<script setup lang="ts">
import dayjs from 'dayjs'
import { computed, nextTick, onMounted, onUnmounted, ref, shallowRef, watch } from 'vue'

interface UltraCourse {
  id: string
  dayIndex: number
  startRow: number
  endRow: number
  laneIndex: number
  laneCount: number
  title: string
  teacherName: string
  studentName: string
  classroomName: string
  fill: string
}

interface PaintedCourseRect {
  course: UltraCourse
  x: number
  y: number
  width: number
  height: number
}

const emit = defineEmits(['week-range-change'])

const TOTAL_DAYS = 7
const TOTAL_ROWS = 720
const ROW_MINUTES = 2
const TOTAL_COURSE_COUNT = 10000
const TIME_GUTTER_WIDTH = 86
const HEADER_HEIGHT = 54
const ROW_HEIGHT = 20
const DAY_COLUMN_WIDTH = 220
const VIEWPORT_MIN_HEIGHT = 560
const VIEWPORT_MAX_HEIGHT = 920
const CANVAS_SHELL_PADDING = 24
const COURSE_HORIZONTAL_GAP = 4
const COURSE_VERTICAL_GAP = 2
const TEXT_LEFT_PADDING = 8
const GRID_BACKGROUND = '#f6f8fc'
const BORDER_COLOR = '#d8dfeb'
const SUB_BORDER_COLOR = '#edf1f7'
const HEADER_BACKGROUND = '#ffffff'
const TIME_GUTTER_BACKGROUND = '#fbfcff'
const TEXT_PRIMARY = '#1f2937'
const TEXT_MUTED = '#6b7280'

const dayLabels = ['周一', '周二', '周三', '周四', '周五', '周六', '周日']
const teacherNames = [
  '张三', '李四', '王五', '赵六', '刘晨', '陈阳', '杨悦', '黄宁', '周婷', '吴轩',
  '徐浩', '孙然', '胡涵', '朱彤', '高辰', '林琳', '何菲', '郭宇', '马可', '罗欣',
]
const studentNames = [
  '安安', '乐乐', '可可', '宁宁', '然然', '晨晨', '小宇', '小涵', '依依', '果果',
  '豆豆', '浩浩', '佳佳', '萌萌', '圆圆', '苗苗', '清清', '朵朵', '元元', '菲菲',
]
const courseNames = [
  '数学', '语文', '英语', '物理', '化学', '生物', '历史', '地理', '美术', '书法',
  '钢琴', '围棋', '口才', '编程', '作文', '舞蹈', '素描', '阅读', '科学', '逻辑',
]
const classroomNames = ['A101', 'A201', 'A301', 'B102', 'B202', 'C103', 'C203', 'D305']
const coursePalette = [
  '#4e6dff',
  '#22c55e',
  '#f59e0b',
  '#ef4444',
  '#0ea5e9',
  '#8b5cf6',
  '#ec4899',
  '#14b8a6',
]

const shellRef = ref<HTMLElement | null>(null)
const viewportRef = ref<HTMLElement | null>(null)
const canvasRef = ref<HTMLCanvasElement | null>(null)
const viewportHeight = ref(VIEWPORT_MIN_HEIGHT)
const viewportWidth = ref(1280)
const scrollTop = ref(0)
const scrollLeft = ref(0)
const selectedCourseId = ref('')
const paintedCourseRects = shallowRef<PaintedCourseRect[]>([])
const resizeTick = ref(0)

let resizeObserver: ResizeObserver | null = null
let pendingFrame = 0

const weekStart = computed(() => {
  const now = dayjs()
  const offset = (now.day() + 6) % 7
  return now.startOf('day').subtract(offset, 'day')
})

const weekDates = computed(() => {
  return Array.from({ length: TOTAL_DAYS }, (_, index) => weekStart.value.add(index, 'day'))
})

const totalGridWidth = computed(() => TIME_GUTTER_WIDTH + TOTAL_DAYS * DAY_COLUMN_WIDTH)
const totalGridHeight = computed(() => HEADER_HEIGHT + TOTAL_ROWS * ROW_HEIGHT)
const totalGridCellCount = TOTAL_DAYS * TOTAL_ROWS

