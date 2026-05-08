package erxinscore

import "time"

const (
	DomainGrossMotor       = "GM"
	DomainFineMotor        = "FM"
	DomainAdaptiveAbility  = "AD"
	DomainLanguage         = "LANG"
	DomainSocialBehavior   = "SOC"
	ScaleCode              = "ERXIN2"
	DefaultScaleVersion    = "WS-T-580-2017"
	MaxSupportedAgeMonths  = 72
	StandardAgeMonthCount  = 28
	ExpectedItemDefinition = 261
)

var StandardAgeMonths = []int{
	1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
	15, 18, 21, 24, 27, 30, 33, 36,
	42, 48, 54, 60, 66, 72, 78, 84,
}

var DomainOrder = []string{
	DomainGrossMotor,
	DomainFineMotor,
	DomainAdaptiveAbility,
	DomainLanguage,
	DomainSocialBehavior,
}

type ItemDefinition struct {
	ItemNo                int     `json:"item_no"`
	ItemTitle             string  `json:"item_title"`
	TestItem              string  `json:"test_item"`
	AgeMonth              int     `json:"age_month"`
	AgeSegment            string  `json:"age_segment"`
	DomainCode            string  `json:"domain_code"`
	DomainName            string  `json:"domain_name"`
	ParentReportAllowed   bool    `json:"parent_report_allowed"`
	AttentionIfFailed     bool    `json:"attention_if_failed"`
	DomainMonthTotalScore float64 `json:"domain_month_total_score"`
	ItemWeight            float64 `json:"item_weight"`
	Method                string  `json:"method"`
	PassCriteria          string  `json:"pass_criteria"`
	SourcePDF             string  `json:"source_pdf"`
	SourcePages           []int   `json:"source_pages"`
	OCRStatus             string  `json:"ocr_status"`
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
	ItemPasses     map[int]bool
}

type AssessmentWindow struct {
	MainAgeMonth int              `json:"mainAgeMonth"`
	AgeMonths    []int            `json:"ageMonths"`
	ItemNumbers  []int            `json:"itemNumbers"`
	DomainItems  map[string][]int `json:"domainItems"`
}

type AssessmentResult struct {
	Age                     Age              `json:"age"`
	MainAgeMonth            int              `json:"mainAgeMonth"`
	Window                  AssessmentWindow `json:"window"`
	Domains                 []DomainResult   `json:"domains"`
	MeanMentalAgeMonths     float64          `json:"meanMentalAgeMonths"`
	MeanMentalAgeMonthsText string           `json:"meanMentalAgeMonthsText"`
	DQ                      float64          `json:"dq"`
	Level                   string           `json:"level"`
	Complete                bool             `json:"complete"`
	Warnings                []string         `json:"warnings,omitempty"`
}

type DomainResult struct {
	DomainCode               string           `json:"domainCode"`
	DomainName               string           `json:"domainName"`
	MentalAgeMonths          float64          `json:"mentalAgeMonths"`
	MentalAgeMonthsText      string           `json:"mentalAgeMonthsText"`
	DQ                       float64          `json:"dq"`
	Level                    string           `json:"level"`
	BasalAgeMonth            int              `json:"basalAgeMonth,omitempty"`
	CeilingAgeMonth          int              `json:"ceilingAgeMonth,omitempty"`
	BasalComplete            bool             `json:"basalComplete"`
	CeilingComplete          bool             `json:"ceilingComplete"`
	Complete                 bool             `json:"complete"`
	AgeMonthResults          []AgeMonthResult `json:"ageMonthResults,omitempty"`
	PassedItemNumbers        []int            `json:"passedItemNumbers,omitempty"`
	FailedItemNumbers        []int            `json:"failedItemNumbers,omitempty"`
	MissingItemNumbers       []int            `json:"missingItemNumbers,omitempty"`
	DefaultPassedItemNumbers []int            `json:"defaultPassedItemNumbers,omitempty"`
	Warnings                 []string         `json:"warnings,omitempty"`
}

type AgeMonthResult struct {
	AgeMonth           int   `json:"ageMonth"`
	ItemNumbers        []int `json:"itemNumbers"`
	PassedItemNumbers  []int `json:"passedItemNumbers,omitempty"`
	FailedItemNumbers  []int `json:"failedItemNumbers,omitempty"`
	MissingItemNumbers []int `json:"missingItemNumbers,omitempty"`
	AllPassed          bool  `json:"allPassed"`
	AllFailed          bool  `json:"allFailed"`
	Complete           bool  `json:"complete"`
}
