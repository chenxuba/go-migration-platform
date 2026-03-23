package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"go-migration-platform/pkg/authx"
	"go-migration-platform/pkg/httpx"
	"go-migration-platform/pkg/tenant"
	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/service"
)

type Handler struct {
	service *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{service: svc}
}

func (handler *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/health", handler.health)
	mux.HandleFunc("/api/v1/students", handler.students)
	mux.HandleFunc("/api/v1/orders", handler.orders)
	mux.HandleFunc("/api/v1/inst-config", handler.getInstConfig)
	mux.HandleFunc("/api/v1/inst-config/update", handler.setInstConfig)
	mux.HandleFunc("/api/v1/inst-config/init-all", handler.initInstAllConfig)
	mux.HandleFunc("/api/v1/approval-configs/save", handler.saveApprovalConfig)
	mux.HandleFunc("/api/v1/approvals/all-paged-list", handler.approvalAllPagedList)
	mux.HandleFunc("/api/v1/approvals/my-paged-list", handler.approvalMyPagedList)
	mux.HandleFunc("/api/v1/approvals/my-approve-statistics-count", handler.approvalMyStatisticsCount)
	mux.HandleFunc("/api/v1/approvals/my-initiated-paged-list", handler.approvalMyInitiatedPagedList)
	mux.HandleFunc("/api/v1/approvals/detail", handler.approvalDetail)
	mux.HandleFunc("/api/v1/approval-templates/list", handler.approvalTemplatesList)
	mux.HandleFunc("/api/v1/approval-templates/save", handler.saveApprovalTemplates)
	mux.HandleFunc("/api/v1/approvals/approve", handler.approveApprovalRecord)
	mux.HandleFunc("/api/v1/approvals/reject", handler.rejectApprovalRecord)
	mux.HandleFunc("/api/v1/approvals/cancel", handler.cancelApprovalRecord)
	mux.HandleFunc("/api/v1/staffs/summaries", handler.staffSummaries)
	mux.HandleFunc("/api/v1/student-field-keys/default", handler.defaultStudentFields)
	mux.HandleFunc("/api/v1/student-field-keys/custom", handler.customStudentFields)
	mux.HandleFunc("/api/v1/student-field-keys/detail", handler.studentFieldDetail)
	mux.HandleFunc("/api/v1/student-field-keys/sort", handler.sortCustomStudentFields)
	mux.HandleFunc("/api/v1/student-field-keys/init", handler.initInstStudentField)
	mux.HandleFunc("/api/v1/student-field-keys/display-status", handler.updateStudentFieldDisplayStatus)
	mux.HandleFunc("/api/v1/student-field-keys/create", handler.addCustomStudentField)
	mux.HandleFunc("/api/v1/student-field-keys/update", handler.updateCustomStudentField)
	mux.HandleFunc("/api/v1/student-field-keys/delete", handler.deleteCustomStudentField)
	mux.HandleFunc("/api/v1/inst-users/page", handler.instUsersPage)
	mux.HandleFunc("/api/v1/inst-users/detail", handler.instUserDetail)
	mux.HandleFunc("/api/v1/inst-users/create", handler.saveInstUser)
	mux.HandleFunc("/api/v1/inst-users/update", handler.updateInstUser)
	mux.HandleFunc("/api/v1/inst-users/batch-disabled", handler.batchDisabledInstUsers)
	mux.HandleFunc("/api/v1/inst-users/batch-dept", handler.batchModifyInstUserDept)
	mux.HandleFunc("/api/v1/inst-users/batch-role", handler.batchModifyInstUserRole)
	mux.HandleFunc("/api/v1/inst-users/check-phone", handler.checkInstUserPhoneUsed)
	mux.HandleFunc("/api/v1/inst-users/change-phone", handler.changeInstUserPhone)
	mux.HandleFunc("/api/v1/qiniu/upload-token", handler.qiniuUploadToken)
	mux.HandleFunc("/api/v1/qiniu/video-upload-token", handler.qiniuVideoUploadToken)
	mux.HandleFunc("/api/v1/tuition-accounts/reading-list", handler.tuitionAccountReadingList)
	mux.HandleFunc("/api/v1/course-properties", handler.coursePropertyList)
	mux.HandleFunc("/api/v1/course-properties/update", handler.updateCourseProperty)
	mux.HandleFunc("/api/v1/course-properties/init", handler.initInstCourseProperty)
	mux.HandleFunc("/api/v1/course-property-options", handler.coursePropertyOptions)
	mux.HandleFunc("/api/v1/course-property-options/create", handler.addCoursePropertyOption)
	mux.HandleFunc("/api/v1/course-property-options/update", handler.updateCoursePropertyOption)
	mux.HandleFunc("/api/v1/course-property-options/delete", handler.deleteCoursePropertyOption)
	mux.HandleFunc("/api/v1/course-property-options/sort", handler.updateCoursePropertyOptionSort)
	mux.HandleFunc("/api/v1/channels/default", handler.defaultChannels)
	mux.HandleFunc("/api/v1/channels", handler.channels)
	mux.HandleFunc("/api/v1/channels/grouped", handler.channelsWithGroups)
	mux.HandleFunc("/api/v1/channel-categories", handler.channelCategories)
	mux.HandleFunc("/api/v1/channels/pc-page", handler.channelPCPage)
	mux.HandleFunc("/api/v1/channel-tree", handler.channelTree)
	mux.HandleFunc("/api/v1/channel-categories/create", handler.addChannelCategory)
	mux.HandleFunc("/api/v1/channel-categories/update", handler.updateChannelCategory)
	mux.HandleFunc("/api/v1/channel-categories/delete", handler.deleteChannelCategory)
	mux.HandleFunc("/api/v1/channels/status", handler.updateChannelStatus)
	mux.HandleFunc("/api/v1/channels/create", handler.addChannel)
	mux.HandleFunc("/api/v1/channels/update", handler.updateChannel)
	mux.HandleFunc("/api/v1/channels/adjust", handler.adjustChannels)
	mux.HandleFunc("/api/v1/course-categories/page", handler.courseCategoriesPage)
	mux.HandleFunc("/api/v1/course-categories/create", handler.addCourseCategory)
	mux.HandleFunc("/api/v1/course-categories/update", handler.updateCourseCategory)
	mux.HandleFunc("/api/v1/course-categories/delete", handler.deleteCourseCategory)
	mux.HandleFunc("/api/v1/courses/page", handler.coursesPage)
	mux.HandleFunc("/api/v1/courses/options", handler.courseIDNamesPage)
	mux.HandleFunc("/api/v1/courses/detail", handler.courseDetail)
	mux.HandleFunc("/api/v1/courses/create", handler.addCourse)
	mux.HandleFunc("/api/v1/courses/update", handler.updateCourse)
	mux.HandleFunc("/api/v1/courses/process-content", handler.processContentPage)
	mux.HandleFunc("/api/v1/courses/delete-restore", handler.batchDeleteOrRestoreCourses)
	mux.HandleFunc("/api/v1/courses/sale-status", handler.batchSaleStatus)
	mux.HandleFunc("/api/v1/courses/micro-school-show", handler.batchOpenMicroSchoolShow)
	mux.HandleFunc("/api/v1/infrastructure/status", handler.infrastructureStatus)
	mux.HandleFunc("/api/v1/mq/event-logs", handler.mqEventLogs)
	mux.HandleFunc("/api/v1/es-sync/intent-student/sync", handler.syncIntentStudents)
	mux.HandleFunc("/api/v1/es-sync/intent-student/rebuild", handler.rebuildIntentStudents)
	mux.HandleFunc("/api/v1/es-sync/intent-student/clear", handler.clearIntentStudents)
	mux.HandleFunc("/api/v1/intent-students/page", handler.intentStudentsPage)
	mux.HandleFunc("/api/v1/intent-students/detail", handler.intentStudentDetail)
	mux.HandleFunc("/api/v1/current-students/page", handler.currentStudentsPage)
	mux.HandleFunc("/api/v1/enrolled-students/page", handler.enrolledStudentsPage)
	mux.HandleFunc("/api/v1/orders/list", handler.orderList)
	mux.HandleFunc("/api/v1/orders/detail", handler.orderDetail)
	mux.HandleFunc("/api/v1/orders/check-quote", handler.checkQuoteInfo)
	mux.HandleFunc("/api/v1/orders/calc-enroll-type", handler.calcCourseEnrollType)
	mux.HandleFunc("/api/v1/orders/create", handler.createOrder)
	mux.HandleFunc("/api/v1/orders/pay", handler.payOrder)
	mux.HandleFunc("/api/v1/orders/registration-list", handler.registrationListPage)
	mux.HandleFunc("/api/v1/orders/set-bad-debt", handler.setBadDebt)
	mux.HandleFunc("/api/v1/orders/cancel-bad-debt", handler.cancelBadDebt)
	mux.HandleFunc("/api/v1/follow-records/page", handler.followRecordsPage)
	mux.HandleFunc("/api/v1/follow-records/create", handler.createFollowUp)
	mux.HandleFunc("/api/v1/follow-records/count", handler.followUpCount)
	mux.HandleFunc("/api/v1/follow-records/visit-status", handler.updateVisitStatus)
	mux.HandleFunc("/api/v1/follow-records/update", handler.updateFollowUpRecord)
	mux.HandleFunc("/api/v1/follow-records/statistics", handler.followUpRecordStatistics)
	mux.HandleFunc("/api/v1/intent-students/status", handler.updateStudentStatus)
	mux.HandleFunc("/api/v1/intent-students/assign-sales", handler.batchAssignSalesperson)
	mux.HandleFunc("/api/v1/intent-students/public-pool", handler.batchTransferToPublicPool)
	mux.HandleFunc("/api/v1/intent-students/delete", handler.batchDeleteIntentStudents)
	mux.HandleFunc("/api/v1/intent-students/create", handler.addIntentStudent)
	mux.HandleFunc("/api/v1/intent-students/update", handler.updateIntentStudent)
	mux.HandleFunc("/api/v1/intent-students/check-repeat", handler.checkStudentRepeat)
	mux.HandleFunc("/api/v1/intent-students/check-tips", handler.checkStudentTips)
	mux.HandleFunc("/api/v1/intent-students/phone", handler.getStudentPhoneNumber)
	mux.HandleFunc("/api/v1/recommenders/page", handler.recommendersPage)
	mux.HandleFunc("/api/v1/birthday-students/page", handler.birthdayStudentsPage)
	mux.HandleFunc("/api/v1/students/change-records", handler.studentChangeRecords)
}

