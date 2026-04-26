package model

import "time"

const (
	CloseTuitionAccountOrderStatusClosed  = 1
	CloseTuitionAccountOrderStatusRevoked = 4
)

type TuitionAccountSubAccountDateInfoQueryDTO struct {
	TuitionAccountID string `json:"tuitionAccountId"`
}

type TuitionAccountSubAccountDateInfoItem struct {
	ID                     string     `json:"id"`
	Quantity               float64    `json:"quantity"`
	CreatedTime            *time.Time `json:"createdTime,omitempty"`
	StartTime              *time.Time `json:"startTime,omitempty"`
	ActivedAt              *time.Time `json:"activedAt,omitempty"`
	RemainDays             float64    `json:"remainDays"`
	RawStatus              int        `json:"rawStatus"`
	Status                 int        `json:"status"`
	IsFree                 bool       `json:"isFree"`
	TotalDays              float64    `json:"totalDays"`
	Tuition                float64    `json:"tuition"`
	TotalTuition           float64    `json:"totalTuition"`
	EndDate                *time.Time `json:"endDate,omitempty"`
	SourceType             int        `json:"sourceType"`
	AccountSourceType      int        `json:"accountSourceType"`
	OrderID                string     `json:"orderId"`
	SourceID               string     `json:"sourceId"`
	UnitPrice              float64    `json:"unitPrice"`
	PaidTuition            float64    `json:"paidTuition"`
	ShouldTuition          float64    `json:"shouldTuition"`
	ArrearTuition          float64    `json:"arrearTuition"`
	ChargeAgainstTuition   float64    `json:"chargeAgainstTuition"`
	TransferredTuition     float64    `json:"transferredTuition"`
	PaidRemaining          float64    `json:"paidRemaining"`
	UsedTuition            float64    `json:"usedTuition"`
	StartDate              *time.Time `json:"startDate,omitempty"`
	ExpiredToClearQuantity bool       `json:"expiredToClearQuantity"`
	ExpireDate             *time.Time `json:"expireDate,omitempty"`
}

type TuitionAccountSubAccountDateInfoResult struct {
	List []TuitionAccountSubAccountDateInfoItem `json:"list"`
}

type RefundTuitionAccountOwedSummaryQueryDTO struct {
	TuitionAccountID string `json:"tuitionAccountId"`
}

type RefundTuitionAccountOwedSummaryResult struct {
	ArrearAmountTotal  float64 `json:"arrearAmountTotal"`
	BadDebtAmountTotal float64 `json:"badDebtAmountTotal"`
	OrderID            string  `json:"orderId"`
	OrderType          int     `json:"orderType"`
}

type RefundTuitionAccountHandlingFeeQueryDTO struct {
	TuitionAccountID string  `json:"tuitionAccountId"`
	RefundQuantity   float64 `json:"refundQuantity"`
}

type RefundTuitionAccountValuableEstimateQueryDTO struct {
	TuitionAccountID string  `json:"tuitionAccountId"`
	Quantity         float64 `json:"quantity"`
}

type RefundTuitionAccountValuableEstimateSubAccount struct {
	TuitionAccountID string  `json:"tuitionAccountId"`
	OrderNumber      string  `json:"orderNumber"`
	Quantity         float64 `json:"quantity"`
	FreeQuantity     float64 `json:"freeQuantity"`
	Tuition          float64 `json:"tuition"`
}

type RefundTuitionAccountValuableEstimateResult struct {
	TuitionAccountID string                                           `json:"tuitionAccountId"`
	Quantity         float64                                          `json:"quantity"`
	FreeQuantity     float64                                          `json:"freeQuantity"`
	Tuition          float64                                          `json:"tuition"`
	SubAccounts      []RefundTuitionAccountValuableEstimateSubAccount `json:"subAccounts"`
}

type RefundTuitionAccountHandlingFeeDetail struct {
	OrderNumber          string  `json:"orderNumber"`
	OriginalUnitPrice    float64 `json:"originalUnitPrice"`
	DiscountedUnitPrice  float64 `json:"discountedUnitPrice"`
	ShouldTuition        float64 `json:"shouldTuition"`
	PaidAmount           float64 `json:"paidAmount"`
	TransferredTuition   float64 `json:"transferredTuition"`
	ConsumedQuantity     float64 `json:"consumedQuantity"`
	ArrearTuition        float64 `json:"arrearTuition"`
	OriginalRefundAmount float64 `json:"originalRefundAmount"`
	LessonChargingMode   int     `json:"lessonChargingMode"`
	DeductionQuantity    float64 `json:"deductionQuantity"`
}

