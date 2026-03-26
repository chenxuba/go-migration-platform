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

func (svc *Service) UploadIntentionStudentImportFile(filename string, data []byte) (model.IntentionStudentImportUploadResult, error) {
	if len(data) == 0 {
		return model.IntentionStudentImportUploadResult{}, errors.New("empty file")
	}
	ticket := saveUploadedImportFile(uploadedImportFile{
		FileName:  strings.TrimSpace(filename),
		Data:      data,
		ExpiresAt: time.Now().Add(2 * time.Hour),
	})
	return model.IntentionStudentImportUploadResult{
		FileURL:  "/api/v1/intent-students/import-uploaded-file?ticket=" + ticket,
		FileName: strings.TrimSpace(filename),
	}, nil
}

func (svc *Service) LoadUploadedIntentionStudentImportFile(ticket string) (string, []byte, bool) {
	file, ok := loadUploadedImportFile(ticket)
	if !ok {
		return "", nil, false
	}
	return file.FileName, file.Data, true
}

func (svc *Service) SubmitIntentionStudentImportTask(userID int64, req model.IntentionStudentImportSubmitRequest) (string, error) {
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

	fileBytes, err := loadImportFileBytes(context.Background(), req.FileURL)
	if err != nil {
		return "", err
	}
	parseResult, err := svc.ParseIntentionStudentImportFile(userID, req.FileName, readerFromBytes(fileBytes))
	if err != nil {
		return "", err
	}

	now := time.Now()
	taskID := parseResult.ImportID
	task := intentionStudentImportTask{
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
	if err := svc.repo.CreateIntentionStudentImportTask(context.Background(), instID, task.Detail, task.Columns, task.Rows); err != nil {
		return "", err
	}
	return taskID, nil
}

func (svc *Service) ListIntentionStudentImportTasks(userID int64) (model.IntentionStudentImportTaskListResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.IntentionStudentImportTaskListResult{}, errors.New("no institution context")
		}
		return model.IntentionStudentImportTaskListResult{}, err
	}
	return svc.repo.ListIntentionStudentImportTasks(context.Background(), instID)
}

func (svc *Service) GetIntentionStudentImportTaskDetail(taskID string) (model.IntentionStudentImportTaskDetail, error) {
	task, err := svc.repo.GetIntentionStudentImportTask(context.Background(), taskID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.IntentionStudentImportTaskDetail{}, errors.New("import task not found")
		}
		return model.IntentionStudentImportTaskDetail{}, err
	}
	return task.Detail, nil
}

func (svc *Service) GetIntentionStudentImportTaskRecordList(taskID string, taskType int) (model.IntentionStudentImportTaskRecordListResult, error) {
	task, err := svc.repo.GetIntentionStudentImportTask(context.Background(), taskID)
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

func (svc *Service) BatchSaveIntentionStudentImportTaskRecords(userID int64, req model.IntentionStudentImportSaveTaskRecordRequest) ([]model.IntentionStudentImportRow, error) {
	task, err := svc.repo.GetIntentionStudentImportTask(context.Background(), req.TaskID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("import task not found")
		}
		return nil, err
	}
	optionMap, err := svc.loadIntentionStudentImportOptionMap(userID)
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
	task.Detail.Status = 3
	if err := svc.repo.UpdateIntentionStudentImportTask(context.Background(), task.Detail, task.Rows); err != nil {
		return nil, err
	}
	return updatedRows, nil
}

func (svc *Service) StartIntentionStudentImportTask(userID int64, taskID string) error {
	task, err := svc.repo.GetIntentionStudentImportTask(context.Background(), taskID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("import task not found")
		}
		return err
	}
	if countImportTaskErrors(task.Rows) > 0 {
		return errors.New("请先处理异常数据")
	}
	instUserID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		return err
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
	if err := svc.repo.UpdateIntentionStudentImportTask(context.Background(), task.Detail, task.Rows); err != nil {
		return err
	}
	go svc.runIntentionStudentImportTask(userID, taskID)
	return nil
}

