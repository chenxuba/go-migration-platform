package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

func ensureStudentTeachingRecordTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS student_teaching_record (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			inst_id BIGINT NOT NULL,
			teaching_record_id BIGINT NOT NULL DEFAULT 0,
			teaching_schedule_id BIGINT NOT NULL DEFAULT 0,
			timetable_source_type INT NOT NULL DEFAULT 0,
			timetable_source_id BIGINT NOT NULL DEFAULT 0,
			student_id BIGINT NOT NULL DEFAULT 0,
			student_name VARCHAR(100) NOT NULL DEFAULT '',
			student_phone VARCHAR(32) NOT NULL DEFAULT '',
			avatar_url VARCHAR(500) NOT NULL DEFAULT '',
			source_type INT NOT NULL DEFAULT 0,
			current_student_status INT NOT NULL DEFAULT 0,
			status INT NOT NULL DEFAULT 0,
			class_id BIGINT NOT NULL DEFAULT 0,
			class_name VARCHAR(150) NOT NULL DEFAULT '',
			one_to_one_id BIGINT NOT NULL DEFAULT 0,
			one_to_one_name VARCHAR(150) NOT NULL DEFAULT '',
			lesson_id BIGINT NOT NULL DEFAULT 0,
			lesson_name VARCHAR(150) NOT NULL DEFAULT '',
			subject_id BIGINT NOT NULL DEFAULT 0,
			subject_name VARCHAR(100) NOT NULL DEFAULT '',
			teaching_content LONGTEXT NULL,
			teaching_content_images_json JSON NULL,
			classroom_id BIGINT NOT NULL DEFAULT 0,
			classroom_name VARCHAR(150) NOT NULL DEFAULT '',
			main_teacher_id BIGINT NOT NULL DEFAULT 0,
			main_teacher_name VARCHAR(100) NOT NULL DEFAULT '',
			teacher_employee_type INT NOT NULL DEFAULT 0,
			assistant_teacher_ids_json JSON NULL,
			assistant_teacher_names_json JSON NULL,
			class_teacher_ids_json JSON NULL,
			class_teacher_names_json JSON NULL,
			roll_call_class_teacher_ids_json JSON NULL,
			roll_call_class_teacher_names_json JSON NULL,
			current_class_teacher_ids_json JSON NULL,
			current_class_teacher_names_json JSON NULL,
			one2one_teacher_ids_json JSON NULL,
			one2one_teacher_names_json JSON NULL,
			tuition_account_id BIGINT NOT NULL DEFAULT 0,
			tuition_account_name VARCHAR(150) NOT NULL DEFAULT '',
			sku_mode INT NOT NULL DEFAULT 0,
			quantity DECIMAL(18,2) NOT NULL DEFAULT 0,
			actual_quantity DECIMAL(18,2) NOT NULL DEFAULT 0,
			amount DECIMAL(18,2) NOT NULL DEFAULT 0,
			actual_deduct DECIMAL(18,2) NOT NULL DEFAULT 0,
			actual_tuition DECIMAL(18,2) NOT NULL DEFAULT 0,
			arrear_quantity DECIMAL(18,2) NOT NULL DEFAULT 0,
			teacher_class_time DECIMAL(18,2) NOT NULL DEFAULT 0,
			remark VARCHAR(1000) NOT NULL DEFAULT '',
			external_remark VARCHAR(1000) NOT NULL DEFAULT '',
			has_compensated TINYINT(1) NOT NULL DEFAULT 0,
			advisor_staff_id BIGINT NOT NULL DEFAULT 0,
			advisor_staff_name VARCHAR(100) NOT NULL DEFAULT '',
			student_manager_id BIGINT NOT NULL DEFAULT 0,
			student_manager_name VARCHAR(100) NOT NULL DEFAULT '',
			start_time DATETIME NOT NULL,
			end_time DATETIME NOT NULL,
			teaching_record_created_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			record_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_staff_id BIGINT NOT NULL DEFAULT 0,
			updated_staff_name VARCHAR(100) NOT NULL DEFAULT '',
			updated_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT NOT NULL DEFAULT 0,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			KEY idx_student_teaching_record_list (inst_id, start_time, updated_time, id),
			KEY idx_student_teaching_record_student (inst_id, student_id, start_time),
			KEY idx_student_teaching_record_teaching (inst_id, teaching_record_id),
			KEY idx_student_teaching_record_schedule (inst_id, teaching_schedule_id)
		)
	`)
	return err
}

type classRecordStudentQueryFragments struct {
	whereSQL string
	args     []any
	orderBy  string
}

func (repo *Repository) buildStudentTeachingRecordQuery(dto model.StudentTeachingRecordPagedQueryDTO, instID int64) classRecordStudentQueryFragments {
	whereParts := []string{
		"inst_id = ?",
		"del_flag = 0",
	}
	args := []any{instID}
	query := dto.QueryModel

	if begin := parseDateStart(strings.TrimSpace(query.BeginStartTime)); begin != nil {
		whereParts = append(whereParts, "start_time >= ?")
		args = append(args, *begin)
	}
	if end := parseDateEnd(strings.TrimSpace(query.EndStartTime)); end != nil {
		whereParts = append(whereParts, "start_time <= ?")
		args = append(args, *end)
	}
	if len(query.TeacherIDs) > 0 {
		values := normalizeStringIDs(query.TeacherIDs)
		if len(values) > 0 {
			whereParts = append(whereParts, "CAST(main_teacher_id AS CHAR) IN ("+sqlPlaceholders(len(values))+")")
			args = append(args, stringSliceToAny(values)...)
		}
	}
	if len(query.AssistantTeacherIDs) > 0 {
		if filter := buildJSONArrayAnyMatch("assistant_teacher_ids_json", normalizeStringIDs(query.AssistantTeacherIDs)); filter != "" {
			whereParts = append(whereParts, filter)
		}
	}
	if len(query.ClassTeacherIDs) > 0 {
		values := normalizeStringIDs(query.ClassTeacherIDs)
		if len(values) > 0 {
			parts := make([]string, 0, 3)
			if filter := buildJSONArrayAnyMatch("class_teacher_ids_json", values); filter != "" {
				parts = append(parts, filter)
			}
			if filter := buildJSONArrayAnyMatch("roll_call_class_teacher_ids_json", values); filter != "" {
				parts = append(parts, filter)
			}
			if filter := buildJSONArrayAnyMatch("current_class_teacher_ids_json", values); filter != "" {
				parts = append(parts, filter)
			}
			if len(parts) > 0 {
				whereParts = append(whereParts, "("+strings.Join(parts, " OR ")+")")
			}
		}
	}
	if len(query.One2OneTeacherIDs) > 0 {
		if filter := buildJSONArrayAnyMatch("one2one_teacher_ids_json", normalizeStringIDs(query.One2OneTeacherIDs)); filter != "" {
			whereParts = append(whereParts, filter)
		}
	}
	if len(query.TimetableSourceTypes) > 0 {
		whereParts = append(whereParts, "timetable_source_type IN ("+sqlPlaceholders(len(query.TimetableSourceTypes))+")")
		args = append(args, intSliceToAny(query.TimetableSourceTypes)...)
	}
	if len(query.StudentSourceTypes) > 0 {
		whereParts = append(whereParts, "source_type IN ("+sqlPlaceholders(len(query.StudentSourceTypes))+")")
		args = append(args, intSliceToAny(query.StudentSourceTypes)...)
	}
	if len(query.LessonChargingModeEnums) > 0 {
		whereParts = append(whereParts, "sku_mode IN ("+sqlPlaceholders(len(query.LessonChargingModeEnums))+")")
		args = append(args, intSliceToAny(query.LessonChargingModeEnums)...)
	}
	if len(query.StudentTeachingRecordStatuses) > 0 {
		whereParts = append(whereParts, "status IN ("+sqlPlaceholders(len(query.StudentTeachingRecordStatuses))+")")
		args = append(args, intSliceToAny(query.StudentTeachingRecordStatuses)...)
	}
	if query.IsArrear != nil {
		if *query.IsArrear {
			whereParts = append(whereParts, "arrear_quantity > 0")
		} else {
			whereParts = append(whereParts, "arrear_quantity <= 0")
		}
	}

	orderBy := "start_time DESC, updated_time DESC, id DESC"
	if dto.SortModel.StartTime == 1 {
		orderBy = "start_time ASC, updated_time DESC, id ASC"
	} else if dto.SortModel.UpdatedTime == 1 {
		orderBy = "updated_time ASC, start_time DESC, id ASC"
	} else if dto.SortModel.UpdatedTime == 2 {
		orderBy = "updated_time DESC, start_time DESC, id DESC"
	}

	return classRecordStudentQueryFragments{
		whereSQL: strings.Join(whereParts, " AND "),
		args:     args,
		orderBy:  orderBy,
	}
}

func (repo *Repository) GetStudentTeachingRecordPagedList(ctx context.Context, instID int64, dto model.StudentTeachingRecordPagedQueryDTO) (model.StudentTeachingRecordPagedResult, error) {
	_, pageSize, offset := normalizeRollCallPage(dto.PageRequestModel)
	fragments := repo.buildStudentTeachingRecordQuery(dto, instID)

	var result model.StudentTeachingRecordPagedResult
	if err := repo.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*),
			IFNULL(SUM(quantity), 0),
			IFNULL(SUM(actual_tuition), 0),
			COUNT(DISTINCT student_id)
		FROM student_teaching_record
		WHERE `+fragments.whereSQL, fragments.args...).Scan(&result.Total, &result.TotalClassTimes, &result.TotalTuition, &result.TotalStudentCount); err != nil {
		return model.StudentTeachingRecordPagedResult{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			CAST(id AS CHAR),
			CAST(teaching_record_id AS CHAR),
			CAST(student_id AS CHAR),
			student_name,
			student_phone,
			avatar_url,
			main_teacher_name,
			teacher_employee_type,
			CAST(IFNULL(assistant_teacher_names_json, JSON_ARRAY()) AS CHAR),
			class_name,
			one_to_one_name,
			lesson_name,
			status,
			source_type,
			DATE_FORMAT(start_time, '%Y-%m-%dT%H:%i:%s'),
			DATE_FORMAT(end_time, '%Y-%m-%dT%H:%i:%s'),
			DATE_FORMAT(teaching_record_created_time, '%Y-%m-%dT%H:%i:%s'),
			timetable_source_type,
			DATE_FORMAT(updated_time, '%Y-%m-%dT%H:%i:%s'),
			updated_staff_name,
			DATE_FORMAT(record_time, '%Y-%m-%dT%H:%i:%s'),
			quantity,
			actual_quantity,
			amount,
			sku_mode,
			actual_deduct,
			actual_tuition,
			arrear_quantity,
			remark,
			external_remark,
			CAST(tuition_account_id AS CHAR),
			tuition_account_name,
			has_compensated,
			CAST(subject_id AS CHAR),
			subject_name,
			CAST(advisor_staff_id AS CHAR),
			advisor_staff_name,
			CAST(student_manager_id AS CHAR),
			student_manager_name,
			IFNULL(teaching_content, ''),
			CAST(IFNULL(teaching_content_images_json, JSON_ARRAY()) AS CHAR),
			classroom_name,
			CAST(IFNULL(one2one_teacher_names_json, JSON_ARRAY()) AS CHAR),
			CAST(IFNULL(class_teacher_names_json, JSON_ARRAY()) AS CHAR),
			CAST(IFNULL(roll_call_class_teacher_names_json, JSON_ARRAY()) AS CHAR),
			CAST(IFNULL(current_class_teacher_names_json, JSON_ARRAY()) AS CHAR)
		FROM student_teaching_record
		WHERE `+fragments.whereSQL+`
		ORDER BY `+fragments.orderBy+`
		LIMIT ? OFFSET ?
	`, append(fragments.args, pageSize, offset)...)
	if err != nil {
		return model.StudentTeachingRecordPagedResult{}, err
	}
	defer rows.Close()

	result.List = make([]model.StudentTeachingRecordItem, 0, pageSize)
	for rows.Next() {
		var (
			item                     model.StudentTeachingRecordItem
			imageBytes               []byte
			rawAssistants            string
			rawOne2OneTeachers       string
			rawClassTeachers         string
			rawRollCallClassTeachers string
			rawCurrentClassTeachers  string
		)
		if err := rows.Scan(
			&item.StudentTeachingRecordID,
			&item.TeachingRecordID,
			&item.StudentID,
			&item.StudentName,
			&item.StudentPhone,
			&item.Avatar,
			&item.TeacherName,
			&item.TeacherEmployeeType,
			&rawAssistants,
			&item.ClassName,
			&item.One2OneName,
			&item.LessonName,
			&item.Status,
			&item.SourceType,
			&item.StartTime,
			&item.EndTime,
			&item.TeachingRecordCreatedTime,
			&item.TimetableSourceType,
			&item.UpdatedTime,
			&item.UpdatedStaffName,
			&item.RecordTime,
			&item.Quantity,
			&item.ActualQuantity,
			&item.Amount,
			&item.SkuMode,
			&item.ActualDeduct,
			&item.ActualTuition,
			&item.ArrearQuantity,
			&item.Remark,
			&item.ExternalRemark,
			&item.TuitionAccountID,
			&item.TuitionAccountName,
			&item.HasCompensated,
			&item.SubjectID,
			&item.SubjectName,
			&item.AdvisorStaffID,
			&item.AdvisorStaffName,
			&item.StudentManagerID,
			&item.StudentManagerName,
			&item.TeachingContent,
			&imageBytes,
			&item.ClassRoomName,
			&rawOne2OneTeachers,
			&rawClassTeachers,
			&rawRollCallClassTeachers,
			&rawCurrentClassTeachers,
		); err != nil {
			return model.StudentTeachingRecordPagedResult{}, err
		}
		item.Assistants = normalizeJSONStringListText(rawAssistants)
		item.One2OneTeachers = normalizeJSONStringListText(rawOne2OneTeachers)
		item.ClassTeachers = normalizeJSONStringListText(rawClassTeachers)
		item.RollCallClassTeachers = normalizeJSONStringListText(rawRollCallClassTeachers)
		item.CurrentClassTeachers = normalizeJSONStringListText(rawCurrentClassTeachers)
		if len(imageBytes) > 0 {
			_ = json.Unmarshal(imageBytes, &item.TeachingContentImages)
		}
		if item.TeachingContentImages == nil {
			item.TeachingContentImages = []string{}
		}
		result.List = append(result.List, item)
	}
	return result, rows.Err()
}

func (repo *Repository) GetScheduleTeachingRecordPagedList(ctx context.Context, instID int64, dto model.ScheduleTeachingRecordPagedQueryDTO) (model.ScheduleTeachingRecordPagedResult, error) {
	_, pageSize, offset := normalizeRollCallPage(dto.PageRequestModel)
	studentFragments := repo.buildStudentTeachingRecordQuery(model.StudentTeachingRecordPagedQueryDTO{
		PageRequestModel: dto.PageRequestModel,
		SortModel: model.StudentTeachingRecordSortModel{
			StartTime:   dto.SortModel.StartTime,
			UpdatedTime: dto.SortModel.UpdatedTime,
		},
		QueryModel: dto.QueryModel,
	}, instID)

	var result model.ScheduleTeachingRecordPagedResult
	if err := repo.db.QueryRowContext(ctx, `
		SELECT
			COUNT(DISTINCT teaching_record_id),
			IFNULL(SUM(record_stat.teacher_class_time), 0),
			IFNULL(SUM(record_stat.actual_tuition), 0)
		FROM (
			SELECT
				teaching_record_id,
				MAX(teacher_class_time) AS teacher_class_time,
				SUM(actual_tuition) AS actual_tuition
			FROM student_teaching_record
			WHERE `+studentFragments.whereSQL+`
			GROUP BY teaching_record_id
		) AS record_stat
	`, studentFragments.args...).Scan(&result.Total, &result.TotalTeacherTimes, &result.TotalTuition); err != nil {
		return model.ScheduleTeachingRecordPagedResult{}, err
	}
	if err := repo.db.QueryRowContext(ctx, `
		SELECT IFNULL(SUM(actual_quantity), 0)
		FROM student_teaching_record
		WHERE `+studentFragments.whereSQL, studentFragments.args...).Scan(&result.TotalClassTimes); err != nil {
		return model.ScheduleTeachingRecordPagedResult{}, err
	}

	orderBy := "MAX(start_time) DESC, MAX(updated_time) DESC, teaching_record_id DESC"
	if dto.SortModel.StartTime == 1 {
		orderBy = "MAX(start_time) ASC, MAX(updated_time) DESC, teaching_record_id ASC"
	} else if dto.SortModel.UpdatedTime == 1 {
		orderBy = "MAX(updated_time) ASC, MAX(start_time) DESC, teaching_record_id ASC"
	} else if dto.SortModel.UpdatedTime == 2 {
		orderBy = "MAX(updated_time) DESC, MAX(start_time) DESC, teaching_record_id DESC"
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			CAST(teaching_record_id AS CHAR),
			DATE_FORMAT(MAX(start_time), '%Y-%m-%dT%H:%i:%s'),
			DATE_FORMAT(MAX(end_time), '%Y-%m-%dT%H:%i:%s'),
			MAX(timetable_source_type),
			MAX(class_name),
			MAX(one_to_one_name),
			MAX(lesson_name),
			CAST(MAX(subject_id) AS CHAR),
			MAX(subject_name),
			CASE WHEN SUM(CASE WHEN status = 4 THEN 1 ELSE 0 END) > 0 THEN 1 ELSE 2 END AS roll_call_status,
			CASE WHEN COUNT(*) = 0 THEN 0 ELSE SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END) / COUNT(*) END AS attendance_rate,
			SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END) AS attend_count,
			COUNT(*) AS should_attend_count,
			IFNULL(SUM(actual_quantity), 0),
			IFNULL(SUM(actual_tuition), 0),
			MAX(main_teacher_name),
			MAX(CAST(assistant_teacher_names_json AS CHAR(1000))),
			MAX(teacher_class_time),
			DATE_FORMAT(MAX(teaching_record_created_time), '%Y-%m-%d %H:%i:%s'),
			DATE_FORMAT(MAX(updated_time), '%Y-%m-%d %H:%i:%s')
		FROM student_teaching_record
		WHERE `+studentFragments.whereSQL+`
		GROUP BY teaching_record_id
		ORDER BY `+orderBy+`
		LIMIT ? OFFSET ?
	`, append(studentFragments.args, pageSize, offset)...)
	if err != nil {
		return model.ScheduleTeachingRecordPagedResult{}, err
	}
	defer rows.Close()

	result.List = make([]model.ScheduleTeachingRecordItem, 0, pageSize)
	for rows.Next() {
		var item model.ScheduleTeachingRecordItem
		var rawAssistants string
		if err := rows.Scan(
			&item.TeachingRecordID,
			&item.StartTime,
			&item.EndTime,
			&item.TimetableSourceType,
			&item.ClassName,
			&item.One2OneName,
			&item.LessonName,
			&item.SubjectID,
			&item.SubjectName,
			&item.RollCallStatus,
			&item.AttendanceRate,
			&item.AttendCount,
			&item.ShouldAttendCount,
			&item.ActualQuantity,
			&item.ActualTuition,
			&item.TeacherName,
			&rawAssistants,
			&item.TeacherClassTime,
			&item.CreatedTime,
			&item.UpdatedTime,
		); err != nil {
			return model.ScheduleTeachingRecordPagedResult{}, err
		}
		item.Assistants = normalizeJSONStringListText(rawAssistants)
		result.List = append(result.List, item)
	}
	return result, rows.Err()
}

