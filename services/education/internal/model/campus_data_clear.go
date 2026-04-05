package model

const CampusDataClearScopeBusinessOnly = "business_only"

type CampusDataClearRequest struct {
	Scope string `json:"scope"`
}

type CampusDataClearSummary struct {
	Students                  int `json:"students"`
	StudentFieldValues        int `json:"studentFieldValues"`
	StudentChangeRecords      int `json:"studentChangeRecords"`
	FollowRecords             int `json:"followRecords"`
	Orders                    int `json:"orders"`
	OrderCourseDetails        int `json:"orderCourseDetails"`
	OrderPaymentDetails       int `json:"orderPaymentDetails"`
	Ledgers                   int `json:"ledgers"`
	ApprovalRecords           int `json:"approvalRecords"`
	ApprovalHistories         int `json:"approvalHistories"`
	TuitionAccounts           int `json:"tuitionAccounts"`
	TuitionAccountFlows       int `json:"tuitionAccountFlows"`
	RechargeAccounts          int `json:"rechargeAccounts"`
	RechargeAccountStudents   int `json:"rechargeAccountStudents"`
	RechargeAccountFlows      int `json:"rechargeAccountFlows"`
	Courses                   int `json:"courses"`
	CourseDetails             int `json:"courseDetails"`
	CourseQuotations          int `json:"courseQuotations"`
	CoursePropertyResults     int `json:"coursePropertyResults"`
	ProductPackages           int `json:"productPackages"`
	ProductPackageItems       int `json:"productPackageItems"`
	ProductPackageProperties  int `json:"productPackageProperties"`
	CourseSaleVolumesReset    int `json:"courseSaleVolumesReset"`
	ImportTasks               int `json:"importTasks"`
	ImportTaskRecords         int `json:"importTaskRecords"`
	OrderImportTasks          int `json:"orderImportTasks"`
	OrderImportTaskRecords    int `json:"orderImportTaskRecords"`
	RechargeImportTasks       int `json:"rechargeImportTasks"`
	RechargeImportTaskRecords int `json:"rechargeImportTaskRecords"`
	ExportRecords             int `json:"exportRecords"`
	TeachingClasses           int `json:"teachingClasses"`
	TeachingClassStudents     int `json:"teachingClassStudents"`
	TeachingClassTeachers     int `json:"teachingClassTeachers"`
	TeachingSchedules         int `json:"teachingSchedules"`
}

type CampusDataClearResult struct {
	Scope                     string                 `json:"scope"`
	ScopeName                 string                 `json:"scopeName"`
	Cleared                   CampusDataClearSummary `json:"cleared"`
	Preserved                 []string               `json:"preserved"`
	IntentStudentIndexCleared bool                   `json:"intentStudentIndexCleared"`
	IntentStudentIndexMessage string                 `json:"intentStudentIndexMessage,omitempty"`
}
