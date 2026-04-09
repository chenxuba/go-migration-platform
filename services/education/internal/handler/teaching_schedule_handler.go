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

func parseTeacherMatrixQuery(r *http.Request) model.TeachingScheduleListQueryDTO {
	query := model.TeachingScheduleListQueryDTO{
		StartDate:           strings.TrimSpace(r.URL.Query().Get("startDate")),
		EndDate:             strings.TrimSpace(r.URL.Query().Get("endDate")),
		StudentID:           strings.TrimSpace(r.URL.Query().Get("studentId")),
		MatrixTeacherFilter: strings.TrimSpace(strings.ToLower(r.URL.Query().Get("teacherFilter"))),
		PeriodGroupUUID:     strings.TrimSpace(r.URL.Query().Get("periodGroupUuid")),
	}
	parseInt64CSV := func(raw string) []int64 {
		out := make([]int64, 0)
		for _, p := range strings.Split(strings.TrimSpace(raw), ",") {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			if v, err := strconv.ParseInt(p, 10, 64); err == nil && v > 0 {
				out = append(out, v)
			}
		}
		return out
	}
	parseStringCSV := func(raw string) []string {
		out := make([]string, 0)
		for _, p := range strings.Split(strings.TrimSpace(raw), ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}
		return out
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("matrixTeacherIds")); raw != "" {
		for _, p := range strings.Split(raw, ",") {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			if v, err := strconv.ParseInt(p, 10, 64); err == nil && v > 0 {
				query.MatrixTeacherIDs = append(query.MatrixTeacherIDs, v)
			}
		}
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("classType")); raw != "" {
		if value, err := strconv.Atoi(raw); err == nil && value > 0 {
			query.ClassType = &value
		}
	}
	if w := strings.TrimSpace(r.URL.Query().Get("weekdays")); w != "" {
		for _, p := range strings.Split(w, ",") {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			v, err := strconv.Atoi(p)
			if err != nil || v < 1 || v > 7 {
				continue
			}
			query.MatrixWeekdays = append(query.MatrixWeekdays, v)
		}
	}
	query.ScheduleTeacherIDs = parseInt64CSV(r.URL.Query().Get("scheduleTeacherIds"))
	query.ClassroomIDs = parseInt64CSV(r.URL.Query().Get("classroomIds"))
	query.GroupClassIDs = parseInt64CSV(r.URL.Query().Get("groupClassIds"))
	query.OneToOneClassIDs = parseInt64CSV(r.URL.Query().Get("oneToOneClassIds"))
	query.LessonIDs = parseInt64CSV(r.URL.Query().Get("lessonIds"))
	query.ScheduleTypeFilters = parseStringCSV(r.URL.Query().Get("scheduleTypes"))
	query.CallStatusFilters = parseStringCSV(r.URL.Query().Get("callStatuses"))
	return query
}

func parseTeachingScheduleListQuery(r *http.Request) model.TeachingScheduleListQueryDTO {
	query := model.TeachingScheduleListQueryDTO{
		StartDate: strings.TrimSpace(r.URL.Query().Get("startDate")),
		EndDate:   strings.TrimSpace(r.URL.Query().Get("endDate")),
		StudentID: strings.TrimSpace(r.URL.Query().Get("studentId")),
	}
	parseInt64CSV := func(raw string) []int64 {
		out := make([]int64, 0)
		for _, p := range strings.Split(strings.TrimSpace(raw), ",") {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			if v, err := strconv.ParseInt(p, 10, 64); err == nil && v > 0 {
				out = append(out, v)
			}
		}
		return out
	}
	parseStringCSV := func(raw string) []string {
		out := make([]string, 0)
		for _, p := range strings.Split(strings.TrimSpace(raw), ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}
		return out
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("classType")); raw != "" {
		if value, err := strconv.Atoi(raw); err == nil && value > 0 {
			query.ClassType = &value
		}
	}
	query.ScheduleTeacherIDs = parseInt64CSV(r.URL.Query().Get("scheduleTeacherIds"))
	query.ClassroomIDs = parseInt64CSV(r.URL.Query().Get("classroomIds"))
	query.GroupClassIDs = parseInt64CSV(r.URL.Query().Get("groupClassIds"))
	query.OneToOneClassIDs = parseInt64CSV(r.URL.Query().Get("oneToOneClassIds"))
	query.LessonIDs = parseInt64CSV(r.URL.Query().Get("lessonIds"))
	query.ConflictTypes = parseStringCSV(r.URL.Query().Get("conflictTypes"))
	query.ScheduleTypeFilters = parseStringCSV(r.URL.Query().Get("scheduleTypes"))
	query.CallStatusFilters = parseStringCSV(r.URL.Query().Get("callStatuses"))
	return query
}

