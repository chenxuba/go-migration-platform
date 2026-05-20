package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go-migration-platform/pkg/vbmappscore"
	"go-migration-platform/services/education/internal/model"
)

func TestScoreVBMAPPWithGeneratedDraftsWhenPresent(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..")
	dataDir := filepath.Join(root, "docs", "vbmapp")
	if _, err := os.Stat(filepath.Join(dataDir, "milestone-items.json")); err != nil {
		t.Skipf("generated VB-MAPP data not present: %s", dataDir)
	}

	result, err := (&Service{}).ScoreVBMAPP(vbmappscore.AssessmentInput{
		MilestoneScores: map[string]float64{
			"MAND_01M": 1,
			"MAND_02M": 0.5,
		},
		BarrierScores: map[string]int{
			"B01": 3,
			"B02": 2,
		},
		TransitionScores: map[string]int{
			"T06": 3,
		},
	})
	if err != nil {
		t.Fatalf("ScoreVBMAPP returned error: %v", err)
	}
	if result.ScaleCode != vbmappScaleCode || result.ScaleVersion != vbmappScaleVersion || result.DataStatus != vbmappDraftDataStatus {
		t.Fatalf("unexpected score data info: %+v", result.VBMAPPScoreDataInfo)
	}
	if result.Result.Milestones.TotalScore != 1.5 || result.Result.Milestones.AnsweredItems != 2 {
		t.Fatalf("unexpected milestone result: %+v", result.Result.Milestones)
	}
	if result.Result.Barriers.TotalScore != 5 || len(result.Result.Barriers.HighRiskItems) != 1 {
		t.Fatalf("unexpected barrier result: %+v", result.Result.Barriers)
	}
	if len(result.Result.Transition.Suggestions) == 0 {
		t.Fatalf("expected transition suggestions: %+v", result.Result.Transition)
	}
}

func TestVBMAPPAssessmentSchemaWithGeneratedDraftsWhenPresent(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..")
	dataDir := filepath.Join(root, "docs", "vbmapp")
	if _, err := os.Stat(filepath.Join(dataDir, "milestone-response-schemas.json")); err != nil {
		t.Skipf("generated VB-MAPP response schema data not present: %s", dataDir)
	}

	schema, err := (&Service{}).VBMAPPAssessmentSchema()
	if err != nil {
		t.Fatalf("VBMAPPAssessmentSchema returned error: %v", err)
	}
	if schema.ScaleCode != vbmappScaleCode || schema.ScaleVersion != vbmappScaleVersion {
		t.Fatalf("unexpected schema data info: %+v", schema.VBMAPPScoreDataInfo)
	}
	if len(schema.MilestoneItems) != 170 || len(schema.MilestoneResponseSchemas) != 170 {
		t.Fatalf("unexpected milestone schema counts: items=%d schemas=%d", len(schema.MilestoneItems), len(schema.MilestoneResponseSchemas))
	}
	if len(schema.Barriers) != 24 || len(schema.BarrierResponseSchemas) != 24 {
		t.Fatalf("unexpected barrier schema counts: items=%d schemas=%d", len(schema.Barriers), len(schema.BarrierResponseSchemas))
	}
	if len(schema.Transitions) != 18 || len(schema.TransitionResponseSchemas) != 18 {
		t.Fatalf("unexpected transition schema counts: items=%d schemas=%d", len(schema.Transitions), len(schema.TransitionResponseSchemas))
	}
	if schema.MilestoneResponseSchemas[0].MilestoneID != "MAND_01M" || schema.MilestoneResponseSchemas[0].UIPattern == "" {
		t.Fatalf("unexpected first milestone schema: %+v", schema.MilestoneResponseSchemas[0])
	}
	if schema.MilestoneResponseSchemas[0].MaterialProfileID != "mand_1m_request_starter_set" {
		t.Fatalf("unexpected MAND_01M material profile: %s", schema.MilestoneResponseSchemas[0].MaterialProfileID)
	}
	if len(schema.ResponseFieldTemplates) == 0 || len(schema.ResponseMaterialProfiles) == 0 || len(schema.ResponseSchemaSummary) == 0 {
		t.Fatalf("expected schema support dictionaries: %+v", schema)
	}
	if _, ok := schema.ResponseMaterialProfiles["mand_2m_visible_request_set"]; !ok {
		t.Fatalf("expected MAND_02M material profile in schema response")
	}
	if !containsString(schema.ResponseMaterialProfiles["potential_reinforcer_set"].QuickPicksByField["mand3_people"], "爸爸") {
		t.Fatalf("expected MAND_03M people quick picks in schema response: %+v", schema.ResponseMaterialProfiles["potential_reinforcer_set"].QuickPicksByField)
	}
}

