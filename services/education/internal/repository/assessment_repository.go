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

type AssessmentRecordEntity struct {
	ID             int64
	InstID         int64
	StudentID      int64
	StudentName    string
	AssessmentCode string
	AssessmentName string
	ScaleVersion   string
	BirthDate      time.Time
	AssessmentDate time.Time
	AgeYears       int
	AgeMonths      int
	AgeDays        int
	NormAgeMonths  int
	ExaminerID     int64
	ExaminerName   string
	Input          any
	Result         any
	DataStatus     string
	Remark         string
	CreatedTime    *time.Time
	UpdatedTime    *time.Time
}

type AssessmentDraftEntity struct {
	ID                int64
	InstID            int64
	StudentID         int64
	StudentName       string
	AssessmentCode    string
	AssessmentName    string
	ScaleVersion      string
	BirthDate         *time.Time
	AssessmentDate    *time.Time
	ExaminerID        int64
	ExaminerName      string
	Input             any
	Progress          any
	AnsweredItemCount int
	RawScoreCount     int
	Status            string
	Remark            string
	CreatedBy         int64
	UpdatedBy         int64
}

func ensureAssessmentTables(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS assessment_record (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			inst_id BIGINT NOT NULL DEFAULT 0,
			student_id BIGINT NOT NULL DEFAULT 0,
			student_name VARCHAR(100) NOT NULL DEFAULT '',
			assessment_code VARCHAR(64) NOT NULL DEFAULT '',
			assessment_name VARCHAR(100) NOT NULL DEFAULT '',
			scale_version VARCHAR(64) NOT NULL DEFAULT '',
			birth_date DATE NULL,
			assessment_date DATE NULL,
			age_years INT NOT NULL DEFAULT 0,
			age_months INT NOT NULL DEFAULT 0,
			age_days INT NOT NULL DEFAULT 0,
			norm_age_months INT NOT NULL DEFAULT 0,
			examiner_id BIGINT NOT NULL DEFAULT 0,
			examiner_name VARCHAR(100) NOT NULL DEFAULT '',
			input_json LONGTEXT NOT NULL,
			result_json LONGTEXT NOT NULL,
			data_status VARCHAR(500) NOT NULL DEFAULT '',
			remark VARCHAR(1000) NOT NULL DEFAULT '',
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			KEY idx_assessment_record_inst_code_date (inst_id, assessment_code, assessment_date, id),
			KEY idx_assessment_record_inst_student (inst_id, student_id, assessment_date, id),
			KEY idx_assessment_record_created (inst_id, create_time, id)
		)
	`); err != nil {
		return err
	}
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS assessment_draft (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			inst_id BIGINT NOT NULL DEFAULT 0,
			student_id BIGINT NOT NULL DEFAULT 0,
			student_name VARCHAR(100) NOT NULL DEFAULT '',
			assessment_code VARCHAR(64) NOT NULL DEFAULT '',
			assessment_name VARCHAR(100) NOT NULL DEFAULT '',
			scale_version VARCHAR(64) NOT NULL DEFAULT '',
			birth_date DATE NULL,
			assessment_date DATE NULL,
			examiner_id BIGINT NOT NULL DEFAULT 0,
			examiner_name VARCHAR(100) NOT NULL DEFAULT '',
			input_json LONGTEXT NOT NULL,
			progress_json LONGTEXT NOT NULL,
			answered_item_count INT NOT NULL DEFAULT 0,
			raw_score_count INT NOT NULL DEFAULT 0,
			status VARCHAR(32) NOT NULL DEFAULT 'draft',
			submitted_record_id BIGINT NOT NULL DEFAULT 0,
			remark VARCHAR(1000) NOT NULL DEFAULT '',
			create_id BIGINT NOT NULL DEFAULT 0,
			update_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			KEY idx_assessment_draft_inst_code_date (inst_id, assessment_code, assessment_date, id),
			KEY idx_assessment_draft_inst_student (inst_id, student_id, update_time, id),
			KEY idx_assessment_draft_status (inst_id, assessment_code, status, update_time, id)
		)
	`)
	return err
}

