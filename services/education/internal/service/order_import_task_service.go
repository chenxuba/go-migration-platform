package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

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
