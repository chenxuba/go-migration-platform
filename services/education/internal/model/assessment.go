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
	TemplateCode        string                    `json:"templateCode"`
	TemplateVersion     string                    `json:"templateVersion"`
	Title               string                    `json:"title"`
	ScaleCode           string                    `json:"scaleCode"`
	ScaleVersion        string                    `json:"scaleVersion"`
	DataStatus          string                    `json:"dataStatus,omitempty"`
	Sources             []string                  `json:"sources,omitempty"`
	Sections            []PEP3TemplateSection     `json:"sections"`
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

type PEP3TemplateField struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Value       string `json:"value"`
	RawValue    any    `json:"rawValue,omitempty"`
	Unit        string `json:"unit,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
}

type PEP3TemplateColumn struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Width int    `json:"width,omitempty"`
	Align string `json:"align,omitempty"`
}

type PEP3TemplateTable struct {
	Columns    []PEP3TemplateColumn `json:"columns"`
	Rows       []map[string]any     `json:"rows"`
	FooterRows []map[string]any     `json:"footerRows,omitempty"`
}

type PEP3TemplateSection struct {
	SectionCode     string              `json:"sectionCode"`
	Title           string              `json:"title"`
	Type            string              `json:"type"`
	Layout          string              `json:"layout,omitempty"`
	SourcePDFPageNo int                 `json:"sourcePdfPageNo,omitempty"`
	BookletPageNo   int                 `json:"bookletPageNo,omitempty"`
	Fields          []PEP3TemplateField `json:"fields,omitempty"`
	Table           *PEP3TemplateTable  `json:"table,omitempty"`
	TextItems       []string            `json:"textItems,omitempty"`
	Meta            map[string]any      `json:"meta,omitempty"`
}

type PEP3BookletVO struct {
	Record          AssessmentRecordSummaryVO `json:"record"`
	TemplateCode    string                    `json:"templateCode"`
	TemplateVersion string                    `json:"templateVersion"`
	Title           string                    `json:"title"`
	ScaleCode       string                    `json:"scaleCode"`
	ScaleVersion    string                    `json:"scaleVersion"`
	DataStatus      string                    `json:"dataStatus,omitempty"`
	Sources         []string                  `json:"sources,omitempty"`
	SourcePDF       string                    `json:"sourcePdf"`
	Pages           []PEP3BookletPage         `json:"pages"`
	Warnings        []string                  `json:"warnings,omitempty"`
}

type PEP3BookletPage struct {
	PageNo          int                   `json:"pageNo"`
	SourcePDFPageNo int                   `json:"sourcePdfPageNo"`
	Title           string                `json:"title"`
	PageType        string                `json:"pageType"`
	Sections        []PEP3TemplateSection `json:"sections"`
	Meta            map[string]any        `json:"meta,omitempty"`
}
