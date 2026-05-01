package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"go-migration-platform/pkg/pep3score"
	"go-migration-platform/services/education/internal/model"
)

const pep3RecordBookletPDF = "测试员记录册彩(1).pdf"

type pep3SavedInputSnapshot struct {
	ItemScores          map[int]int                          `json:"itemScores"`
	ItemScoreList       []pep3SavedItemScore                 `json:"itemScoreList"`
	RawScores           map[string]int                       `json:"rawScores"`
	RawScoreList        []pep3SavedRawScore                  `json:"rawScoreList"`
	ItemRecordValues    map[int]map[string]any               `json:"itemRecordValues"`
	ItemRecordValueList []pep3SavedItemRecordValueRequest    `json:"itemRecordValueList"`
	AllowMissingItems   bool                                 `json:"allowMissingItems"`
	CaregiverReport     *model.PEP3CaregiverReportSubmission `json:"caregiverReport,omitempty"`
}

type pep3SavedItemScore struct {
	ItemNo int `json:"itemNo"`
	Score  int `json:"score"`
}

type pep3SavedRawScore struct {
	ScaleCode string `json:"scaleCode"`
	RawScore  int    `json:"rawScore"`
}

type pep3SavedItemRecordValueRequest struct {
	ItemNo   int    `json:"itemNo"`
	FieldKey string `json:"fieldKey"`
	Value    any    `json:"value"`
}

type pep3BookletItemDefinition struct {
	ItemNo      int    `json:"item_no"`
	ItemTitle   string `json:"item_title"`
	TestItem    string `json:"test_item"`
	Domain      string `json:"domain"`
	DomainCode  string `json:"domain_code"`
	Standard    string `json:"standard"`
	SourcePages []int  `json:"source_pages"`
}

type pep3BookletItemRange struct {
	SourcePDFPageNo int
	BookletPageNo   int
	Layout          string
	StartItemNo     int
	EndItemNo       int
}

type pep3BookletDomainPlacement struct {
	SourcePDFPageNo int
	BookletPageNo   int
	Layout          string
	DomainCode      string
}

func (svc *Service) GetPEP3AssessmentBooklet(userID, recordID int64) (model.PEP3BookletVO, error) {
	detail, err := svc.GetPEP3AssessmentRecord(userID, recordID)
	if err != nil {
		return model.PEP3BookletVO{}, err
	}
	detail = svc.rescorePEP3AssessmentRecordDetail(detail)
	return buildPEP3Booklet(detail)
}

