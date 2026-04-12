package repository

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

const (
	faceAttendanceRollCallTaskStatusPending   = 1
	faceAttendanceRollCallTaskStatusSuccess   = 2
	faceAttendanceRollCallTaskStatusSkipped   = 3
	faceAttendanceRollCallTaskBatchSize       = 100
	faceAttendanceRollCallTaskSessionBatchMax = 20
	faceAttendanceAutoRollCallDelay           = 30 * time.Minute
)

type faceAttendanceAutoRollCallTask struct {
	ID                  int64
	InstID              int64
	AttendanceSessionID int64
	StudentID           int64
	AttendanceDate      string
	TeachingScheduleID  int64
	ExecuteAt           time.Time
	SignInTime          sql.NullTime
	Status              int
	TeachingRecordID    int64
	LastError           string
}

type faceAttendanceSessionSnapshot struct {
	ID             int64
	InstID         int64
	StudentID      int64
	AttendanceDate string
	SignInTime     sql.NullTime
	SignOutTime    sql.NullTime
	Status         int
}

type faceAttendanceAutoRollCallScheduleCandidate struct {
	ScheduleID int64
	ClassType  int
	ClassID    int64
	StudentID  int64
	StartAt    time.Time
	EndAt      time.Time
}

func (repo *Repository) SyncFaceAttendanceAutoRollCallTasks(ctx context.Context, operatorID int64, now time.Time) (int, int, error) {
	created, err := repo.ensureFaceAttendanceAutoRollCallTasksForActiveSessions(ctx, operatorID, now)
	if err != nil {
		return 0, 0, err
	}
	processed, err := repo.ProcessFaceAttendanceAutoRollCallTasks(ctx, operatorID, now, nil, nil)
	if err != nil {
		return created, processed, err
	}
	return created, processed, nil
}

func (repo *Repository) ProcessFaceAttendanceAutoRollCallTasks(ctx context.Context, operatorID int64, now time.Time, instID *int64, sessionID *int64) (int, error) {
	taskIDs, err := repo.listDueFaceAttendanceAutoRollCallTaskIDs(ctx, now, instID, sessionID, faceAttendanceRollCallTaskBatchSize)
	if err != nil {
		return 0, err
	}
	processed := 0
	for _, taskID := range taskIDs {
		done, err := repo.processFaceAttendanceAutoRollCallTaskByID(ctx, operatorID, now, taskID)
		if err != nil {
			return processed, err
		}
		if done {
			processed++
		}
	}
	return processed, nil
}

