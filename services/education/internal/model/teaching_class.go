package model

import "time"

const (
	TeachingClassTypeNormal   = 1
	TeachingClassTypeOneToOne = 2
)

const (
	TeachingClassStatusActive = 1
	TeachingClassStatusClosed = 2
)

const (
	TeachingClassStudentStatusStudying = 1
	TeachingClassStudentStatusStopped  = 2
	TeachingClassStudentStatusClosed   = 3
)

type OneToOneListQueryDTO struct {
	PageRequestModel PageRequestModel       `json:"pageRequestModel"`
	QueryModel       OneToOneListQueryModel `json:"queryModel"`
}

type OneToOneListQueryModel struct {
	SearchKey          string   `json:"searchKey"`
	StudentID          string   `json:"studentId"`
	LessonIDs          []string `json:"lessonIds"`
	ClassTeacherID     string   `json:"classTeacherId"`
	DefaultTeacherID   string   `json:"defaultTeacherId"`
	HasClassTeacher    *bool    `json:"hasClassTeacher"`
	IsScheduled        *bool    `json:"isScheduled"`
	Status             []int    `json:"status"`
	ClassStudentStatus []int    `json:"classStudentStatus"`
	StartDate          string   `json:"startDate"`
	EndDate            string   `json:"endDate"`
}

type OneToOneLessonTimeVO struct {
	ID        string     `json:"id,omitempty"`
	StartTime *time.Time `json:"startTime,omitempty"`
	EndTime   *time.Time `json:"endTime,omitempty"`
}

type OneToOneLessonDayInfoVO struct {
	LessonDayCount         int `json:"lessonDayCount"`
	CompleteLessonDayCount int `json:"completeLessonDayCount"`
}

type OneToOneTeacherVO struct {
	TeacherID string `json:"teacherId"`
	Name      string `json:"name"`
	Status    int    `json:"status"`
	ClassID   string `json:"classId"`
	IsDefault bool   `json:"isDefault,omitempty"`
}

type OneToOneTuitionAccountVO struct {
	ID                 string     `json:"id"`
	TotalTuition       float64    `json:"totalTuition"`
	RemainTuition      float64    `json:"remainTuition"`
	TotalQuantity      float64    `json:"totalQuantity"`
	TotalFreeQuantity  float64    `json:"totalFreeQuantity"`
	RemainQuantity     float64    `json:"remainQuantity"`
	RemainFreeQuantity float64    `json:"remainFreeQuantity"`
	LessonChargingMode int        `json:"lessonChargingMode"`
	ProductName        string     `json:"productName"`
	Status             int        `json:"status"`
	EnableExpireTime   bool       `json:"enableExpireTime"`
	LastSuspendedTime  time.Time  `json:"lastSuspendedTime"`
	ExpireTime         time.Time  `json:"expireTime"`
	StudentID          string     `json:"studentId"`
	LessonID           string     `json:"lessonId"`
	LessonType         int        `json:"lessonType"`
	ChangeStatusTime   time.Time  `json:"changeStatusTime"`
	SuspendedTime      *time.Time `json:"suspendedTime"`
	ClassEndingTime    *time.Time `json:"classEndingTime"`
	AssignedClass      bool       `json:"assignedClass"`
}

