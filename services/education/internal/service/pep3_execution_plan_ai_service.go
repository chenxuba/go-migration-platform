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
	MonthIndex     int      `json:"monthIndex,omitempty"`
	MonthLabel     string   `json:"monthLabel,omitempty"`
	WeekIndex      int      `json:"weekIndex,omitempty"`
	WeekLabel      string   `json:"weekLabel,omitempty"`
	DurationMonths int      `json:"durationMonths,omitempty"`
	StartDate      string   `json:"startDate"`
	EndDate        string   `json:"endDate"`
	WeekDates      []string `json:"weekDates,omitempty"`
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
	target := buildExecutionPlanTarget(sourcePlan, req.DurationMonths, req.TargetMonthIndex, req.TargetWeekIndex)
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
		prepared.Payload.OutputRequest = "只输出JSON：title, student{name,gender,birthDate}, meta{planDate,participant,implementer,startDate,endDate,monthLabel,sourceTitle}, rows[{domain,longGoal,shortGoal,trainingItems[{content,startEndDate}],courseForm}]。title必须写成“康复教学X月计划”；必须基于sourcePlan的康复领域、长期目标、短期目标拆解target.monthLabel当月执行内容；每条短期目标至少2条trainingItems；每一条trainingItems必须有自己的startEndDate，日期必须落在target.startDate到target.endDate内；content必须具体到训练材料、活动、提示方式、步骤或泛化场景；不要照抄短期目标作为训练内容。"
		prepared.Payload.GenerationNote = "月计划是IEP总计划中某一个自然月的执行拆解，一次只生成target.monthIndex对应月份；不批量生成其他月份，不重新创造长期目标和领域；短期目标来自总IEP，本月只选取目标月份相关内容。"
	case "weekly":
		prepared.SystemPrompt = "你是儿童康复机构的周计划日记录卡生成助手。必须根据IEP或月计划生成本周可执行训练记录卡。"
		prepared.Payload.OutputRequest = "只输出JSON：title, student{name,gender,birthDate}, teacherName, courseName, trainingDate, preparation, weekDates[], rows[{project,content,completion[]}]。title必须写成“康复教学周计划日记录卡X月第X周”；必须基于monthlyPlan生成target.monthLabel和target.weekLabel对应周计划；如果monthlyPlan为空，则直接基于sourcePlan生成该周计划。weekDates必须使用target.weekDates；rows用于周计划日记录卡，project是训练项目，content是本周训练内容，completion长度必须等于weekDates长度且先留空字符串。"
		prepared.Payload.GenerationNote = "周计划是某个月份某一周的执行记录卡，不写长期目标列；一次只生成target.weekIndex对应周次，不能批量生成其他周；训练内容要具体到本周能执行的活动、材料、提示和反应标准。没有月计划时允许直接从IEP总计划拆出本周安排，但必须保持依据清晰。"
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

func (svc *Service) GeneratePEP3ExecutionPlanWithAIStream(ctx context.Context, userID int64, req model.PEP3ExecutionPlanGenerateRequest, onDelta func(string) error) (any, error) {
	prepared, err := svc.preparePEP3ExecutionPlanGeneration(ctx, userID, req)
	if err != nil {
		return nil, err
	}
	switch prepared.PlanType {
	case "monthly":
		var result model.PEP3MonthlyPlanAIResult
		if err := callDeepSeekExecutionPlanStream(ctx, prepared.SystemPrompt, prepared.Payload, &result, onDelta); err != nil {
			return nil, err
		}
		normalized := normalizePEP3MonthlyExecutionPlan(result, prepared.SourcePlan, prepared.Target, prepared.CurrentTeacherName)
		return normalized, nil
	case "weekly":
		var result model.PEP3WeeklyPlanAIResult
		if err := callDeepSeekExecutionPlanStream(ctx, prepared.SystemPrompt, prepared.Payload, &result, onDelta); err != nil {
			return nil, err
		}
		normalized := normalizePEP3WeeklyExecutionPlan(result, prepared.SourcePlan, req.MonthlyPlan, prepared.Target, prepared.CurrentTeacherName)
		return normalized, nil
	default:
		return nil, errors.New("invalid execution plan type")
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
	if err := json.Unmarshal([]byte(extractJSONContent(content)), result); err != nil {
		return fmt.Errorf("parse DeepSeek execution plan JSON: %w", err)
	}
	return nil
}

func buildDeepSeekExecutionPlanRequestBody(systemPrompt string, payload pep3ExecutionPlanPromptPayload, stream bool) ([]byte, error) {
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
					systemPrompt,
					"必须输出严格JSON，不要Markdown，不要代码块，不要解释。",
					"不要使用空字符串占位生成目标或训练内容；没有内容的行不要输出。",
					"课程形式必须依据源计划保留或合理判断，常见值为个训、集体课。",
				}, "\n"),
			},
			{Role: "user", Content: string(payloadJSON)},
		},
		Temperature: 0.22,
		MaxTokens:   4096,
		ResponseFormat: map[string]string{
			"type": "json_object",
		},
		Thinking: &deepSeekThinking{Type: "disabled"},
		Stream:   stream,
	})
}

