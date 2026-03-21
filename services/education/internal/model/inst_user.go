package model

import "time"

type InstUserQueryDTO struct {
	PageRequestModel PageRequestModel   `json:"pageRequestModel"`
	QueryModel       InstUserQueryModel `json:"queryModel"`
}

type InstUserQueryModel struct {
	ID              *int64     `json:"id"`
	UserType        *int       `json:"userType"`
	RoleIDs         []int64    `json:"roleIds"`
	Status          *bool      `json:"status"`
	DeptID          *int64     `json:"deptId"`
	CreateTimeBegin *time.Time `json:"createTimeBegin"`
	CreateTimeEnd   *time.Time `json:"createTimeEnd"`
	SearchKey       string     `json:"searchKey"`
}

type InstUserQueryVO struct {
	ID              int64      `json:"id"`
	UUID            string     `json:"uuid,omitempty"`
	Version         int64      `json:"version,omitempty"`
	InstID          int64      `json:"instId"`
	InstName        string     `json:"instName,omitempty"`
	Avatar          string     `json:"avatar,omitempty"`
	NickName        string     `json:"nickName"`
	Mobile          string     `json:"mobile"`
	DepartNames     string     `json:"departNames,omitempty"`
	RoleNum         int        `json:"roleNum"`
	RoleIDs         []int64    `json:"roleIds,omitempty"`
	RoleName        string     `json:"roleName,omitempty"`
	Disabled        bool       `json:"disabled"`
	UserType        *int       `json:"userType,omitempty"`
	CreateTime      *time.Time `json:"createTime,omitempty"`
	IsAdmin         bool       `json:"isAdmin"`
	ActivatedStatus bool       `json:"activatedStatus"`
}

type InstUserSaveDTO struct {
	UserID   *int64  `json:"userId"`
	InstID   *int64  `json:"instId"`
	NickName string  `json:"nickName"`
	Avatar   string  `json:"avatar"`
	Mobile   string  `json:"mobile"`
	DeptIDs  []int64 `json:"deptIds"`
	Admin    *bool   `json:"admin"`
	Sort     *int    `json:"sort"`
	Disabled *bool   `json:"disabled"`
	Username string  `json:"username"`
	RoleIDs  []int64 `json:"roleIds"`
	Password string  `json:"password"`
	UserType *int    `json:"userType"`
}

type InstUserModifyDTO struct {
	ID       int64   `json:"id"`
	NickName string  `json:"nickName"`
	Avatar   string  `json:"avatar"`
	Mobile   string  `json:"mobile"`
	DeptIDs  []int64 `json:"deptIds"`
	Disabled *bool   `json:"disabled"`
	RoleIDs  []int64 `json:"roleIds"`
	UserType *int    `json:"userType"`
}

type ChangePhoneVO struct {
	Mobile   string `json:"mobile"`
	Code     string `json:"code"`
	Password string `json:"password"`
	UserID   int64  `json:"userId"`
}

type InstUserRoleDetail struct {
	RoleID                   int64  `json:"roleId"`
	RoleName                 string `json:"roleName"`
	Description              string `json:"description,omitempty"`
	FunctionalAuthorityCount int    `json:"functionalAuthorityCount"`
	DataAuthorityCount       int    `json:"dataAuthorityCount"`
}

type InstUserDetailVO struct {
	ID         int64                `json:"id"`
	UUID       string               `json:"uuid,omitempty"`
	Version    int64                `json:"version,omitempty"`
	NickName   string               `json:"nickName"`
	Avatar     string               `json:"avatar,omitempty"`
	Mobile     string               `json:"mobile"`
	Disabled   bool                 `json:"disabled"`
	CreateTime *time.Time           `json:"createTime,omitempty"`
	InstName   string               `json:"instName,omitempty"`
	InstID     int64                `json:"instId"`
	UserType   *int                 `json:"userType,omitempty"`
	DeptNames  []string             `json:"deptNames,omitempty"`
	DeptIDs    []int64              `json:"deptIds,omitempty"`
	Roles      []InstUserRoleDetail `json:"roles,omitempty"`
	RoleIDs    []int64              `json:"roleIds,omitempty"`
	IsAdmin    bool                 `json:"isAdmin"`
}
