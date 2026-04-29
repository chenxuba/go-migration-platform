package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"go-migration-platform/pkg/pep3score"
	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) GetPEP3AssessmentReport(userID, recordID int64) (model.PEP3ReportVO, error) {
	detail, err := svc.GetPEP3AssessmentRecord(userID, recordID)
	if err != nil {
		return model.PEP3ReportVO{}, err
	}
	detail = svc.rescorePEP3AssessmentRecordDetail(detail)
	return buildPEP3Report(detail)
}

func (svc *Service) rescorePEP3AssessmentRecordDetail(record model.AssessmentRecordDetailVO) model.AssessmentRecordDetailVO {
	if record.BirthDate == nil || record.BirthDate.IsZero() || record.AssessmentDate == nil || record.AssessmentDate.IsZero() {
		return record
	}
	itemScores, rawScores, err := decodeSavedPEP3InputScores(record.InputJSON)
	if err != nil || (len(itemScores) == 0 && len(rawScores) == 0) {
		return record
	}
	var snapshot pep3SavedInputSnapshot
	_ = json.Unmarshal(record.InputJSON, &snapshot)
	score, err := svc.ScorePEP3(pep3score.AssessmentInput{
		BirthDate:         *record.BirthDate,
		AssessmentDate:    *record.AssessmentDate,
		ItemScores:        itemScores,
		RawScores:         rawScores,
		AllowMissingItems: snapshot.AllowMissingItems,
	})
	if err != nil {
		return record
	}
	raw, err := json.Marshal(score)
	if err != nil {
		return record
	}
	record.ResultJSON = raw
	record.ScaleVersion = score.ScaleVersion
	record.DataStatus = score.DataStatus
	return record
}

func buildPEP3Report(record model.AssessmentRecordDetailVO) (model.PEP3ReportVO, error) {
	score, err := decodeSavedPEP3Score(record.ResultJSON)
	if err != nil {
		return model.PEP3ReportVO{}, err
	}
	developmentRows := buildPEP3ScaleRows(score.Result.Scales, "发展量表", []string{"CVP", "EL", "RL", "FM", "GM", "VMI"})
	behaviorRows := buildPEP3ScaleRows(score.Result.Scales, "适应不良行为", []string{"AE", "SR", "CMB", "CVB"})
	caregiverReportRows := buildPEP3ScaleRows(score.Result.Scales, "照顾者报告", []string{"PB", "PSC", "AB"})
	compositeRows := buildPEP3CompositeRows(score.Result.Composites, score.Result.Scales)
	warnings := collectPEP3ReportWarnings(score)
	summary := buildPEP3ReportSummary(compositeRows, warnings)

	report := model.PEP3ReportVO{
		Record:          record.AssessmentRecordSummaryVO,
		TemplateCode:    "PEP3_EXPLANATORY_REPORT",
		TemplateVersion: nonEmptyString(score.ScaleVersion, record.ScaleVersion, pep3ScaleVersion),
		Title:           "PEP-3解释性报告",
		ScaleCode:       nonEmptyString(score.ScaleCode, record.AssessmentCode, pep3ScaleCode),
		ScaleVersion:    nonEmptyString(score.ScaleVersion, record.ScaleVersion),
		DataStatus:      nonEmptyString(score.DataStatus, record.DataStatus),
		Sources:         append([]string(nil), score.Sources...),
		Sections:        buildPEP3ReportSections(record, developmentRows, behaviorRows, caregiverReportRows, compositeRows, summary, warnings),
	}
	return report, nil
}

func decodeSavedPEP3Score(raw json.RawMessage) (PEP3ScoreResponse, error) {
	if len(raw) == 0 {
		return PEP3ScoreResponse{}, errors.New("assessment result is empty")
	}
	var score PEP3ScoreResponse
	if err := json.Unmarshal(raw, &score); err == nil && score.Result.Scales != nil {
		return score, nil
	}

	var result pep3score.AssessmentResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return PEP3ScoreResponse{}, fmt.Errorf("decode PEP-3 result: %w", err)
	}
	if result.Scales == nil {
		return PEP3ScoreResponse{}, errors.New("assessment result does not contain PEP-3 scales")
	}
	return PEP3ScoreResponse{
		PEP3ScoreDataInfo: PEP3ScoreDataInfo{
			ScaleCode:    pep3ScaleCode,
			ScaleVersion: pep3ScaleVersion,
		},
		Result: result,
	}, nil
}

