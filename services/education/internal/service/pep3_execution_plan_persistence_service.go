package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/repository"
)

const (
	pep3ExecutionPlanTypeMonthly = "monthly"
	pep3ExecutionPlanTypeWeekly  = "weekly"
)

func (svc *Service) SavePEP3ExecutionPlan(userID int64, req model.PEP3ExecutionPlanSaveRequest) (model.PEP3ExecutionPlanSavedVO, error) {
	if svc.repo == nil {
		return model.PEP3ExecutionPlanSavedVO{}, errors.New("assessment repository is not configured")
	}
	if req.ID <= 0 {
		return model.PEP3ExecutionPlanSavedVO{}, errors.New("invalid assessment record id")
	}
	durationMonths := normalizePEP3IEPPlanDuration(req.DurationMonths)
	planType, err := normalizePEP3ExecutionPlanType(req.PlanType)
	if err != nil {
		return model.PEP3ExecutionPlanSavedVO{}, err
	}
	targetMonthIndex := normalizePEP3ExecutionMonthIndex(req.TargetMonthIndex, durationMonths)
	targetWeekIndex := normalizePEP3ExecutionWeekIndex(req.TargetWeekIndex, planType)
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return model.PEP3ExecutionPlanSavedVO{}, err
	}
	record, err := svc.repo.GetAssessmentRecord(context.Background(), instID, req.ID)
	if err != nil {
		return model.PEP3ExecutionPlanSavedVO{}, err
	}

	var plan any
	switch planType {
	case pep3ExecutionPlanTypeMonthly:
		if req.MonthlyPlan == nil || len(req.MonthlyPlan.Rows) == 0 {
			return model.PEP3ExecutionPlanSavedVO{}, errors.New("暂无可保存的月度计划")
		}
		monthlyPlan := applyPEP3MonthlyPlanHeaderValues(*req.MonthlyPlan, pep3IEPPlanHeaderValuesForRecord(record))
		plan = monthlyPlan
	case pep3ExecutionPlanTypeWeekly:
		if req.WeeklyPlan == nil || len(req.WeeklyPlan.Rows) == 0 {
			return model.PEP3ExecutionPlanSavedVO{}, errors.New("暂无可保存的周计划")
		}
		plan = *req.WeeklyPlan
	}
	if err := svc.repo.SavePEP3ExecutionPlan(context.Background(), repository.PEP3ExecutionPlanEntity{
		InstID:           instID,
		RecordID:         req.ID,
		DurationMonths:   durationMonths,
		PlanType:         planType,
		TargetMonthIndex: targetMonthIndex,
		TargetWeekIndex:  targetWeekIndex,
		CreatedBy:        userID,
		UpdatedBy:        userID,
	}, plan); err != nil {
		return model.PEP3ExecutionPlanSavedVO{}, err
	}
	if planType == pep3ExecutionPlanTypeMonthly && !req.PreserveWeeklyPlans {
		if err := svc.repo.DeletePEP3WeeklyExecutionPlansForMonth(context.Background(), instID, req.ID, durationMonths, targetMonthIndex); err != nil {
			return model.PEP3ExecutionPlanSavedVO{}, err
		}
	}
	return svc.GetPEP3ExecutionPlans(userID, req.ID, durationMonths)
}

func (svc *Service) GetPEP3ExecutionPlans(userID, recordID int64, durationMonths int) (model.PEP3ExecutionPlanSavedVO, error) {
	if svc.repo == nil {
		return model.PEP3ExecutionPlanSavedVO{}, errors.New("assessment repository is not configured")
	}
	if recordID <= 0 {
		return model.PEP3ExecutionPlanSavedVO{}, errors.New("invalid assessment record id")
	}
	durationMonths = normalizePEP3IEPPlanDuration(durationMonths)
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return model.PEP3ExecutionPlanSavedVO{}, err
	}
	record, err := svc.repo.GetAssessmentRecord(context.Background(), instID, recordID)
	if err != nil {
		return model.PEP3ExecutionPlanSavedVO{}, err
	}
	entities, err := svc.repo.ListPEP3ExecutionPlans(context.Background(), instID, recordID, durationMonths)
	if err != nil {
		return model.PEP3ExecutionPlanSavedVO{}, err
	}
	result := model.PEP3ExecutionPlanSavedVO{
		Exists:         len(entities) > 0,
		DurationMonths: durationMonths,
		MonthlyPlans:   make([]model.PEP3MonthlyExecutionPlanSaved, 0),
		WeeklyPlans:    make([]model.PEP3WeeklyExecutionPlanSaved, 0),
	}
	for _, entity := range entities {
		switch strings.ToLower(strings.TrimSpace(entity.PlanType)) {
		case pep3ExecutionPlanTypeMonthly:
			var plan model.PEP3MonthlyPlanAIResult
			if err := json.Unmarshal(entity.PlanJSON, &plan); err != nil {
				return model.PEP3ExecutionPlanSavedVO{}, err
			}
			plan = applyPEP3MonthlyPlanHeaderValues(plan, pep3IEPPlanHeaderValuesForRecord(record))
			result.MonthlyPlans = append(result.MonthlyPlans, model.PEP3MonthlyExecutionPlanSaved{
				TargetMonthIndex: entity.TargetMonthIndex,
				Plan:             plan,
				UpdatedTime:      formatPEP3IEPPlanUpdatedTime(entity.UpdatedTime),
			})
		case pep3ExecutionPlanTypeWeekly:
			var plan model.PEP3WeeklyPlanAIResult
			if err := json.Unmarshal(entity.PlanJSON, &plan); err != nil {
				return model.PEP3ExecutionPlanSavedVO{}, err
			}
			result.WeeklyPlans = append(result.WeeklyPlans, model.PEP3WeeklyExecutionPlanSaved{
				TargetMonthIndex: entity.TargetMonthIndex,
				TargetWeekIndex:  entity.TargetWeekIndex,
				Plan:             plan,
				UpdatedTime:      formatPEP3IEPPlanUpdatedTime(entity.UpdatedTime),
			})
		}
	}
	return result, nil
}

func normalizePEP3ExecutionPlanType(planType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(planType)) {
	case pep3ExecutionPlanTypeMonthly:
		return pep3ExecutionPlanTypeMonthly, nil
	case pep3ExecutionPlanTypeWeekly:
		return pep3ExecutionPlanTypeWeekly, nil
	default:
		return "", errors.New("invalid execution plan type")
	}
}

func normalizePEP3ExecutionMonthIndex(value, durationMonths int) int {
	if value < 1 {
		value = 1
	}
	if value > durationMonths {
		value = durationMonths
	}
	return value
}

func normalizePEP3ExecutionWeekIndex(value int, planType string) int {
	if planType != pep3ExecutionPlanTypeWeekly {
		return 0
	}
	if value < 1 {
		return 1
	}
	if value > 6 {
		return 6
	}
	return value
}
