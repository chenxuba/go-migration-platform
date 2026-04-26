package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

type closeTuitionAccountOrderRow struct {
	id                 int64
	flowSourceID       int64
	tuitionAccountID   int64
	studentID          int64
	courseID           int64
	lessonChargingMode int
	quantity           float64
	freeQuantity       float64
	tuition            float64
	remark             string
	status             int
	closeTime          sql.NullTime
	expireDate         sql.NullTime
	orderID            int64
	orderType          int
	arrearAmountTotal  float64
	badDebtAmountTotal float64
	lessonName         string
	lessonType         int
}

type closeTuitionAccountFlowRow struct {
	flowID              int64
	tuitionAccountID    int64
	studentID           int64
	courseID            int64
	orderID             int64
	orderCourseDetailID int64
	lessonType          sql.NullInt64
	lessonChargingMode  int
	orderNumber         string
	quantity            float64
	tuition             float64
	usedQuantity        float64
	remainQuantity      float64
	usedTuition         float64
	remainTuition       float64
	confirmedTuition    float64
	enableExpireTime    int
	validDate           sql.NullTime
	endDate             sql.NullTime
	createTime          sql.NullTime
}

type restoredTimeSlotOrderDetailRange struct {
	orderID             int64
	orderCourseDetailID int64
	validDate           time.Time
	endDate             time.Time
}

type tuitionAccountSubAccountRow struct {
	id                   int64
	createdTime          sql.NullTime
	activedAt            sql.NullTime
	validDate            sql.NullTime
	endDate              sql.NullTime
	remainQuantity       float64
	rawStatus            int
	lessonChargingMode   int
	totalQuantity        float64
	tuition              float64
	totalTuition         float64
	orderID              int64
	orderType            int
	orderStatus          int
	unitPrice            float64
	paidTuition          float64
	shouldTuition        float64
	arrearTuition        float64
	chargeAgainstTuition float64
	transferredTuition   float64
	paidRemaining        float64
	usedTuition          float64
}

const refundAutoCloseRevertIDPrefix = "refund:"
const refundTuitionAccountOrderStatusCompleted = 2

func buildRefundAutoCloseRevertID(refundOrderID int64) string {
	return fmt.Sprintf("%s%d", refundAutoCloseRevertIDPrefix, refundOrderID)
}

func parseRefundAutoCloseRevertID(raw string) (int64, bool, error) {
	value := strings.TrimSpace(raw)
	if !strings.HasPrefix(value, refundAutoCloseRevertIDPrefix) {
		return 0, false, nil
	}
	id, err := strconv.ParseInt(strings.TrimSpace(strings.TrimPrefix(value, refundAutoCloseRevertIDPrefix)), 10, 64)
	if err != nil || id <= 0 {
		return 0, true, errors.New("closeTuitionAccountOrderId 无效")
	}
	return id, true, nil
}

func ensureCloseTuitionAccountOrderTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS close_tuition_account_order (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			uuid VARCHAR(64) NULL,
			version BIGINT NOT NULL DEFAULT 0,
			inst_id BIGINT NOT NULL,
			flow_source_id BIGINT NULL,
			tuition_account_id BIGINT NOT NULL DEFAULT 0,
			student_id BIGINT NOT NULL DEFAULT 0,
			course_id BIGINT NOT NULL DEFAULT 0,
			lesson_charging_mode INT NOT NULL DEFAULT 0,
			quantity DECIMAL(18,2) NOT NULL DEFAULT 0,
			free_quantity DECIMAL(18,2) NOT NULL DEFAULT 0,
			tuition DECIMAL(18,2) NOT NULL DEFAULT 0,
			remark VARCHAR(500) NOT NULL DEFAULT '',
			status INT NOT NULL DEFAULT 1,
			close_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			revert_valid_start_date DATETIME NULL DEFAULT NULL,
			reverted_time DATETIME NULL DEFAULT NULL,
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT NOT NULL DEFAULT 0,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uk_close_tuition_account_order_source (inst_id, flow_source_id),
			KEY idx_close_tuition_account_order_account (inst_id, tuition_account_id, del_flag),
			KEY idx_close_tuition_account_order_student_course (inst_id, student_id, course_id, del_flag),
			KEY idx_close_tuition_account_order_status (inst_id, status, close_time)
		)
	`)
	if err != nil {
		return err
	}
	return backfillCloseTuitionAccountOrders(ctx, db)
}

func backfillCloseTuitionAccountOrders(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
		INSERT INTO close_tuition_account_order (
			uuid, version, inst_id, flow_source_id, tuition_account_id, student_id, course_id, lesson_charging_mode,
			quantity, free_quantity, tuition, remark, status, close_time,
			create_id, create_time, update_id, update_time, del_flag
		)
		SELECT
			UUID(), 0,
			agg.inst_id,
			agg.flow_source_id,
			agg.tuition_account_id,
			agg.student_id,
			agg.course_id,
			agg.lesson_charging_mode,
			agg.quantity,
			0,
			agg.tuition,
			'',
			CASE WHEN IFNULL(rev.revoked_count, 0) > 0 THEN ? ELSE ? END,
			agg.close_time,
			agg.create_id,
			agg.create_time,
			CASE WHEN IFNULL(rev.revoked_count, 0) > 0 THEN rev.update_id ELSE agg.update_id END,
			CASE WHEN IFNULL(rev.revoked_count, 0) > 0 THEN rev.update_time ELSE agg.update_time END,
			0
		FROM (
			SELECT
				taf.inst_id,
				taf.source_id AS flow_source_id,
				MIN(taf.tuition_account_id) AS tuition_account_id,
				MIN(taf.student_id) AS student_id,
				MIN(taf.product_id) AS course_id,
				MAX(IFNULL(taf.lesson_charging_mode, 0)) AS lesson_charging_mode,
				SUM(ABS(IFNULL(taf.quantity, 0))) AS quantity,
				SUM(ABS(IFNULL(taf.tuition, 0))) AS tuition,
				MIN(taf.created_time) AS close_time,
				MIN(IFNULL(taf.create_id, 0)) AS create_id,
				MIN(IFNULL(taf.create_time, taf.created_time)) AS create_time,
				MAX(IFNULL(taf.update_id, 0)) AS update_id,
				MAX(IFNULL(taf.update_time, taf.created_time)) AS update_time
			FROM tuition_account_flow taf
			WHERE taf.source_type = ?
			  AND taf.del_flag = 0
			  AND IFNULL(taf.source_id, 0) > 0
			GROUP BY taf.inst_id, taf.source_id
		) agg
		LEFT JOIN (
			SELECT
				inst_id,
				source_id,
				COUNT(*) AS revoked_count,
				MAX(IFNULL(create_id, 0)) AS update_id,
				MAX(created_time) AS update_time
			FROM tuition_account_flow
			WHERE source_type = ?
			  AND del_flag = 0
			  AND IFNULL(source_id, 0) > 0
			GROUP BY inst_id, source_id
		) rev
		  ON rev.inst_id = agg.inst_id
		 AND rev.source_id = agg.flow_source_id
		LEFT JOIN close_tuition_account_order existing
		  ON existing.inst_id = agg.inst_id
		 AND existing.flow_source_id = agg.flow_source_id
		 AND existing.del_flag = 0
		WHERE existing.id IS NULL
	`, model.CloseTuitionAccountOrderStatusRevoked, model.CloseTuitionAccountOrderStatusClosed, model.TuitionAccountFlowSourceManualCloseCourse, model.TuitionAccountFlowSourceRevokeGraduate); err != nil {
		return err
	}
	return nil
}

