package model

type ShuangxiResultAnalysisGenerateRequest struct {
	ID int64 `json:"id"`
}

type ShuangxiResultAnalysisSaveRequest struct {
	ID       int64                    `json:"id"`
	Analysis ShuangxiResultAnalysisVO `json:"analysis"`
}

type ShuangxiResultAnalysisVO struct {
	Title       string                      `json:"title"`
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
