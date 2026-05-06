package service

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"time"

	"go-migration-platform/pkg/pep3score"
	"go-migration-platform/services/education/internal/model"
)

func TestBuildPEP3BookletFillsItemGrid(t *testing.T) {
	score := PEP3ScoreResponse{
		PEP3ScoreDataInfo: PEP3ScoreDataInfo{
			ScaleCode:    "PEP3",
			ScaleVersion: "2025-92题版",
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
			ScaleVersion:   "2025-92题版",
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
	if booklet.TemplateCode != "PEP3_RECORD_BOOKLET" || len(booklet.Pages) != 26 {
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
	page23 := booklet.Pages[22]
	domainSection := findPEP3BookletSectionForTest(page23.Sections, "domain_fm")
	if domainSection == nil || domainSection.Table == nil {
		t.Fatalf("expected FM education planning section on page 23: %+v", page23.Sections)
	}
	if len(domainSection.Table.FooterRows) != 2 ||
		domainSection.Table.FooterRows[0]["score2"] != 2 ||
		domainSection.Table.FooterRows[0]["score1"] != 0 ||
		domainSection.Table.FooterRows[1]["rawScore"] != 2 {
		t.Fatalf("expected FM education planning footer to summarize scores: %+v", domainSection.Table.FooterRows)
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

func findPEP3BookletSectionForTest(sections []model.PEP3TemplateSection, sectionCode string) *model.PEP3TemplateSection {
	for index := range sections {
		if sections[index].SectionCode == sectionCode {
			return &sections[index]
		}
	}
	return nil
}

func TestPEP3BookletPDFEducationPlanningLayoutsCoverDomains(t *testing.T) {
	layouts := pep3BookletPDFEducationPlanningLayouts()
	got := map[string]bool{}
	for _, layout := range layouts {
		got[layout.DomainCode] = true
		if layout.PageNo < 20 || layout.PageNo > 26 {
			t.Fatalf("unexpected education planning page for %s: %d", layout.DomainCode, layout.PageNo)
		}
		if layout.RowBottomY <= layout.RowTopY || layout.FooterTotalY <= layout.FooterScoreY {
			t.Fatalf("invalid education planning y coordinates for %s: %+v", layout.DomainCode, layout)
		}
		for _, scoreValue := range []int{2, 1, 0} {
			if layout.ScoreCenterXByValue[scoreValue] <= 0 {
				t.Fatalf("missing score column %d for %s: %+v", scoreValue, layout.DomainCode, layout.ScoreCenterXByValue)
			}
		}
	}
	for _, domainCode := range pep3BookletDomainOrder() {
		if !got[domainCode] {
			t.Fatalf("missing education planning layout for %s", domainCode)
		}
	}
}

func TestPEP3BookletPDFCaregiverScoreHelpers(t *testing.T) {
	submission := pep3BookletTestCaregiverReportSubmission(t)
	layouts := pep3BookletPDFCaregiverScoreLayouts()
	if len(layouts) != 3 {
		t.Fatalf("expected three caregiver score layouts, got: %+v", layouts)
	}
	for _, layout := range layouts {
		if layout.CenterX <= 0 || layout.RowStartY <= 0 || layout.RowGap <= 0 || layout.RawScoreY <= 0 || layout.ItemCount <= 0 {
			t.Fatalf("invalid caregiver score layout: %+v", layout)
		}
	}

	rawScore, ok := pep3CaregiverReportRawScore(&submission, nil, nil, "PB")
	if !ok || rawScore != 17 {
		t.Fatalf("expected PB raw score from submission, got %d ok=%v", rawScore, ok)
	}
	sections := pep3CaregiverReportScoredSectionMap()
	problemSection := sections["problem_behavior"]
	score, ok := pep3CaregiverReportItemScore(problemSection.Items[1], submission.Answers["problem_behavior"]["speech_delay_or_absent"])
	if !ok || score != 1 {
		t.Fatalf("expected scored caregiver item, got %d ok=%v", score, ok)
	}
	profileScore, ok := pep3BookletPDFCaregiverProfileScore(&submission, "personal_self_care")
	if !ok || !profileScore.HasBreakdown || profileScore.PassedScore != 22 || profileScore.PartialScore != 1 || profileScore.TotalScore != 23 {
		t.Fatalf("expected PSC profile breakdown from caregiver answers, got %+v ok=%v", profileScore, ok)
	}
}

func TestPEP3BookletPDFExportScopes(t *testing.T) {
	tests := []struct {
		raw   string
		code  string
		pages []int
	}{
		{raw: "", code: "all", pages: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26}},
		{raw: "test_score", code: "test_score", pages: []int{1}},
		{raw: "development_profile", code: "development_profile", pages: []int{19}},
		{raw: "score_and_profile", code: "score_and_profile", pages: []int{1, 19}},
		{raw: "scoring_tables", code: "scoring_tables", pages: []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18}},
		{raw: "education_plan", code: "education_plan", pages: []int{20, 21, 22, 23, 24, 25, 26}},
	}
	for _, tt := range tests {
		got := normalizePEP3BookletPDFExportScope(tt.raw)
		if got.Code != tt.code || !reflect.DeepEqual(got.Pages, tt.pages) {
			t.Fatalf("scope %q = %+v, want code=%s pages=%+v", tt.raw, got, tt.code, tt.pages)
		}
	}
}

func TestBuildPEP3BookletPDFUsesTemplateBackground(t *testing.T) {
	if _, err := loadPEP3PDFFontBytes(); err != nil {
		t.Fatalf("PEP-3 PDF font not available: %v", err)
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
		"itemScoreList":   itemScoreList,
		"caregiverReport": pep3BookletTestCaregiverReportSubmission(t),
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
			ScaleVersion:   "2025-92题版",
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
	}, "测试机构", "all")
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

func pep3BookletTestCaregiverReportSubmission(t *testing.T) model.PEP3CaregiverReportSubmission {
	t.Helper()
	template := pep3CaregiverReportTemplate()
	answers := map[string]map[string]any{}
	for _, section := range template.Sections {
		if !section.Scored {
			continue
		}
		answers[section.SectionCode] = map[string]any{}
		for _, item := range section.Items {
			if !item.Scored {
				continue
			}
			answers[section.SectionCode][item.Key] = item.Options[0].Value
		}
	}
	answers["problem_behavior"]["speech_delay_or_absent"] = "mild_moderate"
	answers["problem_behavior"]["peer_friendship"] = "severe"
	answers["personal_self_care"]["bathe"] = "score_1"
	answers["personal_self_care"]["toileting"] = "score_0"
	answers["adaptive_behavior"]["activity_transition"] = "score_0"
	answers["adaptive_behavior"]["play_with_children"] = "score_1"
	rawScores, missing, err := scorePEP3CaregiverReportAnswers(template, answers)
	if err != nil {
		t.Fatalf("score caregiver report: %v", err)
	}
	if len(missing) > 0 {
		t.Fatalf("unexpected missing caregiver answers: %+v", missing)
	}
	return model.PEP3CaregiverReportSubmission{
		RespondentName: "家长",
		Relationship:   "母亲",
		Answers:        answers,
		RawScores:      rawScores,
		RawScoreList:   pep3CaregiverRawScoreListFromMap(rawScores),
		Source:         "test",
	}
}

func pep3BookletTestScoreRaw(t *testing.T) []byte {
	t.Helper()
	score := PEP3ScoreResponse{
		PEP3ScoreDataInfo: PEP3ScoreDataInfo{
			ScaleCode:    "PEP3",
			ScaleVersion: "2025-92题版",
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
