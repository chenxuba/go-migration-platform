package service

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) GetLedgerList(userID int64, query model.LedgerListQueryDTO) (model.LedgerListResultVO, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.LedgerListResultVO{}, errors.New("no institution context")
		}
		return model.LedgerListResultVO{}, err
	}
	return svc.repo.PageLedgers(context.Background(), instID, query)
}

func (svc *Service) GetLedgerStatistics(userID int64, query model.LedgerListQueryDTO) (model.LedgerStatisticsVO, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.LedgerStatisticsVO{}, errors.New("no institution context")
		}
		return model.LedgerStatisticsVO{}, err
	}
	return svc.repo.GetLedgerStatistics(context.Background(), instID, query)
}

func (svc *Service) ConfirmLedger(userID int64, dto model.LedgerOperateDTO) error {
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
	ledgerID, err := strconv.ParseInt(strings.TrimSpace(dto.ID), 10, 64)
	if err != nil || ledgerID <= 0 {
		return errors.New("账单ID不能为空")
	}
	operatorName := svc.repo.GetStaffNameByID(context.Background(), &instUserID)
	return svc.repo.ConfirmLedger(context.Background(), instID, ledgerID, instUserID, operatorName)
}

func (svc *Service) CancelConfirmLedger(userID int64, dto model.LedgerOperateDTO) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	ledgerID, err := strconv.ParseInt(strings.TrimSpace(dto.ID), 10, 64)
	if err != nil || ledgerID <= 0 {
		return errors.New("账单ID不能为空")
	}
	return svc.repo.CancelConfirmLedger(context.Background(), instID, ledgerID)
}
