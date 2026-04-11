package model

import "time"

const (
	TeachingScheduleStatusActive   = 1
	TeachingScheduleStatusCanceled = 2
)

const (
	TeachingScheduleCancelScopeCurrent = "current"
	TeachingScheduleCancelScopeFuture  = "future"
)

const (
	TeachingScheduleStudentTypeClassMember = 1
	TeachingScheduleStudentTypeTemporary   = 2
	TeachingScheduleStudentTypeTrial       = 3
	TeachingScheduleStudentTypeMakeup      = 4
)

const (
	TeachingScheduleStudentRosterStatusActive  = 1
	TeachingScheduleStudentRosterStatusLeave   = 2
	TeachingScheduleStudentRosterStatusRemoved = 3
)

type TeachingScheduleCreateSlotDTO struct {
	LessonDate             string   `json:"lessonDate"`
	StartTime              string   `json:"startTime"`
	EndTime                string   `json:"endTime"`
	TeacherID              string   `json:"teacherId,omitempty"`
	AssistantIDs           []string `json:"assistantIds,omitempty"`
	ClassroomID            string   `json:"classroomId,omitempty"`
	AllowStudentConflict   bool     `json:"allowStudentConflict,omitempty"`
	AllowClassroomConflict bool     `json:"allowClassroomConflict,omitempty"`
}

type TeachingScheduleBatchMeta struct {
	SchedulingMode    string   `json:"schedulingMode,omitempty"`
	RepeatRule        string   `json:"repeatRule,omitempty"`
	HolidayPolicy     string   `json:"holidayPolicy,omitempty"`
	SelectedWeekdays  []string `json:"selectedWeekdays,omitempty"`
	ScheduleStartDate string   `json:"scheduleStartDate,omitempty"`
	FreeSelectedDates []string `json:"freeSelectedDates,omitempty"`
	PlannedClassCount int      `json:"plannedClassCount,omitempty"`
}

type CreateOneToOneSchedulesDTO struct {
	OneToOneID             string                          `json:"oneToOneId"`
	TeacherID              string                          `json:"teacherId"`
	AssistantIDs           []string                        `json:"assistantIds"`
	ClassroomID            string                          `json:"classroomId"`
	ExcludeIDs             []string                        `json:"excludeIds,omitempty"`
	AllowStudentConflict   bool                            `json:"allowStudentConflict,omitempty"`
	AllowClassroomConflict bool                            `json:"allowClassroomConflict,omitempty"`
	BatchMeta              *TeachingScheduleBatchMeta      `json:"batchMeta,omitempty"`
	Schedules              []TeachingScheduleCreateSlotDTO `json:"schedules"`
}

type CreateGroupClassSchedulesDTO struct {
	GroupClassID           string                          `json:"groupClassId"`
	TeacherID              string                          `json:"teacherId"`
	AssistantIDs           []string                        `json:"assistantIds"`
	ClassroomID            string                          `json:"classroomId"`
	ExcludeIDs             []string                        `json:"excludeIds,omitempty"`
	AllowStudentConflict   bool                            `json:"allowStudentConflict,omitempty"`
	AllowClassroomConflict bool                            `json:"allowClassroomConflict,omitempty"`
	BatchMeta              *TeachingScheduleBatchMeta      `json:"batchMeta,omitempty"`
	Schedules              []TeachingScheduleCreateSlotDTO `json:"schedules"`
}

type OneToOneScheduleAvailabilitySlotDTO struct {
	TeacherID  string `json:"teacherId"`
	LessonDate string `json:"lessonDate"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
}

type CheckOneToOneScheduleAvailabilityDTO struct {
	OneToOneID string                                `json:"oneToOneId"`
	ExcludeIDs []string                              `json:"excludeIds,omitempty"`
	Schedules  []OneToOneScheduleAvailabilitySlotDTO `json:"schedules"`
}

type AssistantScheduleAvailabilitySlotDTO struct {
	LessonDate string `json:"lessonDate"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
}

type CheckAssistantScheduleAvailabilityDTO struct {
	OneToOneID   string                                 `json:"oneToOneId"`
	AssistantIDs []string                               `json:"assistantIds"`
	ExcludeIDs   []string                               `json:"excludeIds,omitempty"`
	Schedules    []AssistantScheduleAvailabilitySlotDTO `json:"schedules"`
}

