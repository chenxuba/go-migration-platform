<script setup lang="ts">
import { CloseOutlined, ExclamationCircleFilled } from '@ant-design/icons-vue'
import type { TeachingScheduleValidationResult } from '@/api/edu-center/teaching-schedule'

const props = defineProps<{
  open: boolean
  validation?: TeachingScheduleValidationResult | null
  title?: string
  compareTitle?: string
  currentTitle?: string
  existingTitle?: string
  currentColumnTitle?: string
  existingColumnTitle?: string
  fallbackMessage?: string
  showFooter?: boolean
  closeText?: string
  continueText?: string
  continueLoading?: boolean
  continueDisabled?: boolean
  continueHint?: string
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'continue'): void
  (e: 'close'): void
}>()

const modalOpen = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const currentSchedules = computed(() => props.validation?.currentSchedules || [])
const existingSchedules = computed(() => props.validation?.existingSchedules || [])
const activeConflictFilter = ref<'all' | '老师' | '学员' | '教室'>('all')
const isExistingConflictMode = computed(() =>
  String(props.currentTitle || '').includes('冲突')
  || String(props.currentColumnTitle || '').includes('冲突'),
)

function hasConflictType(item: { conflictTypes?: string[] }, type: string) {
  return (item.conflictTypes || []).includes(type)
}

function parseTimeText(text?: string) {
  const m = String(text || '').match(/(\d{1,2}:\d{2})[~～](\d{1,2}:\d{2})/)
  if (!m)
    return null
  const toMinutes = (value: string) => {
    const [hour, minute] = value.split(':').map(Number)
    return hour * 60 + minute
  }
  return {
    start: toMinutes(m[1]),
    end: toMinutes(m[2]),
  }
}

function schedulesOverlap(
  current: { date?: string, timeText?: string },
  existing: { date?: string, timeText?: string },
) {
  if (current.date !== existing.date)
    return false
  const currentRange = parseTimeText(current.timeText)
  const existingRange = parseTimeText(existing.timeText)
  if (!currentRange || !existingRange)
    return false
  return currentRange.start < existingRange.end && currentRange.end > existingRange.start
}

const conflictTypeStats = computed(() => {
  const list = currentSchedules.value
  return {
    total: list.length,
    teacher: list.filter(item => hasConflictType(item, '老师')).length,
    student: list.filter(item => hasConflictType(item, '学员')).length,
    classroom: list.filter(item => hasConflictType(item, '教室')).length,
  }
})

const summaryNounText = computed(() =>
  isExistingConflictMode.value ? '冲突日程' : '待处理日程',
)

const conflictFilters = computed(() => {
  const stats = conflictTypeStats.value
  return [
    { key: 'all', label: `全部 ${stats.total}` },
    { key: '老师', label: `老师 ${stats.teacher}` },
    { key: '学员', label: `学员 ${stats.student}` },
    { key: '教室', label: `教室 ${stats.classroom}` },
  ]
})

const conflictGroups = computed(() =>
  currentSchedules.value.map((current, index) => {
    const matches = existingSchedules.value.filter(existing => schedulesOverlap(current, existing))
    return {
      key: `${current.date || 'date'}-${current.timeText || 'time'}-${index}`,
      index: index + 1,
      current,
      matches,
    }
  }),
)

const visibleConflictGroups = computed(() => {
  if (activeConflictFilter.value === 'all')
    return conflictGroups.value
  return conflictGroups.value.filter(group => hasConflictType(group.current, activeConflictFilter.value))
})

watch(
  () => props.open,
  (open) => {
    if (open)
      activeConflictFilter.value = 'all'
  },
)

const conflictTableColumns = computed(() => [
  {
    title: props.currentColumnTitle || '待创建日程',
    key: 'current',
    dataIndex: 'current',
    width: '50%',
  },
  {
    title: props.existingColumnTitle || '与其冲突的日程',
    key: 'existing',
    dataIndex: 'matches',
    width: '50%',
  },
])

function handleClose() {
  emit('close')
  modalOpen.value = false
}

function handleContinue() {
  emit('continue')
}
</script>

