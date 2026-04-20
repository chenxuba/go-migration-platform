package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"go-migration-platform/services/education/internal/model"
)

type leaveStudentSnapshot struct {
	ID        int64
	Name      string
	AvatarURL string
	Phone     string
}

type leaveOperatorSnapshot struct {
	ID   int64
	Name string
}

type leaveApplicableSchedule struct {
	ScheduleID         int64
	ClassType          int
	TeachingClassID    int64
	TeachingClassName  string
	LessonID           int64
	LessonName         string
	TeacherID          int64
	TeacherName        string
	StartTime          time.Time
	EndTime            time.Time
	RosterStatusBefore int
}

func ensureLeaveTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS inst_leave_request (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			uuid VARCHAR(64) NULL,
			version BIGINT NOT NULL DEFAULT 0,
			inst_id BIGINT NOT NULL,
			student_id BIGINT NOT NULL,
			student_name VARCHAR(100) NOT NULL DEFAULT '',
			student_avatar_url VARCHAR(255) NOT NULL DEFAULT '',
			student_phone VARCHAR(32) NOT NULL DEFAULT '',
			start_time DATETIME NOT NULL,
			end_time DATETIME NOT NULL,
			leave_type INT NOT NULL DEFAULT 0,
			reason VARCHAR(1000) NOT NULL DEFAULT '',
			proof_materials_json LONGTEXT NULL,
			remark VARCHAR(100) NOT NULL DEFAULT '',
			schedule_count INT NOT NULL DEFAULT 0,
			initiate_staff_id BIGINT NOT NULL DEFAULT 0,
			initiate_staff_name VARCHAR(100) NOT NULL DEFAULT '',
			is_agent TINYINT(1) NOT NULL DEFAULT 1,
			status INT NOT NULL DEFAULT 1,
			approval_config_version INT NOT NULL DEFAULT 0,
			current_step INT NULL,
			current_approver_ids VARCHAR(255) NOT NULL DEFAULT '',
			current_approver_names VARCHAR(255) NOT NULL DEFAULT '',
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT NOT NULL DEFAULT 0,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			KEY idx_inst_leave_request_inst_status (inst_id, status, create_time),
			KEY idx_inst_leave_request_student (inst_id, student_id),
			KEY idx_inst_leave_request_time (inst_id, start_time, end_time)
		)
	`)
	if err != nil {
		return err
	}
	if err := ensureColumnsOnTable(ctx, db, "inst_leave_request", map[string]string{
		"is_agent": "is_agent TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否代办' AFTER initiate_staff_name",
	}); err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS inst_leave_schedule (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			inst_id BIGINT NOT NULL,
			leave_request_id BIGINT NOT NULL,
			teaching_schedule_id BIGINT NOT NULL,
			class_type INT NOT NULL DEFAULT 0,
			teaching_class_id BIGINT NOT NULL DEFAULT 0,
			teaching_class_name VARCHAR(150) NOT NULL DEFAULT '',
			lesson_id BIGINT NOT NULL DEFAULT 0,
			lesson_name VARCHAR(150) NOT NULL DEFAULT '',
			teacher_id BIGINT NOT NULL DEFAULT 0,
			teacher_name VARCHAR(100) NOT NULL DEFAULT '',
			start_time DATETIME NOT NULL,
			end_time DATETIME NOT NULL,
			roster_status_before INT NOT NULL DEFAULT 1,
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT NOT NULL DEFAULT 0,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uk_inst_leave_schedule_unique (inst_id, leave_request_id, teaching_schedule_id),
			KEY idx_inst_leave_schedule_leave (inst_id, leave_request_id),
			KEY idx_inst_leave_schedule_schedule (inst_id, teaching_schedule_id)
		)
	`)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS inst_leave_action (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			inst_id BIGINT NOT NULL,
			leave_request_id BIGINT NOT NULL,
			action_type INT NOT NULL DEFAULT 0,
			action_staff_id BIGINT NOT NULL DEFAULT 0,
			action_staff_name VARCHAR(100) NOT NULL DEFAULT '',
			status_text VARCHAR(50) NOT NULL DEFAULT '',
			remark VARCHAR(1000) NOT NULL DEFAULT '',
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT NOT NULL DEFAULT 0,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			KEY idx_inst_leave_action_leave (inst_id, leave_request_id, create_time)
		)
	`)
	return err
}

func (repo *Repository) PreviewLeaveSchedules(ctx context.Context, instID int64, dto model.LeavePreviewDTO) ([]model.LeaveScheduleSnapshotVO, error) {
	studentID, startTime, endTime, err := parseLeavePreviewInput(dto)
	if err != nil {
		return nil, err
	}
	schedules, err := repo.listLeaveApplicableSchedules(ctx, repo.db, instID, studentID, startTime, endTime, true)
	if err != nil {
		return nil, err
	}
	result := make([]model.LeaveScheduleSnapshotVO, 0, len(schedules))
	for _, item := range schedules {
		start := item.StartTime
		end := item.EndTime
		result = append(result, model.LeaveScheduleSnapshotVO{
			ScheduleID:         strconv.FormatInt(item.ScheduleID, 10),
			ClassType:          item.ClassType,
			TeachingClassID:    emptyStringIfZero(item.TeachingClassID),
			TeachingClassName:  firstNonEmptyString(strings.TrimSpace(item.TeachingClassName), "-"),
			LessonID:           emptyStringIfZero(item.LessonID),
			LessonName:         firstNonEmptyString(strings.TrimSpace(item.LessonName), "-"),
			TeacherID:          emptyStringIfZero(item.TeacherID),
			TeacherName:        firstNonEmptyString(strings.TrimSpace(item.TeacherName), "-"),
			StartTime:          &start,
			EndTime:            &end,
			RosterStatusBefore: item.RosterStatusBefore,
		})
	}
	return result, nil
}

func (repo *Repository) CreateLeaveRequest(ctx context.Context, instID, operatorID int64, dto model.LeaveCreateDTO) (model.LeaveCreateResult, error) {
	studentID, startTime, endTime, err := parseLeaveCreateInput(dto)
	if err != nil {
		return model.LeaveCreateResult{}, err
	}
	if leaveTypeText(dto.LeaveType) == "-" {
		return model.LeaveCreateResult{}, errors.New("请选择请假类型")
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return model.LeaveCreateResult{}, err
	}
	defer tx.Rollback()

	student, err := repo.getLeaveStudentSnapshotTx(ctx, tx, instID, studentID)
	if err != nil {
		return model.LeaveCreateResult{}, err
	}
	operator, err := repo.getLeaveOperatorSnapshotTx(ctx, tx, instID, operatorID)
	if err != nil {
		return model.LeaveCreateResult{}, err
	}

	schedules, err := repo.listLeaveApplicableSchedules(ctx, tx, instID, studentID, startTime, endTime, true)
	if err != nil {
		return model.LeaveCreateResult{}, err
	}
	if len(schedules) == 0 {
		return model.LeaveCreateResult{}, errors.New("请假期间没有可请假的相关课节")
	}

	currentStep, currentApproverIDs, currentApproverNames, configVersion, err := repo.getLeaveApprovalStepTx(ctx, tx, instID)
	if err != nil {
		return model.LeaveCreateResult{}, err
	}

	status := model.LeaveStatusPending
	if len(currentApproverIDs) == 0 {
		status = model.LeaveStatusApproved
	}
	var currentStepArg any
	if currentStep != nil {
		currentStepArg = *currentStep
	}

	proofMaterialsJSON, err := json.Marshal(compactStrings(dto.ProofMaterials))
	if err != nil {
		return model.LeaveCreateResult{}, err
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO inst_leave_request (
			uuid, version, inst_id, student_id, student_name, student_avatar_url, student_phone,
			start_time, end_time, leave_type, reason, proof_materials_json, remark, schedule_count,
			initiate_staff_id, initiate_staff_name, is_agent, status, approval_config_version, current_step,
			current_approver_ids, current_approver_names, create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), ?, NOW(), 0
		)
	`, instID, student.ID, student.Name, student.AvatarURL, student.Phone,
		startTime, endTime, dto.LeaveType, strings.TrimSpace(dto.Reason), string(proofMaterialsJSON), strings.TrimSpace(dto.Remark), len(schedules),
		operator.ID, operator.Name, true, status, configVersion, currentStepArg,
		joinInt64CSV(currentApproverIDs), strings.TrimSpace(currentApproverNames), operator.ID, operator.ID)
	if err != nil {
		return model.LeaveCreateResult{}, err
	}
	leaveID, err := result.LastInsertId()
	if err != nil {
		return model.LeaveCreateResult{}, err
	}

	for _, item := range schedules {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO inst_leave_schedule (
				inst_id, leave_request_id, teaching_schedule_id, class_type, teaching_class_id,
				teaching_class_name, lesson_id, lesson_name, teacher_id, teacher_name,
				start_time, end_time, roster_status_before, create_id, update_id, del_flag
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
		`, instID, leaveID, item.ScheduleID, item.ClassType, item.TeachingClassID,
			item.TeachingClassName, item.LessonID, item.LessonName, item.TeacherID, item.TeacherName,
			item.StartTime, item.EndTime, item.RosterStatusBefore, operator.ID, operator.ID); err != nil {
			return model.LeaveCreateResult{}, err
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO teaching_schedule_student (
				inst_id, teaching_schedule_id, teaching_class_id, student_id, student_type, roster_status,
				create_id, update_id, del_flag
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0)
			ON DUPLICATE KEY UPDATE
				teaching_class_id = VALUES(teaching_class_id),
				student_type = VALUES(student_type),
				roster_status = VALUES(roster_status),
				update_id = VALUES(update_id),
				update_time = CURRENT_TIMESTAMP,
				del_flag = 0
		`, instID, item.ScheduleID, item.TeachingClassID, student.ID, model.TeachingScheduleStudentTypeClassMember,
			model.TeachingScheduleStudentRosterStatusLeave, operator.ID, operator.ID); err != nil {
			return model.LeaveCreateResult{}, err
		}
	}

	if err := repo.insertLeaveActionTx(ctx, tx, instID, leaveID, model.LeaveActionCreate, operator.ID, operator.Name, "发起", ""); err != nil {
		return model.LeaveCreateResult{}, err
	}
	if status == model.LeaveStatusApproved {
		if err := repo.insertLeaveActionTx(ctx, tx, instID, leaveID, model.LeaveActionAutoApprove, 0, "系统自动执行", "已通过", "未配置审批人，系统自动通过"); err != nil {
			return model.LeaveCreateResult{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return model.LeaveCreateResult{}, err
	}
	return model.LeaveCreateResult{
		ID:     strconv.FormatInt(leaveID, 10),
		Status: status,
	}, nil
}

func (repo *Repository) CancelLeaveRequest(ctx context.Context, instID, operatorID int64, dto model.LeaveCancelDTO) error {
	leaveID, remark, err := parseLeaveCancelInput(dto)
	if err != nil {
		return err
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var (
		studentID   int64
		leaveStatus int
	)
	if err := tx.QueryRowContext(ctx, `
		SELECT student_id, IFNULL(status, 0)
		FROM inst_leave_request
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
		FOR UPDATE
	`, leaveID, instID).Scan(&studentID, &leaveStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("未找到请假记录")
		}
		return err
	}

	switch leaveStatus {
	case model.LeaveStatusApproved:
	case model.LeaveStatusRevoked:
		return errors.New("请假记录已撤销，请勿重复操作")
	default:
		return errors.New("仅已通过的请假可撤销")
	}

	operator, err := repo.getLeaveOperatorSnapshotTx(ctx, tx, instID, operatorID)
	if err != nil {
		return err
	}

	scheduleRows, err := tx.QueryContext(ctx, `
		SELECT
			teaching_schedule_id,
			IFNULL(roster_status_before, 1),
			end_time
		FROM inst_leave_schedule
		WHERE inst_id = ? AND leave_request_id = ? AND del_flag = 0
		ORDER BY start_time ASC, id ASC
	`, instID, leaveID)
	if err != nil {
		return err
	}
	defer scheduleRows.Close()

	type leaveCancelScheduleSnapshot struct {
		ScheduleID         int64
		RosterStatusBefore int
		EndTime            time.Time
	}

	schedules := make([]leaveCancelScheduleSnapshot, 0)
	for scheduleRows.Next() {
		var item leaveCancelScheduleSnapshot
		if err := scheduleRows.Scan(&item.ScheduleID, &item.RosterStatusBefore, &item.EndTime); err != nil {
			return err
		}
		schedules = append(schedules, item)
	}
	if err := scheduleRows.Err(); err != nil {
		return err
	}

	now := time.Now()
	for _, item := range schedules {
		if item.EndTime.Before(now) {
			continue
		}
		restoreStatus := normalizeTeachingScheduleStudentRosterStatus(item.RosterStatusBefore)
		if restoreStatus == model.TeachingScheduleStudentRosterStatusLeave {
			restoreStatus = model.TeachingScheduleStudentRosterStatusActive
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE teaching_schedule_student
			SET roster_status = ?, update_id = ?, update_time = CURRENT_TIMESTAMP
			WHERE inst_id = ?
			  AND teaching_schedule_id = ?
			  AND student_id = ?
			  AND del_flag = 0
			  AND roster_status = ?
		`, restoreStatus, operator.ID, instID, item.ScheduleID, studentID, model.TeachingScheduleStudentRosterStatusLeave); err != nil {
			return err
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE inst_leave_request
		SET
			status = ?,
			current_step = NULL,
			current_approver_ids = '',
			current_approver_names = '',
			update_id = ?,
			update_time = CURRENT_TIMESTAMP
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, model.LeaveStatusRevoked, operator.ID, leaveID, instID); err != nil {
		return err
	}

	if err := repo.insertLeaveActionTx(ctx, tx, instID, leaveID, model.LeaveActionRevoke, operator.ID, operator.Name, "已撤销", remark); err != nil {
		return err
	}

	return tx.Commit()
}

func (repo *Repository) PageLeaveRequests(ctx context.Context, instID int64, query model.LeavePagedQueryDTO) (model.LeavePagedResult, error) {
	current := query.PageRequestModel.PageIndex
	size := query.PageRequestModel.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 20
	}
	offset := (current - 1) * size

	whereParts := []string{"inst_id = ?", "del_flag = 0"}
	args := []any{instID}
	if studentID := strings.TrimSpace(query.QueryModel.StudentID.String()); studentID != "" {
		whereParts = append(whereParts, "CAST(student_id AS CHAR) = ?")
		args = append(args, studentID)
	}
	if from := parseLeaveDateFilterStart(query.QueryModel.ApplyStartTime); from != nil {
		whereParts = append(whereParts, "create_time >= ?")
		args = append(args, *from)
	}
	if to := parseLeaveDateFilterEnd(query.QueryModel.ApplyEndTime); to != nil {
		whereParts = append(whereParts, "create_time <= ?")
		args = append(args, *to)
	}
	if len(query.QueryModel.Statuses) > 0 {
		whereParts = append(whereParts, "status IN ("+sqlPlaceholders(len(query.QueryModel.Statuses))+")")
		for _, status := range query.QueryModel.Statuses {
			args = append(args, status)
		}
	}
	if len(query.QueryModel.LeaveTypes) > 0 {
		whereParts = append(whereParts, "leave_type IN ("+sqlPlaceholders(len(query.QueryModel.LeaveTypes))+")")
		for _, leaveType := range query.QueryModel.LeaveTypes {
			args = append(args, leaveType)
		}
	}

	whereSQL := strings.Join(whereParts, " AND ")
	var total int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM inst_leave_request
		WHERE `+whereSQL, args...).Scan(&total); err != nil {
		return model.LeavePagedResult{}, err
	}

	orderBy := "ORDER BY create_time DESC"
	if query.SortModel.ByApplyTime > 0 {
		orderBy = "ORDER BY create_time ASC"
	} else if query.SortModel.ByApplyTime < 0 {
		orderBy = "ORDER BY create_time DESC"
	}

	listArgs := []any{model.LeaveActionApprove, model.LeaveActionReject, model.LeaveActionAutoApprove}
	listArgs = append(listArgs, args...)
	listArgs = append(listArgs, size, offset)

	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			id,
			CAST(student_id AS CHAR),
			IFNULL(student_name, ''),
			IFNULL(student_avatar_url, ''),
			IFNULL(student_phone, ''),
			start_time,
			end_time,
			IFNULL(is_agent, 1),
			IFNULL(leave_type, 0),
			IFNULL(initiate_staff_name, ''),
			IFNULL(status, 0),
			IFNULL(current_approver_names, ''),
			create_time,
			IFNULL((
				SELECT NULLIF(action_staff_name, '')
				FROM inst_leave_action action_log
				WHERE action_log.inst_id = inst_leave_request.inst_id
				  AND action_log.leave_request_id = inst_leave_request.id
				  AND action_log.del_flag = 0
				  AND action_log.action_type IN (?, ?, ?)
				ORDER BY action_log.create_time DESC, action_log.id DESC
				LIMIT 1
			), '')
		FROM inst_leave_request
		WHERE `+whereSQL+`
		`+orderBy+`
		LIMIT ? OFFSET ?
	`, listArgs...)
	if err != nil {
		return model.LeavePagedResult{}, err
	}
	defer rows.Close()

	list := make([]model.LeavePagedItem, 0, size)
	for rows.Next() {
		var item model.LeavePagedItem
		var (
			startTime sql.NullTime
			endTime   sql.NullTime
			applyTime sql.NullTime
			isAgent   int
		)
		if err := rows.Scan(
			&item.ID,
			&item.StudentID,
			&item.StudentName,
			&item.StudentAvatarURL,
			&item.StudentPhone,
			&startTime,
			&endTime,
			&isAgent,
			&item.LeaveType,
			&item.InitiateStaffName,
			&item.Status,
			&item.CurrentApproverName,
			&applyTime,
			&item.ApproverName,
		); err != nil {
			return model.LeavePagedResult{}, err
		}
		if startTime.Valid {
			t := startTime.Time
			item.StartTime = &t
		}
		if endTime.Valid {
			t := endTime.Time
			item.EndTime = &t
		}
		if applyTime.Valid {
			t := applyTime.Time
			item.ApplyTime = &t
		}
		item.IsAgent = isAgent != 0
		item.StudentPhone = maskApprovalPhone(item.StudentPhone)
		item.LeaveTypeText = leaveTypeText(item.LeaveType)
		item.StatusText = leaveStatusText(item.Status)
		item.OperatorName = item.InitiateStaffName
		list = append(list, item)
	}
	if err := rows.Err(); err != nil {
		return model.LeavePagedResult{}, err
	}

	return model.LeavePagedResult{
		List:  list,
		Total: total,
	}, nil
}

func (repo *Repository) GetLeaveDetail(ctx context.Context, instID, leaveID int64) (model.LeaveDetailVO, error) {
	var detail model.LeaveDetailVO
	var (
		startTime          sql.NullTime
		endTime            sql.NullTime
		applyTime          sql.NullTime
		proofMaterialsJSON sql.NullString
		currentStep        sql.NullInt64
		isAgent            int
	)
	err := repo.db.QueryRowContext(ctx, `
		SELECT
			CAST(r.id AS CHAR),
			CAST(r.student_id AS CHAR),
			IFNULL(r.student_name, ''),
			IFNULL(r.student_avatar_url, ''),
			IFNULL(r.student_phone, ''),
			IFNULL(s.stu_sex, 0),
			r.start_time,
			r.end_time,
			IFNULL(r.is_agent, 1),
			IFNULL(r.leave_type, 0),
			IFNULL(r.reason, ''),
			r.proof_materials_json,
			IFNULL(r.remark, ''),
			IFNULL(r.status, 0),
			IFNULL(r.initiate_staff_name, ''),
			CAST(IFNULL(r.initiate_staff_id, 0) AS CHAR),
			IFNULL(iu.avatar, ''),
			IFNULL(r.current_approver_names, ''),
			r.create_time,
			current_step,
			IFNULL((
				SELECT NULLIF(action_staff_name, '')
				FROM inst_leave_action action_log
				WHERE action_log.inst_id = r.inst_id
				  AND action_log.leave_request_id = r.id
				  AND action_log.del_flag = 0
				  AND action_log.action_type IN (?, ?, ?)
				ORDER BY action_log.create_time DESC, action_log.id DESC
				LIMIT 1
			), '')
		FROM inst_leave_request r
		LEFT JOIN inst_student s
		  ON s.id = r.student_id
		 AND s.inst_id = r.inst_id
		 AND s.del_flag = 0
		LEFT JOIN inst_user iu
		  ON iu.id = r.initiate_staff_id
		 AND iu.inst_id = r.inst_id
		 AND iu.del_flag = 0
		WHERE r.id = ? AND r.inst_id = ? AND r.del_flag = 0
		LIMIT 1
	`, model.LeaveActionApprove, model.LeaveActionReject, model.LeaveActionAutoApprove, leaveID, instID).Scan(
		&detail.ID,
		&detail.StudentID,
		&detail.StudentName,
		&detail.StudentAvatarURL,
		&detail.StudentPhone,
		&detail.StudentSex,
		&startTime,
		&endTime,
		&isAgent,
		&detail.LeaveType,
		&detail.Reason,
		&proofMaterialsJSON,
		&detail.Remark,
		&detail.Status,
		&detail.InitiateStaffName,
		&detail.OperatorID,
		&detail.OperatorAvatar,
		&detail.CurrentApproverName,
		&applyTime,
		&currentStep,
		&detail.ApproverName,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.LeaveDetailVO{}, errors.New("未找到请假记录")
		}
		return model.LeaveDetailVO{}, err
	}
	if startTime.Valid {
		t := startTime.Time
		detail.StartTime = &t
		detail.StartDate = &t
	}
	if endTime.Valid {
		t := endTime.Time
		detail.EndTime = &t
		detail.EndDate = &t
	}
	if applyTime.Valid {
		t := applyTime.Time
		detail.ApplyTime = &t
		detail.OperationDate = &t
	}
	detail.IsAgent = isAgent != 0
	detail.StudentPhone = maskApprovalPhone(detail.StudentPhone)
	detail.LeaveTypeText = leaveTypeText(detail.LeaveType)
	detail.StatusText = leaveStatusText(detail.Status)
	if detail.OperatorID == "0" {
		detail.OperatorID = ""
	}
	detail.OperatorName = firstNonEmptyString(strings.TrimSpace(detail.InitiateStaffName), strings.TrimSpace(detail.OperatorName), "-")
	if proofMaterialsJSON.Valid && strings.TrimSpace(proofMaterialsJSON.String) != "" {
		detail.ProofMaterials = decodeJSONStringArray([]byte(proofMaterialsJSON.String))
	}

	scheduleRows, err := repo.db.QueryContext(ctx, `
		SELECT
			CAST(teaching_schedule_id AS CHAR),
			IFNULL(class_type, 0),
			CAST(teaching_class_id AS CHAR),
			IFNULL(teaching_class_name, ''),
			CAST(lesson_id AS CHAR),
			IFNULL(lesson_name, ''),
			CAST(teacher_id AS CHAR),
			IFNULL(teacher_name, ''),
			start_time,
			end_time,
			IFNULL(roster_status_before, 1)
		FROM inst_leave_schedule
		WHERE inst_id = ? AND leave_request_id = ? AND del_flag = 0
		ORDER BY start_time ASC, id ASC
	`, instID, leaveID)
	if err != nil {
		return model.LeaveDetailVO{}, err
	}
	defer scheduleRows.Close()
	detail.Schedules = make([]model.LeaveScheduleSnapshotVO, 0)
	for scheduleRows.Next() {
		var item model.LeaveScheduleSnapshotVO
		var (
			itemStart sql.NullTime
			itemEnd   sql.NullTime
		)
		if err := scheduleRows.Scan(
			&item.ScheduleID,
			&item.ClassType,
			&item.TeachingClassID,
			&item.TeachingClassName,
			&item.LessonID,
			&item.LessonName,
			&item.TeacherID,
			&item.TeacherName,
			&itemStart,
			&itemEnd,
			&item.RosterStatusBefore,
		); err != nil {
			return model.LeaveDetailVO{}, err
		}
		if itemStart.Valid {
			t := itemStart.Time
			item.StartTime = &t
		}
		if itemEnd.Valid {
			t := itemEnd.Time
			item.EndTime = &t
		}
		detail.Schedules = append(detail.Schedules, item)
	}
	if err := scheduleRows.Err(); err != nil {
		return model.LeaveDetailVO{}, err
	}

	actionRows, err := repo.db.QueryContext(ctx, `
		SELECT
			action_type,
			CAST(IFNULL(action_staff_id, 0) AS CHAR),
			IFNULL(action_staff_name, ''),
			IFNULL(iu.avatar, ''),
			IFNULL(status_text, ''),
			IFNULL(remark, ''),
			action_log.create_time
		FROM inst_leave_action action_log
		LEFT JOIN inst_user iu
		  ON iu.id = action_log.action_staff_id
		 AND iu.inst_id = action_log.inst_id
		 AND iu.del_flag = 0
		WHERE action_log.inst_id = ? AND action_log.leave_request_id = ? AND action_log.del_flag = 0
		ORDER BY action_log.create_time ASC, action_log.id ASC
	`, instID, leaveID)
	if err != nil {
		return model.LeaveDetailVO{}, err
	}
	defer actionRows.Close()
	detail.Processes = make([]model.LeaveProcessVO, 0)
	detail.Approves = make([]model.LeaveDetailApproveVO, 0)
	for actionRows.Next() {
		var (
			item       model.LeaveProcessVO
			approve    model.LeaveDetailApproveVO
			actionTime sql.NullTime
		)
		if err := actionRows.Scan(
			&item.ActionType,
			&approve.OperatorID,
			&item.Name,
			&approve.OperatorAvatar,
			&item.Status,
			&item.Remark,
			&actionTime,
		); err != nil {
			return model.LeaveDetailVO{}, err
		}
		if actionTime.Valid {
			t := actionTime.Time
			item.ActionTime = &t
		}
		if approve.OperatorID == "0" {
			approve.OperatorID = ""
		}
		approve = buildLeaveApproveVO(
			item.ActionType,
			approve.OperatorID,
			item.Name,
			approve.OperatorAvatar,
			item.ActionTime,
			item.Remark,
			item.Status,
			false,
		)
		detail.Processes = append(detail.Processes, item)
		detail.Approves = append(detail.Approves, approve)
	}
	if err := actionRows.Err(); err != nil {
		return model.LeaveDetailVO{}, err
	}
	if detail.Status == model.LeaveStatusPending && strings.TrimSpace(detail.CurrentApproverName) != "" {
		detail.Processes = append(detail.Processes, model.LeaveProcessVO{
			ActionType: model.LeaveActionApprove,
			Name:       detail.CurrentApproverName,
			Status:     "待处理",
			Pending:    true,
		})
		pendingApprove := buildLeaveApproveVO(
			model.LeaveActionApprove,
			"",
			detail.CurrentApproverName,
			"",
			nil,
			"",
			"待处理",
			true,
		)
		detail.Approve = &pendingApprove
	} else if len(detail.Approves) > 0 {
		currentApprove := detail.Approves[len(detail.Approves)-1]
		detail.Approve = &currentApprove
	}

	return detail, nil
}

func (repo *Repository) PageLeaveDetailSchedules(ctx context.Context, instID int64, query model.LeaveDetailScheduleQueryDTO) (model.LeaveDetailSchedulePagedResult, error) {
	studentID, startTime, endTime, err := parseLeaveDetailScheduleInput(query.QueryModel)
	if err != nil {
		return model.LeaveDetailSchedulePagedResult{}, err
	}

	current := query.PageRequestModel.PageIndex
	size := query.PageRequestModel.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 1000
	}

	schedules, err := repo.listLeaveApplicableSchedules(ctx, repo.db, instID, studentID, startTime, endTime, false)
	if err != nil {
		return model.LeaveDetailSchedulePagedResult{}, err
	}

	desc := query.SortModel.ByStartDate == 2
	sort.SliceStable(schedules, func(i, j int) bool {
		if desc {
			if schedules[i].StartTime.Equal(schedules[j].StartTime) {
				return schedules[i].ScheduleID > schedules[j].ScheduleID
			}
			return schedules[i].StartTime.After(schedules[j].StartTime)
		}
		if schedules[i].StartTime.Equal(schedules[j].StartTime) {
			return schedules[i].ScheduleID < schedules[j].ScheduleID
		}
		return schedules[i].StartTime.Before(schedules[j].StartTime)
	})

	total := len(schedules)
	offset := (current - 1) * size
	if offset > total {
		offset = total
	}
	endIndex := offset + size
	if endIndex > total {
		endIndex = total
	}

	studentName := repo.getLeaveStudentName(ctx, instID, studentID)
	list := make([]model.LeaveDetailScheduleItem, 0, endIndex-offset)
	for _, item := range schedules[offset:endIndex] {
		list = append(list, buildLeaveDetailScheduleItem(instID, studentID, studentName, item))
	}

	return model.LeaveDetailSchedulePagedResult{
		List:  list,
		Total: total,
	}, nil
}

func (repo *Repository) getLeaveStudentSnapshotTx(ctx context.Context, tx *sql.Tx, instID, studentID int64) (leaveStudentSnapshot, error) {
	var item leaveStudentSnapshot
	if err := tx.QueryRowContext(ctx, `
		SELECT id, IFNULL(stu_name, ''), IFNULL(avatar_url, ''), IFNULL(mobile, '')
		FROM inst_student
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, studentID, instID).Scan(&item.ID, &item.Name, &item.AvatarURL, &item.Phone); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return item, errors.New("学员不存在或已删除")
		}
		return item, err
	}
	return item, nil
}

