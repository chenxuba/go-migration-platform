export interface TuitionAccountReadingItem {
  id?: string
  lessonId?: string
  lessonName?: string
  lessonType?: number
  totalQuantity?: number
  totalFreeQuantity?: number
  totalTuition?: number
  arrearTuition?: number
  isAdjustable?: boolean
  remainQuantity?: number
  tuition?: number
  remainFreeQuantity?: number
  enableExpireTime?: boolean
  expireTime?: string
  validDate?: string
  endDate?: string
  activedAt?: string
  assignedClass?: boolean
  status?: number
  changeStatusTime?: string
  lessonChargingMode?: number
  planSuspendTime?: string
  planResumeTime?: string
  hasGradeUpgrade?: boolean
  manualSort?: boolean
}

export interface TuitionAccountReadingListQueryParams {
  sortModel?: Record<string, any>
  queryModel: {
    studentId: string
  }
  pageRequestModel: {
    needTotal?: boolean
    pageSize: number
    pageIndex: number
    skipCount?: number
  }
}

export interface TuitionAccountReadingListResult {
  list?: TuitionAccountReadingItem[]
  total?: number
}

export interface TuitionAccountSubAccountDateInfoItem {
  id?: string
  createdTime?: string
  activedAt?: string
  remainDays?: number
  rawStatus?: number
  status?: number
  isFree?: boolean
  totalDays?: number
  tuition?: number
  totalTuition?: number
  endDate?: string
  sourceType?: number
  orderId?: string
  unitPrice?: number
  paidTuition?: number
  shouldTuition?: number
  arrearTuition?: number
  chargeAgainstTuition?: number
  transferredTuition?: number
  paidRemaining?: number
  usedTuition?: number
  startDate?: string
}

export interface TuitionAccountSubAccountDateInfoResult {
  list?: TuitionAccountSubAccountDateInfoItem[]
}

export interface RevertCloseTuitionAccountPreviewSubPeriod {
  startDate?: string
  endDate?: string
}

export interface RevertCloseTuitionAccountPreview {
  tuitionAccountId?: string
  lessonName?: string
  lessonType?: number
  lessonChargingMode?: number
  closeTuitionAccountOrderId?: string
  closeTime?: string
  quantity?: number
  freeQuantity?: number
  tuition?: number
  remark?: string
  expireDate?: string
  arrearAmountTotal?: number
  badDebtAmountTotal?: number
  orderId?: string
  orderType?: number
  subTuitionAccounts?: RevertCloseTuitionAccountPreviewSubPeriod[]
}

export interface RevertCloseTuitionAccountParams {
  tuitionAccountId: string
  closeTuitionAccountOrderId: string
  startDate?: string
  expireDate?: string
  currentValidStartDate?: string
}

export interface RevertCloseTuitionAccountResult {
  id?: string
}

export interface CloseTuitionAccountOrderParams {
  tuitionAccountId: string
  quantity: number
  freeQuantity: number
  tuition: number
  remark?: string
}

export interface CloseTuitionAccountOrderResult {
  id?: string
  name?: string
}

export interface CloseTuitionAccountOrderRecordItem {
  id?: string
  tuitionAccountId?: string
  quantity?: number
  freeQuantity?: number
  status?: number
  updatedStaffId?: string
  updatedStaffName?: string
  updatedTime?: string
  createdTime?: string
}

export interface CloseTuitionAccountOrderRecordResult {
  list?: CloseTuitionAccountOrderRecordItem[]
}

// 查询学生报读列表（学费账户在读列表）
export function getTuitionAccountReadingListApi(data: TuitionAccountReadingListQueryParams) {
  return usePost<TuitionAccountReadingListResult>('/api/v1/tuition-accounts/reading-list', data)
}

export function getTuitionAccountSubAccountDateInfoApi(data: { tuitionAccountId: string }) {
  return usePost<TuitionAccountSubAccountDateInfoResult>('/api/v1/tuition-accounts/sub-account-date-info', data)
}

export function getRevertCloseTuitionAccountPreviewApi(data: { tuitionAccountId: string }) {
  return usePost<RevertCloseTuitionAccountPreview>('/api/v1/tuition-accounts/revert-close-preview', data)
}

export function revertCloseTuitionAccountApi(data: RevertCloseTuitionAccountParams) {
  return usePost<RevertCloseTuitionAccountResult>('/api/v1/tuition-accounts/revert-close', data)
}

export function addCloseTuitionAccountOrderApi(data: CloseTuitionAccountOrderParams) {
  return usePost<CloseTuitionAccountOrderResult>('/api/v1/tuition-accounts/close-order', data)
}

export function getCloseTuitionAccountOrderListApi(data: { tuitionAccountId: string }) {
  return usePost<CloseTuitionAccountOrderRecordResult>('/api/v1/tuition-accounts/close-orders/list', data)
}
