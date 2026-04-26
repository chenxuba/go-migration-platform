<script setup>
import * as qiniu from 'qiniu-js'
import { CloseOutlined, PlusOutlined, QuestionCircleFilled, QuestionCircleOutlined } from '@ant-design/icons-vue'
import { h } from 'vue'
import dayjs from 'dayjs'
import {
  calculateRefundTuitionAccountHandlingFeeApi,
  createRefundTuitionAccountOrderApi,
  estimateRefundTuitionAccountValuableTuitionApi,
  getTuitionAccountRefundOwedSummaryApi,
  payRefundTuitionAccountOrderApi,
} from '@/api/edu-center/tuition-account'
import { getOrderTagListPagedApi } from '@/api/finance-center/order-tag'
import { getQiniuToken } from '@/api/qiniu'
import StaffSelect from '@/components/common/staff-select.vue'
import { useUserStore } from '@/stores/user'
import emitter, { EVENTS } from '@/utils/eventBus'
import messageService from '@/utils/messageService'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  record: {
    type: Object,
    default: () => ({}),
  },
})
const emit = defineEmits(['update:open', 'submitted'])

const openDrawer = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const userStore = useUserStore()
const currentInstUserId = computed(() => userStore.instUserId)
const getCurrentDate = () => dayjs().format('YYYY-MM-DD')

function initFormState() {
  return {
    surplusClass: 0,
    effectiveDate: '',
    dropTheClassNumber: '',
    autoFinishCourse: true,
    originalPriceRefund: false,
    price: '0.00',
    payType: 1,
    dropPayPrice: undefined,
    dropPayAccount: '1',
    date: getCurrentDate(),
    orderLabel: [],
    salesperson: currentInstUserId.value || undefined,
    remarks1: '',
    remarks2: '',
    billRemarks: '',
    fileList: [],
  }
}

function createDefaultOwedSummary() {
  return {
    arrearAmountTotal: 0,
    badDebtAmountTotal: 0,
    orderId: '0',
    orderType: 0,
  }
}

function createDefaultEstimate() {
  return {
    tuitionAccountId: '',
    quantity: 0,
    freeQuantity: 0,
    tuition: 0,
    subAccounts: [],
  }
}

function createDefaultCalcResult() {
  return {
    tuitionAccountId: '',
    refundAmount: 0,
    totalOriginalRefundAmount: 0,
    totalArrearDeduction: 0,
    handlingFee: 0,
    lessonChargingMode: 0,
    paidRefundQuantity: 0,
    giftRefundQuantity: 0,
    arrearAmountTotal: 0,
    badDebtAmountTotal: 0,
    orderId: '0',
    orderType: 0,
    details: [],
  }
}

const openModal = ref(false)
const calcPreviewOpen = ref(false)
const feeCalcOpen = ref(false)
const current = ref(0)
const formRef = ref(null)
const previewVisible = ref(false)
const previewImage = ref('')
const previewTitle = ref('')
const readonly = ref(true)
const nextLoading = ref(false)
const estimateLoading = ref(false)
const orderLabelLoading = ref(false)
const submitLoading = ref(false)
const pendingRefundOrderId = ref('')
const pendingRefundNeedPay = ref(false)
const owedSummary = ref(createDefaultOwedSummary())
const estimateResult = ref(createDefaultEstimate())
const calcResult = ref(createDefaultCalcResult())
const orderLabelOptions = ref([])

const formState = reactive(initFormState())

