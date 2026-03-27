<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import dayjs from 'dayjs'
import { getOrderDetailApi } from '@/api/finance-center/order-manage'
import { useUserStore } from '@/stores/user'
import messageService from '@/utils/messageService'

type ReceiptTemplateMode = 'a4' | 'dot' | 'receipt'

interface ReceiptDetailRow {
  name: string
  quoteLabel: string
  chargingModeText: string
  quantityText: string
  giftText: string
  periodText: string
  originalAmount: number
  discountAmount: number
  amount: number
  noteText: string
}

interface ReceiptKvItem {
  label: string
  value: string
}

function resolveTemplateMode(value: unknown): ReceiptTemplateMode {
  const normalized = String(value || '').toLowerCase()
  if (normalized === 'dot' || normalized === 'receipt')
    return normalized
  return 'a4'
}

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const orderId = computed(() => String(route.query.orderId || route.params.orderId || ''))
const templateMode = ref<ReceiptTemplateMode>(resolveTemplateMode(route.query.template))
const loading = ref(false)
const detail = ref<any>(null)
const hasAutoPrinted = ref(false)
const printedAt = ref(dayjs().format('YYYY-MM-DD HH:mm:ss'))

const orgName = computed(() => userStore.userInfo?.orgName || '总校区')
const printedBy = computed(() => userStore.nickname || '-')
const campusPhone = computed(() => '-')
const campusAddress = computed(() => '-')

const orderTypeMap: Record<number, string> = {
  1: '报名续费',
  2: '储值账户充值',
  3: '退课',
  4: '储值账户退费',
  5: '转课',
  6: '退教材费',
  7: '退学杂费',
}

const orderStatusMap: Record<number, string> = {
  1: '待付款',
  2: '审批中',
  3: '已完成',
  4: '已关闭',
  5: '已作废',
  6: '待处理',
  7: '退费中',
  8: '已退费',
}

const orderSourceMap: Record<number, string> = {
  1: '线下办理',
  2: '微校报名',
  3: '线下导入',
  4: '续费订单',
}

const payMethodMap: Record<number, string> = {
  1: '微信',
  2: '支付宝',
  3: '银行转账',
  4: 'POS机',
  5: '现金',
  6: '其他',
}

const sexMap: Record<number, string> = {
  0: '未知',
  1: '男',
  2: '女',
}

const chargingModeMap: Record<number, string> = {
  1: '按课时',
  2: '按时段',
  3: '按金额',
}

const templateTabs = [
  { key: 'a4', label: 'A4纸模板' },
  { key: 'dot', label: '针式打印模板' },
  { key: 'receipt', label: '小票模板' },
]

function isPlaceholderDate(value?: string | Date | null) {
  const raw = String(value || '').trim()
  if (!raw)
    return true
  return raw.startsWith('0001-01-01') || raw.startsWith('0000-00-00')
}

function formatDateTime(value?: string | Date | null) {
  if (isPlaceholderDate(value))
    return '-'
  const parsed = dayjs(value)
  return parsed.isValid() ? parsed.format('YYYY-MM-DD HH:mm:ss') : '-'
}

function formatDateOnly(value?: string | Date | null) {
  if (isPlaceholderDate(value))
    return '-'
  const parsed = dayjs(value)
  return parsed.isValid() ? parsed.format('YYYY-MM-DD') : '-'
}

function formatMoney(value: unknown) {
  const amount = Number(value || 0)
  return Number.isFinite(amount) ? amount.toFixed(2) : '0.00'
}

function formatCurrency(value: unknown) {
  return `¥${formatMoney(value)}`
}

function formatNegativeCurrency(value: unknown) {
  return `-¥${formatMoney(value)}`
}

