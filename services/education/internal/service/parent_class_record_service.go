package service

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

const (
	parentClassRecordDefaultPageSize = 50
	parentClassRecordMaxPageSize     = 100
)

func (svc *Service) ListParentClassRecordsByPhone(ctx context.Context, phone string, query model.ParentClassRecordQueryDTO) (model.ParentClassRecordSummaryVO, error) {
	if svc == nil || svc.repo == nil {
		return model.ParentClassRecordSummaryVO{}, errors.New("家长端上课记录服务未初始化")
	}

	phone = normalizeParentPhone(phone)
	if phone == "" {
		return model.ParentClassRecordSummaryVO{}, errors.New("手机号不能为空")
	}

	rows, err := svc.repo.ListParentStudentCandidatesByPhone(ctx, phone)
	if err != nil {
		return model.ParentClassRecordSummaryVO{}, err
	}

	targets := buildParentScheduleTargets(rows)
	students := buildParentBoundStudents(targets)
	if len(targets) == 0 {
		return model.ParentClassRecordSummaryVO{
			Students: students,
			Items:    []model.ParentClassRecordVO{},
		}, nil
	}

	target, ok := resolveParentClassRecordTarget(targets, query.StudentID)
	if !ok {
		return model.ParentClassRecordSummaryVO{}, errors.New("未找到对应学员")
	}

	items, err := svc.listParentClassRecordItemsForTarget(ctx, target, normalizeParentClassRecordPageSize(query.PageSize))
	if err != nil {
		return model.ParentClassRecordSummaryVO{}, err
	}

	return model.ParentClassRecordSummaryVO{
		Students: students,
		Items:    items,
	}, nil
}

func (svc *Service) listParentClassRecordItemsForTarget(ctx context.Context, target parentScheduleTarget, pageSize int) ([]model.ParentClassRecordVO, error) {
	sourceStudentIDs := []int64{target.StudentID}
	recordItems, err := svc.listParentClassRecordSourceItems(ctx, target.InstID, sourceStudentIDs, pageSize)
	if err != nil {
		return nil, err
	}

	if len(recordItems) == 0 && shouldFallbackToParentScheduleAliases(target) {
		aliasIDs, err := svc.resolveParentScheduleAliasStudentIDs(ctx, target)
		if err != nil {
			return nil, err
		}
		if len(aliasIDs) > 0 {
			sourceStudentIDs = aliasIDs
			recordItems, err = svc.listParentClassRecordSourceItems(ctx, target.InstID, sourceStudentIDs, pageSize)
			if err != nil {
				return nil, err
			}
		}
	}

	if len(recordItems) == 0 {
		return []model.ParentClassRecordVO{}, nil
	}

	sortParentStudentTeachingRecordItems(recordItems)
	if len(recordItems) > pageSize {
		recordItems = recordItems[:pageSize]
	}

	timeSlotConsumeMap, err := svc.buildParentTimeSlotConsumeMap(ctx, target.InstID, sourceStudentIDs, recordItems)
	if err != nil {
		return nil, err
	}

	items := make([]model.ParentClassRecordVO, 0, len(recordItems))
	assignedTimeSlotDays := make(map[string]struct{}, len(timeSlotConsumeMap))
	for _, item := range recordItems {
		deductDays := 0.0
		if isParentTimeSlotClassRecord(item) {
			key := buildParentTimeSlotConsumeKey(item.StudentID, item.LessonID, parentClassRecordDate(item.StartTime))
			if _, exists := assignedTimeSlotDays[key]; !exists {
				deductDays = timeSlotConsumeMap[key]
				if deductDays > 0 {
					assignedTimeSlotDays[key] = struct{}{}
				}
			}
		}
		items = append(items, buildParentClassRecordVO(target, item, deductDays))
	}

	sortParentClassRecordVOs(items)
	return items, nil
}

