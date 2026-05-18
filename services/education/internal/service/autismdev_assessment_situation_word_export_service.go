package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/pkg/autismdevscore"
	"go-migration-platform/services/education/internal/model"
)

const (
	autismDevAssessmentSituationTitle    = "3.1 发展能力计分汇总表"
	autismDevAssessmentSituationGridUnit = 144
	autismDevAssessmentSituationGridCols = 70
)

type autismDevAssessmentSituationWordExport struct {
	Title              string
	StudentName        string
	Age                string
	BirthDate          string
	AssessmentDate     string
	ExaminerName       string
	AssessmentSequence string
	DevelopmentRows    []autismDevAssessmentSituationDevelopmentRow
	BehaviorRows       []autismDevAssessmentSituationBehaviorRow
}

type autismDevAssessmentSituationDevelopmentRow struct {
	Label        string
	Measured     bool
	PCount       int
	SupportCount int
	TotalScore   int
}

type autismDevAssessmentSituationBehaviorRow struct {
	Label    string
	Measured bool
	ACount   int
	MCount   int
	SCount   int
}

type autismDevAssessmentSituationDevelopmentSpec struct {
	Code  string
	Label string
}

type autismDevAssessmentSituationBehaviorSpec struct {
	Label       string
	StartItemNo int
	EndItemNo   int
}

var autismDevAssessmentSituationDevelopmentOrder = []autismDevAssessmentSituationDevelopmentSpec{
	{Code: autismdevscore.DomainLanguageComm, Label: "语言与沟通能力"},
	{Code: autismdevscore.DomainCognition, Label: "认知能力"},
	{Code: autismdevscore.DomainDailyLiving, Label: "自理能力"},
	{Code: autismdevscore.DomainSensory, Label: "感知觉能力"},
	{Code: autismdevscore.DomainGrossMotor, Label: "粗大动作能力"},
	{Code: autismdevscore.DomainFineMotor, Label: "精细动作能力"},
	{Code: autismdevscore.DomainSocial, Label: "社会交往能力"},
}

var autismDevAssessmentSituationBehaviorSpecs = []autismDevAssessmentSituationBehaviorSpec{
	{Label: "依附情绪行为", StartItemNo: 442, EndItemNo: 443},
	{Label: "情绪理解", StartItemNo: 444, EndItemNo: 447},
	{Label: "情绪表达与调节", StartItemNo: 448, EndItemNo: 455},
	{Label: "关系与情感", StartItemNo: 456, EndItemNo: 466},
	{Label: "对物品的兴趣", StartItemNo: 467, EndItemNo: 475},
	{Label: "感觉偏好", StartItemNo: 476, EndItemNo: 485},
	{Label: "特殊行为", StartItemNo: 486, EndItemNo: 493},
}

func (svc *Service) ExportAutismDevAssessmentSituationWord(userID int64, recordID int64) (string, string, []byte, error) {
	_, record, score, data, itemScores, err := svc.autismDevResultAnalysisContext(userID, recordID)
	if err != nil {
		return "", "", nil, err
	}
	export := buildAutismDevAssessmentSituationWordExport(record, score, data, itemScores)
	content, err := buildAutismDevAssessmentSituationWordDocx(export)
	if err != nil {
		return "", "", nil, err
	}
	fileName := fmt.Sprintf("%s-孤独症儿童评估情况-%s.docx", sanitizeExportFileName(export.StudentName), time.Now().Format("20060102150405"))
	return fileName, iepPlanWordContentType, content, nil
}

func (svc *Service) ExportAutismDevAssessmentSituationPDF(userID int64, recordID int64) (string, string, []byte, error) {
	fileName, content, err := svc.buildAutismDevAssessmentSituationPDF(userID, recordID)
	if err != nil {
		return "", "", nil, err
	}
	return fileName, iepPlanPDFContentType, content, nil
}

func buildAutismDevAssessmentSituationWordExport(record model.AssessmentRecordDetailVO, score autismdevscore.AssessmentResult, data autismDevStaticData, itemScores map[int]string) autismDevAssessmentSituationWordExport {
	return autismDevAssessmentSituationWordExport{
		Title:              autismDevAssessmentSituationTitle,
		StudentName:        strings.TrimSpace(record.StudentName),
		Age:                formatIEPPlanAge(record.AgeYears, record.AgeMonths, record.AgeDays),
		BirthDate:          formatIEPPlanDate(record.BirthDate),
		AssessmentDate:     formatIEPPlanDate(record.AssessmentDate),
		ExaminerName:       strings.TrimSpace(record.ExaminerName),
		AssessmentSequence: formatAutismDevAssessmentSituationSequence(record.AssessmentSequence),
		DevelopmentRows:    autismDevAssessmentSituationDevelopmentRows(record, score, data),
		BehaviorRows:       autismDevAssessmentSituationBehaviorRows(record, score, data, itemScores),
	}
}

