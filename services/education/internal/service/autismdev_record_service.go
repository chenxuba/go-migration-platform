package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"go-migration-platform/pkg/autismdevscore"
	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/repository"
)

type AutismDevAssessmentRecordSaveInput struct {
	StudentID     int64
	StudentName   string
	ExaminerName  string
	Remark        string
	ScoreInput    autismdevscore.AssessmentInput
	InputSnapshot any
}

type AutismDevAssessmentRecordConfigInput struct {
	ExaminerName   string
	AssessmentDate time.Time
}

func (svc *Service) CreateAutismDevAssessmentRecord(userID int64, input AutismDevAssessmentRecordSaveInput) (model.AssessmentRecordDetailVO, error) {
	if svc.repo == nil {
		return model.AssessmentRecordDetailVO{}, errors.New("assessment repository is not configured")
	}
	instID, examinerID, examinerName, err := svc.pep3AssessmentActor(userID, input.ExaminerName)
	if err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	if err := svc.validatePEP3AssessmentStudent(instID, input.StudentID, input.StudentName); err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}

	scoreResult, err := svc.ScoreAutismDev(input.ScoreInput)
	if err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	if !scoreResult.Result.Complete {
		return model.AssessmentRecordDetailVO{}, fmt.Errorf("评估记录未完成，仍有 %d 道题目缺少评分", scoreResult.Result.MissingItemCount)
	}
	recordID, err := svc.repo.CreateAssessmentRecord(context.Background(), repository.AssessmentRecordEntity{
		InstID:         instID,
		StudentID:      input.StudentID,
		StudentName:    input.StudentName,
		AssessmentCode: autismDevScaleCode,
		AssessmentName: autismDevAssessmentName,
		ScaleVersion:   scoreResult.ScaleVersion,
		BirthDate:      input.ScoreInput.BirthDate,
		AssessmentDate: input.ScoreInput.AssessmentDate,
		AgeYears:       scoreResult.Result.Age.Years,
		AgeMonths:      scoreResult.Result.Age.Months,
		AgeDays:        scoreResult.Result.Age.Days,
		NormAgeMonths:  int(scoreResult.Result.Age.TotalMonthsRounded),
		ExaminerID:     examinerID,
		ExaminerName:   examinerName,
		Input:          input.InputSnapshot,
		Result:         scoreResult,
		DataStatus:     scoreResult.DataStatus,
		Remark:         input.Remark,
	})
	if err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	return svc.repo.GetAssessmentRecord(context.Background(), instID, recordID)
}

func (svc *Service) UpdateAutismDevAssessmentRecord(userID, recordID int64, input AutismDevAssessmentRecordSaveInput) (model.AssessmentRecordDetailVO, error) {
	if svc.repo == nil {
		return model.AssessmentRecordDetailVO{}, errors.New("assessment repository is not configured")
	}
	instID, examinerID, examinerName, err := svc.pep3AssessmentActor(userID, input.ExaminerName)
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
	if strings.TrimSpace(record.AssessmentCode) != autismDevScaleCode {
		return model.AssessmentRecordDetailVO{}, errors.New("assessment record is not AutismDev")
	}
	if err := svc.validatePEP3AssessmentStudent(instID, input.StudentID, input.StudentName); err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	scoreResult, err := svc.ScoreAutismDev(input.ScoreInput)
	if err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	if !scoreResult.Result.Complete {
		return model.AssessmentRecordDetailVO{}, fmt.Errorf("评估记录未完成，仍有 %d 道题目缺少评分", scoreResult.Result.MissingItemCount)
	}
	if err := svc.repo.UpdateAssessmentRecordBaseInputAndResult(context.Background(), repository.AssessmentRecordEntity{
		ID:             recordID,
		InstID:         instID,
		StudentID:      input.StudentID,
		StudentName:    input.StudentName,
		AssessmentCode: autismDevScaleCode,
		AssessmentName: autismDevAssessmentName,
		ScaleVersion:   scoreResult.ScaleVersion,
		BirthDate:      input.ScoreInput.BirthDate,
		AssessmentDate: input.ScoreInput.AssessmentDate,
		AgeYears:       scoreResult.Result.Age.Years,
		AgeMonths:      scoreResult.Result.Age.Months,
		AgeDays:        scoreResult.Result.Age.Days,
		NormAgeMonths:  int(scoreResult.Result.Age.TotalMonthsRounded),
		ExaminerID:     examinerID,
		ExaminerName:   examinerName,
		Input:          input.InputSnapshot,
		Result:         scoreResult,
		DataStatus:     scoreResult.DataStatus,
		Remark:         input.Remark,
	}); err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	if err := svc.syncAutismDevSubmittedDraftAfterRecordUpdate(userID, instID, recordID, input); err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	return svc.repo.GetAssessmentRecord(context.Background(), instID, recordID)
}

