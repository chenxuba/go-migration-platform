<script setup>
import { onUnmounted, ref, watch } from 'vue'
import TimetableScheduleHoverPopover from './timetable-schedule-hover-popover.vue'

const props = defineProps({
  spinning: {
    type: Boolean,
    default: false,
  },
  tableDataSource: {
    type: Array,
    default: () => [],
  },
  columns: {
    type: Array,
    default: () => [],
  },
  isSwapTimeGrid: {
    type: Boolean,
    default: false,
  },
  focusedScheduleCellKey: {
    type: String,
    default: '',
  },
  draggingScheduleCellKey: {
    type: String,
    default: '',
  },
  isScheduleColumn: {
    type: Function,
    required: true,
  },
  scheduleCellKey: {
    type: Function,
    required: true,
  },
  scheduleCellStartTime: {
    type: Function,
    required: true,
  },
  scheduleCellEndTime: {
    type: Function,
    required: true,
  },
  scheduleCellContextColumn: {
    type: Function,
    required: true,
  },
  scheduleCellContextRecord: {
    type: Function,
    required: true,
  },
  hasScheduledLesson: {
    type: Function,
    required: true,
  },
  openScheduledLessonDetail: {
    type: Function,
    required: true,
  },
  openScheduledLessonEdit: {
    type: Function,
    required: true,
  },
  openScheduledLessonEditCurrent: {
    type: Function,
    required: true,
  },
  openScheduledLessonCopy: {
    type: Function,
    required: true,
  },
  openScheduledLessonCopyCurrent: {
    type: Function,
    required: true,
  },
  openScheduledConflictDetail: {
    type: Function,
    required: true,
  },
  handleConflictClick: {
    type: Function,
    required: true,
  },
  handleScheduleClick: {
    type: Function,
    required: true,
  },
  consumeScheduledLessonClickSuppressed: {
    type: Function,
    required: true,
  },
  handleSchedulePointerDown: {
    type: Function,
    required: true,
  },
  isScheduleDraggable: {
    type: Function,
    required: true,
  },
  resolveScheduleDragBlockedMessage: {
    type: Function,
    required: true,
  },
  draggingScheduleStyle: {
    type: Object,
    default: () => ({}),
  },
  emptyLessonDragState: {
    type: Function,
    required: true,
  },
  emptyLessonStatusText: {
    type: Function,
    required: true,
  },
  teacherLessonCountLabel: {
    type: Function,
    required: true,
  },
  formatWeek: {
    type: Function,
    required: true,
  },
  formatDate: {
    type: Function,
    required: true,
  },
})

function conflictBadgeTooltip(text) {
  const types = Array.isArray(text?.scheduledConflictTypes)
    ? text.scheduledConflictTypes.filter(Boolean)
    : []
  if (types.length)
    return `冲突原因：${types.join('、')}冲突，点击查看详情`
  return '当前课程存在冲突，点击查看详情'
}

function scheduleStudentText(text) {
  return Array.isArray(text?.studentNames)
    ? text.studentNames.map(item => item?.name).filter(Boolean).join('、')
    : '-'
}

function scheduleLessonTitle(text) {
  const className = String(text?.className || '').trim()
  const courseName = String(text?.courseName || '').trim()
  return className || courseName || '课程'
}

function scheduleLessonSubtitle(text) {
  const className = String(text?.className || '').trim()
  const courseName = String(text?.courseName || '').trim()
  if (className && courseName && className !== courseName)
    return courseName
  return ''
}

function scheduleStudentLine(text, studentName, hasNext) {
  const name = String(studentName || '').trim()
  if (!name)
    return ''
  if (text?.courseType === 1)
    return `${name}${hasNext ? '、' : ''}`
  const courseName = String(text?.courseName || '').trim()
  return `${name}${hasNext ? '、' : ''}${courseName ? `-${courseName}` : ''}`
}

