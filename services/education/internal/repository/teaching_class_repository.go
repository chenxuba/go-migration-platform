package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

// tuitionAccountQuotationJoinForTa 与学费账户列表一致，解析报价单（quote_id / 订单明细 / 量价匹配 / 课程首条报价）
const tuitionAccountQuotationJoinForTa = `
LEFT JOIN sale_order_course_detail sod_ta ON sod_ta.id = ta.order_course_detail_id AND sod_ta.del_flag = 0
LEFT JOIN inst_course_quotation icq ON icq.id = COALESCE(
	NULLIF(ta.quote_id, 0),
	NULLIF(sod_ta.quote_id, 0),
	(SELECT qx.id FROM inst_course_quotation qx
	 WHERE qx.course_id = ta.course_id AND qx.del_flag = 0
	   AND ABS(IFNULL(qx.quantity, 0) - IFNULL(ta.total_quantity, 0)) < 0.000001
	   AND ABS(IFNULL(qx.price, 0) - IFNULL(ta.total_tuition, 0)) < 0.000001
	 ORDER BY qx.id DESC LIMIT 1),
	(SELECT qmin.id FROM inst_course_quotation qmin
	 WHERE qmin.course_id = ta.course_id AND qmin.del_flag = 0
	 ORDER BY qmin.id ASC LIMIT 1)
) AND icq.del_flag = 0`

func ensureTeachingClassTables(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS teaching_class (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			uuid VARCHAR(64) NULL,
			version BIGINT NOT NULL DEFAULT 0,
			inst_id BIGINT NOT NULL,
			class_type INT NOT NULL DEFAULT 1,
			course_id BIGINT NOT NULL DEFAULT 0,
			name VARCHAR(255) NOT NULL DEFAULT '',
			advisor_id BIGINT NOT NULL DEFAULT 0,
			default_teacher_id BIGINT NOT NULL DEFAULT 0,
			status INT NOT NULL DEFAULT 1,
			scheduled_lesson_count INT NOT NULL DEFAULT 0,
			finished_lesson_count INT NOT NULL DEFAULT 0,
			class_room_id BIGINT NOT NULL DEFAULT 0,
			class_room_name VARCHAR(255) NOT NULL DEFAULT '',
			classroom_enabled TINYINT(1) NULL DEFAULT NULL,
			remark VARCHAR(150) NOT NULL DEFAULT '',
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT NOT NULL DEFAULT 0,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			KEY idx_teaching_class_inst_type (inst_id, class_type, del_flag),
			KEY idx_teaching_class_course (inst_id, course_id),
			KEY idx_teaching_class_advisor (inst_id, advisor_id),
			KEY idx_teaching_class_default_teacher (inst_id, default_teacher_id),
			KEY idx_teaching_class_created (inst_id, create_time, id)
		)
	`); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS teaching_class_student (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			uuid VARCHAR(64) NULL,
			version BIGINT NOT NULL DEFAULT 0,
			inst_id BIGINT NOT NULL,
			teaching_class_id BIGINT NOT NULL,
			student_id BIGINT NOT NULL,
			order_id BIGINT NOT NULL DEFAULT 0,
			order_course_detail_id BIGINT NOT NULL DEFAULT 0,
			quote_id BIGINT NOT NULL DEFAULT 0,
			primary_tuition_account_id BIGINT NOT NULL DEFAULT 0,
			class_student_status INT NOT NULL DEFAULT 1,
			class_time DECIMAL(18,2) NOT NULL DEFAULT 1,
			student_class_time DECIMAL(18,2) NOT NULL DEFAULT 1,
			teacher_class_time DECIMAL(18,2) NOT NULL DEFAULT 0,
			class_time_record_mode INT NOT NULL DEFAULT 1,
			last_finished_lesson_day DATETIME NULL DEFAULT NULL,
			class_properties_json TEXT NULL,
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT NOT NULL DEFAULT 0,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uk_teaching_class_student_order_detail (inst_id, order_course_detail_id),
			KEY idx_teaching_class_student_class (inst_id, teaching_class_id),
			KEY idx_teaching_class_student_student (inst_id, student_id),
			KEY idx_teaching_class_student_tuition (inst_id, primary_tuition_account_id)
		)
	`); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS teaching_class_teacher (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			uuid VARCHAR(64) NULL,
			version BIGINT NOT NULL DEFAULT 0,
			inst_id BIGINT NOT NULL,
			teaching_class_id BIGINT NOT NULL,
			teacher_id BIGINT NOT NULL,
			status INT NOT NULL DEFAULT 1,
			is_default TINYINT(1) NOT NULL DEFAULT 0,
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT NOT NULL DEFAULT 0,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uk_teaching_class_teacher (inst_id, teaching_class_id, teacher_id),
			KEY idx_teaching_class_teacher_class (inst_id, teaching_class_id),
			KEY idx_teaching_class_teacher_teacher (inst_id, teacher_id)
		)
	`); err != nil {
		return err
	}
	if err := ensureColumnsOnTable(ctx, db, "teaching_class", map[string]string{
		"remark": "remark VARCHAR(150) NOT NULL DEFAULT '' AFTER classroom_enabled",
	}); err != nil {
		return err
	}
	return ensureColumnsOnTable(ctx, db, "teaching_class_student", map[string]string{
		"class_properties_json":  "class_properties_json TEXT NULL AFTER last_finished_lesson_day",
		"class_time_record_mode": "class_time_record_mode INT NOT NULL DEFAULT 1 AFTER teacher_class_time",
	})
}

func (repo *Repository) CountTeachingClassByName(ctx context.Context, instID int64, classType int, name string, excludeID *int64) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM teaching_class
		WHERE inst_id = ? AND class_type = ? AND name = ? AND del_flag = 0
	`
	args := []any{instID, classType, strings.TrimSpace(name)}
	if excludeID != nil {
		query += " AND id <> ?"
		args = append(args, *excludeID)
	}
	var count int
	err := repo.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

func (repo *Repository) getCourseTeachMethodMapTx(ctx context.Context, tx *sql.Tx, courseIDs []int64) (map[int64]int, error) {
	result := make(map[int64]int, len(courseIDs))
	if len(courseIDs) == 0 {
		return result, nil
	}
	placeholders := make([]string, 0, len(courseIDs))
	args := make([]any, 0, len(courseIDs))
	for _, id := range courseIDs {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT id, IFNULL(teach_method, 0)
		FROM inst_course
		WHERE del_flag = 0 AND id IN (`+strings.Join(placeholders, ",")+`)
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var courseID int64
		var teachMethod int
		if err := rows.Scan(&courseID, &teachMethod); err != nil {
			return nil, err
		}
		result[courseID] = teachMethod
	}
	return result, rows.Err()
}

