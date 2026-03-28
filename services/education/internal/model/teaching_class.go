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
	ID                    string                   `json:"id"`
	Name                  string                   `json:"name"`
	StudentName           string                   `json:"studentName"`
	StudentID             string                   `json:"studentId"`
	Sex                   int                      `json:"sex"`
	Avatar                string                   `json:"avatar"`
	Phone                 string                   `json:"phone"`
	SchoolID              string                   `json:"schoolId"`
	One2OneLessonTimes    []OneToOneLessonTimeVO   `json:"one2OneLessonTimes"`
	IsScheduled           bool                     `json:"isScheduled"`
	Status                int                      `json:"status"`
	ClassStudentStatus    int                      `json:"classStudentStatus"`
	One2OneLessonDayInfo  OneToOneLessonDayInfoVO  `json:"one2OneLessonDayInfo"`
	CreatedTime           time.Time                `json:"createdTime"`
	ClassRoomID           string                   `json:"classRoomId"`
	ClassRoomName         string                   `json:"classRoomName"`
	ClassroomEnabled      *bool                    `json:"classroomEnabled"`
	ClassTime             float64                  `json:"classTime"`
	StudentClassTime      float64                  `json:"studentClassTime"`
	TeacherClassTime      float64                  `json:"teacherClassTime"`
	LessonID              string                   `json:"lessonId"`
	LessonName            string                   `json:"lessonName"`
	TuitionAccountID      string                   `json:"tuitionAccountId"`
	DefaultTeacherID      string                   `json:"defaultTeacherId"`
	DefaultTeacherName    string                   `json:"defaultTeacherName"`
	IsGradeUpgrade        bool                     `json:"isGradeUpgrade"`
	LastFinishedLessonDay time.Time                `json:"lastFinishedLessonDay"`
	TeacherList           []OneToOneTeacherVO      `json:"teacherList"`
	TuitionAccount        OneToOneTuitionAccountVO `json:"tuitionAccount"`
	ClassProperties       []OneToOnePropertyVO     `json:"classProperties"`
	ClassTeacherID        string                   `json:"classTeacherId,omitempty"`
	ClassTeacherName      string                   `json:"classTeacherName,omitempty"`
}

type OneToOneListResultVO struct {
	Total        int              `json:"total"`
	StudentCount int              `json:"studentCount"`
	List         []OneToOneItemVO `json:"list"`
}

type OneToOneBatchAssignTeacherDTO struct {
	IDs            []string `json:"ids"`
	ClassTeacherID string   `json:"classTeacherId"`
}

type OneToOneBatchClassTimeDTO struct {
	IDs              []string `json:"ids"`
	ClassTime        float64  `json:"classTime"`
	StudentClassTime float64  `json:"studentClassTime"`
	TeacherClassTime float64  `json:"teacherClassTime"`
}

type OneToOneBatchAttributeDTO struct {
	IDs                []string `json:"ids"`
	DefaultTeacherID   string   `json:"defaultTeacherId"`
	Status             *int     `json:"status"`
	ClassStudentStatus *int     `json:"classStudentStatus"`
}