type OneToOnePropertyVO struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type OneToOneItemVO struct {
	ID                         string                   `json:"id"`
	Name                       string                   `json:"name"`
	StudentName                string                   `json:"studentName"`
	StudentID                  string                   `json:"studentId"`
	Sex                        int                      `json:"sex"`
	Avatar                     string                   `json:"avatar"`
	Phone                      string                   `json:"phone"`
	SchoolID                   string                   `json:"schoolId"`
	One2OneLessonTimes         []OneToOneLessonTimeVO   `json:"one2OneLessonTimes"`
	IsScheduled                bool                     `json:"isScheduled"`
	Status                     int                      `json:"status"`
	ClassStudentStatus         int                      `json:"classStudentStatus"`
	One2OneLessonDayInfo       OneToOneLessonDayInfoVO  `json:"one2OneLessonDayInfo"`
	CreatedTime                time.Time                `json:"createdTime"`
	ClassRoomID                string                   `json:"classRoomId"`
	ClassRoomName              string                   `json:"classRoomName"`
	ClassroomEnabled           *bool                    `json:"classroomEnabled"`
	ClassTime                  float64                  `json:"classTime"`
	StudentClassTime           float64                  `json:"studentClassTime"`
	TeacherClassTime           float64                  `json:"teacherClassTime"`
	LessonID                   string                   `json:"lessonId"`
	LessonName                 string                   `json:"lessonName"`
	OrderCourseDetailID        string                   `json:"orderCourseDetailId,omitempty"`
	TuitionAccountCount        int                      `json:"tuitionAccountCount"`
	TuitionAccountID           string                   `json:"tuitionAccountId"`
	DefaultTeacherID           string                   `json:"defaultTeacherId"`
	DefaultTeacherName         string                   `json:"defaultTeacherName"`
	DefaultClassTimeRecordMode int                      `json:"defaultClassTimeRecordMode"`
	IsGradeUpgrade             bool                     `json:"isGradeUpgrade"`
	LastFinishedLessonDay      time.Time                `json:"lastFinishedLessonDay"`
	TeacherList                []OneToOneTeacherVO      `json:"teacherList"`
	TuitionAccount             OneToOneTuitionAccountVO `json:"tuitionAccount"`
	ClassProperties            []OneToOnePropertyVO     `json:"classProperties"`
	ClassTeacherID             string                   `json:"classTeacherId,omitempty"`
	ClassTeacherName           string                   `json:"classTeacherName,omitempty"`
	Remark                     string                   `json:"remark,omitempty"`
}

type OneToOneListResultVO struct {
	Total        int              `json:"total"`
	StudentCount int              `json:"studentCount"`
	List         []OneToOneItemVO `json:"list"`
}

type OneToOneDetailVO struct {
	ID                         string                   `json:"id"`
	StudentID                  string                   `json:"studentId"`
	SchoolID                   string                   `json:"schoolId"`
	Name                       string                   `json:"name"`
	StudentName                string                   `json:"studentName"`
	StudentAvatar              string                   `json:"studentAvatar"`
	StudentGender              int                      `json:"studentGender"`
	LessonID                   string                   `json:"lessonId"`
	LessonName                 string                   `json:"lessonName"`
	LessonPrice                float64                  `json:"lessonPrice"`
	ClassroomID                string                   `json:"classroomId"`
	ClassroomName              *string                  `json:"classroomName"`
	OrderCourseDetailID        string                   `json:"orderCourseDetailId,omitempty"`
	TuitionAccountCount        int                      `json:"tuitionAccountCount"`
	TuitionAccountID           string                   `json:"tuitionAccountId"`
	ClassTime                  float64                  `json:"classTime"`
	IsScheduled                bool                     `json:"isScheduled"`
	ClassroomEnabled           *bool                    `json:"classroomEnabled"`
	Status                     int                      `json:"status"`
	ClassStudentStatus         int                      `json:"classStudentStatus"`
	CreatedTime                time.Time                `json:"createdTime"`
	DefaultStudentClassTime    float64                  `json:"defaultStudentClassTime"`
	DefaultTeacherClassTime    float64                  `json:"defaultTeacherClassTime"`
	DefaultClassTimeRecordMode int                      `json:"defaultClassTimeRecordMode"`
	DefaultTeacherID           string                   `json:"defaultTeacherId"`
	DefaultTeacherName         string                   `json:"defaultTeacherName"`
	ClassTeacherName           string                   `json:"classTeacherName,omitempty"`
	IsGradeUpgrade             bool                     `json:"isGradeUpgrade"`
	Remark                     string                   `json:"remark"`
	TeacherList                []OneToOneTeacherVO      `json:"teacherList"`
	TuitionAccount             OneToOneTuitionAccountVO `json:"tuitionAccount"`
	CreatedStaffID             string                   `json:"createdStaffId"`
	CreatedStaffName           string                   `json:"createdStaffName"`
	ClassProperties            []OneToOnePropertyVO     `json:"classProperties"`
	DefaultTeacherStatus       int                      `json:"defaultTeacherStatus"`
}

