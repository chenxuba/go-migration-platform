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

type orderImportStudentDecision struct {
	StudentID  int64
	CreatedNew bool
}

func (svc *Service) UploadOrderImportFile(filename string, data []byte) (model.IntentionStudentImportUploadResult, error) {
	if len(data) == 0 {
		return model.IntentionStudentImportUploadResult{}, errors.New("empty file")
	}
	ticket := saveUploadedImportFile(uploadedImportFile{
		FileName:  strings.TrimSpace(filename),
		Data:      data,
		ExpiresAt: time.Now().Add(2 * time.Hour),
	})
	return model.IntentionStudentImportUploadResult{
		FileURL:  "/api/v1/orders/import-uploaded-file?ticket=" + ticket,
		FileName: strings.TrimSpace(filename),
	}, nil
}

func (svc *Service) LoadUploadedOrderImportFile(ticket string) (string, []byte, bool) {
	file, ok := loadUploadedImportFile(ticket)
	if !ok {
		return "", nil, false
	}
	return file.FileName, file.Data, true
}

func (svc *Service) SubmitOrderImportTask(userID int64, req model.IntentionStudentImportSubmitRequest) (string, error) {
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

	fileBytes, err := loadOrderImportFileBytes(context.Background(), req.FileURL)
	if err != nil {
		return "", err
	}
	parseResult, err := svc.ParseOrderImportFile(userID, req.FileName, readerFromBytes(fileBytes))
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
	if err := svc.repo.CreateOrderImportTask(context.Background(), instID, task, parseResult.Columns, parseResult.Rows); err != nil {
		return "", err
	}
	return taskID, nil
}

func (svc *Service) GetOrderImportTaskDetail(taskID string) (model.IntentionStudentImportTaskDetail, error) {
	task, err := svc.repo.GetOrderImportTask(context.Background(), taskID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.IntentionStudentImportTaskDetail{}, errors.New("import task not found")
		}
		return model.IntentionStudentImportTaskDetail{}, err
	}
	return task.Detail, nil
}

func (svc *Service) GetOrderImportTaskRecordList(taskID string, taskType int) (model.IntentionStudentImportTaskRecordListResult, error) {
	task, err := svc.repo.GetOrderImportTask(context.Background(), taskID)
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
		if taskType == 0 && row.Status == 2 {
			items = append(items, row)
			continue
		}
		if taskType == 1 && row.Status == 1 {
			items = append(items, row)
		}
	}
	return model.IntentionStudentImportTaskRecordListResult{
		List:    items,
		Total:   len(items),
		Columns: task.Columns,
	}, nil
}

func (svc *Service) BatchSaveOrderImportTaskRecords(userID int64, req model.IntentionStudentImportSaveTaskRecordRequest) ([]model.IntentionStudentImportRow, error) {
	task, err := svc.repo.GetOrderImportTask(context.Background(), req.TaskID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("import task not found")
		}
		return nil, err
	}
	importMode := detectOrderImportModeByColumns(task.Columns)
	optionMap, err := svc.loadOrderImportOptionMap(userID, importMode)
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
		applyOrderImportRowValidation(importMode, current.Cells, &current.HasError)
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
	if err := svc.repo.UpdateOrderImportTask(context.Background(), task.Detail, task.Rows); err != nil {
		return nil, err
	}
	return updatedRows, nil
}

func (svc *Service) DeleteOrderImportTask(userID int64, taskID string) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	if err := svc.repo.DeleteOrderImportTask(context.Background(), instID, taskID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("import task not found")
		}
		return err
	}
	return nil
}

