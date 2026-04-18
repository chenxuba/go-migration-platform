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
	ModuleID        int64        `json:"moduleId"`
	UUID            string       `json:"uuid,omitempty"`
	Version         int64        `json:"version,omitempty"`
	ModuleName      string       `json:"moduleName"`
	ModuleType      int          `json:"moduleType"`
	Price           float64      `json:"price"`
	Remark          string       `json:"remark,omitempty"`
	MenuCount       int          `json:"menuCount"`
	OrgCount        int          `json:"orgCount"`
	CreateTime      string       `json:"createTime,omitempty"`
	UpdateTime      string       `json:"updateTime,omitempty"`
	SelectedMenuIDs []int64      `json:"selectedMenuIds,omitempty"`
	MenuIDs         []ModuleMenu `json:"menuIds"`
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
	Remark  string   `json:"remark"`
	MenuIDs []int64  `json:"menuIds"`
}

type ModuleMenu struct {
	MenuID    string       `json:"menuId"`
	MenuName  string       `json:"menuName"`
	IsSelect  bool         `json:"isSelect"`
	Introduce string       `json:"introduce,omitempty"`
	MenuType  int64        `json:"menuType,omitempty"`
	GroupCode string       `json:"groupCode,omitempty"`
	Weight    int64        `json:"weight,omitempty"`
	Children  []ModuleMenu `json:"children,omitempty"`
}

type Module struct {
	ID         int64   `json:"id"`
	Name       string  `json:"name"`
	Type       int     `json:"type"`
	Price      float64 `json:"price"`
	Remark     string  `json:"remark,omitempty"`
	MenuCount  int     `json:"menuCount"`
	OrgCount   int     `json:"orgCount"`
	CreateTime string  `json:"createTime,omitempty"`
	UpdateTime string  `json:"updateTime,omitempty"`
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
	Province         string `json:"province,omitempty"`
	City             string `json:"city,omitempty"`
	Region           string `json:"region,omitempty"`
	Address          string `json:"address,omitempty"`
	Logo             string `json:"logo,omitempty"`
	Enabled          bool   `json:"enabled"`
	Status           int    `json:"status"`
	OpenType         int    `json:"openType"`
	OpenDuration     string `json:"openDuration,omitempty"`
	RegisterTime     string `json:"registerTime,omitempty"`
	ExpireEndTime    string `json:"expireEndTime,omitempty"`
	StaffCount       int    `json:"staffCount"`
	ActiveStaffCount int    `json:"activeStaffCount"`
	AdminCount       int    `json:"adminCount"`
}

type InstitutionProfile struct {
	Description   string   `json:"description,omitempty"`
	BusinessTime  string   `json:"businessTime,omitempty"`
	Video         string   `json:"video,omitempty"`
	GalleryImages []string `json:"galleryImages,omitempty"`
}

type InstitutionDetail struct {
	ID              int64              `json:"id"`
	OrganName       string             `json:"organName"`
	OrganCode       string             `json:"organCode"`
	LoginName       string             `json:"loginName"`
	Mobile          string             `json:"mobile"`
	Principal       string             `json:"principal,omitempty"`
	ProvinceCode    int64              `json:"provinceCode,omitempty"`
	Province        string             `json:"province"`
	CityCode        int64              `json:"cityCode,omitempty"`
	City            string             `json:"city"`
	RegionCode      int64              `json:"regionCode,omitempty"`
	Region          string             `json:"region,omitempty"`
	Address         string             `json:"address,omitempty"`
	ConcatPhone     string             `json:"concatPhone,omitempty"`
	FixedPhone      string             `json:"fixedPhone,omitempty"`
	Remark          string             `json:"remark,omitempty"`
	Logo            string             `json:"logo,omitempty"`
	Enabled         bool               `json:"enabled"`
	Status          int                `json:"status"`
	OpenType        int                `json:"openType"`
	OpenDuration    string             `json:"openDuration,omitempty"`
	ExpireStartTime string             `json:"expireStartTime,omitempty"`
	ExpireEndTime   string             `json:"expireEndTime,omitempty"`
	Lng             float64            `json:"lng,omitempty"`
	Lat             float64            `json:"lat,omitempty"`
	Profile         InstitutionProfile `json:"profile"`
}

