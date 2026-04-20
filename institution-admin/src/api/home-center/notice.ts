import { useGet, usePost } from '~/utils/request'

export interface NoticeTemplateItem {
  id: string
  title: string
  coverUrl: string
  tag: string
  weight: number
  content: string
  summary: string
  orgId: string
  schoolId: string
}

export interface NoticeClassItem {
  classId: string
  noticeId: string
  className: string
}

export interface NoticeListItem {
  noticeId: string
  title: string
  content: string
  summary: string
  isAllSchool: boolean
  isConfirm: boolean
  isRemind: boolean
  isWithdraw: boolean
  classs: NoticeClassItem[]
  studentCount: number
  readStudentCount: number
  confirmStudentCount: number
  operatorId: string
  operatorName: string
  operationDate?: string | null
  isDelaySend: boolean
  publishTime?: string | null
  status: number
  realityPublishTime?: string | null
}

export interface NoticePageResult {
  list: NoticeListItem[]
  total: number
}

export interface NoticePageParams {
  pageRequestModel: {
    pageSize: number
    pageIndex: number
    skipCount?: number
  }
  queryModel: {
    statuses?: number[]
    isWithdraw?: boolean
    beginPublishDate?: string
    endPublishDate?: string
    operatorId?: string
  }
}

export interface NoticeCreateParams {
  noticeTemplateId?: string
  title: string
  content: string
  isAllSchool: boolean
  isDelaySend: boolean
  publishDate?: string
  hour?: number
  isConfirm: boolean
  summary: string
  classIds: string[]
  studentIds: string[]
}

export function listNoticeTemplatesApi() {
  return useGet<NoticeTemplateItem[]>('/api/v1/notices/templates')
}

export function checkNoticeFilterWordsApi(data: {
  title: string
  content: string
  summary: string
}) {
  return usePost<string[]>('/api/v1/notices/check-filter-word', data)
}

export function checkRepeatNoticeStudentApi(data: {
  studentIds: string[]
}) {
  return usePost<boolean>('/api/v1/notices/check-repeat-student', data)
}

export function createNoticeApi(data: NoticeCreateParams) {
  return usePost<string>('/api/v1/notices/create', data)
}

export function withdrawNoticeApi(data: { noticeId: string }) {
  return usePost<boolean>('/api/v1/notices/withdraw', data)
}

export function pageNoticesApi(data: NoticePageParams) {
  return usePost<NoticePageResult>('/api/v1/notices/paged-list', data)
}
