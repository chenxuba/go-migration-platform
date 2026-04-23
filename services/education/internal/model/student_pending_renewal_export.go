package model

import "time"

type PendingRenewalStudentExportCreateRequest struct {
	QueryModel      PendingRenewalStudentQueryDTO `json:"queryModel"`
	QueryConditions []ExportConditionItem         `json:"queryConditions"`
}

type PendingRenewalStudentExportRecord struct {
	ID              int64                 `json:"id"`
	FileName        string                `json:"fileName"`
	ExporterName    string                `json:"exporterName"`
	TotalRows       int                   `json:"totalRows"`
	QueryConditions []ExportConditionItem `json:"queryConditions"`
	CreatedTime     *time.Time            `json:"createdTime,omitempty"`
	ExpiresAt       *time.Time            `json:"expiresAt,omitempty"`
	DownloadURL     string                `json:"downloadUrl,omitempty"`
}
