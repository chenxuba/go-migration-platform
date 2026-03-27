package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/repository"
)

func (svc *Service) UploadRechargeAccountImportFile(filename string, data []byte) (model.IntentionStudentImportUploadResult, error) {
	if len(data) == 0 {
		return model.IntentionStudentImportUploadResult{}, errors.New("empty file")
	}
	ticket := saveUploadedImportFile(uploadedImportFile{
		FileName:  strings.TrimSpace(filename),
		Data:      data,
		ExpiresAt: time.Now().Add(2 * time.Hour),
	})
	return model.IntentionStudentImportUploadResult{
		FileURL:  "/api/v1/recharge-accounts/import-uploaded-file?ticket=" + ticket,
		FileName: strings.TrimSpace(filename),
	}, nil
}

func (svc *Service) LoadUploadedRechargeAccountImportFile(ticket string) (string, []byte, bool) {
	file, ok := loadUploadedImportFile(ticket)
	if !ok {
		return "", nil, false
	}
	return file.FileName, file.Data, true
}

func (svc *Service) SubmitRechargeAccountImportTask(userID int64, req model.IntentionStudentImportSubmitRequest) (string, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("no institution context")
		}
		return "", err
	}
	instUserID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("no institution user context")
		}
		return "", err
	}
	uploadStaffName := svc.repo.GetStaffNameByID(context.Background(), &instUserID)

	fileBytes, err := loadRechargeAccountImportFileBytes(context.Background(), req.FileURL)
	if err != nil {
		return "", err
	}
	parseResult, err := svc.ParseRechargeAccountImportFile(userID, req.FileName, readerFromBytes(fileBytes))
	if err != nil {
		return "", err
	}

	now := time.Now()
	taskID := parseResult.ImportID
	task := model.IntentionStudentImportTaskDetail{
		ID:              taskID,
		FileName:        strings.TrimSpace(req.FileName),
		UploadStaffID:   fmt.Sprintf("%d", instUserID),
		UploadStaffName: uploadStaffName,
		TotalRows:       len(parseResult.Rows),
		ExecutedRows:    0,
		DeletedRows:     0,
		ErrorRows:       parseResult.AbnormalCount,
		CreatedTime:     &now,
		Status:          3,
		InstName:        parseResult.InstName,
	}
	if err := svc.repo.CreateRechargeAccountImportTask(context.Background(), instID, task, parseResult.Columns, parseResult.Rows); err != nil {
		return "", err
	}
	return taskID, nil
}

func (svc *Service) GetRechargeAccountImportTaskDetail(taskID string) (model.IntentionStudentImportTaskDetail, error) {
	task, err := svc.repo.GetRechargeAccountImportTask(context.Background(), taskID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.IntentionStudentImportTaskDetail{}, errors.New("import task not found")
		}
		return model.IntentionStudentImportTaskDetail{}, err
	}
	return task.Detail, nil
}

func (svc *Service) GetRechargeAccountImportTaskRecordList(taskID string, taskType int) (model.IntentionStudentImportTaskRecordListResult, error) {
	task, err := svc.repo.GetRechargeAccountImportTask(context.Background(), taskID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.IntentionStudentImportTaskRecordListResult{}, errors.New("import task not found")
		}
		return model.IntentionStudentImportTaskRecordListResult{}, err
	}
	items := make([]model.IntentionStudentImportRow, 0, len(task.Rows))
	for _, row := range task.Rows {
		if task.Detail.Status == 3 {
			if taskType == 0 && row.HasError {
				items = append(items, row)
				continue
			}
			if taskType == 1 && !row.HasError {
				items = append(items, row)
			}
			continue
		}
		if taskType == 0 && row.Status != 1 {
			items = append(items, row)
			continue
		}
		if taskType == 1 && row.Status == 1 {
			items = append(items, row)
		}
	}
	return model.IntentionStudentImportTaskRecordListResult{List: items, Total: len(items), Columns: task.Columns}, nil
}

