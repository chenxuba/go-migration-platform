package model

type ParentLeaveQueryDTO struct {
	StudentID string `json:"studentId"`
	PageIndex int    `json:"pageIndex"`
	PageSize  int    `json:"pageSize"`
}

type ParentLeaveSummaryVO struct {
	Students  []ParentBoundStudentVO `json:"students"`
	Items     []ParentLeaveVO        `json:"items"`
	PageIndex int                    `json:"pageIndex"`
	PageSize  int                    `json:"pageSize"`
	Total     int                    `json:"total"`
	HasMore   bool                   `json:"hasMore"`
}

type ParentLeaveVO struct {
	ID               string `json:"id"`
	StudentID        string `json:"studentId"`
	StudentName      string `json:"studentName"`
	StudentAvatarURL string `json:"studentAvatarUrl,omitempty"`
	LeaveType        int    `json:"leaveType"`
	LeaveTypeText    string `json:"leaveTypeText"`
	Status           int    `json:"status"`
	StatusText       string `json:"statusText"`
	StartTime        string `json:"startTime"`
	EndTime          string `json:"endTime"`
	ApplyTime        string `json:"applyTime"`
}
