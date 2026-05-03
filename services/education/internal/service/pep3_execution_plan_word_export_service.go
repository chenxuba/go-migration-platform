package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) ExportPEP3ExecutionPlanWord(userID int64, req model.PEP3ExecutionPlanWordExportRequest) (string, string, []byte, error) {
	if _, err := svc.rollCallInstID(userID); err != nil {
		return "", "", nil, err
	}
	planType := strings.ToLower(strings.TrimSpace(req.PlanType))
	var (
		title       string
		studentName string
		data        []byte
		err         error
	)
	switch planType {
	case "monthly":
		if req.MonthlyPlan == nil {
			return "", "", nil, errors.New("暂无可导出的月度计划")
		}
		title = firstNonEmptyExportValue(strings.TrimSpace(req.MonthlyPlan.Title), "康复教学月计划表")
		studentName = firstNonEmptyExportValue(strings.TrimSpace(req.MonthlyPlan.Student.Name), "学员")
		data, err = buildPEP3MonthlyPlanWordDocx(*req.MonthlyPlan)
	case "weekly":
		if req.WeeklyPlan == nil {
			return "", "", nil, errors.New("暂无可导出的周计划")
		}
		title = firstNonEmptyExportValue(strings.TrimSpace(req.WeeklyPlan.Title), "康复教学周计划日记录卡")
		studentName = firstNonEmptyExportValue(strings.TrimSpace(req.WeeklyPlan.Student.Name), "学员")
		data, err = buildPEP3WeeklyPlanWordDocx(*req.WeeklyPlan)
	default:
		return "", "", nil, errors.New("invalid execution plan type")
	}
	if err != nil {
		return "", "", nil, err
	}
	fileName := fmt.Sprintf("%s-%s-%s.docx", sanitizeExportFileName(studentName), sanitizeExportFileName(title), time.Now().Format("20060102150405"))
	return fileName, iepPlanWordContentType, data, nil
}

func buildPEP3MonthlyPlanWordDocx(plan model.PEP3MonthlyPlanAIResult) ([]byte, error) {
	if len(plan.Rows) == 0 {
		return nil, errors.New("暂无可导出的月度计划")
	}
	entries := map[string][]byte{
		"[Content_Types].xml":          []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/></Types>`),
		"_rels/.rels":                  []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>`),
		"word/_rels/document.xml.rels": []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"></Relationships>`),
		"word/document.xml":            []byte(buildPEP3MonthlyPlanDocumentXML(plan)),
	}
	return writeDocxZipEntries(entries)
}

func buildPEP3WeeklyPlanWordDocx(plan model.PEP3WeeklyPlanAIResult) ([]byte, error) {
	if len(plan.Rows) == 0 {
		return nil, errors.New("暂无可导出的周计划")
	}
	entries := map[string][]byte{
		"[Content_Types].xml":          []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/></Types>`),
		"_rels/.rels":                  []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>`),
		"word/_rels/document.xml.rels": []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"></Relationships>`),
		"word/document.xml":            []byte(buildPEP3WeeklyPlanDocumentXML(plan)),
	}
	return writeDocxZipEntries(entries)
}

func buildPEP3MonthlyPlanDocumentXML(plan model.PEP3MonthlyPlanAIResult) string {
	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	builder.WriteString(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><w:body>`)
	builder.WriteString(buildIEPTitleParagraph(firstNonEmptyExportValue(plan.Title, "康复教学月计划表")))
	builder.WriteString(buildPEP3MonthlyPlanTable(plan))
	builder.WriteString(`<w:sectPr><w:pgSz w:w="11906" w:h="16838"/><w:pgMar w:top="900" w:right="720" w:bottom="900" w:left="720" w:header="851" w:footer="992" w:gutter="0"/></w:sectPr>`)
	builder.WriteString(`</w:body></w:document>`)
	return builder.String()
}

func buildPEP3WeeklyPlanDocumentXML(plan model.PEP3WeeklyPlanAIResult) string {
	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	builder.WriteString(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><w:body>`)
	builder.WriteString(buildIEPTitleParagraph(firstNonEmptyExportValue(plan.Title, "康复教学周计划日记录卡")))
	builder.WriteString(buildPEP3WeeklyPlanTable(plan))
	builder.WriteString(`<w:sectPr><w:pgSz w:w="11906" w:h="16838"/><w:pgMar w:top="900" w:right="720" w:bottom="900" w:left="720" w:header="851" w:footer="992" w:gutter="0"/></w:sectPr>`)
	builder.WriteString(`</w:body></w:document>`)
	return builder.String()
}

