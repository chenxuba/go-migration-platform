package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"go-migration-platform/pkg/httpx"
	"go-migration-platform/pkg/tenant"
	"go-migration-platform/pkg/vbmappscore"
	"go-migration-platform/services/education/internal/service"
)

type vbmappScoreRequest struct {
	ScaleVersion string `json:"scaleVersion,omitempty"`

	MilestoneScores     map[string]float64             `json:"milestoneScores,omitempty"`
	MilestoneScoreList  []vbmappMilestoneScoreRequest  `json:"milestoneScoreList,omitempty"`
	BarrierScores       map[string]int                 `json:"barrierScores,omitempty"`
	BarrierScoreList    []vbmappBarrierScoreRequest    `json:"barrierScoreList,omitempty"`
	TransitionScores    map[string]int                 `json:"transitionScores,omitempty"`
	TransitionScoreList []vbmappTransitionScoreRequest `json:"transitionScoreList,omitempty"`

	PreviousMilestoneScores     map[string]float64             `json:"previousMilestoneScores,omitempty"`
	PreviousMilestoneScoreList  []vbmappMilestoneScoreRequest  `json:"previousMilestoneScoreList,omitempty"`
	PreviousBarrierScores       map[string]int                 `json:"previousBarrierScores,omitempty"`
	PreviousBarrierScoreList    []vbmappBarrierScoreRequest    `json:"previousBarrierScoreList,omitempty"`
	PreviousTransitionScores    map[string]int                 `json:"previousTransitionScores,omitempty"`
	PreviousTransitionScoreList []vbmappTransitionScoreRequest `json:"previousTransitionScoreList,omitempty"`

	ItemResponses map[string]map[string]map[string]any `json:"itemResponses,omitempty"`
}

type vbmappMilestoneScoreRequest struct {
	MilestoneID string  `json:"milestoneId"`
	Score       float64 `json:"score"`
}

type vbmappBarrierScoreRequest struct {
	BarrierCode string `json:"barrierCode"`
	Score       int    `json:"score"`
}

type vbmappTransitionScoreRequest struct {
	TransitionCode string `json:"transitionCode"`
	Score          int    `json:"score"`
}

type vbmappAssessmentDraftSaveRequest struct {
	ID             int64  `json:"id,omitempty"`
	StudentID      int64  `json:"studentId,omitempty"`
	StudentName    string `json:"studentName,omitempty"`
	ExaminerName   string `json:"examinerName,omitempty"`
	Remark         string `json:"remark,omitempty"`
	BirthDate      string `json:"birthDate,omitempty"`
	AssessmentDate string `json:"assessmentDate,omitempty"`
	ScaleVersion   string `json:"scaleVersion,omitempty"`

	MilestoneScores     map[string]float64             `json:"milestoneScores,omitempty"`
	MilestoneScoreList  []vbmappMilestoneScoreRequest  `json:"milestoneScoreList,omitempty"`
	BarrierScores       map[string]int                 `json:"barrierScores,omitempty"`
	BarrierScoreList    []vbmappBarrierScoreRequest    `json:"barrierScoreList,omitempty"`
	TransitionScores    map[string]int                 `json:"transitionScores,omitempty"`
	TransitionScoreList []vbmappTransitionScoreRequest `json:"transitionScoreList,omitempty"`

	PreviousMilestoneScores     map[string]float64             `json:"previousMilestoneScores,omitempty"`
	PreviousMilestoneScoreList  []vbmappMilestoneScoreRequest  `json:"previousMilestoneScoreList,omitempty"`
	PreviousBarrierScores       map[string]int                 `json:"previousBarrierScores,omitempty"`
	PreviousBarrierScoreList    []vbmappBarrierScoreRequest    `json:"previousBarrierScoreList,omitempty"`
	PreviousTransitionScores    map[string]int                 `json:"previousTransitionScores,omitempty"`
	PreviousTransitionScoreList []vbmappTransitionScoreRequest `json:"previousTransitionScoreList,omitempty"`

	ItemResponses map[string]map[string]map[string]any `json:"itemResponses,omitempty"`
}

