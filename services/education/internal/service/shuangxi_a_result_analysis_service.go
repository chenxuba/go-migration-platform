package service

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/repository"
)

const shuangxiAResultAnalysisCode = "SHUANGXI_A_RESULT_ANALYSIS"

type shuangxiResultAnalysisPromptPayload struct {
	Student       pep3IEPPlanPromptStudent               `json:"student"`
	Assessment    shuangxiResultAnalysisAssessment       `json:"assessment"`
	Domains       []shuangxiResultAnalysisPromptDomain   `json:"domains"`
	OutputRequest erxinReportInterpretationOutputRequest `json:"outputRequest"`
}

type shuangxiResultAnalysisAssessment struct {
	Date              string  `json:"date,omitempty"`
	ScaleVersion      string  `json:"scaleVersion,omitempty"`
	DataStatus        string  `json:"dataStatus,omitempty"`
	Remark            string  `json:"remark,omitempty"`
	TotalRawScore     int     `json:"totalRawScore"`
	MaxRawScore       int     `json:"maxRawScore"`
	CompletionPercent float64 `json:"completionPercent"`
}

type shuangxiResultAnalysisPromptDomain struct {
	DomainCode        string                             `json:"domainCode"`
	DomainName        string                             `json:"domainName"`
	ItemCount         int                                `json:"itemCount"`
	AnsweredItemCount int                                `json:"answeredItemCount"`
	MissingItemCount  int                                `json:"missingItemCount"`
	RawScore          int                                `json:"rawScore"`
	MaxRawScore       int                                `json:"maxRawScore"`
	ScoreRate         float64                            `json:"scoreRate"`
	StrengthItems     []shuangxiResultAnalysisPromptItem `json:"strengthItems,omitempty"`
	WeakItems         []shuangxiResultAnalysisPromptItem `json:"weakItems,omitempty"`
}

type shuangxiResultAnalysisPromptItem struct {
	ItemNo      int    `json:"itemNo"`
	SkillName   string `json:"skillName,omitempty"`
	Title       string `json:"title"`
	Score       int    `json:"score"`
	ScoreLabel  string `json:"scoreLabel,omitempty"`
	Description string `json:"description,omitempty"`
}

func (svc *Service) GetShuangxiAResultAnalysis(userID, recordID int64) (model.ShuangxiResultAnalysisVO, error) {
	instID, record, data, itemScores, err := svc.shuangxiAResultAnalysisContext(userID, recordID)
	if err != nil {
		return model.ShuangxiResultAnalysisVO{}, err
	}
	if svc.repo == nil || instID <= 0 {
		return buildEmptyShuangxiAResultAnalysis(record, data, itemScores), nil
	}
	var cached model.ShuangxiResultAnalysisVO
	_, err = svc.repo.GetAssessmentReportInterpretationJSON(context.Background(), instID, recordID, shuangxiAResultAnalysisCode, &cached)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return buildEmptyShuangxiAResultAnalysis(record, data, itemScores), nil
		}
		return model.ShuangxiResultAnalysisVO{}, err
	}
	fallback := buildEmptyShuangxiAResultAnalysis(record, data, itemScores)
	if isLegacyShuangxiAResultAnalysisAutoTemplate(cached) {
		return fallback, nil
	}
	return normalizeShuangxiAResultAnalysis(cached, fallback, cached.GeneratedBy), nil
}

func (svc *Service) SaveShuangxiAResultAnalysis(ctx context.Context, userID, recordID int64, analysis model.ShuangxiResultAnalysisVO) (model.ShuangxiResultAnalysisVO, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if svc.repo == nil {
		return model.ShuangxiResultAnalysisVO{}, errors.New("assessment repository is not configured")
	}
	instID, record, data, itemScores, err := svc.shuangxiAResultAnalysisContext(userID, recordID)
	if err != nil {
		return model.ShuangxiResultAnalysisVO{}, err
	}
	fallback := buildEmptyShuangxiAResultAnalysis(record, data, itemScores)
	normalized := normalizeShuangxiAResultAnalysis(analysis, fallback, "manual")
	if err := svc.saveShuangxiAResultAnalysis(ctx, instID, userID, recordID, record, normalized); err != nil {
		return model.ShuangxiResultAnalysisVO{}, err
	}
	return normalized, nil
}