func buildPEP3Booklet(record model.AssessmentRecordDetailVO) (model.PEP3BookletVO, error) {
	score, err := decodeSavedPEP3Score(record.ResultJSON)
	if err != nil {
		return model.PEP3BookletVO{}, err
	}
	items, err := loadPEP3BookletItems()
	if err != nil {
		return model.PEP3BookletVO{}, err
	}
	itemScores, rawScores, err := decodeSavedPEP3InputScores(record.InputJSON)
	if err != nil {
		return model.PEP3BookletVO{}, err
	}
	itemRecordValues, err := decodeSavedPEP3ItemRecordValues(record.InputJSON)
	if err != nil {
		return model.PEP3BookletVO{}, err
	}
	if len(rawScores) == 0 {
		rawScores = rawScoresFromPEP3Result(score.Result.Scales)
	}

	itemsByDomain := groupPEP3BookletItemsByDomain(items)
	warnings := collectPEP3ReportWarnings(score)
	if len(itemScores) == 0 {
		warnings = append(warnings, "该记录未保存逐题得分，记录册逐题格无法自动填充；仅可填充分量表原始分和常模换算结果。")
	}
	warnings = append(warnings, "记录册页面结构已按扫描版PDF的页面和左右版面输出；少数跨页大题的视觉细分行后续可继续补充坐标级布局。")

	pageCount := pep3BookletJSONPageCount()
	normDataInfo := pep3NormalizeNormDataInfo(score.PEP3NormDataInfo)
	booklet := model.PEP3BookletVO{
		Record:           record.AssessmentRecordSummaryVO,
		TemplateCode:     "PEP3_RECORD_BOOKLET",
		TemplateVersion:  nonEmptyString(score.ScaleVersion, record.ScaleVersion, pep3ScaleVersion),
		Title:            "PEP-3测试员记录册",
		ScaleCode:        nonEmptyString(score.ScaleCode, record.AssessmentCode, pep3ScaleCode),
		ScaleVersion:     nonEmptyString(score.ScaleVersion, record.ScaleVersion),
		PEP3NormDataInfo: normDataInfo,
		DataStatus:       nonEmptyString(score.DataStatus, record.DataStatus),
		Sources:          append([]string(nil), score.Sources...),
		SourcePDF:        pep3RecordBookletPDF,
		Pages:            make([]model.PEP3BookletPage, 0, pageCount),
		Warnings:         uniqueNonEmptyStrings(warnings),
	}

	pages := make(map[int]*model.PEP3BookletPage, pageCount)
	for pageNo := 1; pageNo <= pageCount; pageNo++ {
		page := model.PEP3BookletPage{
			PageNo:          pageNo,
			SourcePDFPageNo: pageNo,
			Title:           fmt.Sprintf("记录册PDF第%d页", pageNo),
			PageType:        "spread",
			Meta: map[string]any{
				"sourcePdf": pep3RecordBookletPDF,
			},
		}
		pages[pageNo] = &page
	}

	pages[1].Sections = append(pages[1].Sections, buildPEP3BookletCoverSections(record, score)...)

	for _, placement := range pep3BookletDomainPlacements() {
		page := pages[placement.SourcePDFPageNo]
		page.Sections = append(page.Sections, buildPEP3BookletDomainSection(placement, itemsByDomain[placement.DomainCode], itemScores, rawScores, score.Result.Scales))
	}
	for _, itemRange := range pep3BookletItemRanges() {
		page := pages[itemRange.SourcePDFPageNo]
		rangeItems := filterPEP3BookletItemsByRange(items, itemRange.StartItemNo, itemRange.EndItemNo)
		page.Sections = append(page.Sections, buildPEP3BookletItemGridSection(itemRange, rangeItems, itemScores, itemRecordValues))
		page.Sections = append(page.Sections, buildPEP3BookletPageTallySection(itemRange, rangeItems, itemScores))
	}

	for pageNo := 1; pageNo <= pageCount; pageNo++ {
		booklet.Pages = append(booklet.Pages, *pages[pageNo])
	}
	return booklet, nil
}

func pep3BookletJSONPageCount() int {
	pageCount := 1
	for _, itemRange := range pep3BookletItemRanges() {
		if itemRange.SourcePDFPageNo > pageCount {
			pageCount = itemRange.SourcePDFPageNo
		}
	}
	for _, placement := range pep3BookletDomainPlacements() {
		if placement.SourcePDFPageNo > pageCount {
			pageCount = placement.SourcePDFPageNo
		}
	}
	return pageCount
}

func buildPEP3BookletCoverSections(record model.AssessmentRecordDetailVO, score PEP3ScoreResponse) []model.PEP3TemplateSection {
	basicInfo := model.PEP3TemplateSection{
		SectionCode:     "basic_info",
		Title:           "第1部分 儿童资料",
		Type:            "field_grid",
		Layout:          "right",
		SourcePDFPageNo: 1,
		BookletPageNo:   1,
		Fields: []model.PEP3TemplateField{
			bookletField("studentName", "儿童姓名", record.StudentName, record.StudentName, ""),
			bookletField("gender", "性别", "", "", "男 / 女"),
			bookletField("centerClass", "中心/班别", "", "", "由前端或业务资料填充"),
			bookletField("examinerName", "测试员姓名", record.ExaminerName, record.ExaminerName, ""),
			bookletField("assessmentDate", "评估日期", formatReportDate(record.AssessmentDate), formatReportDate(record.AssessmentDate), ""),
			bookletField("birthDate", "出生日期", formatReportDate(record.BirthDate), formatReportDate(record.BirthDate), ""),
			bookletField("ageText", "年龄", fmt.Sprintf("%d岁%d个月%d天", record.AgeYears, record.AgeMonths, record.AgeDays), score.Result.Age, ""),
			bookletField("remark", "备注", record.Remark, record.Remark, ""),
		},
	}

	scaleRows := append([]model.PEP3ReportScaleRow{}, buildPEP3ScaleRows(score.Result.Scales, "发展及行为副测验", []string{"CVP", "EL", "RL", "FM", "GM", "VMI", "AE", "SR", "CMB", "CVB"})...)
	scaleRows = append(scaleRows, buildPEP3ScaleRows(score.Result.Scales, "儿童照顾者报告副测验", []string{"PB", "PSC", "AB"})...)
	scoreSummary := model.PEP3TemplateSection{
		SectionCode:     "subtest_scores",
		Title:           "第2部分 副测验分数",
		Type:            "score_summary_table",
		Layout:          "right",
		SourcePDFPageNo: 1,
		BookletPageNo:   1,
		Table: &model.PEP3TemplateTable{
			Columns: []model.PEP3TemplateColumn{
				{Key: "scaleName", Label: "副测验", Width: 160},
				{Key: "scaleCode", Label: "编码", Width: 70, Align: "center"},
				{Key: "rawScore", Label: "原始分", Width: 90, Align: "center"},
				{Key: "developmentAge", Label: "发展/发展年龄", Width: 130, Align: "center"},
				{Key: "percentileRank", Label: "百分比级数", Width: 110, Align: "center"},
				{Key: "scaledScore", Label: "级数", Width: 80, Align: "center"},
				{Key: "level", Label: "适应程度", Width: 110, Align: "center"},
			},
			Rows: pep3ScaleTemplateRows(scaleRows),
		},
	}

	compositeSummary := model.PEP3TemplateSection{
		SectionCode:     "composite_scores",
		Title:           "第3部分 合成分数",
		Type:            "composite_score_table",
		Layout:          "right",
		SourcePDFPageNo: 1,
		BookletPageNo:   1,
		Table: &model.PEP3TemplateTable{
			Columns: pep3CompositeTemplateColumns(110, 54, 110, 110, 120, 120, "发展/适应程度"),
			Rows:    pep3CompositeTemplateRows(buildPEP3CompositeRows(score.Result.Composites, score.Result.Scales)),
		},
	}
	return []model.PEP3TemplateSection{basicInfo, scoreSummary, compositeSummary}
}

