package service

import (
	"context"
	"database/sql"
	"errors"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) GetStudentOverviewStatistics(userID int64) (model.StudentOverviewStatistics, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.StudentOverviewStatistics{}, errors.New("no institution context")
		}
		return model.StudentOverviewStatistics{}, err
	}
	return svc.repo.GetStudentOverviewStatistics(context.Background(), instID)
}
