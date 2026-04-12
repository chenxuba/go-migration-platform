package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

const effectiveTuitionAccountStatusSQL = `
			CASE
				WHEN SUM(CASE WHEN IFNULL(ta.status, 0) = 1 THEN 1 ELSE 0 END) > 0 THEN 1
				WHEN SUM(CASE WHEN IFNULL(ta.status, 0) = 2 THEN 1 ELSE 0 END) > 0 THEN 2
				WHEN SUM(CASE WHEN IFNULL(ta.status, 0) = 3 THEN 1 ELSE 0 END) > 0 THEN 3
				ELSE IFNULL(MAX(ta.status), 0)
			END`

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
			GROUP BY ic.id, ic.teach_method, icq.lesson_model
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
				(
					SELECT IFNULL(SUM(IFNULL(str.arrear_quantity, 0)), 0)
					FROM student_teaching_record str
					WHERE str.inst_id = ?
						AND str.del_flag = 0
						AND str.student_id = ?
						AND str.lesson_id = ic.id
						AND (
							CASE
								WHEN IFNULL(str.sku_mode, 0) = 4 THEN 3
								ELSE IFNULL(str.sku_mode, 0)
						END
					) = (
						CASE
							WHEN IFNULL(icq.lesson_model, 0) = 4 THEN 3
							ELSE IFNULL(icq.lesson_model, 0)
						END
					)
					AND IFNULL(str.arrear_quantity, 0) > 0
					AND IFNULL(str.has_compensated, 0) = 0
			) AS lesson_consume_arrear_quantity,
			SUM(CASE 
				WHEN IFNULL(ta.status, 0) = 3 THEN 0
				WHEN icq.lesson_model = 3 THEN ta.remaining_tuition
				WHEN ta.total_quantity > 0 THEN ta.remaining_quantity 
				ELSE 0 
			END) AS remain_quantity,
			SUM(CASE
				WHEN IFNULL(ta.status, 0) = 3 THEN 0
				ELSE IFNULL(ta.remaining_tuition, 0)
			END) AS tuition,
			SUM(CASE 
				WHEN IFNULL(ta.status, 0) = 3 THEN 0
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
			`+effectiveTuitionAccountStatusSQL+` AS status,
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
			GROUP BY ic.id, ic.name, ic.teach_method, icq.lesson_model
			ORDER BY (`+effectiveTuitionAccountStatusSQL+` = 3) ASC, MAX(ta.create_time) DESC
		`, instID, studentID, instID, studentID)
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
			&item.TotalQuantity,
			&item.TotalFreeQuantity,
			&item.TotalTuition,
			&item.ArrearTuition,
			&item.LessonConsumeArrearQuantity,
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

// ListStudentTuitionAccountsByStudentAndLesson 返回某 1 对 1 课程下聚合后的学费账户明细：
// 1. 当前课程本身的学费账户；
// 2. 已绑定到该课程 1 对 1 班级上的扣费账户（例如手动创建时绑定的通用 1v1 账户）。
// 最终按「课程 + 授课方式 + 计费模式」聚合，避免同一账户桶下多笔原始 tuition_account 被重复计数。
// teachingClassID>0 时仅纳入该班级已绑定的扣费账户，避免同学员同课程历史班级串数；
// orderCourseDetailID>0 时兼容旧行为，只收窄到指定报读明细。
func (repo *Repository) ListStudentTuitionAccountsByStudentAndLesson(ctx context.Context, instID, studentID, courseID, teachingClassID, orderCourseDetailID int64) ([]model.StudentLessonTuitionAccountItem, error) {
	directArgs := []any{instID, studentID, courseID}
	directWhere := ""
	bindWhere := ""
	if teachingClassID > 0 {
		bindWhere += " AND tcs2.teaching_class_id = ?"
	}
	if orderCourseDetailID > 0 {
		directWhere += " AND ta0.order_course_detail_id = ?"
		bindWhere += " AND tcs2.order_course_detail_id = ?"
		directArgs = append(directArgs, orderCourseDetailID)
	}
	bindArgs := []any{instID, studentID, model.TeachingClassTypeOneToOne, courseID}
	if teachingClassID > 0 {
		bindArgs = append(bindArgs, teachingClassID)
	}
	if orderCourseDetailID > 0 {
		bindArgs = append(bindArgs, orderCourseDetailID)
	}
	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			CAST(MIN(ta.id) AS CHAR),
			CAST(MIN(ta.student_id) AS CHAR),
			CAST(ta.course_id AS CHAR),
			IFNULL(MAX(ic.name), '') AS lesson_name,
			MAX(IFNULL(NULLIF(TRIM(ic.name), ''), IFNULL(icq.name, ''))) AS product_name,
			IFNULL(MAX(icq.lesson_model), 0) AS lesson_charging_mode,
			SUM(CASE
				WHEN IFNULL(icq.lesson_model, 0) IN (3, 4) THEN IFNULL(ta.total_tuition, 0)
				WHEN IFNULL(ta.total_quantity, 0) > 0 THEN IFNULL(ta.total_quantity, 0)
				ELSE 0
			END) AS total_quantity_display,
			SUM(CASE
				WHEN IFNULL(icq.lesson_model, 0) IN (3, 4) THEN IFNULL(ta.free_quantity, 0)
				WHEN IFNULL(ta.total_quantity, 0) = 0 AND IFNULL(ta.free_quantity, 0) > 0 THEN IFNULL(ta.free_quantity, 0)
				ELSE 0
			END) AS total_free_quantity_display,
			SUM(IFNULL(ta.total_tuition, 0)),
			SUM(CASE
				WHEN IFNULL(ta.status, 0) = 3 THEN 0
				WHEN IFNULL(ta.total_quantity, 0) = 0 AND IFNULL(ta.free_quantity, 0) > 0 THEN IFNULL(ta.remaining_quantity, 0)
				ELSE 0
			END) AS remain_free_quantity,
			SUM(CASE
				WHEN IFNULL(ta.status, 0) = 3 THEN 0
				WHEN IFNULL(icq.lesson_model, 0) IN (3, 4) THEN IFNULL(ta.remaining_tuition, 0)
				WHEN IFNULL(ta.total_quantity, 0) > 0 THEN IFNULL(ta.remaining_quantity, 0)
				ELSE 0
			END) AS remain_quantity_display,
			SUM(CASE
				WHEN IFNULL(ta.status, 0) = 3 THEN 0
				ELSE IFNULL(ta.remaining_tuition, 0)
			END),
			MAX(ta.suspended_time),
			MIN(IFNULL(ta.create_time, NOW())) AS start_time,
			IFNULL(MAX(ta.enable_expire_time), 0) AS enable_expire,
			MAX(ta.expire_time),
			IFNULL(MAX(ta.assigned_class), 0) AS assigned_class,
			MAX(ta.valid_date),
			IFNULL(MAX(ic.teach_method), 0) AS teach_method,
			`+effectiveTuitionAccountStatusSQL+` AS ta_status
		FROM tuition_account ta
		INNER JOIN (
			SELECT DISTINCT account_id
			FROM (
				SELECT ta0.id AS account_id
				FROM tuition_account ta0
				WHERE ta0.inst_id = ?
					AND ta0.del_flag = 0
					AND ta0.student_id = ?
					AND ta0.course_id = ?
					`+directWhere+`
				UNION ALL
				SELECT COALESCE(
					NULLIF(tcs2.primary_tuition_account_id, 0),
					(SELECT MIN(ta1.id)
					 FROM tuition_account ta1
					 WHERE ta1.order_course_detail_id = tcs2.order_course_detail_id
					   AND ta1.inst_id = tcs2.inst_id
					   AND ta1.del_flag = 0)
				) AS account_id
				FROM teaching_class_student tcs2
				INNER JOIN teaching_class tc2
					ON tc2.id = tcs2.teaching_class_id
					AND tc2.inst_id = tcs2.inst_id
					AND tc2.del_flag = 0
				WHERE tcs2.inst_id = ?
					AND tcs2.del_flag = 0
					AND tcs2.student_id = ?
					AND tc2.class_type = ?
					AND tc2.course_id = ?
					`+bindWhere+`
			) candidates
			WHERE account_id IS NOT NULL AND account_id > 0
		) matched ON matched.account_id = ta.id
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
			AND IFNULL(ta.status, 0) <> 3
		GROUP BY ta.course_id, IFNULL(ic.teach_method, 0), IFNULL(icq.lesson_model, -99999)
		ORDER BY MIN(IFNULL(ta.create_time, NOW())) DESC, MIN(ta.id) DESC
	`, append(append(directArgs, bindArgs...), instID)...)
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
			&item.LessonName,
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

