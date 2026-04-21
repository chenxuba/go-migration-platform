package repository

import (
	"context"
	"database/sql"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

func (repo *Repository) PageParentLeaveRequestsByStudentIDs(ctx context.Context, instID int64, studentIDs []int64, pageIndex, pageSize int) (model.LeavePagedResult, error) {
	normalizedStudentIDs := make([]int64, 0, len(studentIDs))
	seen := make(map[int64]struct{}, len(studentIDs))
	for _, studentID := range studentIDs {
		if studentID <= 0 {
			continue
		}
		if _, exists := seen[studentID]; exists {
			continue
		}
		seen[studentID] = struct{}{}
		normalizedStudentIDs = append(normalizedStudentIDs, studentID)
	}
	if instID <= 0 || len(normalizedStudentIDs) == 0 {
		return model.LeavePagedResult{
			List:  []model.LeavePagedItem{},
			Total: 0,
		}, nil
	}

	current := pageIndex
	if current <= 0 {
		current = 1
	}
	size := pageSize
	if size <= 0 {
		size = 20
	}
	offset := (current - 1) * size

	whereSQL := strings.Join([]string{
		"inst_id = ?",
		"del_flag = 0",
		"student_id IN (" + sqlPlaceholders(len(normalizedStudentIDs)) + ")",
	}, " AND ")
	args := make([]any, 0, len(normalizedStudentIDs)+1)
	args = append(args, instID)
	args = append(args, int64SliceToAny(normalizedStudentIDs)...)

	var total int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM inst_leave_request
		WHERE `+whereSQL, args...).Scan(&total); err != nil {
		return model.LeavePagedResult{}, err
	}

	queryArgs := make([]any, 0, len(args)+5)
	queryArgs = append(queryArgs, model.LeaveActionApprove, model.LeaveActionReject, model.LeaveActionAutoApprove)
	queryArgs = append(queryArgs, args...)
	queryArgs = append(queryArgs, size, offset)

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
		ORDER BY create_time DESC, id DESC
		LIMIT ? OFFSET ?
	`, queryArgs...)
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
