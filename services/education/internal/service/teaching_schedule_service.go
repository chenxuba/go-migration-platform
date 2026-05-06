package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) CreateOneToOneSchedules(userID int64, dto model.CreateOneToOneSchedulesDTO) (model.CreateOneToOneSchedulesResult, error) {
	instID, operatorID, err := svc.resolveTeachingScheduleOperator(userID)
	if err != nil {
		return model.CreateOneToOneSchedulesResult{}, err
	}
	if strings.TrimSpace(dto.OneToOneID) == "" {
		return model.CreateOneToOneSchedulesResult{}, errors.New("请选择1对1")
	}
	return svc.repo.CreateOneToOneSchedules(context.Background(), instID, operatorID, dto)
}

func (svc *Service) CreateGroupClassSchedules(userID int64, dto model.CreateGroupClassSchedulesDTO) (model.CreateOneToOneSchedulesResult, error) {
	instID, operatorID, err := svc.resolveTeachingScheduleOperator(userID)
	if err != nil {
		return model.CreateOneToOneSchedulesResult{}, err
	}
	if strings.TrimSpace(dto.GroupClassID) == "" {
		return model.CreateOneToOneSchedulesResult{}, errors.New("请选择班级")
	}
	return svc.repo.CreateGroupClassSchedules(context.Background(), instID, operatorID, dto)
}

func (svc *Service) ValidateOneToOneSchedules(userID int64, dto model.CreateOneToOneSchedulesDTO) (model.TeachingScheduleValidationResult, error) {
	instID, _, err := svc.resolveTeachingScheduleOperator(userID)
	if err != nil {
		return model.TeachingScheduleValidationResult{}, err
	}
	if strings.TrimSpace(dto.OneToOneID) == "" {
		return model.TeachingScheduleValidationResult{}, errors.New("请选择1对1")
	}
	return svc.repo.ValidateOneToOneSchedules(context.Background(), instID, dto)
}

func (svc *Service) ValidateGroupClassSchedules(userID int64, dto model.CreateGroupClassSchedulesDTO) (model.TeachingScheduleValidationResult, error) {
	instID, _, err := svc.resolveTeachingScheduleOperator(userID)
	if err != nil {
		return model.TeachingScheduleValidationResult{}, err
	}
	if strings.TrimSpace(dto.GroupClassID) == "" {
		return model.TeachingScheduleValidationResult{}, errors.New("请选择班级")
	}
	return svc.repo.ValidateGroupClassSchedules(context.Background(), instID, dto)
}

func (svc *Service) CheckOneToOneScheduleAvailability(userID int64, dto model.CheckOneToOneScheduleAvailabilityDTO) (model.OneToOneScheduleAvailabilityResult, error) {
	instID, _, err := svc.resolveTeachingScheduleOperator(userID)
	if err != nil {
		return model.OneToOneScheduleAvailabilityResult{}, err
	}
	if strings.TrimSpace(dto.OneToOneID) == "" {
		return model.OneToOneScheduleAvailabilityResult{}, errors.New("请选择1对1")
	}
	return svc.repo.CheckOneToOneScheduleAvailability(context.Background(), instID, dto)
}

func (svc *Service) CheckAssistantScheduleAvailability(userID int64, dto model.CheckAssistantScheduleAvailabilityDTO) (model.AssistantScheduleAvailabilityResult, error) {
	instID, _, err := svc.resolveTeachingScheduleOperator(userID)
	if err != nil {
		return model.AssistantScheduleAvailabilityResult{}, err
	}
	if strings.TrimSpace(dto.OneToOneID) == "" {
		return model.AssistantScheduleAvailabilityResult{}, errors.New("请选择1对1")
	}
	return svc.repo.CheckAssistantScheduleAvailability(context.Background(), instID, dto)
}

func (svc *Service) CheckGroupClassAssistantScheduleAvailability(userID int64, dto model.CheckGroupClassAssistantScheduleAvailabilityDTO) (model.AssistantScheduleAvailabilityResult, error) {
	instID, _, err := svc.resolveTeachingScheduleOperator(userID)
	if err != nil {
		return model.AssistantScheduleAvailabilityResult{}, err
	}
	if strings.TrimSpace(dto.GroupClassID) == "" {
		return model.AssistantScheduleAvailabilityResult{}, errors.New("请选择班课")
	}
	return svc.repo.CheckGroupClassAssistantScheduleAvailability(context.Background(), instID, dto)
}

func (svc *Service) GetTeachingScheduleConflictDetail(userID int64, query model.TeachingScheduleConflictDetailQueryDTO) (model.TeachingScheduleValidationResult, error) {
	instID, _, err := svc.resolveTeachingScheduleOperator(userID)
	if err != nil {
		return model.TeachingScheduleValidationResult{}, err
	}
	if strings.TrimSpace(query.ID) == "" {
		return model.TeachingScheduleValidationResult{}, errors.New("缺少日程ID")
	}
	return svc.repo.GetTeachingScheduleConflictDetail(context.Background(), instID, query)
}

func (svc *Service) GetTeachingScheduleDetail(userID int64, query model.TeachingScheduleDetailQueryDTO) (model.TeachingScheduleDetailVO, error) {
	instID, _, err := svc.resolveTeachingScheduleOperator(userID)
	if err != nil {
		return model.TeachingScheduleDetailVO{}, err
	}
	if strings.TrimSpace(query.ID) == "" {
		return model.TeachingScheduleDetailVO{}, errors.New("缺少日程ID")
	}
	return svc.repo.GetTeachingScheduleDetail(context.Background(), instID, query)
}

func (svc *Service) PageTeachingScheduleStudentCandidates(userID int64, dto model.TeachingScheduleStudentCandidateQueryDTO) (model.TeachingScheduleStudentCandidatePagedResult, error) {
	instID, _, err := svc.resolveTeachingScheduleOperator(userID)
	if err != nil {
		return model.TeachingScheduleStudentCandidatePagedResult{}, err
	}
	if strings.TrimSpace(dto.QueryModel.ScheduleID) == "" {
		return model.TeachingScheduleStudentCandidatePagedResult{}, errors.New("缺少日程ID")
	}
	return svc.repo.PageTeachingScheduleStudentCandidates(context.Background(), instID, dto)
}

