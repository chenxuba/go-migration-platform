package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
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

type pep3AssessmentDraftSaveRequest struct {
	ID                int64                  `json:"id,omitempty"`
	StudentID         int64                  `json:"studentId,omitempty"`
	StudentName       string                 `json:"studentName,omitempty"`
	ExaminerName      string                 `json:"examinerName,omitempty"`
	Remark            string                 `json:"remark,omitempty"`
	BirthDate         string                 `json:"birthDate,omitempty"`
	AssessmentDate    string                 `json:"assessmentDate,omitempty"`
	ItemScores        map[int]int            `json:"itemScores,omitempty"`
	ItemScoreList     []pep3ItemScoreRequest `json:"itemScoreList,omitempty"`
	RawScores         map[string]int         `json:"rawScores,omitempty"`
	RawScoreList      []pep3RawScoreRequest  `json:"rawScoreList,omitempty"`
	AllowMissingItems bool                   `json:"allowMissingItems,omitempty"`
}

type pep3AssessmentDraftDeleteRequest struct {
	ID int64 `json:"id"`
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

func (handler *Handler) pep3AssessmentFormTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	result, err := handler.service.GetPEP3AssessmentFormTemplate()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) savePEP3AssessmentDraft(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	var req pep3AssessmentDraftSaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	input, err := req.toDraftSaveInput()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	result, err := handler.service.SavePEP3AssessmentDraft(claims.UserID, input)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) pep3AssessmentDraftDetail(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.GetPEP3AssessmentDraft(claims.UserID, id)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) pep3AssessmentDraftsPage(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var query model.AssessmentDraftPageQueryDTO
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.PagePEP3AssessmentDrafts(claims.UserID, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) deletePEP3AssessmentDraft(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req pep3AssessmentDraftDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if req.ID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id", ctx.RequestID)
		return
	}
	result, err := handler.service.DeletePEP3AssessmentDraft(claims.UserID, req.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) submitPEP3AssessmentDraft(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req pep3AssessmentDraftDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if req.ID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id", ctx.RequestID)
		return
	}
	result, err := handler.service.SubmitPEP3AssessmentDraft(claims.UserID, req.ID)
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

func (handler *Handler) pep3AssessmentRecordBooklet(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.GetPEP3AssessmentBooklet(claims.UserID, id)
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

func (req pep3AssessmentDraftSaveRequest) toDraftSaveInput() (service.PEP3AssessmentDraftSaveInput, error) {
	birthDate, err := parseOptionalPEP3Date(req.BirthDate, "birthDate")
	if err != nil {
		return service.PEP3AssessmentDraftSaveInput{}, err
	}
	assessmentDate, err := parseOptionalPEP3Date(req.AssessmentDate, "assessmentDate")
	if err != nil {
		return service.PEP3AssessmentDraftSaveInput{}, err
	}
	itemScores, err := normalizePEP3ItemScores(req.ItemScores, req.ItemScoreList)
	if err != nil {
		return service.PEP3AssessmentDraftSaveInput{}, err
	}
	rawScores, err := normalizePEP3RawScores(req.RawScores, req.RawScoreList)
	if err != nil {
		return service.PEP3AssessmentDraftSaveInput{}, err
	}
	return service.PEP3AssessmentDraftSaveInput{
		ID:                req.ID,
		StudentID:         req.StudentID,
		StudentName:       strings.TrimSpace(req.StudentName),
		ExaminerName:      strings.TrimSpace(req.ExaminerName),
		Remark:            strings.TrimSpace(req.Remark),
		BirthDate:         birthDate,
		AssessmentDate:    assessmentDate,
		ItemScores:        itemScores,
		RawScores:         rawScores,
		AllowMissingItems: req.AllowMissingItems,
		InputSnapshot:     req.normalizedSnapshot(itemScores, rawScores),
	}, nil
}

func (req pep3AssessmentDraftSaveRequest) normalizedSnapshot(itemScores map[int]int, rawScores map[string]int) any {
	return struct {
		ID                int64                  `json:"id,omitempty"`
		StudentID         int64                  `json:"studentId,omitempty"`
		StudentName       string                 `json:"studentName,omitempty"`
		ExaminerName      string                 `json:"examinerName,omitempty"`
		Remark            string                 `json:"remark,omitempty"`
		BirthDate         string                 `json:"birthDate,omitempty"`
		AssessmentDate    string                 `json:"assessmentDate,omitempty"`
		ItemScores        map[int]int            `json:"itemScores,omitempty"`
		ItemScoreList     []pep3ItemScoreRequest `json:"itemScoreList,omitempty"`
		RawScores         map[string]int         `json:"rawScores,omitempty"`
		RawScoreList      []pep3RawScoreRequest  `json:"rawScoreList,omitempty"`
		AllowMissingItems bool                   `json:"allowMissingItems,omitempty"`
	}{
		ID:                req.ID,
		StudentID:         req.StudentID,
		StudentName:       strings.TrimSpace(req.StudentName),
		ExaminerName:      strings.TrimSpace(req.ExaminerName),
		Remark:            strings.TrimSpace(req.Remark),
		BirthDate:         strings.TrimSpace(req.BirthDate),
		AssessmentDate:    strings.TrimSpace(req.AssessmentDate),
		ItemScores:        itemScores,
		ItemScoreList:     itemScoreListFromMap(itemScores),
		RawScores:         rawScores,
		RawScoreList:      rawScoreListFromMap(rawScores),
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

	itemScores, err := normalizePEP3ItemScores(req.ItemScores, req.ItemScoreList)
	if err != nil {
		return pep3score.AssessmentInput{}, err
	}
	if len(itemScores) == 0 {
		itemScores = nil
	}

	rawScores, err := normalizePEP3RawScores(req.RawScores, req.RawScoreList)
	if err != nil {
		return pep3score.AssessmentInput{}, err
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

func parseOptionalPEP3Date(raw, field string) (*time.Time, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	parsed, err := parsePEP3Date(raw, field)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func normalizePEP3ItemScores(itemScores map[int]int, itemScoreList []pep3ItemScoreRequest) (map[int]int, error) {
	normalized := make(map[int]int, len(itemScores)+len(itemScoreList))
	for itemNo, score := range itemScores {
		if itemNo <= 0 {
			return nil, fmt.Errorf("itemScores contains invalid itemNo %d", itemNo)
		}
		normalized[itemNo] = score
	}
	for _, item := range itemScoreList {
		if item.ItemNo <= 0 {
			return nil, fmt.Errorf("itemScoreList contains invalid itemNo %d", item.ItemNo)
		}
		normalized[item.ItemNo] = item.Score
	}
	return normalized, nil
}

func normalizePEP3RawScores(rawScores map[string]int, rawScoreList []pep3RawScoreRequest) (map[string]int, error) {
	normalizedScores := make(map[string]int, len(rawScores)+len(rawScoreList))
	for scaleCode, rawScore := range rawScores {
		normalized := normalizePEP3ScaleCode(scaleCode)
		if normalized == "" {
			return nil, fmt.Errorf("rawScores contains empty scale code")
		}
		normalizedScores[normalized] = rawScore
	}
	for _, item := range rawScoreList {
		normalized := normalizePEP3ScaleCode(item.ScaleCode)
		if normalized == "" {
			return nil, fmt.Errorf("rawScoreList contains empty scaleCode")
		}
		normalizedScores[normalized] = item.RawScore
	}
	return normalizedScores, nil
}

func itemScoreListFromMap(itemScores map[int]int) []pep3ItemScoreRequest {
	if len(itemScores) == 0 {
		return nil
	}
	itemNos := make([]int, 0, len(itemScores))
	for itemNo := range itemScores {
		itemNos = append(itemNos, itemNo)
	}
	sort.Ints(itemNos)
	out := make([]pep3ItemScoreRequest, 0, len(itemNos))
	for _, itemNo := range itemNos {
		out = append(out, pep3ItemScoreRequest{ItemNo: itemNo, Score: itemScores[itemNo]})
	}
	return out
}

func rawScoreListFromMap(rawScores map[string]int) []pep3RawScoreRequest {
	if len(rawScores) == 0 {
		return nil
	}
	scaleCodes := make([]string, 0, len(rawScores))
	for scaleCode := range rawScores {
		scaleCodes = append(scaleCodes, scaleCode)
	}
	sort.Strings(scaleCodes)
	out := make([]pep3RawScoreRequest, 0, len(scaleCodes))
	for _, scaleCode := range scaleCodes {
		out = append(out, pep3RawScoreRequest{ScaleCode: scaleCode, RawScore: rawScores[scaleCode]})
	}
	return out
}

func normalizePEP3ScaleCode(raw string) string {
	return strings.ToUpper(strings.TrimSpace(raw))
}
