package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"go-migration-platform/pkg/autismdevscore"
	"go-migration-platform/pkg/httpx"
	"go-migration-platform/pkg/tenant"
	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/service"
)

type autismDevScoreRequest struct {
	BirthDate                 string                      `json:"birthDate"`
	AssessmentDate            string                      `json:"assessmentDate"`
	QuestionDisplayPreference string                      `json:"questionDisplayPreference,omitempty"`
	ItemScores                map[int]string              `json:"itemScores,omitempty"`
	ItemScoreList             []autismDevItemScoreRequest `json:"itemScoreList,omitempty"`
}

type autismDevItemScoreRequest struct {
	ItemNo int    `json:"itemNo"`
	Score  string `json:"score"`
	Remark string `json:"remark,omitempty"`
}

type autismDevItemRemarkRequest struct {
	ItemNo int    `json:"itemNo"`
	Remark string `json:"remark"`
}

type autismDevAssessmentDraftSaveRequest struct {
	ID                        int64                        `json:"id,omitempty"`
	StudentID                 int64                        `json:"studentId,omitempty"`
	StudentName               string                       `json:"studentName,omitempty"`
	ExaminerName              string                       `json:"examinerName,omitempty"`
	Remark                    string                       `json:"remark,omitempty"`
	BirthDate                 string                       `json:"birthDate,omitempty"`
	AssessmentDate            string                       `json:"assessmentDate,omitempty"`
	ScopeMode                 string                       `json:"scopeMode,omitempty"`
	ScopeDomainCodes          []string                     `json:"scopeDomainCodes,omitempty"`
	QuestionDisplayPreference string                       `json:"questionDisplayPreference,omitempty"`
	ItemScores                map[int]string               `json:"itemScores,omitempty"`
	ItemScoreList             []autismDevItemScoreRequest  `json:"itemScoreList,omitempty"`
	ItemRemarks               map[int]string               `json:"itemRemarks,omitempty"`
	ItemRemarkList            []autismDevItemRemarkRequest `json:"itemRemarkList,omitempty"`
}

type autismDevAssessmentRecordCreateRequest struct {
	ID                        int64                        `json:"id,omitempty"`
	StudentID                 int64                        `json:"studentId,omitempty"`
	StudentName               string                       `json:"studentName,omitempty"`
	ExaminerName              string                       `json:"examinerName,omitempty"`
	Remark                    string                       `json:"remark,omitempty"`
	BirthDate                 string                       `json:"birthDate"`
	AssessmentDate            string                       `json:"assessmentDate"`
	ScopeMode                 string                       `json:"scopeMode,omitempty"`
	ScopeDomainCodes          []string                     `json:"scopeDomainCodes,omitempty"`
	QuestionDisplayPreference string                       `json:"questionDisplayPreference,omitempty"`
	ItemScores                map[int]string               `json:"itemScores,omitempty"`
	ItemScoreList             []autismDevItemScoreRequest  `json:"itemScoreList,omitempty"`
	ItemRemarks               map[int]string               `json:"itemRemarks,omitempty"`
	ItemRemarkList            []autismDevItemRemarkRequest `json:"itemRemarkList,omitempty"`
}

type autismDevAssessmentRecordConfigUpdateRequest struct {
	ID             int64  `json:"id"`
	ExaminerName   string `json:"examinerName"`
	AssessmentDate string `json:"assessmentDate"`
}

type autismDevAssessmentDraftItemSaveRequest struct {
	DraftID int64   `json:"draftId"`
	ItemNo  int     `json:"itemNo"`
	Score   *string `json:"score"`
	Remark  *string `json:"remark,omitempty"`
}

type autismDevAssessmentDeleteRequest struct {
	ID int64 `json:"id"`
}

