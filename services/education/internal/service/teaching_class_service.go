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

func (svc *Service) GetOneToOneDetail(userID int64, id string) (model.OneToOneDetailVO, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.OneToOneDetailVO{}, errors.New("no institution context")
		}
		return model.OneToOneDetailVO{}, err
	}
	classID, err := strconv.ParseInt(strings.TrimSpace(id), 10, 64)
	if err != nil || classID <= 0 {
		return model.OneToOneDetailVO{}, errors.New("1对1ID不能为空")
	}
	return svc.repo.GetOneToOneDetail(context.Background(), instID, classID)
}

func (svc *Service) ListStudentTuitionAccountsByStudentAndLesson(userID int64, dto model.StudentLessonTuitionAccountsQueryDTO) (model.StudentLessonTuitionAccountsResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.StudentLessonTuitionAccountsResult{}, errors.New("no institution context")
		}
		return model.StudentLessonTuitionAccountsResult{}, err
	}
	studentID, err := strconv.ParseInt(strings.TrimSpace(dto.StudentID), 10, 64)
	if err != nil || studentID <= 0 {
		return model.StudentLessonTuitionAccountsResult{}, errors.New("studentId 不能为空")
	}
	courseID, err := strconv.ParseInt(strings.TrimSpace(dto.LessonID), 10, 64)
	if err != nil || courseID <= 0 {
		return model.StudentLessonTuitionAccountsResult{}, errors.New("lessonId 不能为空")
	}
	list, err := svc.repo.ListStudentTuitionAccountsByStudentAndLesson(context.Background(), instID, studentID, courseID)
	if err != nil {
		return model.StudentLessonTuitionAccountsResult{}, err
	}
	if list == nil {
		list = []model.StudentLessonTuitionAccountItem{}
	}
	return model.StudentLessonTuitionAccountsResult{List: list}, nil
}

func (svc *Service) CheckOneToOneName(userID int64, dto model.OneToOneCheckNameDTO) (bool, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, errors.New("no institution context")
		}
		return false, err
	}
	var excludeID *int64
	if value, err := strconv.ParseInt(strings.TrimSpace(dto.ExceptID), 10, 64); err == nil && value > 0 {
		excludeID = &value
	}
	count, err := svc.repo.CountTeachingClassByName(context.Background(), instID, model.TeachingClassTypeOneToOne, dto.Name, excludeID)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (svc *Service) UpdateOneToOne(userID int64, dto model.OneToOneUpdateDTO) error {
	instID, operatorID, err := svc.resolveTeachingClassOperator(userID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(dto.ID) == "" {
		return errors.New("1对1ID不能为空")
	}
	if strings.TrimSpace(dto.StudentID) == "" {
		return errors.New("学员ID不能为空")
	}
	if strings.TrimSpace(dto.LessonID) == "" {
		return errors.New("课程ID不能为空")
	}
	if strings.TrimSpace(dto.Name) == "" {
		return errors.New("1对1名称不能为空")
	}
	if dto.DefaultClassTimeRecordMode <= 0 {
		dto.DefaultClassTimeRecordMode = 1
	}

	excludeID, _ := strconv.ParseInt(strings.TrimSpace(dto.ID), 10, 64)
	count, err := svc.repo.CountTeachingClassByName(context.Background(), instID, model.TeachingClassTypeOneToOne, dto.Name, &excludeID)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("1对1名称已存在")
	}

	return svc.repo.UpdateOneToOne(context.Background(), instID, operatorID, dto)
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