function scheduleClassLine(text) {
  const className = String(text?.className || '').trim()
  const courseName = String(text?.courseName || '').trim()
  if (text?.courseType === 1)
    return className || courseName || '课程'
  if (className && courseName)
    return `${className}-${courseName}`
  return className || courseName || '课程'
}

function scheduleModeLabel(text) {
  if (text?.courseType === 2)
    return text?.isMain === false ? '班课辅教' : '班课主教'
  return '1v1'
}

function scheduleModeShortLabel(text) {
  if (text?.courseType === 2)
    return '班课'
  return '1v1'
}

function scheduleDateText(column, record) {
  const contextRecord = props.scheduleCellContextRecord(column, record)
  const lessonDate = contextRecord?.date
  if (!lessonDate)
    return '-'
  return `${props.formatDate(lessonDate)} ${props.formatWeek(lessonDate)}`
}

function scheduleTimeText(column, record) {
  const contextColumn = props.scheduleCellContextColumn(column, record)
  const slotTitle = String(contextColumn?.title || '').trim()
  const startTime = props.scheduleCellStartTime(column, record)
  const endTime = props.scheduleCellEndTime(column, record)
  const timeRange = startTime && endTime ? `${startTime}-${endTime}` : '-'
  return slotTitle ? `${slotTitle} · ${timeRange}` : timeRange
}

function scheduleConflictText(text) {
  const types = Array.isArray(text?.scheduledConflictTypes)
    ? text.scheduledConflictTypes.filter(Boolean)
    : []
  if (types.length)
    return `${types.join('、')}冲突`
  return text?.scheduledConflict ? '当前课程存在冲突' : ''
}

function scheduleMonthDayText(column, record) {
  const contextRecord = props.scheduleCellContextRecord(column, record)
  const lessonDate = String(contextRecord?.date || '').trim()
  if (!lessonDate)
    return '-'
  const parts = lessonDate.split('-')
  if (parts.length >= 3) {
    const month = Number(parts[1])
    const day = Number(parts[2])
    if (!Number.isNaN(month) && !Number.isNaN(day))
      return `${month}月${day}日`
  }
  return props.formatDate(lessonDate)
}

function scheduleHeaderTimeText(column, record) {
  const startTime = props.scheduleCellStartTime(column, record)
  const endTime = props.scheduleCellEndTime(column, record)
  const contextRecord = props.scheduleCellContextRecord(column, record)
  const lessonDate = contextRecord?.date
  const weekText = lessonDate ? props.formatWeek(lessonDate) : '-'
  return `${startTime} ~ ${endTime} (${weekText}) ${scheduleMonthDayText(column, record)}`
}

function scheduleAssistantSummary(text) {
  const ids = Array.isArray(text?.assistantIds)
    ? text.assistantIds.filter(Boolean)
    : []
  const assistantText = String(text?.assistantText || '').trim()
  if (!assistantText || assistantText === '未安排')
    return '-'
  return `${ids.length || 1}人，${assistantText}`
}

function scheduleStudentSummary(text) {
  const names = Array.isArray(text?.studentNames)
    ? text.studentNames.map(item => item?.name).filter(Boolean)
    : []
  if (!names.length)
    return '-'
  return `${names.length}人，${names.join('、')}`
}

const openSchedulePopoverKey = ref('')
const scheduleCellPressState = ref(null)
let scheduleCellPressMoveHandler = null
let scheduleCellPressUpHandler = null
let suppressedScheduleCellClickKey = ''
let suppressedScheduleCellClickUntil = 0

function schedulePopoverKey(column, record) {
  return String(props.scheduleCellKey(column, record) || '').trim()
}

function clearScheduleCellPressTracking() {
  scheduleCellPressState.value = null
  if (typeof document === 'undefined')
    return
  if (scheduleCellPressMoveHandler)
    document.removeEventListener('mousemove', scheduleCellPressMoveHandler)
  if (scheduleCellPressUpHandler)
    document.removeEventListener('mouseup', scheduleCellPressUpHandler)
  scheduleCellPressMoveHandler = null
  scheduleCellPressUpHandler = null
}