func (repo *Repository) getLeaveOperatorSnapshotTx(ctx context.Context, tx *sql.Tx, instID, instUserID int64) (leaveOperatorSnapshot, error) {
	var item leaveOperatorSnapshot
	if err := tx.QueryRowContext(ctx, `
		SELECT id, IFNULL(nick_name, '')
		FROM inst_user
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, instUserID, instID).Scan(&item.ID, &item.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return item, errors.New("当前操作人不存在")
		}
		return item, err
	}
	item.Name = firstNonEmptyString(strings.TrimSpace(item.Name), strconv.FormatInt(item.ID, 10))
	return item, nil
}

func (repo *Repository) getLeaveStudentName(ctx context.Context, instID, studentID int64) string {
	var name string
	err := repo.db.QueryRowContext(ctx, `
		SELECT IFNULL(stu_name, '')
		FROM inst_student
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, studentID, instID).Scan(&name)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(name)
}

func buildLeaveDetailScheduleItem(instID, studentID int64, studentName string, item leaveApplicableSchedule) model.LeaveDetailScheduleItem {
	titleStudentName := firstNonEmptyString(strings.TrimSpace(studentName), "学员")
	lessonName := firstNonEmptyString(strings.TrimSpace(item.LessonName), strings.TrimSpace(item.TeachingClassName), "课程")
	lessonDay := time.Date(item.StartTime.Year(), item.StartTime.Month(), item.StartTime.Day(), 0, 0, 0, 0, item.StartTime.Location())
	start := item.StartTime
	end := item.EndTime
	scheduleID := strconv.FormatInt(item.ScheduleID, 10)
	studentIDString := strconv.FormatInt(studentID, 10)
	teacherID := emptyStringIfZero(item.TeacherID)
	teacherName := firstNonEmptyString(strings.TrimSpace(item.TeacherName), "-")

	tags := make([]string, 0, 1)
	if item.ClassType == model.TeachingClassTypeOneToOne {
		tags = append(tags, "1v1")
	}
	tagsJSON, _ := json.Marshal(tags)

	return model.LeaveDetailScheduleItem{
		OrgID:             strconv.FormatInt(instID, 10),
		SchoolID:          "",
		SchoolName:        "",
		ID:                scheduleID,
		Title:             fmt.Sprintf("%s-%s", titleStudentName, lessonName),
		LessonDay:         &lessonDay,
		LessonType:        item.ClassType,
		IsFinished:        !item.EndTime.After(time.Now()),
		StartMinutes:      item.StartTime.Hour()*60 + item.StartTime.Minute(),
		EndMinutes:        item.EndTime.Hour()*60 + item.EndTime.Minute(),
		Remark:            "",
		ExternalRemark:    "",
		LessonID:          emptyStringIfZero(item.LessonID),
		LessonName:        lessonName,
		LessonColor:       "",
		MainTeacherID:     teacherID,
		MainTeacherName:   teacherName,
		MainTeacherColor:  "",
		MainTeacherStatus: 0,
		MainTeacherAvatar: "",
		Tags:              tags,
		TagsString:        string(tagsJSON),
		SourceType:        item.ClassType,
		SourceID:          scheduleID,
		Address:           nil,
		AddressType:       0,
		AddressID:         "0",
		AddressName:       "",
		Members: []model.LeaveDetailScheduleMemberVO{
			{
				MemberType:  1,
				MemberID:    studentIDString,
				MemberName:  firstNonEmptyString(strings.TrimSpace(studentName), "-"),
				TimetableID: "0",
			},
		},
		Teachers: []model.LeaveDetailScheduleTeacherVO{
			{
				TeacherColor:  "",
				TeacherID:     teacherID,
				TeacherDuty:   1,
				TeacherName:   teacherName,
				TeacherStatus: 0,
			},
		},
		RepeatSpan:         0,
		WeekDays:           0,
		ScheduleSourceType: item.ClassType,
		ScheduleSourceID:   studentIDString,
		MaxStudentCount:    0,
		BookedStudentCount: 0,
		SubjectID:          "0",
		SubjectName:        "",
		IsOrgCreated:       false,
		IsOpenLiveRecord:   false,
		IsOpenLive:         false,
		StartTime:          &start,
		EndTime:            &end,
	}
}

