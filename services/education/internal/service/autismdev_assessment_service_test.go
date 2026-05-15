package service

import (
	"encoding/json"
	"testing"
	"time"

	"go-migration-platform/pkg/autismdevscore"
	"go-migration-platform/services/education/internal/model"
)

func TestBuildAutismDevAssessmentFormTemplateSummary(t *testing.T) {
	template, err := buildAutismDevAssessmentFormTemplateSummary()
	if err != nil {
		t.Fatalf("buildAutismDevAssessmentFormTemplateSummary returned error: %v", err)
	}
	if template.ScaleCode != autismDevScaleCode {
		t.Fatalf("unexpected scale code: %s", template.ScaleCode)
	}
	if template.ItemCount != autismdevscore.ExpectedItemDefinition {
		t.Fatalf("expected %d items, got %d", autismdevscore.ExpectedItemDefinition, template.ItemCount)
	}
	if len(template.Domains) != 8 {
		t.Fatalf("expected 8 domains, got %d", len(template.Domains))
	}
	if len(template.DomainGroups) != 8 || len(template.DomainGroups[0].Items) != 55 {
		t.Fatalf("unexpected domain groups: groups=%d firstItems=%d", len(template.DomainGroups), len(template.DomainGroups[0].Items))
	}
	if template.DomainGroups[0].DomainCode != autismdevscore.DomainSensory {
		t.Fatalf("unexpected first domain code: %s", template.DomainGroups[0].DomainCode)
	}
	if template.DomainGroups[0].DomainName != "感知觉" || template.DomainGroups[0].Title != "感知觉" {
		t.Fatalf("domain display name should not include code: title=%q domain=%q", template.DomainGroups[0].Title, template.DomainGroups[0].DomainName)
	}
	if template.DomainGroups[0].Items[0].DomainName != "感知觉" {
		t.Fatalf("item domain display name should not include code: %q", template.DomainGroups[0].Items[0].DomainName)
	}
}

func TestBuildAutismDevAssessmentDraftProgressComplete(t *testing.T) {
	data, err := loadAutismDevStaticData()
	if err != nil {
		t.Fatalf("loadAutismDevStaticData returned error: %v", err)
	}
	itemScores := make(map[int]string, len(data.items))
	for _, item := range data.items {
		switch item.ScoreType {
		case autismdevscore.ScoreTypeAMS:
			itemScores[item.ItemNo] = autismdevscore.ScoreA
		default:
			itemScores[item.ItemNo] = autismdevscore.ScoreP
		}
	}
	birthDate := time.Date(2021, 5, 1, 0, 0, 0, 0, time.UTC)
	assessmentDate := time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC)
	progress, err := buildAutismDevAssessmentDraftProgress(&birthDate, &assessmentDate, autismdevscore.QuestionDisplayPreferenceAll, itemScores)
	if err != nil {
		t.Fatalf("buildAutismDevAssessmentDraftProgress returned error: %v", err)
	}
	if !progress.Complete || !progress.CanScore {
		t.Fatalf("expected complete progress: %+v", progress)
	}
	if progress.AnsweredItemCount != autismdevscore.ExpectedItemDefinition || progress.MissingItemCount != 0 {
		t.Fatalf("unexpected progress counts: answered=%d missing=%d", progress.AnsweredItemCount, progress.MissingItemCount)
	}
}

func TestBuildAutismDevAssessmentDraftProgressUsesQuestionDisplayPreference(t *testing.T) {
	data, err := loadAutismDevStaticData()
	if err != nil {
		t.Fatalf("loadAutismDevStaticData returned error: %v", err)
	}
	birthDate := time.Date(2021, 5, 1, 0, 0, 0, 0, time.UTC)
	assessmentDate := time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC)
	requiredItemNos := autismDevRequiredItemNos(
		data.items,
		&birthDate,
		&assessmentDate,
		autismdevscore.QuestionDisplayPreferenceAgeAndBelow,
	)
	if len(requiredItemNos) <= 0 || len(requiredItemNos) >= len(data.items) {
		t.Fatalf("expected age-and-below range to be a subset, got %d/%d", len(requiredItemNos), len(data.items))
	}
	itemScores := make(map[int]string, len(requiredItemNos))
	for _, item := range data.items {
		if !requiredItemNos[item.ItemNo] {
			continue
		}
		switch item.ScoreType {
		case autismdevscore.ScoreTypeAMS:
			itemScores[item.ItemNo] = autismdevscore.ScoreA
		default:
			itemScores[item.ItemNo] = autismdevscore.ScoreP
		}
	}
	progress, err := buildAutismDevAssessmentDraftProgress(
		&birthDate,
		&assessmentDate,
		autismdevscore.QuestionDisplayPreferenceAgeAndBelow,
		itemScores,
	)
	if err != nil {
		t.Fatalf("buildAutismDevAssessmentDraftProgress returned error: %v", err)
	}
	if !progress.Complete || !progress.CanScore {
		t.Fatalf("expected complete age-and-below progress: %+v", progress)
	}
	if progress.ItemCount != len(requiredItemNos) || progress.AnsweredItemCount != len(requiredItemNos) || progress.MissingItemCount != 0 {
		t.Fatalf("unexpected progress counts: item=%d answered=%d missing=%d required=%d", progress.ItemCount, progress.AnsweredItemCount, progress.MissingItemCount, len(requiredItemNos))
	}
}

