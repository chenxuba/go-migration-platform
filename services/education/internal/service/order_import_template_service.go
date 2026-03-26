package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) BuildLessonHourOrderImportTemplate(userID int64) (string, error) {
	return svc.buildOrderImportTemplate(userID, "按课时", buildLessonHourOrderImportColumns, buildLessonHourOrderImportNotesRichText)
}

func (svc *Service) LoadLessonHourOrderImportTemplate(ticket string) (string, string, []byte, bool) {
	return svc.loadOrderImportTemplate(ticket)
}

func (svc *Service) BuildTimeSlotOrderImportTemplate(userID int64) (string, error) {
	return svc.buildOrderImportTemplate(userID, "按时段", buildTimeSlotOrderImportColumns, buildTimeSlotOrderImportNotesRichText)
}

func (svc *Service) LoadTimeSlotOrderImportTemplate(ticket string) (string, string, []byte, bool) {
	return svc.loadOrderImportTemplate(ticket)
}

func (svc *Service) BuildAmountOrderImportTemplate(userID int64) (string, error) {
	return svc.buildOrderImportTemplate(userID, "按金额", buildAmountOrderImportColumns, buildAmountOrderImportNotesRichText)
}

func (svc *Service) LoadAmountOrderImportTemplate(ticket string) (string, string, []byte, bool) {
	return svc.loadOrderImportTemplate(ticket)
}

type orderImportColumnBuilder func(defaultFields, customFields []model.StudentFieldKey, channels []model.ChannelVO, courseNames, staffNames, orderTagNames []string) []model.IntentionStudentImportTemplateColumn
type orderImportNotesBuilder func(orgName string, columns []model.IntentionStudentImportTemplateColumn) []excelize.RichTextRun

func (svc *Service) buildOrderImportTemplate(userID int64, mode string, columnBuilder orderImportColumnBuilder, notesBuilder orderImportNotesBuilder) (string, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("no institution context")
		}
		return "", err
	}

	orgName, err := svc.repo.GetInstitutionName(context.Background(), instID)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(orgName) == "" {
		orgName = "总校区"
	}

	channels, err := svc.repo.GetChannels(context.Background(), instID)
	if err != nil {
		return "", err
	}
	defaultFields, err := svc.repo.ListStudentFields(context.Background(), instID, true)
	if err != nil {
		return "", err
	}
	customFields, err := svc.repo.ListStudentFields(context.Background(), instID, false)
	if err != nil {
		return "", err
	}
	courseNames, err := svc.loadOrderImportCourseNames(context.Background(), instID, mode)
	if err != nil {
		return "", err
	}
	staffNames, err := svc.repo.ListActiveStaffNames(context.Background(), instID)
	if err != nil {
		return "", err
	}
	orderTagNames, err := svc.repo.ListEnabledOrderTagNames(context.Background(), instID)
	if err != nil {
		return "", err
	}

	columns := columnBuilder(defaultFields, customFields, channels, courseNames, staffNames, orderTagNames)
	filename := sanitizeTemplateFileName(fmt.Sprintf("%s导入学员订单模板-%s-%s.xlsx", orgName, mode, time.Now().Format("20060102")))
	content, err := buildOrderImportTemplateWorkbook(columns, notesBuilder(orgName, columns))
	if err != nil {
		return "", err
	}

	ticket := saveTemplateDownloadFile(templateDownloadFile{
		Filename:    filename,
		ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		Data:        content,
		ExpiresAt:   time.Now().Add(30 * time.Minute),
	})

	switch mode {
	case "按时段":
		return "/api/v1/orders/import-template/time-slot/file?ticket=" + url.QueryEscape(ticket), nil
	case "按金额":
		return "/api/v1/orders/import-template/amount/file?ticket=" + url.QueryEscape(ticket), nil
	default:
		return "/api/v1/orders/import-template/lesson-hour/file?ticket=" + url.QueryEscape(ticket), nil
	}
}

func (svc *Service) loadOrderImportCourseNames(ctx context.Context, instID int64, mode string) ([]string, error) {
	switch strings.TrimSpace(mode) {
	case "按课时":
		return svc.repo.ListCourseNamesByLessonModel(ctx, instID, 1)
	case "按时段":
		return svc.repo.ListCourseNamesByLessonModel(ctx, instID, 2)
	case "按金额":
		return svc.repo.ListCourseNamesByLessonModel(ctx, instID, 3)
	default:
		return svc.repo.ListCourseNamesByLessonModel(ctx, instID, 1)
	}
}

