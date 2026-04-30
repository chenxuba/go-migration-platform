package service

import (
	"embed"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/signintech/gopdf"
	"go-migration-platform/pkg/pep3score"
	"go-migration-platform/services/education/internal/model"
)

//go:embed assets/pep3_record_booklet/page_*.jpg
var pep3RecordBookletTemplateImages embed.FS

const (
	pep3BookletPDFSpreadWidth      = 1190.52001953125
	pep3BookletPDFPageWidth        = pep3BookletPDFSpreadWidth / 2
	pep3BookletPDFPageHeight       = 779.4000244140625
	pep3BookletPDFRightPageOffsetX = pep3BookletPDFPageWidth
	pep3BookletPDFFontFamily       = "pep3-cjk"
	pep3BookletPDFLineBaselineGap  = 2.2
)

type pep3BookletPDFRenderer struct {
	pdf *gopdf.GoPdf
}

func (svc *Service) GeneratePEP3AssessmentBookletPDF(userID, recordID int64) (string, []byte, error) {
	detail, err := svc.GetPEP3AssessmentRecord(userID, recordID)
	if err != nil {
		return "", nil, err
	}
	detail = svc.rescorePEP3AssessmentRecordDetail(detail)
	content, err := buildPEP3BookletPDF(detail)
	if err != nil {
		return "", nil, err
	}
	name := nonEmptyString(detail.StudentName, "未命名儿童")
	filename := sanitizeTemplateFileName(fmt.Sprintf("%s-PEP3测试员记录册-%s.pdf", name, time.Now().Format("20060102150405")))
	return filename, content, nil
}

func buildPEP3BookletPDF(record model.AssessmentRecordDetailVO) ([]byte, error) {
	score, err := decodeSavedPEP3Score(record.ResultJSON)
	if err != nil {
		return nil, err
	}
	fontPath, err := resolvePEP3PDFFontPath()
	if err != nil {
		return nil, err
	}

	var pdf gopdf.GoPdf
	pdf.Start(gopdf.Config{
		Unit:     gopdf.UnitPT,
		PageSize: gopdf.Rect{W: pep3BookletPDFPageWidth, H: pep3BookletPDFPageHeight},
	})
	for pageNo := 1; pageNo <= 26; pageNo++ {
		pdf.AddPage()
		raw, err := pep3RecordBookletTemplateImages.ReadFile(fmt.Sprintf("assets/pep3_record_booklet/page_%02d.jpg", pageNo))
		if err != nil {
			return nil, fmt.Errorf("load PEP-3 booklet template page %d: %w", pageNo, err)
		}
		holder, err := gopdf.ImageHolderByBytes(raw)
		if err != nil {
			return nil, fmt.Errorf("decode PEP-3 booklet template page %d: %w", pageNo, err)
		}
		if err := pdf.ImageByHolder(holder, 0, 0, &gopdf.Rect{W: pep3BookletPDFPageWidth, H: pep3BookletPDFPageHeight}); err != nil {
			return nil, fmt.Errorf("draw PEP-3 booklet template page %d: %w", pageNo, err)
		}
	}
	if err := pdf.AddTTFFont(pep3BookletPDFFontFamily, fontPath); err != nil {
		return nil, fmt.Errorf("load PEP-3 PDF font: %w", err)
	}
	renderer := pep3BookletPDFRenderer{pdf: &pdf}
	renderer.drawCoverPage(record, score)

	return pdf.GetBytesPdfReturnErr()
}

