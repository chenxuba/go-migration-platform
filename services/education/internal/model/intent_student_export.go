package model

import "time"

type IntentStudentExportCreateRequest struct {
	QueryModel      IntentStudentFilters  `json:"queryModel"`
	SortModel       SortModel             `json:"sortModel"`
	QueryConditions []ExportConditionItem `json:"queryConditions"`
}

type IntentStudentExportRecord struct {
	ID              int64                 `json:"id"`
	FileName        string                `json:"fileName"`
	ExporterName    string                `json:"exporterName"`
	TotalRows       int                   `json:"totalRows"`
	QueryConditions []ExportConditionItem `json:"queryConditions"`
	CreatedTime     *time.Time            `json:"createdTime,omitempty"`
	ExpiresAt       *time.Time            `json:"expiresAt,omitempty"`
	DownloadURL     string                `json:"downloadUrl,omitempty"`
}