type OneToOneBatchAssignTeacherDTO struct {
	IDs             []string `json:"ids"`
	ClassTeacherID  string   `json:"classTeacherId"`  // 兼容旧版单选
	ClassTeacherIDs []string `json:"classTeacherIds"` // 多选班主任
}

type OneToOneBatchClassTimeDTO struct {
	IDs                 []string `json:"ids"`
	ClassTime           float64  `json:"classTime"`
	StudentClassTime    float64  `json:"studentClassTime"`
	TeacherClassTime    float64  `json:"teacherClassTime"`
	ClassTimeRecordMode int      `json:"classTimeRecordMode"` // 1 按固定课时 2 按上课时长
}

type OneToOneCheckNameDTO struct {
	Name      string `json:"name"`
	ExceptID  string `json:"exceptId"`
	IsOne2One bool   `json:"isOne2One"`
}

// OneToOneExistDTO 对标 SchoolPal ExistOne2One：检测学员是否已有该课程的「开班中」1 对 1
type OneToOneExistDTO struct {
	StudentID string `json:"studentId"`
	LessonID  string `json:"lessonId"`
}

// OneToOneLessonsByStudentQueryDTO 对标 QueryOne2OneLessonByStudentId：学员已有学费账户、且课程为 1v1 授课的可选课程
type OneToOneLessonsByStudentQueryDTO struct {
	StudentID            string `json:"studentId"`
	TuitionAccountStatus []int  `json:"tuitionAccountStatus"`
}

type OneToOneLessonOptionVO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// AlreadyEnrolled 已报名：学费账户已分班，或已存在开班中的 1 对 1 班级（下拉右侧展示「已报名」）
	AlreadyEnrolled bool `json:"alreadyEnrolled"`
}

type OneToOneLessonsByStudentResult struct {
	List []OneToOneLessonOptionVO `json:"list"`
}

// OneToOneCloseDTO 仅结班（
type OneToOneCloseDTO struct {
	ID string `json:"id"`
}

type OneToOneUpdateDTO struct {
	ID                         string               `json:"id"`
	StudentID                  string               `json:"studentId"`
	LessonID                   string               `json:"lessonId"`
	ClassroomID                string               `json:"classroomId"`
	Name                       string               `json:"name"`
	TeacherID                  []string             `json:"teacherId"`
	DefaultTeacherID           string               `json:"defaultTeacherId"`
	DefaultStudentClassTime    float64              `json:"defaultStudentClassTime"`
	DefaultTeacherClassTime    float64              `json:"defaultTeacherClassTime"`
	DefaultClassTimeRecordMode int                  `json:"defaultClassTimeRecordMode"`
	Remark                     string               `json:"remark"`
	ClassProperties            []OneToOnePropertyVO `json:"classProperties"`
	// AllowDuplicateName 为 true 时跳过「1对1名称唯一」校验（前端已二次确认同名）
	AllowDuplicateName bool `json:"allowDuplicateName,omitempty"`
}

// OneToOneCreateDTO 手动创建 1 对 1。TuitionAccountID 为学费账户数字 id，
// 或 agg:{扣费课程id}:{授课方式}:{计费模式}（与下拉汇总行一致，服务端仅绑定该聚合桶下的账户）。
type OneToOneCreateDTO struct {
	StudentID                  string               `json:"studentId"`
	LessonID                   string               `json:"lessonId"`
	ClassroomID                string               `json:"classroomId"`
	TuitionAccountID           string               `json:"tuitionAccountId"`
	Name                       string               `json:"name"`
	TeacherID                  []string             `json:"teacherId"`
	DefaultTeacherID           string               `json:"defaultTeacherId"`
	DefaultStudentClassTime    float64              `json:"defaultStudentClassTime"`
	DefaultTeacherClassTime    float64              `json:"defaultTeacherClassTime"`
	DefaultClassTimeRecordMode int                  `json:"defaultClassTimeRecordMode"`
	Remark                     string               `json:"remark"`
	ClassProperties            []OneToOnePropertyVO `json:"classProperties"`
	AllowDuplicateName         bool                 `json:"allowDuplicateName,omitempty"`
}

