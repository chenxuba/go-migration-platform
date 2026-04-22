package model

const ParentLoginTypeMiniProgram = "parent-miniapp"

type ParentWeChatLoginDTO struct {
	LoginCode  string `json:"loginCode"`
	PhoneCode  string `json:"phoneCode"`
	BindTicket string `json:"bindTicket"`
}

type ParentBindStudentsDTO struct {
	StudentIDs []int64 `json:"studentIds"`
}

type ParentScheduleQueryDTO struct {
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

type ParentWeChatLoginVO struct {
	Token       string                     `json:"token"`
	Phone       string                     `json:"phone"`
	MaskedPhone string                     `json:"maskedPhone"`
	Nickname    string                     `json:"nickname"`
	MiniOpenID  string                     `json:"miniOpenId,omitempty"`
	UnionID     string                     `json:"unionId,omitempty"`
	Candidates  []ParentStudentCandidateVO `json:"candidates,omitempty"`
}

type ParentCampusSummaryVO struct {
	Items []ParentCampusVO `json:"items"`
}

type ParentScheduleSummaryVO struct {
	Items []ParentScheduleVO `json:"items"`
}

type ParentScheduleDateSummaryVO struct {
	Items []ParentScheduleDateVO `json:"items"`
}

type ParentCampusVO struct {
	ID           string `json:"id"`
	InstID       int64  `json:"instId"`
	Name         string `json:"name"`
	BrandName    string `json:"brandName"`
	ShortName    string `json:"shortName"`
	LogoURL      string `json:"logoUrl,omitempty"`
	StudentCount int    `json:"studentCount"`
}

type ParentStudentLookupByPhoneVO struct {
	Phone       string                     `json:"phone"`
	MaskedPhone string                     `json:"maskedPhone"`
	Candidates  []ParentStudentCandidateVO `json:"candidates"`
}

type ParentBoundStudentSummaryVO struct {
	Phone       string                     `json:"phone"`
	MaskedPhone string                     `json:"maskedPhone"`
	Students    []ParentStudentCandidateVO `json:"students"`
}

type ParentPendingStudentSummaryVO struct {
	Phone       string                     `json:"phone"`
	MaskedPhone string                     `json:"maskedPhone"`
	Count       int                        `json:"count"`
	Candidates  []ParentStudentCandidateVO `json:"candidates"`
}

type ParentWeChatOfficialStatusVO struct {
	Subscribed          bool   `json:"subscribed"`
	OfficialAccountName string `json:"officialAccountName"`
	NeedFollowGuide     bool   `json:"needFollowGuide"`
	BoundStudentCount   int    `json:"boundStudentCount"`
	SubscribedBindCount int    `json:"subscribedBindCount"`
	LastUnsubscribeAt   string `json:"lastUnsubscribeAt,omitempty"`
}

type ParentScheduleVO struct {
	ID               string `json:"id"`
	ScheduleID       string `json:"scheduleId"`
	InstID           int64  `json:"instId"`
	CampusID         string `json:"campusId"`
	CampusName       string `json:"campusName"`
	Date             string `json:"date"`
	StudentID        string `json:"studentId"`
	StudentName      string `json:"studentName"`
	StudentAvatarURL string `json:"studentAvatarUrl,omitempty"`
	StartTime        string `json:"startTime"`
	EndTime          string `json:"endTime"`
	CourseName       string `json:"courseName"`
	ClassName        string `json:"className"`
	TeacherName      string `json:"teacherName"`
	Classroom        string `json:"classroom"`
	Note             string `json:"note"`
	StatusText       string `json:"statusText"`
	CallStatus       int    `json:"callStatus"`
	CallStatusText   string `json:"callStatusText,omitempty"`
}

type ParentScheduleDateVO struct {
	InstID        int64  `json:"instId"`
	CampusID      string `json:"campusId"`
	CampusName    string `json:"campusName"`
	Date          string `json:"date"`
	ScheduleCount int    `json:"scheduleCount"`
}

type ParentStudentCandidateVO struct {
	ID                int64  `json:"id"`
	InstID            int64  `json:"instId"`
	CampusID          string `json:"campusId"`
	CampusName        string `json:"campusName"`
	CampusLogoURL     string `json:"campusLogoUrl,omitempty"`
	Name              string `json:"name"`
	AvatarURL         string `json:"avatarUrl,omitempty"`
	Mobile            string `json:"mobile"`
	MaskedMobile      string `json:"maskedMobile"`
	StudentStatus     int    `json:"studentStatus"`
	StudentStatusText string `json:"studentStatusText"`
	PhoneRelationship int    `json:"phoneRelationship"`
	RelationText      string `json:"relationText"`
	IsBound           bool   `json:"isBound"`
	ClassLabel        string `json:"classLabel"`
}
