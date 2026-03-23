package service

import (
	"context"
	"database/sql"
	"errors"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) PageOrderTags(userID int64, query model.OrderTagPagedQueryDTO) (model.OrderTagPagedResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.OrderTagPagedResult{}, errors.New("no institution context")
		}
		return model.OrderTagPagedResult{}, err
	}
	return svc.repo.PageOrderTags(context.Background(), instID, query)
}

func (svc *Service) CreateOrderTag(userID int64, dto model.CreateOrderTagDTO) (int64, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("no institution context")
		}
		return 0, err
	}
	instUserID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("no institution user context")
		}
		return 0, err
	}
	return svc.repo.CreateOrderTag(context.Background(), instID, instUserID, dto)
}

func (svc *Service) UpdateOrderTag(userID int64, dto model.UpdateOrderTagDTO) error {
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
	return svc.repo.UpdateOrderTag(context.Background(), instID, instUserID, dto)
}
