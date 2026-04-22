package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

func ensureHomeworkTables(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS homework_task (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			uuid VARCHAR(64) NULL,
			version BIGINT NOT NULL DEFAULT 0,
			inst_id BIGINT NOT NULL,
			title VARCHAR(100) NOT NULL DEFAULT '',
			content TEXT NULL,
			attachments_json LONGTEXT NULL,
			repeat_rule_json TEXT NULL,
			publish_rule INT NOT NULL DEFAULT 1,
			publish_time DATETIME NULL DEFAULT NULL,
			end_time DATETIME NULL DEFAULT NULL,
			publish_hour INT NOT NULL DEFAULT 0,
			task_duration_hours INT NOT NULL DEFAULT 0,
			end_hour INT NOT NULL DEFAULT 0,
			is_visible_student TINYINT(1) NOT NULL DEFAULT 0,
			source_type INT NOT NULL DEFAULT 1,
			source_id BIGINT NOT NULL DEFAULT 0,
			source_name VARCHAR(255) NOT NULL DEFAULT '',
			selected_students_json LONGTEXT NULL,
			student_count INT NOT NULL DEFAULT 0,
			unsubmitted_count INT NOT NULL DEFAULT 0,
			rejected_count INT NOT NULL DEFAULT 0,
			submitted_count INT NOT NULL DEFAULT 0,
			re_submitted_count INT NOT NULL DEFAULT 0,
			evaluated_count INT NOT NULL DEFAULT 0,
			unevaluated_count INT NOT NULL DEFAULT 0,
			read_count INT NOT NULL DEFAULT 0,
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT NOT NULL DEFAULT 0,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			KEY idx_homework_inst_publish (inst_id, publish_time, id),
			KEY idx_homework_inst_end (inst_id, end_time, id),
			KEY idx_homework_inst_source (inst_id, source_type, source_id),
			KEY idx_homework_inst_creator (inst_id, create_id),
			KEY idx_homework_inst_deleted (inst_id, del_flag, id)
		)
	`); err != nil {
		return err
	}

	for _, columns := range []map[string]string{
		{"attachments_json": "attachments_json LONGTEXT NULL AFTER content"},
		{"repeat_rule_json": "repeat_rule_json TEXT NULL AFTER attachments_json"},
		{"publish_rule": "publish_rule INT NOT NULL DEFAULT 1 AFTER repeat_rule_json"},
		{"publish_time": "publish_time DATETIME NULL DEFAULT NULL AFTER publish_rule"},
		{"end_time": "end_time DATETIME NULL DEFAULT NULL AFTER publish_time"},
		{"publish_hour": "publish_hour INT NOT NULL DEFAULT 0 AFTER end_time"},
		{"task_duration_hours": "task_duration_hours INT NOT NULL DEFAULT 0 AFTER publish_hour"},
		{"end_hour": "end_hour INT NOT NULL DEFAULT 0 AFTER task_duration_hours"},
		{"is_visible_student": "is_visible_student TINYINT(1) NOT NULL DEFAULT 0 AFTER end_hour"},
		{"source_type": "source_type INT NOT NULL DEFAULT 1 AFTER is_visible_student"},
		{"source_id": "source_id BIGINT NOT NULL DEFAULT 0 AFTER source_type"},
		{"source_name": "source_name VARCHAR(255) NOT NULL DEFAULT '' AFTER source_id"},
		{"selected_students_json": "selected_students_json LONGTEXT NULL AFTER source_name"},
		{"student_count": "student_count INT NOT NULL DEFAULT 0 AFTER selected_students_json"},
		{"unsubmitted_count": "unsubmitted_count INT NOT NULL DEFAULT 0 AFTER student_count"},
		{"rejected_count": "rejected_count INT NOT NULL DEFAULT 0 AFTER unsubmitted_count"},
		{"submitted_count": "submitted_count INT NOT NULL DEFAULT 0 AFTER rejected_count"},
		{"re_submitted_count": "re_submitted_count INT NOT NULL DEFAULT 0 AFTER submitted_count"},
		{"evaluated_count": "evaluated_count INT NOT NULL DEFAULT 0 AFTER re_submitted_count"},
		{"unevaluated_count": "unevaluated_count INT NOT NULL DEFAULT 0 AFTER evaluated_count"},
		{"read_count": "read_count INT NOT NULL DEFAULT 0 AFTER unevaluated_count"},
	} {
		if err := ensureColumnsOnTable(ctx, db, "homework_task", columns); err != nil {
			return err
		}
	}
	return nil
}

func (repo *Repository) ListHomeworkTargetStudents(ctx context.Context, instID int64, sourceType int, sourceID int64) (string, []model.HomeworkTargetStudent, error) {
	officialSubscribedExpr := studentOfficialSubscribedExistsSQL("s")
	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			IFNULL(tc.name, ''),
			tcs.student_id,
			IFNULL(s.stu_name, ''),
			CAST(IFNULL(tcs.primary_tuition_account_id, 0) AS CHAR),
			`+officialSubscribedExpr+`
		FROM teaching_class tc
		INNER JOIN teaching_class_student tcs
			ON tcs.teaching_class_id = tc.id
			AND tcs.inst_id = tc.inst_id
			AND tcs.del_flag = 0
		INNER JOIN inst_student s
			ON s.id = tcs.student_id
			AND s.del_flag = 0
		WHERE tc.inst_id = ?
		  AND tc.id = ?
		  AND tc.class_type = ?
		  AND tc.del_flag = 0
		ORDER BY tcs.id ASC
	`, instID, sourceID, sourceType)
	if err != nil {
		return "", nil, err
	}
	defer rows.Close()

	sourceName := ""
	students := make([]model.HomeworkTargetStudent, 0)
	for rows.Next() {
		var (
			rowSourceName string
			item          model.HomeworkTargetStudent
			isBind        int
		)
		if err := rows.Scan(
			&rowSourceName,
			&item.StudentID,
			&item.StudentName,
			&item.TuitionAccountID,
			&isBind,
		); err != nil {
			return "", nil, err
		}
		sourceName = strings.TrimSpace(rowSourceName)
		item.IsBind = isBind != 0
		students = append(students, item)
	}
	if err := rows.Err(); err != nil {
		return "", nil, err
	}
	if len(students) == 0 {
		return "", nil, sql.ErrNoRows
	}
	return sourceName, students, nil
}

