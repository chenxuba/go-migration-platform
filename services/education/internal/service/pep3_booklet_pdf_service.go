package service

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
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
	itemScores, rawScores, err := decodeSavedPEP3InputScores(record.InputJSON)
	if err != nil {
		return nil, err
	}
	itemRecordValues, err := decodeSavedPEP3ItemRecordValues(record.InputJSON)
	if err != nil {
		return nil, err
	}
	if len(rawScores) == 0 {
		rawScores = rawScoresFromPEP3Result(score.Result.Scales)
	}
	items, err := loadPEP3BookletItems()
	if err != nil {
		return nil, err
	}
	normRecords, err := loadPEP3BookletPDFNormRecords()
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
	itemDomainByNo := pep3BookletPDFItemDomainMap(items)
	renderer.drawCoverPage(record, score, institutionName)
	renderer.drawDevelopmentBehaviorScorePages(itemScores, itemDomainByNo)
	renderer.drawDevelopmentBehaviorRecordValues(itemRecordValues)
	renderer.drawDevelopmentBehaviorRawTotalTable(score, rawScores, itemScores, itemDomainByNo)
	renderer.drawDevelopmentProfilePage(record, score, rawScores, itemScores, itemDomainByNo, normRecords)

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

type pep3BookletPDFRect struct {
	X float64
	Y float64
	W float64
	H float64
}

type pep3BookletPDFItemPageLayout struct {
	PageNo       int
	StartItemNo  int
	ItemCenters  []pep3BookletPDFPoint
	DomainX      map[string]float64
	RawSubtotalY float64
	TallyY       map[int]float64
}

type pep3BookletPDFDomainScoreSummary struct {
	Answered    int
	RawSubtotal int
	ScoreCounts map[int]int
}

type pep3BookletPDFIntValue struct {
	Value   int
	Present bool
}

type pep3BookletPDFRecordFieldPlacement struct {
	PageNo      int                                 // PDF页码；第4部分儿童表现记录对应记录册第2-16页。
	ItemNo      int                                 // 题号。
	FieldKey    string                              // 前端/接口里的儿童表现记录字段key。
	X           float64                             // 文本横线左边界；向右调大，向左调小。
	Y           float64                             // 文本横线基线位置；向下调大，向上调小。
	Width       float64                             // 文本在横线上的居中区域宽度。
	Size        float64                             // 字号。
	TextLines   []pep3BookletPDFTextLine            // 多行文本横线；配置后按行左对齐并自动换行。
	OptionRects map[string]pep3BookletPDFRect       // 斜杠选项：每个可圈选文字的独立椭圆坐标。
	OptionMarks map[string]pep3BookletPDFOptionMark // 横线选项：每个可选横线的独立填充坐标。
}

type pep3BookletPDFTextLine struct {
	X     float64 // 该行文字起点；向右调大，向左调小。
	Y     float64 // 该行横线基线；向下调大，向上调小。
	Width float64 // 该行可写宽度；超出会换到下一条 TextLines。
}

type pep3BookletPDFOptionMark struct {
	X     float64 // 标记横线左边界；向右调大，向左调小。
	Y     float64 // 标记横线基线位置；向下调大，向上调小。
	Width float64 // 标记居中区域宽度。
	Size  float64 // 标记字号。
	Text  string  // 为空时默认填“√”。
}

type pep3BookletPDFProfileScore struct {
	PassedScore  int
	PartialScore int
	TotalScore   int
	HasBreakdown bool
	HasTotal     bool
}

func (r pep3BookletPDFRenderer) drawDevelopmentBehaviorScorePages(itemScores map[int]int, itemDomainByNo map[int]string) {
	if len(itemScores) == 0 {
		return
	}
	r.pdf.SetTextColor(58, 58, 58)

	for _, layout := range pep3BookletPDFItemPageLayouts() {
		_ = r.pdf.SetPage(layout.PageNo)
		for index, point := range layout.ItemCenters {
			itemNo := layout.StartItemNo + index
			score, ok := itemScores[itemNo]
			if !ok {
				continue
			}
			r.centerInBox(point.X, point.Y, 18, 8.5, strconv.Itoa(score))
		}

		r.drawDevelopmentBehaviorScorePageTally(layout, itemScores, itemDomainByNo)
	}
}

func (r pep3BookletPDFRenderer) drawDevelopmentBehaviorRecordValues(itemRecordValues map[int]map[string]any) {
	if len(itemRecordValues) == 0 {
		return
	}
	r.pdf.SetTextColor(58, 58, 58)
	r.pdf.SetStrokeColor(92, 92, 92)
	r.pdf.SetLineWidth(0.8)
	for _, placement := range pep3BookletPDFRecordFieldPlacements() {
		values := itemRecordValues[placement.ItemNo]
		if len(values) == 0 {
			continue
		}
		value, ok := values[placement.FieldKey]
		if !ok {
			continue
		}
		_ = r.pdf.SetPage(placement.PageNo)
		if len(placement.OptionRects) > 0 {
			r.drawDevelopmentBehaviorRecordOptionValue(placement, value)
			continue
		}
		if len(placement.OptionMarks) > 0 {
			r.drawDevelopmentBehaviorRecordOptionMark(placement, value)
			continue
		}
		text := pep3BookletPDFRecordValueTextForField(placement.ItemNo, placement.FieldKey, value)
		if text == "" {
			continue
		}
		if len(placement.TextLines) > 0 {
			r.multilineText(placement.TextLines, placement.Size, text)
			continue
		}
		r.center(placement.X, placement.Y, placement.Width, placement.Size, text)
	}
}

func (r pep3BookletPDFRenderer) drawDevelopmentBehaviorRecordOptionValue(placement pep3BookletPDFRecordFieldPlacement, value any) {
	for _, token := range pep3BookletPDFRecordValueTokensForField(placement.ItemNo, placement.FieldKey, value) {
		rect, ok := placement.OptionRects[token]
		if !ok {
			continue
		}
		r.pdf.Oval(rect.X, rect.Y, rect.X+rect.W, rect.Y+rect.H)
	}
}

func (r pep3BookletPDFRenderer) drawDevelopmentBehaviorRecordOptionMark(placement pep3BookletPDFRecordFieldPlacement, value any) {
	for _, token := range pep3BookletPDFRecordValueTokensForField(placement.ItemNo, placement.FieldKey, value) {
		mark, ok := placement.OptionMarks[token]
		if !ok {
			continue
		}
		text := mark.Text
		size := mark.Size
		if text == "" {
			text = "√"
			size += 2.0
			r.centerBold(mark.X, mark.Y, mark.Width, size, text)
			continue
		}
		r.center(mark.X, mark.Y, mark.Width, size, text)
	}
}

func (r pep3BookletPDFRenderer) drawDevelopmentBehaviorScorePageTally(layout pep3BookletPDFItemPageLayout, itemScores map[int]int, itemDomainByNo map[int]string) {
	summaryByDomain := pep3BookletPDFLayoutScoreSummary(layout, itemScores, itemDomainByNo)

	for _, domainCode := range pep3BookletDomainOrder() {
		x, ok := layout.DomainX[domainCode]
		if !ok {
			continue
		}
		summary := summaryByDomain[domainCode]
		for _, scoreValue := range []int{2, 1, 0} {
			count := summary.ScoreCounts[scoreValue]
			if count > 0 {
				r.centerInBox(x, layout.TallyY[scoreValue], 18, 8, strconv.Itoa(count))
			}
		}
		if summary.Answered > 0 {
			r.centerInBox(x, layout.RawSubtotalY, 18, 8, strconv.Itoa(summary.RawSubtotal))
		}
	}
}

func (r pep3BookletPDFRenderer) drawDevelopmentBehaviorRawTotalTable(score PEP3ScoreResponse, rawScores map[string]int, itemScores map[int]int, itemDomainByNo map[int]string) {
	if len(itemScores) == 0 && len(rawScores) == 0 && len(score.Result.Scales) == 0 {
		return
	}
	_ = r.pdf.SetPage(17)
	r.pdf.SetTextColor(58, 58, 58)

	// 第17页：发展及行为副测验原积分总和表，每个副测验列的中心 x 坐标。
	// 如果横向位置有偏差，只调这里的 x 值。
	xByDomain := map[string]float64{
		"CVP": 223.7, // CVP 列
		"EL":  249.7, // EL 列
		"RL":  276.3, // RL 列
		"FM":  303.0, // FM 列
		"GM":  329.6, // GM 列
		"VMI": 356.2, // VMI 列
		"AE":  382.9, // AE 列
		"SR":  409.5, // SR 列
		"CMB": 436.1, // CMB 列
		"CVB": 462.8, // CVB 列
	}
	// 第17页：第2页总和到第16页总和，每一行的中心 y 坐标。
	// 如果整行上下有偏差，只调对应页码的 y 值。
	yByPageNo := map[int]float64{
		2:  115.0, // 第2页总和
		3:  131.6, // 第3页总和
		4:  148.2, // 第4页总和
		5:  165.2, // 第5页总和
		6:  182.2, // 第6页总和
		7:  198.8, // 第7页总和
		8:  215.8, // 第8页总和
		9:  233.5, // 第9页总和
		10: 250.8, // 第10页总和
		11: 267.4, // 第11页总和
		12: 284.4, // 第12页总和
		13: 301.8, // 第13页总和
		14: 318.7, // 第14页总和
		15: 335.4, // 第15页总和
		16: 352.4, // 第16页总和
	}

	for _, layout := range pep3BookletPDFItemPageLayouts() {
		y, ok := yByPageNo[layout.PageNo]
		if !ok {
			continue
		}
		summaryByDomain := pep3BookletPDFLayoutScoreSummary(layout, itemScores, itemDomainByNo)
		for _, domainCode := range pep3BookletDomainOrder() {
			x, ok := xByDomain[domainCode]
			if !ok {
				continue
			}
			summary := summaryByDomain[domainCode]
			if summary.Answered > 0 {
				r.centerInBox(x, y, 18, 8, strconv.Itoa(summary.RawSubtotal))
			}
		}
	}

	// 第17页：底部“原积分总和（抄写到第1页）”粉色格子的中心 y 坐标。
	const rawTotalY = 381.0
	totalByDomain := pep3BookletPDFRawScoreTotals(score.Result.Scales, rawScores, itemScores, itemDomainByNo)
	for _, domainCode := range pep3BookletDomainOrder() {
		total := totalByDomain[domainCode]
		if total.Present {
			r.centerInBox(xByDomain[domainCode], rawTotalY, 18, 8, strconv.Itoa(total.Value))
		}
	}
}

