package service

import (
	"context"
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
	institutionName, err := svc.repo.GetInstitutionName(context.Background(), detail.InstID)
	if err != nil {
		return "", nil, err
	}
	content, err := buildPEP3BookletPDF(detail, institutionName)
	if err != nil {
		return "", nil, err
	}
	name := nonEmptyString(detail.StudentName, "未命名儿童")
	filename := sanitizeTemplateFileName(fmt.Sprintf("%s-PEP3测试员记录册-%s.pdf", name, time.Now().Format("20060102150405")))
	return filename, content, nil
}

func buildPEP3BookletPDF(record model.AssessmentRecordDetailVO, institutionName string) ([]byte, error) {
	score, err := decodeSavedPEP3Score(record.ResultJSON)
	if err != nil {
		return nil, err
	}
	itemScores, _, err := decodeSavedPEP3InputScores(record.InputJSON)
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
	renderer.drawCoverPage(record, score, institutionName)
	renderer.drawDevelopmentBehaviorScorePage2(itemScores)

	return pdf.GetBytesPdfReturnErr()
}

func (r pep3BookletPDFRenderer) drawCoverPage(record model.AssessmentRecordDetailVO, score PEP3ScoreResponse, institutionName string) {
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
		pep3score.CompositeCommunication:       680, // 沟通（C）
		pep3score.CompositeMotor:               698, // 体能（M）
		pep3score.CompositeMaladaptiveBehavior: 715, // 行为（MB）
	}
	// 第3部分：标准分小格每一列的 x 坐标。只调左右位置时改这里。
	standardScoreX := map[string]float64{
		"CVP": 136, // CVP 标准分
		"EL":  162, // EL 标准分
		"RL":  188, // RL 标准分
		"FM":  214, // FM 标准分
		"GM":  240, // GM 标准分
		"VMI": 266, // VMI 标准分
		"AE":  292, // AE 标准分
		"SR":  318, // SR 标准分
		"CMB": 344, // CMB 标准分
		"CVB": 369, // CVB 标准分
	}
	for _, row := range compositeRows {
		y := compositeY[row.CompositeCode]
		for _, code := range pep3CompositeScaleCodes() {
			r.center(standardScoreX[code]-12, y, 24, 8.2, row.MemberScaleScores[code]) // 标准分小格单项得分
		}
		// 第3部分：合成分数右侧汇总列的 x 坐标。
		r.center(384, y, 24, 8.2, row.StandardScoreSumText) // 标准分总和
		r.center(413, y, 32, 8.2, row.PercentileRankText)   // 百分比级数
		r.center(450, y, 62, 8.2, row.Level)                // 发展/适应程度
		r.center(517, y, 44, 8.2, row.DevelopmentAgeText)   // 发展年龄
	}

	r.drawCoverInstitutionMark(institutionName)
}

func pep3RightPageX(spreadX float64) float64 {
	return spreadX - pep3BookletPDFRightPageOffsetX
}

type pep3BookletPDFPoint struct {
	X float64
	Y float64
}

func (r pep3BookletPDFRenderer) drawDevelopmentBehaviorScorePage2(itemScores map[int]int) {
	if len(itemScores) == 0 {
		return
	}
	_ = r.pdf.SetPage(2)
	r.pdf.SetTextColor(58, 58, 58)

	// 第4部分第2页：题目右侧小方框的中心坐标。格式为 题号: {X, Y}。
	itemScoreBoxCenters := map[int]pep3BookletPDFPoint{
		1:  {383.0, 96.3},  // 旋开瓶盖
		2:  {383.0, 115.7}, // 吹肥皂泡
		3:  {383.0, 135.7}, // 目光追视
		4:  {382.7, 155.0}, // 目光追视跨越中线
		5:  {515.7, 194.0}, // 检视触觉块
		6:  {435.7, 237.3}, // 使用万花筒
		7:  {382.3, 271.0}, // 表现出能够使用惯用眼
		8:  {303.0, 321.7}, // 转向手摇铃声
		9:  {435.7, 365.0}, // 模仿按动响铃2次
		10: {435.7, 413.3}, // 手指插入胶泥并做出凹位
		11: {382.3, 433.0}, // 抓握竹棒
		12: {435.3, 453.0}, // 听生日歌假装吹蜡烛
		13: {435.7, 472.3}, // 享受音乐
		14: {435.3, 492.3}, // 搓胶泥条
	}
	for itemNo, point := range itemScoreBoxCenters {
		score, ok := itemScores[itemNo]
		if !ok {
			continue
		}
		r.centerInBox(point.X, point.Y, 18, 8.5, strconv.Itoa(score))
	}

	r.drawDevelopmentBehaviorScorePage2Tally(itemScores)
}

