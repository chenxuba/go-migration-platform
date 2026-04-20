<script setup>
import { computed, ref, watch } from 'vue'
import { CloseOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import { getLeaveDetailApi } from '@/api/home-center/leave'
import messageService from '@/utils/messageService'

const open = defineModel({
  type: Boolean,
  default: false,
})

const props = defineProps({
  leaveId: {
    type: [String, Number],
    default: '',
  },
})

const emit = defineEmits(['closed'])

const defaultStudentAvatar = 'https://cdn.schoolpal.cn/schoolpal/next-erp/avator_male.png?x-oss-process=image/resize,w_120'

const loading = ref(false)
const detail = ref(createEmptyDetail())

const statusStyleMap = {
  1: {
    color: '#1677ff',
    background: '#e6f4ff',
  },
  2: {
    color: '#389e0d',
    background: '#f6ffed',
  },
  3: {
    color: '#cf1322',
    background: '#fff1f0',
  },
  4: {
    color: '#8c8c8c',
    background: '#f5f5f5',
  },
}

const summaryItems = computed(() => [
  { label: '开始时间', value: formatDateTime(detail.value.startTime) },
  { label: '结束时间', value: formatDateTime(detail.value.endTime) },
  { label: '请假类型', value: detail.value.leaveTypeText || '-' },
  { label: '发起人', value: formatInitiateStaffName(detail.value.operatorName || detail.value.initiateStaffName, detail.value.isAgent) },
  { label: '处理状态', value: detail.value.statusText || '-' },
  { label: '审批人', value: formatApproverName(detail.value.approverName || detail.value.currentApproverName, detail.value.status) },
  { label: '申请时间', value: formatDateTime(detail.value.applyTime) },
  { label: '备注', value: detail.value.remark || '-' },
])

function createEmptyDetail() {
  return {
    id: '',
    studentName: '',
    studentPhone: '',
    studentAvatarUrl: '',
    startTime: '',
    endTime: '',
    isAgent: true,
    leaveTypeText: '',
    reason: '',
    proofMaterials: [],
    remark: '',
    status: 1,
    statusText: '',
    initiateStaffName: '',
    operatorName: '',
    currentApproverName: '',
    approverName: '',
    applyTime: '',
    schedules: [],
    processes: [],
  }
}

function resetDetail() {
  detail.value = createEmptyDetail()
}

function formatDateTime(value) {
  if (!value || String(value).startsWith('0001-01-01'))
    return '-'
  const parsed = dayjs(value)
  return parsed.isValid() ? parsed.format('YYYY-MM-DD HH:mm') : String(value).replace('T', ' ').slice(0, 16)
}

function formatTimeOnly(value) {
  if (!value)
    return '--:--'
  const parsed = dayjs(value)
  return parsed.isValid() ? parsed.format('HH:mm') : '--:--'
}

function formatInitiateStaffName(name, isAgent) {
  const normalized = String(name || '').trim()
  if (!normalized)
    return '-'
  if (!isAgent)
    return normalized
  if (normalized.includes('（代办）') || normalized.includes('(代办)'))
    return normalized
  return `${normalized}（代办）`
}

function formatApproverName(name, status) {
  const normalized = String(name || '').trim()
  if (normalized)
    return normalized
  if (Number(status) === 2)
    return '系统自动执行'
  return '-'
}

function formatScheduleTitle(item) {
  return item.teachingClassName || item.lessonName || '-'
}

function getStatusStyle(status) {
  return statusStyleMap[status] || statusStyleMap[4]
}

function getProcessBadge(process) {
  if (process.pending) {
    return {
      text: '待处理',
      color: '#1677ff',
      background: '#e6f4ff',
    }
  }

  if (process.actionType === 1) {
    return {
      text: process.status || '发起',
      color: '#1677ff',
      background: '#e6f4ff',
    }
  }

  if (process.actionType === 2 || process.actionType === 5) {
    return {
      text: process.status || '已通过',
      color: '#389e0d',
      background: '#f6ffed',
    }
  }

  if (process.actionType === 3) {
    return {
      text: process.status || '已拒绝',
      color: '#cf1322',
      background: '#fff1f0',
    }
  }

  return {
    text: process.status || '-',
    color: '#8c8c8c',
    background: '#f5f5f5',
  }
}

function getProcessDotClass(process) {
  if (process.pending)
    return 'pending'
  if (process.actionType === 3)
    return 'danger'
  if (process.actionType === 2 || process.actionType === 5)
    return 'success'
  return 'primary'
}

async function fetchLeaveDetail() {
  if (!props.leaveId)
    return

  try {
    loading.value = true
    const res = await getLeaveDetailApi({ id: String(props.leaveId) })
    if (res.code === 200) {
      detail.value = {
        ...createEmptyDetail(),
        ...(res.result || {}),
        proofMaterials: Array.isArray(res.result?.proofMaterials) ? res.result.proofMaterials : [],
        schedules: Array.isArray(res.result?.schedules) ? res.result.schedules : [],
        processes: Array.isArray(res.result?.processes) ? res.result.processes : [],
      }
    }
  }
  catch (error) {
    console.error('获取请假详情失败:', error)
    messageService.error('获取请假详情失败')
  }
  finally {
    loading.value = false
  }
}

watch(
  () => [open.value, props.leaveId],
  ([visible, leaveId]) => {
    if (visible && leaveId) {
      fetchLeaveDetail()
      return
    }

    if (!visible) {
      resetDetail()
    }
  },
  { immediate: true },
)

function handleClose() {
  open.value = false
  emit('closed')
}
</script>

<template>
  <a-drawer
    v-model:open="open"
    width="860"
    :body-style="{ padding: '24px 28px', background: '#f7f8fa' }"
    :keyboard="false"
    :mask-closable="false"
    :closable="false"
  >
    <template #title>
      <div class="drawer-header">
        <span class="drawer-title">请假详情</span>
        <a-button type="text" class="close-btn" @click="handleClose">
          <template #icon>
            <CloseOutlined class="text-5" />
          </template>
        </a-button>
      </div>
    </template>

    <a-spin :spinning="loading">
      <div class="detail-page">
        <div class="hero-card">
          <div class="hero-main">
            <img class="hero-avatar" :src="detail.studentAvatarUrl || defaultStudentAvatar" alt="">
            <div class="hero-info">
              <div class="hero-name">
                {{ detail.studentName || '-' }}
              </div>
              <div class="hero-phone">
                {{ detail.studentPhone || '-' }}
              </div>
            </div>
          </div>
          <span class="status-chip" :style="getStatusStyle(detail.status)">
            {{ detail.statusText || '-' }}
          </span>
        </div>

        <div class="section-card">
          <div class="section-title">
            基本信息
          </div>
          <div class="summary-grid">
            <div v-for="item in summaryItems" :key="item.label" class="summary-item">
              <div class="summary-label">
                {{ item.label }}
              </div>
              <div class="summary-value">
                {{ item.value }}
              </div>
            </div>
          </div>
        </div>

        <div class="section-card">
          <div class="section-title">
            请假原因
          </div>
          <div class="section-text">
            {{ detail.reason || '未填写请假原因' }}
          </div>
        </div>

        <div class="section-card">
          <div class="section-header">
            <div class="section-title">
              请假课节
            </div>
            <div class="section-extra">
              共 {{ detail.schedules.length }} 节
            </div>
          </div>
          <div v-if="detail.schedules.length" class="schedule-list">
            <div v-for="item in detail.schedules" :key="item.scheduleId" class="schedule-item">
              <div class="schedule-time">
                {{ formatDateTime(item.startTime) }} - {{ formatTimeOnly(item.endTime) }}
              </div>
              <div class="schedule-name">
                {{ formatScheduleTitle(item) }}
              </div>
              <div class="schedule-meta">
                <span>课程：{{ item.lessonName || '-' }}</span>
                <span>教师：{{ item.teacherName || '-' }}</span>
              </div>
            </div>
          </div>
          <a-empty v-else description="暂无课节信息" />
        </div>

        <div class="section-card">
          <div class="section-title">
            请假佐证材料
          </div>
          <div v-if="detail.proofMaterials.length" class="proof-list">
            <a-image
              v-for="(url, index) in detail.proofMaterials"
              :key="`${url}-${index}`"
              :src="url"
              :width="88"
              :height="88"
              class="proof-image"
            />
          </div>
          <a-empty v-else description="未上传佐证材料" />
        </div>

        <div class="section-card">
          <div class="section-title">
            请假流程
          </div>
          <div v-if="detail.processes.length" class="process-list">
            <div
              v-for="(item, index) in detail.processes"
              :key="`${item.actionType}-${index}-${item.name}`"
              class="process-item"
            >
              <div class="process-line" :class="{ hidden: index === detail.processes.length - 1 }" />
              <div class="process-dot" :class="getProcessDotClass(item)" />
              <div class="process-card">
                <div class="process-top">
                  <div class="process-name">
                    {{ item.name || '-' }}
                  </div>
                  <span class="process-badge" :style="getProcessBadge(item)">
                    {{ getProcessBadge(item).text }}
                  </span>
                </div>
                <div class="process-time">
                  {{ item.pending ? '待审批中' : formatDateTime(item.actionTime) }}
                </div>
                <div v-if="item.remark" class="process-remark">
                  {{ item.remark }}
                </div>
              </div>
            </div>
          </div>
          <a-empty v-else description="暂无流程记录" />
        </div>
      </div>
    </a-spin>
  </a-drawer>
</template>

<style lang="less" scoped>
.drawer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.drawer-title {
  color: #1f1f1f;
  font-size: 18px;
  font-weight: 600;
}

.close-btn:hover {
  background: transparent;
}

.detail-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-height: calc(100vh - 120px);
}

