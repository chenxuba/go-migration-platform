<script setup>
import { computed, createVNode, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import dayjs from 'dayjs'
import { CloseOutlined, DownOutlined, ExclamationCircleFilled, EyeInvisibleOutlined, EyeOutlined, QuestionCircleOutlined } from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'
import { useWindowSize } from '@vueuse/core'
import { useUserStore } from '@/stores/user'
import { getStudentPhoneNumberApi } from '~/api/common/config'
import { approveApprovalApi, getApprovalDetailApi, refuseApprovalApi } from '~/api/finance-center/approval-manage'
import { closeOrderApi, getOrderDetailApi } from '~/api/finance-center/order-manage'
import { getRechargeAccountByStudentApi } from '~/api/finance-center/recharge-account'
import messageService from '~/utils/messageService'
import { downloadOrderReceiptPdf, openOrderReceiptPage } from '~/utils/order-receipt'
import RechargeAccountDetailDrawer from './recharge-account-detail-drawer.vue'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  orderId: {
    type: String,
    default: '',
  },
})

const emit = defineEmits(['update:open', 'updated', 'closed'])
const router = useRouter()
const userStore = useUserStore()

const openDrawer = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const { width } = useWindowSize()
const drawerWidth = computed(() => {
  if (width.value < 990) {
    return '100%'
  }
  else if (width.value < 1165) {
    return '80%'
  }
  else {
    return '1165px'
  }
})

const loading = ref(false)
const operateLoading = ref(false)
const detail = ref(null)
const approvalDetail = ref(null)
const openApprovalFlowModal = ref(false)
const openRejectModal = ref(false)
const openApproveModal = ref(false)
const openRechargePayDrawer = ref(false)
const openRechargeAccountDetailDrawer = ref(false)
const rechargeAccountDetailPayload = ref({})
/** 从嵌套的储值账户详情点订单编号时切换当前展示的订单，避免再套一层订单详情抽屉 */
const internalViewOrderId = ref('')
const resolvedOrderId = computed(() => {
  if (!props.open)
    return ''
  const internal = String(internalViewOrderId.value || '').trim()
  if (internal)
    return internal
  return String(props.orderId || '').trim()
})
const rejectRemark = ref('')
const approveRemark = ref('')
const decryptedPhone = ref('')
const isPhoneDecrypted = ref(false)
const phoneLoading = ref(false)
const rechargeAccountLoading = ref(false)
const rechargeAccountInfo = ref(null)

const orderStatusMap = {
  1: '待付款',
  2: '审批中',
  3: '已完成',
  4: '已关闭',
  5: '已作废',
  6: '待处理',
  7: '退费中',
  8: '已退费',
}

const orderTypeMap = {
  1: '报名续费',
  2: '储值账户充值',
  3: '退课',
  4: '储值账户退费',
  5: '转课',
  6: '退教材费',
  7: '退学杂费',
}

const orderSourceMap = {
  1: '线下办理',
  2: '微校报名',
  3: '线下导入',
  4: '续费订单',
}

const approvalStatusMap = {
  0: '审批中',
  1: '审批通过',
  2: '审批拒绝',
}

const payMethodMap = {
  1: '微信',
  2: '支付宝',
  3: '银行转账',
  4: 'POS机',
  5: '现金',
  6: '其他',
}

const lessonTypeMap = {
  1: '班级授课',
  2: '1v1授课',
}

const chargingModeMap = {
  1: '按课时',
  2: '按时段',
  3: '按金额',
}

const handleTypeMap = {
  0: '无',
  1: '新报',
  2: '续费',
  3: '转课',
}

const approvalInfo = computed(() => detail.value?.approvalInfo || null)
const approvalId = computed(() => {
  return String(
    approvalInfo.value?.approvalId
    || detail.value?.approveId
    || '',
  ).trim()
})
const approvalNumberText = computed(() => {
  return approvalInfo.value?.approvalNumber || approvalDetail.value?.approveNumber || '-'
})
const currentApproverText = computed(() => {
  if (approvalStatusText.value !== '审批中') {
    return '-'
  }
  if (approvalInfo.value?.currentApprover) {
    return approvalInfo.value.currentApprover
  }
  const currentStage = Array.isArray(approvalDetail.value?.approveFlows)
    ? approvalDetail.value.approveFlows.find(item => item?.isCurrentStage)
    : null
  const staffNames = Array.isArray(currentStage?.flowStaffs)
    ? currentStage.flowStaffs.map(item => item?.staffName).filter(Boolean)
    : []
  return staffNames.length ? staffNames.join('、') : '-'
})
const applicantNameText = computed(() => {
  return approvalInfo.value?.applicantName || approvalDetail.value?.initiateStaffName || '-'
})
const approvalTimeText = computed(() => {
  return formatDate(approvalInfo.value?.approvalTime || approvalDetail.value?.initiateTime)
})
const approvalFinishTimeText = computed(() => {
  return formatDate(approvalInfo.value?.finishTime || approvalDetail.value?.finishTime)
})
const approvalTriggerReasonText = computed(() => approvalDetail.value?.initiateReason || '-')
const approvalFlowList = computed(() => Array.isArray(approvalDetail.value?.approveFlows) ? approvalDetail.value.approveFlows : [])
const currentApprovalFlow = computed(() => approvalFlowList.value.find(item => item?.isCurrentStage) || null)
const canApprovalOperate = computed(() => {
  if (approvalStatusText.value !== '审批中') {
    return false
  }
  const currentInstUserId = Number(userStore.instUserId || 0)
  if (!currentInstUserId || !currentApprovalFlow.value) {
    return false
  }
  return Array.isArray(currentApprovalFlow.value.flowStaffs)
    && currentApprovalFlow.value.flowStaffs.some(item => Number(item?.staffId || 0) === currentInstUserId)
})
const showApprovalFlowAction = computed(() => approvalFlowList.value.length > 0)
const showApprovalSection = computed(() => {
  return !!(approvalNumberText.value !== '-' || approvalId.value)
})
const orderItems = computed(() => Array.isArray(detail.value?.orderItems) ? detail.value.orderItems : [])
const isRechargeOrderDetail = computed(() => Number(detail.value?.orderType || 0) === 2)
const isRefundRechargeOrderDetail = computed(() => Number(detail.value?.orderType || 0) === 4)
/** 退费类订单：流水区块展示为「退款记录」 */
const isRefundOrderPaymentSection = computed(() =>
  [3, 4, 6, 7].includes(Number(detail.value?.orderType || 0)),
)
const paymentRecords = computed(() => Array.isArray(detail.value?.paymentRecords) ? detail.value.paymentRecords : [])
const actualPaymentAmount = computed(() => Math.abs(Number(
  detail.value?.paidAmount
  ?? detail.value?.totalAmount
  ?? detail.value?.amount
  ?? 0,
)))
/** 实收/实退现金为 0 时，不展示支付/退款记录整块，避免出现空流水区域。 */
const showOrderPaymentRecordsBlock = computed(() => {
  if (!detail.value)
    return false
  return actualPaymentAmount.value > 0 && paymentRecords.value.length > 0
})
const showRechargeAccountSection = computed(
  () => isRechargeOrderDetail.value || isRefundRechargeOrderDetail.value,
)
const rechargeOrderTotalAmount = computed(() => Number(detail.value?.totalAmount ?? detail.value?.amount ?? 0))
const orderTotalAmount = computed(() => Number(detail.value?.totalAmount ?? detail.value?.amount ?? 0))
const orderRechargeDeductionSummary = computed(() => {
  const recharge = Number(detail.value?.rechargeAccountAmount || 0)
  const residual = Number(detail.value?.rechargeAccountResidualAmount || 0)
  const giving = Number(detail.value?.rechargeAccountGivingAmount || 0)
  const studentPhone = detail.value?.studentPhone || ''
  if (recharge <= 0 && residual <= 0 && giving <= 0) {
    return null
  }
  return {
    title: `储值账户（${studentPhone || '-'}）`,
    recharge,
    residual,
    giving,
  }
})
const refundOrderTotalAmount = computed(() => {
  if (!isRefundRechargeOrderDetail.value) {
    return 0
  }
  return Math.abs(
    Number(
      detail.value?.paidAmount
      ?? detail.value?.totalAmount
      ?? detail.value?.amount
      ?? 0,
    ),
  )
})
const refundOrderDetailRows = computed(() => {
  if (!isRefundRechargeOrderDetail.value) {
    return []
  }
  const d = detail.value
  return [
    {
      label: '扣除充值金额（元）',
      value: Math.abs(Number(d?.rechargeAccountAmount || 0)),
    },
    {
      label: '扣除赠送金额（元）',
      value: Math.abs(Number(d?.rechargeAccountGivingAmount || 0)),
    },
    {
      label: '扣除残联金额（元）',
      value: Math.abs(Number(d?.rechargeAccountResidualAmount || 0)),
    },
    {
      label: '实退金额（元）',
      value: Math.abs(Number(d?.paidAmount ?? d?.totalAmount ?? d?.amount ?? 0)),
      highlight: true,
    },
  ]
})
const rechargeOrderDetailRows = computed(() => [
  {
    label: '充值金额（元）',
    value: Number(detail.value?.rechargeAccountAmount || 0),
  },
  {
    label: '赠送金额（元）',
    value: Number(detail.value?.rechargeAccountGivingAmount || 0),
  },
  {
    label: '残联金额（元）',
    value: Number(detail.value?.rechargeAccountResidualAmount || 0),
  },
  {
    label: '应收金额（元）',
    value: Number(detail.value?.totalAmount ?? detail.value?.amount ?? 0),
    highlight: true,
  },
])
const rechargeAccountDisplayName = computed(() => {
  return rechargeAccountInfo.value?.accountName || rechargeAccountInfo.value?.id || '-'
})
const rechargeAccountStudentsText = computed(() => {
  const students = Array.isArray(rechargeAccountInfo.value?.students) ? rechargeAccountInfo.value.students : []
  const names = students.map(item => item?.name).filter(Boolean)
  return names.length ? names.join('、') : '-'
})
const orderTagText = computed(() => {
  if (Array.isArray(detail.value?.orderTags) && detail.value.orderTags.length) {
    return detail.value.orderTags.map(item => item?.tagName).filter(Boolean).join('、') || '-'
  }
  if (Array.isArray(detail.value?.orderTagNames) && detail.value.orderTagNames.length) {
    return detail.value.orderTagNames.join('、')
  }
  return '-'
})
const showBadDebtBanner = computed(() => {
  return !!detail.value?.isBadDebt
})
const showArrearBanner = computed(() => {
  return !!detail.value && detail.value.orderStatus !== 1 && !detail.value.isBadDebt && Number(detail.value.arrearAmount || 0) > 0
})

