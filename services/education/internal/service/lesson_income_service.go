package service

import (
	"context"
	"database/sql"
	"errors"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) GetLessonIncomePagedList(userID int64, query model.LessonIncomeQueryDTO) (model.LessonIncomePagedResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.LessonIncomePagedResult{}, errors.New("no institution context")
		}
		return model.LessonIncomePagedResult{}, err
	}
	return svc.repo.GetLessonIncomePagedList(context.Background(), instID, query)
}

func (svc *Service) GetLessonIncomeStatistics(userID int64, query model.LessonIncomeQueryDTO) (model.LessonIncomeStatistics, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.LessonIncomeStatistics{}, errors.New("no institution context")
		}
		return model.LessonIncomeStatistics{}, err
	}
	return svc.repo.GetLessonIncomeStatistics(context.Background(), instID, query)
}