func autismDevAssessmentSituationDevelopmentRows(record model.AssessmentRecordDetailVO, score autismdevscore.AssessmentResult, data autismDevStaticData) []autismDevAssessmentSituationDevelopmentRow {
	domainByCode := autismDevAssessmentSituationDomainsByCode(score)
	scopeSet, hasScope := autismDevAssessmentSituationScopeSet(record, data)
	rows := make([]autismDevAssessmentSituationDevelopmentRow, 0, len(autismDevAssessmentSituationDevelopmentOrder))
	for _, spec := range autismDevAssessmentSituationDevelopmentOrder {
		domain, ok := domainByCode[spec.Code]
		if !ok || !strings.EqualFold(strings.TrimSpace(domain.ScoreType), autismdevscore.ScoreTypePEF) {
			continue
		}
		if hasScope {
			if !scopeSet[spec.Code] {
				continue
			}
		} else if domain.AnsweredItemCount <= 0 {
			continue
		}
		rows = append(rows, autismDevAssessmentSituationDevelopmentRow{
			Label:        spec.Label,
			Measured:     domain.AnsweredItemCount > 0,
			PCount:       domain.PCount,
			SupportCount: domain.ECount + domain.FCount + domain.XCount,
			TotalScore:   firstNonZeroInt(domain.RawScore, domain.PCount),
		})
	}
	if len(rows) > 0 {
		return rows
	}
	for _, spec := range autismDevAssessmentSituationDevelopmentOrder {
		domain, ok := domainByCode[spec.Code]
		if !ok || !strings.EqualFold(strings.TrimSpace(domain.ScoreType), autismdevscore.ScoreTypePEF) {
			continue
		}
		rows = append(rows, autismDevAssessmentSituationDevelopmentRow{
			Label:        spec.Label,
			Measured:     domain.AnsweredItemCount > 0,
			PCount:       domain.PCount,
			SupportCount: domain.ECount + domain.FCount + domain.XCount,
			TotalScore:   firstNonZeroInt(domain.RawScore, domain.PCount),
		})
	}
	return rows
}

func autismDevAssessmentSituationBehaviorRows(record model.AssessmentRecordDetailVO, score autismdevscore.AssessmentResult, data autismDevStaticData, itemScores map[int]string) []autismDevAssessmentSituationBehaviorRow {
	domainByCode := autismDevAssessmentSituationDomainsByCode(score)
	behaviorDomain, ok := domainByCode[autismdevscore.DomainEmotionBehavior]
	if !ok || !strings.EqualFold(strings.TrimSpace(behaviorDomain.ScoreType), autismdevscore.ScoreTypeAMS) {
		return nil
	}
	scopeSet, hasScope := autismDevAssessmentSituationScopeSet(record, data)
	if hasScope {
		if !scopeSet[autismdevscore.DomainEmotionBehavior] {
			return nil
		}
	} else if behaviorDomain.AnsweredItemCount <= 0 {
		return nil
	}
	rows := make([]autismDevAssessmentSituationBehaviorRow, 0, len(autismDevAssessmentSituationBehaviorSpecs))
	for _, spec := range autismDevAssessmentSituationBehaviorSpecs {
		row := autismDevAssessmentSituationBehaviorScoreRow(spec, itemScores)
		rows = append(rows, row)
	}
	return rows
}

func autismDevAssessmentSituationBehaviorScoreRow(spec autismDevAssessmentSituationBehaviorSpec, itemScores map[int]string) autismDevAssessmentSituationBehaviorRow {
	row := autismDevAssessmentSituationBehaviorRow{Label: spec.Label}
	for itemNo := spec.StartItemNo; itemNo <= spec.EndItemNo; itemNo++ {
		switch normalizeAutismDevScore(itemScores[itemNo]) {
		case autismdevscore.ScoreA:
			row.ACount++
			row.Measured = true
		case autismdevscore.ScoreM:
			row.MCount++
			row.Measured = true
		case autismdevscore.ScoreS:
			row.SCount++
			row.Measured = true
		}
	}
	return row
}

