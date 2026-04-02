<script setup>
import { LeftOutlined, RightOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import CreateSchedulePopover from './create-schedule-popover.vue'

const displayArray = ref([
  'intentionCourse',
  'reference',
  'department',
  'channelCategory',
  'channelStatus',
  'channelType',
  'subject',
])

const timeOptions = [
  { key: 'day', label: '日' },
  { key: 'week', label: '周' },
]

const currentTime = ref('week')
const currentDate = ref(dayjs())
const now = ref(dayjs())

const weekdayLabels = ['周一', '周二', '周三', '周四', '周五', '周六', '周日']
const timelineStart = 8 * 60
const timelineEnd = 21 * 60
const hourRowHeight = 88
const timelineBottomPadding = 28

const scheduleLegend = [
  {
    key: 'unsigned',
    label: '未点名（教师/课程）',
    type: 'bar',
    color: 'linear-gradient(90deg, #39b8ff 0%, #6c5cff 50%, #74d87f 100%)',
  },
  {
    key: 'signed',
    label: '已点名',
    type: 'bar',
    color: '#b7bec8',
  },
  {
    key: 'trial',
    label: '含试听学员',
    type: 'icon',
  },
  {
    key: 'conflict',
    label: '日程冲突',
    type: 'icon-danger',
  },
]

function getWeekStart(value = dayjs()) {
  const current = dayjs(value)
  const diff = current.day() === 0 ? -6 : 1 - current.day()
  return current.add(diff, 'day').startOf('day')
}

let nowTimer = null

onMounted(() => {
  nowTimer = setInterval(() => {
    now.value = dayjs()
  }, 30 * 1000)
})

onUnmounted(() => {
  if (nowTimer) {
    clearInterval(nowTimer)
    nowTimer = null
  }
})

watch(currentTime, () => {
  currentDate.value = dayjs()
})

function formatDateRange(value) {
  if (!value)
    return ''

  if (currentTime.value === 'day')
    return value.format('YYYY年MM月DD日')

  const start = getWeekStart(value)
  const end = start.add(6, 'day')

  if (start.year() === end.year() && start.month() === end.month())
    return `${start.format('YYYY年MM月DD日')} ~ ${end.format('DD日')}`
  if (start.year() === end.year())
    return `${start.format('YYYY年MM月DD日')} ~ ${end.format('MM月DD日')}`
  return `${start.format('YYYY年MM月DD日')} ~ ${end.format('YYYY年MM月DD日')}`
}

function handlePrev() {
  currentDate.value = currentTime.value === 'day'
    ? currentDate.value.subtract(1, 'day')
    : currentDate.value.subtract(1, 'week')
}

function handleNext() {
  currentDate.value = currentTime.value === 'day'
    ? currentDate.value.add(1, 'day')
    : currentDate.value.add(1, 'week')
}

function formatClock(minutes) {
  const hour = String(Math.floor(minutes / 60)).padStart(2, '0')
  const minute = String(minutes % 60).padStart(2, '0')
  return `${hour}:${minute}`
}

const displayDates = computed(() => {
  if (currentTime.value === 'day')
    return [dayjs(currentDate.value).startOf('day')]
  const start = getWeekStart(currentDate.value)
  return Array.from({ length: 7 }, (_, index) => start.add(index, 'day'))
})

const todayKey = computed(() => now.value.format('YYYY-MM-DD'))
const currentTimeMinutes = computed(() => now.value.hour() * 60 + now.value.minute())
const currentTimeLabel = computed(() => now.value.format('HH:mm'))
const showCurrentTimeLine = computed(() => {
  if (currentTimeMinutes.value < timelineStart || currentTimeMinutes.value > timelineEnd)
    return false
  return displayDates.value.some(date => date.format('YYYY-MM-DD') === todayKey.value)
})

const mockSchedules = computed(() => [])

const headerSummaries = computed(() =>
  displayDates.value.map((date) => {
    const key = date.format('YYYY-MM-DD')
    const count = mockSchedules.value.filter(item => item.dateKey === key).length
    return {
      key,
      date,
      count,
    }
  }),
)

const gridTemplateStyle = computed(() => ({
  gridTemplateColumns: `84px repeat(${headerSummaries.value.length}, minmax(0, 1fr))`,
}))

const unsignedCount = computed(() => mockSchedules.value.filter(item => item.status === 'unsigned').length)

const hourMarks = computed(() =>
  Array.from({ length: timelineEnd / 60 - timelineStart / 60 + 1 }, (_, index) => timelineStart + index * 60),
)

const timelineHeight = computed(() => (hourMarks.value.length - 1) * hourRowHeight + timelineBottomPadding)

function minuteOffset(minutes) {
  return ((minutes - timelineStart) / 60) * hourRowHeight
}

function buildDayLayouts(items = []) {
  const sorted = [...items].sort((a, b) => a.startAt.valueOf() - b.startAt.valueOf())
  const columns = []

  return sorted.map((item) => {
    const startMinutes = item.startAt.hour() * 60 + item.startAt.minute()
    const endMinutes = item.endAt.hour() * 60 + item.endAt.minute()

    let columnIndex = columns.findIndex(endValue => endValue <= startMinutes)
    if (columnIndex === -1) {
      columnIndex = columns.length
      columns.push(endMinutes)
    }
    else {
      columns[columnIndex] = endMinutes
    }

    return {
      ...item,
      startMinutes,
      endMinutes,
      columnIndex,
      columnCount: columns.length,
    }
  }).map((item) => {
    const overlapItems = sorted.filter((other) => {
      const otherStart = other.startAt.hour() * 60 + other.startAt.minute()
      const otherEnd = other.endAt.hour() * 60 + other.endAt.minute()
      return !(otherEnd <= item.startMinutes || otherStart >= item.endMinutes)
    })
    const overlapCount = Math.max(...overlapItems.map((other) => {
      const target = sorted.findIndex(x => x.id === other.id)
      return target >= 0 ? 1 : 1
    }), 1)

    return {
      ...item,
      columnCount: Math.max(item.columnCount, overlapCount),
    }
  })
}

const layoutsByDate = computed(() => {
  const map = new Map()
  headerSummaries.value.forEach((item) => {
    const list = mockSchedules.value.filter(schedule => schedule.dateKey === item.key)
    map.set(item.key, buildDayLayouts(list))
  })
  return map
})

function eventStyle(item) {
  const widthPercent = 100 / Math.max(item.columnCount || 1, 1)
  const leftPercent = widthPercent * item.columnIndex
  return {
    top: `${minuteOffset(item.startMinutes)}px`,
    height: `${Math.max(64, ((item.endMinutes - item.startMinutes) / 60) * hourRowHeight)}px`,
    left: `calc(${leftPercent}% + 6px)`,
    width: `calc(${widthPercent}% - 12px)`,
  }
}

function eventClass(item) {
  return {
    'schedule-event': true,
    'schedule-event--unsigned': item.status === 'unsigned',
    'schedule-event--signed': item.status === 'signed',
    'schedule-event--conflict': item.conflict,
  }
}

function isActiveColumn(dateKey) {
  return dateKey === todayKey.value
}

function isMutedTimeLabel(mark) {
  if (!showCurrentTimeLine.value)
    return false
  return Math.abs(mark - currentTimeMinutes.value) <= 20
}
</script>

<template>
<div>
  <div class="filter-wrap bg-white pl-3 pr-3 rounded-4 rounded-lt-0 rounded-rt-0">
    <all-filter :display-array="displayArray" :is-show-search-stu-phonefilter="true" />
  </div>

  <div class="time-page mt2">
    <div class="toolbar-card">
      <div class="toolbar-main">
        <div class="toolbar-group">
          <a-radio-group v-model:value="currentTime" button-style="solid" size="small">
            <a-radio-button v-for="opt in timeOptions" :key="opt.key" :value="opt.key">
              {{ opt.label }}
            </a-radio-button>
          </a-radio-group>
        </div>

        <div class="toolbar-date">
          <span class="nav-btn" @click="handlePrev">
            <LeftOutlined />
          </span>
          <div class="toolbar-date__label">
            {{ formatDateRange(currentDate) }}
            <a-date-picker
              v-if="currentTime === 'day'"
              v-model:value="currentDate"
              class="date-picker-mask"
              :allow-clear="false"
              :bordered="false"
              :format="formatDateRange"
            />
            <a-date-picker
              v-else
              v-model:value="currentDate"
              class="date-picker-mask"
              picker="week"
              :allow-clear="false"
              :bordered="false"
              :format="formatDateRange"
            />
          </div>
          <span class="nav-btn" @click="handleNext">
            <RightOutlined />
          </span>
        </div>

        <a-space>
          <create-schedule-popover />
          <a-button>导出课表</a-button>
        </a-space>
      </div>
    </div>

    <div class="schedule-card">
      <div class="schedule-sticky-shell">
        <div class="schedule-summary">
          <div class="schedule-summary__left">
            <span class="summary-accent" />
            <span>共 {{ mockSchedules.length }} 个日程（未点名 {{ unsignedCount }} 个日程）</span>
          </div>
          <div class="schedule-summary__right">
            <span v-for="item in scheduleLegend" :key="item.key" class="legend-item">
              <span
                v-if="item.type === 'bar'"
                class="legend-item__bar"
                :style="{ background: item.color }"
              />
              <span v-else-if="item.type === 'icon'" class="legend-item__icon legend-item__icon--trial" />
              <span v-else class="legend-item__icon legend-item__icon--danger" />
              {{ item.label }}
            </span>
          </div>
        </div>

        <div class="schedule-header-scroll">
          <div class="schedule-header-grid" :style="gridTemplateStyle">
            <div class="schedule-time-header" />

            <div
              v-for="item in headerSummaries"
              :key="item.key"
              class="schedule-column-header"
              :class="{ 'schedule-column-header--active': isActiveColumn(item.key) }"
            >
              <div class="schedule-column-header__title">
                {{ currentTime === 'day' ? '当日' : weekdayLabels[item.date.day() === 0 ? 6 : item.date.day() - 1] }}
                <span class="schedule-column-header__date">（{{ item.date.format('M-D') }}）</span>
              </div>
              <div class="schedule-column-header__count">
                {{ item.count }}个
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="schedule-board">
        <div class="schedule-grid" :style="gridTemplateStyle">
          <div class="schedule-time-axis">
            <div
              v-for="(mark, index) in hourMarks"
              :key="mark"
              class="schedule-time-axis__label"
              :class="{
                'schedule-time-axis__label--first': index === 0,
                'schedule-time-axis__label--muted': isMutedTimeLabel(mark),
              }"
              :style="{ top: `${minuteOffset(mark)}px` }"
            >
              <span class="schedule-time-axis__text">{{ formatClock(mark) }}</span>
            </div>
            <div
              v-if="showCurrentTimeLine"
              class="schedule-now-axis"
              :style="{ top: `${minuteOffset(currentTimeMinutes)}px` }"
            >
              <span class="schedule-now-axis__text">{{ currentTimeLabel }}</span>
              <span class="schedule-now-axis__dot" />
            </div>
          </div>

          <div
            v-for="item in headerSummaries"
            :key="`${item.key}-body`"
            class="schedule-column"
            :class="{ 'schedule-column--active': isActiveColumn(item.key) }"
          >
            <div class="schedule-column__body" :style="{ height: `${timelineHeight}px` }">
              <div
                v-for="mark in hourMarks"
                :key="`${item.key}-${mark}`"
                class="schedule-column__line"
                :style="{ top: `${minuteOffset(mark)}px` }"
              />
              <div
                v-if="showCurrentTimeLine"
                class="schedule-now-line"
                :style="{ top: `${minuteOffset(currentTimeMinutes)}px` }"
              />

              <div
                v-for="event in layoutsByDate.get(item.key) || []"
                :key="event.id"
                :class="eventClass(event)"
                :style="eventStyle(event)"
              >
                <div class="schedule-event__top">
                  <span class="schedule-event__time">
                    {{ event.startAt.format('HH:mm') }} - {{ event.endAt.format('HH:mm') }}
                  </span>
                  <span v-if="event.hasTrial" class="schedule-event__badge">
                    试听
                  </span>
                </div>
                <div class="schedule-event__title">
                  {{ event.title }}
                </div>
                <div class="schedule-event__meta">
                  {{ event.course }}
                </div>
                <div class="schedule-event__meta">
                  {{ event.teacher }} · {{ event.classroom }}
                </div>
                <div class="schedule-event__footer">
                  {{ event.studentText }}
                </div>
              </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped lang="less">
