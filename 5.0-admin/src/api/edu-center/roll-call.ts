import { usePost } from '~/utils/request'
import type { TeachingScheduleItem } from './teaching-schedule'

export interface RollCallQueryModel {
  startDate?: string
  endDate?: string
  lessonId?: string
  classroomId?: string
  classId?: string
  oneToOneId?: string
  teacherId?: string
  teacherTypes?: number[]
  scheduleTypes?: string[]
}

export interface RollCallStatisticsResult {
  todayCount: number
  allCount: number
  partialCount: number
}

export interface RollCallPagedListParams {
  queryModel: RollCallQueryModel
  pageRequestModel: {
    needTotal?: boolean
    pageSize: number
    pageIndex: number
    skipCount?: number
  }
  sortModel?: {
    byStartDate?: number
  }
}

export interface RollCallPagedListResult {
  list?: TeachingScheduleItem[]
  total?: number
}

export function getRollCallStatisticsApi(data: { queryModel?: RollCallQueryModel }) {
  return usePost<RollCallStatisticsResult>('/api/v1/roll-call/statistics', data)
}

export function getRollCallPagedListApi(data: RollCallPagedListParams) {
  return usePost<RollCallPagedListResult>('/api/v1/roll-call/paged-list', data)
}