func (r pep3BookletPDFRenderer) drawCoverPage(record model.AssessmentRecordDetailVO, score PEP3ScoreResponse) {
	_ = r.pdf.SetPage(1)
	r.pdf.SetTextColor(58, 58, 58)

	r.text(pep3RightPageX(690), 177, 10, record.StudentName)
	r.text(pep3RightPageX(690), 232, 10, record.ExaminerName)
	r.text(pep3RightPageX(690), 252, 10, record.Remark)

	assessmentYear, assessmentMonth, assessmentDay := dateParts(record.AssessmentDate)
	birthYear, birthMonth, birthDay := dateParts(record.BirthDate)
	r.center(pep3RightPageX(956), 217, 48, 9, assessmentYear)
	r.center(pep3RightPageX(1012), 217, 45, 9, assessmentMonth)
	r.center(pep3RightPageX(1069), 217, 45, 9, assessmentDay)
	r.center(pep3RightPageX(956), 237, 48, 9, birthYear)
	r.center(pep3RightPageX(1012), 237, 45, 9, birthMonth)
	r.center(pep3RightPageX(1069), 237, 45, 9, birthDay)
	r.center(pep3RightPageX(956), 257, 48, 9, strconv.Itoa(record.AgeYears))
	r.center(pep3RightPageX(1012), 257, 45, 9, strconv.Itoa(record.AgeMonths))
	r.center(pep3RightPageX(1069), 257, 45, 9, strconv.Itoa(record.AgeDays))

	scaleRows := append([]model.PEP3ReportScaleRow{}, buildPEP3ScaleRows(score.Result.Scales, "发展及行为副测验", []string{"CVP", "EL", "RL", "FM", "GM", "VMI", "AE", "SR", "CMB", "CVB"})...)
	scaleRows = append(scaleRows, buildPEP3ScaleRows(score.Result.Scales, "儿童照顾者报告副测验", []string{"PB", "PSC", "AB"})...)
	scaleRowByCode := make(map[string]model.PEP3ReportScaleRow, len(scaleRows))
	for _, row := range scaleRows {
		scaleRowByCode[row.ScaleCode] = row
	}
	scaleY := map[string]float64{
		"CVP": 347, "EL": 366, "RL": 385, "FM": 404, "GM": 423, "VMI": 442,
		"AE": 461, "SR": 480, "CMB": 499, "CVB": 518,
		"PB": 554, "PSC": 573, "AB": 592,
	}
	for _, code := range []string{"CVP", "EL", "RL", "FM", "GM", "VMI", "AE", "SR", "CMB", "CVB", "PB", "PSC", "AB"} {
		row, ok := scaleRowByCode[code]
		if !ok {
			continue
		}
		y := scaleY[code]
		r.center(pep3RightPageX(827), y, 38, 8.5, strconv.Itoa(row.RawScore))
		r.center(pep3RightPageX(900), y, 45, 8.5, row.DevelopmentAgeText)
		r.center(pep3RightPageX(976), y, 48, 8.5, row.PercentileRankText)
		r.center(pep3RightPageX(1050), y, 85, 8.5, row.Level)
	}

	compositeRows := buildPEP3CompositeRows(score.Result.Composites, score.Result.Scales)
	compositeY := map[string]float64{
		pep3score.CompositeCommunication:       674,
		pep3score.CompositeMotor:               705,
		pep3score.CompositeMaladaptiveBehavior: 735,
	}
	standardScoreX := map[string]float64{
		"CVP": 133, "EL": 160, "RL": 187, "FM": 214, "GM": 241,
		"VMI": 268, "AE": 295, "SR": 322, "CMB": 349, "CVB": 376,
	}
	for _, row := range compositeRows {
		y := compositeY[row.CompositeCode]
		for _, code := range pep3CompositeScaleCodes() {
			r.center(standardScoreX[code]-12, y, 24, 8.2, row.MemberScaleScores[code])
		}
		r.center(386, y, 24, 8.2, row.StandardScoreSumText)
		r.center(424, y, 32, 8.2, row.PercentileRankText)
		r.center(463, y, 62, 8.2, row.Level)
		r.center(525, y, 44, 8.2, row.DevelopmentAgeText)
	}
}

func pep3RightPageX(spreadX float64) float64 {
	return spreadX - pep3BookletPDFRightPageOffsetX
}

func (r pep3BookletPDFRenderer) text(x, y, size float64, value string) {
	value = cleanPEP3BookletPDFValue(value)
	if value == "" {
		return
	}
	_ = r.pdf.SetFont(pep3BookletPDFFontFamily, "", size)
	r.pdf.SetXY(x, y-pep3BookletPDFLineBaselineGap)
	_ = r.pdf.Text(value)
}

func (r pep3BookletPDFRenderer) center(x, y, width, size float64, value string) {
	value = cleanPEP3BookletPDFValue(value)
	if value == "" {
		return
	}
	_ = r.pdf.SetFont(pep3BookletPDFFontFamily, "", size)
	textWidth, err := r.pdf.MeasureTextWidth(value)
	if err != nil || textWidth > width {
		textWidth = 0
	}
	r.pdf.SetXY(x+(width-textWidth)/2, y-pep3BookletPDFLineBaselineGap)
	_ = r.pdf.Text(value)
}

func cleanPEP3BookletPDFValue(value string) string {
	value = strings.TrimSpace(value)
	switch value {
	case "", "-", "--", "待校对":
		return ""
	default:
		return value
	}
}

func dateParts(value *time.Time) (string, string, string) {
	if value == nil || value.IsZero() {
		return "", "", ""
	}
	return strconv.Itoa(value.Year()), strconv.Itoa(int(value.Month())), strconv.Itoa(value.Day())
}

func resolvePEP3PDFFontPath() (string, error) {
	if path := strings.TrimSpace(os.Getenv("PEP3_PDF_FONT_PATH")); path != "" {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
		return "", fmt.Errorf("PEP3_PDF_FONT_PATH does not exist: %s", path)
	}
	for _, path := range []string{
		"/System/Library/Fonts/Supplemental/Arial Unicode.ttf",
		"/Library/Fonts/Arial Unicode.ttf",
		"/usr/share/fonts/truetype/noto/NotoSansCJK-Regular.ttf",
		"/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.otf",
	} {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", errors.New("missing CJK font for PEP-3 PDF; set PEP3_PDF_FONT_PATH to a Chinese-capable .ttf/.otf font")
}
