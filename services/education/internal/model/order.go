package model

import "time"

type OrderManageQueryDTO struct {
	PageRequestModel PageRequestModel  `json:"pageRequestModel"`
	QueryModel       OrderQueryFilters `json:"queryModel"`
}

type OrderQueryFilters struct {
	Keyword     string `json:"keyword"`
	OrderStatus *int   `json:"orderStatus"`
	StudentID   string `json:"studentId"`
	StaffID     string `json:"staffId"`
}

type OrderManageQueryVO struct {
	OrderID        string     `json:"orderId"`
	SourceID       string     `json:"sourceId"`
	OrderNumber    string     `json:"orderNumber"`
	StudentID      string     `json:"studentId,omitempty"`
	StudentName    string     `json:"studentName,omitempty"`
	StudentPhone   string     `json:"studentPhone,omitempty"`
	CreatedTime    time.Time  `json:"createdTime"`
	Amount         float64    `json:"amount"`
	PaidAmount     float64    `json:"paidAmount"`
	OrderStatus    *int       `json:"orderStatus,omitempty"`
	OrderType      *int       `json:"orderType,omitempty"`
	OrderSource    *int       `json:"orderSource,omitempty"`
	StaffID        string     `json:"staffId,omitempty"`
	StaffName      string     `json:"staffName,omitempty"`
	DealDate       *time.Time `json:"dealDate,omitempty"`
	SalePersonID   string     `json:"salePersonId,omitempty"`
	SalePersonName string     `json:"salePersonName,omitempty"`
	ProductItems   []string   `json:"productItems,omitempty"`
	ArrearAmount   float64    `json:"arrearAmount"`
	IsAmountOwed   bool       `json:"isAmountOwed"`
	Remark         string     `json:"remark,omitempty"`
	ExternalRemark string     `json:"externalRemark,omitempty"`
}

type OrderManageResultVO struct {
	List  []OrderManageQueryVO `json:"list"`
	Total int                  `json:"total"`
}

type BadDebtDTO struct {
	OrderID string `json:"orderId"`
	Remark  string `json:"remark"`
}

type CourseEnrollTypeDTO struct {
	StudentID int64                       `json:"studentId"`
	Courses   []CourseEnrollTypeCheckItem `json:"courses"`
}

type CourseEnrollTypeCheckItem struct {
	CourseID   int64 `json:"courseId"`
	IsAudition bool  `json:"isAudition"`
}

type CourseEnrollTypeVO struct {
	CourseID   int64 `json:"courseId"`
	EnrollType int   `json:"enrollType"`
}

type CheckQuoteDTO struct {
	QuoteDetailList []CheckQuoteDetailDTO `json:"quoteDetailList"`
}

type CheckQuoteDetailDTO struct {
	CourseID    int64   `json:"courseId"`
	QuoteID     int64   `json:"quoteId"`
	Price       float64 `json:"price"`
	Quantity    *int    `json:"quantity"`
	LessonModel *int    `json:"lessonModel"`
}

type CheckQuoteVO struct {
	CourseID int64 `json:"courseId"`
	Error    int   `json:"error"`
}
