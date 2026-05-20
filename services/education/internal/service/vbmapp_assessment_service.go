package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"go-migration-platform/pkg/vbmappscore"
)

const (
	vbmappScaleCode             = vbmappscore.ScaleCode
	vbmappScaleVersion          = vbmappscore.DefaultScaleVersion
	vbmappAssessmentName        = "VB-MAPP语言行为里程碑评估及安置计划"
	vbmappDomainFile            = "domains.json"
	vbmappMilestoneItemFile     = "milestone-items.json"
	vbmappMilestoneRuleFile     = "milestone-scoring-rules.json"
	vbmappBarrierFile           = "barriers.json"
	vbmappTransitionFile        = "transition.json"
	vbmappMilestoneSchemaFile   = "milestone-response-schemas.json"
	vbmappBarrierSchemaFile     = "barrier-response-schemas.json"
	vbmappTransitionSchemaFile  = "transition-response-schemas.json"
	vbmappFieldTemplateFile     = "response-field-templates.json"
	vbmappMaterialProfileFile   = "response-material-profiles.json"
	vbmappSchemaSummaryFile     = "response-schema-summary.json"
	vbmappStaticRevision        = "draft-2026-05-20"
	vbmappDraftDataStatus       = "draft"
	vbmappDefaultDataSubdirName = "vbmapp"
)

type VBMAPPScoreDataInfo struct {
	ScaleCode      string   `json:"scaleCode"`
	ScaleVersion   string   `json:"scaleVersion"`
	AssessmentName string   `json:"assessmentName"`
	DataStatus     string   `json:"dataStatus,omitempty"`
	Sources        []string `json:"sources"`
}

type VBMAPPScoreResponse struct {
	VBMAPPScoreDataInfo
	Result vbmappscore.AssessmentResult `json:"result"`
}

type VBMAPPAssessmentSchemaResponse struct {
	VBMAPPScoreDataInfo
	Domains                   []vbmappscore.DomainDefinition                 `json:"domains"`
	MilestoneItems            []vbmappscore.MilestoneItemDefinition          `json:"milestoneItems"`
	Barriers                  []vbmappscore.BarrierDefinition                `json:"barriers"`
	Transitions               []vbmappscore.TransitionDefinition             `json:"transitions"`
	MilestoneResponseSchemas  []vbmappscore.MilestoneResponseSchema          `json:"milestoneResponseSchemas"`
	BarrierResponseSchemas    []vbmappscore.BarrierResponseSchema            `json:"barrierResponseSchemas"`
	TransitionResponseSchemas []vbmappscore.TransitionResponseSchema         `json:"transitionResponseSchemas"`
	ResponseFieldTemplates    map[string]vbmappscore.ResponseFieldTemplate   `json:"responseFieldTemplates"`
	ResponseMaterialProfiles  map[string]vbmappscore.ResponseMaterialProfile `json:"responseMaterialProfiles"`
	ResponseSchemaSummary     map[string]any                                 `json:"responseSchemaSummary,omitempty"`
}

type vbmappStaticData struct {
	domains     []vbmappscore.DomainDefinition
	milestones  []vbmappscore.MilestoneItemDefinition
	rules       []vbmappscore.MilestoneScoringRule
	barriers    []vbmappscore.BarrierDefinition
	transitions []vbmappscore.TransitionDefinition
	sources     []string
	dataStatus  string
}

var (
	vbmappEngineOnce    sync.Once
	vbmappEngine        *vbmappscore.Engine
	vbmappEngineInfo    VBMAPPScoreDataInfo
	vbmappEngineLoadErr error
)

func (svc *Service) ScoreVBMAPP(input vbmappscore.AssessmentInput) (VBMAPPScoreResponse, error) {
	engine, info, err := loadVBMAPPEngine()
	if err != nil {
		return VBMAPPScoreResponse{}, err
	}
	result, err := engine.Score(input)
	if err != nil {
		return VBMAPPScoreResponse{}, err
	}
	return VBMAPPScoreResponse{
		VBMAPPScoreDataInfo: info,
		Result:              result,
	}, nil
}

