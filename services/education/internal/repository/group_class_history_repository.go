package repository

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

func ensureGroupClassHistoryTables(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS teaching_class_operation_log (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			uuid VARCHAR(64) NULL,
			version BIGINT NOT NULL DEFAULT 0,
			inst_id BIGINT NOT NULL,
			teaching_class_id BIGINT NOT NULL,
			teaching_class_student_id BIGINT NOT NULL DEFAULT 0,
			student_id BIGINT NOT NULL DEFAULT 0,
			operation_type INT NOT NULL DEFAULT 0,
			operation_content VARCHAR(500) NOT NULL DEFAULT '',
			operator_id BIGINT NOT NULL DEFAULT 0,
			operate_time DATETIME NOT NULL,
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT NOT NULL DEFAULT 0,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uk_teaching_class_operation_log_unique (inst_id, teaching_class_student_id, operation_type),
			KEY idx_teaching_class_operation_log_class (inst_id, teaching_class_id, operate_time),
			KEY idx_teaching_class_operation_log_student (inst_id, student_id, operate_time)
		)
	`); err != nil {
		return err
	}
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS teaching_class_entry_exit_record (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			uuid VARCHAR(64) NULL,
			version BIGINT NOT NULL DEFAULT 0,
			inst_id BIGINT NOT NULL,
			teaching_class_id BIGINT NOT NULL,
			teaching_class_student_id BIGINT NOT NULL DEFAULT 0,
			student_id BIGINT NOT NULL DEFAULT 0,
			entry_exit_status INT NOT NULL DEFAULT 1,
			entry_exit_time DATETIME NOT NULL,
			operator_id BIGINT NOT NULL DEFAULT 0,
			operate_time DATETIME NOT NULL,
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT NOT NULL DEFAULT 0,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uk_teaching_class_entry_exit_unique (inst_id, teaching_class_student_id, entry_exit_status),
			KEY idx_teaching_class_entry_exit_class (inst_id, teaching_class_id, entry_exit_time),
			KEY idx_teaching_class_entry_exit_student (inst_id, student_id, entry_exit_time)
		)
	`)
	return err
}

func groupClassOperationTypeText(operationType int) string {
	switch operationType {
	case model.GroupClassOperationTypeSwitchTuitionAccount:
		return "切换默认学费账户"
	case model.GroupClassOperationTypeStatusChange:
		return "在班状态变更"
	case model.GroupClassOperationTypeRemoveStudent:
		return "移出班级学员"
	case model.GroupClassOperationTypeAddStudent:
		return "添加班级学员"
	default:
		return "未知操作"
	}
}

func groupClassEntryExitStatusText(status int) string {
	switch status {
	case model.GroupClassEntryExitStatusIn:
		return "入班"
	case model.GroupClassEntryExitStatusOut:
		return "出班"
	default:
		return "-"
	}
}

func teachingClassStudentStatusTextForLog(status int) string {
	switch status {
	case model.TeachingClassStudentStatusStudying:
		return "在读"
	case model.TeachingClassStudentStatusStopped:
		return "停课"
	case model.TeachingClassStudentStatusClosed:
		return "结课"
	default:
		return "在读"
	}
}

