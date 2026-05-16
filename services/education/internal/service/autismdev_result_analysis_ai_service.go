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

type autismDevResultAnalysisPromptPayload struct {
	Student       pep3IEPPlanPromptStudent               `json:"student"`
	Assessment    autismDevResultAnalysisAssessment      `json:"assessment"`
	Domains       []autismDevResultAnalysisPromptDomain  `json:"domains"`
	OutputRequest erxinReportInterpretationOutputRequest `json:"outputRequest"`
}

type autismDevResultAnalysisAssessment struct {
	Date         string `json:"date,omitempty"`
	ScaleVersion string `json:"scaleVersion,omitempty"`
	DataStatus   string `json:"dataStatus,omitempty"`
	Remark       string `json:"remark,omitempty"`
}

type autismDevResultAnalysisPromptDomain struct {
	DomainCode        string                              `json:"domainCode"`
	DomainName        string                              `json:"domainName"`
	ScoreType         string                              `json:"scoreType"`
	ItemCount         int                                 `json:"itemCount"`
	AnsweredItemCount int                                 `json:"answeredItemCount"`
	MissingItemCount  int                                 `json:"missingItemCount"`
	Complete          bool                                `json:"complete"`
	PCount            int                                 `json:"pCount,omitempty"`
	ECount            int                                 `json:"eCount,omitempty"`
	FCount            int                                 `json:"fCount,omitempty"`
	XCount            int                                 `json:"xCount,omitempty"`
	PECount           int                                 `json:"peCount,omitempty"`
	RawScore          int                                 `json:"rawScore,omitempty"`
	ScorableItemCount int                                 `json:"scorableItemCount,omitempty"`
	ScoreRate         float64                             `json:"scoreRate,omitempty"`
	DevelopmentLevel  string                              `json:"developmentLevel,omitempty"`
	PassedItems       []autismDevResultAnalysisPromptItem `json:"passedItems,omitempty"`
	EmergingItems     []autismDevResultAnalysisPromptItem `json:"emergingItems,omitempty"`
	FailedItems       []autismDevResultAnalysisPromptItem `json:"failedItems,omitempty"`
	TargetItems       []autismDevResultAnalysisPromptItem `json:"targetItems,omitempty"`
	Warnings          []string                            `json:"warnings,omitempty"`
}

type autismDevResultAnalysisPromptItem struct {
	ItemNo          int    `json:"itemNo"`
	DomainItemNo    int    `json:"domainItemNo"`
	Title           string `json:"title"`
	Score           string `json:"score"`
	AgeSegment      string `json:"ageSegment,omitempty"`
	AgeMinMonth     int    `json:"ageMinMonth,omitempty"`
	AgeMaxMonth     int    `json:"ageMaxMonth,omitempty"`
	AssessmentRange string `json:"assessmentRange,omitempty"`
	PassCriteria    string `json:"passCriteria,omitempty"`
}

var autismDevResultAnalysisDomains = []string{
	autismdevscore.DomainSensory,
	autismdevscore.DomainGrossMotor,
	autismdevscore.DomainFineMotor,
	autismdevscore.DomainLanguageComm,
	autismdevscore.DomainCognition,
	autismdevscore.DomainSocial,
	autismdevscore.DomainDailyLiving,
}

func (svc *Service) GetAutismDevResultAnalysis(userID, recordID int64) (model.AutismDevResultAnalysisVO, error) {
	if svc.repo == nil {
		return model.AutismDevResultAnalysisVO{}, errors.New("assessment repository is not configured")
	}
	instID, record, score, data, itemScores, err := svc.autismDevResultAnalysisContext(userID, recordID)
	if err != nil {
		return model.AutismDevResultAnalysisVO{}, err
	}
	var cached model.AutismDevResultAnalysisVO
	_, err = svc.repo.GetAssessmentReportInterpretationJSON(context.Background(), instID, recordID, autismDevScaleCode, &cached)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.AutismDevResultAnalysisVO{}, nil
		}
		return model.AutismDevResultAnalysisVO{}, err
	}
	normalized := normalizeAutismDevResultAnalysis(cached, record, score, data, itemScores, cached.GeneratedBy)
	if strings.TrimSpace(cached.GeneratedAt) != "" {
		normalized.GeneratedAt = cached.GeneratedAt
	}
	return normalized, nil
}

func (svc *Service) autismDevResultAnalysisContext(userID, recordID int64) (int64, model.AssessmentRecordDetailVO, autismdevscore.AssessmentResult, autismDevStaticData, map[int]string, error) {
	if recordID <= 0 {
		return 0, model.AssessmentRecordDetailVO{}, autismdevscore.AssessmentResult{}, autismDevStaticData{}, nil, errors.New("invalid assessment record id")
	}
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return 0, model.AssessmentRecordDetailVO{}, autismdevscore.AssessmentResult{}, autismDevStaticData{}, nil, err
	}
	record, err := svc.GetAutismDevAssessmentRecord(userID, recordID)
	if err != nil {
		return 0, model.AssessmentRecordDetailVO{}, autismdevscore.AssessmentResult{}, autismDevStaticData{}, nil, err
	}
	score, err := decodeSavedAutismDevScore(record.ResultJSON)
	if err != nil {
		return 0, model.AssessmentRecordDetailVO{}, autismdevscore.AssessmentResult{}, autismDevStaticData{}, nil, err
	}
	data, err := loadAutismDevStaticData()
	if err != nil {
		return 0, model.AssessmentRecordDetailVO{}, autismdevscore.AssessmentResult{}, autismDevStaticData{}, nil, err
	}
	itemScores, _ := decodeSavedAutismDevInputScores(record.InputJSON)
	return instID, record, score.Result, data, itemScores, nil
}

