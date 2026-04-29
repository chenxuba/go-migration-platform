package model

import (
	"encoding/json"
	"time"
)

type AssessmentRecordPageQueryDTO struct {
	PageRequestModel PageRequestModel           `json:"pageRequestModel"`
	QueryModel       AssessmentRecordQueryModel `json:"queryModel"`
}

type AssessmentRecordQueryModel struct {
	AssessmentCode      string `json:"assessmentCode,omitempty"`
	StudentID           *int64 `json:"studentId,omitempty"`
	SearchKey           string `json:"searchKey,omitempty"`
	AssessmentDateBegin string `json:"assessmentDateBegin,omitempty"`
	AssessmentDateEnd   string `json:"assessmentDateEnd,omitempty"`
}

type AssessmentRecordSummaryVO struct {
	ID             int64      `json:"id"`
	InstID         int64      `json:"instId"`
	StudentID      int64      `json:"studentId,omitempty"`
	StudentName    string     `json:"studentName,omitempty"`
	AssessmentCode string     `json:"assessmentCode"`
	AssessmentName string     `json:"assessmentName"`
	ScaleVersion   string     `json:"scaleVersion"`
	BirthDate      *time.Time `json:"birthDate,omitempty"`
	AssessmentDate *time.Time `json:"assessmentDate,omitempty"`
	AgeYears       int        `json:"ageYears"`
	AgeMonths      int        `json:"ageMonths"`
	AgeDays        int        `json:"ageDays"`
	NormAgeMonths  int        `json:"normAgeMonths"`
	ExaminerID     int64      `json:"examinerId,omitempty"`
	ExaminerName   string     `json:"examinerName,omitempty"`
	DataStatus     string     `json:"dataStatus,omitempty"`
	Remark         string     `json:"remark,omitempty"`
	CreatedTime    *time.Time `json:"createdTime,omitempty"`
	UpdatedTime    *time.Time `json:"updatedTime,omitempty"`
}

type AssessmentRecordDetailVO struct {
	AssessmentRecordSummaryVO
	InputJSON  json.RawMessage `json:"input,omitempty"`
	ResultJSON json.RawMessage `json:"result,omitempty"`
}

type PEP3ReportVO struct {
	Record              AssessmentRecordSummaryVO `json:"record"`
	ScaleCode           string                    `json:"scaleCode"`
	ScaleVersion        string                    `json:"scaleVersion"`
	DataStatus          string                    `json:"dataStatus,omitempty"`
	Sources             []string                  `json:"sources,omitempty"`
	BasicInfo           PEP3ReportBasicInfo       `json:"basicInfo"`
	DevelopmentRows     []PEP3ReportScaleRow      `json:"developmentRows"`
	BehaviorRows        []PEP3ReportScaleRow      `json:"behaviorRows"`
	CaregiverReportRows []PEP3ReportScaleRow      `json:"caregiverReportRows"`
	CompositeRows       []PEP3ReportCompositeRow  `json:"compositeRows"`
	Summary             []string                  `json:"summary"`
	Warnings            []string                  `json:"warnings,omitempty"`
}

type PEP3ReportBasicInfo struct {
	StudentID      int64  `json:"studentId,omitempty"`
	StudentName    string `json:"studentName,omitempty"`
	ExaminerID     int64  `json:"examinerId,omitempty"`
	ExaminerName   string `json:"examinerName,omitempty"`
	BirthDate      string `json:"birthDate,omitempty"`
	AssessmentDate string `json:"assessmentDate,omitempty"`
	AgeText        string `json:"ageText"`
	NormAgeMonths  int    `json:"normAgeMonths"`
	Remark         string `json:"remark,omitempty"`
}

type PEP3ReportScaleRow struct {
	ScaleCode          string   `json:"scaleCode"`
	ScaleName          string   `json:"scaleName"`
	Category           string   `json:"category"`
	RawScore           int      `json:"rawScore"`
	MaxRawScore        *int     `json:"maxRawScore,omitempty"`
	DevelopmentAgeText string   `json:"developmentAgeText,omitempty"`
	PercentileRankText string   `json:"percentileRankText,omitempty"`
	ScaledScoreText    string   `json:"scaledScoreText,omitempty"`
	Level              string   `json:"level,omitempty"`
	Warnings           []string `json:"warnings,omitempty"`
}

type PEP3ReportCompositeRow struct {
	CompositeCode        string   `json:"compositeCode"`
	CompositeName        string   `json:"compositeName"`
	MemberScaleCodes     []string `json:"memberScaleCodes"`
	StandardScoreSumText string   `json:"standardScoreSumText,omitempty"`
	PercentileRankText   string   `json:"percentileRankText,omitempty"`
	DevelopmentAgeText   string   `json:"developmentAgeText,omitempty"`
	Level                string   `json:"level,omitempty"`
	Warnings             []string `json:"warnings,omitempty"`
}
