package model

import "time"

type LoginRequest struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	LoginType     int    `json:"loginType"`
	InstitutionID *int64 `json:"institutionId,omitempty"`
	UserID        *int64 `json:"userId,omitempty"`
}

type InstitutionLoginOptionsRequest struct {
	Identifier string `json:"identifier"`
	LoginType  int    `json:"loginType"`
}

type InstitutionLoginOption struct {
	UserID              int64  `json:"userId"`
	InstID              int64  `json:"instId"`
	OrgName             string `json:"orgName"`
	LoginName           string `json:"loginName"`
	NickName            string `json:"nickName"`
	Mobile              string `json:"mobile"`
	Logo                string `json:"logo,omitempty"`
	Admin               bool   `json:"admin"`
	InstitutionStatus   string `json:"institutionStatus,omitempty"`
	InstitutionReadonly bool   `json:"institutionReadonly"`
}

type User struct {
	ID       int64
	Username string
	Password string
	Mobile   string
	NickName string
	UserType *int
	DeptID   *int64
	IsAdmin  bool
}

type ManageUserInfo struct {
	ID           int64    `json:"id"`
	Username     string   `json:"username"`
	Mobile       string   `json:"mobile"`
	NickName     string   `json:"nickName"`
	DeptID       *int64   `json:"deptId,omitempty"`
	DeptIDs      []int64  `json:"deptIds"`
	DeptName     string   `json:"deptName,omitempty"`
	IsAdmin      bool     `json:"isAdmin"`
	MenuCodeList []string `json:"menuCodeList"`
	RoleID       string   `json:"roleId,omitempty"`
	RoleName     string   `json:"roleName,omitempty"`
}

type InstUserInfo struct {
	InstUserID           int64    `json:"instUserId"`
	UserID               int64    `json:"userId"`
	InstID               int64    `json:"instId"`
	OpenType             int      `json:"openType,omitempty"`
	VersionName          string   `json:"versionName,omitempty"`
	NickName             string   `json:"nickName"`
	Avatar               string   `json:"avatar,omitempty"`
	OrgName              string   `json:"orgName"`
	Username             string   `json:"username,omitempty"`
	Mobile               string   `json:"mobile,omitempty"`
	Logo                 string   `json:"logo,omitempty"`
	Manage               bool     `json:"manage"`
	Admin                bool     `json:"admin"`
	Disabled             bool     `json:"disabled"`
	InstitutionEnabled   bool     `json:"-"`
	InstitutionExpired   bool     `json:"-"`
	InstitutionStatus    string   `json:"institutionStatus,omitempty"`
	InstitutionReadonly  bool     `json:"institutionReadonly"`
	DeptIDs              []int64  `json:"deptIds"`
	MenuCodeList         []string `json:"menuCodeList"`
}

const (
	InstitutionStatusNormal          = "normal"
	InstitutionStatusDisabled        = "disabled"
	InstitutionStatusTrialExpired    = "trial_expired"
	InstitutionStatusExpiredReadonly = "expired_readonly"
)

func ResolveInstitutionStatus(enabled, expired bool, openType int) string {
	if !enabled {
		return InstitutionStatusDisabled
	}
	if expired {
		if openType == 1 {
			return InstitutionStatusTrialExpired
		}
		return InstitutionStatusExpiredReadonly
	}
	return InstitutionStatusNormal
}

func InstitutionStatusMessage(status string) string {
	switch status {
	case InstitutionStatusDisabled:
		return "该机构已被停用"
	case InstitutionStatusTrialExpired:
		return "该机构已到期，请联系售后"
	default:
		return ""
	}
}

type LoginResult struct {
	Token     string `json:"token"`
	LoginType string `json:"loginType"`
	User      any    `json:"user"`
	TenantID  string `json:"tenantId"`
	OrgID     int64  `json:"orgId,omitempty"`
}

type SessionInfo struct {
	UserID       int64    `json:"userId"`
	Username     string   `json:"username"`
	LoginType    string   `json:"loginType"`
	TenantID     string   `json:"tenantId"`
	OrgID        int64    `json:"orgId,omitempty"`
	RoleList     []string `json:"roleList"`
	MenuCodeList []string `json:"menuCodeList"`
	User         any      `json:"user"`
}