const orderStatusText = computed(() => {
  return orderStatusMap[detail.value?.orderStatus] || '-'
})
const voidedOrderActionText = computed(() => {
  if (detail.value?.orderStatus !== 5) {
    return ''
  }
  const handledFlows = approvalFlowList.value.filter(flow => isHandledApprovalFlow(flow))
  const latestHandledFlow = handledFlows.length ? handledFlows[handledFlows.length - 1] : null
  const operatorName = getApprovalFlowOperatorNames(latestHandledFlow) || currentApproverText.value || applicantNameText.value || '-'
  const operateTime = formatDate(approvalInfo.value?.finishTime || approvalDetail.value?.finishTime || latestHandledFlow?.operateTime)
  return `${operatorName}在 ${operateTime} 作废了订单`
})

const approvalStatusText = computed(() => {
  if (approvalInfo.value?.approvalStatus !== undefined && approvalInfo.value?.approvalStatus !== null) {
    return approvalStatusMap[approvalInfo.value.approvalStatus] || '-'
  }
  if (approvalDetail.value?.status === 1) {
    return '审批中'
  }
  if (approvalDetail.value?.status === 2) {
    return '审批通过'
  }
  if (approvalDetail.value?.status === 3) {
    return '审批拒绝'
  }
  return '-'
})

const actionButtonText = computed(() => {
  if (detail.value?.orderStatus === 2 || detail.value?.orderStatus === 3)
    return '废除订单'
  if (detail.value?.orderStatus === 1)
    return '关闭订单'
  return ''
})

const showReceiptAction = computed(() =>
  ![1, 2, 5].includes(Number(detail.value?.orderStatus || 0))
  && showOrderPaymentRecordsBlock.value
  && paymentRecords.value.length > 0,
)
const showRepaymentAction = computed(() => detail.value?.orderStatus !== 1 && detail.value?.arrearAmount > 0 && !detail.value?.isBadDebt)
const showPayAction = computed(() => detail.value?.orderStatus === 1)

watch(
  () => [props.open, resolvedOrderId.value],
  async ([open, orderId]) => {
    if (!open) {
      internalViewOrderId.value = ''
      detail.value = null
      approvalDetail.value = null
      rechargeAccountInfo.value = null
      openRechargeAccountDetailDrawer.value = false
      return
    }
    if (!orderId) {
      detail.value = null
      approvalDetail.value = null
      rechargeAccountInfo.value = null
      return
    }
    await fetchOrderDetail(orderId)
  },
  { immediate: true },
)

watch(
  () => props.orderId,
  () => {
    if (props.open)
      internalViewOrderId.value = ''
  },
)

watch(
  () => props.open,
  (open, prevOpen) => {
    if (prevOpen && !open) {
      emit('closed')
    }
  },
)

watch(
  () => detail.value?.studentId,
  () => {
    decryptedPhone.value = ''
    isPhoneDecrypted.value = false
    phoneLoading.value = false
  },
)

async function fetchOrderDetail(orderId) {
  try {
    loading.value = true
    approvalDetail.value = null
    rechargeAccountInfo.value = null
    const { result } = await getOrderDetailApi({ orderId })
    detail.value = result || null
    await fetchRechargeAccountInfo()
    await fetchApprovalDetail()
  }
  catch (error) {
    console.error('获取订单详情失败:', error)
    detail.value = null
    approvalDetail.value = null
    messageService.error(error?.message || '获取订单详情失败')
  }
  finally {
    loading.value = false
  }
}

async function fetchRechargeAccountInfo() {
  const ot = Number(detail.value?.orderType || 0)
  if (ot !== 2 && ot !== 4) {
    rechargeAccountInfo.value = null
    return
  }

  const studentId = Number(detail.value?.studentId || 0)
  if (!studentId) {
    rechargeAccountInfo.value = null
    return
  }
  try {
    rechargeAccountLoading.value = true
    const res = await getRechargeAccountByStudentApi({ studentId })
    if (res.code === 200) {
      rechargeAccountInfo.value = res.result || null
      return
    }
    rechargeAccountInfo.value = null
  }
  catch (error) {
    console.error('获取储值账户信息失败:', error)
    rechargeAccountInfo.value = null
  }
  finally {
    rechargeAccountLoading.value = false
  }
}