function suppressScheduleCellClick(key, duration = 260) {
  suppressedScheduleCellClickKey = String(key || '').trim()
  suppressedScheduleCellClickUntil = Date.now() + duration
}

function consumeScheduleCellClickSuppressed(key) {
  const normalizedKey = String(key || '').trim()
  if (!normalizedKey)
    return false
  if (normalizedKey === suppressedScheduleCellClickKey && Date.now() <= suppressedScheduleCellClickUntil) {
    suppressedScheduleCellClickKey = ''
    suppressedScheduleCellClickUntil = 0
    return true
  }
  return false
}

function handleSchedulePopoverOpenChange(column, record, open) {
  const key = schedulePopoverKey(column, record)
  openSchedulePopoverKey.value = open ? key : ''
}

function handleSchedulePointerDownWithPopoverClose(event, text, column, record) {
  clearScheduleCellPressTracking()
  openSchedulePopoverKey.value = ''
  if (typeof document !== 'undefined') {
    scheduleCellPressState.value = {
      key: schedulePopoverKey(column, record),
      startX: Number(event?.clientX || 0),
      startY: Number(event?.clientY || 0),
      moved: false,
    }

    scheduleCellPressMoveHandler = (moveEvent) => {
      if (!scheduleCellPressState.value)
        return
      const deltaX = Math.abs(Number(moveEvent?.clientX || 0) - scheduleCellPressState.value.startX)
      const deltaY = Math.abs(Number(moveEvent?.clientY || 0) - scheduleCellPressState.value.startY)
      if (deltaX >= 3 || deltaY >= 3)
        scheduleCellPressState.value.moved = true
    }

    scheduleCellPressUpHandler = () => {
      if (scheduleCellPressState.value?.moved)
        suppressScheduleCellClick(scheduleCellPressState.value.key)
      clearScheduleCellPressTracking()
    }

    document.addEventListener('mousemove', scheduleCellPressMoveHandler)
    document.addEventListener('mouseup', scheduleCellPressUpHandler)
  }
  props.handleSchedulePointerDown(event, text, column, record)
}

function handleScheduleCellClick(text, column, record) {
  if (consumeScheduleCellClickSuppressed(schedulePopoverKey(column, record)))
    return
  if (props.consumeScheduledLessonClickSuppressed())
    return
  props.openScheduledLessonDetail(text, column, record)
}

watch(
  () => props.draggingScheduleCellKey,
  (value) => {
    if (String(value || '').trim())
      openSchedulePopoverKey.value = ''
  },
)

onUnmounted(() => {
  clearScheduleCellPressTracking()
})
</script>

