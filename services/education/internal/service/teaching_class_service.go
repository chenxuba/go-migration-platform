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
	svc.SyncScheduledSuspendResumeTuitionAccountsOnce()
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
	svc.SyncScheduledSuspendResumeTuitionAccountsOnce()
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

func (svc *Service) ListStudentOneToOneDeductionTuitionAccounts(userID int64, dto model.StudentOneToOneDeductionAccountsQueryDTO) (model.StudentLessonTuitionAccountsResult, error) {
	svc.SyncScheduledSuspendResumeTuitionAccountsOnce()
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
	list, err := svc.repo.ListStudentOneToOneDeductionTuitionAccounts(context.Background(), instID, studentID)
	if err != nil {
		return model.StudentLessonTuitionAccountsResult{}, err
	}
	if list == nil {
		list = []model.StudentLessonTuitionAccountItem{}
	}
	return model.StudentLessonTuitionAccountsResult{List: list}, nil
}

func (svc *Service) ListStudentTuitionAccountsByStudentAndLesson(userID int64, dto model.StudentLessonTuitionAccountsQueryDTO) (model.StudentLessonTuitionAccountsResult, error) {
	svc.SyncScheduledSuspendResumeTuitionAccountsOnce()
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
	var teachingClassID int64
	if s := strings.TrimSpace(dto.TeachingClassID); s != "" && s != "0" {
		if v, perr := strconv.ParseInt(s, 10, 64); perr == nil && v > 0 {
			teachingClassID = v
		}
	}
	var orderCourseDetailID int64
	if s := strings.TrimSpace(dto.OrderCourseDetailID); s != "" && s != "0" {
		if v, perr := strconv.ParseInt(s, 10, 64); perr == nil && v > 0 {
			orderCourseDetailID = v
		}
	}
	list, err := svc.repo.ListStudentTuitionAccountsByStudentAndLesson(context.Background(), instID, studentID, courseID, teachingClassID, orderCourseDetailID)
	if err != nil {
		return model.StudentLessonTuitionAccountsResult{}, err
	}
	if list == nil {
		list = []model.StudentLessonTuitionAccountItem{}
	}
	return model.StudentLessonTuitionAccountsResult{List: list}, nil
}

// ListOneToOneLessonsByStudent 对标 QueryOne2OneLessonByStudentId：按学员查可建 1 对 1 的课程列表
func (svc *Service) ListOneToOneLessonsByStudent(userID int64, dto model.OneToOneLessonsByStudentQueryDTO) (model.OneToOneLessonsByStudentResult, error) {
	svc.SyncScheduledSuspendResumeTuitionAccountsOnce()
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.OneToOneLessonsByStudentResult{}, errors.New("no institution context")
		}
		return model.OneToOneLessonsByStudentResult{}, err
	}
	studentID, err := strconv.ParseInt(strings.TrimSpace(dto.StudentID), 10, 64)
	if err != nil || studentID <= 0 {
		return model.OneToOneLessonsByStudentResult{}, errors.New("studentId 不能为空")
	}
	statusFilter := dto.TuitionAccountStatus
	if len(statusFilter) == 0 {
		statusFilter = []int{1}
	}
	list, err := svc.repo.ListOneToOneLessonOptionsByStudent(context.Background(), instID, studentID, statusFilter)
	if err != nil {
		return model.OneToOneLessonsByStudentResult{}, err
	}
	if list == nil {
		list = []model.OneToOneLessonOptionVO{}
	}
	return model.OneToOneLessonsByStudentResult{List: list}, nil
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

// ExistOneToOneForStudentLesson 对标 ExistOne2One：data=true 表示已存在开班中的 1 对 1，不应再创建
func (svc *Service) ExistOneToOneForStudentLesson(userID int64, dto model.OneToOneExistDTO) (bool, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, errors.New("no institution context")
		}
		return false, err
	}
	studentID, err := strconv.ParseInt(strings.TrimSpace(dto.StudentID), 10, 64)
	if err != nil || studentID <= 0 {
		return false, errors.New("studentId 不能为空")
	}
	courseID, err := strconv.ParseInt(strings.TrimSpace(dto.LessonID), 10, 64)
	if err != nil || courseID <= 0 {
		return false, errors.New("lessonId 不能为空")
	}
	return svc.repo.ExistsActiveOneToOneForStudentCourse(context.Background(), instID, studentID, courseID)
}

