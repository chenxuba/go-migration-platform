package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

const (
	iepPlanWordContentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	iepPlanWordTableWidth  = 10080
)

type iepPlanWordExport struct {
	Title         string
	StudentName   string
	Gender        string
	BirthDate     string
	PlanDate      string
	Participant   string
	PlannerName   string
	StartDate     string
	EndDate       string
	DurationLabel string
	Domains       []iepPlanWordDomain
	Rows          []iepPlanWordRow
	HomePlan      []string
}

type iepPlanWordDomain struct {
	Name       string
	LongGoals  []string
	StageGoals []iepPlanWordStage
}

type iepPlanWordRow struct {
	Domain       string
	LongGoal     string
	ShortGoal    string
	CourseForm   string
	StartEndDate string
}

type iepPlanWordStage struct {
	Title string
	Goals []string
}

type iepPlanWordCellOptions struct {
	GridSpan         int
	VMerge           string
	Align            string
	VAlign           string
	NoWrap           bool
	Bold             bool
	IndentLeft       int
	SpacingBefore    int
	SpacingAfter     int
	LineSpacing      int
	CompactParagraph bool
}

func (svc *Service) ExportPEP3IEPPlanWord(userID int64, recordID int64, durationMonths int) (string, string, []byte, error) {
	if durationMonths != 6 {
		durationMonths = 3
	}
	if _, err := svc.rollCallInstID(userID); err != nil {
		return "", "", nil, err
	}

	plan := buildStaticPEP3IEPPlanWordExport(durationMonths)
	data, err := buildPEP3IEPPlanWordDocx(plan)
	if err != nil {
		return "", "", nil, err
	}

	fileName := fmt.Sprintf("%s-康复个别化教育计划-%s.docx", sanitizeExportFileName(plan.StudentName), time.Now().Format("20060102150405"))
	return fileName, iepPlanWordContentType, data, nil
}

func (svc *Service) ExportPEP3IEPPlanWordFromAIResult(userID int64, recordID int64, planResult model.PEP3IEPPlanAIResult, durationMonths int) (string, string, []byte, error) {
	if durationMonths != 6 {
		durationMonths = 3
	}
	instID, err := svc.rollCallInstID(userID)
	if err != nil {
		return "", "", nil, err
	}
	if recordID > 0 {
		record, err := svc.repo.GetAssessmentRecord(context.Background(), instID, recordID)
		if err != nil {
			return "", "", nil, err
		}
		planResult = applyPEP3IEPPlanHeaderValues(planResult, pep3IEPPlanHeaderValuesForRecord(record))
	}

	plan := buildPEP3IEPPlanWordExportFromAIResult(planResult, durationMonths)
	data, err := buildPEP3IEPPlanWordDocx(plan)
	if err != nil {
		return "", "", nil, err
	}
	fileName := fmt.Sprintf("%s-康复教学计划-%s.docx", sanitizeExportFileName(plan.StudentName), time.Now().Format("20060102150405"))
	return fileName, iepPlanWordContentType, data, nil
}

