package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/repository"
)

const classCommentWordExportMaxRows = 500

type rehabRecordWordExportView struct {
	StudentName          string
	Gender               string
	BirthDate            string
	ClassName            string
	TeacherName          string
	TrainingDate         string
	TrainingTarget       string
	TrainingItems        []model.RehabRecordTrainingItem
	Performance          string
	Suggestion           string
	ParentFeedback       string
	ParentSignatureText  string
	ParentSignatureImage string
	FeedbackDate         string
}

func (svc *Service) ExportClassCommentWord(userID int64, baseURL string, dto model.ClassCommentPagedQueryDTO) (string, string, []byte, error) {
	instID, err := svc.rollCallInstID(userID)
	if err != nil {
		return "", "", nil, err
	}

	rows, total, err := svc.repo.ListPublishedRehabRecordWordExportRows(context.Background(), instID, dto, classCommentWordExportMaxRows)
	if err != nil {
		return "", "", nil, err
	}
	if total <= 0 {
		return "", "", nil, errors.New("当前筛选条件下暂无可导出的已发布康复记录")
	}
	if total > classCommentWordExportMaxRows {
		return "", "", nil, fmt.Errorf("当前最多支持导出%d份康复记录，请缩小筛选范围后重试", classCommentWordExportMaxRows)
	}

	views := make([]rehabRecordWordExportView, 0, len(rows))
	for _, row := range rows {
		view, err := buildRehabRecordWordExportView(row, baseURL)
		if err != nil {
			return "", "", nil, err
		}
		views = append(views, view)
	}

	fileName := fmt.Sprintf("个别训练记录表-%s.doc", time.Now().Format("20060102150405"))
	content := buildClassCommentWordDocument(views)
	return fileName, "application/msword; charset=utf-8", []byte(content), nil
}

func buildRehabRecordWordExportView(row repository.RehabRecordWordExportRow, baseURL string) (rehabRecordWordExportView, error) {
	var content model.RehabRecordContent
	if text := strings.TrimSpace(row.PublishedContentJSON); text != "" {
		if err := json.Unmarshal([]byte(text), &content); err != nil {
			return rehabRecordWordExportView{}, fmt.Errorf("解析康复记录失败: %w", err)
		}
	}
	content = normalizeRehabRecordExportContent(content)

	parentSignature := strings.TrimSpace(content.ParentSignature)
	parentSignatureImage := resolveExportImageURL(baseURL, parentSignature)
	parentSignatureText := parentSignature
	if parentSignatureImage != "" {
		parentSignatureText = ""
	}

	className := firstNonEmptyExportValue(
		content.ClassName,
		row.SourceName,
		row.LessonName,
	)
	className = formatExportClassName(className)
	teacherName := firstNonEmptyExportValue(content.TeacherName, row.TeacherName)
	trainingDate := firstNonEmptyExportValue(content.TrainingDate, formatExportDateOnly(row.StartTime))

	return rehabRecordWordExportView{
		StudentName:          firstNonEmptyExportValue(content.StudentName, row.StudentName),
		Gender:               firstNonEmptyExportValue(content.Gender, formatStudentGender(row.Sex)),
		BirthDate:            firstNonEmptyExportValue(content.BirthDate, row.BirthDate),
		ClassName:            className,
		TeacherName:          teacherName,
		TrainingDate:         trainingDate,
		TrainingTarget:       strings.TrimSpace(content.TrainingTarget),
		TrainingItems:        content.TrainingItems,
		Performance:          strings.TrimSpace(content.Performance),
		Suggestion:           strings.TrimSpace(content.Suggestion),
		ParentFeedback:       strings.TrimSpace(content.ParentFeedback),
		ParentSignatureText:  parentSignatureText,
		ParentSignatureImage: parentSignatureImage,
		FeedbackDate:         strings.TrimSpace(content.FeedbackDate),
	}, nil
}

