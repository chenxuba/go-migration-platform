package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/repository"
)

const (
	shuangxiAScaleCode       = "SHUANGXI_A"
	shuangxiAScaleVersion    = "A-2012-doc"
	shuangxiAItemBankFile    = "shuangxi-a-item-bank.json"
	shuangxiADomainMapFile   = "shuangxi-a-domain-map.json"
	shuangxiAMetadataFile    = "shuangxi-a-scale-metadata.json"
	shuangxiAStaticRevision  = "2026-05-18"
	shuangxiADraftDataStatus = "draft"
)

type shuangxiAStaticData struct {
	metadata   shuangxiAScaleMetadata
	items      []shuangxiAItemDefinition
	domains    []shuangxiADomainDefinition
	sources    []string
	dataStatus string
}

type shuangxiAScaleMetadata struct {
	ScaleCode      string   `json:"scale_code"`
	ScaleName      string   `json:"scale_name"`
	ScaleVersion   string   `json:"scale_version"`
	SourceStandard string   `json:"source_standard"`
	SourceFiles    []string `json:"source_files"`
	ItemCount      int      `json:"item_count"`
	DomainCount    int      `json:"domain_count"`
	SkillCount     int      `json:"skill_count"`
	ScoreMin       int      `json:"score_min"`
	ScoreMax       int      `json:"score_max"`
	DataStatus     string   `json:"data_status"`
	Revision       string   `json:"revision"`
	ScoringNote    string   `json:"scoring_note"`
}

type shuangxiADomainDefinition struct {
	ScaleCode   string                 `json:"scale_code"`
	ScaleName   string                 `json:"scale_name"`
	SortNo      int                    `json:"sort_no"`
	Category    string                 `json:"category,omitempty"`
	DomainNo    int                    `json:"domain_no,omitempty"`
	ItemCount   int                    `json:"item_count"`
	MaxRawScore int                    `json:"max_raw_score"`
	ItemNumbers []int                  `json:"item_numbers,omitempty"`
	Skills      []shuangxiASkillDetail `json:"skills,omitempty"`
}

type shuangxiASkillDetail struct {
	SkillCode   string `json:"skill_code"`
	SkillName   string `json:"skill_name"`
	DomainCode  string `json:"domain_code"`
	DomainName  string `json:"domain_name"`
	SortNo      int    `json:"sort_no"`
	ItemCount   int    `json:"item_count"`
	ItemNumbers []int  `json:"item_numbers,omitempty"`
}

type shuangxiAItemDefinition struct {
	ItemNo         int                    `json:"item_no"`
	ItemCode       string                 `json:"item_code"`
	ItemTitle      string                 `json:"item_title"`
	TestItem       string                 `json:"test_item"`
	DomainCode     string                 `json:"domain_code"`
	Domain         string                 `json:"domain"`
	DomainName     string                 `json:"domain_name"`
	DomainSortNo   int                    `json:"domain_sort_no"`
	SkillCode      string                 `json:"skill_code"`
	SkillName      string                 `json:"skill_name"`
	SkillSortNo    int                    `json:"skill_sort_no"`
	SourcePDF      string                 `json:"source_pdf"`
	SourcePages    []int                  `json:"source_pages"`
	OCRStatus      string                 `json:"ocr_status"`
	Materials      string                 `json:"materials"`
	MaterialImages []string               `json:"material_images"`
	Method         string                 `json:"method"`
	Describes      string                 `json:"describes,omitempty"`
	Guidance       string                 `json:"guidance"`
	GuidanceVideo  string                 `json:"guidance_video"`
	ScoreOptions   []shuangxiAScoreOption `json:"score_options"`
	ScoreType      string                 `json:"score_type"`
	Standard       string                 `json:"standard"`
	ScoreMin       int                    `json:"score_min"`
	ScoreMax       int                    `json:"score_max"`
}

