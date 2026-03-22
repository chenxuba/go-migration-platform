package model

import "time"

type StudentFollowUpQueryDTO struct {
	PageRequestModel PageRequestModel       `json:"pageRequestModel"`
	QueryModel       StudentFollowUpFilters `json:"queryModel"`
	SortModel        SortModel              `json:"sortModel"`
}

type StudentFollowUpFilters struct {
	QuickFilter           *int    `json:"quickFilter"`
	QueryAllOrDepartment  *int    `json:"queryAllOrDepartment"`
	DeptID                *int64  `json:"deptId"`
	StudentID             *int64  `json:"studentId"`
	FollowUpStaffID       *int64  `json:"followUpStaffId"`
	SalespersonID         *int64  `json:"salespersonId"`
	SearchKey             string  `json:"searchKey"`
	Sexes                 []int   `json:"sexes"`
	FollowUpTimeBegin     string  `json:"followUpTimeBegin"`
	FollowUpTimeEnd       string  `json:"followUpTimeEnd"`
	NextFollowUpTimeBegin string  `json:"nextFollowUpTimeBegin"`
	NextFollowUpTimeEnd   string  `json:"nextFollowUpTimeEnd"`
	FollowUpTypes         []int   `json:"followUpTypes"`
	VisitStatuses         []int   `json:"visitStatuses"`
	ChannelIDs            []int64 `json:"channelIds"`
	StudentStatuses       []int   `json:"studentStatuses"`
}

// FollowUpIntentionLesson 跟进记录上的意向课程（给前端展示名称，与旧版 intentionLessonList 字段对齐）
type FollowUpIntentionLesson struct {
	LessonID   int64  `json:"lessonId"`
	LessonName string `json:"lessonName"`
}

type StudentFollowUpRecord struct {
	ID                int64      `json:"id"`
	StudentID         int64      `json:"studentId"`
	StuName           string     `json:"stuName"`
	StuSex            *int       `json:"stuSex,omitempty"`
	AvatarURL         string     `json:"avatarUrl,omitempty"`
	Mobile            string     `json:"mobile"`
	PhoneRelationship *int       `json:"phoneRelationship,omitempty"`
	StudentStatus     int        `json:"studentStatus"`
	SalesPersonID     *int64     `json:"salesPersonId,omitempty"`
	SalesPersonName   string     `json:"salesPersonName,omitempty"`
	ChannelID         *int64     `json:"channelId,omitempty"`
	ChannelName       string     `json:"channelName,omitempty"`
	CategoryID        *int64     `json:"categoryId,omitempty"`
	CategoryName      string     `json:"categoryName,omitempty"`
	CreateID          *int64     `json:"createId,omitempty"`
	CreateName        string     `json:"createName,omitempty"`
	CreateTime        time.Time  `json:"createTime"`
	Content           string     `json:"content"`
	FollowImages      string     `json:"followImages,omitempty"`
	FollowMethod      *int       `json:"followMethod,omitempty"`
	IntendedCourse      []int64                   `json:"intendedCourse,omitempty"`
	IntendedCourseName  []string                  `json:"intendedCourseName,omitempty"`
	IntentionLessonList []FollowUpIntentionLesson `json:"intentionLessonList,omitempty"`
	IntentionLevel    *int       `json:"intentionLevel,omitempty"`
	FollowUpStatus    *int       `json:"followUpStatus,omitempty"`
	VisitStatus       *bool      `json:"visitStatus,omitempty"`
	FollowUpTime      *time.Time `json:"followUpTime,omitempty"`
	NextFollowUpTime  *time.Time `json:"nextFollowUpTime,omitempty"`
}

type CreateFollowUpDTO struct {
	StudentID        int64      `json:"studentId"`
	FollowMethod     *int       `json:"followMethod"`
	IntentLevel      *int       `json:"intentLevel"`
	NextFollowUpTime *time.Time `json:"nextFollowUpTime"`
	FollowUpStatus   *int       `json:"followUpStatus"`
	Content          string     `json:"content"`
	FollowImages     string     `json:"followImages"`
	IntentCourseIDs  string     `json:"intentCourseIds"`
}

type FollowUpCountDTO struct {
	DeptID               *int64 `json:"deptId"`
	QueryAllOrDepartment *int   `json:"queryAllOrDepartment"`
}

type FollowUpCountVO struct {
	ToBeFollowedUpTodayCount         int `json:"toBeFollowedUpTodayCount"`
	NewInquiriesAddedWeekCount       int `json:"newInquiriesAddedWeekCount"`
	OverdueForFollowUpInterviewCount int `json:"overdueForFollowUpInterviewCount"`
}

type VisitStatusUpdateDTO struct {
	ID          int64 `json:"id"`
	VisitStatus *bool `json:"visitStatus"`
}

type UpdateFollowUpDTO struct {
	ID               int64      `json:"id"`
	UUID             string     `json:"uuid"`
	Version          *int64     `json:"version"`
	FollowMethod     *int       `json:"followMethod"`
	IntentLevel      *int       `json:"intentLevel"`
	NextFollowUpTime *time.Time `json:"nextFollowUpTime"`
	FollowUpStatus   *int       `json:"followUpStatus"`
	Content          string     `json:"content"`
	FollowImages     string     `json:"followImages"`
	IntentCourseIDs  string     `json:"intentCourseIds"`
}

type FollowVisitCountVO struct {
	StudentCount      int `json:"studentCount"`
	InterviewCount    int `json:"interviewCount"`
	NotInterviewCount int `json:"notInterviewCount"`
}
