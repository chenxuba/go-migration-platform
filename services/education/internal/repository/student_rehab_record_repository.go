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

func sanitizeRehabRecordMediaList(items []model.RehabRecordMediaItem) []model.RehabRecordMediaItem {
	if len(items) == 0 {
		return nil
	}

	result := make([]model.RehabRecordMediaItem, 0, len(items))
	for _, item := range items {
		url := strings.TrimSpace(item.URL)
		if url == "" {
			continue
		}

		mediaType := strings.ToLower(strings.TrimSpace(item.MediaType))
		if mediaType != "image" && mediaType != "video" {
			lowerURL := strings.ToLower(url)
			switch {
			case strings.Contains(lowerURL, ".mp4"), strings.Contains(lowerURL, ".mov"), strings.Contains(lowerURL, ".webm"), strings.Contains(lowerURL, ".ogg"), strings.Contains(lowerURL, ".m4v"):
				mediaType = "video"
			default:
				mediaType = "image"
			}
		}

		size := item.Size
		if size < 0 {
			size = 0
		}

		result = append(result, model.RehabRecordMediaItem{
			MediaType: mediaType,
			URL:       url,
			FileName:  strings.TrimSpace(item.FileName),
			Size:      size,
		})
	}
	return result
}

func sanitizeStudentRehabRecordContent(content model.RehabRecordContent) model.RehabRecordContent {
	content.StudentName = strings.TrimSpace(content.StudentName)
	content.Gender = strings.TrimSpace(content.Gender)
	content.BirthDate = strings.TrimSpace(content.BirthDate)
	content.ClassName = strings.TrimSpace(content.ClassName)
	content.TeacherName = strings.TrimSpace(content.TeacherName)
	content.TrainingDate = strings.TrimSpace(content.TrainingDate)
	content.TrainingTarget = strings.TrimSpace(content.TrainingTarget)
	content.TrainingMediaList = sanitizeRehabRecordMediaList(content.TrainingMediaList)
	content.Performance = strings.TrimSpace(content.Performance)
	content.Suggestion = strings.TrimSpace(content.Suggestion)
	content.ParentFeedback = strings.TrimSpace(content.ParentFeedback)
	content.ParentSignature = strings.TrimSpace(content.ParentSignature)
	content.FeedbackDate = strings.TrimSpace(content.FeedbackDate)
	content.PerformanceMediaList = sanitizeRehabRecordMediaList(content.PerformanceMediaList)
	content.SuggestionMediaList = sanitizeRehabRecordMediaList(content.SuggestionMediaList)

	items := make([]model.RehabRecordTrainingItem, 0, len(content.TrainingItems))
	for _, item := range content.TrainingItems {
		title := strings.TrimSpace(item.Title)
		body := strings.TrimSpace(item.Content)
		mediaList := sanitizeRehabRecordMediaList(item.MediaList)
		if title == "" && body == "" && len(mediaList) == 0 {
			continue
		}
		items = append(items, model.RehabRecordTrainingItem{
			Title:     title,
			Content:   body,
			MediaList: mediaList,
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

func parseStudentRehabRecordSnapshot(template model.RehabRecordTemplateMeta, contentJSON, updatedStaffName string, updatedTime sql.NullTime) (*model.StudentRehabRecordSnapshot, error) {
	text := strings.TrimSpace(contentJSON)
	if text == "" {
		return nil, nil
	}

	var content model.RehabRecordContent
	if err := json.Unmarshal([]byte(text), &content); err != nil {
		return nil, err
	}

	snapshot := &model.StudentRehabRecordSnapshot{
		Template:         template,
		Content:          content,
		UpdatedStaffName: strings.TrimSpace(updatedStaffName),
	}
	if updatedTime.Valid {
		snapshot.UpdatedTime = updatedTime.Time.Format("2006-01-02 15:04:05")
	}
	return snapshot, nil
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
		studentID            int64
		lessonID             int64
		startTime            sql.NullTime
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
			IFNULL(srr.published_staff_name, ''),
			IFNULL(str.student_id, 0),
			IFNULL(str.lesson_id, 0),
			str.start_time
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
		&studentID,
		&lessonID,
		&startTime,
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
	if draft, err := parseStudentRehabRecordSnapshot(template, draftContentJSON.String, draftSavedStaffName, draftSavedTime); err != nil {
		return model.StudentRehabRecordDetailResult{}, err
	} else if draft != nil {
		result.HasDraft = true
		result.Draft = draft
	}
	if published, err := parseStudentRehabRecordSnapshot(template, publishedContentJSON.String, publishedStaffName, publishedTime); err != nil {
		return model.StudentRehabRecordDetailResult{}, err
	} else if published != nil {
		result.HasPublished = true
		result.Published = published
	}

	if studentID > 0 && lessonID > 0 && startTime.Valid {
		var (
			previousContentJSON    sql.NullString
			previousPublishedTime  sql.NullTime
			previousStaffName      string
			previousTemplateCode   string
			previousTemplateName   string
			previousTemplateScope  string
			previousTemplateVerion int
			previousAssignmentID   int64
		)

		err := repo.db.QueryRowContext(ctx, `
			SELECT
				IFNULL(srr.template_code, ''),
				IFNULL(srr.template_name, ''),
				IFNULL(srr.template_version, 0),
				IFNULL(srr.template_scope, ''),
				IFNULL(srr.template_assignment_id, 0),
				IFNULL(srr.published_content_json, ''),
				srr.published_time,
				IFNULL(srr.published_staff_name, '')
			FROM student_rehab_record srr
			INNER JOIN student_teaching_record str
				ON str.inst_id = srr.inst_id
			   AND str.id = srr.student_teaching_record_id
			   AND str.del_flag = 0
			WHERE srr.inst_id = ?
			  AND srr.del_flag = 0
			  AND str.student_id = ?
			  AND IFNULL(str.lesson_id, 0) = ?
			  AND str.id <> ?
			  AND LENGTH(TRIM(IFNULL(srr.published_content_json, ''))) > 0
			  AND (
				str.start_time < ?
				OR (str.start_time = ? AND str.id < ?)
			  )
			ORDER BY str.start_time DESC, srr.published_time DESC, str.id DESC
			LIMIT 1
		`, instID, studentID, lessonID, studentTeachingRecordID, startTime.Time, startTime.Time, studentTeachingRecordID).Scan(
			&previousTemplateCode,
			&previousTemplateName,
			&previousTemplateVerion,
			&previousTemplateScope,
			&previousAssignmentID,
			&previousContentJSON,
			&previousPublishedTime,
			&previousStaffName,
		)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return model.StudentRehabRecordDetailResult{}, err
		}
		if err == nil {
			previousTemplate := defaultStudentRehabRecordTemplate()
			if strings.TrimSpace(previousTemplateCode) != "" {
				previousTemplate.TemplateCode = strings.TrimSpace(previousTemplateCode)
			}
			if strings.TrimSpace(previousTemplateName) != "" {
				previousTemplate.TemplateName = strings.TrimSpace(previousTemplateName)
			}
			if previousTemplateVerion > 0 {
				previousTemplate.TemplateVersion = previousTemplateVerion
			}
			if strings.TrimSpace(previousTemplateScope) != "" {
				previousTemplate.TemplateScope = strings.TrimSpace(previousTemplateScope)
			}
			if previousAssignmentID > 0 {
				previousTemplate.TemplateAssignmentID = strconv.FormatInt(previousAssignmentID, 10)
			}

			previousPublished, err := parseStudentRehabRecordSnapshot(previousTemplate, previousContentJSON.String, previousStaffName, previousPublishedTime)
			if err != nil {
				return model.StudentRehabRecordDetailResult{}, err
			}
			if previousPublished != nil {
				result.HasPreviousPublished = true
				result.PreviousPublished = previousPublished
			}
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

func (repo *Repository) SaveParentRehabFeedback(ctx context.Context, instID int64, dto model.ParentRehabFeedbackSaveDTO) (bool, error) {
	studentTeachingRecordID, err := strconv.ParseInt(strings.TrimSpace(dto.StudentTeachingRecordID), 10, 64)
	if err != nil || studentTeachingRecordID <= 0 {
		return false, errors.New("缺少有效的康复记录学员")
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	if _, err := repo.loadStudentTeachingRecordForUpdateTx(ctx, tx, instID, studentTeachingRecordID); err != nil {
		return false, err
	}

	var publishedContentJSON string
	if err := tx.QueryRowContext(ctx, `
		SELECT IFNULL(published_content_json, '')
		FROM student_rehab_record
		WHERE inst_id = ?
		  AND student_teaching_record_id = ?
		  AND del_flag = 0
		FOR UPDATE
	`, instID, studentTeachingRecordID).Scan(&publishedContentJSON); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, errors.New("当前康复记录暂未发布，无法提交家长反馈")
		}
		return false, err
	}

	publishedContentJSON = strings.TrimSpace(publishedContentJSON)
	if publishedContentJSON == "" {
		return false, errors.New("当前康复记录暂未发布，无法提交家长反馈")
	}

	var content model.RehabRecordContent
	if err := json.Unmarshal([]byte(publishedContentJSON), &content); err != nil {
		return false, fmt.Errorf("解析康复记录失败: %w", err)
	}

	content.ParentFeedback = strings.TrimSpace(dto.ParentFeedback)
	content.ParentSignature = strings.TrimSpace(dto.ParentSignature)
	if content.ParentFeedback != "" || content.ParentSignature != "" {
		content.FeedbackDate = time.Now().Format("2006-01-02")
	} else {
		content.FeedbackDate = ""
	}
	content = sanitizeStudentRehabRecordContent(content)

	contentJSON, err := json.Marshal(content)
	if err != nil {
		return false, err
	}
	summary := buildStudentRehabRecordSummary(content)

	if _, err := tx.ExecContext(ctx, `
		UPDATE student_rehab_record
		SET published_content_json = ?,
		    published_summary = ?,
		    update_id = 0,
		    update_time = NOW()
		WHERE inst_id = ?
		  AND student_teaching_record_id = ?
		  AND del_flag = 0
	`, string(contentJSON), summary, instID, studentTeachingRecordID); err != nil {
		return false, err
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}

func (repo *Repository) ListParentPublishedRehabRecordIDs(ctx context.Context, instID int64, studentIDs []int64, pageIndex, pageSize int) ([]string, int, error) {
	if len(studentIDs) == 0 {
		return []string{}, 0, nil
	}

	current := pageIndex
	if current <= 0 {
		current = 1
	}
	size := pageSize
	if size <= 0 {
		size = 20
	}
	offset := (current - 1) * size
	placeholders := sqlPlaceholders(len(studentIDs))

	args := make([]any, 0, len(studentIDs)+3)
	args = append(args, instID)
	for _, studentID := range studentIDs {
		args = append(args, studentID)
	}

	whereSQL := `
		str.inst_id = ?
		AND str.del_flag = 0
		AND str.student_id IN (` + placeholders + `)
		AND ` + publishedStudentRehabRecordExistsSQL("str")

	var total int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM student_teaching_record str
		WHERE `+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []string{}, 0, nil
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT CAST(str.id AS CHAR)
		FROM student_teaching_record str
		WHERE `+whereSQL+`
		ORDER BY str.start_time DESC, str.id DESC
		LIMIT ? OFFSET ?
	`, append(append([]any{}, args...), size, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]string, 0, size)
	for rows.Next() {
		var recordID string
		if err := rows.Scan(&recordID); err != nil {
			return nil, 0, err
		}
		recordID = strings.TrimSpace(recordID)
		if recordID == "" {
			continue
		}
		items = append(items, recordID)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return items, total, nil
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
