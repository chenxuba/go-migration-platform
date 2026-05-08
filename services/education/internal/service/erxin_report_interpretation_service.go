package service

import (
	"bytes"
	"context"
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
)

type erxinReportInterpretationPromptPayload struct {
	Student       erxinReportInterpretationStudent       `json:"student"`
	Assessment    erxinReportInterpretationAssessment    `json:"assessment"`
	Summary       model.ERXinReportSummary               `json:"summary"`
	DomainRows    []model.ERXinReportDomainRow           `json:"domainRows"`
	Warnings      []string                               `json:"warnings,omitempty"`
	OutputRequest erxinReportInterpretationOutputRequest `json:"outputRequest"`
}

type erxinReportInterpretationStudent struct {
	Name      string `json:"name"`
	Gender    string `json:"gender,omitempty"`
	BirthDate string `json:"birthDate,omitempty"`
	Age       string `json:"age,omitempty"`
}

type erxinReportInterpretationAssessment struct {
	Date           string `json:"date,omitempty"`
	ScaleVersion   string `json:"scaleVersion,omitempty"`
	SourceStandard string `json:"sourceStandard,omitempty"`
}

type erxinReportInterpretationOutputRequest struct {
	RequiredSchema string `json:"requiredSchema"`
	SafetyRules    string `json:"safetyRules"`
}

func (svc *Service) GenerateERXinReportInterpretation(ctx context.Context, userID, recordID int64) (model.ERXinReportInterpretationVO, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if recordID <= 0 {
		return model.ERXinReportInterpretationVO{}, errors.New("invalid assessment record id")
	}
	report, err := svc.GetERXinAssessmentReport(userID, recordID)
	if err != nil {
		return model.ERXinReportInterpretationVO{}, err
	}
	payload := buildERXinReportInterpretationPromptPayload(report)
	result, err := callDeepSeekERXinReportInterpretation(ctx, payload)
	if err != nil {
		fallback := buildRuleBasedERXinReportInterpretation(report)
		fallback.Notes = append(fallback.Notes, "AI解读生成失败，当前展示系统规则解读："+err.Error())
		return fallback, nil
	}
	return normalizeERXinReportInterpretation(result, report, "ai"), nil
}

func buildERXinReportInterpretationPromptPayload(report model.ERXinReportVO) erxinReportInterpretationPromptPayload {
	return erxinReportInterpretationPromptPayload{
		Student: erxinReportInterpretationStudent{
			Name:      strings.TrimSpace(report.Record.StudentName),
			Gender:    erxinBlankDash(report.Record.StudentGender),
			BirthDate: erxinFormatDate(report.Record.BirthDate),
			Age:       erxinActualAgeSummaryText(report.Record),
		},
		Assessment: erxinReportInterpretationAssessment{
			Date:           erxinFormatDate(report.Record.AssessmentDate),
			ScaleVersion:   strings.TrimSpace(report.ScaleVersion),
			SourceStandard: strings.TrimSpace(report.SourceStandard),
		},
		Summary:    report.Summary,
		DomainRows: append([]model.ERXinReportDomainRow(nil), report.DomainRows...),
		Warnings:   append([]string(nil), report.Warnings...),
		OutputRequest: erxinReportInterpretationOutputRequest{
			RequiredSchema: "只输出JSON：title, summary, domainAnalysis[], suggestions[], notes[]。summary为1段综合解读；domainAnalysis为3-5条能区表现；suggestions为3-5条发展建议；notes为注意事项。",
			SafetyRules:    "不得更改任何DQ、智龄、等级、日期等事实数据；不得做医学诊断；不得使用确诊、治疗方案等表述；使用建议关注、可结合训练、建议复评等克制表达；如有warnings或评分不完整，必须在notes提示结果需结合临床观察和日常表现综合判断。",
		},
	}
}

