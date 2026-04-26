package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

type refundTuitionAccountDeductionItem struct {
	row                  refundTuitionAccountPreviewRow
	quantity             float64
	freeQuantity         float64
	tuition              float64
	refundAmount         float64
	originalRefundAmount float64
	arrearDeduction      float64
}

type refundTuitionAccountDeductionPlan struct {
	items                     []refundTuitionAccountDeductionItem
	paidQuantity              float64
	freeQuantity              float64
	tuition                   float64
	refundAmount              float64
	totalOriginalRefundAmount float64
	totalArrearDeduction      float64
	handlingFee               float64
	lessonChargingMode        int
}

func ensureRefundTuitionAccountOrderTables(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS refund_tuition_account_order (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			uuid VARCHAR(64) NULL,
			version BIGINT NOT NULL DEFAULT 0,
			inst_id BIGINT NOT NULL,
			tuition_account_id BIGINT NOT NULL DEFAULT 0,
			sale_order_id BIGINT NOT NULL DEFAULT 0,
			order_number VARCHAR(64) NOT NULL DEFAULT '',
			status INT NOT NULL DEFAULT 1,
			total_amount DECIMAL(18,2) NOT NULL DEFAULT 0,
			real_amount DECIMAL(18,2) NOT NULL DEFAULT 0,
			charge_against_tuition DECIMAL(18,2) NOT NULL DEFAULT 0,
			refund_quantity DECIMAL(18,2) NOT NULL DEFAULT 0,
			refund_free_quantity DECIMAL(18,2) NOT NULL DEFAULT 0,
			handling_fee DECIMAL(18,2) NOT NULL DEFAULT 0,
			is_recharge_account TINYINT(1) NOT NULL DEFAULT 0,
			recharge_account_id BIGINT NOT NULL DEFAULT 0,
			deal_date DATE NULL,
			sale_person_id BIGINT NOT NULL DEFAULT 0,
			collector_staff_id BIGINT NOT NULL DEFAULT 0,
			phone_sell_staff_id BIGINT NOT NULL DEFAULT 0,
			foreground_staff_id BIGINT NOT NULL DEFAULT 0,
			vice_sell_staff_staff_id BIGINT NOT NULL DEFAULT 0,
			remark VARCHAR(500) NOT NULL DEFAULT '',
			external_remark VARCHAR(500) NOT NULL DEFAULT '',
			student_id BIGINT NOT NULL DEFAULT 0,
			course_id BIGINT NOT NULL DEFAULT 0,
			auto_close_tuition TINYINT(1) NOT NULL DEFAULT 0,
			completed_time DATETIME NULL,
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT NOT NULL DEFAULT 0,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uk_refund_tuition_sale_order (inst_id, sale_order_id),
			KEY idx_refund_tuition_account_order_account (inst_id, tuition_account_id, del_flag),
			KEY idx_refund_tuition_account_order_student (inst_id, student_id, create_time),
			KEY idx_refund_tuition_account_order_status (inst_id, status, create_time)
		)
	`); err != nil {
		return err
	}
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS refund_tuition_account_order_item (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			inst_id BIGINT NOT NULL,
			refund_order_id BIGINT NOT NULL,
			sale_order_id BIGINT NOT NULL DEFAULT 0,
			tuition_account_id BIGINT NOT NULL DEFAULT 0,
			source_order_id BIGINT NOT NULL DEFAULT 0,
			source_order_number VARCHAR(64) NOT NULL DEFAULT '',
			course_id BIGINT NOT NULL DEFAULT 0,
			lesson_type INT NULL,
			lesson_charging_mode INT NOT NULL DEFAULT 0,
			quantity DECIMAL(18,2) NOT NULL DEFAULT 0,
			free_quantity DECIMAL(18,2) NOT NULL DEFAULT 0,
			tuition DECIMAL(18,2) NOT NULL DEFAULT 0,
			refund_amount DECIMAL(18,2) NOT NULL DEFAULT 0,
			original_refund_amount DECIMAL(18,2) NOT NULL DEFAULT 0,
			arrear_deduction DECIMAL(18,2) NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			KEY idx_refund_tuition_order_item_order (inst_id, refund_order_id),
			KEY idx_refund_tuition_order_item_account (inst_id, tuition_account_id)
		)
	`)
	return err
}

