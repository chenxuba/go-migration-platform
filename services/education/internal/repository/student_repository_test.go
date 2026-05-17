package repository

import (
	"context"
	"testing"
)

func TestBatchDeleteIntentStudentsClearsRelatedBusinessData(t *testing.T) {
	countTempStudentIDsQuery := `SELECT COUNT(*) FROM tmp_delete_student_ids`
	db, state, cleanup := newCampusClearScriptDB(t, map[string]int64{
		normalizeCampusClearSQL(countTempStudentIDsQuery): 1,
	})
	defer cleanup()

	repo := &Repository{db: db}
	if err := repo.BatchDeleteIntentStudents(context.Background(), 10048, []int64{101, 102}); err != nil {
		t.Fatalf("BatchDeleteIntentStudents returned error: %v", err)
	}

	requiredDeletes := []string{
		`DELETE sod FROM sale_order_course_detail sod
			INNER JOIN tmp_delete_student_order_ids do ON do.id = sod.order_id`,
		`DELETE sop FROM sale_order_pay_detail sop
			INNER JOIN tmp_delete_student_order_ids do ON do.id = sop.order_id`,
		`DELETE so FROM sale_order so
			INNER JOIN tmp_delete_student_order_ids do ON do.id = so.id
			WHERE so.inst_id = ?`,
		`DELETE taf FROM tuition_account_flow taf
			INNER JOIN tmp_delete_student_ids ds ON ds.id = taf.student_id
			WHERE taf.inst_id = ?`,
		`DELETE str FROM student_teaching_record str
			INNER JOIN tmp_delete_student_ids ds ON ds.id = str.student_id
			WHERE str.inst_id = ?`,
		`DELETE tss FROM teaching_schedule_student tss
			INNER JOIN tmp_delete_student_ids ds ON ds.id = tss.student_id
			WHERE tss.inst_id = ?`,
		`DELETE ts FROM teaching_schedule ts
			INNER JOIN tmp_delete_student_ids ds ON ds.id = ts.student_id
			WHERE ts.inst_id = ?`,
		`DELETE tcs FROM teaching_class_student tcs
			INNER JOIN tmp_delete_student_ids ds ON ds.id = tcs.student_id
			WHERE tcs.inst_id = ?`,
		`DELETE s FROM inst_student s
			INNER JOIN tmp_delete_student_ids ds ON ds.id = s.id
			WHERE s.inst_id = ?`,
	}
	for _, query := range requiredDeletes {
		if !containsNormalizedCampusClearSQL(state.execLog, query) {
			t.Fatalf("expected delete query to be executed: %s", normalizeCampusClearSQL(query))
		}
	}

	legacySoftDelete := `
		UPDATE inst_student
		SET del_flag = 1
		WHERE id IN (?,?)
		  AND inst_id = ?
		  AND del_flag = 0`
	if containsNormalizedCampusClearSQL(state.execLog, legacySoftDelete) {
		t.Fatalf("did not expect legacy soft delete query to be executed")
	}
}