func buildPEP3MonthlyPlanTable(plan model.PEP3MonthlyPlanAIResult) string {
	widths := []int{806, 907, 907, 958, 957, 655, 655, 655, 655, 826, 1049, 1050}
	var builder strings.Builder
	builder.WriteString(buildIEPTableStart(widths))
	builder.WriteString(buildIEPTableRowStart(560))
	builder.WriteString(buildIEPCell([]string{"姓名"}, widths[0], iepPlanWordCellOptions{Align: "center", VAlign: "center", Bold: true}))
	builder.WriteString(buildIEPCell([]string{plan.Student.Name}, sumInts(widths[1], widths[2]), iepPlanWordCellOptions{GridSpan: 2, Align: "center", VAlign: "center"}))
	builder.WriteString(buildIEPCell([]string{"性别"}, widths[3], iepPlanWordCellOptions{Align: "center", VAlign: "center", Bold: true}))
	builder.WriteString(buildIEPCell([]string{plan.Student.Gender}, widths[4], iepPlanWordCellOptions{Align: "center", VAlign: "center"}))
	builder.WriteString(buildIEPCell([]string{"出生年月"}, sumInts(widths[5], widths[6]), iepPlanWordCellOptions{GridSpan: 2, Align: "center", VAlign: "center", Bold: true}))
	builder.WriteString(buildIEPCell([]string{plan.Student.BirthDate}, sumInts(widths[7], widths[8], widths[9], widths[10], widths[11]), iepPlanWordCellOptions{GridSpan: 5, Align: "center", VAlign: "center"}))
	builder.WriteString(`</w:tr>`)
	builder.WriteString(buildIEPTableRowStart(560))
	builder.WriteString(buildIEPCell([]string{"制定日期"}, widths[0], iepPlanWordCellOptions{Align: "center", VAlign: "center", Bold: true}))
	builder.WriteString(buildIEPCell([]string{plan.Meta.PlanDate}, sumInts(widths[1], widths[2]), iepPlanWordCellOptions{GridSpan: 2, Align: "center", VAlign: "center"}))
	builder.WriteString(buildIEPCell([]string{"计划参与者"}, sumInts(widths[3], widths[4]), iepPlanWordCellOptions{GridSpan: 2, Align: "center", VAlign: "center", Bold: true}))
	builder.WriteString(buildIEPCell([]string{plan.Meta.Participant}, sumInts(widths[5], widths[6], widths[7], widths[8], widths[9], widths[10], widths[11]), iepPlanWordCellOptions{GridSpan: 7, Align: "center", VAlign: "center"}))
	builder.WriteString(`</w:tr>`)
	builder.WriteString(buildIEPTableRowStart(560))
	builder.WriteString(buildIEPCell([]string{"实施者"}, widths[0], iepPlanWordCellOptions{Align: "center", VAlign: "center", Bold: true}))
	builder.WriteString(buildIEPCell([]string{plan.Meta.Implementer}, sumInts(widths[1], widths[2]), iepPlanWordCellOptions{GridSpan: 2, Align: "center", VAlign: "center"}))
	builder.WriteString(buildIEPCell([]string{"实施起止日期"}, sumInts(widths[3], widths[4]), iepPlanWordCellOptions{GridSpan: 2, Align: "center", VAlign: "center", Bold: true}))
	builder.WriteString(buildIEPCell([]string{plan.Meta.StartDate + " 至 " + plan.Meta.EndDate}, sumInts(widths[5], widths[6], widths[7], widths[8], widths[9], widths[10], widths[11]), iepPlanWordCellOptions{GridSpan: 7, Align: "center", VAlign: "center", NoWrap: true}))
	builder.WriteString(`</w:tr>`)
	builder.WriteString(buildIEPTableRowStart(620))
	builder.WriteString(buildIEPCell([]string{"康复领域"}, widths[0], iepPlanWordCellOptions{Align: "center", VAlign: "center", Bold: true, CompactParagraph: true}))
	builder.WriteString(buildIEPCell([]string{"长期目标"}, sumInts(widths[1], widths[2]), iepPlanWordCellOptions{GridSpan: 2, Align: "center", VAlign: "center", Bold: true, CompactParagraph: true}))
	builder.WriteString(buildIEPCell([]string{"短期目标"}, sumInts(widths[3], widths[4]), iepPlanWordCellOptions{GridSpan: 2, Align: "center", VAlign: "center", Bold: true, CompactParagraph: true}))
	builder.WriteString(buildIEPCell([]string{"训练内容"}, sumInts(widths[5], widths[6], widths[7], widths[8]), iepPlanWordCellOptions{GridSpan: 4, Align: "center", VAlign: "center", Bold: true, CompactParagraph: true}))
	builder.WriteString(buildIEPCell([]string{"课程", "形式"}, widths[9], iepPlanWordCellOptions{Align: "center", VAlign: "center", Bold: true, CompactParagraph: true}))
	builder.WriteString(buildIEPCell([]string{"起止日期"}, sumInts(widths[10], widths[11]), iepPlanWordCellOptions{GridSpan: 2, Align: "center", VAlign: "center", Bold: true, CompactParagraph: true}))
	builder.WriteString(`</w:tr>`)
	for _, row := range plan.Rows {
		items := monthlyTrainingItemsForWord(row, plan.Meta.StartDate, plan.Meta.EndDate)
		for index, item := range items {
			builder.WriteString(buildIEPTableRowStart(620))
			if index == 0 {
				builder.WriteString(buildIEPCell([]string{row.Domain}, widths[0], iepPlanWordCellOptions{VMerge: "restart", Align: "center", VAlign: "center", Bold: true}))
				builder.WriteString(buildIEPCell(splitWordLines(row.LongGoal), sumInts(widths[1], widths[2]), iepPlanWordCellOptions{GridSpan: 2, VMerge: "restart", VAlign: "center", IndentLeft: 120}))
				builder.WriteString(buildIEPCell(splitWordLines(row.ShortGoal), sumInts(widths[3], widths[4]), iepPlanWordCellOptions{GridSpan: 2, VMerge: "restart", VAlign: "center", IndentLeft: 120}))
			} else {
				builder.WriteString(buildIEPCell(nil, widths[0], iepPlanWordCellOptions{VMerge: "continue"}))
				builder.WriteString(buildIEPCell(nil, sumInts(widths[1], widths[2]), iepPlanWordCellOptions{GridSpan: 2, VMerge: "continue"}))
				builder.WriteString(buildIEPCell(nil, sumInts(widths[3], widths[4]), iepPlanWordCellOptions{GridSpan: 2, VMerge: "continue"}))
			}
			builder.WriteString(buildIEPCell(splitWordLines(fmt.Sprintf("%d. %s", index+1, item.Content)), sumInts(widths[5], widths[6], widths[7], widths[8]), iepPlanWordCellOptions{GridSpan: 4, VAlign: "center", IndentLeft: 120}))
			if index == 0 {
				builder.WriteString(buildIEPCell([]string{row.CourseForm}, widths[9], iepPlanWordCellOptions{VMerge: "restart", Align: "center", VAlign: "center", NoWrap: true, CompactParagraph: true}))
			} else {
				builder.WriteString(buildIEPCell(nil, widths[9], iepPlanWordCellOptions{VMerge: "continue"}))
			}
			builder.WriteString(buildIEPCell([]string{item.StartEndDate}, sumInts(widths[10], widths[11]), iepPlanWordCellOptions{GridSpan: 2, Align: "center", VAlign: "center", NoWrap: true, CompactParagraph: true}))
			builder.WriteString(`</w:tr>`)
		}
	}
	builder.WriteString(`</w:tbl>`)
	return builder.String()
}

