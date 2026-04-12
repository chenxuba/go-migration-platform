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

func (svc *Service) GetRollCallStatistics(userID int64, dto model.RollCallStatisticsQueryDTO) (model.RollCallStatisticsVO, error) {
	today := time.Now().Format("2006-01-02")

	todayItems, err := svc.listRollCallSchedules(userID, dto.QueryModel, today, today, "desc", "incomplete")
	if err != nil {
		return model.RollCallStatisticsVO{}, err
	}

	allItems, err := svc.listRollCallSchedules(userID, dto.QueryModel, "", today, "desc", "incomplete")
	if err != nil {
		return model.RollCallStatisticsVO{}, err
	}

	partialItems, err := svc.listRollCallSchedules(userID, dto.QueryModel, "", today, "desc", "partial")
	if err != nil {
		return model.RollCallStatisticsVO{}, err
	}

	return model.RollCallStatisticsVO{
		TodayCount:   len(todayItems),
		AllCount:     len(allItems),
		PartialCount: len(partialItems),
	}, nil
}

func (svc *Service) GetRollCallPagedList(userID int64, dto model.RollCallPagedListQueryDTO) (model.RollCallPagedListResult, error) {
	pageSize := dto.PageRequestModel.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 200 {
		pageSize = 200
	}

	pageIndex := dto.PageRequestModel.PageIndex
	if pageIndex <= 0 {
		pageIndex = 1
	}

	offset := (pageIndex - 1) * pageSize
	if dto.PageRequestModel.SkipCount > 0 {
		offset = dto.PageRequestModel.SkipCount
	}

	sortDirection := "asc"
	if dto.SortModel.ByStartDate == 2 {
		sortDirection = "desc"
	}

	items, err := svc.listRollCallSchedules(userID, dto.QueryModel, dto.QueryModel.StartDate, dto.QueryModel.EndDate, sortDirection, strings.TrimSpace(dto.QueryModel.CallStatusMode))
	if err != nil {
		return model.RollCallPagedListResult{}, err
	}

	sortRollCallSchedules(items, dto.SortModel.ByStartDate)

	total := len(items)
	if offset >= total {
		return model.RollCallPagedListResult{
			List:  []model.TeachingScheduleVO{},
			Total: total,
		}, nil
	}

	end := offset + pageSize
	if end > total {
		end = total
	}

	return model.RollCallPagedListResult{
		List:  items[offset:end],
		Total: total,
	}, nil
}

func (svc *Service) GetRollCallClassTimetable(userID int64, dto model.RollCallClassTimetableQueryDTO) (model.RollCallClassTimetableResult, error) {
	instID, err := svc.rollCallInstID(userID)
	if err != nil {
		return model.RollCallClassTimetableResult{}, err
	}
	return svc.repo.GetRollCallClassTimetable(context.Background(), instID, dto)
}

func (svc *Service) GetRollCallTeachingRecordStudentList(userID int64, dto model.RollCallTeachingRecordStudentListQueryDTO) (model.RollCallTeachingRecordStudentListResult, error) {
	instID, err := svc.rollCallInstID(userID)
	if err != nil {
		return model.RollCallTeachingRecordStudentListResult{}, err
	}
	return svc.repo.GetRollCallTeachingRecordStudentList(context.Background(), instID, dto)
}

func (svc *Service) GetRollCallStudentLeaveCount(userID int64, dto model.RollCallStudentLeaveCountQueryDTO) ([]model.RollCallStudentLeaveCountVO, error) {
	instID, err := svc.rollCallInstID(userID)
	if err != nil {
		return nil, err
	}
	return svc.repo.GetRollCallStudentLeaveCount(context.Background(), instID, dto)
}

func (svc *Service) GetRollCallStudentTuitionExtraInfo(userID int64, dto model.RollCallStudentTuitionExtraInfoQueryDTO) ([]model.RollCallStudentTuitionExtraInfoVO, error) {
	instID, err := svc.rollCallInstID(userID)
	if err != nil {
		return nil, err
	}
	return svc.repo.GetRollCallStudentTuitionExtraInfo(context.Background(), instID, dto)
}

