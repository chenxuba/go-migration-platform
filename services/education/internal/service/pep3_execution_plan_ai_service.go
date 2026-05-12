package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

type pep3ExecutionPlanPromptPayload struct {
	PlanType       string                         `json:"planType"`
	SourcePlan     model.PEP3IEPPlanAIResult      `json:"sourcePlan"`
	MonthlyPlan    *model.PEP3MonthlyPlanAIResult `json:"monthlyPlan,omitempty"`
	Target         pep3ExecutionPlanTarget        `json:"target"`
	OutputRequest  string                         `json:"outputRequest"`
	GenerationNote string                         `json:"generationNote"`
}

type pep3ExecutionPlanTarget struct {
	MonthIndex          int      `json:"monthIndex,omitempty"`
	MonthLabel          string   `json:"monthLabel,omitempty"`
	WeekIndex           int      `json:"weekIndex,omitempty"`
	WeekLabel           string   `json:"weekLabel,omitempty"`
	DurationMonths      int      `json:"durationMonths,omitempty"`
	StartDate           string   `json:"startDate"`
	EndDate             string   `json:"endDate"`
	WeekStartDate       string   `json:"weekStartDate,omitempty"`
	WeekEndDate         string   `json:"weekEndDate,omitempty"`
	WeekRangeText       string   `json:"weekRangeText,omitempty"`
	WeekRanges          []string `json:"weekRanges,omitempty"`
	CandidateWeekCount  int      `json:"candidateWeekCount,omitempty"`
	CandidateWeekRanges []string `json:"candidateWeekRanges,omitempty"`
	WeekDates           []string `json:"weekDates,omitempty"`
	RestWeekdays        []int    `json:"restWeekdays,omitempty"`
}

type pep3ExecutionPlanPrepared struct {
	PlanType           string
	SystemPrompt       string
	Payload            pep3ExecutionPlanPromptPayload
	SourcePlan         model.PEP3IEPPlanAIResult
	Target             pep3ExecutionPlanTarget
	CurrentTeacherName string
}

func (svc *Service) preparePEP3ExecutionPlanGeneration(ctx context.Context, userID int64, req model.PEP3ExecutionPlanGenerateRequest) (pep3ExecutionPlanPrepared, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if req.ID <= 0 {
		return pep3ExecutionPlanPrepared{}, errors.New("invalid assessment record id")
	}
	planType := strings.ToLower(strings.TrimSpace(req.PlanType))
	if planType != "monthly" && planType != "weekly" {
		return pep3ExecutionPlanPrepared{}, errors.New("invalid execution plan type")
	}
	if planType == "monthly" {
		req.RestWeekdays = normalizeExecutionPlanRestWeekdays(req.RestWeekdays)
	} else if planType == "weekly" {
		if len(req.RestWeekdays) == 0 && req.MonthlyPlan != nil && len(req.MonthlyPlan.RestWeekdays) > 0 {
			req.RestWeekdays = append([]int(nil), req.MonthlyPlan.RestWeekdays...)
		}
		req.RestWeekdays = normalizeExecutionPlanRestWeekdays(req.RestWeekdays)
	}
	if len(req.SourcePlan.Rows) == 0 {
		return pep3ExecutionPlanPrepared{}, errors.New("请先生成或选择IEP总计划")
	}
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return pep3ExecutionPlanPrepared{}, err
	}
	record, err := svc.repo.GetAssessmentRecord(ctx, instID, req.ID)
	if err != nil {
		return pep3ExecutionPlanPrepared{}, err
	}
	currentTeacherName := svc.currentIEPPlanTeacherName(ctx, userID)
	sourcePlan := normalizeExecutionSourceIEPPlan(req.SourcePlan, record, currentTeacherName, req.DurationMonths)
	target := buildExecutionPlanTarget(sourcePlan, req.DurationMonths, req.TargetMonthIndex, req.TargetWeekIndex, req.RestWeekdays)
	payload := pep3ExecutionPlanPromptPayload{
		PlanType:    planType,
		SourcePlan:  sourcePlan,
		MonthlyPlan: req.MonthlyPlan,
		Target:      target,
	}
	prepared := pep3ExecutionPlanPrepared{
		PlanType:           planType,
		Payload:            payload,
		SourcePlan:         sourcePlan,
		Target:             target,
		CurrentTeacherName: currentTeacherName,
	}
	switch planType {
	case "monthly":
		prepared.SystemPrompt = "你是儿童康复机构的月度教学计划生成助手。必须根据IEP总计划生成可执行月计划。"
		prepared.Payload.OutputRequest = "只输出JSON：title, student{name,gender,birthDate}, meta{planDate,participant,implementer,startDate,endDate,monthLabel,sourceTitle}, rows[{domain,longGoal,shortGoal,candidateTrainingItems[{content,startEndDate}],courseForm}]。title必须写成“康复教学X月计划”；必须基于sourcePlan的康复领域、长期目标、短期目标拆解target.monthLabel当月执行内容；candidateTrainingItems是全月候选训练内容池，固定输出target.candidateWeekCount条，不受target.weekRanges实际展示周数影响；第N条candidateTrainingItem的startEndDate必须直接等于target.candidateWeekRanges[N-1]；candidateTrainingItems.content必须体现由浅入深的全月递进关系，写清训练材料、活动、提示方式、步骤或泛化场景，不要照抄短期目标本身，也不要在内容开头写“第一周/第二周/第三周/第四周/第五周”这类周序号。不要输出trainingItems字段。"
		prepared.Payload.GenerationNote = "月计划是IEP总计划中某一个自然月的执行拆解，一次只生成target.monthIndex对应月份；不批量生成其他月份，不重新创造长期目标和领域；短期目标来自总IEP，本月只选取目标月份相关内容。target.candidateWeekRanges代表该月完整候选训练节奏，固定为4条或5条；系统会再根据target.weekRanges把候选池映射成当前实际展示周次，并用于后续周计划承接。你只负责把candidateTrainingItems按全月 progression 生成好。"
	case "weekly":
		prepared.SystemPrompt = "你是儿童康复机构的周计划日记录卡生成助手。必须根据IEP或月计划生成本周可执行训练记录卡。"
		prepared.Payload.OutputRequest = "只输出JSON：title, student{name,gender,birthDate}, teacherName, courseName, trainingDate, preparation, weekDates[], rows[{project,content,completion[]}]。title必须写成“康复教学周计划日记录卡X月第X周”；必须基于monthlyPlan生成target.monthLabel和target.weekLabel对应周计划；如果monthlyPlan为空，则直接基于sourcePlan生成该周计划。当前周区间以target.weekRangeText为准，weekDates必须严格使用target.weekDates；如果monthlyPlan存在，优先选用startEndDate覆盖target.weekRangeText的trainingItems作为本周训练依据，保证周计划内容与月计划中的周区间推进一致；rows用于周计划日记录卡，project是训练项目，content是本周训练内容，completion长度必须等于weekDates长度且先留空字符串。"
		prepared.Payload.GenerationNote = "周计划是某个月份某一周的执行记录卡，不写长期目标列；一次只生成target.weekIndex对应周次，不能批量生成其他周；训练内容要具体到本周能执行的活动、材料、提示和反应标准。weekDates就是完成情况下面要展示的真实日期周期，必须与target.weekRangeText和月计划对应trainingItems的起止日期规则一致。没有月计划时允许直接从IEP总计划拆出本周安排，但必须保持依据清晰。"
	}
	return prepared, nil
}

