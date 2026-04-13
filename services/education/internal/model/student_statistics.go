package model

type StudentOverviewStatistics struct {
	TotalStudents            int `json:"totalStudents"`
	RecentMonthNewStudents   int `json:"recentMonthNewStudents"`
	PreviousMonthNewStudents int `json:"previousMonthNewStudents"`
	RecentMonthGrowthRate    int `json:"recentMonthGrowthRate"`
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
