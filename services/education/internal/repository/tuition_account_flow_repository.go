package repository

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

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
			created_time DATETIME NOT NULL,
			quantity DECIMAL(18,2) NOT NULL DEFAULT 0,
			tuition DECIMAL(18,2) NOT NULL DEFAULT 0,
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
	return err
}

func (repo *Repository) ensureHistoricalTuitionAccountFlowRecords(ctx context.Context, instID int64) error {
	if _, err := repo.db.ExecContext(ctx, `
		DELETE FROM tuition_account_flow
		WHERE inst_id = ? AND source_type = ? AND teaching_record_id IS NULL
	`, instID, model.TuitionAccountFlowSourceRegistration); err != nil {
		return err
	}

	_, err := repo.db.ExecContext(ctx, `
		INSERT INTO tuition_account_flow (
			uuid, version, inst_id, tuition_account_id, student_id, product_id, lesson_type, lesson_charging_mode,
			source_type, source_id, teaching_record_id, created_time, quantity, tuition,
			create_id, create_time, update_id, update_time, del_flag
		)
		SELECT
			UUID(), 0, ta.inst_id, MIN(ta.id), ta.student_id, ta.course_id, MAX(ic.teach_method), MAX(icq.lesson_model),
			?, ta.order_course_detail_id, NULL, MIN(ta.create_time),
			CASE
				WHEN IFNULL(MAX(icq.lesson_model), 0) = 3 THEN IFNULL(SUM(ta.total_tuition), 0)
				ELSE IFNULL(SUM(ta.total_quantity), 0) + IFNULL(SUM(ta.free_quantity), 0)
			END,
			IFNULL(SUM(ta.total_tuition), 0),
			MIN(IFNULL(ta.create_id, 0)), MIN(IFNULL(ta.create_time, NOW())),
			MIN(IFNULL(ta.update_id, 0)), MAX(IFNULL(ta.update_time, NOW())), 0
		FROM tuition_account ta
		INNER JOIN inst_course ic ON ic.id = ta.course_id AND ic.del_flag = 0
		LEFT JOIN inst_course_quotation icq ON icq.id = ta.quote_id AND icq.del_flag = 0
		WHERE ta.inst_id = ? AND ta.del_flag = 0
		GROUP BY ta.inst_id, ta.student_id, ta.course_id, ta.order_course_detail_id
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

	whereParts := []string{"taf.inst_id = ?", "taf.del_flag = 0"}
	args := []any{instID}

	if strings.TrimSpace(query.QueryModel.ProductID) != "" {
		whereParts = append(whereParts, "CAST(taf.product_id AS CHAR) = ?")
		args = append(args, strings.TrimSpace(query.QueryModel.ProductID))
	}
	if strings.TrimSpace(query.QueryModel.StudentID) != "" {
		whereParts = append(whereParts, "CAST(taf.student_id AS CHAR) = ?")
		args = append(args, strings.TrimSpace(query.QueryModel.StudentID))
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
	orderBy := "taf.created_time DESC, taf.id DESC"
	if query.SortModel.OrderByCreatedTime > 0 {
		orderBy = "taf.created_time ASC, taf.id ASC"
	}

	var total int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM tuition_account_flow taf
		WHERE `+whereSQL, args...).Scan(&total); err != nil {
		return model.TuitionAccountFlowRecordListResult{}, err
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
			taf.lesson_charging_mode,
			taf.source_type,
			taf.source_id,
			taf.teaching_record_id,
			taf.created_time,
			IFNULL(taf.quantity, 0),
			IFNULL(taf.tuition, 0)
		FROM tuition_account_flow taf
		LEFT JOIN inst_student s ON s.id = taf.student_id AND s.del_flag = 0
		LEFT JOIN inst_course c ON c.id = taf.product_id AND c.del_flag = 0
		WHERE `+whereSQL+`
		ORDER BY `+orderBy+`
		LIMIT ? OFFSET ?
	`, append(args, size, offset)...)
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