// OneToOneCreateResult 创建成功返回班级 ID
type OneToOneCreateResult struct {
	ID string `json:"id"`
}

type OneToOneSwitchDefaultTuitionAccountDTO struct {
	ID               string `json:"id"`
	TuitionAccountID string `json:"tuitionAccountId"`
}

type StudentLessonTuitionAccountsQueryDTO struct {
	StudentID           string `json:"studentId"`
	LessonID            string `json:"lessonId"`
	TeachingClassID     string `json:"teachingClassId,omitempty"`
	OrderCourseDetailID string `json:"orderCourseDetailId,omitempty"`
}

// StudentOneToOneDeductionAccountsQueryDTO 创建 1 对 1 时选扣费账户：按学员查全部在读的报读账户（班级授课或 1v1，不限上课课程）
type StudentOneToOneDeductionAccountsQueryDTO struct {
	StudentID string `json:"studentId"`
}

// StudentLessonTuitionAccountItem 单条学费账户（含竞品常用字段名 quantity/tuition 表示剩余）
type StudentLessonTuitionAccountItem struct {
	ID                     string     `json:"id"`
	StudentID              string     `json:"studentId"`
	LessonID               string     `json:"lessonId"`
	LessonName             string     `json:"lessonName,omitempty"`
	ProductName            string     `json:"productName"`
	LessonChargingMode     int        `json:"lessonChargingMode"`
	TotalQuantity          float64    `json:"totalQuantity"`
	TotalFreeQuantity      float64    `json:"totalFreeQuantity"`
	TotalTuition           float64    `json:"totalTuition"`
	FreeQuantity           float64    `json:"freeQuantity"`
	Quantity               float64    `json:"quantity"`
	Tuition                float64    `json:"tuition"`
	Suspended              bool       `json:"suspended"`
	SuspendedTime          *time.Time `json:"suspendedTime,omitempty"`
	StartTime              time.Time  `json:"startTime"`
	EnableExpireTime       bool       `json:"enableExpireTime"`
	ExpireTime             time.Time  `json:"expireTime"`
	AssignedClass          bool       `json:"assignedClass"`
	LatestStartTime        time.Time  `json:"latestStartTime"`
	LessonType             int        `json:"lessonType"`
	IsTuitionAccountActive bool       `json:"isTuitionAccountActive"`
	Status                 int        `json:"status"`
	// IsAggregate 为 true 时 id 形如 agg:{courseId}:{teachMethod}:{lessonModel}，
	// 表示该课程某个计费桶下多笔在读账户的汇总；创建 1 对 1 时会为该桶下每笔账户各写一条班员绑定。
	IsAggregate bool `json:"isAggregate,omitempty"`
}

type StudentLessonTuitionAccountsResult struct {
	List []StudentLessonTuitionAccountItem `json:"list"`
}

type CloseTuitionAccountOrderDTO struct {
	TuitionAccountID string  `json:"tuitionAccountId"`
	Quantity         float64 `json:"quantity"`
	FreeQuantity     float64 `json:"freeQuantity"`
	Tuition          float64 `json:"tuition"`
	Remark           string  `json:"remark"`
}

// CloseTuitionAccountOrderResult 返回生成的流水/订单 id，供前端展示
type CloseTuitionAccountOrderResult struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// --- 班级授课（集体班）对标 CheckClassName / Create / QueryClassList / QueryClassStatisticsInfo ---

