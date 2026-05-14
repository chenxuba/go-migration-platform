package autismdevscore

import (
	"path/filepath"
	"testing"
	"time"
)

func TestLoadGeneratedItemBank(t *testing.T) {
	items, err := LoadItemDefinitionsFile(filepath.Join("..", "..", "docs", "autismdev-item-bank-draft.json"))
	if err != nil {
		t.Fatalf("LoadItemDefinitionsFile returned error: %v", err)
	}
	if len(items) != ExpectedItemDefinition {
		t.Fatalf("expected %d items, got %d", ExpectedItemDefinition, len(items))
	}
	engine, err := NewEngine(items)
	if err != nil {
		t.Fatalf("NewEngine returned error: %v", err)
	}
	if len(engine.itemsByDomain[DomainEmotionBehavior]) != 52 {
		t.Fatalf("expected 52 EB items, got %d", len(engine.itemsByDomain[DomainEmotionBehavior]))
	}
}

func TestScoreCountsPEFAndAMSDomains(t *testing.T) {
	engine, err := NewEngine(fixtureItems())
	if err != nil {
		t.Fatalf("NewEngine returned error: %v", err)
	}
	result, err := engine.Score(AssessmentInput{
		BirthDate:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		AssessmentDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		ItemScores: map[int]string{
			1: "P",
			2: "E",
			3: "F",
			4: "X",
			5: "A",
			6: "M",
			7: "S",
		},
	})
	if err != nil {
		t.Fatalf("Score returned error: %v", err)
	}
	if !result.Complete {
		t.Fatalf("expected complete result, got missing=%d", result.MissingItemCount)
	}
	if result.Development.PCount != 1 || result.Development.ECount != 1 || result.Development.FCount != 1 || result.Development.XCount != 1 {
		t.Fatalf("unexpected development counts: %+v", result.Development)
	}
	if result.Development.RawScore != 1 || result.Development.PECount != 2 || result.Development.ScorableItemCount != 3 {
		t.Fatalf("unexpected development score fields: %+v", result.Development)
	}
	if result.Behavior.ACount != 1 || result.Behavior.MCount != 1 || result.Behavior.SCount != 1 {
		t.Fatalf("unexpected behavior counts: %+v", result.Behavior)
	}
	if result.Behavior.AdaptiveCount != 2 || result.Behavior.AbnormalCount != 2 {
		t.Fatalf("unexpected behavior aggregate fields: %+v", result.Behavior)
	}
}

func TestScoreRejectsInvalidDomainScore(t *testing.T) {
	engine, err := NewEngine(fixtureItems())
	if err != nil {
		t.Fatalf("NewEngine returned error: %v", err)
	}
	_, err = engine.Score(AssessmentInput{
		BirthDate:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		AssessmentDate: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		ItemScores:     map[int]string{1: "A"},
	})
	if err == nil {
		t.Fatalf("expected invalid score error")
	}
}

func TestScoreRejectsChildrenOverSixYearsOld(t *testing.T) {
	engine, err := NewEngine(fixtureItems())
	if err != nil {
		t.Fatalf("NewEngine returned error: %v", err)
	}
	_, err = engine.Score(AssessmentInput{
		BirthDate:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		AssessmentDate: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		ItemScores:     map[int]string{1: "P"},
	})
	if err == nil {
		t.Fatalf("expected over-age error")
	}
}

func TestScoreUsesQuestionDisplayPreference(t *testing.T) {
	engine, err := NewEngine([]ItemDefinition{
		{ItemNo: 1, DomainCode: DomainSensory, DomainName: "感知觉", ScoreType: ScoreTypePEF, AgeMinMonth: 0, AgeMaxMonth: 12},
		{ItemNo: 2, DomainCode: DomainSensory, DomainName: "感知觉", ScoreType: ScoreTypePEF, AgeMinMonth: 24, AgeMaxMonth: 36},
		{ItemNo: 3, DomainCode: DomainSensory, DomainName: "感知觉", ScoreType: ScoreTypePEF, AgeMinMonth: 60, AgeMaxMonth: 72},
		{ItemNo: 4, DomainCode: DomainEmotionBehavior, DomainName: "情绪与行为", ScoreType: ScoreTypeAMS},
	})
	if err != nil {
		t.Fatalf("NewEngine returned error: %v", err)
	}
	result, err := engine.Score(AssessmentInput{
		BirthDate:                 time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		AssessmentDate:            time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		QuestionDisplayPreference: QuestionDisplayPreferenceAgeAndBelow,
		ItemScores: map[int]string{
			1: "P",
			2: "P",
			4: "A",
		},
	})
	if err != nil {
		t.Fatalf("Score returned error: %v", err)
	}
	if !result.Complete || result.ItemCount != 3 || result.MissingItemCount != 0 {
		t.Fatalf("expected age-and-below range complete, got item=%d missing=%d complete=%v", result.ItemCount, result.MissingItemCount, result.Complete)
	}
}

func fixtureItems() []ItemDefinition {
	return []ItemDefinition{
		{ItemNo: 1, DomainCode: DomainSensory, DomainName: "感知觉", ScoreType: ScoreTypePEF},
		{ItemNo: 2, DomainCode: DomainSensory, DomainName: "感知觉", ScoreType: ScoreTypePEF},
		{ItemNo: 3, DomainCode: DomainSensory, DomainName: "感知觉", ScoreType: ScoreTypePEF},
		{ItemNo: 4, DomainCode: DomainSensory, DomainName: "感知觉", ScoreType: ScoreTypePEF},
		{ItemNo: 5, DomainCode: DomainEmotionBehavior, DomainName: "情绪与行为", ScoreType: ScoreTypeAMS},
		{ItemNo: 6, DomainCode: DomainEmotionBehavior, DomainName: "情绪与行为", ScoreType: ScoreTypeAMS},
		{ItemNo: 7, DomainCode: DomainEmotionBehavior, DomainName: "情绪与行为", ScoreType: ScoreTypeAMS},
	}
}
