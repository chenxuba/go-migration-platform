package model

import "time"

type OrderTagPagedQueryDTO struct {
	PageRequestModel PageRequestModel          `json:"pageRequestModel"`
	QueryModel       OrderTagPagedQueryFilters `json:"queryModel"`
}

type OrderTagPagedQueryFilters struct {
	Enable *bool `json:"enable"`
}

type OrderTagManageVO struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Enable        bool       `json:"enable"`
	OrgOrderTagID string     `json:"orgOrderTagId"`
	CreatedTime   *time.Time `json:"createdTime,omitempty"`
	UpdatedTime   *time.Time `json:"updatedTime,omitempty"`
}

type OrderTagPagedResult struct {
	List  []OrderTagManageVO `json:"list"`
	Total int                `json:"total"`
}

type CreateOrderTagDTO struct {
	Name string `json:"name"`
}

type UpdateOrderTagDTO struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Enable *bool  `json:"enable"`
}