func buildPEP3ScaleRows(scales map[string]pep3score.ScaleResult, category string, order []string) []model.PEP3ReportScaleRow {
	rows := make([]model.PEP3ReportScaleRow, 0, len(order))
	for _, code := range order {
		scale, ok := scales[code]
		if !ok {
			continue
		}
		developmentAgeText := formatNormText(scale.DevelopmentAge)
		if developmentAgeText == "" && !pep3ScaleHasDevelopmentAge(code) {
			developmentAgeText = "--"
		}
		rows = append(rows, model.PEP3ReportScaleRow{
			ScaleCode:          scale.ScaleCode,
			ScaleName:          scale.ScaleName,
			Category:           category,
			RawScore:           scale.RawScore,
			MaxRawScore:        copyIntPtr(scale.MaxRawScore),
			DevelopmentAgeText: developmentAgeText,
			PercentileRankText: formatNormText(scale.PercentileRank),
			ScaledScoreText:    formatNormText(scale.ScaledScore),
			Level:              scale.Level,
			Warnings:           append([]string(nil), scale.Warnings...),
		})
	}
	return rows
}

func buildPEP3CompositeRows(composites map[string]pep3score.CompositeResult, scales map[string]pep3score.ScaleResult) []model.PEP3ReportCompositeRow {
	order := []string{
		pep3score.CompositeCommunication,
		pep3score.CompositeMotor,
		pep3score.CompositeMaladaptiveBehavior,
	}
	rows := make([]model.PEP3ReportCompositeRow, 0, len(order))
	for _, code := range order {
		composite, ok := composites[code]
		if !ok {
			continue
		}
		developmentAgeText := formatFloatMonths(composite.DevelopmentAgeMonths)
		if developmentAgeText == "" && !pep3CompositeHasDevelopmentAge(code) {
			developmentAgeText = "--"
		}
		rows = append(rows, model.PEP3ReportCompositeRow{
			CompositeCode:        composite.CompositeCode,
			CompositeName:        composite.CompositeName,
			MemberScaleCodes:     append([]string(nil), composite.MemberScaleCodes...),
			MemberScaleScores:    pep3CompositeMemberScoreTexts(composite.MemberScaleCodes, scales),
			StandardScoreSumText: formatIntPtr(composite.StandardScoreSum),
			PercentileRankText:   formatNormText(composite.PercentileRank),
			DevelopmentAgeText:   developmentAgeText,
			Level:                composite.Level,
			Warnings:             append([]string(nil), composite.Warnings...),
		})
	}
	return rows
}

func collectPEP3ReportWarnings(score PEP3ScoreResponse) []string {
	var warnings []string
	if strings.TrimSpace(score.DataStatus) != "" {
		warnings = append(warnings, score.DataStatus)
	}
	warnings = append(warnings, score.Result.Warnings...)
	for _, scale := range score.Result.Scales {
		for _, warning := range scale.Warnings {
			warnings = append(warnings, fmt.Sprintf("%s：%s", nonEmptyString(scale.ScaleName, scale.ScaleCode), warning))
		}
	}
	for _, composite := range score.Result.Composites {
		for _, warning := range composite.Warnings {
			warnings = append(warnings, fmt.Sprintf("%s：%s", nonEmptyString(composite.CompositeName, composite.CompositeCode), warning))
		}
	}
	return uniqueNonEmptyStrings(warnings)
}

func buildPEP3ReportSummary(compositeRows []model.PEP3ReportCompositeRow, warnings []string) []string {
	summary := make([]string, 0, 4)
	for _, row := range compositeRows {
		if row.PercentileRankText == "" && row.Level == "" {
			continue
		}
		text := row.CompositeName
		if row.PercentileRankText != "" {
			text += "百分比级数为" + row.PercentileRankText
		}
		if row.Level != "" {
			text += "，等级为" + row.Level
		}
		summary = append(summary, text)
	}
	if len(warnings) > 0 {
		summary = append(summary, "本报告包含数据或换算提示，正式出具前需复核。")
	}
	return summary
}