function createSeededRandom(seed: number) {
  let value = seed >>> 0
  return () => {
    value = (value * 1664525 + 1013904223) >>> 0
    return value / 0x100000000
  }
}

function slotTimeLabel(rowIndex: number) {
  const totalMinutes = rowIndex * ROW_MINUTES
  const hour = Math.floor(totalMinutes / 60)
  const minute = totalMinutes % 60
  return `${String(hour).padStart(2, '0')}:${String(minute).padStart(2, '0')}`
}

function courseTimeText(course: UltraCourse) {
  return `${slotTimeLabel(course.startRow)}-${slotTimeLabel(course.endRow)}`
}

function ellipsisText(ctx: CanvasRenderingContext2D, text: string, maxWidth: number) {
  if (!text)
    return ''
  if (ctx.measureText(text).width <= maxWidth)
    return text
  let output = text
  while (output.length > 1 && ctx.measureText(`${output}…`).width > maxWidth)
    output = output.slice(0, -1)
  return `${output}…`
}

function buildCourseSeedData() {
  const random = createSeededRandom(20260413)
  const courses: UltraCourse[] = []
  for (let index = 0; index < TOTAL_COURSE_COUNT; index += 1) {
    const dayIndex = Math.floor(random() * TOTAL_DAYS)
    const durationRows = 10 + Math.floor(random() * 32)
    const maxStart = Math.max(0, TOTAL_ROWS - durationRows - 1)
    const startRow = Math.floor(random() * Math.max(1, maxStart + 1))
    const endRow = Math.min(TOTAL_ROWS, startRow + durationRows)
    const courseName = `${courseNames[index % courseNames.length]}${String((index % 30) + 1).padStart(2, '0')}`
    const teacherName = teacherNames[index % teacherNames.length]
    const studentName = studentNames[(index * 3) % studentNames.length]
    const classroomName = classroomNames[index % classroomNames.length]
    courses.push({
      id: `canvas-course-${index + 1}`,
      dayIndex,
      startRow,
      endRow,
      laneIndex: 0,
      laneCount: 1,
      title: courseName,
      teacherName,
      studentName,
      classroomName,
      fill: coursePalette[index % coursePalette.length],
    })
  }
  return courses
}

function assignCourseLanes(courses: UltraCourse[]) {
  const byDay = Array.from({ length: TOTAL_DAYS }, () => [] as UltraCourse[])
  courses.forEach((course) => {
    byDay[course.dayIndex].push(course)
  })

  byDay.forEach((items) => {
    items.sort((left, right) => {
      if (left.startRow !== right.startRow)
        return left.startRow - right.startRow
      return left.endRow - right.endRow
    })

    let active: UltraCourse[] = []
    let groupCourses: UltraCourse[] = []
    let groupMaxLane = 1
    let groupEnd = -1

    function flushGroup() {
      if (!groupCourses.length)
        return
      groupCourses.forEach((course) => {
        course.laneCount = groupMaxLane
      })
      groupCourses = []
      groupMaxLane = 1
      groupEnd = -1
    }

    items.forEach((course) => {
      active = active.filter(item => item.endRow > course.startRow)
      if (!active.length && groupCourses.length)
        flushGroup()

      const laneUsed = new Set(active.map(item => item.laneIndex))
      let laneIndex = 0
      while (laneUsed.has(laneIndex))
        laneIndex += 1
      course.laneIndex = laneIndex
      active.push(course)

      if (!groupCourses.length || course.startRow < groupEnd) {
        groupCourses.push(course)
        groupEnd = Math.max(groupEnd, course.endRow)
      }
      else {
        flushGroup()
        groupCourses = [course]
        groupEnd = course.endRow
      }

      const activeMaxLane = active.reduce((max, item) => Math.max(max, item.laneIndex + 1), 1)
      groupMaxLane = Math.max(groupMaxLane, activeMaxLane)
    })

    flushGroup()
  })

  return byDay.map(items => items.sort((left, right) => left.startRow - right.startRow))
}