func (svc *Service) StartOrderImportTask(userID int64, taskID string) (model.OrderImportStartResult, error) {
	task, err := svc.repo.GetOrderImportTask(context.Background(), taskID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.OrderImportStartResult{}, errors.New("import task not found")
		}
		return model.OrderImportStartResult{}, err
	}
	switch task.Detail.Status {
	case 4:
		return model.OrderImportStartResult{}, errors.New("导入任务正在执行，请勿重复操作")
	case 1:
		return model.OrderImportStartResult{}, errors.New("导入任务已完成")
	case 3:
	default:
		return model.OrderImportStartResult{}, errors.New("当前导入任务状态不允许开始导入")
	}
	if countImportTaskErrors(task.Rows) > 0 {
		return model.OrderImportStartResult{}, errors.New("请先处理异常数据")
	}

	instUserID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.OrderImportStartResult{}, errors.New("no institution user context")
		}
		return model.OrderImportStartResult{}, err
	}
	executorName := svc.repo.GetStaffNameByID(context.Background(), &instUserID)

	now := time.Now()
	executorID := fmt.Sprintf("%d", instUserID)
	task.Detail.ExecuteStaffID = &executorID
	task.Detail.ExecuteStaffName = &executorName
	task.Detail.ExecutedRows = 0
	task.Detail.ErrorRows = 0
	task.Detail.Status = 4
	task.Detail.ConfirmTime = &now
	task.Detail.CompleteTime = nil
	for idx := range task.Rows {
		task.Rows[idx].Status = 0
		task.Rows[idx].Result = ""
	}
	if err := svc.repo.MarkOrderImportTaskRunning(context.Background(), task.Detail, task.Rows); err != nil {
		if errors.Is(err, repository.ErrOrderImportTaskStartConflict) {
			return model.OrderImportStartResult{}, errors.New("导入任务正在执行，请勿重复操作")
		}
		return model.OrderImportStartResult{}, err
	}
	go svc.runOrderImportTask(userID, taskID)
	return model.OrderImportStartResult{}, nil
}

func (svc *Service) runOrderImportTask(userID int64, taskID string) {
	task, err := svc.repo.GetOrderImportTask(context.Background(), taskID)
	if err != nil {
		return
	}
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		svc.finishOrderImportTaskWithFatalError(&task.Detail, task.Rows, fmt.Sprintf("获取机构信息失败：%v", err))
		return
	}
	importMode := detectOrderImportModeByColumns(task.Columns)
	if importMode == orderImportModeUnknown {
		svc.finishOrderImportTaskWithFatalError(&task.Detail, task.Rows, "当前导入模板暂不支持执行导入")
		return
	}
	optionMap, err := svc.loadOrderImportOptionMap(userID, importMode)
	if err != nil {
		svc.finishOrderImportTaskWithFatalError(&task.Detail, task.Rows, fmt.Sprintf("加载导入配置失败：%v", err))
		return
	}
	orderTagMap, err := svc.repo.ListEnabledOrderTagNameIDMap(context.Background(), instID)
	if err != nil {
		svc.finishOrderImportTaskWithFatalError(&task.Detail, task.Rows, fmt.Sprintf("加载订单标签失败：%v", err))
		return
	}
	courseNames := collectOrderImportColumnValues(task.Rows, task.Columns, "报读课程")
	quotationMap, err := svc.repo.ListCourseQuotationsByNamesAndLessonModel(context.Background(), instID, courseNames, lessonModelByOrderImportMode(importMode))
	if err != nil {
		svc.finishOrderImportTaskWithFatalError(&task.Detail, task.Rows, fmt.Sprintf("加载课程报价失败：%v", err))
		return
	}
	if len(task.Rows) == 0 {
		completeAt := time.Now()
		task.Detail.CompleteTime = &completeAt
		task.Detail.Status = 1
		_ = retryOrderImportTaskWrite(func() error {
			return svc.repo.UpdateOrderImportTask(context.Background(), task.Detail, task.Rows)
		})
		return
	}

	successCount := 0
	failCount := 0
	columnMap := make(map[string]model.IntentionStudentImportColumn, len(task.Columns))
	for _, column := range task.Columns {
		columnMap[column.Key] = column
	}

	for idx := range task.Rows {
		row := task.Rows[idx]
		resultText := "导入成功"
		if err := svc.importOrderRow(userID, instID, importMode, row, columnMap, optionMap, orderTagMap, quotationMap); err != nil {
			task.Rows[idx].Status = 2
			task.Rows[idx].Result = err.Error()
			task.Rows[idx].HasError = true
			attachOrderImportRowError(&task.Rows[idx], err.Error())
			resultText = err.Error()
			failCount++
		} else {
			task.Rows[idx].Status = 1
			task.Rows[idx].Result = resultText
			task.Rows[idx].HasError = false
			clearOrderImportRowErrors(&task.Rows[idx])
			successCount++
		}
		task.Detail.ExecutedRows = successCount
		task.Detail.ErrorRows = failCount
		if idx == len(task.Rows)-1 {
			completeAt := time.Now()
			task.Detail.CompleteTime = &completeAt
			task.Detail.Status = 1
		} else {
			task.Detail.CompleteTime = nil
			task.Detail.Status = 4
		}
		if err := retryOrderImportTaskWrite(func() error {
			return svc.repo.UpdateOrderImportTaskProgress(context.Background(), task.Detail, task.Rows[idx])
		}); err != nil {
			svc.finishOrderImportTaskWithFatalError(&task.Detail, task.Rows, fmt.Sprintf("保存导入进度失败：%v", err))
			return
		}
	}
}

