package service

import (
	"testing"
	"time"

	"go-migration-platform/services/education/internal/model"
)

func TestBuildPEP3AssessmentDraftProgressForPartialItemScores(t *testing.T) {
	birthDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local)
	assessmentDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)

	progress, err := buildPEP3AssessmentDraftProgress(&birthDate, &assessmentDate, map[int]int{1: 2, 2: 1}, nil, false)
	if err != nil {
		t.Fatalf("buildPEP3AssessmentDraftProgress returned error: %v", err)
	}
	if progress.ItemCount != 172 || progress.AnsweredItemCount != 2 || progress.MissingItemCount != 170 {
		t.Fatalf("unexpected item progress: %+v", progress)
	}
	if progress.CanScore || progress.Complete {
		t.Fatalf("partial item scores without allowMissingItems should not be ready: %+v", progress)
	}
	fmProgress := findPEP3DomainProgressForTest(progress.DomainProgress, "FM")
	if fmProgress == nil || fmProgress.AnsweredItemCount != 2 || fmProgress.Complete {
		t.Fatalf("unexpected FM progress: %+v", fmProgress)
	}
	if fmProgress.RawScore == nil || *fmProgress.RawScore != 3 {
		t.Fatalf("expected FM raw score to be auto-summed from item scores, got: %+v", fmProgress)
	}
	if status := pep3DraftStatus(progress); status != "draft" {
		t.Fatalf("unexpected draft status: %s", status)
	}
}

func TestBuildPEP3AssessmentDraftProgressForCaregiverOnlyScores(t *testing.T) {
	birthDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local)
	assessmentDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)

	progress, err := buildPEP3AssessmentDraftProgress(&birthDate, &assessmentDate, nil, map[string]int{"PB": 7, "PSC": 7, "AB": 10}, true)
	if err != nil {
		t.Fatalf("buildPEP3AssessmentDraftProgress returned error: %v", err)
	}
	if progress.CanScore || progress.Complete {
		t.Fatalf("caregiver-only scores should not be ready for scoring: %+v", progress)
	}
	if progress.RawScoreCount != 3 || progress.CaregiverRawScoreCount != 3 {
		t.Fatalf("unexpected caregiver progress: %+v", progress)
	}
}

func TestBuildPEP3AssessmentDraftProgressForRawScores(t *testing.T) {
	birthDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local)
	assessmentDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)
	rawScores := map[string]int{
		"CVP": 16, "EL": 18, "RL": 12, "FM": 34, "GM": 27, "VMI": 11,
		"AE": 3, "SR": 6, "CMB": 7, "CVB": 10,
		"PB": 7, "PSC": 7, "AB": 10,
	}

	progress, err := buildPEP3AssessmentDraftProgress(&birthDate, &assessmentDate, nil, rawScores, false)
	if err != nil {
		t.Fatalf("buildPEP3AssessmentDraftProgress returned error: %v", err)
	}
	if !progress.CanScore || !progress.Complete || progress.CompletionPercent != 100 {
		t.Fatalf("raw scores should complete the draft: %+v", progress)
	}
	if progress.RawScoreCount != 13 || progress.CaregiverRawScoreCount != 3 {
		t.Fatalf("unexpected raw score progress: %+v", progress)
	}
	if status := pep3DraftStatus(progress); status != "complete" {
		t.Fatalf("unexpected draft status: %s", status)
	}
}

func findPEP3DomainProgressForTest(rows []model.PEP3DomainProgress, scaleCode string) *model.PEP3DomainProgress {
	for i := range rows {
		if rows[i].ScaleCode == scaleCode {
			return &rows[i]
		}
	}
	return nil
}
