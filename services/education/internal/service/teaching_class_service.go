package service

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

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

// CheckClassName 对标 CheckClassName：返回 true 表示名称已存在（不可用）
func (svc *Service) CheckClassName(userID int64, dto model.GroupClassCheckNameDTO) (bool, error) {
	if dto.IsOne2One {
		return svc.CheckOneToOneName(userID, model.OneToOneCheckNameDTO{
			Name:      dto.Name,
			ExceptID:  dto.ExceptID,
			IsOne2One: true,
		})
	}
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
	n, err := svc.repo.CountActiveGroupClassByName(context.Background(), instID, dto.Name, excludeID)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (svc *Service) CreateGroupClass(userID int64, dto model.GroupClassCreateDTO) (model.GroupClassCreateResult, error) {
	instID, operatorID, err := svc.resolveTeachingClassOperator(userID)
	if err != nil {
		return model.GroupClassCreateResult{}, err
	}
	id, err := svc.repo.CreateGroupClass(context.Background(), instID, operatorID, dto)
	if err != nil {
		return model.GroupClassCreateResult{}, err
	}
	return model.GroupClassCreateResult{ID: strconv.FormatInt(id, 10), Name: ""}, nil
}

func (svc *Service) UpdateGroupClass(userID int64, dto model.GroupClassUpdateDTO) error {
	instID, operatorID, err := svc.resolveTeachingClassOperator(userID)
	if err != nil {
		return err
	}
	return svc.repo.UpdateGroupClass(context.Background(), instID, operatorID, dto)
}

// CloseGroupClassOnly 仅结班（更新班级开班状态为已结班）
func (svc *Service) CloseGroupClassOnly(userID int64, id string) error {
	instID, operatorID, err := svc.resolveTeachingClassOperator(userID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(id) == "" {
		return errors.New("班级ID不能为空")
	}
	classID, err := strconv.ParseInt(strings.TrimSpace(id), 10, 64)
	if err != nil || classID <= 0 {
		return errors.New("班级ID无效")
	}
	return svc.repo.CloseGroupClassOnly(context.Background(), instID, operatorID, classID)
}

// ReopenGroupClassOnly 恢复开班（已结班 → 开班中）
func (svc *Service) ReopenGroupClassOnly(userID int64, id string) error {
	instID, operatorID, err := svc.resolveTeachingClassOperator(userID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(id) == "" {
		return errors.New("班级ID不能为空")
	}
	classID, err := strconv.ParseInt(strings.TrimSpace(id), 10, 64)
	if err != nil || classID <= 0 {
		return errors.New("班级ID无效")
	}
	return svc.repo.ReopenGroupClassOnly(context.Background(), instID, operatorID, classID)
}

func (svc *Service) PageGroupClasses(userID int64, body model.GroupClassListBody) (model.GroupClassListPageResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.GroupClassListPageResult{}, errors.New("no institution context")
		}
		return model.GroupClassListPageResult{}, err
	}
	return svc.repo.PageGroupClassList(context.Background(), instID, body.QueryModel, body.PageRequestModel)
}

func (svc *Service) PageMoveGroupClassCandidates(userID int64, body model.GroupClassMoveStudentCandidateListBody) (model.GroupClassListPageResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.GroupClassListPageResult{}, errors.New("no institution context")
		}
		return model.GroupClassListPageResult{}, err
	}

	currentClassID, err := strconv.ParseInt(strings.TrimSpace(body.QueryModel.CurrentClassID), 10, 64)
	if err != nil || currentClassID <= 0 {
		return model.GroupClassListPageResult{}, errors.New("currentClassId 无效")
	}
	studentID, err := strconv.ParseInt(strings.TrimSpace(body.QueryModel.StudentID), 10, 64)
	if err != nil || studentID <= 0 {
		return model.GroupClassListPageResult{}, errors.New("studentId 无效")
	}
	if strings.TrimSpace(body.QueryModel.LessonID) == "" {
		return model.GroupClassListPageResult{}, errors.New("lessonId 不能为空")
	}

	return svc.repo.PageMoveGroupClassCandidates(context.Background(), instID, currentClassID, studentID, body.QueryModel, body.PageRequestModel)
}