type shuangxiAScoreOption struct {
	Value       int    `json:"value"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

func (svc *Service) EnsureShuangxiAScaleData(ctx context.Context) error {
	if svc == nil || svc.repo == nil {
		return nil
	}
	forceReseed := os.Getenv("SHUANGXI_A_STATIC_DATA_RESEED") == "1"
	if !forceReseed {
		hasData, err := hasShuangxiAScaleData(ctx, svc.repo)
		if err != nil {
			return err
		}
		if hasData {
			dataset, err := svc.repo.GetAssessmentScaleDataset(ctx, shuangxiAScaleCode, shuangxiAScaleVersion)
			if err == nil {
				expectedSources, sourceErr := shuangxiAExpectedStaticDataSources()
				if sourceErr == nil && sameStringSlice(dataset.Sources, expectedSources) {
					return nil
				}
			}
		}
	}
	dataDir, err := resolveShuangxiADataDir()
	if err != nil {
		return err
	}
	data, err := loadShuangxiAStaticDataFromFiles(dataDir)
	if err != nil {
		return err
	}
	return svc.repo.ReplaceAssessmentScaleStaticData(ctx, shuangxiAStaticDataEntity(data), 0)
}

func hasShuangxiAScaleData(ctx context.Context, repo *repository.Repository) (bool, error) {
	if _, err := repo.GetAssessmentScaleDataset(ctx, shuangxiAScaleCode, shuangxiAScaleVersion); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	items, err := repo.ListAssessmentScaleItems(ctx, shuangxiAScaleCode, shuangxiAScaleVersion)
	if err != nil {
		return false, err
	}
	domains, err := repo.ListAssessmentScaleDomains(ctx, shuangxiAScaleCode, shuangxiAScaleVersion)
	if err != nil {
		return false, err
	}
	return len(items) == 209 && len(domains) == 7, nil
}

func loadShuangxiAStaticDataFromFiles(dataDir string) (shuangxiAStaticData, error) {
	var items []shuangxiAItemDefinition
	if err := loadShuangxiAJSONFile(filepath.Join(dataDir, shuangxiAItemBankFile), &items); err != nil {
		return shuangxiAStaticData{}, fmt.Errorf("load Shuangxi A item bank: %w", err)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ItemNo < items[j].ItemNo })

	var domains []shuangxiADomainDefinition
	if err := loadShuangxiAJSONFile(filepath.Join(dataDir, shuangxiADomainMapFile), &domains); err != nil {
		return shuangxiAStaticData{}, fmt.Errorf("load Shuangxi A domain map: %w", err)
	}
	sort.Slice(domains, func(i, j int) bool { return domains[i].SortNo < domains[j].SortNo })

	var metadata shuangxiAScaleMetadata
	if err := loadShuangxiAJSONFile(filepath.Join(dataDir, shuangxiAMetadataFile), &metadata); err != nil {
		return shuangxiAStaticData{}, fmt.Errorf("load Shuangxi A metadata: %w", err)
	}

	sources, dataStatus := shuangxiADataSources(metadata)
	return shuangxiAStaticData{
		metadata:   metadata,
		items:      items,
		domains:    domains,
		sources:    sources,
		dataStatus: dataStatus,
	}, nil
}

func loadShuangxiAJSONFile(path string, out any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("decode %s: %w", filepath.Base(path), err)
	}
	return nil
}

func resolveShuangxiADataDir() (string, error) {
	if raw := os.Getenv("SHUANGXI_A_DATA_DIR"); raw != "" {
		if err := requireShuangxiADataFiles(raw); err != nil {
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
		if requireShuangxiADataFiles(candidate) == nil {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
	return "", fmt.Errorf("Shuangxi A data files not found; set SHUANGXI_A_DATA_DIR to the directory containing %s", shuangxiAItemBankFile)
}

func requireShuangxiADataFiles(dir string) error {
	for _, name := range []string{shuangxiAItemBankFile, shuangxiADomainMapFile, shuangxiAMetadataFile} {
		if !fileExists(filepath.Join(dir, name)) {
			return fmt.Errorf("Shuangxi A data file %s not found in %s", name, dir)
		}
	}
	return nil
}

func shuangxiAExpectedStaticDataSources() ([]string, error) {
	dataDir, err := resolveShuangxiADataDir()
	if err != nil {
		return nil, err
	}
	var metadata shuangxiAScaleMetadata
	if err := loadShuangxiAJSONFile(filepath.Join(dataDir, shuangxiAMetadataFile), &metadata); err != nil {
		return nil, err
	}
	sources, _ := shuangxiADataSources(metadata)
	return sources, nil
}

func shuangxiADataSources(metadata shuangxiAScaleMetadata) ([]string, string) {
	sources := make([]string, 0, len(metadata.SourceFiles)+4)
	for _, item := range metadata.SourceFiles {
		item = strings.TrimSpace(item)
		if item != "" {
			sources = append(sources, item)
		}
	}
	sources = append(sources, "revision:"+nonEmptyString(metadata.Revision, shuangxiAStaticRevision))
	dataStatus := nonEmptyString(metadata.DataStatus, shuangxiADraftDataStatus)
	return sources, dataStatus
}

func shuangxiAStaticDataEntity(data shuangxiAStaticData) repository.AssessmentScaleStaticDataEntity {
	items := make([]repository.AssessmentScaleItemEntity, 0, len(data.items))
	for _, item := range data.items {
		raw, _ := json.Marshal(item)
		items = append(items, repository.AssessmentScaleItemEntity{
			ItemNo: item.ItemNo,
			Raw:    raw,
		})
	}
	domains := make([]repository.AssessmentScaleDomainEntity, 0, len(data.domains))
	for idx, domain := range data.domains {
		raw, _ := json.Marshal(domain)
		sortNo := domain.SortNo
		if sortNo == 0 {
			sortNo = idx + 1
		}
		domains = append(domains, repository.AssessmentScaleDomainEntity{
			DomainCode: domain.ScaleCode,
			SortNo:     sortNo,
			Raw:        raw,
		})
	}
	metadataRaw, _ := json.Marshal(data.metadata)
	return repository.AssessmentScaleStaticDataEntity{
		Dataset: repository.AssessmentScaleDatasetEntity{
			ScaleCode:    shuangxiAScaleCode,
			ScaleVersion: shuangxiAScaleVersion,
			DataStatus:   data.dataStatus,
			Sources:      append([]string(nil), data.sources...),
			Metadata:     metadataRaw,
		},
		Items:   items,
		Domains: domains,
	}
}

func (svc *Service) GetShuangxiAAssessmentFormTemplateSummary(ctx context.Context) (model.ShuangxiAssessmentFormTemplateSummaryVO, error) {
	data, err := svc.loadShuangxiAStaticData(ctx)
	if err != nil {
		return model.ShuangxiAssessmentFormTemplateSummaryVO{}, err
	}
	return buildShuangxiAAssessmentFormTemplateSummary(data), nil
}

func (svc *Service) GetShuangxiAAssessmentFormTemplateItem(ctx context.Context, itemNo int) (model.ShuangxiAssessmentItem, error) {
	if itemNo <= 0 {
		return model.ShuangxiAssessmentItem{}, fmt.Errorf("invalid Shuangxi A item number")
	}
	data, err := svc.loadShuangxiAStaticData(ctx)
	if err != nil {
		return model.ShuangxiAssessmentItem{}, err
	}
	for _, item := range data.items {
		if item.ItemNo == itemNo {
			return buildShuangxiAAssessmentItem(item), nil
		}
	}
	return model.ShuangxiAssessmentItem{}, fmt.Errorf("Shuangxi A item %d not found", itemNo)
}

func (svc *Service) loadShuangxiAStaticData(ctx context.Context) (shuangxiAStaticData, error) {
	if svc != nil && svc.repo != nil {
		data, err := loadShuangxiAStaticDataFromRepository(ctx, svc.repo)
		if err == nil {
			return data, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return shuangxiAStaticData{}, err
		}
	}
	dataDir, err := resolveShuangxiADataDir()
	if err != nil {
		return shuangxiAStaticData{}, err
	}
	return loadShuangxiAStaticDataFromFiles(dataDir)
}

func loadShuangxiAStaticDataFromRepository(ctx context.Context, repo *repository.Repository) (shuangxiAStaticData, error) {
	dataset, err := repo.GetAssessmentScaleDataset(ctx, shuangxiAScaleCode, shuangxiAScaleVersion)
	if err != nil {
		return shuangxiAStaticData{}, err
	}
	itemEntities, err := repo.ListAssessmentScaleItems(ctx, shuangxiAScaleCode, shuangxiAScaleVersion)
	if err != nil {
		return shuangxiAStaticData{}, err
	}
	domainEntities, err := repo.ListAssessmentScaleDomains(ctx, shuangxiAScaleCode, shuangxiAScaleVersion)
	if err != nil {
		return shuangxiAStaticData{}, err
	}
	if len(itemEntities) == 0 || len(domainEntities) == 0 {
		return shuangxiAStaticData{}, sql.ErrNoRows
	}
	items := make([]shuangxiAItemDefinition, 0, len(itemEntities))
	for _, entity := range itemEntities {
		var item shuangxiAItemDefinition
		if err := json.Unmarshal(entity.Raw, &item); err != nil {
			return shuangxiAStaticData{}, fmt.Errorf("decode Shuangxi A item %d: %w", entity.ItemNo, err)
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ItemNo < items[j].ItemNo })

	domains := make([]shuangxiADomainDefinition, 0, len(domainEntities))
	for _, entity := range domainEntities {
		var domain shuangxiADomainDefinition
		if err := json.Unmarshal(entity.Raw, &domain); err != nil {
			return shuangxiAStaticData{}, fmt.Errorf("decode Shuangxi A domain %s: %w", entity.DomainCode, err)
		}
		domains = append(domains, domain)
	}
	sort.Slice(domains, func(i, j int) bool { return domains[i].SortNo < domains[j].SortNo })

	var metadata shuangxiAScaleMetadata
	if len(dataset.Metadata) > 0 {
		if err := json.Unmarshal(dataset.Metadata, &metadata); err != nil {
			return shuangxiAStaticData{}, fmt.Errorf("decode Shuangxi A metadata: %w", err)
		}
	}
	if strings.TrimSpace(metadata.ScaleCode) == "" {
		metadata = defaultShuangxiAScaleMetadata(len(items), len(domains))
	}
	sources := append([]string(nil), dataset.Sources...)
	if len(sources) == 0 {
		sources, _ = shuangxiADataSources(metadata)
	}
	return shuangxiAStaticData{
		metadata:   metadata,
		items:      items,
		domains:    domains,
		sources:    sources,
		dataStatus: strings.TrimSpace(dataset.DataStatus),
	}, nil
}

func defaultShuangxiAScaleMetadata(itemCount, domainCount int) shuangxiAScaleMetadata {
	return shuangxiAScaleMetadata{
		ScaleCode:    shuangxiAScaleCode,
		ScaleName:    "双溪课程评量表A",
		ScaleVersion: shuangxiAScaleVersion,
		ItemCount:    itemCount,
		DomainCount:  domainCount,
		ScoreMin:     0,
		ScoreMax:     3,
		Revision:     shuangxiAStaticRevision,
		ScoringNote:  "每题按0、1、2、3四级评分，领域原始分为该领域题目得分合计。",
	}
}

func buildShuangxiAAssessmentFormTemplateSummary(data shuangxiAStaticData) model.ShuangxiAssessmentFormTemplateSummaryVO {
	itemByNo := make(map[int]shuangxiAItemDefinition, len(data.items))
	for _, item := range data.items {
		itemByNo[item.ItemNo] = item
	}
	domains := make([]model.ShuangxiAssessmentDomain, 0, len(data.domains))
	itemGroups := make([]model.ShuangxiAssessmentSkillSummary, 0)
	skillCount := 0
	for _, domain := range data.domains {
		skills := make([]model.ShuangxiAssessmentSkillSummary, 0, len(domain.Skills))
		for _, skill := range domain.Skills {
			summary := buildShuangxiASkillSummary(skill, itemByNo)
			skills = append(skills, summary)
			itemGroups = append(itemGroups, summary)
			skillCount++
		}
		domains = append(domains, model.ShuangxiAssessmentDomain{
			DomainCode:  domain.ScaleCode,
			DomainName:  domain.ScaleName,
			SortNo:      domain.SortNo,
			ItemCount:   domain.ItemCount,
			MaxRawScore: domain.MaxRawScore,
			Skills:      skills,
		})
	}
	itemCount := data.metadata.ItemCount
	if itemCount <= 0 {
		itemCount = len(data.items)
	}
	domainCount := data.metadata.DomainCount
	if domainCount <= 0 {
		domainCount = len(data.domains)
	}
	if skillCount <= 0 {
		skillCount = data.metadata.SkillCount
	}
	dataStatus := nonEmptyString(data.dataStatus, data.metadata.DataStatus)
	return model.ShuangxiAssessmentFormTemplateSummaryVO{
		TemplateCode:    "SHUANGXI_A_ASSESSMENT_FORM",
		TemplateVersion: shuangxiAScaleVersion,
		Title:           nonEmptyString(data.metadata.ScaleName, "双溪课程评量表A"),
		ScaleCode:       shuangxiAScaleCode,
		ScaleVersion:    shuangxiAScaleVersion,
		DataStatus:      dataStatus,
		Sources:         append([]string(nil), data.sources...),
		ItemCount:       itemCount,
		DomainCount:     domainCount,
		SkillCount:      skillCount,
		ScoreMin:        data.metadata.ScoreMin,
		ScoreMax:        data.metadata.ScoreMax,
		ScoringNote:     data.metadata.ScoringNote,
		Domains:         domains,
		ScoreOptions:    shuangxiAScoreOptions(),
		ItemGroups:      itemGroups,
	}
}

func buildShuangxiASkillSummary(skill shuangxiASkillDetail, itemByNo map[int]shuangxiAItemDefinition) model.ShuangxiAssessmentSkillSummary {
	items := make([]model.ShuangxiAssessmentItemSummary, 0, len(skill.ItemNumbers))
	for _, itemNo := range skill.ItemNumbers {
		item, ok := itemByNo[itemNo]
		if !ok {
			continue
		}
		items = append(items, model.ShuangxiAssessmentItemSummary{
			ItemNo:     item.ItemNo,
			ItemCode:   item.ItemCode,
			ItemTitle:  item.ItemTitle,
			TestItem:   item.TestItem,
			DomainCode: item.DomainCode,
			DomainName: item.DomainName,
			SkillCode:  item.SkillCode,
			SkillName:  item.SkillName,
		})
	}
	return model.ShuangxiAssessmentSkillSummary{
		SkillCode:  skill.SkillCode,
		SkillName:  skill.SkillName,
		DomainCode: skill.DomainCode,
		DomainName: skill.DomainName,
		SortNo:     skill.SortNo,
		ItemCount:  skill.ItemCount,
		Items:      items,
	}
}

func buildShuangxiAAssessmentItem(item shuangxiAItemDefinition) model.ShuangxiAssessmentItem {
	scoreOptions := make([]model.PEP3ScoreOption, 0, len(item.ScoreOptions))
	for _, option := range item.ScoreOptions {
		scoreOptions = append(scoreOptions, model.PEP3ScoreOption{
			Value:       option.Value,
			Label:       option.Label,
			Description: option.Description,
		})
	}
	if len(scoreOptions) == 0 {
		scoreOptions = shuangxiAScoreOptions()
	}
	return model.ShuangxiAssessmentItem{
		ItemNo:       item.ItemNo,
		ItemCode:     item.ItemCode,
		ItemTitle:    item.ItemTitle,
		TestItem:     item.TestItem,
		DomainCode:   item.DomainCode,
		DomainName:   item.DomainName,
		DomainSortNo: item.DomainSortNo,
		SkillCode:    item.SkillCode,
		SkillName:    item.SkillName,
		SkillSortNo:  item.SkillSortNo,
		Method:       item.Method,
		Standard:     item.Standard,
		ScoreMin:     item.ScoreMin,
		ScoreMax:     item.ScoreMax,
		ScoreOptions: scoreOptions,
		SourcePDF:    item.SourcePDF,
		SourcePages:  append([]int(nil), item.SourcePages...),
		OCRStatus:    item.OCRStatus,
	}
}

func shuangxiAScoreOptions() []model.PEP3ScoreOption {
	return []model.PEP3ScoreOption{
		{Value: 0, Label: "0分", Description: "尚未出现或无法完成"},
		{Value: 1, Label: "1分", Description: "大量协助下可完成"},
		{Value: 2, Label: "2分", Description: "少量协助或提示下可完成"},
		{Value: 3, Label: "3分", Description: "能独立稳定完成"},
	}
}