func (svc *Service) ListOrderImportTasks(userID int64) (model.IntentionStudentImportTaskListResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.IntentionStudentImportTaskListResult{}, errors.New("no institution context")
		}
		return model.IntentionStudentImportTaskListResult{}, err
	}
	return svc.repo.ListOrderImportTasks(context.Background(), instID)
}

func (svc *Service) ClearOrderImportTasks(userID int64) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	return svc.repo.ClearOrderImportTasks(context.Background(), instID)
}

func (svc *Service) loadOrderImportOptionMap(userID int64, importMode orderImportMode) (map[string][]importOptionItem, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		return nil, err
	}
	defaultFields, err := svc.repo.ListStudentFields(context.Background(), instID, true)
	if err != nil {
		return nil, err
	}
	customFields, err := svc.repo.ListStudentFields(context.Background(), instID, false)
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
	courseNames, err := svc.loadOrderImportCourseNames(context.Background(), instID, orderImportModeLabel(importMode))
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
		"是否为体验价": {
			{Label: "是", Value: "是"},
			{Label: "否", Value: "否"},
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
	for _, name := range courseNames {
		result["报读课程"] = append(result["报读课程"], importOptionItem{Label: name, Value: name})
	}
	for _, name := range orderTagNames {
		result["订单标签"] = append(result["订单标签"], importOptionItem{Label: name, Value: name})
	}
	for _, staff := range staffs {
		result["销售"] = append(result["销售"], importOptionItem{Label: staff.Name, Value: staff.ID})
		result["订单销售员"] = append(result["订单销售员"], importOptionItem{Label: staff.Name, Value: staff.ID})
	}
	for _, field := range customFields {
		if !field.IsDisplay || field.FieldType != 4 {
			continue
		}
		for _, item := range splitTemplateOptions(field.OptionsJSON) {
			result[field.FieldKey] = append(result[field.FieldKey], importOptionItem{Label: item, Value: item})
		}
	}
	return result, nil
}

func (svc *Service) importOrderRow(
	userID int64,
	instID int64,
	importMode orderImportMode,
	row model.IntentionStudentImportRow,
	columns map[string]model.IntentionStudentImportColumn,
	optionMap map[string][]importOptionItem,
	orderTagMap map[string]int64,
	quotationMap map[string][]model.CourseQuotation,
) error {
	studentDTO, err := buildStudentSaveDTOFromImportRow(row, columns, optionMap)
	if err != nil {
		return err
	}
	isTrial := strings.TrimSpace(cellValueByTitle(row, "是否为体验价")) == "是"
	decision, err := svc.resolveOrderImportStudent(userID, instID, studentDTO)
	if err != nil {
		return err
	}

	courseName := strings.TrimSpace(cellValueByTitle(row, "报读课程"))
	quotation, err := pickOrderImportAnchorQuotation(quotationMap[courseName], isTrial, importMode)
	if err != nil {
		return err
	}
	handleType, err := svc.detectOrderImportHandleType(context.Background(), instID, decision.StudentID, quotation.CourseID)
	if err != nil {
		return err
	}

	createDTO, payDTO, hasPayment, err := buildCreateAndPayOrderDTOFromImportRow(decision.StudentID, importMode, handleType, row, columns, optionMap, orderTagMap, quotationMap)
	if err != nil {
		return err
	}
	orderID, err := svc.CreateOrder(userID, createDTO)
	if err != nil {
		return err
	}
	if hasPayment {
		payDTO.OrderID = orderID
		for idx := range payDTO.PayAccounts {
			payDTO.PayAccounts[idx].OrderID = orderID
		}
		if err := svc.PayOrder(userID, payDTO); err != nil {
			return err
		}
	}
	if decision.CreatedNew {
		targetStatus := 0
		if !isTrial {
			targetStatus = 1
		}
		if err := svc.repo.UpdateStudentStatusValue(context.Background(), instID, decision.StudentID, targetStatus); err != nil {
			return err
		}
	}
	return nil
}

