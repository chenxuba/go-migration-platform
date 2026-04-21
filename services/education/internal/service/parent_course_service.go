package service

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/repository"
)

const (
	parentCourseFlowDefaultPageIndex = 1
	parentCourseFlowDefaultPageSize  = 20
	parentCourseFlowMaxPageSize      = 100
	parentCourseLowBalanceThreshold  = 5
)

func (svc *Service) ListParentCourseEnrollmentsByPhone(ctx context.Context, phone string, query model.ParentCourseEnrollmentQueryDTO) (model.ParentCourseEnrollmentSummaryVO, error) {
	if svc == nil || svc.repo == nil {
		return model.ParentCourseEnrollmentSummaryVO{}, errors.New("家长端报读课程服务未初始化")
	}

	phone = normalizeParentPhone(phone)
	if phone == "" {
		return model.ParentCourseEnrollmentSummaryVO{}, errors.New("手机号不能为空")
	}

	rows, err := svc.repo.ListParentStudentCandidatesByPhone(ctx, phone)
	if err != nil {
		return model.ParentCourseEnrollmentSummaryVO{}, err
	}
	displayStatuses, err := svc.resolveParentStudentDisplayStatuses(ctx, rows)
	if err != nil {
		return model.ParentCourseEnrollmentSummaryVO{}, err
	}

	targets := buildParentScheduleTargets(rows, displayStatuses)
	students := buildParentBoundStudents(targets)
	if len(targets) == 0 {
		return model.ParentCourseEnrollmentSummaryVO{
			Students: students,
			Items:    []model.ParentCourseEnrollmentVO{},
		}, nil
	}

	target, ok := resolveParentCourseTarget(targets, query.StudentID)
	if !ok {
		return model.ParentCourseEnrollmentSummaryVO{}, errors.New("未找到对应学员")
	}

	items, _, err := svc.listParentCourseEnrollmentItemsForTarget(ctx, target)
	if err != nil {
		return model.ParentCourseEnrollmentSummaryVO{}, err
	}

	return model.ParentCourseEnrollmentSummaryVO{
		Students: students,
		Items:    items,
	}, nil
}

func (svc *Service) GetParentCourseEnrollmentDetailByPhone(ctx context.Context, phone string, query model.ParentCourseEnrollmentDetailQueryDTO) (model.ParentCourseEnrollmentDetailVO, error) {
	if svc == nil || svc.repo == nil {
		return model.ParentCourseEnrollmentDetailVO{}, errors.New("家长端报读课程详情服务未初始化")
	}

	phone = normalizeParentPhone(phone)
	if phone == "" {
		return model.ParentCourseEnrollmentDetailVO{}, errors.New("手机号不能为空")
	}
	lessonID := strings.TrimSpace(query.LessonID)
	if lessonID == "" {
		return model.ParentCourseEnrollmentDetailVO{}, errors.New("缺少课程标识")
	}
	if query.ChargingMode <= 0 {
		return model.ParentCourseEnrollmentDetailVO{}, errors.New("缺少计费模式")
	}

	rows, err := svc.repo.ListParentStudentCandidatesByPhone(ctx, phone)
	if err != nil {
		return model.ParentCourseEnrollmentDetailVO{}, err
	}
	displayStatuses, err := svc.resolveParentStudentDisplayStatuses(ctx, rows)
	if err != nil {
		return model.ParentCourseEnrollmentDetailVO{}, err
	}

	targets := buildParentScheduleTargets(rows, displayStatuses)
	if len(targets) == 0 {
		return model.ParentCourseEnrollmentDetailVO{}, errors.New("暂未绑定学员")
	}

	target, ok := resolveParentCourseTarget(targets, query.StudentID)
	if !ok {
		return model.ParentCourseEnrollmentDetailVO{}, errors.New("未找到对应学员")
	}

	items, sourceStudentIDs, err := svc.listParentCourseEnrollmentItemsForTarget(ctx, target)
	if err != nil {
		return model.ParentCourseEnrollmentDetailVO{}, err
	}

	var courseItem *model.ParentCourseEnrollmentVO
	for index := range items {
		if items[index].LessonID == lessonID && items[index].ChargingMode == query.ChargingMode {
			courseItem = &items[index]
			break
		}
	}
	if courseItem == nil {
		return model.ParentCourseEnrollmentDetailVO{}, errors.New("未找到对应报读课程")
	}

	accountIDs, err := svc.listParentCourseBucketAccountIDs(ctx, target.InstID, sourceStudentIDs, lessonID, query.ChargingMode)
	if err != nil {
		return model.ParentCourseEnrollmentDetailVO{}, err
	}

	pageIndex := normalizeParentCourseFlowPageIndex(query.PageIndex)
	pageSize := normalizeParentCourseFlowPageSize(query.PageSize)
	flowRows, total, err := svc.repo.ListParentCourseFlowRecordsByTuitionAccountIDs(ctx, target.InstID, accountIDs, pageIndex, pageSize)
	if err != nil {
		return model.ParentCourseEnrollmentDetailVO{}, err
	}

	flowItems := make([]model.ParentCourseEnrollmentFlowVO, 0, len(flowRows))
	for _, row := range flowRows {
		flowItems = append(flowItems, buildParentCourseEnrollmentFlowVO(row, courseItem.ChargingMode))
	}

	return model.ParentCourseEnrollmentDetailVO{
		Student: model.ParentBoundStudentVO{
			ID:                strconv.FormatInt(target.StudentID, 10),
			InstID:            target.InstID,
			CampusID:          target.CampusID,
			CampusName:        target.CampusName,
			Name:              target.StudentName,
			AvatarURL:         target.AvatarURL,
			StudentStatus:     target.StudentStatus,
			StudentStatusText: parentStudentStatusText(target.DisplayStatus),
		},
		Course:    *courseItem,
		Items:     flowItems,
		PageIndex: pageIndex,
		PageSize:  pageSize,
		Total:     total,
		HasMore:   pageIndex*pageSize < total,
	}, nil
}