type CheckGroupClassAssistantScheduleAvailabilityDTO struct {
	GroupClassID string                                 `json:"groupClassId"`
	AssistantIDs []string                               `json:"assistantIds"`
	ExcludeIDs   []string                               `json:"excludeIds,omitempty"`
	Schedules    []AssistantScheduleAvailabilitySlotDTO `json:"schedules"`
}

type TeachingScheduleListQueryDTO struct {
	StartDate     string   `json:"startDate"`
	EndDate       string   `json:"endDate"`
	SortDirection string   `json:"sortDirection,omitempty"`
	ClassType     *int     `json:"classType,omitempty"`
	StudentID     string   `json:"studentId,omitempty"`
	ConflictTypes []string `json:"conflictTypes,omitempty"`
	BatchNo       string   `json:"batchNo,omitempty"`
	IDs           []string `json:"ids,omitempty"`
	// 课表业务筛选（HTTP 上为逗号分隔）
	ScheduleTeacherIDs  []int64  `json:"scheduleTeacherIds,omitempty"`
	ClassroomIDs        []int64  `json:"classroomIds,omitempty"`
	GroupClassIDs       []int64  `json:"groupClassIds,omitempty"`
	OneToOneClassIDs    []int64  `json:"oneToOneClassIds,omitempty"`
	LessonIDs           []int64  `json:"lessonIds,omitempty"`
	ScheduleTypeFilters []string `json:"scheduleTypeFilters,omitempty"`
	CallStatusFilters   []string `json:"callStatusFilters,omitempty"`
	// Matrix API（by-teacher-matrix）可选：日期展示维度，1=周一…7=周日，空=不限制
	MatrixWeekdays []int `json:"matrixWeekdays,omitempty"`
	// Matrix API 可选：教师列筛选，all | has_class | no_class，空等价于 all
	MatrixTeacherFilter string `json:"matrixTeacherFilter,omitempty"`
	// Matrix API 可选：时段组 UUID（与 unified 配置 groups[].id 一致），服务端优先从 inst_period_group_teacher 解析教师列
	PeriodGroupUUID string `json:"periodGroupUuid,omitempty"`
	// Matrix API 可选：教师用户 ID 列表（HTTP 上为逗号分隔），当库中无时段组或组下无关联老师时作为回退筛选
	MatrixTeacherIDs []int64 `json:"matrixTeacherIds,omitempty"`
}

type TeachingScheduleConflictDetailQueryDTO struct {
	ID string `json:"id"`
}

type TeachingScheduleDetailQueryDTO struct {
	ID string `json:"id"`
}

type TeachingScheduleStudentRemoveCurrentDTO struct {
	ScheduleID string `json:"scheduleId"`
	StudentID  string `json:"studentId"`
}

type TeachingScheduleStudentCandidateQueryDTO struct {
	PageRequestModel PageRequestModel                           `json:"pageRequestModel"`
	QueryModel       TeachingScheduleStudentCandidateQueryModel `json:"queryModel"`
}

type TeachingScheduleStudentCandidateQueryModel struct {
	ScheduleID  string `json:"scheduleId"`
	StudentType int    `json:"studentType"`
	Keyword     string `json:"keyword"`
}

type TeachingScheduleStudentCandidateVO struct {
	StudentID             string `json:"studentId"`
	StudentName           string `json:"studentName"`
	AvatarURL             string `json:"avatarUrl,omitempty"`
	Phone                 string `json:"phone,omitempty"`
	MaskedPhone           string `json:"maskedPhone,omitempty"`
	PhoneRelationship     int    `json:"phoneRelationship,omitempty"`
	PhoneRelationshipText string `json:"phoneRelationshipText,omitempty"`
	StudentStatus         int    `json:"studentStatus,omitempty"`
	StudentStatusText     string `json:"studentStatusText,omitempty"`
}

type TeachingScheduleStudentCandidatePagedResult struct {
	List  []TeachingScheduleStudentCandidateVO `json:"list"`
	Total int                                  `json:"total"`
}

type TeachingScheduleStudentsAddCurrentDTO struct {
	ScheduleID  string   `json:"scheduleId"`
	StudentType int      `json:"studentType"`
	StudentIDs  []string `json:"studentIds"`
}

