package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go-migration-platform/pkg/erxinscore"
	"go-migration-platform/services/education/internal/model"
)

func TestBuildERXinAssessmentDraftProgressUsesStoppingRules(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..")
	itemPath := filepath.Join(root, "docs", "erxin-item-bank-draft.json")
	if _, err := os.Stat(itemPath); err != nil {
		t.Skipf("generated ERXin draft data not present: %s", itemPath)
	}

	items, err := erxinscore.LoadItemDefinitionsFile(itemPath)
	if err != nil {
		t.Fatalf("load ERXin items: %v", err)
	}
	itemPasses := make(map[int]bool)
	for _, item := range items {
		switch item.AgeMonth {
		case 15, 18, 21:
			itemPasses[item.ItemNo] = true
		case 24, 27:
			itemPasses[item.ItemNo] = false
		}
	}
	birthDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	assessmentDate := time.Date(2021, 10, 1, 0, 0, 0, 0, time.UTC)

	progress, err := buildERXinAssessmentDraftProgress(&birthDate, &assessmentDate, itemPasses)
	if err != nil {
		t.Fatalf("buildERXinAssessmentDraftProgress returned error: %v", err)
	}
	if !progress.Complete || !progress.CanScore {
		t.Fatalf("expected progress to be score-ready: %+v", progress)
	}
	if progress.AnsweredItemCount != len(itemPasses) || progress.MissingItemCount != 0 {
		t.Fatalf("unexpected item progress: %+v", progress)
	}
	if len(progress.DomainProgress) != len(erxinscore.DomainOrder) {
		t.Fatalf("unexpected domain progress: %+v", progress.DomainProgress)
	}
	for _, domain := range progress.DomainProgress {
		if !domain.Complete {
			t.Fatalf("expected domain %s to be complete: %+v", domain.ScaleCode, domain)
		}
	}
}

func TestBuildERXinReportUsesSavedScore(t *testing.T) {
	raw, err := json.Marshal(ERXinScoreResponse{
		ERXinScoreDataInfo: ERXinScoreDataInfo{
			ScaleCode:      erxinScaleCode,
			ScaleVersion:   erxinScaleVersion,
			SourceStandard: "WS/T 580-2017",
			SourcePDF:      "儿心.pdf",
		},
		Result: erxinscore.AssessmentResult{
			MainAgeMonth:            21,
			MeanMentalAgeMonths:     21,
			MeanMentalAgeMonthsText: "21个月",
			DQ:                      100,
			Level:                   "中等",
			Complete:                true,
			Domains: []erxinscore.DomainResult{
				{
					DomainCode:          "GM",
					DomainName:          "大运动",
					MentalAgeMonths:     21,
					MentalAgeMonthsText: "21个月",
					DQ:                  100,
					Level:               "中等",
					BasalAgeMonth:       18,
					CeilingAgeMonth:     24,
					Complete:            true,
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("marshal score: %v", err)
	}

	report, err := buildERXinReport(model.AssessmentRecordDetailVO{
		AssessmentRecordSummaryVO: model.AssessmentRecordSummaryVO{
			AssessmentCode: erxinScaleCode,
			ScaleVersion:   erxinScaleVersion,
			StudentName:    "测试儿童",
		},
		ResultJSON: raw,
	})
	if err != nil {
		t.Fatalf("buildERXinReport returned error: %v", err)
	}
	if report.TemplateCode != "ERXIN2_DEVELOPMENT_REPORT" || report.Summary.DQ != 100 || !report.Summary.Complete {
		t.Fatalf("unexpected report summary: %+v", report)
	}
	if len(report.DomainRows) != 1 || report.DomainRows[0].DomainCode != "GM" {
		t.Fatalf("unexpected domain rows: %+v", report.DomainRows)
	}
	if len(report.Sections) < 3 {
		t.Fatalf("expected report sections: %+v", report.Sections)
	}
}