func (svc *Service) listParentCourseEnrollmentItemsForTarget(ctx context.Context, target parentScheduleTarget) ([]model.ParentCourseEnrollmentVO, []int64, error) {
	sourceStudentIDs := []int64{target.StudentID}
	readingItems, err := svc.listParentCourseReadingItems(ctx, target.InstID, sourceStudentIDs)
	if err != nil {
		return nil, nil, err
	}

	if len(readingItems) == 0 && shouldFallbackToParentScheduleAliases(target) {
		aliasIDs, err := svc.resolveParentScheduleAliasStudentIDs(ctx, target)
		if err != nil {
			return nil, nil, err
		}
		if len(aliasIDs) > 0 {
			sourceStudentIDs = aliasIDs
			readingItems, err = svc.listParentCourseReadingItems(ctx, target.InstID, sourceStudentIDs)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	items := make([]model.ParentCourseEnrollmentVO, 0, len(readingItems))
	for _, item := range readingItems {
		if parentCourseIsClosed(item) {
			continue
		}
		items = append(items, buildParentCourseEnrollmentVO(target, item))
	}
	sortParentCourseEnrollmentVOs(items)
	return items, sourceStudentIDs, nil
}

func (svc *Service) listParentCourseReadingItems(ctx context.Context, instID int64, studentIDs []int64) ([]model.TuitionAccountReadingItem, error) {
	items := make([]model.TuitionAccountReadingItem, 0, len(studentIDs)*2)
	seen := make(map[string]struct{}, len(studentIDs)*2)
	for _, studentID := range studentIDs {
		if studentID <= 0 {
			continue
		}
		result, err := svc.repo.GetTuitionAccountReadingList(ctx, instID, strconv.FormatInt(studentID, 10))
		if err != nil {
			return nil, err
		}
		for _, item := range result.List {
			key := parentCourseBucketKey(item)
			if key == "" {
				continue
			}
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			items = append(items, item)
		}
	}
	return items, nil
}

func (svc *Service) listParentCourseBucketAccountIDs(ctx context.Context, instID int64, studentIDs []int64, lessonID string, chargingMode int) ([]int64, error) {
	courseID, err := strconv.ParseInt(strings.TrimSpace(lessonID), 10, 64)
	if err != nil || courseID <= 0 {
		return nil, errors.New("课程标识不合法")
	}

	chargingMode = normalizeParentCourseChargingMode(chargingMode)
	accountIDs := make([]int64, 0, len(studentIDs)*2)
	for _, studentID := range studentIDs {
		if studentID <= 0 {
			continue
		}
		readingList, lerr := svc.repo.GetTuitionAccountReadingList(ctx, instID, strconv.FormatInt(studentID, 10))
		if lerr != nil {
			return nil, lerr
		}
		teachMethod := 0
		for _, item := range readingList.List {
			if strings.TrimSpace(item.LessonID) != lessonID {
				continue
			}
			if normalizeParentCourseChargingMode(derefParentCourseInt(item.LessonChargingMode)) != normalizeParentCourseChargingMode(chargingMode) {
				continue
			}
			teachMethod = derefParentCourseInt(item.LessonType)
			break
		}
		if teachMethod <= 0 {
			continue
		}
		items, err := svc.repo.ListTuitionAccountIDsForStudentCourseBucketAllStatuses(ctx, nil, instID, studentID, courseID, teachMethod, chargingMode)
		if err != nil {
			return nil, err
		}
		accountIDs = append(accountIDs, items...)
	}

	return normalizeParentCourseAccountIDs(accountIDs), nil
}

func resolveParentCourseTarget(targets []parentScheduleTarget, studentID string) (parentScheduleTarget, bool) {
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

func buildParentCourseEnrollmentVO(target parentScheduleTarget, item model.TuitionAccountReadingItem) model.ParentCourseEnrollmentVO {
	chargingMode := normalizeParentCourseChargingMode(derefParentCourseInt(item.LessonChargingMode))
	remainingQuantity := item.RemainQuantity
	totalQuantity := item.TotalQuantity
	if chargingMode == 2 {
		remainingQuantity = roundParentCourseMetric(item.RemainQuantity)
		totalQuantity = roundParentCourseMetric(item.TotalQuantity)
	}
	status := derefParentCourseInt(item.Status)
	validDate := formatParentCourseDate(item.ValidDate)
	endDate := formatParentCourseDate(item.EndDate)
	showRemainingQuantity := chargingMode != 3
	lowBalanceText := buildParentCourseLowBalanceText(chargingMode, remainingQuantity, status)
	showValidRange := validDate != "" || endDate != ""

	return model.ParentCourseEnrollmentVO{
		ID:                     parentCourseBucketKey(item),
		InstID:                 target.InstID,
		CampusID:               target.CampusID,
		CampusName:             target.CampusName,
		StudentID:              strconv.FormatInt(target.StudentID, 10),
		StudentName:            target.StudentName,
		StudentAvatarURL:       target.AvatarURL,
		LessonID:               strings.TrimSpace(item.LessonID),
		LessonName:             firstNonEmptyString(strings.TrimSpace(item.LessonName), "报读课程"),
		ChargingMode:           chargingMode,
		ChargingModeText:       parentCourseChargingModeText(chargingMode),
		Status:                 status,
		StatusText:             parentCourseStatusText(status),
		TotalQuantity:          totalQuantity,
		RemainingQuantity:      remainingQuantity,
		TotalTuition:           item.TotalTuition,
		RemainingTuition:       item.Tuition,
		ShowRemainingQuantity:  showRemainingQuantity,
		RemainingQuantityLabel: parentCourseRemainingLabel(chargingMode),
		RemainingQuantityText:  buildParentCourseRemainingQuantityText(chargingMode, remainingQuantity),
		RemainingTuitionText:   buildParentCourseAmountText(item.Tuition),
		ValidDate:              validDate,
		EndDate:                endDate,
		ShowValidRange:         showValidRange,
		ValidRangeText:         buildParentCourseValidRangeText(validDate, endDate),
		LowBalance:             lowBalanceText != "",
		LowBalanceText:         lowBalanceText,
	}
}

func buildParentCourseEnrollmentFlowVO(item repository.ParentCourseFlowRecord, chargingMode int) model.ParentCourseEnrollmentFlowVO {
	return model.ParentCourseEnrollmentFlowVO{
		ID:                strings.TrimSpace(item.ID),
		SourceType:        item.SourceType,
		SourceID:          strings.TrimSpace(item.SourceID),
		Title:             parentCourseFlowTitle(item.SourceType),
		CreatedAt:         formatParentCourseDateTime(item.CreatedAt),
		Quantity:          item.Quantity,
		Tuition:           item.Tuition,
		QuantityText:      buildParentCourseFlowQuantityText(item.SourceType, chargingMode, item.Quantity, item.Tuition),
		HighlightPositive: parentCourseFlowIsPositive(item.SourceType),
	}
}

func sortParentCourseEnrollmentVOs(items []model.ParentCourseEnrollmentVO) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Status != items[j].Status {
			return items[i].Status < items[j].Status
		}
		if items[i].LowBalance != items[j].LowBalance {
			return items[i].LowBalance
		}
		if items[i].EndDate != items[j].EndDate {
			if items[i].EndDate == "" {
				return false
			}
			if items[j].EndDate == "" {
				return true
			}
			return items[i].EndDate < items[j].EndDate
		}
		return items[i].LessonName < items[j].LessonName
	})
}

