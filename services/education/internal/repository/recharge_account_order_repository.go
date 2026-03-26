package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

func ensureRechargeAccountOrderTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS recharge_account_order (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			uuid VARCHAR(64) NULL,
			version BIGINT NOT NULL DEFAULT 0,
			inst_id BIGINT NOT NULL,
			recharge_account_id BIGINT NOT NULL,
			sale_order_id BIGINT NOT NULL DEFAULT 0,
			order_number VARCHAR(64) NOT NULL DEFAULT '',
			status INT NOT NULL DEFAULT 1,
			amount DECIMAL(18,2) NOT NULL DEFAULT 0,
			giving_amount DECIMAL(18,2) NOT NULL DEFAULT 0,
			residual_amount DECIMAL(18,2) NOT NULL DEFAULT 0,
			deal_date DATE NULL,
			sale_person_id BIGINT NOT NULL DEFAULT 0,
			collector_staff_id BIGINT NOT NULL DEFAULT 0,
			phone_sell_staff_id BIGINT NOT NULL DEFAULT 0,
			foreground_staff_id BIGINT NOT NULL DEFAULT 0,
			vice_sell_staff_staff_id BIGINT NOT NULL DEFAULT 0,
			remark VARCHAR(500) NOT NULL DEFAULT '',
			external_remark VARCHAR(500) NOT NULL DEFAULT '',
			student_id BIGINT NOT NULL DEFAULT 0,
			bill_id BIGINT NOT NULL DEFAULT 0,
			approve_id BIGINT NULL,
			order_obsolete VARCHAR(255) NULL,
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT NOT NULL DEFAULT 0,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uk_recharge_account_order_number (inst_id, order_number),
			KEY idx_recharge_account_order_inst (inst_id, create_time, id),
			KEY idx_recharge_account_order_student (inst_id, student_id)
		)
	`)
	if err != nil {
		return err
	}
	if err := ensureColumnsOnTable(ctx, db, "recharge_account_order", map[string]string{
		"sale_order_id": "sale_order_id BIGINT NOT NULL DEFAULT 0 AFTER recharge_account_id",
	}); err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS recharge_account_order_tag (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			inst_id BIGINT NOT NULL,
			recharge_account_order_id BIGINT NOT NULL,
			tag_id BIGINT NOT NULL,
			tag_name VARCHAR(100) NOT NULL DEFAULT '',
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			KEY idx_recharge_account_order_tag_order (inst_id, recharge_account_order_id)
		)
	`)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS recharge_account_bill (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			inst_id BIGINT NOT NULL,
			recharge_account_order_id BIGINT NOT NULL,
			status INT NOT NULL DEFAULT 1,
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT NOT NULL DEFAULT 0,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			KEY idx_recharge_account_bill_order (inst_id, recharge_account_order_id)
		)
	`)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS recharge_account_bill_flow (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			inst_id BIGINT NOT NULL,
			bill_id BIGINT NOT NULL,
			amount DECIMAL(18,2) NOT NULL DEFAULT 0,
			remark VARCHAR(500) NOT NULL DEFAULT '',
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			KEY idx_recharge_account_bill_flow_bill (inst_id, bill_id)
		)
	`)
	return err
}

