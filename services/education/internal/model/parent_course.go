package model

type ParentCourseEnrollmentQueryDTO struct {
	StudentID string `json:"studentId"`
}

type ParentCourseEnrollmentDetailQueryDTO struct {
	StudentID    string `json:"studentId"`
	LessonID     string `json:"lessonId"`
	ChargingMode int    `json:"chargingMode"`
	PageIndex    int    `json:"pageIndex"`
	PageSize     int    `json:"pageSize"`
}

type ParentCourseEnrollmentSummaryVO struct {
	Students []ParentBoundStudentVO     `json:"students"`
	Items    []ParentCourseEnrollmentVO `json:"items"`
}

type ParentCourseEnrollmentDetailVO struct {
	Student   ParentBoundStudentVO           `json:"student"`
	Course    ParentCourseEnrollmentVO       `json:"course"`
	Items     []ParentCourseEnrollmentFlowVO `json:"items"`
	PageIndex int                            `json:"pageIndex"`
	PageSize  int                            `json:"pageSize"`
	Total     int                            `json:"total"`
	HasMore   bool                           `json:"hasMore"`
}

type ParentCourseEnrollmentVO struct {
	ID                     string  `json:"id"`
	InstID                 int64   `json:"instId"`
	CampusID               string  `json:"campusId"`
	CampusName             string  `json:"campusName"`
	StudentID              string  `json:"studentId"`
	StudentName            string  `json:"studentName"`
	StudentAvatarURL       string  `json:"studentAvatarUrl,omitempty"`
	LessonID               string  `json:"lessonId"`
	LessonName             string  `json:"lessonName"`
	ChargingMode           int     `json:"chargingMode"`
	ChargingModeText       string  `json:"chargingModeText"`
	Status                 int     `json:"status"`
	StatusText             string  `json:"statusText"`
	TotalQuantity          float64 `json:"totalQuantity"`
	RemainingQuantity      float64 `json:"remainingQuantity"`
	TotalTuition           float64 `json:"totalTuition"`
	RemainingTuition       float64 `json:"remainingTuition"`
	ShowRemainingQuantity  bool    `json:"showRemainingQuantity"`
	RemainingQuantityLabel string  `json:"remainingQuantityLabel"`
	RemainingQuantityText  string  `json:"remainingQuantityText"`
	RemainingTuitionText   string  `json:"remainingTuitionText"`
	ValidDate              string  `json:"validDate"`
	EndDate                string  `json:"endDate"`
	ShowValidRange         bool    `json:"showValidRange"`
	ValidRangeText         string  `json:"validRangeText"`
	LowBalance             bool    `json:"lowBalance"`
	LowBalanceText         string  `json:"lowBalanceText"`
}

type ParentCourseEnrollmentFlowVO struct {
	ID                string  `json:"id"`
	SourceType        int     `json:"sourceType"`
	SourceID          string  `json:"sourceId"`
	Title             string  `json:"title"`
	CreatedAt         string  `json:"createdAt"`
	Quantity          float64 `json:"quantity"`
	Tuition           float64 `json:"tuition"`
	QuantityText      string  `json:"quantityText"`
	HighlightPositive bool    `json:"highlightPositive"`
}
