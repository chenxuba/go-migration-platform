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

type closeTuitionAccountSnapshot struct {
	id               int64
	studentID        int64
	courseID         int64
	totalQty         float64
	freeQty          float64
	usedQty          float64
	remQty           float64
	totalTuition     float64
	usedTuition      float64
	remTuition       float64
	confirmedTuition float64
	lessonModel      int
	enableExpire     int
	teachMethod      sql.NullInt64
	orderNumber      string
}

func (snap closeTuitionAccountSnapshot) flowChargingMode() int {
	if snap.lessonModel > 0 {
		return snap.lessonModel
	}
	if snap.enableExpire == 1 && snap.totalQty > 0.0001 {
		return 2
	}
	if snap.totalQty > 0.0001 {
		return 1
	}
	if snap.totalTuition > 0.0001 {
		return 3
	}
	return 0
}

func (snap closeTuitionAccountSnapshot) lessonTypeValue() any {
	if snap.teachMethod.Valid {
		return snap.teachMethod.Int64
	}
	return nil
}

func closeOrderMatchesSubmitted(deductQty, tuition float64, lessonModel int, currentRemQty, currentRemTuition float64) bool {
	if !closeOrderAlmostEqual(tuition, currentRemTuition) {
		return false
	}
	if lessonModel == 3 || lessonModel == 4 {
		if currentRemQty > 0.0001 && !closeOrderAlmostEqual(deductQty, currentRemQty) {
			return false
		}
		return true
	}
	return closeOrderAlmostEqual(deductQty, currentRemQty)
}

func closeOrderMismatchError(deductQty, tuition float64, lessonModel int, currentRemQty, currentRemTuition float64) error {
	if !closeOrderAlmostEqual(tuition, currentRemTuition) {
		return fmt.Errorf("剩余学费与提交不一致（当前 ¥%.2f）", currentRemTuition)
	}
	if lessonModel == 3 || lessonModel == 4 {
		if currentRemQty > 0.0001 && !closeOrderAlmostEqual(deductQty, currentRemQty) {
			return fmt.Errorf("剩余数量与提交不一致（当前 %.2f）", currentRemQty)
		}
		return nil
	}
	if !closeOrderAlmostEqual(deductQty, currentRemQty) {
		return fmt.Errorf("剩余课时与提交不一致（当前 %.2f）", currentRemQty)
	}
	return nil
}

func sumCloseTuitionAccountSnapshots(snaps []closeTuitionAccountSnapshot) (float64, float64) {
	var remQty float64
	var remTuition float64
	for _, snap := range snaps {
		remQty += snap.remQty
		remTuition += snap.remTuition
	}
	return closeOrderRoundMoney(remQty), closeOrderRoundMoney(remTuition)
}

