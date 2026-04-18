import axios from 'axios'
import { STORAGE_AUTHORIZE_KEY, useAuthorization } from '~/composables/authorization'
import { usePost } from '~/utils/request'

// 订单项信息
export interface OrderItem {
  orderId: string
  sourceId: string
  orderNumber: string
  studentId: string
  studentName: string
  sex: number
  studentPhone: string
  createdTime: string
  totalAmount: number
  amount: number
  paidAmount: number
  orderStatus: number
  orderType: number
  tranOrderType: number
  remark: string
  externalRemark: string
  staffId: string
  staffName: string
  finishedTime: string
  billFinishedTime: string
  productItemsStr: string
  productItems: string[]
  isAmountOwed: boolean
  dealDate: string
  avatar: string
  salePersonId: string
  salePersonName: string
  collectorStaffId: string
  collectorStaffName: string
  phoneSellStaffId: string
  phoneSellStaffName: string
  foregroundStaffId: string
  foregroundStaffName: string
  viceSellStaffStaffId: string
  viceSellStaffStaffName: string
  arrearAmount: number
  tagNames: string[]
  rechargeAccountId: string
  rechargeAccountAmount: number
  rechargeAccountResidualAmount?: number
  rechargeAccountGivingAmount: number
  wideDiscountId: string
  wideDiscountName: string
  shareAmount: number
  customerRemark: string
  totalActualPaidAmount: number
  totalChargeAgainstAmount: number
  latestPaidTime: string
  isBadDebt: boolean
  badDebtAmount: number
  badDebtRemark: string
  orderSource: number
}

// 查询参数
export interface OrderQueryParams {
  sortModel?: Record<string, any>
  queryModel?: {
    keyword?: string
    keywordType?: string
    orderIds?: string[]
    orderStatus?: number
    orderStatusList?: number[]
    orderType?: number
    orderTypeList?: number[]
    orderTagIds?: string[]
    orderSourceList?: number[]
    studentId?: string
    staffId?: string
    creatorId?: string
    salePersonId?: string
    courseIds?: string[]
    billingModes?: number[]
    isArrears?: boolean
    orderArrearStatus?: number[]
    createdTimeBegin?: string
    createdTimeEnd?: string
    dealDateBegin?: string
    dealDateEnd?: string
    latestPaidTimeBegin?: string
    latestPaidTimeEnd?: string
  }
  pageRequestModel: {
    needTotal?: boolean
    pageSize: number
    pageIndex: number
    skipCount?: number
  }
}

// 返回结果
export interface OrderListResult {
  list?: OrderItem[]
  total?: number
  totalPaid?: number
  totalArrear?: number
  totalBadDebt?: number
}

export interface OrderApprovalInfo {
  approvalId?: string
  approvalNumber?: string
  approvalStatus?: number
  currentStep?: number
  currentApprover?: string
  applicantId?: string
  applicantName?: string
  approvalTime?: string
  finishTime?: string
}

export interface OrderDetailLineItem {
  orderCourseDetailId: string
  courseId?: string
  courseName?: string
  quoteId?: string
  quoteName?: string
  quotePrice?: number
  lessonType?: number
  chargingMode?: number
  handleType?: number
  count?: number
  unit?: number
  quoteQuantity?: number
  freeQuantity?: number
  hasValidDate?: boolean
  validDate?: string
  endDate?: string
  discountType?: number
  discountNumber?: number
  singleDiscountAmount?: number
  shareDiscount?: number
  amount?: number
  receivableAmount?: number
  realQuantity?: number
}

export interface OrderPaymentRecord {
  paymentId: string
  amountId?: string
  accountName?: string
  payMethod?: number
  payAmount?: number
  payTime?: string
  createdTime?: string
  paymentVoucher?: string
  remark?: string
  operatorId?: string
  operatorName?: string
}

export interface OrderTagDetailItem {
  tagId: string
  tagName: string
}

