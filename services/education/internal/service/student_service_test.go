package service

import (
	"context"
	"testing"
)

func TestRefreshStudentBindChildStatusByPhones_DeduplicatesAndRefreshesEachPhone(t *testing.T) {
	phoneA := "17601241636"
	phoneB := "17601241632"

	svc, cleanup := newScriptedService(t, []queryExpectation{
		execResultExpectation(`
			UPDATE inst_student s
			LEFT JOIN (
				SELECT DISTINCT student_id
				FROM wechat_official_student_binding
				WHERE subscribed = 1
			) bound ON bound.student_id = s.id
			SET s.is_bind_child = CASE WHEN bound.student_id IS NULL THEN 0 ELSE 1 END,
				s.update_time = NOW()
			WHERE IFNULL(s.mobile, '') = ? AND s.del_flag = 0
		`, []any{phoneA}, 1),
		execResultExpectation(`
			UPDATE inst_student s
			LEFT JOIN (
				SELECT DISTINCT student_id
				FROM wechat_official_student_binding
				WHERE subscribed = 1
			) bound ON bound.student_id = s.id
			SET s.is_bind_child = CASE WHEN bound.student_id IS NULL THEN 0 ELSE 1 END,
				s.update_time = NOW()
			WHERE IFNULL(s.mobile, '') = ? AND s.del_flag = 0
		`, []any{phoneB}, 1),
	})
	defer cleanup()

	svc.refreshStudentBindChildStatusByPhones(context.Background(), "", phoneA, phoneA, "  ", phoneB)
}
