package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) CreateOneToOneSchedules(userID int64, dto model.CreateOneToOneSchedulesDTO) (model.CreateOneToOneSchedulesResult, error) {
	instID, operatorID, err := svc.resolveTeachingScheduleOperator(userID)
	if err != nil {
		return model.CreateOneToOneSchedulesResult{}, err
	}
	if strings.TrimSpace(dto.OneToOneID) == "" {
		return model.CreateOneToOneSchedulesResult{}, errors.New("请选择1对1")
	}
	return svc.repo.CreateOneToOneSchedules(context.Background(), instID, operatorID, dto)
}

func (svc *Service) ValidateOneToOneSchedules(userID int64, dto model.CreateOneToOneSchedulesDTO) (model.TeachingScheduleValidationResult, error) {
	instID, _, err := svc.resolveTeachingScheduleOperator(userID)
	if err != nil {
		return model.TeachingScheduleValidationResult{}, err
	}
	if strings.TrimSpace(dto.OneToOneID) == "" {
		return model.TeachingScheduleValidationResult{}, errors.New("请选择1对1")
	}
	return svc.repo.ValidateOneToOneSchedules(context.Background(), instID, dto)
}

func (svc *Service) ListTeachingSchedules(userID int64, query model.TeachingScheduleListQueryDTO) ([]model.TeachingScheduleVO, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no institution context")
		}
		return nil, err
	}
	return svc.repo.ListTeachingSchedules(context.Background(), instID, query)
}

func (svc *Service) BatchUpdateTeachingSchedules(userID int64, dto model.TeachingScheduleBatchUpdateDTO) error {
	instID, operatorID, err := svc.resolveTeachingScheduleOperator(userID)
	if err != nil {
		return err
	}
	return svc.repo.BatchUpdateTeachingSchedules(context.Background(), instID, operatorID, dto)
}

func (svc *Service) resolveTeachingScheduleOperator(userID int64) (int64, int64, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, errors.New("no institution context")
		}
		return 0, 0, err
	}
	operatorID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, errors.New("no institution user context")
		}
		return 0, 0, err
	}
	return instID, operatorID, nil
}
