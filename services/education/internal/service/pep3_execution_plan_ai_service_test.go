package service

import (
	"testing"

	"go-migration-platform/services/education/internal/model"
)

func TestNormalizePEP3MonthlyTrainingItemsAlignsTrainingItemCountAndDatesToWeekRanges(t *testing.T) {
	target := pep3ExecutionPlanTarget{
		StartDate:  "2026-05-09",
		EndDate:    "2026-05-31",
		WeekRanges: []string{"2026-05-09 - 2026-05-09", "2026-05-11 - 2026-05-16", "2026-05-18 - 2026-05-23", "2026-05-25 - 2026-05-30"},
	}
	row := model.PEP3MonthlyPlanRow{
		ShortGoal: "提升单脚站立与平衡能力",
		TrainingItems: []model.PEP3MonthlyTrainingItem{
			{Content: "项目一", StartEndDate: "2026-05-09 - 2026-05-10"},
			{Content: "项目二", StartEndDate: "2026-05-11 - 2026-05-20"},
		},
	}

	items := normalizePEP3MonthlyTrainingItems(row, target)

	if len(items) != 4 {
		t.Fatalf("expected 4 training items, got %d", len(items))
	}
	if items[0].StartEndDate != "2026-05-09 - 2026-05-09" {
		t.Fatalf("unexpected first date range: %s", items[0].StartEndDate)
	}
	if items[1].StartEndDate != "2026-05-11 - 2026-05-16" {
		t.Fatalf("unexpected second date range: %s", items[1].StartEndDate)
	}
	if items[2].StartEndDate != "2026-05-18 - 2026-05-23" {
		t.Fatalf("unexpected third date range: %s", items[2].StartEndDate)
	}
	if items[3].StartEndDate != "2026-05-25 - 2026-05-30" {
		t.Fatalf("unexpected fourth date range: %s", items[3].StartEndDate)
	}
}

func TestNormalizePEP3MonthlyTrainingItemsStripsWeekAndStagePrefixes(t *testing.T) {
	target := pep3ExecutionPlanTarget{
		StartDate:  "2026-05-11",
		EndDate:    "2026-05-31",
		WeekRanges: []string{"2026-05-11 - 2026-05-15", "2026-05-18 - 2026-05-22", "2026-05-25 - 2026-05-29"},
	}
	row := model.PEP3MonthlyPlanRow{
		ShortGoal: "提升分类能力",
		TrainingItems: []model.PEP3MonthlyTrainingItem{
			{Content: "第一周：按颜色分类积木", StartEndDate: "2026-05-11 - 2026-05-15"},
			{Content: "强化提升：按大小分类玩具", StartEndDate: "2026-05-18 - 2026-05-22"},
			{Content: "第三周训练内容：按形状分类卡片", StartEndDate: "2026-05-25 - 2026-05-29"},
		},
	}

	items := normalizePEP3MonthlyTrainingItems(row, target)

	if items[0].Content != "按颜色分类积木" {
		t.Fatalf("unexpected first content: %s", items[0].Content)
	}
	if items[1].Content != "按大小分类玩具" {
		t.Fatalf("unexpected second content: %s", items[1].Content)
	}
	if items[2].Content != "按形状分类卡片" {
		t.Fatalf("unexpected third content: %s", items[2].Content)
	}
}

func TestBuildExecutionPlanTargetIncludesMonthWeekRangesAndSelectedWeekRange(t *testing.T) {
	sourcePlan := model.PEP3IEPPlanAIResult{
		Meta: model.PEP3IEPPlanMeta{
			StartDate: "2026-05-09",
			EndDate:   "2026-08-09",
		},
	}

	target := buildExecutionPlanTarget(sourcePlan, 3, 1, 2, []int{7})

	if len(target.WeekRanges) != 4 {
		t.Fatalf("expected 4 month week ranges, got %#v", target.WeekRanges)
	}
	if target.WeekRangeText != "2026-05-11 - 2026-05-16" {
		t.Fatalf("unexpected selected week range: %s", target.WeekRangeText)
	}
	if target.WeekStartDate != "2026-05-11" || target.WeekEndDate != "2026-05-16" {
		t.Fatalf("unexpected selected week bounds: %s - %s", target.WeekStartDate, target.WeekEndDate)
	}
}