type GroupClassCheckNameDTO struct {
	Name      string `json:"name"`
	IsOne2One bool   `json:"isOne2One"`
	ExceptID  string `json:"exceptId"`
}

type GroupClassCreateDTO struct {
	Name                       string   `json:"name"`
	LessonID                   string   `json:"lessonId"`
	ClassroomID                string   `json:"classroomId"`
	MaxCount                   int      `json:"maxCount"`
	TeacherIDs                 []string `json:"teacherIds"`
	DefaultTeacherID           string   `json:"defaultTeacherId"`
	DefaultStudentClassTime    float64  `json:"defaultStudentClassTime"`
	DefaultTeacherClassTime    float64  `json:"defaultTeacherClassTime"`
	DefaultClassTimeRecordMode int      `json:"defaultClassTimeRecordMode"`
	IsCopyStudent              bool     `json:"isCopyStudent"`
	CopiedStudents             []any    `json:"copiedStudents"`
	IsCopyTimetable            bool     `json:"isCopyTimetable"`
	ClassProperties            []any    `json:"classProperties"`
	Remark                     string   `json:"remark"`
}

// GroupClassUpdateDTO 对标 Class/Update，内嵌字段与创建一致，另含 id、copyFromClassId
type GroupClassUpdateDTO struct {
	GroupClassCreateDTO
	ID              string `json:"id"`
	CopyFromClassID string `json:"copyFromClassId"`
}

type GroupClassCreateResult struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type GroupClassPageRequestModel struct {
	NeedTotal bool `json:"needTotal"`
	PageSize  int  `json:"pageSize"`
	PageIndex int  `json:"pageIndex"`
	SkipCount int  `json:"skipCount"`
}

type GroupClassListQueryModel struct {
	ClassIDs         []string `json:"classIds"`
	Statues          []int    `json:"statues"`
	LessonIDs        []string `json:"lessonIds"`
	ClassName        string   `json:"className"`
	TeacherID        string   `json:"teacherId"`
	DefaultTeacherID string   `json:"defaultTeacherId"`
	ClassRoomName    string   `json:"classRoomName"`
	IsMultiProduct   *bool    `json:"isMultiProduct"`
	IsScheduled      *bool    `json:"isScheduled"`
	CreatedStaffIDs  []string `json:"createdStaffIds"`
	CreatedStartTime string   `json:"createdStartTime"`
	CreatedEndTime   string   `json:"createdEndTime"`
	ClosedStartDate  string   `json:"closedStartDate"`
	ClosedEndDate    string   `json:"closedEndDate"`
	ClassProperties  []any    `json:"classProperties"`
}

type GroupClassListBody struct {
	QueryModel       GroupClassListQueryModel   `json:"queryModel"`
	PageRequestModel GroupClassPageRequestModel `json:"pageRequestModel"`
}

type GroupClassListTeacherVO struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Mobile string `json:"mobile,omitempty"`
	Status int    `json:"status"`
	Avatar string `json:"avatar"`
}

type GroupClassLessonDayInfoVO struct {
	LessonDayCount         int `json:"lessonDayCount"`
	CompleteLessonDayCount int `json:"completeLessonDayCount"`
}

type GroupClassListItemVO struct {
	ID                         string                    `json:"id"`
	Name                       string                    `json:"name"`
	ClassTime                  float64                   `json:"classTime"`
	LessonID                   string                    `json:"lessonId"`
	LessonName                 string                    `json:"lessonName"`
	IsMultiProduct             bool                      `json:"isMultiProduct"`
	StudentCount               int                       `json:"studentCount"`
	LockStudentCount           int                       `json:"lockStudentCount"`
	MaxCount                   int                       `json:"maxCount"`
	Teachers                   []GroupClassListTeacherVO `json:"teachers"`
	DefaultTeacherID           string                    `json:"defaultTeacherId"`
	DefaultTeacherName         string                    `json:"defaultTeacherName"`
	ClassRoomName              string                    `json:"classRoomName"`
	ClassLessonTimes           []any                     `json:"classLessonTimes"`
	IsScheduled                bool                      `json:"isScheduled"`
	ClassLessonDayInfos        GroupClassLessonDayInfoVO `json:"classLessonDayInfos"`
	Status                     int                       `json:"status"`
	ClosedTime                 time.Time                 `json:"closedTime"`
	CreatedTime                time.Time                 `json:"createdTime"`
	CreatedStaffName           string                    `json:"createdStaffName"`
	Remark                     string                    `json:"remark"`
	ClassProperties            []any                     `json:"classProperties"`
	DefaultStudentClassTime    float64                   `json:"defaultStudentClassTime"`
	DefaultTeacherClassTime    float64                   `json:"defaultTeacherClassTime"`
	DefaultClassTimeRecordMode int                       `json:"defaultClassTimeRecordMode"`
}

