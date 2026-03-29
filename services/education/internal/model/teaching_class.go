package model

import "time"

const (
	TeachingClassTypeNormal   = 1
	TeachingClassTypeOneToOne = 2
)

const (
	TeachingClassStatusActive = 1
	TeachingClassStatusClosed = 2
)

const (
	TeachingClassStudentStatusStudying = 1
	TeachingClassStudentStatusStopped  = 2
	TeachingClassStudentStatusClosed   = 3
)

type OneToOneListQueryDTO struct {
	PageRequestModel PageRequestModel       `json:"pageRequestModel"`
	QueryModel       OneToOneListQueryModel `json:"queryModel"`
}

type OneToOneListQueryModel struct {
	StudentID          string   `json:"studentId"`
	LessonIDs          []string `json:"lessonIds"`
	ClassTeacherID     string   `json:"classTeacherId"`
	DefaultTeacherID   string   `json:"defaultTeacherId"`
	HasClassTeacher    *bool    `json:"hasClassTeacher"`
	IsScheduled        *bool    `json:"isScheduled"`
	Status             []int    `json:"status"`
	ClassStudentStatus []int    `json:"classStudentStatus"`
	StartDate          string   `json:"startDate"`
	EndDate            string   `json:"endDate"`
}

type OneToOneLessonTimeVO struct {
	ID        string     `json:"id,omitempty"`
	StartTime *time.Time `json:"startTime,omitempty"`
	EndTime   *time.Time `json:"endTime,omitempty"`
}

type OneToOneLessonDayInfoVO struct {
	LessonDayCount         int `json:"lessonDayCount"`
	CompleteLessonDayCount int `json:"completeLessonDayCount"`
}

type OneToOneTeacherVO struct {
	TeacherID string `json:"teacherId"`
	Name      string `json:"name"`
	Status    int    `json:"status"`
	ClassID   string `json:"classId"`
	IsDefault bool   `json:"isDefault,omitempty"`
}

type OneToOneTuitionAccountVO struct {
	ID                 string     `json:"id"`
	TotalTuition       float64    `json:"totalTuition"`
	RemainTuition      float64    `json:"remainTuition"`
	TotalQuantity      float64    `json:"totalQuantity"`
	TotalFreeQuantity  float64    `json:"totalFreeQuantity"`
	RemainQuantity     float64    `json:"remainQuantity"`
	RemainFreeQuantity float64    `json:"remainFreeQuantity"`
	LessonChargingMode int        `json:"lessonChargingMode"`
	LessonScopeModel   int        `json:"lessonScopeModel"`
	ProductName        string     `json:"productName"`
	Status             int        `json:"status"`
	EnableExpireTime   bool       `json:"enableExpireTime"`
	LastSuspendedTime  time.Time  `json:"lastSuspendedTime"`
	ExpireTime         time.Time  `json:"expireTime"`
	StudentID          string     `json:"studentId"`
	LessonID           string     `json:"lessonId"`
	LessonType         int        `json:"lessonType"`
	ChangeStatusTime   time.Time  `json:"changeStatusTime"`
	SuspendedTime      *time.Time `json:"suspendedTime"`
	ClassEndingTime    *time.Time `json:"classEndingTime"`
	AssignedClass      bool       `json:"assignedClass"`
}

type OneToOnePropertyVO struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type OneToOneItemVO struct {
	ID                         string                   `json:"id"`
	Name                       string                   `json:"name"`
	StudentName                string                   `json:"studentName"`
	StudentID                  string                   `json:"studentId"`
	Sex                        int                      `json:"sex"`
	Avatar                     string                   `json:"avatar"`
	Phone                      string                   `json:"phone"`
	SchoolID                   string                   `json:"schoolId"`
	One2OneLessonTimes         []OneToOneLessonTimeVO   `json:"one2OneLessonTimes"`
	IsScheduled                bool                     `json:"isScheduled"`
	Status                     int                      `json:"status"`
	ClassStudentStatus         int                      `json:"classStudentStatus"`
	One2OneLessonDayInfo       OneToOneLessonDayInfoVO  `json:"one2OneLessonDayInfo"`
	CreatedTime                time.Time                `json:"createdTime"`
	ClassRoomID                string                   `json:"classRoomId"`
	ClassRoomName              string                   `json:"classRoomName"`
	ClassroomEnabled           *bool                    `json:"classroomEnabled"`
	ClassTime                  float64                  `json:"classTime"`
	StudentClassTime           float64                  `json:"studentClassTime"`
	TeacherClassTime           float64                  `json:"teacherClassTime"`
	LessonID                   string                   `json:"lessonId"`
	LessonName                 string                   `json:"lessonName"`
	TuitionAccountID           string                   `json:"tuitionAccountId"`
	DefaultTeacherID           string                   `json:"defaultTeacherId"`
	DefaultTeacherName         string                   `json:"defaultTeacherName"`
	DefaultClassTimeRecordMode int                      `json:"defaultClassTimeRecordMode"`
	IsGradeUpgrade             bool                     `json:"isGradeUpgrade"`
	LastFinishedLessonDay      time.Time                `json:"lastFinishedLessonDay"`
	TeacherList                []OneToOneTeacherVO      `json:"teacherList"`
	TuitionAccount             OneToOneTuitionAccountVO `json:"tuitionAccount"`
	ClassProperties            []OneToOnePropertyVO     `json:"classProperties"`
	ClassTeacherID             string                   `json:"classTeacherId,omitempty"`
	ClassTeacherName           string                   `json:"classTeacherName,omitempty"`
	Remark                     string                   `json:"remark,omitempty"`
}

type OneToOneListResultVO struct {
	Total        int              `json:"total"`
	StudentCount int              `json:"studentCount"`
	List         []OneToOneItemVO `json:"list"`
}

