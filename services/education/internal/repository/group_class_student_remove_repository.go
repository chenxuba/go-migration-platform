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

func (repo *Repository) RemoveGroupClassStudent(ctx context.Context, instID, operatorID int64, dto model.GroupClassRemoveStudentDTO) error {
	classID, err := strconv.ParseInt(strings.TrimSpace(dto.ClassID), 10, 64)
	if err != nil || classID <= 0 {
		return errors.New("classId 无效")
	}
	studentID, err := strconv.ParseInt(strings.TrimSpace(dto.StudentID), 10, 64)
	if err != nil || studentID <= 0 {
		return errors.New("studentId 无效")
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var classExists int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM teaching_class
		WHERE inst_id = ?
		  AND id = ?
		  AND class_type = ?
		  AND del_flag = 0
	`, instID, classID, model.TeachingClassTypeNormal).Scan(&classExists); err != nil {
		return err
	}
	if classExists <= 0 {
		return errors.New("班级不存在")
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT
			id,
			IFNULL(class_student_status, ?)
		FROM teaching_class_student
		WHERE inst_id = ?
		  AND teaching_class_id = ?
		  AND student_id = ?
		  AND del_flag = 0
		ORDER BY id ASC
	`, model.TeachingClassStudentStatusStudying, instID, classID, studentID)
	if err != nil {
		return err
	}
	defer rows.Close()

	targetMembershipIDs := make([]int64, 0, 2)
	for rows.Next() {
		var membershipID int64
		var status int
		if err := rows.Scan(&membershipID, &status); err != nil {
			return err
		}
		if membershipID <= 0 || status == model.TeachingClassStudentStatusClosed {
			continue
		}
		targetMembershipIDs = append(targetMembershipIDs, membershipID)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(targetMembershipIDs) == 0 {
		return errors.New("学员已不在当前班级中")
	}

	studentName := fmt.Sprintf("学员%d", studentID)
	if err := tx.QueryRowContext(ctx, `
		SELECT IFNULL(stu_name, '')
		FROM inst_student
		WHERE inst_id = ?
		  AND id = ?
		  AND del_flag = 0
		LIMIT 1
	`, instID, studentID).Scan(&studentName); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if strings.TrimSpace(studentName) == "" {
		studentName = fmt.Sprintf("学员%d", studentID)
	}

	scheduleRows, err := tx.QueryContext(ctx, `
		SELECT
			id,
			lesson_start_at
		FROM teaching_schedule
		WHERE inst_id = ?
		  AND teaching_class_id = ?
		  AND class_type = ?
		  AND status = ?
		  AND del_flag = 0
		ORDER BY id ASC
	`, instID, classID, model.TeachingClassTypeNormal, model.TeachingScheduleStatusActive)
	if err != nil {
		return err
	}
	defer scheduleRows.Close()

	metas := make([]teachingScheduleRollCallMeta, 0, 16)
	for scheduleRows.Next() {
		var scheduleID int64
		var startAt time.Time
		if err := scheduleRows.Scan(&scheduleID, &startAt); err != nil {
			return err
		}
		if scheduleID <= 0 {
			continue
		}
		metas = append(metas, teachingScheduleRollCallMeta{
			ScheduleID: scheduleID,
			ClassType:  model.TeachingClassTypeNormal,
			ClassID:    classID,
			StartAt:    startAt,
		})
	}
	if err := scheduleRows.Err(); err != nil {
		return err
	}

	if len(metas) > 0 {
		statusByID, err := repo.computeTeachingScheduleCallStatusMap(ctx, tx, instID, metas)
		if err != nil {
			return err
		}
		for _, meta := range metas {
			if normalizeTeachingScheduleCallStatus(statusByID[meta.ScheduleID]) != 1 {
				continue
			}
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO teaching_schedule_student (
					inst_id, teaching_schedule_id, teaching_class_id, student_id,
					student_type, roster_status, create_id, update_id, del_flag
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0)
				ON DUPLICATE KEY UPDATE
					teaching_class_id = VALUES(teaching_class_id),
					student_type = VALUES(student_type),
					roster_status = VALUES(roster_status),
					update_id = VALUES(update_id),
					update_time = CURRENT_TIMESTAMP,
					del_flag = 0
			`, instID, meta.ScheduleID, classID, studentID, model.TeachingScheduleStudentTypeClassMember, model.TeachingScheduleStudentRosterStatusRemoved, operatorID, operatorID); err != nil {
				return err
			}
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE teaching_class_student
		SET class_student_status = ?,
		    update_id = ?,
		    update_time = NOW()
		WHERE inst_id = ?
		  AND id IN (`+sqlPlaceholders(len(targetMembershipIDs))+`)
		  AND del_flag = 0
	`, append([]any{
		model.TeachingClassStudentStatusClosed,
		operatorID,
		instID,
	}, int64SliceToAny(targetMembershipIDs)...)...); err != nil {
		return err
	}

	operateTime := time.Now()
	operationContent := strings.TrimSpace(studentName) + "移出班级，在班状态为：转出"
	for _, membershipID := range targetMembershipIDs {
		if err := repo.appendGroupClassEntryExitRecordTx(
			ctx,
			tx,
			instID,
			operatorID,
			classID,
			membershipID,
			studentID,
			model.GroupClassEntryExitStatusOut,
			operateTime,
		); err != nil {
			return err
		}
		if err := repo.appendGroupClassOperationLogTx(
			ctx,
			tx,
			instID,
			operatorID,
			classID,
			membershipID,
			studentID,
			model.GroupClassOperationTypeRemoveStudent,
			operationContent,
			operateTime,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}