func (svc *Service) RemoveTeachingScheduleStudentCurrent(userID int64, dto model.TeachingScheduleStudentRemoveCurrentDTO) error {
	instID, operatorID, err := svc.resolveTeachingScheduleOperator(userID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(dto.ScheduleID) == "" {
		return errors.New("缺少日程ID")
	}
	if strings.TrimSpace(dto.StudentID) == "" {
		return errors.New("缺少学员ID")
	}
	return svc.repo.RemoveTeachingScheduleStudentCurrent(context.Background(), instID, operatorID, dto)
}

func (svc *Service) AddTeachingScheduleStudentsCurrent(userID int64, dto model.TeachingScheduleStudentsAddCurrentDTO) error {
	instID, operatorID, err := svc.resolveTeachingScheduleOperator(userID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(dto.ScheduleID) == "" {
		return errors.New("缺少日程ID")
	}
	if len(dto.StudentIDs) == 0 {
		return errors.New("请至少选择一位学员")
	}
	return svc.repo.AddTeachingScheduleStudentsCurrent(context.Background(), instID, operatorID, dto)
}

func (svc *Service) ListTeachingSchedules(userID int64, query model.TeachingScheduleListQueryDTO) ([]model.TeachingScheduleVO, error) {
	ctx := context.Background()
	instID, err := svc.repo.FindInstIDByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no institution context")
		}
		return nil, err
	}
	schedules, err := svc.repo.ListTeachingSchedules(ctx, instID, query)
	if err != nil {
		return nil, err
	}
	if err := svc.repo.FillTeachingScheduleCallStatus(ctx, instID, schedules); err != nil {
		return nil, err
	}
	if err := svc.annotateTeachingScheduleConflictsForQuery(ctx, instID, query, schedules); err != nil {
		return nil, err
	}
	if len(query.ConflictTypes) > 0 {
		schedules = filterTeachingSchedulesByConflictTypes(schedules, query.ConflictTypes)
	}
	return schedules, nil
}

func (svc *Service) GetPadTimetable(userID int64, query model.PadTimetableQueryDTO) (model.PadTimetableVO, error) {
	ctx := context.Background()
	instID, err := svc.repo.FindInstIDByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PadTimetableVO{}, errors.New("no institution context")
		}
		return model.PadTimetableVO{}, err
	}

	startDate, endDate, err := normalizePadTimetableDateRange(query.StartDate, query.EndDate)
	if err != nil {
		return model.PadTimetableVO{}, err
	}

	allTeachers, err := svc.repo.ListInstUsersForScheduleMatrix(ctx, instID)
	if err != nil {
		return model.PadTimetableVO{}, err
	}
	currentTeacherID, _ := svc.repo.FindInstUserIDByUserID(ctx, userID)

	periodGroups := svc.padTimetableConfiguredGroups(ctx, instID, startDate)
	preferredGroupTeacherID := currentTeacherID
	if strings.TrimSpace(query.PeriodGroupUUID) == "" && query.TeacherID > 0 {
		preferredGroupTeacherID = query.TeacherID
	}
	selectedGroupIndex := selectPadTimetablePeriodGroupIndex(periodGroups, query.PeriodGroupUUID, preferredGroupTeacherID)
	selectedGroupUUID := padTimetableDefaultPeriodGroupID
	var selectedGroup *smartExportGroup
	if selectedGroupIndex >= 0 && selectedGroupIndex < len(periodGroups) {
		selectedGroup = &periodGroups[selectedGroupIndex]
		selectedGroupUUID = strings.TrimSpace(selectedGroup.ID)
	}
	if selectedGroupUUID == "" {
		selectedGroupUUID = padTimetableDefaultPeriodGroupID
	}

	teacherIDs := []int64(nil)
	configuredSlots := []smartExportSlot(nil)
	if selectedGroup != nil {
		teacherIDs = uniquePositiveTeacherIDs(selectedGroup.BoundTeacherIDs)
		configuredSlots = selectedGroup.Slots
	}
	teachers := allTeachers
	if selectedGroup != nil {
		teachers, err = svc.appendMatrixRosterUsersByIDs(ctx, instID, allTeachers, teacherIDs)
		if err != nil {
			return model.PadTimetableVO{}, err
		}
		teachers = filterPadTimetableRosterByTeacherIDs(teachers, teacherIDs)
	}
	if selectedGroup == nil && currentTeacherID > 0 {
		teachers, err = svc.appendMatrixRosterUsersByIDs(ctx, instID, teachers, []int64{currentTeacherID})
		if err != nil {
			return model.PadTimetableVO{}, err
		}
	}

	selectedTeacherID := selectPadTimetableTeacherID(teachers, query.TeacherID, currentTeacherID)

	schedules := []model.TeachingScheduleVO{}
	if selectedTeacherID > 0 {
		schedules, err = svc.repo.ListTeachingSchedules(ctx, instID, model.TeachingScheduleListQueryDTO{
			StartDate:          startDate,
			EndDate:            endDate,
			SortDirection:      "asc",
			ScheduleTeacherIDs: []int64{selectedTeacherID},
		})
		if err != nil {
			return model.PadTimetableVO{}, err
		}
		if err := svc.repo.FillTeachingScheduleCallStatus(ctx, instID, schedules); err != nil {
			return model.PadTimetableVO{}, err
		}
		if err := svc.annotatePadTimetableConflicts(ctx, instID, schedules); err != nil {
			return model.PadTimetableVO{}, err
		}
	}

	teacherVOs, selectedTeacherName := padTimetableTeachers(teachers, currentTeacherID, selectedTeacherID)
	return model.PadTimetableVO{
		StartDate:               startDate,
		EndDate:                 endDate,
		SelectedPeriodGroupUUID: selectedGroupUUID,
		SelectedTeacherID:       selectedTeacherID,
		SelectedTeacherName:     selectedTeacherName,
		PeriodGroups:            padTimetablePeriodGroups(periodGroups, allTeachers),
		Teachers:                teacherVOs,
		Days:                    padTimetableDays(startDate, endDate),
		Slots:                   padTimetableSlots(configuredSlots),
		Items:                   padTimetableItems(schedules),
		Summary:                 padTimetableSummary(schedules),
	}, nil
}

func (svc *Service) annotatePadTimetableConflicts(ctx context.Context, instID int64, schedules []model.TeachingScheduleVO) error {
	if len(schedules) == 0 {
		return nil
	}
	candidates, err := svc.repo.ListTeachingSchedulesForConflictTargets(ctx, instID, schedules)
	if err != nil {
		return err
	}
	if len(candidates) == 0 {
		for i := range schedules {
			schedules[i].Conflict = false
			schedules[i].ConflictTypes = nil
		}
		return nil
	}
	annotateTeachingScheduleConflicts(candidates)
	applyAnnotatedConflictsByID(schedules, candidates)
	return nil
}

func normalizePadTimetableDateRange(startDate, endDate string) (string, string, error) {
	loc := padHomeLocation()
	startDate = strings.TrimSpace(startDate)
	endDate = strings.TrimSpace(endDate)
	if startDate == "" {
		now := time.Now().In(loc)
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		startDate = now.AddDate(0, 0, 1-weekday).Format("2006-01-02")
	}
	start, err := time.ParseInLocation("2006-01-02", startDate, loc)
	if err != nil {
		return "", "", errors.New("startDate 格式须为 YYYY-MM-DD")
	}
	if endDate == "" {
		endDate = start.AddDate(0, 0, 6).Format("2006-01-02")
	}
	end, err := time.ParseInLocation("2006-01-02", endDate, loc)
	if err != nil {
		return "", "", errors.New("endDate 格式须为 YYYY-MM-DD")
	}
	if end.Before(start) {
		return "", "", errors.New("endDate 不能早于 startDate")
	}
	return start.Format("2006-01-02"), end.Format("2006-01-02"), nil
}

