package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

func ensureNoticeTemplateTables(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS notice_template (
			id BIGINT PRIMARY KEY,
			title VARCHAR(100) NOT NULL DEFAULT '',
			cover_url VARCHAR(512) NOT NULL DEFAULT '',
			tag VARCHAR(100) NOT NULL DEFAULT '',
			weight INT NOT NULL DEFAULT 0,
			content LONGTEXT NULL,
			summary TEXT NULL,
			org_id BIGINT NOT NULL DEFAULT 0,
			school_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			KEY idx_notice_template_scope (school_id, org_id, weight, id),
			KEY idx_notice_template_deleted (del_flag, weight, id)
		)
	`); err != nil {
		return err
	}

	for _, columns := range []map[string]string{
		{"cover_url": "cover_url VARCHAR(512) NOT NULL DEFAULT '' AFTER title"},
		{"tag": "tag VARCHAR(100) NOT NULL DEFAULT '' AFTER cover_url"},
		{"weight": "weight INT NOT NULL DEFAULT 0 AFTER tag"},
		{"content": "content LONGTEXT NULL AFTER weight"},
		{"summary": "summary TEXT NULL AFTER content"},
		{"org_id": "org_id BIGINT NOT NULL DEFAULT 0 AFTER summary"},
		{"school_id": "school_id BIGINT NOT NULL DEFAULT 0 AFTER org_id"},
	} {
		if err := ensureColumnsOnTable(ctx, db, "notice_template", columns); err != nil {
			return err
		}
	}

	return seedNoticeTemplates(ctx, db)
}

func seedNoticeTemplates(ctx context.Context, db *sql.DB) error {
	var items []model.NoticeTemplateVO
	if err := json.Unmarshal([]byte(noticeTemplateSeedJSON), &items); err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO notice_template (
			id, title, cover_url, tag, weight, content, summary, org_id, school_id, del_flag
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
		ON DUPLICATE KEY UPDATE
			title = VALUES(title),
			cover_url = VALUES(cover_url),
			tag = VALUES(tag),
			weight = VALUES(weight),
			content = VALUES(content),
			summary = VALUES(summary),
			org_id = VALUES(org_id),
			school_id = VALUES(school_id),
			del_flag = 0
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, item := range items {
		id, parseErr := strconv.ParseInt(strings.TrimSpace(item.ID), 10, 64)
		if parseErr != nil {
			err = parseErr
			return err
		}
		orgID, parseErr := parseNoticeTemplateScopeID(item.OrgID)
		if parseErr != nil {
			err = parseErr
			return err
		}
		schoolID, parseErr := parseNoticeTemplateScopeID(item.SchoolID)
		if parseErr != nil {
			err = parseErr
			return err
		}
		if _, err = stmt.ExecContext(
			ctx,
			id,
			strings.TrimSpace(item.Title),
			strings.TrimSpace(item.CoverURL),
			strings.TrimSpace(item.Tag),
			item.Weight,
			item.Content,
			item.Summary,
			orgID,
			schoolID,
		); err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func parseNoticeTemplateScopeID(value string) (int64, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return 0, nil
	}
	return strconv.ParseInt(normalized, 10, 64)
}

func (repo *Repository) ListNoticeTemplates(ctx context.Context, schoolID int64) ([]model.NoticeTemplateVO, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			id,
			IFNULL(title, ''),
			IFNULL(cover_url, ''),
			IFNULL(tag, ''),
			IFNULL(weight, 0),
			IFNULL(content, ''),
			IFNULL(summary, ''),
			IFNULL(org_id, 0),
			IFNULL(school_id, 0)
		FROM notice_template
		WHERE del_flag = 0
		  AND (school_id = 0 OR school_id = ?)
		ORDER BY weight DESC, id DESC
	`, schoolID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]model.NoticeTemplateVO, 0)
	for rows.Next() {
		var (
			id      int64
			orgID   int64
			scopeID int64
			item    model.NoticeTemplateVO
		)
		if err := rows.Scan(
			&id,
			&item.Title,
			&item.CoverURL,
			&item.Tag,
			&item.Weight,
			&item.Content,
			&item.Summary,
			&orgID,
			&scopeID,
		); err != nil {
			return nil, err
		}
		item.ID = strconv.FormatInt(id, 10)
		item.OrgID = strconv.FormatInt(orgID, 10)
		item.SchoolID = strconv.FormatInt(scopeID, 10)
		result = append(result, item)
	}

	return result, rows.Err()
}
