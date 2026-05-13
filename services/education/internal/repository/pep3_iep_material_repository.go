package repository

import (
	"context"
	"database/sql"
	"sort"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

func ensurePEP3IEPMaterialTables(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS pep3_iep_item_option_rule (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			library_scope VARCHAR(16) NOT NULL DEFAULT 'institution',
			inst_id BIGINT NOT NULL DEFAULT 0,
			item_no INT NOT NULL DEFAULT 0,
			item_title VARCHAR(500) NOT NULL DEFAULT '',
			domain_code VARCHAR(64) NOT NULL DEFAULT '',
			domain VARCHAR(128) NOT NULL DEFAULT '',
			score_value INT NOT NULL DEFAULT -1,
			score_label VARCHAR(64) NOT NULL DEFAULT '',
			score_description VARCHAR(1000) NOT NULL DEFAULT '',
			result_meaning VARCHAR(1000) NOT NULL DEFAULT '',
			generate_policy VARCHAR(32) NOT NULL DEFAULT '',
			priority INT NOT NULL DEFAULT 0,
			ai_instruction TEXT NOT NULL,
			status VARCHAR(16) NOT NULL DEFAULT 'active',
			create_id BIGINT NOT NULL DEFAULT 0,
			update_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			KEY idx_pep3_iep_item_rule_match (library_scope, inst_id, item_no, score_value, status, del_flag),
			KEY idx_pep3_iep_item_rule_domain (library_scope, inst_id, domain_code, status, del_flag)
		)
	`); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS pep3_iep_goal_material (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			library_scope VARCHAR(16) NOT NULL DEFAULT 'institution',
			inst_id BIGINT NOT NULL DEFAULT 0,
			material_type VARCHAR(16) NOT NULL DEFAULT 'long_term',
			parent_goal_material_id BIGINT NOT NULL DEFAULT 0,
			domain_code VARCHAR(64) NOT NULL DEFAULT '',
			domain VARCHAR(128) NOT NULL DEFAULT '',
			long_goal TEXT NOT NULL,
			short_goal TEXT NOT NULL,
			course_form VARCHAR(64) NOT NULL DEFAULT '',
			age_min_months INT NOT NULL DEFAULT 0,
			age_max_months INT NOT NULL DEFAULT 0,
			difficulty_level INT NOT NULL DEFAULT 0,
			applicable_score_values VARCHAR(32) NOT NULL DEFAULT '',
			priority INT NOT NULL DEFAULT 0,
			status VARCHAR(16) NOT NULL DEFAULT 'active',
			create_id BIGINT NOT NULL DEFAULT 0,
			update_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			KEY idx_pep3_iep_goal_parent (library_scope, inst_id, parent_goal_material_id, status, del_flag),
			KEY idx_pep3_iep_goal_type (library_scope, inst_id, material_type, status, del_flag),
			KEY idx_pep3_iep_goal_scope (library_scope, inst_id, status, del_flag),
			KEY idx_pep3_iep_goal_domain (library_scope, inst_id, domain_code, status, del_flag)
		)
	`); err != nil {
		return err
	}
	if err := ensurePEP3IEPMaterialGoalColumns(ctx, db); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS pep3_iep_item_rule_goal_rel (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			rule_id BIGINT NOT NULL DEFAULT 0,
			goal_material_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uk_pep3_iep_rule_goal_rel (rule_id, goal_material_id, del_flag),
			KEY idx_pep3_iep_rule_goal_rule (rule_id, del_flag),
			KEY idx_pep3_iep_rule_goal_goal (goal_material_id, del_flag)
		)
	`); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS pep3_iep_monthly_material`); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS pep3_iep_weekly_material`); err != nil {
		return err
	}
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS pep3_iep_training_material (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			library_scope VARCHAR(16) NOT NULL DEFAULT 'institution',
			inst_id BIGINT NOT NULL DEFAULT 0,
			goal_material_id BIGINT NOT NULL DEFAULT 0,
			training_project VARCHAR(500) NOT NULL DEFAULT '',
			training_content TEXT NOT NULL,
			priority INT NOT NULL DEFAULT 0,
			status VARCHAR(16) NOT NULL DEFAULT 'active',
			create_id BIGINT NOT NULL DEFAULT 0,
			update_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			KEY idx_pep3_iep_training_scope (library_scope, inst_id, status, del_flag),
			KEY idx_pep3_iep_training_goal (goal_material_id, status, del_flag)
		)
	`)
	return err
}

func ensurePEP3IEPMaterialGoalColumns(ctx context.Context, db *sql.DB) error {
	columns := []struct {
		name string
		sql  string
	}{
		{
			name: "material_type",
			sql:  "ALTER TABLE pep3_iep_goal_material ADD COLUMN material_type VARCHAR(16) NOT NULL DEFAULT 'long_term' AFTER inst_id",
		},
		{
			name: "parent_goal_material_id",
			sql:  "ALTER TABLE pep3_iep_goal_material ADD COLUMN parent_goal_material_id BIGINT NOT NULL DEFAULT 0 AFTER material_type",
		},
	}
	for _, column := range columns {
		var count int
		if err := db.QueryRowContext(ctx, `
			SELECT COUNT(*)
			FROM information_schema.COLUMNS
			WHERE TABLE_SCHEMA = DATABASE()
			  AND TABLE_NAME = 'pep3_iep_goal_material'
			  AND COLUMN_NAME = ?
		`, column.name).Scan(&count); err != nil {
			return err
		}
		if count == 0 {
			if _, err := db.ExecContext(ctx, column.sql); err != nil {
				return err
			}
		}
	}
	return nil
}