func (repo *Repository) CreateAssessmentRecord(ctx context.Context, entity AssessmentRecordEntity) (int64, error) {
	inputRaw, err := json.Marshal(entity.Input)
	if err != nil {
		return 0, fmt.Errorf("marshal assessment input: %w", err)
	}
	resultRaw, err := json.Marshal(entity.Result)
	if err != nil {
		return 0, fmt.Errorf("marshal assessment result: %w", err)
	}

	result, err := repo.db.ExecContext(ctx, `
		INSERT INTO assessment_record (
			inst_id, student_id, student_name, assessment_code, assessment_name, scale_version,
			birth_date, assessment_date, age_years, age_months, age_days, norm_age_months,
			examiner_id, examiner_name, input_json, result_json, data_status, remark,
			create_time, update_time, del_flag
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
	`,
		entity.InstID,
		entity.StudentID,
		strings.TrimSpace(entity.StudentName),
		strings.TrimSpace(entity.AssessmentCode),
		strings.TrimSpace(entity.AssessmentName),
		strings.TrimSpace(entity.ScaleVersion),
		entity.BirthDate,
		entity.AssessmentDate,
		entity.AgeYears,
		entity.AgeMonths,
		entity.AgeDays,
		entity.NormAgeMonths,
		entity.ExaminerID,
		strings.TrimSpace(entity.ExaminerName),
		string(inputRaw),
		string(resultRaw),
		strings.TrimSpace(entity.DataStatus),
		strings.TrimSpace(entity.Remark),
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (repo *Repository) GetAssessmentRecord(ctx context.Context, instID, recordID int64) (model.AssessmentRecordDetailVO, error) {
	var (
		item      model.AssessmentRecordDetailVO
		birthDate sql.NullTime
		testDate  sql.NullTime
		createdAt sql.NullTime
		updatedAt sql.NullTime
		inputRaw  string
		resultRaw string
	)
	err := repo.db.QueryRowContext(ctx, `
		SELECT id, inst_id, student_id, student_name, assessment_code, assessment_name, scale_version,
		       birth_date, assessment_date, age_years, age_months, age_days, norm_age_months,
		       examiner_id, examiner_name, input_json, result_json, data_status, remark, create_time, update_time
		FROM assessment_record
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, recordID, instID).Scan(
		&item.ID,
		&item.InstID,
		&item.StudentID,
		&item.StudentName,
		&item.AssessmentCode,
		&item.AssessmentName,
		&item.ScaleVersion,
		&birthDate,
		&testDate,
		&item.AgeYears,
		&item.AgeMonths,
		&item.AgeDays,
		&item.NormAgeMonths,
		&item.ExaminerID,
		&item.ExaminerName,
		&inputRaw,
		&resultRaw,
		&item.DataStatus,
		&item.Remark,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	fillAssessmentRecordTimes(&item.AssessmentRecordSummaryVO, birthDate, testDate, createdAt, updatedAt)
	item.InputJSON = json.RawMessage(inputRaw)
	item.ResultJSON = json.RawMessage(resultRaw)
	return item, nil
}

func (repo *Repository) PageAssessmentRecords(ctx context.Context, instID int64, query model.AssessmentRecordQueryModel, current, size int) (model.PageResult[model.AssessmentRecordSummaryVO], error) {
	current, size = normalizeAssessmentPage(current, size)
	where := []string{"inst_id = ?", "del_flag = 0"}
	args := []any{instID}
	if code := strings.TrimSpace(query.AssessmentCode); code != "" {
		where = append(where, "assessment_code = ?")
		args = append(args, code)
	}
	if query.StudentID != nil && *query.StudentID > 0 {
		where = append(where, "student_id = ?")
		args = append(args, *query.StudentID)
	}
	if searchKey := strings.TrimSpace(query.SearchKey); searchKey != "" {
		like := "%" + searchKey + "%"
		where = append(where, "(student_name LIKE ? OR examiner_name LIKE ?)")
		args = append(args, like, like)
	}
	if begin := strings.TrimSpace(query.AssessmentDateBegin); begin != "" {
		where = append(where, "assessment_date >= ?")
		args = append(args, begin)
	}
	if end := strings.TrimSpace(query.AssessmentDateEnd); end != "" {
		where = append(where, "assessment_date <= ?")
		args = append(args, end)
	}

	whereSQL := strings.Join(where, " AND ")
	var total int
	if err := repo.db.QueryRowContext(ctx, "SELECT COUNT(1) FROM assessment_record WHERE "+whereSQL, args...).Scan(&total); err != nil {
		return model.PageResult[model.AssessmentRecordSummaryVO]{}, err
	}
	offset := (current - 1) * size
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, inst_id, student_id, student_name, assessment_code, assessment_name, scale_version,
		       birth_date, assessment_date, age_years, age_months, age_days, norm_age_months,
		       examiner_id, examiner_name, data_status, remark, create_time, update_time
		FROM assessment_record
		WHERE `+whereSQL+`
		ORDER BY assessment_date DESC, id DESC
		LIMIT ? OFFSET ?
	`, append(args, size, offset)...)
	if err != nil {
		return model.PageResult[model.AssessmentRecordSummaryVO]{}, err
	}
	defer rows.Close()

	items := make([]model.AssessmentRecordSummaryVO, 0, size)
	for rows.Next() {
		var (
			item      model.AssessmentRecordSummaryVO
			birthDate sql.NullTime
			testDate  sql.NullTime
			createdAt sql.NullTime
			updatedAt sql.NullTime
		)
		if err := rows.Scan(
			&item.ID,
			&item.InstID,
			&item.StudentID,
			&item.StudentName,
			&item.AssessmentCode,
			&item.AssessmentName,
			&item.ScaleVersion,
			&birthDate,
			&testDate,
			&item.AgeYears,
			&item.AgeMonths,
			&item.AgeDays,
			&item.NormAgeMonths,
			&item.ExaminerID,
			&item.ExaminerName,
			&item.DataStatus,
			&item.Remark,
			&createdAt,
			&updatedAt,
		); err != nil {
			return model.PageResult[model.AssessmentRecordSummaryVO]{}, err
		}
		fillAssessmentRecordTimes(&item, birthDate, testDate, createdAt, updatedAt)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return model.PageResult[model.AssessmentRecordSummaryVO]{}, err
	}
	return model.PageResult[model.AssessmentRecordSummaryVO]{
		Items:   items,
		Total:   total,
		Current: current,
		Size:    size,
	}, nil
}

func (repo *Repository) SaveAssessmentDraft(ctx context.Context, entity AssessmentDraftEntity) (int64, error) {
	inputRaw, err := json.Marshal(entity.Input)
	if err != nil {
		return 0, fmt.Errorf("marshal assessment draft input: %w", err)
	}
	progressRaw, err := json.Marshal(entity.Progress)
	if err != nil {
		return 0, fmt.Errorf("marshal assessment draft progress: %w", err)
	}
	status := strings.TrimSpace(entity.Status)
	if status == "" {
		status = "draft"
	}

	if entity.ID > 0 {
		result, err := repo.db.ExecContext(ctx, `
			UPDATE assessment_draft
			SET student_id = ?,
			    student_name = ?,
			    assessment_code = ?,
			    assessment_name = ?,
			    scale_version = ?,
			    birth_date = ?,
			    assessment_date = ?,
			    examiner_id = ?,
			    examiner_name = ?,
			    input_json = ?,
			    progress_json = ?,
			    answered_item_count = ?,
			    raw_score_count = ?,
			    status = ?,
			    submitted_record_id = 0,
			    remark = ?,
			    update_id = ?,
			    update_time = NOW()
			WHERE id = ? AND inst_id = ? AND del_flag = 0
		`,
			entity.StudentID,
			strings.TrimSpace(entity.StudentName),
			strings.TrimSpace(entity.AssessmentCode),
			strings.TrimSpace(entity.AssessmentName),
			strings.TrimSpace(entity.ScaleVersion),
			nullableAssessmentDate(entity.BirthDate),
			nullableAssessmentDate(entity.AssessmentDate),
			entity.ExaminerID,
			strings.TrimSpace(entity.ExaminerName),
			string(inputRaw),
			string(progressRaw),
			entity.AnsweredItemCount,
			entity.RawScoreCount,
			status,
			strings.TrimSpace(entity.Remark),
			entity.UpdatedBy,
			entity.ID,
			entity.InstID,
		)
		if err != nil {
			return 0, err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return 0, err
		}
		if affected == 0 {
			return 0, sql.ErrNoRows
		}
		return entity.ID, nil
	}

	result, err := repo.db.ExecContext(ctx, `
		INSERT INTO assessment_draft (
			inst_id, student_id, student_name, assessment_code, assessment_name, scale_version,
			birth_date, assessment_date, examiner_id, examiner_name, input_json, progress_json,
			answered_item_count, raw_score_count, status, submitted_record_id, remark, create_id, update_id,
			create_time, update_time, del_flag
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?, ?, NOW(), NOW(), 0)
	`,
		entity.InstID,
		entity.StudentID,
		strings.TrimSpace(entity.StudentName),
		strings.TrimSpace(entity.AssessmentCode),
		strings.TrimSpace(entity.AssessmentName),
		strings.TrimSpace(entity.ScaleVersion),
		nullableAssessmentDate(entity.BirthDate),
		nullableAssessmentDate(entity.AssessmentDate),
		entity.ExaminerID,
		strings.TrimSpace(entity.ExaminerName),
		string(inputRaw),
		string(progressRaw),
		entity.AnsweredItemCount,
		entity.RawScoreCount,
		status,
		strings.TrimSpace(entity.Remark),
		entity.CreatedBy,
		entity.UpdatedBy,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (repo *Repository) GetAssessmentDraft(ctx context.Context, instID, draftID int64) (model.AssessmentDraftDetailVO, error) {
	var (
		item        model.AssessmentDraftDetailVO
		birthDate   sql.NullTime
		testDate    sql.NullTime
		createdAt   sql.NullTime
		updatedAt   sql.NullTime
		inputRaw    string
		progressRaw string
	)
	err := repo.db.QueryRowContext(ctx, `
		SELECT id, inst_id, student_id, student_name, assessment_code, assessment_name, scale_version,
		       birth_date, assessment_date, examiner_id, examiner_name, input_json, progress_json,
		       answered_item_count, raw_score_count, status, submitted_record_id, remark, create_time, update_time
		FROM assessment_draft
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, draftID, instID).Scan(
		&item.ID,
		&item.InstID,
		&item.StudentID,
		&item.StudentName,
		&item.AssessmentCode,
		&item.AssessmentName,
		&item.ScaleVersion,
		&birthDate,
		&testDate,
		&item.ExaminerID,
		&item.ExaminerName,
		&inputRaw,
		&progressRaw,
		&item.AnsweredItemCount,
		&item.RawScoreCount,
		&item.Status,
		&item.SubmittedRecordID,
		&item.Remark,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	fillAssessmentDraftTimes(&item.AssessmentDraftSummaryVO, birthDate, testDate, createdAt, updatedAt)
	item.InputJSON = json.RawMessage(inputRaw)
	if err := json.Unmarshal([]byte(progressRaw), &item.Progress); err != nil {
		return model.AssessmentDraftDetailVO{}, fmt.Errorf("decode assessment draft progress: %w", err)
	}
	item.CompletionPercent = item.Progress.CompletionPercent
	return item, nil
}

func (repo *Repository) PageAssessmentDrafts(ctx context.Context, instID int64, query model.AssessmentDraftQueryModel, current, size int) (model.PageResult[model.AssessmentDraftSummaryVO], error) {
	current, size = normalizeAssessmentPage(current, size)
	where := []string{"inst_id = ?", "del_flag = 0"}
	args := []any{instID}
	if code := strings.TrimSpace(query.AssessmentCode); code != "" {
		where = append(where, "assessment_code = ?")
		args = append(args, code)
	}
	if status := strings.TrimSpace(query.Status); status != "" {
		where = append(where, "status = ?")
		args = append(args, status)
	}
	if query.StudentID != nil && *query.StudentID > 0 {
		where = append(where, "student_id = ?")
		args = append(args, *query.StudentID)
	}
	if searchKey := strings.TrimSpace(query.SearchKey); searchKey != "" {
		like := "%" + searchKey + "%"
		where = append(where, "(student_name LIKE ? OR examiner_name LIKE ?)")
		args = append(args, like, like)
	}
	if begin := strings.TrimSpace(query.AssessmentDateBegin); begin != "" {
		where = append(where, "assessment_date >= ?")
		args = append(args, begin)
	}
	if end := strings.TrimSpace(query.AssessmentDateEnd); end != "" {
		where = append(where, "assessment_date <= ?")
		args = append(args, end)
	}

	whereSQL := strings.Join(where, " AND ")
	var total int
	if err := repo.db.QueryRowContext(ctx, "SELECT COUNT(1) FROM assessment_draft WHERE "+whereSQL, args...).Scan(&total); err != nil {
		return model.PageResult[model.AssessmentDraftSummaryVO]{}, err
	}
	offset := (current - 1) * size
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, inst_id, student_id, student_name, assessment_code, assessment_name, scale_version,
		       birth_date, assessment_date, examiner_id, examiner_name, progress_json,
		       answered_item_count, raw_score_count, status, submitted_record_id, remark, create_time, update_time
		FROM assessment_draft
		WHERE `+whereSQL+`
		ORDER BY update_time DESC, id DESC
		LIMIT ? OFFSET ?
	`, append(args, size, offset)...)
	if err != nil {
		return model.PageResult[model.AssessmentDraftSummaryVO]{}, err
	}
	defer rows.Close()

	items := make([]model.AssessmentDraftSummaryVO, 0, size)
	for rows.Next() {
		var (
			item        model.AssessmentDraftSummaryVO
			birthDate   sql.NullTime
			testDate    sql.NullTime
			createdAt   sql.NullTime
			updatedAt   sql.NullTime
			progressRaw string
		)
		if err := rows.Scan(
			&item.ID,
			&item.InstID,
			&item.StudentID,
			&item.StudentName,
			&item.AssessmentCode,
			&item.AssessmentName,
			&item.ScaleVersion,
			&birthDate,
			&testDate,
			&item.ExaminerID,
			&item.ExaminerName,
			&progressRaw,
			&item.AnsweredItemCount,
			&item.RawScoreCount,
			&item.Status,
			&item.SubmittedRecordID,
			&item.Remark,
			&createdAt,
			&updatedAt,
		); err != nil {
			return model.PageResult[model.AssessmentDraftSummaryVO]{}, err
		}
		fillAssessmentDraftTimes(&item, birthDate, testDate, createdAt, updatedAt)
		if err := json.Unmarshal([]byte(progressRaw), &item.Progress); err != nil {
			return model.PageResult[model.AssessmentDraftSummaryVO]{}, fmt.Errorf("decode assessment draft progress: %w", err)
		}
		item.CompletionPercent = item.Progress.CompletionPercent
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return model.PageResult[model.AssessmentDraftSummaryVO]{}, err
	}
	return model.PageResult[model.AssessmentDraftSummaryVO]{
		Items:   items,
		Total:   total,
		Current: current,
		Size:    size,
	}, nil
}

func (repo *Repository) DeleteAssessmentDraft(ctx context.Context, instID, draftID, operatorID int64) (bool, error) {
	result, err := repo.db.ExecContext(ctx, `
		UPDATE assessment_draft
		SET del_flag = 1,
		    update_id = ?,
		    update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, operatorID, draftID, instID)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

func (repo *Repository) MarkAssessmentDraftSubmitted(ctx context.Context, instID, draftID, recordID, operatorID int64) (bool, error) {
	result, err := repo.db.ExecContext(ctx, `
		UPDATE assessment_draft
		SET status = 'submitted',
		    submitted_record_id = ?,
		    update_id = ?,
		    update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, recordID, operatorID, draftID, instID)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

func fillAssessmentRecordTimes(item *model.AssessmentRecordSummaryVO, birthDate, assessmentDate, createdAt, updatedAt sql.NullTime) {
	if birthDate.Valid {
		t := birthDate.Time
		item.BirthDate = &t
	}
	if assessmentDate.Valid {
		t := assessmentDate.Time
		item.AssessmentDate = &t
	}
	if createdAt.Valid {
		t := createdAt.Time
		item.CreatedTime = &t
	}
	if updatedAt.Valid {
		t := updatedAt.Time
		item.UpdatedTime = &t
	}
}

func fillAssessmentDraftTimes(item *model.AssessmentDraftSummaryVO, birthDate, assessmentDate, createdAt, updatedAt sql.NullTime) {
	if birthDate.Valid {
		t := birthDate.Time
		item.BirthDate = &t
	}
	if assessmentDate.Valid {
		t := assessmentDate.Time
		item.AssessmentDate = &t
	}
	if createdAt.Valid {
		t := createdAt.Time
		item.CreatedTime = &t
	}
	if updatedAt.Valid {
		t := updatedAt.Time
		item.UpdatedTime = &t
	}
}

func nullableAssessmentDate(value *time.Time) any {
	if value == nil || value.IsZero() {
		return nil
	}
	return *value
}

func normalizeAssessmentPage(current, size int) (int, int) {
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	return current, size
}
