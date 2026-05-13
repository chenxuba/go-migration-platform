package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"go-migration-platform/services/platform/internal/model"
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

func (repo *Repository) PagePlatformPEP3IEPItemOptionRules(ctx context.Context, query model.PEP3IEPItemOptionRulePageQuery) (model.PageResult[model.PEP3IEPItemOptionRule], error) {
	current, size := normalizePlatformPEP3IEPPage(query.PageRequestModel.PageIndex, query.PageRequestModel.PageSize)
	where, args := platformPEP3IEPRuleWhere(query.QueryModel)
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
		ORDER BY priority DESC, item_no ASC, score_value ASC, update_time DESC, id DESC
		LIMIT ? OFFSET ?
	`, listArgs...)
	if err != nil {
		return model.PageResult[model.PEP3IEPItemOptionRule]{}, err
	}
	defer rows.Close()
	items := make([]model.PEP3IEPItemOptionRule, 0, size)
	for rows.Next() {
		item, err := scanPlatformPEP3IEPItemOptionRule(rows)
		if err != nil {
			return model.PageResult[model.PEP3IEPItemOptionRule]{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return model.PageResult[model.PEP3IEPItemOptionRule]{}, err
	}
	if err := repo.attachPlatformPEP3IEPRuleGoalMaterials(ctx, items); err != nil {
		return model.PageResult[model.PEP3IEPItemOptionRule]{}, err
	}
	return model.PageResult[model.PEP3IEPItemOptionRule]{Items: items, Total: total, Current: current, Size: size}, nil
}

func (repo *Repository) SavePlatformPEP3IEPItemOptionRule(ctx context.Context, userID int64, item model.PEP3IEPItemOptionRule) (model.PEP3IEPItemOptionRule, error) {
	item.LibraryScope = "platform"
	item.InstID = 0
	status := platformPEP3IEPMaterialStatus(item.Status)
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return model.PEP3IEPItemOptionRule{}, err
	}
	defer tx.Rollback()
	if item.ID > 0 {
		result, err := tx.ExecContext(ctx, `
			UPDATE pep3_iep_item_option_rule
			SET library_scope = 'platform', inst_id = 0, item_no = ?, item_title = ?, domain_code = ?, domain = ?,
			    score_value = ?, score_label = ?, score_description = ?, result_meaning = ?, generate_policy = ?,
			    priority = ?, ai_instruction = ?, status = ?, update_id = ?, update_time = NOW()
			WHERE id = ? AND library_scope = 'platform' AND inst_id = 0 AND del_flag = 0
		`,
			item.ItemNo, strings.TrimSpace(item.ItemTitle), strings.TrimSpace(item.DomainCode), strings.TrimSpace(item.Domain),
			item.ScoreValue, strings.TrimSpace(item.ScoreLabel), strings.TrimSpace(item.ScoreDescription), strings.TrimSpace(item.ResultMeaning), strings.TrimSpace(item.GeneratePolicy),
			item.Priority, strings.TrimSpace(item.AIInstruction), status, userID, item.ID,
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
			) VALUES ('platform', 0, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
		`,
			item.ItemNo, strings.TrimSpace(item.ItemTitle), strings.TrimSpace(item.DomainCode), strings.TrimSpace(item.Domain),
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
	if err := replacePlatformPEP3IEPRuleGoalRelsTx(ctx, tx, item.ID, item.GoalMaterialIDs); err != nil {
		return model.PEP3IEPItemOptionRule{}, err
	}
	if err := tx.Commit(); err != nil {
		return model.PEP3IEPItemOptionRule{}, err
	}
	return repo.GetPlatformPEP3IEPItemOptionRule(ctx, item.ID)
}