func callDeepSeekERXinReportInterpretation(ctx context.Context, payload erxinReportInterpretationPromptPayload) (model.ERXinReportInterpretationVO, error) {
	apiKey := strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY"))
	if apiKey == "" {
		apiKey = deepSeekIEPPlanFallbackAPIKey
	}
	if apiKey == "" {
		return model.ERXinReportInterpretationVO{}, errors.New("DEEPSEEK_API_KEY is not configured")
	}
	endpoint := strings.TrimSpace(os.Getenv("DEEPSEEK_API_BASE_URL"))
	if endpoint == "" {
		endpoint = deepSeekIEPPlanDefaultURL
	}
	body, err := buildDeepSeekERXinReportInterpretationRequestBody(payload)
	if err != nil {
		return model.ERXinReportInterpretationVO{}, err
	}
	requestCtx, cancel := context.WithTimeout(ctx, deepSeekIEPPlanTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return model.ERXinReportInterpretationVO{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := (&http.Client{Timeout: deepSeekIEPPlanTimeout + 5*time.Second}).Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(requestCtx.Err(), context.DeadlineExceeded) {
			return model.ERXinReportInterpretationVO{}, fmt.Errorf("DeepSeek API 生成超时（%d秒），请稍后重试", int(deepSeekIEPPlanTimeout.Seconds()))
		}
		return model.ERXinReportInterpretationVO{}, err
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.ERXinReportInterpretationVO{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return model.ERXinReportInterpretationVO{}, fmt.Errorf("DeepSeek API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}
	var chatResponse deepSeekChatResponse
	if err := json.Unmarshal(responseBody, &chatResponse); err != nil {
		return model.ERXinReportInterpretationVO{}, err
	}
	if chatResponse.Error != nil && strings.TrimSpace(chatResponse.Error.Message) != "" {
		return model.ERXinReportInterpretationVO{}, errors.New(chatResponse.Error.Message)
	}
	if len(chatResponse.Choices) == 0 {
		return model.ERXinReportInterpretationVO{}, errors.New("DeepSeek API returned empty choices")
	}
	content := strings.TrimSpace(chatResponse.Choices[0].Message.Content)
	if content == "" {
		return model.ERXinReportInterpretationVO{}, errors.New("DeepSeek API returned empty content")
	}
	var result model.ERXinReportInterpretationVO
	if err := json.Unmarshal([]byte(extractJSONContent(content)), &result); err != nil {
		return model.ERXinReportInterpretationVO{}, fmt.Errorf("parse DeepSeek ERXin report interpretation JSON: %w", err)
	}
	result.Model = deepSeekIEPPlanModel
	return result, nil
}

func buildDeepSeekERXinReportInterpretationRequestBody(payload erxinReportInterpretationPromptPayload) ([]byte, error) {
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
					"你是儿童发育评估报告解读助手。",
					"任务是把儿心量表-II的结构化评分结果转换为专业、克制、适合家长和老师阅读的报告解读。",
					"必须输出严格JSON，不要Markdown，不要代码块，不要解释。",
					"不得修改输入中的分数、智龄、DQ、等级、日期等事实。",
					"不得做医学诊断，不得输出治疗方案。",
				}, "\n"),
			},
			{Role: "user", Content: string(payloadJSON)},
		},
		Temperature: 0.2,
		MaxTokens:   2048,
		ResponseFormat: map[string]string{
			"type": "json_object",
		},
		Thinking: &deepSeekThinking{Type: "disabled"},
	})
}