func (svc *Service) BatchSaveRechargeAccountImportTaskRecords(userID int64, req model.IntentionStudentImportSaveTaskRecordRequest) ([]model.IntentionStudentImportRow, error) {
	task, err := svc.repo.GetRechargeAccountImportTask(context.Background(), req.TaskID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("import task not found")
		}
		return nil, err
	}

	importMode := detectRechargeAccountImportModeByColumns(extractImportColumnTitles(task.Columns))
	optionMap, err := svc.loadRechargeAccountImportOptionMap(userID, importMode)
	if err != nil {
		return nil, err
	}

	columnMap := make(map[string]model.IntentionStudentImportColumn, len(task.Columns))
	for _, column := range task.Columns {
		columnMap[column.Key] = column
	}
	rowMap := make(map[string]model.IntentionStudentImportRow, len(task.Rows))
	for _, row := range task.Rows {
		rowMap[row.ID] = row
	}

	updatedRows := make([]model.IntentionStudentImportRow, 0, len(req.Records))
	for _, incoming := range req.Records {
		current, ok := rowMap[incoming.ID]
		if !ok {
			continue
		}
		cellMap := make(map[string]*model.IntentionStudentImportCell, len(current.Cells))
		for idx := range current.Cells {
			cellMap[current.Cells[idx].Key] = &current.Cells[idx]
		}
		for _, incomingCell := range incoming.Cells {
			if cell, ok := cellMap[incomingCell.Key]; ok {
				column := columnMap[incomingCell.Key]
				normalizeImportedCellValue(cell, incomingCell, column, optionMap[column.Title])
				if column.Title == "收款方式" && strings.TrimSpace(cell.Value) == "其他" {
					cell.Value = "其他方式"
					cell.SelectedID = "6"
				}
				if strings.TrimSpace(cell.Value) == "" {
					switch column.Title {
					case "经办日期":
						cell.Value = time.Now().Format("2006-01-02")
					case "收款方式":
						cell.Value = "其他方式"
						cell.SelectedID = "6"
					case "收款账户":
						cell.Value = "默认账户"
						cell.SelectedID = "default"
					}
				}
				cell.Error = validateImportedCell(column, *cell, optionMap[column.Title])
			}
		}
		clearOrderImportRowErrors(&current)
		for idx := range current.Cells {
			column := columnMap[current.Cells[idx].Key]
			current.Cells[idx].Error = validateImportedCell(column, current.Cells[idx], optionMap[column.Title])
		}
		current.HasError = false
		applyRechargeAccountImportRowValidation(importMode, current.Cells, &current.HasError)
		for _, cell := range current.Cells {
			if strings.TrimSpace(cell.Error) != "" {
				current.HasError = true
				break
			}
		}
		rowMap[current.ID] = current
		updatedRows = append(updatedRows, current)
	}

	task.Rows = make([]model.IntentionStudentImportRow, 0, len(rowMap))
	for _, row := range rowMap {
		task.Rows = append(task.Rows, row)
	}
	sort.Slice(task.Rows, func(i, j int) bool { return task.Rows[i].RowNo < task.Rows[j].RowNo })
	task.Detail.TotalRows = len(task.Rows)
	task.Detail.ErrorRows = countImportTaskErrors(task.Rows)
	task.Detail.Status = 3
	if err := svc.repo.UpdateRechargeAccountImportTask(context.Background(), task.Detail, task.Rows); err != nil {
		return nil, err
	}
	return updatedRows, nil
}

