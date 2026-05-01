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

type AssessmentDraftPageQueryDTO struct {
	PageRequestModel PageRequestModel          `json:"pageRequestModel"`
	QueryModel       AssessmentDraftQueryModel `json:"queryModel"`
}

type AssessmentDraftQueryModel struct {
	AssessmentCode      string `json:"assessmentCode,omitempty"`
	StudentID           *int64 `json:"studentId,omitempty"`
	SearchKey           string `json:"searchKey,omitempty"`
	Status              string `json:"status,omitempty"`
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

type AssessmentDraftSummaryVO struct {
	ID                int64                       `json:"id"`
	InstID            int64                       `json:"instId"`
	StudentID         int64                       `json:"studentId,omitempty"`
	StudentName       string                      `json:"studentName,omitempty"`
	AssessmentCode    string                      `json:"assessmentCode"`
	AssessmentName    string                      `json:"assessmentName"`
	ScaleVersion      string                      `json:"scaleVersion"`
	BirthDate         *time.Time                  `json:"birthDate,omitempty"`
	AssessmentDate    *time.Time                  `json:"assessmentDate,omitempty"`
	ExaminerID        int64                       `json:"examinerId,omitempty"`
	ExaminerName      string                      `json:"examinerName,omitempty"`
	Status            string                      `json:"status"`
	SubmittedRecordID int64                       `json:"submittedRecordId,omitempty"`
	AnsweredItemCount int                         `json:"answeredItemCount"`
	RawScoreCount     int                         `json:"rawScoreCount"`
	CompletionPercent float64                     `json:"completionPercent"`
	Progress          PEP3AssessmentDraftProgress `json:"progress"`
	Remark            string                      `json:"remark,omitempty"`
	CreatedTime       *time.Time                  `json:"createdTime,omitempty"`
	UpdatedTime       *time.Time                  `json:"updatedTime,omitempty"`
}

type AssessmentDraftDetailVO struct {
	AssessmentDraftSummaryVO
	InputJSON json.RawMessage `json:"input,omitempty"`
}

type PEP3AssessmentDraftSubmitVO struct {
	DraftID     int64                    `json:"draftId"`
	RecordID    int64                    `json:"recordId"`
	DraftStatus string                   `json:"draftStatus"`
	Record      AssessmentRecordDetailVO `json:"record"`
}

type PEP3AssessmentDraftProgress struct {
	ItemCount              int                  `json:"itemCount"`
	AnsweredItemCount      int                  `json:"answeredItemCount"`
	MissingItemCount       int                  `json:"missingItemCount"`
	RawScoreCount          int                  `json:"rawScoreCount"`
	CaregiverRawScoreCount int                  `json:"caregiverRawScoreCount"`
	TotalInputCount        int                  `json:"totalInputCount"`
	CompletedInputCount    int                  `json:"completedInputCount"`
	CompletionPercent      float64              `json:"completionPercent"`
	Complete               bool                 `json:"complete"`
	CanScore               bool                 `json:"canScore"`
	MissingRequiredFields  []string             `json:"missingRequiredFields,omitempty"`
	MissingItemNos         []int                `json:"missingItemNos,omitempty"`
	DomainProgress         []PEP3DomainProgress `json:"domainProgress,omitempty"`
}

type PEP3DomainProgress struct {
	ScaleCode         string `json:"scaleCode"`
	ScaleName         string `json:"scaleName"`
	Category          string `json:"category"`
	ItemCount         int    `json:"itemCount"`
	AnsweredItemCount int    `json:"answeredItemCount"`
	RawScore          *int   `json:"rawScore,omitempty"`
	MaxRawScore       *int   `json:"maxRawScore,omitempty"`
	Complete          bool   `json:"complete"`
}

type PEP3NormDataInfo struct {
	NormVersion             string `json:"normVersion"`
	DevelopmentAgeMaxMonths int    `json:"developmentAgeMaxMonths"`
	NormAgeBandMaxMonths    int    `json:"normAgeBandMaxMonths"`
	NormSourcePDF           string `json:"normSourcePdf"`
}

type PEP3AssessmentFormTemplateVO struct {
	TemplateCode    string `json:"templateCode"`
	TemplateVersion string `json:"templateVersion"`
	Title           string `json:"title"`
	ScaleCode       string `json:"scaleCode"`
	ScaleVersion    string `json:"scaleVersion"`
	PEP3NormDataInfo
	DataStatus      string                      `json:"dataStatus,omitempty"`
	Sources         []string                    `json:"sources,omitempty"`
	ItemCount       int                         `json:"itemCount"`
	ScoreOptions    []PEP3ScoreOption           `json:"scoreOptions"`
	BasicFields     []PEP3AssessmentFormField   `json:"basicFields"`
	Domains         []PEP3AssessmentDomain      `json:"domains"`
	RawScoreFields  []PEP3RawScoreField         `json:"rawScoreFields"`
	ItemGroups      []PEP3AssessmentItemGroup   `json:"itemGroups"`
	CaregiverReport PEP3CaregiverReportTemplate `json:"caregiverReport"`
	SubmitContract  PEP3SubmitContract          `json:"submitContract"`
}

type PEP3AssessmentFormField struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	FieldType   string `json:"fieldType"`
	Required    bool   `json:"required"`
	Placeholder string `json:"placeholder,omitempty"`
}

