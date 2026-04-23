package model

import "time"

type PendingRenewalStudentPagedQueryDTO struct {
	PageRequestModel PageRequestModel              `json:"pageRequestModel"`
	QueryModel       PendingRenewalStudentQueryDTO `json:"queryModel"`
	SortModel        PendingRenewalStudentSortDTO  `json:"sortModel"`
}

type PendingRenewalStudentQueryDTO struct {
	StudentID      string   `json:"studentId"`
	ProductID      string   `json:"productId"`
	ProductIDs     []string `json:"productIds"`
	ClassTeacherID string   `json:"classTeacherId"`
	ClassIDs       []string `json:"classIds"`
	StatusList     []int    `json:"statusList"`
}

type PendingRenewalStudentSortDTO struct {
	ExpriedTime int `json:"expriedTime"`
}

type PendingRenewalStudentItem struct {
	TuitionAccountID   string                    `json:"tuitionAccountId"`
	StudentID          string                    `json:"studentId"`
	Sex                *int                      `json:"sex,omitempty"`
	Avatar             string                    `json:"avatar"`
	LessonID           string                    `json:"lessonId"`
	LessonName         string                    `json:"lessonName"`
	StudentName        string                    `json:"studentName"`
	LeftQuantity       float64                   `json:"leftQuantity"`
	LeftFreeQuantity   float64                   `json:"leftFreeQuantity"`
	EnableExpireTime   bool                      `json:"enableExpireTime"`
	ExpireTime         *time.Time                `json:"expireTime,omitempty"`
	LessonChargingMode *int                      `json:"lessonChargingMode,omitempty"`
	TotalQuantity      float64                   `json:"totalQuantity"`
	LatestStartTime    *time.Time                `json:"latestStartTime,omitempty"`
	Phone              string                    `json:"phone"`
	Tuition            float64                   `json:"tuition"`
	ClassTeacherList   []RegistrationListTeacher `json:"classTeacherList,omitempty"`
	Status             *int                      `json:"status,omitempty"`
	AdvisorStaffID     *int64                    `json:"advisorStaffId,omitempty"`
	AdvisorStaffName   string                    `json:"advisorStaffName"`
	StudentManagerID   *int64                    `json:"studentManagerId,omitempty"`
	StudentManagerName string                    `json:"studentManagerName"`
}

type PendingRenewalStudentPagedResult struct {
	List         []PendingRenewalStudentItem `json:"list"`
	Total        int                         `json:"total"`
	StudentCount int                         `json:"studentCount"`
}
