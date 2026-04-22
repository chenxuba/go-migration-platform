package service

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

const (
	parentRehabRecordDefaultPageIndex = 1
	parentRehabRecordDefaultPageSize  = 20
	parentRehabRecordMaxPageSize      = 100
)

func (svc *Service) ListParentRehabRecordsByPhone(ctx context.Context, phone string, query model.ParentRehabRecordQueryDTO) (model.ParentRehabRecordSummaryVO, error) {
	if svc == nil || svc.repo == nil {
		return model.ParentRehabRecordSummaryVO{}, errors.New("家长端康复记录服务未初始化")
	}

	phone = normalizeParentPhone(phone)
	if phone == "" {
		return model.ParentRehabRecordSummaryVO{}, errors.New("手机号不能为空")
	}

	rows, err := svc.repo.ListParentStudentCandidatesByPhone(ctx, phone)
	if err != nil {
		return model.ParentRehabRecordSummaryVO{}, err
	}
	displayProfiles, err := svc.resolveParentStudentDisplayProfiles(ctx, rows)
	if err != nil {
		return model.ParentRehabRecordSummaryVO{}, err
	}

	targets := buildParentScheduleTargets(rows, displayProfiles)
	students := buildParentBoundStudents(targets)
	if len(targets) == 0 {
		return model.ParentRehabRecordSummaryVO{
			Students: students,
			Items:    []model.ParentRehabRecordVO{},
		}, nil
	}

	target, ok := resolveParentRehabTarget(targets, query.StudentID)
	if !ok {
		return model.ParentRehabRecordSummaryVO{}, errors.New("未找到对应学员")
	}

	pageIndex := normalizeParentRehabPageIndex(query.PageIndex)
	pageSize := normalizeParentRehabPageSize(query.PageSize)
	items, total, err := svc.listParentRehabRecordItemsForTarget(ctx, target, pageIndex, pageSize)
	if err != nil {
		return model.ParentRehabRecordSummaryVO{}, err
	}

	return model.ParentRehabRecordSummaryVO{
		Students:  students,
		Items:     items,
		PageIndex: pageIndex,
		PageSize:  pageSize,
		Total:     total,
		HasMore:   pageIndex*pageSize < total,
	}, nil
}

func (svc *Service) GetParentRehabRecordDetailByPhone(ctx context.Context, phone string, query model.ParentRehabRecordDetailQueryDTO) (model.ParentRehabRecordDetailVO, error) {
	if svc == nil || svc.repo == nil {
		return model.ParentRehabRecordDetailVO{}, errors.New("家长端康复记录详情服务未初始化")
	}

	phone = normalizeParentPhone(phone)
	if phone == "" {
		return model.ParentRehabRecordDetailVO{}, errors.New("手机号不能为空")
	}
	recordID := strings.TrimSpace(query.StudentTeachingRecordID)
	if recordID == "" {
		return model.ParentRehabRecordDetailVO{}, errors.New("缺少康复记录标识")
	}

	rows, err := svc.repo.ListParentStudentCandidatesByPhone(ctx, phone)
	if err != nil {
		return model.ParentRehabRecordDetailVO{}, err
	}
	displayProfiles, err := svc.resolveParentStudentDisplayProfiles(ctx, rows)
	if err != nil {
		return model.ParentRehabRecordDetailVO{}, err
	}

	targets := buildParentScheduleTargets(rows, displayProfiles)
	if len(targets) == 0 {
		return model.ParentRehabRecordDetailVO{}, errors.New("暂未绑定学员")
	}

	studentID := strings.TrimSpace(query.StudentID)
	if studentID != "" {
		target, ok := resolveParentRehabTarget(targets, studentID)
		if !ok {
			return model.ParentRehabRecordDetailVO{}, errors.New("未找到对应学员")
		}
		detail, found, err := svc.getParentRehabRecordDetailForTarget(ctx, target, recordID)
		if err != nil {
			return model.ParentRehabRecordDetailVO{}, err
		}
		if !found {
			return model.ParentRehabRecordDetailVO{}, errors.New("未找到对应康复记录")
		}
		return detail, nil
	}

	for _, target := range targets {
		detail, found, err := svc.getParentRehabRecordDetailForTarget(ctx, target, recordID)
		if err != nil {
			return model.ParentRehabRecordDetailVO{}, err
		}
		if found {
			return detail, nil
		}
	}

	return model.ParentRehabRecordDetailVO{}, errors.New("未找到对应康复记录")
}

