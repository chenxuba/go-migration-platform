package service

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html"
	"net/url"
	"sort"
	"strings"
	"time"

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
	filename := sanitizeTemplateFileName(fmt.Sprintf("%s导入意向学员模板-%s.xls", orgName, time.Now().Format("20060102")))
	content := buildIntentionStudentImportTemplateContent(orgName, columns)
	ticket := saveTemplateDownloadFile(templateDownloadFile{
		Filename:    filename,
		ContentType: "application/vnd.ms-excel; charset=utf-8",
		Data:        []byte(content),
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
		{Title: "手机号", Required: true, FieldType: 1},
		{Title: "手机号归属人", Required: true, FieldType: 4, Options: []string{"爸爸", "妈妈", "爷爷", "奶奶", "外公", "外婆", "其他"}},
	}

	appendDefaultColumn := func(title string, fieldType int, fallbackOptions []string) {
		field, ok := displayedDefaults[title]
		if !ok {
			return
		}
		options := fallbackOptions
		if len(options) == 0 && strings.TrimSpace(field.OptionsJSON) != "" {
			options = splitTemplateOptions(field.OptionsJSON)
		}
		columns = append(columns, model.IntentionStudentImportTemplateColumn{
			Title:     title,
			Required:  field.Required,
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
			Title:     strings.TrimSpace(field.FieldKey),
			Required:  field.Required,
			FieldType: field.FieldType,
		}
		if field.FieldType == 4 {
			column.Options = splitTemplateOptions(field.OptionsJSON)
		}
		columns = append(columns, column)
	}

	columns = append(columns, model.IntentionStudentImportTemplateColumn{
		Title:     "销售员",
		Required:  false,
		FieldType: 4,
		Options:   staffNames,
	})

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

func buildIntentionStudentImportTemplateContent(orgName string, columns []model.IntentionStudentImportTemplateColumn) string {
	var buffer bytes.Buffer
	buffer.WriteString(`<!DOCTYPE html><html xmlns:o="urn:schemas-microsoft-com:office:office" xmlns:x="urn:schemas-microsoft-com:office:excel"><head><meta http-equiv="Content-Type" content="text/html; charset=utf-8">`)
	buffer.WriteString(`<style>
table{border-collapse:collapse;font-size:12px;color:#222;font-family:'Microsoft YaHei','PingFang SC',sans-serif;}
th,td{border:1px solid #d8dde8;height:26px;padding:0 6px;vertical-align:top;}
th{height:30px;background:#dfe8f7;font-weight:700;font-size:12px;text-align:center;white-space:nowrap;border-color:#8f99ad;}
td{background:#fff;}
.required{color:#ff4d4f;font-weight:700;}
.sheet-cell{min-width:78px;}
.note-head{position:relative;}
.note-head:before{content:'';position:absolute;left:0;top:0;width:0;height:0;border-top:10px solid #36b36b;border-right:10px solid transparent;}
.note-wrap{min-width:560px;padding:10px 12px 14px;line-height:1.8;word-break:break-word;white-space:normal;}
.note-p{margin:0 0 14px;}
.note-title{margin:18px 0 8px;font-size:13px;font-weight:700;color:#111;}
.note-item{margin:0 0 2px;}
.note-red{color:#ff3b30;}
</style></head><body><table>`)

	buffer.WriteString("<tr>")
	for _, column := range columns {
		buffer.WriteString(`<th class="sheet-cell">`)
		if column.Required {
			buffer.WriteString(`<span class="required">*</span>`)
		}
		buffer.WriteString(html.EscapeString(column.Title))
		buffer.WriteString("</th>")
	}
	buffer.WriteString(`<th class="note-head">填写说明</th></tr>`)

	const rowCount = 120
	notes := buildIntentionStudentImportNotesHTML(orgName, columns)
	for row := 0; row < rowCount; row++ {
		buffer.WriteString("<tr>")
		for range columns {
			buffer.WriteString("<td></td>")
		}
		if row == 0 {
			buffer.WriteString(`<td rowspan="120"><div class="note-wrap">`)
			buffer.WriteString(notes)
			buffer.WriteString(`</div></td>`)
		}
		buffer.WriteString("</tr>")
	}

	buffer.WriteString("</table></body></html>")
	return buffer.String()
}

func buildIntentionStudentImportNotesHTML(orgName string, columns []model.IntentionStudentImportTemplateColumn) string {
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

	var builder strings.Builder
	builder.WriteString(`<p class="note-p">你好，当前 excel 表格学员数据用于导入<span class="note-red">【`)
	builder.WriteString(html.EscapeString(orgName))
	builder.WriteString(`】</span>使用</p>`)

	builder.WriteString(`<div class="note-title">【导入提示】</div>`)
	builder.WriteString(`<p class="note-p">「来源渠道」、「销售员」取的是系统内的预设信息，如填写信息不符合预设值，则导入会失败。</p>`)

	builder.WriteString(`<div class="note-title">【填写规范】</div>`)
	builder.WriteString(`<p class="note-item">1、请勿修改顶部字段标题及顺序</p>`)
	if len(requiredFields) > 0 {
		builder.WriteString(`2、<span class="note-red">标*字段，`)
		builder.WriteString(html.EscapeString(strings.Join(requiredFields, "、")))
		builder.WriteString(`</span>为必填项，`)
	} else {
		builder.WriteString(`2、请按字段要求补充完整内容，`)
	}
	builder.WriteString(`<span class="note-red">「来源渠道」、「销售员」、下拉选项字段</span>请按预设值填写</p>`)
	builder.WriteString(`<p class="note-item">3、「手机号」必须为 1 开头的 11 位数字，不支持“-”和中间空格，<span class="note-red">支持样式 13311113333</span></p>`)
	builder.WriteString(`<p class="note-item">4、「出生日期」的日期格式支持年月日输入，支持<span class="note-red">2021-01-21、2021/01/21、2021.01.21、20210121</span>四种样式</p>`)
	builder.WriteString(`<p class="note-item">5、自定义添加的字段请按照对应的格式类型进行填写，如「数字」的格式类型，只支持输入阿拉伯数字，请勿携带单位“节”或“元”</p>`)
	builder.WriteString(`<p class="note-item">6、如更新了自定义字段的字段名称、必填规则，则需重新下载模板，填写信息后再进行导入</p>`)

	if len(optionBlocks) > 0 {
		builder.WriteString(`<div class="note-title">【选项说明】</div>`)
		for _, block := range optionBlocks {
			builder.WriteString(`<p class="note-item">`)
			builder.WriteString(html.EscapeString(block))
			builder.WriteString(`</p>`)
		}
	}

	builder.WriteString(`<div class="note-title">【其他注意】</div>`)
	builder.WriteString(`<p class="note-p">最多导入1000条数据，请控制导入数量。</p>`)

	return builder.String()
}

func sanitizeTemplateFileName(name string) string {
	replacer := strings.NewReplacer("/", "-", "\\", "-", "?", "", "*", "", ":", "-", "\"", "", "<", "", ">", "", "|", "")
	return replacer.Replace(strings.TrimSpace(name))
}