func (repo *Repository) getLeaveApprovalStepTx(ctx context.Context, tx *sql.Tx, instID int64) (*int, []int64, string, int, error) {
	var (
		configID      int64
		configVersion int
	)
	err := tx.QueryRowContext(ctx, `
		SELECT id, IFNULL(config_version, 0)
		FROM inst_approval_config
		WHERE inst_id = ? AND type = ? AND del_flag = 0
		ORDER BY id DESC
		LIMIT 1
	`, instID, model.LeaveApprovalConfigType).Scan(&configID, &configVersion)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, "", 0, nil
		}
		return nil, nil, "", 0, err
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT step, IFNULL(staff_id, '')
		FROM inst_approval_flow
		WHERE config_id = ? AND config_version = ? AND del_flag = 0
		ORDER BY step ASC, id ASC
	`, configID, configVersion)
	if err != nil {
		return nil, nil, "", 0, err
	}
	defer rows.Close()

	flowByStep := make(map[int][]int64)
	steps := make([]int, 0)
	userIDs := make(map[int64]struct{})
	for rows.Next() {
		var (
			step       int
			staffIDRaw string
		)
		if err := rows.Scan(&step, &staffIDRaw); err != nil {
			return nil, nil, "", 0, err
		}
		staffIDs := splitCSV(staffIDRaw)
		if len(staffIDs) == 0 {
			continue
		}
		if _, ok := flowByStep[step]; !ok {
			steps = append(steps, step)
		}
		flowByStep[step] = append(flowByStep[step], staffIDs...)
		for _, staffID := range staffIDs {
			userIDs[staffID] = struct{}{}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, nil, "", 0, err
	}
	if len(steps) == 0 {
		return nil, nil, "", configVersion, nil
	}

	userNames, _, err := repo.getApprovalUsers(ctx, userIDs)
	if err != nil {
		return nil, nil, "", 0, err
	}
	currentStepValue := steps[0]
	currentApproverIDs := dedupeInt64s(flowByStep[currentStepValue])
	return &currentStepValue, currentApproverIDs, joinApprovalUserNames(currentApproverIDs, userNames), configVersion, nil
}

func (repo *Repository) insertLeaveActionTx(ctx context.Context, tx *sql.Tx, instID, leaveID int64, actionType int, staffID int64, staffName, statusText, remark string) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO inst_leave_action (
			inst_id, leave_request_id, action_type, action_staff_id, action_staff_name,
			status_text, remark, create_id, update_id, del_flag
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
	`, instID, leaveID, actionType, staffID, strings.TrimSpace(staffName), strings.TrimSpace(statusText), strings.TrimSpace(remark), staffID, staffID)
	return err
}

