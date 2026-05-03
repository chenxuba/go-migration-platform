package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/repository"
)

const (
	deepSeekIEPPlanModel          = "deepseek-v4-pro"
	deepSeekIEPPlanDefaultURL     = "https://api.deepseek.com/chat/completions"
	deepSeekIEPPlanFallbackAPIKey = "sk-3153518fffd24b1588d7b914c388320d"
	deepSeekIEPPlanTimeout        = 180 * time.Second
)

var iepGoalNumberPrefixPattern = regexp.MustCompile(`(^|\s)(?:\d+[.、．]|[一二三四五六七八九十]+[、.．])\s*`)

type iepShortGoalItem struct {
	goal       string
	courseForm string
}

type deepSeekChatRequest struct {
	Model          string                `json:"model"`
	Messages       []deepSeekChatMessage `json:"messages"`
	Temperature    float64               `json:"temperature,omitempty"`
	MaxTokens      int                   `json:"max_tokens,omitempty"`
	ResponseFormat map[string]string     `json:"response_format,omitempty"`
	Thinking       *deepSeekThinking     `json:"thinking,omitempty"`
	Stream         bool                  `json:"stream,omitempty"`
}

type deepSeekChatMessage struct {
	Role             string `json:"role"`
	Content          string `json:"content,omitempty"`
	ReasoningContent string `json:"reasoning_content,omitempty"`
}

type deepSeekThinking struct {
	Type string `json:"type"`
}