func (svc *Service) GenerateAutismDevResultAnalysisStream(ctx context.Context, userID, recordID int64, onDelta func(string) error) (model.AutismDevResultAnalysisVO, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if svc.repo == nil {
		return model.AutismDevResultAnalysisVO{}, errors.New("assessment repository is not configured")
	}
	instID, record, score, data, itemScores, err := svc.autismDevResultAnalysisContext(userID, recordID)
	if err != nil {
		return model.AutismDevResultAnalysisVO{}, err
	}
	payload := buildAutismDevResultAnalysisPromptPayload(record, score, data, itemScores)
	result, err := callDeepSeekAutismDevResultAnalysisStream(ctx, payload, onDelta)
	if err != nil {
		fallback := buildRuleBasedAutismDevResultAnalysis(record, score, data, itemScores)
		if saveErr := svc.saveAutismDevResultAnalysis(ctx, instID, userID, recordID, record, score, data, itemScores, fallback); saveErr != nil {
			return model.AutismDevResultAnalysisVO{}, saveErr
		}
		return fallback, nil
	}
	normalized := normalizeAutismDevResultAnalysis(result, record, score, data, itemScores, "ai")
	if err := svc.saveAutismDevResultAnalysis(ctx, instID, userID, recordID, record, score, data, itemScores, normalized); err != nil {
		return model.AutismDevResultAnalysisVO{}, err
	}
	return normalized, nil
}

func (svc *Service) SaveAutismDevResultAnalysis(ctx context.Context, userID, recordID int64, analysis model.AutismDevResultAnalysisVO) (model.AutismDevResultAnalysisVO, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if svc.repo == nil {
		return model.AutismDevResultAnalysisVO{}, errors.New("assessment repository is not configured")
	}
	instID, record, score, data, itemScores, err := svc.autismDevResultAnalysisContext(userID, recordID)
	if err != nil {
		return model.AutismDevResultAnalysisVO{}, err
	}
	normalized := normalizeAutismDevResultAnalysis(analysis, record, score, data, itemScores, analysis.GeneratedBy)
	if err := svc.saveAutismDevResultAnalysis(ctx, instID, userID, recordID, record, score, data, itemScores, normalized); err != nil {
		return model.AutismDevResultAnalysisVO{}, err
	}
	return normalized, nil
}

func (svc *Service) saveAutismDevResultAnalysis(ctx context.Context, instID, userID, recordID int64, record model.AssessmentRecordDetailVO, score autismdevscore.AssessmentResult, data autismDevStaticData, itemScores map[int]string, interpretation model.AutismDevResultAnalysisVO) error {
	return svc.repo.UpsertAssessmentReportInterpretationJSON(ctx, repository.AssessmentReportInterpretationJSONEntity{
		InstID:         instID,
		RecordID:       recordID,
		AssessmentCode: autismDevScaleCode,
		SourceHash:     autismDevResultAnalysisSourceHash(record),
		Model:          strings.TrimSpace(nonEmptyString(interpretation.Model, deepSeekIEPPlanModel)),
		GeneratedBy:    strings.TrimSpace(nonEmptyString(interpretation.GeneratedBy, "manual")),
	}, interpretation, userID)
}