.time-page {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.toolbar-card,
.schedule-card {
  border: 1px solid #e5ebf3;
  border-radius: 16px;
  background: #fff;
  box-shadow: 0 10px 24px rgb(15 23 42 / 4%);
}

.toolbar-card {
  padding: 14px 18px;
}

.toolbar-main {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.toolbar-group {
  display: flex;
  align-items: center;
  gap: 12px;
}

.toolbar-date {
  display: flex;
  align-items: center;
  gap: 10px;
}

.toolbar-date__label {
  position: relative;
  min-width: 260px;
  padding: 8px 14px;
  border: 1px solid #dbe5f0;
  border-radius: 999px;
  color: #0f172a;
  font-size: 13px;
  font-weight: 600;
  text-align: center;
}

.date-picker-mask {
  position: absolute;
  inset: 0;
  opacity: 0;
  cursor: pointer;
}

.nav-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border-radius: 999px;
  background: #f1f5f9;
  color: #64748b;
  cursor: pointer;
  transition: all 0.2s ease;
}

.nav-btn:hover {
  background: #e8f3ff;
  color: #1677ff;
}

.schedule-card {
  overflow: visible;
}

.schedule-sticky-shell {
  position: sticky;
  top: 8px;
  z-index: 40;
  background: #fff;
  box-shadow: 0 10px 22px rgb(15 23 42 / 6%);
}

.schedule-summary {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 16px 10px;
  border-bottom: 1px solid #edf2f7;
  background: rgb(255 255 255 / 98%);
  backdrop-filter: blur(12px);
}

.schedule-summary__left,
.schedule-summary__right {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
}

.schedule-summary__left {
  color: #1f2937;
  font-size: 13px;
  font-weight: 600;
}

.summary-accent {
  width: 4px;
  height: 16px;
  border-radius: 999px;
  background: #1677ff;
}

.legend-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #4b5563;
  font-size: 12px;
}

