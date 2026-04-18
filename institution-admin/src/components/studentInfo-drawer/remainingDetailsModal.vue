<script setup>
import { CloseOutlined, QuestionCircleOutlined } from '@ant-design/icons-vue'
import { computed, ref, watch } from 'vue'
import dayjs from 'dayjs'
import OrderDetailDrawer from '@/components/common/order-detail-drawer.vue'
import {
  getSubTuitionAccountPriorityConfigListApi,
  getTuitionAccountSubAccountDateInfoApi,
} from '@/api/edu-center/tuition-account'
import { getSubTuitionAccountFlowRecordListApi } from '@/api/finance-center/tuition-account-flow'
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

const ORDER_TYPE_LABEL_MAP = {
  1: '报名/续费',
  2: '储值账户充值',
  3: '退课',
  4: '储值账户退费',
  5: '转课',
  6: '退教材费',
  7: '退学杂费',
}

const PRIORITY_TEXT_MAP = {
  1: '先进先出',
  2: '优先扣减更早到期账户',
  3: '优先扣减最近生成账户',
}

const LESSON_TYPE_LABEL_MAP = {
  1: '班级授课',
  2: '1v1授课',
}

const CHARGING_MODE_LABEL_MAP = {
  1: '课时',
  2: '时段',
  3: '金额',
}

const FLOW_SOURCE_TYPE_LABEL_MAP = {
  1: '报名',
  2: '转入',
  3: '跨校转入',
  4: '跨校上课转入',
  5: '课消退还',
  6: '撤销结课',
  7: '过期撤回返还',
  8: '撤回退课订单',
  9: '撤销转出',
  10: '撤回导入课消',
  11: '撤回每日自动课消',
  12: '课消',
  13: '导入课消',
  14: '课消补扣',
  15: '每日自动课消',
  16: '课消欠费清算',
  17: '转出',
  18: '跨校转出',
  19: '跨校上课转出',
  20: '结课',
  21: '到期结算',
  22: '退费',
  23: '订单作废',
  24: '作废跨校转入',
  25: '手动结课',
}

const FLOW_OUT_SOURCE_TYPES = new Set([12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25])

const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const loading = ref(false)
const manualSort = ref(false)
const baseRows = ref([])
const tableRows = ref([])
const priorityConfigs = ref([])
const orderDetailOpen = ref(false)
const currentOrderId = ref('')
const feeChangeOpen = ref(false)
const feeChangeLoading = ref(false)
const currentFeeChangeRow = ref(null)
const feeChangeRows = ref([])

const tuitionAccountId = computed(() => String(props.record?.id || props.record?.tuitionAccountId || ''))
const lessonChargingMode = computed(() => Number(props.record?.lessonChargingMode || 0))
const showManualSort = computed(() => lessonChargingMode.value === 1)
const courseName = computed(() => props.record?.lessonName || props.record?.productName || '-')

const summaryTags = computed(() => {
  const tags = []
  const lessonType = Number(props.record?.lessonType || 0)
  if (lessonType) {
    tags.push(LESSON_TYPE_LABEL_MAP[lessonType] || '-')
  }
  if (lessonChargingMode.value) {
    tags.push(CHARGING_MODE_LABEL_MAP[lessonChargingMode.value] || '-')
  }
  return tags
})

const quantityColumnTitle = computed(() => {
  if (lessonChargingMode.value === 2)
    return '剩余天数'
  if (lessonChargingMode.value === 3)
    return '剩余金额'
  return '剩余课时'
})

const periodColumnTitle = computed(() => (
  lessonChargingMode.value === 2 ? '有效时段' : '有效期至'
))

const totalColumnTitle = computed(() => {
  if (lessonChargingMode.value === 2)
    return '总时间'
  if (lessonChargingMode.value === 3)
    return '总金额'
  return '总课时'
})

const smartSortText = computed(() => {
  const enabledConfig = [...priorityConfigs.value]
    .filter(item => item?.isEnabled)
    .sort((a, b) => Number(a?.sortWeight || 0) - Number(b?.sortWeight || 0))[0]
  if (!enabledConfig)
    return '先进先出'
  return PRIORITY_TEXT_MAP[Number(enabledConfig.priorityType || 0)] || `规则${enabledConfig.priorityType || '-'}`
})

