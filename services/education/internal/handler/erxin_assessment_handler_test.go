package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go-migration-platform/pkg/authx"
	"go-migration-platform/pkg/erxinscore"
	"go-migration-platform/pkg/tenant"
	"go-migration-platform/services/education/internal/service"
)

func TestERXinScoreRequestToAssessmentInputKeepsFalsePasses(t *testing.T) {
	req := erxinScoreRequest{
		BirthDate:      "2020-01-01",
		AssessmentDate: "2021-10-01",
		ItemPassList: []erxinItemPassRequest{
			{ItemNo: 112, Passed: true},
			{ItemNo: 138, Passed: false},
		},
	}

	input, err := req.toAssessmentInput()
	if err != nil {
		t.Fatalf("toAssessmentInput returned error: %v", err)
	}
	if input.BirthDate.Year() != 2020 || input.AssessmentDate.Year() != 2021 {
		t.Fatalf("unexpected dates: %+v", input)
	}
	if !input.ItemPasses[112] {
		t.Fatalf("expected item 112 to pass: %+v", input.ItemPasses)
	}
	if passed, ok := input.ItemPasses[138]; !ok || passed {
		t.Fatalf("expected item 138 to be present and false: %+v", input.ItemPasses)
	}
}

func TestERXinDraftSaveRequestAllowsPartialItemPasses(t *testing.T) {
	req := erxinAssessmentDraftSaveRequest{
		StudentName: "测试儿童",
		ItemPassList: []erxinItemPassRequest{
			{ItemNo: 112, Passed: true},
			{ItemNo: 138, Passed: false},
		},
	}

	input, err := req.toDraftSaveInput()
	if err != nil {
		t.Fatalf("toDraftSaveInput returned error: %v", err)
	}
	if input.BirthDate != nil || input.AssessmentDate != nil {
		t.Fatalf("draft dates should be optional: %+v", input)
	}
	if !input.ItemPasses[112] {
		t.Fatalf("expected item 112 to pass: %+v", input.ItemPasses)
	}
	if passed, ok := input.ItemPasses[138]; !ok || passed {
		t.Fatalf("expected item 138 to be present and false: %+v", input.ItemPasses)
	}
}

func TestERXinAssessmentRecordCreateRequestToRecordSaveInput(t *testing.T) {
	req := erxinAssessmentRecordCreateRequest{
		StudentID:      1001,
		StudentName:    " 测试儿童 ",
		ExaminerName:   " 测试员 ",
		BirthDate:      "2020-01-01",
		AssessmentDate: "2021-10-01",
		ItemPassList:   []erxinItemPassRequest{{ItemNo: 112, Passed: true}, {ItemNo: 138, Passed: false}},
		ItemPasses:     map[int]bool{139: true},
	}

	input, err := req.toRecordSaveInput()
	if err != nil {
		t.Fatalf("toRecordSaveInput returned error: %v", err)
	}
	if input.StudentName != "测试儿童" || input.ExaminerName != "测试员" {
		t.Fatalf("expected trimmed names: %+v", input)
	}
	if input.ScoreInput.BirthDate.Year() != 2020 || input.ScoreInput.AssessmentDate.Year() != 2021 {
		t.Fatalf("unexpected dates: %+v", input.ScoreInput)
	}
	if !input.ScoreInput.ItemPasses[112] || !input.ScoreInput.ItemPasses[139] {
		t.Fatalf("expected passed items to be preserved: %+v", input.ScoreInput.ItemPasses)
	}
	if passed, ok := input.ScoreInput.ItemPasses[138]; !ok || passed {
		t.Fatalf("expected failed item to be preserved: %+v", input.ScoreInput.ItemPasses)
	}
}

