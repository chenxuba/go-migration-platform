package service

import (
	"database/sql/driver"
	"testing"
	"time"
)

func mondayOfWeek(value time.Time) time.Time {
	t := time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return t.AddDate(0, 0, -(weekday - 1))
}

func TestPreviewInstPeriodConfigUpdate_UsesCurrentWeekPayloadForAffectedTeachers(t *testing.T) {
	userID := int64(901)
	instID := int64(902)
	teacherID := int64(22)
	now := time.Now()
	currentWeekStart := mondayOfWeek(now)
	currentWeekEnd := currentWeekStart.AddDate(0, 0, 6)
	nextWeekStart := currentWeekStart.AddDate(0, 0, 7)
	nextWeekEnd := nextWeekStart.AddDate(0, 0, 6)

	svc, cleanup := newScriptedService(t, []queryExpectation{
		findInstIDExpectation(userID, instID),
		{
			query: `
				SELECT COUNT(*)
				FROM inst_period_config_version
				WHERE inst_id = ?
			`,
			args:    []any{instID},
			columns: []string{"count"},
			rows: [][]driver.Value{
				{int64(1)},
			},
		},
		{
			query: `
				SELECT payload_json
				FROM inst_period_config_version
				WHERE inst_id = ? AND effective_week_start <= ?
				ORDER BY effective_week_start DESC, id DESC
				LIMIT 1
			`,
			args:    []any{instID, currentWeekStart.Format("2006-01-02")},
			columns: []string{"payload_json"},
			rows: [][]driver.Value{
				{`{"version":1,"groups":[{"id":"group-c","name":"C组","sort":2,"slots":[{"index":1,"start":"09:20","end":"10:00","enabled":true}],"boundTeachers":[{"id":"22","name":"许晶晶"}]}]}`},
			},
		},
		{
			query: `
				SELECT id, IFNULL(nick_name, ''), IFNULL(disabled, 0)
				FROM inst_user
				WHERE inst_id = ? AND del_flag = 0 AND id IN (?)
			`,
			args:    []any{instID, teacherID},
			columns: []string{"id", "nick_name", "disabled"},
			rows: [][]driver.Value{
				{teacherID, "许晶晶", false},
			},
		},
		{
			query: `
				SELECT COUNT(*)
				FROM teaching_schedule
				WHERE inst_id = ?
				  AND del_flag = 0
				  AND status = 1
				  AND lesson_date >= ?
				  AND lesson_date <= ?
				  AND (
					teacher_id IN (?)
					OR JSON_SEARCH(COALESCE(assistant_ids_json, JSON_ARRAY()), 'one', ?) IS NOT NULL
				  )
			`,
			args: []any{
				instID,
				currentWeekStart.Format("2006-01-02"),
				currentWeekEnd.Format("2006-01-02"),
				teacherID,
				"22",
			},
			columns: []string{"count"},
			rows: [][]driver.Value{
				{int64(1)},
			},
		},
		{
			query: `
				SELECT COUNT(*)
				FROM teaching_schedule
				WHERE inst_id = ?
				  AND del_flag = 0
				  AND status = 1
				  AND lesson_date >= ?
				  AND lesson_date <= ?
				  AND (
					teacher_id IN (?)
					OR JSON_SEARCH(COALESCE(assistant_ids_json, JSON_ARRAY()), 'one', ?) IS NOT NULL
				  )
			`,
			args: []any{
				instID,
				nextWeekStart.Format("2006-01-02"),
				nextWeekEnd.Format("2006-01-02"),
				teacherID,
				"22",
			},
			columns: []string{"count"},
			rows: [][]driver.Value{
				{int64(0)},
			},
		},
	})
	defer cleanup()

	result, err := svc.PreviewInstPeriodConfigUpdate(userID, map[string]any{
		"version": 1,
		"groups": []map[string]any{
			{
				"id":            "group-c",
				"name":          "C组",
				"sort":          2,
				"slots":         []map[string]any{{"index": 1, "start": "09:20", "end": "10:00", "enabled": true}},
				"boundTeachers": []map[string]any{},
			},
		},
	})
	if err != nil {
		t.Fatalf("PreviewInstPeriodConfigUpdate returned error: %v", err)
	}
	if result.PeriodAppliedToday {
		t.Fatalf("expected current-week removal with classes to defer, got applied today")
	}
	if result.PeriodWeekStart != nextWeekStart.Format("2006-01-02") {
		t.Fatalf("expected deferred week start %s, got %s", nextWeekStart.Format("2006-01-02"), result.PeriodWeekStart)
	}
}