func buildRuleBasedERXinReportInterpretation(report model.ERXinReportVO) model.ERXinReportInterpretationVO {
	rows := append([]model.ERXinReportDomainRow(nil), report.DomainRows...)
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i].DQ > rows[j].DQ
	})
	strong := firstNonEmptyDomainName(rows, true)
	weak := firstNonEmptyDomainName(rows, false)
	summary := fmt.Sprintf(
		"本次测评显示，儿童全量表发育商 DQ 为 %s，平均智龄为 %s，整体评价等级为%s。五大能区中，%s表现相对较好，%s需要在后续训练和日常观察中重点关注。",
		erxinReportInterpretationDQText(report.Summary.DQ),
		nonEmptyString(report.Summary.MeanMentalAgeMonthsText, erxinMonthText(report.Summary.MeanMentalAgeMonths)),
		nonEmptyString(report.Summary.Level, "待结合完整测评判断"),
		strong,
		weak,
	)
	analysis := make([]string, 0, len(report.DomainRows))
	for _, row := range report.DomainRows {
		analysis = append(analysis, fmt.Sprintf(
			"%s：智龄%s，DQ %s，等级%s。",
			nonEmptyString(row.DomainName, row.DomainCode),
			nonEmptyString(row.MentalAgeMonthsText, erxinMonthText(row.MentalAgeMonths)),
			erxinReportInterpretationDQText(row.DQ),
			nonEmptyString(row.Level, "待判断"),
		))
	}
	suggestions := []string{
		fmt.Sprintf("建议围绕%s相关能力设计持续、可观察的小目标训练，并结合课堂和家庭场景进行泛化。", weak),
		"建议保留儿童相对优势能区的活动机会，用优势能力带动薄弱能力发展。",
		"建议结合日常行为观察、教师记录和家长反馈综合判断，并按机构计划进行阶段复评。",
	}
	notes := []string{"本解读基于当前儿心量表-II结构化评分结果自动生成，仅用于评估沟通和训练计划参考。"}
	if len(report.Warnings) > 0 || !report.Summary.Complete {
		notes = append(notes, "当前记录存在缺测或评分完整性提示，结果需谨慎参考："+strings.Join(report.Warnings, "；"))
	}
	return model.ERXinReportInterpretationVO{
		Title:          "报告解读",
		GeneratedBy:    "rule",
		GeneratedAt:    time.Now().Format(time.RFC3339),
		Summary:        summary,
		DomainAnalysis: analysis,
		Suggestions:    suggestions,
		Notes:          notes,
	}
}

func normalizeERXinReportInterpretation(result model.ERXinReportInterpretationVO, report model.ERXinReportVO, generatedBy string) model.ERXinReportInterpretationVO {
	result.Title = nonEmptyString(result.Title, "报告解读")
	result.GeneratedBy = nonEmptyString(result.GeneratedBy, generatedBy)
	result.GeneratedAt = time.Now().Format(time.RFC3339)
	result.Summary = strings.TrimSpace(result.Summary)
	if result.Summary == "" {
		result.Summary = buildRuleBasedERXinReportInterpretation(report).Summary
	}
	result.DomainAnalysis = compactNonEmptyStrings(result.DomainAnalysis)
	result.Suggestions = compactNonEmptyStrings(result.Suggestions)
	result.Notes = compactNonEmptyStrings(result.Notes)
	if len(result.DomainAnalysis) == 0 || len(result.Suggestions) == 0 {
		fallback := buildRuleBasedERXinReportInterpretation(report)
		if len(result.DomainAnalysis) == 0 {
			result.DomainAnalysis = fallback.DomainAnalysis
		}
		if len(result.Suggestions) == 0 {
			result.Suggestions = fallback.Suggestions
		}
	}
	if len(result.Notes) == 0 {
		result.Notes = []string{"报告解读由结构化评分结果生成，不替代医学诊断。"}
	}
	return result
}

func compactNonEmptyStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		text := strings.TrimSpace(value)
		if text != "" {
			out = append(out, text)
		}
	}
	return out
}

func firstNonEmptyDomainName(rows []model.ERXinReportDomainRow, strongest bool) string {
	if len(rows) == 0 {
		return "各能区"
	}
	if strongest {
		for _, row := range rows {
			if name := nonEmptyString(row.DomainName, row.DomainCode); name != "" {
				return name
			}
		}
	}
	for index := len(rows) - 1; index >= 0; index-- {
		if name := nonEmptyString(rows[index].DomainName, rows[index].DomainCode); name != "" {
			return name
		}
	}
	return "相关能区"
}

func erxinReportInterpretationDQText(value float64) string {
	if value <= 0 || math.IsNaN(value) || math.IsInf(value, 0) {
		return "待判断"
	}
	return erxinFloatText(value)
}