// ListStudentOneToOneDeductionTuitionAccounts 学员名下「在读 + 班级授课或 1v1」扣费账户，
// 按课程 + 授课方式 + 计费模式聚合一行（多笔订单课时/余额相加）。
// id 为 agg:{courseId}:{teachMethod}:{lessonModel}；创建 1 对 1 时服务端仅绑定该聚合桶下的 tuition_account。
func (repo *Repository) ListStudentOneToOneDeductionTuitionAccounts(ctx context.Context, instID, studentID int64) ([]model.StudentLessonTuitionAccountItem, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			CONCAT(
				'agg:',
				CAST(agg.course_id AS CHAR), ':',
				CAST(agg.teach_method AS CHAR), ':',
				CAST(agg.lesson_charging_mode_key AS CHAR)
			),
			CAST(agg.student_id AS CHAR),
			CAST(agg.course_id AS CHAR),
			IFNULL(agg.lesson_name, ''),
			IFNULL(agg.product_name, ''),
			IFNULL(agg.lesson_charging_mode, 0),
			IFNULL(agg.total_quantity_display, 0),
			IFNULL(agg.total_free_quantity_display, 0),
			IFNULL(agg.sum_total_tuition, 0),
			IFNULL(agg.remain_free_sum, 0),
			IFNULL(agg.remain_quantity_display, 0),
			IFNULL(agg.sum_remaining_tuition, 0),
			agg.suspended_time,
			agg.first_start_time,
			IFNULL(agg.enable_expire_max, 0),
			agg.expire_time_max,
			IFNULL(agg.assigned_class_max, 0),
			agg.valid_date_max,
			IFNULL(agg.teach_method, 0),
			1 AS ta_status
		FROM (
			SELECT
				ta.inst_id,
				ta.student_id,
				ta.course_id,
				IFNULL(ic.teach_method, 0) AS teach_method,
				IFNULL(icq.lesson_model, 0) AS lesson_charging_mode_key,
				MAX(IFNULL(ic.name, '')) AS lesson_name,
				MAX(IFNULL(NULLIF(TRIM(ic.name), ''), IFNULL(icq.name, ''))) AS product_name,
				MAX(IFNULL(icq.lesson_model, 0)) AS lesson_charging_mode,
				SUM(CASE
					WHEN IFNULL(icq.lesson_model, 0) IN (3, 4) THEN IFNULL(ta.total_tuition, 0)
					WHEN IFNULL(ta.total_quantity, 0) > 0 THEN IFNULL(ta.total_quantity, 0)
					ELSE 0
				END) AS total_quantity_display,
				SUM(CASE
					WHEN IFNULL(icq.lesson_model, 0) IN (3, 4) THEN IFNULL(ta.free_quantity, 0)
					WHEN IFNULL(ta.total_quantity, 0) = 0 AND IFNULL(ta.free_quantity, 0) > 0 THEN IFNULL(ta.free_quantity, 0)
					ELSE 0
				END) AS total_free_quantity_display,
				SUM(IFNULL(ta.total_tuition, 0)) AS sum_total_tuition,
				SUM(CASE
					WHEN IFNULL(ta.total_quantity, 0) = 0 AND IFNULL(ta.free_quantity, 0) > 0 THEN IFNULL(ta.remaining_quantity, 0)
					ELSE 0
				END) AS remain_free_sum,
				SUM(CASE
					WHEN IFNULL(icq.lesson_model, 0) IN (3, 4) THEN IFNULL(ta.remaining_tuition, 0)
					WHEN IFNULL(ta.total_quantity, 0) > 0 THEN IFNULL(ta.remaining_quantity, 0)
					ELSE 0
				END) AS remain_quantity_display,
				SUM(IFNULL(ta.remaining_tuition, 0)) AS sum_remaining_tuition,
				MAX(ta.suspended_time) AS suspended_time,
				MIN(IFNULL(ta.create_time, NOW())) AS first_start_time,
				MAX(IFNULL(ta.enable_expire_time, 0)) AS enable_expire_max,
				MAX(ta.expire_time) AS expire_time_max,
				MAX(IFNULL(ta.assigned_class, 0)) AS assigned_class_max,
				MAX(ta.valid_date) AS valid_date_max
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
				AND IFNULL(ta.status, 0) = 1
				AND ic.teach_method IN (1, 2)
			GROUP BY ta.inst_id, ta.student_id, ta.course_id, IFNULL(ic.teach_method, 0), IFNULL(icq.lesson_model, 0)
		) agg
		ORDER BY IFNULL(agg.lesson_name, ''), agg.course_id, agg.teach_method, agg.lesson_charging_mode_key
	`, instID, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]model.StudentLessonTuitionAccountItem, 0, 16)
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
			&item.LessonName,
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
		item.IsAggregate = true
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

// ListTuitionAccountIDsForStudentCourse 学员某课程下全部在读、且班级/1v1 授课的学费账户 id（按创建时间、id 升序，与扣费 FIFO 一致）。
func (repo *Repository) ListTuitionAccountIDsForStudentCourse(ctx context.Context, tx *sql.Tx, instID, studentID, courseID int64) ([]int64, error) {
	q := `
		SELECT ta.id
		FROM tuition_account ta
		INNER JOIN inst_course ic ON ic.id = ta.course_id AND ic.del_flag = 0
		WHERE ta.inst_id = ? AND ta.del_flag = 0 AND ta.student_id = ? AND ta.course_id = ?
			AND IFNULL(ta.status, 0) = 1 AND ic.teach_method IN (1, 2)
		ORDER BY ta.create_time ASC, ta.id ASC
	`
	var rows *sql.Rows
	var err error
	if tx != nil {
		rows, err = tx.QueryContext(ctx, q, instID, studentID, courseID)
	} else {
		rows, err = repo.db.QueryContext(ctx, q, instID, studentID, courseID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// ListTuitionAccountIDsForStudentCourseBucket 学员某课程下、指定授课方式/计费模式桶内的全部在读学费账户 id（FIFO）。
func (repo *Repository) ListTuitionAccountIDsForStudentCourseBucket(ctx context.Context, tx *sql.Tx, instID, studentID, courseID int64, teachMethod, lessonChargingMode int) ([]int64, error) {
	q := `
		SELECT ta.id
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
		WHERE ta.inst_id = ? AND ta.del_flag = 0 AND ta.student_id = ? AND ta.course_id = ?
			AND IFNULL(ta.status, 0) = 1
			AND IFNULL(ic.teach_method, 0) = ?
			AND IFNULL(icq.lesson_model, 0) = ?
		ORDER BY ta.create_time ASC, ta.id ASC
	`
	var rows *sql.Rows
	var err error
	if tx != nil {
		rows, err = tx.QueryContext(ctx, q, instID, studentID, courseID, teachMethod, lessonChargingMode)
	} else {
		rows, err = repo.db.QueryContext(ctx, q, instID, studentID, courseID, teachMethod, lessonChargingMode)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// ListTuitionAccountIDsForStudentCourseBucketAllStatuses 学员某课程下、指定授课方式/计费模式桶内的全部学费账户 id（含已结课/停课，FIFO）。
func (repo *Repository) ListTuitionAccountIDsForStudentCourseBucketAllStatuses(ctx context.Context, tx *sql.Tx, instID, studentID, courseID int64, teachMethod, lessonChargingMode int) ([]int64, error) {
	q := `
		SELECT ta.id
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
		WHERE ta.inst_id = ? AND ta.del_flag = 0 AND ta.student_id = ? AND ta.course_id = ?
			AND IFNULL(ic.teach_method, 0) = ?
			AND IFNULL(icq.lesson_model, 0) = ?
		ORDER BY ta.create_time ASC, ta.id ASC
	`
	var rows *sql.Rows
	var err error
	if tx != nil {
		rows, err = tx.QueryContext(ctx, q, instID, studentID, courseID, teachMethod, lessonChargingMode)
	} else {
		rows, err = repo.db.QueryContext(ctx, q, instID, studentID, courseID, teachMethod, lessonChargingMode)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// ListOneToOneLessonOptionsByStudent 机构内全部 1v1 课程（teach_method=2），供创建/选择 1 对 1 上课课程。
// LEFT JOIN 学员学费账户仅用于 assigned_class；「已报名」仍看是否存在该学员在该课下的开班中 1 对 1。
func (repo *Repository) ListOneToOneLessonOptionsByStudent(ctx context.Context, instID, studentID int64, tuitionAccountStatus []int) ([]model.OneToOneLessonOptionVO, error) {
	sqlStr := `
		SELECT
			CAST(ic.id AS CHAR),
			IFNULL(ic.name, ''),
			IFNULL(MAX(IFNULL(ta.assigned_class, 0)), 0),
			MAX(CASE WHEN EXISTS (
				SELECT 1
				FROM teaching_class tc
				INNER JOIN teaching_class_student tcs ON tcs.teaching_class_id = tc.id AND tcs.inst_id = tc.inst_id AND tcs.del_flag = 0
				WHERE tc.inst_id = ?
					AND tc.course_id = ic.id
					AND tc.class_type = ?
					AND tc.del_flag = 0
					AND tc.status = ?
					AND tcs.student_id = ?
				LIMIT 1
			) THEN 1 ELSE 0 END)
		FROM inst_course ic
		LEFT JOIN tuition_account ta ON ta.course_id = ic.id AND ta.inst_id = ic.inst_id AND ta.student_id = ? AND ta.del_flag = 0`
	args := []any{
		instID,
		model.TeachingClassTypeOneToOne,
		model.TeachingClassStatusActive,
		studentID,
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
	sqlStr += `
		WHERE ic.inst_id = ?
			AND ic.del_flag = 0
			AND ic.teach_method = 2`
	args = append(args, instID)
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

func tuitionRemainQuantityForDisplay(lessonModel int, remQty, remTui float64) float64 {
	if lessonModel == 3 || lessonModel == 4 {
		return remTui
	}
	return remQty
}

func appendTuitionAccountLessonStudentFilters(baseFrom string, argsBase []any, f model.TuitionAccountLessonPageFilters) (string, []any) {
	if len(f.Sex) > 0 {
		ph := sqlPlaceholders(len(f.Sex))
		baseFrom += ` AND IFNULL(s.stu_sex, 0) IN (` + ph + `)`
		for _, v := range f.Sex {
			argsBase = append(argsBase, v)
		}
	}
	if f.AgeMin != nil || f.AgeMax != nil {
		baseFrom += ` AND s.birthday IS NOT NULL AND YEAR(s.birthday) > 1900`
		if f.AgeMin != nil && f.AgeMax != nil {
			baseFrom += ` AND TIMESTAMPDIFF(YEAR, s.birthday, CURDATE()) BETWEEN ? AND ?`
			argsBase = append(argsBase, *f.AgeMin, *f.AgeMax)
		} else if f.AgeMin != nil {
			baseFrom += ` AND TIMESTAMPDIFF(YEAR, s.birthday, CURDATE()) >= ?`
			argsBase = append(argsBase, *f.AgeMin)
		} else {
			baseFrom += ` AND TIMESTAMPDIFF(YEAR, s.birthday, CURDATE()) <= ?`
			argsBase = append(argsBase, *f.AgeMax)
		}
	}
	if sn := strings.TrimSpace(f.StudentName); sn != "" {
		baseFrom += ` AND IFNULL(s.stu_name, '') LIKE ?`
		argsBase = append(argsBase, "%"+sn+"%")
	}
	return baseFrom, argsBase
}

// PageTuitionAccountsByLessonForGroupAdd 对标 TuitionAccount/GetTuitionAccountListByLessonId：班课在读账户，供集体班添加学员列表。
func (repo *Repository) PageTuitionAccountsByLessonForGroupAdd(ctx context.Context, instID int64, courseIDs []int64, currentClassID int64, studentIDs []int64, pageIndex, pageSize int, filters model.TuitionAccountLessonPageFilters) (list []model.TuitionAccountByLessonRowVO, total int, err error) {
	if len(courseIDs) == 0 {
		return nil, 0, nil
	}
	phCourses := sqlPlaceholders(len(courseIDs))
	baseFrom := `
		FROM tuition_account ta
		INNER JOIN inst_course ic ON ic.id = ta.course_id AND ic.inst_id = ta.inst_id AND ic.del_flag = 0
		INNER JOIN inst_student s ON s.id = ta.student_id AND s.inst_id = ta.inst_id AND s.del_flag = 0
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
			AND IFNULL(ta.status, 0) = 1
			AND ic.teach_method = 1
			AND ta.course_id IN (` + phCourses + `)
	`
	argsBase := make([]any, 0, 1+len(courseIDs))
	argsBase = append(argsBase, instID)
	for _, cid := range courseIDs {
		argsBase = append(argsBase, cid)
	}
	if len(studentIDs) > 0 {
		phStu := sqlPlaceholders(len(studentIDs))
		baseFrom += ` AND ta.student_id IN (` + phStu + `)`
		for _, sid := range studentIDs {
			argsBase = append(argsBase, sid)
		}
	}
	baseFrom, argsBase = appendTuitionAccountLessonStudentFilters(baseFrom, argsBase, filters)

	countQ := `SELECT COUNT(1) ` + baseFrom
	if err := repo.db.QueryRowContext(ctx, countQ, argsBase...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []model.TuitionAccountByLessonRowVO{}, 0, nil
	}
	if pageIndex <= 0 {
		pageIndex = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (pageIndex - 1) * pageSize

	dataQ := `
		SELECT
			CAST(ta.id AS CHAR),
			CAST(ta.student_id AS CHAR),
			IFNULL(s.stu_name, ''),
			IFNULL(s.avatar_url, ''),
			IFNULL(s.mobile, ''),
			IFNULL(s.stu_sex, 0),
			s.birthday,
			IFNULL(icq.lesson_model, 0),
			IFNULL(ta.remaining_quantity, 0),
			IFNULL(ta.remaining_tuition, 0),
			IFNULL(NULLIF(TRIM(ic.name), ''), IFNULL(icq.name, '')) AS product_name,
			CAST(IFNULL(icq.id, 0) AS CHAR),
			IFNULL(ta.create_time, NOW()),
			CASE WHEN IFNULL(ta.status, 0) = 1 THEN 1 ELSE 0 END,
			CASE WHEN ? > 0 AND EXISTS (
				SELECT 1 FROM teaching_class_student tcs
				WHERE tcs.inst_id = ta.inst_id AND tcs.teaching_class_id = ? AND tcs.student_id = ta.student_id AND tcs.del_flag = 0
			) THEN 1 ELSE 0 END
	` + baseFrom + `
		ORDER BY ta.id DESC
		LIMIT ? OFFSET ?
	`
	dataArgs := make([]any, 0, 2+len(argsBase)+2)
	dataArgs = append(dataArgs, currentClassID, currentClassID)
	dataArgs = append(dataArgs, argsBase...)
	dataArgs = append(dataArgs, pageSize, offset)

	rows, err := repo.db.QueryContext(ctx, dataQ, dataArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list = make([]model.TuitionAccountByLessonRowVO, 0, pageSize)
	for rows.Next() {
		var row model.TuitionAccountByLessonRowVO
		var avatar sql.NullString
		var birthday sql.NullTime
		var lm int
		var remQty, remTui float64
		var productName string
		var productID string
		var startT sql.NullTime
		var activeRaw, assignedRaw int
		if err := rows.Scan(
			&row.TuitionAccountID,
			&row.StudentID,
			&row.StudentName,
			&avatar,
			&row.Phone,
			&row.Sex,
			&birthday,
			&lm,
			&remQty,
			&remTui,
			&productName,
			&productID,
			&startT,
			&activeRaw,
			&assignedRaw,
		); err != nil {
			return nil, 0, err
		}
		row.LessonChargingMode = lm
		row.LessonScope = 2
		row.Quantity = tuitionRemainQuantityForDisplay(lm, remQty, remTui)
		row.ProductName = strings.TrimSpace(productName)
		row.ProductID = productID
		row.Phone = maskPhoneDisplay(row.Phone)
		row.IsTuitionAccountActive = activeRaw != 0
		row.AssignedClass = assignedRaw != 0
		row.IsCrossSchoolStudent = false
		if avatar.Valid && strings.TrimSpace(avatar.String) != "" {
			a := strings.TrimSpace(avatar.String)
			row.Avatar = &a
		}
		if birthday.Valid && birthday.Time.Year() > 1 {
			row.Birthday = birthday.Time
		}
		if startT.Valid {
			row.StartTime = startT.Time
		}
		list = append(list, row)
	}
	return list, total, rows.Err()
}
