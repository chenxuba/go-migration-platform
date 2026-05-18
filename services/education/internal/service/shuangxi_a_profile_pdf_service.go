package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/signintech/gopdf"
	"go-migration-platform/services/education/internal/model"
)

const (
	shuangxiAProfilePDFPageWidth  = 841.89
	shuangxiAProfilePDFPageHeight = 595.28
	shuangxiAProfilePDFFontFamily = "shuangxi-a-cjk"
)

type shuangxiAProfileColor struct {
	Name string
	R    uint8
	G    uint8
	B    uint8
}

type shuangxiAProfilePoint struct {
	X     float64
	Y     float64
	Score int
}

type shuangxiAProfileRect struct {
	X float64
	Y float64
	W float64
	H float64
}

type shuangxiAProfileDomain struct {
	Code        string
	Name        string
	DomainNo    int
	MaxRawScore int
}

var shuangxiAProfileColors = []shuangxiAProfileColor{
	{Name: "红色", R: 220, G: 38, B: 38},
	{Name: "蓝色", R: 37, G: 99, B: 235},
	{Name: "绿色", R: 22, G: 163, B: 74},
	{Name: "黑色", R: 17, G: 24, B: 39},
}

func (svc *Service) GenerateShuangxiADevelopmentProfilePDF(userID, recordID int64) (string, []byte, error) {
	record, err := svc.GetShuangxiAAssessmentRecord(userID, recordID)
	if err != nil {
		return "", nil, err
	}
	data, err := svc.loadShuangxiAStaticData(context.Background())
	if err != nil {
		return "", nil, err
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
			return "", nil, err
		}
		records = shuangxiAProfileRecordsThroughCurrent(history, record)
	}
	content, err := buildShuangxiADevelopmentProfilePDF(data, records)
	if err != nil {
		return "", nil, err
	}
	name := nonEmptyString(record.StudentName, "未命名儿童")
	filename := sanitizeTemplateFileName(fmt.Sprintf("%s-双溪综合发展侧面图-%s.pdf", name, time.Now().Format("20060102150405")))
	return filename, content, nil
}

func shuangxiAProfileRecordsThroughCurrent(history []model.AssessmentRecordDetailVO, current model.AssessmentRecordDetailVO) []model.AssessmentRecordDetailVO {
	records := make([]model.AssessmentRecordDetailVO, 0, len(history)+1)
	seen := make(map[int64]bool, len(history)+1)
	for _, item := range history {
		if item.ID <= 0 || seen[item.ID] {
			continue
		}
		records = append(records, item)
		seen[item.ID] = true
	}
	if current.ID > 0 && !seen[current.ID] {
		records = append(records, current)
	}
	if len(records) == 0 {
		return []model.AssessmentRecordDetailVO{current}
	}
	sort.SliceStable(records, func(i, j int) bool {
		left := shuangxiAProfileRecordSortTime(records[i])
		right := shuangxiAProfileRecordSortTime(records[j])
		if !left.Equal(right) {
			return left.Before(right)
		}
		return records[i].ID < records[j].ID
	})
	currentIndex := -1
	for index, item := range records {
		if item.ID == current.ID {
			currentIndex = index
			break
		}
	}
	if currentIndex < 0 {
		currentIndex = len(records) - 1
	}
	start := currentIndex - 3
	if start < 0 {
		start = 0
	}
	end := currentIndex + 1
	if end > len(records) {
		end = len(records)
	}
	return records[start:end]
}

func shuangxiAProfileRecordSortTime(record model.AssessmentRecordDetailVO) time.Time {
	if record.AssessmentDate != nil {
		return *record.AssessmentDate
	}
	if record.CreatedTime != nil {
		return *record.CreatedTime
	}
	if record.UpdatedTime != nil {
		return *record.UpdatedTime
	}
	return time.Unix(0, 0)
}

