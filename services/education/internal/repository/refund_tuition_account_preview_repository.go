package repository

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

type refundTuitionAccountPreviewRow struct {
	id                  int64
	lessonChargingMode  int
	orderID             int64
	orderNumber         string
	totalQuantity       float64
	freeQuantity        float64
	usedQuantity        float64
	remainingQuantity   float64
	remainingTuition    float64
	shouldTuition       float64
	originalShouldPrice float64
	paidQuantityBase    float64
	originalQtyBase     float64
	paidAmount          float64
	arrearTuition       float64
	createTime          sql.NullTime
}

func (row refundTuitionAccountPreviewRow) remainingPaidMetric() float64 {
	if row.lessonChargingMode == 3 || row.lessonChargingMode == 4 {
		return closeOrderRoundMoney(row.remainingTuition)
	}
	if row.totalQuantity > 0 {
		return closeOrderRoundMoney(row.remainingQuantity)
	}
	return 0
}

func (row refundTuitionAccountPreviewRow) remainingGiftMetric() float64 {
	if row.lessonChargingMode == 3 || row.lessonChargingMode == 4 {
		return 0
	}
	if row.totalQuantity <= 0 && row.freeQuantity > 0 {
		return closeOrderRoundMoney(row.remainingQuantity)
	}
	return 0
}

func (row refundTuitionAccountPreviewRow) originalUnitPrice() float64 {
	if row.lessonChargingMode == 3 || row.lessonChargingMode == 4 || row.originalQtyBase <= 0 {
		return 0
	}
	return closeOrderRoundMoney(row.originalShouldPrice / row.originalQtyBase)
}

func (row refundTuitionAccountPreviewRow) discountedUnitPrice() float64 {
	if row.lessonChargingMode == 3 || row.lessonChargingMode == 4 || row.paidQuantityBase <= 0 {
		return 0
	}
	return closeOrderRoundMoney(row.shouldTuition / row.paidQuantityBase)
}

func (row refundTuitionAccountPreviewRow) refundAmountByDeduction(deduction float64) float64 {
	if row.lessonChargingMode == 3 || row.lessonChargingMode == 4 {
		return closeOrderRoundMoney(deduction)
	}
	return closeOrderRoundMoney(deduction * row.discountedUnitPrice())
}

func (repo *Repository) loadRefundPreviewAccountIDsTx(ctx context.Context, tx *sql.Tx, instID, tuitionAccountID int64) ([]int64, error) {
	selected, err := repo.loadCloseTuitionAccountSnapshotTx(ctx, tx, instID, tuitionAccountID)
	if err != nil {
		return nil, err
	}
	bucket, err := repo.loadOneToOneTuitionBucketTx(ctx, tx, instID, tuitionAccountID)
	if err != nil {
		return nil, err
	}
	accountIDs, err := repo.ListTuitionAccountIDsForStudentCourseBucketAllStatuses(
		ctx,
		tx,
		instID,
		selected.studentID,
		bucket.courseID,
		bucket.teachMethod,
		bucket.lessonModelCode,
	)
	if err != nil {
		return nil, err
	}
	if len(accountIDs) == 0 {
		return []int64{selected.id}, nil
	}
	return accountIDs, nil
}