function formatQuantity(value: unknown, chargingMode?: number | null) {
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

function formatChineseMoney(value: unknown) {
  const amount = Math.abs(Number(value || 0))
  if (!Number.isFinite(amount) || amount === 0)
    return '零元整'

  const cnNums = ['零', '壹', '贰', '叁', '肆', '伍', '陆', '柒', '捌', '玖']
  const cnIntRadice = ['', '拾', '佰', '仟']
  const cnIntUnits = ['', '万', '亿', '兆']
  const cnDecUnits = ['角', '分']

  const rounded = Math.round(amount * 100) / 100
  const integerNum = Math.floor(rounded)
  const decimalNum = Math.round((rounded - integerNum) * 100)

  let chinese = ''

  if (integerNum > 0) {
    const digits = String(integerNum).split('').reverse()
    let zeroCount = 0
    for (let i = 0; i < digits.length; i++) {
      const digit = Number(digits[i])
      const quotient = Math.floor(i / 4)
      const modulus = i % 4
      if (digit === 0) {
        zeroCount++
      }
      else {
        if (zeroCount > 0) {
          chinese = cnNums[0] + chinese
        }
        zeroCount = 0
        chinese = cnNums[digit] + cnIntRadice[modulus] + chinese
      }
      if (modulus === 0 && zeroCount < 4) {
        chinese = cnIntUnits[quotient] + chinese
      }
    }
    chinese += '元'
  }

  if (decimalNum === 0) {
    chinese += '整'
  }
  else {
    const jiao = Math.floor(decimalNum / 10)
    const fen = decimalNum % 10
    if (jiao > 0) {
      chinese += cnNums[jiao] + cnDecUnits[0]
    }
    if (fen > 0) {
      chinese += cnNums[fen] + cnDecUnits[1]
    }
  }

  return chinese
}

function formatPeriodText(item: any) {
  const start = formatDateOnly(item?.validDate)
  const end = formatDateOnly(item?.endDate)
  if (start !== '-' && end !== '-')
    return `${start} 至 ${end}`
  if (end !== '-')
    return `有效期至 ${end}`
  if (start !== '-')
    return `开始于 ${start}`
  return '不限期'
}

function getPurchaseQuantity(item: any) {
  const realQuantity = Number(item?.realQuantity || 0)
  const freeQuantity = Number(item?.freeQuantity || 0)
  if (realQuantity > 0) {
    return Math.max(realQuantity - freeQuantity, 0)
  }
  return Number(item?.count || 0)
}

const isRefundOrder = computed(() => [3, 4, 6, 7].includes(Number(detail.value?.orderType || 0)))
const orderTypeText = computed(() => orderTypeMap[Number(detail.value?.orderType || 0)] || '-')
const orderStatusText = computed(() => orderStatusMap[Number(detail.value?.orderStatus || 0)] || '-')
const orderSourceText = computed(() => orderSourceMap[Number(detail.value?.orderSource || 0)] || '-')
const orderTagText = computed(() =>
  Array.isArray(detail.value?.orderTagNames) && detail.value.orderTagNames.length
    ? detail.value.orderTagNames.join('、')
    : '-',
)
const businessContentText = computed(() =>
  Array.isArray(detail.value?.productItems) && detail.value.productItems.length
    ? detail.value.productItems.join('、')
    : '-',
)
const grandTotalAmount = computed(() => Number(detail.value?.totalAmount ?? detail.value?.amount ?? 0))
const orderDiscountAmount = computed(() => Number(detail.value?.orderDiscountAmount || 0))
const externalPaidAmount = computed(() => Number(detail.value?.paidAmount || 0))
const arrearAmount = computed(() => Number(detail.value?.arrearAmount || 0))
const amountInWords = computed(() => formatChineseMoney(grandTotalAmount.value))

const rechargeDeductionItems = computed(() => {
  if (!detail.value)
    return []
  return [
    { label: '充值余额抵扣', value: Number(detail.value.rechargeAccountAmount || 0) },
    { label: '残联余额抵扣', value: Number(detail.value.rechargeAccountResidualAmount || 0) },
    { label: '赠送余额抵扣', value: Number(detail.value.rechargeAccountGivingAmount || 0) },
  ].filter(item => item.value > 0)
})

const storageDeductionTotal = computed(() =>
  rechargeDeductionItems.value.reduce((sum, item) => sum + Number(item.value || 0), 0),
)

const paymentRecords = computed(() =>
  Array.isArray(detail.value?.paymentRecords)
    ? detail.value.paymentRecords.filter((item: any) => Number(item?.payAmount || 0) > 0)
    : [],
)

const paymentSummaryText = computed(() => {
  if (!paymentRecords.value.length) {
    if (storageDeductionTotal.value > 0)
      return '本单无外部支付，已通过储值账户抵扣完成结算'
    return '暂无支付明细'
  }
  return paymentRecords.value
    .map((item: any) => `${payMethodMap[Number(item?.payMethod || 0)] || '其他'} ${formatCurrency(item?.payAmount)}`)
    .join('；')
})

const printSettingTip = computed(() => {
  if (templateMode.value === 'dot') {
    return '针式打印前，请在浏览器打印预览中将“边距”设为“无”，并关闭页眉页脚。'
  }
  if (templateMode.value === 'receipt') {
    return '小票打印前，建议在打印预览中将“边距”设为“无”，并关闭页眉页脚。'
  }
  return '打印前建议在打印预览中将“边距”设为“无”，并关闭页眉页脚。'
})

const printPreviewSizeTip = computed(() => {
  if (templateMode.value === 'dot') {
    return '建议预览尺寸：自定义 210mm × 150mm'
  }
  if (templateMode.value === 'receipt') {
    return '建议预览尺寸：自定义 80mm × 200mm'
  }
  return '建议预览尺寸：A4 210mm × 297mm'
})

const studentMetaList = computed<ReceiptKvItem[]>(() => [
  { label: '学员姓名', value: detail.value?.studentName || '-' },
  { label: '手机号', value: detail.value?.studentPhone || '-' },
  { label: '性别', value: sexMap[Number(detail.value?.sex || 0)] || '未知' },
  { label: '订单标签', value: orderTagText.value },
])

const businessMetaList = computed<ReceiptKvItem[]>(() => [
  { label: '票据编号', value: detail.value?.orderNumber || '-' },
  { label: '订单类型', value: orderTypeText.value },
  { label: '订单状态', value: orderStatusText.value },
  { label: '办理内容', value: businessContentText.value },
  { label: '订单来源', value: orderSourceText.value },
  { label: '经办日期', value: formatDateOnly(detail.value?.dealDate) },
  { label: '创建时间', value: formatDateTime(detail.value?.createdTime) },
  { label: '完成时间', value: formatDateTime(detail.value?.finishedTime) },
])

const operatorMetaList = computed<ReceiptKvItem[]>(() => [
  { label: '经办人', value: detail.value?.staffName || '-' },
  { label: '订单销售员', value: detail.value?.salePersonName || '-' },
  { label: '打印人', value: printedBy.value },
  { label: '打印时间', value: printedAt.value },
])

const settlementMetaList = computed<ReceiptKvItem[]>(() => [
  { label: isRefundOrder.value ? '业务金额' : '订单应收', value: formatCurrency(grandTotalAmount.value) },
  { label: '整单优惠', value: orderDiscountAmount.value > 0 ? formatNegativeCurrency(orderDiscountAmount.value) : '¥0.00' },
  { label: '储值抵扣', value: storageDeductionTotal.value > 0 ? formatNegativeCurrency(storageDeductionTotal.value) : '¥0.00' },
  { label: '外部实收', value: formatCurrency(externalPaidAmount.value) },
  { label: '待收欠费', value: arrearAmount.value > 0 ? formatCurrency(arrearAmount.value) : '¥0.00' },
  { label: '支付摘要', value: paymentSummaryText.value },
])

const noteMetaList = computed<ReceiptKvItem[]>(() => [
  { label: '对外备注', value: detail.value?.externalRemark || '-' },
  { label: '票据说明', value: '此票据为系统业务凭证，用于学员留存与内部对账。' },
])

const paymentTableRows = computed(() =>
  paymentRecords.value.map((item: any) => ({
    methodText: payMethodMap[Number(item?.payMethod || 0)] || '其他',
    accountText: item?.accountName || '-',
    paidAtText: formatDateTime(item?.payTime || item?.createdTime),
    amountText: formatCurrency(item?.payAmount),
    remarkText: item?.remark || '-',
  })),
)

const receiptRows = computed<ReceiptDetailRow[]>(() => {
  if (!detail.value)
    return []

  const orderType = Number(detail.value.orderType || 0)
  if ([2, 4].includes(orderType)) {
    return [
      { name: '充值金额', value: Number(detail.value.rechargeAccountAmount || 0) },
      { name: '残联金额', value: Number(detail.value.rechargeAccountResidualAmount || 0) },
      { name: '赠送金额', value: Number(detail.value.rechargeAccountGivingAmount || 0) },
    ]
      .filter(item => item.value > 0)
      .map(item => ({
        name: item.name,
        quoteLabel: orderType === 4 ? '储值账户退费' : '储值账户充值',
        chargingModeText: '账户金额',
        quantityText: '1笔',
        giftText: '-',
        periodText: '即时生效',
        originalAmount: item.value,
        discountAmount: 0,
        amount: item.value,
        noteText: `业务类型：${orderTypeText.value}`,
      }))
  }

  const orderItems = Array.isArray(detail.value.orderItems) ? detail.value.orderItems : []
  return orderItems.map((item: any) => {
    const chargingMode = Number(item?.chargingMode || 0)
    const purchasedQuantity = getPurchaseQuantity(item)
    const discountAmount = Math.max(Number(item?.amount || 0) - Number(item?.receivableAmount || 0), 0)
    const noteParts = [
      item?.quoteName ? `报价单：${item.quoteName}` : '',
      chargingModeMap[chargingMode] ? `收费方式：${chargingModeMap[chargingMode]}` : '',
      Number(item?.quotePrice || 0) > 0 ? `报价：${formatCurrency(item.quotePrice)}` : '',
    ].filter(Boolean)

    return {
      name: item?.courseName || '-',
      quoteLabel: item?.quoteName || '-',
      chargingModeText: chargingModeMap[chargingMode] || '-',
      quantityText: formatQuantity(purchasedQuantity, chargingMode),
      giftText: Number(item?.freeQuantity || 0) > 0 ? formatQuantity(item.freeQuantity, chargingMode) : '-',
      periodText: formatPeriodText(item),
      originalAmount: Number(item?.amount || item?.receivableAmount || 0),
      discountAmount,
      amount: Number(item?.receivableAmount || item?.amount || 0),
      noteText: noteParts.join('｜') || '-',
    }
  })
})

async function loadOrderDetail() {
  if (!orderId.value) {
    detail.value = null
    return
  }
  loading.value = true
  try {
    const { result } = await getOrderDetailApi({ orderId: orderId.value })
    detail.value = result || null
  }
  catch (error) {
    console.error('load order receipt detail failed', error)
    messageService.error('加载订单详情失败')
  }
  finally {
    loading.value = false
  }
}

function switchTemplate(key: ReceiptTemplateMode) {
  templateMode.value = key
  router.replace({
    path: route.path,
    query: {
      ...route.query,
      template: key,
    },
  })
}

function printPage() {
  window.print()
}

function downloadPage() {
  window.print()
}

watch(orderId, () => {
  hasAutoPrinted.value = false
  printedAt.value = dayjs().format('YYYY-MM-DD HH:mm:ss')
  loadOrderDetail()
}, { immediate: true })

watch(() => route.query.template, (value) => {
  templateMode.value = resolveTemplateMode(value)
})

watch(
  [loading, detail, () => route.query.autoPrint],
  async ([currentLoading, currentDetail, autoPrint]) => {
    if (hasAutoPrinted.value || currentLoading || !currentDetail || autoPrint !== '1')
      return
    hasAutoPrinted.value = true
    await nextTick()
    setTimeout(() => window.print(), 200)
  },
  { immediate: true },
)
</script>

<template>
  <div class="receipt-print-page">
    <div class="receipt-toolbar print-hidden">
      <div class="receipt-toolbar__left">
        <span class="receipt-toolbar__label">打印模板选择：</span>
        <div class="receipt-toolbar__tabs">
          <button
            v-for="item in templateTabs"
            :key="item.key"
            type="button"
            class="receipt-tab"
            :class="{ 'receipt-tab--active': templateMode === item.key }"
            @click="switchTemplate(item.key as ReceiptTemplateMode)"
          >
            {{ item.label }}
          </button>
        </div>
      </div>
    </div>

    <div class="receipt-print-tip print-hidden">
      <span>{{ printSettingTip }}</span>
      <span class="receipt-print-tip__divider">|</span>
      <span>{{ printPreviewSizeTip }}</span>
    </div>

    <div v-if="loading" class="receipt-loading">
      加载中...
    </div>

    <div v-else-if="!detail" class="receipt-loading">
      暂无订单数据
    </div>

    <div v-else class="receipt-preview-wrap">
      <section v-if="templateMode === 'a4'" class="receipt-paper receipt-paper--a4">
        <header class="receipt-a4-head">
          <div class="receipt-a4-head__org">
            {{ orgName }}
          </div>
          <div class="receipt-a4-head__line" />
          <div class="receipt-a4-head__meta">
            <span>学员姓名：{{ detail.studentName || '-' }}</span>
            <span>订单类型：{{ orderTypeText }}</span>
            <span>订单号：{{ detail.orderNumber || '-' }}</span>
          </div>
        </header>

        <section class="receipt-sheet-section">
          <div class="receipt-sheet-section__title">
            <span class="receipt-sheet-section__bar" />
            校区信息
          </div>
          <table class="receipt-sheet-table">
            <tbody>
              <tr>
                <th>校区名称</th>
                <th>校区电话</th>
                <th>校区地址</th>
              </tr>
              <tr>
                <td>{{ orgName }}</td>
                <td>{{ campusPhone }}</td>
                <td>{{ campusAddress }}</td>
              </tr>
            </tbody>
          </table>
        </section>

        <section class="receipt-sheet-section">
          <div class="receipt-sheet-section__title">
            <span class="receipt-sheet-section__bar" />
            商品信息
          </div>
          <table class="receipt-sheet-table receipt-sheet-table--product">
            <tbody v-for="item in receiptRows" :key="`${item.name}-${item.quoteLabel}-${item.noteText}`">
              <tr>
                <td colspan="2"><strong>课程名称：</strong>{{ item.name }}</td>
                <td colspan="2"><strong>收费方式：</strong>{{ item.chargingModeText }}</td>
                <td colspan="2"><strong>报价单：</strong>{{ item.quoteLabel }}</td>
              </tr>
              <tr>
                <th>购买数量</th>
                <th>赠送数量</th>
                <th>有效期</th>
                <th>原价</th>
                <th>优惠</th>
                <th>应收小计</th>
              </tr>
              <tr>
                <td>{{ item.quantityText }}</td>
                <td>{{ item.giftText }}</td>
                <td>{{ item.periodText }}</td>
                <td>{{ formatCurrency(item.originalAmount) }}</td>
                <td>{{ item.discountAmount > 0 ? formatNegativeCurrency(item.discountAmount) : '¥0.00' }}</td>
                <td>{{ formatCurrency(item.amount) }}</td>
              </tr>
              <tr v-if="item.noteText !== '-'">
                <td colspan="6"><strong>说明：</strong>{{ item.noteText }}</td>
              </tr>
            </tbody>
          </table>
        </section>

        <section class="receipt-sheet-section">
          <div class="receipt-sheet-section__title">
            <span class="receipt-sheet-section__bar" />
            订单信息
          </div>
          <table class="receipt-sheet-table">
            <tbody>
              <tr>
                <th>经办日期</th>
                <th>经办人</th>
                <th>订单销售员</th>
                <th>订单来源</th>
              </tr>
              <tr>
                <td>{{ formatDateOnly(detail.dealDate) }}</td>
                <td>{{ detail.staffName || '-' }}</td>
                <td>{{ detail.salePersonName || '-' }}</td>
                <td>{{ orderSourceText }}</td>
              </tr>
              <tr>
                <th>订单状态</th>
                <th>订单标签</th>
                <th>打印人</th>
                <th>打印时间</th>
              </tr>
              <tr>
                <td>{{ orderStatusText }}</td>
                <td>{{ orderTagText }}</td>
                <td>{{ printedBy }}</td>
                <td>{{ printedAt }}</td>
              </tr>
            </tbody>
          </table>
        </section>

        <section class="receipt-sheet-section">
          <div class="receipt-sheet-section__title">
            <span class="receipt-sheet-section__bar" />
            结算信息
          </div>
          <table class="receipt-sheet-table">
            <tbody>
              <tr>
                <th>订单应收</th>
                <th>整单优惠</th>
                <th>储值抵扣</th>
                <th>外部实收</th>
                <th>待收欠费</th>
                <th>金额大写</th>
              </tr>
              <tr>
                <td>{{ formatCurrency(grandTotalAmount) }}</td>
                <td>{{ orderDiscountAmount > 0 ? formatNegativeCurrency(orderDiscountAmount) : '¥0.00' }}</td>
                <td>{{ storageDeductionTotal > 0 ? formatNegativeCurrency(storageDeductionTotal) : '¥0.00' }}</td>
                <td>{{ formatCurrency(externalPaidAmount) }}</td>
                <td>{{ arrearAmount > 0 ? formatCurrency(arrearAmount) : '¥0.00' }}</td>
                <td>{{ amountInWords }}</td>
              </tr>
              <tr v-if="rechargeDeductionItems.length">
                <th>充值余额抵扣</th>
                <th>残联余额抵扣</th>
                <th>赠送余额抵扣</th>
                <th colspan="3">支付摘要</th>
              </tr>
              <tr v-if="rechargeDeductionItems.length">
                <td>{{ formatNegativeCurrency(rechargeDeductionItems.find(item => item.label === '充值余额抵扣')?.value || 0) }}</td>
                <td>{{ formatNegativeCurrency(rechargeDeductionItems.find(item => item.label === '残联余额抵扣')?.value || 0) }}</td>
                <td>{{ formatNegativeCurrency(rechargeDeductionItems.find(item => item.label === '赠送余额抵扣')?.value || 0) }}</td>
                <td colspan="3">{{ paymentSummaryText }}</td>
              </tr>
              <tr v-else>
                <th colspan="2">支付摘要</th>
                <td colspan="4">{{ paymentSummaryText }}</td>
              </tr>
            </tbody>
          </table>
        </section>

        <section class="receipt-sheet-section">
          <div class="receipt-sheet-section__title">
            <span class="receipt-sheet-section__bar" />
            备注信息
          </div>
          <table class="receipt-sheet-table">
            <tbody>
              <tr>
                <th>对外备注</th>
              </tr>
              <tr>
                <td>{{ detail.externalRemark || '-' }}</td>
              </tr>
            </tbody>
          </table>
        </section>

        <footer class="receipt-signature receipt-signature--sheet">
          <div>经办人：{{ detail.staffName || printedBy }}</div>
          <div>复核人：________________</div>
          <div>家长签字：________________</div>
        </footer>
      </section>

      <section v-else-if="templateMode === 'dot'" class="receipt-paper receipt-paper--dot">
        <div class="dot-head">
          <div class="dot-head__title">
            {{ orgName }}
          </div>
          <div class="dot-head__subtitle">
            业务收据（针式打印联）
          </div>
          <div class="dot-head__meta-grid">
            <div class="dot-head__meta-item">
              <span class="dot-head__meta-label">票据编号：</span>
              <span class="dot-head__meta-value">{{ detail.orderNumber || '-' }}</span>
            </div>
            <div class="dot-head__meta-item">
              <span class="dot-head__meta-label">打印时间：</span>
              <span class="dot-head__meta-value">{{ printedAt }}</span>
            </div>
          </div>
        </div>

        <div class="dot-info-grid">
          <div>学员：{{ detail.studentName || '-' }}</div>
          <div>手机：{{ detail.studentPhone || '-' }}</div>
          <div>类型：{{ orderTypeText }}</div>
          <div>状态：{{ orderStatusText }}</div>
          <div>来源：{{ orderSourceText }}</div>
          <div>经办日期：{{ formatDateOnly(detail.dealDate) }}</div>
          <div>经办人：{{ detail.staffName || '-' }}</div>
          <div>销售员：{{ detail.salePersonName || '-' }}</div>
        </div>

        <div class="dot-divider" />

        <div class="dot-section-title">
          业务明细
        </div>
        <div v-for="item in receiptRows" :key="`${item.name}-${item.quoteLabel}`" class="dot-item">
          <div class="dot-item__top">
            <span>{{ item.name }}</span>
            <span>{{ formatCurrency(item.amount) }}</span>
          </div>
          <div class="dot-item__sub">
            {{ item.quoteLabel }}｜购买 {{ item.quantityText }}｜赠送 {{ item.giftText }}
          </div>
          <div class="dot-item__sub">
            {{ item.periodText }}
          </div>
        </div>

        <div class="dot-divider" />

        <div class="dot-section-title">
          优惠与结算
        </div>
        <div class="dot-summary-grid">
          <div>订单应收：{{ formatCurrency(grandTotalAmount) }}</div>
          <div>整单优惠：{{ orderDiscountAmount > 0 ? formatNegativeCurrency(orderDiscountAmount) : '¥0.00' }}</div>
          <div>储值抵扣：{{ storageDeductionTotal > 0 ? formatNegativeCurrency(storageDeductionTotal) : '¥0.00' }}</div>
          <div>外部实收：{{ formatCurrency(externalPaidAmount) }}</div>
          <div>待收欠费：{{ arrearAmount > 0 ? formatCurrency(arrearAmount) : '¥0.00' }}</div>
          <div>金额大写：{{ amountInWords }}</div>
        </div>

        <div v-if="rechargeDeductionItems.length" class="dot-divider" />
        <div v-if="rechargeDeductionItems.length" class="dot-section-title">
          储值账户抵扣
        </div>
        <div v-for="item in rechargeDeductionItems" :key="item.label" class="dot-row">
          <span>{{ item.label }}</span>
          <span>{{ formatNegativeCurrency(item.value) }}</span>
        </div>

        <div class="dot-divider" />

        <div class="dot-section-title">
          支付记录
        </div>
        <div v-if="paymentTableRows.length">
          <div v-for="item in paymentTableRows" :key="`${item.methodText}-${item.paidAtText}`" class="dot-row">
            <span>{{ item.methodText }} / {{ item.accountText }}</span>
            <span>{{ item.amountText }}</span>
          </div>
          <div v-for="item in paymentTableRows" :key="`${item.paidAtText}-${item.remarkText}`" class="dot-row dot-row--sub">
            <span>{{ item.paidAtText }}</span>
            <span>{{ item.remarkText }}</span>
          </div>
        </div>
        <div v-else class="dot-empty">
          {{ paymentSummaryText }}
        </div>

        <div class="dot-divider" />

        <div class="dot-footer">
          <span>经办人：{{ detail.staffName || printedBy }}</span>
          <span>打印人：{{ printedBy }}</span>
          <span>家长签字：____________</span>
        </div>
      </section>

      <section v-else class="receipt-paper receipt-paper--receipt">
        <div class="mini-head">
          <div class="mini-head__org">
            {{ orgName }}
          </div>
          <div class="mini-head__title">
            业务收据
          </div>
          <div class="mini-head__meta">
            {{ detail.orderNumber || '-' }}
          </div>
        </div>

        <div class="mini-block">
          <div class="mini-row">
            <span>学员</span>
            <span>{{ detail.studentName || '-' }}</span>
          </div>
          <div class="mini-row">
            <span>手机</span>
            <span>{{ detail.studentPhone || '-' }}</span>
          </div>
          <div class="mini-row">
            <span>类型</span>
            <span>{{ orderTypeText }}</span>
          </div>
          <div class="mini-row">
            <span>日期</span>
            <span>{{ formatDateOnly(detail.dealDate) }}</span>
          </div>
          <div class="mini-row">
            <span>经办人</span>
            <span>{{ detail.staffName || '-' }}</span>
          </div>
        </div>

        <div class="mini-divider" />

        <div class="mini-section-title">
          业务明细
        </div>
        <div v-for="item in receiptRows" :key="`${item.name}-${item.quoteLabel}`" class="mini-item">
          <div class="mini-row">
            <span>{{ item.name }}</span>
            <span>{{ formatCurrency(item.amount) }}</span>
          </div>
          <div class="mini-sub">
            {{ item.quoteLabel }}
          </div>
          <div class="mini-sub">
            购买 {{ item.quantityText }} / 赠送 {{ item.giftText }}
          </div>
        </div>

        <div v-if="rechargeDeductionItems.length" class="mini-divider" />
        <div v-if="rechargeDeductionItems.length" class="mini-section-title">
          储值抵扣
        </div>
        <div v-for="item in rechargeDeductionItems" :key="item.label" class="mini-row">
          <span>{{ item.label }}</span>
          <span>{{ formatNegativeCurrency(item.value) }}</span>
        </div>

        <div class="mini-divider" />

        <div class="mini-section-title">
          金额汇总
        </div>
        <div class="mini-row">
          <span>订单应收</span>
          <span>{{ formatCurrency(grandTotalAmount) }}</span>
        </div>
        <div class="mini-row">
          <span>整单优惠</span>
          <span>{{ orderDiscountAmount > 0 ? formatNegativeCurrency(orderDiscountAmount) : '¥0.00' }}</span>
        </div>
        <div class="mini-row">
          <span>外部实收</span>
          <span>{{ formatCurrency(externalPaidAmount) }}</span>
        </div>
        <div class="mini-row">
          <span>待收欠费</span>
          <span>{{ arrearAmount > 0 ? formatCurrency(arrearAmount) : '¥0.00' }}</span>
        </div>

        <div class="mini-divider" />

        <div class="mini-section-title">
          支付摘要
        </div>
        <div class="mini-sub">
          {{ paymentSummaryText }}
        </div>
        <div class="mini-sub">
          金额大写：{{ amountInWords }}
        </div>
        <div class="mini-footer">
          打印人：{{ printedBy }}
        </div>
      </section>
    </div>

    <div class="receipt-action-bar print-hidden">
      <button type="button" class="receipt-print-btn" @click="printPage">
        打印
      </button>
      <button type="button" class="receipt-print-btn receipt-print-btn--ghost" @click="downloadPage">
        下载
      </button>
    </div>
  </div>
</template>

<style scoped lang="less">
.receipt-print-page {
  min-height: 100vh;
  background: linear-gradient(180deg, #f6f8fc 0%, #eef3fb 100%);
  color: #1f2329;
}

.receipt-toolbar {
  position: sticky;
  top: 0;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  padding: 12px 24px;
  background: rgba(255, 255, 255, 0.96);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid #e8eef7;
}

.receipt-toolbar__left {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 18px;
  width: min(1280px, 100%);
}

.receipt-toolbar__label {
  flex-shrink: 0;
  font-size: 16px;
  font-weight: 600;
  color: #30353d;
}

.receipt-toolbar__tabs {
  display: inline-flex;
  border: 1px solid #d7deeb;
  border-radius: 18px;
  overflow: hidden;
  background: #fff;
  width: auto;
  max-width: 100%;
}

.receipt-tab {
  flex: none;
  border: none;
  background: transparent;
  height: 32px;
  padding: 0 18px;
  font-size: 16px;
  font-weight: 600;
  line-height: 32px;
  color: #222;
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
}

.receipt-tab:nth-child(1) {
  min-width: 166px;
}

.receipt-tab:nth-child(2) {
  min-width: 194px;
}

.receipt-tab:nth-child(3) {
  min-width: 162px;
}

.receipt-tab--active {
  background: #1677ff;
  color: #fff;
  box-shadow: inset 0 -2px 0 rgba(255, 255, 255, 0.12);
}

.receipt-print-btn {
  border: none;
  background: #1677ff;
  color: #fff;
  border-radius: 10px;
  min-width: 96px;
  padding: 8px 18px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  box-shadow: 0 10px 20px rgba(22, 119, 255, 0.18);
}

.receipt-loading {
  padding: 100px 24px;
  text-align: center;
  color: #65707f;
  font-size: 16px;
}

.receipt-print-tip {
  display: flex;
  justify-content: center;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  padding: 10px 24px 0;
  color: #a15c00;
  font-size: 13px;
}

.receipt-print-tip__divider {
  color: #d1a15a;
}

.receipt-preview-wrap {
  display: flex;
  justify-content: center;
  padding: 20px 24px 20px;
}

.receipt-paper {
  background: #fff;
  border-radius: 10px;
  border: 1px solid rgba(218, 226, 238, 0.9);
  box-shadow: 0 16px 40px rgba(25, 44, 86, 0.07);
}

.receipt-paper--a4 {
  width: 210mm;
  min-height: 297mm;
  padding: 18mm 16mm 8mm;
  box-sizing: border-box;
}

.receipt-a4-head {
  text-align: center;
}

.receipt-a4-head__org {
  font-size: 28px;
  font-weight: 700;
  letter-spacing: 1px;
  line-height: 1.2;
}

.receipt-a4-head__line {
  width: 100%;
  margin: 8px 0 12px;
  border-top: 2px solid #222;
}

.receipt-a4-head__meta {
  display: grid;
  grid-template-columns: 1fr 1fr 1.5fr;
  gap: 12px;
  margin-bottom: 12px;
  text-align: left;
  font-size: 13px;
  font-weight: 600;
}

.receipt-sheet-section {
  margin-top: 12px;
}

.receipt-sheet-section__title {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-bottom: 6px;
  font-size: 14px;
  font-weight: 700;
}

.receipt-sheet-section__bar {
  width: 4px;
  height: 12px;
  border-radius: 999px;
  background: #696969;
  flex-shrink: 0;
}

.receipt-sheet-table {
  width: 100%;
  border-collapse: collapse;
  table-layout: fixed;
  font-size: 12px;
  margin-bottom: 30px;
}

.receipt-sheet-table th,
.receipt-sheet-table td {
  border: 1px solid #808a98;
  padding: 9px 10px;
  vertical-align: middle;
  line-height: 1.4;
  word-break: break-word;
}

.receipt-sheet-table th {
  font-weight: 700;
  text-align: left;
  background: #fff;
}

.receipt-sheet-table--product tbody + tbody tr:first-child td {
  border-top-width: 2px;
}

.receipt-signature--sheet {
  margin-top: 14px;
  padding-top: 0;
  border-top: none;
  font-size: 13px;
}

.receipt-table {
  width: 100%;
  border-collapse: collapse;
  table-layout: fixed;
  font-size: 13px;
}

.receipt-table th,
.receipt-table td {
  border: 1px solid #dfe6f1;
  padding: 10px 12px;
  vertical-align: top;
  line-height: 1.6;
}

.receipt-table th {
  background: #f7f9fc;
  font-weight: 700;
  color: #273142;
}

.receipt-signature {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
  margin-top: 14px;
  font-size: 13px;
}

.receipt-action-bar {
  position: sticky;
  bottom: 0;
  z-index: 12;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 8px 16px 10px;
  background: rgba(255, 255, 255, 0.96);
  border-top: 1px solid #e8eef7;
  backdrop-filter: blur(12px);
}

.receipt-print-btn--ghost {
  background: #fff;
  color: #1677ff;
  border: 1px solid #1677ff;
  box-shadow: none;
}

.receipt-paper--dot {
  width: 210mm;
  min-height: 150mm;
  padding: 14mm 12mm;
  font-family: 'Courier New', monospace;
  border-radius: 16px;
}

.dot-head {
  text-align: center;
}

.dot-head__title {
  font-size: 22px;
  font-weight: 700;
}

.dot-head__subtitle {
  margin-top: 6px;
  font-size: 16px;
  font-weight: 700;
}

.dot-head__meta-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px 22px;
  margin-top: 10px;
  text-align: left;
}

.dot-head__meta-item {
  display: flex;
  align-items: baseline;
  gap: 6px;
  min-width: 0;
  font-size: 13px;
}

.dot-head__meta-label {
  flex-shrink: 0;
  font-weight: 700;
}

.dot-head__meta-value {
  min-width: 0;
  word-break: break-all;
}

.dot-info-grid,
.dot-summary-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px 20px;
  margin-top: 14px;
  font-size: 14px;
}

