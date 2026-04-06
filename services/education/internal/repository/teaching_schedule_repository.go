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
	if err := ensureColumnsOnTable(ctx, db, "teaching_schedule", map[string]string{
		"assistant_ids_json":   "assistant_ids_json JSON NULL AFTER teacher_name",
		"assistant_names_json": "assistant_names_json JSON NULL AFTER assistant_ids_json",
		"classroom_id":         "classroom_id BIGINT NOT NULL DEFAULT 0 AFTER assistant_names_json",
		"classroom_name":       "classroom_name VARCHAR(150) NOT NULL DEFAULT '' AFTER classroom_id",
		"batch_no":             "batch_no VARCHAR(64) NOT NULL DEFAULT '' AFTER lesson_end_at",
		"batch_size":           "batch_size INT NOT NULL DEFAULT 1 AFTER batch_no",
	}); err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS teaching_schedule_batch_meta (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			inst_id BIGINT NOT NULL,
			batch_key VARCHAR(96) NOT NULL,
			batch_no VARCHAR(64) NOT NULL DEFAULT '',
			class_type INT NOT NULL DEFAULT 0,
			teaching_class_id BIGINT NOT NULL DEFAULT 0,
			meta_json JSON NULL,
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT NOT NULL DEFAULT 0,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uniq_teaching_schedule_batch_meta_key (inst_id, batch_key),
			KEY idx_teaching_schedule_batch_meta_batch_no (inst_id, batch_no)
		)
	`)
	if err != nil {
		return err
	}
	return ensureColumnsOnTable(ctx, db, "teaching_schedule_batch_meta", map[string]string{
		"batch_no":          "batch_no VARCHAR(64) NOT NULL DEFAULT '' AFTER batch_key",
		"class_type":        "class_type INT NOT NULL DEFAULT 0 AFTER batch_no",
		"teaching_class_id": "teaching_class_id BIGINT NOT NULL DEFAULT 0 AFTER class_type",
		"meta_json":         "meta_json JSON NULL AFTER teaching_class_id",
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

	assistantConflicts := []scheduleConflictDetailRow{}
	if len(current.AssistantIDs) > 0 {
		assistantConflicts, err = repo.listScheduleConflictDetailsByAssistantsTx(ctx, tx, instID, parseStringIDs(current.AssistantIDs), []normalizedScheduleSlot{slot}, "", excludeIDs)
		if err != nil {
			return model.TeachingScheduleValidationResult{}, err
		}
	}

	currentItems, existingItems, conflictTypes := buildScheduleConflictResultFromExisting(current, teacherConflicts, classroomConflicts, studentConflicts, assistantConflicts)
	if len(conflictTypes) == 0 {
		return model.TeachingScheduleValidationResult{
			Valid:            true,
			Message:          "当前日程暂无冲突",
			CurrentSchedules: currentItems,
		}, nil
	}

	return model.TeachingScheduleValidationResult{
		Valid:             false,
		Message:           buildExistingConflictSummaryMessage(conflictTypes),
		CurrentSchedules:  currentItems,
		ExistingSchedules: existingItems,
		ConflictTypes:     conflictTypes,
	}, nil
}

func (repo *Repository) ListTeachingSchedules(ctx context.Context, instID int64, query model.TeachingScheduleListQueryDTO) ([]model.TeachingScheduleVO, error) {
	filters := []string{"ts.inst_id = ?", "ts.del_flag = 0", "ts.status = ?"}
	args := []any{instID, model.TeachingScheduleStatusActive}
	if batchNo := strings.TrimSpace(query.BatchNo); batchNo != "" {
		filters = append(filters, "ts.batch_no = ?")
		args = append(args, batchNo)
	}
	if ids := parseStringIDs(query.IDs); len(ids) > 0 {
		filters = append(filters, "ts.id IN ("+sqlPlaceholders(len(ids))+")")
		for _, id := range ids {
			args = append(args, id)
		}
	}
	if strings.TrimSpace(query.StartDate) != "" {
		filters = append(filters, "ts.lesson_date >= ?")
		args = append(args, strings.TrimSpace(query.StartDate))
	}
	if strings.TrimSpace(query.EndDate) != "" {
		filters = append(filters, "ts.lesson_date <= ?")
		args = append(args, strings.TrimSpace(query.EndDate))
	}
	if sid := strings.TrimSpace(query.StudentID); sid != "" {
		filters = append(filters, "CAST(ts.student_id AS CHAR) = ?")
		args = append(args, sid)
	}
	if query.ClassType != nil && *query.ClassType > 0 {
		filters = append(filters, "ts.class_type = ?")
		args = append(args, *query.ClassType)
	}
	appendInt64InFilter := func(column string, values []int64) {
		if len(values) == 0 {
			return
		}
		holders := make([]string, 0, len(values))
		for _, item := range values {
			if item <= 0 {
				continue
			}
			holders = append(holders, "?")
			args = append(args, item)
		}
		if len(holders) > 0 {
			filters = append(filters, column+" IN ("+strings.Join(holders, ",")+")")
		}
	}
	appendScheduleTeacherFilter := func(values []int64) {
		if len(values) == 0 {
			return
		}
		teacherHolders := make([]string, 0, len(values))
		assistantParts := make([]string, 0, len(values))
		assistantArgs := make([]any, 0, len(values))
		for _, item := range values {
			if item <= 0 {
				continue
			}
			teacherHolders = append(teacherHolders, "?")
			args = append(args, item)
			assistantParts = append(assistantParts, "JSON_SEARCH(COALESCE(ts.assistant_ids_json, JSON_ARRAY()), 'one', ?) IS NOT NULL")
			assistantArgs = append(assistantArgs, strconv.FormatInt(item, 10))
		}
		if len(teacherHolders) == 0 {
			return
		}
		args = append(args, assistantArgs...)
		filters = append(filters, "(ts.teacher_id IN ("+strings.Join(teacherHolders, ",")+") OR "+strings.Join(assistantParts, " OR ")+")")
	}
	appendScheduleTeacherFilter(query.ScheduleTeacherIDs)
	appendInt64InFilter("ts.classroom_id", query.ClassroomIDs)
	appendInt64InFilter("ts.lesson_id", query.LessonIDs)

	if len(query.GroupClassIDs) > 0 || len(query.OneToOneClassIDs) > 0 {
		typeParts := make([]string, 0, 2)
		typeArgs := make([]any, 0, len(query.GroupClassIDs)+len(query.OneToOneClassIDs)+2)
		if len(query.GroupClassIDs) > 0 {
			holders := make([]string, 0, len(query.GroupClassIDs))
			typeArgs = append(typeArgs, model.TeachingClassTypeNormal)
			for _, item := range query.GroupClassIDs {
				if item <= 0 {
					continue
				}
				holders = append(holders, "?")
				typeArgs = append(typeArgs, item)
			}
			if len(holders) > 0 {
				typeParts = append(typeParts, "(ts.class_type = ? AND ts.teaching_class_id IN ("+strings.Join(holders, ",")+"))")
			}
		}
		if len(query.OneToOneClassIDs) > 0 {
			holders := make([]string, 0, len(query.OneToOneClassIDs))
			typeArgs = append(typeArgs, model.TeachingClassTypeOneToOne)
			for _, item := range query.OneToOneClassIDs {
				if item <= 0 {
					continue
				}
				holders = append(holders, "?")
				typeArgs = append(typeArgs, item)
			}
			if len(holders) > 0 {
				typeParts = append(typeParts, "(ts.class_type = ? AND ts.teaching_class_id IN ("+strings.Join(holders, ",")+"))")
			}
		}
		if len(typeParts) > 0 {
			filters = append(filters, "("+strings.Join(typeParts, " OR ")+")")
			args = append(args, typeArgs...)
		}
	}

	trialExistsSQL := buildTeachingScheduleTrialExistsSQL()
	if typeSet := normalizeTeachingScheduleTypeFilters(query.ScheduleTypeFilters); len(typeSet) > 0 {
		typeParts := make([]string, 0, len(typeSet))
		if _, ok := typeSet["group_class"]; ok {
			typeParts = append(typeParts, "(ts.class_type = 1 AND NOT "+trialExistsSQL+")")
		}
		if _, ok := typeSet["one_to_one"]; ok {
			typeParts = append(typeParts, "(ts.class_type = 2 AND NOT "+trialExistsSQL+")")
		}
		if _, ok := typeSet["trial"]; ok {
			typeParts = append(typeParts, trialExistsSQL)
		}
		if len(typeParts) > 0 {
			filters = append(filters, "("+strings.Join(typeParts, " OR ")+")")
		}
	}

	if statusSet := normalizeTeachingScheduleCallStatusFilters(query.CallStatusFilters); len(statusSet) > 0 {
		existsSQL, err := repo.buildTeachingScheduleRecordExistsSQL(ctx)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(existsSQL) != "" {
			statusParts := make([]string, 0, len(statusSet))
			if _, ok := statusSet["signed"]; ok {
				statusParts = append(statusParts, existsSQL)
			}
			if _, ok := statusSet["unsigned"]; ok {
				statusParts = append(statusParts, "NOT "+existsSQL)
			}
			if len(statusParts) > 0 {
				filters = append(filters, "("+strings.Join(statusParts, " OR ")+")")
			}
		}
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
		FROM teaching_schedule ts
		WHERE `+strings.Join(filters, " AND ")+`
		ORDER BY ts.lesson_start_at ASC, ts.id ASC
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

func normalizeTeachingScheduleTypeFilters(list []string) map[string]struct{} {
	result := make(map[string]struct{})
	for _, item := range list {
		switch strings.TrimSpace(strings.ToLower(item)) {
		case "group_class", "group", "class", "class_schedule":
			result["group_class"] = struct{}{}
		case "one_to_one", "one2one", "1v1":
			result["one_to_one"] = struct{}{}
		case "trial", "audition":
			result["trial"] = struct{}{}
		}
	}
	return result
}

func normalizeTeachingScheduleCallStatusFilters(list []string) map[string]struct{} {
	result := make(map[string]struct{})
	for _, item := range list {
		switch strings.TrimSpace(strings.ToLower(item)) {
		case "signed", "called":
			result["signed"] = struct{}{}
		case "unsigned", "uncalled":
			result["unsigned"] = struct{}{}
		}
	}
	return result
}

func buildTeachingScheduleTrialExistsSQL() string {
	return `EXISTS (
		SELECT 1
		FROM teaching_class_student tcs
		LEFT JOIN tuition_account ta_ded ON ta_ded.id = COALESCE(
			NULLIF(tcs.primary_tuition_account_id, 0),
			(SELECT MIN(ta0.id) FROM tuition_account ta0
			 WHERE ta0.order_course_detail_id = tcs.order_course_detail_id
			   AND ta0.inst_id = tcs.inst_id AND ta0.del_flag = 0)
		) AND ta_ded.inst_id = tcs.inst_id AND ta_ded.del_flag = 0
		LEFT JOIN sale_order_course_detail sod_ta_ded ON sod_ta_ded.id = ta_ded.order_course_detail_id AND sod_ta_ded.del_flag = 0
		LEFT JOIN inst_course_quotation icq_ded ON icq_ded.id = COALESCE(
			NULLIF(ta_ded.quote_id, 0),
			NULLIF(sod_ta_ded.quote_id, 0),
			(SELECT qx.id FROM inst_course_quotation qx
			 WHERE qx.course_id = ta_ded.course_id AND qx.del_flag = 0
			   AND ABS(IFNULL(qx.quantity, 0) - IFNULL(ta_ded.total_quantity, 0)) < 0.000001
			   AND ABS(IFNULL(qx.price, 0) - IFNULL(ta_ded.total_tuition, 0)) < 0.000001
			 ORDER BY qx.id DESC LIMIT 1),
			(SELECT qmin.id FROM inst_course_quotation qmin
			 WHERE qmin.course_id = ta_ded.course_id AND qmin.del_flag = 0
			 ORDER BY qmin.id ASC LIMIT 1)
		) AND icq_ded.del_flag = 0
		WHERE tcs.inst_id = ts.inst_id
		  AND tcs.teaching_class_id = ts.teaching_class_id
		  AND tcs.del_flag = 0
		  AND (ts.student_id = 0 OR tcs.student_id = ts.student_id)
		  AND IFNULL(icq_ded.lesson_audition, 0) = 1
	)`
}

func (repo *Repository) buildTeachingScheduleRecordExistsSQL(ctx context.Context) (string, error) {
	teachingTable, err := repo.firstExistingTable(ctx, []string{"teaching_record", "inst_teaching_record", "class_teaching_record"})
	if err != nil || strings.TrimSpace(teachingTable) == "" {
		return "", err
	}
	classIDColumn, err := repo.firstExistingColumn(ctx, teachingTable, []string{"class_id"})
	if err != nil || strings.TrimSpace(classIDColumn) == "" {
		return "", err
	}
	lessonDayColumn, err := repo.firstExistingColumn(ctx, teachingTable, []string{"lesson_day", "teaching_day", "class_day", "teaching_date", "lesson_date"})
	if err != nil {
		return "", err
	}
	startMinutesColumn, err := repo.firstExistingColumn(ctx, teachingTable, []string{"start_minutes", "start_minute"})
	if err != nil {
		return "", err
	}
	endMinutesColumn, err := repo.firstExistingColumn(ctx, teachingTable, []string{"end_minutes", "end_minute"})
	if err != nil {
		return "", err
	}
	instIDColumn, err := repo.firstExistingColumn(ctx, teachingTable, []string{"inst_id"})
	if err != nil {
		return "", err
	}
	delFlagColumn, err := repo.firstExistingColumn(ctx, teachingTable, []string{"del_flag"})
	if err != nil {
		return "", err
	}

	parts := []string{
		"CAST(tr." + classIDColumn + " AS CHAR) = CAST(ts.teaching_class_id AS CHAR)",
	}
	if instIDColumn != "" {
		parts = append(parts, "tr."+instIDColumn+" = ts.inst_id")
	}
	if delFlagColumn != "" {
		parts = append(parts, "IFNULL(tr."+delFlagColumn+", 0) = 0")
	}
	if lessonDayColumn != "" {
		parts = append(parts, "DATE(tr."+lessonDayColumn+") = ts.lesson_date")
	}
	if startMinutesColumn != "" {
		parts = append(parts, "IFNULL(tr."+startMinutesColumn+", 0) = (HOUR(ts.lesson_start_at) * 60 + MINUTE(ts.lesson_start_at))")
	}
	if endMinutesColumn != "" {
		parts = append(parts, "IFNULL(tr."+endMinutesColumn+", 0) = (HOUR(ts.lesson_end_at) * 60 + MINUTE(ts.lesson_end_at))")
	}

	return "EXISTS (SELECT 1 FROM " + teachingTable + " tr WHERE " + strings.Join(parts, " AND ") + ")", nil
}

func (repo *Repository) CreateOneToOneSchedules(ctx context.Context, instID, operatorID int64, dto model.CreateOneToOneSchedulesDTO) (model.CreateOneToOneSchedulesResult, error) {
	classID, err := strconv.ParseInt(strings.TrimSpace(dto.OneToOneID), 10, 64)
	if err != nil || classID <= 0 {
		return model.CreateOneToOneSchedulesResult{}, errors.New("请选择1对1")
	}
	fallbackTeacherID, err := parseOptionalPositiveID(dto.TeacherID)
	if err != nil {
		return model.CreateOneToOneSchedulesResult{}, errors.New("请选择上课教师")
	}
	fallbackClassroomID, err := parseOptionalPositiveID(dto.ClassroomID)
	if err != nil {
		return model.CreateOneToOneSchedulesResult{}, errors.New("classroomId 无效")
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

	plans, err := normalizeCreateSchedulePlans(dto.Schedules, fallbackTeacherID, fallbackClassroomID, assistantIDs)
	if err != nil {
		return model.CreateOneToOneSchedulesResult{}, err
	}
	applyCreateScheduleConflictAllowances(plans, dto.AllowStudentConflict, dto.AllowClassroomConflict)
	teacherIDs := collectPlanTeacherIDs(plans)
	if len(teacherIDs) == 0 {
		return model.CreateOneToOneSchedulesResult{}, errors.New("请选择上课教师")
	}

	if n, err := repo.CountInstUsersByIDs(ctx, instID, teacherIDs); err != nil || n != len(teacherIDs) {
		if err != nil {
			return model.CreateOneToOneSchedulesResult{}, err
		}
		return model.CreateOneToOneSchedulesResult{}, errors.New("上课教师无效")
	}
	planAssistantIDs := collectPlanAssistantIDs(plans)
	if len(planAssistantIDs) > 0 {
		if n, err := repo.CountInstUsersByIDs(ctx, instID, planAssistantIDs); err != nil || n != len(planAssistantIDs) {
			if err != nil {
				return model.CreateOneToOneSchedulesResult{}, err
			}
			return model.CreateOneToOneSchedulesResult{}, errors.New("存在无效的上课助教")
		}
	}

	classroomNames, err := repo.resolveClassroomNamesTx(ctx, tx, instID, collectPlanClassroomIDs(plans))
	if err != nil {
		return model.CreateOneToOneSchedulesResult{}, err
	}

	teacherNames := repo.resolveTeacherNames(ctx, teacherIDs)
	assistantNameMap := repo.resolveTeacherNames(ctx, planAssistantIDs)
	for i := range plans {
		plans[i].AssistantNames = compactStrings(func() []string {
			names := make([]string, 0, len(plans[i].AssistantIDs))
			for _, id := range plans[i].AssistantIDs {
				if name := strings.TrimSpace(assistantNameMap[id]); name != "" && name != "-" {
					names = append(names, name)
				}
			}
			return names
		}())
	}

	teacherConflictsByPlan, err := repo.listTeacherConflictsByPlanTx(ctx, tx, instID, plans, nil)
	if err != nil {
		return model.CreateOneToOneSchedulesResult{}, err
	}
	classroomConflictsByPlan, err := repo.listClassroomConflictsByPlanTx(ctx, tx, instID, plans, nil)
	if err != nil {
		return model.CreateOneToOneSchedulesResult{}, err
	}
	studentConflicts, err := repo.listScheduleConflictDetailsTx(ctx, tx, instID, "student_id", base.StudentID, plansToSlots(plans), "", nil)
	if err != nil {
		return model.CreateOneToOneSchedulesResult{}, err
	}
	assistantConflictsByPlan, err := repo.listAssistantConflictsByPlanTx(ctx, tx, instID, plans, nil)
	if err != nil {
		return model.CreateOneToOneSchedulesResult{}, err
	}

	hardConflictTypes := make(map[string]struct{})
	for _, plan := range plans {
		key := schedulePlanKey(plan)
		if len(teacherConflictsByPlan[key]) > 0 {
			hardConflictTypes["老师"] = struct{}{}
		}
		if len(classroomConflictsByPlan[key]) > 0 && !plan.AllowClassroomConflict {
			hardConflictTypes["教室"] = struct{}{}
		}
		if slotHasConflict(normalizedScheduleSlot{
			LessonDate: plan.LessonDate,
			StartAt:    plan.StartAt,
			EndAt:      plan.EndAt,
		}, studentConflicts) && !plan.AllowStudentConflict {
			hardConflictTypes["学员"] = struct{}{}
		}
		if len(assistantConflictsByPlan[key]) > 0 {
			hardConflictTypes["助教"] = struct{}{}
		}
	}
	if len(hardConflictTypes) > 0 {
		conflictTypes := make([]string, 0, len(hardConflictTypes))
		for key := range hardConflictTypes {
			conflictTypes = append(conflictTypes, key)
		}
		sort.Strings(conflictTypes)
		return model.CreateOneToOneSchedulesResult{}, errors.New(buildConflictSummaryMessage(conflictTypes))
	}

	batchNo := ""
	if len(plans) > 1 {
		batchNo = fmt.Sprintf("BATCH-%d", time.Now().UnixNano())
	}

	result := model.CreateOneToOneSchedulesResult{
		BatchNo: batchNo,
		Count:   len(plans),
		List:    make([]model.TeachingScheduleVO, 0, len(plans)),
	}
	createdScheduleIDs := make([]int64, 0, len(plans))

	for _, plan := range plans {
		teacherName := firstNonEmptyString(teacherNames[plan.TeacherID], "-")
		classroomName := classroomNames[plan.ClassroomID]
		classroomID := plan.ClassroomID
		assistantIDsJSON, _ := json.Marshal(stringIDsFromInt64(plan.AssistantIDs))
		assistantNamesJSON, _ := json.Marshal(plan.AssistantNames)
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
			plan.TeacherID,
			teacherName,
			nullJSONBytes(assistantIDsJSON),
			nullJSONBytes(assistantNamesJSON),
			classroomID,
			classroomName,
			plan.LessonDate.Format("2006-01-02"),
			plan.StartAt,
			plan.EndAt,
			batchNo,
			len(plans),
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
		createdScheduleIDs = append(createdScheduleIDs, id)
		result.List = append(result.List, model.TeachingScheduleVO{
			ID:                strconv.FormatInt(id, 10),
			BatchNo:           batchNo,
			BatchSize:         len(plans),
			ClassType:         model.TeachingClassTypeOneToOne,
			TeachingClassID:   strconv.FormatInt(base.ClassID, 10),
			TeachingClassName: base.ClassName,
			StudentID:         strconv.FormatInt(base.StudentID, 10),
			StudentName:       base.StudentName,
			LessonID:          strconv.FormatInt(base.LessonID, 10),
			LessonName:        base.LessonName,
			TeacherID:         strconv.FormatInt(plan.TeacherID, 10),
			TeacherName:       teacherName,
			AssistantIDs:      stringIDsFromInt64(plan.AssistantIDs),
			AssistantNames:    plan.AssistantNames,
			ClassroomID:       emptyStringIfZero(classroomID),
			ClassroomName:     classroomName,
			LessonDate:        plan.LessonDate.Format("2006-01-02"),
			StartAt:           plan.StartAt,
			EndAt:             plan.EndAt,
			Status:            model.TeachingScheduleStatusActive,
		})
	}
	if err := repo.saveTeachingScheduleBatchMetaTx(
		ctx,
		tx,
		instID,
		operatorID,
		batchNo,
		model.TeachingClassTypeOneToOne,
		base.ClassID,
		createdScheduleIDs,
		dto.BatchMeta,
	); err != nil {
		return model.CreateOneToOneSchedulesResult{}, err
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
	fallbackTeacherID, err := parseOptionalPositiveID(dto.TeacherID)
	if err != nil {
		return model.TeachingScheduleValidationResult{}, errors.New("请选择上课教师")
	}
	fallbackClassroomID, err := parseOptionalPositiveID(dto.ClassroomID)
	if err != nil {
		return model.TeachingScheduleValidationResult{}, errors.New("classroomId 无效")
	}
	if len(dto.Schedules) == 0 {
		return model.TeachingScheduleValidationResult{}, errors.New("请至少选择一节日程")
	}
	assistantIDs := parseStringIDs(dto.AssistantIDs)
	excludeIDs := parseStringIDs(dto.ExcludeIDs)

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

	plans, err := normalizeCreateSchedulePlans(dto.Schedules, fallbackTeacherID, fallbackClassroomID, assistantIDs)
	if err != nil {
		return model.TeachingScheduleValidationResult{}, err
	}
	teacherIDs := collectPlanTeacherIDs(plans)
	if len(teacherIDs) == 0 {
		return model.TeachingScheduleValidationResult{}, errors.New("请选择上课教师")
	}

	if n, err := repo.CountInstUsersByIDs(ctx, instID, teacherIDs); err != nil || n != len(teacherIDs) {
		if err != nil {
			return model.TeachingScheduleValidationResult{}, err
		}
		return model.TeachingScheduleValidationResult{}, errors.New("上课教师无效")
	}
	planAssistantIDs := collectPlanAssistantIDs(plans)
	if len(planAssistantIDs) > 0 {
		if n, err := repo.CountInstUsersByIDs(ctx, instID, planAssistantIDs); err != nil || n != len(planAssistantIDs) {
			if err != nil {
				return model.TeachingScheduleValidationResult{}, err
			}
			return model.TeachingScheduleValidationResult{}, errors.New("存在无效的上课助教")
		}
	}

	classroomNames, err := repo.resolveClassroomNamesTx(ctx, tx, instID, collectPlanClassroomIDs(plans))
	if err != nil {
		return model.TeachingScheduleValidationResult{}, err
	}

	teacherNames := repo.resolveTeacherNames(ctx, teacherIDs)
	assistantNameMap := repo.resolveTeacherNames(ctx, planAssistantIDs)
	for i := range plans {
		plans[i].AssistantNames = compactStrings(func() []string {
			names := make([]string, 0, len(plans[i].AssistantIDs))
			for _, id := range plans[i].AssistantIDs {
				if name := strings.TrimSpace(assistantNameMap[id]); name != "" && name != "-" {
					names = append(names, name)
				}
			}
			return names
		}())
	}
	slots := plansToSlots(plans)
	teacherConflictRows, err := repo.listScheduleConflictDetailsByFieldValuesTx(ctx, tx, instID, "teacher_id", teacherIDs, slots, "", excludeIDs)
	if err != nil {
		return model.TeachingScheduleValidationResult{}, err
	}
	classroomConflictRows, err := repo.listScheduleConflictDetailsByFieldValuesTx(ctx, tx, instID, "classroom_id", collectPlanClassroomIDs(plans), slots, "", excludeIDs)
	if err != nil {
		return model.TeachingScheduleValidationResult{}, err
	}
	studentConflictRows, err := repo.listScheduleConflictDetailsByFieldValuesTx(ctx, tx, instID, "student_id", []int64{base.StudentID}, slots, "", excludeIDs)
	if err != nil {
		return model.TeachingScheduleValidationResult{}, err
	}
	assistantConflictRows, err := repo.listScheduleConflictDetailsByAssistantsTx(ctx, tx, instID, planAssistantIDs, slots, "", excludeIDs)
	if err != nil {
		return model.TeachingScheduleValidationResult{}, err
	}

	return buildScheduleValidationResultFromPlans(
		base,
		plans,
		teacherNames,
		classroomNames,
		teacherConflictRows,
		classroomConflictRows,
		studentConflictRows,
		assistantConflictRows,
	), nil
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
	excludeIDs := parseStringIDs(dto.ExcludeIDs)
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

	studentConflicts, err := repo.listAvailabilityConflictsByStudentTx(ctx, tx, instID, base.StudentID, startDate, endDate, excludeIDs)
	if err != nil {
		return model.OneToOneScheduleAvailabilityResult{}, err
	}
	teacherConflicts, err := repo.listAvailabilityConflictsByTeachersTx(ctx, tx, instID, teacherIDs, startDate, endDate, excludeIDs)
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

func (repo *Repository) CheckAssistantScheduleAvailability(ctx context.Context, instID int64, dto model.CheckAssistantScheduleAvailabilityDTO) (model.AssistantScheduleAvailabilityResult, error) {
	classID, err := strconv.ParseInt(strings.TrimSpace(dto.OneToOneID), 10, 64)
	if err != nil || classID <= 0 {
		return model.AssistantScheduleAvailabilityResult{}, errors.New("请选择1对1")
	}
	assistantIDs := parseStringIDs(dto.AssistantIDs)
	if len(assistantIDs) == 0 {
		return model.AssistantScheduleAvailabilityResult{}, errors.New("请至少选择一个上课助教")
	}
	if len(dto.Schedules) == 0 {
		return model.AssistantScheduleAvailabilityResult{}, errors.New("请至少选择一个上课时段")
	}
	if len(dto.Schedules) > 2000 {
		return model.AssistantScheduleAvailabilityResult{}, errors.New("待检测时段过多，请缩小时间范围后重试")
	}

	normalized, err := normalizeAssistantAvailabilityScheduleSlots(dto.Schedules)
	if err != nil {
		return model.AssistantScheduleAvailabilityResult{}, err
	}
	excludeIDs := parseStringIDs(dto.ExcludeIDs)
	if len(normalized) == 0 {
		return model.AssistantScheduleAvailabilityResult{}, errors.New("请至少选择一个上课时段")
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return model.AssistantScheduleAvailabilityResult{}, err
	}
	defer tx.Rollback()

	base, err := repo.GetOneToOneScheduleCreateContextTx(ctx, tx, instID, classID)
	if err != nil {
		return model.AssistantScheduleAvailabilityResult{}, err
	}
	if base.Status != model.TeachingClassStatusActive {
		return buildUnavailableAssistantAvailabilityResult(assistantIDs, repo.resolveTeacherNames(ctx, assistantIDs), "当前1对1已结班，暂不可排课"), nil
	}
	if base.ClassStudentStatus != model.TeachingClassStudentStatusStudying {
		return buildUnavailableAssistantAvailabilityResult(assistantIDs, repo.resolveTeacherNames(ctx, assistantIDs), "当前1对1学员状态不允许排课"), nil
	}
	if n, err := repo.CountInstUsersByIDs(ctx, instID, assistantIDs); err != nil || n != len(assistantIDs) {
		if err != nil {
			return model.AssistantScheduleAvailabilityResult{}, err
		}
		return model.AssistantScheduleAvailabilityResult{}, errors.New("存在无效的上课助教")
	}

	startDate, endDate := scheduleSlotsDateRange(normalized)
	conflicts, err := repo.listAvailabilityConflictsByAssistantsTx(ctx, tx, instID, assistantIDs, startDate, endDate, excludeIDs)
	if err != nil {
		return model.AssistantScheduleAvailabilityResult{}, err
	}

	assistantNames := repo.resolveTeacherNames(ctx, assistantIDs)
	result := model.AssistantScheduleAvailabilityResult{
		Items: make([]model.AssistantScheduleAvailabilityItem, 0, len(assistantIDs)),
	}
	for _, assistantID := range assistantIDs {
		assistantConflictRows := make([]scheduleAvailabilityConflictRow, 0)
		for _, row := range conflicts {
			if row.TeacherID == assistantID || stringSliceHasAnyID(row.AssistantIDs, map[int64]struct{}{assistantID: {}}) {
				assistantConflictRows = append(assistantConflictRows, row)
			}
		}

		existingMap := make(map[int64]model.TeachingScheduleConflictItem)
		conflictTypes := make([]string, 0, 1)
		for _, row := range assistantConflictRows {
			if !availabilityRowOverlapsAnySlot(row, normalized) {
				continue
			}
			appendAvailabilityConflict(existingMap, row, "助教")
			if !containsString(conflictTypes, "助教") {
				conflictTypes = append(conflictTypes, "助教")
			}
		}

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

		item := model.AssistantScheduleAvailabilityItem{
			AssistantID:       strconv.FormatInt(assistantID, 10),
			AssistantName:     firstNonEmptyString(assistantNames[assistantID], "-"),
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

func (repo *Repository) GetTeachingScheduleBatchDetail(ctx context.Context, instID int64, query model.TeachingScheduleBatchDetailQueryDTO) (model.TeachingScheduleBatchDetailVO, error) {
	items, err := repo.ListTeachingSchedules(ctx, instID, model.TeachingScheduleListQueryDTO{
		BatchNo: query.BatchNo,
		IDs:     query.IDs,
	})
	if err != nil {
		return model.TeachingScheduleBatchDetailVO{}, err
	}
	if len(items) == 0 {
		return model.TeachingScheduleBatchDetailVO{}, errors.New("未找到该批次日程")
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].LessonDate == items[j].LessonDate {
			if items[i].StartAt.Equal(items[j].StartAt) {
				return items[i].ID < items[j].ID
			}
			return items[i].StartAt.Before(items[j].StartAt)
		}
		return items[i].LessonDate < items[j].LessonDate
	})
	scheduleIDs := make([]int64, 0, len(items))
	for _, item := range items {
		id, parseErr := strconv.ParseInt(strings.TrimSpace(item.ID), 10, 64)
		if parseErr == nil && id > 0 {
			scheduleIDs = append(scheduleIDs, id)
		}
	}
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return model.TeachingScheduleBatchDetailVO{}, err
	}
	defer tx.Rollback()
	meta, err := repo.loadTeachingScheduleBatchMetaTx(ctx, tx, instID, items[0].BatchNo, scheduleIDs)
	if err != nil {
		return model.TeachingScheduleBatchDetailVO{}, err
	}
	first := items[0]
	return model.TeachingScheduleBatchDetailVO{
		BatchNo:           first.BatchNo,
		BatchSize:         len(items),
		ClassType:         first.ClassType,
		TeachingClassID:   first.TeachingClassID,
		TeachingClassName: first.TeachingClassName,
		StudentID:         first.StudentID,
		StudentName:       first.StudentName,
		LessonID:          first.LessonID,
		LessonName:        first.LessonName,
		BatchMeta:         meta,
		Schedules:         items,
	}, nil
}

func (repo *Repository) ReplaceTeachingScheduleBatch(ctx context.Context, instID, operatorID int64, dto model.TeachingScheduleBatchReplaceDTO) (model.CreateOneToOneSchedulesResult, error) {
	fallbackTeacherID, err := parseOptionalPositiveID(dto.TeacherID)
	if err != nil {
		return model.CreateOneToOneSchedulesResult{}, errors.New("请选择上课教师")
	}
	fallbackClassroomID, err := parseOptionalPositiveID(dto.ClassroomID)
	if err != nil {
		return model.CreateOneToOneSchedulesResult{}, errors.New("classroomId 无效")
	}
	if len(dto.Schedules) == 0 {
		return model.CreateOneToOneSchedulesResult{}, errors.New("请至少选择一节日程")
	}
	assistantIDs := parseStringIDs(dto.AssistantIDs)
	for _, assistantID := range assistantIDs {
		if assistantID == fallbackTeacherID {
			return model.CreateOneToOneSchedulesResult{}, errors.New("主教与助教不能为同一人")
		}
	}

	existing, err := repo.ListTeachingSchedules(ctx, instID, model.TeachingScheduleListQueryDTO{
		BatchNo: strings.TrimSpace(dto.BatchNo),
		IDs:     dto.IDs,
	})
	if err != nil {
		return model.CreateOneToOneSchedulesResult{}, err
	}
	if len(existing) == 0 {
		return model.CreateOneToOneSchedulesResult{}, errors.New("未找到可替换的日程")
	}
	allBatchSchedules := existing
	if batchNo := strings.TrimSpace(dto.BatchNo); batchNo != "" {
		allBatchSchedules, err = repo.ListTeachingSchedules(ctx, instID, model.TeachingScheduleListQueryDTO{
			BatchNo: batchNo,
		})
		if err != nil {
			return model.CreateOneToOneSchedulesResult{}, err
		}
	}

	classID, err := strconv.ParseInt(strings.TrimSpace(existing[0].TeachingClassID), 10, 64)
	if err != nil || classID <= 0 {
		return model.CreateOneToOneSchedulesResult{}, errors.New("当前批次缺少有效的1对1信息")
	}
	if raw := strings.TrimSpace(dto.OneToOneID); raw != "" && raw != strconv.FormatInt(classID, 10) {
		return model.CreateOneToOneSchedulesResult{}, errors.New("当前批次与所选1对1不一致")
	}
	for _, item := range existing {
		if item.ClassType != model.TeachingClassTypeOneToOne {
			return model.CreateOneToOneSchedulesResult{}, errors.New("仅支持按1对1规则重排当前批次")
		}
		if strings.TrimSpace(item.TeachingClassID) != strconv.FormatInt(classID, 10) {
			return model.CreateOneToOneSchedulesResult{}, errors.New("当前批次包含多个1对1，暂不支持合并重排")
		}
	}

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
		return model.CreateOneToOneSchedulesResult{}, errors.New("当前1对1已结班，暂不可调整批次规则")
	}
	if base.ClassStudentStatus != model.TeachingClassStudentStatusStudying {
		return model.CreateOneToOneSchedulesResult{}, errors.New("当前1对1学员状态不允许调整批次规则")
	}

	plans, err := normalizeCreateSchedulePlans(dto.Schedules, fallbackTeacherID, fallbackClassroomID, assistantIDs)
	if err != nil {
		return model.CreateOneToOneSchedulesResult{}, err
	}
	applyCreateScheduleConflictAllowances(plans, dto.AllowStudentConflict, dto.AllowClassroomConflict)
	teacherIDs := collectPlanTeacherIDs(plans)
	if len(teacherIDs) == 0 {
		return model.CreateOneToOneSchedulesResult{}, errors.New("请选择上课教师")
	}
	if n, err := repo.CountInstUsersByIDs(ctx, instID, teacherIDs); err != nil || n != len(teacherIDs) {
		if err != nil {
			return model.CreateOneToOneSchedulesResult{}, err
		}
		return model.CreateOneToOneSchedulesResult{}, errors.New("上课教师无效")
	}
	planAssistantIDs := collectPlanAssistantIDs(plans)
	if len(planAssistantIDs) > 0 {
		if n, err := repo.CountInstUsersByIDs(ctx, instID, planAssistantIDs); err != nil || n != len(planAssistantIDs) {
			if err != nil {
				return model.CreateOneToOneSchedulesResult{}, err
			}
			return model.CreateOneToOneSchedulesResult{}, errors.New("存在无效的上课助教")
		}
	}

	classroomNames, err := repo.resolveClassroomNamesTx(ctx, tx, instID, collectPlanClassroomIDs(plans))
	if err != nil {
		return model.CreateOneToOneSchedulesResult{}, err
	}

	teacherNames := repo.resolveTeacherNames(ctx, teacherIDs)
	assistantNameMap := repo.resolveTeacherNames(ctx, planAssistantIDs)
	for i := range plans {
		plans[i].AssistantNames = compactStrings(func() []string {
			names := make([]string, 0, len(plans[i].AssistantIDs))
			for _, id := range plans[i].AssistantIDs {
				if name := strings.TrimSpace(assistantNameMap[id]); name != "" && name != "-" {
					names = append(names, name)
				}
			}
			return names
		}())
	}

	existingIDs := make([]int64, 0, len(existing))
	for _, item := range existing {
		id, parseErr := strconv.ParseInt(strings.TrimSpace(item.ID), 10, 64)
		if parseErr == nil && id > 0 {
			existingIDs = append(existingIDs, id)
		}
	}
	allBatchIDs := make([]int64, 0, len(allBatchSchedules))
	for _, item := range allBatchSchedules {
		id, parseErr := strconv.ParseInt(strings.TrimSpace(item.ID), 10, 64)
		if parseErr == nil && id > 0 {
			allBatchIDs = append(allBatchIDs, id)
		}
	}
	targetIDSet := make(map[int64]struct{}, len(existingIDs))
	for _, id := range existingIDs {
		targetIDSet[id] = struct{}{}
	}
	untouchedIDs := make([]int64, 0, len(allBatchIDs))
	for _, id := range allBatchIDs {
		if _, ok := targetIDSet[id]; ok {
			continue
		}
		untouchedIDs = append(untouchedIDs, id)
	}

	teacherConflictsByPlan, err := repo.listTeacherConflictsByPlanTx(ctx, tx, instID, plans, existingIDs)
	if err != nil {
		return model.CreateOneToOneSchedulesResult{}, err
	}
	classroomConflictsByPlan, err := repo.listClassroomConflictsByPlanTx(ctx, tx, instID, plans, existingIDs)
	if err != nil {
		return model.CreateOneToOneSchedulesResult{}, err
	}
	studentConflicts, err := repo.listScheduleConflictDetailsTx(ctx, tx, instID, "student_id", base.StudentID, plansToSlots(plans), "", existingIDs)
	if err != nil {
		return model.CreateOneToOneSchedulesResult{}, err
	}
	assistantConflictsByPlan, err := repo.listAssistantConflictsByPlanTx(ctx, tx, instID, plans, existingIDs)
	if err != nil {
		return model.CreateOneToOneSchedulesResult{}, err
	}

	hardConflictTypes := make(map[string]struct{})
	for _, plan := range plans {
		key := schedulePlanKey(plan)
		if len(teacherConflictsByPlan[key]) > 0 {
			hardConflictTypes["老师"] = struct{}{}
		}
		if len(classroomConflictsByPlan[key]) > 0 && !plan.AllowClassroomConflict {
			hardConflictTypes["教室"] = struct{}{}
		}
		if slotHasConflict(normalizedScheduleSlot{
			LessonDate: plan.LessonDate,
			StartAt:    plan.StartAt,
			EndAt:      plan.EndAt,
		}, studentConflicts) && !plan.AllowStudentConflict {
			hardConflictTypes["学员"] = struct{}{}
		}
		if len(assistantConflictsByPlan[key]) > 0 {
			hardConflictTypes["助教"] = struct{}{}
		}
	}
	if len(hardConflictTypes) > 0 {
		conflictTypes := make([]string, 0, len(hardConflictTypes))
		for key := range hardConflictTypes {
			conflictTypes = append(conflictTypes, key)
		}
		sort.Strings(conflictTypes)
		return model.CreateOneToOneSchedulesResult{}, errors.New(buildConflictSummaryMessage(conflictTypes))
	}

	if len(existingIDs) > 0 {
		args := append([]any{
			model.TeachingScheduleStatusCanceled,
			operatorID,
			instID,
			model.TeachingScheduleStatusActive,
		}, int64SliceToAny(existingIDs)...)
		if _, err := tx.ExecContext(ctx, `
			UPDATE teaching_schedule
			SET del_flag = 1,
			    status = ?,
			    update_id = ?,
			    update_time = NOW()
			WHERE inst_id = ?
			  AND del_flag = 0
			  AND status = ?
			  AND id IN (`+sqlPlaceholders(len(existingIDs))+`)
		`, args...); err != nil {
			return model.CreateOneToOneSchedulesResult{}, err
		}
	}

	batchNo := strings.TrimSpace(existing[0].BatchNo)
	if batchNo != "" && len(untouchedIDs) > 0 {
		if err := repo.deleteTeachingScheduleBatchMetaTx(ctx, tx, instID, batchNo, nil); err != nil {
			return model.CreateOneToOneSchedulesResult{}, err
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE teaching_schedule
			SET batch_size = ?,
			    update_id = ?,
			    update_time = NOW()
			WHERE inst_id = ?
			  AND del_flag = 0
			  AND id IN (`+sqlPlaceholders(len(untouchedIDs))+`)
		`, append([]any{len(untouchedIDs), operatorID, instID}, int64SliceToAny(untouchedIDs)...)...); err != nil {
			return model.CreateOneToOneSchedulesResult{}, err
		}
		batchNo = ""
	}
	if batchNo == "" && len(plans) > 1 {
		batchNo = fmt.Sprintf("BATCH-%d", time.Now().UnixNano())
	}

	result := model.CreateOneToOneSchedulesResult{
		BatchNo: batchNo,
		Count:   len(plans),
		List:    make([]model.TeachingScheduleVO, 0, len(plans)),
	}
	createdScheduleIDs := make([]int64, 0, len(plans))

	for _, plan := range plans {
		teacherName := firstNonEmptyString(teacherNames[plan.TeacherID], "-")
		classroomName := classroomNames[plan.ClassroomID]
		classroomID := plan.ClassroomID
		assistantIDsJSON, _ := json.Marshal(stringIDsFromInt64(plan.AssistantIDs))
		assistantNamesJSON, _ := json.Marshal(plan.AssistantNames)
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
			plan.TeacherID,
			teacherName,
			nullJSONBytes(assistantIDsJSON),
			nullJSONBytes(assistantNamesJSON),
			classroomID,
			classroomName,
			plan.LessonDate.Format("2006-01-02"),
			plan.StartAt,
			plan.EndAt,
			batchNo,
			len(plans),
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
		createdScheduleIDs = append(createdScheduleIDs, id)
		result.List = append(result.List, model.TeachingScheduleVO{
			ID:                strconv.FormatInt(id, 10),
			BatchNo:           batchNo,
			BatchSize:         len(plans),
			ClassType:         model.TeachingClassTypeOneToOne,
			TeachingClassID:   strconv.FormatInt(base.ClassID, 10),
			TeachingClassName: base.ClassName,
			StudentID:         strconv.FormatInt(base.StudentID, 10),
			StudentName:       base.StudentName,
			LessonID:          strconv.FormatInt(base.LessonID, 10),
			LessonName:        base.LessonName,
			TeacherID:         strconv.FormatInt(plan.TeacherID, 10),
			TeacherName:       teacherName,
			AssistantIDs:      stringIDsFromInt64(plan.AssistantIDs),
			AssistantNames:    plan.AssistantNames,
			ClassroomID:       emptyStringIfZero(classroomID),
			ClassroomName:     classroomName,
			LessonDate:        plan.LessonDate.Format("2006-01-02"),
			StartAt:           plan.StartAt,
			EndAt:             plan.EndAt,
			Status:            model.TeachingScheduleStatusActive,
		})
	}
	if err := repo.saveTeachingScheduleBatchMetaTx(
		ctx,
		tx,
		instID,
		operatorID,
		batchNo,
		model.TeachingClassTypeOneToOne,
		base.ClassID,
		createdScheduleIDs,
		dto.BatchMeta,
	); err != nil {
		return model.CreateOneToOneSchedulesResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return model.CreateOneToOneSchedulesResult{}, err
	}
	return result, nil
}

