package service

import (
	"bytes"
	"context"
	"database/sql/driver"
	"strings"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"

	"go-migration-platform/services/education/internal/model"
)

func cellByTitle(row model.IntentionStudentImportRow, title string) model.IntentionStudentImportCell {
	for _, cell := range row.Cells {
		if cell.Title == title {
			return cell
		}
	}
	return model.IntentionStudentImportCell{}
}

func TestParseLessonHourOrderImportFile_AppliesDefaultsAndSkipsEmptyRows(t *testing.T) {
	userID := int64(301)
	instID := int64(401)

	svc, cleanup := newScriptedService(t, []queryExpectation{
		findInstIDExpectation(userID, instID),
		{
			query: `
				SELECT IFNULL(organ_name, '')
				FROM org_institution
				WHERE id = ?
				LIMIT 1
			`,
			args:    []any{instID},
			columns: []string{"organ_name"},
			rows:    [][]driver.Value{{"测试校区"}},
		},
		{
			query: `
				SELECT id, IFNULL(uuid, ''), IFNULL(version, 0), IFNULL(channel_name, ''), IFNULL(introduction, ''), IFNULL(category_id, 0), IFNULL(is_disabled, 0), IFNULL(remark, '')
				FROM inst_channel
				WHERE del_flag = 0 AND (inst_id = ? OR inst_id IS NULL)
				ORDER BY inst_id IS NULL DESC, id DESC
			`,
			args:    []any{instID},
			columns: []string{"id", "uuid", "version", "channel_name", "introduction", "category_id", "is_disabled", "remark"},
			rows:    [][]driver.Value{},
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
				{int64(1), "", int64(0), instID, "渠道", int64(4), int64(0), int64(0), "", int64(1), int64(1), int64(0), int64(1), int64(1), ""},
				{int64(2), "", int64(0), instID, "性别", int64(4), int64(0), int64(0), "", int64(1), int64(1), int64(0), int64(1), int64(2), ""},
				{int64(3), "", int64(0), instID, "微信号", int64(1), int64(0), int64(0), "", int64(1), int64(1), int64(0), int64(1), int64(3), ""},
				{int64(4), "", int64(0), instID, "年级", int64(4), int64(0), int64(0), "一年级,二年级", int64(1), int64(1), int64(0), int64(1), int64(4), ""},
			},
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
			args:    []any{instID, false},
			columns: []string{"id", "uuid", "version", "inst_id", "field_key", "field_type", "required", "searched", "options_json", "is_default", "is_display", "can_delete", "can_edit", "sort", "remark"},
			rows:    [][]driver.Value{},
		},
		{
			query: `
				SELECT IFNULL(nick_name, '') AS nick_name
				FROM inst_user
				WHERE inst_id = ? AND del_flag = 0 AND IFNULL(disabled, 0) = 0
				GROUP BY IFNULL(nick_name, '')
				ORDER BY MAX(create_time) DESC, MAX(id) DESC
			`,
			args:    []any{instID},
			columns: []string{"nick_name"},
			rows:    [][]driver.Value{{"销售甲"}},
		},
		{
			query: `
				SELECT IFNULL(name, '')
				FROM inst_order_tag
				WHERE inst_id = ? AND del_flag = 0 AND IFNULL(enable, 0) = 1
				ORDER BY update_time DESC, id DESC
			`,
			args:    []any{instID},
			columns: []string{"name"},
			rows:    [][]driver.Value{{"老带新"}},
		},
		{
			query: `
				SELECT IFNULL(c.name, '')
				FROM inst_course c
				WHERE c.inst_id = ? AND c.del_flag = 0
				  AND EXISTS (
					SELECT 1
					FROM inst_course_quotation q
					WHERE q.course_id = c.id AND q.del_flag = 0 AND q.lesson_model = ?
				  )
				GROUP BY IFNULL(c.name, '')
				ORDER BY MAX(c.update_time) DESC, MAX(c.id) DESC
			`,
			args:    []any{instID, 1},
			columns: []string{"name"},
			rows:    [][]driver.Value{{"数学课"}},
		},
	})
	defer cleanup()

	file := excelize.NewFile()
	sheet := file.GetSheetName(0)
	headers := []string{"*学员姓名", "*手机号", "*手机号归属人", "微信号", "*报读课程", "*购买课时数", "赠送课时数", "*实收金额", "欠费金额", "经办日期", "收款方式", "收款账户"}
	for idx, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(idx+1, 1)
		file.SetCellValue(sheet, cell, header)
	}
	values := []string{"陈瑞瑞", "19822223333", "爸爸", "wx-001", "数学课", "2", "", "1002", "", "", "", ""}
	for idx, value := range values {
		cell, _ := excelize.CoordinatesToCellName(idx+1, 2)
		file.SetCellValue(sheet, cell, value)
	}
	buffer, err := file.WriteToBuffer()
	if err != nil {
		t.Fatalf("write workbook: %v", err)
	}

	result, err := svc.ParseLessonHourOrderImportFile(userID, "导入.xlsx", bytes.NewReader(buffer.Bytes()))
	if err != nil {
		t.Fatalf("parse lesson hour import: %v", err)
	}
	if result.InstName != "测试校区" {
		t.Fatalf("unexpected inst name: %s", result.InstName)
	}
	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 parsed row, got %d", len(result.Rows))
	}
	row := result.Rows[0]
	if row.HasError {
		t.Fatalf("expected normal row, got error row: %+v", row)
	}
	if got := cellByTitle(row, "经办日期").Value; got != time.Now().Format("2006-01-02") {
		t.Fatalf("expected default business date, got %s", got)
	}
	if got := cellByTitle(row, "收款方式").Value; got != "其他方式" {
		t.Fatalf("expected default pay method, got %s", got)
	}
	if got := cellByTitle(row, "收款账户").Value; got != "默认账户" {
		t.Fatalf("expected default pay account, got %s", got)
	}
}