type vbmappAssessmentDraftItemSaveRequest struct {
	DraftID          int64          `json:"draftId"`
	ModuleCode       string         `json:"moduleCode"`
	ItemCode         string         `json:"itemCode"`
	Score            *float64       `json:"score,omitempty"`
	SuggestedScore   *float64       `json:"suggestedScore,omitempty"`
	TeacherConfirmed *bool          `json:"teacherConfirmed,omitempty"`
	OverrideReason   string         `json:"overrideReason,omitempty"`
	RecordStatus     string         `json:"recordStatus,omitempty"`
	Evidence         map[string]any `json:"evidence,omitempty"`
}

type vbmappAssessmentDraftDeleteRequest struct {
	ID int64 `json:"id"`
}

type vbmappAssessmentRecordCreateRequest struct {
	ID             int64  `json:"id,omitempty"`
	StudentID      int64  `json:"studentId,omitempty"`
	StudentName    string `json:"studentName,omitempty"`
	ExaminerName   string `json:"examinerName,omitempty"`
	Remark         string `json:"remark,omitempty"`
	BirthDate      string `json:"birthDate"`
	AssessmentDate string `json:"assessmentDate"`
	ScaleVersion   string `json:"scaleVersion,omitempty"`

	MilestoneScores     map[string]float64             `json:"milestoneScores,omitempty"`
	MilestoneScoreList  []vbmappMilestoneScoreRequest  `json:"milestoneScoreList,omitempty"`
	BarrierScores       map[string]int                 `json:"barrierScores,omitempty"`
	BarrierScoreList    []vbmappBarrierScoreRequest    `json:"barrierScoreList,omitempty"`
	TransitionScores    map[string]int                 `json:"transitionScores,omitempty"`
	TransitionScoreList []vbmappTransitionScoreRequest `json:"transitionScoreList,omitempty"`

	PreviousMilestoneScores     map[string]float64             `json:"previousMilestoneScores,omitempty"`
	PreviousMilestoneScoreList  []vbmappMilestoneScoreRequest  `json:"previousMilestoneScoreList,omitempty"`
	PreviousBarrierScores       map[string]int                 `json:"previousBarrierScores,omitempty"`
	PreviousBarrierScoreList    []vbmappBarrierScoreRequest    `json:"previousBarrierScoreList,omitempty"`
	PreviousTransitionScores    map[string]int                 `json:"previousTransitionScores,omitempty"`
	PreviousTransitionScoreList []vbmappTransitionScoreRequest `json:"previousTransitionScoreList,omitempty"`

	ItemResponses map[string]map[string]map[string]any `json:"itemResponses,omitempty"`
}

