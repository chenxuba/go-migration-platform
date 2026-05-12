package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/repository"
)

func pep3AssessmentPlanDate(record model.AssessmentRecordDetailVO) string {
	return strings.TrimSpace(formatIEPPlanDate(record.AssessmentDate))
}

func (svc *Service) SyncPEP3IEPPlanPeriod(userID int64, req model.PEP3IEPPlanPeriodSyncRequest) (model.PEP3IEPPlanPeriodSyncVO, error) {
	return svc.syncIEPPlanPeriod(userID, req, pep3ScaleCode)
}

func (svc *Service) SyncERXinIEPPlanPeriod(userID int64, req model.PEP3IEPPlanPeriodSyncRequest) (model.PEP3IEPPlanPeriodSyncVO, error) {
	return svc.syncIEPPlanPeriod(userID, req, erxinScaleCode)
}

func (svc *Service) syncIEPPlanPeriod(userID int64, req model.PEP3IEPPlanPeriodSyncRequest, expectedScaleCode string) (model.PEP3IEPPlanPeriodSyncVO, error) {
	if svc.repo == nil {
		return model.PEP3IEPPlanPeriodSyncVO{}, errors.New("assessment repository is not configured")
	}
	if req.ID <= 0 {
		return model.PEP3IEPPlanPeriodSyncVO{}, errors.New("invalid assessment record id")
	}
	durationMonths, err := validateIEPPlanPeriodDuration(req.DurationMonths)
	if err != nil {
		return model.PEP3IEPPlanPeriodSyncVO{}, err
	}
	sourceDurationMonths := durationMonths
	if req.SourceDurationMonths > 0 {
		sourceDurationMonths, err = validateIEPPlanPeriodDuration(req.SourceDurationMonths)
		if err != nil {
			return model.PEP3IEPPlanPeriodSyncVO{}, err
		}
	}
	periodStart, err := parseIEPPlanPeriodStart(req.StartDate, req.StartMonth)
	if err != nil {
		return model.PEP3IEPPlanPeriodSyncVO{}, err
	}
	ctx := context.Background()
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return model.PEP3IEPPlanPeriodSyncVO{}, err
	}
	record, err := svc.repo.GetAssessmentRecord(ctx, instID, req.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PEP3IEPPlanPeriodSyncVO{}, errors.New("assessment record not found")
		}
		return model.PEP3IEPPlanPeriodSyncVO{}, err
	}
	if strings.TrimSpace(expectedScaleCode) != "" && strings.TrimSpace(record.AssessmentCode) != strings.TrimSpace(expectedScaleCode) {
		return model.PEP3IEPPlanPeriodSyncVO{}, errors.New("assessment record type mismatch")
	}

	var planEntity repository.PEP3IEPPlanEntity
	var exists bool
	if sourceDurationMonths != durationMonths {
		planEntity, exists, err = svc.repo.GetPEP3IEPPlan(ctx, instID, req.ID, sourceDurationMonths)
		if err != nil {
			return model.PEP3IEPPlanPeriodSyncVO{}, err
		}
	}
	if !exists {
		planEntity, exists, err = svc.repo.GetPEP3IEPPlan(ctx, instID, req.ID, durationMonths)
		if err != nil {
			return model.PEP3IEPPlanPeriodSyncVO{}, err
		}
	}
	if !exists || len(planEntity.Plan.Rows) == 0 {
		return model.PEP3IEPPlanPeriodSyncVO{}, errors.New("暂无可调整周期的IEP计划")
	}

	currentTeacherName := svc.currentIEPPlanTeacherName(ctx, userID)
	syncMode := normalizeIEPPlanPeriodSyncMode(req.SyncMode)
	previousPlan := planEntity.Plan
	nextPlan := syncPEP3IEPPlanPeriodDates(planEntity.Plan, record, durationMonths, periodStart)
	executionPlans, err := svc.buildPEP3PeriodExecutionPlanEntities(
		ctx,
		instID,
		req.ID,
		durationMonths,
		sourceDurationMonths,
		previousPlan,
		nextPlan,
		currentTeacherName,
		userID,
		syncMode,
	)
	if err != nil {
		return model.PEP3IEPPlanPeriodSyncVO{}, err
	}
	if err := svc.repo.SavePEP3IEPPlanWithExecutionPlans(ctx, repository.PEP3IEPPlanEntity{
		InstID:         instID,
		RecordID:       req.ID,
		DurationMonths: durationMonths,
		Status:         planEntity.Status,
		Plan:           nextPlan,
		CreatedBy:      userID,
		UpdatedBy:      userID,
	}, executionPlans, sourceDurationMonths); err != nil {
		return model.PEP3IEPPlanPeriodSyncVO{}, err
	}

	iepPlan, err := svc.GetPEP3IEPPlan(userID, req.ID, durationMonths)
	if err != nil {
		return model.PEP3IEPPlanPeriodSyncVO{}, err
	}
	executionSaved, err := svc.GetPEP3ExecutionPlans(userID, req.ID, durationMonths)
	if err != nil {
		return model.PEP3IEPPlanPeriodSyncVO{}, err
	}
	return model.PEP3IEPPlanPeriodSyncVO{
		IEPPlan:        iepPlan,
		ExecutionPlans: executionSaved,
	}, nil
}

