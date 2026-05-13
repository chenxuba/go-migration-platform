package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"go-migration-platform/pkg/httpx"
	"go-migration-platform/pkg/tenant"
	"go-migration-platform/services/platform/internal/model"
)

func (handler *Handler) platformPEP3IEPMaterialImportTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requirePlatformManage(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	result, err := handler.service.BuildPlatformPEP3IEPMaterialImportTemplate()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) platformPEP3IEPMaterialImportTemplateFile(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requirePlatformManage(w, r, ctx); !ok {
		return
	}
	ticket := strings.TrimSpace(r.URL.Query().Get("ticket"))
	filename, contentType, data, ok := handler.service.LoadPlatformPEP3IEPMaterialImportTemplate(ticket)
	if !ok {
		httpx.WriteError(w, http.StatusNotFound, "template not found or expired", ctx.RequestID)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", platformAttachmentDisposition(filename, "pep3-iep-material-template.xlsx"))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (handler *Handler) platformPEP3IEPMaterialImportUpload(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requirePlatformManage(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid multipart form", ctx.RequestID)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "file is required", ctx.RequestID)
		return
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "read file failed", ctx.RequestID)
		return
	}
	result, err := handler.service.UploadPlatformPEP3IEPMaterialImportFile(header.Filename, data)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) platformPEP3IEPMaterialImportUploadedFile(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requirePlatformManage(w, r, ctx); !ok {
		return
	}
	ticket := strings.TrimSpace(r.URL.Query().Get("ticket"))
	filename, data, ok := handler.service.LoadUploadedPlatformPEP3IEPMaterialImportFile(ticket)
	if !ok {
		httpx.WriteError(w, http.StatusNotFound, "uploaded file not found or expired", ctx.RequestID)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", platformAttachmentDisposition(filename, "pep3-iep-material-import.xlsx"))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (handler *Handler) platformPEP3IEPMaterialImportSubmit(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requirePlatformManage(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req model.PEP3IEPMaterialImportSubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	taskID, err := handler.service.SubmitPlatformPEP3IEPMaterialImportTask(claims.UserID, claims.Username, req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, taskID, ctx.RequestID)
}

func (handler *Handler) platformPEP3IEPMaterialImportTaskDetail(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requirePlatformManage(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	taskID := strings.TrimSpace(r.URL.Query().Get("taskId"))
	result, err := handler.service.GetPlatformPEP3IEPMaterialImportTaskDetail(taskID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) platformPEP3IEPMaterialImportTaskList(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requirePlatformManage(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	result, err := handler.service.ListPlatformPEP3IEPMaterialImportTasks()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) platformPEP3IEPMaterialImportTaskRecords(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requirePlatformManage(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req model.PEP3IEPMaterialImportTaskRecordListQuery
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.GetPlatformPEP3IEPMaterialImportTaskRecordList(req.QueryModel.TaskID, req.QueryModel.Type)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) platformPEP3IEPMaterialImportBatchSaveRecords(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requirePlatformManage(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req model.PEP3IEPMaterialImportSaveTaskRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.BatchSavePlatformPEP3IEPMaterialImportTaskRecords(req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) platformPEP3IEPMaterialImportStart(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requirePlatformManage(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req model.PEP3IEPMaterialImportStartTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if err := handler.service.StartPlatformPEP3IEPMaterialImportTask(claims.UserID, claims.Username, req.TaskID); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func platformAttachmentDisposition(filename, fallback string) string {
	filename = strings.TrimSpace(filename)
	if filename == "" {
		filename = fallback
	}
	return `attachment; filename="` + fallback + `"; filename*=UTF-8''` + url.PathEscape(filename)
}

func (handler *Handler) platformPEP3IEPMaterialImportClear(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requirePlatformManage(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	if err := handler.service.ClearPlatformPEP3IEPMaterialImportTasks(); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func (handler *Handler) platformPEP3IEPMaterialImportDelete(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requirePlatformManage(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	_ = claims
	var payload struct {
		TaskID string `json:"taskId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if err := handler.service.DeletePlatformPEP3IEPMaterialImportTask(payload.TaskID); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}
