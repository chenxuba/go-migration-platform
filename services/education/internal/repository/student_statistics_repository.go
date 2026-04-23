package repository

import (
	"context"
	"math"

	"go-migration-platform/services/education/internal/model"
)

func (repo *Repository) GetStudentOverviewStatistics(ctx context.Context, instID int64) (model.StudentOverviewStatistics, error) {
	result := model.StudentOverviewStatistics{}
	officialSubscribedExpr := studentOfficialSubscribedExistsSQL("s")

	if err := repo.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*) AS total_students,
			IFNULL(SUM(CASE WHEN s.student_status = 1 THEN 1 ELSE 0 END), 0) AS reading_students,
			IFNULL(SUM(CASE WHEN s.student_status = 2 THEN 1 ELSE 0 END), 0) AS history_students,
			IFNULL(SUM(CASE WHEN s.student_status = 0 THEN 1 ELSE 0 END), 0) AS intent_students,
			IFNULL(SUM(CASE WHEN s.create_time >= DATE_SUB(NOW(), INTERVAL 30 DAY) THEN 1 ELSE 0 END), 0) AS recent_month_new_students,
			IFNULL(SUM(CASE WHEN s.create_time >= DATE_SUB(NOW(), INTERVAL 60 DAY) AND s.create_time < DATE_SUB(NOW(), INTERVAL 30 DAY) THEN 1 ELSE 0 END), 0) AS previous_month_new_students,
			IFNULL(SUM(CASE WHEN s.student_status IN (1, 2) AND `+officialSubscribedExpr+` = 0 THEN 1 ELSE 0 END), 0) AS pending_attention_students
		FROM inst_student s
		WHERE s.inst_id = ? AND s.del_flag = 0
	`, instID).Scan(
		&result.TotalStudents,
		&result.ReadingStudents,
		&result.HistoryStudents,
		&result.IntentStudents,
		&result.RecentMonthNewStudents,
		&result.PreviousMonthNewStudents,
		&result.PendingAttentionStudents,
	); err != nil {
		return model.StudentOverviewStatistics{}, err
	}

	switch {
	case result.PreviousMonthNewStudents > 0:
		result.RecentMonthGrowthRate = int(math.Round((float64(result.RecentMonthNewStudents-result.PreviousMonthNewStudents) / float64(result.PreviousMonthNewStudents)) * 100))
	case result.RecentMonthNewStudents > 0:
		result.RecentMonthGrowthRate = 100
	default:
		result.RecentMonthGrowthRate = 0
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

	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM (
			SELECT 1
			FROM tuition_account ta
			INNER JOIN inst_student s ON s.id = ta.student_id AND s.del_flag = 0
			INNER JOIN inst_course ic ON ic.id = ta.course_id AND ic.del_flag = 0
			LEFT JOIN inst_course_quotation icq ON ta.quote_id = icq.id
			WHERE ta.inst_id = ? AND ta.del_flag = 0
			  AND ta.status IN (1, 2)
			  AND ic.teach_method = 1
			GROUP BY s.id, ic.id, ic.teach_method, icq.lesson_model
			HAVING MAX(GREATEST(
				IFNULL(ta.assigned_class, 0),
				CASE WHEN EXISTS (
					SELECT 1
					FROM teaching_class_student tcs
					INNER JOIN teaching_class tc ON tc.id = tcs.teaching_class_id
						AND tc.inst_id = tcs.inst_id
						AND tc.class_type = 1
						AND tc.del_flag = 0
					WHERE tcs.inst_id = ta.inst_id
					  AND tcs.primary_tuition_account_id = ta.id
					  AND tcs.del_flag = 0
					  AND IFNULL(tcs.class_student_status, 1) IN (1, 2)
				) THEN 1 ELSE 0 END
			)) = 0
		) pending_class_students
	`, instID).Scan(&result.PendingClassStudents); err != nil {
		return model.StudentOverviewStatistics{}, err
	}

	pendingRenewalStudents, err := repo.CountPendingRenewalStudents(ctx, instID)
	if err != nil {
		return model.StudentOverviewStatistics{}, err
	}
	result.PendingRenewalStudents = pendingRenewalStudents

	return result, nil
}