const (
	iepPlanPeriodSyncModeDatesOnly         = "dates_only"
	iepPlanPeriodSyncModeSupplementNewWeeks = "supplement_new_weeks"
)

func normalizeIEPPlanPeriodSyncMode(value string) string {
	switch strings.TrimSpace(value) {
	case iepPlanPeriodSyncModeSupplementNewWeeks:
		return iepPlanPeriodSyncModeSupplementNewWeeks
	default:
		return iepPlanPeriodSyncModeDatesOnly
	}
}

func validateIEPPlanPeriodDuration(durationMonths int) (int, error) {
	switch durationMonths {
	case 3, 6:
		return durationMonths, nil
	default:
		return 0, errors.New("计划周期仅支持3个月或6个月")
	}
}

func parseIEPPlanPeriodStart(startDate, startMonth string) (time.Time, error) {
	value := strings.TrimSpace(startDate)
	if value == "" {
		value = strings.TrimSpace(startMonth)
	}
	if value == "" {
		return time.Time{}, errors.New("请选择计划开始日期")
	}
	if len(value) >= len("2006-01") && len(value) < len("2006-01-02") {
		value = value[:len("2006-01")] + "-01"
	} else if len(value) >= len("2006-01-02") {
		value = value[:len("2006-01-02")]
	}
	parsed, err := time.ParseInLocation("2006-01-02", value, time.Local)
	if err != nil {
		return time.Time{}, errors.New("计划开始日期格式应为YYYY-MM-DD")
	}
	return time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.Local), nil
}

func syncPEP3IEPPlanPeriodDates(plan model.PEP3IEPPlanAIResult, record model.AssessmentRecordDetailVO, durationMonths int, periodStart time.Time) model.PEP3IEPPlanAIResult {
	startDate, endDate := iepPlanWholeMonthDateRangeFromStart(periodStart, durationMonths)
	stageRanges := iepPlanStageDateRangesFromStart(periodStart, durationMonths)
	plan.Title = iepPlanTitle(durationMonths)
	plan.Student.Name = firstNonEmptyExportValue(strings.TrimSpace(plan.Student.Name), strings.TrimSpace(record.StudentName))
	plan.Student.Gender = firstNonEmptyExportValue(strings.TrimSpace(plan.Student.Gender), strings.TrimSpace(record.StudentGender))
	plan.Student.BirthDate = firstNonEmptyExportValue(strings.TrimSpace(plan.Student.BirthDate), formatIEPPlanDate(record.BirthDate))
	plan = applyPEP3IEPPlanHeaderValues(plan, pep3IEPPlanHeaderValuesForRecord(record))
	plan.Meta.StartDate = startDate
	plan.Meta.EndDate = endDate
	for index := range plan.Rows {
		plan.Rows[index].Domain = firstNonEmptyExportValue(strings.TrimSpace(plan.Rows[index].Domain), "综合康复")
		plan.Rows[index].CourseForm = firstNonEmptyExportValue(normalizeIEPCourseForm(plan.Rows[index].CourseForm), "个训")
		plan.Rows[index].StartEndDate = firstNonEmptyExportValue(stageDateForGoal(stageRanges, index, len(plan.Rows)), startDate+" - "+endDate)
	}
	return plan
}

func iepPlanWholeMonthDateRangeFromStart(start time.Time, durationMonths int) (string, string) {
	normalizedStart, end := iepPlanDateRangeFromStart(start, durationMonths)
	return normalizedStart.Format("2006-01-02"), end.Format("2006-01-02")
}