func (repo *Repository) upsertOneToOneTeachingClassTx(ctx context.Context, tx *sql.Tx, instID, operatorID, orderID, studentID, courseID, quoteID, orderCourseDetailID, primaryTuitionAccountID int64, now time.Time) error {
	var existingClassID int64
	err := tx.QueryRowContext(ctx, `
		SELECT tc.id
		FROM teaching_class_student tcs
		INNER JOIN teaching_class tc ON tc.id = tcs.teaching_class_id AND tc.del_flag = 0
		WHERE tcs.inst_id = ? AND tcs.order_course_detail_id = ? AND tcs.del_flag = 0 AND tc.class_type = ?
		LIMIT 1
	`, instID, orderCourseDetailID, model.TeachingClassTypeOneToOne).Scan(&existingClassID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	var (
		studentName string
		courseName  string
	)
	if err := tx.QueryRowContext(ctx, `
		SELECT IFNULL(s.stu_name, ''), IFNULL(c.name, '')
		FROM inst_student s
		INNER JOIN inst_course c ON c.id = ?
		WHERE s.id = ? AND s.del_flag = 0 AND c.del_flag = 0
		LIMIT 1
	`, courseID, studentID).Scan(&studentName, &courseName); err != nil {
		return err
	}
	className := strings.TrimSpace(studentName + "-" + courseName)
	if className == "-" || className == "" {
		className = courseName
	}

	if existingClassID > 0 {
		if _, err := tx.ExecContext(ctx, `
			UPDATE teaching_class
			SET course_id = ?, name = ?, update_id = ?, update_time = ?
			WHERE id = ? AND inst_id = ? AND del_flag = 0
		`, courseID, className, operatorID, now, existingClassID, instID); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `
			UPDATE teaching_class_student
			SET student_id = ?, order_id = ?, quote_id = ?, primary_tuition_account_id = ?, update_id = ?, update_time = ?
			WHERE teaching_class_id = ? AND inst_id = ? AND order_course_detail_id = ? AND del_flag = 0
		`, studentID, orderID, quoteID, primaryTuitionAccountID, operatorID, now, existingClassID, instID, orderCourseDetailID)
		return err
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO teaching_class (
			uuid, version, inst_id, class_type, course_id, name, advisor_id, default_teacher_id, status,
			scheduled_lesson_count, finished_lesson_count, class_room_id, class_room_name, classroom_enabled,
			create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, ?, ?, ?, 0, 0, ?, 0, 0, 0, '', NULL, ?, ?, ?, ?, 0
		)
	`, instID, model.TeachingClassTypeOneToOne, courseID, className, model.TeachingClassStatusActive, operatorID, now, operatorID, now)
	if err != nil {
		return err
	}
	classID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO teaching_class_student (
			uuid, version, inst_id, teaching_class_id, student_id, order_id, order_course_detail_id, quote_id,
			primary_tuition_account_id, class_student_status, class_time, student_class_time, teacher_class_time,
			last_finished_lesson_day, class_properties_json, create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, ?, ?, ?, ?, ?, ?, ?, 1, 1, 0, NULL, NULL, ?, ?, ?, ?, 0
		)
	`, instID, classID, studentID, orderID, orderCourseDetailID, quoteID, primaryTuitionAccountID, model.TeachingClassStudentStatusStudying, operatorID, now, operatorID, now)
	return err
}