func (r pep3BookletPDFRenderer) drawDevelopmentProfilePage(record model.AssessmentRecordDetailVO, score PEP3ScoreResponse, rawScores map[string]int, itemScores map[int]int, itemDomainByNo map[int]string, normRecords []pep3score.NormRecord) {
	_ = r.pdf.SetPage(19)
	r.pdf.SetTextColor(58, 58, 58)

	// 第19页：顶部儿童姓名、评估日期、年龄（月龄）的位置。
	// y 是底图横线位置；如果文字没压在线上，只调对应 y 值。
	const topInfoY = 51.2
	r.text(170.0, topInfoY, 9, record.StudentName)                            // 儿童姓名
	r.text(352.0, topInfoY, 9, formatReportDate(record.AssessmentDate))       // 评估日期
	r.center(431.0, topInfoY, 46.0, 8.5, pep3BookletPDFAgeMonthsText(record)) // 评估年龄（月龄）

	// 第19页：底部两个分数框的中心 x 坐标。
	// 如果底部“通过项目分数/部份通过项目分数”左右不齐，只调这里。
	bottomBoxXByScale := map[string]float64{
		"CVP": 131.2, // CVP 底部分数框中心 x
		"EL":  188.1, // EL 底部分数框中心 x
		"RL":  244.4, // RL 底部分数框中心 x
		"FM":  301.1, // FM 底部分数框中心 x
		"GM":  357.3, // GM 底部分数框中心 x
		"VMI": 411.2, // VMI 底部分数框中心 x
		"PSC": 461.5, // PSC 底部分数框中心 x
	}
	// 第19页：发展表现图里椭圆/折线点的中心 x 坐标。
	// 这里要对准底图每列“原积分数值”的中心；如果圆圈没圈在数值正中间，只调这里。
	pointXByScale := map[string]float64{
		"CVP": 130.9, // CVP 图表数值列中心 x
		"EL":  185.7, // EL 图表数值列中心 x
		"RL":  240.3, // RL 图表数值列中心 x
		"FM":  300.4, // FM 图表数值列中心 x
		"GM":  355.9, // GM 图表数值列中心 x
		"VMI": 409.8, // VMI 图表数值列中心 x
		"PSC": 461.0, // PSC 图表数值列中心 x
	}
	// 第19页：每个圆圈的独立微调量。正数 x 往右，负数 y 往上。
	// 所有列都显式写在这里；某列不需要微调时填 0，后面逐个点手工校准就改对应这一行。
	pointOffsetByScale := map[string]pep3BookletPDFPoint{
		"CVP": {X: 0.0, Y: 0.0}, // CVP 圆圈微调
		"EL":  {X: 0.0, Y: 0.0}, // EL 圆圈微调
		"RL":  {X: 1, Y: 0.0},   // RL 圆圈微调：25 那一列往右一点
		"FM":  {X: 0.0, Y: -1},  // FM 圆圈微调：34 那个往上一点
		"GM":  {X: 1, Y: -1},    // GM 圆圈微调：24 那一列往右一点
		"VMI": {X: 1, Y: -1.8},  // VMI 圆圈微调：12 那一列往上一点
		"PSC": {X: 0.0, Y: 0.0}, // PSC 圆圈微调
	}
	order := []string{"CVP", "EL", "RL", "FM", "GM", "VMI", "PSC"}

	profileScores := pep3BookletPDFProfileScores(score.Result.Scales, rawScores, itemScores, itemDomainByNo)
	// 第19页：底部“通过项目分数”和“部份通过项目分数”两个计数框的中心 y 坐标。
	// 如果底部数字上下不齐，只调这两个 y 值。
	const (
		passedScoreY  = 728.0 // 通过项目分数
		partialScoreY = 746.0 // 部份通过项目分数
	)
	for _, scaleCode := range order {
		x := bottomBoxXByScale[scaleCode]
		profileScore := profileScores[scaleCode]
		if !profileScore.HasBreakdown {
			continue
		}
		r.centerInBox(x, passedScoreY, 24, 8, strconv.Itoa(profileScore.PassedScore))   // 通过项目分数：2分项目贡献的分数总和
		r.centerInBox(x, partialScoreY, 24, 8, strconv.Itoa(profileScore.PartialScore)) // 部份通过项目分数：1分项目贡献的分数总和
	}

	r.drawDevelopmentProfileLine(profileScores, normRecords, order, pointXByScale, pointOffsetByScale)
}

func (r pep3BookletPDFRenderer) drawDevelopmentProfileLine(profileScores map[string]pep3BookletPDFProfileScore, normRecords []pep3score.NormRecord, order []string, pointXByScale map[string]float64, pointOffsetByScale map[string]pep3BookletPDFPoint) {
	// 第19页：发展表现图纵轴。topY 对应92个月这一行，rowGap 是相邻月龄行距。
	// 折线整体上下偏差时调 topY；行距不贴合月份网格时调 rowGap。
	const (
		topY   = 81.6
		rowGap = 7.82
	)
	// 第19页：折线点位椭圆。正版样式是用横向空心椭圆圈住该列数值/范围，不填充。
	// 椭圆过宽/过窄调 markerRadiusX，过高/过矮调 markerRadiusY。
	const (
		markerRadiusX = 7.0
		markerRadiusY = 4.0
	)
	type graphPoint struct {
		x float64
		y float64
	}
	points := make([]graphPoint, 0, len(order))
	// 第19页：折线和椭圆颜色。正版不是纯黑，偏灰；要更深/更浅时调这里。
	r.pdf.SetStrokeColor(92, 92, 92)
	r.pdf.SetLineWidth(1.0)
	var lastPoint graphPoint
	hasLastPoint := false
	for _, scaleCode := range order {
		profileScore := profileScores[scaleCode]
		if !profileScore.HasTotal {
			hasLastPoint = false
			continue
		}
		months, lessThan, greaterThan, ok := pep3BookletPDFDevelopmentAgeMonthsByRawScore(normRecords, scaleCode, profileScore.TotalScore)
		if !ok {
			hasLastPoint = false
			continue
		}
		offset := pointOffsetByScale[scaleCode]
		point := graphPoint{
			x: pointXByScale[scaleCode] + offset.X,
			y: pep3BookletPDFDevelopmentProfileY(months, lessThan, greaterThan, topY, rowGap) + offset.Y,
		}
		if hasLastPoint {
			x1, y1, x2, y2 := pep3BookletPDFLineBetweenEllipseBorders(lastPoint.x, lastPoint.y, point.x, point.y, markerRadiusX, markerRadiusY)
			r.pdf.Line(x1, y1, x2, y2)
		}
		points = append(points, point)
		lastPoint = point
		hasLastPoint = true
	}
	if len(points) == 0 {
		return
	}
	for _, point := range points {
		r.pdf.Oval(point.x-markerRadiusX, point.y-markerRadiusY, point.x+markerRadiusX, point.y+markerRadiusY)
	}
}

func pep3BookletPDFLineBetweenEllipseBorders(x1, y1, x2, y2, radiusX, radiusY float64) (float64, float64, float64, float64) {
	dx := x2 - x1
	dy := y2 - y1
	length := math.Hypot(dx, dy)
	if length == 0 || radiusX <= 0 || radiusY <= 0 {
		return x1, y1, x2, y2
	}
	unitX := dx / length
	unitY := dy / length
	offset := pep3BookletPDFEllipseRadiusOnVector(unitX, unitY, radiusX, radiusY)
	if offset <= 0 || offset*2 >= length {
		return x1, y1, x2, y2
	}
	return x1 + unitX*offset, y1 + unitY*offset, x2 - unitX*offset, y2 - unitY*offset
}

func pep3BookletPDFEllipseRadiusOnVector(unitX, unitY, radiusX, radiusY float64) float64 {
	denominator := math.Sqrt((unitX*unitX)/(radiusX*radiusX) + (unitY*unitY)/(radiusY*radiusY))
	if denominator == 0 {
		return 0
	}
	return 1 / denominator
}

func pep3BookletPDFLayoutScoreSummary(layout pep3BookletPDFItemPageLayout, itemScores map[int]int, itemDomainByNo map[int]string) map[string]pep3BookletPDFDomainScoreSummary {
	summaryByDomain := make(map[string]pep3BookletPDFDomainScoreSummary, len(layout.DomainX))
	for index := range layout.ItemCenters {
		itemNo := layout.StartItemNo + index
		score, ok := itemScores[itemNo]
		if !ok {
			continue
		}
		domainCode := itemDomainByNo[itemNo]
		if domainCode == "" {
			continue
		}
		summary := summaryByDomain[domainCode]
		if summary.ScoreCounts == nil {
			summary.ScoreCounts = map[int]int{}
		}
		summary.Answered++
		summary.RawSubtotal += score
		summary.ScoreCounts[score]++
		summaryByDomain[domainCode] = summary
	}
	return summaryByDomain
}

func pep3BookletPDFAllItemScoreSummary(itemScores map[int]int, itemDomainByNo map[int]string) map[string]pep3BookletPDFDomainScoreSummary {
	out := make(map[string]pep3BookletPDFDomainScoreSummary)
	for _, layout := range pep3BookletPDFItemPageLayouts() {
		for domainCode, pageSummary := range pep3BookletPDFLayoutScoreSummary(layout, itemScores, itemDomainByNo) {
			summary := out[domainCode]
			if summary.ScoreCounts == nil {
				summary.ScoreCounts = map[int]int{}
			}
			summary.Answered += pageSummary.Answered
			summary.RawSubtotal += pageSummary.RawSubtotal
			for scoreValue, count := range pageSummary.ScoreCounts {
				summary.ScoreCounts[scoreValue] += count
			}
			out[domainCode] = summary
		}
	}
	return out
}

func pep3BookletPDFRawScoreTotals(scales map[string]pep3score.ScaleResult, rawScores map[string]int, itemScores map[int]int, itemDomainByNo map[int]string) map[string]pep3BookletPDFIntValue {
	out := make(map[string]pep3BookletPDFIntValue, len(pep3BookletDomainOrder()))
	itemSummary := pep3BookletPDFAllItemScoreSummary(itemScores, itemDomainByNo)
	for _, domainCode := range pep3BookletDomainOrder() {
		if scale, ok := scales[domainCode]; ok {
			out[domainCode] = pep3BookletPDFIntValue{Value: scale.RawScore, Present: true}
			continue
		}
		if rawScore, ok := rawScores[domainCode]; ok {
			out[domainCode] = pep3BookletPDFIntValue{Value: rawScore, Present: true}
			continue
		}
		if summary := itemSummary[domainCode]; summary.Answered > 0 {
			out[domainCode] = pep3BookletPDFIntValue{Value: summary.RawSubtotal, Present: true}
		}
	}
	return out
}

