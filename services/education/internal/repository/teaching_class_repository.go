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

// tuitionAccountQuotationJoinForTaDed 与学费账户列表一致：解析报价单（quote_id / 订单明细 / 量价匹配 / 课程首条报价），别名 ta_ded/icq_ded。
const tuitionAccountQuotationJoinForTaDed = `
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
) AND icq_ded.del_flag = 0`

// oneToOneReadingListBucketFrom 与 GetTuitionAccountReadingList 分桶一致：按课程 id、授课方式、学费账户自身 quote_id 对应的 lesson_model（仅 ta.quote_id→icq，与报读列表 SQL 相同）。
const oneToOneReadingListBucketFrom = `
FROM tuition_account ta2
INNER JOIN inst_course ic2 ON ic2.id = ta2.course_id AND ic2.del_flag = 0
LEFT JOIN inst_course_quotation icq2 ON icq2.id = ta2.quote_id AND icq2.del_flag = 0
WHERE ta2.del_flag = 0
  AND ta2.inst_id = tcs.inst_id AND ta2.student_id = tcs.student_id
  AND ta2.course_id = ta_ded.course_id
  AND IFNULL(ic2.teach_method, 0) = IFNULL(ic_ded.teach_method, 0)
  AND IFNULL(icq2.lesson_model, -99999) = IFNULL(icq_read.lesson_model, -99999)`

// oneToOneTuitionAccountDeductionJoinSQL 列表/详情「当前课程账户」：tuitionAccountId 取扣费账户 ta_ded；展示用课时/学费按报读列表同桶 SUM（多条账户累加），与学员报读列表一致。
const oneToOneTuitionAccountDeductionJoinSQL = `
LEFT JOIN tuition_account ta_ded ON ta_ded.id = COALESCE(
	NULLIF(tcs.primary_tuition_account_id, 0),
	(SELECT MIN(ta0.id) FROM tuition_account ta0
	 WHERE ta0.order_course_detail_id = tcs.order_course_detail_id
	   AND ta0.inst_id = tcs.inst_id AND ta0.del_flag = 0)
) AND ta_ded.inst_id = tcs.inst_id AND ta_ded.del_flag = 0
LEFT JOIN inst_course ic_ded ON ic_ded.id = ta_ded.course_id AND ic_ded.del_flag = 0` + tuitionAccountQuotationJoinForTaDed + `
LEFT JOIN inst_course_quotation icq_read ON icq_read.id = ta_ded.quote_id AND icq_read.del_flag = 0`

// oneToOneTuitionAccountCountSQL 统计当前 1 对 1 相关账户数量：
// 当前课程自身账户 + 当前班级已绑定的扣费账户，最终按「课程 + 授课方式 + 计费模式」聚合计数，
// 避免同一账户桶下多笔原始 tuition_account 被重复计入。
const oneToOneTuitionAccountCountSQL = `
IFNULL((
	SELECT COUNT(DISTINCT CONCAT(
		CAST(ta_cnt.course_id AS CHAR), '#',
		CAST(IFNULL(ic_cnt.teach_method, 0) AS CHAR), '#',
		CAST(IFNULL(icq_cnt.lesson_model, -99999) AS CHAR)
	))
	FROM (
		SELECT ta_key.id AS account_id
		FROM tuition_account ta_key
		WHERE ta_key.inst_id = tc.inst_id
			AND ta_key.del_flag = 0
			AND ta_key.student_id = tcs.student_id
			AND ta_key.course_id = tc.course_id
		UNION ALL
		SELECT COALESCE(
			NULLIF(tcs_cnt.primary_tuition_account_id, 0),
			(SELECT MIN(ta0.id)
			 FROM tuition_account ta0
			 WHERE ta0.order_course_detail_id = tcs_cnt.order_course_detail_id
			   AND ta0.inst_id = tcs_cnt.inst_id
			   AND ta0.del_flag = 0)
		) AS account_id
		FROM teaching_class_student tcs_cnt
		WHERE tcs_cnt.inst_id = tc.inst_id
			AND tcs_cnt.del_flag = 0
			AND tcs_cnt.teaching_class_id = tc.id
	) account_candidates
	INNER JOIN tuition_account ta_cnt
		ON ta_cnt.id = account_candidates.account_id
		AND ta_cnt.inst_id = tc.inst_id
		AND ta_cnt.del_flag = 0
	INNER JOIN inst_course ic_cnt
		ON ic_cnt.id = ta_cnt.course_id
		AND ic_cnt.del_flag = 0
	LEFT JOIN sale_order_course_detail sod_cnt
		ON sod_cnt.id = ta_cnt.order_course_detail_id
		AND sod_cnt.del_flag = 0
	LEFT JOIN inst_course_quotation icq_cnt ON icq_cnt.id = COALESCE(
		NULLIF(ta_cnt.quote_id, 0),
		NULLIF(sod_cnt.quote_id, 0),
		(SELECT qx.id FROM inst_course_quotation qx
		 WHERE qx.course_id = ta_cnt.course_id AND qx.del_flag = 0
		   AND ABS(IFNULL(qx.quantity, 0) - IFNULL(ta_cnt.total_quantity, 0)) < 0.000001
		   AND ABS(IFNULL(qx.price, 0) - IFNULL(ta_cnt.total_tuition, 0)) < 0.000001
		 ORDER BY qx.id DESC LIMIT 1),
		(SELECT qmin.id FROM inst_course_quotation qmin
		 WHERE qmin.course_id = ta_cnt.course_id AND qmin.del_flag = 0
		 ORDER BY qmin.id ASC LIMIT 1)
	) AND icq_cnt.del_flag = 0
	WHERE account_candidates.account_id IS NOT NULL AND account_candidates.account_id > 0
), 0)`

// oneToOneListableJoinSQL 与 PageOneToOneList 主查询一致：仅统计/展示仍有有效班员、且学员与上课课程均未删除的 1 对 1（避免清空校区后残留 teaching_class 导致条数虚高、同名误判）。
const oneToOneListableJoinSQL = `
INNER JOIN (
	SELECT teaching_class_id, MIN(id) AS id
	FROM teaching_class_student
	WHERE inst_id = ? AND del_flag = 0
	GROUP BY teaching_class_id
) tcs_pick ON tcs_pick.teaching_class_id = tc.id
INNER JOIN teaching_class_student tcs ON tcs.id = tcs_pick.id AND tcs.inst_id = tc.inst_id AND tcs.del_flag = 0
INNER JOIN inst_student s ON s.id = tcs.student_id AND s.del_flag = 0
INNER JOIN inst_course c ON c.id = tc.course_id AND c.del_flag = 0`

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
			UNIQUE KEY uk_tcs_inst_class_ocd (inst_id, teaching_class_id, order_course_detail_id),
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
	if err := ensureColumnsOnTable(ctx, db, "teaching_class_student", map[string]string{
		"class_properties_json":  "class_properties_json TEXT NULL AFTER last_finished_lesson_day",
		"class_time_record_mode": "class_time_record_mode INT NOT NULL DEFAULT 1 AFTER teacher_class_time",
	}); err != nil {
		return err
	}
	return ensureTeachingClassStudentOCUniqueMigration(ctx, db)
}