<template>
  <a-modal
    v-model:open="modalOpen"
    centered
    class="schedule-conflict-modal"
    :footer="null"
    :width="1180"
    :body-style="{ paddingTop: '0px' }"
    :keyboard="false"
    :closable="false"
    :mask-closable="true"
  >
    <template #title>
      <div class="schedule-conflict__titlebar">
        <span>{{ props.title || '冲突提示' }}</span>
        <a-button type="text" @click="handleClose">
          <template #icon>
            <CloseOutlined />
          </template>
        </a-button>
      </div>
    </template>

    <div class="schedule-conflict">
      <div class="schedule-conflict__banner">
        <ExclamationCircleFilled />
        <span>{{ validation?.message || props.fallbackMessage || '当前创建日程与已有日程冲突' }}</span>
      </div>

      <div class="schedule-conflict__toolbar">
        <div class="schedule-conflict__toolbar-summary">
          共 {{ conflictTypeStats.total }} 节{{ summaryNounText }}，
          其中老师冲突 {{ conflictTypeStats.teacher }} 节，
          学员冲突 {{ conflictTypeStats.student }} 节，
          教室冲突 {{ conflictTypeStats.classroom }} 节。
        </div>
        <div class="schedule-conflict__filters">
          <button
            v-for="item in conflictFilters"
            :key="item.key"
            type="button"
            class="schedule-conflict__filter"
            :class="{ 'schedule-conflict__filter--active': activeConflictFilter === item.key }"
            @click="activeConflictFilter = item.key"
          >
            {{ item.label }}
          </button>
        </div>
      </div>

      <section class="schedule-conflict__section">
        <div class="schedule-conflict__section-title">
          {{ props.compareTitle || '按当前创建日程逐项查看' }}
        </div>
        <div v-if="!visibleConflictGroups.length" class="schedule-conflict__empty">
          当前筛选条件下暂无冲突日程。
        </div>
        <a-table
          v-else
          class="schedule-conflict__matrix"
          :columns="conflictTableColumns"
          :data-source="visibleConflictGroups"
          :pagination="false"
          row-key="key"
          :scroll="{ x: 980 }"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'current'">
              <div class="schedule-conflict__cell-card schedule-conflict__cell-card--current">
                <div class="schedule-conflict__cell-top">
                  <span v-if="!isExistingConflictMode" class="schedule-conflict__group-index">
                    {{ `第 ${record.index} 节待创建` }}
                  </span>
                  <span v-else class="schedule-conflict__cell-caption">
                    {{ props.currentColumnTitle || '当前冲突日程' }}
                  </span>
                  <span class="schedule-conflict__group-time">{{ record.current.date }} {{ record.current.timeText }}</span>
                </div>
                <div class="schedule-conflict__cell-main">
                  <strong>{{ record.current.name }}</strong>
                  <span>{{ record.current.classTypeText }}</span>
                </div>
                <div class="schedule-conflict__cell-meta">
                  <span>
                    老师：
                    <strong :class="{ 'schedule-conflict__cell--danger': hasConflictType(record.current, '老师') }">{{ record.current.teacherName || '-' }}</strong>
                  </span>
                  <span>
                    教室：
                    <strong :class="{ 'schedule-conflict__cell--danger': hasConflictType(record.current, '教室') }">{{ record.current.classroomName || '-' }}</strong>
                  </span>
                  <span v-if="(record.current.studentNames || []).length">
                    学员：
                    <strong :class="{ 'schedule-conflict__cell--danger': hasConflictType(record.current, '学员') }">{{ (record.current.studentNames || []).join('、') }}</strong>
                  </span>
                </div>
                <div class="schedule-conflict__tags">
                  <a-tag
                    v-for="tag in record.current.conflictTypes || []"
                    :key="`${record.key}-${tag}`"
                    color="error"
                    :bordered="false"
                  >
                    {{ tag }}冲突
                  </a-tag>
                </div>
              </div>
            </template>

            <template v-else-if="column.key === 'existing'">
              <div class="schedule-conflict__cell-stack">
                <div v-if="!record.matches.length" class="schedule-conflict__empty-inline">
                  暂无可直接匹配的冲突明细
                </div>
                <div
                  v-for="(item, index) in record.matches"
                  :key="`${item.date}-${item.timeText}-${index}`"
                  class="schedule-conflict__cell-card schedule-conflict__cell-card--existing"
                >
                  <div class="schedule-conflict__cell-top">
                    <span class="schedule-conflict__cell-caption">
                      {{ props.existingColumnTitle || '与其冲突的日程' }}
                    </span>
                    <span class="schedule-conflict__group-time">{{ item.date }} {{ item.timeText }}</span>
                  </div>
                  <div class="schedule-conflict__cell-main">
                    <strong>{{ item.name }}</strong>
                    <span>{{ item.classTypeText }}</span>
                  </div>
                  <div class="schedule-conflict__cell-meta">
                    <span>
                      老师：
                      <strong :class="{ 'schedule-conflict__cell--danger': hasConflictType(item, '老师') }">{{ item.teacherName || '-' }}</strong>
                    </span>
                    <span>
                      教室：
                      <strong :class="{ 'schedule-conflict__cell--danger': hasConflictType(item, '教室') }">{{ item.classroomName || '-' }}</strong>
                    </span>
                    <span>
                      冲突学员：
                      <strong :class="{ 'schedule-conflict__cell--danger': hasConflictType(item, '学员') }">{{ (item.studentNames || []).join('、') || '-' }}</strong>
                    </span>
                  </div>
                  <div class="schedule-conflict__tags">
                    <a-tag
                      v-for="tag in item.conflictTypes || []"
                      :key="`${record.key}-existing-${index}-${tag}`"
                      color="error"
                      :bordered="false"
                    >
                      {{ tag }}冲突
                    </a-tag>
                  </div>
                </div>
              </div>
            </template>
          </template>
        </a-table>
      </section>

      <div v-if="props.showFooter" class="schedule-conflict__footer">
        <div v-if="props.continueHint" class="schedule-conflict__footer-hint">
          {{ props.continueHint }}
        </div>
        <div class="schedule-conflict__footer-actions">
          <a-button @click="handleClose">
            {{ props.closeText || '返回修改' }}
          </a-button>
          <a-button
            v-if="props.continueText"
            type="primary"
            :loading="props.continueLoading"
            :disabled="props.continueDisabled"
            @click="handleContinue"
          >
            {{ props.continueText }}
          </a-button>
        </div>
      </div>
    </div>
  </a-modal>
