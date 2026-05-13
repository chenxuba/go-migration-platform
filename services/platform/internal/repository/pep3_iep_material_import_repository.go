package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"unicode"

	"go-migration-platform/services/platform/internal/model"
)

type PEP3IEPMaterialImportTaskEntity struct {
	Detail  model.PEP3IEPMaterialImportTaskDetail
	Columns []model.PEP3IEPMaterialImportColumn
	Rows    []model.PEP3IEPMaterialImportRow
}

func ensurePEP3IEPMaterialImportTables(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS pep3_iep_material_import_task (
			id VARCHAR(64) PRIMARY KEY,
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
			KEY idx_pep3_iep_import_task_created (created_time),
			KEY idx_pep3_iep_import_task_status (status, del_flag)
		)
	`); err != nil {
		return err
	}
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS pep3_iep_material_import_task_record (
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
			KEY idx_pep3_iep_import_record_task (task_id),
			KEY idx_pep3_iep_import_record_row (row_no)
		)
	`)
	return err
}

func (repo *Repository) CreatePlatformPEP3IEPMaterialImportTask(ctx context.Context, detail model.PEP3IEPMaterialImportTaskDetail, columns []model.PEP3IEPMaterialImportColumn, rows []model.PEP3IEPMaterialImportRow) error {
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
		INSERT INTO pep3_iep_material_import_task (
			id, file_name, upload_staff_id, upload_staff_name, execute_staff_id, execute_staff_name,
			total_rows, executed_rows, deleted_rows, error_rows, created_time, confirm_time, complete_time,
			status, inst_name, columns_json, del_flag
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
	`,
		detail.ID, detail.FileName, detail.UploadStaffID, detail.UploadStaffName, detail.ExecuteStaffID, detail.ExecuteStaffName,
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
			INSERT INTO pep3_iep_material_import_task_record (
				id, task_id, row_no, has_error, status, result, cells_json, create_time, update_time, del_flag
			) VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
		`, row.ID, detail.ID, row.RowNo, row.HasError, row.Status, row.Result, string(cellsRaw)); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (repo *Repository) GetPlatformPEP3IEPMaterialImportTask(ctx context.Context, taskID string) (PEP3IEPMaterialImportTaskEntity, error) {
	entity := PEP3IEPMaterialImportTaskEntity{}
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
		FROM pep3_iep_material_import_task
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
		return PEP3IEPMaterialImportTaskEntity{}, err
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
			return PEP3IEPMaterialImportTaskEntity{}, err
		}
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, row_no, has_error, status, IFNULL(result, ''), IFNULL(cells_json, '')
		FROM pep3_iep_material_import_task_record
		WHERE task_id = ? AND del_flag = 0
		ORDER BY row_no ASC, id ASC
	`, taskID)
	if err != nil {
		return PEP3IEPMaterialImportTaskEntity{}, err
	}
	defer rows.Close()

	entity.Rows = make([]model.PEP3IEPMaterialImportRow, 0, 32)
	for rows.Next() {
		var row model.PEP3IEPMaterialImportRow
		var cellsRaw string
		if err := rows.Scan(&row.ID, &row.RowNo, &row.HasError, &row.Status, &row.Result, &cellsRaw); err != nil {
			return PEP3IEPMaterialImportTaskEntity{}, err
		}
		if cellsRaw != "" {
			if err := json.Unmarshal([]byte(cellsRaw), &row.Cells); err != nil {
				return PEP3IEPMaterialImportTaskEntity{}, err
			}
		}
		entity.Rows = append(entity.Rows, row)
	}
	if err := rows.Err(); err != nil {
		return PEP3IEPMaterialImportTaskEntity{}, err
	}
	return entity, nil
}

func (repo *Repository) ListPlatformPEP3IEPMaterialImportTasks(ctx context.Context) (model.PEP3IEPMaterialImportTaskListResult, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, file_name, upload_staff_id, upload_staff_name, execute_staff_id, execute_staff_name,
		       total_rows, executed_rows, deleted_rows, error_rows, created_time, confirm_time, complete_time,
		       status, inst_name
		FROM pep3_iep_material_import_task
		WHERE del_flag = 0
		ORDER BY created_time DESC, id DESC
	`)
	if err != nil {
		return model.PEP3IEPMaterialImportTaskListResult{}, err
	}
	defer rows.Close()

	items := make([]model.PEP3IEPMaterialImportTaskDetail, 0, 32)
	for rows.Next() {
		var (
			item         model.PEP3IEPMaterialImportTaskDetail
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
			return model.PEP3IEPMaterialImportTaskListResult{}, err
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
		return model.PEP3IEPMaterialImportTaskListResult{}, err
	}
	return model.PEP3IEPMaterialImportTaskListResult{List: items, Total: len(items)}, nil
}

