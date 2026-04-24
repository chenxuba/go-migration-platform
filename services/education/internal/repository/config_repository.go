package repository

import (
	"context"
	"database/sql"
	"fmt"
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
	"autoSendBirthdayMessage":             {},
	"enableRechargeAccountChangeMessage":  {},
	"enabledClassReminder":                {},
	"enabledClassConsumptionReminder":     {},
	"enableAuditionSmsRemind":             {},
	"enableSendCouponRemindSms":           {},
	"enableSendChildBindNoticeToAdmin":    {},
	"enableTeachingBillRemindSms":         {},
	"studentAbsentClassSwitch":            {},
	"enabledRenewReminder":                {},
	"enableArrearagedSendMessage":         {},
	"enableLiquidationRemindMessage":      {},
	"enablePointChangeRemindMessage":      {},
	"enableOrgSendChildBindNoticeToAdmin": {},
	"enableLeaveApplyNumberLimit":         {},
	"enableLeaveApplyTimeLimit":           {},
	"enableRenewClassNum":                 {},
	"enableRenewValidityDay":              {},
	"enableRenewPrice":                    {},
}

func EnsureInstConfigUnifiedTimePeriodColumns(ctx context.Context, db *sql.DB) error {
	if err := ensureColumnsOnTable(ctx, db, "inst_config", map[string]string{
		"unified_time_period_json":                   "unified_time_period_json LONGTEXT NULL",
		"group_class_roll_call_sheet_template":       "group_class_roll_call_sheet_template VARCHAR(64) NULL",
		"enable_classroom_teaching":                  "enable_classroom_teaching TINYINT(1) NOT NULL DEFAULT 1",
		"enabled_one2one":                            "enabled_one2one TINYINT(1) NOT NULL DEFAULT 1",
		"enable_compose_lesson":                      "enable_compose_lesson TINYINT(1) NOT NULL DEFAULT 1",
		"enable_charge_by_hours":                     "enable_charge_by_hours TINYINT(1) NOT NULL DEFAULT 1",
		"enable_by_date_lesson":                      "enable_by_date_lesson TINYINT(1) NOT NULL DEFAULT 0",
		"enable_charge_by_price":                     "enable_charge_by_price TINYINT(1) NOT NULL DEFAULT 0",
		"enable_filter_holiday":                      "enable_filter_holiday TINYINT(1) NOT NULL DEFAULT 0",
		"enabled_arrears_rollcall":                   "enabled_arrears_rollcall TINYINT(1) NOT NULL DEFAULT 0",
		"enable_by_auto_teaching":                    "enable_by_auto_teaching TINYINT(1) NOT NULL DEFAULT 0",
		"enable_one_to_one_schedule_limit":           "enable_one_to_one_schedule_limit TINYINT(1) NOT NULL DEFAULT 0",
		"enable_schedule_conflict_continue":          "enable_schedule_conflict_continue TINYINT(1) NOT NULL DEFAULT 0",
		"schedule_teacher_selection_range":           "schedule_teacher_selection_range VARCHAR(32) NOT NULL DEFAULT 'all'",
		"enable_face_attendance_relate_teaching":     "enable_face_attendance_relate_teaching TINYINT(1) NOT NULL DEFAULT 0",
		"enable_face_attendance_check_in_notice":     "enable_face_attendance_check_in_notice TINYINT(1) NOT NULL DEFAULT 0",
		"enable_face_attendance_check_out_notice":    "enable_face_attendance_check_out_notice TINYINT(1) NOT NULL DEFAULT 0",
		"enable_by_voice_tips":                       "enable_by_voice_tips TINYINT(1) NOT NULL DEFAULT 0",
		"enable_send_face_attend_notice_to_admin":    "enable_send_face_attend_notice_to_admin TINYINT(1) NOT NULL DEFAULT 0",
		"face_attendance_interval":                   "face_attendance_interval VARCHAR(32) NOT NULL DEFAULT '1'",
		"default_class_time_record_mode":             "default_class_time_record_mode INT NOT NULL DEFAULT 1",
		"default_student_class_time":                 "default_student_class_time DECIMAL(18,2) NOT NULL DEFAULT 1.00",
		"default_teacher_class_time":                 "default_teacher_class_time DECIMAL(18,2) NOT NULL DEFAULT 0.00",
		"charge_by_price_default_price":              "charge_by_price_default_price DECIMAL(18,2) NOT NULL DEFAULT 100.00",
		"enable_limit_single_order_arrears_deduct":   "enable_limit_single_order_arrears_deduct TINYINT(1) NOT NULL DEFAULT 0",
		"enable_hour_leave_normal_record":            "enable_hour_leave_normal_record TINYINT(1) NOT NULL DEFAULT 0",
		"enable_hour_truancy_normal_record":          "enable_hour_truancy_normal_record TINYINT(1) NOT NULL DEFAULT 0",
		"enable_period_makeup":                       "enable_period_makeup TINYINT(1) NOT NULL DEFAULT 0",
		"enable_period_auto_finish_when_zero":        "enable_period_auto_finish_when_zero TINYINT(1) NOT NULL DEFAULT 0",
		"enable_price_leave_normal_record":           "enable_price_leave_normal_record TINYINT(1) NOT NULL DEFAULT 0",
		"enable_price_truancy_normal_record":         "enable_price_truancy_normal_record TINYINT(1) NOT NULL DEFAULT 0",
		"enable_price_makeup":                        "enable_price_makeup TINYINT(1) NOT NULL DEFAULT 0",
		"auto_send_birthday_message":                 "auto_send_birthday_message TINYINT(1) NOT NULL DEFAULT 0",
		"enable_recharge_account_change_message":     "enable_recharge_account_change_message TINYINT(1) NOT NULL DEFAULT 0",
		"enabled_class_reminder":                     "enabled_class_reminder TINYINT(1) NOT NULL DEFAULT 0",
		"enabled_class_consumption_reminder":         "enabled_class_consumption_reminder TINYINT(1) NOT NULL DEFAULT 0",
		"enable_audition_sms_remind":                 "enable_audition_sms_remind TINYINT(1) NOT NULL DEFAULT 0",
		"enable_send_coupon_remind_sms":              "enable_send_coupon_remind_sms TINYINT(1) NOT NULL DEFAULT 0",
		"enable_send_child_bind_notice_to_admin":     "enable_send_child_bind_notice_to_admin TINYINT(1) NOT NULL DEFAULT 0",
		"enable_teaching_bill_remind_sms":            "enable_teaching_bill_remind_sms TINYINT(1) NOT NULL DEFAULT 0",
		"student_absent_class_switch":                "student_absent_class_switch TINYINT(1) NOT NULL DEFAULT 0",
		"enabled_renew_reminder":                     "enabled_renew_reminder TINYINT(1) NOT NULL DEFAULT 0",
		"enable_arrearaged_send_message":             "enable_arrearaged_send_message TINYINT(1) NOT NULL DEFAULT 0",
		"enable_liquidation_remind_message":          "enable_liquidation_remind_message TINYINT(1) NOT NULL DEFAULT 0",
		"enable_point_change_remind_message":         "enable_point_change_remind_message TINYINT(1) NOT NULL DEFAULT 0",
		"enable_org_send_child_bind_notice_to_admin": "enable_org_send_child_bind_notice_to_admin TINYINT(1) NOT NULL DEFAULT 0",
		"send_class_reminder_msg_hour":               "send_class_reminder_msg_hour VARCHAR(16) NOT NULL DEFAULT '19:00'",
		"enable_leave_apply_number_limit":            "enable_leave_apply_number_limit TINYINT(1) NOT NULL DEFAULT 0",
		"leave_apply_cycle_limit":                    "leave_apply_cycle_limit VARCHAR(32) NOT NULL DEFAULT 'month'",
		"leave_apply_number_limit":                   "leave_apply_number_limit VARCHAR(32) NOT NULL DEFAULT '2'",
		"leave_apply_type_limit":                     "leave_apply_type_limit VARCHAR(32) NOT NULL DEFAULT 'course'",
		"enable_leave_apply_time_limit":              "enable_leave_apply_time_limit TINYINT(1) NOT NULL DEFAULT 0",
		"leave_apply_time_limit":                     "leave_apply_time_limit VARCHAR(32) NOT NULL DEFAULT '1.0'",
		"enable_renew_class_num":                     "enable_renew_class_num TINYINT(1) NOT NULL DEFAULT 0",
		"renew_class_num":                            "renew_class_num VARCHAR(32) NOT NULL DEFAULT '5'",
		"enable_renew_validity_day":                  "enable_renew_validity_day TINYINT(1) NOT NULL DEFAULT 0",
		"renew_validity_day":                         "renew_validity_day VARCHAR(32) NOT NULL DEFAULT '15'",
		"enable_renew_price":                         "enable_renew_price TINYINT(1) NOT NULL DEFAULT 0",
		"renew_price":                                "renew_price VARCHAR(32) NOT NULL DEFAULT '500'",
	}); err != nil {
		return err
	}

	if err := ensureInstConfigStringFieldTypes(ctx, db); err != nil {
		return err
	}

	_, err := db.ExecContext(ctx, `
		UPDATE inst_config
		SET charge_by_price_default_price = 100.00
		WHERE charge_by_price_default_price IS NULL
	`)
	return err
}