.dot-divider,
.mini-divider {
  border-top: 1px dashed #8a94a6;
  margin: 12px 0;
}

.dot-section-title,
.mini-section-title {
  margin-bottom: 10px;
  font-weight: 700;
}

.dot-item {
  padding: 8px 0;
}

.dot-item:last-child {
  border-bottom: none;
}

.dot-item__top,
.dot-row,
.mini-row {
  display: flex;
  justify-content: space-between;
  gap: 12px;
}

.dot-item__top {
  font-weight: 700;
}

.dot-item__sub,
.dot-row--sub,
.mini-sub {
  margin-top: 4px;
  color: #5d6777;
  font-size: 13px;
  line-height: 1.5;
}

.dot-empty {
  color: #5d6777;
  line-height: 1.6;
}

.dot-footer {
  display: flex;
  justify-content: space-between;
  gap: 14px;
  margin-top: 10px;
  font-size: 13px;
}

.receipt-paper--receipt {
  width: 80mm;
  min-height: 210mm;
  padding: 8mm 6mm 9mm;
  font-family: 'Courier New', monospace;
  border-radius: 14px;
}

.mini-head {
  text-align: center;
}

.mini-head__org {
  font-size: 16px;
  font-weight: 700;
  line-height: 1.5;
}

.mini-head__title {
  margin-top: 4px;
  font-size: 14px;
  font-weight: 700;
}

.mini-head__meta {
  margin-top: 4px;
  font-size: 11px;
  word-break: break-all;
}

.mini-block,
.mini-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.mini-item + .mini-item {
  margin-top: 8px;
}

.mini-row {
  font-size: 12px;
  line-height: 1.5;
}

.mini-sub {
  font-size: 11px;
}

.mini-footer {
  margin-top: 10px;
  text-align: right;
  font-size: 11px;
  color: #6a7380;
}

@media print {
  @page {
    size: A4;
    margin: 0mm !important;
  }

  .print-hidden {
    display: none !important;
  }

  .receipt-print-page {
    min-height: 0;
    background: #fff;
  }

  .receipt-preview-wrap {
    padding: 0;
  }

  .receipt-paper {
    border: none;
    border-radius: 0;
    box-shadow: none;
    break-inside: avoid;
  }

  .receipt-paper--a4,
  .receipt-paper--dot,
  .receipt-paper--receipt {
    width: 100%;
    min-height: 0;
  }

}
</style>
