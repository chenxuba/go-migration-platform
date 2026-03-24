package model

import "time"

const (
	TuitionAccountFlowSourceRegistration              = 1
	TuitionAccountFlowSourceTransferIn                = 2
	TuitionAccountFlowSourceCrossCampusTransferIn     = 3
	TuitionAccountFlowSourceCrossCampusAttendIn       = 4
	TuitionAccountFlowSourceConsumeReturn             = 5
	TuitionAccountFlowSourceRevokeGraduate            = 6
	TuitionAccountFlowSourceExpireRollback            = 7
	TuitionAccountFlowSourceRevokeRefundOrder         = 8
	TuitionAccountFlowSourceRevokeTransferOut         = 9
	TuitionAccountFlowSourceRevokeImportConsume       = 10
	TuitionAccountFlowSourceRevokeAutoConsume         = 11
	TuitionAccountFlowSourceConsume                   = 12
	TuitionAccountFlowSourceImportConsume             = 13
	TuitionAccountFlowSourceConsumeSupplement         = 14
	TuitionAccountFlowSourceAutoConsume               = 15
	TuitionAccountFlowSourceConsumeArrearsSettlement  = 16
	TuitionAccountFlowSourceTransferOut               = 17
	TuitionAccountFlowSourceCrossCampusTransferOut    = 18
	TuitionAccountFlowSourceCrossCampusAttendOut      = 19
	TuitionAccountFlowSourceGraduate                  = 20
	TuitionAccountFlowSourceExpireGraduate            = 21
	TuitionAccountFlowSourceRefund                    = 22
	TuitionAccountFlowSourceOrderVoid                 = 23
	TuitionAccountFlowSourceVoidCrossCampusTransferIn = 24
)

type TuitionAccountFlowRecordListQueryDTO struct {
	PageRequestModel PageRequestModel                   `json:"pageRequestModel"`
	QueryModel       TuitionAccountFlowRecordQueryModel `json:"queryModel"`
	SortModel        TuitionAccountFlowRecordSortModel  `json:"sortModel"`
}

type TuitionAccountFlowRecordQueryModel struct {
	ProductID   string `json:"productId"`
	StudentID   string `json:"studentId"`
	SourceTypes []int  `json:"sourceTypes"`
	StartTime   string `json:"startTime"`
	EndTime     string `json:"endTime"`
}

type TuitionAccountFlowRecordSortModel struct {
	OrderByCreatedTime int `json:"orderByCreatedTime"`
}

type TuitionAccountFlowRecordItem struct {
	TuitionAccountFlowID string     `json:"tutionAccountFlowId"`
	TuitionAccountID     string     `json:"tuitionAccountId"`
	StudentID            string     `json:"studentId"`
	StudentName          string     `json:"studentName"`
	StudentPhone         string     `json:"studentPhone"`
	StudentAvatar        string     `json:"studentAvatar"`
	ProductID            string     `json:"productId"`
	ProductName          string     `json:"productName"`
	LessonType           *int       `json:"lessonType,omitempty"`
	LessonChargingMode   *int       `json:"lessonChargingMode,omitempty"`
	SourceType           int        `json:"sourceType"`
	SourceID             string     `json:"sourceId"`
	TeachingRecordID     string     `json:"teachingRecordId,omitempty"`
	CreatedTime          *time.Time `json:"createdTime,omitempty"`
	Quantity             float64    `json:"quantity"`
	Tuition              float64    `json:"tuition"`
}

type TuitionAccountFlowRecordListResult struct {
	List  []TuitionAccountFlowRecordItem `json:"list"`
	Total int                            `json:"total"`
}
