package service

import (
	"fmt"
	"os"
	"path/filepath"
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
	vbmappStaticRevision        = "draft-2026-05-19"
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
