// 商品SKU接口
export interface ProductSku {
  id: number
  uuid: string
  version: number
  createTime: string
  updateTime: string | null
  courseId: number
  lessonModel: number
  name: string
  unit: number
  quantity: number | null
  price: number
  lessonAudition: boolean
  onlineSale: boolean
  remark: string | null
}

// 课程接口
export interface Course {
  id: number
  uuid: string
  version: number
  name: string
  courseCategory: string | null
  categoryName: string | null
  courseType: number
  commonCourseList: any | null
  teachMethod: number
  teachMethodName: string | null
  chargeMethods: string
  hasExperiencePrice: boolean
  saleStatus: number
  productSku: ProductSku[]
}

// 分页请求参数接口
export interface ProcessContentPageParams {
  pageRequestModel: {
    needTotal: boolean
    pageSize: number
    pageIndex: number
    skipCount: number
  }
  queryModel: {
    delFlag: boolean
    productType?: number // 商品类型：1-课程商品，2-学杂费，3-教材商品
    name?: string // 搜索关键词
    saleStatus?: boolean // 销售状态
    courseCategory?: string // 课程分类
    teachMethod?: number // 授课方式
    courseType?: number // 课程范围
    [key: string]: any
  }
  sortModel: Record<string, any>
}



// 通用响应数据结构
export interface ResponseData<T = any> {
  code: number
  message: string
  result: T
  total?: number
  [key: string]: any
}

// 报价单检测结果
export interface CheckQuoteInfoResult {
  courseId: number
  error: number // 1表示有错误，0表示无错误
}

// 报名类型计算结果
export interface CalcCourseEnrollTypeResult {
  enrollType: number // 1-新报，2-续费，3-扩科
}

export interface CalcCourseEnrollTypeParams {
  studentId: number
  courses: {
    courseId: number
    isAudition: boolean
  }[]
}

export interface CheckQuoteInfoParams {
  quoteDetailList: {
    courseId: number
    quoteId: number
    price: number
    quantity: number
    lessonModel: number
  }[]
}

// 报价单详情接口
export interface QuoteDetail {
  handleType: number
  courseId: number
  courseType: number
  quoteId: number
  lessonMode: number
  classId: number | null
  count: number
  unit: number
  freeQuantity: number
  discountType: number
  discountNumber: number
  hasValidDate: boolean
  validDate: string | null
  endDate: string | null
  shareDiscount: string
  amount: string
  quantity: number
  realQuantity: number
  realAmount: string
}

// 订单详情接口
export interface OrderDetail {
  quoteDetailList: QuoteDetail[]
  orderDiscountType: number
  orderDiscountNumber: number | undefined
  orderDiscountAmount: string
  orderRealQuantity: number
  orderRealAmount: string
  internalRemark: string
  externalRemark: string
  dealDate: string
  salePerson: number | null
  orderTagIds: number[]
}

// 创建订单参数接口
export interface CreateOrderParams {
  studentId: number
  orderDetail: OrderDetail
}

// 支付账户详情接口
export interface PayAccount {
  orderId: number
  amountId: number
  payMethod: number
  payAmount: number
  payTime: string
  paymentVoucher: string
}

// 订单支付参数接口
export interface PayOrderParams {
  orderId: number
  payAmount: number
  payAccounts: PayAccount[]
}

// 分页获取报名课程商品 /instCourse/getProcessContentPage
export function getProcessContentPageApi(data: ProcessContentPageParams) {
  return usePost<ResponseData<Course[]>>('/api/v1/courses/process-content', data)
}

// 查询报名类型 /saleOrder/calcCourseEnrollType
export function getCalcCourseEnrollTypeApi(data: CalcCourseEnrollTypeParams) {
  return usePost<ResponseData<CalcCourseEnrollTypeResult[]>>('/api/v1/orders/calc-enroll-type', data)
}
// 检测报价单信息/saleOrder/checkQuoteInfo
export function getCheckQuoteInfoApi(data: CheckQuoteInfoParams) {
  return usePost<ResponseData<CheckQuoteInfoResult[]>>('/api/v1/orders/check-quote', data)
}
// 创建订单 /saleOrder/createOrder
export function postCreateOrderApi(data: CreateOrderParams) {
  return usePost<ResponseData<{ orderId: number }>>('/api/v1/orders/create', data)
}
// 订单支付 /saleOrder/payOrder
export function postPayOrderApi(data: PayOrderParams) {
  return usePost<ResponseData>('/api/v1/orders/pay', data)
}
// 获取课程id和name 
export function getCourseIdAndNameApi(data: { searchKey: string }) {
  return usePost<ResponseData<{ id: number, name: string }>>('/api/v1/courses/options', data)
}