func TestParseOrderImportFile_SupportsTimeSlotTemplate(t *testing.T) {
	userID := int64(401)
	instID := int64(501)

	svc, cleanup := newScriptedService(t, []queryExpectation{
		findInstIDExpectation(userID, instID),
		{
			query: `
				SELECT IFNULL(organ_name, '')
				FROM org_institution
				WHERE id = ?
				LIMIT 1
			`,
			args:    []any{instID},
			columns: []string{"organ_name"},
			rows:    [][]driver.Value{{"测试校区"}},
		},
		{
			query: `
				SELECT id, IFNULL(uuid, ''), IFNULL(version, 0), IFNULL(channel_name, ''), IFNULL(introduction, ''), IFNULL(category_id, 0), IFNULL(is_disabled, 0), IFNULL(remark, '')
				FROM inst_channel
				WHERE del_flag = 0 AND (inst_id = ? OR inst_id IS NULL)
				ORDER BY inst_id IS NULL DESC, id DESC
			`,
			args:    []any{instID},
			columns: []string{"id", "uuid", "version", "channel_name", "introduction", "category_id", "is_disabled", "remark"},
			rows:    [][]driver.Value{},
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
				{int64(1), "", int64(0), instID, "渠道", int64(4), int64(0), int64(0), "", int64(1), int64(1), int64(0), int64(1), int64(1), ""},
				{int64(2), "", int64(0), instID, "性别", int64(4), int64(0), int64(0), "", int64(1), int64(1), int64(0), int64(1), int64(2), ""},
			},
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
			args:    []any{instID, false},
			columns: []string{"id", "uuid", "version", "inst_id", "field_key", "field_type", "required", "searched", "options_json", "is_default", "is_display", "can_delete", "can_edit", "sort", "remark"},
			rows:    [][]driver.Value{},
		},
		{
			query: `
				SELECT IFNULL(nick_name, '') AS nick_name
				FROM inst_user
				WHERE inst_id = ? AND del_flag = 0 AND IFNULL(disabled, 0) = 0
				GROUP BY IFNULL(nick_name, '')
				ORDER BY MAX(create_time) DESC, MAX(id) DESC
			`,
			args:    []any{instID},
			columns: []string{"nick_name"},
			rows:    [][]driver.Value{{"销售甲"}},
		},
		{
			query: `
				SELECT IFNULL(name, '')
				FROM inst_order_tag
				WHERE inst_id = ? AND del_flag = 0 AND IFNULL(enable, 0) = 1
				ORDER BY update_time DESC, id DESC
			`,
			args:    []any{instID},
			columns: []string{"name"},
			rows:    [][]driver.Value{},
		},
		{
			query: `
				SELECT IFNULL(c.name, '')
				FROM inst_course c
				WHERE c.inst_id = ? AND c.del_flag = 0
				  AND EXISTS (
					SELECT 1
					FROM inst_course_quotation q
					WHERE q.course_id = c.id AND q.del_flag = 0 AND q.lesson_model = ?
				  )
				GROUP BY IFNULL(c.name, '')
				ORDER BY MAX(c.update_time) DESC, MAX(c.id) DESC
			`,
			args:    []any{instID, 2},
			columns: []string{"name"},
			rows:    [][]driver.Value{{"感统课"}},
		},
	})
	defer cleanup()

	file := excelize.NewFile()
	sheet := file.GetSheetName(0)
	headers := []string{"*学员姓名", "*手机号", "*手机号归属人", "*报读课程", "*有效开始日期", "*有效结束日期(含赠送天数)", "赠送天数", "*实收金额", "欠费金额"}
	for idx, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(idx+1, 1)
		file.SetCellValue(sheet, cell, header)
	}
	values := []string{"陈瑞瑞", "19822223333", "爸爸", "感统课", "2026-03-01", "2026-03-31", "1", "3210", "0"}
	for idx, value := range values {
		cell, _ := excelize.CoordinatesToCellName(idx+1, 2)
		file.SetCellValue(sheet, cell, value)
	}
	buffer, err := file.WriteToBuffer()
	if err != nil {
		t.Fatalf("write workbook: %v", err)
	}

	result, err := svc.ParseOrderImportFile(userID, "按时段导入.xlsx", bytes.NewReader(buffer.Bytes()))
	if err != nil {
		t.Fatalf("parse time-slot import: %v", err)
	}
	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 parsed row, got %d", len(result.Rows))
	}
	if result.Rows[0].HasError {
		t.Fatalf("expected normal row, got error row: %+v", result.Rows[0])
	}
}

