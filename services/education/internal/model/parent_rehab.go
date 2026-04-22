package model

type ParentRehabRecordQueryDTO struct {
	StudentID string `json:"studentId"`
	PageIndex int    `json:"pageIndex"`
	PageSize  int    `json:"pageSize"`
}

type ParentRehabRecordDetailQueryDTO struct {
	StudentID               string `json:"studentId"`
	StudentTeachingRecordID string `json:"studentTeachingRecordId"`
}

type ParentRehabFeedbackSaveDTO struct {
	StudentID               string `json:"studentId"`
	StudentTeachingRecordID string `json:"studentTeachingRecordId"`
	ParentFeedback          string `json:"parentFeedback"`
	ParentSignature         string `json:"parentSignature"`
}

type ParentUploadedFileVO struct {
	URL string `json:"url"`
}

type ParentRehabRecordSummaryVO struct {
	Students  []ParentBoundStudentVO `json:"students"`
	Items     []ParentRehabRecordVO  `json:"items"`
	PageIndex int                    `json:"pageIndex"`
	PageSize  int                    `json:"pageSize"`
	Total     int                    `json:"total"`
	HasMore   bool                   `json:"hasMore"`
}

type ParentRehabRecordVO struct {
	ID                      string `json:"id"`
	StudentTeachingRecordID string `json:"studentTeachingRecordId"`
	TeachingRecordID        string `json:"teachingRecordId"`
	InstID                  int64  `json:"instId"`
	CampusID                string `json:"campusId"`
	CampusName              string `json:"campusName"`
	StudentID               string `json:"studentId"`
	StudentName             string `json:"studentName"`
	StudentAvatarURL        string `json:"studentAvatarUrl,omitempty"`
	Date                    string `json:"date"`
	StartTime               string `json:"startTime"`
	EndTime                 string `json:"endTime"`
	LessonTime              string `json:"lessonTime"`
	ClassName               string `json:"className"`
	CourseName              string `json:"courseName"`
	TeacherName             string `json:"teacherName"`
	Classroom               string `json:"classroom"`
	StatusText              string `json:"statusText"`
	SummaryText             string `json:"summaryText"`
	TrainingTarget          string `json:"trainingTarget"`
	Performance             string `json:"performance"`
	Suggestion              string `json:"suggestion"`
	Remark                  string `json:"remark"`
	UpdatedTime             string `json:"updatedTime"`
	UpdatedStaffName        string `json:"updatedStaffName"`
	HasPublished            bool   `json:"hasPublished"`
}

type ParentRehabRecordDetailVO struct {
	Student           ParentBoundStudentVO        `json:"student"`
	Record            ParentRehabRecordVO         `json:"record"`
	Published         *StudentRehabRecordSnapshot `json:"published,omitempty"`
	PreviousPublished *StudentRehabRecordSnapshot `json:"previousPublished,omitempty"`
}