func normalizeGroupClassOperationTypes(types []int) []int {
	if len(types) == 0 {
		return nil
	}
	seen := make(map[int]struct{}, len(types))
	out := make([]int, 0, len(types))
	for _, item := range types {
		if item < model.GroupClassOperationTypeSwitchTuitionAccount || item > model.GroupClassOperationTypeAddStudent {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func normalizeGroupClassEntryExitStatuses(statuses []int) []int {
	if len(statuses) == 0 {
		return nil
	}
	seen := make(map[int]struct{}, len(statuses))
	out := make([]int, 0, len(statuses))
	for _, item := range statuses {
		if item != model.GroupClassEntryExitStatusIn && item != model.GroupClassEntryExitStatusOut {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func parseDayStart(value string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", strings.TrimSpace(value), time.Local)
}

func parseDayEnd(value string) (time.Time, error) {
	day, err := parseDayStart(value)
	if err != nil {
		return time.Time{}, err
	}
	return day.Add(24*time.Hour - time.Nanosecond), nil
}

func normalizeGroupClassHistoryPage(page model.GroupClassPageRequestModel) (pageSize, pageIndex, offset int) {
	pageSize = page.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	pageIndex = page.PageIndex
	if pageIndex <= 0 {
		pageIndex = 1
	}
	offset = (pageIndex - 1) * pageSize
	if page.SkipCount > 0 {
		offset = page.SkipCount
	}
	return
}

func (repo *Repository) ensureGroupClassHistoryBackfilled(ctx context.Context, instID, classID int64) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := repo.backfillGroupClassHistoryTx(ctx, tx, instID, classID); err != nil {
		return err
	}
	return tx.Commit()
}

func (repo *Repository) backfillGroupClassHistoryTx(ctx context.Context, tx *sql.Tx, instID, classID int64) error {
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO teaching_class_entry_exit_record (
			uuid, version, inst_id, teaching_class_id, teaching_class_student_id, student_id,
			entry_exit_status, entry_exit_time, operator_id, operate_time,
			create_id, create_time, update_id, update_time, del_flag
		)
		SELECT
			UUID(), 0, tcs.inst_id, tcs.teaching_class_id, tcs.id, tcs.student_id,
			?, tcs.create_time, IFNULL(tcs.create_id, 0), tcs.create_time,
			IFNULL(tcs.create_id, 0), tcs.create_time, IFNULL(tcs.create_id, 0), tcs.create_time, 0
		FROM teaching_class_student tcs
		LEFT JOIN teaching_class_entry_exit_record rec
			ON rec.inst_id = tcs.inst_id
			AND rec.teaching_class_student_id = tcs.id
			AND rec.entry_exit_status = ?
			AND rec.del_flag = 0
		WHERE tcs.inst_id = ? AND tcs.teaching_class_id = ? AND tcs.del_flag = 0
			AND rec.id IS NULL
	`, model.GroupClassEntryExitStatusIn, model.GroupClassEntryExitStatusIn, instID, classID); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO teaching_class_entry_exit_record (
			uuid, version, inst_id, teaching_class_id, teaching_class_student_id, student_id,
			entry_exit_status, entry_exit_time, operator_id, operate_time,
			create_id, create_time, update_id, update_time, del_flag
		)
		SELECT
			UUID(), 0, tcs.inst_id, tcs.teaching_class_id, tcs.id, tcs.student_id,
			?,
			CASE
				WHEN IFNULL(tcs.class_student_status, 1) = ? THEN COALESCE(ta.suspended_time, tcs.update_time, tcs.create_time)
				WHEN IFNULL(tcs.class_student_status, 1) = ? THEN COALESCE(ta.class_ending_time, tcs.update_time, tcs.create_time)
				ELSE COALESCE(tcs.update_time, tcs.create_time)
			END,
			IFNULL(NULLIF(tcs.update_id, 0), IFNULL(tcs.create_id, 0)),
			COALESCE(tcs.update_time, tcs.create_time),
			IFNULL(tcs.create_id, 0), tcs.create_time,
			IFNULL(NULLIF(tcs.update_id, 0), IFNULL(tcs.create_id, 0)),
			COALESCE(tcs.update_time, tcs.create_time),
			0
		FROM teaching_class_student tcs
		LEFT JOIN tuition_account ta ON ta.id = tcs.primary_tuition_account_id AND ta.inst_id = tcs.inst_id AND ta.del_flag = 0
		LEFT JOIN teaching_class_entry_exit_record rec
			ON rec.inst_id = tcs.inst_id
			AND rec.teaching_class_student_id = tcs.id
			AND rec.entry_exit_status = ?
			AND rec.del_flag = 0
		WHERE tcs.inst_id = ? AND tcs.teaching_class_id = ? AND tcs.del_flag = 0
			AND IFNULL(tcs.class_student_status, 1) IN (?, ?)
			AND rec.id IS NULL
	`,
		model.GroupClassEntryExitStatusOut,
		model.TeachingClassStudentStatusStopped,
		model.TeachingClassStudentStatusClosed,
		model.GroupClassEntryExitStatusOut,
		instID,
		classID,
		model.TeachingClassStudentStatusStopped,
		model.TeachingClassStudentStatusClosed,
	); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO teaching_class_operation_log (
			uuid, version, inst_id, teaching_class_id, teaching_class_student_id, student_id,
			operation_type, operation_content, operator_id, operate_time,
			create_id, create_time, update_id, update_time, del_flag
		)
		SELECT
			UUID(), 0, tcs.inst_id, tcs.teaching_class_id, tcs.id, tcs.student_id,
			?,
			CONCAT(
				COALESCE(NULLIF(s.stu_name, ''), CONCAT('学员', CAST(tcs.student_id AS CHAR))),
				'入班，学费账户为：',
				COALESCE(NULLIF(ic.name, ''), CONCAT('账户', CAST(IFNULL(ta.id, 0) AS CHAR))),
				'，在班状态为：',
				CASE IFNULL(tcs.class_student_status, 1)
					WHEN ? THEN '在读'
					WHEN ? THEN '停课'
					WHEN ? THEN '结课'
					ELSE '在读'
				END
			),
			IFNULL(tcs.create_id, 0),
			tcs.create_time,
			IFNULL(tcs.create_id, 0), tcs.create_time, IFNULL(tcs.create_id, 0), tcs.create_time, 0
		FROM teaching_class_student tcs
		LEFT JOIN inst_student s ON s.id = tcs.student_id AND s.inst_id = tcs.inst_id AND s.del_flag = 0
		LEFT JOIN tuition_account ta ON ta.id = tcs.primary_tuition_account_id AND ta.inst_id = tcs.inst_id AND ta.del_flag = 0
		LEFT JOIN inst_course ic ON ic.id = ta.course_id AND ic.inst_id = ta.inst_id AND ic.del_flag = 0
		LEFT JOIN teaching_class_operation_log log
			ON log.inst_id = tcs.inst_id
			AND log.teaching_class_student_id = tcs.id
			AND log.operation_type = ?
			AND log.del_flag = 0
		WHERE tcs.inst_id = ? AND tcs.teaching_class_id = ? AND tcs.del_flag = 0
			AND log.id IS NULL
	`,
		model.GroupClassOperationTypeAddStudent,
		model.TeachingClassStudentStatusStudying,
		model.TeachingClassStudentStatusStopped,
		model.TeachingClassStudentStatusClosed,
		model.GroupClassOperationTypeAddStudent,
		instID,
		classID,
	); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO teaching_class_operation_log (
			uuid, version, inst_id, teaching_class_id, teaching_class_student_id, student_id,
			operation_type, operation_content, operator_id, operate_time,
			create_id, create_time, update_id, update_time, del_flag
		)
		SELECT
			UUID(), 0, tcs.inst_id, tcs.teaching_class_id, tcs.id, tcs.student_id,
			CASE
				WHEN IFNULL(tcs.class_student_status, 1) = ? THEN ?
				ELSE ?
			END,
			CASE
				WHEN IFNULL(tcs.class_student_status, 1) = ? THEN CONCAT(
					COALESCE(NULLIF(s.stu_name, ''), CONCAT('学员', CAST(tcs.student_id AS CHAR))),
					'移出班级，在班状态为：结课'
				)
				ELSE CONCAT(
					COALESCE(NULLIF(s.stu_name, ''), CONCAT('学员', CAST(tcs.student_id AS CHAR))),
					'在班状态变更为：停课'
				)
			END,
			IFNULL(NULLIF(tcs.update_id, 0), IFNULL(tcs.create_id, 0)),
			COALESCE(tcs.update_time, tcs.create_time),
			IFNULL(tcs.create_id, 0), tcs.create_time,
			IFNULL(NULLIF(tcs.update_id, 0), IFNULL(tcs.create_id, 0)),
			COALESCE(tcs.update_time, tcs.create_time),
			0
		FROM teaching_class_student tcs
		LEFT JOIN inst_student s ON s.id = tcs.student_id AND s.inst_id = tcs.inst_id AND s.del_flag = 0
		LEFT JOIN teaching_class_operation_log log
			ON log.inst_id = tcs.inst_id
			AND log.teaching_class_student_id = tcs.id
			AND log.operation_type = CASE
				WHEN IFNULL(tcs.class_student_status, 1) = ? THEN ?
				ELSE ?
			END
			AND log.del_flag = 0
		WHERE tcs.inst_id = ? AND tcs.teaching_class_id = ? AND tcs.del_flag = 0
			AND IFNULL(tcs.class_student_status, 1) IN (?, ?)
			AND log.id IS NULL
	`,
		model.TeachingClassStudentStatusClosed,
		model.GroupClassOperationTypeRemoveStudent,
		model.GroupClassOperationTypeStatusChange,
		model.TeachingClassStudentStatusClosed,
		model.TeachingClassStudentStatusClosed,
		model.GroupClassOperationTypeRemoveStudent,
		model.GroupClassOperationTypeStatusChange,
		instID,
		classID,
		model.TeachingClassStudentStatusStopped,
		model.TeachingClassStudentStatusClosed,
	); err != nil {
		return err
	}

	return nil
}

func (repo *Repository) PageGroupClassOperationLogs(ctx context.Context, instID int64, body model.GroupClassOperationLogPagedListBody) (model.GroupClassOperationLogPagedListResult, error) {
	out := model.GroupClassOperationLogPagedListResult{List: []model.GroupClassOperationLogItemVO{}}
	classID, err := strconv.ParseInt(strings.TrimSpace(body.QueryModel.ClassID), 10, 64)
	if err != nil || classID <= 0 {
		return out, errors.New("classId 无效")
	}
	if err := repo.ensureGroupClassHistoryBackfilled(ctx, instID, classID); err != nil {
		return out, err
	}

	filters := []string{"log.inst_id = ?", "log.teaching_class_id = ?", "log.del_flag = 0"}
	args := []any{instID, classID}

	if studentID := strings.TrimSpace(body.QueryModel.StudentID); studentID != "" {
		value, err := strconv.ParseInt(studentID, 10, 64)
		if err != nil || value <= 0 {
			return out, errors.New("studentId 无效")
		}
		filters = append(filters, "log.student_id = ?")
		args = append(args, value)
	}
	if operatorID := strings.TrimSpace(body.QueryModel.OperatorID); operatorID != "" {
		value, err := strconv.ParseInt(operatorID, 10, 64)
		if err != nil || value <= 0 {
			return out, errors.New("operatorId 无效")
		}
		filters = append(filters, "log.operator_id = ?")
		args = append(args, value)
	}
	if startAt := strings.TrimSpace(body.QueryModel.OperateStartAt); startAt != "" {
		startTime, err := parseDayStart(startAt)
		if err != nil {
			return out, errors.New("operateStartAt 无效")
		}
		filters = append(filters, "log.operate_time >= ?")
		args = append(args, startTime)
	}
	if endAt := strings.TrimSpace(body.QueryModel.OperateEndAt); endAt != "" {
		endTime, err := parseDayEnd(endAt)
		if err != nil {
			return out, errors.New("operateEndAt 无效")
		}
		filters = append(filters, "log.operate_time <= ?")
		args = append(args, endTime)
	}
	if operationTypes := normalizeGroupClassOperationTypes(body.QueryModel.OperationTypes); len(operationTypes) > 0 {
		filters = append(filters, "log.operation_type IN ("+sqlPlaceholders(len(operationTypes))+")")
		args = append(args, intSliceToAny(operationTypes)...)
	}

	countSQL := `SELECT COUNT(*) FROM teaching_class_operation_log log WHERE ` + strings.Join(filters, " AND ")
	if err := repo.db.QueryRowContext(ctx, countSQL, args...).Scan(&out.Total); err != nil {
		return out, err
	}
	if out.Total == 0 {
		return out, nil
	}

	pageSize, _, offset := normalizeGroupClassHistoryPage(body.PageRequestModel)
	listArgs := append([]any{}, args...)
	listArgs = append(listArgs, pageSize, offset)

	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			CAST(log.id AS CHAR),
			log.operate_time,
			CAST(log.student_id AS CHAR),
			IFNULL(s.stu_name, ''),
			IFNULL(log.operation_type, 0),
			IFNULL(log.operation_content, ''),
			CAST(IFNULL(log.operator_id, 0) AS CHAR),
			IFNULL(op.nick_name, '')
		FROM teaching_class_operation_log log
		LEFT JOIN inst_student s ON s.id = log.student_id AND s.inst_id = log.inst_id AND s.del_flag = 0
		LEFT JOIN inst_user op ON op.id = log.operator_id AND op.inst_id = log.inst_id AND op.del_flag = 0
		WHERE `+strings.Join(filters, " AND ")+`
		ORDER BY log.operate_time DESC, log.id DESC
		LIMIT ? OFFSET ?
	`, listArgs...)
	if err != nil {
		return out, err
	}
	defer rows.Close()

	for rows.Next() {
		var item model.GroupClassOperationLogItemVO
		if err := rows.Scan(
			&item.ID,
			&item.OperateTime,
			&item.StudentID,
			&item.StudentName,
			&item.OperationType,
			&item.OperationContent,
			&item.OperatorID,
			&item.OperatorName,
		); err != nil {
			return out, err
		}
		item.OperationTypeText = groupClassOperationTypeText(item.OperationType)
		out.List = append(out.List, item)
	}
	if err := rows.Err(); err != nil {
		return out, err
	}
	return out, nil
}

func (repo *Repository) PageGroupClassEntryExitRecords(ctx context.Context, instID int64, body model.GroupClassEntryExitRecordPagedListBody) (model.GroupClassEntryExitRecordPagedListResult, error) {
	out := model.GroupClassEntryExitRecordPagedListResult{List: []model.GroupClassEntryExitRecordItemVO{}}
	classID, err := strconv.ParseInt(strings.TrimSpace(body.QueryModel.ClassID), 10, 64)
	if err != nil || classID <= 0 {
		return out, errors.New("classId 无效")
	}
	if err := repo.ensureGroupClassHistoryBackfilled(ctx, instID, classID); err != nil {
		return out, err
	}

	filters := []string{"rec.inst_id = ?", "rec.teaching_class_id = ?", "rec.del_flag = 0"}
	args := []any{instID, classID}

	if studentID := strings.TrimSpace(body.QueryModel.StudentID); studentID != "" {
		value, err := strconv.ParseInt(studentID, 10, 64)
		if err != nil || value <= 0 {
			return out, errors.New("studentId 无效")
		}
		filters = append(filters, "rec.student_id = ?")
		args = append(args, value)
	}
	if startAt := strings.TrimSpace(body.QueryModel.RecordStartDate); startAt != "" {
		startTime, err := parseDayStart(startAt)
		if err != nil {
			return out, errors.New("recordStartDate 无效")
		}
		filters = append(filters, "rec.entry_exit_time >= ?")
		args = append(args, startTime)
	}
	if endAt := strings.TrimSpace(body.QueryModel.RecordEndDate); endAt != "" {
		endTime, err := parseDayEnd(endAt)
		if err != nil {
			return out, errors.New("recordEndDate 无效")
		}
		filters = append(filters, "rec.entry_exit_time <= ?")
		args = append(args, endTime)
	}
	if statuses := normalizeGroupClassEntryExitStatuses(body.QueryModel.EntryExitStatuses); len(statuses) > 0 {
		filters = append(filters, "rec.entry_exit_status IN ("+sqlPlaceholders(len(statuses))+")")
		args = append(args, intSliceToAny(statuses)...)
	}

	countSQL := `SELECT COUNT(*) FROM teaching_class_entry_exit_record rec WHERE ` + strings.Join(filters, " AND ")
	if err := repo.db.QueryRowContext(ctx, countSQL, args...).Scan(&out.Total); err != nil {
		return out, err
	}
	if out.Total == 0 {
		return out, nil
	}
	studentCountSQL := `SELECT COUNT(DISTINCT rec.student_id) FROM teaching_class_entry_exit_record rec WHERE ` + strings.Join(filters, " AND ")
	if err := repo.db.QueryRowContext(ctx, studentCountSQL, args...).Scan(&out.StudentCount); err != nil {
		return out, err
	}

	pageSize, _, offset := normalizeGroupClassHistoryPage(body.PageRequestModel)
	listArgs := append([]any{}, args...)
	listArgs = append(listArgs, pageSize, offset)

	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			CAST(rec.id AS CHAR),
			CAST(rec.student_id AS CHAR),
			IFNULL(s.stu_name, ''),
			IFNULL(s.avatar_url, ''),
			IFNULL(s.mobile, ''),
			IFNULL(s.phone_relationship, 0),
			IFNULL(rec.entry_exit_status, 0),
			rec.entry_exit_time,
			CAST(IFNULL(rec.operator_id, 0) AS CHAR),
			IFNULL(op.nick_name, ''),
			rec.operate_time,
			(
				SELECT MAX(prev.entry_exit_time)
				FROM teaching_class_entry_exit_record prev
				WHERE prev.inst_id = rec.inst_id
					AND prev.teaching_class_id = rec.teaching_class_id
					AND prev.student_id = rec.student_id
					AND prev.del_flag = 0
					AND (
						prev.entry_exit_time < rec.entry_exit_time
						OR (prev.entry_exit_time = rec.entry_exit_time AND prev.id < rec.id)
					)
			),
			(
				SELECT MIN(next.entry_exit_time)
				FROM teaching_class_entry_exit_record next
				WHERE next.inst_id = rec.inst_id
					AND next.teaching_class_id = rec.teaching_class_id
					AND next.student_id = rec.student_id
					AND next.del_flag = 0
					AND (
						next.entry_exit_time > rec.entry_exit_time
						OR (next.entry_exit_time = rec.entry_exit_time AND next.id > rec.id)
					)
			)
		FROM teaching_class_entry_exit_record rec
		LEFT JOIN inst_student s ON s.id = rec.student_id AND s.inst_id = rec.inst_id AND s.del_flag = 0
		LEFT JOIN inst_user op ON op.id = rec.operator_id AND op.inst_id = rec.inst_id AND op.del_flag = 0
		WHERE `+strings.Join(filters, " AND ")+`
		ORDER BY rec.entry_exit_time DESC, rec.id DESC
		LIMIT ? OFFSET ?
	`, listArgs...)
	if err != nil {
		return out, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			item               model.GroupClassEntryExitRecordItemVO
			avatar             string
			mobile             string
			previousRecordTime sql.NullTime
			nextRecordTime     sql.NullTime
		)
		if err := rows.Scan(
			&item.ID,
			&item.StudentID,
			&item.StudentName,
			&avatar,
			&mobile,
			&item.PhoneRelationship,
			&item.EntryExitStatus,
			&item.EntryExitTime,
			&item.OperatorID,
			&item.OperatorName,
			&item.OperateTime,
			&previousRecordTime,
			&nextRecordTime,
		); err != nil {
			return out, err
		}
		item.Avatar = strings.TrimSpace(avatar)
		item.Phone = maskPhoneDisplay(mobile)
		item.EntryExitStatusText = groupClassEntryExitStatusText(item.EntryExitStatus)
		if previousRecordTime.Valid {
			t := previousRecordTime.Time
			item.PreviousRecordTime = &t
		}
		if nextRecordTime.Valid {
			t := nextRecordTime.Time
			item.NextRecordTime = &t
		}
		out.List = append(out.List, item)
	}
	if err := rows.Err(); err != nil {
		return out, err
	}
	return out, nil
}

func (repo *Repository) UpdateGroupClassEntryExitRecordTime(ctx context.Context, instID, operatorID int64, dto model.GroupClassEntryExitRecordUpdateDTO) error {
	recordID, err := strconv.ParseInt(strings.TrimSpace(dto.ID), 10, 64)
	if err != nil || recordID <= 0 {
		return errors.New("id 无效")
	}
	entryExitDate, err := parseDayStart(dto.EntryExitTime)
	if err != nil {
		return errors.New("entryExitTime 无效")
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var (
		currentTime            time.Time
		classID                int64
		studentID              int64
		entryExitStatus        int
		teachingClassStudentID int64
	)
	err = tx.QueryRowContext(ctx, `
		SELECT teaching_class_id, student_id, entry_exit_status, entry_exit_time, teaching_class_student_id
		FROM teaching_class_entry_exit_record
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, recordID, instID).Scan(&classID, &studentID, &entryExitStatus, &currentTime, &teachingClassStudentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("出入班记录不存在")
		}
		return err
	}

	targetTime := time.Date(
		entryExitDate.Year(),
		entryExitDate.Month(),
		entryExitDate.Day(),
		currentTime.Hour(),
		currentTime.Minute(),
		currentTime.Second(),
		currentTime.Nanosecond(),
		currentTime.Location(),
	)

	var previousTime sql.NullTime
	if err := tx.QueryRowContext(ctx, `
		SELECT MAX(prev.entry_exit_time)
		FROM teaching_class_entry_exit_record prev
		WHERE prev.inst_id = ? AND prev.teaching_class_id = ? AND prev.student_id = ? AND prev.del_flag = 0
			AND (
				prev.entry_exit_time < ?
				OR (prev.entry_exit_time = ? AND prev.id < ?)
			)
	`, instID, classID, studentID, currentTime, currentTime, recordID).Scan(&previousTime); err != nil {
		return err
	}
	if previousTime.Valid && !targetTime.After(previousTime.Time) {
		return errors.New("日期调整范围只能在上一条记录与下一条记录之间")
	}

	var nextTime sql.NullTime
	if err := tx.QueryRowContext(ctx, `
		SELECT MIN(next.entry_exit_time)
		FROM teaching_class_entry_exit_record next
		WHERE next.inst_id = ? AND next.teaching_class_id = ? AND next.student_id = ? AND next.del_flag = 0
			AND (
				next.entry_exit_time > ?
				OR (next.entry_exit_time = ? AND next.id > ?)
			)
	`, instID, classID, studentID, currentTime, currentTime, recordID).Scan(&nextTime); err != nil {
		return err
	}
	if nextTime.Valid && !targetTime.Before(nextTime.Time) {
		return errors.New("日期调整范围只能在上一条记录与下一条记录之间")
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE teaching_class_entry_exit_record
		SET entry_exit_time = ?, operator_id = ?, operate_time = NOW(), update_id = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, targetTime, operatorID, operatorID, recordID, instID); err != nil {
		return err
	}

	if entryExitStatus == model.GroupClassEntryExitStatusIn && teachingClassStudentID > 0 {
		if _, err := tx.ExecContext(ctx, `
			UPDATE teaching_class_student
			SET create_time = ?, update_id = ?, update_time = NOW()
			WHERE id = ? AND inst_id = ? AND del_flag = 0
		`, targetTime, operatorID, teachingClassStudentID, instID); err != nil {
			return err
		}
	}

	return tx.Commit()
}
