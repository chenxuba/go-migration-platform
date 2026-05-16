package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

type autismDevResultAnalysisWordExport struct {
	Title          string
	StudentName    string
	Gender         string
	BirthDate      string
	Age            string
	AssessmentDate string
	ExaminerName   string
	AssessmentName string
	Rows           []model.AutismDevResultAnalysisRow
}

func (svc *Service) ExportAutismDevResultAnalysisWord(userID int64, recordID int64, analysis *model.AutismDevResultAnalysisVO) (string, string, []byte, error) {
	record, normalized, err := svc.autismDevResultAnalysisForExport(userID, recordID, analysis)
	if err != nil {
		return "", "", nil, err
	}
	export := autismDevResultAnalysisWordExport{
		Title:          firstNonEmptyExportValue(strings.TrimSpace(normalized.Title), "孤独症儿童评估结果分析表"),
		StudentName:    strings.TrimSpace(record.StudentName),
		Gender:         strings.TrimSpace(record.StudentGender),
		BirthDate:      formatIEPPlanDate(record.BirthDate),
		Age:            formatIEPPlanAge(record.AgeYears, record.AgeMonths, record.AgeDays),
		AssessmentDate: formatIEPPlanDate(record.AssessmentDate),
		ExaminerName:   strings.TrimSpace(record.ExaminerName),
		AssessmentName: firstNonEmptyExportValue(strings.TrimSpace(record.AssessmentName), "孤独症儿童发展评估表"),
		Rows:           normalized.Rows,
	}
	content, err := buildAutismDevResultAnalysisWordDocx(export)
	if err != nil {
		return "", "", nil, err
	}
	fileName := fmt.Sprintf("%s-孤独症儿童评估结果分析-%s.docx", sanitizeExportFileName(export.StudentName), time.Now().Format("20060102150405"))
	return fileName, iepPlanWordContentType, content, nil
}

func (svc *Service) ExportAutismDevResultAnalysisPDF(userID int64, recordID int64, analysis *model.AutismDevResultAnalysisVO) (string, string, []byte, error) {
	fileName, _, content, err := svc.ExportAutismDevResultAnalysisWord(userID, recordID, analysis)
	if err != nil {
		return "", "", nil, err
	}
	return exportIEPPDFByDOCX(fileName, content)
}

func (svc *Service) autismDevResultAnalysisForExport(userID int64, recordID int64, analysis *model.AutismDevResultAnalysisVO) (model.AssessmentRecordDetailVO, model.AutismDevResultAnalysisVO, error) {
	_, record, score, data, itemScores, err := svc.autismDevResultAnalysisContext(userID, recordID)
	if err != nil {
		return model.AssessmentRecordDetailVO{}, model.AutismDevResultAnalysisVO{}, err
	}
	if analysis == nil || len(analysis.Rows) == 0 {
		saved, err := svc.GetAutismDevResultAnalysis(userID, recordID)
		if err != nil {
			return model.AssessmentRecordDetailVO{}, model.AutismDevResultAnalysisVO{}, err
		}
		analysis = &saved
	}
	if autismDevResultAnalysisIsEmpty(*analysis) {
		return model.AssessmentRecordDetailVO{}, model.AutismDevResultAnalysisVO{}, errors.New("请先生成评估结果分析")
	}
	normalized := normalizeAutismDevResultAnalysis(*analysis, record, score, data, itemScores, analysis.GeneratedBy)
	return record, normalized, nil
}

func autismDevResultAnalysisIsEmpty(analysis model.AutismDevResultAnalysisVO) bool {
	if len(analysis.Rows) == 0 {
		return true
	}
	for _, row := range analysis.Rows {
		if strings.TrimSpace(row.Status) != "" ||
			strings.TrimSpace(row.Strengths) != "" ||
			strings.TrimSpace(row.Weaknesses) != "" ||
			strings.TrimSpace(row.Targets) != "" {
			return false
		}
	}
	return true
}

func buildAutismDevResultAnalysisWordDocx(export autismDevResultAnalysisWordExport) ([]byte, error) {
	if len(export.Rows) == 0 {
		return nil, errors.New("暂无可导出的评估结果分析")
	}
	entries := map[string][]byte{
		"[Content_Types].xml":          []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/></Types>`),
		"_rels/.rels":                  []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>`),
		"word/_rels/document.xml.rels": []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"></Relationships>`),
		"word/document.xml":            []byte(buildAutismDevResultAnalysisDocumentXML(export)),
	}
	return writeDocxZipEntries(entries)
}

