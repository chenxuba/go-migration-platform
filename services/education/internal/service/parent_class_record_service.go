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
	parentClassRecordDefaultPageIndex = 1
	parentClassRecordDefaultPageSize  = 50
	parentClassRecordMaxPageSize      = 100
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
	displayProfiles, err := svc.resolveParentStudentDisplayProfiles(ctx, rows)
	if err != nil {
		return model.ParentClassRecordSummaryVO{}, err
	}

	targets := buildParentScheduleTargets(rows, displayProfiles)
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

	pageIndex := normalizeParentClassRecordPageIndex(query.PageIndex)
	pageSize := normalizeParentClassRecordPageSize(query.PageSize)
	items, total, err := svc.listParentClassRecordItemsForTarget(ctx, target, pageIndex, pageSize)
	if err != nil {
		return model.ParentClassRecordSummaryVO{}, err
	}

	return model.ParentClassRecordSummaryVO{
		Students:  students,
		Items:     items,
		PageIndex: pageIndex,
		PageSize:  pageSize,
		Total:     total,
		HasMore:   pageIndex*pageSize < total,
	}, nil
}

func (svc *Service) GetParentClassRecordDetailByPhone(ctx context.Context, phone string, query model.ParentClassRecordDetailQueryDTO) (model.ParentClassRecordDetailVO, error) {
	if svc == nil || svc.repo == nil {
		return model.ParentClassRecordDetailVO{}, errors.New("家长端上课记录详情服务未初始化")
	}

	phone = normalizeParentPhone(phone)
	if phone == "" {
		return model.ParentClassRecordDetailVO{}, errors.New("手机号不能为空")
	}
	recordID := strings.TrimSpace(query.StudentTeachingRecordID)
	if recordID == "" {
		return model.ParentClassRecordDetailVO{}, errors.New("缺少上课记录标识")
	}

	rows, err := svc.repo.ListParentStudentCandidatesByPhone(ctx, phone)
	if err != nil {
		return model.ParentClassRecordDetailVO{}, err
	}
	displayProfiles, err := svc.resolveParentStudentDisplayProfiles(ctx, rows)
	if err != nil {
		return model.ParentClassRecordDetailVO{}, err
	}

	targets := buildParentScheduleTargets(rows, displayProfiles)
	if len(targets) == 0 {
		return model.ParentClassRecordDetailVO{}, errors.New("暂未绑定学员")
	}

	studentID := strings.TrimSpace(query.StudentID)
	if studentID != "" {
		target, ok := resolveParentClassRecordTarget(targets, studentID)
		if !ok {
			return model.ParentClassRecordDetailVO{}, errors.New("未找到对应学员")
		}
		detail, found, err := svc.getParentClassRecordDetailForTarget(ctx, target, recordID)
		if err != nil {
			return model.ParentClassRecordDetailVO{}, err
		}
		if !found {
			return model.ParentClassRecordDetailVO{}, errors.New("未找到对应上课记录")
		}
		return detail, nil
	}

	for _, target := range targets {
		detail, found, err := svc.getParentClassRecordDetailForTarget(ctx, target, recordID)
		if err != nil {
			return model.ParentClassRecordDetailVO{}, err
		}
		if found {
			return detail, nil
		}
	}

	return model.ParentClassRecordDetailVO{}, errors.New("未找到对应上课记录")
}

func (svc *Service) listParentClassRecordItemsForTarget(ctx context.Context, target parentScheduleTarget, pageIndex, pageSize int) ([]model.ParentClassRecordVO, int, error) {
	fetchSize := pageIndex * pageSize
	sourceStudentIDs := []int64{target.StudentID}
	recordItems, total, err := svc.listParentClassRecordSourceItems(ctx, target.InstID, sourceStudentIDs, fetchSize)
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
			recordItems, total, err = svc.listParentClassRecordSourceItems(ctx, target.InstID, sourceStudentIDs, fetchSize)
			if err != nil {
				return nil, 0, err
			}
		}
	}

	if total <= 0 || len(recordItems) == 0 {
		return []model.ParentClassRecordVO{}, total, nil
	}

	sortParentStudentTeachingRecordItems(recordItems)
	recordDeductDays, err := svc.buildParentClassRecordDeductDays(ctx, target.InstID, sourceStudentIDs, recordItems)
	if err != nil {
		return nil, 0, err
	}

	offset := (pageIndex - 1) * pageSize
	if offset >= len(recordItems) {
		return []model.ParentClassRecordVO{}, total, nil
	}
	end := offset + pageSize
	if end > len(recordItems) {
		end = len(recordItems)
	}

	pageItems := recordItems[offset:end]
	items := make([]model.ParentClassRecordVO, 0, len(pageItems))
	for _, item := range pageItems {
		items = append(items, buildParentClassRecordVO(target, item, recordDeductDays[parentClassRecordItemKey(item)]))
	}
	sortParentClassRecordVOs(items)
	return items, total, nil
}

