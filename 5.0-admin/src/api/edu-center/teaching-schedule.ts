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
