package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

// CopyTeachingSchedulesWeek 将源周内课表复制到目标周；同一源 batch_no 的多条在目标周使用新的 batch_no，并保持相同的 batch_size（复制条数）。
func (repo *Repository) CopyTeachingSchedulesWeek(ctx context.Context, instID, operatorID int64, dto model.TeachingScheduleCopyWeekDTO) (model.TeachingScheduleCopyWeekResult, error) {
	var zero model.TeachingScheduleCopyWeekResult
	srcStart := strings.TrimSpace(dto.SourceStartDate)
	srcEnd := strings.TrimSpace(dto.SourceEndDate)
	tgtStart := strings.TrimSpace(dto.TargetStartDate)
	tgtEnd := strings.TrimSpace(dto.TargetEndDate)
	if srcStart == "" || srcEnd == "" || tgtStart == "" || tgtEnd == "" {
		return zero, errors.New("请填写源周与目标周的 startDate / endDate")
	}
	if srcStart == tgtStart && srcEnd == tgtEnd {
		return zero, errors.New("源周与目标周不能相同")
	}
	srcDays, err := expandScheduleCopyDateRange(srcStart, srcEnd)
	if err != nil {
		return zero, err
	}
	tgtDays, err := expandScheduleCopyDateRange(tgtStart, tgtEnd)
	if err != nil {
		return zero, err
	}
	if len(srcDays) == 0 || len(srcDays) != len(tgtDays) {
		return zero, errors.New("源周与目标周包含的天数须一致且不能为空")
	}
	dateMap := make(map[string]string, len(srcDays))
	for i := range srcDays {
		dateMap[srcDays[i]] = tgtDays[i]
	}

	query := model.TeachingScheduleListQueryDTO{
		StartDate: srcStart,
		EndDate:   srcEnd,
	}
	if scheduleTypes := normalizeNonEmptyStringList(dto.ScheduleTypes); len(scheduleTypes) > 0 {
		query.ScheduleTypeFilters = scheduleTypes
	} else {
		classType := model.TeachingClassTypeOneToOne
		if dto.ClassType != nil && *dto.ClassType > 0 {
			classType = *dto.ClassType
		}
		query.ClassType = &classType
	}
	sourceRows, err := repo.ListTeachingSchedules(ctx, instID, query)
	if err != nil {
		return zero, err
	}
	if len(sourceRows) == 0 {
		return model.TeachingScheduleCopyWeekResult{Created: 0}, nil
	}

	batchGroups, singles := groupSchedulesByBatchForCopy(sourceRows)
	plans, err := buildCopyPlans(batchGroups, singles, dateMap)
	if err != nil {
		return zero, err
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return zero, err
	}
	defer tx.Rollback()

	if err := repo.validateOneToOneClassesReadyForCopyTx(ctx, tx, instID, plans); err != nil {
		return zero, err
	}
	if err := repo.validateCopiedSchedulePlansAgainstDBTx(ctx, tx, instID, plans); err != nil {
		return zero, err
	}

	created, err := insertCopiedTeachingSchedulesTx(ctx, tx, instID, operatorID, plans)
	if err != nil {
		return zero, err
	}
	classIDs := make([]int64, 0, len(plans))
	for _, plan := range plans {
		classID, parseErr := strconv.ParseInt(strings.TrimSpace(plan.Source.TeachingClassID), 10, 64)
		if parseErr == nil && classID > 0 {
			classIDs = append(classIDs, classID)
		}
	}
	if err := repo.refreshTeachingClassScheduleCountsTx(ctx, tx, instID, operatorID, classIDs); err != nil {
		return zero, err
	}
	if err := tx.Commit(); err != nil {
		return zero, err
	}
	return model.TeachingScheduleCopyWeekResult{Created: created}, nil
}

