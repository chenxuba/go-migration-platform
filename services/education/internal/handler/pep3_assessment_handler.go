package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/pkg/httpx"
	"go-migration-platform/pkg/pep3score"
	"go-migration-platform/pkg/tenant"
	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/service"
)

type pep3ScoreRequest struct {
	BirthDate         string                 `json:"birthDate"`
	AssessmentDate    string                 `json:"assessmentDate"`
	ItemScores        map[int]int            `json:"itemScores,omitempty"`
	ItemScoreList     []pep3ItemScoreRequest `json:"itemScoreList,omitempty"`
	RawScores         map[string]int         `json:"rawScores,omitempty"`
	RawScoreList      []pep3RawScoreRequest  `json:"rawScoreList,omitempty"`
	AllowMissingItems bool                   `json:"allowMissingItems,omitempty"`
}

type pep3ItemScoreRequest struct {
	ItemNo int `json:"itemNo"`
	Score  int `json:"score"`
}

type pep3RawScoreRequest struct {
	ScaleCode string `json:"scaleCode"`
	RawScore  int    `json:"rawScore"`
}

type pep3AssessmentRecordCreateRequest struct {
	StudentID         int64                  `json:"studentId,omitempty"`
	StudentName       string                 `json:"studentName,omitempty"`
	ExaminerName      string                 `json:"examinerName,omitempty"`
	Remark            string                 `json:"remark,omitempty"`
	BirthDate         string                 `json:"birthDate"`
	AssessmentDate    string                 `json:"assessmentDate"`
	ItemScores        map[int]int            `json:"itemScores,omitempty"`
	ItemScoreList     []pep3ItemScoreRequest `json:"itemScoreList,omitempty"`
	RawScores         map[string]int         `json:"rawScores,omitempty"`
	RawScoreList      []pep3RawScoreRequest  `json:"rawScoreList,omitempty"`
	AllowMissingItems bool                   `json:"allowMissingItems,omitempty"`
}

func (handler *Handler) scorePEP3(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	var req pep3ScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	input, err := req.toAssessmentInput()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	result, err := handler.service.ScorePEP3(input)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) createPEP3AssessmentRecord(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	var req pep3AssessmentRecordCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	scoreReq := req.toScoreRequest()
	scoreInput, err := scoreReq.toAssessmentInput()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	result, err := handler.service.CreatePEP3AssessmentRecord(claims.UserID, service.PEP3AssessmentRecordSaveInput{
		StudentID:     req.StudentID,
		StudentName:   strings.TrimSpace(req.StudentName),
		ExaminerName:  strings.TrimSpace(req.ExaminerName),
		Remark:        strings.TrimSpace(req.Remark),
		ScoreInput:    scoreInput,
		InputSnapshot: req,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) pep3AssessmentRecordDetail(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	id, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("id")), 10, 64)
	if err != nil || id <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id", ctx.RequestID)
		return
	}
	result, err := handler.service.GetPEP3AssessmentRecord(claims.UserID, id)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) pep3AssessmentRecordReport(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	id, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("id")), 10, 64)
	if err != nil || id <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id", ctx.RequestID)
		return
	}
	result, err := handler.service.GetPEP3AssessmentReport(claims.UserID, id)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) pep3AssessmentRecordsPage(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var query model.AssessmentRecordPageQueryDTO
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.PagePEP3AssessmentRecords(claims.UserID, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (req pep3AssessmentRecordCreateRequest) toScoreRequest() pep3ScoreRequest {
	return pep3ScoreRequest{
		BirthDate:         req.BirthDate,
		AssessmentDate:    req.AssessmentDate,
		ItemScores:        req.ItemScores,
		ItemScoreList:     req.ItemScoreList,
		RawScores:         req.RawScores,
		RawScoreList:      req.RawScoreList,
		AllowMissingItems: req.AllowMissingItems,
	}
}

func (req pep3ScoreRequest) toAssessmentInput() (pep3score.AssessmentInput, error) {
	birthDate, err := parsePEP3Date(req.BirthDate, "birthDate")
	if err != nil {
		return pep3score.AssessmentInput{}, err
	}
	assessmentDate, err := parsePEP3Date(req.AssessmentDate, "assessmentDate")
	if err != nil {
		return pep3score.AssessmentInput{}, err
	}

	itemScores := make(map[int]int, len(req.ItemScores)+len(req.ItemScoreList))
	for itemNo, score := range req.ItemScores {
		itemScores[itemNo] = score
	}
	for _, item := range req.ItemScoreList {
		if item.ItemNo <= 0 {
			return pep3score.AssessmentInput{}, fmt.Errorf("itemScoreList contains invalid itemNo %d", item.ItemNo)
		}
		itemScores[item.ItemNo] = item.Score
	}
	if len(itemScores) == 0 {
		itemScores = nil
	}

	rawScores := make(map[string]int, len(req.RawScores)+len(req.RawScoreList))
	for scaleCode, rawScore := range req.RawScores {
		normalized := normalizePEP3ScaleCode(scaleCode)
		if normalized == "" {
			return pep3score.AssessmentInput{}, fmt.Errorf("rawScores contains empty scale code")
		}
		rawScores[normalized] = rawScore
	}
	for _, item := range req.RawScoreList {
		normalized := normalizePEP3ScaleCode(item.ScaleCode)
		if normalized == "" {
			return pep3score.AssessmentInput{}, fmt.Errorf("rawScoreList contains empty scaleCode")
		}
		rawScores[normalized] = item.RawScore
	}
	if len(rawScores) == 0 {
		rawScores = nil
	}

	if len(itemScores) == 0 && len(rawScores) == 0 {
		return pep3score.AssessmentInput{}, fmt.Errorf("itemScores or rawScores is required")
	}

	return pep3score.AssessmentInput{
		BirthDate:         birthDate,
		AssessmentDate:    assessmentDate,
		ItemScores:        itemScores,
		RawScores:         rawScores,
		AllowMissingItems: req.AllowMissingItems,
	}, nil
}

func parsePEP3Date(raw, field string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, fmt.Errorf("%s is required", field)
	}
	for _, layout := range []string{"2006-01-02", "2006/01/02", time.RFC3339} {
		if parsed, err := time.ParseInLocation(layout, raw, time.Local); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("%s format should be YYYY-MM-DD", field)
}

func normalizePEP3ScaleCode(raw string) string {
	return strings.ToUpper(strings.TrimSpace(raw))
}
