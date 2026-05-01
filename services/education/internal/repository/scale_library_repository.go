package repository

import (
	"context"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

func (repo *Repository) ensureScaleLibrarySchema(ctx context.Context) error {
	exists, err := repo.tableExists(ctx, "sys_scale")
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	if err := ensureColumnsOnTable(ctx, repo.db, "sys_scale", map[string]string{
		"age_min_months":       "age_min_months INT NOT NULL DEFAULT 0 AFTER age_range",
		"age_max_months":       "age_max_months INT NOT NULL DEFAULT 0 AFTER age_min_months",
		"estimated_duration":   "estimated_duration VARCHAR(64) NOT NULL DEFAULT '' AFTER age_range",
		"duration_min_minutes": "duration_min_minutes INT NOT NULL DEFAULT 0 AFTER estimated_duration",
		"duration_max_minutes": "duration_max_minutes INT NOT NULL DEFAULT 0 AFTER duration_min_minutes",
		"poster_url":           "poster_url VARCHAR(500) DEFAULT NULL AFTER summary",
	}); err != nil {
		return err
	}
	_, err = repo.db.ExecContext(ctx, `
		UPDATE sys_scale
		SET age_range = '2.6岁-6岁',
		    age_min_months = 30,
		    age_max_months = 72
		WHERE scale_code = 'PEP3' AND del_flag = 0
		  AND (IFNULL(age_min_months, 0) = 0 OR IFNULL(age_max_months, 0) = 0)
	`)
	return err
}

func (repo *Repository) ListInstitutionScaleLibrary(ctx context.Context, instID int64, query model.ScaleLibraryQuery) ([]model.ScaleLibraryItem, error) {
	filters := []string{"s.del_flag = 0"}
	filterArgs := make([]any, 0, 8)

	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		filters = append(filters, "(s.scale_name LIKE ? OR s.scale_code LIKE ? OR s.current_version LIKE ? OR s.category LIKE ? OR s.scenario LIKE ? OR s.age_range LIKE ? OR s.estimated_duration LIKE ? OR s.summary LIKE ?)")
		filterArgs = append(filterArgs, like, like, like, like, like, like, like, like)
	}
	if category := strings.TrimSpace(query.Category); category != "" {
		filters = append(filters, "s.category = ?")
		filterArgs = append(filterArgs, category)
	}
	if scenario := strings.TrimSpace(query.Scenario); scenario != "" {
		filters = append(filters, "s.scenario = ?")
		filterArgs = append(filterArgs, scenario)
	}

	whereClause := strings.Join(filters, " AND ")
	args := append([]any{instID}, filterArgs...)
	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			s.id,
			s.scale_name,
			s.scale_code,
			s.category,
			s.scenario,
			s.age_range,
			IFNULL(s.age_min_months, 0),
			IFNULL(s.age_max_months, 0),
			IFNULL(s.estimated_duration, ''),
			IFNULL(s.duration_min_minutes, 0),
			IFNULL(s.duration_max_minutes, 0),
			s.current_version,
			s.item_count,
			s.domain_count,
			s.institution_count,
			s.month_usage,
			IFNULL(stats.month_usage, 0),
			IFNULL(stats.usage_count, 0),
			IFNULL(DATE_FORMAT(stats.latest_use, '%Y-%m-%d'), ''),
			IFNULL(s.data_status, ''),
			IFNULL(DATE_FORMAT(s.update_time, '%Y-%m-%d %H:%i:%s'), ''),
			IFNULL(s.summary, ''),
			IFNULL(s.poster_url, ''),
			IFNULL(s.execution_entry, ''),
			IFNULL(s.api_package, '')
		FROM sys_scale s
		LEFT JOIN (
			SELECT
				assessment_code,
				COUNT(1) AS usage_count,
				SUM(CASE WHEN create_time >= DATE_FORMAT(CURDATE(), '%Y-%m-01') THEN 1 ELSE 0 END) AS month_usage,
				MAX(COALESCE(assessment_date, create_time)) AS latest_use
			FROM assessment_record
			WHERE del_flag = 0 AND inst_id = ?
			GROUP BY assessment_code
		) stats ON CONVERT(stats.assessment_code USING utf8mb4) COLLATE utf8mb4_unicode_ci = CONVERT(s.scale_code USING utf8mb4) COLLATE utf8mb4_unicode_ci
		WHERE `+whereClause+`
		ORDER BY s.sort ASC, s.id ASC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.ScaleLibraryItem, 0, 8)
	for rows.Next() {
		item := model.ScaleLibraryItem{
			References:        []model.ScaleLibraryTextResource{},
			Acknowledgements:  []model.ScaleLibraryTextResource{},
			AuthReserved:      true,
			AuthActionEnabled: false,
		}
		var configuredMonthUsage int
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Code,
			&item.Category,
			&item.Scenario,
			&item.AgeRange,
			&item.AgeMinMonths,
			&item.AgeMaxMonths,
			&item.Duration,
			&item.DurationMinMinutes,
			&item.DurationMaxMinutes,
			&item.CurrentVersion,
			&item.ItemCount,
			&item.DomainCount,
			&item.InstitutionCount,
			&configuredMonthUsage,
			&item.MonthUsage,
			&item.UsageCount,
			&item.LatestUse,
			&item.DataStatus,
			&item.UpdatedAt,
			&item.Summary,
			&item.PosterURL,
			&item.ExecutionEntry,
			&item.APIPackage,
		); err != nil {
			return nil, err
		}
		if item.MonthUsage == 0 {
			item.MonthUsage = configuredMonthUsage
		}
		item.Status, item.StatusText = scaleLibraryStatus(item)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return items, nil
	}
	if err := repo.loadScaleLibraryReferences(ctx, items); err != nil {
		return nil, err
	}
	if err := repo.loadScaleLibraryAcknowledgements(ctx, items); err != nil {
		return nil, err
	}
	return filterScaleLibraryItems(items, query), nil
}

