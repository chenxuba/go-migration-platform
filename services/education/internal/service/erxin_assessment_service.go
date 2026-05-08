package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"go-migration-platform/pkg/erxinscore"
	"go-migration-platform/services/education/internal/model"
)

const (
	erxinScaleCode       = erxinscore.ScaleCode
	erxinScaleVersion    = erxinscore.DefaultScaleVersion
	erxinAssessmentName  = "0岁～6岁儿童发育行为评估量表（儿心量表-II）"
	erxinItemBankFile    = "erxin-item-bank-draft.json"
	erxinDomainMapFile   = "erxin-domain-map.json"
	erxinAgeBandsFile    = "erxin-age-bands.json"
	erxinMetadataFile    = "erxin-scale-metadata.json"
	erxinStaticRevision  = "draft-2026-05-08"
	erxinDraftDataStatus = "draft"
)

type ERXinScoreDataInfo struct {
	ScaleCode      string   `json:"scaleCode"`
	ScaleVersion   string   `json:"scaleVersion"`
	SourceStandard string   `json:"sourceStandard,omitempty"`
	SourcePDF      string   `json:"sourcePdf,omitempty"`
	DataStatus     string   `json:"dataStatus,omitempty"`
	Sources        []string `json:"sources"`
}

type ERXinScoreResponse struct {
	ERXinScoreDataInfo
	Result erxinscore.AssessmentResult `json:"result"`
}

type erxinStaticData struct {
	metadata   erxinScaleMetadata
	items      []erxinscore.ItemDefinition
	domains    []erxinDomainDefinition
	ageBands   []erxinAgeBandDefinition
	sources    []string
	dataStatus string
}

type erxinScaleMetadata struct {
	ScaleCode      string `json:"scale_code"`
	ScaleName      string `json:"scale_name"`
	ScaleVersion   string `json:"scale_version"`
	SourcePDF      string `json:"source_pdf"`
	SourceStandard string `json:"source_standard"`
	ItemCount      int    `json:"item_count"`
	DomainCount    int    `json:"domain_count"`
}

type erxinDomainDefinition struct {
	ScaleCode string `json:"scale_code"`
	ScaleName string `json:"scale_name"`
	SortNo    int    `json:"sort_no"`
}

type erxinAgeBandDefinition struct {
	AgeMonth         int     `json:"age_month"`
	Segment          string  `json:"segment"`
	DomainTotalScore float64 `json:"domain_total_score"`
}

var (
	erxinEngineOnce    sync.Once
	erxinEngine        *erxinscore.Engine
	erxinEngineInfo    ERXinScoreDataInfo
	erxinEngineLoadErr error
)

func (svc *Service) ScoreERXin(input erxinscore.AssessmentInput) (ERXinScoreResponse, error) {
	engine, info, err := loadERXinEngine()
	if err != nil {
		return ERXinScoreResponse{}, err
	}
	result, err := engine.Score(input)
	if err != nil {
		return ERXinScoreResponse{}, err
	}
	return ERXinScoreResponse{
		ERXinScoreDataInfo: info,
		Result:             result,
	}, nil
}

func (svc *Service) GetERXinAssessmentFormTemplate() (model.ERXinAssessmentFormTemplateVO, error) {
	return buildERXinAssessmentFormTemplate()
}

func (svc *Service) GetERXinAssessmentFormTemplateSummary() (model.ERXinAssessmentFormTemplateSummaryVO, error) {
	return buildERXinAssessmentFormTemplateSummary()
}

func (svc *Service) GetERXinAssessmentFormTemplateItem(itemNo int) (model.ERXinAssessmentItem, error) {
	data, err := loadERXinStaticData()
	if err != nil {
		return model.ERXinAssessmentItem{}, err
	}
	for _, item := range data.items {
		if item.ItemNo == itemNo {
			return buildERXinAssessmentItem(item), nil
		}
	}
	return model.ERXinAssessmentItem{}, fmt.Errorf("ERXin item %d not found", itemNo)
}

func loadERXinEngine() (*erxinscore.Engine, ERXinScoreDataInfo, error) {
	erxinEngineOnce.Do(func() {
		erxinEngine, erxinEngineInfo, erxinEngineLoadErr = buildERXinEngine()
	})
	if erxinEngineLoadErr != nil {
		return nil, ERXinScoreDataInfo{}, erxinEngineLoadErr
	}
	return erxinEngine, erxinEngineInfo, nil
}

func buildERXinEngine() (*erxinscore.Engine, ERXinScoreDataInfo, error) {
	data, err := loadERXinStaticData()
	if err != nil {
		return nil, ERXinScoreDataInfo{}, err
	}
	engine, err := erxinscore.NewEngine(data.items)
	if err != nil {
		return nil, ERXinScoreDataInfo{}, fmt.Errorf("build ERXin score engine: %w", err)
	}
	return engine, erxinScoreDataInfo(data), nil
}

