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
		"2006-01-02T15:04:05",
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
	if trimmed == "" || strings.HasPrefix(trimmed, "0001-01-01") {
		return time.Time{}, false, nil
	}
	parsed, err := parseFlexibleDateTime(trimmed)
	if err != nil {
		return time.Time{}, false, err
	}
	return parsed, true, nil
}

func endOfDayTime(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 23, 59, 59, 0, value.Location())
}

func (repo *Repository) AddSuspendResumeTuitionAccountOrder(ctx context.Context, instID, operatorID int64, dto model.SuspendResumeTuitionAccountOrderDTO) (model.SuspendResumeTuitionAccountOrderResult, error) {
	taID, err := strconv.ParseInt(strings.TrimSpace(dto.TuitionAccountID), 10, 64)
	if err != nil || taID <= 0 {
		return model.SuspendResumeTuitionAccountOrderResult{}, errors.New("tuitionAccountId 无效")
	}
	if dto.Type != 1 && dto.Type != 2 {
		return model.SuspendResumeTuitionAccountOrderResult{}, errors.New("type 无效")
	}

	suspendDate, hasSuspendDate, err := parseOptionalDateTime(dto.SuspendDate)
	if err != nil {
		return model.SuspendResumeTuitionAccountOrderResult{}, err
	}
	resumeDate, hasResumeDate, err := parseOptionalDateTime(dto.ResumeDate)
	if err != nil {
		return model.SuspendResumeTuitionAccountOrderResult{}, err
	}
	if dto.Type == 1 && !hasSuspendDate {
		return model.SuspendResumeTuitionAccountOrderResult{}, errors.New("计划停课日期不能为空")
	}
	if dto.Type == 2 && !hasResumeDate {
		return model.SuspendResumeTuitionAccountOrderResult{}, errors.New("复课日期不能为空")
	}
	if hasResumeDate && hasSuspendDate && resumeDate.Before(suspendDate) {
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
	var accountIDs []int64
	if dto.Type == 2 {
		accountIDs, err = repo.ListTuitionAccountIDsForStudentCourseBucketAllStatuses(ctx, tx, instID, selected.studentID, selected.courseID, bucket.teachMethod, bucket.lessonModelCode)
	} else {
		accountIDs, err = repo.ListTuitionAccountIDsForStudentCourseBucket(ctx, tx, instID, selected.studentID, selected.courseID, bucket.teachMethod, bucket.lessonModelCode)
	}
	if err != nil {
		return model.SuspendResumeTuitionAccountOrderResult{}, err
	}
	if len(accountIDs) == 0 {
		accountIDs = []int64{taID}
	}

	primaryAccountID := accountIDs[0]
	now := time.Now()
	isImmediateSuspend := dto.Type == 1 && !suspendDate.After(now)
	isImmediateResume := dto.Type == 2 && !resumeDate.After(now)

	var suspendDateArg any
	if hasSuspendDate {
		suspendDateArg = suspendDate
	}
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

	args := make([]any, 0, len(accountIDs)+16)
	statusSQL := ""
	updateExpireSQL := ""
	switch dto.Type {
	case 1:
		args = append(args, suspendDateArg, resumeDateArg, operatorID)
		if isImmediateSuspend {
			statusSQL = `,
				status = ?,
				suspended_time = ?,
				status_change_time = ?`
			args = append(args, model.TuitionAccountStatusSuspended, suspendDateArg, now)
		}
	case 2:
		resumePlanArg := resumeDateArg
		if isImmediateResume {
			resumePlanArg = nil
		}
		args = append(args, nil, resumePlanArg, operatorID)
		pausedDays := 0
		if hasResumeDate {
			baseSuspendDate := suspendDate
			if !hasSuspendDate {
				var suspendedTime sql.NullTime
				if err := tx.QueryRowContext(ctx, `
					SELECT MAX(COALESCE(suspended_time, plan_suspend_time, status_change_time))
					FROM tuition_account
					WHERE id IN (`+buildPlaceholders(len(accountIDs))+`)
					  AND inst_id = ?
					  AND del_flag = 0
				`, append(int64SliceToAny(accountIDs), instID)...).Scan(&suspendedTime); err != nil {
					return model.SuspendResumeTuitionAccountOrderResult{}, err
				}
				if suspendedTime.Valid {
					baseSuspendDate = suspendedTime.Time
					hasSuspendDate = true
				}
			}
			if hasSuspendDate {
				pausedDays = int(startOfDayTime(resumeDate).Sub(startOfDayTime(baseSuspendDate)).Hours() / 24)
				if pausedDays < 0 {
					pausedDays = 0
				}
			}
		}
		var expireType1Arg any = nil
		var expireType1Enabled int = 1
		if dto.ExpireType == 1 {
			if !hasExpireTime {
				return model.SuspendResumeTuitionAccountOrderResult{}, errors.New("请选择现有效期至")
			}
			expireType1Arg = endOfDayTime(expireTime)
			updateExpireSQL = `,
				enable_expire_time = ?,
				expire_time = ?`
			args = append(args, expireType1Enabled, expireType1Arg)
		} else if dto.ExpireType == 3 {
			updateExpireSQL = `,
				enable_expire_time = ?,
				expire_time = NULL,
				valid_date = NULL,
				end_date = NULL`
			args = append(args, 0)
		} else if dto.ExpireType == 2 {
			updateExpireSQL = `,
				expire_time = CASE WHEN expire_time IS NULL THEN NULL ELSE DATE_ADD(expire_time, INTERVAL ? DAY) END,
				valid_date = CASE WHEN valid_date IS NULL THEN NULL ELSE DATE_ADD(valid_date, INTERVAL ? DAY) END,
				end_date = CASE WHEN end_date IS NULL THEN NULL ELSE DATE_ADD(end_date, INTERVAL ? DAY) END`
			args = append(args, pausedDays, pausedDays, pausedDays)
		}
		if isImmediateResume {
			statusSQL = `,
				status = ?,
				suspended_time = NULL,
				status_change_time = ?,
				plan_suspend_time = NULL,
				plan_resume_time = NULL`
			args = append(args, model.TuitionAccountStatusActive, now)
		}
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
		    update_time = NOW()`+updateExpireSQL+statusSQL+`
		WHERE id IN (`+buildPlaceholders(len(accountIDs))+`)
		  AND inst_id = ?
		  AND del_flag = 0
	`, args...); err != nil {
		return model.SuspendResumeTuitionAccountOrderResult{}, err
	}

	if dto.Type == 1 && isImmediateSuspend {
		if err := syncOneToOneClassStudentStatusForTuitionAccountsTx(ctx, tx, instID, operatorID, selected.studentID, selected.courseID, accountIDs, model.TeachingClassStudentStatusStopped); err != nil {
			return model.SuspendResumeTuitionAccountOrderResult{}, err
		}
	}
	if dto.Type == 2 && isImmediateResume {
		if err := syncOneToOneClassStudentStatusForTuitionAccountsTx(ctx, tx, instID, operatorID, selected.studentID, selected.courseID, accountIDs, model.TeachingClassStudentStatusStudying); err != nil {
			return model.SuspendResumeTuitionAccountOrderResult{}, err
		}
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

// syncOneToOneClassStudentStatusForTuitionAccountsTx 学费账户停课/复课生效时，同步对应 1 对 1 班员的「开课状态」，与 tuition_account.status 一致。
func syncOneToOneClassStudentStatusForTuitionAccountsTx(ctx context.Context, tx *sql.Tx, instID, operatorID, studentID, courseID int64, accountIDs []int64, targetStatus int) error {
	if len(accountIDs) == 0 {
		return nil
	}
	ph := buildPlaceholders(len(accountIDs))
	args := []any{targetStatus, operatorID, instID, studentID, courseID, model.TeachingClassTypeOneToOne}
	args = append(args, int64SliceToAny(accountIDs)...)
	extraWhere := ""
	if targetStatus == model.TeachingClassStudentStatusStudying {
		extraWhere = ` AND tcs.class_student_status = ? `
		args = append(args, model.TeachingClassStudentStatusStopped)
	} else if targetStatus == model.TeachingClassStudentStatusStopped {
		extraWhere = ` AND tcs.class_student_status <> ? `
		args = append(args, model.TeachingClassStudentStatusClosed)
	}
	_, err := tx.ExecContext(ctx, `
		UPDATE teaching_class_student tcs
		INNER JOIN teaching_class tc ON tc.id = tcs.teaching_class_id AND tc.inst_id = tcs.inst_id AND tc.del_flag = 0
		SET tcs.class_student_status = ?,
		    tcs.update_id = ?,
		    tcs.update_time = NOW()
		WHERE tcs.inst_id = ?
		  AND tcs.del_flag = 0
		  AND tcs.student_id = ?
		  AND tc.course_id = ?
		  AND tc.class_type = ?
		  AND tcs.primary_tuition_account_id IN (`+ph+`)
	`+extraWhere, args...)
	return err
}