// CopyTeachingSchedulesDay 将源日期的老师课表复制到目标日期；若目标日期已有任意有效日程则整次失败。
func (repo *Repository) CopyTeachingSchedulesDay(ctx context.Context, instID, operatorID int64, dto model.TeachingScheduleCopyDayDTO) (model.TeachingScheduleCopyDayResult, error) {
	var zero model.TeachingScheduleCopyDayResult
	sourceDate := strings.TrimSpace(dto.SourceDate)
	targetDate := strings.TrimSpace(dto.TargetDate)
	if sourceDate == "" || targetDate == "" {
		return zero, errors.New("请填写源日期和目标日期")
	}
	if sourceDate == targetDate {
		return zero, errors.New("源日期与目标日期不能相同")
	}
	if _, err := time.ParseInLocation("2006-01-02", sourceDate, time.Local); err != nil {
		return zero, errors.New("日期格式须为 YYYY-MM-DD")
	}
	if _, err := time.ParseInLocation("2006-01-02", targetDate, time.Local); err != nil {
		return zero, errors.New("日期格式须为 YYYY-MM-DD")
	}

	query := model.TeachingScheduleListQueryDTO{
		StartDate:           sourceDate,
		EndDate:             sourceDate,
		StudentID:           strings.TrimSpace(dto.StudentID),
		ScheduleTeacherIDs:  parsePositiveInt64Strings(dto.ScheduleTeacherIDs),
		ClassroomIDs:        parsePositiveInt64Strings(dto.ClassroomIDs),
		GroupClassIDs:       parsePositiveInt64Strings(dto.GroupClassIDs),
		OneToOneClassIDs:    parsePositiveInt64Strings(dto.OneToOneClassIDs),
		LessonIDs:           parsePositiveInt64Strings(dto.LessonIDs),
		ScheduleTypeFilters: normalizeNonEmptyStringList(dto.ScheduleTypes),
		CallStatusFilters:   normalizeNonEmptyStringList(dto.CallStatuses),
	}
	sourceRows, err := repo.ListTeachingSchedules(ctx, instID, query)
	if err != nil {
		return zero, err
	}
	if len(sourceRows) == 0 {
		if len(query.ScheduleTeacherIDs) > 0 {
			return zero, errors.New("源日期该老师暂无可复制日程")
		}
		return zero, errors.New("源日期暂无可复制日程")
	}

	batchGroups, singles := groupSchedulesByBatchForCopy(sourceRows)
	plans, err := buildCopyPlans(batchGroups, singles, map[string]string{sourceDate: targetDate})
	if err != nil {
		return zero, err
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return zero, err
	}
	defer tx.Rollback()

	if err := repo.ensureTeachingScheduleCopyTargetDateEmptyTx(ctx, tx, instID, targetDate); err != nil {
		return zero, err
	}
	if err := repo.validateOneToOneClassesReadyForCopyTx(ctx, tx, instID, plans); err != nil {
		return zero, err
	}
	if err := repo.validateCopiedSchedulePlansAgainstDBTx(ctx, tx, instID, plans); err != nil {
		return zero, err
	}

	created, err := insertCopiedTeachingSchedulesTx(ctx, tx, instID, operatorID, plans)
	if err != nil {
		return zero, err
	}
	classIDs := make([]int64, 0, len(plans))
	for _, plan := range plans {
		classID, parseErr := strconv.ParseInt(strings.TrimSpace(plan.Source.TeachingClassID), 10, 64)
		if parseErr == nil && classID > 0 {
			classIDs = append(classIDs, classID)
		}
	}
	if err := repo.refreshTeachingClassScheduleCountsTx(ctx, tx, instID, operatorID, classIDs); err != nil {
		return zero, err
	}
	if err := tx.Commit(); err != nil {
		return zero, err
	}
	return model.TeachingScheduleCopyDayResult{Created: created}, nil
}

type scheduleCopyPlan struct {
	Source           model.TeachingScheduleVO
	TargetLessonDate string
	StartAt          time.Time
	EndAt            time.Time
	NewBatchNo       string
	NewBatchSize     int
}

type copyOverlapTeacherKey struct {
	tid int64
	d   string
}

type copyOverlapClassroomKey struct {
	cid int64
	d   string
}

