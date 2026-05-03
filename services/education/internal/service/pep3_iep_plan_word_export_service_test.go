package service

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
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
		`<w:gridCol w:w="1000"/>`,
		`<w:gridCol w:w="2500"/>`,
		`<w:gridCol w:w="3700"/>`,
		`<w:vMerge w:val="restart"/>`,
		`<w:vMerge/>`,
		`<w:ind w:left="160"/>`,
		`<w:spacing w:before="0" w:after="0" w:line="240" w:lineRule="auto"/>`,
		`<w:t>姓名</w:t>`,
		`<w:t>康复领域</w:t>`,
		`<w:t>课程形式</w:t>`,
		`<w:t>起止日期</w:t>`,
		`<w:t>实施起止日期</w:t>`,
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
}
