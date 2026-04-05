package repository

import (
	"context"
	"database/sql"

	"go-migration-platform/services/education/internal/model"
)

func (repo *Repository) ClearCampusBusinessData(ctx context.Context, instID, operatorID int64) (model.CampusDataClearSummary, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return model.CampusDataClearSummary{}, err
	}
	defer tx.Rollback()

	summary, err := repo.countCampusBusinessDataTx(ctx, tx, instID)
	if err != nil {
		return model.CampusDataClearSummary{}, err
	}

	_ = operatorID

	deleteStatements := []string{
		`DELETE FROM teaching_schedule WHERE inst_id = ?`,
		`DELETE FROM teaching_class_teacher WHERE inst_id = ?`,
		`DELETE FROM teaching_class_student WHERE inst_id = ?`,
		`DELETE FROM teaching_class WHERE inst_id = ?`,
		`DELETE FROM inst_student_field_value
			WHERE student_id IN (
				SELECT id FROM (
					SELECT id FROM inst_student WHERE inst_id = ?
				) AS campus_students
			)`,
		`DELETE FROM inst_student_record WHERE inst_id = ?`,
		`DELETE FROM follow_record WHERE inst_id = ?`,
		`DELETE h FROM approval_history h
			INNER JOIN approval_record r ON r.id = h.approval_id
			WHERE r.inst_id = ?`,
		`DELETE FROM approval_record WHERE inst_id = ?`,
		`DELETE FROM tuition_account_flow WHERE inst_id = ?`,
		`DELETE FROM tuition_account WHERE inst_id = ?`,
		`DELETE FROM inst_ledger WHERE inst_id = ?`,
		`DELETE FROM recharge_account_flow WHERE inst_id = ?`,
		`DELETE FROM recharge_account_student WHERE inst_id = ?`,
		`DELETE FROM recharge_account WHERE inst_id = ?`,
		`DELETE r FROM intention_student_import_task_record r
			INNER JOIN intention_student_import_task t ON t.id = r.task_id
			WHERE t.inst_id = ?`,
		`DELETE FROM intention_student_import_task WHERE inst_id = ?`,
		`DELETE r FROM order_import_task_record r
			INNER JOIN order_import_task t ON t.id = r.task_id
			WHERE t.inst_id = ?`,
		`DELETE FROM order_import_task WHERE inst_id = ?`,
		`DELETE r FROM recharge_account_import_task_record r
			INNER JOIN recharge_account_import_task t ON t.id = r.task_id
			WHERE t.inst_id = ?`,
		`DELETE FROM recharge_account_import_task WHERE inst_id = ?`,
		`DELETE FROM enrolled_student_export_record WHERE inst_id = ?`,
		`DELETE d FROM sale_order_course_detail d
			INNER JOIN sale_order so ON so.id = d.order_id
			WHERE so.inst_id = ?`,
		`DELETE pd FROM sale_order_pay_detail pd
			INNER JOIN sale_order so ON so.id = pd.order_id
			WHERE so.inst_id = ?`,
		`DELETE FROM sale_order WHERE inst_id = ?`,
		`DELETE ppr FROM product_package_property_result ppr
			INNER JOIN product_package pp ON pp.id = ppr.product_package_id
			WHERE pp.inst_id = ?`,
		`DELETE ppi FROM product_package_item ppi
			INNER JOIN product_package pp ON pp.id = ppi.product_package_id
			WHERE pp.inst_id = ?`,
		`DELETE FROM product_package WHERE inst_id = ?`,
		`DELETE FROM inst_student WHERE inst_id = ?`,
	}

	for _, query := range deleteStatements {
		if _, err := tx.ExecContext(ctx, query, instID); err != nil {
			return model.CampusDataClearSummary{}, err
		}
	}

	// 保留课程主档与其配置（详情/报价/属性结果/销量），避免清空校区数据后课程基础资料丢失。
	summary.Courses = 0
	summary.CourseDetails = 0
	summary.CourseQuotations = 0
	summary.CoursePropertyResults = 0
	summary.CourseSaleVolumesReset = 0

	if err := tx.Commit(); err != nil {
		return model.CampusDataClearSummary{}, err
	}
	return summary, nil
}

