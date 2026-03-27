package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

type EnrolledStudentExportRecordEntity struct {
	ID              int64
	InstID          int64
	ExportStaffID   int64
	ExportStaffName string
	FileName        string
	ContentType     string
	FileData        []byte
	TotalRows       int
	QueryConditions []model.ExportConditionItem
	CreatedTime     *time.Time
	ExpiresAt       *time.Time
}

func ensureEnrolledStudentExportTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS enrolled_student_export_record (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			inst_id BIGINT NOT NULL DEFAULT 0,
			export_staff_id BIGINT NOT NULL DEFAULT 0,
			export_staff_name VARCHAR(100) NOT NULL DEFAULT '',
			file_name VARCHAR(255) NOT NULL DEFAULT '',
			content_type VARCHAR(120) NOT NULL DEFAULT '',
			file_data LONGBLOB NOT NULL,
			total_rows INT NOT NULL DEFAULT 0,
			query_conditions_json LONGTEXT NOT NULL,
			expires_at DATETIME NOT NULL,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			KEY idx_enrolled_student_export_inst (inst_id, create_time, id),
			KEY idx_enrolled_student_export_expire (expires_at)
		)
	`)
	return err
}

func (repo *Repository) CleanupExpiredEnrolledStudentExportRecords(ctx context.Context) error {
	_, err := repo.db.ExecContext(ctx, `
		DELETE FROM enrolled_student_export_record
		WHERE del_flag = 0 AND expires_at <= NOW()
	`)
	return err
}

func (repo *Repository) CreateEnrolledStudentExportRecord(ctx context.Context, entity EnrolledStudentExportRecordEntity) (int64, error) {
	queryConditionsRaw, err := json.Marshal(entity.QueryConditions)
	if err != nil {
		return 0, err
	}
	result, err := repo.db.ExecContext(ctx, `
		INSERT INTO enrolled_student_export_record (
			inst_id, export_staff_id, export_staff_name, file_name, content_type,
			file_data, total_rows, query_conditions_json, expires_at, create_time, update_time, del_flag
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
	`,
		entity.InstID,
		entity.ExportStaffID,
		entity.ExportStaffName,
		entity.FileName,
		entity.ContentType,
		entity.FileData,
		entity.TotalRows,
		string(queryConditionsRaw),
		entity.ExpiresAt,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (repo *Repository) ListEnrolledStudentExportRecords(ctx context.Context, instID int64) ([]model.EnrolledStudentExportRecord, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, file_name, export_staff_name, total_rows, query_conditions_json, create_time, expires_at
		FROM enrolled_student_export_record
		WHERE inst_id = ? AND del_flag = 0 AND expires_at > NOW()
		ORDER BY create_time DESC, id DESC
		LIMIT 100
	`, instID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.EnrolledStudentExportRecord, 0, 16)
	for rows.Next() {
		var (
			item               model.EnrolledStudentExportRecord
			queryConditionsRaw string
			createdTime        sql.NullTime
			expiresAt          sql.NullTime
		)
		if err := rows.Scan(
			&item.ID,
			&item.FileName,
			&item.ExporterName,
			&item.TotalRows,
			&queryConditionsRaw,
			&createdTime,
			&expiresAt,
		); err != nil {
			return nil, err
		}
		if strings.TrimSpace(queryConditionsRaw) != "" {
			if err := json.Unmarshal([]byte(queryConditionsRaw), &item.QueryConditions); err != nil {
				return nil, err
			}
		}
		if createdTime.Valid {
			t := createdTime.Time
			item.CreatedTime = &t
		}
		if expiresAt.Valid {
			t := expiresAt.Time
			item.ExpiresAt = &t
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (repo *Repository) GetEnrolledStudentExportRecord(ctx context.Context, instID, recordID int64) (EnrolledStudentExportRecordEntity, error) {
	var (
		entity             EnrolledStudentExportRecordEntity
		queryConditionsRaw string
		createdTime        sql.NullTime
		expiresAt          sql.NullTime
	)
	err := repo.db.QueryRowContext(ctx, `
		SELECT id, inst_id, export_staff_id, export_staff_name, file_name, content_type, file_data,
		       total_rows, query_conditions_json, create_time, expires_at
		FROM enrolled_student_export_record
		WHERE id = ? AND inst_id = ? AND del_flag = 0 AND expires_at > NOW()
		LIMIT 1
	`, recordID, instID).Scan(
		&entity.ID,
		&entity.InstID,
		&entity.ExportStaffID,
		&entity.ExportStaffName,
		&entity.FileName,
		&entity.ContentType,
		&entity.FileData,
		&entity.TotalRows,
		&queryConditionsRaw,
		&createdTime,
		&expiresAt,
	)
	if err != nil {
		return EnrolledStudentExportRecordEntity{}, err
	}
	if strings.TrimSpace(queryConditionsRaw) != "" {
		if err := json.Unmarshal([]byte(queryConditionsRaw), &entity.QueryConditions); err != nil {
			return EnrolledStudentExportRecordEntity{}, err
		}
	}
	if createdTime.Valid {
		t := createdTime.Time
		entity.CreatedTime = &t
	}
	if expiresAt.Valid {
		t := expiresAt.Time
		entity.ExpiresAt = &t
	}
	return entity, nil
}

func (repo *Repository) GetRechargeBalancesByStudentIDs(ctx context.Context, instID int64, studentIDs []int64) (map[int64]model.EnrolledStudentBalance, error) {
	result := make(map[int64]model.EnrolledStudentBalance, len(studentIDs))
	if len(studentIDs) == 0 {
		return result, nil
	}

	placeholders := make([]string, 0, len(studentIDs))
	args := make([]any, 0, len(studentIDs)+1)
	args = append(args, instID)
	for _, studentID := range studentIDs {
		placeholders = append(placeholders, "?")
		args = append(args, studentID)
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT ras.student_id,
		       IFNULL(SUM(IFNULL(ra.recharge_balance, 0) + IFNULL(ra.residual_balance, 0)), 0) AS available_balance,
		       IFNULL(SUM(IFNULL(ra.giving_balance, 0)), 0) AS gift_balance
		FROM recharge_account_student ras
		INNER JOIN recharge_account ra ON ra.id = ras.recharge_account_id AND ra.del_flag = 0
		WHERE ras.inst_id = ? AND ras.del_flag = 0 AND ras.student_id IN (`+strings.Join(placeholders, ",")+`)
		GROUP BY ras.student_id
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			studentID int64
			item      model.EnrolledStudentBalance
		)
		if err := rows.Scan(&studentID, &item.AvailableBalance, &item.GiftBalance); err != nil {
			return nil, err
		}
		result[studentID] = item
	}
	return result, rows.Err()
}

func (repo *Repository) GetStudentOrderArrearAmounts(ctx context.Context, instID int64, studentIDs []int64) (map[int64]float64, error) {
	result := make(map[int64]float64, len(studentIDs))
	if len(studentIDs) == 0 {
		return result, nil
	}

	placeholders := make([]string, 0, len(studentIDs))
	args := make([]any, 0, len(studentIDs)+5)
	args = append(args, instID, model.OrderStatusClosed, model.OrderStatusVoided, model.OrderStatusRefunding, model.OrderStatusRefunded)
	for _, studentID := range studentIDs {
		placeholders = append(placeholders, "?")
		args = append(args, studentID)
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT so.student_id,
		       IFNULL(SUM(
		         CASE
		           WHEN IFNULL(so.is_bad_debt, 0) = 1 THEN 0
		           ELSE GREATEST(IFNULL(so.order_real_amount, 0) - IFNULL(pay.paid_amount, 0), 0)
		         END
		       ), 0) AS arrear_amount
		FROM sale_order so
		LEFT JOIN (
			SELECT order_id, SUM(IFNULL(pay_amount, 0)) AS paid_amount
			FROM sale_order_pay_detail
			WHERE del_flag = 0
			GROUP BY order_id
		) pay ON pay.order_id = so.id
		WHERE so.inst_id = ?
		  AND so.del_flag = 0
		  AND IFNULL(so.order_status, 0) NOT IN (?, ?, ?, ?)
		  AND so.student_id IN (`+strings.Join(placeholders, ",")+`)
		GROUP BY so.student_id
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			studentID    int64
			arrearAmount float64
		)
		if err := rows.Scan(&studentID, &arrearAmount); err != nil {
			return nil, err
		}
		result[studentID] = arrearAmount
	}
	return result, rows.Err()
}

func (repo *Repository) GetStudentRawMobileMap(ctx context.Context, instID int64, studentIDs []int64) (map[int64]string, error) {
	result := make(map[int64]string, len(studentIDs))
	if len(studentIDs) == 0 {
		return result, nil
	}

	placeholders := make([]string, 0, len(studentIDs))
	args := make([]any, 0, len(studentIDs)+1)
	args = append(args, instID)
	for _, studentID := range studentIDs {
		placeholders = append(placeholders, "?")
		args = append(args, studentID)
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, IFNULL(mobile, '')
		FROM inst_student
		WHERE inst_id = ? AND del_flag = 0 AND id IN (`+strings.Join(placeholders, ",")+`)
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			studentID int64
			mobile    string
		)
		if err := rows.Scan(&studentID, &mobile); err != nil {
			return nil, err
		}
		result[studentID] = strings.TrimSpace(mobile)
	}
	return result, rows.Err()
}