func TestParseOrderImportFile_SupportsAmountTemplate(t *testing.T) {
	userID := int64(402)
	instID := int64(502)

	svc, cleanup := newScriptedService(t, []queryExpectation{
		findInstIDExpectation(userID, instID),
		{
			query: `
				SELECT IFNULL(organ_name, '')
				FROM org_institution
				WHERE id = ?
				LIMIT 1
			`,
			args:    []any{instID},
			columns: []string{"organ_name"},
			rows:    [][]driver.Value{{"测试校区"}},
		},
		{
			query: `
				SELECT id, IFNULL(uuid, ''), IFNULL(version, 0), IFNULL(channel_name, ''), IFNULL(introduction, ''), IFNULL(category_id, 0), IFNULL(is_disabled, 0), IFNULL(remark, '')
				FROM inst_channel
				WHERE del_flag = 0 AND (inst_id = ? OR inst_id IS NULL)
				ORDER BY inst_id IS NULL DESC, id DESC
			`,
			args:    []any{instID},
			columns: []string{"id", "uuid", "version", "channel_name", "introduction", "category_id", "is_disabled", "remark"},
			rows:    [][]driver.Value{},
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
				{int64(1), "", int64(0), instID, "渠道", int64(4), int64(0), int64(0), "", int64(1), int64(1), int64(0), int64(1), int64(1), ""},
				{int64(2), "", int64(0), instID, "性别", int64(4), int64(0), int64(0), "", int64(1), int64(1), int64(0), int64(1), int64(2), ""},
			},
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
			args:    []any{instID, false},
			columns: []string{"id", "uuid", "version", "inst_id", "field_key", "field_type", "required", "searched", "options_json", "is_default", "is_display", "can_delete", "can_edit", "sort", "remark"},
			rows:    [][]driver.Value{},
		},
		{
			query: `
				SELECT IFNULL(nick_name, '') AS nick_name
				FROM inst_user
				WHERE inst_id = ? AND del_flag = 0 AND IFNULL(disabled, 0) = 0
				GROUP BY IFNULL(nick_name, '')
				ORDER BY MAX(create_time) DESC, MAX(id) DESC
			`,
			args:    []any{instID},
			columns: []string{"nick_name"},
			rows:    [][]driver.Value{{"销售甲"}},
		},
		{
			query: `
				SELECT IFNULL(name, '')
				FROM inst_order_tag
				WHERE inst_id = ? AND del_flag = 0 AND IFNULL(enable, 0) = 1
				ORDER BY update_time DESC, id DESC
			`,
			args:    []any{instID},
			columns: []string{"name"},
			rows:    [][]driver.Value{},
		},
		{
			query: `
				SELECT IFNULL(c.name, '')
				FROM inst_course c
				WHERE c.inst_id = ? AND c.del_flag = 0
				  AND EXISTS (
					SELECT 1
					FROM inst_course_quotation q
					WHERE q.course_id = c.id AND q.del_flag = 0 AND q.lesson_model = ?
				  )
				GROUP BY IFNULL(c.name, '')
				ORDER BY MAX(c.update_time) DESC, MAX(c.id) DESC
			`,
			args:    []any{instID, 3},
			columns: []string{"name"},
			rows:    [][]driver.Value{{"托管费"}},
		},
	})
	defer cleanup()

	file := excelize.NewFile()
	sheet := file.GetSheetName(0)
	headers := []string{"*学员姓名", "*手机号", "*手机号归属人", "*报读课程", "*购买金额", "赠送金额", "*实收金额", "欠费金额", "有效期至"}
	for idx, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(idx+1, 1)
		file.SetCellValue(sheet, cell, header)
	}
	values := []string{"陈瑞瑞", "19822223333", "爸爸", "托管费", "5000", "300", "4500", "0", "2026-12-31"}
	for idx, value := range values {
		cell, _ := excelize.CoordinatesToCellName(idx+1, 2)
		file.SetCellValue(sheet, cell, value)
	}
	buffer, err := file.WriteToBuffer()
	if err != nil {
		t.Fatalf("write workbook: %v", err)
	}

	result, err := svc.ParseOrderImportFile(userID, "按金额导入.xlsx", bytes.NewReader(buffer.Bytes()))
	if err != nil {
		t.Fatalf("parse amount import: %v", err)
	}
	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 parsed row, got %d", len(result.Rows))
	}
	if result.Rows[0].HasError {
		t.Fatalf("expected normal row, got error row: %+v", result.Rows[0])
	}
}