func (svc *Service) SaveParentRehabFeedbackByPhone(ctx context.Context, phone string, dto model.ParentRehabFeedbackSaveDTO) (bool, error) {
	if svc == nil || svc.repo == nil {
		return false, errors.New("家长端康复记录反馈服务未初始化")
	}

	phone = normalizeParentPhone(phone)
	if phone == "" {
		return false, errors.New("手机号不能为空")
	}
	recordID := strings.TrimSpace(dto.StudentTeachingRecordID)
	if recordID == "" {
		return false, errors.New("缺少康复记录标识")
	}

	rows, err := svc.repo.ListParentStudentCandidatesByPhone(ctx, phone)
	if err != nil {
		return false, err
	}
	displayProfiles, err := svc.resolveParentStudentDisplayProfiles(ctx, rows)
	if err != nil {
		return false, err
	}

	targets := buildParentScheduleTargets(rows, displayProfiles)
	if len(targets) == 0 {
		return false, errors.New("暂未绑定学员")
	}

	studentID := strings.TrimSpace(dto.StudentID)
	if studentID != "" {
		target, ok := resolveParentRehabTarget(targets, studentID)
		if !ok {
			return false, errors.New("未找到对应学员")
		}
		return svc.saveParentRehabFeedbackForTarget(ctx, target, dto)
	}

	for _, target := range targets {
		saved, err := svc.saveParentRehabFeedbackForTarget(ctx, target, dto)
		if err != nil {
			return false, err
		}
		if saved {
			return true, nil
		}
	}

	return false, errors.New("未找到对应康复记录")
}

func (svc *Service) listParentRehabRecordItemsForTarget(ctx context.Context, target parentScheduleTarget, pageIndex, pageSize int) ([]model.ParentRehabRecordVO, int, error) {
	sourceStudentIDs := []int64{target.StudentID}
	recordIDs, total, err := svc.repo.ListParentPublishedRehabRecordIDs(ctx, target.InstID, sourceStudentIDs, pageIndex, pageSize)
	if err != nil {
		return nil, 0, err
	}

	if total <= 0 && shouldFallbackToParentScheduleAliases(target) {
		aliasIDs, err := svc.resolveParentScheduleAliasStudentIDs(ctx, target)
		if err != nil {
			return nil, 0, err
		}
		if len(aliasIDs) > 0 {
			sourceStudentIDs = aliasIDs
			recordIDs, total, err = svc.repo.ListParentPublishedRehabRecordIDs(ctx, target.InstID, sourceStudentIDs, pageIndex, pageSize)
			if err != nil {
				return nil, 0, err
			}
		}
	}

	if total <= 0 || len(recordIDs) == 0 {
		return []model.ParentRehabRecordVO{}, total, nil
	}

	recordItems, err := svc.listParentClassRecordSourceItemsByRecordIDs(ctx, target.InstID, sourceStudentIDs, recordIDs)
	if err != nil {
		return nil, 0, err
	}
	if len(recordItems) == 0 {
		return []model.ParentRehabRecordVO{}, total, nil
	}

	sortParentStudentTeachingRecordItems(recordItems)
	items := make([]model.ParentRehabRecordVO, 0, len(recordItems))
	for _, item := range recordItems {
		detail, err := svc.repo.GetStudentRehabRecordDetail(ctx, target.InstID, model.StudentRehabRecordQueryDTO{
			StudentTeachingRecordID: strings.TrimSpace(item.StudentTeachingRecordID),
		})
		if err != nil {
			return nil, 0, err
		}
		if !isParentRehabRecorded(item, detail) {
			continue
		}
		items = append(items, buildParentRehabRecordVO(target, item, detail))
	}

	return items, total, nil
}

func (svc *Service) getParentRehabRecordDetailForTarget(ctx context.Context, target parentScheduleTarget, recordID string) (model.ParentRehabRecordDetailVO, bool, error) {
	sourceStudentIDs := []int64{target.StudentID}
	recordItems, err := svc.listParentClassRecordSourceItemsByRecordIDs(ctx, target.InstID, sourceStudentIDs, []string{recordID})
	if err != nil {
		return model.ParentRehabRecordDetailVO{}, false, err
	}

	if len(recordItems) == 0 && shouldFallbackToParentScheduleAliases(target) {
		aliasIDs, err := svc.resolveParentScheduleAliasStudentIDs(ctx, target)
		if err != nil {
			return model.ParentRehabRecordDetailVO{}, false, err
		}
		if len(aliasIDs) > 0 {
			sourceStudentIDs = aliasIDs
			recordItems, err = svc.listParentClassRecordSourceItemsByRecordIDs(ctx, target.InstID, sourceStudentIDs, []string{recordID})
			if err != nil {
				return model.ParentRehabRecordDetailVO{}, false, err
			}
		}
	}

	if len(recordItems) == 0 {
		return model.ParentRehabRecordDetailVO{}, false, nil
	}

	sortParentStudentTeachingRecordItems(recordItems)
	item := recordItems[0]
	detail, err := svc.repo.GetStudentRehabRecordDetail(ctx, target.InstID, model.StudentRehabRecordQueryDTO{
		StudentTeachingRecordID: strings.TrimSpace(item.StudentTeachingRecordID),
	})
	if err != nil {
		return model.ParentRehabRecordDetailVO{}, false, err
	}
	if !isParentRehabRecorded(item, detail) {
		return model.ParentRehabRecordDetailVO{}, false, nil
	}

	return buildParentRehabRecordDetailVO(target, item, detail), true, nil
}

