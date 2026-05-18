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
	"sort"
	"strings"
	"time"

	"go-migration-platform/pkg/autismdevscore"
	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/repository"
)

const autismDevReportInterpretationCode = "AUTISMDEV_REPORT_INTERPRETATION"

func (svc *Service) GetAutismDevReportInterpretation(userID, recordID int64) (model.ERXinReportInterpretationVO, error) {
	if svc.repo == nil {
		return model.ERXinReportInterpretationVO{}, errors.New("assessment repository is not configured")
	}
	instID, record, score, data, itemScores, err := svc.autismDevResultAnalysisContext(userID, recordID)
	if err != nil {
		return model.ERXinReportInterpretationVO{}, err
	}
	sourceHash := autismDevReportInterpretationSourceHash(record)
	var cachedRaw json.RawMessage
	cached, err := svc.repo.GetAssessmentReportInterpretationJSON(context.Background(), instID, recordID, autismDevReportInterpretationCode, &cachedRaw)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.ERXinReportInterpretationVO{}, nil
		}
		return model.ERXinReportInterpretationVO{}, err
	}
	if strings.TrimSpace(cached.SourceHash) != sourceHash {
		return model.ERXinReportInterpretationVO{}, nil
	}
	interpretation, err := parseAutismDevReportInterpretationJSON(string(cachedRaw))
	if err != nil {
		return model.ERXinReportInterpretationVO{}, err
	}
	if strings.TrimSpace(interpretation.Model) == "" {
		interpretation.Model = cached.Model
	}
	if strings.TrimSpace(interpretation.GeneratedBy) == "" {
		interpretation.GeneratedBy = cached.GeneratedBy
	}
	normalized := normalizeAutismDevReportInterpretation(interpretation, record, score, data, itemScores, interpretation.GeneratedBy)
	if strings.TrimSpace(interpretation.GeneratedAt) != "" {
		normalized.GeneratedAt = interpretation.GeneratedAt
	}
	return normalized, nil
}

func (svc *Service) GenerateAutismDevReportInterpretation(ctx context.Context, userID, recordID int64) (model.ERXinReportInterpretationVO, error) {
	return svc.GenerateAutismDevReportInterpretationStream(ctx, userID, recordID, nil)
}

func (svc *Service) GenerateAutismDevReportInterpretationStream(ctx context.Context, userID, recordID int64, onDelta func(string) error) (model.ERXinReportInterpretationVO, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if svc.repo == nil {
		return model.ERXinReportInterpretationVO{}, errors.New("assessment repository is not configured")
	}
	instID, record, score, data, itemScores, err := svc.autismDevResultAnalysisContext(userID, recordID)
	if err != nil {
		return model.ERXinReportInterpretationVO{}, err
	}
	payload := buildAutismDevReportInterpretationPromptPayload(record, score, data, itemScores)
	result, err := callDeepSeekAutismDevReportInterpretationStream(ctx, payload, onDelta)
	if err != nil {
		fallback := buildRuleBasedAutismDevReportInterpretation(record, score, data, itemScores)
		fallback.Notes = append(fallback.Notes, "AI解读生成失败，当前展示系统规则解读："+err.Error())
		if saveErr := svc.saveAutismDevReportInterpretation(ctx, instID, userID, recordID, record, fallback); saveErr != nil {
			return model.ERXinReportInterpretationVO{}, saveErr
		}
		return fallback, nil
	}
	normalized := normalizeAutismDevReportInterpretation(result, record, score, data, itemScores, "ai")
	if err := svc.saveAutismDevReportInterpretation(ctx, instID, userID, recordID, record, normalized); err != nil {
		return model.ERXinReportInterpretationVO{}, err
	}
	return normalized, nil
}

func (svc *Service) saveAutismDevReportInterpretation(ctx context.Context, instID, userID, recordID int64, record model.AssessmentRecordDetailVO, interpretation model.ERXinReportInterpretationVO) error {
	return svc.repo.UpsertAssessmentReportInterpretation(ctx, repository.AssessmentReportInterpretationEntity{
		InstID:         instID,
		RecordID:       recordID,
		AssessmentCode: autismDevReportInterpretationCode,
		SourceHash:     autismDevReportInterpretationSourceHash(record),
		Interpretation: interpretation,
	}, userID)
}

