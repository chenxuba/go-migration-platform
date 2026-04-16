package model

import "time"

type StudentTeachingRecordPagedQueryDTO struct {
	PageRequestModel RollCallPageRequestModel        `json:"pageRequestModel"`
	SortModel        StudentTeachingRecordSortModel  `json:"sortModel"`
	QueryModel       StudentTeachingRecordQueryModel `json:"queryModel"`
}

type TeachingRecordDetailQueryDTO struct {
	TeachingRecordID string `json:"teachingRecordId"`
}

type UpdateTeachingRecordClassInfoDTO struct {
	TeachingRecordID string   `json:"teachingRecordId"`
	TeacherID        string   `json:"teacherId"`
	AssistantIDs     []string `json:"assistantIds"`
	ClassRoomID      string   `json:"classRoomId"`
	TeacherClassTime float64  `json:"teacherClassTime"`
}

type UpdateStudentTeachingRecordDTO struct {
	StudentTeachingRecordID string  `json:"studentTeachingRecordId"`
	TeachingRecordID        string  `json:"teachingRecordId"`
	StudentID               string  `json:"studentId"`
	SourceType              int     `json:"sourceType"`
	Status                  int     `json:"status"`
	Quantity                float64 `json:"quantity"`
	Remark                  string  `json:"remark"`
	ExternalRemark          string  `json:"externalRemark"`
}

type DeleteStudentTeachingRecordDTO struct {
	StudentTeachingRecordID string `json:"studentTeachingRecordId"`
}

type DeleteTeachingRecordDTO struct {
	TeachingRecordID string `json:"teachingRecordId"`
}