func (repo *Repository) GetStudentDetailView(ctx context.Context, instID, studentID int64) (model.StudentDetailView, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT
			s.id,
			IFNULL(s.stu_name, ''),
			CASE
				WHEN CHAR_LENGTH(IFNULL(s.mobile, '')) >= 7 THEN CONCAT(LEFT(s.mobile, 3), '****', RIGHT(s.mobile, 4))
				ELSE IFNULL(s.mobile, '')
			END,
			IFNULL(s.avatar_url, ''),
			IFNULL(s.stu_sex, 0),
			IFNULL(s.phone_relationship, 0),
			IFNULL(s.sale_person, 0),
			IFNULL(sale.nick_name, ''),
			s.create_time,
			(SELECT MIN(so.create_time) FROM sale_order so WHERE so.student_id = s.id AND so.del_flag = 0),
			NULL,
			IFNULL(s.create_id, 0),
			IFNULL(creator.nick_name, ''),
			IFNULL(s.collector_staff_id, 0),
			IFNULL(collector.nick_name, ''),
			IFNULL(s.phone_sell_staff_id, 0),
			IFNULL(phone_sell.nick_name, ''),
			IFNULL(s.foreground_staff_id, 0),
			IFNULL(foreground.nick_name, ''),
			IFNULL(s.vice_sell_staff_id, 0),
			IFNULL(vice.nick_name, ''),
			IFNULL(s.student_status, 0)
		FROM inst_student s
		LEFT JOIN inst_user sale ON sale.id = s.sale_person
		LEFT JOIN inst_user creator ON creator.id = s.create_id
		LEFT JOIN inst_user collector ON collector.id = s.collector_staff_id
		LEFT JOIN inst_user phone_sell ON phone_sell.id = s.phone_sell_staff_id
		LEFT JOIN inst_user foreground ON foreground.id = s.foreground_staff_id
		LEFT JOIN inst_user vice ON vice.id = s.vice_sell_staff_id
		WHERE s.inst_id = ? AND s.id = ? AND s.del_flag = 0
		LIMIT 1
	`, instID, studentID)

	var (
		item              model.StudentDetailView
		id                int64
		salespersonID     int64
		createdStaffID    int64
		collectorStaffID  int64
		phoneSellStaffID  int64
		foregroundStaffID int64
		viceSellStaffID   int64
		createdTime       sql.NullTime
		firstEnrolledTime sql.NullTime
		turnedHistoryTime sql.NullTime
	)
	if err := row.Scan(
		&id,
		&item.Name,
		&item.Phone,
		&item.Avatar,
		&item.Sex,
		&item.PhoneRelationship,
		&salespersonID,
		&item.SalespersonName,
		&createdTime,
		&firstEnrolledTime,
		&turnedHistoryTime,
		&createdStaffID,
		&item.CreatedStaffName,
		&collectorStaffID,
		&item.CollectorStaffName,
		&phoneSellStaffID,
		&item.PhoneSellStaffName,
		&foregroundStaffID,
		&item.ForegroundStaffName,
		&viceSellStaffID,
		&item.ViceSellStaffStaffName,
		&item.Status,
	); err != nil {
		return model.StudentDetailView{}, err
	}
	item.ID = strconv.FormatInt(id, 10)
	item.SalespersonID = strconv.FormatInt(salespersonID, 10)
	item.CreatedStaffID = strconv.FormatInt(createdStaffID, 10)
	item.CollectorStaffID = strconv.FormatInt(collectorStaffID, 10)
	item.PhoneSellStaffID = strconv.FormatInt(phoneSellStaffID, 10)
	item.ForegroundStaffID = strconv.FormatInt(foregroundStaffID, 10)
	item.ViceSellStaffStaffID = strconv.FormatInt(viceSellStaffID, 10)
	if createdTime.Valid {
		t := createdTime.Time
		item.CreatedTime = &t
	}
	if firstEnrolledTime.Valid {
		t := firstEnrolledTime.Time
		item.FirstEnrolledTime = &t
	}
	if turnedHistoryTime.Valid {
		t := turnedHistoryTime.Time
		item.TurnedHistoryTime = &t
	}
	return item, nil
}

func (repo *Repository) GetRechargeAccountByStudent(ctx context.Context, instID, studentID, operatorID int64) (model.RechargeAccountByStudent, error) {
	if err := repo.EnsureRechargeAccount(ctx, instID, studentID, operatorID); err != nil {
		return model.RechargeAccountByStudent{}, err
	}

	row := repo.db.QueryRowContext(ctx, `
		SELECT
			ra.id,
			IFNULL(ra.account_name, ''),
			IFNULL(ra.phone, ''),
			ra.main_student_id,
			IFNULL(ra.recharge_balance, 0) + IFNULL(ra.residual_balance, 0),
			IFNULL(ra.giving_balance, 0),
			IFNULL(ra.residual_balance, 0),
			ra.create_time
		FROM recharge_account ra
		INNER JOIN recharge_account_student ras ON ras.recharge_account_id = ra.id AND ras.del_flag = 0
		WHERE ra.inst_id = ? AND ra.del_flag = 0 AND ras.student_id = ?
		ORDER BY ra.create_time ASC, ra.id ASC
		LIMIT 1
	`, instID, studentID)

	var (
		item          model.RechargeAccountByStudent
		accountID     int64
		mainStudentID sql.NullInt64
		createdAt     sql.NullTime
	)
	if err := row.Scan(&accountID, &item.AccountName, &item.Phone, &mainStudentID, &item.Balance, &item.GivingBalance, &item.ResidualBalance, &createdAt); err != nil {
		return model.RechargeAccountByStudent{}, err
	}
	item.ID = strconv.FormatInt(accountID, 10)
	if mainStudentID.Valid {
		item.MainStudentID = strconv.FormatInt(mainStudentID.Int64, 10)
	}
	item.AccountName = normalizeRechargeAccountName(item.AccountName, item.MainStudentID, item.ID)
	if createdAt.Valid {
		t := createdAt.Time
		item.CreatedAt = &t
	}

	studentsMap, err := repo.listRechargeAccountStudents(ctx, instID, []int64{accountID})
	if err != nil {
		return model.RechargeAccountByStudent{}, err
	}
	for _, stu := range studentsMap[accountID] {
		detail, detailErr := repo.GetStudentDetailView(ctx, instID, mustInt64(stu.StudentID))
		if detailErr != nil {
			continue
		}
		item.Students = append(item.Students, model.RechargeAccountByStudentStudent{
			ID:            stu.StudentID,
			Name:          stu.StudentName,
			Avatar:        detail.Avatar,
			Sex:           detail.Sex,
			Phone:         detail.Phone,
			IsMainStudent: stu.IsMainStudent,
		})
	}
	return item, nil
}

// GetRechargeAccountByID loads a recharge account by id (SchoolPal-compatible shape for refund drawer).
func (repo *Repository) GetRechargeAccountByID(ctx context.Context, instID, rechargeAccountID int64) (model.RechargeAccountByStudent, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT
			ra.id,
			IFNULL(ra.account_name, ''),
			IFNULL(ra.phone, ''),
			ra.main_student_id,
			IFNULL(ra.recharge_balance, 0) + IFNULL(ra.residual_balance, 0),
			IFNULL(ra.giving_balance, 0),
			IFNULL(ra.residual_balance, 0),
			ra.create_time
		FROM recharge_account ra
		WHERE ra.inst_id = ? AND ra.id = ? AND ra.del_flag = 0
		LIMIT 1
	`, instID, rechargeAccountID)

	var (
		item          model.RechargeAccountByStudent
		accountID     int64
		mainStudentID sql.NullInt64
		createdAt     sql.NullTime
	)
	if err := row.Scan(&accountID, &item.AccountName, &item.Phone, &mainStudentID, &item.Balance, &item.GivingBalance, &item.ResidualBalance, &createdAt); err != nil {
		return model.RechargeAccountByStudent{}, err
	}
	item.ID = strconv.FormatInt(accountID, 10)
	if mainStudentID.Valid {
		item.MainStudentID = strconv.FormatInt(mainStudentID.Int64, 10)
	}
	item.AccountName = normalizeRechargeAccountName(item.AccountName, item.MainStudentID, item.ID)
	if createdAt.Valid {
		t := createdAt.Time
		item.CreatedAt = &t
	}

	studentsMap, err := repo.listRechargeAccountStudents(ctx, instID, []int64{accountID})
	if err != nil {
		return model.RechargeAccountByStudent{}, err
	}
	for _, stu := range studentsMap[accountID] {
		detail, detailErr := repo.GetStudentDetailView(ctx, instID, mustInt64(stu.StudentID))
		if detailErr != nil {
			continue
		}
		item.Students = append(item.Students, model.RechargeAccountByStudentStudent{
			ID:            stu.StudentID,
			Name:          stu.StudentName,
			Avatar:        detail.Avatar,
			Sex:           detail.Sex,
			Phone:         detail.Phone,
			IsMainStudent: stu.IsMainStudent,
		})
	}
	return item, nil
}