func autismDevReportInterpretationSourceHash(record model.AssessmentRecordDetailVO) string {
	raw, err := json.Marshal(struct {
		Input          json.RawMessage `json:"input,omitempty"`
		Result         json.RawMessage `json:"result,omitempty"`
		DataStatus     string          `json:"dataStatus,omitempty"`
		Remark         string          `json:"remark,omitempty"`
		AssessmentDate string          `json:"assessmentDate,omitempty"`
		ExaminerName   string          `json:"examinerName,omitempty"`
	}{
		Input:          record.InputJSON,
		Result:         record.ResultJSON,
		DataStatus:     strings.TrimSpace(record.DataStatus),
		Remark:         strings.TrimSpace(record.Remark),
		AssessmentDate: formatIEPPlanDate(record.AssessmentDate),
		ExaminerName:   strings.TrimSpace(record.ExaminerName),
	})
	if err != nil {
		raw = []byte(fmt.Sprintf("%+v", record.ResultJSON))
	}
	sum := sha256.Sum256(raw)
	return fmt.Sprintf("%x", sum[:])
}

func buildAutismDevReportInterpretationPromptPayload(record model.AssessmentRecordDetailVO, result autismdevscore.AssessmentResult, data autismDevStaticData, itemScores map[int]string) autismDevResultAnalysisPromptPayload {
	payload := buildAutismDevResultAnalysisPromptPayload(record, result, data, itemScores)
	payload.Domains = autismDevIEPPlanPromptDomains(result, data, itemScores)
	payload.OutputRequest = erxinReportInterpretationOutputRequest{
		RequiredSchema: "只输出JSON：title, summary, domainAnalysis[], suggestions[], notes[]。domainAnalysis必须是字符串数组，不允许对象数组；每条格式为“领域名：一句话”，最多60个中文字符，只写1个已会项目和1个下一步关注点，尽量具体到题目或行为。summary最多80个中文字符；suggestions为3-4条，每条最多45个中文字符；notes为1-2条。禁止把domain、analysis、name、content等字段名写进数组内容。",
		SafetyRules:    "必须基于孤独症儿童发展评估表的结构化题目得分生成；不得更改P、E、F、X、A、M、S计数、年龄段、日期等事实；不得做医学诊断；不得使用确诊、治疗方案等表述；不要写复杂长段落；如有X项、缺测或数据复核提示，必须在notes中说明结果需结合临床观察和日常表现综合判断。",
	}
	return payload
}

func callDeepSeekAutismDevReportInterpretation(ctx context.Context, payload autismDevResultAnalysisPromptPayload) (model.ERXinReportInterpretationVO, error) {
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
	body, err := buildDeepSeekAutismDevReportInterpretationRequestBody(payload, false)
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
	result, err := parseAutismDevReportInterpretationJSON(extractJSONContent(content))
	if err != nil {
		return model.ERXinReportInterpretationVO{}, fmt.Errorf("parse DeepSeek AutismDev report interpretation JSON: %w", err)
	}
	result.Model = deepSeekIEPPlanModel
	return result, nil
}

func callDeepSeekAutismDevReportInterpretationStream(ctx context.Context, payload autismDevResultAnalysisPromptPayload, onDelta func(string) error) (model.ERXinReportInterpretationVO, error) {
	if onDelta == nil {
		return callDeepSeekAutismDevReportInterpretation(ctx, payload)
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
	body, err := buildDeepSeekAutismDevReportInterpretationRequestBody(payload, true)
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
		if isDeepSeekDeadlineExceeded(err, requestCtx) {
			return model.ERXinReportInterpretationVO{}, deepSeekTimeoutError()
		}
		return model.ERXinReportInterpretationVO{}, err
	}
	text := strings.TrimSpace(content.String())
	if text == "" {
		if strings.TrimSpace(reasoning.String()) != "" {
			return model.ERXinReportInterpretationVO{}, errors.New("DeepSeek API 只返回了 reasoning_content，没有返回最终JSON内容，请重试")
		}
		return model.ERXinReportInterpretationVO{}, errors.New("DeepSeek API returned empty content")
	}
	result, err := parseAutismDevReportInterpretationJSON(extractJSONContent(text))
	if err != nil {
		return model.ERXinReportInterpretationVO{}, fmt.Errorf("parse DeepSeek AutismDev report interpretation JSON: %w", err)
	}
	result.Model = deepSeekIEPPlanModel
	return result, nil
}

