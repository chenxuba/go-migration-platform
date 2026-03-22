package model

import "time"

type StudentSyncStatus struct {
	IndexName      string         `json:"indexName"`
	ES             map[string]any `json:"es"`
	RocketMQ       map[string]any `json:"rocketmq"`
	TotalStudents  int            `json:"totalStudents"`
	IntentStudents int            `json:"intentStudents"`
}

type IntentStudentQueryDTO struct {
	PageRequestModel PageRequestModel     `json:"pageRequestModel"`
	QueryModel       IntentStudentFilters `json:"queryModel"`
	SortModel        SortModel            `json:"sortModel"`
}

type IntentStudentFilters struct {
	QueryAllOrDepartment     *int             `json:"queryAllOrDepartment"`
	QuickFilter              *int             `json:"quickFilter"`
	StudentID                string           `json:"studentId"`
	SalespersonID            *int64           `json:"salespersonId"`
	CourseID                 *int64           `json:"courseId"`
	SearchKey                string           `json:"searchKey"`
	WechatNumber             string           `json:"wechatNumber"`
	SchoolSearchKey          string           `json:"schoolSearchKey"`
	AddressSearchKey         string           `json:"addressSearchKey"`
	InterestSearchKey        string           `json:"interestSearchKey"`
	IntentionLevels          []int            `json:"intentionLevels"`
	FollowUpStatuses         []int            `json:"followUpStatuses"`
	Sexes                    []int            `json:"sexes"`
	Grades                   []string         `json:"grades"`
	ChannelIDs               []int64          `json:"channelIds"`
	RecommendStudentID       *int64           `json:"recommendStudentId"`
	CreateID                 *int64           `json:"createId"`
	IsRecommend              *bool            `json:"isRecommend"`
	IsHasSalePerson          *bool            `json:"isHasSalePerson"`
	PurchasedAuditionProduct *bool            `json:"purchasedAuditionProduct"`
	NotFollowUpDay           *int             `json:"notFollowUpDay"`
	AgeMin                   *int             `json:"ageMin"`
	AgeMax                   *int             `json:"ageMax"`
	CreateTimeBegin          string           `json:"createTimeBegin"`
	CreateTimeEnd            string           `json:"createTimeEnd"`
	BirthDayBegin            string           `json:"birthDayBegin"`
	BirthDayEnd              string           `json:"birthDayEnd"`
	FollowUpTimeBegin        string           `json:"followUpTimeBegin"`
	FollowUpTimeEnd          string           `json:"followUpTimeEnd"`
	NextFollowUpTimeBegin    string           `json:"nextFollowUpTimeBegin"`
	NextFollowUpTimeEnd      string           `json:"nextFollowUpTimeEnd"`
	SalesAssignedTimeBegin   string           `json:"salesAssignedTimeBegin"`
	SalesAssignedTimeEnd     string           `json:"salesAssignedTimeEnd"`
	CustomFieldSearchList    []map[string]any `json:"customFieldSearchList"`
}

type IntentStudent struct {
	ID                            int64          `json:"id"`
	InstID                        int64          `json:"instId"`
	StuName                       string         `json:"stuName"`
	AvatarURL                     string         `json:"avatarUrl,omitempty"`
	StuSex                        *int           `json:"stuSex,omitempty"`
	Mobile                        string         `json:"mobile"`
	PhoneRelationship             *int           `json:"phoneRelationship,omitempty"`
	SalePerson                    *int64         `json:"salePerson,omitempty"`
	SalePersonName                string         `json:"salePersonName"`
	IntentLevel                   *int           `json:"intentLevel,omitempty"`
	IntendedCourse                []int64        `json:"intendedCourse,omitempty"`
	Lessons                       []CourseIDName `json:"lessons"`
	ChannelID                     *int64         `json:"channelId,omitempty"`
	ChannelName                   string         `json:"channelName"`
	CreateTime                    time.Time      `json:"createTime"`
	BirthDay                      *time.Time     `json:"birthDay,omitempty"`
	WeChatNumber                  string         `json:"weChatNumber,omitempty"`
	StudySchool                   string         `json:"studySchool,omitempty"`
	Grade                         string         `json:"grade,omitempty"`
	Interest                      string         `json:"interest,omitempty"`
	Address                       string         `json:"address,omitempty"`
	RecommendStudentID            *int64         `json:"recommendStudentId,omitempty"`
	RecommendStudentName          string         `json:"recommendStudentName"`
	ChannelCategoryName           string         `json:"channelCategoryName"`
	FollowUpStatus                *int           `json:"followUpStatus,omitempty"`
	StudentStatus                 int            `json:"studentStatus"`
	LastFollowUpTime              *time.Time     `json:"followUpTime,omitempty"`
	NextFollowUpTime              *time.Time     `json:"nextFollowUpTime,omitempty"`
	SalesAssignedTime             *time.Time     `json:"salesAssignedTime,omitempty"`
	CreateID                      *int64         `json:"createId,omitempty"`
	CreateName                    string         `json:"createName"`
	IsRecommend                   bool           `json:"isRecommend"`
	PurchasedAuditionProduct      bool           `json:"purchasedAuditionProduct"`
	ExperienceClassPurchaseStatus string         `json:"experienceClassPurchaseStatus"`
	DaysUntilReturn               *int           `json:"daysUntilReturn,omitempty"`
	CustomInfo                    []CustomInfo   `json:"customInfo"`
	Remark                        string         `json:"remark,omitempty"`
}

