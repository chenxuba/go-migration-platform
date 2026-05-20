package vbmappscore

const (
	ScaleCode           = "VBMAPP"
	DefaultScaleVersion = "VBMAPP_CN_2ND_DRAFT_2026_05"

	ModuleMilestones = "milestones"
	ModuleBarriers   = "barriers"
	ModuleTransition = "transition"
)

type DomainDefinition struct {
	DomainCode string                  `json:"domainCode"`
	DomainName string                  `json:"domainName"`
	SortNo     int                     `json:"sortNo"`
	ItemCount  int                     `json:"itemCount"`
	MaxScore   float64                 `json:"maxScore"`
	Levels     []DomainLevelDefinition `json:"levels,omitempty"`
}

type DomainLevelDefinition struct {
	Level        int    `json:"level"`
	AgeBand      string `json:"ageBand,omitempty"`
	ItemCount    int    `json:"itemCount"`
	MilestoneNos []int  `json:"milestoneNos,omitempty"`
}

type MilestoneItemDefinition struct {
	SequenceNo       int    `json:"sequenceNo"`
	MilestoneID      string `json:"milestoneId"`
	Label            string `json:"label,omitempty"`
	DomainCode       string `json:"domainCode"`
	DomainName       string `json:"domainName,omitempty"`
	DomainSortNo     int    `json:"domainSortNo,omitempty"`
	Level            int    `json:"level"`
	AgeBand          string `json:"ageBand,omitempty"`
	MilestoneNo      int    `json:"milestoneNo"`
	Title            string `json:"title"`
	AssessmentMode   string `json:"assessmentMode,omitempty"`
	ScoreType        string `json:"scoreType,omitempty"`
	SourceFile       string `json:"sourceFile,omitempty"`
	SourceTableIndex int    `json:"sourceTableIndex,omitempty"`
}

type MilestoneScoringRule struct {
	MilestoneID       string `json:"milestoneId"`
	DomainCode        string `json:"domainCode"`
	DomainName        string `json:"domainName,omitempty"`
	Level             int    `json:"level"`
	MilestoneNo       int    `json:"milestoneNo"`
	Description       string `json:"description,omitempty"`
	OnePointCriteria  string `json:"onePointCriteria,omitempty"`
	HalfPointCriteria string `json:"halfPointCriteria,omitempty"`
	SourceFile        string `json:"sourceFile,omitempty"`
}

type ScoreOption struct {
	Score       int    `json:"score"`
	Description string `json:"description,omitempty"`
}

type BarrierDefinition struct {
	BarrierNo    int           `json:"barrierNo"`
	BarrierCode  string        `json:"barrierCode"`
	BarrierName  string        `json:"barrierName"`
	MinScore     int           `json:"minScore"`
	MaxScore     int           `json:"maxScore"`
	ScoreOptions []ScoreOption `json:"scoreOptions,omitempty"`
	SourceFile   string        `json:"sourceFile,omitempty"`
}

type TransitionDefinition struct {
	TransitionNo             int           `json:"transitionNo"`
	TransitionCode           string        `json:"transitionCode"`
	TransitionName           string        `json:"transitionName"`
	Category                 string        `json:"category,omitempty"`
	MinScore                 int           `json:"minScore"`
	MaxScore                 int           `json:"maxScore"`
	ScoreOptions             []ScoreOption `json:"scoreOptions,omitempty"`
	PlacementRecommendations []string      `json:"placementRecommendations,omitempty"`
	SourceFile               string        `json:"sourceFile,omitempty"`
}

type AssessmentInput struct {
	ScaleVersion string `json:"scaleVersion,omitempty"`

	MilestoneScores  map[string]float64 `json:"milestoneScores,omitempty"`
	BarrierScores    map[string]int     `json:"barrierScores,omitempty"`
	TransitionScores map[string]int     `json:"transitionScores,omitempty"`

	PreviousMilestoneScores  map[string]float64 `json:"previousMilestoneScores,omitempty"`
	PreviousBarrierScores    map[string]int     `json:"previousBarrierScores,omitempty"`
	PreviousTransitionScores map[string]int     `json:"previousTransitionScores,omitempty"`
}