func (repo *Repository) PagePEP3IEPItemOptionRules(ctx context.Context, instID int64, query model.PEP3IEPItemOptionRulePageQuery) (model.PageResult[model.PEP3IEPItemOptionRule], error) {
	current, size := normalizeAssessmentPage(query.PageRequestModel.PageIndex, query.PageRequestModel.PageSize)
	where, args := pep3IEPRuleWhere(instID, query.QueryModel)
	var total int
	if err := repo.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM pep3_iep_item_option_rule WHERE `+where, args...).Scan(&total); err != nil {
		return model.PageResult[model.PEP3IEPItemOptionRule]{}, err
	}
	listArgs := append(append([]any{}, args...), size, (current-1)*size)
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, library_scope, inst_id, item_no, item_title, domain_code, domain,
		       score_value, score_label, score_description, result_meaning, generate_policy,
		       priority, ai_instruction, status, create_time, update_time
		FROM pep3_iep_item_option_rule
		WHERE `+where+`
		ORDER BY FIELD(library_scope, 'institution', 'platform'), priority DESC, item_no ASC, score_value ASC, update_time DESC, id DESC
		LIMIT ? OFFSET ?
	`, listArgs...)
	if err != nil {
		return model.PageResult[model.PEP3IEPItemOptionRule]{}, err
	}
	defer rows.Close()
	items := make([]model.PEP3IEPItemOptionRule, 0, size)
	for rows.Next() {
		item, err := scanPEP3IEPItemOptionRule(rows)
		if err != nil {
			return model.PageResult[model.PEP3IEPItemOptionRule]{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return model.PageResult[model.PEP3IEPItemOptionRule]{}, err
	}
	if err := repo.attachPEP3IEPRuleGoalMaterials(ctx, items); err != nil {
		return model.PageResult[model.PEP3IEPItemOptionRule]{}, err
	}
	return model.PageResult[model.PEP3IEPItemOptionRule]{
		Items:   items,
		Total:   total,
		Current: current,
		Size:    size,
	}, nil
}

func (repo *Repository) SavePEP3IEPItemOptionRule(ctx context.Context, instID, userID int64, item model.PEP3IEPItemOptionRule) (model.PEP3IEPItemOptionRule, error) {
	scope, rowInstID := pep3IEPMaterialScopeAndInstID(instID, item.LibraryScope)
	status := pep3IEPMaterialStatus(item.Status)
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return model.PEP3IEPItemOptionRule{}, err
	}
	defer tx.Rollback()
	if item.ID > 0 {
		result, err := tx.ExecContext(ctx, `
			UPDATE pep3_iep_item_option_rule
			SET library_scope = ?, inst_id = ?, item_no = ?, item_title = ?, domain_code = ?, domain = ?,
			    score_value = ?, score_label = ?, score_description = ?, result_meaning = ?, generate_policy = ?,
			    priority = ?, ai_instruction = ?, status = ?, update_id = ?, update_time = NOW()
			WHERE id = ? AND del_flag = 0
			  AND library_scope = 'institution' AND inst_id = ?
		`,
			scope, rowInstID, item.ItemNo, strings.TrimSpace(item.ItemTitle), strings.TrimSpace(item.DomainCode), strings.TrimSpace(item.Domain),
			item.ScoreValue, strings.TrimSpace(item.ScoreLabel), strings.TrimSpace(item.ScoreDescription), strings.TrimSpace(item.ResultMeaning), strings.TrimSpace(item.GeneratePolicy),
			item.Priority, strings.TrimSpace(item.AIInstruction), status, userID, item.ID, instID,
		)
		if err != nil {
			return model.PEP3IEPItemOptionRule{}, err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return model.PEP3IEPItemOptionRule{}, err
		}
		if affected == 0 {
			return model.PEP3IEPItemOptionRule{}, sql.ErrNoRows
		}
	} else {
		result, err := tx.ExecContext(ctx, `
			INSERT INTO pep3_iep_item_option_rule (
				library_scope, inst_id, item_no, item_title, domain_code, domain,
				score_value, score_label, score_description, result_meaning, generate_policy,
				priority, ai_instruction, status, create_id, update_id, create_time, update_time, del_flag
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
		`,
			scope, rowInstID, item.ItemNo, strings.TrimSpace(item.ItemTitle), strings.TrimSpace(item.DomainCode), strings.TrimSpace(item.Domain),
			item.ScoreValue, strings.TrimSpace(item.ScoreLabel), strings.TrimSpace(item.ScoreDescription), strings.TrimSpace(item.ResultMeaning), strings.TrimSpace(item.GeneratePolicy),
			item.Priority, strings.TrimSpace(item.AIInstruction), status, userID, userID,
		)
		if err != nil {
			return model.PEP3IEPItemOptionRule{}, err
		}
		item.ID, err = result.LastInsertId()
		if err != nil {
			return model.PEP3IEPItemOptionRule{}, err
		}
	}
	if err := replacePEP3IEPRuleGoalRelsTx(ctx, tx, item.ID, item.GoalMaterialIDs); err != nil {
		return model.PEP3IEPItemOptionRule{}, err
	}
	if err := tx.Commit(); err != nil {
		return model.PEP3IEPItemOptionRule{}, err
	}
	return repo.GetPEP3IEPItemOptionRule(ctx, instID, item.ID)
}