func autismDevResultAnalysisSourceHash(record model.AssessmentRecordDetailVO) string {
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

func buildAutismDevResultAnalysisPromptPayload(record model.AssessmentRecordDetailVO, result autismdevscore.AssessmentResult, data autismDevStaticData, itemScores map[int]string) autismDevResultAnalysisPromptPayload {
	return autismDevResultAnalysisPromptPayload{
		Student: pep3IEPPlanPromptStudent{
			Name:      strings.TrimSpace(record.StudentName),
			Gender:    strings.TrimSpace(record.StudentGender),
			BirthDate: formatIEPPlanDate(record.BirthDate),
			Age:       formatIEPPlanAge(record.AgeYears, record.AgeMonths, record.AgeDays),
		},
		Assessment: autismDevResultAnalysisAssessment{
			Date:         formatIEPPlanDate(record.AssessmentDate),
			ScaleVersion: strings.TrimSpace(record.ScaleVersion),
			DataStatus:   strings.TrimSpace(record.DataStatus),
			Remark:       strings.TrimSpace(record.Remark),
		},
		Domains: autismDevResultAnalysisPromptDomains(record, result, data, itemScores),
		OutputRequest: erxinReportInterpretationOutputRequest{
			RequiredSchema: "只输出JSON：title, rows[{domain,status,strengths,weaknesses,targets}]。rows必须按输入domains顺序输出，只包含输入的领域；domain使用中文领域名；status写能力现状分析，必须同时包含1-3个passedItems具体项目和1-3个emergingItems/failedItems具体项目，developmentLevel只能作为辅助信息，不能只写“达到同龄正常儿童X岁左右标准”；禁止使用“能力发展明显超前/明显落后/发展较好/发展较弱”这类模板判断；strengths只写优势内容，不要带“优势：”，按实际列出passedItems中的具体项目，不足时少写；weaknesses只写劣势内容，不要带“劣势：”，按实际列出emergingItems/failedItems中的具体项目，不足时少写；targets只挑近期关键训练目标，通常3-6条，最多7条，候选目标少就少写，不要每个领域写相同条数，不要为了凑数量编目标；候选目标多时按E项优先和近端年龄段选择；每条短句，不输出括号里的P通过标准，格式类似“1 能……。”。",
			SafetyRules:    "必须基于孤独症儿童发展评估表的题目得分生成；不得修改P、E、F、X计数、年龄段、日期等事实；不得做医学诊断；不得输出治疗方案；禁止空话套话，如“继续提升综合能力”“加强相关训练”“提高配合度”“能力发展明显超前”等没有具体项目的表达；禁止模板化句式和重复开头；每个单元格要像报告样例一样具体、可观察、可训练。",
		},
	}
}

func autismDevResultAnalysisPromptDomains(record model.AssessmentRecordDetailVO, result autismdevscore.AssessmentResult, data autismDevStaticData, itemScores map[int]string) []autismDevResultAnalysisPromptDomain {
	byCode := make(map[string]autismdevscore.DomainResult, len(result.Domains))
	for _, domain := range result.Domains {
		byCode[strings.TrimSpace(domain.DomainCode)] = domain
	}
	itemsByDomain := autismDevItemsByDomain(data.items)
	out := make([]autismDevResultAnalysisPromptDomain, 0, len(autismDevResultAnalysisDomains))
	for _, code := range autismDevResultAnalysisDomainCodes(record, result, data) {
		domain, ok := byCode[code]
		if !ok || strings.TrimSpace(domain.DomainName) == "" {
			continue
		}
		passedItems, emergingItems, failedItems := autismDevResultAnalysisScoredItems(itemsByDomain[code], itemScores)
		targetItems := autismDevResultAnalysisTargetItems(emergingItems, failedItems)
		out = append(out, autismDevResultAnalysisPromptDomain{
			DomainCode:        strings.TrimSpace(domain.DomainCode),
			DomainName:        strings.TrimSpace(domain.DomainName),
			ScoreType:         strings.TrimSpace(domain.ScoreType),
			ItemCount:         domain.ItemCount,
			AnsweredItemCount: domain.AnsweredItemCount,
			MissingItemCount:  domain.MissingItemCount,
			Complete:          domain.Complete,
			PCount:            domain.PCount,
			ECount:            domain.ECount,
			FCount:            domain.FCount,
			XCount:            domain.XCount,
			PECount:           domain.PECount,
			RawScore:          domain.RawScore,
			ScorableItemCount: domain.ScorableItemCount,
			ScoreRate:         domain.ScoreRate,
			DevelopmentLevel:  autismDevResultAnalysisDevelopmentLevel(passedItems),
			PassedItems:       autismDevResultAnalysisSelectItems(passedItems, 8, true),
			EmergingItems:     autismDevResultAnalysisSelectItems(emergingItems, 8, false),
			FailedItems:       autismDevResultAnalysisSelectItems(failedItems, 8, false),
			TargetItems:       autismDevResultAnalysisSelectItems(targetItems, autismDevResultAnalysisPromptTargetLimit(targetItems), false),
			Warnings:          append([]string(nil), domain.Warnings...),
		})
	}
	return out
}

func autismDevResultAnalysisDomainCodes(record model.AssessmentRecordDetailVO, result autismdevscore.AssessmentResult, data autismDevStaticData) []string {
	pef := make(map[string]bool, len(result.Domains))
	for _, domain := range result.Domains {
		if strings.EqualFold(strings.TrimSpace(domain.ScoreType), autismdevscore.ScoreTypePEF) {
			pef[strings.TrimSpace(domain.DomainCode)] = true
		}
	}
	scope := autismDevTrainingCurrentRecordDomains(record, data)
	if len(scope) == 0 {
		for _, domain := range result.Domains {
			code := strings.TrimSpace(domain.DomainCode)
			if pef[code] && domain.AnsweredItemCount > 0 {
				scope = append(scope, code)
			}
		}
	}
	if len(scope) == 0 {
		scope = append(scope, autismDevResultAnalysisDomains...)
	}
	scoped := make(map[string]bool, len(scope))
	for _, code := range scope {
		scoped[strings.TrimSpace(code)] = true
	}
	out := make([]string, 0, len(scope))
	seen := make(map[string]bool, len(scope))
	for _, code := range autismDevResultAnalysisDomains {
		if !scoped[code] || !pef[code] || seen[code] {
			continue
		}
		seen[code] = true
		out = append(out, code)
	}
	return out
}

func autismDevResultAnalysisScoredItems(items []autismdevscore.ItemDefinition, itemScores map[int]string) ([]autismDevResultAnalysisPromptItem, []autismDevResultAnalysisPromptItem, []autismDevResultAnalysisPromptItem) {
	passed := make([]autismDevResultAnalysisPromptItem, 0)
	emerging := make([]autismDevResultAnalysisPromptItem, 0)
	failed := make([]autismDevResultAnalysisPromptItem, 0)
	for _, item := range items {
		score := normalizeAutismDevScore(itemScores[item.ItemNo])
		if score == "" || score == autismdevscore.ScoreX {
			continue
		}
		promptItem := autismDevResultAnalysisPromptItemFromDefinition(item, score)
		switch score {
		case autismdevscore.ScoreP:
			passed = append(passed, promptItem)
		case autismdevscore.ScoreE:
			emerging = append(emerging, promptItem)
		case autismdevscore.ScoreF:
			failed = append(failed, promptItem)
		}
	}
	return passed, emerging, failed
}

func autismDevResultAnalysisPromptItemFromDefinition(item autismdevscore.ItemDefinition, score string) autismDevResultAnalysisPromptItem {
	return autismDevResultAnalysisPromptItem{
		ItemNo:          item.ItemNo,
		DomainItemNo:    item.DomainItemNo,
		Title:           autismDevResultAnalysisItemTitle(item),
		Score:           score,
		AgeSegment:      strings.TrimSpace(item.AgeSegment),
		AgeMinMonth:     item.AgeMinMonth,
		AgeMaxMonth:     item.AgeMaxMonth,
		AssessmentRange: strings.TrimSpace(item.AssessmentRange),
		PassCriteria:    autismDevResultAnalysisPassCriterion(item.PassCriteria),
	}
}

func autismDevResultAnalysisTargetItems(emergingItems, failedItems []autismDevResultAnalysisPromptItem) []autismDevResultAnalysisPromptItem {
	out := make([]autismDevResultAnalysisPromptItem, 0, len(emergingItems)+len(failedItems))
	out = append(out, emergingItems...)
	out = append(out, failedItems...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			if out[i].Score == autismdevscore.ScoreE {
				return true
			}
			if out[j].Score == autismdevscore.ScoreE {
				return false
			}
		}
		if out[i].AgeMinMonth != out[j].AgeMinMonth {
			return out[i].AgeMinMonth < out[j].AgeMinMonth
		}
		if out[i].DomainItemNo != out[j].DomainItemNo {
			return out[i].DomainItemNo < out[j].DomainItemNo
		}
		return out[i].ItemNo < out[j].ItemNo
	})
	return out
}