func (svc *Service) GenerateShuangxiAResultAnalysisStream(ctx context.Context, userID, recordID int64, onDelta func(string) error) (model.ShuangxiResultAnalysisVO, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if svc.repo == nil {
		return model.ShuangxiResultAnalysisVO{}, errors.New("assessment repository is not configured")
	}
	instID, record, data, itemScores, err := svc.shuangxiAResultAnalysisContext(userID, recordID)
	if err != nil {
		return model.ShuangxiResultAnalysisVO{}, err
	}
	payload := buildShuangxiAResultAnalysisPromptPayload(record, data, itemScores)
	result, err := callDeepSeekShuangxiAResultAnalysisStream(ctx, payload, onDelta)
	if err != nil {
		return model.ShuangxiResultAnalysisVO{}, err
	}
	normalized := normalizeShuangxiAResultAnalysis(result, buildEmptyShuangxiAResultAnalysis(record, data, itemScores), "ai")
	if err := svc.saveShuangxiAResultAnalysis(ctx, instID, userID, recordID, record, normalized); err != nil {
		return model.ShuangxiResultAnalysisVO{}, err
	}
	return normalized, nil
}

func (svc *Service) shuangxiAResultAnalysisContext(userID, recordID int64) (int64, model.AssessmentRecordDetailVO, shuangxiAStaticData, map[int]int, error) {
	if recordID <= 0 {
		return 0, model.AssessmentRecordDetailVO{}, shuangxiAStaticData{}, nil, errors.New("invalid assessment record id")
	}
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return 0, model.AssessmentRecordDetailVO{}, shuangxiAStaticData{}, nil, err
	}
	record, err := svc.GetShuangxiAAssessmentRecord(userID, recordID)
	if err != nil {
		return 0, model.AssessmentRecordDetailVO{}, shuangxiAStaticData{}, nil, err
	}
	data, err := svc.loadShuangxiAStaticData(context.Background())
	if err != nil {
		return 0, model.AssessmentRecordDetailVO{}, shuangxiAStaticData{}, nil, err
	}
	itemScores, err := decodeSavedShuangxiAInputScores(record.InputJSON)
	if err != nil {
		return 0, model.AssessmentRecordDetailVO{}, shuangxiAStaticData{}, nil, err
	}
	return instID, record, data, itemScores, nil
}

func (svc *Service) saveShuangxiAResultAnalysis(ctx context.Context, instID, userID, recordID int64, record model.AssessmentRecordDetailVO, analysis model.ShuangxiResultAnalysisVO) error {
	return svc.repo.UpsertAssessmentReportInterpretationJSON(ctx, repository.AssessmentReportInterpretationJSONEntity{
		InstID:         instID,
		RecordID:       recordID,
		AssessmentCode: shuangxiAResultAnalysisCode,
		SourceHash:     shuangxiAResultAnalysisSourceHash(record),
		Model:          strings.TrimSpace(nonEmptyString(analysis.Model, deepSeekIEPPlanModel)),
		GeneratedBy:    strings.TrimSpace(nonEmptyString(analysis.GeneratedBy, "manual")),
	}, analysis, userID)
}