func (svc *Service) buildPEP3PeriodExecutionPlanEntities(
	ctx context.Context,
	instID, recordID int64,
	durationMonths, sourceDurationMonths int,
	previousPlan model.PEP3IEPPlanAIResult,
	nextPlan model.PEP3IEPPlanAIResult,
	currentTeacherName string,
	userID int64,
	syncMode string,
) ([]repository.PEP3ExecutionPlanEntity, error) {
	sourceEntitiesByKey := make(map[string]repository.PEP3ExecutionPlanEntity)
	targetEntities, err := svc.repo.ListPEP3ExecutionPlans(ctx, instID, recordID, durationMonths)
	if err != nil {
		return nil, err
	}
	for _, entity := range targetEntities {
		sourceEntitiesByKey[pep3ExecutionPlanEntityKey(entity)] = entity
	}
	if sourceDurationMonths != durationMonths {
		sourceEntities, err := svc.repo.ListPEP3ExecutionPlans(ctx, instID, recordID, sourceDurationMonths)
		if err != nil {
			return nil, err
		}
		for _, entity := range sourceEntities {
			sourceEntitiesByKey[pep3ExecutionPlanEntityKey(entity)] = entity
		}
	}
	sourceEntities := make([]repository.PEP3ExecutionPlanEntity, 0, len(sourceEntitiesByKey))
	for _, entity := range sourceEntitiesByKey {
		sourceEntities = append(sourceEntities, entity)
	}

	result := make([]repository.PEP3ExecutionPlanEntity, 0, len(sourceEntities))
	monthlyPlans := make(map[int]model.PEP3MonthlyPlanAIResult)
	supplementedMonths := make(map[int]struct{})
	resetMonths := make(map[int]struct{})
	monthCount := executionPlanMonthCount(nextPlan, durationMonths)
	if monthCount <= 0 {
		monthCount = 1
	}
	for _, entity := range sourceEntities {
		if strings.ToLower(strings.TrimSpace(entity.PlanType)) != pep3ExecutionPlanTypeMonthly {
			continue
		}
		monthIndex := entity.TargetMonthIndex
		if monthIndex < 1 || monthIndex > monthCount {
			continue
		}
		var plan model.PEP3MonthlyPlanAIResult
		if err := json.Unmarshal(entity.PlanJSON, &plan); err != nil {
			return nil, fmt.Errorf("parse monthly execution plan: %w", err)
		}
		restWeekdays := normalizeExecutionPlanRestWeekdays(plan.RestWeekdays)
		oldTarget := buildExecutionPlanTarget(previousPlan, durationMonths, monthIndex, 0, restWeekdays)
		newTarget := buildExecutionPlanTarget(nextPlan, durationMonths, monthIndex, 0, restWeekdays)
		if monthWeekRangesChanged(oldTarget, newTarget) {
			switch syncMode {
			case iepPlanPeriodSyncModeSupplementNewWeeks:
				supplementedPlan, supplementErr := svc.supplementPEP3MonthlyPlanForPeriodSync(
					ctx,
					userID,
					recordID,
					durationMonths,
					monthIndex,
					nextPlan,
					plan,
					newTarget,
				)
				if supplementErr != nil {
					return nil, supplementErr
				}
				plan = supplementedPlan
				supplementedMonths[monthIndex] = struct{}{}
			default:
				resetMonths[monthIndex] = struct{}{}
				continue
			}
		} else {
			plan = syncPEP3MonthlyPlanPeriodDates(plan, nextPlan, newTarget, currentTeacherName)
		}
		monthlyPlans[newTarget.MonthIndex] = plan
		raw, err := json.Marshal(plan)
		if err != nil {
			return nil, fmt.Errorf("marshal monthly execution plan: %w", err)
		}
		result = append(result, repository.PEP3ExecutionPlanEntity{
			InstID:           instID,
			RecordID:         recordID,
			DurationMonths:   durationMonths,
			PlanType:         pep3ExecutionPlanTypeMonthly,
			TargetMonthIndex: newTarget.MonthIndex,
			TargetWeekIndex:  0,
			PlanJSON:         json.RawMessage(raw),
			CreatedBy:        userID,
			UpdatedBy:        userID,
		})
	}

	for _, entity := range sourceEntities {
		if strings.ToLower(strings.TrimSpace(entity.PlanType)) != pep3ExecutionPlanTypeWeekly {
			continue
		}
		monthIndex := entity.TargetMonthIndex
		if monthIndex < 1 || monthIndex > monthCount {
			continue
		}
		if _, reset := resetMonths[monthIndex]; reset {
			continue
		}
		if _, supplemented := supplementedMonths[monthIndex]; supplemented {
			continue
		}
		var plan model.PEP3WeeklyPlanAIResult
		if err := json.Unmarshal(entity.PlanJSON, &plan); err != nil {
			return nil, fmt.Errorf("parse weekly execution plan: %w", err)
		}
		var monthlyPlan *model.PEP3MonthlyPlanAIResult
		if currentMonthlyPlan, ok := monthlyPlans[monthIndex]; ok {
			monthlyPlan = &currentMonthlyPlan
		}
		restWeekdays := inferWeeklyPlanRestWeekdays(plan)
		if len(plan.RestWeekdays) == 0 && monthlyPlan != nil && len(monthlyPlan.RestWeekdays) > 0 {
			restWeekdays = normalizeExecutionPlanRestWeekdays(monthlyPlan.RestWeekdays)
		}
		oldTarget := buildExecutionPlanTarget(previousPlan, durationMonths, monthIndex, entity.TargetWeekIndex, restWeekdays)
		target := buildExecutionPlanTarget(nextPlan, durationMonths, monthIndex, entity.TargetWeekIndex, restWeekdays)
		if monthWeekRangesChanged(
			buildExecutionPlanTarget(previousPlan, durationMonths, monthIndex, 0, restWeekdays),
			buildExecutionPlanTarget(nextPlan, durationMonths, monthIndex, 0, restWeekdays),
		) || weeklyTargetChanged(oldTarget, target) {
			continue
		}
		if currentMonthlyPlan, ok := monthlyPlans[target.MonthIndex]; ok {
			monthlyPlan = &currentMonthlyPlan
		}
		plan = normalizePEP3WeeklyExecutionPlan(plan, nextPlan, monthlyPlan, target, currentTeacherName)
		raw, err := json.Marshal(plan)
		if err != nil {
			return nil, fmt.Errorf("marshal weekly execution plan: %w", err)
		}
		result = append(result, repository.PEP3ExecutionPlanEntity{
			InstID:           instID,
			RecordID:         recordID,
			DurationMonths:   durationMonths,
			PlanType:         pep3ExecutionPlanTypeWeekly,
			TargetMonthIndex: target.MonthIndex,
			TargetWeekIndex:  target.WeekIndex,
			PlanJSON:         json.RawMessage(raw),
			CreatedBy:        userID,
			UpdatedBy:        userID,
		})
	}
	return result, nil
}