// ensureTeachingClassStudentOCUniqueMigration 将 (inst, 订单明细) 全局唯一改为 (inst, 班级, 订单明细) 唯一，允许同一报读账户绑定多个 1 对 1 班级。
func ensureTeachingClassStudentOCUniqueMigration(ctx context.Context, db *sql.DB) error {
	var n int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM information_schema.statistics
		WHERE table_schema = DATABASE() AND table_name = 'teaching_class_student' AND index_name = 'uk_tcs_inst_class_ocd'
	`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	var oldN int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM information_schema.statistics
		WHERE table_schema = DATABASE() AND table_name = 'teaching_class_student' AND index_name = 'uk_teaching_class_student_order_detail'
	`).Scan(&oldN); err != nil {
		return err
	}
	if oldN > 0 {
		if _, err := db.ExecContext(ctx, `ALTER TABLE teaching_class_student DROP INDEX uk_teaching_class_student_order_detail`); err != nil {
			return err
		}
	}
	_, err := db.ExecContext(ctx, `ALTER TABLE teaching_class_student ADD UNIQUE KEY uk_tcs_inst_class_ocd (inst_id, teaching_class_id, order_course_detail_id)`)
	return err
}

// CountTeachingClassByName 按名称统计班级。1 对 1 仅与「开班中」且列表可见（有效班员+学员+课程未删）的班级判重；已结班、已删班级、无有效关联的不参与。
func (repo *Repository) CountTeachingClassByName(ctx context.Context, instID int64, classType int, name string, excludeID *int64) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM teaching_class tc
		WHERE tc.inst_id = ? AND tc.class_type = ? AND tc.name = ? AND tc.del_flag = 0
	`
	args := []any{instID, classType, strings.TrimSpace(name)}
	if classType == model.TeachingClassTypeOneToOne {
		query += ` AND tc.status = ?`
		args = append(args, model.TeachingClassStatusActive)
		query += `
			AND EXISTS (
				SELECT 1
				FROM teaching_class_student tcs
				INNER JOIN inst_student s ON s.id = tcs.student_id AND s.del_flag = 0
				INNER JOIN inst_course c ON c.id = tc.course_id AND c.del_flag = 0
				WHERE tcs.teaching_class_id = tc.id AND tcs.inst_id = tc.inst_id AND tcs.del_flag = 0
			)`
	}
	if excludeID != nil {
		query += " AND tc.id <> ?"
		args = append(args, *excludeID)
	}
	var count int
	err := repo.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

// ExistsActiveOneToOneForStudentCourse 是否存在学员在该课程下、状态为开班中的 1 对 1（对标 ExistOne2One）
func (repo *Repository) ExistsActiveOneToOneForStudentCourse(ctx context.Context, instID, studentID, courseID int64) (bool, error) {
	var one int
	err := repo.db.QueryRowContext(ctx, `
		SELECT 1
		FROM teaching_class tc
		INNER JOIN teaching_class_student tcs ON tcs.teaching_class_id = tc.id AND tcs.inst_id = tc.inst_id AND tcs.del_flag = 0
		INNER JOIN inst_student s ON s.id = tcs.student_id AND s.del_flag = 0
		INNER JOIN inst_course c ON c.id = tc.course_id AND c.del_flag = 0
		WHERE tc.inst_id = ?
			AND tc.course_id = ?
			AND tc.class_type = ?
			AND tc.del_flag = 0
			AND tc.status = ?
			AND tcs.student_id = ?
		LIMIT 1
	`, instID, courseID, model.TeachingClassTypeOneToOne, model.TeachingClassStatusActive, studentID).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
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
		ORDER BY tcs.id ASC
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

	// 学员在该课程下已有「开班中」1 对 1 时复用同一班级，只新增班员绑定；列表按班级汇总多条订单的课时/学费。
	var reuseClassID int64
	switch err = tx.QueryRowContext(ctx, `
		SELECT tc.id
		FROM teaching_class tc
		INNER JOIN teaching_class_student tcs ON tcs.teaching_class_id = tc.id AND tcs.inst_id = tc.inst_id AND tcs.del_flag = 0
		WHERE tc.inst_id = ? AND tc.class_type = ? AND tc.course_id = ? AND tc.del_flag = 0 AND tc.status = ?
			AND tcs.student_id = ?
		ORDER BY tc.id ASC
		LIMIT 1
	`, instID, model.TeachingClassTypeOneToOne, courseID, model.TeachingClassStatusActive, studentID).Scan(&reuseClassID); {
	case errors.Is(err, sql.ErrNoRows):
		reuseClassID = 0
	case err != nil:
		return err
	}

	if reuseClassID > 0 {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO teaching_class_student (
				uuid, version, inst_id, teaching_class_id, student_id, order_id, order_course_detail_id, quote_id,
				primary_tuition_account_id, class_student_status, class_time, student_class_time, teacher_class_time,
				last_finished_lesson_day, class_properties_json, create_id, create_time, update_id, update_time, del_flag
			) VALUES (
				UUID(), 0, ?, ?, ?, ?, ?, ?, ?, 1, 1, 1, 0, NULL, NULL, ?, ?, ?, ?, 0
			)
		`, instID, reuseClassID, studentID, orderID, orderCourseDetailID, quoteID, primaryTuitionAccountID, operatorID, now, operatorID, now)
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
			UUID(), 0, ?, ?, ?, ?, ?, ?, ?, 1, 1, 1, 0, NULL, NULL, ?, ?, ?, ?, 0
		)
	`, instID, classID, studentID, orderID, orderCourseDetailID, quoteID, primaryTuitionAccountID, operatorID, now, operatorID, now)
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

	countArgs := append([]any{instID}, args...)
	var total int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT tc.id)
		FROM teaching_class tc
		`+oneToOneListableJoinSQL+`
		WHERE `+whereSQL, countArgs...).Scan(&total); err != nil {
		return model.OneToOneListResultVO{}, err
	}

	var studentCount int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT tcs.student_id)
		FROM teaching_class tc
		`+oneToOneListableJoinSQL+`
		WHERE `+whereSQL, countArgs...).Scan(&studentCount); err != nil {
		return model.OneToOneListResultVO{}, err
	}

	queryArgs := make([]any, 0, 1+len(args)+2)
	queryArgs = append(queryArgs, instID)
	queryArgs = append(queryArgs, args...)
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
			IFNULL(ic_ded.name, ''),
			IFNULL(ta_ded.course_id, 0),
			IFNULL(ic_ded.teach_method, 0),
			IFNULL(ta_ded.id, IFNULL(tcs.primary_tuition_account_id, 0)),
			CAST(IFNULL(tcs.order_course_detail_id, 0) AS CHAR) AS order_course_detail_id,
			IFNULL(tc.default_teacher_id, 0),
			IFNULL(default_teacher.nick_name, ''),
			IFNULL(tcs.class_time_record_mode, 1),
			IFNULL(ta_ded.has_grade_upgrade, 0),
			tcs.last_finished_lesson_day,
			IFNULL(tcs.class_properties_json, '[]'),
			IFNULL(tc.advisor_id, 0),
			IFNULL(advisor.nick_name, ''),
			IFNULL(tc.remark, ''),
			CASE WHEN IFNULL(tc.scheduled_lesson_count, 0) > 0 THEN 1 ELSE 0 END,
			IFNULL(tc.finished_lesson_count, 0),
			`+oneToOneTuitionAccountCountSQL+`,
			IFNULL((
				SELECT SUM(IFNULL(ta2.total_tuition, 0)) `+oneToOneReadingListBucketFrom+`
			), IFNULL(ta_ded.total_tuition, 0)),
			IFNULL((
				SELECT SUM(IFNULL(ta2.remaining_tuition, 0)) `+oneToOneReadingListBucketFrom+`
			), IFNULL(ta_ded.remaining_tuition, 0)),
			IFNULL((
				SELECT SUM(CASE
					WHEN IFNULL(icq2.lesson_model, 0) = 3 THEN IFNULL(ta2.total_tuition, 0)
					WHEN IFNULL(ta2.total_quantity, 0) > 0 THEN IFNULL(ta2.total_quantity, 0)
					ELSE 0
				END) `+oneToOneReadingListBucketFrom+`
			), CASE
				WHEN IFNULL(icq_ded.lesson_model, 0) = 3 THEN IFNULL(ta_ded.total_tuition, 0)
				WHEN IFNULL(ta_ded.total_quantity, 0) > 0 THEN IFNULL(ta_ded.total_quantity, 0)
				ELSE 0
			END),
			IFNULL((
				SELECT SUM(CASE
					WHEN IFNULL(icq2.lesson_model, 0) = 3 THEN IFNULL(ta2.free_quantity, 0)
					WHEN IFNULL(ta2.total_quantity, 0) = 0 AND IFNULL(ta2.free_quantity, 0) > 0 THEN IFNULL(ta2.free_quantity, 0)
					ELSE 0
				END) `+oneToOneReadingListBucketFrom+`
			), CASE
				WHEN IFNULL(icq_ded.lesson_model, 0) = 3 THEN IFNULL(ta_ded.free_quantity, 0)
				WHEN IFNULL(ta_ded.total_quantity, 0) = 0 AND IFNULL(ta_ded.free_quantity, 0) > 0 THEN IFNULL(ta_ded.free_quantity, 0)
				ELSE 0
			END),
			IFNULL((
				SELECT SUM(CASE
					WHEN IFNULL(icq2.lesson_model, 0) = 3 THEN IFNULL(ta2.remaining_tuition, 0)
					WHEN IFNULL(ta2.total_quantity, 0) > 0 THEN IFNULL(ta2.remaining_quantity, 0)
					ELSE 0
				END) `+oneToOneReadingListBucketFrom+`
			), CASE
				WHEN IFNULL(icq_ded.lesson_model, 0) = 3 THEN IFNULL(ta_ded.remaining_tuition, 0)
				WHEN IFNULL(ta_ded.total_quantity, 0) > 0 THEN IFNULL(ta_ded.remaining_quantity, 0)
				ELSE 0
			END),
			IFNULL((
				SELECT SUM(CASE
					WHEN IFNULL(icq2.lesson_model, 0) = 3 THEN IFNULL(ta2.free_quantity, 0)
					WHEN IFNULL(ta2.total_quantity, 0) = 0 AND IFNULL(ta2.free_quantity, 0) > 0 THEN IFNULL(ta2.remaining_quantity, 0)
					ELSE 0
				END) `+oneToOneReadingListBucketFrom+`
			), CASE
				WHEN IFNULL(icq_ded.lesson_model, 0) = 3 THEN IFNULL(ta_ded.free_quantity, 0)
				WHEN IFNULL(ta_ded.total_quantity, 0) = 0 AND IFNULL(ta_ded.free_quantity, 0) > 0 THEN IFNULL(ta_ded.remaining_quantity, 0)
				ELSE 0
			END),
			CASE
				WHEN IFNULL(icq_ded.lesson_model, 0) > 0 THEN icq_ded.lesson_model
				WHEN IFNULL(ta_ded.enable_expire_time, 0) = 1 AND IFNULL(ta_ded.total_quantity, 0) > 0 THEN 2
				ELSE 0
			END,
			IFNULL(ta_ded.status, 0),
			IFNULL(ta_ded.enable_expire_time, 0),
			ta_ded.expire_time,
			ta_ded.status_change_time,
			ta_ded.suspended_time,
			ta_ded.class_ending_time,
			IFNULL(ta_ded.assigned_class, 0)
		FROM teaching_class tc
		`+oneToOneListableJoinSQL+`
		`+oneToOneTuitionAccountDeductionJoinSQL+`
		LEFT JOIN inst_user advisor ON advisor.id = tc.advisor_id
		LEFT JOIN inst_user default_teacher ON default_teacher.id = tc.default_teacher_id
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
			deductLessonName      string
			deductCourseID        int64
			deductTeachMethod     int
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
			&deductLessonName,
			&deductCourseID,
			&deductTeachMethod,
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
			&item.TuitionAccountCount,
			&item.TuitionAccount.TotalTuition,
			&item.TuitionAccount.RemainTuition,
			&item.TuitionAccount.TotalQuantity,
			&item.TuitionAccount.TotalFreeQuantity,
			&item.TuitionAccount.RemainQuantity,
			&item.TuitionAccount.RemainFreeQuantity,
			&item.TuitionAccount.LessonChargingMode,
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
		item.TuitionAccount.LessonID = strconv.FormatInt(deductCourseID, 10)
		if deductTeachMethod == model.TeachingClassTypeNormal || deductTeachMethod == model.TeachingClassTypeOneToOne {
			item.TuitionAccount.LessonType = deductTeachMethod
		} else {
			item.TuitionAccount.LessonType = model.TeachingClassTypeOneToOne
		}
		if strings.TrimSpace(deductLessonName) != "" {
			item.TuitionAccount.ProductName = strings.TrimSpace(deductLessonName)
		} else {
			item.TuitionAccount.ProductName = item.LessonName
		}
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
			IFNULL(ic_ded.name, ''),
			IFNULL(ta_ded.course_id, 0),
			IFNULL(ic_ded.teach_method, 0),
			IFNULL(icq.price, 0),
			IFNULL(tc.class_room_id, 0),
			tc.class_room_name,
			tc.classroom_enabled,
			IFNULL(ta_ded.id, IFNULL(tcs.primary_tuition_account_id, 0)),
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
			IFNULL(ta_ded.has_grade_upgrade, 0),
			IFNULL(tc.remark, ''),
			IFNULL(tc.create_id, 0),
			IFNULL(created_staff.nick_name, ''),
			IFNULL(default_teacher_rel.status, 0),
			IFNULL(tcs.class_properties_json, '[]'),
			`+oneToOneTuitionAccountCountSQL+`,
			IFNULL((
				SELECT SUM(IFNULL(ta2.total_tuition, 0)) `+oneToOneReadingListBucketFrom+`
			), IFNULL(ta_ded.total_tuition, 0)),
			IFNULL((
				SELECT SUM(IFNULL(ta2.remaining_tuition, 0)) `+oneToOneReadingListBucketFrom+`
			), IFNULL(ta_ded.remaining_tuition, 0)),
			IFNULL((
				SELECT SUM(CASE
					WHEN IFNULL(icq2.lesson_model, 0) = 3 THEN IFNULL(ta2.total_tuition, 0)
					WHEN IFNULL(ta2.total_quantity, 0) > 0 THEN IFNULL(ta2.total_quantity, 0)
					ELSE 0
				END) `+oneToOneReadingListBucketFrom+`
			), CASE
				WHEN IFNULL(icq_ded.lesson_model, 0) = 3 THEN IFNULL(ta_ded.total_tuition, 0)
				WHEN IFNULL(ta_ded.total_quantity, 0) > 0 THEN IFNULL(ta_ded.total_quantity, 0)
				ELSE 0
			END),
			IFNULL((
				SELECT SUM(CASE
					WHEN IFNULL(icq2.lesson_model, 0) = 3 THEN IFNULL(ta2.free_quantity, 0)
					WHEN IFNULL(ta2.total_quantity, 0) = 0 AND IFNULL(ta2.free_quantity, 0) > 0 THEN IFNULL(ta2.free_quantity, 0)
					ELSE 0
				END) `+oneToOneReadingListBucketFrom+`
			), CASE
				WHEN IFNULL(icq_ded.lesson_model, 0) = 3 THEN IFNULL(ta_ded.free_quantity, 0)
				WHEN IFNULL(ta_ded.total_quantity, 0) = 0 AND IFNULL(ta_ded.free_quantity, 0) > 0 THEN IFNULL(ta_ded.free_quantity, 0)
				ELSE 0
			END),
			IFNULL((
				SELECT SUM(CASE
					WHEN IFNULL(icq2.lesson_model, 0) = 3 THEN IFNULL(ta2.remaining_tuition, 0)
					WHEN IFNULL(ta2.total_quantity, 0) > 0 THEN IFNULL(ta2.remaining_quantity, 0)
					ELSE 0
				END) `+oneToOneReadingListBucketFrom+`
			), CASE
				WHEN IFNULL(icq_ded.lesson_model, 0) = 3 THEN IFNULL(ta_ded.remaining_tuition, 0)
				WHEN IFNULL(ta_ded.total_quantity, 0) > 0 THEN IFNULL(ta_ded.remaining_quantity, 0)
				ELSE 0
			END),
			IFNULL((
				SELECT SUM(CASE
					WHEN IFNULL(icq2.lesson_model, 0) = 3 THEN IFNULL(ta2.free_quantity, 0)
					WHEN IFNULL(ta2.total_quantity, 0) = 0 AND IFNULL(ta2.free_quantity, 0) > 0 THEN IFNULL(ta2.remaining_quantity, 0)
					ELSE 0
				END) `+oneToOneReadingListBucketFrom+`
			), CASE
				WHEN IFNULL(icq_ded.lesson_model, 0) = 3 THEN IFNULL(ta_ded.free_quantity, 0)
				WHEN IFNULL(ta_ded.total_quantity, 0) = 0 AND IFNULL(ta_ded.free_quantity, 0) > 0 THEN IFNULL(ta_ded.remaining_quantity, 0)
				ELSE 0
			END),
			CASE
				WHEN IFNULL(icq_ded.lesson_model, 0) > 0 THEN icq_ded.lesson_model
				WHEN IFNULL(ta_ded.enable_expire_time, 0) = 1 AND IFNULL(ta_ded.total_quantity, 0) > 0 THEN 2
				ELSE 0
			END,
			IFNULL(ta_ded.status, 0),
			IFNULL(ta_ded.enable_expire_time, 0),
			ta_ded.expire_time,
			ta_ded.status_change_time,
			ta_ded.suspended_time,
			ta_ded.class_ending_time,
			IFNULL(ta_ded.assigned_class, 0)
		FROM teaching_class tc
		INNER JOIN (
			SELECT MIN(id) AS id
			FROM teaching_class_student
			WHERE teaching_class_id = ? AND inst_id = ? AND del_flag = 0
		) tcs_pick ON 1 = 1
		INNER JOIN teaching_class_student tcs ON tcs.id = tcs_pick.id AND tcs.teaching_class_id = tc.id AND tcs.del_flag = 0
		INNER JOIN inst_student s ON s.id = tcs.student_id AND s.del_flag = 0
		INNER JOIN inst_course c ON c.id = tc.course_id AND c.del_flag = 0
		`+oneToOneTuitionAccountDeductionJoinSQL+`
		LEFT JOIN inst_course_quotation icq ON icq.id = tcs.quote_id AND icq.del_flag = 0
		LEFT JOIN inst_user advisor ON advisor.id = tc.advisor_id
		LEFT JOIN inst_user default_teacher ON default_teacher.id = tc.default_teacher_id
		LEFT JOIN inst_user created_staff ON created_staff.id = tc.create_id
		LEFT JOIN teaching_class_teacher default_teacher_rel
			ON default_teacher_rel.teaching_class_id = tc.id
			AND default_teacher_rel.inst_id = tc.inst_id
			AND default_teacher_rel.teacher_id = tc.default_teacher_id
			AND default_teacher_rel.del_flag = 0
		WHERE tc.inst_id = ? AND tc.id = ? AND tc.class_type = ? AND tc.del_flag = 0
		LIMIT 1
	`, classID, instID, instID, classID, model.TeachingClassTypeOneToOne)

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
		deductLessonName    string
		deductCourseID      int64
		deductTeachMethod   int
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
		&deductLessonName,
		&deductCourseID,
		&deductTeachMethod,
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
		&detail.TuitionAccountCount,
		&detail.TuitionAccount.TotalTuition,
		&detail.TuitionAccount.RemainTuition,
		&detail.TuitionAccount.TotalQuantity,
		&detail.TuitionAccount.TotalFreeQuantity,
		&detail.TuitionAccount.RemainQuantity,
		&detail.TuitionAccount.RemainFreeQuantity,
		&detail.TuitionAccount.LessonChargingMode,
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
	detail.TuitionAccount.LessonID = strconv.FormatInt(deductCourseID, 10)
	if deductTeachMethod == model.TeachingClassTypeNormal || deductTeachMethod == model.TeachingClassTypeOneToOne {
		detail.TuitionAccount.LessonType = deductTeachMethod
	} else {
		detail.TuitionAccount.LessonType = model.TeachingClassTypeOneToOne
	}
	if strings.TrimSpace(deductLessonName) != "" {
		detail.TuitionAccount.ProductName = strings.TrimSpace(deductLessonName)
	} else {
		detail.TuitionAccount.ProductName = detail.LessonName
	}
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