func parentCourseBucketKey(item model.TuitionAccountReadingItem) string {
	lessonID := strings.TrimSpace(item.LessonID)
	if lessonID == "" {
		return ""
	}
	return lessonID + ":" + strconv.Itoa(normalizeParentCourseChargingMode(derefParentCourseInt(item.LessonChargingMode)))
}

func normalizeParentCourseFlowPageSize(pageSize int) int {
	if pageSize <= 0 {
		return parentCourseFlowDefaultPageSize
	}
	if pageSize > parentCourseFlowMaxPageSize {
		return parentCourseFlowMaxPageSize
	}
	return pageSize
}

func normalizeParentCourseFlowPageIndex(pageIndex int) int {
	if pageIndex <= 0 {
		return parentCourseFlowDefaultPageIndex
	}
	return pageIndex
}

func normalizeParentCourseChargingMode(mode int) int {
	if mode == 4 {
		return 3
	}
	if mode <= 0 {
		return 1
	}
	return mode
}

func parentCourseChargingModeText(mode int) string {
	switch normalizeParentCourseChargingMode(mode) {
	case 2:
		return "按时段"
	case 3:
		return "按金额"
	default:
		return "按课时"
	}
}

func parentCourseRemainingLabel(mode int) string {
	switch normalizeParentCourseChargingMode(mode) {
	case 2:
		return "剩余时间"
	case 3:
		return "剩余金额"
	default:
		return "剩余课时"
	}
}

