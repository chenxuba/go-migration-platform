package repository

import (
	"context"
	"database/sql"
	"fmt"

	"go-migration-platform/services/education/internal/model"
)

func (repo *Repository) GetTuitionAccountReadingList(ctx context.Context, instID int64, studentID string) (model.TuitionAccountReadingListResult, error) {
	arrearTuitionExpr := fmt.Sprintf(`
			SUM(CASE
				WHEN IFNULL(ta.total_tuition, 0) <= 0 THEN 0
				WHEN IFNULL(so.is_bad_debt, 0) = 1 THEN 0
				WHEN IFNULL(so.order_status, 0) = %d THEN 0
				WHEN IFNULL(so.order_real_amount, 0) <= 0 THEN 0
				ELSE GREATEST(
					(CASE
						WHEN sod.id IS NOT NULL THEN GREATEST(IFNULL(sod.amount, 0) - IFNULL(sod.share_discount, 0), 0)
						ELSE IFNULL(ta.total_tuition, 0)
					END)
					- (
						IFNULL(pay.paid_amount, 0) * (
							(CASE
								WHEN sod.id IS NOT NULL THEN GREATEST(IFNULL(sod.amount, 0) - IFNULL(sod.share_discount, 0), 0)
								ELSE IFNULL(ta.total_tuition, 0)
							END) / IFNULL(so.order_real_amount, 0)
						)
					),
					0
				)
			END) AS arrear_tuition
	`, model.OrderStatusPendingPayment)

	rows, err := repo.db.QueryContext(ctx, `
		SELECT 
			CAST(MIN(ta.id) AS CHAR) AS id,
			CAST(ic.id AS CHAR) AS lesson_id,
			IFNULL(ic.name, '') AS lesson_name,
			ic.teach_method AS lesson_type,
			icq.lesson_model AS lesson_charging_mode,
			ic.course_type AS lesson_scope,
			SUM(CASE 
				WHEN icq.lesson_model = 3 THEN ta.total_tuition
				WHEN ta.total_quantity > 0 THEN ta.total_quantity 
				ELSE 0 
			END) AS total_quantity,
			SUM(CASE 
				WHEN icq.lesson_model = 3 THEN ta.free_quantity
				WHEN ta.total_quantity = 0 AND ta.free_quantity > 0 THEN ta.free_quantity 
				ELSE 0 
			END) AS total_free_quantity,
			SUM(ta.total_tuition) AS total_tuition,
			`+arrearTuitionExpr+`,
			SUM(CASE 
				WHEN icq.lesson_model = 3 THEN ta.remaining_tuition
				WHEN ta.total_quantity > 0 THEN ta.remaining_quantity 
				ELSE 0 
			END) AS remain_quantity,
			SUM(ta.remaining_tuition) AS tuition,
			SUM(CASE 
				WHEN icq.lesson_model = 3 THEN ta.free_quantity
				WHEN ta.total_quantity = 0 AND ta.free_quantity > 0 THEN ta.remaining_quantity 
				ELSE 0 
			END) AS remain_free_quantity,
			IFNULL(MAX(ta.enable_expire_time), 0) AS enable_expire_time,
			MAX(ta.expire_time) AS expire_time,
			MIN(ta.valid_date) AS valid_date,
			MAX(ta.end_date) AS end_date,
			MAX(ta.create_time) AS actived_at,
			IFNULL(MAX(ta.assigned_class), 0) AS assigned_class,
			MAX(ta.status) AS status,
			MAX(ta.status_change_time) AS change_status_time,
			MAX(ta.plan_suspend_time) AS plan_suspend_time,
			MAX(ta.plan_resume_time) AS plan_resume_time,
			IFNULL(MAX(ta.has_grade_upgrade), 0) AS has_grade_upgrade
		FROM tuition_account ta
		INNER JOIN inst_course ic ON ta.course_id = ic.id AND ic.del_flag = 0
		LEFT JOIN inst_course_quotation icq ON ta.quote_id = icq.id
		LEFT JOIN sale_order so ON ta.order_id = so.id AND so.del_flag = 0
		LEFT JOIN sale_order_course_detail sod ON ta.order_course_detail_id = sod.id AND sod.del_flag = 0
		LEFT JOIN (
			SELECT order_id, SUM(pay_amount) AS paid_amount
			FROM sale_order_pay_detail
			WHERE del_flag = 0
			GROUP BY order_id
		) pay ON pay.order_id = ta.order_id
		WHERE ta.del_flag = 0
			AND ta.inst_id = ?
			AND ta.student_id = ?
		GROUP BY ic.id, ic.name, ic.teach_method, icq.lesson_model, ic.course_type
		ORDER BY MAX(ta.create_time) DESC
	`, instID, studentID)
	if err != nil {
		return model.TuitionAccountReadingListResult{}, err
	}
	defer rows.Close()

	items := make([]model.TuitionAccountReadingItem, 0, 16)
	for rows.Next() {
		var item model.TuitionAccountReadingItem
		var expireTime, validDate, endDate, activedAt, changeStatusTime, planSuspendTime, planResumeTime sql.NullTime
		if err := rows.Scan(
			&item.ID,
			&item.LessonID,
			&item.LessonName,
			&item.LessonType,
			&item.LessonChargingMode,
			&item.LessonScope,
			&item.TotalQuantity,
			&item.TotalFreeQuantity,
			&item.TotalTuition,
			&item.ArrearTuition,
			&item.RemainQuantity,
			&item.Tuition,
			&item.RemainFreeQuantity,
			&item.EnableExpireTime,
			&expireTime,
			&validDate,
			&endDate,
			&activedAt,
			&item.AssignedClass,
			&item.Status,
			&changeStatusTime,
			&planSuspendTime,
			&planResumeTime,
			&item.HasGradeUpgrade,
		); err != nil {
			return model.TuitionAccountReadingListResult{}, err
		}
		item.IsAdjustable = true
		item.ManualSort = false
		if expireTime.Valid {
			t := expireTime.Time
			item.ExpireTime = &t
		}
		if validDate.Valid {
			t := validDate.Time
			item.ValidDate = &t
		}
		if endDate.Valid {
			t := endDate.Time
			item.EndDate = &t
		}
		if activedAt.Valid {
			t := activedAt.Time
			item.ActivedAt = &t
		}
		if changeStatusTime.Valid {
			t := changeStatusTime.Time
			item.ChangeStatusTime = &t
		}
		if planSuspendTime.Valid {
			t := planSuspendTime.Time
			item.PlanSuspendTime = &t
		}
		if planResumeTime.Valid {
			t := planResumeTime.Time
			item.PlanResumeTime = &t
		}
		items = append(items, item)
	}

	return model.TuitionAccountReadingListResult{
		List:  items,
		Total: len(items),
	}, rows.Err()
}