func (svc *Service) runIntentionStudentImportTask(userID int64, taskID string) {
	task, err := svc.repo.GetIntentionStudentImportTask(context.Background(), taskID)
	if err != nil {
		return
	}
	columnMap := make(map[string]model.IntentionStudentImportColumn, len(task.Columns))
	for _, column := range task.Columns {
		columnMap[column.Key] = column
	}
	optionMap, err := svc.loadIntentionStudentImportOptionMap(userID)
	if err != nil {
		return
	}
	successCount := 0
	failCount := 0
	for idx := range task.Rows {
		row := task.Rows[idx]
		dto, err := buildStudentSaveDTOFromImportRow(row, columnMap, optionMap)
		if err != nil {
			task.Rows[idx].Status = 2
			task.Rows[idx].Result = err.Error()
			failCount++
		} else if _, err := svc.AddIntentStudentByImport(userID, dto); err != nil {
			task.Rows[idx].Status = 2
			task.Rows[idx].Result = err.Error()
			failCount++
		} else {
			task.Rows[idx].Status = 1
			task.Rows[idx].Result = "导入成功"
			successCount++
		}
		task.Detail.ExecutedRows = successCount
		task.Detail.ErrorRows = failCount
		task.Detail.Status = 4
		_ = svc.repo.UpdateIntentionStudentImportTask(context.Background(), task.Detail, task.Rows)
	}
	now := time.Now()
	task.Detail.CompleteTime = &now
	task.Detail.Status = 1
	_ = svc.repo.UpdateIntentionStudentImportTask(context.Background(), task.Detail, task.Rows)
}

func (svc *Service) AddIntentStudentByImport(userID int64, dto model.StudentSaveDTO) (int64, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("no institution context")
		}
		return 0, err
	}
	if strings.TrimSpace(dto.StuName) == "" || strings.TrimSpace(dto.Mobile) == "" {
		return 0, errors.New("stuName and mobile are required")
	}
	rule, count, err := svc.studentImportDuplicateCheck(context.Background(), instID, dto.StuName, dto.Mobile, nil)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, errors.New(studentDuplicateMessage(rule))
	}
	limitSameWeChat, err := svc.repo.GetLimitImportSameWeChat(context.Background(), instID)
	if err != nil {
		return 0, err
	}
	if limitSameWeChat && strings.TrimSpace(dto.WeChatNumber) != "" {
		count, err := svc.repo.CountStudentByWeChat(context.Background(), instID, dto.WeChatNumber, nil)
		if err != nil {
			return 0, err
		}
		if count > 0 {
			return 0, errors.New(studentWeChatDuplicateMessage())
		}
	}
	return svc.createIntentStudentRecord(userID, instID, dto)
}

func (svc *Service) ClearIntentionStudentImportTasks(userID int64) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	return svc.repo.ClearIntentionStudentImportTasks(context.Background(), instID)
}

func (svc *Service) DeleteIntentionStudentImportTask(userID int64, taskID string) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	if err := svc.repo.DeleteIntentionStudentImportTask(context.Background(), instID, taskID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("import task not found")
		}
		return err
	}
	return nil
}

type importOptionItem struct {
	Label string
	Value string
}

