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
)

func (handler *Handler) rechargeAccountDetailPage(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var raw map[string]any
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	query := parseRechargeAccountDetailQueryDTO(raw)
	result, err := handler.service.GetRechargeAccountDetailPage(claims.UserID, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) rechargeAccountExpendIncome(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	params := r.URL.Query()
	query := model.RechargeAccountDetailQuery{
		StudentID:         params.Get("studentId"),
		RechargeAccountID: params.Get("rechargeAccountId"),
		StartTime:         params.Get("startTime"),
		EndTime:           params.Get("endTime"),
		FlowTypes:         parseRechargeAccountDetailFlowTypes(params),
	}
	result, err := handler.service.GetRechargeAccountExpendIncome(claims.UserID, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) rechargeAccountDetailExport(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var raw map[string]any
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	query := parseRechargeAccountDetailQueryDTO(raw)
	content, fileName, err := handler.service.ExportRechargeAccountDetails(claims.UserID, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+url.QueryEscape(fileName))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

func parseRechargeAccountDetailFlowTypes(values url.Values) []int {
	keys := []string{"flowTypes", "flowTypes[]"}
	result := make([]int, 0)
	seen := make(map[int]struct{})
	for _, key := range keys {
		for _, raw := range values[key] {
			for _, part := range strings.Split(raw, ",") {
				item := strings.TrimSpace(part)
				if item == "" {
					continue
				}
				value, err := strconv.Atoi(item)
				if err != nil {
					continue
				}
				if _, ok := seen[value]; ok {
					continue
				}
				seen[value] = struct{}{}
				result = append(result, value)
			}
		}
	}
	return result
}
