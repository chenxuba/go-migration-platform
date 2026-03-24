package model

import "time"

const (
	LedgerTypeIncome      = 1
	LedgerTypeExpenditure = 2
)

const (
	LedgerSourceSystem = 1
	LedgerSourceManual = 2
)

const (
	LedgerSystemTypeOrderPayment = 101
)

const (
	LedgerConfirmStatusPending      = 0
	LedgerConfirmStatusConfirmed    = 1
	LedgerConfirmStatusRefunding    = 2
	LedgerConfirmStatusRefundFailed = 3
)

const (
	LedgerCategoryOrderIncome      = "order-income"
	LedgerSubCategoryRegistration  = "registration-renewal"
	LedgerSubCategoryRefundCourse  = "refund-course"
	LedgerSubCategoryTransferOrder = "transfer-course"
)

type LedgerListQueryDTO struct {
	PageRequestModel PageRequestModel  `json:"pageRequestModel"`
	QueryModel       LedgerQueryFilter `json:"queryModel"`
	SortModel        map[string]any    `json:"sortModel"`
}

type LedgerQueryFilter struct {
	AccountIDs            []string `json:"accountIds"`
	LedgerConfirmStatuses []int    `json:"ledgerConfirmStatuses"`
	SourceTypes           []int    `json:"sourceTypes"`
	DealStaffID           string   `json:"dealStaffId"`
	ConfirmStaffID        string   `json:"confirmStaffId"`
	StudentID             string   `json:"studentId"`
	OrderNumber           string   `json:"orderNumber"`
	BankSlipNo            string   `json:"bankSlipNo"`
	LedgerNumber          string   `json:"ledgerNumber"`
	ConfirmStartTime      string   `json:"confirmStartTime"`
	ConfirmEndTime        string   `json:"confirmEndTime"`
	PayStartTime          string   `json:"payStartTime"`
	PayEndTime            string   `json:"payEndTime"`
	LedgerSubCategoryIDs  []string `json:"ledgerSubCategoryIds"`
	OrderID               string   `json:"orderId"`
}

type LedgerRichText struct {
	Text   string   `json:"text"`
	Images []string `json:"images"`
}

type LedgerListItemVO struct {
	ID                    string         `json:"id"`
	OrgID                 string         `json:"orgId,omitempty"`
	SchoolID              string         `json:"schoolId,omitempty"`
	Type                  int            `json:"type"`
	SourceType            int            `json:"sourceType"`
	LedgerCategoryID      string         `json:"ledgerCategoryId"`
	LedgerCategoryName    string         `json:"ledgerCategoryName"`
	LedgerSubCategoryID   string         `json:"ledgerSubCategoryId"`
	LedgerSubCategoryName string         `json:"ledgerSubCategoryName"`
	LedgerCategoryIcon    string         `json:"ledgerCategoryIcon"`
	Amount                float64        `json:"amount"`
	DealStaffID           string         `json:"dealStaffId,omitempty"`
	DealStaffName         string         `json:"dealStaffName,omitempty"`
	PayTime               *time.Time     `json:"payTime,omitempty"`
	CreatedTime           *time.Time     `json:"createdTime,omitempty"`
	PayMethod             *int           `json:"payMethod,omitempty"`
	AccountID             string         `json:"accountId,omitempty"`
	AccountName           string         `json:"accountName,omitempty"`
	ReciprocalAccount     string         `json:"reciprocalAccount"`
	BankSlipNo            string         `json:"bankSlipNo"`
	OrderNumber           string         `json:"orderNumber,omitempty"`
	LedgerNumber          string         `json:"ledgerNumber"`
	StudentID             string         `json:"studentId,omitempty"`
	StudentName           string         `json:"studentName,omitempty"`
	StudentPhone          string         `json:"studentPhone,omitempty"`
	IsConfirmed           bool           `json:"isConfirmed"`
	ConfirmRemark         LedgerRichText `json:"confirmRemark"`
	ProductItems          []string       `json:"productItems"`
	ConfirmStaffName      string         `json:"confirmStaffName"`
	ConfirmTime           *time.Time     `json:"confirmTime,omitempty"`
	SystemType            int            `json:"systemType"`
	OrderID               string         `json:"orderId,omitempty"`
	PaymentVoucher        LedgerRichText `json:"paymentVoucher"`
	BillFlowID            string         `json:"billFlowId,omitempty"`
	BillID                string         `json:"billId,omitempty"`
	LedgerConfirmStatus   int            `json:"ledgerConfirmStatus"`
	ErrorMessage          string         `json:"errorMessage"`
}

type LedgerListResultVO struct {
	List  []LedgerListItemVO `json:"list"`
	Total int                `json:"total"`
}

type LedgerStatisticsVO struct {
	IncomeAmount      float64 `json:"incomeAmount"`
	ExpenditureAmount float64 `json:"expenditureAmount"`
	BalanceAmount     float64 `json:"balanceAmount"`
	TotalConfirm      int     `json:"totalConfirm"`
	TotalUnConfirm    int     `json:"totalUnConfirm"`
	TotalRefunding    int     `json:"totalRefunding"`
	TotalRefundFailed int     `json:"totalRefundFailed"`
}

type LedgerOperateDTO struct {
	ID string `json:"id"`
}
