package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

type rollCallConfirmStudentProfile struct {
	StudentID          int64
	StudentName        string
	StudentPhone       string
	AvatarURL          string
	StudentStatus      int
	AdvisorStaffID     int64
	AdvisorStaffName   string
	StudentManagerID   int64
	StudentManagerName string
}

type rollCallConfirmAccount struct {
	ID                 int64
	StudentID          int64
	CourseID           int64
	OrderNumber        string
	LessonType         int
	LessonChargingMode int
	TotalQuantity      float64
	FreeQuantity       float64
	UsedQuantity       float64
	RemainingQuantity  float64
	TotalTuition       float64
	UsedTuition        float64
	RemainingTuition   float64
	ConfirmedTuition   float64
	Status             int
	ProductName        string
}

type rollCallExistingScheduleRecordSummary struct {
	TeachingRecordID int64
	StudentIDSet     map[int64]struct{}
}

type rollCallConfirmOptions struct {
	RecordTime     *time.Time
	OperatorName   string
	IsAutoRollCall bool
}

func (repo *Repository) CheckRollCallTeachingRecordByTeacherAndTime(ctx context.Context, instID int64, dto model.RollCallCheckTeachingRecordByTeacherAndTimeDTO) error {
	teacherID, err := strconv.ParseInt(strings.TrimSpace(dto.TeacherID), 10, 64)
	if err != nil || teacherID <= 0 {
		return nil
	}
	excludeScheduleID, _ := strconv.ParseInt(strings.TrimSpace(dto.TimetableSourceID), 10, 64)
	excludeTeachingRecordID, _ := strconv.ParseInt(strings.TrimSpace(dto.ExcludeTeachingRecordID), 10, 64)
	startTime, err := parseRollCallConfirmDateTime(dto.StartTime)
	if err != nil {
		return err
	}
	endTime, err := parseRollCallConfirmDateTime(dto.EndTime)
	if err != nil {
		return err
	}
	if !endTime.After(startTime) {
		return errors.New("上课时间不合法")
	}

	var count int
	query := `
		SELECT COUNT(*)
		FROM student_teaching_record
		WHERE inst_id = ?
		  AND del_flag = 0
		  AND main_teacher_id = ?
		  AND start_time < ?
		  AND end_time > ?
	`
	args := []any{instID, teacherID, endTime, startTime}
	if excludeScheduleID > 0 {
		query += " AND teaching_schedule_id <> ?"
		args = append(args, excludeScheduleID)
	}
	if excludeTeachingRecordID > 0 {
		query += " AND teaching_record_id <> ?"
		args = append(args, excludeTeachingRecordID)
	}
	if err := repo.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return errors.New("上课教师在该时间段已有上课记录")
	}
	return nil
}

func (repo *Repository) BatchEstimateRollCallSufficientTuitionAccount(ctx context.Context, instID int64, dto model.RollCallBatchEstimateSufficientTuitionAccountDTO) (model.RollCallBatchEstimateSufficientTuitionAccountResult, error) {
	accountIDs := make([]int64, 0, len(dto.TuitionInfoList))
	for _, item := range dto.TuitionInfoList {
		accountID, err := strconv.ParseInt(strings.TrimSpace(item.TuitionAccountID), 10, 64)
		if err != nil || accountID <= 0 {
			continue
		}
		accountIDs = append(accountIDs, accountID)
	}
	accountMap, err := repo.loadRollCallConfirmAccountMap(ctx, instID, uniquePositiveInt64s(accountIDs))
	if err != nil {
		return model.RollCallBatchEstimateSufficientTuitionAccountResult{}, err
	}

	result := model.RollCallBatchEstimateSufficientTuitionAccountResult{
		TuitionInfoList: make([]model.RollCallEstimateTuitionResultItem, 0, len(dto.TuitionInfoList)),
	}
	for _, item := range dto.TuitionInfoList {
		accountID := strings.TrimSpace(item.TuitionAccountID)
		isSufficient := true
		if accountID != "" && accountID != "0" && item.Quantity > 0 {
			account, ok := accountMap[accountID]
			if !ok {
				isSufficient = false
			} else {
				mode := normalizeRollCallDrawerChargingMode(account.LessonChargingMode)
				if mode == 1 {
					isSufficient = account.Status == model.TuitionAccountStatusActive && account.RemainingQuantity+0.000001 >= item.Quantity
				}
			}
		}
		result.TuitionInfoList = append(result.TuitionInfoList, model.RollCallEstimateTuitionResultItem{
			TuitionAccountID: accountID,
			IsSufficient:     isSufficient,
		})
	}
	return result, nil
}

func (repo *Repository) ConfirmRollCall(ctx context.Context, instID, operatorID int64, dto model.RollCallConfirmDTO) (model.RollCallConfirmResult, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return model.RollCallConfirmResult{}, err
	}
	defer tx.Rollback()

	result, err := repo.confirmRollCallTx(ctx, tx, instID, operatorID, dto, rollCallConfirmOptions{})
	if err != nil {
		return model.RollCallConfirmResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return model.RollCallConfirmResult{}, err
	}
	return result, nil
}

