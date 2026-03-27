package service

import (
	"testing"

	"go-migration-platform/services/education/internal/model"
)

func containsTemplateColumn(columns []model.IntentionStudentImportTemplateColumn, title string) bool {
	for _, column := range columns {
		if column.Title == title {
			return true
		}
	}
	return false
}

func TestBuildConfiguredStudentImportColumns_UsesDisplayedFieldsOnly(t *testing.T) {
	defaultFields := []model.StudentFieldKey{
		{ID: 1, FieldKey: "渠道", FieldType: 4, IsDisplay: true},
		{ID: 2, FieldKey: "微信号", FieldType: 1, IsDisplay: true},
		{ID: 3, FieldKey: "家庭住址", FieldType: 1, IsDisplay: false},
		{ID: 4, FieldKey: "兴趣爱好", FieldType: 1, IsDisplay: false},
	}
	customFields := []model.StudentFieldKey{
		{ID: 11, FieldKey: "自定义A", FieldType: 1, IsDisplay: true, Sort: 1},
		{ID: 12, FieldKey: "自定义B", FieldType: 1, IsDisplay: false, Sort: 2},
	}
	channels := []model.ChannelVO{
		{ID: 101, Name: "自然到访"},
	}
	staffNames := []string{"销售员甲"}

	columns := buildConfiguredStudentImportColumns(defaultFields, customFields, channels, staffNames, "销售员")

	if !containsTemplateColumn(columns, "学员姓名") {
		t.Fatalf("expected fixed column 学员姓名 to exist")
	}
	if !containsTemplateColumn(columns, "手机号") {
		t.Fatalf("expected fixed column 手机号 to exist")
	}
	if !containsTemplateColumn(columns, "手机号归属人") {
		t.Fatalf("expected fixed column 手机号归属人 to exist")
	}
	if !containsTemplateColumn(columns, "渠道") {
		t.Fatalf("expected displayed default field 渠道 to exist")
	}
	if !containsTemplateColumn(columns, "微信号") {
		t.Fatalf("expected displayed default field 微信号 to exist")
	}
	if containsTemplateColumn(columns, "家庭住址") {
		t.Fatalf("expected hidden default field 家庭住址 to be excluded")
	}
	if containsTemplateColumn(columns, "兴趣爱好") {
		t.Fatalf("expected hidden default field 兴趣爱好 to be excluded")
	}
	if !containsTemplateColumn(columns, "自定义A") {
		t.Fatalf("expected displayed custom field 自定义A to exist")
	}
	if containsTemplateColumn(columns, "自定义B") {
		t.Fatalf("expected hidden custom field 自定义B to be excluded")
	}
	if !containsTemplateColumn(columns, "销售员") {
		t.Fatalf("expected sales column 销售员 to exist")
	}
}
