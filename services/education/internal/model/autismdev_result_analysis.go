package model

type AutismDevResultAnalysisGenerateRequest struct {
	ID int64 `json:"id"`
}

type AutismDevResultAnalysisSaveRequest struct {
	ID       int64                     `json:"id"`
	Analysis AutismDevResultAnalysisVO `json:"analysis"`
}

type AutismDevResultAnalysisExportRequest struct {
	ID       int64                      `json:"id"`
	Analysis *AutismDevResultAnalysisVO `json:"analysis,omitempty"`
}

type AutismDevSelectedReportExportRequest struct {
	ID       int64                      `json:"id"`
	Sections []string                   `json:"sections"`
	Analysis *AutismDevResultAnalysisVO `json:"analysis,omitempty"`
}

type AutismDevResultAnalysisVO struct {
	Title       string                       `json:"title"`
	Model       string                       `json:"model,omitempty"`
	GeneratedBy string                       `json:"generatedBy,omitempty"`
	GeneratedAt string                       `json:"generatedAt,omitempty"`
	Rows        []AutismDevResultAnalysisRow `json:"rows"`
}

type AutismDevResultAnalysisRow struct {
	Domain     string `json:"domain"`
	Status     string `json:"status"`
	Strengths  string `json:"strengths"`
	Weaknesses string `json:"weaknesses"`
	Targets    string `json:"targets"`
}
