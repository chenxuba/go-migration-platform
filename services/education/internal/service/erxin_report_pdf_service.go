package service

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/signintech/gopdf"
	"go-migration-platform/services/education/internal/model"
)

const (
	erxinReportPDFPageWidth  = 595.28
	erxinReportPDFPageHeight = 841.89
	erxinReportPDFFontFamily = "erxin-cjk"
)

func (svc *Service) GenerateERXinAssessmentReportPDF(userID, recordID int64) (string, []byte, error) {
	report, err := svc.GetERXinAssessmentReport(userID, recordID)
	if err != nil {
		return "", nil, err
	}
	content, err := buildERXinAssessmentReportPDF(report)
	if err != nil {
		return "", nil, err
	}
	name := nonEmptyString(report.Record.StudentName, "未命名儿童")
	filename := sanitizeTemplateFileName(fmt.Sprintf("%s-儿心量表评估报告-%s.pdf", name, time.Now().Format("20060102150405")))
	return filename, content, nil
}

func buildERXinAssessmentReportPDF(report model.ERXinReportVO) ([]byte, error) {
	fontBytes, err := loadPEP3PDFFontBytes()
	if err != nil {
		return nil, err
	}

	var pdf gopdf.GoPdf
	pdf.Start(gopdf.Config{
		Unit:     gopdf.UnitPT,
		PageSize: gopdf.Rect{W: erxinReportPDFPageWidth, H: erxinReportPDFPageHeight},
	})
	pdf.AddPage()
	if err := pdf.AddTTFFontByReader(erxinReportPDFFontFamily, bytes.NewReader(fontBytes)); err != nil {
		return nil, fmt.Errorf("load ERXin PDF font: %w", err)
	}

	renderer := erxinReportPDFRenderer{
		pdf:             &pdf,
		currentFontSize: 11,
	}
	renderer.draw(report)
	return pdf.GetBytesPdfReturnErr()
}

type erxinReportPDFRenderer struct {
	pdf             *gopdf.GoPdf
	currentFontSize float64
}

func (r erxinReportPDFRenderer) draw(report model.ERXinReportVO) {
	r.drawHeader(report)
	r.drawOriginalResultTable(report)
	r.drawExaminer(report)
}

func (r erxinReportPDFRenderer) drawHeader(report model.ERXinReportVO) {
	r.setTextColor(0, 0, 0)
	r.setFont(14)
	r.drawText(erxinReportPDFPageWidth-155, 42, erxinReportPDFStandardText(report))

	r.setFont(14)
	r.centerText(
		0,
		92,
		erxinReportPDFPageWidth,
		"0 岁～6 岁儿童发育行为评估量表（儿心量表-Ⅱ）基本信息和结果记录",
	)
}