func (svc *Service) GetGroupClassDetail(userID int64, classIDStr string) (model.GroupClassDetailVO, error) {
	var zero model.GroupClassDetailVO
	classID, err := strconv.ParseInt(strings.TrimSpace(classIDStr), 10, 64)
	if err != nil || classID <= 0 {
		return zero, errors.New("id 无效")
	}
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return zero, errors.New("no institution context")
		}
		return zero, err
	}
	vo, err := svc.repo.GetGroupClassByID(context.Background(), instID, classID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return zero, errors.New("班级不存在")
		}
		return zero, err
	}
	return vo, nil
}

func (svc *Service) AggregateGroupClassStatistics(userID int64, q model.GroupClassListQueryModel) (model.GroupClassStatisticsVO, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.GroupClassStatisticsVO{}, errors.New("no institution context")
		}
		return model.GroupClassStatisticsVO{}, err
	}
	return svc.repo.AggregateGroupClassStatistics(context.Background(), instID, q)
}

func (svc *Service) ListGroupClassDrawerSchedules(userID int64, dto model.GroupClassDrawerSchedulesQueryDTO) (model.GroupClassDrawerScheduleListResult, error) {
	var zero model.GroupClassDrawerScheduleListResult
	classID, err := strconv.ParseInt(strings.TrimSpace(dto.ClassID), 10, 64)
	if err != nil || classID <= 0 {
		return zero, errors.New("classId 无效")
	}
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return zero, errors.New("no institution context")
		}
		return zero, err
	}
	ctx := context.Background()
	schedules, err := svc.repo.ListTeachingSchedules(ctx, instID, model.TeachingScheduleListQueryDTO{
		GroupClassIDs: []int64{classID},
		StartDate:     strings.TrimSpace(dto.StartDate),
		EndDate:       strings.TrimSpace(dto.EndDate),
		SortDirection: "asc",
	})
	if err != nil {
		return zero, err
	}
	if err := svc.repo.FillTeachingScheduleCallStatus(ctx, instID, schedules); err != nil {
		return zero, err
	}
	if len(schedules) == 0 {
		return model.GroupClassDrawerScheduleListResult{
			List:  []model.GroupClassDrawerScheduleVO{},
			Total: 0,
		}, nil
	}

	grouped := make(map[string][]model.TeachingScheduleVO)
	order := make([]string, 0, len(schedules))
	for _, item := range schedules {
		key := strings.TrimSpace(item.BatchNo)
		if key == "" {
			key = "single:" + strings.TrimSpace(item.ID)
		}
		if _, ok := grouped[key]; !ok {
			order = append(order, key)
		}
		grouped[key] = append(grouped[key], item)
	}

	result := make([]model.GroupClassDrawerScheduleVO, 0, len(grouped))
	for _, key := range order {
		groupItems := grouped[key]
		if len(groupItems) == 0 {
			continue
		}
		sort.SliceStable(groupItems, func(i, j int) bool {
			if groupItems[i].StartAt.Equal(groupItems[j].StartAt) {
				return strings.TrimSpace(groupItems[i].ID) < strings.TrimSpace(groupItems[j].ID)
			}
			return groupItems[i].StartAt.Before(groupItems[j].StartAt)
		})
		first := groupItems[0]
		last := groupItems[len(groupItems)-1]
		batchNo := strings.TrimSpace(first.BatchNo)
		repeatRule := "单次"
		weekdayText := weekdayTextFromDate(first.LessonDate)
		var batchMeta *model.TeachingScheduleBatchMeta
		if batchNo != "" {
			ids := make([]string, 0, len(groupItems))
			for _, item := range groupItems {
				ids = append(ids, item.ID)
			}
			batchDetail, detailErr := svc.repo.GetTeachingScheduleBatchDetail(ctx, instID, model.TeachingScheduleBatchDetailQueryDTO{
				BatchNo: batchNo,
				IDs:     ids,
			})
			if detailErr == nil {
				batchMeta = batchDetail.BatchMeta
			}
		}
		if batchMeta != nil {
			repeatRule = formatBatchRepeatRule(batchMeta)
			weekdayText = formatBatchWeekdayText(batchMeta, first.LessonDate)
		}
		if batchNo != "" && repeatRule == "单次" && len(groupItems) > 1 {
			repeatRule = "重复排课"
		}

		completedCount := 0
		for _, item := range groupItems {
			if item.CallStatus == 2 {
				completedCount++
			}
		}

		result = append(result, model.GroupClassDrawerScheduleVO{
			Key:              key,
			ClassID:          strconv.FormatInt(classID, 10),
			DetailScheduleID: first.ID,
			BatchNo:          batchNo,
			ScheduleCount:    len(groupItems),
			CompletedCount:   completedCount,
			Type:             groupScheduleType(batchMeta, len(groupItems)),
			RepeatRule:       repeatRule,
			DateRangeText:    buildScheduleDateRangeText(first.LessonDate, last.LessonDate),
			TimeText:         buildScheduleTimeText(first.StartAt, first.EndAt),
			WeekdayText:      weekdayText,
			TeacherName:      first.TeacherName,
			AssistantText:    joinAssistantNames(groupItems),
			ClassroomName:    first.ClassroomName,
			LessonName:       first.LessonName,
			BatchMeta:        batchMeta,
		})
	}
	return model.GroupClassDrawerScheduleListResult{
		List:  result,
		Total: len(result),
	}, nil
}