func padTimetableTeachers(roster []model.InstUserScheduleRosterItem, currentID, selectedID int64) ([]model.PadTimetableTeacher, string) {
	items := make([]model.PadTimetableTeacher, 0, len(roster))
	selectedName := ""
	seen := make(map[int64]struct{}, len(roster))
	for _, teacher := range roster {
		if teacher.ID <= 0 {
			continue
		}
		if _, ok := seen[teacher.ID]; ok {
			continue
		}
		seen[teacher.ID] = struct{}{}
		name := strings.TrimSpace(teacher.Name)
		if name == "" {
			name = "未命名老师"
		}
		if teacher.ID == selectedID {
			selectedName = name
		}
		items = append(items, model.PadTimetableTeacher{
			ID:      teacher.ID,
			Name:    name,
			Current: teacher.ID == currentID,
		})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].ID == selectedID {
			return true
		}
		if items[j].ID == selectedID {
			return false
		}
		if items[i].Current != items[j].Current {
			return items[i].Current
		}
		return items[i].ID < items[j].ID
	})
	if selectedName == "" && len(items) > 0 {
		selectedName = items[0].Name
	}
	return items, selectedName
}

func padTimetableRosterContains(roster []model.InstUserScheduleRosterItem, teacherID int64) bool {
	for _, teacher := range roster {
		if teacher.ID == teacherID {
			return true
		}
	}
	return false
}

func padTimetableDays(startDate, endDate string) []model.PadTimetableDay {
	loc := padHomeLocation()
	start, err := time.ParseInLocation("2006-01-02", startDate, loc)
	if err != nil {
		return nil
	}
	end, err := time.ParseInLocation("2006-01-02", endDate, loc)
	if err != nil {
		return nil
	}
	items := make([]model.PadTimetableDay, 0, 7)
	for day := start; !day.After(end); day = day.AddDate(0, 0, 1) {
		items = append(items, model.PadTimetableDay{
			Date:    day.Format("2006-01-02"),
			Label:   chineseWeekdayShort(day.Weekday()),
			Weekday: chineseWeekday(day.Weekday()),
		})
	}
	return items
}

func chineseWeekdayShort(weekday time.Weekday) string {
	switch weekday {
	case time.Monday:
		return "周一"
	case time.Tuesday:
		return "周二"
	case time.Wednesday:
		return "周三"
	case time.Thursday:
		return "周四"
	case time.Friday:
		return "周五"
	case time.Saturday:
		return "周六"
	default:
		return "周日"
	}
}

const padTimetableDefaultPeriodGroupID = "default"

func (svc *Service) padTimetableConfiguredGroups(ctx context.Context, instID int64, startDate string) []smartExportGroup {
	loc := padHomeLocation()
	targetDate, err := time.ParseInLocation("2006-01-02", startDate, loc)
	if err != nil {
		targetDate = time.Now().In(loc)
	}
	cfg, err := svc.repo.GetInstPeriodConfigJSONForDate(ctx, instID, targetDate)
	if err != nil {
		return nil
	}
	groups := parseSmartExportGroups(cfg)
	if len(groups) == 0 {
		return nil
	}
	return groups
}

func selectPadTimetablePeriodGroupIndex(groups []smartExportGroup, requestedUUID string, currentTeacherID int64) int {
	if len(groups) == 0 {
		return -1
	}
	requestedUUID = strings.TrimSpace(requestedUUID)
	if requestedUUID != "" && requestedUUID != padTimetableDefaultPeriodGroupID {
		for index, group := range groups {
			if strings.TrimSpace(group.ID) == requestedUUID {
				return index
			}
		}
	}
	if currentTeacherID > 0 {
		for index, group := range groups {
			if containsTeacherID(group.BoundTeacherIDs, currentTeacherID) {
				return index
			}
		}
	}
	return 0
}

func selectPadTimetableTeacherID(roster []model.InstUserScheduleRosterItem, requestedID, currentID int64) int64 {
	if requestedID > 0 && padTimetableRosterContains(roster, requestedID) {
		return requestedID
	}
	if currentID > 0 && padTimetableRosterContains(roster, currentID) {
		return currentID
	}
	if len(roster) == 0 {
		return 0
	}
	return roster[0].ID
}

func filterPadTimetableRosterByTeacherIDs(roster []model.InstUserScheduleRosterItem, teacherIDs []int64) []model.InstUserScheduleRosterItem {
	if len(roster) == 0 || len(teacherIDs) == 0 {
		return nil
	}
	rosterByID := make(map[int64]model.InstUserScheduleRosterItem, len(roster))
	for _, teacher := range roster {
		if teacher.ID <= 0 {
			continue
		}
		rosterByID[teacher.ID] = teacher
	}
	out := make([]model.InstUserScheduleRosterItem, 0, len(teacherIDs))
	seen := make(map[int64]struct{}, len(teacherIDs))
	for _, id := range teacherIDs {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		if teacher, ok := rosterByID[id]; ok {
			out = append(out, teacher)
		}
	}
	return out
}

func padTimetablePeriodGroups(groups []smartExportGroup, allTeachers []model.InstUserScheduleRosterItem) []model.PadTimetablePeriodGroup {
	if len(groups) == 0 {
		return []model.PadTimetablePeriodGroup{{
			ID:          padTimetableDefaultPeriodGroupID,
			Name:        "默认时段",
			Sort:        0,
			StartTime:   "08:00",
			EndTime:     "18:20",
			LessonCount: len(defaultPadTimetableSlots()),
			TeacherIDs:  padTimetableTeacherIDsFromRoster(allTeachers),
		}}
	}
	items := make([]model.PadTimetablePeriodGroup, 0, len(groups))
	for index, group := range groups {
		start, end := padTimetableGroupTimeRange(group.Slots)
		name := strings.TrimSpace(group.Name)
		if name == "" {
			name = fmt.Sprintf("时段%d", index+1)
		}
		id := strings.TrimSpace(group.ID)
		if id == "" {
			id = fmt.Sprintf("group-%d", index+1)
		}
		items = append(items, model.PadTimetablePeriodGroup{
			ID:          id,
			Name:        name,
			Sort:        group.Sort,
			StartTime:   start,
			EndTime:     end,
			LessonCount: len(group.Slots),
			TeacherIDs:  uniquePositiveTeacherIDs(group.BoundTeacherIDs),
		})
	}
	return items
}

func padTimetableGroupTimeRange(slots []smartExportSlot) (string, string) {
	start := ""
	end := ""
	for _, slot := range slots {
		slotStart := strings.TrimSpace(slot.Start)
		slotEnd := strings.TrimSpace(slot.End)
		if slotStart == "" || slotEnd == "" {
			continue
		}
		if start == "" || slotStart < start {
			start = slotStart
		}
		if end == "" || slotEnd > end {
			end = slotEnd
		}
	}
	return start, end
}