func buildStaticPEP3IEPPlanWordExport(durationMonths int) iepPlanWordExport {
	startDate := "2026-05-01"
	endDate := "2026-08-01"
	stages := []iepPlanWordStage{
		{
			Title: "第1个月 建立表达基础",
			Goals: []string{
				"用单词表达需求",
				"模仿功能词：要、不要",
				"二选一情境选择表达",
			},
		},
		{
			Title: "第2个月 提升主动沟通",
			Goals: []string{
				"主动使用2词短句提出请求",
				"回应简单问句",
				"小组活动中发起沟通",
			},
		},
		{
			Title: "第3个月 泛化与稳定",
			Goals: []string{
				"表达拒绝或帮助需求",
				"家庭场景简单问答",
				"受挫时使用语言表达情绪",
			},
		},
	}
	if durationMonths == 6 {
		endDate = "2026-11-01"
		stages = []iepPlanWordStage{
			{Title: "第1个月 表达启动", Goals: []string{"用单词表达需求", "模仿功能词：要、不要", "等待情境中主动请求帮助"}},
			{Title: "第2个月 词汇积累", Goals: []string{"进行口型和发音模仿", "练习“要”“不要”等功能词跟读", "功能词在生活场景中泛化"}},
			{Title: "第3个月 短句表达", Goals: []string{"主动使用2词短句提出请求", "回应简单问句", "受挫时使用语言表达情绪"}},
			{Title: "第4个月 问答轮替", Goals: []string{"识别“要什么”“在哪里”等问句类型", "使用实物或图片辅助完成回应", "同伴轮替中发起简单沟通"}},
			{Title: "第5个月 场景泛化", Goals: []string{"表达拒绝或帮助需求", "家庭场景简单问答", "家庭外出场景主动表达"}},
			{Title: "第6个月 稳定维持", Goals: []string{"完成吃饭、玩具、外出等主题问答", "使用短句回应家长提问", "稳定使用短句完成需求表达"}},
		}
	}

	domains := []iepPlanWordDomain{
		{
			Name: "语言沟通",
			LongGoals: []string{
				"在自然情境中主动表达需求，能使用2-3词短句完成请求、拒绝和简单回应。",
				"提升语言模仿与功能性表达能力，减少仅用动作或哭闹表达需求的情况。",
				"将课堂训练中的表达方式泛化到家庭、集体活动和日常生活场景。",
			},
			StageGoals: stages,
		},
		{
			Name: "社交互动",
			LongGoals: []string{
				"提升共同注意和互动回应能力，能在课堂活动中完成简单轮替互动。",
				"在成人提示下参与同伴活动，逐步增加主动发起和等待轮到自己的行为。",
			},
			StageGoals: stages,
		},
		{
			Name: "认知理解",
			LongGoals: []string{
				"提升物品功能、类别配对和两步指令理解能力。",
				"能在课堂和家庭情境中根据口头指令完成简单任务。",
			},
			StageGoals: stages,
		},
		{
			Name: "精细动作",
			LongGoals: []string{
				"提升手眼协调、抓握控制和双手协作能力。",
				"能完成穿珠、描线、夹取等精细动作任务并迁移到生活操作。",
			},
			StageGoals: stages,
		},
		{
			Name: "感统运动",
			LongGoals: []string{
				"提升前庭、本体和动作计划能力，增强课堂坐姿稳定和活动参与。",
				"在规则明确的运动活动中完成等待、启动和停止。",
			},
			StageGoals: stages,
		},
		{
			Name: "生活自理",
			LongGoals: []string{
				"提升穿脱、进餐和如厕相关生活自理能力。",
			},
			StageGoals: stages,
		},
	}

	return iepPlanWordExport{
		Title:         iepPlanWordTitle(durationMonths),
		StudentName:   "张一鸣",
		Gender:        "男",
		BirthDate:     "2023-04-30",
		PlanDate:      time.Now().Format("2006-01-02"),
		Participant:   "1",
		PlannerName:   "陈瑞",
		StartDate:     startDate,
		EndDate:       endDate,
		DurationLabel: fmt.Sprintf("%d个月", durationMonths),
		Domains:       domains,
		HomePlan: []string{
			"家庭训练围绕日常高频场景开展，包括进餐、玩具选择、穿衣、外出准备和睡前互动。家长在孩子出现需求前先等待3-5秒，鼓励其使用单词或2词短句表达，再给予物品或活动强化。",
			"每天安排2-3次短时训练，每次10-15分钟，重点练习请求、拒绝、选择和简单问答。训练时减少直接替孩子表达，优先使用实物提示、图片提示和语言示范，并逐步降低提示强度。",
			"家庭成员保持一致回应方式，当孩子用哭闹、拉人或指物替代表达时，先引导其完成可接受表达，再满足需求；如出现明显抗拒，可降低难度并回到单词表达。",
		},
	}
}

