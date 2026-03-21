package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"go-migration-platform/services/education/internal/model"
	"golang.org/x/crypto/bcrypt"
)

func (svc *Service) PageInstUsers(userID int64, query model.InstUserQueryDTO) (model.PageResult[model.InstUserQueryVO], error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PageResult[model.InstUserQueryVO]{}, errors.New("no institution context")
		}
		return model.PageResult[model.InstUserQueryVO]{}, err
	}
	return svc.repo.PageInstUsers(context.Background(), instID, query.QueryModel, query.PageRequestModel.PageIndex, query.PageRequestModel.PageSize)
}

func (svc *Service) GetInstUserDetail(userID, instUserID int64) (model.InstUserDetailVO, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.InstUserDetailVO{}, errors.New("no institution context")
		}
		return model.InstUserDetailVO{}, err
	}
	return svc.repo.GetInstUserDetail(context.Background(), instUserID, instID)
}

func (svc *Service) SaveInstUser(userID int64, dto model.InstUserSaveDTO) (int64, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("no institution context")
		}
		return 0, err
	}
	if strings.TrimSpace(dto.Mobile) == "" || strings.TrimSpace(dto.NickName) == "" {
		return 0, errors.New("nickName and mobile are required")
	}
	used, err := svc.repo.CheckPhoneUsed(context.Background(), instID, dto.Mobile, nil)
	if err != nil {
		return 0, err
	}
	if used {
		return 0, errors.New("当前机构下此手机号账号已存在")
	}
	if dto.Password == "" {
		dto.Password = "123456"
	}
	if dto.UserType == nil {
		defaultType := 1
		dto.UserType = &defaultType
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	return svc.repo.SaveInstUser(context.Background(), instID, dto, string(hash))
}

func (svc *Service) UpdateInstUser(userID int64, dto model.InstUserModifyDTO) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	if dto.ID <= 0 {
		return errors.New("id is required")
	}
	return svc.repo.UpdateInstUser(context.Background(), instID, dto)
}

func (svc *Service) BatchDisabledInstUsers(userID int64, dto model.BatchCommonDTO) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	if len(dto.UserIDs) == 0 {
		return errors.New("userIds are required")
	}
	disabled := true
	if dto.IsWork != nil {
		disabled = *dto.IsWork
	}
	return svc.repo.BatchSetInstUserDisabled(context.Background(), instID, dto.UserIDs, disabled)
}

func (svc *Service) BatchModifyInstUserDept(userID int64, dto model.BatchCommonDTO) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	if len(dto.UserIDs) == 0 || len(dto.DeptIDs) == 0 {
		return errors.New("userIds and deptIds are required")
	}
	return svc.repo.BatchModifyInstUserDept(context.Background(), instID, dto.UserIDs, dto.DeptIDs)
}

func (svc *Service) BatchModifyInstUserRole(userID int64, dto model.BatchCommonDTO) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	if len(dto.UserIDs) == 0 || len(dto.RoleIDs) == 0 {
		return errors.New("userIds and roleIds are required")
	}
	return svc.repo.BatchModifyInstUserRole(context.Background(), instID, dto.UserIDs, dto.RoleIDs)
}

func (svc *Service) CheckInstUserPhoneUsed(userID int64, vo model.ChangePhoneVO) (bool, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, errors.New("no institution context")
		}
		return false, err
	}
	if strings.TrimSpace(vo.Mobile) == "" {
		return false, errors.New("mobile is required")
	}
	used, err := svc.repo.CheckPhoneUsed(context.Background(), instID, vo.Mobile, nil)
	if err != nil {
		return false, err
	}
	if used {
		return false, errors.New("手机号已占用")
	}
	return true, nil
}

func (svc *Service) ChangeInstUserPhone(userID int64, vo model.ChangePhoneVO) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	if vo.UserID <= 0 || strings.TrimSpace(vo.Mobile) == "" {
		return errors.New("userId and mobile are required")
	}
	used, err := svc.repo.CheckPhoneUsed(context.Background(), instID, vo.Mobile, &vo.UserID)
	if err != nil {
		return err
	}
	if used {
		return errors.New("手机号已占用")
	}
	return svc.repo.ChangeInstUserPhone(context.Background(), vo.UserID, instID, vo.Mobile)
}