func (svc *Service) GeneratePEP3ExecutionPlanWithAI(ctx context.Context, userID int64, req model.PEP3ExecutionPlanGenerateRequest) (any, error) {
	prepared, err := svc.preparePEP3ExecutionPlanGeneration(ctx, userID, req)
	if err != nil {
		return nil, err
	}
	switch prepared.PlanType {
	case "monthly":
		var result model.PEP3MonthlyPlanAIResult
		if err := callDeepSeekExecutionPlan(ctx, prepared.SystemPrompt, prepared.Payload, &result); err != nil {
			return nil, err
		}
		normalized := normalizePEP3MonthlyExecutionPlan(result, prepared.SourcePlan, prepared.Target, prepared.CurrentTeacherName)
		return normalized, nil
	case "weekly":
		var result model.PEP3WeeklyPlanAIResult
		if err := callDeepSeekExecutionPlan(ctx, prepared.SystemPrompt, prepared.Payload, &result); err != nil {
			return nil, err
		}
		normalized := normalizePEP3WeeklyExecutionPlan(result, prepared.SourcePlan, req.MonthlyPlan, prepared.Target, prepared.CurrentTeacherName)
		return normalized, nil
	default:
		return nil, errors.New("invalid execution plan type")
	}
}

func (svc *Service) GeneratePEP3ExecutionPlanWithAIStream(ctx context.Context, userID int64, req model.PEP3ExecutionPlanGenerateRequest, onDelta func(string) error) (any, *model.DeepSeekUsageVO, error) {
	prepared, err := svc.preparePEP3ExecutionPlanGeneration(ctx, userID, req)
	if err != nil {
		return nil, nil, err
	}
	switch prepared.PlanType {
	case "monthly":
		var result model.PEP3MonthlyPlanAIResult
		usage, err := callDeepSeekExecutionPlanStream(ctx, prepared.SystemPrompt, prepared.Payload, &result, onDelta)
		if err != nil {
			return nil, usage, err
		}
		normalized := normalizePEP3MonthlyExecutionPlan(result, prepared.SourcePlan, prepared.Target, prepared.CurrentTeacherName)
		return normalized, usage, nil
	case "weekly":
		var result model.PEP3WeeklyPlanAIResult
		usage, err := callDeepSeekExecutionPlanStream(ctx, prepared.SystemPrompt, prepared.Payload, &result, onDelta)
		if err != nil {
			return nil, usage, err
		}
		normalized := normalizePEP3WeeklyExecutionPlan(result, prepared.SourcePlan, req.MonthlyPlan, prepared.Target, prepared.CurrentTeacherName)
		return normalized, usage, nil
	default:
		return nil, nil, errors.New("invalid execution plan type")
	}
}