func monthlyTrainingItemsForWord(row model.PEP3MonthlyPlanRow, startDate, endDate string) []model.PEP3MonthlyTrainingItem {
	items := make([]model.PEP3MonthlyTrainingItem, 0, len(row.TrainingItems))
	for _, item := range row.TrainingItems {
		content := strings.TrimSpace(item.Content)
		if content == "" {
			continue
		}
		items = append(items, model.PEP3MonthlyTrainingItem{
			Content:      content,
			StartEndDate: strings.TrimSpace(item.StartEndDate),
		})
	}
	if len(items) == 0 {
		items = append(items, model.PEP3MonthlyTrainingItem{Content: ""})
	}
	for index := range items {
		if strings.TrimSpace(items[index].StartEndDate) == "" {
			items[index].StartEndDate = firstNonEmptyExportValue(monthlyItemDateRange(startDate, endDate, index, len(items)), startDate+" - "+endDate)
		}
	}
	return items
}

func buildPEP3WeeklyPlanTable(plan model.PEP3WeeklyPlanAIResult) string {
	widths := []int{1310, 3326, 907, 907, 907, 907, 907, 909}
	weekDates := plan.WeekDates
	if len(weekDates) > 6 {
		weekDates = weekDates[:6]
	}
	for len(weekDates) < 6 {
		weekDates = append(weekDates, "")
	}
	var builder strings.Builder
	builder.WriteString(buildIEPTableStart(widths))
	builder.WriteString(buildIEPTableRowStart(560))
	builder.WriteString(buildIEPCell([]string{"姓名"}, widths[0], iepPlanWordCellOptions{Align: "center", VAlign: "center", Bold: true}))
	builder.WriteString(buildIEPCell([]string{plan.Student.Name}, widths[1], iepPlanWordCellOptions{Align: "center", VAlign: "center"}))
	builder.WriteString(buildIEPCell([]string{"性别"}, widths[2], iepPlanWordCellOptions{Align: "center", VAlign: "center", Bold: true}))
	builder.WriteString(buildIEPCell([]string{plan.Student.Gender}, widths[3], iepPlanWordCellOptions{Align: "center", VAlign: "center"}))
	builder.WriteString(buildIEPCell([]string{"出生年月"}, widths[4], iepPlanWordCellOptions{Align: "center", VAlign: "center", Bold: true}))
	builder.WriteString(buildIEPCell([]string{plan.Student.BirthDate}, widths[5]+widths[6]+widths[7], iepPlanWordCellOptions{GridSpan: 3, Align: "center", VAlign: "center"}))
	builder.WriteString(`</w:tr>`)
	builder.WriteString(buildIEPTableRowStart(620))
	builder.WriteString(buildIEPCell([]string{"任教", "老师"}, widths[0], iepPlanWordCellOptions{Align: "center", VAlign: "center", Bold: true, CompactParagraph: true}))
	builder.WriteString(buildIEPCell([]string{plan.TeacherName}, widths[1], iepPlanWordCellOptions{Align: "center", VAlign: "center"}))
	builder.WriteString(buildIEPCell([]string{"课程", "名称"}, widths[2], iepPlanWordCellOptions{Align: "center", VAlign: "center", Bold: true, CompactParagraph: true}))
	builder.WriteString(buildIEPCell(splitWordLines(plan.CourseName), widths[3], iepPlanWordCellOptions{Align: "center", VAlign: "center", CompactParagraph: true}))
	builder.WriteString(buildIEPCell([]string{"训练日期"}, widths[4], iepPlanWordCellOptions{Align: "center", VAlign: "center", Bold: true}))
	builder.WriteString(buildIEPCell([]string{plan.TrainingDate}, widths[5]+widths[6]+widths[7], iepPlanWordCellOptions{GridSpan: 3, Align: "center", VAlign: "center", NoWrap: true}))
	builder.WriteString(`</w:tr>`)
	builder.WriteString(buildIEPTableRowStart(760))
	builder.WriteString(buildIEPCell([]string{"训练前", "准备"}, widths[0], iepPlanWordCellOptions{Align: "center", VAlign: "center", Bold: true, CompactParagraph: true}))
	builder.WriteString(buildIEPCell(splitWordLines(plan.Preparation), widths[1]+widths[2]+widths[3]+widths[4]+widths[5]+widths[6]+widths[7], iepPlanWordCellOptions{GridSpan: 7, VAlign: "center", IndentLeft: 120}))
	builder.WriteString(`</w:tr>`)
	builder.WriteString(buildIEPTableRowStart(620))
	builder.WriteString(buildIEPCell([]string{"训练项目"}, widths[0], iepPlanWordCellOptions{VMerge: "restart", Align: "center", VAlign: "center", Bold: true}))
	builder.WriteString(buildIEPCell([]string{"训练内容"}, widths[1], iepPlanWordCellOptions{VMerge: "restart", Align: "center", VAlign: "center", Bold: true}))
	builder.WriteString(buildIEPCell([]string{"完成情况"}, widths[2]+widths[3]+widths[4]+widths[5]+widths[6]+widths[7], iepPlanWordCellOptions{GridSpan: 6, Align: "center", VAlign: "center", Bold: true}))
	builder.WriteString(`</w:tr>`)
	builder.WriteString(buildIEPTableRowStart(480))
	builder.WriteString(buildIEPCell(nil, widths[0], iepPlanWordCellOptions{VMerge: "continue"}))
	builder.WriteString(buildIEPCell(nil, widths[1], iepPlanWordCellOptions{VMerge: "continue"}))
	for index, date := range weekDates {
		builder.WriteString(buildIEPCell([]string{date}, widths[2+index], iepPlanWordCellOptions{Align: "center", VAlign: "center", CompactParagraph: true}))
	}
	builder.WriteString(`</w:tr>`)
	for _, row := range plan.Rows {
		builder.WriteString(buildIEPTableRowStart(720))
		builder.WriteString(buildIEPCell(splitWordLines(row.Project), widths[0], iepPlanWordCellOptions{Align: "center", VAlign: "center"}))
		builder.WriteString(buildIEPCell(splitWordLines(row.Content), widths[1], iepPlanWordCellOptions{VAlign: "center", IndentLeft: 120}))
		for index := range weekDates {
			value := ""
			if index < len(row.Completion) {
				value = strings.TrimSpace(row.Completion[index])
			}
			builder.WriteString(buildIEPCell([]string{value}, widths[2+index], iepPlanWordCellOptions{Align: "center", VAlign: "center"}))
		}
		builder.WriteString(`</w:tr>`)
	}
	builder.WriteString(`</w:tbl>`)
	return builder.String()
}