func buildPEP3BookletItemGridSection(itemRange pep3BookletItemRange, items []pep3BookletItemDefinition, itemScores map[int]int, itemRecordValues map[int]map[string]any) model.PEP3TemplateSection {
	return model.PEP3TemplateSection{
		SectionCode:     fmt.Sprintf("page_%d_item_grid", itemRange.BookletPageNo),
		Title:           fmt.Sprintf("第%d页 儿童表现记录", itemRange.BookletPageNo),
		Type:            "item_grid",
		Layout:          itemRange.Layout,
		SourcePDFPageNo: itemRange.SourcePDFPageNo,
		BookletPageNo:   itemRange.BookletPageNo,
		Table: &model.PEP3TemplateTable{
			Columns: pep3BookletItemGridColumns(),
			Rows:    pep3BookletItemGridRows(items, itemScores, itemRecordValues),
		},
		Meta: map[string]any{
			"startItemNo": itemRange.StartItemNo,
			"endItemNo":   itemRange.EndItemNo,
		},
	}
}

func buildPEP3BookletPageTallySection(itemRange pep3BookletItemRange, items []pep3BookletItemDefinition, itemScores map[int]int) model.PEP3TemplateSection {
	return model.PEP3TemplateSection{
		SectionCode:     fmt.Sprintf("page_%d_tally", itemRange.BookletPageNo),
		Title:           fmt.Sprintf("第%d页总和", itemRange.BookletPageNo),
		Type:            "page_tally",
		Layout:          itemRange.Layout,
		SourcePDFPageNo: itemRange.SourcePDFPageNo,
		BookletPageNo:   itemRange.BookletPageNo,
		Table: &model.PEP3TemplateTable{
			Columns: pep3BookletDomainColumns("scoreLabel", "计分"),
			Rows:    pep3BookletPageTallyRows(items, itemScores),
		},
	}
}

func buildPEP3BookletDomainSection(placement pep3BookletDomainPlacement, items []pep3BookletItemDefinition, itemScores map[int]int, rawScores map[string]int, scales map[string]pep3score.ScaleResult) model.PEP3TemplateSection {
	scaleName := pep3BookletDomainName(placement.DomainCode, scales)
	section := model.PEP3TemplateSection{
		SectionCode:     "domain_" + strings.ToLower(placement.DomainCode),
		Title:           scaleName + "（" + placement.DomainCode + "）",
		Type:            "domain_score_table",
		Layout:          placement.Layout,
		SourcePDFPageNo: placement.SourcePDFPageNo,
		BookletPageNo:   placement.BookletPageNo,
		Table: &model.PEP3TemplateTable{
			Columns: []model.PEP3TemplateColumn{
				{Key: "itemNo", Label: "No.", Width: 70, Align: "center"},
				{Key: "itemTitle", Label: "项目", Width: 260},
				{Key: "score2", Label: "2", Width: 70, Align: "center"},
				{Key: "score1", Label: "1", Width: 70, Align: "center"},
				{Key: "score0", Label: "0", Width: 70, Align: "center"},
				{Key: "score", Label: "得分", Width: 80, Align: "center"},
			},
			Rows:       pep3BookletDomainRows(items, itemScores),
			FooterRows: []map[string]any{{"label": placement.DomainCode + " 原始分", "rawScore": rawScores[placement.DomainCode]}},
		},
	}
	return section
}

