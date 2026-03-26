import { usePost } from '~/utils/request'

export interface LedgerRichText {
  text?: string
  images?: string[]
}

export interface LedgerItem {
  id: string
  type: number
  sourceType: number
  ledgerCategoryId: string
  ledgerCategoryName: string
  ledgerSubCategoryId: string
  ledgerSubCategoryName: string
  ledgerCategoryIcon: string
  amount: number
  dealStaffId?: string
  dealStaffName?: string
  payTime?: string
  createdTime?: string
  payMethod?: number
  accountId?: string
  accountName?: string
  reciprocalAccount?: string
  bankSlipNo?: string
  orderNumber?: string
  ledgerNumber: string
  studentId?: string
  studentName?: string
  studentPhone?: string
  isConfirmed?: boolean
  confirmRemark?: LedgerRichText
  productItems?: string[]
  confirmStaffName?: string
  confirmTime?: string
  systemType?: number
  orderId?: string
  /** 关联订单类型：2 储值充值、4 储值退费（账单列表用于展示办理内容等） */
  orderType?: number
  /**
   * 储值账单办理内容用：储值账户名称（勿与 studentName、收款 accountName 混淆）。
   * 若列表接口未返回本字段，前端只显示「-」，需后端在 /api/v1/ledgers/list 结果中补齐。
   */
  rechargeAccountName?: string
  /** 部分后端可能使用的同义字段 */
  storedValueAccountName?: string
  paymentVoucher?: LedgerRichText
  billFlowId?: string
  billId?: string
  ledgerConfirmStatus: number
  errorMessage?: string
}

export interface LedgerListResult {
  list?: LedgerItem[]
  total?: number
}

export interface LedgerStatistics {
  incomeAmount?: number
  expenditureAmount?: number
  balanceAmount?: number
  totalConfirm?: number
  totalUnConfirm?: number
  totalRefunding?: number
  totalRefundFailed?: number
}

export interface LedgerQueryParams {
  sortModel?: Record<string, any>
  queryModel?: {
    accountIds?: string[]
    ledgerConfirmStatuses?: number[]
    sourceTypes?: number[]
    dealStaffId?: string
    confirmStaffId?: string
    studentId?: string
    orderNumber?: string
    bankSlipNo?: string
    ledgerNumber?: string
    confirmStartTime?: string
    confirmEndTime?: string
    payStartTime?: string
    payEndTime?: string
    ledgerSubCategoryIds?: string[]
    orderId?: string
  }
  pageRequestModel: {
    needTotal?: boolean
    pageSize: number
    pageIndex: number
    skipCount?: number
  }
}

export function getLedgerListApi(data: LedgerQueryParams) {
  return usePost<LedgerListResult>('/api/v1/ledgers/list', data)
}

export function getLedgerStatisticsApi(data: LedgerQueryParams) {
  return usePost<LedgerStatistics>('/api/v1/ledgers/statistics', data)
}

export function confirmLedgerApi(data: { id: string, confirmRemark?: LedgerRichText }) {
  return usePost<void>('/api/v1/ledgers/confirm', data)
}

export function cancelConfirmLedgerApi(data: { id: string }) {
  return usePost<void>('/api/v1/ledgers/cancel-confirm', data)
}