func (svc *Service) listParentClassRecordSourceItems(ctx context.Context, instID int64, studentIDs []int64, pageSize int) ([]model.StudentTeachingRecordItem, error) {
	items := make([]model.StudentTeachingRecordItem, 0, len(studentIDs)*2)
	seen := make(map[string]struct{}, len(studentIDs)*2)

	for _, studentID := range studentIDs {
		if studentID <= 0 {
			continue
		}
		result, err := svc.repo.GetStudentTeachingRecordPagedList(ctx, instID, model.StudentTeachingRecordPagedQueryDTO{
			PageRequestModel: model.RollCallPageRequestModel{
				PageIndex: 1,
				PageSize:  pageSize,
			},
			QueryModel: model.StudentTeachingRecordQueryModel{
				StudentID: strconv.FormatInt(studentID, 10),
			},
		})
		if err != nil {
			return nil, err
		}
		for _, item := range result.List {
			recordID := strings.TrimSpace(item.StudentTeachingRecordID)
			if recordID == "" {
				recordID = strings.TrimSpace(item.TeachingRecordID) + "|" + strings.TrimSpace(item.StudentID) + "|" + strings.TrimSpace(item.StartTime)
			}
			if _, exists := seen[recordID]; exists {
				continue
			}
			seen[recordID] = struct{}{}
			items = append(items, item)
		}
	}

	return items, nil
}

func (svc *Service) buildParentTimeSlotConsumeMap(ctx context.Context, instID int64, studentIDs []int64, items []model.StudentTeachingRecordItem) (map[string]float64, error) {
	timeSlotItems := make([]model.StudentTeachingRecordItem, 0, len(items))
	for _, item := range items {
		if isParentTimeSlotClassRecord(item) {
			timeSlotItems = append(timeSlotItems, item)
		}
	}
	if len(timeSlotItems) == 0 {
		return map[string]float64{}, nil
	}

	startDate := ""
	endDate := ""
	for _, item := range timeSlotItems {
		dateText := parentClassRecordDate(item.StartTime)
		if dateText == "" {
			continue
		}
		if startDate == "" || dateText < startDate {
			startDate = dateText
		}
		if endDate == "" || dateText > endDate {
			endDate = dateText
		}
	}
	if startDate == "" || endDate == "" {
		return map[string]float64{}, nil
	}

	rows, err := svc.repo.ListParentTimeSlotConsumeRecords(ctx, instID, studentIDs, startDate, endDate)
	if err != nil {
		return nil, err
	}

	result := make(map[string]float64, len(rows))
	for _, row := range rows {
		key := buildParentTimeSlotConsumeKey(strconv.FormatInt(row.StudentID, 10), strconv.FormatInt(row.LessonID, 10), row.Date)
		if key == "" || row.Quantity <= 0 {
			continue
		}
		result[key] += row.Quantity
	}
	return result, nil
}

