import axios from 'axios'
import { STORAGE_AUTHORIZE_KEY, useAuthorization } from '~/composables/authorization'
import { useGet, usePost } from '~/utils/request'

export interface TeachingScheduleItem {
  id: string
  batchNo?: string
  batchSize?: number
  classType: number
  teachingClassId: string
  teachingClassName: string
  studentId: string
  studentName: string
  lessonId: string
  lessonName: string
  teacherId: string
  teacherName: string
  assistantIds?: string[]
  assistantNames?: string[]
  classroomId?: string
  classroomName?: string
  lessonDate: string
  startAt: string
  endAt: string
  status: number
}

export interface CreateOneToOneSchedulesResult {
  batchNo?: string
  count: number
  list: TeachingScheduleItem[]
}

export interface TeachingScheduleValidationResult {
  valid: boolean
  message?: string
  currentSchedules?: Array<{
    name: string
    classTypeText: string
    date: string
    week?: string
    timeText: string
    teacherName: string
    classroomName?: string
    studentNames?: string[]
    conflictTypes?: string[]
  }>
  existingSchedules?: Array<{
    name: string
    classTypeText: string
    date: string
    week?: string
    timeText: string
    teacherName: string
    classroomName?: string
    studentNames?: string[]
    conflictTypes?: string[]
  }>
  conflictTypes?: string[]
}

export function createOneToOneSchedulesApi(data: {
  oneToOneId: string
  teacherId: string
  assistantIds?: string[]
  classroomId?: string
  schedules: Array<{
    lessonDate: string
    startTime: string
    endTime: string
  }>
}) {
  return usePost<CreateOneToOneSchedulesResult>('/api/v1/teaching-schedules/one-to-one/create', data)
}

export function validateOneToOneSchedulesApi(data: {
  oneToOneId: string
  teacherId: string
  assistantIds?: string[]
  classroomId?: string
  schedules: Array<{
    lessonDate: string
    startTime: string
    endTime: string
  }>
}) {
  return usePost<TeachingScheduleValidationResult>('/api/v1/teaching-schedules/one-to-one/validate', data)
}

export function listTeachingSchedulesApi(params: {
  startDate: string
  endDate: string
  classType?: number
}) {
  return useGet<TeachingScheduleItem[]>('/api/v1/teaching-schedules', params)
}

/** 旧版机构总课表矩阵接口返回的教师列 */
export interface TeachingScheduleMatrixTeacherColumn {
  teacherName: string
  teacherId: number
  scheduleInfoVoList: TeachingScheduleMatrixLegacyItem[]
}

export interface TeachingScheduleMatrixLegacyItem {
  id: number
  scheduleDate: string
  scheduleStartTime: string
  scheduleEndTime: string
  scheduleStatus?: number
  courseStatus?: number
  courseType?: number
  courseName?: string
  className?: string | null
  classId?: number | null
  courseId?: number
  batchId?: number
  teacherList?: Array<{ name: string, id: number, type?: number, disabled?: boolean }>
  studentList?: Array<{ name: string, id: number, type?: number }>
  instId?: number
  width?: number
  courseTime?: number
  courseHour?: number
  finishType?: number
  leaveList?: unknown[]
}

/** 按「日期 × 教师」矩阵 */
export interface TeachingScheduleMatrixDay {
  scheduleDate: string
  width: number
  scheduleInfoVoList?: null
  scheduleListVoList: TeachingScheduleMatrixTeacherColumn[]
}

export type MatrixTeacherFilterParam = 'all' | 'has_class' | 'no_class'

export function listTeachingSchedulesByTeacherMatrixApi(params: {
  startDate: string
  endDate: string
  classType?: number
  /** 逗号分隔 1–7（周一…周日），省略或全开则不传以缩短 URL */
  weekdays?: string
  /** 教师列：仅有课 / 仅无课，与旧版课表展示配置一致 */
  teacherFilter?: Exclude<MatrixTeacherFilterParam, 'all'>
}) {
  return useGet<TeachingScheduleMatrixDay[]>(
    '/api/v1/teaching-schedules/by-teacher-matrix',
    params,
  )
}

/** 导出教师矩阵课表 Excel（每位教师一个 Sheet），查询参数与矩阵列表一致 */
export async function downloadTeachingSchedulesTeacherMatrixExcelApi(params: {
  startDate: string
  endDate: string
  classType?: number
  weekdays?: string
  teacherFilter?: Exclude<MatrixTeacherFilterParam, 'all'>
}) {
  const token = useAuthorization()
  return axios.get('/api/v1/teaching-schedules/by-teacher-matrix/export', {
    params,
    responseType: 'blob',
    headers: {
      [STORAGE_AUTHORIZE_KEY]: token.value || '',
      Authorization: token.value ? `Bearer ${token.value}` : '',
      'Accept-Language': 'zh-CN',
    },
  })
}

export function batchUpdateTeachingSchedulesApi(data: {
  batchNo?: string
  ids?: string[]
  teacherId: string
  assistantIds?: string[]
  classroomId?: string
  startTime: string
  endTime: string
}) {
  return usePost<boolean>('/api/v1/teaching-schedules/batch-update', data)
}

/** 将源周课表按天对齐复制到目标周；源 batch 在目标周使用新 batchNo，batchSize 与复制条数一致 */
export function copyTeachingSchedulesWeekApi(data: {
  sourceStartDate: string
  sourceEndDate: string
  targetStartDate: string
  targetEndDate: string
  /** 省略时后端默认仅复制 1 对 1（classType=2） */
  classType?: number
}) {
  return usePost<{ created: number }>('/api/v1/teaching-schedules/copy-week', data)
}
