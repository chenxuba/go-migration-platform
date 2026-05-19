package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/signintech/gopdf"
	"go-migration-platform/services/education/internal/model"
)

const (
	ShuangxiAReportSectionDevelopmentProfile = "developmentProfile"
	ShuangxiAReportSectionResultAnalysis     = "resultAnalysis"

	shuangxiAResultAnalysisPDFPageWidth    = 595.28
	shuangxiAResultAnalysisPDFPageHeight   = 841.89
	shuangxiAResultAnalysisPDFBottomMargin = 46.0
)

type shuangxiAResultAnalysisPDFExport struct {
	Title          string
	StudentName    string
	Gender         string
	BirthDate      string
	Age            string
	AssessmentDate string
	ExaminerName   string
	AssessmentName string
	Rows           []model.ShuangxiResultAnalysisRow
}

func (svc *Service) ExportShuangxiASelectedReportPDF(userID int64, recordID int64, sections []string, analysis *model.ShuangxiResultAnalysisVO) (string, string, []byte, error) {
	normalizedSections := normalizeShuangxiASelectedReportSections(sections)
	if len(normalizedSections) == 0 {
		return "", "", nil, errors.New("请选择导出内容")
	}

	record, err := svc.GetShuangxiAAssessmentRecord(userID, recordID)
	if err != nil {
		return "", "", nil, err
	}

	if len(normalizedSections) == 1 {
		content, err := svc.shuangxiASelectedReportSectionPDF(userID, recordID, record, normalizedSections[0], analysis)
		if err != nil {
			return "", "", nil, err
		}
		if len(content) == 0 {
			return "", "", nil, fmt.Errorf("%s暂无可导出内容", shuangxiASelectedReportSectionLabel(normalizedSections[0]))
		}
		return shuangxiASelectedReportPDFFileName(record, normalizedSections), iepPlanPDFContentType, content, nil
	}

	builder := newShuangxiASelectedReportPDFBuilder()
	for _, section := range normalizedSections {
		if err := svc.appendShuangxiASelectedReportSectionPDF(builder, userID, recordID, record, section, analysis); err != nil {
			return "", "", nil, err
		}
	}
	if builder.partCount == 0 {
		return "", "", nil, errors.New("暂无可导出内容")
	}
	content, err := builder.bytes()
	if err != nil {
		return "", "", nil, err
	}
	return shuangxiASelectedReportPDFFileName(record, normalizedSections), iepPlanPDFContentType, content, nil
}

func (svc *Service) shuangxiASelectedReportSectionPDF(userID int64, recordID int64, record model.AssessmentRecordDetailVO, section string, analysis *model.ShuangxiResultAnalysisVO) ([]byte, error) {
	switch section {
	case ShuangxiAReportSectionDevelopmentProfile:
		data, records, err := svc.shuangxiADevelopmentProfilePDFSource(record)
		if err != nil {
			return nil, err
		}
		return buildShuangxiADevelopmentProfilePDF(data, records)
	case ShuangxiAReportSectionResultAnalysis:
		export, err := svc.shuangxiAResultAnalysisPDFExport(userID, recordID, analysis)
		if err != nil {
			return nil, err
		}
		return buildShuangxiAResultAnalysisPDF(export)
	default:
		return nil, fmt.Errorf("不支持的导出内容：%s", section)
	}
}

func (svc *Service) appendShuangxiASelectedReportSectionPDF(builder *shuangxiASelectedReportPDFBuilder, userID int64, recordID int64, record model.AssessmentRecordDetailVO, section string, analysis *model.ShuangxiResultAnalysisVO) error {
	switch section {
	case ShuangxiAReportSectionDevelopmentProfile:
		data, records, err := svc.shuangxiADevelopmentProfilePDFSource(record)
		if err != nil {
			return err
		}
		return builder.appendDirectDraw(func(pdf *gopdf.GoPdf) error {
			return drawShuangxiADevelopmentProfilePDFPages(pdf, data, records)
		})
	case ShuangxiAReportSectionResultAnalysis:
		export, err := svc.shuangxiAResultAnalysisPDFExport(userID, recordID, analysis)
		if err != nil {
			return err
		}
		return builder.appendDirectDraw(func(pdf *gopdf.GoPdf) error {
			return drawShuangxiAResultAnalysisPDFPages(pdf, export)
		})
	default:
		return fmt.Errorf("不支持的导出内容：%s", section)
	}
}

