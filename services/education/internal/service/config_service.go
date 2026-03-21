package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"go-migration-platform/pkg/qiniux"
	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) GetQiniuUploadToken() (qiniux.TokenVO, error) {
	if svc.qiniuClient == nil {
		return qiniux.TokenVO{}, errors.New("qiniu not configured")
	}
	return svc.qiniuClient.ImageUploadToken()
}

func (svc *Service) GetQiniuVideoUploadToken() (qiniux.TokenVO, error) {
	if svc.qiniuClient == nil {
		return qiniux.TokenVO{}, errors.New("qiniu not configured")
	}
	return svc.qiniuClient.VideoUploadToken()
}

func (svc *Service) GetInstConfig(userID int64) (map[string]any, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no institution context")
		}
		return nil, err
	}

	config, err := svc.repo.GetInstConfig(context.Background(), instID)
	if err != nil {
		return nil, err
	}
	if len(config) == 0 {
		if err := svc.repo.CreateDefaultInstConfig(context.Background(), instID); err != nil {
			return nil, err
		}
		return svc.repo.GetInstConfig(context.Background(), instID)
	}
	return config, nil
}

func (svc *Service) SetInstConfig(userID int64, payload map[string]any) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}

	config, err := svc.repo.GetInstConfig(context.Background(), instID)
	if err != nil {
		return err
	}
	if len(config) == 0 {
		if err := svc.repo.CreateDefaultInstConfig(context.Background(), instID); err != nil {
			return err
		}
	}
	return svc.repo.UpdateInstConfig(context.Background(), instID, payload)
}

func (svc *Service) GetDefaultStudentFields(userID int64) ([]model.StudentFieldKey, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no institution context")
		}
		return nil, err
	}
	return svc.repo.ListStudentFields(context.Background(), instID, true)
}

func (svc *Service) GetCustomStudentFields(userID int64) ([]model.StudentFieldKey, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no institution context")
		}
		return nil, err
	}
	return svc.repo.ListStudentFields(context.Background(), instID, false)
}

func (svc *Service) AddCustomStudentField(userID int64, field model.StudentFieldKey) (int64, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("no institution context")
		}
		return 0, err
	}
	if strings.TrimSpace(field.FieldKey) == "" {
		return 0, errors.New("fieldKey is required")
	}
	if field.FieldType <= 0 {
		return 0, errors.New("fieldType is required")
	}
	sortValue, err := svc.repo.MaxStudentFieldSort(context.Background(), instID)
	if err != nil {
		return 0, err
	}
	field.Sort = sortValue + 1
	return svc.repo.CreateStudentCustomField(context.Background(), instID, field)
}

func (svc *Service) UpdateStudentFieldDisplayStatus(userID int64, field model.StudentFieldKey) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	current, err := svc.repo.GetStudentFieldByID(context.Background(), field.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("自定义字段不存在")
		}
		return err
	}
	if current.InstID != instID {
		return errors.New("自定义字段不存在")
	}
	if !current.CanDelete {
		return errors.New("当前字段不支持移除")
	}
	if strings.TrimSpace(field.UUID) != strings.TrimSpace(current.UUID) || field.Version != current.Version {
		return errors.New("当前信息已变更，请刷新后重试")
	}
	return svc.repo.UpdateStudentFieldDisplayStatus(context.Background(), field.ID, field.IsDisplay)
}

func (svc *Service) UpdateCustomStudentField(userID int64, field model.StudentFieldKey) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	if field.ID <= 0 {
		return errors.New("id is required")
	}
	current, err := svc.repo.GetStudentFieldByID(context.Background(), field.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("自定义字段不存在")
		}
		return err
	}
	if current.InstID != instID {
		return errors.New("自定义字段不存在")
	}
	if strings.TrimSpace(field.UUID) != strings.TrimSpace(current.UUID) || field.Version != current.Version {
		return errors.New("当前信息已变更，请刷新后重试")
	}
	return svc.repo.UpdateStudentCustomField(context.Background(), field)
}

func (svc *Service) DeleteCustomStudentField(userID int64, field model.StudentFieldKey) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	current, err := svc.repo.GetStudentFieldByID(context.Background(), field.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("自定义字段不存在")
		}
		return err
	}
	if current.InstID != instID {
		return errors.New("自定义字段不存在")
	}
	if strings.TrimSpace(field.UUID) != strings.TrimSpace(current.UUID) || field.Version != current.Version {
		return errors.New("当前信息已变更，请刷新后重试")
	}
	return svc.repo.DeleteStudentCustomField(context.Background(), field.ID)
}

func (svc *Service) GetTuitionAccountReadingList(userID int64, query model.TuitionAccountReadingListQueryDTO) (model.TuitionAccountReadingListResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.TuitionAccountReadingListResult{}, errors.New("no institution context")
		}
		return model.TuitionAccountReadingListResult{}, err
	}
	if strings.TrimSpace(query.QueryModel.StudentID) == "" {
		return model.TuitionAccountReadingListResult{}, errors.New("studentId is required")
	}
	return svc.repo.GetTuitionAccountReadingList(context.Background(), instID, strings.TrimSpace(query.QueryModel.StudentID))
}
