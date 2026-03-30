<script setup>
import { computed, ref, watch } from 'vue'
import { CloseOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import { getTuitionAccountSubAccountDateInfoApi } from '@/api/edu-center/tuition-account'
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

const infoLoading = ref(false)
const subAccountList = ref([])

const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

watch(
  () => props.open,
  async (value) => {
    if (!value)
      return
    await loadSubAccountInfo()
  },
)

const lessonChargingMode = computed(() => Number(props.record?.lessonChargingMode || 0))

const courseName = computed(() => props.record?.lessonName || props.record?.productName || '-')

const summaryTags = computed(() => {
  const tags = []
  const lessonType = Number(props.record?.lessonType || 0)
  if (lessonType === 1)
    tags.push('班级授课')
  else if (lessonType === 2)
    tags.push('1v1授课')

  if (lessonChargingMode.value === 1)
    tags.push('课时')
  else if (lessonChargingMode.value === 2)
    tags.push('时段')
  else if (lessonChargingMode.value === 3)
    tags.push('金额')
  return tags
})

const remainQuantityText = computed(() => {
  if (lessonChargingMode.value === 2) {
    const paidDays = subAccountList.value
      .filter(item => !item?.isFree)
      .reduce((sum, item) => sum + Number(item?.remainDays || 0), 0)
    const freeDays = subAccountList.value
      .filter(item => item?.isFree)
      .reduce((sum, item) => sum + Number(item?.remainDays || 0), 0)
    if (paidDays > 0 && freeDays > 0)
      return `${formatCount(paidDays)}天 + 赠${formatCount(freeDays)}天`
    if (paidDays > 0)
      return `${formatCount(paidDays)}天`
    if (freeDays > 0)
      return `赠${formatCount(freeDays)}天`
    return `${formatCount(Number(props.record?.remainQuantity || 0) + Number(props.record?.remainFreeQuantity || 0))}天`
  }
  return `${formatCount(Number(props.record?.remainQuantity || 0) + Number(props.record?.remainFreeQuantity || 0))}${lessonChargingMode.value === 3 ? '元' : '课时'}`
})

const remainTuitionText = computed(() => `¥ ${formatMoney(props.record?.tuition || 0)}`)

const originalValidityText = computed(() => {
  if (lessonChargingMode.value !== 2)
    return formatDate(props.record?.expireTime)
  const lines = subAccountList.value
    .filter(item => Number(item?.remainDays || 0) > 0)
    .map((item) => {
      const start = formatDate(item?.startDate || item?.activedAt)
      const end = formatDate(item?.endDate)
      if (start === '-' || end === '-')
        return ''
      return `${start} ~ ${end}`
    })
    .filter(Boolean)
  return Array.from(new Set(lines)).join('，') || '-'
})

const suspendDateText = computed(() => {
  const suspendDate = props.record?.changeStatusTime || props.record?.suspendedTime || props.record?.planSuspendTime
  const date = formatDate(suspendDate)
  if (date === '-')
    return '-'
  const diffDays = Math.max(dayjs().startOf('day').diff(dayjs(suspendDate).startOf('day'), 'day'), 0)
  return `${date}（至今已停课${diffDays}天）`
})

const suspendRemarkText = computed(() => '-')

async function loadSubAccountInfo() {
  const tuitionAccountId = String(props.record?.id || props.record?.tuitionAccountId || '')
  if (!tuitionAccountId) {
    subAccountList.value = []
    return
  }
  infoLoading.value = true
  try {
    const res = await getTuitionAccountSubAccountDateInfoApi({ tuitionAccountId })
    if (res.code !== 200)
      throw new Error(res.message || '加载账期明细失败')
    subAccountList.value = Array.isArray(res.result?.list) ? res.result.list : []
  }
  catch (error) {
    subAccountList.value = []
    messageService.error(error?.message || '加载账期明细失败')
  }
  finally {
    infoLoading.value = false
  }
}

function formatDate(value) {
  if (!value || `${value}`.startsWith('0001-01-01'))
    return '-'
  const parsed = dayjs(value)
  if (!parsed.isValid())
    return '-'
  return parsed.format('YYYY-MM-DD')
}

function formatMoney(value) {
  return Number(value || 0).toLocaleString('zh-CN', {
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

function closeFun() {
  openModal.value = false
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
    :width="800"
    :destroy-on-close="true"
  >
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>复课</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <div class="contenter scrollbar">
      <a-spin :spinning="infoLoading">
        <div class="resume-card">
          <div class="text-20px font-800 mb-4px">
            {{ courseName }}
          </div>
          <a-space v-if="summaryTags.length">
            <span
              v-for="tag in summaryTags"
              :key="tag"
              class="bg-#e6f0ff text-#06f text-3 px3 py2px rounded-10"
            >
              {{ tag }}
            </span>
          </a-space>
          <a-descriptions class="mt-20px" :column="2" size="small" :content-style="{ color: '#888' }">
            <a-descriptions-item label="剩余天数">
              {{ remainQuantityText }}
            </a-descriptions-item>
            <a-descriptions-item label="剩余学费">
              {{ remainTuitionText }}
            </a-descriptions-item>
            <a-descriptions-item label="原有效时段" :span="2">
              {{ originalValidityText }}
            </a-descriptions-item>
            <a-descriptions-item label="停课日期">
              {{ suspendDateText }}
            </a-descriptions-item>
            <a-descriptions-item label="停课备注" :span="2">
              {{ suspendRemarkText }}
            </a-descriptions-item>
          </a-descriptions>
          <a-divider class="my-16px" />
          <a-form :label-col="{ span: 4 }" :wrapper-col="{ span: 18 }">
            <a-form-item label="复课日期" required>
              <a-date-picker value-format="YYYY-MM-DD" class="w-200px" placeholder="请选择日期" />
            </a-form-item>
            <a-form-item label="复课备注">
              <a-input placeholder="请输入" />
            </a-form-item>
          </a-form>
        </div>
      </a-spin>
    </div>
    <template #footer>
      <a-button @click="closeFun">
        关闭
      </a-button>
      <a-button type="primary">
        确定
      </a-button>
    </template>
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
  margin: 24px;
  border-radius: 14px;
  background: #fafafa;
}

.resume-card {
  padding: 24px;
  border-radius: 14px;
  background: #fff;
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