func (repo *Repository) PageOneToOneList(ctx context.Context, instID int64, query model.OneToOneListQueryDTO) (model.OneToOneListResultVO, error) {
	current := query.PageRequestModel.PageIndex
	size := query.PageRequestModel.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	whereSQL, args := buildOneToOneWhere(instID, query.QueryModel, false)

	var total int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM teaching_class tc
		INNER JOIN teaching_class_student tcs ON tcs.teaching_class_id = tc.id AND tcs.inst_id = tc.inst_id AND tcs.del_flag = 0
		INNER JOIN inst_student s ON s.id = tcs.student_id AND s.del_flag = 0
		INNER JOIN inst_course c ON c.id = tc.course_id AND c.del_flag = 0
		WHERE `+whereSQL, args...).Scan(&total); err != nil {
		return model.OneToOneListResultVO{}, err
	}

	var studentCount int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT tcs.student_id)
		FROM teaching_class tc
		INNER JOIN teaching_class_student tcs ON tcs.teaching_class_id = tc.id AND tcs.inst_id = tc.inst_id AND tcs.del_flag = 0
		INNER JOIN inst_student s ON s.id = tcs.student_id AND s.del_flag = 0
		INNER JOIN inst_course c ON c.id = tc.course_id AND c.del_flag = 0
		WHERE `+whereSQL, args...).Scan(&studentCount); err != nil {
		return model.OneToOneListResultVO{}, err
	}

	queryArgs := append([]any{instID}, args...)
	queryArgs = append(queryArgs, size, offset)
	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			tc.id,
			IFNULL(tc.name, ''),
			tcs.student_id,
			IFNULL(s.stu_name, ''),
			IFNULL(s.stu_sex, 2),
			IFNULL(s.avatar_url, ''),
			CASE
				WHEN CHAR_LENGTH(IFNULL(s.mobile, '')) >= 7 THEN CONCAT(LEFT(s.mobile, 3), '****', RIGHT(s.mobile, 4))
				ELSE IFNULL(s.mobile, '')
			END AS phone,
			tc.status,
			tcs.class_student_status,
			tc.create_time,
			IFNULL(tc.class_room_id, 0),
			IFNULL(tc.class_room_name, ''),
			tc.classroom_enabled,
			IFNULL(tcs.class_time, 0),
			IFNULL(tcs.student_class_time, 0),
			IFNULL(tcs.teacher_class_time, 0),
			tc.course_id,
			IFNULL(c.name, ''),
			IFNULL(ts.primary_tuition_account_id, IFNULL(tcs.primary_tuition_account_id, 0)),
			CAST(IFNULL(tcs.order_course_detail_id, 0) AS CHAR) AS order_course_detail_id,
			IFNULL(tc.default_teacher_id, 0),
			IFNULL(default_teacher.nick_name, ''),
			IFNULL(tcs.class_time_record_mode, 1),
			IFNULL(ts.has_grade_upgrade, 0),
			tcs.last_finished_lesson_day,
			IFNULL(tcs.class_properties_json, '[]'),
			IFNULL(tc.advisor_id, 0),
			IFNULL(advisor.nick_name, ''),
			IFNULL(tc.remark, ''),
			CASE WHEN IFNULL(tc.scheduled_lesson_count, 0) > 0 THEN 1 ELSE 0 END,
			IFNULL(tc.finished_lesson_count, 0),
			IFNULL(ts.total_tuition, 0),
			IFNULL(ts.remain_tuition, 0),
			IFNULL(ts.total_quantity, 0),
			IFNULL(ts.total_free_quantity, 0),
			IFNULL(ts.remain_quantity, 0),
			IFNULL(ts.remain_free_quantity, 0),
			IFNULL(ts.lesson_charging_mode, 0),
			IFNULL(ts.lesson_scope_model, 0),
			IFNULL(ts.status, 0),
			IFNULL(ts.enable_expire_time, 0),
			ts.expire_time,
			ts.change_status_time,
			ts.suspended_time,
			ts.class_ending_time,
			IFNULL(ts.assigned_class, 0)
		FROM teaching_class tc
		INNER JOIN teaching_class_student tcs ON tcs.teaching_class_id = tc.id AND tcs.inst_id = tc.inst_id AND tcs.del_flag = 0
		INNER JOIN inst_student s ON s.id = tcs.student_id AND s.del_flag = 0
		INNER JOIN inst_course c ON c.id = tc.course_id AND c.del_flag = 0
		LEFT JOIN inst_user advisor ON advisor.id = tc.advisor_id
		LEFT JOIN inst_user default_teacher ON default_teacher.id = tc.default_teacher_id
		LEFT JOIN (
			SELECT
				ta.order_course_detail_id,
				MIN(ta.id) AS primary_tuition_account_id,
				SUM(IFNULL(ta.total_tuition, 0)) AS total_tuition,
				SUM(IFNULL(ta.remaining_tuition, 0)) AS remain_tuition,
				SUM(CASE
					WHEN IFNULL(icq.lesson_model, 0) = 3 THEN IFNULL(ta.total_tuition, 0)
					WHEN IFNULL(ta.total_quantity, 0) > 0 THEN IFNULL(ta.total_quantity, 0)
					ELSE 0
				END) AS total_quantity,
				SUM(CASE
					WHEN IFNULL(icq.lesson_model, 0) = 3 THEN IFNULL(ta.free_quantity, 0)
					WHEN IFNULL(ta.total_quantity, 0) = 0 AND IFNULL(ta.free_quantity, 0) > 0 THEN IFNULL(ta.free_quantity, 0)
					ELSE 0
				END) AS total_free_quantity,
				SUM(CASE
					WHEN IFNULL(icq.lesson_model, 0) = 3 THEN IFNULL(ta.remaining_tuition, 0)
					WHEN IFNULL(ta.total_quantity, 0) > 0 THEN IFNULL(ta.remaining_quantity, 0)
					ELSE 0
				END) AS remain_quantity,
				SUM(CASE
					WHEN IFNULL(icq.lesson_model, 0) = 3 THEN IFNULL(ta.free_quantity, 0)
					WHEN IFNULL(ta.total_quantity, 0) = 0 AND IFNULL(ta.free_quantity, 0) > 0 THEN IFNULL(ta.remaining_quantity, 0)
					ELSE 0
				END) AS remain_free_quantity,
				MAX(
					CASE
						WHEN IFNULL(icq.lesson_model, 0) > 0 THEN icq.lesson_model
						WHEN IFNULL(ta.enable_expire_time, 0) = 1 AND IFNULL(ta.total_quantity, 0) > 0 THEN 2
						ELSE 0
					END
				) AS lesson_charging_mode,
				MAX(IFNULL(ic.course_type, 0)) AS lesson_scope_model,
				MAX(IFNULL(ta.status, 0)) AS status,
				IFNULL(MAX(ta.enable_expire_time), 0) AS enable_expire_time,
				MAX(ta.expire_time) AS expire_time,
				MAX(ta.status_change_time) AS change_status_time,
				MAX(ta.suspended_time) AS suspended_time,
				MAX(ta.class_ending_time) AS class_ending_time,
				IFNULL(MAX(ta.assigned_class), 0) AS assigned_class,
				IFNULL(MAX(ta.has_grade_upgrade), 0) AS has_grade_upgrade
			FROM tuition_account ta
			LEFT JOIN inst_course ic ON ic.id = ta.course_id AND ic.del_flag = 0
			` + tuitionAccountQuotationJoinForTa + `
			WHERE ta.inst_id = ? AND ta.del_flag = 0
			GROUP BY ta.order_course_detail_id
		) ts ON ts.order_course_detail_id = tcs.order_course_detail_id
		WHERE `+whereSQL+`
		ORDER BY tc.create_time DESC, tc.id DESC
		LIMIT ? OFFSET ?
	`, queryArgs...)
	if err != nil {
		return model.OneToOneListResultVO{}, err
	}
	defer rows.Close()

	items := make([]model.OneToOneItemVO, 0, size)
	classIDs := make([]int64, 0, size)
	for rows.Next() {
		var (
			item                  model.OneToOneItemVO
			classID               int64
			studentID             int64
			status                int
			classStudentStatus    int
			courseID              int64
			primaryTuitionAccount int64
			defaultTeacherID      int64
			classTeacherID        int64
			classRoomID           int64
			classroomEnabled      sql.NullBool
			classPropertiesJSON   string
			lastFinishedLessonDay sql.NullTime
			expireTime            sql.NullTime
			changeStatusTime      sql.NullTime
			suspendedTime         sql.NullTime
			classEndingTime       sql.NullTime
			createdTime           sql.NullTime
		)
		if err := rows.Scan(
			&classID,
			&item.Name,
			&studentID,
			&item.StudentName,
			&item.Sex,
			&item.Avatar,
			&item.Phone,
			&status,
			&classStudentStatus,
			&createdTime,
			&classRoomID,
			&item.ClassRoomName,
			&classroomEnabled,
			&item.ClassTime,
			&item.StudentClassTime,
			&item.TeacherClassTime,
			&courseID,
			&item.LessonName,
			&primaryTuitionAccount,
			&item.OrderCourseDetailID,
			&defaultTeacherID,
			&item.DefaultTeacherName,
			&item.DefaultClassTimeRecordMode,
			&item.IsGradeUpgrade,
			&lastFinishedLessonDay,
			&classPropertiesJSON,
			&classTeacherID,
			&item.ClassTeacherName,
			&item.Remark,
			&item.One2OneLessonDayInfo.LessonDayCount,
			&item.One2OneLessonDayInfo.CompleteLessonDayCount,
			&item.TuitionAccount.TotalTuition,
			&item.TuitionAccount.RemainTuition,
			&item.TuitionAccount.TotalQuantity,
			&item.TuitionAccount.TotalFreeQuantity,
			&item.TuitionAccount.RemainQuantity,
			&item.TuitionAccount.RemainFreeQuantity,
			&item.TuitionAccount.LessonChargingMode,
			&item.TuitionAccount.LessonScopeModel,
			&item.TuitionAccount.Status,
			&item.TuitionAccount.EnableExpireTime,
			&expireTime,
			&changeStatusTime,
			&suspendedTime,
			&classEndingTime,
			&item.TuitionAccount.AssignedClass,
		); err != nil {
			return model.OneToOneListResultVO{}, err
		}
		classIDs = append(classIDs, classID)
		item.ID = strconv.FormatInt(classID, 10)
		item.StudentID = strconv.FormatInt(studentID, 10)
		item.SchoolID = strconv.FormatInt(instID, 10)
		item.Status = status
		item.ClassStudentStatus = classStudentStatus
		item.IsScheduled = item.One2OneLessonDayInfo.LessonDayCount > 0
		item.ClassRoomID = strconv.FormatInt(classRoomID, 10)
		item.LessonID = strconv.FormatInt(courseID, 10)
		item.TuitionAccountID = strconv.FormatInt(primaryTuitionAccount, 10)
		item.TuitionAccount.ID = item.TuitionAccountID
		item.TuitionAccount.StudentID = item.StudentID
		item.TuitionAccount.LessonID = item.LessonID
		item.TuitionAccount.LessonType = model.TeachingClassTypeOneToOne
		item.TuitionAccount.ProductName = item.LessonName
		item.TuitionAccount.AssignedClass = item.TuitionAccount.AssignedClass
		item.DefaultTeacherID = strconv.FormatInt(defaultTeacherID, 10)
		if defaultTeacherID <= 0 {
			item.DefaultTeacherID = "0"
		}
		item.ClassTeacherID = strconv.FormatInt(classTeacherID, 10)
		if classTeacherID <= 0 {
			item.ClassTeacherID = "0"
		}
		if createdTime.Valid {
			item.CreatedTime = createdTime.Time
		}
		if classroomEnabled.Valid {
			value := classroomEnabled.Bool
			item.ClassroomEnabled = &value
		}
		if lastFinishedLessonDay.Valid {
			item.LastFinishedLessonDay = lastFinishedLessonDay.Time
		}
		item.TuitionAccount.LastSuspendedTime = zeroTimeFromNull(suspendedTime)
		item.TuitionAccount.ExpireTime = zeroTimeFromNull(expireTime)
		item.TuitionAccount.ChangeStatusTime = zeroTimeFromNull(changeStatusTime)
		if suspendedTime.Valid {
			t := suspendedTime.Time
			item.TuitionAccount.SuspendedTime = &t
		}
		if classEndingTime.Valid {
			t := classEndingTime.Time
			item.TuitionAccount.ClassEndingTime = &t
		}
		item.One2OneLessonTimes = []model.OneToOneLessonTimeVO{}
		item.ClassProperties = []model.OneToOnePropertyVO{}
		if strings.TrimSpace(classPropertiesJSON) != "" {
			_ = json.Unmarshal([]byte(classPropertiesJSON), &item.ClassProperties)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return model.OneToOneListResultVO{}, err
	}

	teacherMap, err := repo.listTeachingClassTeachers(ctx, instID, classIDs)
	if err != nil {
		return model.OneToOneListResultVO{}, err
	}
	for idx := range items {
		classID, _ := strconv.ParseInt(items[idx].ID, 10, 64)
		items[idx].TeacherList = teacherMap[classID]
		items[idx].ClassTeacherName = classTeacherNamesFromTeacherList(items[idx].TeacherList, strings.TrimSpace(items[idx].ClassTeacherName))
	}

	return model.OneToOneListResultVO{
		Total:        total,
		StudentCount: studentCount,
		List:         items,
	}, nil
}

func (repo *Repository) GetOneToOneDetail(ctx context.Context, instID, classID int64) (model.OneToOneDetailVO, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT
			tc.id,
			tcs.student_id,
			IFNULL(tc.name, ''),
			IFNULL(s.stu_name, ''),
			IFNULL(s.avatar_url, ''),
			IFNULL(s.stu_sex, 2),
			tc.course_id,
			IFNULL(c.name, ''),
			IFNULL(icq.price, 0),
			IFNULL(tc.class_room_id, 0),
			tc.class_room_name,
			tc.classroom_enabled,
			IFNULL(ts.primary_tuition_account_id, IFNULL(tcs.primary_tuition_account_id, 0)),
			CAST(IFNULL(tcs.order_course_detail_id, 0) AS CHAR) AS order_course_detail_id,
			IFNULL(tcs.class_time, 0),
			CASE WHEN IFNULL(tc.scheduled_lesson_count, 0) > 0 THEN 1 ELSE 0 END,
			IFNULL(tc.status, 1),
			IFNULL(tcs.class_student_status, 1),
			tc.create_time,
			IFNULL(tcs.student_class_time, 0),
			IFNULL(tcs.teacher_class_time, 0),
			IFNULL(tcs.class_time_record_mode, 1),
			IFNULL(advisor.nick_name, ''),
			IFNULL(tc.default_teacher_id, 0),
			IFNULL(default_teacher.nick_name, ''),
			IFNULL(ts.has_grade_upgrade, 0),
			IFNULL(tc.remark, ''),
			IFNULL(tc.create_id, 0),
			IFNULL(created_staff.nick_name, ''),
			IFNULL(default_teacher_rel.status, 0),
			IFNULL(tcs.class_properties_json, '[]'),
			IFNULL(ts.total_tuition, 0),
			IFNULL(ts.remain_tuition, 0),
			IFNULL(ts.total_quantity, 0),
			IFNULL(ts.total_free_quantity, 0),
			IFNULL(ts.remain_quantity, 0),
			IFNULL(ts.remain_free_quantity, 0),
			IFNULL(ts.lesson_charging_mode, 0),
			IFNULL(ts.lesson_scope_model, 0),
			IFNULL(ts.status, 0),
			IFNULL(ts.enable_expire_time, 0),
			ts.expire_time,
			ts.change_status_time,
			ts.suspended_time,
			ts.class_ending_time,
			IFNULL(ts.assigned_class, 0)
		FROM teaching_class tc
		INNER JOIN teaching_class_student tcs ON tcs.teaching_class_id = tc.id AND tcs.inst_id = tc.inst_id AND tcs.del_flag = 0
		INNER JOIN inst_student s ON s.id = tcs.student_id AND s.del_flag = 0
		INNER JOIN inst_course c ON c.id = tc.course_id AND c.del_flag = 0
		LEFT JOIN inst_course_quotation icq ON icq.id = tcs.quote_id AND icq.del_flag = 0
		LEFT JOIN inst_user advisor ON advisor.id = tc.advisor_id
		LEFT JOIN inst_user default_teacher ON default_teacher.id = tc.default_teacher_id
		LEFT JOIN inst_user created_staff ON created_staff.id = tc.create_id
		LEFT JOIN teaching_class_teacher default_teacher_rel
			ON default_teacher_rel.teaching_class_id = tc.id
			AND default_teacher_rel.inst_id = tc.inst_id
			AND default_teacher_rel.teacher_id = tc.default_teacher_id
			AND default_teacher_rel.del_flag = 0
		LEFT JOIN (
			SELECT
				ta.order_course_detail_id,
				MIN(ta.id) AS primary_tuition_account_id,
				SUM(IFNULL(ta.total_tuition, 0)) AS total_tuition,
				SUM(IFNULL(ta.remaining_tuition, 0)) AS remain_tuition,
				SUM(CASE
					WHEN IFNULL(icq.lesson_model, 0) = 3 THEN IFNULL(ta.total_tuition, 0)
					WHEN IFNULL(ta.total_quantity, 0) > 0 THEN IFNULL(ta.total_quantity, 0)
					ELSE 0
				END) AS total_quantity,
				SUM(CASE
					WHEN IFNULL(icq.lesson_model, 0) = 3 THEN IFNULL(ta.free_quantity, 0)
					WHEN IFNULL(ta.total_quantity, 0) = 0 AND IFNULL(ta.free_quantity, 0) > 0 THEN IFNULL(ta.free_quantity, 0)
					ELSE 0
				END) AS total_free_quantity,
				SUM(CASE
					WHEN IFNULL(icq.lesson_model, 0) = 3 THEN IFNULL(ta.remaining_tuition, 0)
					WHEN IFNULL(ta.total_quantity, 0) > 0 THEN IFNULL(ta.remaining_quantity, 0)
					ELSE 0
				END) AS remain_quantity,
				SUM(CASE
					WHEN IFNULL(icq.lesson_model, 0) = 3 THEN IFNULL(ta.free_quantity, 0)
					WHEN IFNULL(ta.total_quantity, 0) = 0 AND IFNULL(ta.free_quantity, 0) > 0 THEN IFNULL(ta.remaining_quantity, 0)
					ELSE 0
				END) AS remain_free_quantity,
				MAX(
					CASE
						WHEN IFNULL(icq.lesson_model, 0) > 0 THEN icq.lesson_model
						WHEN IFNULL(ta.enable_expire_time, 0) = 1 AND IFNULL(ta.total_quantity, 0) > 0 THEN 2
						ELSE 0
					END
				) AS lesson_charging_mode,
				MAX(IFNULL(ic.course_type, 0)) AS lesson_scope_model,
				MAX(IFNULL(ta.status, 0)) AS status,
				IFNULL(MAX(ta.enable_expire_time), 0) AS enable_expire_time,
				MAX(ta.expire_time) AS expire_time,
				MAX(ta.status_change_time) AS change_status_time,
				MAX(ta.suspended_time) AS suspended_time,
				MAX(ta.class_ending_time) AS class_ending_time,
				IFNULL(MAX(ta.assigned_class), 0) AS assigned_class,
				IFNULL(MAX(ta.has_grade_upgrade), 0) AS has_grade_upgrade
			FROM tuition_account ta
			LEFT JOIN inst_course ic ON ic.id = ta.course_id AND ic.del_flag = 0
			` + tuitionAccountQuotationJoinForTa + `
			WHERE ta.inst_id = ? AND ta.del_flag = 0
			GROUP BY ta.order_course_detail_id
		) ts ON ts.order_course_detail_id = tcs.order_course_detail_id
		WHERE tc.inst_id = ? AND tc.id = ? AND tc.class_type = ? AND tc.del_flag = 0
		LIMIT 1
	`, instID, instID, classID, model.TeachingClassTypeOneToOne)

	var (
		detail              model.OneToOneDetailVO
		classIDValue        int64
		studentID           int64
		courseID            int64
		classRoomID         int64
		tuitionAccountID    int64
		defaultTeacherID    int64
		createdStaffID      int64
		classroomName       sql.NullString
		classroomEnabled    sql.NullBool
		isScheduled         bool
		expireTime          sql.NullTime
		changeStatusTime    sql.NullTime
		suspendedTime       sql.NullTime
		classEndingTime     sql.NullTime
		classPropertiesJSON string
		advisorName         string
	)

	if err := row.Scan(
		&classIDValue,
		&studentID,
		&detail.Name,
		&detail.StudentName,
		&detail.StudentAvatar,
		&detail.StudentGender,
		&courseID,
		&detail.LessonName,
		&detail.LessonPrice,
		&classRoomID,
		&classroomName,
		&classroomEnabled,
		&tuitionAccountID,
		&detail.OrderCourseDetailID,
		&detail.ClassTime,
		&isScheduled,
		&detail.Status,
		&detail.ClassStudentStatus,
		&detail.CreatedTime,
		&detail.DefaultStudentClassTime,
		&detail.DefaultTeacherClassTime,
		&detail.DefaultClassTimeRecordMode,
		&advisorName,
		&defaultTeacherID,
		&detail.DefaultTeacherName,
		&detail.IsGradeUpgrade,
		&detail.Remark,
		&createdStaffID,
		&detail.CreatedStaffName,
		&detail.DefaultTeacherStatus,
		&classPropertiesJSON,
		&detail.TuitionAccount.TotalTuition,
		&detail.TuitionAccount.RemainTuition,
		&detail.TuitionAccount.TotalQuantity,
		&detail.TuitionAccount.TotalFreeQuantity,
		&detail.TuitionAccount.RemainQuantity,
		&detail.TuitionAccount.RemainFreeQuantity,
		&detail.TuitionAccount.LessonChargingMode,
		&detail.TuitionAccount.LessonScopeModel,
		&detail.TuitionAccount.Status,
		&detail.TuitionAccount.EnableExpireTime,
		&expireTime,
		&changeStatusTime,
		&suspendedTime,
		&classEndingTime,
		&detail.TuitionAccount.AssignedClass,
	); err != nil {
		return model.OneToOneDetailVO{}, err
	}

	detail.ID = strconv.FormatInt(classIDValue, 10)
	detail.StudentID = strconv.FormatInt(studentID, 10)
	detail.SchoolID = strconv.FormatInt(instID, 10)
	detail.IsScheduled = isScheduled
	detail.LessonID = strconv.FormatInt(courseID, 10)
	detail.ClassroomID = strconv.FormatInt(classRoomID, 10)
	detail.TuitionAccountID = strconv.FormatInt(tuitionAccountID, 10)
	detail.DefaultTeacherID = strconv.FormatInt(defaultTeacherID, 10)
	detail.CreatedStaffID = strconv.FormatInt(createdStaffID, 10)
	if defaultTeacherID <= 0 {
		detail.DefaultTeacherID = "0"
	}
	if classroomName.Valid {
		value := classroomName.String
		detail.ClassroomName = &value
	}
	if classroomEnabled.Valid {
		value := classroomEnabled.Bool
		detail.ClassroomEnabled = &value
	}
	if strings.TrimSpace(classPropertiesJSON) != "" {
		_ = json.Unmarshal([]byte(classPropertiesJSON), &detail.ClassProperties)
	}
	if detail.ClassProperties == nil {
		detail.ClassProperties = []model.OneToOnePropertyVO{}
	}
	if detail.DefaultTeacherStatus <= 0 && defaultTeacherID > 0 {
		detail.DefaultTeacherStatus = 1
	}
	detail.TeacherList = []model.OneToOneTeacherVO{}
	detail.TuitionAccount.ID = detail.TuitionAccountID
	detail.TuitionAccount.StudentID = detail.StudentID
	detail.TuitionAccount.LessonID = detail.LessonID
	detail.TuitionAccount.ProductName = detail.LessonName
	detail.TuitionAccount.LessonType = model.TeachingClassTypeOneToOne
	detail.TuitionAccount.LastSuspendedTime = zeroTimeFromNull(suspendedTime)
	detail.TuitionAccount.ExpireTime = zeroTimeFromNull(expireTime)
	detail.TuitionAccount.ChangeStatusTime = zeroTimeFromNull(changeStatusTime)
	if suspendedTime.Valid {
		t := suspendedTime.Time
		detail.TuitionAccount.SuspendedTime = &t
	}
	if classEndingTime.Valid {
		t := classEndingTime.Time
		detail.TuitionAccount.ClassEndingTime = &t
	}

	teacherMap, err := repo.listTeachingClassTeachers(ctx, instID, []int64{classIDValue})
	if err != nil {
		return model.OneToOneDetailVO{}, err
	}
	detail.TeacherList = teacherMap[classIDValue]
	detail.ClassTeacherName = classTeacherNamesFromTeacherList(detail.TeacherList, strings.TrimSpace(advisorName))
	return detail, nil
}