const tuitionAccountId = computed(() => String(props.record?.id || props.record?.tuitionAccountId || ''))
const lessonChargingMode = computed(() => Number(props.record?.lessonChargingMode || 0))
const isAmountMode = computed(() => [3, 4].includes(lessonChargingMode.value))
const quantityUnit = computed(() => {
  if (isAmountMode.value)
    return '元'
  return lessonChargingMode.value === 2 ? '天' : '课时'
})
const remainLabelText = computed(() => {
  if (isAmountMode.value)
    return '剩余金额'
  return lessonChargingMode.value === 2 ? '剩余天数' : '剩余课时'
})
const validityLabelText = computed(() => (
  lessonChargingMode.value === 2 ? '有效时段' : '有效期至'
))
const lessonTypeText = computed(() => {
  const type = Number(props.record?.lessonType || 0)
  if (type === 1)
    return '班级授课'
  if (type === 2)
    return '1对1授课'
  return ''
})
const chargingModeText = computed(() => {
  if (lessonChargingMode.value === 1)
    return '课时'
  if (lessonChargingMode.value === 2)
    return '时段'
  if ([3, 4].includes(lessonChargingMode.value))
    return '金额'
  return ''
})
const courseName = computed(() => props.record?.lessonName || props.record?.productName || '-')
const maxRefundQuantity = computed(() => {
  const remainQuantity = Number(props.record?.remainQuantity || 0)
  const remainFreeQuantity = Number(props.record?.remainFreeQuantity || 0)
  if (isAmountMode.value)
    return roundAmount(remainQuantity)
  return roundAmount(remainQuantity + remainFreeQuantity)
})
const remainValueText = computed(() => {
  const remainQuantity = Number(props.record?.remainQuantity || 0)
  const remainFreeQuantity = Number(props.record?.remainFreeQuantity || 0)
  if (isAmountMode.value)
    return `¥ ${formatMoney(remainQuantity)}`
  const total = roundAmount(remainQuantity + remainFreeQuantity)
  if (remainFreeQuantity > 0) {
    return `${formatCount(total)}${quantityUnit.value}（含赠 ${formatCount(remainFreeQuantity)}${quantityUnit.value}）`
  }
  return `${formatCount(total)}${quantityUnit.value}`
})
const validityValueText = computed(() => {
  if (lessonChargingMode.value === 2) {
    const start = formatDate(props.record?.validDate || props.record?.activedAt)
    const end = formatDate(props.record?.endDate || props.record?.expireTime)
    if (start === '-' || end === '-')
      return '-'
    return `${start} ~ ${end}`
  }
  if (!props.record?.enableExpireTime)
    return '不限制'
  return formatDate(props.record?.expireTime)
})
const isFullRefund = computed(() => {
  if (formState.dropTheClassNumber === '' || formState.dropTheClassNumber === null || formState.dropTheClassNumber === undefined) {
    return false
  }
  return almostEqual(Number(formState.dropTheClassNumber), maxRefundQuantity.value)
})
const estimateAmount = computed(() => Number(estimateResult.value?.tuition || 0))
const estimatedPaidQuantity = computed(() => Number(estimateResult.value?.quantity || 0))
const estimatedFreeQuantity = computed(() => Number(estimateResult.value?.freeQuantity || 0))
const footerEstimateSummary = computed(() => {
  return `退还 ${formatCount(estimatedPaidQuantity.value)}${quantityUnit.value}，退赠 ${formatCount(estimatedFreeQuantity.value)}${quantityUnit.value}（赠送不计入总计）`
})
const confirmSummaryText = computed(() => {
  const paidQuantity = Number(calcResult.value?.paidRefundQuantity || 0)
  const freeQuantity = Number(calcResult.value?.giftRefundQuantity || 0)
  return `退还 ${formatCount(paidQuantity)}${quantityUnit.value}，退赠 ${formatCount(freeQuantity)}${quantityUnit.value}`
})
const previewArrearAmount = computed(() => Number(calcResult.value?.arrearAmountTotal || owedSummary.value?.arrearAmountTotal || 0))
const previewBadDebtAmount = computed(() => Number(calcResult.value?.badDebtAmountTotal || owedSummary.value?.badDebtAmountTotal || 0))
const previewOriginalRefundAmount = computed(() => Number(calcResult.value?.totalOriginalRefundAmount || 0))
const previewRefundAmount = computed(() => Number(calcResult.value?.refundAmount || 0))
const previewArrearDeductionAmount = computed(() => Number(calcResult.value?.totalArrearDeduction || 0))
const previewHandlingFeeAmount = computed(() => Number(calcResult.value?.handlingFee || 0))
const cashierRefundAmount = computed(() => Number(calcResult.value?.refundAmount || formState.price || 0))
const cashierActualRefundAmount = computed(() => Number(formState.dropPayPrice || 0))
const cashierRefundDiff = computed(() => roundAmount(cashierRefundAmount.value - cashierActualRefundAmount.value))
const cashierDefaultActualRefundAmount = computed(() => {
  const handlingFee = formState.originalPriceRefund ? previewHandlingFeeAmount.value : 0
  return roundAmount(Math.max(cashierRefundAmount.value - handlingFee, 0))
})
const cashierRefundTipText = computed(() => {
  if (isEmptyAmountValue(formState.dropPayPrice))
    return ''
  if (cashierRefundDiff.value > 0.009)
    return `应退金额：¥${formatMoney(cashierRefundAmount.value)}，手续费 ¥${formatMoney(cashierRefundDiff.value)}`
  if (cashierRefundDiff.value < -0.009)
    return `应退金额：¥${formatMoney(cashierRefundAmount.value)}，亏损费 ¥${formatTrimMoney(Math.abs(cashierRefundDiff.value))}`
  return `应退金额：¥${formatMoney(cashierRefundAmount.value)}`
})
const showCashierFeeTip = computed(() => Math.abs(cashierRefundDiff.value) > 0.009)
const cashierFeeTipTitle = computed(() => (cashierRefundDiff.value < -0.009 ? '亏损费' : '手续费'))
const cashierFeeTipContent = computed(() => (
  cashierRefundDiff.value < -0.009
    ? '指机构需要补偿学员的费用'
    : '指机构在办理业务时额外向学员收取的费用'
))
const shouldShowCalcPreviewModal = computed(() => previewArrearAmount.value > 0.009 || previewBadDebtAmount.value > 0.009)
const showPreviewHandlingFee = computed(() => formState.originalPriceRefund && (previewBadDebtAmount.value > 0.009 || previewHandlingFeeAmount.value > 0.009))
const previewStatusText = computed(() => {
  if (previewArrearAmount.value > 0.009 && previewBadDebtAmount.value > 0.009)
    return '欠费/坏账'
  if (previewBadDebtAmount.value > 0.009)
    return '坏账'
  if (previewArrearAmount.value > 0.009)
    return '欠费'
  return ''
})
const previewDescriptionText = computed(() => {
  if (previewArrearAmount.value > 0.009)
    return '退课成功后，此订单欠费金额将被完全抵扣，无需补费'
  if (previewBadDebtAmount.value > 0.009)
    return '应退金额包含订单坏账的金额。如果要剔除坏账金额可以先前往订单列表取消坏账，或在下一步修改应退金额'
  return '请确认本次应退金额'
})
const previewInfoRows = computed(() => {
  const rows = [
    { label: '退课金额', value: previewRefundAmount.value },
  ]
  if (previewArrearAmount.value > 0.009 || previewBadDebtAmount.value > 0.009)
    rows.push({ label: '订单欠费', value: previewArrearAmount.value })
  if (previewBadDebtAmount.value > 0.009)
    rows.push({ label: '坏账金额', value: previewBadDebtAmount.value })
  return rows
})
const feeCalcDetails = computed(() => (Array.isArray(calcResult.value?.details) ? calcResult.value.details : []))
const feeCalcUnitPriceText = value => (Number(value || 0) > 0 ? `${formatMoney(value)}元/${quantityUnit.value}` : '-')
const feeCalcMoneyText = value => (Number(value || 0) > 0 ? `¥ ${formatMoney(value)}` : '-')
const feeCalcQuantityText = value => `${formatCount(value)}${quantityUnit.value}`
const feeCalcColumns = [
  { title: '订单号', dataIndex: 'orderNumber', key: 'orderNumber', width: 220 },
  { title: '原价', dataIndex: 'originalUnitPriceText', key: 'originalUnitPriceText', width: 160 },
  { title: '优惠后金额', dataIndex: 'discountedUnitPriceText', key: 'discountedUnitPriceText', width: 180 },
  { title: '应收学费金额', dataIndex: 'shouldTuitionText', key: 'shouldTuitionText', width: 150 },
  { title: '退/转学员金额', dataIndex: 'transferredTuitionText', key: 'transferredTuitionText', width: 150 },
  { title: '已课消数量', dataIndex: 'consumedQuantityText', key: 'consumedQuantityText', width: 130 },
  { title: '原价应退金额', dataIndex: 'originalRefundAmountText', key: 'originalRefundAmountText', width: 150 },
]
const feeCalcTableData = computed(() => [
  ...feeCalcDetails.value.map((item, index) => ({
    key: item.orderNumber || index,
    orderNumber: item.orderNumber || '-',
    originalUnitPriceText: feeCalcUnitPriceText(item.originalUnitPrice),
    discountedUnitPriceText: feeCalcUnitPriceText(item.discountedUnitPrice),
    shouldTuitionText: feeCalcMoneyText(item.shouldTuition),
    transferredTuitionText: feeCalcMoneyText(item.transferredTuition),
    consumedQuantityText: feeCalcQuantityText(item.consumedQuantity),
    originalRefundAmountText: feeCalcMoneyText(item.originalRefundAmount),
  })),
  {
    key: 'total',
    orderNumber: '总计',
    originalUnitPriceText: '',
    discountedUnitPriceText: '',
    shouldTuitionText: '',
    transferredTuitionText: '',
    consumedQuantityText: '',
    originalRefundAmountText: `¥ ${formatMoney(previewOriginalRefundAmount.value)}`,
    isTotal: true,
  },
])

watch(isFullRefund, (value) => {
  formState.autoFinishCourse = value
})

watch(currentInstUserId, (value) => {
  if (value && !formState.salesperson) {
    formState.salesperson = value
  }
}, { immediate: true })

watch(
  () => props.open,
  async (value) => {
    resetDrawerState()
    if (!value)
      return
    hydrateFormStateFromRecord()
    await Promise.all([
      loadOwedSummary(),
      loadOrderLabelOptions(),
    ])
  },
)

