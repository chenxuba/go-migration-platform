<script setup>
import { CloseOutlined, PlusOutlined, QuestionCircleFilled } from '@ant-design/icons-vue'
import { h } from 'vue'
import dayjs from 'dayjs'
import {
  calculateRefundTuitionAccountHandlingFeeApi,
  estimateRefundTuitionAccountValuableTuitionApi,
  getTuitionAccountRefundOwedSummaryApi,
} from '@/api/edu-center/tuition-account'
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
const emit = defineEmits(['update:open'])

const openDrawer = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

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
    salesperson: undefined,
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
const owedSummary = ref(createDefaultOwedSummary())
const estimateResult = ref(createDefaultEstimate())
const calcResult = ref(createDefaultCalcResult())

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
const shouldShowCalcPreviewModal = computed(() => previewArrearAmount.value > 0.009 || previewBadDebtAmount.value > 0.009)
const showPreviewHandlingFee = computed(() => previewBadDebtAmount.value > 0.009 || previewHandlingFeeAmount.value > 0.009)
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
    { label: '退课金额', value: previewOriginalRefundAmount.value },
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

watch(isFullRefund, (value) => {
  formState.autoFinishCourse = value
})

watch(
  () => props.open,
  async (value) => {
    resetDrawerState()
    if (!value)
      return
    hydrateFormStateFromRecord()
    await loadOwedSummary()
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

const orderLabelOptions = ref([
  { id: '1', name: '内荐' },
  { id: '2', name: '转介绍' },
  { id: '3', name: '双11订单' },
  { id: '4', name: '会员日订单' },
])
const salespersonOptions = ref([
  { id: '1', name: '陈瑞生', phone: '17601241636' },
  { id: '2', name: '刘明', phone: '18876552232' },
  { id: '3', name: '张望名', phone: '17601241636' },
  { id: '4', name: '李元芳', phone: '17601241636' },
])
function orderLabelFilterOption(input, option) {
  const keyword = input.toLowerCase()
  const nameMatch = option.name.toLowerCase().includes(keyword)
  return nameMatch
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
function roundAmount(value) {
  return Number(Number(value || 0).toFixed(2))
}
function almostEqual(a, b) {
  return Math.abs(Number(a || 0) - Number(b || 0)) < 0.01
}
function formatMoney(value) {
  return Number(value || 0).toFixed(2)
}
function formatCount(value) {
  const num = Number(value || 0)
  if (Number.isInteger(num))
    return String(num)
  return num.toFixed(2)
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
    return
  try {
    const res = await getTuitionAccountRefundOwedSummaryApi({ tuitionAccountId: tuitionAccountId.value })
    if (res.code !== 200) {
      throw new Error(res.message || '加载退课信息失败')
    }
    owedSummary.value = {
      ...createDefaultOwedSummary(),
      ...(res.result || {}),
    }
  }
  catch (error) {
    owedSummary.value = createDefaultOwedSummary()
    messageService.error(error?.message || '加载退课信息失败')
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
    formState.price = formatMoney(calcResult.value.refundAmount)
    formState.dropPayPrice = Number(calcResult.value.refundAmount || 0)
    readonly.value = true
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
function handleConfirm() {
  openModal.value = true
}
function handleOpenFeeCalcModal() {
  feeCalcOpen.value = true
}
function handleCalcPreviewNext() {
  calcPreviewOpen.value = false
  current.value++
}
function handleSubmitRefund() {
  messageService.info('退课下单接口待接入')
  openModal.value = false
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
              <span class="text-#000 text-48px custom-num-font-family">¥ {{ formState.price }} <span
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
                    :class="{ 'animate-border': !formState.dropPayPrice }"
                  >
                    <a-input-number
                      v-model:value="formState.dropPayPrice" :precision="2"
                      :bordered="false" :controls="false" :readonly="readonly" class="h-100% w-100% text-12"
                      :min="1" :max="100000" placeholder="输入实退金额" @blur="formState.dropPayPrice ? readonly = true : readonly = false"
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
                    <span v-if="!formState.dropPayPrice" class="text-3.5 text-#f33 relative top--27px">请输入实退金额</span>
                    <span
                      v-if="formState.dropPayPrice && formState.dropPayPrice < formState.price"
                      class="text-3.5 text-#888 relative top--27px"
                    >应退金额：¥ {{ formState.dropPayPrice }}，手续费 ¥
                      {{ (formState.price
                        - formState.dropPayPrice).toFixed(2) }}</span>
                    <span
                      v-if="formState.dropPayPrice && formState.dropPayPrice == formState.price"
                      class="text-3.5 text-#888 relative top--27px"
                    >应退金额：¥ {{ formState.dropPayPrice }}</span>
                    <span
                      v-if="formState.dropPayPrice && formState.dropPayPrice > formState.price"
                      class="text-3.5 text-#888 relative top--27px"
                    >应退金额：¥ {{ formState.dropPayPrice }}，亏损费 ¥
                      {{ (formState.dropPayPrice - formState.price).toFixed(2) }}</span>
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
                :filter-option="orderLabelFilterOption" :field-names="{ label: 'name', value: 'id' }"
              />
            </a-form-item>
            <!-- 订单销售员 -->
            <a-form-item>
              <div class="text-#666 flex flex-items-center mb-6px">
                订单销售员：
              </div>
              <a-select
                v-model:value="formState.salesperson" placeholder="请选择销售员" show-search style="width: 320px"
                :options="salespersonOptions" :field-names="{ label: 'name', value: 'id' }"
              >
                <template #option="{ name, phone }">
                  <div class="flex justify-between flex-items-center">
                    <span>{{ name }}</span>
                    <span class="text-#999 text-3">{{ phone }}</span>
                  </div>
                </template>
              </a-select>
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
                    action="https://www.mocky.io/v2/5cc8019d300000980a055e76" list-type="picture-card"
                    accept=".jpg,.jpeg,.png" @preview="handlePreview"
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
              <span class="text-14px text-#888">应退金额：¥ {{ formatMoney(calcResult.refundAmount || 0) }}</span>
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
          <table class="fee-calc-table">
            <thead>
              <tr>
                <th>订单号</th>
                <th>原价</th>
                <th>优惠后金额</th>
                <th>应收学费金额</th>
                <th>退/转学员金额</th>
                <th>已课消数量</th>
                <th>原价应退金额</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in feeCalcDetails" :key="item.orderNumber || item.sourceId">
                <td>{{ item.orderNumber || '-' }}</td>
                <td>{{ feeCalcUnitPriceText(item.originalUnitPrice) }}</td>
                <td>{{ feeCalcUnitPriceText(item.discountedUnitPrice) }}</td>
                <td>{{ feeCalcMoneyText(item.shouldTuition) }}</td>
                <td>{{ feeCalcMoneyText(item.transferredTuition) }}</td>
                <td>{{ feeCalcQuantityText(item.consumedQuantity) }}</td>
                <td>{{ feeCalcMoneyText(item.originalRefundAmount) }}</td>
              </tr>
              <tr class="fee-calc-table__total-row">
                <td>总计</td>
                <td />
                <td />
                <td />
                <td />
                <td />
                <td>¥ {{ formatMoney(previewOriginalRefundAmount) }}</td>
              </tr>
            </tbody>
          </table>
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
        <a-button danger ghost class="ml-12px" @click="handleSubmitRefund">
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
  overflow-x: auto;
  border: 1px solid #f0f0f0;
  border-radius: 2px;
  background: #fff;
}

.fee-calc-table {
  min-width: 1060px;
  width: 100%;
  border-collapse: collapse;
  table-layout: fixed;
  font-size: 14px;
}

.fee-calc-table th {
  height: 46px;
  padding: 0 16px;
  background: #fafafa;
  color: #333;
  font-weight: 600;
  text-align: left;
  border-bottom: 1px solid #f0f0f0;
}

.fee-calc-table td {
  height: 42px;
  padding: 0 16px;
  color: #333;
  white-space: nowrap;
  border-bottom: 1px solid #f0f0f0;
}

.fee-calc-table__total-row td {
  color: #666;
  background: #fff;
}

.fee-calc-table__total-row td:last-child {
  color: #333;
  font-weight: 600;
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