func (svc *Service) shuangxiADevelopmentProfilePDFSource(record model.AssessmentRecordDetailVO) (shuangxiAStaticData, []model.AssessmentRecordDetailVO, error) {
	data, err := svc.loadShuangxiAStaticData(context.Background())
	if err != nil {
		return shuangxiAStaticData{}, nil, err
	}
	records := []model.AssessmentRecordDetailVO{record}
	if svc.repo != nil && record.InstID > 0 && record.StudentID > 0 {
		history, err := svc.repo.ListAssessmentRecordsForStudentScale(
			context.Background(),
			record.InstID,
			record.StudentID,
			shuangxiAScaleCode,
			record.ID,
			100,
		)
		if err != nil {
			return shuangxiAStaticData{}, nil, err
		}
		records = shuangxiAProfileRecordsThroughCurrent(history, record)
	}
	return data, records, nil
}

func (svc *Service) shuangxiAResultAnalysisPDFExport(userID int64, recordID int64, analysis *model.ShuangxiResultAnalysisVO) (shuangxiAResultAnalysisPDFExport, error) {
	_, record, data, itemScores, err := svc.shuangxiAResultAnalysisContext(userID, recordID)
	if err != nil {
		return shuangxiAResultAnalysisPDFExport{}, err
	}
	if analysis == nil || len(analysis.Rows) == 0 {
		saved, err := svc.GetShuangxiAResultAnalysis(userID, recordID)
		if err != nil {
			return shuangxiAResultAnalysisPDFExport{}, err
		}
		analysis = &saved
	}
	if shuangxiAResultAnalysisContentIsEmpty(*analysis) {
		return shuangxiAResultAnalysisPDFExport{}, errors.New("请先生成评量结果分析后再导出")
	}
	fallback := buildEmptyShuangxiAResultAnalysis(record, data, itemScores)
	normalized := normalizeShuangxiAResultAnalysis(*analysis, fallback, analysis.GeneratedBy)
	return shuangxiAResultAnalysisPDFExport{
		Title:          firstNonEmptyExportValue(strings.TrimSpace(normalized.Title), "双溪心智障碍个别化教育课程（三）评量结果分析表"),
		StudentName:    strings.TrimSpace(record.StudentName),
		Gender:         strings.TrimSpace(record.StudentGender),
		BirthDate:      formatIEPPlanDate(record.BirthDate),
		Age:            formatIEPPlanAge(record.AgeYears, record.AgeMonths, record.AgeDays),
		AssessmentDate: formatIEPPlanDate(record.AssessmentDate),
		ExaminerName:   strings.TrimSpace(record.ExaminerName),
		AssessmentName: firstNonEmptyExportValue(strings.TrimSpace(record.AssessmentName), "双溪课程评量表A"),
		Rows:           normalized.Rows,
	}, nil
}

func shuangxiAResultAnalysisContentIsEmpty(analysis model.ShuangxiResultAnalysisVO) bool {
	for _, row := range analysis.Rows {
		if strings.TrimSpace(row.Strengths) != "" ||
			strings.TrimSpace(row.Weaknesses) != "" ||
			strings.TrimSpace(row.Reason) != "" ||
			strings.TrimSpace(row.Strategy) != "" {
			return false
		}
	}
	return true
}

func normalizeShuangxiASelectedReportSections(sections []string) []string {
	allowed := map[string]bool{
		ShuangxiAReportSectionDevelopmentProfile: true,
		ShuangxiAReportSectionResultAnalysis:     true,
	}
	out := make([]string, 0, len(sections))
	seen := make(map[string]bool, len(sections))
	for _, raw := range sections {
		section := strings.TrimSpace(raw)
		if !allowed[section] || seen[section] {
			continue
		}
		seen[section] = true
		out = append(out, section)
	}
	return out
}

func shuangxiASelectedReportSectionLabel(section string) string {
	switch section {
	case ShuangxiAReportSectionDevelopmentProfile:
		return "综合发展侧面图"
	case ShuangxiAReportSectionResultAnalysis:
		return "评量结果分析"
	default:
		return "所选内容"
	}
}

func shuangxiASelectedReportPDFFileName(record model.AssessmentRecordDetailVO, sections []string) string {
	name := nonEmptyString(record.StudentName, "未命名儿童")
	title := "双溪课程评量表A报告"
	if len(sections) == 1 {
		title = "双溪" + shuangxiASelectedReportSectionLabel(sections[0])
	}
	return sanitizeTemplateFileName(fmt.Sprintf("%s-%s-%s.pdf", name, title, time.Now().Format("20060102150405")))
}

