package handler

import (
	"net/http"

	"go-migration-platform/pkg/httpx"
	"go-migration-platform/pkg/tenant"
)

func (handler *Handler) parseIntentionStudentImportFile(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
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

	result, err := handler.service.ParseIntentionStudentImportFile(claims.UserID, header.Filename, file)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}
