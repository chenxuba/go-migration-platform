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

func ensureSuspendResumeTuitionAccountOrderTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS suspend_resume_tuition_account_order (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			uuid VARCHAR(64) NULL,
			version BIGINT NOT NULL DEFAULT 0,
			inst_id BIGINT NOT NULL,
			tuition_account_id BIGINT NOT NULL,
			student_id BIGINT NOT NULL,
			course_id BIGINT NOT NULL,
			type INT NOT NULL DEFAULT 0,
			expire_time DATETIME NULL DEFAULT NULL,
			expire_type INT NOT NULL DEFAULT 0,
			remark VARCHAR(500) NOT NULL DEFAULT '',
			suspend_date DATETIME NULL DEFAULT NULL,
			resume_date DATETIME NULL DEFAULT NULL,
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT NOT NULL DEFAULT 0,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			KEY idx_suspend_resume_tuition_account_order_inst (inst_id, tuition_account_id, create_time)
		)
	`)
	return err
}

func parseFlexibleDateTime(value string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, errors.New("日期不能为空")
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, trimmed, time.Local); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, errors.New("日期格式不正确")
}

func parseOptionalDateTime(value string) (time.Time, bool, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, false, nil
	}
	parsed, err := parseFlexibleDateTime(trimmed)
	if err != nil {
		return time.Time{}, false, err
	}
	return parsed, true, nil
}

func (repo *Repository) AddSuspendResumeTuitionAccountOrder(ctx context.Context, instID, operatorID int64, dto model.SuspendResumeTuitionAccountOrderDTO) (model.SuspendResumeTuitionAccountOrderResult, error) {
	taID, err := strconv.ParseInt(strings.TrimSpace(dto.TuitionAccountID), 10, 64)
	if err != nil || taID <= 0 {
		return model.SuspendResumeTuitionAccountOrderResult{}, errors.New("tuitionAccountId 无效")
	}
	if dto.Type != 1 {
		return model.SuspendResumeTuitionAccountOrderResult{}, errors.New("暂只支持停课")
	}

	suspendDate, err := parseFlexibleDateTime(dto.SuspendDate)
	if err != nil {
		return model.SuspendResumeTuitionAccountOrderResult{}, err
	}
	resumeDate, hasResumeDate, err := parseOptionalDateTime(dto.ResumeDate)
	if err != nil {
		return model.SuspendResumeTuitionAccountOrderResult{}, err
	}
	if hasResumeDate && resumeDate.Before(suspendDate) {
		return model.SuspendResumeTuitionAccountOrderResult{}, errors.New("计划复课日期不能早于计划停课日期")
	}
	expireTime, hasExpireTime, err := parseOptionalDateTime(dto.ExpireTime)
	if err != nil {
		return model.SuspendResumeTuitionAccountOrderResult{}, err
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return model.SuspendResumeTuitionAccountOrderResult{}, err
	}
	defer tx.Rollback()

	selected, err := repo.loadCloseTuitionAccountSnapshotTx(ctx, tx, instID, taID)
	if err != nil {
		return model.SuspendResumeTuitionAccountOrderResult{}, err
	}
	bucket, err := repo.loadOneToOneTuitionBucketTx(ctx, tx, instID, taID)
	if err != nil {
		return model.SuspendResumeTuitionAccountOrderResult{}, err
	}
	accountIDs, err := repo.ListTuitionAccountIDsForStudentCourseBucket(ctx, tx, instID, selected.studentID, selected.courseID, bucket.teachMethod, bucket.lessonModelCode)
	if err != nil {
		return model.SuspendResumeTuitionAccountOrderResult{}, err
	}
	if len(accountIDs) == 0 {
		accountIDs = []int64{taID}
	}

	primaryAccountID := accountIDs[0]
	now := time.Now()
	isImmediateSuspend := !suspendDate.After(now)

	var suspendDateArg any = suspendDate
	var resumeDateArg any
	if hasResumeDate {
		resumeDateArg = resumeDate
	}
	var expireTimeArg any
	if hasExpireTime {
		expireTimeArg = expireTime
	}

	res, err := tx.ExecContext(ctx, `
		INSERT INTO suspend_resume_tuition_account_order (
			uuid, version, inst_id, tuition_account_id, student_id, course_id, type,
			expire_time, expire_type, remark, suspend_date, resume_date,
			create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, NOW(), ?, NOW(), 0
		)
	`, instID, primaryAccountID, selected.studentID, selected.courseID, dto.Type, expireTimeArg, dto.ExpireType, strings.TrimSpace(dto.Remark), suspendDateArg, resumeDateArg, operatorID, operatorID)
	if err != nil {
		return model.SuspendResumeTuitionAccountOrderResult{}, err
	}
	orderID, err := res.LastInsertId()
	if err != nil {
		return model.SuspendResumeTuitionAccountOrderResult{}, err
	}

	args := make([]any, 0, len(accountIDs)+7)
	args = append(args, suspendDateArg, resumeDateArg, operatorID)
	statusSQL := ""
	if isImmediateSuspend {
		statusSQL = `,
			status = ?,
			suspended_time = ?,
			status_change_time = ?`
		args = append(args, model.TuitionAccountStatusSuspended, suspendDateArg, now)
	}
	for _, id := range accountIDs {
		args = append(args, id)
	}
	args = append(args, instID)

	if _, err := tx.ExecContext(ctx, `
		UPDATE tuition_account
		SET plan_suspend_time = ?,
		    plan_resume_time = ?,
		    update_id = ?,
		    update_time = NOW()`+statusSQL+`
		WHERE id IN (`+buildPlaceholders(len(accountIDs))+`)
		  AND inst_id = ?
		  AND del_flag = 0
	`, args...); err != nil {
		return model.SuspendResumeTuitionAccountOrderResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return model.SuspendResumeTuitionAccountOrderResult{}, err
	}

	return model.SuspendResumeTuitionAccountOrderResult{
		ID:        strconv.FormatInt(orderID, 10),
		StudentID: strconv.FormatInt(selected.studentID, 10),
		LessonID:  strconv.FormatInt(selected.courseID, 10),
	}, nil
}
