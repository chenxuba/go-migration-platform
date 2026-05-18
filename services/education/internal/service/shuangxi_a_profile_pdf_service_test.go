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

func TestShuangxiAProfileScoreYUsesThreeIntervalAxis(t *testing.T) {
	const (
		top  = 170.0
		rowH = 112.0
		max  = 168
	)
	cases := []struct {
		name string
		raw  int
		want float64
	}{
		{name: "max at level three", raw: 168, want: top},
		{name: "two thirds at level two", raw: 112, want: top + rowH},
		{name: "one third at level one", raw: 56, want: top + rowH*2},
		{name: "zero at level zero", raw: 0, want: top + rowH*3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := shuangxiAProfileScoreY(tc.raw, max, top, rowH)
			if got != tc.want {
				t.Fatalf("score y = %v, want %v", got, tc.want)
			}
		})
	}
	if got := shuangxiAProfileScoreY(76, max, top, rowH); !(got > top+rowH && got < top+rowH*2) {
		t.Fatalf("score 76 should sit between 112 and 56 lines, got y=%v", got)
	}
}

func TestShuangxiAProfileSkillsAndScores(t *testing.T) {
	dataDir, err := resolveShuangxiADataDir()
	if err != nil {
		t.Fatalf("resolve Shuangxi data dir: %v", err)
	}
	data, err := loadShuangxiAStaticDataFromFiles(dataDir)
	if err != nil {
		t.Fatalf("load Shuangxi data: %v", err)
	}
	skills := shuangxiAProfileSkills(data)
	if len(skills) != 34 {
		t.Fatalf("skills = %d, want 34", len(skills))
	}
	if skills[0].Code != "1.1" || skills[0].MaxRawScore != 24 {
		t.Fatalf("first skill = %+v, want 1.1 max 24", skills[0])
	}
	if skills[len(skills)-1].Code != "7.8" || skills[len(skills)-1].MaxRawScore != 12 {
		t.Fatalf("last skill = %+v, want 7.8 max 12", skills[len(skills)-1])
	}
	input, err := json.Marshal(shuangxiASavedInputSnapshot{
		ItemScores: map[int]int{
			1:  2,
			2:  3,
			22: 1,
			23: 2,
		},
	})
	if err != nil {
		t.Fatalf("marshal input: %v", err)
	}
	got := shuangxiAProfileSkillScoreMap(data, model.AssessmentRecordDetailVO{InputJSON: input})
	if got["1.1"] != 5 {
		t.Fatalf("skill 1.1 score = %d, want 5", got["1.1"])
	}
	if got["2.1"] != 3 {
		t.Fatalf("skill 2.1 score = %d, want 3", got["2.1"])
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
