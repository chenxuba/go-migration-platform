package repository

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

func (repo *Repository) PageApprovalConfigs(ctx context.Context, instID int64, query model.ApprovalConfigPageQueryDTO) (model.ApprovalConfigPageResult, error) {
	current := query.PageRequestModel.PageIndex
	size := query.PageRequestModel.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	whereParts := []string{"r.inst_id = ?", "r.del_flag = 0"}
	args := []any{instID}
	q := query.QueryModel
	if strings.TrimSpace(q.ApprovalNumber) != "" {
		whereParts = append(whereParts, "r.approval_number LIKE ?")
		args = append(args, "%"+strings.TrimSpace(q.ApprovalNumber)+"%")
	}
	if q.ApplicantID != nil {
		whereParts = append(whereParts, "r.applicant = ?")
		args = append(args, *q.ApplicantID)
	}
	if strings.TrimSpace(q.OrderNumber) != "" {
		whereParts = append(whereParts, "o.order_number LIKE ?")
		args = append(args, "%"+strings.TrimSpace(q.OrderNumber)+"%")
	}
	if q.CurrentApproverID != nil {
		whereParts = append(whereParts, "FIND_IN_SET(?, IFNULL(r.current_approver, '')) > 0")
		args = append(args, strconv.FormatInt(*q.CurrentApproverID, 10))
	}
	if from := parseDateStart(q.FinishStartTime); from != nil {
		whereParts = append(whereParts, "r.finish_time >= ?")
		args = append(args, *from)
	}
	if to := parseDateEnd(q.FinishEndTime); to != nil {
		whereParts = append(whereParts, "r.finish_time <= ?")
		args = append(args, *to)
	}
	if from := parseDateStart(q.ApplicationStartTime); from != nil {
		whereParts = append(whereParts, "r.approval_time >= ?")
		args = append(args, *from)
	}
	if to := parseDateEnd(q.ApplicationEndTime); to != nil {
		whereParts = append(whereParts, "r.approval_time <= ?")
		args = append(args, *to)
	}
	if q.StudentID != nil {
		whereParts = append(whereParts, "r.student_id = ?")
		args = append(args, *q.StudentID)
	}
	if len(q.Statuses) > 0 {
		holders := make([]string, 0, len(q.Statuses))
		for _, item := range q.Statuses {
			holders = append(holders, "?")
			args = append(args, item)
		}
		whereParts = append(whereParts, "r.approval_status IN ("+strings.Join(holders, ",")+")")
	}

	whereSQL := strings.Join(whereParts, " AND ")
	baseFrom := `
		FROM approval_record r
		LEFT JOIN sale_order o ON o.id = r.order_id AND o.del_flag = 0
		LEFT JOIN inst_student s ON s.id = r.student_id AND s.del_flag = 0
		LEFT JOIN inst_user applicant ON applicant.id = r.applicant
		WHERE ` + whereSQL

	var total int
	if err := repo.db.QueryRowContext(ctx, `SELECT COUNT(*) `+baseFrom, args...).Scan(&total); err != nil {
		return model.ApprovalConfigPageResult{}, err
	}

	orderBy := " ORDER BY r.create_time DESC"
	if query.SortModel.ByInitiateTime != 0 {
		orderBy = fmt.Sprintf(" ORDER BY r.create_time %s", sortDirection(query.SortModel.ByInitiateTime))
	} else if query.SortModel.ByFinishTime != 0 {
		orderBy = fmt.Sprintf(" ORDER BY r.finish_time %s", sortDirection(query.SortModel.ByFinishTime))
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT r.id, IFNULL(r.approval_number, ''), IFNULL(r.approval_type, 0), IFNULL(r.current_approver, ''),
		       IFNULL(r.config_version, 0), r.current_step, IFNULL(applicant.nick_name, ''), IFNULL(s.stu_name, ''),
		       r.student_id, IFNULL(s.avatar_url, ''), IFNULL(s.mobile, ''), r.approval_time, r.finish_time,
		       r.approval_status, IFNULL(o.order_number, ''), r.order_id, o.order_type
		`+baseFrom+orderBy+`
		LIMIT ? OFFSET ?`, append(args, size, offset)...)
	if err != nil {
		return model.ApprovalConfigPageResult{}, err
	}
	defer rows.Close()

	records := make([]model.ApprovalConfigRecord, 0, size)
	recordIDs := make([]int64, 0, size)
	userIDs := make(map[int64]struct{})
	for rows.Next() {
		var record model.ApprovalConfigRecord
		var (
			currentStep    sql.NullInt64
			studentID      sql.NullInt64
			approvalTime   sql.NullTime
			finishTime     sql.NullTime
			approvalStatus sql.NullInt64
			orderID        sql.NullInt64
			orderType      sql.NullInt64
		)
		if err := rows.Scan(
			&record.ID,
			&record.ApprovalNumber,
			&record.ApprovalType,
			&record.CurrentApprover,
			&record.ConfigVersion,
			&currentStep,
			&record.ApplicantName,
			&record.StudentName,
			&studentID,
			&record.StudentAvatar,
			&record.Mobile,
			&approvalTime,
			&finishTime,
			&approvalStatus,
			&record.OrderNumber,
			&orderID,
			&orderType,
		); err != nil {
			return model.ApprovalConfigPageResult{}, err
		}
		if currentStep.Valid {
			value := int(currentStep.Int64)
			record.CurrentStep = &value
		}
		if studentID.Valid {
			record.StudentID = strconv.FormatInt(studentID.Int64, 10)
		}
		if approvalTime.Valid {
			t := approvalTime.Time
			record.ApprovalTime = &t
		}
		if finishTime.Valid {
			t := finishTime.Time
			record.FinishTime = &t
		}
		if approvalStatus.Valid {
			value := int(approvalStatus.Int64)
			record.ApprovalStatus = &value
		}
		if orderID.Valid {
			record.OrderID = strconv.FormatInt(orderID.Int64, 10)
		}
		if orderType.Valid {
			value := int(orderType.Int64)
			record.OrderType = &value
		}
		record.Mobile = maskApprovalPhone(record.Mobile)
		for _, uid := range splitCSV(record.CurrentApprover) {
			userIDs[uid] = struct{}{}
		}
		recordIDs = append(recordIDs, record.ID)
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return model.ApprovalConfigPageResult{}, err
	}

	userNames, userDisabled, err := repo.getApprovalUsers(ctx, userIDs)
	if err != nil {
		return model.ApprovalConfigPageResult{}, err
	}
	configsByType, err := repo.getApprovalConfigsByType(ctx, instID)
	if err != nil {
		return model.ApprovalConfigPageResult{}, err
	}
	flowsByConfig, err := repo.getApprovalFlowsByConfig(ctx, configsByType)
	if err != nil {
		return model.ApprovalConfigPageResult{}, err
	}
	historyByApproval, err := repo.getApprovalHistories(ctx, recordIDs)
	if err != nil {
		return model.ApprovalConfigPageResult{}, err
	}

	for idx := range records {
		record := &records[idx]
		record.CurrentApprover = joinApprovalUserNames(splitCSV(record.CurrentApprover), userNames)
		config, ok := configsByType[record.ApprovalType]
		if !ok {
			continue
		}
		flows := flowsByConfig[config.ID]
		histories := historyByApproval[record.ID]
		record.ApproveFlows = buildApprovalFlowStages(flows, histories, record.CurrentStep, userNames, userDisabled)
	}

	return model.ApprovalConfigPageResult{
		Records: records,
		Total:   total,
		Current: current,
		Size:    size,
	}, nil
}

func (repo *Repository) SaveApprovalConfig(ctx context.Context, instID, operatorID int64, dto model.ApprovalConfigSaveDTO) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var configVersion int
	err = tx.QueryRowContext(ctx, `
		SELECT IFNULL(config_version, 0)
		FROM inst_approval_config
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, dto.ID, instID).Scan(&configVersion)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE inst_approval_flow
		SET del_flag = 1, update_id = ?, update_time = NOW()
		WHERE config_id = ? AND config_version = ? AND del_flag = 0
	`, operatorID, dto.ID, configVersion); err != nil {
		return err
	}

	nextVersion := configVersion + 1
	if _, err := tx.ExecContext(ctx, `
		UPDATE inst_approval_config
		SET enable = ?, rule_json = ?, config_version = ?, update_id = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, boolValue(dto.Enable), strings.TrimSpace(dto.RuleJSON), nextVersion, operatorID, dto.ID, instID); err != nil {
		return err
	}

	for _, flow := range dto.StaffFlowList {
		if flow.Step <= 0 || len(flow.StaffIDs) == 0 {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO inst_approval_flow (
				uuid, version, config_id, config_version, staff_id, step,
				create_id, create_time, update_id, update_time, del_flag
			) VALUES (
				UUID(), 0, ?, ?, ?, ?, ?, NOW(), ?, NOW(), 0
			)
		`, dto.ID, nextVersion, joinInt64CSV(flow.StaffIDs), flow.Step, operatorID, operatorID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (repo *Repository) ApproveApprovalRecord(ctx context.Context, instID, operatorID int64, dto model.ApprovalOperateDTO) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var (
		orderID         int64
		studentID       int64
		approvalType    int
		approvalStatus  int
		configVersion   int
		currentApprover string
		currentStep     sql.NullInt64
	)
	err = tx.QueryRowContext(ctx, `
		SELECT order_id, student_id, IFNULL(approval_type, 0), IFNULL(approval_status, 0),
		       IFNULL(config_version, 0), IFNULL(current_approver, ''), current_step
		FROM approval_record
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, dto.ID, instID).Scan(&orderID, &studentID, &approvalType, &approvalStatus, &configVersion, &currentApprover, &currentStep)
	if err != nil {
		return err
	}
	if approvalStatus != 0 {
		return fmt.Errorf("当前审批状态不可处理")
	}
	allowed := false
	for _, approverID := range splitCSV(currentApprover) {
		if approverID == operatorID {
			allowed = true
			break
		}
	}
	if !allowed {
		return fmt.Errorf("当前用户不是审批人")
	}
	if !currentStep.Valid {
		return fmt.Errorf("当前审批步骤不存在")
	}

	now := time.Now()
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO approval_history (
			uuid, version, approval_id, step, approval_person, approval_time, approval_status,
			create_id, create_time, update_id, update_time, del_flag, remark
		) VALUES (
			UUID(), 0, ?, ?, ?, ?, 1, ?, ?, ?, ?, 0, ?
		)
	`, dto.ID, currentStep.Int64, operatorID, now, operatorID, now, operatorID, now, strings.TrimSpace(dto.Remark)); err != nil {
		return err
	}

	var configID int64
	err = tx.QueryRowContext(ctx, `
		SELECT id
		FROM inst_approval_config
		WHERE inst_id = ? AND type = ? AND del_flag = 0
		ORDER BY id DESC
		LIMIT 1
	`, instID, approvalType).Scan(&configID)
	if err != nil {
		return err
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT step, IFNULL(staff_id, '')
		FROM inst_approval_flow
		WHERE config_id = ? AND config_version = ? AND del_flag = 0 AND step > ?
		ORDER BY step ASC, id ASC
	`, configID, configVersion, currentStep.Int64)
	if err != nil {
		return err
	}
	defer rows.Close()
	type nextFlow struct {
		Step    int
		StaffID string
	}
	var next *nextFlow
	for rows.Next() {
		var flow nextFlow
		if err := rows.Scan(&flow.Step, &flow.StaffID); err != nil {
			return err
		}
		if strings.TrimSpace(flow.StaffID) == "" {
			continue
		}
		next = &flow
		break
	}
	if err := rows.Err(); err != nil {
		return err
	}

	if next != nil {
		if _, err := tx.ExecContext(ctx, `
			UPDATE approval_record
			SET current_step = ?, current_approver = ?, update_id = ?, update_time = NOW()
			WHERE id = ? AND inst_id = ? AND del_flag = 0
		`, next.Step, next.StaffID, operatorID, dto.ID, instID); err != nil {
			return err
		}
		return tx.Commit()
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE approval_record
		SET approval_status = 1, finish_time = ?, update_id = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, now, operatorID, dto.ID, instID); err != nil {
		return err
	}
	if approvalType == 1 {
		if err := repo.completeOrderRegistrationTx(ctx, tx, instID, operatorID, orderID, studentID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE sale_order
			SET order_status = 3, update_id = ?, update_time = NOW()
			WHERE id = ? AND inst_id = ? AND del_flag = 0
		`, operatorID, orderID, instID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (repo *Repository) CancelApprovalRecord(ctx context.Context, instID, operatorID int64, dto model.ApprovalOperateDTO) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var (
		orderID         int64
		approvalType    int
		approvalStatus  int
		applicant       int64
		currentApprover string
		currentStep     sql.NullInt64
	)
	err = tx.QueryRowContext(ctx, `
		SELECT order_id, IFNULL(approval_type, 0), IFNULL(approval_status, 0),
		       IFNULL(applicant, 0), IFNULL(current_approver, ''), current_step
		FROM approval_record
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, dto.ID, instID).Scan(&orderID, &approvalType, &approvalStatus, &applicant, &currentApprover, &currentStep)
	if err != nil {
		return err
	}
	if approvalStatus != 0 {
		return fmt.Errorf("当前审批状态不可处理")
	}

	allowed := operatorID == applicant
	if !allowed {
		for _, approverID := range splitCSV(currentApprover) {
			if approverID == operatorID {
				allowed = true
				break
			}
		}
	}
	if !allowed {
		return fmt.Errorf("当前用户无权作废审批")
	}
	if !currentStep.Valid {
		return fmt.Errorf("当前审批步骤不存在")
	}

	now := time.Now()
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO approval_history (
			uuid, version, approval_id, step, approval_person, approval_time, approval_status,
			create_id, create_time, update_id, update_time, del_flag, remark
		) VALUES (
			UUID(), 0, ?, ?, ?, ?, 2, ?, ?, ?, ?, 0, ?
		)
	`, dto.ID, currentStep.Int64, operatorID, now, operatorID, now, operatorID, now, strings.TrimSpace(dto.Remark)); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE approval_record
		SET approval_status = 2, finish_time = ?, update_id = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, now, operatorID, dto.ID, instID); err != nil {
		return err
	}
	if approvalType == 1 {
		if _, err := tx.ExecContext(ctx, `
			UPDATE sale_order
			SET order_status = 5, update_id = ?, update_time = NOW()
			WHERE id = ? AND inst_id = ? AND del_flag = 0
		`, operatorID, orderID, instID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

type approvalConfigMeta struct {
	ID            int64
	Type          int
	ConfigVersion int
}

type approvalFlowMeta struct {
	Step     int
	StaffIDs []int64
}

type approvalHistoryMeta struct {
	Step           int
	ApprovalPerson int64
	ApprovalStatus *int
	ApprovalTime   *time.Time
	Remark         string
}

func (repo *Repository) getApprovalUsers(ctx context.Context, userIDs map[int64]struct{}) (map[int64]string, map[int64]bool, error) {
	names := make(map[int64]string)
	disabled := make(map[int64]bool)
	if len(userIDs) == 0 {
		return names, disabled, nil
	}
	ids := make([]int64, 0, len(userIDs))
	for id := range userIDs {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	holders := strings.TrimRight(strings.Repeat("?,", len(ids)), ",")
	args := make([]any, 0, len(ids))
	for _, id := range ids {
		args = append(args, id)
	}
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, IFNULL(nick_name, ''), IFNULL(disabled, 0)
		FROM inst_user
		WHERE id IN (`+holders+`)
	`, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var name string
		var isDisabled bool
		if err := rows.Scan(&id, &name, &isDisabled); err != nil {
			return nil, nil, err
		}
		names[id] = name
		disabled[id] = isDisabled
	}
	return names, disabled, rows.Err()
}