func (handler *Handler) scoreAutismDev(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req autismDevScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	input, err := req.toAssessmentInput()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	result, err := handler.service.ScoreAutismDev(input)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) saveAutismDevAssessmentDraft(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req autismDevAssessmentDraftSaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	input, err := req.toDraftSaveInput()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	result, err := handler.service.SaveAutismDevAssessmentDraft(claims.UserID, input)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) saveAutismDevAssessmentDraftItem(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req autismDevAssessmentDraftItemSaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if req.DraftID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "draftId is required", ctx.RequestID)
		return
	}
	if req.ItemNo <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "itemNo is required", ctx.RequestID)
		return
	}
	if req.Score == nil || strings.TrimSpace(*req.Score) == "" {
		httpx.WriteError(w, http.StatusBadRequest, "score is required", ctx.RequestID)
		return
	}
	normalizedScore := normalizeAutismDevRequestScore(*req.Score)
	result, err := handler.service.SaveAutismDevAssessmentDraftItem(claims.UserID, service.AutismDevAssessmentDraftItemSaveInput{
		DraftID: req.DraftID,
		ItemNo:  req.ItemNo,
		Score:   &normalizedScore,
		Remark:  req.Remark,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) autismDevAssessmentDraftDetail(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.GetAutismDevAssessmentDraft(claims.UserID, id)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) autismDevAssessmentDraftsPage(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.PageAutismDevAssessmentDrafts(claims.UserID, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) deleteAutismDevAssessmentDraft(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req autismDevAssessmentDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if req.ID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id", ctx.RequestID)
		return
	}
	result, err := handler.service.DeleteAutismDevAssessmentDraft(claims.UserID, req.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) submitAutismDevAssessmentDraft(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req autismDevAssessmentDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if req.ID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id", ctx.RequestID)
		return
	}
	result, err := handler.service.SubmitAutismDevAssessmentDraft(claims.UserID, req.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) createAutismDevAssessmentRecord(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req autismDevAssessmentRecordCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	input, err := req.toRecordSaveInput()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	result, err := handler.service.CreateAutismDevAssessmentRecord(claims.UserID, input)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) updateAutismDevAssessmentRecord(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req autismDevAssessmentRecordCreateRequest
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
	result, err := handler.service.UpdateAutismDevAssessmentRecord(claims.UserID, req.ID, input)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) updateAutismDevAssessmentRecordConfig(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req autismDevAssessmentRecordConfigUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if req.ID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "id is required", ctx.RequestID)
		return
	}
	assessmentDate, err := parseERXinDate(req.AssessmentDate, "assessmentDate")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	result, err := handler.service.UpdateAutismDevAssessmentRecordConfig(claims.UserID, req.ID, service.AutismDevAssessmentRecordConfigInput{
		ExaminerName:   strings.TrimSpace(req.ExaminerName),
		AssessmentDate: assessmentDate,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) autismDevAssessmentRecordDetail(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.GetAutismDevAssessmentRecord(claims.UserID, id)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) autismDevAssessmentRecordResultAnalysis(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		id, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("id")), 10, 64)
		if err != nil || id <= 0 {
			httpx.WriteError(w, http.StatusBadRequest, "invalid id", ctx.RequestID)
			return
		}
		result, err := handler.service.GetAutismDevResultAnalysis(claims.UserID, id)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
	case http.MethodPost:
		var req model.AutismDevResultAnalysisSaveRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
			return
		}
		result, err := handler.service.SaveAutismDevResultAnalysis(r.Context(), claims.UserID, req.ID, req.Analysis)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
	default:
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
	}
}

