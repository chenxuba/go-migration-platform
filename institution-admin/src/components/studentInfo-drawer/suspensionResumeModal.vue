<script setup>
import { CloseOutlined } from '@ant-design/icons-vue'
import { Empty } from 'ant-design-vue'
import dayjs from 'dayjs'
import { getSuspendResumeTuitionAccountOrderListApi } from '@/api/edu-center/tuition-account'
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

const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const loading = ref(false)
const list = ref([])
const tuitionAccountId = computed(() => String(props.record?.id || props.record?.tuitionAccountId || ''))
const emptyImage = Empty.PRESENTED_IMAGE_SIMPLE

watch(
  () => [props.open, tuitionAccountId.value],
  async ([open]) => {
    if (!open) {
      list.value = []
      return
    }
    await fetchList()
  },
)

function closeFun() {
  openModal.value = false
}

function isResumeRecord(item) {
  return Number(item?.type || 0) === 2
}

function getRecordName(item) {
  return isResumeRecord(item) ? '复课' : '停课'
}

function getPrimaryDateLabel(item) {
  return isResumeRecord(item) ? '复课日期' : '停课日期'
}

function getPrimaryDateValue(item) {
  return formatDate(isResumeRecord(item) ? item?.resumeDate : item?.suspendDate)
}

function getRemarkLabel(item) {
  return isResumeRecord(item) ? '复课备注' : '停课备注'
}

function getRemarkText(item) {
  return item?.remark?.trim() || '-'
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

function getExpireText(item) {
  const lessonChargingMode = Number(props.record?.lessonChargingMode || 0)
  if (isResumeRecord(item)) {
    if (Number(item?.expireType || 0) === 3)
      return '不限制'
    if (Number(item?.expireType || 0) === 2)
      return '顺延'
  }

  if (lessonChargingMode === 2) {
    const start = formatDate(props.record?.validDate || props.record?.activedAt)
    const end = formatDate(item?.expireTime || props.record?.endDate || props.record?.expireTime)
    if (start !== '-' && end !== '-')
      return `${start} ~ ${end}`
  }

  const expireText = formatDate(item?.expireTime || props.record?.expireTime || props.record?.endDate)
  if (expireText !== '-')
    return expireText

  if (props.record?.enableExpireTime === false)
    return '不限制'

  return '-'
}

async function fetchList() {
  if (!tuitionAccountId.value) {
    list.value = []
    return
  }
  loading.value = true
  try {
    const res = await getSuspendResumeTuitionAccountOrderListApi({
      tuitionAccountId: tuitionAccountId.value,
    })
    if (res.code !== 200)
      throw new Error(res.message || '加载停/复课记录失败')
    list.value = Array.isArray(res.result?.list) ? res.result.list : []
  }
  catch (error) {
    list.value = []
    messageService.error(error?.message || '加载停/复课记录失败')
  }
  finally {
    loading.value = false
  }
}
</script>

<template>
  <a-modal
    v-model:open="openModal" centered class="modal-content-box" :keyboard="false" :closable="false"
    :mask-closable="false" :width="800" :footer="false"
  >
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>停/复课记录</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <div class="pb-2px">
      <a-spin :spinning="loading">
        <div v-if="list.length">
          <div v-for="item in list" :key="item.id" class="contenter scrollbar">
            <a-descriptions class="descriptions" :column="3" size="small" :content-style="{ color: '#888' }">
              <template #title>
                <div class="flex flex-items-center mb-8px">
                  <img
                    v-if="isResumeRecord(item)"
                    src="https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12185/static/recovery-course.29f340ff.svg"
                    class="w-14px h-14px mr-4px mt-1px" alt=""
                  >
                  <img
                    v-else
                    src="https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12185/static/stop-course.a6619221.svg"
                    class="w-14px h-14px mr-4px mt-1px" alt=""
                  >
                  <span class="text-14px">{{ getRecordName(item) }}</span>
                </div>
              </template>
              <a-descriptions-item :label="getPrimaryDateLabel(item)">
                {{ getPrimaryDateValue(item) }}
              </a-descriptions-item>
              <a-descriptions-item label="现有效期至">
                {{ getExpireText(item) }}
              </a-descriptions-item>
              <a-descriptions-item label="操作时间">
                {{ formatDateTime(item.createdTime) }}
              </a-descriptions-item>
              <a-descriptions-item label="操作人">
                {{ item.createdStaffName || '-' }}
              </a-descriptions-item>
              <a-descriptions-item :label="getRemarkLabel(item)">
                {{ getRemarkText(item) }}
              </a-descriptions-item>
            </a-descriptions>
          </div>
        </div>
        <a-empty v-else :image="emptyImage" description="暂无停/复课记录" class="py-40px" />
      </a-spin>
    </div>
  </a-modal>
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

.contenter {
  padding: 24px;
  margin:18px 24px;
  border-radius: 14px;
  background: #f6f7f8;

  :deep(.descriptions .ant-descriptions-header) {
    margin-bottom: 0;
  }
}
</style>
