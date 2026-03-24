package model

const CampusDataClearScopeBusinessOnly = "business_only"

type CampusDataClearRequest struct {
	Scope string `json:"scope"`
}

type CampusDataClearSummary struct {
	Students               int `json:"students"`
	StudentFieldValues     int `json:"studentFieldValues"`
	StudentChangeRecords   int `json:"studentChangeRecords"`
	FollowRecords          int `json:"followRecords"`
	Orders                 int `json:"orders"`
	OrderCourseDetails     int `json:"orderCourseDetails"`
	OrderPaymentDetails    int `json:"orderPaymentDetails"`
	ApprovalRecords        int `json:"approvalRecords"`
	ApprovalHistories      int `json:"approvalHistories"`
	TuitionAccounts        int `json:"tuitionAccounts"`
	CourseSaleVolumesReset int `json:"courseSaleVolumesReset"`
}

type CampusDataClearResult struct {
	Scope                     string                 `json:"scope"`
	ScopeName                 string                 `json:"scopeName"`
	Cleared                   CampusDataClearSummary `json:"cleared"`
	Preserved                 []string               `json:"preserved"`
	IntentStudentIndexCleared bool                   `json:"intentStudentIndexCleared"`
	IntentStudentIndexMessage string                 `json:"intentStudentIndexMessage,omitempty"`
}
