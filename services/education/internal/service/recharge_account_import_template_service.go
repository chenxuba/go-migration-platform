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

func (svc *Service) BuildRechargeAccountImportByStudentTemplate(userID int64) (string, error) {
	return svc.buildRechargeAccountImportTemplate(userID, "按关联学员", buildRechargeAccountImportByStudentColumns, buildRechargeAccountImportByStudentNotesRichText)
}

func (svc *Service) LoadRechargeAccountImportByStudentTemplate(ticket string) (string, string, []byte, bool) {
	return svc.loadRechargeAccountImportTemplate(ticket)
}

func (svc *Service) BuildRechargeAccountImportByAccountTemplate(userID int64) (string, error) {
	return svc.buildRechargeAccountImportTemplate(userID, "按储值账户", buildRechargeAccountImportByAccountColumns, buildRechargeAccountImportByAccountNotesRichText)
}

func (svc *Service) LoadRechargeAccountImportByAccountTemplate(ticket string) (string, string, []byte, bool) {
	return svc.loadRechargeAccountImportTemplate(ticket)
}

type rechargeAccountImportColumnBuilder func(defaultFields []model.StudentFieldKey, channels []model.ChannelVO, staffNames, orderTagNames []string) []model.IntentionStudentImportTemplateColumn
type rechargeAccountImportNotesBuilder func(orgName string, columns []model.IntentionStudentImportTemplateColumn) []excelize.RichTextRun

func (svc *Service) buildRechargeAccountImportTemplate(userID int64, mode string, columnBuilder rechargeAccountImportColumnBuilder, notesBuilder rechargeAccountImportNotesBuilder) (string, error) {
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
	staffNames, err := svc.repo.ListActiveStaffNames(context.Background(), instID)
	if err != nil {
		return "", err
	}
	orderTagNames, err := svc.repo.ListEnabledOrderTagNames(context.Background(), instID)
	if err != nil {
		return "", err
	}

	columns := columnBuilder(defaultFields, channels, staffNames, orderTagNames)
	filename := sanitizeTemplateFileName(fmt.Sprintf("%s导入储值账户模板-%s-%s.xlsx", orgName, mode, time.Now().Format("20060102")))
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

	if mode == "按储值账户" {
		return "/api/v1/recharge-accounts/import-template/by-account/file?ticket=" + url.QueryEscape(ticket), nil
	}
	return "/api/v1/recharge-accounts/import-template/by-student/file?ticket=" + url.QueryEscape(ticket), nil
}

func (svc *Service) loadRechargeAccountImportTemplate(ticket string) (string, string, []byte, bool) {
	file, ok := loadTemplateDownloadFile(strings.TrimSpace(ticket))
	if !ok {
		return "", "", nil, false
	}
	return file.Filename, file.ContentType, file.Data, true
}

func buildRechargeChannelOptions(defaultFields []model.StudentFieldKey, channels []model.ChannelVO) []string {
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
	if len(channelOptions) > 0 {
		return channelOptions
	}
	if field, ok := displayedDefaults["渠道"]; ok && strings.TrimSpace(field.OptionsJSON) != "" {
		return splitTemplateOptions(field.OptionsJSON)
	}
	return nil
}

func findRechargeStudentField(defaultFields []model.StudentFieldKey, title string) *model.StudentFieldKey {
	for idx := range defaultFields {
		field := defaultFields[idx]
		if strings.TrimSpace(field.FieldKey) == title && field.IsDisplay {
			return &field
		}
	}
	return nil
}

func appendRechargeStudentSystemColumn(columns *[]model.IntentionStudentImportTemplateColumn, defaultFields []model.StudentFieldKey, title string, fieldType int, options []string) {
	field := findRechargeStudentField(defaultFields, title)
	if field == nil {
		return
	}
	column := model.IntentionStudentImportTemplateColumn{
		FieldID:   field.ID,
		Title:     title,
		Required:  field.Required,
		FieldType: fieldType,
		Options:   options,
	}
	if len(column.Options) == 0 && strings.TrimSpace(field.OptionsJSON) != "" {
		column.Options = splitTemplateOptions(field.OptionsJSON)
	}
	*columns = append(*columns, column)
}

