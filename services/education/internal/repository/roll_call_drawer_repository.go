package repository

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

func (repo *Repository) GetRollCallClassTimetable(ctx context.Context, instID int64, dto model.RollCallClassTimetableQueryDTO) (model.RollCallClassTimetableResult, error) {
	detail, classMeta, err := repo.loadRollCallDrawerContext(ctx, instID, strings.TrimSpace(dto.ID))
	if err != nil {
		return model.RollCallClassTimetableResult{}, err
	}

	lessonDay := formatRollCallLessonDayString(dto.LessonDay, detail.LessonDate)
	students := make([]model.RollCallClassTimetableStudentVO, 0, len(detail.Students))
	for _, item := range detail.Students {
		students = append(students, model.RollCallClassTimetableStudentVO{
			SourceType:                   rollCallClassTimetableSourceType(item.ScheduleStudentType),
			SourceID:                     "0",
			StudentID:                    item.StudentID,
			StudentName:                  item.StudentName,
			StudentAvatar:                strings.TrimSpace(item.AvatarURL),
			StudentPhone:                 firstNonEmptyString(strings.TrimSpace(item.MaskedPhone), maskStudentMobile(strings.TrimSpace(item.Phone))),
			StudentPhoneRelationshipType: item.PhoneRelationship,
		})
	}

	return model.RollCallClassTimetableResult{
		Detail: model.RollCallClassTimetableDetailVO{
			ID:                         detail.ID,
			ClassID:                    classMeta.ClassID,
			ClassName:                  classMeta.ClassName,
			ClassTimes:                 classMeta.DefaultStudentClassTime,
			DefaultStudentClassTime:    classMeta.DefaultStudentClassTime,
			DefaultTeacherClassTime:    classMeta.DefaultTeacherClassTime,
			DefaultClassTimeRecordMode: classMeta.DefaultClassTimeRecordMode,
			LessonPrice:                classMeta.LessonPrice,
			Teachers:                   classMeta.Teachers,
			AddressType:                0,
			AddressID:                  firstNonEmptyString(detail.ClassroomID, "0"),
			AddressName:                detail.ClassroomName,
			LessonID:                   detail.LessonID,
			LessonName:                 detail.LessonName,
			LessonType:                 classMeta.LessonType,
			StartMinutes:               detail.StartAt.Hour()*60 + detail.StartAt.Minute(),
			EndMinutes:                 detail.EndAt.Hour()*60 + detail.EndAt.Minute(),
			RepeatSpan:                 0,
			WeekDays:                   0,
			StartDate:                  zeroRollCallDateTime(),
			EndDate:                    zeroRollCallDateTime(),
			LessonCount:                1,
			Remark:                     detail.Remark,
			ExternalRemark:             "",
			LessonDays: []model.RollCallClassTimetableLessonDayVO{{
				LessonDay:        lessonDay,
				IsFinished:       detail.CallStatus == 2,
				LessonDayIndex:   0,
				Students:         students,
				RemoveStudent:    []model.RollCallClassTimetableStudentVO{},
				TeachingRecordID: "0",
			}},
			IsBookLesson:     false,
			SubjectID:        "0",
			SubjectName:      "",
			IsOrgCreated:     false,
			SchoolID:         "",
			SchoolName:       "",
			IsOpenLiveRecord: false,
			IsOpenLive:       false,
		},
	}, nil
}

func (repo *Repository) GetRollCallTeachingRecordStudentList(ctx context.Context, instID int64, dto model.RollCallTeachingRecordStudentListQueryDTO) (model.RollCallTeachingRecordStudentListResult, error) {
	detail, classMeta, err := repo.loadRollCallDrawerContext(ctx, instID, strings.TrimSpace(dto.TimetableSourceID))
	if err != nil {
		return model.RollCallTeachingRecordStudentListResult{}, err
	}

	sources := make([]rollCallDrawerStudentSource, 0, len(detail.Students)+len(detail.LeaveStudents))
	for _, item := range detail.Students {
		sources = append(sources, rollCallDrawerStudentSource{
			Student:               item,
			DefaultTeachingStatus: 1,
		})
	}
	for _, item := range detail.LeaveStudents {
		sources = append(sources, rollCallDrawerStudentSource{
			Student:               item,
			DefaultTeachingStatus: 3,
		})
	}

	studentIDs := make([]int64, 0, len(sources))
	for _, source := range sources {
		studentID, err := strconv.ParseInt(strings.TrimSpace(source.Student.StudentID), 10, 64)
		if err != nil || studentID <= 0 {
			continue
		}
		studentIDs = append(studentIDs, studentID)
	}
	profileMap, err := repo.loadRollCallDrawerStudentProfileMap(ctx, instID, studentIDs)
	if err != nil {
		return model.RollCallTeachingRecordStudentListResult{}, err
	}

	students := make([]model.RollCallTeachingRecordStudentVO, 0, len(sources))
	for _, source := range sources {
		studentID, _ := strconv.ParseInt(strings.TrimSpace(source.Student.StudentID), 10, 64)
		profile := profileMap[studentID]
		account, ok, err := repo.pickRollCallDrawerStudentAccount(ctx, instID, studentID, detail, source.Student.ScheduleStudentType)
		if err != nil {
			return model.RollCallTeachingRecordStudentListResult{}, err
		}
		chargingMode := 0
		quantity := 0.0
		paidRemaining := 0.0
		isActive := false
		tuitionAccountID := "0"
		if ok {
			chargingMode = rollCallAccountChargingMode(account)
			quantity = account.Quantity + account.FreeQuantity
			paidRemaining = account.Tuition
			isActive = account.IsTuitionAccountActive
			tuitionAccountID = firstNonEmptyString(strings.TrimSpace(account.ID), "0")
		}
		students = append(students, model.RollCallTeachingRecordStudentVO{
			StudentID:                    source.Student.StudentID,
			StudentName:                  source.Student.StudentName,
			Avatar:                       strings.TrimSpace(source.Student.AvatarURL),
			IsBindChild:                  profile.IsBindChild,
			Quantity:                     quantity,
			PaidRemaining:                paidRemaining,
			ChargingMode:                 chargingMode,
			IsTuitionAccountActive:       isActive,
			MakeUpTeachingRecordID:       "0",
			AbsentStudentType:            0,
			TuitionAccountID:             tuitionAccountID,
			SourceType:                   rollCallTeachingRecordSourceType(source.Student.ScheduleStudentType),
			StudentTeachingStatus:        0,
			DefaultStudentTeachingStatus: source.DefaultTeachingStatus,
			HasSignIn:                    false,
			IsCrossSchoolStudent:         false,
		})
	}

	return model.RollCallTeachingRecordStudentListResult{
		Data: model.RollCallTeachingRecordMetaVO{
			SourceName:          classMeta.ClassName,
			SourceType:          rollCallTeachingRecordSourceCategory(detail.ClassType),
			SourceID:            classMeta.ClassID,
			LessonID:            detail.LessonID,
			TimetableSourceType: rollCallTeachingRecordSourceCategory(detail.ClassType),
			Tag:                 1,
			TimetableSourceID:   detail.ID,
			StartTime:           detail.StartAt.Format("2006-01-02T15:04:05"),
			EndTime:             detail.EndAt.Format("2006-01-02T15:04:05"),
			TeacherClassTime:    classMeta.DefaultTeacherClassTime,
			ClassroomID:         firstNonEmptyString(detail.ClassroomID, "0"),
		},
		Teachers: classMeta.TeachingRecordTeachers,
		Students: students,
	}, nil
}

