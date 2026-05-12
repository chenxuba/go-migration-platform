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

type IEPExecutionGenerationTaskEntity struct {
	TaskID              string
	InstID              int64
	UserID              int64
	RecordID            int64
	AssessmentType      string
	DurationMonths      int
	PlanType            string
	TargetMonthIndex    int
	TargetWeekIndex     int
	RestWeekdays        []int
	Status              string
	Message             string
	StreamText          string
	Usage               *model.DeepSeekUsageVO
	CostAmountCNY       float64
	MonthlyPlan         *model.PEP3MonthlyPlanAIResult
	WeeklyPlan          *model.PEP3WeeklyPlanAIResult
	SavedExecutionPlans *model.PEP3ExecutionPlanSavedVO
	Error               string
	CreatedBy           int64
	UpdatedBy           int64
	CreatedTime         *time.Time
	UpdatedTime         *time.Time
}

func ensureIEPExecutionGenerationTaskTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS assessment_iep_execution_generation_task (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			task_id VARCHAR(64) NOT NULL,
			inst_id BIGINT NOT NULL DEFAULT 0,
			user_id BIGINT NOT NULL DEFAULT 0,
			record_id BIGINT NOT NULL DEFAULT 0,
			assessment_type VARCHAR(16) NOT NULL DEFAULT '',
			duration_months INT NOT NULL DEFAULT 3,
			plan_type VARCHAR(16) NOT NULL DEFAULT '',
			target_month_index INT NOT NULL DEFAULT 0,
			target_week_index INT NOT NULL DEFAULT 0,
			rest_weekdays_json VARCHAR(100) NOT NULL DEFAULT '[]',
			status VARCHAR(16) NOT NULL DEFAULT 'pending',
			message VARCHAR(255) NOT NULL DEFAULT '',
			stream_text LONGTEXT NULL,
			usage_json LONGTEXT NULL,
			cost_amount_cny DECIMAL(12,6) NOT NULL DEFAULT 0,
			monthly_plan_json LONGTEXT NULL,
			weekly_plan_json LONGTEXT NULL,
			saved_execution_plans_json LONGTEXT NULL,
			error_message VARCHAR(500) NOT NULL DEFAULT '',
			create_id BIGINT NOT NULL DEFAULT 0,
			update_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uk_assessment_iep_execution_generation_task_task (task_id, del_flag),
			KEY idx_assessment_iep_execution_generation_task_active (inst_id, user_id, record_id, assessment_type, plan_type, target_month_index, target_week_index, status, update_time)
		)
	`)
	if err != nil {
		return err
	}
	if err := ensureTableIndexExists(ctx, db, "assessment_iep_execution_generation_task", "uk_assessment_iep_execution_generation_task_task",
		`ALTER TABLE assessment_iep_execution_generation_task ADD UNIQUE KEY uk_assessment_iep_execution_generation_task_task (task_id, del_flag)`); err != nil {
		return err
	}
	return ensureTableIndexExists(ctx, db, "assessment_iep_execution_generation_task", "idx_assessment_iep_execution_generation_task_active",
		`ALTER TABLE assessment_iep_execution_generation_task ADD KEY idx_assessment_iep_execution_generation_task_active (inst_id, user_id, record_id, assessment_type, plan_type, target_month_index, target_week_index, status, update_time)`)
}

func (repo *Repository) CreateIEPExecutionGenerationTask(ctx context.Context, entity IEPExecutionGenerationTaskEntity) error {
	taskID := strings.TrimSpace(entity.TaskID)
	if taskID == "" {
		return fmt.Errorf("task id is required")
	}
	restWeekdaysJSON, err := json.Marshal(entity.RestWeekdays)
	if err != nil {
		return fmt.Errorf("marshal execution task rest weekdays: %w", err)
	}
	_, err = repo.db.ExecContext(ctx, `
		INSERT INTO assessment_iep_execution_generation_task (
			task_id, inst_id, user_id, record_id, assessment_type, duration_months,
			plan_type, target_month_index, target_week_index, rest_weekdays_json,
			status, message, stream_text, usage_json, cost_amount_cny,
			monthly_plan_json, weekly_plan_json, saved_execution_plans_json, error_message,
			create_id, update_id, create_time, update_time, del_flag
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, '', NULL, 0, NULL, NULL, NULL, '', ?, ?, NOW(), NOW(), 0)
	`,
		taskID,
		entity.InstID,
		entity.UserID,
		entity.RecordID,
		strings.TrimSpace(entity.AssessmentType),
		entity.DurationMonths,
		strings.TrimSpace(entity.PlanType),
		entity.TargetMonthIndex,
		entity.TargetWeekIndex,
		string(restWeekdaysJSON),
		strings.TrimSpace(entity.Status),
		strings.TrimSpace(entity.Message),
		entity.CreatedBy,
		entity.UpdatedBy,
	)
	return err
}

func (repo *Repository) UpdateIEPExecutionGenerationTask(ctx context.Context, entity IEPExecutionGenerationTaskEntity) error {
	taskID := strings.TrimSpace(entity.TaskID)
	if taskID == "" {
		return fmt.Errorf("task id is required")
	}
	var (
		usageJSON               any
		monthlyPlanJSON         any
		weeklyPlanJSON          any
		savedExecutionPlansJSON any
	)
	restWeekdaysJSON, err := json.Marshal(entity.RestWeekdays)
	if err != nil {
		return fmt.Errorf("marshal execution task rest weekdays: %w", err)
	}
	if entity.Usage != nil {
		raw, err := json.Marshal(entity.Usage)
		if err != nil {
			return fmt.Errorf("marshal execution task usage: %w", err)
		}
		usageJSON = string(raw)
	}
	if entity.MonthlyPlan != nil {
		raw, err := json.Marshal(entity.MonthlyPlan)
		if err != nil {
			return fmt.Errorf("marshal execution monthly plan: %w", err)
		}
		monthlyPlanJSON = string(raw)
	}
	if entity.WeeklyPlan != nil {
		raw, err := json.Marshal(entity.WeeklyPlan)
		if err != nil {
			return fmt.Errorf("marshal execution weekly plan: %w", err)
		}
		weeklyPlanJSON = string(raw)
	}
	if entity.SavedExecutionPlans != nil {
		raw, err := json.Marshal(entity.SavedExecutionPlans)
		if err != nil {
			return fmt.Errorf("marshal execution saved plans: %w", err)
		}
		savedExecutionPlansJSON = string(raw)
	}
	_, err = repo.db.ExecContext(ctx, `
		UPDATE assessment_iep_execution_generation_task
		SET duration_months = ?,
		    plan_type = ?,
		    target_month_index = ?,
		    target_week_index = ?,
		    rest_weekdays_json = ?,
		    status = ?,
		    message = ?,
		    stream_text = ?,
		    usage_json = ?,
		    cost_amount_cny = ?,
		    monthly_plan_json = ?,
		    weekly_plan_json = ?,
		    saved_execution_plans_json = ?,
		    error_message = ?,
		    update_id = ?,
		    update_time = NOW(),
		    del_flag = 0
		WHERE task_id = ? AND del_flag = 0
	`,
		entity.DurationMonths,
		strings.TrimSpace(entity.PlanType),
		entity.TargetMonthIndex,
		entity.TargetWeekIndex,
		string(restWeekdaysJSON),
		strings.TrimSpace(entity.Status),
		strings.TrimSpace(entity.Message),
		entity.StreamText,
		usageJSON,
		entity.CostAmountCNY,
		monthlyPlanJSON,
		weeklyPlanJSON,
		savedExecutionPlansJSON,
		strings.TrimSpace(entity.Error),
		entity.UpdatedBy,
		taskID,
	)
	return err
}

func (repo *Repository) GetIEPExecutionGenerationTaskByTaskID(ctx context.Context, taskID string) (IEPExecutionGenerationTaskEntity, bool, error) {
	return repo.getIEPExecutionGenerationTask(ctx, `
		SELECT task_id, inst_id, user_id, record_id, assessment_type, duration_months,
		       plan_type, target_month_index, target_week_index, rest_weekdays_json,
		       status, message, stream_text, usage_json, cost_amount_cny,
		       monthly_plan_json, weekly_plan_json, saved_execution_plans_json, error_message,
		       create_time, update_time
		FROM assessment_iep_execution_generation_task
		WHERE task_id = ? AND del_flag = 0
		LIMIT 1
	`, strings.TrimSpace(taskID))
}

func (repo *Repository) FindActiveIEPExecutionGenerationTask(
	ctx context.Context,
	instID, userID, recordID int64,
	assessmentType, planType string,
	targetMonthIndex, targetWeekIndex int,
) (IEPExecutionGenerationTaskEntity, bool, error) {
	return repo.getIEPExecutionGenerationTask(ctx, `
		SELECT task_id, inst_id, user_id, record_id, assessment_type, duration_months,
		       plan_type, target_month_index, target_week_index, rest_weekdays_json,
		       status, message, stream_text, usage_json, cost_amount_cny,
		       monthly_plan_json, weekly_plan_json, saved_execution_plans_json, error_message,
		       create_time, update_time
		FROM assessment_iep_execution_generation_task
		WHERE inst_id = ?
		  AND user_id = ?
		  AND record_id = ?
		  AND assessment_type = ?
		  AND plan_type = ?
		  AND target_month_index = ?
		  AND target_week_index = ?
		  AND status IN ('pending', 'running')
		  AND del_flag = 0
		ORDER BY update_time DESC, id DESC
		LIMIT 1
	`, instID, userID, recordID, strings.TrimSpace(assessmentType), strings.TrimSpace(planType), targetMonthIndex, targetWeekIndex)
}

func (repo *Repository) getIEPExecutionGenerationTask(ctx context.Context, query string, args ...any) (IEPExecutionGenerationTaskEntity, bool, error) {
	var (
		entity                   IEPExecutionGenerationTaskEntity
		assessmentType           sql.NullString
		planType                 sql.NullString
		restWeekdaysJSON         sql.NullString
		message                  sql.NullString
		streamText               sql.NullString
		usageJSON                sql.NullString
		costAmountCNY            sql.NullFloat64
		monthlyPlanJSON          sql.NullString
		weeklyPlanJSON           sql.NullString
		savedExecutionPlansJSON  sql.NullString
		errorMessage             sql.NullString
		createdTime              sql.NullTime
		updatedTime              sql.NullTime
	)
	err := repo.db.QueryRowContext(ctx, query, args...).Scan(
		&entity.TaskID,
		&entity.InstID,
		&entity.UserID,
		&entity.RecordID,
		&assessmentType,
		&entity.DurationMonths,
		&planType,
		&entity.TargetMonthIndex,
		&entity.TargetWeekIndex,
		&restWeekdaysJSON,
		&entity.Status,
		&message,
		&streamText,
		&usageJSON,
		&costAmountCNY,
		&monthlyPlanJSON,
		&weeklyPlanJSON,
		&savedExecutionPlansJSON,
		&errorMessage,
		&createdTime,
		&updatedTime,
	)
	if err == sql.ErrNoRows {
		return IEPExecutionGenerationTaskEntity{}, false, nil
	}
	if err != nil {
		return IEPExecutionGenerationTaskEntity{}, false, err
	}
	entity.AssessmentType = strings.TrimSpace(assessmentType.String)
	entity.PlanType = strings.TrimSpace(planType.String)
	entity.Message = strings.TrimSpace(message.String)
	entity.StreamText = streamText.String
	entity.CostAmountCNY = costAmountCNY.Float64
	entity.Error = strings.TrimSpace(errorMessage.String)
	if restWeekdaysJSON.Valid && strings.TrimSpace(restWeekdaysJSON.String) != "" {
		_ = json.Unmarshal([]byte(restWeekdaysJSON.String), &entity.RestWeekdays)
	}
	if usageJSON.Valid && strings.TrimSpace(usageJSON.String) != "" {
		var usage model.DeepSeekUsageVO
		if err := json.Unmarshal([]byte(usageJSON.String), &usage); err != nil {
			return IEPExecutionGenerationTaskEntity{}, false, fmt.Errorf("parse execution task usage: %w", err)
		}
		entity.Usage = &usage
	}
	if monthlyPlanJSON.Valid && strings.TrimSpace(monthlyPlanJSON.String) != "" {
		var plan model.PEP3MonthlyPlanAIResult
		if err := json.Unmarshal([]byte(monthlyPlanJSON.String), &plan); err != nil {
			return IEPExecutionGenerationTaskEntity{}, false, fmt.Errorf("parse execution monthly plan: %w", err)
		}
		entity.MonthlyPlan = &plan
	}
	if weeklyPlanJSON.Valid && strings.TrimSpace(weeklyPlanJSON.String) != "" {
		var plan model.PEP3WeeklyPlanAIResult
		if err := json.Unmarshal([]byte(weeklyPlanJSON.String), &plan); err != nil {
			return IEPExecutionGenerationTaskEntity{}, false, fmt.Errorf("parse execution weekly plan: %w", err)
		}
		entity.WeeklyPlan = &plan
	}
	if savedExecutionPlansJSON.Valid && strings.TrimSpace(savedExecutionPlansJSON.String) != "" {
		var saved model.PEP3ExecutionPlanSavedVO
		if err := json.Unmarshal([]byte(savedExecutionPlansJSON.String), &saved); err != nil {
			return IEPExecutionGenerationTaskEntity{}, false, fmt.Errorf("parse execution saved plans: %w", err)
		}
		entity.SavedExecutionPlans = &saved
	}
	if createdTime.Valid {
		entity.CreatedTime = &createdTime.Time
	}
	if updatedTime.Valid {
		entity.UpdatedTime = &updatedTime.Time
	}
	return entity, true, nil
}