const allCourses = shallowRef(buildCourseSeedData())
const coursesByDay = shallowRef(assignCourseLanes(allCourses.value))

const selectedCourse = computed(() => {
  return allCourses.value.find(course => course.id === selectedCourseId.value) || null
})

const summaryCards = computed(() => {
  return [
    { label: '总格子数', value: `${totalGridCellCount}` },
    { label: '课程数量', value: `${TOTAL_COURSE_COUNT}` },
    { label: '时段粒度', value: `${ROW_MINUTES} 分钟` },
    { label: '周区间', value: `${weekDates.value[0].format('MM-DD')} ~ ${weekDates.value[6].format('MM-DD')}` },
  ]
})

function updateViewportMetrics() {
  const shell = shellRef.value
  const viewport = viewportRef.value
  if (!(shell instanceof HTMLElement) || !(viewport instanceof HTMLElement) || typeof window === 'undefined')
    return
  const shellRect = shell.getBoundingClientRect()
  viewportWidth.value = viewport.clientWidth
  const nextHeight = Math.floor(window.innerHeight - shellRect.top - CANVAS_SHELL_PADDING)
  viewportHeight.value = Math.max(VIEWPORT_MIN_HEIGHT, Math.min(VIEWPORT_MAX_HEIGHT, nextHeight))
}

function syncCanvasSize() {
  const canvas = canvasRef.value
  const viewport = viewportRef.value
  if (!(canvas instanceof HTMLCanvasElement) || !(viewport instanceof HTMLElement))
    return null
  const ratio = typeof window !== 'undefined' ? Math.max(window.devicePixelRatio || 1, 1) : 1
  const cssWidth = viewport.clientWidth
  const cssHeight = viewport.clientHeight
  if (!cssWidth || !cssHeight)
    return null
  const pixelWidth = Math.floor(cssWidth * ratio)
  const pixelHeight = Math.floor(cssHeight * ratio)
  if (canvas.width !== pixelWidth || canvas.height !== pixelHeight) {
    canvas.width = pixelWidth
    canvas.height = pixelHeight
  }
  canvas.style.width = `${cssWidth}px`
  canvas.style.height = `${cssHeight}px`
  const ctx = canvas.getContext('2d')
  if (!ctx)
    return null
  ctx.setTransform(ratio, 0, 0, ratio, 0, 0)
  return ctx
}

