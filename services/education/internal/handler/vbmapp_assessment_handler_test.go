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

func TestVBMAPPScoreRequestToAssessmentInputNormalizesLists(t *testing.T) {
	req := vbmappScoreRequest{
		MilestoneScores: map[string]float64{
			" mand_01m ": 0.5,
		},
		MilestoneScoreList: []vbmappMilestoneScoreRequest{
			{MilestoneID: "mand_02m", Score: 1},
		},
		BarrierScoreList: []vbmappBarrierScoreRequest{
			{BarrierCode: "b01", Score: 3},
		},
		TransitionScoreList: []vbmappTransitionScoreRequest{
			{TransitionCode: "t01", Score: 2},
		},
		PreviousMilestoneScoreList: []vbmappMilestoneScoreRequest{
			{MilestoneID: "mand_02m", Score: 0.5},
		},
	}

	input, err := req.toAssessmentInput()
	if err != nil {
		t.Fatalf("toAssessmentInput returned error: %v", err)
	}
	if input.MilestoneScores["MAND_01M"] != 0.5 || input.MilestoneScores["MAND_02M"] != 1 {
		t.Fatalf("unexpected milestone scores: %+v", input.MilestoneScores)
	}
	if input.BarrierScores["B01"] != 3 {
		t.Fatalf("unexpected barrier scores: %+v", input.BarrierScores)
	}
	if input.TransitionScores["T01"] != 2 {
		t.Fatalf("unexpected transition scores: %+v", input.TransitionScores)
	}
	if input.PreviousMilestoneScores["MAND_02M"] != 0.5 {
		t.Fatalf("unexpected previous milestone scores: %+v", input.PreviousMilestoneScores)
	}
}

func TestVBMAPPDraftSaveRequestAllowsPartialScores(t *testing.T) {
	req := vbmappAssessmentDraftSaveRequest{
		StudentID:    1001,
		StudentName:  " 测试儿童 ",
		ExaminerName: " 测试员 ",
		MilestoneScoreList: []vbmappMilestoneScoreRequest{
			{MilestoneID: "mand_02m", Score: 0.5},
		},
		BarrierScores: map[string]int{
			" b01 ": 3,
		},
	}

	input, err := req.toDraftSaveInput()
	if err != nil {
		t.Fatalf("toDraftSaveInput returned error: %v", err)
	}
	if input.BirthDate != nil || input.AssessmentDate != nil {
		t.Fatalf("draft dates should be optional: %+v", input)
	}
	if input.StudentName != "测试儿童" || input.ExaminerName != "测试员" {
		t.Fatalf("expected trimmed names: %+v", input)
	}
	if input.ScoreInput.MilestoneScores["MAND_02M"] != 0.5 {
		t.Fatalf("unexpected milestone scores: %+v", input.ScoreInput.MilestoneScores)
	}
	if input.ScoreInput.BarrierScores["B01"] != 3 {
		t.Fatalf("unexpected barrier scores: %+v", input.ScoreInput.BarrierScores)
	}

	raw, err := json.Marshal(input.InputSnapshot)
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	var snapshot struct {
		StudentName        string                        `json:"studentName"`
		MilestoneScoreList []vbmappMilestoneScoreRequest `json:"milestoneScoreList"`
		BarrierScoreList   []vbmappBarrierScoreRequest   `json:"barrierScoreList"`
	}
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	if snapshot.StudentName != "测试儿童" || len(snapshot.MilestoneScoreList) != 1 || snapshot.MilestoneScoreList[0].MilestoneID != "MAND_02M" {
		t.Fatalf("unexpected normalized snapshot: %+v", snapshot)
	}
	if len(snapshot.BarrierScoreList) != 1 || snapshot.BarrierScoreList[0].BarrierCode != "B01" {
		t.Fatalf("unexpected barrier snapshot: %+v", snapshot)
	}
}

func TestVBMAPPRecordCreateRequestRequiresDatesAndNormalizesScores(t *testing.T) {
	req := vbmappAssessmentRecordCreateRequest{
		StudentID:      1001,
		StudentName:    " 测试儿童 ",
		ExaminerName:   " 测试员 ",
		BirthDate:      "2020-01-01",
		AssessmentDate: "2024-01-01",
		MilestoneScoreList: []vbmappMilestoneScoreRequest{
			{MilestoneID: "mand_01m", Score: 1},
		},
		TransitionScoreList: []vbmappTransitionScoreRequest{
			{TransitionCode: "t01", Score: 3},
		},
	}

	input, err := req.toRecordSaveInput()
	if err != nil {
		t.Fatalf("toRecordSaveInput returned error: %v", err)
	}
	if input.BirthDate.Year() != 2020 || input.AssessmentDate.Year() != 2024 {
		t.Fatalf("unexpected dates: %+v", input)
	}
	if input.StudentName != "测试儿童" || input.ExaminerName != "测试员" {
		t.Fatalf("expected trimmed names: %+v", input)
	}
	if input.ScoreInput.MilestoneScores["MAND_01M"] != 1 {
		t.Fatalf("unexpected milestone scores: %+v", input.ScoreInput.MilestoneScores)
	}
	if input.ScoreInput.TransitionScores["T01"] != 3 {
		t.Fatalf("unexpected transition scores: %+v", input.ScoreInput.TransitionScores)
	}

	_, err = (vbmappAssessmentRecordCreateRequest{
		StudentID:      1001,
		StudentName:    "测试儿童",
		AssessmentDate: "2024-01-01",
	}).toRecordSaveInput()
	if err == nil {
		t.Fatalf("expected missing birthDate to fail")
	}
}

