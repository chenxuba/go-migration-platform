package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"go-migration-platform/pkg/vbmappscore"
	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) ListVBMAPPMaterialProfiles(userID int64) ([]model.VBMAPPMaterialProfile, error) {
	if svc == nil || svc.repo == nil {
		return nil, errors.New("repository is not configured")
	}
	if userID <= 0 {
		return nil, errors.New("用户未登录")
	}
	return svc.repo.ListVBMAPPMaterialProfiles(context.Background(), vbmappScaleVersion)
}

func (svc *Service) SaveVBMAPPMaterialItem(userID int64, item model.VBMAPPMaterialItem) (model.VBMAPPMaterialItem, error) {
	if svc == nil || svc.repo == nil {
		return model.VBMAPPMaterialItem{}, errors.New("repository is not configured")
	}
	if userID <= 0 {
		return model.VBMAPPMaterialItem{}, errors.New("用户未登录")
	}
	item.ProfileID = strings.TrimSpace(item.ProfileID)
	item.MaterialName = strings.TrimSpace(item.MaterialName)
	item.MaterialCode = strings.TrimSpace(item.MaterialCode)
	item.MaterialType = strings.TrimSpace(item.MaterialType)
	if item.ProfileID == "" {
		return model.VBMAPPMaterialItem{}, errors.New("素材分类不能为空")
	}
	if item.MaterialName == "" {
		return model.VBMAPPMaterialItem{}, errors.New("素材名称不能为空")
	}
	profiles, err := svc.repo.ListVBMAPPMaterialProfiles(context.Background(), vbmappScaleVersion)
	if err != nil {
		return model.VBMAPPMaterialItem{}, err
	}
	if !vbmappMaterialProfileExists(profiles, item.ProfileID) {
		return model.VBMAPPMaterialItem{}, errors.New("素材分类不存在")
	}
	return svc.repo.SaveVBMAPPMaterialItem(context.Background(), vbmappScaleVersion, userID, item)
}

func (svc *Service) DeleteVBMAPPMaterialItem(userID, id int64) error {
	if svc == nil || svc.repo == nil {
		return errors.New("repository is not configured")
	}
	if userID <= 0 {
		return errors.New("用户未登录")
	}
	if id <= 0 {
		return errors.New("素材ID不能为空")
	}
	if err := svc.repo.DeleteVBMAPPMaterialItem(context.Background(), vbmappScaleVersion, id, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("素材不存在或已删除")
		}
		return err
	}
	return nil
}

