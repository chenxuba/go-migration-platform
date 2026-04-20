package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

func ensureNoticeRecordTables(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS notice_record (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			inst_id BIGINT NOT NULL,
			notice_template_id BIGINT NOT NULL DEFAULT 0,
			title VARCHAR(100) NOT NULL DEFAULT '',
			content LONGTEXT NULL,
			summary LONGTEXT NULL,
			is_all_school TINYINT(1) NOT NULL DEFAULT 0,
			is_delay_send TINYINT(1) NOT NULL DEFAULT 0,
			is_confirm TINYINT(1) NOT NULL DEFAULT 0,
			is_remind TINYINT(1) NOT NULL DEFAULT 0,
			is_withdraw TINYINT(1) NOT NULL DEFAULT 0,
			class_ids_json LONGTEXT NULL,
			classs_json LONGTEXT NULL,
			student_ids_json LONGTEXT NULL,
			target_student_ids_json LONGTEXT NULL,
			student_count INT NOT NULL DEFAULT 0,
			read_student_count INT NOT NULL DEFAULT 0,
			confirm_student_count INT NOT NULL DEFAULT 0,
			operator_id BIGINT NOT NULL DEFAULT 0,
			operator_name VARCHAR(100) NOT NULL DEFAULT '',
			operation_date DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			publish_hour INT NOT NULL DEFAULT 0,
			publish_time DATETIME NULL DEFAULT NULL,
			reality_publish_time DATETIME NULL DEFAULT NULL,
			status INT NOT NULL DEFAULT 4,
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT NOT NULL DEFAULT 0,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			KEY idx_notice_record_inst_deleted (inst_id, del_flag, id),
			KEY idx_notice_record_inst_status (inst_id, status, is_withdraw, id),
			KEY idx_notice_record_inst_publish (inst_id, publish_time, id),
			KEY idx_notice_record_inst_operation (inst_id, operation_date, id),
			KEY idx_notice_record_inst_operator (inst_id, operator_id, id)
		)
	`); err != nil {
		return err
	}

	for _, columns := range []map[string]string{
		{"notice_template_id": "notice_template_id BIGINT NOT NULL DEFAULT 0 AFTER inst_id"},
		{"summary": "summary LONGTEXT NULL AFTER content"},
		{"is_all_school": "is_all_school TINYINT(1) NOT NULL DEFAULT 0 AFTER summary"},
		{"is_delay_send": "is_delay_send TINYINT(1) NOT NULL DEFAULT 0 AFTER is_all_school"},
		{"is_confirm": "is_confirm TINYINT(1) NOT NULL DEFAULT 0 AFTER is_delay_send"},
		{"is_remind": "is_remind TINYINT(1) NOT NULL DEFAULT 0 AFTER is_confirm"},
		{"is_withdraw": "is_withdraw TINYINT(1) NOT NULL DEFAULT 0 AFTER is_remind"},
		{"class_ids_json": "class_ids_json LONGTEXT NULL AFTER is_withdraw"},
		{"classs_json": "classs_json LONGTEXT NULL AFTER class_ids_json"},
		{"student_ids_json": "student_ids_json LONGTEXT NULL AFTER classs_json"},
		{"target_student_ids_json": "target_student_ids_json LONGTEXT NULL AFTER student_ids_json"},
		{"student_count": "student_count INT NOT NULL DEFAULT 0 AFTER target_student_ids_json"},
		{"read_student_count": "read_student_count INT NOT NULL DEFAULT 0 AFTER student_count"},
		{"confirm_student_count": "confirm_student_count INT NOT NULL DEFAULT 0 AFTER read_student_count"},
		{"operator_id": "operator_id BIGINT NOT NULL DEFAULT 0 AFTER confirm_student_count"},
		{"operator_name": "operator_name VARCHAR(100) NOT NULL DEFAULT '' AFTER operator_id"},
		{"operation_date": "operation_date DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP AFTER operator_name"},
		{"publish_hour": "publish_hour INT NOT NULL DEFAULT 0 AFTER operation_date"},
		{"publish_time": "publish_time DATETIME NULL DEFAULT NULL AFTER publish_hour"},
		{"reality_publish_time": "reality_publish_time DATETIME NULL DEFAULT NULL AFTER publish_time"},
		{"status": "status INT NOT NULL DEFAULT 4 AFTER reality_publish_time"},
	} {
		if err := ensureColumnsOnTable(ctx, db, "notice_record", columns); err != nil {
			return err
		}
	}
	return nil
}

func (repo *Repository) ListNoticeClassSnapshots(ctx context.Context, instID int64, classIDs []string) ([]model.NoticeClassSnapshot, error) {
	ids := parseNoticeIDInts(classIDs)
	if len(ids) == 0 {
		return []model.NoticeClassSnapshot{}, nil
	}
	placeholders := sqlPlaceholders(len(ids))
	args := make([]any, 0, 1+len(ids))
	args = append(args, instID)
	for _, id := range ids {
		args = append(args, id)
	}
	rows, err := repo.db.QueryContext(ctx, `
		SELECT CAST(id AS CHAR), IFNULL(name, '')
		FROM teaching_class
		WHERE inst_id = ? AND del_flag = 0 AND id IN (`+placeholders+`)
		ORDER BY create_time DESC, id DESC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.NoticeClassSnapshot, 0, len(ids))
	for rows.Next() {
		var item model.NoticeClassSnapshot
		if err := rows.Scan(&item.ClassID, &item.ClassName); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repo *Repository) ListNoticeTargetStudentIDsByClassIDs(ctx context.Context, instID int64, classIDs []string) ([]string, error) {
	ids := parseNoticeIDInts(classIDs)
	if len(ids) == 0 {
		return []string{}, nil
	}
	placeholders := sqlPlaceholders(len(ids))
	args := make([]any, 0, 5+len(ids))
	args = append(
		args,
		instID,
		model.TeachingClassStudentStatusStudying,
		model.TeachingClassStudentStatusStudying,
		model.TeachingClassStudentStatusStopped,
	)
	for _, id := range ids {
		args = append(args, id)
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT DISTINCT CAST(tcs.student_id AS CHAR)
		FROM teaching_class_student tcs
		INNER JOIN teaching_class tc
			ON tc.id = tcs.teaching_class_id
			AND tc.inst_id = tcs.inst_id
			AND tc.del_flag = 0
		WHERE tcs.inst_id = ?
		  AND tcs.del_flag = 0
		  AND IFNULL(tcs.class_student_status, ?) IN (?, ?)
		  AND tcs.teaching_class_id IN (`+placeholders+`)
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]string, 0, len(ids)*2)
	for rows.Next() {
		var studentID string
		if err := rows.Scan(&studentID); err != nil {
			return nil, err
		}
		result = append(result, strings.TrimSpace(studentID))
	}
	return result, rows.Err()
}

