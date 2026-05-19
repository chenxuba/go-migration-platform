package repository

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
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
	ReuseOpenDraft    bool
}

type AssessmentCaregiverInviteEntity struct {
	ID                     int64
	Ticket                 string
	InstID                 int64
	DraftID                int64
	RecordID               int64
	WeChatURLLink          string
	MiniProgramCodeDataURL string
	ExpiresAt              time.Time
}

type AssessmentReportInterpretationEntity struct {
	ID             int64
	InstID         int64
	RecordID       int64
	AssessmentCode string
	SourceHash     string
	Interpretation model.ERXinReportInterpretationVO
	CreatedTime    *time.Time
	UpdatedTime    *time.Time
}

type AssessmentReportInterpretationJSONEntity struct {
	ID             int64
	InstID         int64
	RecordID       int64
	AssessmentCode string
	SourceHash     string
	Model          string
	GeneratedBy    string
	CreatedTime    *time.Time
	UpdatedTime    *time.Time
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
	if _, err := db.ExecContext(ctx, `
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
			KEY idx_assessment_draft_reuse_lookup (inst_id, student_id, assessment_code, assessment_date, submitted_record_id, del_flag, update_time, id),
			KEY idx_assessment_draft_status (inst_id, assessment_code, status, update_time, id)
		)
	`); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS assessment_draft_item_score (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			inst_id BIGINT NOT NULL DEFAULT 0,
			draft_id BIGINT NOT NULL DEFAULT 0,
			item_no INT NOT NULL DEFAULT 0,
			score INT NOT NULL DEFAULT 0,
			create_id BIGINT NOT NULL DEFAULT 0,
			update_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uk_assessment_draft_item_score (draft_id, item_no),
			KEY idx_assessment_draft_item_score_inst (inst_id, draft_id, item_no)
		)
	`); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS assessment_draft_raw_score (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			inst_id BIGINT NOT NULL DEFAULT 0,
			draft_id BIGINT NOT NULL DEFAULT 0,
			scale_code VARCHAR(32) NOT NULL DEFAULT '',
			raw_score INT NOT NULL DEFAULT 0,
			create_id BIGINT NOT NULL DEFAULT 0,
			update_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uk_assessment_draft_raw_score (draft_id, scale_code),
			KEY idx_assessment_draft_raw_score_inst (inst_id, draft_id, scale_code)
		)
	`); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS assessment_draft_item_record_value (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			inst_id BIGINT NOT NULL DEFAULT 0,
			draft_id BIGINT NOT NULL DEFAULT 0,
			item_no INT NOT NULL DEFAULT 0,
			field_key VARCHAR(100) NOT NULL DEFAULT '',
			value_json LONGTEXT NOT NULL,
			create_id BIGINT NOT NULL DEFAULT 0,
			update_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uk_assessment_draft_item_record_value (draft_id, item_no, field_key),
			KEY idx_assessment_draft_item_record_inst (inst_id, draft_id, item_no)
		)
	`); err != nil {
		return err
	}
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS assessment_caregiver_invite (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			ticket VARCHAR(64) NOT NULL DEFAULT '',
			inst_id BIGINT NOT NULL DEFAULT 0,
			draft_id BIGINT NOT NULL DEFAULT 0,
			record_id BIGINT NOT NULL DEFAULT 0,
			wechat_url_link VARCHAR(1024) NOT NULL DEFAULT '',
			mini_program_code_data_url LONGTEXT NULL,
			expires_at DATETIME NULL,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uk_assessment_caregiver_invite_ticket (ticket),
			KEY idx_assessment_caregiver_invite_draft (inst_id, draft_id, update_time),
			KEY idx_assessment_caregiver_invite_record (inst_id, record_id, update_time)
		)
	`)
	if err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS assessment_report_interpretation (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			inst_id BIGINT NOT NULL DEFAULT 0,
			record_id BIGINT NOT NULL DEFAULT 0,
			assessment_code VARCHAR(64) NOT NULL DEFAULT '',
			source_hash VARCHAR(64) NOT NULL DEFAULT '',
			content_json LONGTEXT NOT NULL,
			model VARCHAR(100) NOT NULL DEFAULT '',
			generated_by VARCHAR(32) NOT NULL DEFAULT '',
			create_id BIGINT NOT NULL DEFAULT 0,
			update_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uk_assessment_report_interpretation_record (inst_id, record_id, assessment_code),
			KEY idx_assessment_report_interpretation_update (inst_id, assessment_code, update_time)
		)
	`); err != nil {
		return err
	}
	return ensureColumnsOnTable(ctx, db, "assessment_caregiver_invite", map[string]string{
		"wechat_url_link":            "wechat_url_link VARCHAR(1024) NOT NULL DEFAULT '' AFTER record_id",
		"mini_program_code_data_url": "mini_program_code_data_url LONGTEXT NULL AFTER wechat_url_link",
	})
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
	return scanAssessmentRecordDetailRows(repo.db.QueryRowContext(ctx, `
		SELECT ar.id, ar.inst_id, ar.student_id, COALESCE(NULLIF(ar.student_name, ''), IFNULL(s.stu_name, '')),
		       IFNULL(s.stu_sex, -1), IFNULL(s.avatar_url, ''),
		       CASE
		           WHEN CHAR_LENGTH(IFNULL(s.mobile, '')) >= 7 THEN CONCAT(LEFT(s.mobile, 3), '****', RIGHT(s.mobile, 4))
		           ELSE IFNULL(s.mobile, '')
		       END,
		       ar.assessment_code, ar.assessment_name, IFNULL(sc.category, ''), ar.scale_version,
		       ar.birth_date, ar.assessment_date, ar.age_years, ar.age_months, ar.age_days, ar.norm_age_months,
		       (
		         SELECT COUNT(1)
		         FROM assessment_record ar_seq
		         WHERE ar_seq.inst_id = ar.inst_id
		           AND ar_seq.student_id = ar.student_id
		           AND CONVERT(ar_seq.assessment_code USING utf8mb4) COLLATE utf8mb4_unicode_ci = CONVERT(ar.assessment_code USING utf8mb4) COLLATE utf8mb4_unicode_ci
		           AND ar_seq.del_flag = 0
		           AND (
		             ar_seq.assessment_date < ar.assessment_date
		             OR (ar_seq.assessment_date = ar.assessment_date AND ar_seq.id <= ar.id)
		             OR (ar_seq.assessment_date IS NULL AND ar.assessment_date IS NULL AND ar_seq.id <= ar.id)
		           )
		       ) AS assessment_sequence,
		       ar.examiner_id, ar.examiner_name, ar.input_json, ar.result_json, ar.data_status, ar.remark, ar.create_time, ar.update_time
		FROM assessment_record ar
		LEFT JOIN inst_student s ON s.id = ar.student_id AND s.inst_id = ar.inst_id AND s.del_flag = 0
		LEFT JOIN (
			SELECT scale_code, MAX(category) AS category
			FROM sys_scale
			WHERE del_flag = 0
			GROUP BY scale_code
			) sc ON CONVERT(sc.scale_code USING utf8mb4) COLLATE utf8mb4_unicode_ci = CONVERT(ar.assessment_code USING utf8mb4) COLLATE utf8mb4_unicode_ci
		WHERE ar.id = ? AND ar.inst_id = ? AND ar.del_flag = 0
		LIMIT 1
	`, recordID, instID))
}

type assessmentRecordDetailScanner interface {
	Scan(dest ...any) error
}

func scanAssessmentRecordDetailRows(row assessmentRecordDetailScanner) (model.AssessmentRecordDetailVO, error) {
	var (
		item      model.AssessmentRecordDetailVO
		birthDate sql.NullTime
		testDate  sql.NullTime
		createdAt sql.NullTime
		updatedAt sql.NullTime
		inputRaw  string
		resultRaw string
		sex       int
	)
	err := row.Scan(
		&item.ID,
		&item.InstID,
		&item.StudentID,
		&item.StudentName,
		&sex,
		&item.StudentAvatar,
		&item.StudentPhone,
		&item.AssessmentCode,
		&item.AssessmentName,
		&item.ScaleCategory,
		&item.ScaleVersion,
		&birthDate,
		&testDate,
		&item.AgeYears,
		&item.AgeMonths,
		&item.AgeDays,
		&item.NormAgeMonths,
		&item.AssessmentSequence,
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
	item.StudentGender = scaleLibraryStudentGenderText(sex)
	item.StudentAvatar = scaleLibraryStudentAvatarURL(item.StudentAvatar, sex)
	fillAssessmentRecordTimes(&item.AssessmentRecordSummaryVO, birthDate, testDate, createdAt, updatedAt)
	item.InputJSON = json.RawMessage(inputRaw)
	item.ResultJSON = json.RawMessage(resultRaw)
	return item, nil
}

func (repo *Repository) ListAssessmentRecordsForStudentScale(ctx context.Context, instID, studentID int64, assessmentCode string, throughRecordID int64, limit int) ([]model.AssessmentRecordDetailVO, error) {
	assessmentCode = strings.TrimSpace(assessmentCode)
	if instID <= 0 || studentID <= 0 || assessmentCode == "" || throughRecordID <= 0 {
		return nil, nil
	}
	if limit <= 0 {
		limit = 3
	}
	rows, err := repo.db.QueryContext(ctx, `
		SELECT ar.id, ar.inst_id, ar.student_id, COALESCE(NULLIF(ar.student_name, ''), IFNULL(s.stu_name, '')),
		       IFNULL(s.stu_sex, -1), IFNULL(s.avatar_url, ''),
		       CASE
		           WHEN CHAR_LENGTH(IFNULL(s.mobile, '')) >= 7 THEN CONCAT(LEFT(s.mobile, 3), '****', RIGHT(s.mobile, 4))
		           ELSE IFNULL(s.mobile, '')
		       END,
		       ar.assessment_code, ar.assessment_name, IFNULL(sc.category, ''), ar.scale_version,
		       ar.birth_date, ar.assessment_date, ar.age_years, ar.age_months, ar.age_days, ar.norm_age_months,
		       0 AS assessment_sequence,
		       ar.examiner_id, ar.examiner_name, ar.input_json, ar.result_json, ar.data_status, ar.remark, ar.create_time, ar.update_time
		FROM assessment_record ar
		LEFT JOIN inst_student s ON s.id = ar.student_id AND s.inst_id = ar.inst_id AND s.del_flag = 0
		LEFT JOIN (
			SELECT scale_code, MAX(category) AS category
			FROM sys_scale
			WHERE del_flag = 0
			GROUP BY scale_code
			) sc ON CONVERT(sc.scale_code USING utf8mb4) COLLATE utf8mb4_unicode_ci = CONVERT(ar.assessment_code USING utf8mb4) COLLATE utf8mb4_unicode_ci
		WHERE ar.inst_id = ?
		  AND ar.student_id = ?
		  AND CONVERT(ar.assessment_code USING utf8mb4) COLLATE utf8mb4_unicode_ci = CONVERT(? USING utf8mb4) COLLATE utf8mb4_unicode_ci
		  AND ar.del_flag = 0
		  AND (
		    ar.assessment_date < (SELECT assessment_date FROM assessment_record WHERE id = ? AND inst_id = ? AND del_flag = 0)
		    OR (
		      ar.assessment_date = (SELECT assessment_date FROM assessment_record WHERE id = ? AND inst_id = ? AND del_flag = 0)
		      AND ar.id <= ?
		    )
		    OR (
		      ar.assessment_date IS NULL
		      AND (SELECT assessment_date FROM assessment_record WHERE id = ? AND inst_id = ? AND del_flag = 0) IS NULL
		      AND ar.id <= ?
		    )
		  )
		ORDER BY ar.assessment_date ASC, ar.id ASC
		LIMIT ?
	`, instID, studentID, assessmentCode, throughRecordID, instID, throughRecordID, instID, throughRecordID, throughRecordID, instID, throughRecordID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.AssessmentRecordDetailVO, 0, limit)
	for rows.Next() {
		item, err := scanAssessmentRecordDetailRows(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for index := range items {
		items[index].AssessmentSequence = index + 1
	}
	return items, nil
}

func (repo *Repository) UpdateAssessmentRecordInputAndResult(ctx context.Context, instID, recordID int64, input any, result any, scaleVersion string, ageYears, ageMonths, ageDays, normAgeMonths int, dataStatus string) error {
	inputRaw, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("marshal assessment input: %w", err)
	}
	resultRaw, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal assessment result: %w", err)
	}

	updateResult, err := repo.db.ExecContext(ctx, `
		UPDATE assessment_record
		SET input_json = ?,
		    result_json = ?,
		    scale_version = ?,
		    age_years = ?,
		    age_months = ?,
		    age_days = ?,
		    norm_age_months = ?,
		    data_status = ?,
		    update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`,
		string(inputRaw),
		string(resultRaw),
		strings.TrimSpace(scaleVersion),
		ageYears,
		ageMonths,
		ageDays,
		normAgeMonths,
		strings.TrimSpace(dataStatus),
		recordID,
		instID,
	)
	if err != nil {
		return err
	}
	affected, err := updateResult.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (repo *Repository) UpdateAssessmentRecordBaseInputAndResult(ctx context.Context, entity AssessmentRecordEntity) error {
	inputRaw, err := json.Marshal(entity.Input)
	if err != nil {
		return fmt.Errorf("marshal assessment input: %w", err)
	}
	resultRaw, err := json.Marshal(entity.Result)
	if err != nil {
		return fmt.Errorf("marshal assessment result: %w", err)
	}

	updateResult, err := repo.db.ExecContext(ctx, `
		UPDATE assessment_record
		SET student_id = ?,
		    student_name = ?,
		    assessment_name = ?,
		    scale_version = ?,
		    birth_date = ?,
		    assessment_date = ?,
		    age_years = ?,
		    age_months = ?,
		    age_days = ?,
		    norm_age_months = ?,
		    examiner_id = ?,
		    examiner_name = ?,
		    input_json = ?,
		    result_json = ?,
		    data_status = ?,
		    remark = ?,
		    update_time = NOW()
		WHERE id = ? AND inst_id = ? AND assessment_code = ? AND del_flag = 0
	`,
		entity.StudentID,
		strings.TrimSpace(entity.StudentName),
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
		entity.ID,
		entity.InstID,
		strings.TrimSpace(entity.AssessmentCode),
	)
	if err != nil {
		return err
	}
	affected, err := updateResult.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (repo *Repository) UpdateAssessmentRecordConfig(ctx context.Context, instID, recordID int64, examinerName string, assessmentDate time.Time) error {
	updateResult, err := repo.db.ExecContext(ctx, `
		UPDATE assessment_record
		SET examiner_name = ?,
		    assessment_date = ?,
		    update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`,
		strings.TrimSpace(examinerName),
		assessmentDate,
		recordID,
		instID,
	)
	if err != nil {
		return err
	}
	_, _ = updateResult.RowsAffected()
	return nil
}

func (repo *Repository) AssessmentRecordHasIEPPlan(ctx context.Context, instID, recordID int64) (bool, error) {
	var count int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(1)
		FROM assessment_iep_plan
		WHERE inst_id = ? AND record_id = ? AND del_flag = 0
	`, instID, recordID).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (repo *Repository) GetPadHomeAssessmentStats(ctx context.Context, instID int64, assessmentCode string, _ time.Time) (model.PadHomeAssessmentStats, error) {
	assessmentCode = strings.TrimSpace(assessmentCode)
	stats := model.PadHomeAssessmentStats{}

	assessmentCodeFilter := ""
	coverageArgs := []any{}
	if assessmentCode != "" {
		assessmentCodeFilter = " AND ar.assessment_code = ?"
		coverageArgs = append(coverageArgs, assessmentCode)
	}
	coverageArgs = append(coverageArgs, instID, model.InstStudentStatusEnrolled)
	if err := repo.db.QueryRowContext(ctx, `
		SELECT
			COUNT(1),
			COALESCE(SUM(CASE WHEN EXISTS (
				SELECT 1
				FROM assessment_record ar
				WHERE ar.inst_id = s.inst_id
				  AND ar.student_id = s.id
				  AND ar.del_flag = 0`+assessmentCodeFilter+`
			) THEN 1 ELSE 0 END), 0)
		FROM inst_student s
		WHERE s.inst_id = ?
		  AND s.del_flag = 0
		  AND s.student_status = ?
	`, coverageArgs...).Scan(&stats.EnrolledStudents, &stats.AssessedStudents); err != nil {
		return model.PadHomeAssessmentStats{}, err
	}
	stats.UnassessedStudents = stats.EnrolledStudents - stats.AssessedStudents
	if stats.UnassessedStudents < 0 {
		stats.UnassessedStudents = 0
	}

	draftWhere := []string{
		"ad.inst_id = ?",
		"ad.del_flag = 0",
		"ad.status <> 'submitted'",
		"ad.submitted_record_id = 0",
		"s.student_status = ?",
	}
	draftArgs := []any{instID, model.InstStudentStatusEnrolled}
	if assessmentCode != "" {
		draftWhere = append(draftWhere, "ad.assessment_code = ?")
		draftArgs = append(draftArgs, assessmentCode)
	}
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(1)
		FROM (
			SELECT ad.student_id, ad.assessment_code, MAX(ad.id) AS draft_id
			FROM assessment_draft ad
			JOIN inst_student s ON s.id = ad.student_id AND s.inst_id = ad.inst_id AND s.del_flag = 0
			WHERE `+strings.Join(draftWhere, " AND ")+`
			GROUP BY ad.student_id, ad.assessment_code
		) active_draft
	`, draftArgs...).Scan(&stats.InProgressDrafts); err != nil {
		return model.PadHomeAssessmentStats{}, err
	}

	recordWhere := []string{"ar.inst_id = ?", "ar.del_flag = 0", "s.student_status = ?"}
	recordArgs := []any{instID, model.InstStudentStatusEnrolled}
	if assessmentCode != "" {
		recordWhere = append(recordWhere, "ar.assessment_code = ?")
		recordArgs = append(recordArgs, assessmentCode)
	}
	planSQL := `
		SELECT
			inst_id,
			record_id,
			COUNT(1) AS plan_count,
			SUM(CASE WHEN status = 'confirmed' THEN 1 ELSE 0 END) AS confirmed_count
		FROM assessment_iep_plan
		WHERE inst_id = ? AND del_flag = 0
		GROUP BY inst_id, record_id
	`
	queryArgs := append([]any{instID}, recordArgs...)
	if err := repo.db.QueryRowContext(ctx, `
		SELECT
			COUNT(1),
			COALESCE(SUM(CASE WHEN IFNULL(aip.plan_count, 0) = 0 THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN IFNULL(aip.plan_count, 0) > 0 AND IFNULL(aip.confirmed_count, 0) = 0 THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN IFNULL(aip.confirmed_count, 0) > 0 THEN 1 ELSE 0 END), 0)
		FROM assessment_record ar
		JOIN inst_student s ON s.id = ar.student_id AND s.inst_id = ar.inst_id AND s.del_flag = 0
		LEFT JOIN (`+planSQL+`) aip ON aip.inst_id = ar.inst_id AND aip.record_id = ar.id
		WHERE `+strings.Join(recordWhere, " AND "),
		queryArgs...,
	).Scan(&stats.CompletedRecords, &stats.PendingIEP, &stats.DraftIEP, &stats.GeneratedIEP); err != nil {
		return model.PadHomeAssessmentStats{}, err
	}

	stats.Total = stats.EnrolledStudents
	if stats.Total > 0 {
		stats.CoverageRate = float64(stats.AssessedStudents) / float64(stats.Total)
	}
	stats.CompletionRate = stats.CoverageRate
	return stats, nil
}

type assessmentRecordQueryParts struct {
	whereParts      []string
	args            []any
	needStudentJoin bool
	needScaleJoin   bool
}

func buildAssessmentRecordQueryParts(instID int64, query model.AssessmentRecordQueryModel) assessmentRecordQueryParts {
	parts := assessmentRecordQueryParts{
		whereParts: []string{"ar.inst_id = ?", "ar.del_flag = 0"},
		args:       []any{instID},
	}
	if code := strings.TrimSpace(query.AssessmentCode); code != "" {
		parts.whereParts = append(parts.whereParts, "ar.assessment_code = ?")
		parts.args = append(parts.args, code)
	}
	if category := strings.TrimSpace(query.ScaleCategory); category != "" {
		parts.needScaleJoin = true
		parts.whereParts = append(parts.whereParts, "IFNULL(sc.category, '') = ?")
		parts.args = append(parts.args, category)
	}
	if query.StudentID != nil && *query.StudentID > 0 {
		parts.whereParts = append(parts.whereParts, "ar.student_id = ?")
		parts.args = append(parts.args, *query.StudentID)
	}
	if searchKey := strings.TrimSpace(query.SearchKey); searchKey != "" {
		parts.needStudentJoin = true
		like := "%" + searchKey + "%"
		parts.whereParts = append(parts.whereParts, "(ar.student_name LIKE ? OR s.stu_name LIKE ? OR s.mobile LIKE ? OR ar.examiner_name LIKE ?)")
		parts.args = append(parts.args, like, like, like, like)
	}
	if begin := strings.TrimSpace(query.AssessmentDateBegin); begin != "" {
		parts.whereParts = append(parts.whereParts, "ar.assessment_date >= ?")
		parts.args = append(parts.args, begin)
	}
	if end := strings.TrimSpace(query.AssessmentDateEnd); end != "" {
		parts.whereParts = append(parts.whereParts, "ar.assessment_date <= ?")
		parts.args = append(parts.args, end)
	}
	return parts
}

func assessmentRecordStudentJoinSQL() string {
	return "LEFT JOIN inst_student s ON s.id = ar.student_id AND s.inst_id = ar.inst_id AND s.del_flag = 0"
}

func assessmentRecordScaleCategoryJoinSQL() string {
	return `LEFT JOIN (
			SELECT scale_code, MAX(category) AS category
			FROM sys_scale
			WHERE del_flag = 0
			GROUP BY scale_code
		) sc ON CONVERT(sc.scale_code USING utf8mb4) COLLATE utf8mb4_unicode_ci = CONVERT(ar.assessment_code USING utf8mb4) COLLATE utf8mb4_unicode_ci`
}

func (repo *Repository) SummarizeAssessmentRecordCategories(ctx context.Context, instID int64, query model.AssessmentRecordQueryModel) (model.AssessmentRecordCategoryStatsVO, error) {
	parts := buildAssessmentRecordQueryParts(instID, query)
	joins := []string{assessmentRecordScaleCategoryJoinSQL()}
	if parts.needStudentJoin {
		joins = append([]string{assessmentRecordStudentJoinSQL()}, joins...)
	}
	rows, err := repo.db.QueryContext(ctx, `
		SELECT IFNULL(sc.category, '') AS category, COUNT(1) AS total
		FROM assessment_record ar
		`+strings.Join(joins, "\n		")+`
		WHERE `+strings.Join(parts.whereParts, " AND ")+`
		GROUP BY IFNULL(sc.category, '')
		ORDER BY IFNULL(sc.category, '') ASC
	`, parts.args...)
	if err != nil {
		return model.AssessmentRecordCategoryStatsVO{}, err
	}
	defer rows.Close()

	stats := model.AssessmentRecordCategoryStatsVO{
		CategoryCounts: map[string]int{},
	}
	for rows.Next() {
		var (
			category string
			total    int
		)
		if err := rows.Scan(&category, &total); err != nil {
			return model.AssessmentRecordCategoryStatsVO{}, err
		}
		stats.CategoryCounts[category] = total
		stats.Total += total
	}
	if err := rows.Err(); err != nil {
		return model.AssessmentRecordCategoryStatsVO{}, err
	}
	return stats, nil
}

func (repo *Repository) PageAssessmentRecords(ctx context.Context, instID int64, query model.AssessmentRecordQueryModel, current, size int) (model.PageResult[model.AssessmentRecordSummaryVO], error) {
	current, size = normalizeAssessmentPage(current, size)
	parts := buildAssessmentRecordQueryParts(instID, query)
	whereSQL := strings.Join(parts.whereParts, " AND ")
	var total int
	countJoins := make([]string, 0, 2)
	if parts.needStudentJoin {
		countJoins = append(countJoins, assessmentRecordStudentJoinSQL())
	}
	if parts.needScaleJoin {
		countJoins = append(countJoins, assessmentRecordScaleCategoryJoinSQL())
	}
	countFromSQL := "FROM assessment_record ar"
	if len(countJoins) > 0 {
		countFromSQL += "\n\t\t" + strings.Join(countJoins, "\n\t\t")
	}
	if err := repo.db.QueryRowContext(ctx, "SELECT COUNT(1) "+countFromSQL+" WHERE "+whereSQL, parts.args...).Scan(&total); err != nil {
		return model.PageResult[model.AssessmentRecordSummaryVO]{}, err
	}
	offset := (current - 1) * size
	pageFromSQL := `
		FROM assessment_record ar
		` + assessmentRecordStudentJoinSQL() + `
		` + assessmentRecordScaleCategoryJoinSQL() + `
		LEFT JOIN (
			SELECT
				inst_id,
				record_id,
				CASE
					WHEN SUM(CASE WHEN status = 'confirmed' THEN 1 ELSE 0 END) > 0 THEN 'confirmed'
					WHEN COUNT(1) > 0 THEN 'draft'
					ELSE ''
				END AS status
			FROM assessment_iep_plan
			WHERE inst_id = ? AND del_flag = 0
			GROUP BY inst_id, record_id
		) aip ON aip.inst_id = ar.inst_id AND aip.record_id = ar.id
	`
	pageArgs := append([]any{instID}, parts.args...)
	rows, err := repo.db.QueryContext(ctx, `
		SELECT ar.id, ar.inst_id, ar.student_id, COALESCE(NULLIF(ar.student_name, ''), IFNULL(s.stu_name, '')),
		       IFNULL(s.stu_sex, -1), IFNULL(s.avatar_url, ''),
	       CASE
	           WHEN CHAR_LENGTH(IFNULL(s.mobile, '')) >= 7 THEN CONCAT(LEFT(s.mobile, 3), '****', RIGHT(s.mobile, 4))
	           ELSE IFNULL(s.mobile, '')
	       END,
	       ar.assessment_code, ar.assessment_name, IFNULL(sc.category, ''), ar.scale_version,
	       ar.birth_date, ar.assessment_date, ar.age_years, ar.age_months, ar.age_days, ar.norm_age_months,
	       (
	         SELECT COUNT(1)
	         FROM assessment_record ar_seq
	         WHERE ar_seq.inst_id = ar.inst_id
	           AND ar_seq.student_id = ar.student_id
	           AND CONVERT(ar_seq.assessment_code USING utf8mb4) COLLATE utf8mb4_unicode_ci = CONVERT(ar.assessment_code USING utf8mb4) COLLATE utf8mb4_unicode_ci
	           AND ar_seq.del_flag = 0
	           AND (
	             ar_seq.assessment_date < ar.assessment_date
	             OR (ar_seq.assessment_date = ar.assessment_date AND ar_seq.id <= ar.id)
	             OR (ar_seq.assessment_date IS NULL AND ar.assessment_date IS NULL AND ar_seq.id <= ar.id)
	           )
	       ) AS assessment_sequence,
	       ar.examiner_id, ar.examiner_name, ar.remark, IFNULL(aip.status, ''), ar.create_time, ar.update_time
		`+pageFromSQL+`
		WHERE `+whereSQL+`
		ORDER BY ar.create_time DESC, ar.id DESC
		LIMIT ? OFFSET ?
	`, append(pageArgs, size, offset)...)
	if err != nil {
		return model.PageResult[model.AssessmentRecordSummaryVO]{}, err
	}
	defer rows.Close()

	items := make([]model.AssessmentRecordSummaryVO, 0, size)
	for rows.Next() {
		var (
			item      model.AssessmentRecordSummaryVO
			sex       int
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
			&sex,
			&item.StudentAvatar,
			&item.StudentPhone,
			&item.AssessmentCode,
			&item.AssessmentName,
			&item.ScaleCategory,
			&item.ScaleVersion,
			&birthDate,
			&testDate,
			&item.AgeYears,
			&item.AgeMonths,
			&item.AgeDays,
			&item.NormAgeMonths,
			&item.AssessmentSequence,
			&item.ExaminerID,
			&item.ExaminerName,
			&item.Remark,
			&item.IEPPlanStatus,
			&createdAt,
			&updatedAt,
		); err != nil {
			return model.PageResult[model.AssessmentRecordSummaryVO]{}, err
		}
		item.StudentGender = scaleLibraryStudentGenderText(sex)
		item.StudentAvatar = scaleLibraryStudentAvatarURL(item.StudentAvatar, sex)
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

func (repo *Repository) DeleteAssessmentRecord(ctx context.Context, instID, recordID int64) (bool, error) {
	result, err := repo.db.ExecContext(ctx, `
		UPDATE assessment_record
		SET del_flag = 1,
		    update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, recordID, instID)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