func buildShuangxiAResultAnalysisPromptPayload(record model.AssessmentRecordDetailVO, data shuangxiAStaticData, itemScores map[int]int) shuangxiResultAnalysisPromptPayload {
	score := shuangxiAResultAnalysisScoreSummary(data, itemScores)
	return shuangxiResultAnalysisPromptPayload{
		Student: pep3IEPPlanPromptStudent{
			Name:      strings.TrimSpace(record.StudentName),
			Gender:    strings.TrimSpace(record.StudentGender),
			BirthDate: formatIEPPlanDate(record.BirthDate),
			Age:       formatIEPPlanAge(record.AgeYears, record.AgeMonths, record.AgeDays),
		},
		Assessment: shuangxiResultAnalysisAssessment{
			Date:              formatIEPPlanDate(record.AssessmentDate),
			ScaleVersion:      strings.TrimSpace(record.ScaleVersion),
			DataStatus:        strings.TrimSpace(record.DataStatus),
			Remark:            strings.TrimSpace(record.Remark),
			TotalRawScore:     score.raw,
			MaxRawScore:       score.max,
			CompletionPercent: score.completion,
		},
		Domains: shuangxiAResultAnalysisPromptDomains(data, itemScores),
		OutputRequest: erxinReportInterpretationOutputRequest{
			RequiredSchema: "只输出JSON：title, rows[{domainCode,domain,strengths,weaknesses,reason,strategy}]。rows必须只包含输入domains，domain使用中文领域名，domainCode保留输入值；rows按输入domains顺序输出；strengths只写优势内容，不要带“优：”；weaknesses只写弱项内容，不要带“弱：”；reason从生理、心理、教学、环境、互动角度分析可能原因，但不要虚构医学诊断；strategy写可执行教学策略。",
			SafetyRules:    "必须基于双溪课程评量表A的题目得分生成；不得修改题目得分、领域得分、日期等事实；不得做医学诊断；不得输出医疗治疗方案；禁止空话套话，如“继续加强训练”“提升综合能力”等没有具体项目的表达；每个单元格要具体、可观察、可训练。",
		},
	}
}

func shuangxiAResultAnalysisPromptDomains(data shuangxiAStaticData, itemScores map[int]int) []shuangxiResultAnalysisPromptDomain {
	domainItems := shuangxiAResultAnalysisItemsByDomain(data)
	domains := shuangxiAProfileDomains(data)
	out := make([]shuangxiResultAnalysisPromptDomain, 0, len(domains))
	for _, domain := range domains {
		items := domainItems[domain.Code]
		rawScore := 0
		answered := 0
		for _, item := range items {
			score, ok := itemScores[item.ItemNo]
			if !ok {
				continue
			}
			rawScore += score
			answered++
		}
		maxScore := domain.MaxRawScore
		if maxScore <= 0 {
			maxScore = len(items) * 3
		}
		scoreRate := 0.0
		if maxScore > 0 {
			scoreRate = math.Round(float64(rawScore)*1000/float64(maxScore)) / 10
		}
		out = append(out, shuangxiResultAnalysisPromptDomain{
			DomainCode:        domain.Code,
			DomainName:        domain.Name,
			ItemCount:         len(items),
			AnsweredItemCount: answered,
			MissingItemCount:  shuangxiAResultAnalysisMaxInt(0, len(items)-answered),
			RawScore:          rawScore,
			MaxRawScore:       maxScore,
			ScoreRate:         scoreRate,
			StrengthItems: shuangxiAResultAnalysisSelectItems(
				items,
				itemScores,
				func(score int) bool { return score >= 2 },
				8,
				true,
			),
			WeakItems: shuangxiAResultAnalysisSelectItems(
				items,
				itemScores,
				func(score int) bool { return score <= 1 },
				8,
				false,
			),
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].ScoreRate != out[j].ScoreRate {
			return out[i].ScoreRate > out[j].ScoreRate
		}
		return false
	})
	return out
}

func buildEmptyShuangxiAResultAnalysis(record model.AssessmentRecordDetailVO, data shuangxiAStaticData, itemScores map[int]int) model.ShuangxiResultAnalysisVO {
	domains := shuangxiAResultAnalysisPromptDomains(data, itemScores)
	rows := make([]model.ShuangxiResultAnalysisRow, 0, len(domains))
	for _, domain := range domains {
		rows = append(rows, model.ShuangxiResultAnalysisRow{
			DomainCode: strings.TrimSpace(domain.DomainCode),
			Domain:     strings.TrimSpace(domain.DomainName),
		})
	}
	return model.ShuangxiResultAnalysisVO{
		Title:       "双溪心智障碍个别化教育课程（三）评量结果分析表",
		GeneratedAt: shuangxiAResultAnalysisGeneratedAt(record),
		Rows:        rows,
	}
}