func buildPEP3IEPPlanWordExportFromAIResult(result model.PEP3IEPPlanAIResult, durationMonths int) iepPlanWordExport {
	startDate := strings.TrimSpace(result.Meta.StartDate)
	endDate := strings.TrimSpace(result.Meta.EndDate)
	if startDate == "" || endDate == "" {
		fallback := buildStaticPEP3IEPPlanWordExport(durationMonths)
		if startDate == "" {
			startDate = fallback.StartDate
		}
		if endDate == "" {
			endDate = fallback.EndDate
		}
	}
	rows := make([]iepPlanWordRow, 0, len(result.Rows))
	for _, row := range result.Rows {
		domain := strings.TrimSpace(row.Domain)
		longGoal := strings.TrimSpace(row.LongGoal)
		shortGoal := strings.TrimSpace(row.ShortGoal)
		if domain == "" && longGoal == "" && shortGoal == "" {
			continue
		}
		rows = append(rows, iepPlanWordRow{
			Domain:       firstNonEmptyExportValue(domain, "综合康复"),
			LongGoal:     longGoal,
			ShortGoal:    shortGoal,
			CourseForm:   firstNonEmptyExportValue(strings.TrimSpace(row.CourseForm), "个训"),
			StartEndDate: firstNonEmptyExportValue(strings.TrimSpace(row.StartEndDate), startDate+" - "+endDate),
		})
	}
	return iepPlanWordExport{
		Title:         firstNonEmptyExportValue(strings.TrimSpace(result.Title), iepPlanWordTitle(durationMonths)),
		StudentName:   firstNonEmptyExportValue(strings.TrimSpace(result.Student.Name), "学员"),
		Gender:        strings.TrimSpace(result.Student.Gender),
		BirthDate:     strings.TrimSpace(result.Student.BirthDate),
		PlanDate:      firstNonEmptyExportValue(strings.TrimSpace(result.Meta.PlanDate), time.Now().Format("2006-01-02")),
		Participant:   firstNonEmptyExportValue(strings.TrimSpace(result.Meta.Participant), "1"),
		PlannerName:   strings.TrimSpace(result.Meta.Implementer),
		StartDate:     startDate,
		EndDate:       endDate,
		DurationLabel: fmt.Sprintf("%d个月", durationMonths),
		Rows:          rows,
	}
}

func iepPlanWordTitle(durationMonths int) string {
	if durationMonths == 3 {
		return "康复教学季度计划"
	}
	return "康复教学半年计划"
}

func buildPEP3IEPPlanWordDocx(plan iepPlanWordExport) ([]byte, error) {
	if len(plan.Domains) == 0 && len(plan.Rows) == 0 {
		return nil, errors.New("暂无可导出的IEP训练计划")
	}

	entries := map[string][]byte{
		"[Content_Types].xml":          []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/></Types>`),
		"_rels/.rels":                  []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>`),
		"word/_rels/document.xml.rels": []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"></Relationships>`),
		"word/document.xml":            []byte(buildPEP3IEPPlanDocumentXML(plan)),
	}
	return writeDocxZipEntries(entries)
}