type RefundTuitionAccountHandlingFeeResult struct {
	TuitionAccountID          string                                  `json:"tuitionAccountId"`
	RefundAmount              float64                                 `json:"refundAmount"`
	TotalOriginalRefundAmount float64                                 `json:"totalOriginalRefundAmount"`
	TotalArrearDeduction      float64                                 `json:"totalArrearDeduction"`
	HandlingFee               float64                                 `json:"handlingFee"`
	LessonChargingMode        int                                     `json:"lessonChargingMode"`
	PaidRefundQuantity        float64                                 `json:"paidRefundQuantity"`
	GiftRefundQuantity        float64                                 `json:"giftRefundQuantity"`
	ArrearAmountTotal         float64                                 `json:"arrearAmountTotal"`
	BadDebtAmountTotal        float64                                 `json:"badDebtAmountTotal"`
	OrderID                   string                                  `json:"orderId"`
	OrderType                 int                                     `json:"orderType"`
	Details                   []RefundTuitionAccountHandlingFeeDetail `json:"details"`
}

type SubTuitionAccountPriorityConfigItem struct {
	PriorityType  int  `json:"priorityType"`
	SortDirection int  `json:"sortDirection"`
	SortWeight    int  `json:"sortWeight"`
	IsEnabled     bool `json:"isEnabled"`
}

type SubTuitionAccountPriorityConfigResult struct {
	List []SubTuitionAccountPriorityConfigItem `json:"list"`
}

type RevertCloseTuitionAccountPreviewQueryDTO struct {
	TuitionAccountID string `json:"tuitionAccountId"`
}

type RevertCloseTuitionAccountSubPeriod struct {
	Quantity  float64    `json:"quantity"`
	IsFree    bool       `json:"isFree"`
	StartDate *time.Time `json:"startDate,omitempty"`
	EndDate   *time.Time `json:"endDate,omitempty"`
}

type RevertCloseTuitionAccountPreview struct {
	TuitionAccountID           string                               `json:"tuitionAccountId"`
	LessonName                 string                               `json:"lessonName,omitempty"`
	LessonType                 int                                  `json:"lessonType"`
	LessonChargingMode         int                                  `json:"lessonChargingMode"`
	CloseTuitionAccountOrderID string                               `json:"closeTuitionAccountOrderId"`
	CloseTime                  *time.Time                           `json:"closeTime,omitempty"`
	Quantity                   float64                              `json:"quantity"`
	FreeQuantity               float64                              `json:"freeQuantity"`
	Tuition                    float64                              `json:"tuition"`
	Remark                     string                               `json:"remark"`
	ExpireDate                 *time.Time                           `json:"expireDate,omitempty"`
	ArrearAmountTotal          float64                              `json:"arrearAmountTotal"`
	BadDebtAmountTotal         float64                              `json:"badDebtAmountTotal"`
	OrderID                    string                               `json:"orderId"`
	OrderType                  int                                  `json:"orderType"`
	SubTuitionAccounts         []RevertCloseTuitionAccountSubPeriod `json:"subTuitionAccounts"`
}

type RevertCloseTuitionAccountDTO struct {
	TuitionAccountID           string `json:"tuitionAccountId"`
	CloseTuitionAccountOrderID string `json:"closeTuitionAccountOrderId"`
	StartDate                  string `json:"startDate"`
	ExpireDate                 string `json:"expireDate"`
	CurrentValidStartDate      string `json:"currentValidStartDate"`
}

type RevertCloseTuitionAccountResult struct {
	ID string `json:"id"`
}

type CloseTuitionAccountOrderRecordQueryDTO struct {
	TuitionAccountID string `json:"tuitionAccountId"`
}

type CloseTuitionAccountOrderRecordItem struct {
	ID               string     `json:"id"`
	TuitionAccountID string     `json:"tuitionAccountId"`
	Quantity         float64    `json:"quantity"`
	FreeQuantity     float64    `json:"freeQuantity"`
	Status           int        `json:"status"`
	UpdatedStaffID   string     `json:"updatedStaffId"`
	UpdatedStaffName string     `json:"updatedStaffName"`
	UpdatedTime      *time.Time `json:"updatedTime,omitempty"`
	CreatedTime      *time.Time `json:"createdTime,omitempty"`
}

type CloseTuitionAccountOrderRecordResult struct {
	List []CloseTuitionAccountOrderRecordItem `json:"list"`
}