func TestBuildCreateAndPayOrderDTOFromImportRow_UsesCustomQuoteForImportedOrder(t *testing.T) {
	row := model.IntentionStudentImportRow{
		Cells: []model.IntentionStudentImportCell{
			{Title: "报读课程", Value: "数学课"},
			{Title: "购买课时数", Value: "2"},
			{Title: "赠送课时数", Value: "0"},
			{Title: "实收金额", Value: "1002"},
			{Title: "欠费金额", Value: "0"},
			{Title: "收款方式", Value: "其他方式"},
			{Title: "订单备注", Value: "测试导入"},
			{Title: "是否为体验价", Value: "否"},
		},
	}
	quotationMap := map[string][]model.CourseQuotation{
		"数学课": {
			{
				ID:             11,
				CourseID:       22,
				LessonModel:    intPtr(1),
				Unit:           intPtr(1),
				Quantity:       intPtr(10),
				Price:          300,
				LessonAudition: false,
			},
		},
	}

	createDTO, payDTO, hasPayment, err := buildCreateAndPayOrderDTOFromImportRow(
		1001,
		orderImportModeLessonHour,
		1,
		row,
		map[string]model.IntentionStudentImportColumn{},
		map[string][]importOptionItem{
			"收款方式": {
				{Label: "其他方式", Value: "6"},
			},
		},
		map[string]int64{},
		quotationMap,
	)
	if err != nil {
		t.Fatalf("build create/pay dto: %v", err)
	}
	if !hasPayment {
		t.Fatalf("expected payment to exist")
	}
	if createDTO.StudentID != 1001 {
		t.Fatalf("unexpected student id: %d", createDTO.StudentID)
	}
	if len(createDTO.OrderDetail.QuoteDetailList) != 1 {
		t.Fatalf("expected 1 quote detail, got %d", len(createDTO.OrderDetail.QuoteDetailList))
	}
	detail := createDTO.OrderDetail.QuoteDetailList[0]
	if detail.Count == nil || *detail.Count != 1 {
		t.Fatalf("expected imported order sku count to be 1, got %+v", detail.Count)
	}
	if detail.Quantity != 2 {
		t.Fatalf("expected purchased quantity 2, got %.2f", detail.Quantity)
	}
	if detail.RealQuantity != 2 {
		t.Fatalf("expected real quantity 2, got %.2f", detail.RealQuantity)
	}
	if detail.Amount != "1002.00" || detail.RealAmount != "1002.00" {
		t.Fatalf("expected custom amount 1002.00, got amount=%s realAmount=%s", detail.Amount, detail.RealAmount)
	}
	if createDTO.OrderDetail.OrderSource == nil || *createDTO.OrderDetail.OrderSource != model.OrderSourceOfflineImport {
		t.Fatalf("expected offline import order source, got %+v", createDTO.OrderDetail.OrderSource)
	}
	if payDTO.PayAmount != 1002 {
		t.Fatalf("expected pay amount 1002, got %.2f", payDTO.PayAmount)
	}
	if len(payDTO.PayAccounts) != 1 || payDTO.PayAccounts[0].PayMethod == nil || *payDTO.PayAccounts[0].PayMethod != 6 {
		t.Fatalf("expected pay method 6, got %+v", payDTO.PayAccounts)
	}
}