const columns = computed(() => [
  { title: quantityColumnTitle.value, key: 'quantityValue', dataIndex: 'quantityValue', width: 120 },
  { title: periodColumnTitle.value, key: 'periodValue', dataIndex: 'periodValue', width: 220 },
  { title: '来源', key: 'sourceLabel', dataIndex: 'sourceLabel', width: 140 },
  { title: '生成时间', key: 'createdTime', dataIndex: 'createdTime', width: 170 },
  { title: totalColumnTitle.value, key: 'totalValue', dataIndex: 'totalValue', width: 120 },
  { title: '课单价', key: 'unitPrice', dataIndex: 'unitPrice', width: 120 },
  { title: '总学费', key: 'totalTuition', dataIndex: 'totalTuition', width: 120 },
  { title: '已用学费金额', key: 'usedTuition', dataIndex: 'usedTuition', width: 140 },
  { title: '剩余学费金额', key: 'remainTuition', dataIndex: 'remainTuition', width: 140 },
  { title: '转退学费金额', key: 'transferredTuition', dataIndex: 'transferredTuition', width: 140 },
  { title: '实收学费金额', key: 'paidTuition', dataIndex: 'paidTuition', width: 140 },
  { title: '欠费学费金额', key: 'arrearTuition', dataIndex: 'arrearTuition', width: 140 },
  { title: '实收剩余可用', key: 'paidRemaining', dataIndex: 'paidRemaining', width: 140 },
  { title: '操作', key: 'action', dataIndex: 'action', width: 250, fixed: 'right' },
])

const feeChangeColumns = [
  { title: '变动类型', key: 'sourceType', dataIndex: 'sourceType', width: 120 },
  { title: '变动时间', key: 'createdTime', dataIndex: 'createdTime', width: 160 },
  { title: '数量变动', key: 'quantity', dataIndex: 'quantity', width: 120 },
  { title: '学费变动', key: 'tuition', dataIndex: 'tuition', width: 120 },
  { title: '变动后剩余数量', key: 'balanceQuantity', dataIndex: 'balanceQuantity', width: 140 },
  { title: '变动后剩余学费', key: 'balanceTuition', dataIndex: 'balanceTuition', width: 140 },
]

watch(
  () => [props.open, tuitionAccountId.value],
  async ([open]) => {
    if (!open)
      return
    manualSort.value = false
    await loadData()
  },
  { immediate: false },
)

watch(manualSort, (enabled) => {
  if (!enabled) {
    tableRows.value = cloneRows(baseRows.value)
  }
})

function cloneRows(list = []) {
  return list.map(item => ({ ...item }))
}

async function loadData() {
  if (!tuitionAccountId.value) {
    baseRows.value = []
    tableRows.value = []
    priorityConfigs.value = []
    return
  }
  loading.value = true
  try {
    const requests = [getTuitionAccountSubAccountDateInfoApi({ tuitionAccountId: tuitionAccountId.value })]
    if (showManualSort.value) {
      requests.push(getSubTuitionAccountPriorityConfigListApi())
    }
    const [subAccountRes, priorityRes] = await Promise.all(requests)
    if (subAccountRes.code !== 200) {
      throw new Error(subAccountRes.message || '加载剩余详情失败')
    }
    const list = Array.isArray(subAccountRes.result?.list) ? subAccountRes.result.list : []
    baseRows.value = list.map(item => normalizeRow(item))
    tableRows.value = cloneRows(baseRows.value)
    if (showManualSort.value) {
      if (priorityRes?.code !== 200) {
        throw new Error(priorityRes?.message || '加载排序配置失败')
      }
      priorityConfigs.value = Array.isArray(priorityRes?.result?.list) ? priorityRes.result.list : []
    }
    else {
      priorityConfigs.value = []
    }
  }
  catch (error) {
    baseRows.value = []
    tableRows.value = []
    priorityConfigs.value = []
    messageService.error(error?.message || '加载剩余详情失败')
  }
  finally {
    loading.value = false
  }
}

function normalizeRow(item = {}) {
  return {
    ...item,
    id: String(item.id || ''),
    orderId: String(item.orderId || ''),
    quantityValue: getQuantityValue(item),
    periodValue: getPeriodValue(item),
    sourceLabel: getSourceLabel(item),
    totalValue: getTotalValue(item),
  }
}

function getQuantityValue(item) {
  if (lessonChargingMode.value === 3) {
    return formatMoney(item?.tuition || 0)
  }
  const unit = lessonChargingMode.value === 2 ? '天' : '课时'
  return `${formatCount(item?.remainDays ?? item?.quantity ?? 0)}${unit}`
}