func (handler *Handler) health(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"service": "education-service", "status": "ok"}, ctx.RequestID)
}

func (handler *Handler) students(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	httpx.WriteJSON(w, http.StatusOK, handler.service.Students(ctx), ctx.RequestID)
}

func (handler *Handler) orders(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	httpx.WriteJSON(w, http.StatusOK, handler.service.Orders(ctx), ctx.RequestID)
}

func (handler *Handler) getInstConfig(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	result, err := handler.service.GetInstConfig(claims.UserID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) setInstConfig(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if err := handler.service.SetInstConfig(claims.UserID, payload); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func (handler *Handler) initInstAllConfig(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	instID := derefInt64Value(asInt64Ptr(payload["id"]))
	if instID <= 0 {
		instID = derefInt64Value(asInt64Ptr(payload["instId"]))
	}
	if err := handler.service.InitInstAllConfig(instID); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func (handler *Handler) defaultStudentFields(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	result, err := handler.service.GetDefaultStudentFields(claims.UserID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) customStudentFields(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	result, err := handler.service.GetCustomStudentFields(claims.UserID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) updateStudentFieldDisplayStatus(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var input model.StudentFieldKey
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if err := handler.service.UpdateStudentFieldDisplayStatus(claims.UserID, input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func (handler *Handler) addCustomStudentField(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var input model.StudentFieldKey
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	id, err := handler.service.AddCustomStudentField(claims.UserID, input)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"id": id}, ctx.RequestID)
}

func (handler *Handler) updateCustomStudentField(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var input model.StudentFieldKey
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if err := handler.service.UpdateCustomStudentField(claims.UserID, input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func (handler *Handler) deleteCustomStudentField(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var input model.StudentFieldKey
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if err := handler.service.DeleteCustomStudentField(claims.UserID, input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func (handler *Handler) sortCustomStudentFields(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var fields []model.StudentFieldKey
	if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if err := handler.service.SortCustomStudentFields(claims.UserID, fields); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func (handler *Handler) studentFieldDetail(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	id, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("id")), 10, 64)
	if err != nil || id <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id", ctx.RequestID)
		return
	}
	result, err := handler.service.GetStudentFieldDetail(claims.UserID, id)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) initInstStudentField(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	instID, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("instId")), 10, 64)
	if err != nil || instID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid instId", ctx.RequestID)
		return
	}
	if err := handler.service.InitInstStudentField(instID); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func (handler *Handler) initInstCourseProperty(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	instID, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("instId")), 10, 64)
	if err != nil || instID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid instId", ctx.RequestID)
		return
	}
	if err := handler.service.InitInstCourseProperty(instID); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func (handler *Handler) qiniuUploadToken(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	result, err := handler.service.GetQiniuUploadToken()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) qiniuVideoUploadToken(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	result, err := handler.service.GetQiniuVideoUploadToken()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) tuitionAccountReadingList(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var query model.TuitionAccountReadingListQueryDTO
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.GetTuitionAccountReadingList(claims.UserID, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) infrastructureStatus(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	var instID *int64
	if claims.LoginType == "org" {
		value := claims.UserID
		_ = value
	}
	result, err := handler.service.StudentSyncStatus(instID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) mqEventLogs(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
		return
	}
	result, err := handler.service.PageMQEventLogs(parseInt(r.URL.Query().Get("current"), 1), parseInt(r.URL.Query().Get("size"), 20))
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) syncIntentStudents(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	_, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	instID, batchSize := syncParams(r)
	count, err := handler.service.SyncIntentStudentsToES(instID, batchSize)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"count": count, "message": "sync completed"}, ctx.RequestID)
}

func (handler *Handler) rebuildIntentStudents(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	_, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	instID, batchSize := syncParams(r)
	count, err := handler.service.RebuildIntentStudentIndex(instID, batchSize)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"count": count, "message": "rebuild completed"}, ctx.RequestID)
}

