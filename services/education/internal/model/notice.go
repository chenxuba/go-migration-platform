package model

type NoticeTemplateVO struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	CoverURL string `json:"coverUrl"`
	Tag      string `json:"tag"`
	Weight   int    `json:"weight"`
	Content  string `json:"content"`
	Summary  string `json:"summary"`
	OrgID    string `json:"orgId"`
	SchoolID string `json:"schoolId"`
}