type CustomInfo struct {
	FieldID   int64  `json:"fieldId"`
	FieldName string `json:"fieldName"`
	Value     string `json:"value"`
}

type CurrentStudentQueryDTO struct {
	PageRequestModel PageRequestModel      `json:"pageRequestModel"`
	QueryModel       CurrentStudentFilters `json:"queryModel"`
	SortModel        SortModel             `json:"sortModel"`
}

type CurrentStudentFilters struct {
	StudentID         string `json:"studentId"`
	SalespersonID     *int64 `json:"salespersonId"`
	SearchKey         string `json:"searchKey"`
	WechatNumber      string `json:"wechatNumber"`
	SchoolSearchKey   string `json:"schoolSearchKey"`
	AddressSearchKey  string `json:"addressSearchKey"`
	InterestSearchKey string `json:"interestSearchKey"`
}

type CurrentStudent struct {
	ID                 int64      `json:"id"`
	StuName            string     `json:"stuName"`
	Mobile             string     `json:"mobile"`
	StudentStatus      int        `json:"studentStatus"`
	SalePerson         *int64     `json:"salePerson,omitempty"`
	SalePersonName     string     `json:"salePersonName,omitempty"`
	ChannelID          *int64     `json:"channelId,omitempty"`
	ChannelName        string     `json:"channelName,omitempty"`
	CreateTime         time.Time  `json:"createTime"`
	FirstReadTime      *time.Time `json:"firstReadTime,omitempty"`
	FollowUpTime       *time.Time `json:"followUpTime,omitempty"`
	BirthDay           *time.Time `json:"birthDay,omitempty"`
	Grade              string     `json:"grade,omitempty"`
	WeChatNumber       string     `json:"weChatNumber,omitempty"`
	StudySchool        string     `json:"studySchool,omitempty"`
	Interest           string     `json:"interest,omitempty"`
	Address            string     `json:"address,omitempty"`
	CreateID           *int64     `json:"createId,omitempty"`
	CreateName         string     `json:"createName,omitempty"`
	StudentManagerID   *int64     `json:"studentManagerId,omitempty"`
	StudentManagerName string     `json:"studentManagerName,omitempty"`
	AdvisorID          *int64     `json:"advisorId,omitempty"`
	AdvisorName        string     `json:"advisorName,omitempty"`
	FollowUpStatus     *int       `json:"followUpStatus,omitempty"`
}

type EnrolledStudentQueryDTO struct {
	PageRequestModel PageRequestModel      `json:"pageRequestModel"`
	QueryModel       EnrolledStudentFilter `json:"queryModel"`
}

type EnrolledStudentFilter struct {
	StudentID         string   `json:"studentId"`
	StuName           string   `json:"stuName"`
	Mobile            string   `json:"mobile"`
	Sexes             []int    `json:"sexes"`
	StudentStatuses   []int    `json:"studentStatuses"`
	Grades            []string `json:"grades"`
	ChannelIDs        []int64  `json:"channelIds"`
	WechatNumber      string   `json:"wechatNumber"`
	StudySchool       string   `json:"studySchool"`
	SchoolSearchKey   string   `json:"schoolSearchKey"`
	Address           string   `json:"address"`
	AddressSearchKey  string   `json:"addressSearchKey"`
	Interest          string   `json:"interest"`
	InterestSearchKey string   `json:"interestSearchKey"`
	CreateID          *int64   `json:"createId"`
	SalespersonID     *int64   `json:"salespersonId"`
}

