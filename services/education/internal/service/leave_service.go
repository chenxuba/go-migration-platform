package service

import (
	"context"
	"database/sql"
	"errors"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) PreviewLeaveSchedules(userID int64, dto model.LeavePreviewDTO) ([]model.LeaveScheduleSnapshotVO, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no institution context")
		}
		return nil, err
	}
	return svc.repo.PreviewLeaveSchedules(context.Background(), instID, dto)
}

func (svc *Service) CreateLeaveRequest(userID int64, dto model.LeaveCreateDTO) (model.LeaveCreateResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.LeaveCreateResult{}, errors.New("no institution context")
		}
		return model.LeaveCreateResult{}, err
	}
	instUserID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.LeaveCreateResult{}, errors.New("no institution user context")
		}
		return model.LeaveCreateResult{}, err
	}
	return svc.repo.CreateLeaveRequest(context.Background(), instID, instUserID, dto)
}

func (svc *Service) PageLeaveRequests(userID int64, query model.LeavePagedQueryDTO) (model.LeavePagedResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.LeavePagedResult{}, errors.New("no institution context")
		}
		return model.LeavePagedResult{}, err
	}
	return svc.repo.PageLeaveRequests(context.Background(), instID, query)
}

func (svc *Service) GetLeaveDetail(userID, leaveID int64) (model.LeaveDetailVO, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.LeaveDetailVO{}, errors.New("no institution context")
		}
		return model.LeaveDetailVO{}, err
	}
	if leaveID <= 0 {
		return model.LeaveDetailVO{}, errors.New("id不能为空")
	}
	return svc.repo.GetLeaveDetail(context.Background(), instID, leaveID)
}

func (svc *Service) PageLeaveDetailSchedules(userID int64, query model.LeaveDetailScheduleQueryDTO) (model.LeaveDetailSchedulePagedResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.LeaveDetailSchedulePagedResult{}, errors.New("no institution context")
		}
		return model.LeaveDetailSchedulePagedResult{}, err
	}
	return svc.repo.PageLeaveDetailSchedules(context.Background(), instID, query)
}
