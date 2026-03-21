package model

import "time"

type CourseCategory struct {
	ID      int64  `json:"id"`
	UUID    string `json:"uuid,omitempty"`
	Version int64  `json:"version,omitempty"`
	InstID  int64  `json:"instId"`
	Name    string `json:"name"`
	Sort    int    `json:"sort"`
	Remark  string `json:"remark,omitempty"`
}

type CourseCategoryMutation struct {
	ID      *int64 `json:"id"`
	UUID    string `json:"uuid"`
	Version *int64 `json:"version"`
	Name    string `json:"name"`
	Sort    *int   `json:"sort"`
	Remark  string `json:"remark"`
}

type Course struct {
	ID                 int64     `json:"id"`
	UUID               string    `json:"uuid,omitempty"`
	Version            int64     `json:"version,omitempty"`
	Name               string    `json:"name"`
	CourseCategory     *int64    `json:"courseCategory,omitempty"`
	CourseAttribute    *int      `json:"courseAttribute,omitempty"`
	Type               *int      `json:"type,omitempty"`
	CategoryName       string    `json:"categoryName,omitempty"`
	CourseType         *int      `json:"courseType,omitempty"`
	TeachMethod        *int      `json:"teachMethod,omitempty"`
	SaleStatus         *int      `json:"saleStatus,omitempty"`
	ChargeMethods      string    `json:"chargeMethods,omitempty"`
	HasExperiencePrice bool      `json:"hasExperiencePrice"`
	OnlineSale         bool      `json:"onlineSale"`
	QuoteCount         int       `json:"quoteCount"`
	SaleVolume         int       `json:"saleVolume"`
	IsShowMicoSchool   bool      `json:"isShowMicoSchool"`
	UpdateTime         time.Time `json:"updateTime"`
}

type CourseIDName struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type CourseCategoryQueryDTO struct {
	PageRequestModel PageRequestModel      `json:"pageRequestModel"`
	QueryModel       CourseCategoryFilters `json:"queryModel"`
	SortModel        SortModel             `json:"sortModel"`
}

type CourseCategoryFilters struct {
	CourseCategoryID *int64 `json:"courseCategoryId"`
	SearchKey        string `json:"searchKey"`
}

type CourseQueryDTO struct {
	PageRequestModel PageRequestModel `json:"pageRequestModel"`
	QueryModel       CourseFilters    `json:"queryModel"`
	SortModel        SortModel        `json:"sortModel"`
}

type CourseFilters struct {
	SearchKey            string `json:"searchKey"`
	CourseName           string `json:"courseName"`
	CourseCategory       *int64 `json:"courseCategory"`
	CourseAttribute      *int   `json:"courseAttribute"`
	CommonCourse         []int  `json:"commonCourse"`
	TeachMethod          *int   `json:"teachMethod"`
	ChargeTypes          []int  `json:"chargeTypes"`
	SaleStatus           *bool  `json:"saleStatus"`
	LessonAudition       *bool  `json:"lessonAudition"`
	IsOpenMicroSchoolBuy *bool  `json:"isOpenMicroSchoolBuy"`
	IsShowMicroSchool    *bool  `json:"isShowMicroSchool"`
	Deleted              *bool  `json:"delFlag"`
}

type CourseQuotation struct {
	ID             int64   `json:"id"`
	UUID           string  `json:"uuid,omitempty"`
	Version        int64   `json:"version,omitempty"`
	CourseID       int64   `json:"courseId"`
	LessonModel    *int    `json:"lessonModel,omitempty"`
	Name           string  `json:"name"`
	Unit           *int    `json:"unit,omitempty"`
	Quantity       *int    `json:"quantity,omitempty"`
	Price          float64 `json:"price"`
	LessonAudition bool    `json:"lessonAudition"`
	OnlineSale     bool    `json:"onlineSale"`
	Remark         string  `json:"remark,omitempty"`
}

type CourseEntryInfo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type CourseBuyRule struct {
	EnableBuyLimit          bool              `json:"enableBuyLimit"`
	IsAllowReturningStudent bool              `json:"isAllowReturningStudent"`
	RelateProductIds        []int64           `json:"relateProductIds,omitempty"`
	RelateProductInfos      []CourseEntryInfo `json:"relateProductInfos,omitempty"`
	AllowType               *int              `json:"allowType,omitempty"`
	StudentStatuses         []int             `json:"studentStatuses,omitempty"`
	IsAllowFreshmanStudent  bool              `json:"isAllowFreshmanStudent"`
	LimitOnePer             bool              `json:"limitOnePer"`
}