func normalizeRollCallPage(page model.RollCallPageRequestModel) (pageIndex, pageSize, offset int) {
	pageIndex = page.PageIndex
	if pageIndex <= 0 {
		pageIndex = 1
	}
	pageSize = page.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 200 {
		pageSize = 200
	}
	offset = (pageIndex - 1) * pageSize
	if page.SkipCount > 0 {
		offset = page.SkipCount
	}
	return
}

func normalizeStringIDs(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func buildJSONArrayAnyMatch(column string, values []string) string {
	if len(values) == 0 {
		return ""
	}
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, fmt.Sprintf("JSON_CONTAINS(IFNULL(%s, JSON_ARRAY()), JSON_QUOTE('%s'))", column, escapeSQLString(value)))
	}
	return "(" + strings.Join(parts, " OR ") + ")"
}

func escapeSQLString(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}

func stringSliceToAny(values []string) []any {
	result := make([]any, 0, len(values))
	for _, value := range values {
		result = append(result, value)
	}
	return result
}

func normalizeJSONStringListText(value string) string {
	text := strings.TrimSpace(value)
	if text == "" || text == "null" {
		return ""
	}
	var list []string
	if err := json.Unmarshal([]byte(text), &list); err != nil {
		return text
	}
	result := make([]string, 0, len(list))
	for _, item := range list {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}
	return strings.Join(result, "、")
}