func (handler *Handler) autismDevAssessmentRecordAssessmentInfoWord(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	id, err := decodeAutismDevAssessmentSituationExportID(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	fileName, contentType, content, err := handler.service.ExportAutismDevAssessmentSituationWord(claims.UserID, id)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+url.QueryEscape(fileName))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

func (handler *Handler) autismDevAssessmentRecordAssessmentInfoPDF(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	id, err := decodeAutismDevAssessmentSituationExportID(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	fileName, contentType, content, err := handler.service.ExportAutismDevAssessmentSituationPDF(claims.UserID, id)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/pdf"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "inline; filename*=UTF-8''"+url.QueryEscape(fileName))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

func decodeAutismDevAssessmentSituationExportID(r *http.Request) (int64, error) {
	if r.Method != http.MethodGet {
		return 0, errors.New("method not allowed")
	}
	id, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("id")), 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}
	return id, nil
}

func (handler *Handler) autismDevAssessmentRecordResultAnalysisWord(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	recordID, analysis, err := decodeAutismDevResultAnalysisExportRequest(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	fileName, contentType, content, err := handler.service.ExportAutismDevResultAnalysisWord(claims.UserID, recordID, analysis)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+url.QueryEscape(fileName))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

func (handler *Handler) autismDevAssessmentRecordResultAnalysisPDF(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	recordID, analysis, err := decodeAutismDevResultAnalysisExportRequest(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	fileName, contentType, content, err := handler.service.ExportAutismDevResultAnalysisPDF(claims.UserID, recordID, analysis)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/pdf"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "inline; filename*=UTF-8''"+url.QueryEscape(fileName))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

func decodeAutismDevResultAnalysisExportRequest(r *http.Request) (int64, *model.AutismDevResultAnalysisVO, error) {
	switch r.Method {
	case http.MethodGet:
		id, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("id")), 10, 64)
		if err != nil || id <= 0 {
			return 0, nil, errors.New("invalid id")
		}
		return id, nil, nil
	case http.MethodPost:
		var req model.AutismDevResultAnalysisExportRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return 0, nil, errors.New("invalid request body")
		}
		if req.ID <= 0 {
			return 0, nil, errors.New("invalid id")
		}
		return req.ID, req.Analysis, nil
	default:
		return 0, nil, errors.New("method not allowed")
	}
}

func (handler *Handler) autismDevAssessmentRecordResultAnalysisAIStream(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req model.AutismDevResultAnalysisGenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		httpx.WriteError(w, http.StatusInternalServerError, "streaming is not supported", ctx.RequestID)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	writeEvent := func(event string, payload any) error {
		return writePEP3IEPPlanSSE(w, flusher, event, payload)
	}
	if err := writeEvent("status", map[string]any{"type": "status", "message": "正在读取孤独症儿童发展评估结果"}); err != nil {
		return
	}
	result, err := handler.service.GenerateAutismDevResultAnalysisStream(r.Context(), claims.UserID, req.ID, func(text string) error {
		return writeEvent("delta", map[string]any{"type": "delta", "text": text})
	})
	if err != nil {
		_ = writeEvent("error", map[string]any{"type": "error", "message": err.Error()})
		return
	}
	_ = writeEvent("done", map[string]any{"type": "done", "data": result})
}

func (handler *Handler) autismDevAssessmentRecordProfilePDF(w http.ResponseWriter, r *http.Request) {
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
	profile := strings.TrimSpace(r.URL.Query().Get("profile"))
	filename, content, err := handler.service.GenerateAutismDevProfilePDF(claims.UserID, id, profile)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline; filename*=UTF-8''"+url.QueryEscape(filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

func (handler *Handler) autismDevAssessmentRecordSelectedReportPDF(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req model.AutismDevSelectedReportExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if req.ID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id", ctx.RequestID)
		return
	}
	fileName, contentType, content, err := handler.service.ExportAutismDevSelectedReportPDF(claims.UserID, req.ID, req.Sections, req.Analysis)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/pdf"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "inline; filename*=UTF-8''"+url.QueryEscape(fileName))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

func (handler *Handler) autismDevAssessmentRecordsPage(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.PageAutismDevAssessmentRecords(claims.UserID, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) autismDevAssessmentRecordCategoryStats(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.SummarizeAutismDevAssessmentRecordCategories(claims.UserID, query.QueryModel)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) deleteAutismDevAssessmentRecord(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req autismDevAssessmentDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if req.ID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id", ctx.RequestID)
		return
	}
	result, err := handler.service.DeleteAutismDevAssessmentRecord(claims.UserID, req.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) autismDevAssessmentFormTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	result, err := handler.service.GetAutismDevAssessmentFormTemplate()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) autismDevAssessmentFormTemplateSummary(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	result, err := handler.service.GetAutismDevAssessmentFormTemplateSummary()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) autismDevAssessmentFormTemplateItem(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	itemNo, err := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("itemNo")))
	if err != nil || itemNo <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid itemNo", ctx.RequestID)
		return
	}
	result, err := handler.service.GetAutismDevAssessmentFormTemplateItem(itemNo)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (req autismDevScoreRequest) toAssessmentInput() (autismdevscore.AssessmentInput, error) {
	birthDate, err := parseERXinDate(req.BirthDate, "birthDate")
	if err != nil {
		return autismdevscore.AssessmentInput{}, err
	}
	assessmentDate, err := parseERXinDate(req.AssessmentDate, "assessmentDate")
	if err != nil {
		return autismdevscore.AssessmentInput{}, err
	}
	itemScores, err := normalizeAutismDevItemScores(req.ItemScores, req.ItemScoreList)
	if err != nil {
		return autismdevscore.AssessmentInput{}, err
	}
	if len(itemScores) == 0 {
		return autismdevscore.AssessmentInput{}, fmt.Errorf("itemScores or itemScoreList is required")
	}
	return autismdevscore.AssessmentInput{
		BirthDate:                 birthDate,
		AssessmentDate:            assessmentDate,
		QuestionDisplayPreference: req.QuestionDisplayPreference,
		ItemScores:                itemScores,
	}, nil
}