func (repo *Repository) buildRefundTuitionAccountOwedSummaryTx(ctx context.Context, tx *sql.Tx, instID int64, accountIDs []int64) (model.RefundTuitionAccountOwedSummaryResult, error) {
	if len(accountIDs) == 0 {
		return model.RefundTuitionAccountOwedSummaryResult{
			OrderID: "0",
		}, nil
	}

	baseArgs := make([]any, 0, 1+len(accountIDs))
	baseArgs = append(baseArgs, instID)
	for _, accountID := range accountIDs {
		baseArgs = append(baseArgs, accountID)
	}
	baseSQL := `
		FROM (
			SELECT DISTINCT IFNULL(ta.order_id, 0) AS order_id
			FROM tuition_account ta
			WHERE ta.inst_id = ?
			  AND ta.del_flag = 0
			  AND ta.id IN (` + buildPlaceholders(len(accountIDs)) + `)
			  AND IFNULL(ta.order_id, 0) > 0
		) src
		LEFT JOIN sale_order so ON so.id = src.order_id AND so.del_flag = 0
		LEFT JOIN (
			SELECT order_id, SUM(pay_amount) AS paid_amount
			FROM sale_order_pay_detail
			WHERE del_flag = 0
			GROUP BY order_id
		) pay ON pay.order_id = so.id
	`

	var summary model.RefundTuitionAccountOwedSummaryResult
	summary.OrderID = "0"
	if err := tx.QueryRowContext(ctx, `
		SELECT
			IFNULL(SUM(CASE
				WHEN IFNULL(so.is_bad_debt, 0) = 0 AND IFNULL(so.order_real_amount, 0) > IFNULL(pay.paid_amount, 0)
				THEN IFNULL(so.order_real_amount, 0) - IFNULL(pay.paid_amount, 0)
				ELSE 0
			END), 0),
			IFNULL(SUM(CASE
				WHEN IFNULL(so.is_bad_debt, 0) = 1 THEN IFNULL(so.bad_debt_amount, 0)
				ELSE 0
			END), 0)
	`+baseSQL, baseArgs...).Scan(&summary.ArrearAmountTotal, &summary.BadDebtAmountTotal); err != nil {
		return model.RefundTuitionAccountOwedSummaryResult{}, err
	}
	summary.ArrearAmountTotal = closeOrderRoundMoney(summary.ArrearAmountTotal)
	summary.BadDebtAmountTotal = closeOrderRoundMoney(summary.BadDebtAmountTotal)

	row := tx.QueryRowContext(ctx, `
		SELECT
			CAST(so.id AS CHAR),
			IFNULL(so.order_type, 0)
	`+baseSQL+`
		WHERE (
			IFNULL(so.is_bad_debt, 0) = 0 AND IFNULL(so.order_real_amount, 0) > IFNULL(pay.paid_amount, 0)
		) OR IFNULL(so.is_bad_debt, 0) = 1
		ORDER BY so.create_time DESC, so.id DESC
		LIMIT 1
	`, baseArgs...)
	switch err := row.Scan(&summary.OrderID, &summary.OrderType); err {
	case nil:
	case sql.ErrNoRows:
		summary.OrderID = "0"
		summary.OrderType = 0
	default:
		return model.RefundTuitionAccountOwedSummaryResult{}, err
	}
	if strings.TrimSpace(summary.OrderID) == "" {
		summary.OrderID = "0"
	}

	return summary, nil
}

