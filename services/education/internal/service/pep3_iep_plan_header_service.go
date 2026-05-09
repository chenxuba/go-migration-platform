package service

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/repository"
)

type pep3IEPPlanHeaderValues struct {
	PlanDate    string
	Participant string
	Implementer string
}

func pep3IEPPlanHeaderValuesForRecord(record model.AssessmentRecordDetailVO) pep3IEPPlanHeaderValues {
	return pep3IEPPlanHeaderValues{
		PlanDate:    pep3AssessmentPlanDate(record),
		Participant: strings.TrimSpace(record.ExaminerName),
		Implementer: pep3AssessmentOriginalExaminerName(record),
	}
}

func pep3IEPPlanHeaderValuesForRecordConfig(record model.AssessmentRecordDetailVO, examinerName string, assessmentDate time.Time) pep3IEPPlanHeaderValues {
	planDate := ""
	if !assessmentDate.IsZero() {
		planDate = assessmentDate.Format("2006-01-02")
	}
	return pep3IEPPlanHeaderValues{
		PlanDate:    strings.TrimSpace(planDate),
		Participant: strings.TrimSpace(examinerName),
		Implementer: firstNonEmptyExportValue(
			firstPEP3AssessmentExaminerName(pep3AssessmentInputExaminerName(record)),
			firstPEP3AssessmentExaminerName(record.ExaminerName),
			firstPEP3AssessmentExaminerName(examinerName),
		),
	}
}

func pep3AssessmentOriginalExaminerName(record model.AssessmentRecordDetailVO) string {
	return firstNonEmptyExportValue(
		firstPEP3AssessmentExaminerName(pep3AssessmentInputExaminerName(record)),
		firstPEP3AssessmentExaminerName(record.ExaminerName),
	)
}

func pep3AssessmentInputExaminerName(record model.AssessmentRecordDetailVO) string {
	if len(record.InputJSON) == 0 {
		return ""
	}
	var input map[string]json.RawMessage
	if err := json.Unmarshal(record.InputJSON, &input); err != nil {
		return ""
	}
	raw, ok := input["examinerName"]
	if !ok || len(raw) == 0 {
		return ""
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return strings.TrimSpace(text)
	}
	var values []string
	if err := json.Unmarshal(raw, &values); err == nil {
		return strings.Join(trimNonEmptyStrings(values), "、")
	}
	return ""
}

func firstPEP3AssessmentExaminerName(value string) string {
	for _, item := range strings.FieldsFunc(value, func(r rune) bool {
		return r == '、' || r == ',' || r == '，' || r == ';' || r == '；'
	}) {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			return trimmed
		}
	}
	return strings.TrimSpace(value)
}

func trimNonEmptyStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func applyPEP3IEPPlanHeaderValues(plan model.PEP3IEPPlanAIResult, header pep3IEPPlanHeaderValues) model.PEP3IEPPlanAIResult {
	if header.PlanDate != "" {
		plan.Meta.PlanDate = header.PlanDate
	} else if strings.TrimSpace(plan.Meta.PlanDate) == "" {
		plan.Meta.PlanDate = time.Now().Format("2006-01-02")
	}
	if header.Participant != "" {
		plan.Meta.Participant = header.Participant
	}
	if header.Implementer != "" {
		plan.Meta.Implementer = header.Implementer
	}
	return plan
}

func applyPEP3MonthlyPlanHeaderValues(plan model.PEP3MonthlyPlanAIResult, header pep3IEPPlanHeaderValues) model.PEP3MonthlyPlanAIResult {
	if header.PlanDate != "" {
		plan.Meta.PlanDate = header.PlanDate
	} else if strings.TrimSpace(plan.Meta.PlanDate) == "" {
		plan.Meta.PlanDate = time.Now().Format("2006-01-02")
	}
	if header.Participant != "" {
		plan.Meta.Participant = header.Participant
	}
	if header.Implementer != "" {
		plan.Meta.Implementer = header.Implementer
	}
	return plan
}

func (svc *Service) syncPEP3SavedPlanHeadersWithRecordConfig(ctx context.Context, instID, recordID int64, header pep3IEPPlanHeaderValues, userID int64) error {
	for _, durationMonths := range []int{3, 6} {
		if err := svc.syncPEP3SavedIEPPlanHeader(ctx, instID, recordID, durationMonths, header, userID); err != nil {
			return err
		}
		if err := svc.syncPEP3SavedMonthlyPlanHeaders(ctx, instID, recordID, durationMonths, header, userID); err != nil {
			return err
		}
	}
	return nil
}

func (svc *Service) syncPEP3SavedIEPPlanHeader(ctx context.Context, instID, recordID int64, durationMonths int, header pep3IEPPlanHeaderValues, userID int64) error {
	entity, exists, err := svc.repo.GetPEP3IEPPlan(ctx, instID, recordID, durationMonths)
	if err != nil || !exists {
		return err
	}
	nextPlan := applyPEP3IEPPlanHeaderValues(entity.Plan, header)
	if entity.Plan.Meta.PlanDate == nextPlan.Meta.PlanDate &&
		entity.Plan.Meta.Participant == nextPlan.Meta.Participant &&
		entity.Plan.Meta.Implementer == nextPlan.Meta.Implementer {
		return nil
	}
	return svc.repo.SavePEP3IEPPlan(ctx, repository.PEP3IEPPlanEntity{
		InstID:         instID,
		RecordID:       recordID,
		DurationMonths: durationMonths,
		Status:         entity.Status,
		Plan:           nextPlan,
		CreatedBy:      userID,
		UpdatedBy:      userID,
	})
}

func (svc *Service) syncPEP3SavedMonthlyPlanHeaders(ctx context.Context, instID, recordID int64, durationMonths int, header pep3IEPPlanHeaderValues, userID int64) error {
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
		nextPlan := applyPEP3MonthlyPlanHeaderValues(plan, header)
		if plan.Meta.PlanDate == nextPlan.Meta.PlanDate &&
			plan.Meta.Participant == nextPlan.Meta.Participant &&
			plan.Meta.Implementer == nextPlan.Meta.Implementer {
			continue
		}
		if err := svc.repo.SavePEP3ExecutionPlan(ctx, repository.PEP3ExecutionPlanEntity{
			InstID:           instID,
			RecordID:         recordID,
			DurationMonths:   durationMonths,
			PlanType:         entity.PlanType,
			TargetMonthIndex: entity.TargetMonthIndex,
			TargetWeekIndex:  entity.TargetWeekIndex,
			CreatedBy:        userID,
			UpdatedBy:        userID,
		}, nextPlan); err != nil {
			return err
		}
	}
	return nil
}