type TeachingScheduleBatchUpdateDTO struct {
	BatchNo              string   `json:"batchNo"`
	IDs                  []string `json:"ids"`
	TeacherID            string   `json:"teacherId"`
	AssistantIDs         []string `json:"assistantIds"`
	ClassroomID          string   `json:"classroomId"`
	LessonDate           string   `json:"lessonDate"`
	StartTime            string   `json:"startTime"`
	EndTime              string   `json:"endTime"`
	AllowStudentConflict bool     `json:"allowStudentConflict,omitempty"`
}

type TeachingScheduleBatchDetailQueryDTO struct {
	BatchNo string   `json:"batchNo"`
	IDs     []string `json:"ids,omitempty"`
}

type TeachingScheduleBatchDetailVO struct {
	BatchNo           string                     `json:"batchNo,omitempty"`
	BatchSize         int                        `json:"batchSize"`
	ClassType         int                        `json:"classType"`
	TeachingClassID   string                     `json:"teachingClassId"`
	TeachingClassName string                     `json:"teachingClassName"`
	StudentID         string                     `json:"studentId"`
	StudentName       string                     `json:"studentName"`
	LessonID          string                     `json:"lessonId"`
	LessonName        string                     `json:"lessonName"`
	BatchMeta         *TeachingScheduleBatchMeta `json:"batchMeta,omitempty"`
	Schedules         []TeachingScheduleVO       `json:"schedules"`
}

type TeachingScheduleDetailStudentVO struct {
	StudentID               string `json:"studentId"`
	StudentName             string `json:"studentName"`
	AvatarURL               string `json:"avatarUrl,omitempty"`
	Phone                   string `json:"phone,omitempty"`
	MaskedPhone             string `json:"maskedPhone,omitempty"`
	PhoneRelationship       int    `json:"phoneRelationship,omitempty"`
	PhoneRelationshipText   string `json:"phoneRelationshipText,omitempty"`
	ScheduleStudentType     int    `json:"scheduleStudentType"`
	ScheduleStudentTypeText string `json:"scheduleStudentTypeText,omitempty"`
	ClassStatus             int    `json:"classStatus"`
	ClassStatusText         string `json:"classStatusText,omitempty"`
	CallStatus              int    `json:"callStatus"`
	CallStatusText          string `json:"callStatusText,omitempty"`
}

type TeachingScheduleDetailVO struct {
	ID                     string                            `json:"id"`
	BatchNo                string                            `json:"batchNo,omitempty"`
	BatchSize              int                               `json:"batchSize"`
	ClassType              int                               `json:"classType"`
	TeachingClassID        string                            `json:"teachingClassId"`
	TeachingClassName      string                            `json:"teachingClassName"`
	LessonID               string                            `json:"lessonId"`
	LessonName             string                            `json:"lessonName"`
	TeacherID              string                            `json:"teacherId"`
	TeacherName            string                            `json:"teacherName"`
	AssistantIDs           []string                          `json:"assistantIds,omitempty"`
	AssistantNames         []string                          `json:"assistantNames,omitempty"`
	ClassroomID            string                            `json:"classroomId,omitempty"`
	ClassroomName          string                            `json:"classroomName,omitempty"`
	LessonDate             string                            `json:"lessonDate"`
	StartAt                time.Time                         `json:"startAt"`
	EndAt                  time.Time                         `json:"endAt"`
	DurationMinutes        int                               `json:"durationMinutes"`
	CallStatus             int                               `json:"callStatus"`
	CallStatusText         string                            `json:"callStatusText,omitempty"`
	CanRollCall            bool                              `json:"canRollCall"`
	RollCallDisabledReason string                            `json:"rollCallDisabledReason,omitempty"`
	Remark                 string                            `json:"remark,omitempty"`
	BatchMeta              *TeachingScheduleBatchMeta        `json:"batchMeta,omitempty"`
	Students               []TeachingScheduleDetailStudentVO `json:"students"`
	LeaveStudents          []TeachingScheduleDetailStudentVO `json:"leaveStudents,omitempty"`
}

