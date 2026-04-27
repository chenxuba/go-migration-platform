package model

const CampusDataClearScopeBusinessOnly = "business_only"

type CampusDataClearRequest struct {
	Scope string `json:"scope"`
}

type CampusDataClearSummary struct {
	Students                   int `json:"students"`
	StudentFieldValues         int `json:"studentFieldValues"`
	StudentChangeRecords       int `json:"studentChangeRecords"`
	FaceProfiles               int `json:"faceProfiles"`
	FaceAttendanceSessions     int `json:"faceAttendanceSessions"`
	FaceAttendanceRecords      int `json:"faceAttendanceRecords"`
	FaceRollCallTasks          int `json:"faceRollCallTasks"`
	StudentTeachingRecords     int `json:"studentTeachingRecords"`
	StudentTeachingChangeLogs  int `json:"studentTeachingChangeLogs"`
	StudentRehabRecords        int `json:"studentRehabRecords"`
	FollowRecords              int `json:"followRecords"`
	HomeworkTasks              int `json:"homeworkTasks"`
	NoticeRecords              int `json:"noticeRecords"`
	Orders                     int `json:"orders"`
	OrderCourseDetails         int `json:"orderCourseDetails"`
	OrderPaymentDetails        int `json:"orderPaymentDetails"`
	Ledgers                    int `json:"ledgers"`
	ApprovalRecords            int `json:"approvalRecords"`
	ApprovalHistories          int `json:"approvalHistories"`
	TuitionAccounts            int `json:"tuitionAccounts"`
	TuitionAccountFlows        int `json:"tuitionAccountFlows"`
	CloseTuitionAccountOrders  int `json:"closeTuitionAccountOrders"`
	RefundTuitionOrders        int `json:"refundTuitionOrders"`
	RefundTuitionOrderItems    int `json:"refundTuitionOrderItems"`
	SuspendResumeTuitionOrders int `json:"suspendResumeTuitionOrders"`
	RechargeAccounts           int `json:"rechargeAccounts"`
	RechargeAccountStudents    int `json:"rechargeAccountStudents"`
	RechargeAccountFlows       int `json:"rechargeAccountFlows"`
	RechargeAccountOrders      int `json:"rechargeAccountOrders"`
	RechargeAccountOrderTags   int `json:"rechargeAccountOrderTags"`
	RechargeAccountBills       int `json:"rechargeAccountBills"`
	RechargeAccountBillFlows   int `json:"rechargeAccountBillFlows"`
	Courses                    int `json:"courses"`
	CourseDetails              int `json:"courseDetails"`
	CourseQuotations           int `json:"courseQuotations"`
	CoursePropertyResults      int `json:"coursePropertyResults"`
	ProductPackages            int `json:"productPackages"`
	ProductPackageItems        int `json:"productPackageItems"`
	ProductPackageProperties   int `json:"productPackageProperties"`
	CourseSaleVolumesReset     int `json:"courseSaleVolumesReset"`
	ImportTasks                int `json:"importTasks"`
	ImportTaskRecords          int `json:"importTaskRecords"`
	OrderImportTasks           int `json:"orderImportTasks"`
	OrderImportTaskRecords     int `json:"orderImportTaskRecords"`
	RechargeImportTasks        int `json:"rechargeImportTasks"`
	RechargeImportTaskRecords  int `json:"rechargeImportTaskRecords"`
	ExportRecords              int `json:"exportRecords"`
	TemplateMessageRecords     int `json:"templateMessageRecords"`
	TemplateMessageRecordItems int `json:"templateMessageRecordItems"`
	WeChatBindTickets          int `json:"weChatBindTickets"`
	WeChatStudentBindings      int `json:"weChatStudentBindings"`
	TeachingClasses            int `json:"teachingClasses"`
	TeachingClassStudents      int `json:"teachingClassStudents"`
	TeachingClassTeachers      int `json:"teachingClassTeachers"`
	TeachingSchedules          int `json:"teachingSchedules"`
	TeachingScheduleStudents   int `json:"teachingScheduleStudents"`
	TeachingScheduleBatchMetas int `json:"teachingScheduleBatchMetas"`
	TeachingRecords            int `json:"teachingRecords"`
	TeachingClassOperationLogs int `json:"teachingClassOperationLogs"`
	TeachingClassEntryExits    int `json:"teachingClassEntryExits"`
}

type CampusDataClearResult struct {
	Scope                     string                 `json:"scope"`
	ScopeName                 string                 `json:"scopeName"`
	Cleared                   CampusDataClearSummary `json:"cleared"`
	Preserved                 []string               `json:"preserved"`
	IntentStudentIndexCleared bool                   `json:"intentStudentIndexCleared"`
	IntentStudentIndexMessage string                 `json:"intentStudentIndexMessage,omitempty"`
}