func monthWeekRangesChanged(left, right pep3ExecutionPlanTarget) bool {
	if len(left.WeekRanges) != len(right.WeekRanges) {
		return true
	}
	for index := range left.WeekRanges {
		if strings.TrimSpace(left.WeekRanges[index]) != strings.TrimSpace(right.WeekRanges[index]) {
			return true
		}
	}
	return false
}

func weeklyTargetChanged(left, right pep3ExecutionPlanTarget) bool {
	if strings.TrimSpace(left.WeekRangeText) != strings.TrimSpace(right.WeekRangeText) {
		return true
	}
	if len(left.WeekDates) != len(right.WeekDates) {
		return true
	}
	for index := range left.WeekDates {
		if strings.TrimSpace(left.WeekDates[index]) != strings.TrimSpace(right.WeekDates[index]) {
			return true
		}
	}
	return false
}

func (svc *Service) supplementPEP3MonthlyPlanForPeriodSync(
	ctx context.Context,
	userID, recordID int64,
	durationMonths, targetMonthIndex int,
	sourcePlan model.PEP3IEPPlanAIResult,
	existingPlan model.PEP3MonthlyPlanAIResult,
	target pep3ExecutionPlanTarget,
) (model.PEP3MonthlyPlanAIResult, error) {
	generatedAny, err := svc.GeneratePEP3ExecutionPlanWithAI(ctx, userID, model.PEP3ExecutionPlanGenerateRequest{
		ID:               recordID,
		DurationMonths:   durationMonths,
		PlanType:         pep3ExecutionPlanTypeMonthly,
		TargetMonthIndex: targetMonthIndex,
		RestWeekdays:     target.RestWeekdays,
		SourcePlan:       sourcePlan,
	})
	if err != nil {
		return model.PEP3MonthlyPlanAIResult{}, fmt.Errorf("补齐新增周训练内容失败：%w", err)
	}
	generatedPlan, ok := generatedAny.(model.PEP3MonthlyPlanAIResult)
	if !ok {
		return model.PEP3MonthlyPlanAIResult{}, errors.New("补齐新增周训练内容失败：AI返回结果类型异常")
	}
	return mergeSupplementedPEP3MonthlyPlan(existingPlan, generatedPlan), nil
}