func (svc *Service) loadOrderImportTemplate(ticket string) (string, string, []byte, bool) {
	file, ok := loadTemplateDownloadFile(strings.TrimSpace(ticket))
	if !ok {
		return "", "", nil, false
	}
	return file.Filename, file.ContentType, file.Data, true
}

func buildLessonHourOrderImportColumns(defaultFields, customFields []model.StudentFieldKey, channels []model.ChannelVO, courseNames, staffNames, orderTagNames []string) []model.IntentionStudentImportTemplateColumn {
	columns := buildConfiguredStudentImportColumns(defaultFields, customFields, channels, staffNames, "销售")
	columns = append(columns,
		model.IntentionStudentImportTemplateColumn{Title: "报读课程", Required: true, FieldType: 4, Options: courseNames},
		model.IntentionStudentImportTemplateColumn{Title: "报读班级", FieldType: 1},
		model.IntentionStudentImportTemplateColumn{Title: "购买课时数", Required: true, FieldType: 2},
		model.IntentionStudentImportTemplateColumn{Title: "赠送课时数", FieldType: 2},
		model.IntentionStudentImportTemplateColumn{Title: "已上课时数", FieldType: 2},
		model.IntentionStudentImportTemplateColumn{Title: "有效期至", FieldType: 3},
	)
	return appendCommonOrderColumns(columns, staffNames, orderTagNames)
}

func buildTimeSlotOrderImportColumns(defaultFields, customFields []model.StudentFieldKey, channels []model.ChannelVO, courseNames, staffNames, orderTagNames []string) []model.IntentionStudentImportTemplateColumn {
	columns := buildConfiguredStudentImportColumns(defaultFields, customFields, channels, staffNames, "销售")
	columns = append(columns,
		model.IntentionStudentImportTemplateColumn{Title: "报读课程", Required: true, FieldType: 4, Options: courseNames},
		model.IntentionStudentImportTemplateColumn{Title: "报读班级", FieldType: 1},
		model.IntentionStudentImportTemplateColumn{Title: "有效开始日期", Required: true, FieldType: 3},
		model.IntentionStudentImportTemplateColumn{Title: "有效结束日期(含赠送天数)", Required: true, FieldType: 3},
		model.IntentionStudentImportTemplateColumn{Title: "赠送天数", FieldType: 2},
		model.IntentionStudentImportTemplateColumn{Title: "已上天数", FieldType: 2},
	)
	return appendCommonOrderColumns(columns, staffNames, orderTagNames)
}

func buildAmountOrderImportColumns(defaultFields, customFields []model.StudentFieldKey, channels []model.ChannelVO, courseNames, staffNames, orderTagNames []string) []model.IntentionStudentImportTemplateColumn {
	columns := buildConfiguredStudentImportColumns(defaultFields, customFields, channels, staffNames, "销售")
	columns = append(columns,
		model.IntentionStudentImportTemplateColumn{Title: "报读课程", Required: true, FieldType: 4, Options: courseNames},
		model.IntentionStudentImportTemplateColumn{Title: "报读班级", FieldType: 1},
		model.IntentionStudentImportTemplateColumn{Title: "购买金额", Required: true, FieldType: 2},
		model.IntentionStudentImportTemplateColumn{Title: "赠送金额", FieldType: 2},
		model.IntentionStudentImportTemplateColumn{Title: "已上金额", FieldType: 2},
		model.IntentionStudentImportTemplateColumn{Title: "有效期至", FieldType: 3},
	)
	return appendCommonOrderColumns(columns, staffNames, orderTagNames)
}

