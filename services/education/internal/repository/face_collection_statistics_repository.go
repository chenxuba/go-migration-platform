package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

type faceAttendanceTodaySchedulePair struct {
	ScheduleID    int64
	StudentID     int64
	StudentName   string
	ClassTime     string
	ScheduleName  string
	LessonStartAt time.Time
	LessonEndAt   time.Time
}

type faceAttendanceTodayStudentMeta struct {
	StudentName string
	Mobile      string
	AvatarURL   string
	StudentSex  int
	IsCollect   bool
}

type faceAttendanceTodaySessionSnapshot struct {
	SessionID     int64
	SessionStatus int
	SignInTime    *time.Time
	SignInImage   string
	SignOutTime   *time.Time
	SignOutImage  string
}

type faceAttendanceTodayRecordSnapshot struct {
	Status         int
	IsAutoRollCall bool
	RecordTime     *time.Time
}

type faceAttendanceTodayDashboardData struct {
	Pending         []model.FaceAttendanceTodayDetailItem
	Success         []model.FaceAttendanceTodayDetailItem
	SuccessUnrolled []model.FaceAttendanceTodayDetailItem
}

type faceAttendanceTodaySessionAggregate struct {
	Item          model.FaceAttendanceTodayDetailItem
	ClassTimes    []string
	ScheduleNames []string
	SeenSchedule  map[string]struct{}
}

type faceAttendanceTodayPairKey struct {
	ScheduleID int64
	StudentID  int64
}

func (repo *Repository) GetFaceAttendanceTodayStatistics(ctx context.Context, instID int64) (model.FaceAttendanceTodayStatistics, error) {
	data, err := repo.buildFaceAttendanceTodayDashboardData(ctx, instID)
	if err != nil {
		return model.FaceAttendanceTodayStatistics{}, err
	}
	return model.FaceAttendanceTodayStatistics{
		PendingCount:         len(data.Pending),
		SuccessCount:         len(data.Success),
		SuccessUnrolledCount: len(data.SuccessUnrolled),
	}, nil
}

