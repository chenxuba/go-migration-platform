package service

import (
	"context"
	"embed"
	"errors"
	"fmt"
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

	// 第19页：发展表现图七个副测验栏的中心 x 坐标，同时用于折线点和底部两个计数框。
	// 如果折线或底部数字左右不齐，只调这里。
	xByScale := map[string]float64{
		"CVP": 131.2, // CVP 列
		"EL":  188.1, // EL 列
		"RL":  244.4, // RL 列
		"FM":  301.1, // FM 列
		"GM":  357.3, // GM 列
		"VMI": 411.2, // VMI 列
		"PSC": 461.5, // PSC 列
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
		x := xByScale[scaleCode]
		profileScore := profileScores[scaleCode]
		if !profileScore.HasBreakdown {
			continue
		}
		r.centerInBox(x, passedScoreY, 24, 8, strconv.Itoa(profileScore.PassedScore))   // 通过项目分数：2分项目贡献的分数总和
		r.centerInBox(x, partialScoreY, 24, 8, strconv.Itoa(profileScore.PartialScore)) // 部份通过项目分数：1分项目贡献的分数总和
	}

	r.drawDevelopmentProfileLine(profileScores, normRecords, order, xByScale)
}

func (r pep3BookletPDFRenderer) drawDevelopmentProfileLine(profileScores map[string]pep3BookletPDFProfileScore, normRecords []pep3score.NormRecord, order []string, xByScale map[string]float64) {
	// 第19页：发展表现图纵轴。topY 对应92个月这一行，rowGap 是相邻月龄行距。
	// 折线整体上下偏差时调 topY；行距不贴合月份网格时调 rowGap。
	const (
		topY   = 81.6
		rowGap = 7.82
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
		point := graphPoint{
			x: xByScale[scaleCode],
			y: pep3BookletPDFDevelopmentProfileY(months, lessThan, greaterThan, topY, rowGap),
		}
		if hasLastPoint {
			r.pdf.Line(lastPoint.x, lastPoint.y, point.x, point.y)
		}
		points = append(points, point)
		lastPoint = point
		hasLastPoint = true
	}
	if len(points) == 0 {
		return
	}
	for _, point := range points {
		// 第19页：折线点位椭圆。正版样式是用横向空心椭圆圈住该列数值/范围，不填充。
		// 椭圆过宽/过窄调 markerRadiusX，过高/过矮调 markerRadiusY。
		const (
			markerRadiusX = 7.0
			markerRadiusY = 4.0
		)
		r.pdf.Oval(point.x-markerRadiusX, point.y-markerRadiusY, point.x+markerRadiusX, point.y+markerRadiusY)
	}
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