func padTimetableTeacherIDsFromRoster(roster []model.InstUserScheduleRosterItem) []int64 {
	out := make([]int64, 0, len(roster))
	seen := make(map[int64]struct{}, len(roster))
	for _, teacher := range roster {
		if teacher.ID <= 0 {
			continue
		}
		if _, ok := seen[teacher.ID]; ok {
			continue
		}
		seen[teacher.ID] = struct{}{}
		out = append(out, teacher.ID)
	}
	return out
}

func padTimetableSlots(configuredSlots []smartExportSlot) []model.PadTimetableSlot {
	slots := configuredSlots
	if len(slots) == 0 {
		slots = defaultPadTimetableSlots()
	}
	slots = append([]smartExportSlot(nil), slots...)
	sort.SliceStable(slots, func(i, j int) bool {
		if slots[i].Index == slots[j].Index {
			if slots[i].Start == slots[j].Start {
				return slots[i].End < slots[j].End
			}
			return slots[i].Start < slots[j].Start
		}
		return slots[i].Index < slots[j].Index
	})
	items := make([]model.PadTimetableSlot, 0, len(slots))
	for index, item := range slots {
		start := strings.TrimSpace(item.Start)
		end := strings.TrimSpace(item.End)
		if start == "" || end == "" {
			continue
		}
		titleIndex := item.Index
		if titleIndex <= 0 {
			titleIndex = index + 1
		}
		items = append(items, model.PadTimetableSlot{
			Title:     fmt.Sprintf("第%d节", titleIndex),
			Time:      start + " - " + end,
			StartTime: start,
			EndTime:   end,
		})
	}
	return items
}

func defaultPadTimetableSlots() []smartExportSlot {
	return []smartExportSlot{
		{Index: 1, Start: "08:00", End: "08:40"},
		{Index: 2, Start: "08:50", End: "09:30"},
		{Index: 3, Start: "09:40", End: "10:20"},
		{Index: 4, Start: "10:30", End: "11:10"},
		{Index: 5, Start: "11:20", End: "12:00"},
		{Index: 6, Start: "13:30", End: "14:10"},
		{Index: 7, Start: "14:20", End: "15:00"},
		{Index: 8, Start: "15:10", End: "15:50"},
		{Index: 9, Start: "16:00", End: "16:40"},
		{Index: 10, Start: "16:50", End: "17:30"},
		{Index: 11, Start: "17:40", End: "18:20"},
	}
}

func padTimetableItems(schedules []model.TeachingScheduleVO) []model.PadTimetableItem {
	loc := padHomeLocation()
	items := make([]model.PadTimetableItem, 0, len(schedules))
	for _, schedule := range schedules {
		status, statusText := padTimetableScheduleStatus(schedule)
		items = append(items, model.PadTimetableItem{
			ID:                schedule.ID,
			BatchNo:           schedule.BatchNo,
			ClassType:         schedule.ClassType,
			TeachingClassID:   strings.TrimSpace(schedule.TeachingClassID),
			Date:              schedule.LessonDate,
			StartTime:         schedule.StartAt.In(loc).Format("15:04"),
			EndTime:           schedule.EndAt.In(loc).Format("15:04"),
			LessonName:        firstNonEmptyString(strings.TrimSpace(schedule.LessonName), "未命名课程"),
			TeachingClassName: strings.TrimSpace(schedule.TeachingClassName),
			StudentName:       strings.TrimSpace(schedule.StudentName),
			PersonName:        padTimetablePersonName(schedule),
			ClassroomName:     strings.TrimSpace(schedule.ClassroomName),
			TeacherID:         strings.TrimSpace(schedule.TeacherID),
			TeacherName:       strings.TrimSpace(schedule.TeacherName),
			AssistantIDs:      schedule.AssistantIDs,
			ClassroomID:       strings.TrimSpace(schedule.ClassroomID),
			Status:            status,
			StatusText:        statusText,
			Conflict:          schedule.Conflict,
		})
	}
	return items
}

func padTimetablePersonName(schedule model.TeachingScheduleVO) string {
	if name := strings.TrimSpace(schedule.StudentName); name != "" {
		return name
	}
	if name := strings.TrimSpace(schedule.TeachingClassName); name != "" {
		return name
	}
	return "未关联学员"
}

func padTimetableSummary(schedules []model.TeachingScheduleVO) model.PadTimetableSummary {
	var summary model.PadTimetableSummary
	summary.Total = len(schedules)
	for _, schedule := range schedules {
		if schedule.Conflict {
			summary.Conflict++
		}
		if padTimetableLooksTrial(schedule) {
			summary.Trial++
		}
		switch schedule.CallStatus {
		case 2:
			summary.Signed++
		case 3:
			summary.Partial++
		default:
			summary.Unsigned++
		}
	}
	return summary
}

func padTimetableScheduleStatus(schedule model.TeachingScheduleVO) (string, string) {
	if schedule.Conflict {
		return "conflict", "冲突"
	}
	if padTimetableLooksTrial(schedule) {
		return "trial", "试听"
	}
	switch schedule.CallStatus {
	case 2:
		return "signed", "已点名"
	case 3:
		return "partial", "部分"
	default:
		return "unsigned", "未点名"
	}
}

func padTimetableLooksTrial(schedule model.TeachingScheduleVO) bool {
	text := strings.TrimSpace(schedule.LessonName) + " " +
		strings.TrimSpace(schedule.TeachingClassName) + " " +
		strings.TrimSpace(schedule.StudentName)
	return strings.Contains(text, "试听")
}

func (svc *Service) PageConflictTeachingSchedules(userID int64, query model.TeachingScheduleConflictPageQueryDTO) (model.PageResult[model.TeachingScheduleVO], error) {
	pageIndex := query.PageRequestModel.PageIndex
	if pageIndex <= 0 {
		pageIndex = 1
	}
	pageSize := query.PageRequestModel.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 200 {
		pageSize = 200
	}

	schedules, err := svc.ListTeachingSchedules(userID, query.TeachingScheduleListQueryDTO)
	if err != nil {
		return model.PageResult[model.TeachingScheduleVO]{}, err
	}

	conflicts := make([]model.TeachingScheduleVO, 0, len(schedules))
	for _, item := range schedules {
		if item.Conflict {
			conflicts = append(conflicts, item)
		}
	}

	total := len(conflicts)
	offset := (pageIndex - 1) * pageSize
	if offset < 0 {
		offset = 0
	}
	if offset > total {
		offset = total
	}
	end := offset + pageSize
	if end > total {
		end = total
	}

	items := make([]model.TeachingScheduleVO, 0, end-offset)
	if offset < end {
		items = append(items, conflicts[offset:end]...)
	}

	return model.PageResult[model.TeachingScheduleVO]{
		Items:   items,
		Total:   total,
		Current: pageIndex,
		Size:    pageSize,
	}, nil
}