func TestBuildVBMAPPAssessmentDraftProgressWithPartialScoresWhenPresent(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..")
	dataDir := filepath.Join(root, "docs", "vbmapp")
	if _, err := os.Stat(filepath.Join(dataDir, "milestone-items.json")); err != nil {
		t.Skipf("generated VB-MAPP data not present: %s", dataDir)
	}

	birthDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	progress, err := buildVBMAPPAssessmentDraftProgress(&birthDate, nil, vbmappscore.AssessmentInput{
		MilestoneScores: map[string]float64{
			"MAND_01M": 1,
			"MAND_02M": 0.5,
		},
		BarrierScores: map[string]int{
			"B01": 3,
		},
	})
	if err != nil {
		t.Fatalf("buildVBMAPPAssessmentDraftProgress returned error: %v", err)
	}
	if progress.ItemCount != 212 || progress.AnsweredItemCount != 3 || progress.MissingItemCount != 209 {
		t.Fatalf("unexpected item progress: %+v", progress)
	}
	if !progress.CanScore || progress.Complete {
		t.Fatalf("partial VB-MAPP draft should be score-previewable but incomplete: %+v", progress)
	}
	if !containsString(progress.MissingRequiredFields, "assessmentDate") || !containsString(progress.MissingRequiredFields, "transitionScoreList") {
		t.Fatalf("unexpected missing required fields: %+v", progress.MissingRequiredFields)
	}
	if len(progress.DomainProgress) < 3 {
		t.Fatalf("expected module/domain progress: %+v", progress.DomainProgress)
	}
}

func TestBuildVBMAPPAssessmentDraftProgressCompleteRequiresDatesWhenPresent(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..")
	dataDir := filepath.Join(root, "docs", "vbmapp")
	if _, err := os.Stat(filepath.Join(dataDir, "milestone-items.json")); err != nil {
		t.Skipf("generated VB-MAPP data not present: %s", dataDir)
	}

	data, err := loadVBMAPPStaticData()
	if err != nil {
		t.Fatalf("loadVBMAPPStaticData returned error: %v", err)
	}
	input := vbmappscore.AssessmentInput{
		MilestoneScores:  map[string]float64{},
		BarrierScores:    map[string]int{},
		TransitionScores: map[string]int{},
	}
	for _, item := range data.milestones {
		input.MilestoneScores[item.MilestoneID] = 1
	}
	for _, item := range data.barriers {
		input.BarrierScores[item.BarrierCode] = item.MinScore
	}
	for _, item := range data.transitions {
		input.TransitionScores[item.TransitionCode] = item.MinScore
	}

	birthDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	assessmentDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	missingDatesProgress, err := buildVBMAPPAssessmentDraftProgress(nil, nil, input)
	if err != nil {
		t.Fatalf("buildVBMAPPAssessmentDraftProgress without dates returned error: %v", err)
	}
	if missingDatesProgress.Complete {
		t.Fatalf("progress without dates should not be complete: %+v", missingDatesProgress)
	}

	progress, err := buildVBMAPPAssessmentDraftProgress(&birthDate, &assessmentDate, input)
	if err != nil {
		t.Fatalf("buildVBMAPPAssessmentDraftProgress returned error: %v", err)
	}
	if !progress.Complete || progress.MissingItemCount != 0 || len(progress.MissingRequiredFields) != 0 {
		t.Fatalf("expected complete progress: %+v", progress)
	}
}

func TestApplyVBMAPPItemScoreValidatesModuleScoreShape(t *testing.T) {
	input := vbmappscore.AssessmentInput{}
	score := 0.5
	if err := applyVBMAPPItemScore(&input, "里程碑评估", "mand_01m", score); err != nil {
		t.Fatalf("apply milestone score: %v", err)
	}
	if input.MilestoneScores["MAND_01M"] != 0.5 {
		t.Fatalf("unexpected milestone scores: %+v", input.MilestoneScores)
	}
	if err := applyVBMAPPItemScore(&input, "障碍评估", "b01", 2.5); err == nil {
		t.Fatalf("expected non-integer barrier score to fail")
	}
	if err := applyVBMAPPItemScore(&input, "转衔评估", "t01", 4); err != nil {
		t.Fatalf("apply transition score: %v", err)
	}
	if input.TransitionScores["T01"] != 4 {
		t.Fatalf("unexpected transition scores: %+v", input.TransitionScores)
	}
}