const items = computed(() => [
  {
    title: '填写退课单',
    icon: h('img', {
      src: 'https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12181/static/transfer-pre.b79033da.svg',
      style: { width: '32px', height: '32px', marginTop: '-3px', marginRight: '-8px' },
    }),
  },
  {
    title: '确认实退金额',
    icon: h('img', {
      src: current.value === 1
        ? 'https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12181/static/transfer-next.4b9c0674.svg'
        : 'https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12181/static/transfer-next-normal.11eb34d7.svg',
      style: { width: '32px', height: '32px', marginTop: '-3px', marginRight: '-6px' },
    }),
  },
])

function orderLabelFilterOption(input, option) {
  const keyword = input.toLowerCase()
  return String(option?.name || '').toLowerCase().includes(keyword)
}
async function loadOrderLabelOptions() {
  orderLabelLoading.value = true
  try {
    const res = await getOrderTagListPagedApi({
      queryModel: { enable: true },
      sortModel: {},
      pageRequestModel: {
        needTotal: true,
        pageSize: 50,
        pageIndex: 1,
        skipCount: 0,
      },
    })
    if (res.code === 200) {
      orderLabelOptions.value = Array.isArray(res.result?.list) ? res.result.list : []
      return
    }
    orderLabelOptions.value = []
    messageService.error(res.message || '加载订单标签失败')
  }
  catch (error) {
    orderLabelOptions.value = []
    messageService.error(error?.message || '加载订单标签失败')
  }
  finally {
    orderLabelLoading.value = false
  }
}
function handleOrderLabelChange(value) {
  if (Array.isArray(value) && value.length > 5) {
    formState.orderLabel = value.slice(0, 5)
  }
}
function disabledDate(currentDate) {
  return currentDate > dayjs().endOf('day')
}
function handleDateChange(dateObj) {
  if (!dateObj) {
    formState.date = getCurrentDate()
  }
}
function getBase64(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.readAsDataURL(file)
    reader.onload = () => resolve(reader.result)
    reader.onerror = error => reject(error)
  })
}
async function handlePreview(file) {
  if (!file.url && !file.preview) {
    file.preview = await getBase64(file.originFileObj)
  }
  previewImage.value = file.url || file.preview
  previewVisible.value = true
  previewTitle.value = file.name || file.url.substring(file.url.lastIndexOf('/') + 1)
}
function handleCancelPreview() {
  previewVisible.value = false
  previewTitle.value = ''
}
function validateBillImageUpload(file, checkLimit = true) {
  const isImage = file.type === 'image/jpeg' || file.type === 'image/png'
  if (!isImage) {
    messageService.error('只能上传 JPG / JPEG / PNG 格式的图片')
    return false
  }
  const isLt4M = file.size / 1024 / 1024 < 4
  if (!isLt4M) {
    messageService.error('图片大小不能超过 4 MB')
    return false
  }
  if (checkLimit && formState.fileList.length >= 3) {
    messageService.warning('最多上传 3 张图片')
    return false
  }
  return true
}
function beforeBillImageUpload(file) {
  return validateBillImageUpload(file, true)
}
function handleBillImageUpload(options) {
  const { file, onSuccess, onError, onProgress } = options
  const rawFile = file.originFileObj || file
  if (!validateBillImageUpload(rawFile, false)) {
    onError?.(new Error('文件校验未通过'))
    return
  }

  ;(async () => {
    try {
      const tokenRes = await getQiniuToken()
      const { token, uuid, buckethostname } = tokenRes.result || {}
      if (!token || !uuid || !buckethostname)
        throw new Error('上传凭证无效')

      const ext = rawFile.name?.includes('.')
        ? rawFile.name.substring(rawFile.name.lastIndexOf('.'))
        : (rawFile.type === 'image/png' ? '.png' : '.jpg')
      const key = `finance/refund-tuition/${uuid}${ext}`
      const observable = qiniu.upload(rawFile, key, token, {
        fname: rawFile.name,
        mimeType: rawFile.type,
      }, {
        useCdnDomain: true,
        region: qiniu.region.z0,
      })

      observable.subscribe({
        next(res) {
          onProgress?.({ percent: Math.floor(res.total.percent) })
        },
        error(error) {
          console.error('上传退课账单备注图片失败:', error)
          messageService.error(error?.message || '图片上传失败')
          onError?.(error)
        },
        complete(res) {
          onSuccess?.({ url: `${buckethostname}${res.key}` }, file)
        },
      })
    }
    catch (error) {
      console.error('获取上传凭证失败:', error)
      messageService.error(error?.message || '获取上传凭证失败')
      onError?.(error)
    }
  })()
}
function roundAmount(value) {
  return Number(Number(value || 0).toFixed(2))
}
function almostEqual(a, b) {
  return Math.abs(Number(a || 0) - Number(b || 0)) < 0.01
}
function formatMoney(value) {
  return Number(value || 0).toFixed(2)
}
function formatTrimMoney(value) {
  const num = Number(value || 0)
  if (Number.isInteger(num))
    return String(num)
  return num.toFixed(2)
}
function formatCount(value) {
  const num = Number(value || 0)
  if (Number.isInteger(num))
    return String(num)
  return num.toFixed(2)
}
function isEmptyAmountValue(value) {
  return value === '' || value === null || value === undefined
}
function formatDate(value) {
  if (!value || `${value}`.startsWith('0001-01-01'))
    return '-'
  const parsed = dayjs(value)
  if (!parsed.isValid())
    return '-'
  return parsed.format('YYYY-MM-DD')
}
function resetDrawerState() {
  current.value = 0
  openModal.value = false
  calcPreviewOpen.value = false
  feeCalcOpen.value = false
  previewVisible.value = false
  previewTitle.value = ''
  previewImage.value = ''
  nextLoading.value = false
  estimateLoading.value = false
  submitLoading.value = false
  pendingRefundOrderId.value = ''
  pendingRefundNeedPay.value = false
  readonly.value = true
  owedSummary.value = createDefaultOwedSummary()
  estimateResult.value = createDefaultEstimate()
  calcResult.value = createDefaultCalcResult()
  Object.assign(formState, initFormState())
  formRef.value?.resetFields?.()
}
function hydrateFormStateFromRecord() {
  formState.surplusClass = maxRefundQuantity.value
  formState.effectiveDate = validityValueText.value
}
async function loadOwedSummary() {
  if (!tuitionAccountId.value)
    return false
  try {
    const res = await getTuitionAccountRefundOwedSummaryApi({ tuitionAccountId: tuitionAccountId.value })
    if (res.code !== 200) {
      throw new Error(res.message || '加载退课信息失败')
    }
    owedSummary.value = {
      ...createDefaultOwedSummary(),
      ...(res.result || {}),
    }
    return true
  }
  catch (error) {
    owedSummary.value = createDefaultOwedSummary()
    messageService.error(error?.message || '加载退课信息失败')
    return false
  }
}
function handleRefundQuantityChange(value) {
  if (value === null || value === undefined || value === '') {
    formState.dropTheClassNumber = ''
    estimateResult.value = createDefaultEstimate()
    formState.price = '0.00'
    return
  }
  let nextValue = Number(value)
  if (Number.isNaN(nextValue)) {
    formState.dropTheClassNumber = ''
    estimateResult.value = createDefaultEstimate()
    formState.price = '0.00'
    return
  }
  if (nextValue < 0)
    nextValue = 0
  if (!isAmountMode.value)
    nextValue = Math.floor(nextValue)
  else
    nextValue = roundAmount(nextValue)
  if (nextValue > maxRefundQuantity.value)
    nextValue = maxRefundQuantity.value
  formState.dropTheClassNumber = nextValue === 0 ? '' : nextValue
  estimateResult.value = createDefaultEstimate()
  formState.price = '0.00'
}
function validateRefundQuantity(_rule, value) {
  const errorMessage = getRefundQuantityErrorMessage(value)
  if (errorMessage)
    return Promise.reject(new Error(errorMessage))
  return Promise.resolve()
}
function getRefundQuantityErrorMessage(value) {
  if (value === '' || value === null || value === undefined) {
    return '请输入退课数量'
  }
  const amount = Number(value)
  if (Number.isNaN(amount) || amount <= 0) {
    return '请输入退课数量'
  }
  if (amount > maxRefundQuantity.value + 0.009) {
    return '退课数量不能超过当前可退范围'
  }
  return ''
}
async function estimateRefundPreview() {
  const quantity = Number(formState.dropTheClassNumber || 0)
  if (!tuitionAccountId.value || quantity <= 0) {
    estimateResult.value = createDefaultEstimate()
    return
  }
  estimateLoading.value = true
  try {
    const res = await estimateRefundTuitionAccountValuableTuitionApi({
      tuitionAccountId: tuitionAccountId.value,
      quantity,
    })
    if (res.code !== 200) {
      throw new Error(res.message || '退课试算失败')
    }
    estimateResult.value = {
      ...createDefaultEstimate(),
      ...(res.result || {}),
      subAccounts: Array.isArray(res.result?.subAccounts) ? res.result.subAccounts : [],
    }
    formState.price = formatMoney(estimateResult.value.tuition)
  }
  catch (error) {
    estimateResult.value = createDefaultEstimate()
    formState.price = '0.00'
    messageService.error(error?.message || '退课试算失败')
  }
  finally {
    estimateLoading.value = false
  }
}
function handleEstimateBlur() {
  if (getRefundQuantityErrorMessage(formState.dropTheClassNumber))
    return
  estimateRefundPreview()
}
function handleAllReturn() {
  formState.dropTheClassNumber = maxRefundQuantity.value
  handleRefundQuantityChange(maxRefundQuantity.value)
  estimateRefundPreview()
}
async function calculateHandlingFeePreview() {
  if (!tuitionAccountId.value) {
    messageService.error('缺少学费账户ID')
    return false
  }
  try {
    const res = await calculateRefundTuitionAccountHandlingFeeApi({
      tuitionAccountId: tuitionAccountId.value,
      refundQuantity: Number(formState.dropTheClassNumber || 0),
    })
    if (res.code !== 200) {
      throw new Error(res.message || '计算实退金额失败')
    }
    calcResult.value = {
      ...createDefaultCalcResult(),
      ...(res.result || {}),
      details: Array.isArray(res.result?.details) ? res.result.details : [],
    }
    syncCashierRefundAmount()
    return true
  }
  catch (error) {
    messageService.error(error?.message || '计算实退金额失败')
    return false
  }
}
async function handleNext() {
  try {
    await formRef.value?.validate?.()
  }
  catch {
    return
  }
  nextLoading.value = true
  try {
    const success = await calculateHandlingFeePreview()
    if (!success)
      return
    if (shouldShowCalcPreviewModal.value) {
      calcPreviewOpen.value = true
      return
    }
    current.value++
  }
  finally {
    nextLoading.value = false
  }
}
function syncCashierRefundAmount() {
  const refundAmount = roundAmount(Number(calcResult.value?.refundAmount || 0))
  formState.price = formatMoney(refundAmount)
  formState.dropPayPrice = cashierDefaultActualRefundAmount.value
  readonly.value = true
}
function handleConfirm() {
  openModal.value = true
}
function handleOpenFeeCalcModal() {
  feeCalcOpen.value = true
}
function handleCalcPreviewNext() {
  calcPreviewOpen.value = false
  syncCashierRefundAmount()
  current.value++
}
function getVoucherImages() {
  return (Array.isArray(formState.fileList) ? formState.fileList : [])
    .filter(file => file?.status === 'done' || file?.url || file?.response?.url)
    .map(file => file?.url || file?.response?.url || file?.response?.result?.url || file?.response?.result)
    .filter(item => typeof item === 'string' && item.trim())
    .slice(0, 3)
}
async function handleSubmitRefund() {
  if (submitLoading.value)
    return
  const realAmount = Number(formState.dropPayPrice || 0)
  if (isEmptyAmountValue(formState.dropPayPrice) || Number.isNaN(realAmount) || realAmount < 0) {
    readonly.value = false
    messageService.error('请输入实退金额')
    return
  }
  if (!tuitionAccountId.value) {
    messageService.error('缺少学费账户ID')
    return
  }
  if (formState.fileList.some(file => file?.status === 'uploading')) {
    messageService.warning('图片上传中，请稍后再提交')
    return
  }
  if (formState.fileList.some(file => file?.status === 'error' || (!file?.url && !file?.response?.url))) {
    messageService.warning('账单备注图片未上传成功，请删除后重新上传')
    return
  }
  submitLoading.value = true
  try {
    let orderId = pendingRefundOrderId.value
    let needPay = pendingRefundNeedPay.value
    if (!orderId) {
      const owedLoaded = await loadOwedSummary()
      if (!owedLoaded)
        return
      const createRes = await createRefundTuitionAccountOrderApi({
        tuitionAccountId: tuitionAccountId.value,
        totalAmount: roundAmount(cashierRefundAmount.value),
        realAmount: roundAmount(realAmount),
        chargeAgainstTuition: roundAmount(previewArrearDeductionAmount.value),
        refundQuantity: Number(formState.dropTheClassNumber || 0),
        refundFreeQuantity: Number(calcResult.value?.giftRefundQuantity || 0),
        isRechargeAccount: false,
        rechargeAccountId: '0',
        dealDate: formState.date || getCurrentDate(),
        remark: formState.remarks1 || '',
        externalRemark: formState.remarks2 || '',
        salePersonId: formState.salesperson ? String(formState.salesperson) : '0',
        collectorStaffId: '0',
        phoneSellStaffId: '0',
        foregroundStaffId: '0',
        viceSellStaffStaffId: '0',
        orderTagIds: Array.isArray(formState.orderLabel) ? formState.orderLabel.map(String) : [],
        autoCloseTuition: !!formState.autoFinishCourse,
        isOriginalRefund: !!formState.originalPriceRefund,
      })
      if (createRes.code !== 200) {
        throw new Error(createRes.message || '创建退课订单失败')
      }
      orderId = String(createRes.result?.id || '').trim()
      if (!orderId) {
        throw new Error('创建退课订单失败')
      }
      needPay = !!createRes.result?.isNeedPay
      pendingRefundOrderId.value = orderId
      pendingRefundNeedPay.value = needPay
    }
    if (needPay) {
      const payRes = await payRefundTuitionAccountOrderApi({
        orderId,
        payAmount: roundAmount(realAmount),
        isOriginalRefund: !!formState.originalPriceRefund,
        payAccounts: [
          {
            payMethod: Number(formState.payType || 1),
            amount: roundAmount(realAmount),
            accountId: String(formState.dropPayAccount || '1'),
            paymentVoucher: {
              text: formState.billRemarks || '',
              images: getVoucherImages(),
            },
            payTime: formState.date || getCurrentDate(),
          },
        ],
      })
      if (payRes.code !== 200) {
        throw new Error(payRes.message || '退课付款失败')
      }
    }
    messageService.success('退课成功')
    openModal.value = false
    resetDrawerState()
    openDrawer.value = false
    emit('submitted', { orderId })
    emitter.emit(EVENTS.REFRESH_STUDENT_ORDER_RECORD)
  }
  catch (error) {
    messageService.error(error?.message || '退课失败')
  }
  finally {
    submitLoading.value = false
  }
}
function handleBack() {
  current.value--
}
function handleCancel() {
  resetDrawerState()
  openDrawer.value = false
}
const checkOptions = reactive([
  {
    id: 1,
    label: '微信',
    img: 'https://pcsys.admin.ybc365.com/e068b5e2-e27e-4228-8437-fae315326ced.png',
  },
  {
    id: 2,
    label: '支付宝',
    img: 'https://pcsys.admin.ybc365.com/3f396285-ac3a-43fe-94a8-adfef395d47d.png',
  },
  {
    id: 3,
    label: '银行转帐',
    img: 'https://pcsys.admin.ybc365.com/b05c2b92-6c09-46a0-8123-42446727d480.png',
  },
  {
    id: 4,
    label: 'POS机',
    img: 'https://pcsys.admin.ybc365.com/2a6137fb-3e35-470f-b368-2fe71e5439f1.png',
  },
  {
    id: 5,
    label: '现金',
    img: 'https://pcsys.admin.ybc365.com/bab59869-17ea-42c9-9354-d88a59ecf18a.png',
  },
  {
    id: 6,
    label: '其他',
    img: 'https://pcsys.admin.ybc365.com/6d6faf00-aaea-4f6c-b3ab-e581a9bcf7f6.png',
  },
  {
    id: 7,
    label: '储值账户',
    img: 'https://pcsys.admin.ybc365.com/bca6bad7-3180-4b01-b766-4b05013347f7.png',
  },

])
function handleModify() {
  readonly.value = false
}
</script>