func (repo *Repository) ListScaleLibraryCategoryOptions(ctx context.Context) ([]string, error) {
	dictExists, err := repo.tableExists(ctx, "sys_dict")
	if err != nil {
		return nil, err
	}
	if !dictExists {
		return nil, nil
	}
	valueExists, err := repo.tableExists(ctx, "sys_dict_value")
	if err != nil {
		return nil, err
	}
	if !valueExists {
		return nil, nil
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT TRIM(IFNULL(v.dict_value, ''))
		FROM sys_dict_value v
		JOIN sys_dict d ON v.dict_id = d.id
		WHERE d.dict_code = 'scale_category'
		  AND d.del_flag = 0
		  AND v.del_flag = 0
		  AND IFNULL(v.is_enable, 0) = 1
		ORDER BY IFNULL(v.sort, 0) ASC, v.id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]string, 0, 8)
	seen := make(map[string]bool)
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		items = append(items, value)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func scaleLibraryStatus(item model.ScaleLibraryItem) (string, string) {
	if strings.TrimSpace(item.ExecutionEntry) == "" && strings.TrimSpace(item.APIPackage) == "" {
		return "unavailable", "暂不可用"
	}
	return "available", "可用"
}

func filterScaleLibraryItems(items []model.ScaleLibraryItem, query model.ScaleLibraryQuery) []model.ScaleLibraryItem {
	status := strings.TrimSpace(query.Status)
	ageScope := strings.TrimSpace(query.AgeScope)
	duration := strings.TrimSpace(query.Duration)
	if status == "" && ageScope == "" && duration == "" {
		return items
	}
	out := make([]model.ScaleLibraryItem, 0, len(items))
	for _, item := range items {
		if status != "" && item.Status != status {
			continue
		}
		if ageScope != "" && !scaleLibraryAgeRangeMatches(item.AgeMinMonths, item.AgeMaxMonths, ageScope) {
			continue
		}
		if duration != "" && !scaleLibraryDurationMatches(item.DurationMinMinutes, item.DurationMaxMinutes, duration) {
			continue
		}
		out = append(out, item)
	}
	return out
}

