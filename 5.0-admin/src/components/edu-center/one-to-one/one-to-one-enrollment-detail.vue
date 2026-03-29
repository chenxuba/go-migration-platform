<script setup>
import dayjs from 'dayjs'
import { computed } from 'vue'

const props = defineProps({
  record: {
    type: Object,
    default: () => ({}),
  },
  /** 接口返回的学费账户列表；为空时用 record.tuitionAccount 兜底 */
  accounts: {
    type: Array,
    default: () => [],
  },
})

function mapApiAccountToTuitionVO(item) {
  if (!item)
    return null
  return {
    id: item.id,
    totalTuition: item.totalTuition,
    remainTuition: item.tuition,
    totalQuantity: item.totalQuantity,
    totalFreeQuantity: item.totalFreeQuantity,
    remainQuantity: item.quantity,
    remainFreeQuantity: item.freeQuantity,
    lessonChargingMode: item.lessonChargingMode,
    lessonScopeModel: item.lessonScope,
    productName: item.productName,
    status: item.status,
    enableExpireTime: item.enableExpireTime,
    expireTime: item.expireTime,
    studentId: item.studentId,
    lessonId: item.lessonId,
    lessonType: item.lessonType,
    assignedClass: item.assignedClass,
  }
}

const blocks = computed(() => {
  const status = props.record?.classStudentStatus
  if (Array.isArray(props.accounts) && props.accounts.length > 0) {
    return props.accounts.map(acc => ({
      classStudentStatus: status,
      tuitionAccount: mapApiAccountToTuitionVO(acc),
    }))
  }
  return [{
    classStudentStatus: status,
    tuitionAccount: props.record?.tuitionAccount,
  }]
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

function calcUsedQuantity(block) {
  const ta = block.tuitionAccount
  const total = Number(ta?.totalQuantity || 0) + Number(ta?.totalFreeQuantity || 0)
  const remain = Number(ta?.remainQuantity || 0) + Number(ta?.remainFreeQuantity || 0)
  return Math.max(total - remain, 0)
}

function calcRemainQuantity(block) {
  const ta = block.tuitionAccount
  return Number(ta?.remainQuantity || 0) + Number(ta?.remainFreeQuantity || 0)
}

function calcUsedTuition(block) {
  const ta = block.tuitionAccount
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

function lessonChargingModeOf(block) {
  return block.tuitionAccount?.lessonChargingMode
}

function enrollmentTitleOf(block) {
  const ta = block.tuitionAccount
  return ta?.productName || props.record?.lessonName || props.record?.name || '-'
}

function expireTextOf(block) {
  const ta = block.tuitionAccount
  if (!ta?.enableExpireTime)
    return '不限制'
  return formatDate(ta.expireTime)
}

function purchaseQuantityTextOf(block) {
  const ta = block.tuitionAccount
  const q = Number(ta?.totalQuantity || 0)
  const u = getQuantityUnit(lessonChargingModeOf(block))
  return `购 ${q} ${u}`.trim()
}

function usedQuantityTextOf(block) {
  return formatQuantityAmount(calcUsedQuantity(block), lessonChargingModeOf(block))
}

function remainQuantityTextOf(block) {
  return formatQuantityAmount(calcRemainQuantity(block), lessonChargingModeOf(block))
}

function quantityTagLabelOf(block) {
  return getQuantityUnit(lessonChargingModeOf(block)) || '课时'
}

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

function blockKey(block, idx) {
  return block.tuitionAccount?.id ?? `row-${idx}`
}
</script>

<template>
  <div class="m-12px">
    <div
      v-for="(block, idx) in blocks"
      :key="blockKey(block, idx)"
      class="bg-white pt-18px px-20px pb-18px rounded-10px mb-12px last:mb-0"
    >
      <a-space direction="vertical" size="middle" class="w-full">
        <div>
          <div class="text-4 text-#222 font-500 mb-1">
            {{ enrollmentTitleOf(block) }}
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
              {{ quantityTagLabelOf(block) }}
            </a-tag>
          </a-space>
        </div>

        <a-descriptions :column="3" size="small">
          <a-descriptions-item label="当前状态">
            {{ classStudentStatusLabel(block.classStudentStatus) }}
          </a-descriptions-item>
          <a-descriptions-item label="报读数量">
            {{ purchaseQuantityTextOf(block) }}
          </a-descriptions-item>
          <a-descriptions-item label="有效期至">
            {{ expireTextOf(block) }}
          </a-descriptions-item>
          <a-descriptions-item label="已用数量">
            {{ usedQuantityTextOf(block) }}
          </a-descriptions-item>
          <a-descriptions-item label="剩余数量">
            {{ remainQuantityTextOf(block) }}
          </a-descriptions-item>
          <a-descriptions-item label="总学费">
            {{ formatMoney(block.tuitionAccount?.totalTuition) }}
          </a-descriptions-item>
          <a-descriptions-item label="已用学费金额">
            {{ formatMoney(calcUsedTuition(block)) }}
          </a-descriptions-item>
          <a-descriptions-item label="剩余学费金额">
            {{ formatMoney(block.tuitionAccount?.remainTuition) }}
          </a-descriptions-item>
        </a-descriptions>
      </a-space>
    </div>
  </div>
</template>