type InstitutionMutation struct {
	ID           *int64              `json:"id"`
	OrganName    string              `json:"organName"`
	LoginName    string              `json:"loginName"`
	Mobile       string              `json:"mobile"`
	Principal    string              `json:"principal"`
	ProvinceCode *int64              `json:"provinceCode,omitempty"`
	Province     string              `json:"province"`
	CityCode     *int64              `json:"cityCode,omitempty"`
	City         string              `json:"city"`
	RegionCode   *int64              `json:"regionCode,omitempty"`
	Region       string              `json:"region"`
	Address      string              `json:"address"`
	ConcatPhone  string              `json:"concatPhone"`
	FixedPhone   string              `json:"fixedPhone"`
	Remark       string              `json:"remark"`
	Logo         string              `json:"logo"`
	Enabled      *bool               `json:"enabled"`
	OpenType     *int                `json:"openType,omitempty"`
	OpenDuration string              `json:"openDuration"`
	Lng          *float64            `json:"lng,omitempty"`
	Lat          *float64            `json:"lat,omitempty"`
	Profile      *InstitutionProfile `json:"profile,omitempty"`
}

type InstitutionStatusMutation struct {
	ID      *int64 `json:"id"`
	Enabled *bool  `json:"enabled"`
}

type InstitutionGeocodeQuery struct {
	Province string `json:"province"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Address  string `json:"address"`
}

type InstitutionGeocodeResult struct {
	Lng             float64 `json:"lng"`
	Lat             float64 `json:"lat"`
	Source          string  `json:"source"`
	ResolvedAddress string  `json:"resolvedAddress,omitempty"`
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

type InstitutionRenewalRecord struct {
	ID                  int64  `json:"id"`
	InstitutionID       int64  `json:"institutionId"`
	BeforeOpenType      int    `json:"beforeOpenType"`
	BeforeOpenDuration  string `json:"beforeOpenDuration,omitempty"`
	BeforeExpireEndTime string `json:"beforeExpireEndTime,omitempty"`
	AfterOpenType       int    `json:"afterOpenType"`
	RenewDuration       string `json:"renewDuration,omitempty"`
	RenewStartTime      string `json:"renewStartTime,omitempty"`
	AfterExpireEndTime  string `json:"afterExpireEndTime,omitempty"`
	OperatorID          int64  `json:"operatorId,omitempty"`
	CreateTime          string `json:"createTime,omitempty"`
}

type InstitutionRenewalMutation struct {
	InstitutionID *int64 `json:"institutionId"`
	OpenType      *int   `json:"openType"`
	OpenDuration  string `json:"openDuration"`
}

type InstitutionRenewalResult struct {
	InstitutionID   int64  `json:"institutionId"`
	OpenType        int    `json:"openType"`
	OpenDuration    string `json:"openDuration,omitempty"`
	ExpireStartTime string `json:"expireStartTime,omitempty"`
	ExpireEndTime   string `json:"expireEndTime,omitempty"`
}

type InstitutionPermissionDetail struct {
	InstitutionID     int64   `json:"institutionId"`
	OrganName         string  `json:"organName"`
	Mobile            string  `json:"mobile,omitempty"`
	OpenType          int     `json:"openType"`
	OpenDuration      string  `json:"openDuration,omitempty"`
	Status            int     `json:"status"`
	ExpireEndTime     string  `json:"expireEndTime,omitempty"`
	CurrentModuleID   int64   `json:"currentModuleId,omitempty"`
	CurrentModuleName string  `json:"currentModuleName,omitempty"`
	AdminRoleID       int64   `json:"adminRoleId,omitempty"`
	AdminRoleName     string  `json:"adminRoleName,omitempty"`
	TemplateMenuIDs   []int64 `json:"templateMenuIds,omitempty"`
	EffectiveMenuIDs  []int64 `json:"effectiveMenuIds,omitempty"`
}

type InstitutionPermissionMutation struct {
	InstitutionID *int64  `json:"institutionId"`
	ModuleID      *int64  `json:"moduleId,omitempty"`
	MenuIDs       []int64 `json:"menuIds,omitempty"`
}
