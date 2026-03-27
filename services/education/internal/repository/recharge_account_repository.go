package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

func ensureRechargeAccountTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS recharge_account (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			uuid VARCHAR(64) NULL,
			version BIGINT NOT NULL DEFAULT 0,
			inst_id BIGINT NOT NULL,
			account_name VARCHAR(100) NOT NULL DEFAULT '',
			main_student_id BIGINT NOT NULL DEFAULT 0,
			phone VARCHAR(32) NOT NULL DEFAULT '',
			recharge_balance DECIMAL(18,2) NOT NULL DEFAULT 0,
			residual_balance DECIMAL(18,2) NOT NULL DEFAULT 0,
			giving_balance DECIMAL(18,2) NOT NULL DEFAULT 0,
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT NOT NULL DEFAULT 0,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			KEY idx_recharge_account_inst (inst_id, update_time, id),
			KEY idx_recharge_account_main_student (inst_id, main_student_id)
		)
	`)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS recharge_account_student (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			uuid VARCHAR(64) NULL,
			version BIGINT NOT NULL DEFAULT 0,
			inst_id BIGINT NOT NULL,
			recharge_account_id BIGINT NOT NULL,
			student_id BIGINT NOT NULL,
			is_main_student TINYINT(1) NOT NULL DEFAULT 0,
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT NOT NULL DEFAULT 0,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uk_recharge_account_student (recharge_account_id, student_id),
			KEY idx_recharge_account_student_inst (inst_id, student_id),
			KEY idx_recharge_account_student_account (inst_id, recharge_account_id)
		)
	`)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS recharge_account_flow (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			uuid VARCHAR(64) NULL,
			version BIGINT NOT NULL DEFAULT 0,
			inst_id BIGINT NOT NULL,
			recharge_account_id BIGINT NOT NULL,
			student_id BIGINT NOT NULL DEFAULT 0,
			order_number VARCHAR(64) NOT NULL DEFAULT '',
			flow_type INT NOT NULL DEFAULT 1,
			amount DECIMAL(18,2) NOT NULL DEFAULT 0,
			residual_amount DECIMAL(18,2) NOT NULL DEFAULT 0,
			giving_amount DECIMAL(18,2) NOT NULL DEFAULT 0,
			remark VARCHAR(500) NOT NULL DEFAULT '',
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT NOT NULL DEFAULT 0,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			KEY idx_recharge_account_flow_account (inst_id, recharge_account_id, create_time),
			KEY idx_recharge_account_flow_student (inst_id, student_id)
		)
	`)
	if err != nil {
		return err
	}
	return ensureRechargeAccountOrderTables(ctx, db)
}

func (repo *Repository) PageRechargeAccountItems(ctx context.Context, instID int64, query model.RechargeAccountItemPageQueryDTO) (model.RechargeAccountItemPageResult, error) {
	current := query.PageRequestModel.PageIndex
	size := query.PageRequestModel.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	whereParts := []string{"ra.inst_id = ?", "ra.del_flag = 0"}
	args := []any{instID}
	if strings.TrimSpace(query.QueryModel.StudentID) != "" {
		whereParts = append(whereParts, `EXISTS (
			SELECT 1 FROM recharge_account_student ras
			WHERE ras.recharge_account_id = ra.id AND ras.del_flag = 0 AND CAST(ras.student_id AS CHAR) = ?
		)`)
		args = append(args, strings.TrimSpace(query.QueryModel.StudentID))
	}
	showZero := false
	if query.QueryModel.ShowZeroBalanceAccount != nil {
		showZero = *query.QueryModel.ShowZeroBalanceAccount
	}
	if !showZero {
		whereParts = append(whereParts, "(IFNULL(ra.recharge_balance, 0) + IFNULL(ra.residual_balance, 0) + IFNULL(ra.giving_balance, 0)) > 0")
	}
	whereSQL := strings.Join(whereParts, " AND ")
	orderBy := "ra.update_time DESC, ra.id DESC"
	if query.SortModel.OrderByUpdatedTime > 0 {
		orderBy = "ra.update_time ASC, ra.id ASC"
	}

	var total int
	if err := repo.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM recharge_account ra WHERE `+whereSQL, args...).Scan(&total); err != nil {
		return model.RechargeAccountItemPageResult{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			ra.id,
			IFNULL(ra.account_name, ''),
			IFNULL(ra.phone, ''),
			ra.main_student_id,
			ra.update_time,
			(IFNULL(ra.recharge_balance, 0) + IFNULL(ra.residual_balance, 0) + IFNULL(ra.giving_balance, 0)) AS balance_total,
			IFNULL(ra.recharge_balance, 0),
			IFNULL(ra.residual_balance, 0),
			IFNULL(ra.giving_balance, 0)
		FROM recharge_account ra
		WHERE `+whereSQL+`
		ORDER BY `+orderBy+`
		LIMIT ? OFFSET ?
	`, append(args, size, offset)...)
	if err != nil {
		return model.RechargeAccountItemPageResult{}, err
	}
	defer rows.Close()

	items := make([]model.RechargeAccountItem, 0, size)
	accountIDs := make([]int64, 0, size)
	for rows.Next() {
		var (
			item          model.RechargeAccountItem
			accountID     int64
			mainStudentID sql.NullInt64
			updateTime    sql.NullTime
		)
		if err := rows.Scan(&accountID, &item.RechargeAccountName, &item.Phone, &mainStudentID, &updateTime, &item.BalanceTotal, &item.RechargeBalance, &item.ResidualBalance, &item.GivingBalance); err != nil {
			return model.RechargeAccountItemPageResult{}, err
		}
		item.RechargeAccountID = strconv.FormatInt(accountID, 10)
		if mainStudentID.Valid && mainStudentID.Int64 > 0 {
			item.MainStudentID = strconv.FormatInt(mainStudentID.Int64, 10)
		}
		if updateTime.Valid {
			t := updateTime.Time
			item.UpdateTime = &t
		}
		item.RechargeAccountName = normalizeRechargeAccountName(item.RechargeAccountName, item.MainStudentID, item.RechargeAccountID)
		item.Phone = maskRechargePhone(item.Phone)
		items = append(items, item)
		accountIDs = append(accountIDs, accountID)
	}
	if err := rows.Err(); err != nil {
		return model.RechargeAccountItemPageResult{}, err
	}

	studentsMap, err := repo.listRechargeAccountStudents(ctx, instID, accountIDs)
	if err != nil {
		return model.RechargeAccountItemPageResult{}, err
	}
	for i := range items {
		accountID, _ := strconv.ParseInt(items[i].RechargeAccountID, 10, 64)
		items[i].RechargeAccountStudents = studentsMap[accountID]
		if strings.TrimSpace(items[i].Phone) == "" {
			for _, stu := range items[i].RechargeAccountStudents {
				if stu.IsMainStudent {
					phone, _ := repo.getStudentRawPhoneByID(ctx, accountID, instID, stu.StudentID)
					items[i].Phone = maskRechargePhone(phone)
					break
				}
			}
		}
		if strings.TrimSpace(items[i].RechargeAccountName) == "" {
			items[i].RechargeAccountName = normalizeRechargeAccountName("", items[i].MainStudentID, items[i].RechargeAccountID)
		}
	}

	return model.RechargeAccountItemPageResult{
		List:  items,
		Total: total,
	}, nil
}

func (repo *Repository) GetRechargeAccountStatistics(ctx context.Context, instID int64) (model.RechargeAccountStatistics, error) {
	var result model.RechargeAccountStatistics
	err := repo.db.QueryRowContext(ctx, `
		SELECT
			IFNULL(SUM(IFNULL(recharge_balance, 0) + IFNULL(residual_balance, 0) + IFNULL(giving_balance, 0)), 0),
			IFNULL(SUM(IFNULL(recharge_balance, 0) + IFNULL(residual_balance, 0)), 0),
			IFNULL(SUM(IFNULL(giving_balance, 0)), 0),
			IFNULL(SUM(IFNULL(residual_balance, 0)), 0)
		FROM recharge_account
		WHERE inst_id = ? AND del_flag = 0
	`, instID).Scan(&result.RechargeAccountTotal, &result.AmountTotal, &result.GivingAmountTotal, &result.ResidualAmountTotal)
	return result, err
}

func (repo *Repository) UpdateRechargeAccount(ctx context.Context, instID, operatorID int64, dto model.UpdateRechargeAccountDTO) error {
	accountID, err := strconv.ParseInt(strings.TrimSpace(dto.RechargeAccountID), 10, 64)
	if err != nil || accountID <= 0 {
		return fmt.Errorf("rechargeAccountId无效")
	}
	accountName := strings.TrimSpace(dto.RechargeAccountName)
	if accountName == "" {
		return fmt.Errorf("rechargeAccountName不能为空")
	}

	result, err := repo.db.ExecContext(ctx, `
		UPDATE recharge_account
		SET account_name = ?, update_id = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, accountName, operatorID, accountID, instID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("储值账户不存在")
	}
	return nil
}