function drawGrid(ctx: CanvasRenderingContext2D) {
  const canvas = canvasRef.value
  if (!(canvas instanceof HTMLCanvasElement))
    return

  const visibleWidth = canvas.clientWidth
  const visibleHeight = canvas.clientHeight
  const horizontalOffset = scrollLeft.value
  const verticalOffset = scrollTop.value
  const bodyTop = HEADER_HEIGHT

  ctx.clearRect(0, 0, visibleWidth, visibleHeight)
  ctx.fillStyle = GRID_BACKGROUND
  ctx.fillRect(0, 0, visibleWidth, visibleHeight)

  ctx.fillStyle = HEADER_BACKGROUND
  ctx.fillRect(0, 0, visibleWidth, HEADER_HEIGHT)
  ctx.fillStyle = TIME_GUTTER_BACKGROUND
  ctx.fillRect(0, 0, TIME_GUTTER_WIDTH, visibleHeight)

  const firstVisibleDay = Math.max(0, Math.floor(horizontalOffset / DAY_COLUMN_WIDTH))
  const lastVisibleDay = Math.min(
    TOTAL_DAYS - 1,
    Math.ceil((horizontalOffset + Math.max(0, visibleWidth - TIME_GUTTER_WIDTH)) / DAY_COLUMN_WIDTH),
  )
  const firstVisibleRow = Math.max(0, Math.floor(Math.max(0, verticalOffset - HEADER_HEIGHT) / ROW_HEIGHT))
  const lastVisibleRow = Math.min(
    TOTAL_ROWS,
    Math.ceil(Math.max(0, verticalOffset - HEADER_HEIGHT + visibleHeight) / ROW_HEIGHT) + 1,
  )

  ctx.strokeStyle = BORDER_COLOR
  ctx.lineWidth = 1
  ctx.beginPath()
  ctx.moveTo(TIME_GUTTER_WIDTH + 0.5, 0)
  ctx.lineTo(TIME_GUTTER_WIDTH + 0.5, visibleHeight)
  ctx.moveTo(0, HEADER_HEIGHT + 0.5)
  ctx.lineTo(visibleWidth, HEADER_HEIGHT + 0.5)
  ctx.stroke()

  for (let dayIndex = firstVisibleDay; dayIndex <= lastVisibleDay; dayIndex += 1) {
    const x = TIME_GUTTER_WIDTH + dayIndex * DAY_COLUMN_WIDTH - horizontalOffset
    ctx.fillStyle = HEADER_BACKGROUND
    ctx.fillRect(x, 0, DAY_COLUMN_WIDTH, HEADER_HEIGHT)
    ctx.strokeStyle = BORDER_COLOR
    ctx.beginPath()
    ctx.moveTo(x + 0.5, 0)
    ctx.lineTo(x + 0.5, visibleHeight)
    ctx.stroke()

    const weekDate = weekDates.value[dayIndex]
    ctx.fillStyle = TEXT_PRIMARY
    ctx.font = '600 15px sans-serif'
    ctx.fillText(dayLabels[dayIndex], x + 12, 22)
    ctx.fillStyle = TEXT_MUTED
    ctx.font = '12px sans-serif'
    ctx.fillText(weekDate.format('MM-DD'), x + 12, 40)
  }

  for (let rowIndex = firstVisibleRow; rowIndex <= lastVisibleRow; rowIndex += 1) {
    const y = HEADER_HEIGHT + rowIndex * ROW_HEIGHT - verticalOffset
    const isHourLine = rowIndex % 30 === 0
    ctx.strokeStyle = isHourLine ? BORDER_COLOR : SUB_BORDER_COLOR
    ctx.beginPath()
    ctx.moveTo(0, y + 0.5)
    ctx.lineTo(visibleWidth, y + 0.5)
    ctx.stroke()

    if (rowIndex < TOTAL_ROWS && (rowIndex % 30 === 0 || rowIndex === 0)) {
      ctx.fillStyle = TEXT_MUTED
      ctx.font = '12px sans-serif'
      ctx.fillText(slotTimeLabel(rowIndex), 10, y + 14)
    }
  }

  ctx.fillStyle = HEADER_BACKGROUND
  ctx.fillRect(0, 0, TIME_GUTTER_WIDTH, HEADER_HEIGHT)
  ctx.fillStyle = TEXT_PRIMARY
  ctx.font = '600 14px sans-serif'
  ctx.fillText('时间', 16, 22)
  ctx.fillStyle = TEXT_MUTED
  ctx.font = '12px sans-serif'
  ctx.fillText(`${TOTAL_ROWS} 个时段`, 16, 40)

  ctx.save()
  ctx.beginPath()
  ctx.rect(TIME_GUTTER_WIDTH, bodyTop, visibleWidth - TIME_GUTTER_WIDTH, visibleHeight - bodyTop)
  ctx.clip()

  const visibleRects: PaintedCourseRect[] = []
  for (let dayIndex = firstVisibleDay; dayIndex <= lastVisibleDay; dayIndex += 1) {
    const columnX = TIME_GUTTER_WIDTH + dayIndex * DAY_COLUMN_WIDTH - horizontalOffset
    const dayCourses = coursesByDay.value[dayIndex] || []
    for (const course of dayCourses) {
      if (course.endRow < firstVisibleRow || course.startRow > lastVisibleRow)
        continue

      const lanes = Math.max(1, course.laneCount)
      const laneWidth = (DAY_COLUMN_WIDTH - COURSE_HORIZONTAL_GAP * (lanes + 1)) / lanes
      const x = columnX + COURSE_HORIZONTAL_GAP + course.laneIndex * (laneWidth + COURSE_HORIZONTAL_GAP)
      const y = HEADER_HEIGHT + course.startRow * ROW_HEIGHT - verticalOffset + COURSE_VERTICAL_GAP
      const height = Math.max(ROW_HEIGHT - COURSE_VERTICAL_GAP * 2, (course.endRow - course.startRow) * ROW_HEIGHT - COURSE_VERTICAL_GAP * 2)
      const width = Math.max(44, laneWidth)
      const isSelected = selectedCourseId.value === course.id

      ctx.fillStyle = course.fill
      ctx.globalAlpha = isSelected ? 0.98 : 0.9
      ctx.fillRect(x, y, width, height)
      ctx.globalAlpha = 1

      ctx.strokeStyle = isSelected ? '#111827' : 'rgba(255,255,255,0.9)'
      ctx.lineWidth = isSelected ? 2 : 1
      ctx.strokeRect(x + 0.5, y + 0.5, width - 1, height - 1)

      ctx.fillStyle = '#ffffff'
      ctx.font = '600 12px sans-serif'
      const textWidth = Math.max(0, width - TEXT_LEFT_PADDING * 2)
      ctx.fillText(ellipsisText(ctx, course.title, textWidth), x + TEXT_LEFT_PADDING, y + 16)

      if (height >= 34) {
        ctx.font = '11px sans-serif'
        ctx.fillText(ellipsisText(ctx, course.teacherName, textWidth), x + TEXT_LEFT_PADDING, y + 30)
      }

      if (height >= 48) {
        ctx.font = '11px sans-serif'
        ctx.fillText(ellipsisText(ctx, courseTimeText(course), textWidth), x + TEXT_LEFT_PADDING, y + 44)
      }

      visibleRects.push({
        course,
        x,
        y,
        width,
        height,
      })
    }
  }

  paintedCourseRects.value = visibleRects
  ctx.restore()
}