func (r erxinReportPDFRenderer) drawOriginalResultTable(report model.ERXinReportVO) {
	const (
		left       = 67.0
		top        = 122.0
		labelWidth = 96.0
		tableWidth = 450.0
	)
	rowHeights := []float64{30, 25, 25, 25, 30, 27, 27, 27, 27, 27, 27}
	totalHeight := 0.0
	for _, height := range rowHeights {
		totalHeight += height
	}

	r.setTextColor(0, 0, 0)
	r.pdf.SetStrokeColor(0, 0, 0)
	r.pdf.SetLineWidth(1)
	r.pdf.RectFromUpperLeft(left, top, tableWidth, totalHeight)
	y := top
	for _, height := range rowHeights[:len(rowHeights)-1] {
		y += height
		r.pdf.Line(left, y, left+tableWidth, y)
	}

	topRowColumns := []float64{labelWidth, 96, 64, 64, 64, 66}
	for _, x := range erxinReportPDFColumnLines(left, topRowColumns) {
		r.pdf.Line(x, top, x, top+rowHeights[0])
	}
	mainRowTop := top + rowHeights[0]
	for row := 1; row <= 3; row++ {
		rowTop := top + erxinReportPDFRowsOffset(rowHeights, row)
		r.pdf.Line(left+labelWidth, rowTop, left+labelWidth, rowTop+rowHeights[row])
	}
	resultRowTop := mainRowTop + rowHeights[1] + rowHeights[2] + rowHeights[3]
	resultColumns := []float64{labelWidth, 160, 194}
	for _, x := range erxinReportPDFColumnLines(left, resultColumns) {
		r.pdf.Line(x, resultRowTop, x, top+totalHeight)
	}

	r.drawOriginalCellText(left, top, labelWidth, rowHeights[0], "姓    名", 12, true)
	r.drawOriginalCellText(left+labelWidth, top, 96, rowHeights[0], report.Record.StudentName, 11, true)
	r.drawOriginalCellText(left+192, top, 64, rowHeights[0], "性    别", 12, true)
	r.drawOriginalCellText(left+256, top, 64, rowHeights[0], erxinBlankDash(report.Record.StudentGender), 11, true)
	r.drawOriginalCellText(left+320, top, 64, rowHeights[0], "民    族", 12, true)

	dateRowTop := top + rowHeights[0]
	r.drawOriginalCellText(left, dateRowTop, labelWidth, rowHeights[1], "测验日期", 12, true)
	r.drawOriginalCellText(left+labelWidth, dateRowTop, tableWidth-labelWidth, rowHeights[1], erxinReportPDFDateText(report.Record.AssessmentDate), 11, true)

	birthRowTop := dateRowTop + rowHeights[1]
	r.drawOriginalCellText(left, birthRowTop, labelWidth, rowHeights[2], "出生日期", 12, true)
	r.drawOriginalCellText(left+labelWidth, birthRowTop, tableWidth-labelWidth, rowHeights[2], erxinReportPDFDateText(report.Record.BirthDate), 11, true)

	ageRowTop := birthRowTop + rowHeights[2]
	r.drawOriginalCellText(left, ageRowTop, labelWidth, rowHeights[3], "实足年龄", 12, true)
	r.drawOriginalCellText(left+labelWidth, ageRowTop, tableWidth-labelWidth, rowHeights[3], erxinActualAgeSummaryText(report.Record), 11, true)

	headerRowTop := ageRowTop + rowHeights[3]
	r.drawOriginalCellText(left, headerRowTop, labelWidth, rowHeights[4], "项    目", 12, true)
	r.drawOriginalCellText(left+labelWidth, headerRowTop, resultColumns[1], rowHeights[4], "智  龄（月）", 12, true)
	r.drawOriginalCellText(left+labelWidth+resultColumns[1], headerRowTop, resultColumns[2], rowHeights[4], "发育商（DQ）", 12, true)

	domains := []string{"大 运 动", "精细动作", "适应能力", "语    言", "社会行为"}
	domainKeys := []string{"GM", "FM", "AD", "LANG", "SOC"}
	rowsByKey := erxinReportPDFDomainRowsByKey(report.DomainRows)
	for index, label := range domains {
		rowTop := headerRowTop + rowHeights[4] + float64(index)*rowHeights[5]
		row := rowsByKey[domainKeys[index]]
		r.drawOriginalCellText(left, rowTop, labelWidth, rowHeights[5], label, 12, true)
		r.drawOriginalCellText(left+labelWidth, rowTop, resultColumns[1], rowHeights[5], erxinReportPDFMentalAgeText(row), 11, true)
		r.drawOriginalCellText(left+labelWidth+resultColumns[1], rowTop, resultColumns[2], rowHeights[5], erxinReportPDFDQText(row.DQ), 11, true)
	}

	totalRowTop := headerRowTop + rowHeights[4] + 5*rowHeights[5]
	r.drawOriginalCellText(left, totalRowTop, labelWidth, rowHeights[10], "全 量 表", 12, true)
	r.drawOriginalCellText(left+labelWidth, totalRowTop, resultColumns[1], rowHeights[10], erxinReportPDFSummaryMentalAgeText(report.Summary), 11, true)
	r.drawOriginalCellText(left+labelWidth+resultColumns[1], totalRowTop, resultColumns[2], rowHeights[10], erxinReportPDFDQText(report.Summary.DQ), 11, true)
}

func (r erxinReportPDFRenderer) drawExaminer(report model.ERXinReportVO) {
	text := "主试者："
	if examiner := strings.TrimSpace(report.Record.ExaminerName); examiner != "" {
		text += examiner
	}
	r.setTextColor(0, 0, 0)
	r.setFont(13)
	r.drawText(305, 456, text)
}

func (r erxinReportPDFRenderer) drawOriginalCellText(x, y, width, height float64, text string, size float64, center bool) {
	value := strings.TrimSpace(text)
	if value == "" || value == "-" {
		return
	}
	r.setTextColor(0, 0, 0)
	r.setFont(size)
	textWidth, _ := r.pdf.MeasureTextWidth(value)
	textX := x + 10
	if center {
		textX = x + (width-textWidth)/2
	}
	r.drawText(textX, y+height*0.64, value)
}

