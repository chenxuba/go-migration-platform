import { useGet, usePost } from '~/utils/request'

export interface HomeworkAttachmentItem {
  type: number
  url: string
  duration?: number
  name?: string
  extendName?: string
}

export interface HomeworkRepeatRule {
  startDate: string
  endDate: string
  repeatSpan: number
  weekDays: number
}

export interface HomeworkStudentSelection {
  sourceType: number
  sourceId: string
  sourceName: string
  studentId: string
  studentName: string
  tuitionAccountId?: string
  isBind: boolean
}

export interface HomeworkObjectPayload {
  sourceType: number
  sourceId: string
  studentIds: string[]
}

export interface HomeworkMutationPayload {
  title: string
  content: string
  attachments: HomeworkAttachmentItem[]
  repeatRule: HomeworkRepeatRule | null
  publishTime?: string
  endTime?: string
  publishHour?: number
  endHour?: number
  taskDurationHours?: number
  isVisibleStudent?: boolean
  homeworkObjects: HomeworkObjectPayload[]
}

export interface HomeworkOperationResult {
  id: string
  name: string
}

export interface HomeworkListItem {
  id: string
  title: string
  content: string
  studentCount: number
  unsubmittedCount: number
  rejectedCount: number
  submittedCount: number
  reSubmittedCount: number
  evaluatedCount: number
  unevaluatedCount: number
  isVisibleStudent: boolean
  sourceType: number
  sourceId: string
  sourceName: string
  createdStaffId: string
  createdStaffName: string
  createdTime?: string
  publishTime?: string
  endTime?: string
  endHour?: number
  readCount: number
  unreadCount: number
  publishRule?: number
  repeatRule?: HomeworkRepeatRule | null
}

export interface HomeworkDetail extends HomeworkListItem {
  publishHour?: number
  taskDurationHours?: number
  attachments: HomeworkAttachmentItem[]
  selectedStudents: HomeworkStudentSelection[]
}

export interface HomeworkListParams {
  pageRequestModel: {
    needTotal?: boolean
    pageSize: number
    pageIndex: number
    skipCount?: number
  }
  sortModel?: {
    publishTime?: number
  }
  queryModel?: {
    teacherIds?: string[]
    classId?: string
    one2OneId?: string
    publishStartTime?: string
    publishEndTime?: string
    endStartTime?: string
    endEndTime?: string
    hasUnevaluated?: boolean
    hasUnsubmitted?: boolean
  }
}

export interface HomeworkPageResult {
  list: HomeworkListItem[]
  total: number
}

export interface HomeworkStatisticsResult {
  unsubmittedCount: number
  unevaluatedCount: number
}

export function batchCreateHomeworksApi(data: HomeworkMutationPayload) {
  return usePost<HomeworkOperationResult[]>('/api/v1/homeworks/batch-create', data)
}

export function getHomeworkDetailApi(id: string | number) {
  return useGet<HomeworkDetail>('/api/v1/homeworks/detail', { id })
}

export function updateHomeworkApi(data: HomeworkMutationPayload & { id: string }) {
  return usePost<HomeworkOperationResult>('/api/v1/homeworks/update', data)
}

export function deleteHomeworkApi(data: { id: string }) {
  return usePost<boolean>('/api/v1/homeworks/delete', data)
}

export function pageHomeworksApi(data: HomeworkListParams) {
  return usePost<HomeworkPageResult>('/api/v1/homeworks/paged-list', data)
}

export function homeworkStatisticsApi(data: HomeworkListParams['queryModel']) {
  return usePost<HomeworkStatisticsResult>('/api/v1/homeworks/statistics', data)
}