func (svc *Service) CreateOneToOne(userID int64, dto model.OneToOneCreateDTO) (model.OneToOneCreateResult, error) {
	instID, operatorID, err := svc.resolveTeachingClassOperator(userID)
	if err != nil {
		return model.OneToOneCreateResult{}, err
	}
	if strings.TrimSpace(dto.StudentID) == "" {
		return model.OneToOneCreateResult{}, errors.New("学员ID不能为空")
	}
	if strings.TrimSpace(dto.LessonID) == "" {
		return model.OneToOneCreateResult{}, errors.New("课程ID不能为空")
	}
	if strings.TrimSpace(dto.TuitionAccountID) == "" {
		return model.OneToOneCreateResult{}, errors.New("请选择扣费学费账户")
	}
	if strings.TrimSpace(dto.Name) == "" {
		return model.OneToOneCreateResult{}, errors.New("1对1名称不能为空")
	}
	if dto.DefaultClassTimeRecordMode <= 0 {
		dto.DefaultClassTimeRecordMode = 1
	}

	if !dto.AllowDuplicateName {
		count, err := svc.repo.CountTeachingClassByName(context.Background(), instID, model.TeachingClassTypeOneToOne, dto.Name, nil)
		if err != nil {
			return model.OneToOneCreateResult{}, err
		}
		if count > 0 {
			return model.OneToOneCreateResult{}, errors.New("1对1名称已存在")
		}
	}

	classID, err := svc.repo.CreateOneToOne(context.Background(), instID, operatorID, dto)
	if err != nil {
		return model.OneToOneCreateResult{}, err
	}
	return model.OneToOneCreateResult{ID: strconv.FormatInt(classID, 10)}, nil
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

	if !dto.AllowDuplicateName {
		excludeID, _ := strconv.ParseInt(strings.TrimSpace(dto.ID), 10, 64)
		count, err := svc.repo.CountTeachingClassByName(context.Background(), instID, model.TeachingClassTypeOneToOne, dto.Name, &excludeID)
		if err != nil {
			return err
		}
		if count > 0 {
			return errors.New("1对1名称已存在")
		}
	}

	return svc.repo.UpdateOneToOne(context.Background(), instID, operatorID, dto)
}

func (svc *Service) SwitchOneToOneDefaultTuitionAccount(userID int64, dto model.OneToOneSwitchDefaultTuitionAccountDTO) error {
	instID, operatorID, err := svc.resolveTeachingClassOperator(userID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(dto.ID) == "" {
		return errors.New("1对1ID不能为空")
	}
	if strings.TrimSpace(dto.TuitionAccountID) == "" {
		return errors.New("tuitionAccountId不能为空")
	}
	return svc.repo.SwitchOneToOneDefaultTuitionAccount(context.Background(), instID, operatorID, dto)
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
	classTeacherIDs := parseTeachingClassIDs(dto.ClassTeacherIDs)
	if len(classTeacherIDs) == 0 && strings.TrimSpace(dto.ClassTeacherID) != "" {
		if v, e := strconv.ParseInt(strings.TrimSpace(dto.ClassTeacherID), 10, 64); e == nil && v > 0 {
			classTeacherIDs = []int64{v}
		}
	}
	if len(classTeacherIDs) == 0 {
		return errors.New("请选择班主任")
	}
	return svc.repo.BatchAssignOneToOneClassTeacher(context.Background(), instID, operatorID, classTeacherIDs, ids)
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
	if dto.ClassTimeRecordMode <= 0 {
		dto.ClassTimeRecordMode = 1
	}
	return svc.repo.BatchUpdateOneToOneClassTime(context.Background(), instID, operatorID, ids, dto)
}

// CloseOneToOneOnly 仅结班（更新班级开班状态为已结班，不处理结课与日程）
func (svc *Service) CloseOneToOneOnly(userID int64, id string) error {
	instID, operatorID, err := svc.resolveTeachingClassOperator(userID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(id) == "" {
		return errors.New("1对1ID不能为空")
	}
	classID, err := strconv.ParseInt(strings.TrimSpace(id), 10, 64)
	if err != nil || classID <= 0 {
		return errors.New("1对1ID无效")
	}
	return svc.repo.CloseOneToOneOnly(context.Background(), instID, operatorID, classID)
}

// AddCloseTuitionAccountOrder 手动结课下单（扣减账户、写流水，联动课消/学费变动/确认收入）
func (svc *Service) AddCloseTuitionAccountOrder(userID int64, dto model.CloseTuitionAccountOrderDTO) (model.CloseTuitionAccountOrderResult, error) {
	instID, operatorID, err := svc.resolveTeachingClassOperator(userID)
	if err != nil {
		return model.CloseTuitionAccountOrderResult{}, err
	}
	taID, err := strconv.ParseInt(strings.TrimSpace(dto.TuitionAccountID), 10, 64)
	if err != nil || taID <= 0 {
		return model.CloseTuitionAccountOrderResult{}, errors.New("tuitionAccountId 无效")
	}
	flowID, err := svc.repo.AddCloseTuitionAccountOrder(context.Background(), instID, operatorID, taID, dto.Quantity, dto.FreeQuantity, dto.Tuition, dto.Remark)
	if err != nil {
		return model.CloseTuitionAccountOrderResult{}, err
	}
	return model.CloseTuitionAccountOrderResult{
		ID:   strconv.FormatInt(flowID, 10),
		Name: "",
	}, nil
}

// ReopenOneToOneOnly 恢复开班（已结班 → 开班中）
func (svc *Service) ReopenOneToOneOnly(userID int64, id string) error {
	instID, operatorID, err := svc.resolveTeachingClassOperator(userID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(id) == "" {
		return errors.New("1对1ID不能为空")
	}
	classID, err := strconv.ParseInt(strings.TrimSpace(id), 10, 64)
	if err != nil || classID <= 0 {
		return errors.New("1对1ID无效")
	}
	return svc.repo.ReopenOneToOneOnly(context.Background(), instID, operatorID, classID)
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