func (repo *Repository) GetPlatformPEP3IEPItemOptionRule(ctx context.Context, id int64) (model.PEP3IEPItemOptionRule, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT id, library_scope, inst_id, item_no, item_title, domain_code, domain,
		       score_value, score_label, score_description, result_meaning, generate_policy,
		       priority, ai_instruction, status, create_time, update_time
		FROM pep3_iep_item_option_rule
		WHERE id = ? AND library_scope = 'platform' AND inst_id = 0 AND del_flag = 0
		LIMIT 1
	`, id)
	item, err := scanPlatformPEP3IEPItemOptionRule(row)
	if err != nil {
		return model.PEP3IEPItemOptionRule{}, err
	}
	items := []model.PEP3IEPItemOptionRule{item}
	if err := repo.attachPlatformPEP3IEPRuleGoalMaterials(ctx, items); err != nil {
		return model.PEP3IEPItemOptionRule{}, err
	}
	return items[0], nil
}

func (repo *Repository) DeletePlatformPEP3IEPItemOptionRule(ctx context.Context, id int64) error {
	result, err := repo.db.ExecContext(ctx, `
		UPDATE pep3_iep_item_option_rule
		SET del_flag = 1, update_time = NOW()
		WHERE id = ? AND library_scope = 'platform' AND inst_id = 0 AND del_flag = 0
	`, id)
	return platformPEP3IEPDeleteResult(result, err)
}

func (repo *Repository) PagePlatformPEP3IEPGoalMaterials(ctx context.Context, query model.PEP3IEPGoalMaterialPageQuery) (model.PageResult[model.PEP3IEPGoalMaterial], error) {
	current, size := normalizePlatformPEP3IEPPage(query.PageRequestModel.PageIndex, query.PageRequestModel.PageSize)
	where, args := platformPEP3IEPGoalMaterialWhere(query.QueryModel)
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
		ORDER BY priority DESC, update_time DESC, id DESC
		LIMIT ? OFFSET ?
	`, listArgs...)
	if err != nil {
		return model.PageResult[model.PEP3IEPGoalMaterial]{}, err
	}
	defer rows.Close()
	items := make([]model.PEP3IEPGoalMaterial, 0, size)
	for rows.Next() {
		item, err := scanPlatformPEP3IEPGoalMaterial(rows)
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

func (repo *Repository) SavePlatformPEP3IEPGoalMaterial(ctx context.Context, userID int64, item model.PEP3IEPGoalMaterial) (model.PEP3IEPGoalMaterial, error) {
	item.LibraryScope = "platform"
	item.InstID = 0
	status := platformPEP3IEPMaterialStatus(item.Status)
	if item.ID > 0 {
		result, err := repo.db.ExecContext(ctx, `
			UPDATE pep3_iep_goal_material
			SET library_scope = 'platform', inst_id = 0, material_type = ?, parent_goal_material_id = ?,
			    domain_code = ?, domain = ?, long_goal = ?, short_goal = ?,
			    course_form = ?, age_min_months = ?, age_max_months = ?, difficulty_level = ?,
			    applicable_score_values = ?, priority = ?, status = ?, update_id = ?, update_time = NOW()
			WHERE id = ? AND library_scope = 'platform' AND inst_id = 0 AND del_flag = 0
		`,
			strings.TrimSpace(item.MaterialType), item.ParentGoalMaterialID,
			strings.TrimSpace(item.DomainCode), strings.TrimSpace(item.Domain), strings.TrimSpace(item.LongGoal), strings.TrimSpace(item.ShortGoal),
			strings.TrimSpace(item.CourseForm), item.AgeMinMonths, item.AgeMaxMonths, item.DifficultyLevel,
			strings.TrimSpace(item.ApplicableScoreValues), item.Priority, status, userID, item.ID,
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
			) VALUES ('platform', 0, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
		`,
			strings.TrimSpace(item.MaterialType), item.ParentGoalMaterialID,
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
	return repo.GetPlatformPEP3IEPGoalMaterial(ctx, item.ID)
}

func (repo *Repository) GetPlatformPEP3IEPGoalMaterial(ctx context.Context, id int64) (model.PEP3IEPGoalMaterial, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT id, library_scope, inst_id, material_type, parent_goal_material_id, domain_code, domain, long_goal, short_goal, course_form,
		       age_min_months, age_max_months, difficulty_level, applicable_score_values,
		       priority, status, create_time, update_time
		FROM pep3_iep_goal_material
		WHERE id = ? AND library_scope = 'platform' AND inst_id = 0 AND del_flag = 0
		LIMIT 1
	`, id)
	return scanPlatformPEP3IEPGoalMaterial(row)
}

func (repo *Repository) DeletePlatformPEP3IEPGoalMaterial(ctx context.Context, id int64) error {
	result, err := repo.db.ExecContext(ctx, `
		UPDATE pep3_iep_goal_material
		SET del_flag = 1, update_time = NOW()
		WHERE id = ? AND library_scope = 'platform' AND inst_id = 0 AND del_flag = 0
	`, id)
	return platformPEP3IEPDeleteResult(result, err)
}

func (repo *Repository) PagePlatformPEP3IEPTrainingMaterials(ctx context.Context, query model.PEP3IEPTrainingMaterialPageQuery) (model.PageResult[model.PEP3IEPTrainingMaterial], error) {
	current, size := normalizePlatformPEP3IEPPage(query.PageRequestModel.PageIndex, query.PageRequestModel.PageSize)
	where, args := platformPEP3IEPTrainingMaterialWhere(query.QueryModel)
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
		ORDER BY priority DESC, update_time DESC, id DESC
		LIMIT ? OFFSET ?
	`, listArgs...)
	if err != nil {
		return model.PageResult[model.PEP3IEPTrainingMaterial]{}, err
	}
	defer rows.Close()
	items := make([]model.PEP3IEPTrainingMaterial, 0, size)
	for rows.Next() {
		item, err := scanPlatformPEP3IEPTrainingMaterial(rows)
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

func (repo *Repository) SavePlatformPEP3IEPTrainingMaterial(ctx context.Context, userID int64, item model.PEP3IEPTrainingMaterial) (model.PEP3IEPTrainingMaterial, error) {
	item.LibraryScope = "platform"
	item.InstID = 0
	status := platformPEP3IEPMaterialStatus(item.Status)
	if item.ID > 0 {
		result, err := repo.db.ExecContext(ctx, `
			UPDATE pep3_iep_training_material
			SET library_scope = 'platform', inst_id = 0, goal_material_id = ?, training_project = ?,
			    training_content = ?, priority = ?, status = ?, update_id = ?, update_time = NOW()
			WHERE id = ? AND library_scope = 'platform' AND inst_id = 0 AND del_flag = 0
		`,
			item.GoalMaterialID, strings.TrimSpace(item.TrainingProject), strings.TrimSpace(item.TrainingContent),
			item.Priority, status, userID, item.ID,
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
			) VALUES ('platform', 0, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
		`,
			item.GoalMaterialID, strings.TrimSpace(item.TrainingProject), strings.TrimSpace(item.TrainingContent),
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
	return repo.GetPlatformPEP3IEPTrainingMaterial(ctx, item.ID)
}

func (repo *Repository) GetPlatformPEP3IEPTrainingMaterial(ctx context.Context, id int64) (model.PEP3IEPTrainingMaterial, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT id, library_scope, inst_id, goal_material_id, training_project, training_content,
		       priority, status, create_time, update_time
		FROM pep3_iep_training_material
		WHERE id = ? AND library_scope = 'platform' AND inst_id = 0 AND del_flag = 0
		LIMIT 1
	`, id)
	return scanPlatformPEP3IEPTrainingMaterial(row)
}

func (repo *Repository) DeletePlatformPEP3IEPTrainingMaterial(ctx context.Context, id int64) error {
	result, err := repo.db.ExecContext(ctx, `
		UPDATE pep3_iep_training_material
		SET del_flag = 1, update_time = NOW()
		WHERE id = ? AND library_scope = 'platform' AND inst_id = 0 AND del_flag = 0
	`, id)
	return platformPEP3IEPDeleteResult(result, err)
}

func replacePlatformPEP3IEPRuleGoalRelsTx(ctx context.Context, tx *sql.Tx, ruleID int64, goalMaterialIDs []int64) error {
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

func (repo *Repository) attachPlatformPEP3IEPRuleGoalMaterials(ctx context.Context, rules []model.PEP3IEPItemOptionRule) error {
	if len(rules) == 0 {
		return nil
	}
	ruleIDs := make([]int64, 0, len(rules))
	for _, rule := range rules {
		ruleIDs = append(ruleIDs, rule.ID)
	}
	placeholder, args := platformPEP3IEPInt64Placeholders(ruleIDs)
	rows, err := repo.db.QueryContext(ctx, `
		SELECT rel.rule_id,
		       gm.id, gm.library_scope, gm.inst_id, gm.material_type, gm.parent_goal_material_id, gm.domain_code, gm.domain, gm.long_goal, gm.short_goal, gm.course_form,
		       gm.age_min_months, gm.age_max_months, gm.difficulty_level, gm.applicable_score_values,
		       gm.priority, gm.status, gm.create_time, gm.update_time
		FROM pep3_iep_item_rule_goal_rel rel
		JOIN pep3_iep_goal_material gm ON gm.id = rel.goal_material_id AND gm.del_flag = 0
		WHERE rel.del_flag = 0
		  AND gm.library_scope = 'platform'
		  AND gm.inst_id = 0
		  AND rel.rule_id IN (`+placeholder+`)
		ORDER BY rel.rule_id ASC, gm.priority DESC, gm.id DESC
	`, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	goalByRule := make(map[int64][]model.PEP3IEPGoalMaterial, len(rules))
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
			return err
		}
		item.CreatedTime = platformPEP3IEPTimePtr(createTime)
		item.UpdatedTime = platformPEP3IEPTimePtr(updateTime)
		goalByRule[ruleID] = append(goalByRule[ruleID], item)
	}
	if err := rows.Err(); err != nil {
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

func platformPEP3IEPRuleWhere(query model.PEP3IEPItemOptionRuleQuery) (string, []any) {
	where := []string{"del_flag = 0", "library_scope = 'platform'", "inst_id = 0"}
	args := []any{}
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

func platformPEP3IEPGoalMaterialWhere(query model.PEP3IEPMaterialQuery) (string, []any) {
	where, args := platformPEP3IEPMaterialBaseWhere(query)
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

func platformPEP3IEPTrainingMaterialWhere(query model.PEP3IEPMaterialQuery) (string, []any) {
	where := []string{"del_flag = 0", "library_scope = 'platform'", "inst_id = 0"}
	args := []any{}
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

func platformPEP3IEPMaterialBaseWhere(query model.PEP3IEPMaterialQuery) ([]string, []any) {
	where := []string{"del_flag = 0", "library_scope = 'platform'", "inst_id = 0"}
	args := []any{}
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

type platformPEP3IEPRowScanner interface {
	Scan(dest ...any) error
}

func scanPlatformPEP3IEPItemOptionRule(scanner platformPEP3IEPRowScanner) (model.PEP3IEPItemOptionRule, error) {
	var item model.PEP3IEPItemOptionRule
	var createTime, updateTime sql.NullTime
	err := scanner.Scan(
		&item.ID, &item.LibraryScope, &item.InstID, &item.ItemNo, &item.ItemTitle, &item.DomainCode, &item.Domain,
		&item.ScoreValue, &item.ScoreLabel, &item.ScoreDescription, &item.ResultMeaning, &item.GeneratePolicy,
		&item.Priority, &item.AIInstruction, &item.Status, &createTime, &updateTime,
	)
	item.CreatedTime = platformPEP3IEPTimePtr(createTime)
	item.UpdatedTime = platformPEP3IEPTimePtr(updateTime)
	return item, err
}

func scanPlatformPEP3IEPGoalMaterial(scanner platformPEP3IEPRowScanner) (model.PEP3IEPGoalMaterial, error) {
	var item model.PEP3IEPGoalMaterial
	var createTime, updateTime sql.NullTime
	err := scanner.Scan(
		&item.ID, &item.LibraryScope, &item.InstID, &item.MaterialType, &item.ParentGoalMaterialID, &item.DomainCode, &item.Domain, &item.LongGoal, &item.ShortGoal, &item.CourseForm,
		&item.AgeMinMonths, &item.AgeMaxMonths, &item.DifficultyLevel, &item.ApplicableScoreValues,
		&item.Priority, &item.Status, &createTime, &updateTime,
	)
	item.CreatedTime = platformPEP3IEPTimePtr(createTime)
	item.UpdatedTime = platformPEP3IEPTimePtr(updateTime)
	return item, err
}

func scanPlatformPEP3IEPTrainingMaterial(scanner platformPEP3IEPRowScanner) (model.PEP3IEPTrainingMaterial, error) {
	var item model.PEP3IEPTrainingMaterial
	var createTime, updateTime sql.NullTime
	err := scanner.Scan(
		&item.ID, &item.LibraryScope, &item.InstID, &item.GoalMaterialID, &item.TrainingProject, &item.TrainingContent,
		&item.Priority, &item.Status, &createTime, &updateTime,
	)
	item.CreatedTime = platformPEP3IEPTimePtr(createTime)
	item.UpdatedTime = platformPEP3IEPTimePtr(updateTime)
	return item, err
}

func normalizePlatformPEP3IEPPage(current, size int) (int, int) {
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 20
	}
	if size > 200 {
		size = 200
	}
	return current, size
}

func platformPEP3IEPMaterialStatus(status string) string {
	status = strings.TrimSpace(status)
	if status == "" {
		return "active"
	}
	return status
}

func platformPEP3IEPDeleteResult(result sql.Result, err error) error {
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

func platformPEP3IEPTimePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time
	return &t
}

func platformPEP3IEPInt64Placeholders(values []int64) (string, []any) {
	args := make([]any, 0, len(values))
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, "?")
		args = append(args, value)
	}
	return strings.Join(parts, ","), args
}
