package repository

import (
	"context"
	"database/sql"
	"math"
	"time"

	"go-migration-platform/services/education/internal/model"
)

type discountedTimeSlotRepairAccount struct {
	ID                   int64
	InstID               int64
	OrderID              int64
	OrderStatus          int
	ExpectedTotalQty     float64
	ExpectedTotalTuition float64
	UsedQty              float64
	RemainingQty         float64
	TotalTuition         float64
	PaidTuition          float64
	UsedTuition          float64
	RemainingTuition     float64
	ConfirmedTuition     float64
}

type discountedTimeSlotRepairFlow struct {
	ID              int64
	SourceType      int
	SourceID        int64
	Quantity        float64
	Tuition         float64
	BalanceQuantity float64
	BalanceTuition  float64
	CreatedTime     time.Time
}

// fixDiscountedTimeSlotTuitionHistory 将早期“时段课 + 课程分摊优惠”仍按优惠前金额入账的学费账户/流水修正为优惠后净额。
func fixDiscountedTimeSlotTuitionHistory(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
		SELECT
			ta.id,
			ta.inst_id,
			ta.order_id,
			IFNULL(so.order_status, 0),
			CASE
				WHEN sod.valid_date IS NOT NULL AND sod.end_date IS NOT NULL THEN
					GREATEST(DATEDIFF(DATE(sod.end_date), DATE(sod.valid_date)) + 1 - IFNULL(sod.free_quantity, 0), 0)
				ELSE
					GREATEST(
						IFNULL(ta.total_quantity, 0),
						IFNULL(ta.used_quantity, 0) + IFNULL(ta.remaining_quantity, 0),
						0
					)
			END AS expected_total_qty,
			GREATEST(ROUND(IFNULL(sod.amount, 0) - IFNULL(sod.share_discount, 0), 2), 0) AS expected_total_tuition,
			IFNULL(ta.used_quantity, 0),
			IFNULL(ta.remaining_quantity, 0),
			IFNULL(ta.total_tuition, 0),
			IFNULL(ta.paid_tuition, 0),
			IFNULL(ta.used_tuition, 0),
			IFNULL(ta.remaining_tuition, 0),
			IFNULL(ta.confirmed_tuition, 0)
		FROM tuition_account ta
		INNER JOIN sale_order_course_detail sod ON sod.id = ta.order_course_detail_id AND sod.del_flag = 0
		INNER JOIN inst_course_quotation icq
			ON icq.id = COALESCE(NULLIF(ta.quote_id, 0), NULLIF(sod.quote_id, 0))
			AND icq.del_flag = 0
			AND IFNULL(icq.lesson_model, 0) = 2
		LEFT JOIN sale_order so ON so.id = ta.order_id AND so.del_flag = 0
		WHERE ta.del_flag = 0
		  AND IFNULL(sod.share_discount, 0) > 0
		ORDER BY ta.id ASC
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	accounts := make([]discountedTimeSlotRepairAccount, 0, 16)
	for rows.Next() {
		var item discountedTimeSlotRepairAccount
		if err := rows.Scan(
			&item.ID,
			&item.InstID,
			&item.OrderID,
			&item.OrderStatus,
			&item.ExpectedTotalQty,
			&item.ExpectedTotalTuition,
			&item.UsedQty,
			&item.RemainingQty,
			&item.TotalTuition,
			&item.PaidTuition,
			&item.UsedTuition,
			&item.RemainingTuition,
			&item.ConfirmedTuition,
		); err != nil {
			return err
		}
		accounts = append(accounts, item)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, account := range accounts {
		if err := repairDiscountedTimeSlotAccountTx(ctx, tx, account); err != nil {
			return err
		}
		if err := repairDiscountedTimeSlotFlowsTx(ctx, tx, account); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func repairDiscountedTimeSlotAccountTx(ctx context.Context, tx *sql.Tx, account discountedTimeSlotRepairAccount) error {
	if account.OrderStatus == model.OrderStatusVoided {
		return nil
	}

	expectedTotalQty := roundMoney(math.Max(account.ExpectedTotalQty, 0))
	expectedTotalTuition := roundMoney(math.Max(account.ExpectedTotalTuition, 0))

	usedQty := account.UsedQty
	if usedQty <= 0 && account.RemainingQty > 0 && expectedTotalQty > 0 {
		usedQty = roundMoney(math.Max(expectedTotalQty-account.RemainingQty, 0))
	}
	usedQty = roundMoney(math.Min(math.Max(usedQty, 0), expectedTotalQty))
	remainingQty := roundMoney(math.Max(expectedTotalQty-usedQty, 0))
	usedTuition := timeSlotAccumulatedTuition(expectedTotalTuition, expectedTotalQty, usedQty)
	remainingTuition := roundMoney(math.Max(expectedTotalTuition-usedTuition, 0))
	confirmedTuition := usedTuition

	if almostEqualFloat(account.TotalTuition, expectedTotalTuition) &&
		almostEqualFloat(account.PaidTuition, expectedTotalTuition) &&
		almostEqualFloat(account.UsedQty, usedQty) &&
		almostEqualFloat(account.RemainingQty, remainingQty) &&
		almostEqualFloat(account.UsedTuition, usedTuition) &&
		almostEqualFloat(account.RemainingTuition, remainingTuition) &&
		almostEqualFloat(account.ConfirmedTuition, confirmedTuition) {
		return nil
	}

	_, err := tx.ExecContext(ctx, `
		UPDATE tuition_account
		SET total_quantity = ?,
		    used_quantity = ?,
		    remaining_quantity = ?,
		    total_tuition = ?,
		    paid_tuition = ?,
		    used_tuition = ?,
		    remaining_tuition = ?,
		    confirmed_tuition = ?,
		    update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`,
		expectedTotalQty,
		usedQty,
		remainingQty,
		expectedTotalTuition,
		expectedTotalTuition,
		usedTuition,
		remainingTuition,
		confirmedTuition,
		account.ID,
		account.InstID,
	)
	return err
}

func repairDiscountedTimeSlotFlowsTx(ctx context.Context, tx *sql.Tx, account discountedTimeSlotRepairAccount) error {
	rows, err := tx.QueryContext(ctx, `
		SELECT
			id,
			source_type,
			source_id,
			IFNULL(quantity, 0),
			IFNULL(tuition, 0),
			IFNULL(balance_quantity, 0),
			IFNULL(balance_tuition, 0),
			created_time
		FROM tuition_account_flow
		WHERE inst_id = ?
		  AND tuition_account_id = ?
		  AND del_flag = 0
		  AND source_type IN (?, ?, ?)
		ORDER BY created_time ASC, id ASC
	`,
		account.InstID,
		account.ID,
		model.TuitionAccountFlowSourceAutoConsume,
		model.TuitionAccountFlowSourceRevokeAutoConsume,
		model.TuitionAccountFlowSourceOrderVoid,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	flows := make([]discountedTimeSlotRepairFlow, 0, 8)
	for rows.Next() {
		var item discountedTimeSlotRepairFlow
		if err := rows.Scan(
			&item.ID,
			&item.SourceType,
			&item.SourceID,
			&item.Quantity,
			&item.Tuition,
			&item.BalanceQuantity,
			&item.BalanceTuition,
			&item.CreatedTime,
		); err != nil {
			return err
		}
		flows = append(flows, item)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(flows) == 0 {
		return nil
	}

	expectedTotalQty := roundMoney(math.Max(account.ExpectedTotalQty, 0))
	expectedTotalTuition := roundMoney(math.Max(account.ExpectedTotalTuition, 0))
	if expectedTotalQty <= 0 || expectedTotalTuition < 0 {
		return nil
	}

	autoConsumedQty := 0.0
	for _, flow := range flows {
		if flow.SourceType != model.TuitionAccountFlowSourceAutoConsume {
			continue
		}
		stepQty := roundMoney(math.Abs(flow.Quantity))
		prevQty := math.Min(autoConsumedQty, expectedTotalQty)
		nextQty := math.Min(autoConsumedQty+stepQty, expectedTotalQty)
		prevTuition := timeSlotAccumulatedTuition(expectedTotalTuition, expectedTotalQty, prevQty)
		nextTuition := timeSlotAccumulatedTuition(expectedTotalTuition, expectedTotalQty, nextQty)
		rowTuition := roundMoney(nextTuition - prevTuition)
		balanceQty := roundMoney(math.Max(expectedTotalQty-nextQty, 0))
		balanceTuition := roundMoney(math.Max(expectedTotalTuition-nextTuition, 0))
		if err := updateDiscountedTimeSlotFlowTx(ctx, tx, flow.ID, flow.Quantity, rowTuition, balanceQty, balanceTuition); err != nil {
			return err
		}
		autoConsumedQty = nextQty
	}

	restoredQty := 0.0
	for _, flow := range flows {
		if flow.SourceType != model.TuitionAccountFlowSourceRevokeAutoConsume {
			continue
		}
		stepQty := roundMoney(math.Abs(flow.Quantity))
		prevRestoredQty := math.Min(restoredQty, autoConsumedQty)
		nextRestoredQty := math.Min(restoredQty+stepQty, autoConsumedQty)
		prevRestoredTuition := timeSlotAccumulatedTuition(expectedTotalTuition, expectedTotalQty, prevRestoredQty)
		nextRestoredTuition := timeSlotAccumulatedTuition(expectedTotalTuition, expectedTotalQty, nextRestoredQty)
		rowTuition := roundMoney(-(nextRestoredTuition - prevRestoredTuition))
		consumedAfterRestore := math.Max(autoConsumedQty-nextRestoredQty, 0)
		consumedAfterRestoreTuition := timeSlotAccumulatedTuition(expectedTotalTuition, expectedTotalQty, consumedAfterRestore)
		balanceQty := roundMoney(math.Max(expectedTotalQty-consumedAfterRestore, 0))
		balanceTuition := roundMoney(math.Max(expectedTotalTuition-consumedAfterRestoreTuition, 0))
		if err := updateDiscountedTimeSlotFlowTx(ctx, tx, flow.ID, flow.Quantity, rowTuition, balanceQty, balanceTuition); err != nil {
			return err
		}
		restoredQty = nextRestoredQty
	}

	consumedBeforeVoidQty := math.Max(autoConsumedQty-restoredQty, 0)
	voidedQty := 0.0
	for _, flow := range flows {
		if flow.SourceType != model.TuitionAccountFlowSourceOrderVoid {
			continue
		}
		stepQty := roundMoney(math.Abs(flow.Quantity))
		prevVoidQty := voidedQty
		nextVoidQty := math.Min(voidedQty+stepQty, math.Max(expectedTotalQty-consumedBeforeVoidQty, 0))
		prevVoidTuition := timeSlotAccumulatedTuition(expectedTotalTuition, expectedTotalQty, consumedBeforeVoidQty+prevVoidQty)
		nextVoidTuition := timeSlotAccumulatedTuition(expectedTotalTuition, expectedTotalQty, consumedBeforeVoidQty+nextVoidQty)
		rowTuition := roundMoney(nextVoidTuition - prevVoidTuition)
		balanceQty := roundMoney(math.Max(expectedTotalQty-consumedBeforeVoidQty-nextVoidQty, 0))
		balanceTuition := roundMoney(math.Max(expectedTotalTuition-nextVoidTuition, 0))
		if err := updateDiscountedTimeSlotFlowTx(ctx, tx, flow.ID, flow.Quantity, rowTuition, balanceQty, balanceTuition); err != nil {
			return err
		}
		voidedQty = nextVoidQty
	}

	return nil
}

func updateDiscountedTimeSlotFlowTx(
	ctx context.Context,
	tx *sql.Tx,
	flowID int64,
	quantity,
	tuition,
	balanceQty,
	balanceTuition float64,
) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE tuition_account_flow
		SET quantity = ?,
		    tuition = ?,
		    balance_quantity = ?,
		    balance_tuition = ?,
		    update_time = NOW()
		WHERE id = ? AND del_flag = 0
	`, roundMoney(quantity), roundMoney(tuition), roundMoney(balanceQty), roundMoney(balanceTuition), flowID)
	return err
}

func timeSlotAccumulatedTuition(totalTuition, totalQty, usedQty float64) float64 {
	if totalQty <= 0 || usedQty <= 0 {
		return 0
	}
	if usedQty >= totalQty {
		return roundMoney(totalTuition)
	}
	return roundMoney(totalTuition * usedQty / totalQty)
}