func pep3BookletPDFProfileScores(scales map[string]pep3score.ScaleResult, rawScores map[string]int, itemScores map[int]int, itemDomainByNo map[int]string) map[string]pep3BookletPDFProfileScore {
	out := make(map[string]pep3BookletPDFProfileScore)
	itemSummary := pep3BookletPDFAllItemScoreSummary(itemScores, itemDomainByNo)
	for _, scaleCode := range []string{"CVP", "EL", "RL", "FM", "GM", "VMI", "PSC"} {
		if summary := itemSummary[scaleCode]; summary.Answered > 0 {
			passedScore := summary.ScoreCounts[2] * 2
			partialScore := summary.ScoreCounts[1]
			out[scaleCode] = pep3BookletPDFProfileScore{
				PassedScore:  passedScore,
				PartialScore: partialScore,
				TotalScore:   passedScore + partialScore,
				HasBreakdown: true,
				HasTotal:     true,
			}
			continue
		}
		if scale, ok := scales[scaleCode]; ok {
			out[scaleCode] = pep3BookletPDFProfileScore{TotalScore: scale.RawScore, HasTotal: true}
			continue
		}
		if rawScore, ok := rawScores[scaleCode]; ok {
			out[scaleCode] = pep3BookletPDFProfileScore{TotalScore: rawScore, HasTotal: true}
		}
	}
	return out
}

func pep3BookletPDFDevelopmentAgeMonths(value *pep3score.NormValue) (float64, bool, bool, bool) {
	if value == nil {
		return 0, false, false, false
	}
	text := strings.TrimSpace(value.Text)
	lessThan := value.Comparator == "<" || strings.HasPrefix(text, "<") || strings.HasPrefix(text, "＜")
	greaterThan := value.Comparator == ">" || strings.HasPrefix(text, ">") || strings.HasPrefix(text, "＞")
	if value.Number != nil {
		return float64(*value.Number), lessThan, greaterThan, true
	}
	if months, ok := pep3BookletPDFFirstNumberOrRangeAverage(text); ok {
		return months, lessThan, greaterThan, true
	}
	return 0, lessThan, greaterThan, false
}

func pep3BookletPDFFirstNumberOrRangeAverage(value string) (float64, bool) {
	numbers := make([]int, 0, 2)
	current := -1
	for _, char := range value {
		if char >= '0' && char <= '9' {
			if current < 0 {
				current = 0
			}
			current = current*10 + int(char-'0')
			continue
		}
		if current >= 0 {
			numbers = append(numbers, current)
			current = -1
		}
	}
	if current >= 0 {
		numbers = append(numbers, current)
	}
	if len(numbers) == 0 {
		return 0, false
	}
	if len(numbers) >= 2 && (strings.Contains(value, "-") || strings.Contains(value, "－") || strings.Contains(value, "~") || strings.Contains(value, "至")) {
		return float64(numbers[0]+numbers[1]) / 2, true
	}
	return float64(numbers[0]), true
}

func pep3BookletPDFDevelopmentAgeMonthsByRawScore(normRecords []pep3score.NormRecord, scaleCode string, rawScore int) (float64, bool, bool, bool) {
	for _, record := range normRecords {
		if record.TableType != pep3score.TableDevelopmentAge || record.ScaleCode != scaleCode {
			continue
		}
		if !pep3BookletPDFRawScoreInRange(rawScore, record.RawScoreMin, record.RawScoreMax) {
			continue
		}
		lessThan := record.DevelopmentAgeComparator == "<" || strings.HasPrefix(record.DevelopmentAgeMonthsLabel, "<") || strings.HasPrefix(record.DevelopmentAgeMonthsLabel, "＜")
		greaterThan := record.DevelopmentAgeComparator == ">" || strings.HasPrefix(record.DevelopmentAgeMonthsLabel, ">") || strings.HasPrefix(record.DevelopmentAgeMonthsLabel, "＞")
		if record.DevelopmentAgeMonths != nil {
			return float64(*record.DevelopmentAgeMonths), lessThan, greaterThan, true
		}
		if months, ok := pep3BookletPDFFirstNumberOrRangeAverage(record.DevelopmentAgeMonthsLabel); ok {
			return months, lessThan, greaterThan, true
		}
		if lessThan {
			return 12, true, false, true
		}
		if greaterThan {
			return 92, false, true, true
		}
	}
	return 0, false, false, false
}

func pep3BookletPDFRawScoreInRange(rawScore int, min, max *int) bool {
	if min == nil && max == nil {
		return false
	}
	if min != nil && rawScore < *min {
		return false
	}
	if max != nil && rawScore > *max {
		return false
	}
	return true
}

func pep3BookletPDFDevelopmentProfileY(months float64, lessThan, greaterThan bool, topY, rowGap float64) float64 {
	const (
		minMonth = 12.0
		maxMonth = 92.0
	)
	if (lessThan && months <= minMonth) || months < minMonth {
		return topY + 81*rowGap
	}
	if (greaterThan && months >= maxMonth) || months > maxMonth {
		return topY
	}
	return topY + (maxMonth-months)*rowGap
}

func pep3BookletPDFRecordValueText(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(typed)
	case []string:
		return strings.Join(typed, "、")
	case []any:
		parts := make([]string, 0, len(typed))
		for _, item := range typed {
			if text := pep3BookletPDFRecordValueText(item); text != "" {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, "、")
	default:
		return strings.TrimSpace(fmt.Sprint(value))
	}
}

func pep3BookletPDFRecordValueTextForField(itemNo int, fieldKey string, value any) string {
	optionLabels := map[string]string{}
	for _, field := range pep3ItemRecordFields(itemNo) {
		if field.Key != fieldKey {
			continue
		}
		for _, option := range field.Options {
			optionLabels[option.Value] = option.Label
		}
		break
	}
	if len(optionLabels) == 0 {
		return pep3BookletPDFRecordValueText(value)
	}
	return pep3BookletPDFRecordValueLabelText(value, optionLabels)
}

func pep3BookletPDFRecordValueLabelText(value any, optionLabels map[string]string) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		text := strings.TrimSpace(typed)
		if label, ok := optionLabels[text]; ok {
			return label
		}
		return text
	case []string:
		parts := make([]string, 0, len(typed))
		for _, item := range typed {
			if text := pep3BookletPDFRecordValueLabelText(item, optionLabels); text != "" {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, "、")
	case []any:
		parts := make([]string, 0, len(typed))
		for _, item := range typed {
			if text := pep3BookletPDFRecordValueLabelText(item, optionLabels); text != "" {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, "、")
	default:
		text := strings.TrimSpace(fmt.Sprint(value))
		if label, ok := optionLabels[text]; ok {
			return label
		}
		return text
	}
}

func pep3BookletPDFRecordValueTokens(value any) []string {
	switch typed := value.(type) {
	case nil:
		return nil
	case string:
		text := strings.TrimSpace(typed)
		if text == "" {
			return nil
		}
		return []string{text}
	case []string:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if token := strings.TrimSpace(item); token != "" {
				out = append(out, token)
			}
		}
		return out
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			out = append(out, pep3BookletPDFRecordValueTokens(item)...)
		}
		return out
	default:
		text := strings.TrimSpace(fmt.Sprint(value))
		if text == "" {
			return nil
		}
		return []string{text}
	}
}

func pep3BookletPDFRecordValueTokensForField(itemNo int, fieldKey string, value any) []string {
	tokens := pep3BookletPDFRecordValueTokens(value)
	if len(tokens) == 0 {
		return nil
	}
	aliases := map[string]string{}
	for _, field := range pep3ItemRecordFields(itemNo) {
		if field.Key != fieldKey {
			continue
		}
		for _, option := range field.Options {
			aliases[option.Value] = option.Label
			aliases[option.Label] = option.Value
		}
		break
	}
	out := make([]string, 0, len(tokens)*2)
	seen := map[string]bool{}
	for _, token := range tokens {
		if token == "" || seen[token] {
			continue
		}
		out = append(out, token)
		seen[token] = true
		if alias := aliases[token]; alias != "" && !seen[alias] {
			out = append(out, alias)
			seen[alias] = true
		}
	}
	return out
}

func pep3BookletPDFAgeMonthsText(record model.AssessmentRecordDetailVO) string {
	months := record.NormAgeMonths
	if months == 0 {
		months = record.AgeYears*12 + record.AgeMonths
	}
	if months <= 0 {
		return ""
	}
	return strconv.Itoa(months) + "个月"
}

func loadPEP3BookletPDFNormRecords() ([]pep3score.NormRecord, error) {
	dataDir, err := resolvePEP3DataDir()
	if err != nil {
		return nil, err
	}
	paths := []string{filepath.Join(dataDir, pep3NormFile)}
	correctionPath := filepath.Join(dataDir, pep3CorrectionFile)
	if fileExists(correctionPath) {
		paths = append(paths, correctionPath)
	}
	records, err := pep3score.LoadMergedNormRecordsFiles(paths...)
	if err != nil {
		return nil, fmt.Errorf("load PEP-3 booklet PDF norm records: %w", err)
	}
	return records, nil
}

func pep3BookletPDFItemDomainMap(items []pep3BookletItemDefinition) map[int]string {
	out := make(map[int]string, len(items))
	for _, item := range items {
		if item.ItemNo > 0 {
			out[item.ItemNo] = strings.ToUpper(strings.TrimSpace(item.DomainCode))
		}
	}
	return out
}