func (repo *Repository) PageRechargeAccountDetails(ctx context.Context, instID int64, query model.RechargeAccountDetailQueryDTO) (model.RechargeAccountDetailPageResult, error) {
	current := query.PageRequestModel.PageIndex
	size := query.PageRequestModel.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	whereParts := []string{"raf.inst_id = ?", "raf.del_flag = 0"}
	args := []any{instID}
	if strings.TrimSpace(query.QueryModel.StudentID) != "" {
		whereParts = append(whereParts, "CAST(raf.student_id AS CHAR) = ?")
		args = append(args, strings.TrimSpace(query.QueryModel.StudentID))
	}
	if strings.TrimSpace(query.QueryModel.RechargeAccountID) != "" {
		whereParts = append(whereParts, "CAST(raf.recharge_account_id AS CHAR) = ?")
		args = append(args, strings.TrimSpace(query.QueryModel.RechargeAccountID))
	}
	if len(query.QueryModel.FlowTypes) > 0 {
		holders := make([]string, 0, len(query.QueryModel.FlowTypes))
		for _, item := range query.QueryModel.FlowTypes {
			holders = append(holders, "?")
			args = append(args, item)
		}
		whereParts = append(whereParts, "raf.flow_type IN ("+strings.Join(holders, ",")+")")
	}
	if begin := parseDateStart(query.QueryModel.StartTime); begin != nil {
		whereParts = append(whereParts, "raf.create_time >= ?")
		args = append(args, *begin)
	}
	if end := parseDateEnd(query.QueryModel.EndTime); end != nil {
		whereParts = append(whereParts, "raf.create_time <= ?")
		args = append(args, *end)
	}
	whereSQL := strings.Join(whereParts, " AND ")
	orderBy := "raf.create_time DESC, raf.id DESC"
	if query.SortModel.OrderByCreatedTime > 0 {
		orderBy = "raf.create_time ASC, raf.id ASC"
	}

	var total int
	if err := repo.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM recharge_account_flow raf WHERE `+whereSQL, args...).Scan(&total); err != nil {
		return model.RechargeAccountDetailPageResult{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			IFNULL(ra.phone, ''),
			IFNULL(raf.amount, 0),
			IFNULL(raf.giving_amount, 0),
			IFNULL(raf.residual_amount, 0),
			raf.recharge_account_id,
			raf.id,
			IFNULL(ra.account_name, ''),
			IFNULL(raf.remark, ''),
			raf.create_time,
			raf.flow_type,
			DATE_FORMAT(raf.create_time, '%Y-%m-%dT00:00:00'),
			IFNULL(so.id, 0),
			IFNULL(raf.order_number, ''),
			IFNULL(so.order_type, 0),
			raf.student_id,
			IFNULL(s.stu_name, ''),
			CASE
				WHEN CHAR_LENGTH(IFNULL(s.mobile, '')) >= 7 THEN CONCAT(LEFT(s.mobile, 3), '****', RIGHT(s.mobile, 4))
				ELSE IFNULL(s.mobile, '')
			END,
			IFNULL(s.avatar_url, ''),
			IFNULL(raf.amount, 0) + IFNULL(raf.giving_amount, 0) + IFNULL(raf.residual_amount, 0)
		FROM recharge_account_flow raf
		LEFT JOIN recharge_account ra ON ra.id = raf.recharge_account_id AND ra.del_flag = 0
		LEFT JOIN inst_student s ON s.id = raf.student_id AND s.del_flag = 0
		LEFT JOIN sale_order so ON so.order_number = raf.order_number AND so.inst_id = raf.inst_id AND so.del_flag = 0
		WHERE `+whereSQL+`
		ORDER BY `+orderBy+`
		LIMIT ? OFFSET ?
	`, append(args, size, offset)...)
	if err != nil {
		return model.RechargeAccountDetailPageResult{}, err
	}
	defer rows.Close()

	items := make([]model.RechargeAccountDetailItem, 0, size)
	accountIDs := make([]int64, 0, size)
	for rows.Next() {
		var (
			item                  model.RechargeAccountDetailItem
			rechargeAccountID     int64
			rechargeAccountFlowID int64
			studentID             int64
			createTime            sql.NullTime
			sourceID              int64
			sourceOrderType       int
		)
		if err := rows.Scan(
			&item.Phone,
			&item.Amount,
			&item.GivingAmount,
			&item.ResidualAmount,
			&rechargeAccountID,
			&rechargeAccountFlowID,
			&item.RechargeAccountName,
			&item.Remark,
			&createTime,
			&item.RechargeAccountFlowSourceType,
			&item.DealDate,
			&sourceID,
			&item.SourceOrderNumber,
			&sourceOrderType,
			&studentID,
			&item.StudentName,
			&item.StudentPhone,
			&item.StudentAvatar,
			&item.TotalAmount,
		); err != nil {
			return model.RechargeAccountDetailPageResult{}, err
		}
		item.RechargeAccountID = strconv.FormatInt(rechargeAccountID, 10)
		item.RechargeAccountFlowID = strconv.FormatInt(rechargeAccountFlowID, 10)
		item.StudentID = strconv.FormatInt(studentID, 10)
		item.SourceID = strconv.FormatInt(sourceID, 10)
		item.SourceOrderType = sourceOrderType
		if createTime.Valid {
			item.CreateTime = createTime.Time.Format(time.RFC3339)
		}
		item.RechargeAccountName = normalizeRechargeAccountName(item.RechargeAccountName, item.StudentID, item.RechargeAccountID)
		item.Phone = maskRechargePhone(item.Phone)
		items = append(items, item)
		accountIDs = append(accountIDs, rechargeAccountID)
	}
	if err := rows.Err(); err != nil {
		return model.RechargeAccountDetailPageResult{}, err
	}
	studentsMap, err := repo.listRechargeAccountStudents(ctx, instID, accountIDs)
	if err != nil {
		return model.RechargeAccountDetailPageResult{}, err
	}
	for i := range items {
		accountID, _ := strconv.ParseInt(items[i].RechargeAccountID, 10, 64)
		items[i].RechargeAccountStudents = studentsMap[accountID]
	}
	return model.RechargeAccountDetailPageResult{List: items, Total: total}, nil
}

