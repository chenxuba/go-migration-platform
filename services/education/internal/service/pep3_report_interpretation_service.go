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
	"net/http"
	"os"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/repository"
)

type pep3ReportInterpretationPromptPayload struct {
	Student       pep3ReportInterpretationStudent        `json:"student"`
	Assessment    pep3ReportInterpretationAssessment     `json:"assessment"`
	Summary       []string                               `json:"summary,omitempty"`
	Sections      []model.PEP3TemplateSection            `json:"sections"`
	Warnings      []string                               `json:"warnings,omitempty"`
	OutputRequest erxinReportInterpretationOutputRequest `json:"outputRequest"`
}

type pep3ReportInterpretationStudent struct {
	Name      string `json:"name"`
	Gender    string `json:"gender,omitempty"`
	BirthDate string `json:"birthDate,omitempty"`
	Age       string `json:"age,omitempty"`
}

type pep3ReportInterpretationAssessment struct {
	Date         string `json:"date,omitempty"`
	ScaleVersion string `json:"scaleVersion,omitempty"`
	DataStatus   string `json:"dataStatus,omitempty"`
}

func (svc *Service) GetPEP3ReportInterpretation(userID, recordID int64) (model.ERXinReportInterpretationVO, error) {
	if svc.repo == nil {
		return model.ERXinReportInterpretationVO{}, errors.New("assessment repository is not configured")
	}
	if recordID <= 0 {
		return model.ERXinReportInterpretationVO{}, errors.New("invalid assessment record id")
	}
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return model.ERXinReportInterpretationVO{}, err
	}
	report, err := svc.GetPEP3AssessmentReport(userID, recordID)
	if err != nil {
		return model.ERXinReportInterpretationVO{}, err
	}
	sourceHash := pep3ReportInterpretationSourceHash(report)
	cached, err := svc.repo.GetAssessmentReportInterpretation(context.Background(), instID, recordID, pep3ScaleCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.ERXinReportInterpretationVO{}, nil
		}
		return model.ERXinReportInterpretationVO{}, err
	}
	if strings.TrimSpace(cached.SourceHash) != sourceHash {
		return model.ERXinReportInterpretationVO{}, nil
	}
	normalized := normalizePEP3ReportInterpretation(cached.Interpretation, report, cached.Interpretation.GeneratedBy)
	if strings.TrimSpace(cached.Interpretation.GeneratedAt) != "" {
		normalized.GeneratedAt = cached.Interpretation.GeneratedAt
	}
	return normalized, nil
}

func (svc *Service) GeneratePEP3ReportInterpretation(ctx context.Context, userID, recordID int64) (model.ERXinReportInterpretationVO, error) {
	return svc.GeneratePEP3ReportInterpretationStream(ctx, userID, recordID, nil)
}

func (svc *Service) GeneratePEP3ReportInterpretationStream(ctx context.Context, userID, recordID int64, onDelta func(string) error) (model.ERXinReportInterpretationVO, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if svc.repo == nil {
		return model.ERXinReportInterpretationVO{}, errors.New("assessment repository is not configured")
	}
	if recordID <= 0 {
		return model.ERXinReportInterpretationVO{}, errors.New("invalid assessment record id")
	}
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return model.ERXinReportInterpretationVO{}, err
	}
	report, err := svc.GetPEP3AssessmentReport(userID, recordID)
	if err != nil {
		return model.ERXinReportInterpretationVO{}, err
	}
	payload := buildPEP3ReportInterpretationPromptPayload(report)
	result, err := callDeepSeekPEP3ReportInterpretationStream(ctx, payload, onDelta)
	if err != nil {
		fallback := buildRuleBasedPEP3ReportInterpretation(report)
		fallback.Notes = append(fallback.Notes, "AI解读生成失败，当前展示系统规则解读："+err.Error())
		if saveErr := svc.savePEP3ReportInterpretation(ctx, instID, userID, recordID, report, fallback); saveErr != nil {
			return model.ERXinReportInterpretationVO{}, saveErr
		}
		return fallback, nil
	}
	normalized := normalizePEP3ReportInterpretation(result, report, "ai")
	if err := svc.savePEP3ReportInterpretation(ctx, instID, userID, recordID, report, normalized); err != nil {
		return model.ERXinReportInterpretationVO{}, err
	}
	return normalized, nil
}

func (svc *Service) savePEP3ReportInterpretation(ctx context.Context, instID, userID, recordID int64, report model.PEP3ReportVO, interpretation model.ERXinReportInterpretationVO) error {
	return svc.repo.UpsertAssessmentReportInterpretation(ctx, repository.AssessmentReportInterpretationEntity{
		InstID:         instID,
		RecordID:       recordID,
		AssessmentCode: pep3ScaleCode,
		SourceHash:     pep3ReportInterpretationSourceHash(report),
		Interpretation: interpretation,
	}, userID)
}

