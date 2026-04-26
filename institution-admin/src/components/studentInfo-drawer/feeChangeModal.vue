<script setup>
import { CloseOutlined } from '@ant-design/icons-vue'
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import dayjs from 'dayjs'
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

const LESSON_TYPE_MAP = {
  1: '班级授课',
  2: '1v1授课',
}

const LESSON_CHARGING_MODE_MAP = {
  1: '课时',
  2: '时段',
  3: '金额',
}

const LESSON_CHARGING_UNIT_MAP = {
  1: '课时',
  2: '天',
  3: '元',
}

const FLOW_TYPE_GROUPS = [
  {
    children: [
      { id: 1, label: '报名', direction: 'in' },
      { id: 2, label: '转入', direction: 'in' },
      { id: 3, label: '跨校转入', direction: 'in' },
      { id: 4, label: '跨校上课转入', direction: 'in' },
      { id: 5, label: '课消退还', direction: 'in' },
      { id: 6, label: '撤销结课', direction: 'in' },
      { id: 7, label: '过期撤回返还', direction: 'in' },
      { id: 8, label: '撤回退课订单', direction: 'in' },
      { id: 9, label: '撤销转出', direction: 'in' },
      { id: 10, label: '撤回导入课消', direction: 'in' },
      { id: 11, label: '撤回每日自动课消', direction: 'in' },
      { id: 12, label: '课消', direction: 'out' },
      { id: 13, label: '导入课消', direction: 'out' },
      { id: 14, label: '课消补扣', direction: 'out' },
      { id: 15, label: '每日自动课消', direction: 'out' },
      { id: 16, label: '课消欠费清算', direction: 'out' },
      { id: 17, label: '转出', direction: 'out' },
      { id: 18, label: '跨校转出', direction: 'out' },
      { id: 19, label: '跨校上课转出', direction: 'out' },
      { id: 20, label: '结课', direction: 'out' },
      { id: 21, label: '到期结算', direction: 'out' },
      { id: 22, label: '退费', direction: 'out' },
      { id: 23, label: '订单作废', direction: 'out' },
      { id: 24, label: '作废跨校转入', direction: 'out' },
      { id: 25, label: '手动结课', direction: 'out' },
    ],
  },
]

const sourceTypeLabelMap = FLOW_TYPE_GROUPS.reduce((acc, group) => {
  group.children.forEach((child) => {
    acc[child.id] = child.label
  })
  return acc
}, {})

const sourceTypeDirectionMap = FLOW_TYPE_GROUPS.reduce((acc, group) => {
  group.children.forEach((child) => {
    acc[child.id] = child.direction
  })
  return acc
}, {})

const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const totalHeight = ref(window.innerHeight - 420)
const loading = ref(false)
const dataSource = ref([])
const pagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: false,
  showQuickJumper: false,
})

const tuitionAccountId = computed(() => String(props.record?.id || props.record?.tuitionAccountId || ''))
const lessonChargingMode = computed(() => Number(props.record?.lessonChargingMode || 0))
const quantityUnit = computed(() => LESSON_CHARGING_UNIT_MAP[lessonChargingMode.value] || '课时')
const lessonTypeText = computed(() => LESSON_TYPE_MAP[Number(props.record?.lessonType || 0)] || '-')
const lessonChargingModeText = computed(() => LESSON_CHARGING_MODE_MAP[lessonChargingMode.value] || '-')
const remainTuitionText = computed(() => `¥ ${formatMoney(getSummaryRemainTuition())}`)
const totalTuitionText = computed(() => `¥ ${formatMoney(props.record?.totalTuition || 0)}`)
const totalQuantityText = computed(() => `${formatCount(getSummaryTotalQuantity())}${quantityUnit.value}`)
const remainQuantityText = computed(() => `${formatCount(getSummaryRemainQuantity())}${quantityUnit.value}`)

const columns = [
  {
    title: '变动类型',
    dataIndex: 'type',
  },
  {
    title: '变动时间',
    dataIndex: 'time',
  },
  {
    title: '学费变动',
    dataIndex: 'fee',
  },
  {
    title: '数量变动',
    dataIndex: 'num',
  },
]

watch(
  () => [props.open, tuitionAccountId.value],
  async ([open]) => {
    if (!open)
      return
    pagination.value.current = 1
    await fetchData()
  },
  { immediate: false },
)

function closeFun() {
  openModal.value = false
}

function updateTableHeight() {
  totalHeight.value = window.innerHeight - 420
}

function formatMoney(value) {
  const num = Number(value || 0)
  return num.toLocaleString('zh-CN', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })
}

function formatCount(value) {
  const num = Number(value || 0)
  if (Number.isInteger(num))
    return String(num)
  return num.toFixed(2)
}

function getSummaryTotalQuantity() {
  const totalQuantity = Number(props.record?.totalQuantity || 0)
  const totalFreeQuantity = Number(props.record?.totalFreeQuantity || 0)
  return totalQuantity + totalFreeQuantity
}

