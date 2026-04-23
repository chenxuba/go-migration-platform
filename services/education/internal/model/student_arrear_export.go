package model

import "time"

type StudentRegistrationArrearExportCreateRequest struct {
	QueryModel      StudentRegistrationArrearQueryModel `json:"queryModel"`
	QueryConditions []ExportConditionItem               `json:"queryConditions"`
}

type StudentLessonArrearExportCreateRequest struct {
	QueryModel      StudentLessonArrearQueryModel `json:"queryModel"`
	QueryConditions []ExportConditionItem         `json:"queryConditions"`
}

type StudentArrearExportRecord struct {
	ID              int64                 `json:"id"`
	ExportType      string                `json:"exportType"`
	FileName        string                `json:"fileName"`
	ExporterName    string                `json:"exporterName"`
	TotalRows       int                   `json:"totalRows"`
	QueryConditions []ExportConditionItem `json:"queryConditions"`
	CreatedTime     *time.Time            `json:"createdTime,omitempty"`
	ExpiresAt       *time.Time            `json:"expiresAt,omitempty"`
	DownloadURL     string                `json:"downloadUrl,omitempty"`
}
