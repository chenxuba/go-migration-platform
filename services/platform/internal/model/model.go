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

type ScaleInstitutionRow struct {
	Name      string `json:"name"`
	Contact   string `json:"contact"`
	AuthState string `json:"authState"`
	ExpireAt  string `json:"expireAt"`
}

type ScaleTextResource struct {
	ID      int64  `json:"id"`
	ScaleID int64  `json:"scaleId"`
	Content string `json:"content"`
	Sort    int    `json:"sort"`
}

type ScaleTextResourceMutation struct {
	ID      *int64 `json:"id"`
	ScaleID *int64 `json:"scaleId"`
	Content string `json:"content"`
	Sort    *int   `json:"sort"`
}

type ScaleRecord struct {
	ID                 int64                 `json:"id"`
	Name               string                `json:"name"`
	Code               string                `json:"code"`
	Category           string                `json:"category"`
	Scenario           string                `json:"scenario"`
	AgeRange           string                `json:"ageRange"`
	AgeMinMonths       int                   `json:"ageMinMonths"`
	AgeMaxMonths       int                   `json:"ageMaxMonths"`
	Duration           string                `json:"duration"`
	DurationMinMinutes int                   `json:"durationMinMinutes"`
	DurationMaxMinutes int                   `json:"durationMaxMinutes"`
	CurrentVersion     string                `json:"currentVersion"`
	ItemCount          int                   `json:"itemCount"`
	DomainCount        int                   `json:"domainCount"`
	InstitutionCount   int                   `json:"institutionCount"`
	MonthUsage         int                   `json:"monthUsage"`
	DataStatus         string                `json:"dataStatus"`
	UpdatedAt          string                `json:"updatedAt"`
	Summary            string                `json:"summary"`
	PosterURL          string                `json:"posterUrl"`
	ExecutionEntry     string                `json:"executionEntry"`
	APIPackage         string                `json:"apiPackage"`
	References         []ScaleTextResource   `json:"references"`
	Acknowledgements   []ScaleTextResource   `json:"acknowledgements"`
	AuthInstitutions   []ScaleInstitutionRow `json:"authInstitutions"`
}

type ScaleMutation struct {
	ID             *int64 `json:"id"`
	Name           string `json:"name"`
	Code           string `json:"code"`
	Category       string `json:"category"`
	Scenario       string `json:"scenario"`
	AgeRange       string `json:"ageRange"`
	AgeMinMonths   *int   `json:"ageMinMonths"`
	AgeMaxMonths   *int   `json:"ageMaxMonths"`
	CurrentVersion string `json:"currentVersion"`
	ItemCount      *int   `json:"itemCount"`
	DomainCount    *int   `json:"domainCount"`
	Summary        string `json:"summary"`
	PosterURL      string `json:"posterUrl"`
	ExecutionEntry string `json:"executionEntry"`
	APIPackage     string `json:"apiPackage"`
}

type ScaleQuestionBank struct {
	ScaleCode    string                    `json:"scaleCode"`
	ScaleVersion string                    `json:"scaleVersion"`
	DataStatus   string                    `json:"dataStatus"`
	ItemCount    int                       `json:"itemCount"`
	DomainCount  int                       `json:"domainCount"`
	Domains      []ScaleQuestionBankDomain `json:"domains"`
	Items        []ScaleQuestionBankItem   `json:"items"`
	SourceTables []string                  `json:"sourceTables"`
}

type ScaleQuestionBankDomain struct {
	ScaleCode            string `json:"scaleCode"`
	ScaleName            string `json:"scaleName"`
	Category             string `json:"category"`
	ItemCount            *int   `json:"itemCount,omitempty"`
	MaxRawScore          *int   `json:"maxRawScore,omitempty"`
	ItemNumbers          []int  `json:"itemNumbers,omitempty"`
	IsDevelopmentSubtest bool   `json:"isDevelopmentSubtest,omitempty"`
	IsBehaviorSubtest    bool   `json:"isBehaviorSubtest,omitempty"`
	IsCaregiverReport    bool   `json:"isCaregiverReport,omitempty"`
	CompositeCode        string `json:"compositeCode,omitempty"`
}

