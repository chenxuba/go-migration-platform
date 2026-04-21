package model

type ParentClassRecordQueryDTO struct {
	StudentID string `json:"studentId"`
	PageIndex int    `json:"pageIndex"`
	PageSize  int    `json:"pageSize"`
}

type ParentClassRecordSummaryVO struct {
	Students  []ParentBoundStudentVO `json:"students"`
	Items     []ParentClassRecordVO  `json:"items"`
	PageIndex int                    `json:"pageIndex"`
	PageSize  int                    `json:"pageSize"`
	Total     int                    `json:"total"`
	HasMore   bool                   `json:"hasMore"`
}

type ParentBoundStudentVO struct {
	ID                string `json:"id"`
	InstID            int64  `json:"instId"`
	CampusID          string `json:"campusId"`
	CampusName        string `json:"campusName"`
	Name              string `json:"name"`
	AvatarURL         string `json:"avatarUrl,omitempty"`
	StudentStatus     int    `json:"studentStatus"`
	StudentStatusText string `json:"studentStatusText"`
}

type ParentClassRecordVO struct {
	ID                      string  `json:"id"`
	StudentTeachingRecordID string  `json:"studentTeachingRecordId"`
	TeachingRecordID        string  `json:"teachingRecordId"`
	InstID                  int64   `json:"instId"`
	CampusID                string  `json:"campusId"`
	CampusName              string  `json:"campusName"`
	StudentID               string  `json:"studentId"`
	StudentName             string  `json:"studentName"`
	StudentAvatarURL        string  `json:"studentAvatarUrl,omitempty"`
	Date                    string  `json:"date"`
	StartTime               string  `json:"startTime"`
	EndTime                 string  `json:"endTime"`
	ClassName               string  `json:"className"`
	CourseName              string  `json:"courseName"`
	TeacherName             string  `json:"teacherName"`
	Classroom               string  `json:"classroom"`
	Remark                  string  `json:"remark"`
	Status                  int     `json:"status"`
	StatusText              string  `json:"statusText"`
	ChargingMode            int     `json:"chargingMode"`
	ChargingModeText        string  `json:"chargingModeText"`
	DeductQuantity          float64 `json:"deductQuantity"`
	DeductDays              float64 `json:"deductDays"`
	ShowDeductQuantity      bool    `json:"showDeductQuantity"`
	ShowDeductDays          bool    `json:"showDeductDays"`
}
