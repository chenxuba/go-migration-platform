package service

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) GetOneToOneListPage(userID int64, query model.OneToOneListQueryDTO) (model.OneToOneListResultVO, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.OneToOneListResultVO{}, errors.New("no institution context")
		}
		return model.OneToOneListResultVO{}, err
	}
	return svc.repo.PageOneToOneList(context.Background(), instID, query)
}

func (svc *Service) BatchAssignOneToOneClassTeacher(userID int64, dto model.OneToOneBatchAssignTeacherDTO) error {
	instID, operatorID, err := svc.resolveTeachingClassOperator(userID)
	if err != nil {
		return err
	}
	ids := parseTeachingClassIDs(dto.IDs)
	if len(ids) == 0 {
		return errors.New("请选择1对1记录")
	}
	classTeacherID, err := parseRequiredInt64String(dto.ClassTeacherID, "班主任")
	if err != nil {
		return err
	}
	return svc.repo.BatchAssignOneToOneClassTeacher(context.Background(), instID, operatorID, classTeacherID, ids)
}

func (svc *Service) BatchUpdateOneToOneClassTime(userID int64, dto model.OneToOneBatchClassTimeDTO) error {
	instID, operatorID, err := svc.resolveTeachingClassOperator(userID)
	if err != nil {
		return err
	}
	ids := parseTeachingClassIDs(dto.IDs)
	if len(ids) == 0 {
		return errors.New("请选择1对1记录")
	}
	return svc.repo.BatchUpdateOneToOneClassTime(context.Background(), instID, operatorID, ids, dto)
}

func (svc *Service) BatchUpdateOneToOneAttributes(userID int64, dto model.OneToOneBatchAttributeDTO) error {
	instID, operatorID, err := svc.resolveTeachingClassOperator(userID)
	if err != nil {
		return err
	}
	ids := parseTeachingClassIDs(dto.IDs)
	if len(ids) == 0 {
		return errors.New("请选择1对1记录")
	}
	if strings.TrimSpace(dto.DefaultTeacherID) == "" && dto.Status == nil && dto.ClassStudentStatus == nil {
		return errors.New("请至少修改一项1对1属性")
	}
	return svc.repo.BatchUpdateOneToOneAttributes(context.Background(), instID, operatorID, ids, dto)
}

func (svc *Service) resolveTeachingClassOperator(userID int64) (int64, int64, error) {
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

func parseTeachingClassIDs(ids []string) []int64 {
	result := make([]int64, 0, len(ids))
	for _, raw := range ids {
		value, _ := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
		if value <= 0 {
			continue
		}
		result = append(result, value)
	}
	return result
}

func parseRequiredInt64String(raw, field string) (int64, error) {
	value, _ := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if value <= 0 {
		return 0, errors.New(field + "不能为空")
	}
	return value, nil
}
