package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"go-migration-platform/services/platform/internal/model"
	"go-migration-platform/services/platform/internal/repository"
)

func (svc *Service) UploadPlatformPEP3IEPMaterialImportFile(filename string, data []byte) (model.PEP3IEPMaterialImportUploadResult, error) {
	if len(data) == 0 {
		return model.PEP3IEPMaterialImportUploadResult{}, errors.New("empty file")
	}
	ticket := savePlatformImportUploadedFile(platformImportUploadedFile{
		FileName:  strings.TrimSpace(filename),
		Data:      data,
		ExpiresAt: time.Now().Add(2 * time.Hour),
	})
	return model.PEP3IEPMaterialImportUploadResult{
		FileURL:  "/api/v1/platform/scales/pep3-iep-material/import-uploaded-file?ticket=" + ticket,
		FileName: strings.TrimSpace(filename),
	}, nil
}

func (svc *Service) LoadUploadedPlatformPEP3IEPMaterialImportFile(ticket string) (string, []byte, bool) {
	file, ok := loadPlatformImportUploadedFile(ticket)
	if !ok {
		return "", nil, false
	}
	return file.FileName, file.Data, true
}

func (svc *Service) SubmitPlatformPEP3IEPMaterialImportTask(userID int64, username string, req model.PEP3IEPMaterialImportSubmitRequest) (string, error) {
	if svc.repo == nil {
		return "", errors.New("repository is not configured")
	}
	fileBytes, err := loadPlatformImportFileBytes(context.Background(), req.FileURL)
	if err != nil {
		return "", err
	}
	parseResult, err := svc.ParsePlatformPEP3IEPMaterialImportFile(req.FileName, platformReaderFromBytes(fileBytes))
	if err != nil {
		return "", err
	}

	now := time.Now()
	operatorName := strings.TrimSpace(username)
	if operatorName == "" {
		operatorName = "平台管理员"
	}
	taskID := parseResult.ImportID
	task := model.PEP3IEPMaterialImportTaskDetail{
		ID:              taskID,
		FileName:        strings.TrimSpace(req.FileName),
		UploadStaffID:   fmt.Sprintf("%d", userID),
		UploadStaffName: operatorName,
		TotalRows:       len(parseResult.Rows),
		ExecutedRows:    0,
		DeletedRows:     0,
		ErrorRows:       parseResult.AbnormalCount,
		CreatedTime:     &now,
		Status:          3,
		InstName:        parseResult.InstName,
	}
	if err := svc.repo.CreatePlatformPEP3IEPMaterialImportTask(context.Background(), task, parseResult.Columns, parseResult.Rows); err != nil {
		return "", err
	}
	return taskID, nil
}

func (svc *Service) ListPlatformPEP3IEPMaterialImportTasks() (model.PEP3IEPMaterialImportTaskListResult, error) {
	if svc.repo == nil {
		return model.PEP3IEPMaterialImportTaskListResult{}, errors.New("repository is not configured")
	}
	return svc.repo.ListPlatformPEP3IEPMaterialImportTasks(context.Background())
}

func (svc *Service) GetPlatformPEP3IEPMaterialImportTaskDetail(taskID string) (model.PEP3IEPMaterialImportTaskDetail, error) {
	if svc.repo == nil {
		return model.PEP3IEPMaterialImportTaskDetail{}, errors.New("repository is not configured")
	}
	task, err := svc.repo.GetPlatformPEP3IEPMaterialImportTask(context.Background(), strings.TrimSpace(taskID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PEP3IEPMaterialImportTaskDetail{}, errors.New("import task not found")
		}
		return model.PEP3IEPMaterialImportTaskDetail{}, err
	}
	return task.Detail, nil
}

func (svc *Service) GetPlatformPEP3IEPMaterialImportTaskRecordList(taskID string, taskType int) (model.PEP3IEPMaterialImportTaskRecordListResult, error) {
	if svc.repo == nil {
		return model.PEP3IEPMaterialImportTaskRecordListResult{}, errors.New("repository is not configured")
	}
	task, err := svc.repo.GetPlatformPEP3IEPMaterialImportTask(context.Background(), strings.TrimSpace(taskID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PEP3IEPMaterialImportTaskRecordListResult{}, errors.New("import task not found")
		}
		return model.PEP3IEPMaterialImportTaskRecordListResult{}, err
	}

	items := make([]model.PEP3IEPMaterialImportRow, 0, len(task.Rows))
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
	return model.PEP3IEPMaterialImportTaskRecordListResult{
		List:    items,
		Total:   len(items),
		Columns: task.Columns,
	}, nil
}