func TestBuildCreateAndPayOrderDTOFromImportRow_SupportsAmountImport(t *testing.T) {
	row := model.IntentionStudentImportRow{
		Cells: []model.IntentionStudentImportCell{
			{Title: "报读课程", Value: "托管费"},
			{Title: "购买金额", Value: "5000"},
			{Title: "赠送金额", Value: "300"},
			{Title: "有效期至", Value: "2026-12-31"},
			{Title: "实收金额", Value: "4500"},
			{Title: "欠费金额", Value: "0"},
			{Title: "收款方式", Value: "其他方式"},
			{Title: "订单备注", Value: "测试按金额导入"},
			{Title: "是否为体验价", Value: "否"},
		},
	}
	quotationMap := map[string][]model.CourseQuotation{
		"托管费": {
			{
				ID:             31,
				CourseID:       41,
				LessonModel:    intPtr(3),
				Unit:           intPtr(1),
				Quantity:       intPtr(1),
				Price:          5000,
				LessonAudition: false,
			},
		},
	}

	createDTO, payDTO, hasPayment, err := buildCreateAndPayOrderDTOFromImportRow(
		1003,
		orderImportModeAmount,
		1,
		row,
		map[string]model.IntentionStudentImportColumn{},
		map[string][]importOptionItem{
			"收款方式": {
				{Label: "其他方式", Value: "6"},
			},
		},
		map[string]int64{},
		quotationMap,
	)
	if err != nil {
		t.Fatalf("build amount create/pay dto: %v", err)
	}
	if !hasPayment {
		t.Fatalf("expected payment to exist")
	}
	detail := createDTO.OrderDetail.QuoteDetailList[0]
	if detail.LessonMode == nil || *detail.LessonMode != 3 {
		t.Fatalf("expected lesson mode 3, got %+v", detail.LessonMode)
	}
	if detail.Quantity != 5000 {
		t.Fatalf("expected purchased amount quantity 5000, got %.2f", detail.Quantity)
	}
	if detail.FreeQuantity != 300 {
		t.Fatalf("expected free amount 300, got %.2f", detail.FreeQuantity)
	}
	if detail.RealQuantity != 5300 {
		t.Fatalf("expected real quantity 5300, got %.2f", detail.RealQuantity)
	}
	if detail.Amount != "5000.00" {
		t.Fatalf("expected amount 5000.00, got %s", detail.Amount)
	}
	if detail.EndDate == nil || detail.EndDate.Format("2006-01-02") != "2026-12-31" {
		t.Fatalf("unexpected end date: %+v", detail.EndDate)
	}
	if createDTO.OrderDetail.OrderRealAmount != "4500.00" {
		t.Fatalf("expected order real amount 4500.00, got %s", createDTO.OrderDetail.OrderRealAmount)
	}
	if payDTO.PayAmount != 4500 {
		t.Fatalf("expected pay amount 4500, got %.2f", payDTO.PayAmount)
	}
}

