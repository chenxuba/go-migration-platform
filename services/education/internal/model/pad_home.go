package model

type PadHomeSummaryVO struct {
	Date            string                 `json:"date"`
	Weekday         string                 `json:"weekday"`
	AssessmentStats PadHomeAssessmentStats `json:"assessmentStats"`
	Schedule        []PadHomeScheduleItem  `json:"schedule"`
	Weather         PadHomeWeather         `json:"weather"`
}

type PadHomeAssessmentStats struct {
	EnrolledStudents   int     `json:"enrolledStudents"`
	AssessedStudents   int     `json:"assessedStudents"`
	InProgressDrafts   int     `json:"inProgressDrafts"`
	UnassessedStudents int     `json:"unassessedStudents"`
	CompletedRecords   int     `json:"completedRecords"`
	PendingIEP         int     `json:"pendingIep"`
	DraftIEP           int     `json:"draftIep"`
	GeneratedIEP       int     `json:"generatedIep"`
	Total              int     `json:"total"`
	CoverageRate       float64 `json:"coverageRate"`
	CompletionRate     float64 `json:"completionRate"`
}

type PadHomeScheduleItem struct {
	Time  string `json:"time"`
	Title string `json:"title"`
	Place string `json:"place"`
	State string `json:"state"`
}

type PadHomeWeather struct {
	City        string  `json:"city"`
	Condition   string  `json:"condition"`
	DisplayName string  `json:"displayName"`
	Temperature float64 `json:"temperature,omitempty"`
	UpdatedAt   string  `json:"updatedAt,omitempty"`
	Source      string  `json:"source,omitempty"`
}
