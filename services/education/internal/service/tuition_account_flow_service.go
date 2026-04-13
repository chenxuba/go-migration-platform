package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

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

func (svc *Service) ExportTuitionAccountFlowRecordList(userID int64, query model.TuitionAccountFlowRecordListQueryDTO) ([]byte, string, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", errors.New("no institution context")
		}
		return nil, "", err
	}

	exportQuery := query
	exportQuery.PageRequestModel.PageIndex = 1
	exportQuery.PageRequestModel.PageSize = tuitionAccountFlowExportMaxRows

	result, err := svc.repo.GetTuitionAccountFlowRecordList(context.Background(), instID, exportQuery)
	if err != nil {
		return nil, "", err
	}
	if result.Total == 0 || len(result.List) == 0 {
		return nil, "", errors.New("没有符合条件的学费变动记录可以导出")
	}
	if result.Total > tuitionAccountFlowExportMaxRows {
		return nil, "", errors.New("当前列表最多支持导出10000条数据，请缩小筛选范围后重试")
	}

	content, err := buildTuitionAccountFlowExportWorkbook(result.List)
	if err != nil {
		return nil, "", err
	}
	fileName := fmt.Sprintf("学费变动记录-%s.xlsx", time.Now().Format("20060102150405"))
	return content, fileName, nil
}
