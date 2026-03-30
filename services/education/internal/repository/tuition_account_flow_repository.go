package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

// tuitionAccountQuotationJoin matches ListStudentTuitionAccountsByStudentAndLesson: resolve quote when ta.quote_id is empty.
const tuitionAccountQuotationJoinSQL = `
LEFT JOIN sale_order_course_detail sod_taf ON sod_taf.id = ta.order_course_detail_id AND sod_taf.del_flag = 0
LEFT JOIN inst_course_quotation icq_taf ON icq_taf.id = COALESCE(
	NULLIF(ta.quote_id, 0),
	NULLIF(sod_taf.quote_id, 0),
	(SELECT qx.id FROM inst_course_quotation qx
	 WHERE qx.course_id = ta.course_id AND qx.del_flag = 0
	   AND ABS(IFNULL(qx.quantity, 0) - IFNULL(ta.total_quantity, 0)) < 0.000001
	   AND ABS(IFNULL(qx.price, 0) - IFNULL(ta.total_tuition, 0)) < 0.000001
	 ORDER BY qx.id DESC LIMIT 1),
	(SELECT qmin.id FROM inst_course_quotation qmin
	 WHERE qmin.course_id = ta.course_id AND qmin.del_flag = 0
	 ORDER BY qmin.id ASC LIMIT 1)
) AND icq_taf.del_flag = 0`

// resolvedLessonChargingModeExpr:
// 1. Prefer stored flow mode.
// 2. Fallback to quotation lesson_model.
// 3. Infer mode from tuition_account for legacy rows that have no quotation bound.
func resolvedLessonChargingModeExpr(storedModeCol, quotationModelCol string) string {
	return `CASE
		WHEN IFNULL(` + storedModeCol + `, 0) > 0 THEN ` + storedModeCol + `
		WHEN IFNULL(` + quotationModelCol + `, 0) > 0 THEN ` + quotationModelCol + `
		WHEN IFNULL(ta.enable_expire_time, 0) = 1 AND IFNULL(ta.total_quantity, 0) > 0 THEN 2
		WHEN IFNULL(ta.total_quantity, 0) > 0 THEN 1
		WHEN IFNULL(ta.total_tuition, 0) > 0 THEN 3
		ELSE 0
	END`
}

func ensureTuitionAccountFlowTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS tuition_account_flow (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			uuid VARCHAR(64) NULL,
			version BIGINT NOT NULL DEFAULT 0,
			inst_id BIGINT NOT NULL,
			tuition_account_id BIGINT NOT NULL,
			student_id BIGINT NOT NULL,
			product_id BIGINT NOT NULL,
			lesson_type INT NULL,
			lesson_charging_mode INT NULL,
			source_type INT NOT NULL,
			source_id BIGINT NOT NULL DEFAULT 0,
			teaching_record_id BIGINT NULL,
			order_number VARCHAR(64) NOT NULL DEFAULT '',
			created_time DATETIME NOT NULL,
			quantity DECIMAL(18,2) NOT NULL DEFAULT 0,
			tuition DECIMAL(18,2) NOT NULL DEFAULT 0,
			balance_quantity DECIMAL(18,2) NOT NULL DEFAULT 0,
			balance_tuition DECIMAL(18,2) NOT NULL DEFAULT 0,
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT NOT NULL DEFAULT 0,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uk_tuition_account_flow_backfill (inst_id, tuition_account_id, source_type, source_id),
			KEY idx_tuition_account_flow_list (inst_id, created_time, id),
			KEY idx_tuition_account_flow_product (inst_id, product_id),
			KEY idx_tuition_account_flow_student (inst_id, student_id),
			KEY idx_tuition_account_flow_source (inst_id, source_type)
		)
	`)
	if err != nil {
		return err
	}
	for _, statement := range []string{
		"ALTER TABLE tuition_account_flow ADD COLUMN order_number VARCHAR(64) NOT NULL DEFAULT '' AFTER teaching_record_id",
		"ALTER TABLE tuition_account_flow ADD COLUMN balance_quantity DECIMAL(18,2) NOT NULL DEFAULT 0 AFTER tuition",
		"ALTER TABLE tuition_account_flow ADD COLUMN balance_tuition DECIMAL(18,2) NOT NULL DEFAULT 0 AFTER balance_quantity",
	} {
		if _, alterErr := db.ExecContext(ctx, statement); alterErr != nil && !strings.Contains(strings.ToLower(alterErr.Error()), "duplicate column") {
			return alterErr
		}
	}
	return nil
}

func (repo *Repository) ensureHistoricalTuitionAccountFlowRecords(ctx context.Context, instID int64) error {
	_, err := repo.db.ExecContext(ctx, `
		INSERT INTO tuition_account_flow (
			uuid, version, inst_id, tuition_account_id, student_id, product_id, lesson_type, lesson_charging_mode,
			source_type, source_id, teaching_record_id, order_number, created_time, quantity, tuition, balance_quantity, balance_tuition,
			create_id, create_time, update_id, update_time, del_flag
		)
		SELECT
			UUID(), 0, ta.inst_id, ta.id, ta.student_id, ta.course_id, ic.teach_method,
			CASE
				WHEN IFNULL(icq_taf.lesson_model, 0) > 0 THEN icq_taf.lesson_model
				WHEN IFNULL(ta.enable_expire_time, 0) = 1 AND IFNULL(ta.total_quantity, 0) > 0 THEN 2
				WHEN IFNULL(ta.total_quantity, 0) > 0 THEN 1
				WHEN IFNULL(ta.total_tuition, 0) > 0 THEN 3
				ELSE 0
			END,
			?, ta.order_course_detail_id, NULL, IFNULL(so.order_number, ''), ta.create_time,
			CASE
				WHEN IFNULL(icq_taf.lesson_model, 0) = 3 AND IFNULL(ta.total_tuition, 0) > 0 THEN IFNULL(ta.total_tuition, 0)
				WHEN IFNULL(icq_taf.lesson_model, 0) = 3 THEN IFNULL(ta.free_quantity, 0)
				WHEN IFNULL(ta.total_quantity, 0) > 0 THEN IFNULL(ta.total_quantity, 0)
				ELSE IFNULL(ta.free_quantity, 0)
			END,
			IFNULL(ta.total_tuition, 0),
			CASE
				WHEN IFNULL(icq_taf.lesson_model, 0) = 3 AND IFNULL(ta.total_tuition, 0) > 0 THEN IFNULL(ta.total_tuition, 0)
				WHEN IFNULL(icq_taf.lesson_model, 0) = 3 THEN IFNULL(ta.free_quantity, 0)
				WHEN IFNULL(ta.total_quantity, 0) > 0 THEN IFNULL(ta.total_quantity, 0)
				ELSE IFNULL(ta.free_quantity, 0)
			END,
			IFNULL(ta.total_tuition, 0),
			IFNULL(ta.create_id, 0), IFNULL(ta.create_time, NOW()),
			IFNULL(ta.update_id, 0), IFNULL(ta.update_time, NOW()), 0
		FROM tuition_account ta
		INNER JOIN inst_course ic ON ic.id = ta.course_id AND ic.del_flag = 0
		`+tuitionAccountQuotationJoinSQL+`
		LEFT JOIN sale_order so ON so.id = ta.order_id AND so.del_flag = 0
		WHERE ta.inst_id = ? AND ta.del_flag = 0
		ON DUPLICATE KEY UPDATE
			student_id = VALUES(student_id),
			product_id = VALUES(product_id),
			lesson_type = VALUES(lesson_type),
			lesson_charging_mode = VALUES(lesson_charging_mode),
			teaching_record_id = VALUES(teaching_record_id),
			order_number = VALUES(order_number),
			created_time = VALUES(created_time),
			quantity = VALUES(quantity),
			tuition = VALUES(tuition),
			balance_quantity = VALUES(balance_quantity),
			balance_tuition = VALUES(balance_tuition),
			update_id = VALUES(update_id),
			update_time = VALUES(update_time),
			del_flag = VALUES(del_flag)
	`, model.TuitionAccountFlowSourceRegistration, instID)
	return err
}

func (repo *Repository) GetTuitionAccountFlowRecordList(ctx context.Context, instID int64, query model.TuitionAccountFlowRecordListQueryDTO) (model.TuitionAccountFlowRecordListResult, error) {
	if err := repo.ensureHistoricalTuitionAccountFlowRecords(ctx, instID); err != nil {
		return model.TuitionAccountFlowRecordListResult{}, err
	}

	current := query.PageRequestModel.PageIndex
	size := query.PageRequestModel.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	whereParts := []string{
		"taf.inst_id = ?",
		"taf.del_flag = 0",
		"NOT (taf.source_type = 15 AND taf.source_id >= 20000101 AND taf.source_id > CAST(DATE_FORMAT(NOW(), '%Y%m%d') AS UNSIGNED))",
	}
	args := []any{instID}

	if strings.TrimSpace(query.QueryModel.ProductID) != "" {
		whereParts = append(whereParts, "CAST(taf.product_id AS CHAR) = ?")
		args = append(args, strings.TrimSpace(query.QueryModel.ProductID))
	}
	if strings.TrimSpace(query.QueryModel.StudentID) != "" {
		whereParts = append(whereParts, "CAST(taf.student_id AS CHAR) = ?")
		args = append(args, strings.TrimSpace(query.QueryModel.StudentID))
	}
	if strings.TrimSpace(query.QueryModel.OrderNumber) != "" {
		whereParts = append(whereParts, "taf.order_number LIKE ?")
		args = append(args, "%"+strings.TrimSpace(query.QueryModel.OrderNumber)+"%")
	}
	if len(query.QueryModel.SourceTypes) > 0 {
		holders := make([]string, 0, len(query.QueryModel.SourceTypes))
		for _, item := range query.QueryModel.SourceTypes {
			holders = append(holders, "?")
			args = append(args, item)
		}
		whereParts = append(whereParts, "taf.source_type IN ("+strings.Join(holders, ",")+")")
	}
	if begin := parseDateStart(query.QueryModel.StartTime); begin != nil {
		whereParts = append(whereParts, "taf.created_time >= ?")
		args = append(args, *begin)
	}
	if end := parseDateEnd(query.QueryModel.EndTime); end != nil {
		whereParts = append(whereParts, "taf.created_time <= ?")
		args = append(args, *end)
	}

	whereSQL := strings.Join(whereParts, " AND ")
	outerOrderBy := `
		DATE_FORMAT(g.min_created_time, '%Y-%m-%d %H:%i') DESC,
		CASE
			WHEN g.source_type IN (12,13,14,15,16,17,18,19,20,21,22,23,24,25) THEN 0
			WHEN g.source_type NOT IN (12,13,14,15,16,17,18,19,20,21,22,23,24,25) AND IFNULL(g.sum_tuition, 0) > 0 THEN 1
			WHEN g.source_type NOT IN (12,13,14,15,16,17,18,19,20,21,22,23,24,25) AND IFNULL(g.sum_quantity, 0) > 0 THEN 2
			ELSE 3
		END ASC,
		g.flow_id DESC`
	if query.SortModel.OrderByCreatedTime > 0 {
		outerOrderBy = `
			DATE_FORMAT(g.min_created_time, '%Y-%m-%d %H:%i') ASC,
			CASE
				WHEN g.source_type IN (12,13,14,15,16,17,18,19,20,21,22,23,24,25) THEN 0
				WHEN g.source_type NOT IN (12,13,14,15,16,17,18,19,20,21,22,23,24,25) AND IFNULL(g.sum_tuition, 0) > 0 THEN 1
				WHEN g.source_type NOT IN (12,13,14,15,16,17,18,19,20,21,22,23,24,25) AND IFNULL(g.sum_quantity, 0) > 0 THEN 2
				ELSE 3
			END ASC,
			g.flow_id ASC`
	}

	flowListInnerSQL := `
		SELECT
			MIN(taf.id) AS flow_id,
			MIN(taf.tuition_account_id) AS tuition_account_id,
			MIN(taf.student_id) AS student_id,
			IFNULL(s.stu_name, '') AS stu_name,
			CASE
				WHEN CHAR_LENGTH(IFNULL(s.mobile, '')) >= 7 THEN CONCAT(LEFT(s.mobile, 3), '****', RIGHT(s.mobile, 4))
				ELSE IFNULL(s.mobile, '')
			END AS stu_phone,
			IFNULL(s.avatar_url, '') AS stu_avatar,
			MIN(taf.product_id) AS product_id,
			IFNULL(c.name, '') AS course_name,
			MAX(taf.lesson_type) AS agg_lesson_type,
			MAX(taf.lesson_charging_mode) AS agg_lesson_mode,
			taf.source_type AS source_type,
			taf.source_id AS source_id,
			MAX(taf.teaching_record_id) AS agg_teaching_record_id,
			MIN(taf.created_time) AS min_created_time,
			IFNULL(SUM(
				CASE
					WHEN taf.source_type = 1
					 AND IFNULL(taf.lesson_charging_mode, 0) = 3
					 AND IFNULL(taf.quantity, 0) = 0
					 AND IFNULL(taf.tuition, 0) = 0
					 AND IFNULL(taf.balance_tuition, 0) = 0
					 AND IFNULL(taf.balance_quantity, 0) > 0
					THEN IFNULL(taf.balance_quantity, 0)
					ELSE IFNULL(taf.quantity, 0)
				END
			), 0) AS sum_quantity,
			IFNULL(SUM(taf.tuition), 0) AS sum_tuition
		FROM tuition_account_flow taf
		LEFT JOIN inst_student s ON s.id = taf.student_id AND s.del_flag = 0
		LEFT JOIN inst_course c ON c.id = taf.product_id AND c.del_flag = 0
		WHERE ` + whereSQL + `
		GROUP BY taf.source_type, taf.source_id, s.stu_name, s.mobile, s.avatar_url, c.name`

	flowListSQL := `
		SELECT
			g.flow_id,
			g.tuition_account_id,
			g.student_id,
			g.stu_name,
			g.stu_phone,
			g.stu_avatar,
			g.product_id,
			g.course_name,
			g.agg_lesson_type,
			` + resolvedLessonChargingModeExpr("g.agg_lesson_mode", "icq_taf.lesson_model") + ` AS resolved_lesson_mode,
			g.source_type,
			g.source_id,
			g.agg_teaching_record_id,
			g.min_created_time,
			g.sum_quantity,
			g.sum_tuition
		FROM (` + flowListInnerSQL + `) g
		LEFT JOIN tuition_account ta ON ta.id = g.tuition_account_id AND ta.inst_id = ? AND ta.del_flag = 0
		` + tuitionAccountQuotationJoinSQL + `
		ORDER BY ` + outerOrderBy + `
		LIMIT ? OFFSET ?`

	var total int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM (
			SELECT taf.source_type, taf.source_id
			FROM tuition_account_flow taf
			WHERE `+whereSQL+`
			GROUP BY taf.source_type, taf.source_id
		) flow_group
	`, args...).Scan(&total); err != nil {
		return model.TuitionAccountFlowRecordListResult{}, err
	}

	listArgs := append(append([]any{}, args...), instID, size, offset)
	rows, err := repo.db.QueryContext(ctx, flowListSQL, listArgs...)
	if err != nil {
		return model.TuitionAccountFlowRecordListResult{}, err
	}
	defer rows.Close()

	items := make([]model.TuitionAccountFlowRecordItem, 0, size)
	for rows.Next() {
		var (
			item               model.TuitionAccountFlowRecordItem
			flowID             int64
			tuitionAccountID   int64
			studentID          int64
			productID          int64
			sourceID           int64
			lessonType         sql.NullInt64
			lessonChargingMode sql.NullInt64
			teachingRecordID   sql.NullInt64
			createdTime        sql.NullTime
		)
		if err := rows.Scan(
			&flowID,
			&tuitionAccountID,
			&studentID,
			&item.StudentName,
			&item.StudentPhone,
			&item.StudentAvatar,
			&productID,
			&item.ProductName,
			&lessonType,
			&lessonChargingMode,
			&item.SourceType,
			&sourceID,
			&teachingRecordID,
			&createdTime,
			&item.Quantity,
			&item.Tuition,
		); err != nil {
			return model.TuitionAccountFlowRecordListResult{}, err
		}
		item.TuitionAccountFlowID = strconv.FormatInt(flowID, 10)
		item.TuitionAccountID = strconv.FormatInt(tuitionAccountID, 10)
		item.StudentID = strconv.FormatInt(studentID, 10)
		item.ProductID = strconv.FormatInt(productID, 10)
		item.SourceID = strconv.FormatInt(sourceID, 10)
		if lessonType.Valid {
			value := int(lessonType.Int64)
			item.LessonType = &value
		}
		if lessonChargingMode.Valid {
			value := int(lessonChargingMode.Int64)
			item.LessonChargingMode = &value
		}
		if teachingRecordID.Valid {
			item.TeachingRecordID = strconv.FormatInt(teachingRecordID.Int64, 10)
		}
		if createdTime.Valid {
			t := createdTime.Time
			item.CreatedTime = &t
		}
		items = append(items, item)
	}

	return model.TuitionAccountFlowRecordListResult{
		List:  items,
		Total: total,
	}, rows.Err()
}

