package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) BuildIntentionStudentImportTemplate(userID int64) (string, error) {
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

	defaultFields, err := svc.repo.ListStudentFields(context.Background(), instID, true)
	if err != nil {
		return "", err
	}
	customFields, err := svc.repo.ListStudentFields(context.Background(), instID, false)
	if err != nil {
		return "", err
	}
	channels, err := svc.repo.GetChannels(context.Background(), instID)
	if err != nil {
		return "", err
	}
	staffNames, err := svc.repo.ListActiveStaffNames(context.Background(), instID)
	if err != nil {
		return "", err
	}

	columns := buildIntentionStudentImportColumns(defaultFields, customFields, channels, staffNames)
	filename := sanitizeTemplateFileName(fmt.Sprintf("%s导入意向学员模板-%s.xlsx", orgName, time.Now().Format("20060102")))
	content, err := buildIntentionStudentImportTemplateWorkbook(orgName, columns)
	if err != nil {
		return "", err
	}
	ticket := saveTemplateDownloadFile(templateDownloadFile{
		Filename:    filename,
		ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		Data:        content,
		ExpiresAt:   time.Now().Add(30 * time.Minute),
	})
	return "/api/v1/intent-students/import-template/file?ticket=" + url.QueryEscape(ticket), nil
}

func (svc *Service) LoadIntentionStudentImportTemplate(ticket string) (string, string, []byte, bool) {
	file, ok := loadTemplateDownloadFile(strings.TrimSpace(ticket))
	if !ok {
		return "", "", nil, false
	}
	return file.Filename, file.ContentType, file.Data, true
}

func buildIntentionStudentImportColumns(defaultFields, customFields []model.StudentFieldKey, channels []model.ChannelVO, staffNames []string) []model.IntentionStudentImportTemplateColumn {
	return buildConfiguredStudentImportColumns(defaultFields, customFields, channels, staffNames, "销售员")
}

func buildConfiguredStudentImportColumns(defaultFields, customFields []model.StudentFieldKey, channels []model.ChannelVO, staffNames []string, salesTitle string) []model.IntentionStudentImportTemplateColumn {
	displayedDefaults := make(map[string]model.StudentFieldKey, len(defaultFields))
	for _, field := range defaultFields {
		if field.IsDisplay {
			displayedDefaults[strings.TrimSpace(field.FieldKey)] = field
		}
	}

	channelOptions := make([]string, 0, len(channels))
	for _, item := range channels {
		if item.IsDisabled || strings.TrimSpace(item.Name) == "" {
			continue
		}
		channelOptions = append(channelOptions, strings.TrimSpace(item.Name))
	}
	sort.Strings(channelOptions)

	columns := []model.IntentionStudentImportTemplateColumn{
		{Title: "学员姓名", Required: true, FieldType: 1},
		{Title: "手机号", Required: true, FieldType: 1},
		{Title: "手机号归属人", Required: true, FieldType: 4, Options: []string{"爸爸", "妈妈", "爷爷", "奶奶", "外公", "外婆", "其他"}},
	}

	appendDefaultColumn := func(title string, fieldType int, fallbackOptions []string) {
		field, ok := displayedDefaults[title]
		if !ok {
			return
		}
		options := fallbackOptions
		if ok && len(options) == 0 && strings.TrimSpace(field.OptionsJSON) != "" {
			options = splitTemplateOptions(field.OptionsJSON)
		}
		columns = append(columns, model.IntentionStudentImportTemplateColumn{
			FieldID: func() int64 {
				return field.ID
			}(),
			Title: title,
			Required: func() bool {
				return field.Required
			}(),
			FieldType: fieldType,
			Options:   options,
		})
	}

	appendDefaultColumn("渠道", 4, channelOptions)
	appendDefaultColumn("性别", 4, []string{"男", "女", "未知"})
	appendDefaultColumn("生日", 3, nil)
	appendDefaultColumn("微信号", 1, nil)
	appendDefaultColumn("年级", 4, nil)
	appendDefaultColumn("就读学校", 1, nil)
	appendDefaultColumn("家庭住址", 1, nil)
	appendDefaultColumn("兴趣爱好", 1, nil)

	sort.SliceStable(customFields, func(i, j int) bool {
		if customFields[i].Sort != customFields[j].Sort {
			return customFields[i].Sort < customFields[j].Sort
		}
		return customFields[i].ID < customFields[j].ID
	})
	for _, field := range customFields {
		if !field.IsDisplay || strings.TrimSpace(field.FieldKey) == "" {
			continue
		}
		column := model.IntentionStudentImportTemplateColumn{
			FieldID:   field.ID,
			Title:     strings.TrimSpace(field.FieldKey),
			Required:  field.Required,
			FieldType: field.FieldType,
		}
		if field.FieldType == 4 {
			column.Options = splitTemplateOptions(field.OptionsJSON)
		}
		columns = append(columns, column)
	}

	if strings.TrimSpace(salesTitle) != "" {
		columns = append(columns, model.IntentionStudentImportTemplateColumn{
			Title:     strings.TrimSpace(salesTitle),
			Required:  false,
			FieldType: 4,
			Options:   staffNames,
		})
	}

	return columns
}

