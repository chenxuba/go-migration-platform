package service

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) BatchCreateHomeworks(userID int64, dto model.HomeworkBatchCreateDTO) ([]model.HomeworkOperationResult, error) {
	instID, operatorID, err := svc.resolveTeachingClassOperator(userID)
	if err != nil {
		return nil, err
	}
	inputs, err := svc.buildHomeworkMutationInputs(
		context.Background(),
		instID,
		dto.Title,
		dto.Content,
		dto.Attachments,
		dto.RepeatRule,
		dto.PublishTime,
		dto.EndTime,
		dto.PublishHour,
		dto.EndHour,
		dto.IsVisibleStudent,
		dto.HomeworkObjects,
		true,
	)
	if err != nil {
		return nil, err
	}
	return svc.repo.BatchCreateHomeworks(context.Background(), instID, operatorID, inputs)
}

func (svc *Service) UpdateHomework(userID int64, dto model.HomeworkUpdateDTO) (model.HomeworkOperationResult, error) {
	instID, operatorID, err := svc.resolveTeachingClassOperator(userID)
	if err != nil {
		return model.HomeworkOperationResult{}, err
	}
	homeworkID, err := parseRequiredInt64String(dto.ID, "课后任务ID")
	if err != nil {
		return model.HomeworkOperationResult{}, err
	}
	inputs, err := svc.buildHomeworkMutationInputs(
		context.Background(),
		instID,
		dto.Title,
		dto.Content,
		dto.Attachments,
		dto.RepeatRule,
		dto.PublishTime,
		dto.EndTime,
		dto.PublishHour,
		dto.EndHour,
		dto.IsVisibleStudent,
		dto.HomeworkObjects,
		false,
	)
	if err != nil {
		return model.HomeworkOperationResult{}, err
	}
	if len(inputs) != 1 {
		return model.HomeworkOperationResult{}, errors.New("编辑时仅支持一个班级或1对1对象")
	}
	if err := svc.repo.UpdateHomework(context.Background(), instID, operatorID, homeworkID, inputs[0]); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.HomeworkOperationResult{}, errors.New("课后任务不存在")
		}
		return model.HomeworkOperationResult{}, err
	}
	return model.HomeworkOperationResult{
		ID:   strconv.FormatInt(homeworkID, 10),
		Name: strings.TrimSpace(dto.Title),
	}, nil
}

func (svc *Service) DeleteHomework(userID int64, id string) error {
	instID, operatorID, err := svc.resolveTeachingClassOperator(userID)
	if err != nil {
		return err
	}
	homeworkID, err := parseRequiredInt64String(id, "课后任务ID")
	if err != nil {
		return err
	}
	if err := svc.repo.DeleteHomework(context.Background(), instID, operatorID, homeworkID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("课后任务不存在")
		}
		return err
	}
	return nil
}

func (svc *Service) GetHomeworkDetail(userID int64, id string) (model.HomeworkDetailVO, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.HomeworkDetailVO{}, errors.New("no institution context")
		}
		return model.HomeworkDetailVO{}, err
	}
	homeworkID, err := parseRequiredInt64String(id, "课后任务ID")
	if err != nil {
		return model.HomeworkDetailVO{}, err
	}
	item, err := svc.repo.GetHomeworkDetail(context.Background(), instID, homeworkID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.HomeworkDetailVO{}, errors.New("课后任务不存在")
		}
		return model.HomeworkDetailVO{}, err
	}
	return item, nil
}

func (svc *Service) PageHomeworks(userID int64, query model.HomeworkListQueryDTO) (model.HomeworkPageResultVO, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.HomeworkPageResultVO{}, errors.New("no institution context")
		}
		return model.HomeworkPageResultVO{}, err
	}
	return svc.repo.PageHomeworks(context.Background(), instID, query)
}

func (svc *Service) HomeworkStatistics(userID int64, query model.HomeworkListQueryModel) (model.HomeworkStatisticsVO, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.HomeworkStatisticsVO{}, errors.New("no institution context")
		}
		return model.HomeworkStatisticsVO{}, err
	}
	return svc.repo.HomeworkStatistics(context.Background(), instID, query)
}

