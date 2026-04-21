package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

type manualLedgerCategoryMeta struct {
	Type            int
	CategoryID      string
	CategoryName    string
	SubCategoryID   string
	SubCategoryName string
	CategoryIcon    string
}

var manualLedgerCategoryCatalog = map[string]manualLedgerCategoryMeta{
	model.LedgerSubCategoryManualExam: {
		Type:            model.LedgerTypeIncome,
		CategoryID:      model.LedgerCategoryManualOtherBusiness,
		CategoryName:    "其他业务",
		SubCategoryID:   model.LedgerSubCategoryManualExam,
		SubCategoryName: "考试费用",
		CategoryIcon:    "manualTallyBookType1",
	},
	model.LedgerSubCategoryManualShow: {
		Type:            model.LedgerTypeIncome,
		CategoryID:      model.LedgerCategoryManualOtherBusiness,
		CategoryName:    "其他业务",
		SubCategoryID:   model.LedgerSubCategoryManualShow,
		SubCategoryName: "演出费用",
		CategoryIcon:    "manualTallyBookType1",
	},
	model.LedgerSubCategoryManualInstrument: {
		Type:            model.LedgerTypeIncome,
		CategoryID:      model.LedgerCategoryManualOtherBusiness,
		CategoryName:    "其他业务",
		SubCategoryID:   model.LedgerSubCategoryManualInstrument,
		SubCategoryName: "乐器费用",
		CategoryIcon:    "manualTallyBookType1",
	},
	model.LedgerSubCategoryManualMeal: {
		Type:            model.LedgerTypeIncome,
		CategoryID:      model.LedgerCategoryManualOtherBusiness,
		CategoryName:    "其他业务",
		SubCategoryID:   model.LedgerSubCategoryManualMeal,
		SubCategoryName: "餐饮费用",
		CategoryIcon:    "manualTallyBookType1",
	},
	model.LedgerSubCategoryManualOther: {
		Type:            model.LedgerTypeIncome,
		CategoryID:      model.LedgerCategoryManualOtherBusiness,
		CategoryName:    "其他业务",
		SubCategoryID:   model.LedgerSubCategoryManualOther,
		SubCategoryName: "其他",
		CategoryIcon:    "manualTallyBookType1",
	},
	model.LedgerSubCategoryManualOffice: {
		Type:            model.LedgerTypeExpenditure,
		CategoryID:      model.LedgerCategoryManualManagementExpense,
		CategoryName:    "管理费用",
		SubCategoryID:   model.LedgerSubCategoryManualOffice,
		SubCategoryName: "办公用品",
		CategoryIcon:    "manualTallyBookType1",
	},
	model.LedgerSubCategoryManualWater: {
		Type:            model.LedgerTypeExpenditure,
		CategoryID:      model.LedgerCategoryManualManagementExpense,
		CategoryName:    "管理费用",
		SubCategoryID:   model.LedgerSubCategoryManualWater,
		SubCategoryName: "水费",
		CategoryIcon:    "manualTallyBookType1",
	},
	model.LedgerSubCategoryManualElectricity: {
		Type:            model.LedgerTypeExpenditure,
		CategoryID:      model.LedgerCategoryManualManagementExpense,
		CategoryName:    "管理费用",
		SubCategoryID:   model.LedgerSubCategoryManualElectricity,
		SubCategoryName: "电费",
		CategoryIcon:    "manualTallyBookType1",
	},
	model.LedgerSubCategoryManualRent: {
		Type:            model.LedgerTypeExpenditure,
		CategoryID:      model.LedgerCategoryManualManagementExpense,
		CategoryName:    "管理费用",
		SubCategoryID:   model.LedgerSubCategoryManualRent,
		SubCategoryName: "房租",
		CategoryIcon:    "manualTallyBookType1",
	},
	model.LedgerSubCategoryManualProperty: {
		Type:            model.LedgerTypeExpenditure,
		CategoryID:      model.LedgerCategoryManualManagementExpense,
		CategoryName:    "管理费用",
		SubCategoryID:   model.LedgerSubCategoryManualProperty,
		SubCategoryName: "物业",
		CategoryIcon:    "manualTallyBookType1",
	},
	model.LedgerSubCategoryManualSalary: {
		Type:            model.LedgerTypeExpenditure,
		CategoryID:      model.LedgerCategoryManualManagementExpense,
		CategoryName:    "管理费用",
		SubCategoryID:   model.LedgerSubCategoryManualSalary,
		SubCategoryName: "工资",
		CategoryIcon:    "manualTallyBookType1",
	},
	model.LedgerSubCategoryManualFund: {
		Type:            model.LedgerTypeExpenditure,
		CategoryID:      model.LedgerCategoryManualManagementExpense,
		CategoryName:    "管理费用",
		SubCategoryID:   model.LedgerSubCategoryManualFund,
		SubCategoryName: "公积金",
		CategoryIcon:    "manualTallyBookType1",
	},
	model.LedgerSubCategoryManualInsurance: {
		Type:            model.LedgerTypeExpenditure,
		CategoryID:      model.LedgerCategoryManualManagementExpense,
		CategoryName:    "管理费用",
		SubCategoryID:   model.LedgerSubCategoryManualInsurance,
		SubCategoryName: "社保",
		CategoryIcon:    "manualTallyBookType1",
	},
	model.LedgerSubCategoryManualMarketing: {
		Type:            model.LedgerTypeExpenditure,
		CategoryID:      model.LedgerCategoryManualSalesExpense,
		CategoryName:    "销售费用",
		SubCategoryID:   model.LedgerSubCategoryManualMarketing,
		SubCategoryName: "营销费用",
		CategoryIcon:    "manualTallyBookType1",
	},
	model.LedgerSubCategoryManualTax: {
		Type:            model.LedgerTypeExpenditure,
		CategoryID:      model.LedgerCategoryManualFinanceExpense,
		CategoryName:    "财务费用",
		SubCategoryID:   model.LedgerSubCategoryManualTax,
		SubCategoryName: "税",
		CategoryIcon:    "manualTallyBookType1",
	},
}

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
			CASE
				WHEN IFNULL(so.order_type, 1) = ? THEN ?
				WHEN IFNULL(pd.pay_amount, 0) >= 0 THEN ?
				ELSE ?
			END,
			CONCAT(DATE_FORMAT(pd.create_time, '%Y%m%d%H%i%s'), LPAD(MOD(pd.id, 1000000), 6, '0')),
			?, ?, 
			CASE
				WHEN IFNULL(so.order_type, 1) = ? THEN ?
				WHEN IFNULL(so.order_type, 1) = ? THEN ?
				WHEN IFNULL(so.order_type, 1) = ? THEN ?
				WHEN IFNULL(so.order_type, 1) = ? THEN ?
				ELSE ?
			END,
			CASE
				WHEN IFNULL(so.order_type, 1) = ? THEN '储值账户充值'
				WHEN IFNULL(so.order_type, 1) = ? THEN '退课'
				WHEN IFNULL(so.order_type, 1) = ? THEN '转课'
				WHEN IFNULL(so.order_type, 1) = ? THEN '储值账户退费'
				ELSE '报名续费'
			END,
			'systemTallyBookType1',
			ABS(IFNULL(pd.pay_amount, 0)),
			IFNULL(pd.create_id, 0),
			IFNULL(operator.nick_name, ''),
			pd.pay_time,
			pd.pay_method,
			IFNULL(pd.amount_id, 0),
			'默认账户',
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
		WHERE pd.inst_id = ? AND pd.del_flag = 0 AND ABS(IFNULL(pd.pay_amount, 0)) > 0 AND l.id IS NULL
	`,
		model.LedgerSourceSystem,
		model.LedgerSystemTypeOrderPayment,
		1,
		model.OrderTypeRechargeAccountRefund,
		model.LedgerTypeExpenditure,
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
		model.OrderTypeRechargeAccountRefund,
		model.LedgerSubCategoryRechargeAccountRefund,
		model.LedgerSubCategoryRegistration,
		model.OrderTypeRechargeAccount,
		model.OrderTypeRefundCourse,
		model.OrderTypeTransferCourse,
		model.OrderTypeRechargeAccountRefund,
		model.LedgerConfirmStatusPending,
		model.LedgerSourceSystem,
		1,
		instID,
	)
	if err != nil {
		return err
	}
	return repo.normalizeSystemLedgerAccountNames(ctx, instID)
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
			CASE
				WHEN IFNULL(so.order_type, 1) = ? THEN ?
				WHEN IFNULL(pd.pay_amount, 0) >= 0 THEN ?
				ELSE ?
			END,
			CONCAT(DATE_FORMAT(pd.create_time, '%Y%m%d%H%i%s'), LPAD(MOD(pd.id, 1000000), 6, '0')),
			?, ?, 
			CASE
				WHEN IFNULL(so.order_type, 1) = ? THEN ?
				WHEN IFNULL(so.order_type, 1) = ? THEN ?
				WHEN IFNULL(so.order_type, 1) = ? THEN ?
				WHEN IFNULL(so.order_type, 1) = ? THEN ?
				ELSE ?
			END,
			CASE
				WHEN IFNULL(so.order_type, 1) = ? THEN '储值账户充值'
				WHEN IFNULL(so.order_type, 1) = ? THEN '退课'
				WHEN IFNULL(so.order_type, 1) = ? THEN '转课'
				WHEN IFNULL(so.order_type, 1) = ? THEN '储值账户退费'
				ELSE '报名续费'
			END,
			'systemTallyBookType1',
			ABS(IFNULL(pd.pay_amount, 0)),
			IFNULL(pd.create_id, 0),
			IFNULL(operator.nick_name, ''),
			pd.pay_time,
			pd.pay_method,
			IFNULL(pd.amount_id, 0),
			'默认账户',
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
		WHERE pd.id = ? AND pd.inst_id = ? AND ABS(IFNULL(pd.pay_amount, 0)) > 0
		ON DUPLICATE KEY UPDATE
			amount = VALUES(amount),
			type = VALUES(type),
			ledger_sub_category_id = VALUES(ledger_sub_category_id),
			ledger_sub_category_name = VALUES(ledger_sub_category_name),
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
		model.OrderTypeRechargeAccountRefund,
		model.LedgerTypeExpenditure,
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
		model.OrderTypeRechargeAccountRefund,
		model.LedgerSubCategoryRechargeAccountRefund,
		model.LedgerSubCategoryRegistration,
		model.OrderTypeRechargeAccount,
		model.OrderTypeRefundCourse,
		model.OrderTypeTransferCourse,
		model.OrderTypeRechargeAccountRefund,
		model.LedgerConfirmStatusPending,
		paymentDetailID,
		instID,
	)
	// Avoid cross-connection cleanup inside the payment transaction.
	// The normalization pass is maintenance work and can wait for non-transactional paths.
	return err
}

func (repo *Repository) normalizeSystemLedgerAccountNames(ctx context.Context, instID int64) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE inst_ledger
		SET account_name = '默认账户', update_time = NOW()
		WHERE inst_id = ?
		  AND del_flag = 0
		  AND source_type = ?
		  AND system_type = ?
		  AND (account_name = '' OR account_name LIKE '账户%')
	`, instID, model.LedgerSourceSystem, model.LedgerSystemTypeOrderPayment)
	return err
}