func TestBuildCreateAndPayOrderDTOFromImportRow_SupportsTimeSlotImport(t *testing.T) {
	row := model.IntentionStudentImportRow{
		Cells: []model.IntentionStudentImportCell{
			{Title: "报读课程", Value: "感统课"},
			{Title: "有效开始日期", Value: "2026-03-01"},
			{Title: "有效结束日期(含赠送天数)", Value: "2026-03-31"},
			{Title: "赠送天数", Value: "1"},
			{Title: "已上天数", Value: "0"},
			{Title: "实收金额", Value: "3210"},
			{Title: "欠费金额", Value: "0"},
			{Title: "收款方式", Value: "其他方式"},
			{Title: "订单备注", Value: "测试按时段导入"},
			{Title: "是否为体验价", Value: "否"},
		},
	}
	quotationMap := map[string][]model.CourseQuotation{
		"感统课": {
			{
				ID:             21,
				CourseID:       31,
				LessonModel:    intPtr(2),
				Unit:           intPtr(2),
				Quantity:       intPtr(30),
				Price:          3200,
				LessonAudition: false,
			},
		},
	}

	createDTO, payDTO, hasPayment, err := buildCreateAndPayOrderDTOFromImportRow(
		1002,
		orderImportModeTimeSlot,
		1,
		row,
		map[string]model.IntentionStudentImportColumn{},
		map[string][]importOptionItem{
			"收款方式": {
				{Label: "其他方式", Value: "6"},
			},
		},
		map[string]int64{},
		quotationMap,
	)
	if err != nil {
		t.Fatalf("build time-slot create/pay dto: %v", err)
	}
	if !hasPayment {
		t.Fatalf("expected payment to exist")
	}
	detail := createDTO.OrderDetail.QuoteDetailList[0]
	if detail.Quantity != 30 {
		t.Fatalf("expected purchased quantity 30, got %.2f", detail.Quantity)
	}
	if detail.FreeQuantity != 1 {
		t.Fatalf("expected free quantity 1, got %.2f", detail.FreeQuantity)
	}
	if detail.RealQuantity != 31 {
		t.Fatalf("expected real quantity 31, got %.2f", detail.RealQuantity)
	}
	if detail.ValidDate == nil || detail.ValidDate.Format("2006-01-02") != "2026-03-01" {
		t.Fatalf("unexpected valid date: %+v", detail.ValidDate)
	}
	if detail.EndDate == nil || detail.EndDate.Format("2006-01-02") != "2026-03-31" {
		t.Fatalf("unexpected end date: %+v", detail.EndDate)
	}
	if detail.LessonMode == nil || *detail.LessonMode != 2 {
		t.Fatalf("expected lesson mode 2, got %+v", detail.LessonMode)
	}
	if detail.Amount != "3210.00" || detail.RealAmount != "3210.00" {
		t.Fatalf("expected custom amount 3210.00, got amount=%s realAmount=%s", detail.Amount, detail.RealAmount)
	}
	if payDTO.PayAmount != 3210 {
		t.Fatalf("expected pay amount 3210, got %.2f", payDTO.PayAmount)
	}
}

func TestResolveOrderImportStudent_ReusesExistingStudentOnExactNameAndMobile(t *testing.T) {
	userID := int64(302)
	instID := int64(402)
	dto := model.StudentSaveDTO{
		StuName: "陈瑞瑞",
		Mobile:  "19822223333",
	}

	svc, cleanup := newScriptedService(t, []queryExpectation{
		{
			query: `
				SELECT id
				FROM inst_student
				WHERE inst_id = ? AND del_flag = 0 AND stu_name = ? AND mobile = ?
				ORDER BY id ASC
				LIMIT 1
			`,
			args:    []any{instID, dto.StuName, dto.Mobile},
			columns: []string{"id"},
			rows:    [][]driver.Value{{int64(888)}},
		},
	})
	defer cleanup()

	decision, err := svc.resolveOrderImportStudent(userID, instID, dto)
	if err != nil {
		t.Fatalf("resolve student: %v", err)
	}
	if decision.CreatedNew {
		t.Fatalf("expected existing student reuse")
	}
	if decision.StudentID != 888 {
		t.Fatalf("expected reused student 888, got %d", decision.StudentID)
	}
}

