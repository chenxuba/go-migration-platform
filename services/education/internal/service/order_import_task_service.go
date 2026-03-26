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
	parseResult, err := svc.ParseLessonHourOrderImportFile(userID, req.FileName, readerFromBytes(fileBytes))
	if err != nil {
		return "", err
	}

	now := time.Now()
	taskID := parseResult.ImportID
	task := orderImportTask{
		InstID: instID,
		Detail: model.IntentionStudentImportTaskDetail{
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
		},
		Columns: parseResult.Columns,
		Rows:    parseResult.Rows,
	}
	saveOrderImportTask(task)
	_ = instID
	return taskID, nil
}

func (svc *Service) GetOrderImportTaskDetail(taskID string) (model.IntentionStudentImportTaskDetail, error) {
	task, ok := loadOrderImportTask(taskID)
	if !ok {
		return model.IntentionStudentImportTaskDetail{}, errors.New("import task not found")
	}
	return task.Detail, nil
}

func (svc *Service) GetOrderImportTaskRecordList(taskID string, taskType int) (model.IntentionStudentImportTaskRecordListResult, error) {
	task, ok := loadOrderImportTask(taskID)
	if !ok {
		return model.IntentionStudentImportTaskRecordListResult{}, errors.New("import task not found")
	}

	items := make([]model.IntentionStudentImportRow, 0, len(task.Rows))
	for _, row := range task.Rows {
		if taskType == 0 && row.HasError {
			items = append(items, row)
			continue
		}
		if taskType == 1 && !row.HasError {
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
	task, ok := loadOrderImportTask(req.TaskID)
	if !ok {
		return nil, errors.New("import task not found")
	}
	optionMap, err := svc.loadOrderImportOptionMap(userID)
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
		current.HasError = false
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
	saveOrderImportTask(task)
	return updatedRows, nil
}

func (svc *Service) DeleteOrderImportTask(taskID string) error {
	if _, ok := loadOrderImportTask(taskID); !ok {
		return errors.New("import task not found")
	}
	deleteOrderImportTask(taskID)
	return nil
}

func (svc *Service) StartOrderImportTask(userID int64, taskID string) (model.OrderImportStartResult, error) {
	task, ok := loadOrderImportTask(taskID)
	if !ok {
		return model.OrderImportStartResult{}, errors.New("import task not found")
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
	saveOrderImportTask(task)
	go svc.runOrderImportTask(userID, taskID)
	return model.OrderImportStartResult{}, nil
}

func (svc *Service) runOrderImportTask(userID int64, taskID string) {
	task, ok := loadOrderImportTask(taskID)
	if !ok {
		return
	}
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		return
	}
	optionMap, err := svc.loadOrderImportOptionMap(userID)
	if err != nil {
		return
	}
	orderTagMap, err := svc.repo.ListEnabledOrderTagNameIDMap(context.Background(), instID)
	if err != nil {
		return
	}
	courseNames := collectOrderImportColumnValues(task.Rows, task.Columns, "报读课程")
	quotationMap, err := svc.repo.ListCourseQuotationsByNamesAndLessonModel(context.Background(), instID, courseNames, 1)
	if err != nil {
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
		if err := svc.importOrderRow(userID, instID, row, columnMap, optionMap, orderTagMap, quotationMap); err != nil {
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
		saveOrderImportTask(task)
	}

	completeAt := time.Now()
	task.Detail.CompleteTime = &completeAt
	task.Detail.Status = 1
	saveOrderImportTask(task)
}

func (svc *Service) ListOrderImportTasks(userID int64) (model.IntentionStudentImportTaskListResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.IntentionStudentImportTaskListResult{}, errors.New("no institution context")
		}
		return model.IntentionStudentImportTaskListResult{}, err
	}
	tasks := listOrderImportTasks(instID)
	sort.Slice(tasks, func(i, j int) bool {
		left := tasks[i].Detail.CreatedTime
		right := tasks[j].Detail.CreatedTime
		if left == nil || right == nil {
			return tasks[i].Detail.ID > tasks[j].Detail.ID
		}
		return left.After(*right)
	})
	items := make([]model.IntentionStudentImportTaskDetail, 0, len(tasks))
	for _, task := range tasks {
		items = append(items, task.Detail)
	}
	return model.IntentionStudentImportTaskListResult{List: items, Total: len(items)}, nil
}

func (svc *Service) ClearOrderImportTasks(userID int64) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	clearOrderImportTasks(instID)
	return nil
}