type TeachingScheduleBatchReplaceDTO struct {
	BatchNo                string                          `json:"batchNo"`
	IDs                    []string                        `json:"ids"`
	OneToOneID             string                          `json:"oneToOneId,omitempty"`
	GroupClassID           string                          `json:"groupClassId,omitempty"`
	TeacherID              string                          `json:"teacherId"`
	AssistantIDs           []string                        `json:"assistantIds"`
	ClassroomID            string                          `json:"classroomId"`
	AllowStudentConflict   bool                            `json:"allowStudentConflict,omitempty"`
	AllowClassroomConflict bool                            `json:"allowClassroomConflict,omitempty"`
	BatchMeta              *TeachingScheduleBatchMeta      `json:"batchMeta,omitempty"`
	Schedules              []TeachingScheduleCreateSlotDTO `json:"schedules"`
}

type TeachingScheduleCancelDTO struct {
	IDs []string `json:"ids"`
}

type TeachingScheduleScopedCancelDTO struct {
	ID    string `json:"id"`
	Scope string `json:"scope"`
}

type TeachingScheduleCancelResult struct {
	Canceled int `json:"canceled"`
}

// TeachingScheduleCopyWeekDTO 将源日期区间内的课表按「星期对齐」复制到目标区间（两区间须同为连续日历天且天数一致）。
type TeachingScheduleCopyWeekDTO struct {
	SourceStartDate string `json:"sourceStartDate"`
	SourceEndDate   string `json:"sourceEndDate"`
	TargetStartDate string `json:"targetStartDate"`
	TargetEndDate   string `json:"targetEndDate"`
	// ScheduleTypes 可选；支持 group_class / one_to_one / trial，传入时优先按课表业务类型复制。
	ScheduleTypes []string `json:"scheduleTypes,omitempty"`
	// ClassType 兼容旧调用；当未传 ScheduleTypes 时生效。省略时仅复制 1 对 1。
	ClassType *int `json:"classType,omitempty"`
}

type TeachingScheduleCopyWeekResult struct {
	Created int `json:"created"`
}

// TeachingScheduleCopyDayDTO 将源日期下的老师课表复制到目标日期。
// TeacherIDs 为空时复制当前条件下源日期的全部老师课表；传入单个老师时复制该老师在矩阵中可见的当天课表。
// 若目标日期存在任意有效日程，则整次复制终止。
type TeachingScheduleCopyDayDTO struct {
	SourceDate         string   `json:"sourceDate"`
	TargetDate         string   `json:"targetDate"`
	StudentID          string   `json:"studentId,omitempty"`
	ScheduleTeacherIDs []string `json:"scheduleTeacherIds,omitempty"`
	ClassroomIDs       []string `json:"classroomIds,omitempty"`
	GroupClassIDs      []string `json:"groupClassIds,omitempty"`
	OneToOneClassIDs   []string `json:"oneToOneClassIds,omitempty"`
	LessonIDs          []string `json:"lessonIds,omitempty"`
	ScheduleTypes      []string `json:"scheduleTypes,omitempty"`
	CallStatuses       []string `json:"callStatuses,omitempty"`
}

type TeachingScheduleCopyDayResult struct {
	Created int `json:"created"`
}

type TeachingScheduleVO struct {
	ID                     string    `json:"id"`
	BatchNo                string    `json:"batchNo,omitempty"`
	BatchSize              int       `json:"batchSize"`
	ClassType              int       `json:"classType"`
	TeachingClassID        string    `json:"teachingClassId"`
	TeachingClassName      string    `json:"teachingClassName"`
	StudentID              string    `json:"studentId"`
	StudentName            string    `json:"studentName"`
	LessonID               string    `json:"lessonId"`
	LessonName             string    `json:"lessonName"`
	TeacherID              string    `json:"teacherId"`
	TeacherName            string    `json:"teacherName"`
	AssistantIDs           []string  `json:"assistantIds,omitempty"`
	AssistantNames         []string  `json:"assistantNames,omitempty"`
	ClassroomID            string    `json:"classroomId"`
	ClassroomName          string    `json:"classroomName"`
	LessonDate             string    `json:"lessonDate"`
	StartAt                time.Time `json:"startAt"`
	EndAt                  time.Time `json:"endAt"`
	Status                 int       `json:"status"`
	CallStatus             int       `json:"callStatus"`
	CallStatusText         string    `json:"callStatusText,omitempty"`
	CanRollCall            bool      `json:"canRollCall"`
	RollCallDisabledReason string    `json:"rollCallDisabledReason,omitempty"`
	Conflict               bool      `json:"conflict"`
	ConflictTypes          []string  `json:"conflictTypes,omitempty"`
}

