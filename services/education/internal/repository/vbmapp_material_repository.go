package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"go-migration-platform/pkg/vbmappscore"
	"go-migration-platform/services/education/internal/model"
)

func ensureVBMAPPMaterialLibraryTables(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS vbmapp_material_profile (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			scale_version VARCHAR(64) NOT NULL DEFAULT '',
			library_scope VARCHAR(16) NOT NULL DEFAULT 'platform',
			inst_id BIGINT NOT NULL DEFAULT 0,
			profile_id VARCHAR(100) NOT NULL DEFAULT '',
			label VARCHAR(255) NOT NULL DEFAULT '',
			source_logic TEXT NOT NULL,
			suggested_types_json LONGTEXT NOT NULL,
			preparation_checks_json LONGTEXT NOT NULL,
			sort_no INT NOT NULL DEFAULT 0,
			status VARCHAR(16) NOT NULL DEFAULT 'active',
			create_id BIGINT NOT NULL DEFAULT 0,
			update_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uk_vbmapp_material_profile (scale_version, library_scope, inst_id, profile_id),
			KEY idx_vbmapp_material_profile_scope (scale_version, library_scope, inst_id, status, del_flag)
		)
	`); err != nil {
		return err
	}
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS vbmapp_material_item (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			scale_version VARCHAR(64) NOT NULL DEFAULT '',
			library_scope VARCHAR(16) NOT NULL DEFAULT 'platform',
			inst_id BIGINT NOT NULL DEFAULT 0,
			profile_id VARCHAR(100) NOT NULL DEFAULT '',
			material_code VARCHAR(100) NOT NULL DEFAULT '',
			material_name VARCHAR(255) NOT NULL DEFAULT '',
			material_type VARCHAR(100) NOT NULL DEFAULT '',
			sort_no INT NOT NULL DEFAULT 0,
			status VARCHAR(16) NOT NULL DEFAULT 'active',
			create_id BIGINT NOT NULL DEFAULT 0,
			update_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uk_vbmapp_material_item (scale_version, library_scope, inst_id, profile_id, material_code),
			KEY idx_vbmapp_material_item_profile (scale_version, library_scope, inst_id, profile_id, status, del_flag),
			KEY idx_vbmapp_material_item_name (scale_version, material_name, status, del_flag)
		)
	`)
	return err
}

