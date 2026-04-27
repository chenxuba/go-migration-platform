package model

import "time"

type SuspendResumeTuitionAccountOrderDTO struct {
	TuitionAccountID string `json:"tuitionAccountId"`
	Type             int    `json:"type"`
	ExpireTime       string `json:"expireTime"`
	ExpireType       int    `json:"expireType"`
	Remark           string `json:"remark"`
	SuspendDate      string `json:"suspendDate"`
	ResumeDate       string `json:"resumeDate"`
}

type SuspendResumeTuitionAccountOrderResult struct {
	ID        string `json:"id"`
	StudentID string `json:"studentId"`
	LessonID  string `json:"lessonId"`
}

type SuspendResumeTuitionAccountOrderListQueryDTO struct {
	TuitionAccountID string `json:"tuitionAccountId"`
}

type SuspendResumeTuitionAccountOrderListItem struct {
	ID               string     `json:"id"`
	TuitionAccountID string     `json:"tuitionAccountId"`
	Type             int        `json:"type"`
	ExpireTime       *time.Time `json:"expireTime,omitempty"`
	ExpireType       int        `json:"expireType"`
	Remark           string     `json:"remark"`
	SuspendDate      *time.Time `json:"suspendDate,omitempty"`
	ResumeDate       *time.Time `json:"resumeDate,omitempty"`
	CreatedStaffID   string     `json:"createdStaffId"`
	CreatedStaffName string     `json:"createdStaffName"`
	CreatedTime      *time.Time `json:"createdTime,omitempty"`
}

type SuspendResumeTuitionAccountOrderListResult struct {
	List []SuspendResumeTuitionAccountOrderListItem `json:"list"`
}
