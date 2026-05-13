package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"go-migration-platform/services/platform/internal/model"
)

var platformPEP3IEPGoalTimePrefixPattern = regexp.MustCompile(`^\s*(?:在)?(?:未来)?(?:\d+|[一二三四五六七八九十半]+)\s*(?:个)?(?:月|周|星期|年)(?:内|后|期间)?[，,、\s]*`)

const (
	platformPEP3IEPMaterialAIBaseURLDefault         = "https://ai.yiqiu.dev/v1"
	platformPEP3IEPMaterialAIModelDefault           = "gpt-5.5"
	platformPEP3IEPMaterialAIReasoningEffortDefault = "xhigh"
	platformPEP3IEPMaterialAITimeout                = 45 * time.Second
)

type platformPEP3IEPMaterialAIResponsesRequest struct {
	Model           string                                    `json:"model"`
	Input           []platformPEP3IEPMaterialAIResponsesInput `json:"input"`
	Reasoning       platformPEP3IEPMaterialAIReasoning        `json:"reasoning"`
	Store           bool                                      `json:"store"`
	MaxOutputTokens int                                       `json:"max_output_tokens,omitempty"`
	Text            platformPEP3IEPMaterialAIResponseText     `json:"text"`
}

type platformPEP3IEPMaterialAIResponsesInput struct {
	Role    string                                      `json:"role"`
	Content []platformPEP3IEPMaterialAIResponsesContent `json:"content"`
}

type platformPEP3IEPMaterialAIResponsesContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type platformPEP3IEPMaterialAIReasoning struct {
	Effort string `json:"effort"`
}

type platformPEP3IEPMaterialAIResponseText struct {
	Format map[string]string `json:"format"`
}