func (repo *Repository) ListExistingNoticeStudentIDs(ctx context.Context, instID int64, studentIDs []string) ([]string, error) {
	ids := parseNoticeIDInts(studentIDs)
	if len(ids) == 0 {
		return []string{}, nil
	}
	placeholders := sqlPlaceholders(len(ids))
	args := make([]any, 0, 2+len(ids))
	args = append(args, instID, model.InstStudentStatusEnrolled)
	for _, id := range ids {
		args = append(args, id)
	}
	rows, err := repo.db.QueryContext(ctx, `
		SELECT CAST(id AS CHAR)
		FROM inst_student
		WHERE inst_id = ? AND del_flag = 0 AND student_status = ? AND id IN (`+placeholders+`)
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]string, 0, len(ids))
	for rows.Next() {
		var studentID string
		if err := rows.Scan(&studentID); err != nil {
			return nil, err
		}
		result = append(result, strings.TrimSpace(studentID))
	}
	return result, rows.Err()
}

func (repo *Repository) ListAllNoticeStudentIDs(ctx context.Context, instID int64) ([]string, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT CAST(id AS CHAR)
		FROM inst_student
		WHERE inst_id = ? AND del_flag = 0 AND student_status = ?
		ORDER BY id DESC
	`, instID, model.InstStudentStatusEnrolled)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]string, 0, 64)
	for rows.Next() {
		var studentID string
		if err := rows.Scan(&studentID); err != nil {
			return nil, err
		}
		result = append(result, strings.TrimSpace(studentID))
	}
	return result, rows.Err()
}

