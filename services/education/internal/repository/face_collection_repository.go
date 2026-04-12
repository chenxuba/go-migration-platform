package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

const (
	faceAttendanceDuplicateWindow = 5 * time.Minute
	faceAttendanceSignOutGrace    = 1 * time.Minute
)

type faceAttendanceQueryRunner interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type faceAttendanceStudentBase struct {
	StudentID   int64
	StudentName string
	AvatarURL   string
}

func ensureFaceCollectionTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS inst_student_face_profile (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			inst_id BIGINT NOT NULL,
			student_id BIGINT NOT NULL,
			face_descriptor LONGTEXT NOT NULL,
			face_image LONGTEXT NULL,
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT NOT NULL DEFAULT 0,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uk_inst_student_face_profile (inst_id, student_id),
			KEY idx_inst_student_face_profile_student (student_id),
			KEY idx_inst_student_face_profile_inst (inst_id)
		)
	`)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS inst_student_face_attendance_session (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			inst_id BIGINT NOT NULL,
			student_id BIGINT NOT NULL,
			student_name VARCHAR(100) NOT NULL DEFAULT '',
			attendance_date DATE NOT NULL,
			status INT NOT NULL DEFAULT 1,
			sign_in_time DATETIME NULL,
			sign_in_image VARCHAR(1024) NOT NULL DEFAULT '',
			sign_out_time DATETIME NULL,
			sign_out_image VARCHAR(1024) NOT NULL DEFAULT '',
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT NOT NULL DEFAULT 0,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uk_inst_student_face_attendance_session_day (inst_id, student_id, attendance_date),
			KEY idx_inst_student_face_attendance_session_inst_day (inst_id, attendance_date),
			KEY idx_inst_student_face_attendance_session_student (inst_id, student_id)
		)
	`)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS inst_student_face_roll_call_task (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			inst_id BIGINT NOT NULL,
			attendance_session_id BIGINT NOT NULL,
			student_id BIGINT NOT NULL,
			attendance_date DATE NOT NULL,
			teaching_schedule_id BIGINT NOT NULL,
			execute_at DATETIME NOT NULL,
			sign_in_time DATETIME NULL,
			status INT NOT NULL DEFAULT 1,
			teaching_record_id BIGINT NOT NULL DEFAULT 0,
			last_error VARCHAR(500) NOT NULL DEFAULT '',
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT NOT NULL DEFAULT 0,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uk_inst_student_face_roll_call_task (inst_id, attendance_session_id, teaching_schedule_id),
			KEY idx_inst_student_face_roll_call_task_execute (inst_id, status, execute_at, id),
			KEY idx_inst_student_face_roll_call_task_session (inst_id, attendance_session_id),
			KEY idx_inst_student_face_roll_call_task_student (inst_id, student_id, attendance_date)
		)
	`)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS inst_student_face_attendance_record (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			inst_id BIGINT NOT NULL,
			student_id BIGINT NOT NULL,
			student_name VARCHAR(100) NOT NULL DEFAULT '',
			face_image VARCHAR(1024) NOT NULL DEFAULT '',
			record_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT NOT NULL DEFAULT 0,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			KEY idx_inst_student_face_attendance_record_inst_time (inst_id, record_time),
			KEY idx_inst_student_face_attendance_record_student (student_id)
		)
	`)
	return err
}