func (repo *Repository) PageFaceAttendanceTodaySuccessRecords(ctx context.Context, instID int64, query model.FaceAttendanceTodaySuccessRecordQueryDTO) (model.PageResult[model.FaceAttendanceRecordItem], error) {
	current := query.PageRequestModel.PageIndex
	size := query.PageRequestModel.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 20
	}
	offset := (current - 1) * size
	attendanceDate := time.Now().In(time.Local).Format("2006-01-02")
	searchKey := strings.TrimSpace(query.QueryModel.SearchKey)

	total, err := repo.countFaceAttendanceTodaySuccessSessions(ctx, instID, attendanceDate, searchKey)
	if err != nil {
		return model.PageResult[model.FaceAttendanceRecordItem]{}, err
	}
	if total == 0 {
		return model.PageResult[model.FaceAttendanceRecordItem]{
			Items:   []model.FaceAttendanceRecordItem{},
			Total:   0,
			Current: current,
			Size:    size,
		}, nil
	}

	args := []any{instID, attendanceDate, faceAttendanceRollCallTaskStatusSuccess}
	whereParts := []string{
		"fas.inst_id = ?",
		"IFNULL(fas.del_flag, 0) = 0",
		"fas.attendance_date = ?",
		"fas.sign_in_time IS NOT NULL",
		`EXISTS (
			SELECT 1
			FROM inst_student_face_roll_call_task task
			WHERE task.inst_id = fas.inst_id
			  AND IFNULL(task.del_flag, 0) = 0
			  AND task.attendance_session_id = fas.id
			  AND task.status = ?
		)`,
	}
	if searchKey != "" {
		likeText := "%" + searchKey + "%"
		whereParts = append(whereParts, "(IFNULL(stu.stu_name, IFNULL(fas.student_name, '')) LIKE ? OR IFNULL(stu.mobile, '') LIKE ?)")
		args = append(args, likeText, likeText)
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			fas.id,
			fas.student_id,
			DATE_FORMAT(fas.attendance_date, '%Y-%m-%d'),
			IFNULL(fas.status, 0),
			IFNULL(fas.sign_in_image, ''),
			IFNULL(fas.sign_out_image, ''),
			fas.sign_in_time,
			fas.sign_out_time,
			IFNULL(stu.stu_name, IFNULL(fas.student_name, '')),
			IFNULL(stu.mobile, ''),
			IFNULL(stu.avatar_url, ''),
			IFNULL(stu.stu_sex, 0),
			IFNULL(stu.is_collect, 0)
		FROM inst_student_face_attendance_session fas
		LEFT JOIN inst_student stu
			ON stu.id = fas.student_id
		   AND stu.inst_id = fas.inst_id
		   AND IFNULL(stu.del_flag, 0) = 0
		WHERE `+strings.Join(whereParts, " AND ")+`
		ORDER BY fas.sign_in_time DESC, fas.id DESC
		LIMIT ? OFFSET ?
	`, append(args, size, offset)...)
	if err != nil {
		return model.PageResult[model.FaceAttendanceRecordItem]{}, err
	}
	defer rows.Close()

	items := make([]model.FaceAttendanceRecordItem, 0, size)
	sessionIDs := make([]int64, 0, size)
	for rows.Next() {
		var (
			item        model.FaceAttendanceRecordItem
			signInTime  sql.NullTime
			signOutTime sql.NullTime
			isCollect   int
		)
		if err := rows.Scan(
			&item.SessionID,
			&item.StudentID,
			&item.AttendanceDate,
			&item.SessionStatus,
			&item.SignInImage,
			&item.SignOutImage,
			&signInTime,
			&signOutTime,
			&item.StudentName,
			&item.StudentMobile,
			&item.AvatarURL,
			&item.StudentSex,
			&isCollect,
		); err != nil {
			return model.PageResult[model.FaceAttendanceRecordItem]{}, err
		}
		item.ID = fmt.Sprintf("%d", item.SessionID)
		item.AttendanceType = "人脸考勤"
		item.Action = model.FaceAttendanceSessionActionSignIn
		item.ActionLabel = faceAttendanceActionLabel(item.Action)
		item.StudentMobile = maskStudentMobile(item.StudentMobile)
		item.IsCollect = isCollect != 0
		if signInTime.Valid {
			t := signInTime.Time
			item.AttendanceTime = &t
			item.ActionTime = &t
		}
		if signOutTime.Valid {
			t := signOutTime.Time
			item.SignOutTime = &t
		}
		items = append(items, item)
		sessionIDs = append(sessionIDs, item.SessionID)
	}
	if err := rows.Err(); err != nil {
		return model.PageResult[model.FaceAttendanceRecordItem]{}, err
	}

	sessionSummaryMap, err := repo.loadFaceAttendanceRecordSessionSummaries(ctx, instID, sessionIDs)
	if err != nil {
		return model.PageResult[model.FaceAttendanceRecordItem]{}, err
	}
	taskSummaryMap, err := repo.loadFaceAttendanceRecordTaskSummaries(ctx, instID, sessionIDs)
	if err != nil {
		return model.PageResult[model.FaceAttendanceRecordItem]{}, err
	}

	for index := range items {
		item := &items[index]
		if summary, ok := sessionSummaryMap[item.SessionID]; ok {
			item.HasSchedule = summary.HasSchedule
			item.ClassTimes = summary.ClassTimes
			item.RelatedSchedules = summary.RelatedSchedules
			item.RelatedScheduleItems = summary.RelatedScheduleItems
		}
		item.Prompt = buildFaceAttendanceRecordPrompt(model.FaceAttendanceSessionActionSignIn, item.HasSchedule, taskSummaryMap[item.SessionID])
	}

	return model.PageResult[model.FaceAttendanceRecordItem]{
		Items:   items,
		Total:   total,
		Current: current,
		Size:    size,
	}, nil
}

func (repo *Repository) PageFaceAttendanceTodayDetails(ctx context.Context, instID int64, query model.FaceAttendanceTodayDetailQueryDTO) (model.PageResult[model.FaceAttendanceTodayDetailItem], error) {
	data, err := repo.buildFaceAttendanceTodayDashboardData(ctx, instID)
	if err != nil {
		return model.PageResult[model.FaceAttendanceTodayDetailItem]{}, err
	}

	var source []model.FaceAttendanceTodayDetailItem
	switch strings.TrimSpace(query.QueryModel.Type) {
	case model.FaceAttendanceTodayStatisticTypeSuccess:
		source = data.Success
	case model.FaceAttendanceTodayStatisticTypeSuccessUnrolled:
		source = data.SuccessUnrolled
	default:
		source = data.Pending
	}

	filtered := filterFaceAttendanceTodayDetailItems(source, query.QueryModel.SearchKey)

	current := query.PageRequestModel.PageIndex
	size := query.PageRequestModel.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 20
	}
	offset := (current - 1) * size
	if offset >= len(filtered) {
		return model.PageResult[model.FaceAttendanceTodayDetailItem]{
			Items:   []model.FaceAttendanceTodayDetailItem{},
			Total:   len(filtered),
			Current: current,
			Size:    size,
		}, nil
	}
	end := offset + size
	if end > len(filtered) {
		end = len(filtered)
	}
	items := make([]model.FaceAttendanceTodayDetailItem, end-offset)
	copy(items, filtered[offset:end])
	return model.PageResult[model.FaceAttendanceTodayDetailItem]{
		Items:   items,
		Total:   len(filtered),
		Current: current,
		Size:    size,
	}, nil
}

func (repo *Repository) countFaceAttendanceTodaySuccessSessions(ctx context.Context, instID int64, attendanceDate, searchKey string) (int, error) {
	args := []any{instID, attendanceDate, faceAttendanceRollCallTaskStatusSuccess}
	whereParts := []string{
		"fas.inst_id = ?",
		"IFNULL(fas.del_flag, 0) = 0",
		"fas.attendance_date = ?",
		"fas.sign_in_time IS NOT NULL",
		`EXISTS (
			SELECT 1
			FROM inst_student_face_roll_call_task task
			WHERE task.inst_id = fas.inst_id
			  AND IFNULL(task.del_flag, 0) = 0
			  AND task.attendance_session_id = fas.id
			  AND task.status = ?
		)`,
	}
	text := strings.TrimSpace(searchKey)
	if text != "" {
		likeText := "%" + text + "%"
		whereParts = append(whereParts, "(IFNULL(stu.stu_name, IFNULL(fas.student_name, '')) LIKE ? OR IFNULL(stu.mobile, '') LIKE ?)")
		args = append(args, likeText, likeText)
	}

	var total int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM inst_student_face_attendance_session fas
		LEFT JOIN inst_student stu
			ON stu.id = fas.student_id
		   AND stu.inst_id = fas.inst_id
		   AND IFNULL(stu.del_flag, 0) = 0
		WHERE `+strings.Join(whereParts, " AND "), args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (repo *Repository) buildFaceAttendanceTodayDashboardData(ctx context.Context, instID int64) (faceAttendanceTodayDashboardData, error) {
	now := time.Now().In(time.Local)
	attendanceDate := now.Format("2006-01-02")

	pairs, err := repo.listFaceAttendanceTodaySchedulePairs(ctx, instID, attendanceDate)
	if err != nil {
		return faceAttendanceTodayDashboardData{}, err
	}
	if len(pairs) == 0 {
		return faceAttendanceTodayDashboardData{
			Pending:         []model.FaceAttendanceTodayDetailItem{},
			Success:         []model.FaceAttendanceTodayDetailItem{},
			SuccessUnrolled: []model.FaceAttendanceTodayDetailItem{},
		}, nil
	}

	studentIDs := make([]int64, 0, len(pairs))
	scheduleIDs := make([]int64, 0, len(pairs))
	seenStudents := make(map[int64]struct{}, len(pairs))
	seenSchedules := make(map[int64]struct{}, len(pairs))
	for _, pair := range pairs {
		if _, ok := seenStudents[pair.StudentID]; !ok {
			seenStudents[pair.StudentID] = struct{}{}
			studentIDs = append(studentIDs, pair.StudentID)
		}
		if _, ok := seenSchedules[pair.ScheduleID]; !ok {
			seenSchedules[pair.ScheduleID] = struct{}{}
			scheduleIDs = append(scheduleIDs, pair.ScheduleID)
		}
	}

	studentMetaMap, err := repo.loadFaceAttendanceTodayStudentMetaMap(ctx, instID, studentIDs)
	if err != nil {
		return faceAttendanceTodayDashboardData{}, err
	}
	sessionByStudentID, err := repo.loadFaceAttendanceTodaySessionMap(ctx, instID, attendanceDate, studentIDs)
	if err != nil {
		return faceAttendanceTodayDashboardData{}, err
	}
	recordByPairKey, err := repo.loadFaceAttendanceTodayRecordMap(ctx, instID, scheduleIDs)
	if err != nil {
		return faceAttendanceTodayDashboardData{}, err
	}

	data := faceAttendanceTodayDashboardData{
		Pending:         make([]model.FaceAttendanceTodayDetailItem, 0),
		Success:         make([]model.FaceAttendanceTodayDetailItem, 0),
		SuccessUnrolled: make([]model.FaceAttendanceTodayDetailItem, 0),
	}
	successBySession := make(map[int64]*faceAttendanceTodaySessionAggregate)
	successUnrolledBySession := make(map[int64]*faceAttendanceTodaySessionAggregate)
	for _, pair := range pairs {
		key := faceAttendanceTodayPairKey{ScheduleID: pair.ScheduleID, StudentID: pair.StudentID}
		record, hasRecord := recordByPairKey[key]
		if hasRecord && (record.Status == 2 || record.Status == 3) {
			continue
		}

		meta := studentMetaMap[pair.StudentID]
		if strings.TrimSpace(meta.StudentName) == "" {
			meta.StudentName = pair.StudentName
		}
		session, hasSession := sessionByStudentID[pair.StudentID]
		eligibleFaceSession := faceAttendanceTodayPairMatchesSession(pair, session, hasSession)
		canManualRollCall := !pair.LessonEndAt.IsZero() && !pair.LessonEndAt.Add(faceAttendanceAutoRollCallDelay).After(now)

		baseItem := model.FaceAttendanceTodayDetailItem{
			ID:                fmt.Sprintf("%d-%d", pair.ScheduleID, pair.StudentID),
			ScheduleID:        fmt.Sprintf("%d", pair.ScheduleID),
			SessionID:         session.SessionID,
			StudentID:         pair.StudentID,
			StudentName:       firstNonEmptyString(strings.TrimSpace(meta.StudentName), strings.TrimSpace(pair.StudentName), "该学员"),
			StudentMobile:     maskStudentMobile(meta.Mobile),
			AvatarURL:         strings.TrimSpace(meta.AvatarURL),
			StudentSex:        meta.StudentSex,
			IsCollect:         meta.IsCollect,
			HasFaceSession:    hasSession,
			HasSchedule:       true,
			SessionStatus:     session.SessionStatus,
			SignInTime:        session.SignInTime,
			SignOutTime:       session.SignOutTime,
			SignInImage:       strings.TrimSpace(session.SignInImage),
			SignOutImage:      strings.TrimSpace(session.SignOutImage),
			ClassTime:         pair.ClassTime,
			ScheduleName:      pair.ScheduleName,
			CanManualRollCall: canManualRollCall,
		}
		if !pair.LessonStartAt.IsZero() {
			startAt := pair.LessonStartAt
			baseItem.LessonStartAt = &startAt
		}
		if !pair.LessonEndAt.IsZero() {
			endAt := pair.LessonEndAt
			baseItem.LessonEndAt = &endAt
		}

		relatedScheduleItem := model.FaceAttendanceRelatedScheduleItem{
			ScheduleID:   fmt.Sprintf("%d", pair.ScheduleID),
			ClassTime:    pair.ClassTime,
			ScheduleName: pair.ScheduleName,
		}

		switch {
		case hasRecord && record.Status == 1 && eligibleFaceSession:
			relatedScheduleItem.RollCallStatus = "已点名"
			aggregate := ensureFaceAttendanceTodaySessionAggregate(successBySession, session.SessionID, baseItem, model.FaceAttendanceTodayStatisticTypeSuccess)
			aggregate.Item.AttendanceTime = session.SignInTime
			aggregate.Item.AttendanceType = "人脸考勤"
			aggregate.Item.Prompt = "点名到课"
			appendFaceAttendanceTodayAggregateSchedule(aggregate, relatedScheduleItem, pair.LessonStartAt, pair.LessonEndAt)
		case hasRecord && record.Status == 1:
			// 纯手动点名不属于人脸考勤看板口径，避免和刷脸成功记录混在一起。
			continue
		case eligibleFaceSession:
			relatedScheduleItem.RollCallStatus = "未点名"
			aggregate := ensureFaceAttendanceTodaySessionAggregate(successUnrolledBySession, session.SessionID, baseItem, model.FaceAttendanceTodayStatisticTypeSuccessUnrolled)
			aggregate.Item.AttendanceTime = session.SignInTime
			aggregate.Item.AttendanceType = "人脸考勤"
			if canManualRollCall {
				aggregate.Item.Prompt = "待手动点名"
				aggregate.Item.CanManualRollCall = true
			} else {
				if !aggregate.Item.CanManualRollCall {
					aggregate.Item.Prompt = "待自动点名"
				}
			}
			appendFaceAttendanceTodayAggregateSchedule(aggregate, relatedScheduleItem, pair.LessonStartAt, pair.LessonEndAt)
		default:
			baseItem.Type = model.FaceAttendanceTodayStatisticTypePending
			if !meta.IsCollect {
				baseItem.Prompt = "待采集"
			} else if !pair.LessonStartAt.IsZero() && pair.LessonStartAt.After(now) {
				baseItem.Prompt = "待开课"
			} else {
				baseItem.Prompt = "待刷脸"
			}
			data.Pending = append(data.Pending, baseItem)
		}
	}

	data.Success = finalizeFaceAttendanceTodaySessionAggregates(successBySession)
	data.SuccessUnrolled = finalizeFaceAttendanceTodaySessionAggregates(successUnrolledBySession)

	sort.SliceStable(data.Pending, func(i, j int) bool {
		return compareFaceAttendanceTodayItems(data.Pending[i], data.Pending[j], true)
	})
	sort.SliceStable(data.Success, func(i, j int) bool {
		return compareFaceAttendanceTodayItems(data.Success[i], data.Success[j], false)
	})
	sort.SliceStable(data.SuccessUnrolled, func(i, j int) bool {
		return compareFaceAttendanceTodayItems(data.SuccessUnrolled[i], data.SuccessUnrolled[j], false)
	})
	return data, nil
}

func (repo *Repository) listFaceAttendanceTodaySchedulePairs(ctx context.Context, instID int64, attendanceDate string) ([]faceAttendanceTodaySchedulePair, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			IFNULL(ts.id, 0),
			IFNULL(ts.class_type, 0),
			IFNULL(ts.teaching_class_id, 0),
			IFNULL(ts.student_id, 0),
			IFNULL(ts.teaching_class_name, ''),
			IFNULL(ts.student_name, ''),
			IFNULL(ts.lesson_name, ''),
			ts.lesson_start_at,
			ts.lesson_end_at
		FROM teaching_schedule ts
		WHERE ts.inst_id = ?
		  AND IFNULL(ts.del_flag, 0) = 0
		  AND ts.status = ?
		  AND ts.lesson_date = ?
		ORDER BY ts.lesson_start_at ASC, ts.id ASC
	`, instID, model.TeachingScheduleStatusActive, attendanceDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	schedules := make([]faceAttendanceScheduleMatch, 0)
	groupMetas := make([]effectiveGroupClassScheduleMeta, 0)
	for rows.Next() {
		var item faceAttendanceScheduleMatch
		if err := rows.Scan(
			&item.ScheduleID,
			&item.ClassType,
			&item.ClassID,
			&item.StudentID,
			&item.TeachingClassName,
			&item.StudentName,
			&item.LessonName,
			&item.LessonStartAt,
			&item.LessonEndAt,
		); err != nil {
			return nil, err
		}
		schedules = append(schedules, item)
		if item.ClassType == model.TeachingClassTypeNormal && item.ClassID > 0 {
			groupMetas = append(groupMetas, effectiveGroupClassScheduleMeta{
				ScheduleID: item.ScheduleID,
				ClassID:    item.ClassID,
				StartAt:    item.LessonStartAt,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(schedules) == 0 {
		return []faceAttendanceTodaySchedulePair{}, nil
	}

	rosterByScheduleID := map[int64]groupClassScheduleRoster{}
	if len(groupMetas) > 0 {
		rosterByScheduleID, err = repo.loadEffectiveGroupClassScheduleRosterMap(ctx, repo.db, instID, groupMetas)
		if err != nil {
			return nil, err
		}
	}

	result := make([]faceAttendanceTodaySchedulePair, 0, len(schedules))
	seen := make(map[faceAttendanceTodayPairKey]struct{}, len(schedules))
	for _, schedule := range schedules {
		appendPair := func(studentID int64, studentName string) {
			if studentID <= 0 {
				return
			}
			key := faceAttendanceTodayPairKey{ScheduleID: schedule.ScheduleID, StudentID: studentID}
			if _, ok := seen[key]; ok {
				return
			}
			seen[key] = struct{}{}
			result = append(result, faceAttendanceTodaySchedulePair{
				ScheduleID:    schedule.ScheduleID,
				StudentID:     studentID,
				StudentName:   firstNonEmptyString(strings.TrimSpace(studentName), strings.TrimSpace(schedule.StudentName), "该学员"),
				ClassTime:     formatFaceAttendanceScheduleClassTime(schedule),
				ScheduleName:  formatFaceAttendanceScheduleName(schedule),
				LessonStartAt: schedule.LessonStartAt,
				LessonEndAt:   schedule.LessonEndAt,
			})
		}

		switch {
		case schedule.ClassType == model.TeachingClassTypeNormal && schedule.ClassID > 0:
			for _, student := range rosterByScheduleID[schedule.ScheduleID].Active {
				appendPair(student.StudentID, student.StudentName)
			}
		case schedule.ClassType == model.TeachingClassTypeOneToOne && schedule.StudentID > 0:
			appendPair(schedule.StudentID, schedule.StudentName)
		}
	}
	return result, nil
}

func (repo *Repository) loadFaceAttendanceTodayStudentMetaMap(ctx context.Context, instID int64, studentIDs []int64) (map[int64]faceAttendanceTodayStudentMeta, error) {
	studentIDs = uniquePositiveInt64s(studentIDs)
	result := make(map[int64]faceAttendanceTodayStudentMeta, len(studentIDs))
	if len(studentIDs) == 0 {
		return result, nil
	}
	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			IFNULL(id, 0),
			IFNULL(stu_name, ''),
			IFNULL(mobile, ''),
			IFNULL(avatar_url, ''),
			IFNULL(stu_sex, 0),
			IFNULL(is_collect, 0)
		FROM inst_student
		WHERE inst_id = ?
		  AND del_flag = 0
		  AND id IN (`+sqlPlaceholders(len(studentIDs))+`)
	`, append([]any{instID}, int64SliceToAny(studentIDs)...)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			studentID int64
			item      faceAttendanceTodayStudentMeta
			isCollect int
		)
		if err := rows.Scan(&studentID, &item.StudentName, &item.Mobile, &item.AvatarURL, &item.StudentSex, &isCollect); err != nil {
			return nil, err
		}
		item.IsCollect = isCollect != 0
		result[studentID] = item
	}
	return result, rows.Err()
}