async function fetchApprovalDetail() {
  if (!approvalId.value) {
    approvalDetail.value = null
    return
  }

  try {
    const res = await getApprovalDetailApi({ id: approvalId.value })
    if (res.code === 200) {
      approvalDetail.value = res.result || null
      return
    }
    approvalDetail.value = null
  }
  catch (error) {
    console.error('获取审批详情失败:', error)
    approvalDetail.value = null
  }
}

function formatDate(dateStr) {
  if (isPlaceholderDate(dateStr))
    return '-'
  const value = dayjs(dateStr)
  return value.isValid() ? value.format('YYYY-MM-DD HH:mm') : '-'
}

function formatDateOnly(dateStr) {
  if (isPlaceholderDate(dateStr))
    return '-'
  const value = dayjs(dateStr)
  return value.isValid() ? value.format('YYYY-MM-DD') : '-'
}

function isPlaceholderDate(dateStr) {
  const raw = String(dateStr || '').trim()
  if (!raw) {
    return true
  }
  return raw.startsWith('0001-01-01') || raw.startsWith('0000-00-00')
}

/** 仅用于退款记录：账单操作日期（流水 payTime，仅年月日；与订单信息「账单操作时间」无关） */
function formatPaymentBillOperationDate(record) {
  const pay = record?.payTime
  if (pay && !isPlaceholderDate(pay)) {
    return formatDateOnly(pay)
  }
  return '-'
}

/** 订单信息：储值账户退费显示「账单操作时间」，其余订单显示「创建时间」。 */
const orderTimeLabelText = computed(() => {
  return Number(detail.value?.orderType || 0) === 4 ? '账单操作时间' : '创建时间'
})

const orderTimeValueText = computed(() => {
  const d = detail.value
  if (!d) {
    return '-'
  }
  if (Number(d.orderType || 0) === 4) {
    const bill = d.billFinishedTime
    if (bill && !isPlaceholderDate(bill)) {
      return formatDate(bill)
    }
  }
  return formatDate(d.createdTime)
})

function formatMoney(value) {
  return `¥${Number(value || 0).toFixed(2)}`
}

function formatMoneyPlain(value) {
  return Number(value || 0).toFixed(2)
}

const displayPhoneNumber = computed(() => {
  if (isPhoneDecrypted.value && decryptedPhone.value) {
    return decryptedPhone.value
  }
  return maskPhone(detail.value?.studentPhone)
})

function maskPhone(phone) {
  const value = String(phone || '').trim()
  if (!value)
    return '-'
  if (value.includes('****'))
    return value
  return value.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2')
}

function getLessonTypeText(type) {
  return lessonTypeMap[type] || '未知'
}

function getChargingModeText(mode) {
  return chargingModeMap[mode] || '未知'
}

function getHandleTypeText(type) {
  return handleTypeMap[type] || '未知'
}

function getPayMethodText(type) {
  return payMethodMap[type] || '未知'
}

function getOrderTypeText(type) {
  return orderTypeMap[type] || '-'
}

function getOrderSourceText(type) {
  return orderSourceMap[type] || '-'
}

async function handlePhoneToggle() {
  if (isPhoneDecrypted.value) {
    isPhoneDecrypted.value = false
    return
  }

  if (decryptedPhone.value) {
    isPhoneDecrypted.value = true
    return
  }

  const studentId = Number(detail.value?.studentId || 0)
  if (!studentId) {
    messageService.error('缺少学员ID，无法查看完整手机号')
    return
  }

  phoneLoading.value = true
  try {
    const res = await getStudentPhoneNumberApi({ studentId })
    if (res.code === 200 && res.result) {
      decryptedPhone.value = res.result
      isPhoneDecrypted.value = true
      return
    }
    messageService.error(res.message || '解密手机号失败')
  }
  catch (error) {
    console.error('解密手机号失败:', error)
    messageService.error('解密手机号失败')
  }
  finally {
    phoneLoading.value = false
  }
}

function formatQuantity(value, chargingMode) {
  const amount = Number(value || 0)
  const text = Number.isInteger(amount) ? String(amount) : amount.toFixed(2)
  if (chargingMode === 1)
    return `${text}课时`
  if (chargingMode === 2)
    return `${text}天`
  if (chargingMode === 3)
    return `${text}元`
  return text
}

function getChargingUnitText(mode) {
  if (mode === 1)
    return '课时'
  if (mode === 2)
    return '天'
  if (mode === 3)
    return '元'
  return ''
}

function getGiftLabel(mode) {
  if (mode === 1)
    return '赠送课时'
  if (mode === 2)
    return '赠送天数'
  if (mode === 3)
    return '赠送金额'
  return '赠送课时'
}

function getOrderDetailColumns(item) {
  const isTimeSlot = Number(item?.chargingMode || 0) === 2
  const columns = [
    {
      title: '报价单',
      dataIndex: 'priceList',
      key: 'priceList',
      width: 260,
      align: 'center',
      ellipsis: true,
    },
    {
      title: '购买份数',
      dataIndex: 'quantity',
      key: 'quantity',
      width: 120,
      align: 'center',
    },
    {
      title: getGiftLabel(item?.chargingMode),
      dataIndex: 'freeHours',
      key: 'freeHours',
      width: 120,
      align: 'center',
    },
    {
      title: '单课优惠',
      dataIndex: 'singleDiscount',
      key: 'singleDiscount',
      width: 120,
      align: 'center',
    },
    {
      title: '分摊整单优惠(元)',
      dataIndex: 'shareDiscount',
      key: 'shareDiscount',
      width: 160,
      align: 'center',
    },
    {
      title: '应收金额(元)',
      dataIndex: 'receivableAmount',
      key: 'receivableAmount',
      width: 140,
      align: 'center',
    },
    {
      title: '分摊优惠券（元）',
      dataIndex: 'shareCouponAmount',
      key: 'shareCouponAmount',
      width: 150,
      align: 'center',
    },
    {
      title: '分摊储值账户充值余额（元）',
      dataIndex: 'shareRechargeAccountAmount',
      key: 'shareRechargeAccountAmount',
      width: 200,
      align: 'center',
    },
    {
      title: '分摊储值账户赠送余额（元）',
      dataIndex: 'shareRechargeAccountGivingAmount',
      key: 'shareRechargeAccountGivingAmount',
      width: 200,
      align: 'center',
    },
    {
      title: '平账抵扣（元）',
      dataIndex: 'chargeAgainstAmount',
      key: 'chargeAgainstAmount',
      width: 140,
      align: 'center',
    },
    {
      title: '分摊欠费金额（元）',
      dataIndex: 'arrearAmount',
      key: 'arrearAmount',
      width: 160,
      align: 'center',
    },
    {
      title: '实付金额（元）',
      dataIndex: 'actualPaidAmount',
      key: 'actualPaidAmount',
      width: 150,
      align: 'center',
    },
  ]

  if (isTimeSlot) {
    columns.splice(3, 0,
      {
        title: '开始时间',
        dataIndex: 'startTime',
        key: 'startTime',
        width: 180,
        align: 'center',
      },
      {
        title: '结束时间',
        dataIndex: 'endTime',
        key: 'endTime',
        width: 180,
        align: 'center',
      },
      {
        title: '总天数（含赠）',
        dataIndex: 'totalDays',
        key: 'totalDays',
        width: 150,
        align: 'center',
      },
    )
    return columns
  }

  columns.splice(3, 0, {
    title: '有效期至',
    dataIndex: 'validUntil',
    key: 'validUntil',
    width: 140,
    align: 'center',
  })
  return columns
}

