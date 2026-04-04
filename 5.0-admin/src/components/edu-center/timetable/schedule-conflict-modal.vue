<script setup lang="ts">
import { CloseOutlined, ExclamationCircleFilled } from '@ant-design/icons-vue'
import type { TeachingScheduleValidationResult } from '@/api/edu-center/teaching-schedule'

const props = defineProps<{
  open: boolean
  validation?: TeachingScheduleValidationResult | null
  title?: string
  currentTitle?: string
  existingTitle?: string
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

      <section class="schedule-conflict__section">
        <div class="schedule-conflict__section-title">
          {{ props.currentTitle || '当前创建日程' }}
        </div>
        <div class="schedule-conflict__table">
          <div class="schedule-conflict__head">
            <span>日程名称</span>
            <span>日程类型</span>
            <span>上课时间</span>
            <span>上课教师</span>
            <span>上课教室</span>
            <span>冲突类型</span>
          </div>
          <div
            v-for="(item, index) in currentSchedules"
            :key="`${item.date}-${item.timeText}-${index}`"
            class="schedule-conflict__row"
          >
            <span>{{ item.name }}</span>
            <span>{{ item.classTypeText }}</span>
            <span>{{ item.date }} {{ item.timeText }}</span>
            <span
              :class="{
                'schedule-conflict__cell--danger': hasConflictType(item, '老师'),
              }"
            >{{ item.teacherName || '-' }}</span>
            <span
              :class="{
                'schedule-conflict__cell--danger': hasConflictType(item, '教室'),
              }"
            >{{ item.classroomName || '-' }}</span>
            <span class="schedule-conflict__tags">
              <a-tag
                v-for="tag in item.conflictTypes || []"
                :key="tag"
                color="error"
                :bordered="false"
              >
                {{ tag }}冲突
              </a-tag>
            </span>
          </div>
        </div>
      </section>

      <section class="schedule-conflict__section">
        <div class="schedule-conflict__section-title">
          {{ props.existingTitle || '校内已有日程' }}
        </div>
        <div class="schedule-conflict__table">
          <div class="schedule-conflict__head">
            <span>日程名称</span>
            <span>日程类型</span>
            <span>上课时间</span>
            <span>上课教师</span>
            <span>上课教室</span>
            <span>冲突学员</span>
          </div>
          <div
            v-for="(item, index) in existingSchedules"
            :key="`${item.date}-${item.timeText}-${index}`"
            class="schedule-conflict__row"
          >
            <span>{{ item.name }}</span>
            <span>{{ item.classTypeText }}</span>
            <span>{{ item.date }} {{ item.timeText }}</span>
            <span
              :class="{
                'schedule-conflict__cell--danger': hasConflictType(item, '老师'),
              }"
            >{{ item.teacherName || '-' }}</span>
            <span
              :class="{
                'schedule-conflict__cell--danger': hasConflictType(item, '教室'),
              }"
            >{{ item.classroomName || '-' }}</span>
            <span
              :class="{
                'schedule-conflict__cell--danger': hasConflictType(item, '学员'),
              }"
            >{{ (item.studentNames || []).join('、') || '-' }}</span>
          </div>
        </div>
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
