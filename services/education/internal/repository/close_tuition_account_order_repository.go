package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

func closeOrderAlmostEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.02
}

func closeOrderRoundMoney(v float64) float64 {
	return math.Round(v*100) / 100
}

const closeTuitionAccountICQJoin = `
LEFT JOIN sale_order_course_detail sod_ta ON sod_ta.id = ta.order_course_detail_id AND sod_ta.del_flag = 0
LEFT JOIN inst_course_quotation icq_ta ON icq_ta.id = COALESCE(
	NULLIF(ta.quote_id, 0),
	NULLIF(sod_ta.quote_id, 0),
	(SELECT qx.id FROM inst_course_quotation qx
	 WHERE qx.course_id = ta.course_id AND qx.del_flag = 0
	   AND ABS(IFNULL(qx.quantity, 0) - IFNULL(ta.total_quantity, 0)) < 0.000001
	   AND ABS(IFNULL(qx.price, 0) - IFNULL(ta.total_tuition, 0)) < 0.000001
	 ORDER BY qx.id DESC LIMIT 1),
	(SELECT qmin.id FROM inst_course_quotation qmin
	 WHERE qmin.course_id = ta.course_id AND qmin.del_flag = 0
	 ORDER BY qmin.id ASC LIMIT 1)
) AND icq_ta.del_flag = 0`