func (svc *Service) detectOrderImportHandleType(ctx context.Context, instID, studentID, courseID int64) (int, error) {
	active, err := svc.repo.StudentHasActiveCourseEnrollment(ctx, instID, studentID, courseID)
	if err != nil {
		return 0, err
	}
	if active {
		return 2, nil
	}

	purchased, err := svc.repo.StudentHasCompletedOrderForCourse(ctx, instID, studentID, courseID)
	if err != nil {
		return 0, err
	}
	if purchased {
		return 2, nil
	}

	hasAnyPurchased, err := svc.repo.StudentHasCompletedOrders(ctx, instID, studentID)
	if err != nil {
		return 0, err
	}
	if hasAnyPurchased {
		return 3, nil
	}
	return 1, nil
}

func (svc *Service) resolveOrderImportStudent(userID, instID int64, dto model.StudentSaveDTO) (orderImportStudentDecision, error) {
	existingID, err := svc.repo.FindStudentIDByNameMobile(context.Background(), instID, dto.StuName, dto.Mobile)
	if err == nil && existingID > 0 {
		return orderImportStudentDecision{StudentID: existingID, CreatedNew: false}, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return orderImportStudentDecision{}, err
	}

	importRule, err := svc.repo.GetAddImportStudentRule(context.Background(), instID)
	if err != nil {
		return orderImportStudentDecision{}, err
	}
	switch importRule {
	case 2:
		count, err := svc.repo.CountStudentDuplicatesByRule(context.Background(), instID, 2, dto.StuName, dto.Mobile, nil)
		if err != nil {
			return orderImportStudentDecision{}, err
		}
		if count > 0 {
			return orderImportStudentDecision{}, errors.New("已存在相同手机号的学员，不可创建")
		}
	case 3:
		count, err := svc.repo.CountStudentDuplicatesByRule(context.Background(), instID, 3, dto.StuName, dto.Mobile, nil)
		if err != nil {
			return orderImportStudentDecision{}, err
		}
		if count > 0 {
			return orderImportStudentDecision{}, errors.New("已存在相同姓名的学员，不可创建")
		}
	}

	limitImportSameWeChat, err := svc.repo.GetLimitImportSameWeChat(context.Background(), instID)
	if err != nil {
		return orderImportStudentDecision{}, err
	}
	if limitImportSameWeChat && strings.TrimSpace(dto.WeChatNumber) != "" {
		count, err := svc.repo.CountStudentByWeChat(context.Background(), instID, dto.WeChatNumber, nil)
		if err != nil {
			return orderImportStudentDecision{}, err
		}
		if count > 0 {
			return orderImportStudentDecision{}, errors.New("已存在相同微信号的学员，不可创建")
		}
	}

	studentID, err := svc.createIntentStudentRecord(userID, instID, dto)
	if err != nil {
		return orderImportStudentDecision{}, err
	}
	return orderImportStudentDecision{StudentID: studentID, CreatedNew: true}, nil
}

