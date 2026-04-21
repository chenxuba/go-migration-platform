package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

type ParentCourseFlowRecord struct {
	ID         string
	SourceType int
	SourceID   string
	CreatedAt  time.Time
	Quantity   float64
	Tuition    float64
}

func (repo *Repository) ListParentCourseFlowRecordsByTuitionAccountIDs(ctx context.Context, instID int64, tuitionAccountIDs []int64, pageIndex, pageSize int) ([]ParentCourseFlowRecord, int, error) {
	if len(tuitionAccountIDs) == 0 {
		return []ParentCourseFlowRecord{}, 0, nil
	}
	if err := repo.ensureHistoricalTuitionAccountFlowRecords(ctx, instID); err != nil {
		return nil, 0, err
	}

	current := pageIndex
	if current <= 0 {
		current = 1
	}
	size := pageSize
	if size <= 0 {
		size = 20
	}
	offset := (current - 1) * size
	placeholders := sqlPlaceholders(len(tuitionAccountIDs))

	whereParts := []string{
		"taf.inst_id = ?",
		"taf.del_flag = 0",
		"taf.tuition_account_id IN (" + placeholders + ")",
		"NOT (taf.source_type = 15 AND taf.source_id >= 20000101 AND taf.source_id > CAST(DATE_FORMAT(NOW(), '%Y%m%d') AS UNSIGNED))",
	}
	args := make([]any, 0, len(tuitionAccountIDs)+3)
	args = append(args, instID)
	for _, tuitionAccountID := range tuitionAccountIDs {
		args = append(args, tuitionAccountID)
	}
	whereSQL := strings.Join(whereParts, " AND ")

	var total int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM (
			SELECT taf.source_type, taf.source_id
			FROM tuition_account_flow taf
			WHERE `+whereSQL+`
			GROUP BY taf.source_type, taf.source_id
		) grouped_flow
	`, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			CAST(MIN(taf.id) AS CHAR) AS id,
			taf.source_type,
			CAST(taf.source_id AS CHAR) AS source_id,
			MAX(taf.created_time) AS created_at,
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
			), 0) AS quantity,
			IFNULL(SUM(IFNULL(taf.tuition, 0)), 0) AS tuition
		FROM tuition_account_flow taf
		WHERE `+whereSQL+`
		GROUP BY taf.source_type, taf.source_id
		ORDER BY MAX(taf.created_time) DESC, MAX(taf.id) DESC
		LIMIT ? OFFSET ?
	`, append(append([]any{}, args...), size, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]ParentCourseFlowRecord, 0, size)
	for rows.Next() {
		var item ParentCourseFlowRecord
		var createdAt sql.NullTime
		if err := rows.Scan(
			&item.ID,
			&item.SourceType,
			&item.SourceID,
			&createdAt,
			&item.Quantity,
			&item.Tuition,
		); err != nil {
			return nil, 0, err
		}
		if createdAt.Valid {
			item.CreatedAt = createdAt.Time
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}