func callDeepSeekExecutionPlan(ctx context.Context, systemPrompt string, payload pep3ExecutionPlanPromptPayload, result any) error {
	apiKey := strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY"))
	if apiKey == "" {
		apiKey = deepSeekIEPPlanFallbackAPIKey
	}
	if apiKey == "" {
		return errors.New("DEEPSEEK_API_KEY is not configured")
	}
	endpoint := strings.TrimSpace(os.Getenv("DEEPSEEK_API_BASE_URL"))
	if endpoint == "" {
		endpoint = deepSeekIEPPlanDefaultURL
	}
	requestBody, err := buildDeepSeekExecutionPlanRequestBody(systemPrompt, payload, false)
	if err != nil {
		return err
	}
	requestCtx, cancel := context.WithTimeout(ctx, deepSeekIEPPlanTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := (&http.Client{Timeout: deepSeekIEPPlanTimeout + 5*time.Second}).Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(requestCtx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("DeepSeek API 生成超时（%d秒），请稍后重试", int(deepSeekIEPPlanTimeout.Seconds()))
		}
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("DeepSeek API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var chatResponse deepSeekChatResponse
	if err := json.Unmarshal(body, &chatResponse); err != nil {
		return err
	}
	if chatResponse.Error != nil && strings.TrimSpace(chatResponse.Error.Message) != "" {
		return errors.New(chatResponse.Error.Message)
	}
	if len(chatResponse.Choices) == 0 {
		return errors.New("DeepSeek API returned empty choices")
	}
	content := strings.TrimSpace(chatResponse.Choices[0].Message.Content)
	if content == "" {
		return errors.New("DeepSeek API returned empty content")
	}
	if err := parseDeepSeekExecutionPlanResult(content, result); err != nil {
		return fmt.Errorf("parse DeepSeek execution plan JSON: %w", err)
	}
	return nil
}

func buildDeepSeekExecutionPlanRequestBody(systemPrompt string, payload pep3ExecutionPlanPromptPayload, stream bool) ([]byte, error) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	request := deepSeekChatRequest{
		Model: deepSeekIEPPlanModel,
		Messages: []deepSeekChatMessage{
			{
				Role: "system",
				Content: strings.Join([]string{
					systemPrompt,
					"必须输出严格JSON，不要Markdown，不要代码块，不要解释。",
					"不要使用空字符串占位生成目标或训练内容；没有内容的行不要输出。",
					"课程形式必须依据源计划保留或合理判断，常见值为个训、集体课。",
				}, "\n"),
			},
			{Role: "user", Content: string(payloadJSON)},
		},
		Temperature: 0.22,
		MaxTokens:   8192,
		ResponseFormat: map[string]string{
			"type": "json_object",
		},
		Thinking: &deepSeekThinking{Type: "disabled"},
		Stream:   stream,
	}
	if stream {
		request.StreamOptions = &deepSeekStreamOptions{IncludeUsage: true}
	}
	return json.Marshal(request)
}

func callDeepSeekExecutionPlanStream(ctx context.Context, systemPrompt string, payload pep3ExecutionPlanPromptPayload, result any, onDelta func(string) error) (*model.DeepSeekUsageVO, error) {
	apiKey := strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY"))
	if apiKey == "" {
		apiKey = deepSeekIEPPlanFallbackAPIKey
	}
	if apiKey == "" {
		return nil, errors.New("DEEPSEEK_API_KEY is not configured")
	}
	endpoint := strings.TrimSpace(os.Getenv("DEEPSEEK_API_BASE_URL"))
	if endpoint == "" {
		endpoint = deepSeekIEPPlanDefaultURL
	}
	requestBody, err := buildDeepSeekExecutionPlanRequestBody(systemPrompt, payload, true)
	if err != nil {
		return nil, err
	}
	requestCtx, cancel := context.WithTimeout(ctx, deepSeekIEPPlanTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := (&http.Client{Timeout: deepSeekIEPPlanTimeout + 5*time.Second}).Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(requestCtx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf("DeepSeek API 生成超时（%d秒），请稍后重试", int(deepSeekIEPPlanTimeout.Seconds()))
		}
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("DeepSeek API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var content strings.Builder
	var reasoning strings.Builder
	var usage *model.DeepSeekUsageVO
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
			return usage, fmt.Errorf("parse DeepSeek execution plan stream chunk: %w", err)
		}
		if chunk.Error != nil && strings.TrimSpace(chunk.Error.Message) != "" {
			return usage, errors.New(chunk.Error.Message)
		}
		if chunk.Usage != nil {
			usage = toDeepSeekUsageVO(chunk.Usage)
		}
		for _, choice := range chunk.Choices {
			if text := choice.Delta.Content; text != "" {
				content.WriteString(text)
				if onDelta != nil {
					if err := onDelta(text); err != nil {
						return usage, err
					}
				}
			}
			if text := strings.TrimSpace(choice.Delta.ReasoningContent); text != "" {
				reasoning.WriteString(text)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return usage, err
	}
	text := strings.TrimSpace(content.String())
	if text == "" {
		if strings.TrimSpace(reasoning.String()) != "" {
			return usage, errors.New("DeepSeek API 只返回了 reasoning_content，没有返回最终JSON内容，请重试")
		}
		return usage, errors.New("DeepSeek API returned empty content")
	}
	if err := parseDeepSeekExecutionPlanResult(text, result); err != nil {
		return usage, fmt.Errorf("parse DeepSeek execution plan JSON: %w", err)
	}
	return usage, nil
}

func parseDeepSeekExecutionPlanResult(content string, result any) error {
	candidates := extractJSONContentCandidates(content)
	if len(candidates) == 0 {
		return errors.New("no complete JSON object")
	}
	var lastErr error
	for _, candidate := range candidates {
		if err := unmarshalDeepSeekExecutionPlanCandidate(candidate, result); err == nil {
			return nil
		} else {
			lastErr = err
		}
	}
	if lastErr == nil {
		lastErr = errors.New("no complete JSON object")
	}
	return lastErr
}

func unmarshalDeepSeekExecutionPlanCandidate(candidate string, result any) error {
	if err := json.Unmarshal([]byte(candidate), result); err == nil {
		return nil
	} else {
		repaired := escapeBareControlCharsInJSONString(candidate)
		if repaired == candidate {
			return err
		}
		if repairedErr := json.Unmarshal([]byte(repaired), result); repairedErr == nil {
			return nil
		} else {
			return err
		}
	}
}

func normalizeExecutionSourceIEPPlan(plan model.PEP3IEPPlanAIResult, record model.AssessmentRecordDetailVO, currentTeacherName string, durationMonths int) model.PEP3IEPPlanAIResult {
	if durationMonths != 6 {
		durationMonths = 3
	}
	plan.Title = firstNonEmptyExportValue(strings.TrimSpace(plan.Title), iepPlanTitle(durationMonths))
	plan.Student.Name = firstNonEmptyExportValue(strings.TrimSpace(plan.Student.Name), strings.TrimSpace(record.StudentName))
	plan.Student.Gender = firstNonEmptyExportValue(strings.TrimSpace(plan.Student.Gender), strings.TrimSpace(record.StudentGender))
	plan.Student.BirthDate = firstNonEmptyExportValue(strings.TrimSpace(plan.Student.BirthDate), formatIEPPlanDate(record.BirthDate))
	if planDate := pep3AssessmentPlanDate(record); planDate != "" {
		plan.Meta.PlanDate = planDate
	} else {
		plan.Meta.PlanDate = firstNonEmptyExportValue(strings.TrimSpace(plan.Meta.PlanDate), time.Now().Format("2006-01-02"))
	}
	plan = applyPEP3IEPPlanHeaderValues(plan, pep3IEPPlanHeaderValuesForRecord(record))
	startDate, endDate := iepPlanWholeMonthDateRange(record, durationMonths)
	plan.Meta.StartDate = firstNonEmptyExportValue(strings.TrimSpace(plan.Meta.StartDate), startDate)
	plan.Meta.EndDate = firstNonEmptyExportValue(strings.TrimSpace(plan.Meta.EndDate), endDate)
	rows := make([]model.PEP3IEPPlanRow, 0, len(plan.Rows))
	for _, row := range plan.Rows {
		shortGoal := strings.TrimSpace(row.ShortGoal)
		if shortGoal == "" {
			continue
		}
		rows = append(rows, model.PEP3IEPPlanRow{
			Domain:       firstNonEmptyExportValue(strings.TrimSpace(row.Domain), "综合康复"),
			LongGoal:     strings.TrimSpace(row.LongGoal),
			ShortGoal:    shortGoal,
			CourseForm:   firstNonEmptyExportValue(normalizeIEPCourseForm(row.CourseForm), "个训"),
			StartEndDate: strings.TrimSpace(row.StartEndDate),
		})
	}
	plan.Rows = rows
	return plan
}

func buildExecutionPlanTarget(sourcePlan model.PEP3IEPPlanAIResult, durationMonths, targetMonthIndex, targetWeekIndex int, restWeekdays []int) pep3ExecutionPlanTarget {
	if durationMonths != 6 {
		durationMonths = 3
	}
	monthRanges := executionMonthRangesForPlan(sourcePlan, durationMonths)
	if len(monthRanges) == 0 {
		periodStart, periodEnd := executionPlanPeriodRange(sourcePlan, durationMonths)
		monthRanges = []pep3ExecutionMonthRange{{Start: periodStart, End: periodEnd}}
	}
	if targetMonthIndex < 1 {
		targetMonthIndex = 1
	}
	if targetMonthIndex > len(monthRanges) {
		targetMonthIndex = len(monthRanges)
	}
	monthRange := monthRanges[targetMonthIndex-1]
	normalizedRestWeekdays := normalizeExecutionPlanRestWeekdays(restWeekdays)
	weekRanges := calendarWeekRangesForDateRange(monthRange.Start, monthRange.End, normalizedRestWeekdays)
	weekCount := len(weekRanges)
	if weekCount <= 0 && targetWeekIndex > 0 {
		weekRanges = []pep3ExecutionWeekRange{{Start: monthRange.Start, End: monthRange.End}}
		weekCount = 1
	}
	weekIndex := 0
	weekLabel := ""
	weekDates := make([]string, 0, 7)
	weekRangeStart := ""
	weekRangeEnd := ""
	weekRangeText := ""
	weekRangeTexts := make([]string, 0, len(weekRanges))
	for _, item := range weekRanges {
		weekRangeTexts = append(weekRangeTexts, executionWeekRangeText(item))
	}
	candidateWeekRanges := candidateWeekRangesForMonthRange(
		monthRange.Start,
		normalizedRestWeekdays,
	)
	candidateWeekRangeTexts := make([]string, 0, len(candidateWeekRanges))
	for _, item := range candidateWeekRanges {
		candidateWeekRangeTexts = append(
			candidateWeekRangeTexts,
			executionWeekRangeText(item),
		)
	}
	if targetWeekIndex > 0 {
		if targetWeekIndex > weekCount {
			targetWeekIndex = weekCount
		}
		if targetWeekIndex < 1 {
			targetWeekIndex = 1
		}
		weekRange := weekRanges[targetWeekIndex-1]
		weekIndex = targetWeekIndex
		weekLabel = fmt.Sprintf("第%d周", targetWeekIndex)
		weekDates = buildWorkingWeekDatesForRange(weekRange, normalizedRestWeekdays)
		weekRangeStart = weekRange.Start.Format("2006-01-02")
		weekRangeEnd = weekRange.End.Format("2006-01-02")
		weekRangeText = executionWeekRangeText(weekRange)
	}
	return pep3ExecutionPlanTarget{
		MonthIndex:          targetMonthIndex,
		MonthLabel:          fmt.Sprintf("%d月", int(monthRange.Start.Month())),
		WeekIndex:           weekIndex,
		WeekLabel:           weekLabel,
		DurationMonths:      durationMonths,
		StartDate:           monthRange.Start.Format("2006-01-02"),
		EndDate:             monthRange.End.Format("2006-01-02"),
		WeekStartDate:       weekRangeStart,
		WeekEndDate:         weekRangeEnd,
		WeekRangeText:       weekRangeText,
		WeekRanges:          weekRangeTexts,
		CandidateWeekCount:  len(candidateWeekRanges),
		CandidateWeekRanges: candidateWeekRangeTexts,
		WeekDates:           weekDates,
		RestWeekdays:        normalizedRestWeekdays,
	}
}

func parseIEPPlanDateValue(text string) time.Time {
	value := strings.TrimSpace(text)
	if len(value) >= 10 {
		value = value[:10]
	}
	parsed, err := time.ParseInLocation("2006-01-02", value, time.Local)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func normalizePEP3MonthlyExecutionPlan(result model.PEP3MonthlyPlanAIResult, sourcePlan model.PEP3IEPPlanAIResult, target pep3ExecutionPlanTarget, currentTeacherName string) model.PEP3MonthlyPlanAIResult {
	result.Title = fmt.Sprintf("康复教学%s计划", target.MonthLabel)
	result.Model = deepSeekIEPPlanModel
	result.Student.Name = firstNonEmptyExportValue(strings.TrimSpace(result.Student.Name), strings.TrimSpace(sourcePlan.Student.Name))
	result.Student.Gender = firstNonEmptyExportValue(strings.TrimSpace(result.Student.Gender), strings.TrimSpace(sourcePlan.Student.Gender))
	result.Student.BirthDate = firstNonEmptyExportValue(strings.TrimSpace(result.Student.BirthDate), strings.TrimSpace(sourcePlan.Student.BirthDate))
	result.RestWeekdays = append([]int(nil), normalizeExecutionPlanRestWeekdays(target.RestWeekdays)...)
	result.Meta.PlanDate = firstNonEmptyExportValue(strings.TrimSpace(sourcePlan.Meta.PlanDate), strings.TrimSpace(result.Meta.PlanDate), time.Now().Format("2006-01-02"))
	result.Meta.Participant = firstNonEmptyExportValue(strings.TrimSpace(sourcePlan.Meta.Participant), strings.TrimSpace(result.Meta.Participant), currentTeacherName)
	result.Meta.Implementer = firstNonEmptyExportValue(strings.TrimSpace(sourcePlan.Meta.Implementer), strings.TrimSpace(result.Meta.Implementer), currentTeacherName)
	result.Meta.StartDate = firstNonEmptyExportValue(strings.TrimSpace(result.Meta.StartDate), target.StartDate)
	result.Meta.EndDate = firstNonEmptyExportValue(strings.TrimSpace(result.Meta.EndDate), target.EndDate)
	result.Meta.MonthLabel = firstNonEmptyExportValue(strings.TrimSpace(result.Meta.MonthLabel), target.MonthLabel)
	result.Meta.SourceTitle = firstNonEmptyExportValue(strings.TrimSpace(result.Meta.SourceTitle), strings.TrimSpace(sourcePlan.Title))
	rows := make([]model.PEP3MonthlyPlanRow, 0, len(result.Rows))
	for _, row := range result.Rows {
		shortGoal := strings.TrimSpace(row.ShortGoal)
		if shortGoal == "" {
			continue
		}
		candidateTrainingItems := normalizeMonthlyCandidateTrainingItems(
			row,
			target,
		)
		trainingItems := materializeMonthlyTrainingItemsFromCandidates(
			candidateTrainingItems,
			target,
			row.ShortGoal,
		)
		rows = append(rows, model.PEP3MonthlyPlanRow{
			Domain:                 firstNonEmptyExportValue(strings.TrimSpace(row.Domain), "综合康复"),
			LongGoal:               strings.TrimSpace(row.LongGoal),
			ShortGoal:              shortGoal,
			CandidateTrainingItems: candidateTrainingItems,
			TrainingItems:          trainingItems,
			CourseForm:             firstNonEmptyExportValue(normalizeIEPCourseForm(row.CourseForm), "个训"),
		})
	}
	result.Rows = rows
	return result
}

func normalizePEP3MonthlyTrainingItems(row model.PEP3MonthlyPlanRow, target pep3ExecutionPlanTarget) []model.PEP3MonthlyTrainingItem {
	candidateTrainingItems := normalizeMonthlyCandidateTrainingItems(row, target)
	return materializeMonthlyTrainingItemsFromCandidates(
		candidateTrainingItems,
		target,
		row.ShortGoal,
	)
}

func normalizeMonthlyCandidateTrainingItems(row model.PEP3MonthlyPlanRow, target pep3ExecutionPlanTarget) []model.PEP3MonthlyTrainingItem {
	sourceItems := row.CandidateTrainingItems
	if len(sourceItems) == 0 {
		sourceItems = row.TrainingItems
	}
	items := make([]model.PEP3MonthlyTrainingItem, 0, len(sourceItems))
	for _, item := range sourceItems {
		content := normalizeMonthlyTrainingItemContent(item.Content)
		if content == "" {
			continue
		}
		items = append(items, model.PEP3MonthlyTrainingItem{
			Content:      content,
			StartEndDate: strings.TrimSpace(item.StartEndDate),
		})
	}
	if len(items) == 0 {
		items = append(items, model.PEP3MonthlyTrainingItem{Content: "围绕短期目标开展结构化训练，记录提示等级、完成表现和泛化情况。"})
	}
	candidateWeekRanges := parseExecutionWeekRanges(target.CandidateWeekRanges)
	items = normalizeMonthlyCandidateItemPool(
		items,
		candidateWeekRanges,
		expectedMonthlyCandidateItemCountForTarget(target),
		row.ShortGoal,
	)
	for index := range items {
		if index < len(candidateWeekRanges) {
			items[index].StartEndDate = executionWeekRangeText(
				candidateWeekRanges[index],
			)
			continue
		}
		items[index].StartEndDate = firstNonEmptyExportValue(
			monthlyItemDateRangeForWeekRanges(
				candidateWeekRanges,
				target.StartDate,
				target.EndDate,
				index,
				len(items),
			),
			target.StartDate+" - "+target.EndDate,
		)
	}
	return items
}

func normalizeMonthlyCandidateItemPool(
	items []model.PEP3MonthlyTrainingItem,
	candidateWeekRanges []pep3ExecutionWeekRange,
	expectedCount int,
	shortGoal string,
) []model.PEP3MonthlyTrainingItem {
	if expectedCount <= 0 {
		expectedCount = 1
	}
	if len(items) == 0 {
		items = []model.PEP3MonthlyTrainingItem{{
			Content: "围绕短期目标开展结构化训练，记录提示等级、完成表现和泛化情况。",
		}}
	}
	if len(items) >= expectedCount {
		return normalizeMonthlyTrainingItemCount(items, expectedCount, shortGoal)
	}
	slotIndexes := monthlyCandidateSourceIndexes(
		items,
		candidateWeekRanges,
		expectedCount,
	)
	if len(slotIndexes) != len(items) {
		slotIndexes = monthlyTrainingAnchorPositions(len(items), expectedCount)
	}
	slotIndexes = normalizeCandidateIndexes(slotIndexes, expectedCount)
	slots := make([]model.PEP3MonthlyTrainingItem, expectedCount)
	for index, item := range items {
		if index >= len(slotIndexes) {
			break
		}
		slotIndex := slotIndexes[index]
		if slotIndex < 0 {
			slotIndex = 0
		}
		if slotIndex >= expectedCount {
			slotIndex = expectedCount - 1
		}
		slots[slotIndex].Content = mergeMonthlyTrainingItemContent(
			slots[slotIndex].Content,
			item.Content,
		)
	}
	fillMonthlyTrainingItemGaps(slots, shortGoal)
	return slots
}

func expectedMonthlyTrainingItemCountForTarget(target pep3ExecutionPlanTarget) int {
	expectedCount := len(parseExecutionWeekRanges(target.WeekRanges))
	if expectedCount > 0 {
		return expectedCount
	}
	return 1
}

func expectedMonthlyCandidateItemCountForTarget(target pep3ExecutionPlanTarget) int {
	if target.CandidateWeekCount > 0 {
		return target.CandidateWeekCount
	}
	if len(target.CandidateWeekRanges) > 0 {
		return len(target.CandidateWeekRanges)
	}
	return len(candidateWeekRangesForMonthRange(
		parseIEPPlanDateValue(target.StartDate),
		target.RestWeekdays,
	))
}

func materializeMonthlyTrainingItemsFromCandidates(
	candidateItems []model.PEP3MonthlyTrainingItem,
	target pep3ExecutionPlanTarget,
	shortGoal string,
) []model.PEP3MonthlyTrainingItem {
	visibleWeekRanges := parseExecutionWeekRanges(target.WeekRanges)
	if len(visibleWeekRanges) == 0 {
		content := strings.TrimSpace(shortGoal)
		if len(candidateItems) > 0 &&
			strings.TrimSpace(candidateItems[0].Content) != "" {
			content = candidateItems[0].Content
		}
		if content == "" {
			content = "围绕短期目标开展结构化训练，记录提示等级、完成表现和泛化情况。"
		}
		return []model.PEP3MonthlyTrainingItem{{
			Content:      normalizeMonthlyTrainingItemContent(content),
			StartEndDate: strings.TrimSpace(target.StartDate + " - " + target.EndDate),
		}}
	}
	selectedIndexes := selectCandidateIndexesForVisibleWeekRanges(
		visibleWeekRanges,
		parseExecutionWeekRanges(target.CandidateWeekRanges),
		len(candidateItems),
	)
	items := make([]model.PEP3MonthlyTrainingItem, 0, len(visibleWeekRanges))
	fallbackContent := strings.TrimSpace(shortGoal)
	if fallbackContent == "" {
		fallbackContent = "围绕短期目标开展结构化训练，记录提示等级、完成表现和泛化情况。"
	}
	for index, weekRange := range visibleWeekRanges {
		content := fallbackContent
		if index < len(selectedIndexes) {
			selectedIndex := selectedIndexes[index]
			if selectedIndex >= 0 &&
				selectedIndex < len(candidateItems) &&
				strings.TrimSpace(candidateItems[selectedIndex].Content) != "" {
				content = candidateItems[selectedIndex].Content
			}
		}
		items = append(items, model.PEP3MonthlyTrainingItem{
			Content:      normalizeMonthlyTrainingItemContent(content),
			StartEndDate: executionWeekRangeText(weekRange),
		})
	}
	return items
}

func normalizeMonthlyTrainingItemCount(items []model.PEP3MonthlyTrainingItem, expectedCount int, shortGoal string) []model.PEP3MonthlyTrainingItem {
	if expectedCount <= 0 {
		expectedCount = 1
	}
	sourceItems := items
	if len(sourceItems) == 0 {
		sourceItems = []model.PEP3MonthlyTrainingItem{{Content: ""}}
	}
	if len(sourceItems) < expectedCount {
		return expandMonthlyTrainingItems(sourceItems, expectedCount, shortGoal)
	}
	mergedContents := make([]string, expectedCount)
	for index, item := range sourceItems {
		content := normalizeMonthlyTrainingItemContent(item.Content)
		if content == "" {
			continue
		}
		targetIndex := index * expectedCount / len(sourceItems)
		if targetIndex < 0 {
			targetIndex = 0
		}
		if targetIndex >= expectedCount {
			targetIndex = expectedCount - 1
		}
		mergedContents[targetIndex] = mergeMonthlyTrainingItemContent(mergedContents[targetIndex], content)
	}
	fallbackContent := strings.TrimSpace(shortGoal)
	if fallbackContent == "" {
		fallbackContent = "围绕短期目标开展结构化训练，记录提示等级、完成表现和泛化情况。"
	}
	lastContent := ""
	for index := range mergedContents {
		if strings.TrimSpace(mergedContents[index]) == "" {
			if lastContent != "" {
				mergedContents[index] = lastContent
			}
			continue
		}
		lastContent = mergedContents[index]
	}
	lastContent = ""
	for index := len(mergedContents) - 1; index >= 0; index-- {
		if strings.TrimSpace(mergedContents[index]) == "" {
			if lastContent != "" {
				mergedContents[index] = lastContent
			}
			continue
		}
		lastContent = mergedContents[index]
	}
	normalized := make([]model.PEP3MonthlyTrainingItem, 0, expectedCount)
	for index := range mergedContents {
		content := normalizeMonthlyTrainingItemContent(mergedContents[index])
		if content == "" {
			content = fallbackContent
		}
		normalized = append(normalized, model.PEP3MonthlyTrainingItem{Content: content})
	}
	return normalized
}

func expandMonthlyTrainingItems(items []model.PEP3MonthlyTrainingItem, expectedCount int, shortGoal string) []model.PEP3MonthlyTrainingItem {
	if expectedCount <= 0 {
		expectedCount = 1
	}
	sourceItems := items
	if len(sourceItems) == 0 {
		sourceItems = []model.PEP3MonthlyTrainingItem{{Content: ""}}
	}
	if len(sourceItems) == 1 {
		return expandSingleMonthlyTrainingItem(
			normalizeMonthlyTrainingItemContent(sourceItems[0].Content),
			expectedCount,
			shortGoal,
		)
	}
	positions := monthlyTrainingAnchorPositions(len(sourceItems), expectedCount)
	normalized := make([]model.PEP3MonthlyTrainingItem, expectedCount)
	anchors := make(map[int]string, len(sourceItems))
	for index, item := range sourceItems {
		content := normalizeMonthlyTrainingItemContent(item.Content)
		position := positions[index]
		anchors[position] = content
		normalized[position] = model.PEP3MonthlyTrainingItem{Content: content}
	}
	for index := 0; index < expectedCount; index++ {
		if strings.TrimSpace(normalized[index].Content) != "" {
			continue
		}
		prevIndex, prevContent, hasPrev := nearestMonthlyTrainingAnchorBefore(anchors, index)
		nextIndex, nextContent, hasNext := nearestMonthlyTrainingAnchorAfter(anchors, index)
		normalized[index] = model.PEP3MonthlyTrainingItem{
			Content: pickExpandedMonthlyTrainingContent(
				shortGoal,
				index,
				prevIndex,
				prevContent,
				nextIndex,
				nextContent,
				hasPrev,
				hasNext,
			),
		}
	}
	return normalized
}

func expandSingleMonthlyTrainingItem(content string, expectedCount int, shortGoal string) []model.PEP3MonthlyTrainingItem {
	normalized := make([]model.PEP3MonthlyTrainingItem, 0, expectedCount)
	base := normalizeMonthlyTrainingItemContent(content)
	for index := 0; index < expectedCount; index++ {
		if index == 0 && base != "" {
			normalized = append(normalized, model.PEP3MonthlyTrainingItem{Content: base})
			continue
		}
		normalized = append(normalized, model.PEP3MonthlyTrainingItem{
			Content: pickExpandedMonthlyTrainingContent(
				shortGoal,
				index,
				0,
				base,
				0,
				"",
				base != "",
				false,
			),
		})
	}
	return normalized
}

func monthlyCandidateSourceIndexes(
	items []model.PEP3MonthlyTrainingItem,
	candidateWeekRanges []pep3ExecutionWeekRange,
	expectedCount int,
) []int {
	if len(items) == 0 || expectedCount <= 0 {
		return nil
	}
	anchors := monthlyTrainingAnchorPositions(len(items), expectedCount)
	indexes := make([]int, 0, len(items))
	for index, item := range items {
		if weekRange, ok := parseExecutionWeekRangeText(item.StartEndDate); ok && len(candidateWeekRanges) > 0 {
			indexes = append(indexes, bestCandidateIndexForVisibleRange(weekRange, candidateWeekRanges))
			continue
		}
		if index < len(anchors) {
			indexes = append(indexes, anchors[index])
			continue
		}
		indexes = append(indexes, expectedCount-1)
	}
	return indexes
}

func fillMonthlyTrainingItemGaps(items []model.PEP3MonthlyTrainingItem, shortGoal string) {
	fallback := normalizeMonthlyTrainingItemContent(shortGoal)
	if fallback == "" {
		fallback = "围绕短期目标开展结构化训练，记录提示等级、完成表现和泛化情况。"
	}
	for index := range items {
		if strings.TrimSpace(items[index].Content) != "" {
			continue
		}
		prevIndex, prevContent, hasPrev := nearestMonthlyItemBefore(items, index)
		nextIndex, nextContent, hasNext := nearestMonthlyItemAfter(items, index)
		items[index].Content = pickExpandedMonthlyTrainingContent(
			fallback,
			index,
			prevIndex,
			prevContent,
			nextIndex,
			nextContent,
			hasPrev,
			hasNext,
		)
	}
}

func monthlyTrainingAnchorPositions(sourceCount, expectedCount int) []int {
	if sourceCount <= 0 {
		return nil
	}
	if expectedCount <= 1 || sourceCount == 1 {
		return []int{0}
	}
	positions := make([]int, sourceCount)
	last := -1
	for index := 0; index < sourceCount; index++ {
		position := int(math.Round(float64(index*(expectedCount-1)) / float64(sourceCount-1)))
		if position <= last {
			position = last + 1
		}
		maxAllowed := expectedCount - (sourceCount - index)
		if position > maxAllowed {
			position = maxAllowed
		}
		positions[index] = position
		last = position
	}
	return positions
}

func nearestMonthlyTrainingAnchorBefore(anchors map[int]string, index int) (int, string, bool) {
	for cursor := index - 1; cursor >= 0; cursor-- {
		if content, ok := anchors[cursor]; ok && strings.TrimSpace(content) != "" {
			return cursor, content, true
		}
	}
	return 0, "", false
}

func nearestMonthlyTrainingAnchorAfter(anchors map[int]string, index int) (int, string, bool) {
	for cursor := index + 1; cursor < len(anchors)+8; cursor++ {
		if content, ok := anchors[cursor]; ok && strings.TrimSpace(content) != "" {
			return cursor, content, true
		}
	}
	return 0, "", false
}

func nearestMonthlyItemBefore(items []model.PEP3MonthlyTrainingItem, index int) (int, string, bool) {
	for cursor := index - 1; cursor >= 0; cursor-- {
		if content := strings.TrimSpace(items[cursor].Content); content != "" {
			return cursor, content, true
		}
	}
	return 0, "", false
}

func nearestMonthlyItemAfter(items []model.PEP3MonthlyTrainingItem, index int) (int, string, bool) {
	for cursor := index + 1; cursor < len(items); cursor++ {
		if content := strings.TrimSpace(items[cursor].Content); content != "" {
			return cursor, content, true
		}
	}
	return 0, "", false
}

var monthlyWeekPrefixPattern = regexp.MustCompile(`^(?:第)?[一二三四五六七八九十0-9]+周(?:训练内容|训练项目|内容|安排)?[：:、，,\s]*`)
var monthlyTrainingStagePrefixPattern = regexp.MustCompile(`^(?:导入建立|分步练习|强化提升|独立巩固|泛化维持|本周训练)[：:、，,\s]*`)

func normalizeMonthlyTrainingItemContent(content string) string {
	value := strings.TrimSpace(content)
	if value == "" {
		return ""
	}
	value = monthlyWeekPrefixPattern.ReplaceAllString(value, "")
	value = monthlyTrainingStagePrefixPattern.ReplaceAllString(value, "")
	return strings.TrimSpace(value)
}

func pickExpandedMonthlyTrainingContent(shortGoal string, index, prevIndex int, prevContent string, nextIndex int, nextContent string, hasPrev, hasNext bool) string {
	prevContent = normalizeMonthlyTrainingItemContent(prevContent)
	nextContent = normalizeMonthlyTrainingItemContent(nextContent)
	switch {
	case hasPrev && hasNext && prevContent != "" && nextContent != "":
		if index-prevIndex <= nextIndex-index {
			return prevContent
		}
		return nextContent
	case hasPrev && prevContent != "":
		return prevContent
	case hasNext && nextContent != "":
		return nextContent
	default:
		fallback := normalizeMonthlyTrainingItemContent(shortGoal)
		if fallback != "" {
			return fallback
		}
		return "围绕短期目标开展结构化训练，记录提示等级、完成表现和泛化情况。"
	}
}

func monthlyTrainingStageDescriptor(index, total int) (string, string) {
	if total <= 1 {
		return "本周训练", "完成结构化练习并记录提示等级、准确率与泛化表现"
	}
	progress := float64(index) / float64(total-1)
	switch {
	case progress <= 0.2:
		return "导入建立", "先熟悉材料与规则，在示范和较高提示下完成基础尝试"
	case progress <= 0.45:
		return "分步练习", "在口语或手势提示下完成关键步骤，逐步减少辅助"
	case progress <= 0.7:
		return "强化提升", "增加连续完成次数与准确率，提升任务稳定性"
	case progress <= 0.9:
		return "独立巩固", "提高独立完成比例，并关注不同指令下的保持表现"
	default:
		return "泛化维持", "切换材料或场景进行泛化训练，巩固本月目标表现"
	}
}

func monthlyTrainingContentSnippet(content string) string {
	value := normalizeMonthlyTrainingItemContent(content)
	if value == "" {
		return ""
	}
	for _, separator := range []string{"。", "；", ";", "\n"} {
		if index := strings.Index(value, separator); index > 0 {
			value = strings.TrimSpace(value[:index])
			break
		}
	}
	runes := []rune(value)
	if len(runes) > 26 {
		value = strings.TrimSpace(string(runes[:26]))
	}
	return strings.Trim(value, "：:，,；; ")
}

func mergeMonthlyTrainingItemContent(existing, next string) string {
	current := normalizeMonthlyTrainingItemContent(existing)
	value := normalizeMonthlyTrainingItemContent(next)
	if current == "" {
		return value
	}
	if value == "" || current == value {
		return current
	}
	return current + "；" + value
}

func monthlyItemDateRange(startText, endText string, itemIndex, itemCount int) string {
	return monthlyItemDateRangeForWeekRanges(
		calendarWeekRangesForDateRange(parseIEPPlanDateValue(startText), parseIEPPlanDateValue(endText), normalizeExecutionPlanRestWeekdays(nil)),
		startText,
		endText,
		itemIndex,
		itemCount,
	)
}

func monthlyItemDateRangeForTarget(target pep3ExecutionPlanTarget, itemIndex, itemCount int) string {
	return monthlyItemDateRangeForWeekRanges(parseExecutionWeekRanges(target.WeekRanges), target.StartDate, target.EndDate, itemIndex, itemCount)
}

func monthlyItemDateRangeForWeekRanges(weekRanges []pep3ExecutionWeekRange, startText, endText string, itemIndex, itemCount int) string {
	if itemCount <= 0 {
		itemCount = 1
	}
	if itemIndex < 0 {
		itemIndex = 0
	}
	if itemIndex >= itemCount {
		itemIndex = itemCount - 1
	}
	if len(weekRanges) == 0 {
		start := parseIEPPlanDateValue(startText)
		end := parseIEPPlanDateValue(endText)
		if start.IsZero() || end.IsZero() || end.Before(start) {
			return strings.TrimSpace(startText + " - " + endText)
		}
		return start.Format("2006-01-02") + " - " + end.Format("2006-01-02")
	}
	weekCount := len(weekRanges)
	startIndex := itemIndex * weekCount / itemCount
	endIndex := ((itemIndex + 1) * weekCount / itemCount) - 1
	if endIndex < startIndex {
		endIndex = startIndex
	}
	if startIndex >= weekCount {
		startIndex = weekCount - 1
	}
	if endIndex >= weekCount {
		endIndex = weekCount - 1
	}
	itemStart := weekRanges[startIndex].Start
	itemEnd := weekRanges[endIndex].End
	return itemStart.Format("2006-01-02") + " - " + itemEnd.Format("2006-01-02")
}

func normalizePEP3WeeklyExecutionPlan(result model.PEP3WeeklyPlanAIResult, sourcePlan model.PEP3IEPPlanAIResult, monthlyPlan *model.PEP3MonthlyPlanAIResult, target pep3ExecutionPlanTarget, currentTeacherName string) model.PEP3WeeklyPlanAIResult {
	sourceTitle := strings.TrimSpace(sourcePlan.Title)
	if monthlyPlan != nil && strings.TrimSpace(monthlyPlan.Title) != "" {
		sourceTitle = strings.TrimSpace(monthlyPlan.Title)
	}
	result.Title = fmt.Sprintf("康复教学周计划日记录卡%s%s", target.MonthLabel, target.WeekLabel)
	result.Model = deepSeekIEPPlanModel
	result.Student.Name = firstNonEmptyExportValue(strings.TrimSpace(result.Student.Name), strings.TrimSpace(sourcePlan.Student.Name))
	result.Student.Gender = firstNonEmptyExportValue(strings.TrimSpace(result.Student.Gender), strings.TrimSpace(sourcePlan.Student.Gender))
	result.Student.BirthDate = firstNonEmptyExportValue(strings.TrimSpace(result.Student.BirthDate), strings.TrimSpace(sourcePlan.Student.BirthDate))
	result.TeacherName = firstNonEmptyExportValue(strings.TrimSpace(result.TeacherName), strings.TrimSpace(sourcePlan.Meta.Implementer), currentTeacherName)
	result.CourseName = firstNonEmptyExportValue(strings.TrimSpace(result.CourseName), sourceTitle)
	trainingDate := ""
	if len(target.WeekDates) > 0 {
		visibleWeekDates := make([]string, 0, len(target.WeekDates))
		for _, date := range target.WeekDates {
			if strings.TrimSpace(date) != "" {
				visibleWeekDates = append(visibleWeekDates, strings.TrimSpace(date))
			}
		}
		if len(visibleWeekDates) > 0 {
			trainingDate = visibleWeekDates[0] + " 至 " + visibleWeekDates[len(visibleWeekDates)-1]
		}
	}
	result.TrainingDate = trainingDate
	result.Preparation = firstNonEmptyExportValue(strings.TrimSpace(result.Preparation), "依据"+sourceTitle+"准备训练材料、强化物和提示卡，明确本周训练目标与记录方式。")
	result.SourceTitle = firstNonEmptyExportValue(strings.TrimSpace(result.SourceTitle), sourceTitle)
	result.RestWeekdays = append([]int(nil), normalizeExecutionPlanRestWeekdays(target.RestWeekdays)...)
	result.WeekDates = append([]string(nil), target.WeekDates...)
	rows := make([]model.PEP3WeeklyPlanRow, 0, len(result.Rows))
	for _, row := range result.Rows {
		project := strings.TrimSpace(row.Project)
		content := strings.TrimSpace(row.Content)
		if project == "" && content == "" {
			continue
		}
		completion := append([]string(nil), row.Completion...)
		if len(completion) > len(result.WeekDates) {
			completion = completion[:len(result.WeekDates)]
		}
		for len(completion) < len(result.WeekDates) {
			completion = append(completion, "")
		}
		rows = append(rows, model.PEP3WeeklyPlanRow{
			Project:    firstNonEmptyExportValue(project, "训练项目"),
			Content:    content,
			Completion: completion,
		})
	}
	result.Rows = rows
	return result
}

type pep3ExecutionWeekRange struct {
	Start time.Time
	End   time.Time
}

type pep3ExecutionMonthRange struct {
	Start time.Time
	End   time.Time
}

func executionPlanPeriodRange(sourcePlan model.PEP3IEPPlanAIResult, durationMonths int) (time.Time, time.Time) {
	start := parseIEPPlanDateValue(firstNonEmptyExportValue(strings.TrimSpace(sourcePlan.Meta.StartDate), time.Now().Format("2006-01-02")))
	if start.IsZero() {
		start = time.Now()
	}
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.Local)
	end := parseIEPPlanDateValue(strings.TrimSpace(sourcePlan.Meta.EndDate))
	if end.IsZero() || end.Before(start) {
		_, computedEnd := iepPlanDateRangeFromStart(start, durationMonths)
		end = computedEnd
	}
	return start, time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, time.Local)
}

func executionMonthRangesForPlan(sourcePlan model.PEP3IEPPlanAIResult, durationMonths int) []pep3ExecutionMonthRange {
	start, end := executionPlanPeriodRange(sourcePlan, durationMonths)
	if start.IsZero() || end.IsZero() || end.Before(start) {
		return nil
	}
	ranges := make([]pep3ExecutionMonthRange, 0, durationMonths+1)
	cursor := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.Local)
	for !cursor.After(end) {
		monthStart := cursor
		monthEnd := time.Date(cursor.Year(), cursor.Month()+1, 0, 0, 0, 0, 0, time.Local)
		visibleStart := monthStart
		if visibleStart.Before(start) {
			visibleStart = start
		}
		visibleEnd := monthEnd
		if visibleEnd.After(end) {
			visibleEnd = end
		}
		if !visibleEnd.Before(visibleStart) {
			ranges = append(ranges, pep3ExecutionMonthRange{Start: visibleStart, End: visibleEnd})
		}
		cursor = time.Date(cursor.Year(), cursor.Month()+1, 1, 0, 0, 0, 0, time.Local)
	}
	return ranges
}

func executionPlanMonthCount(sourcePlan model.PEP3IEPPlanAIResult, durationMonths int) int {
	return len(executionMonthRangesForPlan(sourcePlan, durationMonths))
}

func candidateWeekRangesForMonthRange(monthDate time.Time, restWeekdays []int) []pep3ExecutionWeekRange {
	if monthDate.IsZero() {
		return nil
	}
	monthStart := time.Date(monthDate.Year(), monthDate.Month(), 1, 0, 0, 0, 0, time.Local)
	monthEnd := time.Date(monthDate.Year(), monthDate.Month()+1, 0, 0, 0, 0, 0, time.Local)
	baseWeekRanges := calendarWeekRangesForDateRange(
		monthStart,
		monthEnd,
		normalizeExecutionPlanRestWeekdays(restWeekdays),
	)
	if len(baseWeekRanges) == 0 {
		return nil
	}
	return rebalanceWeekRangesForCount(
		baseWeekRanges,
		candidateWeekCountFromBaseWeekRanges(baseWeekRanges),
	)
}

func candidateWeekCountFromBaseWeekRanges(weekRanges []pep3ExecutionWeekRange) int {
	count := len(weekRanges)
	switch {
	case count <= 0:
		return 4
	case count < 4:
		return 4
	case count > 5:
		return 5
	default:
		return count
	}
}

func rebalanceWeekRangesForCount(
	weekRanges []pep3ExecutionWeekRange,
	targetCount int,
) []pep3ExecutionWeekRange {
	if len(weekRanges) == 0 {
		return nil
	}
	if targetCount <= 0 {
		targetCount = 1
	}
	if len(weekRanges) == targetCount {
		result := make([]pep3ExecutionWeekRange, len(weekRanges))
		copy(result, weekRanges)
		return result
	}
	result := make([]pep3ExecutionWeekRange, 0, targetCount)
	for index := 0; index < targetCount; index++ {
		startIndex := index * len(weekRanges) / targetCount
		endIndex := ((index + 1) * len(weekRanges) / targetCount) - 1
		if endIndex < startIndex {
			endIndex = startIndex
		}
		if startIndex >= len(weekRanges) {
			startIndex = len(weekRanges) - 1
		}
		if endIndex >= len(weekRanges) {
			endIndex = len(weekRanges) - 1
		}
		result = append(result, pep3ExecutionWeekRange{
			Start: weekRanges[startIndex].Start,
			End:   weekRanges[endIndex].End,
		})
	}
	return result
}

func normalizeExecutionPlanRestWeekdays(restWeekdays []int) []int {
	source := restWeekdays
	if len(source) == 0 {
		source = []int{7}
	}
	seen := make(map[int]struct{}, len(source))
	normalized := make([]int, 0, 2)
	for _, value := range source {
		if value < 1 || value > 7 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
		if len(normalized) >= 2 {
			break
		}
	}
	if len(normalized) == 0 {
		return []int{7}
	}
	return normalized
}

func inferWeeklyPlanRestWeekdays(plan model.PEP3WeeklyPlanAIResult) []int {
	return normalizeExecutionPlanRestWeekdays(plan.RestWeekdays)
}

func weekdaySetFromSelection(restWeekdays []int) map[time.Weekday]struct{} {
	set := make(map[time.Weekday]struct{}, len(restWeekdays))
	for _, value := range normalizeExecutionPlanRestWeekdays(restWeekdays) {
		var weekday time.Weekday
		if value == 7 {
			weekday = time.Sunday
		} else {
			weekday = time.Weekday(value)
		}
		set[weekday] = struct{}{}
	}
	return set
}

func buildWorkingWeekDatesForRange(weekRange pep3ExecutionWeekRange, restWeekdays []int) []string {
	restSet := weekdaySetFromSelection(restWeekdays)
	dates := make([]string, 0, 7)
	for current := weekRange.Start; !current.After(weekRange.End); current = current.AddDate(0, 0, 1) {
		if _, blocked := restSet[current.Weekday()]; blocked {
			continue
		}
		dates = append(dates, current.Format("2006-01-02"))
	}
	return dates
}

func executionWeekRangeText(weekRange pep3ExecutionWeekRange) string {
	if weekRange.Start.IsZero() || weekRange.End.IsZero() || weekRange.End.Before(weekRange.Start) {
		return ""
	}
	return weekRange.Start.Format("2006-01-02") + " - " + weekRange.End.Format("2006-01-02")
}

func parseExecutionWeekRanges(texts []string) []pep3ExecutionWeekRange {
	ranges := make([]pep3ExecutionWeekRange, 0, len(texts))
	for _, text := range texts {
		weekRange, ok := parseExecutionWeekRangeText(text)
		if !ok {
			continue
		}
		ranges = append(ranges, weekRange)
	}
	return ranges
}

func parseExecutionWeekRangeText(text string) (pep3ExecutionWeekRange, bool) {
	parts := strings.Split(strings.TrimSpace(text), " - ")
	if len(parts) != 2 {
		return pep3ExecutionWeekRange{}, false
	}
	start := parseIEPPlanDateValue(parts[0])
	end := parseIEPPlanDateValue(parts[1])
	if start.IsZero() || end.IsZero() || end.Before(start) {
		return pep3ExecutionWeekRange{}, false
	}
	return pep3ExecutionWeekRange{Start: start, End: end}, true
}

func selectCandidateIndexesForVisibleWeekRanges(
	visibleRanges, candidateRanges []pep3ExecutionWeekRange,
	candidateCount int,
) []int {
	if len(visibleRanges) == 0 || candidateCount <= 0 {
		return nil
	}
	if len(candidateRanges) == 0 {
		return evenlyDistributedCandidateIndexes(len(visibleRanges), candidateCount)
	}
	raw := make([]int, 0, len(visibleRanges))
	for _, visibleRange := range visibleRanges {
		raw = append(raw, bestCandidateIndexForVisibleRange(visibleRange, candidateRanges))
	}
	if len(visibleRanges) > candidateCount {
		return raw
	}
	return normalizeCandidateIndexes(raw, candidateCount)
}

func bestCandidateIndexForVisibleRange(
	visibleRange pep3ExecutionWeekRange,
	candidateRanges []pep3ExecutionWeekRange,
) int {
	if len(candidateRanges) == 0 {
		return 0
	}
	bestIndex := 0
	bestOverlap := -1
	bestDistance := math.MaxFloat64
	visibleCenter := rangeCenterTimestamp(visibleRange)
	for index, candidateRange := range candidateRanges {
		overlap := weekRangeOverlapDays(visibleRange, candidateRange)
		distance := math.Abs(visibleCenter - rangeCenterTimestamp(candidateRange))
		if overlap > bestOverlap || (overlap == bestOverlap && distance < bestDistance) {
			bestIndex = index
			bestOverlap = overlap
			bestDistance = distance
		}
	}
	return bestIndex
}

func normalizeCandidateIndexes(raw []int, candidateCount int) []int {
	if len(raw) == 0 {
		return nil
	}
	if candidateCount <= 0 {
		candidateCount = 1
	}
	result := make([]int, len(raw))
	copy(result, raw)
	for index := range result {
		if result[index] < 0 {
			result[index] = 0
		}
		if result[index] >= candidateCount {
			result[index] = candidateCount - 1
		}
	}
	for index := 1; index < len(result); index++ {
		minAllowed := result[index-1] + 1
		if result[index] < minAllowed {
			result[index] = minAllowed
		}
	}
	for index := len(result) - 2; index >= 0; index-- {
		maxAllowed := result[index+1] - 1
		if result[index] > maxAllowed {
			result[index] = maxAllowed
		}
	}
	for index := range result {
		minAllowed := index
		maxAllowed := candidateCount - (len(result) - index)
		if result[index] < minAllowed {
			result[index] = minAllowed
		}
		if result[index] > maxAllowed {
			result[index] = maxAllowed
		}
	}
	return result
}

func evenlyDistributedCandidateIndexes(visibleCount, candidateCount int) []int {
	if visibleCount <= 0 || candidateCount <= 0 {
		return nil
	}
	if visibleCount == 1 {
		return []int{0}
	}
	indexes := make([]int, visibleCount)
	for index := 0; index < visibleCount; index++ {
		position := int(math.Round(float64(index*(candidateCount-1)) / float64(visibleCount-1)))
		if position < 0 {
			position = 0
		}
		if position >= candidateCount {
			position = candidateCount - 1
		}
		indexes[index] = position
	}
	return normalizeCandidateIndexes(indexes, candidateCount)
}

func weekRangeOverlapDays(left, right pep3ExecutionWeekRange) int {
	if left.Start.IsZero() || left.End.IsZero() || right.Start.IsZero() || right.End.IsZero() {
		return 0
	}
	start := left.Start
	if right.Start.After(start) {
		start = right.Start
	}
	end := left.End
	if right.End.Before(end) {
		end = right.End
	}
	if end.Before(start) {
		return 0
	}
	return int(end.Sub(start).Hours()/24) + 1
}

func rangeCenterTimestamp(value pep3ExecutionWeekRange) float64 {
	if value.Start.IsZero() || value.End.IsZero() {
		return 0
	}
	return float64(value.Start.Unix()+value.End.Unix()) / 2
}

func calendarWeekRangesForDateRange(start, end time.Time, restWeekdays []int) []pep3ExecutionWeekRange {
	if start.IsZero() || end.IsZero() || end.Before(start) {
		return nil
	}
	restSet := weekdaySetFromSelection(restWeekdays)
	ranges := make([]pep3ExecutionWeekRange, 0)
	cursor := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.Local)
	normalizedEnd := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, time.Local)
	for !cursor.After(normalizedEnd) {
		rawWeekEnd := cursor.AddDate(0, 0, 6-int(cursor.Weekday()))
		weekEnd := rawWeekEnd
		if weekEnd.After(normalizedEnd) {
			weekEnd = normalizedEnd
		}
		visibleStart, okStart := firstWorkingDayBetweenDates(cursor, weekEnd, restSet)
		visibleEnd, okEnd := lastWorkingDayBetweenDates(weekEnd, cursor, restSet)
		if okStart && okEnd && !visibleEnd.Before(visibleStart) {
			ranges = append(ranges, pep3ExecutionWeekRange{Start: visibleStart, End: visibleEnd})
		}
		cursor = weekEnd.AddDate(0, 0, 1)
	}
	return ranges
}

