package service

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/signintech/gopdf"
)

const (
	autismDevReportPDFMargin       = 42.0
	autismDevReportPDFBottomMargin = 46.0
	autismDevReportPDFLineHeight   = 12.0
)

type autismDevReportPDFRenderer struct {
	pdf             *gopdf.GoPdf
	currentFontSize float64
}

func (svc *Service) buildAutismDevAssessmentSituationPDF(userID int64, recordID int64) (string, []byte, error) {
	_, record, score, data, itemScores, err := svc.autismDevResultAnalysisContext(userID, recordID)
	if err != nil {
		return "", nil, err
	}
	export := buildAutismDevAssessmentSituationWordExport(record, score, data, itemScores)
	content, err := buildAutismDevAssessmentSituationPDF(export)
	if err != nil {
		return "", nil, err
	}
	fileName := fmt.Sprintf("%s-孤独症儿童评估情况-%s.pdf", sanitizeExportFileName(export.StudentName), time.Now().Format("20060102150405"))
	return fileName, content, nil
}

func buildAutismDevAssessmentSituationPDF(export autismDevAssessmentSituationWordExport) ([]byte, error) {
	var pdf gopdf.GoPdf
	pdf.Start(gopdf.Config{
		Unit:     gopdf.UnitPT,
		PageSize: gopdf.Rect{W: autismDevProfilePDFPageWidth, H: autismDevProfilePDFPageHeight},
	})
	if err := addAutismDevProfilePDFFont(&pdf); err != nil {
		return nil, err
	}
	if err := drawAutismDevAssessmentSituationPDFPages(&pdf, export); err != nil {
		return nil, err
	}
	return pdf.GetBytesPdfReturnErr()
}

func drawAutismDevAssessmentSituationPDFPages(pdf *gopdf.GoPdf, export autismDevAssessmentSituationWordExport) error {
	renderer := autismDevReportPDFRenderer{pdf: pdf}
	renderer.drawAssessmentSituation(export)
	return nil
}

func buildAutismDevResultAnalysisPDF(export autismDevResultAnalysisWordExport) ([]byte, error) {
	if len(export.Rows) == 0 {
		return nil, fmt.Errorf("暂无可导出的评估结果分析")
	}
	var pdf gopdf.GoPdf
	pdf.Start(gopdf.Config{
		Unit:     gopdf.UnitPT,
		PageSize: gopdf.Rect{W: autismDevProfilePDFPageWidth, H: autismDevProfilePDFPageHeight},
	})
	if err := addAutismDevProfilePDFFont(&pdf); err != nil {
		return nil, err
	}
	if err := drawAutismDevResultAnalysisPDFPages(&pdf, export); err != nil {
		return nil, err
	}
	return pdf.GetBytesPdfReturnErr()
}

func drawAutismDevResultAnalysisPDFPages(pdf *gopdf.GoPdf, export autismDevResultAnalysisWordExport) error {
	if len(export.Rows) == 0 {
		return fmt.Errorf("暂无可导出的评估结果分析")
	}
	renderer := autismDevReportPDFRenderer{pdf: pdf}
	renderer.drawResultAnalysis(export)
	return nil
}

