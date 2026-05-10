package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type PEP3ExecutionPlanEntity struct {
	InstID           int64
	RecordID         int64
	DurationMonths   int
	PlanType         string
	TargetMonthIndex int
	TargetWeekIndex  int
	PlanJSON         json.RawMessage
	CreatedBy        int64
	UpdatedBy        int64
	UpdatedTime      *time.Time
}

func ensurePEP3ExecutionPlanTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS assessment_iep_execution_plan (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			inst_id BIGINT NOT NULL DEFAULT 0,
			record_id BIGINT NOT NULL DEFAULT 0,
			duration_months INT NOT NULL DEFAULT 3,
			plan_type VARCHAR(16) NOT NULL DEFAULT '',
			target_month_index INT NOT NULL DEFAULT 0,
			target_week_index INT NOT NULL DEFAULT 0,
			plan_json LONGTEXT NOT NULL,
			create_id BIGINT NOT NULL DEFAULT 0,
			update_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uk_assessment_iep_execution_plan_target (inst_id, record_id, duration_months, plan_type, target_month_index, target_week_index, del_flag),
			KEY idx_assessment_iep_execution_plan_record (inst_id, record_id, duration_months, plan_type)
		)
	`)
	if err != nil {
		return err
	}
	return ensureTableIndexExists(ctx, db, "assessment_iep_execution_plan", "uk_assessment_iep_execution_plan_target",
		`ALTER TABLE assessment_iep_execution_plan ADD UNIQUE KEY uk_assessment_iep_execution_plan_target (inst_id, record_id, duration_months, plan_type, target_month_index, target_week_index, del_flag)`)
}

func (repo *Repository) SavePEP3ExecutionPlan(ctx context.Context, entity PEP3ExecutionPlanEntity, plan any) error {
	raw, err := json.Marshal(plan)
	if err != nil {
		return fmt.Errorf("marshal execution plan: %w", err)
	}
	if entity.DurationMonths <= 0 {
		entity.DurationMonths = 3
	}
	_, err = repo.db.ExecContext(ctx, `
		INSERT INTO assessment_iep_execution_plan (
			inst_id, record_id, duration_months, plan_type, target_month_index, target_week_index, plan_json,
			create_id, update_id, create_time, update_time, del_flag
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
		ON DUPLICATE KEY UPDATE
			plan_json = VALUES(plan_json),
			update_id = VALUES(update_id),
			update_time = NOW(),
			del_flag = 0
	`,
		entity.InstID,
		entity.RecordID,
		entity.DurationMonths,
		entity.PlanType,
		entity.TargetMonthIndex,
		entity.TargetWeekIndex,
		string(raw),
		entity.CreatedBy,
		entity.UpdatedBy,
	)
	return err
}

func savePEP3ExecutionPlanTx(ctx context.Context, tx *sql.Tx, entity PEP3ExecutionPlanEntity) error {
	if entity.DurationMonths <= 0 {
		entity.DurationMonths = 3
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO assessment_iep_execution_plan (
			inst_id, record_id, duration_months, plan_type, target_month_index, target_week_index, plan_json,
			create_id, update_id, create_time, update_time, del_flag
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
		ON DUPLICATE KEY UPDATE
			plan_json = VALUES(plan_json),
			update_id = VALUES(update_id),
			update_time = NOW(),
			del_flag = 0
	`,
		entity.InstID,
		entity.RecordID,
		entity.DurationMonths,
		entity.PlanType,
		entity.TargetMonthIndex,
		entity.TargetWeekIndex,
		string(entity.PlanJSON),
		entity.CreatedBy,
		entity.UpdatedBy,
	)
	return err
}

func deletePEP3ExecutionPlansForDurationTx(ctx context.Context, tx *sql.Tx, instID, recordID int64, durationMonths int) error {
	if durationMonths <= 0 {
		durationMonths = 3
	}
	_, err := tx.ExecContext(ctx, `
		DELETE FROM assessment_iep_execution_plan
		WHERE inst_id = ?
		  AND record_id = ?
		  AND duration_months = ?
	`, instID, recordID, durationMonths)
	return err
}

func (repo *Repository) ListPEP3ExecutionPlans(ctx context.Context, instID, recordID int64, durationMonths int) ([]PEP3ExecutionPlanEntity, error) {
	if durationMonths <= 0 {
		durationMonths = 3
	}
	rows, err := repo.db.QueryContext(ctx, `
		SELECT inst_id, record_id, duration_months, plan_type, target_month_index, target_week_index, plan_json, update_time
		FROM assessment_iep_execution_plan
		WHERE inst_id = ? AND record_id = ? AND duration_months = ? AND del_flag = 0
		ORDER BY plan_type, target_month_index, target_week_index
	`, instID, recordID, durationMonths)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]PEP3ExecutionPlanEntity, 0)
	for rows.Next() {
		var (
			entity      PEP3ExecutionPlanEntity
			rawPlan     string
			updatedTime sql.NullTime
		)
		if err := rows.Scan(
			&entity.InstID,
			&entity.RecordID,
			&entity.DurationMonths,
			&entity.PlanType,
			&entity.TargetMonthIndex,
			&entity.TargetWeekIndex,
			&rawPlan,
			&updatedTime,
		); err != nil {
			return nil, err
		}
		entity.PlanJSON = json.RawMessage(rawPlan)
		if updatedTime.Valid {
			entity.UpdatedTime = &updatedTime.Time
		}
		result = append(result, entity)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (repo *Repository) DeletePEP3WeeklyExecutionPlansForMonth(ctx context.Context, instID, recordID int64, durationMonths, targetMonthIndex int) error {
	if durationMonths <= 0 {
		durationMonths = 3
	}
	_, err := repo.db.ExecContext(ctx, `
		DELETE FROM assessment_iep_execution_plan
		WHERE inst_id = ?
		  AND record_id = ?
		  AND duration_months = ?
		  AND plan_type = 'weekly'
		  AND target_month_index = ?
	`, instID, recordID, durationMonths, targetMonthIndex)
	return err
}

func (repo *Repository) DeletePEP3ExecutionPlansForDuration(ctx context.Context, instID, recordID int64, durationMonths int) error {
	if durationMonths <= 0 {
		durationMonths = 3
	}
	_, err := repo.db.ExecContext(ctx, `
		DELETE FROM assessment_iep_execution_plan
		WHERE inst_id = ?
		  AND record_id = ?
		  AND duration_months = ?
	`, instID, recordID, durationMonths)
	return err
}