func (svc *Service) BatchSavePlatformPEP3IEPMaterialImportTaskRecords(req model.PEP3IEPMaterialImportSaveTaskRecordRequest) ([]model.PEP3IEPMaterialImportRow, error) {
	if svc.repo == nil {
		return nil, errors.New("repository is not configured")
	}
	task, err := svc.repo.GetPlatformPEP3IEPMaterialImportTask(context.Background(), strings.TrimSpace(req.TaskID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("import task not found")
		}
		return nil, err
	}
	lookups, err := svc.loadPEP3IEPImportLookups()
	if err != nil {
		return nil, err
	}

	currentByID := make(map[string]model.PEP3IEPMaterialImportRow, len(task.Rows))
	for _, row := range task.Rows {
		currentByID[row.ID] = row
	}
	updatedRows := make([]model.PEP3IEPMaterialImportRow, 0, len(req.Records))
	if len(req.Records) == 1 {
		incoming := req.Records[0]
		current, ok := currentByID[incoming.ID]
		if !ok {
			return nil, errors.New("import row not found")
		}
		current = mergePEP3IEPImportRowCells(current, incoming)
		current = normalizePEP3IEPImportRow(current, task.Columns, lookups)
		current.Status = 0
		current.Result = ""
		currentByID[current.ID] = current
		updatedRows = append(updatedRows, current)
		task.Rows = mapPEP3IEPImportRows(currentByID)
	} else {
		task.Rows = make([]model.PEP3IEPMaterialImportRow, 0, len(req.Records))
		for _, incoming := range req.Records {
			current, ok := currentByID[incoming.ID]
			if ok {
				incoming = mergePEP3IEPImportRowCells(current, incoming)
			}
			incoming = normalizePEP3IEPImportRow(incoming, task.Columns, lookups)
			incoming.Status = 0
			incoming.Result = ""
			task.Rows = append(task.Rows, incoming)
			updatedRows = append(updatedRows, incoming)
		}
		sort.Slice(task.Rows, func(i, j int) bool { return task.Rows[i].RowNo < task.Rows[j].RowNo })
	}

	task.Detail.TotalRows = len(task.Rows)
	task.Detail.ErrorRows = countPEP3IEPImportTaskErrors(task.Rows)
	task.Detail.ExecutedRows = 0
	task.Detail.Status = 3
	task.Detail.ConfirmTime = nil
	task.Detail.CompleteTime = nil
	if err := svc.repo.UpdatePlatformPEP3IEPMaterialImportTask(context.Background(), task.Detail, task.Rows); err != nil {
		return nil, err
	}
	return updatedRows, nil
}

func (svc *Service) StartPlatformPEP3IEPMaterialImportTask(userID int64, username, taskID string) error {
	if svc.repo == nil {
		return errors.New("repository is not configured")
	}
	task, err := svc.repo.GetPlatformPEP3IEPMaterialImportTask(context.Background(), strings.TrimSpace(taskID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("import task not found")
		}
		return err
	}
	if countPEP3IEPImportTaskErrors(task.Rows) > 0 {
		return errors.New("请先处理异常数据")
	}
	now := time.Now()
	executorName := strings.TrimSpace(username)
	if executorName == "" {
		executorName = "平台管理员"
	}
	executorID := fmt.Sprintf("%d", userID)
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
	if err := svc.repo.UpdatePlatformPEP3IEPMaterialImportTask(context.Background(), task.Detail, task.Rows); err != nil {
		return err
	}
	go svc.runPlatformPEP3IEPMaterialImportTask(userID, taskID)
	return nil
}

func (svc *Service) runPlatformPEP3IEPMaterialImportTask(userID int64, taskID string) {
	task, err := svc.repo.GetPlatformPEP3IEPMaterialImportTask(context.Background(), strings.TrimSpace(taskID))
	if err != nil {
		return
	}
	lookups, err := svc.loadPEP3IEPImportLookups()
	if err != nil {
		return
	}
	successCount := 0
	failCount := 0
	createdGroups := map[string]repository.PlatformPEP3IEPMaterialImportSaveResult{}
	for idx := range task.Rows {
		row := normalizePEP3IEPImportRow(task.Rows[idx], task.Columns, lookups)
		if row.HasError {
			task.Rows[idx] = row
			task.Rows[idx].Status = 2
			task.Rows[idx].Result = "数据校验失败"
			failCount++
		} else {
			input, err := buildPEP3IEPImportSaveInputFromRow(row, lookups)
			if err != nil {
				task.Rows[idx].Status = 2
				task.Rows[idx].Result = err.Error()
				failCount++
			} else {
				groupKey := pep3IEPImportAppendGroupKey(input)
				if existing, ok := createdGroups[groupKey]; ok {
					input.ExistingRuleID = existing.RuleID
					input.ExistingLongGoalID = existing.LongGoalID
				}
				saveResult, err := svc.repo.SavePlatformPEP3IEPMaterialImportRow(context.Background(), userID, input)
				if err != nil {
					task.Rows[idx].Status = 2
					task.Rows[idx].Result = err.Error()
					failCount++
				} else {
					createdGroups[groupKey] = saveResult
					task.Rows[idx].Status = 1
					task.Rows[idx].Result = "导入成功"
					successCount++
				}
			}
		}
		task.Detail.ExecutedRows = successCount
		task.Detail.ErrorRows = failCount
		task.Detail.Status = 4
		_ = svc.repo.UpdatePlatformPEP3IEPMaterialImportTask(context.Background(), task.Detail, task.Rows)
	}
	now := time.Now()
	task.Detail.CompleteTime = &now
	task.Detail.Status = 1
	_ = svc.repo.UpdatePlatformPEP3IEPMaterialImportTask(context.Background(), task.Detail, task.Rows)
}