func buildCreateAndPayOrderDTOFromImportRow(
	studentID int64,
	importMode orderImportMode,
	handleType int,
	row model.IntentionStudentImportRow,
	columns map[string]model.IntentionStudentImportColumn,
	optionMap map[string][]importOptionItem,
	orderTagMap map[string]int64,
	quotationMap map[string][]model.CourseQuotation,
) (model.CreateOrderDTO, model.PayOrderDTO, bool, error) {
	rowData := make(map[string]model.IntentionStudentImportCell, len(row.Cells))
	for _, cell := range row.Cells {
		rowData[cell.Title] = cell
	}

	courseName := strings.TrimSpace(rowData["报读课程"].Value)
	if courseName == "" {
		return model.CreateOrderDTO{}, model.PayOrderDTO{}, false, errors.New("报读课程不能为空")
	}
	totalAmount, err := parseOrderImportFloat(rowData["实收金额"].Value, "实收金额")
	if err != nil {
		return model.CreateOrderDTO{}, model.PayOrderDTO{}, false, err
	}
	arrearAmount, err := parseOrderImportFloat(rowData["欠费金额"].Value, "欠费金额")
	if err != nil {
		return model.CreateOrderDTO{}, model.PayOrderDTO{}, false, err
	}
	orderAmount := totalAmount + arrearAmount
	if orderAmount <= 0 {
		return model.CreateOrderDTO{}, model.PayOrderDTO{}, false, errors.New("订单金额必须大于0")
	}

	isTrial := strings.TrimSpace(rowData["是否为体验价"].Value) == "是"

	var (
		dealDate   *time.Time
		validDate  *time.Time
		endDate    *time.Time
		salePerson *int64
		payTime    *time.Time
	)
	if value := strings.TrimSpace(rowData["经办日期"].Value); value != "" {
		if parsed, ok := parseImportDateValue(value); ok {
			dealDate = &parsed
			payTime = &parsed
		}
	}

	if value, ok := resolveImportOptionInt64(rowData["订单销售员"], optionMap["订单销售员"]); ok {
		salePerson = &value
	} else if value, ok := resolveImportOptionInt64(rowData["销售"], optionMap["销售"]); ok {
		salePerson = &value
	}

	orderTagIDs := resolveOrderImportTagIDs(rowData["订单标签"], orderTagMap)
	quotation, err := pickOrderImportAnchorQuotation(quotationMap[courseName], isTrial, importMode)
	if err != nil {
		return model.CreateOrderDTO{}, model.PayOrderDTO{}, false, err
	}
	unit := 1
	if quotation.Unit != nil {
		unit = *quotation.Unit
	}
	lessonMode := lessonModelByOrderImportMode(importMode)
	if quotation.LessonModel != nil {
		lessonMode = *quotation.LessonModel
	}

	countValue := 1
	quoteID := quotation.ID
	courseID := quotation.CourseID
	courseAmount := fmt.Sprintf("%.2f", orderAmount)
	hasValidDate := false
	purchasedQuantity := 0.0
	giftCount := 0.0
	realQuantity := 0.0

	switch importMode {
	case orderImportModeTimeSlot:
		if value := strings.TrimSpace(rowData["有效开始日期"].Value); value != "" {
			if parsed, ok := parseImportDateValue(value); ok {
				validDate = &parsed
			}
		}
		if value := strings.TrimSpace(rowData["有效结束日期(含赠送天数)"].Value); value != "" {
			if parsed, ok := parseImportDateValue(value); ok {
				endDate = &parsed
			}
		}
		if validDate == nil || endDate == nil {
			return model.CreateOrderDTO{}, model.PayOrderDTO{}, false, errors.New("有效开始日期和有效结束日期不能为空")
		}
		if endDate.Before(*validDate) {
			return model.CreateOrderDTO{}, model.PayOrderDTO{}, false, errors.New("有效结束日期不能早于有效开始日期")
		}
		giftCount, err = parseOrderImportIntWithDefault(rowData["赠送天数"].Value, "赠送天数")
		if err != nil {
			return model.CreateOrderDTO{}, model.PayOrderDTO{}, false, err
		}
		totalDays := int(endDate.Sub(*validDate).Hours()/24) + 1
		purchasedQuantity = float64(totalDays) - giftCount
		if purchasedQuantity <= 0 {
			return model.CreateOrderDTO{}, model.PayOrderDTO{}, false, errors.New("有效时段需大于赠送天数")
		}
		realQuantity = purchasedQuantity + giftCount
		hasValidDate = true
	case orderImportModeAmount:
		if value := strings.TrimSpace(rowData["有效期至"].Value); value != "" {
			if parsed, ok := parseImportDateValue(value); ok {
				endDate = &parsed
			}
		}
		purchasedQuantity, err = parseOrderImportPositiveFloat(rowData["购买金额"].Value, "购买金额")
		if err != nil {
			return model.CreateOrderDTO{}, model.PayOrderDTO{}, false, err
		}
		giftCount, err = parseOrderImportFloatWithPrecision(rowData["赠送金额"].Value, "赠送金额", 2)
		if err != nil {
			return model.CreateOrderDTO{}, model.PayOrderDTO{}, false, err
		}
		realQuantity = purchasedQuantity + giftCount
		hasValidDate = endDate != nil
		courseAmount = fmt.Sprintf("%.2f", purchasedQuantity)
	default:
		if value := strings.TrimSpace(rowData["有效期至"].Value); value != "" {
			if parsed, ok := parseImportDateValue(value); ok {
				endDate = &parsed
			}
		}
		purchasedQuantity, err = parseOrderImportLessonHour(rowData["购买课时数"].Value, "购买课时数")
		if err != nil {
			return model.CreateOrderDTO{}, model.PayOrderDTO{}, false, err
		}
		giftCount, err = parseOrderImportFloatWithPrecision(rowData["赠送课时数"].Value, "赠送课时数", 2)
		if err != nil {
			return model.CreateOrderDTO{}, model.PayOrderDTO{}, false, err
		}
		realQuantity = purchasedQuantity + giftCount
		hasValidDate = endDate != nil
	}

	createDTO := model.CreateOrderDTO{
		StudentID: studentID,
		OrderDetail: model.OrderDetailDTO{
			QuoteDetailList: []model.QuoteDetailDTO{
				{
					HandleType:   &handleType,
					CourseID:     courseID,
					QuoteID:      quoteID,
					LessonMode:   &lessonMode,
					Count:        &countValue,
					Unit:         &unit,
					FreeQuantity: giftCount,
					HasValidDate: &hasValidDate,
					ValidDate:    validDate,
					EndDate:      endDate,
					Amount:       courseAmount,
					Quantity:     purchasedQuantity,
					RealQuantity: realQuantity,
					RealAmount:   courseAmount,
				},
			},
			OrderDiscountType:   nil,
			OrderDiscountNumber: 0,
			OrderDiscountAmount: "0.00",
			OrderRealQuantity:   realQuantity,
			OrderRealAmount:     courseAmount,
			InternalRemark:      strings.TrimSpace(rowData["订单备注"].Value),
			DealDate:            dealDate,
			SalePerson:          salePerson,
			OrderTagIDs:         orderTagIDs,
			OrderSource:         intPtr(model.OrderSourceOfflineImport),
		},
	}

	payDTO := model.PayOrderDTO{
		PayAmount: totalAmount,
		PayAccounts: []model.PayAccountDTO{
			{
				PayMethod: intPtr(resolvePayMethod(rowData["收款方式"].Value)),
				PayAmount: totalAmount,
				PayTime:   payTime,
			},
		},
	}
	return createDTO, payDTO, totalAmount > 0, nil
}

