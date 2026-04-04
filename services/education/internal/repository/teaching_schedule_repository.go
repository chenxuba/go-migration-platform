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

func ensureTeachingScheduleTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS teaching_schedule (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			uuid VARCHAR(64) NULL,
			version BIGINT NOT NULL DEFAULT 0,
			inst_id BIGINT NOT NULL,
			class_type INT NOT NULL DEFAULT 0,
			teaching_class_id BIGINT NOT NULL DEFAULT 0,
			teaching_class_name VARCHAR(150) NOT NULL DEFAULT '',
			student_id BIGINT NOT NULL DEFAULT 0,
			student_name VARCHAR(100) NOT NULL DEFAULT '',
			lesson_id BIGINT NOT NULL DEFAULT 0,
			lesson_name VARCHAR(150) NOT NULL DEFAULT '',
			teacher_id BIGINT NOT NULL DEFAULT 0,
			teacher_name VARCHAR(100) NOT NULL DEFAULT '',
			assistant_ids_json JSON NULL,
			assistant_names_json JSON NULL,
			classroom_id BIGINT NOT NULL DEFAULT 0,
			classroom_name VARCHAR(150) NOT NULL DEFAULT '',
			lesson_date DATE NOT NULL,
			lesson_start_at DATETIME NOT NULL,
			lesson_end_at DATETIME NOT NULL,
			batch_no VARCHAR(64) NOT NULL DEFAULT '',
			batch_size INT NOT NULL DEFAULT 1,
			status INT NOT NULL DEFAULT 1,
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT NOT NULL DEFAULT 0,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			KEY idx_teaching_schedule_inst_date (inst_id, lesson_date),
			KEY idx_teaching_schedule_teacher (inst_id, teacher_id, lesson_date),
			KEY idx_teaching_schedule_classroom (inst_id, classroom_id, lesson_date),
			KEY idx_teaching_schedule_batch (inst_id, batch_no)
		)
	`)
	if err != nil {
		return err
	}
	return ensureColumnsOnTable(ctx, db, "teaching_schedule", map[string]string{
		"assistant_ids_json":   "assistant_ids_json JSON NULL AFTER teacher_name",
		"assistant_names_json": "assistant_names_json JSON NULL AFTER assistant_ids_json",
		"classroom_id":         "classroom_id BIGINT NOT NULL DEFAULT 0 AFTER assistant_names_json",
		"classroom_name":       "classroom_name VARCHAR(150) NOT NULL DEFAULT '' AFTER classroom_id",
		"batch_no":             "batch_no VARCHAR(64) NOT NULL DEFAULT '' AFTER lesson_end_at",
		"batch_size":           "batch_size INT NOT NULL DEFAULT 1 AFTER batch_no",
	})
}

func (repo *Repository) GetOneToOneScheduleCreateContextTx(ctx context.Context, tx *sql.Tx, instID, classID int64) (model.OneToOneScheduleCreateContext, error) {
	var item model.OneToOneScheduleCreateContext
	err := tx.QueryRowContext(ctx, `
		SELECT
			tc.id,
			IFNULL(tc.name, ''),
			IFNULL(tcs.student_id, 0),
			IFNULL(s.stu_name, ''),
			IFNULL(tc.course_id, 0),
			IFNULL(c.name, ''),
			IFNULL(tc.status, 0),
			IFNULL(tcs.class_student_status, 0)
		FROM teaching_class tc
		INNER JOIN teaching_class_student tcs ON tcs.teaching_class_id = tc.id AND tcs.inst_id = tc.inst_id AND tcs.del_flag = 0
		LEFT JOIN inst_student s ON s.id = tcs.student_id AND s.inst_id = tcs.inst_id AND s.del_flag = 0
		LEFT JOIN inst_course c ON c.id = tc.course_id AND c.del_flag = 0
		WHERE tc.inst_id = ? AND tc.id = ? AND tc.class_type = ? AND tc.del_flag = 0
		ORDER BY tcs.id ASC
		LIMIT 1
	`, instID, classID, model.TeachingClassTypeOneToOne).Scan(
		&item.ClassID,
		&item.ClassName,
		&item.StudentID,
		&item.StudentName,
		&item.LessonID,
		&item.LessonName,
		&item.Status,
		&item.ClassStudentStatus,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return item, errors.New("1对1不存在")
		}
		return item, err
	}
	return item, nil
}

func (repo *Repository) GetTeachingScheduleConflictDetail(ctx context.Context, instID int64, query model.TeachingScheduleConflictDetailQueryDTO) (model.TeachingScheduleValidationResult, error) {
	scheduleID, err := strconv.ParseInt(strings.TrimSpace(query.ID), 10, 64)
	if err != nil || scheduleID <= 0 {
		return model.TeachingScheduleValidationResult{}, errors.New("缺少有效的日程ID")
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return model.TeachingScheduleValidationResult{}, err
	}
	defer tx.Rollback()

	current, err := repo.loadScheduleConflictDetailByIDTx(ctx, tx, instID, scheduleID)
	if err != nil {
		return model.TeachingScheduleValidationResult{}, err
	}

	slot := normalizedScheduleSlot{
		LessonDate: startOfDay(current.LessonDate),
		StartAt:    current.StartAt,
		EndAt:      current.EndAt,
	}
	excludeIDs := []int64{current.ID}

	teacherConflicts := []scheduleConflictDetailRow{}
	if current.TeacherID > 0 {
		teacherConflicts, err = repo.listScheduleConflictDetailsTx(ctx, tx, instID, "teacher_id", current.TeacherID, []normalizedScheduleSlot{slot}, "", excludeIDs)
		if err != nil {
			return model.TeachingScheduleValidationResult{}, err
		}
	}

	studentConflicts := []scheduleConflictDetailRow{}
	if current.StudentID > 0 {
		studentConflicts, err = repo.listScheduleConflictDetailsTx(ctx, tx, instID, "student_id", current.StudentID, []normalizedScheduleSlot{slot}, "", excludeIDs)
		if err != nil {
			return model.TeachingScheduleValidationResult{}, err
		}
	}

	classroomConflicts := []scheduleConflictDetailRow{}
	if current.ClassroomID > 0 {
		classroomConflicts, err = repo.listScheduleConflictDetailsTx(ctx, tx, instID, "classroom_id", current.ClassroomID, []normalizedScheduleSlot{slot}, "", excludeIDs)
		if err != nil {
			return model.TeachingScheduleValidationResult{}, err
		}
	}

	currentItems, existingItems, conflictTypes := buildScheduleConflictResultFromExisting(current, teacherConflicts, classroomConflicts, studentConflicts)
	if len(conflictTypes) == 0 {
		return model.TeachingScheduleValidationResult{
			Valid:            true,
			Message:          "当前日程暂无冲突",
			CurrentSchedules: currentItems,
		}, nil
	}

	return model.TeachingScheduleValidationResult{
		Valid:             false,
		Message:           buildConflictSummaryMessage(conflictTypes),
		CurrentSchedules:  currentItems,
		ExistingSchedules: existingItems,
		ConflictTypes:     conflictTypes,
	}, nil
}

func (repo *Repository) ListTeachingSchedules(ctx context.Context, instID int64, query model.TeachingScheduleListQueryDTO) ([]model.TeachingScheduleVO, error) {
	filters := []string{"inst_id = ?", "del_flag = 0", "status = ?"}
	args := []any{instID, model.TeachingScheduleStatusActive}
	if strings.TrimSpace(query.StartDate) != "" {
		filters = append(filters, "lesson_date >= ?")
		args = append(args, strings.TrimSpace(query.StartDate))
	}
	if strings.TrimSpace(query.EndDate) != "" {
		filters = append(filters, "lesson_date <= ?")
		args = append(args, strings.TrimSpace(query.EndDate))
	}
	if query.ClassType != nil && *query.ClassType > 0 {
		filters = append(filters, "class_type = ?")
		args = append(args, *query.ClassType)
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			id,
			IFNULL(batch_no, ''),
			IFNULL(batch_size, 1),
			IFNULL(class_type, 0),
			IFNULL(teaching_class_id, 0),
			IFNULL(teaching_class_name, ''),
			IFNULL(student_id, 0),
			IFNULL(student_name, ''),
			IFNULL(lesson_id, 0),
			IFNULL(lesson_name, ''),
			IFNULL(teacher_id, 0),
			IFNULL(teacher_name, ''),
			assistant_ids_json,
			assistant_names_json,
			IFNULL(classroom_id, 0),
			IFNULL(classroom_name, ''),
			lesson_date,
			lesson_start_at,
			lesson_end_at,
			IFNULL(status, 0)
		FROM teaching_schedule
		WHERE `+strings.Join(filters, " AND ")+`
		ORDER BY lesson_start_at ASC, id ASC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.TeachingScheduleVO, 0, 64)
	for rows.Next() {
		var (
			item              model.TeachingScheduleVO
			id                int64
			teachingClassID   int64
			studentID         int64
			lessonID          int64
			teacherID         int64
			classroomID       int64
			lessonDate        time.Time
			assistantIDsRaw   []byte
			assistantNamesRaw []byte
		)
		if err := rows.Scan(
			&id,
			&item.BatchNo,
			&item.BatchSize,
			&item.ClassType,
			&teachingClassID,
			&item.TeachingClassName,
			&studentID,
			&item.StudentName,
			&lessonID,
			&item.LessonName,
			&teacherID,
			&item.TeacherName,
			&assistantIDsRaw,
			&assistantNamesRaw,
			&classroomID,
			&item.ClassroomName,
			&lessonDate,
			&item.StartAt,
			&item.EndAt,
			&item.Status,
		); err != nil {
			return nil, err
		}
		item.ID = strconv.FormatInt(id, 10)
		item.TeachingClassID = strconv.FormatInt(teachingClassID, 10)
		item.StudentID = strconv.FormatInt(studentID, 10)
		item.LessonID = strconv.FormatInt(lessonID, 10)
		item.TeacherID = strconv.FormatInt(teacherID, 10)
		item.ClassroomID = strconv.FormatInt(classroomID, 10)
		if classroomID <= 0 {
			item.ClassroomID = ""
		}
		item.LessonDate = lessonDate.Format("2006-01-02")
		if len(assistantIDsRaw) > 0 {
			_ = json.Unmarshal(assistantIDsRaw, &item.AssistantIDs)
		}
		if len(assistantNamesRaw) > 0 {
			_ = json.Unmarshal(assistantNamesRaw, &item.AssistantNames)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repo *Repository) CreateOneToOneSchedules(ctx context.Context, instID, operatorID int64, dto model.CreateOneToOneSchedulesDTO) (model.CreateOneToOneSchedulesResult, error) {
	classID, err := strconv.ParseInt(strings.TrimSpace(dto.OneToOneID), 10, 64)
	if err != nil || classID <= 0 {
		return model.CreateOneToOneSchedulesResult{}, errors.New("请选择1对1")
	}
	teacherID, err := strconv.ParseInt(strings.TrimSpace(dto.TeacherID), 10, 64)
	if err != nil || teacherID <= 0 {
		return model.CreateOneToOneSchedulesResult{}, errors.New("请选择上课教师")
	}
	if len(dto.Schedules) == 0 {
		return model.CreateOneToOneSchedulesResult{}, errors.New("请至少选择一节日程")
	}
	assistantIDs := parseStringIDs(dto.AssistantIDs)

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return model.CreateOneToOneSchedulesResult{}, err
	}
	defer tx.Rollback()

	base, err := repo.GetOneToOneScheduleCreateContextTx(ctx, tx, instID, classID)
	if err != nil {
		return model.CreateOneToOneSchedulesResult{}, err
	}
	if base.Status != model.TeachingClassStatusActive {
		return model.CreateOneToOneSchedulesResult{}, errors.New("当前1对1已结班，暂不可创建日程")
	}
	if base.ClassStudentStatus != model.TeachingClassStudentStatusStudying {
		return model.CreateOneToOneSchedulesResult{}, errors.New("当前1对1学员状态不允许创建日程")
	}

	if n, err := repo.CountInstUsersByIDs(ctx, instID, []int64{teacherID}); err != nil || n != 1 {
		if err != nil {
			return model.CreateOneToOneSchedulesResult{}, err
		}
		return model.CreateOneToOneSchedulesResult{}, errors.New("上课教师无效")
	}
	if len(assistantIDs) > 0 {
		if n, err := repo.CountInstUsersByIDs(ctx, instID, assistantIDs); err != nil || n != len(assistantIDs) {
			if err != nil {
				return model.CreateOneToOneSchedulesResult{}, err
			}
			return model.CreateOneToOneSchedulesResult{}, errors.New("存在无效的上课助教")
		}
	}

	classroomID, classroomName, _, err := repo.resolveClassroomByIDTx(ctx, tx, instID, dto.ClassroomID)
	if err != nil {
		return model.CreateOneToOneSchedulesResult{}, err
	}

	teacherName := repo.GetStaffNameByID(ctx, &teacherID)
	assistantNames := make([]string, 0, len(assistantIDs))
	for _, id := range assistantIDs {
		copyID := id
		name := strings.TrimSpace(repo.GetStaffNameByID(ctx, &copyID))
		if name != "" && name != "-" {
			assistantNames = append(assistantNames, name)
		}
	}

	normalized, err := normalizeCreateScheduleSlots(dto.Schedules)
	if err != nil {
		return model.CreateOneToOneSchedulesResult{}, err
	}
	if err := repo.validateTeachingScheduleConflictsTx(ctx, tx, instID, teacherID, classroomID, normalized, "", nil); err != nil {
		return model.CreateOneToOneSchedulesResult{}, err
	}
	studentConflicts, err := repo.listScheduleConflictDetailsTx(ctx, tx, instID, "student_id", base.StudentID, normalized, "", nil)
	if err != nil {
		return model.CreateOneToOneSchedulesResult{}, err
	}
	if len(studentConflicts) > 0 && !dto.AllowStudentConflict {
		return model.CreateOneToOneSchedulesResult{}, errors.New(buildConflictSummaryMessage([]string{"学员"}))
	}

	batchNo := ""
	if len(normalized) > 1 {
		batchNo = fmt.Sprintf("BATCH-%d", time.Now().UnixNano())
	}

	assistantIDsJSON, _ := json.Marshal(stringIDsFromInt64(assistantIDs))
	assistantNamesJSON, _ := json.Marshal(assistantNames)
	result := model.CreateOneToOneSchedulesResult{
		BatchNo: batchNo,
		Count:   len(normalized),
		List:    make([]model.TeachingScheduleVO, 0, len(normalized)),
	}

	for _, slot := range normalized {
		res, err := tx.ExecContext(ctx, `
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
			model.TeachingClassTypeOneToOne,
			base.ClassID,
			base.ClassName,
			base.StudentID,
			base.StudentName,
			base.LessonID,
			base.LessonName,
			teacherID,
			teacherName,
			nullJSONBytes(assistantIDsJSON),
			nullJSONBytes(assistantNamesJSON),
			classroomID,
			classroomName,
			slot.LessonDate.Format("2006-01-02"),
			slot.StartAt,
			slot.EndAt,
			batchNo,
			len(normalized),
			model.TeachingScheduleStatusActive,
			operatorID,
			operatorID,
		)
		if err != nil {
			return model.CreateOneToOneSchedulesResult{}, err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return model.CreateOneToOneSchedulesResult{}, err
		}
		result.List = append(result.List, model.TeachingScheduleVO{
			ID:                strconv.FormatInt(id, 10),
			BatchNo:           batchNo,
			BatchSize:         len(normalized),
			ClassType:         model.TeachingClassTypeOneToOne,
			TeachingClassID:   strconv.FormatInt(base.ClassID, 10),
			TeachingClassName: base.ClassName,
			StudentID:         strconv.FormatInt(base.StudentID, 10),
			StudentName:       base.StudentName,
			LessonID:          strconv.FormatInt(base.LessonID, 10),
			LessonName:        base.LessonName,
			TeacherID:         strconv.FormatInt(teacherID, 10),
			TeacherName:       teacherName,
			AssistantIDs:      stringIDsFromInt64(assistantIDs),
			AssistantNames:    assistantNames,
			ClassroomID:       emptyStringIfZero(classroomID),
			ClassroomName:     classroomName,
			LessonDate:        slot.LessonDate.Format("2006-01-02"),
			StartAt:           slot.StartAt,
			EndAt:             slot.EndAt,
			Status:            model.TeachingScheduleStatusActive,
		})
	}

	if err := tx.Commit(); err != nil {
		return model.CreateOneToOneSchedulesResult{}, err
	}
	return result, nil
}