func (svc *Service) ListGroupClassDrawerWaitingRollCallSchedules(userID int64, dto model.GroupClassDrawerSchedulesQueryDTO) (model.GroupClassDrawerWaitingRollCallScheduleListResult, error) {
	var zero model.GroupClassDrawerWaitingRollCallScheduleListResult
	classID, err := strconv.ParseInt(strings.TrimSpace(dto.ClassID), 10, 64)
	if err != nil || classID <= 0 {
		return zero, errors.New("classId 无效")
	}
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return zero, errors.New("no institution context")
		}
		return zero, err
	}
	ctx := context.Background()
	schedules, err := svc.repo.ListTeachingSchedules(ctx, instID, model.TeachingScheduleListQueryDTO{
		GroupClassIDs: []int64{classID},
		StartDate:     strings.TrimSpace(dto.StartDate),
		EndDate:       strings.TrimSpace(dto.EndDate),
		SortDirection: "asc",
	})
	if err != nil {
		return zero, err
	}
	if err := svc.repo.FillTeachingScheduleCallStatus(ctx, instID, schedules); err != nil {
		return zero, err
	}

	result := make([]model.GroupClassDrawerWaitingRollCallScheduleVO, 0, len(schedules))
	for _, item := range schedules {
		if item.CallStatus == 2 {
			continue
		}
		result = append(result, model.GroupClassDrawerWaitingRollCallScheduleVO{
			ID:                     item.ID,
			BatchNo:                item.BatchNo,
			BatchSize:              item.BatchSize,
			ClassID:                item.TeachingClassID,
			LessonName:             item.LessonName,
			LessonDate:             item.LessonDate,
			StartAt:                item.StartAt,
			EndAt:                  item.EndAt,
			TeacherName:            item.TeacherName,
			AssistantText:          strings.Join(item.AssistantNames, "、"),
			ClassroomName:          item.ClassroomName,
			CallStatus:             item.CallStatus,
			CallStatusText:         item.CallStatusText,
			CanRollCall:            item.CanRollCall,
			RollCallDisabledReason: item.RollCallDisabledReason,
		})
	}
	return model.GroupClassDrawerWaitingRollCallScheduleListResult{
		List:  result,
		Total: len(result),
	}, nil
}

func (svc *Service) PageGroupClassOperationLogs(userID int64, body model.GroupClassOperationLogPagedListBody) (model.GroupClassOperationLogPagedListResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.GroupClassOperationLogPagedListResult{}, errors.New("no institution context")
		}
		return model.GroupClassOperationLogPagedListResult{}, err
	}
	return svc.repo.PageGroupClassOperationLogs(context.Background(), instID, body)
}

func (svc *Service) PageGroupClassEntryExitRecords(userID int64, body model.GroupClassEntryExitRecordPagedListBody) (model.GroupClassEntryExitRecordPagedListResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.GroupClassEntryExitRecordPagedListResult{}, errors.New("no institution context")
		}
		return model.GroupClassEntryExitRecordPagedListResult{}, err
	}
	return svc.repo.PageGroupClassEntryExitRecords(context.Background(), instID, body)
}

func (svc *Service) GetGroupClassStudentStatistics(userID int64, q model.GroupClassStudentQueryModel) (model.GroupClassStudentStatisticsVO, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.GroupClassStudentStatisticsVO{}, errors.New("no institution context")
		}
		return model.GroupClassStudentStatisticsVO{}, err
	}
	return svc.repo.GetGroupClassStudentStatistics(context.Background(), instID, q)
}

func (svc *Service) PageGroupClassStudents(userID int64, body model.GroupClassStudentPagedListBody) (model.GroupClassStudentPagedListResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.GroupClassStudentPagedListResult{}, errors.New("no institution context")
		}
		return model.GroupClassStudentPagedListResult{}, err
	}
	return svc.repo.PageGroupClassStudents(context.Background(), instID, body)
}

