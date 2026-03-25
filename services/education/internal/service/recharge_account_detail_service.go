package service

import (
	"context"
	"database/sql"
	"errors"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) GetRechargeAccountDetailPage(userID int64, query model.RechargeAccountDetailQueryDTO) (model.RechargeAccountDetailPageResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.RechargeAccountDetailPageResult{}, errors.New("no institution context")
		}
		return model.RechargeAccountDetailPageResult{}, err
	}
	return svc.repo.PageRechargeAccountDetails(context.Background(), instID, query)
}

func (svc *Service) GetRechargeAccountExpendIncome(userID int64, query model.RechargeAccountDetailQuery) (model.RechargeAccountExpendIncome, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.RechargeAccountExpendIncome{}, errors.New("no institution context")
		}
		return model.RechargeAccountExpendIncome{}, err
	}
	return svc.repo.GetRechargeAccountExpendIncome(context.Background(), instID, query)
}
