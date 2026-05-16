package service

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"

	"github.com/signintech/gopdf"
	"go-migration-platform/services/education/internal/model"
)

func TestBuildAutismDevSelectedReportWordDocxCombinesWordSections(t *testing.T) {
	docxBytes, err := buildAutismDevSelectedReportWordDocx([]autismDevSelectedReportWordSectionExport{
		{
			Section: AutismDevReportSectionAssessmentInfo,
			AssessmentSituation: autismDevAssessmentSituationWordExport{
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
			},
		},
		{
			Section: AutismDevReportSectionResultAnalysis,
			ResultAnalysis: autismDevResultAnalysisWordExport{
				Title:          "孤独症儿童评估结果分析表",
				StudentName:    "林一",
				AssessmentDate: "2026-05-11",
				ExaminerName:   "陈老师",
				Rows: []model.AutismDevResultAnalysisRow{
					{Domain: "语言与沟通", Status: "能理解常用指令。", Strengths: "可配合熟悉任务。", Weaknesses: "主动表达不足。", Targets: "1 能主动表达需求。"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("buildAutismDevSelectedReportWordDocx failed: %v", err)
	}
	reader, err := zip.NewReader(bytes.NewReader(docxBytes), int64(len(docxBytes)))
	if err != nil {
		t.Fatalf("open generated docx failed: %v", err)
	}
	documentXML := readZipEntryForTest(t, reader, "word/document.xml")
	for _, expected := range []string{"3.1 发展能力计分汇总表", "孤独症儿童评估结果分析表", "测评次数", "能力现状描述"} {
		if !strings.Contains(documentXML, expected) {
			t.Fatalf("generated document.xml missing %q", expected)
		}
	}
	if !strings.Contains(documentXML, `<w:br w:type="page"/>`) {
		t.Fatal("combined Word doc should page-break between selected sections")
	}
}

func TestMergeAutismDevReportPDFBytesPreservesPages(t *testing.T) {
	first := buildAutismDevSelectedReportTestPDF(t, 320, 480)
	second := buildAutismDevSelectedReportTestPDF(t, 595, 842)

	merged, err := mergeAutismDevReportPDFBytes([][]byte{first, second})
	if err != nil {
		t.Fatalf("mergeAutismDevReportPDFBytes failed: %v", err)
	}
	if len(merged) == 0 {
		t.Fatal("merged pdf should not be empty")
	}
	source := string(merged)
	if count := strings.Count(source, "/Type /Page"); count < 2 {
		t.Fatalf("merged pdf should contain at least 2 pages, got %d", count)
	}
	if strings.Contains(source, "/Subtype /Image") {
		t.Fatal("merged pdf should import PDF pages instead of rasterizing them as images")
	}
}

func buildAutismDevSelectedReportTestPDF(t *testing.T, width, height float64) []byte {
	t.Helper()
	var pdf gopdf.GoPdf
	pdf.Start(gopdf.Config{
		Unit:     gopdf.UnitPT,
		PageSize: gopdf.Rect{W: width, H: height},
	})
	pdf.AddPage()
	pdf.SetLineWidth(1)
	pdf.SetStrokeColor(0, 0, 0)
	pdf.Line(24, 24, width-24, height-24)
	content, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("build test pdf failed: %v", err)
	}
	return content
}