func (repo *Repository) CreateRechargeAccountOrder(ctx context.Context, instID, operatorID int64, dto model.CreateRechargeAccountOrderDTO) (int64, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	rechargeAccountID, err := strconv.ParseInt(strings.TrimSpace(dto.RechargeAccountID), 10, 64)
	if err != nil || rechargeAccountID <= 0 {
		return 0, errors.New("储值账户不存在")
	}
	studentID, err := strconv.ParseInt(strings.TrimSpace(dto.StudentID), 10, 64)
	if err != nil || studentID <= 0 {
		return 0, errors.New("学生不存在")
	}

	var dealDate any
	if parsed := parseDateStart(dto.DealDate); parsed != nil {
		dealDate = parsed.Format("2006-01-02")
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO recharge_account_order (
			uuid, version, inst_id, recharge_account_id, sale_order_id, order_number, status, amount, giving_amount, residual_amount,
			deal_date, sale_person_id, collector_staff_id, phone_sell_staff_id, foreground_staff_id, vice_sell_staff_staff_id,
			remark, external_remark, student_id, create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, ?, 0, '', 1, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, NOW(), ?, NOW(), 0
		)
	`,
		instID, rechargeAccountID, dto.Amount, dto.GivingAmount, dto.ResidualAmount,
		dealDate, parseInt64String(dto.SalePersonID), parseInt64String(dto.CollectorStaffID), parseInt64String(dto.PhoneSellStaffID),
		parseInt64String(dto.ForegroundStaffID), parseInt64String(dto.ViceSellStaffStaffID),
		strings.TrimSpace(dto.Remark), strings.TrimSpace(dto.ExternalRemark), studentID, operatorID, operatorID,
	)
	if err != nil {
		return 0, err
	}
	orderID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	orderNumber := fmt.Sprintf("%s%06d", time.Now().Format("20060102150405"), orderID%1000000)
	orderTagIDs := joinRechargeOrderTagIDs(dto.OrderTagIDs)
	saleOrderResult, err := tx.ExecContext(ctx, `
		INSERT INTO sale_order (
			uuid, version, inst_id, student_id, order_number, sale_person, deal_date, order_discount_type,
			order_discount_amount, order_discount_number, order_real_amount, order_tag_ids, internal_remark,
			external_remark, order_type, order_status, order_source, create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, ?, ?, ?, ?, NULL, 0, 0, ?, ?, ?, ?, ?, ?, 1, ?, NOW(), ?, NOW(), 0
		)
	`,
		instID,
		studentID,
		orderNumber,
		parseInt64String(dto.SalePersonID),
		dealDate,
		dto.Amount,
		orderTagIDs,
		strings.TrimSpace(dto.Remark),
		strings.TrimSpace(dto.ExternalRemark),
		model.OrderTypeRechargeAccount,
		model.OrderStatusPendingPayment,
		operatorID,
		operatorID,
	)
	if err != nil {
		return 0, err
	}
	saleOrderID, err := saleOrderResult.LastInsertId()
	if err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE recharge_account_order
		SET sale_order_id = ?, order_number = ?, update_id = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ?
	`, saleOrderID, orderNumber, operatorID, orderID, instID); err != nil {
		return 0, err
	}
	billResult, err := tx.ExecContext(ctx, `
		INSERT INTO recharge_account_bill (
			inst_id, recharge_account_order_id, status, create_id, create_time, update_id, update_time, del_flag
		) VALUES (?, ?, 1, ?, NOW(), ?, NOW(), 0)
	`, instID, orderID, operatorID, operatorID)
	if err != nil {
		return 0, err
	}
	billID, err := billResult.LastInsertId()
	if err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE recharge_account_order
		SET bill_id = ?, update_id = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ?
	`, billID, operatorID, orderID, instID); err != nil {
		return 0, err
	}
	for _, tagIDRaw := range dto.OrderTagIDs {
		tagID := parseInt64String(tagIDRaw)
		if tagID <= 0 {
			continue
		}
		var tagName string
		if err := tx.QueryRowContext(ctx, `
			SELECT IFNULL(name, '')
			FROM inst_order_tag
			WHERE id = ? AND inst_id = ? AND del_flag = 0
			LIMIT 1
		`, tagID, instID).Scan(&tagName); err != nil {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO recharge_account_order_tag (
				inst_id, recharge_account_order_id, tag_id, tag_name, create_time, del_flag
			) VALUES (?, ?, ?, ?, NOW(), 0)
		`, instID, orderID, tagID, tagName); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return orderID, nil
}

