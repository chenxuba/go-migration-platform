package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

func ensureClassroomTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS inst_classroom (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			uuid VARCHAR(64) NULL,
			version BIGINT NOT NULL DEFAULT 0,
			inst_id BIGINT NOT NULL,
			name VARCHAR(100) NOT NULL,
			address VARCHAR(255) NOT NULL DEFAULT '',
			enabled TINYINT(1) NOT NULL DEFAULT 1,
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT NOT NULL DEFAULT 0,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			KEY idx_inst_classroom_inst (inst_id),
			KEY idx_inst_classroom_name (name),
			KEY idx_inst_classroom_enabled (enabled)
		)
	`)
	if err != nil {
		return err
	}
	if err := ensureColumnsOnTable(ctx, db, "inst_classroom", map[string]string{
		"address":   "address VARCHAR(255) NOT NULL DEFAULT '' AFTER name",
		"enabled":   "enabled TINYINT(1) NOT NULL DEFAULT 1 AFTER address",
		"create_id": "create_id BIGINT NOT NULL DEFAULT 0 AFTER enabled",
		"update_id": "update_id BIGINT NOT NULL DEFAULT 0 AFTER create_time",
	}); err != nil {
		return err
	}
	for _, col := range []string{"capacity", "remark", "sort"} {
		if err := dropColumnIfExists(ctx, db, "inst_classroom", col); err != nil {
			return err
		}
	}
	return nil
}

func (repo *Repository) ListClassrooms(ctx context.Context, instID int64, query model.ClassroomQueryDTO) ([]model.ClassroomVO, error) {
	filters := []string{"inst_id = ?", "del_flag = 0"}
	args := []any{instID}

	if query.EnabledOnly != nil {
		filters = append(filters, "enabled = ?")
		args = append(args, *query.EnabledOnly)
	}
	if keyword := strings.TrimSpace(query.SearchKey); keyword != "" {
		filters = append(filters, "(name LIKE ? OR address LIKE ?)")
		like := "%" + keyword + "%"
		args = append(args, like, like)
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, IFNULL(uuid, ''), IFNULL(version, 0), inst_id, IFNULL(name, ''), IFNULL(address, ''),
		       IFNULL(enabled, 0), create_time, update_time
		FROM inst_classroom
		WHERE `+strings.Join(filters, " AND ")+`
		ORDER BY enabled DESC, create_time DESC, id DESC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.ClassroomVO, 0, 32)
	for rows.Next() {
		var item model.ClassroomVO
		var createTime sql.NullTime
		var updateTime sql.NullTime
		if err := rows.Scan(
			&item.ID,
			&item.UUID,
			&item.Version,
			&item.InstID,
			&item.Name,
			&item.Address,
			&item.Enabled,
			&createTime,
			&updateTime,
		); err != nil {
			return nil, err
		}
		if createTime.Valid {
			t := createTime.Time
			item.CreateTime = &t
		}
		if updateTime.Valid {
			t := updateTime.Time
			item.UpdateTime = &t
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repo *Repository) CountClassroomsByName(ctx context.Context, instID int64, name string, excludeID *int64) (int, error) {
	query := "SELECT COUNT(*) FROM inst_classroom WHERE inst_id = ? AND name = ? AND del_flag = 0"
	args := []any{instID, strings.TrimSpace(name)}
	if excludeID != nil && *excludeID > 0 {
		query += " AND id <> ?"
		args = append(args, *excludeID)
	}
	var count int
	err := repo.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

func (repo *Repository) CreateClassroom(ctx context.Context, instID, operatorID int64, input model.ClassroomMutation) (int64, error) {
	result, err := repo.db.ExecContext(ctx, `
		INSERT INTO inst_classroom (
			inst_id, name, address, enabled,
			create_id, create_time, update_id, update_time, del_flag, version
		)
		VALUES (?, ?, ?, ?, ?, NOW(), ?, NOW(), 0, 0)
	`,
		instID,
		strings.TrimSpace(input.Name),
		strings.TrimSpace(input.Address),
		boolValueWithFallback(input.Enabled, true),
		operatorID,
		operatorID,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (repo *Repository) UpdateClassroom(ctx context.Context, instID, operatorID int64, input model.ClassroomMutation) error {
	if input.ID == nil || *input.ID <= 0 {
		return fmt.Errorf("id is required")
	}
	_, err := repo.db.ExecContext(ctx, `
		UPDATE inst_classroom
		SET name = ?, address = ?, enabled = ?, update_id = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`,
		strings.TrimSpace(input.Name),
		strings.TrimSpace(input.Address),
		boolValueWithFallback(input.Enabled, true),
		operatorID,
		*input.ID,
		instID,
	)
	return err
}

func (repo *Repository) UpdateClassroomStatus(ctx context.Context, instID, operatorID int64, input model.ClassroomStatusMutation) error {
	if input.ID == nil || *input.ID <= 0 || input.Enabled == nil {
		return fmt.Errorf("id and enabled are required")
	}
	_, err := repo.db.ExecContext(ctx, `
		UPDATE inst_classroom
		SET enabled = ?, update_id = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, *input.Enabled, operatorID, *input.ID, instID)
	return err
}

func (repo *Repository) DeleteClassroom(ctx context.Context, instID, operatorID, classroomID int64) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE inst_classroom
		SET del_flag = 1, update_id = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, operatorID, classroomID, instID)
	return err
}

func (repo *Repository) resolveClassroomByIDTx(ctx context.Context, tx *sql.Tx, instID int64, rawID string) (int64, string, any, error) {
	classroomID, err := strconv.ParseInt(strings.TrimSpace(rawID), 10, 64)
	if strings.TrimSpace(rawID) == "" || strings.TrimSpace(rawID) == "0" {
		return 0, "", nil, nil
	}
	if err != nil || classroomID <= 0 {
		return 0, "", nil, fmt.Errorf("classroomId 无效")
	}
	var (
		name    string
		enabled bool
	)
	if err := tx.QueryRowContext(ctx, `
		SELECT IFNULL(name, ''), IFNULL(enabled, 0)
		FROM inst_classroom
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, classroomID, instID).Scan(&name, &enabled); err != nil {
		if err == sql.ErrNoRows {
			return 0, "", nil, fmt.Errorf("教室不存在")
		}
		return 0, "", nil, err
	}
	return classroomID, strings.TrimSpace(name), enabled, nil
}

func boolValueWithFallback(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}