func (svc *Service) GetRollCallStudentTuitionAccounts(userID int64, dto model.StudentLessonTuitionAccountsQueryDTO) (model.StudentLessonTuitionAccountsResult, error) {
	instID, err := svc.rollCallInstID(userID)
	if err != nil {
		return model.StudentLessonTuitionAccountsResult{}, err
	}
	list, err := svc.repo.GetRollCallStudentTuitionAccounts(context.Background(), instID, dto)
	if err != nil {
		return model.StudentLessonTuitionAccountsResult{}, err
	}
	if list == nil {
		list = []model.StudentLessonTuitionAccountItem{}
	}
	return model.StudentLessonTuitionAccountsResult{List: list}, nil
}

func (svc *Service) CheckRollCallTeachingRecordByTeacherAndTime(userID int64, dto model.RollCallCheckTeachingRecordByTeacherAndTimeDTO) error {
	instID, err := svc.rollCallInstID(userID)
	if err != nil {
		return err
	}
	return svc.repo.CheckRollCallTeachingRecordByTeacherAndTime(context.Background(), instID, dto)
}

func (svc *Service) BatchEstimateRollCallSufficientTuitionAccount(userID int64, dto model.RollCallBatchEstimateSufficientTuitionAccountDTO) (model.RollCallBatchEstimateSufficientTuitionAccountResult, error) {
	instID, err := svc.rollCallInstID(userID)
	if err != nil {
		return model.RollCallBatchEstimateSufficientTuitionAccountResult{}, err
	}
	return svc.repo.BatchEstimateRollCallSufficientTuitionAccount(context.Background(), instID, dto)
}

func (svc *Service) ConfirmRollCall(userID int64, dto model.RollCallConfirmDTO) (model.RollCallConfirmResult, error) {
	instID, err := svc.rollCallInstID(userID)
	if err != nil {
		return model.RollCallConfirmResult{}, err
	}
	operatorID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		return model.RollCallConfirmResult{}, err
	}
	return svc.repo.ConfirmRollCall(context.Background(), instID, operatorID, dto)
}

func (svc *Service) rollCallInstID(userID int64) (int64, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("no institution context")
		}
		return 0, err
	}
	return instID, nil
}

func (svc *Service) listRollCallSchedules(userID int64, query model.RollCallQueryModel, startDate, endDate, sortDirection string, callStatusMode string) ([]model.TeachingScheduleVO, error) {
	listQuery, err := buildRollCallTeachingScheduleQuery(query, startDate, endDate, sortDirection)
	if err != nil {
		return nil, err
	}

	items, err := svc.ListTeachingSchedules(userID, listQuery)
	if err != nil {
		return nil, err
	}

	items = filterRollCallSchedulesByTeacher(items, query.TeacherID, query.TeacherTypes)
	return filterRollCallSchedulesByStatus(items, callStatusMode), nil
}

func buildRollCallTeachingScheduleQuery(query model.RollCallQueryModel, startDate, endDate, sortDirection string) (model.TeachingScheduleListQueryDTO, error) {
	listQuery := model.TeachingScheduleListQueryDTO{
		StartDate:           strings.TrimSpace(startDate),
		EndDate:             strings.TrimSpace(endDate),
		SortDirection:       normalizeRollCallSortDirection(sortDirection),
		ScheduleTypeFilters: normalizeRollCallScheduleTypes(query.ScheduleTypes),
	}

	if strings.TrimSpace(listQuery.StartDate) == "" {
		listQuery.StartDate = strings.TrimSpace(query.StartDate)
	}
	if strings.TrimSpace(listQuery.EndDate) == "" {
		listQuery.EndDate = strings.TrimSpace(query.EndDate)
	}

	if id, ok, err := parsePositiveInt64String(query.LessonID); err != nil {
		return model.TeachingScheduleListQueryDTO{}, err
	} else if ok {
		listQuery.LessonIDs = []int64{id}
	}

	if id, ok, err := parsePositiveInt64String(query.ClassroomID); err != nil {
		return model.TeachingScheduleListQueryDTO{}, err
	} else if ok {
		listQuery.ClassroomIDs = []int64{id}
	}

	if id, ok, err := parsePositiveInt64String(query.ClassID); err != nil {
		return model.TeachingScheduleListQueryDTO{}, err
	} else if ok {
		listQuery.GroupClassIDs = []int64{id}
	}

	if id, ok, err := parsePositiveInt64String(query.OneToOneID); err != nil {
		return model.TeachingScheduleListQueryDTO{}, err
	} else if ok {
		listQuery.OneToOneClassIDs = []int64{id}
	}

	return listQuery, nil
}

