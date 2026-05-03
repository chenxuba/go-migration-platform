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

type PEP3IEPPlanEntity struct {
	InstID         int64
	RecordID       int64
	DurationMonths int
	Status         string
	Plan           model.PEP3IEPPlanAIResult
	CreatedBy      int64
	UpdatedBy      int64
	UpdatedTime    *time.Time
}

func ensurePEP3IEPPlanTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS assessment_iep_plan (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			inst_id BIGINT NOT NULL DEFAULT 0,
			record_id BIGINT NOT NULL DEFAULT 0,
			duration_months INT NOT NULL DEFAULT 3,
			status VARCHAR(32) NOT NULL DEFAULT 'draft',
			plan_json LONGTEXT NOT NULL,
			create_id BIGINT NOT NULL DEFAULT 0,
			update_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uk_assessment_iep_plan_record_duration (inst_id, record_id, duration_months, del_flag),
			KEY idx_assessment_iep_plan_status (inst_id, status, update_time),
			KEY idx_assessment_iep_plan_record_status (inst_id, record_id, status)
		)
	`)
	if err != nil {
		return err
	}
	return ensurePEP3IEPPlanDurationUniqueIndex(ctx, db)
}

func ensurePEP3IEPPlanDurationUniqueIndex(ctx context.Context, db *sql.DB) error {
	if err := dropTableIndexIfExists(ctx, db, "assessment_iep_plan", "uk_assessment_iep_plan_record"); err != nil {
		return err
	}
	return ensureTableIndexExists(ctx, db, "assessment_iep_plan", "uk_assessment_iep_plan_record_duration",
		`ALTER TABLE assessment_iep_plan ADD UNIQUE KEY uk_assessment_iep_plan_record_duration (inst_id, record_id, duration_months, del_flag)`)
}

func (repo *Repository) SavePEP3IEPPlan(ctx context.Context, entity PEP3IEPPlanEntity) error {
	raw, err := json.Marshal(entity.Plan)
	if err != nil {
		return fmt.Errorf("marshal iep plan: %w", err)
	}
	status := strings.TrimSpace(entity.Status)
	if status == "" {
		status = "draft"
	}
	if entity.DurationMonths <= 0 {
		entity.DurationMonths = 3
	}
	_, err = repo.db.ExecContext(ctx, `
		INSERT INTO assessment_iep_plan (
			inst_id, record_id, duration_months, status, plan_json,
			create_id, update_id, create_time, update_time, del_flag
		) VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
		ON DUPLICATE KEY UPDATE
			duration_months = VALUES(duration_months),
			status = VALUES(status),
			plan_json = VALUES(plan_json),
			update_id = VALUES(update_id),
			update_time = NOW(),
			del_flag = 0
	`,
		entity.InstID,
		entity.RecordID,
		entity.DurationMonths,
		status,
		string(raw),
		entity.CreatedBy,
		entity.UpdatedBy,
	)
	return err
}

func (repo *Repository) GetPEP3IEPPlan(ctx context.Context, instID, recordID int64, durationMonths int) (PEP3IEPPlanEntity, bool, error) {
	var (
		entity      PEP3IEPPlanEntity
		rawPlan     string
		updatedTime sql.NullTime
	)
	if durationMonths <= 0 {
		durationMonths = 3
	}
	err := repo.db.QueryRowContext(ctx, `
		SELECT inst_id, record_id, duration_months, status, plan_json, update_time
		FROM assessment_iep_plan
		WHERE inst_id = ? AND record_id = ? AND duration_months = ? AND del_flag = 0
		LIMIT 1
	`, instID, recordID, durationMonths).Scan(
		&entity.InstID,
		&entity.RecordID,
		&entity.DurationMonths,
		&entity.Status,
		&rawPlan,
		&updatedTime,
	)
	if err == sql.ErrNoRows {
		return PEP3IEPPlanEntity{}, false, nil
	}
	if err != nil {
		return PEP3IEPPlanEntity{}, false, err
	}
	if err := json.Unmarshal([]byte(rawPlan), &entity.Plan); err != nil {
		return PEP3IEPPlanEntity{}, false, fmt.Errorf("parse iep plan: %w", err)
	}
	if updatedTime.Valid {
		entity.UpdatedTime = &updatedTime.Time
	}
	return entity, true, nil
}

func (repo *Repository) ListRecentPublishedRehabRecordRows(ctx context.Context, instID, studentID int64, limit int) ([]RehabRecordWordExportRow, error) {
	if limit <= 0 {
		limit = 12
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			CAST(str.id AS CHAR),
			CAST(str.student_id AS CHAR),
			str.student_name,
			CASE
				WHEN LENGTH(TRIM(str.one_to_one_name)) > 0 THEN str.one_to_one_name
				WHEN LENGTH(TRIM(str.class_name)) > 0 THEN str.class_name
				ELSE str.lesson_name
			END AS source_name,
			str.lesson_name,
			str.main_teacher_name,
			CAST(IFNULL(str.assistant_teacher_names_json, JSON_ARRAY()) AS CHAR),
			str.classroom_name,
			DATE_FORMAT(str.start_time, '%Y-%m-%d %H:%i:%s'),
			DATE_FORMAT(str.end_time, '%Y-%m-%d %H:%i:%s'),
			(
				SELECT IFNULL(srr.published_content_json, '')
				FROM student_rehab_record srr
				WHERE srr.inst_id = str.inst_id
				  AND srr.student_teaching_record_id = str.id
				  AND srr.del_flag = 0
				  AND LENGTH(TRIM(IFNULL(srr.published_content_json, ''))) > 0
				ORDER BY srr.id DESC
				LIMIT 1
			) AS published_content_json,
			(
				SELECT s.stu_sex
				FROM inst_student s
				WHERE s.inst_id = str.inst_id
				  AND s.id = str.student_id
				  AND s.del_flag = 0
				LIMIT 1
			) AS stu_sex,
			(
				SELECT s.birthday
				FROM inst_student s
				WHERE s.inst_id = str.inst_id
				  AND s.id = str.student_id
				  AND s.del_flag = 0
				LIMIT 1
			) AS birthday
		FROM student_teaching_record str
		WHERE str.inst_id = ?
		  AND str.student_id = ?
		  AND str.del_flag = 0
		  AND `+publishedStudentRehabRecordExistsSQL("str")+`
		ORDER BY str.start_time DESC, str.id DESC
		LIMIT ?
	`, instID, studentID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]RehabRecordWordExportRow, 0, limit)
	for rows.Next() {
		var (
			item          RehabRecordWordExportRow
			rawAssistants string
			sex           sql.NullInt64
			birthday      sql.NullTime
		)
		if err := rows.Scan(
			&item.StudentTeachingRecordID,
			&item.StudentID,
			&item.StudentName,
			&item.SourceName,
			&item.LessonName,
			&item.TeacherName,
			&rawAssistants,
			&item.ClassRoomName,
			&item.StartTime,
			&item.EndTime,
			&item.PublishedContentJSON,
			&sex,
			&birthday,
		); err != nil {
			return nil, err
		}
		item.Assistants = normalizeJSONStringListText(rawAssistants)
		if sex.Valid {
			value := int(sex.Int64)
			item.Sex = &value
		}
		if birthday.Valid && birthday.Time.Year() > 1 {
			item.BirthDate = birthday.Time.Format("2006-01-02")
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
