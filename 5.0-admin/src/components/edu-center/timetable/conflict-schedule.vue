<script setup lang="ts">
import dayjs from 'dayjs'
import { Modal } from 'ant-design-vue'
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import type { TeachingScheduleItem } from '@/api/edu-center/teaching-schedule'
import { cancelTeachingSchedulesApi, listTeachingSchedulesApi } from '@/api/edu-center/teaching-schedule'
import emitter, { EVENTS } from '@/utils/eventBus'
import messageService from '@/utils/messageService'

type ConflictFilterKey = 'class' | 'teacher' | 'classroom' | 'student' | 'assistant'
type ConflictType = '班级' | '老师' | '教室' | '学员' | '助教'

const weekdayLabels = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
const conflictTypeOrder: ConflictType[] = ['班级', '老师', '教室', '学员', '助教']

const displayArray = ['createTime', 'scheduleType']

const conflictFilterTypeMap: Record<ConflictFilterKey, ConflictType> = {
  class: '班级',
  teacher: '老师',
  classroom: '教室',
  student: '学员',
  assistant: '助教',
}

const scheduleTypeOptions = [
  { id: 'class', value: '班级冲突' },
  { id: 'teacher', value: '上课教师冲突' },
  { id: 'classroom', value: '上课教室冲突' },
  { id: 'student', value: '学员冲突' },
  { id: 'assistant', value: '助教冲突' },
]

const tableColumns = [
  {
    title: '上课日期/时段',
    key: 'dateTime',
    width: 220,
  },
  {
    title: '日程类型',
    key: 'scheduleType',
    width: 130,
  },
  {
    title: '冲突类型',
    key: 'conflictType',
    width: 130,
  },
  {
    title: '所属班级/1对1',
    key: 'teachingClass',
    width: 210,
  },
  {
    title: '所属课程',
    key: 'lessonName',
    width: 130,
  },
  {
    title: '上课教师',
    key: 'teacher',
    width: 120,
  },
  {
    title: '教室',
    key: 'classroom',
    width: 100,
  },
  {
    title: '操作',
    key: 'actions',
    width: 130,
  },
]

function getWeekRange(base = dayjs()) {
  const current = dayjs(base)
  const diff = current.day() === 0 ? -6 : 1 - current.day()
  const start = current.add(diff, 'day').startOf('day')
  const end = start.add(6, 'day')
  return [
    start.format('YYYY-MM-DD'),
    end.format('YYYY-MM-DD'),
  ]
}

function normalizeFilterValues(value: unknown) {
  if (!Array.isArray(value))
    return []
  return value.map(item => String(item ?? '').trim()).filter(Boolean)
}

const defaultCreateTimeVals = getWeekRange()
const queryRange = ref<string[]>([...defaultCreateTimeVals])
const selectedConflictKeys = ref<ConflictFilterKey[]>([])
const loading = ref(false)
const deletingId = ref('')
const scheduleRows = ref<TeachingScheduleItem[]>([])
let requestToken = 0

const selectedConflictTypes = computed(() =>
  selectedConflictKeys.value
    .map(key => conflictFilterTypeMap[key])
    .filter(Boolean),
)

const conflictRows = computed(() => {
  const filtered = scheduleRows.value.filter(item => item?.conflict)

  return filtered.slice().sort((a, b) => {
    const timeDiff = dayjs(a.startAt).valueOf() - dayjs(b.startAt).valueOf()
    if (timeDiff !== 0)
      return timeDiff
    return String(a.id || '').localeCompare(String(b.id || ''))
  })
})

function handleCreateTimeFilter(value: unknown) {
  if (Array.isArray(value) && value.length === 2) {
    queryRange.value = [
      String(value[0] || ''),
      String(value[1] || ''),
    ]
    return
  }
  queryRange.value = [...defaultCreateTimeVals]
}

function handleConflictTypeFilter(value: unknown) {
  selectedConflictKeys.value = normalizeFilterValues(value) as ConflictFilterKey[]
}

async function loadConflictSchedules() {
  const [startDate, endDate] = queryRange.value
  const token = ++requestToken
  loading.value = true
  try {
    const res = await listTeachingSchedulesApi({
      startDate,
      endDate,
      conflictTypes: selectedConflictTypes.value.join(',') || undefined,
    })
    if (token !== requestToken)
      return
    scheduleRows.value = res.code === 200 && Array.isArray(res.result) ? res.result : []
  }
  catch (error) {
    if (token !== requestToken)
      return
    console.error('load conflict schedules failed', error)
    scheduleRows.value = []
    messageService.error(error?.response?.data?.message || error?.message || '加载冲突日程失败')
  }
  finally {
    if (token === requestToken)
      loading.value = false
  }
}