func pep3BookletPDFRecordFieldPlacements() []pep3BookletPDFRecordFieldPlacement {
	// 儿童表现记录填空和选择项坐标。这里按“页码 + 题号 + 字段 key”逐项配置，
	// 不再共用通用偏移。文本类字段 X/Y/Width 对准底图横线；斜杠选项使用
	// OptionRects 圈出底图上的选项文字；横线选项使用 OptionMarks 填到对应横线。
	return []pep3BookletPDFRecordFieldPlacement{
		// 第2页：1-14
		pep3PDFRecordRects(2, 5, "touch_block_reaction", map[string]pep3BookletPDFRect{ // 5 检视触觉块：无兴趣/怪异兴趣
			"no_interest":      pep3PDFCircle(195.0, 186.0, 30.0, 11.5), // 无兴趣
			"unusual_interest": pep3PDFCircle(237.0, 186.0, 40.0, 11.5), // 怪异兴趣
		}),
		pep3PDFRecordMarks(2, 6, "kaleidoscope_action", map[string]pep3BookletPDFOptionMark{ // 6 使用万花筒：观看/扭动/观看+扭动
			"watch":          pep3PDFMark(193.3, 233.3, 22.0),
			"turn":           pep3PDFMark(239.3, 233.3, 22.0),
			"watch_and_turn": pep3PDFMark(223.3, 248.7, 22.7),
		}),
		pep3PDFRecordText(2, 7, "first_observation", 227.3, 268.3, 22.7),  // 7 第1次观察
		pep3PDFRecordText(2, 7, "second_observation", 227.3, 283.7, 22.7), // 7 第2次观察
		pep3PDFRecordMarks(2, 9, "bell_attempts", map[string]pep3BookletPDFOptionMark{ // 9 响铃尝试
			"first_attempt":  pep3PDFMark(226.7, 361.3, 23.3),
			"second_attempt": pep3PDFMark(227.3, 376.7, 22.7),
		}),

		// 第3页：15-27
		pep3PDFRecordMarks(3, 16, "daily_actions", map[string]pep3BookletPDFOptionMark{ // 16 模仿日常动作
			"喂食": pep3PDFMark(196.0, 116.0, 22.7),
			"饮水": pep3PDFMark(242.0, 116.0, 22.0),
			"刷牙": pep3PDFMark(196.0, 131.3, 22.7),
			"抹鼻": pep3PDFMark(242.0, 131.3, 22.0),
		}),
		pep3PDFRecordMarks(3, 17, "puppet_body_parts", map[string]pep3BookletPDFOptionMark{ // 17 手偶身体部位
			"眼": pep3PDFMark(187.3, 151.0, 22.4),
			"耳": pep3PDFMark(224.0, 151.0, 22.7),
			"口": pep3PDFMark(187.3, 166.3, 22.7),
			"鼻": pep3PDFMark(224.0, 166.3, 23.0),
		}),
		pep3PDFRecordMarks(3, 18, "self_body_parts", map[string]pep3BookletPDFOptionMark{ // 18 自己身体部位
			"眼": pep3PDFMark(187.7, 187.7, 21.6),
			"耳": pep3PDFMark(224.0, 187.7, 22.0),
			"口": pep3PDFMark(187.3, 201.3, 22.7),
			"鼻": pep3PDFMark(224.0, 201.3, 22.7),
		}),
		pep3PDFRecordMarks(3, 21, "shape_positions", map[string]pep3BookletPDFOptionMark{ // 21 拼块正确位置
			"三角形": pep3PDFMark(204.7, 279.3, 23.3),
			"圆形":  pep3PDFMark(250.7, 279.3, 23.3),
			"正方形": pep3PDFMark(205.3, 295.0, 23.0),
		}),
		pep3PDFRecordMarks(3, 22, "shape_board_completed", map[string]pep3BookletPDFOptionMark{ // 22 完成形状拼板
			"三角形": pep3PDFMark(205.3, 314.7, 22.7),
			"圆形":  pep3PDFMark(251.3, 314.7, 22.0),
			"正方形": pep3PDFMark(205.3, 330.0, 22.7),
		}),
		pep3PDFRecordMarks(3, 23, "shape_names", map[string]pep3BookletPDFOptionMark{ // 23 说出形状名称
			"三角形": pep3PDFMark(205.3, 349.7, 22.7),
			"圆形":  pep3PDFMark(251.3, 349.7, 22.7),
			"正方形": pep3PDFMark(205.3, 365.3, 23.4),
		}),
		pep3PDFRecordMarks(3, 24, "shape_selection", map[string]pep3BookletPDFOptionMark{ // 24 挑选形状
			"三角形": pep3PDFMark(205.3, 384.7, 22.7),
			"圆形":  pep3PDFMark(251.3, 384.7, 22.0),
			"正方形": pep3PDFMark(206.0, 400.3, 22.0),
		}),
		pep3PDFRecordMarks(3, 25, "object_puzzle_completed", map[string]pep3BookletPDFOptionMark{ // 25 物件拼板
			"小鸡": pep3PDFMark(196.0, 439.3, 23.3),
			"雨伞": pep3PDFMark(242.0, 439.3, 22.7),
			"蝴蝶": pep3PDFMark(196.7, 454.7, 22.6),
			"雪梨": pep3PDFMark(242.7, 454.7, 22.6),
		}),
		pep3PDFRecordMarks(3, 27, "mitten_position_sizes", map[string]pep3BookletPDFOptionMark{ // 27 手套拼块正确位置
			"大": pep3PDFMark(187.3, 513.3, 22.7),
			"中": pep3PDFMark(224.3, 513.0, 21.4),
			"小": pep3PDFMark(188.0, 528.7, 22.7),
		}),

		// 第4页：28-37
		pep3PDFRecordMarks(4, 28, "mitten_completed_sizes", map[string]pep3BookletPDFOptionMark{ // 28 完成手套拼板
			"大": pep3PDFMark(184.0, 86.7, 22.0),
			"中": pep3PDFMark(220.7, 86.7, 22.6),
			"小": pep3PDFMark(184.0, 102.0, 22.7),
		}),
		pep3PDFRecordMarks(4, 29, "size_naming", map[string]pep3BookletPDFOptionMark{ // 29 说出物件大小
			"first_big":    pep3PDFMark(218.0, 121.7, 18.7),
			"first_small":  pep3PDFMark(249.3, 121.7, 18.7),
			"second_big":   pep3PDFMark(218.7, 137.0, 18.6),
			"second_small": pep3PDFMark(249.7, 137.0, 18.6),
		}),
		pep3PDFRecordMarks(4, 30, "size_selection", map[string]pep3BookletPDFOptionMark{ // 30 挑选大小物件
			"first_big":    pep3PDFMark(218.0, 156.7, 18.7),
			"first_small":  pep3PDFMark(249.3, 156.7, 18.7),
			"second_big":   pep3PDFMark(218.7, 172.0, 18.0),
			"second_small": pep3PDFMark(250.0, 172.0, 18.0),
		}),
		pep3PDFRecordRects(4, 31, "cat_puzzle_prompt", map[string]pep3BookletPDFRect{ // 31 完成方式
			// 第31题“自行”的椭圆：pep3PDFCircle(centerX, y, width, height)。
			// centerX 越小越靠左，y 越小越靠上，width/height 控制椭圆大小。
			"自行": pep3PDFCircle(190.0, 202, 24.0, 11.0),
			// 第31题“需示范”的椭圆：当前如果太靠右，就把 232.0 调小；如果太靠下，就把 205.5 调小。
			"需示范": pep3PDFCircle(223.0, 202, 34.0, 11.0),
		}),
		pep3PDFRecordText(4, 31, "completed_piece_count", 227.0, 226.3, 30.0),
		pep3PDFRecordText(4, 32, "interlocked_piece_count", 193.3, 248.0, 22.0),
		pep3PDFRecordRects(4, 33, "cow_puzzle_prompt", map[string]pep3BookletPDFRect{ // 33 完成方式
			"自行":  pep3PDFCircle(190.0, 202, 24.0, 11.0), // 自行
			"需示范": pep3PDFCircle(223.0, 202, 34.0, 11.0), // 需示范
		}),
		pep3PDFRecordText(4, 33, "completed_piece_count", 227.0, 300.3, 33.7),
		pep3PDFRecordMarks(4, 34, "boy_puzzle_parts", map[string]pep3BookletPDFOptionMark{ // 34 男孩拼图部位
			"头":  pep3PDFMark(184.0, 339.3, 30.0),
			"头发": pep3PDFMark(237.3, 339.3, 26.0),
			"双眼": pep3PDFMark(193.3, 354.7, 20.7),
			"鼻":  pep3PDFMark(228.7, 354.7, 34.6),
			"口":  pep3PDFMark(184.0, 370.0, 30.0),
			"身":  pep3PDFMark(228.7, 370.0, 34.6),
			"脚":  pep3PDFMark(184.0, 385.3, 30.0),
		}),
		pep3PDFRecordMarks(4, 35, "sound_objects", map[string]pep3BookletPDFOptionMark{ // 35 发声物
			"响板": pep3PDFMark(192.7, 424.3, 23.0),
			"手铃": pep3PDFMark(239.0, 424.3, 22.3),
			"匙子": pep3PDFMark(193.3, 439.3, 22.7),
		}),
		pep3PDFRecordText(4, 36, "sock", 193.3, 478.7, 22.0),
		pep3PDFRecordText(4, 36, "cup", 239.3, 478.7, 22.0),
		pep3PDFRecordText(4, 36, "toothbrush", 193.3, 494.0, 22.0),
		pep3PDFRecordText(4, 36, "crayon", 239.3, 494.0, 22.0),
		pep3PDFRecordText(4, 36, "scissors", 193.3, 509.3, 22.0),
		pep3PDFRecordText(4, 36, "comb", 239.3, 509.3, 22.0),
		pep3PDFRecordText(4, 36, "pencil", 193.3, 525.0, 22.0),
		pep3PDFRecordMarks(4, 37, "object_use", map[string]pep3BookletPDFOptionMark{ // 37 正确使用物件
			"杯子": pep3PDFMark(192.7, 544.7, 23.3),
			"匙子": pep3PDFMark(238.7, 544.7, 22.6),
			"蜡笔": pep3PDFMark(192.7, 559.7, 23.3),
			"梳子": pep3PDFMark(238.7, 559.7, 22.6),
			"剪刀": pep3PDFMark(192.7, 574.7, 23.3),
		}),

		// 第5页：38-49
		pep3PDFRecordMarks(5, 38, "requested_objects", map[string]pep3BookletPDFOptionMark{ // 38 交出物件
			"杯子": pep3PDFMark(194.0, 82.7, 22.0),
			"匙子": pep3PDFMark(240.0, 82.3, 22.0),
			"蜡笔": pep3PDFMark(194.0, 98.0, 22.0),
			"梳子": pep3PDFMark(240.0, 98.0, 22.0),
			"剪刀": pep3PDFMark(194.0, 113.3, 22.0),
		}),
		pep3PDFRecordMarks(5, 39, "matched_picture_objects", map[string]pep3BookletPDFOptionMark{ // 39 实物配对图片
			"袜子": pep3PDFMark(194.0, 133.0, 22.7),
			"杯子": pep3PDFMark(240.0, 133.0, 23.0),
			"牙刷": pep3PDFMark(194.0, 148.3, 22.7),
			"匙子": pep3PDFMark(240.3, 148.3, 23.0),
			"剪刀": pep3PDFMark(194.0, 163.7, 23.0),
			"梳子": pep3PDFMark(240.3, 163.7, 23.0),
			"铅笔": pep3PDFMark(194.0, 179.3, 23.3),
		}),
		pep3PDFRecordMultilineText(5, 40, "pointed_objects", []pep3BookletPDFTextLine{ // 40 指出物件
			// 第40题第1行：从“指出物件：”后面的横线开始写。X 越小越靠左，Y 越小越靠上，Width 控制本行可写长度。
			{X: 220.0, Y: 198.7, Width: 40.0},
			// 第40题第2行：第一行写不下时自动换到这条横线。
			{X: 178.0, Y: 214.0, Width: 88.0},
		}),
		pep3PDFRecordText(5, 43, "first_attempt", 201.3, 330.7, 19.4),
		pep3PDFRecordText(5, 43, "second_attempt", 250.7, 330.7, 19.3),
		pep3PDFRecordText(5, 43, "third_attempt", 201.3, 346.0, 19.4),
		pep3PDFRecordMarks(5, 44, "tactile_objects", map[string]pep3BookletPDFOptionMark{ // 44 触觉辨别物件
			"球":  pep3PDFMark(187.3, 384.7, 29.4),
			"积木": pep3PDFMark(238.7, 384.7, 23.3),
			"蜡笔": pep3PDFMark(195.7, 400.3, 21.0),
			"硬币": pep3PDFMark(239.0, 400.3, 22.0),
			"匙子": pep3PDFMark(194.7, 416.0, 22.0),
		}),
		pep3PDFRecordRects(5, 45, "material_inspection", map[string]pep3BookletPDFRect{ // 45 短暂/没有
			"短暂": pep3PDFRect(177.0, 424.0, 27.0, 11.0),
			"没有": pep3PDFRect(211.0, 424.0, 24.0, 11.0),
		}),
		pep3PDFRecordRects(5, 46, "visual_inspection", map[string]pep3BookletPDFRect{ // 46 过分兴趣/过分抗拒
			"过分兴趣": pep3PDFRect(177.0, 444.0, 44.0, 11.0),
			"过分抗拒": pep3PDFRect(226.0, 444.0, 44.0, 11.0),
		}),
		pep3PDFRecordRects(5, 49, "body_contact", map[string]pep3BookletPDFRect{ // 49 拒绝/过分抗拒
			"拒绝":   pep3PDFRect(177.0, 502.0, 27.0, 11.0),
			"过分抗拒": pep3PDFRect(216.0, 502.0, 45.0, 11.0),
		}),

		// 第6页：50-63
		pep3PDFRecordRects(6, 50, "tickle_response", map[string]pep3BookletPDFRect{ // 50 拒绝/过分反应
			"拒绝":   pep3PDFRect(177.0, 78.0, 27.0, 11.0),
			"过分反应": pep3PDFRect(216.0, 78.0, 45.0, 11.0),
		}),
		pep3PDFRecordRects(6, 52, "social_communication", map[string]pep3BookletPDFRect{ // 52 被动/完全无反应
			"被动":    pep3PDFRect(177.0, 124.0, 27.0, 11.0),
			"完全无反应": pep3PDFRect(216.0, 124.0, 58.0, 11.0),
		}),
		pep3PDFRecordMarks(6, 54, "gross_motor_imitation", map[string]pep3BookletPDFOptionMark{ // 54 模仿大肌肉动作
			"举手":    pep3PDFMark(200.0, 182.0, 18.0),
			"摸鼻":    pep3PDFMark(242.0, 182.0, 18.0),
			"举手+摸鼻": pep3PDFMark(220.0, 197.3, 18.0),
		}),
		pep3PDFRecordText(6, 58, "first_attempt", 197.3, 352.0, 18.7),
		pep3PDFRecordText(6, 58, "second_attempt", 246.7, 352.0, 18.0),
		pep3PDFRecordText(6, 58, "third_attempt", 198.0, 367.7, 17.3),
		pep3PDFRecordText(6, 59, "first_attempt", 197.3, 387.3, 18.7),
		pep3PDFRecordText(6, 59, "second_attempt", 246.0, 387.3, 18.7),
		pep3PDFRecordText(6, 59, "third_attempt", 197.3, 402.7, 18.7),
		pep3PDFRecordText(6, 60, "first_attempt", 197.3, 422.7, 19.4),
		pep3PDFRecordText(6, 60, "second_attempt", 247.3, 422.7, 18.0),
		pep3PDFRecordText(6, 60, "third_attempt", 197.3, 438.0, 19.4),
		pep3PDFRecordText(6, 61, "kick_ball", 190.3, 457.7, 25.7),
		pep3PDFRecordText(6, 61, "stairs", 199.3, 473.0, 16.7),

		// 第7页：64-77
		pep3PDFRecordRects(7, 64, "string_reaction", map[string]pep3BookletPDFRect{ // 64 无兴趣/怪异反应
			"无兴趣":  pep3PDFRect(177.0, 94.0, 34.0, 11.0),
			"怪异反应": pep3PDFRect(214.0, 96.0, 40.0, 11.0),
		}),
		pep3PDFRecordText(7, 65, "completed_bead_count", 190.7, 126.3, 18.6),
		pep3PDFRecordText(7, 67, "completed_bead_count", 191.0, 185.0, 18.7),
		pep3PDFRecordText(7, 71, "dominant_hand", 207.3, 282.7, 21.4),
		pep3PDFRecordMarks(7, 72, "traced_shapes", map[string]pep3BookletPDFOptionMark{ // 72 沿线描画图形
			"圆形":  pep3PDFMark(191.3, 302.3, 20.7),
			"正方形": pep3PDFMark(246.3, 302.3, 20.7),
			"三角形": pep3PDFMark(191.3, 317.3, 20.7),
			"菱形":  pep3PDFMark(246.3, 317.3, 21.0),
		}),

		// 第8页：78-84
		pep3PDFRecordMarks(8, 84, "pretend_picture_objects", map[string]pep3BookletPDFOptionMark{ // 84 假装使用图画物件
			"哨子": pep3PDFMark(194.0, 240.0, 22.0),
			"球":  pep3PDFMark(232.0, 240.0, 30.0),
			"鼓":  pep3PDFMark(186.7, 255.3, 29.3),
			"钥匙": pep3PDFMark(240.0, 255.3, 22.0),
			"槌子": pep3PDFMark(194.0, 272.7, 22.0),
		}),

		// 第9页：85
		pep3PDFRecordMarks(9, 85, "picture_identification", pep3PDFPictureIdentificationMarks()),

		// 第10页：86-90
		pep3PDFRecordMarks(10, 86, "picture_naming", pep3PDFPictureNamingMarks()),
		pep3PDFRecordText(10, 87, "spoken_phrase", 188.0, 414.7, 67.3),
		pep3PDFRecordMarks(10, 88, "recognized_characters", pep3PDFCharacterMarks(470.0)),
		pep3PDFRecordMarks(10, 89, "read_characters", pep3PDFCharacterMarks(519.7)),
		pep3PDFRecordMarks(10, 90, "matched_characters", pep3PDFCharacterMarks(584.7)),

		// 第11页：91-100
		pep3PDFRecordMarks(11, 92, "read_words", map[string]pep3BookletPDFOptionMark{ // 92 读出单字
			"球": pep3PDFMark(180.0, 123.3, 22.7),
			"狗": pep3PDFMark(217.3, 123.3, 22.7),
			"猫": pep3PDFMark(180.7, 138.0, 22.0),
			"屋": pep3PDFMark(217.3, 138.0, 22.7),
		}),
		pep3PDFRecordMarks(11, 95, "reading_comprehension_questions", map[string]pep3BookletPDFOptionMark{ // 95 阅读理解问题
			"小明有哪些动物呀？":  pep3PDFMark(201.3, 337.3, 23.4),
			"小明在玩什么？":    pep3PDFMark(229.3, 368.3, 22.7),
			"什么跳过小明的皮球？": pep3PDFMark(183.3, 399.3, 23.4),
		}),
		pep3PDFRecordMarks(11, 96, "sentence_commands", map[string]pep3BookletPDFOptionMark{ // 96 阅读句子及遵从指令
			"拿起皮球":   pep3PDFMark(192.0, 439.7, 55.3),
			"皮球放桌子上": pep3PDFMark(210.0, 453.3, 37.3),
		}),
		pep3PDFRecordText(11, 97, "put_in_block_count", 173.3, 494.0, 30.0),
		pep3PDFRecordText(11, 99, "first_attempt", 203.3, 551.0, 33.4),
		pep3PDFRecordText(11, 99, "second_attempt", 203.3, 566.0, 33.4),
		pep3PDFRecordText(11, 99, "third_attempt", 203.3, 581.0, 33.4),

		// 第12页：101-111
		pep3PDFRecordText(12, 101, "two_blocks", 189.3, 102.0, 26.0),
		pep3PDFRecordText(12, 101, "six_blocks", 234.7, 101.7, 25.6),
		pep3PDFRecordText(12, 102, "two_blocks", 189.3, 141.0, 24.4),
		pep3PDFRecordText(12, 102, "seven_blocks", 232.7, 140.7, 25.3),
		pep3PDFRecordText(12, 104, "stack_completed", 211.3, 197.3, 32.7),
		pep3PDFRecordText(12, 104, "jar_completed", 213.0, 213.0, 30.3),
		pep3PDFRecordMarks(12, 105, "matched_colors", pep3PDFColorMarks(251.7)),
		pep3PDFRecordMarks(12, 106, "named_colors", pep3PDFColorMarks(287.0)),
		pep3PDFRecordMarks(12, 107, "selected_colors", pep3PDFColorMarks(322.0)),
		pep3PDFRecordRects(12, 108, "classification_prompt", map[string]pep3BookletPDFRect{ // 108 示范情况
			"自行完成": pep3PDFCircle(198.0, 366.0, 42.0, 11.0), // 自行完成
			"部分示范": pep3PDFCircle(247.0, 366.0, 42.0, 11.0), // 部分示范
			"全部示范": pep3PDFCircle(198.0, 383.0, 42.0, 11.0), // 全部示范
		}),
		pep3PDFRecordRects(12, 108, "classification_basis", map[string]pep3BookletPDFRect{ // 108 分类依据
			"颜色": pep3PDFCircle(189.5, 401.0, 25.0, 11.0), // 颜色
			"形状": pep3PDFCircle(223.5, 401.0, 25.0, 11.0), // 形状
		}),
		pep3PDFRecordText(12, 108, "completed_card_count", 198.0, 422.7, 22.0),
		pep3PDFRecordMarks(12, 111, "imitated_sounds", map[string]pep3BookletPDFOptionMark{ // 111 模仿声音
			"m-m-m": pep3PDFMark(198.0, 500.0, 22.0),
			"ba-ba": pep3PDFMark(206.0, 516.0, 26.0),
			"pa-ta": pep3PDFMark(206.7, 531.3, 26.0),
			"la-la": pep3PDFMark(205.3, 546.7, 28.0),
		}),

		// 第13页：112-122
		pep3PDFRecordText(13, 112, "digits_7_9", 182.7, 87.0, 20.3),
		pep3PDFRecordText(13, 112, "digits_5_3", 220.7, 86.7, 22.0),
		pep3PDFRecordText(13, 113, "digits_2_4_1", 190.0, 106.3, 22.0),
		pep3PDFRecordText(13, 113, "digits_5_7_9", 236.0, 106.3, 22.0),
		pep3PDFRecordText(13, 114, "word_street", 188.7, 126.0, 22.0),
		pep3PDFRecordText(13, 114, "word_car", 234.0, 126.0, 22.0),
		pep3PDFRecordText(13, 114, "word_bye", 188.7, 140.0, 22.0),
		pep3PDFRecordRects(13, 115, "repeated_sentences", map[string]pep3BookletPDFRect{ // 115 正确复述短句
			"bb_looking":    pep3PDFCircle(190.0, 153.0, 45.0, 11.5), // BB 望住
			"want_biscuit":  pep3PDFCircle(196.0, 167.0, 55.0, 11.5), // 又要饼干
			"crying_loudly": pep3PDFCircle(196.0, 181.0, 55.0, 11.5), // 佢大声喊
		}),
		pep3PDFRecordRects(13, 116, "eye_contact", map[string]pep3BookletPDFRect{ // 116 短暂/没有
			"brief": pep3PDFRect(177.0, 196.3, 24.0, 9.5),
			"none":  pep3PDFRect(213.0, 196.3, 24.0, 9.5),
		}),
		pep3PDFRecordRects(13, 117, "delayed_echolalia", map[string]pep3BookletPDFRect{ // 117 不适用/过多
			"not_applicable": pep3PDFRect(177.0, 230.0, 34.0, 9.5),
			"too_much":       pep3PDFRect(220.0, 230.0, 24.0, 9.5),
		}),
		pep3PDFRecordText(13, 119, "pronoun_response", 170.7, 294.7, 40.0),
		pep3PDFRecordText(13, 120, "spoken_words", 169.3, 349.3, 66.7),
		pep3PDFRecordText(13, 121, "spoken_words", 169.3, 403.3, 66.7),
		pep3PDFRecordText(13, 122, "spoken_phrase", 169.3, 473.3, 66.7),

		// 第14页：123-133
		pep3PDFRecordMarks(14, 123, "oral_commands", map[string]pep3BookletPDFOptionMark{ // 123 口语指令
			"拍吓个盒":        pep3PDFMark(217.3, 99.7, 22.0),
			"摸吓只狗":        pep3PDFMark(217.3, 114.0, 22.0),
			"企起身跳":        pep3PDFMark(217.3, 129.7, 22.0),
			"攞个杯给我，然后坐下":  pep3PDFMark(218.0, 160.3, 22.0),
			"敲吓度门，然后摸吓度墙": pep3PDFMark(236.0, 191.3, 22.0),
		}),
		pep3PDFRecordMarks(14, 125, "gesture_responses", map[string]pep3BookletPDFOptionMark{ // 125 对手势反应
			"叫名字+招手":  pep3PDFMark(232.0, 246.3, 29.7),
			"坐下+拿走积木": pep3PDFMark(192.0, 261.0, 14.7),
			"交回颜色笔":   pep3PDFMark(246.0, 261.0, 14.7),
			"其他":      pep3PDFMark(220.0, 292.0, 40.7),
		}),
		pep3PDFRecordMarks(14, 126, "no_stop_commands", map[string]pep3BookletPDFOptionMark{ // 126 不要/停止
			"不要": pep3PDFMark(192.0, 313.3, 29.3),
			"停止": pep3PDFMark(192.0, 327.3, 29.3),
		}),
		pep3PDFRecordText(14, 129, "child_answer", 218.0, 387.7, 36.7),
		pep3PDFRecordText(14, 130, "child_answer", 217.3, 407.3, 36.7),
		pep3PDFRecordMarks(14, 131, "single_actions", map[string]pep3BookletPDFOptionMark{ // 131 动词动作
			"跳":   pep3PDFMark(192.0, 427.0, 29.3),
			"坐下":  pep3PDFMark(228.3, 427.0, 21.7),
			"企起身": pep3PDFMark(192.0, 441.3, 22.0),
		}),
		pep3PDFRecordMarks(14, 133, "wh_questions", map[string]pep3BookletPDFOptionMark{ // 133 问句
			"何人": pep3PDFMark(191.3, 482.0, 22.7),
			"何事": pep3PDFMark(237.3, 482.0, 22.0),
			"何地": pep3PDFMark(192.0, 496.3, 22.0),
			"何时": pep3PDFMark(237.3, 496.3, 22.0),
		}),

		// 第15页：134-151
		pep3PDFRecordMarks(15, 134, "simple_action_commands", map[string]pep3BookletPDFOptionMark{ // 134 简单动作指令
			"坐下":  pep3PDFMark(193.3, 85.3, 22.0),
			"起身":  pep3PDFMark(244.0, 85.0, 17.3),
			"过来":  pep3PDFMark(194.0, 99.0, 20.3),
			"伸我":  pep3PDFMark(244.0, 98.7, 17.3),
			"放低手": pep3PDFMark(202.7, 114.7, 18.0),
			"开门":  pep3PDFMark(244.0, 114.0, 18.0),
			"其他":  pep3PDFMark(202.7, 130.0, 22.0),
		}),
		pep3PDFRecordRects(15, 135, "visual_self_stimulation", pep3PDFSlashRects(139.0, 193.0, 27.0, 226.0, 45.0)),
		pep3PDFRecordRects(15, 136, "space_material_exploration", pep3PDFSlashRects(164.0, 193.0, 34.0, 229.0, 45.0)),
		pep3PDFRecordRects(15, 137, "environment_exploration", pep3PDFSlashRects(204.0, 193.0, 34.0, 229.0, 45.0)),
		pep3PDFRecordRects(15, 138, "sound_response", pep3PDFSlashRects(223.0, 193.0, 34.0, 229.0, 45.0)),
		pep3PDFRecordRects(15, 139, "texture_exploration", pep3PDFSlashRects(242.0, 193.0, 27.0, 226.0, 45.0)),
		pep3PDFRecordRects(15, 140, "taste_use", pep3PDFSlashRects(263.0, 193.0, 27.0, 226.0, 45.0)),
		pep3PDFRecordRects(15, 141, "smell_interest", pep3PDFSlashRects(282.0, 193.0, 27.0, 226.0, 45.0)),
		pep3PDFRecordRects(15, 142, "completed_tests", pep3PDFSlashRects(301.0, 193.0, 27.0, 226.0, 45.0)),
		pep3PDFRecordRects(15, 144, "repeated_sentences_heard", pep3PDFSlashRects(343.0, 193.0, 31.0, 228.0, 28.0)),
		pep3PDFRecordRects(15, 145, "repeated_words_sounds", pep3PDFSlashRects(382.0, 193.0, 31.0, 228.0, 28.0)),
		pep3PDFRecordRects(15, 146, "speech_tone_volume_speed", pep3PDFSlashRects(404.0, 193.0, 31.0, 228.0, 28.0)),
		pep3PDFRecordRects(15, 147, "meaningless_sounds", pep3PDFSlashRects(444.0, 193.0, 31.0, 228.0, 28.0)),
		pep3PDFRecordRects(15, 148, "age_appropriate_vocabulary", pep3PDFSlashRects(464.0, 193.0, 31.0, 228.0, 31.0)),
		pep3PDFRecordRects(15, 149, "self_talk", pep3PDFSlashRects(484.0, 193.0, 31.0, 228.0, 28.0)),
		pep3PDFRecordRects(15, 150, "age_appropriate_articulation", pep3PDFSlashRects(504.0, 193.0, 31.0, 228.0, 45.0)),
		pep3PDFRecordRects(15, 151, "spontaneous_communication", pep3PDFSlashRects(524.0, 193.0, 20.0, 218.0, 45.0)),

		// 第16页：152-172
		pep3PDFRecordRects(16, 154, "cooperation", pep3PDFSlashRects(114.0, 193.0, 27.0, 226.0, 45.0)),
		pep3PDFRecordRects(16, 155, "organization", pep3PDFSlashRects(139.0, 193.0, 27.0, 226.0, 45.0)),
		pep3PDFRecordRects(16, 159, "pleasant_emotion", pep3PDFSlashRects(219.0, 193.0, 27.0, 226.0, 45.0)),
		pep3PDFRecordRects(16, 160, "fear_response", pep3PDFSlashRects(239.0, 193.0, 20.0, 218.0, 28.0)),
		pep3PDFRecordRects(16, 161, "attention", pep3PDFSlashRects(259.0, 193.0, 20.0, 218.0, 45.0)),
		pep3PDFRecordRects(16, 162, "transition", pep3PDFSlashRects(279.0, 193.0, 27.0, 226.0, 45.0)),
		pep3PDFRecordRects(16, 168, "request_help", pep3PDFSlashRects(399.0, 193.0, 20.0, 218.0, 45.0)),
		pep3PDFRecordRects(16, 169, "movement", pep3PDFSlashRects(419.0, 193.0, 27.0, 226.0, 45.0)),
		pep3PDFRecordRects(16, 170, "return_to_examiner", pep3PDFSlashRects(439.0, 193.0, 34.0, 232.0, 28.0)),
		pep3PDFRecordRects(16, 171, "reward_response", pep3PDFSlashRects(459.0, 193.0, 31.0, 228.0, 20.0)),
		pep3PDFRecordRects(16, 172, "social_reward_response", pep3PDFSlashRects(479.0, 193.0, 31.0, 228.0, 20.0)),
	}
}