type GroupClassListPageResult struct {
	List  []GroupClassListItemVO `json:"list"`
	Total int                    `json:"total"`
}

// GroupClassDetailVO 对标 ToB/PC/Class/Get，供编辑弹窗拉取完整班级信息
type GroupClassDetailVO struct {
	ID                         string                    `json:"id"`
	Name                       string                    `json:"name"`
	Status                     int                       `json:"status"`
	LessonID                   string                    `json:"lessonId"`
	LessonName                 string                    `json:"lessonName"`
	StudentCount               int                       `json:"studentCount"`
	LockStudentCount           int                       `json:"lockStudentCount"`
	MaxCount                   int                       `json:"maxCount"`
	ClassroomID                string                    `json:"classroomId"`
	ClassroomName              string                    `json:"classroomName"`
	ClassroomEnabled           bool                      `json:"classroomEnabled"`
	ClassroomAddressCharge     int                       `json:"classroomAddressCharge"`
	Teachers                   []GroupClassListTeacherVO `json:"teachers"`
	TeacherCount               int                       `json:"teacherCount"`
	ClassTime                  float64                   `json:"classTime"`
	DefaultStudentClassTime    float64                   `json:"defaultStudentClassTime"`
	DefaultTeacherClassTime    float64                   `json:"defaultTeacherClassTime"`
	DefaultClassTimeRecordMode int                       `json:"defaultClassTimeRecordMode"`
	DefaultTeacherID           string                    `json:"defaultTeacherId"`
	DefaultTeacherStatus       int                       `json:"defaultTeacherStatus"`
	DefaultTeacherName         string                    `json:"defaultTeacherName"`
	LessonType                 int                       `json:"lessonType"`
	LessonScope                int                       `json:"lessonScope"`
	CreatedTime                time.Time                 `json:"createdTime"`
	ClosedTime                 time.Time                 `json:"closedTime"`
	LessonPrice                float64                   `json:"lessonPrice"`
	IsMultiProduct             bool                      `json:"isMultiProduct"`
	Remark                     string                    `json:"remark"`
	ClassProperties            []any                     `json:"classProperties"`
}

type GroupClassStatisticsVO struct {
	ClassCount        int `json:"classCount"`
	OpenClassCount    int `json:"openClassCount"`
	StudentCount      int `json:"studentCount"`
	StudentPersonTime int `json:"studentPersonTime"`
}

// --- 集体班添加学员：对标 Class/GetStudentListByClassIds + TuitionAccount/GetTuitionAccountListByLessonId ---

type GroupClassStudentListByClassIDsRequest struct {
	ClassIDs []string `json:"classIds"`
}

// BatchAssignGroupClassStudentsRequest 对标 ToB/PC/Class/BatchAssignStudents
type BatchAssignGroupClassStudentsRequest struct {
	ClassIDs           []string                           `json:"classIds"`
	Students           []BatchAssignGroupClassStudentItem `json:"students"`
	EnforceClassAssign bool                               `json:"enforceClassAssign"`
}

type BatchAssignGroupClassStudentItem struct {
	StudentID        string `json:"studentId"`
	TuitionAccountID string `json:"tuitionAccountId"`
}

