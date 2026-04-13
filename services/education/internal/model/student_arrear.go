package model

import "time"

type StudentRegistrationArrearPagedQueryDTO struct {
	PageRequestModel PageRequestModel                    `json:"pageRequestModel"`
	QueryModel       StudentRegistrationArrearQueryModel `json:"queryModel"`
}

type StudentRegistrationArrearQueryModel struct {
	OrderNumber      string `json:"orderNumber"`
	LessonID         string `json:"lessonId"`
	StudentID        string `json:"studentId"`
	Keyword          string `json:"keyword"`
	KeywordType      string `json:"keywordType"`
	CreatedTimeBegin string `json:"createdTimeBegin"`
	CreatedTimeEnd   string `json:"createdTimeEnd"`
}

type StudentRegistrationArrearItem struct {
	OrderID      string     `json:"orderId"`
	OrderNumber  string     `json:"orderNumber"`
	StudentID    string     `json:"studentId"`
	StudentName  string     `json:"studentName"`
	Sex          *int       `json:"sex,omitempty"`
	Avatar       string     `json:"avatar"`
	Phone        string     `json:"phone"`
	ArrearAmount float64    `json:"arrearAmount"`
	OrderAmount  float64    `json:"orderAmount"`
	PaidAmount   float64    `json:"paidAmount"`
	ProductName  string     `json:"productName"`
	CreatedTime  *time.Time `json:"createdTime,omitempty"`
}

type StudentRegistrationArrearPagedResult struct {
	List  []StudentRegistrationArrearItem `json:"list"`
	Total int                             `json:"total"`
}

type StudentRegistrationArrearStatistics struct {
	TotalArrearAmount float64 `json:"totalArrearAmount"`
}

type StudentLessonArrearPagedQueryDTO struct {
	PageRequestModel PageRequestModel             `json:"pageRequestModel"`
	QueryModel       StudentLessonArrearQueryModel `json:"queryModel"`
}

type StudentLessonArrearQueryModel struct {
	LessonID    string `json:"lessonId"`
	StudentID   string `json:"studentId"`
	Keyword     string `json:"keyword"`
	KeywordType string `json:"keywordType"`
}

type StudentLessonArrearItem struct {
	StudentID          string  `json:"studentId"`
	StudentName        string  `json:"studentName"`
	Sex                *int    `json:"sex,omitempty"`
	Avatar             string  `json:"avatar"`
	Phone              string  `json:"phone"`
	LessonID           string  `json:"lessonId"`
	LessonName         string  `json:"lessonName"`
	TuitionAccountID   string  `json:"tuitionAccountId"`
	LessonChargingMode int     `json:"lessonChargingMode"`
	BeInArrearsTotal   float64 `json:"beInArrearsTotal"`
	RecordCount        int     `json:"recordCount"`
	AdvisorStaffID     string  `json:"advisorStaffId"`
	AdvisorStaffName   string  `json:"advisorStaffName"`
	StudentManagerID   string  `json:"studentManagerId"`
	StudentManagerName string  `json:"studentManagerName"`
}

type StudentLessonArrearPagedResult struct {
	List  []StudentLessonArrearItem `json:"list"`
	Total int                       `json:"total"`
}

type StudentLessonArrearStatistics struct {
	TotalArrearAmount float64 `json:"totalArrearAmount"`
	TotalArrearTime   float64 `json:"totalArrearTime"`
}
