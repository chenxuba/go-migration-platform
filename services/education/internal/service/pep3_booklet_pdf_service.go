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
	// 全局竖向微调：数值越大，所有填充值越靠上；数值越小，所有填充值越靠下。
	pep3BookletPDFLineBaselineGap = 2.2
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

	// 第1部分：儿童资料。格式为 text(x, y, 字号, 值)，x/y 对应底图横线位置。
	r.text(pep3RightPageX(690), 185, 10, record.StudentName)  // 儿童姓名
	r.text(pep3RightPageX(690), 240, 10, record.ExaminerName) // 测试员姓名
	r.text(pep3RightPageX(690), 252, 10, record.Remark)       // 备注

	assessmentYear, assessmentMonth, assessmentDay := dateParts(record.AssessmentDate)
	birthYear, birthMonth, birthDay := dateParts(record.BirthDate)
	// 第1部分：右侧日期/年龄。格式为 center(x, y, 居中宽度, 字号, 值)。
	r.center(pep3RightPageX(950), 220, 48, 9, assessmentYear)                  // 评估日期-年
	r.center(pep3RightPageX(1006), 220, 45, 9, assessmentMonth)                // 评估日期-月
	r.center(pep3RightPageX(1063), 220, 45, 9, assessmentDay)                  // 评估日期-日
	r.center(pep3RightPageX(950), 239, 48, 9, birthYear)                       // 出生日期-年
	r.center(pep3RightPageX(1006), 239, 45, 9, birthMonth)                     // 出生日期-月
	r.center(pep3RightPageX(1063), 239, 45, 9, birthDay)                       // 出生日期-日
	r.center(pep3RightPageX(950), 259, 48, 9, strconv.Itoa(record.AgeYears))   // 年龄-年
	r.center(pep3RightPageX(1006), 259, 45, 9, strconv.Itoa(record.AgeMonths)) // 年龄-月
	r.center(pep3RightPageX(1063), 259, 45, 9, strconv.Itoa(record.AgeDays))   // 年龄-日

	scaleRows := append([]model.PEP3ReportScaleRow{}, buildPEP3ScaleRows(score.Result.Scales, "发展及行为副测验", []string{"CVP", "EL", "RL", "FM", "GM", "VMI", "AE", "SR", "CMB", "CVB"})...)
	scaleRows = append(scaleRows, buildPEP3ScaleRows(score.Result.Scales, "儿童照顾者报告副测验", []string{"PB", "PSC", "AB"})...)
	scaleRowByCode := make(map[string]model.PEP3ReportScaleRow, len(scaleRows))
	for _, row := range scaleRows {
		scaleRowByCode[row.ScaleCode] = row
	}
	// 第2部分：副测验分数每一行的 y 坐标。只调上下位置时改这里。
	scaleY := map[string]float64{
		"CVP": 352, // 认知（语言/语前）
		"EL":  370, // 语言表达
		"RL":  389, // 语言理解
		"FM":  407, // 小肌肉
		"GM":  425, // 大肌肉
		"VMI": 443, // 模仿（视觉/动作）
		"AE":  461, // 情感表达
		"SR":  479, // 社交互助
		"CMB": 496, // 行为特征-非语言
		"CVB": 515, // 行为特征-语言
		"PB":  554, // 问题行为
		"PSC": 573, // 个人自理
		"AB":  592, // 适应行为
	}
	for _, code := range []string{"CVP", "EL", "RL", "FM", "GM", "VMI", "AE", "SR", "CMB", "CVB", "PB", "PSC", "AB"} {
		row, ok := scaleRowByCode[code]
		if !ok {
			continue
		}
		y := scaleY[code]
		// 第2部分：副测验分数每一列的 x 坐标。只调左右位置时改这里。
		r.center(pep3RightPageX(826), y, 38, 8.5, strconv.Itoa(row.RawScore)) // 原积/原始分
		r.center(pep3RightPageX(896), y, 45, 8.5, row.DevelopmentAgeText)     // 发展年龄
		r.center(pep3RightPageX(968), y, 48, 8.5, row.PercentileRankText)     // 百分比级数
		r.center(pep3RightPageX(1036), y, 85, 8.5, row.Level)                 // 发展/适应程度
	}

	compositeRows := buildPEP3CompositeRows(score.Result.Composites, score.Result.Scales)
	// 第3部分：合成分数三行的 y 坐标。只调上下位置时改这里。
	compositeY := map[string]float64{
		pep3score.CompositeCommunication:       674, // 沟通（C）
		pep3score.CompositeMotor:               705, // 体能（M）
		pep3score.CompositeMaladaptiveBehavior: 735, // 行为（MB）
	}
	// 第3部分：标准分小格每一列的 x 坐标。只调左右位置时改这里。
	standardScoreX := map[string]float64{
		"CVP": 133, // CVP 标准分
		"EL":  160, // EL 标准分
		"RL":  187, // RL 标准分
		"FM":  214, // FM 标准分
		"GM":  241, // GM 标准分
		"VMI": 268, // VMI 标准分
		"AE":  295, // AE 标准分
		"SR":  322, // SR 标准分
		"CMB": 349, // CMB 标准分
		"CVB": 376, // CVB 标准分
	}
	for _, row := range compositeRows {
		y := compositeY[row.CompositeCode]
		for _, code := range pep3CompositeScaleCodes() {
			r.center(standardScoreX[code]-12, y, 24, 8.2, row.MemberScaleScores[code]) // 标准分小格单项得分
		}
		// 第3部分：合成分数右侧汇总列的 x 坐标。
		r.center(386, y, 24, 8.2, row.StandardScoreSumText) // 标准分总和
		r.center(424, y, 32, 8.2, row.PercentileRankText)   // 百分比级数
		r.center(463, y, 62, 8.2, row.Level)                // 发展/适应程度
		r.center(525, y, 44, 8.2, row.DevelopmentAgeText)   // 发展年龄
	}
}

func pep3RightPageX(spreadX float64) float64 {
	return spreadX - pep3BookletPDFRightPageOffsetX
}

// text 用于左对齐填值：x 是文字起点，y 是底图横线位置，size 是字号。
func (r pep3BookletPDFRenderer) text(x, y, size float64, value string) {
	value = cleanPEP3BookletPDFValue(value)
	if value == "" {
		return
	}
	_ = r.pdf.SetFont(pep3BookletPDFFontFamily, "", size)
	r.pdf.SetXY(x, y-pep3BookletPDFLineBaselineGap)
	_ = r.pdf.Text(value)
}

// center 用于居中填值：x 是居中区域左边界，y 是底图横线位置，width 是居中区域宽度，size 是字号。
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
