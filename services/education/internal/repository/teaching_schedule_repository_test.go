package repository

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"go-migration-platform/services/education/internal/model"
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

func TestBuildTeachingClassScheduleExistsSQLUsesLiveTeachingScheduleTable(t *testing.T) {
	got := buildTeachingClassScheduleExistsSQL("tc", "1")
	if !strings.Contains(got, "FROM teaching_schedule ts") {
		t.Fatalf("expected existence SQL to query teaching_schedule, got %s", got)
	}
	if !strings.Contains(got, "ts.teaching_class_id = tc.id") {
		t.Fatalf("expected existence SQL to bind teaching_class_id to current class, got %s", got)
	}
	if !strings.Contains(got, "ts.status = "+strconv.Itoa(model.TeachingScheduleStatusActive)) {
		t.Fatalf("expected existence SQL to filter active schedules, got %s", got)
	}
}

func TestBuildTeachingClassFinishedCountSQLFallsBackWhenRecordSQLMissing(t *testing.T) {
	got := buildTeachingClassFinishedCountSQL("tc", "tc.class_type", "", "IFNULL(tc.finished_lesson_count, 0)")
	if got != "IFNULL(tc.finished_lesson_count, 0)" {
		t.Fatalf("expected fallback finished count SQL, got %s", got)
	}

	got = buildTeachingClassFinishedCountSQL("tc", "tc.class_type", "EXISTS (SELECT 1 FROM teaching_record tr WHERE tr.class_id = ts.teaching_class_id)", "0")
	if !strings.Contains(got, "FROM teaching_schedule ts") {
		t.Fatalf("expected live finished count SQL to query teaching_schedule, got %s", got)
	}
	if !strings.Contains(got, strconv.Itoa(model.TeachingScheduleStatusActive)) {
		t.Fatalf("expected live finished count SQL to keep active status constraint, got %s", got)
	}
}

func TestBuildGroupClassStudentRosterFromMembershipsUsesScheduleStartAsBoundary(t *testing.T) {
	scheduleStartAt := time.Date(2026, 4, 9, 10, 0, 0, 0, time.Local)
	roster := buildGroupClassStudentRosterFromMemberships([]groupClassStudentMembership{
		{
			StudentID:       1,
			StudentName:     "先加入的学员",
			JoinAt:          scheduleStartAt.Add(-time.Minute),
			ClassStatus:     model.TeachingClassStudentStatusStudying,
			StatusChangedAt: scheduleStartAt.Add(-time.Minute),
		},
		{
			StudentID:       2,
			StudentName:     "同一时刻加入的学员",
			JoinAt:          scheduleStartAt,
			ClassStatus:     model.TeachingClassStudentStatusStudying,
			StatusChangedAt: scheduleStartAt,
		},
		{
			StudentID:       3,
			StudentName:     "后加入的学员",
			JoinAt:          scheduleStartAt.Add(time.Minute),
			ClassStatus:     model.TeachingClassStudentStatusStudying,
			StatusChangedAt: scheduleStartAt.Add(time.Minute),
		},
	}, scheduleStartAt)

	if len(roster.IDs) != 1 || roster.IDs[0] != 1 {
		t.Fatalf("expected only students added before schedule start to be included, got %#v", roster.IDs)
	}
	if len(roster.Names) != 1 || roster.Names[0] != "先加入的学员" {
		t.Fatalf("expected roster names to match effective students, got %#v", roster.Names)
	}
}

func TestResolveGroupClassRosterReferenceAtUsesCreationTimeForRetroSchedules(t *testing.T) {
	scheduleStartAt := time.Date(2026, 4, 9, 9, 15, 0, 0, time.Local)
	scheduleCreatedAt := time.Date(2026, 4, 9, 20, 49, 0, 0, time.Local)

	got := resolveGroupClassRosterReferenceAt(scheduleStartAt, scheduleCreatedAt)
	if !got.Equal(scheduleCreatedAt) {
		t.Fatalf("expected retro schedule roster boundary to use create time, got %s", got)
	}
}

func TestNormalizeCreateSchedulePlansRemovesTeacherFromAssistantIDs(t *testing.T) {
	plans, err := normalizeCreateSchedulePlans([]model.TeachingScheduleCreateSlotDTO{
		{
			LessonDate:   "2026-04-09",
			StartTime:    "10:00",
			EndTime:      "11:00",
			TeacherID:    "100",
			AssistantIDs: []string{"100", "101", "101"},
		},
	}, 0, 0, nil)
	if err != nil {
		t.Fatalf("expected normalizeCreateSchedulePlans to succeed, got %v", err)
	}
	if len(plans) != 1 {
		t.Fatalf("expected one normalized plan, got %d", len(plans))
	}
	if len(plans[0].AssistantIDs) != 1 || plans[0].AssistantIDs[0] != 101 {
		t.Fatalf("expected duplicate teacher assistant to be removed, got %#v", plans[0].AssistantIDs)
	}
}

