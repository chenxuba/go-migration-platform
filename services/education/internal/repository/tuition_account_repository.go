package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

func (repo *Repository) CountStudentPrimaryCourseItems(ctx context.Context, instID, studentID int64) (int, error) {
	var count int
	err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM (
			SELECT 1
			FROM tuition_account ta
			INNER JOIN inst_course ic ON ta.course_id = ic.id AND ic.del_flag = 0
			LEFT JOIN inst_course_quotation icq ON ta.quote_id = icq.id
			WHERE ta.del_flag = 0
				AND ta.inst_id = ?
				AND ta.student_id = ?
			GROUP BY ic.id, ic.teach_method, icq.lesson_model, ic.course_type
		) course_items
	`, instID, studentID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

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

// ListStudentTuitionAccountsByStudentAndLesson 学员在某课程下的学费账户（原始账户行）。orderCourseDetailID>0 时仅返回该订单明细下的账户，供 1 对 1 详情报读明细等与当前报读对齐。
func (repo *Repository) ListStudentTuitionAccountsByStudentAndLesson(ctx context.Context, instID, studentID, courseID int64, orderCourseDetailID int64) ([]model.StudentLessonTuitionAccountItem, error) {
	orderDetailSQL := ""
	args := []any{instID, studentID, courseID}
	if orderCourseDetailID > 0 {
		orderDetailSQL = " AND ta.order_course_detail_id = ?"
		args = append(args, orderCourseDetailID)
	}
	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			CAST(ta.id AS CHAR),
			CAST(ta.student_id AS CHAR),
			CAST(ta.course_id AS CHAR),
			IFNULL(NULLIF(TRIM(ic.name), ''), IFNULL(icq.name, '')) AS product_name,
			IFNULL(icq.lesson_model, 0) AS lesson_charging_mode,
			CASE
				WHEN IFNULL(icq.lesson_model, 0) IN (3, 4) THEN IFNULL(ta.total_tuition, 0)
				WHEN IFNULL(ta.total_quantity, 0) > 0 THEN IFNULL(ta.total_quantity, 0)
				ELSE 0
			END AS total_quantity_display,
			CASE
				WHEN IFNULL(icq.lesson_model, 0) IN (3, 4) THEN IFNULL(ta.free_quantity, 0)
				WHEN IFNULL(ta.total_quantity, 0) = 0 AND IFNULL(ta.free_quantity, 0) > 0 THEN IFNULL(ta.free_quantity, 0)
				ELSE 0
			END AS total_free_quantity_display,
			IFNULL(ta.total_tuition, 0),
			CASE
				WHEN IFNULL(ta.total_quantity, 0) = 0 AND IFNULL(ta.free_quantity, 0) > 0 THEN IFNULL(ta.remaining_quantity, 0)
				ELSE 0
			END AS remain_free_quantity,
			CASE
				WHEN IFNULL(icq.lesson_model, 0) IN (3, 4) THEN IFNULL(ta.remaining_tuition, 0)
				WHEN IFNULL(ta.total_quantity, 0) > 0 THEN IFNULL(ta.remaining_quantity, 0)
				ELSE 0
			END AS remain_quantity_display,
			IFNULL(ta.remaining_tuition, 0),
			ta.suspended_time,
			IFNULL(ta.create_time, NOW()) AS start_time,
			IFNULL(ta.enable_expire_time, 0) AS enable_expire,
			ta.expire_time,
			IFNULL(ta.assigned_class, 0) AS assigned_class,
			IFNULL(ic.course_type, 0) AS lesson_scope,
			ta.valid_date,
			IFNULL(ic.teach_method, 0) AS teach_method,
			IFNULL(ta.status, 0) AS ta_status
		FROM tuition_account ta
		INNER JOIN inst_course ic ON ic.id = ta.course_id AND ic.del_flag = 0
		LEFT JOIN sale_order_course_detail sod ON sod.id = ta.order_course_detail_id AND sod.del_flag = 0
		LEFT JOIN inst_course_quotation icq ON icq.id = COALESCE(
			NULLIF(ta.quote_id, 0),
			NULLIF(sod.quote_id, 0),
			(SELECT qx.id FROM inst_course_quotation qx
			 WHERE qx.course_id = ta.course_id AND qx.del_flag = 0
			   AND ABS(IFNULL(qx.quantity, 0) - IFNULL(ta.total_quantity, 0)) < 0.000001
			   AND ABS(IFNULL(qx.price, 0) - IFNULL(ta.total_tuition, 0)) < 0.000001
			 ORDER BY qx.id DESC LIMIT 1),
			(SELECT qmin.id FROM inst_course_quotation qmin
			 WHERE qmin.course_id = ta.course_id AND qmin.del_flag = 0
			 ORDER BY qmin.id ASC LIMIT 1)
		) AND icq.del_flag = 0
		WHERE ta.inst_id = ?
			AND ta.del_flag = 0
			AND ta.student_id = ?
			AND ta.course_id = ?
		`+orderDetailSQL+`
		ORDER BY ta.create_time DESC, ta.id DESC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]model.StudentLessonTuitionAccountItem, 0, 8)
	for rows.Next() {
		var (
			item                                               model.StudentLessonTuitionAccountItem
			suspendedTime                                      sql.NullTime
			expireTime                                         sql.NullTime
			validDate                                          sql.NullTime
			startTime                                          sql.NullTime
			enableExpireRaw                                    int64
			assignedClassRaw                                   int64
			totalQtyDisp, totalFreeDisp, remainFree, remainQty float64
			taStatus                                           int
		)
		if err := rows.Scan(
			&item.ID,
			&item.StudentID,
			&item.LessonID,
			&item.ProductName,
			&item.LessonChargingMode,
			&totalQtyDisp,
			&totalFreeDisp,
			&item.TotalTuition,
			&remainFree,
			&remainQty,
			&item.Tuition,
			&suspendedTime,
			&startTime,
			&enableExpireRaw,
			&expireTime,
			&assignedClassRaw,
			&item.LessonScope,
			&validDate,
			&item.LessonType,
			&taStatus,
		); err != nil {
			return nil, err
		}
		item.TotalQuantity = totalQtyDisp
		item.TotalFreeQuantity = totalFreeDisp
		item.FreeQuantity = remainFree
		item.Quantity = remainQty
		item.EnableExpireTime = enableExpireRaw != 0
		item.AssignedClass = assignedClassRaw != 0
		item.Status = taStatus
		item.IsTuitionAccountActive = taStatus == 1
		item.GeneralLessonIDList = []string{}
		if suspendedTime.Valid && suspendedTime.Time.Year() > 1 {
			item.Suspended = true
			t := suspendedTime.Time
			item.SuspendedTime = &t
		}
		if startTime.Valid {
			item.StartTime = startTime.Time
		}
		if expireTime.Valid {
			item.ExpireTime = expireTime.Time
		}
		if validDate.Valid {
			item.LatestStartTime = validDate.Time
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// ListOneToOneLessonOptionsByStudent 学员在指定学费账户状态下、可用于 1 对 1 的课程（去重）。teach_method=2 为 1v1。
func (repo *Repository) ListOneToOneLessonOptionsByStudent(ctx context.Context, instID, studentID int64, tuitionAccountStatus []int) ([]model.OneToOneLessonOptionVO, error) {
	sqlStr := `
		SELECT
			CAST(ic.id AS CHAR),
			IFNULL(ic.name, ''),
			MAX(IFNULL(ta.assigned_class, 0)),
			MAX(CASE WHEN EXISTS (
				SELECT 1
				FROM teaching_class tc
				INNER JOIN teaching_class_student tcs ON tcs.teaching_class_id = tc.id AND tcs.inst_id = tc.inst_id AND tcs.del_flag = 0
				WHERE tc.inst_id = ta.inst_id
					AND tc.course_id = ic.id
					AND tc.class_type = ?
					AND tc.del_flag = 0
					AND tc.status = ?
					AND tcs.student_id = ta.student_id
				LIMIT 1
			) THEN 1 ELSE 0 END)
		FROM tuition_account ta
		INNER JOIN inst_course ic ON ic.id = ta.course_id AND ic.del_flag = 0
		WHERE ta.inst_id = ?
			AND ta.student_id = ?
			AND ta.del_flag = 0
			AND ic.teach_method = 2`
	args := []any{
		model.TeachingClassTypeOneToOne,
		model.TeachingClassStatusActive,
		instID,
		studentID,
	}
	if len(tuitionAccountStatus) > 0 {
		placeholders := make([]string, 0, len(tuitionAccountStatus))
		for _, st := range tuitionAccountStatus {
			placeholders = append(placeholders, "?")
			args = append(args, st)
		}
		sqlStr += ` AND ta.status IN (` + strings.Join(placeholders, ",") + `)`
	}
	sqlStr += ` GROUP BY ic.id, ic.name ORDER BY ic.name ASC, ic.id ASC`

	rows, err := repo.db.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]model.OneToOneLessonOptionVO, 0, 8)
	for rows.Next() {
		var (
			item         model.OneToOneLessonOptionVO
			maxAssigned  int64
			hasOneToOneO int64
		)
		if err := rows.Scan(&item.ID, &item.Name, &maxAssigned, &hasOneToOneO); err != nil {
			return nil, err
		}
		item.AlreadyEnrolled = maxAssigned != 0 || hasOneToOneO != 0
		out = append(out, item)
	}
	return out, rows.Err()
}