func pickOrderImportAnchorQuotation(items []model.CourseQuotation, isTrial bool, importMode orderImportMode) (model.CourseQuotation, error) {
	if len(items) == 0 {
		switch importMode {
		case orderImportModeTimeSlot:
			return model.CourseQuotation{}, errors.New("报读课程未配置可用的按时段报价单")
		case orderImportModeAmount:
			return model.CourseQuotation{}, errors.New("报读课程未配置可用的按金额报价单")
		default:
			return model.CourseQuotation{}, errors.New("报读课程未配置可用的课时报价单")
		}
	}
	filtered := make([]model.CourseQuotation, 0, len(items))
	for _, item := range items {
		if item.LessonAudition == isTrial {
			filtered = append(filtered, item)
		}
	}
	if len(filtered) == 0 {
		filtered = items
	}
	return filtered[0], nil
}

func collectOrderImportColumnValues(rows []model.IntentionStudentImportRow, columns []model.IntentionStudentImportColumn, title string) []string {
	keySet := make(map[string]struct{})
	for _, column := range columns {
		if column.Title != title {
			continue
		}
		for _, row := range rows {
			for _, cell := range row.Cells {
				if cell.Key != column.Key {
					continue
				}
				text := strings.TrimSpace(cell.Value)
				if text == "" {
					continue
				}
				keySet[text] = struct{}{}
			}
		}
	}
	result := make([]string, 0, len(keySet))
	for item := range keySet {
		result = append(result, item)
	}
	sort.Strings(result)
	return result
}

func quantityOrDefault(value *int) int {
	if value == nil || *value <= 0 {
		return 1
	}
	return *value
}

func parseOrderImportLessonHour(value string, title string) (float64, error) {
	number, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0, fmt.Errorf("%s格式错误", title)
	}
	if number <= 0 {
		return 0, fmt.Errorf("%s必须大于0", title)
	}
	if !isValidTwoDecimalNumber(value) {
		return 0, fmt.Errorf("%s最多保留2位小数", title)
	}
	return number, nil
}