// CreateRechargeAccountRefundOrder creates a pending 储值账户退费 sale_order and recharge_account_order (same bill flow as recharge).
func (repo *Repository) CreateRechargeAccountRefundOrder(ctx context.Context, instID, operatorID int64, dto model.CreateRechargeAccountOrderDTO) (int64, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	rechargeAccountID, err := strconv.ParseInt(strings.TrimSpace(dto.RechargeAccountID), 10, 64)
	if err != nil || rechargeAccountID <= 0 {
		return 0, errors.New("储值账户不存在")
	}
	var rechBal, resBal, givingBal float64
	if err := tx.QueryRowContext(ctx, `
		SELECT IFNULL(recharge_balance, 0), IFNULL(residual_balance, 0), IFNULL(giving_balance, 0)
		FROM recharge_account
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, rechargeAccountID, instID).Scan(&rechBal, &resBal, &givingBal); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("储值账户不存在")
		}
		return 0, err
	}
	if dto.Amount <= 0 && dto.ResidualAmount <= 0 && dto.GivingAmount <= 0 {
		return 0, errors.New("退费金额、残联金额或赠送扣减至少填写一项")
	}
	if dto.Amount > rechBal+1e-9 {
		return 0, errors.New("充值余额不足")
	}
	if dto.ResidualAmount > resBal+1e-9 {
		return 0, errors.New("残联余额不足")
	}
	if dto.GivingAmount > givingBal+1e-9 {
		return 0, errors.New("赠送余额不足")
	}

	studentID, err := strconv.ParseInt(strings.TrimSpace(dto.StudentID), 10, 64)
	if err != nil || studentID <= 0 {
		var mainStu sql.NullInt64
		if err := tx.QueryRowContext(ctx, `
			SELECT main_student_id FROM recharge_account WHERE id = ? AND inst_id = ? AND del_flag = 0 LIMIT 1
		`, rechargeAccountID, instID).Scan(&mainStu); err != nil || !mainStu.Valid || mainStu.Int64 <= 0 {
			return 0, errors.New("未找到主学员")
		}
		studentID = mainStu.Int64
	}

	var dealDate any
	if parsed := parseDateStart(dto.DealDate); parsed != nil {
		dealDate = parsed.Format("2006-01-02")
	}
	residualRefund := dto.ResidualAmount
	result, err := tx.ExecContext(ctx, `
		INSERT INTO recharge_account_order (
			uuid, version, inst_id, recharge_account_id, sale_order_id, order_number, status, amount, giving_amount, residual_amount,
			deal_date, sale_person_id, collector_staff_id, phone_sell_staff_id, foreground_staff_id, vice_sell_staff_staff_id,
			remark, external_remark, student_id, create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, ?, 0, '', 1, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, NOW(), ?, NOW(), 0
		)
	`,
		instID, rechargeAccountID, dto.Amount, dto.GivingAmount, residualRefund,
		dealDate, parseInt64String(dto.SalePersonID), parseInt64String(dto.CollectorStaffID), parseInt64String(dto.PhoneSellStaffID),
		parseInt64String(dto.ForegroundStaffID), parseInt64String(dto.ViceSellStaffStaffID),
		strings.TrimSpace(dto.Remark), strings.TrimSpace(dto.ExternalRemark), studentID, operatorID, operatorID,
	)
	if err != nil {
		return 0, err
	}
	orderID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	orderNumber := fmt.Sprintf("%s%06d", time.Now().Format("20060102150405"), orderID%1000000)
	orderTagIDs := joinRechargeOrderTagIDs(dto.OrderTagIDs)
	saleOrderResult, err := tx.ExecContext(ctx, `
		INSERT INTO sale_order (
			uuid, version, inst_id, student_id, order_number, sale_person, deal_date, order_discount_type,
			order_discount_amount, order_discount_number, order_real_amount, order_tag_ids, internal_remark,
			external_remark, order_type, order_status, order_source, create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, ?, ?, ?, ?, NULL, 0, 0, ?, ?, ?, ?, ?, ?, 1, ?, NOW(), ?, NOW(), 0
		)
	`,
		instID,
		studentID,
		orderNumber,
		parseInt64String(dto.SalePersonID),
		dealDate,
		dto.Amount,
		orderTagIDs,
		strings.TrimSpace(dto.Remark),
		strings.TrimSpace(dto.ExternalRemark),
		model.OrderTypeRechargeAccountRefund,
		model.OrderStatusPendingPayment,
		operatorID,
		operatorID,
	)
	if err != nil {
		return 0, err
	}
	saleOrderID, err := saleOrderResult.LastInsertId()
	if err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE recharge_account_order
		SET sale_order_id = ?, order_number = ?, update_id = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ?
	`, saleOrderID, orderNumber, operatorID, orderID, instID); err != nil {
		return 0, err
	}
	billResult, err := tx.ExecContext(ctx, `
		INSERT INTO recharge_account_bill (
			inst_id, recharge_account_order_id, status, create_id, create_time, update_id, update_time, del_flag
		) VALUES (?, ?, 1, ?, NOW(), ?, NOW(), 0)
	`, instID, orderID, operatorID, operatorID)
	if err != nil {
		return 0, err
	}
	billID, err := billResult.LastInsertId()
	if err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE recharge_account_order
		SET bill_id = ?, update_id = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ?
	`, billID, operatorID, orderID, instID); err != nil {
		return 0, err
	}
	for _, tagIDRaw := range dto.OrderTagIDs {
		tagID := parseInt64String(tagIDRaw)
		if tagID <= 0 {
			continue
		}
		var tagName string
		if err := tx.QueryRowContext(ctx, `
			SELECT IFNULL(name, '')
			FROM inst_order_tag
			WHERE id = ? AND inst_id = ? AND del_flag = 0
			LIMIT 1
		`, tagID, instID).Scan(&tagName); err != nil {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO recharge_account_order_tag (
				inst_id, recharge_account_order_id, tag_id, tag_name, create_time, del_flag
			) VALUES (?, ?, ?, ?, NOW(), 0)
		`, instID, orderID, tagID, tagName); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return orderID, nil
}