func (repo *Repository) createCloseTuitionAccountOrderTx(
	ctx context.Context,
	tx *sql.Tx,
	instID, operatorID int64,
	selected closeTuitionAccountSnapshot,
	quantity, freeQuantity, tuition float64,
	remark string,
	closeTime time.Time,
) (int64, error) {
	res, err := tx.ExecContext(ctx, `
		INSERT INTO close_tuition_account_order (
			uuid, version, inst_id, flow_source_id, tuition_account_id, student_id, course_id, lesson_charging_mode,
			quantity, free_quantity, tuition, remark, status, close_time,
			create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, NULL, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, NOW(), ?, NOW(), 0
		)
	`,
		instID,
		selected.id,
		selected.studentID,
		selected.courseID,
		selected.flowChargingMode(),
		closeOrderRoundMoney(quantity),
		closeOrderRoundMoney(freeQuantity),
		closeOrderRoundMoney(tuition),
		strings.TrimSpace(remark),
		model.CloseTuitionAccountOrderStatusClosed,
		closeTime,
		operatorID,
		operatorID,
	)
	if err != nil {
		return 0, err
	}
	closeOrderID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE close_tuition_account_order
		SET flow_source_id = ?
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, closeOrderID, closeOrderID, instID); err != nil {
		return 0, err
	}
	return closeOrderID, nil
}

func (repo *Repository) listCloseOrderBucketAccountIDsTx(ctx context.Context, tx *sql.Tx, instID, tuitionAccountID int64) (closeTuitionAccountSnapshot, []int64, error) {
	selected, err := repo.loadCloseTuitionAccountSnapshotTx(ctx, tx, instID, tuitionAccountID)
	if err != nil {
		return closeTuitionAccountSnapshot{}, nil, err
	}
	bucket, err := repo.loadOneToOneTuitionBucketTx(ctx, tx, instID, tuitionAccountID)
	if err != nil {
		return closeTuitionAccountSnapshot{}, nil, err
	}
	ids, err := repo.ListTuitionAccountIDsForStudentCourseBucketAllStatuses(
		ctx, tx, instID, selected.studentID, selected.courseID, bucket.teachMethod, bucket.lessonModelCode,
	)
	if err != nil {
		return closeTuitionAccountSnapshot{}, nil, err
	}
	if len(ids) == 0 {
		ids = []int64{tuitionAccountID}
	}
	return selected, ids, nil
}

func buildPlaceholders(count int) string {
	if count <= 0 {
		return ""
	}
	list := make([]string, 0, count)
	for i := 0; i < count; i++ {
		list = append(list, "?")
	}
	return strings.Join(list, ",")
}

func (repo *Repository) loadLatestClosableCloseOrderTx(ctx context.Context, tx *sql.Tx, instID int64, accountIDs []int64) (closeTuitionAccountOrderRow, error) {
	if len(accountIDs) == 0 {
		return closeTuitionAccountOrderRow{}, sql.ErrNoRows
	}
	args := []any{instID, model.TuitionAccountFlowSourceManualCloseCourse}
	for _, id := range accountIDs {
		args = append(args, id)
	}
	query := `
		SELECT
			co.id,
			IFNULL(co.flow_source_id, 0),
			IFNULL(co.tuition_account_id, 0),
			IFNULL(co.student_id, 0),
			IFNULL(co.course_id, 0),
			IFNULL(co.lesson_charging_mode, 0),
			IFNULL(co.quantity, 0),
			IFNULL(co.free_quantity, 0),
			IFNULL(co.tuition, 0),
			IFNULL(co.remark, ''),
			IFNULL(co.status, 0),
			co.close_time,
			IFNULL(ta.expire_time, NULL),
			IFNULL(ta.order_id, 0),
			IFNULL(so.order_type, 0),
			CASE
				WHEN IFNULL(so.is_bad_debt, 0) = 0 AND IFNULL(so.order_real_amount, 0) > IFNULL(pay.paid_amount, 0)
				THEN IFNULL(so.order_real_amount, 0) - IFNULL(pay.paid_amount, 0)
				ELSE 0
			END AS arrear_amount_total,
			CASE
				WHEN IFNULL(so.is_bad_debt, 0) = 1 THEN IFNULL(so.bad_debt_amount, 0)
				ELSE 0
			END AS bad_debt_amount_total,
			IFNULL(ic.name, ''),
			IFNULL(ic.teach_method, 0)
		FROM close_tuition_account_order co
		INNER JOIN (
			SELECT DISTINCT source_id
			FROM tuition_account_flow
			WHERE inst_id = ?
			  AND source_type = ?
			  AND del_flag = 0
			  AND tuition_account_id IN (` + buildPlaceholders(len(accountIDs)) + `)
		) src ON src.source_id = co.flow_source_id
		LEFT JOIN tuition_account ta ON ta.id = co.tuition_account_id AND ta.inst_id = co.inst_id AND ta.del_flag = 0
		LEFT JOIN sale_order so ON so.id = ta.order_id AND so.del_flag = 0
		LEFT JOIN inst_course ic ON ic.id = co.course_id AND ic.del_flag = 0
		LEFT JOIN (
			SELECT order_id, SUM(pay_amount) AS paid_amount
			FROM sale_order_pay_detail
			WHERE del_flag = 0
			GROUP BY order_id
		) pay ON pay.order_id = ta.order_id
		WHERE co.inst_id = ?
		  AND co.del_flag = 0
		  AND co.status = ?
		ORDER BY co.close_time DESC, co.id DESC
		LIMIT 1
	`
	args = append(args, instID, model.CloseTuitionAccountOrderStatusClosed)

	var row closeTuitionAccountOrderRow
	err := tx.QueryRowContext(ctx, query, args...).Scan(
		&row.id,
		&row.flowSourceID,
		&row.tuitionAccountID,
		&row.studentID,
		&row.courseID,
		&row.lessonChargingMode,
		&row.quantity,
		&row.freeQuantity,
		&row.tuition,
		&row.remark,
		&row.status,
		&row.closeTime,
		&row.expireDate,
		&row.orderID,
		&row.orderType,
		&row.arrearAmountTotal,
		&row.badDebtAmountTotal,
		&row.lessonName,
		&row.lessonType,
	)
	return row, err
}