function renderFrame() {
  pendingFrame = 0
  const ctx = syncCanvasSize()
  if (!ctx)
    return
  drawGrid(ctx)
}

function scheduleRender() {
  if (pendingFrame)
    return
  pendingFrame = window.requestAnimationFrame(renderFrame)
}

function handleScroll(event: Event) {
  const currentTarget = event.currentTarget as HTMLElement | null
  if (!(currentTarget instanceof HTMLElement))
    return
  scrollTop.value = currentTarget.scrollTop
  scrollLeft.value = currentTarget.scrollLeft
  scheduleRender()
}

function handleCanvasClick(event: MouseEvent) {
  const canvas = canvasRef.value
  if (!(canvas instanceof HTMLCanvasElement))
    return
  const rect = canvas.getBoundingClientRect()
  const pointerX = event.clientX - rect.left
  const pointerY = event.clientY - rect.top

  for (let index = paintedCourseRects.value.length - 1; index >= 0; index -= 1) {
    const item = paintedCourseRects.value[index]
    const withinX = pointerX >= item.x && pointerX <= item.x + item.width
    const withinY = pointerY >= item.y && pointerY <= item.y + item.height
    if (withinX && withinY) {
      selectedCourseId.value = item.course.id
      scheduleRender()
      return
    }
  }

  selectedCourseId.value = ''
  scheduleRender()
}

function installResizeObserver() {
  if (!(viewportRef.value instanceof HTMLElement) || typeof ResizeObserver === 'undefined')
    return
  resizeObserver?.disconnect()
  resizeObserver = new ResizeObserver(() => {
    resizeTick.value += 1
    updateViewportMetrics()
    scheduleRender()
  })
  resizeObserver.observe(viewportRef.value)
}

watch([viewportHeight, viewportWidth], () => {
  nextTick(() => scheduleRender())
})

watch(selectedCourseId, () => {
  scheduleRender()
})

watch(resizeTick, () => {
  scheduleRender()
})

