package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"go-migration-platform/pkg/httpx"
	"go-migration-platform/pkg/tenant"
	"go-migration-platform/services/education/internal/model"
)

func (handler *Handler) padHomeSummary(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	query, err := parsePadHomeSummaryQuery(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	result, err := handler.service.GetPadHomeSummary(claims.UserID, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func parsePadHomeSummaryQuery(r *http.Request) (model.PadHomeSummaryQueryDTO, error) {
	values := r.URL.Query()
	dto := model.PadHomeSummaryQueryDTO{
		City: strings.TrimSpace(values.Get("city")),
	}
	latitudeText := firstNonEmptyQueryValue(values.Get("latitude"), values.Get("lat"))
	longitudeText := firstNonEmptyQueryValue(values.Get("longitude"), values.Get("lng"))
	if latitudeText == "" && longitudeText == "" {
		return dto, nil
	}
	if latitudeText == "" || longitudeText == "" {
		return model.PadHomeSummaryQueryDTO{}, errors.New("latitude and longitude must be provided together")
	}
	latitude, latErr := strconv.ParseFloat(latitudeText, 64)
	longitude, lngErr := strconv.ParseFloat(longitudeText, 64)
	if latErr != nil || lngErr != nil || latitude < -90 || latitude > 90 || longitude < -180 || longitude > 180 {
		return model.PadHomeSummaryQueryDTO{}, errors.New("latitude/longitude 参数格式不正确")
	}
	dto.Latitude = latitude
	dto.Longitude = longitude
	dto.HasCoordinates = true
	return dto, nil
}

func firstNonEmptyQueryValue(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