func TestScoreERXinEndpointWithGeneratedDraftsWhenPresent(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..")
	itemPath := filepath.Join(root, "docs", "erxin-item-bank-draft.json")
	if _, err := os.Stat(itemPath); err != nil {
		t.Skipf("generated ERXin draft data not present: %s", itemPath)
	}

	items, err := erxinscore.LoadItemDefinitionsFile(itemPath)
	if err != nil {
		t.Fatalf("load ERXin items: %v", err)
	}
	itemPassList := make([]erxinItemPassRequest, 0)
	for _, item := range items {
		switch item.AgeMonth {
		case 15, 18, 21:
			itemPassList = append(itemPassList, erxinItemPassRequest{ItemNo: item.ItemNo, Passed: true})
		case 24, 27:
			itemPassList = append(itemPassList, erxinItemPassRequest{ItemNo: item.ItemNo, Passed: false})
		}
	}
	body, err := json.Marshal(erxinScoreRequest{
		BirthDate:      "2020-01-01",
		AssessmentDate: "2021-10-01",
		ItemPassList:   itemPassList,
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/assessments/erxin/score", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h := New(&service.Service{})
	tenant.Middleware(http.HandlerFunc(h.scoreERXin)).ServeHTTP(rec, req)

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
				MainAgeMonth        int     `json:"mainAgeMonth"`
				MeanMentalAgeMonths float64 `json:"meanMentalAgeMonths"`
				DQ                  float64 `json:"dq"`
				Complete            bool    `json:"complete"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !envelope.Success {
		t.Fatalf("expected success response: %+v", envelope)
	}
	if envelope.Data.ScaleCode != "ERXIN2" || envelope.Data.ScaleVersion != "WS-T-580-2017" || envelope.Data.DataStatus != "draft" {
		t.Fatalf("unexpected scale info: %+v", envelope.Data)
	}
	if envelope.Data.Result.MainAgeMonth != 21 || !envelope.Data.Result.Complete || envelope.Data.Result.MeanMentalAgeMonths != 21 || envelope.Data.Result.DQ != 100 {
		t.Fatalf("unexpected score result: %+v", envelope.Data.Result)
	}
}

func TestERXinFormTemplateSummaryEndpointWithGeneratedDraftsWhenPresent(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..")
	itemPath := filepath.Join(root, "docs", "erxin-item-bank-draft.json")
	if _, err := os.Stat(itemPath); err != nil {
		t.Skipf("generated ERXin draft data not present: %s", itemPath)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/assessments/erxin/form-template/summary", nil)
	rec := httptest.NewRecorder()

	tokenManager := authx.NewTokenManager("test-secret")
	token, err := tokenManager.Generate(authx.Claims{UserID: 1, Username: "tester", LoginType: "staff", TenantID: "default", OrgID: 1}, time.Hour)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	h := New(service.New(nil, nil, tokenManager, nil, nil, nil))
	tenant.Middleware(http.HandlerFunc(h.erxinAssessmentFormTemplateSummary)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status %d: %s", rec.Code, rec.Body.String())
	}
	var envelope struct {
		Success bool `json:"success"`
		Data    struct {
			TemplateCode string `json:"templateCode"`
			ScaleCode    string `json:"scaleCode"`
			ItemCount    int    `json:"itemCount"`
			AgeGroups    []struct {
				AgeMonth int `json:"ageMonth"`
				Items    []struct {
					ItemNo              int  `json:"itemNo"`
					ParentReportAllowed bool `json:"parentReportAllowed"`
				} `json:"items"`
			} `json:"ageGroups"`
			SubmitContract struct {
				ScoreEndpoint   string `json:"scoreEndpoint"`
				ItemPassListKey string `json:"itemPassListKey"`
			} `json:"submitContract"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !envelope.Success {
		t.Fatalf("expected success response: %+v", envelope)
	}
	if envelope.Data.TemplateCode != "ERXIN2_ASSESSMENT_FORM" || envelope.Data.ScaleCode != "ERXIN2" || envelope.Data.ItemCount != 261 {
		t.Fatalf("unexpected template info: %+v", envelope.Data)
	}
	if len(envelope.Data.AgeGroups) != len(erxinscore.StandardAgeMonths) || envelope.Data.AgeGroups[0].AgeMonth != 1 || len(envelope.Data.AgeGroups[0].Items) == 0 {
		t.Fatalf("unexpected age groups: %+v", envelope.Data.AgeGroups)
	}
	if envelope.Data.SubmitContract.ScoreEndpoint != "/api/v1/assessments/erxin/score" || envelope.Data.SubmitContract.ItemPassListKey != "itemPassList" {
		t.Fatalf("unexpected submit contract: %+v", envelope.Data.SubmitContract)
	}
}