<template>
  <div>
    <a-drawer
      v-model:open="openDrawer" :push="{ distance: 80 }" :body-style="{ padding: '0', background: '#f7f7fd' }"
      :closable="false" width="1165px" placement="right" :mask-closable="false" :keyboard="false"
    >
      <!-- 自定义头部 -->
      <template #title>
        <div class="custom-header flex justify-between h-4 flex-items-center">
          <div class="text-5">
            退课
          </div>
          <a-button type="text" class="close-btn" @click="handleCancel">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>
      <a-alert
        message="退课后，将会扣除退课所填写的课时并生成退费订单，请谨慎操作！" show-icon type="warning"
        class="text-#f90 border-none bg-#fff5e6"
      />
      <div class="contenter">
        <div class="steps mt-24px mx-24px bg-#fff rounded-16px flex flex-center px-28% py-24px">
          <a-steps :current="current" :items="items" />
        </div>
        <a-form ref="formRef" :model="formState">
          <div v-if="current === 0" class="refund-basic-card mt-24px mx-24px p-24px bg-#fff rounded-16px flex flex-col py-24px">
            <h1 class="text-20px">
              {{ courseName }}
            </h1>
            <a-space class="flex-1 flex flex-items-center">
              <span v-if="lessonTypeText" class="bg-#e6f0ff text-#06f text-3 px3 py2px rounded-10 ">{{ lessonTypeText }}</span>
              <span v-if="chargingModeText" class="bg-#e6f0ff text-#06f text-3 px3 py2px rounded-10">{{ chargingModeText }}</span>
            </a-space>
            <div class="flex flex-items-center my-24px">
              <span>{{ remainLabelText }}：</span>
              <span class="text-#888">{{ remainValueText }}</span>
            </div>
            <div class="flex flex-items-center">
              <span>{{ validityLabelText }}：</span>
              <span class="text-#888">{{ validityValueText }}</span>
            </div>
            <!-- 分割线 -->
            <a-divider />
            <!-- 退课数量 -->
            <a-form-item label="退课数量" name="dropTheClassNumber" required class="!mb-12px" :rules="[{ validator: validateRefundQuantity, trigger: ['blur', 'change'] }]">
              <div class="flex flex-items-center">
                <a-input-number
                  v-model:value="formState.dropTheClassNumber" placeholder="请输入退课数量" :min="0" :max="maxRefundQuantity"
                  :precision="isAmountMode ? 2 : 0" class="w-160px" @change="handleRefundQuantityChange" @blur="handleEstimateBlur"
                />
                <a-button type="link" class="text-#06f" @click="handleAllReturn">
                  全部退还
                </a-button>
              </div>
            </a-form-item>
            <a-form-item v-if="isFullRefund" label="是否自动结课" name="autoFinishCourse" class="switch-form-item">
              <div class="refund-switch-inline">
                <a-switch v-model:checked="formState.autoFinishCourse" />
                <span class="refund-switch-inline__desc">开启后，全部退课成功后，本课程将自动结课</span>
              </div>
            </a-form-item>
            <a-form-item label="是否原价退课" name="originalPriceRefund" class="switch-form-item !mb-0">
              <div class="refund-switch-inline">
                <a-switch v-model:checked="formState.originalPriceRefund" />
                <span class="refund-switch-inline__desc">开启后，会根据课程报价单的原价计算学员应退金额，其余学费金额计入本次退课应收手续费</span>
              </div>
            </a-form-item>
          </div>
          <div
            v-if="current === 1"
            class="mt-24px mx-24px p-24px pb-0 mb-24px bg-#fff rounded-16px flex flex-col   py-24px"
          >
            <!-- 应退金额 -->
            <div class="flex flex-col">
              <span class="text-#666 mb-2px">应退金额：</span>
              <span class="text-#000 text-48px custom-num-font-family">¥ {{ formatMoney(cashierRefundAmount) }} <span
                class="text-#888 text-14px "
              >{{ confirmSummaryText }}</span> </span>
            </div>
            <!-- 退款方式 -->
            <a-form-item name="payType">
              <div class="payList mt-4 text-#666">
                <div>
                  <span class=" mr-2px">*</span>退款方式
                  <span class="payList-tip ml-2">请选择</span>
                </div>
                <div class="pay">
                  <a-radio-group v-model:value="formState.payType" class="custom-radio">
                    <a-space :size="16" class="flex-wrap">
                      <label
                        v-for="(item, index) in checkOptions" :key="index"
                        class="pay-box" :class="{ active: formState.payType === item.id }"
                      >
                        <span> <img :src="item.img" alt=""> {{ item.label }}</span>
                        <a-radio :value="item.id" />
                      </label>
                    </a-space>
                  </a-radio-group>
                </div>
              </div>
            </a-form-item>
            <a-form-item name="dropPayPrice">
              <!-- 实退金额 -->
              <div class="flex flex-col mt-20px mb-40px">
                <span class="text-#666 mb-2px"><span class="text-red mr-2px">*</span>实退金额：</span>
                <div class="text-#000 text-48px custom-num-font-family">
                  <div
                    class="payPrice h-20 border border-b-#eee border-solid border-x-none border-t-none w-100%"
                    :class="{ 'animate-border': isEmptyAmountValue(formState.dropPayPrice) }"
                  >
                    <a-input-number
                      v-model:value="formState.dropPayPrice" :precision="2"
                      :bordered="false" :controls="false" :readonly="readonly" class="h-100% w-100% text-12"
                      :min="0" :max="100000" placeholder="输入实退金额" @blur="!isEmptyAmountValue(formState.dropPayPrice) ? readonly = true : readonly = false"
                    >
                      <template #addonBefore>
                        <span class="text-12">¥</span>
                      </template>
                      <template #addonAfter>
                        <a-button v-if="readonly" @click="handleModify">
                          修改
                        </a-button>
                      </template>
                    </a-input-number>
                    <span v-if="isEmptyAmountValue(formState.dropPayPrice)" class="text-3.5 text-#f33 relative top--27px">请输入实退金额</span>
                    <span
                      v-else
                      class="text-3.5 text-#888 relative top--27px"
                    >
                      {{ cashierRefundTipText }}
                      <a-popover v-if="showCashierFeeTip" :title="cashierFeeTipTitle" placement="top">
                        <template #content>
                          <div>{{ cashierFeeTipContent }}</div>
                        </template>
                        <QuestionCircleOutlined class="ml-4px text-#1677ff cursor-pointer" />
                      </a-popover>
                    </span>
                  </div>
                </div>
              </div>
            </a-form-item>
            <!-- 退款账户 -->
            <a-form-item name="dropPayAccount" :rules="[{ required: true, message: '请选择退款账户' }]">
              <div class="text-#666 flex flex-items-center mb-6px">
                <span class="text-red mr-2px">*</span>退款账户：
              </div>
              <a-select v-model:value="formState.dropPayAccount" placeholder="请选择退款账户" style="width: 300px;">
                <a-select-option value="1">
                  默认账户
                </a-select-option>
              </a-select>
            </a-form-item>
            <!-- 经办日期 -->
            <a-form-item class="custom-label mt--2">
              <div class="text-#666 flex flex-items-center mb-6px">
                经办日期：
              </div>
              <a-date-picker
                v-model:value="formState.date" class="w-300px" :disabled-date="disabledDate"
                value-format="YYYY-MM-DD" format="YYYY-MM-DD" @change="handleDateChange"
              />
            </a-form-item>
            <!-- 订单标签 -->
            <a-form-item>
              <div class="text-#666 flex flex-items-center mb-6px">
                订单标签：
              </div>
              <a-select
                v-model:value="formState.orderLabel" mode="multiple" placeholder="请选择订单标签" show-search
                class="multiple-select" style="width: 100%" :options="orderLabelOptions"
                :loading="orderLabelLoading" :filter-option="orderLabelFilterOption"
                :field-names="{ label: 'name', value: 'id' }" @change="handleOrderLabelChange"
              />
            </a-form-item>
            <!-- 订单销售员 -->
            <a-form-item>
              <div class="text-#666 flex flex-items-center mb-6px">
                订单销售员：
              </div>
              <StaffSelect v-model="formState.salesperson" placeholder="请选择销售员" width="320px" :status="0" />
            </a-form-item>
            <a-form-item>
              <div class="text-#666 flex flex-items-center mb-6px">
                对内备注：
              </div>
              <a-input
                v-model:value="formState.remarks1" placeholder="请输入对内备注，此备注仅内部员工可见"
                style="width: 100%"
              />
            </a-form-item>
            <a-form-item>
              <div class="text-#666 flex flex-items-center mb-6px">
                对外备注：
              </div>
              <a-input v-model:value="formState.remarks2" placeholder="请输入对内备注，此备注打印时将显示" style="width: 100%" />
            </a-form-item>
            <a-form-item>
              <div class="text-#666 flex flex-items-center mb-6px">
                账单备注：
              </div>
              <a-textarea
                v-model:value="formState.billRemarks" placeholder="请输入内容，最多100字"
                :auto-size="{ minRows: 2, maxRows: 5 }"
              />
            </a-form-item>
            <a-form-item class="w-80%">
              <a-form-item-rest>
                <div class="mt--10px">
                  <a-upload
                    v-model:file-list="formState.fileList"
                    list-type="picture-card"
                    :custom-request="handleBillImageUpload"
                    :before-upload="beforeBillImageUpload"
                    accept=".jpg,.jpeg,.png"
                    @preview="handlePreview"
                  >
                    <div v-if="formState.fileList.length < 3">
                      <PlusOutlined class="text-20px" />
                    </div>
                  </a-upload>
                  <span class="text-#888 text-12px">最多上传3张，支持JPG、JPEG、PNG，单张图片不超过 4 MB</span>
                  <a-modal :open="previewVisible" :title="previewTitle" :footer="null" @cancel="handleCancelPreview">
                    <img alt="example" style="width: 100%" :src="previewImage">
                  </a-modal>
                </div>
              </a-form-item-rest>
            </a-form-item>
          </div>
        </a-form>
      </div>
      <template #footer>
        <div v-if="current === 0" class="px-24px py-16px flex  justify-end flex-items-center">
          <div class="l flex flex-col items-end">
            <span class="text-16px text-#333 font-bold mb-8px">总计学费：¥ {{ formatMoney(estimateAmount) }}</span>
            <span class="text-14px text-#888">{{ footerEstimateSummary }}</span>
          </div>
          <div class="flex flex-items-center ml-24px">
            <a-button type="primary" class="text-20px h-48px w-140px" :loading="nextLoading" @click="handleNext">
              下一步
            </a-button>
          </div>
        </div>
        <div v-if="current === 1" class="px-24px py-16px flex  justify-between flex-items-center">
          <div>
            <a-button type="primary" class="text-20px h-48px w-140px" ghost @click="handleBack">
              返回上一步
            </a-button>
          </div>
          <div class="flex flex-items-center">
            <div class="l flex flex-col items-end">
              <span class="text-16px text-#333 font-bold mb-8px">实退金额：¥ {{ formatMoney(formState.dropPayPrice || 0) }}</span>
              <span class="text-14px text-#888">应退金额：¥ {{ formatMoney(cashierRefundAmount) }}</span>
            </div>
            <div class="flex flex-items-center ml-24px">
              <a-button type="primary" class="text-20px h-48px w-140px" @click="handleConfirm">
                确定
              </a-button>
            </div>
          </div>
        </div>
      </template>
    </a-drawer>
    <a-modal
      v-model:open="calcPreviewOpen"
      :centered='true'
      class="modal-content-box refund-preview-modal"
      :closable="false"
      :width="800"
      :keyboard="false"
      :mask-closable="false"
    >
      <template #title>
        <div class="text-5 flex justify-between flex-center">
          <span class="refund-preview-modal-title">确认应退金额</span>
          <a-button type="text" class="close-btn" @click="calcPreviewOpen = false">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>
      <div class="contenter refund-preview-shell">
        <div class="refund-preview-card">
          <div class="refund-preview-card__header">
            <div>
              <div class="refund-preview-card__title">
                {{ courseName }}
              </div>
              <a-space class="refund-preview-card__tags" :size="10">
                <span v-if="lessonTypeText" class="bg-#e6f0ff text-#06f text-3 px3 py2px rounded-10 ">{{ lessonTypeText }}</span>
                <span v-if="chargingModeText" class="bg-#e6f0ff text-#06f text-3 px3 py2px rounded-10">{{ chargingModeText }}</span>
              </a-space>
            </div>
            <span v-if="previewStatusText" class="refund-preview-card__badge">{{ previewStatusText }}</span>
          </div>
          <div class="refund-preview-card__meta">
            <div v-for="item in previewInfoRows" :key="item.label" class="refund-preview-card__meta-row">
              <span>{{ item.label }}：</span>
              <span>¥ {{ formatMoney(item.value) }}</span>
            </div>
          </div>
          <a-divider class="refund-preview-card__divider" />
          <div class="refund-preview-card__summary refund-preview-card__summary--inline">
            <div class="refund-preview-card__summary-main">
              <div class="refund-preview-card__amount-label">
                <span>应退金额</span>
                <QuestionCircleFilled class="text-#1677ff text-12px" />
              </div>
              <div class="refund-preview-card__amount-inline">
                ¥ {{ formatMoney(previewRefundAmount) }}
              </div>
            </div>
            <div v-if="showPreviewHandlingFee" class="refund-preview-card__fee">
              <span>手续费：</span>
              <span class="refund-preview-card__fee-value">¥ {{ formatMoney(previewHandlingFeeAmount) }}</span>
              <a-button v-if="previewBadDebtAmount > 0.009" type="link" class="refund-preview-card__calc-link" @click="handleOpenFeeCalcModal">
                查看计算过程
              </a-button>
            </div>
          </div>
          <div class="refund-preview-card__desc refund-preview-card__desc--compact">
            {{ previewDescriptionText }}
          </div>
        </div>
      </div>
      <template #footer>
        <a-button type="primary" class="refund-preview-footer__btn" @click="handleCalcPreviewNext">
          下一步
        </a-button>
      </template>
    </a-modal>
    <a-modal
      v-model:open="feeCalcOpen"
      :centered='true'
      class="modal-content-box fee-calc-modal"
      :closable="false"
      :width="1000"
      :keyboard="false"
      :mask-closable="false"
      :footer="null"
    >
      <template #title>
        <div class="text-5 flex justify-between flex-center">
          <span class="refund-preview-modal-title">手续费计算</span>
          <a-button type="text" class="close-btn" @click="feeCalcOpen = false">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>
      <div class="contenter scrollbar fee-calc-content">
        <div class="fee-calc-formula">
          <span class="fee-calc-formula__num">{{ formatMoney(previewHandlingFeeAmount) }}</span>
          <span class="fee-calc-formula__label">（手续费）=</span>
          <span class="fee-calc-formula__num">{{ formatMoney(previewRefundAmount) }}</span>
          <span class="fee-calc-formula__label">（应退金额）</span>
          <span class="fee-calc-formula__symbol">-</span>
          <span class="fee-calc-formula__symbol">(</span>
          <span class="fee-calc-formula__num">{{ formatMoney(previewOriginalRefundAmount) }}</span>
          <span class="fee-calc-formula__label">（原价应退金额）</span>
          <span class="fee-calc-formula__symbol">-</span>
          <span class="fee-calc-formula__num">{{ formatMoney(previewArrearDeductionAmount) }}</span>
          <span class="fee-calc-formula__label">（欠费抵扣金额）</span>
          <span class="fee-calc-formula__symbol">)</span>
        </div>
        <div class="fee-calc-subtitle">
          本次退课涉及以下订单的学费账户，订单欠费¥{{ formatMoney(previewArrearAmount) }}
        </div>
        <div class="fee-calc-table-wrap">
          <a-table
            class="fee-calc-table"
            :columns="feeCalcColumns"
            :data-source="feeCalcTableData"
            :pagination="false"
            :scroll="{ x: 1140 }"
            size="middle"
          />
        </div>
        <div class="fee-calc-rule">
          <div class="fee-calc-rule__title">
            手续费计算规则：
          </div>
          <div>1. 每个学费账户的原价应退金额 =（应收学费金额 - 退转学费金额 - 原价 × 课消课时数量）÷ 剩余课时数 × 退课课时数</div>
          <div>2. 每个学费子账户的原价应退金额进行欠费抵扣以后，计算出本次退课的原价应退总金额</div>
          <div>3. 如果“原价应退总金额”≤0，则手续费 = 应退金额；如果“原价应退总金额”＞0，则手续费 = 应退金额 - 原价应退金额</div>
          <div>4. 优惠价和原价计算会出现小数点精度问题，可能会导致手续费存在误差，可在下一步收银台页面手动调整实退金额和手续费</div>
        </div>
      </div>
    </a-modal>
    <!-- 确定退课提示 -->
    <a-modal
      v-model:open="openModal" centered :mask-closable="false" :keyboard="false" :footer="false" :width="416"
      :closable="false"
    >
      <div class="text-16px font-bold flex flex-items-center">
        <QuestionCircleFilled class="text-#f33 text-22px mr-12px" />确定退课？
      </div>
      <div class="pl-35px text-#666 mt-12px">
        退课后，将会扣除退课所填写的课时并生成退费订单，请谨慎操作！
      </div>
      <div class="flex flex-items-center justify-end mt-24px">
        <a-button @click="openModal = false">
          再想想
        </a-button>
        <a-button danger ghost class="ml-12px" :loading="submitLoading" @click="handleSubmitRefund">
          确定退课
        </a-button>
      </div>
    </a-modal>
  </div>
