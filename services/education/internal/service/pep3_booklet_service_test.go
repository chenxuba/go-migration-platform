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
			ScaleVersion: "2025-92mo-draft",
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
		"itemRecordValues": map[string]map[string]any{
			"112": {
				"repeated_two_digits": []string{"7-9", "5-3"},
			},
			"116": {
				"eye_contact": "brief",
			},
		},
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
			ScaleVersion:   "2025-92mo-draft",
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
	if booklet.TemplateCode != "PEP3_RECORD_BOOKLET" || len(booklet.Pages) != 16 {
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
	row112 := findPEP3BookletRowForTest(booklet, 112)
	if row112 == nil {
		t.Fatalf("expected item 112 row in booklet")
	}
	recordValues, ok := row112["recordValues"].(map[string]any)
	repeatedDigits, _ := recordValues["repeated_two_digits"].([]any)
	if !ok || len(repeatedDigits) != 2 {
		t.Fatalf("expected item 112 record values, got: %+v", row112["recordValues"])
	}
	recordFields, ok := row112["recordFields"].([]model.PEP3ItemRecordField)
	if !ok || len(recordFields) != 1 || recordFields[0].Key != "repeated_two_digits" || recordFields[0].FieldType != "checkbox_group" {
		t.Fatalf("expected item 112 record fields, got: %+v", row112["recordFields"])
	}
}

func findPEP3BookletRowForTest(booklet model.PEP3BookletVO, itemNo int) map[string]any {
	for _, page := range booklet.Pages {
		for _, section := range page.Sections {
			if section.Type != "item_grid" || section.Table == nil {
				continue
			}
			for _, row := range section.Table.Rows {
				if row["itemNo"] == itemNo {
					return row
				}
			}
		}
	}
	return nil
}

func TestBuildPEP3BookletPDFUsesTemplateBackground(t *testing.T) {
	if _, err := resolvePEP3PDFFontPath(); err != nil {
		t.Skipf("PEP-3 PDF font not available: %v", err)
	}
	resultRaw := pep3BookletTestScoreRaw(t)
	itemScoreList := make([]map[string]int, 0, 172)
	for itemNo := 1; itemNo <= 172; itemNo++ {
		itemScoreList = append(itemScoreList, map[string]int{
			"itemNo": itemNo,
			"score":  itemNo % 3,
		})
	}
	inputRaw, err := json.Marshal(map[string]any{
		"itemScoreList": itemScoreList,
		"itemRecordValues": map[string]map[string]any{
			"5":   {"touch_block_reaction": "no_interest"},
			"6":   {"kaleidoscope_action": []string{"watch", "watch_and_turn"}},
			"7":   {"first_observation": "right_eye", "second_observation": "left_eye"},
			"9":   {"bell_attempts": []string{"first_attempt", "second_attempt"}},
			"29":  {"size_naming": []string{"first_big", "second_small"}},
			"31":  {"cat_puzzle_prompt": "需示范", "completed_piece_count": "3"},
			"33":  {"cow_puzzle_prompt": "自行", "completed_piece_count": "4"},
			"40":  {"pointed_objects": "气球、板凳、手鼓、杯子、剪刀"},
			"85":  {"picture_identification": []string{"A 杯子", "H 系鞋带", "T 火箭起飞"}},
			"86":  {"picture_naming": []string{"A 牛", "J 电风扇", "T 整路"}},
			"95":  {"reading_comprehension_questions": []string{"小明有哪些动物呀？", "什么跳过小明的皮球？"}},
			"108": {"classification_prompt": "部分示范", "classification_basis": "颜色", "completed_card_count": "8"},
			"112": {"repeated_two_digits": []string{"7-9", "5-3"}},
			"113": {"repeated_three_digits": []string{"2-4-1", "5-7-9"}},
			"114": {"repeated_words": []string{"街街", "车车", "拜拜"}},
			"115": {"repeated_sentences": []string{"bb_looking", "want_biscuit"}},
			"116": {"eye_contact": "brief"},
			"117": {"delayed_echolalia": "not_applicable"},
			"119": {"pronoun_response": "玩具"},
			"125": {"gesture_responses": []string{"叫名字+招手", "其他"}},
			"134": {"simple_action_commands": []string{"坐下", "其他"}},
		},
	})
	if err != nil {
		t.Fatalf("marshal input: %v", err)
	}
	birthDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local)
	assessmentDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)

	content, err := buildPEP3BookletPDF(model.AssessmentRecordDetailVO{
		AssessmentRecordSummaryVO: model.AssessmentRecordSummaryVO{
			ID:             1,
			StudentName:    "李东尼",
			AssessmentCode: "PEP3",
			AssessmentName: "PEP-3儿童心理教育评核",
			ScaleVersion:   "2025-92mo-draft",
			BirthDate:      &birthDate,
			AssessmentDate: &assessmentDate,
			AgeYears:       4,
			AgeMonths:      0,
			AgeDays:        0,
			NormAgeMonths:  48,
			ExaminerName:   "测试员A",
		},
		InputJSON:  inputRaw,
		ResultJSON: resultRaw,
	}, "测试机构")
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
			ScaleVersion: "2025-92mo-draft",
			Sources:      []string{"pep3-item-bank-simplified-draft.json"},
		},
		Result: pep3score.AssessmentResult{
			Age: pep3score.Age{Years: 4, Months: 0, Days: 0, TotalMonthsForNorm: 48},
			Scales: map[string]pep3score.ScaleResult{
				"CVP": {
					ScaleCode:      "CVP",
					ScaleName:      "认知（语言/语前）",
					RawScore:       58,
					DevelopmentAge: pep3BookletTestNormValue("49", 49),
				},
				"EL": {
					ScaleCode:      "EL",
					ScaleName:      "语言表达",
					RawScore:       35,
					DevelopmentAge: pep3BookletTestNormValue("48", 48),
				},
				"RL": {
					ScaleCode:      "RL",
					ScaleName:      "语言理解",
					RawScore:       29,
					DevelopmentAge: pep3BookletTestNormValue("32", 32),
				},
				"FM": {
					ScaleCode:      "FM",
					ScaleName:      "小肌肉",
					RawScore:       38,
					DevelopmentAge: pep3BookletTestNormValue("48", 48),
				},
				"GM": {
					ScaleCode:      "GM",
					ScaleName:      "大肌肉",
					RawScore:       28,
					DevelopmentAge: pep3BookletTestNormValue("31", 31),
				},
				"VMI": {
					ScaleCode:      "VMI",
					ScaleName:      "模仿（视觉/动作）",
					RawScore:       20,
					DevelopmentAge: pep3BookletTestNormValue("48", 48),
				},
				"PSC": {
					ScaleCode:      "PSC",
					ScaleName:      "个人自理",
					RawScore:       21,
					DevelopmentAge: pep3BookletTestNormValue("48", 48),
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

func pep3BookletTestNormValue(text string, number int) *pep3score.NormValue {
	return &pep3score.NormValue{Text: text, Number: &number}
}