func (repo *Repository) listRefundTuitionAccountPreviewRowsTx(ctx context.Context, tx *sql.Tx, instID int64, accountIDs []int64) ([]refundTuitionAccountPreviewRow, error) {
	if len(accountIDs) == 0 {
		return []refundTuitionAccountPreviewRow{}, nil
	}

	args := make([]any, 0, 1+len(accountIDs))
	args = append(args, instID)
	for _, accountID := range accountIDs {
		args = append(args, accountID)
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT
			ta.id,
			IFNULL(icq.lesson_model, 0),
			IFNULL(ta.order_id, 0),
			IFNULL(so.order_number, ''),
			IFNULL(ta.total_quantity, 0),
			IFNULL(ta.free_quantity, 0),
			IFNULL(ta.used_quantity, 0),
			IFNULL(ta.remaining_quantity, 0),
			IFNULL(ta.remaining_tuition, 0),
			CASE
				WHEN sod.id IS NOT NULL THEN GREATEST(IFNULL(sod.amount, 0) - IFNULL(sod.share_discount, 0), 0)
				ELSE IFNULL(ta.total_tuition, 0)
			END AS should_tuition,
			CASE
				WHEN sod.id IS NOT NULL THEN GREATEST(IFNULL(sod.amount, 0), 0)
				ELSE IFNULL(ta.total_tuition, 0)
			END AS original_should_price,
			CASE
				WHEN IFNULL(sod.real_quantity, 0) > 0 THEN IFNULL(sod.real_quantity, 0)
				WHEN IFNULL(ta.total_quantity, 0) > 0 THEN IFNULL(ta.total_quantity, 0)
				ELSE 0
			END AS paid_quantity_base,
			CASE
				WHEN IFNULL(sod.count, 0) > 0 THEN IFNULL(sod.count, 0)
				WHEN IFNULL(ta.total_quantity, 0) + IFNULL(ta.free_quantity, 0) > 0
				THEN IFNULL(ta.total_quantity, 0) + IFNULL(ta.free_quantity, 0)
				ELSE 0
			END AS original_qty_base,
			CASE
				WHEN IFNULL(so.order_real_amount, 0) <= 0 THEN 0
				ELSE LEAST(
					CASE
						WHEN sod.id IS NOT NULL THEN GREATEST(IFNULL(sod.amount, 0) - IFNULL(sod.share_discount, 0), 0)
						ELSE IFNULL(ta.total_tuition, 0)
					END,
					IFNULL(pay.paid_amount, 0) * (
						CASE
							WHEN sod.id IS NOT NULL THEN GREATEST(IFNULL(sod.amount, 0) - IFNULL(sod.share_discount, 0), 0)
							ELSE IFNULL(ta.total_tuition, 0)
						END
					) / IFNULL(so.order_real_amount, 0)
				)
			END AS paid_amount,
			CASE
				WHEN IFNULL(so.is_bad_debt, 0) = 1 THEN 0
				WHEN IFNULL(so.order_real_amount, 0) <= 0 THEN 0
				ELSE GREATEST(
					(CASE
						WHEN sod.id IS NOT NULL THEN GREATEST(IFNULL(sod.amount, 0) - IFNULL(sod.share_discount, 0), 0)
						ELSE IFNULL(ta.total_tuition, 0)
					END) - LEAST(
						CASE
							WHEN sod.id IS NOT NULL THEN GREATEST(IFNULL(sod.amount, 0) - IFNULL(sod.share_discount, 0), 0)
							ELSE IFNULL(ta.total_tuition, 0)
						END,
						IFNULL(pay.paid_amount, 0) * (
							CASE
								WHEN sod.id IS NOT NULL THEN GREATEST(IFNULL(sod.amount, 0) - IFNULL(sod.share_discount, 0), 0)
								ELSE IFNULL(ta.total_tuition, 0)
							END
						) / IFNULL(so.order_real_amount, 0)
					),
					0
				)
			END AS arrear_tuition,
			ta.create_time
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
		ORDER BY ta.create_time ASC, ta.id ASC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]refundTuitionAccountPreviewRow, 0, len(accountIDs))
	for rows.Next() {
		var item refundTuitionAccountPreviewRow
		if err := rows.Scan(
			&item.id,
			&item.lessonChargingMode,
			&item.orderID,
			&item.orderNumber,
			&item.totalQuantity,
			&item.freeQuantity,
			&item.usedQuantity,
			&item.remainingQuantity,
			&item.remainingTuition,
			&item.shouldTuition,
			&item.originalShouldPrice,
			&item.paidQuantityBase,
			&item.originalQtyBase,
			&item.paidAmount,
			&item.arrearTuition,
			&item.createTime,
		); err != nil {
			return nil, err
		}
		item.shouldTuition = closeOrderRoundMoney(item.shouldTuition)
		item.originalShouldPrice = closeOrderRoundMoney(item.originalShouldPrice)
		item.paidAmount = closeOrderRoundMoney(item.paidAmount)
		item.arrearTuition = closeOrderRoundMoney(item.arrearTuition)
		list = append(list, item)
	}
	return list, rows.Err()
}

func (repo *Repository) GetRefundTuitionAccountOwedSummary(ctx context.Context, instID int64, dto model.RefundTuitionAccountOwedSummaryQueryDTO) (model.RefundTuitionAccountOwedSummaryResult, error) {
	tuitionAccountID, err := strconv.ParseInt(strings.TrimSpace(dto.TuitionAccountID), 10, 64)
	if err != nil || tuitionAccountID <= 0 {
		return model.RefundTuitionAccountOwedSummaryResult{}, errors.New("tuitionAccountId 无效")
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return model.RefundTuitionAccountOwedSummaryResult{}, err
	}
	defer tx.Rollback()

	accountIDs, err := repo.loadRefundPreviewAccountIDsTx(ctx, tx, instID, tuitionAccountID)
	if err != nil {
		return model.RefundTuitionAccountOwedSummaryResult{}, err
	}
	return repo.buildRefundTuitionAccountOwedSummaryTx(ctx, tx, instID, accountIDs)
}