func callDeepSeekExecutionPlanStream(ctx context.Context, systemPrompt string, payload pep3ExecutionPlanPromptPayload, result any, onDelta func(string) error) error {
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
	requestBody, err := buildDeepSeekExecutionPlanRequestBody(systemPrompt, payload, true)
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
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("DeepSeek API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
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
			return fmt.Errorf("parse DeepSeek execution plan stream chunk: %w", err)
		}
		if chunk.Error != nil && strings.TrimSpace(chunk.Error.Message) != "" {
			return errors.New(chunk.Error.Message)
		}
		for _, choice := range chunk.Choices {
			if text := choice.Delta.Content; text != "" {
				content.WriteString(text)
				if onDelta != nil {
					if err := onDelta(text); err != nil {
						return err
					}
				}
			}
			if text := strings.TrimSpace(choice.Delta.ReasoningContent); text != "" {
				reasoning.WriteString(text)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	text := strings.TrimSpace(content.String())
	if text == "" {
		if strings.TrimSpace(reasoning.String()) != "" {
			return errors.New("DeepSeek API 只返回了 reasoning_content，没有返回最终JSON内容，请重试")
		}
		return errors.New("DeepSeek API returned empty content")
	}
	if err := json.Unmarshal([]byte(extractJSONContent(text)), result); err != nil {
		return fmt.Errorf("parse DeepSeek execution plan JSON: %w", err)
	}
	return nil
}

func normalizeExecutionSourceIEPPlan(plan model.PEP3IEPPlanAIResult, record model.AssessmentRecordDetailVO, currentTeacherName string, durationMonths int) model.PEP3IEPPlanAIResult {
	if durationMonths != 6 {
		durationMonths = 3
	}
	plan.Title = firstNonEmptyExportValue(strings.TrimSpace(plan.Title), iepPlanTitle(durationMonths))
	plan.Student.Name = firstNonEmptyExportValue(strings.TrimSpace(plan.Student.Name), strings.TrimSpace(record.StudentName))
	plan.Student.Gender = firstNonEmptyExportValue(strings.TrimSpace(plan.Student.Gender), strings.TrimSpace(record.StudentGender))
	plan.Student.BirthDate = firstNonEmptyExportValue(strings.TrimSpace(plan.Student.BirthDate), formatIEPPlanDate(record.BirthDate))
	plan.Meta.PlanDate = firstNonEmptyExportValue(strings.TrimSpace(plan.Meta.PlanDate), time.Now().Format("2006-01-02"))
	plan.Meta.Participant = firstNonEmptyExportValue(strings.TrimSpace(plan.Meta.Participant), strings.TrimSpace(currentTeacherName), strings.TrimSpace(record.ExaminerName))
	plan.Meta.Implementer = firstNonEmptyExportValue(strings.TrimSpace(plan.Meta.Implementer), strings.TrimSpace(currentTeacherName), strings.TrimSpace(record.ExaminerName))
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

func buildExecutionPlanTarget(sourcePlan model.PEP3IEPPlanAIResult, durationMonths, targetMonthIndex, targetWeekIndex int) pep3ExecutionPlanTarget {
	if durationMonths != 6 {
		durationMonths = 3
	}
	if targetMonthIndex < 1 {
		targetMonthIndex = 1
	}
	if targetMonthIndex > durationMonths {
		targetMonthIndex = durationMonths
	}
	startDate := firstNonEmptyExportValue(strings.TrimSpace(sourcePlan.Meta.StartDate), time.Now().Format("2006-01-02"))
	start := parseIEPPlanDateValue(startDate)
	if start.IsZero() {
		start = time.Now()
	}
	planStart := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.Local)
	monthStart := planStart.AddDate(0, targetMonthIndex-1, 0)
	monthEnd := time.Date(start.Year(), start.Month()+1, 0, 0, 0, 0, 0, time.Local)
	monthEnd = time.Date(monthStart.Year(), monthStart.Month()+1, 0, 0, 0, 0, 0, time.Local)
	daysInMonth := int(monthEnd.Sub(monthStart).Hours()/24) + 1
	weekCount := (daysInMonth + 5) / 6
	if targetWeekIndex < 1 {
		targetWeekIndex = 1
	}
	if targetWeekIndex > weekCount {
		targetWeekIndex = weekCount
	}
	weekStart := monthStart.AddDate(0, 0, (targetWeekIndex-1)*6)
	weekEnd := weekStart.AddDate(0, 0, 5)
	if weekEnd.After(monthEnd) {
		weekEnd = monthEnd
	}
	weekDates := make([]string, 0, 6)
	for current := weekStart; !current.After(weekEnd); current = current.AddDate(0, 0, 1) {
		weekDates = append(weekDates, current.Format("2006-01-02"))
	}
	return pep3ExecutionPlanTarget{
		MonthIndex:     targetMonthIndex,
		MonthLabel:     fmt.Sprintf("%d月", int(monthStart.Month())),
		WeekIndex:      targetWeekIndex,
		WeekLabel:      fmt.Sprintf("第%d周", targetWeekIndex),
		DurationMonths: durationMonths,
		StartDate:      monthStart.Format("2006-01-02"),
		EndDate:        monthEnd.Format("2006-01-02"),
		WeekDates:      weekDates,
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
	result.Meta.PlanDate = firstNonEmptyExportValue(strings.TrimSpace(result.Meta.PlanDate), time.Now().Format("2006-01-02"))
	result.Meta.Participant = firstNonEmptyExportValue(strings.TrimSpace(result.Meta.Participant), strings.TrimSpace(sourcePlan.Meta.Participant), currentTeacherName)
	result.Meta.Implementer = firstNonEmptyExportValue(strings.TrimSpace(result.Meta.Implementer), strings.TrimSpace(sourcePlan.Meta.Implementer), currentTeacherName)
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
		trainingItems := normalizePEP3MonthlyTrainingItems(row, target)
		rows = append(rows, model.PEP3MonthlyPlanRow{
			Domain:        firstNonEmptyExportValue(strings.TrimSpace(row.Domain), "综合康复"),
			LongGoal:      strings.TrimSpace(row.LongGoal),
			ShortGoal:     shortGoal,
			TrainingItems: trainingItems,
			CourseForm:    firstNonEmptyExportValue(normalizeIEPCourseForm(row.CourseForm), "个训"),
		})
	}
	result.Rows = rows
	return result
}

func normalizePEP3MonthlyTrainingItems(row model.PEP3MonthlyPlanRow, target pep3ExecutionPlanTarget) []model.PEP3MonthlyTrainingItem {
	items := make([]model.PEP3MonthlyTrainingItem, 0, len(row.TrainingItems))
	for _, item := range row.TrainingItems {
		content := strings.TrimSpace(item.Content)
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
	for index := range items {
		if strings.TrimSpace(items[index].StartEndDate) == "" {
			items[index].StartEndDate = firstNonEmptyExportValue(monthlyItemDateRange(target.StartDate, target.EndDate, index, len(items)), target.StartDate+" - "+target.EndDate)
		}
	}
	return items
}

func monthlyItemDateRange(startText, endText string, itemIndex, itemCount int) string {
	if itemCount <= 0 {
		itemCount = 1
	}
	if itemIndex < 0 {
		itemIndex = 0
	}
	if itemIndex >= itemCount {
		itemIndex = itemCount - 1
	}
	start := parseIEPPlanDateValue(startText)
	end := parseIEPPlanDateValue(endText)
	if start.IsZero() || end.IsZero() || end.Before(start) {
		return strings.TrimSpace(startText + " - " + endText)
	}
	totalDays := int(end.Sub(start).Hours()/24) + 1
	offsetStart := itemIndex * totalDays / itemCount
	offsetEnd := ((itemIndex + 1) * totalDays / itemCount) - 1
	if offsetEnd < offsetStart {
		offsetEnd = offsetStart
	}
	itemStart := start.AddDate(0, 0, offsetStart)
	itemEnd := start.AddDate(0, 0, offsetEnd)
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
		trainingDate = target.WeekDates[0] + " 至 " + target.WeekDates[len(target.WeekDates)-1]
	}
	result.TrainingDate = trainingDate
	result.Preparation = firstNonEmptyExportValue(strings.TrimSpace(result.Preparation), "依据"+sourceTitle+"准备训练材料、强化物和提示卡，明确本周训练目标与记录方式。")
	result.SourceTitle = firstNonEmptyExportValue(strings.TrimSpace(result.SourceTitle), sourceTitle)
	result.WeekDates = append([]string(nil), target.WeekDates...)
	if len(result.WeekDates) > 6 {
		result.WeekDates = result.WeekDates[:6]
	}
	for len(result.WeekDates) < 6 {
		result.WeekDates = append(result.WeekDates, "")
	}
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