func (repo *Repository) PageFaceCollectionStudents(ctx context.Context, instID int64, query model.FaceCollectionStudentQueryDTO) (model.PageResult[model.FaceCollectionStudent], error) {
	current := query.PageRequestModel.PageIndex
	size := query.PageRequestModel.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 20
	}
	offset := (current - 1) * size

	filters := []string{
		"s.del_flag = 0",
		"s.inst_id = ?",
		"s.student_status IN (1, 2)",
	}
	args := []any{instID}
	if searchKey := strings.TrimSpace(query.QueryModel.SearchKey); searchKey != "" {
		filters = append(filters, "(s.stu_name LIKE ? OR s.mobile LIKE ?)")
		args = append(args, "%"+searchKey+"%", "%"+searchKey+"%")
	}
	whereClause := strings.Join(filters, " AND ")

	var total int
	if err := repo.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM inst_student s WHERE `+whereClause, args...).Scan(&total); err != nil {
		return model.PageResult[model.FaceCollectionStudent]{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT s.id, IFNULL(s.stu_name, ''), IFNULL(s.avatar_url, ''), IFNULL(s.mobile, ''), IFNULL(s.is_collect, 0)
		FROM inst_student s
		WHERE `+whereClause+`
		ORDER BY IFNULL(s.is_collect, 0) DESC, IFNULL(s.stu_name, '') ASC, s.id DESC
		LIMIT ? OFFSET ?
	`, append(args, size, offset)...)
	if err != nil {
		return model.PageResult[model.FaceCollectionStudent]{}, err
	}
	defer rows.Close()

	items := make([]model.FaceCollectionStudent, 0, size)
	for rows.Next() {
		var item model.FaceCollectionStudent
		if err := rows.Scan(&item.ID, &item.StuName, &item.AvatarURL, &item.Mobile, &item.IsCollect); err != nil {
			return model.PageResult[model.FaceCollectionStudent]{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return model.PageResult[model.FaceCollectionStudent]{}, err
	}
	return model.PageResult[model.FaceCollectionStudent]{
		Items:   items,
		Total:   total,
		Current: current,
		Size:    size,
	}, nil
}

func (repo *Repository) GetFaceCollectionProfile(ctx context.Context, instID, studentID int64) (model.FaceCollectionProfile, error) {
	var (
		item           model.FaceCollectionProfile
		descriptorJSON string
		updateTime     sql.NullTime
	)
	err := repo.db.QueryRowContext(ctx, `
		SELECT p.student_id, IFNULL(s.stu_name, ''), IFNULL(p.face_descriptor, ''), IFNULL(p.face_image, ''), p.update_time
		FROM inst_student_face_profile p
		LEFT JOIN inst_student s ON s.id = p.student_id
		WHERE p.inst_id = ? AND p.student_id = ? AND p.del_flag = 0
		LIMIT 1
	`, instID, studentID).Scan(&item.StudentID, &item.StuName, &descriptorJSON, &item.FaceImage, &updateTime)
	if err != nil {
		return model.FaceCollectionProfile{}, err
	}
	if strings.TrimSpace(descriptorJSON) != "" {
		if err := json.Unmarshal([]byte(descriptorJSON), &item.FaceDescriptor); err != nil {
			return model.FaceCollectionProfile{}, err
		}
	}
	if updateTime.Valid {
		t := updateTime.Time
		item.UpdatedTime = &t
	}
	return item, nil
}

func (repo *Repository) CompareFaceCollectionProfile(ctx context.Context, instID int64, descriptor []float32) (model.FaceCollectionCompareResult, error) {
	if len(descriptor) == 0 {
		return model.FaceCollectionCompareResult{}, errors.New("faceDescriptor 不能为空")
	}
	rows, err := repo.db.QueryContext(ctx, `
		SELECT p.student_id, IFNULL(s.stu_name, ''), IFNULL(p.face_descriptor, '')
		FROM inst_student_face_profile p
		LEFT JOIN inst_student s ON s.id = p.student_id
		WHERE p.inst_id = ? AND p.del_flag = 0
	`, instID)
	if err != nil {
		return model.FaceCollectionCompareResult{}, err
	}
	defer rows.Close()

	bestResult := model.FaceCollectionCompareResult{Matched: false}
	minDistance := 0.6

	for rows.Next() {
		var (
			studentID      int64
			studentName    string
			descriptorJSON string
			stored         []float32
		)
		if err := rows.Scan(&studentID, &studentName, &descriptorJSON); err != nil {
			return model.FaceCollectionCompareResult{}, err
		}
		if strings.TrimSpace(descriptorJSON) == "" {
			continue
		}
		if err := json.Unmarshal([]byte(descriptorJSON), &stored); err != nil {
			return model.FaceCollectionCompareResult{}, err
		}
		distance, ok := calculateEuclideanDistance(descriptor, stored)
		if !ok {
			continue
		}
		if distance < minDistance {
			minDistance = distance
			bestResult = model.FaceCollectionCompareResult{
				Matched:     true,
				StudentID:   studentID,
				StudentName: studentName,
				Distance:    distance,
			}
		}
	}
	if err := rows.Err(); err != nil {
		return model.FaceCollectionCompareResult{}, err
	}
	return bestResult, nil
}

func (repo *Repository) ListFaceCollectionProfiles(ctx context.Context, instID int64) ([]model.FaceCollectionProfile, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT p.student_id, IFNULL(s.stu_name, ''), IFNULL(p.face_descriptor, ''), p.update_time
		FROM inst_student_face_profile p
		LEFT JOIN inst_student s ON s.id = p.student_id
		WHERE p.inst_id = ? AND p.del_flag = 0
		ORDER BY p.update_time DESC, p.id DESC
	`, instID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.FaceCollectionProfile, 0)
	for rows.Next() {
		var (
			item           model.FaceCollectionProfile
			descriptorJSON string
			updateTime     sql.NullTime
		)
		if err := rows.Scan(&item.StudentID, &item.StuName, &descriptorJSON, &updateTime); err != nil {
			return nil, err
		}
		if strings.TrimSpace(descriptorJSON) != "" {
			if err := json.Unmarshal([]byte(descriptorJSON), &item.FaceDescriptor); err != nil {
				return nil, err
			}
		}
		if updateTime.Valid {
			t := updateTime.Time
			item.UpdatedTime = &t
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repo *Repository) SaveFaceCollectionProfile(ctx context.Context, instID, operatorID int64, dto model.FaceCollectionProfileSaveDTO) error {
	if dto.StudentID <= 0 {
		return errors.New("invalid studentId")
	}
	if len(dto.FaceDescriptor) == 0 {
		return errors.New("faceDescriptor 不能为空")
	}
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM inst_student
		WHERE id = ? AND inst_id = ? AND del_flag = 0 AND student_status IN (1, 2)
	`, dto.StudentID, instID).Scan(&exists); err != nil {
		return err
	}
	if exists == 0 {
		return errors.New("学员不存在或当前状态不支持人脸采集")
	}

	descriptorJSON, err := json.Marshal(dto.FaceDescriptor)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO inst_student_face_profile (
			inst_id, student_id, face_descriptor, face_image, create_id, update_id, del_flag
		) VALUES (?, ?, ?, ?, ?, ?, 0)
		ON DUPLICATE KEY UPDATE
			face_descriptor = VALUES(face_descriptor),
			face_image = VALUES(face_image),
			update_id = VALUES(update_id),
			update_time = CURRENT_TIMESTAMP,
			del_flag = 0
	`, instID, dto.StudentID, string(descriptorJSON), dto.FaceImage, operatorID, operatorID); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE inst_student
		SET is_collect = 1
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, dto.StudentID, instID); err != nil {
		return err
	}

	return tx.Commit()
}

func (repo *Repository) DeleteFaceCollectionProfile(ctx context.Context, instID, operatorID int64, studentID int64) error {
	if studentID <= 0 {
		return errors.New("invalid studentId")
	}
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		UPDATE inst_student_face_profile
		SET del_flag = 1, update_id = ?, update_time = CURRENT_TIMESTAMP
		WHERE inst_id = ? AND student_id = ? AND del_flag = 0
	`, operatorID, instID, studentID); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE inst_student
		SET is_collect = 0
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, studentID, instID); err != nil {
		return err
	}

	return tx.Commit()
}

