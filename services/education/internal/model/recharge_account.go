package model

import "time"

type RechargeAccountItemPageQueryDTO struct {
	PageRequestModel PageRequestModel              `json:"pageRequestModel"`
	QueryModel       RechargeAccountItemQueryModel `json:"queryModel"`
	SortModel        RechargeAccountItemSortModel  `json:"sortModel"`
}

type RechargeAccountItemQueryModel struct {
	StudentID              string `json:"studentId"`
	ShowZeroBalanceAccount *bool  `json:"showZeroBalanceAccount"`
}

type RechargeAccountItemSortModel struct {
	OrderByUpdatedTime int `json:"orderByUpdatedTime"`
}

type RechargeAccountStudentItem struct {
	IsMainStudent bool   `json:"isMainStudent"`
	StudentID     string `json:"studentId"`
	StudentName   string `json:"studentName"`
}

type RechargeAccountItem struct {
	RechargeAccountID       string                       `json:"rechargeAccountId"`
	RechargeAccountName     string                       `json:"rechargeAccountName,omitempty"`
	Phone                   string                       `json:"phone"`
	MainStudentID           string                       `json:"mainStudentId"`
	UpdateTime              *time.Time                   `json:"updateTime,omitempty"`
	BalanceTotal            float64                      `json:"balanceTotal"`
	RechargeBalance         float64                      `json:"rechargeBalance,omitempty"`
	ResidualBalance         float64                      `json:"residualBalance,omitempty"`
	GivingBalance           float64                      `json:"givingBalance,omitempty"`
	RechargeAccountStudents []RechargeAccountStudentItem `json:"rechargeAccountStudents"`
}

type RechargeAccountItemPageResult struct {
	List  []RechargeAccountItem `json:"list"`
	Total int                   `json:"total"`
}

type RechargeAccountStatistics struct {
	RechargeAccountTotal float64 `json:"rechargeAccountTotal"`
	AmountTotal          float64 `json:"amountTotal"`
	GivingAmountTotal    float64 `json:"givingAmountTotal"`
	RechargeAmountTotal  float64 `json:"rechargeAmountTotal,omitempty"`
	ResidualAmountTotal  float64 `json:"residualAmountTotal,omitempty"`
}
