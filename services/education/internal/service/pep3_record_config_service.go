package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

type PEP3AssessmentRecordConfigInput struct {
	ExaminerName   string
	AssessmentDate time.Time
}

func (svc *Service) UpdatePEP3AssessmentRecordConfig(userID, recordID int64, input PEP3AssessmentRecordConfigInput) (model.AssessmentRecordDetailVO, error) {
	if svc.repo == nil {
		return model.AssessmentRecordDetailVO{}, errors.New("assessment repository is not configured")
	}
	examinerName := strings.TrimSpace(input.ExaminerName)
	if examinerName == "" {
		return model.AssessmentRecordDetailVO{}, errors.New("评估老师不能为空")
	}
	if input.AssessmentDate.IsZero() {
		return model.AssessmentRecordDetailVO{}, errors.New("评估日期不能为空")
	}

	instID, operatorID, _, err := svc.pep3AssessmentActor(userID, "")
	if err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	record, err := svc.repo.GetAssessmentRecord(context.Background(), instID, recordID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.AssessmentRecordDetailVO{}, errors.New("assessment record not found")
		}
		return model.AssessmentRecordDetailVO{}, err
	}
	if strings.TrimSpace(record.AssessmentCode) != pep3ScaleCode {
		return model.AssessmentRecordDetailVO{}, errors.New("assessment record is not PEP-3")
	}

	assessmentDate := dateOnlyPEP3RecordConfig(input.AssessmentDate)
	if err := svc.repo.UpdateAssessmentRecordConfig(context.Background(), instID, recordID, examinerName, assessmentDate); err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	if err := svc.syncPEP3SubmittedDraftRecordConfig(instID, operatorID, recordID, examinerName, assessmentDate); err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	if err := svc.syncPEP3SavedPlanDatesWithAssessmentDate(context.Background(), instID, recordID, assessmentDate, userID); err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	return svc.repo.GetAssessmentRecord(context.Background(), instID, recordID)
}

func (svc *Service) syncPEP3SubmittedDraftRecordConfig(instID, operatorID, recordID int64, examinerName string, assessmentDate time.Time) error {
	draft, err := svc.repo.GetAssessmentDraftBySubmittedRecordID(context.Background(), instID, recordID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	return svc.repo.UpdateAssessmentDraftConfigIncludingSubmitted(context.Background(), instID, draft.ID, examinerName, assessmentDate, operatorID)
}

func dateOnlyPEP3RecordConfig(value time.Time) time.Time {
	if value.IsZero() {
		return value
	}
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.Local)
}