type AssessmentResult struct {
	ScaleCode     string `json:"scaleCode"`
	ScaleVersion  string `json:"scaleVersion"`
	Complete      bool   `json:"complete"`
	CurrentModule string `json:"currentModule,omitempty"`

	ModuleProgress []ModuleProgressResult `json:"moduleProgress"`
	Milestones     MilestoneModuleResult  `json:"milestones"`
	Barriers       BarrierModuleResult    `json:"barriers"`
	Transition     TransitionModuleResult `json:"transition"`
	Warnings       []string               `json:"warnings,omitempty"`
}

type ModuleProgressResult struct {
	ModuleCode    string  `json:"moduleCode"`
	ModuleName    string  `json:"moduleName"`
	AnsweredItems int     `json:"answeredItems"`
	ItemCount     int     `json:"itemCount"`
	MissingItems  int     `json:"missingItems"`
	Score         float64 `json:"score"`
	MaxScore      float64 `json:"maxScore"`
	Percent       float64 `json:"percent"`
	Complete      bool    `json:"complete"`
}

type MilestoneModuleResult struct {
	TotalScore          float64                `json:"totalScore"`
	MaxScore            float64                `json:"maxScore"`
	AnsweredItems       int                    `json:"answeredItems"`
	ItemCount           int                    `json:"itemCount"`
	MissingMilestoneIDs []string               `json:"missingMilestoneIds,omitempty"`
	Percent             float64                `json:"percent"`
	Complete            bool                   `json:"complete"`
	Levels              []LevelScoreResult     `json:"levels"`
	Domains             []DomainScoreResult    `json:"domains"`
	Items               []MilestoneScoreResult `json:"items,omitempty"`
	LowItems            []MilestoneScoreResult `json:"lowItems,omitempty"`
}

type LevelScoreResult struct {
	Level               int      `json:"level"`
	AgeBand             string   `json:"ageBand,omitempty"`
	TotalScore          float64  `json:"totalScore"`
	MaxScore            float64  `json:"maxScore"`
	AnsweredItems       int      `json:"answeredItems"`
	ItemCount           int      `json:"itemCount"`
	MissingMilestoneIDs []string `json:"missingMilestoneIds,omitempty"`
	Percent             float64  `json:"percent"`
	Complete            bool     `json:"complete"`
}

type DomainScoreResult struct {
	DomainCode          string             `json:"domainCode"`
	DomainName          string             `json:"domainName"`
	SortNo              int                `json:"sortNo"`
	TotalScore          float64            `json:"totalScore"`
	MaxScore            float64            `json:"maxScore"`
	AnsweredItems       int                `json:"answeredItems"`
	ItemCount           int                `json:"itemCount"`
	MissingMilestoneIDs []string           `json:"missingMilestoneIds,omitempty"`
	Percent             float64            `json:"percent"`
	Complete            bool               `json:"complete"`
	Levels              []LevelScoreResult `json:"levels,omitempty"`
}

type MilestoneScoreResult struct {
	MilestoneID       string   `json:"milestoneId"`
	Label             string   `json:"label,omitempty"`
	Title             string   `json:"title"`
	DomainCode        string   `json:"domainCode"`
	DomainName        string   `json:"domainName,omitempty"`
	Level             int      `json:"level"`
	AgeBand           string   `json:"ageBand,omitempty"`
	MilestoneNo       int      `json:"milestoneNo"`
	AssessmentMode    string   `json:"assessmentMode,omitempty"`
	Score             *float64 `json:"score,omitempty"`
	PreviousScore     *float64 `json:"previousScore,omitempty"`
	Change            *float64 `json:"change,omitempty"`
	Status            string   `json:"status"`
	OnePointCriteria  string   `json:"onePointCriteria,omitempty"`
	HalfPointCriteria string   `json:"halfPointCriteria,omitempty"`
}

