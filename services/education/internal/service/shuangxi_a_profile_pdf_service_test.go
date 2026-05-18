package service

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"go-migration-platform/services/education/internal/model"
)

func TestBuildShuangxiADevelopmentProfilePDFLandscape(t *testing.T) {
	dataDir, err := resolveShuangxiADataDir()
	if err != nil {
		t.Fatalf("resolve Shuangxi data dir: %v", err)
	}
	data, err := loadShuangxiAStaticDataFromFiles(dataDir)
	if err != nil {
		t.Fatalf("load Shuangxi data: %v", err)
	}
	assessmentDate := time.Date(2026, 5, 18, 0, 0, 0, 0, time.Local)
	birthDate := time.Date(2018, 1, 1, 0, 0, 0, 0, time.Local)
	record := model.AssessmentRecordDetailVO{
		AssessmentRecordSummaryVO: model.AssessmentRecordSummaryVO{
			ID:             51,
			StudentID:      61,
			StudentName:    "双溪学生",
			StudentGender:  "女",
			AssessmentCode: shuangxiAScaleCode,
			AssessmentName: "双溪课程评量表A",
			ScaleVersion:   shuangxiAScaleVersion,
			BirthDate:      &birthDate,
			AssessmentDate: &assessmentDate,
			ExaminerName:   "陈老师",
		},
		ResultJSON: shuangxiAProfileTestResultJSON(t, data),
	}
	content, err := buildShuangxiADevelopmentProfilePDF(data, []model.AssessmentRecordDetailVO{record})
	if err != nil {
		t.Fatalf("build profile PDF: %v", err)
	}
	if !bytes.HasPrefix(content, []byte("%PDF")) {
		t.Fatalf("expected PDF header, got %q", content[:min(len(content), 8)])
	}
	if len(content) < 1000 {
		t.Fatalf("expected non-empty profile PDF, got %d bytes", len(content))
	}
}

func shuangxiAProfileTestResultJSON(t *testing.T, data shuangxiAStaticData) []byte {
	t.Helper()
	rows := make([]shuangxiADomainScoreResult, 0, len(data.domains))
	totalRaw := 0
	maxRaw := 0
	for _, domain := range data.domains {
		raw := domain.MaxRawScore / 2
		totalRaw += raw
		maxRaw += domain.MaxRawScore
		rows = append(rows, shuangxiADomainScoreResult{
			DomainCode:        domain.ScaleCode,
			DomainName:        domain.ScaleName,
			ItemCount:         domain.ItemCount,
			AnsweredItemCount: domain.ItemCount,
			RawScore:          raw,
			MaxRawScore:       domain.MaxRawScore,
			CompletionPercent: 100,
			Complete:          true,
		})
	}
	score := shuangxiAAssessmentScoreResponse{
		ScaleCode:    shuangxiAScaleCode,
		ScaleVersion: shuangxiAScaleVersion,
		DataStatus:   "ready",
		Result: shuangxiAAssessmentResult{
			ItemCount:         data.metadata.ItemCount,
			AnsweredItemCount: data.metadata.ItemCount,
			TotalRawScore:     totalRaw,
			MaxRawScore:       maxRaw,
			CompletionPercent: 100,
			Complete:          true,
			DomainScores:      rows,
		},
	}
	raw, err := json.Marshal(score)
	if err != nil {
		t.Fatalf("marshal score: %v", err)
	}
	return raw
}