func (repo *Repository) UpdatePlatformPEP3IEPMaterialImportTask(ctx context.Context, detail model.PEP3IEPMaterialImportTaskDetail, rows []model.PEP3IEPMaterialImportRow) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		UPDATE pep3_iep_material_import_task
		SET total_rows = ?, executed_rows = ?, deleted_rows = ?, error_rows = ?, confirm_time = ?, complete_time = ?, status = ?,
		    execute_staff_id = ?, execute_staff_name = ?, update_time = NOW()
		WHERE id = ? AND del_flag = 0
	`, detail.TotalRows, detail.ExecutedRows, detail.DeletedRows, detail.ErrorRows, detail.ConfirmTime, detail.CompleteTime, detail.Status,
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
			UPDATE pep3_iep_material_import_task_record
			SET has_error = ?, status = ?, result = ?, cells_json = ?, update_time = NOW()
			WHERE id = ? AND task_id = ? AND del_flag = 0
		`, row.HasError, row.Status, row.Result, string(cellsRaw), row.ID, detail.ID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (repo *Repository) ClearPlatformPEP3IEPMaterialImportTasks(ctx context.Context) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		UPDATE pep3_iep_material_import_task_record r
		INNER JOIN pep3_iep_material_import_task t ON t.id = r.task_id
		SET r.del_flag = 1, r.update_time = NOW()
		WHERE t.del_flag = 0 AND r.del_flag = 0
	`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE pep3_iep_material_import_task
		SET del_flag = 1, update_time = NOW()
		WHERE del_flag = 0
	`); err != nil {
		return err
	}
	return tx.Commit()
}

func (repo *Repository) DeletePlatformPEP3IEPMaterialImportTask(ctx context.Context, taskID string) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		UPDATE pep3_iep_material_import_task_record r
		INNER JOIN pep3_iep_material_import_task t ON t.id = r.task_id
		SET r.del_flag = 1, r.update_time = NOW()
		WHERE t.id = ? AND t.del_flag = 0 AND r.del_flag = 0
	`, taskID); err != nil {
		return err
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE pep3_iep_material_import_task
		SET del_flag = 1, update_time = NOW()
		WHERE id = ? AND del_flag = 0
	`, taskID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err == nil && affected == 0 {
		return sql.ErrNoRows
	}
	return tx.Commit()
}

type PlatformPEP3IEPMaterialImportSaveInput struct {
	ItemNo             int
	ItemTitle          string
	DomainCode         string
	Domain             string
	ScoreValue         int
	ScoreLabel         string
	ScoreDescription   string
	LongGoal           string
	ShortGoal          string
	CourseForm         string
	TrainingProject    string
	TrainingContent    string
	Status             string
	ExistingRuleID     int64
	ExistingLongGoalID int64
}

type PlatformPEP3IEPMaterialImportSaveResult struct {
	RuleID     int64
	LongGoalID int64
}

