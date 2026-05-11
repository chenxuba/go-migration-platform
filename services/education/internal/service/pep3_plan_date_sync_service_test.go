package service

import (
	"testing"
	"time"

	"go-migration-platform/services/education/internal/model"
)

func TestIEPPlanWholeMonthDateRangeFromStartUsesAssessmentAlignedEndDate(t *testing.T) {
	start := parseIEPPlanDateValue("2026-05-05")
	startText, endText := iepPlanWholeMonthDateRangeFromStart(start, 3)

	if startText != "2026-05-05" {
		t.Fatalf("unexpected start date: %s", startText)
	}
	if endText != "2026-08-05" {
		t.Fatalf("unexpected end date: %s", endText)
	}
}

func TestBuildExecutionPlanTargetUsesCalendarMonthsWithinAssessmentAlignedPeriod(t *testing.T) {
	sourcePlan := fakePEP3IEPPlanResultForDateSync("2026-05-05")
	target := buildExecutionPlanTarget(sourcePlan, 3, 4, 2, []int{7})

	if target.StartDate != "2026-08-01" {
		t.Fatalf("unexpected target start date: %s", target.StartDate)
	}
	if target.EndDate != "2026-08-05" {
		t.Fatalf("unexpected target end date: %s", target.EndDate)
	}
	if target.WeekIndex != 2 {
		t.Fatalf("unexpected target week index: %d", target.WeekIndex)
	}
	if len(target.WeekDates) == 0 || target.WeekDates[len(target.WeekDates)-1] != "2026-08-05" {
		t.Fatalf("unexpected target week dates: %#v", target.WeekDates)
	}
}

func TestBuildExecutionPlanTargetFiltersSelectedRestWeekdays(t *testing.T) {
	sourcePlan := fakePEP3IEPPlanResultForDateSync("2026-05-05")
	target := buildExecutionPlanTarget(sourcePlan, 3, 1, 4, []int{3, 7})

	for _, dateText := range target.WeekDates {
		date := parseIEPPlanDateValue(dateText)
		if date.IsZero() {
			t.Fatalf("unexpected empty week date in %#v", target.WeekDates)
		}
		if date.Weekday() == time.Wednesday || date.Weekday() == time.Sunday {
			t.Fatalf("rest weekday should be filtered out, got %#v", target.WeekDates)
		}
	}
}

func fakePEP3IEPPlanResultForDateSync(startDate string) model.PEP3IEPPlanAIResult {
	return model.PEP3IEPPlanAIResult{
		Title: "康复教学季度计划",
		Student: model.PEP3IEPPlanStudent{
			Name:      "陈旭",
			Gender:    "男",
			BirthDate: "2022-05-11",
		},
		Meta: model.PEP3IEPPlanMeta{
			PlanDate:    "2026-05-07",
			Participant: "陈瑞",
			Implementer: "陈瑞",
			StartDate:   startDate,
		},
		Rows: []model.PEP3IEPPlanRow{
			{Domain: "大肌肉", ShortGoal: "能独立行走3米", CourseForm: "个训"},
		},
	}
}