func scaleLibraryAgeRangeMatches(minMonths, maxMonths int, ageScope string) bool {
	if minMonths <= 0 && maxMonths <= 0 {
		return false
	}
	if maxMonths <= 0 {
		maxMonths = minMonths
	}
	if minMonths <= 0 {
		minMonths = maxMonths
	}
	switch strings.TrimSpace(ageScope) {
	case "", "all":
		return true
	case "0-2":
		return maxMonths >= 0 && minMonths <= 24
	case "2-6":
		return maxMonths >= 24 && minMonths <= 72
	case "6-12":
		return maxMonths > 72 && minMonths <= 144
	case "12+":
		return maxMonths >= 144
	default:
		return false
	}
}

func scaleLibraryDurationMatches(minMinutes, maxMinutes int, duration string) bool {
	if minMinutes <= 0 && maxMinutes <= 0 {
		return false
	}
	if maxMinutes <= 0 {
		maxMinutes = minMinutes
	}
	if minMinutes <= 0 {
		minMinutes = maxMinutes
	}
	switch strings.TrimSpace(duration) {
	case "", "all":
		return true
	case "0-15":
		return maxMinutes <= 15
	case "15-30":
		return maxMinutes >= 15 && minMinutes <= 30
	case "30-60":
		return maxMinutes >= 30 && minMinutes <= 60
	case "60+":
		return maxMinutes > 60
	default:
		return false
	}
}

func (repo *Repository) loadScaleLibraryReferences(ctx context.Context, items []model.ScaleLibraryItem) error {
	scaleIDs := make([]int64, 0, len(items))
	indexByID := make(map[int64]int, len(items))
	for index, item := range items {
		scaleIDs = append(scaleIDs, item.ID)
		indexByID[item.ID] = index
	}
	placeholders, args := buildAssessmentInt64InClause(scaleIDs)
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, scale_id, content, sort
		FROM sys_scale_reference
		WHERE del_flag = 0 AND scale_id IN (`+placeholders+`)
		ORDER BY scale_id ASC, sort ASC, id ASC
	`, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item model.ScaleLibraryTextResource
		if err := rows.Scan(&item.ID, &item.ScaleID, &item.Content, &item.Sort); err != nil {
			return err
		}
		if index, ok := indexByID[item.ScaleID]; ok {
			items[index].References = append(items[index].References, item)
		}
	}
	return rows.Err()
}

func (repo *Repository) loadScaleLibraryAcknowledgements(ctx context.Context, items []model.ScaleLibraryItem) error {
	scaleIDs := make([]int64, 0, len(items))
	indexByID := make(map[int64]int, len(items))
	for index, item := range items {
		scaleIDs = append(scaleIDs, item.ID)
		indexByID[item.ID] = index
	}
	placeholders, args := buildAssessmentInt64InClause(scaleIDs)
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, scale_id, content, sort
		FROM sys_scale_acknowledgement
		WHERE del_flag = 0 AND scale_id IN (`+placeholders+`)
		ORDER BY scale_id ASC, sort ASC, id ASC
	`, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item model.ScaleLibraryTextResource
		if err := rows.Scan(&item.ID, &item.ScaleID, &item.Content, &item.Sort); err != nil {
			return err
		}
		if index, ok := indexByID[item.ScaleID]; ok {
			items[index].Acknowledgements = append(items[index].Acknowledgements, item)
		}
	}
	return rows.Err()
}

func buildAssessmentInt64InClause(values []int64) (string, []any) {
	if len(values) == 0 {
		return "NULL", nil
	}
	placeholders := make([]string, 0, len(values))
	args := make([]any, 0, len(values))
	for _, value := range values {
		placeholders = append(placeholders, "?")
		args = append(args, value)
	}
	return strings.Join(placeholders, ","), args
}
