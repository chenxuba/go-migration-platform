package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func ensureCourseSchema(ctx context.Context, db *sql.DB) error {
	if err := ensureLongTextColumnType(ctx, db, "inst_course_detail", "images"); err != nil {
		return err
	}
	if err := ensureLongTextColumnType(ctx, db, "inst_course_detail", "description"); err != nil {
		return err
	}
	return nil
}

func ensureLongTextColumnType(ctx context.Context, db *sql.DB, tableName, columnName string) error {
	var (
		dataType   string
		isNullable string
	)
	err := db.QueryRowContext(ctx, `
		SELECT DATA_TYPE, IS_NULLABLE
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = ?
		  AND COLUMN_NAME = ?
		LIMIT 1
	`, tableName, columnName).Scan(&dataType, &isNullable)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	if strings.EqualFold(dataType, "longtext") {
		return nil
	}

	nullClause := "NULL"
	if strings.EqualFold(isNullable, "NO") {
		nullClause = "NOT NULL"
	}
	_, err = db.ExecContext(ctx, fmt.Sprintf(
		"ALTER TABLE %s MODIFY COLUMN %s LONGTEXT %s",
		tableName,
		columnName,
		nullClause,
	))
	return err
}