type EnrolledStudent struct {
	ID                     int64            `json:"id"`
	StuName                string           `json:"stuName"`
	AvatarURL              string           `json:"avatarUrl,omitempty"`
	StuSex                 *int             `json:"stuSex,omitempty"`
	Mobile                 string           `json:"mobile"`
	PhoneRelationship      *int             `json:"phoneRelationship,omitempty"`
	IsCollect              bool             `json:"isCollect"`
	IsBindChild            bool             `json:"isBindChild"`
	StudentStatus          int              `json:"studentStatus"`
	CreateTime             *time.Time       `json:"createTime,omitempty"`
	ChannelID              *int64           `json:"channelId,omitempty"`
	ChannelName            string           `json:"channelName,omitempty"`
	AdvisorID              *int64           `json:"advisorId,omitempty"`
	AdvisorName            string           `json:"advisorName,omitempty"`
	StudentManagerID       *int64           `json:"studentManagerId,omitempty"`
	StudentManagerName     string           `json:"studentManagerName,omitempty"`
	FollowUpTime           *time.Time       `json:"followUpTime,omitempty"`
	BirthDay               *time.Time       `json:"birthDay,omitempty"`
	WeChatNumber           string           `json:"weChatNumber,omitempty"`
	SecondPhoneNumber      string           `json:"secondPhoneNumber,omitempty"`
	StudySchool            string           `json:"studySchool,omitempty"`
	Grade                  string           `json:"grade,omitempty"`
	Interest               string           `json:"interest,omitempty"`
	Address                string           `json:"address,omitempty"`
	RecommendStudentID     *int64           `json:"recommendStudentId,omitempty"`
	RecommendStudentName   string           `json:"recommendStudentName,omitempty"`
	Remark                 string           `json:"remark,omitempty"`
	SalesAssignedTime      *time.Time       `json:"salesAssignedTime,omitempty"`
	SalePerson             *int64           `json:"salePerson,omitempty"`
	SalePersonName         string           `json:"salePersonName,omitempty"`
	CustomInfo             []map[string]any `json:"customInfo,omitempty"`
	IsCrossSchoolStudent   bool             `json:"isCrossSchoolStudent"`
	CreateID               *int64           `json:"createId,omitempty"`
	CreateName             string           `json:"createName,omitempty"`
	FollowUpStatus         *int             `json:"followUpStatus,omitempty"`
	CollectorStaffID       *int64           `json:"collectorStaffId,omitempty"`
	CollectorStaffName     string           `json:"collectorStaffName,omitempty"`
	ForegroundStaffID      *int64           `json:"foregroundStaffId,omitempty"`
	ForegroundStaffName    string           `json:"foregroundStaffName,omitempty"`
	PhoneSellStaffID       *int64           `json:"phoneSellStaffId,omitempty"`
	PhoneSellStaffName     string           `json:"phoneSellStaffName,omitempty"`
	ViceSellStaffStaffID   *int64           `json:"viceSellStaffStaffId,omitempty"`
	ViceSellStaffStaffName string           `json:"viceSellStaffStaffName,omitempty"`
	FirstEnrolledTime      *time.Time       `json:"firstEnrolledTime,omitempty"`
}

type StudentStatusUpdateDTO struct {
	ID             int64 `json:"id"`
	FollowUpStatus *int  `json:"followUpStatus"`
	IntentLevel    *int  `json:"intentLevel"`
}

type BatchCommonDTO struct {
	SalespersonID    *int64  `json:"salespersonId"`
	StudentIDs       []int64 `json:"studentIds"`
	UserIDs          []int64 `json:"userIds"`
	DeptIDs          []int64 `json:"deptIds"`
	RoleIDs          []int64 `json:"roleIds"`
	CourseIDs        []int64 `json:"courseIds"`
	DelFlag          *bool   `json:"delFlag"`
	SaleStatus       *bool   `json:"saleStatus"`
	IsShowMicoSchool *bool   `json:"isShowMicoSchool"`
	IsWork           *bool   `json:"isWork"`
}

