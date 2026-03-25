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

type UpdateRechargeAccountDTO struct {
	RechargeAccountID   string `json:"rechargeAccountId"`
	RechargeAccountName string `json:"rechargeAccountName"`
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

type StudentDetailView struct {
	ID                     string     `json:"id"`
	Name                   string     `json:"name"`
	Phone                  string     `json:"phone"`
	Avatar                 string     `json:"avatar"`
	Sex                    int        `json:"sex"`
	PhoneRelationship      int        `json:"phoneRelationship"`
	SalespersonID          string     `json:"salespersonId"`
	SalespersonName        string     `json:"salespersonName"`
	CreatedTime            *time.Time `json:"createdTime,omitempty"`
	FirstEnrolledTime      *time.Time `json:"firstEnrolledTime,omitempty"`
	TurnedHistoryTime      *time.Time `json:"turnedHistoryTime,omitempty"`
	CreatedStaffID         string     `json:"createdStaffId"`
	CreatedStaffName       string     `json:"createdStaffName"`
	CollectorStaffID       string     `json:"collectorStaffId"`
	CollectorStaffName     string     `json:"collectorStaffName"`
	PhoneSellStaffID       string     `json:"phoneSellStaffId"`
	PhoneSellStaffName     string     `json:"phoneSellStaffName"`
	ForegroundStaffID      string     `json:"foregroundStaffId"`
	ForegroundStaffName    string     `json:"foregroundStaffName"`
	ViceSellStaffStaffID   string     `json:"viceSellStaffStaffId"`
	ViceSellStaffStaffName string     `json:"viceSellStaffStaffName"`
	Status                 int        `json:"status"`
}

type RechargeAccountByStudentStudent struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Avatar        string `json:"avatar"`
	Sex           int    `json:"sex"`
	Phone         string `json:"phone"`
	IsMainStudent bool   `json:"isMainStudent"`
}

type RechargeAccountByStudent struct {
	ID              string                            `json:"id"`
	AccountName     string                            `json:"accountName"`
	Phone           string                            `json:"phone"`
	MainStudentID   string                            `json:"mainStudentId"`
	Balance         float64                           `json:"balance"`
	GivingBalance   float64                           `json:"givingBalance"`
	ResidualBalance float64                           `json:"residualBalance"`
	CreatedAt       *time.Time                        `json:"createdAt,omitempty"`
	Students        []RechargeAccountByStudentStudent `json:"students"`
}

type CreateRechargeAccountOrderDTO struct {
	RechargeAccountID    string   `json:"rechargeAccountId"`
	Amount               float64  `json:"amount"`
	GivingAmount         float64  `json:"givingAmount"`
	ResidualAmount       float64  `json:"residualAmount"`
	DealDate             string   `json:"dealDate"`
	SalePersonID         string   `json:"salePersonId"`
	CollectorStaffID     string   `json:"collectorStaffId"`
	PhoneSellStaffID     string   `json:"phoneSellStaffId"`
	ForegroundStaffID    string   `json:"foregroundStaffId"`
	ViceSellStaffStaffID string   `json:"viceSellStaffStaffId"`
	Remark               string   `json:"remark"`
	OrderTagIDs          []string `json:"orderTagIds"`
	ExternalRemark       string   `json:"externalRemark"`
	StudentID            string   `json:"studentId"`
}

type RechargeAccountOrderCreateResult struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type RechargeAccountOrderTag struct {
	TagID   string `json:"tagId"`
	TagName string `json:"tagName"`
}

type RechargeAccountOrderBillDetail struct {
	ID        string `json:"id"`
	Status    int    `json:"status"`
	BillFlows []any  `json:"billFlows"`
}

type RechargeAccountOrderDetail struct {
	ID                string                         `json:"id"`
	RechargeAccountID string                         `json:"rechargeAccountId"`
	SaleOrderID       string                         `json:"saleOrderId"`
	OrderNumber       string                         `json:"orderNumber"`
	Status            int                            `json:"status"`
	Amount            float64                        `json:"amount"`
	GivingAmount      float64                        `json:"givingAmount"`
	ResidualAmount    float64                        `json:"residualAmount"`
	OperatorName      string                         `json:"operatorName"`
	CreatedAt         *time.Time                     `json:"createdAt,omitempty"`
	Bill              RechargeAccountOrderBillDetail `json:"bill"`
	ApproveID         *string                        `json:"approveId,omitempty"`
	OrderTags         []RechargeAccountOrderTag      `json:"orderTags"`
	StudentID         string                         `json:"studentId"`
	StudentName       string                         `json:"studentName"`
	StudentPhone      string                         `json:"studentPhone"`
	OrderObsolete     any                            `json:"orderObsolete"`
}

type RechargeAccountOrderDetailQuery struct {
	RechargeAccountOrderID string `json:"rechargeAccountOrderId"`
	SaleOrderID            string `json:"saleOrderId"`
}

type PayOrderBySchoolPalDTO struct {
	BillID         string  `json:"billId"`
	Amount         float64 `json:"amount"`
	Remark         string  `json:"remark"`
	PayMethod      *int    `json:"payMethod,omitempty"`
	AmountID       *int64  `json:"amountId,omitempty"`
	PayTime        string  `json:"payTime,omitempty"`
	PaymentVoucher string  `json:"paymentVoucher,omitempty"`
}