func pep3ReportInterpretationSourceHash(report model.PEP3ReportVO) string {
	raw, err := json.Marshal(struct {
		RecordUpdatedTime *time.Time                  `json:"recordUpdatedTime,omitempty"`
		Sections          []model.PEP3TemplateSection `json:"sections"`
		DataStatus        string                      `json:"dataStatus,omitempty"`
		AssessmentDate    string                      `json:"assessmentDate,omitempty"`
		ExaminerName      string                      `json:"examinerName,omitempty"`
	}{
		RecordUpdatedTime: report.Record.UpdatedTime,
		Sections:          report.Sections,
		DataStatus:        strings.TrimSpace(report.DataStatus),
		AssessmentDate:    formatIEPPlanDate(report.Record.AssessmentDate),
		ExaminerName:      strings.TrimSpace(report.Record.ExaminerName),
	})
	if err != nil {
		raw = []byte(fmt.Sprintf("%+v", report.Sections))
	}
	sum := sha256.Sum256(raw)
	return fmt.Sprintf("%x", sum[:])
}

func buildPEP3ReportInterpretationPromptPayload(report model.PEP3ReportVO) pep3ReportInterpretationPromptPayload {
	return pep3ReportInterpretationPromptPayload{
		Student: pep3ReportInterpretationStudent{
			Name:      strings.TrimSpace(report.Record.StudentName),
			Gender:    strings.TrimSpace(report.Record.StudentGender),
			BirthDate: formatIEPPlanDate(report.Record.BirthDate),
			Age:       formatIEPPlanAge(report.Record.AgeYears, report.Record.AgeMonths, report.Record.AgeDays),
		},
		Assessment: pep3ReportInterpretationAssessment{
			Date:         formatIEPPlanDate(report.Record.AssessmentDate),
			ScaleVersion: strings.TrimSpace(report.ScaleVersion),
			DataStatus:   strings.TrimSpace(report.DataStatus),
		},
		Summary:  pep3ReportSummaryTextItems(report),
		Sections: append([]model.PEP3TemplateSection(nil), report.Sections...),
		Warnings: pep3ReportWarnings(report),
		OutputRequest: erxinReportInterpretationOutputRequest{
			RequiredSchema: "只输出JSON：title, summary, domainAnalysis[], suggestions[], notes[]。summary为1段综合解读；domainAnalysis为4-6条能力表现；suggestions为3-5条发展建议；notes为注意事项。",
			SafetyRules:    "必须基于PEP-3结果记录生成；不得更改原始分、标准分、百分比、发展年龄、等级、日期等事实数据；不得做医学诊断；不得使用确诊、治疗方案等表述；使用建议关注、可结合训练、建议复评等克制表达；如有warnings或数据复核提示，必须在notes中说明结果需结合临床观察和日常表现综合判断。",
		},
	}
}

func callDeepSeekPEP3ReportInterpretation(ctx context.Context, payload pep3ReportInterpretationPromptPayload) (model.ERXinReportInterpretationVO, error) {
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
	body, err := buildDeepSeekPEP3ReportInterpretationRequestBodyWithStream(payload, false)
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
		return model.ERXinReportInterpretationVO{}, fmt.Errorf("parse DeepSeek PEP3 report interpretation JSON: %w", err)
	}
	result.Model = deepSeekIEPPlanModel
	return result, nil
}

func buildDeepSeekPEP3ReportInterpretationRequestBodyWithStream(payload pep3ReportInterpretationPromptPayload, stream bool) ([]byte, error) {
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
					"任务是把PEP-3结果记录转换为专业、克制、适合家长和老师阅读的报告解读。",
					"必须输出严格JSON，不要Markdown，不要代码块，不要解释。",
					"不得修改输入中的分数、百分比、发展年龄、等级、日期等事实。",
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
		Stream:   stream,
	})
}