func buildPEP3IEPPlanDocumentXML(plan iepPlanWordExport) string {
	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	builder.WriteString(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">`)
	builder.WriteString(`<w:body>`)
	builder.WriteString(buildIEPTitleParagraph(firstNonEmptyExportValue(plan.Title, "康复教学计划")))
	builder.WriteString(buildIEPMainTable(plan))
	if len(plan.HomePlan) > 0 {
		builder.WriteString(buildIEPPageBreakParagraph())
		builder.WriteString(buildIEPHomePlanTable(plan.HomePlan))
	}
	builder.WriteString(`<w:sectPr><w:pgSz w:w="11906" w:h="16838"/><w:pgMar w:top="900" w:right="900" w:bottom="900" w:left="900" w:header="851" w:footer="992" w:gutter="0"/></w:sectPr>`)
	builder.WriteString(`</w:body></w:document>`)
	return builder.String()
}

func buildIEPMainTable(plan iepPlanWordExport) string {
	widths := []int{1038, 1472, 625, 877, 1260, 1562, 927, 2319}
	rows := plan.Rows
	if len(rows) == 0 {
		rows = buildIEPWordRowsFromDomains(plan)
	}
	var builder strings.Builder
	builder.WriteString(buildIEPTableStart(widths))
	builder.WriteString(buildIEPTableRowStart(560))
	builder.WriteString(buildIEPCell([]string{"姓名"}, widths[0], iepPlanWordCellOptions{Align: "center", VAlign: "center", Bold: true, CompactParagraph: true}))
	builder.WriteString(buildIEPCell([]string{plan.StudentName}, widths[1], iepPlanWordCellOptions{Align: "center", VAlign: "center"}))
	builder.WriteString(buildIEPCell([]string{"性别"}, widths[2], iepPlanWordCellOptions{Align: "center", VAlign: "center", Bold: true}))
	builder.WriteString(buildIEPCell([]string{plan.Gender}, widths[3], iepPlanWordCellOptions{Align: "center", VAlign: "center"}))
	builder.WriteString(buildIEPCell([]string{"出生年月"}, widths[4], iepPlanWordCellOptions{Align: "center", VAlign: "center", Bold: true}))
	builder.WriteString(buildIEPCell([]string{plan.BirthDate}, widths[5]+widths[6]+widths[7], iepPlanWordCellOptions{GridSpan: 3, Align: "center", VAlign: "center"}))
	builder.WriteString(`</w:tr>`)
	builder.WriteString(buildIEPTableRowStart(560))
	builder.WriteString(buildIEPCell([]string{"制定日期"}, widths[0], iepPlanWordCellOptions{Align: "center", VAlign: "center", Bold: true}))
	builder.WriteString(buildIEPCell([]string{plan.PlanDate}, widths[1]+widths[2]+widths[3], iepPlanWordCellOptions{GridSpan: 3, Align: "center", VAlign: "center"}))
	builder.WriteString(buildIEPCell([]string{"计划参与者"}, widths[4], iepPlanWordCellOptions{Align: "center", VAlign: "center", Bold: true}))
	builder.WriteString(buildIEPCell([]string{plan.Participant}, widths[5]+widths[6]+widths[7], iepPlanWordCellOptions{GridSpan: 3, Align: "center", VAlign: "center"}))
	builder.WriteString(`</w:tr>`)
	builder.WriteString(buildIEPTableRowStart(560))
	builder.WriteString(buildIEPCell([]string{"实施者"}, widths[0], iepPlanWordCellOptions{Align: "center", VAlign: "center", Bold: true}))
	builder.WriteString(buildIEPCell([]string{plan.PlannerName}, widths[1]+widths[2]+widths[3], iepPlanWordCellOptions{GridSpan: 3, Align: "center", VAlign: "center"}))
	builder.WriteString(buildIEPCell([]string{"实施", "起止日期"}, widths[4], iepPlanWordCellOptions{Align: "center", VAlign: "center", Bold: true, CompactParagraph: true}))
	builder.WriteString(buildIEPCell([]string{plan.StartDate + " 至 " + plan.EndDate}, widths[5]+widths[6]+widths[7], iepPlanWordCellOptions{GridSpan: 3, Align: "center", VAlign: "center", NoWrap: true, CompactParagraph: true}))
	builder.WriteString(`</w:tr>`)
	builder.WriteString(buildIEPTableRowStart(560))
	builder.WriteString(buildIEPCell([]string{"康复", "领域"}, widths[0], iepPlanWordCellOptions{Align: "center", VAlign: "center", Bold: true, CompactParagraph: true}))
	builder.WriteString(buildIEPCell([]string{"长期目标"}, widths[1]+widths[2]+widths[3], iepPlanWordCellOptions{GridSpan: 3, Align: "center", VAlign: "center", Bold: true}))
	builder.WriteString(buildIEPCell([]string{"短期目标"}, widths[4]+widths[5], iepPlanWordCellOptions{GridSpan: 2, Align: "center", VAlign: "center", Bold: true}))
	builder.WriteString(buildIEPCell([]string{"课程", "形式"}, widths[6], iepPlanWordCellOptions{Align: "center", VAlign: "center", Bold: true, CompactParagraph: true}))
	builder.WriteString(buildIEPCell([]string{"起止日期"}, widths[7], iepPlanWordCellOptions{Align: "center", VAlign: "center", NoWrap: true, Bold: true, CompactParagraph: true}))
	builder.WriteString(`</w:tr>`)

	for index, row := range rows {
		rowSpan := 0
		if index == 0 || row.Domain != rows[index-1].Domain {
			for next := index; next < len(rows) && rows[next].Domain == row.Domain; next++ {
				rowSpan++
			}
		}
		builder.WriteString(buildIEPTableRowStart(660))
		if rowSpan > 0 {
			builder.WriteString(buildIEPCell(domainTextLines(row.Domain), widths[0], iepPlanWordCellOptions{VMerge: "restart", Align: "center", VAlign: "center", Bold: true, CompactParagraph: true}))
			builder.WriteString(buildIEPCell(splitWordLines(row.LongGoal), widths[1]+widths[2]+widths[3], iepPlanWordCellOptions{GridSpan: 3, VMerge: "restart", VAlign: "center", IndentLeft: 200}))
		} else {
			builder.WriteString(buildIEPCell(nil, widths[0], iepPlanWordCellOptions{VMerge: "continue"}))
			builder.WriteString(buildIEPCell(nil, widths[1]+widths[2]+widths[3], iepPlanWordCellOptions{GridSpan: 3, VMerge: "continue"}))
		}
		builder.WriteString(buildIEPCell(splitWordLines(row.ShortGoal), widths[4]+widths[5], iepPlanWordCellOptions{GridSpan: 2, VAlign: "center", IndentLeft: 200}))
		builder.WriteString(buildIEPCell([]string{row.CourseForm}, widths[6], iepPlanWordCellOptions{Align: "center", VAlign: "center", NoWrap: true, CompactParagraph: true}))
		builder.WriteString(buildIEPCell([]string{row.StartEndDate}, widths[7], iepPlanWordCellOptions{Align: "center", VAlign: "center", NoWrap: true, CompactParagraph: true}))
		builder.WriteString(`</w:tr>`)
	}
	builder.WriteString(`</w:tbl>`)
	return builder.String()
}

func buildIEPWordRowsFromDomains(plan iepPlanWordExport) []iepPlanWordRow {
	rows := make([]iepPlanWordRow, 0)
	stageRanges := iepPlanWordStageRanges(plan.StartDate, plan.EndDate, 3)
	for _, domain := range plan.Domains {
		shortGoals := shortGoalLines(domain.StageGoals)
		if len(shortGoals) == 0 {
			shortGoals = []string{""}
		}
		longGoal := strings.Join(numberedLines(domain.LongGoals), "\n")
		for index, shortGoal := range shortGoals {
			rows = append(rows, iepPlanWordRow{
				Domain:       domain.Name,
				LongGoal:     longGoal,
				ShortGoal:    shortGoal,
				CourseForm:   inferStaticIEPWordCourseForm(domain.Name, shortGoal),
				StartEndDate: stageDateForWordGoal(stageRanges, index, len(shortGoals)),
			})
		}
	}
	return rows
}

func buildIEPHomePlanTable(items []string) string {
	widths := []int{iepPlanWordTableWidth}
	lines := numberedLines(items)
	var builder strings.Builder
	builder.WriteString(buildIEPSpacingParagraph(120))
	builder.WriteString(buildIEPTableStart(widths))
	builder.WriteString(`<w:tr>`)
	builder.WriteString(buildIEPCell([]string{"家庭干预计划"}, widths[0], iepPlanWordCellOptions{Align: "center", VAlign: "center", Bold: true}))
	builder.WriteString(`</w:tr>`)
	builder.WriteString(`<w:tr>`)
	builder.WriteString(buildIEPCell(lines, widths[0], iepPlanWordCellOptions{VAlign: "top", IndentLeft: 140}))
	builder.WriteString(`</w:tr>`)
	builder.WriteString(`</w:tbl>`)
	return builder.String()
}

func buildIEPTableStart(widths []int) string {
	var builder strings.Builder
	builder.WriteString(`<w:tbl><w:tblPr><w:tblW w:w="`)
	builder.WriteString(strconv.Itoa(sumInts(widths...)))
	builder.WriteString(`" w:type="dxa"/><w:tblLayout w:type="fixed"/><w:tblBorders><w:top w:val="single" w:sz="8" w:space="0" w:color="000000"/><w:left w:val="single" w:sz="8" w:space="0" w:color="000000"/><w:bottom w:val="single" w:sz="8" w:space="0" w:color="000000"/><w:right w:val="single" w:sz="8" w:space="0" w:color="000000"/><w:insideH w:val="single" w:sz="8" w:space="0" w:color="000000"/><w:insideV w:val="single" w:sz="8" w:space="0" w:color="000000"/></w:tblBorders></w:tblPr><w:tblGrid>`)
	for _, width := range widths {
		builder.WriteString(`<w:gridCol w:w="`)
		builder.WriteString(strconv.Itoa(width))
		builder.WriteString(`"/>`)
	}
	builder.WriteString(`</w:tblGrid>`)
	return builder.String()
}

func buildIEPTableRowStart(height int) string {
	if height <= 0 {
		return `<w:tr>`
	}
	return `<w:tr><w:trPr><w:trHeight w:val="` + strconv.Itoa(height) + `" w:hRule="atLeast"/></w:trPr>`
}

func buildIEPCell(lines []string, width int, options iepPlanWordCellOptions) string {
	var builder strings.Builder
	builder.WriteString(`<w:tc><w:tcPr><w:tcW w:w="`)
	builder.WriteString(strconv.Itoa(width))
	builder.WriteString(`" w:type="dxa"/>`)
	if options.GridSpan > 1 {
		builder.WriteString(`<w:gridSpan w:val="`)
		builder.WriteString(strconv.Itoa(options.GridSpan))
		builder.WriteString(`"/>`)
	}
	switch options.VMerge {
	case "restart":
		builder.WriteString(`<w:vMerge w:val="restart"/>`)
	case "continue":
		builder.WriteString(`<w:vMerge/>`)
	}
	if options.VAlign != "" {
		builder.WriteString(`<w:vAlign w:val="`)
		builder.WriteString(options.VAlign)
		builder.WriteString(`"/>`)
	}
	if options.NoWrap {
		builder.WriteString(`<w:noWrap/>`)
	}
	builder.WriteString(`</w:tcPr>`)
	if len(lines) == 0 {
		builder.WriteString(buildIEPParagraph("", options.Align, options.Bold, 21, options))
	}
	for _, line := range lines {
		builder.WriteString(buildIEPParagraph(line, options.Align, options.Bold, 21, options))
	}
	builder.WriteString(`</w:tc>`)
	return builder.String()
}

func buildIEPTitleParagraph(text string) string {
	return `<w:p><w:pPr><w:jc w:val="center"/><w:spacing w:after="240"/>` + iepParagraphRunPropsXML(true, 32) + `</w:pPr>` + buildIEPTextRun(text, true, 32) + `</w:p>`
}

func buildIEPNormalParagraph(text, align string, bold bool) string {
	return buildIEPParagraph(text, align, bold, 20, iepPlanWordCellOptions{})
}

func buildIEPSpacingParagraph(after int) string {
	return `<w:p><w:pPr><w:spacing w:after="` + strconv.Itoa(after) + `"/></w:pPr></w:p>`
}

func buildIEPPageBreakParagraph() string {
	return `<w:p><w:r><w:br w:type="page"/></w:r></w:p>`
}

func buildIEPParagraph(text, align string, bold bool, size int, options iepPlanWordCellOptions) string {
	if align == "" {
		align = "left"
	}
	before := 40
	after := 40
	lineSpacing := 300
	if options.CompactParagraph {
		before = 0
		after = 0
		lineSpacing = 240
	}
	if options.SpacingBefore > 0 {
		before = options.SpacingBefore
	}
	if options.SpacingAfter > 0 {
		after = options.SpacingAfter
	}
	if options.LineSpacing > 0 {
		lineSpacing = options.LineSpacing
	}
	var builder strings.Builder
	builder.WriteString(`<w:p><w:pPr><w:jc w:val="`)
	builder.WriteString(align)
	builder.WriteString(`"/><w:spacing w:before="`)
	builder.WriteString(strconv.Itoa(before))
	builder.WriteString(`" w:after="`)
	builder.WriteString(strconv.Itoa(after))
	builder.WriteString(`" w:line="`)
	builder.WriteString(strconv.Itoa(lineSpacing))
	builder.WriteString(`" w:lineRule="auto"/>`)
	if options.IndentLeft > 0 {
		builder.WriteString(`<w:ind w:left="`)
		builder.WriteString(strconv.Itoa(options.IndentLeft))
		builder.WriteString(`"/>`)
	}
	builder.WriteString(iepParagraphRunPropsXML(bold, size))
	builder.WriteString(`</w:pPr>`)
	builder.WriteString(buildIEPTextRun(text, bold, size))
	builder.WriteString(`</w:p>`)
	return builder.String()
}

func buildIEPTextRun(text string, bold bool, size int) string {
	if strings.TrimSpace(text) == "" {
		text = " "
	}
	var builder strings.Builder
	builder.WriteString(`<w:r>`)
	builder.WriteString(iepRunPropsXML(bold, size))
	builder.WriteString(`<w:t`)
	if shouldPreserveWordTextSpaces(text) {
		builder.WriteString(` xml:space="preserve"`)
	}
	builder.WriteString(`>`)
	builder.WriteString(escapeXMLText(text))
	builder.WriteString(`</w:t></w:r>`)
	return builder.String()
}

func iepParagraphRunPropsXML(bold bool, size int) string {
	return `<w:rPr>` + iepRunPropsInnerXML(bold, size) + `</w:rPr>`
}

func iepRunPropsXML(bold bool, size int) string {
	return `<w:rPr>` + iepRunPropsInnerXML(bold, size) + `</w:rPr>`
}

func iepRunPropsInnerXML(bold bool, size int) string {
	if size <= 0 {
		size = 21
	}
	var builder strings.Builder
	builder.WriteString(`<w:rFonts w:ascii="Times New Roman" w:hAnsi="Times New Roman" w:eastAsia="宋体"/><w:sz w:val="`)
	builder.WriteString(strconv.Itoa(size))
	builder.WriteString(`"/><w:szCs w:val="`)
	builder.WriteString(strconv.Itoa(size))
	builder.WriteString(`"/><w:lang w:val="en-US" w:eastAsia="zh-CN"/>`)
	if bold {
		builder.WriteString(`<w:b/><w:bCs/>`)
	}
	return builder.String()
}

func numberedLines(items []string) []string {
	lines := make([]string, 0, len(items))
	for index, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("%d. %s", index+1, item))
	}
	return lines
}

