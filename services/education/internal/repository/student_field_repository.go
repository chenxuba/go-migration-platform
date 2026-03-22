package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

func (repo *Repository) ListStudentFields(ctx context.Context, instID int64, isDefault bool) ([]model.StudentFieldKey, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, IFNULL(uuid, ''), IFNULL(version, 0), inst_id, field_key, field_type,
		       IFNULL(required, 0), IFNULL(searched, 0), IFNULL(options_json, ''),
		       IFNULL(is_default, 0), IFNULL(is_display, 0), IFNULL(can_delete, 0),
		       IFNULL(can_edit, 0), IFNULL(sort, 0), IFNULL(remark, '')
		FROM inst_student_field_key
		WHERE inst_id = ? AND is_default = ? AND del_flag = 0
		ORDER BY sort ASC, id ASC
	`, instID, isDefault)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.StudentFieldKey, 0, 32)
	for rows.Next() {
		var item model.StudentFieldKey
		if err := rows.Scan(
			&item.ID,
			&item.UUID,
			&item.Version,
			&item.InstID,
			&item.FieldKey,
			&item.FieldType,
			&item.Required,
			&item.Searched,
			&item.OptionsJSON,
			&item.IsDefault,
			&item.IsDisplay,
			&item.CanDelete,
			&item.CanEdit,
			&item.Sort,
			&item.Remark,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repo *Repository) GetStudentFieldByID(ctx context.Context, id int64) (model.StudentFieldKey, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT id, IFNULL(uuid, ''), IFNULL(version, 0), inst_id, field_key, field_type,
		       IFNULL(required, 0), IFNULL(searched, 0), IFNULL(options_json, ''),
		       IFNULL(is_default, 0), IFNULL(is_display, 0), IFNULL(can_delete, 0),
		       IFNULL(can_edit, 0), IFNULL(sort, 0), IFNULL(remark, '')
		FROM inst_student_field_key
		WHERE id = ? AND del_flag = 0
		LIMIT 1
	`, id)

	var item model.StudentFieldKey
	err := row.Scan(
		&item.ID,
		&item.UUID,
		&item.Version,
		&item.InstID,
		&item.FieldKey,
		&item.FieldType,
		&item.Required,
		&item.Searched,
		&item.OptionsJSON,
		&item.IsDefault,
		&item.IsDisplay,
		&item.CanDelete,
		&item.CanEdit,
		&item.Sort,
		&item.Remark,
	)
	return item, err
}

func (repo *Repository) MaxStudentFieldSort(ctx context.Context, instID int64) (int, error) {
	var sort sql.NullInt64
	err := repo.db.QueryRowContext(ctx, `
		SELECT MAX(sort)
		FROM inst_student_field_key
		WHERE inst_id = ? AND del_flag = 0
	`, instID).Scan(&sort)
	if err != nil {
		return 0, err
	}
	if !sort.Valid {
		return 0, nil
	}
	return int(sort.Int64), nil
}

func (repo *Repository) CreateStudentCustomField(ctx context.Context, instID int64, field model.StudentFieldKey) (int64, error) {
	result, err := repo.db.ExecContext(ctx, `
		INSERT INTO inst_student_field_key (
			uuid, version, inst_id, field_key, field_type, required, searched, options_json,
			is_default, is_display, can_delete, can_edit, sort, remark, del_flag, create_time
		) VALUES (
			UUID(), 0, ?, ?, ?, ?, ?, ?, 0, 1, 1, 1, ?, ?, 0, NOW()
		)
	`, instID, strings.TrimSpace(field.FieldKey), field.FieldType, field.Required, field.Searched, strings.TrimSpace(field.OptionsJSON), field.Sort, strings.TrimSpace(field.Remark))
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (repo *Repository) UpdateStudentFieldDisplayStatus(ctx context.Context, id int64, isDisplay bool) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE inst_student_field_key
		SET is_display = ?, version = IFNULL(version, 0) + 1, update_time = NOW()
		WHERE id = ? AND del_flag = 0
	`, isDisplay, id)
	return err
}

func (repo *Repository) UpdateStudentCustomField(ctx context.Context, field model.StudentFieldKey) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE inst_student_field_key
		SET field_key = ?, field_type = ?, required = ?, searched = ?, options_json = ?, remark = ?,
		    version = IFNULL(version, 0) + 1, update_time = NOW()
		WHERE id = ? AND del_flag = 0
	`, strings.TrimSpace(field.FieldKey), field.FieldType, field.Required, field.Searched, strings.TrimSpace(field.OptionsJSON), strings.TrimSpace(field.Remark), field.ID)
	return err
}

