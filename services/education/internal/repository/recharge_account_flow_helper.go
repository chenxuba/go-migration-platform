package repository

import (
	"context"
	"database/sql"
	"errors"

	"go-migration-platform/services/education/internal/model"
)

func (repo *Repository) lockRechargeAccountBalanceTx(ctx context.Context, tx *sql.Tx, instID, rechargeAccountID int64) (float64, float64, float64, error) {
	var (
		rechargeBalance float64
		residualBalance float64
		givingBalance   float64
	)
	if err := tx.QueryRowContext(ctx, `
		SELECT IFNULL(recharge_balance, 0), IFNULL(residual_balance, 0), IFNULL(giving_balance, 0)
		FROM recharge_account
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
		FOR UPDATE
	`, rechargeAccountID, instID).Scan(&rechargeBalance, &residualBalance, &givingBalance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, 0, errors.New("储值账户不存在")
		}
		return 0, 0, 0, err
	}
	return rechargeBalance, residualBalance, givingBalance, nil
}

func (repo *Repository) increaseRechargeAccountBalanceTx(ctx context.Context, tx *sql.Tx, instID, operatorID, rechargeAccountID int64, amount, residualAmount, givingAmount float64) error {
	if rechargeAccountID <= 0 {
		return nil
	}
	amount = closeOrderRoundMoney(amount)
	residualAmount = closeOrderRoundMoney(residualAmount)
	givingAmount = closeOrderRoundMoney(givingAmount)
	if amount == 0 && residualAmount == 0 && givingAmount == 0 {
		return nil
	}
	if _, _, _, err := repo.lockRechargeAccountBalanceTx(ctx, tx, instID, rechargeAccountID); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `
		UPDATE recharge_account
		SET recharge_balance = IFNULL(recharge_balance, 0) + ?,
			residual_balance = IFNULL(residual_balance, 0) + ?,
			giving_balance = IFNULL(giving_balance, 0) + ?,
			update_id = ?,
			update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, amount, residualAmount, givingAmount, operatorID, rechargeAccountID, instID)
	return err
}

func (repo *Repository) decreaseRechargeAccountBalanceTx(ctx context.Context, tx *sql.Tx, instID, operatorID, rechargeAccountID int64, amount, residualAmount, givingAmount float64) error {
	if rechargeAccountID <= 0 {
		return nil
	}
	amount = closeOrderRoundMoney(amount)
	residualAmount = closeOrderRoundMoney(residualAmount)
	givingAmount = closeOrderRoundMoney(givingAmount)
	if amount == 0 && residualAmount == 0 && givingAmount == 0 {
		return nil
	}
	rechargeBalance, residualBalance, givingBalance, err := repo.lockRechargeAccountBalanceTx(ctx, tx, instID, rechargeAccountID)
	if err != nil {
		return err
	}
	if amount > rechargeBalance+1e-9 {
		return errors.New("储值账户充值余额不足，无法作废")
	}
	if residualAmount > residualBalance+1e-9 {
		return errors.New("储值账户残联余额不足，无法作废")
	}
	if givingAmount > givingBalance+1e-9 {
		return errors.New("储值账户赠送余额不足，无法作废")
	}
	_, err = tx.ExecContext(ctx, `
		UPDATE recharge_account
		SET recharge_balance = IFNULL(recharge_balance, 0) - ?,
			residual_balance = IFNULL(residual_balance, 0) - ?,
			giving_balance = IFNULL(giving_balance, 0) - ?,
			update_id = ?,
			update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, amount, residualAmount, givingAmount, operatorID, rechargeAccountID, instID)
	return err
}

func (repo *Repository) insertRechargeAccountFlowTx(ctx context.Context, tx *sql.Tx, instID, operatorID, rechargeAccountID, studentID int64, orderNumber string, flowType int, amount, residualAmount, givingAmount float64, remark string) error {
	amount = closeOrderRoundMoney(amount)
	residualAmount = closeOrderRoundMoney(residualAmount)
	givingAmount = closeOrderRoundMoney(givingAmount)
	if amount == 0 && residualAmount == 0 && givingAmount == 0 {
		return nil
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO recharge_account_flow (
			inst_id, recharge_account_id, student_id, order_number, flow_type,
			amount, residual_amount, giving_amount, remark,
			create_id, create_time, update_id, update_time, del_flag
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), ?, NOW(), 0)
	`, instID, rechargeAccountID, studentID, orderNumber, flowType, amount, residualAmount, givingAmount, remark, operatorID, operatorID)
	return err
}

func rechargeAccountFlowTypeForOrderObsoleteReturn() int {
	return model.RechargeAccountFlowTypeVoidOrderExpend
}