type manualLedgerMutationInput struct {
	LedgerID        int64
	Amount          float64
	Remark          string
	ImagesJSON      string
	PayTime         time.Time
	PayMethod       int
	DealStaffID     int64
	DealStaffName   string
	AccountID       int64
	AccountName     string
	LedgerType      int
	CategoryID      string
	CategoryName    string
	SubCategoryID   string
	SubCategoryName string
	CategoryIcon    string
	SourceBizType   int
	SourceBizID     int64
	LedgerNumber    string
}

func (repo *Repository) CreateManualLedger(ctx context.Context, instID, operatorID int64, dto model.ManualLedgerSaveDTO) (int64, error) {
	input, err := repo.buildManualLedgerMutation(ctx, instID, dto)
	if err != nil {
		return 0, err
	}
	result, err := repo.db.ExecContext(ctx, `
		INSERT INTO inst_ledger (
			uuid, version, inst_id, source_type, system_type, source_biz_type, source_biz_id,
			type, ledger_number, ledger_category_id, ledger_category_name, ledger_sub_category_id,
			ledger_sub_category_name, ledger_category_icon, amount, deal_staff_id, deal_staff_name,
			pay_time, pay_method, account_id, account_name, reciprocal_account, bank_slip_no,
			order_id, order_number, student_id, student_name, student_phone, student_phone_raw,
			payment_voucher_text, payment_voucher_images, ledger_confirm_status, confirm_staff_id,
			confirm_staff_name, confirm_time, confirm_remark_text, confirm_remark_images,
			bill_flow_id, bill_id, error_message, create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, ?, 0, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?, '', '',
			0, '', 0, '', '', '',
			?, ?, ?, 0,
			'', NULL, '', ?,
			0, 0, '', ?, NOW(), ?, NOW(), 0
		)
	`,
		instID, model.LedgerSourceManual, input.SourceBizType, input.SourceBizID,
		input.LedgerType, input.LedgerNumber, input.CategoryID, input.CategoryName, input.SubCategoryID, input.SubCategoryName,
		input.CategoryIcon, input.Amount, input.DealStaffID, input.DealStaffName, input.PayTime, input.PayMethod, input.AccountID, input.AccountName,
		input.Remark, input.ImagesJSON, model.LedgerConfirmStatusPending, "[]", operatorID, operatorID,
	)
	if err != nil {
		return 0, err
	}
	ledgerID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return ledgerID, nil
}

