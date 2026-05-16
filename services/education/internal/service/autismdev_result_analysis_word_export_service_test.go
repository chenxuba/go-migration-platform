package service

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
	"time"

	"go-migration-platform/services/education/internal/model"
)

func TestBuildAutismDevResultAnalysisWordDocx(t *testing.T) {
	docxBytes, err := buildAutismDevResultAnalysisWordDocx(autismDevResultAnalysisWordExport{
		Title:          "孤独症儿童评估结果分析表",
		StudentName:    "林一",
		Gender:         "男",
		BirthDate:      "2021-02-01",
		Age:            "4岁3个月",
		AssessmentDate: "2026-05-11",
		ExaminerName:   "陈老师",
		AssessmentName: "孤独症儿童发展评估表",
		Rows: []model.AutismDevResultAnalysisRow{
			{
				Domain:     "感知觉",
				Status:     "能注视光线刺激，但复杂追视仍未稳定。",
				Strengths:  "可配合熟悉刺激。",
				Weaknesses: "复杂辨别稳定性不足。",
				Targets:    "1 能追视移动物体。\n2 能对突发声音作出反应。",
			},
		},
	})
	if err != nil {
		t.Fatalf("buildAutismDevResultAnalysisWordDocx failed: %v", err)
	}

	reader, err := zip.NewReader(bytes.NewReader(docxBytes), int64(len(docxBytes)))
	if err != nil {
		t.Fatalf("open generated docx failed: %v", err)
	}

	documentXML := readZipEntryForTest(t, reader, "word/document.xml")
	for _, expected := range []string{"孤独症儿童评估结果分析表", "儿童姓名：", "评估者：", "评估时间：", "林一", "感知觉", "能追视移动物体", "优势：可配合熟悉刺激。", "劣势：复杂辨别稳定性不足。"} {
		if !strings.Contains(documentXML, expected) {
			t.Fatalf("generated document.xml missing %q", expected)
		}
	}
	for _, unexpected := range []string{"性别", "出生日期", "年龄", "量表"} {
		if strings.Contains(documentXML, unexpected) {
			t.Fatalf("generated document.xml should not include extra meta table label %q", unexpected)
		}
	}
	if tableCount := strings.Count(documentXML, "<w:tbl>"); tableCount != 1 {
		t.Fatalf("generated document.xml should only include the analysis content table, got %d tables", tableCount)
	}
	if !strings.Contains(documentXML, `<w:tab w:val="left" w:pos="3600"/>`) ||
		!strings.Contains(documentXML, `<w:tab w:val="left" w:pos="7000"/>`) {
		t.Fatal("generated document.xml should spread meta fields across the full row with tab stops")
	}
	metaEnd := strings.Index(documentXML, "评估时间：")
	tableStart := strings.Index(documentXML, "<w:tbl>")
	if metaEnd < 0 || tableStart < 0 || metaEnd > tableStart {
		t.Fatal("generated document.xml should place meta fields before the analysis table")
	}
	if strings.Contains(documentXML[metaEnd:tableStart], `<w:p><w:pPr><w:spacing w:after="`) {
		t.Fatal("generated document.xml should not insert an extra spacer paragraph between meta fields and table")
	}
}

func TestAutismDevStrengthWeaknessLinesAvoidBlankSpacer(t *testing.T) {
	lines := autismDevStrengthWeaknessLines(model.AutismDevResultAnalysisRow{
		Strengths:  "可配合熟悉刺激。",
		Weaknesses: "复杂辨别稳定性不足。",
	})
	expected := []string{"优势：可配合熟悉刺激。", "劣势：复杂辨别稳定性不足。"}
	if strings.Join(lines, "\n") != strings.Join(expected, "\n") {
		t.Fatalf("unexpected strength weakness lines: %#v", lines)
	}
}

func TestAutismDevResultAnalysisSourceHashIgnoresHeaderOnlyChanges(t *testing.T) {
	assessmentDate := time.Date(2026, 5, 10, 0, 0, 0, 0, time.Local)
	updatedTime := time.Date(2026, 5, 10, 9, 30, 0, 0, time.Local)
	record := model.AssessmentRecordDetailVO{
		AssessmentRecordSummaryVO: model.AssessmentRecordSummaryVO{
			AssessmentDate: &assessmentDate,
			ExaminerName:   "陈老师",
			DataStatus:     "completed",
			Remark:         "原备注",
			UpdatedTime:    &updatedTime,
		},
		InputJSON:  []byte(`{"itemScores":{"1":"P","2":"E"}}`),
		ResultJSON: []byte(`{"totalP":1,"totalE":1}`),
	}

	changedDate := time.Date(2026, 5, 12, 0, 0, 0, 0, time.Local)
	changedUpdatedTime := time.Date(2026, 5, 12, 10, 20, 0, 0, time.Local)
	headerChanged := record
	headerChanged.AssessmentDate = &changedDate
	headerChanged.ExaminerName = "李老师"
	headerChanged.UpdatedTime = &changedUpdatedTime
	if autismDevResultAnalysisSourceHash(record) != autismDevResultAnalysisSourceHash(headerChanged) {
		t.Fatal("source hash should not change when only assessment header fields change")
	}

	scoreChanged := record
	scoreChanged.ResultJSON = []byte(`{"totalP":2,"totalE":0}`)
	if autismDevResultAnalysisSourceHash(record) == autismDevResultAnalysisSourceHash(scoreChanged) {
		t.Fatal("source hash should change when score result changes")
	}
}