func (svc *Service) StartRechargeAccountImportTask(userID int64, taskID string) (model.OrderImportStartResult, error) {
	task, err := svc.repo.GetRechargeAccountImportTask(context.Background(), taskID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.OrderImportStartResult{}, errors.New("import task not found")
		}
		return model.OrderImportStartResult{}, err
	}
	if task.Detail.Status != 3 {
		return model.OrderImportStartResult{SuccessCount: task.Detail.ExecutedRows, FailCount: task.Detail.ErrorRows}, nil
	}

	instUserID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.OrderImportStartResult{}, errors.New("no institution user context")
		}
		return model.OrderImportStartResult{}, err
	}
	executeName := svc.repo.GetStaffNameByID(context.Background(), &instUserID)
	now := time.Now()
	task.Detail.ExecuteStaffID = stringPtr(strconv.FormatInt(instUserID, 10))
	task.Detail.ExecuteStaffName = stringPtr(executeName)
	task.Detail.ConfirmTime = &now
	task.Detail.Status = 4
	for idx := range task.Rows {
		if task.Rows[idx].Status == 0 {
			task.Rows[idx].Result = ""
		}
	}
	if err := svc.repo.MarkRechargeAccountImportTaskRunning(context.Background(), task.Detail, task.Rows); err != nil {
		if errors.Is(err, repository.ErrRechargeAccountImportTaskStartConflict) {
			refreshed, getErr := svc.repo.GetRechargeAccountImportTask(context.Background(), taskID)
			if getErr == nil {
				return model.OrderImportStartResult{SuccessCount: refreshed.Detail.ExecutedRows, FailCount: refreshed.Detail.ErrorRows}, nil
			}
		}
		return model.OrderImportStartResult{}, err
	}

	go svc.runRechargeAccountImportTask(userID, taskID)
	return model.OrderImportStartResult{SuccessCount: task.Detail.ExecutedRows, FailCount: task.Detail.ErrorRows}, nil
}

func (svc *Service) runRechargeAccountImportTask(userID int64, taskID string) {
	task, err := svc.repo.GetRechargeAccountImportTask(context.Background(), taskID)
	if err != nil {
		return
	}
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		svc.finishRechargeAccountImportTaskWithFatalError(&task.Detail, task.Rows, fmt.Sprintf("获取机构信息失败：%v", err))
		return
	}
	importMode := detectRechargeAccountImportModeByColumns(extractImportColumnTitles(task.Columns))
	if importMode == rechargeAccountImportModeUnknown {
		svc.finishRechargeAccountImportTaskWithFatalError(&task.Detail, task.Rows, "当前导入模板暂不支持执行导入")
		return
	}
	optionMap, err := svc.loadRechargeAccountImportOptionMap(userID, importMode)
	if err != nil {
		svc.finishRechargeAccountImportTaskWithFatalError(&task.Detail, task.Rows, fmt.Sprintf("加载导入配置失败：%v", err))
		return
	}
	orderTagMap, err := svc.repo.ListEnabledOrderTagNameIDMap(context.Background(), instID)
	if err != nil {
		svc.finishRechargeAccountImportTaskWithFatalError(&task.Detail, task.Rows, fmt.Sprintf("加载订单标签失败：%v", err))
		return
	}

	columnMap := make(map[string]model.IntentionStudentImportColumn, len(task.Columns))
	for _, column := range task.Columns {
		columnMap[column.Key] = column
	}

	for idx := range task.Rows {
		row := &task.Rows[idx]
		if row.Status != 0 {
			continue
		}
		if row.HasError {
			row.Status = 2
			if strings.TrimSpace(row.Result) == "" {
				row.Result = "请先处理异常数据"
			}
		} else {
			var importErr error
			switch importMode {
			case rechargeAccountImportModeByAccount:
				importErr = svc.importRechargeAccountRowByAccount(userID, instID, *row, columnMap, optionMap, orderTagMap)
			default:
				importErr = svc.importRechargeAccountRowByStudent(userID, instID, *row, columnMap, optionMap, orderTagMap)
			}
			if importErr != nil {
				row.Status = 2
				row.Result = importErr.Error()
				row.HasError = true
				attachOrderImportRowError(row, importErr.Error())
			} else {
				row.Status = 1
				row.Result = "导入成功"
				row.HasError = false
			}
		}
		task.Detail.ExecutedRows, task.Detail.ErrorRows = summarizeOrderImportRows(task.Rows)
		if err := svc.repo.UpdateRechargeAccountImportTaskProgress(context.Background(), task.Detail, *row); err != nil {
			svc.finishRechargeAccountImportTaskWithFatalError(&task.Detail, task.Rows, fmt.Sprintf("保存导入进度失败：%v", err))
			return
		}
	}

	completeTime := time.Now()
	task.Detail.CompleteTime = &completeTime
	task.Detail.Status = 1
	_ = svc.repo.UpdateRechargeAccountImportTask(context.Background(), task.Detail, task.Rows)
}