</template>

<style scoped lang="less">
.schedule-conflict__titlebar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.schedule-conflict {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.schedule-conflict__banner {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  border-radius: 12px;
  background: #fff7f7;
  color: #ff7875;
  font-size: 13px;
  font-weight: 600;
  border: 1px solid #ffe1e0;
}

.schedule-conflict__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}

.schedule-conflict__toolbar-summary {
  color: #4b5563;
  font-size: 13px;
  line-height: 1.6;
}

.schedule-conflict__filters {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.schedule-conflict__filter {
  padding: 6px 12px;
  border: 1px solid #d9e1ea;
  border-radius: 999px;
  background: #fff;
  color: #4b5563;
  font-size: 13px;
  line-height: 1;
  cursor: pointer;
  transition: all 0.2s ease;
}

.schedule-conflict__filter:hover {
  border-color: #91caff;
  color: #1677ff;
}

.schedule-conflict__filter--active {
  border-color: #1677ff;
  background: #e6f4ff;
  color: #1677ff;
  font-weight: 600;
}

.schedule-conflict__section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.schedule-conflict__section-title {
  position: relative;
  padding-left: 14px;
  color: #1f2329;
  font-size: 16px;
  font-weight: 700;
}

.schedule-conflict__section-title::before {
  position: absolute;
  left: 0;
  top: 5px;
  width: 5px;
  height: 16px;
  border-radius: 999px;
  background: #1677ff;
  content: '';
}

.schedule-conflict__empty {
  padding: 28px 18px;
  border-radius: 16px;
  border: 1px dashed #d9e1ea;
  background: #fafcff;
  color: #8c8c8c;
  font-size: 14px;
  text-align: center;
}

.schedule-conflict__groups {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.schedule-conflict__matrix {
  :deep(.ant-table) {
    background: transparent;
  }

  :deep(.ant-table-thead > tr > th:nth-child(1)),
  :deep(.ant-table-thead > tr > th:nth-child(2)) {
    width: 50%;
  }

  :deep(.ant-table-thead > tr > th) {
    padding: 12px 14px;
    color: #4b5563;
    font-size: 13px;
    font-weight: 700;
    background: #f8fafc;
  }

  :deep(.ant-table-thead > tr > th::before) {
    display: none;
  }

  :deep(.ant-table-tbody > tr > td) {
    padding: 12px 14px;
    vertical-align: top;
    background: #fff;
  }

  :deep(.ant-table-tbody > tr > td:nth-child(1)),
  :deep(.ant-table-tbody > tr > td:nth-child(2)) {
    width: 50%;
  }
}

.schedule-conflict__cell-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
  height: 100%;
  min-height: 148px;
  padding: 10px 12px;
  border: 1px solid #edf2f7;
  border-radius: 12px;
  background: #f8fafc;
}

.schedule-conflict__cell-card--existing + .schedule-conflict__cell-card--existing {
  margin-top: 8px;
}

.schedule-conflict__cell-stack {
  display: flex;
  flex-direction: column;
  gap: 8px;
  height: 100%;
}

.schedule-conflict__cell-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.schedule-conflict__group-index {
  color: #1f2329;
  font-size: 14px;
  font-weight: 700;
}

.schedule-conflict__group-type {
  color: #8c8c8c;
  font-size: 13px;
}

.schedule-conflict__cell-caption {
  color: #8c8c8c;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.02em;
}

.schedule-conflict__group-time {
  color: #1677ff;
  font-size: 13px;
  font-weight: 600;
}

.schedule-conflict__cell-main {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  color: #8c8c8c;
  font-size: 12px;
  min-height: 24px;
}

.schedule-conflict__cell-main strong {
  color: #1f2329;
  font-size: 15px;
  font-weight: 700;
}

.schedule-conflict__cell-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  color: #4b5563;
  font-size: 13px;
  line-height: 1.7;
  min-height: 28px;
}

