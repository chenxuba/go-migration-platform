<script setup lang="ts">
import { CloseOutlined, ExclamationCircleFilled } from '@ant-design/icons-vue'
import type { TeachingScheduleValidationResult } from '@/api/edu-center/teaching-schedule'

interface ConflictWorkbenchPlan {
  date: string
  week: string
  rule: string
  time: string
  startTime: string
  endTime: string
  teacher: string
  classroom: string
}

interface ConflictWorkbenchSubmitPayload {
  plans: ConflictWorkbenchPlan[]
  allowStudentConflict: boolean
  allowClassroomConflict: boolean
}

const props = defineProps<{
  open: boolean
  validation?: TeachingScheduleValidationResult | null
  plans: ConflictWorkbenchPlan[]
  loading?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'submit', payload: ConflictWorkbenchSubmitPayload): void
}>()

const modalOpen = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const currentSchedules = computed(() => props.validation?.currentSchedules || [])
const existingSchedules = computed(() => props.validation?.existingSchedules || [])

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

function previewPlanMatchesSchedule(plan: ConflictWorkbenchPlan, schedule: NonNullable<TeachingScheduleValidationResult['currentSchedules']>[number]) {
  return plan.date === schedule.date
    && `${plan.startTime}~${plan.endTime}` === String(schedule.timeText || '')
}

const rowState = ref<Record<string, { ignoreStudent: boolean, ignoreClassroom: boolean }>>({})

const rows = computed(() =>
  props.plans.map((plan, index) => {
    const current = currentSchedules.value.find(item => previewPlanMatchesSchedule(plan, item))
    const conflictTypes = current?.conflictTypes || []
    const key = `${plan.date}|${plan.startTime}|${plan.endTime}|${index}`
    const state = rowState.value[key] || { ignoreStudent: false, ignoreClassroom: false }
    const hasTeacherConflict = conflictTypes.includes('老师')
    const hasStudentConflict = conflictTypes.includes('学员')
    const hasClassroomConflict = conflictTypes.includes('教室')
    const matches = existingSchedules.value.filter(item => schedulesOverlap(current || {
      date: plan.date,
      timeText: `${plan.startTime}~${plan.endTime}`,
    }, item))
    const readyBySoftConflict = (!hasStudentConflict || state.ignoreStudent) && (!hasClassroomConflict || state.ignoreClassroom)
    const canCreate = !hasTeacherConflict && readyBySoftConflict
    return {
      key,
      index: index + 1,
      plan,
      current,
      conflictTypes,
      hasTeacherConflict,
      hasStudentConflict,
      hasClassroomConflict,
      matches,
      state,
      canCreate,
    }
  }),
)

watch(
  () => props.open,
  (open) => {
    if (!open)
      return
    const next: Record<string, { ignoreStudent: boolean, ignoreClassroom: boolean }> = {}
    rows.value.forEach((row) => {
      next[row.key] = {
        ignoreStudent: rowState.value[row.key]?.ignoreStudent ?? false,
        ignoreClassroom: rowState.value[row.key]?.ignoreClassroom ?? false,
      }
    })
    rowState.value = next
  },
  { immediate: true },
)

const summary = computed(() => {
  const total = rows.value.length
  const teacher = rows.value.filter(row => row.hasTeacherConflict).length
  const student = rows.value.filter(row => row.hasStudentConflict).length
  const classroom = rows.value.filter(row => row.hasClassroomConflict).length
  const ready = rows.value.filter(row => row.canCreate).length
  return { total, teacher, student, classroom, ready }
})

function updateRowIgnore(key: string, field: 'ignoreStudent' | 'ignoreClassroom', checked: boolean) {
  rowState.value = {
    ...rowState.value,
    [key]: {
      ignoreStudent: field === 'ignoreStudent' ? checked : (rowState.value[key]?.ignoreStudent ?? false),
      ignoreClassroom: field === 'ignoreClassroom' ? checked : (rowState.value[key]?.ignoreClassroom ?? false),
    },
  }
}

function enableAllSoftConflicts() {
  const next = { ...rowState.value }
  rows.value.forEach((row) => {
    next[row.key] = {
      ignoreStudent: row.hasStudentConflict ? true : (next[row.key]?.ignoreStudent ?? false),
      ignoreClassroom: row.hasClassroomConflict ? true : (next[row.key]?.ignoreClassroom ?? false),
    }
  })
  rowState.value = next
}

function submitReadyRows() {
  const selectedRows = rows.value.filter(row => row.canCreate)
  if (!selectedRows.length)
    return
  emit('submit', {
    plans: selectedRows.map(row => row.plan),
    allowStudentConflict: selectedRows.some(row => row.hasStudentConflict && row.state.ignoreStudent),
    allowClassroomConflict: selectedRows.some(row => row.hasClassroomConflict && row.state.ignoreClassroom),
  })
}