func pep3PDFRecordText(pageNo, itemNo int, fieldKey string, x, y, width float64) pep3BookletPDFRecordFieldPlacement {
	return pep3BookletPDFRecordFieldPlacement{PageNo: pageNo, ItemNo: itemNo, FieldKey: fieldKey, X: x, Y: y, Width: width, Size: 7.2}
}

func pep3PDFRecordMultilineText(pageNo, itemNo int, fieldKey string, lines []pep3BookletPDFTextLine) pep3BookletPDFRecordFieldPlacement {
	return pep3BookletPDFRecordFieldPlacement{PageNo: pageNo, ItemNo: itemNo, FieldKey: fieldKey, Size: 7.2, TextLines: lines}
}

func pep3PDFRecordMarks(pageNo, itemNo int, fieldKey string, marks map[string]pep3BookletPDFOptionMark) pep3BookletPDFRecordFieldPlacement {
	return pep3BookletPDFRecordFieldPlacement{PageNo: pageNo, ItemNo: itemNo, FieldKey: fieldKey, OptionMarks: marks}
}

func pep3PDFRecordRects(pageNo, itemNo int, fieldKey string, rects map[string]pep3BookletPDFRect) pep3BookletPDFRecordFieldPlacement {
	return pep3BookletPDFRecordFieldPlacement{PageNo: pageNo, ItemNo: itemNo, FieldKey: fieldKey, OptionRects: rects}
}

