package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) ListClassrooms(userID int64, query model.ClassroomQueryDTO) ([]model.ClassroomVO, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no institution context")
		}
		return nil, err
	}
	return svc.repo.ListClassrooms(context.Background(), instID, query)
}

func (svc *Service) CreateClassroom(userID int64, input model.ClassroomMutation) (int64, error) {
	instID, operatorID, err := svc.resolveClassroomOperator(userID)
	if err != nil {
		return 0, err
	}
	if strings.TrimSpace(input.Name) == "" {
		return 0, errors.New("教室名称不能为空")
	}
	count, err := svc.repo.CountClassroomsByName(context.Background(), instID, input.Name, nil)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, errors.New("教室名称已存在")
	}
	return svc.repo.CreateClassroom(context.Background(), instID, operatorID, input)
}

func (svc *Service) UpdateClassroom(userID int64, input model.ClassroomMutation) error {
	instID, operatorID, err := svc.resolveClassroomOperator(userID)
	if err != nil {
		return err
	}
	if input.ID == nil || *input.ID <= 0 {
		return errors.New("教室ID不能为空")
	}
	if strings.TrimSpace(input.Name) == "" {
		return errors.New("教室名称不能为空")
	}
	count, err := svc.repo.CountClassroomsByName(context.Background(), instID, input.Name, input.ID)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("教室名称已存在")
	}
	return svc.repo.UpdateClassroom(context.Background(), instID, operatorID, input)
}

func (svc *Service) UpdateClassroomStatus(userID int64, input model.ClassroomStatusMutation) error {
	instID, operatorID, err := svc.resolveClassroomOperator(userID)
	if err != nil {
		return err
	}
	if input.ID == nil || *input.ID <= 0 {
		return errors.New("教室ID不能为空")
	}
	if input.Enabled == nil {
		return errors.New("enabled 不能为空")
	}
	return svc.repo.UpdateClassroomStatus(context.Background(), instID, operatorID, input)
}

func (svc *Service) DeleteClassroom(userID int64, classroomID int64) error {
	instID, operatorID, err := svc.resolveClassroomOperator(userID)
	if err != nil {
		return err
	}
	if classroomID <= 0 {
		return errors.New("教室ID不能为空")
	}
	return svc.repo.DeleteClassroom(context.Background(), instID, operatorID, classroomID)
}

func (svc *Service) resolveClassroomOperator(userID int64) (int64, int64, error) {
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
