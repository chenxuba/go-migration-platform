package service

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"

	"go-migration-platform/pkg/autismdevscore"
	"go-migration-platform/services/education/internal/model"
)

func TestBuildAutismDevAssessmentSituationWordDocx(t *testing.T) {
	docxBytes, err := buildAutismDevAssessmentSituationWordDocx(autismDevAssessmentSituationWordExport{
		Title:              autismDevAssessmentSituationTitle,
		StudentName:        "林一",
		Age:                "4岁3月",
		BirthDate:          "2021-02-01",
		AssessmentDate:     "2026-05-11",
		ExaminerName:       "陈老师",
		AssessmentSequence: "第2次",
		DevelopmentRows: []autismDevAssessmentSituationDevelopmentRow{
			{Label: "语言与沟通能力", Measured: true, PCount: 29, SupportCount: 50, TotalScore: 29},
		},
		BehaviorRows: []autismDevAssessmentSituationBehaviorRow{
			{Label: "依附情绪行为", Measured: true, ACount: 1, MCount: 1, SCount: 0},
		},
	})
	if err != nil {
		t.Fatalf("buildAutismDevAssessmentSituationWordDocx failed: %v", err)
	}

	reader, err := zip.NewReader(bytes.NewReader(docxBytes), int64(len(docxBytes)))
	if err != nil {
		t.Fatalf("open generated docx failed: %v", err)
	}
	documentXML := readZipEntryForTest(t, reader, "word/document.xml")
	for _, expected := range []string{
		"3.1 发展能力计分汇总表",
		"儿童姓名",
		"测评年龄",
		"测评次数",
		"林一",
		"第2次",
		"评估结果",
		"E+F(X)",
		"语言与沟通能力",
		"1、依附情绪行为",
	} {
		if !strings.Contains(documentXML, expected) {
			t.Fatalf("generated document.xml missing %q", expected)
		}
	}
	if strings.Contains(documentXML, "量表版本") {
		t.Fatal("generated document.xml should not include scale version")
	}
	if tableCount := strings.Count(documentXML, "<w:tbl>"); tableCount != 1 {
		t.Fatalf("generated document.xml should contain one table, got %d", tableCount)
	}
	if !strings.Contains(documentXML, `<w:gridSpan w:val="39"/>`) ||
		!strings.Contains(documentXML, `<w:vMerge w:val="restart"/>`) {
		t.Fatal("generated document.xml should use merged cells for the report header")
	}
}

func TestAutismDevAssessmentSituationRowsFollowCurrentScope(t *testing.T) {
	record := model.AssessmentRecordDetailVO{
		AssessmentRecordSummaryVO: model.AssessmentRecordSummaryVO{
			StudentName:        "林一",
			AssessmentSequence: 3,
		},
		InputJSON: []byte(`{"scopeMode":"custom","scopeDomainCodes":["LC","EB"]}`),
	}
	score := autismdevscore.AssessmentResult{
		Domains: []autismdevscore.DomainResult{
			{DomainCode: autismdevscore.DomainLanguageComm, ScoreType: autismdevscore.ScoreTypePEF, AnsweredItemCount: 2, PCount: 1, ECount: 1, RawScore: 1},
			{DomainCode: autismdevscore.DomainSensory, ScoreType: autismdevscore.ScoreTypePEF, AnsweredItemCount: 2, PCount: 2, RawScore: 2},
			{DomainCode: autismdevscore.DomainEmotionBehavior, ScoreType: autismdevscore.ScoreTypeAMS, AnsweredItemCount: 2, ACount: 1, MCount: 1},
		},
	}
	data := autismDevStaticData{
		domains: []autismDevDomainDefinition{
			{ScaleCode: autismdevscore.DomainSensory},
			{ScaleCode: autismdevscore.DomainLanguageComm},
			{ScaleCode: autismdevscore.DomainEmotionBehavior},
		},
	}
	itemScores := map[int]string{442: autismdevscore.ScoreA, 443: autismdevscore.ScoreM}

	export := buildAutismDevAssessmentSituationWordExport(record, score, data, itemScores)
	if len(export.DevelopmentRows) != 1 {
		t.Fatalf("expected one scoped development row, got %d", len(export.DevelopmentRows))
	}
	if export.DevelopmentRows[0].Label != "语言与沟通能力" {
		t.Fatalf("unexpected development row: %#v", export.DevelopmentRows[0])
	}
	if len(export.BehaviorRows) == 0 {
		t.Fatal("expected behavior rows when EB is in current scope")
	}
	if export.AssessmentSequence != "第3次" {
		t.Fatalf("unexpected assessment sequence: %q", export.AssessmentSequence)
	}
}
