package repository

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

type TemplateMessageRecordCreateInput struct {
	BusinessType string
	Channel      int
	TemplateID   string
	TemplateName string
	NotifyCount  int
	SuccessCount int
	SkippedCount int
	FailedCount  int
	OperatorID   int64
	OperatorName string
}

type TemplateMessageRecordItemCreateInput struct {
	BusinessType          string
	Channel               int
	RelationID            string
	StudentID             int64
	StudentName           string
	Sex                   *int
	Avatar                string
	Phone                 string
	BizTitle              string
	BizSummary            string
	Status                int
	StatusReason          string
	RecipientCount        int
	SuccessRecipientCount int
}

type TemplateMessageRecord struct {
	ID           int64
	TemplateID   string
	TemplateName string
	Channel      int
	NotifyCount  int
	SuccessCount int
	SkippedCount int
	FailedCount  int
	OperatorID   int64
	OperatorName string
	CreateTime   *time.Time
}

type TemplateMessageRecordItem struct {
	ID                    int64
	RelationID            string
	StudentID             int64
	StudentName           string
	Sex                   *int
	Avatar                string
	Phone                 string
	BizTitle              string
	BizSummary            string
	Status                int
	StatusReason          string
	RecipientCount        int
	SuccessRecipientCount int
}

func ensureTemplateMessageRecordTables(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS template_message_record (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			inst_id BIGINT NOT NULL,
			business_type VARCHAR(64) NOT NULL DEFAULT '',
			channel INT NOT NULL DEFAULT 0,
			template_id VARCHAR(128) NOT NULL DEFAULT '',
			template_name VARCHAR(100) NOT NULL DEFAULT '',
			notify_count INT NOT NULL DEFAULT 0,
			success_count INT NOT NULL DEFAULT 0,
			skipped_count INT NOT NULL DEFAULT 0,
			failed_count INT NOT NULL DEFAULT 0,
			operator_id BIGINT NOT NULL DEFAULT 0,
			operator_name VARCHAR(100) NOT NULL DEFAULT '',
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			KEY idx_template_message_record_inst_business (inst_id, business_type, del_flag, id),
			KEY idx_template_message_record_inst_time (inst_id, business_type, create_time, id)
		)
	`); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS template_message_record_item (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			inst_id BIGINT NOT NULL,
			record_id BIGINT NOT NULL,
			business_type VARCHAR(64) NOT NULL DEFAULT '',
			channel INT NOT NULL DEFAULT 0,
			relation_id VARCHAR(64) NOT NULL DEFAULT '',
			student_id BIGINT NOT NULL DEFAULT 0,
			student_name VARCHAR(100) NOT NULL DEFAULT '',
			sex INT NULL DEFAULT NULL,
			avatar VARCHAR(500) NOT NULL DEFAULT '',
			phone VARCHAR(50) NOT NULL DEFAULT '',
			biz_title VARCHAR(100) NOT NULL DEFAULT '',
			biz_summary VARCHAR(255) NOT NULL DEFAULT '',
			status INT NOT NULL DEFAULT 0,
			status_reason VARCHAR(500) NOT NULL DEFAULT '',
			recipient_count INT NOT NULL DEFAULT 0,
			success_recipient_count INT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			KEY idx_template_message_record_item_record (inst_id, record_id, del_flag, id),
			KEY idx_template_message_record_item_relation (inst_id, business_type, relation_id, id),
			KEY idx_template_message_record_item_student (inst_id, student_id, id)
		)
	`); err != nil {
		return err
	}

	return nil
}

