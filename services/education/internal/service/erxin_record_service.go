package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"go-migration-platform/pkg/erxinscore"
	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/repository"
)

type ERXinAssessmentRecordSaveInput struct {
	StudentID     int64
	StudentName   string
	ExaminerName  string
	Remark        string
	ScoreInput    erxinscore.AssessmentInput
	InputSnapshot any
}

func (svc *Service) CreateERXinAssessmentRecord(userID int64, input ERXinAssessmentRecordSaveInput) (model.AssessmentRecordDetailVO, error) {
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

	scoreResult, err := svc.ScoreERXin(input.ScoreInput)
	if err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	recordID, err := svc.repo.CreateAssessmentRecord(context.Background(), repository.AssessmentRecordEntity{
		InstID:         instID,
		StudentID:      input.StudentID,
		StudentName:    input.StudentName,
		AssessmentCode: erxinScaleCode,
		AssessmentName: erxinAssessmentName,
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

func (svc *Service) UpdateERXinAssessmentRecord(userID, recordID int64, input ERXinAssessmentRecordSaveInput) (model.AssessmentRecordDetailVO, error) {
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
	if strings.TrimSpace(record.AssessmentCode) != erxinScaleCode {
		return model.AssessmentRecordDetailVO{}, errors.New("assessment record is not ERXin")
	}
	if err := svc.validatePEP3AssessmentStudent(instID, input.StudentID, input.StudentName); err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	scoreResult, err := svc.ScoreERXin(input.ScoreInput)
	if err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	if err := svc.repo.UpdateAssessmentRecordBaseInputAndResult(context.Background(), repository.AssessmentRecordEntity{
		ID:             recordID,
		InstID:         instID,
		StudentID:      input.StudentID,
		StudentName:    input.StudentName,
		AssessmentCode: erxinScaleCode,
		AssessmentName: erxinAssessmentName,
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
	if err := svc.syncERXinSubmittedDraftAfterRecordUpdate(userID, instID, recordID, input); err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	return svc.repo.GetAssessmentRecord(context.Background(), instID, recordID)
}

func (svc *Service) syncERXinSubmittedDraftAfterRecordUpdate(userID, instID, recordID int64, input ERXinAssessmentRecordSaveInput) error {
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
	progress, err := buildERXinAssessmentDraftProgress(&input.ScoreInput.BirthDate, &input.ScoreInput.AssessmentDate, input.ScoreInput.ItemPasses)
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
		erxinItemScoresFromPasses(input.ScoreInput.ItemPasses),
		nil,
		nil,
	)
}

func (svc *Service) GetERXinAssessmentRecord(userID, recordID int64) (model.AssessmentRecordDetailVO, error) {
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
	if strings.TrimSpace(record.AssessmentCode) != erxinScaleCode {
		return model.AssessmentRecordDetailVO{}, errors.New("assessment record is not ERXin")
	}
	return record, nil
}

func (svc *Service) PageERXinAssessmentRecords(userID int64, query model.AssessmentRecordPageQueryDTO) (model.PageResult[model.AssessmentRecordSummaryVO], error) {
	if svc.repo == nil {
		return model.PageResult[model.AssessmentRecordSummaryVO]{}, errors.New("assessment repository is not configured")
	}
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return model.PageResult[model.AssessmentRecordSummaryVO]{}, err
	}
	query.QueryModel.AssessmentCode = erxinScaleCode
	return svc.repo.PageAssessmentRecords(context.Background(), instID, query.QueryModel, query.PageRequestModel.PageIndex, query.PageRequestModel.PageSize)
}

func (svc *Service) SummarizeERXinAssessmentRecordCategories(userID int64, query model.AssessmentRecordQueryModel) (model.AssessmentRecordCategoryStatsVO, error) {
	if svc.repo == nil {
		return model.AssessmentRecordCategoryStatsVO{}, errors.New("assessment repository is not configured")
	}
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return model.AssessmentRecordCategoryStatsVO{}, err
	}
	query.AssessmentCode = erxinScaleCode
	query.ScaleCategory = ""
	return svc.repo.SummarizeAssessmentRecordCategories(context.Background(), instID, query)
}

func (svc *Service) DeleteERXinAssessmentRecord(userID, recordID int64) (bool, error) {
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
	if strings.TrimSpace(record.AssessmentCode) != erxinScaleCode {
		return false, errors.New("assessment record is not ERXin")
	}
	return svc.repo.DeleteAssessmentRecord(context.Background(), instID, recordID)
}