function formatQuoteDisplay(item) {
  const quantity = Number(item?.quoteQuantity || 0)
  const price = Number(item?.quotePrice || 0)
  if (Number(item?.chargingMode || 0) === 3 && item?.quoteName === '自定义') {
    return `充值金额${price.toFixed(2)}元（自定义）`
  }
  const unitText = getChargingUnitText(item?.chargingMode)
  const prefix = quantity > 0 && unitText
    ? `${Number.isInteger(quantity) ? quantity : quantity.toFixed(2)}${unitText}/${price.toFixed(2)}元`
    : `${price.toFixed(2)}元`

  if (item?.quoteName) {
    return `${prefix}（${item.quoteName}）`
  }
  return prefix
}

function formatValidUntil(item) {
  if (item.endDate)
    return formatDateOnly(item.endDate)
  if (item.validDate)
    return formatDateOnly(item.validDate)
  return item.hasValidDate ? '长期有效' : '-'
}

function formatDateWithWeekday(dateStr) {
  if (isPlaceholderDate(dateStr))
    return '-'
  const value = dayjs(dateStr)
  if (!value.isValid())
    return '-'
  const weekDays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
  return `${value.format('YYYY-MM-DD')}（${weekDays[value.day()]}）`
}

function formatSingleDiscount(item) {
  if (!item?.discountType || !Number(item.singleDiscountAmount || 0))
    return '-'
  if (item.discountType === 2 && Number(item.discountNumber || 0) > 0) {
    return `${item.discountNumber}折`
  }
  return `-${formatMoneyPlain(item.singleDiscountAmount)}`
}

function getOrderItemAllocationBase() {
  const total = orderItems.value.reduce((sum, current) => sum + Number(current?.receivableAmount || 0), 0)
  if (total > 0) {
    return total
  }
  return Number(orderTotalAmount.value || 0)
}

function getOrderItemShareRatio(item) {
  const base = getOrderItemAllocationBase()
  if (base <= 0) {
    return 0
  }
  return Number(item?.receivableAmount || 0) / base
}

function allocateOrderAmount(item, totalAmount) {
  return Number(totalAmount || 0) * getOrderItemShareRatio(item)
}

function formatDeductionAmount(value) {
  return `-${formatMoneyPlain(Math.abs(Number(value || 0)))}`
}

function formatNumericAmount(value) {
  return formatMoneyPlain(Number(value || 0))
}

function formatCount(value, suffix = '') {
  const amount = Number(value || 0)
  const text = Number.isInteger(amount) ? String(amount) : amount.toFixed(2)
  return `${text}${suffix}`
}

function getOrderDetailTableScrollX(item) {
  return getOrderDetailColumns(item).reduce((total, column) => total + Number(column.width || 0), 0)
}

function toOrderDetailRow(item) {
  const actualPaidAmount = allocateOrderAmount(item, detail.value?.paidAmount)
  const shareCouponAmount = 0
  const shareRechargeAccountAmount = allocateOrderAmount(item, detail.value?.rechargeAccountAmount)
  const shareRechargeAccountGivingAmount = allocateOrderAmount(item, detail.value?.rechargeAccountGivingAmount)
  const chargeAgainstAmount = allocateOrderAmount(item, detail.value?.totalChargeAgainstAmount)
  const arrearAmount = Math.max(
    Number(item?.receivableAmount || 0)
      - actualPaidAmount
      - shareRechargeAccountAmount
      - shareRechargeAccountGivingAmount
      - chargeAgainstAmount
      - shareCouponAmount,
    0,
  )

  return [{
    key: item.orderCourseDetailId,
    priceList: formatQuoteDisplay(item),
    quantity: formatCount(item.count, '份'),
    freeHours: formatQuantity(item.freeQuantity, item.chargingMode),
    validUntil: formatValidUntil(item),
    startTime: formatDateWithWeekday(item.validDate),
    endTime: formatDateWithWeekday(item.endDate),
    totalDays: formatCount(item.realQuantity, '天'),
    singleDiscount: formatSingleDiscount(item),
    shareDiscount: formatDeductionAmount(item.shareDiscount),
    receivableAmount: formatNumericAmount(item.receivableAmount),
    shareCouponAmount: formatDeductionAmount(shareCouponAmount),
    shareRechargeAccountAmount: formatDeductionAmount(shareRechargeAccountAmount),
    shareRechargeAccountGivingAmount: formatDeductionAmount(shareRechargeAccountGivingAmount),
    chargeAgainstAmount: formatNumericAmount(chargeAgainstAmount),
    arrearAmount: formatNumericAmount(arrearAmount),
    actualPaidAmount: formatNumericAmount(actualPaidAmount),
  }]
}

function handleCloseOrder() {
  if (!detail.value?.orderId) {
    return
  }
  Modal.confirm({
    title: '确定关闭订单？',
    icon: createVNode(ExclamationCircleFilled),
    centered: true,
    okText: '确定',
    cancelText: '取消',
    async onOk() {
      try {
        const res = await closeOrderApi({ orderId: detail.value.orderId })
        if (res.code === 200) {
          messageService.success('订单已关闭')
          await fetchOrderDetail(detail.value.orderId)
          emit('updated')
          return
        }
        return Promise.reject(new Error(res.message || '关闭订单失败'))
      }
      catch (error) {
        console.error('close order failed', error)
        messageService.error(error?.message || '关闭订单失败')
        return Promise.reject(error)
      }
    },
  })
}

function handlePrintReceipt() {
  if (!detail.value?.orderId) {
    messageService.warning('订单不存在')
    return
  }
  openOrderReceiptPage(detail.value.orderId, { template: 'a4' })
}

async function handleDownloadReceipt() {
  if (!detail.value?.orderId) {
    messageService.warning('订单不存在')
    return
  }
  try {
    await downloadOrderReceiptPdf(detail.value.orderId, { template: 'a4' })
  }
  catch (error) {
    console.error('download receipt failed', error)
    messageService.error('下载收据失败')
  }
}

function handleSendSms() {
  messageService.info('发送短信功能开发中')
}

function handleRepayment() {
  if (!detail.value?.orderId) {
    return
  }
  if (Number(detail.value?.orderType) === 2) {
    openRechargePayDrawer.value = true
    return
  }
  router.push({
    path: `/edu-center/registr-renewal/${detail.value.orderId}`,
    query: { step: '1' },
  })
  openDrawer.value = false
}

async function handleRechargeSubmitted() {
  if (!detail.value?.orderId) {
    return
  }
  await fetchOrderDetail(detail.value.orderId)
  emit('updated')
}

function handleViewApprovalFlow() {
  if (!approvalFlowList.value.length) {
    return
  }
  openApprovalFlowModal.value = true
}

/** 将 GetRechargeAccountByStudent 结果转为储值账户详情抽屉所需结构 */
function buildRechargeAccountDetailPayload(info) {
  if (!info?.id)
    return null
  const rechargePlusResidual = Number(info.balance || 0)
  const residual = Number(info.residualBalance || 0)
  const giving = Number(info.givingBalance || 0)
  const rechargeOnly = Math.max(0, rechargePlusResidual - residual)
  const students = Array.isArray(info.students) ? info.students : []
  return {
    rechargeAccountId: String(info.id),
    rechargeAccountName: info.accountName || '',
    phone: info.phone || '',
    balanceTotal: rechargePlusResidual + giving,
    rechargeBalance: rechargeOnly,
    residualBalance: residual,
    givingBalance: giving,
    rechargeAccountStudents: students.map(s => ({
      studentId: String(s.id || ''),
      studentName: s.name || '',
      isMainStudent: !!s.isMainStudent,
    })),
  }
}

function handleViewRechargeAccountDetail() {
  const payload = buildRechargeAccountDetailPayload(rechargeAccountInfo.value)
  if (!payload) {
    messageService.warning('暂无储值账户信息')
    return
  }
  rechargeAccountDetailPayload.value = payload
  openRechargeAccountDetailDrawer.value = true
}

