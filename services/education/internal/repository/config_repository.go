package repository

import (
	"context"
	"database/sql"
	"strings"
)

var instConfigBooleanFields = map[string]struct{}{
	"enablePublicPool":                    {},
	"enableCollectorStaff":                {},
	"enablePhoneSellStaff":                {},
	"enableForeground":                    {},
	"enableViceSellStaff":                 {},
	"enableAdvisor":                       {},
	"enableStudentManager":                {},
	"limitSameWeChat":                     {},
	"limitImportSameWeChat":               {},
	"enableClassroomTeaching":             {},
	"enabledOne2one":                      {},
	"enableComposeLesson":                 {},
	"enableChargeByHours":                 {},
	"enableByDateLesson":                  {},
	"enableChargeByPrice":                 {},
	"enableFilterHoliday":                 {},
	"enabledArrearsRollcall":              {},
	"enableByAutoTeaching":                {},
	"enableOneToOneScheduleLimit":         {},
	"enableScheduleConflictContinue":      {},
	"enableFaceAttendanceRelateTeaching":  {},
	"enableFaceAttendanceCheckInNotice":   {},
	"enableFaceAttendanceCheckOutNotice":  {},
	"enableByVoiceTips":                   {},
	"enableSendFaceAttendNoticeToAdmin":   {},
	"enableLimitSingleOrderArrearsDeduct": {},
	"enableHourLeaveNormalRecord":         {},
	"enableHourTruancyNormalRecord":       {},
	"enablePeriodMakeup":                  {},
	"enablePeriodAutoFinishWhenZero":      {},
	"enablePriceLeaveNormalRecord":        {},
	"enablePriceTruancyNormalRecord":      {},
	"enablePriceMakeup":                   {},
}

