package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
			KEY idx_teaching_class_operation_log_member_type (inst_id, teaching_class_student_id, operation_type, operate_time),
			KEY idx_teaching_class_operation_log_class (inst_id, teaching_class_id, operate_time),
			KEY idx_teaching_class_operation_log_student (inst_id, student_id, operate_time)
		)
	`); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `
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
			KEY idx_teaching_class_entry_exit_member_status (inst_id, teaching_class_student_id, entry_exit_status, entry_exit_time),
			KEY idx_teaching_class_entry_exit_class (inst_id, teaching_class_id, entry_exit_time),
			KEY idx_teaching_class_entry_exit_student (inst_id, student_id, entry_exit_time)
		)
	`); err != nil {
		return err
	}
	return ensureGroupClassHistoryAppendOnlyMigration(ctx, db)
}

func ensureGroupClassHistoryAppendOnlyMigration(ctx context.Context, db *sql.DB) error {
	if err := dropTableIndexIfExists(ctx, db, "teaching_class_operation_log", "uk_teaching_class_operation_log_unique"); err != nil {
		return err
	}
	if err := ensureTableIndexExists(ctx, db, "teaching_class_operation_log", "idx_teaching_class_operation_log_member_type",
		`ALTER TABLE teaching_class_operation_log ADD KEY idx_teaching_class_operation_log_member_type (inst_id, teaching_class_student_id, operation_type, operate_time)`); err != nil {
		return err
	}
	if err := dropTableIndexIfExists(ctx, db, "teaching_class_entry_exit_record", "uk_teaching_class_entry_exit_unique"); err != nil {
		return err
	}
	return ensureTableIndexExists(ctx, db, "teaching_class_entry_exit_record", "idx_teaching_class_entry_exit_member_status",
		`ALTER TABLE teaching_class_entry_exit_record ADD KEY idx_teaching_class_entry_exit_member_status (inst_id, teaching_class_student_id, entry_exit_status, entry_exit_time)`)
}

func dropTableIndexIfExists(ctx context.Context, db *sql.DB, tableName, indexName string) error {
	var count int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM information_schema.statistics
		WHERE table_schema = DATABASE() AND table_name = ? AND index_name = ?
	`, tableName, indexName).Scan(&count); err != nil {
		return err
	}
	if count <= 0 {
		return nil
	}
	_, err := db.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s DROP INDEX %s", tableName, indexName))
	return err
}

func ensureTableIndexExists(ctx context.Context, db *sql.DB, tableName, indexName, ddl string) error {
	var count int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM information_schema.statistics
		WHERE table_schema = DATABASE() AND table_name = ? AND index_name = ?
	`, tableName, indexName).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	_, err := db.ExecContext(ctx, ddl)
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

func groupClassStudentDisplayName(studentName string, studentID int64) string {
	if value := strings.TrimSpace(studentName); value != "" {
		return value
	}
	return fmt.Sprintf("学员%d", studentID)
}

func groupClassTuitionAccountDisplayName(courseName string, tuitionAccountID int64) string {
	if value := strings.TrimSpace(courseName); value != "" {
		return value
	}
	return fmt.Sprintf("账户%d", tuitionAccountID)
}

func groupClassAddOperationContent(studentName string, studentID, tuitionAccountID int64, courseName string) string {
	return fmt.Sprintf(
		"%s入班，学费账户为：%s",
		groupClassStudentDisplayName(studentName, studentID),
		groupClassTuitionAccountDisplayName(courseName, tuitionAccountID),
	)
}

func groupClassStatusChangeOperationContent(studentName string, studentID int64, targetStatus int) string {
	return fmt.Sprintf(
		"%s在班状态变更为：%s",
		groupClassStudentDisplayName(studentName, studentID),
		teachingClassStudentStatusTextForLog(targetStatus),
	)
}

func (repo *Repository) appendGroupClassOperationLogTx(
	ctx context.Context,
	tx *sql.Tx,
	instID, operatorID, classID, teachingClassStudentID, studentID int64,
	operationType int,
	operationContent string,
	operateTime time.Time,
) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO teaching_class_operation_log (
			uuid, version, inst_id, teaching_class_id, teaching_class_student_id, student_id,
			operation_type, operation_content, operator_id, operate_time,
			create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?, 0
		)
	`, instID, classID, teachingClassStudentID, studentID,
		operationType, operationContent, operatorID, operateTime,
		operatorID, operateTime, operatorID, operateTime)
	return err
}