func pep3IEPImportAppendGroupKey(input repository.PlatformPEP3IEPMaterialImportSaveInput) string {
	return fmt.Sprintf(
		"%s|%d|%d|%s",
		strings.TrimSpace(input.DomainCode),
		input.ItemNo,
		input.ScoreValue,
		normalizePEP3IEPImportMatchKey(input.LongGoal),
	)
}

func (svc *Service) ClearPlatformPEP3IEPMaterialImportTasks() error {
	if svc.repo == nil {
		return errors.New("repository is not configured")
	}
	return svc.repo.ClearPlatformPEP3IEPMaterialImportTasks(context.Background())
}

func (svc *Service) DeletePlatformPEP3IEPMaterialImportTask(taskID string) error {
	if svc.repo == nil {
		return errors.New("repository is not configured")
	}
	if err := svc.repo.DeletePlatformPEP3IEPMaterialImportTask(context.Background(), strings.TrimSpace(taskID)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("import task not found")
		}
		return err
	}
	return nil
}

func mergePEP3IEPImportRowCells(current, incoming model.PEP3IEPMaterialImportRow) model.PEP3IEPMaterialImportRow {
	cellByKey := make(map[string]model.PEP3IEPMaterialImportCell, len(incoming.Cells))
	for _, cell := range incoming.Cells {
		cellByKey[cell.Key] = cell
	}
	for idx := range current.Cells {
		if next, ok := cellByKey[current.Cells[idx].Key]; ok {
			current.Cells[idx].Value = strings.TrimSpace(next.Value)
			current.Cells[idx].SelectedID = next.SelectedID
		}
	}
	return current
}

func mapPEP3IEPImportRows(rowMap map[string]model.PEP3IEPMaterialImportRow) []model.PEP3IEPMaterialImportRow {
	rows := make([]model.PEP3IEPMaterialImportRow, 0, len(rowMap))
	for _, row := range rowMap {
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].RowNo < rows[j].RowNo })
	return rows
}

func countPEP3IEPImportTaskErrors(rows []model.PEP3IEPMaterialImportRow) int {
	count := 0
	for _, row := range rows {
		if row.HasError {
			count++
		}
	}
	return count
}

func buildPEP3IEPImportSaveInputFromRow(row model.PEP3IEPMaterialImportRow, lookups pep3IEPImportLookups) (repository.PlatformPEP3IEPMaterialImportSaveInput, error) {
	rowData := pep3IEPImportRowValueMap(row)
	domain, ok := lookups.DomainByName[strings.TrimSpace(rowData["领域"])]
	if !ok {
		return repository.PlatformPEP3IEPMaterialImportSaveInput{}, errors.New("领域不存在")
	}
	question, ok := lookups.QuestionByName[strings.TrimSpace(rowData["题目"])]
	if !ok {
		return repository.PlatformPEP3IEPMaterialImportSaveInput{}, errors.New("题目不存在")
	}
	score, ok := parsePEP3IEPScoreValue(rowData["选项"])
	if !ok {
		return repository.PlatformPEP3IEPMaterialImportSaveInput{}, errors.New("选项不正确")
	}
	scoreLabel := fmt.Sprintf("%d分", score)
	scoreDescription := ""
	for _, option := range question.ScoreOptions {
		if option.Value == score {
			if strings.TrimSpace(option.Label) != "" {
				scoreLabel = strings.TrimSpace(option.Label)
			}
			scoreDescription = strings.TrimSpace(option.Description)
			break
		}
	}
	return repository.PlatformPEP3IEPMaterialImportSaveInput{
		ItemNo:           question.ItemNo,
		ItemTitle:        pep3IEPImportQuestionTitle(question),
		DomainCode:       domain.Code,
		Domain:           domain.Name,
		ScoreValue:       score,
		ScoreLabel:       scoreLabel,
		ScoreDescription: scoreDescription,
		LongGoal:         normalizePEP3IEPImportDisplayText(rowData["长期目标"]),
		ShortGoal:        normalizePEP3IEPImportDisplayText(rowData["短期目标"]),
		CourseForm:       strings.TrimSpace(rowData["课程形式"]),
		TrainingProject:  normalizePEP3IEPImportDisplayText(rowData["训练项目"]),
		TrainingContent:  normalizePEP3IEPImportDisplayText(rowData["训练内容"]),
		Status:           pep3IEPImportStatusValue(rowData["状态"]),
	}, nil
}
