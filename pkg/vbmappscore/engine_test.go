package vbmappscore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadGeneratedDraftsAndBuildEngine(t *testing.T) {
	domains, milestones, rules, barriers, transitions := loadGeneratedDrafts(t)

	if len(domains) != 16 {
		t.Fatalf("unexpected domain count: %d", len(domains))
	}
	if len(milestones) != 170 {
		t.Fatalf("unexpected milestone count: %d", len(milestones))
	}
	if len(rules) != 170 {
		t.Fatalf("unexpected milestone scoring rule count: %d", len(rules))
	}
	if len(barriers) != 24 {
		t.Fatalf("unexpected barrier count: %d", len(barriers))
	}
	if len(transitions) != 18 {
		t.Fatalf("unexpected transition count: %d", len(transitions))
	}

	if _, err := NewEngine(domains, milestones, rules, barriers, transitions); err != nil {
		t.Fatalf("NewEngine with generated drafts: %v", err)
	}
}

func TestScoreAggregatesMilestonesBarriersAndTransitionSuggestions(t *testing.T) {
	domains, milestones, rules, barriers, transitions := loadGeneratedDrafts(t)
	engine, err := NewEngine(domains, milestones, rules, barriers, transitions)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}

	result, err := engine.Score(AssessmentInput{
		MilestoneScores: map[string]float64{
			"MAND_01M":   1,
			"MAND_02M":   0.5,
			"TACT_01M":   0,
			"GROUP_11M":  1,
			"SOCIAL_01M": 0.5,
		},
		BarrierScores: map[string]int{
			"B01": 3,
			"B02": 2,
			"B03": 1,
		},
		TransitionScores: map[string]int{
			"T06": 3,
		},
		PreviousMilestoneScores: map[string]float64{
			"MAND_02M": 0,
		},
		PreviousBarrierScores: map[string]int{
			"B01": 2,
		},
	})
	if err != nil {
		t.Fatalf("Score returned error: %v", err)
	}

	if result.ScaleCode != ScaleCode || result.ScaleVersion != DefaultScaleVersion {
		t.Fatalf("unexpected scale metadata: %s %s", result.ScaleCode, result.ScaleVersion)
	}
	if result.Milestones.TotalScore != 3 || result.Milestones.AnsweredItems != 5 {
		t.Fatalf("unexpected milestone totals: %+v", result.Milestones)
	}
	mand := findDomain(t, result.Milestones, "MAND")
	if mand.TotalScore != 1.5 || mand.AnsweredItems != 2 {
		t.Fatalf("unexpected MAND result: %+v", mand)
	}
	if len(result.Milestones.LowItems) != 3 {
		t.Fatalf("expected three low milestone items, got %d", len(result.Milestones.LowItems))
	}
	if result.Barriers.TotalScore != 6 || result.Barriers.AnsweredItems != 3 {
		t.Fatalf("unexpected barrier totals: %+v", result.Barriers)
	}
	if len(result.Barriers.HighRiskItems) != 1 || result.Barriers.HighRiskItems[0].BarrierCode != "B01" {
		t.Fatalf("unexpected high-risk barriers: %+v", result.Barriers.HighRiskItems)
	}
	if result.Transition.TotalScore != 3 || result.Transition.AnsweredItems != 1 {
		t.Fatalf("unexpected transition totals: %+v", result.Transition)
	}

	suggestions := suggestionsByCode(result.Transition.Suggestions)
	assertSuggestion(t, suggestions, "T01", 1)
	assertSuggestion(t, suggestions, "T02", 5)
	assertSuggestion(t, suggestions, "T03", 2)
	assertSuggestion(t, suggestions, "T04", 1)
	assertSuggestion(t, suggestions, "T05", 1)
}

