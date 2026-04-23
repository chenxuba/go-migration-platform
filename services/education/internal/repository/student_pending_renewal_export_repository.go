package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

type PendingRenewalStudentExportRecordEntity struct {
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

func ensurePendingRenewalStudentExportTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS pending_renewal_student_export_record (
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
			KEY idx_pending_renewal_student_export_inst (inst_id, create_time, id),
			KEY idx_pending_renewal_student_export_expire (expires_at)
		)
	`)
	return err
}

func (repo *Repository) CleanupExpiredPendingRenewalStudentExportRecords(ctx context.Context) error {
	_, err := repo.db.ExecContext(ctx, `
		DELETE FROM pending_renewal_student_export_record
		WHERE del_flag = 0 AND expires_at <= NOW()
	`)
	return err
}

func (repo *Repository) CreatePendingRenewalStudentExportRecord(ctx context.Context, entity PendingRenewalStudentExportRecordEntity) (int64, error) {
	queryConditionsRaw, err := json.Marshal(entity.QueryConditions)
	if err != nil {
		return 0, err
	}
	result, err := repo.db.ExecContext(ctx, `
		INSERT INTO pending_renewal_student_export_record (
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

func (repo *Repository) ListPendingRenewalStudentExportRecords(ctx context.Context, instID int64) ([]model.PendingRenewalStudentExportRecord, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, file_name, export_staff_name, total_rows, query_conditions_json, create_time, expires_at
		FROM pending_renewal_student_export_record
		WHERE inst_id = ? AND del_flag = 0 AND expires_at > NOW()
		ORDER BY create_time DESC, id DESC
		LIMIT 100
	`, instID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.PendingRenewalStudentExportRecord, 0, 16)
	for rows.Next() {
		var (
			item               model.PendingRenewalStudentExportRecord
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

func (repo *Repository) GetPendingRenewalStudentExportRecord(ctx context.Context, instID, recordID int64) (PendingRenewalStudentExportRecordEntity, error) {
	var (
		entity             PendingRenewalStudentExportRecordEntity
		queryConditionsRaw string
		createdTime        sql.NullTime
		expiresAt          sql.NullTime
	)
	err := repo.db.QueryRowContext(ctx, `
		SELECT id, inst_id, export_staff_id, export_staff_name, file_name, content_type, file_data,
		       total_rows, query_conditions_json, create_time, expires_at
		FROM pending_renewal_student_export_record
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
		return PendingRenewalStudentExportRecordEntity{}, err
	}
	if strings.TrimSpace(queryConditionsRaw) != "" {
		if err := json.Unmarshal([]byte(queryConditionsRaw), &entity.QueryConditions); err != nil {
			return PendingRenewalStudentExportRecordEntity{}, err
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
