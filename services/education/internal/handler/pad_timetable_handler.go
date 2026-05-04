package handler

import (
	"net/http"
	"strconv"
	"strings"

	"go-migration-platform/pkg/httpx"
	"go-migration-platform/pkg/tenant"
	"go-migration-platform/services/education/internal/model"
)

func (handler *Handler) padTimetable(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	query := model.PadTimetableQueryDTO{
		StartDate:       strings.TrimSpace(r.URL.Query().Get("startDate")),
		EndDate:         strings.TrimSpace(r.URL.Query().Get("endDate")),
		PeriodGroupUUID: strings.TrimSpace(r.URL.Query().Get("periodGroupUuid")),
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("teacherId")); raw != "" {
		if value, err := strconv.ParseInt(raw, 10, 64); err == nil && value > 0 {
			query.TeacherID = value
		}
	}
	result, err := handler.service.GetPadTimetable(claims.UserID, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}