func TestTransitionSuggestionsUseCompletedTotals(t *testing.T) {
	domains, milestones, rules, barriers, transitions := loadGeneratedDrafts(t)
	engine, err := NewEngine(domains, milestones, rules, barriers, transitions)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}

	milestoneScores := make(map[string]float64, len(milestones))
	for _, item := range milestones {
		milestoneScores[item.MilestoneID] = 1
	}
	barrierScores := make(map[string]int, len(barriers))
	for _, barrier := range barriers {
		barrierScores[barrier.BarrierCode] = 0
	}

	result, err := engine.Score(AssessmentInput{
		MilestoneScores: milestoneScores,
		BarrierScores:   barrierScores,
	})
	if err != nil {
		t.Fatalf("Score returned error: %v", err)
	}
	if result.Milestones.TotalScore != 170 || !result.Milestones.Complete {
		t.Fatalf("unexpected milestone result: total=%v complete=%v", result.Milestones.TotalScore, result.Milestones.Complete)
	}
	if result.Barriers.TotalScore != 0 || !result.Barriers.Complete {
		t.Fatalf("unexpected barrier result: total=%v complete=%v", result.Barriers.TotalScore, result.Barriers.Complete)
	}

	suggestions := suggestionsByCode(result.Transition.Suggestions)
	assertSuggestion(t, suggestions, "T01", 5)
	assertSuggestion(t, suggestions, "T02", 5)
	assertSuggestion(t, suggestions, "T03", 5)
	assertSuggestion(t, suggestions, "T04", 5)
	assertSuggestion(t, suggestions, "T05", 5)
}

func TestScoreRejectsInvalidValues(t *testing.T) {
	domains, milestones, rules, barriers, transitions := loadGeneratedDrafts(t)
	engine, err := NewEngine(domains, milestones, rules, barriers, transitions)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}

	if _, err := engine.Score(AssessmentInput{MilestoneScores: map[string]float64{"MAND_01M": 0.25}}); err == nil {
		t.Fatal("expected invalid milestone score error")
	}
	if _, err := engine.Score(AssessmentInput{BarrierScores: map[string]int{"B01": 5}}); err == nil {
		t.Fatal("expected invalid barrier score error")
	}
	if _, err := engine.Score(AssessmentInput{TransitionScores: map[string]int{"T01": 0}}); err == nil {
		t.Fatal("expected invalid transition score error")
	}
}

