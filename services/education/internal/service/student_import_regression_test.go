package service

import (
	"database/sql/driver"
	"testing"

	"go-migration-platform/services/education/internal/model"
)

func TestAddIntentStudent_BlocksDuplicateWeChatWhenManualSettingEnabled(t *testing.T) {
	userID := int64(101)
	instID := int64(201)
	dto := model.StudentSaveDTO{
		StuName:      "张三",
		Mobile:       "13800138000",
		WeChatNumber: "wx-zhangsan",
	}

	svc, cleanup := newScriptedService(t, []queryExpectation{
		findInstIDExpectation(userID, instID),
		{
			query: `
				SELECT add_intention_student_rule
				FROM inst_config
				WHERE inst_id = ? AND del_flag = 0
				ORDER BY id DESC
				LIMIT 1
			`,
			args:    []any{instID},
			columns: []string{"add_intention_student_rule"},
			rows:    [][]driver.Value{{int64(1)}},
		},
		{
			query:   `SELECT COUNT(*) FROM inst_student WHERE inst_id = ? AND del_flag = 0 AND stu_name = ? AND mobile = ?`,
			args:    []any{instID, dto.StuName, dto.Mobile},
			columns: []string{"count"},
			rows:    [][]driver.Value{{int64(0)}},
		},
		{
			query: `
				SELECT IFNULL(limit_same_weChat, 0)
				FROM inst_config
				WHERE inst_id = ? AND del_flag = 0
				ORDER BY id DESC
				LIMIT 1
			`,
			args:    []any{instID},
			columns: []string{"limit_same_weChat"},
			rows:    [][]driver.Value{{int64(1)}},
		},
		{
			query: `
				SELECT COUNT(*)
				FROM inst_student
				WHERE inst_id = ? AND del_flag = 0 AND wechat_number = ?
			`,
			args:    []any{instID, dto.WeChatNumber},
			columns: []string{"count"},
			rows:    [][]driver.Value{{int64(1)}},
		},
	})
	defer cleanup()

	_, err := svc.AddIntentStudent(userID, dto)
	if err == nil || err.Error() != "当前机构已存在微信号相同的学员" {
		t.Fatalf("expected manual wechat duplicate error, got %v", err)
	}
}

func TestAddIntentStudentByImport_UsesImportDuplicateRule(t *testing.T) {
	userID := int64(102)
	instID := int64(202)
	dto := model.StudentSaveDTO{
		StuName: "李四",
		Mobile:  "13900139000",
	}

	svc, cleanup := newScriptedService(t, []queryExpectation{
		findInstIDExpectation(userID, instID),
		{
			query: `
				SELECT add_import_student_rule
				FROM inst_config
				WHERE inst_id = ? AND del_flag = 0
				ORDER BY id DESC
				LIMIT 1
			`,
			args:    []any{instID},
			columns: []string{"add_import_student_rule"},
			rows:    [][]driver.Value{{int64(2)}},
		},
		{
			query:   `SELECT COUNT(*) FROM inst_student WHERE inst_id = ? AND del_flag = 0 AND mobile = ?`,
			args:    []any{instID, dto.Mobile},
			columns: []string{"count"},
			rows:    [][]driver.Value{{int64(1)}},
		},
	})
	defer cleanup()

	_, err := svc.AddIntentStudentByImport(userID, dto)
	if err == nil || err.Error() != "当前机构已存在手机号相同的学员" {
		t.Fatalf("expected import mobile duplicate error, got %v", err)
	}
}

func TestAddIntentStudentByImport_BlocksDuplicateWeChatWhenImportSettingEnabled(t *testing.T) {
	userID := int64(103)
	instID := int64(203)
	dto := model.StudentSaveDTO{
		StuName:      "王五",
		Mobile:       "13700137000",
		WeChatNumber: "wx-wangwu",
	}

	svc, cleanup := newScriptedService(t, []queryExpectation{
		findInstIDExpectation(userID, instID),
		{
			query: `
				SELECT add_import_student_rule
				FROM inst_config
				WHERE inst_id = ? AND del_flag = 0
				ORDER BY id DESC
				LIMIT 1
			`,
			args:    []any{instID},
			columns: []string{"add_import_student_rule"},
			rows:    [][]driver.Value{{int64(1)}},
		},
		{
			query:   `SELECT COUNT(*) FROM inst_student WHERE inst_id = ? AND del_flag = 0 AND stu_name = ? AND mobile = ?`,
			args:    []any{instID, dto.StuName, dto.Mobile},
			columns: []string{"count"},
			rows:    [][]driver.Value{{int64(0)}},
		},
		{
			query: `
				SELECT IFNULL(limit_import_same_weChat, 0)
				FROM inst_config
				WHERE inst_id = ? AND del_flag = 0
				ORDER BY id DESC
				LIMIT 1
			`,
			args:    []any{instID},
			columns: []string{"limit_import_same_weChat"},
			rows:    [][]driver.Value{{int64(1)}},
		},
		{
			query: `
				SELECT COUNT(*)
				FROM inst_student
				WHERE inst_id = ? AND del_flag = 0 AND wechat_number = ?
			`,
			args:    []any{instID, dto.WeChatNumber},
			columns: []string{"count"},
			rows:    [][]driver.Value{{int64(1)}},
		},
	})
	defer cleanup()

	_, err := svc.AddIntentStudentByImport(userID, dto)
	if err == nil || err.Error() != "当前机构已存在微信号相同的学员" {
		t.Fatalf("expected import wechat duplicate error, got %v", err)
	}
}