func buildShuangxiADevelopmentProfilePDF(data shuangxiAStaticData, records []model.AssessmentRecordDetailVO) ([]byte, error) {
	fontBytes, err := loadPEP3PDFFontBytes()
	if err != nil {
		return nil, err
	}
	var pdf gopdf.GoPdf
	pdf.Start(gopdf.Config{
		Unit:     gopdf.UnitPT,
		PageSize: gopdf.Rect{W: shuangxiAProfilePDFPageWidth, H: shuangxiAProfilePDFPageHeight},
	})
	pdf.AddPage()
	if err := pdf.AddTTFFontByReader(shuangxiAProfilePDFFontFamily, bytes.NewReader(fontBytes)); err != nil {
		return nil, fmt.Errorf("load Shuangxi profile PDF font: %w", err)
	}
	renderer := shuangxiAProfilePDFRenderer{pdf: &pdf}
	if err := renderer.draw(data, records); err != nil {
		return nil, err
	}
	return pdf.GetBytesPdfReturnErr()
}

type shuangxiAProfilePDFRenderer struct {
	pdf      *gopdf.GoPdf
	fontSize float64
}

func (r *shuangxiAProfilePDFRenderer) draw(data shuangxiAStaticData, records []model.AssessmentRecordDetailVO) error {
	domains := shuangxiAProfileDomains(data)
	if len(domains) == 0 {
		return fmt.Errorf("Shuangxi A domains are not configured")
	}
	r.pdf.SetFillColor(255, 255, 255)
	r.pdf.RectFromUpperLeftWithStyle(0, 0, shuangxiAProfilePDFPageWidth, shuangxiAProfilePDFPageHeight, "F")
	if err := r.drawHeader(records); err != nil {
		return err
	}
	if err := r.drawScaleLabels(); err != nil {
		return err
	}
	if err := r.drawGrid(domains); err != nil {
		return err
	}
	r.drawProfiles(data, domains, records)
	return nil
}

func shuangxiAProfileDomains(data shuangxiAStaticData) []shuangxiAProfileDomain {
	domains := make([]shuangxiAProfileDomain, 0, len(data.domains))
	for index, domain := range data.domains {
		code := strings.TrimSpace(domain.ScaleCode)
		if code == "" {
			continue
		}
		domainNo := domain.DomainNo
		if domainNo <= 0 {
			domainNo = index + 1
		}
		maxRawScore := domain.MaxRawScore
		if maxRawScore <= 0 {
			maxRawScore = domain.ItemCount * 3
		}
		domains = append(domains, shuangxiAProfileDomain{
			Code:        code,
			Name:        strings.TrimSpace(nonEmptyString(domain.ScaleName, code)),
			DomainNo:    domainNo,
			MaxRawScore: maxRawScore,
		})
	}
	sort.SliceStable(domains, func(i, j int) bool {
		return domains[i].DomainNo < domains[j].DomainNo
	})
	return domains
}

func (r *shuangxiAProfilePDFRenderer) drawHeader(records []model.AssessmentRecordDetailVO) error {
	r.setTextColor(0, 0, 0)
	r.setFont(27)
	if err := r.cell(0, 34, shuangxiAProfilePDFPageWidth, 34, "综合发展侧面图（一）", gopdf.Center|gopdf.Middle); err != nil {
		return err
	}
	current := model.AssessmentRecordDetailVO{}
	if len(records) > 0 {
		current = records[len(records)-1]
	}
	r.setFont(12.2)
	if err := r.drawInfoField(38, 88, 128, "学生姓名：", current.StudentName); err != nil {
		return err
	}
	if err := r.drawInfoField(240, 88, 70, "性别：", shuangxiAProfileGenderText(current.StudentGender)); err != nil {
		return err
	}
	if err := r.drawInfoField(376, 88, 128, "出生日期：", shuangxiAProfileDateText(current.BirthDate)); err != nil {
		return err
	}
	if err := r.drawInfoField(602, 88, 142, "生理年龄：", shuangxiAProfilePhysiologicalAgeText(current.BirthDate, current.AssessmentDate)); err != nil {
		return err
	}

	for slot := 0; slot < 4; slot++ {
		row := slot / 2
		col := slot % 2
		x := 38.0
		if col == 1 {
			x = 434
		}
		y := 115.0 + float64(row)*24
		var record *model.AssessmentRecordDetailVO
		if slot < len(records) {
			record = &records[slot]
		}
		if err := r.drawAssessmentLine(x, y, slot, record); err != nil {
			return err
		}
	}
	return nil
}