func resolveParentClassRecordTarget(targets []parentScheduleTarget, studentID string) (parentScheduleTarget, bool) {
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

func normalizeParentClassRecordPageSize(pageSize int) int {
	if pageSize <= 0 {
		return parentClassRecordDefaultPageSize
	}
	if pageSize > parentClassRecordMaxPageSize {
		return parentClassRecordMaxPageSize
	}
	return pageSize
}

func buildParentBoundStudents(targets []parentScheduleTarget) []model.ParentBoundStudentVO {
	items := make([]model.ParentBoundStudentVO, 0, len(targets))
	for _, target := range targets {
		items = append(items, model.ParentBoundStudentVO{
			ID:                strconv.FormatInt(target.StudentID, 10),
			InstID:            target.InstID,
			CampusID:          target.CampusID,
			CampusName:        target.CampusName,
			Name:              target.StudentName,
			AvatarURL:         target.AvatarURL,
			StudentStatus:     target.StudentStatus,
			StudentStatusText: parentStudentStatusText(target.StudentStatus),
		})
	}
	return items
}

func buildParentClassRecordVO(target parentScheduleTarget, item model.StudentTeachingRecordItem, deductDays float64) model.ParentClassRecordVO {
	recordDate := parentClassRecordDate(item.StartTime)
	startTime := parentClassRecordClock(item.StartTime)
	endTime := parentClassRecordClock(item.EndTime)
	remark := strings.TrimSpace(item.ExternalRemark)
	if remark == "" {
		remark = strings.TrimSpace(item.Remark)
	}

	deductQuantity := item.ActualDeduct
	showDeductQuantity := !isParentTimeSlotClassRecord(item) && !isParentNoCountClassRecord(item) && deductQuantity > 0
	showDeductDays := deductDays > 0

	return model.ParentClassRecordVO{
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
		ClassName:               parentClassRecordClassName(item),
		CourseName:              parentClassRecordCourseName(item),
		TeacherName:             defaultParentScheduleText(item.TeacherName),
		Classroom:               defaultParentScheduleText(item.ClassRoomName),
		Remark:                  firstNonEmptyString(remark, "-"),
		Status:                  item.Status,
		StatusText:              parentClassRecordStatusText(item.Status),
		ChargingMode:            item.SkuMode,
		ChargingModeText:        parentClassRecordChargingModeText(item),
		DeductQuantity:          deductQuantity,
		DeductDays:              deductDays,
		ShowDeductQuantity:      showDeductQuantity,
		ShowDeductDays:          showDeductDays,
	}
}

func sortParentStudentTeachingRecordItems(items []model.StudentTeachingRecordItem) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].StartTime != items[j].StartTime {
			return items[i].StartTime > items[j].StartTime
		}
		if items[i].EndTime != items[j].EndTime {
			return items[i].EndTime > items[j].EndTime
		}
		return items[i].StudentTeachingRecordID > items[j].StudentTeachingRecordID
	})
}

func sortParentClassRecordVOs(items []model.ParentClassRecordVO) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Date != items[j].Date {
			return items[i].Date > items[j].Date
		}
		if items[i].StartTime != items[j].StartTime {
			return items[i].StartTime > items[j].StartTime
		}
		return items[i].StudentTeachingRecordID > items[j].StudentTeachingRecordID
	})
}

func buildParentTimeSlotConsumeKey(studentID, lessonID, date string) string {
	studentID = strings.TrimSpace(studentID)
	date = strings.TrimSpace(date)
	if studentID == "" || date == "" {
		return ""
	}
	lessonID = strings.TrimSpace(lessonID)
	return studentID + "|" + lessonID + "|" + date
}

func isParentTimeSlotClassRecord(item model.StudentTeachingRecordItem) bool {
	return item.SkuMode == 2
}

func isParentNoCountClassRecord(item model.StudentTeachingRecordItem) bool {
	return item.SourceType == 4
}

func parentClassRecordStatusText(status int) string {
	switch status {
	case 2:
		return "旷课"
	case 3:
		return "请假"
	case 4:
		return "未记录"
	default:
		return "到课"
	}
}

func parentClassRecordChargingModeText(item model.StudentTeachingRecordItem) string {
	if isParentNoCountClassRecord(item) {
		return "不记课时"
	}
	switch item.SkuMode {
	case 2:
		return "按时段"
	case 3:
		return "按金额"
	default:
		return "按课时"
	}
}

func parentClassRecordClassName(item model.StudentTeachingRecordItem) string {
	for _, value := range []string{item.ClassName, item.One2OneName, item.LessonName} {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return "课程记录"
}

func parentClassRecordCourseName(item model.StudentTeachingRecordItem) string {
	for _, value := range []string{item.LessonName, item.ClassName, item.One2OneName} {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return "课程"
}

func parentClassRecordDate(raw string) string {
	if parsed, ok := parseParentClassRecordTime(raw); ok {
		return parsed.Format("2006-01-02")
	}
	raw = strings.TrimSpace(raw)
	if len(raw) >= 10 {
		return raw[:10]
	}
	return raw
}

func parentClassRecordClock(raw string) string {
	if parsed, ok := parseParentClassRecordTime(raw); ok {
		return parsed.Format("15:04")
	}
	raw = strings.TrimSpace(raw)
	if len(raw) >= 16 {
		return raw[11:16]
	}
	return raw
}

func parseParentClassRecordTime(raw string) (time.Time, bool) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04",
	} {
		if parsed, err := time.ParseInLocation(layout, text, time.Local); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}
