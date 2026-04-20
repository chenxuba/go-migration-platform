package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	LeaveStatusPending  = 1
	LeaveStatusApproved = 2
	LeaveStatusRejected = 3
	LeaveStatusRevoked  = 4
)

const (
	LeaveTypePersonal = 1
	LeaveTypeSick     = 2
	LeaveTypeSuspend  = 3
)

const (
	LeaveActionCreate      = 1
	LeaveActionApprove     = 2
	LeaveActionReject      = 3
	LeaveActionRevoke      = 4
	LeaveActionAutoApprove = 5
)

const LeaveApprovalConfigType = 7

type FlexibleString string

func (value *FlexibleString) UnmarshalJSON(data []byte) error {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" || trimmed == "null" {
		*value = ""
		return nil
	}
	if strings.HasPrefix(trimmed, `"`) {
		var text string
		if err := json.Unmarshal(data, &text); err != nil {
			return err
		}
		*value = FlexibleString(strings.TrimSpace(text))
		return nil
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	var number json.Number
	if err := decoder.Decode(&number); err == nil {
		*value = FlexibleString(number.String())
		return nil
	}

	return fmt.Errorf("expected string or number, got %s", trimmed)
}

func (value FlexibleString) String() string {
	return string(value)
}

type LeavePagedQueryDTO struct {
	PageRequestModel PageRequestModel      `json:"pageRequestModel"`
	QueryModel       LeavePagedQueryFilter `json:"queryModel"`
	SortModel        LeavePagedQuerySort   `json:"sortModel"`
}

type LeavePagedQueryFilter struct {
	StudentID      FlexibleString `json:"studentId"`
	ApplyStartTime string         `json:"applyStartTime"`
	ApplyEndTime   string         `json:"applyEndTime"`
	LeaveTypes     []int          `json:"leaveTypes"`
	Statuses       []int          `json:"statuses"`
}

type LeavePagedQuerySort struct {
	ByApplyTime int `json:"byApplyTime"`
}

type LeavePreviewDTO struct {
	StudentID FlexibleString `json:"studentId"`
	StartTime string         `json:"startTime"`
	EndTime   string         `json:"endTime"`
}

type LeaveCreateDTO struct {
	StudentID      FlexibleString `json:"studentId"`
	StartTime      string         `json:"startTime"`
	EndTime        string         `json:"endTime"`
	LeaveType      int            `json:"leaveType"`
	Reason         string         `json:"reason"`
	ProofMaterials []string       `json:"proofMaterials"`
	Remark         string         `json:"remark"`
}

type LeavePagedItem struct {
	ID                  string     `json:"id"`
	StudentID           string     `json:"studentId"`
	StudentName         string     `json:"studentName"`
	StudentAvatarURL    string     `json:"studentAvatarUrl,omitempty"`
	StudentPhone        string     `json:"studentPhone"`
	StartTime           *time.Time `json:"startTime,omitempty"`
	EndTime             *time.Time `json:"endTime,omitempty"`
	IsAgent             bool       `json:"isAgent"`
	LeaveType           int        `json:"leaveType"`
	LeaveTypeText       string     `json:"leaveTypeText"`
	InitiateStaffName   string     `json:"initiateStaffName"`
	OperatorName        string     `json:"operatorName"`
	Status              int        `json:"status"`
	StatusText          string     `json:"statusText"`
	CurrentApproverName string     `json:"currentApproverName"`
	ApproverName        string     `json:"approverName"`
	ApplyTime           *time.Time `json:"applyTime,omitempty"`
}

type LeavePagedResult struct {
	List  []LeavePagedItem `json:"list"`
	Total int              `json:"total"`
}

type LeaveScheduleSnapshotVO struct {
	ScheduleID         string     `json:"scheduleId"`
	ClassType          int        `json:"classType"`
	TeachingClassID    string     `json:"teachingClassId"`
	TeachingClassName  string     `json:"teachingClassName"`
	LessonID           string     `json:"lessonId"`
	LessonName         string     `json:"lessonName"`
	TeacherID          string     `json:"teacherId"`
	TeacherName        string     `json:"teacherName"`
	StartTime          *time.Time `json:"startTime,omitempty"`
	EndTime            *time.Time `json:"endTime,omitempty"`
	RosterStatusBefore int        `json:"rosterStatusBefore"`
}

type LeaveProcessVO struct {
	ActionType int        `json:"actionType"`
	Name       string     `json:"name"`
	Status     string     `json:"status"`
	ActionTime *time.Time `json:"actionTime,omitempty"`
	Pending    bool       `json:"pending"`
	Remark     string     `json:"remark,omitempty"`
}

type LeaveDetailVO struct {
	ID                  string                    `json:"id"`
	StudentID           string                    `json:"studentId"`
	StudentName         string                    `json:"studentName"`
	StudentAvatarURL    string                    `json:"studentAvatarUrl,omitempty"`
	StudentPhone        string                    `json:"studentPhone"`
	StartTime           *time.Time                `json:"startTime,omitempty"`
	EndTime             *time.Time                `json:"endTime,omitempty"`
	IsAgent             bool                      `json:"isAgent"`
	LeaveType           int                       `json:"leaveType"`
	LeaveTypeText       string                    `json:"leaveTypeText"`
	Reason              string                    `json:"reason"`
	ProofMaterials      []string                  `json:"proofMaterials"`
	Remark              string                    `json:"remark"`
	Status              int                       `json:"status"`
	StatusText          string                    `json:"statusText"`
	InitiateStaffName   string                    `json:"initiateStaffName"`
	OperatorName        string                    `json:"operatorName"`
	CurrentApproverName string                    `json:"currentApproverName"`
	ApproverName        string                    `json:"approverName"`
	ApplyTime           *time.Time                `json:"applyTime,omitempty"`
	Schedules           []LeaveScheduleSnapshotVO `json:"schedules"`
	Processes           []LeaveProcessVO          `json:"processes"`
}

type LeaveCreateResult struct {
	ID     string `json:"id"`
	Status int    `json:"status"`
}