func autismDevAssessmentSituationDomainsByCode(score autismdevscore.AssessmentResult) map[string]autismdevscore.DomainResult {
	out := make(map[string]autismdevscore.DomainResult, len(score.Domains))
	for _, domain := range score.Domains {
		code := strings.TrimSpace(domain.DomainCode)
		if code != "" {
			out[code] = domain
		}
	}
	return out
}

func autismDevAssessmentSituationScopeSet(record model.AssessmentRecordDetailVO, data autismDevStaticData) (map[string]bool, bool) {
	scope := autismDevTrainingCurrentRecordDomains(record, data)
	out := make(map[string]bool, len(scope))
	for _, code := range scope {
		code = strings.TrimSpace(code)
		if code != "" {
			out[code] = true
		}
	}
	return out, len(out) > 0
}

func buildAutismDevAssessmentSituationWordDocx(export autismDevAssessmentSituationWordExport) ([]byte, error) {
	if len(export.DevelopmentRows) == 0 && len(export.BehaviorRows) == 0 {
		return nil, errors.New("暂无可导出的评估情况")
	}
	entries := map[string][]byte{
		"[Content_Types].xml":          []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/></Types>`),
		"_rels/.rels":                  []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>`),
		"word/_rels/document.xml.rels": []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"></Relationships>`),
		"word/document.xml":            []byte(buildAutismDevAssessmentSituationDocumentXML(export)),
	}
	return writeDocxZipEntries(entries)
}