func normalizeRehabRecordExportContent(content model.RehabRecordContent) model.RehabRecordContent {
	content.StudentName = strings.TrimSpace(content.StudentName)
	content.Gender = strings.TrimSpace(content.Gender)
	content.BirthDate = strings.TrimSpace(content.BirthDate)
	content.ClassName = strings.TrimSpace(content.ClassName)
	content.TeacherName = strings.TrimSpace(content.TeacherName)
	content.TrainingDate = strings.TrimSpace(content.TrainingDate)
	content.TrainingTarget = strings.TrimSpace(content.TrainingTarget)
	content.Performance = strings.TrimSpace(content.Performance)
	content.Suggestion = strings.TrimSpace(content.Suggestion)
	content.ParentFeedback = strings.TrimSpace(content.ParentFeedback)
	content.ParentSignature = strings.TrimSpace(content.ParentSignature)
	content.FeedbackDate = strings.TrimSpace(content.FeedbackDate)

	items := make([]model.RehabRecordTrainingItem, 0, len(content.TrainingItems))
	for _, item := range content.TrainingItems {
		title := strings.TrimSpace(item.Title)
		body := strings.TrimSpace(item.Content)
		if title == "" && body == "" {
			continue
		}
		items = append(items, model.RehabRecordTrainingItem{
			Title:   title,
			Content: body,
		})
	}
	content.TrainingItems = items
	return content
}

