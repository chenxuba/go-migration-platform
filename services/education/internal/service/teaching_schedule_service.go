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
	if err := svc.annotateTeachingScheduleConflictsForQuery(ctx, instID, query, schedules); err != nil {
		return nil, err
	}
	return schedules, nil
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
	if len(preferred) == 0 {
		return base
	}
	out := make([]int64, 0, len(base))
	seen := make(map[int64]struct{}, len(base))
	for _, id := range preferred {
		if id <= 0 {
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

type matrixExtraTeacher struct {
	id   int64
	name string
}

func buildTeacherOrderForMatrix(roster []model.InstUserScheduleRosterItem, schedules []model.TeachingScheduleVO) (order []int64, names map[int64]string) {
	names = make(map[int64]string)
	order = make([]int64, 0, len(roster)+8)
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

	var extras []matrixExtraTeacher
	for _, s := range schedules {
		tid, err := strconv.ParseInt(strings.TrimSpace(s.TeacherID), 10, 64)
		if err == nil && tid > 0 {
			if _, ok := seen[tid]; !ok {
				seen[tid] = struct{}{}
				nm := strings.TrimSpace(s.TeacherName)
				if nm == "" {
					nm = "-"
				}
				names[tid] = nm
				extras = append(extras, matrixExtraTeacher{id: tid, name: nm})
			}
		}
		for i, aid := range s.AssistantIDs {
			aid = strings.TrimSpace(aid)
			assistantID, err := strconv.ParseInt(aid, 10, 64)
			if err != nil || assistantID <= 0 {
				continue
			}
			if _, ok := seen[assistantID]; ok {
				continue
			}
			seen[assistantID] = struct{}{}
			nm := ""
			if i < len(s.AssistantNames) {
				nm = strings.TrimSpace(s.AssistantNames[i])
			}
			if nm == "" {
				nm = "-"
			}
			names[assistantID] = nm
			extras = append(extras, matrixExtraTeacher{id: assistantID, name: nm})
		}
	}
	sort.Slice(extras, func(i, j int) bool {
		if extras[i].name != extras[j].name {
			return extras[i].name < extras[j].name
		}
		return extras[i].id < extras[j].id
	})
	for _, e := range extras {
		order = append(order, e.id)
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
				if sameNonEmptyString(left.TeacherID, right.TeacherID) {
					appendScheduleConflictType(&schedules[leftIndex], "老师")
					appendScheduleConflictType(&schedules[rightIndex], "老师")
				}
				if sameNonEmptyString(left.ClassroomID, right.ClassroomID) {
					appendScheduleConflictType(&schedules[leftIndex], "教室")
					appendScheduleConflictType(&schedules[rightIndex], "教室")
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

func mapTeachingScheduleToLegacyVO(v model.TeachingScheduleVO, instID int64) model.TeachingScheduleInfoLegacyVO {
	id, _ := strconv.ParseInt(v.ID, 10, 64)
	tid, _ := strconv.ParseInt(strings.TrimSpace(v.TeacherID), 10, 64)
	sid, _ := strconv.ParseInt(strings.TrimSpace(v.StudentID), 10, 64)
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
	studentList := []model.ScheduleLegacyPersonVO{
		{Name: v.StudentName, ID: sid, Type: 1},
	}

	batchID, _ := strconv.ParseInt(strings.TrimSpace(v.BatchNo), 10, 64)

	var classIDPtr *int64
	if classIDVal > 0 {
		classIDPtr = &classIDVal
	}

	return model.TeachingScheduleInfoLegacyVO{
		ID:                id,
		InstID:            instID,
		BatchNo:           strings.TrimSpace(v.BatchNo),
		BatchID:           batchID,
		ModifyBatchID:     batchID,
		CourseID:          cid,
		ClassID:           classIDPtr,
		ScheduleDate:      v.LessonDate,
		ScheduleStartTime: v.StartAt.Format("15:04"),
		ScheduleEndTime:   v.EndAt.Format("15:04"),
		ScheduleStatus:    v.Status,
		Conflict:          v.Conflict,
		ConflictTypes:     append([]string(nil), v.ConflictTypes...),
		MissSchedule:      false,
		CourseStatus:      0,
		Width:             0,
		TeacherList:       teacherList,
		StudentList:       studentList,
		ClassroomID:       strings.TrimSpace(v.ClassroomID),
		ClassroomName:     v.ClassroomName,
		CourseName:        v.LessonName,
		CourseType:        v.ClassType,
		ClassName:         v.TeachingClassName,
		LeaveList:         []any{},
		CourseTime:        minutes,
		CourseHour:        1,
		FinishType:        0,
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

// CopyTeachingSchedulesWeek 将源周课表复制到目标周（按日历天对齐）；源 batch 在目标周生成新 batch_no，batch_size 与复制条数一致。
func (svc *Service) CopyTeachingSchedulesWeek(userID int64, dto model.TeachingScheduleCopyWeekDTO) (model.TeachingScheduleCopyWeekResult, error) {
	instID, operatorID, err := svc.resolveTeachingScheduleOperator(userID)
	if err != nil {
		return model.TeachingScheduleCopyWeekResult{}, err
	}
	return svc.repo.CopyTeachingSchedulesWeek(context.Background(), instID, operatorID, dto)
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
