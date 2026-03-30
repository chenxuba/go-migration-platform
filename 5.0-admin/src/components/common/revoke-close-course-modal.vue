<script setup>
import { CloseOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'

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

const emit = defineEmits(['update:open', 'confirm'])

const formRef = ref()

const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const formState = reactive({
  currentValidStartDate: undefined,
})

watch(
  () => props.open,
  (value) => {
    if (value) {
      formState.currentValidStartDate = undefined
    }
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

const courseName = computed(() =>
  props.record?.lessonName || props.record?.productName || '时段课时',
)

const remainQuantityLabel = computed(() => {
  const mode = Number(props.record?.lessonChargingMode || 2)
  if (mode === 2)
    return '剩余天数'
  if (mode === 3)
    return '剩余金额'
  return '剩余课时'
})

const remainQuantityText = computed(() => {
  const mode = Number(props.record?.lessonChargingMode || 2)
  if (props.record?.remainQuantity !== undefined && props.record?.remainQuantity !== null) {
    const unit = getQuantityUnit(mode)
    return unit ? `${formatCount(props.record.remainQuantity)} ${unit}` : formatCount(props.record.remainQuantity)
  }
  return '9 天'
})

const remainTuitionText = computed(() => {
  if (props.record?.tuition !== undefined && props.record?.tuition !== null)
    return `¥${formatMoney(props.record.tuition)}`
  return '¥1800.00'
})

const validPeriodText = computed(() => {
  const start = formatDate(props.record?.validDate || props.record?.activedAt)
  const end = formatDate(props.record?.endDate || props.record?.expireTime)
  if (start !== '-' && end !== '-')
    return `${start} ~ ${end}`
  return '2026-03-29 ~ 2026-03-29'
})

const closeDateText = computed(() => {
  const date = formatDate(props.record?.classEndingTime)
  return date !== '-' ? date : '2026-03-30'
})

const closeRemarkText = computed(() => {
  const text = String(props.record?.closeRemark || props.record?.remark || '').trim()
  return text || '-'
})

const summaryTags = computed(() => [
  { text: getLessonTypeText(props.record?.lessonType), type: 'soft' },
  { text: getChargingModeText(props.record?.lessonChargingMode), type: 'soft' },
])

async function handleSubmit() {
  try {
    await formRef.value?.validate?.()
    emit('confirm', {
      currentValidStartDate: formState.currentValidStartDate,
      record: props.record,
    })
    openModal.value = false
  }
  catch (error) {
    console.log('验证失败:', error)
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
      class="border-none rounded-none text-#06f bg-#e6f0ff"
    />

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
            :color="tag.type === 'primary' ? '#0066ff' : '#e6f0ff'"
            :style="{ color: tag.type === 'primary' ? '#fff' : '#0066ff', borderRadius: '999px', marginRight: '0' }"
          >
            {{ tag.text }}
          </a-tag>
        </a-space>

        <a-descriptions
          class="mt-24px"
          :column="2"
          size="small"
          :label-style="{ color: '#555'}"
          :content-style="{ color: '#666' }"
        >
          <a-descriptions-item :label="remainQuantityLabel">
            {{ remainQuantityText }}
          </a-descriptions-item>
          <a-descriptions-item label="剩余学费">
            {{ remainTuitionText }}
          </a-descriptions-item>
          <a-descriptions-item label="有效时段">
            {{ validPeriodText }}
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
        <a-form ref="formRef" :model="formState"  :wrapper-col="{ span: 17 }">
          <a-form-item
            label="现有效开始时间"
            name="currentValidStartDate"
            :rules="[{ required: true, message: '请选择日期' }]"
          >
            <a-date-picker
              v-model:value="formState.currentValidStartDate"
              value-format="YYYY-MM-DD"
              placeholder="请选择日期"
              class="!w-240px"
            />
          </a-form-item>
        </a-form>
      </div>
    </div>

    <template #footer>
      <a-button @click="closeFun">
        关闭
      </a-button>
      <a-button type="primary" @click="handleSubmit">
        确定
      </a-button>
    </template>
  </a-modal>
</template>