func (svc *Service) VBMAPPAssessmentSchema() (VBMAPPAssessmentSchemaResponse, error) {
	data, err := loadVBMAPPStaticData()
	if err != nil {
		return VBMAPPAssessmentSchemaResponse{}, err
	}
	dataDir, err := resolveVBMAPPDataDir()
	if err != nil {
		return VBMAPPAssessmentSchemaResponse{}, err
	}
	if err := requireVBMAPPResponseSchemaFiles(dataDir); err != nil {
		return VBMAPPAssessmentSchemaResponse{}, err
	}
	milestoneSchemas, err := vbmappscore.LoadMilestoneResponseSchemasFile(filepath.Join(dataDir, vbmappMilestoneSchemaFile))
	if err != nil {
		return VBMAPPAssessmentSchemaResponse{}, fmt.Errorf("load VB-MAPP milestone response schemas: %w", err)
	}
	barrierSchemas, err := vbmappscore.LoadBarrierResponseSchemasFile(filepath.Join(dataDir, vbmappBarrierSchemaFile))
	if err != nil {
		return VBMAPPAssessmentSchemaResponse{}, fmt.Errorf("load VB-MAPP barrier response schemas: %w", err)
	}
	transitionSchemas, err := vbmappscore.LoadTransitionResponseSchemasFile(filepath.Join(dataDir, vbmappTransitionSchemaFile))
	if err != nil {
		return VBMAPPAssessmentSchemaResponse{}, fmt.Errorf("load VB-MAPP transition response schemas: %w", err)
	}
	fieldTemplates, err := vbmappscore.LoadResponseFieldTemplatesFile(filepath.Join(dataDir, vbmappFieldTemplateFile))
	if err != nil {
		return VBMAPPAssessmentSchemaResponse{}, fmt.Errorf("load VB-MAPP response field templates: %w", err)
	}
	materialProfiles, err := svc.loadMergedVBMAPPMaterialProfiles(context.Background(), dataDir)
	if err != nil {
		return VBMAPPAssessmentSchemaResponse{}, err
	}
	summary, err := loadVBMAPPResponseSchemaSummary(filepath.Join(dataDir, vbmappSchemaSummaryFile))
	if err != nil {
		return VBMAPPAssessmentSchemaResponse{}, err
	}
	sources, dataStatus := vbmappSchemaDataSources()
	return VBMAPPAssessmentSchemaResponse{
		VBMAPPScoreDataInfo: VBMAPPScoreDataInfo{
			ScaleCode:      vbmappScaleCode,
			ScaleVersion:   vbmappScaleVersion,
			AssessmentName: vbmappAssessmentName,
			DataStatus:     dataStatus,
			Sources:        sources,
		},
		Domains:                   data.domains,
		MilestoneItems:            data.milestones,
		Barriers:                  data.barriers,
		Transitions:               data.transitions,
		MilestoneResponseSchemas:  milestoneSchemas,
		BarrierResponseSchemas:    barrierSchemas,
		TransitionResponseSchemas: transitionSchemas,
		ResponseFieldTemplates:    fieldTemplates,
		ResponseMaterialProfiles:  materialProfiles,
		ResponseSchemaSummary:     summary,
	}, nil
}

func (svc *Service) loadMergedVBMAPPMaterialProfiles(
	ctx context.Context,
	dataDir string,
) (map[string]vbmappscore.ResponseMaterialProfile, error) {
	materialProfiles, err := vbmappscore.LoadResponseMaterialProfilesFile(filepath.Join(dataDir, vbmappMaterialProfileFile))
	if err != nil {
		return nil, fmt.Errorf("load VB-MAPP response material profiles: %w", err)
	}
	if svc != nil && svc.repo != nil {
		dbProfiles, dbErr := svc.repo.ListVBMAPPResponseMaterialProfiles(ctx, vbmappScaleVersion)
		if dbErr == nil && len(dbProfiles) > 0 {
			for profileID, profile := range dbProfiles {
				materialProfiles[profileID] = mergeVBMAPPResponseMaterialProfile(materialProfiles[profileID], profile)
			}
		} else if dbErr != nil && !isVBMAPPMaterialLibraryFallbackError(dbErr) {
			return nil, fmt.Errorf("load VB-MAPP DB material profiles: %w", dbErr)
		}
	}
	return materialProfiles, nil
}