func (handler *Handler) scoreVBMAPP(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	var req vbmappScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	input, err := req.toAssessmentInput()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	result, err := handler.service.ScoreVBMAPP(input)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) vbmappAssessmentSchema(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	result, err := handler.service.VBMAPPAssessmentSchema()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) saveVBMAPPAssessmentDraft(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	var req vbmappAssessmentDraftSaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	input, err := req.toDraftSaveInput()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	result, err := handler.service.SaveVBMAPPAssessmentDraft(claims.UserID, input)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) saveVBMAPPAssessmentDraftItem(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	var req vbmappAssessmentDraftItemSaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if req.DraftID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "draftId is required", ctx.RequestID)
		return
	}
	if strings.TrimSpace(req.ModuleCode) == "" {
		httpx.WriteError(w, http.StatusBadRequest, "moduleCode is required", ctx.RequestID)
		return
	}
	if strings.TrimSpace(req.ItemCode) == "" {
		httpx.WriteError(w, http.StatusBadRequest, "itemCode is required", ctx.RequestID)
		return
	}
	if req.Score == nil && req.SuggestedScore == nil && req.TeacherConfirmed == nil && strings.TrimSpace(req.OverrideReason) == "" && strings.TrimSpace(req.RecordStatus) == "" && len(req.Evidence) == 0 {
		httpx.WriteError(w, http.StatusBadRequest, "score or evidence is required", ctx.RequestID)
		return
	}

	result, err := handler.service.SaveVBMAPPAssessmentDraftItem(claims.UserID, service.VBMAPPAssessmentDraftItemSaveInput{
		DraftID:          req.DraftID,
		ModuleCode:       strings.TrimSpace(req.ModuleCode),
		ItemCode:         strings.TrimSpace(req.ItemCode),
		Score:            req.Score,
		SuggestedScore:   req.SuggestedScore,
		TeacherConfirmed: req.TeacherConfirmed,
		OverrideReason:   strings.TrimSpace(req.OverrideReason),
		RecordStatus:     strings.TrimSpace(req.RecordStatus),
		Evidence:         req.Evidence,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) vbmappAssessmentDraftDetail(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.GetVBMAPPAssessmentDraft(claims.UserID, id)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) vbmappAssessmentDraftsPage(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	query, err := decodeAssessmentDraftPageQuery(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.PageVBMAPPAssessmentDrafts(claims.UserID, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) deleteVBMAPPAssessmentDraft(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req vbmappAssessmentDraftDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if req.ID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id", ctx.RequestID)
		return
	}
	result, err := handler.service.DeleteVBMAPPAssessmentDraft(claims.UserID, req.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) submitVBMAPPAssessmentDraft(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req vbmappAssessmentDraftDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if req.ID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id", ctx.RequestID)
		return
	}
	result, err := handler.service.SubmitVBMAPPAssessmentDraft(claims.UserID, req.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) createVBMAPPAssessmentRecord(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req vbmappAssessmentRecordCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	input, err := req.toRecordSaveInput()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	result, err := handler.service.CreateVBMAPPAssessmentRecord(claims.UserID, input)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) updateVBMAPPAssessmentRecord(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req vbmappAssessmentRecordCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if req.ID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "id is required", ctx.RequestID)
		return
	}
	input, err := req.toRecordSaveInput()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	result, err := handler.service.UpdateVBMAPPAssessmentRecord(claims.UserID, req.ID, input)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) vbmappAssessmentRecordDetail(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.GetVBMAPPAssessmentRecord(claims.UserID, id)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) vbmappAssessmentRecordsPage(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	query, err := decodeAssessmentRecordPageQuery(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.PageVBMAPPAssessmentRecords(claims.UserID, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) deleteVBMAPPAssessmentRecord(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req vbmappAssessmentDraftDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if req.ID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id", ctx.RequestID)
		return
	}
	result, err := handler.service.DeleteVBMAPPAssessmentRecord(claims.UserID, req.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) vbmappAssessmentRecordHistory(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	studentID, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("studentId")), 10, 64)
	if err != nil || studentID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid studentId", ctx.RequestID)
		return
	}
	limit := 5
	if rawLimit := strings.TrimSpace(r.URL.Query().Get("limit")); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil || parsed <= 0 {
			httpx.WriteError(w, http.StatusBadRequest, "invalid limit", ctx.RequestID)
			return
		}
		limit = parsed
	}
	result, err := handler.service.GetVBMAPPAssessmentHistory(claims.UserID, studentID, limit)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (req vbmappScoreRequest) toAssessmentInput() (vbmappscore.AssessmentInput, error) {
	milestoneScores, err := normalizeVBMAPPMilestoneScores(req.MilestoneScores, req.MilestoneScoreList, "milestoneScores", "milestoneScoreList")
	if err != nil {
		return vbmappscore.AssessmentInput{}, err
	}
	barrierScores, err := normalizeVBMAPPBarrierScores(req.BarrierScores, req.BarrierScoreList, "barrierScores", "barrierScoreList")
	if err != nil {
		return vbmappscore.AssessmentInput{}, err
	}
	transitionScores, err := normalizeVBMAPPTransitionScores(req.TransitionScores, req.TransitionScoreList, "transitionScores", "transitionScoreList")
	if err != nil {
		return vbmappscore.AssessmentInput{}, err
	}
	previousMilestoneScores, err := normalizeVBMAPPMilestoneScores(req.PreviousMilestoneScores, req.PreviousMilestoneScoreList, "previousMilestoneScores", "previousMilestoneScoreList")
	if err != nil {
		return vbmappscore.AssessmentInput{}, err
	}
	previousBarrierScores, err := normalizeVBMAPPBarrierScores(req.PreviousBarrierScores, req.PreviousBarrierScoreList, "previousBarrierScores", "previousBarrierScoreList")
	if err != nil {
		return vbmappscore.AssessmentInput{}, err
	}
	previousTransitionScores, err := normalizeVBMAPPTransitionScores(req.PreviousTransitionScores, req.PreviousTransitionScoreList, "previousTransitionScores", "previousTransitionScoreList")
	if err != nil {
		return vbmappscore.AssessmentInput{}, err
	}
	return vbmappscore.AssessmentInput{
		ScaleVersion:             strings.TrimSpace(req.ScaleVersion),
		MilestoneScores:          milestoneScores,
		BarrierScores:            barrierScores,
		TransitionScores:         transitionScores,
		ItemResponses:            normalizeVBMAPPItemResponses(req.ItemResponses),
		PreviousMilestoneScores:  previousMilestoneScores,
		PreviousBarrierScores:    previousBarrierScores,
		PreviousTransitionScores: previousTransitionScores,
	}, nil
}

