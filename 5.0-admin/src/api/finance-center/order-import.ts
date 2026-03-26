import { useGet } from '~/utils/request'
import { usePost } from '~/utils/request'

export function buildLessonHourOrderImportTemplateApi(data: Record<string, unknown> = {}) {
  return useGet<string>('/api/v1/orders/import-template/lesson-hour', data)
}

export function buildTimeSlotOrderImportTemplateApi(data: Record<string, unknown> = {}) {
  return useGet<string>('/api/v1/orders/import-template/time-slot', data)
}

export function buildAmountOrderImportTemplateApi(data: Record<string, unknown> = {}) {
  return useGet<string>('/api/v1/orders/import-template/amount', data)
}

export interface OrderImportColumn {
  key: string
  title: string
  required: boolean
  fieldType: number
  fieldId?: number
  options?: string[]
}

export interface OrderImportCell {
  key: string
  title: string
  value: string
  selectedId?: any
  error?: string
}

export interface OrderImportRow {
  id: string
  rowNo: number
  hasError: boolean
  cells: OrderImportCell[]
  status?: number
  result?: string
}

export interface OrderImportTaskDetail {
  id: string
  fileName: string
  uploadStaffId: string
  uploadStaffName: string
  executeStaffId?: string
  executeStaffName?: string
  totalRows: number
  executedRows: number
  deletedRows: number
  errorRows: number
  createdTime?: string
  confirmTime?: string
  completeTime?: string
  status: number
  instName: string
}

export interface OrderImportTaskRecordListResult {
  list: OrderImportRow[]
  total: number
  columns: OrderImportColumn[]
}

export interface OrderImportUploadResult {
  fileUrl: string
  fileName: string
}

export function uploadOrderImportApi(data: FormData) {
  return usePost<OrderImportUploadResult, FormData>('/api/v1/orders/import-upload', data, {
    headers: {
      'Content-Type': 'multipart/form-data;charset=UTF-8',
    },
  })
}

export function submitOrderImportTaskApi(data: { fileUrl: string, fileName: string }) {
  return usePost<string>('/api/v1/orders/import-tasks/submit', data)
}

export function getOrderImportTaskDetailApi(params: { taskId: string }) {
  return useGet<OrderImportTaskDetail>('/api/v1/orders/import-tasks/detail', params)
}

export function getOrderImportTaskRecordListApi(data: {
  queryModel: { taskId: string, type: number }
  sortModel?: string
  pageRequestModel?: { needTotal?: boolean, pageSize?: number, pageIndex?: number, skipCount?: number }
}) {
  return usePost<OrderImportTaskRecordListResult>('/api/v1/orders/import-tasks/records', data)
}

export function batchSaveOrderImportTaskRecordsApi(data: { taskId: string, records: OrderImportRow[] }) {
  return usePost<OrderImportRow[]>('/api/v1/orders/import-tasks/batch-save-records', data)
}

export function deleteOrderImportTaskApi(data: { taskId: string }) {
  return usePost<boolean>('/api/v1/orders/import-tasks/delete', data)
}

export interface OrderImportStartResult {
  successCount: number
  failCount: number
}

export function startOrderImportTaskApi(data: { taskId: string }) {
  return usePost<OrderImportStartResult>('/api/v1/orders/import-tasks/start', data)
}

export interface OrderImportTaskListResult {
  list: OrderImportTaskDetail[]
  total: number
}

export function getOrderImportTaskListApi() {
  return useGet<OrderImportTaskListResult>('/api/v1/orders/import-tasks/list')
}

export function clearOrderImportTaskListApi() {
  return usePost<boolean>('/api/v1/orders/import-tasks/clear')
}

export function getOrderImportCourseOptionsApi(data: {
  pageRequestModel: {
    needTotal?: boolean
    pageSize: number
    pageIndex: number
    skipCount?: number
  }
  queryModel?: {
    searchKey?: string
  }
  sortModel?: Record<string, any>
}) {
  return usePost<{ id: number, name: string }[]>('/api/v1/courses/options', data)
}