func (repo *Repository) loadFaceAttendanceTodaySessionMap(ctx context.Context, instID int64, attendanceDate string, studentIDs []int64) (map[int64]faceAttendanceTodaySessionSnapshot, error) {
	studentIDs = uniquePositiveInt64s(studentIDs)
	result := make(map[int64]faceAttendanceTodaySessionSnapshot, len(studentIDs))
	if len(studentIDs) == 0 {
		return result, nil
	}
	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			IFNULL(id, 0),
			IFNULL(student_id, 0),
			IFNULL(status, 0),
			sign_in_time,
			IFNULL(sign_in_image, ''),
			sign_out_time,
			IFNULL(sign_out_image, '')
		FROM inst_student_face_attendance_session
		WHERE inst_id = ?
		  AND attendance_date = ?
		  AND del_flag = 0
		  AND sign_in_time IS NOT NULL
		  AND student_id IN (`+sqlPlaceholders(len(studentIDs))+`)
	`, append([]any{instID, attendanceDate}, int64SliceToAny(studentIDs)...)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			item        faceAttendanceTodaySessionSnapshot
			studentID   int64
			signInTime  sql.NullTime
			signOutTime sql.NullTime
		)
		if err := rows.Scan(&item.SessionID, &studentID, &item.SessionStatus, &signInTime, &item.SignInImage, &signOutTime, &item.SignOutImage); err != nil {
			return nil, err
		}
		if signInTime.Valid {
			t := signInTime.Time
			item.SignInTime = &t
		}
		if signOutTime.Valid {
			t := signOutTime.Time
			item.SignOutTime = &t
		}
		result[studentID] = item
	}
	return result, rows.Err()
}

func (repo *Repository) loadFaceAttendanceTodayRecordMap(ctx context.Context, instID int64, scheduleIDs []int64) (map[faceAttendanceTodayPairKey]faceAttendanceTodayRecordSnapshot, error) {
	scheduleIDs = uniquePositiveInt64s(scheduleIDs)
	result := make(map[faceAttendanceTodayPairKey]faceAttendanceTodayRecordSnapshot, len(scheduleIDs))
	if len(scheduleIDs) == 0 {
		return result, nil
	}
	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			IFNULL(teaching_schedule_id, 0),
			IFNULL(student_id, 0),
			IFNULL(status, 0),
			IFNULL(is_auto_roll_call, 0),
			teaching_record_created_time
		FROM student_teaching_record
		WHERE inst_id = ?
		  AND del_flag = 0
		  AND teaching_schedule_id IN (`+sqlPlaceholders(len(scheduleIDs))+`)
		ORDER BY id ASC
	`, append([]any{instID}, int64SliceToAny(scheduleIDs)...)...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return result, nil
		}
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			scheduleID   int64
			studentID    int64
			item         faceAttendanceTodayRecordSnapshot
			autoRollCall int
			recordTime   sql.NullTime
		)
		if err := rows.Scan(&scheduleID, &studentID, &item.Status, &autoRollCall, &recordTime); err != nil {
			return nil, err
		}
		item.IsAutoRollCall = autoRollCall != 0
		if recordTime.Valid {
			t := recordTime.Time
			item.RecordTime = &t
		}
		if scheduleID > 0 && studentID > 0 {
			result[faceAttendanceTodayPairKey{ScheduleID: scheduleID, StudentID: studentID}] = item
		}
	}
	return result, rows.Err()
}