func buildRefundTuitionAccountDeductionPlan(rows []refundTuitionAccountPreviewRow, refundQuantity float64) (refundTuitionAccountDeductionPlan, error) {
	refundableRows := make([]refundTuitionAccountPreviewRow, 0, len(rows))
	var maxRefundQuantity float64
	var lessonChargingMode int
	for _, row := range rows {
		if lessonChargingMode == 0 && row.lessonChargingMode > 0 {
			lessonChargingMode = row.lessonChargingMode
		}
		maxRefundQuantity += row.remainingPaidMetric() + row.remainingGiftMetric()
		if row.remainingPaidMetric() > 0 || row.remainingGiftMetric() > 0 {
			refundableRows = append(refundableRows, row)
		}
	}
	maxRefundQuantity = closeOrderRoundMoney(maxRefundQuantity)
	if len(refundableRows) == 0 || maxRefundQuantity <= 0 {
		return refundTuitionAccountDeductionPlan{}, errors.New("当前无可退课的剩余数量或金额")
	}
	if closeOrderRoundMoney(refundQuantity) > maxRefundQuantity+0.009 {
		return refundTuitionAccountDeductionPlan{}, errors.New("退款数量超过当前可退范围")
	}

	plan := refundTuitionAccountDeductionPlan{
		items:              make([]refundTuitionAccountDeductionItem, 0, len(refundableRows)),
		lessonChargingMode: lessonChargingMode,
	}
	itemIndexByAccount := map[int64]int{}
	addItem := func(next refundTuitionAccountDeductionItem) {
		if idx, ok := itemIndexByAccount[next.row.id]; ok {
			plan.items[idx].quantity = closeOrderRoundMoney(plan.items[idx].quantity + next.quantity)
			plan.items[idx].freeQuantity = closeOrderRoundMoney(plan.items[idx].freeQuantity + next.freeQuantity)
			plan.items[idx].tuition = closeOrderRoundMoney(plan.items[idx].tuition + next.tuition)
			plan.items[idx].refundAmount = closeOrderRoundMoney(plan.items[idx].refundAmount + next.refundAmount)
			plan.items[idx].originalRefundAmount = closeOrderRoundMoney(plan.items[idx].originalRefundAmount + next.originalRefundAmount)
			plan.items[idx].arrearDeduction = closeOrderRoundMoney(plan.items[idx].arrearDeduction + next.arrearDeduction)
			return
		}
		itemIndexByAccount[next.row.id] = len(plan.items)
		plan.items = append(plan.items, next)
	}

	remainingNeed := closeOrderRoundMoney(refundQuantity)
	for _, row := range refundableRows {
		if remainingNeed <= 0.009 {
			break
		}
		paidMetric := row.remainingPaidMetric()
		if paidMetric <= 0 {
			continue
		}
		deduction := paidMetric
		if deduction > remainingNeed {
			deduction = remainingNeed
		}
		deduction = closeOrderRoundMoney(deduction)
		if deduction <= 0 {
			continue
		}

		tuition := row.refundAmountByDeduction(deduction)
		originalRefundAmount := row.originalRefundAmountByDeduction(deduction)
		arrearDeduction := row.arrearTuition
		if originalRefundAmount <= 0 {
			arrearDeduction = 0
		} else if arrearDeduction > originalRefundAmount {
			arrearDeduction = originalRefundAmount
		}
		arrearDeduction = closeOrderRoundMoney(arrearDeduction)
		refundAmount := closeOrderRoundMoney(tuition - arrearDeduction)

		addItem(refundTuitionAccountDeductionItem{
			row:                  row,
			quantity:             deduction,
			tuition:              tuition,
			refundAmount:         refundAmount,
			originalRefundAmount: originalRefundAmount,
			arrearDeduction:      arrearDeduction,
		})
		plan.paidQuantity += deduction
		plan.tuition += tuition
		plan.refundAmount += refundAmount
		plan.totalOriginalRefundAmount += originalRefundAmount
		plan.totalArrearDeduction += arrearDeduction
		remainingNeed = closeOrderRoundMoney(remainingNeed - deduction)
	}

	if remainingNeed > 0.009 {
		for _, row := range refundableRows {
			if remainingNeed <= 0.009 {
				break
			}
			giftMetric := row.remainingGiftMetric()
			if giftMetric <= 0 {
				continue
			}
			deduction := giftMetric
			if deduction > remainingNeed {
				deduction = remainingNeed
			}
			deduction = closeOrderRoundMoney(deduction)
			if deduction <= 0 {
				continue
			}
			addItem(refundTuitionAccountDeductionItem{
				row:          row,
				freeQuantity: deduction,
			})
			plan.freeQuantity += deduction
			remainingNeed = closeOrderRoundMoney(remainingNeed - deduction)
		}
	}
	if remainingNeed > 0.009 {
		return refundTuitionAccountDeductionPlan{}, errors.New("退款数量超过当前可退范围")
	}

	plan.paidQuantity = closeOrderRoundMoney(plan.paidQuantity)
	plan.freeQuantity = closeOrderRoundMoney(plan.freeQuantity)
	plan.tuition = closeOrderRoundMoney(plan.tuition)
	plan.refundAmount = closeOrderRoundMoney(plan.refundAmount)
	plan.totalOriginalRefundAmount = closeOrderRoundMoney(plan.totalOriginalRefundAmount)
	plan.totalArrearDeduction = closeOrderRoundMoney(plan.totalArrearDeduction)
	originalRefundAfterDeduction := closeOrderRoundMoney(plan.totalOriginalRefundAmount - plan.totalArrearDeduction)
	if originalRefundAfterDeduction <= 0 {
		plan.handlingFee = plan.refundAmount
	} else {
		plan.handlingFee = closeOrderRoundMoney(plan.refundAmount - originalRefundAfterDeduction)
		if plan.handlingFee < 0 {
			plan.handlingFee = 0
		}
	}
	return plan, nil
}