func (repo *Repository) UpdateManualLedger(ctx context.Context, instID, operatorID int64, dto model.ManualLedgerSaveDTO) (int64, error) {
	ledgerID, status, err := repo.loadManualLedgerForMutation(ctx, instID, dto.ID)
	if err != nil {
		return 0, err
	}
	if status != model.LedgerConfirmStatusPending {
		return 0, errors.New("已确认账单不支持编辑")
	}
	input, err := repo.buildManualLedgerMutation(ctx, instID, dto)
	if err != nil {
		return 0, err
	}
	_, err = repo.db.ExecContext(ctx, `
		UPDATE inst_ledger
		SET type = ?,
			ledger_category_id = ?,
			ledger_category_name = ?,
			ledger_sub_category_id = ?,
			ledger_sub_category_name = ?,
			ledger_category_icon = ?,
			amount = ?,
			deal_staff_id = ?,
			deal_staff_name = ?,
			pay_time = ?,
			pay_method = ?,
			account_id = ?,
			account_name = ?,
			payment_voucher_text = ?,
			payment_voucher_images = ?,
			update_id = ?,
			update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`,
		input.LedgerType,
		input.CategoryID,
		input.CategoryName,
		input.SubCategoryID,
		input.SubCategoryName,
		input.CategoryIcon,
		input.Amount,
		input.DealStaffID,
		input.DealStaffName,
		input.PayTime,
		input.PayMethod,
		input.AccountID,
		input.AccountName,
		input.Remark,
		input.ImagesJSON,
		operatorID,
		ledgerID,
		instID,
	)
	if err != nil {
		return 0, err
	}
	return ledgerID, nil
}

