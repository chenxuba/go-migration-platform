export interface OneToOneTeacher {
  teacherId?: string
  name?: string
  status?: number
  classId?: string
  /** 是否为默认上课教师对应行（合并进 teaching_class_teacher） */
  isDefault?: boolean
}

export interface OneToOneLessonDayInfo {
  lessonDayCount?: number
  completeLessonDayCount?: number
}

export interface OneToOneTuitionAccount {
  id?: string
  totalTuition?: number
  remainTuition?: number
  totalQuantity?: number
  totalFreeQuantity?: number
  remainQuantity?: number
  remainFreeQuantity?: number
  lessonChargingMode?: number
  productName?: string
  status?: number
  enableExpireTime?: boolean
  lastSuspendedTime?: string
  expireTime?: string
  studentId?: string
  lessonId?: string
  lessonType?: number
  changeStatusTime?: string
  suspendedTime?: string
  classEndingTime?: string
  assignedClass?: boolean
  /** 为 true 时 id 为 agg:{courseId}:{teachMethod}:{lessonModel}，多笔在读账户按计费桶合并展示 */
  isAggregate?: boolean
}

export interface OneToOneItem {
  id?: string
  name?: string
  studentName?: string
  studentId?: string
  sex?: number
  avatar?: string
  phone?: string
  schoolId?: string
  one2OneLessonTimes?: Array<Record<string, any>>
  isScheduled?: boolean
  status?: number
  classStudentStatus?: number
  one2OneLessonDayInfo?: OneToOneLessonDayInfo
  createdTime?: string
  classRoomId?: string
  classRoomName?: string
  classroomEnabled?: boolean | null
  classTime?: number
  studentClassTime?: number
  teacherClassTime?: number
  lessonId?: string
  lessonName?: string
  tuitionAccountCount?: number
  tuitionAccountId?: string
  defaultTeacherId?: string
  defaultTeacherName?: string
  defaultClassTimeRecordMode?: number
  isGradeUpgrade?: boolean
  lastFinishedLessonDay?: string
  teacherList?: OneToOneTeacher[]
  tuitionAccount?: OneToOneTuitionAccount
  classProperties?: Array<Record<string, any>>
  classTeacherId?: string
  classTeacherName?: string
  remark?: string
}

export interface OneToOneDetail {
  id?: string
  studentId?: string
  schoolId?: string
  name?: string
  studentName?: string
  studentAvatar?: string
  studentGender?: number
  lessonId?: string
  lessonName?: string
  lessonPrice?: number
  classroomId?: string
  classroomName?: string | null
  tuitionAccountCount?: number
  tuitionAccountId?: string
  classTime?: number
  isScheduled?: boolean
  classroomEnabled?: boolean | null
  status?: number
  classStudentStatus?: number
  createdTime?: string
  defaultStudentClassTime?: number
  defaultTeacherClassTime?: number
  defaultClassTimeRecordMode?: number
  defaultTeacherId?: string
  defaultTeacherName?: string
  classTeacherName?: string
  isGradeUpgrade?: boolean
  remark?: string
  teacherList?: OneToOneTeacher[]
  tuitionAccount?: OneToOneTuitionAccount
  createdStaffId?: string
  createdStaffName?: string
  classProperties?: Array<Record<string, any>>
  defaultTeacherStatus?: number
}

export interface OneToOneListResult {
  total?: number
  studentCount?: number
  list?: OneToOneItem[]
}

export interface OneToOneListParams {
  pageRequestModel: {
    needTotal?: boolean
    pageSize: number
    pageIndex: number
    skipCount?: number
  }
  queryModel?: {
    studentId?: string
    lessonIds?: string[]
    classTeacherId?: string
    defaultTeacherId?: string
    hasClassTeacher?: boolean
    isScheduled?: boolean
    status?: number[]
    classStudentStatus?: number[]
    startDate?: string
    endDate?: string
  }
}

export interface OneToOneBatchAssignTeacherParams {
  ids: string[]
  /** 多选班主任（优先） */
  classTeacherIds?: string[]
  /** 兼容旧版单选 */
  classTeacherId?: string
}

export interface OneToOneBatchClassTimeParams {
  ids: string[]
  classTime: number
  studentClassTime: number
  teacherClassTime: number
  /** 1 按固定课时记录 2 按上课时长记录 */
  classTimeRecordMode?: number
}

export interface OneToOneCheckNameParams {
  name: string
  exceptId?: string
  isOne2One: boolean
}

export interface OneToOneCloseParams {
  id: string
}

export interface OneToOneSwitchDefaultTuitionAccountParams {
  id: string
  tuitionAccountId: string
}

export interface OneToOneUpdateParams {
  id: string
  studentId: string
  lessonId: string
  classroomId?: string
  name: string
  teacherId: string[]
  defaultTeacherId?: string
  defaultStudentClassTime: number
  defaultTeacherClassTime: number
  defaultClassTimeRecordMode: number
  remark?: string
  classProperties: Array<Record<string, any>>
  /** 为 true 时在名称与其他班级重复时仍保存（须先经前端二次确认） */
  allowDuplicateName?: boolean
}

