package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type AssessmentScaleDatasetEntity struct {
	ScaleCode    string
	ScaleVersion string
	DataStatus   string
	Sources      []string
}

type AssessmentScaleItemEntity struct {
	ItemNo int
	Raw    json.RawMessage
}

type AssessmentScaleDomainEntity struct {
	DomainCode string
	SortNo     int
	Raw        json.RawMessage
}

type AssessmentScaleNormRecordEntity struct {
	RecordKey string
	SortNo    int
	Raw       json.RawMessage
}

type AssessmentScaleStaticDataEntity struct {
	Dataset     AssessmentScaleDatasetEntity
	Items       []AssessmentScaleItemEntity
	Domains     []AssessmentScaleDomainEntity
	NormRecords []AssessmentScaleNormRecordEntity
}

func ensureAssessmentScaleTables(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS assessment_scale_dataset (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			scale_code VARCHAR(64) NOT NULL DEFAULT '',
			scale_version VARCHAR(64) NOT NULL DEFAULT '',
			data_status VARCHAR(1000) NOT NULL DEFAULT '',
			sources_json LONGTEXT NOT NULL,
			create_id BIGINT NOT NULL DEFAULT 0,
			update_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uk_assessment_scale_dataset (scale_code, scale_version)
		)
	`); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS assessment_scale_item (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			scale_code VARCHAR(64) NOT NULL DEFAULT '',
			scale_version VARCHAR(64) NOT NULL DEFAULT '',
			item_no INT NOT NULL DEFAULT 0,
			item_json LONGTEXT NOT NULL,
			create_id BIGINT NOT NULL DEFAULT 0,
			update_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uk_assessment_scale_item (scale_code, scale_version, item_no),
			KEY idx_assessment_scale_item_version (scale_code, scale_version, item_no)
		)
	`); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS assessment_scale_domain (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			scale_code VARCHAR(64) NOT NULL DEFAULT '',
			scale_version VARCHAR(64) NOT NULL DEFAULT '',
			domain_code VARCHAR(64) NOT NULL DEFAULT '',
			sort_no INT NOT NULL DEFAULT 0,
			domain_json LONGTEXT NOT NULL,
			create_id BIGINT NOT NULL DEFAULT 0,
			update_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uk_assessment_scale_domain (scale_code, scale_version, domain_code),
			KEY idx_assessment_scale_domain_version (scale_code, scale_version, sort_no)
		)
	`); err != nil {
		return err
	}
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS assessment_scale_norm_record (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			scale_code VARCHAR(64) NOT NULL DEFAULT '',
			scale_version VARCHAR(64) NOT NULL DEFAULT '',
			record_key VARCHAR(128) NOT NULL DEFAULT '',
			sort_no INT NOT NULL DEFAULT 0,
			norm_json LONGTEXT NOT NULL,
			create_id BIGINT NOT NULL DEFAULT 0,
			update_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uk_assessment_scale_norm_record (scale_code, scale_version, record_key),
			KEY idx_assessment_scale_norm_version (scale_code, scale_version, sort_no)
		)
	`)
	return err
}

