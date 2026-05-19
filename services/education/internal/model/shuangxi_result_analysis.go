package model

type ShuangxiResultAnalysisGenerateRequest struct {
	ID int64 `json:"id"`
}

type ShuangxiResultAnalysisSaveRequest struct {
	ID       int64                    `json:"id"`
	Analysis ShuangxiResultAnalysisVO `json:"analysis"`
}

type ShuangxiResultAnalysisExportRequest struct {
	ID       int64                     `json:"id"`
	Analysis *ShuangxiResultAnalysisVO `json:"analysis,omitempty"`
}

type ShuangxiSelectedReportExportRequest struct {
	ID       int64                     `json:"id"`
	Sections []string                  `json:"sections"`
	Analysis *ShuangxiResultAnalysisVO `json:"analysis,omitempty"`
}

type ShuangxiResultAnalysisVO struct {
	Title       string                      `json:"title"`
	CourseName  string                      `json:"courseName,omitempty"`
	Model       string                      `json:"model,omitempty"`
	GeneratedBy string                      `json:"generatedBy,omitempty"`
	GeneratedAt string                      `json:"generatedAt,omitempty"`
	Rows        []ShuangxiResultAnalysisRow `json:"rows"`
}

type ShuangxiResultAnalysisRow struct {
	DomainCode string `json:"domainCode,omitempty"`
	Domain     string `json:"domain"`
	Strengths  string `json:"strengths"`
	Weaknesses string `json:"weaknesses"`
	Reason     string `json:"reason"`
	Strategy   string `json:"strategy"`
}