function handleLinkedOrderFromRecharge(orderId) {
  const id = String(orderId || '').trim()
  if (!id)
    return
  openRechargeAccountDetailDrawer.value = false
  internalViewOrderId.value = id
}

function handleOpenRejectModal() {
  rejectRemark.value = ''
  openRejectModal.value = true
}

function handleOpenApproveModal() {
  approveRemark.value = ''
  openApproveModal.value = true
}

async function handleRejectSubmit() {
  if (!approvalId.value) {
    messageService.error('缺少审批单ID')
    return
  }
  try {
    operateLoading.value = true
    const res = await refuseApprovalApi({
      id: approvalId.value,
      remark: rejectRemark.value?.trim() || '',
    })
    if (res.code === 200) {
      messageService.success('审批拒绝成功')
      openRejectModal.value = false
      if (resolvedOrderId.value) {
        await fetchOrderDetail(resolvedOrderId.value)
      }
      emit('updated')
      return
    }
    messageService.error(res.message || '审批拒绝失败')
  }
  catch (error) {
    console.error('审批拒绝失败:', error)
    messageService.error(error?.message || '审批拒绝失败')
  }
  finally {
    operateLoading.value = false
  }
}

async function handleApproveSubmit() {
  if (!approvalId.value) {
    messageService.error('缺少审批单ID')
    return
  }
  try {
    operateLoading.value = true
    const res = await approveApprovalApi({
      id: approvalId.value,
      remark: approveRemark.value?.trim() || '',
    })
    if (res.code === 200) {
      messageService.success('审批通过成功')
      openApproveModal.value = false
      if (resolvedOrderId.value) {
        await fetchOrderDetail(resolvedOrderId.value)
      }
      emit('updated')
      return
    }
    messageService.error(res.message || '审批通过失败')
  }
  catch (error) {
    console.error('审批通过失败:', error)
    messageService.error(error?.message || '审批通过失败')
  }
  finally {
    operateLoading.value = false
  }
}

function getApprovalFlowStageTitle(flow) {
  if (flow?.isCurrentStage || flow?.status === 1) {
    return '审批中'
  }
  if (flow?.status === 2 || flow?.status === 3) {
    return '审批通过'
  }
  if (flow?.status === 4 || flow?.status === 5) {
    return '审批拒绝'
  }
  return '等待审批'
}

function getApprovalFlowStatusTag(flow) {
  if (flow?.status === 2 || flow?.status === 3) {
    return '审批通过'
  }
  if (flow?.status === 4) {
    return '审批拒绝'
  }
  if (flow?.status === 5) {
    return '已作废'
  }
  return ''
}

function getApprovalFlowOperatorNames(flow) {
  const names = Array.isArray(flow?.flowStaffs)
    ? flow.flowStaffs
        .filter(item => item?.isApproveOperate && item?.staffName)
        .map(item => item.staffName)
    : []
  return names.length ? names.join('、') : ''
}

function getApprovalFlowStaffNames(flow) {
  const names = Array.isArray(flow?.flowStaffs)
    ? flow.flowStaffs
        .map((item) => {
          if (!item?.staffName) {
            return ''
          }
          return item.isApproveOperate ? `${item.staffName}（审批人）` : item.staffName
        })
        .filter(Boolean)
    : []
  return names.length ? names.join('、') : '-'
}

function getApprovalFlowCardTitle(flow) {
  const operatorNames = getApprovalFlowOperatorNames(flow)
  if (operatorNames) {
    return operatorNames
  }
  return getApprovalFlowStageTitle(flow)
}

function isHandledApprovalFlow(flow) {
  return !!getApprovalFlowStatusTag(flow)
}
</script>