function getPeriodValue(item) {
  if (lessonChargingMode.value === 2) {
    const start = formatDate(item?.startDate || item?.startTime || item?.activedAt)
    const end = formatDate(item?.endDate)
    if (start === '-' || end === '-')
      return '-'
    return `${start} ~ ${end}`
  }
  const expire = formatDate(item?.expireDate || item?.endDate)
  if (expire !== '-')
    return expire
  return '不限制'
}

function getSourceLabel(item) {
  const baseLabel = ORDER_TYPE_LABEL_MAP[Number(item?.sourceType || 0)] || '报名/续费'
  if (item?.isFree) {
    return baseLabel === '报名/续费' ? '赠送' : `${baseLabel}（赠送）`
  }
  return baseLabel
}

function getTotalValue(item) {
  if (lessonChargingMode.value === 3) {
    return formatMoney(item?.totalTuition || 0)
  }
  const unit = lessonChargingMode.value === 2 ? '天' : '课时'
  return `${formatCount(item?.totalDays || 0)}${unit}`
}

function formatDate(value) {
  if (!value || `${value}`.startsWith('0001-01-01'))
    return '-'
  const parsed = dayjs(value)
  if (!parsed.isValid())
    return '-'
  return parsed.format('YYYY-MM-DD')
}

function formatDateTime(value) {
  if (!value || `${value}`.startsWith('0001-01-01'))
    return '-'
  const parsed = dayjs(value)
  if (!parsed.isValid())
    return '-'
  return parsed.format('YYYY-MM-DD HH:mm')
}

function formatCount(value) {
  const num = Number(value || 0)
  if (Number.isInteger(num))
    return String(num)
  return num.toFixed(2)
}

function formatMoney(value) {
  const num = Number(value || 0)
  return `¥ ${num.toLocaleString('zh-CN', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })}`
}

function closeFun() {
  openModal.value = false
}

function onDragStart(event, index) {
  if (!showManualSort.value || !manualSort.value) {
    event.preventDefault()
    return
  }
  event.dataTransfer.setData('text/plain', String(index))
  event.dataTransfer.effectAllowed = 'move'
}

function onDragOver(event) {
  if (!showManualSort.value || !manualSort.value)
    return
  event.preventDefault()
  event.dataTransfer.dropEffect = 'move'
}

function onDrop(event, dropIndex) {
  if (!showManualSort.value || !manualSort.value)
    return
  event.preventDefault()
  const dragIndex = Number(event.dataTransfer.getData('text/plain'))
  if (Number.isNaN(dragIndex) || dragIndex === dropIndex)
    return
  const nextRows = cloneRows(tableRows.value)
  const [dragged] = nextRows.splice(dragIndex, 1)
  if (!dragged)
    return
  nextRows.splice(dropIndex, 0, dragged)
  tableRows.value = nextRows
}

function handleAdjustPeriod() {
  messageService.info('调整时段待实现')
}

function handleViewOrder(row) {
  const orderId = String(row?.orderId || '').trim()
  if (!orderId) {
    messageService.info('暂无订单信息')
    return
  }
  currentOrderId.value = orderId
  orderDetailOpen.value = true
}

async function handleViewFeeChange(row) {
  const tuitionAccountId = String(row?.id || '').trim()
  if (!tuitionAccountId) {
    messageService.info('暂无学费变动记录')
    return
  }
  feeChangeOpen.value = true
  feeChangeLoading.value = true
  currentFeeChangeRow.value = row || null
  try {
    const res = await getSubTuitionAccountFlowRecordListApi({
      queryModel: {
        tuitionAccountId,
      },
      pageRequestModel: {
        needTotal: true,
        pageSize: 100,
        pageIndex: 1,
        skipCount: 0,
      },
      sortModel: {
        orderByCreatedTime: 0,
      },
    })
    if (res.code !== 200) {
      throw new Error(res.message || '加载学费变动记录失败')
    }
    feeChangeRows.value = Array.isArray(res.result?.list) ? res.result.list : []
  }
  catch (error) {
    feeChangeRows.value = []
    messageService.error(error?.message || '加载学费变动记录失败')
  }
  finally {
    feeChangeLoading.value = false
  }
}

function getFlowPrefix(sourceType) {
  return FLOW_OUT_SOURCE_TYPES.has(Number(sourceType || 0)) ? '-' : '+'
}