export interface OrderDetailListItem {
  orderId: string
  sourceId: string
  orderNumber: string
  studentId: string
  studentName: string
  studentPhone: string
  studentAvatar?: string
  sex?: number
  createdTime?: string
  orderStatus?: number
  orderType?: number
  tranOrderType?: number
  staffId?: string
  staffName?: string
  isAmountOwed?: boolean
  dealDate?: string
  productId?: string
  productName?: string
  quoteName?: string
  enrollType?: number
  orderFlowProductId?: string
  skuId?: string
  skuName?: string
  skuCount?: number
  skuUnit?: number
  freeQuantity?: number
  discountType?: number
  discountNumber?: number
  shareDiscount?: number
  shareCouponAmount?: number
  tuition?: number
  quantity?: number
  realQuantity?: number
  validDate?: string
  endDate?: string
  productType?: number
  remark?: string
  chargingMode?: number
  realTuition?: number
  salePersonId?: string
  salePersonName?: string
  productCategoryId?: string
  productCategoryName?: string
  totalQuantity?: number
  tagNames?: string[]
  externalRemark?: string
  classId?: string
  className?: string
  classAssignStatus?: number
  customerRemark?: string
  actualPaidAmount?: number
  chargeAgainstAmount?: number
  shareRechargeAccountAmount?: number
  shareRechargeAccountGivingAmount?: number
  shouldAmount?: number
  isBadDebt?: boolean
  badDebtAmount?: number
  arrearAmount?: number
  productPackageId?: string
  productPackageName?: string
}

export interface OrderDetailListQueryParams {
  sortModel?: Record<string, any>
  queryModel?: {
    orderNumber?: string
    orderTypeList?: number[]
    orderTagIds?: string[]
    orderSourceList?: number[]
    orderStatusList?: number[]
    courseIds?: string[]
    enrollTypes?: number[]
    productTypes?: number[]
    courseCategoryId?: string | number
    salePersonId?: string
    creatorId?: string
    dealDateBegin?: string
    dealDateEnd?: string
    createdTimeBegin?: string
    createdTimeEnd?: string
    latestPaidTimeBegin?: string
    latestPaidTimeEnd?: string
    orderArrearStatus?: number[]
    studentId?: string
  }
  pageRequestModel: {
    needTotal?: boolean
    pageSize: number
    pageIndex: number
    skipCount?: number
  }
}

export interface OrderDetailListResult {
  list?: OrderDetailListItem[]
  total?: number
}

export interface OrderDetailResult extends OrderItem {
  totalAmount: number
  approveId?: string
  orderDiscountAmount?: number
  orderTagNames?: string[]
  orderTags?: OrderTagDetailItem[]
  approvalInfo?: OrderApprovalInfo
  orderItems?: OrderDetailLineItem[]
  paymentRecords?: OrderPaymentRecord[]
}

// 获取订单列表
export function getOrderListApi(data: OrderQueryParams) {
  return usePost<OrderListResult>('/api/v1/orders/list', data)
}

// 获取订单详情
export function getOrderDetailApi(data: { orderId: string }) {
  return usePost<OrderDetailResult>('/api/v1/orders/detail', data)
}

export function getOrderDetailPagedApi(data: OrderDetailListQueryParams) {
  return usePost<OrderDetailListResult>('/api/v1/orders/detail-paged', data)
}

export async function exportOrderDetailPagedApi(data: {
  queryModel?: OrderDetailListQueryParams['queryModel']
  sortModel?: OrderDetailListQueryParams['sortModel']
}) {
  const token = useAuthorization()
  return axios.post('/api/v1/orders/detail-paged/export', data, {
    responseType: 'blob',
    headers: {
      [STORAGE_AUTHORIZE_KEY]: token.value || '',
      Authorization: token.value ? `Bearer ${token.value}` : '',
      'Accept-Language': 'zh-CN',
    },
  })
}

// 设为坏账
export function setBadDebtApi(data: { orderId: string; remark?: string }) {
  return usePost<void>('/api/v1/orders/set-bad-debt', data)
}

// 取消坏账
export function cancelBadDebtApi(data: { orderId: string }) {
  return usePost<void>('/api/v1/orders/cancel-bad-debt', data)
}

export function closeOrderApi(data: { orderId: string }) {
  return usePost<void>('/api/v1/orders/close', data)
}

export async function exportOrderListApi(data: {
  queryModel?: OrderQueryParams['queryModel']
  sortModel?: OrderQueryParams['sortModel']
}) {
  const token = useAuthorization()
  return axios.post('/api/v1/orders/export', data, {
    responseType: 'blob',
    headers: {
      [STORAGE_AUTHORIZE_KEY]: token.value || '',
      Authorization: token.value ? `Bearer ${token.value}` : '',
      'Accept-Language': 'zh-CN',
    },
  })
}

export async function downloadOrderReceiptPdfApi(data: { orderId: string | number, template?: 'a4' | 'dot' | 'receipt' }) {
  const token = useAuthorization()
  return axios.get('/api/v1/orders/receipt-pdf/download', {
    params: data,
    responseType: 'blob',
    headers: {
      [STORAGE_AUTHORIZE_KEY]: token.value || '',
      Authorization: token.value ? `Bearer ${token.value}` : '',
      'Accept-Language': 'zh-CN',
    },
  })
}