func (repo *Repository) ListFaceAttendanceSessions(ctx context.Context, instID int64, limit int) ([]model.FaceAttendanceSession, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			fas.id,
			fas.student_id,
			IFNULL(fas.student_name, ''),
			IFNULL(s.avatar_url, ''),
			DATE_FORMAT(fas.attendance_date, '%Y-%m-%d'),
			IFNULL(fas.status, 0),
			fas.sign_in_time,
			IFNULL(fas.sign_in_image, ''),
			fas.sign_out_time,
			IFNULL(fas.sign_out_image, ''),
			EXISTS(
				SELECT 1
				FROM teaching_schedule ts
				LEFT JOIN teaching_schedule_student tss
					ON tss.teaching_schedule_id = ts.id
				   AND tss.inst_id = ts.inst_id
				   AND tss.del_flag = 0
				WHERE ts.inst_id = fas.inst_id
				  AND ts.del_flag = 0
				  AND ts.status = ?
				  AND DATE(ts.lesson_date) = fas.attendance_date
				  AND (
					ts.student_id = fas.student_id
					OR (
						tss.student_id = fas.student_id
						AND IFNULL(tss.roster_status, ?) <> ?
					)
				  )
				LIMIT 1
			) AS has_schedule
		FROM inst_student_face_attendance_session fas
		LEFT JOIN inst_student s
			ON s.id = fas.student_id
		   AND s.inst_id = fas.inst_id
		   AND s.del_flag = 0
		WHERE fas.inst_id = ? AND fas.del_flag = 0
		ORDER BY COALESCE(fas.sign_out_time, fas.sign_in_time, fas.update_time) DESC, fas.id DESC
		LIMIT ?
	`, model.TeachingScheduleStatusActive, model.TeachingScheduleStudentRosterStatusActive, model.TeachingScheduleStudentRosterStatusRemoved, instID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.FaceAttendanceSession, 0, limit)
	for rows.Next() {
		item, err := scanFaceAttendanceSession(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repo *Repository) RecognizeFaceAttendanceSession(ctx context.Context, instID int64, dto model.FaceAttendanceSessionRecognizeDTO) (model.FaceAttendanceSessionRecognizeResult, error) {
	compareResult, err := repo.CompareFaceCollectionProfile(ctx, instID, dto.FaceDescriptor)
	if err != nil {
		return model.FaceAttendanceSessionRecognizeResult{}, err
	}
	if !compareResult.Matched || compareResult.StudentID <= 0 {
		return model.FaceAttendanceSessionRecognizeResult{
			Matched:     false,
			Action:      model.FaceAttendanceSessionActionNoMatch,
			ActionLabel: faceAttendanceActionLabel(model.FaceAttendanceSessionActionNoMatch),
			Message:     "未能识别该人脸，请确保已完成人脸采集",
		}, nil
	}

	student, err := repo.getFaceAttendanceStudentBase(ctx, repo.db, instID, compareResult.StudentID)
	if err != nil {
		return model.FaceAttendanceSessionRecognizeResult{}, err
	}

	now := time.Now()
	attendanceDate := now.Format("2006-01-02")
	record, found, err := repo.getFaceAttendanceRecordForDate(ctx, repo.db, instID, compareResult.StudentID, attendanceDate)
	if err != nil {
		return model.FaceAttendanceSessionRecognizeResult{}, err
	}
	hasSchedule, lastLessonEndTime, err := repo.getStudentLastLessonEndTime(ctx, repo.db, instID, compareResult.StudentID, attendanceDate)
	if err != nil {
		return model.FaceAttendanceSessionRecognizeResult{}, err
	}

	result := model.FaceAttendanceSessionRecognizeResult{
		Matched:           true,
		StudentID:         compareResult.StudentID,
		StudentName:       student.StudentName,
		AvatarURL:         student.AvatarURL,
		Distance:          compareResult.Distance,
		HasSchedule:       hasSchedule,
		LastLessonEndTime: lastLessonEndTime,
	}

	if !found {
		if !hasSchedule {
			result.Action = model.FaceAttendanceSessionActionIgnore
			result.ActionLabel = faceAttendanceActionLabel(result.Action)
			result.Message = "今日无日程，无需签到"
			return result, nil
		}
		if !canFaceAttendanceSignIn(now, lastLessonEndTime) {
			result.Action = model.FaceAttendanceSessionActionIgnore
			result.ActionLabel = faceAttendanceActionLabel(result.Action)
			result.Message = "今日课程已结束，无需签到"
			return result, nil
		}
		result.Action = model.FaceAttendanceSessionActionSignIn
		result.ActionLabel = faceAttendanceActionLabel(result.Action)
		result.NeedUpload = true
		result.Message = "识别成功，准备签到"
		return result, nil
	}

	result.SessionID = record.ID
	result.SessionNo = 1

	if record.Status == model.FaceAttendanceSessionStatusSignedIn {
		result.LastActionTime = record.SignInTime
		if withinDuplicateWindow(now, record.SignInTime) {
			result.Action = model.FaceAttendanceSessionActionDuplicateSignIn
			result.ActionLabel = faceAttendanceActionLabel(result.Action)
			result.Message = "已签到，请勿重复签到"
			return result, nil
		}
		if canFaceAttendanceSignOut(now, hasSchedule, lastLessonEndTime) {
			result.Action = model.FaceAttendanceSessionActionSignOut
			result.ActionLabel = faceAttendanceActionLabel(result.Action)
			result.NeedUpload = true
			result.Message = "识别成功，准备签退"
			return result, nil
		}
		result.Action = model.FaceAttendanceSessionActionIgnore
		result.ActionLabel = faceAttendanceActionLabel(result.Action)
		if hasSchedule && lastLessonEndTime != nil {
			result.Message = "已签到，请在最后一节课结束后再签退"
		} else {
			result.Message = "已签到，请勿重复操作"
		}
		return result, nil
	}

	result.LastActionTime = record.SignOutTime
	if shouldOverwriteFaceAttendanceSignOut(record.SignOutTime, lastLessonEndTime) && canFaceAttendanceSignOut(now, hasSchedule, lastLessonEndTime) {
		result.Action = model.FaceAttendanceSessionActionSignOut
		result.ActionLabel = faceAttendanceActionLabel(result.Action)
		result.NeedUpload = true
		result.Message = "识别成功，准备更新签退"
		return result, nil
	}
	if withinDuplicateWindow(now, record.SignOutTime) {
		result.Action = model.FaceAttendanceSessionActionDuplicateSignOut
		result.ActionLabel = faceAttendanceActionLabel(result.Action)
		result.Message = "已签退，请勿重复签退"
		return result, nil
	}
	result.Action = model.FaceAttendanceSessionActionIgnore
	result.ActionLabel = faceAttendanceActionLabel(result.Action)
	if shouldOverwriteFaceAttendanceSignOut(record.SignOutTime, lastLessonEndTime) {
		result.Message = "已签到，请在最后一节课结束后再签退"
	} else {
		result.Message = "已签退，请勿重复操作"
	}
	return result, nil
}

func (repo *Repository) CommitFaceAttendanceSession(ctx context.Context, instID, operatorID int64, dto model.FaceAttendanceSessionCommitDTO) (model.FaceAttendanceSession, error) {
	if dto.StudentID <= 0 {
		return model.FaceAttendanceSession{}, errors.New("studentId 无效")
	}
	if strings.TrimSpace(dto.FaceImage) == "" {
		return model.FaceAttendanceSession{}, errors.New("faceImage 不能为空")
	}
	if dto.Action != model.FaceAttendanceSessionActionSignIn && dto.Action != model.FaceAttendanceSessionActionSignOut {
		return model.FaceAttendanceSession{}, errors.New("action 无效")
	}

	student, err := repo.getFaceAttendanceStudentBase(ctx, repo.db, instID, dto.StudentID)
	if err != nil {
		return model.FaceAttendanceSession{}, err
	}

	now := time.Now()
	attendanceDate := now.Format("2006-01-02")

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return model.FaceAttendanceSession{}, err
	}
	defer tx.Rollback()

	record, found, err := repo.getFaceAttendanceRecordForDate(ctx, tx, instID, dto.StudentID, attendanceDate)
	if err != nil {
		return model.FaceAttendanceSession{}, err
	}
	hasSchedule, lastLessonEndTime, err := repo.getStudentLastLessonEndTime(ctx, tx, instID, dto.StudentID, attendanceDate)
	if err != nil {
		return model.FaceAttendanceSession{}, err
	}

	switch dto.Action {
	case model.FaceAttendanceSessionActionSignIn:
		if !hasSchedule {
			return model.FaceAttendanceSession{}, errors.New("今日无日程，无需签到")
		}
		if !canFaceAttendanceSignIn(now, lastLessonEndTime) {
			return model.FaceAttendanceSession{}, errors.New("今日课程已结束，无需签到")
		}
		if found {
			if record.Status == model.FaceAttendanceSessionStatusSignedIn {
				return model.FaceAttendanceSession{}, errors.New("今日已签到，无需重复提交")
			}
			return model.FaceAttendanceSession{}, errors.New("今日已有签到记录，不支持覆盖签到")
		}
		result, err := tx.ExecContext(ctx, `
			INSERT INTO inst_student_face_attendance_session (
				inst_id, student_id, student_name, attendance_date, status, sign_in_time, sign_in_image,
				create_id, update_id, del_flag
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
		`, instID, dto.StudentID, student.StudentName, attendanceDate, model.FaceAttendanceSessionStatusSignedIn, now, strings.TrimSpace(dto.FaceImage), operatorID, operatorID)
		if err != nil {
			return model.FaceAttendanceSession{}, err
		}
		recordID, err := result.LastInsertId()
		if err != nil {
			return model.FaceAttendanceSession{}, err
		}
		if _, err := repo.enqueueFaceAttendanceAutoRollCallTasksTx(ctx, tx, instID, operatorID, recordID, dto.StudentID, attendanceDate, now); err != nil {
			return model.FaceAttendanceSession{}, err
		}
		if err := tx.Commit(); err != nil {
			return model.FaceAttendanceSession{}, err
		}
		_, _ = repo.ProcessFaceAttendanceAutoRollCallTasks(ctx, operatorID, now, &instID, &recordID)
		return repo.GetFaceAttendanceSessionByID(ctx, instID, recordID)
	case model.FaceAttendanceSessionActionSignOut:
		if !found {
			return model.FaceAttendanceSession{}, errors.New("今日暂无签到记录，无法签退")
		}
		if dto.SessionID > 0 && record.ID != dto.SessionID {
			return model.FaceAttendanceSession{}, errors.New("考勤记录已变化，请重新识别")
		}
		if record.Status == model.FaceAttendanceSessionStatusSignedIn {
			if !canFaceAttendanceSignOut(now, hasSchedule, lastLessonEndTime) {
				return model.FaceAttendanceSession{}, errors.New("已签到，请在最后一节课结束后再签退")
			}
		} else {
			if !shouldOverwriteFaceAttendanceSignOut(record.SignOutTime, lastLessonEndTime) {
				return model.FaceAttendanceSession{}, errors.New("已签退，请勿重复操作")
			}
			if !canFaceAttendanceSignOut(now, hasSchedule, lastLessonEndTime) {
				return model.FaceAttendanceSession{}, errors.New("已签到，请在最后一节课结束后再签退")
			}
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE inst_student_face_attendance_session
			SET status = ?, sign_out_time = ?, sign_out_image = ?, update_id = ?, update_time = CURRENT_TIMESTAMP
			WHERE inst_id = ? AND id = ? AND del_flag = 0
		`, model.FaceAttendanceSessionStatusSignedOut, now, strings.TrimSpace(dto.FaceImage), operatorID, instID, record.ID); err != nil {
			return model.FaceAttendanceSession{}, err
		}
		if err := tx.Commit(); err != nil {
			return model.FaceAttendanceSession{}, err
		}
		return repo.GetFaceAttendanceSessionByID(ctx, instID, record.ID)
	default:
		return model.FaceAttendanceSession{}, errors.New("action 无效")
	}
}