func TestDetectOrderImportHandleType_ReturnsNewWhenCourseNotPurchased(t *testing.T) {
	instID := int64(601)
	studentID := int64(701)
	courseID := int64(801)

	svc, cleanup := newScriptedService(t, []queryExpectation{
		{
			query: `
				SELECT COUNT(*)
				FROM sale_order so
				INNER JOIN sale_order_course_detail d ON d.order_id = so.id AND d.del_flag = 0
				WHERE so.inst_id = ? AND so.student_id = ? AND so.order_status = ? AND so.del_flag = 0
				  AND d.course_id = ?
				  AND (
					IFNULL(d.has_valid_date, 0) = 0
					OR d.end_date IS NULL
					OR d.end_date >= CURDATE()
				  )
			`,
			args:    []any{instID, studentID, model.OrderStatusCompleted, courseID},
			columns: []string{"count"},
			rows:    [][]driver.Value{{int64(0)}},
		},
		{
			query: `
				SELECT COUNT(*)
				FROM sale_order so
				INNER JOIN sale_order_course_detail d ON d.order_id = so.id AND d.del_flag = 0
				WHERE so.inst_id = ? AND so.student_id = ? AND so.order_status = ? AND so.del_flag = 0
				  AND d.course_id = ?
			`,
			args:    []any{instID, studentID, model.OrderStatusCompleted, courseID},
			columns: []string{"count"},
			rows:    [][]driver.Value{{int64(0)}},
		},
		{
			query: `
				SELECT COUNT(*)
				FROM sale_order
				WHERE inst_id = ? AND student_id = ? AND order_status = ? AND del_flag = 0
			`,
			args:    []any{instID, studentID, model.OrderStatusCompleted},
			columns: []string{"count"},
			rows:    [][]driver.Value{{int64(0)}},
		},
	})
	defer cleanup()

	handleType, err := svc.detectOrderImportHandleType(context.Background(), instID, studentID, courseID)
	if err != nil {
		t.Fatalf("detect handle type: %v", err)
	}
	if handleType != 1 {
		t.Fatalf("expected new enroll type 1, got %d", handleType)
	}
}

func TestDetectOrderImportHandleType_ReturnsRenewWhenCourseActive(t *testing.T) {
	instID := int64(602)
	studentID := int64(702)
	courseID := int64(802)

	svc, cleanup := newScriptedService(t, []queryExpectation{
		{
			query: `
				SELECT COUNT(*)
				FROM sale_order so
				INNER JOIN sale_order_course_detail d ON d.order_id = so.id AND d.del_flag = 0
				WHERE so.inst_id = ? AND so.student_id = ? AND so.order_status = ? AND so.del_flag = 0
				  AND d.course_id = ?
				  AND (
					IFNULL(d.has_valid_date, 0) = 0
					OR d.end_date IS NULL
					OR d.end_date >= CURDATE()
				  )
			`,
			args:    []any{instID, studentID, model.OrderStatusCompleted, courseID},
			columns: []string{"count"},
			rows:    [][]driver.Value{{int64(1)}},
		},
	})
	defer cleanup()

	handleType, err := svc.detectOrderImportHandleType(context.Background(), instID, studentID, courseID)
	if err != nil {
		t.Fatalf("detect handle type: %v", err)
	}
	if handleType != 2 {
		t.Fatalf("expected renew enroll type 2, got %d", handleType)
	}
}

func TestDetectOrderImportHandleType_ReturnsRenewWhenCoursePurchasedBefore(t *testing.T) {
	instID := int64(603)
	studentID := int64(703)
	courseID := int64(803)

	svc, cleanup := newScriptedService(t, []queryExpectation{
		{
			query: `
				SELECT COUNT(*)
				FROM sale_order so
				INNER JOIN sale_order_course_detail d ON d.order_id = so.id AND d.del_flag = 0
				WHERE so.inst_id = ? AND so.student_id = ? AND so.order_status = ? AND so.del_flag = 0
				  AND d.course_id = ?
				  AND (
					IFNULL(d.has_valid_date, 0) = 0
					OR d.end_date IS NULL
					OR d.end_date >= CURDATE()
				  )
			`,
			args:    []any{instID, studentID, model.OrderStatusCompleted, courseID},
			columns: []string{"count"},
			rows:    [][]driver.Value{{int64(0)}},
		},
		{
			query: `
				SELECT COUNT(*)
				FROM sale_order so
				INNER JOIN sale_order_course_detail d ON d.order_id = so.id AND d.del_flag = 0
				WHERE so.inst_id = ? AND so.student_id = ? AND so.order_status = ? AND so.del_flag = 0
				  AND d.course_id = ?
			`,
			args:    []any{instID, studentID, model.OrderStatusCompleted, courseID},
			columns: []string{"count"},
			rows:    [][]driver.Value{{int64(1)}},
		},
	})
	defer cleanup()

	handleType, err := svc.detectOrderImportHandleType(context.Background(), instID, studentID, courseID)
	if err != nil {
		t.Fatalf("detect handle type: %v", err)
	}
	if handleType != 2 {
		t.Fatalf("expected renew enroll type 2, got %d", handleType)
	}
}

