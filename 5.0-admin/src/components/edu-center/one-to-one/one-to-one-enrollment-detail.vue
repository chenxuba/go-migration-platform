<script setup>
import dayjs from 'dayjs'
import { computed } from 'vue'

const props = defineProps({
  record: {
    type: Object,
    default: () => ({}),
  },
})

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

function formatMoney(value) {
  return `¥ ${Number(value || 0).toLocaleString('zh-CN', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })}`
}

function getQuantityUnit(mode) {
  if (mode === 1)
    return '课时'
  if (mode === 2)
    return '天'
  if (mode === 3)
    return '元'
  return ''
}

function calcUsedQuantity(record) {
  const ta = record.tuitionAccount
  const total = Number(ta?.totalQuantity || 0) + Number(ta?.totalFreeQuantity || 0)
  const remain = Number(ta?.remainQuantity || 0) + Number(ta?.remainFreeQuantity || 0)
  return Math.max(total - remain, 0)
}

function calcRemainQuantity(record) {
  const ta = record.tuitionAccount
  return Number(ta?.remainQuantity || 0) + Number(ta?.remainFreeQuantity || 0)
}

function calcUsedTuition(record) {
  const ta = record.tuitionAccount
  return Math.max(Number(ta?.totalTuition || 0) - Number(ta?.remainTuition || 0), 0)
}

function classStudentStatusLabel(status) {
  if (status === 3)
    return '已结课'
  if (status === 2)
    return '已开课'
  return '正常'
}

function formatQuantityAmount(n, mode) {
  const u = getQuantityUnit(mode)
  if (mode === 1)
    return `${Number(n).toFixed(2)} ${u}`.trim()
  return `${n}${u}`
}

const lessonChargingMode = computed(() => props.record?.tuitionAccount?.lessonChargingMode)

const enrollmentTitle = computed(() =>
  props.record?.tuitionAccount?.productName || props.record?.lessonName || props.record?.name || '-',
)

const expireText = computed(() => {
  const ta = props.record?.tuitionAccount
  if (!ta?.enableExpireTime)
    return '不限制'
  return formatDate(ta.expireTime)
})

const purchaseQuantityText = computed(() => {
  const ta = props.record?.tuitionAccount
  const q = Number(ta?.totalQuantity || 0)
  const u = getQuantityUnit(lessonChargingMode.value)
  return `购 ${q} ${u}`.trim()
})

const usedQuantityText = computed(() =>
  formatQuantityAmount(calcUsedQuantity(props.record), lessonChargingMode.value),
)

const remainQuantityText = computed(() =>
  formatQuantityAmount(calcRemainQuantity(props.record), lessonChargingMode.value),
)

const quantityTagLabel = computed(() => getQuantityUnit(lessonChargingMode.value) || '课时')

/** 与「选择课程/学杂费/教材用品」弹窗（active-course-modal）中课程行标签一致 */
function getTagStyle(type = 'normal') {
  const baseStyle = {
    borderRadius: '20px',
    marginRight: '0',
    height: '20px',
  }
  if (type === 'primary') {
    return {
      ...baseStyle,
      color: '#fff',
    }
  }
  return {
    ...baseStyle,
    color: '#0066ff',
  }
}
</script>

<template>
  <div class="m-12px">
    <div class="bg-white pt-18px px-20px pb-18px rounded-10px">
      <a-space direction="vertical" size="middle" class="w-full">
        <div>
          <div class="text-4 text-#222 font-500 mb-1">
            {{ enrollmentTitle }}
          </div>
          <a-space :size="5" class="w-100% flex flex-wrap">
            <a-tag :style="getTagStyle('primary')" color="#0066ff">
              通用课
            </a-tag>
            <a-tag :style="getTagStyle('normal')" color="#e6f0ff">
              全部课程通用
            </a-tag>
            <a-tag :style="getTagStyle('normal')" color="#e6f0ff">
              1对1授课
            </a-tag>
            <a-tag :style="getTagStyle('normal')" color="#e6f0ff">
              {{ quantityTagLabel }}
            </a-tag>
          </a-space>
        </div>

        <a-descriptions :column="3" size="small">
          <a-descriptions-item label="当前状态">
            {{ classStudentStatusLabel(record.classStudentStatus) }}
          </a-descriptions-item>
          <a-descriptions-item label="报读数量">
            {{ purchaseQuantityText }}
          </a-descriptions-item>
          <a-descriptions-item label="有效期至">
            {{ expireText }}
          </a-descriptions-item>
          <a-descriptions-item label="已用数量">
            {{ usedQuantityText }}
          </a-descriptions-item>
          <a-descriptions-item label="剩余数量">
            {{ remainQuantityText }}
          </a-descriptions-item>
          <a-descriptions-item label="总学费">
            {{ formatMoney(record.tuitionAccount?.totalTuition) }}
          </a-descriptions-item>
          <a-descriptions-item label="已用学费金额">
            {{ formatMoney(calcUsedTuition(record)) }}
          </a-descriptions-item>
          <a-descriptions-item label="剩余学费金额">
            {{ formatMoney(record.tuitionAccount?.remainTuition) }}
          </a-descriptions-item>
        </a-descriptions>
      </a-space>
    </div>
  </div>
</template>
