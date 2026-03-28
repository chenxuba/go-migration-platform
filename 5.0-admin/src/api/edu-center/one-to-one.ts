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
  isGradeUpgrade?: boolean
  lastFinishedLessonDay?: string
  teacherList?: OneToOneTeacher[]
  tuitionAccount?: OneToOneTuitionAccount
  classProperties?: Array<Record<string, any>>
  classTeacherId?: string
  classTeacherName?: string
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

export interface OneToOneBatchAttributeParams {
  ids: string[]
  defaultTeacherId?: string
  status?: number
  classStudentStatus?: number
}

export function getOneToOneListApi(data: OneToOneListParams) {
  return usePost<OneToOneListResult>('/api/v1/one-to-ones/page', data)
}

export function batchAssignOneToOneClassTeacherApi(data: OneToOneBatchAssignTeacherParams) {
  return usePost('/api/v1/one-to-ones/batch-assign-class-teacher', data)
}

export function batchUpdateOneToOneClassTimeApi(data: OneToOneBatchClassTimeParams) {
  return usePost('/api/v1/one-to-ones/batch-update-class-time', data)
}

export function batchUpdateOneToOneAttributesApi(data: OneToOneBatchAttributeParams) {
  return usePost('/api/v1/one-to-ones/batch-update-attributes', data)
}
