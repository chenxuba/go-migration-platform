export interface TuitionAccountReadingItem {
  id?: string
  lessonId?: string
  lessonName?: string
  lessonType?: number
  totalQuantity?: number
  totalFreeQuantity?: number
  totalTuition?: number
  arrearTuition?: number
  badDebtTuition?: number
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
  orderStatus?: number
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

export interface RefundTuitionAccountOwedSummaryResult {
  arrearAmountTotal?: number
  badDebtAmountTotal?: number
  orderId?: string
  orderType?: number
}

export interface RefundTuitionAccountHandlingFeeDetail {
  orderNumber?: string
  originalUnitPrice?: number
  discountedUnitPrice?: number
  shouldTuition?: number
  paidAmount?: number
  transferredTuition?: number
  consumedQuantity?: number
  arrearTuition?: number
  originalRefundAmount?: number
  lessonChargingMode?: number
  deductionQuantity?: number
}

export interface RefundTuitionAccountValuableEstimateSubAccount {
  tuitionAccountId?: string
  orderNumber?: string
  quantity?: number
  freeQuantity?: number
  tuition?: number
}

export interface RefundTuitionAccountValuableEstimateResult {
  tuitionAccountId?: string
  quantity?: number
  freeQuantity?: number
  tuition?: number
  subAccounts?: RefundTuitionAccountValuableEstimateSubAccount[]
}

export interface RefundTuitionAccountHandlingFeeResult {
  tuitionAccountId?: string
  refundAmount?: number
  totalOriginalRefundAmount?: number
  totalArrearDeduction?: number
  handlingFee?: number
  lessonChargingMode?: number
  paidRefundQuantity?: number
  giftRefundQuantity?: number
  arrearAmountTotal?: number
  badDebtAmountTotal?: number
  orderId?: string
  orderType?: number
  details?: RefundTuitionAccountHandlingFeeDetail[]
}

export interface RefundTuitionAccountCreateOrderParams {
  tuitionAccountId: string
  totalAmount: number
  realAmount: number
  chargeAgainstTuition: number
  refundQuantity: number
  refundFreeQuantity: number
  isRechargeAccount: boolean
  rechargeAccountId: string
  dealDate: string
  remark?: string
  externalRemark?: string
  salePersonId?: string
  collectorStaffId?: string
  phoneSellStaffId?: string
  foregroundStaffId?: string
  viceSellStaffStaffId?: string
  orderTagIds?: string[]
  autoCloseTuition: boolean
  isOriginalRefund?: boolean
}

export interface RefundTuitionAccountCreateOrderResult {
  id?: string
  isNeedPay?: boolean
}

export interface RefundTuitionAccountPayOrderParams {
  orderId: string
  payAmount: number
  isOriginalRefund: boolean
  payAccounts: Array<{
    payMethod: number
    amount: number
    accountId: string
    paymentVoucher?: {
      text?: string
      images?: string[]
    }
    payTime: string
  }>
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

export function getTuitionAccountRefundOwedSummaryApi(data: { tuitionAccountId: string }) {
  return usePost<RefundTuitionAccountOwedSummaryResult>('/api/v1/tuition-accounts/refund-owed-summary', data)
}

export function estimateRefundTuitionAccountValuableTuitionApi(data: { tuitionAccountId: string, quantity: number }) {
  return usePost<RefundTuitionAccountValuableEstimateResult>('/api/v1/tuition-accounts/refund-estimate-valuable-tuition', data)
}

export function calculateRefundTuitionAccountHandlingFeeApi(data: { tuitionAccountId: string, refundQuantity: number }) {
  return usePost<RefundTuitionAccountHandlingFeeResult>('/api/v1/tuition-accounts/refund-calc-handling-fee', data)
}

export function createRefundTuitionAccountOrderApi(data: RefundTuitionAccountCreateOrderParams) {
  return usePost<RefundTuitionAccountCreateOrderResult>('/api/v1/tuition-accounts/refund-orders/create', data)
}

export function payRefundTuitionAccountOrderApi(data: RefundTuitionAccountPayOrderParams) {
  return usePost<string>('/api/v1/tuition-accounts/refund-orders/pay', data)
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
  assignedOtherClass?: boolean
  assignedOtherClassText?: string
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
