package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

func ensureLedgerTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS inst_ledger (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			uuid VARCHAR(64) NULL,
			version BIGINT NOT NULL DEFAULT 0,
			inst_id BIGINT NOT NULL,
			source_type INT NOT NULL DEFAULT 1,
			system_type INT NOT NULL DEFAULT 0,
			source_biz_type INT NOT NULL DEFAULT 0,
			source_biz_id BIGINT NOT NULL DEFAULT 0,
			type INT NOT NULL DEFAULT 1,
			ledger_number VARCHAR(64) NOT NULL,
			ledger_category_id VARCHAR(64) NOT NULL,
			ledger_category_name VARCHAR(100) NOT NULL DEFAULT '',
			ledger_sub_category_id VARCHAR(64) NOT NULL,
			ledger_sub_category_name VARCHAR(100) NOT NULL DEFAULT '',
			ledger_category_icon VARCHAR(100) NOT NULL DEFAULT '',
			amount DECIMAL(18,2) NOT NULL DEFAULT 0,
			deal_staff_id BIGINT NOT NULL DEFAULT 0,
			deal_staff_name VARCHAR(100) NOT NULL DEFAULT '',
			pay_time DATETIME NULL,
			pay_method INT NULL,
			account_id BIGINT NOT NULL DEFAULT 0,
			account_name VARCHAR(100) NOT NULL DEFAULT '',
			reciprocal_account VARCHAR(200) NOT NULL DEFAULT '',
			bank_slip_no VARCHAR(100) NOT NULL DEFAULT '',
			order_id BIGINT NOT NULL DEFAULT 0,
			order_number VARCHAR(64) NOT NULL DEFAULT '',
			student_id BIGINT NOT NULL DEFAULT 0,
			student_name VARCHAR(100) NOT NULL DEFAULT '',
			student_phone VARCHAR(32) NOT NULL DEFAULT '',
			student_phone_raw VARCHAR(32) NOT NULL DEFAULT '',
			payment_voucher_text VARCHAR(1000) NOT NULL DEFAULT '',
			payment_voucher_images JSON NULL,
			ledger_confirm_status INT NOT NULL DEFAULT 0,
			confirm_staff_id BIGINT NOT NULL DEFAULT 0,
			confirm_staff_name VARCHAR(100) NOT NULL DEFAULT '',
			confirm_time DATETIME NULL,
			confirm_remark_text VARCHAR(1000) NOT NULL DEFAULT '',
			confirm_remark_images JSON NULL,
			bill_flow_id BIGINT NOT NULL DEFAULT 0,
			bill_id BIGINT NOT NULL DEFAULT 0,
			error_message VARCHAR(500) NOT NULL DEFAULT '',
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT NOT NULL DEFAULT 0,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uk_inst_ledger_source (inst_id, source_type, source_biz_type, source_biz_id),
			UNIQUE KEY uk_inst_ledger_number (inst_id, ledger_number),
			KEY idx_inst_ledger_list (inst_id, create_time, id),
			KEY idx_inst_ledger_order (inst_id, order_id),
			KEY idx_inst_ledger_student (inst_id, student_id),
			KEY idx_inst_ledger_confirm (inst_id, ledger_confirm_status),
			KEY idx_inst_ledger_sub_category (inst_id, ledger_sub_category_id)
		)
	`)
	return err
}

func (repo *Repository) ensureSystemLedgerRecords(ctx context.Context, instID int64) error {
	_, err := repo.db.ExecContext(ctx, `
		INSERT INTO inst_ledger (
			uuid, version, inst_id, source_type, system_type, source_biz_type, source_biz_id,
			type, ledger_number, ledger_category_id, ledger_category_name, ledger_sub_category_id,
			ledger_sub_category_name, ledger_category_icon, amount, deal_staff_id, deal_staff_name,
			pay_time, pay_method, account_id, account_name, reciprocal_account, bank_slip_no,
			order_id, order_number, student_id, student_name, student_phone, student_phone_raw,
			payment_voucher_text, payment_voucher_images, ledger_confirm_status, confirm_staff_id,
			confirm_staff_name, confirm_time, confirm_remark_text, confirm_remark_images,
			bill_flow_id, bill_id, error_message, create_id, create_time, update_id, update_time, del_flag
		)
		SELECT
			UUID(), 0, pd.inst_id, ?, ?, ?, pd.id,
			CASE WHEN IFNULL(pd.pay_amount, 0) >= 0 THEN ? ELSE ? END,
			CONCAT(DATE_FORMAT(pd.create_time, '%Y%m%d%H%i%s'), LPAD(MOD(pd.id, 1000000), 6, '0')),
			?, ?, 
			CASE
				WHEN IFNULL(so.order_type, 1) = ? THEN ?
				WHEN IFNULL(so.order_type, 1) = ? THEN ?
				WHEN IFNULL(so.order_type, 1) = ? THEN ?
				ELSE ?
			END,
			CASE
				WHEN IFNULL(so.order_type, 1) = ? THEN '储值账户充值'
				WHEN IFNULL(so.order_type, 1) = ? THEN '退课'
				WHEN IFNULL(so.order_type, 1) = ? THEN '转课'
				ELSE '报名续费'
			END,
			'systemTallyBookType1',
			ABS(IFNULL(pd.pay_amount, 0)),
			IFNULL(pd.create_id, 0),
			IFNULL(operator.nick_name, ''),
			pd.pay_time,
			pd.pay_method,
			IFNULL(pd.amount_id, 0),
			CASE
				WHEN IFNULL(pd.amount_id, 0) = 0 THEN '默认账户'
				ELSE CONCAT('账户', pd.amount_id)
			END,
			'',
			'',
			IFNULL(so.id, 0),
			IFNULL(so.order_number, ''),
			IFNULL(so.student_id, 0),
			IFNULL(stu.stu_name, ''),
			CASE
				WHEN CHAR_LENGTH(IFNULL(stu.mobile, '')) >= 7 THEN CONCAT(LEFT(stu.mobile, 3), '****', RIGHT(stu.mobile, 4))
				ELSE IFNULL(stu.mobile, '')
			END,
			IFNULL(stu.mobile, ''),
			IFNULL(pd.payment_voucher, ''),
			JSON_ARRAY(),
			?,
			0,
			'',
			NULL,
			'',
			JSON_ARRAY(),
			pd.id,
			IFNULL(so.id, 0),
			'',
			IFNULL(pd.create_id, 0),
			IFNULL(pd.create_time, NOW()),
			IFNULL(pd.create_id, 0),
			IFNULL(pd.create_time, NOW()),
			0
		FROM sale_order_pay_detail pd
		LEFT JOIN sale_order so ON so.id = pd.order_id AND so.del_flag = 0
		LEFT JOIN inst_student stu ON stu.id = so.student_id AND stu.del_flag = 0
		LEFT JOIN inst_user operator ON operator.id = pd.create_id
		LEFT JOIN inst_ledger l ON l.inst_id = pd.inst_id
			AND l.source_type = ?
			AND l.source_biz_type = ?
			AND l.source_biz_id = pd.id
			AND l.del_flag = 0
		WHERE pd.inst_id = ? AND pd.del_flag = 0 AND l.id IS NULL
	`,
		model.LedgerSourceSystem,
		model.LedgerSystemTypeOrderPayment,
		1,
		model.LedgerTypeIncome,
		model.LedgerTypeExpenditure,
		model.LedgerCategoryOrderIncome,
		"订单收入",
		model.OrderTypeRechargeAccount,
		model.LedgerSubCategoryRechargeAccount,
		model.OrderTypeRefundCourse,
		model.LedgerSubCategoryRefundCourse,
		model.OrderTypeTransferCourse,
		model.LedgerSubCategoryTransferOrder,
		model.LedgerSubCategoryRegistration,
		model.OrderTypeRechargeAccount,
		model.OrderTypeRefundCourse,
		model.OrderTypeTransferCourse,
		model.LedgerConfirmStatusPending,
		model.LedgerSourceSystem,
		1,
		instID,
	)
	return err
}

func (repo *Repository) upsertOrderPaymentLedgerTx(ctx context.Context, tx *sql.Tx, instID, paymentDetailID int64) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO inst_ledger (
			uuid, version, inst_id, source_type, system_type, source_biz_type, source_biz_id,
			type, ledger_number, ledger_category_id, ledger_category_name, ledger_sub_category_id,
			ledger_sub_category_name, ledger_category_icon, amount, deal_staff_id, deal_staff_name,
			pay_time, pay_method, account_id, account_name, reciprocal_account, bank_slip_no,
			order_id, order_number, student_id, student_name, student_phone, student_phone_raw,
			payment_voucher_text, payment_voucher_images, ledger_confirm_status, confirm_staff_id,
			confirm_staff_name, confirm_time, confirm_remark_text, confirm_remark_images,
			bill_flow_id, bill_id, error_message, create_id, create_time, update_id, update_time, del_flag
		)
		SELECT
			UUID(), 0, pd.inst_id, ?, ?, ?, pd.id,
			CASE WHEN IFNULL(pd.pay_amount, 0) >= 0 THEN ? ELSE ? END,
			CONCAT(DATE_FORMAT(pd.create_time, '%Y%m%d%H%i%s'), LPAD(MOD(pd.id, 1000000), 6, '0')),
			?, ?, 
			CASE
				WHEN IFNULL(so.order_type, 1) = ? THEN ?
				WHEN IFNULL(so.order_type, 1) = ? THEN ?
				WHEN IFNULL(so.order_type, 1) = ? THEN ?
				ELSE ?
			END,
			CASE
				WHEN IFNULL(so.order_type, 1) = ? THEN '储值账户充值'
				WHEN IFNULL(so.order_type, 1) = ? THEN '退课'
				WHEN IFNULL(so.order_type, 1) = ? THEN '转课'
				ELSE '报名续费'
			END,
			'systemTallyBookType1',
			ABS(IFNULL(pd.pay_amount, 0)),
			IFNULL(pd.create_id, 0),
			IFNULL(operator.nick_name, ''),
			pd.pay_time,
			pd.pay_method,
			IFNULL(pd.amount_id, 0),
			CASE
				WHEN IFNULL(pd.amount_id, 0) = 0 THEN '默认账户'
				ELSE CONCAT('账户', pd.amount_id)
			END,
			'',
			'',
			IFNULL(so.id, 0),
			IFNULL(so.order_number, ''),
			IFNULL(so.student_id, 0),
			IFNULL(stu.stu_name, ''),
			CASE
				WHEN CHAR_LENGTH(IFNULL(stu.mobile, '')) >= 7 THEN CONCAT(LEFT(stu.mobile, 3), '****', RIGHT(stu.mobile, 4))
				ELSE IFNULL(stu.mobile, '')
			END,
			IFNULL(stu.mobile, ''),
			IFNULL(pd.payment_voucher, ''),
			JSON_ARRAY(),
			?,
			0,
			'',
			NULL,
			'',
			JSON_ARRAY(),
			pd.id,
			IFNULL(so.id, 0),
			'',
			IFNULL(pd.create_id, 0),
			IFNULL(pd.create_time, NOW()),
			IFNULL(pd.create_id, 0),
			IFNULL(pd.create_time, NOW()),
			0
		FROM sale_order_pay_detail pd
		LEFT JOIN sale_order so ON so.id = pd.order_id AND so.del_flag = 0
		LEFT JOIN inst_student stu ON stu.id = so.student_id AND stu.del_flag = 0
		LEFT JOIN inst_user operator ON operator.id = pd.create_id
		WHERE pd.id = ? AND pd.inst_id = ?
		ON DUPLICATE KEY UPDATE
			amount = VALUES(amount),
			deal_staff_id = VALUES(deal_staff_id),
			deal_staff_name = VALUES(deal_staff_name),
			pay_time = VALUES(pay_time),
			pay_method = VALUES(pay_method),
			account_id = VALUES(account_id),
			account_name = VALUES(account_name),
			order_id = VALUES(order_id),
			order_number = VALUES(order_number),
			student_id = VALUES(student_id),
			student_name = VALUES(student_name),
			student_phone = VALUES(student_phone),
			student_phone_raw = VALUES(student_phone_raw),
			payment_voucher_text = VALUES(payment_voucher_text),
			bill_flow_id = VALUES(bill_flow_id),
			bill_id = VALUES(bill_id),
			update_time = NOW()
	`,
		model.LedgerSourceSystem,
		model.LedgerSystemTypeOrderPayment,
		1,
		model.LedgerTypeIncome,
		model.LedgerTypeExpenditure,
		model.LedgerCategoryOrderIncome,
		"订单收入",
		model.OrderTypeRechargeAccount,
		model.LedgerSubCategoryRechargeAccount,
		model.OrderTypeRefundCourse,
		model.LedgerSubCategoryRefundCourse,
		model.OrderTypeTransferCourse,
		model.LedgerSubCategoryTransferOrder,
		model.LedgerSubCategoryRegistration,
		model.OrderTypeRechargeAccount,
		model.OrderTypeRefundCourse,
		model.OrderTypeTransferCourse,
		model.LedgerConfirmStatusPending,
		paymentDetailID,
		instID,
	)
	return err
}

