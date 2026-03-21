package model

import "time"

type Dict struct {
	ID       int64  `json:"id"`
	DictName string `json:"dictName"`
	DictCode string `json:"dictCode"`
	IsEnable bool   `json:"isEnable"`
	Remark   string `json:"remark,omitempty"`
}

type DictMutation struct {
	ID       *int64 `json:"id"`
	DictName string `json:"dictName"`
	DictCode string `json:"dictCode"`
	IsEnable *bool  `json:"isEnable"`
	Remark   string `json:"remark"`
}

type DictValue struct {
	ID        int64  `json:"id"`
	DictID    int64  `json:"dictId"`
	DictLabel string `json:"dictLabel"`
	DictValue string `json:"dictValue"`
	Sort      int    `json:"sort"`
	IsEnable  bool   `json:"isEnable"`
}

type DictValueMutation struct {
	ID        *int64 `json:"id"`
	DictID    *int64 `json:"dictId"`
	DictLabel string `json:"dictLabel"`
	DictValue string `json:"dictValue"`
	Sort      *int   `json:"sort"`
	IsEnable  *bool  `json:"isEnable"`
	Remark    string `json:"remark"`
}

type Notice struct {
	ID         int64     `json:"id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	DisableID  int64     `json:"disableId"`
	Compel     bool      `json:"compel"`
	CreateTime time.Time `json:"createTime"`
}

type NoticeQuery struct {
	Current   int
	Size      int
	Title     string
	StartTime string
	EndTime   string
	DisableID int64
}

type NoticeMutation struct {
	ID        *int64 `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	DisableID *int64 `json:"disableId"`
	Compel    *bool  `json:"compel"`
}

type ModuleDetailVO struct {
	ModuleID   int64        `json:"moduleId"`
	UUID       string       `json:"uuid,omitempty"`
	Version    int64        `json:"version,omitempty"`
	ModuleName string       `json:"moduleName"`
	Price      float64      `json:"price"`
	MenuIDs    []ModuleMenu `json:"menuIds"`
}

type ModulePermissionMutation struct {
	ID        *int64  `json:"id"`
	MenuIDs   []int64 `json:"menuIds"`
	IsAllRole *bool   `json:"isAllRole"`
}

type ModuleMutation struct {
	ID      *int64   `json:"id"`
	Name    string   `json:"name"`
	Type    *int     `json:"type"`
	Price   *float64 `json:"price"`
	MenuIDs []int64  `json:"menuIds"`
}

type ModuleMenu struct {
	MenuID    string       `json:"menuId"`
	MenuName  string       `json:"menuName"`
	IsSelect  bool         `json:"isSelect"`
	Introduce string       `json:"introduce,omitempty"`
	Children  []ModuleMenu `json:"children,omitempty"`
}

type Module struct {
	ID    int64   `json:"id"`
	Name  string  `json:"name"`
	Type  int     `json:"type"`
	Price float64 `json:"price"`
}

type PageResult[T any] struct {
	Items   []T `json:"items"`
	Total   int `json:"total"`
	Current int `json:"current"`
	Size    int `json:"size"`
}
