package model

type SchoolHolidayVO struct {
	ID        int64  `json:"id"`
	InstID    int64  `json:"instId"`
	Name      string `json:"name"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	Source    string `json:"source"`
	Sort      int    `json:"sort"`
}

type SchoolHolidayMutation struct {
	ID        *int64 `json:"id,omitempty"`
	Name      string `json:"name"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	Source    string `json:"source,omitempty"`
	Sort      *int   `json:"sort,omitempty"`
}

type SchoolHolidayDeleteDTO struct {
	ID *int64 `json:"id,omitempty"`
}