func (repo *Repository) GetRechargeAccountOrderDetail(ctx context.Context, instID, orderID, saleOrderID int64) (model.RechargeAccountOrderDetail, error) {
	whereClause := "rao.inst_id = ? AND rao.id = ? AND rao.del_flag = 0"
	args := []any{instID, orderID}
	if orderID <= 0 {
		whereClause = "rao.inst_id = ? AND rao.sale_order_id = ? AND rao.del_flag = 0"
		args = []any{instID, saleOrderID}
	}
	row := repo.db.QueryRowContext(ctx, `
		SELECT
			rao.id, rao.recharge_account_id, IFNULL(rao.sale_order_id, 0), IFNULL(rao.order_number, ''), IFNULL(rao.status, 0),
			IFNULL(rao.amount, 0), IFNULL(rao.giving_amount, 0), IFNULL(rao.residual_amount, 0),
			IFNULL(op.nick_name, ''), rao.create_time, IFNULL(rao.bill_id, 0),
			IFNULL(rao.student_id, 0), IFNULL(stu.stu_name, ''), IFNULL(stu.mobile, ''),
			IFNULL(rb.status, 0)
		FROM recharge_account_order rao
		LEFT JOIN inst_user op ON op.id = rao.create_id
		LEFT JOIN recharge_account_bill rb ON rb.id = rao.bill_id AND rb.del_flag = 0
		LEFT JOIN inst_student stu ON stu.id = rao.student_id AND stu.del_flag = 0
		WHERE `+whereClause+`
		LIMIT 1
	`, args...)
	var (
		item              model.RechargeAccountOrderDetail
		id                int64
		rechargeAccountID int64
		orderSaleOrderID  int64
		billID            int64
		studentID         int64
		createdAt         sql.NullTime
		billStatus        int
	)
	if err := row.Scan(&id, &rechargeAccountID, &orderSaleOrderID, &item.OrderNumber, &item.Status, &item.Amount, &item.GivingAmount, &item.ResidualAmount, &item.OperatorName, &createdAt, &billID, &studentID, &item.StudentName, &item.StudentPhone, &billStatus); err != nil {
		return model.RechargeAccountOrderDetail{}, err
	}
	item.ID = strconv.FormatInt(id, 10)
	item.RechargeAccountID = strconv.FormatInt(rechargeAccountID, 10)
	item.SaleOrderID = strconv.FormatInt(orderSaleOrderID, 10)
	item.StudentID = strconv.FormatInt(studentID, 10)
	if createdAt.Valid {
		t := createdAt.Time
		item.CreatedAt = &t
	}
	item.Bill = model.RechargeAccountOrderBillDetail{
		ID:        strconv.FormatInt(billID, 10),
		Status:    billStatus,
		BillFlows: []any{},
	}
	tags, err := repo.listRechargeAccountOrderTags(ctx, instID, id)
	if err != nil {
		return model.RechargeAccountOrderDetail{}, err
	}
	item.OrderTags = tags
	return item, nil
}

