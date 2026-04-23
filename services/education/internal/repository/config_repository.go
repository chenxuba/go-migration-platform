package repository

import (
	"context"
	"database/sql"
	"strings"
)

var instConfigBooleanFields = map[string]struct{}{
	"enablePublicPool":        {},
	"enableCollectorStaff":    {},
	"enablePhoneSellStaff":    {},
	"enableForeground":        {},
	"enableViceSellStaff":     {},
	"enableAdvisor":           {},
	"enableStudentManager":    {},
	"limitSameWeChat":         {},
	"limitImportSameWeChat":   {},
	"enableClassroomTeaching": {},
	"enabledOne2one":          {},
	"enableComposeLesson":     {},
	"enableChargeByHours":     {},
	"enableByDateLesson":      {},
	"enableChargeByPrice":     {},
}

func EnsureInstConfigUnifiedTimePeriodColumns(ctx context.Context, db *sql.DB) error {
	return ensureColumnsOnTable(ctx, db, "inst_config", map[string]string{
		"unified_time_period_json":             "unified_time_period_json LONGTEXT NULL",
		"group_class_roll_call_sheet_template": "group_class_roll_call_sheet_template VARCHAR(64) NULL",
		"enable_classroom_teaching":            "enable_classroom_teaching TINYINT(1) NOT NULL DEFAULT 1",
		"enabled_one2one":                      "enabled_one2one TINYINT(1) NOT NULL DEFAULT 1",
		"enable_compose_lesson":                "enable_compose_lesson TINYINT(1) NOT NULL DEFAULT 1",
		"enable_charge_by_hours":               "enable_charge_by_hours TINYINT(1) NOT NULL DEFAULT 1",
		"enable_by_date_lesson":                "enable_by_date_lesson TINYINT(1) NOT NULL DEFAULT 0",
		"enable_charge_by_price":               "enable_charge_by_price TINYINT(1) NOT NULL DEFAULT 0",
	})
}

func (repo *Repository) GetGroupClassRollCallSheetTemplateKey(ctx context.Context, instID int64) (string, error) {
	var key string
	err := repo.db.QueryRowContext(ctx, `
		SELECT IFNULL(group_class_roll_call_sheet_template, '')
		FROM inst_config
		WHERE inst_id = ? AND del_flag = 0
		LIMIT 1
	`, instID).Scan(&key)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(key), nil
}

func (repo *Repository) GetInstConfig(ctx context.Context, instID int64) (map[string]any, error) {
	rows, err := repo.db.QueryContext(ctx, "SELECT * FROM inst_config WHERE inst_id = ? AND del_flag = 0 LIMIT 1", instID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		return map[string]any{}, nil
	}

	values := make([]any, len(columns))
	valuePtrs := make([]any, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}
	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, err
	}

	result := make(map[string]any, len(columns))
	for i, col := range columns {
		key := snakeToCamel(col)
		normalized := normalizeDBValue(values[i])
		if _, ok := instConfigBooleanFields[key]; ok {
			switch typed := normalized.(type) {
			case int64:
				normalized = typed != 0
			case int32:
				normalized = typed != 0
			case int:
				normalized = typed != 0
			case uint8:
				normalized = typed != 0
			case string:
				normalized = strings.TrimSpace(typed) == "1" || strings.EqualFold(strings.TrimSpace(typed), "true")
			}
		}
		result[key] = normalized
	}
	return result, nil
}

func (repo *Repository) CreateDefaultInstConfig(ctx context.Context, instID int64) error {
	_, err := repo.db.ExecContext(ctx, `
		INSERT INTO inst_config (
			inst_id,
			add_import_student_rule,
			add_intention_student_rule,
			enable_classroom_teaching,
			enabled_one2one,
			enable_compose_lesson,
			enable_charge_by_hours,
			enable_by_date_lesson,
			enable_charge_by_price,
			enable_collector_staff,
			enable_phone_sell_staff,
			enable_foreground,
			enable_vice_sell_staff,
			enable_advisor,
			enable_student_manager,
			limit_same_weChat,
			limit_import_same_weChat,
			enable_public_pool,
			del_flag,
			create_time,
			version
		)
		SELECT ?, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, NOW(), 0
		FROM DUAL
		WHERE NOT EXISTS (
			SELECT 1
			FROM inst_config
			WHERE inst_id = ? AND del_flag = 0
		)
	`, instID, instID)
	return err
}

func (repo *Repository) UpdateInstConfig(ctx context.Context, instID int64, payload map[string]any) error {
	allowed := map[string]string{
		"addIntentionStudentRule":         "add_intention_student_rule",
		"addImportStudentRule":            "add_import_student_rule",
		"enablePublicPool":                "enable_public_pool",
		"enableClassroomTeaching":         "enable_classroom_teaching",
		"enabledOne2one":                  "enabled_one2one",
		"enableComposeLesson":             "enable_compose_lesson",
		"enableChargeByHours":             "enable_charge_by_hours",
		"enableByDateLesson":              "enable_by_date_lesson",
		"enableChargeByPrice":             "enable_charge_by_price",
		"groupClassRollCallSheetTemplate": "group_class_roll_call_sheet_template",
		"unfollowedTime":                  "unfollowed_time",
		"enableCollectorStaff":            "enable_collector_staff",
		"enablePhoneSellStaff":            "enable_phone_sell_staff",
		"enableForeground":                "enable_foreground",
		"enableViceSellStaff":             "enable_vice_sell_staff",
		"enableAdvisor":                   "enable_advisor",
		"enableStudentManager":            "enable_student_manager",
		"limitSameWeChat":                 "limit_same_weChat",
		"limitImportSameWeChat":           "limit_import_same_weChat",
		// unifiedTimePeriodJson 已改为 inst_period_* 表存储，勿再通过本方法写入 LONGTEXT
	}

	setClauses := make([]string, 0, len(payload)+1)
	args := make([]any, 0, len(payload)+1)
	for key, value := range payload {
		column, ok := allowed[strings.TrimSpace(key)]
		if !ok {
			continue
		}
		setClauses = append(setClauses, column+" = ?")
		args = append(args, normalizeUpdateValue(value))
	}
	if len(setClauses) == 0 {
		return nil
	}

	setClauses = append(setClauses, "update_time = NOW()")
	args = append(args, instID)
	_, err := repo.db.ExecContext(ctx, `
		UPDATE inst_config
		SET `+strings.Join(setClauses, ", ")+`
		WHERE inst_id = ? AND del_flag = 0
	`, args...)
	return err
}