function getFlowSourceLabel(sourceType) {
  return FLOW_SOURCE_TYPE_LABEL_MAP[Number(sourceType || 0)] || '-'
}

function formatSignedMoney(value, sourceType) {
  const num = Number(value || 0)
  return `${getFlowPrefix(sourceType)}${formatMoney(Math.abs(num))}`
}

function formatFlowQuantity(item) {
  const amount = Number(item?.quantity || 0)
  if (lessonChargingMode.value === 3) {
    return formatSignedMoney(amount, item?.sourceType)
  }
  const unit = lessonChargingMode.value === 2 ? '天' : '课时'
  const prefix = getFlowPrefix(item?.sourceType)
  return `${prefix}${formatCount(Math.abs(amount))}${unit}`
}

function formatFlowBalanceQuantity(item) {
  if (lessonChargingMode.value === 3) {
    return formatMoney(item?.balanceQuantity || 0)
  }
  const unit = lessonChargingMode.value === 2 ? '天' : '课时'
  return `${formatCount(item?.balanceQuantity || 0)}${unit}`
}
</script>

<template>
  <a-modal
    v-model:open="openModal"
    centered
    class="modal-content-box"
    :keyboard="false"
    :closable="false"
    :mask-closable="false"
    :width="900"
    :footer="false"
    :destroy-on-close="true"
  >
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>剩余详情</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <div class="contenter scrollbar">
      <a-spin :spinning="loading">
        <div class="header-card">
          <div class="header-main">
            <div class="course-title">
              课程名称：{{ courseName }}
            </div>
            <div v-if="showManualSort" class="course-tip">
              学费变动（课消、转课或退课）会按照当前顺序扣减
            </div>
            <a-space v-if="summaryTags.length" :size="[8, 8]" wrap class="course-tags">
              <span
                v-for="tag in summaryTags"
                :key="tag"
                class="course-tag"
              >
                {{ tag }}
              </span>
            </a-space>
          </div>
          <div v-if="showManualSort" class="header-sort">
            <span class="sort-text">
              <span>手动排序</span>
              <a-tooltip overlay-class-name="manual-sort-tooltip">
                <template #title>
                  <div class="tooltip-content">
                    <div>开启手动排序后，可拖拽调整当前明细顺序。</div>
                    <div>当前智能排序：{{ smartSortText }}</div>
                  </div>
                </template>
                <QuestionCircleOutlined class="cursor-pointer" />
              </a-tooltip>
            </span>
            <a-switch v-model:checked="manualSort" />
          </div>
        </div>

        <a-table
          :columns="columns"
          :data-source="tableRows"
          :pagination="false"
          row-key="id"
          size="small"
          :scroll="{ x: 800 }"
          :custom-row="(record, index) => ({
            draggable: showManualSort && manualSort,
            class: showManualSort && manualSort ? 'draggable-row' : '',
            onDragstart: event => onDragStart(event, index),
            onDragover: onDragOver,
            onDrop: event => onDrop(event, index),
          })"
        >
          <template #bodyCell="{ column, record: row }">
            <template v-if="column.key === 'quantityValue'">
              {{ row.quantityValue }}
            </template>
            <template v-else-if="column.key === 'periodValue'">
              {{ row.periodValue }}
            </template>
            <template v-else-if="column.key === 'sourceLabel'">
              {{ row.sourceLabel }}
            </template>
            <template v-else-if="column.key === 'createdTime'">
              {{ formatDateTime(row.createdTime) }}
            </template>
            <template v-else-if="column.key === 'totalValue'">
              {{ row.totalValue }}
            </template>
            <template v-else-if="column.key === 'unitPrice'">
              {{ Number(row.unitPrice || 0) > 0 ? formatMoney(row.unitPrice) : '-' }}
            </template>
            <template v-else-if="column.key === 'totalTuition'">
              {{ formatMoney(row.totalTuition) }}
            </template>
            <template v-else-if="column.key === 'usedTuition'">
              {{ formatMoney(row.usedTuition) }}
            </template>
            <template v-else-if="column.key === 'remainTuition'">
              {{ formatMoney(row.tuition) }}
            </template>
            <template v-else-if="column.key === 'transferredTuition'">
              {{ formatMoney(row.transferredTuition) }}
            </template>
            <template v-else-if="column.key === 'paidTuition'">
              {{ formatMoney(row.paidTuition) }}
            </template>
            <template v-else-if="column.key === 'arrearTuition'">
              {{ formatMoney(row.arrearTuition) }}
            </template>
            <template v-else-if="column.key === 'paidRemaining'">
              {{ formatMoney(row.paidRemaining) }}
            </template>
            <template v-else-if="column.key === 'action'">
              <a-space :size="8" wrap>
                <a
                  v-if="lessonChargingMode === 2"
                  class="action-link"
                  @click="handleAdjustPeriod(row)"
                >
                  调整时段
                </a>
                <a class="action-link" @click="handleViewOrder(row)">
                  查看订单
                </a>
                <a class="action-link" @click="handleViewFeeChange(row)">
                  学费变动记录
                </a>
              </a-space>
            </template>
          </template>
        </a-table>
      </a-spin>
    </div>
    <OrderDetailDrawer v-model:open="orderDetailOpen" :order-id="currentOrderId" />
    <a-modal
      v-model:open="feeChangeOpen"
      centered
      title="学费变动记录"
      :width="860"
      :footer="false"
      :destroy-on-close="true"
    >
      <a-spin :spinning="feeChangeLoading">
        <div class="mb-12px text-#666">
          {{ currentFeeChangeRow?.sourceLabel || courseName }}
        </div>
        <a-table
          :columns="feeChangeColumns"
          :data-source="feeChangeRows"
          :pagination="false"
          row-key="id"
          size="small"
          :scroll="{ x: 820 }"
        >
          <template #bodyCell="{ column, record: flowRow }">
            <template v-if="column.key === 'sourceType'">
              {{ getFlowSourceLabel(flowRow.sourceType) }}
            </template>
            <template v-else-if="column.key === 'createdTime'">
              {{ formatDateTime(flowRow.createdTime) }}
            </template>
            <template v-else-if="column.key === 'quantity'">
              {{ formatFlowQuantity(flowRow) }}
            </template>
            <template v-else-if="column.key === 'tuition'">
              {{ formatSignedMoney(flowRow.tuition, flowRow.sourceType) }}
            </template>
            <template v-else-if="column.key === 'balanceQuantity'">
              {{ formatFlowBalanceQuantity(flowRow) }}
            </template>
            <template v-else-if="column.key === 'balanceTuition'">
              {{ formatMoney(flowRow.balanceTuition) }}
            </template>
          </template>
        </a-table>
      </a-spin>
    </a-modal>
  </a-modal>
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

