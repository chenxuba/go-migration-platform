package service

import (
	"testing"
	"time"

	"go-migration-platform/pkg/autismdevscore"
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
	progress, err := buildAutismDevAssessmentDraftProgress(&birthDate, &assessmentDate, itemScores)
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