const columns = [
  { title: '待创建日程', key: 'current', dataIndex: 'current', width: '34%' },
  { title: '冲突详情', key: 'conflicts', dataIndex: 'conflicts', width: '42%' },
  { title: '处理方式', key: 'actions', dataIndex: 'actions', width: '24%' },
]
</script>

<template>
  <a-modal
    v-model:open="modalOpen"
    centered
    class="schedule-conflict-workbench-modal"
    :footer="null"
    :width="1240"
    :body-style="{ paddingTop: '0px' }"
    :keyboard="false"
    :closable="false"
    :mask-closable="true"
  >
    <template #title>
      <div class="schedule-conflict__titlebar">
        <span>冲突处理</span>
        <a-button type="text" @click="modalOpen = false">
          <template #icon>
            <CloseOutlined />
          </template>
        </a-button>
      </div>
    </template>

    <div class="schedule-conflict">
      <div class="schedule-conflict__banner">
        <ExclamationCircleFilled />
        <span>{{ props.validation?.message || '当前创建日程存在冲突' }}</span>
      </div>

      <div class="schedule-conflict__toolbar">
        <div class="schedule-conflict__toolbar-summary">
          共 {{ summary.total }} 节待处理日程，其中老师冲突 {{ summary.teacher }} 节，学员冲突 {{ summary.student }} 节，教室冲突 {{ summary.classroom }} 节，当前可直接创建 {{ summary.ready }} 节。
        </div>
        <a-button type="link" class="schedule-conflict__toolbar-link" @click="enableAllSoftConflicts">
          忽略全部软冲突
        </a-button>
      </div>

      <a-table
        class="schedule-conflict__workbench"
        :columns="columns"
        :data-source="rows"
        :pagination="false"
        row-key="key"
        :scroll="{ y: 560 }"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'current'">
            <div class="schedule-conflict__cell-card">
              <div class="schedule-conflict__cell-top">
                <span class="schedule-conflict__group-index">第 {{ record.index }} 节待创建</span>
                <span class="schedule-conflict__group-time">{{ record.plan.date }} {{ record.plan.startTime }}~{{ record.plan.endTime }}</span>
              </div>
              <div class="schedule-conflict__cell-main">
                <strong>{{ record.current?.name || '-' }}</strong>
                <span>{{ record.current?.classTypeText || '1对1日程' }}</span>
              </div>
              <div class="schedule-conflict__cell-meta">
                <span>老师：<strong :class="{ 'schedule-conflict__cell--danger': record.hasTeacherConflict }">{{ record.plan.teacher }}</strong></span>
                <span>教室：<strong :class="{ 'schedule-conflict__cell--danger': record.hasClassroomConflict }">{{ record.plan.classroom || '-' }}</strong></span>
                <span>学员：<strong :class="{ 'schedule-conflict__cell--danger': record.hasStudentConflict }">{{ (record.current?.studentNames || []).join('、') || '-' }}</strong></span>
              </div>
            </div>
          </template>

          <template v-else-if="column.key === 'conflicts'">
            <div class="schedule-conflict__cell-stack">
              <div class="schedule-conflict__panel-title">
                命中冲突
              </div>
              <div
                v-for="(item, index) in record.matches"
                :key="`${record.key}-${index}`"
                class="schedule-conflict__conflict-item"
              >
                <div class="schedule-conflict__cell-main">
                  <strong>{{ item.name }}</strong>
                  <span>{{ item.classTypeText }}</span>
                  <span>{{ item.date }} {{ item.timeText }}</span>
                </div>
                <div class="schedule-conflict__cell-meta">
                  <span>老师：<strong :class="{ 'schedule-conflict__cell--danger': (item.conflictTypes || []).includes('老师') }">{{ item.teacherName || '-' }}</strong></span>
                  <span>教室：<strong :class="{ 'schedule-conflict__cell--danger': (item.conflictTypes || []).includes('教室') }">{{ item.classroomName || '-' }}</strong></span>
                  <span>冲突学员：<strong :class="{ 'schedule-conflict__cell--danger': (item.conflictTypes || []).includes('学员') }">{{ (item.studentNames || []).join('、') || '-' }}</strong></span>
                </div>
                <div class="schedule-conflict__tags">
                  <a-tag
                    v-for="tag in item.conflictTypes || []"
                    :key="`${record.key}-${index}-${tag}`"
                    color="error"
                    :bordered="false"
                  >
                    {{ tag }}冲突
                  </a-tag>
                </div>
              </div>
            </div>
          </template>

          <template v-else-if="column.key === 'actions'">
            <div class="schedule-conflict__action-panel">
              <div class="schedule-conflict__panel-title">
                本行处理
              </div>
              <div v-if="record.hasTeacherConflict" class="schedule-conflict__action-tip schedule-conflict__action-tip--danger">
                <span class="schedule-conflict__action-badge schedule-conflict__action-badge--danger">老师冲突</span>
                <span>需返回修改老师或时间后再创建</span>
              </div>
              <template v-else>
                <label
                  v-if="record.hasStudentConflict"
                  class="schedule-conflict__action-option"
                >
                  <a-checkbox
                    :checked="record.state.ignoreStudent"
                    @change="event => updateRowIgnore(record.key, 'ignoreStudent', Boolean(event.target.checked))"
                  />
                  <div class="schedule-conflict__action-option-main">
                    <span>忽略学员冲突</span>
                    <small>允许学员并行上课，创建后标记冲突</small>
                  </div>
                </label>
                <label
                  v-if="record.hasClassroomConflict"
                  class="schedule-conflict__action-option"
                >
                  <a-checkbox
                    :checked="record.state.ignoreClassroom"
                    @change="event => updateRowIgnore(record.key, 'ignoreClassroom', Boolean(event.target.checked))"
                  />
                  <div class="schedule-conflict__action-option-main">
                    <span>忽略教室冲突</span>
                    <small>允许共享教室资源，创建后标记冲突</small>
                  </div>
                </label>
                <div v-if="!record.hasStudentConflict && !record.hasClassroomConflict" class="schedule-conflict__action-tip">
                  当前行无软冲突，可直接创建
                </div>
              </template>
              <div class="schedule-conflict__action-result" :class="{ 'schedule-conflict__action-result--ready': record.canCreate }">
                <span class="schedule-conflict__action-result-label">处理结果</span>
                <strong>{{ record.canCreate ? '本节可创建' : '本节暂不可创建' }}</strong>
              </div>
            </div>
          </template>
        </template>
      </a-table>

      <div class="schedule-conflict__footer">
        <div class="schedule-conflict__footer-hint">
          学员冲突、教室冲突可忽略后继续创建；老师冲突必须返回调整。
        </div>
        <div class="schedule-conflict__footer-actions">
          <a-button @click="modalOpen = false">
            返回修改
          </a-button>
          <a-button
            type="primary"
            :loading="props.loading"
            :disabled="summary.ready === 0"
            @click="submitReadyRows"
          >
            创建已处理项（{{ summary.ready }} 节）
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

