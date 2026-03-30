<script setup>
import { CloseOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import {
  getRevertCloseTuitionAccountPreviewApi,
  getTuitionAccountSubAccountDateInfoApi,
  revertCloseTuitionAccountApi,
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

const emit = defineEmits(['update:open', 'success'])

const formRef = ref()
const loading = ref(false)
const submitLoading = ref(false)
const previewData = ref(null)
const subAccountList = ref([])

const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const formState = reactive({
  startDate: undefined,
})

const tuitionAccountId = computed(() => String(props.record?.id || props.record?.tuitionAccountId || ''))

watch(
  () => props.open,
  async (value) => {
    if (!value)
      return
    formState.startDate = undefined
    await loadPreviewData()
  },
)

function closeFun() {
  formRef.value?.resetFields?.()
  openModal.value = false
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

function disabledPastDate(current) {
  if (!current)
    return false
  return current.endOf('day').isBefore(dayjs().startOf('day'))
}

function getLessonTypeText(type) {
  const typeMap = {
    1: '班级授课',
    2: '1对1授课',
  }
  return typeMap[type] || '1对1授课'
}

function getChargingModeText(mode) {
  const modeMap = {
    1: '按课时',
    2: '按时段',
    3: '按金额',
  }
  return modeMap[mode] || '按时段'
}

function getQuantityUnit(mode) {
  const unitMap = {
    1: '课时',
    2: '天',
    3: '元',
  }
  return unitMap[mode] || '天'
}

async function loadPreviewData() {
  if (!tuitionAccountId.value) {
    previewData.value = null
    subAccountList.value = []
    return
  }
  loading.value = true
  try {
    const [previewRes, subAccountRes] = await Promise.all([
      getRevertCloseTuitionAccountPreviewApi({ tuitionAccountId: tuitionAccountId.value }),
      getTuitionAccountSubAccountDateInfoApi({ tuitionAccountId: tuitionAccountId.value }),
    ])
    if (previewRes.code !== 200)
      throw new Error(previewRes.message || '加载撤销结课预览失败')
    if (subAccountRes.code !== 200)
      throw new Error(subAccountRes.message || '加载账期明细失败')

    previewData.value = previewRes.result || null
    subAccountList.value = Array.isArray(subAccountRes.result?.list) ? subAccountRes.result.list : []

    formState.startDate = undefined
  }
  catch (error) {
    previewData.value = null
    subAccountList.value = []
    messageService.error(error?.message || '加载撤销结课预览失败')
  }
  finally {
    loading.value = false
  }
}

const lessonChargingMode = computed(() =>
  Number(previewData.value?.lessonChargingMode || props.record?.lessonChargingMode || 0),
)

const startDateRules = computed(() => (
  lessonChargingMode.value === 2 ? [{ required: true, message: '请选择日期' }] : []
))

const courseName = computed(() =>
  previewData.value?.lessonName || props.record?.lessonName || props.record?.productName || '时段课时',
)

const remainQuantityLabel = computed(() => {
  if (lessonChargingMode.value === 2)
    return '剩余天数'
  if (lessonChargingMode.value === 3)
    return '剩余金额'
  return '剩余课时'
})

const remainQuantityText = computed(() => {
  if (previewData.value) {
    const quantity = Number(previewData.value?.quantity || 0)
    const freeQuantity = Number(previewData.value?.freeQuantity || 0)
    const totalQuantity = quantity + freeQuantity
    const unit = getQuantityUnit(lessonChargingMode.value)
    if (lessonChargingMode.value === 2 && freeQuantity > 0) {
      if (quantity > 0)
        return `${formatCount(quantity)}${unit} + 赠${formatCount(freeQuantity)}${unit}`
      return `赠${formatCount(freeQuantity)}${unit}`
    }
    return unit ? `${formatCount(totalQuantity)} ${unit}` : formatCount(totalQuantity)
  }
  return '-'
})

const remainTuitionText = computed(() => {
  if (previewData.value?.tuition !== undefined && previewData.value?.tuition !== null)
    return `¥${formatMoney(previewData.value.tuition)}`
  return '-'
})

const validityLabel = computed(() => (
  lessonChargingMode.value === 2 ? '有效时段' : '有效期至'
))

const currentValidFieldLabel = computed(() => (
  lessonChargingMode.value === 2 ? '现有效开始时间' : '现有效期至'
))

const previewPeriods = computed(() =>
  Array.isArray(previewData.value?.subTuitionAccounts) ? previewData.value.subTuitionAccounts : [],
)

const recomputedValidPeriodLines = computed(() => {
  if (lessonChargingMode.value !== 2 || !formState.startDate)
    return []

  const start = dayjs(formState.startDate)
  if (!start.isValid())
    return []

  let cursor = start
  const lines = []

  previewPeriods.value.forEach((period) => {
    let days = Math.round(Number(period?.quantity || 0))
    if (days <= 0) {
      const rawStart = dayjs(period?.startDate)
      const rawEnd = dayjs(period?.endDate)
      if (rawStart.isValid() && rawEnd.isValid())
        days = rawEnd.diff(rawStart, 'day') + 1
    }
    if (days <= 0)
      return

    const end = cursor.add(days - 1, 'day')
    lines.push(`${cursor.format('YYYY-MM-DD')} ~ ${end.format('YYYY-MM-DD')}`)
    cursor = end.add(1, 'day')
  })

  return lines
})

const validPeriodLines = computed(() => {
  if (lessonChargingMode.value === 1)
    return ['无']

  if (recomputedValidPeriodLines.value.length)
    return recomputedValidPeriodLines.value

  const lines = previewPeriods.value
    .map((period) => {
      const start = formatDate(period?.startDate)
      const end = formatDate(period?.endDate)
      if (start === '-' && end === '-')
        return ''
      return `${start} ~ ${end}`
    })
    .filter(Boolean)

  const uniqueLines = Array.from(new Set(lines))

  if (uniqueLines.length)
    return uniqueLines

  const fallbackLines = (Array.isArray(subAccountList.value) ? subAccountList.value : [])
    .map((item) => {
      const start = formatDate(item?.startDate || item?.activedAt)
      const end = formatDate(item?.endDate)
      if (start === '-' && end === '-')
        return ''
      return `${start} ~ ${end}`
    })
    .filter(Boolean)

  if (fallbackLines.length)
    return fallbackLines

  const start = formatDate(props.record?.validDate || props.record?.activedAt)
  const end = formatDate(props.record?.endDate || props.record?.expireTime)
  if (start !== '-' || end !== '-')
    return [`${start} ~ ${end}`]

  return ['-']
})

const closeDateText = computed(() => {
  const date = formatDate(previewData.value?.closeTime || props.record?.classEndingTime)
  return date !== '-' ? date : '-'
})

const closeRemarkText = computed(() => {
  const text = String(previewData.value?.remark || props.record?.closeRemark || props.record?.remark || '').trim()
  return text || '-'
})

const summaryTags = computed(() => [
  { text: getLessonTypeText(previewData.value?.lessonType || props.record?.lessonType), color: '#e6f0ff', textColor: '#0066ff' },
  { text: getChargingModeText(lessonChargingMode.value), color: '#e6f0ff', textColor: '#0066ff' },
])

async function handleSubmit() {
  try {
    await formRef.value?.validate?.()
    if (!tuitionAccountId.value) {
      messageService.error('缺少学费账户ID')
      return
    }
    if (!previewData.value?.closeTuitionAccountOrderId) {
      messageService.error('缺少结课记录ID')
      return
    }

    let startDate
    if (lessonChargingMode.value === 2) {
      startDate = formState.startDate
    }

    submitLoading.value = true
    const res = await revertCloseTuitionAccountApi({
      tuitionAccountId: tuitionAccountId.value,
      closeTuitionAccountOrderId: String(previewData.value.closeTuitionAccountOrderId),
      startDate,
    })
    if (res.code !== 200)
      throw new Error(res.message || '撤销结课失败')

    messageService.success('撤销结课成功')
    emit('success', {
      tuitionAccountId: tuitionAccountId.value,
      closeTuitionAccountOrderId: String(previewData.value.closeTuitionAccountOrderId),
      result: res.result,
      record: props.record,
    })
    openModal.value = false
  }
  catch (error) {
    if (error?.errorFields)
      return
    messageService.error(error?.message || '撤销结课失败')
  }
  finally {
    submitLoading.value = false
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
    :width="800"
    :body-style="{ padding: '0' }"
  >
    <template #title>
      <div class="flex items-center justify-between text-5">
        <span>撤销结课</span>
        <a-button type="text" class="!w-9 !h-9" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5" />
          </template>
        </a-button>
      </div>
    </template>

    <a-alert
      message="撤销结课后，学员报读课程将恢复计费，并可为其进行点名操作。"
      show-icon
      type="info"
      class="border-none rounded-none text-#06f bg-#e6f0ff mt--8px"
    />

    <a-spin :spinning="loading">
      <div class="m-24px rounded-12px bg-#fafafa overflow-hidden">
        <div class="px-24px pt-28px pb-22px">
          <div class="text-22px leading-7 text-#1f1f1f font-700">
            {{ courseName }}
          </div>
          <a-space :size="[8, 8]" wrap class="mt-6px">
            <a-tag
              v-for="tag in summaryTags"
              :key="tag.text"
              :bordered="false"
              :color="tag.color"
              :style="{ color: tag.textColor, borderRadius: '999px', marginRight: '0' }"
            >
              {{ tag.text }}
            </a-tag>
          </a-space>

          <a-descriptions
            class="mt-24px"
            :column="2"
            size="small"
            :label-style="{ color: '#555' }"
            :content-style="{ color: '#666' }"
          >
            <a-descriptions-item :label="remainQuantityLabel">
              {{ remainQuantityText }}
            </a-descriptions-item>
            <a-descriptions-item label="剩余学费">
              {{ remainTuitionText }}
            </a-descriptions-item>
            <a-descriptions-item :label="validityLabel">
              <div class="flex flex-col gap-4px">
                <span v-for="(line, index) in validPeriodLines" :key="`period-${index}`">
                  {{ line }}
                </span>
              </div>
            </a-descriptions-item>
            <a-descriptions-item label="结课日期">
              {{ closeDateText }}
            </a-descriptions-item>
            <a-descriptions-item label="结课备注" :span="2">
              {{ closeRemarkText }}
            </a-descriptions-item>
          </a-descriptions>
        </div>

        <a-divider class="my-0" />

        <div class="px-24px pt-26px pb-8px">
          <a-form ref="formRef" :model="formState" :wrapper-col="{ span: 17 }">
            <a-form-item
              :label="currentValidFieldLabel"
              name="startDate"
              :rules="startDateRules"
            >
              <a-date-picker
                v-model:value="formState.startDate"
                value-format="YYYY-MM-DD"
                placeholder="请选择日期"
                class="!w-240px"
                :disabled-date="disabledPastDate"
              />
            </a-form-item>
          </a-form>
        </div>
      </div>
    </a-spin>

    <template #footer>
      <a-button @click="closeFun">
        关闭
      </a-button>
      <a-button type="primary" :loading="submitLoading" @click="handleSubmit">
        确定
      </a-button>
    </template>
  </a-modal>
</template>
