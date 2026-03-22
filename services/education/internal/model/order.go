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

type QuoteDetailDTO struct {
	HandleType     *int       `json:"handleType"`
	CourseID       int64      `json:"courseId"`
	CourseType     *int       `json:"courseType"`
	QuoteID        int64      `json:"quoteId"`
	LessonMode     *int       `json:"lessonMode"`
	ClassID        *int64     `json:"classId"`
	Count          *int       `json:"count"`
	Unit           *int       `json:"unit"`
	FreeQuantity   float64    `json:"freeQuantity"`
	DiscountType   *int       `json:"discountType"`
	DiscountNumber float64    `json:"discountNumber"`
	HasValidDate   *bool      `json:"hasValidDate"`
	ValidDate      *time.Time `json:"validDate"`
	EndDate        *time.Time `json:"endDate"`
	ShareDiscount  string     `json:"shareDiscount"`
	Amount         string     `json:"amount"`
	Quantity       float64    `json:"quantity"`
	RealQuantity   float64    `json:"realQuantity"`
	RealAmount     string     `json:"realAmount"`
}

type OrderDetailDTO struct {
	QuoteDetailList     []QuoteDetailDTO `json:"quoteDetailList"`
	OrderDiscountType   *int             `json:"orderDiscountType"`
	OrderDiscountNumber float64          `json:"orderDiscountNumber"`
	OrderDiscountAmount string           `json:"orderDiscountAmount"`
	OrderRealQuantity   float64          `json:"orderRealQuantity"`
	OrderRealAmount     string           `json:"orderRealAmount"`
	InternalRemark      string           `json:"internalRemark"`
	ExternalRemark      string           `json:"externalRemark"`
	DealDate            *time.Time       `json:"dealDate"`
	SalePerson          *int64           `json:"salePerson"`
	OrderTagIDs         []int64          `json:"orderTagIds"`
}

type CreateOrderDTO struct {
	StudentID   int64          `json:"studentId"`
	OrderDetail OrderDetailDTO `json:"orderDetail"`
}

type PayAccountDTO struct {
	OrderID        int64      `json:"orderId"`
	AmountID       *int64     `json:"amountId"`
	PayMethod      *int       `json:"payMethod"`
	PayAmount      float64    `json:"payAmount"`
	PayTime        *time.Time `json:"payTime"`
	PaymentVoucher string     `json:"paymentVoucher"`
}

type PayOrderDTO struct {
	OrderID     int64           `json:"orderId"`
	PayAmount   float64         `json:"payAmount"`
	PayAccounts []PayAccountDTO `json:"payAccounts"`
}

type RegistrationListQueryDTO struct {
	PageRequestModel PageRequestModel        `json:"pageRequestModel"`
	QueryModel       RegistrationListFilters `json:"queryModel"`
	SortModel        SortModel               `json:"sortModel"`
}

type RegistrationListFilters struct {
	FromExpireTime             string   `json:"fromExpireTime"`
	ToExpireTime               string   `json:"toExpireTime"`
	FromSuspendedTime          string   `json:"fromSuspendedTime"`
	ToSuspendedTime            string   `json:"toSuspendedTime"`
	FromClosedTime             string   `json:"fromClosedTime"`
	ToClosedTime               string   `json:"toClosedTime"`
	IsSetExpireTime            *bool    `json:"isSetExpireTime"`
	AssignedClass              *bool    `json:"assignedClass"`
	StudentID                  string   `json:"studentId"`
	LessonType                 *int     `json:"lessonType"`
	RemainLessonChargingMode   *int     `json:"remainLessonChargingMode"`
	FromRemainQuantity         *int     `json:"fromRemainQuantity"`
	ToRemainQuantity           *int     `json:"toRemainQuantity"`
	LessonChargingList         []int    `json:"lessonChargingList"`
	StatusList                 []int    `json:"statusList"`
	ClassTeacherID             string   `json:"classTeacherId"`
	SalespersonID              string   `json:"salespersonId"`
	ClassIDs                   []string `json:"classIds"`
	ProductIDs                 []string `json:"productIds"`
	IsArrears                  *bool    `json:"isArrears"`
	LastestTeachingRecordStart string   `json:"lastestTeachingRecordStartTime"`
	LastestTeachingRecordEnd   string   `json:"lastestTeachingRecordEndTime"`
}