func parsePositiveInt64Strings(list []string) []int64 {
	if len(list) == 0 {
		return nil
	}
	out := make([]int64, 0, len(list))
	seen := make(map[int64]struct{}, len(list))
	for _, raw := range list {
		value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
		if err != nil || value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func normalizeNonEmptyStringList(list []string) []string {
	if len(list) == 0 {
		return nil
	}
	out := make([]string, 0, len(list))
	seen := make(map[string]struct{}, len(list))
	for _, raw := range list {
		value := strings.TrimSpace(raw)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func expandScheduleCopyDateRange(startStr, endStr string) ([]string, error) {
	start, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(startStr), time.Local)
	if err != nil {
		return nil, errors.New("日期格式须为 YYYY-MM-DD")
	}
	end, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(endStr), time.Local)
	if err != nil {
		return nil, errors.New("日期格式须为 YYYY-MM-DD")
	}
	if end.Before(start) {
		return nil, errors.New("结束日期不能早于开始日期")
	}
	var days []string
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		days = append(days, d.Format("2006-01-02"))
	}
	return days, nil
}

func groupSchedulesByBatchForCopy(rows []model.TeachingScheduleVO) (batchGroups map[string][]model.TeachingScheduleVO, singles []model.TeachingScheduleVO) {
	batchGroups = make(map[string][]model.TeachingScheduleVO)
	for _, s := range rows {
		bn := strings.TrimSpace(s.BatchNo)
		if bn == "" {
			singles = append(singles, s)
			continue
		}
		batchGroups[bn] = append(batchGroups[bn], s)
	}
	for k, g := range batchGroups {
		sort.Slice(g, func(i, j int) bool {
			if g[i].LessonDate != g[j].LessonDate {
				return g[i].LessonDate < g[j].LessonDate
			}
			return g[i].StartAt.Before(g[j].StartAt)
		})
		batchGroups[k] = g
	}
	return batchGroups, singles
}

func buildCopyPlans(
	batchGroups map[string][]model.TeachingScheduleVO,
	singles []model.TeachingScheduleVO,
	dateMap map[string]string,
) ([]scheduleCopyPlan, error) {
	nanoBase := time.Now().UnixNano()
	groupIdx := 0
	var plans []scheduleCopyPlan

	keys := make([]string, 0, len(batchGroups))
	for k := range batchGroups {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, bn := range keys {
		group := batchGroups[bn]
		newBatchNo, newSize := "", 1
		if len(group) > 1 {
			newBatchNo = fmt.Sprintf("BATCH-%d-%d", nanoBase, groupIdx)
			groupIdx++
			newSize = len(group)
		}
		for _, s := range group {
			tgtDay, ok := dateMap[s.LessonDate]
			if !ok {
				return nil, fmt.Errorf("日程 %s 不在源周区间内", s.LessonDate)
			}
			ns, ne := shiftScheduleTimesToDate(s.StartAt, s.EndAt, tgtDay)
			if !ne.After(ns) {
				return nil, errors.New("复制后的结束时间须晚于开始时间")
			}
			plans = append(plans, scheduleCopyPlan{
				Source:           s,
				TargetLessonDate: tgtDay,
				StartAt:          ns,
				EndAt:            ne,
				NewBatchNo:       newBatchNo,
				NewBatchSize:     newSize,
			})
		}
	}
	for _, s := range singles {
		tgtDay, ok := dateMap[s.LessonDate]
		if !ok {
			return nil, fmt.Errorf("日程 %s 不在源周区间内", s.LessonDate)
		}
		ns, ne := shiftScheduleTimesToDate(s.StartAt, s.EndAt, tgtDay)
		if !ne.After(ns) {
			return nil, errors.New("复制后的结束时间须晚于开始时间")
		}
		plans = append(plans, scheduleCopyPlan{
			Source:           s,
			TargetLessonDate: tgtDay,
			StartAt:          ns,
			EndAt:            ne,
			NewBatchNo:       "",
			NewBatchSize:     1,
		})
	}
	sort.Slice(plans, func(i, j int) bool {
		if plans[i].TargetLessonDate != plans[j].TargetLessonDate {
			return plans[i].TargetLessonDate < plans[j].TargetLessonDate
		}
		return plans[i].StartAt.Before(plans[j].StartAt)
	})
	return plans, nil
}

func shiftScheduleTimesToDate(startAt, endAt time.Time, targetDateISO string) (time.Time, time.Time) {
	d, err := time.ParseInLocation("2006-01-02", targetDateISO, time.Local)
	if err != nil {
		return startAt, endAt
	}
	y, m, day := d.Date()
	ns := time.Date(y, m, day, startAt.Hour(), startAt.Minute(), startAt.Second(), startAt.Nanosecond(), startAt.Location())
	ne := time.Date(y, m, day, endAt.Hour(), endAt.Minute(), endAt.Second(), endAt.Nanosecond(), endAt.Location())
	return ns, ne
}

func (repo *Repository) ensureTeachingScheduleCopyTargetDateEmptyTx(ctx context.Context, tx *sql.Tx, instID int64, targetDate string) error {
	var count int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(1)
		FROM teaching_schedule
		WHERE inst_id = ?
		  AND lesson_date = ?
		  AND del_flag = 0
		  AND status = ?
	`,
		instID,
		targetDate,
		model.TeachingScheduleStatusActive,
	).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("目标日期 %s 已有日程，请先清空后再复制", targetDate)
	}
	return nil
}

func (repo *Repository) validateOneToOneClassesReadyForCopyTx(ctx context.Context, tx *sql.Tx, instID int64, plans []scheduleCopyPlan) error {
	seen := make(map[int64]struct{})
	for _, p := range plans {
		if p.Source.ClassType != model.TeachingClassTypeOneToOne {
			continue
		}
		cid, perr := strconv.ParseInt(strings.TrimSpace(p.Source.TeachingClassID), 10, 64)
		if perr != nil || cid <= 0 {
			continue
		}
		if _, ok := seen[cid]; ok {
			continue
		}
		seen[cid] = struct{}{}
		if _, err := repo.GetOneToOneScheduleCreateContextTx(ctx, tx, instID, cid); err != nil {
			return err
		}
	}
	return nil
}

// validateCopiedSchedulePlansAgainstDBTx 校验新日程之间及与库中已有日程的教室/教师冲突。
func (repo *Repository) validateCopiedSchedulePlansAgainstDBTx(ctx context.Context, tx *sql.Tx, instID int64, plans []scheduleCopyPlan) error {
	byTeacher := make(map[copyOverlapTeacherKey][]normalizedScheduleSlot)
	byClassroom := make(map[copyOverlapClassroomKey][]normalizedScheduleSlot)

	for _, p := range plans {
		tid, _ := strconv.ParseInt(strings.TrimSpace(p.Source.TeacherID), 10, 64)
		cid, _ := strconv.ParseInt(strings.TrimSpace(p.Source.ClassroomID), 10, 64)
		day := p.TargetLessonDate
		slot := normalizedScheduleSlot{
			LessonDate: startOfDay(p.StartAt),
			StartAt:    p.StartAt,
			EndAt:      p.EndAt,
		}
		if tid > 0 {
			k := copyOverlapTeacherKey{tid, day}
			byTeacher[k] = append(byTeacher[k], slot)
		}
		if cid > 0 {
			k := copyOverlapClassroomKey{cid, day}
			byClassroom[k] = append(byClassroom[k], slot)
		}
	}
	if err := validateSlotSlicesNoInternalOverlap(byTeacher); err != nil {
		return err
	}
	if err := validateSlotSlicesNoInternalOverlapClassroom(byClassroom); err != nil {
		return err
	}
	for _, p := range plans {
		tid, _ := strconv.ParseInt(strings.TrimSpace(p.Source.TeacherID), 10, 64)
		cid, _ := strconv.ParseInt(strings.TrimSpace(p.Source.ClassroomID), 10, 64)
		slot := normalizedScheduleSlot{
			LessonDate: startOfDay(p.StartAt),
			StartAt:    p.StartAt,
			EndAt:      p.EndAt,
		}
		if tid > 0 {
			n, err := repo.countScheduleOverlapTx(ctx, tx, instID, "teacher_id", tid, slot, "", nil)
			if err != nil {
				return err
			}
			if n > 0 {
				return fmt.Errorf("老师在 %s %s-%s 已有日程冲突", p.TargetLessonDate, p.StartAt.Format("15:04"), p.EndAt.Format("15:04"))
			}
		}
		if cid > 0 {
			n, err := repo.countScheduleOverlapTx(ctx, tx, instID, "classroom_id", cid, slot, "", nil)
			if err != nil {
				return err
			}
			if n > 0 {
				return fmt.Errorf("教室在 %s %s-%s 已有日程冲突", p.TargetLessonDate, p.StartAt.Format("15:04"), p.EndAt.Format("15:04"))
			}
		}
	}
	return nil
}

func validateSlotSlicesNoInternalOverlap(m map[copyOverlapTeacherKey][]normalizedScheduleSlot) error {
	for _, slots := range m {
		for i := 0; i < len(slots); i++ {
			for j := i + 1; j < len(slots); j++ {
				if slots[i].LessonDate.Format("2006-01-02") != slots[j].LessonDate.Format("2006-01-02") {
					continue
				}
				if slots[i].StartAt.Before(slots[j].EndAt) && slots[i].EndAt.After(slots[j].StartAt) {
					return errors.New("复制后同一老师在同一天存在时间重叠的日程")
				}
			}
		}
	}
	return nil
}

func validateSlotSlicesNoInternalOverlapClassroom(m map[copyOverlapClassroomKey][]normalizedScheduleSlot) error {
	for _, slots := range m {
		for i := 0; i < len(slots); i++ {
			for j := i + 1; j < len(slots); j++ {
				if slots[i].LessonDate.Format("2006-01-02") != slots[j].LessonDate.Format("2006-01-02") {
					continue
				}
				if slots[i].StartAt.Before(slots[j].EndAt) && slots[i].EndAt.After(slots[j].StartAt) {
					return errors.New("复制后同一教室在同一天存在时间重叠的日程")
				}
			}
		}
	}
	return nil
}

func insertCopiedTeachingSchedulesTx(ctx context.Context, tx *sql.Tx, instID, operatorID int64, plans []scheduleCopyPlan) (int, error) {
	created := 0
	for _, p := range plans {
		s := p.Source
		classroomID, _ := strconv.ParseInt(strings.TrimSpace(s.ClassroomID), 10, 64)
		teachingClassID, err := strconv.ParseInt(strings.TrimSpace(s.TeachingClassID), 10, 64)
		if err != nil || teachingClassID <= 0 {
			return created, errors.New("无效的班级 ID")
		}
		studentID, err := strconv.ParseInt(strings.TrimSpace(s.StudentID), 10, 64)
		if err != nil {
			studentID = 0
		}
		lessonID, err := strconv.ParseInt(strings.TrimSpace(s.LessonID), 10, 64)
		if err != nil {
			lessonID = 0
		}
		teacherID, err := strconv.ParseInt(strings.TrimSpace(s.TeacherID), 10, 64)
		if err != nil || teacherID <= 0 {
			return created, errors.New("无效的教师")
		}
		aids := parseStringIDs(s.AssistantIDs)
		assistantIDsJSON, _ := json.Marshal(stringIDsFromInt64(aids))
		names := s.AssistantNames
		if names == nil {
			names = []string{}
		}
		assistantNamesJSON, _ := json.Marshal(names)

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO teaching_schedule (
				uuid, version, inst_id, class_type, teaching_class_id, teaching_class_name,
				student_id, student_name, lesson_id, lesson_name,
				teacher_id, teacher_name, assistant_ids_json, assistant_names_json,
				classroom_id, classroom_name, lesson_date, lesson_start_at, lesson_end_at,
				batch_no, batch_size, status, create_id, create_time, update_id, update_time, del_flag
			) VALUES (
				UUID(), 0, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), ?, NOW(), 0
			)
		`,
			instID,
			s.ClassType,
			teachingClassID,
			s.TeachingClassName,
			studentID,
			s.StudentName,
			lessonID,
			s.LessonName,
			teacherID,
			s.TeacherName,
			nullJSONBytes(assistantIDsJSON),
			nullJSONBytes(assistantNamesJSON),
			classroomID,
			s.ClassroomName,
			p.TargetLessonDate,
			p.StartAt,
			p.EndAt,
			p.NewBatchNo,
			p.NewBatchSize,
			model.TeachingScheduleStatusActive,
			operatorID,
			operatorID,
		); err != nil {
			return created, err
		}
		created++
	}
	return created, nil
}