func (svc *Service) PageGroupClassFinishCoursePreview(userID int64, body model.GroupClassFinishCoursePreviewBody) (model.GroupClassStudentPagedListResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.GroupClassStudentPagedListResult{}, errors.New("no institution context")
		}
		return model.GroupClassStudentPagedListResult{}, err
	}
	statuses := append([]int(nil), body.QueryModel.ClassStudentStatus...)
	if len(statuses) == 0 {
		statuses = []int{1}
	}
	return svc.repo.PageGroupClassStudents(context.Background(), instID, model.GroupClassStudentPagedListBody{
		QueryModel: model.GroupClassStudentQueryModel{
			ID:                            strings.TrimSpace(body.QueryModel.ID),
			ClassID:                       strings.TrimSpace(body.QueryModel.ID),
			Status:                        statuses,
			IgnoreSuspendedTuitionAccount: false,
		},
		PageRequestModel: body.PageRequestModel,
	})
}

func (svc *Service) GetGroupClassStudentTeachingRecordCount(userID int64, dto model.GroupClassStudentTeachingRecordCountQueryDTO) ([]model.GroupClassStudentTeachingRecordCountVO, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no institution context")
		}
		return nil, err
	}
	return svc.repo.GetGroupClassStudentTeachingRecordCount(context.Background(), instID, dto)
}

// ListGroupClassStudentsByClassIDs 对标 Class/GetStudentListByClassIds
func (svc *Service) ListGroupClassStudentsByClassIDs(userID int64, dto model.GroupClassStudentListByClassIDsRequest) ([]model.GroupClassStudentListBucketVO, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no institution context")
		}
		return nil, err
	}
	return svc.repo.ListGroupClassStudentsByClassIDs(context.Background(), instID, dto.ClassIDs)
}

// BatchAssignGroupClassStudents 对标 Class/BatchAssignStudents
func (svc *Service) BatchAssignGroupClassStudents(userID int64, dto model.BatchAssignGroupClassStudentsRequest) error {
	instID, operatorID, err := svc.resolveTeachingClassOperator(userID)
	if err != nil {
		return err
	}
	classIDs := make([]int64, 0, len(dto.ClassIDs))
	seenC := make(map[int64]struct{})
	for _, raw := range dto.ClassIDs {
		id, perr := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
		if perr != nil || id <= 0 {
			return errors.New("classId 无效")
		}
		if _, ok := seenC[id]; ok {
			continue
		}
		seenC[id] = struct{}{}
		classIDs = append(classIDs, id)
	}
	if len(classIDs) == 0 {
		return errors.New("classIds 不能为空")
	}
	return svc.repo.BatchAssignGroupClassStudents(context.Background(), instID, operatorID, classIDs, dto.Students, dto.EnforceClassAssign)
}