func parseAutismDevReportInterpretationJSON(raw string) (model.ERXinReportInterpretationVO, error) {
	content := strings.TrimSpace(raw)
	if content == "" {
		return model.ERXinReportInterpretationVO{}, errors.New("empty report interpretation JSON")
	}
	var data map[string]any
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return model.ERXinReportInterpretationVO{}, err
	}
	return model.ERXinReportInterpretationVO{
		Title:          autismDevReportInterpretationValueText(data["title"], false),
		Model:          autismDevReportInterpretationValueText(data["model"], false),
		GeneratedBy:    autismDevReportInterpretationValueText(data["generatedBy"], false),
		GeneratedAt:    autismDevReportInterpretationValueText(data["generatedAt"], false),
		Summary:        autismDevReportInterpretationValueText(data["summary"], false),
		DomainAnalysis: autismDevReportInterpretationValueList(data["domainAnalysis"], true),
		Suggestions:    autismDevReportInterpretationValueList(data["suggestions"], false),
		Notes:          autismDevReportInterpretationValueList(data["notes"], false),
	}, nil
}

func autismDevReportInterpretationValueList(value any, withDomainPrefix bool) []string {
	var values []string
	switch typed := value.(type) {
	case []any:
		values = make([]string, 0, len(typed))
		for _, item := range typed {
			if text := autismDevReportInterpretationValueText(item, withDomainPrefix); text != "" {
				values = append(values, text)
			}
		}
	case []string:
		values = append(values, typed...)
	case string:
		values = append(values, typed)
	case map[string]any:
		if text := autismDevReportInterpretationValueText(typed, withDomainPrefix); text != "" {
			values = append(values, text)
		}
	}
	return values
}

func autismDevReportInterpretationValueText(value any, withDomainPrefix bool) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(typed)
	case json.Number:
		return typed.String()
	case float64:
		return strings.TrimSpace(fmt.Sprintf("%v", typed))
	case bool:
		if typed {
			return "true"
		}
		return "false"
	case map[string]any:
		domain := firstAutismDevReportInterpretationMapText(typed, "domain", "domainName", "name", "title")
		body := firstAutismDevReportInterpretationMapText(typed, "analysis", "text", "content", "summary", "description", "value", "suggestion", "note")
		if body == "" {
			for key, item := range typed {
				if autismDevReportInterpretationIsFieldName(key) {
					continue
				}
				if text := autismDevReportInterpretationValueText(item, false); text != "" && !autismDevReportInterpretationIsFieldName(text) {
					body = text
					break
				}
			}
		}
		if body == "" {
			return ""
		}
		if withDomainPrefix && domain != "" && !strings.HasPrefix(body, domain+"：") && !strings.HasPrefix(body, domain+":") {
			return domain + "：" + body
		}
		return body
	case []any:
		parts := make([]string, 0, len(typed))
		for _, item := range typed {
			if text := autismDevReportInterpretationValueText(item, withDomainPrefix); text != "" {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, "；")
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", typed))
	}
}

func firstAutismDevReportInterpretationMapText(values map[string]any, keys ...string) string {
	for _, key := range keys {
		if text := autismDevReportInterpretationValueText(values[key], false); text != "" && !autismDevReportInterpretationIsFieldName(text) {
			return text
		}
	}
	return ""
}

