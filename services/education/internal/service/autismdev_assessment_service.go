package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"go-migration-platform/pkg/autismdevscore"
	"go-migration-platform/services/education/internal/model"
)

const (
	autismDevScaleCode       = autismdevscore.ScaleCode
	autismDevScaleVersion    = autismdevscore.DefaultScaleVersion
	autismDevAssessmentName  = "孤独症儿童发展评估表"
	autismDevItemBankFile    = "autismdev-item-bank-draft.json"
	autismDevDomainMapFile   = "autismdev-domain-map.json"
	autismDevMetadataFile    = "autismdev-scale-metadata.json"
	autismDevStaticRevision  = "draft-2026-05-14-title-cleanup"
	autismDevDraftDataStatus = "draft"
)

type AutismDevScoreDataInfo struct {
	ScaleCode      string   `json:"scaleCode"`
	ScaleVersion   string   `json:"scaleVersion"`
	SourceStandard string   `json:"sourceStandard,omitempty"`
	SourcePDF      string   `json:"sourcePdf,omitempty"`
	DataStatus     string   `json:"dataStatus,omitempty"`
	Sources        []string `json:"sources"`
}

type AutismDevScoreResponse struct {
	AutismDevScoreDataInfo
	Result autismdevscore.AssessmentResult `json:"result"`
}

type autismDevStaticData struct {
	metadata   autismDevScaleMetadata
	items      []autismdevscore.ItemDefinition
	domains    []autismDevDomainDefinition
	sources    []string
	dataStatus string
}

type autismDevScaleMetadata struct {
	ScaleCode      string `json:"scale_code"`
	ScaleName      string `json:"scale_name"`
	ScaleVersion   string `json:"scale_version"`
	SourcePDF      string `json:"source_pdf"`
	SourceStandard string `json:"source_standard"`
	ItemCount      int    `json:"item_count"`
	DomainCount    int    `json:"domain_count"`
	DataStatus     string `json:"data_status"`
	ScoringNote    string `json:"scoring_note"`
}

type autismDevDomainDefinition struct {
	ScaleCode string `json:"scale_code"`
	ScaleName string `json:"scale_name"`
	SortNo    int    `json:"sort_no"`
	ItemCount int    `json:"item_count"`
	ScoreType string `json:"score_type"`
}

var (
	autismDevStaticDataMu             sync.RWMutex
	autismDevStaticDataCache          autismDevStaticData
	autismDevStaticDataCacheSignature string
	autismDevStaticDataCacheReady     bool
)

func (svc *Service) ScoreAutismDev(input autismdevscore.AssessmentInput) (AutismDevScoreResponse, error) {
	engine, info, err := loadAutismDevEngine()
	if err != nil {
		return AutismDevScoreResponse{}, err
	}
	result, err := engine.Score(input)
	if err != nil {
		return AutismDevScoreResponse{}, err
	}
	return AutismDevScoreResponse{
		AutismDevScoreDataInfo: info,
		Result:                 result,
	}, nil
}

func (svc *Service) GetAutismDevAssessmentFormTemplate() (model.AutismDevAssessmentFormTemplateVO, error) {
	return buildAutismDevAssessmentFormTemplate()
}

func (svc *Service) GetAutismDevAssessmentFormTemplateSummary() (model.AutismDevAssessmentFormTemplateSummaryVO, error) {
	return buildAutismDevAssessmentFormTemplateSummary()
}

func (svc *Service) GetAutismDevAssessmentFormTemplateItem(itemNo int) (model.AutismDevAssessmentItem, error) {
	if item, loaded, err := loadAutismDevTemplateItemFromConfiguredDB(itemNo); loaded || err != nil {
		if err != nil {
			return model.AutismDevAssessmentItem{}, err
		}
		return buildAutismDevAssessmentItem(item), nil
	}
	data, err := loadAutismDevStaticData()
	if err != nil {
		return model.AutismDevAssessmentItem{}, err
	}
	for _, item := range data.items {
		if item.ItemNo == itemNo {
			return buildAutismDevAssessmentItem(item), nil
		}
	}
	return model.AutismDevAssessmentItem{}, fmt.Errorf("AutismDev item %d not found", itemNo)
}

