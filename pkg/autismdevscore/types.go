package autismdevscore

import "time"

const (
	DomainSensory         = "SP"
	DomainGrossMotor      = "GM"
	DomainFineMotor       = "FM"
	DomainLanguageComm    = "LC"
	DomainCognition       = "COG"
	DomainSocial          = "SOC"
	DomainDailyLiving     = "ADL"
	DomainEmotionBehavior = "EB"

	ScoreTypePEF = "PEF"
	ScoreTypeAMS = "AMS"

	ScoreP = "P"
	ScoreE = "E"
	ScoreF = "F"
	ScoreX = "X"
	ScoreA = "A"
	ScoreM = "M"
	ScoreS = "S"

	ScaleCode              = "AUTISMDEV"
	DefaultScaleVersion    = "2010-revised-trainer"
	MaxSupportedAgeMonths  = 72
	ExpectedItemDefinition = 493
)

var DomainOrder = []string{
	DomainSensory,
	DomainGrossMotor,
	DomainFineMotor,
	DomainLanguageComm,
	DomainCognition,
	DomainSocial,
	DomainDailyLiving,
	DomainEmotionBehavior,
}

var PEFDomainOrder = []string{
	DomainSensory,
	DomainGrossMotor,
	DomainFineMotor,
	DomainLanguageComm,
	DomainCognition,
	DomainSocial,
	DomainDailyLiving,
}

type ItemDefinition struct {
	ItemNo          int      `json:"item_no"`
	DomainItemNo    int      `json:"domain_item_no"`
	ItemTitle       string   `json:"item_title"`
	TestItem        string   `json:"test_item"`
	AssessmentRange string   `json:"assessment_range"`
	Materials       string   `json:"materials"`
	AgeSegment      string   `json:"age_segment"`
	AgeMinMonth     int      `json:"age_min_month"`
	AgeMaxMonth     int      `json:"age_max_month"`
	DomainCode      string   `json:"domain_code"`
	DomainName      string   `json:"domain_name"`
	ScoreType       string   `json:"score_type"`
	AssessmentModes []string `json:"assessment_modes"`
	Method          string   `json:"method"`
	PassCriteria    string   `json:"pass_criteria"`
	SourcePDF       string   `json:"source_pdf"`
	SourcePages     []int    `json:"source_pages"`
	OCRStatus       string   `json:"ocr_status"`
}

type Age struct {
	Years              int     `json:"years"`
	Months             int     `json:"months"`
	Days               int     `json:"days"`
	TotalMonths        float64 `json:"totalMonths"`
	TotalMonthsRounded float64 `json:"totalMonthsRounded"`
}

type AssessmentInput struct {
	BirthDate      time.Time
	AssessmentDate time.Time
	ItemScores     map[int]string
}

type AssessmentResult struct {
	Age               Age                    `json:"age"`
	ItemCount         int                    `json:"itemCount"`
	AnsweredItemCount int                    `json:"answeredItemCount"`
	MissingItemCount  int                    `json:"missingItemCount"`
	Complete          bool                   `json:"complete"`
	Domains           []DomainResult         `json:"domains"`
	Development       DevelopmentSummary     `json:"development"`
	Behavior          EmotionBehaviorSummary `json:"behavior"`
	Warnings          []string               `json:"warnings,omitempty"`
}

type DomainResult struct {
	DomainCode         string   `json:"domainCode"`
	DomainName         string   `json:"domainName"`
	ScoreType          string   `json:"scoreType"`
	ItemCount          int      `json:"itemCount"`
	AnsweredItemCount  int      `json:"answeredItemCount"`
	MissingItemCount   int      `json:"missingItemCount"`
	Complete           bool     `json:"complete"`
	PCount             int      `json:"pCount,omitempty"`
	ECount             int      `json:"eCount,omitempty"`
	FCount             int      `json:"fCount,omitempty"`
	XCount             int      `json:"xCount,omitempty"`
	PECount            int      `json:"peCount,omitempty"`
	RawScore           int      `json:"rawScore,omitempty"`
	ScorableItemCount  int      `json:"scorableItemCount,omitempty"`
	ScoreRate          float64  `json:"scoreRate,omitempty"`
	ACount             int      `json:"aCount,omitempty"`
	MCount             int      `json:"mCount,omitempty"`
	SCount             int      `json:"sCount,omitempty"`
	AdaptiveCount      int      `json:"adaptiveCount,omitempty"`
	AbnormalCount      int      `json:"abnormalCount,omitempty"`
	MissingItemNumbers []int    `json:"missingItemNumbers,omitempty"`
	Warnings           []string `json:"warnings,omitempty"`
}

type DevelopmentSummary struct {
	DomainCount       int     `json:"domainCount"`
	ItemCount         int     `json:"itemCount"`
	AnsweredItemCount int     `json:"answeredItemCount"`
	PCount            int     `json:"pCount"`
	ECount            int     `json:"eCount"`
	FCount            int     `json:"fCount"`
	XCount            int     `json:"xCount"`
	PECount           int     `json:"peCount"`
	RawScore          int     `json:"rawScore"`
	ScorableItemCount int     `json:"scorableItemCount"`
	ScoreRate         float64 `json:"scoreRate"`
	Complete          bool    `json:"complete"`
}

type EmotionBehaviorSummary struct {
	ItemCount         int     `json:"itemCount"`
	AnsweredItemCount int     `json:"answeredItemCount"`
	ACount            int     `json:"aCount"`
	MCount            int     `json:"mCount"`
	SCount            int     `json:"sCount"`
	AdaptiveCount     int     `json:"adaptiveCount"`
	AbnormalCount     int     `json:"abnormalCount"`
	AdaptiveRate      float64 `json:"adaptiveRate"`
	Complete          bool    `json:"complete"`
}
