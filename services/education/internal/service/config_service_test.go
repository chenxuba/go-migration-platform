package service

import (
	"database/sql/driver"
	"testing"
)

func TestGetDefaultStudentFields_RepairsDuplicateDefaults(t *testing.T) {
	userID := int64(301)
	instID := int64(10048)

	svc, cleanup := newScriptedService(t, []queryExpectation{
		{
			query: `
				SELECT current_inst_id
				FROM sso_user
				WHERE id = ? AND del_flag = 0
				LIMIT 1
			`,
			args:    []any{userID},
			columns: []string{"current_inst_id"},
			rows:    [][]driver.Value{{instID}},
		},
		{
			query: `
				SELECT u.inst_id
				FROM inst_user u
				LEFT JOIN org_institution i ON u.inst_id = i.id
				WHERE u.del_flag = 0 AND u.disabled = 0
				  AND i.del_flag = 0 AND i.enabled = 1
				  AND (
				    i.expire_end_time > NOW()
				    OR IFNULL(i.open_type, 0) <> 1
				    OR i.expire_end_time IS NULL
				  )
				  AND u.user_id = ?
				  AND u.inst_id = ?
				  AND i.organ_type != 2 AND i.organ_type != 10 AND i.organ_type != 11
				LIMIT 1
			`,
			args:    []any{userID, instID},
			columns: []string{"inst_id"},
			rows:    [][]driver.Value{{instID}},
		},
		{
			query: `
				SELECT id, IFNULL(uuid, ''), IFNULL(version, 0), inst_id, field_key, field_type,
				       IFNULL(required, 0), IFNULL(searched, 0), IFNULL(options_json, ''),
				       IFNULL(is_default, 0), IFNULL(is_display, 0), IFNULL(can_delete, 0),
				       IFNULL(can_edit, 0), IFNULL(sort, 0), IFNULL(remark, '')
				FROM inst_student_field_key
				WHERE inst_id = ? AND is_default = ? AND del_flag = 0
				ORDER BY sort ASC, id ASC
			`,
			args:    []any{instID, true},
			columns: []string{"id", "uuid", "version", "inst_id", "field_key", "field_type", "required", "searched", "options_json", "is_default", "is_display", "can_delete", "can_edit", "sort", "remark"},
			rows: [][]driver.Value{
				{int64(49), "", int64(0), instID, "学员姓名", int64(1), int64(1), int64(1), "", int64(1), int64(1), int64(0), int64(0), int64(0), ""},
				{int64(58), "", int64(11), instID, "兴趣爱好", int64(1), int64(0), int64(1), "", int64(1), int64(0), int64(1), int64(1), int64(0), ""},
				{int64(102), "", int64(0), instID, "学员姓名", int64(1), int64(1), int64(1), "", int64(1), int64(1), int64(0), int64(0), int64(1), ""},
				{int64(111), "", int64(0), instID, "兴趣爱好", int64(1), int64(1), int64(0), "", int64(1), int64(1), int64(1), int64(1), int64(10), ""},
				{int64(113), "", int64(0), instID, "学员姓名", int64(1), int64(1), int64(1), "", int64(1), int64(1), int64(0), int64(0), int64(1), ""},
				{int64(122), "", int64(0), instID, "兴趣爱好", int64(1), int64(1), int64(0), "", int64(1), int64(1), int64(1), int64(1), int64(10), ""},
			},
		},
		execResultExpectation(`
			UPDATE inst_student_field_key
			SET del_flag = 1, version = IFNULL(version, 0) + 1, update_time = NOW()
			WHERE id IN (?,?,?,?) AND del_flag = 0
		`, []any{49, 58, 102, 111}, 4),
		{
			query: `
				SELECT id, IFNULL(uuid, ''), IFNULL(version, 0), inst_id, field_key, field_type,
				       IFNULL(required, 0), IFNULL(searched, 0), IFNULL(options_json, ''),
				       IFNULL(is_default, 0), IFNULL(is_display, 0), IFNULL(can_delete, 0),
				       IFNULL(can_edit, 0), IFNULL(sort, 0), IFNULL(remark, '')
				FROM inst_student_field_key
				WHERE inst_id = ? AND is_default = ? AND del_flag = 0
				ORDER BY sort ASC, id ASC
			`,
			args:    []any{instID, true},
			columns: []string{"id", "uuid", "version", "inst_id", "field_key", "field_type", "required", "searched", "options_json", "is_default", "is_display", "can_delete", "can_edit", "sort", "remark"},
			rows: [][]driver.Value{
				{int64(113), "", int64(0), instID, "学员姓名", int64(1), int64(1), int64(1), "", int64(1), int64(1), int64(0), int64(0), int64(1), ""},
				{int64(122), "", int64(0), instID, "兴趣爱好", int64(1), int64(1), int64(0), "", int64(1), int64(1), int64(1), int64(1), int64(10), ""},
			},
		},
	})
	defer cleanup()

	result, err := svc.GetDefaultStudentFields(userID)
	if err != nil {
		t.Fatalf("GetDefaultStudentFields returned error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 default fields, got %d", len(result))
	}
	if result[0].ID != 113 || result[1].ID != 122 {
		t.Fatalf("unexpected repaired fields: %+v", result)
	}
}
