package model

import "time"

type LessonIncomeQueryDTO struct {
	PageRequestModel PageRequestModel      `json:"pageRequestModel"`
	QueryModel       LessonIncomeQueryVO   `json:"queryModel"`
	SortModel        LessonIncomeSortModel `json:"sortModel"`
}

type LessonIncomeQueryVO struct {
	StartDate                  string `json:"startDate"`
	EndDate                    string `json:"endDate"`
	SourceTypes                []int  `json:"sourceTypes"`
	StudentID                  string `json:"studentId"`
	StaffID                    string `json:"staffId"`
	LessonID                   string `json:"lessonId"`
	LessonDayStartDate         string `json:"lessonDayStartDate"`
	LessonDayEndDate           string `json:"lessonDayEndDate"`
	ClassID                    string `json:"classId"`
	ProductCategoryID          string `json:"productCategoryId"`
	ConformIncomeTimeStartDate string `json:"conformIncomeTimeStartDate"`
	ConformIncomeTimeEndDate   string `json:"conformIncomeTimeEndDate"`
}

type LessonIncomeSortModel struct {
	OrderByCreatedTime int `json:"orderByCreatedTime"`
}

type LessonIncomeTeacher struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type LessonIncomeItem struct {
	ID                  string                `json:"id"`
	StudentID           string                `json:"studentId"`
	StudentName         string                `json:"studentName"`
	StudentPhone        string                `json:"studentPhone,omitempty"`
	StudentAvatar       string                `json:"studentAvatar,omitempty"`
	LessonID            string                `json:"lessonId"`
	LessonName          string                `json:"lessonName"`
	LessonType          *int                  `json:"lessonType,omitempty"`
	TeachingMethod      *int                  `json:"teachingMethod,omitempty"`
	SourceType          int                   `json:"sourceType"`
	LessonDay           *time.Time            `json:"lessonDay,omitempty"`
	StartMinutes        int                   `json:"startMinutes"`
	EndMinutes          int                   `json:"endMinutes"`
	TeachingTime        *time.Time            `json:"teachingTime,omitempty"`
	RollCallTime        *time.Time            `json:"rollCallTime,omitempty"`
	Quantity            float64               `json:"quantity"`
	LessonChargingMode  *int                  `json:"lessonChargingMode,omitempty"`
	Tuition             float64               `json:"tuition"`
	CreatedTime         *time.Time            `json:"createdTime,omitempty"`
	Teachers            []LessonIncomeTeacher `json:"teachers,omitempty"`
	TeacherName         string                `json:"teacherName,omitempty"`
	AssistantTeachers   []LessonIncomeTeacher `json:"assistantTeachers,omitempty"`
	AssistantName       string                `json:"assistantName,omitempty"`
	ProductCategoryID   string                `json:"productCategoryId,omitempty"`
	ProductCategoryName string                `json:"productCategoryName,omitempty"`
	ClassID             string                `json:"classId,omitempty"`
	ClassName           string                `json:"className,omitempty"`
	ConformIncomeTime   *time.Time            `json:"conformIncomeTime,omitempty"`
	TeachingRecordID    string                `json:"teachingRecordId,omitempty"`
}

type LessonIncomePagedResult struct {
	List  []LessonIncomeItem `json:"list"`
	Total int                `json:"total"`
}

type LessonIncomeStatistics struct {
	TotalCount   int     `json:"totalCount"`
	TotalTuition float64 `json:"totalTuition"`
}

const (
	LessonIncomeSourceLessonConsume      = 1
	LessonIncomeSourceManualGraduate     = 2
	LessonIncomeSourceExpireGraduate     = 3
	LessonIncomeSourceImportConsume      = 4
	LessonIncomeSourceConsumeReturn      = 5
	LessonIncomeSourceConsumeSupplement  = 6
	LessonIncomeSourceDailyAutoConsume   = 7
	LessonIncomeSourceConsumeArrearClear = 8
	LessonIncomeSourceRefundFee          = 9
	LessonIncomeSourceRevokeGraduate     = 10
	LessonIncomeSourceExpireRollback     = 11
	LessonIncomeSourceVoidReturn         = 12
	LessonIncomeSourceRevokeRefundFee    = 13
	LessonIncomeSourceManualCloseCourse  = 14
)