func (repo *Repository) getApprovalConfigsByType(ctx context.Context, instID int64) (map[int]approvalConfigMeta, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, IFNULL(type, 0), IFNULL(config_version, 0)
		FROM inst_approval_config
		WHERE inst_id = ? AND del_flag = 0
	`, instID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[int]approvalConfigMeta)
	for rows.Next() {
		var item approvalConfigMeta
		if err := rows.Scan(&item.ID, &item.Type, &item.ConfigVersion); err != nil {
			return nil, err
		}
		result[item.Type] = item
	}
	return result, rows.Err()
}

func (repo *Repository) getApprovalFlowsByConfig(ctx context.Context, configsByType map[int]approvalConfigMeta) (map[int64][]approvalFlowMeta, error) {
	result := make(map[int64][]approvalFlowMeta)
	for _, config := range configsByType {
		rows, err := repo.db.QueryContext(ctx, `
			SELECT step, IFNULL(staff_id, '')
			FROM inst_approval_flow
			WHERE config_id = ? AND config_version = ? AND del_flag = 0
			ORDER BY step ASC, id ASC
		`, config.ID, config.ConfigVersion)
		if err != nil {
			return nil, err
		}
		flows := make([]approvalFlowMeta, 0, 4)
		for rows.Next() {
			var step int
			var staffID string
			if err := rows.Scan(&step, &staffID); err != nil {
				rows.Close()
				return nil, err
			}
			flows = append(flows, approvalFlowMeta{Step: step, StaffIDs: splitCSV(staffID)})
		}
		rows.Close()
		sort.Slice(flows, func(i, j int) bool { return flows[i].Step < flows[j].Step })
		result[config.ID] = flows
	}
	return result, nil
}

func (repo *Repository) getApprovalHistories(ctx context.Context, approvalIDs []int64) (map[int64][]approvalHistoryMeta, error) {
	result := make(map[int64][]approvalHistoryMeta)
	if len(approvalIDs) == 0 {
		return result, nil
	}
	holders := strings.TrimRight(strings.Repeat("?,", len(approvalIDs)), ",")
	args := make([]any, 0, len(approvalIDs))
	for _, id := range approvalIDs {
		args = append(args, id)
	}
	rows, err := repo.db.QueryContext(ctx, `
		SELECT approval_id, step, approval_person, approval_status, approval_time, IFNULL(remark, '')
		FROM approval_history
		WHERE del_flag = 0 AND approval_id IN (`+holders+`)
		ORDER BY step ASC, id ASC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			approvalID int64
			item       approvalHistoryMeta
			status     sql.NullInt64
			operateAt  sql.NullTime
		)
		if err := rows.Scan(&approvalID, &item.Step, &item.ApprovalPerson, &status, &operateAt, &item.Remark); err != nil {
			return nil, err
		}
		if status.Valid {
			value := int(status.Int64)
			item.ApprovalStatus = &value
		}
		if operateAt.Valid {
			t := operateAt.Time
			item.ApprovalTime = &t
		}
		result[approvalID] = append(result[approvalID], item)
	}
	return result, rows.Err()
}

