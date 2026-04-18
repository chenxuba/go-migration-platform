<script setup>
import { computed, createVNode, reactive, ref, watch } from 'vue'
import { CloseOutlined, QuestionCircleFilled } from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'
import dayjs from 'dayjs'
import { addCloseTuitionAccountOrderApi } from '@/api/edu-center/tuition-account'
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
  endTheClassDate: '',
  remark: '',
})

watch(
  () => props.open,
  (value) => {
    if (!value)
      return
    formState.endTheClassDate = dayjs().format('YYYY-MM-DD')
    formState.remark = ''
  },
)

const lessonChargingMode = computed(() => Number(props.record?.lessonChargingMode || 0))

const quantityLabel = computed(() => {
  if (lessonChargingMode.value === 2)
    return '剩余天数'
  if (lessonChargingMode.value === 3)
    return '剩余数量'
  return '剩余课时'
})

const validityLabel = computed(() => (
  lessonChargingMode.value === 2 ? '有效时段' : '有效期至'
))

const titleText = computed(() => props.record?.lessonName || props.record?.productName || '-')

const summaryTags = computed(() => {
  const tags = []
  const lessonType = Number(props.record?.lessonType || 0)
  if (lessonType === 1)
    tags.push('班级授课')
  else if (lessonType === 2)
    tags.push('1对1授课')

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

const remainTuitionText = computed(() => {
  const remainTuition = Number(props.record?.tuition ?? props.record?.remainTuition ?? 0)
  return `¥ ${formatMoney(remainTuition)}`
})

const validityText = computed(() => {
  if (lessonChargingMode.value === 2) {
    const start = formatDate(props.record?.validDate || props.record?.activedAt)
    const end = formatDate(props.record?.endDate || props.record?.expireTime)
    if (start !== '-' && end !== '-')
      return `${start} ~ ${end}`
    return '-'
  }
  if (!props.record?.enableExpireTime)
    return '不限制'
  return formatDate(props.record?.expireTime)
})

const endDateDisplayText = computed(() => {
  if (!formState.endTheClassDate)
    return '-'
  const parsed = dayjs(formState.endTheClassDate)
  if (!parsed.isValid())
    return formState.endTheClassDate
  return parsed.isSame(dayjs(), 'day')
    ? `${formState.endTheClassDate}（今天）`
    : formState.endTheClassDate
})

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

function formatDate(value) {
  if (!value || `${value}`.startsWith('0001-01-01'))
    return '-'
  const parsed = dayjs(value)
  if (!parsed.isValid())
    return '-'
  return parsed.format('YYYY-MM-DD')
}

function closeFun() {
  formRef.value?.resetFields?.()
  formState.remark = ''
  openModal.value = false
}

async function submitCloseCourse() {
  const tuitionAccountId = String(props.record?.id || props.record?.tuitionAccountId || '')
  if (!tuitionAccountId) {
    messageService.error('缺少学费账户ID')
    return
  }

  const quantity = Number(props.record?.remainQuantity || 0)
  const freeQuantity = Number(props.record?.remainFreeQuantity || 0)
  const tuition = Number(props.record?.tuition ?? props.record?.remainTuition ?? 0)
  if (quantity + freeQuantity <= 0 && tuition <= 0) {
    messageService.error('当前无可结课的剩余课时或学费')
    return
  }

  submitLoading.value = true
  try {
    const res = await addCloseTuitionAccountOrderApi({
      tuitionAccountId,
      quantity,
      freeQuantity,
      tuition,
      remark: formState.remark?.trim() || '',
    })
    if (res.code !== 200)
      throw new Error(res.message || '结课失败')

    messageService.success('结课成功')
    emit('success', {
      tuitionAccountId,
      result: res.result,
      record: props.record,
    })
    closeFun()
  }
  catch (error) {
    messageService.error(error?.message || '结课失败')
    throw error
  }
  finally {
    submitLoading.value = false
  }
}

async function handleSubmit() {
  try {
    await formRef.value?.validate?.()
    Modal.confirm({
      title: '确定结课?',
      centered: true,
      closable: false,
      maskClosable: false,
      keyboard: false,
      icon: createVNode(QuestionCircleFilled, { style: { color: '#ff4d4f', fontSize: '22px' } }),
      content: '结课后，学员报读课程的剩余课时将全部扣除，机构获得相应课消收入。',
      okText: '确定结课',
      cancelText: '再想想',
      okButtonProps: { danger: true, ghost: true },
      wrapClassName: 'end-course-pre-confirm-modal',
      onOk: submitCloseCourse,
    })
  }
  catch {
    // 无必填校验
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
    :destroy-on-close="true"
    :body-style="{ padding: '0' }"
  >
    <template #title>
      <div class="flex items-center justify-between text-5">
        <span>结课</span>
        <a-button type="text" class="!w-9 !h-9" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5" />
          </template>
        </a-button>
      </div>
    </template>

    <a-alert
      message="结课后，学员报读课程的剩余课时将全部扣除，机构获得相应课消收入。"
      show-icon
      type="warning"
      class="text-#f90 border-none rounded-0 bg-#fff5e6"
    />

    <div class="contenter scrollbar">
      <a-form ref="formRef" :model="formState" :label-col="{ span: 3 }" :wrapper-col="{ span: 20 }">
        <div class="text-20px font-800 mb-4px">
          {{ titleText }}
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
        <a-form-item label="结课日期" name="endTheClassDate">
          <span class="text-#ff3333">{{ endDateDisplayText }}</span>
        </a-form-item>
        <a-form-item label="结课备注" name="remark" class="!mb-0">
          <a-input v-model:value="formState.remark" placeholder="选填" />
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

.contenter {
  padding: 24px;
  margin: 24px;
  border-radius: 14px;
  background: #fafafa;
}
</style>

<style>
.end-course-pre-confirm-modal .ant-modal-content {
  border-radius: 20px;
  overflow: hidden;
}
</style>
