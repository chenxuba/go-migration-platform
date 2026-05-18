package service

import (
	"context"
	"testing"
	"time"
)

func TestLoadShuangxiAStaticDataFromFiles(t *testing.T) {
	dataDir, err := resolveShuangxiADataDir()
	if err != nil {
		t.Fatalf("resolveShuangxiADataDir returned error: %v", err)
	}
	data, err := loadShuangxiAStaticDataFromFiles(dataDir)
	if err != nil {
		t.Fatalf("loadShuangxiAStaticDataFromFiles returned error: %v", err)
	}
	if data.metadata.ScaleCode != shuangxiAScaleCode || data.metadata.ScaleVersion != shuangxiAScaleVersion {
		t.Fatalf("unexpected metadata: %+v", data.metadata)
	}
	if len(data.items) != 209 {
		t.Fatalf("expected 209 items, got %d", len(data.items))
	}
	if len(data.domains) != 7 {
		t.Fatalf("expected 7 domains, got %d", len(data.domains))
	}
	if data.items[0].ItemCode != "1.1.1" || data.items[0].ScoreMax != 3 {
		t.Fatalf("unexpected first item: %+v", data.items[0])
	}
	last := data.items[len(data.items)-1]
	if last.ItemCode != "7.8.4" || last.DomainCode != "SOCIAL_SKILLS" {
		t.Fatalf("unexpected last item: %+v", last)
	}
	wantDomainCounts := map[string]int{
		"SENSORY":       21,
		"GROSS_MOTOR":   25,
		"FINE_MOTOR":    14,
		"SELF_CARE":     24,
		"COMMUNICATION": 56,
		"COGNITION":     21,
		"SOCIAL_SKILLS": 48,
	}
	for _, domain := range data.domains {
		if domain.ItemCount != wantDomainCounts[domain.ScaleCode] {
			t.Fatalf("domain %s item count = %d, want %d", domain.ScaleCode, domain.ItemCount, wantDomainCounts[domain.ScaleCode])
		}
		if domain.MaxRawScore != domain.ItemCount*3 {
			t.Fatalf("domain %s max raw score = %d, want %d", domain.ScaleCode, domain.MaxRawScore, domain.ItemCount*3)
		}
	}
}

func TestBuildShuangxiAAssessmentFormTemplateSummary(t *testing.T) {
	dataDir, err := resolveShuangxiADataDir()
	if err != nil {
		t.Fatalf("resolveShuangxiADataDir returned error: %v", err)
	}
	data, err := loadShuangxiAStaticDataFromFiles(dataDir)
	if err != nil {
		t.Fatalf("load Shuangxi A data: %v", err)
	}
	summary := buildShuangxiAAssessmentFormTemplateSummary(data)
	if summary.ItemCount != 209 || len(summary.Domains) != 7 {
		t.Fatalf("unexpected summary counts: %+v", summary)
	}
	if len(summary.Domains[0].Skills) != 5 {
		t.Fatalf("first domain skills = %d, want 5", len(summary.Domains[0].Skills))
	}
	firstSkill := summary.Domains[0].Skills[0]
	if firstSkill.SkillCode != "1.1" || len(firstSkill.Items) != 8 {
		t.Fatalf("unexpected first skill: %+v", firstSkill)
	}
	firstItem, err := (&Service{}).GetShuangxiAAssessmentFormTemplateItem(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetShuangxiAAssessmentFormTemplateItem returned error: %v", err)
	}
	if firstItem.ItemCode != "1.1.1" || len(firstItem.ScoreOptions) != 4 {
		t.Fatalf("unexpected first item: %+v", firstItem)
	}
}