type StudentTeachingRecordQueryModel struct {
	BeginStartTime                string   `json:"beginStartTime"`
	EndStartTime                  string   `json:"endStartTime"`
	BeginCreateTime               string   `json:"beginCreateTime"`
	EndCreateTime                 string   `json:"endCreateTime"`
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
	ScheduleCallStatus            *int     `json:"scheduleCallStatus,omitempty"`
	IsArrear                      *bool    `json:"isArrear"`
	LessonIDs                     []string `json:"lessonIds"`
	ClassIDs                      []string `json:"classIds"`
	StudentTeachingRecordIDs      []string `json:"studentTeachingRecordIds"`
	TeachingRecordIDs             []string `json:"teachingRecordIds"`
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

type TeachingRecordDetailTeacher struct {
	TeacherID   string  `json:"teacherId"`
	TeacherName string  `json:"teacherName"`
	Type        int     `json:"type"`
	Status      int     `json:"status"`
	Quantity    float64 `json:"quantity"`
}

type TeachingRecordDetailStudent struct {
	StudentTeachingRecordID string  `json:"studentTeachingRecordId"`
	StudentID               string  `json:"studentId"`
	StudentName             string  `json:"studentName"`
	StudentPhone            string  `json:"studentPhone"`
	Avatar                  string  `json:"avatar"`
	Status                  int     `json:"status"`
	SourceType              int     `json:"sourceType"`
	Quantity                float64 `json:"quantity"`
	ActualQuantity          float64 `json:"actualQuantity"`
	Remark                  string  `json:"remark"`
	ExternalRemark          string  `json:"externalRemark"`
	TuitionAccountID        string  `json:"tuitionAccountId"`
	TuitionAccountName      string  `json:"tuitionAccountName"`
	IsTuitionAccountActive  bool    `json:"isTuitionAccountActive"`
	LeftQuantity            float64 `json:"leftQuantity"`
	SkuMode                 int     `json:"skuMode"`
	Amount                  float64 `json:"amount"`
	ActualDeduct            float64 `json:"actualDeduct"`
	ActualTuition           float64 `json:"actualTuition"`
	ArrearQuantity          float64 `json:"arrearQuantity"`
	RecordTime              string  `json:"recordTime"`
	UpdatedTime             string  `json:"updatedTime"`
	UpdatedStaffName        string  `json:"updatedStaffName"`
}

type TeachingRecordDetailResult struct {
	TeachingRecordID        string                        `json:"teachingRecordId"`
	SourceName              string                        `json:"sourceName"`
	SourceType              int                           `json:"sourceType"`
	SourceID                string                        `json:"sourceId"`
	LessonID                string                        `json:"lessonId"`
	LessonType              int                           `json:"lessonType"`
	StartTime               string                        `json:"startTime"`
	EndTime                 string                        `json:"endTime"`
	ShouldAttendanceCount   int                           `json:"shouldAttendanceCount"`
	ActualAttendanceCount   int                           `json:"actualAttendanceCount"`
	LeaveCount              int                           `json:"leaveCount"`
	TruancyCount            int                           `json:"truancyCount"`
	TeacherClassTime        float64                       `json:"teacherClassTime"`
	StudentTotalClassTime   float64                       `json:"studentTotalClassTime"`
	StudentActualTuition    float64                       `json:"studentActualTuition"`
	TeacherList             []TeachingRecordDetailTeacher `json:"teacherList"`
	StudentList             []TeachingRecordDetailStudent `json:"studentList"`
	CreatedTime             string                        `json:"createdTime"`
	CreatedStaffName        string                        `json:"createdStaffName"`
	TimetableSourceType     int                           `json:"timetableSourceType"`
	ClassRoomName           string                        `json:"classRoomName"`
	ClassRoomID             string                        `json:"classRoomId"`
	DefaultStudentClassTime float64                       `json:"defaultStudentClassTime"`
	TimetableSourceID       string                        `json:"timetableSourceId"`
	LessonName              string                        `json:"lessonName"`
	TeachingContent         string                        `json:"teachingContent"`
	SubjectID               string                        `json:"subjectId"`
	SubjectName             string                        `json:"subjectName"`
	TeachingContentImages   []string                      `json:"teachingContentImages"`
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
	TimetableSourceID   string  `json:"timetableSourceId,omitempty"`
	StartTime           string  `json:"startTime"`
	EndTime             string  `json:"endTime"`
	TimetableSourceType int     `json:"timetableSourceType"`
	SourceName          string  `json:"sourceName"`
	SourceType          int     `json:"sourceType"`
	SourceID            string  `json:"sourceId"`
	LessonID            string  `json:"lessonId"`
	ClassName           string  `json:"className"`
	One2OneName         string  `json:"one2OneName"`
	LessonName          string  `json:"lessonName"`
	SubjectID           string  `json:"subjectId"`
	SubjectName         string  `json:"subjectName"`
	RollCallStatus      int     `json:"rollCallStatus"`
	AttendanceRate      float64 `json:"attendanceRate"`
	AttendCount         int     `json:"attendCount"`
	ShouldAttendCount   int     `json:"shouldAttendCount"`
	LeaveCount          int     `json:"leaveCount"`
	AbsentCount         int     `json:"absentCount"`
	UnrecordedCount     int     `json:"unrecordedCount"`
	ActualQuantity      float64 `json:"actualQuantity"`
	ActualTuition       float64 `json:"actualTuition"`
	TeacherID           string  `json:"teacherId"`
	TeacherName         string  `json:"teacherName"`
	Assistants          string  `json:"assistants"`
	ClassRoomName       string  `json:"classRoomName"`
	TeacherClassTime    float64 `json:"teacherClassTime"`
	CommentCount        int     `json:"commentCount"`
	UnCommentCount      int     `json:"unCommentCount"`
	ReadCount           int     `json:"readCount"`
	UnReadCount         int     `json:"unReadCount"`
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

type ClassCommentPagedQueryDTO struct {
	PageRequestModel RollCallPageRequestModel `json:"pageRequestModel"`
	SortModel        ClassCommentSortModel    `json:"sortModel"`
	QueryModel       ClassCommentQueryModel   `json:"queryModel"`
}

type ClassCommentSortModel struct {
	StartTime int `json:"startTime"`
}

type ClassCommentQueryModel struct {
	TeachingStartTime   string   `json:"teachingStartTime"`
	TeachingEndTime     string   `json:"teachingEndTime"`
	TeachingRecordTypes []int    `json:"teachingRecordTypes"`
	LessonID            string   `json:"lessonId"`
	TeacherIDs          []string `json:"teacherIds"`
	ClassID             string   `json:"classId"`
	One2OneID           string   `json:"one2OneId"`
	ClassTeacherIDs     []string `json:"classTeacherIds"`
	One2OneTeacherIDs   []string `json:"one2OneTeacherIds"`
}

type ClassCommentPagedItem struct {
	TeachingRecordID string `json:"teachingRecordId"`
	SourceName       string `json:"sourceName"`
	SourceType       int    `json:"sourceType"`
	SourceID         string `json:"sourceId"`
	LessonID         string `json:"lessonId"`
	LessonName       string `json:"lessonName"`
	CreatedTime      string `json:"createdTime"`
	TeacherID        string `json:"teacherId"`
	TeacherName      string `json:"teacherName"`
	StartTime        string `json:"startTime"`
	EndTime          string `json:"endTime"`
	ReadCount        int    `json:"readCount"`
	UnReadCount      int    `json:"unReadCount"`
	CommentCount     int    `json:"commentCount"`
	UnCommentCount   int    `json:"unCommentCount"`
	Assistants       string `json:"assistants"`
	ClassRoomName    string `json:"classRoomName"`
}

type ClassCommentPagedResult struct {
	List  []ClassCommentPagedItem `json:"list"`
	Total int                     `json:"total"`
}

type ClassRecordExportCreateRequest struct {
	ExportType      string                `json:"exportType"`
	RecordIDs       []string              `json:"recordIds"`
	QueryConditions []ExportConditionItem `json:"queryConditions"`
}

type ClassRecordExportRecord struct {
	ID              int64                 `json:"id"`
	ExportType      string                `json:"exportType"`
	FileName        string                `json:"fileName"`
	ExporterName    string                `json:"exporterName"`
	TotalRows       int                   `json:"totalRows"`
	QueryConditions []ExportConditionItem `json:"queryConditions"`
	CreatedTime     *time.Time            `json:"createdTime,omitempty"`
	ExpiresAt       *time.Time            `json:"expiresAt,omitempty"`
	DownloadURL     string                `json:"downloadUrl,omitempty"`
}
