package service

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"go-migration-platform/pkg/pep3score"
	"go-migration-platform/services/education/internal/model"
)

func TestBuildPEP3BookletFillsItemGrid(t *testing.T) {
	score := PEP3ScoreResponse{
		PEP3ScoreDataInfo: PEP3ScoreDataInfo{
			ScaleCode:    "PEP3",
			ScaleVersion: "2025-draft",
			Sources:      []string{"pep3-item-bank-simplified-draft.json"},
		},
		Result: pep3score.AssessmentResult{
			Age: pep3score.Age{Years: 3, Months: 3, Days: 11, TotalMonthsForNorm: 39},
			Scales: map[string]pep3score.ScaleResult{
				"FM": {
					ScaleCode: "FM",
					ScaleName: "小肌肉",
					RawScore:  2,
				},
			},
			Composites: map[string]pep3score.CompositeResult{},
		},
	}
	resultRaw, err := json.Marshal(score)
	if err != nil {
		t.Fatalf("marshal score: %v", err)
	}
	inputRaw, err := json.Marshal(map[string]any{
		"itemScoreList": []map[string]int{{"itemNo": 1, "score": 2}},
		"rawScores":     map[string]int{"FM": 2},
	})
	if err != nil {
		t.Fatalf("marshal input: %v", err)
	}
	birthDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local)
	assessmentDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)

	booklet, err := buildPEP3Booklet(model.AssessmentRecordDetailVO{
		AssessmentRecordSummaryVO: model.AssessmentRecordSummaryVO{
			ID:             1,
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
			ExaminerName:   "测试员A",
		},
		InputJSON:  inputRaw,
		ResultJSON: resultRaw,
	})
	if err != nil {
		t.Fatalf("buildPEP3Booklet returned error: %v", err)
	}
	if booklet.TemplateCode != "PEP3_RECORD_BOOKLET" || len(booklet.Pages) != 14 {
		t.Fatalf("unexpected booklet metadata: %+v", booklet)
	}
	page2 := booklet.Pages[1]
	if len(page2.Sections) == 0 || page2.Sections[0].Type != "item_grid" {
		t.Fatalf("expected page 2 item grid section: %+v", page2.Sections)
	}
	rows := page2.Sections[0].Table.Rows
	if len(rows) == 0 || rows[0]["itemNo"] != 1 || rows[0]["score"] != 2 || rows[0]["FM"] != 2 {
		t.Fatalf("expected item 1 score to fill FM cell: %+v", rows[:1])
	}
}

func TestBuildPEP3BookletPDFUsesTemplateBackground(t *testing.T) {
	if _, err := resolvePEP3PDFFontPath(); err != nil {
		t.Skipf("PEP-3 PDF font not available: %v", err)
	}
	resultRaw := pep3BookletTestScoreRaw(t)
	birthDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local)
	assessmentDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)

	content, err := buildPEP3BookletPDF(model.AssessmentRecordDetailVO{
		AssessmentRecordSummaryVO: model.AssessmentRecordSummaryVO{
			ID:             1,
			StudentName:    "李东尼",
			AssessmentCode: "PEP3",
			AssessmentName: "PEP-3儿童心理教育评核",
			ScaleVersion:   "2025-draft",
			BirthDate:      &birthDate,
			AssessmentDate: &assessmentDate,
			AgeYears:       4,
			AgeMonths:      0,
			AgeDays:        0,
			NormAgeMonths:  48,
			ExaminerName:   "测试员A",
		},
		ResultJSON: resultRaw,
	})
	if err != nil {
		t.Fatalf("buildPEP3BookletPDF returned error: %v", err)
	}
	if output := os.Getenv("PEP3_BOOKLET_PDF_TEST_OUTPUT"); output != "" {
		if err := os.WriteFile(output, content, 0o644); err != nil {
			t.Fatalf("write sample booklet PDF: %v", err)
		}
	}
	header := ""
	if len(content) >= 4 {
		header = string(content[:4])
	}
	if len(content) < 1_000_000 || header != "%PDF" {
		t.Fatalf("expected non-empty PDF bytes, len=%d header=%q", len(content), header)
	}
}

func pep3BookletTestScoreRaw(t *testing.T) []byte {
	t.Helper()
	score := PEP3ScoreResponse{
		PEP3ScoreDataInfo: PEP3ScoreDataInfo{
			ScaleCode:    "PEP3",
			ScaleVersion: "2025-draft",
			Sources:      []string{"pep3-item-bank-simplified-draft.json"},
		},
		Result: pep3score.AssessmentResult{
			Age: pep3score.Age{Years: 4, Months: 0, Days: 0, TotalMonthsForNorm: 48},
			Scales: map[string]pep3score.ScaleResult{
				"FM": {
					ScaleCode: "FM",
					ScaleName: "小肌肉",
					RawScore:  2,
				},
			},
			Composites: map[string]pep3score.CompositeResult{},
		},
	}
	raw, err := json.Marshal(score)
	if err != nil {
		t.Fatalf("marshal score: %v", err)
	}
	return raw
}
