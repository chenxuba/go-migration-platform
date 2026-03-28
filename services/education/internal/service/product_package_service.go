package service

import (
	"context"
	"database/sql"
	"errors"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) PageProductPackages(userID int64, query model.ProductPackageQueryDTO) (model.ProductPackagePagedResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.ProductPackagePagedResult{}, errors.New("no institution context")
		}
		return model.ProductPackagePagedResult{}, err
	}
	return svc.repo.PageProductPackages(context.Background(), instID, query)
}

func (svc *Service) GetProductPackageStatistics(userID int64, filters model.ProductPackageQueryFilter) (model.ProductPackageStatistics, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.ProductPackageStatistics{}, errors.New("no institution context")
		}
		return model.ProductPackageStatistics{}, err
	}
	return svc.repo.GetProductPackageStatistics(context.Background(), instID, filters)
}

func (svc *Service) CreateProductPackage(userID int64, dto model.ProductPackageMutation) (int64, error) {
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
	if dto.Name == "" {
		return 0, errors.New("套餐名称不能为空")
	}
	if dto.Title == "" {
		dto.Title = dto.Name
	}
	if len(dto.Items) == 0 {
		return 0, errors.New("套餐内商品不能为空")
	}
	return svc.repo.CreateProductPackage(context.Background(), instID, instUserID, dto)
}
