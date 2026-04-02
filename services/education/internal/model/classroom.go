package model

import "time"

type ClassroomQueryDTO struct {
	EnabledOnly *bool  `json:"enabledOnly,omitempty"`
	SearchKey   string `json:"searchKey,omitempty"`
}

type ClassroomVO struct {
	ID         int64      `json:"id"`
	UUID       string     `json:"uuid,omitempty"`
	Version    int64      `json:"version,omitempty"`
	InstID     int64      `json:"instId"`
	Name       string     `json:"name"`
	Address    string     `json:"address,omitempty"`
	Capacity   int        `json:"capacity"`
	Enabled    bool       `json:"enabled"`
	Remark     string     `json:"remark,omitempty"`
	Sort       int        `json:"sort"`
	CreateTime *time.Time `json:"createTime,omitempty"`
	UpdateTime *time.Time `json:"updateTime,omitempty"`
}

type ClassroomMutation struct {
	ID       *int64 `json:"id,omitempty"`
	UUID     string `json:"uuid,omitempty"`
	Version  *int64 `json:"version,omitempty"`
	Name     string `json:"name"`
	Address  string `json:"address,omitempty"`
	Capacity int    `json:"capacity"`
	Enabled  *bool  `json:"enabled,omitempty"`
	Remark   string `json:"remark,omitempty"`
	Sort     int    `json:"sort"`
}

type ClassroomStatusMutation struct {
	ID      *int64 `json:"id,omitempty"`
	Enabled *bool  `json:"enabled,omitempty"`
}

type ClassroomDeleteDTO struct {
	ID *int64 `json:"id,omitempty"`
}