func (r pep3BookletPDFRenderer) drawDevelopmentBehaviorScorePage2Tally(itemScores map[int]int) {
	// 第4部分第2页总和：底部汇总表各副测验列的中心 x 坐标。
	domainX := map[string]float64{
		"CVP": 302.3,
		"EL":  328.7,
		"RL":  355.3,
		"FM":  382.0,
		"GM":  408.7,
		"VMI": 435.3,
		"AE":  462.0,
		"SR":  489.0,
		"CMB": 516.0,
		"CVB": 542.3,
	}
	// 第4部分第2页总和：底部汇总表各得分行的中心 y 坐标。
	tallyY := map[int]float64{
		2: 544.3, // （2）通过 / 恰当
		1: 559.3, // （1）部分通过 / 轻微
		0: 574.3, // （0）未能通过 / 严重
	}
	rawSubtotalY := 529.7 // 第2页总和/原始分小计

	page2Items := map[int]string{
		1: "FM", 2: "FM", 3: "FM", 4: "FM", 5: "CMB", 6: "VMI", 7: "FM",
		8: "CVP", 9: "VMI", 10: "VMI", 11: "FM", 12: "VMI", 13: "VMI", 14: "VMI",
	}
	tally := make(map[string]map[int]int, len(domainX))
	rawSubtotal := make(map[string]int, len(domainX))
	for itemNo, domainCode := range page2Items {
		score, ok := itemScores[itemNo]
		if !ok {
			continue
		}
		if tally[domainCode] == nil {
			tally[domainCode] = map[int]int{}
		}
		tally[domainCode][score]++
		rawSubtotal[domainCode] += score
	}

	for _, domainCode := range pep3BookletDomainOrder() {
		x, ok := domainX[domainCode]
		if !ok {
			continue
		}
		for _, scoreValue := range []int{2, 1, 0} {
			count := tally[domainCode][scoreValue]
			if count > 0 {
				r.centerInBox(x, tallyY[scoreValue], 18, 8, strconv.Itoa(count))
			}
		}
		if subtotal := rawSubtotal[domainCode]; subtotal > 0 {
			r.centerInBox(x, rawSubtotalY, 18, 8, strconv.Itoa(subtotal))
		}
	}
}

func (r pep3BookletPDFRenderer) drawCoverInstitutionMark(institutionName string) {
	institutionName = strings.TrimSpace(institutionName)
	if institutionName == "" {
		return
	}

	// 第1页右下角版权标记遮盖区域。需要改遮盖范围时调 x/y/width/height。
	const (
		x      = 452.0 // 遮盖区域左边界
		y      = 732.0 // 遮盖区域上边界
		width  = 120.0 // 遮盖区域宽度
		height = 28.0  // 遮盖区域高度
	)
	r.pdf.SetFillColor(255, 255, 255)
	r.pdf.RectFromUpperLeftWithStyle(x, y, width, height, "F")

	// 第1页右下角机构名称。需要改文字位置时调 textX/textY/fontSize。
	const (
		textX    = x + 4
		textY    = y + 17
		fontSize = 8.5
	)
	r.pdf.SetTextColor(58, 58, 58)
	r.text(textX, textY, fontSize, institutionName)
}

func (r pep3BookletPDFRenderer) centerInBox(centerX, centerY, width, size float64, value string) {
	value = cleanPEP3BookletPDFValue(value)
	if value == "" {
		return
	}
	_ = r.pdf.SetFont(pep3BookletPDFFontFamily, "", size)
	textWidth, err := r.pdf.MeasureTextWidth(value)
	if err != nil || textWidth > width {
		textWidth = 0
	}
	r.pdf.SetXY(centerX-textWidth/2, centerY+size*0.35)
	_ = r.pdf.Text(value)
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
