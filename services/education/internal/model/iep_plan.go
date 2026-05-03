package model

type PEP3IEPPlanGenerateRequest struct {
	ID             int64 `json:"id"`
	DurationMonths int   `json:"durationMonths,omitempty"`
}

type PEP3IEPPlanWordExportRequest struct {
	ID             int64                `json:"id,omitempty"`
	DurationMonths int                  `json:"durationMonths,omitempty"`
	Plan           *PEP3IEPPlanAIResult `json:"plan,omitempty"`
}

type PEP3IEPPlanAIResult struct {
	Title   string             `json:"title"`
	Model   string             `json:"model,omitempty"`
	Student PEP3IEPPlanStudent `json:"student"`
	Meta    PEP3IEPPlanMeta    `json:"meta"`
	Rows    []PEP3IEPPlanRow   `json:"rows"`
}

type PEP3IEPPlanStudent struct {
	Name      string `json:"name"`
	Gender    string `json:"gender"`
	BirthDate string `json:"birthDate"`
}

type PEP3IEPPlanMeta struct {
	PlanDate    string `json:"planDate"`
	Participant string `json:"participant"`
	Implementer string `json:"implementer"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
}

type PEP3IEPPlanRow struct {
	Domain       string `json:"domain"`
	LongGoal     string `json:"longGoal"`
	ShortGoal    string `json:"shortGoal"`
	CourseForm   string `json:"courseForm"`
	StartEndDate string `json:"startEndDate"`
}
