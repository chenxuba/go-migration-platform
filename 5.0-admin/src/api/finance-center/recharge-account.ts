import axios from 'axios'
import { STORAGE_AUTHORIZE_KEY, useAuthorization } from '~/composables/authorization'
import { useGet, usePost } from '~/utils/request'

export interface RechargeAccountStudentItem {
  isMainStudent?: boolean
  studentId?: string
  studentName?: string
}

export interface RechargeAccountItem {
  rechargeAccountId: string
  rechargeAccountName?: string
  phone?: string
  mainStudentId?: string
  updateTime?: string
  balanceTotal?: number
  rechargeBalance?: number
  residualBalance?: number
  givingBalance?: number
  rechargeAccountStudents?: RechargeAccountStudentItem[]
}

export interface RechargeAccountItemPageResult {
  list?: RechargeAccountItem[]
  total?: number
}

export interface RechargeAccountStatistics {
  rechargeAccountTotal?: number
  amountTotal?: number
  givingAmountTotal?: number
  residualAmountTotal?: number
}

export interface RechargeAccountDetailStudentItem {
  isMainStudent?: boolean
  studentId?: string
  studentName?: string
}

export interface RechargeAccountDetailItem {
  rechargeAccountId: string
  rechargeAccountFlowId?: string
  rechargeAccountName?: string
  phone?: string
  amount?: number
  givingAmount?: number
  residualAmount?: number
  remark?: string
  createTime?: string
  rechargeAccountFlowSourceType?: number
  dealDate?: string
  sourceId?: string
  sourceOrderNumber?: string
  sourceOrderType?: number
  rechargeAccountStudents?: RechargeAccountDetailStudentItem[]
  studentId?: string
  studentName?: string
  studentPhone?: string
  studentAvatar?: string
  totalAmount?: number
}

export interface RechargeAccountDetailPageResult {
  list?: RechargeAccountDetailItem[]
  total?: number
}

export interface RechargeAccountExpendIncome {
  expend?: number
  income?: number
}

export interface RechargeAccountItemPageQueryParams {
  queryModel?: {
    studentId?: string
    showZeroBalanceAccount?: boolean
  }
  pageRequestModel: {
    needTotal?: boolean
    pageSize: number
    pageIndex: number
    skipCount?: number
  }
  sortModel?: {
    orderByUpdatedTime?: number
  }
}