func TestDetectOrderImportHandleType_ReturnsExpandWhenOtherCoursePurchased(t *testing.T) {
	instID := int64(604)
	studentID := int64(704)
	courseID := int64(804)

	svc, cleanup := newScriptedService(t, []queryExpectation{
		{
			query: `
				SELECT COUNT(*)
				FROM sale_order so
				INNER JOIN sale_order_course_detail d ON d.order_id = so.id AND d.del_flag = 0
				WHERE so.inst_id = ? AND so.student_id = ? AND so.order_status = ? AND so.del_flag = 0
				  AND d.course_id = ?
				  AND (
					IFNULL(d.has_valid_date, 0) = 0
					OR d.end_date IS NULL
					OR d.end_date >= CURDATE()
				  )
			`,
			args:    []any{instID, studentID, model.OrderStatusCompleted, courseID},
			columns: []string{"count"},
			rows:    [][]driver.Value{{int64(0)}},
		},
		{
			query: `
				SELECT COUNT(*)
				FROM sale_order so
				INNER JOIN sale_order_course_detail d ON d.order_id = so.id AND d.del_flag = 0
				WHERE so.inst_id = ? AND so.student_id = ? AND so.order_status = ? AND so.del_flag = 0
				  AND d.course_id = ?
			`,
			args:    []any{instID, studentID, model.OrderStatusCompleted, courseID},
			columns: []string{"count"},
			rows:    [][]driver.Value{{int64(0)}},
		},
		{
			query: `
				SELECT COUNT(*)
				FROM sale_order
				WHERE inst_id = ? AND student_id = ? AND order_status = ? AND del_flag = 0
			`,
			args:    []any{instID, studentID, model.OrderStatusCompleted},
			columns: []string{"count"},
			rows:    [][]driver.Value{{int64(2)}},
		},
	})
	defer cleanup()

	handleType, err := svc.detectOrderImportHandleType(context.Background(), instID, studentID, courseID)
	if err != nil {
		t.Fatalf("detect handle type: %v", err)
	}
	if handleType != 3 {
		t.Fatalf("expected expand enroll type 3, got %d", handleType)
	}
}

func TestValidateOrderImportValue_RejectsOldOtherPayMethodAlias(t *testing.T) {
	column := model.IntentionStudentImportColumn{
		Title:     "收款方式",
		FieldType: 4,
		Options:   []string{"微信", "支付宝", "银行转账", "POS机", "现金", "其他方式"},
	}
	if got := validateOrderImportValue(column, "其他"); got != "请选择预设值" {
		t.Fatalf("expected preset value error, got %q", got)
	}
}

func TestBuildCreateAndPayOrderDTOFromImportRow_RequiresCourseName(t *testing.T) {
	_, _, _, err := buildCreateAndPayOrderDTOFromImportRow(
		1001,
		orderImportModeLessonHour,
		1,
		model.IntentionStudentImportRow{
			Cells: []model.IntentionStudentImportCell{
				{Title: "购买课时数", Value: "2"},
				{Title: "实收金额", Value: "100"},
			},
		},
		map[string]model.IntentionStudentImportColumn{},
		map[string][]importOptionItem{},
		map[string]int64{},
		map[string][]model.CourseQuotation{},
	)
	if err == nil || !strings.Contains(err.Error(), "报读课程不能为空") {
		t.Fatalf("expected missing course error, got %v", err)
	}
}