func replacePEP3IEPRuleGoalRelsTx(ctx context.Context, tx *sql.Tx, ruleID int64, goalMaterialIDs []int64) error {
	if ruleID <= 0 {
		return nil
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM pep3_iep_item_rule_goal_rel WHERE rule_id = ?`, ruleID); err != nil {
		return err
	}
	seen := map[int64]bool{}
	for _, goalID := range goalMaterialIDs {
		if goalID <= 0 || seen[goalID] {
			continue
		}
		seen[goalID] = true
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO pep3_iep_item_rule_goal_rel (rule_id, goal_material_id, create_time, update_time, del_flag)
			VALUES (?, ?, NOW(), NOW(), 0)
		`, ruleID, goalID); err != nil {
			return err
		}
	}
	return nil
}

func (repo *Repository) GetPEP3IEPItemOptionRule(ctx context.Context, instID, id int64) (model.PEP3IEPItemOptionRule, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT id, library_scope, inst_id, item_no, item_title, domain_code, domain,
		       score_value, score_label, score_description, result_meaning, generate_policy,
		       priority, ai_instruction, status, create_time, update_time
		FROM pep3_iep_item_option_rule
		WHERE id = ? AND del_flag = 0
		  AND library_scope = 'institution' AND inst_id = ?
		LIMIT 1
	`, id, instID)
	item, err := scanPEP3IEPItemOptionRule(row)
	if err != nil {
		return model.PEP3IEPItemOptionRule{}, err
	}
	items := []model.PEP3IEPItemOptionRule{item}
	if err := repo.attachPEP3IEPRuleGoalMaterials(ctx, items); err != nil {
		return model.PEP3IEPItemOptionRule{}, err
	}
	return items[0], nil
}

func (repo *Repository) DeletePEP3IEPItemOptionRule(ctx context.Context, instID, id int64) error {
	result, err := repo.db.ExecContext(ctx, `
		UPDATE pep3_iep_item_option_rule
		SET del_flag = 1, update_time = NOW()
		WHERE id = ? AND del_flag = 0
		  AND library_scope = 'institution' AND inst_id = ?
	`, id, instID)
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

func (repo *Repository) PagePEP3IEPGoalMaterials(ctx context.Context, instID int64, query model.PEP3IEPGoalMaterialPageQuery) (model.PageResult[model.PEP3IEPGoalMaterial], error) {
	current, size := normalizeAssessmentPage(query.PageRequestModel.PageIndex, query.PageRequestModel.PageSize)
	where, args := pep3IEPGoalMaterialWhere(instID, query.QueryModel)
	var total int
	if err := repo.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM pep3_iep_goal_material WHERE `+where, args...).Scan(&total); err != nil {
		return model.PageResult[model.PEP3IEPGoalMaterial]{}, err
	}
	listArgs := append(append([]any{}, args...), size, (current-1)*size)
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, library_scope, inst_id, material_type, parent_goal_material_id, domain_code, domain, long_goal, short_goal, course_form,
		       age_min_months, age_max_months, difficulty_level, applicable_score_values,
		       priority, status, create_time, update_time
		FROM pep3_iep_goal_material
		WHERE `+where+`
		ORDER BY FIELD(library_scope, 'institution', 'platform'), priority DESC, update_time DESC, id DESC
		LIMIT ? OFFSET ?
	`, listArgs...)
	if err != nil {
		return model.PageResult[model.PEP3IEPGoalMaterial]{}, err
	}
	defer rows.Close()
	items := make([]model.PEP3IEPGoalMaterial, 0, size)
	for rows.Next() {
		item, err := scanPEP3IEPGoalMaterial(rows)
		if err != nil {
			return model.PageResult[model.PEP3IEPGoalMaterial]{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return model.PageResult[model.PEP3IEPGoalMaterial]{}, err
	}
	return model.PageResult[model.PEP3IEPGoalMaterial]{Items: items, Total: total, Current: current, Size: size}, nil
}

func (repo *Repository) SavePEP3IEPGoalMaterial(ctx context.Context, instID, userID int64, item model.PEP3IEPGoalMaterial) (model.PEP3IEPGoalMaterial, error) {
	scope, rowInstID := pep3IEPMaterialScopeAndInstID(instID, item.LibraryScope)
	status := pep3IEPMaterialStatus(item.Status)
	if item.ID > 0 {
		result, err := repo.db.ExecContext(ctx, `
			UPDATE pep3_iep_goal_material
			SET library_scope = ?, inst_id = ?, material_type = ?, parent_goal_material_id = ?,
			    domain_code = ?, domain = ?, long_goal = ?, short_goal = ?,
			    course_form = ?, age_min_months = ?, age_max_months = ?, difficulty_level = ?,
			    applicable_score_values = ?, priority = ?, status = ?, update_id = ?, update_time = NOW()
			WHERE id = ? AND del_flag = 0
			  AND library_scope = 'institution' AND inst_id = ?
		`,
			scope, rowInstID, strings.TrimSpace(item.MaterialType), item.ParentGoalMaterialID,
			strings.TrimSpace(item.DomainCode), strings.TrimSpace(item.Domain), strings.TrimSpace(item.LongGoal), strings.TrimSpace(item.ShortGoal),
			strings.TrimSpace(item.CourseForm), item.AgeMinMonths, item.AgeMaxMonths, item.DifficultyLevel,
			strings.TrimSpace(item.ApplicableScoreValues), item.Priority, status, userID, item.ID, instID,
		)
		if err != nil {
			return model.PEP3IEPGoalMaterial{}, err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return model.PEP3IEPGoalMaterial{}, err
		}
		if affected == 0 {
			return model.PEP3IEPGoalMaterial{}, sql.ErrNoRows
		}
	} else {
		result, err := repo.db.ExecContext(ctx, `
			INSERT INTO pep3_iep_goal_material (
				library_scope, inst_id, material_type, parent_goal_material_id, domain_code, domain, long_goal, short_goal, course_form,
				age_min_months, age_max_months, difficulty_level, applicable_score_values,
				priority, status, create_id, update_id, create_time, update_time, del_flag
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
		`,
			scope, rowInstID, strings.TrimSpace(item.MaterialType), item.ParentGoalMaterialID,
			strings.TrimSpace(item.DomainCode), strings.TrimSpace(item.Domain), strings.TrimSpace(item.LongGoal), strings.TrimSpace(item.ShortGoal),
			strings.TrimSpace(item.CourseForm), item.AgeMinMonths, item.AgeMaxMonths, item.DifficultyLevel,
			strings.TrimSpace(item.ApplicableScoreValues), item.Priority, status, userID, userID,
		)
		if err != nil {
			return model.PEP3IEPGoalMaterial{}, err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return model.PEP3IEPGoalMaterial{}, err
		}
		item.ID = id
	}
	return repo.GetPEP3IEPGoalMaterial(ctx, instID, item.ID)
}

func (repo *Repository) GetPEP3IEPGoalMaterial(ctx context.Context, instID, id int64) (model.PEP3IEPGoalMaterial, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT id, library_scope, inst_id, material_type, parent_goal_material_id, domain_code, domain, long_goal, short_goal, course_form,
		       age_min_months, age_max_months, difficulty_level, applicable_score_values,
		       priority, status, create_time, update_time
		FROM pep3_iep_goal_material
		WHERE id = ? AND del_flag = 0
		  AND library_scope = 'institution' AND inst_id = ?
		LIMIT 1
	`, id, instID)
	return scanPEP3IEPGoalMaterial(row)
}

