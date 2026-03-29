<script setup>
import { computed, createVNode, reactive, ref, watch } from 'vue'
import { CloseOutlined, QuestionCircleFilled } from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'
import dayjs from 'dayjs'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  title: {
    type: String,
    default: '结班并结课',
  },
  /** 列表行数据，用于展示课程与学费信息 */
  record: {
    type: Object,
    default: null,
  },
})

const emit = defineEmits(['update:open', 'confirm'])

const formRef = ref()
const endDateValue = ref('')

const modalBodyStyle = { padding: 0, background: '#fafafa' }

const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const formState = reactive({
  remark: '',
})

const recordId = computed(() => (props.record?.id != null ? String(props.record.id) : ''))

const courseHeadline = computed(() => {
  const r = props.record
  if (!r)
    return '-'
  const student = r.studentName || '-'
  const lesson = r.lessonName || '-'
  return `${student} — ${lesson}`
})

function normalizeChargingMode(mode) {
  const m = Number(mode)
  if (m === 4)
    return 3
  return m
}

function lessonChargingModeOf(block) {
  return block?.tuitionAccount?.lessonChargingMode
}

function effectiveLessonChargingMode(block) {
  const raw = Number(lessonChargingModeOf(block) ?? 0)
  if (raw !== 0)
    return raw
  const ta = block?.tuitionAccount
  if (!ta)
    return 0
  const q = Number(ta.totalQuantity || 0)
  const total = Number(ta.totalTuition || 0)
  if (q > 0 && total > 0 && ta.enableExpireTime)
    return 2
  return 0
}

function getQuantityUnit(mode) {
  const m = normalizeChargingMode(mode)
  if (m === 1)
    return '课时'
  if (m === 2)
    return '天'
  if (m === 3)
    return '元'
  return ''
}

function getChargingTagText(mode) {
  const m = normalizeChargingMode(mode)
  if (m === 1)
    return '课时'
  if (m === 2)
    return '时段'
  if (m === 3)
    return '按金额'
  return ''
}

function calcRemainQuantity(block) {
  const ta = block?.tuitionAccount
  return Number(ta?.remainQuantity || 0) + Number(ta?.remainFreeQuantity || 0)
}

/** 与 one-to-one-enrollment-detail 一致；scope 4 文案对齐设计稿「全部课程通用」 */
function getEnrollmentTags(block) {
  const ta = block?.tuitionAccount
  const scope = Number(ta?.lessonScopeModel ?? ta?.lessonScope ?? 0)
  const teach = Number(ta?.lessonType ?? 0)
  const mode = effectiveLessonChargingMode(block)
  const tags = []

  if (scope === 2 || scope === 3 || scope === 4)
    tags.push({ text: '通用课', type: 'primary' })

  if (teach === 1)
    tags.push({ text: '班级授课', type: 'normal' })
  else if (teach === 2)
    tags.push({ text: '1对1授课', type: 'normal' })

  if (scope === 4)
    tags.push({ text: '全部课程通用', type: 'normal' })
  else if (scope === 2) {
    if (teach === 1)
      tags.push({ text: '全部班课', type: 'normal' })
    else if (teach === 2)
      tags.push({ text: '全部1对1', type: 'normal' })
    else
      tags.push({ text: '全部通用', type: 'normal' })
  }
  else if (scope === 3) {
    tags.push({ text: '部分课程', type: 'normal' })
  }

  const chargingTag = getChargingTagText(mode)
  if (chargingTag)
    tags.push({ text: chargingTag, type: 'normal' })

  return tags
}

const enrollmentTags = computed(() => getEnrollmentTags(props.record || {}))

function formatMoney(value) {
  return `¥ ${Number(value || 0).toLocaleString('zh-CN', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })}`
}

function isZeroDateValue(value) {
  if (!value)
    return true
  if (typeof value === 'string' && value.startsWith('0001-01-01'))
    return true
  const parsed = dayjs(value)
  return !parsed.isValid() || parsed.year() <= 1
}

function formatDate(value) {
  if (isZeroDateValue(value))
    return '-'
  return dayjs(value).format('YYYY-MM-DD')
}

const remainQuantityText = computed(() => {
  const r = props.record
  if (!r)
    return '-'
  const mode = effectiveLessonChargingMode(r)
  const n = calcRemainQuantity(r)
  const u = getQuantityUnit(mode)
  return u ? `${n} ${u}` : String(n)
})

const remainTuitionText = computed(() => {
  const r = props.record
  if (!r)
    return '-'
  return formatMoney(r.tuitionAccount?.remainTuition)
})

const expireDisplayText = computed(() => {
  const r = props.record
  const ta = r?.tuitionAccount
  if (!ta?.enableExpireTime)
    return '无'
  return formatDate(ta.expireTime)
})