func (repo *Repository) GetAssessmentReportInterpretation(ctx context.Context, instID, recordID int64, assessmentCode string) (AssessmentReportInterpretationEntity, error) {
	var (
		entity    AssessmentReportInterpretationEntity
		raw       string
		createdAt sql.NullTime
		updatedAt sql.NullTime
	)
	err := repo.db.QueryRowContext(ctx, `
		SELECT id, inst_id, record_id, assessment_code, source_hash, content_json, create_time, update_time
		FROM assessment_report_interpretation
		WHERE inst_id = ? AND record_id = ? AND assessment_code = ? AND del_flag = 0
		LIMIT 1
	`, instID, recordID, strings.TrimSpace(assessmentCode)).Scan(
		&entity.ID,
		&entity.InstID,
		&entity.RecordID,
		&entity.AssessmentCode,
		&entity.SourceHash,
		&raw,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return AssessmentReportInterpretationEntity{}, err
	}
	if err := json.Unmarshal([]byte(raw), &entity.Interpretation); err != nil {
		return AssessmentReportInterpretationEntity{}, fmt.Errorf("decode assessment report interpretation: %w", err)
	}
	if createdAt.Valid {
		t := createdAt.Time
		entity.CreatedTime = &t
	}
	if updatedAt.Valid {
		t := updatedAt.Time
		entity.UpdatedTime = &t
	}
	return entity, nil
}

