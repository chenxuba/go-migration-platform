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

func (svc *Service) syncPEP3SavedPlanDatesWithAssessmentDate(ctx context.Context, instID, recordID int64, assessmentDate time.Time, userID int64) error {
	planDate := strings.TrimSpace(assessmentDate.Format("2006-01-02"))
	if assessmentDate.IsZero() || planDate == "" {
		return nil
	}
	for _, durationMonths := range []int{3, 6} {
		if err := svc.syncPEP3SavedIEPPlanDate(ctx, instID, recordID, durationMonths, planDate, userID); err != nil {
			return err
		}
		if err := svc.syncPEP3SavedMonthlyPlanDates(ctx, instID, recordID, durationMonths, planDate, userID); err != nil {
			return err
		}
	}
	return nil
}

func pep3AssessmentPlanDate(record model.AssessmentRecordDetailVO) string {
	return strings.TrimSpace(formatIEPPlanDate(record.AssessmentDate))
}

func syncPEP3IEPPlanDateForDisplay(plan model.PEP3IEPPlanAIResult, record model.AssessmentRecordDetailVO) model.PEP3IEPPlanAIResult {
	if planDate := pep3AssessmentPlanDate(record); planDate != "" {
		plan.Meta.PlanDate = planDate
	} else if strings.TrimSpace(plan.Meta.PlanDate) == "" {
		plan.Meta.PlanDate = time.Now().Format("2006-01-02")
	}
	return plan
}

func syncPEP3MonthlyPlanDateForDisplay(plan model.PEP3MonthlyPlanAIResult, record model.AssessmentRecordDetailVO) model.PEP3MonthlyPlanAIResult {
	if planDate := pep3AssessmentPlanDate(record); planDate != "" {
		plan.Meta.PlanDate = planDate
	} else if strings.TrimSpace(plan.Meta.PlanDate) == "" {
		plan.Meta.PlanDate = time.Now().Format("2006-01-02")
	}
	return plan
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
	nextPlan := syncPEP3IEPPlanPeriodDates(planEntity.Plan, record, currentTeacherName, durationMonths, periodStart)
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
		return time.Time{}, errors.New("请选择计划开始月份")
	}
	if len(value) >= len("2006-01") && len(value) < len("2006-01-02") {
		value = value[:len("2006-01")] + "-01"
	} else if len(value) >= len("2006-01-02") {
		value = value[:len("2006-01-02")]
	}
	parsed, err := time.ParseInLocation("2006-01-02", value, time.Local)
	if err != nil {
		return time.Time{}, errors.New("计划开始月份格式应为YYYY-MM")
	}
	return time.Date(parsed.Year(), parsed.Month(), 1, 0, 0, 0, 0, time.Local), nil
}

func syncPEP3IEPPlanPeriodDates(plan model.PEP3IEPPlanAIResult, record model.AssessmentRecordDetailVO, currentTeacherName string, durationMonths int, periodStart time.Time) model.PEP3IEPPlanAIResult {
	startDate, endDate := iepPlanWholeMonthDateRangeFromStart(periodStart, durationMonths)
	stageRanges := iepPlanStageDateRangesFromStart(periodStart, durationMonths)
	plan.Title = iepPlanTitle(durationMonths)
	plan.Student.Name = firstNonEmptyExportValue(strings.TrimSpace(plan.Student.Name), strings.TrimSpace(record.StudentName))
	plan.Student.Gender = firstNonEmptyExportValue(strings.TrimSpace(plan.Student.Gender), strings.TrimSpace(record.StudentGender))
	plan.Student.BirthDate = firstNonEmptyExportValue(strings.TrimSpace(plan.Student.BirthDate), formatIEPPlanDate(record.BirthDate))
	plan = syncPEP3IEPPlanDateForDisplay(plan, record)
	plan.Meta.Participant = firstNonEmptyExportValue(strings.TrimSpace(plan.Meta.Participant), strings.TrimSpace(currentTeacherName), strings.TrimSpace(record.ExaminerName))
	plan.Meta.Implementer = firstNonEmptyExportValue(strings.TrimSpace(plan.Meta.Implementer), strings.TrimSpace(currentTeacherName), strings.TrimSpace(record.ExaminerName))
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
	normalizedStart := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.Local)
	end := normalizedStart.AddDate(0, durationMonths, 0).AddDate(0, 0, -1)
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
	current := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.Local)
	for index := 0; index < stageCount; index++ {
		months := monthBase
		if index < monthRemainder {
			months++
		}
		if months <= 0 {
			months = 1
		}
		end := current.AddDate(0, months, 0).AddDate(0, 0, -1)
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

func (svc *Service) syncPEP3SavedIEPPlanDate(ctx context.Context, instID, recordID int64, durationMonths int, planDate string, userID int64) error {
	entity, exists, err := svc.repo.GetPEP3IEPPlan(ctx, instID, recordID, durationMonths)
	if err != nil || !exists {
		return err
	}
	if strings.TrimSpace(entity.Plan.Meta.PlanDate) == planDate {
		return nil
	}
	entity.Plan.Meta.PlanDate = planDate
	return svc.repo.SavePEP3IEPPlan(ctx, repository.PEP3IEPPlanEntity{
		InstID:         instID,
		RecordID:       recordID,
		DurationMonths: durationMonths,
		Status:         entity.Status,
		Plan:           entity.Plan,
		CreatedBy:      userID,
		UpdatedBy:      userID,
	})
}

func (svc *Service) syncPEP3SavedMonthlyPlanDates(ctx context.Context, instID, recordID int64, durationMonths int, planDate string, userID int64) error {
	entities, err := svc.repo.ListPEP3ExecutionPlans(ctx, instID, recordID, durationMonths)
	if err != nil {
		return err
	}
	for _, entity := range entities {
		if strings.ToLower(strings.TrimSpace(entity.PlanType)) != pep3ExecutionPlanTypeMonthly {
			continue
		}
		var plan model.PEP3MonthlyPlanAIResult
		if err := json.Unmarshal(entity.PlanJSON, &plan); err != nil {
			return err
		}
		if strings.TrimSpace(plan.Meta.PlanDate) == planDate {
			continue
		}
		plan.Meta.PlanDate = planDate
		if err := svc.repo.SavePEP3ExecutionPlan(ctx, repository.PEP3ExecutionPlanEntity{
			InstID:           instID,
			RecordID:         recordID,
			DurationMonths:   durationMonths,
			PlanType:         entity.PlanType,
			TargetMonthIndex: entity.TargetMonthIndex,
			TargetWeekIndex:  entity.TargetWeekIndex,
			CreatedBy:        userID,
			UpdatedBy:        userID,
		}, plan); err != nil {
			return err
		}
	}
	return nil
}
