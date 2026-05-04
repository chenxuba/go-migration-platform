package service

import (
	"context"
	"encoding/json"
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
