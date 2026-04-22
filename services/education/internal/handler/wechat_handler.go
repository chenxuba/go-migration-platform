package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"go-migration-platform/pkg/httpx"
	"go-migration-platform/pkg/logx"
	"go-migration-platform/pkg/tenant"
	"go-migration-platform/services/education/internal/model"
)

func (handler *Handler) wechatOfficialCallback(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())

	signature := r.URL.Query().Get("signature")
	timestamp := r.URL.Query().Get("timestamp")
	nonce := r.URL.Query().Get("nonce")

	switch r.Method {
	case http.MethodGet:
		echoStr := r.URL.Query().Get("echostr")
		value, err := handler.service.VerifyWeChatOfficialCallback(signature, timestamp, nonce, echoStr)
		if err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, err.Error(), ctx.RequestID)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(value))
	case http.MethodPost:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
			return
		}
		if err := handler.service.HandleWeChatOfficialCallback(r.Context(), signature, timestamp, nonce, body, ctx.RequestID); err != nil {
			logx.Error("wechat official callback handling failed", logx.Entry{
				"requestId": ctx.RequestID,
				"error":     err.Error(),
			})
			httpx.WriteError(w, http.StatusUnauthorized, err.Error(), ctx.RequestID)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	default:
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
	}
}

func (handler *Handler) wechatOfficialBindTicketPreview(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	var dto model.WeChatOfficialBindTicketPreviewDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}

	result, err := handler.service.GetWeChatOfficialBindTicketPreview(r.Context(), dto)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) wechatOfficialBindTicketStudents(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	var dto model.WeChatOfficialBindStudentLookupDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}

	result, err := handler.service.LookupWeChatOfficialBindStudents(r.Context(), dto)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) wechatOfficialBindTicketConfirm(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	var dto model.WeChatOfficialConfirmStudentBindingDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}

	result, err := handler.service.ConfirmWeChatOfficialStudentBinding(r.Context(), dto)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}
