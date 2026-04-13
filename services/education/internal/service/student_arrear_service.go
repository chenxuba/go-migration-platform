package service

import (
	"context"
	"database/sql"
	"errors"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) GetStudentRegistrationArrearPagedList(userID int64, query model.StudentRegistrationArrearPagedQueryDTO) (model.StudentRegistrationArrearPagedResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.StudentRegistrationArrearPagedResult{}, errors.New("no institution context")
		}
		return model.StudentRegistrationArrearPagedResult{}, err
	}
	return svc.repo.GetStudentRegistrationArrearPagedList(context.Background(), instID, query)
}

func (svc *Service) GetStudentRegistrationArrearStatistics(userID int64, query model.StudentRegistrationArrearQueryModel) (model.StudentRegistrationArrearStatistics, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.StudentRegistrationArrearStatistics{}, errors.New("no institution context")
		}
		return model.StudentRegistrationArrearStatistics{}, err
	}
	return svc.repo.GetStudentRegistrationArrearStatistics(context.Background(), instID, query)
}

func (svc *Service) GetStudentLessonArrearPagedList(userID int64, query model.StudentLessonArrearPagedQueryDTO) (model.StudentLessonArrearPagedResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.StudentLessonArrearPagedResult{}, errors.New("no institution context")
		}
		return model.StudentLessonArrearPagedResult{}, err
	}
	return svc.repo.GetStudentLessonArrearPagedList(context.Background(), instID, query)
}

func (svc *Service) GetStudentLessonArrearStatistics(userID int64, query model.StudentLessonArrearQueryModel) (model.StudentLessonArrearStatistics, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.StudentLessonArrearStatistics{}, errors.New("no institution context")
		}
		return model.StudentLessonArrearStatistics{}, err
	}
	return svc.repo.GetStudentLessonArrearStatistics(context.Background(), instID, query)
}
