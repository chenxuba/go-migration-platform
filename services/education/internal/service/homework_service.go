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
		dto.TaskDurationHours,
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
		dto.TaskDurationHours,
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

func (svc *Service) buildHomeworkMutationInputs(ctx context.Context, instID int64, title, content string, attachments []model.HomeworkAttachment, repeatRule *model.HomeworkRepeatRule, publishTimeText, endTimeText string, publishHour, endHour, taskDurationHours int, isVisibleStudent bool, objects []model.HomeworkObjectDTO, fillImmediatePublishTime bool) ([]model.HomeworkMutationInput, error) {
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

	schedulePlans, err := buildHomeworkSchedules(repeatRule, publishTimeText, endTimeText, publishHour, endHour, taskDurationHours, fillImmediatePublishTime)
	if err != nil {
		return nil, err
	}
	normalizedAttachments := normalizeHomeworkAttachments(attachments)

	inputs := make([]model.HomeworkMutationInput, 0, len(objects)*len(schedulePlans))
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

		for _, schedulePlan := range schedulePlans {
			inputs = append(inputs, model.HomeworkMutationInput{
				Title:             title,
				Content:           content,
				Attachments:       normalizedAttachments,
				RepeatRule:        schedulePlan.RepeatRule,
				PublishRule:       schedulePlan.PublishRule,
				PublishTime:       schedulePlan.PublishTime,
				EndTime:           schedulePlan.EndTime,
				PublishHour:       schedulePlan.PublishHour,
				EndHour:           schedulePlan.EndHour,
				TaskDurationHours: schedulePlan.TaskDurationHours,
				IsVisibleStudent:  rowVisibleStudent,
				SourceType:        sourceType,
				SourceID:          sourceID,
				SourceName:        sourceName,
				SelectedStudents:  selectedStudents,
			})
		}
	}
	return inputs, nil
}

type homeworkSchedulePlan struct {
	PublishRule       int
	RepeatRule        *model.HomeworkRepeatRule
	PublishTime       *time.Time
	EndTime           *time.Time
	PublishHour       int
	EndHour           int
	TaskDurationHours int
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

func buildHomeworkSchedules(repeatRule *model.HomeworkRepeatRule, publishTimeText, endTimeText string, publishHour, endHour, taskDurationHours int, fillImmediatePublishTime bool) ([]homeworkSchedulePlan, error) {
	if repeatRule != nil {
		startDate, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(repeatRule.StartDate), time.Local)
		if err != nil {
			return nil, errors.New("自动任务开始日期无效")
		}
		finishDate, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(repeatRule.EndDate), time.Local)
		if err != nil {
			return nil, errors.New("自动任务结束日期无效")
		}
		if finishDate.Before(startDate) {
			return nil, errors.New("自动任务结束日期不能早于开始日期")
		}
		if publishHour < 0 || publishHour > 23 {
			return nil, errors.New("自动任务推送时间无效")
		}
		if taskDurationHours < 0 {
			return nil, errors.New("单次任务时长无效")
		}
		if taskDurationHours <= 0 && endHour > 0 {
			if endHour > publishHour {
				taskDurationHours = endHour - publishHour
			}
		}
		if repeatRule.RepeatSpan <= 0 {
			repeatRule.RepeatSpan = 1
		}
		if repeatRule.WeekDays <= 0 {
			return nil, errors.New("请选择自动任务周期")
		}

		occurrenceDates := enumerateHomeworkOccurrenceDates(startDate, finishDate, repeatRule.RepeatSpan, repeatRule.WeekDays)
		if len(occurrenceDates) == 0 {
			return nil, errors.New("任务日期范围内没有符合周期的任务")
		}

		plans := make([]homeworkSchedulePlan, 0, len(occurrenceDates))
		for _, occurrenceDate := range occurrenceDates {
			publishTime := time.Date(occurrenceDate.Year(), occurrenceDate.Month(), occurrenceDate.Day(), publishHour, 0, 0, 0, time.Local)
			var currentEndTime *time.Time
			normalizedEndHour := 0
			if taskDurationHours > 0 {
				deadline := publishTime.Add(time.Duration(taskDurationHours) * time.Hour)
				currentEndTime = &deadline
				normalizedEndHour = publishHour + taskDurationHours
			}
			plans = append(plans, homeworkSchedulePlan{
				PublishRule:       model.HomeworkPublishRuleOnce,
				RepeatRule:        nil,
				PublishTime:       &publishTime,
				EndTime:           currentEndTime,
				PublishHour:       publishHour,
				EndHour:           normalizedEndHour,
				TaskDurationHours: taskDurationHours,
			})
		}
		return plans, nil
	}

	var publishTime *time.Time
	if strings.TrimSpace(publishTimeText) != "" {
		parsed, err := parseHomeworkDateTime(strings.TrimSpace(publishTimeText))
		if err != nil {
			return nil, errors.New("发布时间无效")
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
			return nil, errors.New("截止时间无效")
		}
		endTime = &parsed
	}

	if publishTime != nil && endTime != nil && !endTime.After(*publishTime) {
		return nil, errors.New("截止时间需晚于发布时间")
	}

	return []homeworkSchedulePlan{
		{
			PublishRule: model.HomeworkPublishRuleOnce,
			PublishTime: publishTime,
			EndTime:     endTime,
		},
	}, nil
}

func enumerateHomeworkOccurrenceDates(startDate, finishDate time.Time, repeatSpan, weekDays int) []time.Time {
	if repeatSpan <= 0 || weekDays <= 0 {
		return nil
	}

	normalizedStart := normalizeHomeworkDate(startDate)
	normalizedFinish := normalizeHomeworkDate(finishDate)
	anchorWeekStart := homeworkWeekStart(normalizedStart)

	result := make([]time.Time, 0)
	for current := normalizedStart; !current.After(normalizedFinish); current = current.AddDate(0, 0, 1) {
		homeworkWeekday := timeWeekdayToHomeworkWeekday(current.Weekday())
		if !homeworkWeekdaySelected(weekDays, homeworkWeekday) {
			continue
		}
		currentWeekStart := homeworkWeekStart(current)
		weekIndex := int(currentWeekStart.Sub(anchorWeekStart) / (7 * 24 * time.Hour))
		if weekIndex < 0 || weekIndex%repeatSpan != 0 {
			continue
		}
		result = append(result, current)
	}
	return result
}

func normalizeHomeworkDate(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.Local)
}

func homeworkWeekStart(value time.Time) time.Time {
	normalized := normalizeHomeworkDate(value)
	offset := 0
	if normalized.Weekday() == time.Sunday {
		offset = 6
	} else {
		offset = int(normalized.Weekday()) - 1
	}
	return normalized.AddDate(0, 0, -offset)
}

func timeWeekdayToHomeworkWeekday(weekday time.Weekday) int {
	if weekday == time.Sunday {
		return 7
	}
	return int(weekday)
}

func homeworkWeekdaySelected(weekDays, weekday int) bool {
	switch weekday {
	case 1:
		return weekDays&2 == 2
	case 2:
		return weekDays&4 == 4
	case 3:
		return weekDays&8 == 8
	case 4:
		return weekDays&16 == 16
	case 5:
		return weekDays&32 == 32
	case 6:
		return weekDays&64 == 64
	case 7:
		return weekDays&1 == 1
	default:
		return false
	}
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
