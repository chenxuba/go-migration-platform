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

/** 创建课程提交时「按金额」报价单 lessonModel 3 会写成 4 */
function normalizeChargingMode(mode) {
  const m = Number(mode)
  if (m === 4)
    return 3
  return m
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

/** 标签文案：按时段显示「时段」，与列表「按时段」一致 */
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
  const m = normalizeChargingMode(mode)
  const u = getQuantityUnit(mode)
  if (!u)
    return `${Number(n)}`
  if (m === 1)
    return `${Number(n).toFixed(2)} ${u}`.trim()
  return `${Number(n)} ${u}`.trim()
}

function lessonChargingModeOf(block) {
  return block.tuitionAccount?.lessonChargingMode
}

/** 接口 lessonChargingMode 为 0 时兜底：有购买数量、总价且开启有效期时，多为按时段/天（与后端报价解析失败时的展示一致） */
function effectiveLessonChargingMode(block) {
  const raw = Number(lessonChargingModeOf(block) ?? 0)
  if (raw !== 0)
    return raw
  const ta = block.tuitionAccount
  if (!ta)
    return 0
  const q = Number(ta.totalQuantity || 0)
  const total = Number(ta.totalTuition || 0)
  if (q > 0 && total > 0 && ta.enableExpireTime)
    return 2
  return 0
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
  const u = getQuantityUnit(effectiveLessonChargingMode(block))
  return u ? `购 ${q} ${u}` : `购 ${q}`
}

function usedQuantityTextOf(block) {
  return formatQuantityAmount(calcUsedQuantity(block), effectiveLessonChargingMode(block))
}

function remainQuantityTextOf(block) {
  return formatQuantityAmount(calcRemainQuantity(block), effectiveLessonChargingMode(block))
}

function getEnrollmentTags(block) {
  const ta = block.tuitionAccount
  const teach = Number(ta?.lessonType ?? 0)
  const mode = effectiveLessonChargingMode(block)

  const tags = []

  if (teach === 1)
    tags.push({ text: '班级授课', type: 'normal' })
  else if (teach === 2)
    tags.push({ text: '1对1授课', type: 'normal' })

  const chargingTag = getChargingTagText(mode)
  if (chargingTag)
    tags.push({ text: chargingTag, type: 'normal' })

  return tags
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
            <a-tag
              v-for="(tag, tIdx) in getEnrollmentTags(block)"
              :key="`${blockKey(block, idx)}-tag-${tIdx}`"
              :style="getTagStyle(tag.type)"
              :color="tag.type === 'primary' ? '#0066ff' : '#e6f0ff'"
            >
              {{ tag.text }}
            </a-tag>
          </a-space>
        </div>

        <a-descriptions :column="3" size="small">
          <a-descriptions-item label="当前状态">
            <span class="text-#666666">
              {{ classStudentStatusLabel(block.classStudentStatus) }}
            </span>
          </a-descriptions-item>
          <a-descriptions-item label="报读数量">
            <span class="text-#666666">
              {{ purchaseQuantityTextOf(block) }}
            </span>
          </a-descriptions-item>
          <a-descriptions-item label="有效期至">
            <span class="text-#666666">
              {{ expireTextOf(block) }}
            </span>
          </a-descriptions-item>
          <a-descriptions-item label="已用数量">
            <span class="text-#666666">
              {{ usedQuantityTextOf(block) }}
            </span>
          </a-descriptions-item>
          <a-descriptions-item label="剩余数量">
            <span class="text-#666666">
              {{ remainQuantityTextOf(block) }}
            </span>
          </a-descriptions-item>
          <a-descriptions-item label="总学费">
            <span class="text-#666666">
              {{ formatMoney(block.tuitionAccount?.totalTuition) }}
            </span>
          </a-descriptions-item>
          <a-descriptions-item label="已用学费金额">
            <!-- 颜色#666666 -->
            <span class="text-#666666">
              {{ formatMoney(calcUsedTuition(block)) }}
            </span>
          </a-descriptions-item>
          <a-descriptions-item label="剩余学费金额">
            <span class="text-#666666">
              {{ formatMoney(block.tuitionAccount?.remainTuition) }}
            </span>
          </a-descriptions-item>
        </a-descriptions>
      </a-space>
    </div>
  </div>
</template>