func (req vbmappAssessmentDraftSaveRequest) toDraftSaveInput() (service.VBMAPPAssessmentDraftSaveInput, error) {
	birthDate, err := parseOptionalERXinDate(req.BirthDate, "birthDate")
	if err != nil {
		return service.VBMAPPAssessmentDraftSaveInput{}, err
	}
	assessmentDate, err := parseOptionalERXinDate(req.AssessmentDate, "assessmentDate")
	if err != nil {
		return service.VBMAPPAssessmentDraftSaveInput{}, err
	}
	scoreReq := req.toScoreRequest()
	scoreInput, err := scoreReq.toAssessmentInput()
	if err != nil {
		return service.VBMAPPAssessmentDraftSaveInput{}, err
	}
	return service.VBMAPPAssessmentDraftSaveInput{
		ID:             req.ID,
		StudentID:      req.StudentID,
		StudentName:    strings.TrimSpace(req.StudentName),
		ExaminerName:   strings.TrimSpace(req.ExaminerName),
		Remark:         strings.TrimSpace(req.Remark),
		BirthDate:      birthDate,
		AssessmentDate: assessmentDate,
		ScoreInput:     scoreInput,
		InputSnapshot:  req.normalizedSnapshot(scoreInput),
	}, nil
}

func (req vbmappAssessmentDraftSaveRequest) toScoreRequest() vbmappScoreRequest {
	return vbmappScoreRequest{
		ScaleVersion:                req.ScaleVersion,
		MilestoneScores:             req.MilestoneScores,
		MilestoneScoreList:          req.MilestoneScoreList,
		BarrierScores:               req.BarrierScores,
		BarrierScoreList:            req.BarrierScoreList,
		TransitionScores:            req.TransitionScores,
		TransitionScoreList:         req.TransitionScoreList,
		PreviousMilestoneScores:     req.PreviousMilestoneScores,
		PreviousMilestoneScoreList:  req.PreviousMilestoneScoreList,
		PreviousBarrierScores:       req.PreviousBarrierScores,
		PreviousBarrierScoreList:    req.PreviousBarrierScoreList,
		PreviousTransitionScores:    req.PreviousTransitionScores,
		PreviousTransitionScoreList: req.PreviousTransitionScoreList,
		ItemResponses:               req.ItemResponses,
	}
}