func TestScoreAutoCalculatesMAND08MFromEvidence(t *testing.T) {
	domains, milestones, rules, barriers, transitions := loadGeneratedDrafts(t)
	engine, err := NewEngine(domains, milestones, rules, barriers, transitions)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}

	result, err := engine.Score(AssessmentInput{
		ItemResponses: map[string]map[string]map[string]any{
			ModuleMilestones: {
				"MAND_08M": {
					"evidence": map[string]any{
						"mandEvents": []any{
							map[string]any{"utterance": "跑快点", "promptLevel": "无辅助", "functional": true},
							map[string]any{"utterance": "该我了", "promptLevel": "无辅助", "functional": true},
							map[string]any{"utterance": "泡泡", "promptLevel": "无辅助", "functional": true},
							map[string]any{"utterance": "饼干", "promptLevel": "无辅助", "functional": true},
							map[string]any{"utterance": "倒果汁", "promptLevel": "无辅助", "functional": true},
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Score returned error: %v", err)
	}

	item := findMilestone(t, result.Milestones, "MAND_08M")
	if item.Score == nil || *item.Score != 1 {
		t.Fatalf("expected MAND_08M score 1 from evidence, got %+v", item)
	}
}

func TestScoreAutoCalculatesMAND04MFromEvidence(t *testing.T) {
	domains, milestones, rules, barriers, transitions := loadGeneratedDrafts(t)
	engine, err := NewEngine(domains, milestones, rules, barriers, transitions)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}

	result, err := engine.Score(AssessmentInput{
		ItemResponses: map[string]map[string]map[string]any{
			ModuleMilestones: {
				"MAND_04M": {
					"evidence": map[string]any{
						"timer": map[string]any{
							"startTime":      "2026-05-21T10:00:00Z",
							"plannedMinutes": 60,
						},
						"mandEvents": []any{
							map[string]any{"utterance": "泡泡", "environment": "呈现物品", "responseMode": "自发要求", "recordedAtIso": "2026-05-21T10:05:00Z", "functional": true},
							map[string]any{"utterance": "饼干", "environment": "呈现物品", "responseMode": "自发要求", "recordedAtIso": "2026-05-21T10:10:00Z", "functional": true},
							map[string]any{"utterance": "车车", "environment": "呈现物品", "responseMode": "自发要求", "recordedAtIso": "2026-05-21T10:20:00Z", "functional": true},
							map[string]any{"utterance": "积木", "environment": "呈现物品", "responseMode": "自发要求", "recordedAtIso": "2026-05-21T10:30:00Z", "functional": true},
							map[string]any{"utterance": "球", "environment": "呈现物品", "responseMode": "自发要求", "recordedAtIso": "2026-05-21T10:40:00Z", "functional": true},
							map[string]any{"utterance": "不要这个", "environment": "未呈现物品", "responseMode": "自发要求", "recordedAtIso": "2026-05-21T10:15:00Z", "functional": true},
							map[string]any{"utterance": "再来", "environment": "呈现物品", "promptLevel": "提问下", "responseMode": "提问下要求", "recordedAtIso": "2026-05-21T10:25:00Z", "functional": true},
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Score returned error: %v", err)
	}

	item := findMilestone(t, result.Milestones, "MAND_04M")
	if item.Score == nil || *item.Score != 1 {
		t.Fatalf("expected MAND_04M score 1 from evidence, got %+v", item)
	}
}

func TestScoreAutoCalculatesMAND03MFromGeneralizationEvidence(t *testing.T) {
	domains, milestones, rules, barriers, transitions := loadGeneratedDrafts(t)
	engine, err := NewEngine(domains, milestones, rules, barriers, transitions)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}

	result, err := engine.Score(AssessmentInput{
		ItemResponses: map[string]map[string]map[string]any{
			ModuleMilestones: {
				"MAND_03M": {
					"evidence": map[string]any{
						"mandEvents": []any{
							map[string]any{"utterance": "泡泡", "person": "妈妈", "functional": true},
							map[string]any{"utterance": "泡泡", "person": "老师", "functional": true},
							map[string]any{"utterance": "泡泡", "setting": "客厅", "functional": true},
							map[string]any{"utterance": "泡泡", "setting": "教室", "functional": true},
							map[string]any{"utterance": "泡泡", "example": "红瓶泡泡", "functional": true},
							map[string]any{"utterance": "泡泡", "example": "蓝瓶泡泡", "functional": true},
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Score returned error: %v", err)
	}

	item := findMilestone(t, result.Milestones, "MAND_03M")
	if item.Score == nil || *item.Score != 1 {
		t.Fatalf("expected MAND_03M score 1 from evidence, got %+v", item)
	}
}

func TestScoreAutoCalculatesMAND03MHalfPointFromGeneralizationEvidence(t *testing.T) {
	domains, milestones, rules, barriers, transitions := loadGeneratedDrafts(t)
	engine, err := NewEngine(domains, milestones, rules, barriers, transitions)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}

	result, err := engine.Score(AssessmentInput{
		ItemResponses: map[string]map[string]map[string]any{
			ModuleMilestones: {
				"MAND_03M": {
					"evidence": map[string]any{
						"mandEvents": []any{
							map[string]any{"utterance": "球", "person": "妈妈", "functional": true},
							map[string]any{"utterance": "球", "setting": "客厅", "functional": true},
							map[string]any{"utterance": "球", "example": "大龙球", "functional": true},
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Score returned error: %v", err)
	}

	item := findMilestone(t, result.Milestones, "MAND_03M")
	if item.Score == nil || *item.Score != 0.5 {
		t.Fatalf("expected MAND_03M score 0.5 from evidence, got %+v", item)
	}
}

func TestScoreAutoCalculatesMAND08MHalfPointWhenOnlyTwoQualified(t *testing.T) {
	domains, milestones, rules, barriers, transitions := loadGeneratedDrafts(t)
	engine, err := NewEngine(domains, milestones, rules, barriers, transitions)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}

	result, err := engine.Score(AssessmentInput{
		ItemResponses: map[string]map[string]map[string]any{
			ModuleMilestones: {
				"MAND_08M": {
					"evidence": map[string]any{
						"mandEvents": []any{
							map[string]any{"utterance": "泡泡", "promptLevel": "无辅助", "functional": true},
							map[string]any{"utterance": "饼干", "promptLevel": "无辅助", "functional": true},
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Score returned error: %v", err)
	}

	item := findMilestone(t, result.Milestones, "MAND_08M")
	if item.Score == nil || *item.Score != 0.5 {
		t.Fatalf("expected MAND_08M score 0.5 from evidence, got %+v", item)
	}
}

func TestScoreAutoCalculatesMAND09MFiltersPromptedAndOutOfWindowEvents(t *testing.T) {
	domains, milestones, rules, barriers, transitions := loadGeneratedDrafts(t)
	engine, err := NewEngine(domains, milestones, rules, barriers, transitions)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}

	result, err := engine.Score(AssessmentInput{
		ItemResponses: map[string]map[string]map[string]any{
			ModuleMilestones: {
				"MAND_09M": {
					"evidence": map[string]any{
						"timer": map[string]any{
							"startTime":      "2026-05-21T10:00:00Z",
							"plannedMinutes": 30,
						},
						"mandEvents": []any{
							map[string]any{"utterance": "一起玩", "promptLevel": "自发地", "responseMode": "自发要求", "recordedAtIso": "2026-05-21T10:01:00Z", "functional": true},
							map[string]any{"utterance": "打开", "promptLevel": "自发地", "responseMode": "自发要求", "recordedAtIso": "2026-05-21T10:02:00Z", "functional": true},
							map[string]any{"utterance": "我要书", "promptLevel": "自发地", "responseMode": "自发要求", "recordedAtIso": "2026-05-21T10:03:00Z", "functional": true},
							map[string]any{"utterance": "推高点", "promptLevel": "自发地", "responseMode": "自发要求", "recordedAtIso": "2026-05-21T10:04:00Z", "functional": true},
							map[string]any{"utterance": "帮我开", "promptLevel": "自发地", "responseMode": "自发要求", "recordedAtIso": "2026-05-21T10:05:00Z", "functional": true},
							map[string]any{"utterance": "该我了", "promptLevel": "自发地", "responseMode": "自发要求", "recordedAtIso": "2026-05-21T10:06:00Z", "functional": true},
							map[string]any{"utterance": "再来", "promptLevel": "自发地", "responseMode": "自发要求", "recordedAtIso": "2026-05-21T10:07:00Z", "functional": true},
							map[string]any{"utterance": "不要停", "promptLevel": "自发地", "responseMode": "自发要求", "recordedAtIso": "2026-05-21T10:08:00Z", "functional": true},
							map[string]any{"utterance": "给我球", "promptLevel": "提问下", "responseMode": "提问下要求", "recordedAtIso": "2026-05-21T10:09:00Z", "functional": true},
							map[string]any{"utterance": "我要拼图", "promptLevel": "自发地", "responseMode": "自发要求", "recordedAtIso": "2026-05-21T10:31:00Z", "functional": true},
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Score returned error: %v", err)
	}

	item := findMilestone(t, result.Milestones, "MAND_09M")
	if item.Score == nil || *item.Score != 0.5 {
		t.Fatalf("expected MAND_09M score 0.5 from evidence, got %+v", item)
	}
}

func TestScoreAutoCalculatesMAND08MDoesNotCountWoXiangYaoAsMultiWord(t *testing.T) {
	domains, milestones, rules, barriers, transitions := loadGeneratedDrafts(t)
	engine, err := NewEngine(domains, milestones, rules, barriers, transitions)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}

	result, err := engine.Score(AssessmentInput{
		ItemResponses: map[string]map[string]map[string]any{
			ModuleMilestones: {
				"MAND_08M": {
					"evidence": map[string]any{
						"mandEvents": []any{
							map[string]any{"utterance": "跑快点", "promptLevel": "无辅助", "functional": true},
							map[string]any{"utterance": "泡泡", "promptLevel": "无辅助", "functional": true},
							map[string]any{"utterance": "饼干", "promptLevel": "无辅助", "functional": true},
							map[string]any{"utterance": "车车", "promptLevel": "无辅助", "functional": true},
							map[string]any{"utterance": "我想要书", "promptLevel": "无辅助", "functional": true},
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Score returned error: %v", err)
	}

	item := findMilestone(t, result.Milestones, "MAND_08M")
	if item.Score == nil || *item.Score != 0.5 {
		t.Fatalf("expected MAND_08M score 0.5 when only one true multi-word request exists, got %+v", item)
	}
}

func loadGeneratedDrafts(t *testing.T) (
	[]DomainDefinition,
	[]MilestoneItemDefinition,
	[]MilestoneScoringRule,
	[]BarrierDefinition,
	[]TransitionDefinition,
) {
	t.Helper()
	root := filepath.Join("..", "..")
	dataDir := filepath.Join(root, "docs", "vbmapp")
	for _, name := range []string{
		"domains.json",
		"milestone-items.json",
		"milestone-scoring-rules.json",
		"barriers.json",
		"transition.json",
	} {
		if _, err := os.Stat(filepath.Join(dataDir, name)); err != nil {
			t.Skipf("generated VB-MAPP data not present: %s", filepath.Join(dataDir, name))
		}
	}

	domains, err := LoadDomainDefinitionsFile(filepath.Join(dataDir, "domains.json"))
	if err != nil {
		t.Fatalf("LoadDomainDefinitionsFile: %v", err)
	}
	milestones, err := LoadMilestoneItemDefinitionsFile(filepath.Join(dataDir, "milestone-items.json"))
	if err != nil {
		t.Fatalf("LoadMilestoneItemDefinitionsFile: %v", err)
	}
	rules, err := LoadMilestoneScoringRulesFile(filepath.Join(dataDir, "milestone-scoring-rules.json"))
	if err != nil {
		t.Fatalf("LoadMilestoneScoringRulesFile: %v", err)
	}
	barriers, err := LoadBarrierDefinitionsFile(filepath.Join(dataDir, "barriers.json"))
	if err != nil {
		t.Fatalf("LoadBarrierDefinitionsFile: %v", err)
	}
	transitions, err := LoadTransitionDefinitionsFile(filepath.Join(dataDir, "transition.json"))
	if err != nil {
		t.Fatalf("LoadTransitionDefinitionsFile: %v", err)
	}
	return domains, milestones, rules, barriers, transitions
}

func findDomain(t *testing.T, result MilestoneModuleResult, code string) DomainScoreResult {
	t.Helper()
	for _, domain := range result.Domains {
		if domain.DomainCode == code {
			return domain
		}
	}
	t.Fatalf("domain %s not found", code)
	return DomainScoreResult{}
}

func findMilestone(t *testing.T, result MilestoneModuleResult, milestoneID string) MilestoneScoreResult {
	t.Helper()
	for _, item := range result.Items {
		if item.MilestoneID == milestoneID {
			return item
		}
	}
	t.Fatalf("milestone %s not found", milestoneID)
	return MilestoneScoreResult{}
}

func suggestionsByCode(suggestions []TransitionSuggestion) map[string]TransitionSuggestion {
	result := make(map[string]TransitionSuggestion, len(suggestions))
	for _, suggestion := range suggestions {
		result[suggestion.TransitionCode] = suggestion
	}
	return result
}

func assertSuggestion(t *testing.T, suggestions map[string]TransitionSuggestion, code string, score int) {
	t.Helper()
	suggestion, ok := suggestions[code]
	if !ok {
		t.Fatalf("expected suggestion %s", code)
	}
	if suggestion.Score != score {
		t.Fatalf("unexpected suggestion %s score: got %d want %d; %+v", code, suggestion.Score, score, suggestion)
	}
}