<template>
  <div>
    <a-drawer
      v-model:open="openDrawer"
      :body-style="{ padding: '0', background: '#fff' }"
      :closable="false"
      :width="drawerWidth"
      placement="right"
    >
      <template #title>
        <div class="custom-header flex justify-between h-4 flex-items-center">
          <div class="text-5">
            订单详情
          </div>
          <a-button type="text" class="close-btn" @click="openDrawer = false">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>

      <div class="px6 py4">
        <div v-if="loading" class="py-20 flex justify-center">
          <a-spin size="large" tip="加载订单详情中..." />
        </div>

        <a-empty v-else-if="!detail" description="暂无订单数据" />

        <template v-else>
          <div class="flex flex-items-start justify-between mb-5">
            <div class="flex-1 min-w-0">
              <div>订单状态：{{ orderStatusText }}</div>
              <div v-if="voidedOrderActionText" class="mt-3 w-full px-4 py-3 rounded-2 bg-#f6f7f8 text-#666">
                {{ voidedOrderActionText }}
              </div>
            </div>
            <a-space class="ml-4 flex-shrink-0">
              <a-button v-if="actionButtonText" @click="handleCloseOrder">
                {{ actionButtonText }}
              </a-button>
              <a-dropdown v-if="showReceiptAction">
                <template #overlay>
                  <a-menu>
                    <a-menu-item key="1" @click="handlePrintReceipt">
                      打印收据
                    </a-menu-item>
                    <a-menu-item key="2" @click="handleDownloadReceipt">
                      下载收据
                    </a-menu-item>
                    <a-menu-item key="3" @click="handleSendSms">
                      发送短信
                    </a-menu-item>
                  </a-menu>
                </template>
                <a-button>
                  查看收据
                  <DownOutlined :style="{ fontSize: '10px' }" />
                </a-button>
              </a-dropdown>
              <a-button v-if="showRepaymentAction" type="primary" @click="handleRepayment">
                补费
              </a-button>
              <a-button v-if="showPayAction" type="primary" @click="handleRepayment">
                去付款
              </a-button>
            </a-space>
          </div>

          <div
            v-if="showBadDebtBanner"
            class="pl-4 pr-4 py-4 mt-2 rounded-1 bg-#fff7e8 text-3.5 mb-4"
          >
            <div class="flex flex-items-center">
              <ExclamationCircleFilled class="text-#ff7a00 mr-2" />
              <span class="text-#222">
                此订单已为<span class="text-#ff7a00">坏账</span>，尚有<span class="text-#ff7a00">{{ formatMoney(detail.badDebtAmount || detail.arrearAmount) }}</span>元欠费未缴清。
              </span>
            </div>
            <div class="mt-4 text-#222">
              坏账原因：{{ detail.badDebtRemark || '-' }}
            </div>
          </div>

          <div
            v-else-if="showArrearBanner"
            class="error pl-4 mt-2 flex flex-items-center rounded-1 bg-#ffe6e6 text-#ff3333 h-10 text-3.5 mb-4"
          >
            <ExclamationCircleFilled class="text-#ff3333 mr-2" />
            此订单尚有{{ formatMoney(detail.arrearAmount) }}欠费未缴清
          </div>

          <div v-if="showApprovalSection" class="basic-info bg-#f6f7f8 rounded-4 p6 mb-6">
            <div class="t text-5 font-500 mb-5 flex justify-between flex-items-center">
              <span>审批信息</span>
              <a-space v-if="canApprovalOperate">
                <a-button danger @click="handleOpenRejectModal">
                  审批拒绝
                </a-button>
                <a-button type="primary" @click="handleOpenApproveModal">
                  审批通过
                </a-button>
              </a-space>
            </div>
            <a-descriptions :column="3" :content-style="{ color: '#888' }">
              <a-descriptions-item label="审批编号">
                <clamped-text :text="approvalNumberText" :lines="1" class="pr-30px" />
              </a-descriptions-item>
              <a-descriptions-item label="审批状态">
                <div>
                  <span class="dot" :class="{ rejected: approvalStatusText === '审批拒绝', pending: approvalStatusText === '审批中' }" />
                  <span>{{ approvalStatusText }}</span>
                </div>
              </a-descriptions-item>
              <a-descriptions-item label="当前审批人">
                <div class="flex flex-items-center">
                  <span>{{ currentApproverText }}</span>
                  <a-button
                    v-if="showApprovalFlowAction"
                    type="link"
                    class="px-0 ml-2 h-auto leading-none"
                    @click="handleViewApprovalFlow"
                  >
                    查看审批流程
                  </a-button>
                </div>
              </a-descriptions-item>
              <a-descriptions-item label="申请人">
                {{ applicantNameText }}
              </a-descriptions-item>
              <a-descriptions-item label="申请时间">
                {{ approvalTimeText }}
              </a-descriptions-item>
              <a-descriptions-item label="审批完成时间">
                {{ approvalFinishTimeText }}
              </a-descriptions-item>
              <a-descriptions-item label="触发审批条件" :span="3">
                <!-- 颜色改成 #f30 -->
                <span class="text-#f30">
                  {{ approvalTriggerReasonText }}
                </span>
              </a-descriptions-item>
            </a-descriptions>
          </div>

          <div class="stu-info mb6">
            <div class="t text-5 font-500 mb-5 flex justify-between flex-center">
              <span>学员信息</span>
            </div>
            <a-descriptions :column="3" :content-style="{ color: '#888' }">
              <a-descriptions-item label="学员姓名">
                {{ detail.studentName || '-' }}
              </a-descriptions-item>
              <a-descriptions-item label="手机号">
                <span class="flex flex-items-center">
                  {{ displayPhoneNumber }}
                  <a-spin :spinning="phoneLoading" size="small">
                    <EyeOutlined
                      v-if="isPhoneDecrypted"
                      class="cursor-pointer text-#06f text-4 ml-1"
                      @click="handlePhoneToggle"
                    />
                    <EyeInvisibleOutlined
                      v-else
                      class="cursor-pointer text-#06f text-4 ml-1"
                      @click="handlePhoneToggle"
                    />
                  </a-spin>
                </span>
              </a-descriptions-item>
            </a-descriptions>

            <div v-if="showRechargeAccountSection" class="mt-6">
              <div class="t text-5 font-500 mb-5 flex justify-between flex-center">
                <span>储值账户信息</span>
              </div>
              <a-spin :spinning="rechargeAccountLoading">
                <a-descriptions :column="3" :content-style="{ color: '#888' }">
                  <a-descriptions-item label="储值账户">
                    <span>{{ rechargeAccountDisplayName }}</span>
                    <a class="ml-2 text-#1677ff cursor-pointer" @click="handleViewRechargeAccountDetail">
                      查看储值账户详情
                    </a>
                  </a-descriptions-item>
                  <a-descriptions-item label="关联学员">
                    {{ rechargeAccountStudentsText }}
                  </a-descriptions-item>
                </a-descriptions>
              </a-spin>
            </div>
          </div>

          <div class="order-info mb6">
            <div class="t text-5 font-500 mb-5 flex justify-between flex-center">
              <span>订单信息</span>
            </div>
            <a-descriptions :column="3" :content-style="{ color: '#888' }">
              <a-descriptions-item label="订单编号">
                <clamped-text :text="detail.orderNumber || '-'" :lines="1" class="pr-30px" />
              </a-descriptions-item>
              <a-descriptions-item label="办理类型">
                {{ getOrderTypeText(detail.orderType) }}
              </a-descriptions-item>
              <a-descriptions-item label="经办人">
                {{ detail.staffName || '-' }}
              </a-descriptions-item>
              <a-descriptions-item label="经办日期">
                {{ formatDateOnly(detail.dealDate) }}
              </a-descriptions-item>
              <a-descriptions-item label="订单来源">
                {{ getOrderSourceText(detail.orderSource) }}
              </a-descriptions-item>
              <a-descriptions-item :label="orderTimeLabelText">
                {{ orderTimeValueText }}
              </a-descriptions-item>
              <a-descriptions-item label="完成时间">
                {{ formatDate(detail.finishedTime) }}
              </a-descriptions-item>
              <a-descriptions-item label="对内备注">
                {{ detail.remark || '-' }}
              </a-descriptions-item>
              <a-descriptions-item label="对外备注">
                {{ detail.externalRemark || '-' }}
              </a-descriptions-item>
              <a-descriptions-item label="订单销售员">
                {{ detail.salePersonName || '-' }}
              </a-descriptions-item>
              <a-descriptions-item label="订单标签">
                {{ orderTagText }}
              </a-descriptions-item>
            </a-descriptions>
          </div>

          <div class="order-list mb-8">
            <div class="t text-5 font-500 mb-5 flex justify-between flex-center">
              <span>{{ isRefundRechargeOrderDetail ? '退款明细' : '订单明细' }}</span>
              <div v-if="isRechargeOrderDetail" class="recharge-total">
                <span class="recharge-total-label">订单总金额：</span>
                <span class="recharge-total-value ">{{ formatMoney(rechargeOrderTotalAmount) }}</span>
              </div>
              <div v-else-if="isRefundRechargeOrderDetail" class="refund-total">
                <span class="refund-total-label">退款总金额：</span>
                <span class="refund-total-value">-¥{{ refundOrderTotalAmount.toFixed(2) }}</span>
              </div>
            </div>
            <template v-if="isRechargeOrderDetail">
              <div class="recharge-detail-card">
                <div
                  v-for="(item, index) in rechargeOrderDetailRows"
                  :key="item.label"
                  class="recharge-detail-item"
                  :class="{ 'recharge-detail-item-last': index === rechargeOrderDetailRows.length - 1 }"
                >
                  <div class="recharge-detail-label">
                    {{ item.label }}
                  </div>
                  <div class="recharge-detail-value" :class="{ 'recharge-detail-value-highlight': item.highlight }">
                    {{ formatMoneyPlain(item.value) }}
                  </div>
                </div>
              </div>
            </template>
            <template v-else-if="isRefundRechargeOrderDetail">
              <div class="recharge-detail-card refund-detail-card">
                <div
                  v-for="(item, index) in refundOrderDetailRows"
                  :key="item.label"
                  class="recharge-detail-item"
                  :class="{ 'recharge-detail-item-last': index === refundOrderDetailRows.length - 1 }"
                >
                  <div class="recharge-detail-label">
                    {{ item.label }}
                  </div>
                  <div class="recharge-detail-value" :class="{ 'recharge-detail-value-highlight': item.highlight }">
                    {{ formatMoneyPlain(item.value) }}
                  </div>
                </div>
              </div>
            </template>
            <template v-else>
              <a-descriptions :column="3" :content-style="{ color: '#888' }">
                <a-descriptions-item label="整单总优惠">
                  -{{ formatMoney(detail.orderDiscountAmount) }}
                </a-descriptions-item>
                <a-descriptions-item label="整单优惠活动">
                  -
                </a-descriptions-item>
                <a-descriptions-item label="订单总金额">
                  <div class="order-total-row">
                    <div class="order-total-amount">
                      <span class="order-total-amount-currency">¥</span>
                      <span class="order-total-amount-value">{{ formatMoneyPlain(orderTotalAmount) }}</span>
                    </div>
                    <div class="order-total-formula">
                      <a-popover placement="topRight">
                        <template #title>
                          <div style="border-bottom:1px solid #eee;padding-bottom: 6px;">
                            说明
                          </div>
                        </template>
                        <template #content>
                          <div class="text-#666">
                            报价单金额 - 单课优惠 - 分摊整单优惠 = 应收金额
                          </div>
                        </template>
                        <QuestionCircleOutlined class="text-#06f cursor-pointer" />
                      </a-popover>
                      <span>计算公式</span>
                    </div>
                  </div>
                </a-descriptions-item>
              </a-descriptions>
              <div v-if="orderRechargeDeductionSummary" class="order-recharge-summary-row">
                <span class="storage-deduction-summary__title">储值账户抵扣：</span>
                <div class="storage-deduction-summary">
                  <span>{{ orderRechargeDeductionSummary.title }}</span>
                  <span class="storage-deduction-summary__value">充值余额 <span class="storage-deduction-summary__value-value">¥{{ formatDeductionAmount(orderRechargeDeductionSummary.recharge) }}</span></span>
                  <span class="storage-deduction-summary__value">残联余额 <span class="storage-deduction-summary__value-value">¥{{ formatDeductionAmount(orderRechargeDeductionSummary.residual) }}</span></span>
                  <span class="storage-deduction-summary__value">赠送余额 <span class="storage-deduction-summary__value-value">¥{{ formatDeductionAmount(orderRechargeDeductionSummary.giving) }}</span></span>
                </div>
              </div>
            </template>

            <a-empty
              v-if="!isRechargeOrderDetail && !isRefundRechargeOrderDetail && !orderItems.length"
              description="暂无订单明细"
            />

            <template v-if="!isRechargeOrderDetail && !isRefundRechargeOrderDetail">
              <div v-for="item in orderItems" :key="item.orderCourseDetailId" class="list mt-4 mb-4 last:mb-0">
                <a-row class="bg-#005ce60f px6 py4 border border-solid border-#eee border-b-none">
                  <a-col :span="24" class="flex justify-between">
                    <div class="flex flex-items-center">
                      <div class="text-4 font-500 mr-3">
                        {{ item.courseName || '-' }}
                      </div>
                      <a-space>
                        <span v-if="item.lessonType" class="bg-#e6f0ff text-#06f text-3 rounded-10 px3 py1">
                          {{ getLessonTypeText(item.lessonType) }}
                        </span>
                        <span v-if="item.chargingMode" class="bg-#e6f0ff text-#06f text-3 rounded-10 px3 py1">
                          {{ getChargingModeText(item.chargingMode) }}
                        </span>
                      </a-space>
                    </div>
                    <div class="flex flex-items-center">
                      {{ getHandleTypeText(item.handleType) }}
                    </div>
                  </a-col>
                </a-row>
                <a-table
                  :data-source="toOrderDetailRow(item)"
                  :columns="getOrderDetailColumns(item)"
                  :pagination="false"
                  :scroll="{ x: getOrderDetailTableScrollX(item) }"
                  size="small"
                  bordered
                />
              </div>
            </template>
          </div>

          <div v-if="showOrderPaymentRecordsBlock" class="pay-data">
            <div class="t text-5 font-500 mb-5 flex justify-between flex-center">
              <span>{{ isRefundOrderPaymentSection ? '退款记录' : '支付记录' }}</span>
            </div>

            <a-empty
              v-if="!paymentRecords.length"
              :description="isRefundOrderPaymentSection ? '暂无退款记录' : '暂无支付记录'"
            />

            <div v-for="(item, index) in paymentRecords" :key="item.paymentId" class="items-list mb3">
              <div class="t inline-flex items-center bg-#f0f5fe px-4 py-1 rounded-lt-2 rounded-rt-2 whitespace-nowrap">
                {{ isRefundOrderPaymentSection ? '退款记录' : '支付记录' }}{{ index + 1 }}
                <span v-if="item.createdTime" class="text-#666 ml-1">
                  （实际交易时间：{{ formatDate(item.createdTime) }}）
                </span>
              </div>
              <div class="bg-#fbfbfb px6 py4">
                <a-descriptions
                  v-if="isRefundOrderPaymentSection"
                  :column="3"
                  :content-style="{ color: '#888' }"
                >
                  <a-descriptions-item label="退款方式">
                    {{ getPayMethodText(item.payMethod) }}
                  </a-descriptions-item>
                  <a-descriptions-item label="实退金额（元）">
                    {{ formatMoneyPlain(item.payAmount) }}
                  </a-descriptions-item>
                  <a-descriptions-item label="创建时间">
                    {{ formatDate(item.createdTime) }}
                  </a-descriptions-item>
                  <a-descriptions-item label="退款账户">
                    {{ item.accountName || '默认账户' }}
                  </a-descriptions-item>
                  <a-descriptions-item label="支付单号">
                    -
                  </a-descriptions-item>
                  <a-descriptions-item label="对方账户">
                    -
                  </a-descriptions-item>
                  <a-descriptions-item label="账单操作日期" :span="3">
                    {{ formatPaymentBillOperationDate(item) }}
                  </a-descriptions-item>
                  <a-descriptions-item label="账单备注" :span="3">
                    {{ item.remark || '-' }}
                  </a-descriptions-item>
                </a-descriptions>
                <a-descriptions
                  v-else
                  :column="3"
                  :content-style="{ color: '#888' }"
                >
                  <a-descriptions-item label="支付金额">
                    {{ formatMoney(item.payAmount) }}
                  </a-descriptions-item>
                  <a-descriptions-item label="创建时间">
                    {{ formatDate(item.createdTime) }}
                  </a-descriptions-item>
                  <a-descriptions-item label="收款方式">
                    {{ getPayMethodText(item.payMethod) }}
                  </a-descriptions-item>
                  <a-descriptions-item label="收款账户">
                    {{ item.accountName || '默认账户' }}
                  </a-descriptions-item>
                  <a-descriptions-item label="支付单号">
                    -
                  </a-descriptions-item>
                  <a-descriptions-item label="对方账户">
                    -
                  </a-descriptions-item>
                  <a-descriptions-item label="支付日期">
                    {{ formatDateOnly(item.payTime) }}
                  </a-descriptions-item>
                  <a-descriptions-item label="账单备注">
                    {{ item.remark || '-' }}
                  </a-descriptions-item>
                </a-descriptions>
              </div>
            </div>
          </div>
        </template>
      </div>
      <a-modal
        v-model:open="openApprovalFlowModal"
        title="审批流程"
        :footer="null"
        width="560px"
        centered
        :body-style="{ padding: '0 24px 24px' }"
      >
        <div class="approval-flow-modal pt-4">
          <div class="approval-flow-item approval-flow-item-start">
            <div class="approval-flow-start-text">
              申请人：{{ applicantNameText }}
            </div>
          </div>
          <div v-for="flow in approvalFlowList" :key="flow.step" class="approval-flow-item">
            <div class="approval-flow-card">
              <div class="approval-flow-card-header">
                <div class="approval-flow-card-title">
                  {{ getApprovalFlowCardTitle(flow) }}
                </div>
                <span
                  v-if="getApprovalFlowStatusTag(flow)"
                  class="approval-flow-status-tag"
                  :class="{ rejected: getApprovalFlowStatusTag(flow) === '审批拒绝' }"
                >
                  {{ getApprovalFlowStatusTag(flow) }}
                </span>
              </div>
              <div class="approval-flow-card-body">
                <template v-if="isHandledApprovalFlow(flow)">
                  <div class="approval-flow-card-line">
                    审批时间：{{ formatDate(flow.operateTime) }}
                  </div>
                  <div class="approval-flow-card-line">
                    审批备注：{{ flow.remark || '-' }}
                  </div>
                </template>
                <template v-else>
                  <div class="approval-flow-card-line">
                    可审批人：{{ getApprovalFlowStaffNames(flow) }}
                  </div>
                </template>
              </div>
            </div>
          </div>
        </div>
      </a-modal>
      <a-modal
        v-model:open="openRejectModal"
        title="审批拒绝"
        width="560px"
        centered
        :footer="null"
        :body-style="{ padding: '0 24px 24px' }"
      >
        <div class="pt-4 text-#666 mb-4">
          拒绝后，审批关联的订单会被作废处理
        </div>
        <a-textarea
          v-model:value="rejectRemark"
          :maxlength="100"
          :auto-size="{ minRows: 6, maxRows: 6 }"
          placeholder="选填，备注最多100字"
        />
        <div class="flex justify-end mt-6">
          <a-space>
            <a-button danger :loading="operateLoading" @click="handleRejectSubmit">
              拒绝
            </a-button>
            <a-button :disabled="operateLoading" @click="openRejectModal = false">
              再想想
            </a-button>
          </a-space>
        </div>
      </a-modal>
      <a-modal
        v-model:open="openApproveModal"
        title="审批通过"
        width="560px"
        centered
        :footer="null"
        :body-style="{ padding: '0 24px 24px' }"
      >
        <div class="pt-4">
          <a-textarea
            v-model:value="approveRemark"
            :maxlength="100"
            :auto-size="{ minRows: 6, maxRows: 6 }"
            placeholder="选填，备注最多100字"
          />
        </div>
        <div class="flex justify-end mt-6">
          <a-space>
            <a-button :disabled="operateLoading" @click="openApproveModal = false">
              取消
            </a-button>
            <a-button type="primary" :loading="operateLoading" @click="handleApproveSubmit">
              确定
            </a-button>
          </a-space>
        </div>
      </a-modal>
      <recharge-order-pay-drawer
        v-model:open="openRechargePayDrawer"
        :sale-order-id="detail?.orderId"
        @submitted="handleRechargeSubmitted"
      />
      <recharge-account-detail-drawer
        v-model:open="openRechargeAccountDetailDrawer"
        :account="rechargeAccountDetailPayload"
        @updated="fetchRechargeAccountInfo"
        @open-linked-order-detail="handleLinkedOrderFromRecharge"
      />
    </a-drawer>
  </div>
