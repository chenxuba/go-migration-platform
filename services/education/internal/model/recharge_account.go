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
	ResidualAmountTotal  float64 `json:"residualAmountTotal"`
}

const (
	RechargeAccountFlowTypeRecharge           = 1
	RechargeAccountFlowTypeRefund             = 2
	RechargeAccountFlowTypeOrderExpend        = 3
	RechargeAccountFlowTypeRefundOrderReturn  = 4
	RechargeAccountFlowTypeVoidRecharge       = 5
	RechargeAccountFlowTypeTransferExpend     = 6
	RechargeAccountFlowTypeTransferReturn     = 7
	RechargeAccountFlowTypeVoidTransferExpend = 8
	RechargeAccountFlowTypeVoidTransferReturn = 9
	RechargeAccountFlowTypeVenueExpend        = 10
	RechargeAccountFlowTypeVenueReturn        = 11
	RechargeAccountFlowTypeVoidRefund         = 12
)

type RechargeAccountDetailQueryDTO struct {
	PageRequestModel PageRequestModel               `json:"pageRequestModel"`
	QueryModel       RechargeAccountDetailQuery     `json:"queryModel"`
	SortModel        RechargeAccountDetailSortModel `json:"sortModel"`
}

type RechargeAccountDetailQuery struct {
	StudentID string `json:"studentId"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	FlowTypes []int  `json:"flowTypes"`
}

type RechargeAccountDetailSortModel struct {
	OrderByCreatedTime int `json:"orderByCreatedTime"`
}

type RechargeAccountDetailItem struct {
	Phone                         string                       `json:"phone"`
	Amount                        float64                      `json:"amount"`
	GivingAmount                  float64                      `json:"givingAmount"`
	ResidualAmount                float64                      `json:"residualAmount"`
	RechargeAccountID             string                       `json:"rechargeAccountId"`
	RechargeAccountFlowID         string                       `json:"rechargeAccountFlowId"`
	RechargeAccountName           string                       `json:"rechargeAccountName"`
	Remark                        string                       `json:"remark"`
	CreateTime                    string                       `json:"createTime"`
	RechargeAccountFlowSourceType int                          `json:"rechargeAccountFlowSourceType"`
	DealDate                      string                       `json:"dealDate"`
	SourceID                      string                       `json:"sourceId"`
	SourceOrderNumber             string                       `json:"sourceOrderNumber"`
	SourceOrderType               int                          `json:"sourceOrderType"`
	RechargeAccountStudents       []RechargeAccountStudentItem `json:"rechargeAccountStudents"`
	StudentID                     string                       `json:"studentId"`
	StudentName                   string                       `json:"studentName"`
	StudentPhone                  string                       `json:"studentPhone"`
	StudentAvatar                 string                       `json:"studentAvatar"`
	TotalAmount                   float64                      `json:"totalAmount"`
}

type RechargeAccountDetailPageResult struct {
	List  []RechargeAccountDetailItem `json:"list"`
	Total int                         `json:"total"`
}

type RechargeAccountExpendIncome struct {
	Expend float64 `json:"expend"`
	Income float64 `json:"income"`
}