func (repo *Repository) GetRollCallStudentLeaveCount(ctx context.Context, instID int64, dto model.RollCallStudentLeaveCountQueryDTO) ([]model.RollCallStudentLeaveCountVO, error) {
	studentIDs := make([]int64, 0, len(dto.StudentIDs))
	seen := make(map[int64]struct{}, len(dto.StudentIDs))
	for _, raw := range dto.StudentIDs {
		studentID, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
		if err != nil || studentID <= 0 {
			continue
		}
		if _, ok := seen[studentID]; ok {
			continue
		}
		seen[studentID] = struct{}{}
		studentIDs = append(studentIDs, studentID)
	}
	if len(studentIDs) == 0 {
		return []model.RollCallStudentLeaveCountVO{}, nil
	}
	lessonID, err := strconv.ParseInt(strings.TrimSpace(dto.LessonID), 10, 64)
	if err != nil || lessonID <= 0 {
		return []model.RollCallStudentLeaveCountVO{}, nil
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			CAST(tss.student_id AS CHAR),
			COUNT(*)
		FROM teaching_schedule_student tss
		INNER JOIN teaching_schedule ts
			ON ts.id = tss.teaching_schedule_id
		   AND ts.inst_id = tss.inst_id
		   AND ts.del_flag = 0
		   AND ts.status = ?
		WHERE tss.inst_id = ?
		  AND tss.del_flag = 0
		  AND tss.student_id IN (`+sqlPlaceholders(len(studentIDs))+`)
		  AND tss.roster_status = ?
		  AND ts.lesson_id = ?
		GROUP BY tss.student_id
	`, append([]any{model.TeachingScheduleStatusActive, instID}, append(int64SliceToAny(studentIDs), model.TeachingScheduleStudentRosterStatusLeave, lessonID)...)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]model.RollCallStudentLeaveCountVO, 0, len(studentIDs))
	for rows.Next() {
		var item model.RollCallStudentLeaveCountVO
		if err := rows.Scan(&item.StudentID, &item.LeaveCount); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (repo *Repository) GetRollCallStudentTuitionExtraInfo(ctx context.Context, instID int64, dto model.RollCallStudentTuitionExtraInfoQueryDTO) ([]model.RollCallStudentTuitionExtraInfoVO, error) {
	lessonID := strings.TrimSpace(dto.LessonID)
	if lessonID == "" {
		return []model.RollCallStudentTuitionExtraInfoVO{}, nil
	}
	courseScope, err := repo.resolveRollCallDrawerLessonCourseScope(ctx, instID, lessonID)
	if err != nil {
		return nil, err
	}
	parsedStudentIDs := make([]int64, 0, len(dto.StudentIDs))
	for _, rawStudentID := range dto.StudentIDs {
		studentID, err := strconv.ParseInt(strings.TrimSpace(rawStudentID), 10, 64)
		if err != nil || studentID <= 0 {
			continue
		}
		parsedStudentIDs = append(parsedStudentIDs, studentID)
	}
	scheduleOnlyStudentMap, err := repo.loadRollCallDrawerLessonScheduleOnlyStudentMap(ctx, instID, lessonID, parsedStudentIDs)
	if err != nil {
		return nil, err
	}

	result := make([]model.RollCallStudentTuitionExtraInfoVO, 0, len(parsedStudentIDs))
	for _, studentID := range parsedStudentIDs {
		accounts, bestIndex, err := repo.listRollCallDrawerStudentAccountsByCourseScope(ctx, instID, studentID, courseScope, 0, model.TeachingClassTypeNormal)
		if err != nil {
			return nil, err
		}
		if len(accounts) == 0 && scheduleOnlyStudentMap[studentID] {
			accounts, err = repo.listRollCallDrawerAllStudentAccounts(ctx, instID, studentID)
			if err != nil {
				return nil, err
			}
			bestIndex = pickBestRollCallDrawerAccountIndex(accounts)
		}
		bestName := ""
		if len(accounts) > 0 {
			if bestIndex < 0 || bestIndex >= len(accounts) {
				bestIndex = 0
			}
			bestName = firstNonEmptyString(strings.TrimSpace(accounts[bestIndex].ProductName), strings.TrimSpace(accounts[bestIndex].LessonName))
		}
		result = append(result, model.RollCallStudentTuitionExtraInfoVO{
			StudentID:            strconv.FormatInt(studentID, 10),
			MutilTuition:         len(accounts) > 1,
			BestMatchProductName: bestName,
		})
	}
	return result, nil
}

func (repo *Repository) GetRollCallStudentTuitionAccounts(ctx context.Context, instID int64, dto model.StudentLessonTuitionAccountsQueryDTO) ([]model.StudentLessonTuitionAccountItem, error) {
	lessonID := strings.TrimSpace(dto.LessonID)
	if lessonID == "" {
		return []model.StudentLessonTuitionAccountItem{}, nil
	}
	studentID, err := strconv.ParseInt(strings.TrimSpace(dto.StudentID), 10, 64)
	if err != nil || studentID <= 0 {
		return []model.StudentLessonTuitionAccountItem{}, nil
	}

	courseScope, err := repo.resolveRollCallDrawerLessonCourseScope(ctx, instID, lessonID)
	if err != nil {
		return nil, err
	}
	accounts, _, err := repo.listRollCallDrawerStudentAccountsByCourseScope(ctx, instID, studentID, courseScope, 0, model.TeachingClassTypeNormal)
	if err != nil {
		return nil, err
	}

	scheduleOnlyStudentMap, err := repo.loadRollCallDrawerLessonScheduleOnlyStudentMap(ctx, instID, lessonID, []int64{studentID})
	if err != nil {
		return nil, err
	}
	if len(accounts) == 0 && scheduleOnlyStudentMap[studentID] {
		accounts, err = repo.listRollCallDrawerAllStudentAccounts(ctx, instID, studentID)
		if err != nil {
			return nil, err
		}
	}
	if accounts == nil {
		return []model.StudentLessonTuitionAccountItem{}, nil
	}
	return accounts, nil
}

type rollCallDrawerContext struct {
	ClassID                    string
	ClassName                  string
	DefaultStudentClassTime    float64
	DefaultTeacherClassTime    float64
	DefaultClassTimeRecordMode int
	LessonPrice                float64
	LessonType                 int
	Teachers                   []model.RollCallClassTimetableTeacherVO
	TeachingRecordTeachers     []model.RollCallTeachingRecordTeacherVO
}

type rollCallDrawerStudentSource struct {
	Student               model.TeachingScheduleDetailStudentVO
	DefaultTeachingStatus int
}

type rollCallDrawerStudentProfile struct {
	IsBindChild bool
}

func (repo *Repository) loadRollCallDrawerContext(ctx context.Context, instID int64, scheduleID string) (model.TeachingScheduleDetailVO, rollCallDrawerContext, error) {
	detail, err := repo.GetTeachingScheduleDetail(ctx, instID, model.TeachingScheduleDetailQueryDTO{ID: scheduleID})
	if err != nil {
		return model.TeachingScheduleDetailVO{}, rollCallDrawerContext{}, err
	}

	ctxVO := rollCallDrawerContext{
		ClassID:                    detail.TeachingClassID,
		ClassName:                  detail.TeachingClassName,
		DefaultStudentClassTime:    1,
		DefaultTeacherClassTime:    0,
		DefaultClassTimeRecordMode: 1,
		LessonType:                 1,
		Teachers:                   buildRollCallClassTeachers(detail),
		TeachingRecordTeachers:     buildRollCallTeachingRecordTeachers(detail),
	}

	classID, _ := strconv.ParseInt(strings.TrimSpace(detail.TeachingClassID), 10, 64)
	if detail.ClassType == model.TeachingClassTypeNormal && classID > 0 {
		classDetail, err := repo.GetGroupClassByID(ctx, instID, classID)
		if err == nil {
			ctxVO.ClassID = classDetail.ID
			ctxVO.ClassName = classDetail.Name
			ctxVO.DefaultStudentClassTime = classDetail.DefaultStudentClassTime
			ctxVO.DefaultTeacherClassTime = classDetail.DefaultTeacherClassTime
			ctxVO.DefaultClassTimeRecordMode = classDetail.DefaultClassTimeRecordMode
			ctxVO.LessonPrice = classDetail.LessonPrice
			ctxVO.LessonType = classDetail.LessonType
		}
		return detail, ctxVO, nil
	}

	if detail.ClassType == model.TeachingClassTypeOneToOne && classID > 0 {
		oneToOneDetail, err := repo.GetOneToOneDetail(ctx, instID, classID)
		if err == nil {
			ctxVO.ClassID = oneToOneDetail.ID
			ctxVO.ClassName = oneToOneDetail.Name
			ctxVO.DefaultStudentClassTime = oneToOneDetail.DefaultStudentClassTime
			ctxVO.DefaultTeacherClassTime = oneToOneDetail.DefaultTeacherClassTime
			ctxVO.DefaultClassTimeRecordMode = oneToOneDetail.DefaultClassTimeRecordMode
			ctxVO.LessonPrice = oneToOneDetail.LessonPrice
			ctxVO.LessonType = 1
			ctxVO.Teachers = make([]model.RollCallClassTimetableTeacherVO, 0, len(oneToOneDetail.TeacherList))
			ctxVO.TeachingRecordTeachers = make([]model.RollCallTeachingRecordTeacherVO, 0, len(oneToOneDetail.TeacherList))
			for _, teacher := range oneToOneDetail.TeacherList {
				duty := 3
				typeCode := 3
				if teacher.IsDefault {
					duty = 1
					typeCode = 1
				}
				ctxVO.Teachers = append(ctxVO.Teachers, model.RollCallClassTimetableTeacherVO{
					TeacherID:     teacher.TeacherID,
					TeacherDuty:   duty,
					TeacherName:   teacher.Name,
					TeacherStatus: teacher.Status,
				})
				ctxVO.TeachingRecordTeachers = append(ctxVO.TeachingRecordTeachers, model.RollCallTeachingRecordTeacherVO{
					TeacherID: teacher.TeacherID,
					Type:      typeCode,
				})
			}
		}
	}

	return detail, ctxVO, nil
}

func buildRollCallClassTeachers(detail model.TeachingScheduleDetailVO) []model.RollCallClassTimetableTeacherVO {
	teachers := make([]model.RollCallClassTimetableTeacherVO, 0, 1+len(detail.AssistantIDs))
	if strings.TrimSpace(detail.TeacherID) != "" {
		teachers = append(teachers, model.RollCallClassTimetableTeacherVO{
			TeacherID:     detail.TeacherID,
			TeacherDuty:   1,
			TeacherName:   detail.TeacherName,
			TeacherStatus: 0,
		})
	}
	for index, assistantID := range detail.AssistantIDs {
		name := ""
		if index < len(detail.AssistantNames) {
			name = detail.AssistantNames[index]
		}
		if strings.TrimSpace(assistantID) == "" {
			continue
		}
		teachers = append(teachers, model.RollCallClassTimetableTeacherVO{
			TeacherID:     assistantID,
			TeacherDuty:   3,
			TeacherName:   name,
			TeacherStatus: 0,
		})
	}
	return teachers
}

func buildRollCallTeachingRecordTeachers(detail model.TeachingScheduleDetailVO) []model.RollCallTeachingRecordTeacherVO {
	teachers := make([]model.RollCallTeachingRecordTeacherVO, 0, 1+len(detail.AssistantIDs))
	if strings.TrimSpace(detail.TeacherID) != "" {
		teachers = append(teachers, model.RollCallTeachingRecordTeacherVO{
			TeacherID: detail.TeacherID,
			Type:      1,
		})
	}
	for _, assistantID := range detail.AssistantIDs {
		if strings.TrimSpace(assistantID) == "" {
			continue
		}
		teachers = append(teachers, model.RollCallTeachingRecordTeacherVO{
			TeacherID: assistantID,
			Type:      3,
		})
	}
	return teachers
}

func (repo *Repository) loadRollCallDrawerStudentProfileMap(ctx context.Context, instID int64, studentIDs []int64) (map[int64]rollCallDrawerStudentProfile, error) {
	studentIDs = uniquePositiveInt64s(studentIDs)
	if len(studentIDs) == 0 {
		return map[int64]rollCallDrawerStudentProfile{}, nil
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT IFNULL(id, 0), IFNULL(is_bind_child, 0)
		FROM inst_student
		WHERE inst_id = ?
		  AND del_flag = 0
		  AND id IN (`+sqlPlaceholders(len(studentIDs))+`)
	`, append([]any{instID}, int64SliceToAny(studentIDs)...)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64]rollCallDrawerStudentProfile, len(studentIDs))
	for rows.Next() {
		var (
			studentID   int64
			isBindChild int
		)
		if err := rows.Scan(&studentID, &isBindChild); err != nil {
			return nil, err
		}
		result[studentID] = rollCallDrawerStudentProfile{
			IsBindChild: isBindChild != 0,
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (repo *Repository) pickRollCallDrawerStudentAccount(ctx context.Context, instID, studentID int64, detail model.TeachingScheduleDetailVO, studentType int) (model.StudentLessonTuitionAccountItem, bool, error) {
	courseScope, err := repo.resolveRollCallDrawerCourseScopeByDetail(ctx, instID, detail)
	if err != nil {
		return model.StudentLessonTuitionAccountItem{}, false, err
	}
	teachingClassID, _ := strconv.ParseInt(strings.TrimSpace(detail.TeachingClassID), 10, 64)
	accounts, bestIndex, err := repo.listRollCallDrawerStudentAccountsByCourseScope(ctx, instID, studentID, courseScope, teachingClassID, detail.ClassType)
	if err != nil {
		return model.StudentLessonTuitionAccountItem{}, false, err
	}
	if len(accounts) == 0 && isScheduleOnlyStudentType(studentType) {
		accounts, err = repo.listRollCallDrawerAllStudentAccounts(ctx, instID, studentID)
		if err != nil {
			return model.StudentLessonTuitionAccountItem{}, false, err
		}
		bestIndex = pickBestRollCallDrawerAccountIndex(accounts)
	}
	if len(accounts) == 0 {
		return model.StudentLessonTuitionAccountItem{}, false, nil
	}
	if bestIndex < 0 || bestIndex >= len(accounts) {
		bestIndex = 0
	}
	return accounts[bestIndex], true, nil
}

func (repo *Repository) listRollCallDrawerStudentAccountsByCourseScope(ctx context.Context, instID, studentID int64, courseScope []int64, teachingClassID int64, classType int) ([]model.StudentLessonTuitionAccountItem, int, error) {
	if len(courseScope) == 0 || studentID <= 0 {
		return nil, -1, nil
	}

	accounts := make([]model.StudentLessonTuitionAccountItem, 0, 4)
	seen := make(map[string]int)
	for _, courseID := range courseScope {
		currentAccounts, err := repo.ListStudentTuitionAccountsByStudentAndLesson(ctx, instID, studentID, courseID, teachingClassID, 0)
		if err != nil {
			return nil, -1, err
		}
		if len(currentAccounts) == 0 && teachingClassID > 0 {
			currentAccounts, err = repo.ListStudentTuitionAccountsByStudentAndLesson(ctx, instID, studentID, courseID, 0, 0)
			if err != nil {
				return nil, -1, err
			}
		}
		for _, account := range currentAccounts {
			accounts = appendRollCallDrawerAccount(accounts, account, seen)
		}
	}

	if len(accounts) == 0 && classType == model.TeachingClassTypeNormal {
		groupAccounts, _, err := repo.PageTuitionAccountsByLessonForGroupAdd(ctx, instID, courseScope, teachingClassID, []int64{studentID}, 1, 20, model.TuitionAccountLessonPageFilters{})
		if err != nil {
			return nil, -1, err
		}
		accountIDs := make([]int64, 0, len(groupAccounts))
		for _, row := range groupAccounts {
			accountID, _ := strconv.ParseInt(strings.TrimSpace(row.TuitionAccountID), 10, 64)
			if accountID > 0 {
				accountIDs = append(accountIDs, accountID)
			}
		}
		fallbackMap, err := repo.loadRollCallDrawerTuitionAccountMap(ctx, instID, accountIDs)
		if err != nil {
			return nil, -1, err
		}
		for _, row := range groupAccounts {
			account := fallbackMap[strings.TrimSpace(row.TuitionAccountID)]
			if strings.TrimSpace(account.ID) == "" {
				account = model.StudentLessonTuitionAccountItem{
					ID:                     strings.TrimSpace(row.TuitionAccountID),
					StudentID:              strings.TrimSpace(row.StudentID),
					ProductName:            strings.TrimSpace(row.ProductName),
					LessonChargingMode:     row.LessonChargingMode,
					Quantity:               row.Quantity,
					LessonType:             1,
					IsTuitionAccountActive: row.IsTuitionAccountActive,
					StartTime:              row.StartTime,
				}
			}
			accounts = appendRollCallDrawerAccount(accounts, account, seen)
		}
	}

	if len(accounts) == 0 {
		return nil, -1, nil
	}
	return accounts, pickBestRollCallDrawerAccountIndex(accounts), nil
}

func (repo *Repository) listRollCallDrawerAllStudentAccounts(ctx context.Context, instID, studentID int64) ([]model.StudentLessonTuitionAccountItem, error) {
	if studentID <= 0 {
		return nil, nil
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			CAST(MIN(ta.id) AS CHAR),
			CAST(MIN(ta.student_id) AS CHAR),
			CAST(ta.course_id AS CHAR),
			IFNULL(MAX(ic.name), '') AS lesson_name,
			MAX(IFNULL(NULLIF(TRIM(ic.name), ''), IFNULL(icq.name, ''))) AS product_name,
			IFNULL(MAX(icq.lesson_model), 0) AS lesson_charging_mode,
			SUM(CASE
				WHEN IFNULL(icq.lesson_model, 0) IN (3, 4) THEN IFNULL(ta.total_tuition, 0)
				WHEN IFNULL(ta.total_quantity, 0) > 0 THEN IFNULL(ta.total_quantity, 0)
				ELSE 0
			END) AS total_quantity_display,
			SUM(CASE
				WHEN IFNULL(icq.lesson_model, 0) IN (3, 4) THEN IFNULL(ta.free_quantity, 0)
				WHEN IFNULL(ta.total_quantity, 0) = 0 AND IFNULL(ta.free_quantity, 0) > 0 THEN IFNULL(ta.free_quantity, 0)
				ELSE 0
			END) AS total_free_quantity_display,
			SUM(IFNULL(ta.total_tuition, 0)),
			SUM(CASE
				WHEN IFNULL(ta.status, 0) = 3 THEN 0
				WHEN IFNULL(ta.total_quantity, 0) = 0 AND IFNULL(ta.free_quantity, 0) > 0 THEN IFNULL(ta.remaining_quantity, 0)
				ELSE 0
			END) AS remain_free_quantity,
			SUM(CASE
				WHEN IFNULL(ta.status, 0) = 3 THEN 0
				WHEN IFNULL(icq.lesson_model, 0) IN (3, 4) THEN IFNULL(ta.remaining_tuition, 0)
				WHEN IFNULL(ta.total_quantity, 0) > 0 THEN IFNULL(ta.remaining_quantity, 0)
				ELSE 0
			END) AS remain_quantity_display,
			SUM(CASE
				WHEN IFNULL(ta.status, 0) = 3 THEN 0
				ELSE IFNULL(ta.remaining_tuition, 0)
			END),
			MIN(IFNULL(ta.create_time, NOW())) AS start_time,
			IFNULL(MAX(ta.enable_expire_time), 0) AS enable_expire,
			MAX(ta.expire_time),
			MAX(ta.valid_date),
			IFNULL(MAX(ic.teach_method), 0) AS teach_method,
			`+effectiveTuitionAccountStatusSQL+` AS ta_status
		FROM tuition_account ta
		INNER JOIN inst_course ic ON ic.id = ta.course_id AND ic.inst_id = ta.inst_id AND ic.del_flag = 0
		LEFT JOIN sale_order_course_detail sod ON sod.id = ta.order_course_detail_id AND sod.del_flag = 0
		LEFT JOIN inst_course_quotation icq ON icq.id = COALESCE(
			NULLIF(ta.quote_id, 0),
			NULLIF(sod.quote_id, 0),
			(SELECT qx.id FROM inst_course_quotation qx
			 WHERE qx.course_id = ta.course_id AND qx.del_flag = 0
			   AND ABS(IFNULL(qx.quantity, 0) - IFNULL(ta.total_quantity, 0)) < 0.000001
			   AND ABS(IFNULL(qx.price, 0) - IFNULL(ta.total_tuition, 0)) < 0.000001
			 ORDER BY qx.id DESC LIMIT 1),
			(SELECT qmin.id FROM inst_course_quotation qmin
			 WHERE qmin.course_id = ta.course_id AND qmin.del_flag = 0
			 ORDER BY qmin.id ASC LIMIT 1)
		) AND icq.del_flag = 0
		WHERE ta.inst_id = ?
		  AND ta.del_flag = 0
		  AND ta.student_id = ?
		  AND IFNULL(ta.status, 0) <> 3
		GROUP BY ta.course_id, IFNULL(ic.teach_method, 0), IFNULL(icq.lesson_model, -99999)
		ORDER BY
			MAX(IFNULL(ta.enable_expire_time, 0)) DESC,
			MAX(IFNULL(ta.expire_time, '9999-12-31 23:59:59')) ASC,
			MIN(IFNULL(ta.create_time, NOW())) DESC,
			MIN(ta.id) DESC
	`, instID, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]model.StudentLessonTuitionAccountItem, 0, 4)
	for rows.Next() {
		var (
			item                                               model.StudentLessonTuitionAccountItem
			startTime, expireTime, validDate                   sql.NullTime
			enableExpireRaw                                    int64
			totalQtyDisp, totalFreeDisp, remainFree, remainQty float64
			taStatus                                           int
		)
		if err := rows.Scan(
			&item.ID,
			&item.StudentID,
			&item.LessonID,
			&item.LessonName,
			&item.ProductName,
			&item.LessonChargingMode,
			&totalQtyDisp,
			&totalFreeDisp,
			&item.TotalTuition,
			&remainFree,
			&remainQty,
			&item.Tuition,
			&startTime,
			&enableExpireRaw,
			&expireTime,
			&validDate,
			&item.LessonType,
			&taStatus,
		); err != nil {
			return nil, err
		}
		item.TotalQuantity = totalQtyDisp
		item.TotalFreeQuantity = totalFreeDisp
		item.FreeQuantity = remainFree
		item.Quantity = remainQty
		item.EnableExpireTime = enableExpireRaw != 0
		item.Status = taStatus
		item.IsTuitionAccountActive = taStatus == 1
		if startTime.Valid {
			item.StartTime = startTime.Time
		}
		if expireTime.Valid {
			item.ExpireTime = expireTime.Time
		}
		if validDate.Valid {
			item.LatestStartTime = validDate.Time
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func appendRollCallDrawerAccount(accounts []model.StudentLessonTuitionAccountItem, account model.StudentLessonTuitionAccountItem, seen map[string]int) []model.StudentLessonTuitionAccountItem {
	key := strings.TrimSpace(account.ID)
	if key == "" {
		key = strings.Join([]string{
			strings.TrimSpace(account.LessonID),
			strings.TrimSpace(account.ProductName),
			strconv.Itoa(rollCallAccountChargingMode(account)),
		}, "|")
	}
	if index, ok := seen[key]; ok {
		accounts[index] = mergeRollCallDrawerAccount(accounts[index], account)
		return accounts
	}
	seen[key] = len(accounts)
	return append(accounts, account)
}

func mergeRollCallDrawerAccount(base, extra model.StudentLessonTuitionAccountItem) model.StudentLessonTuitionAccountItem {
	base.ID = firstNonEmptyString(base.ID, extra.ID)
	base.StudentID = firstNonEmptyString(base.StudentID, extra.StudentID)
	base.LessonID = firstNonEmptyString(base.LessonID, extra.LessonID)
	base.LessonName = firstNonEmptyString(base.LessonName, extra.LessonName)
	base.ProductName = firstNonEmptyString(base.ProductName, extra.ProductName)
	if base.LessonChargingMode == 0 && extra.LessonChargingMode != 0 {
		base.LessonChargingMode = extra.LessonChargingMode
	}
	if base.TotalQuantity == 0 && extra.TotalQuantity != 0 {
		base.TotalQuantity = extra.TotalQuantity
	}
	if base.TotalFreeQuantity == 0 && extra.TotalFreeQuantity != 0 {
		base.TotalFreeQuantity = extra.TotalFreeQuantity
	}
	if base.TotalTuition == 0 && extra.TotalTuition != 0 {
		base.TotalTuition = extra.TotalTuition
	}
	if base.FreeQuantity == 0 && extra.FreeQuantity != 0 {
		base.FreeQuantity = extra.FreeQuantity
	}
	if base.Quantity == 0 && extra.Quantity != 0 {
		base.Quantity = extra.Quantity
	}
	if base.Tuition == 0 && extra.Tuition != 0 {
		base.Tuition = extra.Tuition
	}
	if !base.EnableExpireTime && extra.EnableExpireTime {
		base.EnableExpireTime = true
	}
	if base.ExpireTime.IsZero() && !extra.ExpireTime.IsZero() {
		base.ExpireTime = extra.ExpireTime
	}
	if base.StartTime.IsZero() && !extra.StartTime.IsZero() {
		base.StartTime = extra.StartTime
	}
	if base.LatestStartTime.IsZero() && !extra.LatestStartTime.IsZero() {
		base.LatestStartTime = extra.LatestStartTime
	}
	if base.LessonType == 0 && extra.LessonType != 0 {
		base.LessonType = extra.LessonType
	}
	if !base.IsTuitionAccountActive && extra.IsTuitionAccountActive {
		base.IsTuitionAccountActive = true
	}
	if base.Status == 0 && extra.Status != 0 {
		base.Status = extra.Status
	}
	return base
}

func pickBestRollCallDrawerAccountIndex(accounts []model.StudentLessonTuitionAccountItem) int {
	bestIndex := 0
	bestScore := -1
	for index, account := range accounts {
		score := 0
		mode := rollCallAccountChargingMode(account)
		switch mode {
		case 2:
			score += 13
		case 1:
			score += 12
		case 3:
			score += 11
		}
		if account.IsTuitionAccountActive {
			score += 4
		}
		if account.EnableExpireTime {
			score += 2
		}
		if account.LessonType > 0 && account.LessonType != 1 {
			score++
		}
		if (account.Quantity+account.FreeQuantity) > 0 || account.Tuition > 0 {
			score += 2
		}
		if strings.TrimSpace(account.ProductName) != "" {
			score++
		}
		if score > bestScore {
			bestScore = score
			bestIndex = index
		}
	}
	return bestIndex
}

func (repo *Repository) resolveRollCallDrawerCourseScopeByDetail(ctx context.Context, instID int64, detail model.TeachingScheduleDetailVO) ([]int64, error) {
	if detail.ClassType == model.TeachingClassTypeNormal {
		return repo.resolveRollCallDrawerLessonCourseScope(ctx, instID, detail.LessonID)
	}
	lessonID, err := strconv.ParseInt(strings.TrimSpace(detail.LessonID), 10, 64)
	if err != nil || lessonID <= 0 {
		return nil, nil
	}
	return []int64{lessonID}, nil
}

func (repo *Repository) resolveRollCallDrawerLessonCourseScope(ctx context.Context, instID int64, lessonID string) ([]int64, error) {
	trimmed := strings.TrimSpace(lessonID)
	if trimmed == "" {
		return nil, nil
	}
	courseScope, _, err := repo.ResolveGroupClassLessonCourseScope(ctx, instID, trimmed)
	if err == nil && len(courseScope) > 0 {
		return courseScope, nil
	}
	parsedLessonID, parseErr := strconv.ParseInt(trimmed, 10, 64)
	if parseErr != nil || parsedLessonID <= 0 {
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, nil
	}
	return []int64{parsedLessonID}, nil
}

func (repo *Repository) loadRollCallDrawerLessonScheduleOnlyStudentMap(ctx context.Context, instID int64, lessonID string, studentIDs []int64) (map[int64]bool, error) {
	studentIDs = uniquePositiveInt64s(studentIDs)
	if len(studentIDs) == 0 {
		return map[int64]bool{}, nil
	}

	parsedLessonID, err := strconv.ParseInt(strings.TrimSpace(lessonID), 10, 64)
	if err != nil || parsedLessonID <= 0 {
		return map[int64]bool{}, nil
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT DISTINCT tss.student_id
		FROM teaching_schedule_student tss
		INNER JOIN teaching_schedule ts
			ON ts.id = tss.teaching_schedule_id
		   AND ts.inst_id = tss.inst_id
		   AND ts.del_flag = 0
		   AND ts.status = ?
		WHERE tss.inst_id = ?
		  AND tss.del_flag = 0
		  AND ts.lesson_id = ?
		  AND tss.student_id IN (`+sqlPlaceholders(len(studentIDs))+`)
		  AND tss.roster_status = ?
		  AND tss.student_type IN (?, ?, ?)
	`, append(
		[]any{
			model.TeachingScheduleStatusActive,
			instID,
			parsedLessonID,
		},
		append(
			int64SliceToAny(studentIDs),
			model.TeachingScheduleStudentRosterStatusActive,
			model.TeachingScheduleStudentTypeTemporary,
			model.TeachingScheduleStudentTypeTrial,
			model.TeachingScheduleStudentTypeMakeup,
		)...,
	)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64]bool, len(studentIDs))
	for rows.Next() {
		var studentID int64
		if err := rows.Scan(&studentID); err != nil {
			return nil, err
		}
		result[studentID] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (repo *Repository) loadRollCallDrawerTuitionAccountMap(ctx context.Context, instID int64, accountIDs []int64) (map[string]model.StudentLessonTuitionAccountItem, error) {
	accountIDs = uniquePositiveInt64s(accountIDs)
	if len(accountIDs) == 0 {
		return map[string]model.StudentLessonTuitionAccountItem{}, nil
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			CAST(ta.id AS CHAR),
			CAST(ta.student_id AS CHAR),
			CAST(ta.course_id AS CHAR),
			IFNULL(ic.name, ''),
			IFNULL(NULLIF(TRIM(ic.name), ''), IFNULL(icq.name, '')) AS product_name,
			IFNULL(icq.lesson_model, 0) AS lesson_charging_mode,
			CASE
				WHEN IFNULL(icq.lesson_model, 0) IN (3, 4) THEN IFNULL(ta.total_tuition, 0)
				WHEN IFNULL(ta.total_quantity, 0) > 0 THEN IFNULL(ta.total_quantity, 0)
				ELSE 0
			END,
			CASE
				WHEN IFNULL(icq.lesson_model, 0) IN (3, 4) THEN IFNULL(ta.free_quantity, 0)
				WHEN IFNULL(ta.total_quantity, 0) = 0 AND IFNULL(ta.free_quantity, 0) > 0 THEN IFNULL(ta.free_quantity, 0)
				ELSE 0
			END,
			IFNULL(ta.total_tuition, 0),
			CASE
				WHEN IFNULL(ta.status, 0) = 3 THEN 0
				WHEN IFNULL(ta.total_quantity, 0) = 0 AND IFNULL(ta.free_quantity, 0) > 0 THEN IFNULL(ta.remaining_quantity, 0)
				ELSE 0
			END,
			CASE
				WHEN IFNULL(ta.status, 0) = 3 THEN 0
				WHEN IFNULL(icq.lesson_model, 0) IN (3, 4) THEN IFNULL(ta.remaining_tuition, 0)
				WHEN IFNULL(ta.total_quantity, 0) > 0 THEN IFNULL(ta.remaining_quantity, 0)
				ELSE 0
			END,
			CASE
				WHEN IFNULL(ta.status, 0) = 3 THEN 0
				ELSE IFNULL(ta.remaining_tuition, 0)
			END,
			IFNULL(ta.create_time, NOW()),
			IFNULL(ta.enable_expire_time, 0),
			ta.expire_time,
			ta.valid_date,
			IFNULL(ic.teach_method, 0),
			IFNULL(ta.status, 0)
		FROM tuition_account ta
		INNER JOIN inst_course ic ON ic.id = ta.course_id AND ic.del_flag = 0
		LEFT JOIN sale_order_course_detail sod ON sod.id = ta.order_course_detail_id AND sod.del_flag = 0
		LEFT JOIN inst_course_quotation icq ON icq.id = COALESCE(
			NULLIF(ta.quote_id, 0),
			NULLIF(sod.quote_id, 0),
			(SELECT qx.id FROM inst_course_quotation qx
			 WHERE qx.course_id = ta.course_id AND qx.del_flag = 0
			   AND ABS(IFNULL(qx.quantity, 0) - IFNULL(ta.total_quantity, 0)) < 0.000001
			   AND ABS(IFNULL(qx.price, 0) - IFNULL(ta.total_tuition, 0)) < 0.000001
			 ORDER BY qx.id DESC LIMIT 1),
			(SELECT qmin.id FROM inst_course_quotation qmin
			 WHERE qmin.course_id = ta.course_id AND qmin.del_flag = 0
			 ORDER BY qmin.id ASC LIMIT 1)
		) AND icq.del_flag = 0
		WHERE ta.inst_id = ?
		  AND ta.del_flag = 0
		  AND ta.id IN (`+sqlPlaceholders(len(accountIDs))+`)
	`, append([]any{instID}, int64SliceToAny(accountIDs)...)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]model.StudentLessonTuitionAccountItem, len(accountIDs))
	for rows.Next() {
		var (
			item                                             model.StudentLessonTuitionAccountItem
			startTime, expireTime, validDate                 sql.NullTime
			enableExpireRaw                                  int64
			totalQty, totalFreeQty, remainFreeQty, remainQty float64
			status                                           int
		)
		if err := rows.Scan(
			&item.ID,
			&item.StudentID,
			&item.LessonID,
			&item.LessonName,
			&item.ProductName,
			&item.LessonChargingMode,
			&totalQty,
			&totalFreeQty,
			&item.TotalTuition,
			&remainFreeQty,
			&remainQty,
			&item.Tuition,
			&startTime,
			&enableExpireRaw,
			&expireTime,
			&validDate,
			&item.LessonType,
			&status,
		); err != nil {
			return nil, err
		}
		item.TotalQuantity = totalQty
		item.TotalFreeQuantity = totalFreeQty
		item.FreeQuantity = remainFreeQty
		item.Quantity = remainQty
		item.EnableExpireTime = enableExpireRaw != 0
		item.Status = status
		item.IsTuitionAccountActive = status == 1
		if startTime.Valid {
			item.StartTime = startTime.Time
		}
		if expireTime.Valid {
			item.ExpireTime = expireTime.Time
		}
		if validDate.Valid {
			item.LatestStartTime = validDate.Time
		}
		result[strings.TrimSpace(item.ID)] = item
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func rollCallAccountChargingMode(account model.StudentLessonTuitionAccountItem) int {
	mode := normalizeRollCallDrawerChargingMode(account.LessonChargingMode)
	if mode > 0 {
		return mode
	}
	totalQty := account.TotalQuantity + account.TotalFreeQuantity
	remainQty := account.Quantity + account.FreeQuantity
	if account.EnableExpireTime && (totalQty > 0 || remainQty > 0) {
		return 2
	}
	if totalQty > 0 || remainQty > 0 {
		return 1
	}
	if account.TotalTuition > 0 || account.Tuition > 0 {
		return 3
	}
	return 0
}

func normalizeRollCallDrawerChargingMode(mode int) int {
	if mode == 4 {
		return 3
	}
	switch mode {
	case 1, 2, 3:
		return mode
	default:
		return 0
	}
}

func rollCallClassTimetableSourceType(studentType int) int {
	switch normalizeTeachingScheduleStudentType(studentType) {
	case model.TeachingScheduleStudentTypeTemporary:
		return 4
	case model.TeachingScheduleStudentTypeTrial:
		return 2
	case model.TeachingScheduleStudentTypeMakeup:
		return 3
	default:
		return 6
	}
}

func rollCallTeachingRecordSourceType(studentType int) int {
	switch normalizeTeachingScheduleStudentType(studentType) {
	case model.TeachingScheduleStudentTypeTemporary:
		return 2
	case model.TeachingScheduleStudentTypeTrial:
		return 4
	case model.TeachingScheduleStudentTypeMakeup:
		return 3
	default:
		return 5
	}
}

func rollCallTeachingRecordSourceCategory(classType int) int {
	if classType == model.TeachingClassTypeOneToOne {
		return 2
	}
	return 1
}

func formatRollCallLessonDayString(raw string, fallbackDate string) string {
	text := strings.TrimSpace(raw)
	if text == "" {
		text = strings.TrimSpace(fallbackDate)
	}
	if text == "" {
		return zeroRollCallDateTime()
	}
	if parsed, err := time.Parse(time.RFC3339, text); err == nil {
		return parsed.Format("2006-01-02T15:04:05")
	}
	if parsed, err := time.ParseInLocation("2006-01-02", text, time.Local); err == nil {
		return parsed.Format("2006-01-02T15:04:05")
	}
	if parsed, err := time.ParseInLocation("2006-01-02 15:04:05", text, time.Local); err == nil {
		return parsed.Format("2006-01-02T15:04:05")
	}
	return text
}

func zeroRollCallDateTime() string {
	return "0001-01-01T00:00:00"
}