func buildDeepSeekAutismDevReportInterpretationRequestBody(payload autismDevResultAnalysisPromptPayload, stream bool) ([]byte, error) {
	return buildDeepSeekIEPPlanRequestBodyWithPrompt(payload, strings.Join([]string{
		"你是儿童康复机构的评估报告解读助手。",
		"任务是把孤独症儿童发展评估表8大项结构化评分结果转换为专业、克制、适合家长和老师阅读的报告解读。",
		"写法要短，避免复杂长段落；每个领域尽量写成“已会什么 + 还要练什么 + 下一步怎么练”，并尽量点到1个具体题目或行为。",
		"domainAnalysis必须按输入domains顺序写成字符串数组，覆盖已测领域；不要输出对象数组，不要输出domain、analysis等字段名。",
		"必须输出严格JSON，不要Markdown，不要代码块，不要解释。",
		"不得修改输入中的分数、计数、年龄段、日期等事实。",
		"不得做医学诊断，不得输出治疗方案。",
	}, "\n"), stream)
}

func buildRuleBasedAutismDevReportInterpretation(record model.AssessmentRecordDetailVO, score autismdevscore.AssessmentResult, data autismDevStaticData, itemScores map[int]string) model.ERXinReportInterpretationVO {
	domains := autismDevIEPPlanPromptDomains(score, data, itemScores)
	if len(domains) == 0 {
		domains = autismDevResultAnalysisPromptDomains(record, score, data, itemScores)
	}
	weakDomains := append([]autismDevResultAnalysisPromptDomain(nil), domains...)
	sort.SliceStable(weakDomains, func(i, j int) bool {
		return autismDevReportInterpretationConcernScore(weakDomains[i]) > autismDevReportInterpretationConcernScore(weakDomains[j])
	})
	primaryConcern := "相关领域"
	if len(weakDomains) > 0 && strings.TrimSpace(weakDomains[0].DomainName) != "" {
		primaryConcern = strings.TrimSpace(weakDomains[0].DomainName)
	}
	summary := fmt.Sprintf("本次结果显示，%s仍是当前重点，近期建议围绕少提示、稳定完成和场景泛化把目标拆小。", primaryConcern)
	domainAnalysis := make([]string, 0, len(domains))
	for _, domain := range domains {
		domainAnalysis = append(domainAnalysis, autismDevReportInterpretationDomainBrief(domain))
	}
	if len(domainAnalysis) == 0 {
		domainAnalysis = []string{"当前记录已保存孤独症儿童发展评估结构化结果，建议结合各领域题目得分、课堂观察和家庭反馈进行综合理解。"}
	}
	suggestions := []string{
		fmt.Sprintf("优先围绕%s挑1-2个可观察的小目标。", primaryConcern),
		"先用已通过项目做活动入口，再逐步减少提示。",
		"每次只看完成情况、提示量和是否能在不同场景里做出来。",
		"如有X项或缺测，先补观察再做阶段调整。",
	}
	notes := []string{"本解读仅用于评估沟通和训练计划参考。"}
	if autismDevReportInterpretationHasMissing(domains) {
		notes = append(notes, "记录存在X项或缺测，需结合观察和家长反馈判断。")
	}
	return model.ERXinReportInterpretationVO{
		Title:          "孤独症儿童发展评估报告解读",
		GeneratedBy:    "rule",
		GeneratedAt:    time.Now().Format(time.RFC3339),
		Summary:        summary,
		DomainAnalysis: domainAnalysis,
		Suggestions:    suggestions,
		Notes:          notes,
	}
}