func (repo *Repository) GetFaceAttendanceSessionByID(ctx context.Context, instID, recordID int64) (model.FaceAttendanceSession, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT
			fas.id,
			fas.student_id,
			IFNULL(fas.student_name, ''),
			IFNULL(s.avatar_url, ''),
			DATE_FORMAT(fas.attendance_date, '%Y-%m-%d'),
			IFNULL(fas.status, 0),
			fas.sign_in_time,
			IFNULL(fas.sign_in_image, ''),
			fas.sign_out_time,
			IFNULL(fas.sign_out_image, ''),
			EXISTS(
				SELECT 1
				FROM teaching_schedule ts
				LEFT JOIN teaching_schedule_student tss
					ON tss.teaching_schedule_id = ts.id
				   AND tss.inst_id = ts.inst_id
				   AND tss.del_flag = 0
				WHERE ts.inst_id = fas.inst_id
				  AND ts.del_flag = 0
				  AND ts.status = ?
				  AND DATE(ts.lesson_date) = fas.attendance_date
				  AND (
					ts.student_id = fas.student_id
					OR (
						tss.student_id = fas.student_id
						AND IFNULL(tss.roster_status, ?) <> ?
					)
				  )
				LIMIT 1
			) AS has_schedule
		FROM inst_student_face_attendance_session fas
		LEFT JOIN inst_student s
			ON s.id = fas.student_id
		   AND s.inst_id = fas.inst_id
		   AND s.del_flag = 0
		WHERE fas.inst_id = ? AND fas.id = ? AND fas.del_flag = 0
		LIMIT 1
	`, model.TeachingScheduleStatusActive, model.TeachingScheduleStudentRosterStatusActive, model.TeachingScheduleStudentRosterStatusRemoved, instID, recordID)
	item, err := scanFaceAttendanceSession(row)
	if err != nil {
		return model.FaceAttendanceSession{}, err
	}
	_, lastLessonEndTime, err := repo.getStudentLastLessonEndTime(ctx, repo.db, instID, item.StudentID, item.AttendanceDate)
	if err == nil {
		item.LastLessonEndTime = lastLessonEndTime
	}
	return item, nil
}

func (repo *Repository) ListFaceAttendanceRecords(ctx context.Context, instID int64, limit int) ([]model.FaceAttendanceRecord, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, student_id, IFNULL(student_name, ''), IFNULL(face_image, ''), record_time
		FROM inst_student_face_attendance_record
		WHERE inst_id = ? AND del_flag = 0
		ORDER BY record_time DESC, id DESC
		LIMIT ?
	`, instID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.FaceAttendanceRecord, 0, limit)
	for rows.Next() {
		var (
			item       model.FaceAttendanceRecord
			recordTime sql.NullTime
		)
		if err := rows.Scan(&item.ID, &item.StudentID, &item.StudentName, &item.FaceImage, &recordTime); err != nil {
			return nil, err
		}
		if recordTime.Valid {
			t := recordTime.Time
			item.RecordTime = &t
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repo *Repository) SaveFaceAttendanceRecord(ctx context.Context, instID, operatorID int64, dto model.FaceAttendanceRecordSaveDTO) (model.FaceAttendanceRecord, error) {
	if dto.StudentID <= 0 {
		return model.FaceAttendanceRecord{}, errors.New("invalid studentId")
	}
	if strings.TrimSpace(dto.FaceImage) == "" {
		return model.FaceAttendanceRecord{}, errors.New("faceImage 不能为空")
	}

	var studentName string
	if err := repo.db.QueryRowContext(ctx, `
		SELECT IFNULL(stu_name, '')
		FROM inst_student
		WHERE id = ? AND inst_id = ? AND del_flag = 0 AND student_status IN (1, 2)
		LIMIT 1
	`, dto.StudentID, instID).Scan(&studentName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.FaceAttendanceRecord{}, errors.New("学员不存在或当前状态不支持人脸考勤")
		}
		return model.FaceAttendanceRecord{}, err
	}

	result, err := repo.db.ExecContext(ctx, `
		INSERT INTO inst_student_face_attendance_record (
			inst_id, student_id, student_name, face_image, record_time, create_id, update_id, del_flag
		) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, ?, ?, 0)
	`, instID, dto.StudentID, studentName, strings.TrimSpace(dto.FaceImage), operatorID, operatorID)
	if err != nil {
		return model.FaceAttendanceRecord{}, err
	}
	recordID, err := result.LastInsertId()
	if err != nil {
		return model.FaceAttendanceRecord{}, err
	}

	return repo.GetFaceAttendanceRecordByID(ctx, instID, recordID)
}

