package handler

import (
	"encoding/json"
	"net/http"

	"go-migration-platform/pkg/authx"
	"go-migration-platform/pkg/httpx"
	"go-migration-platform/pkg/tenant"
	"go-migration-platform/services/education/internal/model"
)

func (handler *Handler) parentWeChatLogin(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	var dto model.ParentWeChatLoginDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}

	result, err := handler.service.ParentWeChatLogin(r.Context(), ctx.TenantID, dto)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) parentStudentsByPhone(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireParentAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	result, err := handler.service.LookupParentStudentsByPhone(r.Context(), claims.Username)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) requireParentAuth(w http.ResponseWriter, r *http.Request, ctx tenant.Context) (authx.Claims, bool) {
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return authx.Claims{}, false
	}
	if claims.LoginType != model.ParentLoginTypeMiniProgram {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", ctx.RequestID)
		return authx.Claims{}, false
	}
	if claims.TenantID != "" && ctx.TenantID != "" && claims.TenantID != ctx.TenantID {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", ctx.RequestID)
		return authx.Claims{}, false
	}
	return claims, true
}