func (svc *Service) saveParentRehabFeedbackForTarget(ctx context.Context, target parentScheduleTarget, dto model.ParentRehabFeedbackSaveDTO) (bool, error) {
	detail, found, err := svc.getParentRehabRecordDetailForTarget(ctx, target, strings.TrimSpace(dto.StudentTeachingRecordID))
	if err != nil {
		return false, err
	}
	if !found {
		return false, nil
	}
	if detail.Published == nil {
		return false, errors.New("当前康复记录暂不支持填写家长反馈")
	}
	return svc.repo.SaveParentRehabFeedback(ctx, target.InstID, dto)
}

func resolveParentRehabTarget(targets []parentScheduleTarget, studentID string) (parentScheduleTarget, bool) {
	studentID = strings.TrimSpace(studentID)
	if studentID == "" {
		return targets[0], true
	}
	for _, target := range targets {
		if strconv.FormatInt(target.StudentID, 10) == studentID {
			return target, true
		}
	}
	return parentScheduleTarget{}, false
}

func normalizeParentRehabPageSize(pageSize int) int {
	if pageSize <= 0 {
		return parentRehabRecordDefaultPageSize
	}
	if pageSize > parentRehabRecordMaxPageSize {
		return parentRehabRecordMaxPageSize
	}
	return pageSize
}

func normalizeParentRehabPageIndex(pageIndex int) int {
	if pageIndex <= 0 {
		return parentRehabRecordDefaultPageIndex
	}
	return pageIndex
}

func buildParentRehabRecordVO(target parentScheduleTarget, item model.StudentTeachingRecordItem, detail model.StudentRehabRecordDetailResult) model.ParentRehabRecordVO {
	recordDate := parentClassRecordDate(item.StartTime)
	startTime := parentClassRecordClock(item.StartTime)
	endTime := parentClassRecordClock(item.EndTime)
	remark := parentRehabRemark(item)
	summaryText, trainingTarget, performance, suggestion := buildParentRehabContentSummary(detail, remark)
	updatedTime, updatedStaffName := buildParentRehabUpdatedMeta(item, detail)

	return model.ParentRehabRecordVO{
		ID:                      firstNonEmptyString(strings.TrimSpace(item.StudentTeachingRecordID), strings.TrimSpace(item.TeachingRecordID)),
		StudentTeachingRecordID: strings.TrimSpace(item.StudentTeachingRecordID),
		TeachingRecordID:        strings.TrimSpace(item.TeachingRecordID),
		InstID:                  target.InstID,
		CampusID:                target.CampusID,
		CampusName:              target.CampusName,
		StudentID:               strconv.FormatInt(target.StudentID, 10),
		StudentName:             target.StudentName,
		StudentAvatarURL:        target.AvatarURL,
		Date:                    firstNonEmptyString(recordDate, "-"),
		StartTime:               firstNonEmptyString(startTime, "-"),
		EndTime:                 firstNonEmptyString(endTime, "-"),
		LessonTime:              buildParentRehabLessonTime(recordDate, startTime, endTime),
		ClassName:               parentClassRecordClassName(item),
		CourseName:              parentClassRecordCourseName(item),
		TeacherName:             defaultParentScheduleText(item.TeacherName),
		Classroom:               defaultParentScheduleText(item.ClassRoomName),
		StatusText:              "已记录",
		SummaryText:             firstNonEmptyString(summaryText, "已记录康复内容"),
		TrainingTarget:          trainingTarget,
		Performance:             performance,
		Suggestion:              suggestion,
		Remark:                  firstNonEmptyString(remark, "-"),
		UpdatedTime:             firstNonEmptyString(updatedTime, "-"),
		UpdatedStaffName:        firstNonEmptyString(updatedStaffName, "-"),
		HasPublished:            detail.HasPublished && detail.Published != nil,
	}
}