func pep3PDFMark(x, y, width float64) pep3BookletPDFOptionMark {
	return pep3BookletPDFOptionMark{X: x, Y: y, Width: width, Size: 7.2}
}

func pep3PDFRect(x, y, w, h float64) pep3BookletPDFRect {
	return pep3BookletPDFRect{X: x, Y: y, W: w, H: h}
}

func pep3PDFCircle(centerX, y, w, h float64) pep3BookletPDFRect {
	return pep3BookletPDFRect{X: centerX - w/2, Y: y, W: w, H: h}
}

func pep3PDFSlashRects(y, firstX, firstW, secondX, secondW float64) map[string]pep3BookletPDFRect {
	// 扫描底图里星号在选项左侧，圈选时需要避开星号并贴住文字本身。
	firstX -= 16.0
	secondX -= 12.0
	return map[string]pep3BookletPDFRect{
		"正常":   pep3PDFRect(firstX, y, firstW, 11.0),
		"无兴趣":  pep3PDFRect(firstX, y, firstW, 11.0),
		"无反应":  pep3PDFRect(firstX, y, firstW, 11.0),
		"抗拒":   pep3PDFRect(firstX, y, firstW, 11.0),
		"不适用":  pep3PDFRect(firstX, y, firstW, 11.0),
		"无":    pep3PDFRect(firstX, y, firstW, 11.0),
		"短":    pep3PDFRect(firstX, y, firstW, 11.0),
		"接受":   pep3PDFRect(firstX, y, firstW, 11.0),
		"不察觉":  pep3PDFRect(firstX, y, firstW, 11.0),
		"不一致":  pep3PDFRect(firstX, y, firstW, 11.0),
		"反复":   pep3PDFRect(firstX, y, firstW, 11.0),
		"间中":   pep3PDFRect(firstX, y, firstW, 11.0),
		"无变化":  pep3PDFRect(firstX, y, firstW, 11.0),
		"过分刺激": pep3PDFRect(secondX, y, secondW, 11.0),
		"怪异行为": pep3PDFRect(secondX, y, secondW, 11.0),
		"过度反应": pep3PDFRect(secondX, y, secondW, 11.0),
		"怪异兴趣": pep3PDFRect(secondX, y, secondW, 11.0),
		"过分兴趣": pep3PDFRect(secondX, y, secondW, 11.0),
		"混乱":   pep3PDFRect(firstX, y, firstW, 11.0),
		"过多":   pep3PDFRect(secondX, y, secondW, 11.0),
		"怪异":   pep3PDFRect(secondX, y, secondW, 11.0),
		"不恰当":  pep3PDFRect(secondX, y, secondW, 11.0),
		"无法明白": pep3PDFRect(secondX, y, secondW, 11.0),
		"怪异沉迷": pep3PDFRect(secondX, y, secondW, 11.0),
		"过分反抗": pep3PDFRect(secondX, y, secondW, 11.0),
		"经常混乱": pep3PDFRect(secondX, y, secondW, 11.0),
		"怪异反应": pep3PDFRect(secondX, y, secondW, 11.0),
		"极端反应": pep3PDFRect(secondX, y, secondW, 11.0),
		"过分要求": pep3PDFRect(secondX, y, secondW, 11.0),
		"怪异动作": pep3PDFRect(secondX, y, secondW, 11.0),
		"被动":   pep3PDFRect(secondX, y, secondW, 11.0),
	}
}