func appendCommonOrderColumns(columns []model.IntentionStudentImportTemplateColumn, staffNames, orderTagNames []string) []model.IntentionStudentImportTemplateColumn {
	return append(columns,
		model.IntentionStudentImportTemplateColumn{Title: "实收金额", Required: true, FieldType: 2},
		model.IntentionStudentImportTemplateColumn{Title: "欠费金额", FieldType: 2},
		model.IntentionStudentImportTemplateColumn{Title: "经办日期", FieldType: 3},
		model.IntentionStudentImportTemplateColumn{Title: "收款方式", FieldType: 4, Options: []string{"微信", "支付宝", "银行转账", "POS机", "现金", "其他方式"}},
		model.IntentionStudentImportTemplateColumn{Title: "收款账户", FieldType: 4, Options: []string{"默认账户"}},
		model.IntentionStudentImportTemplateColumn{Title: "支付单号", FieldType: 1},
		model.IntentionStudentImportTemplateColumn{Title: "对方账户", FieldType: 1},
		model.IntentionStudentImportTemplateColumn{Title: "订单备注", FieldType: 1},
		model.IntentionStudentImportTemplateColumn{Title: "订单标签", FieldType: 4, Options: orderTagNames},
		model.IntentionStudentImportTemplateColumn{Title: "是否为体验价", FieldType: 4, Options: []string{"是", "否"}},
		model.IntentionStudentImportTemplateColumn{Title: "订单销售员", FieldType: 4, Options: staffNames},
	)
}

func buildOrderImportTemplateWorkbook(columns []model.IntentionStudentImportTemplateColumn, notes []excelize.RichTextRun) ([]byte, error) {
	file := excelize.NewFile()
	sheetName := file.GetSheetName(0)
	const rowCount = 120

	headerStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   10,
			Family: "Microsoft YaHei",
			Color:  "#222222",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"#D8E5F7"},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return nil, err
	}

	dataCellStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   11,
			Family: "Microsoft YaHei",
			Color:  "#222222",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return nil, err
	}

	noteStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   12,
			Family: "Microsoft YaHei",
			Color:  "#111111",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "top",
			WrapText:   true,
		},
	})
	if err != nil {
		return nil, err
	}

	for idx, column := range columns {
		cell, _ := excelize.CoordinatesToCellName(idx+1, 1)
		col := columnName(idx + 1)
		file.SetColWidth(sheetName, col, col, orderImportTemplateColumnWidth(column.Title))
		if err := file.SetCellRichText(sheetName, cell, buildHeaderRichText(column)); err != nil {
			return nil, err
		}
		if err := file.SetCellStyle(sheetName, cell, cell, headerStyle); err != nil {
			return nil, err
		}
		dataStartCell := fmt.Sprintf("%s%d", col, 2)
		dataEndCell := fmt.Sprintf("%s%d", col, rowCount+1)
		if err := file.SetCellStyle(sheetName, dataStartCell, dataEndCell, dataCellStyle); err != nil {
			return nil, err
		}
		if len(column.Options) > 0 {
			if strings.TrimSpace(column.Title) == "报读课程" {
				if err := addTemplateDropdownValidationBySheetRange(file, sheetName, "template_options", "A", col, 2, rowCount+1, column.Options, !column.Required); err != nil {
					return nil, err
				}
			} else if err := addTemplateDropdownValidation(file, sheetName, col, 2, rowCount+1, column.Options, !column.Required); err != nil {
				return nil, err
			}
		} else if column.FieldType == 3 {
			if err := addTemplateDateValidation(file, sheetName, col, 2, rowCount+1, !column.Required); err != nil {
				return nil, err
			}
		}
	}

	noteCol := len(columns) + 1
	noteHeaderCell, _ := excelize.CoordinatesToCellName(noteCol, 1)
	file.SetColWidth(sheetName, columnName(noteCol), columnName(noteCol), 56)
	if err := file.SetCellValue(sheetName, noteHeaderCell, "填写说明"); err != nil {
		return nil, err
	}
	if err := file.SetCellStyle(sheetName, noteHeaderCell, noteHeaderCell, headerStyle); err != nil {
		return nil, err
	}

	for row := 1; row <= rowCount+1; row++ {
		height := 22.0
		if row == 1 {
			height = 20
		}
		if err := file.SetRowHeight(sheetName, row, height); err != nil {
			return nil, err
		}
	}

	noteStartCell, _ := excelize.CoordinatesToCellName(noteCol, 2)
	noteEndCell, _ := excelize.CoordinatesToCellName(noteCol, rowCount+1)
	if err := file.MergeCell(sheetName, noteStartCell, noteEndCell); err != nil {
		return nil, err
	}
	if err := file.SetCellRichText(sheetName, noteStartCell, notes); err != nil {
		return nil, err
	}
	if err := file.SetCellStyle(sheetName, noteStartCell, noteStartCell, noteStyle); err != nil {
		return nil, err
	}

	buffer, err := file.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func orderImportTemplateColumnWidth(title string) float64 {
	switch strings.TrimSpace(title) {
	case "学员姓名", "手机号", "手机号归属人", "渠道", "性别", "生日", "微信号":
		return 16
	case "年级", "会员号", "推荐人", "就读学校", "兴趣爱好", "家庭住址":
		return 16
	case "报读课程", "报读班级":
		return 22
	case "购买课时数", "赠送课时数", "已上课时数", "有效期至":
		return 16
	case "有效结束日期(含赠送天数)":
		return 22
	case "有效开始日期", "有效结束日期", "赠送天数", "已上天数":
		return 16
	case "购买金额", "赠送金额", "已上金额":
		return 16
	case "实收金额", "欠费金额", "经办日期":
		return 14
	case "收款方式", "收款账户", "支付单号", "对方账户":
		return 16
	case "订单备注", "订单标签", "是否为体验价", "销售", "订单销售员":
		return 16
	default:
		return 16
	}
}