watch(
  [queryRange, selectedConflictTypes],
  () => {
    loadConflictSchedules()
  },
  { deep: true, immediate: true },
)

function handleExternalRefresh() {
  loadConflictSchedules()
}

onMounted(() => {
  emitter.on(EVENTS.REFRESH_DATA, handleExternalRefresh)
})

onUnmounted(() => {
  emitter.off(EVENTS.REFRESH_DATA, handleExternalRefresh)
})

function formatLessonDate(record: Record<string, any>) {
  const value = dayjs(record.lessonDate)
  return `${value.format('YYYY-MM-DD')}（${weekdayLabels[value.day()] || ''}）`
}

function formatLessonTime(record: Record<string, any>) {
  return `${dayjs(record.startAt).format('HH:mm')} ~ ${dayjs(record.endAt).format('HH:mm')}`
}

function conflictTypesForRecord(record: Record<string, any>) {
  const source = Array.isArray(record.conflictTypes) ? record.conflictTypes : []
  const normalized = source
    .map(item => String(item || '').trim())
    .filter(Boolean)
  const unique = Array.from(new Set(normalized))
  const ordered = conflictTypeOrder.filter(type => unique.includes(type))
  const rest = unique.filter(type => !conflictTypeOrder.includes(type as ConflictType))
  return [...ordered, ...rest]
}

function conflictTypeText(type: string) {
  return `${type}冲突`
}

function conflictTypeChipClass(type: string) {
  if (type === '班级')
    return 'conflict-type-chip--class'
  if (type === '老师')
    return 'conflict-type-chip--teacher'
  if (type === '教室')
    return 'conflict-type-chip--classroom'
  if (type === '学员')
    return 'conflict-type-chip--student'
  if (type === '助教')
    return 'conflict-type-chip--assistant'
  return ''
}

function conflictTypeDescription(type: string, record: Record<string, any>) {
  const className = String(record.teachingClassName || '').trim() || '当前班级'
  const teacherName = String(record.teacherName || '').trim() || '当前老师'
  const classroomName = String(record.classroomName || '').trim()
  const assistantText = Array.isArray(record.assistantNames) && record.assistantNames.length
    ? record.assistantNames.join('、')
    : '当前助教'

  if (type === '班级')
    return `同一时间 ${className} 已有其他日程安排。`
  if (type === '老师')
    return `同一时间老师 ${teacherName} 已有其他日程安排。`
  if (type === '教室')
    return classroomName
      ? `同一时间教室 ${classroomName} 已有其他日程安排。`
      : '同一时间存在教室占用冲突。'
  if (type === '学员')
    return `同一时间该日程内至少 1 位学员已有其他日程安排。`
  if (type === '助教')
    return `同一时间助教 ${assistantText} 已有其他日程安排。`
  return `同一时间存在${type}相关冲突。`
}

function scheduleTypeLabel(record: Record<string, any>) {
  return Number(record.classType) === 2 ? '1对1日程' : '班课日程'
}

function teachingClassBadge(record: Record<string, any>) {
  return Number(record.classType) === 2 ? '1v1' : '班课'
}

function teachingClassBadgeClass(record: Record<string, any>) {
  return Number(record.classType) === 2
    ? 'text-#ff7d7d bg-#fff1f1 rounded-2.5 inline-block text-2.5 pt-0.5 pb-0.5 pl-1.5 pr-1.5'
    : 'text-#d46b08 bg-#fff5e8 rounded-2.5 inline-block text-2.5 pt-0.5 pb-0.5 pl-1.5 pr-1.5'
}

function isOneToOneSchedule(record: Record<string, any>) {
  return Number(record.classType) === 2
}

function isSameScheduleSlot(left: Record<string, any>, right: Record<string, any>) {
  return String(left.lessonDate || '') === String(right.lessonDate || '')
    && String(left.startAt || '') === String(right.startAt || '')
    && String(left.endAt || '') === String(right.endAt || '')
}

function hasAssistantOverlap(left: Record<string, any>, right: Record<string, any>) {
  const leftSet = new Set((Array.isArray(left.assistantIds) ? left.assistantIds : []).map(item => String(item || '').trim()).filter(Boolean))
  const rightList = (Array.isArray(right.assistantIds) ? right.assistantIds : []).map(item => String(item || '').trim()).filter(Boolean)
  return rightList.some(item => leftSet.has(item))
}