func parseOrderImportPositiveFloat(value string, title string) (float64, error) {
	number, err := parseOrderImportFloatWithPrecision(value, title, 2)
	if err != nil {
		return 0, err
	}
	if number <= 0 {
		return 0, fmt.Errorf("%s必须大于0", title)
	}
	return number, nil
}

func parseOrderImportFloat(value string, title string) (float64, error) {
	text := strings.TrimSpace(value)
	if text == "" {
		return 0, nil
	}
	number, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0, fmt.Errorf("%s格式错误", title)
	}
	if number < 0 {
		return 0, fmt.Errorf("%s不能小于0", title)
	}
	return number, nil
}

func parseOrderImportIntWithDefault(value string, title string) (float64, error) {
	text := strings.TrimSpace(value)
	if text == "" {
		return 0, nil
	}
	if !isValidIntegerNumber(text) {
		return 0, fmt.Errorf("%s请输入整数", title)
	}
	number, err := strconv.Atoi(text)
	if err != nil {
		return 0, fmt.Errorf("%s格式错误", title)
	}
	if number < 0 {
		return 0, fmt.Errorf("%s不能小于0", title)
	}
	return float64(number), nil
}

func parseOrderImportFloatWithPrecision(value string, title string, maxDecimals int) (float64, error) {
	number, err := parseOrderImportFloat(value, title)
	if err != nil {
		return 0, err
	}
	if strings.TrimSpace(value) == "" {
		return 0, nil
	}
	if maxDecimals == 2 && !isValidTwoDecimalNumber(value) {
		return 0, fmt.Errorf("%s最多保留2位小数", title)
	}
	return number, nil
}

func resolveOrderImportTagIDs(cell model.IntentionStudentImportCell, orderTagMap map[string]int64) []int64 {
	text := strings.TrimSpace(cell.Value)
	if text == "" {
		return nil
	}
	parts := strings.FieldsFunc(text, func(r rune) bool {
		return r == ',' || r == '，' || r == '、'
	})
	result := make([]int64, 0, len(parts))
	seen := make(map[int64]struct{})
	for _, item := range parts {
		name := strings.TrimSpace(item)
		if name == "" {
			continue
		}
		id, ok := orderTagMap[name]
		if !ok {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func resolvePayMethod(text string) int {
	switch strings.TrimSpace(text) {
	case "微信":
		return 1
	case "支付宝":
		return 2
	case "银行转账":
		return 3
	case "POS机":
		return 4
	case "现金":
		return 5
	default:
		return 6
	}
}

func summarizeOrderImportRows(rows []model.IntentionStudentImportRow) (int, int) {
	successCount := 0
	failCount := 0
	for _, row := range rows {
		switch row.Status {
		case 1:
			successCount++
		case 2:
			failCount++
		}
	}
	return successCount, failCount
}

func retryOrderImportTaskWrite(fn func() error) error {
	const maxAttempts = 3
	var err error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}
		if !isRetryableOrderImportTaskWriteError(err) || attempt == maxAttempts {
			return err
		}
		time.Sleep(time.Duration(attempt) * 200 * time.Millisecond)
	}
	return err
}

func isRetryableOrderImportTaskWriteError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "lock wait timeout exceeded") || strings.Contains(message, "deadlock found")
}

func (svc *Service) finishOrderImportTaskWithFatalError(detail *model.IntentionStudentImportTaskDetail, rows []model.IntentionStudentImportRow, message string) {
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
	_ = retryOrderImportTaskWrite(func() error {
		return svc.repo.UpdateOrderImportTask(context.Background(), *detail, rows)
	})
}

func clearOrderImportRowErrors(row *model.IntentionStudentImportRow) {
	for idx := range row.Cells {
		row.Cells[idx].Error = ""
	}
}

func attachOrderImportRowError(row *model.IntentionStudentImportRow, errText string) {
	if strings.TrimSpace(errText) == "" {
		return
	}
	hasCellError := false
	for _, cell := range row.Cells {
		if strings.TrimSpace(cell.Error) != "" {
			hasCellError = true
			break
		}
	}
	if hasCellError || len(row.Cells) == 0 {
		return
	}
	row.Cells[0].Error = errText
}

func cellValueByTitle(row model.IntentionStudentImportRow, title string) string {
	for _, cell := range row.Cells {
		if cell.Title == title {
			return cell.Value
		}
	}
	return ""
}

func intPtr(v int) *int {
	return &v
}
