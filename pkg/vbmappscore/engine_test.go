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
