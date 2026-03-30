<script setup>
import { CloseOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import revokeCloseCourseModal from '@/components/common/revoke-close-course-modal.vue'
import { getCloseTuitionAccountOrderListApi } from '@/api/edu-center/tuition-account'
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

const emit = defineEmits(['update:open', 'success'])

const loading = ref(false)
const dataSource = ref([])
const revokeModalOpen = ref(false)
const revokeRecord = ref(null)

const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const columns = [
  { title: '结课时间', dataIndex: 'createdTime', key: 'createdTime', width: 160 },
  { title: '结课数量', dataIndex: 'quantity', key: 'quantity', width: 110 },
  { title: '状态', dataIndex: 'status', key: 'status', width: 110 },
  { title: '更新人', dataIndex: 'updatedStaffName', key: 'updatedStaffName', width: 110 },
  { title: '更新时间', dataIndex: 'updatedTime', key: 'updatedTime', width: 160 },
  { title: '操作', dataIndex: 'action', key: 'action', width: 110 },
]

const quantityUnit = computed(() => {
  const mode = Number(props.record?.lessonChargingMode || 0)
  if (mode === 2)
    return '天'
  if (mode === 3)
    return '元'
  return '课时'
})

const latestClosableOrderId = computed(() =>
  String(dataSource.value.find(item => Number(item?.status) !== 4)?.id || ''),
)

watch(
  () => props.open,
  async (value) => {
    if (value) {
      await loadCloseRecords()
      return
    }
    closeRevokeModal()
  },
)

function closeFun() {
  openModal.value = false
}

function closeRevokeModal() {
  revokeModalOpen.value = false
  revokeRecord.value = null
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

function getStatusMeta(status) {
  if (Number(status) === 4) {
    return {
      text: '已撤销',
      dotClass: 'bg-#bfbfbf',
      textClass: 'text-#8c8c8c',
    }
  }
  return {
    text: '正常',
    dotClass: 'bg-#1677ff',
    textClass: 'text-#262626',
  }
}

function canRevoke(record) {
  return Number(record?.status) !== 4 && String(record?.id || '') === latestClosableOrderId.value
}

function handleRevoke(record) {
  if (!canRevoke(record))
    return
  revokeRecord.value = props.record || {}
  revokeModalOpen.value = true
}

async function handleRevokeSuccess(payload) {
  closeRevokeModal()
  await loadCloseRecords()
  emit('success', payload)
}

async function loadCloseRecords() {
  const tuitionAccountId = String(props.record?.id || props.record?.tuitionAccountId || '')
  if (!tuitionAccountId) {
    dataSource.value = []
    return
  }
  loading.value = true
  try {
    const res = await getCloseTuitionAccountOrderListApi({ tuitionAccountId })
    if (res.code !== 200) {
      throw new Error(res.message || '加载结课记录失败')
    }
    dataSource.value = Array.isArray(res.result?.list) ? res.result.list : []
  }
  catch (error) {
    dataSource.value = []
    messageService.error(error?.message || '加载结课记录失败')
  }
  finally {
    loading.value = false
  }
}
</script>

<template>
  <a-modal
    v-model:open="openModal"
    centered
    :keyboard="false"
    :closable="false"
    :mask-closable="false"
    :width="900"
    :body-style="{ padding: '0' }"
    :footer="false"
  >
    <template #title>
      <div class="flex items-center justify-between text-4.5 ">
        <span>结课记录</span>
        <a-button type="text" class="!w-9 !h-9" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-4.5" />
          </template>
        </a-button>
      </div>
    </template>

    <div class="close-record-modal">
      <div class="close-record-panel">
        <a-table
          :columns="columns"
          :data-source="dataSource"
          :loading="loading"
          :pagination="false"
          row-key="id"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'createdTime'">
              {{ formatDateTime(record.createdTime) }}
            </template>
            <template v-else-if="column.key === 'quantity'">
              {{ formatCount(Number(record.quantity || 0) + Number(record.freeQuantity || 0)) }}{{ quantityUnit }}
            </template>
            <template v-else-if="column.key === 'status'">
              <span
                class="inline-flex items-center gap-6px text-14px"
                :class="getStatusMeta(record.status).textClass"
              >
                <span class="inline-block w-6px h-6px rounded-full" :class="getStatusMeta(record.status).dotClass" />
                {{ getStatusMeta(record.status).text }}
              </span>
            </template>
            <template v-else-if="column.key === 'updatedTime'">
              {{ formatDateTime(record.updatedTime) }}
            </template>
            <template v-else-if="column.key === 'action'">
              <a-button
                v-if="canRevoke(record)"
                type="link"
                size="small"
                class="!px-0"
                @click="handleRevoke(record)"
              >
                撤销结课
              </a-button>
              <span v-else class="text-#bfbfbf">-</span>
            </template>
          </template>
        </a-table>
      </div>

      <div class="close-record-footer">
        <a-button @click="closeFun">
          取消
        </a-button>
      </div>
    </div>

    <revokeCloseCourseModal
      v-model:open="revokeModalOpen"
      :record="revokeRecord"
      @success="handleRevokeSuccess"
    />
  </a-modal>
</template>

<style lang="less" scoped>
.close-record-modal {
  display: flex;
  min-height: 560px;
  flex-direction: column;
}

.close-record-panel {
  margin: 18px 20px 0;
  flex: 1;
  overflow: hidden;
  background: #fafbfc;
}

.close-record-footer {
  display: flex;
  justify-content: flex-end;
  border-top: 1px solid #f0f0f0;
  padding: 12px 20px 16px;
}

</style>
