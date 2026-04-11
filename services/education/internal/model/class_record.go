package model

type StudentTeachingRecordPagedQueryDTO struct {
	PageRequestModel RollCallPageRequestModel        `json:"pageRequestModel"`
	SortModel        StudentTeachingRecordSortModel  `json:"sortModel"`
	QueryModel       StudentTeachingRecordQueryModel `json:"queryModel"`
}

type StudentTeachingRecordQueryModel struct {
	BeginStartTime                string   `json:"beginStartTime"`
	EndStartTime                  string   `json:"endStartTime"`
	BeginUpdatedTime              string   `json:"beginUpdatedTime"`
	EndUpdatedTime                string   `json:"endUpdatedTime"`
	StudentID                     string   `json:"studentId"`
	TeacherIDs                    []string `json:"teacherIds"`
	AssistantTeacherIDs           []string `json:"assistantTeacherIds"`
	One2OneIDs                    []string `json:"one2OneIds"`
	TimetableSourceTypes          []int    `json:"timetableSourceTypes"`
	StudentSourceTypes            []int    `json:"studentSourceTypes"`
	LessonChargingModeEnums       []int    `json:"lessonChargingModeEnums"`
	StudentTeachingRecordStatuses []int    `json:"studentTeachingRecordStatuses"`
	IsArrear                      *bool    `json:"isArrear"`
	LessonIDs                     []string `json:"lessonIds"`
	ClassIDs                      []string `json:"classIds"`
}

type StudentTeachingRecordSortModel struct {
	StartTime   int `json:"startTime"`
	UpdatedTime int `json:"updatedTime"`
}

type StudentTeachingRecordItem struct {
	StudentTeachingRecordID   string   `json:"studentTeachingRecordId"`
	TeachingRecordID          string   `json:"teachingRecordId"`
	StudentID                 string   `json:"studentId"`
	StudentName               string   `json:"studentName"`
	StudentPhone              string   `json:"studentPhone"`
	Avatar                    string   `json:"avatar"`
	TeacherName               string   `json:"teacherName"`
	TeacherEmployeeType       int      `json:"teacherEmployeeType"`
	Assistants                string   `json:"assistants"`
	ClassName                 string   `json:"className"`
	One2OneName               string   `json:"one2OneName"`
	LessonName                string   `json:"lessonName"`
	Status                    int      `json:"status"`
	SourceType                int      `json:"sourceType"`
	StartTime                 string   `json:"startTime"`
	EndTime                   string   `json:"endTime"`
	TeachingRecordCreatedTime string   `json:"teachingRecordCreatedTime"`
	TimetableSourceType       int      `json:"timetableSourceType"`
	UpdatedTime               string   `json:"updatedTime"`
	UpdatedStaffName          string   `json:"updatedStaffName"`
	RecordTime                string   `json:"recordTime"`
	Quantity                  float64  `json:"quantity"`
	ActualQuantity            float64  `json:"actualQuantity"`
	Amount                    float64  `json:"amount"`
	SkuMode                   int      `json:"skuMode"`
	ActualDeduct              float64  `json:"actualDeduct"`
	ActualTuition             float64  `json:"actualTuition"`
	ArrearQuantity            float64  `json:"arrearQuantity"`
	Remark                    string   `json:"remark"`
	ExternalRemark            string   `json:"externalRemark"`
	TuitionAccountID          string   `json:"tuitionAccountId"`
	TuitionAccountName        string   `json:"tuitionAccountName"`
	HasCompensated            bool     `json:"hasCompensated"`
	SubjectID                 string   `json:"subjectId"`
	SubjectName               string   `json:"subjectName"`
	AdvisorStaffID            string   `json:"advisorStaffId"`
	AdvisorStaffName          string   `json:"advisorStaffName"`
	StudentManagerID          string   `json:"studentManagerId"`
	StudentManagerName        string   `json:"studentManagerName"`
	TeachingContent           string   `json:"teachingContent"`
	TeachingContentImages     []string `json:"teachingContentImages"`
	ClassRoomName             string   `json:"classRoomName"`
	One2OneTeachers           string   `json:"one2OneTeachers"`
	ClassTeachers             string   `json:"classTeachers"`
	RollCallClassTeachers     string   `json:"rollCallClassTeachers"`
	CurrentClassTeachers      string   `json:"currentClassTeachers"`
}

type StudentTeachingRecordPagedResult struct {
	TotalClassTimes   float64                     `json:"totalClassTimes"`
	TotalTuition      float64                     `json:"totalTuition"`
	TotalStudentCount int                         `json:"totalStudentCount"`
	List              []StudentTeachingRecordItem `json:"list"`
	Total             int                         `json:"total"`
}

type ScheduleTeachingRecordPagedQueryDTO struct {
	PageRequestModel RollCallPageRequestModel        `json:"pageRequestModel"`
	SortModel        ScheduleTeachingRecordSortModel `json:"sortModel"`
	QueryModel       StudentTeachingRecordQueryModel `json:"queryModel"`
}

type ScheduleTeachingRecordSortModel struct {
	StartTime   int `json:"startTime"`
	UpdatedTime int `json:"updatedTime"`
}

type ScheduleTeachingRecordItem struct {
	TeachingRecordID    string  `json:"teachingRecordId"`
	StartTime           string  `json:"startTime"`
	EndTime             string  `json:"endTime"`
	TimetableSourceType int     `json:"timetableSourceType"`
	ClassName           string  `json:"className"`
	One2OneName         string  `json:"one2OneName"`
	LessonName          string  `json:"lessonName"`
	SubjectID           string  `json:"subjectId"`
	SubjectName         string  `json:"subjectName"`
	RollCallStatus      int     `json:"rollCallStatus"`
	AttendanceRate      float64 `json:"attendanceRate"`
	AttendCount         int     `json:"attendCount"`
	ShouldAttendCount   int     `json:"shouldAttendCount"`
	ActualQuantity      float64 `json:"actualQuantity"`
	ActualTuition       float64 `json:"actualTuition"`
	TeacherName         string  `json:"teacherName"`
	Assistants          string  `json:"assistants"`
	TeacherClassTime    float64 `json:"teacherClassTime"`
	CreatedTime         string  `json:"createdTime"`
	UpdatedTime         string  `json:"updatedTime"`
}

type ScheduleTeachingRecordPagedResult struct {
	TotalClassTimes   float64                      `json:"totalClassTimes"`
	TotalTeacherTimes float64                      `json:"totalTeacherTimes"`
	TotalTuition      float64                      `json:"totalTuition"`
	List              []ScheduleTeachingRecordItem `json:"list"`
	Total             int                          `json:"total"`
}