func buildClassCommentWordDocument(items []rehabRecordWordExportView) string {
	var builder strings.Builder
	builder.WriteString(`<!DOCTYPE html><html xmlns:o="urn:schemas-microsoft-com:office:office" xmlns:w="urn:schemas-microsoft-com:office:word"><head><meta charset="utf-8"><meta name="ProgId" content="Word.Document"><meta name="Generator" content="Microsoft Word 15"><!--[if gte mso 9]><xml><w:WordDocument><w:View>Print</w:View><w:Zoom>100</w:Zoom><w:DoNotOptimizeForBrowser/></w:WordDocument></xml><![endif]--><style>@page Section1{size:595.3pt 841.9pt;mso-page-orientation:portrait;margin:72pt 90pt 72pt 90pt;} div.Section1{page:Section1;} body{margin:0;color:#000;background:#fff;font-family:"SimSun","Songti SC",serif;font-size:10.5pt;} .page{width:415.3pt;margin:0 auto;page-break-after:always;} .page:last-child{page-break-after:auto;} .title{margin:0 0 10pt;text-align:center;font-size:22pt;font-weight:700;letter-spacing:1pt;line-height:1.2;} .sheet{width:100%;border-collapse:collapse;table-layout:fixed;mso-table-layout-alt:fixed;} .sheet td{border:1pt solid #000;padding:6pt 5pt;vertical-align:middle;line-height:1.6;word-break:break-word;font-size:10.5pt;} .label{font-weight:400;text-align:center;} .label-multi{font-weight:400;text-align:center;line-height:1.35;} .center{text-align:center;} .left{text-align:left;} .top{vertical-align:top;} .nowrap{white-space:nowrap !important;word-break:keep-all !important;} .date-center{text-align:center;mso-paragraph-align:center;} .training-content{padding:0 !important;} .training-item-label{font-weight:400;text-align:center;} .target-block{min-height:72pt;} .summary-block{min-height:84pt;} .advice-block{min-height:96pt;} .feedback-block{min-height:60pt;} .class-name-cell{padding-left:4pt !important;padding-right:2pt !important;word-break:keep-all;line-height:1.45;} .signature-row{height:20pt;mso-height-source:userset;mso-height-rule:exactly;} .signature-block{height:20pt;padding:0 5pt 1pt !important;line-height:1 !important;vertical-align:bottom !important;} .signature-layout{position:relative;width:100%;height:16pt;font-size:10.5pt;line-height:16pt;} .signature-left-block{display:block;padding-right:120pt;white-space:nowrap;line-height:16pt;} .signature-right-block{position:absolute;right:0;bottom:0;text-align:right;white-space:nowrap;word-break:keep-all;line-height:16pt;} .signature-content{display:inline-block;line-height:16pt;white-space:nowrap;vertical-align:bottom;} .signature-prefix{display:inline-block;vertical-align:bottom;} .signature-image{display:inline-block;height:16pt;width:auto;max-width:100pt;vertical-align:bottom;} .signature-placeholder{display:inline-block;min-width:100pt;height:10pt;vertical-align:bottom;} .date-inline{display:inline-block;white-space:nowrap;word-break:keep-all;line-height:16pt;vertical-align:bottom;} .date-inline span{display:inline-block;min-width:18pt;text-align:center;} .training-list{margin:0;padding:0;list-style:none;} .training-list li{margin:0;padding:5pt 6pt;border-top:1pt solid #000;line-height:1.6;} .training-list li:first-child{border-top:none;} .training-row{display:flex;} .training-name{width:78pt;flex:0 0 78pt;border-right:1pt solid #000;display:flex;align-items:center;justify-content:center;padding:5pt 4pt;box-sizing:border-box;} .training-text{flex:1;padding:5pt 6pt;box-sizing:border-box;} .training-head{font-weight:400;text-align:center;} .date-label{white-space:nowrap !important;word-break:keep-all !important;font-size:10.5pt;}</style></head><body>`)
	for _, item := range items {
		builder.WriteString(`<div class="Section1 page">`)
		builder.WriteString(`<h1 class="title">个别训练记录表</h1>`)
		builder.WriteString(`<table class="sheet"><colgroup><col width="42" style="mso-width-source:userset;mso-width-alt:840;width:42pt"><col width="56" style="mso-width-source:userset;mso-width-alt:1120;width:56pt"><col width="32" style="mso-width-source:userset;mso-width-alt:640;width:32pt"><col width="26" style="mso-width-source:userset;mso-width-alt:520;width:26pt"><col width="70" style="mso-width-source:userset;mso-width-alt:1400;width:70pt"><col width="87" style="mso-width-source:userset;mso-width-alt:1746;width:87.3pt"><col width="24" style="mso-width-source:userset;mso-width-alt:480;width:24pt"><col width="74" style="mso-width-source:userset;mso-width-alt:1480;width:74pt"></colgroup>`)
		builder.WriteString(`<tr>`)
		builder.WriteString(`<td class="label">姓名</td><td class="center" width="56" style="mso-width-source:userset;mso-width-alt:1120;width:56pt;">` + renderWordText(item.StudentName) + `</td>`)
		builder.WriteString(`<td class="label" width="32" style="mso-width-source:userset;mso-width-alt:640;width:32pt;">性别</td><td class="center" width="26" style="mso-width-source:userset;mso-width-alt:520;width:26pt;">` + renderWordText(item.Gender) + `</td>`)
		builder.WriteString(`<td class="label date-label date-center" align="center" width="70" style="mso-width-source:userset;mso-width-alt:1400;width:70pt;white-space:nowrap;">出生年月</td><td class="date-center nowrap" align="center" width="87" style="mso-width-source:userset;mso-width-alt:1746;width:87.3pt;white-space:nowrap;">` + renderWordText(item.BirthDate) + `</td>`)
		builder.WriteString(`<td class="label-multi" width="24" style="mso-width-source:userset;mso-width-alt:480;width:24pt;">班<br>别</td><td class="left class-name-cell" width="74" style="mso-width-source:userset;mso-width-alt:1480;width:74pt;">` + renderWordText(item.ClassName) + `</td>`)
		builder.WriteString(`</tr>`)
		builder.WriteString(`<tr>`)
		builder.WriteString(`<td class="label-multi">任教<br>老师</td><td colspan="3" class="left">` + renderWordText(item.TeacherName) + `</td>`)
		builder.WriteString(`<td class="label date-label date-center" align="center">训练日期</td><td colspan="3" class="date-center nowrap" align="center">` + renderWordText(item.TrainingDate) + `</td>`)
		builder.WriteString(`</tr>`)
		builder.WriteString(`<tr>`)
		builder.WriteString(`<td class="label-multi">训练<br>目标</td><td colspan="7" class="left top target-block">` + renderWordText(item.TrainingTarget) + `</td>`)
		builder.WriteString(`</tr>`)
		builder.WriteString(`<tr><td class="label-multi">训练<br>项目</td><td colspan="7" class="training-head">训练内容</td></tr>`)
		builder.WriteString(buildTrainingItemRows(item.TrainingItems))
		builder.WriteString(`<tr><td class="label-multi">学生<br>综合<br>表现</td><td colspan="7" class="left top summary-block">` + renderWordText(item.Performance) + `</td></tr>`)
		builder.WriteString(`<tr style="height:96pt;mso-height-rule:at-least;"><td class="label-multi">康复<br>建议</td><td colspan="7" class="left top advice-block" style="height:96pt;">` + renderWordText(item.Suggestion) + `</td></tr>`)
		builder.WriteString(`<tr style="height:60pt;mso-height-rule:at-least;"><td class="label-multi" rowspan="2">家长<br>意见<br>反馈</td><td colspan="7" class="left top feedback-block" style="height:60pt;">` + renderWordText(item.ParentFeedback) + `</td></tr>`)
		builder.WriteString(`<tr class="signature-row" style="height:20pt;mso-height-source:userset;mso-height-rule:exactly;"><td colspan="7" class="left signature-block" valign="bottom" style="height:20pt;vertical-align:bottom;padding:0 5pt 1pt 5pt;">` + renderSignatureDateRow(item) + `</td></tr>`)
		builder.WriteString(`</table></div>`)
	}
	builder.WriteString(`</body></html>`)
	return builder.String()
}

