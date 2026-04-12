package service

import (
	"database/sql/driver"
	"testing"
	"time"

	"go-migration-platform/services/education/internal/model"
)

func TestListTeachingSchedulesByTeacherMatrix_IncludesAssistantMatchesInTeacherFilter(t *testing.T) {
	userID := int64(501)
	instID := int64(601)
	assistantTeacherID := int64(200)
	lessonDate := time.Date(2026, 4, 9, 0, 0, 0, 0, time.Local)
	startAt := time.Date(2026, 4, 9, 10, 50, 0, 0, time.Local)
	endAt := time.Date(2026, 4, 9, 11, 35, 0, 0, time.Local)

	svc, cleanup := newScriptedService(t, []queryExpectation{
		findInstIDExpectation(userID, instID),
		{
			query: `
				SELECT id,
					COALESCE(NULLIF(TRIM(nick_name), ''), NULLIF(TRIM(username), ''), '') AS display_name
				FROM inst_user
				WHERE inst_id = ? AND del_flag = 0 AND disabled = 0
				ORDER BY id ASC
			`,
			args:    []any{instID},
			columns: []string{"id", "display_name"},
			rows: [][]driver.Value{
				{assistantTeacherID, "李文亮"},
			},
		},
		{
			query: `
				SELECT
					id,
					IFNULL(batch_no, ''),
					IFNULL(batch_size, 1),
					IFNULL(class_type, 0),
					IFNULL(teaching_class_id, 0),
					IFNULL(teaching_class_name, ''),
					IFNULL(student_id, 0),
					IFNULL(student_name, ''),
					IFNULL(lesson_id, 0),
					IFNULL(lesson_name, ''),
					IFNULL(teacher_id, 0),
					IFNULL(teacher_name, ''),
					assistant_ids_json,
					assistant_names_json,
					IFNULL(classroom_id, 0),
					IFNULL(classroom_name, ''),
					lesson_date,
					lesson_start_at,
					lesson_end_at,
					IFNULL(status, 0)
				FROM teaching_schedule ts
				WHERE ts.inst_id = ? AND ts.del_flag = 0 AND ts.status = ? AND ts.lesson_date >= ? AND ts.lesson_date <= ? AND ts.class_type = ? AND (ts.teacher_id IN (?) OR JSON_SEARCH(COALESCE(ts.assistant_ids_json, JSON_ARRAY()), 'one', ?) IS NOT NULL)
				ORDER BY ts.lesson_start_at ASC, ts.id ASC
			`,
			args: []any{
				instID,
				model.TeachingScheduleStatusActive,
				"2026-04-06",
				"2026-04-12",
				model.TeachingClassTypeOneToOne,
				assistantTeacherID,
				"200",
			},
			columns: []string{
				"id",
				"batch_no",
				"batch_size",
				"class_type",
				"teaching_class_id",
				"teaching_class_name",
				"student_id",
				"student_name",
				"lesson_id",
				"lesson_name",
				"teacher_id",
				"teacher_name",
				"assistant_ids_json",
				"assistant_names_json",
				"classroom_id",
				"classroom_name",
				"lesson_date",
				"lesson_start_at",
				"lesson_end_at",
				"status",
			},
			rows: [][]driver.Value{
				{
					int64(301),
					"",
					int64(1),
					int64(model.TeachingClassTypeOneToOne),
					int64(701),
					"王安全-时段课程",
					int64(801),
					"王安全",
					int64(901),
					"时段课程",
					int64(999),
					"主教老师",
					[]byte(`["200"]`),
					[]byte(`["李文亮"]`),
					int64(0),
					"",
					lessonDate,
					startAt,
					endAt,
					int64(model.TeachingScheduleStatusActive),
				},
			},
		},
		{
			query: `
				SELECT
					id,
					IFNULL(batch_no, ''),
					IFNULL(batch_size, 1),
					IFNULL(class_type, 0),
					IFNULL(teaching_class_id, 0),
					IFNULL(teaching_class_name, ''),
					IFNULL(student_id, 0),
					IFNULL(student_name, ''),
					IFNULL(lesson_id, 0),
					IFNULL(lesson_name, ''),
					IFNULL(teacher_id, 0),
					IFNULL(teacher_name, ''),
					assistant_ids_json,
					assistant_names_json,
					IFNULL(classroom_id, 0),
					IFNULL(classroom_name, ''),
					lesson_date,
					lesson_start_at,
					lesson_end_at,
					IFNULL(status, 0)
				FROM teaching_schedule ts
				WHERE ts.inst_id = ? AND ts.del_flag = 0 AND ts.status = ? AND ts.lesson_date >= ? AND ts.lesson_date <= ?
				ORDER BY ts.lesson_start_at ASC, ts.id ASC
			`,
			args: []any{
				instID,
				model.TeachingScheduleStatusActive,
				"2026-04-06",
				"2026-04-12",
			},
			columns: []string{
				"id",
				"batch_no",
				"batch_size",
				"class_type",
				"teaching_class_id",
				"teaching_class_name",
				"student_id",
				"student_name",
				"lesson_id",
				"lesson_name",
				"teacher_id",
				"teacher_name",
				"assistant_ids_json",
				"assistant_names_json",
				"classroom_id",
				"classroom_name",
				"lesson_date",
				"lesson_start_at",
				"lesson_end_at",
				"status",
			},
			rows: [][]driver.Value{
				{
					int64(301),
					"",
					int64(1),
					int64(model.TeachingClassTypeOneToOne),
					int64(701),
					"王安全-时段课程",
					int64(801),
					"王安全",
					int64(901),
					"时段课程",
					int64(999),
					"主教老师",
					[]byte(`["200"]`),
					[]byte(`["李文亮"]`),
					int64(0),
					"",
					lessonDate,
					startAt,
					endAt,
					int64(model.TeachingScheduleStatusActive),
				},
			},
		},
	})
	defer cleanup()

	matrix, err := svc.ListTeachingSchedulesByTeacherMatrix(userID, model.TeachingScheduleListQueryDTO{
		StartDate:          "2026-04-06",
		EndDate:            "2026-04-12",
		ClassType:          matrixIntPtr(model.TeachingClassTypeOneToOne),
		ScheduleTeacherIDs: []int64{assistantTeacherID},
	})
	if err != nil {
		t.Fatalf("ListTeachingSchedulesByTeacherMatrix returned error: %v", err)
	}
	if len(matrix) != 7 {
		t.Fatalf("expected 7 matrix days, got %d", len(matrix))
	}

	var targetDay *model.TeachingScheduleMatrixDayVO
	for i := range matrix {
		if matrix[i].ScheduleDate == "2026-04-09" {
			targetDay = &matrix[i]
			break
		}
	}
	if targetDay == nil {
		t.Fatalf("expected to find matrix day for 2026-04-09")
	}
	if len(targetDay.ScheduleListVoList) != 1 {
		t.Fatalf("expected one teacher column on target day, got %d", len(targetDay.ScheduleListVoList))
	}
	if targetDay.ScheduleListVoList[0].TeacherID != assistantTeacherID {
		t.Fatalf("expected teacher %d, got %d", assistantTeacherID, targetDay.ScheduleListVoList[0].TeacherID)
	}
	if len(targetDay.ScheduleListVoList[0].ScheduleInfoVoList) != 1 {
		t.Fatalf("expected assistant teacher row to include matched schedule, got %d items", len(targetDay.ScheduleListVoList[0].ScheduleInfoVoList))
	}
}

