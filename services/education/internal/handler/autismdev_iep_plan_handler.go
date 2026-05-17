package handler

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"go-migration-platform/pkg/httpx"
	"go-migration-platform/pkg/tenant"
	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/service"
)

func (handler *Handler) autismDevAssessmentRecordIEPPlanWord(w http.ResponseWriter, r *http.Request) {
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
		fileName, contentType, content, err = handler.service.ExportAutismDevIEPPlanWordFromAIResult(claims.UserID, recordID, *plan, durationMonths)
	} else {
		fileName, contentType, content, err = handler.service.ExportAutismDevIEPPlanWord(claims.UserID, recordID, durationMonths)
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

func (handler *Handler) autismDevAssessmentRecordIEPPlanPDF(w http.ResponseWriter, r *http.Request) {
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
		fileName, contentType, content, err = handler.service.ExportAutismDevIEPPlanPDFFromAIResult(claims.UserID, recordID, *plan, durationMonths)
	} else {
		fileName, contentType, content, err = handler.service.ExportAutismDevIEPPlanPDF(claims.UserID, recordID, durationMonths)
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

func (handler *Handler) autismDevAssessmentRecordExecutionPlanWord(w http.ResponseWriter, r *http.Request) {
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
	fileName, contentType, content, err := handler.service.ExportAutismDevExecutionPlanWord(claims.UserID, req)
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

func (handler *Handler) autismDevAssessmentRecordExecutionPlanPDF(w http.ResponseWriter, r *http.Request) {
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
	fileName, contentType, content, err := handler.service.ExportAutismDevExecutionPlanPDF(claims.UserID, req)
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

func (handler *Handler) autismDevAssessmentRecordExecutionPlanAI(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.GenerateAutismDevExecutionPlanWithAI(r.Context(), claims.UserID, req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) autismDevAssessmentRecordExecutionPlanAITask(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.CreateAutismDevExecutionPlanGenerationTask(claims.UserID, req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) autismDevAssessmentRecordExecutionPlanAITaskDetail(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.GetExecutionPlanGenerationTask(claims.UserID, taskID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) autismDevAssessmentRecordExecutionPlanAITaskActive(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	req, err := parseExecutionPlanActiveTaskRequest(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	result, err := handler.service.GetAutismDevActiveExecutionPlanGenerationTask(claims.UserID, req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) autismDevAssessmentRecordExecutionPlanAITaskStream(w http.ResponseWriter, r *http.Request) {
	handler.executionPlanGenerationTaskStream(w, r)
}

func (handler *Handler) autismDevAssessmentRecordExecutionPlanAIStream(w http.ResponseWriter, r *http.Request) {
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
	result, usage, err := handler.service.GenerateAutismDevExecutionPlanWithAIStream(r.Context(), claims.UserID, req, func(text string) error {
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

func (handler *Handler) autismDevAssessmentRecordExecutionPlanDetail(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.GetAutismDevExecutionPlans(claims.UserID, id, durationMonths)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) autismDevAssessmentRecordExecutionPlanSave(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.SaveAutismDevExecutionPlan(claims.UserID, req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) autismDevAssessmentRecordLessonSessionWeekState(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.GetAutismDevLessonSessionWeekState(claims.UserID, model.PEP3LessonSessionWeekQueryRequest{
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

func (handler *Handler) autismDevAssessmentRecordLessonSessionStart(w http.ResponseWriter, r *http.Request) {
	handler.handleAutismDevLessonSessionOperate(w, r, func(userID int64, req model.PEP3LessonSessionOperateRequest) (model.PEP3LessonSessionWeekStateVO, error) {
		return handler.service.StartAutismDevLessonSession(userID, req)
	})
}

func (handler *Handler) autismDevAssessmentRecordLessonSessionPause(w http.ResponseWriter, r *http.Request) {
	handler.handleAutismDevLessonSessionOperate(w, r, func(userID int64, req model.PEP3LessonSessionOperateRequest) (model.PEP3LessonSessionWeekStateVO, error) {
		return handler.service.PauseAutismDevLessonSession(userID, req)
	})
}

func (handler *Handler) autismDevAssessmentRecordLessonSessionComplete(w http.ResponseWriter, r *http.Request) {
	handler.handleAutismDevLessonSessionOperate(w, r, func(userID int64, req model.PEP3LessonSessionOperateRequest) (model.PEP3LessonSessionWeekStateVO, error) {
		return handler.service.CompleteAutismDevLessonSession(userID, req)
	})
}

func (handler *Handler) autismDevAssessmentRecordLessonSessionHeartbeat(w http.ResponseWriter, r *http.Request) {
	handler.handleAutismDevLessonSessionOperate(w, r, func(userID int64, req model.PEP3LessonSessionOperateRequest) (model.PEP3LessonSessionWeekStateVO, error) {
		return handler.service.HeartbeatAutismDevLessonSession(userID, req)
	})
}

func (handler *Handler) handleAutismDevLessonSessionOperate(
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

func (handler *Handler) autismDevAssessmentRecordIEPPlanAI(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.GenerateAutismDevIEPPlanWithAI(claims.UserID, req.ID, req.DurationMonths)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) autismDevAssessmentRecordIEPPlanAITask(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.CreateAutismDevIEPPlanGenerationTask(claims.UserID, req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) autismDevAssessmentRecordIEPPlanAITaskDetail(w http.ResponseWriter, r *http.Request) {
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

func (handler *Handler) autismDevAssessmentRecordIEPPlanAITaskActive(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.GetAutismDevActiveIEPPlanGenerationTask(claims.UserID, recordID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) autismDevAssessmentRecordIEPPlanAITaskStream(w http.ResponseWriter, r *http.Request) {
	handler.iepPlanGenerationTaskStream(w, r)
}

func (handler *Handler) autismDevAssessmentRecordIEPPlanDetail(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.GetAutismDevIEPPlan(claims.UserID, id, durationMonths)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) autismDevAssessmentRecordIEPPlanSave(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.SaveAutismDevIEPPlan(claims.UserID, req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) autismDevAssessmentRecordIEPPlanPeriodSync(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.SyncAutismDevIEPPlanPeriod(claims.UserID, req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) autismDevAssessmentRecordIEPPlanAIStream(w http.ResponseWriter, r *http.Request) {
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
	if err := writeEvent("status", map[string]any{"type": "status", "message": "正在读取孤独症8大项评估结果和报告分析"}); err != nil {
		return
	}
	result, _, err := handler.service.GenerateAutismDevIEPPlanWithAIStream(r.Context(), claims.UserID, req.ID, req.DurationMonths, func(text string) error {
		return writeEvent("delta", map[string]any{"type": "delta", "text": text})
	})
	if err != nil {
		_ = writeEvent("error", map[string]any{"type": "error", "message": err.Error()})
		return
	}
	_ = writeEvent("done", map[string]any{"type": "done", "data": result})
}