// ListTeachingSchedulesByTeacherMatrix 按「日期 × 教师」矩阵返回课表（结构对齐旧版机构总课表 scheduleListVoList）
func (svc *Service) ListTeachingSchedulesByTeacherMatrix(userID int64, query model.TeachingScheduleListQueryDTO) ([]model.TeachingScheduleMatrixDayVO, error) {
	if strings.TrimSpace(query.StartDate) == "" || strings.TrimSpace(query.EndDate) == "" {
		return nil, errors.New("startDate 与 endDate 不能为空")
	}
	days, err := expandInclusiveDates(query.StartDate, query.EndDate)
	if err != nil {
		return nil, err
	}

	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no institution context")
		}
		return nil, err
	}

	ctx := context.Background()
	roster, err := svc.repo.ListInstUsersForScheduleMatrix(ctx, instID)
	if err != nil {
		return nil, err
	}
	schedules, err := svc.repo.ListTeachingSchedules(ctx, instID, query)
	if err != nil {
		return nil, err
	}
	if err := svc.repo.FillTeachingScheduleCallStatus(ctx, instID, schedules); err != nil {
		return nil, err
	}
	if err := svc.annotateTeachingScheduleConflictsForQuery(ctx, instID, query, schedules); err != nil {
		return nil, err
	}

	allowTeachers, allowTeacherOrder, err := svc.resolveMatrixTeacherAllowList(ctx, instID, query)
	if err != nil {
		return nil, err
	}
	if len(query.ScheduleTeacherIDs) > 0 {
		selectedTeachers := make(map[int64]struct{}, len(query.ScheduleTeacherIDs))
		for _, id := range query.ScheduleTeacherIDs {
			if id > 0 {
				selectedTeachers[id] = struct{}{}
			}
		}
		if allowTeachers == nil {
			allowTeachers = selectedTeachers
			allowTeacherOrder = uniquePositiveTeacherIDs(query.ScheduleTeacherIDs)
		} else {
			for id := range allowTeachers {
				if _, ok := selectedTeachers[id]; !ok {
					delete(allowTeachers, id)
				}
			}
			allowTeacherOrder = filterTeacherOrderByAllowList(allowTeacherOrder, allowTeachers)
			for _, id := range query.ScheduleTeacherIDs {
				if id <= 0 {
					continue
				}
				if _, ok := allowTeachers[id]; ok && !containsTeacherID(allowTeacherOrder, id) {
					allowTeacherOrder = append(allowTeacherOrder, id)
				}
			}
		}
	}
	if len(allowTeacherOrder) > 0 {
		roster, err = svc.appendMatrixRosterUsersByIDs(ctx, instID, roster, allowTeacherOrder)
		if err != nil {
			return nil, err
		}
	}
	if allowTeachers == nil {
		roster, err = svc.appendMatrixRosterUsersByIDs(ctx, instID, roster, collectMatrixScheduleTeacherIDs(schedules))
		if err != nil {
			return nil, err
		}
	}
	teacherOrder, teacherNames := buildTeacherOrderForMatrix(roster, schedules)
	teacherOrder = prioritizeTeacherOrder(teacherOrder, allowTeacherOrder)
	keyed := make(map[string][]model.TeachingScheduleVO)
	for _, s := range schedules {
		tid := strings.TrimSpace(s.TeacherID)
		if tid != "" {
			k := s.LessonDate + "\t" + tid
			keyed[k] = append(keyed[k], s)
		}
		for _, aid := range s.AssistantIDs {
			aid = strings.TrimSpace(aid)
			if aid == "" {
				continue
			}
			k := s.LessonDate + "\t" + aid
			keyed[k] = append(keyed[k], s)
		}
	}

	matrixTeacherFilter := normalizeMatrixTeacherFilter(query.MatrixTeacherFilter)
	out := make([]model.TeachingScheduleMatrixDayVO, 0, len(days))
	for _, d := range days {
		if len(query.MatrixWeekdays) > 0 {
			wd := dateWeekdayMonToSun(d)
			if wd == 0 || !intSliceContains(query.MatrixWeekdays, wd) {
				continue
			}
		}

		cols := make([]model.TeachingScheduleMatrixTeacher, 0, len(teacherOrder))
		for _, tid := range teacherOrder {
			if allowTeachers != nil {
				if _, ok := allowTeachers[tid]; !ok {
					continue
				}
			}
			k := d + "\t" + strconv.FormatInt(tid, 10)
			raw := keyed[k]
			legacy := make([]model.TeachingScheduleInfoLegacyVO, 0, len(raw))
			for _, item := range raw {
				legacy = append(legacy, mapTeachingScheduleToLegacyVO(item, instID))
			}
			n := len(legacy)
			switch matrixTeacherFilter {
			case "has_class":
				if n == 0 {
					continue
				}
			case "no_class":
				if n > 0 {
					continue
				}
			}
			cols = append(cols, model.TeachingScheduleMatrixTeacher{
				TeacherName:        teacherNames[tid],
				TeacherID:          tid,
				ScheduleInfoVoList: legacy,
			})
		}

		if len(cols) == 0 && (matrixTeacherFilter == "has_class" || matrixTeacherFilter == "no_class") {
			continue
		}

		out = append(out, model.TeachingScheduleMatrixDayVO{
			ScheduleDate:       d,
			Width:              len(cols),
			ScheduleInfoVoList: nil, // 输出 JSON null
			ScheduleListVoList: cols,
		})
	}
	return out, nil
}

// resolveMatrixTeacherAllowList 非 nil 时表示仅展示这些教师列；nil 表示不做 ID 级筛选（与未传时段组一致）。
// 优先使用 periodGroupUuid 在库中的关联；若无则使用 matrixTeacherIds。
func (svc *Service) resolveMatrixTeacherAllowList(ctx context.Context, instID int64, query model.TeachingScheduleListQueryDTO) (map[int64]struct{}, []int64, error) {
	u := strings.TrimSpace(query.PeriodGroupUUID)
	if u != "" {
		targetDate := time.Now()
		if strings.TrimSpace(query.StartDate) != "" {
			if parsed, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(query.StartDate), time.Local); err == nil {
				targetDate = parsed
			}
		}
		ids, err := svc.repo.ListPeriodTeacherUserIDsByGroupUUIDForDate(ctx, instID, u, targetDate)
		if err != nil {
			return nil, nil, err
		}
		ordered := uniquePositiveTeacherIDs(ids)
		if len(ordered) > 0 {
			m := make(map[int64]struct{}, len(ordered))
			for _, id := range ordered {
				m[id] = struct{}{}
			}
			return m, ordered, nil
		}
		exists, err := svc.repo.HasPeriodGroupUUID(ctx, instID, u)
		if err != nil {
			return nil, nil, err
		}
		if exists {
			return map[int64]struct{}{}, []int64{}, nil
		}
	}
	ordered := uniquePositiveTeacherIDs(query.MatrixTeacherIDs)
	if len(ordered) > 0 {
		m := make(map[int64]struct{}, len(ordered))
		for _, id := range ordered {
			m[id] = struct{}{}
		}
		return m, ordered, nil
	}
	return nil, nil, nil
}