func (repo *Repository) CalculateRefundTuitionAccountHandlingFee(ctx context.Context, instID int64, dto model.RefundTuitionAccountHandlingFeeQueryDTO) (model.RefundTuitionAccountHandlingFeeResult, error) {
	tuitionAccountID, err := strconv.ParseInt(strings.TrimSpace(dto.TuitionAccountID), 10, 64)
	if err != nil || tuitionAccountID <= 0 {
		return model.RefundTuitionAccountHandlingFeeResult{}, errors.New("tuitionAccountId 无效")
	}
	if dto.RefundQuantity <= 0 {
		return model.RefundTuitionAccountHandlingFeeResult{}, errors.New("退款数量需大于 0")
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return model.RefundTuitionAccountHandlingFeeResult{}, err
	}
	defer tx.Rollback()

	accountIDs, err := repo.loadRefundPreviewAccountIDsTx(ctx, tx, instID, tuitionAccountID)
	if err != nil {
		return model.RefundTuitionAccountHandlingFeeResult{}, err
	}
	owedSummary, err := repo.buildRefundTuitionAccountOwedSummaryTx(ctx, tx, instID, accountIDs)
	if err != nil {
		return model.RefundTuitionAccountHandlingFeeResult{}, err
	}
	rows, err := repo.listRefundTuitionAccountPreviewRowsTx(ctx, tx, instID, accountIDs)
	if err != nil {
		return model.RefundTuitionAccountHandlingFeeResult{}, err
	}

	refundableRows := make([]refundTuitionAccountPreviewRow, 0, len(rows))
	var lessonChargingMode int
	var maxRefundQuantity float64
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
		return model.RefundTuitionAccountHandlingFeeResult{}, errors.New("当前无可退课的剩余数量或金额")
	}
	if closeOrderRoundMoney(dto.RefundQuantity) > maxRefundQuantity+0.009 {
		return model.RefundTuitionAccountHandlingFeeResult{}, errors.New("退款数量超过当前可退范围")
	}

	result := model.RefundTuitionAccountHandlingFeeResult{
		TuitionAccountID:   strings.TrimSpace(dto.TuitionAccountID),
		LessonChargingMode: lessonChargingMode,
		ArrearAmountTotal:  owedSummary.ArrearAmountTotal,
		BadDebtAmountTotal: owedSummary.BadDebtAmountTotal,
		OrderID:            owedSummary.OrderID,
		OrderType:          owedSummary.OrderType,
		Details:            make([]model.RefundTuitionAccountHandlingFeeDetail, 0, len(refundableRows)),
	}

	remainingNeed := closeOrderRoundMoney(dto.RefundQuantity)
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

		originalRefundAmount := row.refundAmountByDeduction(deduction)
		arrearDeduction := row.arrearTuition
		if arrearDeduction > originalRefundAmount {
			arrearDeduction = originalRefundAmount
		}
		arrearDeduction = closeOrderRoundMoney(arrearDeduction)
		refundAmount := closeOrderRoundMoney(originalRefundAmount - arrearDeduction)

		result.TotalOriginalRefundAmount += originalRefundAmount
		result.TotalArrearDeduction += arrearDeduction
		result.RefundAmount += refundAmount
		result.PaidRefundQuantity += deduction
		result.Details = append(result.Details, model.RefundTuitionAccountHandlingFeeDetail{
			OrderNumber:          row.orderNumber,
			OriginalUnitPrice:    row.originalUnitPrice(),
			DiscountedUnitPrice:  row.discountedUnitPrice(),
			ShouldTuition:        row.shouldTuition,
			PaidAmount:           row.paidAmount,
			TransferredTuition:   0,
			ConsumedQuantity:     closeOrderRoundMoney(row.usedQuantity),
			ArrearTuition:        row.arrearTuition,
			OriginalRefundAmount: originalRefundAmount,
			LessonChargingMode:   row.lessonChargingMode,
			DeductionQuantity:    deduction,
		})
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
			result.GiftRefundQuantity += deduction
			remainingNeed = closeOrderRoundMoney(remainingNeed - deduction)
		}
	}

	if remainingNeed > 0.009 {
		return model.RefundTuitionAccountHandlingFeeResult{}, errors.New("退款数量超过当前可退范围")
	}

	result.TotalOriginalRefundAmount = closeOrderRoundMoney(result.TotalOriginalRefundAmount)
	result.TotalArrearDeduction = closeOrderRoundMoney(result.TotalArrearDeduction)
	result.RefundAmount = closeOrderRoundMoney(result.RefundAmount)
	originalRefundAfterDeduction := closeOrderRoundMoney(result.TotalOriginalRefundAmount - result.TotalArrearDeduction)
	if originalRefundAfterDeduction <= 0 {
		result.HandlingFee = result.RefundAmount
	} else {
		result.HandlingFee = closeOrderRoundMoney(result.RefundAmount - originalRefundAfterDeduction)
	}
	result.PaidRefundQuantity = closeOrderRoundMoney(result.PaidRefundQuantity)
	result.GiftRefundQuantity = closeOrderRoundMoney(result.GiftRefundQuantity)

	return result, nil
}