func buildShuangxiAResultAnalysisPDF(export shuangxiAResultAnalysisPDFExport) ([]byte, error) {
	if !shuangxiAResultAnalysisPDFHasContent(export.Rows) {
		return nil, errors.New("请先生成评量结果分析后再导出")
	}
	var pdf gopdf.GoPdf
	pdf.Start(gopdf.Config{
		Unit:     gopdf.UnitPT,
		PageSize: gopdf.Rect{W: shuangxiAResultAnalysisPDFPageWidth, H: shuangxiAResultAnalysisPDFPageHeight},
	})
	if err := addShuangxiAPDFFont(&pdf); err != nil {
		return nil, err
	}
	if err := drawShuangxiAResultAnalysisPDFPages(&pdf, export); err != nil {
		return nil, err
	}
	return pdf.GetBytesPdfReturnErr()
}

func drawShuangxiAResultAnalysisPDFPages(pdf *gopdf.GoPdf, export shuangxiAResultAnalysisPDFExport) error {
	if !shuangxiAResultAnalysisPDFHasContent(export.Rows) {
		return errors.New("请先生成评量结果分析后再导出")
	}
	renderer := shuangxiAResultAnalysisPDFRenderer{pdf: pdf}
	renderer.draw(export)
	return nil
}

func shuangxiAResultAnalysisPDFHasContent(rows []model.ShuangxiResultAnalysisRow) bool {
	for _, row := range rows {
		if strings.TrimSpace(row.Strengths) != "" ||
			strings.TrimSpace(row.Weaknesses) != "" ||
			strings.TrimSpace(row.Reason) != "" ||
			strings.TrimSpace(row.Strategy) != "" {
			return true
		}
	}
	return false
}

type shuangxiAResultAnalysisPDFRenderer struct {
	pdf             *gopdf.GoPdf
	currentFontSize float64
}

func (r *shuangxiAResultAnalysisPDFRenderer) draw(export shuangxiAResultAnalysisPDFExport) {
	firstPage := true
	widths := []float64{64, 148, 132, 167}
	left := (shuangxiAResultAnalysisPDFPageWidth - sumFloat64(widths)) / 2

	r.beginPage(export, firstPage, left, widths)
	firstPage = false
	for _, source := range export.Rows {
		row := trimShuangxiAResultAnalysisRow(source)
		cells := [][]string{
			r.cellLines([]string{row.Domain}, widths[0]-10, 8.5),
			r.cellLines(shuangxiAResultAnalysisStatusLines(row), widths[1]-12, 8.2),
			r.cellLines(splitWordLines(row.Reason), widths[2]-12, 8.2),
			r.cellLines(splitWordLines(row.Strategy), widths[3]-12, 8.2),
		}
		r.drawRow(export, left, widths, cells, &firstPage)
	}
}

func (r *shuangxiAResultAnalysisPDFRenderer) beginPage(export shuangxiAResultAnalysisPDFExport, firstPage bool, left float64, widths []float64) {
	r.pdf.AddPageWithOption(gopdf.PageOption{PageSize: &gopdf.Rect{W: shuangxiAResultAnalysisPDFPageWidth, H: shuangxiAResultAnalysisPDFPageHeight}})
	title := firstNonEmptyExportValue(export.Title, "双溪心智障碍个别化教育课程（三）评量结果分析表")
	if !firstPage {
		title += "（续）"
	}
	r.drawTitle(title, 46, 15)
	y := 74.0
	if firstPage {
		r.setFont(9.5)
		r.drawText(left, y, "儿童姓名："+firstNonEmptyExportValue(export.StudentName, "-"))
		r.drawText(left+170, y, "评估者："+firstNonEmptyExportValue(export.ExaminerName, "-"))
		r.drawText(left+330, y, "评量日期："+firstNonEmptyExportValue(export.AssessmentDate, "-"))
		y += 24
	} else {
		y += 12
	}
	r.drawCells(left, y, widths, 32, []string{"领   域", "现况分析", "原因推断", "建议策略"}, 9.5, true)
	r.pdf.SetY(y + 32)
}