func (repo *Repository) loadLatestRefundAutoCloseOrderTx(ctx context.Context, tx *sql.Tx, instID int64, accountIDs []int64) (closeTuitionAccountOrderRow, error) {
	if len(accountIDs) == 0 {
		return closeTuitionAccountOrderRow{}, sql.ErrNoRows
	}
	args := []any{instID}
	for _, id := range accountIDs {
		args = append(args, id)
	}
	query := `
		SELECT
			ro.id,
			IFNULL(ro.tuition_account_id, 0),
			IFNULL(ro.student_id, 0),
			IFNULL(ro.course_id, 0),
			IFNULL(MAX(roi.lesson_charging_mode), 0),
			IFNULL(ro.remark, ''),
			COALESCE(ro.completed_time, ro.update_time, ro.create_time),
			NULL,
			IFNULL(ro.sale_order_id, 0),
			IFNULL(so.order_type, 0),
			IFNULL(ic.name, ''),
			IFNULL(ic.teach_method, 0)
		FROM refund_tuition_account_order ro
		INNER JOIN refund_tuition_account_order_item roi
			ON roi.refund_order_id = ro.id
			AND roi.inst_id = ro.inst_id
			AND roi.del_flag = 0
		INNER JOIN tuition_account ta
			ON ta.id = roi.tuition_account_id
			AND ta.inst_id = ro.inst_id
			AND ta.del_flag = 0
		LEFT JOIN sale_order so ON so.id = ro.sale_order_id AND so.del_flag = 0
		LEFT JOIN inst_course ic ON ic.id = ro.course_id AND ic.del_flag = 0
		WHERE ro.inst_id = ?
		  AND ro.del_flag = 0
		  AND IFNULL(ro.auto_close_tuition, 0) = 1
		  AND IFNULL(ro.status, 0) = ?
		  AND IFNULL(ta.status, 0) = ?
		  AND roi.tuition_account_id IN (` + buildPlaceholders(len(accountIDs)) + `)
		GROUP BY ro.id, ro.tuition_account_id, ro.student_id, ro.course_id, ro.remark, ro.completed_time, ro.update_time, ro.create_time, ro.sale_order_id, so.order_type, ic.name, ic.teach_method
		ORDER BY COALESCE(ro.completed_time, ro.update_time, ro.create_time) DESC, ro.id DESC
		LIMIT 1
	`
	args = append([]any{instID, refundTuitionAccountOrderStatusCompleted, model.TuitionAccountStatusClosed}, args[1:]...)

	var row closeTuitionAccountOrderRow
	err := tx.QueryRowContext(ctx, query, args...).Scan(
		&row.id,
		&row.tuitionAccountID,
		&row.studentID,
		&row.courseID,
		&row.lessonChargingMode,
		&row.remark,
		&row.closeTime,
		&row.expireDate,
		&row.orderID,
		&row.orderType,
		&row.lessonName,
		&row.lessonType,
	)
	if err != nil {
		return closeTuitionAccountOrderRow{}, err
	}
	row.status = model.CloseTuitionAccountOrderStatusClosed
	return row, nil
}

func (repo *Repository) listCloseFlowsByOrderTx(ctx context.Context, tx *sql.Tx, instID int64, flowSourceID int64) ([]closeTuitionAccountFlowRow, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT
			taf.id,
			taf.tuition_account_id,
			taf.student_id,
			taf.product_id,
			IFNULL(ta.order_id, 0),
			IFNULL(ta.order_course_detail_id, 0),
			taf.lesson_type,
			IFNULL(taf.lesson_charging_mode, 0),
			IFNULL(taf.order_number, ''),
			IFNULL(taf.quantity, 0),
			IFNULL(taf.tuition, 0),
			IFNULL(ta.used_quantity, 0),
			IFNULL(ta.remaining_quantity, 0),
			IFNULL(ta.used_tuition, 0),
			IFNULL(ta.remaining_tuition, 0),
			IFNULL(ta.confirmed_tuition, 0),
			IFNULL(ta.enable_expire_time, 0),
			ta.valid_date,
			ta.end_date,
			ta.create_time
		FROM tuition_account_flow taf
		INNER JOIN tuition_account ta
			ON ta.id = taf.tuition_account_id
			AND ta.inst_id = taf.inst_id
			AND ta.del_flag = 0
		WHERE taf.inst_id = ?
		  AND taf.source_type = ?
		  AND taf.source_id = ?
		  AND taf.del_flag = 0
		ORDER BY ta.create_time ASC, ta.id ASC, taf.id ASC
		FOR UPDATE
	`, instID, model.TuitionAccountFlowSourceManualCloseCourse, flowSourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]closeTuitionAccountFlowRow, 0, 4)
	for rows.Next() {
		var item closeTuitionAccountFlowRow
		if err := rows.Scan(
			&item.flowID,
			&item.tuitionAccountID,
			&item.studentID,
			&item.courseID,
			&item.orderID,
			&item.orderCourseDetailID,
			&item.lessonType,
			&item.lessonChargingMode,
			&item.orderNumber,
			&item.quantity,
			&item.tuition,
			&item.usedQuantity,
			&item.remainQuantity,
			&item.usedTuition,
			&item.remainTuition,
			&item.confirmedTuition,
			&item.enableExpireTime,
			&item.validDate,
			&item.endDate,
			&item.createTime,
		); err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, rows.Err()
}

func mergeRestoredTimeSlotOrderDetailRange(detailOrder []int64, detailRanges map[int64]*restoredTimeSlotOrderDetailRange, orderID, detailID int64, validDate, endDate time.Time) []int64 {
	if detailID <= 0 || orderID <= 0 {
		return detailOrder
	}
	if current, ok := detailRanges[detailID]; ok {
		if validDate.Before(current.validDate) {
			current.validDate = validDate
		}
		if endDate.After(current.endDate) {
			current.endDate = endDate
		}
		return detailOrder
	}
	detailRanges[detailID] = &restoredTimeSlotOrderDetailRange{
		orderID:             orderID,
		orderCourseDetailID: detailID,
		validDate:           validDate,
		endDate:             endDate,
	}
	return append(detailOrder, detailID)
}

func (repo *Repository) updateRestoredTimeSlotOrderDetailsTx(ctx context.Context, tx *sql.Tx, operatorID int64, detailOrder []int64, detailRanges map[int64]*restoredTimeSlotOrderDetailRange) error {
	for _, detailID := range detailOrder {
		rng := detailRanges[detailID]
		if rng == nil {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE sale_order_course_detail
			SET has_valid_date = 1,
			    valid_date = ?,
			    end_date = ?,
			    update_id = ?,
			    update_time = NOW()
			WHERE id = ? AND order_id = ? AND del_flag = 0
		`, rng.validDate, rng.endDate, operatorID, rng.orderCourseDetailID, rng.orderID); err != nil {
			return err
		}
	}
	return nil
}

func (repo *Repository) syncRestoredTimeSlotAutoIncomeTx(ctx context.Context, tx *sql.Tx, instID, operatorID, studentID int64, detailOrder []int64, detailRanges map[int64]*restoredTimeSlotOrderDetailRange, now time.Time) error {
	if len(detailOrder) == 0 {
		return nil
	}
	details := make([]orderCourseDetail, 0, len(detailOrder))
	detailByID := make(map[int64]orderCourseDetail, len(detailOrder))
	for _, detailID := range detailOrder {
		detail, err := repo.loadOrderCourseDetailTx(ctx, tx, detailID)
		if err != nil {
			return err
		}
		details = append(details, detail)
		detailByID[detailID] = detail
	}
	quotationMap, err := repo.getCourseQuotationsByIDsTx(ctx, tx, collectOrderDetailQuoteIDs(details))
	if err != nil {
		return err
	}
	for _, detailID := range detailOrder {
		rng := detailRanges[detailID]
		if rng == nil {
			continue
		}
		detail, ok := detailByID[detailID]
		if !ok || !detail.QuoteID.Valid {
			continue
		}
		quotation, ok := quotationMap[detail.QuoteID.Int64]
		if !ok {
			continue
		}
		if err := repo.initializeTimeSlotIncomeTx(ctx, tx, instID, operatorID, rng.orderID, studentID, detail, quotation, now); err != nil {
			return err
		}
	}
	return nil
}