function parseScheduleStudentIds(value: unknown) {
  return new Set(
    String(value || '')
      .split(',')
      .map(item => item.trim())
      .filter(item => item && item !== '0'),
  )
}

function hasStudentOverlap(left: Record<string, any>, right: Record<string, any>) {
  const leftSet = parseScheduleStudentIds(left.studentId)
  if (!leftSet.size)
    return false
  return [...parseScheduleStudentIds(right.studentId)].some(item => leftSet.has(item))
}

function findLinkedConflictRows(record: Record<string, any>) {
  return conflictRows.value.filter(item =>
    String(item.id || '') !== String(record.id || '')
    && isSameScheduleSlot(item, record)
    && (
      (item.teachingClassId && item.teachingClassId === record.teachingClassId)
      || hasStudentOverlap(item, record)
      || (item.teacherId && item.teacherId === record.teacherId)
      || (item.classroomId && item.classroomId === record.classroomId)
      || hasAssistantOverlap(item, record)
    ),
  )
}

function buildDeleteSuccessMessage(record: Record<string, any>, linkedCount = 0) {
  const dateText = dayjs(record.lessonDate).format('M月D日')
  const scheduleLabel = isOneToOneSchedule(record) ? '1对1' : '班课'
  if (linkedCount > 0)
    return `已删除 1 条日程；同冲突中的另外 ${linkedCount} 条因不再冲突，会从“冲突日程”列表移除，但不会被删除`
  return `已删除 ${dateText} ${formatLessonTime(record)} 的${scheduleLabel}冲突日程`
}

function confirmDelete(record: Record<string, any>) {
  const isOneToOne = isOneToOneSchedule(record)
  const linkedConflictRows = findLinkedConflictRows(record)
  const linkedHint = linkedConflictRows.length
    ? `删除当前这条后，和它成对冲突的另外 ${linkedConflictRows.length} 条记录如果不再冲突，会一起从“冲突日程”列表消失，但不会被删除。`
    : ''
  const deleteImpact = isOneToOne
    ? '删除后会同步从当前 1 对 1 的排课记录中移除，该操作不可恢复。'
    : '删除后会同步从当前班课的主教与全部助教课表中移除，仅删除这一节，不会删除整批班课。'

  Modal.confirm({
    centered: true,
    title: '确认删除这条冲突日程？',
    content: `${deleteImpact}${linkedHint}`,
    okText: '删除',
    cancelText: '取消',
    okButtonProps: {
      danger: true,
    },
    onOk: async () => {
      deletingId.value = String(record.id || '')
      try {
        const res = await cancelTeachingSchedulesApi({
          ids: [String(record.id)],
        })
        if (res.code !== 200)
          throw new Error(res.message || '删除冲突日程失败')
        messageService.success(buildDeleteSuccessMessage(record, linkedConflictRows.length))
        emitter.emit(EVENTS.REFRESH_DATA)
      }
      catch (error) {
        console.error('delete conflict schedule failed', error)
        messageService.error(error?.response?.data?.message || error?.message || '删除冲突日程失败')
        throw error
      }
      finally {
        deletingId.value = ''
      }
    },
  })
}
</script>

