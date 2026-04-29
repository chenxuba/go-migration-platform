package service

import (
	"encoding/json"
	"testing"
	"time"

	"go-migration-platform/pkg/pep3score"
	"go-migration-platform/services/education/internal/model"
)

func TestBuildPEP3ReportFromSavedScore(t *testing.T) {
	scaledScore := 9
	percentile := 35
	maxRaw := 68
	standardSum := 32
	devAge := 20.0
	score := PEP3ScoreResponse{
		PEP3ScoreDataInfo: PEP3ScoreDataInfo{
			ScaleCode:    "PEP3",
			ScaleVersion: "2025-draft",
			DataStatus:   "常模OCR草稿，需复核",
			Sources:      []string{"pep3-norm-conversion-ocr-draft.json"},
		},
		Result: pep3score.AssessmentResult{
			Age: pep3score.Age{Years: 3, Months: 3, Days: 11, TotalMonthsForNorm: 39},
			Scales: map[string]pep3score.ScaleResult{
				"CVP": {
					ScaleCode:   "CVP",
					ScaleName:   "认知（语言/语前）",
					RawScore:    16,
					MaxRawScore: &maxRaw,
					DevelopmentAge: &pep3score.NormValue{
						Text:   "18",
						Number: intPtrForPEP3ReportTest(18),
					},
					PercentileRank: &pep3score.NormValue{Text: "35", Number: &percentile},
					ScaledScore:    &pep3score.NormValue{Text: "9", Number: &scaledScore},
					Level:          "中度",
				},
			},
			Composites: map[string]pep3score.CompositeResult{
				pep3score.CompositeCommunication: {
					CompositeCode:        pep3score.CompositeCommunication,
					CompositeName:        "沟通",
					MemberScaleCodes:     []string{"CVP", "EL", "RL"},
					StandardScoreSum:     &standardSum,
					PercentileRank:       &pep3score.NormValue{Text: "50", Number: intPtrForPEP3ReportTest(50)},
					DevelopmentAgeMonths: &devAge,
					Level:                "中度",
				},
			},
		},
	}
	raw, err := json.Marshal(score)
	if err != nil {
		t.Fatalf("marshal score: %v", err)
	}
	birthDate := time.Date(2000, 10, 29, 0, 0, 0, 0, time.Local)
	assessmentDate := time.Date(2004, 2, 10, 0, 0, 0, 0, time.Local)

	report, err := buildPEP3Report(model.AssessmentRecordDetailVO{
		AssessmentRecordSummaryVO: model.AssessmentRecordSummaryVO{
			ID:             1,
			StudentID:      1001,
			StudentName:    "李东尼",
			AssessmentCode: "PEP3",
			AssessmentName: "PEP-3儿童心理教育评核",
			ScaleVersion:   "2025-draft",
			BirthDate:      &birthDate,
			AssessmentDate: &assessmentDate,
			AgeYears:       3,
			AgeMonths:      3,
			AgeDays:        11,
			NormAgeMonths:  39,
		},
		ResultJSON: raw,
	})
	if err != nil {
		t.Fatalf("buildPEP3Report returned error: %v", err)
	}
	if report.BasicInfo.AgeText != "3岁3个月11天" {
		t.Fatalf("unexpected age text: %s", report.BasicInfo.AgeText)
	}
	if len(report.DevelopmentRows) != 1 || report.DevelopmentRows[0].ScaleCode != "CVP" || report.DevelopmentRows[0].ScaledScoreText != "9" {
		t.Fatalf("unexpected development rows: %+v", report.DevelopmentRows)
	}
	if len(report.CompositeRows) != 1 || report.CompositeRows[0].StandardScoreSumText != "32" || report.CompositeRows[0].DevelopmentAgeText != "20个月" {
		t.Fatalf("unexpected composite rows: %+v", report.CompositeRows)
	}
	if len(report.Warnings) == 0 {
		t.Fatal("expected data status warning")
	}
}

func intPtrForPEP3ReportTest(value int) *int {
	return &value
}