func (svc *Service) buildHomeworkMutationInputs(ctx context.Context, instID int64, title, content string, attachments []model.HomeworkAttachment, repeatRule *model.HomeworkRepeatRule, publishTimeText, endTimeText string, publishHour, endHour int, isVisibleStudent bool, objects []model.HomeworkObjectDTO, fillImmediatePublishTime bool) ([]model.HomeworkMutationInput, error) {
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)
	if title == "" {
		return nil, errors.New("任务标题不能为空")
	}
	if content == "" {
		return nil, errors.New("任务内容不能为空")
	}
	if len(objects) == 0 {
		return nil, errors.New("请选择班级/学员")
	}

	publishRule, normalizedRepeatRule, publishTime, endTime, err := buildHomeworkSchedule(repeatRule, publishTimeText, endTimeText, publishHour, endHour, fillImmediatePublishTime)
	if err != nil {
		return nil, err
	}
	normalizedAttachments := normalizeHomeworkAttachments(attachments)

	inputs := make([]model.HomeworkMutationInput, 0, len(objects))
	for _, object := range objects {
		sourceID, err := parseRequiredInt64String(object.SourceID, "班级/1对1")
		if err != nil {
			return nil, err
		}
		sourceType := object.SourceType
		if sourceType != model.HomeworkSourceTypeClass && sourceType != model.HomeworkSourceTypeOneToOne {
			return nil, errors.New("班级/1对1类型无效")
		}

		sourceName, sourceStudents, err := svc.repo.ListHomeworkTargetStudents(ctx, instID, sourceType, sourceID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, errors.New("班级/1对1不存在或暂无可选学员")
			}
			return nil, err
		}

		selectedStudents, err := buildHomeworkSelectedStudents(sourceType, sourceID, sourceName, sourceStudents, object.StudentIDs)
		if err != nil {
			return nil, err
		}

		rowVisibleStudent := isVisibleStudent
		if sourceType == model.HomeworkSourceTypeOneToOne {
			rowVisibleStudent = false
		}

		inputs = append(inputs, model.HomeworkMutationInput{
			Title:            title,
			Content:          content,
			Attachments:      normalizedAttachments,
			RepeatRule:       normalizedRepeatRule,
			PublishRule:      publishRule,
			PublishTime:      publishTime,
			EndTime:          endTime,
			PublishHour:      publishHour,
			EndHour:          endHour,
			IsVisibleStudent: rowVisibleStudent,
			SourceType:       sourceType,
			SourceID:         sourceID,
			SourceName:       sourceName,
			SelectedStudents: selectedStudents,
		})
	}
	return inputs, nil
}

func normalizeHomeworkAttachments(attachments []model.HomeworkAttachment) []model.HomeworkAttachment {
	result := make([]model.HomeworkAttachment, 0, len(attachments))
	for _, attachment := range attachments {
		url := strings.TrimSpace(attachment.URL)
		if url == "" {
			continue
		}
		item := model.HomeworkAttachment{
			Type:       attachment.Type,
			URL:        url,
			Duration:   attachment.Duration,
			Name:       strings.TrimSpace(attachment.Name),
			ExtendName: strings.TrimSpace(attachment.ExtendName),
		}
		if item.Type != model.HomeworkAttachmentTypeVideo {
			item.Type = model.HomeworkAttachmentTypeImage
		}
		result = append(result, item)
	}
	return result
}

