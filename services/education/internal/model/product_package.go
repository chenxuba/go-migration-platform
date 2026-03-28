package model

import "time"

type ProductPackageQueryDTO struct {
	PageRequestModel PageRequestModel          `json:"pageRequestModel"`
	QueryModel       ProductPackageQueryFilter `json:"queryModel"`
	SortModel        map[string]any            `json:"sortModel"`
}

type ProductPackageQueryFilter struct {
	Name                     string                      `json:"name"`
	SearchKey                string                      `json:"searchKey"`
	OnlineSale               *bool                       `json:"onlineSale"`
	IsOnlineSaleMicoSchool   *bool                       `json:"isOnlineSaleMicoSchool"`
	IsShowMicoSchool         *bool                       `json:"isShowMicoSchool"`
	ProductPackageProperties []ProductPackagePropertyRef `json:"productPackageProperties"`
}

type ProductPackagePropertyRef struct {
	ProductPackagePropertyID    string `json:"productPackagePropertyId"`
	ProductPackagePropertyValue string `json:"productPackagePropertyValue"`
}

type ProductPackageItemMutation struct {
	ProductID      string  `json:"productId"`
	SkuID          string  `json:"skuId"`
	SkuCount       float64 `json:"skuCount"`
	FreeQuantity   float64 `json:"freeQuantity"`
	DiscountType   *int    `json:"discountType,omitempty"`
	DiscountNumber float64 `json:"discountNumber"`
}

type ProductPackageMutation struct {
	ID                       string                       `json:"id,omitempty"`
	Name                     string                       `json:"name"`
	OnlineSale               bool                         `json:"onlineSale"`
	IsAllowEditWhenEnroll    bool                         `json:"isAllowEditWhenEnroll"`
	Title                    string                       `json:"title"`
	Images                   string                       `json:"images"`
	Description              string                       `json:"description"`
	IsShowMicoSchool         bool                         `json:"isShowMicoSchool"`
	IsOnlineSaleMicoSchool   bool                         `json:"isOnlineSaleMicoSchool"`
	BuyRule                  map[string]any               `json:"buyRule"`
	Items                    []ProductPackageItemMutation `json:"items"`
	SubjectIDs               []int64                      `json:"subjectIds"`
	ProductPackageProperties []ProductPackagePropertyRef  `json:"productPackageProperties"`
}

type ProductPackageOperateMutation struct {
	ID                      string `json:"id"`
	OnlineSale              *bool  `json:"onlineSale,omitempty"`
	IsShowMicoSchool        *bool  `json:"isShowMicoSchool,omitempty"`
	IsOnlineSaleMicoSchool  *bool  `json:"isOnlineSaleMicoSchool,omitempty"`
	IsAllowEditWhenEnroll   *bool  `json:"isAllowEditWhenEnroll,omitempty"`
}

type ProductPackagePropertyVO struct {
	ProductPackagePropertyID        string `json:"productPackagePropertyId"`
	ProductPackagePropertyName      string `json:"productPackagePropertyName,omitempty"`
	ProductPackagePropertyValue     string `json:"productPackagePropertyValue"`
	ProductPackagePropertyValueName string `json:"productPackagePropertyValueName,omitempty"`
}

type ProductPackageItemVO struct {
	ID             string `json:"id"`
	ProductType    int    `json:"productType"`
	ProductID      string `json:"productId"`
	ProductName    string `json:"productName"`
	SkuID          string `json:"skuId"`
	SkuName        string `json:"skuName"`
	LessonScope    int    `json:"lessonScope"`
	LessonType     int    `json:"lessonType"`
	LessonMode     int    `json:"lessonMode"`
	LessonAudition bool   `json:"lessonAudition"`
}

type ProductPackageSubjectVO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ProductPackageVO struct {
	ID                      string                     `json:"id"`
	Name                    string                     `json:"name"`
	Title                   string                     `json:"title"`
	OnlineSale              bool                       `json:"onlineSale"`
	IsOnlineSaleMicoSchool  bool                       `json:"isOnlineSaleMicoSchool"`
	IsShowMicoSchool        bool                       `json:"isShowMicoSchool"`
	OrgProductPackageID     string                     `json:"orgProductPackageId"`
	Editable                bool                       `json:"editable"`
	IsSyncOrgProductPackage bool                       `json:"isSyncOrgProductPackage"`
	Sale                    int                        `json:"sale"`
	TotalAmount             float64                    `json:"totalAmount"`
	DiscountAmount          float64                    `json:"discountAmount"`
	FinalAmount             float64                    `json:"finalAmount"`
	Images                  string                     `json:"images"`
	Subjects                []ProductPackageSubjectVO  `json:"subjects"`
	ExtendProperties        []ProductPackagePropertyVO `json:"extendProperties"`
	UpdatedTime             *time.Time                 `json:"updatedTime,omitempty"`
	Items                   []ProductPackageItemVO     `json:"items"`
}

type ProductPackagePagedResult struct {
	List  []ProductPackageVO `json:"list"`
	Total int                `json:"total"`
}

type ProductPackageStatistics struct {
	TotalCount  int `json:"totalCount"`
	OnSaleCount int `json:"onSaleCount"`
}