func (repo *Repository) BatchCreateHomeworks(ctx context.Context, instID, operatorID int64, inputs []model.HomeworkMutationInput) ([]model.HomeworkOperationResult, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	results := make([]model.HomeworkOperationResult, 0, len(inputs))
	for _, input := range inputs {
		attachmentsJSON, err := marshalHomeworkValue(input.Attachments)
		if err != nil {
			return nil, err
		}
		repeatRuleJSON, err := marshalHomeworkOptionalValue(input.RepeatRule)
		if err != nil {
			return nil, err
		}
		selectedStudentsJSON, err := marshalHomeworkValue(input.SelectedStudents)
		if err != nil {
			return nil, err
		}
		studentCount := len(input.SelectedStudents)
		result, execErr := tx.ExecContext(ctx, `
			INSERT INTO homework_task (
				inst_id, title, content, attachments_json, repeat_rule_json, publish_rule,
				publish_time, end_time, publish_hour, task_duration_hours, end_hour, is_visible_student,
				source_type, source_id, source_name, selected_students_json,
				student_count, unsubmitted_count, rejected_count, submitted_count,
				re_submitted_count, evaluated_count, unevaluated_count, read_count,
				create_id, create_time, update_id, update_time, del_flag
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0, 0, 0, 0, 0, ?, NOW(), ?, NOW(), 0)
		`,
			instID,
			strings.TrimSpace(input.Title),
			strings.TrimSpace(input.Content),
			attachmentsJSON,
			repeatRuleJSON,
			input.PublishRule,
			input.PublishTime,
			input.EndTime,
			input.PublishHour,
			input.TaskDurationHours,
			input.EndHour,
			boolToTinyInt(input.IsVisibleStudent),
			input.SourceType,
			input.SourceID,
			strings.TrimSpace(input.SourceName),
			selectedStudentsJSON,
			studentCount,
			studentCount,
			operatorID,
			operatorID,
		)
		if execErr != nil {
			err = execErr
			return nil, err
		}
		insertID, execErr := result.LastInsertId()
		if execErr != nil {
			err = execErr
			return nil, err
		}
		results = append(results, model.HomeworkOperationResult{
			ID:   strconv.FormatInt(insertID, 10),
			Name: strings.TrimSpace(input.Title),
		})
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return results, nil
}

func (repo *Repository) UpdateHomework(ctx context.Context, instID, operatorID, id int64, input model.HomeworkMutationInput) error {
	attachmentsJSON, err := marshalHomeworkValue(input.Attachments)
	if err != nil {
		return err
	}
	repeatRuleJSON, err := marshalHomeworkOptionalValue(input.RepeatRule)
	if err != nil {
		return err
	}
	selectedStudentsJSON, err := marshalHomeworkValue(input.SelectedStudents)
	if err != nil {
		return err
	}
	studentCount := len(input.SelectedStudents)
	result, err := repo.db.ExecContext(ctx, `
		UPDATE homework_task
		SET title = ?,
			content = ?,
			attachments_json = ?,
			repeat_rule_json = ?,
			publish_rule = ?,
			publish_time = ?,
			end_time = ?,
			publish_hour = ?,
			task_duration_hours = ?,
			end_hour = ?,
			is_visible_student = ?,
			source_type = ?,
			source_id = ?,
			source_name = ?,
			selected_students_json = ?,
			student_count = ?,
			unsubmitted_count = GREATEST(? - IFNULL(submitted_count, 0), 0),
			update_id = ?,
			update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`,
		strings.TrimSpace(input.Title),
		strings.TrimSpace(input.Content),
		attachmentsJSON,
		repeatRuleJSON,
		input.PublishRule,
		input.PublishTime,
		input.EndTime,
		input.PublishHour,
		input.TaskDurationHours,
		input.EndHour,
		boolToTinyInt(input.IsVisibleStudent),
		input.SourceType,
		input.SourceID,
		strings.TrimSpace(input.SourceName),
		selectedStudentsJSON,
		studentCount,
		studentCount,
		operatorID,
		id,
		instID,
	)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected > 0 {
		return nil
	}
	exists, err := repo.homeworkExists(ctx, instID, id)
	if err != nil {
		return err
	}
	if !exists {
		return sql.ErrNoRows
	}
	return nil
}

func (repo *Repository) DeleteHomework(ctx context.Context, instID, operatorID, id int64) error {
	result, err := repo.db.ExecContext(ctx, `
		UPDATE homework_task
		SET del_flag = 1, update_id = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, operatorID, id, instID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected > 0 {
		return nil
	}
	exists, err := repo.homeworkExists(ctx, instID, id)
	if err != nil {
		return err
	}
	if !exists {
		return sql.ErrNoRows
	}
	return nil
}

func (repo *Repository) GetHomeworkDetail(ctx context.Context, instID, id int64) (model.HomeworkDetailVO, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT
			h.id,
			IFNULL(h.title, ''),
			IFNULL(h.content, ''),
			IFNULL(h.attachments_json, ''),
			IFNULL(h.repeat_rule_json, ''),
			IFNULL(h.publish_rule, 1),
			h.publish_time,
			h.end_time,
			IFNULL(h.publish_hour, 0),
			IFNULL(h.task_duration_hours, 0),
			IFNULL(h.end_hour, 0),
			IFNULL(h.is_visible_student, 0),
			IFNULL(h.source_type, 1),
			IFNULL(h.source_id, 0),
			IFNULL(h.source_name, ''),
			IFNULL(h.selected_students_json, ''),
			IFNULL(h.student_count, 0),
			IFNULL(h.unsubmitted_count, 0),
			IFNULL(h.rejected_count, 0),
			IFNULL(h.submitted_count, 0),
			IFNULL(h.re_submitted_count, 0),
			IFNULL(h.evaluated_count, 0),
			IFNULL(h.unevaluated_count, 0),
			IFNULL(h.read_count, 0),
			h.create_time,
			CAST(IFNULL(h.create_id, 0) AS CHAR),
			IFNULL(u.nick_name, '')
		FROM homework_task h
		LEFT JOIN inst_user u ON u.id = h.create_id AND u.del_flag = 0
		WHERE h.inst_id = ? AND h.id = ? AND h.del_flag = 0
		LIMIT 1
	`, instID, id)

	var (
		item                 model.HomeworkDetailVO
		dbID                 int64
		attachmentsJSON      string
		repeatRuleJSON       string
		selectedStudentsJSON string
		sourceID             int64
		isVisibleStudent     int
		createTime           sql.NullTime
		publishTime          sql.NullTime
		endTime              sql.NullTime
	)
	if err := row.Scan(
		&dbID,
		&item.Title,
		&item.Content,
		&attachmentsJSON,
		&repeatRuleJSON,
		&item.PublishRule,
		&publishTime,
		&endTime,
		&item.PublishHour,
		&item.TaskDurationHours,
		&item.EndHour,
		&isVisibleStudent,
		&item.SourceType,
		&sourceID,
		&item.SourceName,
		&selectedStudentsJSON,
		&item.StudentCount,
		&item.UnsubmittedCount,
		&item.RejectedCount,
		&item.SubmittedCount,
		&item.ReSubmittedCount,
		&item.EvaluatedCount,
		&item.UnevaluatedCount,
		&item.ReadCount,
		&createTime,
		&item.CreatedStaffID,
		&item.CreatedStaffName,
	); err != nil {
		return model.HomeworkDetailVO{}, err
	}

	item.ID = strconv.FormatInt(dbID, 10)
	item.SourceID = strconv.FormatInt(sourceID, 10)
	item.IsVisibleStudent = isVisibleStudent != 0
	item.UnreadCount = maxInt(item.StudentCount-item.ReadCount, 0)
	if createTime.Valid {
		t := createTime.Time
		item.CreatedTime = &t
	}
	if publishTime.Valid {
		t := publishTime.Time
		item.PublishTime = &t
	}
	if endTime.Valid {
		t := endTime.Time
		item.EndTime = &t
	}
	if strings.TrimSpace(attachmentsJSON) != "" {
		if err := json.Unmarshal([]byte(attachmentsJSON), &item.Attachments); err != nil {
			return model.HomeworkDetailVO{}, err
		}
	}
	if strings.TrimSpace(repeatRuleJSON) != "" {
		item.RepeatRule = &model.HomeworkRepeatRule{}
		if err := json.Unmarshal([]byte(repeatRuleJSON), item.RepeatRule); err != nil {
			return model.HomeworkDetailVO{}, err
		}
	}
	if strings.TrimSpace(selectedStudentsJSON) != "" {
		if err := json.Unmarshal([]byte(selectedStudentsJSON), &item.SelectedStudents); err != nil {
			return model.HomeworkDetailVO{}, err
		}
	}
	if item.Attachments == nil {
		item.Attachments = []model.HomeworkAttachment{}
	}
	if item.SelectedStudents == nil {
		item.SelectedStudents = []model.HomeworkSelectedStudent{}
	}
	return item, nil
}

func (repo *Repository) PageHomeworks(ctx context.Context, instID int64, query model.HomeworkListQueryDTO) (model.HomeworkPageResultVO, error) {
	current := query.PageRequestModel.PageIndex
	size := query.PageRequestModel.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	if size > 200 {
		size = 200
	}
	offset := (current - 1) * size
	if query.PageRequestModel.SkipCount > 0 {
		offset = query.PageRequestModel.SkipCount
	}

	whereSQL, args, err := buildHomeworkWhere(query.QueryModel)
	if err != nil {
		return model.HomeworkPageResultVO{}, err
	}

	countArgs := append([]any{instID}, args...)
	var total int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM homework_task h
		WHERE h.inst_id = ? AND h.del_flag = 0 `+whereSQL,
		countArgs...,
	).Scan(&total); err != nil {
		return model.HomeworkPageResultVO{}, err
	}

	orderBy := " ORDER BY h.publish_time DESC, h.id DESC"
	if query.SortModel.PublishTime > 0 {
		orderBy = " ORDER BY h.publish_time ASC, h.id ASC"
	}

	dataArgs := append([]any{instID}, args...)
	dataArgs = append(dataArgs, size, offset)
	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			h.id,
			IFNULL(h.title, ''),
			IFNULL(h.content, ''),
			IFNULL(h.student_count, 0),
			IFNULL(h.unsubmitted_count, 0),
			IFNULL(h.rejected_count, 0),
			IFNULL(h.submitted_count, 0),
			IFNULL(h.re_submitted_count, 0),
			IFNULL(h.evaluated_count, 0),
			IFNULL(h.unevaluated_count, 0),
			IFNULL(h.is_visible_student, 0),
			IFNULL(h.source_type, 1),
			IFNULL(h.source_id, 0),
			IFNULL(h.source_name, ''),
			CAST(IFNULL(h.create_id, 0) AS CHAR),
			IFNULL(u.nick_name, ''),
			h.create_time,
			h.publish_time,
			h.end_time,
			IFNULL(h.end_hour, 0),
			IFNULL(h.read_count, 0)
		FROM homework_task h
		LEFT JOIN inst_user u ON u.id = h.create_id AND u.del_flag = 0
		WHERE h.inst_id = ? AND h.del_flag = 0 `+whereSQL+orderBy+`
		LIMIT ? OFFSET ?
	`, dataArgs...)
	if err != nil {
		return model.HomeworkPageResultVO{}, err
	}
	defer rows.Close()

	items := make([]model.HomeworkListItemVO, 0, size)
	for rows.Next() {
		var (
			item             model.HomeworkListItemVO
			dbID             int64
			sourceID         int64
			isVisibleStudent int
			createTime       sql.NullTime
			publishTime      sql.NullTime
			endTime          sql.NullTime
		)
		if err := rows.Scan(
			&dbID,
			&item.Title,
			&item.Content,
			&item.StudentCount,
			&item.UnsubmittedCount,
			&item.RejectedCount,
			&item.SubmittedCount,
			&item.ReSubmittedCount,
			&item.EvaluatedCount,
			&item.UnevaluatedCount,
			&isVisibleStudent,
			&item.SourceType,
			&sourceID,
			&item.SourceName,
			&item.CreatedStaffID,
			&item.CreatedStaffName,
			&createTime,
			&publishTime,
			&endTime,
			&item.EndHour,
			&item.ReadCount,
		); err != nil {
			return model.HomeworkPageResultVO{}, err
		}
		item.ID = strconv.FormatInt(dbID, 10)
		item.SourceID = strconv.FormatInt(sourceID, 10)
		item.IsVisibleStudent = isVisibleStudent != 0
		item.UnreadCount = maxInt(item.StudentCount-item.ReadCount, 0)
		if createTime.Valid {
			t := createTime.Time
			item.CreatedTime = &t
		}
		if publishTime.Valid {
			t := publishTime.Time
			item.PublishTime = &t
		}
		if endTime.Valid {
			t := endTime.Time
			item.EndTime = &t
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return model.HomeworkPageResultVO{}, err
	}
	return model.HomeworkPageResultVO{
		List:  items,
		Total: total,
	}, nil
}

func (repo *Repository) HomeworkStatistics(ctx context.Context, instID int64, query model.HomeworkListQueryModel) (model.HomeworkStatisticsVO, error) {
	whereSQL, args, err := buildHomeworkWhere(query)
	if err != nil {
		return model.HomeworkStatisticsVO{}, err
	}

	row := repo.db.QueryRowContext(ctx, `
		SELECT
			IFNULL(SUM(h.unsubmitted_count), 0),
			IFNULL(SUM(h.unevaluated_count), 0)
		FROM homework_task h
		WHERE h.inst_id = ? AND h.del_flag = 0 `+whereSQL,
		append([]any{instID}, args...)...,
	)

	var result model.HomeworkStatisticsVO
	if err := row.Scan(&result.UnsubmittedCount, &result.UnevaluatedCount); err != nil {
		return model.HomeworkStatisticsVO{}, err
	}
	return result, nil
}

func (repo *Repository) homeworkExists(ctx context.Context, instID, id int64) (bool, error) {
	var count int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM homework_task
		WHERE inst_id = ? AND id = ?
	`, instID, id).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func buildHomeworkWhere(query model.HomeworkListQueryModel) (string, []any, error) {
	parts := make([]string, 0, 8)
	args := make([]any, 0, 12)

	if classID := strings.TrimSpace(query.ClassID); classID != "" {
		value, err := strconv.ParseInt(classID, 10, 64)
		if err != nil || value <= 0 {
			return "", nil, errors.New("班级ID无效")
		}
		parts = append(parts, " AND h.source_type = ? AND h.source_id = ?")
		args = append(args, model.HomeworkSourceTypeClass, value)
	}
	if oneToOneID := strings.TrimSpace(query.OneToOneID); oneToOneID != "" {
		value, err := strconv.ParseInt(oneToOneID, 10, 64)
		if err != nil || value <= 0 {
			return "", nil, errors.New("1对1ID无效")
		}
		parts = append(parts, " AND h.source_type = ? AND h.source_id = ?")
		args = append(args, model.HomeworkSourceTypeOneToOne, value)
	}
	if len(query.TeacherIDs) > 0 {
		ids := sanitizeHomeworkIDs(query.TeacherIDs)
		if len(ids) > 0 {
			placeholders := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
			parts = append(parts, " AND CAST(h.create_id AS CHAR) IN ("+placeholders+")")
			for _, id := range ids {
				args = append(args, id)
			}
		}
	}
	if strings.TrimSpace(query.PublishStartTime) != "" {
		startTime, err := parseHomeworkDateOnly(strings.TrimSpace(query.PublishStartTime))
		if err != nil {
			return "", nil, errors.New("发布时间开始日期无效")
		}
		parts = append(parts, " AND h.publish_time >= ?")
		args = append(args, startTime)
	}
	if strings.TrimSpace(query.PublishEndTime) != "" {
		endTime, err := parseHomeworkDateOnly(strings.TrimSpace(query.PublishEndTime))
		if err != nil {
			return "", nil, errors.New("发布时间结束日期无效")
		}
		parts = append(parts, " AND h.publish_time < ?")
		args = append(args, endTime.AddDate(0, 0, 1))
	}
	if strings.TrimSpace(query.EndStartTime) != "" {
		startTime, err := parseHomeworkDateOnly(strings.TrimSpace(query.EndStartTime))
		if err != nil {
			return "", nil, errors.New("截止时间开始日期无效")
		}
		parts = append(parts, " AND h.end_time >= ?")
		args = append(args, startTime)
	}
	if strings.TrimSpace(query.EndEndTime) != "" {
		endTime, err := parseHomeworkDateOnly(strings.TrimSpace(query.EndEndTime))
		if err != nil {
			return "", nil, errors.New("截止时间结束日期无效")
		}
		parts = append(parts, " AND h.end_time < ?")
		args = append(args, endTime.AddDate(0, 0, 1))
	}
	if query.HasUnevaluated != nil {
		if *query.HasUnevaluated {
			parts = append(parts, " AND IFNULL(h.unevaluated_count, 0) > 0")
		} else {
			parts = append(parts, " AND IFNULL(h.unevaluated_count, 0) = 0")
		}
	}
	if query.HasUnsubmitted != nil {
		if *query.HasUnsubmitted {
			parts = append(parts, " AND IFNULL(h.unsubmitted_count, 0) > 0")
		} else {
			parts = append(parts, " AND IFNULL(h.unsubmitted_count, 0) = 0")
		}
	}

	return strings.Join(parts, ""), args, nil
}

func parseHomeworkDateOnly(value string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", value, time.Local)
}

func marshalHomeworkValue(value any) (string, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func marshalHomeworkOptionalValue(value any) (*string, error) {
	if value == nil {
		return nil, nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	text := string(raw)
	return &text, nil
}

func sanitizeHomeworkIDs(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		text := strings.TrimSpace(value)
		if text == "" {
			continue
		}
		if _, ok := seen[text]; ok {
			continue
		}
		seen[text] = struct{}{}
		result = append(result, text)
	}
	return result
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
