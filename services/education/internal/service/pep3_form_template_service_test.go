package service

import (
	"strings"
	"testing"

	"go-migration-platform/services/education/internal/model"
)

func TestBuildPEP3AssessmentFormTemplate(t *testing.T) {
	template, err := buildPEP3AssessmentFormTemplate()
	if err != nil {
		t.Fatalf("buildPEP3AssessmentFormTemplate returned error: %v", err)
	}
	if template.TemplateCode != "PEP3_ASSESSMENT_FORM" || template.ItemCount != 172 {
		t.Fatalf("unexpected template metadata: %+v", template)
	}
	if len(template.ItemGroups) != len(pep3BookletItemRanges()) {
		t.Fatalf("unexpected item groups: %d", len(template.ItemGroups))
	}
	firstGroup := template.ItemGroups[0]
	if firstGroup.StartItemNo != 1 || firstGroup.EndItemNo != 14 || len(firstGroup.Items) != 14 {
		t.Fatalf("unexpected first item group: %+v", firstGroup)
	}
	firstItem := firstGroup.Items[0]
	if firstItem.ItemNo != 1 || firstItem.DomainCode != "FM" {
		t.Fatalf("unexpected first item: %+v", firstItem)
	}
	if len(firstItem.ScoreOptions) != 3 || firstItem.ScoreOptions[0].Value != 2 {
		t.Fatalf("expected 2/1/0 score options: %+v", firstItem.ScoreOptions)
	}
	if !strings.Contains(firstItem.ScoreOptions[0].Description, "旋开瓶盖") {
		t.Fatalf("expected parsed score criterion, got: %+v", firstItem.ScoreOptions[0])
	}
	if len(template.Domains) != 13 {
		t.Fatalf("expected 13 PEP-3 domains, got %d", len(template.Domains))
	}
	pbField := findPEP3RawScoreFieldForTest(template.RawScoreFields, "PB")
	if pbField == nil || pbField.InputMode != "manual_raw_score" || pbField.Category != "caregiver_report" {
		t.Fatalf("expected caregiver raw score field: %+v", pbField)
	}
	if template.SubmitContract.ItemScoreListKey != "itemScoreList" || template.SubmitContract.CreateRecordEndpoint == "" {
		t.Fatalf("unexpected submit contract: %+v", template.SubmitContract)
	}
}

func findPEP3RawScoreFieldForTest(fields []model.PEP3RawScoreField, scaleCode string) *model.PEP3RawScoreField {
	for i := range fields {
		if fields[i].ScaleCode == scaleCode {
			return &fields[i]
		}
	}
	return nil
}
