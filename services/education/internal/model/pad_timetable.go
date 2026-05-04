package model

type PadTimetableQueryDTO struct {
	StartDate       string `json:"startDate"`
	EndDate         string `json:"endDate"`
	TeacherID       int64  `json:"teacherId,omitempty"`
	PeriodGroupUUID string `json:"periodGroupUuid,omitempty"`
}

type PadTimetableVO struct {
	StartDate               string                    `json:"startDate"`
	EndDate                 string                    `json:"endDate"`
	SelectedPeriodGroupUUID string                    `json:"selectedPeriodGroupUuid"`
	SelectedTeacherID       int64                     `json:"selectedTeacherId"`
	SelectedTeacherName     string                    `json:"selectedTeacherName"`
	PeriodGroups            []PadTimetablePeriodGroup `json:"periodGroups"`
	Teachers                []PadTimetableTeacher     `json:"teachers"`
	Days                    []PadTimetableDay         `json:"days"`
	Slots                   []PadTimetableSlot        `json:"slots"`
	Items                   []PadTimetableItem        `json:"items"`
	Summary                 PadTimetableSummary       `json:"summary"`
}

type PadTimetablePeriodGroup struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Sort        int     `json:"sort"`
	StartTime   string  `json:"startTime"`
	EndTime     string  `json:"endTime"`
	LessonCount int     `json:"lessonCount"`
	TeacherIDs  []int64 `json:"teacherIds,omitempty"`
}

type PadTimetableTeacher struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Current bool   `json:"current,omitempty"`
}

type PadTimetableDay struct {
	Date    string `json:"date"`
	Label   string `json:"label"`
	Weekday string `json:"weekday"`
}

type PadTimetableSlot struct {
	Title     string `json:"title"`
	Time      string `json:"time"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

type PadTimetableItem struct {
	ID                string `json:"id"`
	Date              string `json:"date"`
	StartTime         string `json:"startTime"`
	EndTime           string `json:"endTime"`
	LessonName        string `json:"lessonName"`
	TeachingClassName string `json:"teachingClassName,omitempty"`
	StudentName       string `json:"studentName,omitempty"`
	PersonName        string `json:"personName"`
	ClassroomName     string `json:"classroomName,omitempty"`
	TeacherID         string `json:"teacherId"`
	TeacherName       string `json:"teacherName"`
	Status            string `json:"status"`
	StatusText        string `json:"statusText"`
	Conflict          bool   `json:"conflict"`
}

type PadTimetableSummary struct {
	Total    int `json:"total"`
	Unsigned int `json:"unsigned"`
	Signed   int `json:"signed"`
	Partial  int `json:"partial"`
	Trial    int `json:"trial"`
	Conflict int `json:"conflict"`
}