func (svc *Service) appendMatrixRosterUsersByIDs(ctx context.Context, instID int64, roster []model.InstUserScheduleRosterItem, teacherIDs []int64) ([]model.InstUserScheduleRosterItem, error) {
	if len(teacherIDs) == 0 {
		return roster, nil
	}
	existing := make(map[int64]struct{}, len(roster))
	for _, item := range roster {
		if item.ID > 0 {
			existing[item.ID] = struct{}{}
		}
	}
	missing := make([]int64, 0, len(teacherIDs))
	for _, id := range teacherIDs {
		if id <= 0 {
			continue
		}
		if _, ok := existing[id]; ok {
			continue
		}
		existing[id] = struct{}{}
		missing = append(missing, id)
	}
	if len(missing) == 0 {
		return roster, nil
	}
	extras, err := svc.repo.ListInstUsersForScheduleMatrixByIDs(ctx, instID, missing)
	if err != nil {
		return nil, err
	}
	if len(extras) == 0 {
		return roster, nil
	}
	return append(roster, extras...), nil
}

func collectMatrixScheduleTeacherIDs(schedules []model.TeachingScheduleVO) []int64 {
	if len(schedules) == 0 {
		return nil
	}
	out := make([]int64, 0, len(schedules))
	seen := make(map[int64]struct{})
	appendID := func(raw string) {
		id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
		if err != nil || id <= 0 {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	for _, item := range schedules {
		appendID(item.TeacherID)
		for _, assistantID := range item.AssistantIDs {
			appendID(assistantID)
		}
	}
	return out
}

func uniquePositiveTeacherIDs(ids []int64) []int64 {
	if len(ids) == 0 {
		return nil
	}
	out := make([]int64, 0, len(ids))
	seen := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func filterTeacherOrderByAllowList(order []int64, allow map[int64]struct{}) []int64 {
	if len(order) == 0 || allow == nil {
		return order
	}
	out := make([]int64, 0, len(order))
	for _, id := range order {
		if _, ok := allow[id]; ok {
			out = append(out, id)
		}
	}
	return out
}

func containsTeacherID(ids []int64, target int64) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}
	return false
}

func prioritizeTeacherOrder(base, preferred []int64) []int64 {
	if len(preferred) == 0 || len(base) == 0 {
		return base
	}
	baseSet := make(map[int64]struct{}, len(base))
	for _, id := range base {
		if id > 0 {
			baseSet[id] = struct{}{}
		}
	}
	out := make([]int64, 0, len(base))
	seen := make(map[int64]struct{}, len(base))
	for _, id := range preferred {
		if id <= 0 {
			continue
		}
		if _, ok := baseSet[id]; !ok {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	for _, id := range base {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func normalizeMatrixTeacherFilter(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "has_class", "has-class", "hasclass":
		return "has_class"
	case "no_class", "no-class", "noclass":
		return "no_class"
	default:
		return ""
	}
}

func dateWeekdayMonToSun(dateISO string) int {
	t, err := time.ParseInLocation("2006-01-02", dateISO, time.Local)
	if err != nil {
		return 0
	}
	w := int(t.Weekday())
	if w == 0 {
		return 7
	}
	return w
}

func intSliceContains(slice []int, v int) bool {
	for _, x := range slice {
		if x == v {
			return true
		}
	}
	return false
}

func expandInclusiveDates(startStr, endStr string) ([]string, error) {
	start, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(startStr), time.Local)
	if err != nil {
		return nil, errors.New("startDate 格式须为 YYYY-MM-DD")
	}
	end, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(endStr), time.Local)
	if err != nil {
		return nil, errors.New("endDate 格式须为 YYYY-MM-DD")
	}
	if end.Before(start) {
		return nil, errors.New("endDate 不能早于 startDate")
	}
	var days []string
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		days = append(days, d.Format("2006-01-02"))
	}
	return days, nil
}

func buildTeacherOrderForMatrix(roster []model.InstUserScheduleRosterItem, _ []model.TeachingScheduleVO) (order []int64, names map[int64]string) {
	names = make(map[int64]string)
	order = make([]int64, 0, len(roster))
	seen := make(map[int64]struct{})

	for _, r := range roster {
		if r.ID <= 0 {
			continue
		}
		nm := strings.TrimSpace(r.Name)
		if nm == "" {
			nm = "-"
		}
		names[r.ID] = nm
		seen[r.ID] = struct{}{}
		order = append(order, r.ID)
	}

	return order, names
}

func annotateTeachingScheduleConflicts(schedules []model.TeachingScheduleVO) {
	for i := range schedules {
		schedules[i].Conflict = false
		schedules[i].ConflictTypes = nil
	}
	if len(schedules) <= 1 {
		return
	}

	studentIDSets := make([]map[string]struct{}, len(schedules))
	assistantIDSets := make([]map[string]struct{}, len(schedules))
	for i, item := range schedules {
		studentIDSets[i] = parseDelimitedPositiveIDSet(item.StudentID)
		assistantIDSets[i] = parseStringSliceIDSet(item.AssistantIDs)
	}

	grouped := make(map[string][]int)
	for i, item := range schedules {
		grouped[item.LessonDate] = append(grouped[item.LessonDate], i)
	}

	for _, indexes := range grouped {
		sort.Slice(indexes, func(i, j int) bool {
			left := schedules[indexes[i]]
			right := schedules[indexes[j]]
			if left.StartAt.Equal(right.StartAt) {
				return left.EndAt.Before(right.EndAt)
			}
			return left.StartAt.Before(right.StartAt)
		})

		for i := 0; i < len(indexes); i++ {
			leftIndex := indexes[i]
			left := schedules[leftIndex]
			for j := i + 1; j < len(indexes); j++ {
				rightIndex := indexes[j]
				right := schedules[rightIndex]
				if !left.EndAt.After(right.StartAt) {
					break
				}
				if !right.EndAt.After(left.StartAt) {
					continue
				}
				if sameNonEmptyString(left.StudentID, right.StudentID) {
					appendScheduleConflictType(&schedules[leftIndex], "学员")
					appendScheduleConflictType(&schedules[rightIndex], "学员")
				}
				if nonEmptyIDSetsOverlap(studentIDSets[leftIndex], studentIDSets[rightIndex]) {
					appendScheduleConflictType(&schedules[leftIndex], "学员")
					appendScheduleConflictType(&schedules[rightIndex], "学员")
				}
				if sameNonEmptyString(left.TeacherID, right.TeacherID) {
					appendScheduleConflictType(&schedules[leftIndex], "老师")
					appendScheduleConflictType(&schedules[rightIndex], "老师")
				}
				if sameNonEmptyString(left.ClassroomID, right.ClassroomID) {
					appendScheduleConflictType(&schedules[leftIndex], "教室")
					appendScheduleConflictType(&schedules[rightIndex], "教室")
				}
				if left.ClassType == model.TeachingClassTypeNormal &&
					right.ClassType == model.TeachingClassTypeNormal &&
					sameNonEmptyString(left.TeachingClassID, right.TeachingClassID) {
					appendScheduleConflictType(&schedules[leftIndex], "班级")
					appendScheduleConflictType(&schedules[rightIndex], "班级")
				}
				if nonEmptyIDSetsOverlap(assistantIDSets[leftIndex], assistantIDSets[rightIndex]) {
					appendScheduleConflictType(&schedules[leftIndex], "助教")
					appendScheduleConflictType(&schedules[rightIndex], "助教")
				}
			}
		}
	}
}

func (svc *Service) annotateTeachingScheduleConflictsForQuery(ctx context.Context, instID int64, query model.TeachingScheduleListQueryDTO, schedules []model.TeachingScheduleVO) error {
	if len(schedules) == 0 {
		return nil
	}
	if !needsFullConflictAnnotation(query) {
		annotateTeachingScheduleConflicts(schedules)
		return nil
	}

	conflictScopeQuery := buildConflictAnnotationQuery(query)
	allSchedules, err := svc.repo.ListTeachingSchedules(ctx, instID, conflictScopeQuery)
	if err != nil {
		return err
	}
	annotateTeachingScheduleConflicts(allSchedules)
	applyAnnotatedConflictsByID(schedules, allSchedules)
	return nil
}

func needsFullConflictAnnotation(query model.TeachingScheduleListQueryDTO) bool {
	return strings.TrimSpace(query.StudentID) != "" ||
		len(query.ScheduleTeacherIDs) > 0 ||
		len(query.ClassroomIDs) > 0 ||
		len(query.GroupClassIDs) > 0 ||
		len(query.OneToOneClassIDs) > 0 ||
		len(query.LessonIDs) > 0 ||
		len(query.ScheduleTypeFilters) > 0 ||
		len(query.CallStatusFilters) > 0
}

func buildConflictAnnotationQuery(query model.TeachingScheduleListQueryDTO) model.TeachingScheduleListQueryDTO {
	return model.TeachingScheduleListQueryDTO{
		StartDate: strings.TrimSpace(query.StartDate),
		EndDate:   strings.TrimSpace(query.EndDate),
	}
}

func applyAnnotatedConflictsByID(targets []model.TeachingScheduleVO, annotated []model.TeachingScheduleVO) {
	conflictByID := make(map[string]model.TeachingScheduleVO, len(annotated))
	for _, item := range annotated {
		id := strings.TrimSpace(item.ID)
		if id == "" {
			continue
		}
		conflictByID[id] = item
	}
	for i := range targets {
		targets[i].Conflict = false
		targets[i].ConflictTypes = nil
		id := strings.TrimSpace(targets[i].ID)
		if id == "" {
			continue
		}
		if matched, ok := conflictByID[id]; ok {
			targets[i].Conflict = matched.Conflict
			targets[i].ConflictTypes = append([]string(nil), matched.ConflictTypes...)
		}
	}
}

func appendScheduleConflictType(item *model.TeachingScheduleVO, conflictType string) {
	if item == nil || strings.TrimSpace(conflictType) == "" {
		return
	}
	if !containsStringValue(item.ConflictTypes, conflictType) {
		item.ConflictTypes = append(item.ConflictTypes, conflictType)
		sort.Strings(item.ConflictTypes)
	}
	item.Conflict = len(item.ConflictTypes) > 0
}

func containsStringValue(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

func sameNonEmptyString(left, right string) bool {
	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	return left != "" && right != "" && left == right
}

func parseDelimitedPositiveIDSet(raw string) map[string]struct{} {
	result := make(map[string]struct{})
	for _, part := range strings.Split(strings.TrimSpace(raw), ",") {
		part = strings.TrimSpace(part)
		if part == "" || part == "0" {
			continue
		}
		result[part] = struct{}{}
	}
	return result
}

func parseStringSliceIDSet(values []string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		text := strings.TrimSpace(value)
		if text == "" || text == "0" {
			continue
		}
		result[text] = struct{}{}
	}
	return result
}

func nonEmptyIDSetsOverlap(left, right map[string]struct{}) bool {
	if len(left) == 0 || len(right) == 0 {
		return false
	}
	if len(left) > len(right) {
		left, right = right, left
	}
	for id := range left {
		if _, ok := right[id]; ok {
			return true
		}
	}
	return false
}

func filterTeachingSchedulesByConflictTypes(schedules []model.TeachingScheduleVO, conflictTypes []string) []model.TeachingScheduleVO {
	if len(conflictTypes) == 0 || len(schedules) == 0 {
		return schedules
	}
	typeSet := make(map[string]struct{}, len(conflictTypes))
	for _, item := range conflictTypes {
		text := strings.TrimSpace(item)
		if text == "" {
			continue
		}
		typeSet[text] = struct{}{}
	}
	if len(typeSet) == 0 {
		return schedules
	}

	filtered := make([]model.TeachingScheduleVO, 0, len(schedules))
	for _, item := range schedules {
		if !item.Conflict {
			continue
		}
		matched := false
		for _, conflictType := range item.ConflictTypes {
			if _, ok := typeSet[strings.TrimSpace(conflictType)]; ok {
				matched = true
				break
			}
		}
		if matched {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func mapTeachingScheduleToLegacyVO(v model.TeachingScheduleVO, instID int64) model.TeachingScheduleInfoLegacyVO {
	id, _ := strconv.ParseInt(v.ID, 10, 64)
	tid, _ := strconv.ParseInt(strings.TrimSpace(v.TeacherID), 10, 64)
	cid, _ := strconv.ParseInt(strings.TrimSpace(v.LessonID), 10, 64)
	classIDVal, _ := strconv.ParseInt(strings.TrimSpace(v.TeachingClassID), 10, 64)

	minutes := int(v.EndAt.Sub(v.StartAt).Minutes())
	if minutes < 0 {
		minutes = 0
	}

	teacherList := []model.ScheduleLegacyPersonVO{
		{Name: v.TeacherName, ID: tid, Type: 0, Disabled: false},
	}
	for i, aid := range v.AssistantIDs {
		aid = strings.TrimSpace(aid)
		if aid == "" {
			continue
		}
		aidInt, err := strconv.ParseInt(aid, 10, 64)
		if err != nil {
			continue
		}
		nm := ""
		if i < len(v.AssistantNames) {
			nm = strings.TrimSpace(v.AssistantNames[i])
		}
		teacherList = append(teacherList, model.ScheduleLegacyPersonVO{Name: nm, ID: aidInt, Type: 0, Disabled: false})
	}
	studentIDParts := make([]string, 0)
	for _, part := range strings.Split(strings.TrimSpace(v.StudentID), ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			studentIDParts = append(studentIDParts, part)
		}
	}
	studentNameParts := make([]string, 0)
	for _, part := range strings.Split(strings.TrimSpace(v.StudentName), "、") {
		part = strings.TrimSpace(part)
		if part != "" {
			studentNameParts = append(studentNameParts, part)
		}
	}
	studentCount := len(studentIDParts)
	if len(studentNameParts) > studentCount {
		studentCount = len(studentNameParts)
	}
	studentList := make([]model.ScheduleLegacyPersonVO, 0, studentCount)
	fallbackStudentName := strings.TrimSpace(v.TeachingClassName)
	if fallbackStudentName == "" {
		fallbackStudentName = "-"
	}
	requireRealStudentID := v.ClassType == model.TeachingClassTypeNormal
	for i := 0; i < studentCount; i++ {
		var studentID int64
		if i < len(studentIDParts) {
			studentID, _ = strconv.ParseInt(strings.TrimSpace(studentIDParts[i]), 10, 64)
		}
		studentName := ""
		if i < len(studentNameParts) {
			studentName = strings.TrimSpace(studentNameParts[i])
		}
		if requireRealStudentID {
			if studentID <= 0 {
				continue
			}
		} else if studentID <= 0 && studentName == "" {
			continue
		}
		studentList = append(studentList, model.ScheduleLegacyPersonVO{
			Name: func() string {
				if studentName != "" {
					return studentName
				}
				return fallbackStudentName
			}(),
			ID:   studentID,
			Type: 1,
		})
	}
	if len(studentList) == 0 && !requireRealStudentID {
		name := strings.TrimSpace(v.StudentName)
		if name == "" {
			name = fallbackStudentName
		}
		studentList = []model.ScheduleLegacyPersonVO{{
			Name: name,
			ID:   0,
			Type: 1,
		}}
	}

	batchID, _ := strconv.ParseInt(strings.TrimSpace(v.BatchNo), 10, 64)

	var classIDPtr *int64
	if classIDVal > 0 {
		classIDPtr = &classIDVal
	}

	return model.TeachingScheduleInfoLegacyVO{
		ID:                     id,
		InstID:                 instID,
		BatchNo:                strings.TrimSpace(v.BatchNo),
		BatchID:                batchID,
		ModifyBatchID:          batchID,
		CourseID:               cid,
		ClassID:                classIDPtr,
		ScheduleDate:           v.LessonDate,
		ScheduleStartTime:      v.StartAt.Format("15:04"),
		ScheduleEndTime:        v.EndAt.Format("15:04"),
		ScheduleStatus:         v.Status,
		CallStatus:             v.CallStatus,
		CallStatusText:         v.CallStatusText,
		CanRollCall:            v.CanRollCall,
		RollCallDisabledReason: v.RollCallDisabledReason,
		Conflict:               v.Conflict,
		ConflictTypes:          append([]string(nil), v.ConflictTypes...),
		MissSchedule:           false,
		CourseStatus:           0,
		Width:                  0,
		TeacherList:            teacherList,
		StudentList:            studentList,
		ClassroomID:            strings.TrimSpace(v.ClassroomID),
		ClassroomName:          v.ClassroomName,
		CourseName:             v.LessonName,
		CourseType:             v.ClassType,
		ClassName:              v.TeachingClassName,
		LeaveList:              []any{},
		CourseTime:             minutes,
		CourseHour:             1,
		FinishType:             0,
	}
}

func (svc *Service) BatchUpdateTeachingSchedules(userID int64, dto model.TeachingScheduleBatchUpdateDTO) error {
	instID, operatorID, err := svc.resolveTeachingScheduleOperator(userID)
	if err != nil {
		return err
	}
	return svc.repo.BatchUpdateTeachingSchedules(context.Background(), instID, operatorID, dto)
}

func (svc *Service) GetTeachingScheduleBatchDetail(userID int64, query model.TeachingScheduleBatchDetailQueryDTO) (model.TeachingScheduleBatchDetailVO, error) {
	instID, _, err := svc.resolveTeachingScheduleOperator(userID)
	if err != nil {
		return model.TeachingScheduleBatchDetailVO{}, err
	}
	return svc.repo.GetTeachingScheduleBatchDetail(context.Background(), instID, query)
}

func (svc *Service) ReplaceTeachingScheduleBatch(userID int64, dto model.TeachingScheduleBatchReplaceDTO) (model.CreateOneToOneSchedulesResult, error) {
	instID, operatorID, err := svc.resolveTeachingScheduleOperator(userID)
	if err != nil {
		return model.CreateOneToOneSchedulesResult{}, err
	}
	return svc.repo.ReplaceTeachingScheduleBatch(context.Background(), instID, operatorID, dto)
}

func (svc *Service) CancelTeachingSchedules(userID int64, dto model.TeachingScheduleCancelDTO) (model.TeachingScheduleCancelResult, error) {
	instID, operatorID, err := svc.resolveTeachingScheduleOperator(userID)
	if err != nil {
		return model.TeachingScheduleCancelResult{}, err
	}
	return svc.repo.CancelTeachingSchedules(context.Background(), instID, operatorID, dto)
}

func (svc *Service) CancelTeachingScheduleScoped(userID int64, dto model.TeachingScheduleScopedCancelDTO) (model.TeachingScheduleCancelResult, error) {
	instID, operatorID, err := svc.resolveTeachingScheduleOperator(userID)
	if err != nil {
		return model.TeachingScheduleCancelResult{}, err
	}
	return svc.repo.CancelTeachingScheduleScoped(context.Background(), instID, operatorID, dto)
}

// CopyTeachingSchedulesWeek 将源周课表复制到目标周（按日历天对齐）；源 batch 在目标周生成新 batch_no，batch_size 与复制条数一致。
func (svc *Service) CopyTeachingSchedulesWeek(userID int64, dto model.TeachingScheduleCopyWeekDTO) (model.TeachingScheduleCopyWeekResult, error) {
	instID, operatorID, err := svc.resolveTeachingScheduleOperator(userID)
	if err != nil {
		return model.TeachingScheduleCopyWeekResult{}, err
	}
	return svc.repo.CopyTeachingSchedulesWeek(context.Background(), instID, operatorID, dto)
}

// CopyTeachingSchedulesDay 将源日期的老师课表复制到目标日期；若目标日期已有任意有效日程则整次失败。
func (svc *Service) CopyTeachingSchedulesDay(userID int64, dto model.TeachingScheduleCopyDayDTO) (model.TeachingScheduleCopyDayResult, error) {
	instID, operatorID, err := svc.resolveTeachingScheduleOperator(userID)
	if err != nil {
		return model.TeachingScheduleCopyDayResult{}, err
	}
	return svc.repo.CopyTeachingSchedulesDay(context.Background(), instID, operatorID, dto)
}

// ClearAllTeachingSchedules 清空当前登录用户所在机构的全部排课记录（软删）
func (svc *Service) ClearAllTeachingSchedules(userID int64) (deleted int64, err error) {
	instID, operatorID, err := svc.resolveTeachingScheduleOperator(userID)
	if err != nil {
		return 0, err
	}
	n, err := svc.repo.SoftDeleteAllTeachingSchedulesForInst(context.Background(), instID, operatorID)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (svc *Service) ClearWeekTeachingSchedules(userID int64, startDate, endDate time.Time) (deleted int64, err error) {
	instID, operatorID, err := svc.resolveTeachingScheduleOperator(userID)
	if err != nil {
		return 0, err
	}
	if endDate.Before(startDate) {
		return 0, errors.New("结束日期不能早于开始日期")
	}
	n, err := svc.repo.HardDeleteTeachingSchedulesInDateRange(context.Background(), instID, operatorID, startDate, endDate)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (svc *Service) resolveTeachingScheduleOperator(userID int64) (int64, int64, error) {
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
