package model

import "time"

type PageRequestModel struct {
	PageIndex int `json:"pageIndex"`
	PageSize  int `json:"pageSize"`
}

type PEP3IEPItemOptionRulePageQuery struct {
	PageRequestModel PageRequestModel           `json:"pageRequestModel"`
	QueryModel       PEP3IEPItemOptionRuleQuery `json:"queryModel"`
}

type PEP3IEPItemOptionRuleQuery struct {
	ItemNo     *int   `json:"itemNo,omitempty"`
	ScoreValue *int   `json:"scoreValue,omitempty"`
	DomainCode string `json:"domainCode,omitempty"`
	Domain     string `json:"domain,omitempty"`
	Status     string `json:"status,omitempty"`
	Keyword    string `json:"keyword,omitempty"`
}

type PEP3IEPItemOptionRule struct {
	ID               int64                 `json:"id,omitempty"`
	LibraryScope     string                `json:"libraryScope"`
	InstID           int64                 `json:"instId,omitempty"`
	ItemNo           int                   `json:"itemNo"`
	ItemTitle        string                `json:"itemTitle,omitempty"`
	DomainCode       string                `json:"domainCode,omitempty"`
	Domain           string                `json:"domain,omitempty"`
	ScoreValue       int                   `json:"scoreValue"`
	ScoreLabel       string                `json:"scoreLabel,omitempty"`
	ScoreDescription string                `json:"scoreDescription,omitempty"`
	ResultMeaning    string                `json:"resultMeaning,omitempty"`
	GeneratePolicy   string                `json:"generatePolicy,omitempty"`
	Priority         int                   `json:"priority,omitempty"`
	AIInstruction    string                `json:"aiInstruction,omitempty"`
	Status           string                `json:"status,omitempty"`
	GoalMaterialIDs  []int64               `json:"goalMaterialIds,omitempty"`
	GoalMaterials    []PEP3IEPGoalMaterial `json:"goalMaterials,omitempty"`
	CreatedTime      *time.Time            `json:"createdTime,omitempty"`
	UpdatedTime      *time.Time            `json:"updatedTime,omitempty"`
}

type PEP3IEPGoalMaterialPageQuery struct {
	PageRequestModel PageRequestModel     `json:"pageRequestModel"`
	QueryModel       PEP3IEPMaterialQuery `json:"queryModel"`
}

type PEP3IEPMaterialQuery struct {
	MaterialType         string `json:"materialType,omitempty"`
	ParentGoalMaterialID *int64 `json:"parentGoalMaterialId,omitempty"`
	GoalMaterialID       *int64 `json:"goalMaterialId,omitempty"`
	DomainCode           string `json:"domainCode,omitempty"`
	Domain               string `json:"domain,omitempty"`
	CourseForm           string `json:"courseForm,omitempty"`
	Status               string `json:"status,omitempty"`
	Keyword              string `json:"keyword,omitempty"`
}

type PEP3IEPGoalMaterial struct {
	ID                    int64      `json:"id,omitempty"`
	LibraryScope          string     `json:"libraryScope"`
	InstID                int64      `json:"instId,omitempty"`
	MaterialType          string     `json:"materialType,omitempty"`
	ParentGoalMaterialID  int64      `json:"parentGoalMaterialId,omitempty"`
	DomainCode            string     `json:"domainCode,omitempty"`
	Domain                string     `json:"domain,omitempty"`
	LongGoal              string     `json:"longGoal"`
	ShortGoal             string     `json:"shortGoal"`
	CourseForm            string     `json:"courseForm,omitempty"`
	AgeMinMonths          int        `json:"ageMinMonths,omitempty"`
	AgeMaxMonths          int        `json:"ageMaxMonths,omitempty"`
	DifficultyLevel       int        `json:"difficultyLevel,omitempty"`
	ApplicableScoreValues string     `json:"applicableScoreValues,omitempty"`
	Priority              int        `json:"priority,omitempty"`
	Status                string     `json:"status,omitempty"`
	CreatedTime           *time.Time `json:"createdTime,omitempty"`
	UpdatedTime           *time.Time `json:"updatedTime,omitempty"`
}

type PEP3IEPTrainingMaterialPageQuery struct {
	PageRequestModel PageRequestModel     `json:"pageRequestModel"`
	QueryModel       PEP3IEPMaterialQuery `json:"queryModel"`
}

type PEP3IEPTrainingMaterial struct {
	ID              int64      `json:"id,omitempty"`
	LibraryScope    string     `json:"libraryScope"`
	InstID          int64      `json:"instId,omitempty"`
	GoalMaterialID  int64      `json:"goalMaterialId,omitempty"`
	TrainingProject string     `json:"trainingProject"`
	TrainingContent string     `json:"trainingContent"`
	Priority        int        `json:"priority,omitempty"`
	Status          string     `json:"status,omitempty"`
	CreatedTime     *time.Time `json:"createdTime,omitempty"`
	UpdatedTime     *time.Time `json:"updatedTime,omitempty"`
}