type PEP3AssessmentDomain struct {
	ScaleCode            string `json:"scaleCode"`
	ScaleName            string `json:"scaleName"`
	Category             string `json:"category"`
	ItemCount            *int   `json:"itemCount,omitempty"`
	MaxRawScore          *int   `json:"maxRawScore,omitempty"`
	ItemNumbers          []int  `json:"itemNumbers,omitempty"`
	IsDevelopmentSubtest bool   `json:"isDevelopmentSubtest,omitempty"`
	IsBehaviorSubtest    bool   `json:"isBehaviorSubtest,omitempty"`
	IsCaregiverReport    bool   `json:"isCaregiverReport,omitempty"`
	CompositeCode        string `json:"compositeCode,omitempty"`
}

type PEP3RawScoreField struct {
	ScaleCode   string `json:"scaleCode"`
	ScaleName   string `json:"scaleName"`
	Category    string `json:"category"`
	MinScore    int    `json:"minScore"`
	MaxScore    *int   `json:"maxScore,omitempty"`
	InputMode   string `json:"inputMode"`
	Required    bool   `json:"required"`
	Description string `json:"description,omitempty"`
}

type PEP3AssessmentItemGroup struct {
	GroupCode       string               `json:"groupCode"`
	Title           string               `json:"title"`
	BookletPageNo   int                  `json:"bookletPageNo"`
	SourcePDFPageNo int                  `json:"sourcePdfPageNo,omitempty"`
	Layout          string               `json:"layout,omitempty"`
	StartItemNo     int                  `json:"startItemNo"`
	EndItemNo       int                  `json:"endItemNo"`
	Items           []PEP3AssessmentItem `json:"items"`
}

type PEP3AssessmentItem struct {
	ItemNo       int                   `json:"itemNo"`
	ItemTitle    string                `json:"itemTitle"`
	TestItem     string                `json:"testItem"`
	Materials    string                `json:"materials,omitempty"`
	Method       string                `json:"method,omitempty"`
	Guidance     string                `json:"guidance,omitempty"`
	DomainCode   string                `json:"domainCode"`
	DomainName   string                `json:"domainName"`
	Standard     string                `json:"standard"`
	ScoreOptions []PEP3ScoreOption     `json:"scoreOptions"`
	RecordFields []PEP3ItemRecordField `json:"recordFields,omitempty"`
	SourcePDF    string                `json:"sourcePdf,omitempty"`
	SourcePages  []int                 `json:"sourcePages,omitempty"`
	OCRStatus    string                `json:"ocrStatus,omitempty"`
}

