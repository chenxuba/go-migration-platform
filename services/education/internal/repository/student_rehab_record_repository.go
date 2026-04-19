package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

func ensureStudentRehabRecordTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS student_rehab_record (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			inst_id BIGINT NOT NULL,
			student_teaching_record_id BIGINT NOT NULL DEFAULT 0,
			teaching_record_id BIGINT NOT NULL DEFAULT 0,
			student_id BIGINT NOT NULL DEFAULT 0,
			template_code VARCHAR(64) NOT NULL DEFAULT '',
			template_name VARCHAR(100) NOT NULL DEFAULT '',
			template_version INT NOT NULL DEFAULT 1,
			template_scope VARCHAR(32) NOT NULL DEFAULT 'default',
			template_assignment_id BIGINT NOT NULL DEFAULT 0,
			template_snapshot_json LONGTEXT NULL,
			draft_content_json LONGTEXT NULL,
			draft_saved_staff_id BIGINT NOT NULL DEFAULT 0,
			draft_saved_staff_name VARCHAR(100) NOT NULL DEFAULT '',
			draft_saved_time DATETIME NULL,
			published_content_json LONGTEXT NULL,
			published_summary VARCHAR(500) NOT NULL DEFAULT '',
			published_staff_id BIGINT NOT NULL DEFAULT 0,
			published_staff_name VARCHAR(100) NOT NULL DEFAULT '',
			published_time DATETIME NULL,
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT NOT NULL DEFAULT 0,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uq_student_rehab_record_inst_student (inst_id, student_teaching_record_id),
			KEY idx_student_rehab_record_teaching (inst_id, teaching_record_id),
			KEY idx_student_rehab_record_student (inst_id, student_id)
		)
	`)
	return err
}

func publishedStudentRehabRecordExistsSQL(studentAlias string) string {
	alias := strings.TrimSpace(studentAlias)
	if alias == "" {
		alias = "student_teaching_record"
	}
	return fmt.Sprintf(`EXISTS (
		SELECT 1
		FROM student_rehab_record srr
		WHERE srr.inst_id = %s.inst_id
		  AND srr.student_teaching_record_id = %s.id
		  AND srr.del_flag = 0
		  AND LENGTH(TRIM(IFNULL(srr.published_content_json, ''))) > 0
	)`, alias, alias)
}

func draftStudentRehabRecordExistsSQL(studentAlias string) string {
	alias := strings.TrimSpace(studentAlias)
	if alias == "" {
		alias = "student_teaching_record"
	}
	return fmt.Sprintf(`EXISTS (
		SELECT 1
		FROM student_rehab_record srr
		WHERE srr.inst_id = %s.inst_id
		  AND srr.student_teaching_record_id = %s.id
		  AND srr.del_flag = 0
		  AND LENGTH(TRIM(IFNULL(srr.draft_content_json, ''))) > 0
	)`, alias, alias)
}

func studentTeachingRecordHasCommentSQL(studentAlias string) string {
	alias := strings.TrimSpace(studentAlias)
	if alias == "" {
		alias = "student_teaching_record"
	}
	return fmt.Sprintf("(LENGTH(TRIM(IFNULL(%s.external_remark, ''))) > 0 OR %s)", alias, publishedStudentRehabRecordExistsSQL(alias))
}

func defaultStudentRehabRecordTemplate() model.RehabRecordTemplateMeta {
	return model.RehabRecordTemplateMeta{
		TemplateCode:    "rehab-record-universal",
		TemplateName:    "通用康复记录模板",
		TemplateVersion: 1,
		TemplateScope:   "default",
	}
}

// 预留模板分配扩展点，后续可在这里根据机构/学员/教学记录解析实际分配模板。
func resolveStudentRehabRecordTemplate(template model.RehabRecordTemplateMeta) model.RehabRecordTemplateMeta {
	resolved := defaultStudentRehabRecordTemplate()
	if code := strings.TrimSpace(template.TemplateCode); code != "" {
		resolved.TemplateCode = code
	}
	if name := strings.TrimSpace(template.TemplateName); name != "" {
		resolved.TemplateName = name
	}
	if template.TemplateVersion > 0 {
		resolved.TemplateVersion = template.TemplateVersion
	}
	if scope := strings.TrimSpace(template.TemplateScope); scope != "" {
		resolved.TemplateScope = scope
	}
	resolved.TemplateAssignmentID = strings.TrimSpace(template.TemplateAssignmentID)
	return resolved
}

func sanitizeStudentRehabRecordContent(content model.RehabRecordContent) model.RehabRecordContent {
	content.StudentName = strings.TrimSpace(content.StudentName)
	content.Gender = strings.TrimSpace(content.Gender)
	content.BirthDate = strings.TrimSpace(content.BirthDate)
	content.ClassName = strings.TrimSpace(content.ClassName)
	content.TeacherName = strings.TrimSpace(content.TeacherName)
	content.TrainingDate = strings.TrimSpace(content.TrainingDate)
	content.TrainingTarget = strings.TrimSpace(content.TrainingTarget)
	content.Performance = strings.TrimSpace(content.Performance)
	content.Suggestion = strings.TrimSpace(content.Suggestion)
	content.ParentFeedback = strings.TrimSpace(content.ParentFeedback)
	content.ParentSignature = strings.TrimSpace(content.ParentSignature)
	content.FeedbackDate = strings.TrimSpace(content.FeedbackDate)

	items := make([]model.RehabRecordTrainingItem, 0, len(content.TrainingItems))
	for _, item := range content.TrainingItems {
		title := strings.TrimSpace(item.Title)
		body := strings.TrimSpace(item.Content)
		if title == "" && body == "" {
			continue
		}
		items = append(items, model.RehabRecordTrainingItem{
			Title:   title,
			Content: body,
		})
	}
	content.TrainingItems = items
	return content
}

func buildStudentRehabRecordSummary(content model.RehabRecordContent) string {
	candidates := []string{
		strings.TrimSpace(content.TrainingTarget),
		strings.TrimSpace(content.Performance),
		strings.TrimSpace(content.Suggestion),
	}
	for _, item := range content.TrainingItems {
		candidates = append(candidates, strings.TrimSpace(item.Title), strings.TrimSpace(item.Content))
	}
	for _, item := range candidates {
		if item != "" {
			runes := []rune(item)
			if len(runes) > 120 {
				return string(runes[:120])
			}
			return item
		}
	}
	return "已发布康复记录"
}

func parseTemplateAssignmentID(raw string) int64 {
	value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || value <= 0 {
		return 0
	}
	return value
}

func (repo *Repository) GetStudentRehabRecordDetail(ctx context.Context, instID int64, query model.StudentRehabRecordQueryDTO) (model.StudentRehabRecordDetailResult, error) {
	studentTeachingRecordID, err := strconv.ParseInt(strings.TrimSpace(query.StudentTeachingRecordID), 10, 64)
	if err != nil || studentTeachingRecordID <= 0 {
		return model.StudentRehabRecordDetailResult{}, errors.New("缺少有效的康复记录学员")
	}

	var (
		templateCode         string
		templateName         string
		templateVersion      int
		templateScope        string
		templateAssignmentID int64
		draftContentJSON     sql.NullString
		draftSavedTime       sql.NullTime
		draftSavedStaffName  string
		publishedContentJSON sql.NullString
		publishedTime        sql.NullTime
		publishedStaffName   string
	)

	if err := repo.db.QueryRowContext(ctx, `
		SELECT
			IFNULL(srr.template_code, ''),
			IFNULL(srr.template_name, ''),
			IFNULL(srr.template_version, 0),
			IFNULL(srr.template_scope, ''),
			IFNULL(srr.template_assignment_id, 0),
			IFNULL(srr.draft_content_json, ''),
			srr.draft_saved_time,
			IFNULL(srr.draft_saved_staff_name, ''),
			IFNULL(srr.published_content_json, ''),
			srr.published_time,
			IFNULL(srr.published_staff_name, '')
		FROM student_teaching_record str
		LEFT JOIN student_rehab_record srr
			ON srr.inst_id = str.inst_id
		   AND srr.student_teaching_record_id = str.id
		   AND srr.del_flag = 0
		WHERE str.inst_id = ?
		  AND str.id = ?
		  AND str.del_flag = 0
		LIMIT 1
	`, instID, studentTeachingRecordID).Scan(
		&templateCode,
		&templateName,
		&templateVersion,
		&templateScope,
		&templateAssignmentID,
		&draftContentJSON,
		&draftSavedTime,
		&draftSavedStaffName,
		&publishedContentJSON,
		&publishedTime,
		&publishedStaffName,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.StudentRehabRecordDetailResult{}, errors.New("未找到对应的康复记录学员")
		}
		return model.StudentRehabRecordDetailResult{}, err
	}

	template := defaultStudentRehabRecordTemplate()
	if strings.TrimSpace(templateCode) != "" {
		template.TemplateCode = strings.TrimSpace(templateCode)
	}
	if strings.TrimSpace(templateName) != "" {
		template.TemplateName = strings.TrimSpace(templateName)
	}
	if templateVersion > 0 {
		template.TemplateVersion = templateVersion
	}
	if strings.TrimSpace(templateScope) != "" {
		template.TemplateScope = strings.TrimSpace(templateScope)
	}
	if templateAssignmentID > 0 {
		template.TemplateAssignmentID = strconv.FormatInt(templateAssignmentID, 10)
	}

	result := model.StudentRehabRecordDetailResult{}
	if text := strings.TrimSpace(draftContentJSON.String); text != "" {
		var content model.RehabRecordContent
		if err := json.Unmarshal([]byte(text), &content); err != nil {
			return model.StudentRehabRecordDetailResult{}, err
		}
		result.HasDraft = true
		result.Draft = &model.StudentRehabRecordSnapshot{
			Template:         template,
			Content:          content,
			UpdatedStaffName: strings.TrimSpace(draftSavedStaffName),
		}
		if draftSavedTime.Valid {
			result.Draft.UpdatedTime = draftSavedTime.Time.Format("2006-01-02 15:04:05")
		}
	}
	if text := strings.TrimSpace(publishedContentJSON.String); text != "" {
		var content model.RehabRecordContent
		if err := json.Unmarshal([]byte(text), &content); err != nil {
			return model.StudentRehabRecordDetailResult{}, err
		}
		result.HasPublished = true
		result.Published = &model.StudentRehabRecordSnapshot{
			Template:         template,
			Content:          content,
			UpdatedStaffName: strings.TrimSpace(publishedStaffName),
		}
		if publishedTime.Valid {
			result.Published.UpdatedTime = publishedTime.Time.Format("2006-01-02 15:04:05")
		}
	}
	return result, nil
}

func (repo *Repository) SaveStudentRehabRecordDraft(ctx context.Context, instID, operatorID int64, dto model.SaveStudentRehabRecordDraftDTO) (bool, error) {
	return repo.upsertStudentRehabRecord(ctx, instID, operatorID, strings.TrimSpace(dto.StudentTeachingRecordID), dto.Template, dto.Content, false)
}

func (repo *Repository) PublishStudentRehabRecord(ctx context.Context, instID, operatorID int64, dto model.PublishStudentRehabRecordDTO) (bool, error) {
	return repo.upsertStudentRehabRecord(ctx, instID, operatorID, strings.TrimSpace(dto.StudentTeachingRecordID), dto.Template, dto.Content, true)
}

func (repo *Repository) upsertStudentRehabRecord(ctx context.Context, instID, operatorID int64, studentTeachingRecordIDRaw string, rawTemplate model.RehabRecordTemplateMeta, rawContent model.RehabRecordContent, publish bool) (bool, error) {
	studentTeachingRecordID, err := strconv.ParseInt(studentTeachingRecordIDRaw, 10, 64)
	if err != nil || studentTeachingRecordID <= 0 {
		return false, errors.New("缺少有效的康复记录学员")
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	row, err := repo.loadStudentTeachingRecordForUpdateTx(ctx, tx, instID, studentTeachingRecordID)
	if err != nil {
		return false, err
	}

	template := resolveStudentRehabRecordTemplate(rawTemplate)
	content := sanitizeStudentRehabRecordContent(rawContent)
	templateSnapshotJSON, err := json.Marshal(template)
	if err != nil {
		return false, err
	}
	contentJSON, err := json.Marshal(content)
	if err != nil {
		return false, err
	}

	operatorName := firstNonEmptyString(repo.GetStaffNameByID(ctx, &operatorID), "系统")
	now := time.Now()
	assignmentID := parseTemplateAssignmentID(template.TemplateAssignmentID)
	summary := buildStudentRehabRecordSummary(content)

	if publish {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO student_rehab_record (
				inst_id, student_teaching_record_id, teaching_record_id, student_id,
				template_code, template_name, template_version, template_scope, template_assignment_id, template_snapshot_json,
				draft_content_json, draft_saved_staff_id, draft_saved_staff_name, draft_saved_time,
				published_content_json, published_summary, published_staff_id, published_staff_name, published_time,
				create_id, create_time, update_id, update_time, del_flag
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, '', 0, '', NULL, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
			ON DUPLICATE KEY UPDATE
				teaching_record_id = VALUES(teaching_record_id),
				student_id = VALUES(student_id),
				template_code = VALUES(template_code),
				template_name = VALUES(template_name),
				template_version = VALUES(template_version),
				template_scope = VALUES(template_scope),
				template_assignment_id = VALUES(template_assignment_id),
				template_snapshot_json = VALUES(template_snapshot_json),
				draft_content_json = '',
				draft_saved_staff_id = 0,
				draft_saved_staff_name = '',
				draft_saved_time = NULL,
				published_content_json = VALUES(published_content_json),
				published_summary = VALUES(published_summary),
				published_staff_id = VALUES(published_staff_id),
				published_staff_name = VALUES(published_staff_name),
				published_time = VALUES(published_time),
				update_id = VALUES(update_id),
				update_time = VALUES(update_time),
				del_flag = 0
		`,
			instID, row.StudentTeachingRecordID, row.TeachingRecordID, row.StudentID,
			template.TemplateCode, template.TemplateName, template.TemplateVersion, template.TemplateScope, assignmentID, string(templateSnapshotJSON),
			string(contentJSON), summary, operatorID, operatorName, now,
			operatorID, now, operatorID, now,
		); err != nil {
			return false, err
		}

		if _, err := tx.ExecContext(ctx, `
			UPDATE student_teaching_record
			SET updated_staff_id = ?,
			    updated_staff_name = ?,
			    updated_time = NOW(),
			    update_id = ?,
			    update_time = NOW()
			WHERE inst_id = ?
			  AND id = ?
			  AND del_flag = 0
		`, operatorID, operatorName, operatorID, instID, row.StudentTeachingRecordID); err != nil {
			return false, err
		}
	} else {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO student_rehab_record (
				inst_id, student_teaching_record_id, teaching_record_id, student_id,
				template_code, template_name, template_version, template_scope, template_assignment_id, template_snapshot_json,
				draft_content_json, draft_saved_staff_id, draft_saved_staff_name, draft_saved_time,
				create_id, create_time, update_id, update_time, del_flag
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
			ON DUPLICATE KEY UPDATE
				teaching_record_id = VALUES(teaching_record_id),
				student_id = VALUES(student_id),
				template_code = VALUES(template_code),
				template_name = VALUES(template_name),
				template_version = VALUES(template_version),
				template_scope = VALUES(template_scope),
				template_assignment_id = VALUES(template_assignment_id),
				template_snapshot_json = VALUES(template_snapshot_json),
				draft_content_json = VALUES(draft_content_json),
				draft_saved_staff_id = VALUES(draft_saved_staff_id),
				draft_saved_staff_name = VALUES(draft_saved_staff_name),
				draft_saved_time = VALUES(draft_saved_time),
				update_id = VALUES(update_id),
				update_time = VALUES(update_time),
				del_flag = 0
		`,
			instID, row.StudentTeachingRecordID, row.TeachingRecordID, row.StudentID,
			template.TemplateCode, template.TemplateName, template.TemplateVersion, template.TemplateScope, assignmentID, string(templateSnapshotJSON),
			string(contentJSON), operatorID, operatorName, now,
			operatorID, now, operatorID, now,
		); err != nil {
			return false, err
		}
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}