func mergeVBMAPPResponseMaterialProfile(fileProfile, dbProfile vbmappscore.ResponseMaterialProfile) vbmappscore.ResponseMaterialProfile {
	merged := fileProfile
	if strings.TrimSpace(dbProfile.Label) != "" {
		merged.Label = dbProfile.Label
	}
	if strings.TrimSpace(dbProfile.SourceLogic) != "" {
		merged.SourceLogic = dbProfile.SourceLogic
	}
	if len(dbProfile.SuggestedTypes) > 0 {
		merged.SuggestedTypes = dbProfile.SuggestedTypes
	}
	if len(dbProfile.RecommendedMaterials) > 0 {
		merged.RecommendedMaterials = dbProfile.RecommendedMaterials
	}
	if len(dbProfile.QuickPicksByField) > 0 {
		merged.QuickPicksByField = dbProfile.QuickPicksByField
	}
	if len(dbProfile.PreparationChecks) > 0 {
		merged.PreparationChecks = dbProfile.PreparationChecks
	}
	return merged
}

func (svc *Service) EnsureVBMAPPScaleData(ctx context.Context) error {
	if svc == nil || svc.repo == nil {
		return nil
	}
	forceReseed := os.Getenv("VBMAPP_MATERIAL_LIBRARY_RESEED") == "1"
	if !forceReseed {
		hasProfiles, err := svc.repo.HasVBMAPPResponseMaterialProfiles(ctx, vbmappScaleVersion)
		if err != nil {
			return err
		}
		if hasProfiles {
			return nil
		}
	}
	profiles, err := loadVBMAPPResponseMaterialProfilesFromFiles()
	if err != nil {
		return err
	}
	return svc.repo.ReplaceVBMAPPResponseMaterialProfiles(ctx, vbmappScaleVersion, profiles, 0)
}

func loadVBMAPPResponseMaterialProfilesFromFiles() (map[string]vbmappscore.ResponseMaterialProfile, error) {
	dataDir, err := resolveVBMAPPDataDir()
	if err != nil {
		return nil, err
	}
	profiles, err := vbmappscore.LoadResponseMaterialProfilesFile(filepath.Join(dataDir, vbmappMaterialProfileFile))
	if err != nil {
		return nil, fmt.Errorf("load VB-MAPP response material profiles: %w", err)
	}
	return profiles, nil
}

func isVBMAPPMaterialLibraryFallbackError(err error) bool {
	if err == nil {
		return false
	}
	return err == sql.ErrNoRows ||
		strings.Contains(err.Error(), "vbmapp_material_profile") ||
		strings.Contains(err.Error(), "vbmapp_material_item") ||
		strings.Contains(err.Error(), "database is closed")
}

func loadVBMAPPEngine() (*vbmappscore.Engine, VBMAPPScoreDataInfo, error) {
	vbmappEngineOnce.Do(func() {
		vbmappEngine, vbmappEngineInfo, vbmappEngineLoadErr = buildVBMAPPEngine()
	})
	if vbmappEngineLoadErr != nil {
		return nil, VBMAPPScoreDataInfo{}, vbmappEngineLoadErr
	}
	return vbmappEngine, vbmappEngineInfo, nil
}

func buildVBMAPPEngine() (*vbmappscore.Engine, VBMAPPScoreDataInfo, error) {
	data, err := loadVBMAPPStaticData()
	if err != nil {
		return nil, VBMAPPScoreDataInfo{}, err
	}
	engine, err := vbmappscore.NewEngine(data.domains, data.milestones, data.rules, data.barriers, data.transitions)
	if err != nil {
		return nil, VBMAPPScoreDataInfo{}, fmt.Errorf("build VB-MAPP score engine: %w", err)
	}
	return engine, VBMAPPScoreDataInfo{
		ScaleCode:      vbmappScaleCode,
		ScaleVersion:   vbmappScaleVersion,
		AssessmentName: vbmappAssessmentName,
		DataStatus:     data.dataStatus,
		Sources:        data.sources,
	}, nil
}