func (repo *Repository) SwitchOneToOneDefaultTuitionAccount(ctx context.Context, instID, operatorID int64, dto model.OneToOneSwitchDefaultTuitionAccountDTO) error {
	classID, err := strconv.ParseInt(strings.TrimSpace(dto.ID), 10, 64)
	if err != nil || classID <= 0 {
		return errors.New("1对1ID无效")
	}
	taID, err := strconv.ParseInt(strings.TrimSpace(dto.TuitionAccountID), 10, 64)
	if err != nil || taID <= 0 {
		return errors.New("tuitionAccountId无效")
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var (
		exists    int
		studentID int64
	)
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*), IFNULL(MIN(tcs.student_id), 0)
		FROM teaching_class tc
		INNER JOIN teaching_class_student tcs ON tcs.teaching_class_id = tc.id AND tcs.inst_id = tc.inst_id AND tcs.del_flag = 0
		WHERE tc.id = ? AND tc.inst_id = ? AND tc.class_type = ? AND tc.del_flag = 0
	`, classID, instID, model.TeachingClassTypeOneToOne).Scan(&exists, &studentID); err != nil {
		return err
	}
	if exists == 0 || studentID <= 0 {
		return sql.ErrNoRows
	}

	selectedBind, err := repo.loadOneToOneDeductBindTx(ctx, tx, instID, studentID, taID)
	if err != nil {
		return err
	}
	selectedBucket, err := repo.loadOneToOneTuitionBucketTx(ctx, tx, instID, taID)
	if err != nil {
		return err
	}

	var (
		pickRowID     int64
		pickOrderID   int64
		pickOCDID     int64
		pickQuoteID   int64
		pickPrimaryTA int64
	)
	if err := tx.QueryRowContext(ctx, `
		SELECT id, IFNULL(order_id, 0), IFNULL(order_course_detail_id, 0), IFNULL(quote_id, 0), IFNULL(primary_tuition_account_id, 0)
		FROM teaching_class_student
		WHERE inst_id = ? AND teaching_class_id = ? AND del_flag = 0
		ORDER BY id ASC
		LIMIT 1
	`, instID, classID).Scan(&pickRowID, &pickOrderID, &pickOCDID, &pickQuoteID, &pickPrimaryTA); err != nil {
		return err
	}
	if pickPrimaryTA == taID {
		return nil
	}

	var (
		targetRowID     int64
		targetOrderID   int64
		targetOCDID     int64
		targetQuoteID   int64
		targetPrimaryTA int64
	)
	switch err := tx.QueryRowContext(ctx, `
		SELECT
			tcs.id,
			IFNULL(tcs.order_id, 0),
			IFNULL(tcs.order_course_detail_id, 0),
			IFNULL(tcs.quote_id, 0),
			IFNULL(tcs.primary_tuition_account_id, 0)
		FROM teaching_class_student tcs
		LEFT JOIN tuition_account ta_eff ON ta_eff.id = COALESCE(
			NULLIF(tcs.primary_tuition_account_id, 0),
			(SELECT MIN(ta0.id) FROM tuition_account ta0
			 WHERE ta0.order_course_detail_id = tcs.order_course_detail_id
			   AND ta0.inst_id = tcs.inst_id AND ta0.del_flag = 0)
		) AND ta_eff.inst_id = tcs.inst_id AND ta_eff.del_flag = 0
		INNER JOIN inst_course ic_eff ON ic_eff.id = ta_eff.course_id AND ic_eff.del_flag = 0
		LEFT JOIN sale_order_course_detail sod_eff ON sod_eff.id = ta_eff.order_course_detail_id AND sod_eff.del_flag = 0
		LEFT JOIN inst_course_quotation icq_eff ON icq_eff.id = COALESCE(
			NULLIF(ta_eff.quote_id, 0),
			NULLIF(sod_eff.quote_id, 0),
			(SELECT qx.id FROM inst_course_quotation qx
			 WHERE qx.course_id = ta_eff.course_id AND qx.del_flag = 0
			   AND ABS(IFNULL(qx.quantity, 0) - IFNULL(ta_eff.total_quantity, 0)) < 0.000001
			   AND ABS(IFNULL(qx.price, 0) - IFNULL(ta_eff.total_tuition, 0)) < 0.000001
			 ORDER BY qx.id DESC LIMIT 1),
			(SELECT qmin.id FROM inst_course_quotation qmin
			 WHERE qmin.course_id = ta_eff.course_id AND qmin.del_flag = 0
			 ORDER BY qmin.id ASC LIMIT 1)
		) AND icq_eff.del_flag = 0
		WHERE tcs.inst_id = ? AND tcs.teaching_class_id = ? AND tcs.del_flag = 0
			AND ta_eff.id IS NOT NULL
			AND ta_eff.course_id = ?
			AND IFNULL(ic_eff.teach_method, 0) = ?
			AND IFNULL(icq_eff.lesson_model, -99999) = ?
		ORDER BY tcs.id ASC
		LIMIT 1
	`, instID, classID, selectedBucket.courseID, selectedBucket.teachMethod, selectedBucket.lessonModelCode).Scan(&targetRowID, &targetOrderID, &targetOCDID, &targetQuoteID, &targetPrimaryTA); {
	case errors.Is(err, sql.ErrNoRows):
		targetRowID = 0
	case err != nil:
		return err
	}

	now := time.Now()
	if targetRowID == pickRowID {
		return nil
	}
	if targetRowID > 0 && targetRowID != pickRowID {
		tempOCDID := -pickRowID
		if tempOCDID == pickOCDID || tempOCDID == targetOCDID {
			tempOCDID = -(pickRowID + targetRowID + 1)
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE teaching_class_student
			SET order_course_detail_id = ?, update_id = ?, update_time = ?
			WHERE id = ? AND inst_id = ? AND del_flag = 0
		`, tempOCDID, operatorID, now, pickRowID, instID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE teaching_class_student
			SET order_id = ?, order_course_detail_id = ?, quote_id = ?, primary_tuition_account_id = ?, update_id = ?, update_time = ?
			WHERE id = ? AND inst_id = ? AND del_flag = 0
		`, pickOrderID, pickOCDID, pickQuoteID, pickPrimaryTA, operatorID, now, targetRowID, instID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE teaching_class_student
			SET order_id = ?, order_course_detail_id = ?, quote_id = ?, primary_tuition_account_id = ?, update_id = ?, update_time = ?
			WHERE id = ? AND inst_id = ? AND del_flag = 0
		`, targetOrderID, targetOCDID, targetQuoteID, targetPrimaryTA, operatorID, now, pickRowID, instID); err != nil {
			return err
		}
	} else {
		if _, err := tx.ExecContext(ctx, `
			UPDATE teaching_class_student
			SET order_id = ?, order_course_detail_id = ?, quote_id = ?, primary_tuition_account_id = ?, update_id = ?, update_time = ?
			WHERE id = ? AND inst_id = ? AND del_flag = 0
		`, selectedBind.orderID, selectedBind.ocdID, selectedBind.quoteID, selectedBind.taID, operatorID, now, pickRowID, instID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

type oneToOneDeductBind struct {
	taID    int64
	orderID int64
	ocdID   int64
	quoteID int64
}

type oneToOneTuitionBucket struct {
	courseID        int64
	teachMethod     int
	lessonModelCode int
}

type oneToOneDeductionAggregateKey struct {
	courseID            int64
	teachMethod         int
	lessonChargingMode  int
	hasBucketConstraint bool
}

func (repo *Repository) loadOneToOneDeductBindTx(ctx context.Context, tx *sql.Tx, instID, studentID, taID int64) (oneToOneDeductBind, error) {
	var orderID, ocdID, quoteID sql.NullInt64
	var taStudentID int64
	var taStatus int
	var deductCourseTeachMethod int
	err := tx.QueryRowContext(ctx, `
		SELECT IFNULL(ta.order_id, 0), IFNULL(ta.order_course_detail_id, 0), IFNULL(ta.quote_id, 0),
			ta.student_id, IFNULL(ta.status, 0),
			IFNULL(ic.teach_method, 0)
		FROM tuition_account ta
		INNER JOIN inst_course ic ON ic.id = ta.course_id AND ic.del_flag = 0
		WHERE ta.id = ? AND ta.inst_id = ? AND ta.del_flag = 0
		LIMIT 1
	`, taID, instID).Scan(&orderID, &ocdID, &quoteID, &taStudentID, &taStatus, &deductCourseTeachMethod)
	if errors.Is(err, sql.ErrNoRows) {
		return oneToOneDeductBind{}, errors.New("学费账户不存在")
	}
	if err != nil {
		return oneToOneDeductBind{}, err
	}
	if taStudentID != studentID {
		return oneToOneDeductBind{}, errors.New("学费账户不属于所选学员")
	}
	if taStatus != 1 {
		return oneToOneDeductBind{}, errors.New("学费账户须为在读状态方可创建1对1")
	}
	if deductCourseTeachMethod != 1 && deductCourseTeachMethod != 2 {
		return oneToOneDeductBind{}, errors.New("扣费学费账户须为班级授课或1对1课程的报读账户")
	}
	var b oneToOneDeductBind
	b.taID = taID
	if orderID.Valid {
		b.orderID = orderID.Int64
	}
	if ocdID.Valid {
		b.ocdID = ocdID.Int64
	}
	if quoteID.Valid {
		b.quoteID = quoteID.Int64
	}
	return b, nil
}

func (repo *Repository) loadOneToOneTuitionBucketTx(ctx context.Context, tx *sql.Tx, instID, taID int64) (oneToOneTuitionBucket, error) {
	var bucket oneToOneTuitionBucket
	err := tx.QueryRowContext(ctx, `
		SELECT
			ta.course_id,
			IFNULL(ic.teach_method, 0),
			IFNULL(icq.lesson_model, -99999)
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
		WHERE ta.id = ? AND ta.inst_id = ? AND ta.del_flag = 0
		LIMIT 1
	`, taID, instID).Scan(&bucket.courseID, &bucket.teachMethod, &bucket.lessonModelCode)
	if err != nil {
		return oneToOneTuitionBucket{}, err
	}
	return bucket, nil
}

func parseOneToOneDeductionAggregateKey(raw string) (oneToOneDeductionAggregateKey, error) {
	key := strings.TrimSpace(strings.ToLower(raw))
	if !strings.HasPrefix(key, "agg:") {
		return oneToOneDeductionAggregateKey{}, errors.New("请选择扣费学费账户")
	}
	parts := strings.Split(strings.TrimSpace(raw[4:]), ":")
	if len(parts) <= 0 {
		return oneToOneDeductionAggregateKey{}, errors.New("请选择扣费学费账户")
	}
	courseID, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
	if err != nil || courseID <= 0 {
		return oneToOneDeductionAggregateKey{}, errors.New("请选择扣费学费账户")
	}
	keyDTO := oneToOneDeductionAggregateKey{courseID: courseID}
	if len(parts) >= 3 {
		teachMethod, terr := strconv.Atoi(strings.TrimSpace(parts[1]))
		lessonModel, merr := strconv.Atoi(strings.TrimSpace(parts[2]))
		if terr != nil || merr != nil {
			return oneToOneDeductionAggregateKey{}, errors.New("请选择扣费学费账户")
		}
		keyDTO.teachMethod = teachMethod
		keyDTO.lessonChargingMode = lessonModel
		keyDTO.hasBucketConstraint = true
	}
	return keyDTO, nil
}

func (repo *Repository) loadOrderCourseDetailTx(ctx context.Context, tx *sql.Tx, orderCourseDetailID int64) (orderCourseDetail, error) {
	var detail orderCourseDetail
	err := tx.QueryRowContext(ctx, `
		SELECT id, course_id, quote_id, handle_type, count, unit,
		       IFNULL(free_quantity, 0), IFNULL(amount, 0), IFNULL(real_quantity, 0),
		       IFNULL(has_valid_date, 0), valid_date, end_date
		FROM sale_order_course_detail
		WHERE id = ? AND del_flag = 0
		LIMIT 1
	`, orderCourseDetailID).Scan(
		&detail.ID,
		&detail.CourseID,
		&detail.QuoteID,
		&detail.HandleType,
		&detail.Count,
		&detail.Unit,
		&detail.FreeQuantity,
		&detail.Amount,
		&detail.RealQuantity,
		&detail.HasValidDate,
		&detail.ValidDate,
		&detail.EndDate,
	)
	if err != nil {
		return orderCourseDetail{}, err
	}
	return detail, nil
}

func (repo *Repository) syncOneToOneTimeSlotAutoConsumeTx(ctx context.Context, tx *sql.Tx, instID, operatorID, teachingCourseID, studentID int64, binds []oneToOneDeductBind, now time.Time) error {
	if len(binds) == 0 {
		return nil
	}

	orderByDetail := make(map[int64]int64, len(binds))
	detailOrder := make([]int64, 0, len(binds))
	for _, bind := range binds {
		if bind.orderID <= 0 || bind.ocdID <= 0 {
			continue
		}
		if _, exists := orderByDetail[bind.ocdID]; exists {
			continue
		}
		orderByDetail[bind.ocdID] = bind.orderID
		detailOrder = append(detailOrder, bind.ocdID)
	}
	if len(detailOrder) == 0 {
		return nil
	}

	details := make([]orderCourseDetail, 0, len(detailOrder))
	for _, detailID := range detailOrder {
		detail, err := repo.loadOrderCourseDetailTx(ctx, tx, detailID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return err
		}
		details = append(details, detail)
	}
	if len(details) == 0 {
		return nil
	}

	quotationMap, err := repo.getCourseQuotationsByIDsTx(ctx, tx, collectOrderDetailQuoteIDs(details))
	if err != nil {
		return err
	}

	for _, detail := range details {
		quotation, ok := quotationMap[detail.QuoteID.Int64]
		if !ok {
			continue
		}
		orderID := orderByDetail[detail.ID]
		if orderID <= 0 {
			continue
		}
		if err := repo.consumeOneTimeSlotDayForOneToOneCreateTx(ctx, tx, instID, operatorID, teachingCourseID, orderID, studentID, detail, quotation, now); err != nil {
			return err
		}
	}
	return nil
}

func oneToOneTimeSlotAutoConsumeSourceID(teachingCourseID int64, consumeDate time.Time) int64 {
	coursePart := teachingCourseID
	if coursePart <= 0 {
		coursePart = 1
	}
	datePart := timeSlotSourceIDByDate(startOfDayTime(consumeDate))
	return -((coursePart * 100000000) + datePart)
}

func (repo *Repository) consumeOneTimeSlotDayForOneToOneCreateTx(
	ctx context.Context,
	tx *sql.Tx,
	instID, operatorID, teachingCourseID, orderID, studentID int64,
	detail orderCourseDetail,
	quotation model.CourseQuotation,
	now time.Time,
) error {
	if quotation.LessonModel == nil || *quotation.LessonModel != 2 {
		return nil
	}
	if !detail.ValidDate.Valid || !detail.EndDate.Valid {
		return nil
	}

	// 仅按「上课课程 + 同一天」去重：
	// 同一上课课程同一天不重复课消；不同上课课程仍允许各自正常课消。
	if teachingCourseID == detail.CourseID {
		return repo.initializeTimeSlotIncomeTx(ctx, tx, instID, operatorID, orderID, studentID, detail, quotation, now)
	}

	startDate := startOfDayTime(detail.ValidDate.Time)
	if startOfDayTime(now).Before(startDate) {
		return nil
	}

	var (
		paidAccountID           int64
		paidCourseID            int64
		paidTotalDays           float64
		paidTotalTuition        float64
		usedQuantity            float64
		usedTuition             float64
		confirmedTuitionCurrent float64
		orderNumber             string
		lessonType              sql.NullInt64
	)
	err := tx.QueryRowContext(ctx, `
		SELECT ta.id, ta.course_id, IFNULL(ta.total_quantity, 0), IFNULL(ta.total_tuition, 0),
		       IFNULL(ta.used_quantity, 0), IFNULL(ta.used_tuition, 0), IFNULL(ta.confirmed_tuition, 0),
		       IFNULL(so.order_number, ''), c.teach_method
		FROM tuition_account ta
		LEFT JOIN sale_order so ON so.id = ta.order_id AND so.del_flag = 0
		LEFT JOIN inst_course c ON c.id = ta.course_id AND c.del_flag = 0
		WHERE ta.inst_id = ? AND ta.order_id = ? AND ta.order_course_detail_id = ? AND ta.del_flag = 0 AND IFNULL(ta.total_tuition, 0) > 0
		ORDER BY ta.id ASC
		LIMIT 1
	`, instID, orderID, detail.ID).Scan(&paidAccountID, &paidCourseID, &paidTotalDays, &paidTotalTuition, &usedQuantity, &usedTuition, &confirmedTuitionCurrent, &orderNumber, &lessonType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	consumedByOther, err := repo.hasOtherTeachingCourseAutoConsumeOnDateTx(ctx, tx, instID, studentID, teachingCourseID, paidAccountID, now)
	if err != nil {
		return err
	}
	if consumedByOther {
		return nil
	}

	sourceID := oneToOneTimeSlotAutoConsumeSourceID(teachingCourseID, now)

	var paidFlowExists int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM tuition_account_flow
		WHERE inst_id = ? AND tuition_account_id = ? AND source_type = ? AND source_id = ? AND del_flag = 0
	`, instID, paidAccountID, model.TuitionAccountFlowSourceAutoConsume, sourceID).Scan(&paidFlowExists); err != nil {
		return err
	}
	if paidFlowExists > 0 {
		return nil
	}
	targetPaidDays := math.Min(usedQuantity+1, paidTotalDays)
	paidDelta := roundMoney(targetPaidDays - usedQuantity)
	if paidDelta > 0.00001 {
		prevConfirmed := confirmedTuitionCurrent
		if prevConfirmed <= 0 {
			prevConfirmed = usedTuition
		}
		targetConfirmed := cumulativeTimeSlotTuition(paidTotalTuition, paidTotalDays, int(math.Round(targetPaidDays)))
		targetRemainDays := math.Max(paidTotalDays-targetPaidDays, 0)
		targetRemainTuition := roundMoney(math.Max(paidTotalTuition-targetConfirmed, 0))
		if _, err := tx.ExecContext(ctx, `
			UPDATE tuition_account
			SET used_quantity = ?, remaining_quantity = ?, used_tuition = ?, remaining_tuition = ?, confirmed_tuition = ?, update_id = ?, update_time = NOW()
			WHERE id = ? AND inst_id = ? AND del_flag = 0
		`, targetPaidDays, targetRemainDays, targetConfirmed, targetRemainTuition, targetConfirmed, operatorID, paidAccountID, instID); err != nil {
			return err
		}

		var lessonTypeValue any
		if lessonType.Valid {
			lessonTypeValue = lessonType.Int64
		}
		rowTuition := roundMoney(targetConfirmed - prevConfirmed)
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO tuition_account_flow (
				uuid, version, inst_id, tuition_account_id, student_id, product_id, lesson_type, lesson_charging_mode,
				source_type, source_id, teaching_record_id, order_number, created_time, quantity, tuition, balance_quantity, balance_tuition,
				create_id, create_time, update_id, update_time, del_flag
			) VALUES (
				UUID(), 0, ?, ?, ?, ?, ?, ?,
				?, ?, NULL, ?, ?, ?, ?, ?, ?,
				?, NOW(), ?, NOW(), 0
			)
		`,
			instID,
			paidAccountID,
			studentID,
			paidCourseID,
			lessonTypeValue,
			2,
			model.TuitionAccountFlowSourceAutoConsume,
			sourceID,
			orderNumber,
			now,
			paidDelta,
			rowTuition,
			targetRemainDays,
			targetRemainTuition,
			operatorID,
			operatorID,
		); err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				return err
			}
		}
		return nil
	}

	if detail.FreeQuantity <= 0 {
		return nil
	}
	var (
		freeAccountID       int64
		freeCourseID        int64
		freeUsedQuantity    float64
		freeTotalFreeAmount float64
	)
	err = tx.QueryRowContext(ctx, `
		SELECT ta.id, ta.course_id, IFNULL(ta.used_quantity, 0), IFNULL(ta.free_quantity, 0)
		FROM tuition_account ta
		WHERE ta.inst_id = ? AND ta.order_id = ? AND ta.order_course_detail_id = ? AND ta.del_flag = 0
		  AND IFNULL(ta.total_tuition, 0) = 0 AND IFNULL(ta.free_quantity, 0) > 0
		ORDER BY ta.id ASC
		LIMIT 1
	`, instID, orderID, detail.ID).Scan(&freeAccountID, &freeCourseID, &freeUsedQuantity, &freeTotalFreeAmount)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	var freeFlowExists int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM tuition_account_flow
		WHERE inst_id = ? AND tuition_account_id = ? AND source_type = ? AND source_id = ? AND del_flag = 0
	`, instID, freeAccountID, model.TuitionAccountFlowSourceAutoConsume, sourceID).Scan(&freeFlowExists); err != nil {
		return err
	}
	if freeFlowExists > 0 {
		return nil
	}
	targetFreeDays := math.Min(freeUsedQuantity+1, freeTotalFreeAmount)
	freeDelta := roundMoney(targetFreeDays - freeUsedQuantity)
	if freeDelta <= 0.00001 {
		return nil
	}
	targetFreeRemain := math.Max(freeTotalFreeAmount-targetFreeDays, 0)
	if _, err := tx.ExecContext(ctx, `
		UPDATE tuition_account
		SET used_quantity = ?, remaining_quantity = ?, update_id = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, targetFreeDays, targetFreeRemain, operatorID, freeAccountID, instID); err != nil {
		return err
	}
	var lessonTypeValue any
	if lessonType.Valid {
		lessonTypeValue = lessonType.Int64
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO tuition_account_flow (
			uuid, version, inst_id, tuition_account_id, student_id, product_id, lesson_type, lesson_charging_mode,
			source_type, source_id, teaching_record_id, order_number, created_time, quantity, tuition, balance_quantity, balance_tuition,
			create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, ?, ?, ?, ?, ?,
			?, ?, NULL, ?, ?, ?, ?, ?, ?,
			?, NOW(), ?, NOW(), 0
		)
	`,
		instID,
		freeAccountID,
		studentID,
		freeCourseID,
		lessonTypeValue,
		2,
		model.TuitionAccountFlowSourceAutoConsume,
		sourceID,
		orderNumber,
		now,
		freeDelta,
		0,
		targetFreeRemain,
		0,
		operatorID,
		operatorID,
	); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return err
		}
	}
	return nil
}

// CreateOneToOne 手动创建 1 对 1：tuitionAccountId 为数字 id 或 agg:{扣费课程id}:{授课方式}:{计费模式}（按计费桶汇总时多笔账户各写一条班员，列表课时与下拉一致）
func (repo *Repository) CreateOneToOne(ctx context.Context, instID, operatorID int64, dto model.OneToOneCreateDTO) (int64, error) {
	studentID, err := strconv.ParseInt(strings.TrimSpace(dto.StudentID), 10, 64)
	if err != nil || studentID <= 0 {
		return 0, errors.New("学员ID不能为空")
	}
	courseID, err := strconv.ParseInt(strings.TrimSpace(dto.LessonID), 10, 64)
	if err != nil || courseID <= 0 {
		return 0, errors.New("课程ID不能为空")
	}
	tuitionKey := strings.TrimSpace(dto.TuitionAccountID)
	if tuitionKey == "" {
		return 0, errors.New("请选择扣费学费账户")
	}
	name := strings.TrimSpace(dto.Name)
	if name == "" {
		return 0, errors.New("1对1名称不能为空")
	}
	defaultTeacherID, _ := strconv.ParseInt(strings.TrimSpace(dto.DefaultTeacherID), 10, 64)
	recordMode := dto.DefaultClassTimeRecordMode
	if recordMode <= 0 {
		recordMode = 1
	}
	stuClassTime := dto.DefaultStudentClassTime
	if stuClassTime <= 0 {
		stuClassTime = 1
	}
	teacherClassTime := dto.DefaultTeacherClassTime
	if teacherClassTime < 0 {
		teacherClassTime = 0
	}

	classProperties := dto.ClassProperties
	if classProperties == nil {
		classProperties = []model.OneToOnePropertyVO{}
	}
	classPropertiesJSON, err := json.Marshal(classProperties)
	if err != nil {
		return 0, err
	}

	exists, err := repo.ExistsActiveOneToOneForStudentCourse(ctx, instID, studentID, courseID)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, errors.New("该学员在此课程下已有开班中的1对1")
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var teachMethod int
	err = tx.QueryRowContext(ctx, `
		SELECT IFNULL(teach_method, 0)
		FROM inst_course
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, courseID, instID).Scan(&teachMethod)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, errors.New("课程不存在")
	}
	if err != nil {
		return 0, err
	}
	if teachMethod != 2 {
		return 0, errors.New("所选课程不是1对1授课，无法创建1对1")
	}

	var binds []oneToOneDeductBind
	if strings.HasPrefix(strings.ToLower(tuitionKey), "agg:") {
		aggKey, perr := parseOneToOneDeductionAggregateKey(tuitionKey)
		if perr != nil {
			return 0, perr
		}
		var (
			taIDs []int64
			qerr  error
		)
		if aggKey.hasBucketConstraint {
			taIDs, qerr = repo.ListTuitionAccountIDsForStudentCourseBucket(ctx, tx, instID, studentID, aggKey.courseID, aggKey.teachMethod, aggKey.lessonChargingMode)
		} else {
			taIDs, qerr = repo.ListTuitionAccountIDsForStudentCourse(ctx, tx, instID, studentID, aggKey.courseID)
		}
		if qerr != nil {
			return 0, qerr
		}
		if len(taIDs) == 0 {
			return 0, errors.New("所选扣费账户下暂无在读学费账户")
		}
		for _, tid := range taIDs {
			b, verr := repo.loadOneToOneDeductBindTx(ctx, tx, instID, studentID, tid)
			if verr != nil {
				return 0, verr
			}
			binds = append(binds, b)
		}
	} else {
		taID, perr := strconv.ParseInt(tuitionKey, 10, 64)
		if perr != nil || taID <= 0 {
			return 0, errors.New("请选择扣费学费账户")
		}
		b, verr := repo.loadOneToOneDeductBindTx(ctx, tx, instID, studentID, taID)
		if verr != nil {
			return 0, verr
		}
		binds = append(binds, b)
	}

	// 同一报名明细下可能存在「付费账户 + 赠送账户」两条 tuition_account。
	// 创建 1 对 1 时 teaching_class_student 仍应只保留一条绑定，沿用主账户（通常为较早创建的付费账户）。
	if len(binds) > 1 {
		deduped := make([]oneToOneDeductBind, 0, len(binds))
		seen := make(map[string]struct{}, len(binds))
		for _, bind := range binds {
			key := strconv.FormatInt(bind.ocdID, 10)
			if bind.ocdID <= 0 {
				key = "ta:" + strconv.FormatInt(bind.taID, 10)
			}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			deduped = append(deduped, bind)
		}
		binds = deduped
	}

	teacherIDs := normalizeTeacherIDs(dto.TeacherID, defaultTeacherID)
	advisorID := int64(0)
	if len(teacherIDs) > 0 {
		advisorID = teacherIDs[0]
	}

	now := time.Now()
	res, err := tx.ExecContext(ctx, `
		INSERT INTO teaching_class (
			uuid, version, inst_id, class_type, course_id, name, advisor_id, default_teacher_id, status,
			scheduled_lesson_count, finished_lesson_count, class_room_id, class_room_name, classroom_enabled, remark,
			create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, ?, ?, ?, ?, ?, ?, 0, 0, 0, '', NULL, ?, ?, NOW(), ?, NOW(), 0
		)
	`, instID, model.TeachingClassTypeOneToOne, courseID, name, advisorID, defaultTeacherID, model.TeachingClassStatusActive,
		strings.TrimSpace(dto.Remark), operatorID, operatorID)
	if err != nil {
		return 0, err
	}
	classID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, b := range binds {
		if _, err = tx.ExecContext(ctx, `
		INSERT INTO teaching_class_student (
			uuid, version, inst_id, teaching_class_id, student_id, order_id, order_course_detail_id, quote_id,
			primary_tuition_account_id, class_student_status, class_time, student_class_time, teacher_class_time,
			class_time_record_mode, last_finished_lesson_day, class_properties_json,
			create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?, ?, ?, NULL, ?, ?, NOW(), ?, NOW(), 0
		)
	`, instID, classID, studentID, b.orderID, b.ocdID, b.quoteID, b.taID,
			stuClassTime, stuClassTime, teacherClassTime, recordMode,
			string(classPropertiesJSON),
			operatorID, operatorID); err != nil {
			return 0, err
		}
	}

	if err := repo.syncOneToOneTimeSlotAutoConsumeTx(ctx, tx, instID, operatorID, courseID, studentID, binds, now); err != nil {
		return 0, err
	}

	for _, tid := range teacherIDs {
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
			instID, classID, tid, boolToTinyInt(tid == defaultTeacherID), operatorID, now, operatorID, now,
		); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return classID, nil
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
		whereParts = append(whereParts, `EXISTS (
			SELECT 1 FROM teaching_class_student tcs_w
			WHERE tcs_w.teaching_class_id = tc.id AND tcs_w.inst_id = tc.inst_id AND tcs_w.del_flag = 0
				AND CAST(tcs_w.student_id AS CHAR) = ?
		)`)
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
		whereParts = append(whereParts, `EXISTS (
			SELECT 1 FROM teaching_class_student tcs_w
			WHERE tcs_w.teaching_class_id = tc.id AND tcs_w.inst_id = tc.inst_id AND tcs_w.del_flag = 0
				AND tcs_w.class_student_status IN (`+strings.Join(placeholders, ",")+`)
		)`)
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