func (svc *Service) loadOrderImportOptionMap(userID int64) (map[string][]importOptionItem, error) {
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
	courseNames, err := svc.repo.ListCourseNames(context.Background(), instID)
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

	createDTO, payDTO, hasPayment, err := buildCreateAndPayOrderDTOFromImportRow(decision.StudentID, row, columns, optionMap, orderTagMap, quotationMap)
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
		count, err := svc.repo.CountStudentByWeChat(context.Background(), instID, dto.WeChatNumber)
		if err != nil {
			return orderImportStudentDecision{}, err
		}
		if count > 0 {
			return orderImportStudentDecision{}, errors.New("已存在相同微信号的学员，不可创建")
		}
	}

	studentID, err := svc.AddIntentStudent(userID, dto)
	if err != nil {
		return orderImportStudentDecision{}, err
	}
	return orderImportStudentDecision{StudentID: studentID, CreatedNew: true}, nil
}

func buildCreateAndPayOrderDTOFromImportRow(
	studentID int64,
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
	purchaseCount, err := parseOrderImportInt(rowData["购买课时数"].Value, "购买课时数")
	if err != nil {
		return model.CreateOrderDTO{}, model.PayOrderDTO{}, false, err
	}
	giftCount, err := parseOrderImportFloat(rowData["赠送课时数"].Value, "赠送课时数")
	if err != nil {
		return model.CreateOrderDTO{}, model.PayOrderDTO{}, false, err
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
	quotation, count, err := pickLessonHourQuotation(quotationMap[courseName], purchaseCount, isTrial)
	if err != nil {
		return model.CreateOrderDTO{}, model.PayOrderDTO{}, false, err
	}

	var (
		dealDate   *time.Time
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
	if value := strings.TrimSpace(rowData["有效期至"].Value); value != "" {
		if parsed, ok := parseImportDateValue(value); ok {
			endDate = &parsed
		}
	}

	if value, ok := resolveImportOptionInt64(rowData["订单销售员"], optionMap["订单销售员"]); ok {
		salePerson = &value
	} else if value, ok := resolveImportOptionInt64(rowData["销售"], optionMap["销售"]); ok {
		salePerson = &value
	}

	orderTagIDs := resolveOrderImportTagIDs(rowData["订单标签"], orderTagMap)
	handleType := 0
	unit := 1
	if quotation.Unit != nil {
		unit = *quotation.Unit
	}
	lessonMode := 1
	if quotation.LessonModel != nil {
		lessonMode = *quotation.LessonModel
	}

	countValue := count
	quoteID := quotation.ID
	courseID := quotation.CourseID
	hasValidDate := endDate != nil
	courseAmount := fmt.Sprintf("%.2f", orderAmount)
	realQuantity := float64(purchaseCount) + giftCount

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
					EndDate:      endDate,
					Amount:       courseAmount,
					Quantity:     float64(quantityOrDefault(quotation.Quantity)),
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

func pickLessonHourQuotation(items []model.CourseQuotation, purchaseCount int, isTrial bool) (model.CourseQuotation, int, error) {
	if len(items) == 0 {
		return model.CourseQuotation{}, 0, errors.New("报读课程未找到匹配的课时报价单")
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

	bestIdx := -1
	bestQty := 0
	bestCount := 0
	for idx, item := range filtered {
		qty := quantityOrDefault(item.Quantity)
		if qty <= 0 || purchaseCount%qty != 0 {
			continue
		}
		count := purchaseCount / qty
		if qty > bestQty || (qty == bestQty && (bestIdx == -1 || count < bestCount)) {
			bestIdx = idx
			bestQty = qty
			bestCount = count
		}
	}
	if bestIdx == -1 {
		return model.CourseQuotation{}, 0, errors.New("购买课时数无法匹配当前课程的报价单规格")
	}
	return filtered[bestIdx], bestCount, nil
}

func quantityOrDefault(value *int) int {
	if value == nil || *value <= 0 {
		return 1
	}
	return *value
}

func parseOrderImportInt(value string, title string) (int, error) {
	number, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0, fmt.Errorf("%s格式错误", title)
	}
	if number <= 0 {
		return 0, fmt.Errorf("%s必须大于0", title)
	}
	if number != float64(int(number)) {
		return 0, fmt.Errorf("%s必须为整数", title)
	}
	return int(number), nil
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
