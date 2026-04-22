package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

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

func (handler *Handler) parentBoundStudents(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireParentAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	result, err := handler.service.ListParentBoundStudentsByPhone(r.Context(), claims.Username)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) parentPendingStudents(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireParentAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	result, err := handler.service.ListParentPendingStudentsByPhone(r.Context(), claims.Username)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) parentConfirmStudents(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireParentAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	var dto model.ParentBindStudentsDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}

	result, err := handler.service.ConfirmParentStudentsByPhone(r.Context(), claims.Username, dto)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) parentCampuses(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireParentAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	result, err := handler.service.ListParentCampusesByPhone(r.Context(), claims.Username)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) parentSchedules(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireParentAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	query := model.ParentScheduleQueryDTO{
		StartDate: strings.TrimSpace(r.URL.Query().Get("startDate")),
		EndDate:   strings.TrimSpace(r.URL.Query().Get("endDate")),
	}
	result, err := handler.service.ListParentSchedulesByPhone(r.Context(), claims.Username, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) parentScheduleDates(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireParentAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	query := model.ParentScheduleQueryDTO{
		StartDate: strings.TrimSpace(r.URL.Query().Get("startDate")),
		EndDate:   strings.TrimSpace(r.URL.Query().Get("endDate")),
	}
	result, err := handler.service.ListParentScheduleDatesByPhone(r.Context(), claims.Username, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) parentLeaves(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireParentAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	pageSize, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("pageSize")))
	pageIndex, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("pageIndex")))
	query := model.ParentLeaveQueryDTO{
		StudentID: strings.TrimSpace(r.URL.Query().Get("studentId")),
		PageIndex: pageIndex,
		PageSize:  pageSize,
	}
	result, err := handler.service.ListParentLeavesByPhone(r.Context(), claims.Username, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) parentClassRecords(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireParentAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	pageSize, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("pageSize")))
	pageIndex, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("pageIndex")))
	query := model.ParentClassRecordQueryDTO{
		StudentID: strings.TrimSpace(r.URL.Query().Get("studentId")),
		PageIndex: pageIndex,
		PageSize:  pageSize,
	}
	result, err := handler.service.ListParentClassRecordsByPhone(r.Context(), claims.Username, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) parentClassRecordDetail(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireParentAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	query := model.ParentClassRecordDetailQueryDTO{
		StudentID:               strings.TrimSpace(r.URL.Query().Get("studentId")),
		StudentTeachingRecordID: strings.TrimSpace(r.URL.Query().Get("studentTeachingRecordId")),
	}
	result, err := handler.service.GetParentClassRecordDetailByPhone(r.Context(), claims.Username, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) parentCourseEnrollments(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireParentAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	query := model.ParentCourseEnrollmentQueryDTO{
		StudentID: strings.TrimSpace(r.URL.Query().Get("studentId")),
	}
	result, err := handler.service.ListParentCourseEnrollmentsByPhone(r.Context(), claims.Username, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) parentCourseEnrollmentDetail(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireParentAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	pageSize, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("pageSize")))
	pageIndex, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("pageIndex")))
	chargingMode, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("chargingMode")))
	query := model.ParentCourseEnrollmentDetailQueryDTO{
		StudentID:    strings.TrimSpace(r.URL.Query().Get("studentId")),
		LessonID:     strings.TrimSpace(r.URL.Query().Get("lessonId")),
		ChargingMode: chargingMode,
		PageIndex:    pageIndex,
		PageSize:     pageSize,
	}
	result, err := handler.service.GetParentCourseEnrollmentDetailByPhone(r.Context(), claims.Username, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) parentCourseArrears(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireParentAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	chargingMode, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("chargingMode")))
	query := model.ParentCourseArrearQueryDTO{
		StudentID:    strings.TrimSpace(r.URL.Query().Get("studentId")),
		LessonID:     strings.TrimSpace(r.URL.Query().Get("lessonId")),
		ChargingMode: chargingMode,
	}
	result, err := handler.service.ListParentCourseArrearsByPhone(r.Context(), claims.Username, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) parentRehabRecords(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireParentAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	pageSize, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("pageSize")))
	pageIndex, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("pageIndex")))
	query := model.ParentRehabRecordQueryDTO{
		StudentID: strings.TrimSpace(r.URL.Query().Get("studentId")),
		PageIndex: pageIndex,
		PageSize:  pageSize,
	}
	result, err := handler.service.ListParentRehabRecordsByPhone(r.Context(), claims.Username, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) parentRehabRecordDetail(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireParentAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	query := model.ParentRehabRecordDetailQueryDTO{
		StudentID:               strings.TrimSpace(r.URL.Query().Get("studentId")),
		StudentTeachingRecordID: strings.TrimSpace(r.URL.Query().Get("studentTeachingRecordId")),
	}
	result, err := handler.service.GetParentRehabRecordDetailByPhone(r.Context(), claims.Username, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) parentRehabRecordFeedback(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireParentAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	var dto model.ParentRehabFeedbackSaveDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}

	result, err := handler.service.SaveParentRehabFeedbackByPhone(r.Context(), claims.Username, dto)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) parentRehabRecordSignatureUpload(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireParentAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid multipart form", ctx.RequestID)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "file is required", ctx.RequestID)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "read file failed", ctx.RequestID)
		return
	}

	result, err := handler.service.UploadParentRehabSignatureByPhone(
		r.Context(),
		claims.Username,
		header.Filename,
		header.Header.Get("Content-Type"),
		data,
	)
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
