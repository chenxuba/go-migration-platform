package model

import "time"

const (
	TemplateMessageBusinessTypePendingRenewal = "pending_renewal"

	TemplateMessageChannelWeChat = 1
	TemplateMessageChannelSMS    = 2

	TemplateMessageRecordItemStatusSuccess = 1
	TemplateMessageRecordItemStatusSkipped = 2
	TemplateMessageRecordItemStatusFailed  = 3

	PendingRenewalReminderHomeSchoolStatusUnknown    = 0
	PendingRenewalReminderHomeSchoolStatusFollowed   = 1
	PendingRenewalReminderHomeSchoolStatusUnfollowed = 2
)

type PendingRenewalReminderSendDTO struct {
	TuitionAccountIDs []string `json:"tuitionAccountIds"`
}

type PendingRenewalReminderSendResult struct {
	RecordID     string `json:"recordId"`
	NotifyCount  int    `json:"notifyCount"`
	SuccessCount int    `json:"successCount"`
	SkippedCount int    `json:"skippedCount"`
	FailedCount  int    `json:"failedCount"`
}

type PendingRenewalReminderRecordPageQueryDTO struct {
	PageRequestModel PageRequestModel `json:"pageRequestModel"`
}

type PendingRenewalReminderRecordPageItem struct {
	RecordID     string     `json:"recordId"`
	TemplateID   string     `json:"templateId"`
	TemplateName string     `json:"templateName"`
	Channel      int        `json:"channel"`
	ChannelName  string     `json:"channelName"`
	ReadCount    int        `json:"readCount"`
	NotifyCount  int        `json:"notifyCount"`
	SuccessCount int        `json:"successCount"`
	SkippedCount int        `json:"skippedCount"`
	FailedCount  int        `json:"failedCount"`
	OperatorID   string     `json:"operatorId"`
	OperatorName string     `json:"operatorName"`
	SendTime     *time.Time `json:"sendTime,omitempty"`
	UnsentCount  int        `json:"unsentCount"`
}

type PendingRenewalReminderRecordPageResult struct {
	List  []PendingRenewalReminderRecordPageItem `json:"list"`
	Total int                                    `json:"total"`
}

type PendingRenewalReminderRecordDetailItem struct {
	ItemID               string `json:"itemId"`
	TuitionAccountID     string `json:"tuitionAccountId"`
	StudentID            string `json:"studentId"`
	StudentName          string `json:"studentName"`
	Sex                  *int   `json:"sex,omitempty"`
	Avatar               string `json:"avatar"`
	Phone                string `json:"phone"`
	LessonName           string `json:"lessonName"`
	RemainingText        string `json:"remainingText"`
	HomeSchoolStatus     int    `json:"homeSchoolStatus"`
	HomeSchoolStatusText string `json:"homeSchoolStatusText"`
}

type PendingRenewalReminderRecordDetailResult struct {
	RecordID     string                                   `json:"recordId"`
	TemplateID   string                                   `json:"templateId"`
	TemplateName string                                   `json:"templateName"`
	Channel      int                                      `json:"channel"`
	ChannelName  string                                   `json:"channelName"`
	ReadCount    int                                      `json:"readCount"`
	NotifyCount  int                                      `json:"notifyCount"`
	SuccessCount int                                      `json:"successCount"`
	SkippedCount int                                      `json:"skippedCount"`
	FailedCount  int                                      `json:"failedCount"`
	UnsentCount  int                                      `json:"unsentCount"`
	OperatorID   string                                   `json:"operatorId"`
	OperatorName string                                   `json:"operatorName"`
	SendTime     *time.Time                               `json:"sendTime,omitempty"`
	SentList     []PendingRenewalReminderRecordDetailItem `json:"sentList"`
	UnsentList   []PendingRenewalReminderRecordDetailItem `json:"unsentList"`
}
