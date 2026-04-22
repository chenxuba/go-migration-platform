package service

import (
	"context"
	"database/sql/driver"
	"testing"
)

func TestRefreshStudentBindChildStatusByPhones_DeduplicatesAndRefreshesEachPhone(t *testing.T) {
	phoneA := "17601241636"
	phoneB := "17601241632"

	svc, cleanup := newScriptedService(t, []queryExpectation{
		{
			query: `
				SELECT COUNT(*)
				FROM wechat_official_user_link
				WHERE phone = ? AND subscribed = 1
			`,
			args:    []any{phoneA},
			columns: []string{"count"},
			rows:    [][]driver.Value{{int64(0)}},
		},
		{
			query: `
				SELECT COUNT(*)
				FROM wechat_official_student_binding
				WHERE phone = ? AND subscribed = 1
			`,
			args:    []any{phoneA},
			columns: []string{"count"},
			rows:    [][]driver.Value{{int64(0)}},
		},
		execResultExpectation(`
			UPDATE inst_student
			SET is_bind_child = ?, update_time = NOW()
			WHERE mobile = ? AND del_flag = 0
		`, []any{0, phoneA}, 1),
		{
			query: `
				SELECT COUNT(*)
				FROM wechat_official_user_link
				WHERE phone = ? AND subscribed = 1
			`,
			args:    []any{phoneB},
			columns: []string{"count"},
			rows:    [][]driver.Value{{int64(1)}},
		},
		{
			query: `
				SELECT COUNT(*)
				FROM wechat_official_student_binding
				WHERE phone = ? AND subscribed = 1
			`,
			args:    []any{phoneB},
			columns: []string{"count"},
			rows:    [][]driver.Value{{int64(0)}},
		},
		execResultExpectation(`
			UPDATE inst_student
			SET is_bind_child = ?, update_time = NOW()
			WHERE mobile = ? AND del_flag = 0
		`, []any{1, phoneB}, 1),
	})
	defer cleanup()

	svc.refreshStudentBindChildStatusByPhones(context.Background(), "", phoneA, phoneA, "  ", phoneB)
}