func buildRechargeAccountImportByStudentColumns(defaultFields []model.StudentFieldKey, channels []model.ChannelVO, staffNames, orderTagNames []string) []model.IntentionStudentImportTemplateColumn {
	channelOptions := buildRechargeChannelOptions(defaultFields, channels)
	columns := []model.IntentionStudentImportTemplateColumn{
		{Title: "学员姓名", Required: true, FieldType: 1},
		{Title: "手机号", Required: true, FieldType: 1},
		{Title: "手机号归属人", Required: true, FieldType: 4, Options: []string{"爸爸", "妈妈", "爷爷", "奶奶", "外公", "外婆", "其他"}},
		{
			FieldID: func() int64 {
				if field := findRechargeStudentField(defaultFields, "性别"); field != nil {
					return field.ID
				}
				return 0
			}(),
			Title: "性别", Required: func() bool {
				if field := findRechargeStudentField(defaultFields, "性别"); field != nil {
					return field.Required
				}
				return false
			}(),
			FieldType: 4,
			Options:   []string{"男", "女", "未知"},
		},
	}

	appendRechargeStudentSystemColumn(&columns, defaultFields, "渠道", 4, channelOptions)
	appendRechargeStudentSystemColumn(&columns, defaultFields, "生日", 3, nil)
	columns = append(columns, model.IntentionStudentImportTemplateColumn{Title: "销售员", FieldType: 4, Options: staffNames})
	columns = append(columns, model.IntentionStudentImportTemplateColumn{Title: "推荐人", FieldType: 1})
	appendRechargeStudentSystemColumn(&columns, defaultFields, "微信号", 1, nil)
	appendRechargeStudentSystemColumn(&columns, defaultFields, "年级", 4, nil)
	appendRechargeStudentSystemColumn(&columns, defaultFields, "就读学校", 1, nil)
	appendRechargeStudentSystemColumn(&columns, defaultFields, "兴趣爱好", 1, nil)
	appendRechargeStudentSystemColumn(&columns, defaultFields, "家庭住址", 1, nil)
	columns = append(columns, model.IntentionStudentImportTemplateColumn{Title: "学员备注", FieldType: 1})
	columns = append(columns,
		model.IntentionStudentImportTemplateColumn{Title: "充值金额", FieldType: 2},
		model.IntentionStudentImportTemplateColumn{Title: "赠送金额", FieldType: 2},
		model.IntentionStudentImportTemplateColumn{Title: "经办日期", FieldType: 3},
		model.IntentionStudentImportTemplateColumn{Title: "订单销售员", FieldType: 4, Options: staffNames},
		model.IntentionStudentImportTemplateColumn{Title: "对内备注", FieldType: 1},
		model.IntentionStudentImportTemplateColumn{Title: "对外备注", FieldType: 1},
		model.IntentionStudentImportTemplateColumn{Title: "订单标签", FieldType: 4, Options: orderTagNames},
		model.IntentionStudentImportTemplateColumn{Title: "收款方式", FieldType: 4, Options: []string{"微信", "支付宝", "银行转账", "POS机", "现金", "其他方式"}},
		model.IntentionStudentImportTemplateColumn{Title: "收款账户", FieldType: 4, Options: []string{"默认账户"}},
		model.IntentionStudentImportTemplateColumn{Title: "支付单号", FieldType: 1},
		model.IntentionStudentImportTemplateColumn{Title: "对方账户", FieldType: 1},
		model.IntentionStudentImportTemplateColumn{Title: "支付日期", FieldType: 3},
		model.IntentionStudentImportTemplateColumn{Title: "账单备注", FieldType: 1},
	)
	return columns
}

func buildRechargeAccountImportByAccountColumns(_ []model.StudentFieldKey, _ []model.ChannelVO, staffNames, orderTagNames []string) []model.IntentionStudentImportTemplateColumn {
	return []model.IntentionStudentImportTemplateColumn{
		{Title: "储值账户号", Required: true, FieldType: 1},
		{Title: "充值金额", FieldType: 2},
		{Title: "赠送金额", FieldType: 2},
		{Title: "经办日期", FieldType: 3},
		{Title: "订单销售员", FieldType: 4, Options: staffNames},
		{Title: "对内备注", FieldType: 1},
		{Title: "对外备注", FieldType: 1},
		{Title: "订单标签", FieldType: 4, Options: orderTagNames},
		{Title: "收款方式", FieldType: 4, Options: []string{"微信", "支付宝", "银行转账", "POS机", "现金", "其他方式"}},
		{Title: "收款账户", FieldType: 4, Options: []string{"默认账户"}},
		{Title: "支付单号", FieldType: 1},
		{Title: "对方账户", FieldType: 1},
		{Title: "支付日期", FieldType: 3},
		{Title: "账单备注", FieldType: 1},
	}
}

