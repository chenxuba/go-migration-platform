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
