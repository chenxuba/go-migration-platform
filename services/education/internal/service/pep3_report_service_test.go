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
				"AE": {
					ScaleCode:      "AE",
					ScaleName:      "情感表达",
					RawScore:       17,
					MaxRawScore:    intPtrForPEP3ReportTest(22),
					PercentileRank: &pep3score.NormValue{Text: "65", Number: intPtrForPEP3ReportTest(65)},
					ScaledScore:    &pep3score.NormValue{Text: "12", Number: intPtrForPEP3ReportTest(12)},
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
	if report.TemplateCode != "PEP3_EXPLANATORY_REPORT" || report.Title == "" {
		t.Fatalf("unexpected template metadata: %+v", report)
	}
	if len(report.Sections) == 0 || report.Sections[0].SectionCode != "basic_info" {
		t.Fatalf("expected frontend-fillable report sections: %+v", report.Sections)
	}
	basicFields := report.Sections[0].Fields
	if len(basicFields) < 7 || basicFields[6].Key != "ageText" || basicFields[6].Value != "3岁3个月11天" {
		t.Fatalf("unexpected basic info fields: %+v", basicFields)
	}
	developmentSection := findPEP3ReportTestSection(report.Sections, "development_scores")
	if developmentSection == nil || developmentSection.Table == nil || len(developmentSection.Table.Rows) != 1 {
		t.Fatalf("expected score table rows in report sections: %+v", report.Sections[1])
	}
	if developmentSection.Table.Rows[0]["scaleCode"] != "CVP" || developmentSection.Table.Rows[0]["scaledScore"] != "9" || developmentSection.Table.Rows[0]["developmentAge"] != "18个月" {
		t.Fatalf("unexpected development score row: %+v", developmentSection.Table.Rows[0])
	}
	behaviorSection := findPEP3ReportTestSection(report.Sections, "behavior_scores")
	if behaviorSection == nil || behaviorSection.Table == nil || len(behaviorSection.Table.Rows) != 1 {
		t.Fatalf("expected behavior section row: %+v", behaviorSection)
	}
	if behaviorSection.Table.Rows[0]["scaleCode"] != "AE" || behaviorSection.Table.Rows[0]["developmentAge"] != "--" {
		t.Fatalf("behavior subtests should keep the development age column with --: %+v", behaviorSection.Table.Rows[0])
	}
	compositeSection := findPEP3ReportTestSection(report.Sections, "composite_scores")
	if compositeSection == nil || compositeSection.Table == nil || len(compositeSection.Table.Rows) != 1 {
		t.Fatalf("expected composite section row: %+v", compositeSection)
	}
	if len(compositeSection.Table.Columns) < 12 || compositeSection.Table.Columns[1].Key != "CVP" || compositeSection.Table.Columns[1].Group == "" {
		t.Fatalf("expected composite table to expose member standard score columns: %+v", compositeSection.Table.Columns)
	}
	compositeRow := compositeSection.Table.Rows[0]
	if compositeRow["CVP"] != "9" || compositeRow["EL"] != "待校对" || compositeRow["FM"] != "--" {
		t.Fatalf("unexpected composite member scores: %+v", compositeRow)
	}
	if compositeRow["standardScoreSum"] != "32" || compositeRow["developmentAge"] != "20个月" {
		t.Fatalf("unexpected composite row: %+v", compositeRow)
	}
	warningSection := findPEP3ReportTestSection(report.Sections, "warnings")
	if warningSection == nil || len(warningSection.TextItems) == 0 {
		t.Fatal("expected data status warning section")
	}
	reportRaw, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal report: %v", err)
	}
	var reportObject map[string]json.RawMessage
	if err := json.Unmarshal(reportRaw, &reportObject); err != nil {
		t.Fatalf("unmarshal report object: %v", err)
	}
	for _, legacyField := range []string{`"basicInfo"`, `"developmentRows"`, `"behaviorRows"`, `"caregiverReportRows"`, `"compositeRows"`, `"summary"`, `"warnings"`} {
		if _, ok := reportObject[legacyField[1:len(legacyField)-1]]; ok {
			t.Fatalf("report should not expose legacy top-level field %s: %s", legacyField, string(reportRaw))
		}
	}
}

func TestRescorePEP3AssessmentRecordDetailUsesCurrentNormCorrections(t *testing.T) {
	birthDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local)
	assessmentDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.Local)
	inputRaw := json.RawMessage(`{
		"birthDate": "2020-01-01",
		"assessmentDate": "2023-01-01",
		"allowMissingItems": true,
		"rawScoreList": [
			{"scaleCode":"CVP","rawScore":52},
			{"scaleCode":"EL","rawScore":38},
			{"scaleCode":"RL","rawScore":25},
			{"scaleCode":"FM","rawScore":34},
			{"scaleCode":"GM","rawScore":24},
			{"scaleCode":"VMI","rawScore":12},
			{"scaleCode":"AE","rawScore":17},
			{"scaleCode":"SR","rawScore":18},
			{"scaleCode":"CMB","rawScore":24},
			{"scaleCode":"CVB","rawScore":16},
			{"scaleCode":"PB","rawScore":7},
			{"scaleCode":"PSC","rawScore":7},
			{"scaleCode":"AB","rawScore":10}
		]
	}`)

	refreshed := (&Service{}).rescorePEP3AssessmentRecordDetail(model.AssessmentRecordDetailVO{
		AssessmentRecordSummaryVO: model.AssessmentRecordSummaryVO{
			BirthDate:      &birthDate,
			AssessmentDate: &assessmentDate,
		},
		InputJSON:  inputRaw,
		ResultJSON: json.RawMessage(`{"scaleCode":"PEP3","scaleVersion":"old","result":{"scales":{}}}`),
	})
	score, err := decodeSavedPEP3Score(refreshed.ResultJSON)
	if err != nil {
		t.Fatalf("decode refreshed score: %v", err)
	}
	if score.Result.Scales["CVP"].PercentileRank == nil || score.Result.Scales["CVP"].PercentileRank.Text != "88" {
		t.Fatalf("expected refreshed CVP percentile from manual corrections, got: %+v", score.Result.Scales["CVP"])
	}
	if score.Result.Scales["EL"].PercentileRank == nil || score.Result.Scales["EL"].PercentileRank.Text != "92" {
		t.Fatalf("expected refreshed EL percentile from manual corrections, got: %+v", score.Result.Scales["EL"])
	}
	if score.Result.Scales["GM"].ScaledScore == nil || score.Result.Scales["GM"].ScaledScore.Text != "11" {
		t.Fatalf("expected refreshed GM scaled score from manual corrections, got: %+v", score.Result.Scales["GM"])
	}
}

func intPtrForPEP3ReportTest(value int) *int {
	return &value
}

func findPEP3ReportTestSection(sections []model.PEP3TemplateSection, code string) *model.PEP3TemplateSection {
	for i := range sections {
		if sections[i].SectionCode == code {
			return &sections[i]
		}
	}
	return nil
}