func (repo *Repository) UpsertAssessmentReportInterpretation(ctx context.Context, entity AssessmentReportInterpretationEntity, operatorID int64) error {
	raw, err := json.Marshal(entity.Interpretation)
	if err != nil {
		return fmt.Errorf("marshal assessment report interpretation: %w", err)
	}
	_, err = repo.db.ExecContext(ctx, `
		INSERT INTO assessment_report_interpretation (
			inst_id, record_id, assessment_code, source_hash, content_json, model, generated_by,
			create_id, update_id, create_time, update_time, del_flag
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
		ON DUPLICATE KEY UPDATE
			source_hash = VALUES(source_hash),
			content_json = VALUES(content_json),
			model = VALUES(model),
			generated_by = VALUES(generated_by),
			update_id = VALUES(update_id),
			update_time = NOW(),
			del_flag = 0
	`,
		entity.InstID,
		entity.RecordID,
		strings.TrimSpace(entity.AssessmentCode),
		strings.TrimSpace(entity.SourceHash),
		string(raw),
		strings.TrimSpace(entity.Interpretation.Model),
		strings.TrimSpace(entity.Interpretation.GeneratedBy),
		operatorID,
		operatorID,
	)
	return err
}

func (repo *Repository) GetAssessmentReportInterpretationJSON(ctx context.Context, instID, recordID int64, assessmentCode string, dest any) (AssessmentReportInterpretationJSONEntity, error) {
	var (
		entity    AssessmentReportInterpretationJSONEntity
		raw       string
		createdAt sql.NullTime
		updatedAt sql.NullTime
	)
	err := repo.db.QueryRowContext(ctx, `
		SELECT id, inst_id, record_id, assessment_code, source_hash, content_json, model, generated_by, create_time, update_time
		FROM assessment_report_interpretation
		WHERE inst_id = ? AND record_id = ? AND assessment_code = ? AND del_flag = 0
		LIMIT 1
	`, instID, recordID, strings.TrimSpace(assessmentCode)).Scan(
		&entity.ID,
		&entity.InstID,
		&entity.RecordID,
		&entity.AssessmentCode,
		&entity.SourceHash,
		&raw,
		&entity.Model,
		&entity.GeneratedBy,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return AssessmentReportInterpretationJSONEntity{}, err
	}
	if dest != nil {
		if err := json.Unmarshal([]byte(raw), dest); err != nil {
			return AssessmentReportInterpretationJSONEntity{}, fmt.Errorf("decode assessment report interpretation: %w", err)
		}
	}
	if createdAt.Valid {
		t := createdAt.Time
		entity.CreatedTime = &t
	}
	if updatedAt.Valid {
		t := updatedAt.Time
		entity.UpdatedTime = &t
	}
	return entity, nil
}