<template>
  <a-spin :spinning="spinning">
    <a-table
      :scroll="{ x: 1300 }"
      :sticky="{ offsetHeader: 100 }"
      size="small"
      :pagination="false"
      bordered
      :data-source="tableDataSource"
      :columns="columns"
    >
      <template #headerCell="{ column }">
        <template v-if="!isSwapTimeGrid && column.startTime && column.endTime">
          <div>{{ column.title }}</div>
          <div class="text-12px text-#666 line-height-2">
            {{ column.startTime }}-{{ column.endTime }}
          </div>
        </template>
        <template v-else-if="isSwapTimeGrid && column.date">
          <div>{{ column.title }}</div>
          <div class="text-12px text-#666 line-height-2">
            {{ column.dateText }}
          </div>
        </template>
        <template v-else>
          {{ column.title }}
        </template>
      </template>

      <template #bodyCell="{ column, record, text }">
        <template v-if="isScheduleColumn(column)">
          <TimetableScheduleHoverPopover
            v-if="hasScheduledLesson(text)"
            :open="!draggingScheduleCellKey && openSchedulePopoverKey === schedulePopoverKey(column, record)"
            :schedule-id="String(text.scheduleId || '')"
            :editable="Boolean(text.scheduleId) && !(text.courseType === 1 && text.isMain === false)"
            :batch-no="String(text.batchNo || '')"
            :batch-size="Number(text.batchSize || 0)"
            :mode-label="scheduleModeShortLabel(text)"
            :lesson-title="scheduleLessonTitle(text)"
            :teacher-name="text.teacherName || record.name || '-'"
            :course-name="scheduleLessonSubtitle(text) || scheduleLessonTitle(text)"
            :assistant-text="scheduleAssistantSummary(text)"
            :student-text="scheduleStudentSummary(text)"
            :time-text="scheduleHeaderTimeText(column, record)"
            :conflict-text="text.scheduledConflict ? scheduleConflictText(text) : ''"
            @open-change="handleSchedulePopoverOpenChange(column, record, $event)"
            @detail="openScheduledLessonDetail(text, scheduleCellContextColumn(column, record), scheduleCellContextRecord(column, record))"
            @copy="payload => openScheduledLessonCopy(text, scheduleCellContextColumn(column, record), scheduleCellContextRecord(column, record), payload)"
            @copy-current="payload => openScheduledLessonCopyCurrent(text, scheduleCellContextColumn(column, record), scheduleCellContextRecord(column, record), payload)"
            @edit="payload => openScheduledLessonEdit(text, scheduleCellContextColumn(column, record), scheduleCellContextRecord(column, record), payload)"
            @edit-current="payload => openScheduledLessonEditCurrent(text, scheduleCellContextColumn(column, record), scheduleCellContextRecord(column, record), payload)"
          >
            <div
              :data-schedule-cell-key="scheduleCellKey(column, record)"
              class="st-schedule-cell st-schedule-cell--unsigned flex h-11 cursor-pointer flex-col rounded-1 text-3"
              :class="{
                'st-schedule-cell--focused': focusedScheduleCellKey === scheduleCellKey(column, record),
                'st-schedule-cell--conflict': text.scheduledConflict,
                'st-schedule-cell--signed': text.callStatusKey === 'signed',
                'st-schedule-cell--draggable': isScheduleDraggable(text),
                'st-schedule-cell--drag-disabled': !isScheduleDraggable(text),
                'st-schedule-cell--floating': draggingScheduleCellKey === scheduleCellKey(column, record),
                'st-schedule-cell--dragging': draggingScheduleCellKey === scheduleCellKey(column, record),
              }"
              :title="!isScheduleDraggable(text) ? resolveScheduleDragBlockedMessage(text) : undefined"
              :style="draggingScheduleCellKey === scheduleCellKey(column, record) ? draggingScheduleStyle : undefined"
              @click="handleScheduleCellClick(text, scheduleCellContextColumn(column, record), scheduleCellContextRecord(column, record))"
              @mousedown.left="handleSchedulePointerDownWithPopoverClose($event, text, scheduleCellContextColumn(column, record), scheduleCellContextRecord(column, record))"
            >
              <div class="st-schedule-cell__header flex h-5 rounded-1 rounded-lb-0 rounded-rb-0 pl1 relative">
                {{ scheduleCellStartTime(column, record) }}-{{ scheduleCellEndTime(column, record) }}
                <a-tooltip v-if="text.scheduledConflict" :title="conflictBadgeTooltip(text)" placement="top">
                  <span
                    class="st-schedule-cell__badge st-schedule-cell__badge--edge st-schedule-cell__badge--conflict absolute right-0 h-4 rounded-lb-2 rounded-rt-1 pl-2 pr-1 text-2.5 font-500"
                    style="cursor: pointer"
                    @click.stop="openScheduledConflictDetail(text)"
                  >
                    冲突
                  </span>
                </a-tooltip>
                <span
                  v-else
                  class="st-schedule-cell__badge st-schedule-cell__badge--edge st-schedule-cell__badge--mode absolute right-0 h-4 rounded-lb-2 rounded-rt-1 pl-2 pr-1 text-2.5 font-500"
                >
                  <span v-if="text.courseType === 1">1v1</span>
                  <span v-else-if="text.courseType === 2">班课(<span>{{ text.isMain ? '主教' : '辅教' }}</span>)</span>
                </span>
              </div>

              <div v-if="!text.classId" class="st-schedule-cell__body st-schedule-cell__body--students flex flex-1 flex-items-center pl-1">
                <span v-for="(item, index) in text.studentNames" :key="index">
                  <div class="flex">{{ scheduleStudentLine(text, item.name, index !== text.studentNames.length - 1) }}</div>
                </span>
              </div>

              <div v-else class="st-schedule-cell__body st-schedule-cell__body--class flex flex-1 flex-items-center pl-1 line-height-4">
                <div class="flex">
                  {{ scheduleClassLine(text) }}
                </div>
              </div>
            </div>
          </TimetableScheduleHoverPopover>

          <div
            v-else
            :data-empty-schedule-cell-key="scheduleCellKey(column, record)"
            :data-drag-target-teacher-id="record?.teacherId || ''"
            :data-drag-target-teacher-name="record?.name || ''"
            :data-drag-target-lesson-date="scheduleCellContextRecord(column, record)?.date || ''"
            :data-drag-target-start-time="scheduleCellContextColumn(column, record)?.startTime || ''"
            :data-drag-target-end-time="scheduleCellContextColumn(column, record)?.endTime || ''"
            class="st-empty-cell h-11 rounded-1 text-3 flex-center cursor-pointer"
            :title="emptyLessonDragState(column, record)?.message || undefined"
            :class="[
              emptyLessonDragState(column, record)?.checking
                ? 'st-empty-cell--drag-checking'
                : emptyLessonDragState(column, record)?.valid === true
                  ? 'st-empty-cell--drag-valid'
                  : emptyLessonDragState(column, record)?.valid === false
                    ? 'st-empty-cell--drag-invalid'
                    : emptyLessonStatusText(text)
                      ? (text.conflict ? 'bg-#ffe6e6 text-#a31616' : 'bg-#e6ffe6 text-#16a34a')
                      : 'st-empty-cell--idle',
            ]"
            @click="text.conflict ? handleConflictClick(text, scheduleCellContextColumn(column, record), scheduleCellContextRecord(column, record)) : handleScheduleClick(text, scheduleCellContextColumn(column, record), scheduleCellContextRecord(column, record))"
          >
            {{ emptyLessonDragState(column, record)?.label || emptyLessonStatusText(text) }}
          </div>
        </template>

        <template v-if="column.key === 'date'">
          <div class="text-3.5 ">
            {{ formatWeek(text) }}
          </div>
          <div class="text-3 font-500 line-height-3 text-#666">
            {{ formatDate(text) }}
          </div>
        </template>

        <template v-if="column.key === 'name'">
          <div>{{ text }}</div>
          <div class="text-3 text-#666 leading-snug">
            {{ teacherLessonCountLabel(record.teacherId) }}
          </div>
        </template>

        <template v-if="column.key === 'slot'">
          <div class="text-3.5">
            {{ text }}
          </div>
          <div class="text-3 font-500 line-height-3 text-#666">
            {{ record.startTime }}-{{ record.endTime }}
          </div>
        </template>
      </template>
    </a-table>
  </a-spin>
