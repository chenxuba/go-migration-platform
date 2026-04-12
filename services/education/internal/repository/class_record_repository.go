package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

type classRecordScheduleMeta struct {
	ScheduleID int64
	ClassType  int
	ClassID    int64
	StudentID  int64
	StartAt    time.Time
}

type teachingRecordDeleteStudentRow struct {
	StudentTeachingRecordID int64
	TeachingScheduleID      int64
	StudentID               int64
	TuitionAccountID        int64
	ActualDeduct            float64
	ActualTuition           float64
}

func classRecordRollCallStatus(status int) int {
	switch normalizeTeachingScheduleCallStatus(status) {
	case 2:
		return 2
	default:
		return 1
	}
}

const teachingRecordDetailStudentStatusPendingRollCall = 0

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
			is_late TINYINT(1) NOT NULL DEFAULT 0,
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
			is_auto_roll_call TINYINT(1) NOT NULL DEFAULT 0,
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
	if err != nil {
		return err
	}
	for _, statement := range []string{
		"ALTER TABLE student_teaching_record ADD COLUMN is_late TINYINT(1) NOT NULL DEFAULT 0 AFTER status",
		"ALTER TABLE student_teaching_record ADD COLUMN is_auto_roll_call TINYINT(1) NOT NULL DEFAULT 0 AFTER external_remark",
	} {
		if _, alterErr := db.ExecContext(ctx, statement); alterErr != nil && !strings.Contains(strings.ToLower(alterErr.Error()), "duplicate column") {
			return alterErr
		}
	}
	return nil
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
	if begin := parseDateStart(strings.TrimSpace(query.BeginCreateTime)); begin != nil {
		whereParts = append(whereParts, "teaching_record_created_time >= ?")
		args = append(args, *begin)
	}
	if end := parseDateEnd(strings.TrimSpace(query.EndCreateTime)); end != nil {
		whereParts = append(whereParts, "teaching_record_created_time <= ?")
		args = append(args, *end)
	}
	if begin := parseDateStart(strings.TrimSpace(query.BeginUpdatedTime)); begin != nil {
		whereParts = append(whereParts, "updated_time >= ?")
		args = append(args, *begin)
	}
	if end := parseDateEnd(strings.TrimSpace(query.EndUpdatedTime)); end != nil {
		whereParts = append(whereParts, "updated_time <= ?")
		args = append(args, *end)
	}
	if studentID := strings.TrimSpace(query.StudentID); studentID != "" {
		whereParts = append(whereParts, "CAST(student_id AS CHAR) = ?")
		args = append(args, studentID)
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
	if len(query.One2OneIDs) > 0 {
		values := normalizeStringIDs(query.One2OneIDs)
		if len(values) > 0 {
			whereParts = append(whereParts, "CAST(one_to_one_id AS CHAR) IN ("+sqlPlaceholders(len(values))+")")
			args = append(args, stringSliceToAny(values)...)
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
		modeValues := make([]int, 0, len(query.LessonChargingModeEnums))
		includeNoCountMode := false
		for _, item := range query.LessonChargingModeEnums {
			if item == 4 {
				includeNoCountMode = true
				continue
			}
			modeValues = append(modeValues, item)
		}
		switch {
		case len(modeValues) > 0 && includeNoCountMode:
			whereParts = append(whereParts, "(sku_mode IN ("+sqlPlaceholders(len(modeValues))+") OR source_type = 4)")
			args = append(args, intSliceToAny(modeValues)...)
		case len(modeValues) > 0:
			whereParts = append(whereParts, "sku_mode IN ("+sqlPlaceholders(len(modeValues))+")")
			args = append(args, intSliceToAny(modeValues)...)
		case includeNoCountMode:
			whereParts = append(whereParts, "source_type = 4")
		}
	}
	if len(query.StudentTeachingRecordStatuses) > 0 {
		statusValues := make([]int, 0, len(query.StudentTeachingRecordStatuses))
		for _, item := range query.StudentTeachingRecordStatuses {
			if item == 0 {
				statusValues = append(statusValues, 4)
				continue
			}
			statusValues = append(statusValues, item)
		}
		whereParts = append(whereParts, "status IN ("+sqlPlaceholders(len(statusValues))+")")
		args = append(args, intSliceToAny(statusValues)...)
	}
	if query.IsArrear != nil {
		if *query.IsArrear {
			whereParts = append(whereParts, "arrear_quantity > 0")
		} else {
			whereParts = append(whereParts, "arrear_quantity <= 0")
		}
	}
	if len(query.LessonIDs) > 0 {
		values := normalizeStringIDs(query.LessonIDs)
		if len(values) > 0 {
			whereParts = append(whereParts, "CAST(lesson_id AS CHAR) IN ("+sqlPlaceholders(len(values))+")")
			args = append(args, stringSliceToAny(values)...)
		}
	}
	if len(query.ClassIDs) > 0 {
		values := normalizeStringIDs(query.ClassIDs)
		if len(values) > 0 {
			whereParts = append(whereParts, "CAST(class_id AS CHAR) IN ("+sqlPlaceholders(len(values))+")")
			args = append(args, stringSliceToAny(values)...)
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

func buildScheduleTeachingRecordHaving(query model.StudentTeachingRecordQueryModel) (string, []any) {
	conditions := make([]string, 0, 1)
	args := make([]any, 0, 1)
	if query.ScheduleCallStatus != nil {
		status := *query.ScheduleCallStatus
		if status == 1 || status == 2 {
			conditions = append(conditions, "CASE WHEN SUM(CASE WHEN status = 4 THEN 1 ELSE 0 END) > 0 THEN 1 ELSE 2 END = ?")
			args = append(args, status)
		}
	}
	if len(conditions) == 0 {
		return "", nil
	}
	return " HAVING " + strings.Join(conditions, " AND "), args
}

func filterScheduleTeachingRecordItemsByStatus(items []model.ScheduleTeachingRecordItem, status int) []model.ScheduleTeachingRecordItem {
	if status != 1 && status != 2 {
		return items
	}
	result := make([]model.ScheduleTeachingRecordItem, 0, len(items))
	for _, item := range items {
		if item.RollCallStatus == status {
			result = append(result, item)
		}
	}
	return result
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
			CASE
				WHEN IFNULL(arrear_quantity, 0) > 0 AND IFNULL(actual_tuition, 0) <= 0 THEN 0
				ELSE IFNULL(actual_quantity, 0)
			END AS actual_quantity,
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
	statusFilter := 0
	if dto.QueryModel.ScheduleCallStatus != nil {
		statusFilter = *dto.QueryModel.ScheduleCallStatus
	}
	filterByComputedStatus := statusFilter == 1 || statusFilter == 2
	havingSQL := ""
	var havingArgs []any

	var result model.ScheduleTeachingRecordPagedResult
	if !filterByComputedStatus {
		havingSQL, havingArgs = buildScheduleTeachingRecordHaving(dto.QueryModel)
		if err := repo.db.QueryRowContext(ctx, `
			SELECT
				COUNT(*),
				IFNULL(SUM(record_stat.actual_quantity), 0),
				IFNULL(SUM(record_stat.teacher_class_time), 0),
				IFNULL(SUM(record_stat.actual_tuition), 0)
			FROM (
				SELECT
					teaching_record_id,
					MAX(teacher_class_time) AS teacher_class_time,
					SUM(actual_tuition) AS actual_tuition,
					SUM(CASE
						WHEN IFNULL(arrear_quantity, 0) > 0 AND IFNULL(actual_tuition, 0) <= 0 THEN 0
						ELSE IFNULL(actual_quantity, 0)
					END) AS actual_quantity
				FROM student_teaching_record
				WHERE `+studentFragments.whereSQL+`
				GROUP BY teaching_record_id
				`+havingSQL+`
			) AS record_stat
		`, append(studentFragments.args, havingArgs...)...).Scan(&result.Total, &result.TotalClassTimes, &result.TotalTeacherTimes, &result.TotalTuition); err != nil {
			return model.ScheduleTeachingRecordPagedResult{}, err
		}
	}

	orderBy := "MAX(start_time) DESC, MAX(updated_time) DESC, teaching_record_id DESC"
	if dto.SortModel.StartTime == 1 {
		orderBy = "MAX(start_time) ASC, MAX(updated_time) DESC, teaching_record_id ASC"
	} else if dto.SortModel.UpdatedTime == 1 {
		orderBy = "MAX(updated_time) ASC, MAX(start_time) DESC, teaching_record_id ASC"
	} else if dto.SortModel.UpdatedTime == 2 {
		orderBy = "MAX(updated_time) DESC, MAX(start_time) DESC, teaching_record_id DESC"
	}

	querySQL := `
		SELECT
			CAST(teaching_record_id AS CHAR),
			CAST(MAX(teaching_schedule_id) AS CHAR),
			DATE_FORMAT(MAX(start_time), '%Y-%m-%dT%H:%i:%s'),
			DATE_FORMAT(MAX(end_time), '%Y-%m-%dT%H:%i:%s'),
			MAX(timetable_source_type),
			MAX(class_name),
			MAX(one_to_one_name),
			MAX(lesson_name),
			CAST(MAX(subject_id) AS CHAR),
			MAX(subject_name),
			CASE WHEN SUM(CASE WHEN status = 4 THEN 1 ELSE 0 END) > 0 THEN 1 ELSE 2 END AS roll_call_status,
			CASE
				WHEN SUM(CASE WHEN source_type <> 4 THEN 1 ELSE 0 END) = 0 THEN 0
				ELSE SUM(CASE WHEN source_type <> 4 AND status = 1 THEN 1 ELSE 0 END) / SUM(CASE WHEN source_type <> 4 THEN 1 ELSE 0 END)
			END AS attendance_rate,
			SUM(CASE WHEN source_type <> 4 AND status = 1 THEN 1 ELSE 0 END) AS attend_count,
			SUM(CASE WHEN source_type <> 4 THEN 1 ELSE 0 END) AS should_attend_count,
			IFNULL(SUM(
				CASE
					WHEN IFNULL(arrear_quantity, 0) > 0 AND IFNULL(actual_tuition, 0) <= 0 THEN 0
					ELSE IFNULL(actual_quantity, 0)
				END
			), 0),
			IFNULL(SUM(actual_tuition), 0),
			MAX(main_teacher_name),
			MAX(CAST(assistant_teacher_names_json AS CHAR(1000))),
			MAX(teacher_class_time),
			DATE_FORMAT(MAX(teaching_record_created_time), '%Y-%m-%d %H:%i:%s'),
			DATE_FORMAT(MAX(updated_time), '%Y-%m-%d %H:%i:%s')
		FROM student_teaching_record
		WHERE ` + studentFragments.whereSQL + `
		GROUP BY teaching_record_id
		` + havingSQL + `
		ORDER BY ` + orderBy
	queryArgs := append(studentFragments.args, havingArgs...)
	if !filterByComputedStatus {
		querySQL += `
		LIMIT ? OFFSET ?`
		queryArgs = append(queryArgs, pageSize, offset)
	}
	rows, err := repo.db.QueryContext(ctx, querySQL, queryArgs...)
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
			&item.TimetableSourceID,
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
	if err := rows.Err(); err != nil {
		return model.ScheduleTeachingRecordPagedResult{}, err
	}
	if err := repo.fillScheduleTeachingRecordStats(ctx, instID, result.List); err != nil {
		return model.ScheduleTeachingRecordPagedResult{}, err
	}
	if filterByComputedStatus {
		filteredItems := filterScheduleTeachingRecordItemsByStatus(result.List, statusFilter)
		result.Total = len(filteredItems)
		result.TotalClassTimes = 0
		result.TotalTeacherTimes = 0
		result.TotalTuition = 0
		for _, item := range filteredItems {
			result.TotalClassTimes += item.ActualQuantity
			result.TotalTeacherTimes += item.TeacherClassTime
			result.TotalTuition += item.ActualTuition
		}
		if offset >= len(filteredItems) {
			result.List = []model.ScheduleTeachingRecordItem{}
			return result, nil
		}
		end := offset + pageSize
		if end > len(filteredItems) {
			end = len(filteredItems)
		}
		result.List = filteredItems[offset:end]
	}
	return result, nil
}

func (repo *Repository) GetTeachingRecordDetail(ctx context.Context, instID int64, query model.TeachingRecordDetailQueryDTO) (model.TeachingRecordDetailResult, error) {
	teachingRecordID, err := strconv.ParseInt(strings.TrimSpace(query.TeachingRecordID), 10, 64)
	if err != nil || teachingRecordID <= 0 {
		return model.TeachingRecordDetailResult{}, errors.New("上课记录ID无效")
	}

	var result model.TeachingRecordDetailResult
	var mainTeacherID int64
	var mainTeacherName string
	var rawAssistantTeacherIDs string
	var rawAssistantTeacherNames string
	var rawTeachingContentImages string
	err = repo.db.QueryRowContext(ctx, `
		SELECT
			CAST(MAX(teaching_record_id) AS CHAR),
			MAX(
				CASE
					WHEN LENGTH(TRIM(one_to_one_name)) > 0 THEN one_to_one_name
					WHEN LENGTH(TRIM(class_name)) > 0 THEN class_name
					ELSE lesson_name
				END
			) AS source_name,
			CASE
				WHEN MAX(one_to_one_id) > 0 THEN 2
				WHEN MAX(class_id) > 0 THEN 1
				WHEN MAX(timetable_source_type) = 3 THEN 3
				ELSE 0
			END AS source_type,
			CAST(
				CASE
					WHEN MAX(one_to_one_id) > 0 THEN MAX(one_to_one_id)
					WHEN MAX(class_id) > 0 THEN MAX(class_id)
					ELSE 0
				END AS CHAR
			) AS source_id,
			CAST(MAX(lesson_id) AS CHAR),
			CASE WHEN MAX(one_to_one_id) > 0 THEN 2 ELSE 1 END AS lesson_type,
			DATE_FORMAT(MAX(start_time), '%Y-%m-%dT%H:%i:%s'),
			DATE_FORMAT(MAX(end_time), '%Y-%m-%dT%H:%i:%s'),
			SUM(CASE WHEN source_type <> 4 AND status IN (1, 2, 3) THEN 1 ELSE 0 END) AS should_attendance_count,
			SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END) AS actual_attendance_count,
			SUM(CASE WHEN source_type <> 4 AND status = 3 THEN 1 ELSE 0 END) AS leave_count,
			SUM(CASE WHEN source_type <> 4 AND status = 2 THEN 1 ELSE 0 END) AS truancy_count,
			IFNULL(MAX(teacher_class_time), 0),
			IFNULL(SUM(quantity), 0),
			IFNULL(SUM(actual_tuition), 0),
			DATE_FORMAT(MIN(teaching_record_created_time), '%Y-%m-%d %H:%i:%s'),
			MAX(updated_staff_name),
			MAX(timetable_source_type),
			MAX(classroom_name),
			CAST(MAX(classroom_id) AS CHAR),
			CAST(MAX(teaching_schedule_id) AS CHAR),
			MAX(lesson_name),
			MAX(teaching_content),
			CAST(MAX(subject_id) AS CHAR),
			MAX(subject_name),
			MAX(main_teacher_id),
			MAX(main_teacher_name),
			MAX(CAST(IFNULL(assistant_teacher_ids_json, JSON_ARRAY()) AS CHAR)),
			MAX(CAST(IFNULL(assistant_teacher_names_json, JSON_ARRAY()) AS CHAR)),
			MAX(CAST(IFNULL(teaching_content_images_json, JSON_ARRAY()) AS CHAR))
		FROM student_teaching_record
		WHERE inst_id = ?
		  AND del_flag = 0
		  AND teaching_record_id = ?
	`, instID, teachingRecordID).Scan(
		&result.TeachingRecordID,
		&result.SourceName,
		&result.SourceType,
		&result.SourceID,
		&result.LessonID,
		&result.LessonType,
		&result.StartTime,
		&result.EndTime,
		&result.ShouldAttendanceCount,
		&result.ActualAttendanceCount,
		&result.LeaveCount,
		&result.TruancyCount,
		&result.TeacherClassTime,
		&result.StudentTotalClassTime,
		&result.StudentActualTuition,
		&result.CreatedTime,
		&result.CreatedStaffName,
		&result.TimetableSourceType,
		&result.ClassRoomName,
		&result.ClassRoomID,
		&result.TimetableSourceID,
		&result.LessonName,
		&result.TeachingContent,
		&result.SubjectID,
		&result.SubjectName,
		&mainTeacherID,
		&mainTeacherName,
		&rawAssistantTeacherIDs,
		&rawAssistantTeacherNames,
		&rawTeachingContentImages,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.TeachingRecordDetailResult{}, errors.New("未找到上课记录")
		}
		return model.TeachingRecordDetailResult{}, err
	}

	if strings.TrimSpace(rawTeachingContentImages) != "" {
		_ = json.Unmarshal([]byte(rawTeachingContentImages), &result.TeachingContentImages)
	}
	if result.TeachingContentImages == nil {
		result.TeachingContentImages = []string{}
	}

	result.TeacherList = buildTeachingRecordDetailTeachers(mainTeacherID, mainTeacherName, rawAssistantTeacherIDs, rawAssistantTeacherNames)

	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			CAST(str.id AS CHAR),
			CAST(str.student_id AS CHAR),
			str.student_name,
			str.student_phone,
			str.avatar_url,
			str.status,
			str.source_type,
			IFNULL(str.quantity, 0),
			CASE
				WHEN IFNULL(str.arrear_quantity, 0) > 0 AND IFNULL(str.actual_tuition, 0) <= 0 THEN 0
				ELSE IFNULL(str.actual_quantity, 0)
			END,
			str.remark,
			str.external_remark,
			CAST(str.tuition_account_id AS CHAR),
			str.tuition_account_name,
			IFNULL(ta.status, 0),
			IFNULL(ta.remaining_quantity, 0),
			IFNULL(str.sku_mode, 0),
			IFNULL(str.amount, 0),
			IFNULL(str.actual_deduct, 0),
			IFNULL(str.actual_tuition, 0),
			IFNULL(str.arrear_quantity, 0),
			DATE_FORMAT(str.record_time, '%Y-%m-%d %H:%i:%s'),
			DATE_FORMAT(str.updated_time, '%Y-%m-%d %H:%i:%s'),
			str.updated_staff_name
		FROM student_teaching_record str
		LEFT JOIN tuition_account ta
			ON ta.id = str.tuition_account_id
		   AND ta.inst_id = str.inst_id
		   AND ta.del_flag = 0
		WHERE str.inst_id = ?
		  AND str.del_flag = 0
		  AND str.teaching_record_id = ?
		ORDER BY str.id ASC
	`, instID, teachingRecordID)
	if err != nil {
		return model.TeachingRecordDetailResult{}, err
	}
	defer rows.Close()

	result.StudentList = make([]model.TeachingRecordDetailStudent, 0)
	for rows.Next() {
		var item model.TeachingRecordDetailStudent
		var accountStatus int
		if err := rows.Scan(
			&item.StudentTeachingRecordID,
			&item.StudentID,
			&item.StudentName,
			&item.StudentPhone,
			&item.Avatar,
			&item.Status,
			&item.SourceType,
			&item.Quantity,
			&item.ActualQuantity,
			&item.Remark,
			&item.ExternalRemark,
			&item.TuitionAccountID,
			&item.TuitionAccountName,
			&accountStatus,
			&item.LeftQuantity,
			&item.SkuMode,
			&item.Amount,
			&item.ActualDeduct,
			&item.ActualTuition,
			&item.ArrearQuantity,
			&item.RecordTime,
			&item.UpdatedTime,
			&item.UpdatedStaffName,
		); err != nil {
			return model.TeachingRecordDetailResult{}, err
		}
		item.IsTuitionAccountActive = accountStatus == model.TuitionAccountStatusActive
		result.StudentList = append(result.StudentList, item)
	}
	if err := rows.Err(); err != nil {
		return model.TeachingRecordDetailResult{}, err
	}
	result.StudentList, err = repo.mergeTeachingRecordDetailStudentList(ctx, instID, result.TimetableSourceID, result.StudentList)
	if err != nil {
		return model.TeachingRecordDetailResult{}, err
	}
	if err := repo.fillTeachingRecordDetailAttendanceStats(ctx, instID, &result); err != nil {
		return model.TeachingRecordDetailResult{}, err
	}
	return result, nil
}

func (repo *Repository) DeleteTeachingRecord(ctx context.Context, instID, operatorID int64, dto model.DeleteTeachingRecordDTO) (bool, error) {
	teachingRecordID, err := strconv.ParseInt(strings.TrimSpace(dto.TeachingRecordID), 10, 64)
	if err != nil || teachingRecordID <= 0 {
		return false, errors.New("缺少有效的上课点名记录")
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	rows, err := repo.loadTeachingRecordDeleteRowsTx(ctx, tx, instID, teachingRecordID)
	if err != nil {
		return false, err
	}
	if len(rows) == 0 {
		return false, errors.New("未找到上课点名记录")
	}

	accountMap, err := repo.loadTeachingRecordDeleteAccountMapTx(ctx, tx, instID, rows)
	if err != nil {
		return false, err
	}

	for _, row := range rows {
		if row.TuitionAccountID <= 0 || row.ActualDeduct <= 0 {
			continue
		}
		key := strconv.FormatInt(row.TuitionAccountID, 10)
		account, ok := accountMap[key]
		if !ok {
			continue
		}
		if err := repo.revertTeachingRecordConsumeTx(ctx, tx, instID, operatorID, teachingRecordID, row, account); err != nil {
			return false, err
		}
		account.UsedQuantity = math.Max(roundMoney(account.UsedQuantity-row.ActualDeduct), 0)
		account.UsedTuition = math.Max(roundMoney(account.UsedTuition-row.ActualTuition), 0)
		account.ConfirmedTuition = math.Max(roundMoney(account.ConfirmedTuition-row.ActualTuition), 0)
		account.RemainingQuantity = roundMoney(math.Max(account.TotalQuantity-account.UsedQuantity, 0))
		account.RemainingTuition = roundMoney(math.Max(account.TotalTuition-account.UsedTuition, 0))
		accountMap[key] = account
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE student_teaching_record
		SET del_flag = 1,
		    updated_staff_id = ?,
		    updated_staff_name = ?,
		    updated_time = NOW(),
		    update_id = ?,
		    update_time = NOW()
		WHERE inst_id = ?
		  AND teaching_record_id = ?
		  AND del_flag = 0
	`, operatorID, firstNonEmptyString(repo.GetStaffNameByID(ctx, &operatorID), "系统"), operatorID, instID, teachingRecordID); err != nil {
		return false, err
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}

func (repo *Repository) loadTeachingRecordDeleteRowsTx(ctx context.Context, tx *sql.Tx, instID, teachingRecordID int64) ([]teachingRecordDeleteStudentRow, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT
			IFNULL(id, 0),
			IFNULL(teaching_schedule_id, 0),
			IFNULL(student_id, 0),
			IFNULL(tuition_account_id, 0),
			IFNULL(actual_deduct, 0),
			IFNULL(actual_tuition, 0)
		FROM student_teaching_record
		WHERE inst_id = ?
		  AND teaching_record_id = ?
		  AND del_flag = 0
		ORDER BY id ASC
	`, instID, teachingRecordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]teachingRecordDeleteStudentRow, 0)
	for rows.Next() {
		var item teachingRecordDeleteStudentRow
		if err := rows.Scan(
			&item.StudentTeachingRecordID,
			&item.TeachingScheduleID,
			&item.StudentID,
			&item.TuitionAccountID,
			&item.ActualDeduct,
			&item.ActualTuition,
		); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (repo *Repository) loadTeachingRecordDeleteAccountMapTx(ctx context.Context, tx *sql.Tx, instID int64, rows []teachingRecordDeleteStudentRow) (map[string]rollCallConfirmAccount, error) {
	accountIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		if row.TuitionAccountID > 0 && row.ActualDeduct > 0 {
			accountIDs = append(accountIDs, row.TuitionAccountID)
		}
	}
	accountIDs = uniquePositiveInt64s(accountIDs)
	if len(accountIDs) == 0 {
		return map[string]rollCallConfirmAccount{}, nil
	}

	queryArgs := append([]any{instID}, int64SliceToAny(accountIDs)...)
	query := `
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
		  AND ta.id IN (` + sqlPlaceholders(len(accountIDs)) + `)
		FOR UPDATE
	`

	queryRows, err := tx.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer queryRows.Close()

	result := make(map[string]rollCallConfirmAccount, len(accountIDs))
	for queryRows.Next() {
		var item rollCallConfirmAccount
		if err := queryRows.Scan(
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
	if err := queryRows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (repo *Repository) revertTeachingRecordConsumeTx(ctx context.Context, tx *sql.Tx, instID, operatorID, teachingRecordID int64, row teachingRecordDeleteStudentRow, account rollCallConfirmAccount) error {
	newUsedQuantity := math.Max(roundMoney(account.UsedQuantity-row.ActualDeduct), 0)
	newUsedTuition := math.Max(roundMoney(account.UsedTuition-row.ActualTuition), 0)
	newConfirmedTuition := math.Max(roundMoney(account.ConfirmedTuition-row.ActualTuition), 0)
	newRemainingQuantity := roundMoney(math.Max(account.TotalQuantity-newUsedQuantity, 0))
	newRemainingTuition := roundMoney(math.Max(account.TotalTuition-newUsedTuition, 0))

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
	`,
		instID,
		account.ID,
		account.StudentID,
		account.CourseID,
		account.LessonType,
		normalizeRollCallDrawerChargingMode(account.LessonChargingMode),
		model.TuitionAccountFlowSourceConsumeReturn,
		teachingRecordID,
		teachingRecordID,
		account.OrderNumber,
		roundMoney(row.ActualDeduct),
		roundMoney(-row.ActualTuition),
		newRemainingQuantity,
		newRemainingTuition,
		operatorID,
		operatorID,
	)
	return err
}

func (repo *Repository) mergeTeachingRecordDetailStudentList(ctx context.Context, instID int64, scheduleIDText string, existing []model.TeachingRecordDetailStudent) ([]model.TeachingRecordDetailStudent, error) {
	scheduleID, err := strconv.ParseInt(strings.TrimSpace(scheduleIDText), 10, 64)
	if err != nil || scheduleID <= 0 {
		return existing, nil
	}
	scheduleMetaMap, err := repo.loadClassRecordScheduleMetaMap(ctx, instID, []int64{scheduleID})
	if err != nil {
		return nil, err
	}
	meta, ok := scheduleMetaMap[scheduleID]
	if !ok || meta.ClassType != model.TeachingClassTypeNormal || meta.ClassID <= 0 {
		return existing, nil
	}

	rosterByScheduleID, err := repo.loadEffectiveGroupClassScheduleRosterMap(ctx, repo.db, instID, []effectiveGroupClassScheduleMeta{{
		ScheduleID: meta.ScheduleID,
		ClassID:    meta.ClassID,
		StartAt:    meta.StartAt,
	}})
	if err != nil {
		return nil, err
	}
	roster := rosterByScheduleID[scheduleID]
	if len(roster.Active) == 0 {
		return existing, nil
	}

	recordByStudentID := make(map[int64]model.TeachingRecordDetailStudent, len(existing))
	result := make([]model.TeachingRecordDetailStudent, 0, len(roster.Active)+len(existing))
	appendedStudentIDs := make(map[int64]struct{}, len(roster.Active))

	for _, student := range existing {
		studentID, parseErr := strconv.ParseInt(strings.TrimSpace(student.StudentID), 10, 64)
		if parseErr != nil || studentID <= 0 {
			continue
		}
		if _, exists := recordByStudentID[studentID]; !exists {
			recordByStudentID[studentID] = student
		}
	}

	for _, student := range roster.Active {
		if student.StudentID <= 0 {
			continue
		}
		if record, ok := recordByStudentID[student.StudentID]; ok {
			result = append(result, record)
		} else {
			result = append(result, buildPendingTeachingRecordDetailStudent(student))
		}
		appendedStudentIDs[student.StudentID] = struct{}{}
	}

	for _, student := range existing {
		studentID, parseErr := strconv.ParseInt(strings.TrimSpace(student.StudentID), 10, 64)
		if parseErr != nil || studentID <= 0 {
			result = append(result, student)
			continue
		}
		if _, exists := appendedStudentIDs[studentID]; exists {
			continue
		}
		result = append(result, student)
	}

	return result, nil
}

func (repo *Repository) loadClassRecordScheduleMetaMap(ctx context.Context, instID int64, scheduleIDs []int64) (map[int64]classRecordScheduleMeta, error) {
	scheduleIDs = uniquePositiveInt64s(scheduleIDs)
	if len(scheduleIDs) == 0 {
		return map[int64]classRecordScheduleMeta{}, nil
	}
	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			IFNULL(id, 0),
			IFNULL(class_type, 0),
			IFNULL(teaching_class_id, 0),
			IFNULL(student_id, 0),
			lesson_start_at
		FROM teaching_schedule
		WHERE inst_id = ?
		  AND del_flag = 0
		  AND status = ?
		  AND id IN (`+sqlPlaceholders(len(scheduleIDs))+`)
	`, append([]any{instID, model.TeachingScheduleStatusActive}, int64SliceToAny(scheduleIDs)...)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64]classRecordScheduleMeta, len(scheduleIDs))
	for rows.Next() {
		var item classRecordScheduleMeta
		if err := rows.Scan(&item.ScheduleID, &item.ClassType, &item.ClassID, &item.StudentID, &item.StartAt); err != nil {
			return nil, err
		}
		if item.ScheduleID > 0 {
			result[item.ScheduleID] = item
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (repo *Repository) fillScheduleTeachingRecordStats(ctx context.Context, instID int64, items []model.ScheduleTeachingRecordItem) error {
	if len(items) == 0 {
		return nil
	}

	scheduleIDs := make([]int64, 0, len(items))
	for _, item := range items {
		scheduleID, err := strconv.ParseInt(strings.TrimSpace(item.TimetableSourceID), 10, 64)
		if err == nil && scheduleID > 0 {
			scheduleIDs = append(scheduleIDs, scheduleID)
		}
	}
	scheduleMetaMap, err := repo.loadClassRecordScheduleMetaMap(ctx, instID, scheduleIDs)
	if err != nil {
		return err
	}
	if len(scheduleMetaMap) == 0 {
		return nil
	}

	rollCallMetas := make([]teachingScheduleRollCallMeta, 0, len(scheduleMetaMap))
	groupMetas := make([]effectiveGroupClassScheduleMeta, 0, len(scheduleMetaMap))
	for _, meta := range scheduleMetaMap {
		rollCallMetas = append(rollCallMetas, teachingScheduleRollCallMeta{
			ScheduleID: meta.ScheduleID,
			ClassType:  meta.ClassType,
			ClassID:    meta.ClassID,
			StudentID:  meta.StudentID,
			StartAt:    meta.StartAt,
		})
		if meta.ClassType == model.TeachingClassTypeNormal && meta.ClassID > 0 {
			groupMetas = append(groupMetas, effectiveGroupClassScheduleMeta{
				ScheduleID: meta.ScheduleID,
				ClassID:    meta.ClassID,
				StartAt:    meta.StartAt,
			})
		}
	}
	statusByID, err := repo.computeTeachingScheduleCallStatusMap(ctx, repo.db, instID, rollCallMetas)
	if err != nil {
		return err
	}
	rosterByScheduleID := map[int64]groupClassScheduleRoster{}
	if len(groupMetas) > 0 {
		rosterByScheduleID, err = repo.loadEffectiveGroupClassScheduleRosterMap(ctx, repo.db, instID, groupMetas)
		if err != nil {
			return err
		}
	}

	for i := range items {
		scheduleID, err := strconv.ParseInt(strings.TrimSpace(items[i].TimetableSourceID), 10, 64)
		if err != nil || scheduleID <= 0 {
			continue
		}
		meta, ok := scheduleMetaMap[scheduleID]
		if !ok {
			continue
		}
		items[i].RollCallStatus = classRecordRollCallStatus(statusByID[scheduleID])
		switch {
		case meta.ClassType == model.TeachingClassTypeNormal && meta.ClassID > 0:
			items[i].ShouldAttendCount = len(rosterByScheduleID[scheduleID].activeIDs())
		case meta.StudentID > 0:
			items[i].ShouldAttendCount = 1
		}
		if items[i].ShouldAttendCount > 0 {
			items[i].AttendanceRate = float64(items[i].AttendCount) / float64(items[i].ShouldAttendCount)
		} else {
			items[i].AttendanceRate = 0
		}
	}
	return nil
}

func (repo *Repository) fillTeachingRecordDetailAttendanceStats(ctx context.Context, instID int64, detail *model.TeachingRecordDetailResult) error {
	if detail == nil {
		return nil
	}
	scheduleID, err := strconv.ParseInt(strings.TrimSpace(detail.TimetableSourceID), 10, 64)
	if err != nil || scheduleID <= 0 {
		return nil
	}
	scheduleMetaMap, err := repo.loadClassRecordScheduleMetaMap(ctx, instID, []int64{scheduleID})
	if err != nil {
		return err
	}
	meta, ok := scheduleMetaMap[scheduleID]
	if !ok {
		return nil
	}
	switch {
	case meta.ClassType == model.TeachingClassTypeNormal && meta.ClassID > 0:
		rosterByScheduleID, err := repo.loadEffectiveGroupClassScheduleRosterMap(ctx, repo.db, instID, []effectiveGroupClassScheduleMeta{{
			ScheduleID: meta.ScheduleID,
			ClassID:    meta.ClassID,
			StartAt:    meta.StartAt,
		}})
		if err != nil {
			return err
		}
		detail.ShouldAttendanceCount = len(rosterByScheduleID[scheduleID].activeIDs())
	case meta.StudentID > 0:
		detail.ShouldAttendanceCount = 1
	}
	return nil
}

func buildPendingTeachingRecordDetailStudent(student groupClassScheduleStudent) model.TeachingRecordDetailStudent {
	return model.TeachingRecordDetailStudent{
		StudentID:      emptyStringIfZero(student.StudentID),
		StudentName:    firstNonEmptyString(strings.TrimSpace(student.StudentName), "该学员"),
		StudentPhone:   strings.TrimSpace(student.Phone),
		Avatar:         strings.TrimSpace(student.AvatarURL),
		Status:         teachingRecordDetailStudentStatusPendingRollCall,
		SourceType:     rollCallTeachingRecordSourceType(student.ScheduleStudentType),
		Quantity:       0,
		ActualQuantity: 0,
		Amount:         0,
		ActualDeduct:   0,
		ActualTuition:  0,
		ArrearQuantity: 0,
	}
}

func buildTeachingRecordDetailTeachers(mainTeacherID int64, mainTeacherName, rawAssistantTeacherIDs, rawAssistantTeacherNames string) []model.TeachingRecordDetailTeacher {
	result := make([]model.TeachingRecordDetailTeacher, 0)
	name := strings.TrimSpace(mainTeacherName)
	if mainTeacherID > 0 || name != "" {
		result = append(result, model.TeachingRecordDetailTeacher{
			TeacherID:   emptyStringIfZero(mainTeacherID),
			TeacherName: firstNonEmptyString(name, "-"),
			Type:        1,
			Status:      1,
			Quantity:    0,
		})
	}

	assistantIDs := decodeJSONStringArray([]byte(rawAssistantTeacherIDs))
	assistantNames := decodeJSONStringArray([]byte(rawAssistantTeacherNames))
	for index, assistantName := range assistantNames {
		assistantName = strings.TrimSpace(assistantName)
		if assistantName == "" {
			continue
		}
		assistantID := ""
		if index < len(assistantIDs) {
			assistantID = strings.TrimSpace(assistantIDs[index])
		}
		result = append(result, model.TeachingRecordDetailTeacher{
			TeacherID:   assistantID,
			TeacherName: assistantName,
			Type:        3,
			Status:      1,
			Quantity:    0,
		})
	}
	return result
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