type CreateOneToOneSchedulesResult struct {
	BatchNo string               `json:"batchNo,omitempty"`
	Count   int                  `json:"count"`
	List    []TeachingScheduleVO `json:"list"`
}

type TeachingScheduleValidationResult struct {
	Valid             bool                             `json:"valid"`
	Message           string                           `json:"message,omitempty"`
	CurrentSchedules  []TeachingScheduleConflictItem   `json:"currentSchedules,omitempty"`
	ExistingSchedules []TeachingScheduleConflictItem   `json:"existingSchedules,omitempty"`
	ConflictTypes     []string                         `json:"conflictTypes,omitempty"`
	Items             []TeachingScheduleValidationItem `json:"items,omitempty"`
}

type TeachingScheduleValidationItem struct {
	TeacherID               string                         `json:"teacherId,omitempty"`
	LessonDate              string                         `json:"lessonDate"`
	StartTime               string                         `json:"startTime"`
	EndTime                 string                         `json:"endTime"`
	Valid                   bool                           `json:"valid"`
	Message                 string                         `json:"message,omitempty"`
	ExistingSchedules       []TeachingScheduleConflictItem `json:"existingSchedules,omitempty"`
	ConflictTypes           []string                       `json:"conflictTypes,omitempty"`
	ConflictingStudentNames []string                       `json:"conflictingStudentNames,omitempty"`
}

type OneToOneScheduleAvailabilityItem struct {
	TeacherID         string                         `json:"teacherId"`
	LessonDate        string                         `json:"lessonDate"`
	StartTime         string                         `json:"startTime"`
	EndTime           string                         `json:"endTime"`
	Valid             bool                           `json:"valid"`
	Message           string                         `json:"message,omitempty"`
	ExistingSchedules []TeachingScheduleConflictItem `json:"existingSchedules,omitempty"`
	ConflictTypes     []string                       `json:"conflictTypes,omitempty"`
}

type OneToOneScheduleAvailabilityResult struct {
	ValidCount   int                                `json:"validCount"`
	InvalidCount int                                `json:"invalidCount"`
	Items        []OneToOneScheduleAvailabilityItem `json:"items"`
}

type TeachingScheduleConflictItem struct {
	Date           string   `json:"date"`
	Week           string   `json:"week,omitempty"`
	Name           string   `json:"name"`
	ClassTypeText  string   `json:"classTypeText"`
	TimeText       string   `json:"timeText"`
	TeacherID      string   `json:"teacherId,omitempty"`
	TeacherName    string   `json:"teacherName"`
	AssistantNames []string `json:"assistantNames,omitempty"`
	ClassroomName  string   `json:"classroomName,omitempty"`
	StudentNames   []string `json:"studentNames,omitempty"`
	ConflictTypes  []string `json:"conflictTypes,omitempty"`
}

type AssistantScheduleAvailabilityItem struct {
	AssistantID       string                         `json:"assistantId"`
	AssistantName     string                         `json:"assistantName,omitempty"`
	Valid             bool                           `json:"valid"`
	Message           string                         `json:"message,omitempty"`
	ExistingSchedules []TeachingScheduleConflictItem `json:"existingSchedules,omitempty"`
	ConflictTypes     []string                       `json:"conflictTypes,omitempty"`
}

type AssistantScheduleAvailabilityResult struct {
	ValidCount   int                                 `json:"validCount"`
	InvalidCount int                                 `json:"invalidCount"`
	Items        []AssistantScheduleAvailabilityItem `json:"items"`
}

type OneToOneScheduleCreateContext struct {
	ClassID            int64
	ClassName          string
	StudentID          int64
	StudentName        string
	LessonID           int64
	LessonName         string
	Status             int
	ClassStudentStatus int
}

