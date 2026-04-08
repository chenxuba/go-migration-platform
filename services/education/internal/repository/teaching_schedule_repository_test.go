package repository

import (
	"testing"
	"time"
)

func TestApplyCreateScheduleConflictAllowances(t *testing.T) {
	plans := []normalizedSchedulePlan{
		{},
		{AllowStudentConflict: true},
		{AllowClassroomConflict: true},
	}

	applyCreateScheduleConflictAllowances(plans, true, true)

	for i, plan := range plans {
		if !plan.AllowStudentConflict {
			t.Fatalf("expected plan %d to allow student conflict", i)
		}
		if !plan.AllowClassroomConflict {
			t.Fatalf("expected plan %d to allow classroom conflict", i)
		}
	}
}

func TestIsAssistantRemovalOnlyBatchUpdate(t *testing.T) {
	lessonDate := time.Date(2026, 4, 7, 0, 0, 0, 0, time.Local)
	startAt := time.Date(2026, 4, 7, 10, 55, 0, 0, time.Local)
	endAt := time.Date(2026, 4, 7, 11, 35, 0, 0, time.Local)

	baseSchedule := teachingScheduleRow{
		ID:            1,
		ClassType:     2,
		TeacherID:     100,
		ClassroomID:   200,
		ClassroomName: "A101",
		AssistantIDs:  []string{"300", "301"},
		LessonDate:    lessonDate,
		StartAt:       startAt,
		EndAt:         endAt,
	}
	baseSlot := normalizedScheduleSlot{
		LessonDate: lessonDate,
		StartAt:    startAt,
		EndAt:      endAt,
	}

	if !isAssistantRemovalOnlyBatchUpdate([]teachingScheduleRow{baseSchedule}, 100, 200, true, []int64{300}, []normalizedScheduleSlot{baseSlot}) {
		t.Fatalf("expected assistant removal only update to be recognized")
	}

	if !isAssistantRemovalOnlyBatchUpdate([]teachingScheduleRow{baseSchedule}, 100, 0, false, []int64{300}, []normalizedScheduleSlot{baseSlot}) {
		t.Fatalf("expected empty classroom request to still count as removal only update")
	}

	if isAssistantRemovalOnlyBatchUpdate([]teachingScheduleRow{baseSchedule}, 100, 200, true, []int64{300, 301}, []normalizedScheduleSlot{baseSlot}) {
		t.Fatalf("expected unchanged assistant list to not count as removal only update")
	}

	if isAssistantRemovalOnlyBatchUpdate([]teachingScheduleRow{baseSchedule}, 100, 200, true, []int64{300, 302}, []normalizedScheduleSlot{baseSlot}) {
		t.Fatalf("expected adding a new assistant to not count as removal only update")
	}

	shiftedSlot := normalizedScheduleSlot{
		LessonDate: lessonDate,
		StartAt:    time.Date(2026, 4, 7, 10, 40, 0, 0, time.Local),
		EndAt:      endAt,
	}
	if isAssistantRemovalOnlyBatchUpdate([]teachingScheduleRow{baseSchedule}, 100, 200, true, []int64{300}, []normalizedScheduleSlot{shiftedSlot}) {
		t.Fatalf("expected time changes to break removal only update detection")
	}
}