type PEP3ScoreOption struct {
	Value       int    `json:"value"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

type PEP3ItemRecordField struct {
	Key         string                      `json:"key"`
	Label       string                      `json:"label"`
	FieldType   string                      `json:"fieldType"`
	DisplayType string                      `json:"displayType,omitempty"`
	Required    bool                        `json:"required,omitempty"`
	Placeholder string                      `json:"placeholder,omitempty"`
	Options     []PEP3ItemRecordFieldOption `json:"options,omitempty"`
}

type PEP3ItemRecordFieldOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type PEP3CaregiverReportTemplate struct {
	ReportName   string                       `json:"reportName"`
	SourcePDF    string                       `json:"sourcePdf"`
	SubmitMode   string                       `json:"submitMode"`
	Instructions string                       `json:"instructions"`
	ScoreRules   []PEP3CaregiverScoreRule     `json:"scoreRules"`
	Sections     []PEP3CaregiverReportSection `json:"sections"`
}

type PEP3CaregiverScoreRule struct {
	ScaleCode   string `json:"scaleCode"`
	ScaleName   string `json:"scaleName"`
	SectionCode string `json:"sectionCode"`
	MaxRawScore int    `json:"maxRawScore"`
	Description string `json:"description"`
}

type PEP3CaregiverReportSection struct {
	SectionCode         string                           `json:"sectionCode"`
	Title               string                           `json:"title"`
	Description         string                           `json:"description,omitempty"`
	InputType           string                           `json:"inputType"`
	ScaleCode           string                           `json:"scaleCode,omitempty"`
	ScaleName           string                           `json:"scaleName,omitempty"`
	Scored              bool                             `json:"scored"`
	MaxRawScore         *int                             `json:"maxRawScore,omitempty"`
	Items               []PEP3CaregiverReportItem        `json:"items,omitempty"`
	DiagnosisCategories []PEP3CaregiverDiagnosisCategory `json:"diagnosisCategories,omitempty"`
}

type PEP3CaregiverDiagnosisCategory struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

type PEP3CaregiverReportItem struct {
	ItemNo    int                         `json:"itemNo"`
	Key       string                      `json:"key"`
	Prompt    string                      `json:"prompt"`
	FieldType string                      `json:"fieldType"`
	Unit      string                      `json:"unit,omitempty"`
	Required  bool                        `json:"required,omitempty"`
	Scored    bool                        `json:"scored"`
	Options   []PEP3CaregiverReportOption `json:"options,omitempty"`
}

type PEP3CaregiverReportOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Score *int   `json:"score,omitempty"`
}

type PEP3CaregiverReportInviteVO struct {
	DraftID                int64      `json:"draftId"`
	RecordID               int64      `json:"recordId,omitempty"`
	StudentName            string     `json:"studentName,omitempty"`
	Ticket                 string     `json:"ticket,omitempty"`
	Token                  string     `json:"token"`
	ExpiresAt              *time.Time `json:"expiresAt,omitempty"`
	MiniProgramPath        string     `json:"miniProgramPath"`
	MiniProgramEnvVersion  string     `json:"miniProgramEnvVersion,omitempty"`
	MiniProgramCodeDataURL string     `json:"miniProgramCodeDataUrl,omitempty"`
	WeChatURLLink          string     `json:"wechatUrlLink,omitempty"`
	QRCodeValue            string     `json:"qrCodeValue,omitempty"`
	QRCodeType             string     `json:"qrCodeType,omitempty"`
	QRCodeMessage          string     `json:"qrCodeMessage,omitempty"`
	URL                    string     `json:"url"`
}

type PEP3CaregiverReportPublicTemplateVO struct {
	DraftID        int64                          `json:"draftId"`
	StudentID      int64                          `json:"studentId,omitempty"`
	StudentName    string                         `json:"studentName,omitempty"`
	BirthDate      *time.Time                     `json:"birthDate,omitempty"`
	AssessmentDate *time.Time                     `json:"assessmentDate,omitempty"`
	Template       PEP3CaregiverReportTemplate    `json:"template"`
	Submission     *PEP3CaregiverReportSubmission `json:"submission,omitempty"`
}

type PEP3CaregiverReportSubmissionInput struct {
	Token          string                    `json:"token"`
	Ticket         string                    `json:"ticket,omitempty"`
	RespondentName string                    `json:"respondentName,omitempty"`
	Relationship   string                    `json:"relationship,omitempty"`
	Answers        map[string]map[string]any `json:"answers"`
}

type PEP3CaregiverReportSubmission struct {
	RespondentName string                    `json:"respondentName,omitempty"`
	Relationship   string                    `json:"relationship,omitempty"`
	Answers        map[string]map[string]any `json:"answers"`
	RawScores      map[string]int            `json:"rawScores"`
	RawScoreList   []PEP3CaregiverRawScore   `json:"rawScoreList"`
	SubmittedAt    *time.Time                `json:"submittedAt,omitempty"`
	Source         string                    `json:"source"`
}

type PEP3CaregiverRawScore struct {
	ScaleCode string `json:"scaleCode"`
	RawScore  int    `json:"rawScore"`
}

type PEP3CaregiverReportSubmitVO struct {
	DraftID       int64                       `json:"draftId"`
	RecordID      int64                       `json:"recordId,omitempty"`
	RecordUpdated bool                        `json:"recordUpdated"`
	StudentName   string                      `json:"studentName,omitempty"`
	RawScores     map[string]int              `json:"rawScores"`
	Progress      PEP3AssessmentDraftProgress `json:"progress"`
	SubmittedAt   *time.Time                  `json:"submittedAt,omitempty"`
}

type PEP3SubmitContract struct {
	ScoreEndpoint          string   `json:"scoreEndpoint"`
	CreateRecordEndpoint   string   `json:"createRecordEndpoint"`
	DateFormat             string   `json:"dateFormat"`
	ItemScoreListKey       string   `json:"itemScoreListKey"`
	RawScoreListKey        string   `json:"rawScoreListKey"`
	ItemRecordValuesKey    string   `json:"itemRecordValuesKey,omitempty"`
	ItemRecordValueListKey string   `json:"itemRecordValueListKey,omitempty"`
	RequiredBaseFields     []string `json:"requiredBaseFields"`
	AllowedItemScores      []int    `json:"allowedItemScores"`
}

type PEP3ReportVO struct {
	Record          AssessmentRecordSummaryVO `json:"record"`
	TemplateCode    string                    `json:"templateCode"`
	TemplateVersion string                    `json:"templateVersion"`
	Title           string                    `json:"title"`
	ScaleCode       string                    `json:"scaleCode"`
	ScaleVersion    string                    `json:"scaleVersion"`
	PEP3NormDataInfo
	DataStatus string                `json:"dataStatus,omitempty"`
	Sources    []string              `json:"sources,omitempty"`
	Sections   []PEP3TemplateSection `json:"sections"`
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
	CompositeCode        string            `json:"compositeCode"`
	CompositeName        string            `json:"compositeName"`
	MemberScaleCodes     []string          `json:"memberScaleCodes"`
	MemberScaleScores    map[string]string `json:"memberScaleScores,omitempty"`
	StandardScoreSumText string            `json:"standardScoreSumText,omitempty"`
	PercentileRankText   string            `json:"percentileRankText,omitempty"`
	DevelopmentAgeText   string            `json:"developmentAgeText,omitempty"`
	Level                string            `json:"level,omitempty"`
	Warnings             []string          `json:"warnings,omitempty"`
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
	Group string `json:"group,omitempty"`
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
	PEP3NormDataInfo
	DataStatus string            `json:"dataStatus,omitempty"`
	Sources    []string          `json:"sources,omitempty"`
	SourcePDF  string            `json:"sourcePdf"`
	Pages      []PEP3BookletPage `json:"pages"`
	Warnings   []string          `json:"warnings,omitempty"`
}

type PEP3BookletPage struct {
	PageNo          int                   `json:"pageNo"`
	SourcePDFPageNo int                   `json:"sourcePdfPageNo"`
	Title           string                `json:"title"`
	PageType        string                `json:"pageType"`
	Sections        []PEP3TemplateSection `json:"sections"`
	Meta            map[string]any        `json:"meta,omitempty"`
}