func (repo *Repository) GetFaceAttendanceRecordByID(ctx context.Context, instID, recordID int64) (model.FaceAttendanceRecord, error) {
	var (
		item       model.FaceAttendanceRecord
		recordTime sql.NullTime
	)
	err := repo.db.QueryRowContext(ctx, `
		SELECT id, student_id, IFNULL(student_name, ''), IFNULL(face_image, ''), record_time
		FROM inst_student_face_attendance_record
		WHERE inst_id = ? AND id = ? AND del_flag = 0
		LIMIT 1
	`, instID, recordID).Scan(&item.ID, &item.StudentID, &item.StudentName, &item.FaceImage, &recordTime)
	if err != nil {
		return model.FaceAttendanceRecord{}, err
	}
	if recordTime.Valid {
		t := recordTime.Time
		item.RecordTime = &t
	}
	return item, nil
}

func (repo *Repository) getFaceAttendanceStudentBase(ctx context.Context, runner faceAttendanceQueryRunner, instID, studentID int64) (faceAttendanceStudentBase, error) {
	var item faceAttendanceStudentBase
	err := runner.QueryRowContext(ctx, `
		SELECT id, IFNULL(stu_name, ''), IFNULL(avatar_url, '')
		FROM inst_student
		WHERE id = ? AND inst_id = ? AND del_flag = 0 AND student_status IN (1, 2)
		LIMIT 1
	`, studentID, instID).Scan(&item.StudentID, &item.StudentName, &item.AvatarURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return faceAttendanceStudentBase{}, errors.New("学员不存在或当前状态不支持人脸考勤")
		}
		return faceAttendanceStudentBase{}, err
	}
	return item, nil
}