func parentCourseStatusText(status int) string {
	switch status {
	case model.TuitionAccountStatusSuspended:
		return "已停课"
	case model.TuitionAccountStatusClosed:
		return "已结课"
	default:
		return "已生效"
	}
}

func parentCourseIsClosed(item model.TuitionAccountReadingItem) bool {
	return derefParentCourseInt(item.Status) == model.TuitionAccountStatusClosed
}

func buildParentCourseRemainingQuantityText(mode int, value float64) string {
	switch normalizeParentCourseChargingMode(mode) {
	case 2:
		return formatParentCourseMetric(value) + "天"
	case 3:
		return buildParentCourseAmountText(value)
	default:
		return formatParentCourseMetric(value) + "课时"
	}
}

func buildParentCourseAmountText(value float64) string {
	return formatParentCourseMetric(value) + " 元"
}

func buildParentCourseValidRangeText(validDate, endDate string) string {
	if validDate == "" && endDate == "" {
		return ""
	}
	if validDate == "" {
		return endDate
	}
	if endDate == "" {
		return validDate
	}
	return validDate + "~" + endDate
}

func buildParentCourseLowBalanceText(mode int, remainingQuantity float64, status int) string {
	if status != model.TuitionAccountStatusActive {
		return ""
	}
	if normalizeParentCourseChargingMode(mode) == 3 {
		return ""
	}
	if remainingQuantity > parentCourseLowBalanceThreshold {
		return ""
	}
	switch normalizeParentCourseChargingMode(mode) {
	case 2:
		return "剩余时间不足，请及时续费"
	default:
		return "剩余课时不足，请及时续费"
	}
}

func buildParentCourseFlowQuantityText(sourceType, chargingMode int, quantity, tuition float64) string {
	valueText := ""
	mode := normalizeParentCourseChargingMode(chargingMode)
	switch mode {
	case 3:
		valueText = buildParentCourseAmountText(absParentCourseMetric(tuition))
	default:
		valueText = buildParentCourseRemainingQuantityText(mode, absParentCourseMetric(quantity))
	}
	if parentCourseFlowIsPositive(sourceType) {
		return "+" + valueText
	}
	return valueText
}