func callDeepSeekPEP3ReportInterpretationStream(ctx context.Context, payload pep3ReportInterpretationPromptPayload, onDelta func(string) error) (model.ERXinReportInterpretationVO, error) {
	if onDelta == nil {
		return callDeepSeekPEP3ReportInterpretation(ctx, payload)
	}
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
	body, err := buildDeepSeekPEP3ReportInterpretationRequestBodyWithStream(payload, true)
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
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		responseBody, _ := io.ReadAll(resp.Body)
		return model.ERXinReportInterpretationVO{}, fmt.Errorf("DeepSeek API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
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
			return model.ERXinReportInterpretationVO{}, fmt.Errorf("parse DeepSeek stream chunk: %w", err)
		}
		if chunk.Error != nil && strings.TrimSpace(chunk.Error.Message) != "" {
			return model.ERXinReportInterpretationVO{}, errors.New(chunk.Error.Message)
		}
		for _, choice := range chunk.Choices {
			if text := choice.Delta.Content; text != "" {
				content.WriteString(text)
				if err := onDelta(text); err != nil {
					return model.ERXinReportInterpretationVO{}, err
				}
			}
			if text := strings.TrimSpace(choice.Delta.ReasoningContent); text != "" {
				reasoning.WriteString(text)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return model.ERXinReportInterpretationVO{}, err
	}

	text := strings.TrimSpace(content.String())
	if text == "" {
		if strings.TrimSpace(reasoning.String()) != "" {
			return model.ERXinReportInterpretationVO{}, errors.New("DeepSeek API 只返回了 reasoning_content，没有返回最终JSON内容，请重试")
		}
		return model.ERXinReportInterpretationVO{}, errors.New("DeepSeek API returned empty content")
	}
	var result model.ERXinReportInterpretationVO
	if err := json.Unmarshal([]byte(extractJSONContent(text)), &result); err != nil {
		return model.ERXinReportInterpretationVO{}, fmt.Errorf("parse DeepSeek PEP3 report interpretation JSON: %w", err)
	}
	result.Model = deepSeekIEPPlanModel
	return result, nil
}

func buildRuleBasedPEP3ReportInterpretation(report model.PEP3ReportVO) model.ERXinReportInterpretationVO {
	summaryItems := pep3ReportSummaryTextItems(report)
	summary := strings.Join(summaryItems, "；")
	if strings.TrimSpace(summary) == "" {
		summary = "本次PEP-3评估已形成发展量表、适应不良行为、照顾者报告及合成分数结果，可作为后续教学目标和阶段复评的参考。"
	}
	domainAnalysis := compactNonEmptyStrings([]string{
		pep3ScaleSectionBrief(report, "development_scores", "发展量表"),
		pep3ScaleSectionBrief(report, "behavior_scores", "适应不良行为"),
		pep3ScaleSectionBrief(report, "caregiver_scores", "照顾者报告"),
		pep3CompositeSectionBrief(report),
	})
	if len(domainAnalysis) == 0 {
		domainAnalysis = []string{"当前记录已保存PEP-3结构化评分结果，建议结合各副测验原始分、发展年龄、百分比级数和等级进行综合理解。"}
	}
	warnings := pep3ReportWarnings(report)
	suggestions := []string{
		"建议优先围绕相对薄弱的沟通、动作或行为适应相关能力设置可观察的小目标，并在课程中持续记录表现。",
		"建议保留儿童相对优势领域的活动机会，用优势能力带动薄弱能力的参与和泛化。",
		"建议结合课堂观察、家庭反馈和阶段复评结果动态调整训练重点，避免仅依据单次分数做结论。",
	}
	notes := []string{"本解读基于当前PEP-3结果记录自动生成，仅用于评估沟通和训练计划参考。"}
	if len(warnings) > 0 {
		notes = append(notes, "当前记录存在数据或换算复核提示，结果需谨慎参考："+strings.Join(warnings, "；"))
	}
	return model.ERXinReportInterpretationVO{
		Title:          "PEP-3报告解读",
		GeneratedBy:    "rule",
		GeneratedAt:    time.Now().Format(time.RFC3339),
		Summary:        summary,
		DomainAnalysis: domainAnalysis,
		Suggestions:    suggestions,
		Notes:          notes,
	}
}

func normalizePEP3ReportInterpretation(result model.ERXinReportInterpretationVO, report model.PEP3ReportVO, generatedBy string) model.ERXinReportInterpretationVO {
	result.Title = nonEmptyString(result.Title, "PEP-3报告解读")
	result.GeneratedBy = nonEmptyString(result.GeneratedBy, generatedBy)
	result.GeneratedAt = time.Now().Format(time.RFC3339)
	result.Summary = strings.TrimSpace(result.Summary)
	if result.Summary == "" {
		result.Summary = buildRuleBasedPEP3ReportInterpretation(report).Summary
	}
	result.DomainAnalysis = compactNonEmptyStrings(result.DomainAnalysis)
	result.Suggestions = compactNonEmptyStrings(result.Suggestions)
	result.Notes = compactNonEmptyStrings(result.Notes)
	if len(result.DomainAnalysis) == 0 || len(result.Suggestions) == 0 {
		fallback := buildRuleBasedPEP3ReportInterpretation(report)
		if len(result.DomainAnalysis) == 0 {
			result.DomainAnalysis = fallback.DomainAnalysis
		}
		if len(result.Suggestions) == 0 {
			result.Suggestions = fallback.Suggestions
		}
	}
	if len(result.Notes) == 0 {
		result.Notes = []string{"报告解读由PEP-3结构化评分结果生成，不替代医学诊断。"}
	}
	return result
}

func pep3ReportSummaryTextItems(report model.PEP3ReportVO) []string {
	for _, section := range report.Sections {
		if section.SectionCode == "summary" {
			return compactNonEmptyStrings(section.TextItems)
		}
	}
	return nil
}

func pep3ReportWarnings(report model.PEP3ReportVO) []string {
	values := make([]string, 0)
	if strings.TrimSpace(report.DataStatus) != "" {
		values = append(values, report.DataStatus)
	}
	for _, section := range report.Sections {
		if section.SectionCode == "warnings" {
			values = append(values, section.TextItems...)
		}
		if section.Table == nil {
			continue
		}
		for _, row := range section.Table.Rows {
			values = append(values, pep3StringSliceFromAny(row["warnings"])...)
		}
	}
	return uniqueNonEmptyStrings(values)
}

func pep3ScaleSectionBrief(report model.PEP3ReportVO, sectionCode, fallbackTitle string) string {
	section := pep3ReportSection(report, sectionCode)
	if section == nil || section.Table == nil || len(section.Table.Rows) == 0 {
		return ""
	}
	parts := make([]string, 0, len(section.Table.Rows))
	for _, row := range section.Table.Rows {
		name := pep3StringFromAny(row["scaleName"])
		code := pep3StringFromAny(row["scaleCode"])
		rawScore := pep3StringFromAny(row["rawScore"])
		developmentAge := pep3StringFromAny(row["developmentAge"])
		percentile := pep3StringFromAny(row["percentileRank"])
		level := pep3StringFromAny(row["level"])
		item := nonEmptyString(name, code)
		details := compactNonEmptyStrings([]string{
			pep3LabelValue("原始分", rawScore),
			pep3LabelValue("发展年龄", developmentAge),
			pep3LabelValue("百分比级数", percentile),
			pep3LabelValue("等级", level),
		})
		if len(details) > 0 {
			item += "（" + strings.Join(details, "，") + "）"
		}
		parts = append(parts, item)
	}
	title := nonEmptyString(section.Title, fallbackTitle)
	return title + "：" + strings.Join(compactNonEmptyStrings(parts), "；") + "。"
}

func pep3CompositeSectionBrief(report model.PEP3ReportVO) string {
	section := pep3ReportSection(report, "composite_scores")
	if section == nil || section.Table == nil || len(section.Table.Rows) == 0 {
		return ""
	}
	parts := make([]string, 0, len(section.Table.Rows))
	for _, row := range section.Table.Rows {
		name := pep3StringFromAny(row["compositeName"])
		percentile := pep3StringFromAny(row["percentileRank"])
		level := pep3StringFromAny(row["level"])
		developmentAge := pep3StringFromAny(row["developmentAge"])
		details := compactNonEmptyStrings([]string{
			pep3LabelValue("百分比级数", percentile),
			pep3LabelValue("等级", level),
			pep3LabelValue("发展年龄", developmentAge),
		})
		item := nonEmptyString(name, "合成项目")
		if len(details) > 0 {
			item += "（" + strings.Join(details, "，") + "）"
		}
		parts = append(parts, item)
	}
	return "合成分数：" + strings.Join(compactNonEmptyStrings(parts), "；") + "。"
}

func pep3ReportSection(report model.PEP3ReportVO, sectionCode string) *model.PEP3TemplateSection {
	for index := range report.Sections {
		if report.Sections[index].SectionCode == sectionCode {
			return &report.Sections[index]
		}
	}
	return nil
}

func pep3LabelValue(label, value string) string {
	value = strings.TrimSpace(value)
	if value == "" || value == "--" {
		return ""
	}
	return label + value
}

func pep3StringFromAny(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(v)
	case fmt.Stringer:
		return strings.TrimSpace(v.String())
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case float64:
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%.1f", v)
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", value))
	}
}

func pep3StringSliceFromAny(value any) []string {
	switch v := value.(type) {
	case nil:
		return nil
	case []string:
		return compactNonEmptyStrings(v)
	case []any:
		values := make([]string, 0, len(v))
		for _, item := range v {
			values = append(values, pep3StringFromAny(item))
		}
		return compactNonEmptyStrings(values)
	default:
		text := pep3StringFromAny(value)
		if text == "" {
			return nil
		}
		return []string{text}
	}
}
