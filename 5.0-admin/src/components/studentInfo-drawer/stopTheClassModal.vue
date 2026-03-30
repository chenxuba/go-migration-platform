<script setup>
import { computed, reactive, ref, watch } from 'vue'
import { CloseOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import { addSuspendResumeTuitionAccountOrderApi } from '@/api/edu-center/tuition-account'
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
const submitLoading = ref(false)

const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const formState = reactive({
  plannedSuspensionDate: undefined,
  plannedResumptionDate: undefined,
  remark: '',
})

watch(
  () => props.open,
  (value) => {
    if (!value)
      return
    formState.plannedSuspensionDate = undefined
    formState.plannedResumptionDate = undefined
    formState.remark = ''
  },
)

const lessonChargingMode = computed(() => Number(props.record?.lessonChargingMode || 0))

const quantityLabel = computed(() => {
  if (lessonChargingMode.value === 2)
    return '剩余天数'
  if (lessonChargingMode.value === 3)
    return '剩余金额'
  return '剩余课时'
})

const validityLabel = computed(() => (
  lessonChargingMode.value === 2 ? '有效时段' : '有效期至'
))

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
  const remainQuantity = Number(props.record?.remainQuantity || 0)
  const remainFreeQuantity = Number(props.record?.remainFreeQuantity || 0)
  const total = remainQuantity + remainFreeQuantity
  if (lessonChargingMode.value === 2)
    return `${formatCount(total)}天`
  if (lessonChargingMode.value === 3)
    return `${formatCount(total)}元`
  return `${formatCount(total)}课时`
})

const remainTuitionText = computed(() => `¥ ${formatMoney(props.record?.tuition || 0)}`)

const validityText = computed(() => {
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

function disabledResumeDate(current) {
  if (!current)
    return false
  if (current.endOf('day').isBefore(dayjs().startOf('day')))
    return true
  if (!formState.plannedSuspensionDate)
    return false
  return current.endOf('day').isBefore(dayjs(formState.plannedSuspensionDate).startOf('day'))
}

function toPayloadDateTime(value) {
  if (!value)
    return ''
  const parsed = dayjs(value).startOf('day')
  if (!parsed.isValid())
    return ''
  return parsed.format('YYYY-MM-DDTHH:mm:ssZ')
}

function closeFun() {
  formRef.value?.resetFields?.()
  formState.remark = ''
  openModal.value = false
}

async function handleSubmit() {
  try {
    await formRef.value?.validate?.()
    if (formState.plannedResumptionDate && formState.plannedSuspensionDate
      && dayjs(formState.plannedResumptionDate).isBefore(dayjs(formState.plannedSuspensionDate), 'day')) {
      messageService.error('计划复课日期不能早于计划停课日期')
      return
    }

    const tuitionAccountId = String(props.record?.id || props.record?.tuitionAccountId || '')
    if (!tuitionAccountId) {
      messageService.error('缺少学费账户ID')
      return
    }

    submitLoading.value = true
    const res = await addSuspendResumeTuitionAccountOrderApi({
      tuitionAccountId,
      type: 1,
      expireTime: toPayloadDateTime(props.record?.endDate || props.record?.expireTime),
      expireType: 0,
      remark: formState.remark?.trim() || '',
      suspendDate: toPayloadDateTime(formState.plannedSuspensionDate),
      resumeDate: toPayloadDateTime(formState.plannedResumptionDate),
    })
    if (res.code !== 200)
      throw new Error(res.message || '停课失败')

    messageService.success('停课设置成功')
    emit('success', {
      result: res.result,
      record: props.record,
    })
    closeFun()
  }
  catch (error) {
    if (error?.errorFields)
      return
    messageService.error(error?.message || '停课失败')
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
    class="modal-content-box"
    :keyboard="false"
    :closable="false"
    :mask-closable="false"
    :width="800"
    :destroy-on-close="true"
  >
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>停课</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <a-alert message="停课后，学员报读课程将停止计费，且无法进行点名操作。" show-icon type="info" class="text-#06f border-none bg-#e6f0ff" />
    <div class="contenter scrollbar">
      <a-form ref="formRef" :model="formState" :label-col="{ span: 4 }" :wrapper-col="{ span: 18 }">
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
          <a-descriptions-item :label="quantityLabel">
            {{ remainQuantityText }}
          </a-descriptions-item>
          <a-descriptions-item label="剩余学费">
            {{ remainTuitionText }}
          </a-descriptions-item>
          <a-descriptions-item :label="validityLabel">
            {{ validityText }}
          </a-descriptions-item>
        </a-descriptions>
        <a-divider class="my-16px" />
        <a-form-item label="计划停课日期" name="plannedSuspensionDate" :rules="[{ required: true, message: '请选择计划停课日期' }]">
          <a-date-picker
            v-model:value="formState.plannedSuspensionDate"
            value-format="YYYY-MM-DD"
            class="w-200px"
            :disabled-date="disabledPastDate"
          />
        </a-form-item>
        <a-form-item label="停课备注">
          <a-input v-model:value="formState.remark" placeholder="选填" />
        </a-form-item>
        <a-form-item label="计划复课日期">
          <a-date-picker
            v-model:value="formState.plannedResumptionDate"
            value-format="YYYY-MM-DD"
            class="w-200px"
            :disabled-date="disabledResumeDate"
          />
        </a-form-item>
      </a-form>
    </div>
    <template #footer>
      <a-button danger ghost @click="closeFun">
        关闭
      </a-button>
      <a-button type="primary" ghost :loading="submitLoading" @click="handleSubmit">
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
