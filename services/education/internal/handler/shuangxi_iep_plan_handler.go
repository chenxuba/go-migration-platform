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

func (handler *Handler) shuangxiAAssessmentRecordIEPPlanWord(w http.ResponseWriter, r *http.Request) {
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
		fileName, contentType, content, err = handler.service.ExportShuangxiAIEPPlanWordFromAIResult(claims.UserID, recordID, *plan, durationMonths)
	} else {
		fileName, contentType, content, err = handler.service.ExportShuangxiAIEPPlanWord(claims.UserID, recordID, durationMonths)
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

func (handler *Handler) shuangxiAAssessmentRecordIEPPlanPDF(w http.ResponseWriter, r *http.Request) {
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
		fileName, contentType, content, err = handler.service.ExportShuangxiAIEPPlanPDFFromAIResult(claims.UserID, recordID, *plan, durationMonths)
	} else {
		fileName, contentType, content, err = handler.service.ExportShuangxiAIEPPlanPDF(claims.UserID, recordID, durationMonths)
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

func (handler *Handler) shuangxiAAssessmentRecordIEPPlanDetail(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.GetShuangxiAIEPPlan(claims.UserID, id, durationMonths)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) shuangxiAAssessmentRecordIEPPlanSave(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.SaveShuangxiAIEPPlan(claims.UserID, req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) shuangxiAAssessmentRecordIEPPlanPeriodSync(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.SyncShuangxiAIEPPlanPeriod(claims.UserID, req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) shuangxiAAssessmentRecordIEPPlanAIStream(w http.ResponseWriter, r *http.Request) {
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
	if err := writeEvent("status", map[string]any{"type": "status", "message": "正在读取双溪课程评量结果和评量结果分析"}); err != nil {
		return
	}
	result, _, err := handler.service.GenerateShuangxiAIEPPlanWithAIStream(r.Context(), claims.UserID, req.ID, req.DurationMonths, func(text string) error {
		return writeEvent("delta", map[string]any{"type": "delta", "text": text})
	})
	if err != nil {
		_ = writeEvent("error", map[string]any{"type": "error", "message": err.Error()})
		return
	}
	_ = writeEvent("done", map[string]any{"type": "done", "data": result})
}

func (handler *Handler) shuangxiAAssessmentRecordIEPPlanAI(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.GenerateShuangxiAIEPPlanWithAI(claims.UserID, req.ID, req.DurationMonths)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) shuangxiAAssessmentRecordIEPPlanAITask(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.CreateShuangxiAIEPPlanGenerationTask(claims.UserID, req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) shuangxiAAssessmentRecordIEPPlanAITaskDetail(w http.ResponseWriter, r *http.Request) {
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

func (handler *Handler) shuangxiAAssessmentRecordIEPPlanAITaskActive(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.GetShuangxiAActiveIEPPlanGenerationTask(claims.UserID, recordID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) shuangxiAAssessmentRecordIEPPlanAITaskStream(w http.ResponseWriter, r *http.Request) {
	handler.iepPlanGenerationTaskStream(w, r)
}

func (handler *Handler) shuangxiAAssessmentRecordExecutionPlanWord(w http.ResponseWriter, r *http.Request) {
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
	fileName, contentType, content, err := handler.service.ExportShuangxiAExecutionPlanWord(claims.UserID, req)
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

func (handler *Handler) shuangxiAAssessmentRecordExecutionPlanPDF(w http.ResponseWriter, r *http.Request) {
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
	fileName, contentType, content, err := handler.service.ExportShuangxiAExecutionPlanPDF(claims.UserID, req)
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

func (handler *Handler) shuangxiAAssessmentRecordExecutionPlanAI(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.GenerateShuangxiAExecutionPlanWithAI(r.Context(), claims.UserID, req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) shuangxiAAssessmentRecordExecutionPlanAITask(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.CreateShuangxiAExecutionPlanGenerationTask(claims.UserID, req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) shuangxiAAssessmentRecordExecutionPlanAITaskDetail(w http.ResponseWriter, r *http.Request) {
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

func (handler *Handler) shuangxiAAssessmentRecordExecutionPlanAITaskActive(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.GetShuangxiAActiveExecutionPlanGenerationTask(claims.UserID, req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) shuangxiAAssessmentRecordExecutionPlanAITaskStream(w http.ResponseWriter, r *http.Request) {
	handler.executionPlanGenerationTaskStream(w, r)
}

func (handler *Handler) shuangxiAAssessmentRecordExecutionPlanAIStream(w http.ResponseWriter, r *http.Request) {
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
	result, usage, err := handler.service.GenerateShuangxiAExecutionPlanWithAIStream(r.Context(), claims.UserID, req, func(text string) error {
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

func (handler *Handler) shuangxiAAssessmentRecordExecutionPlanDetail(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.GetShuangxiAExecutionPlans(claims.UserID, id, durationMonths)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) shuangxiAAssessmentRecordExecutionPlanSave(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.SaveShuangxiAExecutionPlan(claims.UserID, req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) shuangxiAAssessmentRecordLessonSessionWeekState(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.GetShuangxiALessonSessionWeekState(claims.UserID, model.PEP3LessonSessionWeekQueryRequest{
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

func (handler *Handler) shuangxiAAssessmentRecordLessonSessionStart(w http.ResponseWriter, r *http.Request) {
	handler.handleShuangxiALessonSessionOperate(w, r, func(userID int64, req model.PEP3LessonSessionOperateRequest) (model.PEP3LessonSessionWeekStateVO, error) {
		return handler.service.StartShuangxiALessonSession(userID, req)
	})
}

func (handler *Handler) shuangxiAAssessmentRecordLessonSessionPause(w http.ResponseWriter, r *http.Request) {
	handler.handleShuangxiALessonSessionOperate(w, r, func(userID int64, req model.PEP3LessonSessionOperateRequest) (model.PEP3LessonSessionWeekStateVO, error) {
		return handler.service.PauseShuangxiALessonSession(userID, req)
	})
}

func (handler *Handler) shuangxiAAssessmentRecordLessonSessionComplete(w http.ResponseWriter, r *http.Request) {
	handler.handleShuangxiALessonSessionOperate(w, r, func(userID int64, req model.PEP3LessonSessionOperateRequest) (model.PEP3LessonSessionWeekStateVO, error) {
		return handler.service.CompleteShuangxiALessonSession(userID, req)
	})
}

func (handler *Handler) shuangxiAAssessmentRecordLessonSessionHeartbeat(w http.ResponseWriter, r *http.Request) {
	handler.handleShuangxiALessonSessionOperate(w, r, func(userID int64, req model.PEP3LessonSessionOperateRequest) (model.PEP3LessonSessionWeekStateVO, error) {
		return handler.service.HeartbeatShuangxiALessonSession(userID, req)
	})
}

func (handler *Handler) handleShuangxiALessonSessionOperate(
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
