package model

import "time"

type ExportConditionItem struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type EnrolledStudentExportCreateRequest struct {
	QueryModel      EnrolledStudentFilter `json:"queryModel"`
	QueryConditions []ExportConditionItem `json:"queryConditions"`
}

type EnrolledStudentExportRecord struct {
	ID              int64                 `json:"id"`
	FileName        string                `json:"fileName"`
	ExporterName    string                `json:"exporterName"`
	TotalRows       int                   `json:"totalRows"`
	QueryConditions []ExportConditionItem `json:"queryConditions"`
	CreatedTime     *time.Time            `json:"createdTime,omitempty"`
	ExpiresAt       *time.Time            `json:"expiresAt,omitempty"`
	DownloadURL     string                `json:"downloadUrl,omitempty"`
}

type EnrolledStudentBalance struct {
	AvailableBalance float64
	GiftBalance      float64
}