func (r *autismDevReportPDFRenderer) drawAssessmentSituation(export autismDevAssessmentSituationWordExport) {
	r.pdf.AddPage()
	r.drawTitle(firstNonEmptyExportValue(export.Title, autismDevAssessmentSituationTitle), 52, 15.5)

	left := 58.0
	top := 86.0
	tableWidth := autismDevProfilePDFPageWidth - left*2
	infoWidths := []float64{50, 112, 58, 112, 58, tableWidth - 390}
	mainWidths := autismDevAssessmentSituationPDFMainWidths(tableWidth)

	y := top
	rowHeight := 22.0
	r.drawCells(left, y, infoWidths, rowHeight, []string{
		"儿童姓名", autismDevAssessmentSituationValue(export.StudentName),
		"测评年龄", autismDevAssessmentSituationValue(export.Age),
		"测评日期", autismDevAssessmentSituationValue(export.AssessmentDate),
	}, 9.5, true)
	y += rowHeight
	r.drawCells(left, y, infoWidths, rowHeight, []string{
		"评估者", autismDevAssessmentSituationValue(export.ExaminerName),
		"出生日期", autismDevAssessmentSituationValue(export.BirthDate),
		"测评次数", autismDevAssessmentSituationValue(export.AssessmentSequence),
	}, 9.5, true)
	y += rowHeight

	headerTop := y
	headerHeight := 46.0
	r.drawCell(left, headerTop, mainWidths[0], headerHeight, []string{"领   域"}, 9.5, "center", "middle")
	r.drawCell(left+mainWidths[0], headerTop, mainWidths[1]+mainWidths[2]+mainWidths[3], 24, []string{"评估结果"}, 9.5, "center", "middle")
	r.drawCell(left+mainWidths[0]+mainWidths[1]+mainWidths[2]+mainWidths[3], headerTop, mainWidths[4], headerHeight, []string{"备注"}, 9.5, "center", "middle")
	subY := headerTop + 24
	r.drawCell(left+mainWidths[0], subY, mainWidths[1], headerHeight-24, []string{"P"}, 9, "center", "middle")
	r.drawCell(left+mainWidths[0]+mainWidths[1], subY, mainWidths[2], headerHeight-24, []string{"E+F(X)"}, 9, "center", "middle")
	r.drawCell(left+mainWidths[0]+mainWidths[1]+mainWidths[2], subY, mainWidths[3], headerHeight-24, []string{"总分"}, 9, "center", "middle")
	y += headerHeight

	for _, row := range export.DevelopmentRows {
		r.drawCells(left, y, mainWidths, 24, []string{
			row.Label,
			autismDevAssessmentSituationScoreText(row.Measured, row.PCount),
			autismDevAssessmentSituationScoreText(row.Measured, row.SupportCount),
			autismDevAssessmentSituationScoreText(row.Measured, row.TotalScore),
			"",
		}, 9.2, true)
		y += 24
	}
	if len(export.BehaviorRows) > 0 {
		r.drawCells(left, y, mainWidths, 23, []string{"情绪与行为能力", "A", "M", "S", ""}, 9.2, true)
		y += 23
		for index, row := range export.BehaviorRows {
			r.drawCells(left, y, mainWidths, 23, []string{
				fmt.Sprintf("%d、%s", index+1, row.Label),
				autismDevAssessmentSituationScoreText(row.Measured, row.ACount),
				autismDevAssessmentSituationScoreText(row.Measured, row.MCount),
				autismDevAssessmentSituationScoreText(row.Measured, row.SCount),
				"",
			}, 9.0, true)
			y += 23
		}
	}
}

func (r *autismDevReportPDFRenderer) drawResultAnalysis(export autismDevResultAnalysisWordExport) {
	firstPage := true
	widths := []float64{58, 132, 176, 145}
	left := (autismDevProfilePDFPageWidth - sumFloat64(widths)) / 2

	r.beginResultAnalysisPage(export, firstPage, left, widths)
	firstPage = false
	for _, source := range export.Rows {
		row := trimAutismDevResultAnalysisRow(source)
		cells := [][]string{
			r.cellLines([]string{row.Domain}, widths[0]-10, 8.5),
			r.cellLines(splitWordLines(row.Status), widths[1]-12, 8.2),
			r.cellLines(autismDevStrengthWeaknessLines(row), widths[2]-12, 8.2),
			r.cellLines(splitWordLines(row.Targets), widths[3]-12, 8.2),
		}
		r.drawResultAnalysisRow(export, left, widths, cells, &firstPage)
	}
}

func (r *autismDevReportPDFRenderer) beginResultAnalysisPage(export autismDevResultAnalysisWordExport, firstPage bool, left float64, widths []float64) {
	r.pdf.AddPage()
	title := firstNonEmptyExportValue(export.Title, "孤独症儿童评估结果分析表")
	if !firstPage {
		title += "（续）"
	}
	r.drawTitle(title, 46, 15)
	y := 74.0
	if firstPage {
		r.setFont(9.5)
		r.drawText(left, y, "儿童姓名："+autismDevAssessmentSituationValue(export.StudentName))
		r.drawText(left+180, y, "评估者："+autismDevAssessmentSituationValue(export.ExaminerName))
		r.drawText(left+350, y, "评估时间："+autismDevAssessmentSituationValue(export.AssessmentDate))
		y += 26
	} else {
		y += 12
	}
	r.drawResultAnalysisHeader(left, y, widths)
	r.pdf.SetY(y + 32)
}

func (r *autismDevReportPDFRenderer) drawResultAnalysisHeader(left, y float64, widths []float64) {
	r.drawCells(left, y, widths, 32, []string{"领   域", "能力现状描述", "优劣分析", "训练目标"}, 9.5, true)
}

func (r *autismDevReportPDFRenderer) drawResultAnalysisRow(export autismDevResultAnalysisWordExport, left float64, widths []float64, cells [][]string, firstPage *bool) {
	const (
		fontSize   = 8.2
		lineHeight = 11.2
		padding    = 6.0
		minHeight  = 62.0
	)
	bottom := autismDevProfilePDFPageHeight - autismDevReportPDFBottomMargin
	for hasPDFCellLines(cells) {
		y := r.pdf.GetY()
		available := bottom - y
		maxLines := int(math.Floor((available - padding*2) / lineHeight))
		if maxLines < 3 {
			r.beginResultAnalysisPage(export, *firstPage, left, widths)
			*firstPage = false
			y = r.pdf.GetY()
			available = bottom - y
			maxLines = int(math.Floor((available - padding*2) / lineHeight))
		}
		if maxLines < 3 {
			maxLines = 3
		}
		chunks := make([][]string, len(cells))
		for index := range cells {
			take := minAutismDevPDFInt(maxLines, len(cells[index]))
			if take > 0 {
				chunks[index] = cells[index][:take]
				cells[index] = cells[index][take:]
			}
		}
		rowHeight := math.Max(minHeight, float64(maxPDFCellLineCount(chunks))*lineHeight+padding*2)
		if y+rowHeight > bottom && rowHeight < bottom-104 {
			r.beginResultAnalysisPage(export, *firstPage, left, widths)
			*firstPage = false
			y = r.pdf.GetY()
		}
		x := left
		for index, lines := range chunks {
			align := "left"
			if index == 0 {
				align = "center"
			}
			r.drawCell(x, y, widths[index], rowHeight, lines, fontSize, align, "top")
			x += widths[index]
		}
		r.pdf.SetY(y + rowHeight)
	}
}

