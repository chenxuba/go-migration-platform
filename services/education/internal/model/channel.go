package model

import "time"

type ChannelVO struct {
	ID           int64  `json:"id"`
	UUID         string `json:"uuid,omitempty"`
	Version      int64  `json:"version,omitempty"`
	Name         string `json:"name"`
	Introduction string `json:"introduction,omitempty"`
	CategoryID   int64  `json:"categoryId,omitempty"`
	IsDisabled   bool   `json:"isDisabled"`
	Remark       string `json:"remark,omitempty"`
}

type ChannelCategoryMutation struct {
	ID           *int64 `json:"id"`
	UUID         string `json:"uuid"`
	Version      *int64 `json:"version"`
	CategoryName string `json:"categoryName"`
	Remark       string `json:"remark"`
}

type ChannelStatusMutation struct {
	ID         *int64 `json:"id"`
	IsDisabled *bool  `json:"isDisabled"`
}

type ChannelMutation struct {
	ID           *int64 `json:"id"`
	UUID         string `json:"uuid"`
	Version      *int64 `json:"version"`
	CategoryID   *int64 `json:"categoryId"`
	ChannelName  string `json:"channelName"`
	Introduction string `json:"introduction"`
	Remark       string `json:"remark"`
}

type AdjustChannelDTO struct {
	ChannelIDs []int64 `json:"channelIds"`
	CategoryID *int64  `json:"categoryId"`
}

type CustomChannelVO struct {
	ID          int64       `json:"id"`
	UUID        string      `json:"uuid,omitempty"`
	Version     int64       `json:"version,omitempty"`
	Name        string      `json:"name"`
	IsDisabled  bool        `json:"isDisabled"`
	Type        int         `json:"type"`
	ChannelList []ChannelVO `json:"channelList,omitempty"`
	Remark      string      `json:"remark,omitempty"`
}

type ChannelCategoryVO struct {
	ID           int64  `json:"id"`
	UUID         string `json:"uuid,omitempty"`
	Version      int64  `json:"version,omitempty"`
	InstID       int64  `json:"instId"`
	CategoryName string `json:"categoryName"`
	Remark       string `json:"remark,omitempty"`
	ChannelCount int    `json:"channelCount"`
}

type ChannelTreeVO struct {
	ID          int64       `json:"id"`
	Name        string      `json:"name"`
	IsDisabled  bool        `json:"isDisabled"`
	Type        int         `json:"type"`
	ChannelList []ChannelVO `json:"channelList"`
}

type ChannelPCQueryDTO struct {
	PageRequestModel PageRequestModel    `json:"pageRequestModel"`
	QueryModel       ChannelPCQueryModel `json:"queryModel"`
	SortModel        ChannelPCSortModel  `json:"sortModel"`
}

type ChannelPCQueryModel struct {
	ChannelTypeIDs any   `json:"channelTypeIds"`
	IsDefault      *bool `json:"isDefault"`
	IsDisabled     *bool `json:"isDisabled"`
}

type ChannelPCSortModel struct {
	ByCreatedTime *int `json:"byCreatedTime"`
}

type ChannelPCVO struct {
	ID                 int64      `json:"id"`
	UUID               string     `json:"uuid,omitempty"`
	Version            int64      `json:"version,omitempty"`
	ChannelName        string     `json:"channelName"`
	CategoryID         *int64     `json:"categoryId,omitempty"`
	CategoryName       string     `json:"categoryName,omitempty"`
	IsDisabled         bool       `json:"isDisabled"`
	IsDefault          bool       `json:"isDefault"`
	InvalidCount       int        `json:"invalidCount"`
	DealTransformCount int        `json:"dealTransformCount"`
	DealTransformRate  float64    `json:"dealTransformRate"`
	Remark             string     `json:"remark,omitempty"`
	CreateTime         *time.Time `json:"createTime,omitempty"`
}