func (repo *Repository) DeletePEP3IEPGoalMaterial(ctx context.Context, instID, id int64) error {
	result, err := repo.db.ExecContext(ctx, `
		UPDATE pep3_iep_goal_material
		SET del_flag = 1, update_time = NOW()
		WHERE id = ? AND del_flag = 0
		  AND library_scope = 'institution' AND inst_id = ?
	`, id, instID)
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

func (repo *Repository) PagePEP3IEPTrainingMaterials(ctx context.Context, instID int64, query model.PEP3IEPTrainingMaterialPageQuery) (model.PageResult[model.PEP3IEPTrainingMaterial], error) {
	current, size := normalizeAssessmentPage(query.PageRequestModel.PageIndex, query.PageRequestModel.PageSize)
	where, args := pep3IEPTrainingMaterialWhere(instID, query.QueryModel)
	var total int
	if err := repo.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM pep3_iep_training_material WHERE `+where, args...).Scan(&total); err != nil {
		return model.PageResult[model.PEP3IEPTrainingMaterial]{}, err
	}
	listArgs := append(append([]any{}, args...), size, (current-1)*size)
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, library_scope, inst_id, goal_material_id, training_project, training_content,
		       priority, status, create_time, update_time
		FROM pep3_iep_training_material
		WHERE `+where+`
		ORDER BY FIELD(library_scope, 'institution', 'platform'), priority DESC, update_time DESC, id DESC
		LIMIT ? OFFSET ?
	`, listArgs...)
	if err != nil {
		return model.PageResult[model.PEP3IEPTrainingMaterial]{}, err
	}
	defer rows.Close()
	items := make([]model.PEP3IEPTrainingMaterial, 0, size)
	for rows.Next() {
		item, err := scanPEP3IEPTrainingMaterial(rows)
		if err != nil {
			return model.PageResult[model.PEP3IEPTrainingMaterial]{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return model.PageResult[model.PEP3IEPTrainingMaterial]{}, err
	}
	return model.PageResult[model.PEP3IEPTrainingMaterial]{Items: items, Total: total, Current: current, Size: size}, nil
}

func (repo *Repository) SavePEP3IEPTrainingMaterial(ctx context.Context, instID, userID int64, item model.PEP3IEPTrainingMaterial) (model.PEP3IEPTrainingMaterial, error) {
	scope, rowInstID := pep3IEPMaterialScopeAndInstID(instID, item.LibraryScope)
	status := pep3IEPMaterialStatus(item.Status)
	if item.ID > 0 {
		result, err := repo.db.ExecContext(ctx, `
			UPDATE pep3_iep_training_material
			SET library_scope = ?, inst_id = ?, goal_material_id = ?, training_project = ?,
			    training_content = ?, priority = ?, status = ?, update_id = ?, update_time = NOW()
			WHERE id = ? AND del_flag = 0
			  AND library_scope = 'institution' AND inst_id = ?
		`,
			scope, rowInstID, item.GoalMaterialID, strings.TrimSpace(item.TrainingProject), strings.TrimSpace(item.TrainingContent), item.Priority, status,
			userID, item.ID, instID,
		)
		if err != nil {
			return model.PEP3IEPTrainingMaterial{}, err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return model.PEP3IEPTrainingMaterial{}, err
		}
		if affected == 0 {
			return model.PEP3IEPTrainingMaterial{}, sql.ErrNoRows
		}
	} else {
		result, err := repo.db.ExecContext(ctx, `
			INSERT INTO pep3_iep_training_material (
				library_scope, inst_id, goal_material_id, training_project, training_content,
				priority, status, create_id, update_id, create_time, update_time, del_flag
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
		`,
			scope, rowInstID, item.GoalMaterialID, strings.TrimSpace(item.TrainingProject), strings.TrimSpace(item.TrainingContent),
			item.Priority, status, userID, userID,
		)
		if err != nil {
			return model.PEP3IEPTrainingMaterial{}, err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return model.PEP3IEPTrainingMaterial{}, err
		}
		item.ID = id
	}
	return repo.GetPEP3IEPTrainingMaterial(ctx, instID, item.ID)
}

func (repo *Repository) GetPEP3IEPTrainingMaterial(ctx context.Context, instID, id int64) (model.PEP3IEPTrainingMaterial, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT id, library_scope, inst_id, goal_material_id, training_project, training_content,
		       priority, status, create_time, update_time
		FROM pep3_iep_training_material
		WHERE id = ? AND del_flag = 0
		  AND library_scope = 'institution' AND inst_id = ?
		LIMIT 1
	`, id, instID)
	return scanPEP3IEPTrainingMaterial(row)
}

func (repo *Repository) DeletePEP3IEPTrainingMaterial(ctx context.Context, instID, id int64) error {
	result, err := repo.db.ExecContext(ctx, `
		UPDATE pep3_iep_training_material
		SET del_flag = 1, update_time = NOW()
		WHERE id = ? AND del_flag = 0
		  AND library_scope = 'institution' AND inst_id = ?
	`, id, instID)
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

func (repo *Repository) MatchPEP3IEPMaterialCandidates(ctx context.Context, instID int64, itemScores map[int]int) ([]model.PEP3IEPMaterialMatchCandidate, error) {
	if len(itemScores) == 0 {
		return nil, nil
	}
	itemNos := make([]int, 0, len(itemScores))
	for itemNo := range itemScores {
		if itemNo > 0 {
			itemNos = append(itemNos, itemNo)
		}
	}
	sort.Ints(itemNos)
	if len(itemNos) == 0 {
		return nil, nil
	}
	placeholder, inArgs := pep3IEPIntPlaceholders(itemNos)
	args := append([]any{instID}, inArgs...)
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, library_scope, inst_id, item_no, item_title, domain_code, domain,
		       score_value, score_label, score_description, result_meaning, generate_policy,
		       priority, ai_instruction, status, create_time, update_time
		FROM pep3_iep_item_option_rule
		WHERE del_flag = 0
		  AND status = 'active'
		  AND ((library_scope = 'platform' AND inst_id = 0) OR (library_scope = 'institution' AND inst_id = ?))
		  AND item_no IN (`+placeholder+`)
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rules := make([]model.PEP3IEPItemOptionRule, 0)
	for rows.Next() {
		rule, err := scanPEP3IEPItemOptionRule(rows)
		if err != nil {
			return nil, err
		}
		if score, ok := itemScores[rule.ItemNo]; ok && score == rule.ScoreValue {
			rules = append(rules, rule)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return nil, nil
	}
	ruleIDs := make([]int64, 0, len(rules))
	for _, rule := range rules {
		ruleIDs = append(ruleIDs, rule.ID)
	}
	goalByRule, err := repo.loadPEP3IEPRuleCandidateGoalMaterials(ctx, ruleIDs)
	if err != nil {
		return nil, err
	}
	candidates := make([]model.PEP3IEPMaterialMatchCandidate, 0, len(rules))
	for _, rule := range rules {
		goals := goalByRule[rule.ID]
		if len(goals) == 0 {
			candidates = append(candidates, pep3IEPMaterialCandidateFromRule(rule, model.PEP3IEPGoalMaterial{}))
			continue
		}
		for _, goal := range goals {
			candidates = append(candidates, pep3IEPMaterialCandidateFromRule(rule, goal))
		}
	}
	if err := repo.attachPEP3IEPTrainingMaterialsToCandidates(ctx, instID, candidates); err != nil {
		return nil, err
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if scoreRank(candidates[i].ScoreValue) != scoreRank(candidates[j].ScoreValue) {
			return scoreRank(candidates[i].ScoreValue) < scoreRank(candidates[j].ScoreValue)
		}
		if candidates[i].Priority != candidates[j].Priority {
			return candidates[i].Priority > candidates[j].Priority
		}
		if candidates[i].Domain != candidates[j].Domain {
			return candidates[i].Domain < candidates[j].Domain
		}
		if candidates[i].ItemNo != candidates[j].ItemNo {
			return candidates[i].ItemNo < candidates[j].ItemNo
		}
		return candidates[i].GoalMaterialID < candidates[j].GoalMaterialID
	})
	return candidates, nil
}

func (repo *Repository) attachPEP3IEPTrainingMaterialsToCandidates(ctx context.Context, instID int64, candidates []model.PEP3IEPMaterialMatchCandidate) error {
	if len(candidates) == 0 {
		return nil
	}
	goalSeen := map[int64]bool{}
	goalIDs := make([]int64, 0)
	for _, candidate := range candidates {
		if candidate.GoalMaterialID <= 0 || goalSeen[candidate.GoalMaterialID] {
			continue
		}
		goalSeen[candidate.GoalMaterialID] = true
		goalIDs = append(goalIDs, candidate.GoalMaterialID)
	}
	if len(goalIDs) == 0 {
		return nil
	}
	placeholder, inArgs := pep3IEPInt64Placeholders(goalIDs)
	args := append([]any{instID}, inArgs...)
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, library_scope, inst_id, goal_material_id, training_project, training_content,
		       priority, status, create_time, update_time
		FROM pep3_iep_training_material
		WHERE del_flag = 0
		  AND status = 'active'
		  AND ((library_scope = 'platform' AND inst_id = 0) OR (library_scope = 'institution' AND inst_id = ?))
		  AND goal_material_id IN (`+placeholder+`)
		ORDER BY goal_material_id ASC, FIELD(library_scope, 'institution', 'platform'), priority DESC, id DESC
	`, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	byGoal := map[int64][]model.PEP3IEPTrainingMaterial{}
	for rows.Next() {
		item, err := scanPEP3IEPTrainingMaterial(rows)
		if err != nil {
			return err
		}
		byGoal[item.GoalMaterialID] = append(byGoal[item.GoalMaterialID], item)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for i := range candidates {
		candidates[i].TrainingMaterials = byGoal[candidates[i].GoalMaterialID]
	}
	return nil
}

func pep3IEPMaterialCandidateFromRule(rule model.PEP3IEPItemOptionRule, goal model.PEP3IEPGoalMaterial) model.PEP3IEPMaterialMatchCandidate {
	priority := rule.Priority + goal.Priority
	return model.PEP3IEPMaterialMatchCandidate{
		RuleID:           rule.ID,
		GoalMaterialID:   goal.ID,
		LibraryScope:     firstNonEmptyPEP3IEPString(rule.LibraryScope, goal.LibraryScope),
		ItemNo:           rule.ItemNo,
		ItemTitle:        rule.ItemTitle,
		DomainCode:       firstNonEmptyPEP3IEPString(goal.DomainCode, rule.DomainCode),
		Domain:           firstNonEmptyPEP3IEPString(goal.Domain, rule.Domain),
		ScoreValue:       rule.ScoreValue,
		ScoreLabel:       rule.ScoreLabel,
		ScoreDescription: rule.ScoreDescription,
		ResultMeaning:    rule.ResultMeaning,
		GeneratePolicy:   rule.GeneratePolicy,
		Priority:         priority,
		LongGoal:         goal.LongGoal,
		ShortGoal:        goal.ShortGoal,
		CourseForm:       goal.CourseForm,
		AIInstruction:    rule.AIInstruction,
	}
}

func scoreRank(score int) int {
	switch score {
	case 1:
		return 0
	case 0:
		return 1
	case 2:
		return 2
	default:
		return 3
	}
}

type pep3IEPRowScanner interface {
	Scan(dest ...any) error
}

func scanPEP3IEPItemOptionRule(scanner pep3IEPRowScanner) (model.PEP3IEPItemOptionRule, error) {
	var item model.PEP3IEPItemOptionRule
	var createTime, updateTime sql.NullTime
	err := scanner.Scan(
		&item.ID, &item.LibraryScope, &item.InstID, &item.ItemNo, &item.ItemTitle, &item.DomainCode, &item.Domain,
		&item.ScoreValue, &item.ScoreLabel, &item.ScoreDescription, &item.ResultMeaning, &item.GeneratePolicy,
		&item.Priority, &item.AIInstruction, &item.Status, &createTime, &updateTime,
	)
	item.CreatedTime = pep3IEPTimePtr(createTime)
	item.UpdatedTime = pep3IEPTimePtr(updateTime)
	return item, err
}

func scanPEP3IEPGoalMaterial(scanner pep3IEPRowScanner) (model.PEP3IEPGoalMaterial, error) {
	var item model.PEP3IEPGoalMaterial
	var createTime, updateTime sql.NullTime
	err := scanner.Scan(
		&item.ID, &item.LibraryScope, &item.InstID, &item.MaterialType, &item.ParentGoalMaterialID, &item.DomainCode, &item.Domain, &item.LongGoal, &item.ShortGoal, &item.CourseForm,
		&item.AgeMinMonths, &item.AgeMaxMonths, &item.DifficultyLevel, &item.ApplicableScoreValues,
		&item.Priority, &item.Status, &createTime, &updateTime,
	)
	item.CreatedTime = pep3IEPTimePtr(createTime)
	item.UpdatedTime = pep3IEPTimePtr(updateTime)
	return item, err
}

func scanPEP3IEPTrainingMaterial(scanner pep3IEPRowScanner) (model.PEP3IEPTrainingMaterial, error) {
	var item model.PEP3IEPTrainingMaterial
	var createTime, updateTime sql.NullTime
	err := scanner.Scan(
		&item.ID, &item.LibraryScope, &item.InstID, &item.GoalMaterialID, &item.TrainingProject, &item.TrainingContent,
		&item.Priority, &item.Status, &createTime, &updateTime,
	)
	item.CreatedTime = pep3IEPTimePtr(createTime)
	item.UpdatedTime = pep3IEPTimePtr(updateTime)
	return item, err
}

func (repo *Repository) attachPEP3IEPRuleGoalMaterials(ctx context.Context, rules []model.PEP3IEPItemOptionRule) error {
	if len(rules) == 0 {
		return nil
	}
	ruleIDs := make([]int64, 0, len(rules))
	for _, rule := range rules {
		ruleIDs = append(ruleIDs, rule.ID)
	}
	goalByRule, err := repo.loadPEP3IEPRuleGoalMaterials(ctx, ruleIDs)
	if err != nil {
		return err
	}
	for i := range rules {
		goals := goalByRule[rules[i].ID]
		rules[i].GoalMaterials = goals
		rules[i].GoalMaterialIDs = make([]int64, 0, len(goals))
		for _, goal := range goals {
			rules[i].GoalMaterialIDs = append(rules[i].GoalMaterialIDs, goal.ID)
		}
	}
	return nil
}

func (repo *Repository) loadPEP3IEPRuleGoalMaterials(ctx context.Context, ruleIDs []int64) (map[int64][]model.PEP3IEPGoalMaterial, error) {
	result := make(map[int64][]model.PEP3IEPGoalMaterial, len(ruleIDs))
	if len(ruleIDs) == 0 {
		return result, nil
	}
	placeholder, args := pep3IEPInt64Placeholders(ruleIDs)
	rows, err := repo.db.QueryContext(ctx, `
		SELECT rel.rule_id,
		       gm.id, gm.library_scope, gm.inst_id, gm.material_type, gm.parent_goal_material_id, gm.domain_code, gm.domain, gm.long_goal, gm.short_goal, gm.course_form,
		       gm.age_min_months, gm.age_max_months, gm.difficulty_level, gm.applicable_score_values,
		       gm.priority, gm.status, gm.create_time, gm.update_time
		FROM pep3_iep_item_rule_goal_rel rel
		JOIN pep3_iep_goal_material gm ON gm.id = rel.goal_material_id AND gm.del_flag = 0 AND gm.status = 'active'
		WHERE rel.del_flag = 0 AND rel.rule_id IN (`+placeholder+`)
		ORDER BY rel.rule_id ASC, gm.priority DESC, gm.id DESC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var ruleID int64
		var item model.PEP3IEPGoalMaterial
		var createTime, updateTime sql.NullTime
		if err := rows.Scan(
			&ruleID,
			&item.ID, &item.LibraryScope, &item.InstID, &item.MaterialType, &item.ParentGoalMaterialID, &item.DomainCode, &item.Domain, &item.LongGoal, &item.ShortGoal, &item.CourseForm,
			&item.AgeMinMonths, &item.AgeMaxMonths, &item.DifficultyLevel, &item.ApplicableScoreValues,
			&item.Priority, &item.Status, &createTime, &updateTime,
		); err != nil {
			return nil, err
		}
		item.CreatedTime = pep3IEPTimePtr(createTime)
		item.UpdatedTime = pep3IEPTimePtr(updateTime)
		result[ruleID] = append(result[ruleID], item)
	}
	return result, rows.Err()
}

func (repo *Repository) loadPEP3IEPRuleCandidateGoalMaterials(ctx context.Context, ruleIDs []int64) (map[int64][]model.PEP3IEPGoalMaterial, error) {
	directGoals, err := repo.loadPEP3IEPRuleGoalMaterials(ctx, ruleIDs)
	if err != nil {
		return nil, err
	}
	result := make(map[int64][]model.PEP3IEPGoalMaterial, len(directGoals))
	type parentRef struct {
		ruleID int64
		parent model.PEP3IEPGoalMaterial
	}
	parentRefs := make(map[int64][]parentRef)
	parentSeen := map[int64]bool{}
	parentIDs := make([]int64, 0)
	for ruleID, goals := range directGoals {
		for _, goal := range goals {
			if strings.TrimSpace(goal.MaterialType) == "long_term" && goal.ID > 0 {
				parentRefs[goal.ID] = append(parentRefs[goal.ID], parentRef{ruleID: ruleID, parent: goal})
				if !parentSeen[goal.ID] {
					parentSeen[goal.ID] = true
					parentIDs = append(parentIDs, goal.ID)
				}
				continue
			}
			result[ruleID] = append(result[ruleID], goal)
		}
	}
	if len(parentIDs) == 0 {
		return result, nil
	}
	placeholder, args := pep3IEPInt64Placeholders(parentIDs)
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, library_scope, inst_id, material_type, parent_goal_material_id, domain_code, domain, long_goal, short_goal, course_form,
		       age_min_months, age_max_months, difficulty_level, applicable_score_values,
		       priority, status, create_time, update_time
		FROM pep3_iep_goal_material
		WHERE del_flag = 0
		  AND status = 'active'
		  AND material_type = 'short_term'
		  AND parent_goal_material_id IN (`+placeholder+`)
		ORDER BY parent_goal_material_id ASC, priority DESC, id DESC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	childCountByParent := map[int64]int{}
	for rows.Next() {
		child, err := scanPEP3IEPGoalMaterial(rows)
		if err != nil {
			return nil, err
		}
		for _, ref := range parentRefs[child.ParentGoalMaterialID] {
			candidate := child
			candidate.LongGoal = firstNonEmptyPEP3IEPString(candidate.LongGoal, ref.parent.LongGoal)
			candidate.DomainCode = firstNonEmptyPEP3IEPString(candidate.DomainCode, ref.parent.DomainCode)
			candidate.Domain = firstNonEmptyPEP3IEPString(candidate.Domain, ref.parent.Domain)
			result[ref.ruleID] = append(result[ref.ruleID], candidate)
		}
		childCountByParent[child.ParentGoalMaterialID]++
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for parentID, refs := range parentRefs {
		if childCountByParent[parentID] > 0 {
			continue
		}
		for _, ref := range refs {
			if strings.TrimSpace(ref.parent.ShortGoal) != "" {
				result[ref.ruleID] = append(result[ref.ruleID], ref.parent)
			}
		}
	}
	return result, nil
}

func pep3IEPRuleWhere(instID int64, query model.PEP3IEPItemOptionRuleQuery) (string, []any) {
	where, args := pep3IEPVisibleScopeWhere(instID, query.LibraryScope)
	if query.ItemNo != nil {
		where = append(where, "item_no = ?")
		args = append(args, *query.ItemNo)
	}
	if query.ScoreValue != nil {
		where = append(where, "score_value = ?")
		args = append(args, *query.ScoreValue)
	}
	if value := strings.TrimSpace(query.DomainCode); value != "" {
		where = append(where, "domain_code = ?")
		args = append(args, value)
	}
	if value := strings.TrimSpace(query.Domain); value != "" {
		where = append(where, "domain LIKE ?")
		args = append(args, "%"+value+"%")
	}
	if value := strings.TrimSpace(query.Status); value != "" {
		where = append(where, "status = ?")
		args = append(args, value)
	}
	if value := strings.TrimSpace(query.Keyword); value != "" {
		where = append(where, "(item_title LIKE ? OR result_meaning LIKE ? OR ai_instruction LIKE ?)")
		like := "%" + value + "%"
		args = append(args, like, like, like)
	}
	return strings.Join(where, " AND "), args
}

func pep3IEPGoalMaterialWhere(instID int64, query model.PEP3IEPMaterialQuery) (string, []any) {
	where, args := pep3IEPMaterialBaseWhere(instID, query)
	if value := strings.TrimSpace(query.MaterialType); value != "" {
		where = append(where, "material_type = ?")
		args = append(args, value)
	}
	if query.ParentGoalMaterialID != nil {
		where = append(where, "parent_goal_material_id = ?")
		args = append(args, *query.ParentGoalMaterialID)
	}
	if value := strings.TrimSpace(query.Keyword); value != "" {
		where = append(where, "(long_goal LIKE ? OR short_goal LIKE ?)")
		like := "%" + value + "%"
		args = append(args, like, like)
	}
	return strings.Join(where, " AND "), args
}

func pep3IEPTrainingMaterialWhere(instID int64, query model.PEP3IEPMaterialQuery) (string, []any) {
	where, args := pep3IEPVisibleScopeWhere(instID, query.LibraryScope)
	if query.GoalMaterialID != nil {
		where = append(where, "goal_material_id = ?")
		args = append(args, *query.GoalMaterialID)
	}
	if value := strings.TrimSpace(query.Keyword); value != "" {
		where = append(where, "(training_project LIKE ? OR training_content LIKE ?)")
		like := "%" + value + "%"
		args = append(args, like, like)
	}
	if value := strings.TrimSpace(query.Status); value != "" {
		where = append(where, "status = ?")
		args = append(args, value)
	}
	return strings.Join(where, " AND "), args
}

func pep3IEPMaterialBaseWhere(instID int64, query model.PEP3IEPMaterialQuery) ([]string, []any) {
	where, args := pep3IEPVisibleScopeWhere(instID, query.LibraryScope)
	if value := strings.TrimSpace(query.DomainCode); value != "" {
		where = append(where, "domain_code = ?")
		args = append(args, value)
	}
	if value := strings.TrimSpace(query.Domain); value != "" {
		where = append(where, "domain LIKE ?")
		args = append(args, "%"+value+"%")
	}
	if value := strings.TrimSpace(query.CourseForm); value != "" {
		where = append(where, "course_form = ?")
		args = append(args, value)
	}
	if value := strings.TrimSpace(query.Status); value != "" {
		where = append(where, "status = ?")
		args = append(args, value)
	}
	return where, args
}

func pep3IEPVisibleScopeWhere(instID int64, scope string) ([]string, []any) {
	where := []string{"del_flag = 0"}
	args := []any{}
	switch strings.ToLower(strings.TrimSpace(scope)) {
	case "platform":
		where = append(where, "library_scope = 'platform'", "inst_id = 0")
	case "institution":
		where = append(where, "library_scope = 'institution'", "inst_id = ?")
		args = append(args, instID)
	default:
		where = append(where, "((library_scope = 'platform' AND inst_id = 0) OR (library_scope = 'institution' AND inst_id = ?))")
		args = append(args, instID)
	}
	return where, args
}

func pep3IEPMaterialScopeAndInstID(instID int64, scope string) (string, int64) {
	if strings.EqualFold(strings.TrimSpace(scope), "platform") {
		return "platform", 0
	}
	return "institution", instID
}

func pep3IEPMaterialStatus(status string) string {
	status = strings.TrimSpace(status)
	if status == "" {
		return "active"
	}
	return status
}

func pep3IEPTimePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time
	return &t
}

func pep3IEPIntPlaceholders(values []int) (string, []any) {
	args := make([]any, 0, len(values))
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, "?")
		args = append(args, value)
	}
	return strings.Join(parts, ","), args
}

func pep3IEPInt64Placeholders(values []int64) (string, []any) {
	args := make([]any, 0, len(values))
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, "?")
		args = append(args, value)
	}
	return strings.Join(parts, ","), args
}

func firstNonEmptyPEP3IEPString(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