func (svc *Service) getParentClassRecordDetailForTarget(ctx context.Context, target parentScheduleTarget, recordID string) (model.ParentClassRecordDetailVO, bool, error) {
	sourceStudentIDs := []int64{target.StudentID}
	recordItems, err := svc.listParentClassRecordSourceItemsByRecordIDs(ctx, target.InstID, sourceStudentIDs, []string{recordID})
	if err != nil {
		return model.ParentClassRecordDetailVO{}, false, err
	}

	if len(recordItems) == 0 && shouldFallbackToParentScheduleAliases(target) {
		aliasIDs, err := svc.resolveParentScheduleAliasStudentIDs(ctx, target)
		if err != nil {
			return model.ParentClassRecordDetailVO{}, false, err
		}
		if len(aliasIDs) > 0 {
			sourceStudentIDs = aliasIDs
			recordItems, err = svc.listParentClassRecordSourceItemsByRecordIDs(ctx, target.InstID, sourceStudentIDs, []string{recordID})
			if err != nil {
				return model.ParentClassRecordDetailVO{}, false, err
			}
		}
	}

	if len(recordItems) == 0 {
		return model.ParentClassRecordDetailVO{}, false, nil
	}

	sortParentStudentTeachingRecordItems(recordItems)
	recordDeductDays, err := svc.buildParentClassRecordDeductDays(ctx, target.InstID, sourceStudentIDs, recordItems)
	if err != nil {
		return model.ParentClassRecordDetailVO{}, false, err
	}

	item := recordItems[0]
	return buildParentClassRecordDetailVO(target, item, recordDeductDays[parentClassRecordItemKey(item)]), true, nil
}

func (svc *Service) listParentClassRecordSourceItems(ctx context.Context, instID int64, studentIDs []int64, pageSize int) ([]model.StudentTeachingRecordItem, int, error) {
	items := make([]model.StudentTeachingRecordItem, 0, len(studentIDs)*2)
	seen := make(map[string]struct{}, len(studentIDs)*2)
	total := 0

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
			return nil, 0, err
		}
		total += result.Total
		for _, item := range result.List {
			recordID := parentClassRecordItemKey(item)
			if _, exists := seen[recordID]; exists {
				continue
			}
			seen[recordID] = struct{}{}
			items = append(items, item)
		}
	}

	return items, total, nil
}

func (svc *Service) listParentClassRecordSourceItemsByRecordIDs(ctx context.Context, instID int64, studentIDs []int64, recordIDs []string) ([]model.StudentTeachingRecordItem, error) {
	recordIDs = normalizeParentClassRecordIDs(recordIDs)
	if len(recordIDs) == 0 {
		return []model.StudentTeachingRecordItem{}, nil
	}

	items := make([]model.StudentTeachingRecordItem, 0, len(recordIDs))
	seen := make(map[string]struct{}, len(recordIDs))
	pageSize := len(recordIDs)
	if pageSize < 1 {
		pageSize = 1
	}

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
				StudentID:                strconv.FormatInt(studentID, 10),
				StudentTeachingRecordIDs: recordIDs,
			},
		})
		if err != nil {
			return nil, err
		}
		for _, item := range result.List {
			key := parentClassRecordItemKey(item)
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			items = append(items, item)
		}
		if len(items) >= len(recordIDs) {
			break
		}
	}

	return items, nil
}

func (svc *Service) buildParentClassRecordDeductDays(ctx context.Context, instID int64, studentIDs []int64, items []model.StudentTeachingRecordItem) (map[string]float64, error) {
	if len(items) == 0 {
		return map[string]float64{}, nil
	}

	timeSlotConsumeMap, err := svc.buildParentTimeSlotConsumeMap(ctx, instID, studentIDs, items)
	if err != nil {
		return nil, err
	}

	result := make(map[string]float64, len(items))
	assignedTimeSlotDays := make(map[string]struct{}, len(timeSlotConsumeMap))
	for _, item := range items {
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
		result[parentClassRecordItemKey(item)] = deductDays
	}
	return result, nil
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

func normalizeParentClassRecordPageIndex(pageIndex int) int {
	if pageIndex <= 0 {
		return parentClassRecordDefaultPageIndex
	}
	return pageIndex
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
			StudentStatus:     target.DisplayStatus,
			StudentStatusText: parentStudentStatusText(target.DisplayStatus),
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
		ArrearQuantity:          roundParentCourseMetric(item.ArrearQuantity),
		ShowDeductQuantity:      showDeductQuantity,
		ShowDeductDays:          showDeductDays,
	}
}