type GroupClassScheduleCreateContext struct {
	ClassID            int64
	ClassName          string
	LessonID           int64
	LessonName         string
	Status             int
	DefaultTeacherID   int64
	DefaultTeacherName string
	ClassroomID        int64
	ClassroomName      string
}

// InstUserScheduleRosterItem 课表矩阵机构在职人员（未禁用的机构用户）
type InstUserScheduleRosterItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// TeachingScheduleMatrixDayVO 按「日期 → 教师分列」矩阵（对齐旧版 scheduleListVoList 结构）
type TeachingScheduleMatrixDayVO struct {
	ScheduleDate       string                          `json:"scheduleDate"`
	Width              int                             `json:"width"`
	ScheduleInfoVoList any                             `json:"scheduleInfoVoList"` // 旧版字段，输出 null
	ScheduleListVoList []TeachingScheduleMatrixTeacher `json:"scheduleListVoList"`
}

// TeachingScheduleMatrixTeacher 单日下单个教师列
type TeachingScheduleMatrixTeacher struct {
	TeacherName        string                         `json:"teacherName"`
	TeacherID          int64                          `json:"teacherId"`
	ScheduleInfoVoList []TeachingScheduleInfoLegacyVO `json:"scheduleInfoVoList"`
}

// ScheduleLegacyPersonVO 旧版 teacherList / studentList 元素
type ScheduleLegacyPersonVO struct {
	Name     string `json:"name"`
	ID       int64  `json:"id"`
	Type     int    `json:"type"`
	Disabled bool   `json:"disabled,omitempty"`
	UUID     any    `json:"uuid,omitempty"`
	Version  any    `json:"version,omitempty"`
}

// TeachingScheduleInfoLegacyVO 旧版日程明细（在能力范围内从 TeachingScheduleVO 映射）
type TeachingScheduleInfoLegacyVO struct {
	ID                     int64                    `json:"id"`
	UUID                   string                   `json:"uuid,omitempty"`
	Version                int64                    `json:"version,omitempty"`
	CreateTime             string                   `json:"createTime,omitempty"`
	UpdateTime             *string                  `json:"updateTime"`
	InstID                 int64                    `json:"instId,omitempty"`
	BatchNo                string                   `json:"batchNo,omitempty"`
	BatchID                int64                    `json:"batchId,omitempty"`
	ModifyBatchID          int64                    `json:"modifyBatchId,omitempty"`
	CourseID               int64                    `json:"courseId,omitempty"`
	ClassID                *int64                   `json:"classId"`
	OriginID               *int64                   `json:"originId"`
	ScheduleDate           string                   `json:"scheduleDate"`
	ScheduleStartTime      string                   `json:"scheduleStartTime"`
	ScheduleEndTime        string                   `json:"scheduleEndTime"`
	ScheduleStatus         int                      `json:"scheduleStatus"`
	CallStatus             int                      `json:"callStatus"`
	CallStatusText         string                   `json:"callStatusText,omitempty"`
	CanRollCall            bool                     `json:"canRollCall"`
	RollCallDisabledReason string                   `json:"rollCallDisabledReason,omitempty"`
	Conflict               bool                     `json:"conflict"`
	ConflictTypes          []string                 `json:"conflictTypes,omitempty"`
	MissSchedule           bool                     `json:"missSchedule"`
	HasIgnore              *bool                    `json:"hasIgnore"`
	CourseStatus           int                      `json:"courseStatus"`
	Remark                 *string                  `json:"remark"`
	Width                  int                      `json:"width"`
	TeacherList            []ScheduleLegacyPersonVO `json:"teacherList"`
	StudentList            []ScheduleLegacyPersonVO `json:"studentList"`
	ClassroomID            string                   `json:"classroomId,omitempty"`
	ClassroomName          string                   `json:"classroomName,omitempty"`
	CourseName             string                   `json:"courseName"`
	CourseType             int                      `json:"courseType"`
	ClassName              string                   `json:"className,omitempty"`
	CrossOver              any                      `json:"crossOver"`
	InfoVoList             any                      `json:"infoVoList"`
	LeaveList              []any                    `json:"leaveList"`
	InstName               string                   `json:"instName,omitempty"`
	CourseTime             int                      `json:"courseTime"`
	CourseHour             int                      `json:"courseHour"`
	FinishType             int                      `json:"finishType"`
}