func (repo *Repository) CreateRefundTuitionAccountOrder(ctx context.Context, instID, operatorID int64, dto model.RefundTuitionAccountCreateOrderDTO) (int64, bool, error) {
	tuitionAccountID, err := strconv.ParseInt(strings.TrimSpace(dto.TuitionAccountID), 10, 64)
	if err != nil || tuitionAccountID <= 0 {
		return 0, false, errors.New("tuitionAccountId 无效")
	}
	if dto.RefundQuantity <= 0 {
		return 0, false, errors.New("退课数量需大于0")
	}
	if dto.RealAmount < 0 {
		return 0, false, errors.New("实退金额不能小于0")
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, false, err
	}
	defer tx.Rollback()

	accountIDs, err := repo.loadRefundPreviewAccountIDsTx(ctx, tx, instID, tuitionAccountID)
	if err != nil {
		return 0, false, err
	}
	rows, err := repo.listRefundTuitionAccountPreviewRowsTx(ctx, tx, instID, accountIDs)
	if err != nil {
		return 0, false, err
	}
	plan, err := buildRefundTuitionAccountDeductionPlan(rows, dto.RefundQuantity)
	if err != nil {
		return 0, false, err
	}
	if len(plan.items) == 0 {
		return 0, false, errors.New("当前无可退课的剩余数量或金额")
	}
	selected := plan.items[0].row
	if selected.studentID <= 0 || selected.courseID <= 0 {
		return 0, false, errors.New("学费账户信息不完整")
	}

	totalAmount := plan.refundAmount
	realAmount := closeOrderRoundMoney(dto.RealAmount)
	isNeedPay := realAmount > 0.000001
	if realAmount > 999999999 {
		return 0, false, errors.New("实退金额不能超过999999999")
	}
	status := model.OrderStatusCompleted
	if isNeedPay {
		status = model.OrderStatusPendingPayment
	}
	now := time.Now()
	orderNumber := fmt.Sprintf("%s%06d", now.Format("20060102150405"), now.UnixNano()%1000000)
	orderTagIDs := joinRechargeOrderTagIDs(dto.OrderTagIDs)
	var dealDate any
	if parsed := parseDateStart(dto.DealDate); parsed != nil {
		dealDate = parsed.Format("2006-01-02")
	}

	saleOrderResult, err := tx.ExecContext(ctx, `
		INSERT INTO sale_order (
			uuid, version, inst_id, student_id, order_number, sale_person, deal_date, order_discount_type,
			order_discount_amount, order_discount_number, order_real_amount, order_tag_ids, internal_remark,
			external_remark, order_type, order_status, order_source, create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, ?, ?, ?, ?, NULL, ?, 0, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), ?, NOW(), 0
		)
	`,
		instID,
		selected.studentID,
		orderNumber,
		parseInt64String(dto.SalePersonID),
		dealDate,
		plan.handlingFee,
		totalAmount,
		orderTagIDs,
		strings.TrimSpace(dto.Remark),
		strings.TrimSpace(dto.ExternalRemark),
		model.OrderTypeRefundCourse,
		status,
		model.OrderSourceOffline,
		operatorID,
		operatorID,
	)
	if err != nil {
		return 0, false, err
	}
	saleOrderID, err := saleOrderResult.LastInsertId()
	if err != nil {
		return 0, false, err
	}

	quoteID := any(nil)
	if selected.quoteID > 0 {
		quoteID = selected.quoteID
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO sale_order_course_detail (
			uuid, version, order_id, handle_type, course_id, quote_id, count, unit, free_quantity,
			amount, discount_type, discount_number, share_discount, real_quantity, has_valid_date,
			valid_date, end_date, create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, 1, ?, ?, 1, ?, ?, ?, NULL, 0, 0, ?, 0, NULL, NULL, ?, NOW(), ?, NOW(), 0
		)
	`,
		saleOrderID,
		selected.courseID,
		quoteID,
		selected.courseDetailUnit(),
		plan.freeQuantity,
		plan.tuition,
		closeOrderRoundMoney(plan.paidQuantity+plan.freeQuantity),
		operatorID,
		operatorID,
	); err != nil {
		return 0, false, err
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO refund_tuition_account_order (
			uuid, version, inst_id, tuition_account_id, sale_order_id, order_number, status,
			total_amount, real_amount, charge_against_tuition, refund_quantity, refund_free_quantity, handling_fee,
			is_recharge_account, recharge_account_id, deal_date,
			sale_person_id, collector_staff_id, phone_sell_staff_id, foreground_staff_id, vice_sell_staff_staff_id,
			remark, external_remark, student_id, course_id, auto_close_tuition,
			completed_time, create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			NULL, ?, NOW(), ?, NOW(), 0
		)
	`,
		instID,
		tuitionAccountID,
		saleOrderID,
		orderNumber,
		statusToRefundOrderStatus(status),
		totalAmount,
		realAmount,
		closeOrderRoundMoney(plan.totalArrearDeduction),
		plan.paidQuantity,
		plan.freeQuantity,
		plan.handlingFee,
		dto.IsRechargeAccount,
		parseInt64String(dto.RechargeAccountID),
		dealDate,
		parseInt64String(dto.SalePersonID),
		parseInt64String(dto.CollectorStaffID),
		parseInt64String(dto.PhoneSellStaffID),
		parseInt64String(dto.ForegroundStaffID),
		parseInt64String(dto.ViceSellStaffStaffID),
		strings.TrimSpace(dto.Remark),
		strings.TrimSpace(dto.ExternalRemark),
		selected.studentID,
		selected.courseID,
		dto.AutoCloseTuition,
		operatorID,
		operatorID,
	)
	if err != nil {
		return 0, false, err
	}
	refundOrderID, err := result.LastInsertId()
	if err != nil {
		return 0, false, err
	}

	for _, item := range plan.items {
		sourceOrderNumber := strings.TrimSpace(item.row.orderNumber)
		if sourceOrderNumber == "" {
			sourceOrderNumber = orderNumber
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO refund_tuition_account_order_item (
				inst_id, refund_order_id, sale_order_id, tuition_account_id, source_order_id, source_order_number,
				course_id, lesson_type, lesson_charging_mode, quantity, free_quantity, tuition, refund_amount,
				original_refund_amount, arrear_deduction, create_time, del_flag
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), 0)
		`,
			instID,
			refundOrderID,
			saleOrderID,
			item.row.id,
			item.row.orderID,
			sourceOrderNumber,
			item.row.courseID,
			item.row.lessonTypeValue(),
			item.row.lessonChargingMode,
			item.quantity,
			item.freeQuantity,
			item.tuition,
			item.refundAmount,
			item.originalRefundAmount,
			item.arrearDeduction,
		); err != nil {
			return 0, false, err
		}
	}

	if !isNeedPay {
		if err := repo.completeRefundTuitionAccountOrderTx(ctx, tx, instID, operatorID, refundOrderID, saleOrderID); err != nil {
			return 0, false, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, false, err
	}
	return saleOrderID, isNeedPay, nil
}

func statusToRefundOrderStatus(orderStatus int) int {
	if orderStatus == model.OrderStatusCompleted {
		return 2
	}
	return 1
}

func (repo *Repository) PayRefundTuitionAccountOrderBySchoolPal(ctx context.Context, instID, operatorID int64, dto model.RefundTuitionAccountPayOrderDTO) (int64, error) {
	orderID, err := strconv.ParseInt(strings.TrimSpace(dto.OrderID), 10, 64)
	if err != nil || orderID <= 0 {
		return 0, errors.New("orderId不能为空")
	}
	if len(dto.PayAccounts) == 0 {
		return 0, errors.New("请选择退款方式")
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var (
		refundOrderID int64
		status        int
		realAmount    float64
	)
	if err := tx.QueryRowContext(ctx, `
		SELECT id, status, IFNULL(real_amount, 0)
		FROM refund_tuition_account_order
		WHERE sale_order_id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
		FOR UPDATE
	`, orderID, instID).Scan(&refundOrderID, &status, &realAmount); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("退课订单不存在")
		}
		return 0, err
	}
	if status != 1 {
		return 0, errors.New("订单状态异常")
	}

	var totalPayAmount float64
	for _, account := range dto.PayAccounts {
		totalPayAmount += account.Amount
	}
	totalPayAmount = closeOrderRoundMoney(totalPayAmount)
	if math.Abs(totalPayAmount-realAmount) > 0.000001 || math.Abs(closeOrderRoundMoney(dto.PayAmount)-realAmount) > 0.000001 {
		return 0, errors.New("支付金额必须等于实退金额")
	}
	if totalPayAmount <= 0 {
		return 0, errors.New("支付金额不能小于0")
	}

	var firstPaymentDetailID int64
	for _, account := range dto.PayAccounts {
		if account.Amount <= 0 {
			continue
		}
		if account.PayMethod <= 0 {
			return 0, errors.New("请选择退款方式")
		}
		payTime := any(time.Now())
		if parsed := parseDateStart(account.PayTime); parsed != nil {
			payTime = *parsed
		}
		accountID := parseInt64String(account.AccountID)
		if accountID <= 0 {
			accountID = 1
		}
		voucherJSON, err := json.Marshal(model.RefundTuitionAccountPaymentVoucher{
			Text:   strings.TrimSpace(account.PaymentVoucher.Text),
			Images: normalizeLedgerImages(account.PaymentVoucher.Images),
		})
		if err != nil {
			return 0, err
		}
		result, err := tx.ExecContext(ctx, `
			INSERT INTO sale_order_pay_detail (
				uuid, version, inst_id, order_id, amount_id, pay_method, pay_amount, pay_time, payment_voucher,
				create_id, create_time, update_id, update_time, del_flag, remark
			) VALUES (
				UUID(), 0, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), ?, NOW(), 0, ?
			)
		`, instID, orderID, accountID, account.PayMethod, account.Amount, payTime, string(voucherJSON), operatorID, operatorID, strings.TrimSpace(account.PaymentVoucher.Text))
		if err != nil {
			return 0, err
		}
		paymentDetailID, err := result.LastInsertId()
		if err != nil {
			return 0, err
		}
		if firstPaymentDetailID == 0 {
			firstPaymentDetailID = paymentDetailID
		}
		if err := repo.upsertOrderPaymentLedgerTx(ctx, tx, instID, paymentDetailID); err != nil {
			return 0, err
		}
	}
	if firstPaymentDetailID == 0 {
		return 0, errors.New("支付金额不能小于0")
	}

	if err := repo.completeRefundTuitionAccountOrderTx(ctx, tx, instID, operatorID, refundOrderID, orderID); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return firstPaymentDetailID, nil
}

func (repo *Repository) completeRefundTuitionAccountOrderTx(ctx context.Context, tx *sql.Tx, instID, operatorID, refundOrderID, saleOrderID int64) error {
	var (
		studentID        int64
		courseID         int64
		orderNumber      string
		autoCloseTuition bool
	)
	if err := tx.QueryRowContext(ctx, `
		SELECT student_id, course_id, IFNULL(order_number, ''), IFNULL(auto_close_tuition, 0)
		FROM refund_tuition_account_order
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, refundOrderID, instID).Scan(&studentID, &courseID, &orderNumber, &autoCloseTuition); err != nil {
		return err
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT tuition_account_id, source_order_number, course_id, lesson_type, lesson_charging_mode,
		       IFNULL(quantity, 0), IFNULL(free_quantity, 0), IFNULL(tuition, 0)
		FROM refund_tuition_account_order_item
		WHERE refund_order_id = ? AND inst_id = ? AND del_flag = 0
		ORDER BY id ASC
	`, refundOrderID, instID)
	if err != nil {
		return err
	}
	defer rows.Close()

	type refundItem struct {
		tuitionAccountID   int64
		sourceOrderNumber  string
		courseID           int64
		lessonType         sql.NullInt64
		lessonChargingMode int
		quantity           float64
		freeQuantity       float64
		tuition            float64
	}
	items := make([]refundItem, 0, 4)
	for rows.Next() {
		var item refundItem
		if err := rows.Scan(
			&item.tuitionAccountID,
			&item.sourceOrderNumber,
			&item.courseID,
			&item.lessonType,
			&item.lessonChargingMode,
			&item.quantity,
			&item.freeQuantity,
			&item.tuition,
		); err != nil {
			return err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	now := time.Now()
	for _, item := range items {
		var (
			remainingQuantity float64
			remainingTuition  float64
		)
		if err := tx.QueryRowContext(ctx, `
			SELECT IFNULL(remaining_quantity, 0), IFNULL(remaining_tuition, 0)
			FROM tuition_account
			WHERE id = ? AND inst_id = ? AND del_flag = 0
			LIMIT 1
			FOR UPDATE
		`, item.tuitionAccountID, instID).Scan(&remainingQuantity, &remainingTuition); err != nil {
			return err
		}
		deductQuantity := closeOrderRoundMoney(item.quantity + item.freeQuantity)
		if deductQuantity > remainingQuantity+0.02 && item.lessonChargingMode != 3 && item.lessonChargingMode != 4 {
			return errors.New("退课数量超过当前可退范围")
		}
		if item.tuition > remainingTuition+0.02 {
			return errors.New("退课金额超过当前可退范围")
		}
		newRemainingQuantity := closeOrderRoundMoney(math.Max(remainingQuantity-deductQuantity, 0))
		newRemainingTuition := closeOrderRoundMoney(math.Max(remainingTuition-item.tuition, 0))

		statusSQL := ""
		args := []any{newRemainingQuantity, newRemainingTuition, operatorID}
		if autoCloseTuition && newRemainingQuantity <= 0.02 && newRemainingTuition <= 0.02 {
			statusSQL = ", status = ?, status_change_time = ?, class_ending_time = ?"
			args = append(args, model.TuitionAccountStatusClosed, now, now)
		}
		args = append(args, item.tuitionAccountID, instID)
		if _, err := tx.ExecContext(ctx, `
			UPDATE tuition_account
			SET remaining_quantity = ?,
			    remaining_tuition = ?,
			    update_id = ?,
			    update_time = NOW()
			    `+statusSQL+`
			WHERE id = ? AND inst_id = ? AND del_flag = 0
		`, args...); err != nil {
			return err
		}

		lessonTypeValue := any(nil)
		if item.lessonType.Valid {
			lessonTypeValue = item.lessonType.Int64
		}
		flowOrderNumber := strings.TrimSpace(item.sourceOrderNumber)
		if flowOrderNumber == "" {
			flowOrderNumber = orderNumber
		}
		if _, err := tx.ExecContext(ctx, `
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
			item.tuitionAccountID,
			studentID,
			item.courseID,
			lessonTypeValue,
			item.lessonChargingMode,
			model.TuitionAccountFlowSourceRefund,
			saleOrderID,
			flowOrderNumber,
			now,
			deductQuantity,
			item.tuition,
			newRemainingQuantity,
			newRemainingTuition,
			operatorID,
			operatorID,
		); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				return errors.New("请勿重复提交退课")
			}
			return err
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE refund_tuition_account_order
		SET status = 2, completed_time = ?, update_id = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, now, operatorID, refundOrderID, instID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE sale_order
		SET order_status = ?, update_id = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, model.OrderStatusCompleted, operatorID, saleOrderID, instID); err != nil {
		return err
	}

	if autoCloseTuition {
		if err := repo.closeRelatedOneToOneClassesByDeductCourseTx(ctx, tx, instID, operatorID, studentID, courseID); err != nil {
			return err
		}
		if err := repo.closeRelatedGroupClassesByDeductCourseTx(ctx, tx, instID, operatorID, studentID, courseID); err != nil {
			return err
		}
	}
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
		return err
	}
	return nil
}