func (r *autismDevReportPDFRenderer) drawTitle(title string, baselineY float64, size float64) {
	r.setTextColor(0, 0, 0)
	r.setFont(size)
	r.centerText(0, baselineY, autismDevProfilePDFPageWidth, title)
}

func (r *autismDevReportPDFRenderer) drawCells(left, y float64, widths []float64, height float64, values []string, fontSize float64, center bool) {
	x := left
	for index, width := range widths {
		value := ""
		if index < len(values) {
			value = values[index]
		}
		align := "left"
		if center {
			align = "center"
		}
		r.drawCell(x, y, width, height, []string{value}, fontSize, align, "middle")
		x += width
	}
}

func (r *autismDevReportPDFRenderer) drawCell(x, y, width, height float64, lines []string, fontSize float64, align, valign string) {
	r.pdf.SetStrokeColor(0, 0, 0)
	r.pdf.SetLineWidth(0.55)
	r.pdf.RectFromUpperLeft(x, y, width, height)

	drawLines := r.cellLines(lines, width-10, fontSize)
	if len(drawLines) == 0 {
		return
	}
	lineHeight := fontSize + 3.0
	totalTextHeight := float64(len(drawLines)) * lineHeight
	textY := y + 7 + fontSize
	if valign == "middle" {
		textY = y + (height-totalTextHeight)/2 + fontSize
	}
	r.setTextColor(0, 0, 0)
	r.setFont(fontSize)
	for _, line := range drawLines {
		textX := x + 5
		if align == "center" {
			textWidth, _ := r.pdf.MeasureTextWidth(line)
			textX = x + (width-textWidth)/2
		}
		if textY > y+height-3 {
			break
		}
		r.drawText(textX, textY, line)
		textY += lineHeight
	}
}

func (r *autismDevReportPDFRenderer) cellLines(values []string, width float64, fontSize float64) []string {
	r.setFont(fontSize)
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		rawLines := splitWordLines(value)
		if len(rawLines) == 0 {
			rawLines = []string{value}
		}
		for _, raw := range rawLines {
			lines, err := r.pdf.SplitText(raw, width)
			if err != nil || len(lines) == 0 {
				out = append(out, raw)
				continue
			}
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" {
					out = append(out, line)
				}
			}
		}
	}
	return out
}

func (r *autismDevReportPDFRenderer) centerText(x, y, width float64, text string) {
	value := strings.TrimSpace(text)
	if value == "" {
		return
	}
	textWidth, _ := r.pdf.MeasureTextWidth(value)
	r.drawText(x+(width-textWidth)/2, y, value)
}

func (r *autismDevReportPDFRenderer) drawText(x, y float64, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	_ = r.pdf.SetFont(autismDevProfilePDFFontFamily, "", r.currentFontSize)
	r.pdf.SetX(x)
	r.pdf.SetY(y)
	_ = r.pdf.Text(value)
}

func (r *autismDevReportPDFRenderer) setTextColor(red, green, blue uint8) {
	r.pdf.SetTextColor(red, green, blue)
}

func (r *autismDevReportPDFRenderer) setFont(size float64) {
	r.currentFontSize = size
	_ = r.pdf.SetFont(autismDevProfilePDFFontFamily, "", size)
}

func autismDevAssessmentSituationPDFMainWidths(total float64) []float64 {
	spans := []float64{18, 13, 13, 13, 13}
	widths := make([]float64, len(spans))
	for index, span := range spans {
		widths[index] = total * span / 70
	}
	return widths
}

func sumFloat64(values []float64) float64 {
	total := 0.0
	for _, value := range values {
		total += value
	}
	return total
}

func hasPDFCellLines(cells [][]string) bool {
	for _, lines := range cells {
		if len(lines) > 0 {
			return true
		}
	}
	return false
}

func maxPDFCellLineCount(cells [][]string) int {
	maxCount := 0
	for _, lines := range cells {
		if len(lines) > maxCount {
			maxCount = len(lines)
		}
	}
	if maxCount == 0 {
		return 1
	}
	return maxCount
}

func minAutismDevPDFInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