onMounted(() => {
  updateViewportMetrics()
  installResizeObserver()
  const viewport = viewportRef.value
  if (viewport) {
    scrollTop.value = viewport.scrollTop
    scrollLeft.value = viewport.scrollLeft
  }
  emit('week-range-change', {
    startDate: weekDates.value[0].format('YYYY-MM-DD'),
    endDate: weekDates.value[6].format('YYYY-MM-DD'),
  })
  scheduleRender()
  window.addEventListener('resize', updateViewportMetrics)
})

onUnmounted(() => {
  resizeObserver?.disconnect()
  resizeObserver = null
  if (pendingFrame)
    window.cancelAnimationFrame(pendingFrame)
  window.removeEventListener('resize', updateViewportMetrics)
})
</script>

<template>
  <div ref="shellRef" class="canvas-ultra-shell">
    <div class="canvas-ultra-summary">
      <div
        v-for="item in summaryCards"
        :key="item.label"
        class="canvas-ultra-summary__card"
      >
        <div class="canvas-ultra-summary__label">
          {{ item.label }}
        </div>
        <div class="canvas-ultra-summary__value">
          {{ item.value }}
        </div>
      </div>
    </div>

    <div
      ref="viewportRef"
      class="canvas-ultra-viewport"
      :style="{ height: `${viewportHeight}px` }"
      @scroll.passive="handleScroll"
    >
      <canvas
        ref="canvasRef"
        class="canvas-ultra-canvas"
        @click="handleCanvasClick"
      />
      <div
        class="canvas-ultra-spacer"
        :style="{
          width: `${totalGridWidth}px`,
          height: `${totalGridHeight}px`,
        }"
      />
    </div>

    <div class="canvas-ultra-footer">
      <div class="canvas-ultra-footer__meta">
        全部格子与课程均由 Canvas 绘制，滚动时只重绘当前可视区域。
      </div>
      <div class="canvas-ultra-footer__detail">
        <template v-if="selectedCourse">
          <strong>{{ selectedCourse.title }}</strong>
          <span>{{ dayLabels[selectedCourse.dayIndex] }} {{ courseTimeText(selectedCourse) }}</span>
          <span>{{ selectedCourse.teacherName }} / {{ selectedCourse.studentName }} / {{ selectedCourse.classroomName }}</span>
        </template>
        <template v-else>
          点击任意课程块查看详情
        </template>
      </div>
    </div>
  </div>
</template>

<style scoped lang="less">
.canvas-ultra-shell {
  padding: 16px;
  background: linear-gradient(180deg, #f7f9fc 0%, #eef3fb 100%);
}

.canvas-ultra-summary {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 16px;
}

.canvas-ultra-summary__card {
  padding: 14px 16px;
  border: 1px solid #dce5f2;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: 0 12px 28px rgba(148, 163, 184, 0.14);
}

.canvas-ultra-summary__label {
  color: #64748b;
  font-size: 12px;
  line-height: 18px;
}

.canvas-ultra-summary__value {
  margin-top: 4px;
  color: #0f172a;
  font-size: 22px;
  font-weight: 700;
  line-height: 28px;
}

.canvas-ultra-viewport {
  position: relative;
  overflow: auto;
  border: 1px solid #d8dfeb;
  border-radius: 16px;
  background: #f6f8fc;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.9);
}

.canvas-ultra-spacer {
  position: relative;
}

.canvas-ultra-canvas {
  position: sticky;
  top: 0;
  left: 0;
  z-index: 2;
  display: block;
  width: 100%;
  height: 100%;
  cursor: default;
}

.canvas-ultra-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-top: 14px;
  padding: 12px 16px;
  border: 1px solid #dce5f2;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.9);
}

.canvas-ultra-footer__meta {
  color: #64748b;
  font-size: 12px;
  line-height: 18px;
}

.canvas-ultra-footer__detail {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 12px;
  color: #0f172a;
  font-size: 13px;
  line-height: 20px;
}

@media (max-width: 960px) {
  .canvas-ultra-summary {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .canvas-ultra-footer {
    flex-direction: column;
    align-items: flex-start;
  }

  .canvas-ultra-footer__detail {
    justify-content: flex-start;
  }
}
</style>
