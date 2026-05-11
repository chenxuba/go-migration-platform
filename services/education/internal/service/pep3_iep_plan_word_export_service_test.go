package service

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"

	"go-migration-platform/services/education/internal/model"
)

func TestBuildPEP3IEPPlanWordDocxShortGoalsAreTableRows(t *testing.T) {
	plan := buildStaticPEP3IEPPlanWordExport(3)
	docxBytes, err := buildPEP3IEPPlanWordDocx(plan)
	if err != nil {
		t.Fatalf("buildPEP3IEPPlanWordDocx failed: %v", err)
	}

	reader, err := zip.NewReader(bytes.NewReader(docxBytes), int64(len(docxBytes)))
	if err != nil {
		t.Fatalf("open generated docx failed: %v", err)
	}
	documentXML := readZipEntryForTest(t, reader, "word/document.xml")

	for _, expected := range []string{
		`<w:pgSz w:w="11906" w:h="16838"/>`,
		`<w:pgMar w:top="900" w:right="900" w:bottom="900" w:left="900"`,
		`<w:tblW w:w="10080" w:type="dxa"/>`,
		`<w:gridCol w:w="1038"/>`,
		`<w:gridCol w:w="1472"/>`,
		`<w:gridCol w:w="625"/>`,
		`<w:gridCol w:w="877"/>`,
		`<w:gridCol w:w="1260"/>`,
		`<w:gridCol w:w="2122"/>`,
		`<w:gridCol w:w="927"/>`,
		`<w:gridCol w:w="1759"/>`,
		`<w:trHeight w:val="660" w:hRule="atLeast"/>`,
		`<w:noWrap/>`,
		`<w:vMerge w:val="restart"/>`,
		`<w:vMerge/>`,
		`<w:ind w:left="100"/>`,
		`<w:ind w:left="120"/>`,
		`<w:spacing w:before="0" w:after="0" w:line="240" w:lineRule="auto"/>`,
		`<w:t>姓名</w:t>`,
		`<w:t>康复</w:t>`,
		`<w:t>领域</w:t>`,
		`<w:t>课程</w:t>`,
		`<w:t>形式</w:t>`,
		`<w:t>起止日期</w:t>`,
		`<w:t>实施</w:t>`,
		`<w:br w:type="page"/>`,
		`1. 用单词表达需求`,
		`2. 模仿功能词：要、不要`,
		`3. 二选一情境选择表达`,
		`4. 主动使用2词短句提出请求`,
	} {
		if !strings.Contains(documentXML, expected) {
			t.Fatalf("generated IEP document.xml missing %q", expected)
		}
	}

	firstGoalIndex := strings.Index(documentXML, `<w:t>1. 用单词表达需求</w:t>`)
	secondGoalIndex := strings.Index(documentXML, `<w:t>2. 模仿功能词：要、不要</w:t>`)
	if firstGoalIndex < 0 || secondGoalIndex < 0 || secondGoalIndex <= firstGoalIndex {
		t.Fatalf("generated IEP document.xml has unexpected short goal order")
	}
	if !strings.Contains(documentXML[firstGoalIndex:secondGoalIndex], `</w:tr><w:tr>`) {
		t.Fatalf("short-term goals should be split into separate table rows")
	}
	if strings.Contains(documentXML, "第1个月 建立表达基础") {
		t.Fatalf("short-term goal rows should not include stage headers")
	}
	if strings.Contains(documentXML, "儿童姓名") {
		t.Fatalf("student name label should be split into two lines")
	}
	pageBreakIndex := strings.Index(documentXML, `<w:br w:type="page"/>`)
	homePlanIndex := strings.Index(documentXML, "家庭干预计划")
	if pageBreakIndex < 0 || homePlanIndex < 0 || !(pageBreakIndex < homePlanIndex) {
		t.Fatalf("home intervention plan should start after page break")
	}
	if strings.Count(documentXML, "<w:tbl>") != 2 {
		t.Fatalf("main IEP content should be rendered as a single table before the home plan section")
	}
}

func TestBuildPEP3WeeklyPlanWordDocxMergesDateLabels(t *testing.T) {
	plan := modelPEP3WeeklyPlanForWordTest()
	docxBytes, err := buildPEP3WeeklyPlanWordDocx(plan)
	if err != nil {
		t.Fatalf("buildPEP3WeeklyPlanWordDocx failed: %v", err)
	}

	reader, err := zip.NewReader(bytes.NewReader(docxBytes), int64(len(docxBytes)))
	if err != nil {
		t.Fatalf("open generated docx failed: %v", err)
	}
	documentXML := readZipEntryForTest(t, reader, "word/document.xml")

	for _, expected := range []string{
		`<w:gridCol w:w="1925"/>`,
		`<w:gridCol w:w="698"/>`,
		`<w:trHeight w:val="540" w:hRule="atLeast"/>`,
		`<w:tcW w:w="1396" w:type="dxa"/><w:gridSpan w:val="2"/><w:vAlign w:val="center"/><w:noWrap/>`,
		`<w:tcW w:w="2792" w:type="dxa"/><w:gridSpan w:val="4"/><w:vAlign w:val="center"/>`,
		`<w:t>出生年月</w:t>`,
		`<w:t>训练日期</w:t>`,
	} {
		if !strings.Contains(documentXML, expected) {
			t.Fatalf("generated weekly document.xml missing %q", expected)
		}
	}

	birthLabelIndex := strings.Index(documentXML, `<w:t>出生年月</w:t>`)
	birthDateIndex := strings.Index(documentXML, `<w:t>2023-04-30</w:t>`)
	trainingLabelIndex := strings.Index(documentXML, `<w:t>训练日期</w:t>`)
	trainingDateIndex := strings.Index(documentXML, `<w:t>2026-05-01 至 2026-05-06</w:t>`)
	if birthLabelIndex < 0 || birthDateIndex < 0 || birthDateIndex <= birthLabelIndex {
		t.Fatalf("birth date value should follow the merged birth date label")
	}
	if trainingLabelIndex < 0 || trainingDateIndex < 0 || trainingDateIndex <= trainingLabelIndex {
		t.Fatalf("training date value should follow the merged training date label")
	}
}

func modelPEP3WeeklyPlanForWordTest() model.PEP3WeeklyPlanAIResult {
	return model.PEP3WeeklyPlanAIResult{
		Title:        "康复教学周计划日记录卡5月第1周",
		Student:      model.PEP3IEPPlanStudent{Name: "张一鸣", Gender: "男", BirthDate: "2023-04-30"},
		TeacherName:  "陈瑞",
		CourseName:   "康复教学",
		TrainingDate: "2026-05-01 至 2026-05-06",
		Preparation:  "彩色珠子、透明塑料杯",
		WeekDates:    []string{"2026-05-01", "2026-05-02", "2026-05-03", "2026-05-04", "2026-05-05", "2026-05-06"},
		Rows: []model.PEP3WeeklyPlanRow{
			{
				Project:    "大肌肉：扶持下单脚站立",
				Content:    "治疗师双手扶住儿童腰部，辅助抬起右脚，保持左脚站立。",
				Completion: []string{"", "", "", "", "", ""},
			},
		},
	}
}