.hero-card,
.section-card {
  background: #fff;
  border-radius: 14px;
  box-shadow: 0 8px 24px rgba(15, 23, 42, 0.04);
}

.hero-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 20px 22px;
}

.hero-main {
  display: flex;
  align-items: center;
  gap: 14px;
}

.hero-avatar {
  width: 64px;
  height: 64px;
  border-radius: 999px;
  object-fit: cover;
}

.hero-name {
  color: #1f1f1f;
  font-size: 20px;
  font-weight: 600;
  line-height: 28px;
}

.hero-phone {
  margin-top: 4px;
  color: #8c8c8c;
  font-size: 14px;
  line-height: 22px;
}

.status-chip {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 84px;
  padding: 4px 12px;
  border-radius: 999px;
  font-size: 13px;
  font-weight: 600;
  line-height: 24px;
}

.section-card {
  padding: 20px 22px;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.section-title {
  color: #1f1f1f;
  font-size: 16px;
  font-weight: 600;
  line-height: 24px;
}

.section-extra {
  color: #8c8c8c;
  font-size: 12px;
}

.section-text {
  margin-top: 12px;
  color: #595959;
  font-size: 14px;
  line-height: 24px;
  white-space: pre-wrap;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px 16px;
  margin-top: 14px;
}

.summary-item {
  padding: 12px 14px;
  background: #fafbfc;
  border-radius: 10px;
}

.summary-label {
  color: #8c8c8c;
  font-size: 12px;
  line-height: 20px;
}

.summary-value {
  margin-top: 4px;
  color: #1f1f1f;
  font-size: 14px;
  line-height: 22px;
  word-break: break-all;
}

.schedule-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-top: 14px;
}