func (req vbmappAssessmentDraftSaveRequest) normalizedSnapshot(input vbmappscore.AssessmentInput) any {
	return struct {
		ID             int64  `json:"id,omitempty"`
		StudentID      int64  `json:"studentId,omitempty"`
		StudentName    string `json:"studentName,omitempty"`
		ExaminerName   string `json:"examinerName,omitempty"`
		Remark         string `json:"remark,omitempty"`
		BirthDate      string `json:"birthDate,omitempty"`
		AssessmentDate string `json:"assessmentDate,omitempty"`
		ScaleVersion   string `json:"scaleVersion,omitempty"`

		MilestoneScores     map[string]float64             `json:"milestoneScores,omitempty"`
		MilestoneScoreList  []vbmappMilestoneScoreRequest  `json:"milestoneScoreList,omitempty"`
		BarrierScores       map[string]int                 `json:"barrierScores,omitempty"`
		BarrierScoreList    []vbmappBarrierScoreRequest    `json:"barrierScoreList,omitempty"`
		TransitionScores    map[string]int                 `json:"transitionScores,omitempty"`
		TransitionScoreList []vbmappTransitionScoreRequest `json:"transitionScoreList,omitempty"`

		PreviousMilestoneScores     map[string]float64             `json:"previousMilestoneScores,omitempty"`
		PreviousMilestoneScoreList  []vbmappMilestoneScoreRequest  `json:"previousMilestoneScoreList,omitempty"`
		PreviousBarrierScores       map[string]int                 `json:"previousBarrierScores,omitempty"`
		PreviousBarrierScoreList    []vbmappBarrierScoreRequest    `json:"previousBarrierScoreList,omitempty"`
		PreviousTransitionScores    map[string]int                 `json:"previousTransitionScores,omitempty"`
		PreviousTransitionScoreList []vbmappTransitionScoreRequest `json:"previousTransitionScoreList,omitempty"`

		ItemResponses map[string]map[string]map[string]any `json:"itemResponses,omitempty"`
	}{
		ID:                          req.ID,
		StudentID:                   req.StudentID,
		StudentName:                 strings.TrimSpace(req.StudentName),
		ExaminerName:                strings.TrimSpace(req.ExaminerName),
		Remark:                      strings.TrimSpace(req.Remark),
		BirthDate:                   strings.TrimSpace(req.BirthDate),
		AssessmentDate:              strings.TrimSpace(req.AssessmentDate),
		ScaleVersion:                strings.TrimSpace(input.ScaleVersion),
		MilestoneScores:             input.MilestoneScores,
		MilestoneScoreList:          vbmappMilestoneScoreListFromMap(input.MilestoneScores),
		BarrierScores:               input.BarrierScores,
		BarrierScoreList:            vbmappBarrierScoreListFromMap(input.BarrierScores),
		TransitionScores:            input.TransitionScores,
		TransitionScoreList:         vbmappTransitionScoreListFromMap(input.TransitionScores),
		PreviousMilestoneScores:     input.PreviousMilestoneScores,
		PreviousMilestoneScoreList:  vbmappMilestoneScoreListFromMap(input.PreviousMilestoneScores),
		PreviousBarrierScores:       input.PreviousBarrierScores,
		PreviousBarrierScoreList:    vbmappBarrierScoreListFromMap(input.PreviousBarrierScores),
		PreviousTransitionScores:    input.PreviousTransitionScores,
		PreviousTransitionScoreList: vbmappTransitionScoreListFromMap(input.PreviousTransitionScores),
		ItemResponses:               normalizeVBMAPPItemResponses(req.ItemResponses),
	}
}

func (req vbmappAssessmentRecordCreateRequest) toRecordSaveInput() (service.VBMAPPAssessmentRecordSaveInput, error) {
	birthDate, err := parseERXinDate(req.BirthDate, "birthDate")
	if err != nil {
		return service.VBMAPPAssessmentRecordSaveInput{}, err
	}
	assessmentDate, err := parseERXinDate(req.AssessmentDate, "assessmentDate")
	if err != nil {
		return service.VBMAPPAssessmentRecordSaveInput{}, err
	}
	scoreReq := req.toScoreRequest()
	scoreInput, err := scoreReq.toAssessmentInput()
	if err != nil {
		return service.VBMAPPAssessmentRecordSaveInput{}, err
	}
	return service.VBMAPPAssessmentRecordSaveInput{
		StudentID:      req.StudentID,
		StudentName:    strings.TrimSpace(req.StudentName),
		ExaminerName:   strings.TrimSpace(req.ExaminerName),
		Remark:         strings.TrimSpace(req.Remark),
		BirthDate:      birthDate,
		AssessmentDate: assessmentDate,
		ScoreInput:     scoreInput,
		InputSnapshot:  req.normalizedSnapshot(scoreInput),
	}, nil
}