func (r erxinReportPDFRenderer) centerText(x, y, width float64, text string) {
	value := strings.TrimSpace(text)
	if value == "" {
		return
	}
	textWidth, _ := r.pdf.MeasureTextWidth(value)
	r.drawText(x+(width-textWidth)/2, y, value)
}

func (r erxinReportPDFRenderer) drawText(x, y float64, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	_ = r.pdf.SetFont(erxinReportPDFFontFamily, "", r.currentFontSize)
	r.pdf.SetX(x)
	r.pdf.SetY(y)
	_ = r.pdf.Text(value)
}

func (r *erxinReportPDFRenderer) setTextColor(red, green, blue uint8) {
	r.pdf.SetTextColor(red, green, blue)
}

func (r *erxinReportPDFRenderer) setFont(size float64) {
	r.currentFontSize = size
	_ = r.pdf.SetFont(erxinReportPDFFontFamily, "", size)
}

func erxinReportPDFColumnLines(left float64, widths []float64) []float64 {
	lines := make([]float64, 0, len(widths)-1)
	x := left
	for _, width := range widths[:len(widths)-1] {
		x += width
		lines = append(lines, x)
	}
	return lines
}

func erxinReportPDFRowsOffset(rowHeights []float64, row int) float64 {
	offset := 0.0
	for index := 0; index < row; index++ {
		offset += rowHeights[index]
	}
	return offset
}

func erxinActualAgeSummaryText(record model.AssessmentRecordSummaryVO) string {
	parts := make([]string, 0, 3)
	if record.AgeYears > 0 {
		parts = append(parts, fmt.Sprintf("%d岁", record.AgeYears))
	}
	if record.AgeMonths > 0 {
		parts = append(parts, fmt.Sprintf("%d个月", record.AgeMonths))
	}
	if record.AgeDays > 0 {
		parts = append(parts, fmt.Sprintf("%d天", record.AgeDays))
	}
	return strings.Join(parts, "")
}

func erxinReportPDFStandardText(report model.ERXinReportVO) string {
	for _, value := range []string{report.SourceStandard, report.ScaleVersion, report.TemplateVersion} {
		if strings.Contains(value, "580") {
			return "WS/T 580—2017"
		}
	}
	return "WS/T 580—2017"
}

func erxinReportPDFDateText(value *time.Time) string {
	if value == nil || value.IsZero() {
		return "年        月        日"
	}
	return fmt.Sprintf("%d 年    %d 月    %d 日", value.Year(), int(value.Month()), value.Day())
}

func erxinBlankDash(value string) string {
	value = strings.TrimSpace(value)
	if value == "-" {
		return ""
	}
	return value
}

func erxinReportPDFDomainRowsByKey(rows []model.ERXinReportDomainRow) map[string]model.ERXinReportDomainRow {
	out := make(map[string]model.ERXinReportDomainRow, len(rows))
	for _, row := range rows {
		code := strings.ToUpper(strings.TrimSpace(row.DomainCode))
		if code != "" {
			out[code] = row
			continue
		}
		switch strings.ReplaceAll(strings.TrimSpace(row.DomainName), " ", "") {
		case "大运动":
			out["GM"] = row
		case "精细动作":
			out["FM"] = row
		case "适应能力":
			out["AD"] = row
		case "语言":
			out["LANG"] = row
		case "社会行为":
			out["SOC"] = row
		}
	}
	return out
}

func erxinReportPDFMentalAgeText(row model.ERXinReportDomainRow) string {
	if row.MentalAgeMonths <= 0 && strings.TrimSpace(row.MentalAgeMonthsText) == "" {
		return ""
	}
	return erxinTrimMonthUnit(nonEmptyString(row.MentalAgeMonthsText, erxinMonthNumberText(row.MentalAgeMonths)))
}

func erxinReportPDFSummaryMentalAgeText(summary model.ERXinReportSummary) string {
	if summary.MeanMentalAgeMonths <= 0 && strings.TrimSpace(summary.MeanMentalAgeMonthsText) == "" {
		return ""
	}
	return erxinTrimMonthUnit(nonEmptyString(summary.MeanMentalAgeMonthsText, erxinMonthNumberText(summary.MeanMentalAgeMonths)))
}

func erxinReportPDFDQText(value float64) string {
	if value <= 0 {
		return ""
	}
	return erxinFloatText(value)
}

func erxinMonthNumberText(value float64) string {
	text := erxinMonthText(value)
	return strings.TrimSuffix(text, "个月")
}

func erxinTrimMonthUnit(value string) string {
	return strings.TrimSuffix(strings.TrimSpace(value), "个月")
}