func (repo *Repository) PageLedgers(ctx context.Context, instID int64, query model.LedgerListQueryDTO) (model.LedgerListResultVO, error) {
	if err := repo.ensureSystemLedgerRecords(ctx, instID); err != nil {
		return model.LedgerListResultVO{}, err
	}

	current := query.PageRequestModel.PageIndex
	size := query.PageRequestModel.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	whereClause, args := buildLedgerWhereClause(instID, query.QueryModel)

	var total int
	if err := repo.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM inst_ledger l WHERE `+whereClause, args...).Scan(&total); err != nil {
		return model.LedgerListResultVO{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			l.id, l.type, l.source_type, l.ledger_category_id, l.ledger_category_name,
			l.ledger_sub_category_id, l.ledger_sub_category_name, l.ledger_category_icon,
			IFNULL(l.amount, 0), l.deal_staff_id, IFNULL(l.deal_staff_name, ''),
			l.pay_time, l.create_time, l.pay_method, l.account_id, IFNULL(l.account_name, ''),
			IFNULL(l.reciprocal_account, ''), IFNULL(l.bank_slip_no, ''), IFNULL(l.order_number, ''),
			IFNULL(l.ledger_number, ''), l.student_id, IFNULL(l.student_name, ''), IFNULL(l.student_phone, ''),
			l.ledger_confirm_status, l.confirm_staff_id, IFNULL(l.confirm_staff_name, ''), l.confirm_time,
			IFNULL(l.confirm_remark_text, ''), IFNULL(l.confirm_remark_images, JSON_ARRAY()),
			l.system_type, l.order_id, IFNULL(l.payment_voucher_text, ''), IFNULL(l.payment_voucher_images, JSON_ARRAY()),
			l.bill_flow_id, l.bill_id, IFNULL(l.error_message, '')
		FROM inst_ledger l
		WHERE `+whereClause+`
		ORDER BY l.create_time DESC, l.id DESC
		LIMIT ? OFFSET ?
	`, append(args, size, offset)...)
	if err != nil {
		return model.LedgerListResultVO{}, err
	}
	defer rows.Close()

	orderProducts := map[int64][]string{}
	items := make([]model.LedgerListItemVO, 0, size)
	for rows.Next() {
		var (
			item                model.LedgerListItemVO
			id                  int64
			dealStaffID         sql.NullInt64
			payTime             sql.NullTime
			createdTime         sql.NullTime
			payMethod           sql.NullInt64
			accountID           sql.NullInt64
			studentID           sql.NullInt64
			confirmStaffID      sql.NullInt64
			confirmTime         sql.NullTime
			orderID             sql.NullInt64
			systemType          sql.NullInt64
			billFlowID          sql.NullInt64
			billID              sql.NullInt64
			confirmRemarkImages string
			paymentImages       string
		)
		if err := rows.Scan(
			&id, &item.Type, &item.SourceType, &item.LedgerCategoryID, &item.LedgerCategoryName,
			&item.LedgerSubCategoryID, &item.LedgerSubCategoryName, &item.LedgerCategoryIcon,
			&item.Amount, &dealStaffID, &item.DealStaffName, &payTime, &createdTime, &payMethod,
			&accountID, &item.AccountName, &item.ReciprocalAccount, &item.BankSlipNo, &item.OrderNumber,
			&item.LedgerNumber, &studentID, &item.StudentName, &item.StudentPhone, &item.LedgerConfirmStatus,
			&confirmStaffID, &item.ConfirmStaffName, &confirmTime, &item.ConfirmRemark.Text,
			&confirmRemarkImages, &systemType, &orderID, &item.PaymentVoucher.Text, &paymentImages,
			&billFlowID, &billID, &item.ErrorMessage,
		); err != nil {
			return model.LedgerListResultVO{}, err
		}
		item.ID = strconv.FormatInt(id, 10)
		item.IsConfirmed = item.LedgerConfirmStatus == model.LedgerConfirmStatusConfirmed
		item.ConfirmRemark.Images = parseJSONStringArray(confirmRemarkImages)
		item.PaymentVoucher.Images = parseJSONStringArray(paymentImages)
		if dealStaffID.Valid && dealStaffID.Int64 > 0 {
			item.DealStaffID = strconv.FormatInt(dealStaffID.Int64, 10)
		}
		if payTime.Valid {
			t := payTime.Time
			item.PayTime = &t
		}
		if createdTime.Valid {
			t := createdTime.Time
			item.CreatedTime = &t
		}
		if payMethod.Valid {
			value := int(payMethod.Int64)
			item.PayMethod = &value
		}
		if accountID.Valid && accountID.Int64 > 0 {
			item.AccountID = strconv.FormatInt(accountID.Int64, 10)
		}
		if studentID.Valid && studentID.Int64 > 0 {
			item.StudentID = strconv.FormatInt(studentID.Int64, 10)
		}
		if confirmStaffID.Valid && confirmStaffID.Int64 > 0 {
			_ = strconv.FormatInt(confirmStaffID.Int64, 10)
		}
		if confirmTime.Valid {
			t := confirmTime.Time
			item.ConfirmTime = &t
		}
		if systemType.Valid {
			item.SystemType = int(systemType.Int64)
		}
		if orderID.Valid && orderID.Int64 > 0 {
			item.OrderID = strconv.FormatInt(orderID.Int64, 10)
			products, ok := orderProducts[orderID.Int64]
			if !ok {
				products, _ = repo.getOrderCourseNames(ctx, orderID.Int64)
				orderProducts[orderID.Int64] = products
			}
			item.ProductItems = products
		}
		if billFlowID.Valid && billFlowID.Int64 > 0 {
			item.BillFlowID = strconv.FormatInt(billFlowID.Int64, 10)
		}
		if billID.Valid && billID.Int64 > 0 {
			item.BillID = strconv.FormatInt(billID.Int64, 10)
		}
		items = append(items, item)
	}
	return model.LedgerListResultVO{
		List:  items,
		Total: total,
	}, rows.Err()
}

