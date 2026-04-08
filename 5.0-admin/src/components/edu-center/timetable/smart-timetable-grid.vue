<script setup>
import { CopyOutlined, EditOutlined } from '@ant-design/icons-vue'
import { useRouter } from 'vue-router'

const router = useRouter()

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
  openScheduledLessonDetail: {
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
  handleSchedulePointerDown: {
    type: Function,
    required: true,
  },
  isScheduleDraggable: {
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

const schedulePopoverInnerStyle = {
  padding: '0px',
}

function goRollCall() {
  router.push('/edu-center/roll-call-list')
}
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
          <a-popover
            v-if="text.studentId"
            trigger="hover"
            placement="rightTop"
            overlay-class-name="st-schedule-cell-popover"
            :overlay-inner-style="schedulePopoverInnerStyle"
            :mouse-enter-delay="0.12"
            :mouse-leave-delay="0.06"
          >
            <template #content>
              <div class="st-schedule-hover-card">
                <div class="st-schedule-hover-card__header">
                  <div class="st-schedule-hover-card__hero">
                    <div class="st-schedule-hover-card__badge-shell">
                      <div class="st-schedule-hover-card__badge">
                        {{ scheduleModeShortLabel(text) }}
                      </div>
                    </div>

                    <div class="st-schedule-hover-card__hero-main">
                      <div class="st-schedule-hover-card__hero-top">
                        <div class="st-schedule-hover-card__title" :title="scheduleLessonTitle(text)">
                          {{ scheduleLessonTitle(text) }}
                        </div>
                        <button
                          type="button"
                          class="st-schedule-hover-card__detail-link"
                          @click.stop="openScheduledLessonDetail(text, scheduleCellContextColumn(column, record), scheduleCellContextRecord(column, record))"
                        >
                          详情
                        </button>
                      </div>
                      <div class="st-schedule-hover-card__time" :title="scheduleHeaderTimeText(column, record)">
                        {{ scheduleHeaderTimeText(column, record) }}
                      </div>
                    </div>
                  </div>
                </div>

                <div class="st-schedule-hover-card__body">
                  <div class="st-schedule-hover-card__row">
                    <span>上课教师：</span>
                    <strong :title="text.teacherName || record.name || '-'">{{ text.teacherName || record.name || '-' }}</strong>
                  </div>
                  <div class="st-schedule-hover-card__row">
                    <span>课程：</span>
                    <strong :title="scheduleLessonSubtitle(text) || scheduleLessonTitle(text)">{{ scheduleLessonSubtitle(text) || scheduleLessonTitle(text) }}</strong>
                  </div>
                  <div class="st-schedule-hover-card__row">
                    <span>上课助教：</span>
                    <strong :title="scheduleAssistantSummary(text)">{{ scheduleAssistantSummary(text) }}</strong>
                  </div>
                  <div class="st-schedule-hover-card__row">
                    <span>上课学员：</span>
                    <strong class="st-schedule-hover-card__value--primary" :title="scheduleStudentSummary(text)">{{ scheduleStudentSummary(text) }}</strong>
                  </div>
                  <div class="st-schedule-hover-card__row">
                    <span>试听学员：</span>
                    <strong>-</strong>
                  </div>
                  <div class="st-schedule-hover-card__row">
                    <span>请假学员：</span>
                    <strong>-</strong>
                  </div>
                  <div class="st-schedule-hover-card__row">
                    <span>对内备注：</span>
                    <strong>-</strong>
                  </div>
                  <div v-if="text.scheduledConflict" class="st-schedule-hover-card__row st-schedule-hover-card__row--danger">
                    <span>冲突说明：</span>
                    <strong :title="scheduleConflictText(text)">{{ scheduleConflictText(text) }}</strong>
                  </div>
                </div>

                <div class="st-schedule-hover-card__footer">
                  <div class="st-schedule-hover-card__actions">
                    <a-tooltip title="编辑日程" placement="top">
                      <button
                        type="button"
                        class="st-schedule-hover-card__icon-btn"
                        @click.stop="openScheduledLessonDetail(text, scheduleCellContextColumn(column, record), scheduleCellContextRecord(column, record))"
                      >
                        <EditOutlined />
                      </button>
                    </a-tooltip>

                    <a-tooltip title="复制日程" placement="top">
                      <button
                        type="button"
                        class="st-schedule-hover-card__icon-btn"
                        @click.stop
                      >
                        <CopyOutlined />
                      </button>
                    </a-tooltip>
                  </div>

                  <button
                    type="button"
                    class="st-schedule-hover-card__primary-btn"
                    @click.stop="goRollCall()"
                  >
                    去点名
                  </button>
                </div>
              </div>
            </template>

            <div
              :data-schedule-cell-key="scheduleCellKey(column, record)"
              class="st-schedule-cell flex flex-col bg-#4e6dff1f h-11 rounded-1 text-3 text-#fff cursor-pointer"
              :class="{
                'st-schedule-cell--focused': focusedScheduleCellKey === scheduleCellKey(column, record),
                'st-schedule-cell--conflict': text.scheduledConflict,
                'st-schedule-cell--draggable': isScheduleDraggable(text),
                'st-schedule-cell--drag-disabled': !isScheduleDraggable(text) && text.courseType === 1 && text.isMain === false,
                'st-schedule-cell--floating': draggingScheduleCellKey === scheduleCellKey(column, record),
                'st-schedule-cell--dragging': draggingScheduleCellKey === scheduleCellKey(column, record),
              }"
              :title="!isScheduleDraggable(text) && text.courseType === 1 && text.isMain === false ? '助教课暂不支持拖拽调课，请在主教老师所在行操作' : undefined"
              :style="draggingScheduleCellKey === scheduleCellKey(column, record) ? draggingScheduleStyle : undefined"
              @click="openScheduledLessonDetail(text, scheduleCellContextColumn(column, record), scheduleCellContextRecord(column, record))"
              @mousedown.left="handleSchedulePointerDown($event, text, scheduleCellContextColumn(column, record), scheduleCellContextRecord(column, record))"
            >
              <div class="pl1 bg-#06f rounded-1 rounded-lb-0 rounded-rb-0 flex relative h-5">
                {{ scheduleCellStartTime(column, record) }}-{{ scheduleCellEndTime(column, record) }}
                <a-tooltip v-if="text.scheduledConflict" :title="conflictBadgeTooltip(text)" placement="top">
                  <span
                    class="absolute right-0 pl-2 pr-1 h-4 text-#fff text-2.5 font-500 rounded-rt-1 rounded-lb-2 st-schedule-cell__badge--conflict"
                    style="cursor: pointer"
                    @click.stop="openScheduledConflictDetail(text)"
                  >
                    冲突
                  </span>
                </a-tooltip>
                <span
                  v-else
                  class="absolute right-0 pl-2 pr-1 h-4 text-#fff text-2.5 font-500 rounded-rt-1 rounded-lb-2 bg-#00000080"
                >
                  <span v-if="text.courseType === 1">1v1</span>
                  <span v-else-if="text.courseType === 2">班课(<span>{{ text.isMain ? '主教' : '辅教' }}</span>)</span>
                </span>
              </div>

              <div v-if="!text.classId" class="flex pl-1 flex-1 text-#002cfd flex-items-center">
                <span v-for="(item, index) in text.studentNames" :key="index">
                  <div class="flex">{{ item.name }}{{ index !== text.studentNames.length - 1 ? '、' : '' }}-{{ text.courseName }}</div>
                </span>
              </div>

              <div v-else class="flex pl-1 flex-1 text-#002cfd line-height-4 flex-items-center">
                <div class="flex">
                  {{ text.className }}-{{ text.courseName }}
                </div>
              </div>
            </div>
          </a-popover>

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
                    : (emptyLessonStatusText(text)
                        ? (text.conflict ? 'bg-#ffe6e6 text-#a31616' : 'bg-#e6ffe6 text-#16a34a')
                        : 'st-empty-cell--idle'),
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
  transition: box-shadow 0.25s ease, transform 0.25s ease, opacity 0.2s ease;
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