.schedule-item {
  padding: 14px;
  background: #fafbfc;
  border-radius: 10px;
}

.schedule-time {
  color: #1f1f1f;
  font-size: 14px;
  font-weight: 600;
  line-height: 22px;
}

.schedule-name {
  margin-top: 4px;
  color: #1f1f1f;
  font-size: 14px;
  line-height: 22px;
}

.schedule-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  margin-top: 8px;
  color: #8c8c8c;
  font-size: 12px;
  line-height: 20px;
}

.proof-list {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 14px;
}

.proof-image {
  overflow: hidden;
  border-radius: 10px;
}

.process-list {
  margin-top: 14px;
}

.process-item {
  position: relative;
  padding-left: 28px;
  padding-bottom: 16px;
}

.process-line {
  position: absolute;
  top: 16px;
  left: 7px;
  width: 2px;
  height: calc(100% - 4px);
  background: #d9d9d9;
}

.process-line.hidden {
  display: none;
}

.process-dot {
  position: absolute;
  top: 4px;
  left: 0;
  width: 16px;
  height: 16px;
  border-radius: 999px;
  border: 3px solid #fff;
  box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.06);
}

.process-dot.primary {
  background: #1677ff;
}

.process-dot.success {
  background: #52c41a;
}

.process-dot.danger {
  background: #ff4d4f;
}

.process-dot.pending {
  background: #faad14;
}

.process-card {
  padding: 14px 16px;
  background: #fafbfc;
  border-radius: 10px;
}

.process-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.process-name {
  color: #1f1f1f;
  font-size: 14px;
  font-weight: 600;
  line-height: 22px;
}

.process-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 500;
  line-height: 20px;
}

.process-time,
.process-remark {
  margin-top: 8px;
  color: #8c8c8c;
  font-size: 12px;
  line-height: 20px;
}
</style>