func (repo *Repository) DeleteStudentCustomField(ctx context.Context, id int64) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE inst_student_field_key
		SET del_flag = 1, version = IFNULL(version, 0) + 1, update_time = NOW()
		WHERE id = ? AND del_flag = 0
	`, id)
	return err
}

func (repo *Repository) SortStudentCustomFields(ctx context.Context, fields []model.StudentFieldKey) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, field := range fields {
		if field.ID <= 0 {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE inst_student_field_key
			SET sort = ?, version = IFNULL(version, 0) + 1, update_time = NOW()
			WHERE id = ? AND del_flag = 0
		`, field.Sort, field.ID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (repo *Repository) GetStudentFieldDetail(ctx context.Context, id int64) (model.StudentFieldDetail, error) {
	field, err := repo.GetStudentFieldByID(ctx, id)
	if err != nil {
		return model.StudentFieldDetail{}, err
	}
	_, err = repo.db.ExecContext(ctx, `
		UPDATE inst_student_field_key
		SET can_delete = 1, is_display = 1, update_time = NOW()
		WHERE id = ? AND del_flag = 0
	`, id)
	if err != nil {
		return model.StudentFieldDetail{}, err
	}
	return model.StudentFieldDetail{
		ID:        field.ID,
		Name:      field.FieldKey,
		Type:      field.FieldType,
		IsDisplay: true,
		Required:  field.Required,
		Searched:  field.Searched,
		Sort:      field.Sort,
	}, nil
}

func (repo *Repository) InitInstStudentField(ctx context.Context, instID int64) error {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT field_key, field_type, IFNULL(required, 0), IFNULL(searched, 0), IFNULL(options_json, ''),
		       IFNULL(is_default, 0), IFNULL(is_display, 0), IFNULL(can_delete, 0), IFNULL(can_edit, 0), IFNULL(sort, 0), IFNULL(remark, '')
		FROM inst_student_field_key
		WHERE inst_id IS NULL AND is_default = 1 AND del_flag = 0
		ORDER BY sort ASC, id ASC
	`)
	if err != nil {
		return err
	}
	defer rows.Close()
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for rows.Next() {
		var field model.StudentFieldKey
		if err := rows.Scan(&field.FieldKey, &field.FieldType, &field.Required, &field.Searched, &field.OptionsJSON, &field.IsDefault, &field.IsDisplay, &field.CanDelete, &field.CanEdit, &field.Sort, &field.Remark); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO inst_student_field_key (
				uuid, version, inst_id, field_key, field_type, required, searched, options_json,
				is_default, is_display, can_delete, can_edit, sort, remark, del_flag, create_time
			) VALUES (
				UUID(), 0, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, NOW()
			)
		`, instID, field.FieldKey, field.FieldType, field.Required, field.Searched, field.OptionsJSON, field.IsDefault, field.IsDisplay, field.CanDelete, field.CanEdit, field.Sort, field.Remark); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return tx.Commit()
}

func ensureFieldExists(field model.StudentFieldKey) error {
	if strings.TrimSpace(field.FieldKey) == "" {
		return fmt.Errorf("fieldKey is required")
	}
	if field.FieldType <= 0 {
		return fmt.Errorf("fieldType is required")
	}
	return nil
}