func buildLessonHourOrderImportNotesRichText(orgName string, columns []model.IntentionStudentImportTemplateColumn) []excelize.RichTextRun {
	return buildOrderImportNotesRichText(
		orgName,
		"按课时",
		columns,
		[]string{
			"4、「生日」、「经办日期」、「有效期至」支持 2021-01-21、2021/01/21 等可识别日期格式。",
			"5、「报读课程」请从下拉框选择当前机构内已创建的课程名称。",
			"6、「收款方式」、「收款账户」、「支付单号」、「对方账户」仅在存在实收金额时填写。",
			"7、「订单标签」、「销售」、「订单销售员」如需填写，请使用系统中已存在的名称。",
			"8、「是否为体验价」请填写“是”或“否”。",
			"9、最多导入1000条数据，请控制导入数量。",
		},
	)
}

func buildTimeSlotOrderImportNotesRichText(orgName string, columns []model.IntentionStudentImportTemplateColumn) []excelize.RichTextRun {
	return buildOrderImportNotesRichText(
		orgName,
		"按时段",
		columns,
		[]string{
			"4、「生日」、「经办日期」、「有效开始日期」、「有效结束日期(含赠送天数)」支持 2021-01-21、2021/01/21 等可识别日期格式。",
			"5、「有效开始日期」与「有效结束日期(含赠送天数)」需同时填写，且结束日期不能早于开始日期。",
			"6、「报读课程」请从下拉框选择当前机构内已创建的课程名称。",
			"7、「赠送天数」、「已上天数」请填写数字。",
			"8、「收款方式」、「收款账户」、「支付单号」、「对方账户」仅在存在实收金额时填写。",
			"9、「订单标签」、「销售」、「订单销售员」如需填写，请使用系统中已存在的名称。",
			"10、「是否为体验价」请填写“是”或“否”。",
			"11、最多导入1000条数据，请控制导入数量。",
		},
	)
}

func buildAmountOrderImportNotesRichText(orgName string, columns []model.IntentionStudentImportTemplateColumn) []excelize.RichTextRun {
	return buildOrderImportNotesRichText(
		orgName,
		"按金额",
		columns,
		[]string{
			"4、「生日」、「经办日期」、「有效期至」支持 2021-01-21、2021/01/21 等可识别日期格式。",
			"5、「报读课程」请从下拉框选择当前机构内已创建的课程名称。",
			"6、「购买金额」、「赠送金额」、「已上金额」请填写数字。",
			"7、「收款方式」、「收款账户」、「支付单号」、「对方账户」仅在存在实收金额时填写。",
			"8、「订单标签」、「销售」、「订单销售员」如需填写，请使用系统中已存在的名称。",
			"9、「是否为体验价」请填写“是”或“否”。",
			"10、最多导入1000条数据，请控制导入数量。",
		},
	)
}