/** 手动创建 1 对 1，须指定在读学费账户（与上课课程一致） */
export interface OneToOneCreateParams {
  studentId: string
  lessonId: string
  classroomId?: string
  tuitionAccountId: string
  name: string
  teacherId: string[]
  defaultTeacherId?: string
  defaultStudentClassTime: number
  defaultTeacherClassTime: number
  defaultClassTimeRecordMode: number
  remark?: string
  classProperties: Array<Record<string, any>>
  allowDuplicateName?: boolean
}

export interface OneToOneCreateResult {
  id: string
}

/** 学员在某课程下的学费账户（对齐后端 / 竞品 GetStudentAllTuitionAccountByLessonId） */
export interface StudentLessonTuitionAccountItem {
  id?: string
  studentId?: string
  lessonId?: string
  /** 账户所属课程名称（创建 1 对 1 扣费下拉用） */
  lessonName?: string
  productName?: string
  lessonChargingMode?: number
  totalQuantity?: number
  totalFreeQuantity?: number
  totalTuition?: number
  freeQuantity?: number
  quantity?: number
  tuition?: number
  suspended?: boolean
  suspendedTime?: string
  startTime?: string
  enableExpireTime?: boolean
  expireTime?: string
  assignedClass?: boolean
  latestStartTime?: string
  lessonType?: number
  isTuitionAccountActive?: boolean
  status?: number
}

export interface StudentLessonTuitionAccountsResult {
  list?: StudentLessonTuitionAccountItem[]
}

/** 对标 QueryOne2OneLessonByStudentId：学员可开 1 对 1 的课程（有学费账户且为 1v1 课程） */
export interface OneToOneLessonOption {
  id?: string
  name?: string
  /** 已分班或已有开班中 1 对 1，下拉展示「已报名」 */
  alreadyEnrolled?: boolean
}

export interface OneToOneLessonsByStudentResult {
  list?: OneToOneLessonOption[]
}

export function listOneToOneLessonsByStudentApi(data: {
  studentId: string
  /** 不传时后端默认按 status=1（在读/有效） */
  tuitionAccountStatus?: number[]
}) {
  return usePost<OneToOneLessonsByStudentResult>('/api/v1/one-to-ones/lessons-by-student', data)
}

export function getOneToOneListApi(data: OneToOneListParams) {
  return usePost<OneToOneListResult>('/api/v1/one-to-ones/page', data)
}

export function getOneToOneByIdApi(id: string | number) {
  return useGet<OneToOneDetail>('/api/v1/one-to-ones/detail', { id })
}

export function batchAssignOneToOneClassTeacherApi(data: OneToOneBatchAssignTeacherParams) {
  return usePost('/api/v1/one-to-ones/batch-assign-class-teacher', data)
}

export function batchUpdateOneToOneClassTimeApi(data: OneToOneBatchClassTimeParams) {
  return usePost('/api/v1/one-to-ones/batch-update-class-time', data)
}

export function checkOneToOneNameApi(data: OneToOneCheckNameParams) {
  return usePost<boolean>('/api/v1/one-to-ones/check-name', data)
}

/** 对标 SchoolPal ExistOne2One：result/data 为 true 表示该学员在该课程下已有开班中的 1 对 1 */
export function existOneToOneApi(data: { studentId: string, lessonId: string }) {
  return usePost<boolean>('/api/v1/one-to-ones/exist', data)
}

export function updateOneToOneApi(data: OneToOneUpdateParams) {
  return usePost<boolean>('/api/v1/one-to-ones/update', data)
}

export function switchOneToOneDefaultTuitionAccountApi(data: OneToOneSwitchDefaultTuitionAccountParams) {
  return usePost<boolean>('/api/v1/one-to-ones/switch-default-tuition-account', data)
}

export function createOneToOneApi(data: OneToOneCreateParams) {
  return usePost<OneToOneCreateResult>('/api/v1/one-to-ones/create', data)
}

export function closeOneToOneApi(data: OneToOneCloseParams) {
  return usePost<boolean>('/api/v1/one-to-ones/close', data)
}

/** 恢复开班，请求体与结班相同：{ id } */
export function reopenOneToOneApi(data: OneToOneCloseParams) {
  return usePost<boolean>('/api/v1/one-to-ones/reopen', data)
}

export function listTuitionAccountsByStudentAndLessonApi(data: {
  studentId: string
  lessonId: string
  /** 传当前 1 对 1 班级 id 时，补齐该班级已绑定的扣费账户，避免同学员历史班级串数 */
  teachingClassId?: string
  /** 有值且非 0 时只查该订单明细下的学费账户（1 对 1 详情报读明细） */
  orderCourseDetailId?: string
}) {
  return usePost<StudentLessonTuitionAccountsResult>('/api/v1/tuition-accounts/by-student-and-lesson', data)
}

/** 创建 1 对 1 选扣费账户：学员名下全部在读报读账户（班级授课或 1v1，不限当前所选上课课程） */
export function listTuitionAccountsForOneToOneDeductionApi(data: { studentId: string }) {
  return usePost<StudentLessonTuitionAccountsResult>('/api/v1/tuition-accounts/for-one-to-one-deduction', data)
}