func (repo *Repository) GetRechargeAccountExpendIncome(ctx context.Context, instID int64, query model.RechargeAccountDetailQuery) (model.RechargeAccountExpendIncome, error) {
	whereParts := []string{"inst_id = ?", "del_flag = 0"}
	args := []any{instID}
	if strings.TrimSpace(query.StudentID) != "" {
		whereParts = append(whereParts, "CAST(student_id AS CHAR) = ?")
		args = append(args, strings.TrimSpace(query.StudentID))
	}
	if begin := parseDateStart(query.StartTime); begin != nil {
		whereParts = append(whereParts, "create_time >= ?")
		args = append(args, *begin)
	}
	if end := parseDateEnd(query.EndTime); end != nil {
		whereParts = append(whereParts, "create_time <= ?")
		args = append(args, *end)
	}
	whereSQL := strings.Join(whereParts, " AND ")

	var result model.RechargeAccountExpendIncome
	err := repo.db.QueryRowContext(ctx, `
		SELECT
			IFNULL(SUM(CASE
				WHEN (IFNULL(amount, 0) + IFNULL(giving_amount, 0) + IFNULL(residual_amount, 0)) < 0
				THEN ABS(IFNULL(amount, 0) + IFNULL(giving_amount, 0) + IFNULL(residual_amount, 0))
				ELSE 0
			END), 0),
			IFNULL(SUM(CASE
				WHEN (IFNULL(amount, 0) + IFNULL(giving_amount, 0) + IFNULL(residual_amount, 0)) > 0
				THEN IFNULL(amount, 0) + IFNULL(giving_amount, 0) + IFNULL(residual_amount, 0)
				ELSE 0
			END), 0)
		FROM recharge_account_flow
		WHERE `+whereSQL, args...).Scan(&result.Expend, &result.Income)
	return result, err
}