func ensureInstConfigStringFieldTypes(ctx context.Context, db *sql.DB) error {
	type fieldSpec struct {
		Column       string
		Definition   string
		NumericValue string
		StringValue  string
	}
	fields := []fieldSpec{
		{Column: "send_class_reminder_msg_hour", Definition: "VARCHAR(16) NOT NULL DEFAULT '19:00'", NumericValue: "19", StringValue: "19:00"},
		{Column: "face_attendance_interval", Definition: "VARCHAR(32) NOT NULL DEFAULT '1'", NumericValue: "1", StringValue: "1"},
		{Column: "leave_apply_cycle_limit", Definition: "VARCHAR(32) NOT NULL DEFAULT 'month'", NumericValue: "1", StringValue: "month"},
		{Column: "leave_apply_number_limit", Definition: "VARCHAR(32) NOT NULL DEFAULT '2'", NumericValue: "2", StringValue: "2"},
		{Column: "leave_apply_type_limit", Definition: "VARCHAR(32) NOT NULL DEFAULT 'course'", NumericValue: "1", StringValue: "course"},
		{Column: "leave_apply_time_limit", Definition: "VARCHAR(32) NOT NULL DEFAULT '1.0'", NumericValue: "1", StringValue: "1.0"},
		{Column: "renew_class_num", Definition: "VARCHAR(32) NOT NULL DEFAULT '5'", NumericValue: "5", StringValue: "5"},
		{Column: "renew_validity_day", Definition: "VARCHAR(32) NOT NULL DEFAULT '15'", NumericValue: "15", StringValue: "15"},
		{Column: "renew_price", Definition: "VARCHAR(32) NOT NULL DEFAULT '500'", NumericValue: "500", StringValue: "500"},
	}
	for _, field := range fields {
		if _, err := db.ExecContext(ctx, fmt.Sprintf("UPDATE inst_config SET %s = %s WHERE %s IS NULL", field.Column, field.NumericValue, field.Column)); err != nil {
			return err
		}
		if err := ensureColumnTypeOnTable(ctx, db, "inst_config", field.Column, "varchar", fmt.Sprintf("ALTER TABLE inst_config MODIFY COLUMN %s %s", field.Column, field.Definition)); err != nil {
			return err
		}
		if _, err := db.ExecContext(ctx, fmt.Sprintf("UPDATE inst_config SET %s = ? WHERE TRIM(%s) = ''", field.Column, field.Column), field.StringValue); err != nil {
			return err
		}
	}
	if _, err := db.ExecContext(ctx, `
		UPDATE inst_config
		SET send_class_reminder_msg_hour = CONCAT(LPAD(send_class_reminder_msg_hour, 2, '0'), ':00')
		WHERE send_class_reminder_msg_hour REGEXP '^[0-9]{1,2}$'
	`); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `
		UPDATE inst_config
		SET leave_apply_cycle_limit = 'month'
		WHERE leave_apply_cycle_limit REGEXP '^[0-9]+$'
	`); err != nil {
		return err
	}
	_, err := db.ExecContext(ctx, `
		UPDATE inst_config
		SET leave_apply_type_limit = 'course'
		WHERE leave_apply_type_limit REGEXP '^[0-9]+$'
	`)
	return err

}

