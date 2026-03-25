package service

import (
	"context"
	"database/sql"
	"errors"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) GetTuitionAccountFlowRecordList(userID int64, query model.TuitionAccountFlowRecordListQueryDTO) (model.TuitionAccountFlowRecordListResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.TuitionAccountFlowRecordListResult{}, errors.New("no institution context")
		}
		return model.TuitionAccountFlowRecordListResult{}, err
	}
	return svc.repo.GetTuitionAccountFlowRecordList(context.Background(), instID, query)
}

func (svc *Service) GetSubTuitionAccountFlowRecordList(userID int64, query model.SubTuitionAccountFlowRecordListQueryDTO) (model.SubTuitionAccountFlowRecordListResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.SubTuitionAccountFlowRecordListResult{}, errors.New("no institution context")
		}
		return model.SubTuitionAccountFlowRecordListResult{}, err
	}
	return svc.repo.GetSubTuitionAccountFlowRecordList(context.Background(), instID, query)
}