type StudentSaveDTO struct {
	StudentID          *int64       `json:"studentId"`
	OperatorID         *int64       `json:"-"`
	UUID               string       `json:"uuid"`
	Version            *int64       `json:"version"`
	StuName            string       `json:"stuName"`
	Mobile             string       `json:"mobile"`
	Avatar             string       `json:"avatar"`
	Sex                *int         `json:"sex"`
	Birthday           *time.Time   `json:"birthday"`
	Grade              string       `json:"grade"`
	StudySchool        string       `json:"studySchool"`
	Interest           string       `json:"interest"`
	PhoneRelationship  *int         `json:"phoneRelationship"`
	Address            string       `json:"address"`
	ChannelID          *int64       `json:"channelId"`
	WeChatNumber       string       `json:"weChatNumber"`
	RecommendStudentID *int64       `json:"recommendStudentId"`
	SalespersonID      *int64       `json:"salespersonId"`
	CollectorStaffID   *int64       `json:"collectorStaffId"`
	PhoneSellStaffID   *int64       `json:"phoneSellStaffId"`
	ForegroundStaffID  *int64       `json:"foregroundStaffId"`
	ViceSellStaffID    *int64       `json:"viceSellStaffId"`
	StudentManagerID   *int64       `json:"studentManagerId"`
	AdvisorID          *int64       `json:"advisorId"`
	CustomInfo         []CustomInfo `json:"customInfo,omitempty"`
	Remark             string       `json:"remark"`
}

type StudentDuplicateCheckRequest struct {
	ID      *int64 `json:"id"`
	StuName string `json:"stuName"`
	Mobile  string `json:"mobile"`
}

type IntentStudentRepeatVO struct {
	AddStudentRepeatRuleEnum int `json:"addStudentRepeatRuleEnum"`
}

type RecommenderQueryDTO struct {
	PageRequestModel PageRequestModel   `json:"pageRequestModel"`
	QueryModel       RecommenderFilters `json:"queryModel"`
	SortModel        SortModel          `json:"sortModel"`
}

type RecommenderFilters struct {
	StudentID     *int64 `json:"studentId"`
	SearchKey     string `json:"searchKey"`
	StudentStatus *int   `json:"studentStatus"`
}

type RecommenderQueryVO struct {
	ID            int64  `json:"id"`
	StuName       string `json:"stuName"`
	AvatarURL     string `json:"avatarUrl,omitempty"`
	Mobile        string `json:"mobile"`
	StudentStatus int    `json:"studentStatus"`
}

type BirthdayStudentQueryDTO struct {
	PageRequestModel PageRequestModel `json:"pageRequestModel"`
	QueryModel       BirthdayFilters  `json:"queryModel"`
	SortModel        SortModel        `json:"sortModel"`
}

type BirthdayFilters struct {
	StudentManagerID *int64 `json:"studentManagerId"`
	AdvisorID        *int64 `json:"advisorId"`
	BirthMonth       *int   `json:"birthMonth"`
	AgeMin           *int   `json:"ageMin"`
	AgeMax           *int   `json:"ageMax"`
}

type BirthdayStudentQueryVO struct {
	ID                 int64      `json:"id"`
	StuName            string     `json:"stuName"`
	AvatarURL          string     `json:"avatarUrl,omitempty"`
	StuSex             *int       `json:"stuSex,omitempty"`
	Mobile             string     `json:"mobile"`
	PhoneRelationship  *int       `json:"phoneRelationship,omitempty"`
	StudentStatus      int        `json:"studentStatus"`
	BirthDay           *time.Time `json:"birthDay,omitempty"`
	StudentManagerID   *int64     `json:"studentManagerId,omitempty"`
	StudentManagerName string     `json:"studentManagerName,omitempty"`
	AdvisorID          *int64     `json:"advisorId,omitempty"`
	AdvisorName        string     `json:"advisorName,omitempty"`
}

type StudentChangeRecord struct {
	ID            int64     `json:"id"`
	StuID         int64     `json:"stuId"`
	ChangeContent string    `json:"changeContent"`
	ChangeID      *int64    `json:"changeId,omitempty"`
	ChangeName    string    `json:"changeName,omitempty"`
	CreateTime    time.Time `json:"createTime"`
	Remark        string    `json:"remark,omitempty"`
}

type StudentFieldKey struct {
	ID          int64  `json:"id"`
	UUID        string `json:"uuid,omitempty"`
	Version     int64  `json:"version,omitempty"`
	InstID      int64  `json:"instId,omitempty"`
	FieldKey    string `json:"fieldKey"`
	FieldType   int    `json:"fieldType"`
	Required    bool   `json:"required"`
	Searched    bool   `json:"searched"`
	OptionsJSON string `json:"optionsJson,omitempty"`
	IsDefault   bool   `json:"isDefault"`
	IsDisplay   bool   `json:"isDisplay"`
	CanDelete   bool   `json:"canDelete"`
	CanEdit     bool   `json:"canEdit"`
	Sort        int    `json:"sort"`
	Remark      string `json:"remark,omitempty"`
}