func (repo *Repository) GetLedgerStatistics(ctx context.Context, instID int64, query model.LedgerListQueryDTO) (model.LedgerStatisticsVO, error) {
	if err := repo.ensureSystemLedgerRecords(ctx, instID); err != nil {
		return model.LedgerStatisticsVO{}, err
	}

	whereClause, args := buildLedgerWhereClause(instID, query.QueryModel)
	var result model.LedgerStatisticsVO
	err := repo.db.QueryRowContext(ctx, `
		SELECT
			IFNULL(SUM(CASE WHEN l.type = ? THEN l.amount ELSE 0 END), 0),
			IFNULL(SUM(CASE WHEN l.type = ? THEN l.amount ELSE 0 END), 0),
			IFNULL(SUM(CASE WHEN l.ledger_confirm_status = ? THEN 1 ELSE 0 END), 0),
			IFNULL(SUM(CASE WHEN l.ledger_confirm_status = ? THEN 1 ELSE 0 END), 0),
			IFNULL(SUM(CASE WHEN l.ledger_confirm_status = ? THEN 1 ELSE 0 END), 0),
			IFNULL(SUM(CASE WHEN l.ledger_confirm_status = ? THEN 1 ELSE 0 END), 0)
		FROM inst_ledger l
		WHERE `+whereClause,
		append([]any{
			model.LedgerTypeIncome,
			model.LedgerTypeExpenditure,
			model.LedgerConfirmStatusConfirmed,
			model.LedgerConfirmStatusPending,
			model.LedgerConfirmStatusRefunding,
			model.LedgerConfirmStatusRefundFailed,
		}, args...)...,
	).Scan(
		&result.IncomeAmount,
		&result.ExpenditureAmount,
		&result.TotalConfirm,
		&result.TotalUnConfirm,
		&result.TotalRefunding,
		&result.TotalRefundFailed,
	)
	if err != nil {
		return model.LedgerStatisticsVO{}, err
	}
	result.BalanceAmount = result.IncomeAmount - result.ExpenditureAmount
	return result, nil
}

