package vbmappscore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadGeneratedResponseSchemas(t *testing.T) {
	dataDir := filepath.Join("..", "..", "docs", "vbmapp")
	for _, name := range []string{
		"milestone-response-schemas.json",
		"barrier-response-schemas.json",
		"transition-response-schemas.json",
		"response-field-templates.json",
		"response-material-profiles.json",
	} {
		if _, err := os.Stat(filepath.Join(dataDir, name)); err != nil {
			t.Skipf("generated VB-MAPP response schema data not present: %s", filepath.Join(dataDir, name))
		}
	}

	milestones, err := LoadMilestoneResponseSchemasFile(filepath.Join(dataDir, "milestone-response-schemas.json"))
	if err != nil {
		t.Fatalf("LoadMilestoneResponseSchemasFile: %v", err)
	}
	if len(milestones) != 170 {
		t.Fatalf("unexpected milestone response schema count: %d", len(milestones))
	}
	barriers, err := LoadBarrierResponseSchemasFile(filepath.Join(dataDir, "barrier-response-schemas.json"))
	if err != nil {
		t.Fatalf("LoadBarrierResponseSchemasFile: %v", err)
	}
	if len(barriers) != 24 {
		t.Fatalf("unexpected barrier response schema count: %d", len(barriers))
	}
	transitions, err := LoadTransitionResponseSchemasFile(filepath.Join(dataDir, "transition-response-schemas.json"))
	if err != nil {
		t.Fatalf("LoadTransitionResponseSchemasFile: %v", err)
	}
	if len(transitions) != 18 {
		t.Fatalf("unexpected transition response schema count: %d", len(transitions))
	}

	templates, err := LoadResponseFieldTemplatesFile(filepath.Join(dataDir, "response-field-templates.json"))
	if err != nil {
		t.Fatalf("LoadResponseFieldTemplatesFile: %v", err)
	}
	if _, ok := templates["mand_event_log"]; !ok {
		t.Fatal("expected mand_event_log field template")
	}
	profiles, err := LoadResponseMaterialProfilesFile(filepath.Join(dataDir, "response-material-profiles.json"))
	if err != nil {
		t.Fatalf("LoadResponseMaterialProfilesFile: %v", err)
	}
	if _, ok := profiles["potential_reinforcer_set"]; !ok {
		t.Fatal("expected potential_reinforcer_set material profile")
	}
	if _, ok := profiles["mand_1m_request_starter_set"]; !ok {
		t.Fatal("expected mand_1m_request_starter_set material profile")
	}
	if _, ok := profiles["mand_2m_visible_request_set"]; !ok {
		t.Fatal("expected mand_2m_visible_request_set material profile")
	}
	if !containsMaterialSuggestion(profiles["mand_1m_request_starter_set"].RecommendedMaterials, "饼干") {
		t.Fatalf("expected MAND 1M material quick picks: %+v", profiles["mand_1m_request_starter_set"].RecommendedMaterials)
	}
	if !containsMaterialSuggestion(profiles["mand_2m_visible_request_set"].RecommendedMaterials, "彩虹弹簧") {
		t.Fatalf("expected MAND 2M material quick picks: %+v", profiles["mand_2m_visible_request_set"].RecommendedMaterials)
	}
	if !contains(profiles["potential_reinforcer_set"].QuickPicksByField["mand3_people"], "爸爸") {
		t.Fatalf("expected MAND 3M people quick picks: %+v", profiles["potential_reinforcer_set"].QuickPicksByField)
	}
	if !contains(profiles["potential_reinforcer_set"].QuickPicksByField["mand3_settings"], "屋里") {
		t.Fatalf("expected MAND 3M setting quick picks: %+v", profiles["potential_reinforcer_set"].QuickPicksByField)
	}
	if !contains(profiles["potential_reinforcer_set"].QuickPicksByField["mand3_examples_bubbles"], "红瓶泡泡") {
		t.Fatalf("expected MAND 3M example quick picks: %+v", profiles["potential_reinforcer_set"].QuickPicksByField)
	}

	mand1 := findMilestoneResponseSchema(t, milestones, "MAND_01M")
	if mand1.UIPattern != "mand_event_recorder" {
		t.Fatalf("unexpected MAND_01M ui pattern: %s", mand1.UIPattern)
	}
	if mand1.MaterialProfileID != "mand_1m_request_starter_set" {
		t.Fatalf("unexpected MAND_01M material profile: %s", mand1.MaterialProfileID)
	}
	if !contains(mand1.FieldTemplateIDs, "mand_event_log") {
		t.Fatalf("MAND_01M should require mand_event_log: %+v", mand1.FieldTemplateIDs)
	}
	if mand1.AutoCompletion.ScoreStrategy != "count_qualified_unique_mand_events" {
		t.Fatalf("unexpected MAND_01M auto strategy: %s", mand1.AutoCompletion.ScoreStrategy)
	}

	mand2 := findMilestoneResponseSchema(t, milestones, "MAND_02M")
	if mand2.MaterialProfileID != "mand_2m_visible_request_set" {
		t.Fatalf("unexpected MAND_02M material profile: %s", mand2.MaterialProfileID)
	}

	mand3 := findMilestoneResponseSchema(t, milestones, "MAND_03M")
	if !contains(mand3.FieldTemplateIDs, "generalization_matrix") {
		t.Fatalf("MAND_03M should require generalization matrix: %+v", mand3.FieldTemplateIDs)
	}
	if !contains(mand3.AutoCompletion.ComputedIndicators, "person_generalization_count") {
		t.Fatalf("MAND_03M should compute generalization indicators: %+v", mand3.AutoCompletion.ComputedIndicators)
	}

	if barriers[0].UIPattern != "barrier_rubric_with_behavior_log" {
		t.Fatalf("unexpected barrier ui pattern: %s", barriers[0].UIPattern)
	}
	if !transitions[0].AutoCompletion.CanSuggestScore {
		t.Fatal("T01 should support score suggestion from computed totals")
	}
}

func findMilestoneResponseSchema(t *testing.T, schemas []MilestoneResponseSchema, id string) MilestoneResponseSchema {
	t.Helper()
	for _, schema := range schemas {
		if schema.MilestoneID == id {
			return schema
		}
	}
	t.Fatalf("milestone response schema %s not found", id)
	return MilestoneResponseSchema{}
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func containsMaterialSuggestion(values []ResponseMaterialSuggestion, value string) bool {
	for _, candidate := range values {
		if candidate.Name == value {
			return true
		}
	}
	return false
}
