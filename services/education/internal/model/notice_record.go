package model

import "time"

const (
	NoticeStatusPendingAudit   = 1
	NoticeStatusAuditRejected  = 2
	NoticeStatusPendingPublish = 3
	NoticeStatusPublished      = 4
)

type NoticeFilterWordCheckDTO struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Summary string `json:"summary"`
}

type NoticeRepeatStudentCheckDTO struct {
	StudentIDs []string `json:"studentIds"`
}

type NoticeCreateDTO struct {
	NoticeTemplateID string   `json:"noticeTemplateId"`
	Title            string   `json:"title"`
	Content          string   `json:"content"`
	IsAllSchool      bool     `json:"isAllSchool"`
	IsDelaySend      bool     `json:"isDelaySend"`
	PublishDate      string   `json:"publishDate"`
	Hour             int      `json:"hour"`
	IsConfirm        bool     `json:"isConfirm"`
	Summary          string   `json:"summary"`
	ClassIDs         []string `json:"classIds"`
	StudentIDs       []string `json:"studentIds"`
}

type NoticeWithdrawDTO struct {
	NoticeID string `json:"noticeId"`
}

type NoticePageQueryDTO struct {
	PageRequestModel PageRequestModel     `json:"pageRequestModel"`
	QueryModel       NoticePageQueryModel `json:"queryModel"`
}

type NoticePageQueryModel struct {
	Statuses         []int  `json:"statuses"`
	IsWithdraw       *bool  `json:"isWithdraw"`
	BeginPublishDate string `json:"beginPublishDate"`
	EndPublishDate   string `json:"endPublishDate"`
	OperatorID       string `json:"operatorId"`
}

type NoticeClassSnapshot struct {
	ClassID   string `json:"classId"`
	ClassName string `json:"className"`
}

type NoticeClassVO struct {
	ClassID   string `json:"classId"`
	NoticeID  string `json:"noticeId"`
	ClassName string `json:"className"`
}

type NoticeCreateInput struct {
	NoticeTemplateID   int64
	Title              string
	Content            string
	Summary            string
	IsAllSchool        bool
	IsDelaySend        bool
	IsConfirm          bool
	PublishHour        int
	PublishTime        *time.Time
	RealityPublishTime *time.Time
	Status             int
	OperatorID         int64
	OperatorName       string
	StudentCount       int
	ClassIDs           []string
	ClassSnapshots     []NoticeClassSnapshot
	ExplicitStudentIDs []string
	TargetStudentIDs   []string
}

type NoticePageItemVO struct {
	NoticeID            string          `json:"noticeId"`
	Title               string          `json:"title"`
	Content             string          `json:"content"`
	Summary             string          `json:"summary"`
	IsAllSchool         bool            `json:"isAllSchool"`
	IsConfirm           bool            `json:"isConfirm"`
	IsRemind            bool            `json:"isRemind"`
	IsWithdraw          bool            `json:"isWithdraw"`
	Classs              []NoticeClassVO `json:"classs"`
	StudentCount        int             `json:"studentCount"`
	ReadStudentCount    int             `json:"readStudentCount"`
	ConfirmStudentCount int             `json:"confirmStudentCount"`
	OperatorID          string          `json:"operatorId"`
	OperatorName        string          `json:"operatorName"`
	OperationDate       *time.Time      `json:"operationDate"`
	IsDelaySend         bool            `json:"isDelaySend"`
	PublishTime         *time.Time      `json:"publishTime"`
	Status              int             `json:"status"`
	RealityPublishTime  *time.Time      `json:"realityPublishTime"`
}

type NoticePageResultVO struct {
	List  []NoticePageItemVO `json:"list"`
	Total int                `json:"total"`
}