type BarrierModuleResult struct {
	TotalScore          int                  `json:"totalScore"`
	MaxScore            int                  `json:"maxScore"`
	AnsweredItems       int                  `json:"answeredItems"`
	ItemCount           int                  `json:"itemCount"`
	MissingBarrierCodes []string             `json:"missingBarrierCodes,omitempty"`
	Percent             float64              `json:"percent"`
	Complete            bool                 `json:"complete"`
	Items               []BarrierScoreResult `json:"items,omitempty"`
	AttentionItems      []BarrierScoreResult `json:"attentionItems,omitempty"`
	HighRiskItems       []BarrierScoreResult `json:"highRiskItems,omitempty"`
}

type BarrierScoreResult struct {
	BarrierCode   string `json:"barrierCode"`
	BarrierName   string `json:"barrierName"`
	Score         *int   `json:"score,omitempty"`
	PreviousScore *int   `json:"previousScore,omitempty"`
	Change        *int   `json:"change,omitempty"`
	Severity      string `json:"severity"`
	Scored        bool   `json:"scored"`
}

type TransitionModuleResult struct {
	TotalScore             int                     `json:"totalScore"`
	MaxScore               int                     `json:"maxScore"`
	AnsweredItems          int                     `json:"answeredItems"`
	ItemCount              int                     `json:"itemCount"`
	MissingTransitionCodes []string                `json:"missingTransitionCodes,omitempty"`
	Percent                float64                 `json:"percent"`
	Complete               bool                    `json:"complete"`
	Items                  []TransitionScoreResult `json:"items,omitempty"`
	Suggestions            []TransitionSuggestion  `json:"suggestions,omitempty"`
}

type TransitionScoreResult struct {
	TransitionCode           string   `json:"transitionCode"`
	TransitionName           string   `json:"transitionName"`
	Category                 string   `json:"category,omitempty"`
	Score                    *int     `json:"score,omitempty"`
	SuggestedScore           *int     `json:"suggestedScore,omitempty"`
	PreviousScore            *int     `json:"previousScore,omitempty"`
	Change                   *int     `json:"change,omitempty"`
	Scored                   bool     `json:"scored"`
	PlacementRecommendations []string `json:"placementRecommendations,omitempty"`
}

type TransitionSuggestion struct {
	TransitionCode string `json:"transitionCode"`
	TransitionName string `json:"transitionName"`
	Score          int    `json:"score"`
	Basis          string `json:"basis"`
}

type ResponseFieldTemplate struct {
	Label         string   `json:"label"`
	Control       string   `json:"control"`
	Required      bool     `json:"required"`
	RecordPurpose string   `json:"recordPurpose,omitempty"`
	Columns       []string `json:"columns,omitempty"`
	Options       []string `json:"options,omitempty"`
}

type ResponseMaterialProfile struct {
	Label                string                       `json:"label"`
	SourceLogic          string                       `json:"sourceLogic,omitempty"`
	SuggestedTypes       []string                     `json:"suggestedTypes,omitempty"`
	RecommendedMaterials []ResponseMaterialSuggestion `json:"recommendedMaterials,omitempty"`
	PreparationChecks    []string                     `json:"preparationChecks,omitempty"`
}

type ResponseMaterialSuggestion struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
}

type ResponseSchemaScoreEvidence struct {
	PrimaryMetric            string            `json:"primaryMetric,omitempty"`
	OnePointCriteria         string            `json:"onePointCriteria,omitempty"`
	HalfPointCriteria        string            `json:"halfPointCriteria,omitempty"`
	RawCriteria              map[string]string `json:"rawCriteria,omitempty"`
	ScoreRange               []int             `json:"scoreRange,omitempty"`
	ScoreOptions             []ScoreOption     `json:"scoreOptions,omitempty"`
	PlacementRecommendations []string          `json:"placementRecommendations,omitempty"`
	RequiresTeacherConfirm   bool              `json:"requiresTeacherConfirmation"`
}