func loadVBMAPPStaticData() (vbmappStaticData, error) {
	dataDir, err := resolveVBMAPPDataDir()
	if err != nil {
		return vbmappStaticData{}, err
	}
	domains, err := vbmappscore.LoadDomainDefinitionsFile(filepath.Join(dataDir, vbmappDomainFile))
	if err != nil {
		return vbmappStaticData{}, fmt.Errorf("load VB-MAPP domains: %w", err)
	}
	milestones, err := vbmappscore.LoadMilestoneItemDefinitionsFile(filepath.Join(dataDir, vbmappMilestoneItemFile))
	if err != nil {
		return vbmappStaticData{}, fmt.Errorf("load VB-MAPP milestone items: %w", err)
	}
	rules, err := vbmappscore.LoadMilestoneScoringRulesFile(filepath.Join(dataDir, vbmappMilestoneRuleFile))
	if err != nil {
		return vbmappStaticData{}, fmt.Errorf("load VB-MAPP milestone scoring rules: %w", err)
	}
	barriers, err := vbmappscore.LoadBarrierDefinitionsFile(filepath.Join(dataDir, vbmappBarrierFile))
	if err != nil {
		return vbmappStaticData{}, fmt.Errorf("load VB-MAPP barriers: %w", err)
	}
	transitions, err := vbmappscore.LoadTransitionDefinitionsFile(filepath.Join(dataDir, vbmappTransitionFile))
	if err != nil {
		return vbmappStaticData{}, fmt.Errorf("load VB-MAPP transition items: %w", err)
	}
	sources, dataStatus := vbmappDataSources()
	return vbmappStaticData{
		domains:     domains,
		milestones:  milestones,
		rules:       rules,
		barriers:    barriers,
		transitions: transitions,
		sources:     sources,
		dataStatus:  dataStatus,
	}, nil
}

func resolveVBMAPPDataDir() (string, error) {
	if raw := os.Getenv("VBMAPP_DATA_DIR"); raw != "" {
		if err := requireVBMAPPDataFiles(raw); err != nil {
			return "", err
		}
		return raw, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for dir := cwd; ; dir = filepath.Dir(dir) {
		candidate := filepath.Join(dir, "docs", vbmappDefaultDataSubdirName)
		if requireVBMAPPDataFiles(candidate) == nil {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
	return "", fmt.Errorf("VB-MAPP data files not found; set VBMAPP_DATA_DIR to the directory containing %s", vbmappMilestoneItemFile)
}

func loadVBMAPPResponseSchemaSummary(path string) (map[string]any, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("load VB-MAPP response schema summary: %w", err)
	}
	defer file.Close()
	var summary map[string]any
	if err := json.NewDecoder(file).Decode(&summary); err != nil {
		return nil, fmt.Errorf("decode VB-MAPP response schema summary: %w", err)
	}
	return summary, nil
}

func requireVBMAPPDataFiles(dir string) error {
	for _, name := range []string{
		vbmappDomainFile,
		vbmappMilestoneItemFile,
		vbmappMilestoneRuleFile,
		vbmappBarrierFile,
		vbmappTransitionFile,
	} {
		if !fileExists(filepath.Join(dir, name)) {
			return fmt.Errorf("VB-MAPP data file %s not found in %s", name, dir)
		}
	}
	return nil
}

func requireVBMAPPResponseSchemaFiles(dir string) error {
	for _, name := range []string{
		vbmappMilestoneSchemaFile,
		vbmappBarrierSchemaFile,
		vbmappTransitionSchemaFile,
		vbmappFieldTemplateFile,
		vbmappMaterialProfileFile,
		vbmappSchemaSummaryFile,
	} {
		if !fileExists(filepath.Join(dir, name)) {
			return fmt.Errorf("VB-MAPP response schema file %s not found in %s", name, dir)
		}
	}
	return nil
}

func vbmappDataSources() ([]string, string) {
	return []string{
		vbmappDefaultDataSubdirName + "/" + vbmappDomainFile,
		vbmappDefaultDataSubdirName + "/" + vbmappMilestoneItemFile,
		vbmappDefaultDataSubdirName + "/" + vbmappMilestoneRuleFile,
		vbmappDefaultDataSubdirName + "/" + vbmappBarrierFile,
		vbmappDefaultDataSubdirName + "/" + vbmappTransitionFile,
		"revision:" + vbmappStaticRevision,
	}, vbmappDraftDataStatus
}

func vbmappSchemaDataSources() ([]string, string) {
	sources, dataStatus := vbmappDataSources()
	return append(sources,
		vbmappDefaultDataSubdirName+"/"+vbmappMilestoneSchemaFile,
		vbmappDefaultDataSubdirName+"/"+vbmappBarrierSchemaFile,
		vbmappDefaultDataSubdirName+"/"+vbmappTransitionSchemaFile,
		vbmappDefaultDataSubdirName+"/"+vbmappFieldTemplateFile,
		vbmappDefaultDataSubdirName+"/"+vbmappMaterialProfileFile,
		vbmappDefaultDataSubdirName+"/"+vbmappSchemaSummaryFile,
	), dataStatus
}