func (repo *Repository) UpdateOneToOne(ctx context.Context, instID, operatorID int64, dto model.OneToOneUpdateDTO) error {
	classID, err := strconv.ParseInt(strings.TrimSpace(dto.ID), 10, 64)
	if err != nil || classID <= 0 {
		return sql.ErrNoRows
	}
	studentID, _ := strconv.ParseInt(strings.TrimSpace(dto.StudentID), 10, 64)
	lessonID, _ := strconv.ParseInt(strings.TrimSpace(dto.LessonID), 10, 64)
	defaultTeacherID, _ := strconv.ParseInt(strings.TrimSpace(dto.DefaultTeacherID), 10, 64)
	teacherIDs := normalizeTeacherIDs(dto.TeacherID, defaultTeacherID)
	classProperties := dto.ClassProperties
	if classProperties == nil {
		classProperties = []model.OneToOnePropertyVO{}
	}
	classPropertiesJSON, err := json.Marshal(classProperties)
	if err != nil {
		return err
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM teaching_class tc
		INNER JOIN teaching_class_student tcs ON tcs.teaching_class_id = tc.id AND tcs.inst_id = tc.inst_id AND tcs.del_flag = 0
		WHERE tc.id = ? AND tc.inst_id = ? AND tc.class_type = ? AND tc.del_flag = 0
	`, classID, instID, model.TeachingClassTypeOneToOne).Scan(&exists); err != nil {
		return err
	}
	if exists == 0 {
		return sql.ErrNoRows
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE teaching_class
		SET name = ?, course_id = ?, default_teacher_id = ?, remark = ?, update_id = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND class_type = ? AND del_flag = 0
	`,
		strings.TrimSpace(dto.Name),
		lessonID,
		defaultTeacherID,
		strings.TrimSpace(dto.Remark),
		operatorID,
		classID,
		instID,
		model.TeachingClassTypeOneToOne,
	); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE teaching_class_student
		SET student_id = ?, class_time = ?, student_class_time = ?, teacher_class_time = ?, class_time_record_mode = ?,
		    class_properties_json = ?, update_id = ?, update_time = NOW()
		WHERE teaching_class_id = ? AND inst_id = ? AND del_flag = 0
	`,
		studentID,
		dto.DefaultStudentClassTime,
		dto.DefaultStudentClassTime,
		dto.DefaultTeacherClassTime,
		dto.DefaultClassTimeRecordMode,
		string(classPropertiesJSON),
		operatorID,
		classID,
		instID,
	); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE teaching_class_teacher
		SET del_flag = 1, update_id = ?, update_time = NOW()
		WHERE inst_id = ? AND teaching_class_id = ? AND del_flag = 0
	`, operatorID, instID, classID); err != nil {
		return err
	}

	now := time.Now()
	for _, teacherID := range teacherIDs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO teaching_class_teacher (
				uuid, version, inst_id, teaching_class_id, teacher_id, status, is_default,
				create_id, create_time, update_id, update_time, del_flag
			) VALUES (
				UUID(), 0, ?, ?, ?, 1, ?, ?, ?, ?, ?, 0
			)
			ON DUPLICATE KEY UPDATE
				status = VALUES(status),
				is_default = VALUES(is_default),
				del_flag = 0,
				update_id = VALUES(update_id),
				update_time = VALUES(update_time)
		`,
			instID,
			classID,
			teacherID,
			boolToTinyInt(teacherID == defaultTeacherID),
			operatorID,
			now,
			operatorID,
			now,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// CloseOneToOneOnly 将 1 对 1 班级标记为已结班（不结课、不删日程）
func (repo *Repository) CloseOneToOneOnly(ctx context.Context, instID, operatorID, classID int64) error {
	var currentStatus int
	err := repo.db.QueryRowContext(ctx, `
		SELECT tc.status
		FROM teaching_class tc
		WHERE tc.id = ? AND tc.inst_id = ? AND tc.class_type = ? AND tc.del_flag = 0
	`, classID, instID, model.TeachingClassTypeOneToOne).Scan(&currentStatus)
	if err != nil {
		return err
	}
	if currentStatus == model.TeachingClassStatusClosed {
		return nil
	}
	if currentStatus != model.TeachingClassStatusActive {
		return errors.New("班级状态不允许结班")
	}
	res, err := repo.db.ExecContext(ctx, `
		UPDATE teaching_class
		SET status = ?, update_id = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND class_type = ? AND del_flag = 0 AND status = ?
	`, model.TeachingClassStatusClosed, operatorID, classID, instID, model.TeachingClassTypeOneToOne, model.TeachingClassStatusActive)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ReopenOneToOneOnly 将已结班的 1 对 1 恢复为开班中
func (repo *Repository) ReopenOneToOneOnly(ctx context.Context, instID, operatorID, classID int64) error {
	var currentStatus int
	err := repo.db.QueryRowContext(ctx, `
		SELECT tc.status
		FROM teaching_class tc
		WHERE tc.id = ? AND tc.inst_id = ? AND tc.class_type = ? AND tc.del_flag = 0
	`, classID, instID, model.TeachingClassTypeOneToOne).Scan(&currentStatus)
	if err != nil {
		return err
	}
	if currentStatus == model.TeachingClassStatusActive {
		return nil
	}
	if currentStatus != model.TeachingClassStatusClosed {
		return errors.New("班级状态不允许恢复开班")
	}
	var courseClosedCount int
	err = repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM teaching_class_student tcs
		WHERE tcs.teaching_class_id = ? AND tcs.inst_id = ? AND tcs.del_flag = 0
		  AND IFNULL(tcs.primary_tuition_account_id, 0) > 0
		  AND tcs.class_student_status = ?
	`, classID, instID, model.TeachingClassStudentStatusClosed).Scan(&courseClosedCount)
	if err != nil {
		return err
	}
	if courseClosedCount > 0 {
		return errors.New("该1对1默认账户的课程已结课，无法恢复开班")
	}
	res, err := repo.db.ExecContext(ctx, `
		UPDATE teaching_class
		SET status = ?, update_id = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND class_type = ? AND del_flag = 0 AND status = ?
	`, model.TeachingClassStatusActive, operatorID, classID, instID, model.TeachingClassTypeOneToOne, model.TeachingClassStatusClosed)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func normalizeTeacherIDs(ids []string, defaultTeacherID int64) []int64 {
	result := make([]int64, 0, len(ids)+1)
	seen := make(map[int64]struct{}, len(ids)+1)
	for _, raw := range ids {
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
	if defaultTeacherID > 0 {
		if _, ok := seen[defaultTeacherID]; !ok {
			result = append(result, defaultTeacherID)
		}
	}
	return result
}

func boolToTinyInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

// classTeacherNamesFromTeacherList 列表/详情「班主任」：展示本班 teaching_class_teacher 全部关联教师（按 teacher_id 去重）。
// 默认上课教师在库中常为 is_default=1，若只展示 is_default=0 会漏掉与班主任重复的默认教师（如 王明+汪洋 只显示一人）。
func classTeacherNamesFromTeacherList(list []model.OneToOneTeacherVO, advisorFallback string) string {
	if len(list) == 0 {
		return advisorFallback
	}
	names := make([]string, 0)
	seen := make(map[string]struct{})
	for _, t := range list {
		n := strings.TrimSpace(t.Name)
		if n == "" {
			continue
		}
		key := strings.TrimSpace(t.TeacherID)
		if key == "" {
			key = n
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		names = append(names, n)
	}
	if len(names) > 0 {
		return strings.Join(names, "、")
	}
	return advisorFallback
}

func buildOneToOneWhere(instID int64, query model.OneToOneListQueryModel, excludeQuickFilters bool) (string, []any) {
	whereParts := []string{
		"tc.inst_id = ?",
		"tc.class_type = ?",
		"tc.del_flag = 0",
	}
	args := []any{instID, model.TeachingClassTypeOneToOne}

	if strings.TrimSpace(query.StudentID) != "" {
		whereParts = append(whereParts, "CAST(tcs.student_id AS CHAR) = ?")
		args = append(args, strings.TrimSpace(query.StudentID))
	}
	if len(query.LessonIDs) > 0 {
		placeholders := make([]string, 0, len(query.LessonIDs))
		for _, lessonID := range query.LessonIDs {
			lessonID = strings.TrimSpace(lessonID)
			if lessonID == "" {
				continue
			}
			placeholders = append(placeholders, "?")
			args = append(args, lessonID)
		}
		if len(placeholders) > 0 {
			whereParts = append(whereParts, "CAST(tc.course_id AS CHAR) IN ("+strings.Join(placeholders, ",")+")")
		}
	}
	if tid := strings.TrimSpace(query.ClassTeacherID); tid != "" {
		whereParts = append(whereParts, `(
			CAST(tc.advisor_id AS CHAR) = ?
			OR EXISTS (
				SELECT 1 FROM teaching_class_teacher tct
				WHERE tct.teaching_class_id = tc.id AND tct.inst_id = tc.inst_id AND tct.del_flag = 0
					AND CAST(tct.teacher_id AS CHAR) = ?
			)
		)`)
		args = append(args, tid, tid)
	}
	if strings.TrimSpace(query.DefaultTeacherID) != "" {
		whereParts = append(whereParts, "CAST(tc.default_teacher_id AS CHAR) = ?")
		args = append(args, strings.TrimSpace(query.DefaultTeacherID))
	}
	if !excludeQuickFilters && query.HasClassTeacher != nil {
		if boolValue(query.HasClassTeacher) {
			whereParts = append(whereParts, `(
				IFNULL(tc.advisor_id, 0) > 0
				OR EXISTS (
					SELECT 1 FROM teaching_class_teacher tct
					WHERE tct.teaching_class_id = tc.id AND tct.inst_id = tc.inst_id AND tct.del_flag = 0
				)
			)`)
		} else {
			whereParts = append(whereParts, `(
				IFNULL(tc.advisor_id, 0) = 0
				AND NOT EXISTS (
					SELECT 1 FROM teaching_class_teacher tct
					WHERE tct.teaching_class_id = tc.id AND tct.inst_id = tc.inst_id AND tct.del_flag = 0
				)
			)`)
		}
	}
	if !excludeQuickFilters && query.IsScheduled != nil {
		if boolValue(query.IsScheduled) {
			whereParts = append(whereParts, "IFNULL(tc.scheduled_lesson_count, 0) > 0")
		} else {
			whereParts = append(whereParts, "IFNULL(tc.scheduled_lesson_count, 0) <= 0")
		}
	}
	if len(query.Status) > 0 {
		placeholders := make([]string, 0, len(query.Status))
		for _, item := range query.Status {
			placeholders = append(placeholders, "?")
			args = append(args, item)
		}
		whereParts = append(whereParts, "tc.status IN ("+strings.Join(placeholders, ",")+")")
	}
	if len(query.ClassStudentStatus) > 0 {
		placeholders := make([]string, 0, len(query.ClassStudentStatus))
		for _, item := range query.ClassStudentStatus {
			placeholders = append(placeholders, "?")
			args = append(args, item)
		}
		whereParts = append(whereParts, "tcs.class_student_status IN ("+strings.Join(placeholders, ",")+")")
	}
	if start := parseDateStart(query.StartDate); start != nil {
		whereParts = append(whereParts, "tc.create_time >= ?")
		args = append(args, *start)
	}
	if end := parseDateEnd(query.EndDate); end != nil {
		whereParts = append(whereParts, "tc.create_time <= ?")
		args = append(args, *end)
	}
	return strings.Join(whereParts, " AND "), args
}

func zeroTimeFromNull(value sql.NullTime) time.Time {
	if value.Valid {
		return value.Time
	}
	return time.Time{}
}

func (repo *Repository) listTeachingClassTeachers(ctx context.Context, instID int64, classIDs []int64) (map[int64][]model.OneToOneTeacherVO, error) {
	result := make(map[int64][]model.OneToOneTeacherVO)
	if len(classIDs) == 0 {
		return result, nil
	}
	placeholders := make([]string, 0, len(classIDs))
	args := make([]any, 0, len(classIDs)+1)
	args = append(args, instID)
	seen := make(map[int64]struct{}, len(classIDs))
	for _, id := range classIDs {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	rows, err := repo.db.QueryContext(ctx, `
		SELECT t.teaching_class_id, t.teacher_id, IFNULL(u.nick_name, ''), IFNULL(t.status, 1), IFNULL(t.is_default, 0)
		FROM teaching_class_teacher t
		LEFT JOIN inst_user u ON u.id = t.teacher_id
		WHERE t.inst_id = ? AND t.del_flag = 0 AND t.teaching_class_id IN (`+strings.Join(placeholders, ",")+`)
		ORDER BY t.is_default ASC, t.id ASC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			classID   int64
			teacherID int64
			item      model.OneToOneTeacherVO
			isDef     int64
		)
		if err := rows.Scan(&classID, &teacherID, &item.Name, &item.Status, &isDef); err != nil {
			return nil, err
		}
		item.IsDefault = isDef != 0
		item.ClassID = strconv.FormatInt(classID, 10)
		item.TeacherID = strconv.FormatInt(teacherID, 10)
		result[classID] = append(result[classID], item)
	}
	return result, rows.Err()
}

func mergeAdvisorTeachersWithDefault(selected []int64, defaultTeacherID int64) []int64 {
	seen := make(map[int64]struct{}, len(selected)+1)
	out := make([]int64, 0, len(selected)+1)
	for _, id := range selected {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	if defaultTeacherID > 0 {
		if _, ok := seen[defaultTeacherID]; !ok {
			out = append(out, defaultTeacherID)
		}
	}
	return out
}

// BatchAssignOneToOneClassTeacher 批量设置班主任：列表主班主任取所选第一位；与单条编辑一致写入 teaching_class_teacher，并合并各校区的默认上课教师。
func (repo *Repository) BatchAssignOneToOneClassTeacher(ctx context.Context, instID, operatorID int64, classTeacherIDs []int64, teachingClassIDs []int64) error {
	if len(teachingClassIDs) == 0 {
		return nil
	}
	if len(classTeacherIDs) == 0 {
		return errors.New("请选择班主任")
	}
	firstAdvisor := classTeacherIDs[0]

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	ph := make([]string, 0, len(teachingClassIDs))
	args := make([]any, 0, len(teachingClassIDs)+5)
	args = append(args, firstAdvisor, operatorID)
	for _, id := range teachingClassIDs {
		ph = append(ph, "?")
		args = append(args, id)
	}
	args = append(args, instID, model.TeachingClassTypeOneToOne)
	if _, err := tx.ExecContext(ctx, `
		UPDATE teaching_class
		SET advisor_id = ?, update_id = ?, update_time = NOW()
		WHERE id IN (`+strings.Join(ph, ",")+`) AND inst_id = ? AND class_type = ? AND del_flag = 0
	`, args...); err != nil {
		return err
	}

	now := time.Now()
	for _, classID := range teachingClassIDs {
		var defaultTeacherID int64
		if err := tx.QueryRowContext(ctx, `
			SELECT IFNULL(default_teacher_id, 0) FROM teaching_class
			WHERE id = ? AND inst_id = ? AND class_type = ? AND del_flag = 0
		`, classID, instID, model.TeachingClassTypeOneToOne).Scan(&defaultTeacherID); err != nil {
			return err
		}
		merged := mergeAdvisorTeachersWithDefault(classTeacherIDs, defaultTeacherID)

		if _, err := tx.ExecContext(ctx, `
			UPDATE teaching_class_teacher
			SET del_flag = 1, update_id = ?, update_time = NOW()
			WHERE inst_id = ? AND teaching_class_id = ? AND del_flag = 0
		`, operatorID, instID, classID); err != nil {
			return err
		}

		for _, teacherID := range merged {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO teaching_class_teacher (
					uuid, version, inst_id, teaching_class_id, teacher_id, status, is_default,
					create_id, create_time, update_id, update_time, del_flag
				) VALUES (
					UUID(), 0, ?, ?, ?, 1, ?, ?, ?, ?, ?, 0
				)
				ON DUPLICATE KEY UPDATE
					status = VALUES(status),
					is_default = VALUES(is_default),
					del_flag = 0,
					update_id = VALUES(update_id),
					update_time = VALUES(update_time)
			`,
				instID,
				classID,
				teacherID,
				boolToTinyInt(teacherID == defaultTeacherID),
				operatorID,
				now,
				operatorID,
				now,
			); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (repo *Repository) BatchUpdateOneToOneClassTime(ctx context.Context, instID, operatorID int64, ids []int64, dto model.OneToOneBatchClassTimeDTO) error {
	if len(ids) == 0 {
		return nil
	}
	recordMode := dto.ClassTimeRecordMode
	if recordMode <= 0 {
		recordMode = 1
	}
	placeholders := make([]string, 0, len(ids))
	args := make([]any, 0, len(ids)+7)
	args = append(args, dto.ClassTime, dto.StudentClassTime, dto.TeacherClassTime, recordMode, operatorID)
	for _, id := range ids {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	args = append(args, instID, model.TeachingClassTypeOneToOne)
	_, err := repo.db.ExecContext(ctx, `
		UPDATE teaching_class_student tcs
		INNER JOIN teaching_class tc ON tc.id = tcs.teaching_class_id AND tc.inst_id = tcs.inst_id AND tc.del_flag = 0
		SET tcs.class_time = ?, tcs.student_class_time = ?, tcs.teacher_class_time = ?, tcs.class_time_record_mode = ?, tcs.update_id = ?, tcs.update_time = NOW()
		WHERE tc.id IN (`+strings.Join(placeholders, ",")+`) AND tc.inst_id = ? AND tc.class_type = ? AND tcs.del_flag = 0
	`, args...)
	return err
}

func parseIDStrings(ids []string) []int64 {
	result := make([]int64, 0, len(ids))
	for _, raw := range ids {
		value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
		if err != nil || value <= 0 {
			continue
		}
		result = append(result, value)
	}
	return result
}