func buildParentClassRecordDetailVO(target parentScheduleTarget, item model.StudentTeachingRecordItem, deductDays float64) model.ParentClassRecordDetailVO {
	base := buildParentClassRecordVO(target, item, deductDays)
	summaryText, summaryLabel := buildParentClassRecordSummary(item, base.DeductQuantity, base.DeductDays, base.ArrearQuantity)
	deductLabel := parentClassRecordDeductLabel(item, base.ShowDeductDays)
	deductText := buildParentClassRecordDeductText(item, base.DeductQuantity, base.DeductDays, base.ShowDeductDays)

	return model.ParentClassRecordDetailVO{
		ID:                      base.ID,
		StudentTeachingRecordID: base.StudentTeachingRecordID,
		TeachingRecordID:        base.TeachingRecordID,
		InstID:                  base.InstID,
		CampusID:                base.CampusID,
		CampusName:              base.CampusName,
		StudentID:               base.StudentID,
		StudentName:             base.StudentName,
		StudentAvatarURL:        base.StudentAvatarURL,
		Date:                    base.Date,
		StartTime:               base.StartTime,
		EndTime:                 base.EndTime,
		LessonTime:              buildParentClassRecordLessonTime(base.Date, base.StartTime, base.EndTime),
		ClassName:               base.ClassName,
		CourseName:              base.CourseName,
		TeacherName:             base.TeacherName,
		Classroom:               base.Classroom,
		Remark:                  base.Remark,
		Status:                  base.Status,
		StatusText:              base.StatusText,
		ChargingMode:            base.ChargingMode,
		ChargingModeText:        base.ChargingModeText,
		DeductQuantity:          base.DeductQuantity,
		DeductDays:              base.DeductDays,
		ArrearQuantity:          base.ArrearQuantity,
		ShowDeductQuantity:      base.ShowDeductQuantity,
		ShowDeductDays:          base.ShowDeductDays,
		SummaryText:             summaryText,
		SummaryLabel:            summaryLabel,
		DeductLabel:             deductLabel,
		DeductText:              deductText,
		ArrearLabel:             "拖欠数量",
		ArrearText:              buildParentCourseArrearDisplayText(item.SkuMode, base.ArrearQuantity),
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

func parentClassRecordItemKey(item model.StudentTeachingRecordItem) string {
	recordID := strings.TrimSpace(item.StudentTeachingRecordID)
	if recordID != "" {
		return recordID
	}
	return strings.TrimSpace(item.TeachingRecordID) + "|" + strings.TrimSpace(item.StudentID) + "|" + strings.TrimSpace(item.StartTime)
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

func normalizeParentClassRecordIDs(ids []string) []string {
	result := make([]string, 0, len(ids))
	seen := make(map[string]struct{}, len(ids))
	for _, item := range ids {
		id := strings.TrimSpace(item)
		if id == "" {
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

func buildParentClassRecordLessonTime(date, startTime, endTime string) string {
	date = strings.TrimSpace(date)
	startTime = strings.TrimSpace(startTime)
	endTime = strings.TrimSpace(endTime)
	switch {
	case date != "" && startTime != "" && endTime != "":
		return date + " " + startTime + "~" + endTime
	case date != "" && startTime != "":
		return date + " " + startTime
	case date != "":
		return date
	case startTime != "" && endTime != "":
		return startTime + "~" + endTime
	default:
		return firstNonEmptyString(startTime, "-")
	}
}

func buildParentClassRecordSummary(item model.StudentTeachingRecordItem, deductQuantity, deductDays, arrearQuantity float64) (string, string) {
	if isParentNoCountClassRecord(item) {
		return "-", "本次不计课时"
	}

	if roundParentCourseMetric(deductDays) > 0 {
		return "-" + formatParentCourseMetric(deductDays), "扣除天数"
	}

	if roundParentCourseMetric(deductQuantity) > 0 {
		return "-" + formatParentCourseMetric(deductQuantity), parentClassRecordDeductLabel(item, false)
	}

	if roundParentCourseMetric(arrearQuantity) > 0 {
		return "-" + formatParentCourseMetric(arrearQuantity), parentClassRecordArrearSummaryLabel(item.SkuMode)
	}

	return "-", parentClassRecordDefaultSummaryLabel(item.SkuMode)
}

func buildParentClassRecordDeductText(item model.StudentTeachingRecordItem, deductQuantity, deductDays float64, showDeductDays bool) string {
	if showDeductDays && roundParentCourseMetric(deductDays) > 0 {
		return formatParentCourseMetric(deductDays) + "天"
	}
	if roundParentCourseMetric(deductQuantity) <= 0 {
		return ""
	}
	switch normalizeParentCourseChargingMode(item.SkuMode) {
	case 3:
		return formatParentCourseMetric(deductQuantity) + "元"
	default:
		return formatParentCourseMetric(deductQuantity) + "课时"
	}
}

func parentClassRecordDeductLabel(item model.StudentTeachingRecordItem, showDeductDays bool) string {
	if showDeductDays {
		return "扣除天数"
	}
	switch normalizeParentCourseChargingMode(item.SkuMode) {
	case 3:
		return "扣除金额"
	default:
		return "扣除课时"
	}
}

func parentClassRecordArrearSummaryLabel(mode int) string {
	switch normalizeParentCourseChargingMode(mode) {
	case 2:
		return "拖欠天数"
	case 3:
		return "拖欠金额"
	default:
		return "拖欠课时"
	}
}

func parentClassRecordDefaultSummaryLabel(mode int) string {
	switch normalizeParentCourseChargingMode(mode) {
	case 2:
		return "扣除天数"
	case 3:
		return "扣除金额"
	default:
		return "扣除课时"
	}
}