func (svc *Service) loadIntentionStudentImportOptionMap(userID int64) (map[string][]importOptionItem, error) {
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
	}

	gradeOptions := ""
	for _, field := range defaultFields {
		if strings.TrimSpace(field.FieldKey) == "年级" {
			gradeOptions = field.OptionsJSON
			break
		}
	}
	if strings.TrimSpace(gradeOptions) != "" {
		for _, item := range splitTemplateOptions(gradeOptions) {
			result["年级"] = append(result["年级"], importOptionItem{Label: item, Value: item})
		}
	}
	for _, channel := range channels {
		if channel.IsDisabled || strings.TrimSpace(channel.Name) == "" {
			continue
		}
		result["渠道"] = append(result["渠道"], importOptionItem{Label: channel.Name, Value: fmt.Sprintf("%d", channel.ID)})
	}
	for _, staff := range staffs {
		result["销售员"] = append(result["销售员"], importOptionItem{Label: staff.Name, Value: staff.ID})
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

func normalizeImportedCellValue(target *model.IntentionStudentImportCell, incoming model.IntentionStudentImportCell, column model.IntentionStudentImportColumn, options []importOptionItem) {
	target.SelectedID = normalizeImportSelectedID(incoming.SelectedID)
	target.Value = strings.TrimSpace(incoming.Value)
	if column.FieldType == 3 || strings.TrimSpace(column.Title) == "生日" {
		target.Value = normalizeImportDateText(target.Value)
	}
	if len(options) == 0 {
		return
	}
	selectedID := normalizeImportSelectedID(target.SelectedID)
	if selectedID != "" {
		for _, option := range options {
			if option.Value == selectedID {
				target.Value = option.Label
				target.SelectedID = selectedID
				return
			}
		}
	}
	for _, option := range options {
		if option.Label == target.Value {
			target.SelectedID = option.Value
			return
		}
	}
}

func validateImportedCell(column model.IntentionStudentImportColumn, cell model.IntentionStudentImportCell, options []importOptionItem) string {
	text := strings.TrimSpace(cell.Value)
	if column.Required && text == "" {
		return "请填写"
	}
	if text == "" {
		return ""
	}
	if len(options) > 0 {
		valid := false
		selectedID := normalizeImportSelectedID(cell.SelectedID)
		for _, option := range options {
			if option.Label == text || option.Value == selectedID {
				valid = true
				break
			}
		}
		if !valid {
			return "请选择预设值"
		}
	}
	if column.Title == "手机号" && !phoneDigitsPattern.MatchString(text) {
		return "手机号格式错误"
	}
	if requiresIntegerPrecision(column.Title) && !isValidIntegerNumber(text) {
		return "请输入整数"
	}
	if requiresTwoDecimalPrecision(column.Title) && !isValidTwoDecimalNumber(text) {
		return "最多保留2位小数"
	}
	switch column.FieldType {
	case 2:
		if !isNumericImportValue(text) {
			return "请输入数字"
		}
	case 3:
		if _, ok := parseImportDateValue(text); !ok {
			return "日期格式错误"
		}
	}
	return ""
}

func countImportTaskErrors(rows []model.IntentionStudentImportRow) int {
	count := 0
	for _, row := range rows {
		if row.HasError {
			count++
		}
	}
	return count
}

func countImportTaskImportFailures(rows []model.IntentionStudentImportRow) int {
	count := 0
	for _, row := range rows {
		if row.Status == 2 {
			count++
		}
	}
	return count
}

func buildStudentSaveDTOFromImportRow(row model.IntentionStudentImportRow, columns map[string]model.IntentionStudentImportColumn, optionMap map[string][]importOptionItem) (model.StudentSaveDTO, error) {
	dto := model.StudentSaveDTO{}
	defaultSex := 2
	dto.Sex = &defaultSex
	for _, cell := range row.Cells {
		column := columns[cell.Key]
		text := strings.TrimSpace(cell.Value)
		switch column.Title {
		case "学员姓名":
			dto.StuName = text
		case "手机号":
			dto.Mobile = text
		case "手机号归属人":
			if value, ok := resolveImportOptionInt(cell, optionMap[column.Title]); ok {
				dto.PhoneRelationship = &value
			}
		case "性别":
			if value, ok := resolveImportOptionInt(cell, optionMap[column.Title]); ok {
				dto.Sex = &value
			}
		case "生日":
			if text != "" {
				if parsed, ok := parseImportDateValue(text); ok {
					dto.Birthday = &parsed
				}
			}
		case "渠道":
			if value, ok := resolveImportOptionInt64(cell, optionMap[column.Title]); ok {
				dto.ChannelID = &value
			}
		case "微信号":
			dto.WeChatNumber = text
		case "年级":
			dto.Grade = text
		case "就读学校":
			dto.StudySchool = text
		case "家庭住址":
			dto.Address = text
		case "兴趣爱好":
			dto.Interest = text
		case "销售员", "销售":
			if value, ok := resolveImportOptionInt64(cell, optionMap[column.Title]); ok {
				dto.SalespersonID = &value
			}
		default:
			if column.FieldID > 0 {
				dto.CustomInfo = append(dto.CustomInfo, model.CustomInfo{
					FieldID:   column.FieldID,
					FieldName: column.Title,
					Value:     text,
				})
			}
		}
	}
	if dto.Sex != nil {
		dto.Avatar = defaultStudentAvatarBySex(*dto.Sex)
	}
	return dto, nil
}

func normalizeImportSelectedID(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case float64:
		return strconv.FormatInt(int64(typed), 10)
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	case nil:
		return ""
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", typed))
	}
}

func resolveImportOptionInt(cell model.IntentionStudentImportCell, options []importOptionItem) (int, bool) {
	if value, ok := resolveImportOptionInt64(cell, options); ok {
		return int(value), true
	}
	return 0, false
}

func resolveImportOptionInt64(cell model.IntentionStudentImportCell, options []importOptionItem) (int64, bool) {
	selectedID := normalizeImportSelectedID(cell.SelectedID)
	if selectedID != "" {
		if value, err := strconv.ParseInt(selectedID, 10, 64); err == nil {
			return value, true
		}
	}
	text := strings.TrimSpace(cell.Value)
	for _, option := range options {
		if option.Label == text {
			if value, err := strconv.ParseInt(option.Value, 10, 64); err == nil {
				return value, true
			}
		}
	}
	return 0, false
}

func defaultStudentAvatarBySex(sex int) string {
	switch sex {
	case 0:
		return "https://pcsys.admin.ybc365.com/d92afddc-ffac-40aa-aa61-bd97d91aa1ec.png"
	case 1:
		return "https://pcsys.admin.ybc365.com/c04d0ea2-a8b0-4001-b19b-946a980cb726.png"
	default:
		return "https://pcsys.admin.ybc365.com/a369a751-2be5-4929-974d-9ae4439f54c4.png"
	}
}