func (repo *Repository) SavePlatformPEP3IEPMaterialImportRow(ctx context.Context, userID int64, input PlatformPEP3IEPMaterialImportSaveInput) (PlatformPEP3IEPMaterialImportSaveResult, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return PlatformPEP3IEPMaterialImportSaveResult{}, err
	}
	defer tx.Rollback()

	status := platformPEP3IEPMaterialStatus(input.Status)
	longGoalID, err := createPlatformPEP3IEPLongGoalTx(ctx, tx, userID, input, status)
	if err != nil {
		return PlatformPEP3IEPMaterialImportSaveResult{}, err
	}
	ruleID, err := createPlatformPEP3IEPRuleTx(ctx, tx, userID, input, status)
	if err != nil {
		return PlatformPEP3IEPMaterialImportSaveResult{}, err
	}
	if err := ensurePlatformPEP3IEPRuleGoalRelTx(ctx, tx, ruleID, longGoalID); err != nil {
		return PlatformPEP3IEPMaterialImportSaveResult{}, err
	}
	if input.ShortGoal != "" {
		shortGoalID, err := findOrCreatePlatformPEP3IEPShortGoalTx(ctx, tx, userID, input, status, longGoalID)
		if err != nil {
			return PlatformPEP3IEPMaterialImportSaveResult{}, err
		}
		if input.TrainingProject != "" || input.TrainingContent != "" {
			if err := findOrCreatePlatformPEP3IEPTrainingTx(ctx, tx, userID, input, status, shortGoalID); err != nil {
				return PlatformPEP3IEPMaterialImportSaveResult{}, err
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return PlatformPEP3IEPMaterialImportSaveResult{}, err
	}
	return PlatformPEP3IEPMaterialImportSaveResult{
		RuleID:     ruleID,
		LongGoalID: longGoalID,
	}, nil
}

func createPlatformPEP3IEPLongGoalTx(ctx context.Context, tx *sql.Tx, userID int64, input PlatformPEP3IEPMaterialImportSaveInput, status string) (int64, error) {
	if input.ExistingLongGoalID > 0 {
		return input.ExistingLongGoalID, nil
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO pep3_iep_goal_material (
			library_scope, inst_id, material_type, parent_goal_material_id, domain_code, domain,
			long_goal, short_goal, course_form, age_min_months, age_max_months, difficulty_level,
			applicable_score_values, priority, status, create_id, update_id, create_time, update_time, del_flag
		) VALUES ('platform', 0, 'long_term', 0, ?, ?, ?, '', '', 0, 0, 0, '', 100, ?, ?, ?, NOW(), NOW(), 0)
	`, input.DomainCode, input.Domain, input.LongGoal, status, userID, userID)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func createPlatformPEP3IEPRuleTx(ctx context.Context, tx *sql.Tx, userID int64, input PlatformPEP3IEPMaterialImportSaveInput, status string) (int64, error) {
	if input.ExistingRuleID > 0 {
		return input.ExistingRuleID, nil
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO pep3_iep_item_option_rule (
			library_scope, inst_id, item_no, item_title, domain_code, domain, score_value, score_label,
			score_description, result_meaning, generate_policy, priority, ai_instruction, status,
			create_id, update_id, create_time, update_time, del_flag
		) VALUES ('platform', 0, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, '', ?, ?, ?, NOW(), NOW(), 0)
	`,
		input.ItemNo, input.ItemTitle, input.DomainCode, input.Domain, input.ScoreValue, input.ScoreLabel,
		input.ScoreDescription, platformPEP3IEPDefaultResultMeaning(input.ScoreValue), platformPEP3IEPDefaultGeneratePolicy(input.ScoreValue),
		platformPEP3IEPDefaultPriority(input.ScoreValue), status, userID, userID,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func ensurePlatformPEP3IEPRuleGoalRelTx(ctx context.Context, tx *sql.Tx, ruleID, goalID int64) error {
	var id int64
	err := tx.QueryRowContext(ctx, `
		SELECT id
		FROM pep3_iep_item_rule_goal_rel
		WHERE rule_id = ? AND goal_material_id = ? AND del_flag = 0
		LIMIT 1
	`, ruleID, goalID).Scan(&id)
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO pep3_iep_item_rule_goal_rel (rule_id, goal_material_id, create_time, update_time, del_flag)
		VALUES (?, ?, NOW(), NOW(), 0)
	`, ruleID, goalID)
	return err
}

func findOrCreatePlatformPEP3IEPShortGoalTx(ctx context.Context, tx *sql.Tx, userID int64, input PlatformPEP3IEPMaterialImportSaveInput, status string, parentID int64) (int64, error) {
	targetShortGoalKey := pep3IEPImportTextMatchKey(input.ShortGoal)
	rows, err := tx.QueryContext(ctx, `
		SELECT id, IFNULL(short_goal, ''), IFNULL(course_form, '')
		FROM pep3_iep_goal_material
		WHERE library_scope = 'platform' AND inst_id = 0 AND material_type = 'short_term'
		  AND parent_goal_material_id = ? AND del_flag = 0
		ORDER BY id ASC
	`, parentID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			id         int64
			shortGoal  string
			courseForm string
		)
		if err := rows.Scan(&id, &shortGoal, &courseForm); err != nil {
			return 0, err
		}
		if strings.TrimSpace(courseForm) != strings.TrimSpace(input.CourseForm) {
			continue
		}
		if pep3IEPImportTextMatchKey(shortGoal) != targetShortGoalKey {
			continue
		}
		_, _ = tx.ExecContext(ctx, `
			UPDATE pep3_iep_goal_material
			SET domain_code = ?, domain = ?, long_goal = ?, short_goal = ?, course_form = ?, status = ?, update_id = ?, update_time = NOW()
			WHERE id = ? AND del_flag = 0
		`, input.DomainCode, input.Domain, input.LongGoal, input.ShortGoal, input.CourseForm, status, userID, id)
		return id, nil
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO pep3_iep_goal_material (
			library_scope, inst_id, material_type, parent_goal_material_id, domain_code, domain,
			long_goal, short_goal, course_form, age_min_months, age_max_months, difficulty_level,
			applicable_score_values, priority, status, create_id, update_id, create_time, update_time, del_flag
		) VALUES ('platform', 0, 'short_term', ?, ?, ?, ?, ?, ?, 0, 0, 0, '', 100, ?, ?, ?, NOW(), NOW(), 0)
	`, parentID, input.DomainCode, input.Domain, input.LongGoal, input.ShortGoal, input.CourseForm, status, userID, userID)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func findOrCreatePlatformPEP3IEPTrainingTx(ctx context.Context, tx *sql.Tx, userID int64, input PlatformPEP3IEPMaterialImportSaveInput, status string, shortGoalID int64) error {
	targetProjectKey := pep3IEPImportTextMatchKey(input.TrainingProject)
	targetContentKey := pep3IEPImportTextMatchKey(input.TrainingContent)
	rows, err := tx.QueryContext(ctx, `
		SELECT id, IFNULL(training_project, ''), IFNULL(training_content, '')
		FROM pep3_iep_training_material
		WHERE library_scope = 'platform' AND inst_id = 0 AND goal_material_id = ?
		  AND del_flag = 0
		ORDER BY id ASC
	`, shortGoalID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			id              int64
			trainingProject string
			trainingContent string
		)
		if err := rows.Scan(&id, &trainingProject, &trainingContent); err != nil {
			return err
		}
		if pep3IEPImportTextMatchKey(trainingProject) != targetProjectKey || pep3IEPImportTextMatchKey(trainingContent) != targetContentKey {
			continue
		}
		_, _ = tx.ExecContext(ctx, `
			UPDATE pep3_iep_training_material
			SET training_project = ?, training_content = ?, status = ?, update_id = ?, update_time = NOW()
			WHERE id = ? AND del_flag = 0
		`, input.TrainingProject, input.TrainingContent, status, userID, id)
		return nil
	}
	if err := rows.Err(); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO pep3_iep_training_material (
			library_scope, inst_id, goal_material_id, training_project, training_content,
			priority, status, create_id, update_id, create_time, update_time, del_flag
		) VALUES ('platform', 0, ?, ?, ?, 100, ?, ?, ?, NOW(), NOW(), 0)
	`, shortGoalID, input.TrainingProject, input.TrainingContent, status, userID, userID)
	return err
}

func pep3IEPImportTextMatchKey(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	var builder strings.Builder
	builder.Grow(len(text))
	for _, r := range text {
		if unicode.IsSpace(r) {
			continue
		}
		builder.WriteRune(r)
	}
	return builder.String()
}

func platformPEP3IEPDefaultResultMeaning(score int) string {
	switch score {
	case 2:
		return "已通过，默认用于维持、泛化或提高独立性。"
	case 1:
		return "部分通过，优先转化为季度或半年度IEP目标。"
	case 0:
		return "未通过，生成前备或基础目标。"
	default:
		return ""
	}
}

func platformPEP3IEPDefaultGeneratePolicy(score int) string {
	switch score {
	case 2:
		return "skip_or_generalize"
	case 0:
		return "prerequisite_goal"
	default:
		return "primary_goal"
	}
}

func platformPEP3IEPDefaultPriority(score int) int {
	switch score {
	case 1:
		return 100
	case 0:
		return 80
	case 2:
		return 20
	default:
		return 0
	}
}