func (repo *Repository) ConfirmLedger(ctx context.Context, instID, ledgerID, confirmStaffID int64, confirmStaffName string) error {
	return repo.ConfirmLedgerWithRemark(ctx, instID, ledgerID, confirmStaffID, confirmStaffName, model.LedgerRichText{})
}

func (repo *Repository) ConfirmLedgerWithRemark(ctx context.Context, instID, ledgerID, confirmStaffID int64, confirmStaffName string, remark model.LedgerRichText) error {
	var status int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT ledger_confirm_status
		FROM inst_ledger
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, ledgerID, instID).Scan(&status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("账单不存在")
		}
		return err
	}
	if status == model.LedgerConfirmStatusConfirmed {
		return nil
	}
	if status != model.LedgerConfirmStatusPending {
		return errors.New("当前账单状态不支持确认到账")
	}
	imagesJSON, err := json.Marshal(normalizeLedgerImages(remark.Images))
	if err != nil {
		return err
	}
	_, err = repo.db.ExecContext(ctx, `
		UPDATE inst_ledger
		SET ledger_confirm_status = ?,
			confirm_staff_id = ?,
			confirm_staff_name = ?,
			confirm_time = NOW(),
			confirm_remark_text = ?,
			confirm_remark_images = ?,
			update_id = ?,
			update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, model.LedgerConfirmStatusConfirmed, confirmStaffID, strings.TrimSpace(confirmStaffName), strings.TrimSpace(remark.Text), string(imagesJSON), confirmStaffID, ledgerID, instID)
	return err
}

func (repo *Repository) CancelConfirmLedger(ctx context.Context, instID, ledgerID int64) error {
	var status int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT ledger_confirm_status
		FROM inst_ledger
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, ledgerID, instID).Scan(&status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("账单不存在")
		}
		return err
	}
	if status == model.LedgerConfirmStatusPending {
		return nil
	}
	if status != model.LedgerConfirmStatusConfirmed {
		return errors.New("当前账单状态不支持取消确认")
	}
	_, err := repo.db.ExecContext(ctx, `
		UPDATE inst_ledger
		SET ledger_confirm_status = ?,
			confirm_staff_id = 0,
			confirm_staff_name = '',
			confirm_time = NULL,
			confirm_remark_text = '',
			confirm_remark_images = JSON_ARRAY(),
			update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, model.LedgerConfirmStatusPending, ledgerID, instID)
	return err
}