func (r *shuangxiAProfilePDFRenderer) drawInfoField(x, y, valueWidth float64, label string, value string) error {
	const (
		labelValueGap = 2.0
		fieldHeight   = 18.0
		underlinePad  = 1.8
	)
	r.setTextColor(0, 0, 0)
	labelWidth, err := r.pdf.MeasureTextWidth(label)
	if err != nil {
		return err
	}
	if err := r.cell(x, y, labelWidth, fieldHeight, label, gopdf.Left|gopdf.Middle); err != nil {
		return err
	}
	lineX := x + labelWidth + labelValueGap
	lineY := y + fieldHeight - underlinePad
	r.pdf.SetStrokeColor(0, 0, 0)
	r.pdf.SetLineWidth(0.45)
	r.pdf.Line(lineX, lineY, lineX+valueWidth, lineY)
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return r.cell(lineX, y, valueWidth, fieldHeight, value, gopdf.Center|gopdf.Middle)
}

func (r *shuangxiAProfilePDFRenderer) drawAssessmentLine(x, y float64, slot int, record *model.AssessmentRecordDetailVO) error {
	labels := []string{"第一次评量：", "第二次评量：", "第三次评量：", "第四次评量："}
	label := labels[slot]
	color := shuangxiAProfileColors[slot%len(shuangxiAProfileColors)]
	dateText := ""
	examiner := ""
	colorName := ""
	if record != nil {
		dateText = shuangxiAProfileDateText(record.AssessmentDate)
		examiner = strings.TrimSpace(record.ExaminerName)
		colorName = color.Name
	}
	r.setTextColor(0, 0, 0)
	r.setFont(10.8)
	if err := r.drawInfoField(x, y, 92, label, dateText); err != nil {
		return err
	}
	examinerX := x + 168
	if err := r.drawInfoField(examinerX, y, 50, "评量者：", examiner); err != nil {
		return err
	}
	colorX := x + 272
	if err := r.drawInfoField(colorX, y, 40, "颜色：", ""); err != nil {
		return err
	}
	if colorName != "" {
		r.setTextColor(color.R, color.G, color.B)
		colorLabelWidth, err := r.pdf.MeasureTextWidth("颜色：")
		if err != nil {
			return err
		}
		if err := r.cell(colorX+colorLabelWidth+2, y, 40, 18, colorName, gopdf.Center|gopdf.Middle); err != nil {
			return err
		}
		r.setTextColor(0, 0, 0)
	}
	return nil
}

func (r *shuangxiAProfilePDFRenderer) drawScaleLabels() error {
	const (
		leftX  = 34.0
		scoreX = 138.0
		top    = 170.0
		height = 336.0
	)
	rowH := height / 3
	labels := [][]string{
		{"已发展出适应环", "境需要之能力"},
		{"已发展较多能力，", "只需重点协助，", "便能适应环境之", "需要"},
		{"仅发展些微能力，", "需要特别协助，", "才能适应环境之", "需要"},
		{"尚未开始发展，", "无法适应环境之", "需要"},
	}
	for index, lines := range labels {
		lineY := top + float64(index)*rowH
		y := lineY + 10
		r.setTextColor(0, 0, 0)
		r.setFont(15)
		for lineIndex, line := range lines {
			if err := r.text(leftX, y+float64(lineIndex)*17, line); err != nil {
				return err
			}
		}
		r.setFont(15)
		if err := r.cell(scoreX, lineY-2, 24, 18, fmt.Sprintf("%d", 3-index), gopdf.Center|gopdf.Middle); err != nil {
			return err
		}
	}
	return nil
}

func (r *shuangxiAProfilePDFRenderer) drawGrid(domains []shuangxiAProfileDomain) error {
	const (
		left   = 164.0
		top    = 170.0
		width  = 644.0
		height = 336.0
	)
	colW := width / float64(len(domains))
	rowH := height / 3
	r.pdf.SetStrokeColor(0, 0, 0)
	r.pdf.SetLineWidth(0.95)
	r.pdf.RectFromUpperLeft(left, top, width, height)
	for row := 1; row < 3; row++ {
		y := top + float64(row)*rowH
		r.pdf.Line(left, y, left+width, y)
	}
	for col := 1; col < len(domains); col++ {
		x := left + float64(col)*colW
		r.pdf.Line(x, top, x, top+height)
	}
	r.setTextColor(0, 0, 0)
	r.setFont(14)
	for col, domain := range domains {
		cellX := left + float64(col)*colW
		for level := 3; level >= 1; level-- {
			value := int(math.Round(float64(domain.MaxRawScore) * float64(level) / 3))
			cellY := top + float64(3-level)*rowH
			if err := r.cell(cellX, cellY+2, colW, 18, fmt.Sprintf("%d", value), gopdf.Center|gopdf.Top); err != nil {
				return err
			}
		}
		if err := r.cell(cellX, top+height+8, colW, 18, shuangxiAProfileDomainLabel(domain.Name), gopdf.Center|gopdf.Middle); err != nil {
			return err
		}
		if err := r.cell(cellX, top+height+28, colW, 18, fmt.Sprintf("%d", domain.DomainNo), gopdf.Center|gopdf.Middle); err != nil {
			return err
		}
	}
	return nil
}