func (repo *Repository) GetSubTuitionAccountFlowRecordList(ctx context.Context, instID int64, query model.SubTuitionAccountFlowRecordListQueryDTO) (model.SubTuitionAccountFlowRecordListResult, error) {
	if err := repo.ensureHistoricalTuitionAccountFlowRecords(ctx, instID); err != nil {
		return model.SubTuitionAccountFlowRecordListResult{}, err
	}

	current := query.PageRequestModel.PageIndex
	size := query.PageRequestModel.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	whereParts := []string{
		"taf.inst_id = ?",
		"taf.del_flag = 0",
		"NOT (taf.source_type = 15 AND taf.source_id >= 20000101 AND taf.source_id > CAST(DATE_FORMAT(NOW(), '%Y%m%d') AS UNSIGNED))",
	}
	args := []any{instID}

	if strings.TrimSpace(query.QueryModel.ProductID) != "" {
		whereParts = append(whereParts, "CAST(taf.product_id AS CHAR) = ?")
		args = append(args, strings.TrimSpace(query.QueryModel.ProductID))
	}
	if strings.TrimSpace(query.QueryModel.StudentID) != "" {
		whereParts = append(whereParts, "CAST(taf.student_id AS CHAR) = ?")
		args = append(args, strings.TrimSpace(query.QueryModel.StudentID))
	}
	if strings.TrimSpace(query.QueryModel.OrderNumber) != "" {
		whereParts = append(whereParts, "taf.order_number LIKE ?")
		args = append(args, "%"+strings.TrimSpace(query.QueryModel.OrderNumber)+"%")
	}
	if len(query.QueryModel.SourceTypes) > 0 {
		holders := make([]string, 0, len(query.QueryModel.SourceTypes))
		for _, item := range query.QueryModel.SourceTypes {
			holders = append(holders, "?")
			args = append(args, item)
		}
		whereParts = append(whereParts, "taf.source_type IN ("+strings.Join(holders, ",")+")")
	}
	if begin := parseDateStart(query.QueryModel.StartTime); begin != nil {
		whereParts = append(whereParts, "taf.created_time >= ?")
		args = append(args, *begin)
	}
	if end := parseDateEnd(query.QueryModel.EndTime); end != nil {
		whereParts = append(whereParts, "taf.created_time <= ?")
		args = append(args, *end)
	}

	whereSQL := strings.Join(whereParts, " AND ")
	orderBy := `
		DATE_FORMAT(taf.created_time, '%Y-%m-%d %H:%i') DESC,
		CASE
			WHEN taf.source_type NOT IN (12,13,14,15,16,17,18,19,20,21,22,23,24,25) AND IFNULL(taf.quantity, 0) > 0 AND IFNULL(taf.tuition, 0) = 0 THEN 0
			WHEN taf.source_type IN (12,13,14,15,16,17,18,19,20,21,22,23,24,25) THEN 1
			WHEN taf.source_type NOT IN (12,13,14,15,16,17,18,19,20,21,22,23,24,25) AND IFNULL(taf.tuition, 0) > 0 THEN 2
			ELSE 3
		END ASC,
		taf.id DESC`
	if query.SortModel.OrderByCreatedTime > 0 {
		orderBy = `
			DATE_FORMAT(taf.created_time, '%Y-%m-%d %H:%i') ASC,
			CASE
				WHEN taf.source_type NOT IN (12,13,14,15,16,17,18,19,20,21,22,23,24,25) AND IFNULL(taf.quantity, 0) > 0 AND IFNULL(taf.tuition, 0) = 0 THEN 0
				WHEN taf.source_type IN (12,13,14,15,16,17,18,19,20,21,22,23,24,25) THEN 1
				WHEN taf.source_type NOT IN (12,13,14,15,16,17,18,19,20,21,22,23,24,25) AND IFNULL(taf.tuition, 0) > 0 THEN 2
				ELSE 3
			END ASC,
			taf.id ASC`
	}

	var total int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM tuition_account_flow taf
		WHERE `+whereSQL, args...).Scan(&total); err != nil {
		return model.SubTuitionAccountFlowRecordListResult{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			taf.id,
			taf.tuition_account_id,
			taf.student_id,
			IFNULL(s.stu_name, ''),
			CASE
				WHEN CHAR_LENGTH(IFNULL(s.mobile, '')) >= 7 THEN CONCAT(LEFT(s.mobile, 3), '****', RIGHT(s.mobile, 4))
				ELSE IFNULL(s.mobile, '')
			END,
			IFNULL(s.avatar_url, ''),
			taf.product_id,
			IFNULL(c.name, ''),
			taf.lesson_type,
			`+resolvedLessonChargingModeExpr("taf.lesson_charging_mode", "icq_taf.lesson_model")+` AS resolved_lesson_mode,
			taf.source_type,
			taf.source_id,
			taf.teaching_record_id,
			taf.created_time,
			CASE
				WHEN taf.source_type = 1
				 AND IFNULL(taf.lesson_charging_mode, 0) = 3
				 AND IFNULL(taf.quantity, 0) = 0
				 AND IFNULL(taf.tuition, 0) = 0
				 AND IFNULL(taf.balance_tuition, 0) = 0
				 AND IFNULL(taf.balance_quantity, 0) > 0
				THEN IFNULL(taf.balance_quantity, 0)
				ELSE IFNULL(taf.quantity, 0)
			END,
			IFNULL(taf.tuition, 0),
			IFNULL(taf.balance_quantity, 0),
			IFNULL(taf.balance_tuition, 0),
			IFNULL(taf.order_number, '')
		FROM tuition_account_flow taf
		LEFT JOIN inst_student s ON s.id = taf.student_id AND s.del_flag = 0
		LEFT JOIN inst_course c ON c.id = taf.product_id AND c.del_flag = 0
		LEFT JOIN tuition_account ta ON ta.id = taf.tuition_account_id AND ta.inst_id = taf.inst_id AND ta.del_flag = 0
		`+tuitionAccountQuotationJoinSQL+`
		WHERE `+whereSQL+`
		ORDER BY `+orderBy+`
		LIMIT ? OFFSET ?
	`, append(args, size, offset)...)
	if err != nil {
		return model.SubTuitionAccountFlowRecordListResult{}, err
	}
	defer rows.Close()

	items := make([]model.SubTuitionAccountFlowRecordItem, 0, size)
	for rows.Next() {
		var (
			item               model.SubTuitionAccountFlowRecordItem
			id                 int64
			tuitionAccountID   int64
			studentID          int64
			productID          int64
			sourceID           int64
			lessonType         sql.NullInt64
			lessonChargingMode sql.NullInt64
			teachingRecordID   sql.NullInt64
			createdTime        sql.NullTime
		)
		if err := rows.Scan(
			&id,
			&tuitionAccountID,
			&studentID,
			&item.StudentName,
			&item.StudentPhone,
			&item.StudentAvatar,
			&productID,
			&item.ProductName,
			&lessonType,
			&lessonChargingMode,
			&item.SourceType,
			&sourceID,
			&teachingRecordID,
			&createdTime,
			&item.Quantity,
			&item.Tuition,
			&item.BalanceQuantity,
			&item.BalanceTuition,
			&item.OrderNumber,
		); err != nil {
			return model.SubTuitionAccountFlowRecordListResult{}, err
		}
		item.ID = strconv.FormatInt(id, 10)
		item.TuitionAccountID = strconv.FormatInt(tuitionAccountID, 10)
		item.StudentID = strconv.FormatInt(studentID, 10)
		item.ProductID = strconv.FormatInt(productID, 10)
		item.SourceID = strconv.FormatInt(sourceID, 10)
		if lessonType.Valid {
			value := int(lessonType.Int64)
			item.LessonType = &value
		}
		if lessonChargingMode.Valid {
			value := int(lessonChargingMode.Int64)
			item.LessonChargingMode = &value
		}
		if teachingRecordID.Valid {
			item.TeachingRecordID = strconv.FormatInt(teachingRecordID.Int64, 10)
		}
		if createdTime.Valid {
			t := createdTime.Time
			item.CreatedTime = &t
		}
		items = append(items, item)
	}

	return model.SubTuitionAccountFlowRecordListResult{
		List:  items,
		Total: total,
	}, rows.Err()
}

func ensureColumnsOnTable(ctx context.Context, db *sql.DB, tableName string, columns map[string]string) error {
	for columnName, definition := range columns {
		var exists int
		if err := db.QueryRowContext(ctx, `
			SELECT COUNT(*)
			FROM information_schema.COLUMNS
			WHERE TABLE_SCHEMA = DATABASE()
			  AND TABLE_NAME = ?
			  AND COLUMN_NAME = ?
		`, tableName, columnName).Scan(&exists); err != nil {
			return err
		}
		if exists > 0 {
			continue
		}
		if _, err := db.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", tableName, definition)); err != nil {
			return err
		}
	}
	return nil
}