type UserPageItem struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Mobile   string `json:"mobile"`
	NickName string `json:"nickName"`
	DeptName string `json:"deptName,omitempty"`
	RoleID   string `json:"roleId,omitempty"`
	RoleName string `json:"roleName,omitempty"`
}

type UserPage struct {
	Items   []UserPageItem `json:"items"`
	Total   int            `json:"total"`
	Current int            `json:"current"`
	Size    int            `json:"size"`
}

type LoginLogSearchDTO struct {
	UserType  *int   `json:"userType"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	NickName  string `json:"nickName"`
	OrgName   string `json:"orgName"`
	Result    *int   `json:"result"`
}

type LoginLogItem struct {
	ID         int64     `json:"id"`
	UserID     *int64    `json:"userId,omitempty"`
	NickName   string    `json:"nickName,omitempty"`
	UserType   *int      `json:"userType,omitempty"`
	UserIP     string    `json:"userIp,omitempty"`
	UserAgent  string    `json:"userAgent,omitempty"`
	OrgID      *int64    `json:"orgId,omitempty"`
	OrgName    string    `json:"orgName,omitempty"`
	Result     *int      `json:"result,omitempty"`
	CreateTime time.Time `json:"createTime"`
}

type LoginLogPage struct {
	Items   []LoginLogItem `json:"items"`
	Total   int            `json:"total"`
	Current int            `json:"current"`
	Size    int            `json:"size"`
}

type Depart struct {
	ID           int64  `json:"id"`
	DepartName   string `json:"departName"`
	DepartCode   string `json:"departCode,omitempty"`
	DepartMan    string `json:"departMan,omitempty"`
	DepartConcat string `json:"departConcat,omitempty"`
	OrgID        int64  `json:"orgId"`
	PID          int64  `json:"pid"`
	Enable       *bool  `json:"enable,omitempty"`
	Sort         *int   `json:"sort,omitempty"`
	Remark       string `json:"remark,omitempty"`
}

func (d Depart) OrgIDPtr() *int64 {
	if d.OrgID <= 0 {
		return nil
	}
	value := d.OrgID
	return &value
}

type DepartTreeNode struct {
	Depart
	PName    string           `json:"pName,omitempty"`
	Children []DepartTreeNode `json:"children"`
}

type Menu struct {
	ID                int64  `json:"id"`
	MenuName          string `json:"menuName"`
	Icon              string `json:"icon,omitempty"`
	MenuCode          string `json:"menuCode,omitempty"`
	MenuType          *int   `json:"menuType,omitempty"`
	OwnType           *int   `json:"ownType,omitempty"`
	PID               int64  `json:"pid"`
	Sort              *int   `json:"sort,omitempty"`
	Weight            *int   `json:"weight,omitempty"`
	GroupCode         string `json:"groupCode,omitempty"`
	Remark            string `json:"remark,omitempty"`
	Introduce         string `json:"introduce,omitempty"`
	AccessDeniedImage string `json:"accessDeniedImage,omitempty"`
	Level             *int   `json:"level,omitempty"`
}

type MenuTreeNode struct {
	Menu
	Children []MenuTreeNode `json:"children"`
}

type MenuAccessCheck struct {
	MenuCode string `json:"menuCode"`
	Allowed  bool   `json:"allowed"`
}

type RoleQueryDTO struct {
	PageRequestModel PageRequestModel   `json:"pageRequestModel"`
	QueryModel       RoleQueryCondition `json:"queryModel"`
}

type PageRequestModel struct {
	PageIndex int `json:"pageIndex"`
	PageSize  int `json:"pageSize"`
}

type RoleQueryCondition struct {
	RoleID          *int64 `json:"roleId"`
	UpdateTimeBegin string `json:"updateTimeBegin"`
	UpdateTimeEnd   string `json:"updateTimeEnd"`
	SearchKey       string `json:"searchKey"`
}

type RoleQueryVO struct {
	ID                       int64    `json:"id"`
	UUID                     string   `json:"uuid,omitempty"`
	Version                  int64    `json:"version,omitempty"`
	RoleName                 string   `json:"roleName"`
	Sort                     *int     `json:"sort,omitempty"`
	RoleType                 *int     `json:"roleType,omitempty"`
	OrgID                    *int64   `json:"orgId,omitempty"`
	Admin                    *bool    `json:"admin,omitempty"`
	IsDefault                *bool    `json:"isDefault,omitempty"`
	FunctionalAuthorityCount int      `json:"functionalAuthorityCount"`
	DataAuthorityCount       int      `json:"dataAuthorityCount"`
	Description              string   `json:"description,omitempty"`
	StaffCount               int      `json:"staffCount"`
	StaffNames               []string `json:"staffNames,omitempty"`
	MenuIDs                  []int64  `json:"menuIds,omitempty"`
	UpdateName               string   `json:"updateName,omitempty"`
	CreateName               string   `json:"createName,omitempty"`
}

type RolePage struct {
	Items   []RoleQueryVO `json:"items"`
	Total   int           `json:"total"`
	Current int           `json:"current"`
	Size    int           `json:"size"`
}

type Role struct {
	ID          int64  `json:"id"`
	UUID        string `json:"uuid,omitempty"`
	Version     int64  `json:"version,omitempty"`
	RoleName    string `json:"roleName"`
	Description string `json:"description"`
	OrgID       int64  `json:"orgId"`
	RoleType    int    `json:"roleType"`
	Admin       bool   `json:"admin"`
	IsDefault   bool   `json:"isDefault"`
}

type SaveRoleRequest struct {
	RoleID       *int64  `json:"roleId,omitempty"`
	RoleName     string  `json:"roleName"`
	Description  string  `json:"description"`
	RoleTemplate []int64 `json:"roleTemplate,omitempty"`
	MenuIDs      []int64 `json:"menuIds,omitempty"`
}

type SaveDefaultRoleRequest struct {
	RoleName    string  `json:"roleName"`
	Description string  `json:"description"`
	MenuIDs     []int64 `json:"menuIds,omitempty"`
	RoleType    *int    `json:"roleType,omitempty"`
}

type DeleteDefaultRoleRequest struct {
	RoleID int64 `json:"roleId"`
}

type DeleteDefaultRoleResult struct {
	DetachedUsers int `json:"detachedUsers"`
}

type InstMenuListRequest struct {
	RoleType *int   `json:"roleType,omitempty"`
	InstID   *int64 `json:"instId,omitempty"`
	RoleID   *int64 `json:"roleId,omitempty"`
}

type RoleMenuCompareRequest struct {
	RoleIDs []int64 `json:"roleIds,omitempty"`
	MenuIDs []int64 `json:"menuIds,omitempty"`
}

type MenuTreeVO struct {
	MenuID    int64        `json:"menuId"`
	PID       int64        `json:"pid"`
	MenuName  string       `json:"menuName"`
	Checked   bool         `json:"checked"`
	Introduce string       `json:"introduce,omitempty"`
	Level     *int         `json:"level,omitempty"`
	Children  []MenuTreeVO `json:"children,omitempty"`
}

type RoleTemplateVO struct {
	RoleID                   int64   `json:"roleId"`
	UUID                     string  `json:"uuid,omitempty"`
	Version                  int64   `json:"version,omitempty"`
	RoleName                 string  `json:"roleName"`
	IsDefault                *bool   `json:"isDefault,omitempty"`
	RoleIDs                  []int64 `json:"roleIds,omitempty"`
	FunctionalAuthorityCount int     `json:"functionalAuthorityCount,omitempty"`
	DataAuthorityCount       int     `json:"dataAuthorityCount,omitempty"`
}

type DefaultRoleDetailVO struct {
	RoleID      int64          `json:"roleId"`
	UUID        string         `json:"uuid,omitempty"`
	Version     int64          `json:"version,omitempty"`
	RoleName    string         `json:"roleName"`
	Description string         `json:"description,omitempty"`
	IsAdmin     *bool          `json:"isAdmin,omitempty"`
	IsDefault   *bool          `json:"isDefault,omitempty"`
	UpdateName  string         `json:"updateName,omitempty"`
	MenuIDs     []MenuTreeNode `json:"menuIds"`
}

type InstUserSimple struct {
	ID              int64      `json:"id"`
	UserID          *int64     `json:"userId,omitempty"`
	NickName        string     `json:"nickName"`
	Mobile          string     `json:"mobile,omitempty"`
	Disabled        bool       `json:"disabled"`
	CreateTime      *time.Time `json:"createTime,omitempty"`
	ActivatedStatus bool       `json:"activatedStatus"`
	IsAdmin         bool       `json:"isAdmin"`
	RoleName        string     `json:"roleName,omitempty"`
}
