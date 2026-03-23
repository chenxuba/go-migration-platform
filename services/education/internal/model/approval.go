package model

import "time"

type ApprovalConfigPageQueryDTO struct {
	PageRequestModel PageRequestModel               `json:"pageRequestModel"`
	QueryModel       ApprovalConfigPageQueryFilters `json:"queryModel"`
	SortModel        ApprovalConfigPageQuerySort    `json:"sortModel"`
}

type ApprovalConfigPageQueryFilters struct {
	ApprovalNumber       string `json:"approvalNumber"`
	ApplicantID          *int64 `json:"applicantId"`
	OrderNumber          string `json:"orderNumber"`
	CurrentApproverID    *int64 `json:"currentApproverId"`
	FinishStartTime      string `json:"finishStartTime"`
	FinishEndTime        string `json:"finishEndTime"`
	ApplicationStartTime string `json:"applicationStartTime"`
	ApplicationEndTime   string `json:"applicationEndTime"`
	Statuses             []int  `json:"statuses"`
	StudentID            *int64 `json:"studentId"`
	QuickFilter          int    `json:"quickFilter"`
}

type ApprovalConfigPageQuerySort struct {
	ByInitiateTime int `json:"byInitiateTime"`
	ByFinishTime   int `json:"byFinishTime"`
}

type ApprovalConfigPageResult struct {
	Records []ApprovalConfigRecord `json:"records"`
	Total   int                    `json:"total"`
	Current int                    `json:"current"`
	Size    int                    `json:"size"`
}

type ApprovalConfigRecord struct {
	ID              int64                 `json:"id"`
	ApprovalNumber  string                `json:"approvalNumber"`
	ApprovalType    int                   `json:"approvalType"`
	CurrentApprover string                `json:"currentApprover"`
	ConfigVersion   int                   `json:"configVersion"`
	CurrentStep     *int                  `json:"currentStep,omitempty"`
	ApplicantName   string                `json:"applicantName"`
	StudentName     string                `json:"studentName"`
	StudentID       string                `json:"studentId"`
	StudentAvatar   string                `json:"studentAvatar"`
	Mobile          string                `json:"mobile"`
	ApprovalTime    *time.Time            `json:"approvalTime,omitempty"`
	FinishTime      *time.Time            `json:"finishTime,omitempty"`
	ApprovalStatus  *int                  `json:"approvalStatus,omitempty"`
	OrderNumber     string                `json:"orderNumber"`
	OrderID         string                `json:"orderId"`
	OrderType       *int                  `json:"orderType,omitempty"`
	ApproveFlows    []ApprovalFlowStageVO `json:"approveFlows,omitempty"`
}

type ApprovalDetailVO struct {
	ID                int64                 `json:"id"`
	ApprovalNumber    string                `json:"approvalNumber"`
	ApprovalType      int                   `json:"approvalType"`
	InitiateStaffName string                `json:"initiateStaffName"`
	InitiateTime      *time.Time            `json:"initiateTime,omitempty"`
	FinishTime        *time.Time            `json:"finishTime,omitempty"`
	Status            *int                  `json:"status,omitempty"`
	InitiateReason    string                `json:"initiateReason"`
	ApproveFlows      []ApprovalFlowStageVO `json:"approveFlows,omitempty"`
}

type ApprovalMyStatisticsVO struct {
	TruntoMyApproveCount int `json:"truntoMyApproveCount"`
	MyHaveApprovedCount  int `json:"myHaveApprovedCount"`
	InitiateApproveCount int `json:"initiateApproveCount"`
}

type ApprovalFlowStageVO struct {
	IsCurrentStage bool                  `json:"isCurrentStage"`
	Status         *int                  `json:"status,omitempty"`
	Remark         string                `json:"remark,omitempty"`
	OperateTime    *time.Time            `json:"operateTime,omitempty"`
	Step           int                   `json:"step"`
	FlowStaffs     []ApprovalFlowStaffVO `json:"flowStaffs"`
}

type ApprovalFlowStaffVO struct {
	StaffID          string `json:"staffId"`
	StaffName        string `json:"staffName"`
	TeacherStatus    int    `json:"teacherStatus"`
	IsApproveOperate bool   `json:"isApproveOperate"`
}

type ApprovalConfigSaveDTO struct {
	ID            int64                     `json:"id"`
	Enable        *bool                     `json:"enable"`
	RuleJSON      string                    `json:"ruleJson"`
	StaffFlowList []ApprovalConfigStaffFlow `json:"staffFlowList"`
}

type ApprovalConfigStaffFlow struct {
	Step     int     `json:"step"`
	StaffIDs []int64 `json:"staffIds"`
}

type ApprovalOperateDTO struct {
	ID     int64  `json:"id"`
	Remark string `json:"remark"`
}

type ApprovalTemplateFlowVO struct {
	Step       int      `json:"step"`
	StaffIDs   []string `json:"staffIds,omitempty"`
	StaffNames []string `json:"staffNames,omitempty"`
}

type ApprovalTemplateVO struct {
	ID               string                   `json:"id"`
	Enable           bool                     `json:"enable"`
	Type             int                      `json:"type"`
	Name             string                   `json:"name"`
	UpdatedStaffName string                   `json:"updatedStaffName,omitempty"`
	UpdatedTime      *time.Time               `json:"updatedTime,omitempty"`
	RuleJSON         string                   `json:"ruleJson"`
	FlowModels       []ApprovalTemplateFlowVO `json:"flowModels,omitempty"`
}

type ApprovalTemplateSaveRequest struct {
	ApproveTemplateRequests []ApprovalTemplateSaveItem `json:"approveTemplateRequests"`
}

type ApprovalTemplateSaveItem struct {
	ID                int64                          `json:"id"`
	Type              int                            `json:"type"`
	Enable            bool                           `json:"enable"`
	RuleJSON          string                         `json:"ruleJson"`
	FlowRequestModels []ApprovalTemplateFlowSaveItem `json:"flowRequestModels"`
}

type ApprovalTemplateFlowSaveItem struct {
	Step     int     `json:"step"`
	StaffIDs []int64 `json:"staffIds"`
}

type StaffSummaryQueryDTO struct {
	PageRequestModel PageRequestModel      `json:"pageRequestModel"`
	QueryModel       StaffSummaryQueryVOIn `json:"queryModel"`
}

type StaffSummaryQueryVOIn struct {
	SchoolID  string `json:"schoolId"`
	SearchKey string `json:"searchKey"`
}

type StaffSummaryVO struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Phone        string     `json:"phone"`
	SuperAdmin   bool       `json:"superAdmin"`
	Avatar       string     `json:"avatar"`
	Color        string     `json:"color"`
	Status       int        `json:"status"`
	CreatedAt    *time.Time `json:"createdAt,omitempty"`
	EmployeeType int        `json:"employeeType"`
}

type StaffSummaryPageVO struct {
	List  []StaffSummaryVO `json:"list"`
	Total int              `json:"total"`
}
