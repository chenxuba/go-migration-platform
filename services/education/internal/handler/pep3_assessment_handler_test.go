package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"go-migration-platform/pkg/tenant"
	"go-migration-platform/services/education/internal/service"
)

func TestPEP3ScoreRequestToAssessmentInputUsesRawScoreList(t *testing.T) {
	req := pep3ScoreRequest{
		BirthDate:      "2000-10-29",
		AssessmentDate: "2004-02-10",
		RawScoreList: []pep3RawScoreRequest{
			{ScaleCode: "cvp", RawScore: 16},
			{ScaleCode: " EL ", RawScore: 18},
		},
	}

	input, err := req.toAssessmentInput()
	if err != nil {
		t.Fatalf("toAssessmentInput returned error: %v", err)
	}
	if input.BirthDate.Year() != 2000 || input.BirthDate.Month() != 10 || input.BirthDate.Day() != 29 {
		t.Fatalf("unexpected birth date: %s", input.BirthDate)
	}
	if input.AssessmentDate.Year() != 2004 || input.AssessmentDate.Month() != 2 || input.AssessmentDate.Day() != 10 {
		t.Fatalf("unexpected assessment date: %s", input.AssessmentDate)
	}
	if input.RawScores["CVP"] != 16 || input.RawScores["EL"] != 18 {
		t.Fatalf("unexpected raw scores: %+v", input.RawScores)
	}
}

func TestPEP3ScoreRequestToAssessmentInputRequiresScores(t *testing.T) {
	req := pep3ScoreRequest{
		BirthDate:      "2000-10-29",
		AssessmentDate: "2004-02-10",
	}

	if _, err := req.toAssessmentInput(); err == nil {
		t.Fatal("expected missing score error")
	}
}

func TestPEP3AssessmentRecordCreateRequestToScoreRequest(t *testing.T) {
	req := pep3AssessmentRecordCreateRequest{
		StudentID:      1001,
		StudentName:    "李东尼",
		BirthDate:      "2000-10-29",
		AssessmentDate: "2004-02-10",
		RawScores:      map[string]int{"cvp": 16},
	}

	scoreReq := req.toScoreRequest()
	input, err := scoreReq.toAssessmentInput()
	if err != nil {
		t.Fatalf("toAssessmentInput returned error: %v", err)
	}
	if input.RawScores["CVP"] != 16 {
		t.Fatalf("unexpected raw scores: %+v", input.RawScores)
	}
}

func TestScorePEP3EndpointWithGeneratedDraftsWhenPresent(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..")
	itemPath := filepath.Join(root, "docs", "pep3-item-bank-simplified-draft.json")
	if _, err := os.Stat(itemPath); err != nil {
		t.Skipf("generated PEP-3 draft data not present: %s", itemPath)
	}

	reqBody := []byte(`{
		"birthDate": "2000-10-29",
		"assessmentDate": "2004-02-10",
		"rawScores": {
			"CVP": 16, "EL": 18, "RL": 12, "FM": 34, "GM": 27, "VMI": 11,
			"AE": 3, "SR": 6, "CMB": 7, "CVB": 10,
			"PB": 7, "PSC": 7, "AB": 10
		}
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/assessments/pep3/score", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	h := New(&service.Service{})
	tenant.Middleware(http.HandlerFunc(h.scorePEP3)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status %d: %s", rec.Code, rec.Body.String())
	}
	var envelope struct {
		Success bool `json:"success"`
		Data    struct {
			ScaleCode    string `json:"scaleCode"`
			ScaleVersion string `json:"scaleVersion"`
			Result       struct {
				Age struct {
					Years              int `json:"years"`
					Months             int `json:"months"`
					Days               int `json:"days"`
					TotalMonthsForNorm int `json:"total_months_for_norm"`
				} `json:"age"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !envelope.Success {
		t.Fatalf("expected success response: %+v", envelope)
	}
	if envelope.Data.ScaleCode != "PEP3" || envelope.Data.ScaleVersion != "2025-draft" {
		t.Fatalf("unexpected scale info: %+v", envelope.Data)
	}
	if envelope.Data.Result.Age.Years != 3 || envelope.Data.Result.Age.Months != 3 || envelope.Data.Result.Age.Days != 11 || envelope.Data.Result.Age.TotalMonthsForNorm != 39 {
		t.Fatalf("unexpected age: %+v", envelope.Data.Result.Age)
	}
}