func buildTrainingItemRows(items []model.RehabRecordTrainingItem) string {
	if len(items) == 0 {
		items = []model.RehabRecordTrainingItem{{}}
	}

	var builder strings.Builder
	for _, item := range items {
		builder.WriteString(`<tr><td class="training-item-label">` + renderWordText(item.Title) + `</td><td colspan="7" class="left top">` + renderWordText(item.Content) + `</td></tr>`)
	}
	return builder.String()
}

func renderSignatureHTML(item rehabRecordWordExportView) string {
	if item.ParentSignatureImage != "" {
		return `<img class="signature-image" src="` + html.EscapeString(item.ParentSignatureImage) + `" alt="家长签名">`
	}
	text := strings.TrimSpace(item.ParentSignatureText)
	if text == "" {
		return `<span class="signature-placeholder"></span>`
	}
	return renderWordText(text)
}

func renderWordText(value string) string {
	text := strings.TrimSpace(value)
	if text == "" {
		return "&nbsp;"
	}
	escaped := html.EscapeString(text)
	escaped = strings.ReplaceAll(escaped, "\r\n", "\n")
	escaped = strings.ReplaceAll(escaped, "\r", "\n")
	return strings.ReplaceAll(escaped, "\n", "<br>")
}

func resolveExportImageURL(baseURL, raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	lower := strings.ToLower(value)
	switch {
	case strings.HasPrefix(lower, "data:image/"):
		return value
	case strings.HasPrefix(lower, "http://"), strings.HasPrefix(lower, "https://"):
		return value
	case strings.HasPrefix(value, "//"):
		if strings.HasPrefix(strings.ToLower(baseURL), "https://") {
			return "https:" + value
		}
		return "http:" + value
	case strings.HasPrefix(value, "/"):
		if strings.TrimSpace(baseURL) == "" {
			return value
		}
		return strings.TrimRight(baseURL, "/") + value
	case looksLikeImagePath(value):
		if strings.TrimSpace(baseURL) == "" {
			return value
		}
		return strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(value, "/")
	default:
		return ""
	}
}

func looksLikeImagePath(value string) bool {
	lower := strings.ToLower(strings.TrimSpace(value))
	for _, suffix := range []string{".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp", ".svg"} {
		if strings.Contains(lower, suffix+"?") || strings.HasSuffix(lower, suffix) {
			return true
		}
	}
	return false
}

func renderFeedbackDateHTML(value string) string {
	text := strings.TrimSpace(value)
	if text == "" {
		return `年&nbsp;&nbsp;&nbsp;&nbsp;月&nbsp;&nbsp;&nbsp;&nbsp;日`
	}
	return renderWordText(formatDisplayDateForDocument(text))
}

func renderSignatureDateRow(item rehabRecordWordExportView) string {
	var builder strings.Builder
	builder.WriteString(`<div class="signature-layout">`)
	builder.WriteString(`<span class="signature-left-block"><span class="signature-content"><span class="signature-prefix">家长签名：</span>` + renderSignatureHTML(item) + `</span></span>`)
	builder.WriteString(`<span class="signature-right-block"><span class="date-inline">` + renderFeedbackDateHTML(item.FeedbackDate) + `</span></span>`)
	builder.WriteString(`</div>`)
	return builder.String()
}

func formatStudentGender(sex *int) string {
	if sex == nil {
		return ""
	}
	switch *sex {
	case 1:
		return "男"
	case 0:
		return "女"
	default:
		return ""
	}
}

func formatExportDateOnly(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
	} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.Format("2006-01-02")
		}
	}
	return value
}

func formatDisplayDateForDocument(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
		"2006/01/02",
		"2006.01.02",
	} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.Format("2006年01月02日")
		}
	}
	return value
}

func formatExportClassName(value string) string {
	text := strings.TrimSpace(value)
	if text == "" {
		return ""
	}
	for _, separator := range []string{"-", " - ", "－", "—", "——"} {
		if index := strings.Index(text, separator); index >= 0 {
			prefix := strings.TrimRight(text[:index+len(separator)], " ")
			suffix := strings.TrimLeft(text[index+len(separator):], " ")
			if suffix == "" {
				return text
			}
			return prefix + "\n" + suffix
		}
	}
	return text
}

func firstNonEmptyExportValue(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