</template>

<style lang="less" scoped>
/* 添加旋转动画 */
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

.payList {
  span {
    color: red;
  }

  span.payList-tip {
    color: var(--pro-ant-color-primary);
  }
}

.pay {
  margin-top: 10px;

  .pay-box {
    border: 1px solid #eee;
    padding: 12px 16px;
    display: flex;
    align-items: center;
    border-radius: 6px;
    margin-right: 6px;
    user-select: none;
    cursor: pointer;

    &:hover {
      border-color: var(--pro-ant-color-primary);
    }

    span {
      color: #000;
      margin-right: 20px;
      display: flex;
      align-items: center;

      img {
        width: 20px;
        height: 20px;
        margin-right: 6px;
      }
    }
  }

  .active {
    border-color: var(--pro-ant-color-primary);
  }
}

.payPrice {
  :deep(.ant-input-number-input) {
    height: 80px;
    line-height: 80px;
    font-family: "DIN alternate", sans-serif;
    padding-left: 0;
  }

  :deep(.ant-input-number-group-addon) {
    background: transparent;
    border: none;
  }

}

/* 动画关键帧 */
@keyframes borderExpand {
  0% {
    transform: scaleX(0);
    opacity: 0;
  }

  100% {
    transform: scaleX(1);
    opacity: 1;
  }
}