function getSummaryRemainQuantity() {
  if (Number(props.record?.status || 0) === 3)
    return 0
  const remainQuantity = Number(props.record?.remainQuantity || 0)
  const remainFreeQuantity = Number(props.record?.remainFreeQuantity || 0)
  return remainQuantity + remainFreeQuantity
}

function getSummaryRemainTuition() {
  if (Number(props.record?.status || 0) === 3)
    return 0
  return Number(props.record?.tuition || 0)
}

function getSourceTypeText(sourceType) {
  return sourceTypeLabelMap[Number(sourceType || 0)] || '-'
}

function getChangePrefix(sourceType) {
  return sourceTypeDirectionMap[Number(sourceType || 0)] === 'out' ? '-' : '+'
}

function formatDateTime(value) {
  if (!value || `${value}`.startsWith('0001-01-01'))
    return '-'
  const parsed = dayjs(value)
  if (!parsed.isValid())
    return '-'
  return parsed.format('YYYY-MM-DD HH:mm')
}

function formatTuition(value, sourceType) {
  return `${getChangePrefix(sourceType)} ¥ ${formatMoney(Math.abs(Number(value || 0)))}`
}

function formatQuantity(value, sourceType) {
  return `${getChangePrefix(sourceType)} ${formatCount(Math.abs(Number(value || 0)))} ${quantityUnit.value}`
}

function normalizeRows(list = []) {
  return list.map(item => ({
    id: String(item?.id || ''),
    type: getSourceTypeText(item?.sourceType),
    time: formatDateTime(item?.createdTime),
    fee: formatTuition(item?.tuition, item?.sourceType),
    num: formatQuantity(item?.quantity, item?.sourceType),
  }))
}

async function fetchData() {
  if (!tuitionAccountId.value) {
    dataSource.value = []
    pagination.value.total = 0
    return
  }
  loading.value = true
  try {
    const res = await getSubTuitionAccountFlowRecordListApi({
      queryModel: {
        tuitionAccountId: tuitionAccountId.value,
      },
      pageRequestModel: {
        needTotal: true,
        pageSize: pagination.value.pageSize,
        pageIndex: pagination.value.current,
        skipCount: (pagination.value.current - 1) * pagination.value.pageSize,
      },
      sortModel: {
        orderByCreatedTime: 0,
      },
    })
    if (res.code !== 200) {
      throw new Error(res.message || '加载学费变动记录失败')
    }
    const list = Array.isArray(res.result?.list) ? res.result.list : []
    dataSource.value = normalizeRows(list)
    pagination.value.total = Number(res.result?.total || 0)
  }
  catch (error) {
    dataSource.value = []
    pagination.value.total = 0
    messageService.error(error?.message || '加载学费变动记录失败')
  }
  finally {
    loading.value = false
  }
}

async function handleTableChange(pag) {
  pagination.value.current = Number(pag?.current || 1)
  await fetchData()
}

onMounted(() => {
  window.addEventListener('resize', updateTableHeight)
})

onUnmounted(() => {
  window.removeEventListener('resize', updateTableHeight)
})
</script>

<template>
  <a-modal
    v-model:open="openModal" centered class="modal-content-box" :keyboard="false" :closable="false"
    :mask-closable="false" :width="800" :footer="false"
  >
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>学费变动记录</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <div class="contenter scrollbar">
      <div class="bg-#fff py-16px px-24px">
        <div class="top mb-16px flex justify-between flex-center">
          <div class="top-left flex flex-col">
            <span class="text-20px text-#222 font-500">{{ props.record?.lessonName || '-' }}</span>
            <a-space>
              <span class="bg-#e6f0ff text-#06f text-3 px3 py2px rounded-10">{{ lessonTypeText }}</span>
              <span class="bg-#e6f0ff text-#06f text-3 px3 py2px rounded-10">{{ lessonChargingModeText }}</span>
            </a-space>
          </div>
          <div class="top-right">
            <div class="custom-num-font-family text-20px text-#06f font-500">
              {{ remainTuitionText }}
            </div>
            <span class="text-12px text-#999 flex justify-end">剩余学费</span>
          </div>
        </div>
        <a-descriptions :column="3" :content-style="{ color: '#888' }">
          <a-descriptions-item label="总学费">
            {{ totalTuitionText }}
          </a-descriptions-item>
          <a-descriptions-item label="总课时">
            {{ totalQuantityText }}
          </a-descriptions-item>
          <a-descriptions-item label="剩余课时">
            {{ remainQuantityText }}
          </a-descriptions-item>
        </a-descriptions>
      </div>
      <div class="tables p-16px rounded-10">
        <div class="bg-#fff rounded-12px p-16px h-600px">
          <custom-title :title="`共 ${pagination.total} 条记录`" font-size="14px" class="mb-12px" />
          <a-spin :spinning="loading">
            <a-table
              row-key="id"
              :columns="columns"
              :data-source="dataSource"
              :pagination="pagination"
              :scroll="{ y: totalHeight }"
              @change="handleTableChange"
            />
          </a-spin>
        </div>
      </div>
    </div>
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
  background: #f7f6fd;
  border-radius: 14px;
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