func filterRollCallSchedulesByStatus(items []model.TeachingScheduleVO, mode string) []model.TeachingScheduleVO {
	mode = strings.TrimSpace(mode)
	filtered := make([]model.TeachingScheduleVO, 0, len(items))
	for _, item := range items {
		status := 1
		if item.CallStatus == 2 || item.CallStatus == 3 {
			status = item.CallStatus
		}
		switch mode {
		case "partial":
			if status == 3 {
				filtered = append(filtered, item)
			}
		case "all", "pending", "incomplete", "":
			if status == 1 || status == 3 {
				filtered = append(filtered, item)
			}
		default:
			if status == 1 || status == 3 {
				filtered = append(filtered, item)
			}
		}
	}
	return filtered
}

func normalizeRollCallScheduleTypes(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, item := range values {
		normalized := strings.TrimSpace(item)
		switch normalized {
		case "1":
			normalized = "group_class"
		case "2":
			normalized = "one_to_one"
		case "3":
			normalized = "trial"
		}
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}

func normalizeRollCallTeacherTypes(values []int) map[int]struct{} {
	if len(values) == 0 {
		return map[int]struct{}{
			1: {},
			2: {},
		}
	}

	result := make(map[int]struct{}, len(values))
	for _, item := range values {
		if item == 1 || item == 2 {
			result[item] = struct{}{}
		}
	}

	if len(result) == 0 {
		result[1] = struct{}{}
		result[2] = struct{}{}
	}
	return result
}

func filterRollCallSchedulesByTeacher(items []model.TeachingScheduleVO, teacherID string, teacherTypes []int) []model.TeachingScheduleVO {
	teacherID = strings.TrimSpace(teacherID)
	if teacherID == "" {
		return items
	}

	allowedTypes := normalizeRollCallTeacherTypes(teacherTypes)
	filtered := make([]model.TeachingScheduleVO, 0, len(items))
	for _, item := range items {
		matched := false
		if _, ok := allowedTypes[1]; ok && strings.TrimSpace(item.TeacherID) == teacherID {
			matched = true
		}
		if !matched {
			if _, ok := allowedTypes[2]; ok {
				for _, assistantID := range item.AssistantIDs {
					if strings.TrimSpace(assistantID) == teacherID {
						matched = true
						break
					}
				}
			}
		}
		if matched {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func normalizeRollCallSortDirection(value string) string {
	if strings.EqualFold(strings.TrimSpace(value), "desc") {
		return "desc"
	}
	return "asc"
}

func sortRollCallSchedules(items []model.TeachingScheduleVO, byStartDate int) {
	desc := byStartDate == 2
	sort.SliceStable(items, func(i, j int) bool {
		left := rollCallScheduleSortTime(items[i])
		right := rollCallScheduleSortTime(items[j])
		if left.Equal(right) {
			if desc {
				return strings.TrimSpace(items[i].ID) > strings.TrimSpace(items[j].ID)
			}
			return strings.TrimSpace(items[i].ID) < strings.TrimSpace(items[j].ID)
		}
		if desc {
			return left.After(right)
		}
		return left.Before(right)
	})
}

func rollCallScheduleSortTime(item model.TeachingScheduleVO) time.Time {
	if !item.StartAt.IsZero() {
		return item.StartAt
	}
	if strings.TrimSpace(item.LessonDate) != "" {
		if parsed, err := time.ParseInLocation("2006-01-02", item.LessonDate, time.Local); err == nil {
			return parsed
		}
	}
	return time.Time{}
}

func parsePositiveInt64String(value string) (int64, bool, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, false, nil
	}
	parsed, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil || parsed <= 0 {
		return 0, false, errors.New("筛选ID格式无效")
	}
	return parsed, true, nil
}