func TestScoreVBMAPPEndpointWithGeneratedDraftsWhenPresent(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..")
	dataDir := filepath.Join(root, "docs", "vbmapp")
	if _, err := os.Stat(filepath.Join(dataDir, "milestone-items.json")); err != nil {
		t.Skipf("generated VB-MAPP data not present: %s", dataDir)
	}

	body, err := json.Marshal(vbmappScoreRequest{
		MilestoneScoreList: []vbmappMilestoneScoreRequest{
			{MilestoneID: "MAND_01M", Score: 1},
			{MilestoneID: "MAND_02M", Score: 0.5},
		},
		BarrierScoreList: []vbmappBarrierScoreRequest{
			{BarrierCode: "B01", Score: 3},
			{BarrierCode: "B02", Score: 2},
		},
		TransitionScoreList: []vbmappTransitionScoreRequest{
			{TransitionCode: "T06", Score: 3},
		},
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/assessments/vbmapp/score", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h := New(&service.Service{})
	tenant.Middleware(http.HandlerFunc(h.scoreVBMAPP)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status %d: %s", rec.Code, rec.Body.String())
	}
	var envelope struct {
		Success bool `json:"success"`
		Data    struct {
			ScaleCode    string `json:"scaleCode"`
			ScaleVersion string `json:"scaleVersion"`
			DataStatus   string `json:"dataStatus"`
			Result       struct {
				ModuleProgress []struct {
					ModuleCode    string  `json:"moduleCode"`
					AnsweredItems int     `json:"answeredItems"`
					Score         float64 `json:"score"`
				} `json:"moduleProgress"`
				Milestones struct {
					TotalScore    float64 `json:"totalScore"`
					AnsweredItems int     `json:"answeredItems"`
				} `json:"milestones"`
				Barriers struct {
					TotalScore    int `json:"totalScore"`
					AnsweredItems int `json:"answeredItems"`
				} `json:"barriers"`
				Transition struct {
					Suggestions []struct {
						TransitionCode string `json:"transitionCode"`
						Score          int    `json:"score"`
					} `json:"suggestions"`
				} `json:"transition"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !envelope.Success {
		t.Fatalf("expected success response: %+v", envelope)
	}
	if envelope.Data.ScaleCode != "VBMAPP" || envelope.Data.ScaleVersion == "" || envelope.Data.DataStatus != "draft" {
		t.Fatalf("unexpected scale info: %+v", envelope.Data)
	}
	if envelope.Data.Result.Milestones.TotalScore != 1.5 || envelope.Data.Result.Milestones.AnsweredItems != 2 {
		t.Fatalf("unexpected milestone result: %+v", envelope.Data.Result.Milestones)
	}
	if envelope.Data.Result.Barriers.TotalScore != 5 || envelope.Data.Result.Barriers.AnsweredItems != 2 {
		t.Fatalf("unexpected barrier result: %+v", envelope.Data.Result.Barriers)
	}
	if len(envelope.Data.Result.ModuleProgress) != 3 {
		t.Fatalf("unexpected module progress: %+v", envelope.Data.Result.ModuleProgress)
	}
	if len(envelope.Data.Result.Transition.Suggestions) == 0 {
		t.Fatalf("expected transition suggestions: %+v", envelope.Data.Result.Transition)
	}
}

func TestVBMAPPSchemaEndpointWithGeneratedDraftsWhenPresent(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..")
	dataDir := filepath.Join(root, "docs", "vbmapp")
	if _, err := os.Stat(filepath.Join(dataDir, "milestone-response-schemas.json")); err != nil {
		t.Skipf("generated VB-MAPP response schema data not present: %s", dataDir)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/assessments/vbmapp/schema", nil)
	rec := httptest.NewRecorder()

	h := New(&service.Service{})
	tenant.Middleware(http.HandlerFunc(h.vbmappAssessmentSchema)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status %d: %s", rec.Code, rec.Body.String())
	}
	var envelope struct {
		Success bool `json:"success"`
		Data    struct {
			ScaleCode                 string `json:"scaleCode"`
			MilestoneItems            []any  `json:"milestoneItems"`
			MilestoneResponseSchemas  []any  `json:"milestoneResponseSchemas"`
			BarrierResponseSchemas    []any  `json:"barrierResponseSchemas"`
			TransitionResponseSchemas []any  `json:"transitionResponseSchemas"`
			ResponseSchemaSummary     struct {
				ItemCount int `json:"itemCount"`
			} `json:"responseSchemaSummary"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !envelope.Success || envelope.Data.ScaleCode != "VBMAPP" {
		t.Fatalf("expected success response: %+v", envelope)
	}
	if len(envelope.Data.MilestoneItems) != 170 || len(envelope.Data.MilestoneResponseSchemas) != 170 {
		t.Fatalf("unexpected milestone schema counts: %+v", envelope.Data)
	}
	if len(envelope.Data.BarrierResponseSchemas) != 24 || len(envelope.Data.TransitionResponseSchemas) != 18 {
		t.Fatalf("unexpected module schema counts: %+v", envelope.Data)
	}
	if envelope.Data.ResponseSchemaSummary.ItemCount != 212 {
		t.Fatalf("unexpected schema summary: %+v", envelope.Data.ResponseSchemaSummary)
	}
}
