package service

import (
	"strings"
	"testing"

	"github.com/signintech/gopdf"
	"go-migration-platform/services/education/internal/model"
)

func TestBuildAutismDevSelectedReportDirectPDFCombinesSections(t *testing.T) {
	assessment := autismDevAssessmentSituationWordExport{
		Title:              autismDevAssessmentSituationTitle,
		StudentName:        "林一",
		Age:                "4岁3月",
		BirthDate:          "2021-02-01",
		AssessmentDate:     "2026-05-11",
		ExaminerName:       "陈老师",
		AssessmentSequence: "第2次",
		DevelopmentRows: []autismDevAssessmentSituationDevelopmentRow{
			{Label: "语言与沟通能力", Measured: true, PCount: 29, SupportCount: 50, TotalScore: 29},
			{Label: "认知能力", Measured: true, PCount: 12, SupportCount: 44, TotalScore: 12},
		},
		BehaviorRows: []autismDevAssessmentSituationBehaviorRow{
			{Label: "情绪理解", Measured: true, ACount: 3, MCount: 1, SCount: 0},
		},
	}
	analysis := autismDevResultAnalysisWordExport{
		Title:          "孤独症儿童评估结果分析表",
		StudentName:    "林一",
		AssessmentDate: "2026-05-11",
		ExaminerName:   "陈老师",
		Rows: []model.AutismDevResultAnalysisRow{
			{Domain: "语言与沟通", Status: "能理解常用指令。", Strengths: "可配合熟悉任务。", Weaknesses: "主动表达不足。", Targets: "1 能主动表达需求。"},
			{Domain: "认知能力", Status: "能完成部分配对和分类任务。", Strengths: "对熟悉材料反应较稳定。", Weaknesses: "新任务中提示依赖较多。", Targets: "1 能按颜色或形状分类。\n2 能在提示减少后完成配对。"},
		},
	}

	builder := newAutismDevSelectedReportPDFBuilder()
	if err := builder.appendDirectDraw(funcPDFDrawAutismDevAssessmentSituation(assessment)); err != nil {
		t.Fatalf("append assessment situation PDF failed: %v", err)
	}
	if err := builder.appendDirectDraw(funcPDFDrawAutismDevResultAnalysis(analysis)); err != nil {
		t.Fatalf("append result analysis PDF failed: %v", err)
	}
	content, err := builder.bytes()
	if err != nil {
		t.Fatalf("build selected report PDF failed: %v", err)
	}
	if len(content) == 0 {
		t.Fatal("selected report PDF should not be empty")
	}
	if count := strings.Count(string(content), "/Type /Page"); count < 2 {
		t.Fatalf("selected report PDF should contain at least 2 pages, got %d", count)
	}
}

func funcPDFDrawAutismDevAssessmentSituation(export autismDevAssessmentSituationWordExport) func(pdf *gopdf.GoPdf) error {
	return func(pdf *gopdf.GoPdf) error {
		return drawAutismDevAssessmentSituationPDFPages(pdf, export)
	}
}

func funcPDFDrawAutismDevResultAnalysis(export autismDevResultAnalysisWordExport) func(pdf *gopdf.GoPdf) error {
	return func(pdf *gopdf.GoPdf) error {
		return drawAutismDevResultAnalysisPDFPages(pdf, export)
	}
}