func (handler *Handler) createOneToOneSchedules(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var dto model.CreateOneToOneSchedulesDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.CreateOneToOneSchedules(claims.UserID, dto)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) createGroupClassSchedules(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var dto model.CreateGroupClassSchedulesDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.CreateGroupClassSchedules(claims.UserID, dto)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) validateOneToOneSchedules(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var dto model.CreateOneToOneSchedulesDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.ValidateOneToOneSchedules(claims.UserID, dto)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) validateGroupClassSchedules(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var dto model.CreateGroupClassSchedulesDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.ValidateGroupClassSchedules(claims.UserID, dto)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) checkOneToOneScheduleAvailability(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var dto model.CheckOneToOneScheduleAvailabilityDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.CheckOneToOneScheduleAvailability(claims.UserID, dto)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) checkAssistantScheduleAvailability(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var dto model.CheckAssistantScheduleAvailabilityDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.CheckAssistantScheduleAvailability(claims.UserID, dto)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) teachingScheduleBatchDetail(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	query := model.TeachingScheduleBatchDetailQueryDTO{
		BatchNo: strings.TrimSpace(r.URL.Query().Get("batchNo")),
	}
	for _, raw := range strings.Split(strings.TrimSpace(r.URL.Query().Get("ids")), ",") {
		value := strings.TrimSpace(raw)
		if value == "" {
			continue
		}
		query.IDs = append(query.IDs, value)
	}
	if id := strings.TrimSpace(r.URL.Query().Get("id")); id != "" {
		query.IDs = append(query.IDs, id)
	}
	result, err := handler.service.GetTeachingScheduleBatchDetail(claims.UserID, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) teachingScheduleConflictDetail(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	query := model.TeachingScheduleConflictDetailQueryDTO{
		ID: strings.TrimSpace(r.URL.Query().Get("id")),
	}
	result, err := handler.service.GetTeachingScheduleConflictDetail(claims.UserID, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) teachingScheduleDetail(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	query := model.TeachingScheduleDetailQueryDTO{
		ID: strings.TrimSpace(r.URL.Query().Get("id")),
	}
	result, err := handler.service.GetTeachingScheduleDetail(claims.UserID, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) removeTeachingScheduleStudentCurrent(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var dto model.TeachingScheduleStudentRemoveCurrentDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if err := handler.service.RemoveTeachingScheduleStudentCurrent(claims.UserID, dto); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, true, ctx.RequestID)
}

func (handler *Handler) teachingSchedules(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	query := parseTeachingScheduleListQuery(r)
	result, err := handler.service.ListTeachingSchedules(claims.UserID, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) teachingSchedulesByTeacherMatrix(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	query := parseTeacherMatrixQuery(r)
	result, err := handler.service.ListTeachingSchedulesByTeacherMatrix(claims.UserID, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) teachingSchedulesTeacherMatrixExport(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	query := parseTeacherMatrixQuery(r)
	buf, filename, err := handler.service.ExportTeachingSchedulesTeacherMatrixExcel(claims.UserID, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	// RFC 5987：UTF-8 文件名必须用 filename*，否则许多客户端会把中文当 Latin-1 显示成乱码
	w.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+url.QueryEscape(filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf)
}

func (handler *Handler) smartTeachingSchedulesExport(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	query := parseTeacherMatrixQuery(r)
	viewMode := strings.TrimSpace(r.URL.Query().Get("viewMode"))
	buf, filename, err := handler.service.ExportSmartTimetableExcel(claims.UserID, query, viewMode)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+url.QueryEscape(filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf)
}

func (handler *Handler) timeTeachingSchedulesExport(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	query := parseTeachingScheduleListQuery(r)
	buf, filename, err := handler.service.ExportTimeTimetableExcel(claims.UserID, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+url.QueryEscape(filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf)
}

func (handler *Handler) clearAllTeachingSchedules(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var body struct {
		Confirm bool `json:"confirm"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || !body.Confirm {
		httpx.WriteError(w, http.StatusBadRequest, "请传 JSON：{\"confirm\":true} 以确认清空本机构全部排课", ctx.RequestID)
		return
	}
	n, err := handler.service.ClearAllTeachingSchedules(claims.UserID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"deleted": n}, ctx.RequestID)
}

func (handler *Handler) replaceTeachingScheduleBatch(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var dto model.TeachingScheduleBatchReplaceDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.ReplaceTeachingScheduleBatch(claims.UserID, dto)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) batchUpdateTeachingSchedules(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var dto model.TeachingScheduleBatchUpdateDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if err := handler.service.BatchUpdateTeachingSchedules(claims.UserID, dto); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func (handler *Handler) cancelTeachingSchedules(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var dto model.TeachingScheduleCancelDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.CancelTeachingSchedules(claims.UserID, dto)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) copyTeachingSchedulesWeek(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var dto model.TeachingScheduleCopyWeekDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.CopyTeachingSchedulesWeek(claims.UserID, dto)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}