func (repo *Repository) getFaceAttendanceRecordForDate(ctx context.Context, runner faceAttendanceQueryRunner, instID, studentID int64, attendanceDate string) (model.FaceAttendanceSession, bool, error) {
	row := runner.QueryRowContext(ctx, `
		SELECT
			id,
			student_id,
			IFNULL(student_name, ''),
			'' AS avatar_url,
			DATE_FORMAT(attendance_date, '%Y-%m-%d'),
			IFNULL(status, 0),
			sign_in_time,
			IFNULL(sign_in_image, ''),
			sign_out_time,
			IFNULL(sign_out_image, ''),
			0 AS has_schedule
		FROM inst_student_face_attendance_session
		WHERE inst_id = ? AND student_id = ? AND attendance_date = ? AND del_flag = 0
		LIMIT 1
	`, instID, studentID, attendanceDate)
	item, err := scanFaceAttendanceSession(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.FaceAttendanceSession{}, false, nil
		}
		return model.FaceAttendanceSession{}, false, err
	}
	return item, true, nil
}

func (repo *Repository) getStudentLastLessonEndTime(ctx context.Context, runner faceAttendanceQueryRunner, instID, studentID int64, attendanceDate string) (bool, *time.Time, error) {
	var (
		total         int
		lastLessonEnd sql.NullTime
	)
	err := runner.QueryRowContext(ctx, `
		SELECT COUNT(*), MAX(ts.lesson_end_at)
		FROM teaching_schedule ts
		LEFT JOIN teaching_schedule_student tss
			ON tss.teaching_schedule_id = ts.id
		   AND tss.inst_id = ts.inst_id
		   AND tss.del_flag = 0
		WHERE ts.inst_id = ?
		  AND ts.del_flag = 0
		  AND ts.status = ?
		  AND DATE(ts.lesson_date) = ?
		  AND (
			ts.student_id = ?
			OR (
				tss.student_id = ?
				AND IFNULL(tss.roster_status, ?) <> ?
			)
		  )
	`, instID, model.TeachingScheduleStatusActive, attendanceDate, studentID, studentID, model.TeachingScheduleStudentRosterStatusActive, model.TeachingScheduleStudentRosterStatusRemoved).Scan(&total, &lastLessonEnd)
	if err != nil {
		return false, nil, err
	}
	if !lastLessonEnd.Valid || total == 0 {
		return false, nil, nil
	}
	t := lastLessonEnd.Time
	return true, &t, nil
}