func (repo *Repository) ValidateOneToOneSchedules(ctx context.Context, instID int64, dto model.CreateOneToOneSchedulesDTO) (model.TeachingScheduleValidationResult, error) {
	classID, err := strconv.ParseInt(strings.TrimSpace(dto.OneToOneID), 10, 64)
	if err != nil || classID <= 0 {
		return model.TeachingScheduleValidationResult{}, errors.New("请选择1对1")
	}
	teacherID, err := strconv.ParseInt(strings.TrimSpace(dto.TeacherID), 10, 64)
	if err != nil || teacherID <= 0 {
		return model.TeachingScheduleValidationResult{}, errors.New("请选择上课教师")
	}
	if len(dto.Schedules) == 0 {
		return model.TeachingScheduleValidationResult{}, errors.New("请至少选择一节日程")
	}
	assistantIDs := parseStringIDs(dto.AssistantIDs)

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return model.TeachingScheduleValidationResult{}, err
	}
	defer tx.Rollback()

	base, err := repo.GetOneToOneScheduleCreateContextTx(ctx, tx, instID, classID)
	if err != nil {
		return model.TeachingScheduleValidationResult{}, err
	}
	if base.Status != model.TeachingClassStatusActive {
		return model.TeachingScheduleValidationResult{}, errors.New("当前1对1已结班，暂不可创建日程")
	}
	if base.ClassStudentStatus != model.TeachingClassStudentStatusStudying {
		return model.TeachingScheduleValidationResult{}, errors.New("当前1对1学员状态不允许创建日程")
	}

	if n, err := repo.CountInstUsersByIDs(ctx, instID, []int64{teacherID}); err != nil || n != 1 {
		if err != nil {
			return model.TeachingScheduleValidationResult{}, err
		}
		return model.TeachingScheduleValidationResult{}, errors.New("上课教师无效")
	}
	if len(assistantIDs) > 0 {
		if n, err := repo.CountInstUsersByIDs(ctx, instID, assistantIDs); err != nil || n != len(assistantIDs) {
			if err != nil {
				return model.TeachingScheduleValidationResult{}, err
			}
			return model.TeachingScheduleValidationResult{}, errors.New("存在无效的上课助教")
		}
	}

	classroomID, classroomName, _, err := repo.resolveClassroomByIDTx(ctx, tx, instID, dto.ClassroomID)
	if err != nil {
		return model.TeachingScheduleValidationResult{}, err
	}

	normalized, err := normalizeCreateScheduleSlots(dto.Schedules)
	if err != nil {
		return model.TeachingScheduleValidationResult{}, err
	}
	teacherName := strings.TrimSpace(repo.GetStaffNameByID(ctx, &teacherID))
	teacherConflicts, err := repo.listScheduleConflictDetailsTx(ctx, tx, instID, "teacher_id", teacherID, normalized, "", nil)
	if err != nil {
		return model.TeachingScheduleValidationResult{}, err
	}
	studentConflicts, err := repo.listScheduleConflictDetailsTx(ctx, tx, instID, "student_id", base.StudentID, normalized, "", nil)
	if err != nil {
		return model.TeachingScheduleValidationResult{}, err
	}
	classroomConflicts := []scheduleConflictDetailRow{}
	if classroomID > 0 {
		classroomConflicts, err = repo.listScheduleConflictDetailsTx(ctx, tx, instID, "classroom_id", classroomID, normalized, "", nil)
		if err != nil {
			return model.TeachingScheduleValidationResult{}, err
		}
	}
	if len(teacherConflicts) > 0 || len(classroomConflicts) > 0 || len(studentConflicts) > 0 {
		currentItems, existingItems, conflictTypes := buildScheduleConflictResult(base, teacherName, classroomName, normalized, teacherConflicts, classroomConflicts, studentConflicts)
		return model.TeachingScheduleValidationResult{
			Valid:             false,
			Message:           buildConflictSummaryMessage(conflictTypes),
			CurrentSchedules:  currentItems,
			ExistingSchedules: existingItems,
			ConflictTypes:     conflictTypes,
		}, nil
	}
	return model.TeachingScheduleValidationResult{Valid: true}, nil
}