func (svc *Service) syncAutismDevSubmittedDraftAfterRecordUpdate(userID, instID, recordID int64, input AutismDevAssessmentRecordSaveInput) error {
	draft, err := svc.repo.GetAssessmentDraftBySubmittedRecordID(context.Background(), instID, recordID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	examinerID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution user context")
		}
		return err
	}
	progress, err := buildAutismDevAssessmentDraftProgress(&input.ScoreInput.BirthDate, &input.ScoreInput.AssessmentDate, input.ScoreInput.ItemScores)
	if err != nil {
		return err
	}
	return svc.repo.UpdateAssessmentDraftInputAndProgressIncludingSubmitted(
		context.Background(),
		instID,
		draft.ID,
		input.InputSnapshot,
		progress,
		progress.AnsweredItemCount,
		0,
		"submitted",
		examinerID,
		autismDevItemScoresForDB(input.ScoreInput.ItemScores),
		nil,
		nil,
	)
}

func (svc *Service) GetAutismDevAssessmentRecord(userID, recordID int64) (model.AssessmentRecordDetailVO, error) {
	if svc.repo == nil {
		return model.AssessmentRecordDetailVO{}, errors.New("assessment repository is not configured")
	}
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	record, err := svc.repo.GetAssessmentRecord(context.Background(), instID, recordID)
	if err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	if strings.TrimSpace(record.AssessmentCode) != autismDevScaleCode {
		return model.AssessmentRecordDetailVO{}, errors.New("assessment record is not AutismDev")
	}
	return record, nil
}

func (svc *Service) PageAutismDevAssessmentRecords(userID int64, query model.AssessmentRecordPageQueryDTO) (model.PageResult[model.AssessmentRecordSummaryVO], error) {
	if svc.repo == nil {
		return model.PageResult[model.AssessmentRecordSummaryVO]{}, errors.New("assessment repository is not configured")
	}
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return model.PageResult[model.AssessmentRecordSummaryVO]{}, err
	}
	query.QueryModel.AssessmentCode = autismDevScaleCode
	return svc.repo.PageAssessmentRecords(context.Background(), instID, query.QueryModel, query.PageRequestModel.PageIndex, query.PageRequestModel.PageSize)
}

func (svc *Service) SummarizeAutismDevAssessmentRecordCategories(userID int64, query model.AssessmentRecordQueryModel) (model.AssessmentRecordCategoryStatsVO, error) {
	if svc.repo == nil {
		return model.AssessmentRecordCategoryStatsVO{}, errors.New("assessment repository is not configured")
	}
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return model.AssessmentRecordCategoryStatsVO{}, err
	}
	query.AssessmentCode = autismDevScaleCode
	query.ScaleCategory = ""
	return svc.repo.SummarizeAssessmentRecordCategories(context.Background(), instID, query)
}

func (svc *Service) DeleteAutismDevAssessmentRecord(userID, recordID int64) (bool, error) {
	if svc.repo == nil {
		return false, errors.New("assessment repository is not configured")
	}
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return false, err
	}
	record, err := svc.repo.GetAssessmentRecord(context.Background(), instID, recordID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, errors.New("assessment record not found")
		}
		return false, err
	}
	if strings.TrimSpace(record.AssessmentCode) != autismDevScaleCode {
		return false, errors.New("assessment record is not AutismDev")
	}
	return svc.repo.DeleteAssessmentRecord(context.Background(), instID, recordID)
}

func (svc *Service) UpdateAutismDevAssessmentRecordConfig(userID, recordID int64, input AutismDevAssessmentRecordConfigInput) (model.AssessmentRecordDetailVO, error) {
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
	if strings.TrimSpace(record.AssessmentCode) != autismDevScaleCode {
		return model.AssessmentRecordDetailVO{}, errors.New("assessment record is not AutismDev")
	}

	assessmentDate := dateOnlyERXinRecordConfig(input.AssessmentDate)
	if err := svc.repo.UpdateAssessmentRecordConfig(context.Background(), instID, recordID, examinerName, assessmentDate); err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	if err := svc.syncAutismDevSubmittedDraftRecordConfig(instID, operatorID, recordID, examinerName, assessmentDate); err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	if err := svc.syncPEP3SavedPlanHeadersWithRecordConfig(context.Background(), instID, recordID, pep3IEPPlanHeaderValuesForRecordConfig(record, examinerName, assessmentDate), userID); err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	return svc.repo.GetAssessmentRecord(context.Background(), instID, recordID)
}

func (svc *Service) syncAutismDevSubmittedDraftRecordConfig(instID, operatorID, recordID int64, examinerName string, assessmentDate time.Time) error {
	draft, err := svc.repo.GetAssessmentDraftBySubmittedRecordID(context.Background(), instID, recordID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	return svc.repo.UpdateAssessmentDraftConfigIncludingSubmitted(context.Background(), instID, draft.ID, examinerName, assessmentDate, operatorID)
}
