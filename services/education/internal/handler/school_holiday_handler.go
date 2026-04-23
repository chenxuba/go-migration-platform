package handler

import (
	"encoding/json"
	"net/http"

	"go-migration-platform/pkg/httpx"
	"go-migration-platform/pkg/tenant"
	"go-migration-platform/services/education/internal/model"
)

func (handler *Handler) schoolHolidays(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	items, err := handler.service.ListSchoolHolidays(claims.UserID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items, ctx.RequestID)
}

func (handler *Handler) saveSchoolHoliday(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var input model.SchoolHolidayMutation
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	id, err := handler.service.SaveSchoolHoliday(claims.UserID, input)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"id": id}, ctx.RequestID)
}

func (handler *Handler) deleteSchoolHoliday(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var input model.SchoolHolidayDeleteDTO
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if input.ID == nil || *input.ID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "节假日ID不能为空", ctx.RequestID)
		return
	}
	if err := handler.service.DeleteSchoolHoliday(claims.UserID, *input.ID); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func (handler *Handler) resetSchoolHolidays(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	if err := handler.service.ResetSchoolHolidaysToDefault(claims.UserID); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}
