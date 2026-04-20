package service

import (
	"testing"
	"time"

	"go-migration-platform/services/education/internal/model"
)

func TestBuildHomeworkSchedules_AutoExpandsOccurrences(t *testing.T) {
	plans, err := buildHomeworkSchedules(&model.HomeworkRepeatRule{
		StartDate:  "2026-04-21",
		EndDate:    "2026-04-24",
		RepeatSpan: 1,
		WeekDays:   4 + 8 + 16 + 32,
	}, "", "", 2, 0, 12, false)
	if err != nil {
		t.Fatalf("buildHomeworkSchedules returned error: %v", err)
	}
	if len(plans) != 4 {
		t.Fatalf("expected 4 plans, got %d", len(plans))
	}

	expectedPublishTimes := []string{
		"2026-04-21 02:00",
		"2026-04-22 02:00",
		"2026-04-23 02:00",
		"2026-04-24 02:00",
	}
	expectedEndTimes := []string{
		"2026-04-21 14:00",
		"2026-04-22 14:00",
		"2026-04-23 14:00",
		"2026-04-24 14:00",
	}

	for index, plan := range plans {
		if plan.PublishRule != model.HomeworkPublishRuleOnce {
			t.Fatalf("plan %d publish rule = %d, want %d", index, plan.PublishRule, model.HomeworkPublishRuleOnce)
		}
		if plan.RepeatRule != nil {
			t.Fatalf("plan %d repeat rule should be nil after expansion", index)
		}
		if plan.PublishTime == nil {
			t.Fatalf("plan %d publish time should not be nil", index)
		}
		if plan.EndTime == nil {
			t.Fatalf("plan %d end time should not be nil", index)
		}
		if got := plan.PublishTime.In(time.Local).Format("2006-01-02 15:04"); got != expectedPublishTimes[index] {
			t.Fatalf("plan %d publish time = %s, want %s", index, got, expectedPublishTimes[index])
		}
		if got := plan.EndTime.In(time.Local).Format("2006-01-02 15:04"); got != expectedEndTimes[index] {
			t.Fatalf("plan %d end time = %s, want %s", index, got, expectedEndTimes[index])
		}
		if plan.TaskDurationHours != 12 {
			t.Fatalf("plan %d task duration hours = %d, want 12", index, plan.TaskDurationHours)
		}
		if plan.EndHour != 14 {
			t.Fatalf("plan %d end hour = %d, want 14", index, plan.EndHour)
		}
	}
}

func TestEnumerateHomeworkOccurrenceDates_RespectsRepeatSpan(t *testing.T) {
	startDate := time.Date(2026, 4, 20, 0, 0, 0, 0, time.Local)
	finishDate := time.Date(2026, 5, 10, 0, 0, 0, 0, time.Local)

	dates := enumerateHomeworkOccurrenceDates(startDate, finishDate, 2, 2)
	expected := []string{"2026-04-20", "2026-05-04"}
	if len(dates) != len(expected) {
		t.Fatalf("expected %d dates, got %d", len(expected), len(dates))
	}
	for index, date := range dates {
		if got := date.In(time.Local).Format("2006-01-02"); got != expected[index] {
			t.Fatalf("date %d = %s, want %s", index, got, expected[index])
		}
	}
}