func scanFaceAttendanceSession(scanner interface {
	Scan(dest ...any) error
}) (model.FaceAttendanceSession, error) {
	var (
		item        model.FaceAttendanceSession
		signInTime  sql.NullTime
		signOutTime sql.NullTime
		hasSchedule int
	)
	if err := scanner.Scan(
		&item.ID,
		&item.StudentID,
		&item.StudentName,
		&item.AvatarURL,
		&item.AttendanceDate,
		&item.Status,
		&signInTime,
		&item.SignInImage,
		&signOutTime,
		&item.SignOutImage,
		&hasSchedule,
	); err != nil {
		return model.FaceAttendanceSession{}, err
	}
	item.SessionNo = 1
	item.HasSchedule = hasSchedule > 0
	if signInTime.Valid {
		t := signInTime.Time
		item.SignInTime = &t
	}
	if signOutTime.Valid {
		t := signOutTime.Time
		item.SignOutTime = &t
	}
	fillFaceAttendanceSessionDerivedFields(&item)
	return item, nil
}

func fillFaceAttendanceSessionDerivedFields(item *model.FaceAttendanceSession) {
	if item == nil {
		return
	}
	if item.SignOutTime != nil {
		item.LatestAction = model.FaceAttendanceSessionActionSignOut
		item.LatestActionLabel = faceAttendanceActionLabel(item.LatestAction)
		item.LatestTime = item.SignOutTime
		item.LatestImage = strings.TrimSpace(item.SignOutImage)
		if item.LatestImage == "" {
			item.LatestImage = strings.TrimSpace(item.SignInImage)
		}
		return
	}
	item.LatestAction = model.FaceAttendanceSessionActionSignIn
	item.LatestActionLabel = faceAttendanceActionLabel(item.LatestAction)
	item.LatestTime = item.SignInTime
	item.LatestImage = strings.TrimSpace(item.SignInImage)
}

