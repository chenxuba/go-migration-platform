export interface OneToOneTeacher {
  teacherId?: string
  name?: string
  status?: number
  classId?: string
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
  lessonScopeModel?: number
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
  classTeacherId: string
}

export interface OneToOneBatchClassTimeParams {
  ids: string[]
  classTime: number
  studentClassTime: number
  teacherClassTime: number
}

export interface OneToOneCheckNameParams {
  name: string
  exceptId?: string
  isOne2One: boolean
}

export interface OneToOneUpdateParams {
  id: string
  studentId: string
  lessonId: string
  name: string
  teacherId: string[]
  defaultTeacherId?: string
  defaultStudentClassTime: number
  defaultTeacherClassTime: number
  defaultClassTimeRecordMode: number
  remark?: string
  classProperties: Array<Record<string, any>>
}

/** 学员在某课程下的学费账户（对齐后端 / 竞品 GetStudentAllTuitionAccountByLessonId） */
export interface StudentLessonTuitionAccountItem {
  id?: string
  studentId?: string
  lessonId?: string
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
  lessonScope?: number
  generalLessonIdList?: string[]
  latestStartTime?: string
  lessonType?: number
  isTuitionAccountActive?: boolean
  status?: number
}

export interface StudentLessonTuitionAccountsResult {
  list?: StudentLessonTuitionAccountItem[]
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

export function updateOneToOneApi(data: OneToOneUpdateParams) {
  return usePost<boolean>('/api/v1/one-to-ones/update', data)
}

export function listTuitionAccountsByStudentAndLessonApi(data: { studentId: string, lessonId: string }) {
  return usePost<StudentLessonTuitionAccountsResult>('/api/v1/tuition-accounts/by-student-and-lesson', data)
}
