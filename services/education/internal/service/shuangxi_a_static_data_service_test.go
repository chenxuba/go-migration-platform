package service

import (
	"context"
	"testing"
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