func (repo *Repository) confirmRollCallTx(ctx context.Context, tx *sql.Tx, instID, operatorID int64, dto model.RollCallConfirmDTO, options rollCallConfirmOptions) (model.RollCallConfirmResult, error) {
	startTime, err := parseRollCallConfirmDateTime(dto.StartTime)
	if err != nil {
		return model.RollCallConfirmResult{}, err
	}
	endTime, err := parseRollCallConfirmDateTime(dto.EndTime)
	if err != nil {
		return model.RollCallConfirmResult{}, err
	}
	if !endTime.After(startTime) {
		return model.RollCallConfirmResult{}, errors.New("上课时间不合法")
	}
	if len(dto.StudentList) == 0 {
		return model.RollCallConfirmResult{}, errors.New("暂无可提交的点名学员")
	}
	scheduleID, _ := strconv.ParseInt(strings.TrimSpace(dto.TimetableSourceID), 10, 64)
	if scheduleID <= 0 {
		return repo.confirmGroupClassUnscheduledRollCallTx(ctx, tx, instID, operatorID, dto, startTime, endTime, options)
	}

	detail, classMeta, err := repo.loadRollCallDrawerContext(ctx, instID, dto.TimetableSourceID)
	if err != nil {
		return model.RollCallConfirmResult{}, err
	}
	if detail.CallStatus == 2 {
		return model.RollCallConfirmResult{}, errors.New("当前日程已完成点名")
	}
	if !detail.CanRollCall && strings.TrimSpace(detail.RollCallDisabledReason) != "" {
		return model.RollCallConfirmResult{}, errors.New(strings.TrimSpace(detail.RollCallDisabledReason))
	}

	allowedStudentIDs := make(map[string]struct{}, len(detail.Students)+len(detail.LeaveStudents))
	for _, item := range detail.Students {
		allowedStudentIDs[strings.TrimSpace(item.StudentID)] = struct{}{}
	}
	for _, item := range detail.LeaveStudents {
		allowedStudentIDs[strings.TrimSpace(item.StudentID)] = struct{}{}
	}

	studentIDs := make([]int64, 0, len(dto.StudentList))
	accountIDs := make([]int64, 0, len(dto.StudentList))
	for _, item := range dto.StudentList {
		studentIDText := strings.TrimSpace(item.StudentID)
		if _, ok := allowedStudentIDs[studentIDText]; !ok {
			return model.RollCallConfirmResult{}, errors.New("当前学员不在本节日程中")
		}
		studentID, err := strconv.ParseInt(studentIDText, 10, 64)
		if err != nil || studentID <= 0 {
			return model.RollCallConfirmResult{}, errors.New("学员信息无效")
		}
		studentIDs = append(studentIDs, studentID)
		accountID, err := strconv.ParseInt(strings.TrimSpace(item.TuitionAccountID), 10, 64)
		if err == nil && accountID > 0 {
			accountIDs = append(accountIDs, accountID)
		}
	}

	profileMap, err := repo.loadRollCallConfirmStudentProfileMap(ctx, instID, uniquePositiveInt64s(studentIDs))
	if err != nil {
		return model.RollCallConfirmResult{}, err
	}
	accountMap, err := repo.loadRollCallConfirmAccountMap(ctx, instID, uniquePositiveInt64s(accountIDs))
	if err != nil {
		return model.RollCallConfirmResult{}, err
	}

	operatorName := strings.TrimSpace(options.OperatorName)
	if operatorName == "" {
		operatorName = repo.GetStaffNameByID(ctx, &operatorID)
	}
	if operatorID <= 0 && operatorName == "" {
		operatorName = "系统自动点名"
	}
	if strings.TrimSpace(operatorName) == "" || strings.HasPrefix(operatorName, "未知(") {
		operatorName = "系统自动点名"
	}

	existingSummary, err := repo.loadRollCallExistingScheduleRecordSummaryTx(ctx, tx, instID, scheduleID)
	if err != nil {
		return model.RollCallConfirmResult{}, err
	}
	if detail.CallStatus == 2 {
		return model.RollCallConfirmResult{}, errors.New("当前日程已完成点名")
	}
	if detail.ClassType != model.TeachingClassTypeNormal && len(existingSummary.StudentIDSet) > 0 {
		return model.RollCallConfirmResult{}, errors.New("当前日程已完成点名")
	}

	pendingStudents := make([]model.RollCallConfirmStudent, 0, len(dto.StudentList))
	for _, item := range dto.StudentList {
		studentID, _ := strconv.ParseInt(strings.TrimSpace(item.StudentID), 10, 64)
		if studentID > 0 {
			if _, exists := existingSummary.StudentIDSet[studentID]; exists {
				continue
			}
		}
		pendingStudents = append(pendingStudents, item)
	}
	if len(pendingStudents) == 0 {
		if existingSummary.TeachingRecordID > 0 {
			return model.RollCallConfirmResult{
				ID:   strconv.FormatInt(existingSummary.TeachingRecordID, 10),
				Name: "",
			}, nil
		}
		return model.RollCallConfirmResult{}, errors.New("暂无可提交的点名学员")
	}

	teachingRecordID := existingSummary.TeachingRecordID
	if teachingRecordID <= 0 {
		teachingRecordID, err = repo.nextRollCallTeachingRecordIDTx(ctx, tx)
		if err != nil {
			return model.RollCallConfirmResult{}, err
		}
	}

	teacherNamesJSON, teacherIDsJSON := rollCallConfirmTeacherJSON(detail.AssistantNames, detail.AssistantIDs)
	teachingContentImagesJSON, err := json.Marshal(dto.TeachingContentImages)
	if err != nil {
		return model.RollCallConfirmResult{}, err
	}
	classTeacherNamesJSON, currentClassTeacherNamesJSON, oneToOneTeacherNamesJSON := rollCallConfirmClassTeacherJSON(detail, classMeta)
	emptyIDsJSON, _ := json.Marshal([]string{})
	recordTime := startTime
	if options.RecordTime != nil && !options.RecordTime.IsZero() {
		recordTime = *options.RecordTime
	}

	insertedCount := 0
	for _, item := range pendingStudents {
		studentID, _ := strconv.ParseInt(strings.TrimSpace(item.StudentID), 10, 64)
		profile := profileMap[studentID]
		account, hasAccount := accountMap[strings.TrimSpace(item.TuitionAccountID)]

		status := normalizeRollCallConfirmStudentStatus(item.Status)
		quantity := normalizeRollCallConfirmQuantity(item)
		if status != 1 {
			quantity = 0
		}

		actualQuantity := quantity
		actualDeduct := 0.0
		actualTuition := 0.0
		arrearQuantity := 0.0
		tuitionAccountName := ""
		if hasAccount {
			tuitionAccountName = firstNonEmptyString(strings.TrimSpace(account.ProductName), strings.TrimSpace(detail.LessonName))
		}

		if normalizeRollCallDrawerChargingMode(item.SkuMode) == 1 && quantity > 0 {
			if hasAccount && account.Status == model.TuitionAccountStatusActive {
				canCoverQuantity := math.Max(account.RemainingQuantity, 0)+0.000001 >= quantity
				if !canCoverQuantity && options.IsAutoRollCall {
					actualQuantity = 0
					actualDeduct = 0
					actualTuition = 0
					arrearQuantity = quantity
				} else {
					actualDeduct = math.Min(quantity, math.Max(account.RemainingQuantity, 0))
					arrearQuantity = roundMoney(math.Max(quantity-actualDeduct, 0))
				}
				if actualDeduct > 0 {
					actualTuition = repo.rollCallConfirmLessonHourTuition(actualDeduct, account)
				}
				if options.IsAutoRollCall && arrearQuantity > 0 {
					actualQuantity = 0
				}
			} else {
				arrearQuantity = quantity
				if options.IsAutoRollCall {
					actualQuantity = 0
				}
			}
		}

		recordClassName := ""
		recordOneToOneName := ""
		if detail.ClassType == model.TeachingClassTypeOneToOne {
			recordOneToOneName = classMeta.ClassName
		} else {
			recordClassName = classMeta.ClassName
		}
		result, err := tx.ExecContext(ctx, `
			INSERT INTO student_teaching_record (
				inst_id, teaching_record_id, teaching_schedule_id, timetable_source_type, timetable_source_id,
				student_id, student_name, student_phone, avatar_url, source_type, current_student_status, status, is_late,
				class_id, class_name, one_to_one_id, one_to_one_name, lesson_id, lesson_name, subject_id, subject_name,
				teaching_content, teaching_content_images_json, classroom_id, classroom_name, main_teacher_id, main_teacher_name,
				teacher_employee_type, assistant_teacher_ids_json, assistant_teacher_names_json, class_teacher_ids_json, class_teacher_names_json,
				roll_call_class_teacher_ids_json, roll_call_class_teacher_names_json, current_class_teacher_ids_json, current_class_teacher_names_json,
				one2one_teacher_ids_json, one2one_teacher_names_json, tuition_account_id, tuition_account_name, sku_mode, quantity, actual_quantity,
				amount, actual_deduct, actual_tuition, arrear_quantity, teacher_class_time, remark, external_remark, is_auto_roll_call, has_compensated,
				advisor_staff_id, advisor_staff_name, student_manager_id, student_manager_name, start_time, end_time, teaching_record_created_time,
				record_time, updated_staff_id, updated_staff_name, updated_time, create_id, create_time, update_id, update_time, del_flag
			) VALUES (
				?, ?, ?, ?, ?,
				?, ?, ?, ?, ?, ?, ?, ?,
				?, ?, ?, ?, ?, ?, ?, ?,
				?, ?, ?, ?, ?, ?,
				?, ?, ?, ?, ?,
				?, ?, ?, ?,
				?, ?, ?, ?, ?, ?, ?, ?,
				?, ?, ?, ?, ?, ?, ?, ?,
				?, ?, ?, ?, ?, ?, ?,
				?, ?, ?, NOW(), ?, NOW(), ?, NOW(), 0
			)
		`,
			instID, teachingRecordID, scheduleID, dto.TimetableSourceType, scheduleID,
			studentID, firstNonEmptyString(strings.TrimSpace(item.StudentName), profile.StudentName), profile.StudentPhone, firstNonEmptyString(profile.AvatarURL, defaultStudentAvatarURL()),
			item.SourceType, profile.StudentStatus, status, false,
			parseRollCallConfirmInt64(dto.SourceID), recordClassName, parseRollCallConfirmOneToOneID(detail, dto.SourceID), recordOneToOneName,
			parseRollCallConfirmInt64(detail.LessonID), detail.LessonName, parseRollCallConfirmInt64(dto.SubjectID), "",
			strings.TrimSpace(dto.TeachingContent), teachingContentImagesJSON, parseRollCallConfirmInt64(dto.ClassRoomID), strings.TrimSpace(detail.ClassroomName),
			parseRollCallConfirmInt64(detail.TeacherID), strings.TrimSpace(detail.TeacherName), 0, teacherIDsJSON, teacherNamesJSON, emptyIDsJSON, classTeacherNamesJSON,
			emptyIDsJSON, classTeacherNamesJSON, emptyIDsJSON, currentClassTeacherNamesJSON, emptyIDsJSON, oneToOneTeacherNamesJSON,
			parseRollCallConfirmInt64(item.TuitionAccountID), tuitionAccountName, normalizeRollCallDrawerChargingMode(item.SkuMode), quantity, actualQuantity,
			roundMoney(item.Amount), actualDeduct, actualTuition, arrearQuantity, dto.TeacherClassTime, strings.TrimSpace(item.Remark), strings.TrimSpace(item.ExternalRemark), options.IsAutoRollCall, false,
			profile.AdvisorStaffID, profile.AdvisorStaffName, profile.StudentManagerID, profile.StudentManagerName, startTime, endTime, recordTime,
			recordTime, operatorID, operatorName, operatorID, operatorID,
		)
		if err != nil {
			return model.RollCallConfirmResult{}, err
		}
		insertedStudentTeachingRecordID, err := result.LastInsertId()
		if err != nil || insertedStudentTeachingRecordID <= 0 {
			return model.RollCallConfirmResult{}, errors.New("点名记录创建失败")
		}
		if actualDeduct > 0 {
			if err := repo.applyRollCallLessonHourConsumeTx(ctx, tx, instID, operatorID, teachingRecordID, insertedStudentTeachingRecordID, actualDeduct, actualTuition, account); err != nil {
				return model.RollCallConfirmResult{}, err
			}
			account.UsedQuantity = roundMoney(account.UsedQuantity + actualDeduct)
			account.RemainingQuantity = roundMoney(math.Max(account.RemainingQuantity-actualDeduct, 0))
			account.UsedTuition = roundMoney(account.UsedTuition + actualTuition)
			account.RemainingTuition = roundMoney(math.Max(account.RemainingTuition-actualTuition, 0))
			account.ConfirmedTuition = roundMoney(account.ConfirmedTuition + actualTuition)
			accountMap[strings.TrimSpace(item.TuitionAccountID)] = account
		}
		insertedCount++
	}

	if insertedCount == 0 && len(existingSummary.StudentIDSet) == 0 {
		return model.RollCallConfirmResult{}, errors.New("暂无可提交的点名学员")
	}

	return model.RollCallConfirmResult{
		ID:   strconv.FormatInt(teachingRecordID, 10),
		Name: "",
	}, nil
}

