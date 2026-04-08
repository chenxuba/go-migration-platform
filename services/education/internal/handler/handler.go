package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	mux.HandleFunc("/api/v1/students/detail", handler.studentDetailView)
	mux.HandleFunc("/api/v1/orders", handler.orders)
	mux.HandleFunc("/api/v1/inst-config", handler.getInstConfig)
	mux.HandleFunc("/api/v1/inst-config/update", handler.setInstConfig)
	mux.HandleFunc("/api/v1/inst-config/period-effective-preview", handler.previewInstPeriodEffective)
	mux.HandleFunc("/api/v1/inst-config/period-repair", handler.repairInstPeriodConfigVersions)
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
	mux.HandleFunc("/api/v1/approvals/refuse", handler.rejectApprovalRecord)
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
	mux.HandleFunc("/api/v1/tuition-accounts/suspend-resume-orders/create", handler.addSuspendResumeTuitionAccountOrder)
	mux.HandleFunc("/api/v1/tuition-accounts/sub-account-date-info", handler.tuitionAccountSubAccountDateInfo)
	mux.HandleFunc("/api/v1/tuition-accounts/sub-account-priority-configs/list", handler.subTuitionAccountPriorityConfigList)
	mux.HandleFunc("/api/v1/tuition-accounts/revert-close-preview", handler.revertCloseTuitionAccountPreview)
	mux.HandleFunc("/api/v1/tuition-accounts/revert-close", handler.revertCloseTuitionAccount)
	mux.HandleFunc("/api/v1/tuition-accounts/close-orders/list", handler.closeTuitionAccountOrderList)
	mux.HandleFunc("/api/v1/tuition-account-flows/list", handler.tuitionAccountFlowRecordList)
	mux.HandleFunc("/api/v1/tuition-account-flows/sub-list", handler.subTuitionAccountFlowRecordList)
	mux.HandleFunc("/api/v1/lesson-incomes/query-paged-list", handler.lessonIncomeQueryPagedList)
	mux.HandleFunc("/api/v1/lesson-incomes/statistics", handler.lessonIncomeStatistics)
	mux.HandleFunc("/api/v1/course-properties", handler.coursePropertyList)
	mux.HandleFunc("/api/v1/course-properties/update", handler.updateCourseProperty)
	mux.HandleFunc("/api/v1/course-properties/init", handler.initInstCourseProperty)
	mux.HandleFunc("/api/v1/course-property-options", handler.coursePropertyOptions)
	mux.HandleFunc("/api/v1/course-property-options/create", handler.addCoursePropertyOption)
	mux.HandleFunc("/api/v1/course-property-options/update", handler.updateCoursePropertyOption)
	mux.HandleFunc("/api/v1/course-property-options/delete", handler.deleteCoursePropertyOption)
	mux.HandleFunc("/api/v1/course-property-options/sort", handler.updateCoursePropertyOptionSort)
	mux.HandleFunc("/api/v1/classrooms", handler.classrooms)
	mux.HandleFunc("/api/v1/classrooms/create", handler.createClassroom)
	mux.HandleFunc("/api/v1/classrooms/update", handler.updateClassroom)
	mux.HandleFunc("/api/v1/classrooms/status", handler.updateClassroomStatus)
	mux.HandleFunc("/api/v1/classrooms/delete", handler.deleteClassroom)
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
	mux.HandleFunc("/api/v1/product-packages/page", handler.productPackagePagedList)
	mux.HandleFunc("/api/v1/product-packages/statistics", handler.productPackageStatistics)
	mux.HandleFunc("/api/v1/product-packages/create", handler.createProductPackage)
	mux.HandleFunc("/api/v1/product-packages/sale-status", handler.updateProductPackageSaleStatus)
	mux.HandleFunc("/api/v1/product-packages/micro-school-rules", handler.updateProductPackageMicroSchoolRules)
	mux.HandleFunc("/api/v1/product-packages/enroll-edit", handler.updateProductPackageAllowEditWhenEnroll)
	mux.HandleFunc("/api/v1/product-packages/delete", handler.deleteProductPackage)
	mux.HandleFunc("/api/v1/courses/process-content", handler.processContentPage)
	mux.HandleFunc("/api/v1/courses/delete-restore", handler.batchDeleteOrRestoreCourses)
	mux.HandleFunc("/api/v1/courses/sale-status", handler.batchSaleStatus)
	mux.HandleFunc("/api/v1/courses/micro-school-show", handler.batchOpenMicroSchoolShow)
	mux.HandleFunc("/api/v1/compose-lessons/create", handler.createComposeLesson)
	mux.HandleFunc("/api/v1/compose-lessons/page", handler.pageComposeLessonsForPC)
	mux.HandleFunc("/api/v1/group-classes/check-name", handler.checkClassName)
	mux.HandleFunc("/api/v1/group-classes/create", handler.createGroupClass)
	mux.HandleFunc("/api/v1/group-classes/update", handler.updateGroupClass)
	mux.HandleFunc("/api/v1/group-classes/page", handler.pageGroupClasses)
	mux.HandleFunc("/api/v1/group-classes/detail", handler.getGroupClassDetail)
	mux.HandleFunc("/api/v1/group-classes/statistics", handler.groupClassStatistics)
	mux.HandleFunc("/api/v1/group-classes/students-by-class-ids", handler.listGroupClassStudentsByClassIDs)
	mux.HandleFunc("/api/v1/group-classes/batch-assign-students", handler.batchAssignGroupClassStudents)
	mux.HandleFunc("/api/v1/teaching-schedules/smart/export", handler.smartTeachingSchedulesExport)
	mux.HandleFunc("/api/v1/teaching-schedules/time/export", handler.timeTeachingSchedulesExport)
	mux.HandleFunc("/api/v1/teaching-schedules/by-teacher-matrix/export", handler.teachingSchedulesTeacherMatrixExport)
	mux.HandleFunc("/api/v1/teaching-schedules/by-teacher-matrix", handler.teachingSchedulesByTeacherMatrix)
	mux.HandleFunc("/api/v1/teaching-schedules/clear-all", handler.clearAllTeachingSchedules)
	mux.HandleFunc("/api/v1/teaching-schedules", handler.teachingSchedules)
	mux.HandleFunc("/api/v1/teaching-schedules/one-to-one/slot-availability", handler.checkOneToOneScheduleAvailability)
	mux.HandleFunc("/api/v1/teaching-schedules/one-to-one/assistant-availability", handler.checkAssistantScheduleAvailability)
	mux.HandleFunc("/api/v1/teaching-schedules/one-to-one/validate", handler.validateOneToOneSchedules)
	mux.HandleFunc("/api/v1/teaching-schedules/one-to-one/create", handler.createOneToOneSchedules)
	mux.HandleFunc("/api/v1/teaching-schedules/batch-detail", handler.teachingScheduleBatchDetail)
	mux.HandleFunc("/api/v1/teaching-schedules/batch-replace", handler.replaceTeachingScheduleBatch)
	mux.HandleFunc("/api/v1/teaching-schedules/conflict-detail", handler.teachingScheduleConflictDetail)
	mux.HandleFunc("/api/v1/teaching-schedules/cancel", handler.cancelTeachingSchedules)
	mux.HandleFunc("/api/v1/teaching-schedules/copy-week", handler.copyTeachingSchedulesWeek)
	mux.HandleFunc("/api/v1/teaching-schedules/batch-update", handler.batchUpdateTeachingSchedules)
	mux.HandleFunc("/api/v1/infrastructure/status", handler.infrastructureStatus)
	mux.HandleFunc("/api/v1/mq/event-logs", handler.mqEventLogs)
	mux.HandleFunc("/api/v1/es-sync/intent-student/sync", handler.syncIntentStudents)
	mux.HandleFunc("/api/v1/es-sync/intent-student/rebuild", handler.rebuildIntentStudents)
	mux.HandleFunc("/api/v1/es-sync/intent-student/clear", handler.clearIntentStudents)
	mux.HandleFunc("/api/v1/campus-data/clear", handler.clearCampusData)
	mux.HandleFunc("/api/v1/intent-students/page", handler.intentStudentsPage)
	mux.HandleFunc("/api/v1/intent-students/detail", handler.intentStudentDetail)
	mux.HandleFunc("/api/v1/current-students/page", handler.currentStudentsPage)
	mux.HandleFunc("/api/v1/enrolled-students/page", handler.enrolledStudentsPage)
	mux.HandleFunc("/api/v1/enrolled-students/export", handler.exportEnrolledStudents)
	mux.HandleFunc("/api/v1/enrolled-students/export-records", handler.listEnrolledStudentExportRecords)
	mux.HandleFunc("/api/v1/enrolled-students/export-records/download", handler.downloadEnrolledStudentExportRecord)
	mux.HandleFunc("/api/v1/students/overview-statistics", handler.studentOverviewStatistics)
	mux.HandleFunc("/api/v1/orders/list", handler.orderList)
	mux.HandleFunc("/api/v1/orders/detail-paged", handler.orderDetailPaged)
	mux.HandleFunc("/api/v1/orders/detail", handler.orderDetail)
	mux.HandleFunc("/api/v1/orders/receipt-pdf/download", handler.downloadOrderReceiptPDF)
	mux.HandleFunc("/api/v1/orders/close", handler.closeOrder)
	mux.HandleFunc("/api/v1/orders/import-template/lesson-hour", handler.buildLessonHourOrderImportTemplate)
	mux.HandleFunc("/api/v1/orders/import-template/lesson-hour/file", handler.downloadLessonHourOrderImportTemplate)
	mux.HandleFunc("/api/v1/orders/import-template/time-slot", handler.buildTimeSlotOrderImportTemplate)
	mux.HandleFunc("/api/v1/orders/import-template/time-slot/file", handler.downloadTimeSlotOrderImportTemplate)
	mux.HandleFunc("/api/v1/orders/import-template/amount", handler.buildAmountOrderImportTemplate)
	mux.HandleFunc("/api/v1/orders/import-template/amount/file", handler.downloadAmountOrderImportTemplate)
	mux.HandleFunc("/api/v1/orders/import-upload", handler.uploadOrderImportFile)
	mux.HandleFunc("/api/v1/orders/import-uploaded-file", handler.downloadUploadedOrderImportFile)
	mux.HandleFunc("/api/v1/orders/import-tasks/submit", handler.submitOrderImportTask)
	mux.HandleFunc("/api/v1/orders/import-tasks/list", handler.listOrderImportTasks)
	mux.HandleFunc("/api/v1/orders/import-tasks/detail", handler.getOrderImportTaskDetail)
	mux.HandleFunc("/api/v1/orders/import-tasks/records", handler.getOrderImportTaskRecordList)
	mux.HandleFunc("/api/v1/orders/import-tasks/batch-save-records", handler.batchSaveOrderImportTaskRecords)
	mux.HandleFunc("/api/v1/orders/import-tasks/start", handler.startOrderImportTask)
	mux.HandleFunc("/api/v1/orders/import-tasks/clear", handler.clearOrderImportTasks)
	mux.HandleFunc("/api/v1/orders/import-tasks/delete", handler.deleteOrderImportTask)
	mux.HandleFunc("/api/v1/ledgers/list", handler.ledgerList)
	mux.HandleFunc("/api/v1/ledgers/statistics", handler.ledgerStatistics)
	mux.HandleFunc("/api/v1/ledgers/confirm", handler.confirmLedger)
	mux.HandleFunc("/api/v1/ledgers/cancel-confirm", handler.cancelConfirmLedger)
	mux.HandleFunc("/api/v1/recharge-accounts/page", handler.rechargeAccountItemPage)
	mux.HandleFunc("/api/v1/recharge-accounts/statistics", handler.rechargeAccountStatistics)
	mux.HandleFunc("/api/v1/recharge-accounts/update", handler.updateRechargeAccount)
	mux.HandleFunc("/api/v1/recharge-accounts/import-template/by-student", handler.buildRechargeAccountImportByStudentTemplate)
	mux.HandleFunc("/api/v1/recharge-accounts/import-template/by-student/file", handler.downloadRechargeAccountImportByStudentTemplate)
	mux.HandleFunc("/api/v1/recharge-accounts/import-template/by-account", handler.buildRechargeAccountImportByAccountTemplate)
	mux.HandleFunc("/api/v1/recharge-accounts/import-template/by-account/file", handler.downloadRechargeAccountImportByAccountTemplate)
	mux.HandleFunc("/api/v1/recharge-accounts/import-upload", handler.uploadRechargeAccountImportFile)
	mux.HandleFunc("/api/v1/recharge-accounts/import-uploaded-file", handler.downloadUploadedRechargeAccountImportFile)
	mux.HandleFunc("/api/v1/recharge-accounts/import-tasks/submit", handler.submitRechargeAccountImportTask)
	mux.HandleFunc("/api/v1/recharge-accounts/import-tasks/list", handler.listRechargeAccountImportTasks)
	mux.HandleFunc("/api/v1/recharge-accounts/import-tasks/detail", handler.getRechargeAccountImportTaskDetail)
	mux.HandleFunc("/api/v1/recharge-accounts/import-tasks/records", handler.getRechargeAccountImportTaskRecordList)
	mux.HandleFunc("/api/v1/recharge-accounts/import-tasks/batch-save-records", handler.batchSaveRechargeAccountImportTaskRecords)
	mux.HandleFunc("/api/v1/recharge-accounts/import-tasks/start", handler.startRechargeAccountImportTask)
	mux.HandleFunc("/api/v1/recharge-accounts/import-tasks/clear", handler.clearRechargeAccountImportTasks)
	mux.HandleFunc("/api/v1/recharge-accounts/import-tasks/delete", handler.deleteRechargeAccountImportTask)
	mux.HandleFunc("/api/v1/recharge-accounts/by-student", handler.rechargeAccountByStudent)
	mux.HandleFunc("/api/v1/recharge-accounts/get", handler.getRechargeAccount)
	mux.HandleFunc("/api/v1/recharge-accounts/details/page", handler.rechargeAccountDetailPage)
	mux.HandleFunc("/api/v1/recharge-accounts/expend-income", handler.rechargeAccountExpendIncome)
	mux.HandleFunc("/api/v1/recharge-account-orders/create", handler.createRechargeAccountOrder)
	mux.HandleFunc("/api/v1/recharge-account-orders/create-refund", handler.createRechargeAccountRefundOrder)
	mux.HandleFunc("/api/v1/recharge-account-orders/detail", handler.rechargeAccountOrderDetail)
	mux.HandleFunc("/api/v1/orders/pay-by-schoolpal", handler.payOrderBySchoolPal)
	mux.HandleFunc("/api/v1/order-tags/list-paged", handler.orderTagListPaged)
	mux.HandleFunc("/api/v1/order-tags/create", handler.createOrderTag)
	mux.HandleFunc("/api/v1/order-tags/update", handler.updateOrderTag)
	mux.HandleFunc("/api/v1/orders/check-quote", handler.checkQuoteInfo)
	mux.HandleFunc("/api/v1/orders/calc-enroll-type", handler.calcCourseEnrollType)
	mux.HandleFunc("/api/v1/orders/create", handler.createOrder)
	mux.HandleFunc("/api/v1/orders/pay", handler.payOrder)
	mux.HandleFunc("/api/v1/orders/registration-list", handler.registrationListPage)
	mux.HandleFunc("/api/v1/one-to-ones/page", handler.oneToOnePage)
	mux.HandleFunc("/api/v1/one-to-ones/detail", handler.oneToOneDetail)
	mux.HandleFunc("/api/v1/one-to-ones/lessons-by-student", handler.listOneToOneLessonsByStudent)
	mux.HandleFunc("/api/v1/tuition-accounts/by-student-and-lesson", handler.listTuitionAccountsByStudentAndLesson)
	mux.HandleFunc("/api/v1/tuition-accounts/page-by-lesson-id", handler.pageTuitionAccountsByLessonID)
	mux.HandleFunc("/api/v1/tuition-accounts/for-one-to-one-deduction", handler.listOneToOneDeductionTuitionAccounts)
	mux.HandleFunc("/api/v1/one-to-ones/check-name", handler.checkOneToOneName)
	mux.HandleFunc("/api/v1/one-to-ones/exist", handler.existOneToOne)
	mux.HandleFunc("/api/v1/one-to-ones/create", handler.createOneToOne)
	mux.HandleFunc("/api/v1/one-to-ones/update", handler.updateOneToOne)
	mux.HandleFunc("/api/v1/one-to-ones/switch-default-tuition-account", handler.switchOneToOneDefaultTuitionAccount)
	mux.HandleFunc("/api/v1/one-to-ones/close", handler.closeOneToOne)
	mux.HandleFunc("/api/v1/tuition-accounts/close-order", handler.addCloseTuitionAccountOrder)
	mux.HandleFunc("/api/v1/one-to-ones/reopen", handler.reopenOneToOne)
	mux.HandleFunc("/api/v1/one-to-ones/batch-assign-class-teacher", handler.batchAssignOneToOneClassTeacher)
	mux.HandleFunc("/api/v1/one-to-ones/batch-update-class-time", handler.batchUpdateOneToOneClassTime)
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
	mux.HandleFunc("/api/v1/intent-students/import-template", handler.buildIntentionStudentImportTemplate)
	mux.HandleFunc("/api/v1/intent-students/import-template/file", handler.downloadIntentionStudentImportTemplate)
	mux.HandleFunc("/api/v1/intent-students/import-parse", handler.parseIntentionStudentImportFile)
	mux.HandleFunc("/api/v1/intent-students/import-upload", handler.uploadIntentionStudentImportFile)
	mux.HandleFunc("/api/v1/intent-students/import-uploaded-file", handler.downloadUploadedIntentionStudentImportFile)
	mux.HandleFunc("/api/v1/intent-students/import-tasks/submit", handler.submitIntentionStudentImportTask)
	mux.HandleFunc("/api/v1/intent-students/import-tasks/list", handler.listIntentionStudentImportTasks)
	mux.HandleFunc("/api/v1/intent-students/import-tasks/detail", handler.getIntentionStudentImportTaskDetail)
	mux.HandleFunc("/api/v1/intent-students/import-tasks/records", handler.getIntentionStudentImportTaskRecordList)
	mux.HandleFunc("/api/v1/intent-students/import-tasks/batch-save-records", handler.batchSaveIntentionStudentImportTaskRecords)
	mux.HandleFunc("/api/v1/intent-students/import-tasks/start", handler.startIntentionStudentImportTask)
	mux.HandleFunc("/api/v1/intent-students/import-tasks/clear", handler.clearIntentionStudentImportTasks)
	mux.HandleFunc("/api/v1/intent-students/import-tasks/delete", handler.deleteIntentionStudentImportTask)
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
	var effectiveDate *time.Time
	if raw := strings.TrimSpace(r.URL.Query().Get("effectiveDate")); raw != "" {
		t, parseErr := time.ParseInLocation("2006-01-02", raw, time.Local)
		if parseErr != nil {
			httpx.WriteError(w, http.StatusBadRequest, "effectiveDate 格式应为 YYYY-MM-DD", ctx.RequestID)
			return
		}
		effectiveDate = &t
	}
	result, err := handler.service.GetInstConfig(claims.UserID, effectiveDate)
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
	result, err := handler.service.SetInstConfig(claims.UserID, payload)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) previewInstPeriodEffective(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.PreviewInstPeriodConfigUpdate(claims.UserID, payload["unifiedTimePeriodJson"])
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) repairInstPeriodConfigVersions(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	result, err := handler.service.RepairInstPeriodConfigVersions(claims.UserID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
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
