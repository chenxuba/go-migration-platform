package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

const (
	IEPPlanGenerationTaskStatusPending = "pending"
	IEPPlanGenerationTaskStatusRunning = "running"
	IEPPlanGenerationTaskStatusDone    = "done"
	IEPPlanGenerationTaskStatusFailed  = "failed"
)

type IEPPlanGenerationTaskEntity struct {
	TaskID         string
	InstID         int64
	UserID         int64
	RecordID       int64
	AssessmentType string
	DurationMonths int
	Status         string
	Message        string
	StreamText     string
	Usage          *model.DeepSeekUsageVO
	CostAmountCNY  float64
	Plan           *model.PEP3IEPPlanAIResult
	SavedPlan      *model.PEP3IEPPlanSavedVO
	Error          string
	CreatedBy      int64
	UpdatedBy      int64
	CreatedTime    *time.Time
	UpdatedTime    *time.Time
}

func ensureIEPPlanGenerationTaskTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS assessment_iep_generation_task (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			task_id VARCHAR(64) NOT NULL,
			inst_id BIGINT NOT NULL DEFAULT 0,
			user_id BIGINT NOT NULL DEFAULT 0,
			record_id BIGINT NOT NULL DEFAULT 0,
			assessment_type VARCHAR(16) NOT NULL DEFAULT '',
			duration_months INT NOT NULL DEFAULT 3,
			status VARCHAR(16) NOT NULL DEFAULT 'pending',
			message VARCHAR(255) NOT NULL DEFAULT '',
			stream_text LONGTEXT NULL,
			usage_json LONGTEXT NULL,
			cost_amount_cny DECIMAL(12,6) NOT NULL DEFAULT 0,
			plan_json LONGTEXT NULL,
			saved_plan_json LONGTEXT NULL,
			error_message VARCHAR(500) NOT NULL DEFAULT '',
			create_id BIGINT NOT NULL DEFAULT 0,
			update_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uk_assessment_iep_generation_task_task (task_id, del_flag),
			KEY idx_assessment_iep_generation_task_active (inst_id, user_id, record_id, assessment_type, status, update_time),
			KEY idx_assessment_iep_generation_task_record (inst_id, record_id, assessment_type, update_time)
		)
	`)
	if err != nil {
		return err
	}
	if err := ensureTableIndexExists(ctx, db, "assessment_iep_generation_task", "uk_assessment_iep_generation_task_task",
		`ALTER TABLE assessment_iep_generation_task ADD UNIQUE KEY uk_assessment_iep_generation_task_task (task_id, del_flag)`); err != nil {
		return err
	}
	if err := ensureTableIndexExists(ctx, db, "assessment_iep_generation_task", "idx_assessment_iep_generation_task_active",
		`ALTER TABLE assessment_iep_generation_task ADD KEY idx_assessment_iep_generation_task_active (inst_id, user_id, record_id, assessment_type, status, update_time)`); err != nil {
		return err
	}
	return ensureColumnsOnTable(ctx, db, "assessment_iep_generation_task", map[string]string{
		"usage_json":       "usage_json LONGTEXT NULL AFTER stream_text",
		"cost_amount_cny":  "cost_amount_cny DECIMAL(12,6) NOT NULL DEFAULT 0 AFTER usage_json",
	})
}

func (repo *Repository) CreateIEPPlanGenerationTask(ctx context.Context, entity IEPPlanGenerationTaskEntity) error {
	status := strings.TrimSpace(entity.Status)
	if status == "" {
		status = IEPPlanGenerationTaskStatusPending
	}
	message := strings.TrimSpace(entity.Message)
	if message == "" {
		message = "正在准备AI生成任务"
	}
	taskID := strings.TrimSpace(entity.TaskID)
	if taskID == "" {
		return fmt.Errorf("task id is required")
	}
	_, err := repo.db.ExecContext(ctx, `
		INSERT INTO assessment_iep_generation_task (
			task_id, inst_id, user_id, record_id, assessment_type, duration_months,
			status, message, stream_text, usage_json, cost_amount_cny, plan_json, saved_plan_json, error_message,
			create_id, update_id, create_time, update_time, del_flag
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, '', NULL, 0, NULL, NULL, '', ?, ?, NOW(), NOW(), 0)
	`,
		taskID,
		entity.InstID,
		entity.UserID,
		entity.RecordID,
		strings.TrimSpace(entity.AssessmentType),
		entity.DurationMonths,
		status,
		message,
		entity.CreatedBy,
		entity.UpdatedBy,
	)
	return err
}

func (repo *Repository) UpdateIEPPlanGenerationTask(ctx context.Context, entity IEPPlanGenerationTaskEntity) error {
	taskID := strings.TrimSpace(entity.TaskID)
	if taskID == "" {
		return fmt.Errorf("task id is required")
	}
	var (
		usageJSON     any
		planJSON      any
		savedPlanJSON any
	)
	if entity.Usage != nil {
		raw, err := json.Marshal(entity.Usage)
		if err != nil {
			return fmt.Errorf("marshal iep generation usage: %w", err)
		}
		usageJSON = string(raw)
	}
	if entity.Plan != nil {
		raw, err := json.Marshal(entity.Plan)
		if err != nil {
			return fmt.Errorf("marshal iep generation plan: %w", err)
		}
		planJSON = string(raw)
	}
	if entity.SavedPlan != nil {
		raw, err := json.Marshal(entity.SavedPlan)
		if err != nil {
			return fmt.Errorf("marshal iep generation saved plan: %w", err)
		}
		savedPlanJSON = string(raw)
	}
	_, err := repo.db.ExecContext(ctx, `
		UPDATE assessment_iep_generation_task
		SET duration_months = ?,
		    status = ?,
		    message = ?,
		    stream_text = ?,
		    usage_json = ?,
		    cost_amount_cny = ?,
		    plan_json = ?,
		    saved_plan_json = ?,
		    error_message = ?,
		    update_id = ?,
		    update_time = NOW(),
		    del_flag = 0
		WHERE task_id = ? AND del_flag = 0
	`,
		entity.DurationMonths,
		strings.TrimSpace(entity.Status),
		strings.TrimSpace(entity.Message),
		entity.StreamText,
		usageJSON,
		entity.CostAmountCNY,
		planJSON,
		savedPlanJSON,
		strings.TrimSpace(entity.Error),
		entity.UpdatedBy,
		taskID,
	)
	return err
}

func (repo *Repository) GetIEPPlanGenerationTaskByTaskID(ctx context.Context, taskID string) (IEPPlanGenerationTaskEntity, bool, error) {
	return repo.getIEPPlanGenerationTask(ctx, `
		SELECT task_id, inst_id, user_id, record_id, assessment_type, duration_months,
		       status, message, stream_text, usage_json, cost_amount_cny, plan_json, saved_plan_json, error_message,
		       create_time, update_time
		FROM assessment_iep_generation_task
		WHERE task_id = ? AND del_flag = 0
		LIMIT 1
	`, strings.TrimSpace(taskID))
}

func (repo *Repository) FindActiveIEPPlanGenerationTask(
	ctx context.Context,
	instID, userID, recordID int64,
	assessmentType string,
) (IEPPlanGenerationTaskEntity, bool, error) {
	return repo.getIEPPlanGenerationTask(ctx, `
		SELECT task_id, inst_id, user_id, record_id, assessment_type, duration_months,
		       status, message, stream_text, usage_json, cost_amount_cny, plan_json, saved_plan_json, error_message,
		       create_time, update_time
		FROM assessment_iep_generation_task
		WHERE inst_id = ?
		  AND user_id = ?
		  AND record_id = ?
		  AND assessment_type = ?
		  AND status IN ('pending', 'running')
		  AND del_flag = 0
		ORDER BY update_time DESC, id DESC
		LIMIT 1
	`, instID, userID, recordID, strings.TrimSpace(assessmentType))
}

func (repo *Repository) getIEPPlanGenerationTask(ctx context.Context, query string, args ...any) (IEPPlanGenerationTaskEntity, bool, error) {
	var (
		entity           IEPPlanGenerationTaskEntity
		streamText       sql.NullString
		usageJSON        sql.NullString
		costAmountCNY    sql.NullFloat64
		planJSON         sql.NullString
		savedPlanJSON    sql.NullString
		errorMessage     sql.NullString
		createdTime      sql.NullTime
		updatedTime      sql.NullTime
		assessmentType   sql.NullString
		message          sql.NullString
	)
	err := repo.db.QueryRowContext(ctx, query, args...).Scan(
		&entity.TaskID,
		&entity.InstID,
		&entity.UserID,
		&entity.RecordID,
		&assessmentType,
		&entity.DurationMonths,
		&entity.Status,
		&message,
		&streamText,
		&usageJSON,
		&costAmountCNY,
		&planJSON,
		&savedPlanJSON,
		&errorMessage,
		&createdTime,
		&updatedTime,
	)
	if err == sql.ErrNoRows {
		return IEPPlanGenerationTaskEntity{}, false, nil
	}
	if err != nil {
		return IEPPlanGenerationTaskEntity{}, false, err
	}
	entity.AssessmentType = strings.TrimSpace(assessmentType.String)
	entity.Message = strings.TrimSpace(message.String)
	entity.StreamText = streamText.String
	entity.CostAmountCNY = costAmountCNY.Float64
	entity.Error = strings.TrimSpace(errorMessage.String)
	if usageJSON.Valid && strings.TrimSpace(usageJSON.String) != "" {
		var usage model.DeepSeekUsageVO
		if err := json.Unmarshal([]byte(usageJSON.String), &usage); err != nil {
			return IEPPlanGenerationTaskEntity{}, false, fmt.Errorf("parse iep generation usage: %w", err)
		}
		entity.Usage = &usage
	}
	if planJSON.Valid && strings.TrimSpace(planJSON.String) != "" {
		var plan model.PEP3IEPPlanAIResult
		if err := json.Unmarshal([]byte(planJSON.String), &plan); err != nil {
			return IEPPlanGenerationTaskEntity{}, false, fmt.Errorf("parse iep generation plan: %w", err)
		}
		entity.Plan = &plan
	}
	if savedPlanJSON.Valid && strings.TrimSpace(savedPlanJSON.String) != "" {
		var saved model.PEP3IEPPlanSavedVO
		if err := json.Unmarshal([]byte(savedPlanJSON.String), &saved); err != nil {
			return IEPPlanGenerationTaskEntity{}, false, fmt.Errorf("parse iep generation saved plan: %w", err)
		}
		entity.SavedPlan = &saved
	}
	if createdTime.Valid {
		entity.CreatedTime = &createdTime.Time
	}
	if updatedTime.Valid {
		entity.UpdatedTime = &updatedTime.Time
	}
	return entity, true, nil
}