func ensureColumnTypeOnTable(ctx context.Context, db *sql.DB, tableName, columnName, expectedType, ddl string) error {
	var dataType string
	if err := db.QueryRowContext(ctx, `
		SELECT DATA_TYPE
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = ?
		  AND COLUMN_NAME = ?
		LIMIT 1
	`, tableName, columnName).Scan(&dataType); err != nil {
		return err
	}
	if strings.EqualFold(strings.TrimSpace(dataType), strings.TrimSpace(expectedType)) {
		return nil
	}
	_, err := db.ExecContext(ctx, ddl)
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
			auto_send_birthday_message,
			enable_recharge_account_change_message,
			enabled_class_reminder,
			enabled_class_consumption_reminder,
			enable_audition_sms_remind,
			enable_send_coupon_remind_sms,
			enable_send_child_bind_notice_to_admin,
			enable_teaching_bill_remind_sms,
			student_absent_class_switch,
			enabled_renew_reminder,
			enable_arrearaged_send_message,
			enable_liquidation_remind_message,
			enable_point_change_remind_message,
			enable_org_send_child_bind_notice_to_admin,
			send_class_reminder_msg_hour,
			enable_leave_apply_number_limit,
			leave_apply_cycle_limit,
			leave_apply_number_limit,
			leave_apply_type_limit,
			enable_leave_apply_time_limit,
			leave_apply_time_limit,
			enable_renew_class_num,
			renew_class_num,
			enable_renew_validity_day,
			renew_validity_day,
			enable_renew_price,
			renew_price,
			del_flag,
			create_time,
			version
		)
		SELECT ?, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 'all', 0, 0, 0, 0, 0, '1', 1, 1.00, 0.00, 100.00, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, '19:00', 0, 'month', '2', 'course', 0, '1.0', 0, '5', 0, '15', 0, '500', 0, NOW(), 0
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
		"autoSendBirthdayMessage":             "auto_send_birthday_message",
		"enableRechargeAccountChangeMessage":  "enable_recharge_account_change_message",
		"enabledClassReminder":                "enabled_class_reminder",
		"enabledClassConsumptionReminder":     "enabled_class_consumption_reminder",
		"enableAuditionSmsRemind":             "enable_audition_sms_remind",
		"enableSendCouponRemindSms":           "enable_send_coupon_remind_sms",
		"enableSendChildBindNoticeToAdmin":    "enable_send_child_bind_notice_to_admin",
		"enableTeachingBillRemindSms":         "enable_teaching_bill_remind_sms",
		"studentAbsentClassSwitch":            "student_absent_class_switch",
		"enabledRenewReminder":                "enabled_renew_reminder",
		"enableArrearagedSendMessage":         "enable_arrearaged_send_message",
		"enableLiquidationRemindMessage":      "enable_liquidation_remind_message",
		"enablePointChangeRemindMessage":      "enable_point_change_remind_message",
		"enableOrgSendChildBindNoticeToAdmin": "enable_org_send_child_bind_notice_to_admin",
		"sendClassReminderMsgHour":            "send_class_reminder_msg_hour",
		"enableLeaveApplyNumberLimit":         "enable_leave_apply_number_limit",
		"leaveApplyCycleLimit":                "leave_apply_cycle_limit",
		"leaveApplyNumberLimit":               "leave_apply_number_limit",
		"leaveApplyTypeLimit":                 "leave_apply_type_limit",
		"enableLeaveApplyTimeLimit":           "enable_leave_apply_time_limit",
		"leaveApplyTimeLimit":                 "leave_apply_time_limit",
		"enableRenewClassNum":                 "enable_renew_class_num",
		"renewClassNum":                       "renew_class_num",
		"enableRenewValidityDay":              "enable_renew_validity_day",
		"renewValidityDay":                    "renew_validity_day",
		"enableRenewPrice":                    "enable_renew_price",
		"renewPrice":                          "renew_price",
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
