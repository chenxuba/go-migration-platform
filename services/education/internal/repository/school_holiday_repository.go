package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

func ensureSchoolHolidayTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS inst_school_holiday (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			inst_id BIGINT NOT NULL,
			name VARCHAR(100) NOT NULL,
			start_date DATE NOT NULL,
			end_date DATE NOT NULL,
			source VARCHAR(16) NOT NULL DEFAULT 'custom',
			sort INT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			KEY idx_inst_school_holiday_inst (inst_id, del_flag),
			KEY idx_inst_school_holiday_dates (inst_id, start_date, end_date)
		)
	`)
	return err
}

func defaultSchoolHolidaySeeds(year int) []model.SchoolHolidayMutation {
	switch year {
	case 2026:
		return []model.SchoolHolidayMutation{
			{Name: "元旦", StartDate: "2026-01-01", EndDate: "2026-01-03", Source: "statutory"},
			{Name: "春节", StartDate: "2026-02-15", EndDate: "2026-02-23", Source: "statutory"},
			{Name: "清明节", StartDate: "2026-04-04", EndDate: "2026-04-06", Source: "statutory"},
			{Name: "劳动节", StartDate: "2026-05-01", EndDate: "2026-05-05", Source: "statutory"},
			{Name: "端午节", StartDate: "2026-06-19", EndDate: "2026-06-21", Source: "statutory"},
			{Name: "中秋节", StartDate: "2026-09-25", EndDate: "2026-09-27", Source: "statutory"},
			{Name: "国庆节", StartDate: "2026-10-01", EndDate: "2026-10-07", Source: "statutory"},
		}
	default:
		return defaultSchoolHolidaySeeds(2026)
	}
}

func normalizeSchoolHolidaySource(source string) string {
	if strings.TrimSpace(source) == "statutory" {
		return "statutory"
	}
	return "custom"
}

func parseSchoolHolidayDate(value string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", strings.TrimSpace(value), time.Local)
}

func (repo *Repository) CountSchoolHolidayRecords(ctx context.Context, instID int64) (int, error) {
	var count int
	err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM inst_school_holiday
		WHERE inst_id = ?
	`, instID).Scan(&count)
	return count, err
}