.legend-item__bar {
  display: inline-block;
  width: 18px;
  height: 4px;
  border-radius: 999px;
}

.legend-item__icon {
  position: relative;
  display: inline-block;
  width: 12px;
  height: 12px;
  border-radius: 3px;
  background: #fff;
  border: 1px solid #cbd5e1;
}

.legend-item__icon--trial::after {
  position: absolute;
  left: 2px;
  top: 2px;
  width: 6px;
  height: 6px;
  background: #b5bfcf;
  border-radius: 1px;
  content: "";
}

.legend-item__icon--danger {
  border-color: #ff7875;
}

.legend-item__icon--danger::after {
  position: absolute;
  left: 1px;
  right: 1px;
  top: 50%;
  height: 2px;
  background: #ff4d4f;
  transform: translateY(-50%);
  content: "";
}

.schedule-header-scroll {
  background: #fff;
}

.schedule-header-grid {
  display: grid;
  width: 100%;
}

.schedule-board {
  overflow: visible;
  background: #fff;
}

.schedule-grid {
  display: grid;
  width: 100%;
  position: relative;
}

.schedule-time-header,
.schedule-column-header {
  height: 48px;
  border-right: 1px solid #dde5f0;
  border-bottom: 1px solid #dde5f0;
  background: #eff4fb;
}

