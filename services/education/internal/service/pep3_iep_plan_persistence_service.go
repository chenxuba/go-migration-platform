package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/repository"
)

const (
	pep3IEPPlanStatusDraft     = "draft"
	pep3IEPPlanStatusConfirmed = "confirmed"
)

func (svc *Service) SavePEP3IEPPlan(userID int64, req model.PEP3IEPPlanSaveRequest) (model.PEP3IEPPlanSavedVO, error) {
	if svc.repo == nil {
		return model.PEP3IEPPlanSavedVO{}, errors.New("assessment repository is not configured")
	}
	if req.ID <= 0 {
		return model.PEP3IEPPlanSavedVO{}, errors.New("invalid assessment record id")
	}
	status, err := normalizePEP3IEPPlanSaveStatus(req.Status)
	if err != nil {
		return model.PEP3IEPPlanSavedVO{}, err
	}
	durationMonths := normalizePEP3IEPPlanDuration(req.DurationMonths)
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return model.PEP3IEPPlanSavedVO{}, err
	}
	record, err := svc.repo.GetAssessmentRecord(context.Background(), instID, req.ID)
	if err != nil {
		return model.PEP3IEPPlanSavedVO{}, err
	}
	plan := normalizePEP3IEPPlanForSave(req.Plan, record, durationMonths)
	if len(plan.Rows) == 0 {
		return model.PEP3IEPPlanSavedVO{}, errors.New("请先生成或填写IEP计划")
	}
	if err := svc.repo.SavePEP3IEPPlan(context.Background(), repository.PEP3IEPPlanEntity{
		InstID:         instID,
		RecordID:       req.ID,
		DurationMonths: durationMonths,
		Status:         status,
		Plan:           plan,
		CreatedBy:      userID,
		UpdatedBy:      userID,
	}); err != nil {
		return model.PEP3IEPPlanSavedVO{}, err
	}
	return svc.GetPEP3IEPPlan(userID, req.ID, durationMonths)
}

func (svc *Service) GetPEP3IEPPlan(userID, recordID int64, durationMonths int) (model.PEP3IEPPlanSavedVO, error) {
	if svc.repo == nil {
		return model.PEP3IEPPlanSavedVO{}, errors.New("assessment repository is not configured")
	}
	if recordID <= 0 {
		return model.PEP3IEPPlanSavedVO{}, errors.New("invalid assessment record id")
	}
	durationMonths = normalizePEP3IEPPlanDuration(durationMonths)
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return model.PEP3IEPPlanSavedVO{}, err
	}
	record, err := svc.repo.GetAssessmentRecord(context.Background(), instID, recordID)
	if err != nil {
		return model.PEP3IEPPlanSavedVO{}, err
	}
	entity, exists, err := svc.repo.GetPEP3IEPPlan(context.Background(), instID, recordID, durationMonths)
	if err != nil {
		return model.PEP3IEPPlanSavedVO{}, err
	}
	if !exists {
		return model.PEP3IEPPlanSavedVO{Exists: false, DurationMonths: durationMonths}, nil
	}
	plan := entity.Plan
	plan = applyPEP3IEPPlanHeaderValues(plan, pep3IEPPlanHeaderValuesForRecord(record))
	return model.PEP3IEPPlanSavedVO{
		Exists:         true,
		Status:         entity.Status,
		DurationMonths: entity.DurationMonths,
		Plan:           &plan,
		UpdatedTime:    formatPEP3IEPPlanUpdatedTime(entity.UpdatedTime),
	}, nil
}

func normalizePEP3IEPPlanSaveStatus(status string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "", pep3IEPPlanStatusDraft:
		return pep3IEPPlanStatusDraft, nil
	case pep3IEPPlanStatusConfirmed:
		return pep3IEPPlanStatusConfirmed, nil
	default:
		return "", errors.New("invalid IEP plan status")
	}
}

func normalizePEP3IEPPlanDuration(durationMonths int) int {
	if durationMonths == 6 {
		return 6
	}
	return 3
}

func normalizePEP3IEPPlanForSave(plan model.PEP3IEPPlanAIResult, record model.AssessmentRecordDetailVO, durationMonths int) model.PEP3IEPPlanAIResult {
	plan.Title = iepPlanTitle(durationMonths)
	plan.Student.Name = firstNonEmptyExportValue(record.StudentName, plan.Student.Name)
	plan.Student.Gender = firstNonEmptyExportValue(record.StudentGender, plan.Student.Gender)
	plan.Student.BirthDate = firstNonEmptyExportValue(formatIEPPlanDate(record.BirthDate), plan.Student.BirthDate)
	plan = applyPEP3IEPPlanHeaderValues(plan, pep3IEPPlanHeaderValuesForRecord(record))
	defaultStartDate, defaultEndDate := iepPlanWholeMonthDateRange(record, durationMonths)
	plan.Meta.StartDate = firstNonEmptyExportValue(plan.Meta.StartDate, defaultStartDate)
	plan.Meta.EndDate = firstNonEmptyExportValue(plan.Meta.EndDate, defaultEndDate)
	plan.Rows = sanitizePEP3IEPPlanRowsForSave(plan.Rows, iepPlanStageDateRanges(record, durationMonths))
	return plan
}

func sanitizePEP3IEPPlanRowsForSave(rows []model.PEP3IEPPlanRow, stageRanges []string) []model.PEP3IEPPlanRow {
	result := make([]model.PEP3IEPPlanRow, 0, len(rows))
	for index, row := range rows {
		domain := strings.TrimSpace(row.Domain)
		longGoal := strings.TrimSpace(row.LongGoal)
		shortGoal := strings.TrimSpace(row.ShortGoal)
		if shortGoal == "" {
			continue
		}
		if domain == "" {
			domain = "综合康复"
		}
		courseForm := normalizeIEPCourseForm(row.CourseForm)
		if courseForm == "" {
			courseForm = "个训"
		}
		startEndDate := strings.TrimSpace(row.StartEndDate)
		if startEndDate == "" {
			startEndDate = stageDateForGoal(stageRanges, index, len(rows))
		}
		result = append(result, model.PEP3IEPPlanRow{
			Domain:       domain,
			LongGoal:     longGoal,
			ShortGoal:    shortGoal,
			CourseForm:   courseForm,
			StartEndDate: startEndDate,
		})
	}
	return result
}

func formatPEP3IEPPlanUpdatedTime(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02 15:04:05")
}