func (repo *Repository) BatchUpdateTeachingSchedules(ctx context.Context, instID, operatorID int64, dto model.TeachingScheduleBatchUpdateDTO) error {
	teacherID, err := strconv.ParseInt(strings.TrimSpace(dto.TeacherID), 10, 64)
	if err != nil || teacherID <= 0 {
		return errors.New("请选择上课教师")
	}
	assistantIDs := parseStringIDs(dto.AssistantIDs)
	for _, assistantID := range assistantIDs {
		if assistantID == teacherID {
			return errors.New("主教与助教不能为同一人")
		}
	}
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
	targetLessonDate := strings.TrimSpace(dto.LessonDate)
	if targetLessonDate != "" && len(schedules) > 1 {
		return errors.New("批量修改暂不支持统一调整日期")
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
		lessonDate := item.LessonDate.Format("2006-01-02")
		if targetLessonDate != "" {
			lessonDate = targetLessonDate
		}
		startAt, endAt, err := buildScheduleDateTime(lessonDate, dto.StartTime, dto.EndTime)
		if err != nil {
			return err
		}
		updatedSlots = append(updatedSlots, normalizedScheduleSlot{
			LessonDate: startOfDay(startAt),
			StartAt:    startAt,
			EndAt:      endAt,
		})
		excludeIDs = append(excludeIDs, item.ID)
	}

	if err := repo.validateTeachingScheduleConflictsTx(ctx, tx, instID, teacherID, classroomID, updatedSlots, strings.TrimSpace(dto.BatchNo), excludeIDs, false); err != nil {
		return err
	}

	for index, item := range schedules {
		slot := updatedSlots[index]
		if item.ClassType == model.TeachingClassTypeOneToOne {
			if item.StudentID > 0 {
				studentConflicts, err := repo.listScheduleConflictDetailsTx(ctx, tx, instID, "student_id", item.StudentID, []normalizedScheduleSlot{slot}, strings.TrimSpace(dto.BatchNo), excludeIDs)
				if err != nil {
					return err
				}
				if len(studentConflicts) > 0 {
					return fmt.Errorf("学员在 %s %s-%s 已有日程冲突", slot.LessonDate.Format("2006-01-02"), slot.StartAt.Format("15:04"), slot.EndAt.Format("15:04"))
				}
			}
			if len(assistantIDs) > 0 {
				assistantConflicts, err := repo.listScheduleConflictDetailsByAssistantsTx(ctx, tx, instID, assistantIDs, []normalizedScheduleSlot{slot}, strings.TrimSpace(dto.BatchNo), excludeIDs)
				if err != nil {
					return err
				}
				if len(assistantConflicts) > 0 {
					return fmt.Errorf("助教在 %s %s-%s 已有日程冲突", slot.LessonDate.Format("2006-01-02"), slot.StartAt.Format("15:04"), slot.EndAt.Format("15:04"))
				}
			}
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE teaching_schedule
			SET teacher_id = ?, teacher_name = ?, assistant_ids_json = ?, assistant_names_json = ?,
			    classroom_id = ?, classroom_name = ?, lesson_date = ?, lesson_start_at = ?, lesson_end_at = ?,
			    update_id = ?, update_time = NOW()
			WHERE id = ? AND inst_id = ? AND del_flag = 0
		`,
			teacherID,
			teacherName,
			nullJSONBytes(assistantIDsJSON),
			nullJSONBytes(assistantNamesJSON),
			classroomID,
			classroomName,
			slot.LessonDate.Format("2006-01-02"),
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

type normalizedSchedulePlan struct {
	LessonDate             time.Time
	StartAt                time.Time
	EndAt                  time.Time
	TeacherID              int64
	AssistantIDs           []int64
	AssistantNames         []string
	ClassroomID            int64
	AllowStudentConflict   bool
	AllowClassroomConflict bool
}

func schedulePlanKey(plan normalizedSchedulePlan) string {
	return plan.LessonDate.Format("2006-01-02") + "|" +
		plan.StartAt.Format(time.RFC3339) + "|" +
		plan.EndAt.Format(time.RFC3339) + "|" +
		strconv.FormatInt(plan.TeacherID, 10) + "|" +
		strings.Join(stringIDsFromInt64(plan.AssistantIDs), ",") + "|" +
		strconv.FormatInt(plan.ClassroomID, 10)
}

func plansToSlots(plans []normalizedSchedulePlan) []normalizedScheduleSlot {
	result := make([]normalizedScheduleSlot, 0, len(plans))
	for _, plan := range plans {
		result = append(result, normalizedScheduleSlot{
			LessonDate: plan.LessonDate,
			StartAt:    plan.StartAt,
			EndAt:      plan.EndAt,
		})
	}
	return result
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
	AssistantIDs      []string
	AssistantNames    []string
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
	AssistantIDs      []string
	AssistantNames    []string
	ClassroomName     string
	LessonDate        time.Time
	StartAt           time.Time
	EndAt             time.Time
}

type teachingScheduleRow struct {
	ID         int64
	BatchNo    string
	ClassType  int
	StudentID  int64
	LessonDate time.Time
}

func buildTeachingScheduleBatchMetaKey(batchNo string, scheduleIDs []int64) (string, error) {
	batchNo = strings.TrimSpace(batchNo)
	if batchNo != "" {
		return "batch:" + batchNo, nil
	}
	if len(scheduleIDs) == 1 && scheduleIDs[0] > 0 {
		return fmt.Sprintf("schedule:%d", scheduleIDs[0]), nil
	}
	return "", errors.New("缺少批次元数据键")
}

func normalizeTeachingScheduleBatchMeta(meta *model.TeachingScheduleBatchMeta) *model.TeachingScheduleBatchMeta {
	if meta == nil {
		return nil
	}
	next := &model.TeachingScheduleBatchMeta{
		SchedulingMode:    strings.TrimSpace(meta.SchedulingMode),
		RepeatRule:        strings.TrimSpace(meta.RepeatRule),
		HolidayPolicy:     strings.TrimSpace(meta.HolidayPolicy),
		SelectedWeekdays:  compactStrings(meta.SelectedWeekdays),
		ScheduleStartDate: strings.TrimSpace(meta.ScheduleStartDate),
		FreeSelectedDates: compactStrings(meta.FreeSelectedDates),
		PlannedClassCount: meta.PlannedClassCount,
	}
	if next.PlannedClassCount < 0 {
		next.PlannedClassCount = 0
	}
	if next.SchedulingMode == "" &&
		next.RepeatRule == "" &&
		next.HolidayPolicy == "" &&
		len(next.SelectedWeekdays) == 0 &&
		next.ScheduleStartDate == "" &&
		len(next.FreeSelectedDates) == 0 &&
		next.PlannedClassCount == 0 {
		return nil
	}
	return next
}

func (repo *Repository) loadTeachingScheduleBatchMetaTx(ctx context.Context, tx *sql.Tx, instID int64, batchNo string, scheduleIDs []int64) (*model.TeachingScheduleBatchMeta, error) {
	batchKey, err := buildTeachingScheduleBatchMetaKey(batchNo, scheduleIDs)
	if err != nil {
		return nil, nil
	}
	var raw []byte
	err = tx.QueryRowContext(ctx, `
		SELECT meta_json
		FROM teaching_schedule_batch_meta
		WHERE inst_id = ?
		  AND batch_key = ?
		  AND del_flag = 0
		LIMIT 1
	`, instID, batchKey).Scan(&raw)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if len(raw) == 0 {
		return nil, nil
	}
	var meta model.TeachingScheduleBatchMeta
	if err := json.Unmarshal(raw, &meta); err != nil {
		return nil, err
	}
	return normalizeTeachingScheduleBatchMeta(&meta), nil
}

func (repo *Repository) saveTeachingScheduleBatchMetaTx(ctx context.Context, tx *sql.Tx, instID, operatorID int64, batchNo string, classType, teachingClassID int64, scheduleIDs []int64, meta *model.TeachingScheduleBatchMeta) error {
	normalized := normalizeTeachingScheduleBatchMeta(meta)
	if normalized == nil {
		return nil
	}
	batchKey, err := buildTeachingScheduleBatchMetaKey(batchNo, scheduleIDs)
	if err != nil {
		return nil
	}
	raw, err := json.Marshal(normalized)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO teaching_schedule_batch_meta (
			inst_id, batch_key, batch_no, class_type, teaching_class_id, meta_json,
			create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, NOW(), ?, NOW(), 0
		)
		ON DUPLICATE KEY UPDATE
			batch_no = VALUES(batch_no),
			class_type = VALUES(class_type),
			teaching_class_id = VALUES(teaching_class_id),
			meta_json = VALUES(meta_json),
			update_id = VALUES(update_id),
			update_time = NOW(),
			del_flag = 0
	`, instID, batchKey, strings.TrimSpace(batchNo), classType, teachingClassID, nullJSONBytes(raw), operatorID, operatorID)
	return err
}

func (repo *Repository) deleteTeachingScheduleBatchMetaTx(ctx context.Context, tx *sql.Tx, instID int64, batchNo string, scheduleIDs []int64) error {
	batchKey, err := buildTeachingScheduleBatchMetaKey(batchNo, scheduleIDs)
	if err != nil {
		return nil
	}
	_, err = tx.ExecContext(ctx, `
		UPDATE teaching_schedule_batch_meta
		SET del_flag = 1,
		    update_time = NOW()
		WHERE inst_id = ?
		  AND batch_key = ?
		  AND del_flag = 0
	`, instID, batchKey)
	return err
}

func (repo *Repository) loadScheduleConflictDetailByIDTx(ctx context.Context, tx *sql.Tx, instID, scheduleID int64) (scheduleConflictDetailRow, error) {
	var item scheduleConflictDetailRow
	var assistantIDsRaw []byte
	var assistantNamesRaw []byte
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
			assistant_ids_json,
			assistant_names_json,
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
		&assistantIDsRaw,
		&assistantNamesRaw,
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
	item.AssistantIDs = decodeJSONStringArray(assistantIDsRaw)
	item.AssistantNames = decodeJSONStringArray(assistantNamesRaw)
	return item, nil
}

func collectPlanTeacherIDs(plans []normalizedSchedulePlan) []int64 {
	seen := make(map[int64]struct{}, len(plans))
	result := make([]int64, 0, len(plans))
	for _, plan := range plans {
		if plan.TeacherID <= 0 {
			continue
		}
		if _, ok := seen[plan.TeacherID]; ok {
			continue
		}
		seen[plan.TeacherID] = struct{}{}
		result = append(result, plan.TeacherID)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func collectPlanClassroomIDs(plans []normalizedSchedulePlan) []int64 {
	seen := make(map[int64]struct{}, len(plans))
	result := make([]int64, 0, len(plans))
	for _, plan := range plans {
		if plan.ClassroomID <= 0 {
			continue
		}
		if _, ok := seen[plan.ClassroomID]; ok {
			continue
		}
		seen[plan.ClassroomID] = struct{}{}
		result = append(result, plan.ClassroomID)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func collectPlanAssistantIDs(plans []normalizedSchedulePlan) []int64 {
	seen := make(map[int64]struct{})
	result := make([]int64, 0)
	for _, plan := range plans {
		for _, assistantID := range plan.AssistantIDs {
			if assistantID <= 0 {
				continue
			}
			if _, ok := seen[assistantID]; ok {
				continue
			}
			seen[assistantID] = struct{}{}
			result = append(result, assistantID)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func (repo *Repository) resolveTeacherNames(ctx context.Context, teacherIDs []int64) map[int64]string {
	result := make(map[int64]string, len(teacherIDs))
	for _, teacherID := range teacherIDs {
		copyID := teacherID
		result[teacherID] = strings.TrimSpace(repo.GetStaffNameByID(ctx, &copyID))
	}
	return result
}

func (repo *Repository) resolveClassroomNamesTx(ctx context.Context, tx *sql.Tx, instID int64, classroomIDs []int64) (map[int64]string, error) {
	result := make(map[int64]string, len(classroomIDs))
	for _, classroomID := range classroomIDs {
		_, name, _, err := repo.resolveClassroomByIDTx(ctx, tx, instID, strconv.FormatInt(classroomID, 10))
		if err != nil {
			return nil, err
		}
		result[classroomID] = name
	}
	return result, nil
}

func (repo *Repository) listTeacherConflictsByPlanTx(ctx context.Context, tx *sql.Tx, instID int64, plans []normalizedSchedulePlan, excludeIDs []int64) (map[string][]scheduleConflictDetailRow, error) {
	result := make(map[string][]scheduleConflictDetailRow, len(plans))
	for _, plan := range plans {
		if plan.TeacherID <= 0 {
			continue
		}
		rows, err := repo.listScheduleConflictDetailsTx(ctx, tx, instID, "teacher_id", plan.TeacherID, []normalizedScheduleSlot{{
			LessonDate: plan.LessonDate,
			StartAt:    plan.StartAt,
			EndAt:      plan.EndAt,
		}}, "", excludeIDs)
		if err != nil {
			return nil, err
		}
		result[schedulePlanKey(plan)] = rows
	}
	return result, nil
}

func (repo *Repository) listClassroomConflictsByPlanTx(ctx context.Context, tx *sql.Tx, instID int64, plans []normalizedSchedulePlan, excludeIDs []int64) (map[string][]scheduleConflictDetailRow, error) {
	result := make(map[string][]scheduleConflictDetailRow, len(plans))
	for _, plan := range plans {
		if plan.ClassroomID <= 0 {
			continue
		}
		rows, err := repo.listScheduleConflictDetailsTx(ctx, tx, instID, "classroom_id", plan.ClassroomID, []normalizedScheduleSlot{{
			LessonDate: plan.LessonDate,
			StartAt:    plan.StartAt,
			EndAt:      plan.EndAt,
		}}, "", excludeIDs)
		if err != nil {
			return nil, err
		}
		result[schedulePlanKey(plan)] = rows
	}
	return result, nil
}

func (repo *Repository) listAssistantConflictsByPlanTx(ctx context.Context, tx *sql.Tx, instID int64, plans []normalizedSchedulePlan, excludeIDs []int64) (map[string][]scheduleConflictDetailRow, error) {
	result := make(map[string][]scheduleConflictDetailRow, len(plans))
	for _, plan := range plans {
		if len(plan.AssistantIDs) == 0 {
			continue
		}
		rows, err := repo.listScheduleConflictDetailsByAssistantsTx(ctx, tx, instID, plan.AssistantIDs, []normalizedScheduleSlot{{
			LessonDate: plan.LessonDate,
			StartAt:    plan.StartAt,
			EndAt:      plan.EndAt,
		}}, "", excludeIDs)
		if err != nil {
			return nil, err
		}
		result[schedulePlanKey(plan)] = rows
	}
	return result, nil
}

func (repo *Repository) listScheduleConflictDetailsByFieldValuesTx(ctx context.Context, tx *sql.Tx, instID int64, field string, fieldValues []int64, slots []normalizedScheduleSlot, excludeBatchNo string, excludeIDs []int64) ([]scheduleConflictDetailRow, error) {
	if len(fieldValues) == 0 || len(slots) == 0 {
		return []scheduleConflictDetailRow{}, nil
	}

	startDate, endDate := scheduleSlotsDateRange(slots)
	if startDate == "" || endDate == "" {
		return []scheduleConflictDetailRow{}, nil
	}

	uniqueFieldValues := make([]int64, 0, len(fieldValues))
	seenFieldValues := make(map[int64]struct{}, len(fieldValues))
	for _, value := range fieldValues {
		if value <= 0 {
			continue
		}
		if _, ok := seenFieldValues[value]; ok {
			continue
		}
		seenFieldValues[value] = struct{}{}
		uniqueFieldValues = append(uniqueFieldValues, value)
	}
	if len(uniqueFieldValues) == 0 {
		return []scheduleConflictDetailRow{}, nil
	}

	filters := []string{
		"inst_id = ?",
		"del_flag = 0",
		"status = ?",
		field + " IN (" + sqlPlaceholders(len(uniqueFieldValues)) + ")",
		"lesson_date >= ?",
		"lesson_date <= ?",
	}
	args := []any{
		instID,
		model.TeachingScheduleStatusActive,
	}
	for _, value := range uniqueFieldValues {
		args = append(args, value)
	}
	args = append(args, startDate, endDate)
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
			assistant_ids_json,
			assistant_names_json,
			IFNULL(classroom_name, ''),
			lesson_date,
			lesson_start_at,
			lesson_end_at
		FROM teaching_schedule
		WHERE `+strings.Join(filters, " AND ")+`
		ORDER BY lesson_date ASC, lesson_start_at ASC, id ASC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]scheduleConflictDetailRow, 0, 32)
	seenRows := make(map[int64]struct{})
	for rows.Next() {
		var item scheduleConflictDetailRow
		var assistantIDsRaw []byte
		var assistantNamesRaw []byte
		if err := rows.Scan(
			&item.ID,
			&item.StudentID,
			&item.TeacherID,
			&item.ClassroomID,
			&item.ClassType,
			&item.TeachingClassName,
			&item.StudentName,
			&item.TeacherName,
			&assistantIDsRaw,
			&assistantNamesRaw,
			&item.ClassroomName,
			&item.LessonDate,
			&item.StartAt,
			&item.EndAt,
		); err != nil {
			return nil, err
		}
		if !scheduleRowOverlapsAnySlot(item.LessonDate, item.StartAt, item.EndAt, slots) {
			continue
		}
		if _, ok := seenRows[item.ID]; ok {
			continue
		}
		item.AssistantIDs = decodeJSONStringArray(assistantIDsRaw)
		item.AssistantNames = decodeJSONStringArray(assistantNamesRaw)
		seenRows[item.ID] = struct{}{}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (repo *Repository) listScheduleConflictDetailsByAssistantsTx(ctx context.Context, tx *sql.Tx, instID int64, assistantIDs []int64, slots []normalizedScheduleSlot, excludeBatchNo string, excludeIDs []int64) ([]scheduleConflictDetailRow, error) {
	if len(assistantIDs) == 0 || len(slots) == 0 {
		return []scheduleConflictDetailRow{}, nil
	}

	startDate, endDate := scheduleSlotsDateRange(slots)
	if startDate == "" || endDate == "" {
		return []scheduleConflictDetailRow{}, nil
	}

	candidateSet := make(map[int64]struct{}, len(assistantIDs))
	for _, id := range assistantIDs {
		if id > 0 {
			candidateSet[id] = struct{}{}
		}
	}

	filters := []string{
		"inst_id = ?",
		"del_flag = 0",
		"status = ?",
		"lesson_date >= ?",
		"lesson_date <= ?",
	}
	args := []any{
		instID,
		model.TeachingScheduleStatusActive,
		startDate,
		endDate,
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
			assistant_ids_json,
			assistant_names_json,
			IFNULL(classroom_name, ''),
			lesson_date,
			lesson_start_at,
			lesson_end_at
		FROM teaching_schedule
		WHERE `+strings.Join(filters, " AND ")+`
		ORDER BY lesson_date ASC, lesson_start_at ASC, id ASC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]scheduleConflictDetailRow, 0, 32)
	seen := make(map[int64]struct{})
	for rows.Next() {
		var item scheduleConflictDetailRow
		var assistantIDsRaw []byte
		var assistantNamesRaw []byte
		if err := rows.Scan(
			&item.ID,
			&item.StudentID,
			&item.TeacherID,
			&item.ClassroomID,
			&item.ClassType,
			&item.TeachingClassName,
			&item.StudentName,
			&item.TeacherName,
			&assistantIDsRaw,
			&assistantNamesRaw,
			&item.ClassroomName,
			&item.LessonDate,
			&item.StartAt,
			&item.EndAt,
		); err != nil {
			return nil, err
		}
		item.AssistantIDs = decodeJSONStringArray(assistantIDsRaw)
		item.AssistantNames = decodeJSONStringArray(assistantNamesRaw)
		if _, ok := candidateSet[item.TeacherID]; !ok && !stringSliceHasAnyID(item.AssistantIDs, candidateSet) {
			continue
		}
		if !scheduleRowOverlapsAnySlot(item.LessonDate, item.StartAt, item.EndAt, slots) {
			continue
		}
		if _, ok := seen[item.ID]; ok {
			continue
		}
		seen[item.ID] = struct{}{}
		result = append(result, item)
	}
	return result, rows.Err()
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
		SELECT
			id,
			IFNULL(batch_no, ''),
			IFNULL(class_type, 0),
			IFNULL(student_id, 0),
			lesson_date
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
		if err := rows.Scan(&item.ID, &item.BatchNo, &item.ClassType, &item.StudentID, &item.LessonDate); err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, rows.Err()
}

func (repo *Repository) validateTeachingScheduleConflictsTx(ctx context.Context, tx *sql.Tx, instID, teacherID, classroomID int64, slots []normalizedScheduleSlot, excludeBatchNo string, excludeIDs []int64, allowClassroomConflict bool) error {
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
		if classroomID > 0 && !allowClassroomConflict {
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
				assistant_ids_json,
				assistant_names_json,
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
			var assistantIDsRaw []byte
			var assistantNamesRaw []byte
			if err := rows.Scan(
				&item.ID,
				&item.StudentID,
				&item.TeacherID,
				&item.ClassroomID,
				&item.ClassType,
				&item.TeachingClassName,
				&item.StudentName,
				&item.TeacherName,
				&assistantIDsRaw,
				&assistantNamesRaw,
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
			item.AssistantIDs = decodeJSONStringArray(assistantIDsRaw)
			item.AssistantNames = decodeJSONStringArray(assistantNamesRaw)
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

func (repo *Repository) listAvailabilityConflictsByStudentTx(ctx context.Context, tx *sql.Tx, instID, studentID int64, startDate, endDate string, excludeIDs []int64) ([]scheduleAvailabilityConflictRow, error) {
	if studentID <= 0 || strings.TrimSpace(startDate) == "" || strings.TrimSpace(endDate) == "" {
		return []scheduleAvailabilityConflictRow{}, nil
	}
	filters := []string{
		"inst_id = ?",
		"del_flag = 0",
		"status = ?",
		"student_id = ?",
		"lesson_date >= ?",
		"lesson_date <= ?",
	}
	args := []any{instID, model.TeachingScheduleStatusActive, studentID, startDate, endDate}
	if len(excludeIDs) > 0 {
		filters = append(filters, "id NOT IN ("+sqlPlaceholders(len(excludeIDs))+")")
		for _, id := range excludeIDs {
			args = append(args, id)
		}
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT
			id,
			IFNULL(teacher_id, 0),
			IFNULL(class_type, 0),
			IFNULL(teaching_class_name, ''),
			IFNULL(student_name, ''),
			IFNULL(teacher_name, ''),
			assistant_ids_json,
			assistant_names_json,
			IFNULL(classroom_name, ''),
			lesson_date,
			lesson_start_at,
			lesson_end_at
		FROM teaching_schedule
		WHERE `+strings.Join(filters, " AND ")+`
		ORDER BY lesson_date ASC, lesson_start_at ASC, id ASC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]scheduleAvailabilityConflictRow, 0, 16)
	for rows.Next() {
		var item scheduleAvailabilityConflictRow
		var assistantIDsRaw []byte
		var assistantNamesRaw []byte
		if err := rows.Scan(
			&item.ID,
			&item.TeacherID,
			&item.ClassType,
			&item.TeachingClassName,
			&item.StudentName,
			&item.TeacherName,
			&assistantIDsRaw,
			&assistantNamesRaw,
			&item.ClassroomName,
			&item.LessonDate,
			&item.StartAt,
			&item.EndAt,
		); err != nil {
			return nil, err
		}
		item.AssistantIDs = decodeJSONStringArray(assistantIDsRaw)
		item.AssistantNames = decodeJSONStringArray(assistantNamesRaw)
		result = append(result, item)
	}
	return result, rows.Err()
}

func (repo *Repository) listAvailabilityConflictsByTeachersTx(ctx context.Context, tx *sql.Tx, instID int64, teacherIDs []int64, startDate, endDate string, excludeIDs []int64) ([]scheduleAvailabilityConflictRow, error) {
	if len(teacherIDs) == 0 || strings.TrimSpace(startDate) == "" || strings.TrimSpace(endDate) == "" {
		return []scheduleAvailabilityConflictRow{}, nil
	}

	args := []any{instID, model.TeachingScheduleStatusActive}
	placeholders := sqlPlaceholders(len(teacherIDs))
	for _, teacherID := range teacherIDs {
		args = append(args, teacherID)
	}
	args = append(args, startDate, endDate)
	filters := []string{
		"inst_id = ?",
		"del_flag = 0",
		"status = ?",
		"teacher_id IN (" + placeholders + ")",
		"lesson_date >= ?",
		"lesson_date <= ?",
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
			IFNULL(teacher_id, 0),
			IFNULL(class_type, 0),
			IFNULL(teaching_class_name, ''),
			IFNULL(student_name, ''),
			IFNULL(teacher_name, ''),
			assistant_ids_json,
			assistant_names_json,
			IFNULL(classroom_name, ''),
			lesson_date,
			lesson_start_at,
			lesson_end_at
		FROM teaching_schedule
		WHERE `+strings.Join(filters, " AND ")+`
		ORDER BY lesson_date ASC, lesson_start_at ASC, id ASC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]scheduleAvailabilityConflictRow, 0, 32)
	for rows.Next() {
		var item scheduleAvailabilityConflictRow
		var assistantIDsRaw []byte
		var assistantNamesRaw []byte
		if err := rows.Scan(
			&item.ID,
			&item.TeacherID,
			&item.ClassType,
			&item.TeachingClassName,
			&item.StudentName,
			&item.TeacherName,
			&assistantIDsRaw,
			&assistantNamesRaw,
			&item.ClassroomName,
			&item.LessonDate,
			&item.StartAt,
			&item.EndAt,
		); err != nil {
			return nil, err
		}
		item.AssistantIDs = decodeJSONStringArray(assistantIDsRaw)
		item.AssistantNames = decodeJSONStringArray(assistantNamesRaw)
		result = append(result, item)
	}
	return result, rows.Err()
}

func (repo *Repository) listAvailabilityConflictsByAssistantsTx(ctx context.Context, tx *sql.Tx, instID int64, assistantIDs []int64, startDate, endDate string, excludeIDs []int64) ([]scheduleAvailabilityConflictRow, error) {
	if len(assistantIDs) == 0 || strings.TrimSpace(startDate) == "" || strings.TrimSpace(endDate) == "" {
		return []scheduleAvailabilityConflictRow{}, nil
	}

	candidateSet := make(map[int64]struct{}, len(assistantIDs))
	for _, id := range assistantIDs {
		if id > 0 {
			candidateSet[id] = struct{}{}
		}
	}

	filters := []string{
		"inst_id = ?",
		"del_flag = 0",
		"status = ?",
		"lesson_date >= ?",
		"lesson_date <= ?",
	}
	args := []any{instID, model.TeachingScheduleStatusActive, startDate, endDate}
	if len(excludeIDs) > 0 {
		filters = append(filters, "id NOT IN ("+sqlPlaceholders(len(excludeIDs))+")")
		for _, id := range excludeIDs {
			args = append(args, id)
		}
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT
			id,
			IFNULL(teacher_id, 0),
			IFNULL(class_type, 0),
			IFNULL(teaching_class_name, ''),
			IFNULL(student_name, ''),
			IFNULL(teacher_name, ''),
			assistant_ids_json,
			assistant_names_json,
			IFNULL(classroom_name, ''),
			lesson_date,
			lesson_start_at,
			lesson_end_at
		FROM teaching_schedule
		WHERE `+strings.Join(filters, " AND ")+`
		ORDER BY lesson_date ASC, lesson_start_at ASC, id ASC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]scheduleAvailabilityConflictRow, 0, 32)
	for rows.Next() {
		var item scheduleAvailabilityConflictRow
		var assistantIDsRaw []byte
		var assistantNamesRaw []byte
		if err := rows.Scan(
			&item.ID,
			&item.TeacherID,
			&item.ClassType,
			&item.TeachingClassName,
			&item.StudentName,
			&item.TeacherName,
			&assistantIDsRaw,
			&assistantNamesRaw,
			&item.ClassroomName,
			&item.LessonDate,
			&item.StartAt,
			&item.EndAt,
		); err != nil {
			return nil, err
		}
		item.AssistantIDs = decodeJSONStringArray(assistantIDsRaw)
		item.AssistantNames = decodeJSONStringArray(assistantNamesRaw)
		if _, ok := candidateSet[item.TeacherID]; !ok && !stringSliceHasAnyID(item.AssistantIDs, candidateSet) {
			continue
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

func normalizeAssistantAvailabilityScheduleSlots(slots []model.AssistantScheduleAvailabilitySlotDTO) ([]normalizedScheduleSlot, error) {
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

func normalizeCreateSchedulePlans(slots []model.TeachingScheduleCreateSlotDTO, fallbackTeacherID, fallbackClassroomID int64, fallbackAssistantIDs []int64) ([]normalizedSchedulePlan, error) {
	result := make([]normalizedSchedulePlan, 0, len(slots))
	seen := make(map[string]struct{}, len(slots))
	for _, item := range slots {
		startAt, endAt, err := buildScheduleDateTime(item.LessonDate, item.StartTime, item.EndTime)
		if err != nil {
			return nil, err
		}
		teacherID, err := parseOptionalPositiveID(item.TeacherID)
		if err != nil {
			return nil, errors.New("存在无效的上课教师")
		}
		if teacherID <= 0 {
			teacherID = fallbackTeacherID
		}
		if teacherID <= 0 {
			return nil, errors.New("请选择上课教师")
		}
		classroomID, err := parseOptionalPositiveID(item.ClassroomID)
		if err != nil {
			return nil, errors.New("存在无效的教室")
		}
		if classroomID <= 0 {
			classroomID = fallbackClassroomID
		}
		assistantIDs := parseStringIDs(item.AssistantIDs)
		if len(assistantIDs) == 0 && len(fallbackAssistantIDs) > 0 {
			assistantIDs = append([]int64{}, fallbackAssistantIDs...)
		}

		key := strconv.FormatInt(teacherID, 10) + "|" +
			strings.Join(stringIDsFromInt64(assistantIDs), ",") + "|" +
			strconv.FormatInt(classroomID, 10) + "|" +
			startAt.Format(time.RFC3339) + "|" +
			endAt.Format(time.RFC3339)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, normalizedSchedulePlan{
			LessonDate:             startOfDay(startAt),
			StartAt:                startAt,
			EndAt:                  endAt,
			TeacherID:              teacherID,
			AssistantIDs:           assistantIDs,
			ClassroomID:            classroomID,
			AllowStudentConflict:   item.AllowStudentConflict,
			AllowClassroomConflict: item.AllowClassroomConflict,
		})
	}
	return result, nil
}

func applyCreateScheduleConflictAllowances(plans []normalizedSchedulePlan, allowStudentConflict bool, allowClassroomConflict bool) {
	for i := range plans {
		if allowStudentConflict {
			plans[i].AllowStudentConflict = true
		}
		if allowClassroomConflict {
			plans[i].AllowClassroomConflict = true
		}
	}
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

func parseOptionalPositiveID(raw string) (int64, error) {
	value := strings.TrimSpace(raw)
	if value == "" || value == "0" {
		return 0, nil
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id < 0 {
		return 0, errors.New("invalid id")
	}
	return id, nil
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
	assistantNames []string,
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
			Name:           base.ClassName,
			ClassTypeText:  "1对1日程",
			Date:           slot.LessonDate.Format("2006-01-02"),
			Week:           weekDisplay(slot.LessonDate),
			TimeText:       slot.StartAt.Format("15:04") + "~" + slot.EndAt.Format("15:04"),
			TeacherID:      "",
			TeacherName:    firstNonEmptyString(teacherName, "-"),
			AssistantNames: compactStrings(assistantNames),
			ClassroomName:  firstNonEmptyString(classroomName, "-"),
			StudentNames:   compactStrings([]string{base.StudentName}),
			ConflictTypes:  conflictTypes,
		})
	}

	existingMap := make(map[int64]model.TeachingScheduleConflictItem)
	appendExisting := func(row scheduleConflictDetailRow, conflictType string) {
		item, ok := existingMap[row.ID]
		if !ok {
			item = model.TeachingScheduleConflictItem{
				Name:           row.TeachingClassName,
				ClassTypeText:  scheduleClassTypeText(row.ClassType),
				Date:           row.LessonDate.Format("2006-01-02"),
				Week:           weekDisplay(row.LessonDate),
				TimeText:       row.StartAt.Format("15:04") + "~" + row.EndAt.Format("15:04"),
				TeacherID:      emptyStringIfZero(row.TeacherID),
				TeacherName:    firstNonEmptyString(row.TeacherName, "-"),
				AssistantNames: compactStrings(row.AssistantNames),
				ClassroomName:  firstNonEmptyString(row.ClassroomName, "-"),
				StudentNames:   compactStrings([]string{row.StudentName}),
				ConflictTypes:  []string{},
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

func buildScheduleConflictResultFromPlans(
	base model.OneToOneScheduleCreateContext,
	plans []normalizedSchedulePlan,
	teacherNames map[int64]string,
	classroomNames map[int64]string,
	teacherConflicts map[string][]scheduleConflictDetailRow,
	classroomConflicts map[string][]scheduleConflictDetailRow,
	studentConflicts []scheduleConflictDetailRow,
	assistantConflicts map[string][]scheduleConflictDetailRow,
) ([]model.TeachingScheduleConflictItem, []model.TeachingScheduleConflictItem, []string) {
	typeSet := make(map[string]struct{})
	current := make([]model.TeachingScheduleConflictItem, 0, len(plans))
	existingMap := make(map[int64]model.TeachingScheduleConflictItem)

	appendExisting := func(row scheduleConflictDetailRow, conflictType string) {
		item, ok := existingMap[row.ID]
		if !ok {
			item = model.TeachingScheduleConflictItem{
				Name:           row.TeachingClassName,
				ClassTypeText:  scheduleClassTypeText(row.ClassType),
				Date:           row.LessonDate.Format("2006-01-02"),
				Week:           weekDisplay(row.LessonDate),
				TimeText:       row.StartAt.Format("15:04") + "~" + row.EndAt.Format("15:04"),
				TeacherID:      emptyStringIfZero(row.TeacherID),
				TeacherName:    firstNonEmptyString(row.TeacherName, "-"),
				AssistantNames: compactStrings(row.AssistantNames),
				ClassroomName:  firstNonEmptyString(row.ClassroomName, "-"),
				StudentNames:   compactStrings([]string{row.StudentName}),
				ConflictTypes:  []string{},
			}
		}
		if !containsString(item.ConflictTypes, conflictType) {
			item.ConflictTypes = append(item.ConflictTypes, conflictType)
		}
		existingMap[row.ID] = item
	}

	for _, plan := range plans {
		key := schedulePlanKey(plan)
		conflictTypes := make([]string, 0, 3)

		if rows := teacherConflicts[key]; len(rows) > 0 {
			conflictTypes = append(conflictTypes, "老师")
			typeSet["老师"] = struct{}{}
			for _, row := range rows {
				appendExisting(row, "老师")
			}
		}
		if rows := classroomConflicts[key]; len(rows) > 0 {
			conflictTypes = append(conflictTypes, "教室")
			typeSet["教室"] = struct{}{}
			for _, row := range rows {
				appendExisting(row, "教室")
			}
		}

		slot := normalizedScheduleSlot{
			LessonDate: plan.LessonDate,
			StartAt:    plan.StartAt,
			EndAt:      plan.EndAt,
		}
		if slotHasConflict(slot, studentConflicts) {
			conflictTypes = append(conflictTypes, "学员")
			typeSet["学员"] = struct{}{}
			for _, row := range studentConflicts {
				if row.LessonDate.Format("2006-01-02") == slot.LessonDate.Format("2006-01-02") &&
					row.StartAt.Before(slot.EndAt) &&
					row.EndAt.After(slot.StartAt) {
					appendExisting(row, "学员")
				}
			}
		}
		if rows := assistantConflicts[key]; len(rows) > 0 {
			conflictTypes = append(conflictTypes, "助教")
			typeSet["助教"] = struct{}{}
			for _, row := range rows {
				appendExisting(row, "助教")
			}
		}

		sort.Strings(conflictTypes)
		current = append(current, model.TeachingScheduleConflictItem{
			Name:           base.ClassName,
			ClassTypeText:  "1对1日程",
			Date:           plan.LessonDate.Format("2006-01-02"),
			Week:           weekDisplay(plan.LessonDate),
			TimeText:       plan.StartAt.Format("15:04") + "~" + plan.EndAt.Format("15:04"),
			TeacherID:      emptyStringIfZero(plan.TeacherID),
			TeacherName:    firstNonEmptyString(teacherNames[plan.TeacherID], "-"),
			AssistantNames: compactStrings(plan.AssistantNames),
			ClassroomName:  firstNonEmptyString(classroomNames[plan.ClassroomID], "-"),
			StudentNames:   compactStrings([]string{base.StudentName}),
			ConflictTypes:  conflictTypes,
		})
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

	return current, existing, conflictTypes
}

func buildTeachingScheduleConflictItemFromRow(row scheduleConflictDetailRow) model.TeachingScheduleConflictItem {
	return model.TeachingScheduleConflictItem{
		Name:           row.TeachingClassName,
		ClassTypeText:  scheduleClassTypeText(row.ClassType),
		Date:           row.LessonDate.Format("2006-01-02"),
		Week:           weekDisplay(row.LessonDate),
		TimeText:       row.StartAt.Format("15:04") + "~" + row.EndAt.Format("15:04"),
		TeacherID:      emptyStringIfZero(row.TeacherID),
		TeacherName:    firstNonEmptyString(row.TeacherName, "-"),
		AssistantNames: compactStrings(row.AssistantNames),
		ClassroomName:  firstNonEmptyString(row.ClassroomName, "-"),
		StudentNames:   compactStrings([]string{row.StudentName}),
		ConflictTypes:  []string{},
	}
}

func appendConflictRowsToItemMap(existingMap map[int64]model.TeachingScheduleConflictItem, rows []scheduleConflictDetailRow, conflictType string) {
	for _, row := range rows {
		item, ok := existingMap[row.ID]
		if !ok {
			item = buildTeachingScheduleConflictItemFromRow(row)
		}
		if !containsString(item.ConflictTypes, conflictType) {
			item.ConflictTypes = append(item.ConflictTypes, conflictType)
		}
		existingMap[row.ID] = item
	}
}

func finalizeConflictItemMap(existingMap map[int64]model.TeachingScheduleConflictItem) []model.TeachingScheduleConflictItem {
	existing := make([]model.TeachingScheduleConflictItem, 0, len(existingMap))
	for _, item := range existingMap {
		sort.Strings(item.ConflictTypes)
		existing = append(existing, item)
	}
	sort.Slice(existing, func(i, j int) bool {
		if existing[i].Date == existing[j].Date {
			return existing[i].TimeText < existing[j].TimeText
		}
		return existing[i].Date < existing[j].Date
	})
	return existing
}

func groupScheduleConflictRowsByIDAndDate(rows []scheduleConflictDetailRow, resolveID func(scheduleConflictDetailRow) int64) map[int64]map[string][]scheduleConflictDetailRow {
	result := make(map[int64]map[string][]scheduleConflictDetailRow)
	for _, row := range rows {
		id := resolveID(row)
		if id <= 0 {
			continue
		}
		dateKey := row.LessonDate.Format("2006-01-02")
		if _, ok := result[id]; !ok {
			result[id] = make(map[string][]scheduleConflictDetailRow)
		}
		result[id][dateKey] = append(result[id][dateKey], row)
	}
	return result
}

func groupScheduleConflictRowsByDate(rows []scheduleConflictDetailRow) map[string][]scheduleConflictDetailRow {
	result := make(map[string][]scheduleConflictDetailRow)
	for _, row := range rows {
		dateKey := row.LessonDate.Format("2006-01-02")
		result[dateKey] = append(result[dateKey], row)
	}
	return result
}

func groupAssistantConflictRowsByCandidateIDAndDate(rows []scheduleConflictDetailRow, candidateSet map[int64]struct{}) map[int64]map[string][]scheduleConflictDetailRow {
	result := make(map[int64]map[string][]scheduleConflictDetailRow)
	for _, row := range rows {
		dateKey := row.LessonDate.Format("2006-01-02")
		attachIDs := make([]int64, 0, 1+len(row.AssistantIDs))
		if _, ok := candidateSet[row.TeacherID]; ok && row.TeacherID > 0 {
			attachIDs = append(attachIDs, row.TeacherID)
		}
		for _, raw := range row.AssistantIDs {
			assistantID, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
			if err != nil || assistantID <= 0 {
				continue
			}
			if _, ok := candidateSet[assistantID]; ok && !containsInt64(attachIDs, assistantID) {
				attachIDs = append(attachIDs, assistantID)
			}
		}
		for _, attachID := range attachIDs {
			if _, ok := result[attachID]; !ok {
				result[attachID] = make(map[string][]scheduleConflictDetailRow)
			}
			result[attachID][dateKey] = append(result[attachID][dateKey], row)
		}
	}
	return result
}

func filterOverlappingScheduleConflictRows(rows []scheduleConflictDetailRow, lessonDate, startAt, endAt time.Time) []scheduleConflictDetailRow {
	if len(rows) == 0 {
		return nil
	}
	result := make([]scheduleConflictDetailRow, 0, len(rows))
	for _, row := range rows {
		if row.LessonDate.Format("2006-01-02") != lessonDate.Format("2006-01-02") {
			continue
		}
		if row.StartAt.Before(endAt) && row.EndAt.After(startAt) {
			result = append(result, row)
		}
	}
	return result
}

func collectPlanAssistantConflictRows(plan normalizedSchedulePlan, assistantIndex map[int64]map[string][]scheduleConflictDetailRow) []scheduleConflictDetailRow {
	if len(plan.AssistantIDs) == 0 {
		return nil
	}
	dateKey := plan.LessonDate.Format("2006-01-02")
	result := make([]scheduleConflictDetailRow, 0)
	seen := make(map[int64]struct{})
	for _, assistantID := range plan.AssistantIDs {
		rowsByDate := assistantIndex[assistantID]
		if len(rowsByDate) == 0 {
			continue
		}
		rows := filterOverlappingScheduleConflictRows(rowsByDate[dateKey], plan.LessonDate, plan.StartAt, plan.EndAt)
		for _, row := range rows {
			if _, ok := seen[row.ID]; ok {
				continue
			}
			seen[row.ID] = struct{}{}
			result = append(result, row)
		}
	}
	return result
}

func buildScheduleValidationResultFromPlans(
	base model.OneToOneScheduleCreateContext,
	plans []normalizedSchedulePlan,
	teacherNames map[int64]string,
	classroomNames map[int64]string,
	teacherConflictRows []scheduleConflictDetailRow,
	classroomConflictRows []scheduleConflictDetailRow,
	studentConflictRows []scheduleConflictDetailRow,
	assistantConflictRows []scheduleConflictDetailRow,
) model.TeachingScheduleValidationResult {
	teacherIndex := groupScheduleConflictRowsByIDAndDate(teacherConflictRows, func(row scheduleConflictDetailRow) int64 {
		return row.TeacherID
	})
	classroomIndex := groupScheduleConflictRowsByIDAndDate(classroomConflictRows, func(row scheduleConflictDetailRow) int64 {
		return row.ClassroomID
	})
	studentIndex := groupScheduleConflictRowsByDate(studentConflictRows)

	assistantCandidateSet := make(map[int64]struct{})
	for _, id := range collectPlanAssistantIDs(plans) {
		if id > 0 {
			assistantCandidateSet[id] = struct{}{}
		}
	}
	assistantIndex := groupAssistantConflictRowsByCandidateIDAndDate(assistantConflictRows, assistantCandidateSet)

	globalTypeSet := make(map[string]struct{})
	globalExistingMap := make(map[int64]model.TeachingScheduleConflictItem)
	currentItems := make([]model.TeachingScheduleConflictItem, 0, len(plans))
	validationItems := make([]model.TeachingScheduleValidationItem, 0, len(plans))

	for _, plan := range plans {
		dateKey := plan.LessonDate.Format("2006-01-02")
		conflictTypes := make([]string, 0, 4)
		itemExistingMap := make(map[int64]model.TeachingScheduleConflictItem)

		teacherRows := filterOverlappingScheduleConflictRows(teacherIndex[plan.TeacherID][dateKey], plan.LessonDate, plan.StartAt, plan.EndAt)
		if len(teacherRows) > 0 {
			conflictTypes = append(conflictTypes, "老师")
			globalTypeSet["老师"] = struct{}{}
			appendConflictRowsToItemMap(itemExistingMap, teacherRows, "老师")
			appendConflictRowsToItemMap(globalExistingMap, teacherRows, "老师")
		}

		classroomRows := filterOverlappingScheduleConflictRows(classroomIndex[plan.ClassroomID][dateKey], plan.LessonDate, plan.StartAt, plan.EndAt)
		if len(classroomRows) > 0 {
			conflictTypes = append(conflictTypes, "教室")
			globalTypeSet["教室"] = struct{}{}
			appendConflictRowsToItemMap(itemExistingMap, classroomRows, "教室")
			appendConflictRowsToItemMap(globalExistingMap, classroomRows, "教室")
		}

		studentRows := filterOverlappingScheduleConflictRows(studentIndex[dateKey], plan.LessonDate, plan.StartAt, plan.EndAt)
		if len(studentRows) > 0 {
			conflictTypes = append(conflictTypes, "学员")
			globalTypeSet["学员"] = struct{}{}
			appendConflictRowsToItemMap(itemExistingMap, studentRows, "学员")
			appendConflictRowsToItemMap(globalExistingMap, studentRows, "学员")
		}

		assistantRows := collectPlanAssistantConflictRows(plan, assistantIndex)
		if len(assistantRows) > 0 {
			conflictTypes = append(conflictTypes, "助教")
			globalTypeSet["助教"] = struct{}{}
			appendConflictRowsToItemMap(itemExistingMap, assistantRows, "助教")
			appendConflictRowsToItemMap(globalExistingMap, assistantRows, "助教")
		}

		sort.Strings(conflictTypes)
		currentItems = append(currentItems, model.TeachingScheduleConflictItem{
			Name:           base.ClassName,
			ClassTypeText:  "1对1日程",
			Date:           dateKey,
			Week:           weekDisplay(plan.LessonDate),
			TimeText:       plan.StartAt.Format("15:04") + "~" + plan.EndAt.Format("15:04"),
			TeacherID:      emptyStringIfZero(plan.TeacherID),
			TeacherName:    firstNonEmptyString(teacherNames[plan.TeacherID], "-"),
			AssistantNames: compactStrings(plan.AssistantNames),
			ClassroomName:  firstNonEmptyString(classroomNames[plan.ClassroomID], "-"),
			StudentNames:   compactStrings([]string{base.StudentName}),
			ConflictTypes:  conflictTypes,
		})

		item := model.TeachingScheduleValidationItem{
			TeacherID:     emptyStringIfZero(plan.TeacherID),
			LessonDate:    dateKey,
			StartTime:     plan.StartAt.Format("15:04"),
			EndTime:       plan.EndAt.Format("15:04"),
			Valid:         len(conflictTypes) == 0,
			ConflictTypes: conflictTypes,
		}
		if !item.Valid {
			item.Message = buildAvailabilityConflictSummaryMessage(conflictTypes)
			item.ExistingSchedules = finalizeConflictItemMap(itemExistingMap)
		}
		validationItems = append(validationItems, item)
	}

	conflictTypes := make([]string, 0, len(globalTypeSet))
	for key := range globalTypeSet {
		conflictTypes = append(conflictTypes, key)
	}
	sort.Strings(conflictTypes)

	if len(conflictTypes) == 0 {
		return model.TeachingScheduleValidationResult{
			Valid: true,
			Items: validationItems,
		}
	}

	return model.TeachingScheduleValidationResult{
		Valid:             false,
		Message:           buildConflictSummaryMessage(conflictTypes),
		CurrentSchedules:  currentItems,
		ExistingSchedules: finalizeConflictItemMap(globalExistingMap),
		ConflictTypes:     conflictTypes,
		Items:             validationItems,
	}
}

func buildScheduleConflictResultFromExisting(
	current scheduleConflictDetailRow,
	teacherConflicts []scheduleConflictDetailRow,
	classroomConflicts []scheduleConflictDetailRow,
	studentConflicts []scheduleConflictDetailRow,
	assistantConflicts []scheduleConflictDetailRow,
) ([]model.TeachingScheduleConflictItem, []model.TeachingScheduleConflictItem, []string) {
	typeSet := make(map[string]struct{})
	currentConflictTypes := make([]string, 0, 4)
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
	if len(assistantConflicts) > 0 {
		currentConflictTypes = append(currentConflictTypes, "助教")
		typeSet["助教"] = struct{}{}
	}

	currentItems := []model.TeachingScheduleConflictItem{{
		Name:           current.TeachingClassName,
		ClassTypeText:  scheduleClassTypeText(current.ClassType),
		Date:           current.LessonDate.Format("2006-01-02"),
		Week:           weekDisplay(current.LessonDate),
		TimeText:       current.StartAt.Format("15:04") + "~" + current.EndAt.Format("15:04"),
		TeacherID:      emptyStringIfZero(current.TeacherID),
		TeacherName:    firstNonEmptyString(current.TeacherName, "-"),
		AssistantNames: compactStrings(current.AssistantNames),
		ClassroomName:  firstNonEmptyString(current.ClassroomName, "-"),
		StudentNames:   compactStrings([]string{current.StudentName}),
		ConflictTypes:  currentConflictTypes,
	}}

	existingMap := make(map[int64]model.TeachingScheduleConflictItem)
	appendExisting := func(row scheduleConflictDetailRow, conflictType string) {
		if row.ID == current.ID {
			return
		}
		item, ok := existingMap[row.ID]
		if !ok {
			item = model.TeachingScheduleConflictItem{
				Name:           row.TeachingClassName,
				ClassTypeText:  scheduleClassTypeText(row.ClassType),
				Date:           row.LessonDate.Format("2006-01-02"),
				Week:           weekDisplay(row.LessonDate),
				TimeText:       row.StartAt.Format("15:04") + "~" + row.EndAt.Format("15:04"),
				TeacherID:      emptyStringIfZero(row.TeacherID),
				TeacherName:    firstNonEmptyString(row.TeacherName, "-"),
				AssistantNames: compactStrings(row.AssistantNames),
				ClassroomName:  firstNonEmptyString(row.ClassroomName, "-"),
				StudentNames:   compactStrings([]string{row.StudentName}),
				ConflictTypes:  []string{},
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
	for _, row := range assistantConflicts {
		appendExisting(row, "助教")
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

func buildExistingConflictSummaryMessage(conflictTypes []string) string {
	if len(conflictTypes) == 0 {
		return "当前日程存在冲突"
	}
	if len(conflictTypes) == 1 {
		return "当前日程存在" + conflictTypes[0] + "冲突"
	}
	return "当前日程存在" + strings.Join(conflictTypes, "、") + "冲突"
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

func scheduleRowOverlapsAnySlot(lessonDate, startAt, endAt time.Time, slots []normalizedScheduleSlot) bool {
	for _, slot := range slots {
		if lessonDate.Format("2006-01-02") != slot.LessonDate.Format("2006-01-02") {
			continue
		}
		if startAt.Before(slot.EndAt) && endAt.After(slot.StartAt) {
			return true
		}
	}
	return false
}

func scheduleSlotsDateRange(slots []normalizedScheduleSlot) (string, string) {
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

func availabilitySlotsOverlap(slot normalizedAvailabilityScheduleSlot, lessonDate, startAt, endAt time.Time) bool {
	return lessonDate.Format("2006-01-02") == slot.LessonDate.Format("2006-01-02") &&
		startAt.Before(slot.EndAt) &&
		endAt.After(slot.StartAt)
}

func availabilityRowOverlapsAnySlot(row scheduleAvailabilityConflictRow, slots []normalizedScheduleSlot) bool {
	for _, slot := range slots {
		if row.LessonDate.Format("2006-01-02") != slot.LessonDate.Format("2006-01-02") {
			continue
		}
		if row.StartAt.Before(slot.EndAt) && row.EndAt.After(slot.StartAt) {
			return true
		}
	}
	return false
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
			Name:           row.TeachingClassName,
			ClassTypeText:  scheduleClassTypeText(row.ClassType),
			Date:           row.LessonDate.Format("2006-01-02"),
			Week:           weekDisplay(row.LessonDate),
			TimeText:       row.StartAt.Format("15:04") + "~" + row.EndAt.Format("15:04"),
			TeacherID:      emptyStringIfZero(row.TeacherID),
			TeacherName:    firstNonEmptyString(row.TeacherName, "-"),
			AssistantNames: compactStrings(row.AssistantNames),
			ClassroomName:  firstNonEmptyString(row.ClassroomName, "-"),
			StudentNames:   compactStrings([]string{row.StudentName}),
			ConflictTypes:  []string{},
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

func buildUnavailableAssistantAvailabilityResult(assistantIDs []int64, assistantNames map[int64]string, message string) model.AssistantScheduleAvailabilityResult {
	result := model.AssistantScheduleAvailabilityResult{
		InvalidCount: len(assistantIDs),
		Items:        make([]model.AssistantScheduleAvailabilityItem, 0, len(assistantIDs)),
	}
	for _, assistantID := range assistantIDs {
		result.Items = append(result.Items, model.AssistantScheduleAvailabilityItem{
			AssistantID:   strconv.FormatInt(assistantID, 10),
			AssistantName: firstNonEmptyString(assistantNames[assistantID], "-"),
			Valid:         false,
			Message:       message,
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

func decodeJSONStringArray(raw []byte) []string {
	if len(raw) == 0 {
		return nil
	}
	var result []string
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil
	}
	return compactStrings(result)
}

func stringSliceHasAnyID(values []string, candidateSet map[int64]struct{}) bool {
	for _, value := range values {
		id, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err != nil || id <= 0 {
			continue
		}
		if _, ok := candidateSet[id]; ok {
			return true
		}
	}
	return false
}

func containsString(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

func containsInt64(list []int64, value int64) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}