func firstWorkingDayBetweenDates(start, end time.Time, restWeekdays map[time.Weekday]struct{}) (time.Time, bool) {
	for current := start; !current.After(end); current = current.AddDate(0, 0, 1) {
		if _, blocked := restWeekdays[current.Weekday()]; !blocked {
			return current, true
		}
	}
	return time.Time{}, false
}

func lastWorkingDayBetweenDates(start, end time.Time, restWeekdays map[time.Weekday]struct{}) (time.Time, bool) {
	for current := start; !current.Before(end); current = current.AddDate(0, 0, -1) {
		if _, blocked := restWeekdays[current.Weekday()]; !blocked {
			return current, true
		}
	}
	return time.Time{}, false
}

func inferExecutionPlanCourseForm(sourcePlan model.PEP3IEPPlanAIResult, monthlyPlan *model.PEP3MonthlyPlanAIResult) string {
	values := make([]string, 0)
	if monthlyPlan != nil {
		for _, row := range monthlyPlan.Rows {
			if courseForm := normalizeIEPCourseForm(row.CourseForm); courseForm != "" {
				values = append(values, courseForm)
			}
		}
	}
	for _, row := range sourcePlan.Rows {
		if courseForm := normalizeIEPCourseForm(row.CourseForm); courseForm != "" {
			values = append(values, courseForm)
		}
	}
	if len(values) == 0 {
		return "个训"
	}
	groupCount := 0
	for _, value := range values {
		if value == "集体课" {
			groupCount++
		}
	}
	if groupCount*2 >= len(values) {
		return "集体课"
	}
	return "个训"
}
