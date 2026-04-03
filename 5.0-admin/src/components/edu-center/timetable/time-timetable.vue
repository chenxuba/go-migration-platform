<script setup>
import { LeftOutlined, RightOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import CreateSchedulePopover from './create-schedule-popover.vue'
import ScheduleBatchEditModal from './schedule-batch-edit-modal.vue'
import { listTeachingSchedulesApi } from '@/api/edu-center/teaching-schedule'
import emitter, { EVENTS } from '@/utils/eventBus'

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
const scheduleLoading = ref(false)
const scheduleRows = ref([])
const scheduleEditOpen = ref(false)
const currentSchedule = ref(null)
const headerScrollRef = ref(null)
const boardScrollRef = ref(null)
let syncingScroll = false

const weekdayLabels = ['周一', '周二', '周三', '周四', '周五', '周六', '周日']
const timelineStart = 8 * 60
const timelineEnd = 22 * 60
const timelineTopPadding = 18
const hourRowHeight = 96
const timelineBottomPadding = 28
const scheduleCardMinWidth = 170
const scheduleCardGap = 5
/** 列宽 = 左 6px + 卡片 + 右 6px；与 eventStyle 里 +6 对齐，避免只左边留白、右边贴线 */
const scheduleColumnHorizontalInset = 6
const baseDateColumnWidth
  = scheduleCardMinWidth + scheduleColumnHorizontalInset * 2
const overlapExtraWidth = scheduleCardMinWidth + scheduleCardGap

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
  loadSchedules()
  emitter.on(EVENTS.REFRESH_DATA, loadSchedules)
})