func (repo *Repository) loadSubAccountDateInfoRowsTx(ctx context.Context, tx *sql.Tx, instID int64, accountIDs []int64) ([]tuitionAccountSubAccountRow, error) {
	if len(accountIDs) == 0 {
		return []tuitionAccountSubAccountRow{}, nil
	}
	args := []any{instID}
	for _, id := range accountIDs {
		args = append(args, id)
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT
			ta.id,
			ta.create_time,
			ta.valid_date,
			ta.valid_date,
			ta.end_date,
			IFNULL(ta.remaining_quantity, 0),
			IFNULL(ta.status, 0),
			IFNULL(icq.lesson_model, 0),
			CASE
				WHEN IFNULL(icq.lesson_model, 0) IN (3, 4) THEN IFNULL(ta.total_tuition, 0)
				WHEN IFNULL(ta.total_quantity, 0) > 0 THEN IFNULL(ta.total_quantity, 0)
				ELSE 0
			END AS total_days,
			IFNULL(ta.remaining_tuition, 0),
			IFNULL(ta.total_tuition, 0),
			IFNULL(ta.order_id, 0),
			IFNULL(so.order_type, 0),
			IFNULL(so.order_status, 0),
			CASE
				WHEN IFNULL(icq.lesson_model, 0) IN (3, 4) THEN 0
				WHEN IFNULL(ta.total_quantity, 0) > 0 THEN round(IFNULL(ta.total_tuition, 0) / NULLIF(ta.total_quantity, 0), 2)
				ELSE 0
			END AS unit_price,
			IFNULL(ta.paid_tuition, 0),
			IFNULL(ta.total_tuition, 0),
			CASE
				WHEN `+orderArrearActiveStatusSQL("so.order_status")+` AND IFNULL(so.is_bad_debt, 0) = 0 AND IFNULL(so.order_real_amount, 0) > IFNULL(pay.paid_amount, 0)
				THEN IFNULL(so.order_real_amount, 0) - IFNULL(pay.paid_amount, 0)
				ELSE 0
			END AS arrear_tuition,
			0 AS charge_against_tuition,
			0 AS transferred_tuition,
			IFNULL(ta.remaining_tuition, 0),
			IFNULL(ta.used_tuition, 0)
		FROM tuition_account ta
		INNER JOIN inst_course ic ON ic.id = ta.course_id AND ic.del_flag = 0
		LEFT JOIN sale_order so ON so.id = ta.order_id AND so.del_flag = 0
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
		LEFT JOIN (
			SELECT order_id, SUM(pay_amount) AS paid_amount
			FROM sale_order_pay_detail
			WHERE del_flag = 0
			GROUP BY order_id
		) pay ON pay.order_id = ta.order_id
		WHERE ta.inst_id = ?
		  AND ta.del_flag = 0
		  AND ta.id IN (`+buildPlaceholders(len(accountIDs))+`)
		ORDER BY IFNULL(ta.valid_date, ta.create_time) ASC, ta.create_time ASC, ta.id ASC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]tuitionAccountSubAccountRow, 0, len(accountIDs))
	for rows.Next() {
		var item tuitionAccountSubAccountRow
		if err := rows.Scan(
			&item.id,
			&item.createdTime,
			&item.activedAt,
			&item.validDate,
			&item.endDate,
			&item.remainQuantity,
			&item.rawStatus,
			&item.lessonChargingMode,
			&item.totalQuantity,
			&item.tuition,
			&item.totalTuition,
			&item.orderID,
			&item.orderType,
			&item.orderStatus,
			&item.unitPrice,
			&item.paidTuition,
			&item.shouldTuition,
			&item.arrearTuition,
			&item.chargeAgainstTuition,
			&item.transferredTuition,
			&item.paidRemaining,
			&item.usedTuition,
		); err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, rows.Err()
}

func convertSubAccountDateInfoRows(rows []tuitionAccountSubAccountRow) []model.TuitionAccountSubAccountDateInfoItem {
	out := make([]model.TuitionAccountSubAccountDateInfoItem, 0, len(rows))
	for _, row := range rows {
		item := model.TuitionAccountSubAccountDateInfoItem{
			ID:                     strconv.FormatInt(row.id, 10),
			Quantity:               closeOrderRoundMoney(row.remainQuantity),
			RemainDays:             closeOrderRoundMoney(row.remainQuantity),
			RawStatus:              row.rawStatus,
			Status:                 row.rawStatus,
			IsFree:                 closeOrderAlmostEqual(row.totalTuition, 0),
			TotalDays:              closeOrderRoundMoney(row.totalQuantity),
			Tuition:                closeOrderRoundMoney(row.tuition),
			TotalTuition:           closeOrderRoundMoney(row.totalTuition),
			SourceType:             row.orderType,
			AccountSourceType:      row.orderType,
			OrderID:                strconv.FormatInt(row.orderID, 10),
			OrderStatus:            row.orderStatus,
			SourceID:               strconv.FormatInt(row.orderID, 10),
			UnitPrice:              closeOrderRoundMoney(row.unitPrice),
			PaidTuition:            closeOrderRoundMoney(row.paidTuition),
			ShouldTuition:          closeOrderRoundMoney(row.shouldTuition),
			ArrearTuition:          closeOrderRoundMoney(row.arrearTuition),
			ChargeAgainstTuition:   closeOrderRoundMoney(row.chargeAgainstTuition),
			TransferredTuition:     closeOrderRoundMoney(row.transferredTuition),
			PaidRemaining:          closeOrderRoundMoney(row.paidRemaining),
			UsedTuition:            closeOrderRoundMoney(row.usedTuition),
			ExpiredToClearQuantity: false,
		}
		if row.createdTime.Valid {
			t := row.createdTime.Time
			item.CreatedTime = &t
		}
		if row.activedAt.Valid {
			t := row.activedAt.Time
			item.ActivedAt = &t
			item.StartTime = &t
			item.StartDate = &t
		}
		if row.endDate.Valid {
			t := row.endDate.Time
			item.EndDate = &t
			item.ExpireDate = &t
		}
		out = append(out, item)
	}
	return out
}

func buildPreviewSubPeriods(rows []tuitionAccountSubAccountRow) []model.RevertCloseTuitionAccountSubPeriod {
	out := make([]model.RevertCloseTuitionAccountSubPeriod, 0, len(rows))
	seen := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		period := model.RevertCloseTuitionAccountSubPeriod{}
		if row.validDate.Valid {
			t := row.validDate.Time
			period.StartDate = &t
		}
		if row.endDate.Valid {
			t := row.endDate.Time
			period.EndDate = &t
		}
		if period.StartDate == nil && period.EndDate == nil {
			continue
		}
		key := fmt.Sprintf("%s|%s", nullableDateKey(period.StartDate), nullableDateKey(period.EndDate))
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, period)
	}
	return out
}

func buildPreviewSubPeriodsFromCloseFlows(rows []closeTuitionAccountFlowRow) []model.RevertCloseTuitionAccountSubPeriod {
	out := make([]model.RevertCloseTuitionAccountSubPeriod, 0, len(rows))
	for _, row := range rows {
		period := model.RevertCloseTuitionAccountSubPeriod{
			Quantity: closeOrderRoundMoney(row.quantity),
			IsFree:   closeOrderAlmostEqual(row.tuition, 0),
		}
		if row.validDate.Valid {
			t := row.validDate.Time
			period.StartDate = &t
		}
		if row.endDate.Valid {
			t := row.endDate.Time
			period.EndDate = &t
		}
		if period.StartDate == nil && period.EndDate == nil {
			continue
		}
		out = append(out, period)
	}
	return out
}

func nullableDateKey(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format("2006-01-02")
}

// minCloseFlowValidDate 结课流水对应账户在结课前合约首日（多子账户取最早 valid_date）。
func minCloseFlowValidDate(flows []closeTuitionAccountFlowRow) (time.Time, bool) {
	var out time.Time
	var ok bool
	for _, f := range flows {
		if !f.validDate.Valid {
			continue
		}
		t := startOfDayTime(f.validDate.Time)
		if !ok || t.Before(out) {
			out = t
			ok = true
		}
	}
	return out, ok
}

