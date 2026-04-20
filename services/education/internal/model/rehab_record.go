package model

type StudentRehabRecordQueryDTO struct {
	StudentTeachingRecordID string `json:"studentTeachingRecordId"`
}

type RehabRecordTemplateMeta struct {
	TemplateCode         string `json:"templateCode"`
	TemplateName         string `json:"templateName"`
	TemplateVersion      int    `json:"templateVersion"`
	TemplateScope        string `json:"templateScope"`
	TemplateAssignmentID string `json:"templateAssignmentId"`
}

type RehabRecordTrainingItem struct {
	Title     string                 `json:"title"`
	Content   string                 `json:"content"`
	MediaList []RehabRecordMediaItem `json:"mediaList,omitempty"`
}

type RehabRecordMediaItem struct {
	MediaType string `json:"mediaType"`
	URL       string `json:"url"`
	FileName  string `json:"fileName"`
	Size      int64  `json:"size"`
}

type RehabRecordContent struct {
	StudentName          string                    `json:"studentName"`
	Gender               string                    `json:"gender"`
	BirthDate            string                    `json:"birthDate"`
	ClassName            string                    `json:"className"`
	TeacherName          string                    `json:"teacherName"`
	TrainingDate         string                    `json:"trainingDate"`
	TrainingTarget       string                    `json:"trainingTarget"`
	TrainingItems        []RehabRecordTrainingItem `json:"trainingItems"`
	TrainingMediaList    []RehabRecordMediaItem    `json:"trainingMediaList,omitempty"`
	Performance          string                    `json:"performance"`
	PerformanceMediaList []RehabRecordMediaItem    `json:"performanceMediaList,omitempty"`
	Suggestion           string                    `json:"suggestion"`
	SuggestionMediaList  []RehabRecordMediaItem    `json:"suggestionMediaList,omitempty"`
	ParentFeedback       string                    `json:"parentFeedback"`
	ParentSignature      string                    `json:"parentSignature"`
	FeedbackDate         string                    `json:"feedbackDate"`
}

type SaveStudentRehabRecordDraftDTO struct {
	StudentTeachingRecordID string                  `json:"studentTeachingRecordId"`
	Template                RehabRecordTemplateMeta `json:"template"`
	Content                 RehabRecordContent      `json:"content"`
}

type PublishStudentRehabRecordDTO struct {
	StudentTeachingRecordID string                  `json:"studentTeachingRecordId"`
	Template                RehabRecordTemplateMeta `json:"template"`
	Content                 RehabRecordContent      `json:"content"`
}

type StudentRehabRecordSnapshot struct {
	Template         RehabRecordTemplateMeta `json:"template"`
	Content          RehabRecordContent      `json:"content"`
	UpdatedTime      string                  `json:"updatedTime"`
	UpdatedStaffName string                  `json:"updatedStaffName"`
}

type StudentRehabRecordDetailResult struct {
	HasDraft             bool                        `json:"hasDraft"`
	Draft                *StudentRehabRecordSnapshot `json:"draft,omitempty"`
	HasPublished         bool                        `json:"hasPublished"`
	Published            *StudentRehabRecordSnapshot `json:"published,omitempty"`
	HasPreviousPublished bool                        `json:"hasPreviousPublished"`
	PreviousPublished    *StudentRehabRecordSnapshot `json:"previousPublished,omitempty"`
}
