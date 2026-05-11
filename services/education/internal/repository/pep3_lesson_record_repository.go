package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go-migration-platform/services/education/internal/model"
)

type PEP3LessonRecordEntity struct {
	InstID           int64
	RecordID         int64
	StudentID        int64
	StudentName      string
	AssessmentCode   string
	AssessmentName   string
	DurationMonths   int
	TargetMonthIndex int
	TargetWeekIndex  int
	LessonDate       time.Time
	WeekDateIndex    int
	WeeklyRowIndex   int
	Project          string
	Content          string
	CompletionCode   string
	TeacherName      string
	CourseName       string
	CreatedBy        int64
	UpdatedBy        int64
}

func ensurePEP3LessonRecordTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS assessment_iep_lesson_record (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			inst_id BIGINT NOT NULL DEFAULT 0,
			record_id BIGINT NOT NULL DEFAULT 0,
			student_id BIGINT NOT NULL DEFAULT 0,
			student_name VARCHAR(64) NOT NULL DEFAULT '',
			assessment_code VARCHAR(64) NOT NULL DEFAULT '',
			assessment_name VARCHAR(128) NOT NULL DEFAULT '',
			duration_months INT NOT NULL DEFAULT 3,
			target_month_index INT NOT NULL DEFAULT 0,
			target_week_index INT NOT NULL DEFAULT 0,
			lesson_date DATE NOT NULL,
			week_date_index INT NOT NULL DEFAULT 0,
			weekly_row_index INT NOT NULL DEFAULT 0,
			project VARCHAR(255) NOT NULL DEFAULT '',
			content TEXT,
			completion_code VARCHAR(16) NOT NULL DEFAULT '',
			teacher_name VARCHAR(64) NOT NULL DEFAULT '',
			course_name VARCHAR(255) NOT NULL DEFAULT '',
			create_id BIGINT NOT NULL DEFAULT 0,
			update_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uk_assessment_iep_lesson_record_target (inst_id, record_id, duration_months, target_month_index, target_week_index, lesson_date, weekly_row_index, del_flag),
			KEY idx_assessment_iep_lesson_record_week (inst_id, record_id, duration_months, target_month_index, target_week_index, lesson_date),
			KEY idx_assessment_iep_lesson_record_student_date (inst_id, student_id, lesson_date)
		)
	`)
	if err != nil {
		return err
	}
	if err := ensureTableIndexExists(ctx, db, "assessment_iep_lesson_record", "uk_assessment_iep_lesson_record_target",
		`ALTER TABLE assessment_iep_lesson_record ADD UNIQUE KEY uk_assessment_iep_lesson_record_target (inst_id, record_id, duration_months, target_month_index, target_week_index, lesson_date, weekly_row_index, del_flag)`); err != nil {
		return err
	}
	if err := ensureTableIndexExists(ctx, db, "assessment_iep_lesson_record", "idx_assessment_iep_lesson_record_week",
		`ALTER TABLE assessment_iep_lesson_record ADD KEY idx_assessment_iep_lesson_record_week (inst_id, record_id, duration_months, target_month_index, target_week_index, lesson_date)`); err != nil {
		return err
	}
	return ensureTableIndexExists(ctx, db, "assessment_iep_lesson_record", "idx_assessment_iep_lesson_record_student_date",
		`ALTER TABLE assessment_iep_lesson_record ADD KEY idx_assessment_iep_lesson_record_student_date (inst_id, student_id, lesson_date)`)
}

func (repo *Repository) SavePEP3WeeklyExecutionPlanWithLessonRecords(ctx context.Context, entity PEP3ExecutionPlanEntity, plan model.PEP3WeeklyPlanAIResult, lessonRecords []PEP3LessonRecordEntity) error {
	raw, err := json.Marshal(plan)
	if err != nil {
		return fmt.Errorf("marshal weekly execution plan: %w", err)
	}
	entity.PlanJSON = raw
	if entity.DurationMonths <= 0 {
		entity.DurationMonths = 3
	}
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := savePEP3ExecutionPlanTx(ctx, tx, entity); err != nil {
		return err
	}
	if err := replacePEP3LessonRecordsForWeekTx(
		ctx,
		tx,
		entity.InstID,
		entity.RecordID,
		entity.DurationMonths,
		entity.TargetMonthIndex,
		entity.TargetWeekIndex,
		lessonRecords,
	); err != nil {
		return err
	}
	return tx.Commit()
}

func (repo *Repository) DeletePEP3LessonRecordsForMonth(ctx context.Context, instID, recordID int64, durationMonths, targetMonthIndex int) error {
	if durationMonths <= 0 {
		durationMonths = 3
	}
	_, err := repo.db.ExecContext(ctx, `
		DELETE FROM assessment_iep_lesson_record
		WHERE inst_id = ?
		  AND record_id = ?
		  AND duration_months = ?
		  AND target_month_index = ?
	`, instID, recordID, durationMonths, targetMonthIndex)
	return err
}

func replacePEP3LessonRecordsForWeekTx(ctx context.Context, tx *sql.Tx, instID, recordID int64, durationMonths, targetMonthIndex, targetWeekIndex int, records []PEP3LessonRecordEntity) error {
	if err := deletePEP3LessonRecordsForWeekTx(ctx, tx, instID, recordID, durationMonths, targetMonthIndex, targetWeekIndex); err != nil {
		return err
	}
	for _, record := range records {
		if err := insertPEP3LessonRecordTx(ctx, tx, record); err != nil {
			return err
		}
	}
	return nil
}

func deletePEP3LessonRecordsForWeekTx(ctx context.Context, tx *sql.Tx, instID, recordID int64, durationMonths, targetMonthIndex, targetWeekIndex int) error {
	if durationMonths <= 0 {
		durationMonths = 3
	}
	_, err := tx.ExecContext(ctx, `
		DELETE FROM assessment_iep_lesson_record
		WHERE inst_id = ?
		  AND record_id = ?
		  AND duration_months = ?
		  AND target_month_index = ?
		  AND target_week_index = ?
	`, instID, recordID, durationMonths, targetMonthIndex, targetWeekIndex)
	return err
}

func deletePEP3LessonRecordsForDurationTx(ctx context.Context, tx *sql.Tx, instID, recordID int64, durationMonths int) error {
	if durationMonths <= 0 {
		durationMonths = 3
	}
	_, err := tx.ExecContext(ctx, `
		DELETE FROM assessment_iep_lesson_record
		WHERE inst_id = ?
		  AND record_id = ?
		  AND duration_months = ?
	`, instID, recordID, durationMonths)
	return err
}

func insertPEP3LessonRecordTx(ctx context.Context, tx *sql.Tx, entity PEP3LessonRecordEntity) error {
	if entity.DurationMonths <= 0 {
		entity.DurationMonths = 3
	}
	lessonDate := entity.LessonDate
	if lessonDate.IsZero() {
		return nil
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO assessment_iep_lesson_record (
			inst_id, record_id, student_id, student_name, assessment_code, assessment_name,
			duration_months, target_month_index, target_week_index, lesson_date, week_date_index,
			weekly_row_index, project, content, completion_code, teacher_name, course_name,
			create_id, update_id, create_time, update_time, del_flag
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
		ON DUPLICATE KEY UPDATE
			student_id = VALUES(student_id),
			student_name = VALUES(student_name),
			assessment_code = VALUES(assessment_code),
			assessment_name = VALUES(assessment_name),
			week_date_index = VALUES(week_date_index),
			project = VALUES(project),
			content = VALUES(content),
			completion_code = VALUES(completion_code),
			teacher_name = VALUES(teacher_name),
			course_name = VALUES(course_name),
			update_id = VALUES(update_id),
			update_time = NOW(),
			del_flag = 0
	`,
		entity.InstID,
		entity.RecordID,
		entity.StudentID,
		entity.StudentName,
		entity.AssessmentCode,
		entity.AssessmentName,
		entity.DurationMonths,
		entity.TargetMonthIndex,
		entity.TargetWeekIndex,
		lessonDate.Format("2006-01-02"),
		entity.WeekDateIndex,
		entity.WeeklyRowIndex,
		entity.Project,
		entity.Content,
		entity.CompletionCode,
		entity.TeacherName,
		entity.CourseName,
		entity.CreatedBy,
		entity.UpdatedBy,
	)
	return err
}