func buildParentRehabRecordDetailVO(target parentScheduleTarget, item model.StudentTeachingRecordItem, detail model.StudentRehabRecordDetailResult) model.ParentRehabRecordDetailVO {
	record := buildParentRehabRecordVO(target, item, detail)
	return model.ParentRehabRecordDetailVO{
		Student: model.ParentBoundStudentVO{
			ID:                strconv.FormatInt(target.StudentID, 10),
			InstID:            target.InstID,
			CampusID:          target.CampusID,
			CampusName:        target.CampusName,
			Name:              target.StudentName,
			AvatarURL:         target.AvatarURL,
			StudentStatus:     target.DisplayStatus,
			StudentStatusText: parentStudentStatusText(target.DisplayStatus),
		},
		Record:            record,
		Published:         detail.Published,
		PreviousPublished: detail.PreviousPublished,
	}
}

func isParentRehabRecorded(item model.StudentTeachingRecordItem, detail model.StudentRehabRecordDetailResult) bool {
	return detail.HasPublished && detail.Published != nil
}

func parentRehabRemark(item model.StudentTeachingRecordItem) string {
	remark := strings.TrimSpace(item.ExternalRemark)
	if remark != "" {
		return remark
	}
	return strings.TrimSpace(item.Remark)
}

func buildParentRehabUpdatedMeta(item model.StudentTeachingRecordItem, detail model.StudentRehabRecordDetailResult) (string, string) {
	if detail.HasPublished && detail.Published != nil {
		return strings.TrimSpace(detail.Published.UpdatedTime), strings.TrimSpace(detail.Published.UpdatedStaffName)
	}
	updatedTime := strings.TrimSpace(item.UpdatedTime)
	if updatedTime == "" {
		updatedTime = strings.TrimSpace(item.RecordTime)
	}
	return updatedTime, strings.TrimSpace(item.UpdatedStaffName)
}

func buildParentRehabContentSummary(detail model.StudentRehabRecordDetailResult, fallbackRemark string) (string, string, string, string) {
	if detail.HasPublished && detail.Published != nil {
		content := detail.Published.Content
		trainingTarget := strings.TrimSpace(content.TrainingTarget)
		performance := strings.TrimSpace(content.Performance)
		suggestion := strings.TrimSpace(content.Suggestion)
		candidates := []string{
			trainingTarget,
			performance,
			suggestion,
		}
		for _, item := range content.TrainingItems {
			candidates = append(candidates, strings.TrimSpace(item.Title), strings.TrimSpace(item.Content))
		}
		for _, item := range candidates {
			if item == "" {
				continue
			}
			runes := []rune(item)
			if len(runes) > 120 {
				return string(runes[:120]), trainingTarget, performance, suggestion
			}
			return item, trainingTarget, performance, suggestion
		}
	}
	fallbackRemark = strings.TrimSpace(fallbackRemark)
	if fallbackRemark == "" {
		return "", "", "", ""
	}
	runes := []rune(fallbackRemark)
	if len(runes) > 120 {
		return string(runes[:120]), "", "", ""
	}
	return fallbackRemark, "", "", ""
}

func buildParentRehabLessonTime(dateText, startTime, endTime string) string {
	dateText = strings.TrimSpace(dateText)
	startTime = strings.TrimSpace(startTime)
	endTime = strings.TrimSpace(endTime)
	switch {
	case dateText != "" && startTime != "" && endTime != "":
		return dateText + " " + startTime + "~" + endTime
	case dateText != "" && startTime != "":
		return dateText + " " + startTime
	case dateText != "":
		return dateText
	case startTime != "" && endTime != "":
		return startTime + "~" + endTime
	default:
		return firstNonEmptyString(startTime, "-")
	}
}

func sortParentRehabRecordVOs(items []model.ParentRehabRecordVO) {
	sort.SliceStable(items, func(i, j int) bool {
		leftTime := strings.TrimSpace(items[i].UpdatedTime)
		rightTime := strings.TrimSpace(items[j].UpdatedTime)
		if leftTime != rightTime {
			return leftTime > rightTime
		}
		leftDate := strings.TrimSpace(items[i].LessonTime)
		rightDate := strings.TrimSpace(items[j].LessonTime)
		if leftDate != rightDate {
			return leftDate > rightDate
		}
		return strings.TrimSpace(items[i].ID) > strings.TrimSpace(items[j].ID)
	})
}