func (repo *Repository) CreateTemplateMessageRecordWithItems(ctx context.Context, instID int64, input TemplateMessageRecordCreateInput, items []TemplateMessageRecordItemCreateInput) (int64, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	result, err := tx.ExecContext(ctx, `
		INSERT INTO template_message_record (
			inst_id, business_type, channel, template_id, template_name,
			notify_count, success_count, skipped_count, failed_count,
			operator_id, operator_name, create_time, update_time, del_flag
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
	`,
		instID,
		strings.TrimSpace(input.BusinessType),
		input.Channel,
		strings.TrimSpace(input.TemplateID),
		strings.TrimSpace(input.TemplateName),
		input.NotifyCount,
		input.SuccessCount,
		input.SkippedCount,
		input.FailedCount,
		input.OperatorID,
		strings.TrimSpace(input.OperatorName),
	)
	if err != nil {
		return 0, err
	}

	recordID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	if len(items) > 0 {
		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO template_message_record_item (
				inst_id, record_id, business_type, channel, relation_id, student_id,
				student_name, sex, avatar, phone, biz_title, biz_summary,
				status, status_reason, recipient_count, success_recipient_count,
				create_time, update_time, del_flag
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
		`)
		if err != nil {
			return 0, err
		}
		defer stmt.Close()

		for _, item := range items {
			var sex any
			if item.Sex != nil {
				sex = *item.Sex
			}
			if _, err := stmt.ExecContext(ctx,
				instID,
				recordID,
				strings.TrimSpace(item.BusinessType),
				item.Channel,
				strings.TrimSpace(item.RelationID),
				item.StudentID,
				strings.TrimSpace(item.StudentName),
				sex,
				strings.TrimSpace(item.Avatar),
				strings.TrimSpace(item.Phone),
				strings.TrimSpace(item.BizTitle),
				strings.TrimSpace(item.BizSummary),
				item.Status,
				strings.TrimSpace(item.StatusReason),
				item.RecipientCount,
				item.SuccessRecipientCount,
			); err != nil {
				return 0, err
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return recordID, nil
}

func (repo *Repository) PageTemplateMessageRecords(ctx context.Context, instID int64, businessType string, page model.PageRequestModel) ([]TemplateMessageRecord, int, error) {
	current := page.PageIndex
	size := page.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	if size > 200 {
		size = 200
	}
	offset := (current - 1) * size
	if page.SkipCount > 0 {
		offset = page.SkipCount
	}

	var total int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM template_message_record
		WHERE inst_id = ? AND del_flag = 0 AND business_type = ?
	`, instID, strings.TrimSpace(businessType)).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			id,
			IFNULL(template_id, ''),
			IFNULL(template_name, ''),
			IFNULL(channel, 0),
			IFNULL(notify_count, 0),
			IFNULL(success_count, 0),
			IFNULL(skipped_count, 0),
			IFNULL(failed_count, 0),
			IFNULL(operator_id, 0),
			IFNULL(operator_name, ''),
			create_time
		FROM template_message_record
		WHERE inst_id = ? AND del_flag = 0 AND business_type = ?
		ORDER BY create_time DESC, id DESC
		LIMIT ? OFFSET ?
	`, instID, strings.TrimSpace(businessType), size, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]TemplateMessageRecord, 0, size)
	for rows.Next() {
		var (
			item       TemplateMessageRecord
			createTime sql.NullTime
		)
		if err := rows.Scan(
			&item.ID,
			&item.TemplateID,
			&item.TemplateName,
			&item.Channel,
			&item.NotifyCount,
			&item.SuccessCount,
			&item.SkippedCount,
			&item.FailedCount,
			&item.OperatorID,
			&item.OperatorName,
			&createTime,
		); err != nil {
			return nil, 0, err
		}
		if createTime.Valid {
			t := createTime.Time
			item.CreateTime = &t
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (repo *Repository) GetTemplateMessageRecord(ctx context.Context, instID, recordID int64, businessType string) (TemplateMessageRecord, error) {
	var (
		item       TemplateMessageRecord
		createTime sql.NullTime
	)
	if err := repo.db.QueryRowContext(ctx, `
		SELECT
			id,
			IFNULL(template_id, ''),
			IFNULL(template_name, ''),
			IFNULL(channel, 0),
			IFNULL(notify_count, 0),
			IFNULL(success_count, 0),
			IFNULL(skipped_count, 0),
			IFNULL(failed_count, 0),
			IFNULL(operator_id, 0),
			IFNULL(operator_name, ''),
			create_time
		FROM template_message_record
		WHERE inst_id = ? AND id = ? AND del_flag = 0 AND business_type = ?
		LIMIT 1
	`, instID, recordID, strings.TrimSpace(businessType)).Scan(
		&item.ID,
		&item.TemplateID,
		&item.TemplateName,
		&item.Channel,
		&item.NotifyCount,
		&item.SuccessCount,
		&item.SkippedCount,
		&item.FailedCount,
		&item.OperatorID,
		&item.OperatorName,
		&createTime,
	); err != nil {
		return TemplateMessageRecord{}, err
	}
	if createTime.Valid {
		t := createTime.Time
		item.CreateTime = &t
	}
	return item, nil
}

func (repo *Repository) ListTemplateMessageRecordItems(ctx context.Context, instID, recordID int64) ([]TemplateMessageRecordItem, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			id,
			IFNULL(relation_id, ''),
			IFNULL(student_id, 0),
			IFNULL(student_name, ''),
			sex,
			IFNULL(avatar, ''),
			IFNULL(phone, ''),
			IFNULL(biz_title, ''),
			IFNULL(biz_summary, ''),
			IFNULL(status, 0),
			IFNULL(status_reason, ''),
			IFNULL(recipient_count, 0),
			IFNULL(success_recipient_count, 0)
		FROM template_message_record_item
		WHERE inst_id = ? AND record_id = ? AND del_flag = 0
		ORDER BY id DESC
	`, instID, recordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]TemplateMessageRecordItem, 0, 16)
	for rows.Next() {
		var (
			item TemplateMessageRecordItem
			sex  sql.NullInt64
		)
		if err := rows.Scan(
			&item.ID,
			&item.RelationID,
			&item.StudentID,
			&item.StudentName,
			&sex,
			&item.Avatar,
			&item.Phone,
			&item.BizTitle,
			&item.BizSummary,
			&item.Status,
			&item.StatusReason,
			&item.RecipientCount,
			&item.SuccessRecipientCount,
		); err != nil {
			return nil, err
		}
		if sex.Valid {
			value := int(sex.Int64)
			item.Sex = &value
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func formatTemplateMessageRecordID(id int64) string {
	if id <= 0 {
		return ""
	}
	return strconv.FormatInt(id, 10)
}
