package service

import (
	"context"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) GetStudentTeachingRecordPagedList(userID int64, dto model.StudentTeachingRecordPagedQueryDTO) (model.StudentTeachingRecordPagedResult, error) {
	instID, err := svc.rollCallInstID(userID)
	if err != nil {
		return model.StudentTeachingRecordPagedResult{}, err
	}
	return svc.repo.GetStudentTeachingRecordPagedList(context.Background(), instID, dto)
}

func (svc *Service) GetScheduleTeachingRecordPagedList(userID int64, dto model.ScheduleTeachingRecordPagedQueryDTO) (model.ScheduleTeachingRecordPagedResult, error) {
	instID, err := svc.rollCallInstID(userID)
	if err != nil {
		return model.ScheduleTeachingRecordPagedResult{}, err
	}
	return svc.repo.GetScheduleTeachingRecordPagedList(context.Background(), instID, dto)
}

func (svc *Service) GetTeachingRecordDetail(userID int64, query model.TeachingRecordDetailQueryDTO) (model.TeachingRecordDetailResult, error) {
	instID, err := svc.rollCallInstID(userID)
	if err != nil {
		return model.TeachingRecordDetailResult{}, err
	}
	return svc.repo.GetTeachingRecordDetail(context.Background(), instID, query)
}