type CoursePropertyBinding struct {
	CoursePropertyID    int64  `json:"coursePropertyId"`
	PropertyIDName      string `json:"propertyIdName,omitempty"`
	CoursePropertyValue int64  `json:"coursePropertyValue"`
	PropertyValueName   string `json:"propertyValueName,omitempty"`
}

type CourseProductSaveDTO struct {
	ID                      *int64                  `json:"id"`
	UUID                    string                  `json:"uuid"`
	Version                 *int64                  `json:"version"`
	Name                    string                  `json:"name"`
	CourseCategory          *int64                  `json:"courseCategory"`
	CourseAttribute         *int                    `json:"courseAttribute"`
	Type                    *int                    `json:"type"`
	Title                   string                  `json:"title"`
	Images                  string                  `json:"images"`
	Description             string                  `json:"description"`
	IsShowMicoSchool        bool                    `json:"isShowMicoSchool"`
	BuyRule                 CourseBuyRule           `json:"buyRule"`
	SaleStatus              *bool                   `json:"saleStatus"`
	ProductSku              []CourseQuotation       `json:"productSku"`
	CourseType              *int                    `json:"courseType"`
	CourseScope             []int64                 `json:"courseScope"`
	TeachMethod             *int                    `json:"teachMethod"`
	AllowedLessonIDs        []int64                 `json:"allowedLessonIds"`
	SubjectIDs              []int64                 `json:"subjectIds"`
	CourseProductProperties []CoursePropertyBinding `json:"courseProductProperties"`
}

type CourseDetail struct {
	ID                      int64                   `json:"id"`
	UUID                    string                  `json:"uuid,omitempty"`
	Version                 int64                   `json:"version,omitempty"`
	Name                    string                  `json:"name"`
	CourseCategory          *int64                  `json:"courseCategory,omitempty"`
	CourseAttribute         *int                    `json:"courseAttribute,omitempty"`
	Type                    *int                    `json:"type,omitempty"`
	CourseType              *int                    `json:"courseType,omitempty"`
	TeachMethod             *int                    `json:"teachMethod,omitempty"`
	SaleStatus              *int                    `json:"saleStatus,omitempty"`
	Title                   string                  `json:"title,omitempty"`
	Images                  string                  `json:"images,omitempty"`
	Description             string                  `json:"description,omitempty"`
	IsShowMicoSchool        bool                    `json:"isShowMicoSchool"`
	CourseScope             []int64                 `json:"courseScope,omitempty"`
	CourseScopeInfo         []CourseEntryInfo       `json:"courseScopeInfo,omitempty"`
	SubjectIDs              []int64                 `json:"subjectIds,omitempty"`
	ProductSku              []CourseQuotation       `json:"productSku,omitempty"`
	BuyRule                 CourseBuyRule           `json:"buyRule,omitempty"`
	CourseProductProperties []CoursePropertyBinding `json:"courseProductProperties,omitempty"`
}

type ProcessContentQueryVO struct {
	ID                 int64             `json:"id"`
	UUID               string            `json:"uuid,omitempty"`
	Version            int64             `json:"version,omitempty"`
	Name               string            `json:"name"`
	CourseCategory     *int64            `json:"courseCategory,omitempty"`
	CategoryName       string            `json:"categoryName,omitempty"`
	CourseType         *int              `json:"courseType,omitempty"`
	TeachMethod        *int              `json:"teachMethod,omitempty"`
	ChargeMethods      string            `json:"chargeMethods,omitempty"`
	HasExperiencePrice bool              `json:"hasExperiencePrice"`
	SaleStatus         *int              `json:"saleStatus,omitempty"`
	ProductSku         []CourseQuotation `json:"productSku,omitempty"`
}

type CourseProperty struct {
	ID                 int64  `json:"id"`
	UUID               string `json:"uuid,omitempty"`
	Version            int64  `json:"version,omitempty"`
	InstID             int64  `json:"instId,omitempty"`
	Name               string `json:"name"`
	Enable             bool   `json:"enable"`
	EnableOnlineFilter bool   `json:"enableOnlineFilter"`
	Remark             string `json:"remark,omitempty"`
}

type CoursePropertyOption struct {
	ID         int64  `json:"id"`
	UUID       string `json:"uuid,omitempty"`
	Version    int64  `json:"version,omitempty"`
	PropertyID int64  `json:"propertyId"`
	Name       string `json:"name"`
	Sort       int    `json:"sort"`
	Remark     string `json:"remark,omitempty"`
}