func TestNormalizeCreateSchedulePlansRemovesTeacherFromFallbackAssistantIDs(t *testing.T) {
	plans, err := normalizeCreateSchedulePlans([]model.TeachingScheduleCreateSlotDTO{
		{
			LessonDate: "2026-04-09",
			StartTime:  "14:00",
			EndTime:    "15:00",
			TeacherID:  "200",
		},
	}, 0, 0, []int64{200, 201})
	if err != nil {
		t.Fatalf("expected normalizeCreateSchedulePlans to succeed with fallback assistants, got %v", err)
	}
	if len(plans) != 1 {
		t.Fatalf("expected one normalized plan, got %d", len(plans))
	}
	if len(plans[0].AssistantIDs) != 1 || plans[0].AssistantIDs[0] != 201 {
		t.Fatalf("expected fallback assistant matching teacher to be removed, got %#v", plans[0].AssistantIDs)
	}
}

func TestBuildGroupClassStudentRosterFromMembershipsIncludesStudentsJoinedBeforeRetroScheduleCreation(t *testing.T) {
	scheduleStartAt := time.Date(2026, 4, 9, 9, 15, 0, 0, time.Local)
	scheduleCreatedAt := time.Date(2026, 4, 9, 20, 49, 0, 0, time.Local)
	referenceAt := resolveGroupClassRosterReferenceAt(scheduleStartAt, scheduleCreatedAt)

	roster := buildGroupClassStudentRosterFromMemberships([]groupClassStudentMembership{
		{
			StudentID:   1,
			StudentName: "补排前已在班里的学员",
			JoinAt:      time.Date(2026, 4, 9, 12, 58, 42, 0, time.Local),
		},
		{
			StudentID:   2,
			StudentName: "补排后才加入的学员",
			JoinAt:      scheduleCreatedAt.Add(time.Minute),
		},
	}, referenceAt)

	if len(roster.IDs) != 1 || roster.IDs[0] != 1 {
		t.Fatalf("expected retro schedule roster to include only students joined before create time, got %#v", roster.IDs)
	}
	if len(roster.Names) != 1 || roster.Names[0] != "补排前已在班里的学员" {
		t.Fatalf("expected retro schedule roster names to use create time boundary, got %#v", roster.Names)
	}
}

func TestBuildGroupClassStudentRosterFromMembershipsExcludesStudentsRemovedBeforeSchedule(t *testing.T) {
	scheduleStartAt := time.Date(2026, 4, 9, 10, 0, 0, 0, time.Local)
	roster := buildGroupClassStudentRosterFromMemberships([]groupClassStudentMembership{
		{
			StudentID:       1,
			StudentName:     "已移出的学员",
			JoinAt:          scheduleStartAt.Add(-2 * time.Hour),
			ClassStatus:     model.TeachingClassStudentStatusClosed,
			StatusChangedAt: scheduleStartAt.Add(-time.Minute),
		},
		{
			StudentID:       2,
			StudentName:     "稍后才移出的学员",
			JoinAt:          scheduleStartAt.Add(-2 * time.Hour),
			ClassStatus:     model.TeachingClassStudentStatusClosed,
			StatusChangedAt: scheduleStartAt.Add(time.Minute),
		},
		{
			StudentID:       3,
			StudentName:     "正好此刻移出的学员",
			JoinAt:          scheduleStartAt.Add(-2 * time.Hour),
			ClassStatus:     model.TeachingClassStudentStatusClosed,
			StatusChangedAt: scheduleStartAt,
		},
	}, scheduleStartAt)

	if len(roster.IDs) != 1 || roster.IDs[0] != 2 {
		t.Fatalf("expected only students removed after schedule start to remain, got %#v", roster.IDs)
	}
	if len(roster.Names) != 1 || roster.Names[0] != "稍后才移出的学员" {
		t.Fatalf("expected roster names to exclude already removed students, got %#v", roster.Names)
	}
}

func TestBuildGroupClassScheduleRosterFromMembershipsAndOverrides_RemovesCurrentLessonOnly(t *testing.T) {
	scheduleStartAt := time.Date(2026, 4, 9, 10, 0, 0, 0, time.Local)
	roster := buildGroupClassScheduleRosterFromMembershipsAndOverrides(
		[]groupClassStudentMembership{
			{
				StudentID:   1,
				StudentName: "被移出本节的学员",
				JoinAt:      scheduleStartAt.Add(-time.Hour),
			},
			{
				StudentID:   2,
				StudentName: "仍然保留的学员",
				JoinAt:      scheduleStartAt.Add(-time.Hour),
			},
		},
		[]teachingScheduleStudentOverride{
			{
				StudentID:    1,
				StudentType:  model.TeachingScheduleStudentTypeClassMember,
				RosterStatus: model.TeachingScheduleStudentRosterStatusRemoved,
			},
		},
		scheduleStartAt,
	)

	if len(roster.Active) != 1 || roster.Active[0].StudentID != 2 {
		t.Fatalf("expected current-lesson removal to exclude only the overridden student, got %#v", roster.Active)
	}
	if len(roster.Leave) != 0 {
		t.Fatalf("expected no leave students after removal override, got %#v", roster.Leave)
	}
}

