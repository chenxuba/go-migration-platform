package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"go-migration-platform/services/education/internal/model"
)

type intentionStudentImportTaskEntity struct {
	Detail  model.IntentionStudentImportTaskDetail
	Columns []model.IntentionStudentImportColumn
	Rows    []model.IntentionStudentImportRow
}

func (repo *Repository) CreateIntentionStudentImportTask(ctx context.Context, instID int64, detail model.IntentionStudentImportTaskDetail, columns []model.IntentionStudentImportColumn, rows []model.IntentionStudentImportRow) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	columnsRaw, err := json.Marshal(columns)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO intention_student_import_task (
			id, inst_id, file_name, upload_staff_id, upload_staff_name, execute_staff_id, execute_staff_name,
			total_rows, executed_rows, deleted_rows, error_rows, created_time, confirm_time, complete_time,
			status, inst_name, columns_json, del_flag
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
	`,
		detail.ID, instID, detail.FileName, detail.UploadStaffID, detail.UploadStaffName, detail.ExecuteStaffID, detail.ExecuteStaffName,
		detail.TotalRows, detail.ExecutedRows, detail.DeletedRows, detail.ErrorRows, detail.CreatedTime, detail.ConfirmTime, detail.CompleteTime,
		detail.Status, detail.InstName, string(columnsRaw),
	); err != nil {
		return err
	}

	for _, row := range rows {
		cellsRaw, err := json.Marshal(row.Cells)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO intention_student_import_task_record (
				id, task_id, row_no, has_error, status, result, cells_json, create_time, update_time, del_flag
			) VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
		`, row.ID, detail.ID, row.RowNo, row.HasError, row.Status, row.Result, string(cellsRaw)); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (repo *Repository) GetIntentionStudentImportTask(ctx context.Context, taskID string) (intentionStudentImportTaskEntity, error) {
	entity := intentionStudentImportTaskEntity{}
	var (
		confirmTime  sql.NullTime
		completeTime sql.NullTime
		columnsRaw   string
		executeID    sql.NullString
		executeName  sql.NullString
	)
	err := repo.db.QueryRowContext(ctx, `
		SELECT id, file_name, upload_staff_id, upload_staff_name, execute_staff_id, execute_staff_name,
		       total_rows, executed_rows, deleted_rows, error_rows, created_time, confirm_time, complete_time,
		       status, inst_name, IFNULL(columns_json, '')
		FROM intention_student_import_task
		WHERE id = ? AND del_flag = 0
		LIMIT 1
	`, taskID).Scan(
		&entity.Detail.ID,
		&entity.Detail.FileName,
		&entity.Detail.UploadStaffID,
		&entity.Detail.UploadStaffName,
		&executeID,
		&executeName,
		&entity.Detail.TotalRows,
		&entity.Detail.ExecutedRows,
		&entity.Detail.DeletedRows,
		&entity.Detail.ErrorRows,
		&entity.Detail.CreatedTime,
		&confirmTime,
		&completeTime,
		&entity.Detail.Status,
		&entity.Detail.InstName,
		&columnsRaw,
	)
	if err != nil {
		return intentionStudentImportTaskEntity{}, err
	}
	if executeID.Valid {
		value := executeID.String
		entity.Detail.ExecuteStaffID = &value
	}
	if executeName.Valid {
		value := executeName.String
		entity.Detail.ExecuteStaffName = &value
	}
	if confirmTime.Valid {
		t := confirmTime.Time
		entity.Detail.ConfirmTime = &t
	}
	if completeTime.Valid {
		t := completeTime.Time
		entity.Detail.CompleteTime = &t
	}
	if columnsRaw != "" {
		if err := json.Unmarshal([]byte(columnsRaw), &entity.Columns); err != nil {
			return intentionStudentImportTaskEntity{}, err
		}
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, row_no, has_error, status, IFNULL(result, ''), IFNULL(cells_json, '')
		FROM intention_student_import_task_record
		WHERE task_id = ? AND del_flag = 0
		ORDER BY row_no ASC, id ASC
	`, taskID)
	if err != nil {
		return intentionStudentImportTaskEntity{}, err
	}
	defer rows.Close()

	entity.Rows = make([]model.IntentionStudentImportRow, 0, 32)
	for rows.Next() {
		var (
			row      model.IntentionStudentImportRow
			cellsRaw string
		)
		if err := rows.Scan(&row.ID, &row.RowNo, &row.HasError, &row.Status, &row.Result, &cellsRaw); err != nil {
			return intentionStudentImportTaskEntity{}, err
		}
		if cellsRaw != "" {
			if err := json.Unmarshal([]byte(cellsRaw), &row.Cells); err != nil {
				return intentionStudentImportTaskEntity{}, err
			}
		}
		entity.Rows = append(entity.Rows, row)
	}
	if err := rows.Err(); err != nil {
		return intentionStudentImportTaskEntity{}, err
	}
	return entity, nil
}

func (repo *Repository) ListIntentionStudentImportTasks(ctx context.Context, instID int64) (model.IntentionStudentImportTaskListResult, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, file_name, upload_staff_id, upload_staff_name, execute_staff_id, execute_staff_name,
		       total_rows, executed_rows, deleted_rows, error_rows, created_time, confirm_time, complete_time,
		       status, inst_name
		FROM intention_student_import_task
		WHERE del_flag = 0 AND inst_id = ?
		ORDER BY created_time DESC, id DESC
	`, instID)
	if err != nil {
		return model.IntentionStudentImportTaskListResult{}, err
	}
	defer rows.Close()

	items := make([]model.IntentionStudentImportTaskDetail, 0, 32)
	for rows.Next() {
		var (
			item         model.IntentionStudentImportTaskDetail
			confirmTime  sql.NullTime
			completeTime sql.NullTime
			executeID    sql.NullString
			executeName  sql.NullString
		)
		if err := rows.Scan(
			&item.ID, &item.FileName, &item.UploadStaffID, &item.UploadStaffName, &executeID, &executeName,
			&item.TotalRows, &item.ExecutedRows, &item.DeletedRows, &item.ErrorRows, &item.CreatedTime, &confirmTime, &completeTime,
			&item.Status, &item.InstName,
		); err != nil {
			return model.IntentionStudentImportTaskListResult{}, err
		}
		if executeID.Valid {
			value := executeID.String
			item.ExecuteStaffID = &value
		}
		if executeName.Valid {
			value := executeName.String
			item.ExecuteStaffName = &value
		}
		if confirmTime.Valid {
			t := confirmTime.Time
			item.ConfirmTime = &t
		}
		if completeTime.Valid {
			t := completeTime.Time
			item.CompleteTime = &t
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return model.IntentionStudentImportTaskListResult{}, err
	}
	return model.IntentionStudentImportTaskListResult{List: items, Total: len(items)}, nil
}

func (repo *Repository) UpdateIntentionStudentImportTask(ctx context.Context, detail model.IntentionStudentImportTaskDetail, rows []model.IntentionStudentImportRow) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		UPDATE intention_student_import_task
		SET executed_rows = ?, deleted_rows = ?, error_rows = ?, confirm_time = ?, complete_time = ?, status = ?,
		    execute_staff_id = ?, execute_staff_name = ?, update_time = NOW()
		WHERE id = ? AND del_flag = 0
	`, detail.ExecutedRows, detail.DeletedRows, detail.ErrorRows, detail.ConfirmTime, detail.CompleteTime, detail.Status,
		detail.ExecuteStaffID, detail.ExecuteStaffName, detail.ID,
	); err != nil {
		return err
	}

	for _, row := range rows {
		cellsRaw, err := json.Marshal(row.Cells)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE intention_student_import_task_record
			SET has_error = ?, status = ?, result = ?, cells_json = ?, update_time = NOW()
			WHERE id = ? AND task_id = ? AND del_flag = 0
		`, row.HasError, row.Status, row.Result, string(cellsRaw), row.ID, detail.ID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (repo *Repository) ClearIntentionStudentImportTasks(ctx context.Context, instID int64) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		UPDATE intention_student_import_task_record r
		INNER JOIN intention_student_import_task t ON t.id = r.task_id
		SET r.del_flag = 1, r.update_time = NOW()
		WHERE t.inst_id = ? AND t.del_flag = 0 AND r.del_flag = 0
	`, instID); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE intention_student_import_task
		SET del_flag = 1, update_time = NOW()
		WHERE inst_id = ? AND del_flag = 0
	`, instID); err != nil {
		return err
	}

	return tx.Commit()
}