// AddCloseTuitionAccountOrder 手动结课：扣减学费账户剩余、写入 tuition_account_flow（联动课消明细 / 学费变动 / 确认收入列表）
func (repo *Repository) AddCloseTuitionAccountOrder(ctx context.Context, instID, operatorID, tuitionAccountID int64, quantity, freeQuantity, tuition float64, remark string) (int64, error) {
	if tuitionAccountID <= 0 {
		return 0, errors.New("tuitionAccountId 无效")
	}
	deductQty := quantity + freeQuantity
	if deductQty < 0 || tuition < 0 {
		return 0, errors.New("数量或学费不能为负")
	}
	if deductQty < 0.0001 && tuition < 0.0001 {
		return 0, errors.New("无可结课的剩余课时或学费")
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var (
		studentID, courseID                   int64
		totalQty, freeQty, usedQty, remQty    float64
		totalTuition, usedTuition, remTuition float64
		confirmedTuition                      float64
		lessonModel                           int
		enableExpire                          int
		teachMethod                           sql.NullInt64
		orderNumber                           string
	)
	err = tx.QueryRowContext(ctx, `
		SELECT
			ta.student_id,
			ta.course_id,
			IFNULL(ta.total_quantity, 0),
			IFNULL(ta.free_quantity, 0),
			IFNULL(ta.used_quantity, 0),
			IFNULL(ta.remaining_quantity, 0),
			IFNULL(ta.total_tuition, 0),
			IFNULL(ta.used_tuition, 0),
			IFNULL(ta.remaining_tuition, 0),
			IFNULL(ta.confirmed_tuition, 0),
			IFNULL(icq_ta.lesson_model, 0),
			IFNULL(ta.enable_expire_time, 0),
			ic.teach_method,
			IFNULL(so.order_number, '')
		FROM tuition_account ta
		INNER JOIN inst_course ic ON ic.id = ta.course_id AND ic.del_flag = 0
		`+closeTuitionAccountICQJoin+`
		LEFT JOIN sale_order so ON so.id = ta.order_id AND so.del_flag = 0
		WHERE ta.id = ? AND ta.inst_id = ? AND ta.del_flag = 0
		FOR UPDATE
	`, tuitionAccountID, instID).Scan(
		&studentID, &courseID, &totalQty, &freeQty, &usedQty, &remQty,
		&totalTuition, &usedTuition, &remTuition, &confirmedTuition,
		&lessonModel, &enableExpire, &teachMethod, &orderNumber,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("学费账户不存在")
		}
		return 0, err
	}

	if !closeOrderAlmostEqual(tuition, remTuition) {
		return 0, fmt.Errorf("剩余学费与提交不一致（当前 ¥%.2f）", remTuition)
	}
	if lessonModel == 3 || lessonModel == 4 {
		if remQty > 0.0001 && !closeOrderAlmostEqual(deductQty, remQty) {
			return 0, fmt.Errorf("剩余数量与提交不一致（当前 %.2f）", remQty)
		}
	} else {
		if !closeOrderAlmostEqual(deductQty, remQty) {
			return 0, fmt.Errorf("剩余课时与提交不一致（当前 %.2f）", remQty)
		}
	}

	newRemQty := closeOrderRoundMoney(remQty - deductQty)
	newRemTuition := closeOrderRoundMoney(remTuition - tuition)
	if newRemQty < -0.03 || newRemTuition < -0.03 {
		return 0, errors.New("扣减后余额异常，请刷新后重试")
	}
	if newRemQty < 0 {
		newRemQty = 0
	}
	if newRemTuition < 0 {
		newRemTuition = 0
	}
	newUsedQty := closeOrderRoundMoney(usedQty + deductQty)
	newUsedTuition := closeOrderRoundMoney(usedTuition + tuition)
	newConfirmed := closeOrderRoundMoney(confirmedTuition + tuition)

	_, err = tx.ExecContext(ctx, `
		UPDATE tuition_account
		SET remaining_quantity = ?,
		    remaining_tuition = ?,
		    used_quantity = ?,
		    used_tuition = ?,
		    confirmed_tuition = ?,
		    update_id = ?,
		    update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, newRemQty, newRemTuition, newUsedQty, newUsedTuition, newConfirmed, operatorID, tuitionAccountID, instID)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	sourceID := now.UnixNano()
	if sourceID < 0 {
		sourceID = -sourceID
	}

	_ = remark // API 保留备注；流水表暂无备注列，结课说明由 source_type=25 体现
	// order_number：与其它学费流水一致，存 tuition_account 关联的 sale_order.order_number（不再用「结课」占位）。
	flowOrderNumber := strings.TrimSpace(orderNumber)
	if flowOrderNumber == "" {
		flowOrderNumber = "-"
	}
	if len(flowOrderNumber) > 64 {
		flowOrderNumber = flowOrderNumber[:64]
	}

	var lessonTypeVal any
	if teachMethod.Valid {
		lessonTypeVal = teachMethod.Int64
	} else {
		lessonTypeVal = nil
	}

	flowChargingMode := lessonModel
	if flowChargingMode <= 0 && enableExpire == 1 && totalQty > 0.0001 {
		flowChargingMode = 2 // 与 tuition_account_flow 回填逻辑一致：按时段/天账户
	} else if flowChargingMode <= 0 && totalQty > 0.0001 {
		flowChargingMode = 1 // 课时账户
	} else if flowChargingMode <= 0 && totalTuition > 0.0001 {
		flowChargingMode = 3 // 金额账户
	}

	// 与课消类流水一致：quantity/tuition 存正数，供「确认收入」SUM/列表展示为机构收入；
	// 学费变动列表用 sourceType=25 + direction=out 展示为「-」扣减。
	flowQty := deductQty
	flowTuition := tuition

	res, err := tx.ExecContext(ctx, `
		INSERT INTO tuition_account_flow (
			uuid, version, inst_id, tuition_account_id, student_id, product_id, lesson_type, lesson_charging_mode,
			source_type, source_id, teaching_record_id, order_number, created_time, quantity, tuition, balance_quantity, balance_tuition,
			create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, ?, ?, ?, ?, ?,
			?, ?, NULL, ?, ?, ?, ?, ?, ?,
			?, NOW(), ?, NOW(), 0
		)
	`,
		instID,
		tuitionAccountID,
		studentID,
		courseID,
		lessonTypeVal,
		flowChargingMode,
		model.TuitionAccountFlowSourceManualCloseCourse,
		sourceID,
		flowOrderNumber,
		now,
		flowQty,
		flowTuition,
		newRemQty,
		newRemTuition,
		operatorID,
		operatorID,
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return 0, errors.New("请勿重复提交结课")
		}
		return 0, err
	}
	flowID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	_, _ = tx.ExecContext(ctx, `
		UPDATE teaching_class_student
		SET class_student_status = ?, update_id = ?, update_time = NOW()
		WHERE inst_id = ? AND primary_tuition_account_id = ? AND del_flag = 0
	`, model.TeachingClassStudentStatusClosed, operatorID, instID, tuitionAccountID)

	// 全部学费账户已无剩余数量与剩余学费时，将在读学员标为历史学员
	if _, err := tx.ExecContext(ctx, `
		UPDATE inst_student s
		SET s.student_status = ?,
		    s.update_id = ?,
		    s.update_time = NOW()
		WHERE s.id = ? AND s.inst_id = ? AND s.del_flag = 0
		  AND s.student_status = ?
		  AND NOT EXISTS (
			SELECT 1 FROM tuition_account ta
			WHERE ta.del_flag = 0 AND ta.inst_id = s.inst_id AND ta.student_id = s.id
			  AND (IFNULL(ta.remaining_quantity, 0) > 0.02 OR IFNULL(ta.remaining_tuition, 0) > 0.02)
			LIMIT 1
		  )
	`, model.InstStudentStatusHistory, operatorID, studentID, instID, model.InstStudentStatusEnrolled); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return flowID, nil
}

// fixManualCloseTuitionFlowNegativeAmounts 修正早期手动结课流水误存的负数，与课消/确认收入列表 SUM 口径一致（正数表示本次确认量与金额）。
func fixManualCloseTuitionFlowNegativeAmounts(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		UPDATE tuition_account_flow
		SET quantity = ABS(quantity), tuition = ABS(tuition)
		WHERE source_type = ?
		  AND del_flag = 0
		  AND (quantity < 0 OR tuition < 0)
	`, model.TuitionAccountFlowSourceManualCloseCourse)
	return err
}

// fixManualCloseCourseFlowOrderNumbers 将早期误写入的「结课」占位 order_number 回填为学费账户关联的销售订单号（与其它流水一致）。
func fixManualCloseCourseFlowOrderNumbers(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		UPDATE tuition_account_flow taf
		INNER JOIN tuition_account ta ON ta.id = taf.tuition_account_id AND ta.inst_id = taf.inst_id AND ta.del_flag = 0
		LEFT JOIN sale_order so ON so.id = ta.order_id AND so.del_flag = 0
		SET taf.order_number = SUBSTRING(COALESCE(NULLIF(TRIM(IFNULL(so.order_number, '')), ''), '-'), 1, 64)
		WHERE taf.source_type = ?
		  AND taf.del_flag = 0
		  AND (taf.order_number = '结课' OR taf.order_number LIKE '结课|%')
	`, model.TuitionAccountFlowSourceManualCloseCourse)
	return err
}