func pep3BookletItemGridColumns() []model.PEP3TemplateColumn {
	columns := []model.PEP3TemplateColumn{
		{Key: "itemNo", Label: "项目", Width: 70, Align: "center"},
		{Key: "itemTitle", Label: "儿童表现记录", Width: 280},
		{Key: "recordValues", Label: "记录值", Width: 180},
		{Key: "score", Label: "得分", Width: 70, Align: "center"},
	}
	return append(columns, pep3BookletDomainColumns("", "")...)
}

func pep3BookletDomainColumns(firstKey, firstLabel string) []model.PEP3TemplateColumn {
	columns := make([]model.PEP3TemplateColumn, 0, 11)
	if firstKey != "" {
		columns = append(columns, model.PEP3TemplateColumn{Key: firstKey, Label: firstLabel, Width: 160})
	}
	for _, code := range pep3BookletDomainOrder() {
		columns = append(columns, model.PEP3TemplateColumn{Key: code, Label: code, Width: 70, Align: "center"})
	}
	return columns
}

func pep3BookletItemGridRows(items []pep3BookletItemDefinition, itemScores map[int]int, itemRecordValues map[int]map[string]any) []map[string]any {
	rows := make([]map[string]any, 0, len(items))
	for _, item := range items {
		score, ok := itemScores[item.ItemNo]
		recordFields := pep3ItemRecordFields(item.ItemNo)
		recordValues := copyPEP3ItemRecordValues(itemRecordValues[item.ItemNo])
		row := map[string]any{
			"itemNo":       item.ItemNo,
			"itemTitle":    nonEmptyString(item.ItemTitle, item.TestItem),
			"domainCode":   item.DomainCode,
			"domainName":   item.Domain,
			"recordFields": recordFields,
			"recordValues": recordValues,
			"score":        "",
			"standard":     item.Standard,
			"sourcePages":  append([]int(nil), item.SourcePages...),
		}
		if ok {
			row["score"] = score
		}
		for _, code := range pep3BookletDomainOrder() {
			row[code] = ""
		}
		if ok {
			row[item.DomainCode] = score
		}
		rows = append(rows, row)
	}
	return rows
}

func copyPEP3ItemRecordValues(values map[string]any) map[string]any {
	if len(values) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(values))
	for key, value := range values {
		out[key] = value
	}
	return out
}

func pep3BookletDomainRows(items []pep3BookletItemDefinition, itemScores map[int]int) []map[string]any {
	rows := make([]map[string]any, 0, len(items))
	for _, item := range items {
		score, ok := itemScores[item.ItemNo]
		row := map[string]any{
			"itemNo":    item.ItemNo,
			"itemTitle": nonEmptyString(item.ItemTitle, item.TestItem),
			"score2":    false,
			"score1":    false,
			"score0":    false,
			"score":     "",
		}
		if ok {
			row["score"] = score
			row["score"+strconv.Itoa(score)] = true
		}
		rows = append(rows, row)
	}
	return rows
}

func pep3BookletPageTallyRows(items []pep3BookletItemDefinition, itemScores map[int]int) []map[string]any {
	rows := make([]map[string]any, 0, 4)
	for _, scoreValue := range []int{2, 1, 0} {
		row := map[string]any{
			"scoreValue": scoreValue,
			"scoreLabel": pep3BookletScoreLabel(scoreValue),
		}
		for _, code := range pep3BookletDomainOrder() {
			row[code] = 0
		}
		for _, item := range items {
			if itemScores[item.ItemNo] == scoreValue {
				row[item.DomainCode] = row[item.DomainCode].(int) + 1
			}
		}
		rows = append(rows, row)
	}
	rawSubtotal := map[string]any{"scoreValue": "rawSubtotal", "scoreLabel": "原始分小计"}
	for _, code := range pep3BookletDomainOrder() {
		rawSubtotal[code] = 0
	}
	for _, item := range items {
		if score, ok := itemScores[item.ItemNo]; ok {
			rawSubtotal[item.DomainCode] = rawSubtotal[item.DomainCode].(int) + score
		}
	}
	rows = append(rows, rawSubtotal)
	return rows
}