</template>

<style lang="less" scoped>
@keyframes icon-rotate {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(180deg);
  }
}

.close-btn {
  &:hover {
    background: transparent;

    .close-icon {
      animation: icon-rotate 0.3s linear;
    }
  }
}

span.dot {
  border-radius: 50%;
  display: inline-block;
  height: 6px;
  position: relative;
  vertical-align: middle;
  width: 6px;
  margin-right: 4px;
  background: #00cc33;

  &.rejected {
    background: #ff4d4f;
  }

  &.pending {
    background: #faad14;
  }
}

:deep(.ant-table) {
  .ant-table-thead > tr > th {
    background: #f6f7f8 !important;
    color: #222 !important;
    font-weight: 500;
    border-bottom: 1px solid #eee;
  }

  .ant-table-tbody > tr > td {
    color: #000000a6;
    padding: 12px 16px;
  }

  .ant-table-tbody > tr {
    background: #fbfbfb;
  }

  .ant-table-tbody > tr:hover > td {
    background: #f5f5f5 !important;
  }
}

.approval-flow-modal {
  padding-left: 8px;
}

.approval-flow-item {
  position: relative;
  padding-left: 34px;
  padding-bottom: 18px;
}

.approval-flow-item:last-child {
  padding-bottom: 0;
}

