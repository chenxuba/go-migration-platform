package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/pkg/erxinscore"
	"go-migration-platform/pkg/httpx"
	"go-migration-platform/pkg/tenant"
	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/service"
)

type erxinScoreRequest struct {
	BirthDate      string                 `json:"birthDate"`
	AssessmentDate string                 `json:"assessmentDate"`
	ItemPasses     map[int]bool           `json:"itemPasses,omitempty"`
	ItemPassList   []erxinItemPassRequest `json:"itemPassList,omitempty"`
}

type erxinItemPassRequest struct {
	ItemNo int    `json:"itemNo"`
	Passed bool   `json:"passed"`
	Remark string `json:"remark,omitempty"`
}

type erxinItemRemarkRequest struct {
	ItemNo int    `json:"itemNo"`
	Remark string `json:"remark"`
}

type erxinAssessmentDraftSaveRequest struct {
	ID             int64                    `json:"id,omitempty"`
	StudentID      int64                    `json:"studentId,omitempty"`
	StudentName    string                   `json:"studentName,omitempty"`
	ExaminerName   string                   `json:"examinerName,omitempty"`
	Remark         string                   `json:"remark,omitempty"`
	BirthDate      string                   `json:"birthDate,omitempty"`
	AssessmentDate string                   `json:"assessmentDate,omitempty"`
	ItemPasses     map[int]bool             `json:"itemPasses,omitempty"`
	ItemPassList   []erxinItemPassRequest   `json:"itemPassList,omitempty"`
	ItemRemarks    map[int]string           `json:"itemRemarks,omitempty"`
	ItemRemarkList []erxinItemRemarkRequest `json:"itemRemarkList,omitempty"`
}

type erxinAssessmentRecordCreateRequest struct {
	ID             int64                    `json:"id,omitempty"`
	StudentID      int64                    `json:"studentId,omitempty"`
	StudentName    string                   `json:"studentName,omitempty"`
	ExaminerName   string                   `json:"examinerName,omitempty"`
	Remark         string                   `json:"remark,omitempty"`
	BirthDate      string                   `json:"birthDate"`
	AssessmentDate string                   `json:"assessmentDate"`
	ItemPasses     map[int]bool             `json:"itemPasses,omitempty"`
	ItemPassList   []erxinItemPassRequest   `json:"itemPassList,omitempty"`
	ItemRemarks    map[int]string           `json:"itemRemarks,omitempty"`
	ItemRemarkList []erxinItemRemarkRequest `json:"itemRemarkList,omitempty"`
}

type erxinAssessmentRecordConfigUpdateRequest struct {
	ID             int64  `json:"id"`
	ExaminerName   string `json:"examinerName"`
	AssessmentDate string `json:"assessmentDate"`
}

type erxinAssessmentDraftItemSaveRequest struct {
	DraftID int64   `json:"draftId"`
	ItemNo  int     `json:"itemNo"`
	Passed  *bool   `json:"passed"`
	Remark  *string `json:"remark,omitempty"`
}

type erxinAssessmentDraftDeleteRequest struct {
	ID int64 `json:"id"`
}