func (repo *Repository) listLeaveApplicableSchedules(ctx context.Context, queryer sqlQueryer, instID, studentID int64, startTime, endTime time.Time, excludeLeave bool) ([]leaveApplicableSchedule, error) {
	if studentID <= 0 {
		return nil, errors.New("请选择学员")
	}
	if endTime.Before(startTime) {
		return nil, errors.New("结束时间不能早于开始时间")
	}

	args := []any{
		studentID,
		model.TeachingScheduleStudentRosterStatusActive,
		model.TeachingClassTypeNormal,
		studentID,
		model.TeachingClassStudentStatusStudying,
		studentID,
		instID,
		model.TeachingScheduleStatusActive,
		endTime,
		startTime,
		model.TeachingScheduleStudentRosterStatusRemoved,
	}
	query := `
		SELECT
			ts.id,
			IFNULL(ts.class_type, 0),
			IFNULL(ts.teaching_class_id, 0),
			IFNULL(ts.teaching_class_name, ''),
			IFNULL(ts.lesson_id, 0),
			IFNULL(ts.lesson_name, ''),
			IFNULL(ts.teacher_id, 0),
			IFNULL(ts.teacher_name, ''),
			ts.lesson_start_at,
			ts.lesson_end_at,
			CASE
				WHEN tss.student_id IS NOT NULL THEN IFNULL(tss.roster_status, ?)
				ELSE ?
			END AS roster_status_before
		FROM teaching_schedule ts
		LEFT JOIN teaching_schedule_student tss
		  ON tss.inst_id = ts.inst_id
		 AND tss.teaching_schedule_id = ts.id
		 AND tss.student_id = ?
		 AND tss.del_flag = 0
		WHERE ts.inst_id = ?
		  AND ts.del_flag = 0
		  AND ts.status = ?
		  AND ts.lesson_start_at <= ?
		  AND ts.lesson_end_at >= ?
		  AND (
				(ts.student_id = ?)
				OR (
					ts.class_type = ?
					AND EXISTS (
						SELECT 1
						FROM teaching_class_student tcs
						WHERE tcs.inst_id = ts.inst_id
						  AND tcs.teaching_class_id = ts.teaching_class_id
						  AND tcs.del_flag = 0
						  AND tcs.student_id = ?
						  AND tcs.create_time < GREATEST(ts.lesson_start_at, ts.create_time)
						  AND (
							IFNULL(tcs.class_student_status, ?) = ?
							OR tcs.update_time > GREATEST(ts.lesson_start_at, ts.create_time)
						  )
					)
				)
		  )
		  AND IFNULL(tss.roster_status, ?) <> ?
	`
	args = []any{
		model.TeachingScheduleStudentRosterStatusActive,
		model.TeachingScheduleStudentRosterStatusActive,
		studentID,
		instID,
		model.TeachingScheduleStatusActive,
		endTime,
		startTime,
		studentID,
		model.TeachingClassTypeNormal,
		studentID,
		model.TeachingClassStudentStatusStudying,
		model.TeachingClassStudentStatusStudying,
		model.TeachingScheduleStudentRosterStatusActive,
		model.TeachingScheduleStudentRosterStatusRemoved,
	}
	if excludeLeave {
		query += ` AND IFNULL(tss.roster_status, ?) <> ? `
		args = append(args, model.TeachingScheduleStudentRosterStatusActive, model.TeachingScheduleStudentRosterStatusLeave)
	}
	query += ` ORDER BY ts.lesson_start_at ASC, ts.id ASC `

	rows, err := queryer.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]leaveApplicableSchedule, 0)
	for rows.Next() {
		var item leaveApplicableSchedule
		if err := rows.Scan(
			&item.ScheduleID,
			&item.ClassType,
			&item.TeachingClassID,
			&item.TeachingClassName,
			&item.LessonID,
			&item.LessonName,
			&item.TeacherID,
			&item.TeacherName,
			&item.StartTime,
			&item.EndTime,
			&item.RosterStatusBefore,
		); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func parseLeavePreviewInput(dto model.LeavePreviewDTO) (int64, time.Time, time.Time, error) {
	studentID, err := parseOptionalPositiveID(dto.StudentID.String())
	if err != nil || studentID <= 0 {
		return 0, time.Time{}, time.Time{}, errors.New("请选择学员")
	}
	startTime, err := parseLeaveDateTime(dto.StartTime)
	if err != nil {
		return 0, time.Time{}, time.Time{}, errors.New("开始时间格式不正确")
	}
	endTime, err := parseLeaveDateTime(dto.EndTime)
	if err != nil {
		return 0, time.Time{}, time.Time{}, errors.New("结束时间格式不正确")
	}
	if endTime.Before(startTime) {
		return 0, time.Time{}, time.Time{}, errors.New("结束时间不能早于开始时间")
	}
	return studentID, startTime, endTime, nil
}

func parseLeaveCreateInput(dto model.LeaveCreateDTO) (int64, time.Time, time.Time, error) {
	studentID, startTime, endTime, err := parseLeavePreviewInput(model.LeavePreviewDTO{
		StudentID: dto.StudentID,
		StartTime: dto.StartTime,
		EndTime:   dto.EndTime,
	})
	if err != nil {
		return 0, time.Time{}, time.Time{}, err
	}
	return studentID, startTime, endTime, nil
}

func parseLeaveCancelInput(dto model.LeaveCancelDTO) (int64, string, error) {
	leaveID, err := parseOptionalPositiveID(dto.ID.String())
	if err != nil || leaveID <= 0 {
		return 0, "", errors.New("id不能为空")
	}
	remark := strings.TrimSpace(dto.Remark)
	if utf8.RuneCountInString(remark) > 200 {
		return 0, "", errors.New("备注不能超过200字")
	}
	return leaveID, remark, nil
}

func parseLeaveDetailScheduleInput(query model.LeaveDetailScheduleQueryModel) (int64, time.Time, time.Time, error) {
	studentID, err := parseOptionalPositiveID(query.StudentID.String())
	if err != nil || studentID <= 0 {
		return 0, time.Time{}, time.Time{}, errors.New("请选择学员")
	}
	startTime, err := parseLeaveDateTime(query.StartDateTime)
	if err != nil {
		return 0, time.Time{}, time.Time{}, errors.New("开始时间格式不正确")
	}
	endTime, err := parseLeaveDateTime(query.EndDateTime)
	if err != nil {
		return 0, time.Time{}, time.Time{}, errors.New("结束时间格式不正确")
	}
	if endTime.Before(startTime) {
		return 0, time.Time{}, time.Time{}, errors.New("结束时间不能早于开始时间")
	}
	return studentID, startTime, endTime, nil
}

func parseLeaveDateTime(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, errors.New("empty leave datetime")
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
	}
	for _, layout := range layouts {
		if value, err := time.ParseInLocation(layout, raw, time.Local); err == nil {
			return value, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported leave datetime: %s", raw)
}

func parseLeaveDateFilterStart(raw string) *time.Time {
	parsed, err := parseLeaveDateTime(raw)
	if err == nil {
		return &parsed
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	if parsed, err := time.ParseInLocation("2006-01-02", raw, time.Local); err == nil {
		value := parsed
		return &value
	}
	return nil
}

func parseLeaveDateFilterEnd(raw string) *time.Time {
	if value := parseLeaveDateFilterStart(raw); value != nil {
		endOfDay := value.Add(24*time.Hour - time.Nanosecond)
		return &endOfDay
	}
	return nil
}

func dedupeInt64s(values []int64) []int64 {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[int64]struct{}, len(values))
	result := make([]int64, 0, len(values))
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

func leaveTypeText(value int) string {
	switch value {
	case model.LeaveTypePersonal:
		return "事假"
	case model.LeaveTypeSick:
		return "病假"
	case model.LeaveTypeSuspend:
		return "休学"
	default:
		return "-"
	}
}

func leaveStatusText(value int) string {
	switch value {
	case model.LeaveStatusPending:
		return "待处理"
	case model.LeaveStatusApproved:
		return "已通过"
	case model.LeaveStatusRejected:
		return "已拒绝"
	case model.LeaveStatusRevoked:
		return "已撤销"
	default:
		return "-"
	}
}

func leaveApproveStatusValue(actionType int, pending bool) int {
	if pending {
		return 1
	}
	switch actionType {
	case model.LeaveActionCreate:
		return 1
	case model.LeaveActionReject:
		return 2
	case model.LeaveActionApprove, model.LeaveActionAutoApprove:
		return 3
	case model.LeaveActionRevoke:
		return 4
	default:
		return 0
	}
}

func leaveApproveStatusText(actionType int, pending bool) string {
	if pending {
		return "待处理"
	}
	switch actionType {
	case model.LeaveActionCreate:
		return "发起"
	case model.LeaveActionReject:
		return "已拒绝"
	case model.LeaveActionApprove, model.LeaveActionAutoApprove:
		return "已通过"
	case model.LeaveActionRevoke:
		return "已撤销"
	default:
		return "-"
	}
}

func buildLeaveApproveVO(actionType int, operatorID, operatorName, operatorAvatar string, operationDate *time.Time, remark, rawStatus string, pending bool) model.LeaveDetailApproveVO {
	statusText := firstNonEmptyString(strings.TrimSpace(rawStatus), leaveApproveStatusText(actionType, pending))
	return model.LeaveDetailApproveVO{
		OperatorID:        strings.TrimSpace(operatorID),
		OperatorName:      firstNonEmptyString(strings.TrimSpace(operatorName), "-"),
		OperatorAvatar:    strings.TrimSpace(operatorAvatar),
		OperationDate:     operationDate,
		Remark:            strings.TrimSpace(remark),
		ApproveStatus:     leaveApproveStatusValue(actionType, pending),
		ApproveStatusText: statusText,
		ActionType:        actionType,
	}
}