func loadAutismDevEngine() (*autismdevscore.Engine, AutismDevScoreDataInfo, error) {
	return buildAutismDevEngine()
}

func buildAutismDevEngine() (*autismdevscore.Engine, AutismDevScoreDataInfo, error) {
	data, err := loadAutismDevStaticData()
	if err != nil {
		return nil, AutismDevScoreDataInfo{}, err
	}
	engine, err := autismdevscore.NewEngine(data.items)
	if err != nil {
		return nil, AutismDevScoreDataInfo{}, fmt.Errorf("build AutismDev score engine: %w", err)
	}
	return engine, autismDevScoreDataInfo(data), nil
}

func buildAutismDevAssessmentFormTemplate() (model.AutismDevAssessmentFormTemplateVO, error) {
	data, err := loadAutismDevStaticData()
	if err != nil {
		return model.AutismDevAssessmentFormTemplateVO{}, err
	}
	return model.AutismDevAssessmentFormTemplateVO{
		TemplateCode:    "AUTISMDEV_ASSESSMENT_FORM",
		TemplateVersion: autismDevScaleVersion,
		Title:           "孤独症儿童发展评估表录入表",
		ScaleCode:       autismDevScaleCode,
		ScaleVersion:    autismDevScaleVersion,
		SourceStandard:  strings.TrimSpace(data.metadata.SourceStandard),
		SourcePDF:       strings.TrimSpace(data.metadata.SourcePDF),
		DataStatus:      data.dataStatus,
		Sources:         data.sources,
		ItemCount:       len(data.items),
		ScoreOptions:    autismDevScoreOptions(),
		BasicFields:     autismDevAssessmentBasicFields(),
		Domains:         autismDevAssessmentDomains(data.domains),
		DomainGroups:    autismDevAssessmentDomainGroups(data.items, data.domains),
		SubmitContract:  autismDevSubmitContract(),
	}, nil
}

func buildAutismDevAssessmentFormTemplateSummary() (model.AutismDevAssessmentFormTemplateSummaryVO, error) {
	data, err := loadAutismDevStaticData()
	if err != nil {
		return model.AutismDevAssessmentFormTemplateSummaryVO{}, err
	}
	return model.AutismDevAssessmentFormTemplateSummaryVO{
		TemplateCode:    "AUTISMDEV_ASSESSMENT_FORM",
		TemplateVersion: autismDevScaleVersion,
		Title:           "孤独症儿童发展评估表录入表",
		ScaleCode:       autismDevScaleCode,
		ScaleVersion:    autismDevScaleVersion,
		SourceStandard:  strings.TrimSpace(data.metadata.SourceStandard),
		SourcePDF:       strings.TrimSpace(data.metadata.SourcePDF),
		DataStatus:      data.dataStatus,
		Sources:         data.sources,
		ItemCount:       len(data.items),
		ScoreOptions:    autismDevScoreOptions(),
		BasicFields:     autismDevAssessmentBasicFields(),
		Domains:         autismDevAssessmentDomains(data.domains),
		DomainGroups:    autismDevAssessmentDomainGroupSummaries(data.items, data.domains),
		SubmitContract:  autismDevSubmitContract(),
	}, nil
}

func loadAutismDevStaticData() (autismDevStaticData, error) {
	if data, loaded, err := loadAutismDevStaticDataFromConfiguredDB(); loaded || err != nil {
		if err != nil {
			return autismDevStaticData{}, err
		}
		return data, nil
	}
	return loadAutismDevFallbackStaticData()
}