func (req autismDevAssessmentDraftSaveRequest) toDraftSaveInput() (service.AutismDevAssessmentDraftSaveInput, error) {
	birthDate, err := parseOptionalERXinDate(req.BirthDate, "birthDate")
	if err != nil {
		return service.AutismDevAssessmentDraftSaveInput{}, err
	}
	assessmentDate, err := parseOptionalERXinDate(req.AssessmentDate, "assessmentDate")
	if err != nil {
		return service.AutismDevAssessmentDraftSaveInput{}, err
	}
	itemScores, err := normalizeAutismDevItemScores(req.ItemScores, req.ItemScoreList)
	if err != nil {
		return service.AutismDevAssessmentDraftSaveInput{}, err
	}
	itemRemarks := normalizeAutismDevItemRemarks(req.ItemRemarks, req.ItemRemarkList, req.ItemScoreList)
	return service.AutismDevAssessmentDraftSaveInput{
		ID:                        req.ID,
		StudentID:                 req.StudentID,
		StudentName:               strings.TrimSpace(req.StudentName),
		ExaminerName:              strings.TrimSpace(req.ExaminerName),
		Remark:                    strings.TrimSpace(req.Remark),
		BirthDate:                 birthDate,
		AssessmentDate:            assessmentDate,
		QuestionDisplayPreference: req.QuestionDisplayPreference,
		ItemScores:                itemScores,
		InputSnapshot:             req.normalizedSnapshot(itemScores, itemRemarks),
	}, nil
}

func (req autismDevAssessmentRecordCreateRequest) toRecordSaveInput() (service.AutismDevAssessmentRecordSaveInput, error) {
	scoreReq := req.toScoreRequest()
	scoreInput, err := scoreReq.toAssessmentInput()
	if err != nil {
		return service.AutismDevAssessmentRecordSaveInput{}, err
	}
	itemRemarks := normalizeAutismDevItemRemarks(req.ItemRemarks, req.ItemRemarkList, req.ItemScoreList)
	return service.AutismDevAssessmentRecordSaveInput{
		StudentID:     req.StudentID,
		StudentName:   strings.TrimSpace(req.StudentName),
		ExaminerName:  strings.TrimSpace(req.ExaminerName),
		Remark:        strings.TrimSpace(req.Remark),
		ScoreInput:    scoreInput,
		InputSnapshot: req.normalizedSnapshot(scoreInput.ItemScores, itemRemarks),
	}, nil
}

func (req autismDevAssessmentRecordCreateRequest) toScoreRequest() autismDevScoreRequest {
	return autismDevScoreRequest{
		BirthDate:                 req.BirthDate,
		AssessmentDate:            req.AssessmentDate,
		QuestionDisplayPreference: req.QuestionDisplayPreference,
		ItemScores:                req.ItemScores,
		ItemScoreList:             req.ItemScoreList,
	}
}

