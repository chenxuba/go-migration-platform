package repository

import (
	"context"

	"go-migration-platform/services/education/internal/model"
)

func (repo *Repository) GetStudentOverviewStatistics(ctx context.Context, instID int64) (model.StudentOverviewStatistics, error) {
	result := model.StudentOverviewStatistics{}

	if err := repo.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*) AS total_students,
			IFNULL(SUM(CASE WHEN student_status = 1 THEN 1 ELSE 0 END), 0) AS reading_students,
			IFNULL(SUM(CASE WHEN student_status = 2 THEN 1 ELSE 0 END), 0) AS history_students,
			IFNULL(SUM(CASE WHEN student_status = 0 THEN 1 ELSE 0 END), 0) AS intent_students
		FROM inst_student
		WHERE inst_id = ? AND del_flag = 0
	`, instID).Scan(&result.TotalStudents, &result.ReadingStudents, &result.HistoryStudents, &result.IntentStudents); err != nil {
		return model.StudentOverviewStatistics{}, err
	}

	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM (
			SELECT DISTINCT so.student_id
			FROM sale_order so
			INNER JOIN inst_student s ON s.id = so.student_id AND s.del_flag = 0
			WHERE so.inst_id = ? AND so.del_flag = 0
			  AND IFNULL(so.is_bad_debt, 0) = 0
			  AND so.order_type = ?
			  AND so.order_status <> ?
			  AND IFNULL(so.order_real_amount, 0) > (
				SELECT IFNULL(SUM(pd.pay_amount), 0)
				FROM sale_order_pay_detail pd
				WHERE pd.del_flag = 0 AND pd.order_id = so.id
			  )
			UNION
			SELECT DISTINCT str.student_id
			FROM student_teaching_record str
			INNER JOIN inst_student s ON s.id = str.student_id AND s.del_flag = 0
			WHERE str.inst_id = ? AND str.del_flag = 0 AND IFNULL(str.arrear_quantity, 0) > 0
		) AS arrear_students
	`, instID, model.OrderTypeRegistrationRenewal, model.OrderStatusPendingPayment, instID).Scan(&result.ArrearStudents); err != nil {
		return model.StudentOverviewStatistics{}, err
	}

	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM inst_student s
		WHERE s.inst_id = ? AND s.del_flag = 0 AND s.birthday IS NOT NULL
		  AND TIMESTAMPDIFF(
			DAY,
			CURDATE(),
			STR_TO_DATE(
				CONCAT(
					YEAR(CURDATE()) + (DATE_FORMAT(s.birthday, '%m-%d') < DATE_FORMAT(CURDATE(), '%m-%d')),
					'-',
					DATE_FORMAT(s.birthday, '%m-%d')
				),
				'%Y-%m-%d'
			)
		  ) BETWEEN 0 AND 30
	`, instID).Scan(&result.BirthdayStudents); err != nil {
		return model.StudentOverviewStatistics{}, err
	}

	return result, nil
}
