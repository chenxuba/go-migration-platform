package service

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) GetOrderList(userID int64, query model.OrderManageQueryDTO) (model.OrderManageResultVO, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.OrderManageResultVO{}, errors.New("no institution context")
		}
		return model.OrderManageResultVO{}, err
	}
	return svc.repo.PageOrders(context.Background(), instID, query)
}

func (svc *Service) GetOrderDetail(userID, orderID int64) (model.OrderManageQueryVO, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.OrderManageQueryVO{}, errors.New("no institution context")
		}
		return model.OrderManageQueryVO{}, err
	}
	return svc.repo.GetOrderDetail(context.Background(), instID, orderID)
}

func (svc *Service) SetBadDebt(userID int64, dto model.BadDebtDTO) error {
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
	orderID, err := strconv.ParseInt(strings.TrimSpace(dto.OrderID), 10, 64)
	if err != nil || orderID <= 0 {
		return errors.New("订单ID不能为空")
	}
	return svc.repo.SetBadDebt(context.Background(), instID, orderID, instUserID, dto.Remark)
}

func (svc *Service) CancelBadDebt(userID int64, orderIDRaw string) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	orderID, err := strconv.ParseInt(strings.TrimSpace(orderIDRaw), 10, 64)
	if err != nil || orderID <= 0 {
		return errors.New("订单ID不能为空")
	}
	return svc.repo.CancelBadDebt(context.Background(), instID, orderID)
}