func buildPEP3ReportSections(
	record model.AssessmentRecordDetailVO,
	developmentRows []model.PEP3ReportScaleRow,
	behaviorRows []model.PEP3ReportScaleRow,
	caregiverReportRows []model.PEP3ReportScaleRow,
	compositeRows []model.PEP3ReportCompositeRow,
	summary []string,
	warnings []string,
) []model.PEP3TemplateSection {
	sections := []model.PEP3TemplateSection{
		{
			SectionCode: "basic_info",
			Title:       "基本资料",
			Type:        "field_grid",
			Fields: []model.PEP3TemplateField{
				templateField("studentId", "学员ID", formatInt64(record.StudentID), record.StudentID),
				templateField("studentName", "儿童姓名", record.StudentName, record.StudentName),
				templateField("examinerId", "测试员ID", formatInt64(record.ExaminerID), record.ExaminerID),
				templateField("examinerName", "测试员姓名", record.ExaminerName, record.ExaminerName),
				templateField("birthDate", "出生日期", formatReportDate(record.BirthDate), formatReportDate(record.BirthDate)),
				templateField("assessmentDate", "评估日期", formatReportDate(record.AssessmentDate), formatReportDate(record.AssessmentDate)),
				templateField("ageText", "实足年龄", fmt.Sprintf("%d岁%d个月%d天", record.AgeYears, record.AgeMonths, record.AgeDays), map[string]int{"years": record.AgeYears, "months": record.AgeMonths, "days": record.AgeDays}),
				templateField("normAgeMonths", "常模月龄", strconv.Itoa(record.NormAgeMonths), record.NormAgeMonths),
				templateField("remark", "备注", record.Remark, record.Remark),
			},
		},
		buildPEP3ReportScaleSection("development_scores", "发展量表", developmentRows),
		buildPEP3ReportScaleSection("behavior_scores", "适应不良行为", behaviorRows),
		buildPEP3ReportScaleSection("caregiver_scores", "照顾者报告", caregiverReportRows),
		{
			SectionCode: "composite_scores",
			Title:       "合成分数",
			Type:        "composite_table",
			Table: &model.PEP3TemplateTable{
				Columns: pep3CompositeTemplateColumns(160, 70, 120, 120, 120, 120, "适应程度"),
				Rows:    pep3CompositeTemplateRows(compositeRows),
			},
		},
		{
			SectionCode: "summary",
			Title:       "解释摘要",
			Type:        "summary",
			TextItems:   append([]string(nil), summary...),
		},
	}
	if len(warnings) > 0 {
		sections = append(sections, model.PEP3TemplateSection{
			SectionCode: "warnings",
			Title:       "复核提示",
			Type:        "warnings",
			TextItems:   append([]string(nil), warnings...),
		})
	}
	return sections
}

func buildPEP3ReportScaleSection(sectionCode, title string, rows []model.PEP3ReportScaleRow) model.PEP3TemplateSection {
	return model.PEP3TemplateSection{
		SectionCode: sectionCode,
		Title:       title,
		Type:        "score_table",
		Table: &model.PEP3TemplateTable{
			Columns: []model.PEP3TemplateColumn{
				{Key: "scaleCode", Label: "编码", Width: 80, Align: "center"},
				{Key: "scaleName", Label: "副测验", Width: 180},
				{Key: "rawScore", Label: "原始分", Width: 90, Align: "center"},
				{Key: "maxRawScore", Label: "满分", Width: 80, Align: "center"},
				{Key: "developmentAge", Label: "发展年龄", Width: 120, Align: "center"},
				{Key: "percentileRank", Label: "百分比级数", Width: 120, Align: "center"},
				{Key: "scaledScore", Label: "级数", Width: 90, Align: "center"},
				{Key: "level", Label: "适应程度", Width: 120, Align: "center"},
			},
			Rows: pep3ScaleTemplateRows(rows),
		},
	}
}

func pep3ScaleHasDevelopmentAge(scaleCode string) bool {
	switch scaleCode {
	case "CVP", "EL", "RL", "FM", "GM", "VMI", "PSC":
		return true
	default:
		return false
	}
}

func pep3CompositeHasDevelopmentAge(compositeCode string) bool {
	switch compositeCode {
	case pep3score.CompositeCommunication, pep3score.CompositeMotor:
		return true
	default:
		return false
	}
}

func pep3CompositeScaleCodes() []string {
	return []string{"CVP", "EL", "RL", "FM", "GM", "VMI", "AE", "SR", "CMB", "CVB"}
}

