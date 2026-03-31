package handler

import (
	"encoding/json"
	"net/http"

	"go-migration-platform/pkg/httpx"
	"go-migration-platform/pkg/tenant"
)

type createComposeLessonBody struct {
	LessonName string   `json:"lessonName"`
	ProductIDs []string `json:"productIds"`
}

type pageComposeLessonBody struct {
	QueryModel struct {
		SearchKey string `json:"searchKey"`
	} `json:"queryModel"`
	PageRequestModel struct {
		NeedTotal bool `json:"needTotal"`
		SkipCount int  `json:"skipCount"`
		PageSize  int  `json:"pageSize"`
		PageIndex int  `json:"pageIndex"`
	} `json:"pageRequestModel"`
}

func (handler *Handler) createComposeLesson(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var body createComposeLessonBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	id, name, err := handler.service.CreateComposeLesson(claims.UserID, body.LessonName, body.ProductIDs)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"id": id, "name": name}, ctx.RequestID)
}

func (handler *Handler) pageComposeLessonsForPC(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var body pageComposeLessonBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	list, total, err := handler.service.PageComposeLessonsForPC(
		claims.UserID,
		body.QueryModel.SearchKey,
		body.PageRequestModel.PageIndex,
		body.PageRequestModel.PageSize,
		body.PageRequestModel.SkipCount,
	)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"list":  list,
		"total": total,
	}, ctx.RequestID)
}