func loadAutismDevFallbackStaticData() (autismDevStaticData, error) {
	dataDir, err := resolveAutismDevDataDir()
	if err != nil {
		return autismDevStaticData{}, err
	}
	signature, err := autismDevDataFilesSignature(dataDir)
	if err != nil {
		return autismDevStaticData{}, err
	}

	autismDevStaticDataMu.RLock()
	if autismDevStaticDataCacheReady &&
		autismDevStaticDataCacheSignature == signature {
		data := autismDevStaticDataCache
		autismDevStaticDataMu.RUnlock()
		return data, nil
	}
	autismDevStaticDataMu.RUnlock()

	autismDevStaticDataMu.Lock()
	defer autismDevStaticDataMu.Unlock()
	if autismDevStaticDataCacheReady &&
		autismDevStaticDataCacheSignature == signature {
		return autismDevStaticDataCache, nil
	}

	data, err := loadAutismDevStaticDataFromFiles(dataDir)
	if err != nil {
		return autismDevStaticData{}, err
	}
	autismDevStaticDataCache = data
	autismDevStaticDataCacheSignature = signature
	autismDevStaticDataCacheReady = true
	return data, nil
}

func resetAutismDevFallbackStaticDataCache() {
	autismDevStaticDataMu.Lock()
	defer autismDevStaticDataMu.Unlock()
	autismDevStaticDataCache = autismDevStaticData{}
	autismDevStaticDataCacheSignature = ""
	autismDevStaticDataCacheReady = false
}

func loadAutismDevStaticDataFromFiles(dataDir string) (autismDevStaticData, error) {
	items, err := autismdevscore.LoadItemDefinitionsFile(filepath.Join(dataDir, autismDevItemBankFile))
	if err != nil {
		return autismDevStaticData{}, fmt.Errorf("load AutismDev item bank: %w", err)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ItemNo < items[j].ItemNo })

	var domains []autismDevDomainDefinition
	if err := loadAutismDevJSONFile(filepath.Join(dataDir, autismDevDomainMapFile), &domains); err != nil {
		return autismDevStaticData{}, fmt.Errorf("load AutismDev domain map: %w", err)
	}
	sort.Slice(domains, func(i, j int) bool { return domains[i].SortNo < domains[j].SortNo })

	var metadata autismDevScaleMetadata
	if err := loadAutismDevJSONFile(filepath.Join(dataDir, autismDevMetadataFile), &metadata); err != nil {
		return autismDevStaticData{}, fmt.Errorf("load AutismDev metadata: %w", err)
	}

	sources, dataStatus := autismDevDataSources(metadata.DataStatus)
	return autismDevStaticData{
		metadata:   metadata,
		items:      items,
		domains:    domains,
		sources:    sources,
		dataStatus: dataStatus,
	}, nil
}

func loadAutismDevJSONFile(path string, out any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("decode %s: %w", filepath.Base(path), err)
	}
	return nil
}

func autismDevDataFilesSignature(dataDir string) (string, error) {
	var builder strings.Builder
	for _, name := range []string{autismDevItemBankFile, autismDevDomainMapFile, autismDevMetadataFile} {
		path := filepath.Join(dataDir, name)
		info, err := os.Stat(path)
		if err != nil {
			return "", err
		}
		if info.IsDir() {
			return "", fmt.Errorf("AutismDev data file %s is a directory", path)
		}
		builder.WriteString(name)
		builder.WriteString(":")
		builder.WriteString(fmt.Sprint(info.Size()))
		builder.WriteString(":")
		builder.WriteString(fmt.Sprint(info.ModTime().UnixNano()))
		builder.WriteString(";")
	}
	return builder.String(), nil
}