type deepSeekChatResponse struct {
	Choices []struct {
		Message      deepSeekChatMessage `json:"message"`
		FinishReason string              `json:"finish_reason,omitempty"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

type deepSeekChatStreamChunk struct {
	Choices []struct {
		Delta        deepSeekChatMessage `json:"delta"`
		FinishReason string              `json:"finish_reason,omitempty"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

type pep3IEPPlanPromptPayload struct {
	Student       pep3IEPPlanPromptStudent       `json:"student"`
	Assessment    pep3IEPPlanPromptAssessment    `json:"assessment"`
	RehabRecords  []pep3IEPPlanPromptRehabRecord `json:"rehabRecords"`
	OutputRequest pep3IEPPlanPromptOutput        `json:"outputRequest"`
}

type pep3IEPPlanPromptStudent struct {
	Name      string `json:"name"`
	Gender    string `json:"gender"`
	BirthDate string `json:"birthDate"`
	Age       string `json:"age"`
}

type pep3IEPPlanPromptAssessment struct {
	Date    string                        `json:"date"`
	Version string                        `json:"version"`
	Scores  []pep3IEPPlanPromptScaleScore `json:"scores"`
	Remark  string                        `json:"remark,omitempty"`
}

type pep3IEPPlanPromptScaleScore struct {
	ScaleName      string `json:"scaleName"`
	RawScore       int    `json:"rawScore"`
	MaxRawScore    string `json:"maxRawScore,omitempty"`
	DevelopmentAge string `json:"developmentAge,omitempty"`
	PercentileRank string `json:"percentileRank,omitempty"`
	ScaledScore    string `json:"scaledScore,omitempty"`
	Level          string `json:"level,omitempty"`
}

type pep3IEPPlanPromptRehabRecord struct {
	Date           string   `json:"date"`
	Course         string   `json:"course"`
	Teacher        string   `json:"teacher"`
	TrainingTarget string   `json:"trainingTarget,omitempty"`
	TrainingItems  []string `json:"trainingItems,omitempty"`
	Performance    string   `json:"performance,omitempty"`
	Suggestion     string   `json:"suggestion,omitempty"`
	ParentFeedback string   `json:"parentFeedback,omitempty"`
}

type pep3IEPPlanPromptOutput struct {
	Title          string `json:"title"`
	DurationMonths int    `json:"durationMonths"`
	RequiredSchema string `json:"requiredSchema"`
}

func (svc *Service) GeneratePEP3IEPPlanWithAI(userID int64, recordID int64, durationMonths int) (model.PEP3IEPPlanAIResult, error) {
	if recordID <= 0 {
		return model.PEP3IEPPlanAIResult{}, errors.New("invalid assessment record id")
	}
	if durationMonths <= 0 {
		durationMonths = 6
	}

	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return model.PEP3IEPPlanAIResult{}, err
	}
	currentTeacherName := svc.currentIEPPlanTeacherName(context.Background(), userID)
	record, err := svc.repo.GetAssessmentRecord(context.Background(), instID, recordID)
	if err != nil {
		return model.PEP3IEPPlanAIResult{}, err
	}

	var rehabRows []pep3IEPPlanPromptRehabRecord
	if record.StudentID > 0 {
		rows, err := svc.repo.ListRecentPublishedRehabRecordRows(context.Background(), instID, record.StudentID, 12)
		if err != nil {
			return model.PEP3IEPPlanAIResult{}, err
		}
		rehabRows = buildPEP3IEPPlanPromptRehabRecords(rows)
	}

	payload := buildPEP3IEPPlanPromptPayload(record, rehabRows, durationMonths)
	result, err := callDeepSeekIEPPlan(context.Background(), payload)
	if err != nil {
		return model.PEP3IEPPlanAIResult{}, err
	}
	return normalizePEP3IEPPlanAIResult(result, record, rehabRows, currentTeacherName, durationMonths), nil
}

func (svc *Service) GeneratePEP3IEPPlanWithAIStream(ctx context.Context, userID int64, recordID int64, durationMonths int, onDelta func(string) error) (model.PEP3IEPPlanAIResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if recordID <= 0 {
		return model.PEP3IEPPlanAIResult{}, errors.New("invalid assessment record id")
	}
	if durationMonths <= 0 {
		durationMonths = 6
	}

	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return model.PEP3IEPPlanAIResult{}, err
	}
	currentTeacherName := svc.currentIEPPlanTeacherName(ctx, userID)
	record, err := svc.repo.GetAssessmentRecord(ctx, instID, recordID)
	if err != nil {
		return model.PEP3IEPPlanAIResult{}, err
	}

	var rehabRows []pep3IEPPlanPromptRehabRecord
	if record.StudentID > 0 {
		rows, err := svc.repo.ListRecentPublishedRehabRecordRows(ctx, instID, record.StudentID, 12)
		if err != nil {
			return model.PEP3IEPPlanAIResult{}, err
		}
		rehabRows = buildPEP3IEPPlanPromptRehabRecords(rows)
	}

	payload := buildPEP3IEPPlanPromptPayload(record, rehabRows, durationMonths)
	result, err := callDeepSeekIEPPlanStream(ctx, payload, onDelta)
	if err != nil {
		return model.PEP3IEPPlanAIResult{}, err
	}
	return normalizePEP3IEPPlanAIResult(result, record, rehabRows, currentTeacherName, durationMonths), nil
}

func buildPEP3IEPPlanPromptPayload(record model.AssessmentRecordDetailVO, rehabRecords []pep3IEPPlanPromptRehabRecord, durationMonths int) pep3IEPPlanPromptPayload {
	return pep3IEPPlanPromptPayload{
		Student: pep3IEPPlanPromptStudent{
			Name:      strings.TrimSpace(record.StudentName),
			Gender:    strings.TrimSpace(record.StudentGender),
			BirthDate: formatIEPPlanDate(record.BirthDate),
			Age:       formatIEPPlanAge(record.AgeYears, record.AgeMonths, record.AgeDays),
		},
		Assessment: pep3IEPPlanPromptAssessment{
			Date:    formatIEPPlanDate(record.AssessmentDate),
			Version: strings.TrimSpace(record.ScaleVersion),
			Scores:  buildPEP3IEPPlanPromptScores(record.ResultJSON),
			Remark:  strings.TrimSpace(record.Remark),
		},
		RehabRecords: rehabRecords,
		OutputRequest: pep3IEPPlanPromptOutput{
			Title:          "康复教学半年计划",
			DurationMonths: durationMonths,
			RequiredSchema: "只输出JSON：title, student{name,gender,birthDate}, meta{planDate,participant,implementer,startDate,endDate}, rows[{domain,longGoal,shortGoal,courseForm,startEndDate}]。rows是表格行，不要输出家庭干预计划。每个康复领域至少3行rows；每行shortGoal只能放1条短期目标；同一领域longGoal必须完全相同，写成至少2条编号长期目标并用\\n分隔；courseForm必须根据评估结果和近期训练记录判断，常见值为个训、集体课；startEndDate按自然月份阶段填写，不能每行都写整个计划周期。",
		},
	}
}