func (svc *Service) RemoveGroupClassStudent(userID int64, dto model.GroupClassRemoveStudentDTO) error {
	instID, operatorID, err := svc.resolveTeachingClassOperator(userID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(dto.ClassID) == "" {
		return errors.New("classId 不能为空")
	}
	if strings.TrimSpace(dto.StudentID) == "" {
		return errors.New("studentId 不能为空")
	}
	return svc.repo.RemoveGroupClassStudent(context.Background(), instID, operatorID, dto)
}

func (svc *Service) MoveGroupClassStudent(userID int64, dto model.GroupClassMoveStudentDTO) error {
	instID, operatorID, err := svc.resolveTeachingClassOperator(userID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(dto.FromClassID) == "" {
		return errors.New("fromClassId 不能为空")
	}
	if strings.TrimSpace(dto.ToClassID) == "" {
		return errors.New("toClassId 不能为空")
	}
	if strings.TrimSpace(dto.StudentID) == "" {
		return errors.New("studentId 不能为空")
	}
	return svc.repo.MoveGroupClassStudent(context.Background(), instID, operatorID, dto)
}

// PageTuitionAccountsByLessonID 对标 TuitionAccount/GetTuitionAccountListByLessonId（集体班添加学员）
func (svc *Service) PageTuitionAccountsByLessonID(userID int64, body model.TuitionAccountListByLessonIDBody) (model.TuitionAccountListByLessonIDResult, error) {
	var zero model.TuitionAccountListByLessonIDResult
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return zero, errors.New("no institution context")
		}
		return zero, err
	}
	lessonID := strings.TrimSpace(body.QueryModel.LessonID)
	if lessonID == "" {
		return zero, errors.New("lessonId 不能为空")
	}
	courseIDs, _, err := svc.repo.ResolveGroupClassLessonCourseScope(context.Background(), instID, lessonID)
	if err != nil {
		return zero, err
	}
	var curClassID int64
	if s := strings.TrimSpace(body.QueryModel.ClassID); s != "" {
		if v, perr := strconv.ParseInt(s, 10, 64); perr == nil && v > 0 {
			curClassID = v
		}
	}
	stuIDs := make([]int64, 0, len(body.QueryModel.StudentIDs))
	for _, raw := range body.QueryModel.StudentIDs {
		if v, perr := strconv.ParseInt(strings.TrimSpace(raw), 10, 64); perr == nil && v > 0 {
			stuIDs = append(stuIDs, v)
		}
	}
	page := body.PageRequestModel.PageIndex
	size := body.PageRequestModel.PageSize
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	fil := body.QueryModel.TuitionAccountLessonPageFilters
	fil.StudentName = strings.TrimSpace(fil.StudentName)
	if fil.AgeMin != nil && fil.AgeMax != nil && *fil.AgeMin > *fil.AgeMax {
		fil.AgeMin, fil.AgeMax = fil.AgeMax, fil.AgeMin
	}
	list, total, err := svc.repo.PageTuitionAccountsByLessonForGroupAdd(context.Background(), instID, courseIDs, curClassID, stuIDs, page, size, fil)
	if err != nil {
		return zero, err
	}
	return model.TuitionAccountListByLessonIDResult{List: list, Total: total}, nil
}

func formatBatchRepeatRule(meta *model.TeachingScheduleBatchMeta) string {
	if meta == nil {
		return "单次"
	}
	switch strings.TrimSpace(strings.ToLower(meta.RepeatRule)) {
	case "daily":
		return "每天重复"
	case "weekly":
		return "每周重复"
	case "biweekly":
		return "隔周重复"
	case "alternateday":
		return "隔天重复"
	case "alternate_day":
		return "隔天重复"
	case "none", "":
		if strings.EqualFold(strings.TrimSpace(meta.SchedulingMode), "free") && len(meta.FreeSelectedDates) > 1 {
			return "自由排课"
		}
		return "单次"
	default:
		return meta.RepeatRule
	}
}

func formatBatchWeekdayText(meta *model.TeachingScheduleBatchMeta, fallbackDate string) string {
	if meta != nil && len(meta.SelectedWeekdays) > 0 {
		return strings.Join(meta.SelectedWeekdays, "、")
	}
	if meta != nil {
		switch strings.TrimSpace(strings.ToLower(meta.RepeatRule)) {
		case "daily":
			return "每天"
		case "alternateday", "alternate_day":
			return "隔天"
		}
	}
	return weekdayTextFromDate(fallbackDate)
}

func weekdayTextFromDate(dateText string) string {
	t, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(dateText), time.Local)
	if err != nil {
		return "-"
	}
	weeks := []string{"周日", "周一", "周二", "周三", "周四", "周五", "周六"}
	return weeks[int(t.Weekday())]
}

func buildScheduleDateRangeText(startDate, endDate string) string {
	startDate = strings.TrimSpace(startDate)
	endDate = strings.TrimSpace(endDate)
	if startDate == "" {
		return "-"
	}
	if endDate == "" || startDate == endDate {
		return startDate
	}
	return startDate + "~" + endDate
}

func buildScheduleTimeText(startAt, endAt time.Time) string {
	if startAt.IsZero() || endAt.IsZero() {
		return "-"
	}
	return startAt.Format("15:04") + "~" + endAt.Format("15:04")
}

func joinAssistantNames(items []model.TeachingScheduleVO) string {
	names := make([]string, 0, 4)
	seen := make(map[string]struct{})
	for _, item := range items {
		for _, name := range item.AssistantNames {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return "-"
	}
	return strings.Join(names, "、")
}

func groupScheduleType(meta *model.TeachingScheduleBatchMeta, scheduleCount int) int {
	if scheduleCount > 1 {
		return 1
	}
	if meta == nil {
		return 2
	}
	if strings.TrimSpace(strings.ToLower(meta.RepeatRule)) != "" && !strings.EqualFold(strings.TrimSpace(meta.RepeatRule), "none") {
		return 1
	}
	if len(meta.FreeSelectedDates) > 1 {
		return 1
	}
	return 2
}