/* 动画容器 */
.animate-border {
  position: relative;
}

/* 动画线条 */
.animate-border::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: #ff3333;
  transform-origin: center;
  /* 缩放中心点 */
  animation: borderExpand 0.4s cubic-bezier(0.68, -0.55, 0.27, 1.2) forwards;
}

.multiple-select {
  :deep(.ant-select-selection-item) {
    background-color: #e6f0ff;
    border: 1px solid #99c2ff;
  }
}

.refund-basic-card {
  :deep(.ant-form-item-row) {
    flex-wrap: nowrap;
  }

  :deep(.ant-form-item-label) {
    flex: 0 0 108px;
    max-width: 108px;
    padding-right: 8px;
  }

  :deep(.ant-form-item-control) {
    flex: 1;
    min-width: 0;
  }
}

.switch-form-item {
  margin-bottom: 12px;

  :deep(.ant-form-item-row) {
    align-items: center;
  }

  :deep(.ant-form-item-control-input) {
    min-height: 32px;
  }

  :deep(.ant-form-item-control-input-content) {
    display: flex;
    align-items: center;
  }
}

.refund-switch-inline {
  display: flex;
  align-items: center;
  gap: 12px;
}

.refund-switch-inline__desc {
  display: inline-flex;
  align-items: center;
  color: #888;
  font-size: 12px;
  line-height: 20px;
  white-space: nowrap;
}

