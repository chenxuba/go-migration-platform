export interface TuitionAccountReadingItem {
  id?: string
  lessonId?: string
  lessonName?: string
  lessonType?: number
  totalQuantity?: number
  totalFreeQuantity?: number
  totalTuition?: number
  arrearTuition?: number
  lessonConsumeArrearQuantity?: number
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
  quantity?: number
  createdTime?: string
  startTime?: string
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
  accountSourceType?: number
  orderId?: string
  sourceId?: string
  unitPrice?: number
  paidTuition?: number
  shouldTuition?: number
  arrearTuition?: number
  chargeAgainstTuition?: number
  transferredTuition?: number
  paidRemaining?: number
  usedTuition?: number
  startDate?: string
  expiredToClearQuantity?: boolean
  expireDate?: string
}

export interface TuitionAccountSubAccountDateInfoResult {
  list?: TuitionAccountSubAccountDateInfoItem[]
}

export interface SubTuitionAccountPriorityConfigItem {
  priorityType?: number
  sortDirection?: number
  sortWeight?: number
  isEnabled?: boolean
}

export interface SubTuitionAccountPriorityConfigResult {
  list?: SubTuitionAccountPriorityConfigItem[]
}

export interface RevertCloseTuitionAccountPreviewSubPeriod {
  quantity?: number
  isFree?: boolean
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

export interface SuspendResumeTuitionAccountOrderParams {
  tuitionAccountId: string
  type: number
  expireTime?: string
  expireType?: number
  remark?: string
  suspendDate?: string
  resumeDate?: string
}

export interface SuspendResumeTuitionAccountOrderResult {
  id?: string
  studentId?: string
  lessonId?: string
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

export function getSubTuitionAccountPriorityConfigListApi() {
  return usePost<SubTuitionAccountPriorityConfigResult>('/api/v1/tuition-accounts/sub-account-priority-configs/list', {})
}

export function getRevertCloseTuitionAccountPreviewApi(data: { tuitionAccountId: string }) {
  return usePost<RevertCloseTuitionAccountPreview>('/api/v1/tuition-accounts/revert-close-preview', data)
}

export function revertCloseTuitionAccountApi(data: RevertCloseTuitionAccountParams) {
  return usePost<RevertCloseTuitionAccountResult>('/api/v1/tuition-accounts/revert-close', data)
}

export function addSuspendResumeTuitionAccountOrderApi(data: SuspendResumeTuitionAccountOrderParams) {
  return usePost<SuspendResumeTuitionAccountOrderResult>('/api/v1/tuition-accounts/suspend-resume-orders/create', data)
}

export function addCloseTuitionAccountOrderApi(data: CloseTuitionAccountOrderParams) {
  return usePost<CloseTuitionAccountOrderResult>('/api/v1/tuition-accounts/close-order', data)
}

export function getCloseTuitionAccountOrderListApi(data: { tuitionAccountId: string }) {
  return usePost<CloseTuitionAccountOrderRecordResult>('/api/v1/tuition-accounts/close-orders/list', data)
}

/** 对标 TuitionAccount/GetTuitionAccountListByLessonId（集体班添加学员） */
export interface TuitionAccountByLessonRow {
  studentId: string
  tuitionAccountId: string
  studentName: string
  assignedClass: boolean
  quantity: number
  avatar?: string | null
  phone: string
  lessonChargingMode: number
  lessonScope?: number
  isTuitionAccountActive?: boolean
  sex?: number
  birthday?: string
  productId?: string
  productName?: string
}

export interface PageTuitionAccountsByLessonIdBody {
  pageRequestModel: {
    needTotal?: boolean
    pageSize: number
    pageIndex: number
    skipCount?: number
  }
  queryModel: {
    lessonId: string
    studentIds: string[]
    /** 当前集体班 id，用于本班已入班标记与勾选禁用 */
    classId?: string
    /** 服务端筛选：1 男 2 女 0 未知，与竞品 GetTuitionAccountListByLessonId 一致 */
    sex?: number[]
    ageMin?: number
    ageMax?: number
    /** 学员姓名模糊 */
    studentName?: string
  }
}

export interface PageTuitionAccountsByLessonIdResult {
  list: TuitionAccountByLessonRow[]
  total: number
}

export function pageTuitionAccountsByLessonIdApi(data: PageTuitionAccountsByLessonIdBody) {
  return usePost<PageTuitionAccountsByLessonIdResult>('/api/v1/tuition-accounts/page-by-lesson-id', data)
}