func autismDevResultAnalysisPromptTargetLimit(items []autismDevResultAnalysisPromptItem) int {
	count := len(items)
	switch {
	case count <= 0:
		return 0
	case count <= 3:
		return count
	case count <= 5:
		return count
	default:
		return 7
	}
}

func autismDevResultAnalysisSelectItems(items []autismDevResultAnalysisPromptItem, limit int, preferHigherAge bool) []autismDevResultAnalysisPromptItem {
	if len(items) == 0 || limit <= 0 {
		return nil
	}
	ordered := append([]autismDevResultAnalysisPromptItem(nil), items...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if !preferHigherAge && ordered[i].Score != ordered[j].Score {
			if ordered[i].Score == autismdevscore.ScoreE {
				return true
			}
			if ordered[j].Score == autismdevscore.ScoreE {
				return false
			}
		}
		if ordered[i].AgeMaxMonth != ordered[j].AgeMaxMonth {
			if preferHigherAge {
				return ordered[i].AgeMaxMonth > ordered[j].AgeMaxMonth
			}
			return ordered[i].AgeMaxMonth < ordered[j].AgeMaxMonth
		}
		if ordered[i].DomainItemNo != ordered[j].DomainItemNo {
			if preferHigherAge {
				return ordered[i].DomainItemNo > ordered[j].DomainItemNo
			}
			return ordered[i].DomainItemNo < ordered[j].DomainItemNo
		}
		return ordered[i].ItemNo < ordered[j].ItemNo
	})
	if len(ordered) > limit {
		ordered = ordered[:limit]
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].DomainItemNo < ordered[j].DomainItemNo
	})
	return ordered
}

func autismDevResultAnalysisDevelopmentLevel(passedItems []autismDevResultAnalysisPromptItem) string {
	maxMonth := 0
	for _, item := range passedItems {
		if item.AgeMaxMonth > maxMonth {
			maxMonth = item.AgeMaxMonth
		}
	}
	return autismDevResultAnalysisAgeLabel(maxMonth)
}