func (repo *Repository) appendGroupClassEntryExitRecordTx(
	ctx context.Context,
	tx *sql.Tx,
	instID, operatorID, classID, teachingClassStudentID, studentID int64,
	entryExitStatus int,
	entryExitTime time.Time,
) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO teaching_class_entry_exit_record (
			uuid, version, inst_id, teaching_class_id, teaching_class_student_id, student_id,
			entry_exit_status, entry_exit_time, operator_id, operate_time,
			create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?, 0
		)
	`, instID, classID, teachingClassStudentID, studentID,
		entryExitStatus, entryExitTime, operatorID, entryExitTime,
		operatorID, entryExitTime, operatorID, entryExitTime)
	return err
}

func (repo *Repository) syncGroupClassStudentAddHistoryTx(
	ctx context.Context,
	tx *sql.Tx,
	instID, operatorID, classID, teachingClassStudentID int64,
	operateTime time.Time,
) error {
	var (
		studentID        int64
		tuitionAccountID int64
		studentName      string
		courseName       string
	)
	err := tx.QueryRowContext(ctx, `
		SELECT
			tcs.student_id,
			IFNULL(tcs.primary_tuition_account_id, 0),
			IFNULL(s.stu_name, ''),
			IFNULL(ic.name, '')
		FROM teaching_class_student tcs
		LEFT JOIN inst_student s ON s.id = tcs.student_id AND s.inst_id = tcs.inst_id AND s.del_flag = 0
		LEFT JOIN tuition_account ta ON ta.id = tcs.primary_tuition_account_id AND ta.inst_id = tcs.inst_id AND ta.del_flag = 0
		LEFT JOIN inst_course ic ON ic.id = ta.course_id AND ic.inst_id = ta.inst_id AND ic.del_flag = 0
		WHERE tcs.inst_id = ?
		  AND tcs.teaching_class_id = ?
		  AND tcs.id = ?
		  AND tcs.del_flag = 0
		LIMIT 1
	`, instID, classID, teachingClassStudentID).Scan(
		&studentID,
		&tuitionAccountID,
		&studentName,
		&courseName,
	)
	if err != nil {
		return err
	}

	operationContent := groupClassAddOperationContent(studentName, studentID, tuitionAccountID, courseName)

	if err := repo.appendGroupClassOperationLogTx(
		ctx,
		tx,
		instID,
		operatorID,
		classID,
		teachingClassStudentID,
		studentID,
		model.GroupClassOperationTypeAddStudent,
		operationContent,
		operateTime,
	); err != nil {
		return err
	}
	return repo.appendGroupClassEntryExitRecordTx(
		ctx,
		tx,
		instID,
		operatorID,
		classID,
		teachingClassStudentID,
		studentID,
		model.GroupClassEntryExitStatusIn,
		operateTime,
	)
}

func (repo *Repository) syncGroupClassStudentStatusChangeHistoryTx(
	ctx context.Context,
	tx *sql.Tx,
	instID, operatorID, classID, teachingClassStudentID int64,
	targetStatus int,
	operateTime time.Time,
) error {
	var (
		studentID   int64
		studentName string
	)
	err := tx.QueryRowContext(ctx, `
		SELECT
			tcs.student_id,
			IFNULL(s.stu_name, '')
		FROM teaching_class_student tcs
		LEFT JOIN inst_student s ON s.id = tcs.student_id AND s.inst_id = tcs.inst_id AND s.del_flag = 0
		WHERE tcs.inst_id = ?
		  AND tcs.teaching_class_id = ?
		  AND tcs.id = ?
		  AND tcs.del_flag = 0
		LIMIT 1
	`, instID, classID, teachingClassStudentID).Scan(&studentID, &studentName)
	if err != nil {
		return err
	}
	return repo.appendGroupClassOperationLogTx(
		ctx,
		tx,
		instID,
		operatorID,
		classID,
		teachingClassStudentID,
		studentID,
		model.GroupClassOperationTypeStatusChange,
		groupClassStatusChangeOperationContent(studentName, studentID, targetStatus),
		operateTime,
	)
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
		UPDATE teaching_class_operation_log log
		INNER JOIN teaching_class_student tcs
			ON tcs.id = log.teaching_class_student_id
			AND tcs.inst_id = log.inst_id
			AND tcs.del_flag = 0
		LEFT JOIN inst_student s ON s.id = tcs.student_id AND s.inst_id = tcs.inst_id AND s.del_flag = 0
		LEFT JOIN tuition_account ta ON ta.id = tcs.primary_tuition_account_id AND ta.inst_id = tcs.inst_id AND ta.del_flag = 0
		LEFT JOIN inst_course ic ON ic.id = ta.course_id AND ic.inst_id = ta.inst_id AND ic.del_flag = 0
		SET log.operation_content = CONCAT(
			COALESCE(NULLIF(s.stu_name, ''), CONCAT('学员', CAST(tcs.student_id AS CHAR))),
			'入班，学费账户为：',
			COALESCE(NULLIF(ic.name, ''), CONCAT('账户', CAST(IFNULL(ta.id, 0) AS CHAR)))
		),
		    log.update_time = NOW()
		WHERE log.inst_id = ?
		  AND log.teaching_class_id = ?
		  AND log.operation_type = ?
		  AND log.del_flag = 0
	`, instID, classID, model.GroupClassOperationTypeAddStudent); err != nil {
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
			?, tcs.create_time, IFNULL(tcs.create_id, 0), tcs.create_time,
			IFNULL(tcs.create_id, 0), tcs.create_time, IFNULL(tcs.create_id, 0), tcs.create_time, 0
		FROM teaching_class_student tcs
		LEFT JOIN teaching_class_entry_exit_record rec
			ON rec.inst_id = tcs.inst_id
			AND rec.teaching_class_student_id = tcs.id
			AND rec.entry_exit_status = ?
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
				COALESCE(NULLIF(ic.name, ''), CONCAT('账户', CAST(IFNULL(ta.id, 0) AS CHAR)))
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
		WHERE tcs.inst_id = ? AND tcs.teaching_class_id = ? AND tcs.del_flag = 0
			AND log.id IS NULL
	`,
		model.GroupClassOperationTypeAddStudent,
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
			?,
			CASE
				WHEN IFNULL(tcs.class_student_status, 1) = ? THEN CONCAT(
					COALESCE(NULLIF(s.stu_name, ''), CONCAT('学员', CAST(tcs.student_id AS CHAR))),
					'在班状态变更为：结课'
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
		LEFT JOIN teaching_class_operation_log statusLog
			ON statusLog.inst_id = tcs.inst_id
			AND statusLog.teaching_class_student_id = tcs.id
			AND statusLog.operation_type = ?
		LEFT JOIN teaching_class_operation_log removeLog
			ON removeLog.inst_id = tcs.inst_id
			AND removeLog.teaching_class_student_id = tcs.id
			AND removeLog.operation_type = ?
		WHERE tcs.inst_id = ? AND tcs.teaching_class_id = ? AND tcs.del_flag = 0
			AND IFNULL(tcs.class_student_status, 1) IN (?, ?)
			AND statusLog.id IS NULL
			AND removeLog.id IS NULL
	`,
		model.GroupClassOperationTypeStatusChange,
		model.TeachingClassStudentStatusClosed,
		model.GroupClassOperationTypeStatusChange,
		model.GroupClassOperationTypeRemoveStudent,
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
			rec.operate_time
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
			item   model.GroupClassEntryExitRecordItemVO
			avatar string
			mobile string
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
		); err != nil {
			return out, err
		}
		item.Avatar = strings.TrimSpace(avatar)
		item.Phone = maskPhoneDisplay(mobile)
		item.EntryExitStatusText = groupClassEntryExitStatusText(item.EntryExitStatus)
		out.List = append(out.List, item)
	}
	if err := rows.Err(); err != nil {
		return out, err
	}
	return out, nil
}