func (r *shuangxiAProfilePDFRenderer) drawProfiles(data shuangxiAStaticData, domains []shuangxiAProfileDomain, records []model.AssessmentRecordDetailVO) {
	const (
		left   = 164.0
		top    = 170.0
		width  = 644.0
		height = 336.0
	)
	colW := width / float64(len(domains))
	rowH := height / 3
	baseScoreRects := shuangxiAProfileBaseScoreRects(domains, left, top, width, height)
	chartRect := shuangxiAProfileRect{X: left, Y: top, W: width, H: height}
	for recordIndex, record := range records {
		color := shuangxiAProfileColors[recordIndex%len(shuangxiAProfileColors)]
		scoreByDomain := shuangxiAProfileDomainScoreMap(data, record)
		points := make([]shuangxiAProfilePoint, 0, len(domains))
		for col, domain := range domains {
			raw, ok := scoreByDomain[domain.Code]
			if !ok || domain.MaxRawScore <= 0 {
				continue
			}
			points = append(points, shuangxiAProfilePoint{
				X:     left + float64(col)*colW + colW/2,
				Y:     shuangxiAProfileScoreY(raw, domain.MaxRawScore, top, rowH),
				Score: raw,
			})
		}
		if len(points) == 0 {
			continue
		}
		r.pdf.SetLineType("solid")
		r.pdf.SetLineWidth(1.45)
		r.pdf.SetStrokeColor(color.R, color.G, color.B)
		for index := 1; index < len(points); index++ {
			r.pdf.Line(points[index-1].X, points[index-1].Y, points[index].X, points[index].Y)
		}
		for _, point := range points {
			shuangxiAProfileDrawDot(r.pdf, point, 3.3, color)
			r.drawProfilePointScore(point, color, baseScoreRects, chartRect)
		}
	}
	r.pdf.SetStrokeColor(0, 0, 0)
	r.pdf.SetLineWidth(0.7)
	r.pdf.SetLineType("solid")
}

func shuangxiAProfileScoreY(rawScore, maxRawScore int, top, rowH float64) float64 {
	if maxRawScore <= 0 {
		return top + rowH*3
	}
	ratio := math.Max(0, math.Min(1, float64(rawScore)/float64(maxRawScore)))
	return top + rowH*3 - ratio*rowH*3
}

func shuangxiAProfileBaseScoreRects(domains []shuangxiAProfileDomain, left, top, width, height float64) []shuangxiAProfileRect {
	if len(domains) == 0 {
		return nil
	}
	colW := width / float64(len(domains))
	rowH := height / 3
	rects := make([]shuangxiAProfileRect, 0, len(domains)*3)
	for col := range domains {
		centerX := left + float64(col)*colW + colW/2
		for level := 3; level >= 1; level-- {
			cellY := top + float64(3-level)*rowH
			rects = append(rects, shuangxiAProfileRect{
				X: centerX - 28,
				Y: cellY,
				W: 56,
				H: 22,
			})
		}
	}
	return rects
}

func (r *shuangxiAProfilePDFRenderer) drawProfilePointScore(point shuangxiAProfilePoint, color shuangxiAProfileColor, avoidRects []shuangxiAProfileRect, chartRect shuangxiAProfileRect) {
	r.setFont(13)
	r.setTextColor(color.R, color.G, color.B)
	value := fmt.Sprintf("%d", point.Score)
	textWidth, err := r.pdf.MeasureTextWidth(value)
	if err != nil {
		textWidth = 20
	}
	labelWidth := math.Max(20, textWidth+6)
	labelRect := shuangxiAProfilePointScoreRect(point, labelWidth, 16, avoidRects, chartRect)
	_ = r.cell(labelRect.X-0.25, labelRect.Y, labelRect.W, labelRect.H, value, gopdf.Center|gopdf.Middle)
	_ = r.cell(labelRect.X+0.25, labelRect.Y, labelRect.W, labelRect.H, value, gopdf.Center|gopdf.Middle)
	_ = r.cell(labelRect.X, labelRect.Y-0.15, labelRect.W, labelRect.H, value, gopdf.Center|gopdf.Middle)
	r.setTextColor(0, 0, 0)
}