func (repo *Repository) GetStudentRechargeAccountBalance(ctx context.Context, instID, studentID int64) (total, recharge, residual, giving float64, err error) {
	err = repo.db.QueryRowContext(ctx, `
		SELECT
			IFNULL(SUM(IFNULL(ra.recharge_balance, 0) + IFNULL(ra.residual_balance, 0) + IFNULL(ra.giving_balance, 0)), 0),
			IFNULL(SUM(IFNULL(ra.recharge_balance, 0)), 0),
			IFNULL(SUM(IFNULL(ra.residual_balance, 0)), 0),
			IFNULL(SUM(IFNULL(ra.giving_balance, 0)), 0)
		FROM recharge_account ra
		INNER JOIN recharge_account_student ras ON ras.recharge_account_id = ra.id AND ras.del_flag = 0
		WHERE ra.inst_id = ? AND ra.del_flag = 0 AND ras.student_id = ?
	`, instID, studentID).Scan(&total, &recharge, &residual, &giving)
	return
}

func (repo *Repository) EnsureRechargeAccount(ctx context.Context, instID, studentID, operatorID int64) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := repo.ensureRechargeAccountTx(ctx, tx, instID, studentID, operatorID); err != nil {
		return err
	}
	return tx.Commit()
}

func (repo *Repository) ensureRechargeAccountTx(ctx context.Context, tx *sql.Tx, instID, studentID, operatorID int64) error {
	if studentID <= 0 {
		return nil
	}

	var exists int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM recharge_account_student
		WHERE inst_id = ? AND student_id = ? AND del_flag = 0
	`, instID, studentID).Scan(&exists); err != nil {
		return err
	}
	if exists > 0 {
		return nil
	}

	var (
		studentName string
		phone       string
	)
	if err := tx.QueryRowContext(ctx, `
		SELECT IFNULL(stu_name, ''), IFNULL(mobile, '')
		FROM inst_student
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, studentID, instID).Scan(&studentName, &phone); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO recharge_account (
			uuid, version, inst_id, account_name, main_student_id, phone,
			recharge_balance, residual_balance, giving_balance,
			create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, ?, ?, ?,
			0, 0, 0,
			?, NOW(), ?, NOW(), 0
		)
	`, instID, "", studentID, strings.TrimSpace(phone), operatorID, operatorID)
	if err != nil {
		return err
	}
	rechargeAccountID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	accountName := buildRechargeAccountName(studentID, rechargeAccountID)
	if _, err := tx.ExecContext(ctx, `
		UPDATE recharge_account
		SET account_name = ?, update_id = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, accountName, operatorID, rechargeAccountID, instID); err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO recharge_account_student (
			uuid, version, inst_id, recharge_account_id, student_id, is_main_student,
			create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, ?, ?, 1,
			?, NOW(), ?, NOW(), 0
		)
	`, instID, rechargeAccountID, studentID, operatorID, operatorID)
	return err
}

func (repo *Repository) listRechargeAccountStudents(ctx context.Context, instID int64, accountIDs []int64) (map[int64][]model.RechargeAccountStudentItem, error) {
	result := make(map[int64][]model.RechargeAccountStudentItem)
	if len(accountIDs) == 0 {
		return result, nil
	}
	holders := make([]string, 0, len(accountIDs))
	args := make([]any, 0, len(accountIDs)+1)
	args = append(args, instID)
	for _, id := range accountIDs {
		holders = append(holders, "?")
		args = append(args, id)
	}
	rows, err := repo.db.QueryContext(ctx, `
		SELECT ras.recharge_account_id, ras.is_main_student, ras.student_id, IFNULL(s.stu_name, '')
		FROM recharge_account_student ras
		LEFT JOIN inst_student s ON s.id = ras.student_id AND s.del_flag = 0
		WHERE ras.inst_id = ? AND ras.del_flag = 0 AND ras.recharge_account_id IN (`+strings.Join(holders, ",")+`)
		ORDER BY ras.is_main_student DESC, ras.id ASC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			accountID int64
			item      model.RechargeAccountStudentItem
			studentID int64
		)
		if err := rows.Scan(&accountID, &item.IsMainStudent, &studentID, &item.StudentName); err != nil {
			return nil, err
		}
		item.StudentID = strconv.FormatInt(studentID, 10)
		result[accountID] = append(result[accountID], item)
	}
	return result, rows.Err()
}

