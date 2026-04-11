package model

type RollCallQueryModel struct {
	StartDate     string   `json:"startDate"`
	EndDate       string   `json:"endDate"`
	LessonID      string   `json:"lessonId"`
	ClassroomID   string   `json:"classroomId"`
	ClassID       string   `json:"classId"`
	OneToOneID    string   `json:"oneToOneId"`
	TeacherID     string   `json:"teacherId"`
	TeacherTypes  []int    `json:"teacherTypes"`
	ScheduleTypes []string `json:"scheduleTypes"`
}

type RollCallPageRequestModel struct {
	NeedTotal bool `json:"needTotal"`
	PageSize  int  `json:"pageSize"`
	PageIndex int  `json:"pageIndex"`
	SkipCount int  `json:"skipCount"`
}

type RollCallSortModel struct {
	ByStartDate int `json:"byStartDate"`
}

type RollCallStatisticsQueryDTO struct {
	QueryModel RollCallQueryModel `json:"queryModel"`
}

type RollCallPagedListQueryDTO struct {
	QueryModel       RollCallQueryModel       `json:"queryModel"`
	PageRequestModel RollCallPageRequestModel `json:"pageRequestModel"`
	SortModel        RollCallSortModel        `json:"sortModel"`
}

type RollCallStatisticsVO struct {
	TodayCount   int `json:"todayCount"`
	AllCount     int `json:"allCount"`
	PartialCount int `json:"partialCount"`
}

type RollCallPagedListResult struct {
	List  []TeachingScheduleVO `json:"list"`
	Total int                  `json:"total"`
}
