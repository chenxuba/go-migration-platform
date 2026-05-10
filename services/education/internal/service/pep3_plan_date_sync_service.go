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
	nextPlan := syncPEP3IEPPlanPeriodDates(planEntity.Plan, record, durationMonths, periodStart)
	executionPlans, err := svc.buildPEP3PeriodExecutionPlanEntities(ctx, instID, req.ID, durationMonths, sourceDurationMonths, nextPlan, currentTeacherName, userID)
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
	if durationMonths <= 0 {
		durationMonths = 6
	}
	normalizedStart := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.Local)
	end := time.Date(normalizedStart.Year(), normalizedStart.Month()+time.Month(durationMonths), 0, 0, 0, 0, 0, time.Local)
	return normalizedStart.Format("2006-01-02"), end.Format("2006-01-02")
}

func iepPlanStageDateRangesFromStart(start time.Time, durationMonths int) []string {
	if durationMonths <= 0 {
		durationMonths = 6
	}
	stageCount := 3
	monthBase := durationMonths / stageCount
	monthRemainder := durationMonths % stageCount
	ranges := make([]string, 0, stageCount)
	current := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.Local)
	_, periodEndText := iepPlanWholeMonthDateRangeFromStart(start, durationMonths)
	periodEnd, _ := time.ParseInLocation("2006-01-02", periodEndText, time.Local)
	for index := 0; index < stageCount; index++ {
		months := monthBase
		if index < monthRemainder {
			months++
		}
		if months <= 0 {
			months = 1
		}
		end := time.Date(current.Year(), current.Month()+time.Month(months), 0, 0, 0, 0, 0, time.Local)
		if !periodEnd.IsZero() && end.After(periodEnd) {
			end = periodEnd
		}
		ranges = append(ranges, current.Format("2006-01-02")+" - "+end.Format("2006-01-02"))
		current = end.AddDate(0, 0, 1)
	}
	return ranges
}

func (svc *Service) buildPEP3PeriodExecutionPlanEntities(ctx context.Context, instID, recordID int64, durationMonths, sourceDurationMonths int, sourcePlan model.PEP3IEPPlanAIResult, currentTeacherName string, userID int64) ([]repository.PEP3ExecutionPlanEntity, error) {
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
	for _, entity := range sourceEntities {
		if strings.ToLower(strings.TrimSpace(entity.PlanType)) != pep3ExecutionPlanTypeMonthly {
			continue
		}
		monthIndex := entity.TargetMonthIndex
		if monthIndex < 1 || monthIndex > durationMonths {
			continue
		}
		var plan model.PEP3MonthlyPlanAIResult
		if err := json.Unmarshal(entity.PlanJSON, &plan); err != nil {
			return nil, fmt.Errorf("parse monthly execution plan: %w", err)
		}
		target := buildExecutionPlanTarget(sourcePlan, durationMonths, monthIndex, 0)
		plan = syncPEP3MonthlyPlanPeriodDates(plan, sourcePlan, target, currentTeacherName)
		monthlyPlans[target.MonthIndex] = plan
		raw, err := json.Marshal(plan)
		if err != nil {
			return nil, fmt.Errorf("marshal monthly execution plan: %w", err)
		}
		result = append(result, repository.PEP3ExecutionPlanEntity{
			InstID:           instID,
			RecordID:         recordID,
			DurationMonths:   durationMonths,
			PlanType:         pep3ExecutionPlanTypeMonthly,
			TargetMonthIndex: target.MonthIndex,
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
		if monthIndex < 1 || monthIndex > durationMonths {
			continue
		}
		var plan model.PEP3WeeklyPlanAIResult
		if err := json.Unmarshal(entity.PlanJSON, &plan); err != nil {
			return nil, fmt.Errorf("parse weekly execution plan: %w", err)
		}
		target := buildExecutionPlanTarget(sourcePlan, durationMonths, monthIndex, entity.TargetWeekIndex)
		var monthlyPlan *model.PEP3MonthlyPlanAIResult
		if currentMonthlyPlan, ok := monthlyPlans[target.MonthIndex]; ok {
			monthlyPlan = &currentMonthlyPlan
		}
		plan = normalizePEP3WeeklyExecutionPlan(plan, sourcePlan, monthlyPlan, target, currentTeacherName)
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
			items[itemIndex].StartEndDate = firstNonEmptyExportValue(monthlyItemDateRange(target.StartDate, target.EndDate, itemIndex, len(items)), target.StartDate+" - "+target.EndDate)
		}
		plan.Rows[rowIndex].TrainingItems = items
	}
	return plan
}
