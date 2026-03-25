package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

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

func (svc *Service) UpdateRechargeAccount(userID int64, dto model.UpdateRechargeAccountDTO) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	instUserID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution user context")
		}
		return err
	}

	name := strings.TrimSpace(dto.RechargeAccountName)
	if strings.TrimSpace(dto.RechargeAccountID) == "" {
		return errors.New("rechargeAccountId不能为空")
	}
	if name == "" {
		return errors.New("rechargeAccountName不能为空")
	}
	if len([]rune(name)) > 30 {
		return errors.New("rechargeAccountName长度不能超过30")
	}

	dto.RechargeAccountName = name
	return svc.repo.UpdateRechargeAccount(context.Background(), instID, instUserID, dto)
}