func shuangxiAProfilePointScoreRect(point shuangxiAProfilePoint, width, height float64, avoidRects []shuangxiAProfileRect, chartRect shuangxiAProfileRect) shuangxiAProfileRect {
	candidates := []shuangxiAProfileRect{
		{X: point.X - width/2, Y: point.Y - 23, W: width, H: height},
		{X: point.X - width/2, Y: point.Y + 7, W: width, H: height},
		{X: point.X + 7, Y: point.Y - 20, W: width, H: height},
		{X: point.X - width - 7, Y: point.Y - 20, W: width, H: height},
		{X: point.X + 7, Y: point.Y - 8, W: width, H: height},
		{X: point.X - width - 7, Y: point.Y - 8, W: width, H: height},
	}
	for _, candidate := range candidates {
		candidate = shuangxiAProfileClampRectToChart(candidate, chartRect)
		if !shuangxiAProfileRectOverlapsAny(candidate, avoidRects) {
			return candidate
		}
	}
	return shuangxiAProfileClampRectToChart(candidates[0], chartRect)
}

func shuangxiAProfileClampRectToChart(rect shuangxiAProfileRect, chart shuangxiAProfileRect) shuangxiAProfileRect {
	if chart.W <= 0 || chart.H <= 0 {
		return rect
	}
	minX := chart.X + 2
	maxX := chart.X + chart.W - rect.W - 2
	if maxX >= minX {
		rect.X = math.Max(minX, math.Min(maxX, rect.X))
	}
	minY := chart.Y - 28
	maxY := chart.Y + chart.H - rect.H + 4
	if maxY >= minY {
		rect.Y = math.Max(minY, math.Min(maxY, rect.Y))
	}
	return rect
}

func shuangxiAProfileRectOverlapsAny(rect shuangxiAProfileRect, others []shuangxiAProfileRect) bool {
	for _, other := range others {
		if shuangxiAProfileRectsOverlap(rect, other) {
			return true
		}
	}
	return false
}

func shuangxiAProfileRectsOverlap(a, b shuangxiAProfileRect) bool {
	const pad = 1.5
	return a.X < b.X+b.W+pad &&
		a.X+a.W+pad > b.X &&
		a.Y < b.Y+b.H+pad &&
		a.Y+a.H+pad > b.Y
}

func shuangxiAProfileDomainScoreMap(data shuangxiAStaticData, record model.AssessmentRecordDetailVO) map[string]int {
	out := make(map[string]int)
	var score shuangxiAAssessmentScoreResponse
	if len(record.ResultJSON) > 0 && json.Unmarshal(record.ResultJSON, &score) == nil {
		for _, row := range score.Result.DomainScores {
			code := strings.TrimSpace(row.DomainCode)
			if code != "" {
				out[code] = row.RawScore
			}
		}
	}
	if len(out) > 0 {
		return out
	}
	itemScores, err := decodeSavedShuangxiAInputScores(record.InputJSON)
	if err != nil {
		return out
	}
	return shuangxiARawScoresByDomainWithData(data, itemScores)
}

func shuangxiAProfileDrawDot(pdf *gopdf.GoPdf, point shuangxiAProfilePoint, radius float64, color shuangxiAProfileColor) {
	pdf.SetLineWidth(0.45)
	pdf.SetFillColor(255, 255, 255)
	pdf.SetStrokeColor(color.R, color.G, color.B)
	pdf.Polygon(shuangxiAProfileCirclePoints(point, radius*1.08, 18), "DF")
	pdf.SetFillColor(color.R, color.G, color.B)
	pdf.Polygon(shuangxiAProfileCirclePoints(point, radius*.78, 16), "F")
}