func buildApprovalFlowStages(flows []approvalFlowMeta, histories []approvalHistoryMeta, currentStep *int, userNames map[int64]string, userDisabled map[int64]bool) []model.ApprovalFlowStageVO {
	if len(flows) == 0 {
		return nil
	}
	historyByStep := make(map[int]approvalHistoryMeta, len(histories))
	for _, history := range histories {
		historyByStep[history.Step] = history
	}
	stages := make([]model.ApprovalFlowStageVO, 0, len(flows))
	for _, flow := range flows {
		stage := model.ApprovalFlowStageVO{
			Step:       flow.Step,
			FlowStaffs: make([]model.ApprovalFlowStaffVO, 0, len(flow.StaffIDs)),
		}
		if currentStep != nil && flow.Step == *currentStep {
			stage.IsCurrentStage = true
		}
		if history, ok := historyByStep[flow.Step]; ok {
			stage.Status = history.ApprovalStatus
			stage.OperateTime = history.ApprovalTime
			stage.Remark = history.Remark
		}
		for _, staffID := range flow.StaffIDs {
			staff := model.ApprovalFlowStaffVO{
				StaffID:       strconv.FormatInt(staffID, 10),
				StaffName:     userNames[staffID],
				TeacherStatus: 1,
			}
			if userDisabled[staffID] {
				staff.TeacherStatus = 2
			}
			if history, ok := historyByStep[flow.Step]; ok && history.ApprovalPerson == staffID {
				staff.IsApproveOperate = true
			}
			stage.FlowStaffs = append(stage.FlowStaffs, staff)
		}
		stages = append(stages, stage)
	}
	return stages
}

func joinApprovalUserNames(userIDs []int64, userNames map[int64]string) string {
	names := make([]string, 0, len(userIDs))
	for _, userID := range userIDs {
		if name := strings.TrimSpace(userNames[userID]); name != "" {
			names = append(names, name)
		}
	}
	return strings.Join(names, ",")
}

func sortDirection(flag int) string {
	if flag < 0 {
		return "ASC"
	}
	return "DESC"
}

func maskApprovalPhone(phone string) string {
	phone = strings.TrimSpace(phone)
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}