func (req vbmappAssessmentRecordCreateRequest) toScoreRequest() vbmappScoreRequest {
	return vbmappScoreRequest{
		ScaleVersion:                req.ScaleVersion,
		MilestoneScores:             req.MilestoneScores,
		MilestoneScoreList:          req.MilestoneScoreList,
		BarrierScores:               req.BarrierScores,
		BarrierScoreList:            req.BarrierScoreList,
		TransitionScores:            req.TransitionScores,
		TransitionScoreList:         req.TransitionScoreList,
		PreviousMilestoneScores:     req.PreviousMilestoneScores,
		PreviousMilestoneScoreList:  req.PreviousMilestoneScoreList,
		PreviousBarrierScores:       req.PreviousBarrierScores,
		PreviousBarrierScoreList:    req.PreviousBarrierScoreList,
		PreviousTransitionScores:    req.PreviousTransitionScores,
		PreviousTransitionScoreList: req.PreviousTransitionScoreList,
		ItemResponses:               req.ItemResponses,
	}
}

func (req vbmappAssessmentRecordCreateRequest) normalizedSnapshot(input vbmappscore.AssessmentInput) any {
	return vbmappAssessmentDraftSaveRequest{
		ID:                          req.ID,
		StudentID:                   req.StudentID,
		StudentName:                 strings.TrimSpace(req.StudentName),
		ExaminerName:                strings.TrimSpace(req.ExaminerName),
		Remark:                      strings.TrimSpace(req.Remark),
		BirthDate:                   strings.TrimSpace(req.BirthDate),
		AssessmentDate:              strings.TrimSpace(req.AssessmentDate),
		ScaleVersion:                strings.TrimSpace(input.ScaleVersion),
		MilestoneScores:             input.MilestoneScores,
		MilestoneScoreList:          vbmappMilestoneScoreListFromMap(input.MilestoneScores),
		BarrierScores:               input.BarrierScores,
		BarrierScoreList:            vbmappBarrierScoreListFromMap(input.BarrierScores),
		TransitionScores:            input.TransitionScores,
		TransitionScoreList:         vbmappTransitionScoreListFromMap(input.TransitionScores),
		PreviousMilestoneScores:     input.PreviousMilestoneScores,
		PreviousMilestoneScoreList:  vbmappMilestoneScoreListFromMap(input.PreviousMilestoneScores),
		PreviousBarrierScores:       input.PreviousBarrierScores,
		PreviousBarrierScoreList:    vbmappBarrierScoreListFromMap(input.PreviousBarrierScores),
		PreviousTransitionScores:    input.PreviousTransitionScores,
		PreviousTransitionScoreList: vbmappTransitionScoreListFromMap(input.PreviousTransitionScores),
		ItemResponses:               normalizeVBMAPPItemResponses(req.ItemResponses),
	}
}

func normalizeVBMAPPMilestoneScores(scores map[string]float64, list []vbmappMilestoneScoreRequest, mapField, listField string) (map[string]float64, error) {
	normalized := make(map[string]float64, len(scores)+len(list))
	for milestoneID, score := range scores {
		id := normalizeVBMAPPCode(milestoneID)
		if id == "" {
			return nil, fmt.Errorf("%s contains empty milestoneId", mapField)
		}
		normalized[id] = score
	}
	for _, item := range list {
		id := normalizeVBMAPPCode(item.MilestoneID)
		if id == "" {
			return nil, fmt.Errorf("%s contains empty milestoneId", listField)
		}
		normalized[id] = item.Score
	}
	return normalized, nil
}

func normalizeVBMAPPBarrierScores(scores map[string]int, list []vbmappBarrierScoreRequest, mapField, listField string) (map[string]int, error) {
	normalized := make(map[string]int, len(scores)+len(list))
	for barrierCode, score := range scores {
		code := normalizeVBMAPPCode(barrierCode)
		if code == "" {
			return nil, fmt.Errorf("%s contains empty barrierCode", mapField)
		}
		normalized[code] = score
	}
	for _, item := range list {
		code := normalizeVBMAPPCode(item.BarrierCode)
		if code == "" {
			return nil, fmt.Errorf("%s contains empty barrierCode", listField)
		}
		normalized[code] = item.Score
	}
	return normalized, nil
}

