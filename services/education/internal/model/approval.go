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