func TestFillShuangxiAMissingItemScoresWithZeroCompletesAllItems(t *testing.T) {
	dataDir, err := resolveShuangxiADataDir()
	if err != nil {
		t.Fatalf("resolveShuangxiADataDir returned error: %v", err)
	}
	data, err := loadShuangxiAStaticDataFromFiles(dataDir)
	if err != nil {
		t.Fatalf("load Shuangxi A data: %v", err)
	}

	filled := fillShuangxiAMissingItemScoresWithZero(data, map[int]int{1: 2})
	if len(filled) != 209 {
		t.Fatalf("filled item score count = %d, want 209", len(filled))
	}
	if filled[1] != 2 {
		t.Fatalf("item 1 score = %d, want preserved score 2", filled[1])
	}
	if filled[2] != 0 || filled[209] != 0 {
		t.Fatalf("missing item scores were not defaulted to zero: item2=%d item209=%d", filled[2], filled[209])
	}

	birthDate := time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
	assessmentDate := time.Date(2026, 5, 18, 0, 0, 0, 0, time.UTC)
	progress, err := buildShuangxiAAssessmentDraftProgressWithData(data, &birthDate, &assessmentDate, filled)
	if err != nil {
		t.Fatalf("build progress returned error: %v", err)
	}
	if !progress.Complete || progress.MissingItemCount != 0 || progress.AnsweredItemCount != 209 {
		t.Fatalf("unexpected completed progress: %+v", progress)
	}
}

func TestApplyShuangxiAGenderDefaults(t *testing.T) {
	maleScores := applyShuangxiAGenderDefaults(map[int]int{
		shuangxiAUseSanitaryPadItemNo: 0,
		shuangxiAShaveItemNo:          1,
	}, "男")
	if maleScores[shuangxiAUseSanitaryPadItemNo] != 3 {
		t.Fatalf("male sanitary pad score = %d, want 3", maleScores[shuangxiAUseSanitaryPadItemNo])
	}
	if maleScores[shuangxiAShaveItemNo] != 1 {
		t.Fatalf("male shave score = %d, want preserved score 1", maleScores[shuangxiAShaveItemNo])
	}

	femaleScores := applyShuangxiAGenderDefaults(map[int]int{
		shuangxiAUseSanitaryPadItemNo: 2,
		shuangxiAShaveItemNo:          0,
	}, "female")
	if femaleScores[shuangxiAUseSanitaryPadItemNo] != 2 {
		t.Fatalf("female sanitary pad score = %d, want preserved score 2", femaleScores[shuangxiAUseSanitaryPadItemNo])
	}
	if femaleScores[shuangxiAShaveItemNo] != 3 {
		t.Fatalf("female shave score = %d, want 3", femaleScores[shuangxiAShaveItemNo])
	}

	unknownScores := applyShuangxiAGenderDefaults(map[int]int{
		shuangxiAUseSanitaryPadItemNo: 0,
		shuangxiAShaveItemNo:          0,
	}, "未填")
	if unknownScores[shuangxiAUseSanitaryPadItemNo] != 0 || unknownScores[shuangxiAShaveItemNo] != 0 {
		t.Fatalf("unknown gender scores were changed: %+v", unknownScores)
	}
}

func TestScaleAssessmentStudentGenderValue(t *testing.T) {
	cases := []struct {
		raw        string
		wantSex    int
		wantGender string
	}{
		{raw: "男", wantSex: 1, wantGender: "男"},
		{raw: "male", wantSex: 1, wantGender: "男"},
		{raw: "女", wantSex: 0, wantGender: "女"},
		{raw: "female", wantSex: 0, wantGender: "女"},
	}
	for _, tc := range cases {
		sex, gender, err := scaleAssessmentStudentGenderValue(tc.raw)
		if err != nil {
			t.Fatalf("scaleAssessmentStudentGenderValue(%q) returned error: %v", tc.raw, err)
		}
		if sex != tc.wantSex || gender != tc.wantGender {
			t.Fatalf("scaleAssessmentStudentGenderValue(%q) = (%d, %q), want (%d, %q)", tc.raw, sex, gender, tc.wantSex, tc.wantGender)
		}
	}

	if _, _, err := scaleAssessmentStudentGenderValue("-"); err == nil {
		t.Fatalf("scaleAssessmentStudentGenderValue(-) expected error")
	}
}