func (svc *Service) ListRechargeAccountImportTasks(userID int64) (model.IntentionStudentImportTaskListResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.IntentionStudentImportTaskListResult{}, errors.New("no institution context")
		}
		return model.IntentionStudentImportTaskListResult{}, err
	}
	return svc.repo.ListRechargeAccountImportTasks(context.Background(), instID)
}

func (svc *Service) ClearRechargeAccountImportTasks(userID int64) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	return svc.repo.ClearRechargeAccountImportTasks(context.Background(), instID)
}

func (svc *Service) DeleteRechargeAccountImportTask(userID int64, taskID string) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	if err := svc.repo.DeleteRechargeAccountImportTask(context.Background(), instID, taskID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("import task not found")
		}
		return err
	}
	return nil
}

func (svc *Service) loadRechargeAccountImportOptionMap(userID int64, importMode rechargeAccountImportMode) (map[string][]importOptionItem, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		return nil, err
	}
	defaultFields, err := svc.repo.ListStudentFields(context.Background(), instID, true)
	if err != nil {
		return nil, err
	}
	channels, err := svc.repo.GetChannels(context.Background(), instID)
	if err != nil {
		return nil, err
	}
	staffs, err := svc.repo.ListActiveStaffOptions(context.Background(), instID)
	if err != nil {
		return nil, err
	}
	orderTagNames, err := svc.repo.ListEnabledOrderTagNames(context.Background(), instID)
	if err != nil {
		return nil, err
	}

	result := map[string][]importOptionItem{
		"手机号归属人": {
			{Label: "爸爸", Value: "1"},
			{Label: "妈妈", Value: "2"},
			{Label: "爷爷", Value: "3"},
			{Label: "奶奶", Value: "4"},
			{Label: "外公", Value: "5"},
			{Label: "外婆", Value: "6"},
			{Label: "其他", Value: "7"},
		},
		"性别": {
			{Label: "男", Value: "1"},
			{Label: "女", Value: "0"},
			{Label: "未知", Value: "2"},
		},
		"收款方式": {
			{Label: "微信", Value: "1"},
			{Label: "支付宝", Value: "2"},
			{Label: "银行转账", Value: "3"},
			{Label: "POS机", Value: "4"},
			{Label: "现金", Value: "5"},
			{Label: "其他方式", Value: "6"},
		},
		"收款账户": {
			{Label: "默认账户", Value: "default"},
		},
	}
	for _, field := range defaultFields {
		if strings.TrimSpace(field.FieldKey) == "年级" && strings.TrimSpace(field.OptionsJSON) != "" {
			for _, item := range splitTemplateOptions(field.OptionsJSON) {
				result["年级"] = append(result["年级"], importOptionItem{Label: item, Value: item})
			}
		}
	}
	for _, channel := range channels {
		if channel.IsDisabled || strings.TrimSpace(channel.Name) == "" {
			continue
		}
		result["渠道"] = append(result["渠道"], importOptionItem{Label: channel.Name, Value: fmt.Sprintf("%d", channel.ID)})
	}
	for _, name := range orderTagNames {
		result["订单标签"] = append(result["订单标签"], importOptionItem{Label: name, Value: name})
	}
	for _, staff := range staffs {
		result["销售员"] = append(result["销售员"], importOptionItem{Label: staff.Name, Value: staff.ID})
		result["订单销售员"] = append(result["订单销售员"], importOptionItem{Label: staff.Name, Value: staff.ID})
	}
	_ = importMode
	return result, nil
}

