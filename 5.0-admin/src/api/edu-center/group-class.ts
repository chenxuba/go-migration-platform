import { usePost } from '~/utils/request'

/** 对标 CheckClassName：true = 名称已存在 */
export function checkGroupClassNameApi(data: {
  name: string
  isOne2One: boolean
  exceptId?: string
}) {
  return usePost<boolean>('/api/v1/group-classes/check-name', data)
}

/** 对标 Create 集体班 */
export function createGroupClassApi(data: {
  name: string
  lessonId: string
  maxCount: number
  teacherIds: string[]
  defaultTeacherId: string
  defaultStudentClassTime: number
  defaultTeacherClassTime: number
  defaultClassTimeRecordMode: number
  isCopyStudent?: boolean
  copiedStudents?: unknown[]
  isCopyTimetable?: boolean
  classProperties?: unknown[]
  remark?: string
}) {
  // 业务错误用 HTTP 400，需当作「有 body 的成功响应」解析，否则会进 reject 且看不到 message
  return usePost<{ id: string; name: string }>('/api/v1/group-classes/create', data, {
    validateStatus: status => (status >= 200 && status < 300) || status === 400,
  })
}

/** 对标 QueryClassList */
export function pageGroupClassesApi(data: {
  queryModel: Record<string, unknown>
  pageRequestModel: {
    needTotal?: boolean
    pageSize: number
    pageIndex: number
    skipCount?: number
  }
}) {
  return usePost<{ list: GroupClassRow[]; total: number }>('/api/v1/group-classes/page', data)
}

/** 对标 QueryClassStatisticsInfo（请求体与 queryModel 字段一致） */
export function groupClassStatisticsApi(queryModel: Record<string, unknown>) {
  return usePost<{
    classCount: number
    openClassCount: number
    studentCount: number
    studentPersonTime: number
  }>('/api/v1/group-classes/statistics', queryModel)
}

export interface GroupClassTeacher {
  id: string
  name: string
  status: number
  avatar?: string
}

export interface GroupClassRow {
  id: string
  name: string
  classTime: number
  lessonId: string
  lessonName: string
  isMultiProduct: boolean
  studentCount: number
  lockStudentCount: number
  maxCount: number
  teachers: GroupClassTeacher[]
  defaultTeacherId: string
  defaultTeacherName: string
  classRoomName: string
  classLessonTimes: unknown[]
  isScheduled: boolean
  classLessonDayInfos: {
    lessonDayCount: number
    completeLessonDayCount: number
  }
  status: number
  closedTime: string
  createdTime: string
  createdStaffName: string
  remark: string
  classProperties: unknown[]
}