func (repo *Repository) ListSchoolHolidays(ctx context.Context, instID int64) ([]model.SchoolHolidayVO, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, inst_id, IFNULL(name, ''), DATE_FORMAT(start_date, '%Y-%m-%d'), DATE_FORMAT(end_date, '%Y-%m-%d'),
		       IFNULL(source, 'custom'), IFNULL(sort, 0)
		FROM inst_school_holiday
		WHERE inst_id = ? AND del_flag = 0
		ORDER BY start_date ASC, end_date ASC, sort ASC, id ASC
	`, instID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.SchoolHolidayVO, 0, 16)
	for rows.Next() {
		var item model.SchoolHolidayVO
		if err := rows.Scan(&item.ID, &item.InstID, &item.Name, &item.StartDate, &item.EndDate, &item.Source, &item.Sort); err != nil {
			return nil, err
		}
		item.Source = normalizeSchoolHolidaySource(item.Source)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repo *Repository) CreateSchoolHoliday(ctx context.Context, instID int64, input model.SchoolHolidayMutation) (int64, error) {
	result, err := repo.db.ExecContext(ctx, `
		INSERT INTO inst_school_holiday (
			inst_id, name, start_date, end_date, source, sort, create_time, update_time, del_flag
		) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
	`,
		instID,
		strings.TrimSpace(input.Name),
		strings.TrimSpace(input.StartDate),
		strings.TrimSpace(input.EndDate),
		normalizeSchoolHolidaySource(input.Source),
		intValueWithFallback(input.Sort, 0),
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (repo *Repository) UpdateSchoolHoliday(ctx context.Context, instID int64, input model.SchoolHolidayMutation) error {
	if input.ID == nil || *input.ID <= 0 {
		return errors.New("id is required")
	}
	_, err := repo.db.ExecContext(ctx, `
		UPDATE inst_school_holiday
		SET name = ?, start_date = ?, end_date = ?, source = ?, sort = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`,
		strings.TrimSpace(input.Name),
		strings.TrimSpace(input.StartDate),
		strings.TrimSpace(input.EndDate),
		normalizeSchoolHolidaySource(input.Source),
		intValueWithFallback(input.Sort, 0),
		*input.ID,
		instID,
	)
	return err
}

func (repo *Repository) DeleteSchoolHoliday(ctx context.Context, instID int64, holidayID int64) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE inst_school_holiday
		SET del_flag = 1, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, holidayID, instID)
	return err
}

func (repo *Repository) SeedDefaultSchoolHolidays(ctx context.Context, instID int64, year int) error {
	totalCount, err := repo.CountSchoolHolidayRecords(ctx, instID)
	if err != nil {
		return err
	}
	if totalCount > 0 {
		return nil
	}
	seeds := defaultSchoolHolidaySeeds(year)
	for index, item := range seeds {
		sortValue := index + 1
		payload := item
		payload.Sort = &sortValue
		if _, err := repo.CreateSchoolHoliday(ctx, instID, payload); err != nil {
			return err
		}
	}
	return nil
}

func (repo *Repository) ResetSchoolHolidaysToDefault(ctx context.Context, instID int64, year int) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `
		UPDATE inst_school_holiday
		SET del_flag = 1, update_time = NOW()
		WHERE inst_id = ? AND del_flag = 0
	`, instID); err != nil {
		return err
	}

	for index, item := range defaultSchoolHolidaySeeds(year) {
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO inst_school_holiday (
				inst_id, name, start_date, end_date, source, sort, create_time, update_time, del_flag
			) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
		`,
			instID,
			strings.TrimSpace(item.Name),
			strings.TrimSpace(item.StartDate),
			strings.TrimSpace(item.EndDate),
			normalizeSchoolHolidaySource(item.Source),
			index+1,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func validateSchoolHolidayInput(input model.SchoolHolidayMutation) error {
	if strings.TrimSpace(input.Name) == "" {
		return errors.New("节假日名称不能为空")
	}
	startDate, err := parseSchoolHolidayDate(input.StartDate)
	if err != nil {
		return errors.New("开始日期格式不正确")
	}
	endDate, err := parseSchoolHolidayDate(input.EndDate)
	if err != nil {
		return errors.New("结束日期格式不正确")
	}
	if endDate.Before(startDate) {
		return errors.New("结束日期不能早于开始日期")
	}
	return nil
}

func intValueWithFallback(value *int, fallback int) int {
	if value == nil {
		return fallback
	}
	return *value
}

func (repo *Repository) SaveSchoolHoliday(ctx context.Context, instID int64, input model.SchoolHolidayMutation) (int64, error) {
	if err := validateSchoolHolidayInput(input); err != nil {
		return 0, err
	}
	if input.ID != nil && *input.ID > 0 {
		return *input.ID, repo.UpdateSchoolHoliday(ctx, instID, input)
	}
	return repo.CreateSchoolHoliday(ctx, instID, input)
}

func (repo *Repository) GetSchoolHolidayByID(ctx context.Context, instID, holidayID int64) (model.SchoolHolidayVO, error) {
	var item model.SchoolHolidayVO
	err := repo.db.QueryRowContext(ctx, `
		SELECT id, inst_id, IFNULL(name, ''), DATE_FORMAT(start_date, '%Y-%m-%d'), DATE_FORMAT(end_date, '%Y-%m-%d'),
		       IFNULL(source, 'custom'), IFNULL(sort, 0)
		FROM inst_school_holiday
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, holidayID, instID).Scan(&item.ID, &item.InstID, &item.Name, &item.StartDate, &item.EndDate, &item.Source, &item.Sort)
	if err != nil {
		if err == sql.ErrNoRows {
			return item, fmt.Errorf("节假日配置不存在")
		}
		return item, err
	}
	item.Source = normalizeSchoolHolidaySource(item.Source)
	return item, nil
}