export interface RechargeAccountDetailPageQueryParams {
  queryModel?: {
    studentId?: string
    rechargeAccountId?: string
    startTime?: string
    endTime?: string
    flowTypes?: number[]
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

export interface StudentDetailView {
  id: string
  name: string
  phone: string
  avatar?: string
  sex?: number
  phoneRelationship?: number
  salespersonId?: string
  salespersonName?: string
  createdTime?: string
  firstEnrolledTime?: string
  turnedHistoryTime?: string
  createdStaffId?: string
  createdStaffName?: string
  collectorStaffId?: string
  collectorStaffName?: string
  phoneSellStaffId?: string
  phoneSellStaffName?: string
  foregroundStaffId?: string
  foregroundStaffName?: string
  viceSellStaffStaffId?: string
  viceSellStaffStaffName?: string
  status?: number
}

export interface RechargeAccountByStudentStudent {
  id: string
  name: string
  avatar?: string
  sex?: number
  phone?: string
  isMainStudent?: boolean
}

export interface RechargeAccountByStudent {
  id: string
  accountName?: string
  phone?: string
  mainStudentId?: string
  balance?: number
  givingBalance?: number
  residualBalance?: number
  createdAt?: string
  students?: RechargeAccountByStudentStudent[]
}

export interface RechargeAccountOrderCreateParams {
  rechargeAccountId: string
  amount: number
  givingAmount: number
  residualAmount?: number
  dealDate: string
  salePersonId: string
  collectorStaffId: string
  phoneSellStaffId: string
  foregroundStaffId: string
  viceSellStaffStaffId: string
  remark?: string
  orderTagIds?: string[]
  externalRemark?: string
  /** 退费单可不传，后端按账户主学员解析 */
  studentId?: string
}

export interface RechargeAccountOrderCreateResult {
  id: string
  name: string
}

export interface RechargeAccountOrderDetail {
  id: string
  rechargeAccountId: string
  saleOrderId?: string
  orderNumber: string
  status: number
  amount: number
  givingAmount: number
  residualAmount?: number
  operatorName?: string
  createdAt?: string
  bill?: {
    id: string
    status: number
    billFlows?: any[]
  }
  approveId?: string | null
  orderTags?: Array<{
    tagId: string
    tagName: string
  }>
  studentId?: string
  studentName?: string
  studentPhone?: string
  orderObsolete?: any
}

export interface UpdateRechargeAccountParams {
  rechargeAccountId: string
  rechargeAccountName: string
}

export function getRechargeAccountItemPageApi(data: RechargeAccountItemPageQueryParams) {
  return usePost<RechargeAccountItemPageResult>('/api/v1/recharge-accounts/page', data)
}

export function getRechargeAccountStatisticsApi() {
  return useGet<RechargeAccountStatistics>('/api/v1/recharge-accounts/statistics')
}

export async function exportRechargeAccountItemPageApi(data: {
  queryModel?: RechargeAccountItemPageQueryParams['queryModel']
  sortModel?: RechargeAccountItemPageQueryParams['sortModel']
}) {
  const token = useAuthorization()
  return axios.post('/api/v1/recharge-accounts/export', data, {
    responseType: 'blob',
    headers: {
      [STORAGE_AUTHORIZE_KEY]: token.value || '',
      Authorization: token.value ? `Bearer ${token.value}` : '',
      'Accept-Language': 'zh-CN',
    },
  })
}

export function buildRechargeAccountImportByStudentTemplateApi() {
  return useGet<string>('/api/v1/recharge-accounts/import-template/by-student')
}

export function buildRechargeAccountImportByAccountTemplateApi() {
  return useGet<string>('/api/v1/recharge-accounts/import-template/by-account')
}

export interface RechargeAccountImportColumn {
  key: string
  title: string
  required: boolean
  fieldType: number
  fieldId?: number
  options?: string[]
}

export interface RechargeAccountImportCell {
  key: string
  title: string
  value: string
  selectedId?: any
  error?: string
}

export interface RechargeAccountImportRow {
  id: string
  rowNo: number
  hasError: boolean
  cells: RechargeAccountImportCell[]
  status?: number
  result?: string
}

export interface RechargeAccountImportTaskDetail {
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

export interface RechargeAccountImportTaskRecordListResult {
  list: RechargeAccountImportRow[]
  total: number
  columns: RechargeAccountImportColumn[]
}

export interface RechargeAccountImportUploadResult {
  fileUrl: string
  fileName: string
}

export interface RechargeAccountImportTaskListResult {
  list: RechargeAccountImportTaskDetail[]
  total: number
}

export interface RechargeAccountImportStartResult {
  successCount: number
  failCount: number
}

export function uploadRechargeAccountImportApi(data: FormData) {
  return usePost<RechargeAccountImportUploadResult, FormData>('/api/v1/recharge-accounts/import-upload', data, {
    headers: {
      'Content-Type': 'multipart/form-data;charset=UTF-8',
    },
  })
}

export function submitRechargeAccountImportTaskApi(data: { fileUrl: string, fileName: string }) {
  return usePost<string>('/api/v1/recharge-accounts/import-tasks/submit', data)
}

export function getRechargeAccountImportTaskDetailApi(params: { taskId: string }) {
  return useGet<RechargeAccountImportTaskDetail>('/api/v1/recharge-accounts/import-tasks/detail', params)
}

export function getRechargeAccountImportTaskRecordListApi(data: {
  queryModel: { taskId: string, type: number }
  sortModel?: string
  pageRequestModel?: { needTotal?: boolean, pageSize?: number, pageIndex?: number, skipCount?: number }
}) {
  return usePost<RechargeAccountImportTaskRecordListResult>('/api/v1/recharge-accounts/import-tasks/records', data)
}

export function batchSaveRechargeAccountImportTaskRecordsApi(data: { taskId: string, records: RechargeAccountImportRow[] }) {
  return usePost<RechargeAccountImportRow[]>('/api/v1/recharge-accounts/import-tasks/batch-save-records', data)
}

export function startRechargeAccountImportTaskApi(data: { taskId: string }) {
  return usePost<RechargeAccountImportStartResult>('/api/v1/recharge-accounts/import-tasks/start', data)
}

export function getRechargeAccountImportTaskListApi() {
  return useGet<RechargeAccountImportTaskListResult>('/api/v1/recharge-accounts/import-tasks/list')
}

export function clearRechargeAccountImportTaskListApi() {
  return usePost<boolean>('/api/v1/recharge-accounts/import-tasks/clear')
}

export function deleteRechargeAccountImportTaskApi(data: { taskId: string }) {
  return usePost<boolean>('/api/v1/recharge-accounts/import-tasks/delete', data)
}

export function getRechargeAccountDetailPageApi(data: RechargeAccountDetailPageQueryParams) {
  return usePost<RechargeAccountDetailPageResult>('/api/v1/recharge-accounts/details/page', data)
}

export function getRechargeAccountExpendIncomeApi(params: {
  studentId?: string
  startTime?: string
  endTime?: string
}) {
  return useGet<RechargeAccountExpendIncome>('/api/v1/recharge-accounts/expend-income', params)
}

export function getStudentDetailApi(params: { studentId: string | number }) {
  return useGet<StudentDetailView>('/api/v1/students/detail', params)
}

export function getRechargeAccountByStudentApi(data: { studentId: string | number }) {
  return usePost<RechargeAccountByStudent>('/api/v1/recharge-accounts/by-student', data)
}

/** 按储值账户 ID 拉取详情（退费抽屉，对标 GetRechargeAccount） */
export function getRechargeAccountApi(data: { rechargeAccountId: string | number }) {
  return usePost<RechargeAccountByStudent>('/api/v1/recharge-accounts/get', data)
}

export function updateRechargeAccountApi(data: UpdateRechargeAccountParams) {
  return usePost<boolean>('/api/v1/recharge-accounts/update', data)
}

export function createRechargeAccountOrderApi(data: RechargeAccountOrderCreateParams) {
  return usePost<RechargeAccountOrderCreateResult>('/api/v1/recharge-account-orders/create', data)
}

/** 创建储值账户退费订单（待支付/待平账，后续 PayBill 同充值） */
export function createRechargeAccountRefundOrderApi(data: RechargeAccountOrderCreateParams) {
  return usePost<RechargeAccountOrderCreateResult>('/api/v1/recharge-account-orders/create-refund', data)
}

export function getRechargeAccountOrderDetailApi(data: { rechargeAccountOrderId?: string, saleOrderId?: string }) {
  return usePost<RechargeAccountOrderDetail>('/api/v1/recharge-account-orders/detail', data)
}

export function payOrderBySchoolPalApi(data: {
  billId: string
  amount: number
  remark?: string
  payMethod?: number
  amountId?: number
  payTime?: string
  paymentVoucher?: string
}) {
  return usePost<string>('/api/v1/orders/pay-by-schoolpal', data)
}