var lessonIncomeSourceTypeToInternal = map[int][]int{
	LessonIncomeSourceLessonConsume:      {TuitionAccountFlowSourceConsume},
	LessonIncomeSourceManualGraduate:     {TuitionAccountFlowSourceGraduate},
	LessonIncomeSourceExpireGraduate:     {TuitionAccountFlowSourceExpireGraduate},
	LessonIncomeSourceImportConsume:      {TuitionAccountFlowSourceImportConsume},
	LessonIncomeSourceConsumeReturn:      {TuitionAccountFlowSourceConsumeReturn},
	LessonIncomeSourceConsumeSupplement:  {TuitionAccountFlowSourceConsumeSupplement},
	LessonIncomeSourceDailyAutoConsume:   {TuitionAccountFlowSourceAutoConsume},
	LessonIncomeSourceConsumeArrearClear: {TuitionAccountFlowSourceConsumeArrearsSettlement},
	LessonIncomeSourceRefundFee:          {TuitionAccountFlowSourceRefund},
	LessonIncomeSourceRevokeGraduate:     {TuitionAccountFlowSourceRevokeGraduate},
	LessonIncomeSourceExpireRollback:     {TuitionAccountFlowSourceExpireRollback},
	LessonIncomeSourceVoidReturn:         {TuitionAccountFlowSourceOrderVoid},
	LessonIncomeSourceRevokeRefundFee:    {TuitionAccountFlowSourceRevokeRefundOrder},
	LessonIncomeSourceManualCloseCourse:  {TuitionAccountFlowSourceManualCloseCourse},
}

var lessonIncomeInternalToSourceType = map[int]int{
	TuitionAccountFlowSourceConsume:                  LessonIncomeSourceLessonConsume,
	TuitionAccountFlowSourceGraduate:                 LessonIncomeSourceManualGraduate,
	TuitionAccountFlowSourceExpireGraduate:           LessonIncomeSourceExpireGraduate,
	TuitionAccountFlowSourceImportConsume:            LessonIncomeSourceImportConsume,
	TuitionAccountFlowSourceConsumeReturn:            LessonIncomeSourceConsumeReturn,
	TuitionAccountFlowSourceConsumeSupplement:        LessonIncomeSourceConsumeSupplement,
	TuitionAccountFlowSourceAutoConsume:              LessonIncomeSourceDailyAutoConsume,
	TuitionAccountFlowSourceConsumeArrearsSettlement: LessonIncomeSourceConsumeArrearClear,
	TuitionAccountFlowSourceRefund:                   LessonIncomeSourceRefundFee,
	TuitionAccountFlowSourceRevokeGraduate:           LessonIncomeSourceRevokeGraduate,
	TuitionAccountFlowSourceExpireRollback:           LessonIncomeSourceExpireRollback,
	TuitionAccountFlowSourceOrderVoid:                LessonIncomeSourceVoidReturn,
	TuitionAccountFlowSourceRevokeRefundOrder:        LessonIncomeSourceRevokeRefundFee,
	TuitionAccountFlowSourceManualCloseCourse:        LessonIncomeSourceManualCloseCourse,
}

func ExpandLessonIncomeSourceTypes(sourceTypes []int) []int {
	if len(sourceTypes) == 0 {
		result := make([]int, 0, len(lessonIncomeInternalToSourceType))
		for internal := range lessonIncomeInternalToSourceType {
			result = append(result, internal)
		}
		return result
	}

	result := make([]int, 0, len(sourceTypes))
	seen := make(map[int]struct{})
	for _, sourceType := range sourceTypes {
		internalList, ok := lessonIncomeSourceTypeToInternal[sourceType]
		if !ok {
			internalList = []int{sourceType}
		}
		for _, internal := range internalList {
			if _, exists := seen[internal]; exists {
				continue
			}
			seen[internal] = struct{}{}
			result = append(result, internal)
		}
	}
	return result
}

func CompressLessonIncomeSourceType(internalSourceType int) int {
	if mapped, ok := lessonIncomeInternalToSourceType[internalSourceType]; ok {
		return mapped
	}
	return internalSourceType
}