func (handler *Handler) clearIntentStudents(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	_, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if err := handler.service.ClearIntentStudentIndex(); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"message": "intent student index cleared"}, ctx.RequestID)
}

func (handler *Handler) requireAuth(w http.ResponseWriter, r *http.Request, ctx tenant.Context) (authx.Claims, bool) {
	token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	if token == "" {
		token = strings.TrimSpace(r.Header.Get("ybcToken"))
	}
	if token == "" {
		if cookie, err := r.Cookie("ybcToken"); err == nil {
			token = strings.TrimSpace(cookie.Value)
		}
	}
	if token == "" {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", ctx.RequestID)
		return authx.Claims{}, false
	}

	claims, err := handler.service.ParseToken(token)
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", ctx.RequestID)
		return authx.Claims{}, false
	}
	return claims, true
}

func syncParams(r *http.Request) (*int64, int) {
	var instID *int64
	if raw := strings.TrimSpace(r.URL.Query().Get("instId")); raw != "" {
		if value, err := strconv.ParseInt(raw, 10, 64); err == nil {
			instID = &value
		}
	}
	batchSize := 1000
	if raw := strings.TrimSpace(r.URL.Query().Get("batchSize")); raw != "" {
		if value, err := strconv.Atoi(raw); err == nil && value > 0 {
			batchSize = value
		}
	}
	return instID, batchSize
}

func parseInstUserQuery(raw map[string]any) model.InstUserQueryDTO {
	query := model.InstUserQueryDTO{}
	if page, ok := raw["pageRequestModel"].(map[string]any); ok {
		query.PageRequestModel.PageIndex = asInt(page["pageIndex"], 1)
		query.PageRequestModel.PageSize = asInt(page["pageSize"], 10)
	}
	if qm, ok := raw["queryModel"].(map[string]any); ok {
		query.QueryModel.ID = asInt64Ptr(qm["id"])
		query.QueryModel.UserType = asIntPtr(qm["userType"])
		query.QueryModel.RoleIDs = asInt64Slice(qm["roleIds"])
		query.QueryModel.Status = asBoolPtr(qm["status"])
		query.QueryModel.DeptID = asInt64Ptr(qm["deptId"])
		query.QueryModel.SearchKey = asString(qm["searchKey"])
		query.QueryModel.CreateTimeBegin = asDateStartPtr(qm["createTimeBegin"])
		query.QueryModel.CreateTimeEnd = asDateEndPtr(qm["createTimeEnd"])
	}
	return query
}