func (repo *Repository) confirmGroupClassUnscheduledRollCallTx(ctx context.Context, tx *sql.Tx, instID, operatorID int64, dto model.RollCallConfirmDTO, startTime, endTime time.Time, options rollCallConfirmOptions) (model.RollCallConfirmResult, error) {
	if parseRollCallConfirmInt64(dto.SourceID) <= 0 || dto.SourceType != 1 {
		return model.RollCallConfirmResult{}, errors.New("未排课点名仅支持班课")
	}

	classID := parseRollCallConfirmInt64(dto.SourceID)
	classDetail, err := repo.GetGroupClassByID(ctx, instID, classID)
	if err != nil {
		return model.RollCallConfirmResult{}, err
	}
	if classDetail.Status == model.TeachingClassStatusClosed {
		return model.RollCallConfirmResult{}, errors.New("已结班班级不可点名")
	}

	mainTeacherID, assistantIDs := extractRollCallConfirmTeacherIDs(dto.TeacherList, classDetail.DefaultTeacherID)
	classTeachers, _, mainTeacherName, assistantNames, err := repo.buildRollCallTeacherEntriesForIDs(ctx, mainTeacherID, assistantIDs)
	if err != nil {
		return model.RollCallConfirmResult{}, err
	}

	classroomID := strings.TrimSpace(dto.ClassRoomID)
	if classroomID == "" || classroomID == "0" {
		classroomID = strings.TrimSpace(classDetail.ClassroomID)
	}
	classroomName, err := repo.loadRollCallDrawerClassroomName(ctx, instID, classroomID, classDetail.ClassroomName)
	if err != nil {
		return model.RollCallConfirmResult{}, err
	}

	syntheticDetail := model.TeachingScheduleDetailVO{
		ID:                "0",
		ClassType:         model.TeachingClassTypeNormal,
		TeachingClassID:   classDetail.ID,
		TeachingClassName: classDetail.Name,
		LessonID:          classDetail.LessonID,
		LessonName:        classDetail.LessonName,
		TeacherID:         strconv.FormatInt(mainTeacherID, 10),
		TeacherName:       mainTeacherName,
		AssistantIDs:      assistantIDs,
		AssistantNames:    assistantNames,
		ClassroomID:       firstNonEmptyString(classroomID, "0"),
		ClassroomName:     classroomName,
		LessonDate:        startTime.Format("2006-01-02"),
		StartAt:           startTime,
		EndAt:             endTime,
		CanRollCall:       true,
		Remark:            classDetail.Remark,
	}

	membershipsByClassID, err := repo.loadGroupClassStudentMembershipMapWithQueryer(ctx, tx, instID, []int64{classID})
	if err != nil {
		return model.RollCallConfirmResult{}, err
	}
	existingSummary, err := repo.loadRollCallExistingGroupClassUnscheduledRecordSummaryTx(
		ctx,
		tx,
		instID,
		classID,
		classDetail.LessonID,
		startTime,
		endTime,
	)
	if err != nil {
		return model.RollCallConfirmResult{}, err
	}
	roster := buildGroupClassScheduleRosterFromMembershipsAndOverrides(
		membershipsByClassID[classID],
		nil,
		time.Now(),
	)
	allowedStudentIDs := make(map[string]struct{}, len(roster.Active))
	classMemberStudentIDs := make(map[string]struct{}, len(roster.Active))
	for _, item := range roster.Active {
		if item.StudentID <= 0 {
			continue
		}
		studentIDText := strconv.FormatInt(item.StudentID, 10)
		allowedStudentIDs[studentIDText] = struct{}{}
		classMemberStudentIDs[studentIDText] = struct{}{}
	}
	for studentID := range existingSummary.StudentIDSet {
		if studentID <= 0 {
			continue
		}
		allowedStudentIDs[strconv.FormatInt(studentID, 10)] = struct{}{}
	}

	studentIDs := make([]int64, 0, len(dto.StudentList))
	accountIDs := make([]int64, 0, len(dto.StudentList))
	for _, item := range dto.StudentList {
		studentIDText := strings.TrimSpace(item.StudentID)
		_, inAllowedStudentIDs := allowedStudentIDs[studentIDText]
		if !inAllowedStudentIDs && !isRollCallConfirmScheduleOnlySourceType(item.SourceType) {
			return model.RollCallConfirmResult{}, errors.New("当前学员不在本次点名班级中")
		}
		if inAllowedStudentIDs && isRollCallConfirmScheduleOnlySourceType(item.SourceType) {
			studentID, _ := strconv.ParseInt(studentIDText, 10, 64)
			if _, exists := existingSummary.StudentIDSet[studentID]; !exists {
				if _, ok := classMemberStudentIDs[studentIDText]; ok && item.SourceType == 4 {
					return model.RollCallConfirmResult{}, errors.New("班级在班学员不可作为试听学员添加")
				}
				return model.RollCallConfirmResult{}, errors.New("该学员已在本次点名名单中")
			}
		}
		studentID, err := strconv.ParseInt(studentIDText, 10, 64)
		if err != nil || studentID <= 0 {
			return model.RollCallConfirmResult{}, errors.New("学员信息无效")
		}
		studentIDs = append(studentIDs, studentID)
		accountID, err := strconv.ParseInt(strings.TrimSpace(item.TuitionAccountID), 10, 64)
		if err == nil && accountID > 0 {
			accountIDs = append(accountIDs, accountID)
		}
	}

	profileMap, err := repo.loadRollCallConfirmStudentProfileMap(ctx, instID, uniquePositiveInt64s(studentIDs))
	if err != nil {
		return model.RollCallConfirmResult{}, err
	}
	accountMap, err := repo.loadRollCallConfirmAccountMap(ctx, instID, uniquePositiveInt64s(accountIDs))
	if err != nil {
		return model.RollCallConfirmResult{}, err
	}

	operatorName := strings.TrimSpace(options.OperatorName)
	if operatorName == "" {
		operatorName = repo.GetStaffNameByID(ctx, &operatorID)
	}
	if operatorID <= 0 && operatorName == "" {
		operatorName = "系统自动点名"
	}
	if strings.TrimSpace(operatorName) == "" || strings.HasPrefix(operatorName, "未知(") {
		operatorName = "系统自动点名"
	}

	pendingStudents := make([]model.RollCallConfirmStudent, 0, len(dto.StudentList))
	for _, item := range dto.StudentList {
		studentID, _ := strconv.ParseInt(strings.TrimSpace(item.StudentID), 10, 64)
		if studentID > 0 {
			if _, exists := existingSummary.StudentIDSet[studentID]; exists {
				continue
			}
		}
		pendingStudents = append(pendingStudents, item)
	}
	if len(pendingStudents) == 0 {
		if existingSummary.TeachingRecordID > 0 {
			return model.RollCallConfirmResult{
				ID:   strconv.FormatInt(existingSummary.TeachingRecordID, 10),
				Name: "",
			}, nil
		}
		return model.RollCallConfirmResult{}, errors.New("暂无可提交的点名学员")
	}

	teachingRecordID := existingSummary.TeachingRecordID
	if teachingRecordID <= 0 {
		teachingRecordID, err = repo.nextRollCallTeachingRecordIDTx(ctx, tx)
		if err != nil {
			return model.RollCallConfirmResult{}, err
		}
	}

	classMeta := rollCallDrawerContext{
		ClassID:                    classDetail.ID,
		ClassName:                  classDetail.Name,
		DefaultStudentClassTime:    classDetail.DefaultStudentClassTime,
		DefaultTeacherClassTime:    classDetail.DefaultTeacherClassTime,
		DefaultClassTimeRecordMode: classDetail.DefaultClassTimeRecordMode,
		LessonPrice:                classDetail.LessonPrice,
		LessonType:                 classDetail.LessonType,
		Teachers:                   classTeachers,
	}

	teacherNamesJSON, teacherIDsJSON := rollCallConfirmTeacherJSON(syntheticDetail.AssistantNames, syntheticDetail.AssistantIDs)
	teachingContentImagesJSON, err := json.Marshal(dto.TeachingContentImages)
	if err != nil {
		return model.RollCallConfirmResult{}, err
	}
	classTeacherNamesJSON, currentClassTeacherNamesJSON, oneToOneTeacherNamesJSON := rollCallConfirmClassTeacherJSON(syntheticDetail, classMeta)
	emptyIDsJSON, _ := json.Marshal([]string{})
	recordTime := startTime
	if options.RecordTime != nil && !options.RecordTime.IsZero() {
		recordTime = *options.RecordTime
	}

	insertedCount := 0
	for _, item := range pendingStudents {
		studentID, _ := strconv.ParseInt(strings.TrimSpace(item.StudentID), 10, 64)
		profile := profileMap[studentID]
		account, hasAccount := accountMap[strings.TrimSpace(item.TuitionAccountID)]

		status := normalizeRollCallConfirmStudentStatus(item.Status)
		quantity := normalizeRollCallConfirmQuantity(item)
		if status != 1 {
			quantity = 0
		}

		actualQuantity := quantity
		actualDeduct := 0.0
		actualTuition := 0.0
		arrearQuantity := 0.0
		tuitionAccountName := ""
		if hasAccount {
			tuitionAccountName = firstNonEmptyString(strings.TrimSpace(account.ProductName), strings.TrimSpace(syntheticDetail.LessonName))
		}

		if normalizeRollCallDrawerChargingMode(item.SkuMode) == 1 && quantity > 0 {
			if hasAccount && account.Status == model.TuitionAccountStatusActive {
				canCoverQuantity := math.Max(account.RemainingQuantity, 0)+0.000001 >= quantity
				if !canCoverQuantity && options.IsAutoRollCall {
					actualQuantity = 0
					actualDeduct = 0
					actualTuition = 0
					arrearQuantity = quantity
				} else {
					actualDeduct = math.Min(quantity, math.Max(account.RemainingQuantity, 0))
					arrearQuantity = roundMoney(math.Max(quantity-actualDeduct, 0))
				}
				if actualDeduct > 0 {
					actualTuition = repo.rollCallConfirmLessonHourTuition(actualDeduct, account)
				}
				if options.IsAutoRollCall && arrearQuantity > 0 {
					actualQuantity = 0
				}
			} else {
				arrearQuantity = quantity
				if options.IsAutoRollCall {
					actualQuantity = 0
				}
			}
		}

		result, err := tx.ExecContext(ctx, `
			INSERT INTO student_teaching_record (
				inst_id, teaching_record_id, teaching_schedule_id, timetable_source_type, timetable_source_id,
				student_id, student_name, student_phone, avatar_url, source_type, current_student_status, status, is_late,
				class_id, class_name, one_to_one_id, one_to_one_name, lesson_id, lesson_name, subject_id, subject_name,
				teaching_content, teaching_content_images_json, classroom_id, classroom_name, main_teacher_id, main_teacher_name,
				teacher_employee_type, assistant_teacher_ids_json, assistant_teacher_names_json, class_teacher_ids_json, class_teacher_names_json,
				roll_call_class_teacher_ids_json, roll_call_class_teacher_names_json, current_class_teacher_ids_json, current_class_teacher_names_json,
				one2one_teacher_ids_json, one2one_teacher_names_json, tuition_account_id, tuition_account_name, sku_mode, quantity, actual_quantity,
				amount, actual_deduct, actual_tuition, arrear_quantity, teacher_class_time, remark, external_remark, is_auto_roll_call, has_compensated,
				advisor_staff_id, advisor_staff_name, student_manager_id, student_manager_name, start_time, end_time, teaching_record_created_time,
				record_time, updated_staff_id, updated_staff_name, updated_time, create_id, create_time, update_id, update_time, del_flag
			) VALUES (
				?, ?, ?, ?, ?,
				?, ?, ?, ?, ?, ?, ?, ?,
				?, ?, ?, ?, ?, ?, ?, ?,
				?, ?, ?, ?, ?, ?,
				?, ?, ?, ?, ?,
				?, ?, ?, ?,
				?, ?, ?, ?, ?, ?, ?, ?,
				?, ?, ?, ?, ?, ?, ?, ?,
				?, ?, ?, ?, ?, ?, ?,
				?, ?, ?, NOW(), ?, NOW(), ?, NOW(), 0
			)
		`,
			instID, teachingRecordID, 0, dto.TimetableSourceType, 0,
			studentID, firstNonEmptyString(strings.TrimSpace(item.StudentName), profile.StudentName), profile.StudentPhone, firstNonEmptyString(profile.AvatarURL, defaultStudentAvatarURL()),
			item.SourceType, profile.StudentStatus, status, false,
			classID, classMeta.ClassName, 0, "", parseRollCallConfirmInt64(syntheticDetail.LessonID), syntheticDetail.LessonName, parseRollCallConfirmInt64(dto.SubjectID), "",
			strings.TrimSpace(dto.TeachingContent), teachingContentImagesJSON, parseRollCallConfirmInt64(classroomID), classroomName,
			mainTeacherID, mainTeacherName, 0, teacherIDsJSON, teacherNamesJSON, emptyIDsJSON, classTeacherNamesJSON,
			emptyIDsJSON, classTeacherNamesJSON, emptyIDsJSON, currentClassTeacherNamesJSON, emptyIDsJSON, oneToOneTeacherNamesJSON,
			parseRollCallConfirmInt64(item.TuitionAccountID), tuitionAccountName, normalizeRollCallDrawerChargingMode(item.SkuMode), quantity, actualQuantity,
			roundMoney(item.Amount), actualDeduct, actualTuition, arrearQuantity, dto.TeacherClassTime, strings.TrimSpace(item.Remark), strings.TrimSpace(item.ExternalRemark), options.IsAutoRollCall, false,
			profile.AdvisorStaffID, profile.AdvisorStaffName, profile.StudentManagerID, profile.StudentManagerName, startTime, endTime, recordTime,
			recordTime, operatorID, operatorName, operatorID, operatorID,
		)
		if err != nil {
			return model.RollCallConfirmResult{}, err
		}
		insertedStudentTeachingRecordID, err := result.LastInsertId()
		if err != nil || insertedStudentTeachingRecordID <= 0 {
			return model.RollCallConfirmResult{}, errors.New("点名记录创建失败")
		}
		if actualDeduct > 0 {
			if err := repo.applyRollCallLessonHourConsumeTx(ctx, tx, instID, operatorID, teachingRecordID, insertedStudentTeachingRecordID, actualDeduct, actualTuition, account); err != nil {
				return model.RollCallConfirmResult{}, err
			}
			account.UsedQuantity = roundMoney(account.UsedQuantity + actualDeduct)
			account.RemainingQuantity = roundMoney(math.Max(account.RemainingQuantity-actualDeduct, 0))
			account.UsedTuition = roundMoney(account.UsedTuition + actualTuition)
			account.RemainingTuition = roundMoney(math.Max(account.RemainingTuition-actualTuition, 0))
			account.ConfirmedTuition = roundMoney(account.ConfirmedTuition + actualTuition)
			accountMap[strings.TrimSpace(item.TuitionAccountID)] = account
		}
		insertedCount++
	}

	if insertedCount == 0 && len(existingSummary.StudentIDSet) == 0 {
		return model.RollCallConfirmResult{}, errors.New("暂无可提交的点名学员")
	}

	return model.RollCallConfirmResult{
		ID:   strconv.FormatInt(teachingRecordID, 10),
		Name: "",
	}, nil
}