func (repo *Repository) UpsertAssessmentReportInterpretationJSON(ctx context.Context, entity AssessmentReportInterpretationJSONEntity, content any, operatorID int64) error {
	raw, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("marshal assessment report interpretation: %w", err)
	}
	_, err = repo.db.ExecContext(ctx, `
		INSERT INTO assessment_report_interpretation (
			inst_id, record_id, assessment_code, source_hash, content_json, model, generated_by,
			create_id, update_id, create_time, update_time, del_flag
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
		ON DUPLICATE KEY UPDATE
			source_hash = VALUES(source_hash),
			content_json = VALUES(content_json),
			model = VALUES(model),
			generated_by = VALUES(generated_by),
			update_id = VALUES(update_id),
			update_time = NOW(),
			del_flag = 0
	`,
		entity.InstID,
		entity.RecordID,
		strings.TrimSpace(entity.AssessmentCode),
		strings.TrimSpace(entity.SourceHash),
		string(raw),
		strings.TrimSpace(entity.Model),
		strings.TrimSpace(entity.GeneratedBy),
		operatorID,
		operatorID,
	)
	return err
}

func (repo *Repository) SoftDeleteAssessmentReportInterpretation(ctx context.Context, instID, recordID int64, assessmentCode string, operatorID int64) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE assessment_report_interpretation
		SET del_flag = 1,
		    update_id = ?,
		    update_time = NOW()
		WHERE inst_id = ? AND record_id = ? AND assessment_code = ? AND del_flag = 0
	`, operatorID, instID, recordID, strings.TrimSpace(assessmentCode))
	return err
}

func (repo *Repository) SaveAssessmentDraft(
	ctx context.Context,
	entity AssessmentDraftEntity,
	itemScores map[int]int,
	rawScores map[string]int,
	itemRecordValues map[int]map[string]any,
	operatorID int64,
) (int64, error) {
	if entity.ID <= 0 && entity.ReuseOpenDraft {
		conn, err := repo.db.Conn(ctx)
		if err != nil {
			return 0, err
		}
		defer conn.Close()
		releaseLock, err := acquireAssessmentDraftReuseLock(ctx, conn, assessmentDraftReuseLockKey(entity))
		if err != nil {
			return 0, err
		}
		defer releaseLock()
		return saveAssessmentDraftInTransaction(ctx, conn, entity, itemScores, rawScores, itemRecordValues, operatorID, true)
	}
	return saveAssessmentDraftInTransaction(ctx, repo.db, entity, itemScores, rawScores, itemRecordValues, operatorID, false)
}

type assessmentDraftTxStarter interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

func saveAssessmentDraftInTransaction(
	ctx context.Context,
	txStarter assessmentDraftTxStarter,
	entity AssessmentDraftEntity,
	itemScores map[int]int,
	rawScores map[string]int,
	itemRecordValues map[int]map[string]any,
	operatorID int64,
	reuseOpenDraft bool,
) (int64, error) {
	tx, err := txStarter.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	if entity.ID <= 0 && reuseOpenDraft {
		draftID, err := findReusableAssessmentDraftIDForUpdate(ctx, tx, entity)
		if err != nil && err != sql.ErrNoRows {
			return 0, err
		}
		if draftID > 0 {
			entity.ID = draftID
		}
	}
	draftID, err := saveAssessmentDraftWithExecutor(ctx, tx, entity)
	if err != nil {
		return 0, err
	}
	if err = replaceAssessmentDraftItemScores(ctx, tx, entity.InstID, draftID, itemScores, operatorID); err != nil {
		return 0, err
	}
	if err = replaceAssessmentDraftRawScores(ctx, tx, entity.InstID, draftID, rawScores, operatorID); err != nil {
		return 0, err
	}
	if err = replaceAssessmentDraftItemRecordValues(ctx, tx, entity.InstID, draftID, itemRecordValues, operatorID); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	committed = true
	return draftID, nil
}

func acquireAssessmentDraftReuseLock(ctx context.Context, conn *sql.Conn, lockKey string) (func(), error) {
	var locked sql.NullInt64
	if err := conn.QueryRowContext(ctx, `SELECT GET_LOCK(?, 5)`, lockKey).Scan(&locked); err != nil {
		return nil, err
	}
	if !locked.Valid || locked.Int64 != 1 {
		return nil, fmt.Errorf("assessment draft reuse lock timeout")
	}
	return func() {
		releaseCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		var released sql.NullInt64
		_ = conn.QueryRowContext(releaseCtx, `SELECT RELEASE_LOCK(?)`, lockKey).Scan(&released)
	}, nil
}

func assessmentDraftReuseLockKey(entity AssessmentDraftEntity) string {
	datePart := "null"
	if entity.AssessmentDate != nil && !entity.AssessmentDate.IsZero() {
		datePart = entity.AssessmentDate.Format("2006-01-02")
	}
	raw := fmt.Sprintf(
		"%d:%d:%s:%s",
		entity.InstID,
		entity.StudentID,
		strings.TrimSpace(entity.AssessmentCode),
		datePart,
	)
	sum := sha1.Sum([]byte(raw))
	return "assessment_draft_reuse:" + hex.EncodeToString(sum[:])
}

func findReusableAssessmentDraftIDForUpdate(
	ctx context.Context,
	tx *sql.Tx,
	entity AssessmentDraftEntity,
) (int64, error) {
	if entity.InstID <= 0 || entity.StudentID <= 0 || strings.TrimSpace(entity.AssessmentCode) == "" {
		return 0, sql.ErrNoRows
	}
	var (
		query string
		args  []any
	)
	baseArgs := []any{
		entity.InstID,
		entity.StudentID,
		strings.TrimSpace(entity.AssessmentCode),
	}
	if entity.AssessmentDate == nil || entity.AssessmentDate.IsZero() {
		query = `
			SELECT id
			FROM assessment_draft
			WHERE inst_id = ?
			  AND student_id = ?
			  AND assessment_code = ?
			  AND assessment_date IS NULL
			  AND submitted_record_id = 0
			  AND del_flag = 0
			  AND status <> 'submitted'
			ORDER BY update_time DESC, id DESC
			LIMIT 1
			FOR UPDATE
		`
		args = baseArgs
	} else {
		query = `
			SELECT id
			FROM assessment_draft
			WHERE inst_id = ?
			  AND student_id = ?
			  AND assessment_code = ?
			  AND assessment_date = ?
			  AND submitted_record_id = 0
			  AND del_flag = 0
			  AND status <> 'submitted'
			ORDER BY update_time DESC, id DESC
			LIMIT 1
			FOR UPDATE
		`
		args = append(baseArgs, nullableAssessmentDate(entity.AssessmentDate))
	}
	var draftID int64
	if err := tx.QueryRowContext(ctx, query, args...).Scan(&draftID); err != nil {
		return 0, err
	}
	return draftID, nil
}

type assessmentDraftExecContext interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func saveAssessmentDraftWithExecutor(
	ctx context.Context,
	exec assessmentDraftExecContext,
	entity AssessmentDraftEntity,
) (int64, error) {
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
		result, err := exec.ExecContext(ctx, `
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

	result, err := exec.ExecContext(ctx, `
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

func (repo *Repository) UpsertAssessmentDraftItemDetails(ctx context.Context, instID, draftID int64, itemNo int, score *int, recordValues map[string]any, replaceRecordValues bool, operatorID int64) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	if err = upsertAssessmentDraftItemDetailsInTx(ctx, tx, instID, draftID, itemNo, score, recordValues, replaceRecordValues, operatorID); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}

func (repo *Repository) UpdateAssessmentDraftInputProgressAndItemDetails(ctx context.Context, instID, draftID int64, input any, progress any, answeredItemCount, rawScoreCount int, status string, itemNo int, score *int, recordValues map[string]any, replaceRecordValues bool, operatorID int64) error {
	inputRaw, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("marshal assessment draft input: %w", err)
	}
	progressRaw, err := json.Marshal(progress)
	if err != nil {
		return fmt.Errorf("marshal assessment draft progress: %w", err)
	}
	status = strings.TrimSpace(status)
	if status == "" {
		status = "draft"
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	var existingDraftID int64
	if err = tx.QueryRowContext(ctx, `
		SELECT id
		FROM assessment_draft
		WHERE id = ? AND inst_id = ? AND del_flag = 0 AND submitted_record_id = 0
		FOR UPDATE
	`, draftID, instID).Scan(&existingDraftID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `
		UPDATE assessment_draft
		SET input_json = ?,
		    progress_json = ?,
		    answered_item_count = ?,
		    raw_score_count = ?,
		    status = ?,
		    update_id = ?,
		    update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0 AND submitted_record_id = 0
	`,
		string(inputRaw),
		string(progressRaw),
		answeredItemCount,
		rawScoreCount,
		status,
		operatorID,
		draftID,
		instID,
	); err != nil {
		return err
	}
	if err = upsertAssessmentDraftItemDetailsInTx(ctx, tx, instID, draftID, itemNo, score, recordValues, replaceRecordValues, operatorID); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}

func upsertAssessmentDraftItemDetailsInTx(ctx context.Context, tx *sql.Tx, instID, draftID int64, itemNo int, score *int, recordValues map[string]any, replaceRecordValues bool, operatorID int64) error {
	if itemNo <= 0 {
		return fmt.Errorf("invalid item no %d", itemNo)
	}
	if score != nil {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO assessment_draft_item_score (
				inst_id, draft_id, item_no, score, create_id, update_id, create_time, update_time, del_flag
			) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
			ON DUPLICATE KEY UPDATE
				inst_id = VALUES(inst_id),
				score = VALUES(score),
				update_id = VALUES(update_id),
				update_time = NOW(),
				del_flag = 0
		`, instID, draftID, itemNo, *score, operatorID, operatorID); err != nil {
			return err
		}
	}
	if replaceRecordValues {
		if _, err := tx.ExecContext(ctx, `
			UPDATE assessment_draft_item_record_value
			SET del_flag = 1, update_id = ?, update_time = NOW()
			WHERE inst_id = ? AND draft_id = ? AND item_no = ? AND del_flag = 0
		`, operatorID, instID, draftID, itemNo); err != nil {
			return err
		}
		for fieldKey, value := range recordValues {
			key := strings.TrimSpace(fieldKey)
			if key == "" || isAssessmentDraftEmptyDetailValue(value) {
				continue
			}
			raw, marshalErr := json.Marshal(value)
			if marshalErr != nil {
				return fmt.Errorf("marshal draft item record value %d/%s: %w", itemNo, key, marshalErr)
			}
			if _, err := tx.ExecContext(ctx, `
					INSERT INTO assessment_draft_item_record_value (
						inst_id, draft_id, item_no, field_key, value_json, create_id, update_id, create_time, update_time, del_flag
					) VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
				ON DUPLICATE KEY UPDATE
					inst_id = VALUES(inst_id),
					value_json = VALUES(value_json),
					update_id = VALUES(update_id),
					update_time = NOW(),
					del_flag = 0
			`, instID, draftID, itemNo, key, string(raw), operatorID, operatorID); err != nil {
				return err
			}
		}
	}
	return nil
}