func shortGoalLines(stages []iepPlanWordStage) []string {
	lines := make([]string, 0, len(stages)*4)
	goalNumber := 1
	for _, stage := range stages {
		for _, goal := range stage.Goals {
			goal = strings.TrimSpace(goal)
			if goal == "" {
				continue
			}
			lines = append(lines, fmt.Sprintf("%d. %s", goalNumber, goal))
			goalNumber++
		}
	}
	return lines
}

func domainNameLines(index int, name string) []string {
	lines := []string{fmt.Sprintf("%d.", index+1)}
	lines = append(lines, verticalCharacterLines(name)...)
	return lines
}

func domainTextLines(name string) []string {
	text := strings.TrimSpace(name)
	if text == "" {
		return []string{"综合康复"}
	}
	return []string{text}
}

func splitWordLines(text string) []string {
	value := strings.TrimSpace(text)
	if value == "" {
		return nil
	}
	rawLines := strings.Split(strings.NewReplacer("\r\n", "\n", "\r", "\n").Replace(value), "\n")
	lines := make([]string, 0, len(rawLines))
	for _, line := range rawLines {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func iepPlanWordStageRanges(startText, endText string, stageCount int) []string {
	if stageCount <= 0 {
		stageCount = 3
	}
	start, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(startText), time.Local)
	if err != nil {
		return nil
	}
	end, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(endText), time.Local)
	if err != nil || end.Before(start) {
		return nil
	}
	totalDays := int(end.Sub(start).Hours()/24) + 1
	ranges := make([]string, 0, stageCount)
	current := start
	for index := 0; index < stageCount; index++ {
		days := totalDays / stageCount
		if index < totalDays%stageCount {
			days++
		}
		if days <= 0 {
			days = 1
		}
		stageEnd := current.AddDate(0, 0, days-1)
		if stageEnd.After(end) {
			stageEnd = end
		}
		ranges = append(ranges, current.Format("2006-01-02")+" - "+stageEnd.Format("2006-01-02"))
		current = stageEnd.AddDate(0, 0, 1)
		if current.After(end) {
			break
		}
	}
	return ranges
}

