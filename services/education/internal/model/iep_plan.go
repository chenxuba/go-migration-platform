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

type PEP3ExecutionPlanGenerateRequest struct {
	ID               int64                    `json:"id"`
	DurationMonths   int                      `json:"durationMonths,omitempty"`
	PlanType         string                   `json:"planType"`
	TargetMonthIndex int                      `json:"targetMonthIndex,omitempty"`
	TargetWeekIndex  int                      `json:"targetWeekIndex,omitempty"`
	SourcePlan       PEP3IEPPlanAIResult      `json:"sourcePlan"`
	MonthlyPlan      *PEP3MonthlyPlanAIResult `json:"monthlyPlan,omitempty"`
}

type PEP3ExecutionPlanWordExportRequest struct {
	ID          int64                    `json:"id,omitempty"`
	PlanType    string                   `json:"planType"`
	MonthlyPlan *PEP3MonthlyPlanAIResult `json:"monthlyPlan,omitempty"`
	WeeklyPlan  *PEP3WeeklyPlanAIResult  `json:"weeklyPlan,omitempty"`
}

type PEP3ExecutionPlanSaveRequest struct {
	ID               int64                    `json:"id"`
	DurationMonths   int                      `json:"durationMonths,omitempty"`
	PlanType         string                   `json:"planType"`
	TargetMonthIndex int                      `json:"targetMonthIndex,omitempty"`
	TargetWeekIndex  int                      `json:"targetWeekIndex,omitempty"`
	MonthlyPlan      *PEP3MonthlyPlanAIResult `json:"monthlyPlan,omitempty"`
	WeeklyPlan       *PEP3WeeklyPlanAIResult  `json:"weeklyPlan,omitempty"`
}

type PEP3ExecutionPlanSavedVO struct {
	Exists         bool                            `json:"exists"`
	DurationMonths int                             `json:"durationMonths,omitempty"`
	MonthlyPlans   []PEP3MonthlyExecutionPlanSaved `json:"monthlyPlans,omitempty"`
	WeeklyPlans    []PEP3WeeklyExecutionPlanSaved  `json:"weeklyPlans,omitempty"`
}

type PEP3MonthlyExecutionPlanSaved struct {
	TargetMonthIndex int                     `json:"targetMonthIndex"`
	Plan             PEP3MonthlyPlanAIResult `json:"plan"`
	UpdatedTime      string                  `json:"updatedTime,omitempty"`
}

type PEP3WeeklyExecutionPlanSaved struct {
	TargetMonthIndex int                    `json:"targetMonthIndex"`
	TargetWeekIndex  int                    `json:"targetWeekIndex"`
	Plan             PEP3WeeklyPlanAIResult `json:"plan"`
	UpdatedTime      string                 `json:"updatedTime,omitempty"`
}

type PEP3IEPPlanSaveRequest struct {
	ID             int64               `json:"id"`
	DurationMonths int                 `json:"durationMonths,omitempty"`
	Status         string              `json:"status,omitempty"`
	Plan           PEP3IEPPlanAIResult `json:"plan"`
}

type PEP3IEPPlanSavedVO struct {
	Exists         bool                 `json:"exists"`
	Status         string               `json:"status,omitempty"`
	DurationMonths int                  `json:"durationMonths,omitempty"`
	Plan           *PEP3IEPPlanAIResult `json:"plan,omitempty"`
	UpdatedTime    string               `json:"updatedTime,omitempty"`
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

type PEP3MonthlyPlanAIResult struct {
	Title   string               `json:"title"`
	Model   string               `json:"model,omitempty"`
	Student PEP3IEPPlanStudent   `json:"student"`
	Meta    PEP3MonthlyPlanMeta  `json:"meta"`
	Rows    []PEP3MonthlyPlanRow `json:"rows"`
}

type PEP3MonthlyPlanMeta struct {
	PlanDate    string `json:"planDate"`
	Participant string `json:"participant"`
	Implementer string `json:"implementer"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	MonthLabel  string `json:"monthLabel,omitempty"`
	SourceTitle string `json:"sourceTitle,omitempty"`
}

type PEP3MonthlyPlanRow struct {
	Domain        string                    `json:"domain"`
	LongGoal      string                    `json:"longGoal"`
	ShortGoal     string                    `json:"shortGoal"`
	TrainingItems []PEP3MonthlyTrainingItem `json:"trainingItems"`
	CourseForm    string                    `json:"courseForm"`
}

type PEP3MonthlyTrainingItem struct {
	Content      string `json:"content"`
	StartEndDate string `json:"startEndDate"`
}

type PEP3WeeklyPlanAIResult struct {
	Title        string              `json:"title"`
	Model        string              `json:"model,omitempty"`
	Student      PEP3IEPPlanStudent  `json:"student"`
	TeacherName  string              `json:"teacherName"`
	CourseName   string              `json:"courseName"`
	TrainingDate string              `json:"trainingDate"`
	Preparation  string              `json:"preparation"`
	WeekDates    []string            `json:"weekDates"`
	Rows         []PEP3WeeklyPlanRow `json:"rows"`
	SourceTitle  string              `json:"sourceTitle,omitempty"`
}

type PEP3WeeklyPlanRow struct {
	Project    string   `json:"project"`
	Content    string   `json:"content"`
	Completion []string `json:"completion,omitempty"`
}