func (req autismDevAssessmentDraftSaveRequest) normalizedSnapshot(itemScores map[int]string, itemRemarks map[int]string) any {
	normalizedScoreList := autismDevItemScoreListFromMap(itemScores, itemRemarks)
	normalizedRemarkList := autismDevItemRemarkListFromMap(itemRemarks)
	return struct {
		ID                        int64                        `json:"id,omitempty"`
		StudentID                 int64                        `json:"studentId,omitempty"`
		StudentName               string                       `json:"studentName,omitempty"`
		ExaminerName              string                       `json:"examinerName,omitempty"`
		Remark                    string                       `json:"remark,omitempty"`
		BirthDate                 string                       `json:"birthDate,omitempty"`
		AssessmentDate            string                       `json:"assessmentDate,omitempty"`
		ScopeMode                 string                       `json:"scopeMode,omitempty"`
		ScopeDomainCodes          []string                     `json:"scopeDomainCodes,omitempty"`
		QuestionDisplayPreference string                       `json:"questionDisplayPreference,omitempty"`
		ItemScores                map[int]string               `json:"itemScores,omitempty"`
		ItemScoreList             []autismDevItemScoreRequest  `json:"itemScoreList,omitempty"`
		ItemRemarks               map[int]string               `json:"itemRemarks,omitempty"`
		ItemRemarkList            []autismDevItemRemarkRequest `json:"itemRemarkList,omitempty"`
	}{
		ID:                        req.ID,
		StudentID:                 req.StudentID,
		StudentName:               strings.TrimSpace(req.StudentName),
		ExaminerName:              strings.TrimSpace(req.ExaminerName),
		Remark:                    strings.TrimSpace(req.Remark),
		BirthDate:                 strings.TrimSpace(req.BirthDate),
		AssessmentDate:            strings.TrimSpace(req.AssessmentDate),
		ScopeMode:                 strings.TrimSpace(req.ScopeMode),
		ScopeDomainCodes:          normalizedAutismDevScopeDomainCodes(req.ScopeDomainCodes),
		QuestionDisplayPreference: autismdevscore.NormalizeQuestionDisplayPreference(req.QuestionDisplayPreference),
		ItemScores:                itemScores,
		ItemScoreList:             normalizedScoreList,
		ItemRemarks:               itemRemarks,
		ItemRemarkList:            normalizedRemarkList,
	}
}

func (req autismDevAssessmentRecordCreateRequest) normalizedSnapshot(itemScores map[int]string, itemRemarks map[int]string) any {
	normalizedScoreList := autismDevItemScoreListFromMap(itemScores, itemRemarks)
	normalizedRemarkList := autismDevItemRemarkListFromMap(itemRemarks)
	return struct {
		ID                        int64                        `json:"id,omitempty"`
		StudentID                 int64                        `json:"studentId,omitempty"`
		StudentName               string                       `json:"studentName,omitempty"`
		ExaminerName              string                       `json:"examinerName,omitempty"`
		Remark                    string                       `json:"remark,omitempty"`
		BirthDate                 string                       `json:"birthDate"`
		AssessmentDate            string                       `json:"assessmentDate"`
		ScopeMode                 string                       `json:"scopeMode,omitempty"`
		ScopeDomainCodes          []string                     `json:"scopeDomainCodes,omitempty"`
		QuestionDisplayPreference string                       `json:"questionDisplayPreference,omitempty"`
		ItemScores                map[int]string               `json:"itemScores,omitempty"`
		ItemScoreList             []autismDevItemScoreRequest  `json:"itemScoreList,omitempty"`
		ItemRemarks               map[int]string               `json:"itemRemarks,omitempty"`
		ItemRemarkList            []autismDevItemRemarkRequest `json:"itemRemarkList,omitempty"`
	}{
		ID:                        req.ID,
		StudentID:                 req.StudentID,
		StudentName:               strings.TrimSpace(req.StudentName),
		ExaminerName:              strings.TrimSpace(req.ExaminerName),
		Remark:                    strings.TrimSpace(req.Remark),
		BirthDate:                 strings.TrimSpace(req.BirthDate),
		AssessmentDate:            strings.TrimSpace(req.AssessmentDate),
		ScopeMode:                 strings.TrimSpace(req.ScopeMode),
		ScopeDomainCodes:          normalizedAutismDevScopeDomainCodes(req.ScopeDomainCodes),
		QuestionDisplayPreference: autismdevscore.NormalizeQuestionDisplayPreference(req.QuestionDisplayPreference),
		ItemScores:                itemScores,
		ItemScoreList:             normalizedScoreList,
		ItemRemarks:               itemRemarks,
		ItemRemarkList:            normalizedRemarkList,
	}
}