func (repo *Repository) HasRepeatNoticeStudentsOnDate(ctx context.Context, instID int64, studentIDs []string, day string) (bool, error) {
	normalized := uniqueTrimmedStrings(studentIDs)
	if len(normalized) == 0 {
		return false, nil
	}
	studentSet := make(map[string]struct{}, len(normalized))
	for _, studentID := range normalized {
		studentSet[studentID] = struct{}{}
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT IFNULL(is_all_school, 0), IFNULL(target_student_ids_json, '')
		FROM notice_record
		WHERE inst_id = ? AND del_flag = 0 AND IFNULL(is_withdraw, 0) = 0
		  AND DATE(COALESCE(reality_publish_time, publish_time, operation_date)) = ?
	`, instID, strings.TrimSpace(day))
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			isAllSchool        int
			targetStudentsJSON string
		)
		if err := rows.Scan(&isAllSchool, &targetStudentsJSON); err != nil {
			return false, err
		}
		if isAllSchool != 0 {
			return true, nil
		}
		existing := make([]string, 0)
		if strings.TrimSpace(targetStudentsJSON) != "" {
			if err := json.Unmarshal([]byte(targetStudentsJSON), &existing); err != nil {
				return false, err
			}
		}
		for _, studentID := range existing {
			if _, ok := studentSet[strings.TrimSpace(studentID)]; ok {
				return true, nil
			}
		}
	}
	return false, rows.Err()
}

func (repo *Repository) CreateNoticeRecord(ctx context.Context, instID int64, input model.NoticeCreateInput) (int64, error) {
	classIDsJSON, err := marshalNoticeJSON(input.ClassIDs)
	if err != nil {
		return 0, err
	}
	classsJSON, err := marshalNoticeJSON(input.ClassSnapshots)
	if err != nil {
		return 0, err
	}
	studentIDsJSON, err := marshalNoticeJSON(input.ExplicitStudentIDs)
	if err != nil {
		return 0, err
	}
	targetStudentIDsJSON, err := marshalNoticeJSON(input.TargetStudentIDs)
	if err != nil {
		return 0, err
	}

	result, err := repo.db.ExecContext(ctx, `
		INSERT INTO notice_record (
			inst_id, notice_template_id, title, content, summary, is_all_school, is_delay_send,
			is_confirm, is_remind, is_withdraw, class_ids_json, classs_json, student_ids_json,
			target_student_ids_json, student_count, read_student_count, confirm_student_count,
			operator_id, operator_name, operation_date, publish_hour, publish_time,
			reality_publish_time, status, create_id, create_time, update_id, update_time, del_flag
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, 0, ?, ?, ?, ?, ?, 0, 0, ?, ?, NOW(), ?, ?, ?, ?, ?, NOW(), ?, NOW(), 0)
	`,
		instID,
		input.NoticeTemplateID,
		strings.TrimSpace(input.Title),
		strings.TrimSpace(input.Content),
		strings.TrimSpace(input.Summary),
		boolToTinyInt(input.IsAllSchool),
		boolToTinyInt(input.IsDelaySend),
		boolToTinyInt(input.IsConfirm),
		classIDsJSON,
		classsJSON,
		studentIDsJSON,
		targetStudentIDsJSON,
		input.StudentCount,
		input.OperatorID,
		strings.TrimSpace(input.OperatorName),
		input.PublishHour,
		input.PublishTime,
		input.RealityPublishTime,
		input.Status,
		input.OperatorID,
		input.OperatorID,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (repo *Repository) WithdrawNoticeRecord(ctx context.Context, instID, operatorID, noticeID int64) error {
	result, err := repo.db.ExecContext(ctx, `
		UPDATE notice_record
		SET is_withdraw = 1, update_id = ?, update_time = NOW()
		WHERE inst_id = ? AND id = ? AND del_flag = 0 AND IFNULL(is_withdraw, 0) = 0
	`, operatorID, instID, noticeID)
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
	exists, err := repo.noticeRecordExists(ctx, instID, noticeID)
	if err != nil {
		return err
	}
	if !exists {
		return sql.ErrNoRows
	}
	return nil
}

func (repo *Repository) PageNoticeRecords(ctx context.Context, instID int64, query model.NoticePageQueryDTO) (model.NoticePageResultVO, error) {
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

	whereSQL, args := buildNoticeRecordWhere(query.QueryModel)
	countArgs := append([]any{instID}, args...)
	var total int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM notice_record
		WHERE inst_id = ? AND del_flag = 0 `+whereSQL,
		countArgs...,
	).Scan(&total); err != nil {
		return model.NoticePageResultVO{}, err
	}

	dataArgs := append([]any{instID}, args...)
	dataArgs = append(dataArgs, size, offset)
	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			id,
			IFNULL(title, ''),
			IFNULL(content, ''),
			IFNULL(summary, ''),
			IFNULL(is_all_school, 0),
			IFNULL(is_confirm, 0),
			IFNULL(is_remind, 0),
			IFNULL(is_withdraw, 0),
			IFNULL(classs_json, ''),
			IFNULL(student_count, 0),
			IFNULL(read_student_count, 0),
			IFNULL(confirm_student_count, 0),
			CAST(IFNULL(operator_id, 0) AS CHAR),
			IFNULL(operator_name, ''),
			operation_date,
			IFNULL(is_delay_send, 0),
			publish_time,
			IFNULL(status, 0),
			reality_publish_time
		FROM notice_record
		WHERE inst_id = ? AND del_flag = 0 `+whereSQL+`
		ORDER BY operation_date DESC, id DESC
		LIMIT ? OFFSET ?
	`, dataArgs...)
	if err != nil {
		return model.NoticePageResultVO{}, err
	}
	defer rows.Close()

	items := make([]model.NoticePageItemVO, 0, size)
	for rows.Next() {
		var (
			item           model.NoticePageItemVO
			dbID           int64
			isAllSchool    int
			isConfirm      int
			isRemind       int
			isWithdraw     int
			classsJSON     string
			isDelaySend    int
			operationDate  sql.NullTime
			publishTime    sql.NullTime
			realityPublish sql.NullTime
		)
		if err := rows.Scan(
			&dbID,
			&item.Title,
			&item.Content,
			&item.Summary,
			&isAllSchool,
			&isConfirm,
			&isRemind,
			&isWithdraw,
			&classsJSON,
			&item.StudentCount,
			&item.ReadStudentCount,
			&item.ConfirmStudentCount,
			&item.OperatorID,
			&item.OperatorName,
			&operationDate,
			&isDelaySend,
			&publishTime,
			&item.Status,
			&realityPublish,
		); err != nil {
			return model.NoticePageResultVO{}, err
		}

		item.NoticeID = strconv.FormatInt(dbID, 10)
		item.IsAllSchool = isAllSchool != 0
		item.IsConfirm = isConfirm != 0
		item.IsRemind = isRemind != 0
		item.IsWithdraw = isWithdraw != 0
		item.IsDelaySend = isDelaySend != 0
		if operationDate.Valid {
			t := operationDate.Time
			item.OperationDate = &t
		}
		if publishTime.Valid {
			t := publishTime.Time
			item.PublishTime = &t
		}
		if realityPublish.Valid {
			t := realityPublish.Time
			item.RealityPublishTime = &t
		}

		snapshots := make([]model.NoticeClassSnapshot, 0)
		if strings.TrimSpace(classsJSON) != "" {
			if err := json.Unmarshal([]byte(classsJSON), &snapshots); err != nil {
				return model.NoticePageResultVO{}, err
			}
		}
		item.Classs = make([]model.NoticeClassVO, 0, len(snapshots))
		for _, snapshot := range snapshots {
			item.Classs = append(item.Classs, model.NoticeClassVO{
				ClassID:   strings.TrimSpace(snapshot.ClassID),
				NoticeID:  item.NoticeID,
				ClassName: strings.TrimSpace(snapshot.ClassName),
			})
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return model.NoticePageResultVO{}, err
	}

	return model.NoticePageResultVO{
		List:  items,
		Total: total,
	}, nil
}

func (repo *Repository) noticeRecordExists(ctx context.Context, instID, noticeID int64) (bool, error) {
	var count int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM notice_record
		WHERE inst_id = ? AND id = ? AND del_flag = 0
	`, instID, noticeID).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func buildNoticeRecordWhere(query model.NoticePageQueryModel) (string, []any) {
	var (
		builder strings.Builder
		args    []any
	)
	if query.IsWithdraw != nil {
		builder.WriteString(" AND IFNULL(is_withdraw, 0) = ?")
		args = append(args, boolToTinyInt(*query.IsWithdraw))
	}
	if statuses := uniqueInts(query.Statuses); len(statuses) > 0 {
		builder.WriteString(" AND status IN (")
		for idx, status := range statuses {
			if idx > 0 {
				builder.WriteString(",")
			}
			builder.WriteString("?")
			args = append(args, status)
		}
		builder.WriteString(")")
	}
	if operatorID, _ := strconv.ParseInt(strings.TrimSpace(query.OperatorID), 10, 64); operatorID > 0 {
		builder.WriteString(" AND operator_id = ?")
		args = append(args, operatorID)
	}
	if beginDate := strings.TrimSpace(query.BeginPublishDate); beginDate != "" {
		builder.WriteString(" AND DATE(COALESCE(reality_publish_time, publish_time, operation_date)) >= ?")
		args = append(args, beginDate)
	}
	if endDate := strings.TrimSpace(query.EndPublishDate); endDate != "" {
		builder.WriteString(" AND DATE(COALESCE(reality_publish_time, publish_time, operation_date)) <= ?")
		args = append(args, endDate)
	}
	return builder.String(), args
}

func marshalNoticeJSON(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func parseNoticeIDInts(values []string) []int64 {
	result := make([]int64, 0, len(values))
	seen := make(map[int64]struct{}, len(values))
	for _, raw := range values {
		value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
		if err != nil || value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func uniqueTrimmedStrings(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, raw := range values {
		value := strings.TrimSpace(raw)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func uniqueInts(values []int) []int {
	result := make([]int, 0, len(values))
	seen := make(map[int]struct{}, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