func (repo *Repository) getStudentRawPhoneByID(ctx context.Context, rechargeAccountID, instID int64, studentID string) (string, error) {
	_ = rechargeAccountID
	if strings.TrimSpace(studentID) == "" {
		return "", nil
	}
	var phone string
	err := repo.db.QueryRowContext(ctx, `
		SELECT IFNULL(mobile, '')
		FROM inst_student
		WHERE inst_id = ? AND del_flag = 0 AND CAST(id AS CHAR) = ?
		LIMIT 1
	`, instID, strings.TrimSpace(studentID)).Scan(&phone)
	return phone, err
}

func normalizeRechargeAccountName(currentName, mainStudentID, rechargeAccountID string) string {
	currentName = strings.TrimSpace(currentName)
	if currentName != "" {
		return currentName
	}
	return buildRechargeAccountNameByString(mainStudentID, rechargeAccountID)
}

func maskRechargePhone(phone string) string {
	phone = strings.TrimSpace(phone)
	if len(phone) < 7 {
		return phone
	}
	if len(phone) == 11 {
		return phone[:3] + "****" + phone[len(phone)-4:]
	}
	return phone
}

func buildRechargeAccountName(studentID, rechargeAccountID int64) string {
	return fmt.Sprintf("RA-%d-%d", studentID, rechargeAccountID)
}

func buildRechargeAccountNameByString(studentID, rechargeAccountID string) string {
	studentID = strings.TrimSpace(studentID)
	rechargeAccountID = strings.TrimSpace(rechargeAccountID)
	if studentID == "" {
		studentID = "0"
	}
	if rechargeAccountID == "" {
		rechargeAccountID = "0"
	}
	return "RA-" + studentID + "-" + rechargeAccountID
}

func (repo *Repository) FindRechargeAccountImportTargetByName(ctx context.Context, instID int64, accountName string) (int64, int64, error) {
	accountName = strings.TrimSpace(accountName)
	var (
		accountID     int64
		mainStudentID int64
	)
	err := repo.db.QueryRowContext(ctx, `
		SELECT id, main_student_id
		FROM recharge_account
		WHERE inst_id = ? AND del_flag = 0
		  AND (
			account_name = ?
			OR CONCAT('RA-', CAST(main_student_id AS CHAR), '-', CAST(id AS CHAR)) = ?
		  )
		ORDER BY id ASC
		LIMIT 1
	`, instID, accountName, accountName).Scan(&accountID, &mainStudentID)
	if err != nil {
		return 0, 0, err
	}
	return accountID, mainStudentID, nil
}
