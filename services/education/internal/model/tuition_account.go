package model

import "time"

type TuitionAccountReadingListQueryDTO struct {
	PageRequestModel PageRequestModel                `json:"pageRequestModel"`
	QueryModel       TuitionAccountReadingQueryModel `json:"queryModel"`
}

type TuitionAccountReadingQueryModel struct {
	StudentID string `json:"studentId"`
}

type TuitionAccountReadingItem struct {
	ID                 string     `json:"id"`
	LessonID           string     `json:"lessonId"`
	LessonName         string     `json:"lessonName"`
	LessonType         *int       `json:"lessonType,omitempty"`
	TotalQuantity      float64    `json:"totalQuantity"`
	TotalFreeQuantity  float64    `json:"totalFreeQuantity"`
	TotalTuition       float64    `json:"totalTuition"`
	ArrearTuition      float64    `json:"arrearTuition"`
	IsAdjustable       bool       `json:"isAdjustable"`
	RemainQuantity     float64    `json:"remainQuantity"`
	Tuition            float64    `json:"tuition"`
	RemainFreeQuantity float64    `json:"remainFreeQuantity"`
	EnableExpireTime   bool       `json:"enableExpireTime"`
	ExpireTime         *time.Time `json:"expireTime,omitempty"`
	ValidDate          *time.Time `json:"validDate,omitempty"`
	EndDate            *time.Time `json:"endDate,omitempty"`
	ActivedAt          *time.Time `json:"activedAt,omitempty"`
	AssignedClass      bool       `json:"assignedClass"`
	Status             *int       `json:"status,omitempty"`
	ChangeStatusTime   *time.Time `json:"changeStatusTime,omitempty"`
	LessonChargingMode *int       `json:"lessonChargingMode,omitempty"`
	LessonScope        *int       `json:"lessonScope,omitempty"`
	PlanSuspendTime    *time.Time `json:"planSuspendTime,omitempty"`
	PlanResumeTime     *time.Time `json:"planResumeTime,omitempty"`
	HasGradeUpgrade    bool       `json:"hasGradeUpgrade"`
	ManualSort         bool       `json:"manualSort"`
}

type TuitionAccountReadingListResult struct {
	List  []TuitionAccountReadingItem `json:"list"`
	Total int                         `json:"total"`
}
