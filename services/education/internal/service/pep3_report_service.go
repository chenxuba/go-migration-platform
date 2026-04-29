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
	return buildPEP3Report(detail)
}

func buildPEP3Report(record model.AssessmentRecordDetailVO) (model.PEP3ReportVO, error) {
	score, err := decodeSavedPEP3Score(record.ResultJSON)
	if err != nil {
		return model.PEP3ReportVO{}, err
	}

	report := model.PEP3ReportVO{
		Record:          record.AssessmentRecordSummaryVO,
		TemplateCode:    "PEP3_EXPLANATORY_REPORT",
		TemplateVersion: nonEmptyString(score.ScaleVersion, record.ScaleVersion, pep3ScaleVersion),
		Title:           "PEP-3解释性报告",
		ScaleCode:       nonEmptyString(score.ScaleCode, record.AssessmentCode, pep3ScaleCode),
		ScaleVersion:    nonEmptyString(score.ScaleVersion, record.ScaleVersion),
		DataStatus:      nonEmptyString(score.DataStatus, record.DataStatus),
		Sources:         append([]string(nil), score.Sources...),
		BasicInfo: model.PEP3ReportBasicInfo{
			StudentID:      record.StudentID,
			StudentName:    record.StudentName,
			ExaminerID:     record.ExaminerID,
			ExaminerName:   record.ExaminerName,
			BirthDate:      formatReportDate(record.BirthDate),
			AssessmentDate: formatReportDate(record.AssessmentDate),
			AgeText:        fmt.Sprintf("%d岁%d个月%d天", record.AgeYears, record.AgeMonths, record.AgeDays),
			NormAgeMonths:  record.NormAgeMonths,
			Remark:         record.Remark,
		},
		DevelopmentRows:     buildPEP3ScaleRows(score.Result.Scales, "发展量表", []string{"CVP", "EL", "RL", "FM", "GM", "VMI"}),
		BehaviorRows:        buildPEP3ScaleRows(score.Result.Scales, "适应不良行为", []string{"AE", "SR", "CMB", "CVB"}),
		CaregiverReportRows: buildPEP3ScaleRows(score.Result.Scales, "照顾者报告", []string{"PB", "PSC", "AB"}),
		CompositeRows:       buildPEP3CompositeRows(score.Result.Composites),
		Warnings:            collectPEP3ReportWarnings(score),
	}
	report.Summary = buildPEP3ReportSummary(report)
	report.Sections = buildPEP3ReportSections(report)
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
		rows = append(rows, model.PEP3ReportScaleRow{
			ScaleCode:          scale.ScaleCode,
			ScaleName:          scale.ScaleName,
			Category:           category,
			RawScore:           scale.RawScore,
			MaxRawScore:        copyIntPtr(scale.MaxRawScore),
			DevelopmentAgeText: formatNormText(scale.DevelopmentAge),
			PercentileRankText: formatNormText(scale.PercentileRank),
			ScaledScoreText:    formatNormText(scale.ScaledScore),
			Level:              scale.Level,
			Warnings:           append([]string(nil), scale.Warnings...),
		})
	}
	return rows
}

func buildPEP3CompositeRows(composites map[string]pep3score.CompositeResult) []model.PEP3ReportCompositeRow {
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
		rows = append(rows, model.PEP3ReportCompositeRow{
			CompositeCode:        composite.CompositeCode,
			CompositeName:        composite.CompositeName,
			MemberScaleCodes:     append([]string(nil), composite.MemberScaleCodes...),
			StandardScoreSumText: formatIntPtr(composite.StandardScoreSum),
			PercentileRankText:   formatNormText(composite.PercentileRank),
			DevelopmentAgeText:   formatFloatMonths(composite.DevelopmentAgeMonths),
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

func buildPEP3ReportSummary(report model.PEP3ReportVO) []string {
	summary := make([]string, 0, 4)
	for _, row := range report.CompositeRows {
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
	if len(report.Warnings) > 0 {
		summary = append(summary, "本报告包含数据或换算提示，正式出具前需复核。")
	}
	return summary
}

func buildPEP3ReportSections(report model.PEP3ReportVO) []model.PEP3TemplateSection {
	sections := []model.PEP3TemplateSection{
		{
			SectionCode: "basic_info",
			Title:       "基本资料",
			Type:        "field_grid",
			Fields: []model.PEP3TemplateField{
				templateField("studentId", "学员ID", formatInt64(report.BasicInfo.StudentID), report.BasicInfo.StudentID),
				templateField("studentName", "儿童姓名", report.BasicInfo.StudentName, report.BasicInfo.StudentName),
				templateField("examinerId", "测试员ID", formatInt64(report.BasicInfo.ExaminerID), report.BasicInfo.ExaminerID),
				templateField("examinerName", "测试员姓名", report.BasicInfo.ExaminerName, report.BasicInfo.ExaminerName),
				templateField("birthDate", "出生日期", report.BasicInfo.BirthDate, report.BasicInfo.BirthDate),
				templateField("assessmentDate", "评估日期", report.BasicInfo.AssessmentDate, report.BasicInfo.AssessmentDate),
				templateField("ageText", "实足年龄", report.BasicInfo.AgeText, report.BasicInfo.AgeText),
				templateField("normAgeMonths", "常模月龄", strconv.Itoa(report.BasicInfo.NormAgeMonths), report.BasicInfo.NormAgeMonths),
				templateField("remark", "备注", report.BasicInfo.Remark, report.BasicInfo.Remark),
			},
		},
		buildPEP3ReportScaleSection("development_scores", "发展量表", report.DevelopmentRows),
		buildPEP3ReportScaleSection("behavior_scores", "适应不良行为", report.BehaviorRows),
		buildPEP3ReportScaleSection("caregiver_scores", "照顾者报告", report.CaregiverReportRows),
		{
			SectionCode: "composite_scores",
			Title:       "合成分数",
			Type:        "composite_table",
			Table: &model.PEP3TemplateTable{
				Columns: []model.PEP3TemplateColumn{
					{Key: "compositeName", Label: "合成项目", Width: 160},
					{Key: "memberScaleCodes", Label: "包含副测验", Width: 180},
					{Key: "standardScoreSum", Label: "标准分总和", Width: 120, Align: "center"},
					{Key: "percentileRank", Label: "百分比级数", Width: 120, Align: "center"},
					{Key: "developmentAge", Label: "发展年龄", Width: 120, Align: "center"},
					{Key: "level", Label: "适应程度", Width: 120, Align: "center"},
				},
				Rows: pep3CompositeTemplateRows(report.CompositeRows),
			},
		},
		{
			SectionCode: "summary",
			Title:       "解释摘要",
			Type:        "summary",
			TextItems:   append([]string(nil), report.Summary...),
		},
	}
	if len(report.Warnings) > 0 {
		sections = append(sections, model.PEP3TemplateSection{
			SectionCode: "warnings",
			Title:       "复核提示",
			Type:        "warnings",
			TextItems:   append([]string(nil), report.Warnings...),
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
		out = append(out, map[string]any{
			"compositeCode":    row.CompositeCode,
			"compositeName":    row.CompositeName,
			"memberScaleCodes": append([]string(nil), row.MemberScaleCodes...),
			"standardScoreSum": row.StandardScoreSumText,
			"percentileRank":   row.PercentileRankText,
			"developmentAge":   row.DevelopmentAgeText,
			"level":            row.Level,
			"warnings":         append([]string(nil), row.Warnings...),
		})
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
