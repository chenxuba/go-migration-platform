package service

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) GetTuitionAccountSubAccountDateInfo(userID int64, dto model.TuitionAccountSubAccountDateInfoQueryDTO) (model.TuitionAccountSubAccountDateInfoResult, error) {
	svc.SyncScheduledSuspendResumeTuitionAccountsOnce()
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.TuitionAccountSubAccountDateInfoResult{}, errors.New("no institution context")
		}
		return model.TuitionAccountSubAccountDateInfoResult{}, err
	}
	return svc.repo.GetTuitionAccountSubAccountDateInfo(context.Background(), instID, dto)
}

func (svc *Service) ListSubTuitionAccountPriorityConfigs(userID int64) (model.SubTuitionAccountPriorityConfigResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.SubTuitionAccountPriorityConfigResult{}, errors.New("no institution context")
		}
		return model.SubTuitionAccountPriorityConfigResult{}, err
	}
	return svc.repo.ListSubTuitionAccountPriorityConfigs(context.Background(), instID)
}

func (svc *Service) GetRevertCloseTuitionAccountPreview(userID int64, dto model.RevertCloseTuitionAccountPreviewQueryDTO) (model.RevertCloseTuitionAccountPreview, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.RevertCloseTuitionAccountPreview{}, errors.New("no institution context")
		}
		return model.RevertCloseTuitionAccountPreview{}, err
	}
	return svc.repo.GetRevertCloseTuitionAccountPreview(context.Background(), instID, dto)
}

func (svc *Service) RevertCloseTuitionAccount(userID int64, dto model.RevertCloseTuitionAccountDTO) (model.RevertCloseTuitionAccountResult, error) {
	instID, operatorID, err := svc.resolveTeachingClassOperator(userID)
	if err != nil {
		return model.RevertCloseTuitionAccountResult{}, err
	}
	id, err := svc.repo.RevertCloseTuitionAccount(context.Background(), instID, operatorID, dto)
	if err != nil {
		return model.RevertCloseTuitionAccountResult{}, err
	}
	return model.RevertCloseTuitionAccountResult{ID: strconv.FormatInt(id, 10)}, nil
}

func (svc *Service) ListCloseTuitionAccountOrders(userID int64, dto model.CloseTuitionAccountOrderRecordQueryDTO) (model.CloseTuitionAccountOrderRecordResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.CloseTuitionAccountOrderRecordResult{}, errors.New("no institution context")
		}
		return model.CloseTuitionAccountOrderRecordResult{}, err
	}
	return svc.repo.ListCloseTuitionAccountOrders(context.Background(), instID, dto)
}
