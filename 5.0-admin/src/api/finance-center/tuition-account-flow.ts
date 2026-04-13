import axios from 'axios'
import { STORAGE_AUTHORIZE_KEY, useAuthorization } from '~/composables/authorization'
import { usePost } from '~/utils/request'

export interface TuitionAccountFlowRecordItem {
  tutionAccountFlowId: string
  tuitionAccountId: string
  studentId: string
  studentName: string
  studentPhone: string
  studentAvatar?: string
  teachingCourseId?: string
  teachingCourseName?: string
  productId: string
  productName: string
  lessonType?: number
  lessonChargingMode?: number
  sourceType: number
  sourceId: string
  teachingRecordId?: string
  createdTime?: string
  quantity: number
  tuition: number
}

export interface TuitionAccountFlowRecordListResult {
  list?: TuitionAccountFlowRecordItem[]
  total?: number
}

export interface SubTuitionAccountFlowRecordItem {
  id: string
  tuitionAccountId: string
  studentId: string
  studentName: string
  studentPhone: string
  studentAvatar?: string
  teachingCourseId?: string
  teachingCourseName?: string
  productId: string
  productName: string
  lessonType?: number
  lessonChargingMode?: number
  sourceType: number
  sourceId: string
  teachingRecordId?: string
  createdTime?: string
  quantity: number
  tuition: number
  balanceQuantity: number
  balanceTuition: number
  orderNumber?: string
}

export interface SubTuitionAccountFlowRecordListResult {
  list?: SubTuitionAccountFlowRecordItem[]
  total?: number
}

export interface TuitionAccountFlowRecordListQueryParams {
  queryModel?: {
    tuitionAccountId?: string
    productId?: string
    studentId?: string
    orderNumber?: string
    sourceTypes?: number[]
    startTime?: string
    endTime?: string
  }
  pageRequestModel: {
    needTotal?: boolean
    pageSize: number
    pageIndex: number
    skipCount?: number
  }
  sortModel?: {
    orderByCreatedTime?: number
  }
}

export function getTuitionAccountFlowRecordListApi(data: TuitionAccountFlowRecordListQueryParams) {
  return usePost<TuitionAccountFlowRecordListResult>('/api/v1/tuition-account-flows/list', data)
}

export function getSubTuitionAccountFlowRecordListApi(data: TuitionAccountFlowRecordListQueryParams) {
  return usePost<SubTuitionAccountFlowRecordListResult>('/api/v1/tuition-account-flows/sub-list', data)
}

export async function exportTuitionAccountFlowRecordListApi(data: {
  queryModel?: TuitionAccountFlowRecordListQueryParams['queryModel']
  sortModel?: TuitionAccountFlowRecordListQueryParams['sortModel']
}) {
  const token = useAuthorization()
  return axios.post('/api/v1/tuition-account-flows/export', data, {
    responseType: 'blob',
    headers: {
      [STORAGE_AUTHORIZE_KEY]: token.value || '',
      Authorization: token.value ? `Bearer ${token.value}` : '',
      'Accept-Language': 'zh-CN',
    },
  })
}

export async function exportSubTuitionAccountFlowRecordListApi(data: {
  queryModel?: TuitionAccountFlowRecordListQueryParams['queryModel']
  sortModel?: TuitionAccountFlowRecordListQueryParams['sortModel']
}) {
  const token = useAuthorization()
  return axios.post('/api/v1/tuition-account-flows/sub-export', data, {
    responseType: 'blob',
    headers: {
      [STORAGE_AUTHORIZE_KEY]: token.value || '',
      Authorization: token.value ? `Bearer ${token.value}` : '',
      'Accept-Language': 'zh-CN',
    },
  })
}