func (handler *Handler) scoreERXin(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	var req erxinScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	input, err := req.toAssessmentInput()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	result, err := handler.service.ScoreERXin(input)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) saveERXinAssessmentDraft(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	var req erxinAssessmentDraftSaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	input, err := req.toDraftSaveInput()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	result, err := handler.service.SaveERXinAssessmentDraft(claims.UserID, input)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) saveERXinAssessmentDraftItem(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	var req erxinAssessmentDraftItemSaveRequest
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
	if req.Passed == nil {
		httpx.WriteError(w, http.StatusBadRequest, "passed is required", ctx.RequestID)
		return
	}

	result, err := handler.service.SaveERXinAssessmentDraftItem(claims.UserID, service.ERXinAssessmentDraftItemSaveInput{
		DraftID: req.DraftID,
		ItemNo:  req.ItemNo,
		Passed:  req.Passed,
		Remark:  req.Remark,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) erxinAssessmentDraftDetail(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.GetERXinAssessmentDraft(claims.UserID, id)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) erxinAssessmentDraftsPage(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.PageERXinAssessmentDrafts(claims.UserID, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) deleteERXinAssessmentDraft(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req erxinAssessmentDraftDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if req.ID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id", ctx.RequestID)
		return
	}
	result, err := handler.service.DeleteERXinAssessmentDraft(claims.UserID, req.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) submitERXinAssessmentDraft(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req erxinAssessmentDraftDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if req.ID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id", ctx.RequestID)
		return
	}
	result, err := handler.service.SubmitERXinAssessmentDraft(claims.UserID, req.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) createERXinAssessmentRecord(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req erxinAssessmentRecordCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	input, err := req.toRecordSaveInput()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	result, err := handler.service.CreateERXinAssessmentRecord(claims.UserID, input)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) updateERXinAssessmentRecord(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req erxinAssessmentRecordCreateRequest
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
	result, err := handler.service.UpdateERXinAssessmentRecord(claims.UserID, req.ID, input)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) updateERXinAssessmentRecordConfig(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req erxinAssessmentRecordConfigUpdateRequest
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
	result, err := handler.service.UpdateERXinAssessmentRecordConfig(claims.UserID, req.ID, service.ERXinAssessmentRecordConfigInput{
		ExaminerName:   strings.TrimSpace(req.ExaminerName),
		AssessmentDate: assessmentDate,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) erxinAssessmentRecordDetail(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.GetERXinAssessmentRecord(claims.UserID, id)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) erxinAssessmentRecordReport(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.GetERXinAssessmentReport(claims.UserID, id)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) erxinAssessmentRecordReportPDF(w http.ResponseWriter, r *http.Request) {
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
	filename, content, err := handler.service.GenerateERXinAssessmentReportPDF(claims.UserID, id)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline; filename*=UTF-8''"+url.QueryEscape(filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

func (handler *Handler) erxinAssessmentRecordReportInterpretation(w http.ResponseWriter, r *http.Request) {
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
		httpx.WriteError(w, http.StatusBadRequest, "invalid assessment record id", ctx.RequestID)
		return
	}
	result, err := handler.service.GetERXinReportInterpretation(claims.UserID, id)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) erxinAssessmentRecordReportInterpretationPDF(w http.ResponseWriter, r *http.Request) {
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
		httpx.WriteError(w, http.StatusBadRequest, "invalid assessment record id", ctx.RequestID)
		return
	}
	filename, content, err := handler.service.GenerateERXinReportInterpretationPDF(claims.UserID, id)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline; filename*=UTF-8''"+url.QueryEscape(filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

func (handler *Handler) erxinAssessmentRecordReportCombinedPDF(w http.ResponseWriter, r *http.Request) {
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
		httpx.WriteError(w, http.StatusBadRequest, "invalid assessment record id", ctx.RequestID)
		return
	}
	filename, content, err := handler.service.GenerateERXinCombinedReportPDF(claims.UserID, id)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline; filename*=UTF-8''"+url.QueryEscape(filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

func (handler *Handler) erxinAssessmentRecordReportInterpretationAI(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req model.ERXinReportInterpretationGenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.GenerateERXinReportInterpretation(r.Context(), claims.UserID, req.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) erxinAssessmentRecordReportInterpretationAIStream(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req model.ERXinReportInterpretationGenerateRequest
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
	if err := writeEvent("status", map[string]any{"type": "status", "message": "正在读取儿心评估结果"}); err != nil {
		return
	}
	result, err := handler.service.GenerateERXinReportInterpretationStream(r.Context(), claims.UserID, req.ID, func(text string) error {
		return writeEvent("delta", map[string]any{"type": "delta", "text": text})
	})
	if err != nil {
		_ = writeEvent("error", map[string]any{"type": "error", "message": err.Error()})
		return
	}
	_ = writeEvent("done", map[string]any{"type": "done", "data": result})
}

func (handler *Handler) erxinAssessmentRecordIEPPlanWord(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var recordID int64
	var durationMonths int
	var plan *model.PEP3IEPPlanAIResult
	if r.Method == http.MethodPost {
		var req model.PEP3IEPPlanWordExportRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
			return
		}
		recordID = req.ID
		durationMonths = req.DurationMonths
		plan = req.Plan
	} else if rawID := strings.TrimSpace(r.URL.Query().Get("id")); rawID != "" {
		parsedID, err := strconv.ParseInt(rawID, 10, 64)
		if err != nil || parsedID < 0 {
			httpx.WriteError(w, http.StatusBadRequest, "invalid id", ctx.RequestID)
			return
		}
		recordID = parsedID
	}
	if r.Method == http.MethodGet {
		durationMonths = parsePEP3IEPPlanDurationQuery(r)
	}
	var fileName string
	var contentType string
	var content []byte
	var err error
	if plan != nil {
		fileName, contentType, content, err = handler.service.ExportERXinIEPPlanWordFromAIResult(claims.UserID, recordID, *plan, durationMonths)
	} else {
		fileName, contentType, content, err = handler.service.ExportERXinIEPPlanWord(claims.UserID, recordID, durationMonths)
	}
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

func (handler *Handler) erxinAssessmentRecordIEPPlanPDF(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var recordID int64
	var durationMonths int
	var plan *model.PEP3IEPPlanAIResult
	if r.Method == http.MethodPost {
		var req model.PEP3IEPPlanWordExportRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
			return
		}
		recordID = req.ID
		durationMonths = req.DurationMonths
		plan = req.Plan
	} else if rawID := strings.TrimSpace(r.URL.Query().Get("id")); rawID != "" {
		parsedID, err := strconv.ParseInt(rawID, 10, 64)
		if err != nil || parsedID < 0 {
			httpx.WriteError(w, http.StatusBadRequest, "invalid id", ctx.RequestID)
			return
		}
		recordID = parsedID
	}
	if r.Method == http.MethodGet {
		durationMonths = parsePEP3IEPPlanDurationQuery(r)
	}
	var fileName string
	var contentType string
	var content []byte
	var err error
	if plan != nil {
		fileName, contentType, content, err = handler.service.ExportERXinIEPPlanPDFFromAIResult(claims.UserID, recordID, *plan, durationMonths)
	} else {
		fileName, contentType, content, err = handler.service.ExportERXinIEPPlanPDF(claims.UserID, recordID, durationMonths)
	}
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

func (handler *Handler) erxinAssessmentRecordExecutionPlanWord(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req model.PEP3ExecutionPlanWordExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	fileName, contentType, content, err := handler.service.ExportERXinExecutionPlanWord(claims.UserID, req)
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

func (handler *Handler) erxinAssessmentRecordExecutionPlanPDF(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req model.PEP3ExecutionPlanWordExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	fileName, contentType, content, err := handler.service.ExportERXinExecutionPlanPDF(claims.UserID, req)
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

func (handler *Handler) erxinAssessmentRecordExecutionPlanAI(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req model.PEP3ExecutionPlanGenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.GenerateERXinExecutionPlanWithAI(r.Context(), claims.UserID, req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) erxinAssessmentRecordExecutionPlanAIStream(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req model.PEP3ExecutionPlanGenerateRequest
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
	planTypeLabel := "执行计划"
	switch strings.ToLower(strings.TrimSpace(req.PlanType)) {
	case "monthly":
		planTypeLabel = "月度计划"
	case "weekly":
		planTypeLabel = "周计划"
	}
	if err := writeEvent("status", map[string]any{"type": "status", "message": "正在准备" + planTypeLabel + "生成上下文"}); err != nil {
		return
	}
	var streamText strings.Builder
	result, usage, err := handler.service.GenerateERXinExecutionPlanWithAIStream(r.Context(), claims.UserID, req, func(text string) error {
		streamText.WriteString(text)
		return writeEvent("delta", map[string]any{
			"type":          "delta",
			"text":          text,
			"costAmountCny": service.EstimateDeepSeekOutputCostCNY(streamText.String()),
		})
	})
	if err != nil {
		_ = writeEvent("error", map[string]any{"type": "error", "message": err.Error()})
		return
	}
	costAmountCny := service.ComputeDeepSeekUsageCostCNY(usage, "")
	_ = writeEvent("done", map[string]any{
		"type":          "done",
		"data":          result,
		"costAmountCny": costAmountCny,
	})
}

func (handler *Handler) erxinAssessmentRecordExecutionPlanDetail(w http.ResponseWriter, r *http.Request) {
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
	durationMonths := parsePEP3IEPPlanDurationQuery(r)
	result, err := handler.service.GetERXinExecutionPlans(claims.UserID, id, durationMonths)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) erxinAssessmentRecordExecutionPlanSave(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req model.PEP3ExecutionPlanSaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.SaveERXinExecutionPlan(claims.UserID, req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) erxinAssessmentRecordLessonSessionWeekState(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	recordID, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("id")), 10, 64)
	if err != nil || recordID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id", ctx.RequestID)
		return
	}
	targetMonthIndex, err := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("targetMonthIndex")))
	if err != nil || targetMonthIndex <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid targetMonthIndex", ctx.RequestID)
		return
	}
	targetWeekIndex, err := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("targetWeekIndex")))
	if err != nil || targetWeekIndex <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid targetWeekIndex", ctx.RequestID)
		return
	}
	result, err := handler.service.GetERXinLessonSessionWeekState(claims.UserID, model.PEP3LessonSessionWeekQueryRequest{
		ID:               recordID,
		DurationMonths:   parsePEP3IEPPlanDurationQuery(r),
		TargetMonthIndex: targetMonthIndex,
		TargetWeekIndex:  targetWeekIndex,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) erxinAssessmentRecordLessonSessionStart(w http.ResponseWriter, r *http.Request) {
	handler.handleERXinLessonSessionOperate(w, r, func(userID int64, req model.PEP3LessonSessionOperateRequest) (model.PEP3LessonSessionWeekStateVO, error) {
		return handler.service.StartERXinLessonSession(userID, req)
	})
}

func (handler *Handler) erxinAssessmentRecordLessonSessionPause(w http.ResponseWriter, r *http.Request) {
	handler.handleERXinLessonSessionOperate(w, r, func(userID int64, req model.PEP3LessonSessionOperateRequest) (model.PEP3LessonSessionWeekStateVO, error) {
		return handler.service.PauseERXinLessonSession(userID, req)
	})
}

func (handler *Handler) erxinAssessmentRecordLessonSessionComplete(w http.ResponseWriter, r *http.Request) {
	handler.handleERXinLessonSessionOperate(w, r, func(userID int64, req model.PEP3LessonSessionOperateRequest) (model.PEP3LessonSessionWeekStateVO, error) {
		return handler.service.CompleteERXinLessonSession(userID, req)
	})
}

func (handler *Handler) erxinAssessmentRecordLessonSessionHeartbeat(w http.ResponseWriter, r *http.Request) {
	handler.handleERXinLessonSessionOperate(w, r, func(userID int64, req model.PEP3LessonSessionOperateRequest) (model.PEP3LessonSessionWeekStateVO, error) {
		return handler.service.HeartbeatERXinLessonSession(userID, req)
	})
}

func (handler *Handler) handleERXinLessonSessionOperate(
	w http.ResponseWriter,
	r *http.Request,
	operate func(int64, model.PEP3LessonSessionOperateRequest) (model.PEP3LessonSessionWeekStateVO, error),
) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req model.PEP3LessonSessionOperateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := operate(claims.UserID, req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) erxinAssessmentRecordIEPPlanAI(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req model.PEP3IEPPlanGenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.GenerateERXinIEPPlanWithAI(claims.UserID, req.ID, req.DurationMonths)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) erxinAssessmentRecordIEPPlanAITask(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req model.PEP3IEPPlanGenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.CreateERXinIEPPlanGenerationTask(claims.UserID, req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) erxinAssessmentRecordIEPPlanAITaskDetail(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	taskID := strings.TrimSpace(r.URL.Query().Get("taskId"))
	result, err := handler.service.GetIEPPlanGenerationTask(claims.UserID, taskID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) erxinAssessmentRecordIEPPlanAITaskActive(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	recordID, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("id")), 10, 64)
	if err != nil || recordID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id", ctx.RequestID)
		return
	}
	result, err := handler.service.GetERXinActiveIEPPlanGenerationTask(claims.UserID, recordID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) erxinAssessmentRecordIEPPlanAITaskStream(w http.ResponseWriter, r *http.Request) {
	handler.iepPlanGenerationTaskStream(w, r)
}

func (handler *Handler) erxinAssessmentRecordIEPPlanDetail(w http.ResponseWriter, r *http.Request) {
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
	durationMonths := parsePEP3IEPPlanDurationQuery(r)
	result, err := handler.service.GetERXinIEPPlan(claims.UserID, id, durationMonths)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) erxinAssessmentRecordIEPPlanSave(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req model.PEP3IEPPlanSaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.SaveERXinIEPPlan(claims.UserID, req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) erxinAssessmentRecordIEPPlanPeriodSync(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req model.PEP3IEPPlanPeriodSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.SyncERXinIEPPlanPeriod(claims.UserID, req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) erxinAssessmentRecordIEPPlanAIStream(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req model.PEP3IEPPlanGenerateRequest
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
	if err := writeEvent("status", map[string]any{"type": "status", "message": "正在读取儿心评估结果和报告解读"}); err != nil {
		return
	}
	result, _, err := handler.service.GenerateERXinIEPPlanWithAIStream(r.Context(), claims.UserID, req.ID, req.DurationMonths, func(text string) error {
		return writeEvent("delta", map[string]any{"type": "delta", "text": text})
	})
	if err != nil {
		_ = writeEvent("error", map[string]any{"type": "error", "message": err.Error()})
		return
	}
	_ = writeEvent("done", map[string]any{"type": "done", "data": result})
}

func (handler *Handler) erxinAssessmentRecordsPage(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.PageERXinAssessmentRecords(claims.UserID, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) erxinAssessmentRecordCategoryStats(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.SummarizeERXinAssessmentRecordCategories(claims.UserID, query.QueryModel)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) deleteERXinAssessmentRecord(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req erxinAssessmentDraftDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if req.ID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id", ctx.RequestID)
		return
	}
	result, err := handler.service.DeleteERXinAssessmentRecord(claims.UserID, req.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) erxinAssessmentFormTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	result, err := handler.service.GetERXinAssessmentFormTemplate()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) erxinAssessmentFormTemplateSummary(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	result, err := handler.service.GetERXinAssessmentFormTemplateSummary()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) erxinAssessmentFormTemplateItem(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.GetERXinAssessmentFormTemplateItem(itemNo)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (req erxinScoreRequest) toAssessmentInput() (erxinscore.AssessmentInput, error) {
	birthDate, err := parseERXinDate(req.BirthDate, "birthDate")
	if err != nil {
		return erxinscore.AssessmentInput{}, err
	}
	assessmentDate, err := parseERXinDate(req.AssessmentDate, "assessmentDate")
	if err != nil {
		return erxinscore.AssessmentInput{}, err
	}
	itemPasses, err := normalizeERXinItemPasses(req.ItemPasses, req.ItemPassList)
	if err != nil {
		return erxinscore.AssessmentInput{}, err
	}
	if len(itemPasses) == 0 {
		return erxinscore.AssessmentInput{}, fmt.Errorf("itemPasses or itemPassList is required")
	}
	return erxinscore.AssessmentInput{
		BirthDate:      birthDate,
		AssessmentDate: assessmentDate,
		ItemPasses:     itemPasses,
	}, nil
}

func (req erxinAssessmentDraftSaveRequest) toDraftSaveInput() (service.ERXinAssessmentDraftSaveInput, error) {
	birthDate, err := parseOptionalERXinDate(req.BirthDate, "birthDate")
	if err != nil {
		return service.ERXinAssessmentDraftSaveInput{}, err
	}
	assessmentDate, err := parseOptionalERXinDate(req.AssessmentDate, "assessmentDate")
	if err != nil {
		return service.ERXinAssessmentDraftSaveInput{}, err
	}
	itemPasses, err := normalizeERXinItemPasses(req.ItemPasses, req.ItemPassList)
	if err != nil {
		return service.ERXinAssessmentDraftSaveInput{}, err
	}
	return service.ERXinAssessmentDraftSaveInput{
		ID:             req.ID,
		StudentID:      req.StudentID,
		StudentName:    strings.TrimSpace(req.StudentName),
		ExaminerName:   strings.TrimSpace(req.ExaminerName),
		Remark:         strings.TrimSpace(req.Remark),
		BirthDate:      birthDate,
		AssessmentDate: assessmentDate,
		ItemPasses:     itemPasses,
		InputSnapshot:  req.normalizedSnapshot(itemPasses, normalizeERXinItemRemarks(req.ItemRemarks, req.ItemRemarkList, req.ItemPassList)),
	}, nil
}

func (req erxinAssessmentRecordCreateRequest) toRecordSaveInput() (service.ERXinAssessmentRecordSaveInput, error) {
	scoreReq := req.toScoreRequest()
	scoreInput, err := scoreReq.toAssessmentInput()
	if err != nil {
		return service.ERXinAssessmentRecordSaveInput{}, err
	}
	return service.ERXinAssessmentRecordSaveInput{
		StudentID:    req.StudentID,
		StudentName:  strings.TrimSpace(req.StudentName),
		ExaminerName: strings.TrimSpace(req.ExaminerName),
		Remark:       strings.TrimSpace(req.Remark),
		ScoreInput:   scoreInput,
		InputSnapshot: req.normalizedSnapshot(
			scoreInput.ItemPasses,
			normalizeERXinItemRemarks(req.ItemRemarks, req.ItemRemarkList, req.ItemPassList),
		),
	}, nil
}

func (req erxinAssessmentRecordCreateRequest) toScoreRequest() erxinScoreRequest {
	return erxinScoreRequest{
		BirthDate:      req.BirthDate,
		AssessmentDate: req.AssessmentDate,
		ItemPasses:     req.ItemPasses,
		ItemPassList:   req.ItemPassList,
	}
}

func (req erxinAssessmentDraftSaveRequest) normalizedSnapshot(
	itemPasses map[int]bool,
	itemRemarks map[int]string,
) any {
	normalizedPassList := erxinItemPassListFromMap(itemPasses, itemRemarks)
	normalizedRemarkList := erxinItemRemarkListFromMap(itemRemarks)
	return struct {
		ID             int64                    `json:"id,omitempty"`
		StudentID      int64                    `json:"studentId,omitempty"`
		StudentName    string                   `json:"studentName,omitempty"`
		ExaminerName   string                   `json:"examinerName,omitempty"`
		Remark         string                   `json:"remark,omitempty"`
		BirthDate      string                   `json:"birthDate,omitempty"`
		AssessmentDate string                   `json:"assessmentDate,omitempty"`
		ItemPasses     map[int]bool             `json:"itemPasses,omitempty"`
		ItemPassList   []erxinItemPassRequest   `json:"itemPassList,omitempty"`
		ItemRemarks    map[int]string           `json:"itemRemarks,omitempty"`
		ItemRemarkList []erxinItemRemarkRequest `json:"itemRemarkList,omitempty"`
	}{
		ID:             req.ID,
		StudentID:      req.StudentID,
		StudentName:    strings.TrimSpace(req.StudentName),
		ExaminerName:   strings.TrimSpace(req.ExaminerName),
		Remark:         strings.TrimSpace(req.Remark),
		BirthDate:      strings.TrimSpace(req.BirthDate),
		AssessmentDate: strings.TrimSpace(req.AssessmentDate),
		ItemPasses:     itemPasses,
		ItemPassList:   normalizedPassList,
		ItemRemarks:    itemRemarks,
		ItemRemarkList: normalizedRemarkList,
	}
}

func normalizeERXinItemRemarks(
	itemRemarks map[int]string,
	itemRemarkList []erxinItemRemarkRequest,
	itemPassList []erxinItemPassRequest,
) map[int]string {
	out := make(map[int]string, len(itemRemarks)+len(itemRemarkList)+len(itemPassList))
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
	for _, item := range itemPassList {
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

func (req erxinAssessmentRecordCreateRequest) normalizedSnapshot(
	itemPasses map[int]bool,
	itemRemarks map[int]string,
) any {
	normalizedPassList := erxinItemPassListFromMap(itemPasses, itemRemarks)
	normalizedRemarkList := erxinItemRemarkListFromMap(itemRemarks)
	return struct {
		ID             int64                    `json:"id,omitempty"`
		StudentID      int64                    `json:"studentId,omitempty"`
		StudentName    string                   `json:"studentName,omitempty"`
		ExaminerName   string                   `json:"examinerName,omitempty"`
		Remark         string                   `json:"remark,omitempty"`
		BirthDate      string                   `json:"birthDate"`
		AssessmentDate string                   `json:"assessmentDate"`
		ItemPasses     map[int]bool             `json:"itemPasses,omitempty"`
		ItemPassList   []erxinItemPassRequest   `json:"itemPassList,omitempty"`
		ItemRemarks    map[int]string           `json:"itemRemarks,omitempty"`
		ItemRemarkList []erxinItemRemarkRequest `json:"itemRemarkList,omitempty"`
	}{
		ID:             req.ID,
		StudentID:      req.StudentID,
		StudentName:    strings.TrimSpace(req.StudentName),
		ExaminerName:   strings.TrimSpace(req.ExaminerName),
		Remark:         strings.TrimSpace(req.Remark),
		BirthDate:      strings.TrimSpace(req.BirthDate),
		AssessmentDate: strings.TrimSpace(req.AssessmentDate),
		ItemPasses:     itemPasses,
		ItemPassList:   normalizedPassList,
		ItemRemarks:    itemRemarks,
		ItemRemarkList: normalizedRemarkList,
	}
}

func parseERXinDate(raw, field string) (time.Time, error) {
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

func parseOptionalERXinDate(raw, field string) (*time.Time, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	parsed, err := parseERXinDate(raw, field)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func normalizeERXinItemPasses(itemPasses map[int]bool, itemPassList []erxinItemPassRequest) (map[int]bool, error) {
	normalized := make(map[int]bool, len(itemPasses)+len(itemPassList))
	for itemNo, passed := range itemPasses {
		if itemNo <= 0 {
			return nil, fmt.Errorf("itemPasses contains invalid itemNo %d", itemNo)
		}
		normalized[itemNo] = passed
	}
	for _, item := range itemPassList {
		if item.ItemNo <= 0 {
			return nil, fmt.Errorf("itemPassList contains invalid itemNo %d", item.ItemNo)
		}
		normalized[item.ItemNo] = item.Passed
	}
	return normalized, nil
}

func erxinItemPassListFromMap(itemPasses map[int]bool, itemRemarks map[int]string) []erxinItemPassRequest {
	if len(itemPasses) == 0 {
		return nil
	}
	itemNos := make([]int, 0, len(itemPasses))
	for itemNo := range itemPasses {
		itemNos = append(itemNos, itemNo)
	}
	sort.Ints(itemNos)
	out := make([]erxinItemPassRequest, 0, len(itemNos))
	for _, itemNo := range itemNos {
		out = append(out, erxinItemPassRequest{
			ItemNo: itemNo,
			Passed: itemPasses[itemNo],
			Remark: strings.TrimSpace(itemRemarks[itemNo]),
		})
	}
	return out
}

func erxinItemRemarkListFromMap(itemRemarks map[int]string) []erxinItemRemarkRequest {
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
	out := make([]erxinItemRemarkRequest, 0, len(itemNos))
	for _, itemNo := range itemNos {
		remark := strings.TrimSpace(itemRemarks[itemNo])
		if remark == "" {
			continue
		}
		out = append(out, erxinItemRemarkRequest{ItemNo: itemNo, Remark: remark})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