func pep3PDFCharacterMarks(topY float64) map[string]pep3BookletPDFOptionMark {
	return map[string]pep3BookletPDFOptionMark{
		"人": pep3PDFMark(184.0, topY, 18.7),
		"口": pep3PDFMark(215.3, topY, 18.7),
		"上": pep3PDFMark(250.0, topY, 18.7),
		"山": pep3PDFMark(184.0, topY+14.3, 18.7),
		"刀": pep3PDFMark(216.0, topY+14.3, 18.7),
		"天": pep3PDFMark(250.0, topY+14.3, 18.7),
		"火": pep3PDFMark(184.0, topY+30.0, 18.7),
		"手": pep3PDFMark(216.0, topY+30.0, 18.7),
		"田": pep3PDFMark(250.0, topY+30.0, 18.7),
	}
}

func pep3PDFColorMarks(topY float64) map[string]pep3BookletPDFOptionMark {
	return map[string]pep3BookletPDFOptionMark{
		"红": pep3PDFMark(183.3, topY, 15.4),
		"黄": pep3PDFMark(212.3, topY, 15.0),
		"蓝": pep3PDFMark(239.7, topY, 14.6),
		"白": pep3PDFMark(184.0, topY+15.3, 14.7),
		"绿": pep3PDFMark(212.7, topY+15.3, 15.3),
	}
}

func pep3PDFPictureIdentificationMarks() map[string]pep3BookletPDFOptionMark {
	return map[string]pep3BookletPDFOptionMark{
		"A 杯子":   pep3PDFMark(187.7, 119.7, 29.3),
		"B 洋娃娃":  pep3PDFMark(187.3, 148.7, 30.0),
		"C 锁匙":   pep3PDFMark(187.3, 177.3, 30.0),
		"D 飞机":   pep3PDFMark(187.7, 206.3, 29.0),
		"E 鸟笼":   pep3PDFMark(187.3, 235.3, 30.0),
		"F 雨伞":   pep3PDFMark(187.3, 264.0, 30.0),
		"G 煮食":   pep3PDFMark(187.3, 293.3, 30.0),
		"H 系鞋带":  pep3PDFMark(187.3, 322.0, 30.0),
		"I 门":    pep3PDFMark(187.3, 350.7, 30.0),
		"J 木偶":   pep3PDFMark(187.3, 380.0, 30.0),
		"K 公鸡":   pep3PDFMark(187.3, 408.7, 30.0),
		"L 接球":   pep3PDFMark(187.7, 437.7, 28.6),
		"M 建筑":   pep3PDFMark(187.3, 466.7, 30.0),
		"N 警察":   pep3PDFMark(188.0, 495.7, 29.3),
		"O 砌积木":  pep3PDFMark(187.3, 524.7, 30.0),
		"P 烧烤":   pep3PDFMark(187.3, 553.3, 30.0),
		"Q 洗澡":   pep3PDFMark(187.3, 582.0, 30.0),
		"R 水壶":   pep3PDFMark(187.3, 610.7, 30.0),
		"S 火车头":  pep3PDFMark(187.3, 639.7, 30.0),
		"T 火箭起飞": pep3PDFMark(187.3, 668.7, 30.0),
	}
}

func pep3PDFPictureNamingMarks() map[string]pep3BookletPDFOptionMark {
	return map[string]pep3BookletPDFOptionMark{
		"A 牛":    pep3PDFMark(222.7, 100.7, 22.0),
		"B 皮球/波": pep3PDFMark(222.7, 116.7, 22.0),
		"C 花":    pep3PDFMark(222.7, 132.0, 22.0),
		"D 婴儿车":  pep3PDFMark(222.7, 147.3, 22.0),
		"E 牙刷":   pep3PDFMark(222.7, 162.7, 22.0),
		"F 雪柜":   pep3PDFMark(222.7, 178.3, 22.0),
		"G 油油":   pep3PDFMark(222.7, 194.0, 22.0),
		"H 打秋千":  pep3PDFMark(222.7, 209.3, 22.0),
		"I 樽":    pep3PDFMark(222.7, 224.7, 22.0),
		"J 电风扇":  pep3PDFMark(222.7, 240.0, 22.0),
		"K 企鹅":   pep3PDFMark(222.7, 255.3, 22.0),
		"L 溜冰":   pep3PDFMark(222.7, 271.0, 21.3),
		"M 抱着狗":  pep3PDFMark(222.7, 286.7, 22.0),
		"N 医生":   pep3PDFMark(222.7, 302.0, 22.0),
		"O 跳水":   pep3PDFMark(222.7, 317.3, 22.6),
		"P 踏车":   pep3PDFMark(246.0, 332.7, 22.7),
		"Q 举起":   pep3PDFMark(222.7, 348.0, 22.0),
		"R 教堂":   pep3PDFMark(222.7, 364.0, 22.6),
		"S 炉":    pep3PDFMark(222.7, 379.3, 22.6),
		"T 整路":   pep3PDFMark(222.7, 394.7, 22.6),
	}
}

