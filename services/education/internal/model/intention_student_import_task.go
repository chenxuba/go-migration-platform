package model

import "time"

type IntentionStudentImportUploadResult struct {
	FileURL  string `json:"fileUrl"`
	FileName string `json:"fileName"`
}

type IntentionStudentImportSubmitRequest struct {
	FileURL  string `json:"fileUrl"`
	FileName string `json:"fileName"`
}

type IntentionStudentImportTaskDetail struct {
	ID               string     `json:"id"`
	FileName         string     `json:"fileName"`
	UploadStaffID    string     `json:"uploadStaffId"`
	UploadStaffName  string     `json:"uploadStaffName"`
	ExecuteStaffID   *string    `json:"executeStaffId,omitempty"`
	ExecuteStaffName *string    `json:"executeStaffName,omitempty"`
	TotalRows        int        `json:"totalRows"`
	ExecutedRows     int        `json:"executedRows"`
	DeletedRows      int        `json:"deletedRows"`
	ErrorRows        int        `json:"errorRows"`
	CreatedTime      *time.Time `json:"createdTime,omitempty"`
	ConfirmTime      *time.Time `json:"confirmTime,omitempty"`
	CompleteTime     *time.Time `json:"completeTime,omitempty"`
	Status           int        `json:"status"`
	InstName         string     `json:"instName"`
}

type IntentionStudentImportTaskRecordListQuery struct {
	QueryModel struct {
		TaskID string `json:"taskId"`
		Type   int    `json:"type"`
	} `json:"queryModel"`
}

type IntentionStudentImportTaskRecordListResult struct {
	List    []IntentionStudentImportRow    `json:"list"`
	Total   int                            `json:"total"`
	Columns []IntentionStudentImportColumn `json:"columns"`
}

type IntentionStudentImportTaskListResult struct {
	List  []IntentionStudentImportTaskDetail `json:"list"`
	Total int                                `json:"total"`
}

type OrderImportStartResult struct {
	SuccessCount int `json:"successCount"`
	FailCount    int `json:"failCount"`
}