func replaceAssessmentDraftItemScores(ctx context.Context, tx *sql.Tx, instID, draftID int64, itemScores map[int]int, operatorID int64) error {
	if _, err := tx.ExecContext(ctx, `
		UPDATE assessment_draft_item_score
		SET del_flag = 1, update_id = ?, update_time = NOW()
		WHERE inst_id = ? AND draft_id = ? AND del_flag = 0
	`, operatorID, instID, draftID); err != nil {
		return err
	}
	itemNos := make([]int, 0, len(itemScores))
	for itemNo := range itemScores {
		itemNos = append(itemNos, itemNo)
	}
	sort.Ints(itemNos)
	for _, itemNo := range itemNos {
		if itemNo <= 0 {
			continue
		}
		score := itemScores[itemNo]
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO assessment_draft_item_score (
				inst_id, draft_id, item_no, score, create_id, update_id, create_time, update_time, del_flag
			) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
			ON DUPLICATE KEY UPDATE
				inst_id = VALUES(inst_id),
				score = VALUES(score),
				update_id = VALUES(update_id),
				update_time = NOW(),
				del_flag = 0
		`, instID, draftID, itemNo, score, operatorID, operatorID); err != nil {
			return err
		}
	}
	return nil
}

func replaceAssessmentDraftRawScores(ctx context.Context, tx *sql.Tx, instID, draftID int64, rawScores map[string]int, operatorID int64) error {
	if _, err := tx.ExecContext(ctx, `
		UPDATE assessment_draft_raw_score
		SET del_flag = 1, update_id = ?, update_time = NOW()
		WHERE inst_id = ? AND draft_id = ? AND del_flag = 0
	`, operatorID, instID, draftID); err != nil {
		return err
	}
	normalizedScores := make(map[string]int, len(rawScores))
	for scaleCode, rawScore := range rawScores {
		if normalized := strings.ToUpper(strings.TrimSpace(scaleCode)); normalized != "" {
			normalizedScores[normalized] = rawScore
		}
	}
	scaleCodes := make([]string, 0, len(normalizedScores))
	for scaleCode := range normalizedScores {
		scaleCodes = append(scaleCodes, scaleCode)
	}
	sort.Strings(scaleCodes)
	for _, scaleCode := range scaleCodes {
		rawScore := normalizedScores[scaleCode]
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO assessment_draft_raw_score (
				inst_id, draft_id, scale_code, raw_score, create_id, update_id, create_time, update_time, del_flag
			) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
			ON DUPLICATE KEY UPDATE
				inst_id = VALUES(inst_id),
				raw_score = VALUES(raw_score),
				update_id = VALUES(update_id),
				update_time = NOW(),
				del_flag = 0
		`, instID, draftID, scaleCode, rawScore, operatorID, operatorID); err != nil {
			return err
		}
	}
	return nil
}