.refund-preview-modal-title {
  color: #1f1f1f;
  font-size: 16px;
  font-weight: 600;
  line-height: 22px;
}

.refund-preview-shell {
  padding: 28px 24px 32px;
}

.refund-preview-card {
  min-height: 278px;
  padding: 24px 24px 20px;
  background: #fafafa;
  border-radius: 14px;
}

.refund-preview-card__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
}

.refund-preview-card__title {
  color: #1f1f1f;
  font-size: 22px;
  font-weight: 700;
  line-height: 30px;
}

.refund-preview-card__tags {
  margin-top: 8px;
}

.refund-preview-card__badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 56px;
  height: 20px;
  padding: 0 12px;
  background: #fff1f0;
  border-radius: 999px;
  color: #ff4d4f;
  font-size: 12px;
  line-height: 20px;
}

.refund-preview-card__meta {
  margin-top: 12px;
}

.refund-preview-card__meta-row {
  display: flex;
  align-items: center;
  color: #8c8c8c;
  font-size: 14px;
  line-height: 28px;
}

.refund-preview-card__divider {
  margin: 22px 0 24px;
}

.refund-preview-card__amount-label {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: #666;
  font-size: 14px;
  line-height: 22px;
}

.refund-preview-card__summary--inline {
  display: flex;
  align-items: center;
  gap: 52px;
  flex-wrap: wrap;
}