func normalizeAutismDevReportInterpretation(result model.ERXinReportInterpretationVO, record model.AssessmentRecordDetailVO, score autismdevscore.AssessmentResult, data autismDevStaticData, itemScores map[int]string, generatedBy string) model.ERXinReportInterpretationVO {
	result.Title = nonEmptyString(result.Title, "孤独症儿童发展评估报告解读")
	result.Model = strings.TrimSpace(result.Model)
	if result.Model == "" && strings.TrimSpace(generatedBy) == "ai" {
		result.Model = deepSeekIEPPlanModel
	}
	result.GeneratedBy = nonEmptyString(result.GeneratedBy, generatedBy)
	if strings.TrimSpace(result.GeneratedBy) == "" {
		result.GeneratedBy = "ai"
	}
	if isLegacyAutismDevReportInterpretation(result) {
		return buildRuleBasedAutismDevReportInterpretation(record, score, data, itemScores)
	}
	result.GeneratedAt = time.Now().Format(time.RFC3339)
	result.Summary = cleanAutismDevReportInterpretationText(result.Summary, 100)
	fallback := buildRuleBasedAutismDevReportInterpretation(record, score, data, itemScores)
	if result.Summary == "" {
		result.Summary = fallback.Summary
	}
	result.DomainAnalysis = compactAutismDevReportInterpretationItems(result.DomainAnalysis, 8, 80)
	result.Suggestions = compactAutismDevReportInterpretationItems(result.Suggestions, 4, 60)
	result.Notes = compactAutismDevReportInterpretationItems(result.Notes, 2, 80)
	if len(result.DomainAnalysis) == 0 {
		result.DomainAnalysis = fallback.DomainAnalysis
	}
	if len(result.Suggestions) == 0 {
		result.Suggestions = fallback.Suggestions
	}
	if len(result.Notes) == 0 {
		result.Notes = []string{"报告解读由孤独症儿童发展评估结构化结果生成，不替代医学诊断。"}
	}
	return result
}

func isLegacyAutismDevReportInterpretation(value model.ERXinReportInterpretationVO) bool {
	if strings.TrimSpace(value.GeneratedBy) != "rule" {
		return false
	}
	summary := strings.TrimSpace(value.Summary)
	if strings.Contains(summary, "本次共") && strings.Contains(summary, "已作答项目") {
		return true
	}
	for _, line := range value.DomainAnalysis {
		text := strings.TrimSpace(line)
		if text == "" {
			continue
		}
		if (strings.Contains(text, "P ") && strings.Contains(text, "E ") && strings.Contains(text, "F ")) ||
			(strings.Contains(text, "A ") && strings.Contains(text, "M ") && strings.Contains(text, "S ")) {
			return true
		}
	}
	return false
}

func compactAutismDevReportInterpretationItems(values []string, maxItems, maxRunes int) []string {
	items := compactNonEmptyStrings(values)
	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		text := cleanAutismDevReportInterpretationText(item, maxRunes)
		if text == "" || autismDevReportInterpretationIsFieldName(text) || isAutismDevReportInterpretationTechnicalFailureNote(text) {
			continue
		}
		key := strings.ToLower(text)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, text)
		if maxItems > 0 && len(result) >= maxItems {
			break
		}
	}
	return result
}

func cleanAutismDevReportInterpretationText(value string, maxRunes int) string {
	text := strings.TrimSpace(value)
	text = strings.Trim(text, "\"'` \t\r\n")
	for {
		runes := []rune(strings.TrimSpace(text))
		index := 0
		for index < len(runes) && runes[index] >= '0' && runes[index] <= '9' {
			index++
		}
		if index == 0 || index >= len(runes) || !strings.ContainsRune(".、)）", runes[index]) {
			break
		}
		text = strings.TrimSpace(string(runes[index+1:]))
	}
	text = strings.TrimSpace(text)
	if maxRunes <= 0 {
		return text
	}
	runes := []rune(text)
	if len(runes) <= maxRunes {
		return text
	}
	cut := maxRunes
	for index := maxRunes; index > maxRunes-20 && index > 0; index-- {
		if strings.ContainsRune("。；;，,、", runes[index-1]) {
			cut = index
			break
		}
	}
	return strings.TrimSpace(string(runes[:cut]))
}

func autismDevReportInterpretationIsFieldName(value string) bool {
	text := strings.ToLower(strings.Trim(strings.TrimSpace(value), "：: "))
	switch text {
	case "domain", "domainname", "name", "title", "analysis", "content", "text", "value", "summary", "description", "suggestion", "suggestions", "note", "notes", "domainanalysis":
		return true
	default:
		return false
	}
}

func isAutismDevReportInterpretationTechnicalFailureNote(value string) bool {
	text := strings.TrimSpace(value)
	return strings.Contains(text, "AI解读生成失败") &&
		(strings.Contains(text, "parse DeepSeek AutismDev report interpretation JSON") ||
			strings.Contains(text, "cannot unmarshal object into Go struct field ERXinReportInterpretationVO.domainAnalysis"))
}

