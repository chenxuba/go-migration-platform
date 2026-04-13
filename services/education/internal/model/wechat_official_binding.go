package model

import "time"

type WeChatOfficialBindTicketPreviewDTO struct {
	BindTicket string `json:"bindTicket"`
}

type WeChatOfficialBindStudentLookupDTO struct {
	BindTicket string `json:"bindTicket"`
	Phone      string `json:"phone"`
}

type WeChatOfficialConfirmStudentBindingDTO struct {
	BindTicket string `json:"bindTicket"`
	StudentID  int64  `json:"studentId"`
	Phone      string `json:"phone"`
	MiniOpenID string `json:"miniOpenId"`
	UnionID    string `json:"unionId"`
}

type WeChatOfficialBindTicketPreviewVO struct {
	BindTicket      string     `json:"bindTicket"`
	Status          string     `json:"status"`
	InstitutionID   int64      `json:"institutionId"`
	InstitutionName string     `json:"institutionName"`
	SceneValue      string     `json:"sceneValue"`
	SceneStudentID  int64      `json:"sceneStudentId"`
	ExpiresAt       *time.Time `json:"expiresAt,omitempty"`
	UsedAt          *time.Time `json:"usedAt,omitempty"`
	HasBoundStudent bool       `json:"hasBoundStudent"`
}

type WeChatOfficialBindStudentCandidateVO struct {
	ID            int64  `json:"id"`
	StuName       string `json:"stuName"`
	AvatarURL     string `json:"avatarUrl"`
	Mobile        string `json:"mobile"`
	StudentStatus int    `json:"studentStatus"`
	IsBound       bool   `json:"isBound"`
}

type WeChatOfficialConfirmStudentBindingVO struct {
	StudentID       int64  `json:"studentId"`
	StudentName     string `json:"studentName"`
	InstitutionID   int64  `json:"institutionId"`
	InstitutionName string `json:"institutionName"`
	Subscribed      bool   `json:"subscribed"`
}