func TestListTeachingSchedulesByTeacherMatrix_IncludesDisabledTeacherFromPeriodGroupAllowList(t *testing.T) {
	userID := int64(511)
	instID := int64(611)
	disabledTeacherID := int64(300)
	groupUUID := "group-c"

	svc, cleanup := newScriptedService(t, []queryExpectation{
		findInstIDExpectation(userID, instID),
		{
			query: `
				SELECT id,
					COALESCE(NULLIF(TRIM(nick_name), ''), NULLIF(TRIM(username), ''), '') AS display_name
				FROM inst_user
				WHERE inst_id = ? AND del_flag = 0 AND disabled = 0
				ORDER BY id ASC
			`,
			args:    []any{instID},
			columns: []string{"id", "display_name"},
			rows:    [][]driver.Value{},
		},
		{
			query: `
				SELECT
					id,
					IFNULL(batch_no, ''),
					IFNULL(batch_size, 1),
					IFNULL(class_type, 0),
					IFNULL(teaching_class_id, 0),
					IFNULL(teaching_class_name, ''),
					IFNULL(student_id, 0),
					IFNULL(student_name, ''),
					IFNULL(lesson_id, 0),
					IFNULL(lesson_name, ''),
					IFNULL(teacher_id, 0),
					IFNULL(teacher_name, ''),
					assistant_ids_json,
					assistant_names_json,
					IFNULL(classroom_id, 0),
					IFNULL(classroom_name, ''),
					lesson_date,
					lesson_start_at,
					lesson_end_at,
					IFNULL(status, 0)
				FROM teaching_schedule ts
				WHERE ts.inst_id = ? AND ts.del_flag = 0 AND ts.status = ? AND ts.lesson_date >= ? AND ts.lesson_date <= ? AND ts.class_type = ?
				ORDER BY ts.lesson_start_at ASC, ts.id ASC
			`,
			args: []any{
				instID,
				model.TeachingScheduleStatusActive,
				"2026-04-06",
				"2026-04-12",
				model.TeachingClassTypeOneToOne,
			},
			columns: []string{
				"id",
				"batch_no",
				"batch_size",
				"class_type",
				"teaching_class_id",
				"teaching_class_name",
				"student_id",
				"student_name",
				"lesson_id",
				"lesson_name",
				"teacher_id",
				"teacher_name",
				"assistant_ids_json",
				"assistant_names_json",
				"classroom_id",
				"classroom_name",
				"lesson_date",
				"lesson_start_at",
				"lesson_end_at",
				"status",
			},
			rows: [][]driver.Value{},
		},
		{
			query: `
				SELECT t.teacher_user_id
				FROM inst_period_group_teacher t
				INNER JOIN inst_period_group g ON g.id = t.group_id AND g.inst_id = ? AND g.del_flag = 0 AND g.group_uuid = ?
				ORDER BY t.id ASC
			`,
			args:    []any{instID, groupUUID},
			columns: []string{"teacher_user_id"},
			rows: [][]driver.Value{
				{disabledTeacherID},
			},
		},
		{
			query: `
				SELECT id,
					COALESCE(NULLIF(TRIM(nick_name), ''), NULLIF(TRIM(username), ''), '') AS display_name,
					IFNULL(disabled, 0) AS disabled
				FROM inst_user
				WHERE inst_id = ? AND del_flag = 0 AND id IN (?)
				ORDER BY id ASC
			`,
			args:    []any{instID, disabledTeacherID},
			columns: []string{"id", "display_name", "disabled"},
			rows: [][]driver.Value{
				{disabledTeacherID, "孙悟空", true},
			},
		},
	})
	defer cleanup()

	matrix, err := svc.ListTeachingSchedulesByTeacherMatrix(userID, model.TeachingScheduleListQueryDTO{
		StartDate:       "2026-04-06",
		EndDate:         "2026-04-12",
		ClassType:       matrixIntPtr(model.TeachingClassTypeOneToOne),
		PeriodGroupUUID: groupUUID,
		MatrixTeacherIDs: []int64{
			disabledTeacherID,
		},
	})
	if err != nil {
		t.Fatalf("ListTeachingSchedulesByTeacherMatrix returned error: %v", err)
	}
	if len(matrix) != 7 {
		t.Fatalf("expected 7 matrix days, got %d", len(matrix))
	}
	for _, day := range matrix {
		if len(day.ScheduleListVoList) != 1 {
			t.Fatalf("expected one teacher column on %s, got %d", day.ScheduleDate, len(day.ScheduleListVoList))
		}
		if day.ScheduleListVoList[0].TeacherID != disabledTeacherID {
			t.Fatalf("expected disabled teacher %d on %s, got %d", disabledTeacherID, day.ScheduleDate, day.ScheduleListVoList[0].TeacherID)
		}
		if day.ScheduleListVoList[0].TeacherName != "孙悟空（离职）" {
			t.Fatalf("expected disabled teacher display name with suffix, got %q", day.ScheduleListVoList[0].TeacherName)
		}
	}
}

func TestMapTeachingScheduleToLegacyVO_DoesNotFabricateGroupClassStudentsFromNameOnly(t *testing.T) {
	startAt := time.Date(2026, 4, 8, 10, 5, 0, 0, time.Local)
	endAt := time.Date(2026, 4, 8, 10, 45, 0, 0, time.Local)

	legacy := mapTeachingScheduleToLegacyVO(model.TeachingScheduleVO{
		ID:                "1",
		ClassType:         model.TeachingClassTypeNormal,
		TeachingClassID:   "101",
		TeachingClassName: "认知班课",
		StudentID:         "",
		StudentName:       "王认知",
		LessonID:          "201",
		LessonName:        "班级认知课",
		TeacherID:         "301",
		TeacherName:       "许晶晶",
		LessonDate:        "2026-04-08",
		StartAt:           startAt,
		EndAt:             endAt,
		Status:            model.TeachingScheduleStatusActive,
		CallStatus:        1,
	}, 1)

	if len(legacy.StudentList) != 0 {
		t.Fatalf("expected group class legacy mapping to keep empty student list when only stale studentName exists, got %#v", legacy.StudentList)
	}
}

func matrixIntPtr(v int) *int {
	return &v
}