func buildERXinAssessmentFormTemplate() (model.ERXinAssessmentFormTemplateVO, error) {
	data, err := loadERXinStaticData()
	if err != nil {
		return model.ERXinAssessmentFormTemplateVO{}, err
	}
	return model.ERXinAssessmentFormTemplateVO{
		TemplateCode:    "ERXIN2_ASSESSMENT_FORM",
		TemplateVersion: erxinScaleVersion,
		Title:           "儿心量表-II测评录入表",
		ScaleCode:       erxinScaleCode,
		ScaleVersion:    erxinScaleVersion,
		SourceStandard:  strings.TrimSpace(data.metadata.SourceStandard),
		SourcePDF:       strings.TrimSpace(data.metadata.SourcePDF),
		DataStatus:      data.dataStatus,
		Sources:         data.sources,
		ItemCount:       len(data.items),
		AgeBands:        erxinAgeBands(data.ageBands),
		ScoreOptions:    erxinScoreOptions(),
		BasicFields:     erxinAssessmentBasicFields(),
		Domains:         erxinAssessmentDomains(data.domains),
		AgeGroups:       erxinAssessmentAgeGroups(data.items, data.ageBands),
		SubmitContract:  erxinSubmitContract(),
	}, nil
}

func buildERXinAssessmentFormTemplateSummary() (model.ERXinAssessmentFormTemplateSummaryVO, error) {
	data, err := loadERXinStaticData()
	if err != nil {
		return model.ERXinAssessmentFormTemplateSummaryVO{}, err
	}
	return model.ERXinAssessmentFormTemplateSummaryVO{
		TemplateCode:    "ERXIN2_ASSESSMENT_FORM",
		TemplateVersion: erxinScaleVersion,
		Title:           "儿心量表-II测评录入表",
		ScaleCode:       erxinScaleCode,
		ScaleVersion:    erxinScaleVersion,
		SourceStandard:  strings.TrimSpace(data.metadata.SourceStandard),
		SourcePDF:       strings.TrimSpace(data.metadata.SourcePDF),
		DataStatus:      data.dataStatus,
		Sources:         data.sources,
		ItemCount:       len(data.items),
		AgeBands:        erxinAgeBands(data.ageBands),
		ScoreOptions:    erxinScoreOptions(),
		BasicFields:     erxinAssessmentBasicFields(),
		Domains:         erxinAssessmentDomains(data.domains),
		AgeGroups:       erxinAssessmentAgeGroupSummaries(data.items, data.ageBands),
		SubmitContract:  erxinSubmitContract(),
	}, nil
}

func loadERXinStaticData() (erxinStaticData, error) {
	dataDir, err := resolveERXinDataDir()
	if err != nil {
		return erxinStaticData{}, err
	}

	items, err := erxinscore.LoadItemDefinitionsFile(filepath.Join(dataDir, erxinItemBankFile))
	if err != nil {
		return erxinStaticData{}, fmt.Errorf("load ERXin item bank: %w", err)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ItemNo < items[j].ItemNo })

	var domains []erxinDomainDefinition
	if err := loadERXinJSONFile(filepath.Join(dataDir, erxinDomainMapFile), &domains); err != nil {
		return erxinStaticData{}, fmt.Errorf("load ERXin domain map: %w", err)
	}
	sort.Slice(domains, func(i, j int) bool { return domains[i].SortNo < domains[j].SortNo })

	var ageBands []erxinAgeBandDefinition
	if err := loadERXinJSONFile(filepath.Join(dataDir, erxinAgeBandsFile), &ageBands); err != nil {
		return erxinStaticData{}, fmt.Errorf("load ERXin age bands: %w", err)
	}
	sort.Slice(ageBands, func(i, j int) bool { return ageBands[i].AgeMonth < ageBands[j].AgeMonth })

	var metadata erxinScaleMetadata
	if err := loadERXinJSONFile(filepath.Join(dataDir, erxinMetadataFile), &metadata); err != nil {
		return erxinStaticData{}, fmt.Errorf("load ERXin metadata: %w", err)
	}

	sources, dataStatus := erxinDataSources()
	return erxinStaticData{
		metadata:   metadata,
		items:      items,
		domains:    domains,
		ageBands:   ageBands,
		sources:    sources,
		dataStatus: dataStatus,
	}, nil
}

func loadERXinJSONFile(path string, out any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("decode %s: %w", filepath.Base(path), err)
	}
	return nil
}