func replaceAssessmentDraftItemRecordValues(ctx context.Context, tx *sql.Tx, instID, draftID int64, itemRecordValues map[int]map[string]any, operatorID int64) error {
	if _, err := tx.ExecContext(ctx, `
		UPDATE assessment_draft_item_record_value
		SET del_flag = 1, update_id = ?, update_time = NOW()
		WHERE inst_id = ? AND draft_id = ? AND del_flag = 0
	`, operatorID, instID, draftID); err != nil {
		return err
	}
	itemNos := make([]int, 0, len(itemRecordValues))
	for itemNo := range itemRecordValues {
		itemNos = append(itemNos, itemNo)
	}
	sort.Ints(itemNos)
	for _, itemNo := range itemNos {
		if itemNo <= 0 {
			continue
		}
		fieldKeys := make([]string, 0, len(itemRecordValues[itemNo]))
		for fieldKey := range itemRecordValues[itemNo] {
			fieldKeys = append(fieldKeys, fieldKey)
		}
		sort.Strings(fieldKeys)
		for _, fieldKey := range fieldKeys {
			key := strings.TrimSpace(fieldKey)
			value := itemRecordValues[itemNo][fieldKey]
			if key == "" || isAssessmentDraftEmptyDetailValue(value) {
				continue
			}
			raw, err := json.Marshal(value)
			if err != nil {
				return fmt.Errorf("marshal draft item record value %d/%s: %w", itemNo, key, err)
			}
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO assessment_draft_item_record_value (
					inst_id, draft_id, item_no, field_key, value_json, create_id, update_id, create_time, update_time, del_flag
				) VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
				ON DUPLICATE KEY UPDATE
					inst_id = VALUES(inst_id),
					value_json = VALUES(value_json),
					update_id = VALUES(update_id),
					update_time = NOW(),
					del_flag = 0
			`, instID, draftID, itemNo, key, string(raw), operatorID, operatorID); err != nil {
				return err
			}
		}
	}
	return nil
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

func isAssessmentDraftEmptyDetailValue(value any) bool {
	if value == nil {
		return true
	}
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	case []any:
		return len(v) == 0
	case []string:
		return len(v) == 0
	default:
		return false
	}
}

func (repo *Repository) GetAssessmentDraftBySubmittedRecordID(ctx context.Context, instID, recordID int64) (model.AssessmentDraftDetailVO, error) {
	var draftID int64
	if err := repo.db.QueryRowContext(ctx, `
		SELECT id
		FROM assessment_draft
		WHERE inst_id = ? AND submitted_record_id = ? AND del_flag = 0
		ORDER BY update_time DESC, id DESC
		LIMIT 1
	`, instID, recordID).Scan(&draftID); err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	return repo.GetAssessmentDraft(ctx, instID, draftID)
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
	} else {
		where = append(where, "status <> 'submitted'")
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
	selectSQL := `
		SELECT id, inst_id, student_id, student_name, assessment_code, assessment_name, scale_version,
		       birth_date, assessment_date, examiner_id, examiner_name, progress_json,
		       answered_item_count, raw_score_count, status, submitted_record_id, remark, create_time, update_time
		FROM assessment_draft
		WHERE ` + whereSQL
	listSQL := selectSQL
	listArgs := append([]any{}, args...)
	countSQL := "SELECT COUNT(1) FROM assessment_draft WHERE " + whereSQL
	countArgs := append([]any{}, args...)
	if query.LatestOnly {
		listSQL = `
			SELECT *
			FROM (` + selectSQL + `) ranked
			WHERE NOT EXISTS (
				SELECT 1
				FROM (` + selectSQL + `) newer
				WHERE newer.student_id = ranked.student_id
				  AND newer.assessment_code = ranked.assessment_code
				  AND (
				  	newer.update_time > ranked.update_time
				  	OR (newer.update_time = ranked.update_time AND newer.id > ranked.id)
				  )
			)
		`
		listArgs = append(append([]any{}, args...), args...)
		countSQL = "SELECT COUNT(1) FROM (" + listSQL + ") latest"
		countArgs = append([]any{}, listArgs...)
	}
	var total int
	if err := repo.db.QueryRowContext(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return model.PageResult[model.AssessmentDraftSummaryVO]{}, err
	}
	offset := (current - 1) * size
	rows, err := repo.db.QueryContext(ctx, listSQL+`
		ORDER BY update_time DESC, id DESC
		LIMIT ? OFFSET ?
	`, append(listArgs, size, offset)...)
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

func (repo *Repository) UpdateAssessmentDraftInputAndProgressIncludingSubmitted(
	ctx context.Context,
	instID, draftID int64,
	input any,
	progress any,
	answeredItemCount, rawScoreCount int,
	status string,
	operatorID int64,
	itemScores map[int]int,
	rawScores map[string]int,
	itemRecordValues map[int]map[string]any,
) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	if err = updateAssessmentDraftInputAndProgressIncludingSubmittedWithExecutor(
		ctx,
		tx,
		instID,
		draftID,
		input,
		progress,
		answeredItemCount,
		rawScoreCount,
		status,
		operatorID,
	); err != nil {
		return err
	}
	if err = replaceAssessmentDraftItemScores(ctx, tx, instID, draftID, itemScores, operatorID); err != nil {
		return err
	}
	if err = replaceAssessmentDraftRawScores(ctx, tx, instID, draftID, rawScores, operatorID); err != nil {
		return err
	}
	if err = replaceAssessmentDraftItemRecordValues(ctx, tx, instID, draftID, itemRecordValues, operatorID); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}

func updateAssessmentDraftInputAndProgressIncludingSubmittedWithExecutor(
	ctx context.Context,
	exec assessmentDraftExecContext,
	instID, draftID int64,
	input any,
	progress any,
	answeredItemCount, rawScoreCount int,
	status string,
	operatorID int64,
) error {
	inputRaw, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("marshal assessment draft input: %w", err)
	}
	progressRaw, err := json.Marshal(progress)
	if err != nil {
		return fmt.Errorf("marshal assessment draft progress: %w", err)
	}
	status = strings.TrimSpace(status)
	if status == "" {
		status = "draft"
	}

	result, err := exec.ExecContext(ctx, `
		UPDATE assessment_draft
		SET input_json = ?,
		    progress_json = ?,
		    answered_item_count = ?,
		    raw_score_count = ?,
		    status = ?,
		    update_id = ?,
		    update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`,
		string(inputRaw),
		string(progressRaw),
		answeredItemCount,
		rawScoreCount,
		status,
		operatorID,
		draftID,
		instID,
	)
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

func (repo *Repository) UpdateAssessmentDraftConfigIncludingSubmitted(ctx context.Context, instID, draftID int64, examinerName string, assessmentDate time.Time, operatorID int64) error {
	result, err := repo.db.ExecContext(ctx, `
		UPDATE assessment_draft
		SET examiner_name = ?,
		    assessment_date = ?,
		    input_json = JSON_SET(
		        JSON_SET(input_json, '$.examinerName', ?),
		        '$.assessmentDate', DATE_FORMAT(?, '%Y-%m-%d')
		    ),
		    update_id = ?,
		    update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`,
		strings.TrimSpace(examinerName),
		assessmentDate,
		strings.TrimSpace(examinerName),
		assessmentDate,
		operatorID,
		draftID,
		instID,
	)
	if err != nil {
		return err
	}
	_, _ = result.RowsAffected()
	return nil
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

func (repo *Repository) CreateAssessmentCaregiverInvite(ctx context.Context, entity AssessmentCaregiverInviteEntity) error {
	_, err := repo.db.ExecContext(ctx, `
		INSERT INTO assessment_caregiver_invite (
			ticket, inst_id, draft_id, record_id, wechat_url_link, mini_program_code_data_url, expires_at, create_time, update_time, del_flag
		) VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
	`,
		strings.TrimSpace(entity.Ticket),
		entity.InstID,
		entity.DraftID,
		entity.RecordID,
		strings.TrimSpace(entity.WeChatURLLink),
		strings.TrimSpace(entity.MiniProgramCodeDataURL),
		entity.ExpiresAt,
	)
	return err
}

func (repo *Repository) UpsertAssessmentCaregiverInvite(ctx context.Context, entity AssessmentCaregiverInviteEntity) error {
	_, err := repo.db.ExecContext(ctx, `
		INSERT INTO assessment_caregiver_invite (
			ticket, inst_id, draft_id, record_id, wechat_url_link, mini_program_code_data_url, expires_at, create_time, update_time, del_flag
		) VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
		ON DUPLICATE KEY UPDATE
			inst_id = VALUES(inst_id),
			draft_id = VALUES(draft_id),
			record_id = VALUES(record_id),
			wechat_url_link = VALUES(wechat_url_link),
			mini_program_code_data_url = VALUES(mini_program_code_data_url),
			expires_at = VALUES(expires_at),
			update_time = NOW(),
			del_flag = 0
	`,
		strings.TrimSpace(entity.Ticket),
		entity.InstID,
		entity.DraftID,
		entity.RecordID,
		strings.TrimSpace(entity.WeChatURLLink),
		strings.TrimSpace(entity.MiniProgramCodeDataURL),
		entity.ExpiresAt,
	)
	return err
}

func (repo *Repository) GetActiveAssessmentCaregiverInviteByDraft(ctx context.Context, instID, draftID int64) (AssessmentCaregiverInviteEntity, error) {
	var (
		item                   AssessmentCaregiverInviteEntity
		wechatURLLink          sql.NullString
		miniProgramCodeDataURL sql.NullString
		expiresAt              sql.NullTime
	)
	err := repo.db.QueryRowContext(ctx, `
		SELECT id, ticket, inst_id, draft_id, record_id, wechat_url_link, mini_program_code_data_url, expires_at
		FROM assessment_caregiver_invite
		WHERE inst_id = ? AND draft_id = ? AND del_flag = 0
		  AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY update_time DESC, id DESC
		LIMIT 1
	`, instID, draftID).Scan(
		&item.ID,
		&item.Ticket,
		&item.InstID,
		&item.DraftID,
		&item.RecordID,
		&wechatURLLink,
		&miniProgramCodeDataURL,
		&expiresAt,
	)
	if err != nil {
		return AssessmentCaregiverInviteEntity{}, err
	}
	if expiresAt.Valid {
		item.ExpiresAt = expiresAt.Time
	}
	if wechatURLLink.Valid {
		item.WeChatURLLink = strings.TrimSpace(wechatURLLink.String)
	}
	if miniProgramCodeDataURL.Valid {
		item.MiniProgramCodeDataURL = strings.TrimSpace(miniProgramCodeDataURL.String)
	}
	return item, nil
}

func (repo *Repository) GetAssessmentCaregiverInviteByTicket(ctx context.Context, ticket string) (AssessmentCaregiverInviteEntity, error) {
	var (
		item                   AssessmentCaregiverInviteEntity
		wechatURLLink          sql.NullString
		miniProgramCodeDataURL sql.NullString
		expiresAt              sql.NullTime
	)
	err := repo.db.QueryRowContext(ctx, `
		SELECT id, ticket, inst_id, draft_id, record_id, wechat_url_link, mini_program_code_data_url, expires_at
		FROM assessment_caregiver_invite
		WHERE ticket = ? AND del_flag = 0
		LIMIT 1
	`, strings.TrimSpace(ticket)).Scan(
		&item.ID,
		&item.Ticket,
		&item.InstID,
		&item.DraftID,
		&item.RecordID,
		&wechatURLLink,
		&miniProgramCodeDataURL,
		&expiresAt,
	)
	if err != nil {
		return AssessmentCaregiverInviteEntity{}, err
	}
	if expiresAt.Valid {
		item.ExpiresAt = expiresAt.Time
	}
	if wechatURLLink.Valid {
		item.WeChatURLLink = strings.TrimSpace(wechatURLLink.String)
	}
	if miniProgramCodeDataURL.Valid {
		item.MiniProgramCodeDataURL = strings.TrimSpace(miniProgramCodeDataURL.String)
	}
	return item, nil
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