func (repo *Repository) ensureFaceAttendanceAutoRollCallTasksForActiveSessions(ctx context.Context, operatorID int64, now time.Time) (int, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			id,
			inst_id,
			student_id,
			DATE_FORMAT(attendance_date, '%Y-%m-%d'),
			sign_in_time
		FROM inst_student_face_attendance_session
		WHERE del_flag = 0
		  AND status = ?
		  AND attendance_date = ?
		  AND sign_in_time IS NOT NULL
	`, model.FaceAttendanceSessionStatusSignedIn, now.Format("2006-01-02"))
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	totalCreated := 0
	for rows.Next() {
		var (
			sessionID      int64
			instID         int64
			studentID      int64
			attendanceDate string
			signInTime     time.Time
		)
		if err := rows.Scan(&sessionID, &instID, &studentID, &attendanceDate, &signInTime); err != nil {
			return totalCreated, err
		}
		created, err := repo.ensureFaceAttendanceAutoRollCallTasksForSession(ctx, instID, operatorID, sessionID, studentID, attendanceDate, signInTime)
		if err != nil {
			return totalCreated, err
		}
		totalCreated += created
	}
	if err := rows.Err(); err != nil {
		return totalCreated, err
	}
	return totalCreated, nil
}

func (repo *Repository) ensureFaceAttendanceAutoRollCallTasksForSession(ctx context.Context, instID, operatorID, sessionID, studentID int64, attendanceDate string, signInTime time.Time) (int, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	created, err := repo.enqueueFaceAttendanceAutoRollCallTasksTx(ctx, tx, instID, operatorID, sessionID, studentID, attendanceDate, signInTime)
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return created, nil
}

func (repo *Repository) enqueueFaceAttendanceAutoRollCallTasksTx(ctx context.Context, tx *sql.Tx, instID, operatorID, sessionID, studentID int64, attendanceDate string, signInTime time.Time) (int, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT
			ts.id,
			IFNULL(ts.class_type, 0),
			IFNULL(ts.teaching_class_id, 0),
			IFNULL(ts.student_id, 0),
			ts.lesson_start_at,
			ts.lesson_end_at
		FROM teaching_schedule ts
		WHERE ts.inst_id = ?
		  AND ts.del_flag = 0
		  AND ts.status = ?
		  AND ts.lesson_date = ?
		  AND ts.lesson_end_at >= ?
		  AND ts.class_type IN (?, ?)
	`, instID, model.TeachingScheduleStatusActive, attendanceDate, signInTime, model.TeachingClassTypeNormal, model.TeachingClassTypeOneToOne)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	candidates := make([]faceAttendanceAutoRollCallScheduleCandidate, 0)
	groupMetas := make([]effectiveGroupClassScheduleMeta, 0)
	for rows.Next() {
		var item faceAttendanceAutoRollCallScheduleCandidate
		if err := rows.Scan(
			&item.ScheduleID,
			&item.ClassType,
			&item.ClassID,
			&item.StudentID,
			&item.StartAt,
			&item.EndAt,
		); err != nil {
			return 0, err
		}
		candidates = append(candidates, item)
		if item.ClassType == model.TeachingClassTypeNormal && item.ClassID > 0 {
			groupMetas = append(groupMetas, effectiveGroupClassScheduleMeta{
				ScheduleID: item.ScheduleID,
				ClassID:    item.ClassID,
				StartAt:    item.StartAt,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	rosterByScheduleID := map[int64]groupClassScheduleRoster{}
	if len(groupMetas) > 0 {
		rosterByScheduleID, err = repo.loadEffectiveGroupClassScheduleRosterMap(ctx, tx, instID, groupMetas)
		if err != nil {
			return 0, err
		}
	}

	totalAffected := 0
	for _, candidate := range candidates {
		shouldEnqueue := false
		switch candidate.ClassType {
		case model.TeachingClassTypeOneToOne:
			shouldEnqueue = candidate.StudentID == studentID
		case model.TeachingClassTypeNormal:
			if candidate.ClassID > 0 {
				if _, ok := rosterByScheduleID[candidate.ScheduleID].activeStudentNameMap()[studentID]; ok {
					shouldEnqueue = true
				}
			}
		}
		if !shouldEnqueue {
			continue
		}

		result, err := tx.ExecContext(ctx, `
			INSERT INTO inst_student_face_roll_call_task (
				inst_id, attendance_session_id, student_id, attendance_date, teaching_schedule_id,
				execute_at, sign_in_time, status, teaching_record_id, last_error, create_id, update_id, del_flag
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, '', ?, ?, 0)
			ON DUPLICATE KEY UPDATE
				execute_at = VALUES(execute_at),
				sign_in_time = VALUES(sign_in_time),
				update_id = VALUES(update_id),
				update_time = CURRENT_TIMESTAMP,
				del_flag = 0
		`, instID, sessionID, studentID, attendanceDate, candidate.ScheduleID, candidate.EndAt.Add(faceAttendanceAutoRollCallDelay), signInTime, faceAttendanceRollCallTaskStatusPending, operatorID, operatorID)
		if err != nil {
			return totalAffected, err
		}
		affected, _ := result.RowsAffected()
		totalAffected += int(affected)
	}
	return totalAffected, nil
}

func (repo *Repository) listDueFaceAttendanceAutoRollCallTaskIDs(ctx context.Context, now time.Time, instID *int64, sessionID *int64, limit int) ([]int64, error) {
	if limit <= 0 {
		limit = faceAttendanceRollCallTaskBatchSize
	}
	filters := []string{
		"del_flag = 0",
		"status = ?",
		"execute_at <= ?",
	}
	args := []any{faceAttendanceRollCallTaskStatusPending, now}
	if instID != nil && *instID > 0 {
		filters = append(filters, "inst_id = ?")
		args = append(args, *instID)
	}
	if sessionID != nil && *sessionID > 0 {
		filters = append(filters, "attendance_session_id = ?")
		args = append(args, *sessionID)
	}
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id
		FROM inst_student_face_roll_call_task
		WHERE `+strings.Join(filters, " AND ")+`
		ORDER BY execute_at ASC, id ASC
		LIMIT ?
	`, append(args, limit)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]int64, 0, limit)
	for rows.Next() {
		var taskID int64
		if err := rows.Scan(&taskID); err != nil {
			return nil, err
		}
		result = append(result, taskID)
	}
	return result, rows.Err()
}

func (repo *Repository) processFaceAttendanceAutoRollCallTaskByID(ctx context.Context, operatorID int64, now time.Time, taskID int64) (bool, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	task, found, err := repo.getFaceAttendanceAutoRollCallTaskForUpdate(ctx, tx, taskID)
	if err != nil {
		return false, err
	}
	if !found || task.Status != faceAttendanceRollCallTaskStatusPending || task.ExecuteAt.After(now) {
		return false, nil
	}

	session, found, err := repo.getFaceAttendanceSessionSnapshotForUpdate(ctx, tx, task.InstID, task.AttendanceSessionID)
	if err != nil {
		return false, err
	}
	if !found || !session.SignInTime.Valid {
		if err := repo.updateFaceAttendanceAutoRollCallTaskStatusTx(ctx, tx, task.ID, operatorID, faceAttendanceRollCallTaskStatusSkipped, 0, "缺少首次签到记录"); err != nil {
			return false, err
		}
		return true, tx.Commit()
	}

	if existingTeachingRecordID, err := repo.findExistingTeachingRecordIDByScheduleStudentTx(ctx, tx, task.InstID, task.TeachingScheduleID, task.StudentID); err != nil {
		return false, err
	} else if existingTeachingRecordID > 0 {
		if err := repo.updateFaceAttendanceAutoRollCallTaskStatusTx(ctx, tx, task.ID, operatorID, faceAttendanceRollCallTaskStatusSuccess, existingTeachingRecordID, ""); err != nil {
			return false, err
		}
		return true, tx.Commit()
	}

	teachingRecordID, err := repo.confirmFaceAttendanceAutoRollCallTaskTx(ctx, tx, operatorID, now, task, session)
	if err != nil {
		if updateErr := repo.updateFaceAttendanceAutoRollCallTaskStatusTx(ctx, tx, task.ID, operatorID, faceAttendanceRollCallTaskStatusSkipped, 0, trimTaskError(err)); updateErr != nil {
			return false, updateErr
		}
		return true, tx.Commit()
	}
	if err := repo.updateFaceAttendanceAutoRollCallTaskStatusTx(ctx, tx, task.ID, operatorID, faceAttendanceRollCallTaskStatusSuccess, teachingRecordID, ""); err != nil {
		return false, err
	}
	return true, tx.Commit()
}

func (repo *Repository) getFaceAttendanceAutoRollCallTaskForUpdate(ctx context.Context, tx *sql.Tx, taskID int64) (faceAttendanceAutoRollCallTask, bool, error) {
	var item faceAttendanceAutoRollCallTask
	err := tx.QueryRowContext(ctx, `
		SELECT
			id,
			inst_id,
			attendance_session_id,
			student_id,
			DATE_FORMAT(attendance_date, '%Y-%m-%d'),
			teaching_schedule_id,
			execute_at,
			sign_in_time,
			status,
			teaching_record_id,
			IFNULL(last_error, '')
		FROM inst_student_face_roll_call_task
		WHERE id = ? AND del_flag = 0
		FOR UPDATE
	`, taskID).Scan(
		&item.ID,
		&item.InstID,
		&item.AttendanceSessionID,
		&item.StudentID,
		&item.AttendanceDate,
		&item.TeachingScheduleID,
		&item.ExecuteAt,
		&item.SignInTime,
		&item.Status,
		&item.TeachingRecordID,
		&item.LastError,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return faceAttendanceAutoRollCallTask{}, false, nil
		}
		return faceAttendanceAutoRollCallTask{}, false, err
	}
	return item, true, nil
}

func (repo *Repository) getFaceAttendanceSessionSnapshotForUpdate(ctx context.Context, tx *sql.Tx, instID, sessionID int64) (faceAttendanceSessionSnapshot, bool, error) {
	var item faceAttendanceSessionSnapshot
	err := tx.QueryRowContext(ctx, `
		SELECT
			id,
			inst_id,
			student_id,
			DATE_FORMAT(attendance_date, '%Y-%m-%d'),
			sign_in_time,
			sign_out_time,
			status
		FROM inst_student_face_attendance_session
		WHERE inst_id = ? AND id = ? AND del_flag = 0
		FOR UPDATE
	`, instID, sessionID).Scan(
		&item.ID,
		&item.InstID,
		&item.StudentID,
		&item.AttendanceDate,
		&item.SignInTime,
		&item.SignOutTime,
		&item.Status,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return faceAttendanceSessionSnapshot{}, false, nil
		}
		return faceAttendanceSessionSnapshot{}, false, err
	}
	return item, true, nil
}

func (repo *Repository) findExistingTeachingRecordIDByScheduleStudentTx(ctx context.Context, tx *sql.Tx, instID, scheduleID, studentID int64) (int64, error) {
	var teachingRecordID int64
	err := tx.QueryRowContext(ctx, `
		SELECT IFNULL(teaching_record_id, 0)
		FROM student_teaching_record
		WHERE inst_id = ?
		  AND teaching_schedule_id = ?
		  AND student_id = ?
		  AND del_flag = 0
		ORDER BY id ASC
		LIMIT 1
	`, instID, scheduleID, studentID).Scan(&teachingRecordID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return teachingRecordID, nil
}

func (repo *Repository) confirmFaceAttendanceAutoRollCallTaskTx(ctx context.Context, tx *sql.Tx, operatorID int64, now time.Time, task faceAttendanceAutoRollCallTask, session faceAttendanceSessionSnapshot) (int64, error) {
	scheduleIDText := strconv.FormatInt(task.TeachingScheduleID, 10)
	rollCallData, err := repo.GetRollCallTeachingRecordStudentList(ctx, task.InstID, model.RollCallTeachingRecordStudentListQueryDTO{
		TimetableSourceID: scheduleIDText,
	})
	if err != nil {
		return 0, err
	}
	detail, classMeta, err := repo.loadRollCallDrawerContext(ctx, task.InstID, scheduleIDText)
	if err != nil {
		return 0, err
	}

	var target *model.RollCallTeachingRecordStudentVO
	for index := range rollCallData.Students {
		item := &rollCallData.Students[index]
		currentStudentID, parseErr := strconv.ParseInt(strings.TrimSpace(item.StudentID), 10, 64)
		if parseErr == nil && currentStudentID == task.StudentID {
			target = item
			break
		}
	}
	if target == nil {
		return 0, errors.New("当前学员不在本节日程中")
	}

	quantity := roundMoney(classMeta.DefaultStudentClassTime)
	if quantity <= 0 {
		quantity = 1
	}
	teacherList := make([]model.RollCallConfirmTeacher, 0, len(rollCallData.Teachers))
	for _, teacher := range rollCallData.Teachers {
		teacherList = append(teacherList, model.RollCallConfirmTeacher{
			TeacherID: teacher.TeacherID,
			Type:      teacher.Type,
		})
	}

	recordTime := task.ExecuteAt
	if session.SignInTime.Valid && session.SignInTime.Time.After(recordTime) {
		recordTime = session.SignInTime.Time
	}
	if recordTime.After(now) {
		recordTime = now
	}

	result, err := repo.confirmRollCallTx(ctx, tx, task.InstID, operatorID, model.RollCallConfirmDTO{
		SourceName:          rollCallData.Data.SourceName,
		TimetableSourceType: rollCallData.Data.TimetableSourceType,
		TimetableSourceID:   rollCallData.Data.TimetableSourceID,
		SourceID:            rollCallData.Data.SourceID,
		SourceType:          rollCallData.Data.SourceType,
		LessonID:            rollCallData.Data.LessonID,
		StartTime:           rollCallData.Data.StartTime,
		EndTime:             rollCallData.Data.EndTime,
		TeacherClassTime:    rollCallData.Data.TeacherClassTime,
		StudentShouldDeduct: int(math.Round(quantity)),
		StudentList: []model.RollCallConfirmStudent{{
			StudentShouldDeduct:    int(math.Round(quantity)),
			StudentName:            target.StudentName,
			StudentID:              target.StudentID,
			TuitionAccountID:       target.TuitionAccountID,
			AbsentTeachingRecordID: "0",
			Status:                 1,
			SourceType:             target.SourceType,
			Remark:                 "",
			ExternalRemark:         "",
			SkuMode:                target.ChargingMode,
			Amount:                 0,
			Quantity:               quantity,
		}},
		TeacherList: teacherList,
		SubjectID:   "0",
		ClassRoomID: firstNonEmptyString(rollCallData.Data.ClassroomID, detail.ClassroomID, "0"),
	}, rollCallConfirmOptions{
		RecordTime:     &recordTime,
		OperatorName:   "系统自动点名",
		IsAutoRollCall: true,
	})
	if err != nil {
		return 0, err
	}
	teachingRecordID, _ := strconv.ParseInt(strings.TrimSpace(result.ID), 10, 64)
	return teachingRecordID, nil
}

func (repo *Repository) updateFaceAttendanceAutoRollCallTaskStatusTx(ctx context.Context, tx *sql.Tx, taskID, operatorID int64, status int, teachingRecordID int64, lastError string) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE inst_student_face_roll_call_task
		SET status = ?, teaching_record_id = ?, last_error = ?, update_id = ?, update_time = CURRENT_TIMESTAMP
		WHERE id = ? AND del_flag = 0
	`, status, teachingRecordID, strings.TrimSpace(lastError), operatorID, taskID)
	return err
}

func trimTaskError(err error) string {
	if err == nil {
		return ""
	}
	text := strings.TrimSpace(err.Error())
	if len(text) > 500 {
		return text[:500]
	}
	return text
}