func (repo *Repository) CheckOneToOneScheduleAvailability(ctx context.Context, instID int64, dto model.CheckOneToOneScheduleAvailabilityDTO) (model.OneToOneScheduleAvailabilityResult, error) {
	classID, err := strconv.ParseInt(strings.TrimSpace(dto.OneToOneID), 10, 64)
	if err != nil || classID <= 0 {
		return model.OneToOneScheduleAvailabilityResult{}, errors.New("请选择1对1")
	}
	if len(dto.Schedules) == 0 {
		return model.OneToOneScheduleAvailabilityResult{}, errors.New("请至少选择一个空位")
	}
	if len(dto.Schedules) > 2000 {
		return model.OneToOneScheduleAvailabilityResult{}, errors.New("待检测空位过多，请缩小时间范围后重试")
	}

	normalized, err := normalizeAvailabilityScheduleSlots(dto.Schedules)
	if err != nil {
		return model.OneToOneScheduleAvailabilityResult{}, err
	}
	if len(normalized) == 0 {
		return model.OneToOneScheduleAvailabilityResult{}, errors.New("请至少选择一个空位")
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return model.OneToOneScheduleAvailabilityResult{}, err
	}
	defer tx.Rollback()

	base, err := repo.GetOneToOneScheduleCreateContextTx(ctx, tx, instID, classID)
	if err != nil {
		return model.OneToOneScheduleAvailabilityResult{}, err
	}
	if base.Status != model.TeachingClassStatusActive {
		return buildUnavailableAvailabilityResult(normalized, "当前1对1已结班，暂不可排课"), nil
	}
	if base.ClassStudentStatus != model.TeachingClassStudentStatusStudying {
		return buildUnavailableAvailabilityResult(normalized, "当前1对1学员状态不允许排课"), nil
	}

	startDate, endDate := availabilityDateRange(normalized)
	teacherIDs := collectAvailabilityTeacherIDs(normalized)

	studentConflicts, err := repo.listAvailabilityConflictsByStudentTx(ctx, tx, instID, base.StudentID, startDate, endDate)
	if err != nil {
		return model.OneToOneScheduleAvailabilityResult{}, err
	}
	teacherConflicts, err := repo.listAvailabilityConflictsByTeachersTx(ctx, tx, instID, teacherIDs, startDate, endDate)
	if err != nil {
		return model.OneToOneScheduleAvailabilityResult{}, err
	}

	teacherConflictMap := make(map[int64][]scheduleAvailabilityConflictRow, len(teacherIDs))
	for _, row := range teacherConflicts {
		teacherConflictMap[row.TeacherID] = append(teacherConflictMap[row.TeacherID], row)
	}

	result := model.OneToOneScheduleAvailabilityResult{
		Items: make([]model.OneToOneScheduleAvailabilityItem, 0, len(normalized)),
	}
	for _, slot := range normalized {
		existingMap := make(map[int64]model.TeachingScheduleConflictItem)
		typeSet := make(map[string]struct{}, 2)

		for _, row := range studentConflicts {
			if availabilitySlotsOverlap(slot, row.LessonDate, row.StartAt, row.EndAt) {
				appendAvailabilityConflict(existingMap, row, "学员")
				typeSet["学员"] = struct{}{}
			}
		}
		for _, row := range teacherConflictMap[slot.TeacherID] {
			if availabilitySlotsOverlap(slot, row.LessonDate, row.StartAt, row.EndAt) {
				appendAvailabilityConflict(existingMap, row, "老师")
				typeSet["老师"] = struct{}{}
			}
		}

		conflictTypes := make([]string, 0, len(typeSet))
		for key := range typeSet {
			conflictTypes = append(conflictTypes, key)
		}
		sort.Strings(conflictTypes)

		existingSchedules := make([]model.TeachingScheduleConflictItem, 0, len(existingMap))
		for _, item := range existingMap {
			sort.Strings(item.ConflictTypes)
			existingSchedules = append(existingSchedules, item)
		}
		sort.Slice(existingSchedules, func(i, j int) bool {
			if existingSchedules[i].Date == existingSchedules[j].Date {
				return existingSchedules[i].TimeText < existingSchedules[j].TimeText
			}
			return existingSchedules[i].Date < existingSchedules[j].Date
		})

		item := model.OneToOneScheduleAvailabilityItem{
			TeacherID:         strconv.FormatInt(slot.TeacherID, 10),
			LessonDate:        slot.LessonDate.Format("2006-01-02"),
			StartTime:         slot.StartAt.Format("15:04"),
			EndTime:           slot.EndAt.Format("15:04"),
			Valid:             len(conflictTypes) == 0,
			ConflictTypes:     conflictTypes,
			ExistingSchedules: existingSchedules,
		}
		if item.Valid {
			result.ValidCount++
		} else {
			item.Message = buildAvailabilityConflictSummaryMessage(conflictTypes)
			result.InvalidCount++
		}
		result.Items = append(result.Items, item)
	}

	return result, nil
}