func (repo *Repository) EstimateRefundTuitionAccountValuableTuition(ctx context.Context, instID int64, dto model.RefundTuitionAccountValuableEstimateQueryDTO) (model.RefundTuitionAccountValuableEstimateResult, error) {
	tuitionAccountID, err := strconv.ParseInt(strings.TrimSpace(dto.TuitionAccountID), 10, 64)
	if err != nil || tuitionAccountID <= 0 {
		return model.RefundTuitionAccountValuableEstimateResult{}, errors.New("tuitionAccountId 无效")
	}
	if dto.Quantity <= 0 {
		return model.RefundTuitionAccountValuableEstimateResult{}, errors.New("退款数量需大于 0")
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return model.RefundTuitionAccountValuableEstimateResult{}, err
	}
	defer tx.Rollback()

	accountIDs, err := repo.loadRefundPreviewAccountIDsTx(ctx, tx, instID, tuitionAccountID)
	if err != nil {
		return model.RefundTuitionAccountValuableEstimateResult{}, err
	}
	rows, err := repo.listRefundTuitionAccountPreviewRowsTx(ctx, tx, instID, accountIDs)
	if err != nil {
		return model.RefundTuitionAccountValuableEstimateResult{}, err
	}

	refundableRows := make([]refundTuitionAccountPreviewRow, 0, len(rows))
	var maxQuantity float64
	for _, row := range rows {
		maxQuantity += row.remainingPaidMetric() + row.remainingGiftMetric()
		if row.remainingPaidMetric() > 0 || row.remainingGiftMetric() > 0 {
			refundableRows = append(refundableRows, row)
		}
	}
	maxQuantity = closeOrderRoundMoney(maxQuantity)
	if len(refundableRows) == 0 || maxQuantity <= 0 {
		return model.RefundTuitionAccountValuableEstimateResult{}, errors.New("当前无可退课的剩余数量或金额")
	}
	if closeOrderRoundMoney(dto.Quantity) > maxQuantity+0.009 {
		return model.RefundTuitionAccountValuableEstimateResult{}, errors.New("退款数量超过当前可退范围")
	}

	result := model.RefundTuitionAccountValuableEstimateResult{
		TuitionAccountID: strings.TrimSpace(dto.TuitionAccountID),
		SubAccounts:      make([]model.RefundTuitionAccountValuableEstimateSubAccount, 0, len(refundableRows)),
	}

	remainingNeed := closeOrderRoundMoney(dto.Quantity)
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
		result.Quantity += deduction
		result.Tuition += tuition
		result.SubAccounts = append(result.SubAccounts, model.RefundTuitionAccountValuableEstimateSubAccount{
			TuitionAccountID: strconv.FormatInt(row.id, 10),
			OrderNumber:      row.orderNumber,
			Quantity:         deduction,
			FreeQuantity:     0,
			Tuition:          tuition,
		})
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
			result.FreeQuantity += deduction
			result.SubAccounts = append(result.SubAccounts, model.RefundTuitionAccountValuableEstimateSubAccount{
				TuitionAccountID: strconv.FormatInt(row.id, 10),
				OrderNumber:      row.orderNumber,
				Quantity:         0,
				FreeQuantity:     deduction,
				Tuition:          0,
			})
			remainingNeed = closeOrderRoundMoney(remainingNeed - deduction)
		}
	}

	if remainingNeed > 0.009 {
		return model.RefundTuitionAccountValuableEstimateResult{}, errors.New("退款数量超过当前可退范围")
	}

	result.Quantity = closeOrderRoundMoney(result.Quantity)
	result.FreeQuantity = closeOrderRoundMoney(result.FreeQuantity)
	result.Tuition = closeOrderRoundMoney(result.Tuition)
	return result, nil
}