func buildAutismDevAssessmentSituationDocumentXML(export autismDevAssessmentSituationWordExport) string {
	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	builder.WriteString(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">`)
	builder.WriteString(`<w:body>`)
	builder.WriteString(buildAutismDevAssessmentSituationTitleParagraph(firstNonEmptyExportValue(export.Title, autismDevAssessmentSituationTitle)))
	builder.WriteString(buildAutismDevAssessmentSituationTable(export))
	builder.WriteString(`<w:sectPr><w:pgSz w:w="11906" w:h="16838"/><w:pgMar w:top="720" w:right="720" w:bottom="720" w:left="720" w:header="851" w:footer="720" w:gutter="0"/></w:sectPr>`)
	builder.WriteString(`</w:body></w:document>`)
	return builder.String()
}

func buildAutismDevAssessmentSituationTitleParagraph(text string) string {
	return buildIEPParagraph(text, "center", true, 32, iepPlanWordCellOptions{
		CompactParagraph: true,
		SpacingAfter:     140,
		LineSpacing:      260,
	})
}

func buildAutismDevAssessmentSituationTable(export autismDevAssessmentSituationWordExport) string {
	var builder strings.Builder
	builder.WriteString(buildIEPTableStart(autismDevAssessmentSituationTableGrid()))
	builder.WriteString(buildAutismDevAssessmentSituationInfoRows(export))
	builder.WriteString(buildAutismDevAssessmentSituationHeaderRows())
	for _, row := range export.DevelopmentRows {
		builder.WriteString(buildAutismDevAssessmentSituationDevelopmentRow(row))
	}
	if len(export.BehaviorRows) > 0 {
		builder.WriteString(buildAutismDevAssessmentSituationBehaviorHeaderRow())
		for index, row := range export.BehaviorRows {
			builder.WriteString(buildAutismDevAssessmentSituationBehaviorRow(index, row))
		}
	}
	builder.WriteString(`</w:tbl>`)
	return builder.String()
}

func buildAutismDevAssessmentSituationInfoRows(export autismDevAssessmentSituationWordExport) string {
	var builder strings.Builder
	builder.WriteString(buildIEPTableRowStart(340))
	builder.WriteString(autismDevAssessmentSituationCell([]string{"儿童姓名"}, 7, autismDevAssessmentSituationInfoLabelCellOptions()))
	builder.WriteString(autismDevAssessmentSituationCell([]string{autismDevAssessmentSituationValue(export.StudentName)}, 16, autismDevAssessmentSituationInfoValueCellOptions()))
	builder.WriteString(autismDevAssessmentSituationCell([]string{"测评年龄"}, 7, autismDevAssessmentSituationInfoLabelCellOptions()))
	builder.WriteString(autismDevAssessmentSituationCell([]string{autismDevAssessmentSituationValue(export.Age)}, 16, autismDevAssessmentSituationInfoValueCellOptions()))
	builder.WriteString(autismDevAssessmentSituationCell([]string{"测评日期"}, 7, autismDevAssessmentSituationInfoLabelCellOptions()))
	builder.WriteString(autismDevAssessmentSituationCell([]string{autismDevAssessmentSituationValue(export.AssessmentDate)}, 17, autismDevAssessmentSituationInfoValueCellOptions()))
	builder.WriteString(`</w:tr>`)
	builder.WriteString(buildIEPTableRowStart(340))
	builder.WriteString(autismDevAssessmentSituationCell([]string{"评估者"}, 7, autismDevAssessmentSituationInfoLabelCellOptions()))
	builder.WriteString(autismDevAssessmentSituationCell([]string{autismDevAssessmentSituationValue(export.ExaminerName)}, 16, autismDevAssessmentSituationInfoValueCellOptions()))
	builder.WriteString(autismDevAssessmentSituationCell([]string{"出生日期"}, 7, autismDevAssessmentSituationInfoLabelCellOptions()))
	builder.WriteString(autismDevAssessmentSituationCell([]string{autismDevAssessmentSituationValue(export.BirthDate)}, 16, autismDevAssessmentSituationInfoValueCellOptions()))
	builder.WriteString(autismDevAssessmentSituationCell([]string{"测评次数"}, 7, autismDevAssessmentSituationInfoLabelCellOptions()))
	builder.WriteString(autismDevAssessmentSituationCell([]string{autismDevAssessmentSituationValue(export.AssessmentSequence)}, 17, autismDevAssessmentSituationInfoValueCellOptions()))
	builder.WriteString(`</w:tr>`)
	return builder.String()
}

func buildAutismDevAssessmentSituationHeaderRows() string {
	var builder strings.Builder
	builder.WriteString(buildIEPTableRowStart(380))
	builder.WriteString(autismDevAssessmentSituationCell([]string{"领   域"}, 18, iepPlanWordCellOptions{VMerge: "restart", Align: "center", VAlign: "center", CompactParagraph: true, LineSpacing: 220, FontSize: 20}))
	builder.WriteString(autismDevAssessmentSituationCell([]string{"评估结果"}, 39, iepPlanWordCellOptions{Align: "center", VAlign: "center", CompactParagraph: true, LineSpacing: 220, FontSize: 20}))
	builder.WriteString(autismDevAssessmentSituationCell([]string{"备注"}, 13, iepPlanWordCellOptions{VMerge: "restart", Align: "center", VAlign: "center", CompactParagraph: true, LineSpacing: 220, FontSize: 20}))
	builder.WriteString(`</w:tr>`)
	builder.WriteString(buildIEPTableRowStart(360))
	builder.WriteString(autismDevAssessmentSituationCell(nil, 18, iepPlanWordCellOptions{VMerge: "continue"}))
	builder.WriteString(autismDevAssessmentSituationCell([]string{"P"}, 13, autismDevAssessmentSituationBodyCenterCellOptions()))
	builder.WriteString(autismDevAssessmentSituationCell([]string{"E+F(X)"}, 13, autismDevAssessmentSituationBodyCenterCellOptions()))
	builder.WriteString(autismDevAssessmentSituationCell([]string{"总分"}, 13, autismDevAssessmentSituationBodyCenterCellOptions()))
	builder.WriteString(autismDevAssessmentSituationCell(nil, 13, iepPlanWordCellOptions{VMerge: "continue"}))
	builder.WriteString(`</w:tr>`)
	return builder.String()
}

func buildAutismDevAssessmentSituationDevelopmentRow(row autismDevAssessmentSituationDevelopmentRow) string {
	var builder strings.Builder
	builder.WriteString(buildIEPTableRowStart(420))
	builder.WriteString(autismDevAssessmentSituationCell([]string{row.Label}, 18, autismDevAssessmentSituationBodyCenterCellOptions()))
	builder.WriteString(autismDevAssessmentSituationCell([]string{autismDevAssessmentSituationScoreText(row.Measured, row.PCount)}, 13, autismDevAssessmentSituationBodyCenterCellOptions()))
	builder.WriteString(autismDevAssessmentSituationCell([]string{autismDevAssessmentSituationScoreText(row.Measured, row.SupportCount)}, 13, autismDevAssessmentSituationBodyCenterCellOptions()))
	builder.WriteString(autismDevAssessmentSituationCell([]string{autismDevAssessmentSituationScoreText(row.Measured, row.TotalScore)}, 13, autismDevAssessmentSituationBodyCenterCellOptions()))
	builder.WriteString(autismDevAssessmentSituationCell(nil, 13, autismDevAssessmentSituationBodyCenterCellOptions()))
	builder.WriteString(`</w:tr>`)
	return builder.String()
}

func buildAutismDevAssessmentSituationBehaviorHeaderRow() string {
	var builder strings.Builder
	builder.WriteString(buildIEPTableRowStart(340))
	builder.WriteString(autismDevAssessmentSituationCell([]string{"情绪与行为能力"}, 18, autismDevAssessmentSituationBodyCenterCellOptions()))
	builder.WriteString(autismDevAssessmentSituationCell([]string{"A"}, 13, autismDevAssessmentSituationBodyCenterCellOptions()))
	builder.WriteString(autismDevAssessmentSituationCell([]string{"M"}, 13, autismDevAssessmentSituationBodyCenterCellOptions()))
	builder.WriteString(autismDevAssessmentSituationCell([]string{"S"}, 13, autismDevAssessmentSituationBodyCenterCellOptions()))
	builder.WriteString(autismDevAssessmentSituationCell(nil, 13, autismDevAssessmentSituationBodyCenterCellOptions()))
	builder.WriteString(`</w:tr>`)
	return builder.String()
}

func buildAutismDevAssessmentSituationBehaviorRow(index int, row autismDevAssessmentSituationBehaviorRow) string {
	var builder strings.Builder
	builder.WriteString(buildIEPTableRowStart(400))
	builder.WriteString(autismDevAssessmentSituationCell([]string{fmt.Sprintf("%d、%s", index+1, row.Label)}, 18, autismDevAssessmentSituationBodyCenterCellOptions()))
	builder.WriteString(autismDevAssessmentSituationCell([]string{autismDevAssessmentSituationScoreText(row.Measured, row.ACount)}, 13, autismDevAssessmentSituationBodyCenterCellOptions()))
	builder.WriteString(autismDevAssessmentSituationCell([]string{autismDevAssessmentSituationScoreText(row.Measured, row.MCount)}, 13, autismDevAssessmentSituationBodyCenterCellOptions()))
	builder.WriteString(autismDevAssessmentSituationCell([]string{autismDevAssessmentSituationScoreText(row.Measured, row.SCount)}, 13, autismDevAssessmentSituationBodyCenterCellOptions()))
	builder.WriteString(autismDevAssessmentSituationCell(nil, 13, autismDevAssessmentSituationBodyCenterCellOptions()))
	builder.WriteString(`</w:tr>`)
	return builder.String()
}

func autismDevAssessmentSituationCell(lines []string, span int, options iepPlanWordCellOptions) string {
	if span <= 0 {
		span = 1
	}
	options.GridSpan = span
	return buildIEPCell(lines, span*autismDevAssessmentSituationGridUnit, options)
}

func autismDevAssessmentSituationTableGrid() []int {
	widths := make([]int, autismDevAssessmentSituationGridCols)
	for index := range widths {
		widths[index] = autismDevAssessmentSituationGridUnit
	}
	return widths
}

func autismDevAssessmentSituationInfoLabelCellOptions() iepPlanWordCellOptions {
	return iepPlanWordCellOptions{
		Align:            "center",
		VAlign:           "center",
		CompactParagraph: true,
		LineSpacing:      210,
		FontSize:         20,
	}
}

func autismDevAssessmentSituationInfoValueCellOptions() iepPlanWordCellOptions {
	return iepPlanWordCellOptions{
		Align:            "center",
		VAlign:           "center",
		CompactParagraph: true,
		LineSpacing:      210,
		FontSize:         20,
	}
}

func autismDevAssessmentSituationBodyCenterCellOptions() iepPlanWordCellOptions {
	return iepPlanWordCellOptions{
		Align:            "center",
		VAlign:           "center",
		CompactParagraph: true,
		LineSpacing:      210,
		FontSize:         20,
	}
}

func autismDevAssessmentSituationValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "-"
	}
	return value
}

func autismDevAssessmentSituationScoreText(measured bool, value int) string {
	if !measured {
		return ""
	}
	return strconv.Itoa(value)
}

func formatAutismDevAssessmentSituationSequence(value int) string {
	if value <= 0 {
		return "-"
	}
	return fmt.Sprintf("第%d次", value)
}

func firstNonZeroInt(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}
