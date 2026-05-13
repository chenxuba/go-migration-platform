package model

import "time"

type PEP3IEPMaterialImportUploadResult struct {
	FileURL  string `json:"fileUrl"`
	FileName string `json:"fileName"`
}

type PEP3IEPMaterialImportSubmitRequest struct {
	FileURL  string `json:"fileUrl"`
	FileName string `json:"fileName"`
}

type PEP3IEPMaterialImportTemplateColumn struct {
	Title     string   `json:"title"`
	Required  bool     `json:"required"`
	FieldType int      `json:"fieldType"`
	Options   []string `json:"options,omitempty"`
}

type PEP3IEPMaterialImportColumn struct {
	Key       string   `json:"key"`
	Title     string   `json:"title"`
	Required  bool     `json:"required"`
	FieldType int      `json:"fieldType"`
	Options   []string `json:"options,omitempty"`
}

type PEP3IEPMaterialImportCell struct {
	Key        string `json:"key"`
	Title      string `json:"title"`
	Value      string `json:"value"`
	SelectedID any    `json:"selectedId,omitempty"`
	Error      string `json:"error,omitempty"`
}

type PEP3IEPMaterialImportRow struct {
	ID       string                      `json:"id"`
	RowNo    int                         `json:"rowNo"`
	HasError bool                        `json:"hasError"`
	Cells    []PEP3IEPMaterialImportCell `json:"cells"`
	Status   int                         `json:"status"`
	Result   string                      `json:"result,omitempty"`
}

type PEP3IEPMaterialImportParseResult struct {
	ImportID      string                        `json:"importId"`
	FileName      string                        `json:"fileName"`
	InstName      string                        `json:"instName"`
	Columns       []PEP3IEPMaterialImportColumn `json:"columns"`
	Rows          []PEP3IEPMaterialImportRow    `json:"rows"`
	NormalCount   int                           `json:"normalCount"`
	AbnormalCount int                           `json:"abnormalCount"`
}

type PEP3IEPMaterialImportTaskDetail struct {
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

type PEP3IEPMaterialImportTaskRecordListQuery struct {
	QueryModel struct {
		TaskID string `json:"taskId"`
		Type   int    `json:"type"`
	} `json:"queryModel"`
}

type PEP3IEPMaterialImportTaskRecordListResult struct {
	List    []PEP3IEPMaterialImportRow    `json:"list"`
	Total   int                           `json:"total"`
	Columns []PEP3IEPMaterialImportColumn `json:"columns"`
}

type PEP3IEPMaterialImportTaskListResult struct {
	List  []PEP3IEPMaterialImportTaskDetail `json:"list"`
	Total int                               `json:"total"`
}

type PEP3IEPMaterialImportSaveTaskRecordRequest struct {
	TaskID  string                     `json:"taskId"`
	Records []PEP3IEPMaterialImportRow `json:"records"`
}

type PEP3IEPMaterialImportStartTaskRequest struct {
	TaskID string `json:"taskId"`
}