</template>

<style scoped lang="less">
:deep(.ant-table-wrapper),
:deep(.ant-table-container) {
  border-radius: 0 !important;
}

:deep(.ant-table-thead > tr:first-child > th:first-child),
:deep(.ant-table-thead > tr:first-child > th:last-child) {
  border-radius: 0 !important;
}

:deep(td.ant-table-cell.ant-table-cell-row-hover) {
  background-color: rgb(231, 236, 255) !important;
}

:deep(td.ant-table-cell) {
  padding: 4px !important;
}

.st-schedule-cell {
  position: relative;
  overflow: hidden;
  background: rgba(78, 109, 255, 0.12);
  color: #fff;
  user-select: none;
  transition: box-shadow 0.25s ease, transform 0.25s ease, opacity 0.2s ease;
}

.st-schedule-cell--unsigned {
  background: rgba(78, 109, 255, 0.12);
}

.st-schedule-cell--signed {
  background: linear-gradient(180deg, #fbfbfc 0%, #f2f3f5 100%);
  box-shadow: 0 4px 12px rgba(15, 23, 42, 0.06);
}

.st-schedule-cell__header {
  align-items: center;
  background: #06f;
  color: #fff;
}

.st-schedule-cell--signed .st-schedule-cell__header {
  background: linear-gradient(180deg, #aab0bb 0%, #9ba2ae 100%);
}

.st-schedule-cell__badge {
  color: #fff;
}

.st-schedule-cell__badge--edge {
  top: -1px;
  right: -1px;
}

.st-schedule-cell__badge--mode {
  background: rgba(0, 0, 0, 0.5);
}

.st-schedule-cell--signed .st-schedule-cell__badge--mode {
  background: rgba(0, 0, 0, 0.42);
}

.st-schedule-cell--conflict {
  box-shadow: 0 0 0 2px rgba(255, 77, 79, 0.4);
}

.st-schedule-cell--draggable {
  cursor: grab;
}

.st-schedule-cell--drag-disabled {
  cursor: not-allowed;
}

.st-schedule-cell--dragging {
  cursor: grabbing;
}

.st-schedule-cell--floating {
  margin: 0 !important;
  pointer-events: none;
  box-shadow:
    0 18px 40px rgba(31, 35, 41, 0.2),
    0 8px 18px rgba(31, 35, 41, 0.12);
  transform: rotate(-1deg);
  z-index: 1200;
}

.st-schedule-cell__badge--conflict {
  background: #ff4d4f;
}

.st-schedule-cell__body {
  color: #002cfd;
}

.st-schedule-cell--signed .st-schedule-cell__body {
  color: #707784;
  font-weight: 600;
}

.st-schedule-cell--focused {
  animation: st-schedule-cell-flash 0.5s ease-in-out 6;
  box-shadow:
    0 0 0 3px rgba(255, 77, 79, 0.98),
    0 0 0 8px rgba(255, 77, 79, 0.26),
    0 0 20px rgba(255, 77, 79, 0.5);
  transform: scale(1.015);
  z-index: 2;
}

:deep(.st-schedule-cell-popover .ant-popover-inner) {
  padding: 0 !important;
  border-radius: 8px;
  overflow: hidden;
  box-shadow:
    0 14px 32px rgba(15, 23, 42, 0.14),
    0 4px 12px rgba(15, 23, 42, 0.08);
}

:deep(.st-schedule-cell-popover .ant-popover-inner-content) {
  padding: 0 !important;
}

.st-schedule-hover-card {
  width: 344px;
  max-width: min(344px, 90vw);
  min-height: 273px;
  background: #fff;
  border-radius: 8px;
  overflow: hidden;
}

.st-schedule-hover-card__header {
  padding: 0 0 1px;
  background: linear-gradient(135deg, #166dff 0%, #1d98ff 100%);
}

.st-schedule-hover-card__hero {
  display: flex;
  gap: 14px;
  align-items: flex-start;
  padding: 16px 18px 14px;
  color: #fff;
}

.st-schedule-hover-card__badge-shell {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 46px;
  height: 46px;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: 0 8px 18px rgba(7, 55, 143, 0.16);
}

.st-schedule-hover-card__badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 8px;
  background: linear-gradient(180deg, #ff8a85 0%, #ff5353 100%);
  color: #fff;
  font-size: 9px;
  font-weight: 700;
  line-height: 1;
}

.st-schedule-hover-card__hero-main {
  min-width: 0;
  flex: 1;
}

.st-schedule-hover-card__hero-top {
  display: flex;
  gap: 12px;
  align-items: flex-start;
  justify-content: space-between;
}

.st-schedule-hover-card__detail-link {
  padding: 0;
  border: 0;
  background: transparent;
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  line-height: 24px;
  cursor: pointer;
  white-space: nowrap;
}

.st-schedule-hover-card__detail-link::after {
  content: ' >';
}

.st-schedule-hover-card__title {
  overflow: hidden;
  color: #fff;
  font-size: 16px;
  font-weight: 700;
  line-height: 24px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.st-schedule-hover-card__time {
  margin-top: 4px;
  overflow: hidden;
  color: rgba(255, 255, 255, 0.96);
  font-size: 13px;
  font-weight: 600;
  line-height: 18px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.st-schedule-hover-card__body {
  display: flex;
  flex-direction: column;
  gap: 0;
  padding: 10px 18px 2px;
}

.st-schedule-hover-card__row {
  display: grid;
  grid-template-columns: max-content minmax(0, 1fr);
  column-gap: 8px;
  row-gap: 0;
  align-items: start;
  font-size: 12px;
  line-height: 22px;
}

.st-schedule-hover-card__row > span {
  color: #8f8f8f;
  font-weight: 400;
}

.st-schedule-hover-card__row > strong {
  overflow: hidden;
  color: #6c6c6c;
  font-weight: 400;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}

.st-schedule-hover-card__value--primary {
  color: #166dff !important;
}

.st-schedule-hover-card__row--danger > strong {
  color: #cf1322;
}

.st-schedule-hover-card__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 18px 14px;
  margin-top: auto;
}

.st-schedule-hover-card__actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.st-schedule-hover-card__icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  padding: 0;
  border: 0;
  border-radius: 50%;
  background: transparent;
  color: #9f9f9f;
  font-size: 18px;
  cursor: pointer;
  transition: background-color 0.18s ease, color 0.18s ease;
}

.st-schedule-hover-card__icon-btn:hover,
.st-schedule-hover-card__icon-btn--active {
  background: #e8f1ff;
  color: #166dff;
}

.st-schedule-hover-card__primary-btn {
  width: 74px;
  min-width: 74px;
  height: 28px;
  padding: 0;
  border: 0;
  border-radius: 6px;
  background: linear-gradient(180deg, #1970ff 0%, #1660e8 100%);
  color: #fff;
  font-size: 12px;
  font-weight: 700;
  line-height: 28px;
  cursor: pointer;
  box-shadow: 0 6px 14px rgba(22, 96, 232, 0.18);
}

.st-empty-cell {
  transition:
    background-color 0.18s ease,
    color 0.18s ease,
    box-shadow 0.18s ease,
    transform 0.18s ease;
}

.st-empty-cell--idle {
  background: transparent;
  color: transparent;
  box-shadow: none;
}

.st-empty-cell--drag-checking {
  background: #fff7e6;
  color: #d48806;
  box-shadow: inset 0 0 0 1px rgba(250, 173, 20, 0.28);
}

.st-empty-cell--drag-valid {
  background: #fff7e6;
  color: #d48806;
  box-shadow:
    inset 0 0 0 1px rgba(250, 173, 20, 0.34),
    0 0 0 2px rgba(250, 173, 20, 0.12);
}

.st-empty-cell--drag-invalid {
  background: #ffe1df;
  color: #cf1322;
  box-shadow:
    inset 0 0 0 1px rgba(255, 77, 79, 0.4),
    0 0 0 2px rgba(255, 77, 79, 0.16);
}

@keyframes st-schedule-cell-flash {
  0%,
  100% {
    box-shadow:
      0 0 0 3px rgba(255, 77, 79, 0.98),
      0 0 0 8px rgba(255, 77, 79, 0.26),
      0 0 20px rgba(255, 77, 79, 0.5);
  }

  50% {
    box-shadow:
      0 0 0 3px rgba(255, 77, 79, 0.25),
      0 0 0 8px rgba(255, 77, 79, 0.08),
      0 0 10px rgba(255, 77, 79, 0.18);
  }
}
</style>