func decodeSavedPEP3InputScores(raw json.RawMessage) (map[int]int, map[string]int, error) {
	if len(raw) == 0 {
		return map[int]int{}, map[string]int{}, nil
	}
	var snapshot pep3SavedInputSnapshot
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		return nil, nil, fmt.Errorf("decode PEP-3 input snapshot: %w", err)
	}
	itemScores := make(map[int]int, len(snapshot.ItemScores)+len(snapshot.ItemScoreList))
	for itemNo, score := range snapshot.ItemScores {
		itemScores[itemNo] = score
	}
	for _, item := range snapshot.ItemScoreList {
		if item.ItemNo > 0 {
			itemScores[item.ItemNo] = item.Score
		}
	}
	rawScores := make(map[string]int, len(snapshot.RawScores)+len(snapshot.RawScoreList))
	for scaleCode, rawScore := range snapshot.RawScores {
		if normalized := strings.ToUpper(strings.TrimSpace(scaleCode)); normalized != "" {
			rawScores[normalized] = rawScore
		}
	}
	for _, item := range snapshot.RawScoreList {
		if normalized := strings.ToUpper(strings.TrimSpace(item.ScaleCode)); normalized != "" {
			rawScores[normalized] = item.RawScore
		}
	}
	return itemScores, rawScores, nil
}

func decodeSavedPEP3ItemRecordValues(raw json.RawMessage) (map[int]map[string]any, error) {
	if len(raw) == 0 {
		return map[int]map[string]any{}, nil
	}
	var snapshot pep3SavedInputSnapshot
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		return nil, fmt.Errorf("decode PEP-3 input snapshot: %w", err)
	}
	return normalizeSavedPEP3ItemRecordValues(snapshot), nil
}

func normalizeSavedPEP3ItemRecordValues(snapshot pep3SavedInputSnapshot) map[int]map[string]any {
	out := make(map[int]map[string]any, len(snapshot.ItemRecordValues)+len(snapshot.ItemRecordValueList))
	for itemNo, values := range snapshot.ItemRecordValues {
		if itemNo <= 0 {
			continue
		}
		for fieldKey, value := range values {
			addPEP3ItemRecordValue(out, itemNo, fieldKey, value)
		}
	}
	for _, item := range snapshot.ItemRecordValueList {
		addPEP3ItemRecordValue(out, item.ItemNo, item.FieldKey, item.Value)
	}
	return out
}

func addPEP3ItemRecordValue(out map[int]map[string]any, itemNo int, fieldKey string, value any) {
	fieldKey = strings.TrimSpace(fieldKey)
	if itemNo <= 0 || fieldKey == "" {
		return
	}
	if out[itemNo] == nil {
		out[itemNo] = map[string]any{}
	}
	out[itemNo][fieldKey] = value
}

func loadPEP3BookletItems() ([]pep3BookletItemDefinition, error) {
	data, err := loadPEP3StaticData()
	if err != nil {
		return nil, err
	}
	return unmarshalPEP3BookletItems(data.itemRows)
}

func groupPEP3BookletItemsByDomain(items []pep3BookletItemDefinition) map[string][]pep3BookletItemDefinition {
	groups := make(map[string][]pep3BookletItemDefinition)
	for _, item := range items {
		groups[item.DomainCode] = append(groups[item.DomainCode], item)
	}
	return groups
}

func filterPEP3BookletItemsByRange(items []pep3BookletItemDefinition, start, end int) []pep3BookletItemDefinition {
	out := make([]pep3BookletItemDefinition, 0, end-start+1)
	for _, item := range items {
		if item.ItemNo >= start && item.ItemNo <= end {
			out = append(out, item)
		}
	}
	return out
}

func rawScoresFromPEP3Result(scales map[string]pep3score.ScaleResult) map[string]int {
	rawScores := make(map[string]int, len(scales))
	for scaleCode, scale := range scales {
		rawScores[scaleCode] = scale.RawScore
	}
	return rawScores
}

func pep3BookletDomainName(domainCode string, scales map[string]pep3score.ScaleResult) string {
	if scale, ok := scales[domainCode]; ok && strings.TrimSpace(scale.ScaleName) != "" {
		return scale.ScaleName
	}
	names := map[string]string{
		"CVP": "认知（语言/语前）",
		"EL":  "语言表达",
		"RL":  "语言理解",
		"FM":  "小肌肉",
		"GM":  "大肌肉",
		"VMI": "模仿（视觉/动作）",
		"AE":  "情感表达",
		"SR":  "社交互动",
		"CMB": "行为特征-非语言",
		"CVB": "行为特征-语言",
	}
	return nonEmptyString(names[domainCode], domainCode)
}