func faceAttendanceActionLabel(action string) string {
	switch action {
	case model.FaceAttendanceSessionActionSignIn:
		return "自动签到"
	case model.FaceAttendanceSessionActionSignOut:
		return "自动签退"
	case model.FaceAttendanceSessionActionDuplicateSignIn:
		return "重复签到"
	case model.FaceAttendanceSessionActionDuplicateSignOut:
		return "重复签退"
	case model.FaceAttendanceSessionActionIgnore:
		return "无需处理"
	case model.FaceAttendanceSessionActionNoMatch:
		return "未识别"
	default:
		return ""
	}
}

func withinDuplicateWindow(now time.Time, lastTime *time.Time) bool {
	if lastTime == nil || lastTime.IsZero() {
		return false
	}
	delta := now.Sub(*lastTime)
	return delta >= 0 && delta < faceAttendanceDuplicateWindow
}

func canFaceAttendanceSignIn(now time.Time, lastLessonEndTime *time.Time) bool {
	if lastLessonEndTime == nil || lastLessonEndTime.IsZero() {
		return false
	}
	return !now.After(*lastLessonEndTime)
}

func canFaceAttendanceSignOut(now time.Time, hasSchedule bool, lastLessonEndTime *time.Time) bool {
	if !hasSchedule || lastLessonEndTime == nil || lastLessonEndTime.IsZero() {
		return true
	}
	return !now.Before(lastLessonEndTime.Add(faceAttendanceSignOutGrace))
}

func shouldOverwriteFaceAttendanceSignOut(signOutTime *time.Time, lastLessonEndTime *time.Time) bool {
	if signOutTime == nil || signOutTime.IsZero() || lastLessonEndTime == nil || lastLessonEndTime.IsZero() {
		return false
	}
	return lastLessonEndTime.After(*signOutTime)
}

func calculateEuclideanDistance(current, stored []float32) (float64, bool) {
	if len(current) == 0 || len(current) != len(stored) {
		return 0, false
	}
	var sum float64
	for i := range current {
		diff := float64(current[i] - stored[i])
		sum += diff * diff
	}
	return math.Sqrt(sum), true
}