func extractRollCallConfirmTeacherIDs(teachers []model.RollCallConfirmTeacher, fallbackTeacherID string) (int64, []string) {
	mainTeacherID := parseRollCallConfirmInt64(fallbackTeacherID)
	assistantIDs := make([]string, 0, len(teachers))
	seenAssistants := make(map[string]struct{}, len(teachers))
	for _, item := range teachers {
		teacherIDText := strings.TrimSpace(item.TeacherID)
		if teacherIDText == "" || teacherIDText == "0" {
			continue
		}
		if item.Type == 1 && mainTeacherID <= 0 {
			mainTeacherID = parseRollCallConfirmInt64(teacherIDText)
			continue
		}
		if item.Type == 1 {
			mainTeacherID = parseRollCallConfirmInt64(teacherIDText)
			continue
		}
		if _, ok := seenAssistants[teacherIDText]; ok {
			continue
		}
		seenAssistants[teacherIDText] = struct{}{}
		assistantIDs = append(assistantIDs, teacherIDText)
	}
	return mainTeacherID, assistantIDs
}

func (repo *Repository) loadRollCallExistingGroupClassUnscheduledRecordSummaryTx(ctx context.Context, tx *sql.Tx, instID, classID int64, lessonID string, startTime, endTime time.Time) (rollCallExistingScheduleRecordSummary, error) {
	parsedLessonID, err := strconv.ParseInt(strings.TrimSpace(lessonID), 10, 64)
	if err != nil || parsedLessonID <= 0 {
		return rollCallExistingScheduleRecordSummary{
			StudentIDSet: make(map[int64]struct{}),
		}, nil
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT
			IFNULL(student_id, 0),
			IFNULL(teaching_record_id, 0)
		FROM student_teaching_record
		WHERE inst_id = ?
		  AND IFNULL(teaching_schedule_id, 0) = 0
		  AND IFNULL(class_id, 0) = ?
		  AND IFNULL(one_to_one_id, 0) = 0
		  AND IFNULL(lesson_id, 0) = ?
		  AND start_time = ?
		  AND end_time = ?
		  AND del_flag = 0
		ORDER BY id ASC
	`, instID, classID, parsedLessonID, startTime, endTime)
	if err != nil {
		return rollCallExistingScheduleRecordSummary{}, err
	}
	defer rows.Close()

	result := rollCallExistingScheduleRecordSummary{
		StudentIDSet: make(map[int64]struct{}),
	}
	for rows.Next() {
		var studentID int64
		var teachingRecordID int64
		if err := rows.Scan(&studentID, &teachingRecordID); err != nil {
			return rollCallExistingScheduleRecordSummary{}, err
		}
		if result.TeachingRecordID <= 0 && teachingRecordID > 0 {
			result.TeachingRecordID = teachingRecordID
		}
		if studentID > 0 {
			result.StudentIDSet[studentID] = struct{}{}
		}
	}
	if err := rows.Err(); err != nil {
		return rollCallExistingScheduleRecordSummary{}, err
	}
	return result, nil
}

func (repo *Repository) loadRollCallExistingScheduleRecordSummaryTx(ctx context.Context, tx *sql.Tx, instID, scheduleID int64) (rollCallExistingScheduleRecordSummary, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT
			IFNULL(student_id, 0),
			IFNULL(teaching_record_id, 0)
		FROM student_teaching_record
		WHERE inst_id = ?
		  AND teaching_schedule_id = ?
		  AND del_flag = 0
		ORDER BY id ASC
	`, instID, scheduleID)
	if err != nil {
		return rollCallExistingScheduleRecordSummary{}, err
	}
	defer rows.Close()

	result := rollCallExistingScheduleRecordSummary{
		StudentIDSet: make(map[int64]struct{}),
	}
	for rows.Next() {
		var studentID int64
		var teachingRecordID int64
		if err := rows.Scan(&studentID, &teachingRecordID); err != nil {
			return rollCallExistingScheduleRecordSummary{}, err
		}
		if result.TeachingRecordID <= 0 && teachingRecordID > 0 {
			result.TeachingRecordID = teachingRecordID
		}
		if studentID > 0 {
			result.StudentIDSet[studentID] = struct{}{}
		}
	}
	if err := rows.Err(); err != nil {
		return rollCallExistingScheduleRecordSummary{}, err
	}
	return result, nil
}