func (svc *Service) importRechargeAccountRowByStudent(userID, instID int64, row model.IntentionStudentImportRow, columns map[string]model.IntentionStudentImportColumn, optionMap map[string][]importOptionItem, orderTagMap map[string]int64) error {
	studentDTO, err := buildStudentSaveDTOFromImportRow(row, columns, optionMap)
	if err != nil {
		return err
	}
	studentDTO.Remark = strings.TrimSpace(cellValueByTitle(row, "学员备注"))
	decision, err := svc.resolveOrderImportStudent(userID, instID, studentDTO)
	if err != nil {
		return err
	}
	account, err := svc.GetRechargeAccountByStudent(userID, strconv.FormatInt(decision.StudentID, 10))
	if err != nil {
		return err
	}
	return svc.importRechargeAccountOrder(userID, decision.StudentID, account.ID, row, optionMap, orderTagMap)
}

func (svc *Service) importRechargeAccountRowByAccount(userID, instID int64, row model.IntentionStudentImportRow, _ map[string]model.IntentionStudentImportColumn, optionMap map[string][]importOptionItem, orderTagMap map[string]int64) error {
	accountName := strings.TrimSpace(cellValueByTitle(row, "储值账户号"))
	accountID, studentID, err := svc.repo.FindRechargeAccountImportTargetByName(context.Background(), instID, accountName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("储值账户号不存在")
		}
		return err
	}
	return svc.importRechargeAccountOrder(userID, studentID, strconv.FormatInt(accountID, 10), row, optionMap, orderTagMap)
}

func (svc *Service) importRechargeAccountOrder(userID, studentID int64, rechargeAccountID string, row model.IntentionStudentImportRow, optionMap map[string][]importOptionItem, orderTagMap map[string]int64) error {
	orderDTO, payDTO, hasPayment, err := buildCreateRechargeAccountOrderDTOFromImportRow(userID, studentID, rechargeAccountID, row, optionMap, orderTagMap, svc)
	if err != nil {
		return err
	}
	result, err := svc.CreateRechargeAccountOrder(userID, orderDTO)
	if err != nil {
		return err
	}
	if hasPayment {
		detail, err := svc.GetRechargeAccountOrderDetail(userID, model.RechargeAccountOrderDetailQuery{RechargeAccountOrderID: result.ID})
		if err != nil {
			return err
		}
		if detail.Bill.ID == "" {
			return errors.New("账单不存在")
		}
		payDTO.BillID = detail.Bill.ID
		if _, err := svc.PayOrderBySchoolPal(userID, payDTO); err != nil {
			return err
		}
	}
	return nil
}

func finishRechargeAccountImportResultMessage(err error) string {
	if err == nil {
		return "导入成功"
	}
	return err.Error()
}

func (svc *Service) finishRechargeAccountImportTaskWithFatalError(detail *model.IntentionStudentImportTaskDetail, rows []model.IntentionStudentImportRow, message string) {
	message = strings.TrimSpace(message)
	if message == "" {
		message = "导入任务执行失败"
	}
	for idx := range rows {
		if rows[idx].Status != 0 {
			continue
		}
		rows[idx].Status = 2
		rows[idx].Result = message
		rows[idx].HasError = true
		attachOrderImportRowError(&rows[idx], message)
	}
	detail.ExecutedRows, detail.ErrorRows = summarizeOrderImportRows(rows)
	now := time.Now()
	detail.CompleteTime = &now
	detail.Status = 1
	_ = svc.repo.UpdateRechargeAccountImportTask(context.Background(), *detail, rows)
}

func extractImportColumnTitles(columns []model.IntentionStudentImportColumn) []string {
	items := make([]string, 0, len(columns))
	for _, column := range columns {
		items = append(items, column.Title)
	}
	return items
}

func stringPtr(value string) *string {
	text := strings.TrimSpace(value)
	if text == "" {
		return nil
	}
	return &text
}