onUnmounted(() => {
  if (nowTimer) {
    clearInterval(nowTimer)
    nowTimer = null
  }
  emitter.off(EVENTS.REFRESH_DATA, loadSchedules)
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
  currentDate.value
    = currentTime.value === 'day'
      ? currentDate.value.subtract(1, 'day')
      : currentDate.value.subtract(1, 'week')
}

function handleNext() {
  currentDate.value
    = currentTime.value === 'day'
      ? currentDate.value.add(1, 'day')
      : currentDate.value.add(1, 'week')
}

function handleGoThisWeek() {
  currentDate.value = dayjs()
}

const isViewingTodayOrThisWeek = computed(() => {
  if (currentTime.value === 'day')
    return currentDate.value.isSame(now.value, 'day')
  return getWeekStart(currentDate.value).isSame(getWeekStart(now.value), 'day')
})

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
const currentTimeMinutes = computed(
  () => now.value.hour() * 60 + now.value.minute(),
)
const currentTimeLabel = computed(() => now.value.format('HH:mm'))
const showCurrentTimeLine = computed(() => {
  if (
    currentTimeMinutes.value < timelineStart
    || currentTimeMinutes.value > timelineEnd
  ) {
    return false
  }
  return displayDates.value.some(
    date => date.format('YYYY-MM-DD') === todayKey.value,
  )
})

const queryDateRange = computed(() => {
  const dates = displayDates.value
  return {
    startDate: dates[0]?.format('YYYY-MM-DD') || dayjs().format('YYYY-MM-DD'),
    endDate:
      dates[dates.length - 1]?.format('YYYY-MM-DD')
      || dayjs().format('YYYY-MM-DD'),
  }
})

async function loadSchedules() {
  scheduleLoading.value = true
  try {
    const res = await listTeachingSchedulesApi({
      startDate: queryDateRange.value.startDate,
      endDate: queryDateRange.value.endDate,
      classType: 2,
    })
    if (res.code === 200) {
      scheduleRows.value = Array.isArray(res.result) ? res.result : []
      return
    }
    scheduleRows.value = []
  }
  catch (error) {
    console.error('load schedules failed', error)
    scheduleRows.value = []
  }
  finally {
    scheduleLoading.value = false
  }
}

watch(
  queryDateRange,
  () => {
    loadSchedules()
  },
  { deep: true },
)

const mockSchedules = computed(() =>
  scheduleRows.value.map(item => ({
    id: item.id,
    batchNo: item.batchNo,
    batchSize: item.batchSize || 1,
    classType: item.classType,
    dateKey: dayjs(item.startAt).format('YYYY-MM-DD'),
    startAt: dayjs(item.startAt),
    endAt: dayjs(item.endAt),
    title: item.teachingClassName || item.studentName || '1对1日程',
    course: item.lessonName || '-',
    teacher: item.teacherName || '-',
    classroom: item.classroomName || '-',
    studentText: item.studentName ? `学员：${item.studentName}` : '-',
    status: 'unsigned',
    conflict: false,
    hasTrial: false,
    raw: item,
  })),
)

const headerSummaries = computed(() =>
  displayDates.value.map((date) => {
    const key = date.format('YYYY-MM-DD')
    const count = mockSchedules.value.filter(
      item => item.dateKey === key,
    ).length
    return {
      key,
      date,
      count,
    }
  }),
)

const unsignedCount = computed(
  () => mockSchedules.value.filter(item => item.status === 'unsigned').length,
)

const hourMarks = computed(() =>
  Array.from(
    { length: timelineEnd / 60 - timelineStart / 60 + 1 },
    (_, index) => timelineStart + index * 60,
  ),
)

const hoverSlots = computed(() =>
  hourMarks.value.slice(0, -1).map((startMinutes, index) => ({
    key: `slot-${startMinutes}`,
    startMinutes,
    endMinutes: hourMarks.value[index + 1],
  })),
)

const timelineHeight = computed(
  () =>
    timelineTopPadding
    + (hourMarks.value.length - 1) * hourRowHeight
    + timelineBottomPadding,
)

function minuteOffset(minutes) {
  return timelineTopPadding + ((minutes - timelineStart) / 60) * hourRowHeight
}

function normalizeScheduleItem(item) {
  return {
    ...item,
    startMinutes: item.startAt.hour() * 60 + item.startAt.minute(),
    endMinutes: item.endAt.hour() * 60 + item.endAt.minute(),
  }
}

function computeOverlapPeak(items = []) {
  const columns = []
  let peak = 1
  items.forEach((item) => {
    let columnIndex = columns.findIndex(
      endValue => endValue <= item.startMinutes,
    )
    if (columnIndex === -1) {
      columnIndex = columns.length
      columns.push(item.endMinutes)
    }
    else {
      columns[columnIndex] = item.endMinutes
    }
    peak = Math.max(peak, columns.length)
  })
  return peak
}

const dateColumnWidths = computed(() => {
  const map = new Map()
  headerSummaries.value.forEach((item) => {
    const list = mockSchedules.value
      .filter(schedule => schedule.dateKey === item.key)
      .map(normalizeScheduleItem)
      .sort((a, b) => a.startAt.valueOf() - b.startAt.valueOf())
    const peak = computeOverlapPeak(list)
    map.set(
      item.key,
      baseDateColumnWidth + Math.max(0, peak - 1) * overlapExtraWidth,
    )
  })
  return map
})

const gridTemplateStyle = computed(() => {
  const isDaySingleColumn
    = currentTime.value === 'day' && headerSummaries.value.length === 1
  const dateCols = headerSummaries.value.map((item) => {
    const w = dateColumnWidths.value.get(item.key) || baseDateColumnWidth
    // 日视图仅一列：用 1fr 撑满剩余宽度，避免右侧大片留白
    return isDaySingleColumn ? `minmax(${w}px, 1fr)` : `${w}px`
  })
  return {
    gridTemplateColumns: `84px ${dateCols.join(' ')}`,
    width: isDaySingleColumn ? '100%' : 'max-content',
    minWidth: '100%',
  }
})

function buildClusterLayouts(clusterItems = []) {
  const columns = []
  let peakColumns = 0

  const assigned = clusterItems.map((item) => {
    let columnIndex = columns.findIndex(
      endValue => endValue <= item.startMinutes,
    )
    if (columnIndex === -1) {
      columnIndex = columns.length
      columns.push(item.endMinutes)
    }
    else {
      columns[columnIndex] = item.endMinutes
    }
    peakColumns = Math.max(peakColumns, columns.length)
    return {
      ...item,
      columnIndex,
      timeText: `${item.startAt.format('HH:mm')} - ${item.endAt.format(
        'HH:mm',
      )}`,
    }
  })

  return assigned.map(item => ({
    ...item,
    displayColumnIndex: item.columnIndex,
    displayColumnCount: peakColumns,
  }))
}

function buildDayLayouts(items = []) {
  const sorted = [...items]
    .map(normalizeScheduleItem)
    .sort((a, b) => a.startAt.valueOf() - b.startAt.valueOf())

  const clusters = []
  let currentCluster = []
  let currentEnd = -1

  sorted.forEach((item) => {
    if (currentCluster.length === 0) {
      currentCluster = [item]
      currentEnd = item.endMinutes
      return
    }

    if (item.startMinutes < currentEnd) {
      currentCluster.push(item)
      currentEnd = Math.max(currentEnd, item.endMinutes)
      return
    }

    clusters.push(currentCluster)
    currentCluster = [item]
    currentEnd = item.endMinutes
  })

  if (currentCluster.length)
    clusters.push(currentCluster)

  return clusters.flatMap(cluster => buildClusterLayouts(cluster))
}

const layoutsByDate = computed(() => {
  const map = new Map()
  headerSummaries.value.forEach((item) => {
    const list = mockSchedules.value.filter(
      schedule => schedule.dateKey === item.key,
    )
    map.set(item.key, buildDayLayouts(list))
  })
  return map
})

function eventStyle(item) {
  const leftOffset
    = (item.displayColumnIndex || 0) * (scheduleCardMinWidth + scheduleCardGap)
      + scheduleColumnHorizontalInset
  return {
    top: `${minuteOffset(item.startMinutes)}px`,
    height: `${Math.max(
      82,
      ((item.endMinutes - item.startMinutes) / 60) * hourRowHeight,
    )}px`,
    left: `${leftOffset}px`,
    width: `${scheduleCardMinWidth}px`,
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

function openScheduleEdit(item) {
  currentSchedule.value = item?.raw || null
  scheduleEditOpen.value = true
}

function handleScheduleUpdated() {
  loadSchedules()
}

function isActiveColumn(dateKey) {
  return dateKey === todayKey.value
}

function isMutedTimeLabel(mark) {
  if (!showCurrentTimeLine.value)
    return false
  return Math.abs(mark - currentTimeMinutes.value) <= 20
}

function isFutureDateKey(dateKey) {
  const current = dayjs(dateKey)
  const today = now.value.startOf('day')
  if (current.isAfter(today, 'day'))
    return true
  if (current.isBefore(today, 'day'))
    return false
  return true
}

function isFutureSlot(dateKey, startMinutes, endMinutes) {
  if (!isFutureDateKey(dateKey))
    return false
  const current = dayjs(dateKey)
  const today = now.value.startOf('day')
  if (current.isAfter(today, 'day'))
    return true
  return endMinutes > currentTimeMinutes.value
}

function hasEventInSlot(dateKey, startMinutes, endMinutes) {
  const list = layoutsByDate.value.get(dateKey) || []
  return list.some(
    event =>
      !(event.endMinutes <= startMinutes || event.startMinutes >= endMinutes),
  )
}

function createSlotStyle(startMinutes, endMinutes) {
  return {
    top: `${minuteOffset(startMinutes)}px`,
    height: `${((endMinutes - startMinutes) / 60) * hourRowHeight}px`,
  }
}

function syncScroll(source, target) {
  if (!source || !target || syncingScroll)
    return
  syncingScroll = true
  target.scrollLeft = source.scrollLeft
  requestAnimationFrame(() => {
    syncingScroll = false
  })
}

function handleHeaderScroll(event) {
  syncScroll(event.target, boardScrollRef.value)
}

function handleBoardScroll(event) {
  syncScroll(event.target, headerScrollRef.value)
}
</script>

<template>
  <div>
    <div
      class="filter-wrap bg-white pl-3 pr-3 rounded-4 rounded-lt-0 rounded-rt-0"
    >
      <all-filter
        :display-array="displayArray"
        :is-show-search-stu-phonefilter="true"
      />
    </div>

    <div class="time-page mt2">
      <div class="toolbar-card">
        <div class="toolbar-main">
          <div class="toolbar-group">
            <a-radio-group
              v-model:value="currentTime"
              button-style="solid"
              size="small"
            >
              <a-radio-button
                v-for="opt in timeOptions"
                :key="opt.key"
                :value="opt.key"
              >
                {{ opt.label }}
              </a-radio-button>
            </a-radio-group>
          </div>

          <div class="toolbar-date time-selector ml3 text-#0061ff font-800 text-5 flex-center">
            <a-popover trigger="hover">
              <template #content>
                {{ currentTime === 'day' ? '前一天' : '上一周' }}
              </template>
              <span
                class="cursor-pointer text-3 text-#888 flex w6 h6 bg-#eee rounded-10 flex-center font-500 hover-text-#06f hover-bg-#e6f0ff"
                @click="handlePrev"
              >
                <LeftOutlined />
              </span>
            </a-popover>
            <span class="mx-2">
              <div class="relative cursor-pointer">
                {{ formatDateRange(currentDate) }}
                <a-date-picker
                  v-if="currentTime === 'day'"
                  v-model:value="currentDate"
                  class="absolute left-0 top-0 right-0 bottom-0 z-10 opacity-0"
                  :allow-clear="false"
                  :bordered="false"
                  :format="formatDateRange"
                  style="cursor: pointer"
                />
                <a-date-picker
                  v-else
                  v-model:value="currentDate"
                  class="absolute left-0 top-0 right-0 bottom-0 z-10 opacity-0"
                  picker="week"
                  :allow-clear="false"
                  :bordered="false"
                  :format="formatDateRange"
                  style="cursor: pointer"
                />
              </div>
            </span>
            <a-popover trigger="hover">
              <template #content>
                {{ currentTime === 'day' ? '后一天' : '下一周' }}
              </template>
              <span
                class="cursor-pointer text-3 text-#888 flex w6 h6 bg-#eee rounded-10 flex-center font-500 hover-text-#06f hover-bg-#e6f0ff"
                @click="handleNext"
              >
                <RightOutlined />
              </span>
            </a-popover>
            <a-popover trigger="hover">
              <template #content>
                {{ currentTime === 'day' ? '回到今天' : '回到本周' }}
              </template>
              <a-button
                type="default"
                size="small"
                class="toolbar-today-week-btn ml2"
                :disabled="isViewingTodayOrThisWeek"
                @click="handleGoThisWeek"
              >
                {{ currentTime === 'day' ? '今天' : '本周' }}
              </a-button>
            </a-popover>
          </div>

          <a-space>
            <CreateSchedulePopover />
            <a-button>导出课表</a-button>
          </a-space>
        </div>
      </div>

      <div class="schedule-card">
        <a-spin
          :spinning="scheduleLoading"
          :delay="120"
          size="small"
          class="schedule-area-spin"
        >
        <div class="schedule-sticky-shell">
          <div class="schedule-summary">
            <div class="schedule-summary__left">
              <span class="summary-accent" />
              <span>共 {{ mockSchedules.length }} 个日程（未点名
                {{ unsignedCount }} 个日程）</span>
            </div>
            <div class="schedule-summary__right">
              <span
                v-for="item in scheduleLegend"
                :key="item.key"
                class="legend-item"
              >
                <span
                  v-if="item.type === 'bar'"
                  class="legend-item__bar"
                  :style="{ background: item.color }"
                />
                <span
                  v-else-if="item.type === 'icon'"
                  class="legend-item__icon legend-item__icon--trial"
                />
                <span
                  v-else
                  class="legend-item__icon legend-item__icon--danger"
                />
                {{ item.label }}
              </span>
            </div>
          </div>

          <div
            ref="headerScrollRef"
            class="schedule-header-scroll"
            @scroll="handleHeaderScroll"
          >
            <div class="schedule-header-grid" :style="gridTemplateStyle">
              <div class="schedule-time-header" />

              <div
                v-for="item in headerSummaries"
                :key="item.key"
                class="schedule-column-header"
                :class="{
                  'schedule-column-header--active': isActiveColumn(item.key),
                }"
              >
                <div class="schedule-column-header__title">
                  {{
                    currentTime === "day"
                      ? "当日"
                      : weekdayLabels[
                        item.date.day() === 0 ? 6 : item.date.day() - 1
                      ]
                  }}
                  <span class="schedule-column-header__date">（{{ item.date.format("M-D") }}）</span>
                </div>
                <div class="schedule-column-header__count">
                  {{ item.count }}个
                </div>
              </div>
            </div>
          </div>
        </div>

        <div
          ref="boardScrollRef"
          class="schedule-board"
          @scroll="handleBoardScroll"
        >
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
                <span class="schedule-time-axis__text">{{
                  formatClock(mark)
                }}</span>
              </div>
              <div
                v-if="showCurrentTimeLine"
                class="schedule-now-axis"
                :style="{ top: `${minuteOffset(currentTimeMinutes)}px` }"
              >
                <span class="schedule-now-axis__text">{{
                  currentTimeLabel
                }}</span>
                <span class="schedule-now-axis__dot" />
              </div>
            </div>

            <div
              v-for="item in headerSummaries"
              :key="`${item.key}-body`"
              class="schedule-column"
              :class="{ 'schedule-column--active': isActiveColumn(item.key) }"
            >
              <div
                class="schedule-column__body"
                :style="{ height: `${timelineHeight}px` }"
              >
                <div
                  v-for="mark in hourMarks"
                  :key="`${item.key}-${mark}`"
                  class="schedule-column__line"
                  :style="{ top: `${minuteOffset(mark)}px` }"
                />
                <template
                  v-for="slot in hoverSlots"
                  :key="`${item.key}-${slot.key}`"
                >
                  <div
                    v-if="
                      isFutureSlot(
                        item.key,
                        slot.startMinutes,
                        slot.endMinutes,
                      )
                        && !hasEventInSlot(
                          item.key,
                          slot.startMinutes,
                          slot.endMinutes,
                        )
                    "
                    class="schedule-create-slot"
                    :style="createSlotStyle(slot.startMinutes, slot.endMinutes)"
                  >
                    <CreateSchedulePopover trigger="click">
                      <button
                        type="button"
                        class="schedule-create-slot__trigger"
                      >
                        点击创建排课日程
                      </button>
                    </CreateSchedulePopover>
                  </div>
                </template>
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
                  @click="openScheduleEdit(event)"
                >
                  <div class="schedule-event__top">
                    <div class="schedule-event__time">
                      {{ event.timeText }}
                    </div>
                    <div class="schedule-event__badges">
                      <span
                        v-if="event.classType === 2"
                        class="schedule-event__badge schedule-event__badge--one-to-one"
                      >
                        1v1
                      </span>
                      <span v-if="event.hasTrial" class="schedule-event__badge">
                        试听
                      </span>
                    </div>
                  </div>
                  <div class="schedule-event__body">
                    <div class="schedule-event__title">
                      {{ event.title }}
                    </div>
                    <div class="schedule-event__meta schedule-event__meta__course">
                      {{ event.course }}
                    </div>
                    <div
                      class="schedule-event__meta schedule-event__meta--muted"
                    >
                      {{ event.teacher }}
                      <template
                        v-if="event.classroom && event.classroom !== '-'"
                      >
                        · {{ event.classroom }}
                      </template>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
        </a-spin>
      </div>
    </div>
    <ScheduleBatchEditModal
      v-model:open="scheduleEditOpen"
      :schedule="currentSchedule"
      @updated="handleScheduleUpdated"
    />
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
  justify-content: center;
  flex: 1;
  min-width: 0;
}

.schedule-card {
  overflow: visible;
}

/* 课表区域轻量 loading（替代 v-loading 黑色半透明蒙层） */
.schedule-area-spin {
  display: block;
  width: 100%;

  :deep(.ant-spin-nested-loading) {
    width: 100%;
  }

  :deep(.ant-spin-blur) {
    opacity: 0.72;
    filter: none;
    -webkit-filter: none;
  }

  :deep(.ant-spin) {
    max-height: none;
  }

  :deep(.ant-spin-dot) {
    font-size: 14px;
  }
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
  overflow-x: auto;
  overflow-y: hidden;
  background: #fff;
}

.schedule-header-grid {
  display: grid;
  width: max-content;
  min-width: 100%;
}

.schedule-board {
  overflow-x: auto;
  overflow-y: visible;
  background: #fff;
}

/* 表头与正文横向滚动条样式一致（两处 scrollLeft 会同步） */
.schedule-header-scroll,
.schedule-board {
  scrollbar-width: thin;
  scrollbar-color: #b8c9de #eef2f8;

  &::-webkit-scrollbar {
    height: 8px;
  }

  &::-webkit-scrollbar-track {
    margin: 0 4px;
    border-radius: 999px;
    background: #eef2f8;
  }

  &::-webkit-scrollbar-thumb {
    border-radius: 999px;
    background: linear-gradient(180deg, #c5d4ea 0%, #a8bad4 100%);
    box-shadow: inset 0 1px 0 rgb(255 255 255 / 45%);

    &:hover {
      background: linear-gradient(180deg, #a8bad4 0%, #8fa3bd 100%);
    }
  }
}

.schedule-grid {
  display: grid;
  width: max-content;
  min-width: 100%;
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
  position: sticky;
  left: 0;
  z-index: 20;
  background: #fff;
  box-shadow: 4px 0 14px -6px rgb(15 23 42 / 12%);
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
  position: sticky;
  left: 0;
  z-index: 20;
  border-right: 1px solid #dde5f0;
  background: #fff;
  box-shadow: 4px 0 14px -6px rgb(15 23 42 / 12%);
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
  transform: translateY(-50%);
}

.schedule-time-axis__label--muted .schedule-time-axis__text {
  color: rgb(31 41 55 / 22%);
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
  top: -7px;
  left: -3px;
  right: -4px;
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

.schedule-create-slot {
  position: absolute;
  left: 0;
  right: 0;
  z-index: 1;
}

.schedule-create-slot__trigger {
  width: calc(100% - 12px);
  height: calc(100% - 12px);
  margin: 6px;
  border: 1px dashed transparent;
  border-radius: 8px;
  background: transparent;
  color: transparent;
  font: inherit;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: color 0.18s ease, border-color 0.18s ease,
    background-color 0.18s ease;
}

.schedule-create-slot__trigger:hover {
  border-color: #cfe0ff;
  background: rgb(24 119 255 / 4%);
  color: #1677ff;
}

.schedule-event {
  position: absolute;
  z-index: 2;
  display: flex;
  flex-direction: column;
  border: none;
  border-radius: 4px;
  background: #ffffff;
  overflow: hidden;
  box-shadow: 0 6px 16px rgb(22 119 255 / 10%);
}

.schedule-event--unsigned {
  background: #ffffff;
}

.schedule-event--signed {
  background: #f5f7fa;
}

.schedule-event--conflict {
  box-shadow: 0 0 0 2px rgb(255 77 79 / 28%), 0 16px 32px rgb(255 77 79 / 16%);
}

.schedule-event__top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  min-height: 24px;
  padding: 3px 4px 3px 10px;
  background: #1677ff;
}

.schedule-event--signed .schedule-event__top {
  background: #98a2b3;
}

.schedule-event__time {
  flex: 1;
  min-width: 0;
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  line-height: 1.2;
  white-space: nowrap;
  letter-spacing: 0.01em;
}

.schedule-event__badges {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.schedule-event__badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 16px;
  padding: 0 7px;
  border-radius: 10px;
  background: rgb(9 61 149 / 24%);
  color: #fff;
  font-size: 10px;
  font-weight: 700;
  line-height: 1;
  border: none;
}

.schedule-event__badge--one-to-one {
  background: rgb(9 61 149 / 24%);
  color: #fff;
}

.schedule-event__body {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 4px 0 0 10px;
}

.schedule-event__title {
  color: #0f172a;
  font-size: 13px;
  font-weight: 700;
  line-height: 1;
  letter-spacing: 0.01em;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  overflow: hidden;
  -webkit-line-clamp: 2;
}

.schedule-event__meta {
  color: #64748b;
  font-size: 12px;
  line-height: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.schedule-event__meta__course{
  margin: 4px 0;
}

.schedule-event__meta--muted {
  color: #334155;
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
}

@media (max-width: 768px) {
  .schedule-summary {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
