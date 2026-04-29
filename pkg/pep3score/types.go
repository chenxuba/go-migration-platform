package pep3score

import "time"

const (
	TableDevelopmentAge = "development_age_raw_score_range"
	TablePercentile     = "raw_score_to_percentile_rank"
	TableScaledScore    = "raw_score_to_scaled_score"
	TableComposite      = "composite_standard_score_sum_to_percentile_rank"
)

const (
	CompositeCommunication       = "communication"
	CompositeMotor               = "motor"
	CompositeMaladaptiveBehavior = "maladaptive_behavior"
)

type ItemDefinition struct {
	ItemNo    int    `json:"item_no"`
	ItemTitle string `json:"item_title,omitempty"`
	ScaleCode string `json:"domain_code"`
	ScaleName string `json:"domain,omitempty"`
}

type DomainDefinition struct {
	ScaleCode            string `json:"scale_code"`
	ScaleName            string `json:"scale_name"`
	ItemCount            *int   `json:"item_count,omitempty"`
	MaxRawScore          *int   `json:"max_raw_score,omitempty"`
	ItemNumbers          []int  `json:"item_numbers,omitempty"`
	IsDevelopmentSubtest bool   `json:"is_development_subtest,omitempty"`
	IsBehaviorSubtest    bool   `json:"is_behavior_subtest,omitempty"`
	IsCaregiverReport    bool   `json:"is_caregiver_report,omitempty"`
	CompositeCode        string `json:"composite_code,omitempty"`
}

type NormRecord struct {
	SourcePDF                 string `json:"source_pdf,omitempty"`
	Appendix                  string `json:"appendix,omitempty"`
	TableNo                   string `json:"table_no,omitempty"`
	TableType                 string `json:"table_type"`
	AgeRangeLabel             string `json:"age_range_label,omitempty"`
	AgeMinMonths              *int   `json:"age_min_months,omitempty"`
	AgeMaxMonths              *int   `json:"age_max_months,omitempty"`
	SourcePages               []int  `json:"source_pages,omitempty"`
	ScaleCode                 string `json:"scale_code,omitempty"`
	RawScore                  *int   `json:"raw_score,omitempty"`
	RawScoreMin               *int   `json:"raw_score_min,omitempty"`
	RawScoreMax               *int   `json:"raw_score_max,omitempty"`
	RawScoreRangeText         string `json:"raw_score_range_text,omitempty"`
	DevelopmentAgeMonthsLabel string `json:"development_age_months_label,omitempty"`
	DevelopmentAgeMonths      *int   `json:"development_age_months,omitempty"`
	DevelopmentAgeComparator  string `json:"development_age_comparator,omitempty"`
	CompositeCode             string `json:"composite_code,omitempty"`
	CompositeName             string `json:"composite_name,omitempty"`
	StandardScoreSum          *int   `json:"standard_score_sum,omitempty"`
	StandardScoreSumLabel     string `json:"standard_score_sum_label,omitempty"`
	ValueText                 string `json:"value_text,omitempty"`
	ValueComparator           string `json:"value_comparator,omitempty"`
	ValueNumber               *int   `json:"value_number,omitempty"`
	OCRStatus                 string `json:"ocr_status,omitempty"`
}

type AssessmentInput struct {
	BirthDate         time.Time
	AssessmentDate    time.Time
	ItemScores        map[int]int
	RawScores         map[string]int
	AllowMissingItems bool
}

type Age struct {
	Years              int `json:"years"`
	Months             int `json:"months"`
	Days               int `json:"days"`
	TotalMonthsForNorm int `json:"total_months_for_norm"`
}

type NormValue struct {
	Text        string `json:"text"`
	Comparator  string `json:"comparator,omitempty"`
	Number      *int   `json:"number,omitempty"`
	TableNo     string `json:"table_no,omitempty"`
	Appendix    string `json:"appendix,omitempty"`
	SourcePDF   string `json:"source_pdf,omitempty"`
	SourcePages []int  `json:"source_pages,omitempty"`
	OCRStatus   string `json:"ocr_status,omitempty"`
}

type ScaleResult struct {
	ScaleCode      string     `json:"scale_code"`
	ScaleName      string     `json:"scale_name,omitempty"`
	RawScore       int        `json:"raw_score"`
	MaxRawScore    *int       `json:"max_raw_score,omitempty"`
	AnsweredItems  int        `json:"answered_items,omitempty"`
	MissingItems   []int      `json:"missing_items,omitempty"`
	DevelopmentAge *NormValue `json:"development_age,omitempty"`
	PercentileRank *NormValue `json:"percentile_rank,omitempty"`
	ScaledScore    *NormValue `json:"scaled_score,omitempty"`
	Level          string     `json:"level,omitempty"`
	Warnings       []string   `json:"warnings,omitempty"`
}

type CompositeResult struct {
	CompositeCode        string     `json:"composite_code"`
	CompositeName        string     `json:"composite_name"`
	MemberScaleCodes     []string   `json:"member_scale_codes"`
	StandardScoreSum     *int       `json:"standard_score_sum,omitempty"`
	PercentileRank       *NormValue `json:"percentile_rank,omitempty"`
	Level                string     `json:"level,omitempty"`
	DevelopmentAgeMonths *float64   `json:"development_age_months,omitempty"`
	Warnings             []string   `json:"warnings,omitempty"`
}

type AssessmentResult struct {
	Age        Age                        `json:"age"`
	Scales     map[string]ScaleResult     `json:"scales"`
	Composites map[string]CompositeResult `json:"composites"`
	Warnings   []string                   `json:"warnings,omitempty"`
}