func (repo *Repository) BatchUpdateTeachingSchedules(ctx context.Context, instID, operatorID int64, dto model.TeachingScheduleBatchUpdateDTO) error {
	teacherID, err := strconv.ParseInt(strings.TrimSpace(dto.TeacherID), 10, 64)
	if err != nil || teacherID <= 0 {
		return errors.New("请选择上课教师")
	}
	assistantIDs := parseStringIDs(dto.AssistantIDs)
	targetIDs := parseStringIDs(dto.IDs)
	if strings.TrimSpace(dto.BatchNo) == "" && len(targetIDs) == 0 {
		return errors.New("缺少待修改日程")
	}
	if strings.TrimSpace(dto.StartTime) == "" || strings.TrimSpace(dto.EndTime) == "" {
		return errors.New("请补全开始与结束时间")
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if n, err := repo.CountInstUsersByIDs(ctx, instID, []int64{teacherID}); err != nil || n != 1 {
		if err != nil {
			return err
		}
		return errors.New("上课教师无效")
	}
	if len(assistantIDs) > 0 {
		if n, err := repo.CountInstUsersByIDs(ctx, instID, assistantIDs); err != nil || n != len(assistantIDs) {
			if err != nil {
				return err
			}
			return errors.New("存在无效的上课助教")
		}
	}
	classroomID, classroomName, _, err := repo.resolveClassroomByIDTx(ctx, tx, instID, dto.ClassroomID)
	if err != nil {
		return err
	}

	schedules, err := repo.loadSchedulesForBatchUpdateTx(ctx, tx, instID, strings.TrimSpace(dto.BatchNo), targetIDs)
	if err != nil {
		return err
	}
	if len(schedules) == 0 {
		return errors.New("未找到可修改的日程")
	}

	teacherName := repo.GetStaffNameByID(ctx, &teacherID)
	assistantNames := make([]string, 0, len(assistantIDs))
	for _, id := range assistantIDs {
		copyID := id
		name := strings.TrimSpace(repo.GetStaffNameByID(ctx, &copyID))
		if name != "" && name != "-" {
			assistantNames = append(assistantNames, name)
		}
	}
	assistantIDsJSON, _ := json.Marshal(stringIDsFromInt64(assistantIDs))
	assistantNamesJSON, _ := json.Marshal(assistantNames)

	updatedSlots := make([]normalizedScheduleSlot, 0, len(schedules))
	excludeIDs := make([]int64, 0, len(schedules))
	for _, item := range schedules {
		startAt, endAt, err := buildScheduleDateTime(item.LessonDate.Format("2006-01-02"), dto.StartTime, dto.EndTime)
		if err != nil {
			return err
		}
		updatedSlots = append(updatedSlots, normalizedScheduleSlot{
			LessonDate: item.LessonDate,
			StartAt:    startAt,
			EndAt:      endAt,
		})
		excludeIDs = append(excludeIDs, item.ID)
	}

	if err := repo.validateTeachingScheduleConflictsTx(ctx, tx, instID, teacherID, classroomID, updatedSlots, strings.TrimSpace(dto.BatchNo), excludeIDs); err != nil {
		return err
	}

	for index, item := range schedules {
		slot := updatedSlots[index]
		if _, err := tx.ExecContext(ctx, `
			UPDATE teaching_schedule
			SET teacher_id = ?, teacher_name = ?, assistant_ids_json = ?, assistant_names_json = ?,
			    classroom_id = ?, classroom_name = ?, lesson_start_at = ?, lesson_end_at = ?,
			    update_id = ?, update_time = NOW()
			WHERE id = ? AND inst_id = ? AND del_flag = 0
		`,
			teacherID,
			teacherName,
			nullJSONBytes(assistantIDsJSON),
			nullJSONBytes(assistantNamesJSON),
			classroomID,
			classroomName,
			slot.StartAt,
			slot.EndAt,
			operatorID,
			item.ID,
			instID,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (repo *Repository) CancelTeachingSchedules(ctx context.Context, instID, operatorID int64, dto model.TeachingScheduleCancelDTO) (model.TeachingScheduleCancelResult, error) {
	ids := parseStringIDs(dto.IDs)
	if len(ids) == 0 {
		return model.TeachingScheduleCancelResult{}, errors.New("缺少待撤销的日程")
	}

	res, err := repo.db.ExecContext(ctx, `
		UPDATE teaching_schedule
		SET del_flag = 1,
		    status = ?,
		    update_id = ?,
		    update_time = NOW()
		WHERE inst_id = ?
		  AND del_flag = 0
		  AND status = ?
		  AND id IN (`+sqlPlaceholders(len(ids))+`)
	`, append([]any{
		model.TeachingScheduleStatusCanceled,
		operatorID,
		instID,
		model.TeachingScheduleStatusActive,
	}, int64SliceToAny(ids)...)...)
	if err != nil {
		return model.TeachingScheduleCancelResult{}, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return model.TeachingScheduleCancelResult{}, err
	}
	if affected <= 0 {
		return model.TeachingScheduleCancelResult{}, errors.New("未找到可撤销的日程")
	}
	return model.TeachingScheduleCancelResult{Canceled: int(affected)}, nil
}

type normalizedScheduleSlot struct {
	LessonDate time.Time
	StartAt    time.Time
	EndAt      time.Time
}

type normalizedAvailabilityScheduleSlot struct {
	TeacherID  int64
	LessonDate time.Time
	StartAt    time.Time
	EndAt      time.Time
}

type scheduleConflictDetailRow struct {
	ID                int64
	StudentID         int64
	TeacherID         int64
	ClassroomID       int64
	ClassType         int
	TeachingClassName string
	StudentName       string
	TeacherName       string
	ClassroomName     string
	LessonDate        time.Time
	StartAt           time.Time
	EndAt             time.Time
}

type scheduleAvailabilityConflictRow struct {
	ID                int64
	TeacherID         int64
	ClassType         int
	TeachingClassName string
	StudentName       string
	TeacherName       string
	ClassroomName     string
	LessonDate        time.Time
	StartAt           time.Time
	EndAt             time.Time
}

type teachingScheduleRow struct {
	ID         int64
	BatchNo    string
	LessonDate time.Time
}

func (repo *Repository) loadScheduleConflictDetailByIDTx(ctx context.Context, tx *sql.Tx, instID, scheduleID int64) (scheduleConflictDetailRow, error) {
	var item scheduleConflictDetailRow
	err := tx.QueryRowContext(ctx, `
		SELECT
			id,
			IFNULL(student_id, 0),
			IFNULL(teacher_id, 0),
			IFNULL(classroom_id, 0),
			IFNULL(class_type, 0),
			IFNULL(teaching_class_name, ''),
			IFNULL(student_name, ''),
			IFNULL(teacher_name, ''),
			IFNULL(classroom_name, ''),
			lesson_date,
			lesson_start_at,
			lesson_end_at
		FROM teaching_schedule
		WHERE inst_id = ?
		  AND id = ?
		  AND del_flag = 0
		  AND status = ?
		LIMIT 1
	`, instID, scheduleID, model.TeachingScheduleStatusActive).Scan(
		&item.ID,
		&item.StudentID,
		&item.TeacherID,
		&item.ClassroomID,
		&item.ClassType,
		&item.TeachingClassName,
		&item.StudentName,
		&item.TeacherName,
		&item.ClassroomName,
		&item.LessonDate,
		&item.StartAt,
		&item.EndAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return item, errors.New("未找到该日程")
		}
		return item, err
	}
	return item, nil
}

func (repo *Repository) loadSchedulesForBatchUpdateTx(ctx context.Context, tx *sql.Tx, instID int64, batchNo string, ids []int64) ([]teachingScheduleRow, error) {
	filters := []string{"inst_id = ?", "del_flag = 0", "status = ?"}
	args := []any{instID, model.TeachingScheduleStatusActive}
	if batchNo != "" {
		filters = append(filters, "batch_no = ?")
		args = append(args, batchNo)
	} else {
		if len(ids) == 0 {
			return nil, nil
		}
		filters = append(filters, "id IN ("+sqlPlaceholders(len(ids))+")")
		for _, id := range ids {
			args = append(args, id)
		}
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT id, IFNULL(batch_no, ''), lesson_date
		FROM teaching_schedule
		WHERE `+strings.Join(filters, " AND ")+`
		ORDER BY lesson_start_at ASC, id ASC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]teachingScheduleRow, 0, 16)
	for rows.Next() {
		var item teachingScheduleRow
		if err := rows.Scan(&item.ID, &item.BatchNo, &item.LessonDate); err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, rows.Err()
}

func (repo *Repository) validateTeachingScheduleConflictsTx(ctx context.Context, tx *sql.Tx, instID, teacherID, classroomID int64, slots []normalizedScheduleSlot, excludeBatchNo string, excludeIDs []int64) error {
	if len(slots) == 0 {
		return nil
	}
	for i := 0; i < len(slots); i++ {
		for j := i + 1; j < len(slots); j++ {
			if slots[i].LessonDate.Format("2006-01-02") != slots[j].LessonDate.Format("2006-01-02") {
				continue
			}
			if slots[i].StartAt.Before(slots[j].EndAt) && slots[i].EndAt.After(slots[j].StartAt) {
				return fmt.Errorf("所选日程在 %s 存在重叠，请调整时间", slots[i].LessonDate.Format("2006-01-02"))
			}
		}
	}

	for _, slot := range slots {
		if teacherID > 0 {
			if conflict, err := repo.countScheduleOverlapTx(ctx, tx, instID, "teacher_id", teacherID, slot, excludeBatchNo, excludeIDs); err != nil {
				return err
			} else if conflict > 0 {
				return fmt.Errorf("老师在 %s %s-%s 已有日程冲突", slot.LessonDate.Format("2006-01-02"), slot.StartAt.Format("15:04"), slot.EndAt.Format("15:04"))
			}
		}
		if classroomID > 0 {
			if conflict, err := repo.countScheduleOverlapTx(ctx, tx, instID, "classroom_id", classroomID, slot, excludeBatchNo, excludeIDs); err != nil {
				return err
			} else if conflict > 0 {
				return fmt.Errorf("教室在 %s %s-%s 已有日程冲突", slot.LessonDate.Format("2006-01-02"), slot.StartAt.Format("15:04"), slot.EndAt.Format("15:04"))
			}
		}
	}
	return nil
}

func (repo *Repository) countScheduleOverlapTx(ctx context.Context, tx *sql.Tx, instID int64, field string, fieldValue int64, slot normalizedScheduleSlot, excludeBatchNo string, excludeIDs []int64) (int, error) {
	filters := []string{
		"inst_id = ?",
		"del_flag = 0",
		"status = ?",
		field + " = ?",
		"lesson_date = ?",
		"lesson_start_at < ?",
		"lesson_end_at > ?",
	}
	args := []any{
		instID,
		model.TeachingScheduleStatusActive,
		fieldValue,
		slot.LessonDate.Format("2006-01-02"),
		slot.EndAt,
		slot.StartAt,
	}
	if excludeBatchNo != "" {
		filters = append(filters, "batch_no <> ?")
		args = append(args, excludeBatchNo)
	}
	if len(excludeIDs) > 0 {
		filters = append(filters, "id NOT IN ("+sqlPlaceholders(len(excludeIDs))+")")
		for _, id := range excludeIDs {
			args = append(args, id)
		}
	}
	var count int
	err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM teaching_schedule
		WHERE `+strings.Join(filters, " AND ")+`
	`, args...).Scan(&count)
	return count, err
}

func (repo *Repository) listScheduleConflictDetailsTx(ctx context.Context, tx *sql.Tx, instID int64, field string, fieldValue int64, slots []normalizedScheduleSlot, excludeBatchNo string, excludeIDs []int64) ([]scheduleConflictDetailRow, error) {
	if fieldValue <= 0 || len(slots) == 0 {
		return []scheduleConflictDetailRow{}, nil
	}
	result := make([]scheduleConflictDetailRow, 0)
	seen := make(map[int64]struct{})
	for _, slot := range slots {
		filters := []string{
			"inst_id = ?",
			"del_flag = 0",
			"status = ?",
			field + " = ?",
			"lesson_date = ?",
			"lesson_start_at < ?",
			"lesson_end_at > ?",
		}
		args := []any{
			instID,
			model.TeachingScheduleStatusActive,
			fieldValue,
			slot.LessonDate.Format("2006-01-02"),
			slot.EndAt,
			slot.StartAt,
		}
		if excludeBatchNo != "" {
			filters = append(filters, "batch_no <> ?")
			args = append(args, excludeBatchNo)
		}
		if len(excludeIDs) > 0 {
			filters = append(filters, "id NOT IN ("+sqlPlaceholders(len(excludeIDs))+")")
			for _, id := range excludeIDs {
				args = append(args, id)
			}
		}
		rows, err := tx.QueryContext(ctx, `
			SELECT
				id,
				IFNULL(student_id, 0),
				IFNULL(teacher_id, 0),
				IFNULL(classroom_id, 0),
				IFNULL(class_type, 0),
				IFNULL(teaching_class_name, ''),
				IFNULL(student_name, ''),
				IFNULL(teacher_name, ''),
				IFNULL(classroom_name, ''),
				lesson_date,
				lesson_start_at,
				lesson_end_at
			FROM teaching_schedule
			WHERE `+strings.Join(filters, " AND ")+`
			ORDER BY lesson_start_at ASC, id ASC
		`, args...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var item scheduleConflictDetailRow
			if err := rows.Scan(
				&item.ID,
				&item.StudentID,
				&item.TeacherID,
				&item.ClassroomID,
				&item.ClassType,
				&item.TeachingClassName,
				&item.StudentName,
				&item.TeacherName,
				&item.ClassroomName,
				&item.LessonDate,
				&item.StartAt,
				&item.EndAt,
			); err != nil {
				rows.Close()
				return nil, err
			}
			if _, ok := seen[item.ID]; ok {
				continue
			}
			seen[item.ID] = struct{}{}
			result = append(result, item)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, err
		}
		rows.Close()
	}
	return result, nil
}

func (repo *Repository) listAvailabilityConflictsByStudentTx(ctx context.Context, tx *sql.Tx, instID, studentID int64, startDate, endDate string) ([]scheduleAvailabilityConflictRow, error) {
	if studentID <= 0 || strings.TrimSpace(startDate) == "" || strings.TrimSpace(endDate) == "" {
		return []scheduleAvailabilityConflictRow{}, nil
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT
			id,
			IFNULL(teacher_id, 0),
			IFNULL(class_type, 0),
			IFNULL(teaching_class_name, ''),
			IFNULL(student_name, ''),
			IFNULL(teacher_name, ''),
			IFNULL(classroom_name, ''),
			lesson_date,
			lesson_start_at,
			lesson_end_at
		FROM teaching_schedule
		WHERE inst_id = ?
		  AND del_flag = 0
		  AND status = ?
		  AND student_id = ?
		  AND lesson_date >= ?
		  AND lesson_date <= ?
		ORDER BY lesson_date ASC, lesson_start_at ASC, id ASC
	`, instID, model.TeachingScheduleStatusActive, studentID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]scheduleAvailabilityConflictRow, 0, 16)
	for rows.Next() {
		var item scheduleAvailabilityConflictRow
		if err := rows.Scan(
			&item.ID,
			&item.TeacherID,
			&item.ClassType,
			&item.TeachingClassName,
			&item.StudentName,
			&item.TeacherName,
			&item.ClassroomName,
			&item.LessonDate,
			&item.StartAt,
			&item.EndAt,
		); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (repo *Repository) listAvailabilityConflictsByTeachersTx(ctx context.Context, tx *sql.Tx, instID int64, teacherIDs []int64, startDate, endDate string) ([]scheduleAvailabilityConflictRow, error) {
	if len(teacherIDs) == 0 || strings.TrimSpace(startDate) == "" || strings.TrimSpace(endDate) == "" {
		return []scheduleAvailabilityConflictRow{}, nil
	}

	args := []any{instID, model.TeachingScheduleStatusActive}
	placeholders := sqlPlaceholders(len(teacherIDs))
	for _, teacherID := range teacherIDs {
		args = append(args, teacherID)
	}
	args = append(args, startDate, endDate)

	rows, err := tx.QueryContext(ctx, `
		SELECT
			id,
			IFNULL(teacher_id, 0),
			IFNULL(class_type, 0),
			IFNULL(teaching_class_name, ''),
			IFNULL(student_name, ''),
			IFNULL(teacher_name, ''),
			IFNULL(classroom_name, ''),
			lesson_date,
			lesson_start_at,
			lesson_end_at
		FROM teaching_schedule
		WHERE inst_id = ?
		  AND del_flag = 0
		  AND status = ?
		  AND teacher_id IN (`+placeholders+`)
		  AND lesson_date >= ?
		  AND lesson_date <= ?
		ORDER BY lesson_date ASC, lesson_start_at ASC, id ASC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]scheduleAvailabilityConflictRow, 0, 32)
	for rows.Next() {
		var item scheduleAvailabilityConflictRow
		if err := rows.Scan(
			&item.ID,
			&item.TeacherID,
			&item.ClassType,
			&item.TeachingClassName,
			&item.StudentName,
			&item.TeacherName,
			&item.ClassroomName,
			&item.LessonDate,
			&item.StartAt,
			&item.EndAt,
		); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func normalizeAvailabilityScheduleSlots(slots []model.OneToOneScheduleAvailabilitySlotDTO) ([]normalizedAvailabilityScheduleSlot, error) {
	result := make([]normalizedAvailabilityScheduleSlot, 0, len(slots))
	seen := make(map[string]struct{}, len(slots))
	for _, item := range slots {
		teacherID, err := strconv.ParseInt(strings.TrimSpace(item.TeacherID), 10, 64)
		if err != nil || teacherID <= 0 {
			return nil, errors.New("存在无效的教师")
		}
		startAt, endAt, err := buildScheduleDateTime(item.LessonDate, item.StartTime, item.EndTime)
		if err != nil {
			return nil, err
		}
		key := strconv.FormatInt(teacherID, 10) + "|" + startAt.Format(time.RFC3339) + "|" + endAt.Format(time.RFC3339)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, normalizedAvailabilityScheduleSlot{
			TeacherID:  teacherID,
			LessonDate: startOfDay(startAt),
			StartAt:    startAt,
			EndAt:      endAt,
		})
	}
	return result, nil
}

func normalizeCreateScheduleSlots(slots []model.TeachingScheduleCreateSlotDTO) ([]normalizedScheduleSlot, error) {
	result := make([]normalizedScheduleSlot, 0, len(slots))
	seen := make(map[string]struct{}, len(slots))
	for _, item := range slots {
		startAt, endAt, err := buildScheduleDateTime(item.LessonDate, item.StartTime, item.EndTime)
		if err != nil {
			return nil, err
		}
		key := startAt.Format(time.RFC3339) + "|" + endAt.Format(time.RFC3339)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, normalizedScheduleSlot{
			LessonDate: startOfDay(startAt),
			StartAt:    startAt,
			EndAt:      endAt,
		})
	}
	return result, nil
}

func buildScheduleDateTime(dateStr, startTimeStr, endTimeStr string) (time.Time, time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	startTimeStr = strings.TrimSpace(startTimeStr)
	endTimeStr = strings.TrimSpace(endTimeStr)
	if dateStr == "" || startTimeStr == "" || endTimeStr == "" {
		return time.Time{}, time.Time{}, errors.New("日程日期和时间不能为空")
	}
	startAt, err := time.ParseInLocation("2006-01-02 15:04", dateStr+" "+startTimeStr, time.Local)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("开始时间格式无效")
	}
	endAt, err := time.ParseInLocation("2006-01-02 15:04", dateStr+" "+endTimeStr, time.Local)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("结束时间格式无效")
	}
	if !endAt.After(startAt) {
		return time.Time{}, time.Time{}, errors.New("结束时间需晚于开始时间")
	}
	return startAt, endAt, nil
}

func parseStringIDs(values []string) []int64 {
	result := make([]int64, 0, len(values))
	seen := make(map[int64]struct{}, len(values))
	for _, raw := range values {
		value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
		if err != nil || value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func stringIDsFromInt64(values []int64) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		result = append(result, strconv.FormatInt(value, 10))
	}
	return result
}

func nullJSONBytes(value []byte) any {
	if len(value) == 0 || string(value) == "null" || string(value) == "[]" {
		return nil
	}
	return value
}

func emptyStringIfZero(value int64) string {
	if value <= 0 {
		return ""
	}
	return strconv.FormatInt(value, 10)
}

func startOfDay(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}

func buildScheduleConflictResult(
	base model.OneToOneScheduleCreateContext,
	teacherName string,
	classroomName string,
	slots []normalizedScheduleSlot,
	teacherConflicts []scheduleConflictDetailRow,
	classroomConflicts []scheduleConflictDetailRow,
	studentConflicts []scheduleConflictDetailRow,
) ([]model.TeachingScheduleConflictItem, []model.TeachingScheduleConflictItem, []string) {
	typeSet := make(map[string]struct{})
	current := make([]model.TeachingScheduleConflictItem, 0, len(slots))
	for _, slot := range slots {
		conflictTypes := make([]string, 0, 3)
		if slotHasConflict(slot, teacherConflicts) {
			conflictTypes = append(conflictTypes, "老师")
			typeSet["老师"] = struct{}{}
		}
		if slotHasConflict(slot, classroomConflicts) {
			conflictTypes = append(conflictTypes, "教室")
			typeSet["教室"] = struct{}{}
		}
		if slotHasConflict(slot, studentConflicts) {
			conflictTypes = append(conflictTypes, "学员")
			typeSet["学员"] = struct{}{}
		}
		current = append(current, model.TeachingScheduleConflictItem{
			Name:          base.ClassName,
			ClassTypeText: "1对1日程",
			Date:          slot.LessonDate.Format("2006-01-02"),
			Week:          weekDisplay(slot.LessonDate),
			TimeText:      slot.StartAt.Format("15:04") + "~" + slot.EndAt.Format("15:04"),
			TeacherID:     "",
			TeacherName:   firstNonEmptyString(teacherName, "-"),
			ClassroomName: firstNonEmptyString(classroomName, "-"),
			StudentNames:  compactStrings([]string{base.StudentName}),
			ConflictTypes: conflictTypes,
		})
	}

	existingMap := make(map[int64]model.TeachingScheduleConflictItem)
	appendExisting := func(row scheduleConflictDetailRow, conflictType string) {
		item, ok := existingMap[row.ID]
		if !ok {
			item = model.TeachingScheduleConflictItem{
				Name:          row.TeachingClassName,
				ClassTypeText: scheduleClassTypeText(row.ClassType),
				Date:          row.LessonDate.Format("2006-01-02"),
				Week:          weekDisplay(row.LessonDate),
				TimeText:      row.StartAt.Format("15:04") + "~" + row.EndAt.Format("15:04"),
				TeacherID:     emptyStringIfZero(row.TeacherID),
				TeacherName:   firstNonEmptyString(row.TeacherName, "-"),
				ClassroomName: firstNonEmptyString(row.ClassroomName, "-"),
				StudentNames:  compactStrings([]string{row.StudentName}),
				ConflictTypes: []string{},
			}
		}
		if !containsString(item.ConflictTypes, conflictType) {
			item.ConflictTypes = append(item.ConflictTypes, conflictType)
		}
		existingMap[row.ID] = item
	}
	for _, row := range teacherConflicts {
		appendExisting(row, "老师")
	}
	for _, row := range classroomConflicts {
		appendExisting(row, "教室")
	}
	for _, row := range studentConflicts {
		appendExisting(row, "学员")
	}

	existing := make([]model.TeachingScheduleConflictItem, 0, len(existingMap))
	for _, item := range existingMap {
		existing = append(existing, item)
	}
	conflictTypes := make([]string, 0, len(typeSet))
	for key := range typeSet {
		conflictTypes = append(conflictTypes, key)
	}
	sort.Strings(conflictTypes)
	sort.Slice(existing, func(i, j int) bool {
		if existing[i].Date == existing[j].Date {
			return existing[i].TimeText < existing[j].TimeText
		}
		return existing[i].Date < existing[j].Date
	})
	return current, existing, conflictTypes
}

func buildScheduleConflictResultFromExisting(
	current scheduleConflictDetailRow,
	teacherConflicts []scheduleConflictDetailRow,
	classroomConflicts []scheduleConflictDetailRow,
	studentConflicts []scheduleConflictDetailRow,
) ([]model.TeachingScheduleConflictItem, []model.TeachingScheduleConflictItem, []string) {
	typeSet := make(map[string]struct{})
	currentConflictTypes := make([]string, 0, 3)
	if len(teacherConflicts) > 0 {
		currentConflictTypes = append(currentConflictTypes, "老师")
		typeSet["老师"] = struct{}{}
	}
	if len(classroomConflicts) > 0 {
		currentConflictTypes = append(currentConflictTypes, "教室")
		typeSet["教室"] = struct{}{}
	}
	if len(studentConflicts) > 0 {
		currentConflictTypes = append(currentConflictTypes, "学员")
		typeSet["学员"] = struct{}{}
	}

	currentItems := []model.TeachingScheduleConflictItem{{
		Name:          current.TeachingClassName,
		ClassTypeText: scheduleClassTypeText(current.ClassType),
		Date:          current.LessonDate.Format("2006-01-02"),
		Week:          weekDisplay(current.LessonDate),
		TimeText:      current.StartAt.Format("15:04") + "~" + current.EndAt.Format("15:04"),
		TeacherID:     emptyStringIfZero(current.TeacherID),
		TeacherName:   firstNonEmptyString(current.TeacherName, "-"),
		ClassroomName: firstNonEmptyString(current.ClassroomName, "-"),
		StudentNames:  compactStrings([]string{current.StudentName}),
		ConflictTypes: currentConflictTypes,
	}}

	existingMap := make(map[int64]model.TeachingScheduleConflictItem)
	appendExisting := func(row scheduleConflictDetailRow, conflictType string) {
		if row.ID == current.ID {
			return
		}
		item, ok := existingMap[row.ID]
		if !ok {
			item = model.TeachingScheduleConflictItem{
				Name:          row.TeachingClassName,
				ClassTypeText: scheduleClassTypeText(row.ClassType),
				Date:          row.LessonDate.Format("2006-01-02"),
				Week:          weekDisplay(row.LessonDate),
				TimeText:      row.StartAt.Format("15:04") + "~" + row.EndAt.Format("15:04"),
				TeacherID:     emptyStringIfZero(row.TeacherID),
				TeacherName:   firstNonEmptyString(row.TeacherName, "-"),
				ClassroomName: firstNonEmptyString(row.ClassroomName, "-"),
				StudentNames:  compactStrings([]string{row.StudentName}),
				ConflictTypes: []string{},
			}
		}
		if !containsString(item.ConflictTypes, conflictType) {
			item.ConflictTypes = append(item.ConflictTypes, conflictType)
		}
		existingMap[row.ID] = item
	}
	for _, row := range teacherConflicts {
		appendExisting(row, "老师")
	}
	for _, row := range classroomConflicts {
		appendExisting(row, "教室")
	}
	for _, row := range studentConflicts {
		appendExisting(row, "学员")
	}

	existing := make([]model.TeachingScheduleConflictItem, 0, len(existingMap))
	for _, item := range existingMap {
		sort.Strings(item.ConflictTypes)
		existing = append(existing, item)
	}
	conflictTypes := make([]string, 0, len(typeSet))
	for key := range typeSet {
		conflictTypes = append(conflictTypes, key)
	}
	sort.Strings(conflictTypes)
	sort.Slice(existing, func(i, j int) bool {
		if existing[i].Date == existing[j].Date {
			return existing[i].TimeText < existing[j].TimeText
		}
		return existing[i].Date < existing[j].Date
	})
	return currentItems, existing, conflictTypes
}

func buildConflictSummaryMessage(conflictTypes []string) string {
	if len(conflictTypes) == 0 {
		return "当前排课方案存在冲突"
	}
	if len(conflictTypes) == 1 {
		return "当前创建日程存在" + conflictTypes[0] + "冲突"
	}
	return "当前创建日程存在" + strings.Join(conflictTypes, "、") + "冲突"
}

func buildAvailabilityConflictSummaryMessage(conflictTypes []string) string {
	if len(conflictTypes) == 0 {
		return ""
	}
	if len(conflictTypes) == 1 {
		return "当前空位存在" + conflictTypes[0] + "冲突"
	}
	return "当前空位存在" + strings.Join(conflictTypes, "、") + "冲突"
}

func slotHasConflict(slot normalizedScheduleSlot, rows []scheduleConflictDetailRow) bool {
	for _, row := range rows {
		if row.LessonDate.Format("2006-01-02") == slot.LessonDate.Format("2006-01-02") &&
			row.StartAt.Before(slot.EndAt) &&
			row.EndAt.After(slot.StartAt) {
			return true
		}
	}
	return false
}

func availabilitySlotsOverlap(slot normalizedAvailabilityScheduleSlot, lessonDate, startAt, endAt time.Time) bool {
	return lessonDate.Format("2006-01-02") == slot.LessonDate.Format("2006-01-02") &&
		startAt.Before(slot.EndAt) &&
		endAt.After(slot.StartAt)
}

func availabilityDateRange(slots []normalizedAvailabilityScheduleSlot) (string, string) {
	if len(slots) == 0 {
		return "", ""
	}
	start := slots[0].LessonDate
	end := slots[0].LessonDate
	for _, slot := range slots[1:] {
		if slot.LessonDate.Before(start) {
			start = slot.LessonDate
		}
		if slot.LessonDate.After(end) {
			end = slot.LessonDate
		}
	}
	return start.Format("2006-01-02"), end.Format("2006-01-02")
}

func collectAvailabilityTeacherIDs(slots []normalizedAvailabilityScheduleSlot) []int64 {
	result := make([]int64, 0, len(slots))
	seen := make(map[int64]struct{}, len(slots))
	for _, slot := range slots {
		if slot.TeacherID <= 0 {
			continue
		}
		if _, ok := seen[slot.TeacherID]; ok {
			continue
		}
		seen[slot.TeacherID] = struct{}{}
		result = append(result, slot.TeacherID)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func appendAvailabilityConflict(existingMap map[int64]model.TeachingScheduleConflictItem, row scheduleAvailabilityConflictRow, conflictType string) {
	item, ok := existingMap[row.ID]
	if !ok {
		item = model.TeachingScheduleConflictItem{
			Name:          row.TeachingClassName,
			ClassTypeText: scheduleClassTypeText(row.ClassType),
			Date:          row.LessonDate.Format("2006-01-02"),
			Week:          weekDisplay(row.LessonDate),
			TimeText:      row.StartAt.Format("15:04") + "~" + row.EndAt.Format("15:04"),
			TeacherID:     emptyStringIfZero(row.TeacherID),
			TeacherName:   firstNonEmptyString(row.TeacherName, "-"),
			ClassroomName: firstNonEmptyString(row.ClassroomName, "-"),
			StudentNames:  compactStrings([]string{row.StudentName}),
			ConflictTypes: []string{},
		}
	}
	if !containsString(item.ConflictTypes, conflictType) {
		item.ConflictTypes = append(item.ConflictTypes, conflictType)
	}
	existingMap[row.ID] = item
}

func buildUnavailableAvailabilityResult(slots []normalizedAvailabilityScheduleSlot, message string) model.OneToOneScheduleAvailabilityResult {
	result := model.OneToOneScheduleAvailabilityResult{
		InvalidCount: len(slots),
		Items:        make([]model.OneToOneScheduleAvailabilityItem, 0, len(slots)),
	}
	for _, slot := range slots {
		result.Items = append(result.Items, model.OneToOneScheduleAvailabilityItem{
			TeacherID:  strconv.FormatInt(slot.TeacherID, 10),
			LessonDate: slot.LessonDate.Format("2006-01-02"),
			StartTime:  slot.StartAt.Format("15:04"),
			EndTime:    slot.EndAt.Format("15:04"),
			Valid:      false,
			Message:    message,
		})
	}
	return result
}

func scheduleClassTypeText(classType int) string {
	if classType == model.TeachingClassTypeOneToOne {
		return "1对1日程"
	}
	if classType == model.TeachingClassTypeNormal {
		return "班级日程"
	}
	return "日程"
}

func weekDisplay(value time.Time) string {
	switch value.Weekday() {
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

// SoftDeleteAllTeachingSchedulesForInst 软删本机构全部未删除排课（列表与矩阵仅展示 del_flag=0）
func (repo *Repository) SoftDeleteAllTeachingSchedulesForInst(ctx context.Context, instID, operatorID int64) (int64, error) {
	res, err := repo.db.ExecContext(ctx, `
		UPDATE teaching_schedule
		SET del_flag = 1,
		    status = ?,
		    update_id = ?,
		    update_time = NOW()
		WHERE inst_id = ? AND del_flag = 0
	`, model.TeachingScheduleStatusCanceled, operatorID, instID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func compactStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || value == "-" {
			continue
		}
		result = append(result, value)
	}
	return result
}

func containsString(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}
