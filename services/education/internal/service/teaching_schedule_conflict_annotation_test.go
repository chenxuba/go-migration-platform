package service

import (
	"testing"
	"time"

	"go-migration-platform/services/education/internal/model"
)

func TestAnnotateTeachingScheduleConflicts_MarksGroupClassStudentOverlap(t *testing.T) {
	schedules := []model.TeachingScheduleVO{
		{
			ID:              "1",
			ClassType:       model.TeachingClassTypeNormal,
			TeachingClassID: "1001",
			StudentID:       "201,202,203",
			LessonDate:      "2026-04-09",
			StartAt:         time.Date(2026, 4, 9, 10, 30, 0, 0, time.Local),
			EndAt:           time.Date(2026, 4, 9, 11, 10, 0, 0, time.Local),
		},
		{
			ID:              "2",
			ClassType:       model.TeachingClassTypeOneToOne,
			TeachingClassID: "2001",
			StudentID:       "202",
			LessonDate:      "2026-04-09",
			StartAt:         time.Date(2026, 4, 9, 10, 50, 0, 0, time.Local),
			EndAt:           time.Date(2026, 4, 9, 11, 30, 0, 0, time.Local),
		},
	}

	annotateTeachingScheduleConflicts(schedules)

	if !schedules[0].Conflict || !containsStringValue(schedules[0].ConflictTypes, "学员") {
		t.Fatalf("expected group class schedule to be marked with 学员 conflict, got conflict=%v conflictTypes=%v", schedules[0].Conflict, schedules[0].ConflictTypes)
	}
	if !schedules[1].Conflict || !containsStringValue(schedules[1].ConflictTypes, "学员") {
		t.Fatalf("expected overlapping one-to-one schedule to be marked with 学员 conflict, got conflict=%v conflictTypes=%v", schedules[1].Conflict, schedules[1].ConflictTypes)
	}
}

func TestAnnotateTeachingScheduleConflicts_MarksGroupClassAndAssistantOverlap(t *testing.T) {
	schedules := []model.TeachingScheduleVO{
		{
			ID:              "1",
			ClassType:       model.TeachingClassTypeNormal,
			TeachingClassID: "3001",
			AssistantIDs:    []string{"900"},
			LessonDate:      "2026-04-10",
			StartAt:         time.Date(2026, 4, 10, 9, 0, 0, 0, time.Local),
			EndAt:           time.Date(2026, 4, 10, 9, 45, 0, 0, time.Local),
		},
		{
			ID:              "2",
			ClassType:       model.TeachingClassTypeNormal,
			TeachingClassID: "3001",
			AssistantIDs:    []string{"900", "901"},
			LessonDate:      "2026-04-10",
			StartAt:         time.Date(2026, 4, 10, 9, 15, 0, 0, time.Local),
			EndAt:           time.Date(2026, 4, 10, 10, 0, 0, 0, time.Local),
		},
	}

	annotateTeachingScheduleConflicts(schedules)

	for _, item := range schedules {
		if !item.Conflict {
			t.Fatalf("expected schedule %s to be marked as conflict", item.ID)
		}
		if !containsStringValue(item.ConflictTypes, "班级") {
			t.Fatalf("expected schedule %s to include 班级 conflict, got %v", item.ID, item.ConflictTypes)
		}
		if !containsStringValue(item.ConflictTypes, "助教") {
			t.Fatalf("expected schedule %s to include 助教 conflict, got %v", item.ID, item.ConflictTypes)
		}
	}
}

func TestFilterTeachingSchedulesByConflictTypes_ReturnsOnlyMatchedConflicts(t *testing.T) {
	schedules := []model.TeachingScheduleVO{
		{
			ID:            "1",
			Conflict:      true,
			ConflictTypes: []string{"老师", "学员"},
		},
		{
			ID:            "2",
			Conflict:      true,
			ConflictTypes: []string{"教室"},
		},
		{
			ID:            "3",
			Conflict:      false,
			ConflictTypes: []string{"老师"},
		},
	}

	filtered := filterTeachingSchedulesByConflictTypes(schedules, []string{"老师"})

	if len(filtered) != 1 {
		t.Fatalf("expected 1 matched schedule, got %d", len(filtered))
	}
	if filtered[0].ID != "1" {
		t.Fatalf("expected schedule 1 to remain after filtering, got %s", filtered[0].ID)
	}
}
