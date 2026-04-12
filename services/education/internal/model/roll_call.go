package model

type RollCallQueryModel struct {
	StartDate      string   `json:"startDate"`
	EndDate        string   `json:"endDate"`
	LessonID       string   `json:"lessonId"`
	ClassroomID    string   `json:"classroomId"`
	ClassID        string   `json:"classId"`
	OneToOneID     string   `json:"oneToOneId"`
	TeacherID      string   `json:"teacherId"`
	TeacherTypes   []int    `json:"teacherTypes"`
	ScheduleTypes  []string `json:"scheduleTypes"`
	CallStatusMode string   `json:"callStatusMode,omitempty"`
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

type RollCallClassTimetableQueryDTO struct {
	ID        string `json:"id"`
	LessonDay string `json:"lessonDay"`
}

type RollCallClassTimetableTeacherVO struct {
	TeacherID     string `json:"teacherId"`
	TeacherDuty   int    `json:"teacherDuty"`
	TeacherName   string `json:"teacherName"`
	TeacherStatus int    `json:"teacherStatus"`
}

type RollCallClassTimetableStudentVO struct {
	SourceType                   int    `json:"sourceType"`
	SourceID                     string `json:"sourceId"`
	StudentID                    string `json:"studentId"`
	StudentName                  string `json:"studentName"`
	StudentAvatar                string `json:"studentAvatar,omitempty"`
	StudentPhone                 string `json:"studentPhone,omitempty"`
	StudentPhoneRelationshipType int    `json:"studentPhoneRelationshipType"`
}

type RollCallClassTimetableLessonDayVO struct {
	LessonDay        string                            `json:"lessonDay"`
	IsFinished       bool                              `json:"isFinished"`
	LessonDayIndex   int                               `json:"lessonDayIndex"`
	Students         []RollCallClassTimetableStudentVO `json:"students"`
	RemoveStudent    []RollCallClassTimetableStudentVO `json:"removeStudent"`
	TeachingRecordID string                            `json:"teachingRecordId"`
}

type RollCallClassTimetableDetailVO struct {
	ID                         string                              `json:"id"`
	ClassID                    string                              `json:"classId"`
	ClassName                  string                              `json:"className"`
	ClassTimes                 float64                             `json:"classTimes"`
	DefaultStudentClassTime    float64                             `json:"defaultStudentClassTime"`
	DefaultTeacherClassTime    float64                             `json:"defaultTeacherClassTime"`
	DefaultClassTimeRecordMode int                                 `json:"defaultClassTimeRecordMode"`
	LessonPrice                float64                             `json:"lessonPrice"`
	Teachers                   []RollCallClassTimetableTeacherVO   `json:"teachers"`
	AddressType                int                                 `json:"addressType"`
	AddressID                  string                              `json:"addressId"`
	AddressName                string                              `json:"addressName"`
	LessonID                   string                              `json:"lessonId"`
	LessonName                 string                              `json:"lessonName"`
	LessonType                 int                                 `json:"lessonType"`
	StartMinutes               int                                 `json:"startMinutes"`
	EndMinutes                 int                                 `json:"endMinutes"`
	RepeatSpan                 int                                 `json:"repeatSpan"`
	WeekDays                   int                                 `json:"weekDays"`
	StartDate                  string                              `json:"startDate"`
	EndDate                    string                              `json:"endDate"`
	LessonCount                int                                 `json:"lessonCount"`
	Remark                     string                              `json:"remark"`
	ExternalRemark             string                              `json:"externalRemark"`
	LessonDays                 []RollCallClassTimetableLessonDayVO `json:"lessonDays"`
	IsBookLesson               bool                                `json:"isBookLesson"`
	SubjectID                  string                              `json:"subjectId"`
	SubjectName                string                              `json:"subjectName"`
	IsOrgCreated               bool                                `json:"isOrgCreated"`
	SchoolID                   string                              `json:"schoolId"`
	SchoolName                 string                              `json:"schoolName"`
	IsOpenLiveRecord           bool                                `json:"isOpenLiveRecord"`
	IsOpenLive                 bool                                `json:"isOpenLive"`
}

type RollCallClassTimetableResult struct {
	Detail RollCallClassTimetableDetailVO `json:"detail"`
}

type RollCallTeachingRecordStudentListQueryDTO struct {
	TimetableSourceID   string `json:"timetableSourceId"`
	TimetableSourceType int    `json:"timetableSourceType"`
	ClassID             string `json:"classId"`
	LessonID            string `json:"lessonId"`
	OneToOneID          string `json:"one2OneId"`
	StartDate           string `json:"startDate"`
	EndDate             string `json:"endDate"`
	LessonDay           string `json:"lessonDay"`
}

type RollCallTeachingRecordMetaVO struct {
	SourceName          string  `json:"sourceName"`
	SourceType          int     `json:"sourceType"`
	SourceID            string  `json:"sourceId"`
	LessonID            string  `json:"lessonId"`
	TimetableSourceType int     `json:"timetableSourceType"`
	Tag                 int     `json:"tag"`
	TimetableSourceID   string  `json:"timetableSourceId"`
	StartTime           string  `json:"startTime"`
	EndTime             string  `json:"endTime"`
	TeacherClassTime    float64 `json:"teacherClassTime"`
	ClassroomID         string  `json:"classroomId"`
}

type RollCallTeachingRecordTeacherVO struct {
	TeacherID string `json:"teacherId"`
	Type      int    `json:"type"`
}

type RollCallTeachingRecordStudentVO struct {
	StudentID                    string  `json:"studentId"`
	StudentName                  string  `json:"studentName"`
	Avatar                       string  `json:"avatar,omitempty"`
	IsBindChild                  bool    `json:"isBindChild"`
	Quantity                     float64 `json:"quantity"`
	PaidRemaining                float64 `json:"paidRemaining"`
	ChargingMode                 int     `json:"chargingMode"`
	IsTuitionAccountActive       bool    `json:"isTuitionAccountActive"`
	MakeUpTeachingRecordID       string  `json:"makeUpTeachingRecordId"`
	AbsentStudentType            int     `json:"absentStudentType"`
	TuitionAccountID             string  `json:"tuitionAccountId"`
	SourceType                   int     `json:"sourceType"`
	StudentTeachingStatus        int     `json:"studentTeachingStatus"`
	DefaultStudentTeachingStatus int     `json:"defaultStudentTeachingStatus"`
	HasSignIn                    bool    `json:"hasSignIn"`
	IsCrossSchoolStudent         bool    `json:"isCrossSchoolStudent"`
	HasTeachingRecord            bool    `json:"hasTeachingRecord"`
	RecordedQuantity             float64 `json:"recordedQuantity"`
	RecordedRemark               string  `json:"recordedRemark,omitempty"`
	RecordedExternalRemark       string  `json:"recordedExternalRemark,omitempty"`
	RecordedSkuMode              int     `json:"recordedSkuMode"`
	RecordedTuitionAccountID     string  `json:"recordedTuitionAccountId,omitempty"`
	RecordedTuitionAccountName   string  `json:"recordedTuitionAccountName,omitempty"`
	Locked                       bool    `json:"locked"`
	AutoRollCall                 bool    `json:"autoRollCall"`
}

type RollCallTeachingRecordStudentListResult struct {
	Data     RollCallTeachingRecordMetaVO      `json:"data"`
	Teachers []RollCallTeachingRecordTeacherVO `json:"teachers"`
	Students []RollCallTeachingRecordStudentVO `json:"students"`
}

type RollCallStudentLeaveCountQueryDTO struct {
	StudentIDs []string `json:"studentIds"`
	LessonID   string   `json:"lessonId"`
}

type RollCallStudentLeaveCountVO struct {
	StudentID  string `json:"studentId"`
	LeaveCount int    `json:"leaveCount"`
}

type RollCallStudentTuitionExtraInfoQueryDTO struct {
	StudentIDs []string `json:"studentIds"`
	LessonID   string   `json:"lessonId"`
}

type RollCallStudentTuitionExtraInfoVO struct {
	StudentID            string `json:"studentId"`
	MutilTuition         bool   `json:"mutilTuition"`
	BestMatchProductName string `json:"bestMatchProductName"`
}

type RollCallCheckTeachingRecordByTeacherAndTimeDTO struct {
	StartTime         string `json:"startTime"`
	EndTime           string `json:"endTime"`
	TeacherID         string `json:"teacherId"`
	TimetableSourceID string `json:"timetableSourceId"`
}

type RollCallEstimateTuitionInfo struct {
	Quantity         float64 `json:"quantity"`
	TuitionAccountID string  `json:"tuitionAccountId"`
	StudentName      string  `json:"studentName"`
}

type RollCallBatchEstimateSufficientTuitionAccountDTO struct {
	TuitionInfoList []RollCallEstimateTuitionInfo `json:"tuitionInfoList"`
}

type RollCallEstimateTuitionResultItem struct {
	TuitionAccountID string `json:"tuitionAccountId"`
	IsSufficient     bool   `json:"isSufficient"`
}

type RollCallBatchEstimateSufficientTuitionAccountResult struct {
	TuitionInfoList []RollCallEstimateTuitionResultItem `json:"tuitionInfoList"`
}

type RollCallConfirmTeacher struct {
	TeacherID string `json:"teacherId"`
	Type      int    `json:"type"`
}

type RollCallConfirmStudent struct {
	StudentShouldDeduct    int     `json:"studentShouldDeduct"`
	StudentName            string  `json:"studentName"`
	StudentID              string  `json:"studentId"`
	TuitionAccountID       string  `json:"tuitionAccountId"`
	AbsentTeachingRecordID string  `json:"absentTeachingRecordId"`
	Status                 int     `json:"status"`
	SourceType             int     `json:"sourceType"`
	Remark                 string  `json:"remark"`
	ExternalRemark         string  `json:"externalRemark"`
	SkuMode                int     `json:"skuMode"`
	Amount                 float64 `json:"amount"`
	Quantity               float64 `json:"quantity"`
}

type RollCallConfirmDTO struct {
	SourceName            string                   `json:"sourceName"`
	TeachingContent       string                   `json:"teachingContent"`
	TeachingContentImages []string                 `json:"teachingContentImages"`
	TimetableSourceType   int                      `json:"timetableSourceType"`
	TimetableSourceID     string                   `json:"timetableSourceId"`
	SourceID              string                   `json:"sourceId"`
	SourceType            int                      `json:"sourceType"`
	LessonID              string                   `json:"lessonId"`
	StartTime             string                   `json:"startTime"`
	EndTime               string                   `json:"endTime"`
	TeacherClassTime      float64                  `json:"teacherClassTime"`
	StudentShouldDeduct   int                      `json:"studentShouldDeduct"`
	StudentList           []RollCallConfirmStudent `json:"studentList"`
	TeacherList           []RollCallConfirmTeacher `json:"teacherList"`
	SubjectID             string                   `json:"subjectId"`
	ClassRoomID           string                   `json:"classRoomId"`
}

type RollCallConfirmResult struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