func parentCourseFlowTitle(sourceType int) string {
	switch sourceType {
	case model.TuitionAccountFlowSourceRegistration:
		return "报名"
	case model.TuitionAccountFlowSourceTransferIn:
		return "转入"
	case model.TuitionAccountFlowSourceCrossCampusTransferIn:
		return "跨校区转入"
	case model.TuitionAccountFlowSourceCrossCampusAttendIn:
		return "跨校区就读转入"
	case model.TuitionAccountFlowSourceConsumeReturn:
		return "课消退还"
	case model.TuitionAccountFlowSourceRevokeGraduate:
		return "撤销结课"
	case model.TuitionAccountFlowSourceExpireRollback:
		return "到期回退"
	case model.TuitionAccountFlowSourceRevokeRefundOrder:
		return "撤销退费"
	case model.TuitionAccountFlowSourceRevokeTransferOut:
		return "撤销转出"
	case model.TuitionAccountFlowSourceRevokeImportConsume:
		return "撤销导入课消"
	case model.TuitionAccountFlowSourceRevokeAutoConsume:
		return "撤销自动课消"
	case model.TuitionAccountFlowSourceConsume:
		return "课消"
	case model.TuitionAccountFlowSourceImportConsume:
		return "导入课消"
	case model.TuitionAccountFlowSourceConsumeSupplement:
		return "补扣课消"
	case model.TuitionAccountFlowSourceAutoConsume:
		return "每日自动课消"
	case model.TuitionAccountFlowSourceConsumeArrearsSettlement:
		return "欠费补扣"
	case model.TuitionAccountFlowSourceTransferOut:
		return "转出"
	case model.TuitionAccountFlowSourceCrossCampusTransferOut:
		return "跨校区转出"
	case model.TuitionAccountFlowSourceCrossCampusAttendOut:
		return "跨校区就读转出"
	case model.TuitionAccountFlowSourceGraduate:
		return "结课"
	case model.TuitionAccountFlowSourceExpireGraduate:
		return "到期结课"
	case model.TuitionAccountFlowSourceRefund:
		return "退费"
	case model.TuitionAccountFlowSourceOrderVoid:
		return "订单作废"
	case model.TuitionAccountFlowSourceVoidCrossCampusTransferIn:
		return "撤销跨校区转入"
	case model.TuitionAccountFlowSourceManualCloseCourse:
		return "手动结课"
	default:
		return "账户变动"
	}
}

func parentCourseFlowIsPositive(sourceType int) bool {
	switch sourceType {
	case model.TuitionAccountFlowSourceRegistration,
		model.TuitionAccountFlowSourceTransferIn,
		model.TuitionAccountFlowSourceCrossCampusTransferIn,
		model.TuitionAccountFlowSourceCrossCampusAttendIn,
		model.TuitionAccountFlowSourceConsumeReturn,
		model.TuitionAccountFlowSourceRevokeGraduate,
		model.TuitionAccountFlowSourceExpireRollback,
		model.TuitionAccountFlowSourceRevokeRefundOrder,
		model.TuitionAccountFlowSourceRevokeTransferOut,
		model.TuitionAccountFlowSourceRevokeImportConsume,
		model.TuitionAccountFlowSourceRevokeAutoConsume,
		model.TuitionAccountFlowSourceVoidCrossCampusTransferIn:
		return true
	default:
		return false
	}
}

func derefParentCourseInt(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func formatParentCourseDate(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02")
}

func formatParentCourseDateTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02 15:04")
}

func formatParentCourseMetric(value float64) string {
	return strconv.FormatFloat(roundParentCourseMetric(value), 'f', -1, 64)
}

func roundParentCourseMetric(value float64) float64 {
	text := strconv.FormatFloat(value, 'f', 2, 64)
	parsed, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return value
	}
	return parsed
}

func absParentCourseMetric(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}

func normalizeParentCourseAccountIDs(ids []int64) []int64 {
	result := make([]int64, 0, len(ids))
	seen := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
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
