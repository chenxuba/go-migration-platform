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

	if _, err := tx.ExecContext(ctx, `
		UPDATE inst_student_field_value v
		INNER JOIN inst_student s ON s.id = v.student_id
		SET v.del_flag = 1, v.update_id = ?, v.update_time = NOW()
		WHERE s.inst_id = ? AND s.del_flag = 0 AND v.del_flag = 0
	`, operatorID, instID); err != nil {
		return model.CampusDataClearSummary{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE inst_student_record
		SET del_flag = 1
		WHERE inst_id = ? AND del_flag = 0
	`, instID); err != nil {
		return model.CampusDataClearSummary{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE follow_record
		SET del_flag = 1, update_time = NOW()
		WHERE inst_id = ? AND del_flag = 0
	`, instID); err != nil {
		return model.CampusDataClearSummary{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE approval_history h
		INNER JOIN approval_record r ON r.id = h.approval_id
		SET h.del_flag = 1, h.update_id = ?, h.update_time = NOW()
		WHERE r.inst_id = ? AND r.del_flag = 0 AND h.del_flag = 0
	`, operatorID, instID); err != nil {
		return model.CampusDataClearSummary{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE approval_record
		SET del_flag = 1, update_id = ?, update_time = NOW()
		WHERE inst_id = ? AND del_flag = 0
	`, operatorID, instID); err != nil {
		return model.CampusDataClearSummary{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE sale_order_pay_detail pd
		INNER JOIN sale_order so ON so.id = pd.order_id
		SET pd.del_flag = 1, pd.update_id = ?, pd.update_time = NOW()
		WHERE so.inst_id = ? AND so.del_flag = 0 AND pd.del_flag = 0
	`, operatorID, instID); err != nil {
		return model.CampusDataClearSummary{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE sale_order_course_detail d
		INNER JOIN sale_order so ON so.id = d.order_id
		SET d.del_flag = 1, d.update_id = ?, d.update_time = NOW()
		WHERE so.inst_id = ? AND so.del_flag = 0 AND d.del_flag = 0
	`, operatorID, instID); err != nil {
		return model.CampusDataClearSummary{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE tuition_account
		SET del_flag = 1, update_id = ?, update_time = NOW()
		WHERE inst_id = ? AND del_flag = 0
	`, operatorID, instID); err != nil {
		return model.CampusDataClearSummary{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE tuition_account_flow
		SET del_flag = 1, update_id = ?, update_time = NOW()
		WHERE inst_id = ? AND del_flag = 0
	`, operatorID, instID); err != nil {
		return model.CampusDataClearSummary{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE inst_ledger
		SET del_flag = 1, update_id = ?, update_time = NOW()
		WHERE inst_id = ? AND del_flag = 0
	`, operatorID, instID); err != nil {
		return model.CampusDataClearSummary{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE recharge_account_flow
		SET del_flag = 1, update_id = ?, update_time = NOW()
		WHERE inst_id = ? AND del_flag = 0
	`, operatorID, instID); err != nil {
		return model.CampusDataClearSummary{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE recharge_account_student
		SET del_flag = 1, update_id = ?, update_time = NOW()
		WHERE inst_id = ? AND del_flag = 0
	`, operatorID, instID); err != nil {
		return model.CampusDataClearSummary{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE recharge_account
		SET del_flag = 1, update_id = ?, update_time = NOW()
		WHERE inst_id = ? AND del_flag = 0
	`, operatorID, instID); err != nil {
		return model.CampusDataClearSummary{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE intention_student_import_task_record r
		INNER JOIN intention_student_import_task t ON t.id = r.task_id
		SET r.del_flag = 1, r.update_time = NOW()
		WHERE t.inst_id = ? AND t.del_flag = 0 AND r.del_flag = 0
	`, instID); err != nil {
		return model.CampusDataClearSummary{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE intention_student_import_task
		SET del_flag = 1, update_time = NOW()
		WHERE inst_id = ? AND del_flag = 0
	`, instID); err != nil {
		return model.CampusDataClearSummary{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE order_import_task_record r
		INNER JOIN order_import_task t ON t.id = r.task_id
		SET r.del_flag = 1, r.update_time = NOW()
		WHERE t.inst_id = ? AND t.del_flag = 0 AND r.del_flag = 0
	`, instID); err != nil {
		return model.CampusDataClearSummary{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE order_import_task
		SET del_flag = 1, update_time = NOW()
		WHERE inst_id = ? AND del_flag = 0
	`, instID); err != nil {
		return model.CampusDataClearSummary{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE enrolled_student_export_record
		SET del_flag = 1, update_time = NOW()
		WHERE inst_id = ? AND del_flag = 0
	`, instID); err != nil {
		return model.CampusDataClearSummary{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE sale_order
		SET del_flag = 1, update_id = ?, update_time = NOW()
		WHERE inst_id = ? AND del_flag = 0
	`, operatorID, instID); err != nil {
		return model.CampusDataClearSummary{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE inst_student
		SET del_flag = 1, update_id = ?, update_time = NOW()
		WHERE inst_id = ? AND del_flag = 0
	`, operatorID, instID); err != nil {
		return model.CampusDataClearSummary{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE inst_course_property_result cpr
		INNER JOIN inst_course c ON c.id = cpr.course_id
		SET cpr.del_flag = 1, cpr.update_id = ?, cpr.update_time = NOW()
		WHERE c.inst_id = ? AND c.del_flag = 0 AND cpr.del_flag = 0
	`, operatorID, instID); err != nil {
		return model.CampusDataClearSummary{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE inst_course_quotation cq
		INNER JOIN inst_course c ON c.id = cq.course_id
		SET cq.del_flag = 1, cq.update_id = ?, cq.update_time = NOW()
		WHERE c.inst_id = ? AND c.del_flag = 0 AND cq.del_flag = 0
	`, operatorID, instID); err != nil {
		return model.CampusDataClearSummary{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE inst_course_detail cd
		INNER JOIN inst_course c ON c.id = cd.course_id
		SET cd.del_flag = 1, cd.update_id = ?, cd.update_time = NOW()
		WHERE c.inst_id = ? AND c.del_flag = 0 AND cd.del_flag = 0
	`, operatorID, instID); err != nil {
		return model.CampusDataClearSummary{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE product_package_property_result ppr
		INNER JOIN product_package pp ON pp.id = ppr.product_package_id
		SET ppr.del_flag = 1, ppr.update_id = ?, ppr.update_time = NOW()
		WHERE pp.inst_id = ? AND pp.del_flag = 0 AND ppr.del_flag = 0
	`, operatorID, instID); err != nil {
		return model.CampusDataClearSummary{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE product_package_item ppi
		INNER JOIN product_package pp ON pp.id = ppi.product_package_id
		SET ppi.del_flag = 1, ppi.update_id = ?, ppi.update_time = NOW()
		WHERE pp.inst_id = ? AND pp.del_flag = 0 AND ppi.del_flag = 0
	`, operatorID, instID); err != nil {
		return model.CampusDataClearSummary{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE product_package
		SET del_flag = 1, update_id = ?, update_time = NOW()
		WHERE inst_id = ? AND del_flag = 0
	`, operatorID, instID); err != nil {
		return model.CampusDataClearSummary{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE inst_course
		SET del_flag = 1, sale_volume = 0, update_id = ?, update_time = NOW()
		WHERE inst_id = ? AND del_flag = 0
	`, operatorID, instID); err != nil {
		return model.CampusDataClearSummary{}, err
	}

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
	}

	for _, item := range counts {
		if err := tx.QueryRowContext(ctx, item.query, item.args...).Scan(item.target); err != nil {
			return model.CampusDataClearSummary{}, err
		}
	}

	return summary, nil
}