type GroupClassStudentTuitionSnapVO struct {
	TuitionAccountID       string    `json:"tuitionAccountId"`
	ProductName            string    `json:"productName"`
	ProductID              string    `json:"productId"`
	RemainQuantity         float64   `json:"remainQuantity"`
	RemainFreeQuantity     float64   `json:"remainFreeQuantity"`
	RemainTuition          float64   `json:"remainTuition"`
	LessonChargingMode     int       `json:"lessonChargingMode"`
	EnableExpireTime       bool      `json:"enableExpireTime"`
	StartTime              time.Time `json:"startTime"`
	ExpireTime             time.Time `json:"expireTime"`
	IsTuitionAccountActive bool      `json:"isTuitionAccountActive"`
	TotalQuantity          float64   `json:"totalQuantity"`
	TotalFreeQuantity      float64   `json:"totalFreeQuantity"`
	TotalTuition           float64   `json:"totalTuition"`
}

type GroupClassStudentInClassItemVO struct {
	ID                             string                          `json:"id"`
	Name                           string                          `json:"name"`
	Avatar                         string                          `json:"avatar"`
	IsBind                         bool                            `json:"isBind"`
	ClassID                        string                          `json:"classId"`
	Phone                          string                          `json:"phone"`
	Sex                            int                             `json:"sex"`
	TuitionAccountID               string                          `json:"tuitionAccountId"`
	Birthday                       time.Time                       `json:"birthday"`
	JoinTime                       time.Time                       `json:"joinTime"`
	ClassStudentTuitionAccountInfo *GroupClassStudentTuitionSnapVO `json:"classStudentTuitionAccountInfo"`
}

type GroupClassStudentListBucketVO struct {
	ClassID  string                           `json:"classId"`
	Students []GroupClassStudentInClassItemVO `json:"students"`
}

type TuitionAccountListByLessonIDBody struct {
	PageRequestModel GroupClassPageRequestModel      `json:"pageRequestModel"`
	QueryModel       TuitionAccountListByLessonQuery `json:"queryModel"`
}

// TuitionAccountLessonPageFilters 对标竞品 GetTuitionAccountListByLessonId 的 queryModel 筛选（性别、年龄区间、姓名模糊）。
type TuitionAccountLessonPageFilters struct {
	Sex         []int  `json:"sex,omitempty"`
	AgeMin      *int   `json:"ageMin,omitempty"`
	AgeMax      *int   `json:"ageMax,omitempty"`
	StudentName string `json:"studentName,omitempty"`
}

type TuitionAccountListByLessonQuery struct {
	LessonID   string   `json:"lessonId"`
	StudentIDs []string `json:"studentIds"`
	/** 当前集体班 id：用于标记本班已有学员（assignedClass），与前端勾选禁用一致 */
	ClassID string `json:"classId,omitempty"`
	TuitionAccountLessonPageFilters
}

type TuitionAccountByLessonRowVO struct {
	StudentID              string    `json:"studentId"`
	TuitionAccountID       string    `json:"tuitionAccountId"`
	StudentName            string    `json:"studentName"`
	AssignedClass          bool      `json:"assignedClass"`
	Quantity               float64   `json:"quantity"`
	Avatar                 *string   `json:"avatar"`
	Phone                  string    `json:"phone"`
	LessonChargingMode     int       `json:"lessonChargingMode"`
	LessonScope            int       `json:"lessonScope"`
	IsTuitionAccountActive bool      `json:"isTuitionAccountActive"`
	StartTime              time.Time `json:"startTime"`
	IsCrossSchoolStudent   bool      `json:"isCrossSchoolStudent"`
	Sex                    int       `json:"sex"`
	Birthday               time.Time `json:"birthday"`
	ProductID              string    `json:"productId"`
	ProductName            string    `json:"productName"`
}

type TuitionAccountListByLessonIDResult struct {
	List  []TuitionAccountByLessonRowVO `json:"list"`
	Total int                           `json:"total"`
}