.schedule-conflict__empty-inline {
  padding: 12px;
  color: #8c8c8c;
  font-size: 13px;
  text-align: center;
}

.schedule-conflict__meta-sep {
  margin: 0 6px;
  color: #d0d7e2;
}

.schedule-conflict__table {
  overflow: hidden;
  border-radius: 18px;
  background: #fff;
  border: 1px solid #edf2f7;
}

.schedule-conflict__head,
.schedule-conflict__row {
  display: grid;
  grid-template-columns: 1.5fr 1fr 1.5fr 1fr 1fr 1fr;
  gap: 16px;
  align-items: center;
  padding: 18px 20px;
}

.schedule-conflict__head {
  background: #f8fafc;
  color: #4b5563;
  font-size: 13px;
  font-weight: 700;
}

.schedule-conflict__row {
  border-top: 1px solid #f0f2f5;
  color: #1f2329;
  font-size: 14px;
  line-height: 1.6;
}

.schedule-conflict__cell--danger {
  color: #ff4d4f;
  font-weight: 700;
}

.schedule-conflict__tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.schedule-conflict__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding-top: 8px;
}

.schedule-conflict__footer-hint {
  color: #8c8c8c;
  font-size: 13px;
  line-height: 1.6;
}

.schedule-conflict__footer-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-left: auto;
}

@media (max-width: 1200px) {
  .schedule-conflict__cell-top,
  .schedule-conflict__cell-meta {
    align-items: flex-start;
    flex-direction: column;
    gap: 8px;
  }

  .schedule-conflict__footer {
    align-items: stretch;
    flex-direction: column;
  }

  .schedule-conflict__footer-actions {
    width: 100%;
    justify-content: flex-end;
  }

  .schedule-conflict__table {
    overflow-x: auto;
  }

  .schedule-conflict__head,
  .schedule-conflict__row {
    min-width: 960px;
  }
}
</style>