func stageDateForWordGoal(stageRanges []string, goalIndex int, goalCount int) string {
	if len(stageRanges) == 0 {
		return ""
	}
	rangeIndex := goalIndex * len(stageRanges) / maxIEPPlanWordInt(goalCount, 1)
	if rangeIndex >= len(stageRanges) {
		rangeIndex = len(stageRanges) - 1
	}
	return stageRanges[rangeIndex]
}

func inferStaticIEPWordCourseForm(domainName, goal string) string {
	text := domainName + " " + goal
	switch {
	case strings.Contains(text, "同伴"), strings.Contains(text, "小组"), strings.Contains(text, "集体"), strings.Contains(text, "融合"):
		return "集体课"
	default:
		return "个训"
	}
}

func maxIEPPlanWordInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func verticalCharacterLines(text string) []string {
	value := strings.TrimSpace(text)
	if value == "" {
		return nil
	}
	lines := make([]string, 0, len([]rune(value)))
	for _, char := range value {
		part := strings.TrimSpace(string(char))
		if part == "" {
			continue
		}
		lines = append(lines, part)
	}
	return lines
}

func sumInts(values ...int) int {
	total := 0
	for _, value := range values {
		total += value
	}
	return total
}

func sanitizeExportFileName(value string) string {
	text := strings.TrimSpace(value)
	if text == "" {
		return "学员"
	}
	replacer := strings.NewReplacer("\\", "", "/", "", ":", "", "*", "", "?", "", "\"", "", "<", "", ">", "", "|", "")
	return replacer.Replace(text)
}