func EnsureInstConfigUnifiedTimePeriodColumns(ctx context.Context, db *sql.DB) error {
	if err := ensureColumnsOnTable(ctx, db, "inst_config", map[string]string{
		"unified_time_period_json":                 "unified_time_period_json LONGTEXT NULL",
		"group_class_roll_call_sheet_template":     "group_class_roll_call_sheet_template VARCHAR(64) NULL",
		"enable_classroom_teaching":                "enable_classroom_teaching TINYINT(1) NOT NULL DEFAULT 1",
		"enabled_one2one":                          "enabled_one2one TINYINT(1) NOT NULL DEFAULT 1",
		"enable_compose_lesson":                    "enable_compose_lesson TINYINT(1) NOT NULL DEFAULT 1",
		"enable_charge_by_hours":                   "enable_charge_by_hours TINYINT(1) NOT NULL DEFAULT 1",
		"enable_by_date_lesson":                    "enable_by_date_lesson TINYINT(1) NOT NULL DEFAULT 0",
		"enable_charge_by_price":                   "enable_charge_by_price TINYINT(1) NOT NULL DEFAULT 0",
		"enable_filter_holiday":                    "enable_filter_holiday TINYINT(1) NOT NULL DEFAULT 0",
		"enabled_arrears_rollcall":                 "enabled_arrears_rollcall TINYINT(1) NOT NULL DEFAULT 0",
		"enable_by_auto_teaching":                  "enable_by_auto_teaching TINYINT(1) NOT NULL DEFAULT 0",
		"enable_one_to_one_schedule_limit":         "enable_one_to_one_schedule_limit TINYINT(1) NOT NULL DEFAULT 0",
		"enable_schedule_conflict_continue":        "enable_schedule_conflict_continue TINYINT(1) NOT NULL DEFAULT 0",
		"schedule_teacher_selection_range":         "schedule_teacher_selection_range VARCHAR(32) NOT NULL DEFAULT 'all'",
		"enable_face_attendance_relate_teaching":   "enable_face_attendance_relate_teaching TINYINT(1) NOT NULL DEFAULT 0",
		"enable_face_attendance_check_in_notice":   "enable_face_attendance_check_in_notice TINYINT(1) NOT NULL DEFAULT 0",
		"enable_face_attendance_check_out_notice":  "enable_face_attendance_check_out_notice TINYINT(1) NOT NULL DEFAULT 0",
		"enable_by_voice_tips":                     "enable_by_voice_tips TINYINT(1) NOT NULL DEFAULT 0",
		"enable_send_face_attend_notice_to_admin":  "enable_send_face_attend_notice_to_admin TINYINT(1) NOT NULL DEFAULT 0",
		"face_attendance_interval":                 "face_attendance_interval VARCHAR(32) NOT NULL DEFAULT '1'",
		"default_class_time_record_mode":           "default_class_time_record_mode INT NOT NULL DEFAULT 1",
		"default_student_class_time":               "default_student_class_time DECIMAL(18,2) NOT NULL DEFAULT 1.00",
		"default_teacher_class_time":               "default_teacher_class_time DECIMAL(18,2) NOT NULL DEFAULT 0.00",
		"charge_by_price_default_price":            "charge_by_price_default_price DECIMAL(18,2) NOT NULL DEFAULT 100.00",
		"enable_limit_single_order_arrears_deduct": "enable_limit_single_order_arrears_deduct TINYINT(1) NOT NULL DEFAULT 0",
		"enable_hour_leave_normal_record":          "enable_hour_leave_normal_record TINYINT(1) NOT NULL DEFAULT 0",
		"enable_hour_truancy_normal_record":        "enable_hour_truancy_normal_record TINYINT(1) NOT NULL DEFAULT 0",
		"enable_period_makeup":                     "enable_period_makeup TINYINT(1) NOT NULL DEFAULT 0",
		"enable_period_auto_finish_when_zero":      "enable_period_auto_finish_when_zero TINYINT(1) NOT NULL DEFAULT 0",
		"enable_price_leave_normal_record":         "enable_price_leave_normal_record TINYINT(1) NOT NULL DEFAULT 0",
		"enable_price_truancy_normal_record":       "enable_price_truancy_normal_record TINYINT(1) NOT NULL DEFAULT 0",
		"enable_price_makeup":                      "enable_price_makeup TINYINT(1) NOT NULL DEFAULT 0",
	}); err != nil {
		return err
	}

	_, err := db.ExecContext(ctx, `
		UPDATE inst_config
		SET charge_by_price_default_price = 100.00
		WHERE charge_by_price_default_price IS NULL
	`)
	return err
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
			enable_filter_holiday,
			enabled_arrears_rollcall,
			enable_by_auto_teaching,
			enable_one_to_one_schedule_limit,
			enable_schedule_conflict_continue,
			schedule_teacher_selection_range,
			enable_face_attendance_relate_teaching,
			enable_face_attendance_check_in_notice,
			enable_face_attendance_check_out_notice,
			enable_by_voice_tips,
			enable_send_face_attend_notice_to_admin,
			face_attendance_interval,
			default_class_time_record_mode,
			default_student_class_time,
			default_teacher_class_time,
			charge_by_price_default_price,
			enable_limit_single_order_arrears_deduct,
			enable_hour_leave_normal_record,
			enable_hour_truancy_normal_record,
			enable_period_makeup,
			enable_period_auto_finish_when_zero,
			enable_price_leave_normal_record,
			enable_price_truancy_normal_record,
			enable_price_makeup,
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
		SELECT ?, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 'all', 0, 0, 0, 0, 0, '1', 1, 1.00, 0.00, 100.00, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, NOW(), 0
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
		"addIntentionStudentRule":             "add_intention_student_rule",
		"addImportStudentRule":                "add_import_student_rule",
		"enablePublicPool":                    "enable_public_pool",
		"enableClassroomTeaching":             "enable_classroom_teaching",
		"enabledOne2one":                      "enabled_one2one",
		"enableComposeLesson":                 "enable_compose_lesson",
		"enableChargeByHours":                 "enable_charge_by_hours",
		"enableByDateLesson":                  "enable_by_date_lesson",
		"enableChargeByPrice":                 "enable_charge_by_price",
		"enableFilterHoliday":                 "enable_filter_holiday",
		"enabledArrearsRollcall":              "enabled_arrears_rollcall",
		"enableByAutoTeaching":                "enable_by_auto_teaching",
		"enableOneToOneScheduleLimit":         "enable_one_to_one_schedule_limit",
		"enableScheduleConflictContinue":      "enable_schedule_conflict_continue",
		"scheduleTeacherSelectionRange":       "schedule_teacher_selection_range",
		"enableFaceAttendanceRelateTeaching":  "enable_face_attendance_relate_teaching",
		"enableFaceAttendanceCheckInNotice":   "enable_face_attendance_check_in_notice",
		"enableFaceAttendanceCheckOutNotice":  "enable_face_attendance_check_out_notice",
		"enableByVoiceTips":                   "enable_by_voice_tips",
		"enableSendFaceAttendNoticeToAdmin":   "enable_send_face_attend_notice_to_admin",
		"faceAttendanceInterval":              "face_attendance_interval",
		"defaultClassTimeRecordMode":          "default_class_time_record_mode",
		"defaultStudentClassTime":             "default_student_class_time",
		"defaultTeacherClassTime":             "default_teacher_class_time",
		"chargeByPriceDefaultPrice":           "charge_by_price_default_price",
		"enableLimitSingleOrderArrearsDeduct": "enable_limit_single_order_arrears_deduct",
		"enableHourLeaveNormalRecord":         "enable_hour_leave_normal_record",
		"enableHourTruancyNormalRecord":       "enable_hour_truancy_normal_record",
		"enablePeriodMakeup":                  "enable_period_makeup",
		"enablePeriodAutoFinishWhenZero":      "enable_period_auto_finish_when_zero",
		"enablePriceLeaveNormalRecord":        "enable_price_leave_normal_record",
		"enablePriceTruancyNormalRecord":      "enable_price_truancy_normal_record",
		"enablePriceMakeup":                   "enable_price_makeup",
		"groupClassRollCallSheetTemplate":     "group_class_roll_call_sheet_template",
		"unfollowedTime":                      "unfollowed_time",
		"enableCollectorStaff":                "enable_collector_staff",
		"enablePhoneSellStaff":                "enable_phone_sell_staff",
		"enableForeground":                    "enable_foreground",
		"enableViceSellStaff":                 "enable_vice_sell_staff",
		"enableAdvisor":                       "enable_advisor",
		"enableStudentManager":                "enable_student_manager",
		"limitSameWeChat":                     "limit_same_weChat",
		"limitImportSameWeChat":               "limit_import_same_weChat",
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