.approval-flow-item::before {
  content: '';
  position: absolute;
  left: 0;
  top: 6px;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #4096ff;
  box-shadow: 0 0 0 3px rgba(64, 150, 255, 0.12);
}

.approval-flow-item::after {
  content: '';
  position: absolute;
  left: 4px;
  top: 18px;
  bottom: -2px;
  width: 1px;
  background: #d0e4ff;
}

.approval-flow-item:last-child::after {
  display: none;
}

.approval-flow-item-start {
  display: flex;
  align-items: center;
  min-height: 22px;
}

.approval-flow-start-text {
  color: #222;
  font-size: 16px;
  line-height: 24px;
  font-weight: 600;
}

.approval-flow-card {
  overflow: hidden;
  border-radius: 12px;
  background: #fafafa;
}

.approval-flow-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid #f0f0f0;
}

.approval-flow-card-title {
  color: #222;
  font-size: 16px;
  line-height: 24px;
  font-weight: 600;
}

.approval-flow-status-tag {
  flex-shrink: 0;
  padding: 2px 10px;
  border-radius: 999px;
  color: #16a34a;
  font-size: 12px;
  line-height: 20px;
  background: #dcfce7;

  &.rejected {
    color: #f33;
    background: #ffe6e6;
  }
}

.approval-flow-card-body {
  padding: 14px 20px;
  color: #666;
  font-size: 14px;
  line-height: 22px;
}

.approval-flow-card-line + .approval-flow-card-line {
  margin-top: 8px;
}

.recharge-total {
  display: flex;
  align-items: baseline;
  gap: 6px;
}

.recharge-total-label {
  color: #666;
  font-size: 14px;
}

.recharge-total-value {
  color: #1677ff;
  font-size: 24px;
  font-weight: 700;
  font-family: DINAlternate-Bold, DINAlternate, sans-serif;
}

.refund-total {
  display: flex;
  align-items: baseline;
  gap: 6px;
}

.refund-total-label {
  color: #666;
  font-size: 14px;
}

.refund-total-value {
  color: #ff3333;
  font-size: 24px;
  font-weight: 700;
  font-family: DINAlternate-Bold, DINAlternate, sans-serif;
}

.order-total-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  gap: 12px;
}

.order-total-amount {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  white-space: nowrap;
  transform: translateY(-5px);
}

.order-total-formula {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  color: #666;
  font-size: 14px;
  line-height: 22px;
  white-space: nowrap;
  flex-shrink: 0;
}

.order-total-amount-currency {
  color: #1677ff;
  font-size: 20px;
  line-height: 1;
  font-weight: 700;
  font-family: DINAlternate-Bold, DINAlternate, sans-serif;
}

.order-total-amount-value {
  color: #1677ff;
  font-size: 30px;
  line-height: 1;
  font-weight: 700;
  font-family: DINAlternate-Bold, DINAlternate, sans-serif;
}

.storage-deduction-summary {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 16px;
}

.order-recharge-summary-row {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 30px;
}

.storage-deduction-summary__title {
  color: #333;
  white-space: nowrap;
}

.storage-deduction-summary__value {
  color: #333;
}

.storage-deduction-summary__value-value {
  color: #ff4d4f;
}

.recharge-detail-card {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  border: 1px solid #d9e5ff;
  border-radius: 12px;
  overflow: hidden;
  background: #fbfcff;
  margin-bottom: 8px;
}

.refund-detail-card {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.recharge-detail-item {
  padding: 16px 28px;
  border-right: 1px solid #d9e5ff;
}

.recharge-detail-item-last {
  border-right: 0;
}

.recharge-detail-label {
  font-size: 14px;
  color: #222;
  line-height: 1.3;
  font-weight: 600;
}

.recharge-detail-value {
  margin-top: 8px;
  font-size: 20px;
  line-height: 1;
  color: #666;
  font-weight: 600;
  font-family: DINAlternate-Bold, DINAlternate, sans-serif;
}

.recharge-detail-value-highlight {
  color: #222;
  font-weight: 700;
}
</style>
