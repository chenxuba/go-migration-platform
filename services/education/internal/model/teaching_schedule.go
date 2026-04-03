package model

import "time"

const (
	TeachingScheduleStatusActive   = 1
	TeachingScheduleStatusCanceled = 2
)

type TeachingScheduleCreateSlotDTO struct {
	LessonDate string `json:"lessonDate"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
}

type CreateOneToOneSchedulesDTO struct {
	OneToOneID   string                          `json:"oneToOneId"`
	TeacherID    string                          `json:"teacherId"`
	AssistantIDs []string                        `json:"assistantIds"`
	ClassroomID  string                          `json:"classroomId"`
	Schedules    []TeachingScheduleCreateSlotDTO `json:"schedules"`
}

type TeachingScheduleListQueryDTO struct {
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	ClassType *int   `json:"classType,omitempty"`
}

type TeachingScheduleBatchUpdateDTO struct {
	BatchNo      string   `json:"batchNo"`
	IDs          []string `json:"ids"`
	TeacherID    string   `json:"teacherId"`
	AssistantIDs []string `json:"assistantIds"`
	ClassroomID  string   `json:"classroomId"`
	StartTime    string   `json:"startTime"`
	EndTime      string   `json:"endTime"`
}

type TeachingScheduleVO struct {
	ID                string    `json:"id"`
	BatchNo           string    `json:"batchNo,omitempty"`
	BatchSize         int       `json:"batchSize"`
	ClassType         int       `json:"classType"`
	TeachingClassID   string    `json:"teachingClassId"`
	TeachingClassName string    `json:"teachingClassName"`
	StudentID         string    `json:"studentId"`
	StudentName       string    `json:"studentName"`
	LessonID          string    `json:"lessonId"`
	LessonName        string    `json:"lessonName"`
	TeacherID         string    `json:"teacherId"`
	TeacherName       string    `json:"teacherName"`
	AssistantIDs      []string  `json:"assistantIds,omitempty"`
	AssistantNames    []string  `json:"assistantNames,omitempty"`
	ClassroomID       string    `json:"classroomId"`
	ClassroomName     string    `json:"classroomName"`
	LessonDate        string    `json:"lessonDate"`
	StartAt           time.Time `json:"startAt"`
	EndAt             time.Time `json:"endAt"`
	Status            int       `json:"status"`
}

type CreateOneToOneSchedulesResult struct {
	BatchNo string               `json:"batchNo,omitempty"`
	Count   int                  `json:"count"`
	List    []TeachingScheduleVO `json:"list"`
}

type TeachingScheduleValidationResult struct {
	Valid             bool                           `json:"valid"`
	Message           string                         `json:"message,omitempty"`
	CurrentSchedules  []TeachingScheduleConflictItem `json:"currentSchedules,omitempty"`
	ExistingSchedules []TeachingScheduleConflictItem `json:"existingSchedules,omitempty"`
	ConflictTypes     []string                       `json:"conflictTypes,omitempty"`
}

type TeachingScheduleConflictItem struct {
	Date          string   `json:"date"`
	Week          string   `json:"week,omitempty"`
	Name          string   `json:"name"`
	ClassTypeText string   `json:"classTypeText"`
	TimeText      string   `json:"timeText"`
	TeacherName   string   `json:"teacherName"`
	ClassroomName string   `json:"classroomName,omitempty"`
	StudentNames  []string `json:"studentNames,omitempty"`
	ConflictTypes []string `json:"conflictTypes,omitempty"`
}

type OneToOneScheduleCreateContext struct {
	ClassID            int64
	ClassName          string
	StudentID          int64
	StudentName        string
	LessonID           int64
	LessonName         string
	Status             int
	ClassStudentStatus int
}

// InstUserScheduleRosterItem 课表矩阵机构在职人员（未禁用的机构用户）
type InstUserScheduleRosterItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// TeachingScheduleMatrixDayVO 按「日期 → 教师分列」矩阵（对齐旧版 scheduleListVoList 结构）
type TeachingScheduleMatrixDayVO struct {
	ScheduleDate       string                          `json:"scheduleDate"`
	Width              int                             `json:"width"`
	ScheduleInfoVoList any                             `json:"scheduleInfoVoList"` // 旧版字段，输出 null
	ScheduleListVoList []TeachingScheduleMatrixTeacher `json:"scheduleListVoList"`
}

// TeachingScheduleMatrixTeacher 单日下单个教师列
type TeachingScheduleMatrixTeacher struct {
	TeacherName        string                         `json:"teacherName"`
	TeacherID          int64                          `json:"teacherId"`
	ScheduleInfoVoList []TeachingScheduleInfoLegacyVO `json:"scheduleInfoVoList"`
}

// ScheduleLegacyPersonVO 旧版 teacherList / studentList 元素
type ScheduleLegacyPersonVO struct {
	Name     string `json:"name"`
	ID       int64  `json:"id"`
	Type     int    `json:"type"`
	Disabled bool   `json:"disabled,omitempty"`
	UUID     any    `json:"uuid,omitempty"`
	Version  any    `json:"version,omitempty"`
}

// TeachingScheduleInfoLegacyVO 旧版日程明细（在能力范围内从 TeachingScheduleVO 映射）
type TeachingScheduleInfoLegacyVO struct {
	ID                 int64                  `json:"id"`
	UUID               string                 `json:"uuid,omitempty"`
	Version            int64                  `json:"version,omitempty"`
	CreateTime         string                 `json:"createTime,omitempty"`
	UpdateTime         *string                `json:"updateTime"`
	InstID             int64                  `json:"instId,omitempty"`
	BatchID            int64                  `json:"batchId,omitempty"`
	ModifyBatchID      int64                  `json:"modifyBatchId,omitempty"`
	CourseID           int64                  `json:"courseId,omitempty"`
	ClassID            *int64                 `json:"classId"`
	OriginID           *int64                 `json:"originId"`
	ScheduleDate       string                 `json:"scheduleDate"`
	ScheduleStartTime  string                 `json:"scheduleStartTime"`
	ScheduleEndTime    string                 `json:"scheduleEndTime"`
	ScheduleStatus     int                    `json:"scheduleStatus"`
	MissSchedule       bool                   `json:"missSchedule"`
	HasIgnore          *bool                  `json:"hasIgnore"`
	CourseStatus       int                    `json:"courseStatus"`
	Remark             *string                `json:"remark"`
	Width              int                    `json:"width"`
	TeacherList        []ScheduleLegacyPersonVO `json:"teacherList"`
	StudentList        []ScheduleLegacyPersonVO `json:"studentList"`
	CourseName         string                 `json:"courseName"`
	CourseType         int                    `json:"courseType"`
	ClassName          string                 `json:"className,omitempty"`
	CrossOver          any                    `json:"crossOver"`
	InfoVoList         any                    `json:"infoVoList"`
	LeaveList          []any                  `json:"leaveList"`
	InstName           string                 `json:"instName,omitempty"`
	CourseTime         int                    `json:"courseTime"`
	CourseHour         int                    `json:"courseHour"`
	FinishType         int                    `json:"finishType"`
}