func resolveAutismDevDataDir() (string, error) {
	if raw := os.Getenv("AUTISMDEV_DATA_DIR"); raw != "" {
		if err := requireAutismDevDataFiles(raw); err != nil {
			return "", err
		}
		return raw, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for dir := cwd; ; dir = filepath.Dir(dir) {
		candidate := filepath.Join(dir, "docs")
		if requireAutismDevDataFiles(candidate) == nil {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
	return "", fmt.Errorf("AutismDev data files not found; set AUTISMDEV_DATA_DIR to the directory containing %s", autismDevItemBankFile)
}

func requireAutismDevDataFiles(dir string) error {
	for _, name := range []string{autismDevItemBankFile, autismDevDomainMapFile, autismDevMetadataFile} {
		if !fileExists(filepath.Join(dir, name)) {
			return fmt.Errorf("AutismDev data file %s not found in %s", name, dir)
		}
	}
	return nil
}

func autismDevDataSources(metadataStatus string) ([]string, string) {
	dataStatus := strings.TrimSpace(metadataStatus)
	if dataStatus == "" {
		dataStatus = autismDevDraftDataStatus
	}
	return []string{
		autismDevItemBankFile,
		autismDevDomainMapFile,
		autismDevMetadataFile,
		"revision:" + autismDevStaticRevision,
	}, dataStatus
}

func autismDevScoreDataInfo(data autismDevStaticData) AutismDevScoreDataInfo {
	scaleVersion := strings.TrimSpace(data.metadata.ScaleVersion)
	if scaleVersion == "" {
		scaleVersion = autismDevScaleVersion
	}
	return AutismDevScoreDataInfo{
		ScaleCode:      autismDevScaleCode,
		ScaleVersion:   scaleVersion,
		SourceStandard: strings.TrimSpace(data.metadata.SourceStandard),
		SourcePDF:      strings.TrimSpace(data.metadata.SourcePDF),
		DataStatus:     data.dataStatus,
		Sources:        append([]string(nil), data.sources...),
	}
}

func autismDevAssessmentBasicFields() []model.PEP3AssessmentFormField {
	return []model.PEP3AssessmentFormField{
		{Key: "studentId", Label: "学员ID", FieldType: "number", Required: false},
		{Key: "studentName", Label: "儿童姓名", FieldType: "text", Required: false},
		{Key: "examinerName", Label: "评估员姓名", FieldType: "text", Required: false},
		{Key: "birthDate", Label: "出生日期", FieldType: "date", Required: true, Placeholder: "YYYY-MM-DD"},
		{Key: "assessmentDate", Label: "评估日期", FieldType: "date", Required: true, Placeholder: "YYYY-MM-DD"},
		{Key: "remark", Label: "备注", FieldType: "textarea", Required: false},
	}
}

func autismDevScoreOptions() []model.AutismDevScoreOption {
	return []model.AutismDevScoreOption{
		{Value: autismdevscore.ScoreP, Label: "P 通过", Description: "通过，记1分", ScoreType: autismdevscore.ScoreTypePEF},
		{Value: autismdevscore.ScoreE, Label: "E 中间反应", Description: "不计分，可作为训练目标", ScoreType: autismdevscore.ScoreTypePEF},
		{Value: autismdevscore.ScoreF, Label: "F 不通过", Description: "不通过，记0分", ScoreType: autismdevscore.ScoreTypePEF},
		{Value: autismdevscore.ScoreX, Label: "X 不适用", Description: "不适用，不计分", ScoreType: autismdevscore.ScoreTypePEF},
		{Value: autismdevscore.ScoreA, Label: "A 无异常", Description: "没有异常", ScoreType: autismdevscore.ScoreTypeAMS},
		{Value: autismdevscore.ScoreM, Label: "M 轻度异常", Description: "轻度异常", ScoreType: autismdevscore.ScoreTypeAMS},
		{Value: autismdevscore.ScoreS, Label: "S 重度异常", Description: "重度异常", ScoreType: autismdevscore.ScoreTypeAMS},
	}
}

func autismDevScoreOptionsForType(scoreType string) []model.AutismDevScoreOption {
	options := autismDevScoreOptions()
	out := make([]model.AutismDevScoreOption, 0, len(options))
	for _, option := range options {
		if strings.EqualFold(option.ScoreType, scoreType) {
			out = append(out, option)
		}
	}
	return out
}

func autismDevSubmitContract() model.AutismDevSubmitContract {
	return model.AutismDevSubmitContract{
		ScoreEndpoint:        "/api/v1/assessments/autismdev/score",
		CreateRecordEndpoint: "/api/v1/assessments/autismdev/records/create",
		DateFormat:           "YYYY-MM-DD",
		ItemScoreListKey:     "itemScoreList",
		RequiredBaseFields:   []string{"birthDate", "assessmentDate"},
		AllowedItemScores: []string{
			autismdevscore.ScoreP,
			autismdevscore.ScoreE,
			autismdevscore.ScoreF,
			autismdevscore.ScoreX,
			autismdevscore.ScoreA,
			autismdevscore.ScoreM,
			autismdevscore.ScoreS,
		},
	}
}

func autismDevAssessmentDomains(domains []autismDevDomainDefinition) []model.AutismDevAssessmentDomain {
	out := make([]model.AutismDevAssessmentDomain, 0, len(domains))
	for _, domain := range domains {
		domainCode := strings.TrimSpace(domain.ScaleCode)
		out = append(out, model.AutismDevAssessmentDomain{
			DomainCode: domainCode,
			DomainName: autismDevDomainDisplayName(domain.ScaleName, domainCode),
			SortNo:     domain.SortNo,
			ItemCount:  domain.ItemCount,
			ScoreType:  strings.TrimSpace(domain.ScoreType),
		})
	}
	return out
}

func autismDevAssessmentDomainGroups(items []autismdevscore.ItemDefinition, domains []autismDevDomainDefinition) []model.AutismDevAssessmentItemGroup {
	itemsByDomain := autismDevItemsByDomain(items)
	groups := make([]model.AutismDevAssessmentItemGroup, 0, len(domains))
	for _, domain := range domains {
		domainCode := strings.TrimSpace(domain.ScaleCode)
		domainName := autismDevDomainDisplayName(domain.ScaleName, domainCode)
		groupItems := make([]model.AutismDevAssessmentItem, 0, len(itemsByDomain[domainCode]))
		for _, item := range itemsByDomain[domainCode] {
			groupItems = append(groupItems, buildAutismDevAssessmentItem(item))
		}
		groups = append(groups, model.AutismDevAssessmentItemGroup{
			GroupCode:  "domain_" + strings.ToLower(domainCode),
			Title:      domainName,
			DomainCode: domainCode,
			DomainName: domainName,
			ScoreType:  strings.TrimSpace(domain.ScoreType),
			ItemCount:  len(groupItems),
			Items:      groupItems,
		})
	}
	return groups
}

func autismDevAssessmentDomainGroupSummaries(items []autismdevscore.ItemDefinition, domains []autismDevDomainDefinition) []model.AutismDevAssessmentItemGroupSummary {
	itemsByDomain := autismDevItemsByDomain(items)
	groups := make([]model.AutismDevAssessmentItemGroupSummary, 0, len(domains))
	for _, domain := range domains {
		domainCode := strings.TrimSpace(domain.ScaleCode)
		domainName := autismDevDomainDisplayName(domain.ScaleName, domainCode)
		groupItems := make([]model.AutismDevAssessmentItemSummary, 0, len(itemsByDomain[domainCode]))
		for _, item := range itemsByDomain[domainCode] {
			groupItems = append(groupItems, buildAutismDevAssessmentItemSummary(item))
		}
		groups = append(groups, model.AutismDevAssessmentItemGroupSummary{
			GroupCode:  "domain_" + strings.ToLower(domainCode),
			Title:      domainName,
			DomainCode: domainCode,
			DomainName: domainName,
			ScoreType:  strings.TrimSpace(domain.ScoreType),
			ItemCount:  len(groupItems),
			Items:      groupItems,
		})
	}
	return groups
}

func buildAutismDevAssessmentItem(item autismdevscore.ItemDefinition) model.AutismDevAssessmentItem {
	domainCode := strings.TrimSpace(item.DomainCode)
	return model.AutismDevAssessmentItem{
		ItemNo:          item.ItemNo,
		DomainItemNo:    item.DomainItemNo,
		ItemTitle:       strings.TrimSpace(item.ItemTitle),
		TestItem:        strings.TrimSpace(nonEmptyString(item.TestItem, item.ItemTitle)),
		AssessmentRange: strings.TrimSpace(item.AssessmentRange),
		Materials:       strings.TrimSpace(item.Materials),
		AgeSegment:      strings.TrimSpace(item.AgeSegment),
		AgeMinMonth:     item.AgeMinMonth,
		AgeMaxMonth:     item.AgeMaxMonth,
		DomainCode:      domainCode,
		DomainName:      autismDevDomainDisplayName(item.DomainName, domainCode),
		ScoreType:       strings.TrimSpace(item.ScoreType),
		ScoreOptions:    autismDevScoreOptionsForType(item.ScoreType),
		AssessmentModes: append([]string(nil), item.AssessmentModes...),
		Method:          strings.TrimSpace(item.Method),
		PassCriteria:    strings.TrimSpace(item.PassCriteria),
		SourcePDF:       strings.TrimSpace(item.SourcePDF),
		SourcePages:     append([]int(nil), item.SourcePages...),
		OCRStatus:       strings.TrimSpace(item.OCRStatus),
	}
}

func buildAutismDevAssessmentItemSummary(item autismdevscore.ItemDefinition) model.AutismDevAssessmentItemSummary {
	domainCode := strings.TrimSpace(item.DomainCode)
	return model.AutismDevAssessmentItemSummary{
		ItemNo:          item.ItemNo,
		DomainItemNo:    item.DomainItemNo,
		ItemTitle:       strings.TrimSpace(item.ItemTitle),
		AssessmentRange: strings.TrimSpace(item.AssessmentRange),
		Materials:       strings.TrimSpace(item.Materials),
		Method:          strings.TrimSpace(item.Method),
		PassCriteria:    strings.TrimSpace(item.PassCriteria),
		AgeSegment:      strings.TrimSpace(item.AgeSegment),
		AgeMinMonth:     item.AgeMinMonth,
		AgeMaxMonth:     item.AgeMaxMonth,
		DomainCode:      domainCode,
		DomainName:      autismDevDomainDisplayName(item.DomainName, domainCode),
		ScoreType:       strings.TrimSpace(item.ScoreType),
		AssessmentModes: append([]string(nil), item.AssessmentModes...),
	}
}

func autismDevItemsByDomain(items []autismdevscore.ItemDefinition) map[string][]autismdevscore.ItemDefinition {
	out := make(map[string][]autismdevscore.ItemDefinition, len(autismdevscore.DomainOrder))
	for _, item := range items {
		out[item.DomainCode] = append(out[item.DomainCode], item)
	}
	for domainCode := range out {
		sort.Slice(out[domainCode], func(i, j int) bool { return out[domainCode][i].ItemNo < out[domainCode][j].ItemNo })
	}
	return out
}

func autismDevDomainDisplayName(raw, domainCode string) string {
	name := strings.TrimSpace(raw)
	code := strings.TrimSpace(domainCode)
	if name == "" || code == "" {
		return name
	}
	escapedCode := regexp.QuoteMeta(code)
	cleaned := regexp.MustCompile(`(?i)^\s*`+escapedCode+`\s*[-_/：:]*\s*`).ReplaceAllString(name, "")
	cleaned = regexp.MustCompile(`(?i)\s*[（(]\s*`+escapedCode+`\s*[）)]\s*$`).ReplaceAllString(cleaned, "")
	cleaned = regexp.MustCompile(`(?i)\s*[-_/：:]*\s*`+escapedCode+`\s*$`).ReplaceAllString(cleaned, "")
	cleaned = strings.TrimSpace(cleaned)
	if cleaned == "" {
		return name
	}
	return cleaned
}