type ResponseSchemaAutoCompletion struct {
	CanAutoCompleteDraft   bool     `json:"canAutoCompleteDraft"`
	CanSuggestScore        bool     `json:"canSuggestScore"`
	ScoreStrategy          string   `json:"scoreStrategy"`
	PrimaryMetric          string   `json:"primaryMetric,omitempty"`
	ComputedIndicators     []string `json:"computedIndicators,omitempty"`
	RequiresTeacherConfirm bool     `json:"requiresTeacherConfirmation"`
	TeacherCanOverride     bool     `json:"teacherCanOverride"`
}

type ResponseSchemaRelatedItem struct {
	Relation    string `json:"relation"`
	MilestoneID string `json:"milestoneId"`
	Reason      string `json:"reason,omitempty"`
}

type ResponseSchemaItemDesign struct {
	RecordingInstructions []string `json:"recordingInstructions,omitempty"`
	UIEmphasis            []string `json:"uiEmphasis,omitempty"`
}

type MilestoneResponseSchema struct {
	ModuleCode        string                       `json:"moduleCode"`
	MilestoneID       string                       `json:"milestoneId"`
	Label             string                       `json:"label,omitempty"`
	DomainCode        string                       `json:"domainCode"`
	DomainName        string                       `json:"domainName,omitempty"`
	Level             int                          `json:"level"`
	AgeBand           string                       `json:"ageBand,omitempty"`
	MilestoneNo       int                          `json:"milestoneNo"`
	AssessmentMode    string                       `json:"assessmentMode,omitempty"`
	RecordDepth       string                       `json:"recordDepth"`
	RecordType        string                       `json:"recordType"`
	UIPattern         string                       `json:"uiPattern"`
	MaterialProfileID string                       `json:"materialProfileId,omitempty"`
	WhyRecord         string                       `json:"whyRecord,omitempty"`
	EvidenceTargets   []string                     `json:"evidenceTargets,omitempty"`
	FieldTemplateIDs  []string                     `json:"fieldTemplateIds,omitempty"`
	QualityChecks     []string                     `json:"qualityChecks,omitempty"`
	ScoreEvidence     ResponseSchemaScoreEvidence  `json:"scoreEvidence"`
	AutoCompletion    ResponseSchemaAutoCompletion `json:"autoCompletion"`
	RelatedItems      []ResponseSchemaRelatedItem  `json:"relatedItems,omitempty"`
	ItemDesign        ResponseSchemaItemDesign     `json:"itemDesign,omitempty"`
	SourceFiles       []string                     `json:"sourceFiles,omitempty"`
}

type BarrierResponseSchema struct {
	ModuleCode       string                       `json:"moduleCode"`
	BarrierCode      string                       `json:"barrierCode"`
	BarrierName      string                       `json:"barrierName"`
	BarrierNo        int                          `json:"barrierNo"`
	RecordDepth      string                       `json:"recordDepth"`
	UIPattern        string                       `json:"uiPattern"`
	FieldTemplateIDs []string                     `json:"fieldTemplateIds,omitempty"`
	EvidenceTargets  []string                     `json:"evidenceTargets,omitempty"`
	ScoreEvidence    ResponseSchemaScoreEvidence  `json:"scoreEvidence"`
	AutoCompletion   ResponseSchemaAutoCompletion `json:"autoCompletion"`
	SourceFiles      []string                     `json:"sourceFiles,omitempty"`
}

type TransitionResponseSchema struct {
	ModuleCode       string                       `json:"moduleCode"`
	TransitionCode   string                       `json:"transitionCode"`
	TransitionName   string                       `json:"transitionName"`
	TransitionNo     int                          `json:"transitionNo"`
	Category         string                       `json:"category,omitempty"`
	RecordDepth      string                       `json:"recordDepth"`
	UIPattern        string                       `json:"uiPattern"`
	FieldTemplateIDs []string                     `json:"fieldTemplateIds,omitempty"`
	EvidenceTargets  []string                     `json:"evidenceTargets,omitempty"`
	ScoreEvidence    ResponseSchemaScoreEvidence  `json:"scoreEvidence"`
	AutoCompletion   ResponseSchemaAutoCompletion `json:"autoCompletion"`
	SourceFiles      []string                     `json:"sourceFiles,omitempty"`
}
