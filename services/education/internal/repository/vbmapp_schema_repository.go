package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type VBMAPPResponseSchemaPreset struct {
	ModuleCode string
	ItemCode   string
	SchemaJSON json.RawMessage
}

type VBMAPPResponseSchemaOverride struct {
	ModuleCode   string
	ItemCode     string
	OverrideJSON json.RawMessage
}

func ensureVBMAPPResponseSchemaTables(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS vbmapp_response_schema_preset (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			scale_version VARCHAR(64) NOT NULL DEFAULT '',
			module_code VARCHAR(32) NOT NULL DEFAULT '',
			item_code VARCHAR(100) NOT NULL DEFAULT '',
			schema_json LONGTEXT NOT NULL,
			sort_no INT NOT NULL DEFAULT 0,
			status VARCHAR(16) NOT NULL DEFAULT 'active',
			create_id BIGINT NOT NULL DEFAULT 0,
			update_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uk_vbmapp_response_schema_preset (scale_version, module_code, item_code),
			KEY idx_vbmapp_response_schema_preset_scope (scale_version, module_code, status, del_flag)
		)
	`); err != nil {
		return err
	}
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS vbmapp_response_schema_override (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			inst_id BIGINT NOT NULL DEFAULT 0,
			scale_version VARCHAR(64) NOT NULL DEFAULT '',
			module_code VARCHAR(32) NOT NULL DEFAULT '',
			item_code VARCHAR(100) NOT NULL DEFAULT '',
			override_json LONGTEXT NOT NULL,
			status VARCHAR(16) NOT NULL DEFAULT 'active',
			create_id BIGINT NOT NULL DEFAULT 0,
			update_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uk_vbmapp_response_schema_override (inst_id, scale_version, module_code, item_code),
			KEY idx_vbmapp_response_schema_override_scope (inst_id, scale_version, module_code, status, del_flag)
		)
	`)
	return err
}

func (repo *Repository) HasVBMAPPResponseSchemaPresets(ctx context.Context, scaleVersion string) (bool, error) {
	var count int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(1)
		FROM vbmapp_response_schema_preset
		WHERE scale_version = ? AND status = 'active' AND del_flag = 0
	`, strings.TrimSpace(scaleVersion)).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (repo *Repository) ReplaceVBMAPPResponseSchemaPresets(
	ctx context.Context,
	scaleVersion string,
	presets []VBMAPPResponseSchemaPreset,
	operatorID int64,
) error {
	scaleVersion = strings.TrimSpace(scaleVersion)
	if scaleVersion == "" {
		return fmt.Errorf("scale version is required")
	}
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if _, err = tx.ExecContext(ctx, `
		UPDATE vbmapp_response_schema_preset
		SET del_flag = 1, update_id = ?, update_time = NOW()
		WHERE scale_version = ? AND del_flag = 0
	`, operatorID, scaleVersion); err != nil {
		return err
	}
	rows := append([]VBMAPPResponseSchemaPreset(nil), presets...)
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].ModuleCode == rows[j].ModuleCode {
			return rows[i].ItemCode < rows[j].ItemCode
		}
		return rows[i].ModuleCode < rows[j].ModuleCode
	})
	for index, preset := range rows {
		moduleCode := strings.TrimSpace(preset.ModuleCode)
		itemCode := strings.TrimSpace(preset.ItemCode)
		if moduleCode == "" || itemCode == "" || len(preset.SchemaJSON) == 0 {
			continue
		}
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO vbmapp_response_schema_preset (
				scale_version, module_code, item_code, schema_json, sort_no, status,
				create_id, update_id, create_time, update_time, del_flag
			) VALUES (?, ?, ?, ?, ?, 'active', ?, ?, NOW(), NOW(), 0)
			ON DUPLICATE KEY UPDATE
				schema_json = VALUES(schema_json),
				sort_no = VALUES(sort_no),
				status = VALUES(status),
				update_id = VALUES(update_id),
				update_time = NOW(),
				del_flag = 0
		`, scaleVersion, moduleCode, itemCode, string(preset.SchemaJSON), index+1, operatorID, operatorID); err != nil {
			return err
		}
	}
	err = tx.Commit()
	return err
}

func (repo *Repository) ListVBMAPPResponseSchemaPresets(
	ctx context.Context,
	scaleVersion string,
) ([]VBMAPPResponseSchemaPreset, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT module_code, item_code, schema_json
		FROM vbmapp_response_schema_preset
		WHERE scale_version = ? AND status = 'active' AND del_flag = 0
		ORDER BY sort_no, id
	`, strings.TrimSpace(scaleVersion))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]VBMAPPResponseSchemaPreset, 0)
	for rows.Next() {
		var row VBMAPPResponseSchemaPreset
		var raw string
		if err := rows.Scan(&row.ModuleCode, &row.ItemCode, &raw); err != nil {
			return nil, err
		}
		row.ModuleCode = strings.TrimSpace(row.ModuleCode)
		row.ItemCode = strings.TrimSpace(row.ItemCode)
		row.SchemaJSON = json.RawMessage(raw)
		out = append(out, row)
	}
	return out, rows.Err()
}

func (repo *Repository) ListVBMAPPResponseSchemaOverrides(
	ctx context.Context,
	scaleVersion string,
	instID int64,
) ([]VBMAPPResponseSchemaOverride, error) {
	if instID <= 0 {
		return nil, nil
	}
	rows, err := repo.db.QueryContext(ctx, `
		SELECT module_code, item_code, override_json
		FROM vbmapp_response_schema_override
		WHERE inst_id = ? AND scale_version = ? AND status = 'active' AND del_flag = 0
		ORDER BY module_code, item_code, id
	`, instID, strings.TrimSpace(scaleVersion))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]VBMAPPResponseSchemaOverride, 0)
	for rows.Next() {
		var row VBMAPPResponseSchemaOverride
		var raw string
		if err := rows.Scan(&row.ModuleCode, &row.ItemCode, &raw); err != nil {
			return nil, err
		}
		row.ModuleCode = strings.TrimSpace(row.ModuleCode)
		row.ItemCode = strings.TrimSpace(row.ItemCode)
		row.OverrideJSON = json.RawMessage(raw)
		out = append(out, row)
	}
	return out, rows.Err()
}