func (svc *Service) GetVBMAPPMaterialCatalog(
	userID int64,
	moduleCode string,
	itemCode string,
) (model.VBMAPPMaterialCatalog, error) {
	if svc == nil {
		return model.VBMAPPMaterialCatalog{}, errors.New("service is not configured")
	}
	if userID <= 0 {
		return model.VBMAPPMaterialCatalog{}, errors.New("用户未登录")
	}
	data, err := loadVBMAPPStaticData()
	if err != nil {
		return model.VBMAPPMaterialCatalog{}, err
	}
	dataDir, err := resolveVBMAPPDataDir()
	if err != nil {
		return model.VBMAPPMaterialCatalog{}, err
	}
	milestoneSchemas, err := vbmappscore.LoadMilestoneResponseSchemasFile(filepath.Join(dataDir, vbmappMilestoneSchemaFile))
	if err != nil {
		return model.VBMAPPMaterialCatalog{}, fmt.Errorf("load VB-MAPP milestone response schemas: %w", err)
	}
	materialProfiles, err := svc.loadMergedVBMAPPMaterialProfiles(context.Background(), dataDir)
	if err != nil {
		return model.VBMAPPMaterialCatalog{}, err
	}

	moduleCode = strings.ToLower(strings.TrimSpace(moduleCode))
	itemCode = strings.ToUpper(strings.TrimSpace(itemCode))
	itemByCode := make(map[string]vbmappscore.MilestoneItemDefinition, len(data.milestones))
	for _, item := range data.milestones {
		itemByCode[strings.ToUpper(strings.TrimSpace(item.MilestoneID))] = item
	}
	items := make([]model.VBMAPPMaterialCatalogItem, 0, len(milestoneSchemas))
	for _, schema := range milestoneSchemas {
		if strings.TrimSpace(schema.MaterialProfileID) == "" {
			continue
		}
		if moduleCode != "" && strings.ToLower(strings.TrimSpace(schema.ModuleCode)) != moduleCode {
			continue
		}
		code := strings.ToUpper(strings.TrimSpace(schema.MilestoneID))
		if itemCode != "" && code != itemCode {
			continue
		}
		definition := itemByCode[code]
		profile := materialProfiles[strings.TrimSpace(schema.MaterialProfileID)]
		items = append(items, model.VBMAPPMaterialCatalogItem{
			ModuleCode:           schema.ModuleCode,
			ItemCode:             code,
			Label:                firstNonEmptyVBMAPPString(definition.Label, schema.Label),
			Title:                strings.TrimSpace(definition.Title),
			DomainCode:           firstNonEmptyVBMAPPString(definition.DomainCode, schema.DomainCode),
			DomainName:           firstNonEmptyVBMAPPString(definition.DomainName, schema.DomainName),
			Level:                definition.Level,
			AgeBand:              definition.AgeBand,
			AssessmentMode:       definition.AssessmentMode,
			MaterialProfileID:    strings.TrimSpace(schema.MaterialProfileID),
			MaterialProfileLabel: strings.TrimSpace(profile.Label),
			WhyRecord:            strings.TrimSpace(schema.WhyRecord),
			QualityChecks:        append([]string(nil), schema.QualityChecks...),
			PreparationChecks:    append([]string(nil), profile.PreparationChecks...),
			SuggestedTypes:       append([]string(nil), profile.SuggestedTypes...),
			RecommendedMaterials: vbmappMaterialSuggestionsFromProfile(profile),
			QuickPicksByField:    cloneVBMAPPMaterialQuickPicks(profile.QuickPicksByField),
		})
	}
	return model.VBMAPPMaterialCatalog{
		ScaleCode:      vbmappScaleCode,
		ScaleVersion:   vbmappScaleVersion,
		AssessmentName: vbmappAssessmentName,
		DataStatus:     data.dataStatus,
		Sources:        data.sources,
		Items:          items,
	}, nil
}

func vbmappMaterialProfileExists(profiles []model.VBMAPPMaterialProfile, profileID string) bool {
	for _, profile := range profiles {
		if profile.ProfileID == profileID {
			return true
		}
	}
	return false
}

func vbmappMaterialSuggestionsFromProfile(profile vbmappscore.ResponseMaterialProfile) []model.VBMAPPMaterialSuggestion {
	if len(profile.RecommendedMaterials) == 0 {
		return nil
	}
	out := make([]model.VBMAPPMaterialSuggestion, 0, len(profile.RecommendedMaterials))
	for _, material := range profile.RecommendedMaterials {
		if strings.TrimSpace(material.Name) == "" {
			continue
		}
		out = append(out, model.VBMAPPMaterialSuggestion{
			MaterialCode: strings.TrimSpace(material.ID),
			MaterialName: strings.TrimSpace(material.Name),
			MaterialType: strings.TrimSpace(material.Type),
		})
	}
	return out
}

func cloneVBMAPPMaterialQuickPicks(input map[string][]string) map[string][]string {
	if len(input) == 0 {
		return nil
	}
	out := make(map[string][]string, len(input))
	for key, values := range input {
		normalizedKey := strings.TrimSpace(key)
		if normalizedKey == "" || len(values) == 0 {
			continue
		}
		out[normalizedKey] = append([]string(nil), values...)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func firstNonEmptyVBMAPPString(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