func buildLedgerWhereClause(instID int64, query model.LedgerQueryFilter) (string, []any) {
	filters := []string{"l.inst_id = ?", "l.del_flag = 0"}
	args := []any{instID}

	if len(query.AccountIDs) > 0 {
		holders := make([]string, 0, len(query.AccountIDs))
		for _, item := range query.AccountIDs {
			if strings.TrimSpace(item) == "" {
				continue
			}
			holders = append(holders, "?")
			args = append(args, strings.TrimSpace(item))
		}
		if len(holders) > 0 {
			filters = append(filters, "CAST(l.account_id AS CHAR) IN ("+strings.Join(holders, ",")+")")
		}
	}
	if len(query.LedgerConfirmStatuses) > 0 {
		holders := make([]string, 0, len(query.LedgerConfirmStatuses))
		for _, item := range query.LedgerConfirmStatuses {
			holders = append(holders, "?")
			args = append(args, item)
		}
		filters = append(filters, "l.ledger_confirm_status IN ("+strings.Join(holders, ",")+")")
	}
	if len(query.SourceTypes) > 0 {
		holders := make([]string, 0, len(query.SourceTypes))
		for _, item := range query.SourceTypes {
			holders = append(holders, "?")
			args = append(args, item)
		}
		filters = append(filters, "l.source_type IN ("+strings.Join(holders, ",")+")")
	}
	if keyword := strings.TrimSpace(query.DealStaffID); keyword != "" {
		filters = append(filters, "(CAST(l.deal_staff_id AS CHAR) = ? OR l.deal_staff_name LIKE ?)")
		args = append(args, keyword, "%"+keyword+"%")
	}
	if keyword := strings.TrimSpace(query.ConfirmStaffID); keyword != "" {
		filters = append(filters, "(CAST(l.confirm_staff_id AS CHAR) = ? OR l.confirm_staff_name LIKE ?)")
		args = append(args, keyword, "%"+keyword+"%")
	}
	if keyword := strings.TrimSpace(query.StudentID); keyword != "" {
		filters = append(filters, "(CAST(l.student_id AS CHAR) = ? OR l.student_name LIKE ? OR l.student_phone_raw LIKE ?)")
		args = append(args, keyword, "%"+keyword+"%", "%"+keyword+"%")
	}
	if keyword := strings.TrimSpace(query.OrderNumber); keyword != "" {
		filters = append(filters, "l.order_number LIKE ?")
		args = append(args, "%"+keyword+"%")
	}
	if keyword := strings.TrimSpace(query.BankSlipNo); keyword != "" {
		filters = append(filters, "l.bank_slip_no LIKE ?")
		args = append(args, "%"+keyword+"%")
	}
	if keyword := strings.TrimSpace(query.LedgerNumber); keyword != "" {
		filters = append(filters, "l.ledger_number LIKE ?")
		args = append(args, "%"+keyword+"%")
	}
	if begin := parseDateStart(query.ConfirmStartTime); begin != nil {
		filters = append(filters, "l.confirm_time >= ?")
		args = append(args, *begin)
	}
	if end := parseDateEnd(query.ConfirmEndTime); end != nil {
		filters = append(filters, "l.confirm_time <= ?")
		args = append(args, *end)
	}
	if begin := parseDateStart(query.PayStartTime); begin != nil {
		filters = append(filters, "l.pay_time >= ?")
		args = append(args, *begin)
	}
	if end := parseDateEnd(query.PayEndTime); end != nil {
		filters = append(filters, "l.pay_time <= ?")
		args = append(args, *end)
	}
	if len(query.LedgerSubCategoryIDs) > 0 {
		holders := make([]string, 0, len(query.LedgerSubCategoryIDs))
		for _, item := range query.LedgerSubCategoryIDs {
			if strings.TrimSpace(item) == "" {
				continue
			}
			holders = append(holders, "?")
			args = append(args, strings.TrimSpace(item))
		}
		if len(holders) > 0 {
			filters = append(filters, "l.ledger_sub_category_id IN ("+strings.Join(holders, ",")+")")
		}
	}
	if keyword := strings.TrimSpace(query.OrderID); keyword != "" {
		filters = append(filters, "(CAST(l.order_id AS CHAR) = ? OR l.order_number LIKE ?)")
		args = append(args, keyword, "%"+keyword+"%")
	}

	return strings.Join(filters, " AND "), args
}

func parseJSONStringArray(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "null" {
		return []string{}
	}
	var items []string
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return []string{}
	}
	return items
}

func normalizeLedgerImages(images []string) []string {
	if len(images) == 0 {
		return []string{}
	}
	result := make([]string, 0, len(images))
	for _, item := range images {
		value := strings.TrimSpace(item)
		if value == "" {
			continue
		}
		result = append(result, value)
	}
	return result
}