func (repo *Repository) loadCloseTuitionAccountSnapshotTx(ctx context.Context, tx *sql.Tx, instID, tuitionAccountID int64) (closeTuitionAccountSnapshot, error) {
	var snap closeTuitionAccountSnapshot
	err := tx.QueryRowContext(ctx, `
		SELECT
			ta.id,
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
		&snap.id,
		&snap.studentID,
		&snap.courseID,
		&snap.totalQty,
		&snap.freeQty,
		&snap.usedQty,
		&snap.remQty,
		&snap.totalTuition,
		&snap.usedTuition,
		&snap.remTuition,
		&snap.confirmedTuition,
		&snap.lessonModel,
		&snap.enableExpire,
		&snap.teachMethod,
		&snap.orderNumber,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return closeTuitionAccountSnapshot{}, errors.New("学费账户不存在")
		}
		return closeTuitionAccountSnapshot{}, err
	}
	return snap, nil
}

func (repo *Repository) loadCloseTuitionAccountBucketSnapshotsTx(ctx context.Context, tx *sql.Tx, instID int64, selected closeTuitionAccountSnapshot) ([]closeTuitionAccountSnapshot, error) {
	teachMethod := 0
	if selected.teachMethod.Valid {
		teachMethod = int(selected.teachMethod.Int64)
	}
	ids, err := repo.ListTuitionAccountIDsForStudentCourseBucket(ctx, tx, instID, selected.studentID, selected.courseID, teachMethod, selected.lessonModel)
	if err != nil {
		return nil, err
	}
	out := make([]closeTuitionAccountSnapshot, 0, len(ids))
	for _, id := range ids {
		snap, snapErr := repo.loadCloseTuitionAccountSnapshotTx(ctx, tx, instID, id)
		if snapErr != nil {
			return nil, snapErr
		}
		out = append(out, snap)
	}
	return out, nil
}

func (repo *Repository) closeTuitionAccountSnapshotTx(ctx context.Context, tx *sql.Tx, instID, operatorID int64, snap closeTuitionAccountSnapshot, sourceID int64, now time.Time) (int64, error) {
	deductQty := closeOrderRoundMoney(snap.remQty)
	tuition := closeOrderRoundMoney(snap.remTuition)
	newUsedQty := closeOrderRoundMoney(snap.usedQty + snap.remQty)
	newUsedTuition := closeOrderRoundMoney(snap.usedTuition + snap.remTuition)
	newConfirmed := closeOrderRoundMoney(snap.confirmedTuition + snap.remTuition)

	if _, err := tx.ExecContext(ctx, `
		UPDATE tuition_account
		SET remaining_quantity = 0,
		    remaining_tuition = 0,
		    used_quantity = ?,
		    used_tuition = ?,
		    confirmed_tuition = ?,
		    status = ?,
		    status_change_time = ?,
		    class_ending_time = ?,
		    update_id = ?,
		    update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, newUsedQty, newUsedTuition, newConfirmed, model.TuitionAccountStatusClosed, now, now, operatorID, snap.id, instID); err != nil {
		return 0, err
	}

	if deductQty < 0.0001 && tuition < 0.0001 {
		return 0, nil
	}

	flowOrderNumber := strings.TrimSpace(snap.orderNumber)
	if flowOrderNumber == "" {
		flowOrderNumber = "-"
	}
	if len(flowOrderNumber) > 64 {
		flowOrderNumber = flowOrderNumber[:64]
	}

	res, err := tx.ExecContext(ctx, `
		INSERT INTO tuition_account_flow (
			uuid, version, inst_id, tuition_account_id, student_id, product_id, lesson_type, lesson_charging_mode,
			source_type, source_id, teaching_record_id, order_number, created_time, quantity, tuition, balance_quantity, balance_tuition,
			create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, ?, ?, ?, ?, ?,
			?, ?, NULL, ?, ?, ?, ?, 0, 0,
			?, NOW(), ?, NOW(), 0
		)
	`,
		instID,
		snap.id,
		snap.studentID,
		snap.courseID,
		snap.lessonTypeValue(),
		snap.flowChargingMode(),
		model.TuitionAccountFlowSourceManualCloseCourse,
		sourceID,
		flowOrderNumber,
		now,
		deductQty,
		tuition,
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
	return flowID, nil
}

func (repo *Repository) closeRelatedOneToOneClassesByDeductCourseTx(ctx context.Context, tx *sql.Tx, instID, operatorID, studentID, courseID int64) error {
	if studentID <= 0 || courseID <= 0 {
		return nil
	}

	studentStatusArgs := []any{
		model.TeachingClassTypeOneToOne,
		model.TeachingClassStudentStatusClosed,
		operatorID,
		instID,
		studentID,
		courseID,
		model.TeachingClassStudentStatusClosed,
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE teaching_class_student tcs
		INNER JOIN teaching_class tc
			ON tc.id = tcs.teaching_class_id
			AND tc.inst_id = tcs.inst_id
			AND tc.class_type = ?
			AND tc.del_flag = 0
		LEFT JOIN tuition_account ta_eff ON ta_eff.id = COALESCE(
			NULLIF(tcs.primary_tuition_account_id, 0),
			(SELECT MIN(ta0.id)
			 FROM tuition_account ta0
			 WHERE ta0.order_course_detail_id = tcs.order_course_detail_id
			   AND ta0.inst_id = tcs.inst_id
			   AND ta0.del_flag = 0)
		) AND ta_eff.inst_id = tcs.inst_id AND ta_eff.del_flag = 0
		SET tcs.class_student_status = ?,
		    tcs.update_id = ?,
		    tcs.update_time = NOW()
		WHERE tcs.inst_id = ?
		  AND tcs.del_flag = 0
		  AND tcs.student_id = ?
		  AND ta_eff.course_id = ?
		  AND tcs.class_student_status <> ?
	`, studentStatusArgs...); err != nil {
		return err
	}

	classStatusArgs := []any{
		model.TeachingClassStatusClosed,
		operatorID,
		model.TeachingClassTypeOneToOne,
		instID,
		studentID,
		courseID,
		instID,
		model.TeachingClassStatusClosed,
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE teaching_class tc
		INNER JOIN teaching_class_student tcs
			ON tcs.teaching_class_id = tc.id
			AND tcs.inst_id = tc.inst_id
			AND tcs.del_flag = 0
		LEFT JOIN tuition_account ta_eff ON ta_eff.id = COALESCE(
			NULLIF(tcs.primary_tuition_account_id, 0),
			(SELECT MIN(ta0.id)
			 FROM tuition_account ta0
			 WHERE ta0.order_course_detail_id = tcs.order_course_detail_id
			   AND ta0.inst_id = tcs.inst_id
			   AND ta0.del_flag = 0)
		) AND ta_eff.inst_id = tcs.inst_id AND ta_eff.del_flag = 0
		SET tc.status = ?,
		    tc.update_id = ?,
		    tc.update_time = NOW()
		WHERE tc.class_type = ?
		  AND tc.del_flag = 0
		  AND tcs.inst_id = ?
		  AND tcs.student_id = ?
		  AND ta_eff.course_id = ?
		  AND tc.inst_id = ?
		  AND tc.status <> ?
	`, classStatusArgs...); err != nil {
		return err
	}

	return nil
}

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

	selectedSnap, err := repo.loadCloseTuitionAccountSnapshotTx(ctx, tx, instID, tuitionAccountID)
	if err != nil {
		return 0, err
	}

	targetSnapshots := []closeTuitionAccountSnapshot{selectedSnap}
	if !closeOrderMatchesSubmitted(deductQty, tuition, selectedSnap.lessonModel, selectedSnap.remQty, selectedSnap.remTuition) {
		bucketSnapshots, bucketErr := repo.loadCloseTuitionAccountBucketSnapshotsTx(ctx, tx, instID, selectedSnap)
		if bucketErr != nil {
			return 0, bucketErr
		}
		bucketRemQty, bucketRemTuition := sumCloseTuitionAccountSnapshots(bucketSnapshots)
		if len(bucketSnapshots) > 1 && closeOrderMatchesSubmitted(deductQty, tuition, selectedSnap.lessonModel, bucketRemQty, bucketRemTuition) {
			targetSnapshots = bucketSnapshots
		} else {
			return 0, closeOrderMismatchError(deductQty, tuition, selectedSnap.lessonModel, selectedSnap.remQty, selectedSnap.remTuition)
		}
	}

	now := time.Now()
	flowID := int64(0)
	closeOrderID, err := repo.createCloseTuitionAccountOrderTx(ctx, tx, instID, operatorID, selectedSnap, quantity, freeQuantity, tuition, remark, now)
	if err != nil {
		return 0, err
	}
	for idx, snap := range targetSnapshots {
		_ = idx
		insertedFlowID, closeErr := repo.closeTuitionAccountSnapshotTx(ctx, tx, instID, operatorID, snap, closeOrderID, now)
		if closeErr != nil {
			return 0, closeErr
		}
		if flowID == 0 && insertedFlowID > 0 {
			flowID = insertedFlowID
		}
	}

	if err := repo.closeRelatedOneToOneClassesByDeductCourseTx(ctx, tx, instID, operatorID, selectedSnap.studentID, selectedSnap.courseID); err != nil {
		return 0, err
	}

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
	`, model.InstStudentStatusHistory, operatorID, selectedSnap.studentID, instID, model.InstStudentStatusEnrolled); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return closeOrderID, nil
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