func callDeepSeekShuangxiAResultAnalysisStream(ctx context.Context, payload shuangxiResultAnalysisPromptPayload, onDelta func(string) error) (model.ShuangxiResultAnalysisVO, error) {
	if onDelta == nil {
		return callDeepSeekShuangxiAResultAnalysis(ctx, payload)
	}
	apiKey := strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY"))
	if apiKey == "" {
		apiKey = deepSeekIEPPlanFallbackAPIKey
	}
	if apiKey == "" {
		return model.ShuangxiResultAnalysisVO{}, errors.New("DEEPSEEK_API_KEY is not configured")
	}
	endpoint := strings.TrimSpace(os.Getenv("DEEPSEEK_API_BASE_URL"))
	if endpoint == "" {
		endpoint = deepSeekIEPPlanDefaultURL
	}
	body, err := buildDeepSeekShuangxiAResultAnalysisRequestBody(payload, true)
	if err != nil {
		return model.ShuangxiResultAnalysisVO{}, err
	}
	requestCtx, cancel := context.WithTimeout(ctx, deepSeekIEPPlanTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return model.ShuangxiResultAnalysisVO{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := (&http.Client{Timeout: deepSeekIEPPlanTimeout + 5*time.Second}).Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(requestCtx.Err(), context.DeadlineExceeded) {
			return model.ShuangxiResultAnalysisVO{}, deepSeekTimeoutError()
		}
		return model.ShuangxiResultAnalysisVO{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		responseBody, _ := io.ReadAll(resp.Body)
		return model.ShuangxiResultAnalysisVO{}, fmt.Errorf("DeepSeek API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
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
			return model.ShuangxiResultAnalysisVO{}, fmt.Errorf("parse DeepSeek stream chunk: %w", err)
		}
		if chunk.Error != nil && strings.TrimSpace(chunk.Error.Message) != "" {
			return model.ShuangxiResultAnalysisVO{}, errors.New(chunk.Error.Message)
		}
		for _, choice := range chunk.Choices {
			if text := choice.Delta.Content; text != "" {
				content.WriteString(text)
				if err := onDelta(text); err != nil {
					return model.ShuangxiResultAnalysisVO{}, err
				}
			}
			if text := strings.TrimSpace(choice.Delta.ReasoningContent); text != "" {
				reasoning.WriteString(text)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		if isDeepSeekDeadlineExceeded(err, requestCtx) {
			return model.ShuangxiResultAnalysisVO{}, deepSeekTimeoutError()
		}
		return model.ShuangxiResultAnalysisVO{}, err
	}
	text := strings.TrimSpace(content.String())
	if text == "" {
		if strings.TrimSpace(reasoning.String()) != "" {
			return model.ShuangxiResultAnalysisVO{}, errors.New("DeepSeek API 只返回了 reasoning_content，没有返回最终JSON内容，请重试")
		}
		return model.ShuangxiResultAnalysisVO{}, errors.New("DeepSeek API returned empty content")
	}
	var result model.ShuangxiResultAnalysisVO
	if err := json.Unmarshal([]byte(extractJSONContent(text)), &result); err != nil {
		return model.ShuangxiResultAnalysisVO{}, fmt.Errorf("parse DeepSeek Shuangxi result analysis JSON: %w", err)
	}
	result.Model = deepSeekIEPPlanModel
	return result, nil
}

func callDeepSeekShuangxiAResultAnalysis(ctx context.Context, payload shuangxiResultAnalysisPromptPayload) (model.ShuangxiResultAnalysisVO, error) {
	apiKey := strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY"))
	if apiKey == "" {
		apiKey = deepSeekIEPPlanFallbackAPIKey
	}
	if apiKey == "" {
		return model.ShuangxiResultAnalysisVO{}, errors.New("DEEPSEEK_API_KEY is not configured")
	}
	endpoint := strings.TrimSpace(os.Getenv("DEEPSEEK_API_BASE_URL"))
	if endpoint == "" {
		endpoint = deepSeekIEPPlanDefaultURL
	}
	body, err := buildDeepSeekShuangxiAResultAnalysisRequestBody(payload, false)
	if err != nil {
		return model.ShuangxiResultAnalysisVO{}, err
	}
	requestCtx, cancel := context.WithTimeout(ctx, deepSeekIEPPlanTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return model.ShuangxiResultAnalysisVO{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := (&http.Client{Timeout: deepSeekIEPPlanTimeout + 5*time.Second}).Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(requestCtx.Err(), context.DeadlineExceeded) {
			return model.ShuangxiResultAnalysisVO{}, deepSeekTimeoutError()
		}
		return model.ShuangxiResultAnalysisVO{}, err
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.ShuangxiResultAnalysisVO{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return model.ShuangxiResultAnalysisVO{}, fmt.Errorf("DeepSeek API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}
	var chatResponse deepSeekChatResponse
	if err := json.Unmarshal(responseBody, &chatResponse); err != nil {
		return model.ShuangxiResultAnalysisVO{}, err
	}
	if chatResponse.Error != nil && strings.TrimSpace(chatResponse.Error.Message) != "" {
		return model.ShuangxiResultAnalysisVO{}, errors.New(chatResponse.Error.Message)
	}
	if len(chatResponse.Choices) == 0 {
		return model.ShuangxiResultAnalysisVO{}, errors.New("DeepSeek API returned empty choices")
	}
	content := strings.TrimSpace(chatResponse.Choices[0].Message.Content)
	if content == "" {
		return model.ShuangxiResultAnalysisVO{}, errors.New("DeepSeek API returned empty content")
	}
	var result model.ShuangxiResultAnalysisVO
	if err := json.Unmarshal([]byte(extractJSONContent(content)), &result); err != nil {
		return model.ShuangxiResultAnalysisVO{}, fmt.Errorf("parse DeepSeek Shuangxi result analysis JSON: %w", err)
	}
	result.Model = deepSeekIEPPlanModel
	return result, nil
}

func buildDeepSeekShuangxiAResultAnalysisRequestBody(payload shuangxiResultAnalysisPromptPayload, stream bool) ([]byte, error) {
	return buildDeepSeekIEPPlanRequestBodyWithPrompt(payload, strings.Join([]string{
		"你是儿童康复机构的双溪课程评量表A评量结果分析助手。",
		"任务是把双溪心智障碍个别化教育课程评量表A的结构化评分结果转换为报告中的“评量结果分析表”内容。",
		"表格列为：领域（依优弱序）、现况分析、原因推断（生理、心理、教学、环境、互动）、建议策略。",
		"现况分析要分为strengths和weaknesses两个字段，分别对应前端显示的“优：”和“弱：”。",
		"每个字段必须引用具体项目表现，不能只写领域名称或总分。",
		"建议策略要可落地，包含提示方式、练习场景、步骤拆解或泛化安排。",
		"不要让每个领域都写成同一句式；不要统一以“经过评量”“该儿童”开头。",
		"必须输出严格JSON，不要Markdown，不要代码块，不要解释。",
		"不得修改输入中的题目得分、领域得分、日期等事实。",
		"不得做医学诊断，不得输出治疗方案。",
	}, "\n"), stream)
}

func normalizeShuangxiAResultAnalysis(result, fallback model.ShuangxiResultAnalysisVO, generatedBy string) model.ShuangxiResultAnalysisVO {
	result.Title = "双溪心智障碍个别化教育课程（三）评量结果分析表"
	result.CourseName = strings.TrimSpace(result.CourseName)
	if strings.TrimSpace(result.Model) == "" && strings.TrimSpace(generatedBy) == "ai" {
		result.Model = deepSeekIEPPlanModel
	}
	result.GeneratedBy = strings.TrimSpace(generatedBy)
	if result.GeneratedBy == "" {
		result.GeneratedBy = strings.TrimSpace(fallback.GeneratedBy)
	}
	if result.GeneratedBy == "" {
		result.GeneratedBy = "manual"
	}
	if strings.TrimSpace(result.GeneratedAt) == "" {
		result.GeneratedAt = time.Now().Format("2006-01-02 15:04:05")
	}
	byDomain := make(map[string]model.ShuangxiResultAnalysisRow, len(result.Rows))
	for _, row := range result.Rows {
		key := normalizeShuangxiAResultAnalysisDomainKey(row.DomainCode, row.Domain, fallback.Rows)
		if key == "" {
			continue
		}
		byDomain[key] = row
	}
	rows := make([]model.ShuangxiResultAnalysisRow, 0, len(fallback.Rows))
	for _, fallbackRow := range fallback.Rows {
		key := normalizeShuangxiAResultAnalysisDomainKey(fallbackRow.DomainCode, fallbackRow.Domain, fallback.Rows)
		row, ok := byDomain[key]
		if !ok {
			row = fallbackRow
		}
		row.DomainCode = fallbackRow.DomainCode
		row.Domain = fallbackRow.Domain
		if strings.TrimSpace(row.Strengths) == "" {
			row.Strengths = fallbackRow.Strengths
		}
		if strings.TrimSpace(row.Weaknesses) == "" {
			row.Weaknesses = fallbackRow.Weaknesses
		}
		if strings.TrimSpace(row.Reason) == "" {
			row.Reason = fallbackRow.Reason
		}
		if strings.TrimSpace(row.Strategy) == "" {
			row.Strategy = fallbackRow.Strategy
		}
		rows = append(rows, trimShuangxiAResultAnalysisRow(row))
	}
	result.Rows = rows
	return result
}

func trimShuangxiAResultAnalysisRow(row model.ShuangxiResultAnalysisRow) model.ShuangxiResultAnalysisRow {
	row.DomainCode = strings.TrimSpace(row.DomainCode)
	row.Domain = strings.TrimSpace(row.Domain)
	row.Strengths = strings.TrimSpace(row.Strengths)
	row.Weaknesses = strings.TrimSpace(row.Weaknesses)
	row.Reason = strings.TrimSpace(row.Reason)
	row.Strategy = strings.TrimSpace(row.Strategy)
	return row
}

func isLegacyShuangxiAResultAnalysisAutoTemplate(analysis model.ShuangxiResultAnalysisVO) bool {
	switch strings.ToLower(strings.TrimSpace(analysis.GeneratedBy)) {
	case "rule", "system", "template":
		return true
	default:
		return false
	}
}

func normalizeShuangxiAResultAnalysisDomainKey(code, domain string, rows []model.ShuangxiResultAnalysisRow) string {
	if key := strings.TrimSpace(code); key != "" {
		return key
	}
	normalized := strings.ReplaceAll(strings.TrimSpace(domain), " ", "")
	for _, row := range rows {
		candidate := strings.ReplaceAll(strings.TrimSpace(row.Domain), " ", "")
		if normalized != "" && normalized == candidate {
			if row.DomainCode != "" {
				return row.DomainCode
			}
			return candidate
		}
	}
	return normalized
}

func shuangxiAResultAnalysisItemsByDomain(data shuangxiAStaticData) map[string][]shuangxiAItemDefinition {
	out := make(map[string][]shuangxiAItemDefinition, len(data.domains))
	for _, item := range data.items {
		code := strings.TrimSpace(item.DomainCode)
		if code == "" {
			continue
		}
		out[code] = append(out[code], item)
	}
	for code := range out {
		sort.SliceStable(out[code], func(i, j int) bool {
			return out[code][i].ItemNo < out[code][j].ItemNo
		})
	}
	return out
}

func shuangxiAResultAnalysisSelectItems(items []shuangxiAItemDefinition, scores map[int]int, accept func(int) bool, limit int, highFirst bool) []shuangxiResultAnalysisPromptItem {
	candidates := make([]shuangxiResultAnalysisPromptItem, 0, limit)
	for _, item := range items {
		score, ok := scores[item.ItemNo]
		if !ok || !accept(score) {
			continue
		}
		candidates = append(candidates, shuangxiResultAnalysisPromptItem{
			ItemNo:      item.ItemNo,
			SkillName:   strings.TrimSpace(item.SkillName),
			Title:       shuangxiAResultAnalysisItemTitle(item),
			Score:       score,
			ScoreLabel:  shuangxiAResultAnalysisScoreLabel(item, score),
			Description: shuangxiAResultAnalysisScoreDescription(item, score),
		})
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Score != candidates[j].Score {
			if highFirst {
				return candidates[i].Score > candidates[j].Score
			}
			return candidates[i].Score < candidates[j].Score
		}
		return candidates[i].ItemNo < candidates[j].ItemNo
	})
	if limit > 0 && len(candidates) > limit {
		candidates = candidates[:limit]
	}
	return candidates
}

func shuangxiAResultAnalysisItemTitle(item shuangxiAItemDefinition) string {
	for _, value := range []string{item.TestItem, item.ItemTitle, item.Describes} {
		if text := strings.TrimSpace(value); text != "" {
			return text
		}
	}
	return fmt.Sprintf("%d号项目", item.ItemNo)
}

func shuangxiAResultAnalysisScoreLabel(item shuangxiAItemDefinition, score int) string {
	for _, option := range item.ScoreOptions {
		if option.Value == score {
			return strings.TrimSpace(option.Label)
		}
	}
	return fmt.Sprintf("%d分", score)
}

func shuangxiAResultAnalysisScoreDescription(item shuangxiAItemDefinition, score int) string {
	for _, option := range item.ScoreOptions {
		if option.Value == score {
			return strings.TrimSpace(option.Description)
		}
	}
	return ""
}

type shuangxiAResultAnalysisScoreTotal struct {
	raw        int
	max        int
	completion float64
}

func shuangxiAResultAnalysisScoreSummary(data shuangxiAStaticData, scores map[int]int) shuangxiAResultAnalysisScoreTotal {
	total := shuangxiAResultAnalysisScoreTotal{}
	itemsByDomain := shuangxiAResultAnalysisItemsByDomain(data)
	for _, domain := range shuangxiAProfileDomains(data) {
		items := itemsByDomain[domain.Code]
		domainMax := domain.MaxRawScore
		if domainMax <= 0 {
			domainMax = len(items) * 3
		}
		total.max += domainMax
		for _, item := range items {
			if score, ok := scores[item.ItemNo]; ok {
				total.raw += score
			}
		}
	}
	if len(data.items) > 0 {
		total.completion = math.Round(float64(len(scores))*1000/float64(len(data.items))) / 10
	}
	return total
}

func shuangxiAResultAnalysisGeneratedAt(record model.AssessmentRecordDetailVO) string {
	if record.UpdatedTime != nil {
		return record.UpdatedTime.Format("2006-01-02 15:04:05")
	}
	if record.CreatedTime != nil {
		return record.CreatedTime.Format("2006-01-02 15:04:05")
	}
	return time.Now().Format("2006-01-02 15:04:05")
}

func shuangxiAResultAnalysisSourceHash(record model.AssessmentRecordDetailVO) string {
	raw, err := json.Marshal(struct {
		Input      json.RawMessage `json:"input,omitempty"`
		Result     json.RawMessage `json:"result,omitempty"`
		DataStatus string          `json:"dataStatus,omitempty"`
		Remark     string          `json:"remark,omitempty"`
	}{
		Input:      record.InputJSON,
		Result:     record.ResultJSON,
		DataStatus: strings.TrimSpace(record.DataStatus),
		Remark:     strings.TrimSpace(record.Remark),
	})
	if err != nil {
		raw = []byte(fmt.Sprintf("%+v", record.ResultJSON))
	}
	sum := sha256.Sum256(raw)
	return fmt.Sprintf("%x", sum[:])
}

func shuangxiAResultAnalysisMaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