type OneToOneDetailVO struct {
	ID                         string                   `json:"id"`
	StudentID                  string                   `json:"studentId"`
	SchoolID                   string                   `json:"schoolId"`
	Name                       string                   `json:"name"`
	StudentName                string                   `json:"studentName"`
	StudentAvatar              string                   `json:"studentAvatar"`
	StudentGender              int                      `json:"studentGender"`
	LessonID                   string                   `json:"lessonId"`
	LessonName                 string                   `json:"lessonName"`
	LessonPrice                float64                  `json:"lessonPrice"`
	ClassroomID                string                   `json:"classroomId"`
	ClassroomName              *string                  `json:"classroomName"`
	TuitionAccountID           string                   `json:"tuitionAccountId"`
	ClassTime                  float64                  `json:"classTime"`
	IsScheduled                bool                     `json:"isScheduled"`
	ClassroomEnabled           *bool                    `json:"classroomEnabled"`
	Status                     int                      `json:"status"`
	ClassStudentStatus         int                      `json:"classStudentStatus"`
	CreatedTime                time.Time                `json:"createdTime"`
	DefaultStudentClassTime    float64                  `json:"defaultStudentClassTime"`
	DefaultTeacherClassTime    float64                  `json:"defaultTeacherClassTime"`
	DefaultClassTimeRecordMode int                      `json:"defaultClassTimeRecordMode"`
	DefaultTeacherID           string                   `json:"defaultTeacherId"`
	DefaultTeacherName         string                   `json:"defaultTeacherName"`
	IsGradeUpgrade             bool                     `json:"isGradeUpgrade"`
	Remark                     string                   `json:"remark"`
	TeacherList                []OneToOneTeacherVO      `json:"teacherList"`
	TuitionAccount             OneToOneTuitionAccountVO `json:"tuitionAccount"`
	CreatedStaffID             string                   `json:"createdStaffId"`
	CreatedStaffName           string                   `json:"createdStaffName"`
	ClassProperties            []OneToOnePropertyVO     `json:"classProperties"`
	DefaultTeacherStatus       int                      `json:"defaultTeacherStatus"`
}

type OneToOneBatchAssignTeacherDTO struct {
	IDs              []string `json:"ids"`
	ClassTeacherID   string   `json:"classTeacherId"`   // 兼容旧版单选
	ClassTeacherIDs  []string `json:"classTeacherIds"`  // 多选班主任
}

type OneToOneBatchClassTimeDTO struct {
	IDs                  []string `json:"ids"`
	ClassTime            float64  `json:"classTime"`
	StudentClassTime     float64  `json:"studentClassTime"`
	TeacherClassTime     float64  `json:"teacherClassTime"`
	ClassTimeRecordMode  int      `json:"classTimeRecordMode"` // 1 按固定课时 2 按上课时长
}

type OneToOneCheckNameDTO struct {
	Name      string `json:"name"`
	ExceptID  string `json:"exceptId"`
	IsOne2One bool   `json:"isOne2One"`
}

type OneToOneUpdateDTO struct {
	ID                         string               `json:"id"`
	StudentID                  string               `json:"studentId"`
	LessonID                   string               `json:"lessonId"`
	Name                       string               `json:"name"`
	TeacherID                  []string             `json:"teacherId"`
	DefaultTeacherID           string               `json:"defaultTeacherId"`
	DefaultStudentClassTime    float64              `json:"defaultStudentClassTime"`
	DefaultTeacherClassTime    float64              `json:"defaultTeacherClassTime"`
	DefaultClassTimeRecordMode int                  `json:"defaultClassTimeRecordMode"`
	Remark                     string               `json:"remark"`
	ClassProperties            []OneToOnePropertyVO `json:"classProperties"`
}

// StudentLessonTuitionAccountsQueryDTO 按学员+课程查询学费账户（对齐竞品 GetStudentAllTuitionAccountByLessonId）
type StudentLessonTuitionAccountsQueryDTO struct {
	StudentID string `json:"studentId"`
	LessonID  string `json:"lessonId"`
}

// StudentLessonTuitionAccountItem 单条学费账户（含竞品常用字段名 quantity/tuition 表示剩余）
type StudentLessonTuitionAccountItem struct {
	ID                     string     `json:"id"`
	StudentID              string     `json:"studentId"`
	LessonID               string     `json:"lessonId"`
	ProductName            string     `json:"productName"`
	LessonChargingMode     int        `json:"lessonChargingMode"`
	TotalQuantity          float64    `json:"totalQuantity"`
	TotalFreeQuantity      float64    `json:"totalFreeQuantity"`
	TotalTuition           float64    `json:"totalTuition"`
	FreeQuantity           float64    `json:"freeQuantity"`
	Quantity               float64    `json:"quantity"`
	Tuition                float64    `json:"tuition"`
	Suspended              bool       `json:"suspended"`
	SuspendedTime          *time.Time `json:"suspendedTime,omitempty"`
	StartTime              time.Time  `json:"startTime"`
	EnableExpireTime       bool       `json:"enableExpireTime"`
	ExpireTime             time.Time  `json:"expireTime"`
	AssignedClass          bool       `json:"assignedClass"`
	LessonScope            int        `json:"lessonScope"`
	GeneralLessonIDList    []string   `json:"generalLessonIdList"`
	LatestStartTime        time.Time  `json:"latestStartTime"`
	LessonType             int        `json:"lessonType"`
	IsTuitionAccountActive bool       `json:"isTuitionAccountActive"`
	Status                 int        `json:"status"`
}

type StudentLessonTuitionAccountsResult struct {
	List []StudentLessonTuitionAccountItem `json:"list"`
}