func normalizeVBMAPPTransitionScores(scores map[string]int, list []vbmappTransitionScoreRequest, mapField, listField string) (map[string]int, error) {
	normalized := make(map[string]int, len(scores)+len(list))
	for transitionCode, score := range scores {
		code := normalizeVBMAPPCode(transitionCode)
		if code == "" {
			return nil, fmt.Errorf("%s contains empty transitionCode", mapField)
		}
		normalized[code] = score
	}
	for _, item := range list {
		code := normalizeVBMAPPCode(item.TransitionCode)
		if code == "" {
			return nil, fmt.Errorf("%s contains empty transitionCode", listField)
		}
		normalized[code] = item.Score
	}
	return normalized, nil
}

func normalizeVBMAPPItemResponses(input map[string]map[string]map[string]any) map[string]map[string]map[string]any {
	if len(input) == 0 {
		return nil
	}
	normalized := make(map[string]map[string]map[string]any, len(input))
	for moduleCode, moduleItems := range input {
		module := normalizeVBMAPPModuleCode(moduleCode)
		if module == "" || len(moduleItems) == 0 {
			continue
		}
		for itemCode, response := range moduleItems {
			code := normalizeVBMAPPCode(itemCode)
			if code == "" || len(response) == 0 {
				continue
			}
			if normalized[module] == nil {
				normalized[module] = map[string]map[string]any{}
			}
			itemResponse := make(map[string]any, len(response)+2)
			for key, value := range response {
				itemResponse[key] = value
			}
			itemResponse["moduleCode"] = module
			itemResponse["itemCode"] = code
			normalized[module][code] = itemResponse
		}
	}
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

func normalizeVBMAPPModuleCode(moduleCode string) string {
	switch strings.ToLower(strings.TrimSpace(moduleCode)) {
	case "milestone", "milestones", "里程碑", "里程碑评估":
		return vbmappscore.ModuleMilestones
	case "barrier", "barriers", "障碍", "障碍评估":
		return vbmappscore.ModuleBarriers
	case "transition", "transitions", "转衔", "转衔评估", "转线", "转线评估", "转型", "转型评估":
		return vbmappscore.ModuleTransition
	default:
		return strings.ToLower(strings.TrimSpace(moduleCode))
	}
}

func normalizeVBMAPPCode(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func vbmappMilestoneScoreListFromMap(scores map[string]float64) []vbmappMilestoneScoreRequest {
	if len(scores) == 0 {
		return nil
	}
	codes := make([]string, 0, len(scores))
	for code := range scores {
		if normalizedCode := normalizeVBMAPPCode(code); normalizedCode != "" {
			codes = append(codes, normalizedCode)
		}
	}
	sort.Strings(codes)
	out := make([]vbmappMilestoneScoreRequest, 0, len(codes))
	for _, code := range codes {
		out = append(out, vbmappMilestoneScoreRequest{MilestoneID: code, Score: scores[code]})
	}
	return out
}

func vbmappBarrierScoreListFromMap(scores map[string]int) []vbmappBarrierScoreRequest {
	if len(scores) == 0 {
		return nil
	}
	codes := make([]string, 0, len(scores))
	for code := range scores {
		if normalizedCode := normalizeVBMAPPCode(code); normalizedCode != "" {
			codes = append(codes, normalizedCode)
		}
	}
	sort.Strings(codes)
	out := make([]vbmappBarrierScoreRequest, 0, len(codes))
	for _, code := range codes {
		out = append(out, vbmappBarrierScoreRequest{BarrierCode: code, Score: scores[code]})
	}
	return out
}

func vbmappTransitionScoreListFromMap(scores map[string]int) []vbmappTransitionScoreRequest {
	if len(scores) == 0 {
		return nil
	}
	codes := make([]string, 0, len(scores))
	for code := range scores {
		if normalizedCode := normalizeVBMAPPCode(code); normalizedCode != "" {
			codes = append(codes, normalizedCode)
		}
	}
	sort.Strings(codes)
	out := make([]vbmappTransitionScoreRequest, 0, len(codes))
	for _, code := range codes {
		out = append(out, vbmappTransitionScoreRequest{TransitionCode: code, Score: scores[code]})
	}
	return out
}