func resolveERXinDataDir() (string, error) {
	if raw := os.Getenv("ERXIN_DATA_DIR"); raw != "" {
		if err := requireERXinDataFiles(raw); err != nil {
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
		if requireERXinDataFiles(candidate) == nil {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
	return "", fmt.Errorf("ERXin data files not found; set ERXIN_DATA_DIR to the directory containing %s", erxinItemBankFile)
}

func requireERXinDataFiles(dir string) error {
	for _, name := range []string{erxinItemBankFile, erxinDomainMapFile, erxinAgeBandsFile, erxinMetadataFile} {
		if !fileExists(filepath.Join(dir, name)) {
			return fmt.Errorf("ERXin data file %s not found in %s", name, dir)
		}
	}
	return nil
}

func erxinDataSources() ([]string, string) {
	return []string{
		erxinItemBankFile,
		erxinDomainMapFile,
		erxinAgeBandsFile,
		erxinMetadataFile,
		"revision:" + erxinStaticRevision,
	}, erxinDraftDataStatus
}

func erxinScoreDataInfo(data erxinStaticData) ERXinScoreDataInfo {
	scaleVersion := strings.TrimSpace(data.metadata.ScaleVersion)
	if scaleVersion == "" {
		scaleVersion = erxinScaleVersion
	}
	return ERXinScoreDataInfo{
		ScaleCode:      erxinScaleCode,
		ScaleVersion:   scaleVersion,
		SourceStandard: strings.TrimSpace(data.metadata.SourceStandard),
		SourcePDF:      strings.TrimSpace(data.metadata.SourcePDF),
		DataStatus:     data.dataStatus,
		Sources:        append([]string(nil), data.sources...),
	}
}

func erxinAssessmentBasicFields() []model.PEP3AssessmentFormField {
	return []model.PEP3AssessmentFormField{
		{Key: "studentId", Label: "学员ID", FieldType: "number", Required: false},
		{Key: "studentName", Label: "儿童姓名", FieldType: "text", Required: false},
		{Key: "examinerName", Label: "测试员姓名", FieldType: "text", Required: false},
		{Key: "birthDate", Label: "出生日期", FieldType: "date", Required: true, Placeholder: "YYYY-MM-DD"},
		{Key: "assessmentDate", Label: "测查日期", FieldType: "date", Required: true, Placeholder: "YYYY-MM-DD"},
		{Key: "remark", Label: "备注", FieldType: "textarea", Required: false},
	}
}

func erxinScoreOptions() []model.ERXinScoreOption {
	return []model.ERXinScoreOption{
		{Value: true, Label: "通过", Description: "测查通过，记○"},
		{Value: false, Label: "不通过", Description: "测查不通过，记×"},
	}
}

func erxinSubmitContract() model.ERXinSubmitContract {
	return model.ERXinSubmitContract{
		ScoreEndpoint:         "/api/v1/assessments/erxin/score",
		CreateRecordEndpoint:  "/api/v1/assessments/erxin/records/create",
		DateFormat:            "YYYY-MM-DD",
		ItemPassListKey:       "itemPassList",
		RequiredBaseFields:    []string{"birthDate", "assessmentDate"},
		AllowedItemPassValues: []bool{true, false},
	}
}

func erxinAssessmentDomains(domains []erxinDomainDefinition) []model.ERXinAssessmentDomain {
	out := make([]model.ERXinAssessmentDomain, 0, len(domains))
	for _, domain := range domains {
		out = append(out, model.ERXinAssessmentDomain{
			DomainCode: strings.TrimSpace(domain.ScaleCode),
			DomainName: strings.TrimSpace(domain.ScaleName),
			SortNo:     domain.SortNo,
		})
	}
	return out
}

func erxinAgeBands(ageBands []erxinAgeBandDefinition) []model.ERXinAgeBand {
	out := make([]model.ERXinAgeBand, 0, len(ageBands))
	for _, ageBand := range ageBands {
		out = append(out, model.ERXinAgeBand{
			AgeMonth:         ageBand.AgeMonth,
			Segment:          strings.TrimSpace(ageBand.Segment),
			DomainTotalScore: ageBand.DomainTotalScore,
		})
	}
	return out
}

func erxinAssessmentAgeGroups(items []erxinscore.ItemDefinition, ageBands []erxinAgeBandDefinition) []model.ERXinAssessmentAgeGroup {
	itemsByAge := erxinItemsByAge(items)
	ageBandByMonth := erxinAgeBandByMonth(ageBands)
	groups := make([]model.ERXinAssessmentAgeGroup, 0, len(erxinscore.StandardAgeMonths))
	for _, ageMonth := range erxinscore.StandardAgeMonths {
		ageBand := ageBandByMonth[ageMonth]
		groupItems := make([]model.ERXinAssessmentItem, 0, len(itemsByAge[ageMonth]))
		for _, item := range itemsByAge[ageMonth] {
			groupItems = append(groupItems, buildERXinAssessmentItem(item))
		}
		groups = append(groups, model.ERXinAssessmentAgeGroup{
			GroupCode:        fmt.Sprintf("age_%02d_month", ageMonth),
			Title:            fmt.Sprintf("%d月龄", ageMonth),
			AgeMonth:         ageMonth,
			Segment:          strings.TrimSpace(ageBand.Segment),
			DomainTotalScore: ageBand.DomainTotalScore,
			Items:            groupItems,
		})
	}
	return groups
}

func erxinAssessmentAgeGroupSummaries(items []erxinscore.ItemDefinition, ageBands []erxinAgeBandDefinition) []model.ERXinAssessmentAgeGroupSummary {
	itemsByAge := erxinItemsByAge(items)
	ageBandByMonth := erxinAgeBandByMonth(ageBands)
	groups := make([]model.ERXinAssessmentAgeGroupSummary, 0, len(erxinscore.StandardAgeMonths))
	for _, ageMonth := range erxinscore.StandardAgeMonths {
		ageBand := ageBandByMonth[ageMonth]
		groupItems := make([]model.ERXinAssessmentItemSummary, 0, len(itemsByAge[ageMonth]))
		for _, item := range itemsByAge[ageMonth] {
			groupItems = append(groupItems, buildERXinAssessmentItemSummary(item))
		}
		groups = append(groups, model.ERXinAssessmentAgeGroupSummary{
			GroupCode:        fmt.Sprintf("age_%02d_month", ageMonth),
			Title:            fmt.Sprintf("%d月龄", ageMonth),
			AgeMonth:         ageMonth,
			Segment:          strings.TrimSpace(ageBand.Segment),
			DomainTotalScore: ageBand.DomainTotalScore,
			Items:            groupItems,
		})
	}
	return groups
}

func buildERXinAssessmentItem(item erxinscore.ItemDefinition) model.ERXinAssessmentItem {
	return model.ERXinAssessmentItem{
		ItemNo:                item.ItemNo,
		ItemTitle:             strings.TrimSpace(item.ItemTitle),
		TestItem:              strings.TrimSpace(nonEmptyString(item.TestItem, item.ItemTitle)),
		AgeMonth:              item.AgeMonth,
		AgeSegment:            strings.TrimSpace(item.AgeSegment),
		DomainCode:            strings.TrimSpace(item.DomainCode),
		DomainName:            strings.TrimSpace(item.DomainName),
		ParentReportAllowed:   item.ParentReportAllowed,
		AttentionIfFailed:     item.AttentionIfFailed,
		DomainMonthTotalScore: item.DomainMonthTotalScore,
		ItemWeight:            item.ItemWeight,
		Method:                strings.TrimSpace(item.Method),
		PassCriteria:          strings.TrimSpace(item.PassCriteria),
		SourcePDF:             strings.TrimSpace(item.SourcePDF),
		SourcePages:           append([]int(nil), item.SourcePages...),
		OCRStatus:             strings.TrimSpace(item.OCRStatus),
	}
}

func buildERXinAssessmentItemSummary(item erxinscore.ItemDefinition) model.ERXinAssessmentItemSummary {
	return model.ERXinAssessmentItemSummary{
		ItemNo:              item.ItemNo,
		ItemTitle:           strings.TrimSpace(item.ItemTitle),
		TestItem:            strings.TrimSpace(nonEmptyString(item.TestItem, item.ItemTitle)),
		AgeMonth:            item.AgeMonth,
		AgeSegment:          strings.TrimSpace(item.AgeSegment),
		DomainCode:          strings.TrimSpace(item.DomainCode),
		DomainName:          strings.TrimSpace(item.DomainName),
		ParentReportAllowed: item.ParentReportAllowed,
		AttentionIfFailed:   item.AttentionIfFailed,
	}
}

func erxinItemsByAge(items []erxinscore.ItemDefinition) map[int][]erxinscore.ItemDefinition {
	out := make(map[int][]erxinscore.ItemDefinition)
	for _, item := range items {
		out[item.AgeMonth] = append(out[item.AgeMonth], item)
	}
	for ageMonth := range out {
		sort.Slice(out[ageMonth], func(i, j int) bool { return out[ageMonth][i].ItemNo < out[ageMonth][j].ItemNo })
	}
	return out
}

func erxinAgeBandByMonth(ageBands []erxinAgeBandDefinition) map[int]erxinAgeBandDefinition {
	out := make(map[int]erxinAgeBandDefinition, len(ageBands))
	for _, ageBand := range ageBands {
		out[ageBand.AgeMonth] = ageBand
	}
	return out
}