func TestMergeVBMAPPDraftInputSnapshotStoresItemEvidence(t *testing.T) {
	score := 1.0
	suggestedScore := 0.5
	confirmed := true
	snapshot, err := mergeVBMAPPDraftInputSnapshot(nil, vbmappscore.AssessmentInput{
		MilestoneScores: map[string]float64{"MAND_01M": 1},
	}, vbmappItemResponsePatch{
		ModuleCode:       "里程碑评估",
		ItemCode:         "mand_01m",
		Score:            &score,
		SuggestedScore:   &suggestedScore,
		TeacherConfirmed: &confirmed,
		RecordStatus:     "confirmed",
		Evidence: map[string]any{
			"spokenWords": []any{"饼干", "球"},
			"promptLevel": "无肢体辅助",
		},
	})
	if err != nil {
		t.Fatalf("mergeVBMAPPDraftInputSnapshot returned error: %v", err)
	}
	raw, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	itemResponses := decoded["itemResponses"].(map[string]any)
	milestones := itemResponses["milestones"].(map[string]any)
	response := milestones["MAND_01M"].(map[string]any)
	if response["recordStatus"] != "confirmed" || response["teacherConfirmed"] != true {
		t.Fatalf("unexpected item response status: %+v", response)
	}
	evidence := response["evidence"].(map[string]any)
	words := evidence["spokenWords"].([]any)
	if len(words) != 2 || words[0] != "饼干" {
		t.Fatalf("unexpected evidence: %+v", evidence)
	}
}

func TestBuildVBMAPPAssessmentHistoryCalculatesRecordChanges(t *testing.T) {
	firstResult := VBMAPPScoreResponse{
		VBMAPPScoreDataInfo: VBMAPPScoreDataInfo{
			ScaleCode:      vbmappScaleCode,
			ScaleVersion:   vbmappScaleVersion,
			AssessmentName: vbmappAssessmentName,
		},
		Result: vbmappscore.AssessmentResult{
			Milestones: vbmappscore.MilestoneModuleResult{
				TotalScore: 10,
				MaxScore:   170,
				Domains: []vbmappscore.DomainScoreResult{
					{DomainCode: "MAND", DomainName: "提要求", TotalScore: 3, MaxScore: 15, Percent: 20},
				},
			},
			Barriers: vbmappscore.BarrierModuleResult{TotalScore: 6, MaxScore: 96},
			Transition: vbmappscore.TransitionModuleResult{
				TotalScore: 18,
				MaxScore:   90,
			},
		},
	}
	secondResult := firstResult
	secondResult.Result.Milestones.TotalScore = 14
	secondResult.Result.Milestones.Domains = []vbmappscore.DomainScoreResult{
		{DomainCode: "MAND", DomainName: "提要求", TotalScore: 5, MaxScore: 15, Percent: 33.3},
	}
	secondResult.Result.Barriers.TotalScore = 4
	secondResult.Result.Transition.TotalScore = 20

	firstDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	secondDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	firstRaw, err := json.Marshal(firstResult)
	if err != nil {
		t.Fatalf("marshal first result: %v", err)
	}
	secondRaw, err := json.Marshal(secondResult)
	if err != nil {
		t.Fatalf("marshal second result: %v", err)
	}
	history, err := buildVBMAPPAssessmentHistory(1001, []model.AssessmentRecordDetailVO{
		{
			AssessmentRecordSummaryVO: model.AssessmentRecordSummaryVO{
				ID:             1,
				AssessmentDate: &firstDate,
			},
			ResultJSON: firstRaw,
		},
		{
			AssessmentRecordSummaryVO: model.AssessmentRecordSummaryVO{
				ID:             2,
				AssessmentDate: &secondDate,
			},
			ResultJSON: secondRaw,
		},
	})
	if err != nil {
		t.Fatalf("buildVBMAPPAssessmentHistory returned error: %v", err)
	}
	if len(history.Records) != 2 || history.Records[1].MilestoneChange == nil || *history.Records[1].MilestoneChange != 4 {
		t.Fatalf("unexpected milestone history: %+v", history)
	}
	if history.Records[1].BarrierChange == nil || *history.Records[1].BarrierChange != -2 {
		t.Fatalf("unexpected barrier history: %+v", history.Records[1])
	}
	if len(history.Records[1].Domains) != 1 || history.Records[1].Domains[0].Change == nil || *history.Records[1].Domains[0].Change != 2 {
		t.Fatalf("unexpected domain history: %+v", history.Records[1].Domains)
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
