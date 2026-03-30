package model

type SuspendResumeTuitionAccountOrderDTO struct {
	TuitionAccountID string `json:"tuitionAccountId"`
	Type             int    `json:"type"`
	ExpireTime       string `json:"expireTime"`
	ExpireType       int    `json:"expireType"`
	Remark           string `json:"remark"`
	SuspendDate      string `json:"suspendDate"`
	ResumeDate       string `json:"resumeDate"`
}

type SuspendResumeTuitionAccountOrderResult struct {
	ID        string `json:"id"`
	StudentID string `json:"studentId"`
	LessonID  string `json:"lessonId"`
}