func (repo *Repository) PayRechargeAccountOrderBySchoolPal(ctx context.Context, instID, operatorID int64, dto model.PayOrderBySchoolPalDTO) (int64, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	billID, err := strconv.ParseInt(strings.TrimSpace(dto.BillID), 10, 64)
	if err != nil || billID <= 0 {
		return 0, errors.New("billId不能为空")
	}

	var (
		orderID           int64
		rechargeAccountID int64
		saleOrderID       int64
		studentID         int64
		amount            float64
		givingAmount      float64
		residualAmount    float64
		orderNumber       string
		status            int
		orderType         int
	)
	if err := tx.QueryRowContext(ctx, `
		SELECT rao.id, rao.recharge_account_id, IFNULL(rao.sale_order_id, 0), rao.student_id, IFNULL(rao.amount, 0), IFNULL(rao.giving_amount, 0), IFNULL(rao.residual_amount, 0), IFNULL(rao.order_number, ''), IFNULL(rao.status, 0), IFNULL(so.order_type, 2)
		FROM recharge_account_bill rb
		INNER JOIN recharge_account_order rao ON rao.id = rb.recharge_account_order_id AND rao.del_flag = 0
		LEFT JOIN sale_order so ON so.id = rao.sale_order_id AND so.inst_id = rao.inst_id AND so.del_flag = 0
		WHERE rb.id = ? AND rb.inst_id = ? AND rb.del_flag = 0
		LIMIT 1
	`, billID, instID).Scan(&orderID, &rechargeAccountID, &saleOrderID, &studentID, &amount, &givingAmount, &residualAmount, &orderNumber, &status, &orderType); err != nil {
		return 0, err
	}
	if status != 1 {
		return 0, errors.New("订单状态异常")
	}
	if saleOrderID <= 0 {
		return 0, errors.New("系统订单不存在")
	}
	isRefund := orderType == model.OrderTypeRechargeAccountRefund
	if !isRefund && dto.Amount <= 0 {
		return 0, errors.New("支付金额不能小于0")
	}
	if isRefund && dto.Amount < 0 {
		return 0, errors.New("支付金额不能小于0")
	}
	if math.Abs(dto.Amount-amount) > 0.000001 {
		return 0, errors.New("支付金额必须等于订单金额")
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO recharge_account_bill_flow (
			inst_id, bill_id, amount, remark, create_id, create_time, del_flag
		) VALUES (?, ?, ?, ?, ?, NOW(), 0)
	`, instID, billID, dto.Amount, strings.TrimSpace(dto.Remark), operatorID)
	if err != nil {
		return 0, err
	}
	billFlowID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE recharge_account_order
		SET status = 2, update_id = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, operatorID, orderID, instID); err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE recharge_account_bill
		SET status = 2, update_id = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, operatorID, billID, instID); err != nil {
		return 0, err
	}
	payTime := any(time.Now())
	if parsed := parseDateStart(dto.PayTime); parsed != nil {
		payTime = *parsed
	}
	payMethod := dto.PayMethod
	if payMethod == nil {
		defaultMethod := 4
		payMethod = &defaultMethod
	}
	result, err = tx.ExecContext(ctx, `
		INSERT INTO sale_order_pay_detail (
			uuid, version, inst_id, order_id, amount_id, pay_method, pay_amount, pay_time, payment_voucher,
			create_id, create_time, update_id, update_time, del_flag, remark
		) VALUES (
			UUID(), 0, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), ?, NOW(), 0, ?
		)
	`, instID, saleOrderID, dto.AmountID, *payMethod, dto.Amount, payTime, strings.TrimSpace(dto.PaymentVoucher), operatorID, operatorID, strings.TrimSpace(dto.Remark))
	if err != nil {
		return 0, err
	}
	paymentDetailID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE sale_order
		SET order_status = ?, update_id = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, model.OrderStatusCompleted, operatorID, saleOrderID, instID); err != nil {
		return 0, err
	}
	if isRefund {
		var rechBal, resBal, givingBal float64
		if err := tx.QueryRowContext(ctx, `
			SELECT IFNULL(recharge_balance, 0), IFNULL(residual_balance, 0), IFNULL(giving_balance, 0)
			FROM recharge_account
			WHERE id = ? AND inst_id = ? AND del_flag = 0
			LIMIT 1
			FOR UPDATE
		`, rechargeAccountID, instID).Scan(&rechBal, &resBal, &givingBal); err != nil {
			return 0, err
		}
		if amount > rechBal+1e-9 {
			return 0, errors.New("充值余额不足")
		}
		if residualAmount > resBal+1e-9 {
			return 0, errors.New("残联余额不足")
		}
		if givingAmount > givingBal+1e-9 {
			return 0, errors.New("赠送余额不足")
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE recharge_account
			SET recharge_balance = IFNULL(recharge_balance, 0) - ?,
				residual_balance = IFNULL(residual_balance, 0) - ?,
				giving_balance = IFNULL(giving_balance, 0) - ?,
				update_id = ?,
				update_time = NOW()
			WHERE id = ? AND inst_id = ? AND del_flag = 0
		`, amount, residualAmount, givingAmount, operatorID, rechargeAccountID, instID); err != nil {
			return 0, err
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO recharge_account_flow (
				inst_id, recharge_account_id, student_id, order_number, flow_type,
				amount, residual_amount, giving_amount, remark,
				create_id, create_time, update_id, update_time, del_flag
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), ?, NOW(), 0)
		`, instID, rechargeAccountID, studentID, orderNumber, model.RechargeAccountFlowTypeRefund, -amount, -residualAmount, -givingAmount, strings.TrimSpace(dto.Remark), operatorID, operatorID); err != nil {
			return 0, err
		}
	} else {
		if _, err := tx.ExecContext(ctx, `
			UPDATE recharge_account
			SET recharge_balance = IFNULL(recharge_balance, 0) + ?,
				residual_balance = IFNULL(residual_balance, 0) + ?,
				giving_balance = IFNULL(giving_balance, 0) + ?,
				update_id = ?,
				update_time = NOW()
			WHERE id = ? AND inst_id = ? AND del_flag = 0
		`, amount, residualAmount, givingAmount, operatorID, rechargeAccountID, instID); err != nil {
			return 0, err
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO recharge_account_flow (
				inst_id, recharge_account_id, student_id, order_number, flow_type,
				amount, residual_amount, giving_amount, remark,
				create_id, create_time, update_id, update_time, del_flag
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), ?, NOW(), 0)
		`, instID, rechargeAccountID, studentID, orderNumber, model.RechargeAccountFlowTypeRecharge, amount, residualAmount, givingAmount, strings.TrimSpace(dto.Remark), operatorID, operatorID); err != nil {
			return 0, err
		}
	}
	if err := repo.upsertOrderPaymentLedgerTx(ctx, tx, instID, paymentDetailID); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return billFlowID, nil
}

func (repo *Repository) listRechargeAccountOrderTags(ctx context.Context, instID, orderID int64) ([]model.RechargeAccountOrderTag, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT tag_id, IFNULL(tag_name, '')
		FROM recharge_account_order_tag
		WHERE inst_id = ? AND recharge_account_order_id = ? AND del_flag = 0
		ORDER BY id ASC
	`, instID, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]model.RechargeAccountOrderTag, 0, 4)
	for rows.Next() {
		var (
			item  model.RechargeAccountOrderTag
			tagID int64
		)
		if err := rows.Scan(&tagID, &item.TagName); err != nil {
			return nil, err
		}
		item.TagID = strconv.FormatInt(tagID, 10)
		items = append(items, item)
	}
	return items, rows.Err()
}

func parseInt64String(raw string) int64 {
	value, _ := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	return value
}

func mustInt64(raw string) int64 {
	value, _ := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	return value
}

func joinRechargeOrderTagIDs(tagIDs []string) string {
	values := make([]string, 0, len(tagIDs))
	for _, item := range tagIDs {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		values = append(values, item)
	}
	return strings.Join(values, ",")
}