func (repo *Repository) DeleteManualLedger(ctx context.Context, instID, operatorID, ledgerID int64) error {
	_, status, err := repo.loadManualLedgerForMutationByID(ctx, instID, ledgerID)
	if err != nil {
		return err
	}
	if status != model.LedgerConfirmStatusPending {
		return errors.New("已确认账单不支持删除")
	}
	_, err = repo.db.ExecContext(ctx, `
		UPDATE inst_ledger
		SET del_flag = 1,
			update_id = ?,
			update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, operatorID, ledgerID, instID)
	return err
}

func (repo *Repository) buildManualLedgerMutation(ctx context.Context, instID int64, dto model.ManualLedgerSaveDTO) (manualLedgerMutationInput, error) {
	amount := dto.Amount
	if amount <= 0 {
		return manualLedgerMutationInput{}, errors.New("账单金额必须大于0")
	}
	if amount > 999999999 {
		return manualLedgerMutationInput{}, errors.New("账单金额不能超过999999999")
	}

	meta, err := resolveManualLedgerCategory(dto.Type, dto.LedgerCategoryID, dto.LedgerSubCategoryID)
	if err != nil {
		return manualLedgerMutationInput{}, err
	}
	if dto.PayMethod < 1 || dto.PayMethod > 6 {
		return manualLedgerMutationInput{}, errors.New("请选择收款方式")
	}

	payTime, err := parseManualLedgerPayTime(dto.PayTime)
	if err != nil {
		return manualLedgerMutationInput{}, err
	}

	dealStaffID, err := strconv.ParseInt(strings.TrimSpace(dto.DealStaffID), 10, 64)
	if err != nil || dealStaffID <= 0 {
		return manualLedgerMutationInput{}, errors.New("请选择经办人")
	}
	dealStaffName, err := repo.findInstStaffNameByID(ctx, instID, dealStaffID)
	if err != nil {
		return manualLedgerMutationInput{}, err
	}

	accountID := int64(1)
	if trimmed := strings.TrimSpace(dto.AccountID); trimmed != "" {
		parsedAccountID, parseErr := strconv.ParseInt(trimmed, 10, 64)
		if parseErr != nil || parsedAccountID <= 0 {
			return manualLedgerMutationInput{}, errors.New("支付账户无效")
		}
		accountID = parsedAccountID
	}

	images := normalizeLedgerImages(dto.Images)
	if len(images) > 3 {
		return manualLedgerMutationInput{}, errors.New("最多上传3张图片")
	}
	imagesJSON, err := json.Marshal(images)
	if err != nil {
		return manualLedgerMutationInput{}, err
	}

	now := time.Now()
	return manualLedgerMutationInput{
		Amount:          amount,
		Remark:          strings.TrimSpace(dto.Remark),
		ImagesJSON:      string(imagesJSON),
		PayTime:         payTime,
		PayMethod:       dto.PayMethod,
		DealStaffID:     dealStaffID,
		DealStaffName:   dealStaffName,
		AccountID:       accountID,
		AccountName:     "默认账户",
		LedgerType:      meta.Type,
		CategoryID:      meta.CategoryID,
		CategoryName:    meta.CategoryName,
		SubCategoryID:   meta.SubCategoryID,
		SubCategoryName: meta.SubCategoryName,
		CategoryIcon:    meta.CategoryIcon,
		SourceBizType:   model.LedgerManualBizTypeDefault,
		SourceBizID:     now.UnixNano(),
		LedgerNumber:    generateManualLedgerNumber(now),
	}, nil
}

func resolveManualLedgerCategory(ledgerType int, categoryID, subCategoryID string) (manualLedgerCategoryMeta, error) {
	meta, ok := manualLedgerCategoryCatalog[strings.TrimSpace(subCategoryID)]
	if !ok {
		return manualLedgerCategoryMeta{}, errors.New("请选择账单分类")
	}
	if ledgerType != meta.Type {
		return manualLedgerCategoryMeta{}, errors.New("收支类型与账单分类不匹配")
	}
	if trimmedCategoryID := strings.TrimSpace(categoryID); trimmedCategoryID != "" && trimmedCategoryID != meta.CategoryID {
		return manualLedgerCategoryMeta{}, errors.New("账单分类无效")
	}
	return meta, nil
}

func parseManualLedgerPayTime(raw string) (time.Time, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Time{}, errors.New("请选择支付日期")
	}
	layouts := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		time.RFC3339,
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("支付日期格式无效")
}

func (repo *Repository) findInstStaffNameByID(ctx context.Context, instID, staffID int64) (string, error) {
	var name string
	err := repo.db.QueryRowContext(ctx, `
		SELECT IFNULL(nick_name, '')
		FROM inst_user
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, staffID, instID).Scan(&name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("经办人不存在")
		}
		return "", err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("经办人不存在")
	}
	return name, nil
}

