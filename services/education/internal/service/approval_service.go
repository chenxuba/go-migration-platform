package service

import (
	"context"
	"database/sql"
	"errors"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) ApprovalConfigPaged(userID int64, query model.ApprovalConfigPageQueryDTO) (model.ApprovalConfigPageResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.ApprovalConfigPageResult{}, errors.New("no institution context")
		}
		return model.ApprovalConfigPageResult{}, err
	}
	return svc.repo.PageApprovalConfigs(context.Background(), instID, query)
}

func (svc *Service) ListApprovalTemplates(userID int64) ([]model.ApprovalTemplateVO, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no institution context")
		}
		return nil, err
	}
	return svc.repo.ListApprovalTemplates(context.Background(), instID)
}

func (svc *Service) SaveApprovalTemplates(userID int64, dto model.ApprovalTemplateSaveRequest) error {
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
	return svc.repo.SaveApprovalTemplates(context.Background(), instID, instUserID, dto)
}

func (svc *Service) StaffSummaries(userID int64, query model.StaffSummaryQueryDTO) (model.StaffSummaryPageVO, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.StaffSummaryPageVO{}, errors.New("no institution context")
		}
		return model.StaffSummaryPageVO{}, err
	}
	return svc.repo.PageStaffSummaries(context.Background(), instID, query)
}

func (svc *Service) SaveApprovalConfig(userID int64, dto model.ApprovalConfigSaveDTO) error {
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
	if dto.ID <= 0 {
		return errors.New("id不能为空")
	}
	return svc.repo.SaveApprovalConfig(context.Background(), instID, instUserID, dto)
}

func (svc *Service) ApproveApprovalRecord(userID int64, dto model.ApprovalOperateDTO) error {
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
	if dto.ID <= 0 {
		return errors.New("id不能为空")
	}
	return svc.repo.ApproveApprovalRecord(context.Background(), instID, instUserID, dto)
}

func (svc *Service) CancelApprovalRecord(userID int64, dto model.ApprovalOperateDTO) error {
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
	if dto.ID <= 0 {
		return errors.New("id不能为空")
	}
	return svc.repo.CancelApprovalRecord(context.Background(), instID, instUserID, dto)
}