func isSameCalendarDay(a, b time.Time) bool {
	return startOfDayTime(a).Equal(startOfDayTime(b))
}

func parseRequiredDate(value string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, errors.New("请选择日期")
	}
	parsed, err := time.ParseInLocation("2006-01-02", trimmed, time.Local)
	if err != nil {
		return time.Time{}, errors.New("日期格式不正确")
	}
	return startOfDayTime(parsed), nil
}

func parseOptionalDate(value string) (time.Time, bool, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, false, nil
	}
	parsed, err := parseRequiredDate(trimmed)
	if err != nil {
		return time.Time{}, false, err
	}
	return parsed, true, nil
}

func (repo *Repository) reopenRelatedOneToOneClassesByDeductCourseTx(ctx context.Context, tx *sql.Tx, instID, operatorID, studentID, courseID int64) error {
	if studentID <= 0 || courseID <= 0 {
		return nil
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
		  AND tcs.class_student_status = ?
	`, model.TeachingClassTypeOneToOne, model.TeachingClassStudentStatusStudying, operatorID, instID, studentID, courseID, model.TeachingClassStudentStatusClosed); err != nil {
		return err
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
		  AND tc.inst_id = ?
		  AND tcs.student_id = ?
		  AND ta_eff.course_id = ?
		  AND tc.status = ?
	`, model.TeachingClassStatusActive, operatorID, model.TeachingClassTypeOneToOne, instID, studentID, courseID, model.TeachingClassStatusClosed); err != nil {
		return err
	}
	return nil
}

func (repo *Repository) reopenRelatedGroupClassesByDeductCourseTx(ctx context.Context, tx *sql.Tx, instID, operatorID, studentID, courseID int64) error {
	if studentID <= 0 || courseID <= 0 {
		return nil
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT tcs.teaching_class_id, tcs.id
		FROM teaching_class_student tcs
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
		WHERE tcs.inst_id = ?
		  AND tcs.del_flag = 0
		  AND tcs.student_id = ?
		  AND ta_eff.course_id = ?
		  AND tcs.class_student_status = ?
	`, model.TeachingClassTypeNormal, instID, studentID, courseID, model.TeachingClassStudentStatusClosed)
	if err != nil {
		return err
	}
	type statusChangeTarget struct {
		classID      int64
		membershipID int64
	}
	targets := make([]statusChangeTarget, 0, 4)
	for rows.Next() {
		var item statusChangeTarget
		if err := rows.Scan(&item.classID, &item.membershipID); err != nil {
			rows.Close()
			return err
		}
		targets = append(targets, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	_, err = tx.ExecContext(ctx, `
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
		  AND tcs.class_student_status = ?
	`, model.TeachingClassTypeNormal, model.TeachingClassStudentStatusStudying, operatorID, instID, studentID, courseID, model.TeachingClassStudentStatusClosed)
	if err != nil {
		return err
	}

	operateTime := time.Now()
	for _, item := range targets {
		if err := repo.syncGroupClassStudentStatusChangeHistoryTx(
			ctx,
			tx,
			instID,
			operatorID,
			item.classID,
			item.membershipID,
			model.TeachingClassStudentStatusStudying,
			operateTime,
		); err != nil {
			return err
		}
	}
	return nil
}

func (repo *Repository) GetTuitionAccountSubAccountDateInfo(ctx context.Context, instID int64, dto model.TuitionAccountSubAccountDateInfoQueryDTO) (model.TuitionAccountSubAccountDateInfoResult, error) {
	tuitionAccountID, err := strconv.ParseInt(strings.TrimSpace(dto.TuitionAccountID), 10, 64)
	if err != nil || tuitionAccountID <= 0 {
		return model.TuitionAccountSubAccountDateInfoResult{}, errors.New("tuitionAccountId 无效")
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return model.TuitionAccountSubAccountDateInfoResult{}, err
	}
	defer tx.Rollback()

	selected, accountIDs, err := repo.listCloseOrderBucketAccountIDsTx(ctx, tx, instID, tuitionAccountID)
	if err != nil {
		return model.TuitionAccountSubAccountDateInfoResult{}, err
	}
	_ = selected
	rows, err := repo.loadSubAccountDateInfoRowsTx(ctx, tx, instID, accountIDs)
	if err != nil {
		return model.TuitionAccountSubAccountDateInfoResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return model.TuitionAccountSubAccountDateInfoResult{}, err
	}
	return model.TuitionAccountSubAccountDateInfoResult{
		List: convertSubAccountDateInfoRows(rows),
	}, nil
}

func (repo *Repository) GetRevertCloseTuitionAccountPreview(ctx context.Context, instID int64, dto model.RevertCloseTuitionAccountPreviewQueryDTO) (model.RevertCloseTuitionAccountPreview, error) {
	tuitionAccountID, err := strconv.ParseInt(strings.TrimSpace(dto.TuitionAccountID), 10, 64)
	if err != nil || tuitionAccountID <= 0 {
		return model.RevertCloseTuitionAccountPreview{}, errors.New("tuitionAccountId 无效")
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return model.RevertCloseTuitionAccountPreview{}, err
	}
	defer tx.Rollback()

	selected, accountIDs, err := repo.listCloseOrderBucketAccountIDsTx(ctx, tx, instID, tuitionAccountID)
	if err != nil {
		return model.RevertCloseTuitionAccountPreview{}, err
	}
	orderRow, err := repo.loadLatestClosableCloseOrderTx(ctx, tx, instID, accountIDs)
	isRefundAutoClose := false
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			orderRow, err = repo.loadLatestRefundAutoCloseOrderTx(ctx, tx, instID, accountIDs)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return model.RevertCloseTuitionAccountPreview{}, errors.New("暂无可撤销的结课记录")
				}
				return model.RevertCloseTuitionAccountPreview{}, err
			}
			isRefundAutoClose = true
		} else {
			return model.RevertCloseTuitionAccountPreview{}, err
		}
	}
	closeFlows := []closeTuitionAccountFlowRow{}
	if !isRefundAutoClose {
		closeFlows, err = repo.listCloseFlowsByOrderTx(ctx, tx, instID, orderRow.flowSourceID)
		if err != nil {
			return model.RevertCloseTuitionAccountPreview{}, err
		}
	}
	subRows, err := repo.loadSubAccountDateInfoRowsTx(ctx, tx, instID, accountIDs)
	if err != nil {
		return model.RevertCloseTuitionAccountPreview{}, err
	}
	if err := tx.Commit(); err != nil {
		return model.RevertCloseTuitionAccountPreview{}, err
	}

	preview := model.RevertCloseTuitionAccountPreview{
		TuitionAccountID:           strconv.FormatInt(tuitionAccountID, 10),
		LessonName:                 firstNonEmpty(strings.TrimSpace(orderRow.lessonName), strconv.FormatInt(orderRow.courseID, 10)),
		LessonType:                 orderRow.lessonType,
		LessonChargingMode:         orderRow.lessonChargingMode,
		CloseTuitionAccountOrderID: strconv.FormatInt(orderRow.id, 10),
		Quantity:                   closeOrderRoundMoney(orderRow.quantity),
		FreeQuantity:               closeOrderRoundMoney(orderRow.freeQuantity),
		Tuition:                    closeOrderRoundMoney(orderRow.tuition),
		Remark:                     strings.TrimSpace(orderRow.remark),
		ArrearAmountTotal:          closeOrderRoundMoney(orderRow.arrearAmountTotal),
		BadDebtAmountTotal:         closeOrderRoundMoney(orderRow.badDebtAmountTotal),
		OrderID:                    strconv.FormatInt(orderRow.orderID, 10),
		OrderType:                  orderRow.orderType,
		SubTuitionAccounts:         buildPreviewSubPeriodsFromCloseFlows(closeFlows),
	}
	if isRefundAutoClose {
		preview.CloseTuitionAccountOrderID = buildRefundAutoCloseRevertID(orderRow.id)
		preview.Quantity = 0
		preview.FreeQuantity = 0
		preview.Tuition = 0
		preview.SubTuitionAccounts = []model.RevertCloseTuitionAccountSubPeriod{}
	}
	if len(preview.SubTuitionAccounts) == 0 {
		preview.SubTuitionAccounts = buildPreviewSubPeriods(subRows)
	}
	if preview.LessonType == 0 && selected.teachMethod.Valid {
		preview.LessonType = int(selected.teachMethod.Int64)
	}
	if orderRow.closeTime.Valid {
		t := orderRow.closeTime.Time
		preview.CloseTime = &t
	}
	if orderRow.expireDate.Valid {
		t := orderRow.expireDate.Time
		preview.ExpireDate = &t
	}
	return preview, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func (repo *Repository) revertRefundAutoClosedTuitionAccount(ctx context.Context, instID, operatorID, tuitionAccountID, refundOrderID int64) (int64, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	selected, accountIDs, err := repo.listCloseOrderBucketAccountIDsTx(ctx, tx, instID, tuitionAccountID)
	if err != nil {
		return 0, err
	}
	if len(accountIDs) == 0 {
		return 0, errors.New("暂无可撤销的结课记录")
	}

	args := []any{refundOrderID, instID, refundTuitionAccountOrderStatusCompleted, model.TuitionAccountStatusClosed}
	for _, id := range accountIDs {
		args = append(args, id)
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT DISTINCT ta.id
		FROM refund_tuition_account_order ro
		INNER JOIN refund_tuition_account_order_item roi
			ON roi.refund_order_id = ro.id
			AND roi.inst_id = ro.inst_id
			AND roi.del_flag = 0
		INNER JOIN tuition_account ta
			ON ta.id = roi.tuition_account_id
			AND ta.inst_id = ro.inst_id
			AND ta.del_flag = 0
		WHERE ro.id = ?
		  AND ro.inst_id = ?
		  AND ro.del_flag = 0
		  AND IFNULL(ro.auto_close_tuition, 0) = 1
		  AND IFNULL(ro.status, 0) = ?
		  AND IFNULL(ta.status, 0) = ?
		  AND roi.tuition_account_id IN (`+buildPlaceholders(len(accountIDs))+`)
		FOR UPDATE
	`, args...)
	if err != nil {
		return 0, err
	}
	revertAccountIDs := make([]int64, 0, len(accountIDs))
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return 0, err
		}
		revertAccountIDs = append(revertAccountIDs, id)
	}
	if err := rows.Close(); err != nil {
		return 0, err
	}
	if len(revertAccountIDs) == 0 {
		return 0, errors.New("暂无可撤销的结课记录")
	}

	updateArgs := []any{model.TuitionAccountStatusActive, time.Now(), operatorID, instID}
	for _, id := range revertAccountIDs {
		updateArgs = append(updateArgs, id)
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE tuition_account
		SET status = ?,
		    status_change_time = ?,
		    class_ending_time = NULL,
		    update_id = ?,
		    update_time = NOW()
		WHERE inst_id = ?
		  AND del_flag = 0
		  AND id IN (`+buildPlaceholders(len(revertAccountIDs))+`)
	`, updateArgs...); err != nil {
		return 0, err
	}

	if err := repo.reopenRelatedOneToOneClassesByDeductCourseTx(ctx, tx, instID, operatorID, selected.studentID, selected.courseID); err != nil {
		return 0, err
	}
	if err := repo.reopenRelatedGroupClassesByDeductCourseTx(ctx, tx, instID, operatorID, selected.studentID, selected.courseID); err != nil {
		return 0, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE inst_student s
		SET s.student_status = ?,
		    s.update_id = ?,
		    s.update_time = NOW()
		WHERE s.id = ? AND s.inst_id = ? AND s.del_flag = 0
	`, 1, operatorID, selected.studentID, instID); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return refundOrderID, nil
}