func buildHomeworkSelectedStudents(sourceType int, sourceID int64, sourceName string, sourceStudents []model.HomeworkTargetStudent, selectedIDs []string) ([]model.HomeworkSelectedStudent, error) {
	studentMap := make(map[string]model.HomeworkTargetStudent, len(sourceStudents))
	orderedIDs := make([]string, 0, len(sourceStudents))
	for _, item := range sourceStudents {
		studentID := strconv.FormatInt(item.StudentID, 10)
		studentMap[studentID] = item
		orderedIDs = append(orderedIDs, studentID)
	}

	selectedSet := make(map[string]struct{}, len(selectedIDs))
	for _, value := range selectedIDs {
		text := strings.TrimSpace(value)
		if text == "" {
			continue
		}
		selectedSet[text] = struct{}{}
	}
	if len(selectedSet) == 0 {
		return nil, errors.New("请选择学员")
	}

	selectedStudents := make([]model.HomeworkSelectedStudent, 0, len(selectedSet))
	for _, studentID := range orderedIDs {
		if _, ok := selectedSet[studentID]; !ok {
			continue
		}
		student := studentMap[studentID]
		selectedStudents = append(selectedStudents, model.HomeworkSelectedStudent{
			SourceType:       sourceType,
			SourceID:         strconv.FormatInt(sourceID, 10),
			SourceName:       strings.TrimSpace(sourceName),
			StudentID:        studentID,
			StudentName:      strings.TrimSpace(student.StudentName),
			TuitionAccountID: strings.TrimSpace(student.TuitionAccountID),
			IsBind:           student.IsBind,
		})
	}
	if len(selectedStudents) != len(selectedSet) {
		return nil, errors.New("所选学员不在当前班级/1对1内")
	}
	if sourceType == model.HomeworkSourceTypeOneToOne && len(selectedStudents) > 1 {
		return nil, errors.New("1对1任务仅支持一个学员")
	}

	sort.SliceStable(selectedStudents, func(i, j int) bool {
		return selectedStudents[i].StudentID < selectedStudents[j].StudentID
	})
	return selectedStudents, nil
}

func buildHomeworkSchedule(repeatRule *model.HomeworkRepeatRule, publishTimeText, endTimeText string, publishHour, endHour int, fillImmediatePublishTime bool) (int, *model.HomeworkRepeatRule, *time.Time, *time.Time, error) {
	if repeatRule != nil {
		startDate, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(repeatRule.StartDate), time.Local)
		if err != nil {
			return 0, nil, nil, nil, errors.New("自动任务开始日期无效")
		}
		finishDate, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(repeatRule.EndDate), time.Local)
		if err != nil {
			return 0, nil, nil, nil, errors.New("自动任务结束日期无效")
		}
		if finishDate.Before(startDate) {
			return 0, nil, nil, nil, errors.New("自动任务结束日期不能早于开始日期")
		}
		if publishHour < 0 || publishHour > 23 {
			return 0, nil, nil, nil, errors.New("自动任务推送时间无效")
		}
		if endHour < 0 || endHour > 23 {
			return 0, nil, nil, nil, errors.New("自动任务截止时间无效")
		}
		if repeatRule.RepeatSpan <= 0 {
			repeatRule.RepeatSpan = 1
		}
		if repeatRule.WeekDays <= 0 {
			return 0, nil, nil, nil, errors.New("请选择自动任务周期")
		}

		publishTime := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), publishHour, 0, 0, 0, time.Local)
		endTime := time.Date(finishDate.Year(), finishDate.Month(), finishDate.Day(), endHour, 0, 0, 0, time.Local)
		if endTime.Before(publishTime) {
			endTime = publishTime
		}

		normalized := &model.HomeworkRepeatRule{
			StartDate:  strings.TrimSpace(repeatRule.StartDate),
			EndDate:    strings.TrimSpace(repeatRule.EndDate),
			RepeatSpan: repeatRule.RepeatSpan,
			WeekDays:   repeatRule.WeekDays,
		}
		return model.HomeworkPublishRuleAuto, normalized, &publishTime, &endTime, nil
	}

	var publishTime *time.Time
	if strings.TrimSpace(publishTimeText) != "" {
		parsed, err := parseHomeworkDateTime(strings.TrimSpace(publishTimeText))
		if err != nil {
			return 0, nil, nil, nil, errors.New("发布时间无效")
		}
		publishTime = &parsed
	} else if fillImmediatePublishTime {
		now := time.Now()
		publishTime = &now
	}

	var endTime *time.Time
	if strings.TrimSpace(endTimeText) != "" {
		parsed, err := parseHomeworkDateTime(strings.TrimSpace(endTimeText))
		if err != nil {
			return 0, nil, nil, nil, errors.New("截止时间无效")
		}
		endTime = &parsed
	}

	if publishTime != nil && endTime != nil && !endTime.After(*publishTime) {
		return 0, nil, nil, nil, errors.New("截止时间需晚于发布时间")
	}

	return model.HomeworkPublishRuleOnce, nil, publishTime, endTime, nil
}

func parseHomeworkDateTime(value string) (time.Time, error) {
	for _, layout := range []string{
		"2006-01-02T15:04",
		"2006-01-02 15:04",
		time.RFC3339,
	} {
		if parsed, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, errors.New("invalid datetime")
}