func (repo *Repository) HasVBMAPPResponseMaterialProfiles(ctx context.Context, scaleVersion string) (bool, error) {
	var count int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(1)
		FROM vbmapp_material_profile
		WHERE scale_version = ? AND library_scope = 'platform' AND inst_id = 0
		  AND status = 'active' AND del_flag = 0
	`, strings.TrimSpace(scaleVersion)).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (repo *Repository) ReplaceVBMAPPResponseMaterialProfiles(ctx context.Context, scaleVersion string, profiles map[string]vbmappscore.ResponseMaterialProfile, operatorID int64) error {
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
		UPDATE vbmapp_material_profile
		SET del_flag = 1, update_id = ?, update_time = NOW()
		WHERE scale_version = ? AND library_scope = 'platform' AND inst_id = 0 AND del_flag = 0
	`, operatorID, scaleVersion); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `
		UPDATE vbmapp_material_item
		SET del_flag = 1, update_id = ?, update_time = NOW()
		WHERE scale_version = ? AND library_scope = 'platform' AND inst_id = 0 AND del_flag = 0
	`, operatorID, scaleVersion); err != nil {
		return err
	}
	keys := make([]string, 0, len(profiles))
	for key := range profiles {
		if strings.TrimSpace(key) != "" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	for profileIndex, profileID := range keys {
		profile := profiles[profileID]
		suggestedTypesRaw, marshalErr := json.Marshal(profile.SuggestedTypes)
		if marshalErr != nil {
			return fmt.Errorf("marshal VB-MAPP material profile %s suggested types: %w", profileID, marshalErr)
		}
		preparationChecksRaw, marshalErr := json.Marshal(profile.PreparationChecks)
		if marshalErr != nil {
			return fmt.Errorf("marshal VB-MAPP material profile %s preparation checks: %w", profileID, marshalErr)
		}
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO vbmapp_material_profile (
				scale_version, library_scope, inst_id, profile_id, label, source_logic,
				suggested_types_json, preparation_checks_json, sort_no, status,
				create_id, update_id, create_time, update_time, del_flag
			) VALUES (?, 'platform', 0, ?, ?, ?, ?, ?, ?, 'active', ?, ?, NOW(), NOW(), 0)
			ON DUPLICATE KEY UPDATE
				label = VALUES(label),
				source_logic = VALUES(source_logic),
				suggested_types_json = VALUES(suggested_types_json),
				preparation_checks_json = VALUES(preparation_checks_json),
				sort_no = VALUES(sort_no),
				status = VALUES(status),
				update_id = VALUES(update_id),
				update_time = NOW(),
				del_flag = 0
		`, scaleVersion, profileID, strings.TrimSpace(profile.Label), strings.TrimSpace(profile.SourceLogic),
			string(suggestedTypesRaw), string(preparationChecksRaw), profileIndex+1, operatorID, operatorID); err != nil {
			return err
		}
		for itemIndex, material := range profile.RecommendedMaterials {
			name := strings.TrimSpace(material.Name)
			if name == "" {
				continue
			}
			code := strings.TrimSpace(material.ID)
			if code == "" {
				code = fmt.Sprintf("%s_%03d", profileID, itemIndex+1)
			}
			if _, err = tx.ExecContext(ctx, `
				INSERT INTO vbmapp_material_item (
					scale_version, library_scope, inst_id, profile_id, material_code, material_name,
					material_type, sort_no, status, create_id, update_id, create_time, update_time, del_flag
				) VALUES (?, 'platform', 0, ?, ?, ?, ?, ?, 'active', ?, ?, NOW(), NOW(), 0)
				ON DUPLICATE KEY UPDATE
					material_name = VALUES(material_name),
					material_type = VALUES(material_type),
					sort_no = VALUES(sort_no),
					status = VALUES(status),
					update_id = VALUES(update_id),
					update_time = NOW(),
					del_flag = 0
			`, scaleVersion, profileID, code, name, strings.TrimSpace(material.Type), itemIndex+1, operatorID, operatorID); err != nil {
				return err
			}
		}
	}
	err = tx.Commit()
	return err
}

func (repo *Repository) ListVBMAPPResponseMaterialProfiles(ctx context.Context, scaleVersion string) (map[string]vbmappscore.ResponseMaterialProfile, error) {
	scaleVersion = strings.TrimSpace(scaleVersion)
	rows, err := repo.db.QueryContext(ctx, `
		SELECT profile_id, label, source_logic, suggested_types_json, preparation_checks_json
		FROM vbmapp_material_profile
		WHERE scale_version = ? AND library_scope = 'platform' AND inst_id = 0
		  AND status = 'active' AND del_flag = 0
		ORDER BY sort_no, id
	`, scaleVersion)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	profiles := make(map[string]vbmappscore.ResponseMaterialProfile)
	for rows.Next() {
		var (
			profileID            string
			profile              vbmappscore.ResponseMaterialProfile
			suggestedTypesRaw    string
			preparationChecksRaw string
		)
		if err := rows.Scan(&profileID, &profile.Label, &profile.SourceLogic, &suggestedTypesRaw, &preparationChecksRaw); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(suggestedTypesRaw), &profile.SuggestedTypes)
		_ = json.Unmarshal([]byte(preparationChecksRaw), &profile.PreparationChecks)
		profiles[strings.TrimSpace(profileID)] = profile
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(profiles) == 0 {
		return profiles, nil
	}
	itemRows, err := repo.db.QueryContext(ctx, `
		SELECT profile_id, material_code, material_name, material_type
		FROM vbmapp_material_item
		WHERE scale_version = ? AND library_scope = 'platform' AND inst_id = 0
		  AND status = 'active' AND del_flag = 0
		ORDER BY profile_id, sort_no, id
	`, scaleVersion)
	if err != nil {
		return nil, err
	}
	defer itemRows.Close()
	for itemRows.Next() {
		var (
			profileID string
			material  vbmappscore.ResponseMaterialSuggestion
		)
		if err := itemRows.Scan(&profileID, &material.ID, &material.Name, &material.Type); err != nil {
			return nil, err
		}
		profile, ok := profiles[strings.TrimSpace(profileID)]
		if !ok {
			continue
		}
		profile.RecommendedMaterials = append(profile.RecommendedMaterials, material)
		profiles[strings.TrimSpace(profileID)] = profile
	}
	return profiles, itemRows.Err()
}

func (repo *Repository) ListVBMAPPMaterialProfiles(ctx context.Context, scaleVersion string) ([]model.VBMAPPMaterialProfile, error) {
	scaleVersion = strings.TrimSpace(scaleVersion)
	rows, err := repo.db.QueryContext(ctx, `
		SELECT profile_id, label, source_logic, suggested_types_json, preparation_checks_json
		FROM vbmapp_material_profile
		WHERE scale_version = ? AND library_scope = 'platform' AND inst_id = 0
		  AND status = 'active' AND del_flag = 0
		ORDER BY sort_no, id
	`, scaleVersion)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]model.VBMAPPMaterialProfile, 0)
	index := make(map[string]int)
	for rows.Next() {
		var (
			item                 model.VBMAPPMaterialProfile
			suggestedTypesRaw    string
			preparationChecksRaw string
		)
		if err := rows.Scan(&item.ProfileID, &item.Label, &item.SourceLogic, &suggestedTypesRaw, &preparationChecksRaw); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(suggestedTypesRaw), &item.SuggestedTypes)
		_ = json.Unmarshal([]byte(preparationChecksRaw), &item.PreparationChecks)
		index[item.ProfileID] = len(out)
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	itemRows, err := repo.db.QueryContext(ctx, `
		SELECT id, scale_version, library_scope, inst_id, profile_id, material_code, material_name,
		       material_type, sort_no, status, create_time, update_time
		FROM vbmapp_material_item
		WHERE scale_version = ? AND library_scope = 'platform' AND inst_id = 0
		  AND status = 'active' AND del_flag = 0
		ORDER BY profile_id, sort_no, id
	`, scaleVersion)
	if err != nil {
		return nil, err
	}
	defer itemRows.Close()
	for itemRows.Next() {
		var item model.VBMAPPMaterialItem
		if err := itemRows.Scan(
			&item.ID,
			&item.ScaleVersion,
			&item.LibraryScope,
			&item.InstID,
			&item.ProfileID,
			&item.MaterialCode,
			&item.MaterialName,
			&item.MaterialType,
			&item.SortNo,
			&item.Status,
			&item.CreatedTime,
			&item.UpdatedTime,
		); err != nil {
			return nil, err
		}
		if profileIndex, ok := index[item.ProfileID]; ok {
			out[profileIndex].Materials = append(out[profileIndex].Materials, item)
		}
	}
	return out, itemRows.Err()
}

func (repo *Repository) SaveVBMAPPMaterialItem(ctx context.Context, scaleVersion string, operatorID int64, item model.VBMAPPMaterialItem) (model.VBMAPPMaterialItem, error) {
	scaleVersion = strings.TrimSpace(scaleVersion)
	profileID := strings.TrimSpace(item.ProfileID)
	materialName := strings.TrimSpace(item.MaterialName)
	if scaleVersion == "" || profileID == "" || materialName == "" {
		return model.VBMAPPMaterialItem{}, fmt.Errorf("scale version, profile id and material name are required")
	}
	materialCode := strings.TrimSpace(item.MaterialCode)
	if materialCode == "" {
		materialCode = fmt.Sprintf("custom_%d", time.Now().UnixNano())
	}
	status := strings.TrimSpace(item.Status)
	if status == "" {
		status = "active"
	}
	if item.ID > 0 {
		result, err := repo.db.ExecContext(ctx, `
			UPDATE vbmapp_material_item
			SET material_code = ?, material_name = ?, material_type = ?, sort_no = ?,
			    status = ?, update_id = ?, update_time = NOW(), del_flag = 0
			WHERE id = ? AND scale_version = ? AND library_scope = 'platform' AND inst_id = 0 AND del_flag = 0
		`, materialCode, materialName, strings.TrimSpace(item.MaterialType), item.SortNo, status, operatorID, item.ID, scaleVersion)
		if err != nil {
			return model.VBMAPPMaterialItem{}, err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return model.VBMAPPMaterialItem{}, err
		}
		if affected == 0 {
			return model.VBMAPPMaterialItem{}, sql.ErrNoRows
		}
		return repo.GetVBMAPPMaterialItem(ctx, scaleVersion, item.ID)
	}
	result, err := repo.db.ExecContext(ctx, `
		INSERT INTO vbmapp_material_item (
			scale_version, library_scope, inst_id, profile_id, material_code, material_name,
			material_type, sort_no, status, create_id, update_id, create_time, update_time, del_flag
		) VALUES (?, 'platform', 0, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
		ON DUPLICATE KEY UPDATE
			material_name = VALUES(material_name),
			material_type = VALUES(material_type),
			sort_no = VALUES(sort_no),
			status = VALUES(status),
			update_id = VALUES(update_id),
			update_time = NOW(),
			del_flag = 0
	`, scaleVersion, profileID, materialCode, materialName, strings.TrimSpace(item.MaterialType), item.SortNo, status, operatorID, operatorID)
	if err != nil {
		return model.VBMAPPMaterialItem{}, err
	}
	id, _ := result.LastInsertId()
	if id > 0 {
		return repo.GetVBMAPPMaterialItem(ctx, scaleVersion, id)
	}
	return repo.GetVBMAPPMaterialItemByCode(ctx, scaleVersion, profileID, materialCode)
}

func (repo *Repository) GetVBMAPPMaterialItem(ctx context.Context, scaleVersion string, id int64) (model.VBMAPPMaterialItem, error) {
	var item model.VBMAPPMaterialItem
	if err := repo.db.QueryRowContext(ctx, `
		SELECT id, scale_version, library_scope, inst_id, profile_id, material_code, material_name,
		       material_type, sort_no, status, create_time, update_time
		FROM vbmapp_material_item
		WHERE id = ? AND scale_version = ? AND library_scope = 'platform' AND inst_id = 0 AND del_flag = 0
		LIMIT 1
	`, id, strings.TrimSpace(scaleVersion)).Scan(
		&item.ID,
		&item.ScaleVersion,
		&item.LibraryScope,
		&item.InstID,
		&item.ProfileID,
		&item.MaterialCode,
		&item.MaterialName,
		&item.MaterialType,
		&item.SortNo,
		&item.Status,
		&item.CreatedTime,
		&item.UpdatedTime,
	); err != nil {
		return model.VBMAPPMaterialItem{}, err
	}
	return item, nil
}

func (repo *Repository) GetVBMAPPMaterialItemByCode(ctx context.Context, scaleVersion, profileID, materialCode string) (model.VBMAPPMaterialItem, error) {
	var item model.VBMAPPMaterialItem
	if err := repo.db.QueryRowContext(ctx, `
		SELECT id, scale_version, library_scope, inst_id, profile_id, material_code, material_name,
		       material_type, sort_no, status, create_time, update_time
		FROM vbmapp_material_item
		WHERE scale_version = ? AND library_scope = 'platform' AND inst_id = 0
		  AND profile_id = ? AND material_code = ? AND del_flag = 0
		LIMIT 1
	`, strings.TrimSpace(scaleVersion), strings.TrimSpace(profileID), strings.TrimSpace(materialCode)).Scan(
		&item.ID,
		&item.ScaleVersion,
		&item.LibraryScope,
		&item.InstID,
		&item.ProfileID,
		&item.MaterialCode,
		&item.MaterialName,
		&item.MaterialType,
		&item.SortNo,
		&item.Status,
		&item.CreatedTime,
		&item.UpdatedTime,
	); err != nil {
		return model.VBMAPPMaterialItem{}, err
	}
	return item, nil
}

func (repo *Repository) DeleteVBMAPPMaterialItem(ctx context.Context, scaleVersion string, id, operatorID int64) error {
	result, err := repo.db.ExecContext(ctx, `
		UPDATE vbmapp_material_item
		SET del_flag = 1, update_id = ?, update_time = NOW()
		WHERE id = ? AND scale_version = ? AND library_scope = 'platform' AND inst_id = 0 AND del_flag = 0
	`, operatorID, id, strings.TrimSpace(scaleVersion))
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
