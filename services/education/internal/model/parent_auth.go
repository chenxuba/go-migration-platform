package model

const ParentLoginTypeMiniProgram = "parent-miniapp"

type ParentWeChatLoginDTO struct {
	LoginCode string `json:"loginCode"`
	PhoneCode string `json:"phoneCode"`
}

type ParentBindStudentsDTO struct {
	StudentIDs []int64 `json:"studentIds"`
}

type ParentWeChatLoginVO struct {
	Token       string                     `json:"token"`
	Phone       string                     `json:"phone"`
	MaskedPhone string                     `json:"maskedPhone"`
	Nickname    string                     `json:"nickname"`
	MiniOpenID  string                     `json:"miniOpenId,omitempty"`
	UnionID     string                     `json:"unionId,omitempty"`
	Candidates  []ParentStudentCandidateVO `json:"candidates"`
}

type ParentCampusSummaryVO struct {
	Items []ParentCampusVO `json:"items"`
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