func (r *shuangxiAResultAnalysisPDFRenderer) drawRow(export shuangxiAResultAnalysisPDFExport, left float64, widths []float64, cells [][]string, firstPage *bool) {
	const (
		fontSize   = 8.2
		lineHeight = 11.2
		padding    = 6.0
		minHeight  = 58.0
	)
	bottom := shuangxiAResultAnalysisPDFPageHeight - shuangxiAResultAnalysisPDFBottomMargin
	for hasPDFCellLines(cells) {
		y := r.pdf.GetY()
		available := bottom - y
		maxLines := int(math.Floor((available - padding*2) / lineHeight))
		if maxLines < 3 {
			r.beginPage(export, *firstPage, left, widths)
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
			take := minShuangxiAPDFInt(maxLines, len(cells[index]))
			if take > 0 {
				chunks[index] = cells[index][:take]
				cells[index] = cells[index][take:]
			}
		}
		rowHeight := math.Max(minHeight, float64(maxPDFCellLineCount(chunks))*lineHeight+padding*2)
		if y+rowHeight > bottom && rowHeight < bottom-104 {
			r.beginPage(export, *firstPage, left, widths)
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

func (r *shuangxiAResultAnalysisPDFRenderer) drawTitle(title string, baselineY float64, size float64) {
	r.setTextColor(0, 0, 0)
	r.setFont(size)
	r.centerText(0, baselineY, shuangxiAResultAnalysisPDFPageWidth, title)
}

func (r *shuangxiAResultAnalysisPDFRenderer) drawCells(left, y float64, widths []float64, height float64, values []string, fontSize float64, center bool) {
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

func (r *shuangxiAResultAnalysisPDFRenderer) drawCell(x, y, width, height float64, lines []string, fontSize float64, align, valign string) {
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

func (r *shuangxiAResultAnalysisPDFRenderer) cellLines(values []string, width float64, fontSize float64) []string {
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

func (r *shuangxiAResultAnalysisPDFRenderer) centerText(x, y, width float64, text string) {
	value := strings.TrimSpace(text)
	if value == "" {
		return
	}
	textWidth, _ := r.pdf.MeasureTextWidth(value)
	r.drawText(x+(width-textWidth)/2, y, value)
}

func (r *shuangxiAResultAnalysisPDFRenderer) drawText(x, y float64, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	_ = r.pdf.SetFont(shuangxiAProfilePDFFontFamily, "", r.currentFontSize)
	r.pdf.SetX(x)
	r.pdf.SetY(y)
	_ = r.pdf.Text(value)
}

func (r *shuangxiAResultAnalysisPDFRenderer) setTextColor(red, green, blue uint8) {
	r.pdf.SetTextColor(red, green, blue)
}

func (r *shuangxiAResultAnalysisPDFRenderer) setFont(size float64) {
	r.currentFontSize = size
	_ = r.pdf.SetFont(shuangxiAProfilePDFFontFamily, "", size)
}

func shuangxiAResultAnalysisStatusLines(row model.ShuangxiResultAnalysisRow) []string {
	strengths := strings.TrimSpace(row.Strengths)
	weaknesses := strings.TrimSpace(row.Weaknesses)
	lines := make([]string, 0, 2)
	if strengths != "" {
		lines = append(lines, "优："+strengths)
	}
	if weaknesses != "" {
		lines = append(lines, "弱："+weaknesses)
	}
	if len(lines) == 0 {
		return []string{"无"}
	}
	return lines
}

func minShuangxiAPDFInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type shuangxiASelectedReportPDFBuilder struct {
	pdf       gopdf.GoPdf
	fontReady bool
	partCount int
}

func newShuangxiASelectedReportPDFBuilder() *shuangxiASelectedReportPDFBuilder {
	builder := &shuangxiASelectedReportPDFBuilder{}
	builder.pdf.Start(gopdf.Config{
		Unit:     gopdf.UnitPT,
		PageSize: gopdf.Rect{W: shuangxiAProfilePDFPageWidth, H: shuangxiAProfilePDFPageHeight},
	})
	return builder
}

func (b *shuangxiASelectedReportPDFBuilder) appendDirectDraw(draw func(*gopdf.GoPdf) error) error {
	if !b.fontReady {
		if err := addShuangxiAPDFFont(&b.pdf); err != nil {
			return err
		}
		b.fontReady = true
	}
	if err := draw(&b.pdf); err != nil {
		return err
	}
	b.partCount++
	return nil
}

func (b *shuangxiASelectedReportPDFBuilder) bytes() ([]byte, error) {
	var output bytes.Buffer
	if err := b.pdf.Write(&output); err != nil {
		return nil, fmt.Errorf("生成PDF失败：%w", err)
	}
	return output.Bytes(), nil
}
