package model

type IntentionStudentImportColumn struct {
	Key       string   `json:"key"`
	Title     string   `json:"title"`
	Required  bool     `json:"required"`
	FieldType int      `json:"fieldType"`
	FieldID   int64    `json:"fieldId"`
	Options   []string `json:"options,omitempty"`
}

type IntentionStudentImportCell struct {
	Key        string `json:"key"`
	Title      string `json:"title"`
	Value      string `json:"value"`
	SelectedID any    `json:"selectedId,omitempty"`
	Error      string `json:"error,omitempty"`
}

type IntentionStudentImportRow struct {
	ID       string                       `json:"id"`
	RowNo    int                          `json:"rowNo"`
	HasError bool                         `json:"hasError"`
	Cells    []IntentionStudentImportCell `json:"cells"`
	Status   int                          `json:"status"`
	Result   string                       `json:"result,omitempty"`
}

type IntentionStudentImportParseResult struct {
	ImportID      string                         `json:"importId"`
	FileName      string                         `json:"fileName"`
	InstName      string                         `json:"instName"`
	Columns       []IntentionStudentImportColumn `json:"columns"`
	Rows          []IntentionStudentImportRow    `json:"rows"`
	NormalCount   int                            `json:"normalCount"`
	AbnormalCount int                            `json:"abnormalCount"`
}

type IntentionStudentImportSaveTaskRecordRequest struct {
	TaskID  string                      `json:"taskId"`
	Records []IntentionStudentImportRow `json:"records"`
}

type IntentionStudentImportStartTaskRequest struct {
	TaskID string `json:"taskId"`
}