type RegistrationListTeacher struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
}

type RegistrationListItem struct {
	TuitionAccountID          string                    `json:"tuitionAccountId"`
	StudentID                 string                    `json:"studentId"`
	StudentName               string                    `json:"studentName"`
	Avatar                    string                    `json:"avatar"`
	Sex                       *int                      `json:"sex,omitempty"`
	Phone                     string                    `json:"phone"`
	LessonID                  string                    `json:"lessonId"`
	LessonName                string                    `json:"lessonName"`
	LessonType                *int                      `json:"lessonType,omitempty"`
	LessonChargingMode        *int                      `json:"lessonChargingMode,omitempty"`
	Type                      *int                      `json:"type,omitempty"`
	TotalQuantity             float64                   `json:"totalQuantity"`
	TotalFreeQuantity         float64                   `json:"totalFreeQuantity"`
	TotalTuition              float64                   `json:"totalTuition"`
	Quantity                  float64                   `json:"quantity"`
	FreeQuantity              float64                   `json:"freeQuantity"`
	Tuition                   float64                   `json:"tuition"`
	ConfirmedTuition          float64                   `json:"confirmedTuition"`
	TuitionAccountStatus      *int                      `json:"tuitionAccountStatus,omitempty"`
	AssignedClass             bool                      `json:"assignedClass"`
	EnableExpireTime          bool                      `json:"enableExpireTime"`
	ExpireTime                *time.Time                `json:"expireTime,omitempty"`
	PlanSuspendTime           *time.Time                `json:"planSuspendTime,omitempty"`
	PlanResumeTime            *time.Time                `json:"planResumeTime,omitempty"`
	ChangeStatusTime          *time.Time                `json:"changeStatusTime,omitempty"`
	CanTransferTuitionAccount bool                      `json:"canTransferTuitionAccount"`
	AdvisorStaffID            *int64                    `json:"advisorStaffId,omitempty"`
	AdvisorStaffName          string                    `json:"advisorStaffName"`
	StudentManagerID          *int64                    `json:"studentManagerId,omitempty"`
	StudentManagerName        string                    `json:"studentManagerName"`
	ClassTeacherList          []RegistrationListTeacher `json:"classTeacherList,omitempty"`
	CreateTime                *time.Time                `json:"createTime,omitempty"`
	SuspendedTime             *time.Time                `json:"suspendedTime,omitempty"`
	ClassEndingTime           *time.Time                `json:"classEndingTime,omitempty"`
	PaidTuition               float64                   `json:"paidTuition"`
	ShouldTuition             float64                   `json:"shouldTuition"`
	ArrearTuition             float64                   `json:"arrearTuition"`
	ChargeAgainstTuition      float64                   `json:"chargeAgainstTuition"`
	TransferredTuition        float64                   `json:"transferredTuition"`
	PaidRemaining             float64                   `json:"paidRemaining"`
	HasGradeUpgrade           bool                      `json:"hasGradeUpgrade"`
	LessonScope               *int                      `json:"lessonScope,omitempty"`
	LastestTeachingRecordTime *time.Time                `json:"lastestTeachingRecordTime,omitempty"`
	ValidDate                 *time.Time                `json:"validDate,omitempty"`
	EndDate                   *time.Time                `json:"endDate,omitempty"`
}

type RegistrationListResultVO struct {
	TotalRemainedTuition     float64                `json:"totalRemainedTuition"`
	TotalConfirmedTuition    float64                `json:"totalConfirmedTuition"`
	TotalPaidRemainedTuition float64                `json:"totalPaidRemainedTuition"`
	Total                    int                    `json:"total"`
	StudentCount             int                    `json:"studentCount"`
	StudentTutionAccounts    []RegistrationListItem `json:"studentTutionAccounts"`
}