type platformPEP3IEPMaterialAIResponsesResponse struct {
	OutputText string `json:"output_text,omitempty"`
	Output     []struct {
		Type    string `json:"type"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content,omitempty"`
	} `json:"output,omitempty"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type platformPEP3IEPMaterialAIPromptPayload struct {
	Target                   string   `json:"target"`
	Domain                   string   `json:"domain,omitempty"`
	DomainCode               string   `json:"domainCode,omitempty"`
	ItemNo                   int      `json:"itemNo,omitempty"`
	ItemTitle                string   `json:"itemTitle,omitempty"`
	ScoreValue               int      `json:"scoreValue"`
	ScoreLabel               string   `json:"scoreLabel,omitempty"`
	ScoreDescription         string   `json:"scoreDescription,omitempty"`
	LongGoal                 string   `json:"longGoal,omitempty"`
	ShortGoal                string   `json:"shortGoal,omitempty"`
	CourseForm               string   `json:"courseForm,omitempty"`
	ExistingShortGoals       []string `json:"existingShortGoals,omitempty"`
	ExistingTrainingProjects []string `json:"existingTrainingProjects,omitempty"`
	ExistingTrainingContents []string `json:"existingTrainingContents,omitempty"`
	OutputRule               string   `json:"outputRule"`
}

func (svc *Service) GeneratePlatformPEP3IEPMaterialAI(req model.PEP3IEPMaterialAIGenerateRequest) (model.PEP3IEPMaterialAIGenerateResult, error) {
	prepared, err := svc.preparePlatformPEP3IEPMaterialAIRequest(req)
	if err != nil {
		return model.PEP3IEPMaterialAIGenerateResult{}, err
	}
	if err := validatePlatformPEP3IEPMaterialAIRequest(prepared); err != nil {
		return model.PEP3IEPMaterialAIGenerateResult{}, err
	}

	apiKey := strings.TrimSpace(os.Getenv("PEP3_IEP_MATERIAL_AI_API_KEY"))
	if apiKey == "" {
		apiKey = strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	}
	if apiKey == "" {
		return model.PEP3IEPMaterialAIGenerateResult{}, errors.New("AI密钥未配置，请设置 PEP3_IEP_MATERIAL_AI_API_KEY")
	}

	baseURL := strings.TrimSpace(os.Getenv("PEP3_IEP_MATERIAL_AI_BASE_URL"))
	if baseURL == "" {
		baseURL = platformPEP3IEPMaterialAIBaseURLDefault
	}
	endpoint := buildPlatformPEP3IEPMaterialAIEndpoint(baseURL)
	modelName := strings.TrimSpace(os.Getenv("PEP3_IEP_MATERIAL_AI_MODEL"))
	if modelName == "" {
		modelName = platformPEP3IEPMaterialAIModelDefault
	}
	reasoningEffort := strings.TrimSpace(os.Getenv("PEP3_IEP_MATERIAL_AI_REASONING_EFFORT"))
	if reasoningEffort == "" {
		reasoningEffort = platformPEP3IEPMaterialAIReasoningEffortDefault
	}

	payload := platformPEP3IEPMaterialAIPromptPayload{
		Target:                   prepared.Target,
		Domain:                   prepared.Domain,
		DomainCode:               prepared.DomainCode,
		ItemNo:                   prepared.ItemNo,
		ItemTitle:                prepared.ItemTitle,
		ScoreValue:               platformPEP3IEPMaterialScoreValue(prepared),
		ScoreLabel:               prepared.ScoreLabel,
		ScoreDescription:         prepared.ScoreDescription,
		LongGoal:                 prepared.LongGoal,
		ShortGoal:                prepared.ShortGoal,
		CourseForm:               prepared.CourseForm,
		ExistingShortGoals:       prepared.ExistingShortGoals,
		ExistingTrainingProjects: prepared.ExistingTrainingProjects,
		ExistingTrainingContents: prepared.ExistingTrainingContents,
		OutputRule:               platformPEP3IEPMaterialAIOutputRule(prepared.Target),
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return model.PEP3IEPMaterialAIGenerateResult{}, err
	}
	requestBody, err := json.Marshal(platformPEP3IEPMaterialAIResponsesRequest{
		Model: modelName,
		Input: []platformPEP3IEPMaterialAIResponsesInput{
			{
				Role: "system",
				Content: []platformPEP3IEPMaterialAIResponsesContent{
					{Type: "input_text", Text: platformPEP3IEPMaterialAISystemPrompt()},
				},
			},
			{
				Role: "user",
				Content: []platformPEP3IEPMaterialAIResponsesContent{
					{Type: "input_text", Text: string(payloadJSON)},
				},
			},
		},
		Reasoning: platformPEP3IEPMaterialAIReasoning{Effort: reasoningEffort},
		Store:     false,
		Text: platformPEP3IEPMaterialAIResponseText{
			Format: map[string]string{"type": "json_object"},
		},
	})
	if err != nil {
		return model.PEP3IEPMaterialAIGenerateResult{}, err
	}

	requestCtx, cancel := context.WithTimeout(context.Background(), platformPEP3IEPMaterialAITimeout)
	defer cancel()
	httpReq, err := http.NewRequestWithContext(requestCtx, http.MethodPost, endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return model.PEP3IEPMaterialAIGenerateResult{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := (&http.Client{Timeout: platformPEP3IEPMaterialAITimeout + 5*time.Second}).Do(httpReq)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(requestCtx.Err(), context.DeadlineExceeded) {
			return model.PEP3IEPMaterialAIGenerateResult{}, fmt.Errorf("AI生成超时（%d秒）", int(platformPEP3IEPMaterialAITimeout.Seconds()))
		}
		return model.PEP3IEPMaterialAIGenerateResult{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.PEP3IEPMaterialAIGenerateResult{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return model.PEP3IEPMaterialAIGenerateResult{}, fmt.Errorf("AI接口返回 %d：%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var aiResponse platformPEP3IEPMaterialAIResponsesResponse
	if err := json.Unmarshal(body, &aiResponse); err != nil {
		return model.PEP3IEPMaterialAIGenerateResult{}, err
	}
	if aiResponse.Error != nil && strings.TrimSpace(aiResponse.Error.Message) != "" {
		return model.PEP3IEPMaterialAIGenerateResult{}, errors.New(aiResponse.Error.Message)
	}

	content := extractPlatformPEP3IEPMaterialAIResponseText(aiResponse)
	if content == "" {
		return model.PEP3IEPMaterialAIGenerateResult{}, errors.New("AI接口返回内容为空")
	}

	result, err := parsePlatformPEP3IEPMaterialAIResult(content)
	if err != nil {
		return model.PEP3IEPMaterialAIGenerateResult{}, err
	}
	result = normalizePlatformPEP3IEPMaterialAIResult(result)
	if err := validatePlatformPEP3IEPMaterialAIResult(prepared, result); err != nil {
		return model.PEP3IEPMaterialAIGenerateResult{}, err
	}
	result.Source = "ai"
	return result, nil
}

func (svc *Service) preparePlatformPEP3IEPMaterialAIRequest(req model.PEP3IEPMaterialAIGenerateRequest) (model.PEP3IEPMaterialAIGenerateRequest, error) {
	req.Target = strings.TrimSpace(req.Target)
	req.Domain = strings.TrimSpace(req.Domain)
	req.DomainCode = strings.TrimSpace(req.DomainCode)
	req.ItemTitle = cleanPlatformPEP3IEPMaterialText(req.ItemTitle)
	req.ScoreLabel = strings.TrimSpace(req.ScoreLabel)
	req.ScoreDescription = strings.TrimSpace(req.ScoreDescription)
	req.LongGoal = cleanPlatformPEP3IEPMaterialText(req.LongGoal)
	req.ShortGoal = cleanPlatformPEP3IEPMaterialText(req.ShortGoal)
	req.CourseForm = strings.TrimSpace(req.CourseForm)
	req.ExistingShortGoals = normalizePlatformPEP3IEPExistingTexts(req.ExistingShortGoals, 20)
	req.ExistingTrainingProjects = normalizePlatformPEP3IEPExistingTexts(req.ExistingTrainingProjects, 20)
	req.ExistingTrainingContents = normalizePlatformPEP3IEPExistingTexts(req.ExistingTrainingContents, 20)

	if req.ItemNo <= 0 {
		return req, nil
	}

	item, ok, err := svc.lookupPlatformPEP3AssessmentItem(req.ItemNo)
	if err != nil {
		return req, err
	}
	if !ok {
		return req, errors.New("PEP-3题目不存在")
	}
	if req.ItemTitle == "" {
		req.ItemTitle = cleanPlatformPEP3IEPMaterialText(firstPlatformPEP3IEPMaterialText(item.ItemTitle, item.TestItem))
	}
	if req.DomainCode == "" {
		req.DomainCode = strings.TrimSpace(item.DomainCode)
	}
	if req.Domain == "" {
		req.Domain = strings.TrimSpace(item.DomainName)
	}
	if req.ScoreValue != nil {
		for _, option := range item.ScoreOptions {
			if option.Value != *req.ScoreValue {
				continue
			}
			if req.ScoreLabel == "" {
				req.ScoreLabel = option.Label
			}
			if req.ScoreDescription == "" {
				req.ScoreDescription = option.Description
			}
			break
		}
	}
	return req, nil
}

func validatePlatformPEP3IEPMaterialAIRequest(req model.PEP3IEPMaterialAIGenerateRequest) error {
	switch req.Target {
	case "long_goal":
		if req.ItemNo <= 0 || req.ItemTitle == "" {
			return errors.New("请先选择题目")
		}
		if req.ScoreValue == nil || (*req.ScoreValue != 0 && *req.ScoreValue != 1 && *req.ScoreValue != 2) {
			return errors.New("请先选择选项")
		}
	case "short_goal":
		if req.ItemNo <= 0 || req.ItemTitle == "" {
			return errors.New("请先选择题目")
		}
		if req.ScoreValue == nil || (*req.ScoreValue != 0 && *req.ScoreValue != 1 && *req.ScoreValue != 2) {
			return errors.New("请先选择选项")
		}
		if req.LongGoal == "" {
			return errors.New("请先保存或填写长期目标")
		}
	case "training":
		if req.ShortGoal == "" {
			return errors.New("请先选择或填写短期目标")
		}
	default:
		return errors.New("不支持的AI生成类型")
	}
	return nil
}

func buildPlatformPEP3IEPMaterialAIEndpoint(baseURL string) string {
	baseURL = strings.TrimSpace(baseURL)
	baseURL = strings.TrimRight(baseURL, "/")
	if strings.HasSuffix(baseURL, "/responses") {
		return baseURL
	}
	if strings.HasSuffix(baseURL, "/chat/completions") {
		return strings.TrimSuffix(baseURL, "/chat/completions") + "/responses"
	}
	return baseURL + "/responses"
}

func platformPEP3IEPMaterialAISystemPrompt() string {
	return strings.TrimSpace(`你是PEP3 IEP素材库编辑助手。你要输出可以直接保存到数据库的中文内容。
规则：
1. 测评题目只是判断入口，不是训练内容模板；必须从题目提炼同类功能方向，不要机械复刻测评原题。
2. 生成目标和训练内容时，要围绕题目反映出的底层操作、沟通、认知或社交功能，泛化到日常生活、课堂或游戏中的同类任务。
3. 必须先判断 target，再按对应字段层级生成，不能把短期目标或训练内容的细节写进长期目标。
4. 长期目标只写泛化后的目标行为、独立或提示程度；不要写时间周期、材料规格、动作拆解、练习次数、百分比、连续成功标准、训练步骤。
5. 短期目标是阶段目标，可以写泛化任务、支持方式和达标标准，但不要写时间周期或完整训练流程。
6. 训练内容是一段简明活动描述，直接写准备哪些同类材料、教师怎么提示、儿童完成什么操作、如何更换材料或场景泛化；不要写成分段教案。
7. 训练内容至少包含2种同类材料或活动，不能只练原题材料。例如“旋开瓶盖”应泛化为旋拧类双手操作，“吹泡泡”应泛化为口部吹气控制。
8. 如果输入里有 existingShortGoals、existingTrainingProjects、existingTrainingContents，必须避开已有目标、训练项目、材料组合、提示语和活动方式，生成明显不同的一条。
9. 不要只做同义改写；要换泛化方向、换材料、换场景或换任务形式。
10. 不要写“自然情境”“相关活动”“能力点”这类空话。
11. 短期目标必须同时输出课程形式，只能是“个训”或“集体课”。
12. 不要输出题号，不要输出解释，只返回JSON对象。`)
}

func platformPEP3IEPMaterialAIOutputRule(target string) string {
	switch target {
	case "long_goal":
		return `只返回 {"longGoal":"..."}。长期目标不超过35个汉字，要写泛化后的同类功能，不要只照搬题目；不写时间周期、材料尺寸、手部动作细节、训练次数或成功标准。示例：{"longGoal":"儿童能独立完成日常旋拧类双手操作。"}`
	case "short_goal":
		return `只返回 {"shortGoal":"...","courseForm":"个训或集体课"}。短期目标要从题目泛化到同类任务，可包含提示方式和达标标准，不写时间周期或训练步骤；必须避开 existingShortGoals 中已有表达和任务方向。示例：{"shortGoal":"儿童能在口头提示下扶稳物品并旋转打开常见旋拧材料，连续3次中至少完成2次。","courseForm":"个训"}`
	case "training":
		return `只返回 {"trainingProject":"...","trainingContent":"..."}。训练项目写泛化功能名称，不要只写题目名；训练内容控制在60-100个汉字，一段话写完，不要使用“材料：”“步骤：”“提示方式：”“完成标准：”这些标签，不要写多轮次数标准；必须避开 existingTrainingProjects 和 existingTrainingContents 中已有的材料、提示语、活动方式。示例：{"trainingProject":"旋拧类双手操作","trainingContent":"准备儿童容易握持的水瓶、小罐和旋钮玩具，教师只用口头提示“扶住、转一转”，让儿童完成打开和关上；更换不同大小和松紧的材料练习泛化。"}`
	default:
		return `只返回JSON对象`
	}
}

func parsePlatformPEP3IEPMaterialAIResult(content string) (model.PEP3IEPMaterialAIGenerateResult, error) {
	candidate := strings.TrimSpace(content)
	if !strings.HasPrefix(candidate, "{") {
		start := strings.Index(candidate, "{")
		end := strings.LastIndex(candidate, "}")
		if start >= 0 && end > start {
			candidate = candidate[start : end+1]
		}
	}
	var result model.PEP3IEPMaterialAIGenerateResult
	if err := json.Unmarshal([]byte(candidate), &result); err != nil {
		return model.PEP3IEPMaterialAIGenerateResult{}, fmt.Errorf("解析AI结果失败：%w", err)
	}
	return result, nil
}

func extractPlatformPEP3IEPMaterialAIResponseText(response platformPEP3IEPMaterialAIResponsesResponse) string {
	if text := strings.TrimSpace(response.OutputText); text != "" {
		return text
	}
	parts := make([]string, 0, 2)
	for _, item := range response.Output {
		if item.Type != "" && item.Type != "message" {
			continue
		}
		for _, content := range item.Content {
			if content.Type != "" && content.Type != "output_text" && content.Type != "text" {
				continue
			}
			if text := strings.TrimSpace(content.Text); text != "" {
				parts = append(parts, text)
			}
		}
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}

func normalizePlatformPEP3IEPMaterialAIResult(result model.PEP3IEPMaterialAIGenerateResult) model.PEP3IEPMaterialAIGenerateResult {
	result.LongGoal = stripPlatformPEP3IEPGoalTimePrefix(cleanPlatformPEP3IEPMaterialText(result.LongGoal))
	result.ShortGoal = stripPlatformPEP3IEPGoalTimePrefix(cleanPlatformPEP3IEPMaterialText(result.ShortGoal))
	result.CourseForm = normalizePlatformPEP3IEPMaterialCourseForm(result.CourseForm)
	result.TrainingProject = cleanPlatformPEP3IEPMaterialText(result.TrainingProject)
	result.TrainingContent = cleanPlatformPEP3IEPMaterialText(result.TrainingContent)
	return result
}

func validatePlatformPEP3IEPMaterialAIResult(req model.PEP3IEPMaterialAIGenerateRequest, result model.PEP3IEPMaterialAIGenerateResult) error {
	switch req.Target {
	case "long_goal":
		if result.LongGoal == "" {
			return errors.New("AI没有生成长期目标，请重试")
		}
		if isOverDetailedPlatformPEP3IEPLongGoal(result.LongGoal) {
			return errors.New("AI生成的长期目标过细，请重新生成")
		}
	case "short_goal":
		if result.ShortGoal == "" {
			return errors.New("AI没有生成短期目标，请重试")
		}
		if result.CourseForm == "" {
			return errors.New("AI没有生成有效课程形式，请重试")
		}
		if isTooSimilarPlatformPEP3IEPText(result.ShortGoal, req.ExistingShortGoals, 0.58) {
			return errors.New("AI生成的短期目标与已有目标过于相似，请重新生成")
		}
	case "training":
		if result.TrainingProject == "" {
			return errors.New("AI没有生成训练项目，请重试")
		}
		if result.TrainingContent == "" {
			return errors.New("AI没有生成训练内容，请重试")
		}
		if isOverDetailedPlatformPEP3IEPTrainingContent(result.TrainingContent) {
			return errors.New("AI生成的训练内容过细，请重新生成")
		}
		if isTooSimilarPlatformPEP3IEPText(result.TrainingProject, req.ExistingTrainingProjects, 0.8) ||
			isTooSimilarPlatformPEP3IEPText(result.TrainingContent, req.ExistingTrainingContents, 0.55) {
			return errors.New("AI生成的训练内容与已有内容过于相似，请重新生成")
		}
	}
	return nil
}

func isOverDetailedPlatformPEP3IEPLongGoal(text string) bool {
	text = strings.TrimSpace(text)
	if len([]rune(text)) > 55 {
		return true
	}
	overDetailedWords := []string{
		"材料", "步骤", "标准", "连续", "至少", "成功", "直径", "厘米", "前三指", "一手", "另一手", "练习中", "百分",
		"个月", "周内", "星期", "年内", "未来", "本阶段", "本月", "本季度",
	}
	for _, word := range overDetailedWords {
		if strings.Contains(text, word) {
			return true
		}
	}
	return false
}

func isOverDetailedPlatformPEP3IEPTrainingContent(text string) bool {
	text = strings.TrimSpace(text)
	if len([]rune(text)) > 140 {
		return true
	}
	overDetailedWords := []string{
		"材料：", "步骤：", "提示方式", "完成标准", "标准：", "连续", "至少", "5次", "4次", "百分",
	}
	for _, word := range overDetailedWords {
		if strings.Contains(text, word) {
			return true
		}
	}
	return false
}

func isTooSimilarPlatformPEP3IEPText(candidate string, existing []string, threshold float64) bool {
	candidate = normalizePlatformPEP3IEPTextForSimilarity(candidate)
	if candidate == "" || len(existing) == 0 {
		return false
	}
	for _, item := range existing {
		item = normalizePlatformPEP3IEPTextForSimilarity(item)
		if item == "" {
			continue
		}
		if candidate == item || strings.Contains(candidate, item) || strings.Contains(item, candidate) {
			return true
		}
		if platformPEP3IEPBigramJaccard(candidate, item) >= threshold {
			return true
		}
	}
	return false
}

func normalizePlatformPEP3IEPExistingTexts(values []string, limit int) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = cleanPlatformPEP3IEPMaterialText(value)
		if value == "" {
			continue
		}
		key := normalizePlatformPEP3IEPTextForSimilarity(value)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, value)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out
}

func normalizePlatformPEP3IEPTextForSimilarity(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	replacer := strings.NewReplacer(
		" ", "", "\t", "", "\n", "", "\r", "",
		"，", "", "。", "", "；", "", "、", "", "：", "", ":", "", ";", "", ",", "", ".", "",
		"“", "", "”", "", "\"", "", "'", "", "‘", "", "’", "",
		"(", "", ")", "", "（", "", "）", "", "-", "", "—", "",
		"儿童", "", "教师", "", "成人", "", "能够", "能", "可以", "能",
	)
	return replacer.Replace(text)
}

func platformPEP3IEPBigramJaccard(left, right string) float64 {
	leftSet := platformPEP3IEPBigramSet(left)
	rightSet := platformPEP3IEPBigramSet(right)
	if len(leftSet) == 0 || len(rightSet) == 0 {
		return 0
	}
	intersection := 0
	for key := range leftSet {
		if _, ok := rightSet[key]; ok {
			intersection++
		}
	}
	union := len(leftSet) + len(rightSet) - intersection
	if union <= 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

func platformPEP3IEPBigramSet(text string) map[string]struct{} {
	runes := []rune(text)
	out := map[string]struct{}{}
	if len(runes) < 2 {
		if len(runes) == 1 {
			out[string(runes)] = struct{}{}
		}
		return out
	}
	for idx := 0; idx < len(runes)-1; idx++ {
		out[string(runes[idx:idx+2])] = struct{}{}
	}
	return out
}

func stripPlatformPEP3IEPGoalTimePrefix(text string) string {
	text = strings.TrimSpace(text)
	for {
		next := platformPEP3IEPGoalTimePrefixPattern.ReplaceAllString(text, "")
		next = strings.TrimSpace(next)
		if next == text {
			return text
		}
		text = next
	}
}

func normalizePlatformPEP3IEPMaterialCourseForm(value string) string {
	value = strings.TrimSpace(value)
	if value == "个训" || value == "集体课" {
		return value
	}
	return ""
}

func platformPEP3IEPMaterialScoreValue(req model.PEP3IEPMaterialAIGenerateRequest) int {
	if req.ScoreValue == nil {
		return 1
	}
	return *req.ScoreValue
}

func firstPlatformPEP3IEPMaterialText(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func cleanPlatformPEP3IEPMaterialText(text string) string {
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, "\r\n", "\n")
	lines := strings.Split(text, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = strings.TrimLeft(line, "0123456789.、．)）- ")
		line = strings.Trim(line, "，。；;:：、")
		if line != "" {
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
}
