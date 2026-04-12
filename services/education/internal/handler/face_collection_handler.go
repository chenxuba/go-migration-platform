package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"go-migration-platform/pkg/httpx"
	"go-migration-platform/pkg/tenant"
	"go-migration-platform/services/education/internal/model"
)

func (handler *Handler) pageFaceCollectionStudents(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var query model.FaceCollectionStudentQueryDTO
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.PageFaceCollectionStudents(claims.UserID, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	for idx := range result.Items {
		result.Items[idx].AvatarURL = normalizeStudentAvatar(result.Items[idx].AvatarURL, nil)
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) getFaceCollectionProfile(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	studentID, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("studentId")), 10, 64)
	if err != nil || studentID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid studentId", ctx.RequestID)
		return
	}
	result, err := handler.service.GetFaceCollectionProfile(claims.UserID, studentID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	result.FaceDescriptor = nil
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) listFaceCollectionProfiles(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	result, err := handler.service.ListFaceCollectionProfiles(claims.UserID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	for idx := range result {
		result[idx].FaceDescriptor = nil
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) compareFaceCollectionProfile(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var dto model.FaceCollectionCompareDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.CompareFaceCollectionProfile(claims.UserID, dto)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) saveFaceCollectionProfile(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var dto model.FaceCollectionProfileSaveDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if err := handler.service.SaveFaceCollectionProfile(claims.UserID, dto); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, true, ctx.RequestID)
}

func (handler *Handler) deleteFaceCollectionProfile(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var dto model.FaceCollectionProfileDeleteDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if err := handler.service.DeleteFaceCollectionProfile(claims.UserID, dto.StudentID); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, true, ctx.RequestID)
}

func (handler *Handler) listFaceAttendanceSessions(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	limit := 50
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err == nil && parsed > 0 {
			limit = parsed
		}
	}
	result, err := handler.service.ListFaceAttendanceSessions(claims.UserID, limit)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	for idx := range result {
		result[idx].AvatarURL = normalizeStudentAvatar(result[idx].AvatarURL, nil)
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) recognizeFaceAttendanceSession(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var dto model.FaceAttendanceSessionRecognizeDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.RecognizeFaceAttendanceSession(claims.UserID, dto)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	result.AvatarURL = normalizeStudentAvatar(result.AvatarURL, nil)
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) commitFaceAttendanceSession(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var dto model.FaceAttendanceSessionCommitDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.CommitFaceAttendanceSession(claims.UserID, dto)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	result.AvatarURL = normalizeStudentAvatar(result.AvatarURL, nil)
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) listFaceAttendanceRecords(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	limit := 50
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err == nil && parsed > 0 {
			limit = parsed
		}
	}
	result, err := handler.service.ListFaceAttendanceRecords(claims.UserID, limit)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) saveFaceAttendanceRecord(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var dto model.FaceAttendanceRecordSaveDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.SaveFaceAttendanceRecord(claims.UserID, dto)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}
