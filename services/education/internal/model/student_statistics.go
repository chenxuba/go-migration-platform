package model

type StudentOverviewStatistics struct {
	TotalStudents            int `json:"totalStudents"`
	ReadingStudents          int `json:"readingStudents"`
	HistoryStudents          int `json:"historyStudents"`
	IntentStudents           int `json:"intentStudents"`
	PendingRenewalStudents   int `json:"pendingRenewalStudents"`
	ArrearStudents           int `json:"arrearStudents"`
	BirthdayStudents         int `json:"birthdayStudents"`
	PendingClassStudents     int `json:"pendingClassStudents"`
	PendingAttentionStudents int `json:"pendingAttentionStudents"`
	AbsentStudents           int `json:"absentStudents"`
}