func TestAutismDevTrainingCurrentRecordDomainsUsesCurrentScopeOrScores(t *testing.T) {
	data, err := loadAutismDevStaticData()
	if err != nil {
		t.Fatalf("loadAutismDevStaticData returned error: %v", err)
	}

	scopedInput, _ := json.Marshal(map[string]any{
		"scopeMode":        "custom",
		"scopeDomainCodes": []string{autismdevscore.DomainLanguageComm, autismdevscore.DomainDailyLiving},
		"itemScoreList": []map[string]any{
			{"itemNo": 1, "score": autismdevscore.ScoreP},
		},
	})
	scopedDomains := autismDevTrainingCurrentRecordDomains(model.AssessmentRecordDetailVO{
		InputJSON: scopedInput,
	}, data)
	if len(scopedDomains) != 2 ||
		scopedDomains[0] != autismdevscore.DomainLanguageComm ||
		scopedDomains[1] != autismdevscore.DomainDailyLiving {
		t.Fatalf("expected scoped domains only, got %#v", scopedDomains)
	}

	scoreOnlyInput, _ := json.Marshal(map[string]any{
		"itemScoreList": []map[string]any{
			{"itemNo": 1, "score": autismdevscore.ScoreP},
			{"itemNo": 442, "score": autismdevscore.ScoreM},
		},
	})
	scoreDomains := autismDevTrainingCurrentRecordDomains(model.AssessmentRecordDetailVO{
		InputJSON: scoreOnlyInput,
	}, data)
	if len(scoreDomains) != 2 ||
		scoreDomains[0] != autismdevscore.DomainSensory ||
		scoreDomains[1] != autismdevscore.DomainEmotionBehavior {
		t.Fatalf("expected domains inferred from scored items, got %#v", scoreDomains)
	}
}

func TestAutismDevTrainingEffectForScores(t *testing.T) {
	tests := []struct {
		name      string
		scoreType string
		before    string
		after     string
		want      string
	}{
		{name: "PEF E to P significant", scoreType: autismdevscore.ScoreTypePEF, before: autismdevscore.ScoreE, after: autismdevscore.ScoreP, want: "significant"},
		{name: "PEF F to P significant", scoreType: autismdevscore.ScoreTypePEF, before: autismdevscore.ScoreF, after: autismdevscore.ScoreP, want: "significant"},
		{name: "PEF F to E effective", scoreType: autismdevscore.ScoreTypePEF, before: autismdevscore.ScoreF, after: autismdevscore.ScoreE, want: "effective"},
		{name: "AMS S to A significant", scoreType: autismdevscore.ScoreTypeAMS, before: autismdevscore.ScoreS, after: autismdevscore.ScoreA, want: "significant"},
		{name: "AMS M to A significant", scoreType: autismdevscore.ScoreTypeAMS, before: autismdevscore.ScoreM, after: autismdevscore.ScoreA, want: "significant"},
		{name: "AMS S to M effective", scoreType: autismdevscore.ScoreTypeAMS, before: autismdevscore.ScoreS, after: autismdevscore.ScoreM, want: "effective"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := autismDevTrainingEffectForScores(tt.scoreType, tt.before, tt.after)
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestAutismDevTrainingEffectTemplateLayouts(t *testing.T) {
	for _, page := range autismDevTrainingEffectTemplatePages() {
		layout, err := autismDevTrainingEffectTemplateLayoutForPage(page)
		if err != nil {
			t.Fatalf("layout page %d returned error: %v", page.PageNo, err)
		}
		if len(layout.ColumnXs) != 11 {
			t.Fatalf("page %d expected 11 table columns, got %d", page.PageNo, len(layout.ColumnXs))
		}
		itemCount := page.LastItemNo - page.FirstItemNo + 1
		if len(layout.RowBounds) != itemCount+1 {
			t.Fatalf("page %d expected %d row bounds, got %d", page.PageNo, itemCount+1, len(layout.RowBounds))
		}
		for index := 1; index < len(layout.RowBounds); index++ {
			gap := layout.RowBounds[index] - layout.RowBounds[index-1]
			if gap < 20 || gap > 125 {
				t.Fatalf("page %d row bound %d has suspicious gap %.1f", page.PageNo, index, gap)
			}
		}
	}
}
