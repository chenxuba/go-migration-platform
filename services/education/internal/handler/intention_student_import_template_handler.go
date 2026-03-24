package handler

import (
	"net/http"
	"net/url"
	"strings"

	"go-migration-platform/pkg/httpx"
	"go-migration-platform/pkg/tenant"
)

func (handler *Handler) buildIntentionStudentImportTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	result, err := handler.service.BuildIntentionStudentImportTemplate(claims.UserID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) downloadIntentionStudentImportTemplate(w http.ResponseWriter, r *http.Request) {
	ticket, _ := url.QueryUnescape(strings.TrimSpace(r.URL.Query().Get("ticket")))
	filename, contentType, data, ok := handler.service.LoadIntentionStudentImportTemplate(ticket)
	if !ok {
		httpx.WriteError(w, http.StatusNotFound, "template not found or expired", "")
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+url.QueryEscape(filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
