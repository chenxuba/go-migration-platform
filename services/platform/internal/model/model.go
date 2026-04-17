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

type Institution struct {
	ID               int64  `json:"id"`
	OrganName        string `json:"organName"`
	OrganCode        string `json:"organCode,omitempty"`
	LoginName        string `json:"loginName,omitempty"`
	Mobile           string `json:"mobile,omitempty"`
	Principal        string `json:"principal,omitempty"`
	Address          string `json:"address,omitempty"`
	Logo             string `json:"logo,omitempty"`
	Enabled          bool   `json:"enabled"`
	StaffCount       int    `json:"staffCount"`
	ActiveStaffCount int    `json:"activeStaffCount"`
	AdminCount       int    `json:"adminCount"`
}

type InstitutionDetail struct {
	ID          int64  `json:"id"`
	OrganName   string `json:"organName"`
	OrganCode   string `json:"organCode"`
	LoginName   string `json:"loginName"`
	Mobile      string `json:"mobile"`
	Principal   string `json:"principal,omitempty"`
	Province    string `json:"province"`
	City        string `json:"city"`
	Region      string `json:"region,omitempty"`
	Address     string `json:"address,omitempty"`
	ConcatPhone string `json:"concatPhone,omitempty"`
	FixedPhone  string `json:"fixedPhone,omitempty"`
	Remark      string `json:"remark,omitempty"`
	Logo        string `json:"logo,omitempty"`
	Enabled     bool   `json:"enabled"`
	Status      int    `json:"status"`
}

type InstitutionMutation struct {
	ID          *int64 `json:"id"`
	OrganName   string `json:"organName"`
	LoginName   string `json:"loginName"`
	Mobile      string `json:"mobile"`
	Principal   string `json:"principal"`
	Province    string `json:"province"`
	City        string `json:"city"`
	Region      string `json:"region"`
	Address     string `json:"address"`
	ConcatPhone string `json:"concatPhone"`
	FixedPhone  string `json:"fixedPhone"`
	Remark      string `json:"remark"`
	Logo        string `json:"logo"`
	Enabled     *bool  `json:"enabled"`
}

type InstitutionStatusMutation struct {
	ID      *int64 `json:"id"`
	Enabled *bool  `json:"enabled"`
}

type InstitutionSummary struct {
	TotalCount    int `json:"totalCount"`
	EnabledCount  int `json:"enabledCount"`
	DisabledCount int `json:"disabledCount"`
}

type InstitutionPage struct {
	Items   []Institution       `json:"items"`
	Total   int                 `json:"total"`
	Current int                 `json:"current"`
	Size    int                 `json:"size"`
	Summary *InstitutionSummary `json:"summary,omitempty"`
}