.contenter {
  padding: 24px;
}

.header-card {
  align-items: flex-start;
  background: #fff;
  border: 1px solid #f0f0f0;
  border-radius: 16px;
  display: flex;
  justify-content: space-between;
  margin-bottom: 16px;
  padding: 20px 24px;
}

.header-main {
  min-width: 0;
}

.course-title {
  color: #222;
  font-size: 16px;
  font-weight: 500;
  line-height: 24px;
}

.course-tip {
  color: #888;
  font-size: 14px;
  line-height: 22px;
  margin-top: 8px;
}

.course-tags {
  margin-top: 12px;
}

.course-tag {
  background: #e6f0ff;
  border-radius: 999px;
  color: #06f;
  display: inline-flex;
  font-size: 12px;
  line-height: 20px;
  padding: 0 12px;
}

.header-sort {
  align-items: center;
  display: flex;
  gap: 8px;
  padding-left: 16px;
  white-space: nowrap;
}

.sort-text {
  align-items: center;
  color: #666;
  display: inline-flex;
  font-size: 14px;
  gap: 4px;
}

.action-link {
  color: #06f;
}

:deep(.manual-sort-tooltip) {
  .ant-tooltip-inner {
    background-color: #4a4a4a !important;
    border-radius: 6px !important;
    color: #fff !important;
    font-size: 12px !important;
    line-height: 1.6 !important;
    padding: 8px 12px !important;
  }

  .ant-tooltip-arrow::before {
    background-color: #4a4a4a !important;
  }
}

.tooltip-content {
  div + div {
    margin-top: 2px;
  }
}

:deep(.draggable-row) {
  cursor: move;
}

:deep(.draggable-row:hover > td) {
  background: #fafcff !important;
}
</style>

<style>
.modal-content-box .ant-modal-header {
  padding: 10px 16px !important;
  margin-bottom: 0;
}

.modal-content-box .ant-modal-body {
  padding: 0 !important;
}
</style>