func buildAutismDevResultAnalysisDocumentXML(export autismDevResultAnalysisWordExport) string {
	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	builder.WriteString(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">`)
	builder.WriteString(`<w:body>`)
	builder.WriteString(buildAutismDevResultAnalysisTitleParagraph(firstNonEmptyExportValue(export.Title, "孤独症儿童评估结果分析表")))
	builder.WriteString(buildAutismDevResultAnalysisMetaParagraph(export))
	builder.WriteString(buildAutismDevResultAnalysisTable(export.Rows))
	builder.WriteString(`<w:sectPr><w:pgSz w:w="11906" w:h="16838"/><w:pgMar w:top="720" w:right="720" w:bottom="720" w:left="720" w:header="851" w:footer="720" w:gutter="0"/></w:sectPr>`)
	builder.WriteString(`</w:body></w:document>`)
	return builder.String()
}

func buildAutismDevResultAnalysisTitleParagraph(text string) string {
	return buildIEPParagraph(text, "center", false, 32, iepPlanWordCellOptions{
		CompactParagraph: true,
		SpacingAfter:     160,
		LineSpacing:      260,
	})
}

func buildAutismDevResultAnalysisMetaParagraph(export autismDevResultAnalysisWordExport) string {
	var builder strings.Builder
	builder.WriteString(`<w:p><w:pPr><w:jc w:val="left"/><w:tabs><w:tab w:val="left" w:pos="3600"/><w:tab w:val="left" w:pos="7000"/></w:tabs><w:spacing w:before="0" w:after="0" w:line="240" w:lineRule="auto"/>`)
	builder.WriteString(iepParagraphRunPropsXML(false, 21))
	builder.WriteString(`</w:pPr>`)
	builder.WriteString(autismDevAnalysisMetaLabelRun("儿童姓名："))
	builder.WriteString(autismDevAnalysisMetaValueRun(export.StudentName, 10))
	builder.WriteString(autismDevAnalysisMetaTabRun())
	builder.WriteString(autismDevAnalysisMetaLabelRun("评估者："))
	builder.WriteString(autismDevAnalysisMetaValueRun(export.ExaminerName, 10))
	builder.WriteString(autismDevAnalysisMetaTabRun())
	builder.WriteString(autismDevAnalysisMetaLabelRun("评估时间："))
	builder.WriteString(autismDevAnalysisMetaValueRun(export.AssessmentDate, 12))
	builder.WriteString(`</w:p>`)
	return builder.String()
}

func autismDevAnalysisMetaLabelRun(text string) string {
	return `<w:r>` + iepRunPropsXML(false, 21) + `<w:t>` + escapeXMLText(text) + `</w:t></w:r>`
}

func autismDevAnalysisMetaValueRun(text string, minChars int) string {
	value := strings.TrimSpace(text)
	if minChars > 0 {
		padding := minChars - len([]rune(value))
		if padding > 0 {
			value += strings.Repeat(" ", padding)
		}
	}
	return `<w:r><w:rPr>` + iepRunPropsInnerXML(false, 20) + `<w:u w:val="single"/></w:rPr><w:t xml:space="preserve">` + escapeXMLText(value) + `</w:t></w:r>`
}

func autismDevAnalysisMetaTabRun() string {
	return `<w:r><w:tab/></w:r>`
}

func buildAutismDevResultAnalysisTable(rows []model.AutismDevResultAnalysisRow) string {
	widths := []int{1050, 2500, 3450, 3080}
	var builder strings.Builder
	builder.WriteString(buildIEPTableStart(widths))
	builder.WriteString(buildIEPTableRowStart(560))
	builder.WriteString(buildIEPCell([]string{"领   域"}, widths[0], autismDevAnalysisHeaderCellOptions()))
	builder.WriteString(buildIEPCell([]string{"能力现状描述"}, widths[1], autismDevAnalysisHeaderCellOptions()))
	builder.WriteString(buildIEPCell([]string{"优劣分析"}, widths[2], autismDevAnalysisHeaderCellOptions()))
	builder.WriteString(buildIEPCell([]string{"训练目标"}, widths[3], autismDevAnalysisHeaderCellOptions()))
	builder.WriteString(`</w:tr>`)
	for _, row := range rows {
		row = trimAutismDevResultAnalysisRow(row)
		builder.WriteString(buildIEPTableRowStart(1680))
		builder.WriteString(buildIEPCell(domainTextLines(row.Domain), widths[0], iepPlanWordCellOptions{Align: "center", VAlign: "center", CompactParagraph: true, LineSpacing: 220, FontSize: 20}))
		builder.WriteString(buildIEPCell(splitWordLines(row.Status), widths[1], autismDevAnalysisBodyCellOptions()))
		builder.WriteString(buildIEPCell(autismDevStrengthWeaknessLines(row), widths[2], autismDevAnalysisBodyCellOptions()))
		builder.WriteString(buildIEPCell(splitWordLines(row.Targets), widths[3], autismDevAnalysisBodyCellOptions()))
		builder.WriteString(`</w:tr>`)
	}
	builder.WriteString(`</w:tbl>`)
	return builder.String()
}

func autismDevAnalysisHeaderCellOptions() iepPlanWordCellOptions {
	return iepPlanWordCellOptions{
		Align:            "center",
		VAlign:           "center",
		CompactParagraph: true,
		LineSpacing:      220,
		FontSize:         20,
	}
}

func autismDevAnalysisBodyCellOptions() iepPlanWordCellOptions {
	return iepPlanWordCellOptions{
		VAlign:        "top",
		IndentLeft:    80,
		PaddingLeft:   60,
		PaddingRight:  60,
		SpacingBefore: 0,
		SpacingAfter:  0,
		LineSpacing:   230,
	}
}

func autismDevStrengthWeaknessLines(row model.AutismDevResultAnalysisRow) []string {
	lines := make([]string, 0, 4)
	if text := strings.TrimSpace(row.Strengths); text != "" {
		lines = append(lines, "优势："+text)
	}
	if strings.TrimSpace(row.Strengths) != "" && strings.TrimSpace(row.Weaknesses) != "" {
		lines = append(lines, "")
	}
	if text := strings.TrimSpace(row.Weaknesses); text != "" {
		lines = append(lines, "劣势："+text)
	}
	if len(lines) == 0 {
		return nil
	}
	return lines
}