type ScaleQuestionBankItem struct {
	ItemNo          int                            `json:"itemNo"`
	ItemTitle       string                         `json:"itemTitle"`
	TestItem        string                         `json:"testItem"`
	Materials       string                         `json:"materials"`
	Method          string                         `json:"method"`
	Describes       string                         `json:"describes,omitempty"`
	Guidance        string                         `json:"guidance"`
	DomainCode      string                         `json:"domainCode"`
	DomainName      string                         `json:"domainName"`
	Standard        string                         `json:"standard"`
	ScoreOptions    []ScaleQuestionBankScoreOption `json:"scoreOptions"`
	ScoreOptionText string                         `json:"scoreOptionText"`
	RecordFields    []ScaleQuestionBankRecordField `json:"recordFields"`
	SourcePDF       string                         `json:"sourcePdf"`
	SourcePages     []int                          `json:"sourcePages"`
	OCRStatus       string                         `json:"ocrStatus"`
	UpdatedAt       string                         `json:"updatedAt"`
}

type ScaleQuestionBankScoreOption struct {
	Value       int    `json:"value"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

type ScaleQuestionBankRecordField struct {
	Key         string                               `json:"key"`
	Label       string                               `json:"label"`
	FieldType   string                               `json:"fieldType"`
	DisplayType string                               `json:"displayType,omitempty"`
	Required    bool                                 `json:"required,omitempty"`
	Placeholder string                               `json:"placeholder,omitempty"`
	Options     []ScaleQuestionBankRecordFieldOption `json:"options,omitempty"`
}

type ScaleQuestionBankRecordFieldOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type ScaleQuestionBankItemMutation struct {
	ScaleCode       string                         `json:"scaleCode"`
	ScaleVersion    string                         `json:"scaleVersion"`
	ItemNo          int                            `json:"itemNo"`
	ItemTitle       string                         `json:"itemTitle"`
	TestItem        string                         `json:"testItem"`
	Materials       string                         `json:"materials"`
	Method          string                         `json:"method"`
	Describes       string                         `json:"describes,omitempty"`
	Guidance        string                         `json:"guidance"`
	DomainCode      string                         `json:"domainCode"`
	DomainName      string                         `json:"domainName"`
	Standard        string                         `json:"standard"`
	ScoreOptions    []ScaleQuestionBankScoreOption `json:"scoreOptions"`
	ScoreOptionText string                         `json:"scoreOptionText"`
	RecordFields    []ScaleQuestionBankRecordField `json:"recordFields"`
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
	TenantID        string       `json:"tenantId,omitempty"`
	TenantName      string       `json:"tenantName,omitempty"`
	OwnerType       string       `json:"ownerType,omitempty"`
	SourceModuleID  int64        `json:"sourceModuleId,omitempty"`
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
	ID             *int64   `json:"id"`
	TenantID       string   `json:"tenantId,omitempty"`
	OwnerType      string   `json:"ownerType,omitempty"`
	SourceModuleID *int64   `json:"sourceModuleId,omitempty"`
	Name           string   `json:"name"`
	Type           *int     `json:"type"`
	Price          *float64 `json:"price"`
	Remark         string   `json:"remark"`
	MenuIDs        []int64  `json:"menuIds"`
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
	ID             int64   `json:"id"`
	TenantID       string  `json:"tenantId,omitempty"`
	TenantName     string  `json:"tenantName,omitempty"`
	OwnerType      string  `json:"ownerType,omitempty"`
	SourceModuleID int64   `json:"sourceModuleId,omitempty"`
	Name           string  `json:"name"`
	Type           int     `json:"type"`
	Price          float64 `json:"price"`
	Remark         string  `json:"remark,omitempty"`
	MenuCount      int     `json:"menuCount"`
	OrgCount       int     `json:"orgCount"`
	CreateTime     string  `json:"createTime,omitempty"`
	UpdateTime     string  `json:"updateTime,omitempty"`
}

type PageResult[T any] struct {
	Items   []T `json:"items"`
	Total   int `json:"total"`
	Current int `json:"current"`
	Size    int `json:"size"`
}

type TenantLoginBrandConfig struct {
	Template        string `json:"template,omitempty"`
	BrandName       string `json:"brandName,omitempty"`
	LogoURL         string `json:"logoUrl,omitempty"`
	LoginTitle      string `json:"loginTitle,omitempty"`
	LoginSubtitle   string `json:"loginSubtitle,omitempty"`
	BackgroundURL   string `json:"backgroundUrl,omitempty"`
	PrimaryColor    string `json:"primaryColor,omitempty"`
	Copyright       string `json:"copyright,omitempty"`
	HeroBadge       string `json:"heroBadge,omitempty"`
	HeroTitle       string `json:"heroTitle,omitempty"`
	HeroDescription string `json:"heroDescription,omitempty"`
}

type TenantLoginBrandSet struct {
	PlatformAdmin    TenantLoginBrandConfig `json:"platformAdmin,omitempty"`
	InstitutionAdmin TenantLoginBrandConfig `json:"institutionAdmin,omitempty"`
}

type TenantPublicLoginTheme struct {
	TenantID        string                 `json:"tenantId"`
	TenantName      string                 `json:"tenantName"`
	EntryType       string                 `json:"entryType"`
	InstitutionID   int64                  `json:"institutionId,omitempty"`
	InstitutionName string                 `json:"institutionName,omitempty"`
	LoginBrand      TenantLoginBrandConfig `json:"loginBrand"`
	MatchedBy       string                 `json:"matchedBy,omitempty"`
}

type TenantBootstrapSummary struct {
	TenantID              string                 `json:"tenantId"`
	TenantName            string                 `json:"tenantName"`
	TenantType            string                 `json:"tenantType"`
	Edition               string                 `json:"edition,omitempty"`
	Status                string                 `json:"status,omitempty"`
	IsolationMode         string                 `json:"isolationMode,omitempty"`
	InstitutionCount      int                    `json:"institutionCount"`
	InstitutionIDs        []int64                `json:"institutionIds,omitempty"`
	MenuCount             int                    `json:"menuCount"`
	ModuleCount           int                    `json:"moduleCount"`
	ModuleIDs             []int64                `json:"moduleIds,omitempty"`
	ModuleNames           []string               `json:"moduleNames,omitempty"`
	AdminUsernames        []string               `json:"adminUsernames"`
	Domains               []string               `json:"domains"`
	AdminDomains          []string               `json:"adminDomains,omitempty"`
	InstitutionDomains    []string               `json:"institutionDomains,omitempty"`
	LoginBrand            TenantLoginBrandConfig `json:"loginBrand,omitempty"`
	PlatformLoginBrand    TenantLoginBrandConfig `json:"platformLoginBrand,omitempty"`
	InstitutionLoginBrand TenantLoginBrandConfig `json:"institutionLoginBrand,omitempty"`
}

type TenantListItem = TenantBootstrapSummary

type TenantMutation struct {
	TenantID              string                 `json:"tenantId"`
	TenantName            string                 `json:"tenantName"`
	TenantType            string                 `json:"tenantType,omitempty"`
	Edition               string                 `json:"edition,omitempty"`
	Status                string                 `json:"status,omitempty"`
	IsolationMode         string                 `json:"isolationMode,omitempty"`
	Domains               []string               `json:"domains,omitempty"`
	AdminDomains          []string               `json:"adminDomains,omitempty"`
	InstitutionDomains    []string               `json:"institutionDomains,omitempty"`
	InstitutionIDs        []int64                `json:"institutionIds,omitempty"`
	MenuIDs               []int64                `json:"menuIds,omitempty"`
	ModuleIDs             []int64                `json:"moduleIds,omitempty"`
	AdminUsername         string                 `json:"adminUsername,omitempty"`
	AdminPassword         string                 `json:"adminPassword,omitempty"`
	AdminNickName         string                 `json:"adminNickName,omitempty"`
	AdminMobile           string                 `json:"adminMobile,omitempty"`
	LoginBrand            TenantLoginBrandConfig `json:"loginBrand,omitempty"`
	PlatformLoginBrand    TenantLoginBrandConfig `json:"platformLoginBrand,omitempty"`
	InstitutionLoginBrand TenantLoginBrandConfig `json:"institutionLoginBrand,omitempty"`
	Remark                string                 `json:"remark,omitempty"`
}

type Institution struct {
	ID                int64  `json:"id"`
	OrganName         string `json:"organName"`
	OrganCode         string `json:"organCode,omitempty"`
	LoginName         string `json:"loginName,omitempty"`
	Mobile            string `json:"mobile,omitempty"`
	Principal         string `json:"principal,omitempty"`
	Province          string `json:"province,omitempty"`
	City              string `json:"city,omitempty"`
	Region            string `json:"region,omitempty"`
	Address           string `json:"address,omitempty"`
	TenantID          string `json:"tenantId,omitempty"`
	TenantName        string `json:"tenantName,omitempty"`
	Logo              string `json:"logo,omitempty"`
	Enabled           bool   `json:"enabled"`
	Status            int    `json:"status"`
	OpenType          int    `json:"openType"`
	OpenDuration      string `json:"openDuration,omitempty"`
	CurrentModuleID   int64  `json:"currentModuleId,omitempty"`
	CurrentModuleName string `json:"currentModuleName,omitempty"`
	RegisterTime      string `json:"registerTime,omitempty"`
	ExpireEndTime     string `json:"expireEndTime,omitempty"`
	StaffCount        int    `json:"staffCount"`
	ActiveStaffCount  int    `json:"activeStaffCount"`
	AdminCount        int    `json:"adminCount"`
}

type InstitutionProfile struct {
	Description   string                 `json:"description,omitempty"`
	BusinessTime  string                 `json:"businessTime,omitempty"`
	Video         string                 `json:"video,omitempty"`
	GalleryImages []string               `json:"galleryImages,omitempty"`
	LoginSlug     string                 `json:"loginSlug,omitempty"`
	LoginBrand    TenantLoginBrandConfig `json:"loginBrand,omitempty"`
}

type InstitutionDetail struct {
	ID                int64              `json:"id"`
	OrganName         string             `json:"organName"`
	OrganCode         string             `json:"organCode"`
	LoginName         string             `json:"loginName"`
	Mobile            string             `json:"mobile"`
	Principal         string             `json:"principal,omitempty"`
	ProvinceCode      int64              `json:"provinceCode,omitempty"`
	Province          string             `json:"province"`
	CityCode          int64              `json:"cityCode,omitempty"`
	City              string             `json:"city"`
	RegionCode        int64              `json:"regionCode,omitempty"`
	Region            string             `json:"region,omitempty"`
	Address           string             `json:"address,omitempty"`
	ConcatPhone       string             `json:"concatPhone,omitempty"`
	FixedPhone        string             `json:"fixedPhone,omitempty"`
	Remark            string             `json:"remark,omitempty"`
	Logo              string             `json:"logo,omitempty"`
	Enabled           bool               `json:"enabled"`
	Status            int                `json:"status"`
	OpenType          int                `json:"openType"`
	OpenDuration      string             `json:"openDuration,omitempty"`
	CurrentModuleID   int64              `json:"currentModuleId,omitempty"`
	CurrentModuleName string             `json:"currentModuleName,omitempty"`
	RegisterTime      string             `json:"registerTime,omitempty"`
	ExpireStartTime   string             `json:"expireStartTime,omitempty"`
	ExpireEndTime     string             `json:"expireEndTime,omitempty"`
	Lng               float64            `json:"lng,omitempty"`
	Lat               float64            `json:"lat,omitempty"`
	Profile           InstitutionProfile `json:"profile"`
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
	ModuleID     *int64              `json:"moduleId,omitempty"`
	OpenDuration string              `json:"openDuration"`
	Lng          *float64            `json:"lng,omitempty"`
	Lat          *float64            `json:"lat,omitempty"`
	Profile      *InstitutionProfile `json:"profile,omitempty"`
}

type InstitutionLoginNameAvailability struct {
	LoginName string `json:"loginName"`
	Available bool   `json:"available"`
	Message   string `json:"message,omitempty"`
}

// TenantAdminUsernameAvailability is the frontend pre-check result for tenant admin login account.
type TenantAdminUsernameAvailability struct {
	Username  string `json:"username"`
	Available bool   `json:"available"`
	Message   string `json:"message,omitempty"`
}

// TenantIDAvailability is the frontend pre-check result for tenant identifier.
type TenantIDAvailability struct {
	TenantID  string `json:"tenantId"`
	Available bool   `json:"available"`
	Message   string `json:"message,omitempty"`
}

type LoginTemplate struct {
	ID             int64    `json:"id"`
	TemplateKey    string   `json:"templateKey"`
	TemplateName   string   `json:"templateName"`
	EntryType      string   `json:"entryType"`
	LayoutType     string   `json:"layoutType,omitempty"`
	Description    string   `json:"description,omitempty"`
	PreviewImage   string   `json:"previewImage,omitempty"`
	Enabled        bool     `json:"enabled"`
	Sort           int      `json:"sort"`
	TenantIDs      []string `json:"tenantIds,omitempty"`
	InstitutionIDs []int64  `json:"institutionIds,omitempty"`
	ReferenceCount int      `json:"referenceCount"`
	CreateTime     string   `json:"createTime,omitempty"`
	UpdateTime     string   `json:"updateTime,omitempty"`
}

type LoginTemplateMutation struct {
	ID             *int64   `json:"id,omitempty"`
	TemplateKey    string   `json:"templateKey"`
	TemplateName   string   `json:"templateName"`
	EntryType      string   `json:"entryType"`
	LayoutType     string   `json:"layoutType,omitempty"`
	Description    string   `json:"description,omitempty"`
	PreviewImage   string   `json:"previewImage,omitempty"`
	Enabled        *bool    `json:"enabled,omitempty"`
	Sort           int      `json:"sort"`
	TenantIDs      []string `json:"tenantIds,omitempty"`
	InstitutionIDs []int64  `json:"institutionIds,omitempty"`
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

type GovernmentOverview struct {
	Level                  string                    `json:"level"`
	LevelLabel             string                    `json:"levelLabel"`
	ScopeText              string                    `json:"scopeText"`
	ScopeCodeText          string                    `json:"scopeCodeText"`
	ScopeCount             int                       `json:"scopeCount"`
	InstitutionCount       int                       `json:"institutionCount"`
	SubordinateRegionCount int                       `json:"subordinateRegionCount"`
	ReadingStudentCount    int                       `json:"readingStudentCount"`
	OrderCount             int                       `json:"orderCount"`
	RegionalSummary        []GovernmentOverviewEntry `json:"regionalSummary"`
}

type GovernmentOverviewEntry struct {
	RegionCode          string `json:"regionCode"`
	RegionName          string `json:"regionName"`
	LevelLabel          string `json:"levelLabel"`
	InstitutionCount    int    `json:"institutionCount"`
	ReadingStudentCount int    `json:"readingStudentCount"`
	IntentStudentCount  int    `json:"intentStudentCount"`
	OrderCount          int    `json:"orderCount"`
}

type GovernmentInstitution struct {
	ID                  int64  `json:"id"`
	OrganName           string `json:"organName"`
	OrganCode           string `json:"organCode,omitempty"`
	LoginName           string `json:"loginName,omitempty"`
	Mobile              string `json:"mobile,omitempty"`
	Principal           string `json:"principal,omitempty"`
	Province            string `json:"province,omitempty"`
	City                string `json:"city,omitempty"`
	Region              string `json:"region,omitempty"`
	Address             string `json:"address,omitempty"`
	Enabled             bool   `json:"enabled"`
	Status              int    `json:"status"`
	OpenType            int    `json:"openType"`
	OpenDuration        string `json:"openDuration,omitempty"`
	RegisterTime        string `json:"registerTime,omitempty"`
	ExpireEndTime       string `json:"expireEndTime,omitempty"`
	StaffCount          int    `json:"staffCount"`
	ActiveStaffCount    int    `json:"activeStaffCount"`
	AdminCount          int    `json:"adminCount"`
	ReadingStudentCount int    `json:"readingStudentCount"`
	IntentStudentCount  int    `json:"intentStudentCount"`
	OrderCount          int    `json:"orderCount"`
}

type GovernmentInstitutionSummary struct {
	TotalCount          int `json:"totalCount"`
	EnabledCount        int `json:"enabledCount"`
	WarningCount        int `json:"warningCount"`
	DisabledCount       int `json:"disabledCount"`
	ExpiredCount        int `json:"expiredCount"`
	ReadingStudentCount int `json:"readingStudentCount"`
	IntentStudentCount  int `json:"intentStudentCount"`
	OrderCount          int `json:"orderCount"`
}

type GovernmentInstitutionPage struct {
	Items         []GovernmentInstitution       `json:"items"`
	Total         int                           `json:"total"`
	Current       int                           `json:"current"`
	Size          int                           `json:"size"`
	Level         string                        `json:"level"`
	LevelLabel    string                        `json:"levelLabel"`
	ScopeText     string                        `json:"scopeText"`
	ScopeCodeText string                        `json:"scopeCodeText"`
	ScopeCount    int                           `json:"scopeCount"`
	Summary       *GovernmentInstitutionSummary `json:"summary,omitempty"`
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
	OperatorName        string `json:"operatorName,omitempty"`
	IsTenantOperator    bool   `json:"isTenantOperator"`
	CreateTime          string `json:"createTime,omitempty"`
}

type InstitutionRenewalMutation struct {
	InstitutionID       *int64 `json:"institutionId"`
	OpenType            *int   `json:"openType,omitempty"`
	ModuleID            *int64 `json:"moduleId,omitempty"`
	OpenDuration        string `json:"openDuration"`
	CustomExpireEndTime string `json:"customExpireEndTime,omitempty"`
}

type InstitutionRenewalResult struct {
	InstitutionID   int64  `json:"institutionId"`
	OpenType        int    `json:"openType"`
	ModuleID        int64  `json:"moduleId,omitempty"`
	ModuleName      string `json:"moduleName,omitempty"`
	OpenDuration    string `json:"openDuration,omitempty"`
	ExpireStartTime string `json:"expireStartTime,omitempty"`
	ExpireEndTime   string `json:"expireEndTime,omitempty"`
}

type InstitutionVersionChangeRecord struct {
	ID                int64  `json:"id"`
	InstitutionID     int64  `json:"institutionId"`
	BeforeOpenType    int    `json:"beforeOpenType"`
	BeforeModuleID    int64  `json:"beforeModuleId,omitempty"`
	BeforeVersionName string `json:"beforeVersionName,omitempty"`
	AfterOpenType     int    `json:"afterOpenType"`
	AfterModuleID     int64  `json:"afterModuleId,omitempty"`
	AfterVersionName  string `json:"afterVersionName,omitempty"`
	OperatorID        int64  `json:"operatorId,omitempty"`
	OperatorName      string `json:"operatorName,omitempty"`
	IsTenantOperator  bool   `json:"isTenantOperator"`
	CreateTime        string `json:"createTime,omitempty"`
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

type InstitutionPermissionBatchMutation struct {
	InstitutionIDs []int64 `json:"institutionIds"`
	ModuleID       *int64  `json:"moduleId,omitempty"`
	MenuIDs        []int64 `json:"menuIds,omitempty"`
}