.schedule-time-header {
  background: #fff;
}

.schedule-column-header {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: #374151;
  font-size: 14px;
  font-weight: 700;
}

.schedule-column-header--active {
  color: #1677ff;
  box-shadow: inset 0 3px 0 #1677ff;
}

.schedule-column-header__date {
  color: #6b7280;
  font-weight: 600;
}

.schedule-column-header__count {
  color: #6b7280;
  font-size: 13px;
  font-weight: 600;
}

.schedule-time-axis {
  position: relative;
  border-right: 1px solid #dde5f0;
  background: #fff;
  z-index: 5;
}

.schedule-time-axis__label {
  position: absolute;
  left: 0;
  right: 0;
  transform: translateY(-50%);
  color: #1f2937;
  font-size: 14px;
  text-align: center;
  z-index: 6;
}

.schedule-time-axis__label::before {
  position: absolute;
  left: 0;
  right: 0;
  top: 50%;
  border-top: 1px solid #dde5f0;
  content: "";
  z-index: 0;
}

.schedule-time-axis__label--first {
  transform: translateY(0);
}

.schedule-time-axis__label--muted .schedule-time-axis__text {
  opacity: 0.28;
}

.schedule-time-axis__text {
  position: relative;
  z-index: 1;
  display: inline-block;
  padding: 0 10px;
  background: #fff;
}