func autismDevResultAnalysisAgeLabel(months int) string {
	if months <= 0 {
		return ""
	}
	years := months / 12
	remainMonths := months % 12
	if years <= 0 {
		return fmt.Sprintf("%d个月", remainMonths)
	}
	if remainMonths == 0 {
		return fmt.Sprintf("%d岁", years)
	}
	return fmt.Sprintf("%d岁%d个月", years, remainMonths)
}

func autismDevResultAnalysisItemTitle(item autismdevscore.ItemDefinition) string {
	for _, value := range []string{item.ItemTitle, item.TestItem, item.AssessmentRange} {
		if text := strings.TrimSpace(value); text != "" {
			return text
		}
	}
	return fmt.Sprintf("%d号项目", item.DomainItemNo)
}

func autismDevResultAnalysisPassCriterion(raw string) string {
	for _, line := range strings.Split(raw, "\n") {
		text := strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToUpper(text), "P-") {
			return text
		}
	}
	return strings.TrimSpace(raw)
}

func callDeepSeekAutismDevResultAnalysisStream(ctx context.Context, payload autismDevResultAnalysisPromptPayload, onDelta func(string) error) (model.AutismDevResultAnalysisVO, error) {
	if onDelta == nil {
		return callDeepSeekAutismDevResultAnalysis(ctx, payload)
	}
	apiKey := strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY"))
	if apiKey == "" {
		apiKey = deepSeekIEPPlanFallbackAPIKey
	}
	if apiKey == "" {
		return model.AutismDevResultAnalysisVO{}, errors.New("DEEPSEEK_API_KEY is not configured")
	}
	endpoint := strings.TrimSpace(os.Getenv("DEEPSEEK_API_BASE_URL"))
	if endpoint == "" {
		endpoint = deepSeekIEPPlanDefaultURL
	}
	body, err := buildDeepSeekAutismDevResultAnalysisRequestBody(payload, true)
	if err != nil {
		return model.AutismDevResultAnalysisVO{}, err
	}
	requestCtx, cancel := context.WithTimeout(ctx, deepSeekIEPPlanTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return model.AutismDevResultAnalysisVO{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := (&http.Client{Timeout: deepSeekIEPPlanTimeout + 5*time.Second}).Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(requestCtx.Err(), context.DeadlineExceeded) {
			return model.AutismDevResultAnalysisVO{}, fmt.Errorf("DeepSeek API 生成超时（%d秒），请稍后重试", int(deepSeekIEPPlanTimeout.Seconds()))
		}
		return model.AutismDevResultAnalysisVO{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		responseBody, _ := io.ReadAll(resp.Body)
		return model.AutismDevResultAnalysisVO{}, fmt.Errorf("DeepSeek API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
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
			return model.AutismDevResultAnalysisVO{}, fmt.Errorf("parse DeepSeek stream chunk: %w", err)
		}
		if chunk.Error != nil && strings.TrimSpace(chunk.Error.Message) != "" {
			return model.AutismDevResultAnalysisVO{}, errors.New(chunk.Error.Message)
		}
		for _, choice := range chunk.Choices {
			if text := choice.Delta.Content; text != "" {
				content.WriteString(text)
				if err := onDelta(text); err != nil {
					return model.AutismDevResultAnalysisVO{}, err
				}
			}
			if text := strings.TrimSpace(choice.Delta.ReasoningContent); text != "" {
				reasoning.WriteString(text)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return model.AutismDevResultAnalysisVO{}, err
	}
	text := strings.TrimSpace(content.String())
	if text == "" {
		if strings.TrimSpace(reasoning.String()) != "" {
			return model.AutismDevResultAnalysisVO{}, errors.New("DeepSeek API 只返回了 reasoning_content，没有返回最终JSON内容，请重试")
		}
		return model.AutismDevResultAnalysisVO{}, errors.New("DeepSeek API returned empty content")
	}
	var result model.AutismDevResultAnalysisVO
	if err := json.Unmarshal([]byte(extractJSONContent(text)), &result); err != nil {
		return model.AutismDevResultAnalysisVO{}, fmt.Errorf("parse DeepSeek AutismDev result analysis JSON: %w", err)
	}
	result.Model = deepSeekIEPPlanModel
	return result, nil
}

func callDeepSeekAutismDevResultAnalysis(ctx context.Context, payload autismDevResultAnalysisPromptPayload) (model.AutismDevResultAnalysisVO, error) {
	apiKey := strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY"))
	if apiKey == "" {
		apiKey = deepSeekIEPPlanFallbackAPIKey
	}
	if apiKey == "" {
		return model.AutismDevResultAnalysisVO{}, errors.New("DEEPSEEK_API_KEY is not configured")
	}
	endpoint := strings.TrimSpace(os.Getenv("DEEPSEEK_API_BASE_URL"))
	if endpoint == "" {
		endpoint = deepSeekIEPPlanDefaultURL
	}
	body, err := buildDeepSeekAutismDevResultAnalysisRequestBody(payload, false)
	if err != nil {
		return model.AutismDevResultAnalysisVO{}, err
	}
	requestCtx, cancel := context.WithTimeout(ctx, deepSeekIEPPlanTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return model.AutismDevResultAnalysisVO{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := (&http.Client{Timeout: deepSeekIEPPlanTimeout + 5*time.Second}).Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(requestCtx.Err(), context.DeadlineExceeded) {
			return model.AutismDevResultAnalysisVO{}, fmt.Errorf("DeepSeek API 生成超时（%d秒），请稍后重试", int(deepSeekIEPPlanTimeout.Seconds()))
		}
		return model.AutismDevResultAnalysisVO{}, err
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.AutismDevResultAnalysisVO{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return model.AutismDevResultAnalysisVO{}, fmt.Errorf("DeepSeek API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}
	var chatResponse deepSeekChatResponse
	if err := json.Unmarshal(responseBody, &chatResponse); err != nil {
		return model.AutismDevResultAnalysisVO{}, err
	}
	if chatResponse.Error != nil && strings.TrimSpace(chatResponse.Error.Message) != "" {
		return model.AutismDevResultAnalysisVO{}, errors.New(chatResponse.Error.Message)
	}
	if len(chatResponse.Choices) == 0 {
		return model.AutismDevResultAnalysisVO{}, errors.New("DeepSeek API returned empty choices")
	}
	content := strings.TrimSpace(chatResponse.Choices[0].Message.Content)
	if content == "" {
		return model.AutismDevResultAnalysisVO{}, errors.New("DeepSeek API returned empty content")
	}
	var result model.AutismDevResultAnalysisVO
	if err := json.Unmarshal([]byte(extractJSONContent(content)), &result); err != nil {
		return model.AutismDevResultAnalysisVO{}, fmt.Errorf("parse DeepSeek AutismDev result analysis JSON: %w", err)
	}
	result.Model = deepSeekIEPPlanModel
	return result, nil
}

func buildDeepSeekAutismDevResultAnalysisRequestBody(payload autismDevResultAnalysisPromptPayload, stream bool) ([]byte, error) {
	return buildDeepSeekIEPPlanRequestBodyWithPrompt(payload, strings.Join([]string{
		"你是儿童康复机构的评估结果分析助手。",
		"任务是把孤独症儿童发展评估表的结构化评分结果转换为报告中的评估结果分析表内容。",
		"写法要贴近纸质样表：能力现状必须写具体题目表现，例如“已能完成A、B，但C、D仍未稳定；目前约达到X左右标准”。发展年龄水平只能作为补充，不能替代题目表现。",
		"优劣势分析写具体已会/未会项目；训练目标写编号条目，每条都是可以训练和观察的具体行为，通常3-6条，最多7条，不能把所有未通过项目全部列进去，也不要让每个领域条数相同。",
		"训练目标要短，类似样表中“1 会使用拇指和食指沿撕纸条。”这种句子；不要输出括号里的评分标准或长解释。",
		"示例风格：感知觉方面已能完成注视光线刺激、叫名字有反应等项目，但灵活追视和辨认常见物品仍未稳定，目前约达到同龄正常儿童6个月左右标准。优势：能注视光线刺激、叫名字有反应。劣势：不能辨认自己的影像、不能辨认常见物品。训练目标：1 能灵活追视移动物体。2 能对突发声音作出反应。",
		"不要把每个领域都写成同一句式；尤其不要统一以“经过观察”“经过评估”开头。",
		"不要使用“能力发展明显超前”“能力发展较好”“能力发展较弱”这类模板句。",
		"必须输出严格JSON，不要Markdown，不要代码块，不要解释。",
		"不得修改输入中的P、E、F、X计数、日期等事实。",
		"不得做医学诊断，不得输出治疗方案。",
	}, "\n"), stream)
}

func normalizeAutismDevResultAnalysis(result model.AutismDevResultAnalysisVO, record model.AssessmentRecordDetailVO, score autismdevscore.AssessmentResult, data autismDevStaticData, itemScores map[int]string, generatedBy string) model.AutismDevResultAnalysisVO {
	fallback := buildRuleBasedAutismDevResultAnalysis(record, score, data, itemScores)
	result.Title = "孤独症儿童评估结果分析表"
	result.Model = deepSeekIEPPlanModel
	result.GeneratedBy = strings.TrimSpace(generatedBy)
	if result.GeneratedBy == "" {
		result.GeneratedBy = "ai"
	}
	if strings.TrimSpace(result.GeneratedAt) == "" {
		result.GeneratedAt = time.Now().Format("2006-01-02 15:04:05")
	}
	byDomain := make(map[string]model.AutismDevResultAnalysisRow, len(result.Rows))
	for _, row := range result.Rows {
		domain := normalizeAutismDevResultAnalysisDomainName(row.Domain, fallback.Rows)
		if domain == "" {
			continue
		}
		byDomain[domain] = row
	}
	rows := make([]model.AutismDevResultAnalysisRow, 0, len(fallback.Rows))
	for _, fallbackRow := range fallback.Rows {
		row, ok := byDomain[fallbackRow.Domain]
		if !ok {
			row = fallbackRow
		}
		row.Domain = fallbackRow.Domain
		if strings.TrimSpace(row.Status) == "" {
			row.Status = fallbackRow.Status
		}
		if strings.TrimSpace(row.Strengths) == "" {
			row.Strengths = fallbackRow.Strengths
		}
		if strings.TrimSpace(row.Weaknesses) == "" {
			row.Weaknesses = fallbackRow.Weaknesses
		}
		if strings.TrimSpace(row.Targets) == "" {
			row.Targets = fallbackRow.Targets
		}
		rows = append(rows, trimAutismDevResultAnalysisRow(row))
	}
	result.Rows = rows
	return result
}

func normalizeAutismDevResultAnalysisDomainName(value string, rows []model.AutismDevResultAnalysisRow) string {
	normalized := strings.ReplaceAll(strings.TrimSpace(value), "和", "与")
	normalized = strings.ReplaceAll(normalized, " ", "")
	if normalized == "" {
		return ""
	}
	for _, row := range rows {
		domain := strings.TrimSpace(row.Domain)
		candidate := strings.ReplaceAll(domain, "和", "与")
		candidate = strings.ReplaceAll(candidate, " ", "")
		if normalized == candidate {
			return domain
		}
	}
	return strings.TrimSpace(value)
}

func buildRuleBasedAutismDevResultAnalysis(record model.AssessmentRecordDetailVO, score autismdevscore.AssessmentResult, data autismDevStaticData, itemScores map[int]string) model.AutismDevResultAnalysisVO {
	domains := autismDevResultAnalysisPromptDomains(record, score, data, itemScores)
	rows := make([]model.AutismDevResultAnalysisRow, 0, len(domains))
	for _, domain := range domains {
		rows = append(rows, buildRuleBasedAutismDevResultAnalysisRow(domain))
	}
	return model.AutismDevResultAnalysisVO{
		Title:       "孤独症儿童评估结果分析表",
		GeneratedBy: "system",
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
		Rows:        rows,
	}
}

func buildRuleBasedAutismDevResultAnalysisRow(domain autismDevResultAnalysisPromptDomain) model.AutismDevResultAnalysisRow {
	status := autismDevResultAnalysisStatusSentence(domain)
	if domain.XCount > 0 {
		status += fmt.Sprintf("其中%d项为X不适用，分析时不计入训练目标。", domain.XCount)
	}

	strengths := "目前稳定独立完成项目较少，需先建立基础反应和任务参与。"
	if titles := autismDevResultAnalysisItemTitles(domain.PassedItems, 5); len(titles) > 0 {
		strengths = "能完成" + autismDevResultAnalysisQuotedList(titles) + "等项目，可作为后续教学的基础。"
	}

	weakItems := append([]autismDevResultAnalysisPromptItem{}, domain.EmergingItems...)
	weakItems = append(weakItems, domain.FailedItems...)
	weaknesses := "未稳定通过项目较多，需结合具体题目安排分步训练。"
	if titles := autismDevResultAnalysisItemTitles(weakItems, 6); len(titles) > 0 {
		weaknesses = "不能稳定完成" + autismDevResultAnalysisQuotedList(titles) + "等项目。"
		if len(domain.EmergingItems) > 0 {
			weaknesses += "其中E项可作为近期优先训练切入点。"
		}
	}

	targets := autismDevResultAnalysisTargetLines(domain.TargetItems, autismDevResultAnalysisTargetLineLimit(domain.TargetItems))
	if strings.TrimSpace(targets) == "" {
		targets = "1 建立本领域基础反应的稳定完成。\n2 提高在少提示条件下完成目标项目的能力。"
	}
	return trimAutismDevResultAnalysisRow(model.AutismDevResultAnalysisRow{
		Domain:     strings.TrimSpace(domain.DomainName),
		Status:     status,
		Strengths:  strengths,
		Weaknesses: weaknesses,
		Targets:    targets,
	})
}

func autismDevResultAnalysisStatusSentence(domain autismDevResultAnalysisPromptDomain) string {
	level := strings.TrimSpace(domain.DevelopmentLevel)
	passedTitles := autismDevResultAnalysisItemTitles(domain.PassedItems, 3)
	weakItems := append([]autismDevResultAnalysisPromptItem{}, domain.EmergingItems...)
	weakItems = append(weakItems, domain.FailedItems...)
	weakTitles := autismDevResultAnalysisItemTitles(weakItems, 3)
	var parts []string
	if len(passedTitles) > 0 {
		parts = append(parts, fmt.Sprintf("%s方面已能完成%s等项目", domain.DomainName, autismDevResultAnalysisQuotedList(passedTitles)))
	} else {
		parts = append(parts, fmt.Sprintf("%s方面目前稳定通过项目较少", domain.DomainName))
	}
	if len(weakTitles) > 0 {
		parts = append(parts, fmt.Sprintf("但%s仍未稳定", autismDevResultAnalysisQuotedList(weakTitles)))
	}
	if level != "" {
		parts = append(parts, fmt.Sprintf("目前约达到同龄正常儿童%s左右标准", level))
	} else {
		parts = append(parts, "发展年龄水平暂不稳定")
	}
	status := strings.Join(parts, "，") + "。"
	notPassedCount := domain.ECount + domain.FCount
	if notPassedCount > len(weakTitles) && notPassedCount > 0 {
		status += fmt.Sprintf("另有%d个项目未达到P标准。", notPassedCount)
	}
	return status
}

func autismDevResultAnalysisItemTitles(items []autismDevResultAnalysisPromptItem, limit int) []string {
	if limit <= 0 || len(items) == 0 {
		return nil
	}
	out := make([]string, 0, limit)
	seen := make(map[string]bool, limit)
	for _, item := range items {
		title := strings.TrimSpace(item.Title)
		if title == "" || seen[title] {
			continue
		}
		seen[title] = true
		out = append(out, title)
		if len(out) >= limit {
			break
		}
	}
	return out
}

func autismDevResultAnalysisQuotedList(values []string) string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		text := strings.TrimSpace(value)
		if text != "" {
			out = append(out, "“"+text+"”")
		}
	}
	return strings.Join(out, "、")
}

func autismDevResultAnalysisTargetLines(items []autismDevResultAnalysisPromptItem, limit int) string {
	if limit <= 0 || len(items) == 0 {
		return ""
	}
	var builder strings.Builder
	count := 0
	seen := make(map[string]bool, limit)
	for _, item := range items {
		title := strings.TrimSpace(item.Title)
		if title == "" || seen[title] {
			continue
		}
		seen[title] = true
		count++
		if builder.Len() > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString(fmt.Sprintf("%d 能%s", count, autismDevResultAnalysisTargetAction(title)))
		builder.WriteString("。")
		if count >= limit {
			break
		}
	}
	return builder.String()
}

func autismDevResultAnalysisTargetAction(title string) string {
	text := strings.TrimSpace(title)
	text = strings.TrimPrefix(text, "能够")
	text = strings.TrimPrefix(text, "能")
	text = strings.TrimPrefix(text, "可以")
	text = strings.TrimPrefix(text, "完成")
	text = strings.TrimSpace(text)
	if text == "" {
		return "完成目标项目"
	}
	runes := []rune(text)
	if len(runes) > 24 {
		text = string(runes[:24])
	}
	return text
}

func autismDevResultAnalysisTargetLineLimit(items []autismDevResultAnalysisPromptItem) int {
	count := len(items)
	switch {
	case count <= 0:
		return 0
	case count <= 7:
		return count
	default:
		return 7
	}
}

func trimAutismDevResultAnalysisRow(row model.AutismDevResultAnalysisRow) model.AutismDevResultAnalysisRow {
	row.Domain = strings.TrimSpace(row.Domain)
	row.Status = trimAutismDevResultAnalysisStatusOpening(row.Status)
	row.Strengths = strings.TrimSpace(strings.TrimPrefix(row.Strengths, "优势："))
	row.Weaknesses = strings.TrimSpace(strings.TrimPrefix(row.Weaknesses, "劣势："))
	row.Targets = normalizeAutismDevResultAnalysisTargets(row.Targets)
	return row
}

func normalizeAutismDevResultAnalysisTargets(targets string) string {
	normalized := strings.ReplaceAll(targets, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	for index := 2; index <= 20; index++ {
		marker := fmt.Sprint(index)
		normalized = strings.ReplaceAll(normalized, "。"+marker, "。\n"+marker)
		normalized = strings.ReplaceAll(normalized, "；"+marker, "；\n"+marker)
		normalized = strings.ReplaceAll(normalized, ";"+marker, ";\n"+marker)
	}
	lines := strings.Split(normalized, "\n")
	out := make([]string, 0, 7)
	for _, line := range lines {
		text := strings.TrimSpace(stripAutismDevParenthetical(line))
		if text == "" {
			continue
		}
		runes := []rune(text)
		if len(runes) > 42 {
			text = strings.TrimRight(string(runes[:42]), "，,；;、 ")
		}
		out = append(out, text)
		if len(out) >= 7 {
			break
		}
	}
	return strings.Join(out, "\n")
}

func stripAutismDevParenthetical(text string) string {
	var builder strings.Builder
	depth := 0
	for _, r := range text {
		switch r {
		case '（', '(':
			depth++
			continue
		case '）', ')':
			if depth > 0 {
				depth--
				continue
			}
		}
		if depth == 0 {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func trimAutismDevResultAnalysisStatusOpening(status string) string {
	text := strings.TrimSpace(status)
	prefixes := []string{
		"经过观察发现，",
		"经过观察发现,",
		"经过评估发现，",
		"经过评估发现,",
		"经过观察，",
		"经过观察,",
		"经过评估，",
		"经过评估,",
		"经观察，",
		"经观察,",
		"经评估，",
		"经评估,",
	}
	changed := true
	for changed {
		changed = false
		for _, prefix := range prefixes {
			if strings.HasPrefix(text, prefix) {
				text = strings.TrimSpace(strings.TrimPrefix(text, prefix))
				changed = true
			}
		}
	}
	return text
}