func (repo *Repository) loadRollCallConfirmStudentProfileMap(ctx context.Context, instID int64, studentIDs []int64) (map[int64]rollCallConfirmStudentProfile, error) {
	if len(studentIDs) == 0 {
		return map[int64]rollCallConfirmStudentProfile{}, nil
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			s.id,
			IFNULL(s.stu_name, ''),
			IFNULL(s.mobile, ''),
			IFNULL(s.avatar_url, ''),
			IFNULL(s.student_status, 0),
			IFNULL(s.advisor_id, 0),
			IFNULL(advisor.nick_name, ''),
			IFNULL(s.student_manager_id, 0),
			IFNULL(manager.nick_name, '')
		FROM inst_student s
		LEFT JOIN inst_user advisor ON advisor.id = s.advisor_id AND advisor.del_flag = 0
		LEFT JOIN inst_user manager ON manager.id = s.student_manager_id AND manager.del_flag = 0
		WHERE s.inst_id = ?
		  AND s.del_flag = 0
		  AND s.id IN (`+sqlPlaceholders(len(studentIDs))+`)
	`, append([]any{instID}, int64SliceToAny(studentIDs)...)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64]rollCallConfirmStudentProfile, len(studentIDs))
	for rows.Next() {
		var item rollCallConfirmStudentProfile
		if err := rows.Scan(
			&item.StudentID,
			&item.StudentName,
			&item.StudentPhone,
			&item.AvatarURL,
			&item.StudentStatus,
			&item.AdvisorStaffID,
			&item.AdvisorStaffName,
			&item.StudentManagerID,
			&item.StudentManagerName,
		); err != nil {
			return nil, err
		}
		item.StudentPhone = maskStudentMobile(strings.TrimSpace(item.StudentPhone))
		result[item.StudentID] = item
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (repo *Repository) loadRollCallConfirmAccountMap(ctx context.Context, instID int64, accountIDs []int64) (map[string]rollCallConfirmAccount, error) {
	if len(accountIDs) == 0 {
		return map[string]rollCallConfirmAccount{}, nil
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			ta.id,
			ta.student_id,
			ta.course_id,
			IFNULL(so.order_number, ''),
			IFNULL(ic.teach_method, 0),
			CASE
				WHEN IFNULL(icq.lesson_model, 0) = 4 THEN 3
				ELSE IFNULL(icq.lesson_model, 0)
			END AS lesson_charging_mode,
			IFNULL(ta.total_quantity, 0),
			IFNULL(ta.free_quantity, 0),
			IFNULL(ta.used_quantity, 0),
			IFNULL(ta.remaining_quantity, 0),
			IFNULL(ta.total_tuition, 0),
			IFNULL(ta.used_tuition, 0),
			IFNULL(ta.remaining_tuition, 0),
			IFNULL(ta.confirmed_tuition, 0),
			IFNULL(ta.status, 0),
			IFNULL(NULLIF(TRIM(ic.name), ''), IFNULL(icq.name, ''))
		FROM tuition_account ta
		INNER JOIN inst_course ic ON ic.id = ta.course_id AND ic.del_flag = 0
		LEFT JOIN sale_order so ON so.id = ta.order_id AND so.del_flag = 0
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

	result := make(map[string]rollCallConfirmAccount, len(accountIDs))
	for rows.Next() {
		var item rollCallConfirmAccount
		if err := rows.Scan(
			&item.ID,
			&item.StudentID,
			&item.CourseID,
			&item.OrderNumber,
			&item.LessonType,
			&item.LessonChargingMode,
			&item.TotalQuantity,
			&item.FreeQuantity,
			&item.UsedQuantity,
			&item.RemainingQuantity,
			&item.TotalTuition,
			&item.UsedTuition,
			&item.RemainingTuition,
			&item.ConfirmedTuition,
			&item.Status,
			&item.ProductName,
		); err != nil {
			return nil, err
		}
		result[strconv.FormatInt(item.ID, 10)] = item
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (repo *Repository) nextRollCallTeachingRecordIDTx(ctx context.Context, tx *sql.Tx) (int64, error) {
	var nextID int64
	if err := tx.QueryRowContext(ctx, `
		SELECT IFNULL(GREATEST(MAX(id), MAX(teaching_record_id)), 0) + 1
		FROM student_teaching_record
	`).Scan(&nextID); err != nil {
		return 0, err
	}
	if nextID <= 0 {
		nextID = 1
	}
	return nextID, nil
}

func (repo *Repository) rollCallConfirmLessonHourTuition(quantity float64, account rollCallConfirmAccount) float64 {
	if quantity <= 0 || account.TotalQuantity <= 0 || account.TotalTuition <= 0 {
		return 0
	}
	return roundMoney(account.TotalTuition * quantity / account.TotalQuantity)
}

func (repo *Repository) applyRollCallLessonHourConsumeTx(ctx context.Context, tx *sql.Tx, instID, operatorID, teachingRecordID, flowSourceID int64, quantity, tuition float64, account rollCallConfirmAccount) error {
	newUsedQuantity := roundMoney(account.UsedQuantity + quantity)
	newUsedTuition := roundMoney(account.UsedTuition + tuition)
	newConfirmedTuition := roundMoney(account.ConfirmedTuition + tuition)
	// 点名扣减后的剩余值统一按“总量 - 已用量”回算，避免历史脏数据继续向后传染。
	newRemainingQuantity := roundMoney(math.Max(account.TotalQuantity-newUsedQuantity, 0))
	newRemainingTuition := roundMoney(math.Max(account.TotalTuition-newUsedTuition, 0))
	if flowSourceID <= 0 {
		flowSourceID = teachingRecordID
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE tuition_account
		SET used_quantity = ?,
		    remaining_quantity = ?,
		    used_tuition = ?,
		    remaining_tuition = ?,
		    confirmed_tuition = ?,
		    update_id = ?,
		    update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, newUsedQuantity, newRemainingQuantity, newUsedTuition, newRemainingTuition, newConfirmedTuition, operatorID, account.ID, instID); err != nil {
		return err
	}

	_, err := tx.ExecContext(ctx, `
		INSERT INTO tuition_account_flow (
			uuid, version, inst_id, tuition_account_id, student_id, product_id, lesson_type, lesson_charging_mode,
			source_type, source_id, teaching_record_id, order_number, created_time, quantity, tuition, balance_quantity, balance_tuition,
			create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, NOW(), ?, ?, ?, ?,
			?, NOW(), ?, NOW(), 0
		)
		ON DUPLICATE KEY UPDATE
			quantity = round(IFNULL(quantity, 0) + VALUES(quantity), 2),
			tuition = round(IFNULL(tuition, 0) + VALUES(tuition), 2),
			balance_quantity = VALUES(balance_quantity),
			balance_tuition = VALUES(balance_tuition),
			order_number = VALUES(order_number),
			teaching_record_id = VALUES(teaching_record_id),
			update_id = VALUES(update_id),
			update_time = VALUES(update_time),
			del_flag = VALUES(del_flag)
	`,
		instID,
		account.ID,
		account.StudentID,
		account.CourseID,
		account.LessonType,
		1,
		model.TuitionAccountFlowSourceConsume,
		flowSourceID,
		teachingRecordID,
		account.OrderNumber,
		quantity,
		tuition,
		newRemainingQuantity,
		newRemainingTuition,
		operatorID,
		operatorID,
	)
	return err
}

func parseRollCallConfirmDateTime(raw string) (time.Time, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return time.Time{}, errors.New("缺少上课时间")
	}
	layouts := []string{
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		time.RFC3339,
	}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, text, time.Local); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, errors.New("上课时间格式不正确")
}

func normalizeRollCallConfirmStudentStatus(status int) int {
	switch status {
	case 2, 3, 4:
		return status
	default:
		return 1
	}
}

func isRollCallConfirmScheduleOnlySourceType(sourceType int) bool {
	switch sourceType {
	case 2, 3, 4:
		return true
	default:
		return false
	}
}

func normalizeRollCallConfirmQuantity(item model.RollCallConfirmStudent) float64 {
	if item.Quantity > 0 {
		return roundMoney(item.Quantity)
	}
	if item.StudentShouldDeduct > 0 {
		return roundMoney(float64(item.StudentShouldDeduct))
	}
	return 0
}

func parseRollCallConfirmInt64(raw string) int64 {
	value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || value <= 0 {
		return 0
	}
	return value
}

func parseRollCallConfirmOneToOneID(detail model.TeachingScheduleDetailVO, sourceID string) int64 {
	if detail.ClassType == model.TeachingClassTypeOneToOne {
		return parseRollCallConfirmInt64(sourceID)
	}
	return 0
}

func rollCallConfirmTeacherJSON(names, ids []string) ([]byte, []byte) {
	nameJSON, _ := json.Marshal(names)
	idJSON, _ := json.Marshal(ids)
	return nameJSON, idJSON
}

func rollCallConfirmClassTeacherJSON(detail model.TeachingScheduleDetailVO, classMeta rollCallDrawerContext) ([]byte, []byte, []byte) {
	classTeacherNames := []string{}
	currentClassTeacherNames := []string{}
	oneToOneTeacherNames := []string{}
	if detail.ClassType == model.TeachingClassTypeOneToOne && strings.TrimSpace(classMeta.ClassName) != "" {
		oneToOneTeacherNames = append(oneToOneTeacherNames, strings.TrimSpace(detail.TeacherName))
	}
	classTeacherJSON, _ := json.Marshal(classTeacherNames)
	currentClassTeacherJSON, _ := json.Marshal(currentClassTeacherNames)
	oneToOneTeacherJSON, _ := json.Marshal(oneToOneTeacherNames)
	return classTeacherJSON, currentClassTeacherJSON, oneToOneTeacherJSON
}

func defaultStudentAvatarURL() string {
	return "https://cdn.schoolpal.cn/schoolpal/next-erp/avator_male.png"
}