.schedule-column {
  position: relative;
  border-right: 1px solid #dde5f0;
  background: #fff;
}

.schedule-column--active {
  background: #f3f9ff;
}

.schedule-column__body {
  position: relative;
}

.schedule-column__line {
  position: absolute;
  left: 0;
  right: 0;
  border-top: 1px solid #dde5f0;
}

.schedule-now-axis {
  position: absolute;
  z-index: 8;
  pointer-events: none;
  left: 0;
  right: 0;
}

.schedule-now-axis__text {
  position: absolute;
  top: -10px;
  left: 0;
  right: 0;
  color: #ff4d4f;
  font-size: 12px;
  font-weight: 600;
  line-height: 1.2;
  text-align: center;
}

.schedule-now-axis__dot {
  position: absolute;
  top: -3px;
  left: 81px;
  width: 6px;
  height: 6px;
  border-radius: 999px;
  background: #ff4d4f;
}

.schedule-now-line {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  border-top: 1px solid #ffb3b3;
  z-index: 1;
  pointer-events: none;
}

.schedule-event {
  position: absolute;
  z-index: 2;
  padding: 4px 10px 4px;
  border-radius: 8px;
  border: 1px solid transparent;
  overflow: hidden;
  box-shadow: 0 8px 20px rgb(15 23 42 / 8%);
}

.schedule-event--unsigned {
  background: linear-gradient(180deg, #f5f8ff 0%, #ffffff 100%);
  border-color: #dbeafe;
}

.schedule-event--signed {
  background: linear-gradient(180deg, #f5f7fa 0%, #fff 100%);
  border-color: #d7dbe2;
}

.schedule-event--unsigned::before,
.schedule-event--signed::before {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 4px;
  content: "";
}

.schedule-event--unsigned::before {
  background: linear-gradient(180deg, #39b8ff 0%, #6c5cff 52%, #74d87f 100%);
}

.schedule-event--signed::before {
  background: #b7bec8;
}

.schedule-event--conflict {
  box-shadow: 0 0 0 1px rgb(255 77 79 / 35%), 0 8px 20px rgb(255 77 79 / 10%);
}

.schedule-event__top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.schedule-event__time {
  color: #3b82f6;
  font-size: 11px;
  font-weight: 700;
}

.schedule-event__badge {
  padding: 1px 6px;
  border-radius: 999px;
  background: #eef2f7;
  color: #64748b;
  font-size: 11px;
  font-weight: 700;
}

.schedule-event__title {
  color: #0f172a;
  font-size: 14px;
  font-weight: 700;
  line-height: 1.4;
}

.schedule-event__meta {
  color: #64748b;
  font-size: 12px;
  line-height: 1.4;
}

.schedule-event__footer {
  margin-top: 6px;
  color: #334155;
  font-size: 12px;
  font-weight: 600;
}

.time-selector {
  font-family: DINAlternate-Bold, DINAlternate;
}

.toolbar-card :deep(.ant-radio-button-wrapper) {
  padding: 0 16px;
}

@media (max-width: 1200px) {
  .toolbar-main {
    flex-wrap: wrap;
    justify-content: flex-start;
  }

  .toolbar-date__label {
    min-width: 220px;
  }
}

@media (max-width: 768px) {
  .schedule-summary {
    align-items: flex-start;
    flex-direction: column;
  }

  .toolbar-date__label {
    min-width: 180px;
  }
}
</style>