func pep3BookletPDFItemPageLayouts() []pep3BookletPDFItemPageLayout {
	// 第4部分：儿童表现记录页。PageNo 是 PDF 逻辑页码，StartItemNo 是该页第一题题号。
	// ItemCenters 按页面上题号顺序排列；需要微调单题得分位置时，改对应页面里的 {X, Y}。
	return []pep3BookletPDFItemPageLayout{
		{
			PageNo:      2,
			StartItemNo: 1,
			ItemCenters: []pep3BookletPDFPoint{
				{383.0, 96.3},
				{383.0, 115.7},
				{383.0, 135.7},
				{382.7, 155.0},
				{515.7, 194.0},
				{435.7, 237.3},
				{382.3, 271.0},
				{303.0, 321.7},
				{435.7, 365.0},
				{435.7, 413.3},
				{382.3, 433.0},
				{435.3, 453.0},
				{435.7, 472.3},
				{435.3, 492.3},
			},
			DomainX:      pep3BookletPDFDomainX(302.3, 328.7, 355.3, 382.0, 408.7, 435.3, 462.0, 489.0, 516.0, 542.3),
			RawSubtotalY: 529.8,
			TallyY:       pep3BookletPDFTallyY(544.5, 559.3, 574.3),
		},
		{
			PageNo:      3,
			StartItemNo: 15,
			ItemCenters: []pep3BookletPDFPoint{
				{438.3, 95.0},
				{438.3, 118.0},
				{358.3, 153.0},
				{358.7, 185.0},
				{332.3, 220.0},
				{491.7, 239.7},
				{306.0, 281.7},
				{385.7, 318.0},
				{332.3, 351.7},
				{359.0, 386.3},
				{306.3, 441.7},
				{412.0, 473.0},
				{306.3, 515.7},
			},
			DomainX:      pep3BookletPDFDomainX(306.0, 332.0, 358.7, 385.3, 412.0, 438.7, 465.3, 491.8, 518.2, 544.7),
			RawSubtotalY: 556.3,
			TallyY:       pep3BookletPDFTallyY(571.0, 586.0, 601.0),
		},
		{
			PageNo:      4,
			StartItemNo: 28,
			ItemCenters: []pep3BookletPDFPoint{
				{382.0, 89.0},
				{328.7, 124.0},
				{355.0, 160.3},
				{302.3, 212.3},
				{382.3, 245.0},
				{302.7, 287.0},
				{303.0, 361.7},
				{302.7, 427.0},
				{329.0, 501.7},
				{435.7, 551.7},
			},
			DomainX:      pep3BookletPDFDomainX(303.3, 329.3, 356.0, 382.7, 409.3, 436.0, 462.7, 489.5, 516.2, 542.5),
			RawSubtotalY: 603.0,
			TallyY:       pep3BookletPDFTallyY(617.7, 632.5, 647.3),
		},
		{
			PageNo:      5,
			StartItemNo: 38,
			ItemCenters: []pep3BookletPDFPoint{
				{356.0, 89.7},
				{303.7, 149.7},
				{356.3, 202.0},
				{383.7, 251.7},
				{304.0, 290.0},
				{304.0, 332.3},
				{304.3, 393.0},
				{516.0, 433.7},
				{516.0, 453.7},
				{410.3, 473.7},
				{410.3, 493.3},
				{463.7, 513.0},
			},
			DomainX:      pep3BookletPDFDomainX(304.8, 331.0, 357.7, 384.5, 411.2, 437.7, 464.2, 490.7, 517.0, 543.3),
			RawSubtotalY: 544.8,
			TallyY:       pep3BookletPDFTallyY(559.7, 574.5, 589.3),
		},
		{
			PageNo:      6,
			StartItemNo: 50,
			ItemCenters: []pep3BookletPDFPoint{
				{460.0, 85.3},
				{486.3, 122.3},
				{486.3, 142.0},
				{406.7, 158.7},
				{433.0, 185.0},
				{406.3, 232.7},
				{406.7, 268.0},
				{513.3, 310.0},
				{406.3, 359.0},
				{406.3, 390.7},
				{406.3, 421.0},
				{406.3, 461.7},
				{406.3, 490.7},
				{406.7, 509.3},
			},
			DomainX:      pep3BookletPDFDomainX(300.0, 326.0, 352.7, 379.3, 406.0, 432.7, 459.7, 486.7, 513.3, 539.7),
			RawSubtotalY: 540.7,
			TallyY:       pep3BookletPDFTallyY(555.5, 570.2, 585.0),
		},
		{
			PageNo:      7,
			StartItemNo: 64,
			ItemCenters: []pep3BookletPDFPoint{
				{510.0, 105.3},
				{377.7, 122.3},
				{404.3, 141.3},
				{377.7, 181.0},
				{378.0, 202.3},
				{404.3, 220.3},
				{378.0, 260.3},
				{404.7, 278.7},
				{378.0, 308.7},
				{299.0, 333.7},
				{299.0, 355.0},
				{299.0, 373.0},
				{299.0, 394.3},
				{299.0, 413.0},
			},
			DomainX:      pep3BookletPDFDomainX(299.3, 325.3, 352.0, 378.7, 405.3, 432.0, 458.7, 485.2, 511.5, 538.5),
			RawSubtotalY: 444.0,
			TallyY:       pep3BookletPDFTallyY(459.0, 474.0, 489.0),
		},
		{
			PageNo:      8,
			StartItemNo: 78,
			ItemCenters: []pep3BookletPDFPoint{
				{382.7, 82.3},
				{303.3, 102.3},
				{303.3, 122.0},
				{303.3, 141.7},
				{383.0, 180.3},
				{303.7, 219.0},
				{303.7, 254.7},
			},
			DomainX:      pep3BookletPDFDomainX(303.7, 329.7, 356.3, 383.0, 409.7, 436.3, 463.0, 489.7, 516.3, 542.7),
			RawSubtotalY: 300.0,
			TallyY:       pep3BookletPDFTallyY(314.7, 329.7, 344.7),
		},
		{
			PageNo:      9,
			StartItemNo: 85,
			ItemCenters: []pep3BookletPDFPoint{
				{352.7, 82.0},
			},
			DomainX:      pep3BookletPDFDomainX(299.7, 326.0, 352.7, 379.3, 406.0, 432.7, 459.0, 485.3, 511.7, 537.7),
			RawSubtotalY: 700.0,
			TallyY:       pep3BookletPDFTallyY(714.7, 729.3, 744.3),
		},
		{
			PageNo:      10,
			StartItemNo: 86,
			ItemCenters: []pep3BookletPDFPoint{
				{327.7, 81.0},
				{329.3, 413.0},
				{356.0, 467.3},
				{329.7, 518.3},
				{304.0, 588.0},
			},
			DomainX:      pep3BookletPDFDomainX(303.7, 330.0, 356.7, 383.2, 409.8, 436.7, 463.3, 490.0, 516.5, 542.8),
			RawSubtotalY: 651.0,
			TallyY:       pep3BookletPDFTallyY(665.8, 680.5, 695.3),
		},
		{
			PageNo:      11,
			StartItemNo: 91,
			ItemCenters: []pep3BookletPDFPoint{
				{325.0, 101.7},
				{325.0, 121.3},
				{325.0, 156.3},
				{325.0, 301.3},
				{325.0, 321.0},
				{325.0, 437.0},
				{378.7, 491.3},
				{484.7, 511.0},
				{378.7, 559.0},
				{351.7, 613.3},
			},
			DomainX:      pep3BookletPDFDomainX(299.0, 325.3, 352.0, 378.7, 405.3, 432.0, 458.5, 484.8, 511.2, 537.5),
			RawSubtotalY: 653.7,
			TallyY:       pep3BookletPDFTallyY(668.7, 683.3, 698.3),
		},
		{
			PageNo:      12,
			StartItemNo: 101,
			ItemCenters: []pep3BookletPDFPoint{
				{354.7, 99.0},
				{328.3, 138.0},
				{328.3, 157.7},
				{302.7, 199.7},
				{302.3, 253.7},
				{328.7, 288.7},
				{355.3, 324.3},
				{303.0, 392.3},
				{303.3, 441.7},
				{303.3, 464.0},
				{303.7, 483.0},
			},
			DomainX:      pep3BookletPDFDomainX(303.7, 330.0, 356.7, 383.3, 410.0, 436.7, 463.3, 490.0, 516.7, 543.0),
			RawSubtotalY: 574.8,
			TallyY:       pep3BookletPDFTallyY(589.5, 604.2, 619.0),
		},
		{
			PageNo:      13,
			StartItemNo: 112,
			ItemCenters: []pep3BookletPDFPoint{
				{298.3, 84.3},
				{298.3, 103.7},
				{298.7, 123.7},
				{298.3, 167.7},
				{484.0, 202.0},
				{537.0, 235.7},
				{324.7, 274.3},
				{351.0, 293.7},
				{324.7, 332.7},
				{325.0, 387.0},
				{325.0, 456.7},
			},
			DomainX:      pep3BookletPDFDomainX(298.7, 324.7, 351.3, 378.0, 404.7, 431.3, 458.0, 484.3, 510.7, 537.0),
			RawSubtotalY: 504.2,
			TallyY:       pep3BookletPDFTallyY(518.8, 533.5, 548.5),
		},
		{
			PageNo:      14,
			StartItemNo: 123,
			ItemCenters: []pep3BookletPDFPoint{
				{354.0, 97.7},
				{381.3, 224.7},
				{354.3, 244.3},
				{354.3, 310.3},
				{354.3, 345.7},
				{354.3, 365.7},
				{328.0, 385.0},
				{328.0, 405.0},
				{354.7, 424.7},
				{328.3, 460.0},
				{354.7, 479.7},
			},
			DomainX:      pep3BookletPDFDomainX(301.7, 328.0, 354.7, 381.3, 408.0, 434.7, 461.3, 488.0, 514.7, 541.3),
			RawSubtotalY: 524.2,
			TallyY:       pep3BookletPDFTallyY(539.0, 553.8, 568.7),
		},
		{
			PageNo:      15,
			StartItemNo: 134,
			ItemCenters: []pep3BookletPDFPoint{
				{303.3, 82.3},
				{514.7, 147.7},
				{514.7, 167.0},
				{515.0, 202.3},
				{515.0, 222.0},
				{515.0, 241.7},
				{515.0, 261.7},
				{515.0, 281.3},
				{515.3, 301.0},
				{330.3, 321.3},
				{542.3, 340.3},
				{542.3, 375.7},
				{542.7, 410.7},
				{543.0, 445.7},
				{543.0, 465.7},
				{543.0, 485.7},
				{543.0, 505.0},
				{543.3, 525.3},
			},
			DomainX:      pep3BookletPDFDomainX(305.0, 331.3, 358.0, 384.7, 411.3, 438.0, 464.5, 490.8, 517.0, 543.3),
			RawSubtotalY: 557.2,
			TallyY:       pep3BookletPDFTallyY(571.8, 586.7, 601.7),
		},
		{
			PageNo:      16,
			StartItemNo: 152,
			ItemCenters: []pep3BookletPDFPoint{
				{535.3, 84.7},
				{535.7, 104.3},
				{482.0, 124.0},
				{322.3, 143.7},
				{455.0, 163.3},
				{455.0, 183.0},
				{455.0, 203.0},
				{455.0, 222.3},
				{455.0, 242.3},
				{455.0, 262.0},
				{508.3, 281.7},
				{455.0, 301.7},
				{455.0, 321.0},
				{455.0, 341.0},
				{482.0, 360.7},
				{482.0, 380.3},
				{482.0, 400.0},
				{508.3, 419.7},
				{482.0, 439.7},
				{482.0, 459.0},
				{482.3, 478.0},
			},
			DomainX:      pep3BookletPDFDomainX(296.3, 322.7, 349.3, 376.0, 402.7, 429.3, 456.0, 482.7, 509.3, 535.7),
			RawSubtotalY: 512.3,
			TallyY:       pep3BookletPDFTallyY(527.0, 541.8, 556.5),
		},
	}
}

func pep3BookletPDFDomainX(cvp, el, rl, fm, gm, vmi, ae, sr, cmb, cvb float64) map[string]float64 {
	return map[string]float64{
		"CVP": cvp,
		"EL":  el,
		"RL":  rl,
		"FM":  fm,
		"GM":  gm,
		"VMI": vmi,
		"AE":  ae,
		"SR":  sr,
		"CMB": cmb,
		"CVB": cvb,
	}
}

func pep3BookletPDFTallyY(score2, score1, score0 float64) map[int]float64 {
	return map[int]float64{
		2: score2,
		1: score1,
		0: score0,
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

func (r pep3BookletPDFRenderer) multilineText(lines []pep3BookletPDFTextLine, size float64, value string) {
	value = cleanPEP3BookletPDFValue(value)
	if value == "" || len(lines) == 0 {
		return
	}
	_ = r.pdf.SetFont(pep3BookletPDFFontFamily, "", size)
	remaining := value
	for _, line := range lines {
		if strings.TrimSpace(remaining) == "" {
			return
		}
		text, rest := r.consumeTextLine(remaining, line.Width)
		if strings.TrimSpace(text) == "" {
			continue
		}
		r.pdf.SetXY(line.X, line.Y-pep3BookletPDFLineBaselineGap)
		_ = r.pdf.Text(text)
		remaining = rest
	}
}

func (r pep3BookletPDFRenderer) consumeTextLine(value string, width float64) (string, string) {
	value = strings.TrimSpace(value)
	if value == "" || width <= 0 {
		return value, ""
	}
	runes := []rune(value)
	current := ""
	for index, char := range runes {
		candidate := current + string(char)
		textWidth, err := r.pdf.MeasureTextWidth(candidate)
		if err == nil && textWidth > width && current != "" {
			if pep3BookletPDFIsLineSeparator(char) {
				return current + string(char), strings.TrimSpace(string(runes[index+1:]))
			}
			return current, strings.TrimSpace(string(runes[index:]))
		}
		current = candidate
	}
	return current, ""
}

func pep3BookletPDFIsLineSeparator(char rune) bool {
	switch char {
	case '、', '，', ',', '；', ';':
		return true
	default:
		return false
	}
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

func (r pep3BookletPDFRenderer) centerBold(x, y, width, size float64, value string) {
	value = cleanPEP3BookletPDFValue(value)
	if value == "" {
		return
	}
	_ = r.pdf.SetFont(pep3BookletPDFFontFamily, "", size)
	textWidth, err := r.pdf.MeasureTextWidth(value)
	if err != nil || textWidth > width {
		textWidth = 0
	}
	baseX := x + (width-textWidth)/2
	baseY := y - pep3BookletPDFLineBaselineGap
	for _, offsetX := range []float64{0, 0.35, -0.35} {
		r.pdf.SetXY(baseX+offsetX, baseY)
		_ = r.pdf.Text(value)
	}
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