const endDateLabel = computed(() => {
  const d = endDateValue.value
  if (!d)
    return '-'
  const parsed = dayjs(d)
  if (!parsed.isValid())
    return d
  const isToday = parsed.isSame(dayjs(), 'day')
  return isToday ? `${d}（今天）` : d
})

const grayCardBodyStyle = { padding: '24px', background: 'transparent' }
const whiteCardBodyStyle = { padding: '20px 20px 16px' }

watch(
  () => props.open,
  (v) => {
    if (v) {
      formState.remark = ''
      endDateValue.value = dayjs().format('YYYY-MM-DD')
    }
  },
)

async function handleSubmit() {
  try {
    await formRef.value?.validate?.()
    const payload = {
      recordId: recordId.value,
      endDate: endDateValue.value,
      remark: formState.remark?.trim() || '',
    }
    Modal.confirm({
      title: '确定结课?',
      centered: true,
      closable: false,
      maskClosable: false,
      keyboard: false,
      icon: createVNode(QuestionCircleFilled, { style: { color: '#ff4d4f', fontSize: '22px' } }),
      content:
        '结课后，学员报读课程的剩余课时将全部扣除，机构获得相应课消收入。',
      okText: '确定结课',
      cancelText: '再想想',
      okButtonProps: { danger: true, ghost: true },
      wrapClassName: 'end-course-pre-confirm-modal',
      onOk() {
        emit('confirm', payload)
      },
    })
  }
  catch {
    // 校验未通过
  }
}

function closeFun() {
  formState.remark = ''
  formRef.value?.resetFields?.()
  openModal.value = false
}
</script>

<template>
  <a-modal
    v-model:open="openModal"
    :keyboard="false"
    :closable="false"
    :mask-closable="false"
    :width="800"
    :destroy-on-close="true"
    :body-style="modalBodyStyle"
  >
    <template #title>
      <div class="flex-between w-full pr-2">
        <span class="text-5">{{ title }}</span>
        <a-button type="text" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5" />
          </template>
        </a-button>
      </div>
    </template>

    <div class="scrollbar max-h-[calc(100vh-220px)] overflow-y-auto  pb-16px mt--8px">
      <a-alert
        type="warning"
        show-icon
        border="0"
        color="#f90"
        message="结课后，学员报读课程的剩余课时将全部扣除，机构获得相应课消收入。"
      />

      <a-card
        :bordered="false"
        class="rounded-8px !bg-#f5f6f8 !shadow-none"
        :body-style="grayCardBodyStyle"
      >
        <a-card
          :bordered="false"
          class="rounded-8px !shadow-none"
          :body-style="whiteCardBodyStyle"
        >
          <a-space direction="vertical" :size="16" class="w-full">
            <div class="text-lg font-600 text-#000000e0 leading-snug">
              {{ courseHeadline }}
            </div>

            <a-space v-if="enrollmentTags.length" :size="8" wrap class="w-full">
              <a-tag
                v-for="(tag, tIdx) in enrollmentTags"
                :key="`tag-${tIdx}`"
                class="m-0 rounded-full !border-none h-22px inline-flex items-center leading-20px text-3"
                :class="tag.type === 'primary' ? '!text-white' : '!text-#06f'"
                :color="tag.type === 'primary' ? '#0066ff' : '#e6f0ff'"
              >
                {{ tag.text }}
              </a-tag>
            </a-space>

            <a-descriptions
              :column="2"
              size="small"
              layout="horizontal"
              :colon="false"
            >
              <a-descriptions-item label="剩余课时：">
                <span class="text-#888">{{ remainQuantityText }}</span>
              </a-descriptions-item>
              <a-descriptions-item label="剩余学费：">
                <span class="text-#888">{{ remainTuitionText }}</span>
              </a-descriptions-item>
              <a-descriptions-item label="有效期至：" :span="2">
                <span class="text-#888">{{ expireDisplayText }}</span>
              </a-descriptions-item>
            </a-descriptions>

            <a-divider class="my-0" />

            <a-form ref="formRef" :model="formState">
              <a-form-item label="结课日期" class="!mb-12px">
                <span class="text-14px text-#ff4d4f">{{ endDateLabel }}</span>
              </a-form-item>
              <a-form-item name="remark" label="结课备注" class="!mb-0">
                <a-input
                  v-model:value="formState.remark"
                  placeholder="请输入"
                />
              </a-form-item>
            </a-form>
          </a-space>
        </a-card>
      </a-card>
    </div>

    <template #footer>
      <a-button danger ghost @click="closeFun">
        关闭
      </a-button>
      <a-button type="primary" ghost @click="handleSubmit">
        确定
      </a-button>
    </template>
  </a-modal>
</template>

<style>
.end-course-pre-confirm-modal .ant-modal-content {
  border-radius: 20px;
  overflow: hidden;
}
</style>