func (repo *Repository) DeleteIntentionStudentImportTask(ctx context.Context, instID int64, taskID string) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		UPDATE intention_student_import_task_record r
		INNER JOIN intention_student_import_task t ON t.id = r.task_id
		SET r.del_flag = 1, r.update_time = NOW()
		WHERE t.inst_id = ? AND t.id = ? AND t.del_flag = 0 AND r.del_flag = 0
	`, instID, taskID); err != nil {
		return err
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE intention_student_import_task
		SET del_flag = 1, update_time = NOW()
		WHERE inst_id = ? AND id = ? AND del_flag = 0
	`, instID, taskID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err == nil && affected == 0 {
		return sql.ErrNoRows
	}
	return tx.Commit()
}

func ensureIntentionStudentImportTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS intention_student_import_task (
			id VARCHAR(64) PRIMARY KEY,
			inst_id BIGINT NOT NULL DEFAULT 0,
			file_name VARCHAR(255) NOT NULL,
			upload_staff_id VARCHAR(64) NOT NULL DEFAULT '',
			upload_staff_name VARCHAR(100) NOT NULL DEFAULT '',
			execute_staff_id VARCHAR(64) NULL DEFAULT NULL,
			execute_staff_name VARCHAR(100) NULL DEFAULT NULL,
			total_rows INT NOT NULL DEFAULT 0,
			executed_rows INT NOT NULL DEFAULT 0,
			deleted_rows INT NOT NULL DEFAULT 0,
			error_rows INT NOT NULL DEFAULT 0,
			created_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			confirm_time DATETIME NULL DEFAULT NULL,
			complete_time DATETIME NULL DEFAULT NULL,
			status INT NOT NULL DEFAULT 0,
			inst_name VARCHAR(255) NOT NULL DEFAULT '',
			columns_json LONGTEXT NOT NULL,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			KEY idx_import_task_inst (inst_id),
			KEY idx_import_task_created (created_time)
		)
	`)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS intention_student_import_task_record (
			id VARCHAR(64) PRIMARY KEY,
			task_id VARCHAR(64) NOT NULL,
			row_no INT NOT NULL DEFAULT 0,
			has_error TINYINT(1) NOT NULL DEFAULT 0,
			status INT NOT NULL DEFAULT 0,
			result VARCHAR(1000) NOT NULL DEFAULT '',
			cells_json LONGTEXT NOT NULL,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			KEY idx_import_task_record_task (task_id),
			KEY idx_import_task_record_row (row_no)
		)
	`)
	return err
}

var errImportTaskNotFound = errors.New("import task not found")
