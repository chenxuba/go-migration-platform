package service

import (
	"context"
	"database/sql"
	"errors"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) GetRechargeAccountItemPage(userID int64, query model.RechargeAccountItemPageQueryDTO) (model.RechargeAccountItemPageResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.RechargeAccountItemPageResult{}, errors.New("no institution context")
		}
		return model.RechargeAccountItemPageResult{}, err
	}
	return svc.repo.PageRechargeAccountItems(context.Background(), instID, query)
}

func (svc *Service) GetRechargeAccountStatistics(userID int64) (model.RechargeAccountStatistics, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.RechargeAccountStatistics{}, errors.New("no institution context")
		}
		return model.RechargeAccountStatistics{}, err
	}
	return svc.repo.GetRechargeAccountStatistics(context.Background(), instID)
}
