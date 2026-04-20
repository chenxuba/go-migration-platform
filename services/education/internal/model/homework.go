package model

import "time"

const (
	HomeworkSourceTypeClass    = 1
	HomeworkSourceTypeOneToOne = 2
)

const (
	HomeworkPublishRuleOnce = 1
	HomeworkPublishRuleAuto = 2
)

const (
	HomeworkAttachmentTypeImage = 1
	HomeworkAttachmentTypeVideo = 2
)

type HomeworkAttachment struct {
	Type       int    `json:"type"`
	URL        string `json:"url"`
	Duration   int    `json:"duration"`
	Name       string `json:"name"`
	ExtendName string `json:"extendName"`
}

type HomeworkRepeatRule struct {
	StartDate  string `json:"startDate"`
	EndDate    string `json:"endDate"`
	RepeatSpan int    `json:"repeatSpan"`
	WeekDays   int    `json:"weekDays"`
}

type HomeworkObjectDTO struct {
	SourceType int      `json:"sourceType"`
	SourceID   string   `json:"sourceId"`
	StudentIDs []string `json:"studentIds"`
}

type HomeworkSelectedStudent struct {
	SourceType       int    `json:"sourceType"`
	SourceID         string `json:"sourceId"`
	SourceName       string `json:"sourceName"`
	StudentID        string `json:"studentId"`
	StudentName      string `json:"studentName"`
	TuitionAccountID string `json:"tuitionAccountId"`
	IsBind           bool   `json:"isBind"`
}

type HomeworkOperationResult struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type HomeworkBatchCreateDTO struct {
	Title            string               `json:"title"`
	Content          string               `json:"content"`
	Attachments      []HomeworkAttachment `json:"attachments"`
	RepeatRule       *HomeworkRepeatRule  `json:"repeatRule"`
	PublishTime      string               `json:"publishTime"`
	EndTime          string               `json:"endTime"`
	PublishHour      int                  `json:"publishHour"`
	EndHour          int                  `json:"endHour"`
	IsVisibleStudent bool                 `json:"isVisibleStudent"`
	HomeworkObjects  []HomeworkObjectDTO  `json:"homeworkObjects"`
}

type HomeworkUpdateDTO struct {
	ID               string               `json:"id"`
	Title            string               `json:"title"`
	Content          string               `json:"content"`
	Attachments      []HomeworkAttachment `json:"attachments"`
	RepeatRule       *HomeworkRepeatRule  `json:"repeatRule"`
	PublishTime      string               `json:"publishTime"`
	EndTime          string               `json:"endTime"`
	PublishHour      int                  `json:"publishHour"`
	EndHour          int                  `json:"endHour"`
	IsVisibleStudent bool                 `json:"isVisibleStudent"`
	HomeworkObjects  []HomeworkObjectDTO  `json:"homeworkObjects"`
}

type HomeworkDeleteDTO struct {
	ID string `json:"id"`
}

type HomeworkDetailQueryDTO struct {
	ID string `json:"id"`
}

type HomeworkListQueryDTO struct {
	PageRequestModel PageRequestModel       `json:"pageRequestModel"`
	SortModel        HomeworkListSortModel  `json:"sortModel"`
	QueryModel       HomeworkListQueryModel `json:"queryModel"`
}

type HomeworkListSortModel struct {
	PublishTime int `json:"publishTime"`
}

type HomeworkListQueryModel struct {
	TeacherIDs       []string `json:"teacherIds"`
	ClassID          string   `json:"classId"`
	OneToOneID       string   `json:"one2OneId"`
	PublishStartTime string   `json:"publishStartTime"`
	PublishEndTime   string   `json:"publishEndTime"`
	EndStartTime     string   `json:"endStartTime"`
	EndEndTime       string   `json:"endEndTime"`
	HasUnevaluated   *bool    `json:"hasUnevaluated"`
	HasUnsubmitted   *bool    `json:"hasUnsubmitted"`
}

type HomeworkListItemVO struct {
	ID               string               `json:"id"`
	Title            string               `json:"title"`
	Content          string               `json:"content"`
	StudentCount     int                  `json:"studentCount"`
	UnsubmittedCount int                  `json:"unsubmittedCount"`
	RejectedCount    int                  `json:"rejectedCount"`
	SubmittedCount   int                  `json:"submittedCount"`
	ReSubmittedCount int                  `json:"reSubmittedCount"`
	EvaluatedCount   int                  `json:"evaluatedCount"`
	UnevaluatedCount int                  `json:"unevaluatedCount"`
	IsVisibleStudent bool                 `json:"isVisibleStudent"`
	SourceType       int                  `json:"sourceType"`
	SourceID         string               `json:"sourceId"`
	SourceName       string               `json:"sourceName"`
	CreatedStaffID   string               `json:"createdStaffId"`
	CreatedStaffName string               `json:"createdStaffName"`
	CreatedTime      *time.Time           `json:"createdTime"`
	PublishTime      *time.Time           `json:"publishTime"`
	EndTime          *time.Time           `json:"endTime"`
	EndHour          int                  `json:"endHour"`
	ReadCount        int                  `json:"readCount"`
	UnreadCount      int                  `json:"unreadCount"`
	Attachments      []HomeworkAttachment `json:"attachments,omitempty"`
	PublishRule      int                  `json:"publishRule"`
	RepeatRule       *HomeworkRepeatRule  `json:"repeatRule,omitempty"`
}

type HomeworkDetailVO struct {
	HomeworkListItemVO
	PublishHour      int                       `json:"publishHour"`
	SelectedStudents []HomeworkSelectedStudent `json:"selectedStudents"`
}

type HomeworkPageResultVO struct {
	List  []HomeworkListItemVO `json:"list"`
	Total int                  `json:"total"`
}

type HomeworkStatisticsVO struct {
	UnsubmittedCount int `json:"unsubmittedCount"`
	UnevaluatedCount int `json:"unevaluatedCount"`
}

type HomeworkTargetStudent struct {
	StudentID        int64
	StudentName      string
	TuitionAccountID string
	IsBind           bool
}

type HomeworkMutationInput struct {
	Title            string
	Content          string
	Attachments      []HomeworkAttachment
	RepeatRule       *HomeworkRepeatRule
	PublishRule      int
	PublishTime      *time.Time
	EndTime          *time.Time
	PublishHour      int
	EndHour          int
	IsVisibleStudent bool
	SourceType       int
	SourceID         int64
	SourceName       string
	SelectedStudents []HomeworkSelectedStudent
}
