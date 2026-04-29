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
		Record:       record.AssessmentRecordSummaryVO,
		ScaleCode:    nonEmptyString(score.ScaleCode, record.AssessmentCode, pep3ScaleCode),
		ScaleVersion: nonEmptyString(score.ScaleVersion, record.ScaleVersion),
		DataStatus:   nonEmptyString(score.DataStatus, record.DataStatus),
		Sources:      append([]string(nil), score.Sources...),
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
