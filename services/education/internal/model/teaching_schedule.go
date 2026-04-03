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