.refund-preview-card__summary-main {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.refund-preview-card__amount-inline {
  color: #262626;
  font-family: "DIN alternate", sans-serif;
  font-size: 21px;
  font-weight: 700;
  line-height: 28px;
}

.refund-preview-card__fee {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #666;
  font-size: 14px;
  line-height: 22px;
}

.refund-preview-card__fee-value {
  color: #ff7a00;
  font-family: "DIN alternate", sans-serif;
  font-size: 20px;
  font-weight: 700;
  line-height: 28px;
}

.refund-preview-card__calc-link {
  height: 22px;
  padding: 0 0 0 14px;
  font-size: 14px;
  line-height: 22px;
}

.refund-preview-card__desc {
  margin-top: 14px;
  color: #8c8c8c;
  font-size: 13px;
  line-height: 20px;
}

.refund-preview-card__desc--compact {
  margin-top: 10px;
}

.refund-preview-footer__btn {
  min-width: 84px;
  height: 34px;
  border-radius: 8px;
  font-size: 14px;
}

.fee-calc-content {
  padding: 24px;
  color: #333;
}

.fee-calc-formula {
  display: flex;
  align-items: baseline;
  flex-wrap: wrap;
  gap: 6px;
  color: #2b2f36;
  font-size: 16px;
  line-height: 32px;
}

.fee-calc-formula__num {
  color: #1f1f1f;
  font-size: 22px;
  font-weight: 700;
}

.fee-calc-formula__label {
  color: #666;
  font-size: 13px;
}

.fee-calc-formula__symbol {
  color: #1f1f1f;
  font-size: 22px;
  font-weight: 600;
}

.fee-calc-subtitle {
  margin-top: 14px;
  color: #666;
  font-size: 14px;
  line-height: 22px;
}

.fee-calc-table-wrap {
  margin-top: 14px;
  border: 1px solid #f0f0f0;
  border-radius: 2px;
  background: #fff;
}

.fee-calc-table {
  font-size: 14px;

  :deep(.ant-table) {
    border-radius: 2px;
  }

  :deep(.ant-table-thead > tr > th) {
    height: 46px;
    padding: 0 16px;
    background: #fafafa;
    color: #333;
    font-weight: 600;
    white-space: nowrap;
  }

  :deep(.ant-table-tbody > tr > td) {
    height: 42px;
    padding: 0 16px;
    color: #333;
    white-space: nowrap;
  }

  :deep(.ant-table-tbody > tr:last-child > td) {
    color: #666;
    background: #fff;
  }

  :deep(.ant-table-tbody > tr:last-child > td:last-child) {
    color: #333;
    font-weight: 600;
  }
}

.fee-calc-rule {
  margin-top: 12px;
  color: #666;
  font-size: 13px;
  line-height: 24px;
}

.fee-calc-rule__title {
  margin-bottom: 4px;
  color: #333;
  font-size: 14px;
  font-weight: 600;
}
</style>

<style>
.modal-content-box.refund-preview-modal .ant-modal-content {
  overflow: hidden;
  border-radius: 8px;
}

.modal-content-box.refund-preview-modal .ant-modal-header {
  padding: 12px 16px !important;
  margin-bottom: 0;
  border-bottom: 1px solid #f0f0f0;
}

.modal-content-box.refund-preview-modal .ant-modal-body {
  padding: 0 !important;
}

.modal-content-box.refund-preview-modal .ant-modal-footer {
  padding: 14px 24px 20px;
  border-top: 1px solid #f0f0f0;
}

.modal-content-box.fee-calc-modal .ant-modal-content {
  overflow: hidden;
  border-radius: 8px;
}

.modal-content-box.fee-calc-modal .ant-modal-header {
  padding: 12px 16px !important;
  margin-bottom: 0;
  border-bottom: 1px solid #f0f0f0;
}

.modal-content-box.fee-calc-modal .ant-modal-body {
  padding: 0 !important;
}
</style>