func (repo *Repository) countCampusBusinessDataTx(ctx context.Context, tx *sql.Tx, instID int64) (model.CampusDataClearSummary, error) {
	summary := model.CampusDataClearSummary{}
	counts := []struct {
		target *int
		query  string
		args   []any
	}{
		{
			target: &summary.Students,
			query:  `SELECT COUNT(*) FROM inst_student WHERE inst_id = ? AND del_flag = 0`,
			args:   []any{instID},
		},
		{
			target: &summary.StudentFieldValues,
			query: `
				SELECT COUNT(*)
				FROM inst_student_field_value v
				INNER JOIN inst_student s ON s.id = v.student_id
				WHERE s.inst_id = ? AND s.del_flag = 0 AND v.del_flag = 0
			`,
			args: []any{instID},
		},
		{
			target: &summary.StudentChangeRecords,
			query:  `SELECT COUNT(*) FROM inst_student_record WHERE inst_id = ? AND del_flag = 0`,
			args:   []any{instID},
		},
		{
			target: &summary.FollowRecords,
			query:  `SELECT COUNT(*) FROM follow_record WHERE inst_id = ? AND del_flag = 0`,
			args:   []any{instID},
		},
		{
			target: &summary.Orders,
			query:  `SELECT COUNT(*) FROM sale_order WHERE inst_id = ? AND del_flag = 0`,
			args:   []any{instID},
		},
		{
			target: &summary.OrderCourseDetails,
			query: `
				SELECT COUNT(*)
				FROM sale_order_course_detail d
				INNER JOIN sale_order so ON so.id = d.order_id
				WHERE so.inst_id = ? AND so.del_flag = 0 AND d.del_flag = 0
			`,
			args: []any{instID},
		},
		{
			target: &summary.OrderPaymentDetails,
			query: `
				SELECT COUNT(*)
				FROM sale_order_pay_detail pd
				INNER JOIN sale_order so ON so.id = pd.order_id
				WHERE so.inst_id = ? AND so.del_flag = 0 AND pd.del_flag = 0
			`,
			args: []any{instID},
		},
		{
			target: &summary.Ledgers,
			query:  `SELECT COUNT(*) FROM inst_ledger WHERE inst_id = ? AND del_flag = 0`,
			args:   []any{instID},
		},
		{
			target: &summary.ApprovalRecords,
			query:  `SELECT COUNT(*) FROM approval_record WHERE inst_id = ? AND del_flag = 0`,
			args:   []any{instID},
		},
		{
			target: &summary.ApprovalHistories,
			query: `
				SELECT COUNT(*)
				FROM approval_history h
				INNER JOIN approval_record r ON r.id = h.approval_id
				WHERE r.inst_id = ? AND r.del_flag = 0 AND h.del_flag = 0
			`,
			args: []any{instID},
		},
		{
			target: &summary.TuitionAccounts,
			query:  `SELECT COUNT(*) FROM tuition_account WHERE inst_id = ? AND del_flag = 0`,
			args:   []any{instID},
		},
		{
			target: &summary.TuitionAccountFlows,
			query:  `SELECT COUNT(*) FROM tuition_account_flow WHERE inst_id = ? AND del_flag = 0`,
			args:   []any{instID},
		},
		{
			target: &summary.RechargeAccounts,
			query:  `SELECT COUNT(*) FROM recharge_account WHERE inst_id = ? AND del_flag = 0`,
			args:   []any{instID},
		},
		{
			target: &summary.RechargeAccountStudents,
			query:  `SELECT COUNT(*) FROM recharge_account_student WHERE inst_id = ? AND del_flag = 0`,
			args:   []any{instID},
		},
		{
			target: &summary.RechargeAccountFlows,
			query:  `SELECT COUNT(*) FROM recharge_account_flow WHERE inst_id = ? AND del_flag = 0`,
			args:   []any{instID},
		},
		{
			target: &summary.ImportTasks,
			query:  `SELECT COUNT(*) FROM intention_student_import_task WHERE inst_id = ? AND del_flag = 0`,
			args:   []any{instID},
		},
		{
			target: &summary.ImportTaskRecords,
			query: `
				SELECT COUNT(*)
				FROM intention_student_import_task_record r
				INNER JOIN intention_student_import_task t ON t.id = r.task_id
				WHERE t.inst_id = ? AND t.del_flag = 0 AND r.del_flag = 0
			`,
			args: []any{instID},
		},
		{
			target: &summary.ExportRecords,
			query:  `SELECT COUNT(*) FROM enrolled_student_export_record WHERE inst_id = ? AND del_flag = 0`,
			args:   []any{instID},
		},
		{
			target: &summary.OrderImportTasks,
			query:  `SELECT COUNT(*) FROM order_import_task WHERE inst_id = ? AND del_flag = 0`,
			args:   []any{instID},
		},
		{
			target: &summary.OrderImportTaskRecords,
			query: `
				SELECT COUNT(*)
				FROM order_import_task_record r
				INNER JOIN order_import_task t ON t.id = r.task_id
				WHERE t.inst_id = ? AND t.del_flag = 0 AND r.del_flag = 0
			`,
			args: []any{instID},
		},
		{
			target: &summary.RechargeImportTasks,
			query:  `SELECT COUNT(*) FROM recharge_account_import_task WHERE inst_id = ? AND del_flag = 0`,
			args:   []any{instID},
		},
		{
			target: &summary.RechargeImportTaskRecords,
			query: `
				SELECT COUNT(*)
				FROM recharge_account_import_task_record r
				INNER JOIN recharge_account_import_task t ON t.id = r.task_id
				WHERE t.inst_id = ? AND t.del_flag = 0 AND r.del_flag = 0
			`,
			args: []any{instID},
		},
		{
			target: &summary.Courses,
			query:  `SELECT COUNT(*) FROM inst_course WHERE inst_id = ? AND del_flag = 0`,
			args:   []any{instID},
		},
		{
			target: &summary.CourseDetails,
			query: `
				SELECT COUNT(*)
				FROM inst_course_detail cd
				INNER JOIN inst_course c ON c.id = cd.course_id
				WHERE c.inst_id = ? AND c.del_flag = 0 AND cd.del_flag = 0
			`,
			args: []any{instID},
		},
		{
			target: &summary.CourseQuotations,
			query: `
				SELECT COUNT(*)
				FROM inst_course_quotation cq
				INNER JOIN inst_course c ON c.id = cq.course_id
				WHERE c.inst_id = ? AND c.del_flag = 0 AND cq.del_flag = 0
			`,
			args: []any{instID},
		},
		{
			target: &summary.CoursePropertyResults,
			query: `
				SELECT COUNT(*)
				FROM inst_course_property_result cpr
				INNER JOIN inst_course c ON c.id = cpr.course_id
				WHERE c.inst_id = ? AND c.del_flag = 0 AND cpr.del_flag = 0
			`,
			args: []any{instID},
		},
		{
			target: &summary.ProductPackages,
			query:  `SELECT COUNT(*) FROM product_package WHERE inst_id = ? AND del_flag = 0`,
			args:   []any{instID},
		},
		{
			target: &summary.ProductPackageItems,
			query: `
				SELECT COUNT(*)
				FROM product_package_item ppi
				INNER JOIN product_package pp ON pp.id = ppi.product_package_id
				WHERE pp.inst_id = ? AND pp.del_flag = 0 AND ppi.del_flag = 0
			`,
			args: []any{instID},
		},
		{
			target: &summary.ProductPackageProperties,
			query: `
				SELECT COUNT(*)
				FROM product_package_property_result ppr
				INNER JOIN product_package pp ON pp.id = ppr.product_package_id
				WHERE pp.inst_id = ? AND pp.del_flag = 0 AND ppr.del_flag = 0
			`,
			args: []any{instID},
		},
		{
			target: &summary.CourseSaleVolumesReset,
			query:  `SELECT COUNT(*) FROM inst_course WHERE inst_id = ? AND del_flag = 0 AND IFNULL(sale_volume, 0) <> 0`,
			args:   []any{instID},
		},
		{
			target: &summary.TeachingClasses,
			query:  `SELECT COUNT(*) FROM teaching_class WHERE inst_id = ? AND del_flag = 0`,
			args:   []any{instID},
		},
		{
			target: &summary.TeachingClassStudents,
			query:  `SELECT COUNT(*) FROM teaching_class_student WHERE inst_id = ? AND del_flag = 0`,
			args:   []any{instID},
		},
		{
			target: &summary.TeachingClassTeachers,
			query:  `SELECT COUNT(*) FROM teaching_class_teacher WHERE inst_id = ? AND del_flag = 0`,
			args:   []any{instID},
		},
		{
			target: &summary.TeachingSchedules,
			query:  `SELECT COUNT(*) FROM teaching_schedule WHERE inst_id = ? AND del_flag = 0`,
			args:   []any{instID},
		},
	}

	for _, item := range counts {
		if err := tx.QueryRowContext(ctx, item.query, item.args...).Scan(item.target); err != nil {
			return model.CampusDataClearSummary{}, err
		}
	}

	return summary, nil
}
