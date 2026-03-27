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
		UPDATE inst_course
		SET sale_volume = 0, update_id = ?, update_time = NOW()
		WHERE inst_id = ? AND del_flag = 0 AND IFNULL(sale_volume, 0) <> 0
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
