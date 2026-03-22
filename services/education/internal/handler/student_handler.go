package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"go-migration-platform/pkg/httpx"
	"go-migration-platform/pkg/tenant"
	"go-migration-platform/services/education/internal/model"
)

var (
	intentLevelChangePattern     = regexp.MustCompile(`意向度从"([^"]+)"修改为"([^"]+)"`)
	followUpStatusChangePattern  = regexp.MustCompile(`跟进状态从"([^"]+)"修改为"([^"]+)"`)
	phoneRelationshipChangeRegex = regexp.MustCompile(`手机关联人关系(?:从|:)"([^"]+)"修改为"([^"]+)"`)
	visitStatusChangePattern     = regexp.MustCompile(`回访状态从"([^"]+)"修改为"([^"]+)"`)
	mobileChangePattern          = regexp.MustCompile(`手机号码从"([^"]+)"修改为"([^"]+)"`)
)

func (handler *Handler) intentStudentsPage(w http.ResponseWriter, r *http.Request) {
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
	query := parseIntentStudentQueryDTO(raw)

	result, err := handler.service.PageIntentStudents(claims.UserID, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	for idx := range result.Items {
		result.Items[idx].AvatarURL = normalizeStudentAvatar(result.Items[idx].AvatarURL, result.Items[idx].StuSex)
		result.Items[idx].Mobile = maskPhone(result.Items[idx].Mobile)
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) intentStudentDetail(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	studentID, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("studentId")), 10, 64)
	if err != nil || studentID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid studentId", ctx.RequestID)
		return
	}

	item, err := handler.service.GetIntentStudentDetail(claims.UserID, studentID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, formatIntentStudentDetail(item), ctx.RequestID)
}

func (handler *Handler) currentStudentsPage(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var query model.CurrentStudentQueryDTO
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.PageCurrentStudents(claims.UserID, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) enrolledStudentsPage(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var query model.EnrolledStudentQueryDTO
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.PageEnrolledStudents(claims.UserID, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) updateStudentStatus(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var dto model.StudentStatusUpdateDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if err := handler.service.UpdateStudentStatus(claims.UserID, dto); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "ok"}, ctx.RequestID)
}

func (handler *Handler) batchAssignSalesperson(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var dto model.BatchCommonDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if err := handler.service.BatchAssignSalesperson(claims.UserID, dto); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func (handler *Handler) batchTransferToPublicPool(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var dto model.BatchCommonDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if err := handler.service.BatchTransferToPublicPool(claims.UserID, dto); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func (handler *Handler) batchDeleteIntentStudents(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var dto model.BatchCommonDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if err := handler.service.BatchDeleteIntentStudents(claims.UserID, dto); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func (handler *Handler) addIntentStudent(w http.ResponseWriter, r *http.Request) {
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
	dto := parseStudentSaveDTO(raw)
	id, err := handler.service.AddIntentStudent(claims.UserID, dto)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"studentId": id}, ctx.RequestID)
}

func (handler *Handler) updateIntentStudent(w http.ResponseWriter, r *http.Request) {
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
	dto := parseStudentSaveDTO(raw)
	if err := handler.service.UpdateIntentStudent(claims.UserID, dto); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func (handler *Handler) checkStudentRepeat(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req model.StudentDuplicateCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.CheckStudentRepeat(claims.UserID, req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) checkStudentTips(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req model.StudentDuplicateCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.CheckStudentTips(claims.UserID, req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) getStudentPhoneNumber(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	studentID, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("studentId")), 10, 64)
	if err != nil || studentID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid studentId", ctx.RequestID)
		return
	}
	phone, err := handler.service.GetStudentPhoneNumber(claims.UserID, studentID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, phone, ctx.RequestID)
}

func (handler *Handler) recommendersPage(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var query model.RecommenderQueryDTO
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.PageRecommenders(claims.UserID, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) birthdayStudentsPage(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var query model.BirthdayStudentQueryDTO
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.PageBirthdayStudents(claims.UserID, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) studentChangeRecords(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	stuID, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("stuId")), 10, 64)
	if err != nil || stuID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid stuId", ctx.RequestID)
		return
	}
	result, err := handler.service.ListStudentChangeRecords(claims.UserID, stuID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	for idx := range result {
		result[idx].ChangeContent = formatStudentChangeContent(result[idx].ChangeContent)
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func formatStudentChangeContent(content string) string {
	if strings.TrimSpace(content) == "" {
		return content
	}

	result := content
	result = replaceLabeledChange(result, intentLevelChangePattern, "意向度", map[string]string{
		"1": "未知",
		"2": "低",
		"3": "中",
		"4": "高",
	})
	result = replaceLabeledChange(result, followUpStatusChangePattern, "跟进状态", map[string]string{
		"0": "待跟进",
		"1": "跟进中",
		"2": "未接听",
		"3": "已邀约",
		"4": "已试听",
		"5": "已到访",
		"6": "已失效",
	})
	result = replaceLabeledChange(result, phoneRelationshipChangeRegex, "手机关联人关系", map[string]string{
		"1": "爸爸",
		"2": "妈妈",
		"3": "爷爷",
		"4": "奶奶",
		"5": "外公",
		"6": "外婆",
		"7": "其他",
	})
	result = replaceLabeledChange(result, visitStatusChangePattern, "回访状态", map[string]string{
		"0": "未回访",
		"1": "已回访",
	})
	result = mobileChangePattern.ReplaceAllStringFunc(result, func(segment string) string {
		matches := mobileChangePattern.FindStringSubmatch(segment)
		if len(matches) != 3 {
			return segment
		}
		return fmt.Sprintf(`手机号码从"%s"修改为"%s"`, maskChangeMobile(matches[1]), maskChangeMobile(matches[2]))
	})
	return result
}

func replaceLabeledChange(content string, pattern *regexp.Regexp, fieldName string, labels map[string]string) string {
	return pattern.ReplaceAllStringFunc(content, func(segment string) string {
		matches := pattern.FindStringSubmatch(segment)
		if len(matches) != 3 {
			return segment
		}
		return fmt.Sprintf(`%s从"%s"修改为"%s"`, fieldName, lookupChangeLabel(labels, matches[1]), lookupChangeLabel(labels, matches[2]))
	})
}

func lookupChangeLabel(labels map[string]string, raw string) string {
	value := strings.TrimSpace(raw)
	if label, ok := labels[value]; ok {
		return label
	}
	return value
}

func maskChangeMobile(value string) string {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) == 11 && strings.IndexFunc(trimmed, func(r rune) bool { return r < '0' || r > '9' }) == -1 {
		return trimmed[:3] + "****" + trimmed[7:]
	}
	return trimmed
}
