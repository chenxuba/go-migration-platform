<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
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
  viewportHeight: {
    type: Number,
    default: 0,
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

const HEADER_HEIGHT = 52
const ROW_HEIGHT = 52
const CELL_PADDING = 4
const CELL_RADIUS = 4
const HEADER_BG = '#ffffff'
const GRID_BG = '#ffffff'
const GRID_LINE = '#f0f0f0'
const STRONG_GRID_LINE = '#d9d9d9'
const TEXT_PRIMARY = '#262626'
const TEXT_SECONDARY = '#666666'
const TEXT_MUTED = '#8c8c8c'
const EMPTY_IDLE_TEXT = 'transparent'
const VIEWPORT_MIN_HEIGHT = 520
const VIEWPORT_MAX_HEIGHT = 1320
const VIEWPORT_BOTTOM_GAP = 0
const HEADER_FONT = '500 14px sans-serif'
const SUB_HEADER_FONT = '12px sans-serif'
const CELL_FONT = '14px sans-serif'
const CELL_SUB_FONT = '12px sans-serif'
const BLOCK_HEADER_HEIGHT = 20
const BADGE_HEIGHT = 16

const shellRef = ref(null)
const viewportRef = ref(null)
const canvasRef = ref(null)
const hoverScheduleAnchorRef = ref(null)
const viewportHeight = ref(420)
const viewportWidth = ref(0)
const scrollTop = ref(0)
const scrollLeft = ref(0)
const openSchedulePopoverKey = ref('')
const hoveredScheduleCellKey = ref('')
const hoveredScheduleRect = ref(null)
const emptyDragCellStateMap = ref({})
const scheduleCellPressState = ref(null)
let pendingEmptyDragCellStateMap = {}
let emptyDragCellStateFlushFrame = 0

let resizeObserver = null
let pendingFrame = 0
let hoverCloseTimer = null
let scheduleCellPressMoveHandler = null
let scheduleCellPressUpHandler = null
let suppressedScheduleCellClickKey = ''
let suppressedScheduleCellClickUntil = 0

const fixedColumns = computed(() =>
  props.columns.filter(column => !props.isScheduleColumn(column)),
)

const scheduleColumns = computed(() =>
  props.columns.filter(column => props.isScheduleColumn(column)),
)

const fixedColumnMetas = computed(() => {
  let left = 0
  return fixedColumns.value.map((column, index) => {
    const width = Number(column?.width || (index === 0 ? 120 : 80))
    const meta = {
      column,
      index,
      left,
      width,
    }
    left += width
    return meta
  })
})

const scheduleColumnMetas = computed(() => {
  const baseWidths = scheduleColumns.value.map(column => Number(column?.width || 160))
  const baseTotalWidth = baseWidths.reduce((sum, width) => sum + width, 0)
  const targetBodyWidth = Math.max(
    baseTotalWidth,
    Math.max(0, viewportWidth.value - fixedLeftWidth.value),
  )
  let left = 0
  return scheduleColumns.value.map((column, index) => {
    const baseWidth = baseWidths[index] || 0
    const accumulatedBaseWidth = baseWidths
      .slice(0, index + 1)
      .reduce((sum, width) => sum + width, 0)
    const width = index === scheduleColumns.value.length - 1
      ? Math.max(baseWidth, targetBodyWidth - left)
      : Math.max(
          baseWidth,
          Math.round(accumulatedBaseWidth * targetBodyWidth / Math.max(baseTotalWidth, 1)) - left,
        )
    const meta = {
      column,
      index,
      left,
      width,
    }
    left += width
    return meta
  })
})

const fixedLeftWidth = computed(() =>
  fixedColumnMetas.value.reduce((sum, item) => sum + item.width, 0),
)

const bodyWidth = computed(() =>
  scheduleColumnMetas.value.reduce((sum, item) => sum + item.width, 0),
)

const totalGridWidth = computed(() => fixedLeftWidth.value + bodyWidth.value)
const totalGridHeight = computed(() => HEADER_HEIGHT + props.tableDataSource.length * ROW_HEIGHT)

const rowMetas = computed(() => {
  return props.tableDataSource.map((record, rowIndex) => ({
    record,
    rowIndex,
    top: HEADER_HEIGHT + rowIndex * ROW_HEIGHT,
    height: ROW_HEIGHT,
  }))
})

const teacherSpanByStartRow = computed(() => {
  const map = new Map()
  const rows = rowMetas.value
  let index = 0
  while (index < rows.length) {
    const startIndex = index
    const teacherId = String(rows[index]?.record?.teacherId || '')
    while (
      index + 1 < rows.length
      && String(rows[index + 1]?.record?.teacherId || '') === teacherId
    ) {
      index += 1
    }
    map.set(startIndex, {
      rowIndex: startIndex,
      rowCount: index - startIndex + 1,
      record: rows[startIndex]?.record || null,
      top: HEADER_HEIGHT + startIndex * ROW_HEIGHT,
      height: (index - startIndex + 1) * ROW_HEIGHT,
    })
    index += 1
  }
  return map
})

function getBodyCellValue(column, record) {
  const root = column?.dataIndex?.[0]
  const index = Number(column?.dataIndex?.[1])
  if (!Number.isInteger(index))
    return null
  if (root === 'lessons')
    return record?.lessons?.[index] || null
  if (root === 'cells')
    return record?.cells?.[index] || null
  return null
}

function buildCellEntry(rowMeta, columnMeta) {
  const column = columnMeta.column
  const record = rowMeta.record
  const text = getBodyCellValue(column, record)
  const key = String(props.scheduleCellKey(column, record) || '').trim()
  const contextColumn = props.scheduleCellContextColumn(column, record)
  const contextRecord = props.scheduleCellContextRecord(column, record)
  return {
    key,
    rowIndex: rowMeta.rowIndex,
    columnIndex: columnMeta.index,
    contentX: fixedLeftWidth.value + columnMeta.left,
    contentY: HEADER_HEIGHT + rowMeta.rowIndex * ROW_HEIGHT,
    width: columnMeta.width,
    height: ROW_HEIGHT,
    text,
    column,
    record,
    contextColumn,
    contextRecord,
  }
}

const scheduleCellEntryMap = computed(() => {
  const map = new Map()
  rowMetas.value.forEach((rowMeta) => {
    scheduleColumnMetas.value.forEach((columnMeta) => {
      const entry = buildCellEntry(rowMeta, columnMeta)
      if (entry.key && props.hasScheduledLesson(entry.text))
        map.set(entry.key, entry)
    })
  })
  return map
})

const emptyCellEntryMap = computed(() => {
  const map = new Map()
  rowMetas.value.forEach((rowMeta) => {
    scheduleColumnMetas.value.forEach((columnMeta) => {
      const entry = buildCellEntry(rowMeta, columnMeta)
      if (entry.key && !props.hasScheduledLesson(entry.text))
        map.set(entry.key, entry)
    })
  })
  return map
})

function schedulePopoverKeyByEntry(entry) {
  return String(entry?.key || '').trim()
}

function conflictBadgeTooltip(text) {
  const types = Array.isArray(text?.scheduledConflictTypes)
    ? text.scheduledConflictTypes.filter(Boolean)
    : []
  if (types.length)
    return `冲突原因：${types.join('、')}冲突，点击查看详情`
  return '当前课程存在冲突，点击查看详情'
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

function scheduleModeShortLabel(text) {
  if (text?.courseType === 2)
    return '班课'
  return '1v1'
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

function ellipsisText(ctx, text, maxWidth) {
  const input = String(text || '')
  if (!input)
    return ''
  if (ctx.measureText(input).width <= maxWidth)
    return input
  let output = input
  while (output.length > 1 && ctx.measureText(`${output}…`).width > maxWidth)
    output = output.slice(0, -1)
  return `${output}…`
}

function drawRoundedRect(ctx, x, y, width, height, radius) {
  const safeRadius = Math.max(0, Math.min(radius, width / 2, height / 2))
  ctx.beginPath()
  ctx.moveTo(x + safeRadius, y)
  ctx.arcTo(x + width, y, x + width, y + height, safeRadius)
  ctx.arcTo(x + width, y + height, x, y + height, safeRadius)
  ctx.arcTo(x, y + height, x, y, safeRadius)
  ctx.arcTo(x, y, x + width, y, safeRadius)
  ctx.closePath()
}

function viewportClientRect() {
  if (!(viewportRef.value instanceof HTMLElement))
    return null
  return viewportRef.value.getBoundingClientRect()
}

function clamp(value, min, max) {
  return Math.min(max, Math.max(min, value))
}

function contentRectToViewRect(entry) {
  return {
    left: entry.contentX - scrollLeft.value,
    top: entry.contentY - scrollTop.value,
    width: entry.width,
    height: entry.height,
    right: entry.contentX - scrollLeft.value + entry.width,
    bottom: entry.contentY - scrollTop.value + entry.height,
  }
}

function viewRectToClientRect(rect) {
  const viewportRect = viewportClientRect()
  if (!viewportRect || !rect)
    return null
  return {
    left: viewportRect.left + rect.left,
    top: viewportRect.top + rect.top,
    width: rect.width,
    height: rect.height,
    right: viewportRect.left + rect.right,
    bottom: viewportRect.top + rect.bottom,
  }
}

function getScheduleCellRect(key) {
  const entry = scheduleCellEntryMap.value.get(String(key || '').trim())
  if (!entry)
    return null
  return viewRectToClientRect(contentRectToViewRect(entry))
}

function getEmptyDragCellRect(key) {
  const entry = emptyCellEntryMap.value.get(String(key || '').trim())
  if (!entry)
    return null
  return viewRectToClientRect(contentRectToViewRect(entry))
}

function getScheduleDragStartInfo(key, clientX = 0, clientY = 0) {
  const rect = getScheduleCellRect(key)
  if (!rect)
    return null
  return {
    rect,
    offsetX: clamp(Number(clientX || 0) - rect.left, 8, Math.max(8, rect.width - 8)),
    offsetY: clamp(Number(clientY || 0) - rect.top, 8, Math.max(8, rect.height - 8)),
  }
}

function getViewportClientRect() {
  return viewportClientRect()
}

function buildDragEventPayload(nativeEvent, entry, currentTarget) {
  const rect = viewRectToClientRect(contentRectToViewRect(entry))
  return {
    clientX: Number(nativeEvent?.clientX || 0),
    clientY: Number(nativeEvent?.clientY || 0),
    currentTarget,
    __smartTimetableRect: rect,
    __smartTimetableOffsetX: rect
      ? clamp(Number(nativeEvent?.clientX || 0) - rect.left, 8, Math.max(8, rect.width - 8))
      : 0,
    __smartTimetableOffsetY: rect
      ? clamp(Number(nativeEvent?.clientY || 0) - rect.top, 8, Math.max(8, rect.height - 8))
      : 0,
    __smartTimetableUseFloatingPreview: true,
  }
}

function resolveCanvasPoint(clientX, clientY) {
  const viewportRect = viewportClientRect()
  if (!viewportRect)
    return null
  const x = Number(clientX || 0) - viewportRect.left
  const y = Number(clientY || 0) - viewportRect.top
  if (x < 0 || y < 0 || x > viewportRect.width || y > viewportRect.height)
    return null
  return { x, y }
}

function resolveVisibleRowRange(visibleHeight) {
  const firstRow = Math.max(0, Math.floor(Math.max(0, scrollTop.value) / ROW_HEIGHT) - 1)
  const lastRow = Math.min(
    rowMetas.value.length - 1,
    Math.ceil((scrollTop.value + Math.max(0, visibleHeight - HEADER_HEIGHT)) / ROW_HEIGHT) + 1,
  )
  return {
    firstRow,
    lastRow,
  }
}

function resolveVisibleBodyColumns(visibleWidth) {
  const viewportBodyWidth = Math.max(0, visibleWidth - fixedLeftWidth.value)
  const left = scrollLeft.value
  const right = scrollLeft.value + viewportBodyWidth
  return scheduleColumnMetas.value.filter(item =>
    item.left + item.width >= left && item.left <= right,
  )
}

function resolveEntryFromViewPoint(viewX, viewY) {
  if (viewX < fixedLeftWidth.value || viewY < HEADER_HEIGHT)
    return null

  const bodyX = scrollLeft.value + viewX - fixedLeftWidth.value
  const bodyY = scrollTop.value + viewY - HEADER_HEIGHT
  if (bodyX < 0 || bodyY < 0)
    return null

  const rowIndex = Math.floor(bodyY / ROW_HEIGHT)
  if (rowIndex < 0 || rowIndex >= rowMetas.value.length)
    return null

  const columnMeta = scheduleColumnMetas.value.find(item =>
    bodyX >= item.left && bodyX <= item.left + item.width,
  )
  if (!columnMeta)
    return null

  return buildCellEntry(rowMetas.value[rowIndex], columnMeta)
}

function resolvePointerDragTarget(clientX, clientY) {
  const point = resolveCanvasPoint(clientX, clientY)
  if (!point)
    return null
  const entry = resolveEntryFromViewPoint(point.x, point.y)
  if (!entry || props.hasScheduledLesson(entry.text))
    return null
  return {
    key: entry.key,
    teacherId: String(entry?.record?.teacherId || '').trim(),
    teacherName: String(entry?.record?.name || '').trim() || '-',
    lessonDate: String(entry?.contextRecord?.date || '').trim(),
    startTime: String(entry?.contextColumn?.startTime || '').trim(),
    endTime: String(entry?.contextColumn?.endTime || '').trim(),
  }
}

function setEmptyDragCellState(target, state) {
  const key = String(target?.key || state?.key || '').trim()
  if (!key)
    return
  pendingEmptyDragCellStateMap[key] = {
    ...(emptyDragCellStateMap.value[key] || {}),
    ...(pendingEmptyDragCellStateMap[key] || {}),
    ...(target || {}),
    ...(state || {}),
    key,
  }
  if (!emptyDragCellStateFlushFrame && typeof window !== 'undefined') {
    emptyDragCellStateFlushFrame = window.requestAnimationFrame(() => {
      emptyDragCellStateFlushFrame = 0
      if (Object.keys(pendingEmptyDragCellStateMap).length) {
        emptyDragCellStateMap.value = {
          ...emptyDragCellStateMap.value,
          ...pendingEmptyDragCellStateMap,
        }
        pendingEmptyDragCellStateMap = {}
      }
      scheduleRender()
    })
  }
  else if (!emptyDragCellStateFlushFrame) {
    emptyDragCellStateMap.value = {
      ...emptyDragCellStateMap.value,
      [key]: pendingEmptyDragCellStateMap[key],
    }
    pendingEmptyDragCellStateMap = {}
    scheduleRender()
  }
}

function clearEmptyDragCellStates() {
  pendingEmptyDragCellStateMap = {}
  if (emptyDragCellStateFlushFrame && typeof window !== 'undefined') {
    window.cancelAnimationFrame(emptyDragCellStateFlushFrame)
    emptyDragCellStateFlushFrame = 0
  }
  emptyDragCellStateMap.value = {}
  scheduleRender()
}

async function scrollToScheduleCell(key) {
  await nextTick()
  const entry = scheduleCellEntryMap.value.get(String(key || '').trim())
  const viewport = viewportRef.value
  if (!entry || !(viewport instanceof HTMLElement))
    return false

  const bodyVisibleWidth = Math.max(0, viewport.clientWidth - fixedLeftWidth.value)
  const desiredLeft = clamp(
    entry.contentX - fixedLeftWidth.value - Math.max(0, bodyVisibleWidth - entry.width) / 2,
    0,
    Math.max(0, totalGridWidth.value - viewport.clientWidth),
  )
  const desiredTop = clamp(
    entry.contentY - HEADER_HEIGHT - Math.max(0, viewport.clientHeight - HEADER_HEIGHT - entry.height) / 2,
    0,
    Math.max(0, totalGridHeight.value - viewport.clientHeight),
  )
  viewport.scrollTo({
    left: desiredLeft,
    top: desiredTop,
    behavior: 'smooth',
  })
  scrollLeft.value = desiredLeft
  scrollTop.value = desiredTop
  updateHoveredScheduleRect()
  scheduleRender()
  return true
}

function schedulePaintStyle(text) {
  if (text?.callStatusKey === 'signed') {
    return {
      background: '#f4f5f7',
      header: '#9ea4b0',
      text: '#707784',
      border: 'rgba(15, 23, 42, 0.06)',
      badge: 'rgba(0, 0, 0, 0.42)',
    }
  }
  if (text?.callStatusKey === 'partial') {
    return {
      background: '#fff1db',
      header: '#f59e0b',
      text: '#9a3412',
      border: 'rgba(245, 158, 11, 0.18)',
      badge: 'rgba(120, 53, 15, 0.24)',
    }
  }
  return {
    background: 'rgba(78, 109, 255, 0.12)',
    header: '#0066ff',
    text: '#002cfd',
    border: 'rgba(78, 109, 255, 0.08)',
    badge: 'rgba(0, 0, 0, 0.5)',
  }
}

function emptyCellPaintStyle(entry) {
  const dragState = emptyDragCellStateMap.value[entry.key]
  if (dragState?.checking) {
    return {
      fill: '#fff7e6',
      stroke: 'rgba(250, 173, 20, 0.28)',
      text: '#d48806',
      label: String(dragState.label || '检测中').trim(),
    }
  }
  if (dragState?.valid === true) {
    return {
      fill: '#fff7e6',
      stroke: 'rgba(250, 173, 20, 0.34)',
      text: '#d48806',
      label: String(dragState.label || '可调课').trim(),
    }
  }
  if (dragState?.valid === false) {
    return {
      fill: '#ffe1df',
      stroke: 'rgba(255, 77, 79, 0.4)',
      text: '#cf1322',
      label: String(dragState.label || '不可调').trim(),
    }
  }

  const label = String(props.emptyLessonStatusText(entry.text) || '').trim()
  if (label) {
    if (entry.text?.conflict) {
      return {
        fill: '#ffe6e6',
        stroke: 'rgba(255, 77, 79, 0.24)',
        text: '#a31616',
        label,
      }
    }
    return {
      fill: '#e6ffe6',
      stroke: 'rgba(22, 163, 74, 0.18)',
      text: '#16a34a',
      label,
    }
  }

  return {
    fill: 'transparent',
    stroke: 'transparent',
    text: EMPTY_IDLE_TEXT,
    label: '',
  }
}

function drawHeader(ctx, visibleWidth) {
  ctx.fillStyle = HEADER_BG
  ctx.fillRect(0, 0, visibleWidth, HEADER_HEIGHT)

  const visibleColumns = resolveVisibleBodyColumns(visibleWidth)
  ctx.save()
  ctx.beginPath()
  ctx.rect(fixedLeftWidth.value, 0, Math.max(0, visibleWidth - fixedLeftWidth.value), HEADER_HEIGHT)
  ctx.clip()

  visibleColumns.forEach((columnMeta) => {
    const x = fixedLeftWidth.value + columnMeta.left - scrollLeft.value
    if (x > visibleWidth || x + columnMeta.width < fixedLeftWidth.value)
      return

    ctx.fillStyle = HEADER_BG
    ctx.fillRect(x, 0, columnMeta.width, HEADER_HEIGHT)
    ctx.strokeStyle = GRID_LINE
    ctx.beginPath()
    ctx.moveTo(x + 0.5, 0)
    ctx.lineTo(x + 0.5, HEADER_HEIGHT)
    ctx.stroke()
    ctx.strokeStyle = STRONG_GRID_LINE
    ctx.beginPath()
    ctx.moveTo(x + columnMeta.width + 0.5, 0)
    ctx.lineTo(x + columnMeta.width + 0.5, HEADER_HEIGHT)
    ctx.stroke()

    ctx.fillStyle = TEXT_PRIMARY
    ctx.font = HEADER_FONT
    ctx.textAlign = 'center'
    ctx.textBaseline = 'middle'

    if (!props.isSwapTimeGrid && columnMeta.column?.startTime && columnMeta.column?.endTime) {
      ctx.fillText(String(columnMeta.column?.title || ''), x + columnMeta.width / 2, 18)
      ctx.fillStyle = TEXT_SECONDARY
      ctx.font = SUB_HEADER_FONT
      ctx.fillText(`${columnMeta.column.startTime}-${columnMeta.column.endTime}`, x + columnMeta.width / 2, 36)
      return
    }

    if (props.isSwapTimeGrid && columnMeta.column?.date) {
      ctx.fillText(String(columnMeta.column?.title || ''), x + columnMeta.width / 2, 18)
      ctx.fillStyle = TEXT_SECONDARY
      ctx.font = SUB_HEADER_FONT
      ctx.fillText(String(columnMeta.column?.dateText || ''), x + columnMeta.width / 2, 36)
      return
    }

    ctx.fillText(String(columnMeta.column?.title || ''), x + columnMeta.width / 2, HEADER_HEIGHT / 2)
  })
  ctx.restore()

  fixedColumnMetas.value.forEach((columnMeta) => {
    const x = columnMeta.left
    ctx.fillStyle = HEADER_BG
    ctx.fillRect(x, 0, columnMeta.width, HEADER_HEIGHT)
    ctx.strokeStyle = GRID_LINE
    ctx.beginPath()
    ctx.moveTo(x + columnMeta.width + 0.5, 0)
    ctx.lineTo(x + columnMeta.width + 0.5, HEADER_HEIGHT)
    ctx.stroke()

    ctx.fillStyle = TEXT_PRIMARY
    ctx.font = HEADER_FONT
    ctx.textAlign = 'center'
    ctx.textBaseline = 'middle'
    ctx.fillText(String(columnMeta.column?.title || ''), x + columnMeta.width / 2, HEADER_HEIGHT / 2)
  })

  ctx.strokeStyle = STRONG_GRID_LINE
  ctx.beginPath()
  ctx.moveTo(0, HEADER_HEIGHT + 0.5)
  ctx.lineTo(visibleWidth, HEADER_HEIGHT + 0.5)
  ctx.stroke()
}

function drawFixedColumns(ctx, visibleHeight) {
  const { firstRow, lastRow } = resolveVisibleRowRange(visibleHeight)
  const rows = rowMetas.value.slice(firstRow, lastRow + 1)

  ctx.fillStyle = HEADER_BG
  ctx.fillRect(0, HEADER_HEIGHT, fixedLeftWidth.value, Math.max(0, visibleHeight - HEADER_HEIGHT))

  rows.forEach((rowMeta) => {
    const y = HEADER_HEIGHT + rowMeta.rowIndex * ROW_HEIGHT - scrollTop.value
    ctx.strokeStyle = GRID_LINE
    ctx.beginPath()
    ctx.moveTo(0, y + ROW_HEIGHT + 0.5)
    ctx.lineTo(fixedLeftWidth.value, y + ROW_HEIGHT + 0.5)
    ctx.stroke()
  })

  const secondFixed = fixedColumnMetas.value[1]
  if (secondFixed) {
    rows.forEach((rowMeta) => {
      const record = rowMeta.record
      const y = HEADER_HEIGHT + rowMeta.rowIndex * ROW_HEIGHT - scrollTop.value
      const x = secondFixed.left
      ctx.fillStyle = '#ffffff'
      ctx.fillRect(x, y, secondFixed.width, ROW_HEIGHT)
      ctx.strokeStyle = GRID_LINE
      ctx.strokeRect(x + 0.5, y + 0.5, secondFixed.width - 1, ROW_HEIGHT - 1)

      ctx.fillStyle = TEXT_PRIMARY
      ctx.font = CELL_FONT
      ctx.textAlign = 'center'
      ctx.textBaseline = 'middle'

      if (secondFixed.column?.key === 'date') {
        ctx.fillText(String(props.formatWeek(record?.date) || ''), x + secondFixed.width / 2, y + 18)
        ctx.fillStyle = TEXT_SECONDARY
        ctx.font = CELL_SUB_FONT
        ctx.fillText(String(props.formatDate(record?.date) || ''), x + secondFixed.width / 2, y + 35)
        return
      }

      if (secondFixed.column?.key === 'slot') {
        ctx.fillStyle = TEXT_PRIMARY
        ctx.font = CELL_FONT
        ctx.fillText(String(record?.slotLabel || ''), x + secondFixed.width / 2, y + 18)
        ctx.fillStyle = TEXT_SECONDARY
        ctx.font = CELL_SUB_FONT
        ctx.fillText(`${record?.startTime || ''}-${record?.endTime || ''}`, x + secondFixed.width / 2, y + 35)
      }
    })
  }

  const firstFixed = fixedColumnMetas.value[0]
  if (!firstFixed)
    return

  teacherSpanByStartRow.value.forEach((spanMeta) => {
    const y = spanMeta.top - scrollTop.value
    const bottom = y + spanMeta.height
    if (bottom < HEADER_HEIGHT || y > visibleHeight)
      return

    ctx.fillStyle = '#ffffff'
    ctx.fillRect(firstFixed.left, y, firstFixed.width, spanMeta.height)
    ctx.strokeStyle = GRID_LINE
    ctx.strokeRect(firstFixed.left + 0.5, y + 0.5, firstFixed.width - 1, spanMeta.height - 1)

    const teacherName = String(spanMeta.record?.name || '')
    const lessonCount = String(props.teacherLessonCountLabel(spanMeta.record?.teacherId) || '')
    const centerY = y + spanMeta.height / 2
    const hasLessonCount = Boolean(lessonCount)

    ctx.fillStyle = TEXT_PRIMARY
    ctx.font = CELL_FONT
    ctx.textAlign = 'center'
    ctx.textBaseline = 'middle'
    ctx.fillText(
      teacherName,
      firstFixed.left + firstFixed.width / 2,
      hasLessonCount ? centerY - 10 : centerY,
    )

    if (hasLessonCount) {
      ctx.fillStyle = TEXT_SECONDARY
      ctx.font = CELL_SUB_FONT
      ctx.fillText(
        lessonCount,
        firstFixed.left + firstFixed.width / 2,
        centerY + 10,
      )
    }
  })

  ctx.strokeStyle = GRID_LINE
  ctx.beginPath()
  ctx.moveTo(fixedLeftWidth.value + 0.5, HEADER_HEIGHT)
  ctx.lineTo(fixedLeftWidth.value + 0.5, visibleHeight)
  ctx.stroke()
}

function drawEmptyCell(ctx, entry, viewRect) {
  const innerX = viewRect.left + CELL_PADDING
  const innerY = viewRect.top + CELL_PADDING
  const innerWidth = Math.max(0, viewRect.width - CELL_PADDING * 2)
  const innerHeight = Math.max(0, viewRect.height - CELL_PADDING * 2)
  const paint = emptyCellPaintStyle(entry)

  if (paint.fill !== 'transparent') {
    drawRoundedRect(ctx, innerX, innerY, innerWidth, innerHeight, CELL_RADIUS)
    ctx.fillStyle = paint.fill
    ctx.fill()
    if (paint.stroke !== 'transparent') {
      ctx.strokeStyle = paint.stroke
      ctx.stroke()
    }
  }

  if (paint.label) {
    ctx.fillStyle = paint.text
    ctx.font = CELL_FONT
    ctx.textAlign = 'center'
    ctx.textBaseline = 'middle'
    ctx.fillText(ellipsisText(ctx, paint.label, innerWidth - 12), innerX + innerWidth / 2, innerY + innerHeight / 2)
  }
}

function drawScheduleCell(ctx, entry, viewRect) {
  const text = entry.text
  const style = schedulePaintStyle(text)
  const innerX = viewRect.left + CELL_PADDING
  const innerY = viewRect.top + CELL_PADDING
  const innerWidth = Math.max(0, viewRect.width - CELL_PADDING * 2)
  const innerHeight = Math.max(0, viewRect.height - CELL_PADDING * 2)
  const bodyWidth = innerWidth - 12

  drawRoundedRect(ctx, innerX, innerY, innerWidth, innerHeight, CELL_RADIUS)
  ctx.fillStyle = style.background
  ctx.fill()

  if (text?.scheduledConflict) {
    ctx.strokeStyle = 'rgba(255, 77, 79, 0.4)'
    ctx.lineWidth = 2
    ctx.stroke()
    ctx.lineWidth = 1
  }
  else {
    ctx.strokeStyle = style.border
    ctx.stroke()
  }

  ctx.save()
  drawRoundedRect(ctx, innerX, innerY, innerWidth, innerHeight, CELL_RADIUS)
  ctx.clip()

  ctx.save()
  drawRoundedRect(ctx, innerX, innerY, innerWidth, BLOCK_HEADER_HEIGHT, CELL_RADIUS)
  ctx.clip()
  ctx.fillStyle = style.header
  ctx.fillRect(innerX, innerY, innerWidth, BLOCK_HEADER_HEIGHT + 2)
  ctx.restore()

  ctx.fillStyle = '#ffffff'
  ctx.font = '12px sans-serif'
  ctx.textAlign = 'left'
  ctx.textBaseline = 'middle'
  ctx.fillText(
    `${props.scheduleCellStartTime(entry.column, entry.record)}-${props.scheduleCellEndTime(entry.column, entry.record)}`,
    innerX + 6,
    innerY + BLOCK_HEADER_HEIGHT / 2,
  )

  const badgeText = text?.scheduledConflict
    ? '冲突'
    : (text?.courseType === 1 ? '1v1' : `班课(${text?.isMain ? '主教' : '辅教'})`)
  ctx.font = '500 10px sans-serif'
  const badgeWidth = Math.min(innerWidth - 10, Math.max(26, ctx.measureText(badgeText).width + 12))
  const badgeX = innerX + innerWidth - badgeWidth
  const badgeY = innerY

  drawRoundedRect(ctx, badgeX, badgeY, badgeWidth, BADGE_HEIGHT, 3)
  ctx.fillStyle = text?.scheduledConflict ? '#ff4d4f' : style.badge
  ctx.fill()
  ctx.fillStyle = '#ffffff'
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'
  ctx.fillText(badgeText, badgeX + badgeWidth / 2, badgeY + BADGE_HEIGHT / 2)

  ctx.fillStyle = style.text
  ctx.textAlign = 'left'
  ctx.textBaseline = 'top'
  ctx.font = '13px sans-serif'

  if (!text?.classId) {
    const lines = Array.isArray(text?.studentNames)
      ? text.studentNames.map((item, index) => scheduleStudentLine(text, item?.name, index !== text.studentNames.length - 1)).filter(Boolean)
      : []
    const lineText = lines.join('')
    ctx.fillText(ellipsisText(ctx, lineText, bodyWidth), innerX + 6, innerY + 26)
  }
  else {
    ctx.fillText(ellipsisText(ctx, scheduleClassLine(text), bodyWidth), innerX + 6, innerY + 26)
  }

  ctx.restore()

  if (String(props.focusedScheduleCellKey || '').trim() === entry.key) {
    drawRoundedRect(ctx, innerX - 1, innerY - 1, innerWidth + 2, innerHeight + 2, CELL_RADIUS + 1)
    ctx.strokeStyle = 'rgba(255, 77, 79, 0.95)'
    ctx.lineWidth = 3
    ctx.stroke()
    ctx.lineWidth = 1
  }
}

function drawBodyGrid(ctx, visibleWidth, visibleHeight) {
  const visibleColumns = resolveVisibleBodyColumns(visibleWidth)
  const { firstRow, lastRow } = resolveVisibleRowRange(visibleHeight)

  ctx.save()
  ctx.beginPath()
  ctx.rect(fixedLeftWidth.value, HEADER_HEIGHT, Math.max(0, visibleWidth - fixedLeftWidth.value), Math.max(0, visibleHeight - HEADER_HEIGHT))
  ctx.clip()

  ctx.fillStyle = GRID_BG
  ctx.fillRect(fixedLeftWidth.value, HEADER_HEIGHT, Math.max(0, visibleWidth - fixedLeftWidth.value), Math.max(0, visibleHeight - HEADER_HEIGHT))

  visibleColumns.forEach((columnMeta) => {
    const x = fixedLeftWidth.value + columnMeta.left - scrollLeft.value
    ctx.strokeStyle = GRID_LINE
    ctx.beginPath()
    ctx.moveTo(x + 0.5, HEADER_HEIGHT)
    ctx.lineTo(x + 0.5, visibleHeight)
    ctx.stroke()
  })

  for (let rowIndex = firstRow; rowIndex <= lastRow; rowIndex += 1) {
    const y = HEADER_HEIGHT + rowIndex * ROW_HEIGHT - scrollTop.value
    ctx.strokeStyle = GRID_LINE
    ctx.beginPath()
    ctx.moveTo(fixedLeftWidth.value, y + 0.5)
    ctx.lineTo(visibleWidth, y + 0.5)
    ctx.stroke()

    visibleColumns.forEach((columnMeta) => {
      const entry = buildCellEntry(rowMetas.value[rowIndex], columnMeta)
      const viewRect = contentRectToViewRect(entry)
      if (String(props.draggingScheduleCellKey || '').trim() === entry.key)
        return
      if (props.hasScheduledLesson(entry.text))
        drawScheduleCell(ctx, entry, viewRect)
      else
        drawEmptyCell(ctx, entry, viewRect)
    })
  }

  ctx.restore()
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

function renderFrame() {
  pendingFrame = 0
  const ctx = syncCanvasSize()
  if (!ctx || !(canvasRef.value instanceof HTMLCanvasElement))
    return
  const visibleWidth = canvasRef.value.clientWidth
  const visibleHeight = canvasRef.value.clientHeight
  ctx.clearRect(0, 0, visibleWidth, visibleHeight)
  ctx.fillStyle = '#fafafa'
  ctx.fillRect(0, 0, visibleWidth, visibleHeight)
  drawBodyGrid(ctx, visibleWidth, visibleHeight)
  drawFixedColumns(ctx, visibleHeight)
  drawHeader(ctx, visibleWidth)
}

function scheduleRender() {
  if (pendingFrame)
    return
  pendingFrame = window.requestAnimationFrame(renderFrame)
}

function updateViewportMetrics() {
  const shell = shellRef.value
  const viewport = viewportRef.value
  if (!(shell instanceof HTMLElement) || !(viewport instanceof HTMLElement) || typeof window === 'undefined')
    return
  if (Number(props.viewportHeight || 0) > 0) {
    viewportWidth.value = viewport.clientWidth
    viewportHeight.value = Math.max(180, Number(props.viewportHeight || 0))
    return
  }
  const parentHeight = shell.parentElement instanceof HTMLElement
    ? shell.parentElement.clientHeight
    : 0
  if (parentHeight > 0) {
    viewportWidth.value = viewport.clientWidth
    viewportHeight.value = Math.max(180, parentHeight)
    return
  }
  const shellRect = shell.getBoundingClientRect()
  const available = Math.max(VIEWPORT_MIN_HEIGHT, Math.min(VIEWPORT_MAX_HEIGHT, Math.floor(window.innerHeight - shellRect.top - VIEWPORT_BOTTOM_GAP)))
  viewportWidth.value = viewport.clientWidth
  viewportHeight.value = Math.max(180, available)
}

function clearHoverCloseTimer() {
  if (hoverCloseTimer)
    clearTimeout(hoverCloseTimer)
  hoverCloseTimer = null
}

function clearScheduleHover() {
  clearHoverCloseTimer()
  hoveredScheduleCellKey.value = ''
  hoveredScheduleRect.value = null
  openSchedulePopoverKey.value = ''
}

function scheduleCloseHover(delay = 24) {
  clearHoverCloseTimer()
  hoverCloseTimer = setTimeout(() => {
    clearScheduleHover()
  }, delay)
}

function updateHoveredScheduleRect() {
  const key = String(hoveredScheduleCellKey.value || '').trim()
  if (!key) {
    hoveredScheduleRect.value = null
    return
  }
  const entry = scheduleCellEntryMap.value.get(key)
  if (!entry) {
    clearScheduleHover()
    return
  }
  hoveredScheduleRect.value = contentRectToViewRect(entry)
}

function setHoveredScheduleEntry(entry) {
  const key = schedulePopoverKeyByEntry(entry)
  if (!key) {
    scheduleCloseHover()
    return
  }
  clearHoverCloseTimer()
  hoveredScheduleCellKey.value = key
  hoveredScheduleRect.value = contentRectToViewRect(entry)
  if (!String(props.draggingScheduleCellKey || '').trim())
    openSchedulePopoverKey.value = key
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

function handleScheduleCellClickByEntry(entry) {
  if (!entry)
    return
  if (consumeScheduleCellClickSuppressed(schedulePopoverKeyByEntry(entry)))
    return
  if (props.consumeScheduledLessonClickSuppressed())
    return
  props.openScheduledLessonDetail(entry.text, entry.contextColumn, entry.contextRecord)
}

function handleEmptyCellClickByEntry(entry) {
  if (!entry)
    return
  if (entry.text?.conflict)
    props.handleConflictClick(entry.text, entry.contextColumn, entry.contextRecord)
  else
    props.handleScheduleClick(entry.text, entry.contextColumn, entry.contextRecord)
}

function handleSchedulePopoverOpenChange(entry, open) {
  const key = schedulePopoverKeyByEntry(entry)
  clearHoverCloseTimer()
  openSchedulePopoverKey.value = open ? key : ''
  if (open) {
    hoveredScheduleCellKey.value = key
    hoveredScheduleRect.value = contentRectToViewRect(entry)
  }
  else if (hoveredScheduleCellKey.value === key) {
    scheduleCloseHover(32)
  }
}

function handleSchedulePointerDownByEntry(nativeEvent, entry) {
  if (!entry)
    return
  clearScheduleCellPressTracking()
  clearHoverCloseTimer()
  openSchedulePopoverKey.value = ''

  if (typeof document !== 'undefined') {
    scheduleCellPressState.value = {
      key: schedulePopoverKeyByEntry(entry),
      startX: Number(nativeEvent?.clientX || 0),
      startY: Number(nativeEvent?.clientY || 0),
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

  props.handleSchedulePointerDown(
    buildDragEventPayload(nativeEvent, entry, nativeEvent?.currentTarget || hoverScheduleAnchorRef.value || null),
    entry.text,
    entry.contextColumn,
    entry.contextRecord,
  )
}

function conflictBadgeHit(viewX, viewY, entry) {
  if (!entry?.text?.scheduledConflict)
    return false
  const rect = contentRectToViewRect(entry)
  const innerX = rect.left + CELL_PADDING
  const innerY = rect.top + CELL_PADDING
  const innerWidth = Math.max(0, rect.width - CELL_PADDING * 2)
  const badgeWidth = 34
  const badgeX = innerX + innerWidth - badgeWidth
  return viewX >= badgeX && viewX <= badgeX + badgeWidth && viewY >= innerY && viewY <= innerY + BADGE_HEIGHT
}

function setCanvasCursor(cursor) {
  if (canvasRef.value instanceof HTMLCanvasElement)
    canvasRef.value.style.cursor = cursor
}

function handleCanvasMouseMove(event) {
  if (String(props.draggingScheduleCellKey || '').trim()) {
    setCanvasCursor('default')
    return
  }
  const point = resolveCanvasPoint(event?.clientX, event?.clientY)
  if (!point) {
    setCanvasCursor('default')
    scheduleCloseHover(24)
    return
  }

  const entry = resolveEntryFromViewPoint(point.x, point.y)
  if (!entry) {
    setCanvasCursor('default')
    scheduleCloseHover(24)
    return
  }

  if (props.hasScheduledLesson(entry.text)) {
    setHoveredScheduleEntry(entry)
    setCanvasCursor(props.isScheduleDraggable(entry.text) ? 'grab' : 'pointer')
    return
  }

  setCanvasCursor('pointer')
  scheduleCloseHover(24)
}

function handleCanvasLeave() {
  setCanvasCursor('default')
  scheduleCloseHover(32)
}

function handleCanvasMouseDown(event) {
  const point = resolveCanvasPoint(event?.clientX, event?.clientY)
  if (!point)
    return
  const entry = resolveEntryFromViewPoint(point.x, point.y)
  if (!entry || !props.hasScheduledLesson(entry.text))
    return
  handleSchedulePointerDownByEntry(event, entry)
}

function handleCanvasClick(event) {
  const point = resolveCanvasPoint(event?.clientX, event?.clientY)
  if (!point)
    return
  const entry = resolveEntryFromViewPoint(point.x, point.y)
  if (!entry)
    return

  if (props.hasScheduledLesson(entry.text)) {
    if (conflictBadgeHit(point.x, point.y, entry)) {
      props.openScheduledConflictDetail(entry.text)
      return
    }
    handleScheduleCellClickByEntry(entry)
    return
  }

  handleEmptyCellClickByEntry(entry)
}

function handleViewportScroll(event) {
  const currentTarget = event.currentTarget
  if (!(currentTarget instanceof HTMLElement))
    return
  scrollTop.value = currentTarget.scrollTop
  scrollLeft.value = currentTarget.scrollLeft
  updateHoveredScheduleRect()
  scheduleRender()
}

function installResizeObserver() {
  if (!(viewportRef.value instanceof HTMLElement) || typeof ResizeObserver === 'undefined')
    return
  resizeObserver?.disconnect()
  resizeObserver = new ResizeObserver(() => {
    updateViewportMetrics()
    updateHoveredScheduleRect()
    scheduleRender()
  })
  resizeObserver.observe(viewportRef.value)
}

watch(
  () => props.draggingScheduleCellKey,
  (value) => {
    if (String(value || '').trim())
      clearScheduleHover()
    scheduleRender()
  },
)

watch(
  () => [props.tableDataSource, props.columns, props.focusedScheduleCellKey],
  () => {
    updateViewportMetrics()
    updateHoveredScheduleRect()
    scheduleRender()
  },
  { deep: true },
)

watch(
  () => props.viewportHeight,
  () => {
    updateViewportMetrics()
    scheduleRender()
  },
)

watch(emptyDragCellStateMap, () => {
  scheduleRender()
}, { deep: true })

watch([scrollTop, scrollLeft], () => {
  updateHoveredScheduleRect()
})

onMounted(() => {
  updateViewportMetrics()
  installResizeObserver()
  if (viewportRef.value instanceof HTMLElement) {
    scrollTop.value = viewportRef.value.scrollTop
    scrollLeft.value = viewportRef.value.scrollLeft
  }
  scheduleRender()
  window.addEventListener('resize', updateViewportMetrics)
})

onUnmounted(() => {
  resizeObserver?.disconnect()
  resizeObserver = null
  clearHoverCloseTimer()
  clearScheduleCellPressTracking()
  if (pendingFrame)
    window.cancelAnimationFrame(pendingFrame)
  if (emptyDragCellStateFlushFrame)
    window.cancelAnimationFrame(emptyDragCellStateFlushFrame)
  window.removeEventListener('resize', updateViewportMetrics)
})

defineExpose({
  clearEmptyDragCellStates,
  getEmptyDragCellRect,
  getScheduleCellRect,
  getScheduleDragStartInfo,
  getViewportClientRect,
  resolvePointerDragTarget,
  scrollToScheduleCell,
  setEmptyDragCellState,
})
</script>

<template>
  <div ref="shellRef" class="st-canvas-grid">
    <div
      ref="viewportRef"
      class="st-canvas-grid__viewport"
      :style="{ height: `${viewportHeight}px` }"
      @scroll.passive="handleViewportScroll"
    >
      <div class="st-canvas-grid__sticky-layer">
        <canvas
          ref="canvasRef"
          class="st-canvas-grid__canvas"
          @mousemove="handleCanvasMouseMove"
          @mouseleave="handleCanvasLeave"
          @mousedown.left="handleCanvasMouseDown"
          @click="handleCanvasClick"
        />

        <div class="st-canvas-grid__overlay">
          <TimetableScheduleHoverPopover
            v-if="hoveredScheduleCellKey && hoveredScheduleRect && scheduleCellEntryMap.get(hoveredScheduleCellKey)"
            :open="!draggingScheduleCellKey && openSchedulePopoverKey === hoveredScheduleCellKey"
            :schedule-id="String(scheduleCellEntryMap.get(hoveredScheduleCellKey)?.text?.scheduleId || '')"
            :editable="Boolean(scheduleCellEntryMap.get(hoveredScheduleCellKey)?.text?.scheduleId) && !(scheduleCellEntryMap.get(hoveredScheduleCellKey)?.text?.courseType === 1 && scheduleCellEntryMap.get(hoveredScheduleCellKey)?.text?.isMain === false)"
            :batch-no="String(scheduleCellEntryMap.get(hoveredScheduleCellKey)?.text?.batchNo || '')"
            :batch-size="Number(scheduleCellEntryMap.get(hoveredScheduleCellKey)?.text?.batchSize || 0)"
            :lesson-date="String(scheduleCellEntryMap.get(hoveredScheduleCellKey)?.text?.lessonDate || '')"
            :call-status-key="String(scheduleCellEntryMap.get(hoveredScheduleCellKey)?.text?.callStatusKey || 'unsigned')"
            :mode-label="scheduleModeShortLabel(scheduleCellEntryMap.get(hoveredScheduleCellKey)?.text)"
            :lesson-title="scheduleLessonTitle(scheduleCellEntryMap.get(hoveredScheduleCellKey)?.text)"
            :teacher-name="scheduleCellEntryMap.get(hoveredScheduleCellKey)?.text?.teacherName || scheduleCellEntryMap.get(hoveredScheduleCellKey)?.record?.name || '-'"
            :course-name="scheduleLessonSubtitle(scheduleCellEntryMap.get(hoveredScheduleCellKey)?.text) || scheduleLessonTitle(scheduleCellEntryMap.get(hoveredScheduleCellKey)?.text)"
            :assistant-text="scheduleAssistantSummary(scheduleCellEntryMap.get(hoveredScheduleCellKey)?.text)"
            :student-text="scheduleStudentSummary(scheduleCellEntryMap.get(hoveredScheduleCellKey)?.text)"
            :time-text="scheduleHeaderTimeText(scheduleCellEntryMap.get(hoveredScheduleCellKey)?.column, scheduleCellEntryMap.get(hoveredScheduleCellKey)?.record)"
            :conflict-text="scheduleCellEntryMap.get(hoveredScheduleCellKey)?.text?.scheduledConflict ? scheduleConflictText(scheduleCellEntryMap.get(hoveredScheduleCellKey)?.text) : ''"
            @open-change="handleSchedulePopoverOpenChange(scheduleCellEntryMap.get(hoveredScheduleCellKey), $event)"
            @detail="openScheduledLessonDetail(scheduleCellEntryMap.get(hoveredScheduleCellKey)?.text, scheduleCellEntryMap.get(hoveredScheduleCellKey)?.contextColumn, scheduleCellEntryMap.get(hoveredScheduleCellKey)?.contextRecord)"
            @copy="payload => openScheduledLessonCopy(scheduleCellEntryMap.get(hoveredScheduleCellKey)?.text, scheduleCellEntryMap.get(hoveredScheduleCellKey)?.contextColumn, scheduleCellEntryMap.get(hoveredScheduleCellKey)?.contextRecord, payload)"
            @copy-current="payload => openScheduledLessonCopyCurrent(scheduleCellEntryMap.get(hoveredScheduleCellKey)?.text, scheduleCellEntryMap.get(hoveredScheduleCellKey)?.contextColumn, scheduleCellEntryMap.get(hoveredScheduleCellKey)?.contextRecord, payload)"
            @edit="payload => openScheduledLessonEdit(scheduleCellEntryMap.get(hoveredScheduleCellKey)?.text, scheduleCellEntryMap.get(hoveredScheduleCellKey)?.contextColumn, scheduleCellEntryMap.get(hoveredScheduleCellKey)?.contextRecord, payload)"
            @edit-current="payload => openScheduledLessonEditCurrent(scheduleCellEntryMap.get(hoveredScheduleCellKey)?.text, scheduleCellEntryMap.get(hoveredScheduleCellKey)?.contextColumn, scheduleCellEntryMap.get(hoveredScheduleCellKey)?.contextRecord, payload)"
          >
            <div
              ref="hoverScheduleAnchorRef"
              class="st-canvas-grid__hover-anchor"
              :data-schedule-cell-key="hoveredScheduleCellKey"
              :title="!isScheduleDraggable(scheduleCellEntryMap.get(hoveredScheduleCellKey)?.text) ? resolveScheduleDragBlockedMessage(scheduleCellEntryMap.get(hoveredScheduleCellKey)?.text) : undefined"
              :style="{
                left: `${hoveredScheduleRect.left}px`,
                top: `${hoveredScheduleRect.top}px`,
                width: `${hoveredScheduleRect.width}px`,
                height: `${hoveredScheduleRect.height}px`,
                cursor: isScheduleDraggable(scheduleCellEntryMap.get(hoveredScheduleCellKey)?.text) ? 'grab' : 'pointer',
              }"
              @mouseenter="setHoveredScheduleEntry(scheduleCellEntryMap.get(hoveredScheduleCellKey))"
              @mouseleave="scheduleCloseHover(240)"
              @click="handleScheduleCellClickByEntry(scheduleCellEntryMap.get(hoveredScheduleCellKey))"
              @mousedown.left="handleSchedulePointerDownByEntry($event, scheduleCellEntryMap.get(hoveredScheduleCellKey))"
            />
          </TimetableScheduleHoverPopover>
        </div>
      </div>

      <div
        class="st-canvas-grid__spacer"
        :style="{
          width: `${totalGridWidth}px`,
          height: `${totalGridHeight}px`,
        }"
      />
    </div>
  </div>
</template>

<style scoped lang="less">
.st-canvas-grid {
  position: relative;
  height: 100%;
  min-height: 0;
}

.st-canvas-grid__viewport {
  position: relative;
  overflow: auto;
  border: 1px solid #f0f0f0;
  background: #fff;
}

.st-canvas-grid__sticky-layer {
  position: sticky;
  top: 0;
  left: 0;
  z-index: 2;
  height: 0;
  overflow: visible;
}

.st-canvas-grid__canvas {
  display: block;
  pointer-events: auto;
}

.st-canvas-grid__sticky-layer,
.st-canvas-grid__overlay {
  width: 0;
}

.st-canvas-grid__spacer {
  position: relative;
  z-index: 1;
  pointer-events: none;
}

.st-canvas-grid__overlay {
  position: absolute;
  top: 0;
  left: 0;
  z-index: 3;
  height: 0;
  pointer-events: none;
}

.st-canvas-grid__hover-anchor {
  position: absolute;
  pointer-events: auto;
  background: transparent;
}

.st-canvas-grid__canvas {
  display: block;
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
</style>