func normalizeAutismDevItemScores(itemScores map[int]string, itemScoreList []autismDevItemScoreRequest) (map[int]string, error) {
	normalized := make(map[int]string, len(itemScores)+len(itemScoreList))
	for itemNo, score := range itemScores {
		if itemNo <= 0 {
			return nil, fmt.Errorf("itemScores contains invalid itemNo %d", itemNo)
		}
		normalizedScore := normalizeAutismDevRequestScore(score)
		if normalizedScore == "" {
			return nil, fmt.Errorf("itemScores contains empty score for itemNo %d", itemNo)
		}
		normalized[itemNo] = normalizedScore
	}
	for _, item := range itemScoreList {
		if item.ItemNo <= 0 {
			return nil, fmt.Errorf("itemScoreList contains invalid itemNo %d", item.ItemNo)
		}
		normalizedScore := normalizeAutismDevRequestScore(item.Score)
		if normalizedScore == "" {
			return nil, fmt.Errorf("itemScoreList contains empty score for itemNo %d", item.ItemNo)
		}
		normalized[item.ItemNo] = normalizedScore
	}
	return normalized, nil
}

func normalizeAutismDevItemRemarks(itemRemarks map[int]string, itemRemarkList []autismDevItemRemarkRequest, itemScoreList []autismDevItemScoreRequest) map[int]string {
	out := make(map[int]string, len(itemRemarks)+len(itemRemarkList)+len(itemScoreList))
	for itemNo, remark := range itemRemarks {
		normalized := strings.TrimSpace(remark)
		if itemNo > 0 && normalized != "" {
			out[itemNo] = normalized
		}
	}
	for _, item := range itemRemarkList {
		normalized := strings.TrimSpace(item.Remark)
		if item.ItemNo > 0 && normalized != "" {
			out[item.ItemNo] = normalized
		}
	}
	for _, item := range itemScoreList {
		if item.ItemNo <= 0 {
			continue
		}
		remark := strings.TrimSpace(item.Remark)
		if remark == "" {
			continue
		}
		out[item.ItemNo] = remark
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func normalizedAutismDevScopeDomainCodes(domainCodes []string) []string {
	if len(domainCodes) == 0 {
		return nil
	}
	seen := make(map[string]bool, len(domainCodes))
	out := make([]string, 0, len(domainCodes))
	for _, raw := range domainCodes {
		code := strings.TrimSpace(raw)
		if code == "" || seen[code] {
			continue
		}
		seen[code] = true
		out = append(out, code)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func autismDevItemScoreListFromMap(itemScores map[int]string, itemRemarks map[int]string) []autismDevItemScoreRequest {
	if len(itemScores) == 0 {
		return nil
	}
	itemNos := make([]int, 0, len(itemScores))
	for itemNo := range itemScores {
		itemNos = append(itemNos, itemNo)
	}
	sort.Ints(itemNos)
	out := make([]autismDevItemScoreRequest, 0, len(itemNos))
	for _, itemNo := range itemNos {
		out = append(out, autismDevItemScoreRequest{
			ItemNo: itemNo,
			Score:  normalizeAutismDevRequestScore(itemScores[itemNo]),
			Remark: strings.TrimSpace(itemRemarks[itemNo]),
		})
	}
	return out
}

func autismDevItemRemarkListFromMap(itemRemarks map[int]string) []autismDevItemRemarkRequest {
	if len(itemRemarks) == 0 {
		return nil
	}
	itemNos := make([]int, 0, len(itemRemarks))
	for itemNo, remark := range itemRemarks {
		if itemNo <= 0 || strings.TrimSpace(remark) == "" {
			continue
		}
		itemNos = append(itemNos, itemNo)
	}
	if len(itemNos) == 0 {
		return nil
	}
	sort.Ints(itemNos)
	out := make([]autismDevItemRemarkRequest, 0, len(itemNos))
	for _, itemNo := range itemNos {
		remark := strings.TrimSpace(itemRemarks[itemNo])
		if remark == "" {
			continue
		}
		out = append(out, autismDevItemRemarkRequest{ItemNo: itemNo, Remark: remark})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func normalizeAutismDevRequestScore(score string) string {
	return strings.ToUpper(strings.TrimSpace(score))
}
