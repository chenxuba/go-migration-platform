package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

var approvalTemplateTypeNames = map[int]string{
	1: "报名续费",
	2: "转课",
	3: "退课",
	4: "储值充值",
	5: "储值退费",
	6: "退学杂教材费",
}

var staffSummaryColors = []string{
	"#0098BE",
	"#009C66",
	"#4E6DFF",
	"#1FC0BE",
	"#DDBA00",
	"#6E93FF",
	"#FF6767",
	"#00C785",
	"#EF4AA9",
	"#97B527",
	"#00BAF2",
	"#FFAF00",
	"#CA6CF8",
	"#00C350",
	"#0A86FF",
	"#FC6B9C",
	"#C77B2B",
}

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

func (repo *Repository) ListApprovalTemplates(ctx context.Context, instID int64) ([]model.ApprovalTemplateVO, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT c.id, c.type, IFNULL(c.name, ''), IFNULL(c.enable, 0), IFNULL(c.rule_json, ''), IFNULL(c.config_version, 0),
		       c.update_id, c.update_time, IFNULL(u.nick_name, '')
		FROM inst_approval_config c
		LEFT JOIN inst_user u ON u.id = c.update_id
		WHERE c.del_flag = 0 AND c.inst_id = ?
		ORDER BY c.type ASC, c.id ASC
	`, instID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type configRow struct {
		model.ApprovalTemplateVO
		configVersion int
	}
	configs := make(map[int]configRow)
	configIDs := make([]int64, 0, 8)
	for rows.Next() {
		var (
			item        configRow
			id          int64
			updatedByID sql.NullInt64
			updatedAt   sql.NullTime
		)
		if err := rows.Scan(&id, &item.Type, &item.Name, &item.Enable, &item.RuleJSON, &item.configVersion, &updatedByID, &updatedAt, &item.UpdatedStaffName); err != nil {
			return nil, err
		}
		item.ID = strconv.FormatInt(id, 10)
		if strings.TrimSpace(item.Name) == "" {
			item.Name = approvalTemplateTypeNames[item.Type]
		}
		if updatedAt.Valid {
			t := updatedAt.Time
			item.UpdatedTime = &t
		}
		configs[item.Type] = item
		configIDs = append(configIDs, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	flowMap, err := repo.getApprovalTemplateFlows(ctx, configIDs)
	if err != nil {
		return nil, err
	}

	result := make([]model.ApprovalTemplateVO, 0, len(approvalTemplateTypeNames))
	for typeID := 1; typeID <= 6; typeID++ {
		if cfg, ok := configs[typeID]; ok {
			configID, _ := strconv.ParseInt(cfg.ID, 10, 64)
			cfg.FlowModels = flowMap[configID]
			result = append(result, cfg.ApprovalTemplateVO)
			continue
		}
		result = append(result, model.ApprovalTemplateVO{
			ID:         "0",
			Type:       typeID,
			Name:       approvalTemplateTypeNames[typeID],
			Enable:     false,
			RuleJSON:   "",
			FlowModels: []model.ApprovalTemplateFlowVO{},
		})
	}
	return result, nil
}

func (repo *Repository) GetApprovalDetail(ctx context.Context, instID, approvalID int64) (model.ApprovalDetailVO, error) {
	var (
		record         model.ApprovalDetailVO
		configVersion  int
		currentStep    sql.NullInt64
		currentRaw     string
		initiateTime   sql.NullTime
		finishTime     sql.NullTime
		status         sql.NullInt64
		initiateReason string
		ruleJSON       string
		configID       int64
	)
	err := repo.db.QueryRowContext(ctx, `
		SELECT r.id, IFNULL(r.approval_number, ''), IFNULL(r.approval_type, 0), IFNULL(applicant.nick_name, ''),
		       r.approval_time, r.finish_time, r.approval_status, IFNULL(r.config_version, 0),
		       IFNULL(r.initiate_reason, ''),
		       r.current_step, IFNULL(r.current_approver, ''), IFNULL(c.rule_json, ''), IFNULL(c.id, 0)
		FROM approval_record r
		LEFT JOIN inst_user applicant ON applicant.id = r.applicant
		LEFT JOIN inst_approval_config c ON c.inst_id = r.inst_id AND c.type = r.approval_type AND c.del_flag = 0
		WHERE r.id = ? AND r.inst_id = ? AND r.del_flag = 0
		ORDER BY c.id DESC
		LIMIT 1
	`, approvalID, instID).Scan(
		&record.ID,
		&record.ApprovalNumber,
		&record.ApprovalType,
		&record.InitiateStaffName,
		&initiateTime,
		&finishTime,
		&status,
		&configVersion,
		&initiateReason,
		&currentStep,
		&currentRaw,
		&ruleJSON,
		&configID,
	)
	if err != nil {
		return model.ApprovalDetailVO{}, err
	}

	if initiateTime.Valid {
		t := initiateTime.Time
		record.InitiateTime = &t
	}
	if finishTime.Valid {
		t := finishTime.Time
		record.FinishTime = &t
	}
	if status.Valid {
		value := int(status.Int64)
		record.Status = &value
	}
	record.InitiateReason = strings.TrimSpace(initiateReason)
	if record.InitiateReason == "" {
		record.InitiateReason = buildApprovalInitiateReason(record.ApprovalType, ruleJSON)
	}

	if configID <= 0 {
		return record, nil
	}

	flows, err := repo.getApprovalFlowsForConfigVersion(ctx, configID, configVersion)
	if err != nil {
		return model.ApprovalDetailVO{}, err
	}
	if len(flows) == 0 {
		return record, nil
	}

	userIDs := make(map[int64]struct{})
	for _, uid := range splitCSV(currentRaw) {
		userIDs[uid] = struct{}{}
	}
	for _, flow := range flows {
		for _, uid := range flow.StaffIDs {
			userIDs[uid] = struct{}{}
		}
	}
	userNames, userDisabled, err := repo.getApprovalUsers(ctx, userIDs)
	if err != nil {
		return model.ApprovalDetailVO{}, err
	}
	historyByApproval, err := repo.getApprovalHistories(ctx, []int64{approvalID})
	if err != nil {
		return model.ApprovalDetailVO{}, err
	}

	var currentStepValue *int
	if currentStep.Valid {
		value := int(currentStep.Int64)
		currentStepValue = &value
	}
	record.ApproveFlows = buildApprovalFlowStages(flows, historyByApproval[approvalID], currentStepValue, userNames, userDisabled)
	return record, nil
}

func (repo *Repository) SaveApprovalTemplates(ctx context.Context, instID, operatorID int64, dto model.ApprovalTemplateSaveRequest) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	existingByType := make(map[int]struct {
		ID            int64
		ConfigVersion int
	})
	rows, err := tx.QueryContext(ctx, `
		SELECT id, type, IFNULL(config_version, 0)
		FROM inst_approval_config
		WHERE inst_id = ? AND del_flag = 0
	`, instID)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id int64
		var typeID int
		var version int
		if err := rows.Scan(&id, &typeID, &version); err != nil {
			rows.Close()
			return err
		}
		existingByType[typeID] = struct {
			ID            int64
			ConfigVersion int
		}{ID: id, ConfigVersion: version}
	}
	rows.Close()

	for _, item := range dto.ApproveTemplateRequests {
		if _, ok := approvalTemplateTypeNames[item.Type]; !ok {
			continue
		}
		configID := item.ID
		currentVersion := 0
		if existing, ok := existingByType[item.Type]; ok {
			configID = existing.ID
			currentVersion = existing.ConfigVersion
		}

		nextVersion := currentVersion + 1
		if configID <= 0 {
			result, err := tx.ExecContext(ctx, `
				INSERT INTO inst_approval_config (
					uuid, version, inst_id, name, type, enable, rule_json, config_version,
					create_id, create_time, update_id, update_time, del_flag
				) VALUES (
					UUID(), 0, ?, ?, ?, ?, ?, ?, ?, NOW(), ?, NOW(), 0
				)
			`, instID, approvalTemplateTypeNames[item.Type], item.Type, item.Enable, strings.TrimSpace(item.RuleJSON), nextVersion, operatorID, operatorID)
			if err != nil {
				return err
			}
			configID, err = result.LastInsertId()
			if err != nil {
				return err
			}
		} else {
			if _, err := tx.ExecContext(ctx, `
				UPDATE inst_approval_flow
				SET del_flag = 1, update_id = ?, update_time = NOW()
				WHERE config_id = ? AND del_flag = 0
			`, operatorID, configID); err != nil {
				return err
			}

			if _, err := tx.ExecContext(ctx, `
				UPDATE inst_approval_config
				SET enable = ?, rule_json = ?, config_version = ?, update_id = ?, update_time = NOW()
				WHERE id = ? AND inst_id = ? AND del_flag = 0
			`, item.Enable, strings.TrimSpace(item.RuleJSON), nextVersion, operatorID, configID, instID); err != nil {
				return err
			}
		}

		for _, flow := range item.FlowRequestModels {
			if flow.Step <= 0 || len(flow.StaffIDs) == 0 {
				continue
			}
			staffNames, err := repo.getApprovalStaffNamesTx(ctx, tx, flow.StaffIDs)
			if err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO inst_approval_flow (
					uuid, version, config_id, config_version, staff_id, staff_name, step,
					create_id, create_time, update_id, update_time, del_flag
				) VALUES (
					UUID(), 0, ?, ?, ?, ?, ?, ?, NOW(), ?, NOW(), 0
				)
			`, configID, nextVersion, joinInt64CSV(flow.StaffIDs), strings.Join(staffNames, ","), flow.Step, operatorID, operatorID); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (repo *Repository) PageStaffSummaries(ctx context.Context, instID int64, query model.StaffSummaryQueryDTO) (model.StaffSummaryPageVO, error) {
	current := query.PageRequestModel.PageIndex
	size := query.PageRequestModel.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 20
	}
	offset := (current - 1) * size

	filters := []string{"iu.del_flag = 0", "iu.inst_id = ?"}
	args := []any{instID}
	if strings.TrimSpace(query.QueryModel.SearchKey) != "" {
		kw := "%" + strings.TrimSpace(query.QueryModel.SearchKey) + "%"
		filters = append(filters, "(iu.nick_name LIKE ? OR iu.mobile LIKE ?)")
		args = append(args, kw, kw)
	}
	whereSQL := strings.Join(filters, " AND ")

	var total int
	if err := repo.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM inst_user iu WHERE `+whereSQL, args...).Scan(&total); err != nil {
		return model.StaffSummaryPageVO{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT iu.id, IFNULL(iu.nick_name, ''), IFNULL(iu.mobile, ''), IFNULL(iu.is_admin, 0), IFNULL(iu.avatar, ''),
		       IFNULL(iu.disabled, 0), iu.create_time, IFNULL(iu.user_type, 0)
		FROM inst_user iu
		WHERE `+whereSQL+`
		ORDER BY iu.create_time DESC, iu.id DESC
		LIMIT ? OFFSET ?
	`, append(args, size, offset)...)
	if err != nil {
		return model.StaffSummaryPageVO{}, err
	}
	defer rows.Close()

	items := make([]model.StaffSummaryVO, 0, size)
	for rows.Next() {
		var (
			item      model.StaffSummaryVO
			id        int64
			createdAt sql.NullTime
			disabled  bool
		)
		if err := rows.Scan(&id, &item.Name, &item.Phone, &item.SuperAdmin, &item.Avatar, &disabled, &createdAt, &item.EmployeeType); err != nil {
			return model.StaffSummaryPageVO{}, err
		}
		item.ID = strconv.FormatInt(id, 10)
		item.Color = staffSummaryColors[int(id)%len(staffSummaryColors)]
		if disabled {
			item.Status = 2
		} else {
			item.Status = 1
		}
		if createdAt.Valid {
			t := createdAt.Time
			item.CreatedAt = &t
		}
		items = append(items, item)
	}
	return model.StaffSummaryPageVO{List: items, Total: total}, rows.Err()
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
	if err := repo.insertApprovalHistoryTx(ctx, tx, dto.ID, int(currentStep.Int64), operatorID, 1, now, strings.TrimSpace(dto.Remark)); err != nil {
		return err
	}
	approvedSet := map[int64]struct{}{operatorID: {}}

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

	historyRows, err := tx.QueryContext(ctx, `
		SELECT approval_person
		FROM approval_history
		WHERE approval_id = ? AND del_flag = 0 AND approval_status = 1
	`, dto.ID)
	if err != nil {
		return err
	}
	for historyRows.Next() {
		var approvalPerson sql.NullInt64
		if err := historyRows.Scan(&approvalPerson); err != nil {
			historyRows.Close()
			return err
		}
		if approvalPerson.Valid {
			approvedSet[approvalPerson.Int64] = struct{}{}
		}
	}
	historyRows.Close()
	if err := historyRows.Err(); err != nil {
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
	nextFlows := make([]approvalFlowStep, 0, 4)
	for rows.Next() {
		var flow approvalFlowStep
		if err := rows.Scan(&flow.Step, &flow.StaffIDRaw); err != nil {
			return err
		}
		if strings.TrimSpace(flow.StaffIDRaw) == "" {
			continue
		}
		flow.StaffIDs = splitCSV(flow.StaffIDRaw)
		nextFlows = append(nextFlows, flow)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	allApproved, err := repo.advanceApprovalRecordTx(ctx, tx, dto.ID, instID, operatorID, nextFlows, approvedSet, now)
	if err != nil {
		return err
	}
	if allApproved && approvalType == 1 {
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
			UUID(), 0, ?, ?, ?, ?, 3, ?, ?, ?, ?, 0, ?
		)
	`, dto.ID, currentStep.Int64, operatorID, now, operatorID, now, operatorID, now, strings.TrimSpace(dto.Remark)); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE approval_record
		SET approval_status = 3, finish_time = ?, update_id = ?, update_time = NOW()
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

func (repo *Repository) RejectApprovalRecord(ctx context.Context, instID, operatorID int64, dto model.ApprovalOperateDTO) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var (
		orderID         int64
		approvalType    int
		approvalStatus  int
		currentApprover string
		currentStep     sql.NullInt64
	)
	err = tx.QueryRowContext(ctx, `
		SELECT order_id, IFNULL(approval_type, 0), IFNULL(approval_status, 0),
		       IFNULL(current_approver, ''), current_step
		FROM approval_record
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, dto.ID, instID).Scan(&orderID, &approvalType, &approvalStatus, &currentApprover, &currentStep)
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

type approvalFlowStep struct {
	Step       int
	StaffIDRaw string
	StaffIDs   []int64
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

func (repo *Repository) getApprovalTemplateFlows(ctx context.Context, configIDs []int64) (map[int64][]model.ApprovalTemplateFlowVO, error) {
	result := make(map[int64][]model.ApprovalTemplateFlowVO)
	if len(configIDs) == 0 {
		return result, nil
	}
	holders := strings.TrimRight(strings.Repeat("?,", len(configIDs)), ",")
	args := make([]any, 0, len(configIDs))
	for _, id := range configIDs {
		args = append(args, id)
	}
	rows, err := repo.db.QueryContext(ctx, `
		SELECT f.config_id, f.step, IFNULL(f.staff_id, ''), IFNULL(f.staff_name, '')
		FROM inst_approval_flow f
		INNER JOIN (
			SELECT config_id, MAX(config_version) AS config_version
			FROM inst_approval_flow
			WHERE del_flag = 0 AND config_id IN (`+holders+`)
			GROUP BY config_id
		) current_flow ON current_flow.config_id = f.config_id AND current_flow.config_version = f.config_version
		WHERE f.del_flag = 0
		ORDER BY f.config_id ASC, f.step ASC, f.id ASC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			configID   int64
			step       int
			staffIDs   string
			staffNames string
		)
		if err := rows.Scan(&configID, &step, &staffIDs, &staffNames); err != nil {
			return nil, err
		}
		result[configID] = append(result[configID], model.ApprovalTemplateFlowVO{
			Step:       step,
			StaffIDs:   splitStringCSV(staffIDs),
			StaffNames: splitStringCSV(staffNames),
		})
	}
	return result, rows.Err()
}

func (repo *Repository) getApprovalFlowsForConfigVersion(ctx context.Context, configID int64, configVersion int) ([]approvalFlowMeta, error) {
	if configID <= 0 {
		return nil, nil
	}
	if configVersion <= 0 {
		var currentVersion int
		if err := repo.db.QueryRowContext(ctx, `
			SELECT IFNULL(MAX(config_version), 0)
			FROM inst_approval_flow
			WHERE config_id = ? AND del_flag = 0
		`, configID).Scan(&currentVersion); err != nil {
			return nil, err
		}
		configVersion = currentVersion
	}
	rows, err := repo.db.QueryContext(ctx, `
		SELECT step, IFNULL(staff_id, '')
		FROM inst_approval_flow
		WHERE config_id = ? AND config_version = ? AND del_flag = 0
		ORDER BY step ASC, id ASC
	`, configID, configVersion)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	flows := make([]approvalFlowMeta, 0, 4)
	for rows.Next() {
		var step int
		var staffID string
		if err := rows.Scan(&step, &staffID); err != nil {
			return nil, err
		}
		flows = append(flows, approvalFlowMeta{Step: step, StaffIDs: splitCSV(staffID)})
	}
	sort.Slice(flows, func(i, j int) bool { return flows[i].Step < flows[j].Step })
	return flows, rows.Err()
}

func (repo *Repository) getApprovalStaffNamesTx(ctx context.Context, tx *sql.Tx, staffIDs []int64) ([]string, error) {
	if len(staffIDs) == 0 {
		return nil, nil
	}
	holders := strings.TrimRight(strings.Repeat("?,", len(staffIDs)), ",")
	args := make([]any, 0, len(staffIDs))
	for _, id := range staffIDs {
		args = append(args, id)
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT id, IFNULL(nick_name, '')
		FROM inst_user
		WHERE id IN (`+holders+`)
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	namesByID := make(map[int64]string, len(staffIDs))
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		namesByID[id] = name
	}
	names := make([]string, 0, len(staffIDs))
	for _, id := range staffIDs {
		names = append(names, namesByID[id])
	}
	return names, rows.Err()
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

func firstMatchedApprovedStaff(staffIDs []int64, approvedSet map[int64]struct{}) (int64, bool) {
	for _, staffID := range staffIDs {
		if _, ok := approvedSet[staffID]; ok {
			return staffID, true
		}
	}
	return 0, false
}

func (repo *Repository) insertApprovalHistoryTx(ctx context.Context, tx *sql.Tx, approvalID int64, step int, approvalPerson int64, status int, now time.Time, remark string) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO approval_history (
			uuid, version, approval_id, step, approval_person, approval_time, approval_status,
			create_id, create_time, update_id, update_time, del_flag, remark
		) VALUES (
			UUID(), 0, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?
		)
	`, approvalID, step, approvalPerson, now, status, approvalPerson, now, approvalPerson, now, strings.TrimSpace(remark))
	return err
}

func (repo *Repository) advanceApprovalRecordTx(ctx context.Context, tx *sql.Tx, approvalID, instID, operatorID int64, flows []approvalFlowStep, approvedSet map[int64]struct{}, now time.Time) (bool, error) {
	for _, flow := range flows {
		if matchedID, ok := firstMatchedApprovedStaff(flow.StaffIDs, approvedSet); ok {
			if err := repo.insertApprovalHistoryTx(ctx, tx, approvalID, flow.Step, matchedID, 1, now, "系统自动执行，原因：审批人此前已审批通过"); err != nil {
				return false, err
			}
			approvedSet[matchedID] = struct{}{}
			continue
		}

		if _, err := tx.ExecContext(ctx, `
			UPDATE approval_record
			SET current_step = ?, current_approver = ?, approval_status = 0, update_id = ?, update_time = NOW()
			WHERE id = ? AND inst_id = ? AND del_flag = 0
		`, flow.Step, flow.StaffIDRaw, operatorID, approvalID, instID); err != nil {
			return false, err
		}
		return false, nil
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE approval_record
		SET approval_status = 1, finish_time = ?, update_id = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, now, operatorID, approvalID, instID); err != nil {
		return false, err
	}
	return true, nil
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

func buildApprovalInitiateReason(approvalType int, ruleJSON string) string {
	raw := strings.TrimSpace(ruleJSON)
	if raw == "" {
		return "不限制，订单提交/支付后即生成审批"
	}

	switch approvalType {
	case 1:
		var rule approvalRegistrationRule
		if err := json.Unmarshal([]byte(raw), &rule); err != nil {
			return "不限制，订单提交/支付后即生成审批"
		}
		conditions := make([]string, 0, 5)
		if rule.ClassTimeFreeQuantity > 0 {
			conditions = append(conditions, fmt.Sprintf("赠送课时＞%s课时", formatApprovalThreshold(rule.ClassTimeFreeQuantity)))
		}
		if rule.PriceFreeQuantity > 0 {
			conditions = append(conditions, fmt.Sprintf("赠送金额＞%s元", formatApprovalThreshold(rule.PriceFreeQuantity)))
		}
		if rule.DateFreeQuantity > 0 {
			conditions = append(conditions, fmt.Sprintf("赠送天数＞%s天", formatApprovalThreshold(rule.DateFreeQuantity)))
		}
		if rule.Discount > 0 {
			conditions = append(conditions, fmt.Sprintf("整单优惠折扣低于%s折", formatApprovalThreshold(rule.Discount)))
		}
		if rule.DiscountPrice > 0 {
			conditions = append(conditions, fmt.Sprintf("整单优惠金额超过%s元", formatApprovalThreshold(rule.DiscountPrice)))
		}
		if len(conditions) == 0 {
			return "不限制，订单提交/支付后即生成审批"
		}
		return "限制条件，" + strings.Join(conditions, "；") + "；满足任一条件即生成审批"
	default:
		return "不限制，订单提交/支付后即生成审批"
	}
}

func formatApprovalThreshold(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
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

func splitStringCSV(raw string) []string {
	parts := strings.Split(strings.TrimSpace(raw), ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		result = append(result, part)
	}
	return result
}