func shuangxiAProfileCirclePoints(center shuangxiAProfilePoint, radius float64, segments int) []gopdf.Point {
	if segments < 8 {
		segments = 8
	}
	points := make([]gopdf.Point, 0, segments)
	for index := 0; index < segments; index++ {
		angle := float64(index) * 2 * math.Pi / float64(segments)
		points = append(points, gopdf.Point{
			X: center.X + math.Cos(angle)*radius,
			Y: center.Y + math.Sin(angle)*radius,
		})
	}
	return points
}

func shuangxiAProfileGenderText(value string) string {
	switch normalizeShuangxiAGender(value) {
	case "male":
		return "男"
	case "female":
		return "女"
	default:
		return strings.TrimSpace(value)
	}
}

func shuangxiAProfileDateText(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return fmt.Sprintf("%d年%d月%d日", value.Year(), int(value.Month()), value.Day())
}

func shuangxiAProfilePhysiologicalAgeText(birthDate, assessmentDate *time.Time) string {
	if birthDate == nil || birthDate.IsZero() || assessmentDate == nil || assessmentDate.IsZero() {
		return ""
	}
	birth := time.Date(birthDate.Year(), birthDate.Month(), birthDate.Day(), 0, 0, 0, 0, time.UTC)
	target := time.Date(assessmentDate.Year(), assessmentDate.Month(), assessmentDate.Day(), 0, 0, 0, 0, time.UTC)
	if birth.After(target) {
		return ""
	}
	years := target.Year() - birth.Year()
	yearAnchor := time.Date(birth.Year()+years, birth.Month(), birth.Day(), 0, 0, 0, 0, time.UTC)
	if yearAnchor.After(target) {
		years--
		yearAnchor = time.Date(birth.Year()+years, birth.Month(), birth.Day(), 0, 0, 0, 0, time.UTC)
	}
	months := (target.Year()-yearAnchor.Year())*12 + int(target.Month()) - int(yearAnchor.Month())
	monthAnchor := time.Date(yearAnchor.Year(), yearAnchor.Month()+time.Month(months), yearAnchor.Day(), 0, 0, 0, 0, time.UTC)
	if monthAnchor.After(target) {
		months--
		monthAnchor = time.Date(yearAnchor.Year(), yearAnchor.Month()+time.Month(months), yearAnchor.Day(), 0, 0, 0, 0, time.UTC)
	}
	days := int(target.Sub(monthAnchor).Hours() / 24)
	if years > 0 {
		if months > 0 {
			return fmt.Sprintf("%d岁%d个月", years, months)
		}
		if days > 0 {
			return fmt.Sprintf("%d岁%d天", years, days)
		}
		return fmt.Sprintf("%d岁", years)
	}
	if months > 0 {
		return fmt.Sprintf("%d个月", months)
	}
	if days < 0 {
		days = 0
	}
	return fmt.Sprintf("%d天", days)
}

func shuangxiAProfileDomainLabel(name string) string {
	name = strings.TrimSpace(name)
	switch name {
	case "沟通":
		return "沟    通"
	case "认知":
		return "认    知"
	default:
		return name
	}
}

func (r *shuangxiAProfilePDFRenderer) setFont(size float64) {
	r.fontSize = size
	_ = r.pdf.SetFont(shuangxiAProfilePDFFontFamily, "", size)
}

func (r *shuangxiAProfilePDFRenderer) setTextColor(red, green, blue uint8) {
	r.pdf.SetTextColor(red, green, blue)
}

func (r *shuangxiAProfilePDFRenderer) text(x, y float64, value string) error {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	_ = r.pdf.SetFont(shuangxiAProfilePDFFontFamily, "", r.fontSize)
	r.pdf.SetX(x)
	r.pdf.SetY(y)
	return r.pdf.Text(value)
}

func (r *shuangxiAProfilePDFRenderer) cell(x, y, width, height float64, value string, align int) error {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	_ = r.pdf.SetFont(shuangxiAProfilePDFFontFamily, "", r.fontSize)
	r.pdf.SetXY(x, y)
	return r.pdf.CellWithOption(&gopdf.Rect{W: width, H: height}, value, gopdf.CellOption{Align: align})
}

func shuangxiAProfileTextWidth(value string, size float64) float64 {
	width := 0.0
	for _, char := range []rune(value) {
		if char < 128 {
			width += size * 0.55
		} else {
			width += size
		}
	}
	return width
}