func mergeSupplementedPEP3MonthlyPlan(
	existingPlan model.PEP3MonthlyPlanAIResult,
	generatedPlan model.PEP3MonthlyPlanAIResult,
) model.PEP3MonthlyPlanAIResult {
	result := generatedPlan
	existingRowsByKey := make(map[string]model.PEP3MonthlyPlanRow, len(existingPlan.Rows))
	for _, row := range existingPlan.Rows {
		existingRowsByKey[monthlyPlanRowIdentity(row)] = row
	}
	for rowIndex := range result.Rows {
		row := result.Rows[rowIndex]
		existingRow, ok := existingRowsByKey[monthlyPlanRowIdentity(row)]
		if !ok {
			continue
		}
		existingItemsByRange := make(map[string]model.PEP3MonthlyTrainingItem, len(existingRow.TrainingItems))
		for _, item := range existingRow.TrainingItems {
			rangeText := strings.TrimSpace(item.StartEndDate)
			if rangeText == "" || strings.TrimSpace(item.Content) == "" {
				continue
			}
			existingItemsByRange[rangeText] = item
		}
		if strings.TrimSpace(existingRow.CourseForm) != "" {
			row.CourseForm = existingRow.CourseForm
		}
		for itemIndex := range row.TrainingItems {
			rangeText := strings.TrimSpace(row.TrainingItems[itemIndex].StartEndDate)
			if existingItem, exists := existingItemsByRange[rangeText]; exists {
				row.TrainingItems[itemIndex] = existingItem
			}
		}
		result.Rows[rowIndex] = row
	}
	return result
}

func monthlyPlanRowIdentity(row model.PEP3MonthlyPlanRow) string {
	return strings.TrimSpace(row.Domain) + "\n" +
		strings.TrimSpace(row.LongGoal) + "\n" +
		strings.TrimSpace(row.ShortGoal)
}

func pep3ExecutionPlanEntityKey(entity repository.PEP3ExecutionPlanEntity) string {
	return strings.ToLower(strings.TrimSpace(entity.PlanType)) + ":" + fmt.Sprint(entity.TargetMonthIndex) + ":" + fmt.Sprint(entity.TargetWeekIndex)
}

func syncPEP3MonthlyPlanPeriodDates(plan model.PEP3MonthlyPlanAIResult, sourcePlan model.PEP3IEPPlanAIResult, target pep3ExecutionPlanTarget, currentTeacherName string) model.PEP3MonthlyPlanAIResult {
	plan = normalizePEP3MonthlyExecutionPlan(plan, sourcePlan, target, currentTeacherName)
	plan.Title = fmt.Sprintf("康复教学%s计划", target.MonthLabel)
	plan.Meta.PlanDate = firstNonEmptyExportValue(strings.TrimSpace(sourcePlan.Meta.PlanDate), strings.TrimSpace(plan.Meta.PlanDate), time.Now().Format("2006-01-02"))
	plan.Meta.StartDate = target.StartDate
	plan.Meta.EndDate = target.EndDate
	plan.Meta.MonthLabel = target.MonthLabel
	plan.Meta.SourceTitle = firstNonEmptyExportValue(strings.TrimSpace(sourcePlan.Title), strings.TrimSpace(plan.Meta.SourceTitle))
	for rowIndex := range plan.Rows {
		items := plan.Rows[rowIndex].TrainingItems
		for itemIndex := range items {
			items[itemIndex].StartEndDate = firstNonEmptyExportValue(monthlyItemDateRangeForTarget(target, itemIndex, len(items)), target.StartDate+" - "+target.EndDate)
		}
		plan.Rows[rowIndex].TrainingItems = items
	}
	return plan
}
