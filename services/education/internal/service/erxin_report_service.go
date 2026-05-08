package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/pkg/erxinscore"
	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) GetERXinAssessmentReport(userID, recordID int64) (model.ERXinReportVO, error) {
	detail, err := svc.GetERXinAssessmentRecord(userID, recordID)
	if err != nil {
		return model.ERXinReportVO{}, err
	}
	detail = svc.rescoreERXinAssessmentRecordDetail(detail)
	return buildERXinReport(detail)
}

func (svc *Service) rescoreERXinAssessmentRecordDetail(record model.AssessmentRecordDetailVO) model.AssessmentRecordDetailVO {
	if record.BirthDate == nil || record.BirthDate.IsZero() || record.AssessmentDate == nil || record.AssessmentDate.IsZero() {
		return record
	}
	itemPasses, err := decodeSavedERXinInputPasses(record.InputJSON)
	if err != nil || len(itemPasses) == 0 {
		return record
	}
	score, err := svc.ScoreERXin(erxinscore.AssessmentInput{
		BirthDate:      *record.BirthDate,
		AssessmentDate: *record.AssessmentDate,
		ItemPasses:     itemPasses,
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

func buildERXinReport(record model.AssessmentRecordDetailVO) (model.ERXinReportVO, error) {
	score, err := decodeSavedERXinScore(record.ResultJSON)
	if err != nil {
		return model.ERXinReportVO{}, err
	}
	domainRows := buildERXinReportDomainRows(score.Result.Domains)
	warnings := collectERXinReportWarnings(score)
	summary := model.ERXinReportSummary{
		MainAgeMonth:            score.Result.MainAgeMonth,
		MeanMentalAgeMonths:     score.Result.MeanMentalAgeMonths,
		MeanMentalAgeMonthsText: score.Result.MeanMentalAgeMonthsText,
		DQ:                      score.Result.DQ,
		Level:                   score.Result.Level,
		Complete:                score.Result.Complete,
	}
	report := model.ERXinReportVO{
		Record:          record.AssessmentRecordSummaryVO,
		TemplateCode:    "ERXIN2_DEVELOPMENT_REPORT",
		TemplateVersion: nonEmptyString(score.ScaleVersion, record.ScaleVersion, erxinScaleVersion),
		Title:           "儿心量表-II发育行为评估报告",
		ScaleCode:       nonEmptyString(score.ScaleCode, record.AssessmentCode, erxinScaleCode),
		ScaleVersion:    nonEmptyString(score.ScaleVersion, record.ScaleVersion, erxinScaleVersion),
		SourceStandard:  score.SourceStandard,
		SourcePDF:       score.SourcePDF,
		DataStatus:      nonEmptyString(score.DataStatus, record.DataStatus),
		Sources:         append([]string(nil), score.Sources...),
		Summary:         summary,
		DomainRows:      domainRows,
		Sections:        buildERXinReportSections(record, score, domainRows, warnings),
		Warnings:        warnings,
	}
	return report, nil
}

func decodeSavedERXinScore(raw json.RawMessage) (ERXinScoreResponse, error) {
	if len(raw) == 0 {
		return ERXinScoreResponse{}, errors.New("assessment result is empty")
	}
	var score ERXinScoreResponse
	if err := json.Unmarshal(raw, &score); err == nil && len(score.Result.Domains) > 0 {
		score.ERXinScoreDataInfo = erxinNormalizeScoreDataInfo(score.ERXinScoreDataInfo)
		return score, nil
	}

	var result erxinscore.AssessmentResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return ERXinScoreResponse{}, fmt.Errorf("decode ERXin result: %w", err)
	}
	if len(result.Domains) == 0 {
		return ERXinScoreResponse{}, errors.New("assessment result does not contain ERXin domains")
	}
	return ERXinScoreResponse{
		ERXinScoreDataInfo: erxinDefaultScoreDataInfo(),
		Result:             result,
	}, nil
}

func erxinNormalizeScoreDataInfo(info ERXinScoreDataInfo) ERXinScoreDataInfo {
	defaults := erxinDefaultScoreDataInfo()
	if strings.TrimSpace(info.ScaleCode) == "" {
		info.ScaleCode = defaults.ScaleCode
	}
	if strings.TrimSpace(info.ScaleVersion) == "" {
		info.ScaleVersion = defaults.ScaleVersion
	}
	if strings.TrimSpace(info.SourceStandard) == "" {
		info.SourceStandard = defaults.SourceStandard
	}
	if strings.TrimSpace(info.SourcePDF) == "" {
		info.SourcePDF = defaults.SourcePDF
	}
	if len(info.Sources) == 0 {
		info.Sources = defaults.Sources
	}
	return info
}

func erxinDefaultScoreDataInfo() ERXinScoreDataInfo {
	data, err := loadERXinStaticData()
	if err == nil {
		return erxinScoreDataInfo(data)
	}
	return ERXinScoreDataInfo{
		ScaleCode:    erxinScaleCode,
		ScaleVersion: erxinScaleVersion,
	}
}

func buildERXinReportDomainRows(domains []erxinscore.DomainResult) []model.ERXinReportDomainRow {
	rows := make([]model.ERXinReportDomainRow, 0, len(domains))
	for _, domain := range domains {
		rows = append(rows, model.ERXinReportDomainRow{
			DomainCode:          domain.DomainCode,
			DomainName:          domain.DomainName,
			MentalAgeMonths:     domain.MentalAgeMonths,
			MentalAgeMonthsText: domain.MentalAgeMonthsText,
			DQ:                  domain.DQ,
			Level:               domain.Level,
			BasalAgeMonth:       domain.BasalAgeMonth,
			CeilingAgeMonth:     domain.CeilingAgeMonth,
			Complete:            domain.Complete,
			MissingItemNumbers:  append([]int(nil), domain.MissingItemNumbers...),
			Warnings:            append([]string(nil), domain.Warnings...),
		})
	}
	return rows
}

func collectERXinReportWarnings(score ERXinScoreResponse) []string {
	warnings := make([]string, 0)
	if strings.TrimSpace(score.DataStatus) != "" {
		warnings = append(warnings, score.DataStatus)
	}
	warnings = append(warnings, score.Result.Warnings...)
	for _, domain := range score.Result.Domains {
		for _, warning := range domain.Warnings {
			warnings = append(warnings, fmt.Sprintf("%s：%s", nonEmptyString(domain.DomainName, domain.DomainCode), warning))
		}
	}
	return uniqueNonEmptyStrings(warnings)
}

func buildERXinReportSections(record model.AssessmentRecordDetailVO, score ERXinScoreResponse, domainRows []model.ERXinReportDomainRow, warnings []string) []model.PEP3TemplateSection {
	sections := []model.PEP3TemplateSection{
		buildERXinReportBasicInfoSection(record),
		buildERXinReportSummarySection(score),
		buildERXinReportDomainSection(domainRows),
	}
	if len(warnings) > 0 {
		sections = append(sections, model.PEP3TemplateSection{
			SectionCode: "warnings",
			Title:       "提示",
			Type:        "text",
			TextItems:   warnings,
		})
	}
	return sections
}

func buildERXinReportBasicInfoSection(record model.AssessmentRecordDetailVO) model.PEP3TemplateSection {
	return model.PEP3TemplateSection{
		SectionCode: "basic_info",
		Title:       "基本信息",
		Type:        "fields",
		Fields: []model.PEP3TemplateField{
			{Key: "studentName", Label: "儿童姓名", Value: record.StudentName, RawValue: record.StudentName},
			{Key: "birthDate", Label: "出生日期", Value: erxinFormatDate(record.BirthDate), RawValue: record.BirthDate},
			{Key: "assessmentDate", Label: "测查日期", Value: erxinFormatDate(record.AssessmentDate), RawValue: record.AssessmentDate},
			{Key: "actualAge", Label: "实足年龄", Value: erxinActualAgeText(record), RawValue: map[string]int{"years": record.AgeYears, "months": record.AgeMonths, "days": record.AgeDays}},
			{Key: "examinerName", Label: "测试员", Value: record.ExaminerName, RawValue: record.ExaminerName},
		},
	}
}

func buildERXinReportSummarySection(score ERXinScoreResponse) model.PEP3TemplateSection {
	return model.PEP3TemplateSection{
		SectionCode: "score_summary",
		Title:       "评分概览",
		Type:        "fields",
		Fields: []model.PEP3TemplateField{
			{Key: "mainAgeMonth", Label: "主测月龄", Value: strconv.Itoa(score.Result.MainAgeMonth) + "个月", RawValue: score.Result.MainAgeMonth},
			{Key: "meanMentalAgeMonths", Label: "平均智龄", Value: nonEmptyString(score.Result.MeanMentalAgeMonthsText, erxinMonthText(score.Result.MeanMentalAgeMonths)), RawValue: score.Result.MeanMentalAgeMonths},
			{Key: "dq", Label: "发育商DQ", Value: erxinFloatText(score.Result.DQ), RawValue: score.Result.DQ},
			{Key: "level", Label: "评价等级", Value: score.Result.Level, RawValue: score.Result.Level},
			{Key: "complete", Label: "评分完整", Value: erxinBoolText(score.Result.Complete), RawValue: score.Result.Complete},
		},
	}
}

func buildERXinReportDomainSection(rows []model.ERXinReportDomainRow) model.PEP3TemplateSection {
	tableRows := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		tableRows = append(tableRows, map[string]any{
			"domainCode":          row.DomainCode,
			"domainName":          row.DomainName,
			"mentalAgeMonths":     row.MentalAgeMonths,
			"mentalAgeMonthsText": nonEmptyString(row.MentalAgeMonthsText, erxinMonthText(row.MentalAgeMonths)),
			"dq":                  erxinFloatText(row.DQ),
			"level":               row.Level,
			"basalAgeMonth":       erxinAgeMonthText(row.BasalAgeMonth),
			"ceilingAgeMonth":     erxinAgeMonthText(row.CeilingAgeMonth),
			"complete":            erxinBoolText(row.Complete),
			"missingItemNumbers":  erxinIntListText(row.MissingItemNumbers),
			"warnings":            strings.Join(row.Warnings, "；"),
		})
	}
	return model.PEP3TemplateSection{
		SectionCode: "domain_scores",
		Title:       "五大能区结果",
		Type:        "table",
		Table: &model.PEP3TemplateTable{
			Columns: []model.PEP3TemplateColumn{
				{Key: "domainName", Label: "能区", Width: 120},
				{Key: "mentalAgeMonthsText", Label: "智龄", Width: 90, Align: "center"},
				{Key: "dq", Label: "DQ", Width: 70, Align: "center"},
				{Key: "level", Label: "等级", Width: 100, Align: "center"},
				{Key: "basalAgeMonth", Label: "基线", Width: 90, Align: "center"},
				{Key: "ceilingAgeMonth", Label: "封顶", Width: 90, Align: "center"},
				{Key: "complete", Label: "完整", Width: 70, Align: "center"},
				{Key: "missingItemNumbers", Label: "缺测题号", Width: 180},
				{Key: "warnings", Label: "提示", Width: 220},
			},
			Rows: tableRows,
		},
	}
}

func erxinFormatDate(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02")
}

func erxinActualAgeText(record model.AssessmentRecordDetailVO) string {
	parts := make([]string, 0, 3)
	if record.AgeYears > 0 {
		parts = append(parts, strconv.Itoa(record.AgeYears)+"岁")
	}
	if record.AgeMonths > 0 {
		parts = append(parts, strconv.Itoa(record.AgeMonths)+"个月")
	}
	if record.AgeDays > 0 {
		parts = append(parts, strconv.Itoa(record.AgeDays)+"天")
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "")
}

func erxinMonthText(value float64) string {
	rounded := math.Round(value*10) / 10
	if rounded == math.Trunc(rounded) {
		return fmt.Sprintf("%.0f个月", rounded)
	}
	return fmt.Sprintf("%.1f个月", rounded)
}

func erxinAgeMonthText(value int) string {
	if value <= 0 {
		return ""
	}
	return strconv.Itoa(value) + "个月"
}

func erxinFloatText(value float64) string {
	rounded := math.Round(value*10) / 10
	if rounded == math.Trunc(rounded) {
		return fmt.Sprintf("%.0f", rounded)
	}
	return fmt.Sprintf("%.1f", rounded)
}

func erxinBoolText(value bool) string {
	if value {
		return "是"
	}
	return "否"
}

func erxinIntListText(values []int) string {
	if len(values) == 0 {
		return ""
	}
	copied := append([]int(nil), values...)
	sort.Ints(copied)
	parts := make([]string, 0, len(copied))
	for _, value := range copied {
		parts = append(parts, strconv.Itoa(value))
	}
	return strings.Join(parts, "、")
}