func pep3CompositeTemplateColumns(nameWidth, scoreWidth, sumWidth, percentileWidth, levelWidth, devAgeWidth int, levelLabel string) []model.PEP3TemplateColumn {
	columns := []model.PEP3TemplateColumn{
		{Key: "compositeName", Label: "合成项目", Width: nameWidth},
	}
	for _, code := range pep3CompositeScaleCodes() {
		columns = append(columns, model.PEP3TemplateColumn{
			Key:   code,
			Label: code,
			Width: scoreWidth,
			Align: "center",
			Group: "标准分 / Standard Scores (SSs)",
		})
	}
	columns = append(columns,
		model.PEP3TemplateColumn{Key: "standardScoreSum", Label: "标准分总和", Width: sumWidth, Align: "center"},
		model.PEP3TemplateColumn{Key: "percentileRank", Label: "百分比级数", Width: percentileWidth, Align: "center"},
		model.PEP3TemplateColumn{Key: "level", Label: levelLabel, Width: levelWidth, Align: "center"},
		model.PEP3TemplateColumn{Key: "developmentAge", Label: "发展年龄", Width: devAgeWidth, Align: "center"},
	)
	return columns
}

func pep3CompositeMemberScoreTexts(memberScaleCodes []string, scales map[string]pep3score.ScaleResult) map[string]string {
	members := make(map[string]bool, len(memberScaleCodes))
	for _, code := range memberScaleCodes {
		members[code] = true
	}
	out := make(map[string]string, len(pep3CompositeScaleCodes()))
	for _, code := range pep3CompositeScaleCodes() {
		if !members[code] {
			out[code] = "--"
			continue
		}
		scale, ok := scales[code]
		if !ok {
			out[code] = "待校对"
			continue
		}
		text := formatNormText(scale.ScaledScore)
		if text == "" {
			text = "待校对"
		}
		out[code] = text
	}
	return out
}

func pep3ScaleTemplateRows(rows []model.PEP3ReportScaleRow) []map[string]any {
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		item := map[string]any{
			"scaleCode":      row.ScaleCode,
			"scaleName":      row.ScaleName,
			"category":       row.Category,
			"rawScore":       row.RawScore,
			"developmentAge": row.DevelopmentAgeText,
			"percentileRank": row.PercentileRankText,
			"scaledScore":    row.ScaledScoreText,
			"level":          row.Level,
			"warnings":       append([]string(nil), row.Warnings...),
		}
		if row.MaxRawScore != nil {
			item["maxRawScore"] = *row.MaxRawScore
		} else {
			item["maxRawScore"] = ""
		}
		out = append(out, item)
	}
	return out
}

func pep3CompositeTemplateRows(rows []model.PEP3ReportCompositeRow) []map[string]any {
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		item := map[string]any{
			"compositeCode":     row.CompositeCode,
			"compositeName":     row.CompositeName,
			"memberScaleCodes":  append([]string(nil), row.MemberScaleCodes...),
			"memberScaleScores": copyStringMap(row.MemberScaleScores),
			"standardScoreSum":  row.StandardScoreSumText,
			"percentileRank":    row.PercentileRankText,
			"developmentAge":    row.DevelopmentAgeText,
			"level":             row.Level,
			"warnings":          append([]string(nil), row.Warnings...),
		}
		for _, code := range pep3CompositeScaleCodes() {
			item[code] = row.MemberScaleScores[code]
		}
		out = append(out, item)
	}
	return out
}

func templateField(key, label, value string, rawValue any) model.PEP3TemplateField {
	return model.PEP3TemplateField{
		Key:      key,
		Label:    label,
		Value:    value,
		RawValue: rawValue,
	}
}

func formatInt64(value int64) string {
	if value == 0 {
		return ""
	}
	return strconv.FormatInt(value, 10)
}

func formatReportDate(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case interface{ Format(string) string }:
		return v.Format("2006-01-02")
	default:
		return ""
	}
}

func formatNormText(value *pep3score.NormValue) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(value.Text)
}

func formatIntPtr(value *int) string {
	if value == nil {
		return ""
	}
	return strconv.Itoa(*value)
}

func formatFloatMonths(value *float64) string {
	if value == nil {
		return ""
	}
	rounded := math.Round(*value*10) / 10
	if rounded == math.Trunc(rounded) {
		return fmt.Sprintf("%.0f个月", rounded)
	}
	return fmt.Sprintf("%.1f个月", rounded)
}

func copyIntPtr(value *int) *int {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

func copyStringMap(value map[string]string) map[string]string {
	if len(value) == 0 {
		return nil
	}
	copied := make(map[string]string, len(value))
	for key, item := range value {
		copied[key] = item
	}
	return copied
}

func nonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func uniqueNonEmptyStrings(values []string) []string {
	seen := make(map[string]bool, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}