func ensureFaceAttendanceTodaySessionAggregate(
	target map[int64]*faceAttendanceTodaySessionAggregate,
	sessionID int64,
	baseItem model.FaceAttendanceTodayDetailItem,
	itemType string,
) *faceAttendanceTodaySessionAggregate {
	if aggregate, ok := target[sessionID]; ok {
		return aggregate
	}

	baseItem.ID = fmt.Sprintf("%d", sessionID)
	baseItem.Type = itemType
	baseItem.ScheduleID = ""
	baseItem.ClassTime = ""
	baseItem.ScheduleName = ""
	baseItem.RelatedScheduleItems = nil

	aggregate := &faceAttendanceTodaySessionAggregate{
		Item:          baseItem,
		ClassTimes:    make([]string, 0, 4),
		ScheduleNames: make([]string, 0, 4),
		SeenSchedule:  make(map[string]struct{}),
	}
	target[sessionID] = aggregate
	return aggregate
}

func appendFaceAttendanceTodayAggregateSchedule(
	aggregate *faceAttendanceTodaySessionAggregate,
	scheduleItem model.FaceAttendanceRelatedScheduleItem,
	lessonStartAt time.Time,
	lessonEndAt time.Time,
) {
	if aggregate == nil {
		return
	}

	key := strings.TrimSpace(scheduleItem.ScheduleID)
	if key == "" {
		key = fmt.Sprintf("%s|%s|%s", scheduleItem.ClassTime, scheduleItem.ScheduleName, scheduleItem.RollCallStatus)
	}
	if _, exists := aggregate.SeenSchedule[key]; exists {
		return
	}
	aggregate.SeenSchedule[key] = struct{}{}

	aggregate.Item.RelatedScheduleItems = append(aggregate.Item.RelatedScheduleItems, scheduleItem)
	if text := strings.TrimSpace(scheduleItem.ClassTime); text != "" {
		aggregate.ClassTimes = append(aggregate.ClassTimes, text)
	}
	if text := strings.TrimSpace(scheduleItem.ScheduleName); text != "" {
		aggregate.ScheduleNames = append(aggregate.ScheduleNames, text)
	}
	if aggregate.Item.ScheduleID == "" {
		aggregate.Item.ScheduleID = scheduleItem.ScheduleID
	}
	if !lessonStartAt.IsZero() && (aggregate.Item.LessonStartAt == nil || lessonStartAt.Before(*aggregate.Item.LessonStartAt)) {
		startAt := lessonStartAt
		aggregate.Item.LessonStartAt = &startAt
	}
	if !lessonEndAt.IsZero() && (aggregate.Item.LessonEndAt == nil || lessonEndAt.After(*aggregate.Item.LessonEndAt)) {
		endAt := lessonEndAt
		aggregate.Item.LessonEndAt = &endAt
	}
}