func TestBuildGroupClassScheduleRosterFromMembershipsAndOverrides_SplitsLeaveAndScheduleOnlyStudents(t *testing.T) {
	scheduleStartAt := time.Date(2026, 4, 9, 10, 0, 0, 0, time.Local)
	roster := buildGroupClassScheduleRosterFromMembershipsAndOverrides(
		[]groupClassStudentMembership{
			{
				StudentID:   1,
				StudentName: "班级学员",
				JoinAt:      scheduleStartAt.Add(-time.Hour),
			},
		},
		[]teachingScheduleStudentOverride{
			{
				StudentID:    1,
				StudentType:  model.TeachingScheduleStudentTypeClassMember,
				RosterStatus: model.TeachingScheduleStudentRosterStatusLeave,
			},
			{
				StudentID:    2,
				StudentName:  "临时学员",
				StudentType:  model.TeachingScheduleStudentTypeTemporary,
				RosterStatus: model.TeachingScheduleStudentRosterStatusActive,
			},
			{
				StudentID:    3,
				StudentName:  "补课学员",
				StudentType:  model.TeachingScheduleStudentTypeMakeup,
				RosterStatus: model.TeachingScheduleStudentRosterStatusActive,
			},
		},
		scheduleStartAt,
	)

	if len(roster.Leave) != 1 || roster.Leave[0].StudentID != 1 {
		t.Fatalf("expected class member leave override to move the student into leave tab, got %#v", roster.Leave)
	}
	if len(roster.Active) != 2 {
		t.Fatalf("expected schedule-only active students to stay in active roster, got %#v", roster.Active)
	}
	if roster.Active[0].StudentID != 2 || roster.Active[0].ScheduleStudentType != model.TeachingScheduleStudentTypeTemporary {
		t.Fatalf("expected temporary student to appear first in active roster, got %#v", roster.Active)
	}
	if roster.Active[1].StudentID != 3 || roster.Active[1].ScheduleStudentType != model.TeachingScheduleStudentTypeMakeup {
		t.Fatalf("expected makeup student to appear in active roster, got %#v", roster.Active)
	}
}

func TestBuildAssociatedGroupClassStudentSet_IncludesRemovedClassMemberOverrides(t *testing.T) {
	scheduleStartAt := time.Date(2026, 4, 9, 10, 0, 0, 0, time.Local)
	got := buildAssociatedGroupClassStudentSet(
		[]groupClassStudentMembership{
			{
				StudentID:   1,
				StudentName: "正式班课学员",
				JoinAt:      scheduleStartAt.Add(-time.Hour),
			},
		},
		[]teachingScheduleStudentOverride{
			{
				StudentID:    2,
				StudentName:  "本节移除的班课学员",
				StudentType:  model.TeachingScheduleStudentTypeClassMember,
				RosterStatus: model.TeachingScheduleStudentRosterStatusRemoved,
			},
			{
				StudentID:    3,
				StudentName:  "试听学员",
				StudentType:  model.TeachingScheduleStudentTypeTrial,
				RosterStatus: model.TeachingScheduleStudentRosterStatusActive,
			},
		},
		scheduleStartAt,
	)

	if _, ok := got[1]; !ok {
		t.Fatalf("expected studying class member to be marked as associated, got %#v", got)
	}
	if _, ok := got[2]; !ok {
		t.Fatalf("expected removed class-member override to still be marked as associated, got %#v", got)
	}
	if _, ok := got[3]; ok {
		t.Fatalf("expected trial override to not be marked as associated, got %#v", got)
	}
}

func TestCollectGroupClassConflictingStudentNames(t *testing.T) {
	currentRoster := groupClassScheduleRoster{
		Active: []groupClassScheduleStudent{
			{StudentID: 1, StudentName: "学员甲"},
			{StudentID: 2, StudentName: "学员乙"},
			{StudentID: 3, StudentName: "学员丙"},
		},
	}

	got := collectGroupClassConflictingStudentNames(
		currentRoster,
		[]scheduleConflictDetailRow{
			{
				ID:        10,
				ClassType: model.TeachingClassTypeNormal,
			},
			{
				ID:        11,
				ClassType: model.TeachingClassTypeOneToOne,
				StudentID: 3,
			},
		},
		map[int64]groupClassScheduleRoster{
			10: {
				Active: []groupClassScheduleStudent{
					{StudentID: 2, StudentName: "学员乙"},
					{StudentID: 4, StudentName: "其他学员"},
				},
			},
		},
	)

	if len(got) != 2 {
		t.Fatalf("expected two conflicting student names, got %#v", got)
	}
	if got[0] != "学员乙" || got[1] != "学员丙" {
		t.Fatalf("expected conflicting student names to keep current roster order, got %#v", got)
	}
}
