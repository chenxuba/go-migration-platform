package service

import (
	"context"
	"database/sql"
	"errors"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) AddSuspendResumeTuitionAccountOrder(userID int64, dto model.SuspendResumeTuitionAccountOrderDTO) (model.SuspendResumeTuitionAccountOrderResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.SuspendResumeTuitionAccountOrderResult{}, errors.New("no institution context")
		}
		return model.SuspendResumeTuitionAccountOrderResult{}, err
	}
	operatorID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.SuspendResumeTuitionAccountOrderResult{}, errors.New("no institution user context")
		}
		return model.SuspendResumeTuitionAccountOrderResult{}, err
	}
	return svc.repo.AddSuspendResumeTuitionAccountOrder(context.Background(), instID, operatorID, dto)
}