func splitTemplateOptions(raw string) []string {
	parts := strings.Split(strings.TrimSpace(raw), ",")
	result := make([]string, 0, len(parts))
	for _, item := range parts {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		result = append(result, item)
	}
	return result
}

func buildIntentionStudentImportTemplateWorkbook(orgName string, columns []model.IntentionStudentImportTemplateColumn) ([]byte, error) {
	file := excelize.NewFile()
	sheetName := file.GetSheetName(0)
	const rowCount = 120

	headerStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			// Header font size for the first row of the template.
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
		file.SetColWidth(sheetName, col, col, intentionStudentTemplateColumnWidth(column))
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
			if err := addTemplateDropdownValidation(file, sheetName, col, 2, rowCount+1, column.Options, !column.Required); err != nil {
				return nil, err
			}
		} else if column.FieldType == 3 || strings.TrimSpace(column.Title) == "生日" {
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
		// Row height tuning:
		// row 1 is the table header row, the rest are template data rows.
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
	if err := file.SetCellRichText(sheetName, noteStartCell, buildIntentionStudentImportNotesRichText(orgName, columns)); err != nil {
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

func buildIntentionStudentImportNotesRichText(orgName string, columns []model.IntentionStudentImportTemplateColumn) []excelize.RichTextRun {
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
		{Text: "你好，当前 excel 表格学员数据用于导入", Font: black},
		{Text: "【" + orgName + "】", Font: red},
		{Text: "使用\n\n", Font: black},
		{Text: "【导入提示】\n", Font: title},
		{Text: "「来源渠道」、「销售员」取的是系统内的预设信息，如填写信息不符合预设值，则导入会失败。\n\n", Font: black},
		{Text: "【填写规范】\n", Font: title},
		{Text: "1、请勿修改顶部字段标题及顺序\n", Font: black},
	}

	if len(requiredFields) > 0 {
		runs = append(runs,
			excelize.RichTextRun{Text: "2、标*字段，", Font: red},
			excelize.RichTextRun{Text: strings.Join(requiredFields, "、"), Font: red},
			excelize.RichTextRun{Text: "为必填项，", Font: black},
			excelize.RichTextRun{Text: "「来源渠道」、「销售员」、下拉选项字段", Font: red},
			excelize.RichTextRun{Text: "请按预设值填写\n", Font: black},
		)
	} else {
		runs = append(runs, excelize.RichTextRun{Text: "2、请按字段要求补充完整内容\n", Font: black})
	}

	runs = append(runs,
		excelize.RichTextRun{Text: "3、「手机号」必须为 1 开头的 11 位数字，不支持“-”和中间空格，", Font: black},
		excelize.RichTextRun{Text: "支持样式 13311113333\n", Font: red},
		excelize.RichTextRun{Text: "4、「出生日期」的日期格式支持年月日输入，支持", Font: black},
		excelize.RichTextRun{Text: "2021-01-21、2021/01/21", Font: red},
		excelize.RichTextRun{Text: "四种样式\n", Font: black},
		excelize.RichTextRun{Text: "5、自定义添加的字段请按照对应的格式类型进行填写，如「数字」的格式类型，只支持输入阿拉伯数字，请勿携带单位“节”或“元”\n", Font: black},
		excelize.RichTextRun{Text: "6、如更新了自定义字段的字段名称、必填规则，则需重新下载模板，填写信息后再进行导入\n", Font: black},
	)

	if len(optionBlocks) > 0 {
		runs = append(runs, excelize.RichTextRun{Text: "\n【选项说明】\n", Font: title})
		for _, block := range optionBlocks {
			runs = append(runs, excelize.RichTextRun{Text: block + "\n", Font: black})
		}
	}

	runs = append(runs,
		excelize.RichTextRun{Text: "\n【其他注意】\n", Font: title},
		excelize.RichTextRun{Text: "最多导入1000条数据，请控制导入数量。", Font: black},
	)
	return runs
}

// intentionStudentTemplateColumnWidth controls the visible Excel column width
// for each template field. The value is Excel's native column width unit,
// not pixels. Increase the returned number to make a column wider.
func intentionStudentTemplateColumnWidth(column model.IntentionStudentImportTemplateColumn) float64 {
	title := strings.TrimSpace(column.Title)
	switch title {
	case "学员姓名":
		return 16
	case "手机号", "手机号码":
		return 16
	case "手机号归属人":
		return 16
	case "渠道", "性别", "生日", "年级", "销售员":
		return 16
	case "微信号":
		return 16
	case "就读学校", "家庭住址", "兴趣爱好":
		return 16
	case "残疾证号", "身份证号":
		return 16
	default:
		switch column.FieldType {
		case 2, 3, 4:
			return 16
		default:
			return 16
		}
	}
}

func buildHeaderRichText(column model.IntentionStudentImportTemplateColumn) []excelize.RichTextRun {
	if !column.Required {
		return []excelize.RichTextRun{{Text: column.Title, Font: &excelize.Font{Bold: true, Size: 10, Family: "Microsoft YaHei", Color: "#222222"}}}
	}
	return []excelize.RichTextRun{
		{Text: "*", Font: &excelize.Font{Bold: true, Size: 10, Family: "Microsoft YaHei", Color: "#FF4D4F"}},
		{Text: column.Title, Font: &excelize.Font{Bold: true, Size: 10, Family: "Microsoft YaHei", Color: "#222222"}},
	}
}

func columnName(index int) string {
	name, _ := excelize.ColumnNumberToName(index)
	return name
}

func addTemplateDropdownValidation(file *excelize.File, sheetName, col string, startRow, endRow int, options []string, allowBlank bool) error {
	if len(options) == 0 {
		return nil
	}
	optionText := strings.Join(options, ",")
	if len(optionText) > 255 {
		// Excel inline list validation has a 255-char limit.
		return nil
	}
	dv := excelize.NewDataValidation(allowBlank)
	dv.Sqref = fmt.Sprintf("%s%d:%s%d", col, startRow, col, endRow)
	if err := dv.SetDropList(options); err != nil {
		return err
	}
	dv.ShowErrorMessage = true
	dv.SetError(excelize.DataValidationErrorStyleStop, "填写有误", "请选择下拉选项中的预设值")
	dv.ShowInputMessage = true
	if allowBlank {
		dv.SetInput("可选项", "请从下拉列表中选择")
	} else {
		dv.SetInput("必填项", "请从下拉列表中选择，不能为空")
	}
	return file.AddDataValidation(sheetName, dv)
}

func addTemplateDateValidation(file *excelize.File, sheetName, col string, startRow, endRow int, allowBlank bool) error {
	dv := excelize.NewDataValidation(allowBlank)
	dv.Sqref = fmt.Sprintf("%s%d:%s%d", col, startRow, col, endRow)
	if err := dv.SetRange("DATE(1900,1,1)", "DATE(2999,12,31)", excelize.DataValidationTypeDate, excelize.DataValidationOperatorBetween); err != nil {
		return err
	}
	dv.SetError(excelize.DataValidationErrorStyleStop, "日期格式不正确", "请输入有效日期，例如 2021-01-21")
	if allowBlank {
		dv.SetInput("日期输入", "请输入可识别的日期格式，例如 2021-01-21")
	} else {
		dv.SetInput("必填日期", "请输入可识别的日期格式，例如 2021-01-21，且不能为空")
	}
	return file.AddDataValidation(sheetName, dv)
}

func sanitizeTemplateFileName(name string) string {
	replacer := strings.NewReplacer("/", "-", "\\", "-", "?", "", "*", "", ":", "-", "\"", "", "<", "", ">", "", "|", "")
	return replacer.Replace(strings.TrimSpace(name))
}