func autismDevReportInterpretationConcernScore(domain autismDevResultAnalysisPromptDomain) int {
	if strings.EqualFold(strings.TrimSpace(domain.ScoreType), autismdevscore.ScoreTypePEF) {
		return domain.ECount*3 + domain.FCount*2 + domain.MissingItemCount
	}
	return domain.SCount*3 + domain.MCount*2 + domain.AbnormalCount + domain.MissingItemCount
}

func autismDevReportInterpretationDomainBrief(domain autismDevResultAnalysisPromptDomain) string {
	name := nonEmptyString(domain.DomainName, domain.DomainCode)
	if strings.EqualFold(strings.TrimSpace(domain.ScoreType), autismdevscore.ScoreTypePEF) {
		passedTitle := autismDevReportInterpretationFirstItemTitle(domain.PassedItems, 1)
		weakItems := append([]autismDevResultAnalysisPromptItem(nil), domain.EmergingItems...)
		weakItems = append(weakItems, domain.FailedItems...)
		weakTitle := autismDevReportInterpretationFirstItemTitle(weakItems, 1)
		strength := "基础反应已出现，但稳定性还不够。"
		if passedTitle != "" {
			strength = fmt.Sprintf("已能完成“%s”等基础项目。", passedTitle)
		}
		weakness := autismDevReportInterpretationDomainNextStep(name, weakTitle)
		return cleanAutismDevReportInterpretationText(fmt.Sprintf("%s：%s%s", name, strength, weakness), 80)
	}
	parts := []string{
		fmt.Sprintf("适应性%d项", domain.ACount),
		fmt.Sprintf("中等%d项", domain.MCount),
		fmt.Sprintf("严重%d项", domain.SCount),
	}
	if domain.XCount > 0 {
		parts = append(parts, fmt.Sprintf("X%d项", domain.XCount))
	}
	status := "近期重点是等待、规则配合和替代表达。"
	if domain.SCount > 0 || domain.AbnormalCount > 0 {
		status = "当前先聚焦情绪稳定、替代表达和减少冲动行为。"
	} else if domain.MCount > 0 {
		status = "部分行为表现需要在结构化场景中持续引导。"
	}
	return cleanAutismDevReportInterpretationText(fmt.Sprintf("%s：%s，%s", name, strings.Join(parts, "、"), status), 80)
}

func autismDevReportInterpretationFirstItemTitle(items []autismDevResultAnalysisPromptItem, limit int) string {
	titles := autismDevResultAnalysisItemTitles(items, limit)
	if len(titles) == 0 {
		return ""
	}
	return stripAutismDevParenthetical(strings.TrimSpace(titles[0]))
}

func autismDevReportInterpretationDomainNextStep(domainName, weakTitle string) string {
	focus := autismDevReportInterpretationFocusPhrase(domainName)
	if weakTitle != "" {
		return fmt.Sprintf("“%s”仍需继续练习，近期先练%s", weakTitle, focus)
	}
	return "近期先练" + focus
}

func autismDevReportInterpretationFocusPhrase(domainName string) string {
	switch strings.TrimSpace(domainName) {
	case "感知觉":
		return "视觉追踪和稳定注意。"
	case "粗大动作":
		return "平衡、协调和动作模仿。"
	case "精细动作":
		return "抓握、配对和手眼协调。"
	case "语言与沟通":
		return "主动回应、表达需求和简单对话。"
	case "认知":
		return "配对、分类、理解和记忆。"
	case "社会交往":
		return "目光接触、轮流和同伴回应。"
	case "生活自理":
		return "穿脱、进食和如厕等独立步骤。"
	case "情绪与行为":
		return "等待、情绪表达和替代行为。"
	default:
		return "稳定完成和少提示配合。"
	}
}

func autismDevReportInterpretationHasMissing(domains []autismDevResultAnalysisPromptDomain) bool {
	for _, domain := range domains {
		if domain.XCount > 0 || domain.MissingItemCount > 0 || len(domain.Warnings) > 0 {
			return true
		}
	}
	return false
}