<template>
  <div class="conflict-page">
    <div class="filter-wrap bg-white pl-3 pr-3 rounded-4 rounded-lt-0 rounded-rt-0">
      <all-filter
        type="noDelCreateTime"
        :display-array="displayArray"
        :default-create-time-vals="defaultCreateTimeVals"
        create-time-label="上课日期"
        :schedule-type-options="scheduleTypeOptions"
        schedule-type-label="冲突类型"
        @update:create-time-filter="handleCreateTimeFilter"
        @update:schedule-type-filter="handleConflictTypeFilter"
      />
    </div>

    <div class="conflict-board mt-2 bg-white rounded-4">
      <div class="conflict-board__head">
        <div class="conflict-board__title">
          共 {{ conflictRows.length }} 条冲突日程
        </div>
      </div>

      <a-table
        row-key="id"
        size="small"
        :columns="tableColumns"
        :data-source="conflictRows"
        :pagination="false"
        :loading="loading"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'dateTime'">
            <div class="flex flex-col">
              <div class="text-3.5 text-#333 whitespace-nowrap">
                {{ formatLessonDate(record) }}
              </div>
              <div class="text-3 text-#888 whitespace-nowrap">
                {{ formatLessonTime(record) }}
              </div>
            </div>
          </template>

          <template v-else-if="column.key === 'scheduleType'">
            <div class="text-#222 whitespace-nowrap">
              {{ scheduleTypeLabel(record) }}
            </div>
          </template>

          <template v-else-if="column.key === 'conflictType'">
            <div v-if="conflictTypesForRecord(record).length" class="conflict-type-cell">
              <a-tooltip placement="topLeft">
                <template #title>
                  <div class="conflict-type-tooltip">
                    <div class="conflict-type-tooltip__title">
                      冲突详情
                    </div>
                    <div
                      v-for="type in conflictTypesForRecord(record)"
                      :key="type"
                      class="conflict-type-tooltip__item"
                    >
                      <strong class="conflict-type-tooltip__item-label">{{ conflictTypeText(type) }}</strong>
                      <span class="conflict-type-tooltip__item-desc">{{ conflictTypeDescription(type, record) }}</span>
                    </div>
                  </div>
                </template>

                <div class="conflict-type-cell__chips">
                  <span
                    v-for="type in conflictTypesForRecord(record)"
                    :key="type"
                    class="conflict-type-chip"
                    :class="conflictTypeChipClass(type)"
                  >
                    {{ conflictTypeText(type) }}
                  </span>
                </div>
              </a-tooltip>
            </div>
            <div v-else class="text-#bbb whitespace-nowrap">
              -
            </div>
          </template>

          <template v-else-if="column.key === 'teachingClass'">
            <div class="flex items-center gap-1.5">
              <span :class="teachingClassBadgeClass(record)">
                {{ teachingClassBadge(record) }}
              </span>
              <span class="text-#ff6b6b whitespace-nowrap">
                {{ record.teachingClassName || '-' }}
              </span>
            </div>
          </template>

          <template v-else-if="column.key === 'lessonName'">
            <div class="text-#222 whitespace-nowrap">
              {{ record.lessonName || '-' }}
            </div>
          </template>

          <template v-else-if="column.key === 'teacher'">
            <div class="text-#222 whitespace-nowrap">
              {{ record.teacherName || '-' }}
            </div>
          </template>

          <template v-else-if="column.key === 'classroom'">
            <div class="text-#222 whitespace-nowrap">
              {{ record.classroomName || '-' }}
            </div>
          </template>

          <template v-else-if="column.key === 'actions'">
            <a-space :size="8">
              <a-tooltip title="编辑暂未开放">
                <span>
                  <a-button type="link" size="small" disabled>
                    编辑
                  </a-button>
                </span>
              </a-tooltip>
              <a-button
                type="link"
                size="small"
                danger
                :loading="deletingId === String(record.id)"
                @click="confirmDelete(record)"
              >
                删除
              </a-button>
            </a-space>
          </template>
        </template>
      </a-table>
    </div>
  </div>
</template>

<style scoped lang="less">
.conflict-page {
  padding-top: 0;
}

.conflict-board {
  padding: 16px 18px 18px;
  box-shadow: 0 10px 28px rgba(15, 35, 95, 0.04);
}

.conflict-board__head {
  margin-bottom: 16px;
}

.conflict-board__title {
  position: relative;
  padding-left: 14px;
  color: #222;
  font-size: 14px;
}

.conflict-board__title::before {
  position: absolute;
  left: 0;
  top: 5px;
  width: 5px;
  height: 14px;
  border-radius: 999px;
  background: #1677ff;
  content: '';
}

.conflict-type-cell {
  max-width: 220px;
}

.conflict-type-cell__chips {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  cursor: help;
}

.conflict-type-chip {
  display: inline-flex;
  align-items: center;
  min-height: 24px;
  padding: 0 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
  line-height: 1;
  white-space: nowrap;
}

.conflict-type-chip--class {
  color: #ad6800;
  background: #fff7e6;
}

.conflict-type-chip--teacher {
  color: #cf1322;
  background: #fff1f0;
}

.conflict-type-chip--classroom {
  color: #0958d9;
  background: #e6f4ff;
}

.conflict-type-chip--student {
  color: #c41d7f;
  background: #fff0f6;
}

.conflict-type-chip--assistant {
  color: #531dab;
  background: #f9f0ff;
}

.conflict-type-tooltip {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-width: 340px;
}

.conflict-type-tooltip__title {
  color: #fff;
  font-size: 13px;
  font-weight: 700;
}

.conflict-type-tooltip__item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.conflict-type-tooltip__item-label {
  color: #fff;
  font-size: 12px;
  line-height: 18px;
}

.conflict-type-tooltip__item-desc {
  color: rgba(255, 255, 255, 0.85);
  font-size: 12px;
  line-height: 18px;
  white-space: normal;
}

</style>