func buildRechargeAccountImportByStudentNotesRichText(orgName string, columns []model.IntentionStudentImportTemplateColumn) []excelize.RichTextRun {
	return buildRechargeAccountImportNotesRichText(
		orgName,
		"按关联学员",
		columns,
		[]string{
			"3、「学员姓名」支持最多 20 个字。",
			"4、「手机号」必须为 1 开头的 11 位数字，不支持“-”和中间空格，例如 13311113333。",
			"5、「手机号归属人」、「渠道」、「性别」、「销售员」、「订单销售员」、「收款方式」、「收款账户」为下拉选择项。",
			"6、「充值金额」、「赠送金额」只支持输入两位小数，请勿携带单位“元”或“￥”。",
			"7、「经办日期」、「支付日期」支持财年、月份、日期等常见日期格式，例如 2021-01-21、2021/01/21、2021.01.21、20210121。",
			"8、「订单标签」如需填写，请使用系统中已存在的标签名称，多个标签用英文（,）分隔。",
			"9、如更新了自定义字段的字段名称或选项，请重新下载模板后填写。",
			"10、最多导入1000条数据，请控制导入数量。",
		},
		[]excelize.RichTextRun{
			{Text: "按关联学员导入时，请确保所「关联的学员」已存在。如学员不存在，系统将自动为您创建新学员。\n", Font: &excelize.Font{Size: 12, Family: "Microsoft YaHei", Color: "#111111"}},
			{Text: "按储值账户导入时，请确保「储值账户号」已存在。", Font: &excelize.Font{Size: 12, Family: "Microsoft YaHei", Color: "#111111"}},
		},
	)
}

func buildRechargeAccountImportByAccountNotesRichText(orgName string, columns []model.IntentionStudentImportTemplateColumn) []excelize.RichTextRun {
	return buildRechargeAccountImportNotesRichText(
		orgName,
		"按储值账户",
		columns,
		[]string{
			"3、「充值金额」、「赠送金额」只支持输入两位小数，请勿携带单位“元”或“￥”。",
			"4、「经办日期」、「支付日期」支持财年、月份、日期等常见日期格式，例如 2021-01-21、2021/01/21、2021.01.21、20210121。",
			"5、「订单销售员」、「收款方式」、「收款账户」为下拉选择项。",
			"6、「订单标签」如需填写，请使用系统中已存在的标签名称，多个标签用英文（,）分隔。",
			"7、最多导入1000条数据，请控制导入数量。",
		},
		[]excelize.RichTextRun{
			{Text: "导入前请确保对应的储值账户号准确无误。", Font: &excelize.Font{Size: 12, Family: "Microsoft YaHei", Color: "#111111"}},
		},
	)
}

func buildRechargeAccountImportNotesRichText(orgName, mode string, columns []model.IntentionStudentImportTemplateColumn, extraRules []string, tips []excelize.RichTextRun) []excelize.RichTextRun {
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
		{Text: "导入储值账户-" + mode + "模板，用于向", Font: black},
		{Text: "【" + orgName + "】", Font: red},
		{Text: "导入储值账户数据。\n\n", Font: black},
		{Text: "【导入提示】\n", Font: title},
	}
	runs = append(runs, tips...)
	runs = append(runs, excelize.RichTextRun{Text: "\n\n【填写规范】\n", Font: title})

	if len(requiredFields) > 0 {
		runs = append(runs,
			excelize.RichTextRun{Text: "1、标*字段，", Font: red},
			excelize.RichTextRun{Text: strings.Join(requiredFields, "、"), Font: red},
			excelize.RichTextRun{Text: "为必填项。\n", Font: black},
		)
	} else {
		runs = append(runs, excelize.RichTextRun{Text: "1、请按字段要求补充完整内容。\n", Font: black})
	}
	runs = append(runs, excelize.RichTextRun{Text: "2、请勿修改顶部字段标题及顺序。\n", Font: black})
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
		excelize.RichTextRun{Text: "请严格按模板填写。当前页面已完成模板下载对接，上传解析流程后续接入。", Font: black},
	)
	return runs
}