func (svc *Service) currentIEPPlanTeacherName(ctx context.Context, userID int64) string {
	instUserID, err := svc.repo.FindInstUserIDByUserID(ctx, userID)
	if err != nil || instUserID <= 0 {
		return ""
	}
	return strings.TrimSpace(svc.repo.GetStaffNameByID(ctx, &instUserID))
}

func buildPEP3IEPPlanPromptScores(raw json.RawMessage) []pep3IEPPlanPromptScaleScore {
	var score PEP3ScoreResponse
	if len(raw) == 0 || json.Unmarshal(raw, &score) != nil {
		return nil
	}
	items := make([]pep3IEPPlanPromptScaleScore, 0, len(score.Result.Scales))
	for _, scale := range score.Result.Scales {
		item := pep3IEPPlanPromptScaleScore{
			ScaleName: strings.TrimSpace(firstNonEmptyExportValue(scale.ScaleName, scale.ScaleCode)),
			RawScore:  scale.RawScore,
			Level:     strings.TrimSpace(scale.Level),
		}
		if scale.MaxRawScore != nil {
			item.MaxRawScore = strconv.Itoa(*scale.MaxRawScore)
		}
		if scale.DevelopmentAge != nil {
			item.DevelopmentAge = strings.TrimSpace(scale.DevelopmentAge.Text)
		}
		if scale.PercentileRank != nil {
			item.PercentileRank = strings.TrimSpace(scale.PercentileRank.Text)
		}
		if scale.ScaledScore != nil {
			item.ScaledScore = strings.TrimSpace(scale.ScaledScore.Text)
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ScaleName < items[j].ScaleName
	})
	return items
}

func buildPEP3IEPPlanPromptRehabRecords(rows []repository.RehabRecordWordExportRow) []pep3IEPPlanPromptRehabRecord {
	result := make([]pep3IEPPlanPromptRehabRecord, 0, len(rows))
	for _, row := range rows {
		view, err := buildRehabRecordWordExportView(row, "")
		if err != nil {
			continue
		}
		items := make([]string, 0, len(view.TrainingItems))
		for _, item := range view.TrainingItems {
			title := strings.TrimSpace(item.Title)
			content := strings.TrimSpace(item.Content)
			switch {
			case title != "" && content != "":
				items = append(items, title+"："+content)
			case title != "":
				items = append(items, title)
			case content != "":
				items = append(items, content)
			}
		}
		result = append(result, pep3IEPPlanPromptRehabRecord{
			Date:           strings.TrimSpace(view.TrainingDate),
			Course:         strings.TrimSpace(firstNonEmptyExportValue(view.ClassName, row.SourceName, row.LessonName)),
			Teacher:        strings.TrimSpace(view.TeacherName),
			TrainingTarget: strings.TrimSpace(view.TrainingTarget),
			TrainingItems:  items,
			Performance:    strings.TrimSpace(view.Performance),
			Suggestion:     strings.TrimSpace(view.Suggestion),
			ParentFeedback: strings.TrimSpace(view.ParentFeedback),
		})
	}
	return result
}