.schedule-conflict__toolbar-link {
  padding: 0;
  height: auto;
  font-weight: 600;
}

.schedule-conflict__workbench {
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
}

.schedule-conflict__cell-card,
.schedule-conflict__conflict-item,
.schedule-conflict__action-panel {
  padding: 12px 14px;
  border: 1px solid #edf2f7;
  border-radius: 12px;
  background: #f8fafc;
}

.schedule-conflict__conflict-item + .schedule-conflict__conflict-item {
  margin-top: 8px;
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
  margin-top: 8px;
  color: #8c8c8c;
  font-size: 12px;
}

.schedule-conflict__cell-main strong {
  color: #1f2329;
  font-size: 15px;
  font-weight: 700;
}

.schedule-conflict__panel-title {
  color: #8c8c8c;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.02em;
}

.schedule-conflict__cell-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  margin-top: 8px;
  color: #4b5563;
  font-size: 13px;
  line-height: 1.7;
}

.schedule-conflict__cell--danger {
  color: #ff4d4f;
  font-weight: 700;
}

.schedule-conflict__tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 8px;
}

.schedule-conflict__action-panel {
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-height: 100%;
}

.schedule-conflict__action-option {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 10px;
  background: #fff;
  border: 1px solid #e9eef5;
  color: #4b5563;
  font-size: 13px;
}

.schedule-conflict__action-option-main {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.schedule-conflict__action-option-main span {
  color: #1f2329;
  font-weight: 600;
}

.schedule-conflict__action-option-main small {
  color: #8c8c8c;
  font-size: 12px;
  line-height: 1.5;
}

.schedule-conflict__action-tip {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border-radius: 10px;
  background: #fff;
  border: 1px solid #e9eef5;
  color: #8c8c8c;
  font-size: 13px;
  line-height: 1.6;
}

.schedule-conflict__action-tip--danger {
  color: #ff7875;
  font-weight: 600;
}

.schedule-conflict__action-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 56px;
  height: 24px;
  padding: 0 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 700;
}

.schedule-conflict__action-badge--danger {
  background: #fff1f0;
  color: #ff4d4f;
}

.schedule-conflict__action-result {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-top: auto;
  padding: 10px 12px;
  border-radius: 10px;
  background: #fff;
  border: 1px dashed #d9e1ea;
  color: #8c8c8c;
  font-size: 13px;
}

.schedule-conflict__action-result-label {
  color: #8c8c8c;
  font-size: 12px;
  font-weight: 600;
}

.schedule-conflict__action-result strong {
  font-size: 14px;
  font-weight: 700;
}

.schedule-conflict__action-result--ready {
  border-color: #b7d9ff;
  background: #f3f9ff;
  color: #1677ff;
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
}
</style>