func (repo *Repository) loadManualLedgerForMutation(ctx context.Context, instID int64, rawLedgerID string) (int64, int, error) {
	ledgerID, err := strconv.ParseInt(strings.TrimSpace(rawLedgerID), 10, 64)
	if err != nil || ledgerID <= 0 {
		return 0, 0, errors.New("账单ID不能为空")
	}
	return repo.loadManualLedgerForMutationByID(ctx, instID, ledgerID)
}

func (repo *Repository) loadManualLedgerForMutationByID(ctx context.Context, instID, ledgerID int64) (int64, int, error) {
	var (
		sourceType int
		status     int
	)
	err := repo.db.QueryRowContext(ctx, `
		SELECT source_type, ledger_confirm_status
		FROM inst_ledger
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, ledgerID, instID).Scan(&sourceType, &status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, errors.New("账单不存在")
		}
		return 0, 0, err
	}
	if sourceType != model.LedgerSourceManual {
		return 0, 0, errors.New("系统同步账单不支持此操作")
	}
	return ledgerID, status, nil
}

func generateManualLedgerNumber(now time.Time) string {
	return fmt.Sprintf("%s%06d", now.Format("20060102150405"), now.UnixNano()%1000000)
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
				var orderType *int
				switch item.LedgerSubCategoryID {
				case model.LedgerSubCategoryRechargeAccount:
					t := model.OrderTypeRechargeAccount
					orderType = &t
				case model.LedgerSubCategoryRechargeAccountRefund:
					t := model.OrderTypeRechargeAccountRefund
					orderType = &t
				}
				var err error
				products, err = repo.getOrderDisplayItems(ctx, orderID.Int64, orderType)
				if err != nil {
					products = nil
				}
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

	if len(query.LedgerIDs) > 0 {
		holders := make([]string, 0, len(query.LedgerIDs))
		for _, item := range query.LedgerIDs {
			if strings.TrimSpace(item) == "" {
				continue
			}
			holders = append(holders, "?")
			args = append(args, strings.TrimSpace(item))
		}
		if len(holders) > 0 {
			filters = append(filters, "CAST(l.id AS CHAR) IN ("+strings.Join(holders, ",")+")")
		}
	}
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