func (repo *Repository) HasAssessmentScaleStaticData(ctx context.Context, scaleCode, scaleVersion string) (bool, error) {
	scaleCode = strings.TrimSpace(scaleCode)
	scaleVersion = strings.TrimSpace(scaleVersion)
	var itemCount, domainCount, normCount int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM assessment_scale_item
		WHERE scale_code = ? AND scale_version = ? AND del_flag = 0
	`, scaleCode, scaleVersion).Scan(&itemCount); err != nil {
		return false, err
	}
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM assessment_scale_domain
		WHERE scale_code = ? AND scale_version = ? AND del_flag = 0
	`, scaleCode, scaleVersion).Scan(&domainCount); err != nil {
		return false, err
	}
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM assessment_scale_norm_record
		WHERE scale_code = ? AND scale_version = ? AND del_flag = 0
	`, scaleCode, scaleVersion).Scan(&normCount); err != nil {
		return false, err
	}
	return itemCount > 0 && domainCount > 0 && normCount > 0, nil
}

func (repo *Repository) ReplaceAssessmentScaleStaticData(ctx context.Context, data AssessmentScaleStaticDataEntity, operatorID int64) error {
	scaleCode := strings.TrimSpace(data.Dataset.ScaleCode)
	scaleVersion := strings.TrimSpace(data.Dataset.ScaleVersion)
	if scaleCode == "" || scaleVersion == "" {
		return fmt.Errorf("scale code and version are required")
	}
	sourcesRaw, err := json.Marshal(data.Dataset.Sources)
	if err != nil {
		return fmt.Errorf("marshal assessment scale sources: %w", err)
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
		INSERT INTO assessment_scale_dataset (
			scale_code, scale_version, data_status, sources_json, create_id, update_id, create_time, update_time, del_flag
		) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
		ON DUPLICATE KEY UPDATE
			data_status = VALUES(data_status),
			sources_json = VALUES(sources_json),
			update_id = VALUES(update_id),
			update_time = NOW(),
			del_flag = 0
	`, scaleCode, scaleVersion, strings.TrimSpace(data.Dataset.DataStatus), string(sourcesRaw), operatorID, operatorID); err != nil {
		return err
	}
	for _, table := range []string{"assessment_scale_item", "assessment_scale_domain", "assessment_scale_norm_record"} {
		if _, err = tx.ExecContext(ctx, fmt.Sprintf(`
			UPDATE %s
			SET del_flag = 1, update_id = ?, update_time = NOW()
			WHERE scale_code = ? AND scale_version = ? AND del_flag = 0
		`, table), operatorID, scaleCode, scaleVersion); err != nil {
			return err
		}
	}
	for _, item := range data.Items {
		if item.ItemNo <= 0 || len(item.Raw) == 0 {
			continue
		}
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO assessment_scale_item (
				scale_code, scale_version, item_no, item_json, create_id, update_id, create_time, update_time, del_flag
			) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
			ON DUPLICATE KEY UPDATE
				item_json = VALUES(item_json),
				update_id = VALUES(update_id),
				update_time = NOW(),
				del_flag = 0
		`, scaleCode, scaleVersion, item.ItemNo, string(item.Raw), operatorID, operatorID); err != nil {
			return err
		}
	}
	for _, domain := range data.Domains {
		domainCode := strings.TrimSpace(domain.DomainCode)
		if domainCode == "" || len(domain.Raw) == 0 {
			continue
		}
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO assessment_scale_domain (
				scale_code, scale_version, domain_code, sort_no, domain_json, create_id, update_id, create_time, update_time, del_flag
			) VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
			ON DUPLICATE KEY UPDATE
				sort_no = VALUES(sort_no),
				domain_json = VALUES(domain_json),
				update_id = VALUES(update_id),
				update_time = NOW(),
				del_flag = 0
		`, scaleCode, scaleVersion, domainCode, domain.SortNo, string(domain.Raw), operatorID, operatorID); err != nil {
			return err
		}
	}
	for _, record := range data.NormRecords {
		recordKey := strings.TrimSpace(record.RecordKey)
		if recordKey == "" {
			recordKey = strconv.Itoa(record.SortNo)
		}
		if len(record.Raw) == 0 {
			continue
		}
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO assessment_scale_norm_record (
				scale_code, scale_version, record_key, sort_no, norm_json, create_id, update_id, create_time, update_time, del_flag
			) VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
			ON DUPLICATE KEY UPDATE
				sort_no = VALUES(sort_no),
				norm_json = VALUES(norm_json),
				update_id = VALUES(update_id),
				update_time = NOW(),
				del_flag = 0
		`, scaleCode, scaleVersion, recordKey, record.SortNo, string(record.Raw), operatorID, operatorID); err != nil {
			return err
		}
	}
	err = tx.Commit()
	return err
}

func (repo *Repository) GetAssessmentScaleDataset(ctx context.Context, scaleCode, scaleVersion string) (AssessmentScaleDatasetEntity, error) {
	var (
		data       AssessmentScaleDatasetEntity
		sourcesRaw string
	)
	if err := repo.db.QueryRowContext(ctx, `
		SELECT scale_code, scale_version, data_status, sources_json
		FROM assessment_scale_dataset
		WHERE scale_code = ? AND scale_version = ? AND del_flag = 0
		LIMIT 1
	`, strings.TrimSpace(scaleCode), strings.TrimSpace(scaleVersion)).Scan(&data.ScaleCode, &data.ScaleVersion, &data.DataStatus, &sourcesRaw); err != nil {
		return AssessmentScaleDatasetEntity{}, err
	}
	_ = json.Unmarshal([]byte(sourcesRaw), &data.Sources)
	return data, nil
}

func (repo *Repository) ListAssessmentScaleItems(ctx context.Context, scaleCode, scaleVersion string) ([]AssessmentScaleItemEntity, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT item_no, item_json
		FROM assessment_scale_item
		WHERE scale_code = ? AND scale_version = ? AND del_flag = 0
		ORDER BY item_no
	`, strings.TrimSpace(scaleCode), strings.TrimSpace(scaleVersion))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]AssessmentScaleItemEntity, 0)
	for rows.Next() {
		var item AssessmentScaleItemEntity
		var raw string
		if err := rows.Scan(&item.ItemNo, &raw); err != nil {
			return nil, err
		}
		item.Raw = json.RawMessage(raw)
		out = append(out, item)
	}
	return out, rows.Err()
}

func (repo *Repository) ListAssessmentScaleDomains(ctx context.Context, scaleCode, scaleVersion string) ([]AssessmentScaleDomainEntity, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT domain_code, sort_no, domain_json
		FROM assessment_scale_domain
		WHERE scale_code = ? AND scale_version = ? AND del_flag = 0
		ORDER BY sort_no, id
	`, strings.TrimSpace(scaleCode), strings.TrimSpace(scaleVersion))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]AssessmentScaleDomainEntity, 0)
	for rows.Next() {
		var item AssessmentScaleDomainEntity
		var raw string
		if err := rows.Scan(&item.DomainCode, &item.SortNo, &raw); err != nil {
			return nil, err
		}
		item.Raw = json.RawMessage(raw)
		out = append(out, item)
	}
	return out, rows.Err()
}

func (repo *Repository) ListAssessmentScaleNormRecords(ctx context.Context, scaleCode, scaleVersion string) ([]AssessmentScaleNormRecordEntity, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT record_key, sort_no, norm_json
		FROM assessment_scale_norm_record
		WHERE scale_code = ? AND scale_version = ? AND del_flag = 0
		ORDER BY sort_no, id
	`, strings.TrimSpace(scaleCode), strings.TrimSpace(scaleVersion))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]AssessmentScaleNormRecordEntity, 0)
	for rows.Next() {
		var item AssessmentScaleNormRecordEntity
		var raw string
		if err := rows.Scan(&item.RecordKey, &item.SortNo, &raw); err != nil {
			return nil, err
		}
		item.Raw = json.RawMessage(raw)
		out = append(out, item)
	}
	return out, rows.Err()
}