func (repo *Repository) RevertCloseTuitionAccount(ctx context.Context, instID, operatorID int64, dto model.RevertCloseTuitionAccountDTO) (int64, error) {
	tuitionAccountID, err := strconv.ParseInt(strings.TrimSpace(dto.TuitionAccountID), 10, 64)
	if err != nil || tuitionAccountID <= 0 {
		return 0, errors.New("tuitionAccountId 无效")
	}
	refundOrderID, isRefundAutoClose, err := parseRefundAutoCloseRevertID(dto.CloseTuitionAccountOrderID)
	if err != nil {
		return 0, err
	}
	if isRefundAutoClose {
		return repo.revertRefundAutoClosedTuitionAccount(ctx, instID, operatorID, tuitionAccountID, refundOrderID)
	}

	closeOrderID, err := strconv.ParseInt(strings.TrimSpace(dto.CloseTuitionAccountOrderID), 10, 64)
	if err != nil || closeOrderID <= 0 {
		return 0, errors.New("closeTuitionAccountOrderId 无效")
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	selected, accountIDs, err := repo.listCloseOrderBucketAccountIDsTx(ctx, tx, instID, tuitionAccountID)
	if err != nil {
		return 0, err
	}

	var orderRow closeTuitionAccountOrderRow
	err = tx.QueryRowContext(ctx, `
		SELECT
			co.id,
			IFNULL(co.flow_source_id, 0),
			IFNULL(co.tuition_account_id, 0),
			IFNULL(co.student_id, 0),
			IFNULL(co.course_id, 0),
			IFNULL(co.lesson_charging_mode, 0),
			IFNULL(co.quantity, 0),
			IFNULL(co.free_quantity, 0),
			IFNULL(co.tuition, 0),
			IFNULL(co.remark, ''),
			IFNULL(co.status, 0),
			co.close_time,
			NULL,
			0,
			0,
			0,
			0,
			IFNULL(ic.name, ''),
			IFNULL(ic.teach_method, 0)
		FROM close_tuition_account_order co
		LEFT JOIN inst_course ic ON ic.id = co.course_id AND ic.del_flag = 0
		WHERE co.id = ? AND co.inst_id = ? AND co.del_flag = 0
		LIMIT 1
		FOR UPDATE
	`, closeOrderID, instID).Scan(
		&orderRow.id,
		&orderRow.flowSourceID,
		&orderRow.tuitionAccountID,
		&orderRow.studentID,
		&orderRow.courseID,
		&orderRow.lessonChargingMode,
		&orderRow.quantity,
		&orderRow.freeQuantity,
		&orderRow.tuition,
		&orderRow.remark,
		&orderRow.status,
		&orderRow.closeTime,
		&orderRow.expireDate,
		&orderRow.orderID,
		&orderRow.orderType,
		&orderRow.arrearAmountTotal,
		&orderRow.badDebtAmountTotal,
		&orderRow.lessonName,
		&orderRow.lessonType,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("结课记录不存在")
		}
		return 0, err
	}
	if orderRow.status == model.CloseTuitionAccountOrderStatusRevoked {
		return 0, errors.New("该结课记录已撤销")
	}

	accountIDSet := make(map[int64]struct{}, len(accountIDs))
	for _, id := range accountIDs {
		accountIDSet[id] = struct{}{}
	}
	flows, err := repo.listCloseFlowsByOrderTx(ctx, tx, instID, orderRow.flowSourceID)
	if err != nil {
		return 0, err
	}
	if len(flows) == 0 {
		return 0, errors.New("该结课记录缺少结课流水")
	}
	for _, flow := range flows {
		if _, ok := accountIDSet[flow.tuitionAccountID]; !ok {
			return 0, errors.New("当前课程账户与结课记录不匹配")
		}
	}

	expireDate, hasExpireDate, err := parseOptionalDate(dto.ExpireDate)
	if err != nil {
		return 0, err
	}
	startDateFromPayload, hasStartDate, err := parseOptionalDate(dto.StartDate)
	if err != nil {
		return 0, err
	}

	var startDate time.Time
	if orderRow.lessonChargingMode == 2 {
		if hasStartDate {
			startDate = startDateFromPayload
		} else if hasExpireDate {
			totalRestoreDays := 0
			for _, flow := range flows {
				totalRestoreDays += int(math.Round(flow.quantity))
			}
			startDate = expireDate
			if totalRestoreDays > 1 {
				startDate = expireDate.AddDate(0, 0, -(totalRestoreDays - 1))
			}
		} else {
			startDate, err = parseRequiredDate(dto.CurrentValidStartDate)
			if err != nil {
				return 0, err
			}
		}
	}
	now := time.Now()
	cursorDate := startDate
	firstRevokeFlowID := int64(0)
	origContractStart, hasOrigContractStart := minCloseFlowValidDate(flows)
	bonusConsumedFirstDay := false
	restoredDetailOrder := make([]int64, 0, 4)
	restoredDetailRanges := make(map[int64]*restoredTimeSlotOrderDetailRange)

	for _, flow := range flows {
		newUsedQty := math.Max(closeOrderRoundMoney(flow.usedQuantity-flow.quantity), 0)
		newRemainQty := closeOrderRoundMoney(flow.remainQuantity + flow.quantity)
		newUsedTuition := math.Max(closeOrderRoundMoney(flow.usedTuition-flow.tuition), 0)
		newRemainTuition := closeOrderRoundMoney(flow.remainTuition + flow.tuition)
		newConfirmed := math.Max(closeOrderRoundMoney(flow.confirmedTuition-flow.tuition), 0)

		var validDateArg any = nil
		var endDateArg any = nil
		var expireDateArg any = nil
		var restoredValidDate time.Time
		var restoredEndDate time.Time
		hasRestoredRange := false
		if orderRow.lessonChargingMode == 2 {
			restoreDays := int(math.Round(flow.quantity))
			if restoreDays > 0 {
				// 现有效开始日若与结课前合约首日同一天，第一段要把「当日已课消」计回时段长度（剩余+1 日）；
				// 若从次日等起算且无该情形，则仍按剩余天数铺区间。
				if hasOrigContractStart && !bonusConsumedFirstDay && isSameCalendarDay(cursorDate, origContractStart) {
					restoreDays++
					bonusConsumedFirstDay = true
				}
				validDate := cursorDate
				endDate := cursorDate.AddDate(0, 0, restoreDays-1)
				validDateArg = validDate
				endDateArg = endDate
				restoredValidDate = validDate
				restoredEndDate = endDate
				hasRestoredRange = true
				cursorDate = endDate.AddDate(0, 0, 1)
			}
		}
		if hasExpireDate {
			expireDateArg = expireDate
		}

		if _, err := tx.ExecContext(ctx, `
			UPDATE tuition_account
			SET used_quantity = ?,
			    remaining_quantity = ?,
			    used_tuition = ?,
			    remaining_tuition = ?,
			    confirmed_tuition = ?,
			    status = ?,
			    status_change_time = ?,
			    class_ending_time = NULL,
			    enable_expire_time = CASE
			    	WHEN ? IS NOT NULL THEN 1
			    	ELSE enable_expire_time
			    END,
			    expire_time = COALESCE(?, expire_time),
			    valid_date = COALESCE(?, valid_date),
			    end_date = COALESCE(?, end_date),
			    update_id = ?,
			    update_time = NOW()
			WHERE id = ? AND inst_id = ? AND del_flag = 0
		`, newUsedQty, newRemainQty, newUsedTuition, newRemainTuition, newConfirmed,
			model.TuitionAccountStatusActive, now, expireDateArg, expireDateArg, validDateArg, endDateArg, operatorID, flow.tuitionAccountID, instID); err != nil {
			return 0, err
		}

		var lessonTypeValue any
		if flow.lessonType.Valid {
			lessonTypeValue = flow.lessonType.Int64
		}
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
			flow.tuitionAccountID,
			flow.studentID,
			flow.courseID,
			lessonTypeValue,
			flow.lessonChargingMode,
			model.TuitionAccountFlowSourceRevokeGraduate,
			orderRow.flowSourceID,
			flow.orderNumber,
			now,
			closeOrderRoundMoney(flow.quantity),
			closeOrderRoundMoney(-flow.tuition),
			newRemainQty,
			newRemainTuition,
			operatorID,
			operatorID,
		)
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				return 0, errors.New("请勿重复提交撤销结课")
			}
			return 0, err
		}
		revokeFlowID, err := res.LastInsertId()
		if err != nil {
			return 0, err
		}
		if firstRevokeFlowID == 0 {
			firstRevokeFlowID = revokeFlowID
		}
		if hasRestoredRange {
			restoredDetailOrder = mergeRestoredTimeSlotOrderDetailRange(
				restoredDetailOrder,
				restoredDetailRanges,
				flow.orderID,
				flow.orderCourseDetailID,
				restoredValidDate,
				restoredEndDate,
			)
		}
	}

	if orderRow.lessonChargingMode == 2 {
		if err := repo.updateRestoredTimeSlotOrderDetailsTx(ctx, tx, operatorID, restoredDetailOrder, restoredDetailRanges); err != nil {
			return 0, err
		}
		if err := repo.syncRestoredTimeSlotAutoIncomeTx(ctx, tx, instID, operatorID, selected.studentID, restoredDetailOrder, restoredDetailRanges, now); err != nil {
			return 0, err
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE close_tuition_account_order
		SET status = ?,
		    reverted_time = ?,
		    revert_valid_start_date = ?,
		    update_id = ?,
		    update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, model.CloseTuitionAccountOrderStatusRevoked, now, nullableDateTime(startDate, orderRow.lessonChargingMode == 2), operatorID, closeOrderID, instID); err != nil {
		return 0, err
	}

	if err := repo.reopenRelatedOneToOneClassesByDeductCourseTx(ctx, tx, instID, operatorID, selected.studentID, selected.courseID); err != nil {
		return 0, err
	}
	if err := repo.reopenRelatedGroupClassesByDeductCourseTx(ctx, tx, instID, operatorID, selected.studentID, selected.courseID); err != nil {
		return 0, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE inst_student s
		SET s.student_status = ?,
		    s.update_id = ?,
		    s.update_time = NOW()
		WHERE s.id = ? AND s.inst_id = ? AND s.del_flag = 0
		  AND s.student_status <> ?
		  AND EXISTS (
			SELECT 1 FROM tuition_account ta
			WHERE ta.del_flag = 0 AND ta.inst_id = s.inst_id AND ta.student_id = s.id
			  AND (IFNULL(ta.remaining_quantity, 0) > 0.02 OR IFNULL(ta.remaining_tuition, 0) > 0.02)
			LIMIT 1
		  )
	`, model.InstStudentStatusEnrolled, operatorID, selected.studentID, instID, model.InstStudentStatusEnrolled); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return firstRevokeFlowID, nil
}

func nullableDateTime(value time.Time, enabled bool) any {
	if !enabled {
		return nil
	}
	return value
}

func (repo *Repository) ListCloseTuitionAccountOrders(ctx context.Context, instID int64, dto model.CloseTuitionAccountOrderRecordQueryDTO) (model.CloseTuitionAccountOrderRecordResult, error) {
	tuitionAccountID, err := strconv.ParseInt(strings.TrimSpace(dto.TuitionAccountID), 10, 64)
	if err != nil || tuitionAccountID <= 0 {
		return model.CloseTuitionAccountOrderRecordResult{}, errors.New("tuitionAccountId 无效")
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return model.CloseTuitionAccountOrderRecordResult{}, err
	}
	defer tx.Rollback()

	_, accountIDs, err := repo.listCloseOrderBucketAccountIDsTx(ctx, tx, instID, tuitionAccountID)
	if err != nil {
		return model.CloseTuitionAccountOrderRecordResult{}, err
	}
	if len(accountIDs) == 0 {
		return model.CloseTuitionAccountOrderRecordResult{List: []model.CloseTuitionAccountOrderRecordItem{}}, nil
	}

	args := []any{instID}
	for _, id := range accountIDs {
		args = append(args, id)
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT
			co.id,
			CAST(IFNULL(co.tuition_account_id, 0) AS CHAR),
			IFNULL(co.quantity, 0),
			IFNULL(co.free_quantity, 0),
			IFNULL(co.status, 0),
			CAST(IFNULL(co.update_id, 0) AS CHAR),
			IFNULL(u.nick_name, ''),
			co.update_time,
			co.close_time
		FROM close_tuition_account_order co
		INNER JOIN (
			SELECT DISTINCT source_id
			FROM tuition_account_flow
			WHERE inst_id = ?
			  AND source_type = ?
			  AND del_flag = 0
			  AND tuition_account_id IN (`+buildPlaceholders(len(accountIDs))+`)
		) src ON src.source_id = co.flow_source_id
		LEFT JOIN inst_user u ON u.id = co.update_id AND u.del_flag = 0
		WHERE co.inst_id = ?
		  AND co.del_flag = 0
		ORDER BY co.close_time DESC, co.id DESC
	`, append(append([]any{instID, model.TuitionAccountFlowSourceManualCloseCourse}, int64SliceToAny(accountIDs)...), instID)...)
	if err != nil {
		return model.CloseTuitionAccountOrderRecordResult{}, err
	}
	defer rows.Close()

	out := make([]model.CloseTuitionAccountOrderRecordItem, 0, 8)
	for rows.Next() {
		var item model.CloseTuitionAccountOrderRecordItem
		if err := rows.Scan(
			&item.ID,
			&item.TuitionAccountID,
			&item.Quantity,
			&item.FreeQuantity,
			&item.Status,
			&item.UpdatedStaffID,
			&item.UpdatedStaffName,
			&item.UpdatedTime,
			&item.CreatedTime,
		); err != nil {
			return model.CloseTuitionAccountOrderRecordResult{}, err
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return model.CloseTuitionAccountOrderRecordResult{}, err
	}
	refundAutoCloseRows, err := repo.listRefundAutoCloseTuitionAccountOrdersTx(ctx, tx, instID, accountIDs)
	if err != nil {
		return model.CloseTuitionAccountOrderRecordResult{}, err
	}
	out = append(out, refundAutoCloseRows...)
	sort.SliceStable(out, func(i, j int) bool {
		leftTime := closeOrderRecordSortTime(out[i])
		rightTime := closeOrderRecordSortTime(out[j])
		if !leftTime.Equal(rightTime) {
			return leftTime.After(rightTime)
		}
		return strings.Compare(strings.TrimSpace(out[i].ID), strings.TrimSpace(out[j].ID)) > 0
	})
	if err := tx.Commit(); err != nil {
		return model.CloseTuitionAccountOrderRecordResult{}, err
	}
	return model.CloseTuitionAccountOrderRecordResult{List: out}, nil
}

func (repo *Repository) listRefundAutoCloseTuitionAccountOrdersTx(ctx context.Context, tx *sql.Tx, instID int64, accountIDs []int64) ([]model.CloseTuitionAccountOrderRecordItem, error) {
	if len(accountIDs) == 0 {
		return []model.CloseTuitionAccountOrderRecordItem{}, nil
	}
	args := []any{instID, refundTuitionAccountOrderStatusCompleted, model.RefundTuitionAccountOrderStatusVoided}
	for _, id := range accountIDs {
		args = append(args, id)
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT
			ro.id,
			CAST(MIN(roi.tuition_account_id) AS CHAR),
			SUM(IFNULL(roi.quantity, 0)),
			SUM(IFNULL(roi.free_quantity, 0)),
			CASE
				WHEN IFNULL(ro.status, 0) = ? THEN ?
				WHEN SUM(CASE WHEN IFNULL(ta.status, 0) = ? THEN 1 ELSE 0 END) > 0 THEN ?
				ELSE ?
			END AS close_status,
			CAST(IFNULL(ro.update_id, 0) AS CHAR),
			IFNULL(u.nick_name, ''),
			ro.update_time,
			COALESCE(ro.completed_time, ro.update_time, ro.create_time)
		FROM refund_tuition_account_order ro
		INNER JOIN refund_tuition_account_order_item roi
			ON roi.refund_order_id = ro.id
			AND roi.inst_id = ro.inst_id
			AND roi.del_flag = 0
		INNER JOIN tuition_account ta
			ON ta.id = roi.tuition_account_id
			AND ta.inst_id = ro.inst_id
			AND ta.del_flag = 0
		LEFT JOIN inst_user u ON u.id = ro.update_id AND u.del_flag = 0
		WHERE ro.inst_id = ?
		  AND ro.del_flag = 0
		  AND IFNULL(ro.auto_close_tuition, 0) = 1
		  AND IFNULL(ro.status, 0) IN (?, ?)
		  AND roi.tuition_account_id IN (`+buildPlaceholders(len(accountIDs))+`)
		GROUP BY ro.id, ro.update_id, u.nick_name, ro.update_time, ro.completed_time, ro.create_time
	`, append([]any{
		model.RefundTuitionAccountOrderStatusVoided,
		model.CloseTuitionAccountOrderStatusRevoked,
		model.TuitionAccountStatusClosed,
		model.CloseTuitionAccountOrderStatusClosed,
		model.CloseTuitionAccountOrderStatusRevoked,
	}, args...)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]model.CloseTuitionAccountOrderRecordItem, 0, 4)
	for rows.Next() {
		var (
			item    model.CloseTuitionAccountOrderRecordItem
			rawID   int64
			rowTime sql.NullTime
		)
		if err := rows.Scan(
			&rawID,
			&item.TuitionAccountID,
			&item.Quantity,
			&item.FreeQuantity,
			&item.Status,
			&item.UpdatedStaffID,
			&item.UpdatedStaffName,
			&item.UpdatedTime,
			&rowTime,
		); err != nil {
			return nil, err
		}
		item.ID = buildRefundAutoCloseRevertID(rawID)
		if rowTime.Valid {
			t := rowTime.Time
			item.CreatedTime = &t
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func closeOrderRecordSortTime(item model.CloseTuitionAccountOrderRecordItem) time.Time {
	if item.CreatedTime != nil {
		return *item.CreatedTime
	}
	if item.UpdatedTime != nil {
		return *item.UpdatedTime
	}
	return time.Time{}
}

func int64SliceToAny(list []int64) []any {
	out := make([]any, 0, len(list))
	for _, item := range list {
		out = append(out, item)
	}
	return out
}