func callDeepSeekIEPPlan(ctx context.Context, payload pep3IEPPlanPromptPayload) (model.PEP3IEPPlanAIResult, error) {
	apiKey := strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY"))
	if apiKey == "" {
		apiKey = deepSeekIEPPlanFallbackAPIKey
	}
	if apiKey == "" {
		return model.PEP3IEPPlanAIResult{}, errors.New("DEEPSEEK_API_KEY is not configured")
	}
	endpoint := strings.TrimSpace(os.Getenv("DEEPSEEK_API_BASE_URL"))
	if endpoint == "" {
		endpoint = deepSeekIEPPlanDefaultURL
	}

	requestBody, err := buildDeepSeekIEPPlanRequestBody(payload, false)
	if err != nil {
		return model.PEP3IEPPlanAIResult{}, err
	}

	requestCtx, cancel := context.WithTimeout(ctx, deepSeekIEPPlanTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return model.PEP3IEPPlanAIResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := (&http.Client{Timeout: deepSeekIEPPlanTimeout + 5*time.Second}).Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(requestCtx.Err(), context.DeadlineExceeded) {
			return model.PEP3IEPPlanAIResult{}, fmt.Errorf("DeepSeek API 生成超时（%d秒），请稍后重试", int(deepSeekIEPPlanTimeout.Seconds()))
		}
		return model.PEP3IEPPlanAIResult{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.PEP3IEPPlanAIResult{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return model.PEP3IEPPlanAIResult{}, fmt.Errorf("DeepSeek API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var chatResponse deepSeekChatResponse
	if err := json.Unmarshal(body, &chatResponse); err != nil {
		return model.PEP3IEPPlanAIResult{}, err
	}
	if chatResponse.Error != nil && strings.TrimSpace(chatResponse.Error.Message) != "" {
		return model.PEP3IEPPlanAIResult{}, errors.New(chatResponse.Error.Message)
	}
	if len(chatResponse.Choices) == 0 {
		return model.PEP3IEPPlanAIResult{}, errors.New("DeepSeek API returned empty choices")
	}
	content := strings.TrimSpace(chatResponse.Choices[0].Message.Content)
	if content == "" {
		finishReason := strings.TrimSpace(chatResponse.Choices[0].FinishReason)
		reasoning := strings.TrimSpace(chatResponse.Choices[0].Message.ReasoningContent)
		if reasoning != "" {
			return model.PEP3IEPPlanAIResult{}, errors.New("DeepSeek API 只返回了 reasoning_content，没有返回最终JSON内容，请重试")
		}
		if finishReason != "" {
			return model.PEP3IEPPlanAIResult{}, fmt.Errorf("DeepSeek API returned empty content, finish_reason=%s", finishReason)
		}
		return model.PEP3IEPPlanAIResult{}, errors.New("DeepSeek API returned empty content")
	}

	var result model.PEP3IEPPlanAIResult
	if err := json.Unmarshal([]byte(extractJSONContent(content)), &result); err != nil {
		return model.PEP3IEPPlanAIResult{}, fmt.Errorf("parse DeepSeek IEP JSON: %w", err)
	}
	result.Model = deepSeekIEPPlanModel
	return result, nil
}

func buildDeepSeekIEPPlanRequestBody(payload pep3IEPPlanPromptPayload, stream bool) ([]byte, error) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return json.Marshal(deepSeekChatRequest{
		Model: deepSeekIEPPlanModel,
		Messages: []deepSeekChatMessage{
			{
				Role: "system",
				Content: strings.Join([]string{
					"你是儿童康复机构的IEP计划生成助手。",
					"根据PEP-3评估结果和近期儿童训练记录，生成可落地的康复教学半年计划。",
					"必须输出严格JSON，不要Markdown，不要代码块，不要解释。",
					"表格结构只能是：康复领域、长期目标、短期目标、课程形式、起止日期。",
					"不要输出家庭干预计划。长期目标和短期目标要具体、可训练、可观察。",
					"每个康复领域至少输出3行短期目标，一行只能放1条短期目标。",
					"同一康复领域的longGoal要写成同一个字符串，至少包含2条长期目标，用\\n分隔并编号；同领域每行longGoal保持完全相同，便于合并成一个单元格。",
					"courseForm要根据近期儿童训练记录、目标场景和实际干预组织方式判断；一对一、个训、个别训练写个训；集体、小组、融合、团体场景写集体课；不能所有目标无依据地统一写同一种课程形式。",
					"起止日期要按自然月份阶段划分：3个月计划分3段，每段1个月；6个月计划分3段，每段2个月；不要把每条短期目标都写成生成当天到半年后。",
				}, "\n"),
			},
			{
				Role:    "user",
				Content: string(payloadJSON),
			},
		},
		Temperature: 0.25,
		MaxTokens:   4096,
		ResponseFormat: map[string]string{
			"type": "json_object",
		},
		Thinking: &deepSeekThinking{Type: "disabled"},
		Stream:   stream,
	})
}

func callDeepSeekIEPPlanStream(ctx context.Context, payload pep3IEPPlanPromptPayload, onDelta func(string) error) (model.PEP3IEPPlanAIResult, error) {
	apiKey := strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY"))
	if apiKey == "" {
		apiKey = deepSeekIEPPlanFallbackAPIKey
	}
	if apiKey == "" {
		return model.PEP3IEPPlanAIResult{}, errors.New("DEEPSEEK_API_KEY is not configured")
	}
	endpoint := strings.TrimSpace(os.Getenv("DEEPSEEK_API_BASE_URL"))
	if endpoint == "" {
		endpoint = deepSeekIEPPlanDefaultURL
	}

	requestBody, err := buildDeepSeekIEPPlanRequestBody(payload, true)
	if err != nil {
		return model.PEP3IEPPlanAIResult{}, err
	}
	requestCtx, cancel := context.WithTimeout(ctx, deepSeekIEPPlanTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return model.PEP3IEPPlanAIResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := (&http.Client{Timeout: deepSeekIEPPlanTimeout + 5*time.Second}).Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(requestCtx.Err(), context.DeadlineExceeded) {
			return model.PEP3IEPPlanAIResult{}, fmt.Errorf("DeepSeek API 生成超时（%d秒），请稍后重试", int(deepSeekIEPPlanTimeout.Seconds()))
		}
		return model.PEP3IEPPlanAIResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return model.PEP3IEPPlanAIResult{}, fmt.Errorf("DeepSeek API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var content strings.Builder
	var reasoning strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ":") || !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" {
			continue
		}
		if data == "[DONE]" {
			break
		}
		var chunk deepSeekChatStreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			return model.PEP3IEPPlanAIResult{}, fmt.Errorf("parse DeepSeek stream chunk: %w", err)
		}
		if chunk.Error != nil && strings.TrimSpace(chunk.Error.Message) != "" {
			return model.PEP3IEPPlanAIResult{}, errors.New(chunk.Error.Message)
		}
		for _, choice := range chunk.Choices {
			if text := choice.Delta.Content; text != "" {
				content.WriteString(text)
				if onDelta != nil {
					if err := onDelta(text); err != nil {
						return model.PEP3IEPPlanAIResult{}, err
					}
				}
			}
			if text := strings.TrimSpace(choice.Delta.ReasoningContent); text != "" {
				reasoning.WriteString(text)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return model.PEP3IEPPlanAIResult{}, err
	}

	text := strings.TrimSpace(content.String())
	if text == "" {
		if strings.TrimSpace(reasoning.String()) != "" {
			return model.PEP3IEPPlanAIResult{}, errors.New("DeepSeek API 只返回了 reasoning_content，没有返回最终JSON内容，请重试")
		}
		return model.PEP3IEPPlanAIResult{}, errors.New("DeepSeek API returned empty content")
	}
	var result model.PEP3IEPPlanAIResult
	if err := json.Unmarshal([]byte(extractJSONContent(text)), &result); err != nil {
		return model.PEP3IEPPlanAIResult{}, fmt.Errorf("parse DeepSeek IEP JSON: %w", err)
	}
	result.Model = deepSeekIEPPlanModel
	return result, nil
}

func extractJSONContent(content string) string {
	text := strings.TrimSpace(content)
	if strings.HasPrefix(text, "```") {
		text = strings.TrimPrefix(text, "```json")
		text = strings.TrimPrefix(text, "```")
		text = strings.TrimSuffix(text, "```")
		text = strings.TrimSpace(text)
	}
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start >= 0 && end > start {
		return text[start : end+1]
	}
	return text
}

func normalizePEP3IEPPlanAIResult(result model.PEP3IEPPlanAIResult, record model.AssessmentRecordDetailVO, rehabRecords []pep3IEPPlanPromptRehabRecord, currentTeacherName string, durationMonths int) model.PEP3IEPPlanAIResult {
	result.Title = iepPlanTitle(durationMonths)
	result.Model = deepSeekIEPPlanModel
	if strings.TrimSpace(result.Student.Name) == "" {
		result.Student.Name = strings.TrimSpace(record.StudentName)
	}
	if strings.TrimSpace(result.Student.Gender) == "" {
		result.Student.Gender = strings.TrimSpace(record.StudentGender)
	}
	if strings.TrimSpace(result.Student.BirthDate) == "" {
		result.Student.BirthDate = formatIEPPlanDate(record.BirthDate)
	}
	if strings.TrimSpace(result.Meta.PlanDate) == "" {
		result.Meta.PlanDate = time.Now().Format("2006-01-02")
	}
	result.Meta.Participant = firstNonEmptyExportValue(strings.TrimSpace(currentTeacherName), strings.TrimSpace(result.Meta.Participant), strings.TrimSpace(record.ExaminerName))
	if strings.TrimSpace(result.Meta.Implementer) == "" {
		result.Meta.Implementer = firstNonEmptyExportValue(strings.TrimSpace(currentTeacherName), strings.TrimSpace(record.ExaminerName))
	}
	startDate, endDate := iepPlanWholeMonthDateRange(record, durationMonths)
	result.Meta.StartDate = startDate
	result.Meta.EndDate = endDate
	result.Rows = normalizeIEPPlanAIRows(result.Rows, iepPlanStageDateRanges(record, durationMonths), inferIEPCourseFormFromRehabRecords(rehabRecords))
	return result
}

func iepPlanTitle(durationMonths int) string {
	if durationMonths == 3 {
		return "康复教学三个月计划"
	}
	return "康复教学半年计划"
}

func iepPlanBaseDate(record model.AssessmentRecordDetailVO) time.Time {
	candidates := []string{
		formatIEPPlanDate(record.AssessmentDate),
		formatIEPPlanDate(record.CreatedTime),
		time.Now().Format("2006-01-02"),
	}
	for _, candidate := range candidates {
		parsed, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(candidate), time.Local)
		if err == nil {
			return parsed
		}
	}
	return time.Now()
}

func iepPlanWholeMonthDateRange(record model.AssessmentRecordDetailVO, durationMonths int) (string, string) {
	if durationMonths <= 0 {
		durationMonths = 6
	}
	base := iepPlanBaseDate(record)
	start := time.Date(base.Year(), base.Month(), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, durationMonths, 0).AddDate(0, 0, -1)
	return start.Format("2006-01-02"), end.Format("2006-01-02")
}

func iepPlanStageDateRanges(record model.AssessmentRecordDetailVO, durationMonths int) []string {
	if durationMonths <= 0 {
		durationMonths = 6
	}
	base := iepPlanBaseDate(record)
	start := time.Date(base.Year(), base.Month(), 1, 0, 0, 0, 0, time.Local)
	stageCount := 3
	monthBase := durationMonths / stageCount
	monthRemainder := durationMonths % stageCount
	ranges := make([]string, 0, stageCount)
	current := start
	for i := 0; i < stageCount; i++ {
		months := monthBase
		if i < monthRemainder {
			months++
		}
		if months <= 0 {
			months = 1
		}
		end := current.AddDate(0, months, 0).AddDate(0, 0, -1)
		ranges = append(ranges, current.Format("2006-01-02")+" - "+end.Format("2006-01-02"))
		current = end.AddDate(0, 0, 1)
	}
	return ranges
}

func normalizeIEPPlanAIRows(rows []model.PEP3IEPPlanRow, stageRanges []string, defaultCourseForm string) []model.PEP3IEPPlanRow {
	type domainGroup struct {
		domain     string
		longGoals  []string
		shortGoals []iepShortGoalItem
	}
	groups := make([]*domainGroup, 0)
	groupIndex := make(map[string]*domainGroup)
	for _, row := range rows {
		domain := strings.TrimSpace(row.Domain)
		if domain == "" {
			domain = "综合康复"
		}
		if strings.TrimSpace(row.Domain) == "" && strings.TrimSpace(row.LongGoal) == "" && strings.TrimSpace(row.ShortGoal) == "" {
			continue
		}
		group := groupIndex[domain]
		if group == nil {
			group = &domainGroup{domain: domain}
			groupIndex[domain] = group
			groups = append(groups, group)
		}
		group.longGoals = appendUniqueGoalLines(group.longGoals, splitIEPGoalLines(row.LongGoal)...)
		courseForm := normalizeIEPCourseForm(row.CourseForm)
		for _, shortGoal := range splitIEPGoalLines(row.ShortGoal) {
			group.shortGoals = appendUniqueShortGoalItems(group.shortGoals, iepShortGoalItem{
				goal:       shortGoal,
				courseForm: firstNonEmptyExportValue(courseForm, defaultCourseForm, inferIEPCourseFormFromText(shortGoal, group.domain)),
			})
		}
	}
	if len(groups) == 0 {
		groups = append(groups, &domainGroup{domain: "综合康复"})
	}

	normalized := make([]model.PEP3IEPPlanRow, 0, len(groups)*3)
	for _, group := range groups {
		longGoals := ensureIEPLongGoalLines(group.domain, group.longGoals)
		shortGoals := ensureIEPShortGoalItems(group.domain, group.shortGoals, defaultCourseForm)
		longGoalText := numberedIEPGoalText(longGoals)
		for index, shortGoal := range shortGoals {
			normalized = append(normalized, model.PEP3IEPPlanRow{
				Domain:       group.domain,
				LongGoal:     longGoalText,
				ShortGoal:    shortGoal.goal,
				CourseForm:   firstNonEmptyExportValue(shortGoal.courseForm, defaultCourseForm, inferIEPCourseFormFromText(shortGoal.goal, group.domain), "个训"),
				StartEndDate: stageDateForGoal(stageRanges, index, len(shortGoals)),
			})
		}
	}
	return normalized
}

func splitIEPGoalLines(text string) []string {
	value := strings.TrimSpace(text)
	if value == "" {
		return nil
	}
	value = strings.NewReplacer("\r\n", "\n", "\r", "\n", "；", "\n", ";", "\n").Replace(value)
	value = iepGoalNumberPrefixPattern.ReplaceAllString(value, "\n")
	parts := strings.Split(value, "\n")
	lines := make([]string, 0, len(parts))
	for _, part := range parts {
		line := strings.Trim(strings.TrimSpace(part), "，,。.；;、")
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}

func appendUniqueGoalLines(lines []string, additions ...string) []string {
	for _, addition := range additions {
		addition = strings.TrimSpace(addition)
		if addition == "" {
			continue
		}
		exists := false
		for _, existing := range lines {
			if existing == addition {
				exists = true
				break
			}
		}
		if !exists {
			lines = append(lines, addition)
		}
	}
	return lines
}

func appendUniqueShortGoalItems(items []iepShortGoalItem, additions ...iepShortGoalItem) []iepShortGoalItem {
	for _, addition := range additions {
		addition.goal = strings.TrimSpace(addition.goal)
		addition.courseForm = normalizeIEPCourseForm(addition.courseForm)
		if addition.goal == "" {
			continue
		}
		exists := false
		for _, existing := range items {
			if existing.goal == addition.goal {
				exists = true
				break
			}
		}
		if !exists {
			items = append(items, addition)
		}
	}
	return items
}

func ensureIEPLongGoalLines(domain string, goals []string) []string {
	result := append([]string(nil), goals...)
	result = appendUniqueGoalLines(result,
		"提升"+domain+"相关核心能力，能在适当提示下稳定参与并完成目标任务。",
		"将"+domain+"训练内容泛化到课堂活动、同伴互动和日常生活中，提高主动性、持续性和独立完成度。",
	)
	if len(result) > 2 {
		return result[:2]
	}
	return result
}

func ensureIEPShortGoalItems(domain string, goals []iepShortGoalItem, defaultCourseForm string) []iepShortGoalItem {
	result := append([]iepShortGoalItem(nil), goals...)
	fallbacks := []string{
		"在教师示范和语言提示下，能参与" + domain + "相关活动并完成基础目标任务，连续3次课程中达到70%以上。",
		"在少量提示下，能将" + domain + "目标应用到对应训练流程中，连续3次课程中达到75%以上。",
		"在自然活动中，能较稳定完成" + domain + "目标并减少成人辅助，连续3次课程中达到80%以上。",
	}
	for _, fallback := range fallbacks {
		if len(result) >= 3 {
			break
		}
		result = appendUniqueShortGoalItems(result, iepShortGoalItem{
			goal:       fallback,
			courseForm: firstNonEmptyExportValue(defaultCourseForm, inferIEPCourseFormFromText(fallback, domain)),
		})
	}
	return result
}

func normalizeIEPCourseForm(value string) string {
	text := strings.TrimSpace(value)
	if text == "" {
		return ""
	}
	if form := detectIEPCourseForm(text); form != "" {
		return form
	}
	if len([]rune(text)) > 8 {
		return ""
	}
	return text
}

func detectIEPCourseForm(text string) string {
	value := strings.TrimSpace(text)
	if value == "" {
		return ""
	}
	lower := strings.ToLower(value)
	switch {
	case strings.Contains(value, "一对一"), strings.Contains(value, "1对1"), strings.Contains(lower, "1v1"), strings.Contains(value, "个训"), strings.Contains(value, "个别"):
		return "个训"
	case strings.Contains(value, "集体"), strings.Contains(value, "小组"), strings.Contains(value, "团体"), strings.Contains(value, "融合"):
		return "集体课"
	default:
		return ""
	}
}

func inferIEPCourseFormFromRehabRecords(records []pep3IEPPlanPromptRehabRecord) string {
	counts := map[string]int{}
	for _, record := range records {
		form := inferIEPCourseFormFromText(record.Course, record.TrainingTarget)
		if form != "" {
			counts[form]++
		}
	}
	if counts["个训"] == 0 && counts["集体课"] == 0 {
		return ""
	}
	if counts["集体课"] > counts["个训"] {
		return "集体课"
	}
	return "个训"
}

func inferIEPCourseFormFromText(parts ...string) string {
	text := strings.TrimSpace(strings.Join(parts, " "))
	if text == "" {
		return ""
	}
	return detectIEPCourseForm(text)
}

func numberedIEPGoalText(goals []string) string {
	lines := make([]string, 0, len(goals))
	for index, goal := range goals {
		goal = strings.TrimSpace(goal)
		if goal == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("%d. %s", index+1, goal))
	}
	return strings.Join(lines, "\n")
}

func stageDateForGoal(stageRanges []string, goalIndex int, goalCount int) string {
	if len(stageRanges) == 0 {
		return ""
	}
	if goalCount <= 0 {
		return stageRanges[0]
	}
	rangeIndex := goalIndex * len(stageRanges) / goalCount
	if rangeIndex < 0 {
		rangeIndex = 0
	}
	if rangeIndex >= len(stageRanges) {
		rangeIndex = len(stageRanges) - 1
	}
	return stageRanges[rangeIndex]
}

func formatIEPPlanDate(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02")
}

func formatIEPPlanAge(years, months, days int) string {
	parts := make([]string, 0, 3)
	if years > 0 {
		parts = append(parts, fmt.Sprintf("%d岁", years))
	}
	if months > 0 {
		parts = append(parts, fmt.Sprintf("%d月", months))
	}
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%d天", days))
	}
	return strings.Join(parts, "")
}