func buildOrderImportNotesRichText(orgName, mode string, columns []model.IntentionStudentImportTemplateColumn, extraRules []string) []excelize.RichTextRun {
	requiredFields := make([]string, 0, len(columns))
	optionBlocks := make([]string, 0, len(columns))
	for _, column := range columns {
		if column.Required {
			requiredFields = append(requiredFields, "「"+column.Title+"」")
		}
		if len(column.Options) > 0 {
			optionBlocks = append(optionBlocks, fmt.Sprintf("「%s」：%s", column.Title, strings.Join(column.Options, "、")))
		}
	}

	black := &excelize.Font{Size: 12, Family: "Microsoft YaHei", Color: "#111111"}
	red := &excelize.Font{Size: 12, Family: "Microsoft YaHei", Color: "#FF3B30"}
	title := &excelize.Font{Size: 12, Family: "Microsoft YaHei", Bold: true, Color: "#111111"}

	runs := []excelize.RichTextRun{
		{Text: "导入学员订单-" + mode + "模板，用于向", Font: black},
		{Text: "【" + orgName + "】", Font: red},
		{Text: "导入" + mode + "订单。\n\n", Font: black},
		{Text: "【导入提示】\n", Font: title},
		{Text: "请务必先创建「即将导入的学员」在当前校区内的课程、学杂费或教材商品；如需分班，请先创建班级。\n\n", Font: black},
		{Text: "【填写规范】\n", Font: title},
		{Text: "1、请勿修改顶部字段标题及顺序。\n", Font: black},
	}

	if len(requiredFields) > 0 {
		runs = append(runs,
			excelize.RichTextRun{Text: "2、标*字段，", Font: red},
			excelize.RichTextRun{Text: strings.Join(requiredFields, "、"), Font: red},
			excelize.RichTextRun{Text: "为必填项。\n", Font: black},
		)
	} else {
		runs = append(runs, excelize.RichTextRun{Text: "2、请按字段要求补充完整内容。\n", Font: black})
	}

	runs = append(runs, excelize.RichTextRun{Text: "3、「手机号」必须为 1 开头的 11 位数字，不支持“-”和中间空格，例如 13311113333。\n", Font: black})
	for _, rule := range extraRules {
		runs = append(runs, excelize.RichTextRun{Text: rule + "\n", Font: black})
	}

	if len(optionBlocks) > 0 {
		runs = append(runs, excelize.RichTextRun{Text: "\n【选项说明】\n", Font: title})
		for _, block := range optionBlocks {
			runs = append(runs, excelize.RichTextRun{Text: block + "\n", Font: black})
		}
	}

	runs = append(runs,
		excelize.RichTextRun{Text: "\n【其他注意】\n", Font: title},
		excelize.RichTextRun{Text: "本模板当前用于下载示例与字段规范，请严格按模板填写。", Font: black},
	)
	return runs
}

func addTemplateDropdownValidationBySheetRange(file *excelize.File, sheetName, helperSheetName, helperCol, targetCol string, startRow, endRow int, options []string, allowBlank bool) error {
	if len(options) == 0 {
		return nil
	}

	helperSheetIndex, err := file.GetSheetIndex(helperSheetName)
	if err != nil {
		return err
	}
	if helperSheetIndex == -1 {
		if _, err := file.NewSheet(helperSheetName); err != nil {
			return err
		}
	}

	for idx, option := range options {
		cell := fmt.Sprintf("%s%d", helperCol, idx+1)
		if err := file.SetCellValue(helperSheetName, cell, option); err != nil {
			return err
		}
	}
	if err := file.SetSheetVisible(helperSheetName, false); err != nil {
		return err
	}

	dv := excelize.NewDataValidation(allowBlank)
	dv.Sqref = fmt.Sprintf("%s%d:%s%d", targetCol, startRow, targetCol, endRow)
	dv.SetSqrefDropList(fmt.Sprintf("%s!$%s$1:$%s$%d", helperSheetName, helperCol, helperCol, len(options)))
	dv.ShowErrorMessage = true
	dv.SetError(excelize.DataValidationErrorStyleStop, "填写有误", "请从下拉选项中选择当前机构已创建的课程")
	dv.ShowInputMessage = true
	if allowBlank {
		dv.SetInput("可选项", "请从下拉列表中选择")
	} else {
		dv.SetInput("必填项", "请从下拉列表中选择，不能为空")
	}
	return file.AddDataValidation(sheetName, dv)
}