func pep3BookletDomainOrder() []string {
	return []string{"CVP", "EL", "RL", "FM", "GM", "VMI", "AE", "SR", "CMB", "CVB"}
}

func pep3BookletScoreLabel(score int) string {
	switch score {
	case 2:
		return "（2）通过 / 恰当"
	case 1:
		return "（1）部分通过 / 轻微"
	case 0:
		return "（0）未能通过 / 严重"
	default:
		return strconv.Itoa(score)
	}
}

func pep3BookletItemRanges() []pep3BookletItemRange {
	return []pep3BookletItemRange{
		{SourcePDFPageNo: 2, BookletPageNo: 2, Layout: "full", StartItemNo: 1, EndItemNo: 14},
		{SourcePDFPageNo: 3, BookletPageNo: 3, Layout: "right", StartItemNo: 15, EndItemNo: 27},
		{SourcePDFPageNo: 4, BookletPageNo: 4, Layout: "left", StartItemNo: 28, EndItemNo: 37},
		{SourcePDFPageNo: 5, BookletPageNo: 5, Layout: "right", StartItemNo: 38, EndItemNo: 49},
		{SourcePDFPageNo: 6, BookletPageNo: 6, Layout: "left", StartItemNo: 50, EndItemNo: 63},
		{SourcePDFPageNo: 7, BookletPageNo: 7, Layout: "right", StartItemNo: 64, EndItemNo: 77},
		{SourcePDFPageNo: 8, BookletPageNo: 8, Layout: "left", StartItemNo: 78, EndItemNo: 84},
		{SourcePDFPageNo: 9, BookletPageNo: 9, Layout: "right", StartItemNo: 85, EndItemNo: 85},
		{SourcePDFPageNo: 10, BookletPageNo: 10, Layout: "left", StartItemNo: 86, EndItemNo: 90},
		{SourcePDFPageNo: 11, BookletPageNo: 11, Layout: "right", StartItemNo: 91, EndItemNo: 100},
		{SourcePDFPageNo: 12, BookletPageNo: 12, Layout: "left", StartItemNo: 101, EndItemNo: 111},
		{SourcePDFPageNo: 13, BookletPageNo: 13, Layout: "full", StartItemNo: 112, EndItemNo: 122},
		{SourcePDFPageNo: 14, BookletPageNo: 14, Layout: "full", StartItemNo: 123, EndItemNo: 133},
		{SourcePDFPageNo: 15, BookletPageNo: 15, Layout: "full", StartItemNo: 134, EndItemNo: 151},
		{SourcePDFPageNo: 16, BookletPageNo: 16, Layout: "full", StartItemNo: 152, EndItemNo: 172},
	}
}

func pep3BookletDomainPlacements() []pep3BookletDomainPlacement {
	return []pep3BookletDomainPlacement{
		{SourcePDFPageNo: 3, BookletPageNo: 26, Layout: "left", DomainCode: "CMB"},
		{SourcePDFPageNo: 3, BookletPageNo: 26, Layout: "left", DomainCode: "CVB"},
		{SourcePDFPageNo: 4, BookletPageNo: 25, Layout: "right", DomainCode: "AE"},
		{SourcePDFPageNo: 4, BookletPageNo: 25, Layout: "right", DomainCode: "SR"},
		{SourcePDFPageNo: 5, BookletPageNo: 24, Layout: "left", DomainCode: "GM"},
		{SourcePDFPageNo: 5, BookletPageNo: 24, Layout: "left", DomainCode: "VMI"},
		{SourcePDFPageNo: 6, BookletPageNo: 23, Layout: "right", DomainCode: "FM"},
		{SourcePDFPageNo: 7, BookletPageNo: 22, Layout: "left", DomainCode: "RL"},
		{SourcePDFPageNo: 8, BookletPageNo: 21, Layout: "right", DomainCode: "EL"},
		{SourcePDFPageNo: 9, BookletPageNo: 20, Layout: "left", DomainCode: "CVP"},
	}
}

func bookletField(key, label, value string, rawValue any, placeholder string) model.PEP3TemplateField {
	field := templateField(key, label, value, rawValue)
	field.Placeholder = placeholder
	return field
}