func finalizeFaceAttendanceTodaySessionAggregates(source map[int64]*faceAttendanceTodaySessionAggregate) []model.FaceAttendanceTodayDetailItem {
	if len(source) == 0 {
		return []model.FaceAttendanceTodayDetailItem{}
	}

	items := make([]model.FaceAttendanceTodayDetailItem, 0, len(source))
	for _, aggregate := range source {
		if aggregate == nil {
			continue
		}
		item := aggregate.Item
		item.ClassTime = strings.Join(aggregate.ClassTimes, "\n")
		item.ScheduleName = strings.Join(aggregate.ScheduleNames, "\n")
		if len(item.RelatedScheduleItems) > 1 {
			item.ScheduleID = ""
		}
		items = append(items, item)
	}
	return items
}

func faceAttendanceTodayPairMatchesSession(pair faceAttendanceTodaySchedulePair, session faceAttendanceTodaySessionSnapshot, hasSession bool) bool {
	if !hasSession || session.SignInTime == nil {
		return false
	}
	if pair.LessonEndAt.IsZero() {
		return true
	}
	return !pair.LessonEndAt.Before(*session.SignInTime)
}

func filterFaceAttendanceTodayDetailItems(items []model.FaceAttendanceTodayDetailItem, searchKey string) []model.FaceAttendanceTodayDetailItem {
	text := strings.TrimSpace(searchKey)
	if text == "" {
		result := make([]model.FaceAttendanceTodayDetailItem, len(items))
		copy(result, items)
		return result
	}
	result := make([]model.FaceAttendanceTodayDetailItem, 0, len(items))
	for _, item := range items {
		if strings.Contains(strings.TrimSpace(item.StudentName), text) || strings.Contains(strings.ReplaceAll(strings.TrimSpace(item.StudentMobile), "*", ""), text) {
			result = append(result, item)
		}
	}
	return result
}

func compareFaceAttendanceTodayItems(left, right model.FaceAttendanceTodayDetailItem, ascending bool) bool {
	leftTime := faceAttendanceTodaySortTime(left, ascending)
	rightTime := faceAttendanceTodaySortTime(right, ascending)
	if !leftTime.Equal(rightTime) {
		if ascending {
			return leftTime.Before(rightTime)
		}
		return leftTime.After(rightTime)
	}
	if strings.TrimSpace(left.StudentName) != strings.TrimSpace(right.StudentName) {
		return strings.TrimSpace(left.StudentName) < strings.TrimSpace(right.StudentName)
	}
	return strings.TrimSpace(left.ID) < strings.TrimSpace(right.ID)
}

func faceAttendanceTodaySortTime(item model.FaceAttendanceTodayDetailItem, ascending bool) time.Time {
	if ascending {
		if item.LessonStartAt != nil {
			return *item.LessonStartAt
		}
		if item.AttendanceTime != nil {
			return *item.AttendanceTime
		}
		return time.Time{}
	}
	if item.AttendanceTime != nil {
		return *item.AttendanceTime
	}
	if item.LessonStartAt != nil {
		return *item.LessonStartAt
	}
	return time.Time{}
}
