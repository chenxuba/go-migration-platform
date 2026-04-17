<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import TimetableScheduleHoverPopover from './timetable-schedule-hover-popover.vue'

interface ScheduleBoardColumn {
  key: string
  left: number
  width: number
  background?: string
  dividerWidth?: number
  dividerColor?: string
  showCurrentLine?: boolean
}

interface ScheduleBoardPopover {
  scheduleId?: string
  editable?: boolean
  batchNo?: string
  batchSize?: number
  lessonDate?: string
  callStatusKey?: string
  modeLabel?: string
  lessonTitle?: string
  teacherName?: string
  courseName?: string
  assistantText?: string
  studentText?: string
  classroomName?: string
  timeText?: string
  conflictText?: string
}

interface ScheduleBoardEvent {
  key: string
  id: string
  columnKey: string
  left: number
  top: number
  width: number
  height: number
  timeLabel: string
  title: string
  metaPrimary?: string
  metaSecondary?: string
  badgeText?: string
  badgeVariant?: 'group' | 'oneToOne' | 'conflict' | 'default'
  badgeTooltip?: string
  topBackground?: string
  bodyBackground?: string
  titleColor?: string
  metaPrimaryColor?: string
  metaSecondaryColor?: string
  focused?: boolean
  popover?: ScheduleBoardPopover
  raw?: any
}

interface TimeMark {
  key: string
  top: number
  label: string
  muted?: boolean
}

interface CurrentLine {
  top: number
  label: string
}

const props = withDefaults(defineProps<{
  columns?: ScheduleBoardColumn[]
  events?: ScheduleBoardEvent[]
  timeMarks?: TimeMark[]
  currentLine?: CurrentLine | null
  timeAxisWidth?: number
  timelineHeight?: number
  totalBodyWidth?: number
}>(), {
  columns: () => [],
  events: () => [],
  timeMarks: () => [],
  currentLine: null,
  timeAxisWidth: 84,
  timelineHeight: 0,
  totalBodyWidth: 0,
})

const emit = defineEmits<{
  (e: 'scroll', value: number): void
  (e: 'detail', event: ScheduleBoardEvent): void
  (e: 'copy', payload: { event: ScheduleBoardEvent, value?: any }): void
  (e: 'copy-current', payload: { event: ScheduleBoardEvent, value?: any }): void
  (e: 'edit', payload: { event: ScheduleBoardEvent, value?: any }): void
  (e: 'edit-current', payload: { event: ScheduleBoardEvent, value?: any }): void
  (e: 'conflict', event: ScheduleBoardEvent): void
}>()

const viewportRef = ref<HTMLElement | null>(null)
const canvasRef = ref<HTMLCanvasElement | null>(null)
const hoverAnchorRef = ref<HTMLElement | null>(null)

const scrollLeft = ref(0)
const hoveredEventKey = ref('')
const openPopoverKey = ref('')
const hoveredEventRect = ref<{ left: number, top: number, width: number, height: number } | null>(null)

let pendingFrame = 0
let resizeObserver: ResizeObserver | null = null
let hoverCloseTimer: number | null = null
const textMeasureCache = new Map<string, number>()
const ellipsisCache = new Map<string, string>()

const eventMap = computed(() => {
  const map = new Map<string, ScheduleBoardEvent>()
  props.events.forEach((item) => {
    const key = String(item.key || item.id || '').trim()
    if (key)
      map.set(key, item)
  })
  return map
})

const totalWidth = computed(() => props.timeAxisWidth + props.totalBodyWidth)

const overlayColumns = computed(() =>
  props.columns.map(column => ({
    ...column,
    viewLeft: props.timeAxisWidth + column.left - scrollLeft.value,
  })),
)

const searchableEvents = computed(() =>
  props.events
    .map((event, index) => ({
      event,
      index,
      left: Number(event.left || 0),
      right: Number(event.left || 0) + Number(event.width || 0),
    }))
    .sort((a, b) => a.left - b.left || a.index - b.index),
)

const maxEventWidth = computed(() =>
  props.events.reduce((max, event) => Math.max(max, Number(event.width || 0)), 0),
)

function clamp(value: number, min: number, max: number) {
  return Math.min(max, Math.max(min, value))
}

function clearHoverCloseTimer() {
  if (hoverCloseTimer != null)
    window.clearTimeout(hoverCloseTimer)
  hoverCloseTimer = null
}

function scheduleCloseHover(delay = 24) {
  clearHoverCloseTimer()
  hoverCloseTimer = window.setTimeout(() => {
    hoveredEventKey.value = ''
    hoveredEventRect.value = null
    openPopoverKey.value = ''
  }, delay)
}

function eventContentRect(event: ScheduleBoardEvent) {
  return {
    left: props.timeAxisWidth + event.left,
    top: event.top,
    width: event.width,
    height: event.height,
  }
}

function eventViewRect(event: ScheduleBoardEvent) {
  const rect = eventContentRect(event)
  return {
    left: rect.left - scrollLeft.value,
    top: rect.top,
    width: rect.width,
    height: rect.height,
  }
}

function updateHoveredEventRect() {
  const key = String(hoveredEventKey.value || '').trim()
  const event = key ? eventMap.value.get(key) : null
  hoveredEventRect.value = event ? eventViewRect(event) : null
}

function setHoveredEvent(event: ScheduleBoardEvent | null | undefined) {
  const key = String(event?.key || event?.id || '').trim()
  if (!key) {
    scheduleCloseHover()
    return
  }
  clearHoverCloseTimer()
  hoveredEventKey.value = key
  hoveredEventRect.value = eventViewRect(event as ScheduleBoardEvent)
  openPopoverKey.value = key
}

function currentHoveredEvent() {
  return eventMap.value.get(String(hoveredEventKey.value || '').trim()) || null
}

function emitDetailFromHover() {
  const event = currentHoveredEvent()
  if (event)
    emit('detail', event)
}

function emitActionFromHover(type: 'copy' | 'copy-current' | 'edit' | 'edit-current', value?: any) {
  const event = currentHoveredEvent()
  if (!event)
    return
  ;(emit as any)(type, { event, value })
}

function handlePopoverOpenChange(event: ScheduleBoardEvent | null | undefined, open: boolean) {
  clearHoverCloseTimer()
  const key = String(event?.key || event?.id || '').trim()
  if (!key)
    return
  if (open) {
    hoveredEventKey.value = key
    hoveredEventRect.value = eventViewRect(event as ScheduleBoardEvent)
    openPopoverKey.value = key
    return
  }
  if (hoveredEventKey.value === key)
    scheduleCloseHover(32)
}

function rememberCacheValue<T>(cache: Map<string, T>, key: string, value: T) {
  if (cache.size >= 4000)
    cache.clear()
  cache.set(key, value)
  return value
}

function measureTextWidth(ctx: CanvasRenderingContext2D, text: string) {
  const cacheKey = `${ctx.font}\n${text}`
  const cached = textMeasureCache.get(cacheKey)
  if (cached != null)
    return cached
  return rememberCacheValue(textMeasureCache, cacheKey, ctx.measureText(text).width)
}

function ellipsisText(ctx: CanvasRenderingContext2D, text: string, maxWidth: number) {
  const input = String(text || '')
  if (!input)
    return ''
  const safeMaxWidth = Math.max(0, Math.floor(maxWidth))
  if (!safeMaxWidth)
    return ''
  const cacheKey = `${ctx.font}\n${safeMaxWidth}\n${input}`
  const cached = ellipsisCache.get(cacheKey)
  if (cached != null)
    return cached
  if (measureTextWidth(ctx, input) <= safeMaxWidth)
    return input
  let output = input
  while (output.length > 1 && measureTextWidth(ctx, `${output}…`) > safeMaxWidth)
    output = output.slice(0, -1)
  return rememberCacheValue(ellipsisCache, cacheKey, `${output}…`)
}

function drawRoundedRect(ctx: CanvasRenderingContext2D, x: number, y: number, width: number, height: number, radius: number) {
  const safeRadius = Math.max(0, Math.min(radius, width / 2, height / 2))
  ctx.beginPath()
  ctx.moveTo(x + safeRadius, y)
  ctx.arcTo(x + width, y, x + width, y + height, safeRadius)
  ctx.arcTo(x + width, y + height, x, y + height, safeRadius)
  ctx.arcTo(x, y + height, x, y, safeRadius)
  ctx.arcTo(x, y, x + width, y, safeRadius)
  ctx.closePath()
}

function syncCanvasSize() {
  const canvas = canvasRef.value
  const viewport = viewportRef.value
  if (!canvas || !viewport)
    return null
  const ratio = typeof window !== 'undefined' ? Math.max(window.devicePixelRatio || 1, 1) : 1
  const cssWidth = viewport.clientWidth
  const cssHeight = Math.max(1, props.timelineHeight)
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

function visibleColumns(viewportWidth: number) {
  const bodyLeft = scrollLeft.value
  const bodyRight = scrollLeft.value + Math.max(0, viewportWidth - props.timeAxisWidth)
  return props.columns.filter((column) => {
    const right = column.left + column.width
    return right >= bodyLeft && column.left <= bodyRight
  })
}

function visibleEvents(viewportWidth: number) {
  const bodyLeft = scrollLeft.value
  const bodyRight = scrollLeft.value + Math.max(0, viewportWidth - props.timeAxisWidth)
  const searchable = searchableEvents.value
  if (!searchable.length)
    return []
  const searchStart = bodyLeft - maxEventWidth.value
  let low = 0
  let high = searchable.length
  while (low < high) {
    const middle = Math.floor((low + high) / 2)
    if (searchable[middle].left < searchStart)
      low = middle + 1
    else
      high = middle
  }
  const result: Array<{ event: ScheduleBoardEvent, index: number }> = []
  for (let index = Math.max(0, low - 1); index < searchable.length; index += 1) {
    const item = searchable[index]
    if (item.left > bodyRight)
      break
    if (item.right >= bodyLeft)
      result.push({ event: item.event, index: item.index })
  }
  result.sort((a, b) => a.index - b.index)
  return result.map(item => item.event)
}

function drawTimeAxis(ctx: CanvasRenderingContext2D, viewportWidth: number) {
  ctx.fillStyle = '#ffffff'
  ctx.fillRect(0, 0, props.timeAxisWidth, props.timelineHeight)
  ctx.strokeStyle = '#dde5f0'
  ctx.beginPath()
  ctx.moveTo(props.timeAxisWidth + 0.5, 0)
  ctx.lineTo(props.timeAxisWidth + 0.5, props.timelineHeight)
  ctx.stroke()

  props.timeMarks.forEach((mark) => {
    const y = mark.top

    ctx.strokeStyle = '#dde5f0'
    ctx.beginPath()
    ctx.moveTo(0, y + 0.5)
    ctx.lineTo(props.timeAxisWidth, y + 0.5)
    ctx.stroke()

    ctx.fillStyle = '#ffffff'
    ctx.fillRect(16, y - 10, Math.max(0, props.timeAxisWidth - 32), 20)
    ctx.fillStyle = mark.muted ? 'rgba(31, 41, 55, 0.22)' : '#1f2937'
    ctx.font = '600 14px sans-serif'
    ctx.textAlign = 'center'
    ctx.textBaseline = 'middle'
    ctx.fillText(mark.label, props.timeAxisWidth / 2, y)
  })

  if (props.currentLine) {
    ctx.fillStyle = '#ff4d4f'
    ctx.font = '600 12px sans-serif'
    ctx.textAlign = 'center'
    ctx.textBaseline = 'middle'
    ctx.fillText(props.currentLine.label, props.timeAxisWidth / 2, props.currentLine.top - 8)

    ctx.beginPath()
    ctx.arc(props.timeAxisWidth - 3, props.currentLine.top, 3, 0, Math.PI * 2)
    ctx.fill()
  }
}

function drawBodyHorizontalLines(ctx: CanvasRenderingContext2D, viewportWidth: number) {
  ctx.save()
  ctx.beginPath()
  ctx.rect(props.timeAxisWidth, 0, Math.max(0, viewportWidth - props.timeAxisWidth), props.timelineHeight)
  ctx.clip()

  props.timeMarks.forEach((mark) => {
    ctx.strokeStyle = '#dde5f0'
    ctx.beginPath()
    ctx.moveTo(props.timeAxisWidth, mark.top + 0.5)
    ctx.lineTo(viewportWidth, mark.top + 0.5)
    ctx.stroke()
  })

  ctx.restore()
}

function drawColumns(ctx: CanvasRenderingContext2D, viewportWidth: number) {
  const columns = visibleColumns(viewportWidth)
  const dividers: Array<{ x: number, width: number, color: string }> = []

  columns.forEach((column) => {
    const viewLeft = props.timeAxisWidth + column.left - scrollLeft.value
    ctx.fillStyle = column.background || '#ffffff'
    ctx.fillRect(viewLeft, 0, column.width, props.timelineHeight)

    dividers.push({
      x: viewLeft + column.width,
      width: Number(column.dividerWidth || 1),
      color: column.dividerColor || '#dde5f0',
    })

    if (props.currentLine?.top != null && column.showCurrentLine) {
      ctx.strokeStyle = '#ffb3b3'
      ctx.beginPath()
      ctx.moveTo(viewLeft, props.currentLine.top + 0.5)
      ctx.lineTo(viewLeft + column.width, props.currentLine.top + 0.5)
      ctx.stroke()
    }
  })

  dividers.forEach((divider) => {
    const lineWidth = Math.max(1, divider.width)
    ctx.fillStyle = divider.color
    ctx.fillRect(
      divider.x - lineWidth / 2,
      0,
      lineWidth,
      props.timelineHeight,
    )
  })
}

function eventBadgeWidth(ctx: CanvasRenderingContext2D, event: ScheduleBoardEvent) {
  const text = String(event.badgeText || '').trim()
  if (!text)
    return 0
  const baseWidth = Math.max(34, measureTextWidth(ctx, text) + 16)
  if (event.badgeVariant === 'group' || event.badgeVariant === 'oneToOne' || event.badgeVariant === 'conflict')
    return baseWidth
  return baseWidth
}

function eventBadgeFill(event: ScheduleBoardEvent) {
  if (event.badgeVariant === 'conflict')
    return '#ff4d4f'
  if (event.badgeVariant === 'group')
    return '#d46b08'
  if (event.badgeVariant === 'oneToOne')
    return 'rgba(0, 0, 0, 0.5)'
  return 'rgba(9, 61, 149, 0.24)'
}

function drawEvent(ctx: CanvasRenderingContext2D, event: ScheduleBoardEvent) {
  const viewLeft = props.timeAxisWidth + event.left - scrollLeft.value
  const viewTop = event.top
  const width = event.width
  const height = event.height
  const headerHeight = 24

  ctx.save()
  ctx.shadowColor = 'rgba(22, 119, 255, 0.1)'
  ctx.shadowBlur = 16
  ctx.shadowOffsetY = 6
  drawRoundedRect(ctx, viewLeft, viewTop, width, height, 4)
  ctx.fillStyle = event.bodyBackground || '#ffffff'
  ctx.fill()
  ctx.restore()

  if (event.focused) {
    drawRoundedRect(ctx, viewLeft - 1, viewTop - 1, width + 2, height + 2, 5)
    ctx.strokeStyle = 'rgba(22, 119, 255, 0.45)'
    ctx.lineWidth = 3
    ctx.stroke()
    ctx.lineWidth = 1
  }
  else if (event.badgeVariant === 'conflict') {
    drawRoundedRect(ctx, viewLeft, viewTop, width, height, 4)
    ctx.strokeStyle = 'rgba(255, 77, 79, 0.4)'
    ctx.lineWidth = 2
    ctx.stroke()
    ctx.lineWidth = 1
  }

  ctx.save()
  drawRoundedRect(ctx, viewLeft, viewTop, width, height, 4)
  ctx.clip()

  ctx.save()
  drawRoundedRect(ctx, viewLeft, viewTop, width, headerHeight, 4)
  ctx.clip()
  ctx.fillStyle = event.topBackground || '#1677ff'
  ctx.fillRect(viewLeft, viewTop, width, headerHeight + 2)
  ctx.restore()

  ctx.fillStyle = '#ffffff'
  ctx.font = '600 12px sans-serif'
  ctx.textAlign = 'left'
  ctx.textBaseline = 'middle'
  const badgeText = String(event.badgeText || '').trim()
  const badgeGap = badgeText ? 8 : 0
  const timeMaxWidth = Math.max(0, width - 20 - (badgeText ? eventBadgeWidth(ctx, event) : 0) - badgeGap)
  ctx.fillText(
    ellipsisText(ctx, event.timeLabel, timeMaxWidth),
    viewLeft + 10,
    viewTop + headerHeight / 2,
  )

  if (badgeText) {
    ctx.font = '700 10px sans-serif'
    const badgeWidth = eventBadgeWidth(ctx, event)
    const badgeX = viewLeft + width - badgeWidth
    const badgeY = viewTop
    drawRoundedRect(ctx, badgeX, badgeY, badgeWidth, 16, 4)
    ctx.fillStyle = eventBadgeFill(event)
    ctx.fill()

    ctx.fillStyle = '#ffffff'
    ctx.textAlign = 'center'
    ctx.textBaseline = 'middle'
    ctx.fillText(badgeText, badgeX + badgeWidth / 2, badgeY + 8)
  }

  ctx.textAlign = 'left'
  ctx.textBaseline = 'top'
  ctx.fillStyle = event.titleColor || '#0f172a'
  ctx.font = '700 13px sans-serif'
  ctx.fillText(
    ellipsisText(ctx, event.title, Math.max(0, width - 20)),
    viewLeft + 10,
    viewTop + 29,
  )

  if (event.metaPrimary) {
    ctx.fillStyle = event.metaPrimaryColor || '#334155'
    ctx.font = '12px sans-serif'
    ctx.fillText(
      ellipsisText(ctx, event.metaPrimary, Math.max(0, width - 20)),
      viewLeft + 10,
      viewTop + 49,
    )
  }

  if (event.metaSecondary) {
    ctx.fillStyle = event.metaSecondaryColor || '#64748b'
    ctx.font = '12px sans-serif'
    ctx.fillText(
      ellipsisText(ctx, event.metaSecondary, Math.max(0, width - 20)),
      viewLeft + 10,
      viewTop + 67,
    )
  }

  ctx.restore()
}

function renderFrame() {
  pendingFrame = 0
  const canvas = canvasRef.value
  const viewport = viewportRef.value
  const ctx = syncCanvasSize()
  if (!canvas || !viewport || !ctx)
    return

  const viewportWidth = viewport.clientWidth
  ctx.clearRect(0, 0, viewportWidth, props.timelineHeight)
  ctx.fillStyle = '#ffffff'
  ctx.fillRect(0, 0, viewportWidth, props.timelineHeight)

  ctx.save()
  ctx.beginPath()
  ctx.rect(props.timeAxisWidth, 0, Math.max(0, viewportWidth - props.timeAxisWidth), props.timelineHeight)
  ctx.clip()
  drawColumns(ctx, viewportWidth)
  drawBodyHorizontalLines(ctx, viewportWidth)
  visibleEvents(viewportWidth).forEach(event => drawEvent(ctx, event))
  ctx.restore()

  drawTimeAxis(ctx, viewportWidth)
}

function scheduleRender() {
  if (pendingFrame)
    return
  pendingFrame = window.requestAnimationFrame(renderFrame)
}

function resolveCanvasPoint(clientX: number, clientY: number) {
  const viewport = viewportRef.value
  if (!viewport)
    return null
  const rect = viewport.getBoundingClientRect()
  const x = clientX - rect.left
  const y = clientY - rect.top
  if (x < 0 || y < 0 || x > rect.width || y > rect.height)
    return null
  return { x, y }
}

function resolveEventFromPoint(clientX: number, clientY: number) {
  const point = resolveCanvasPoint(clientX, clientY)
  if (!point)
    return null
  for (let index = props.events.length - 1; index >= 0; index -= 1) {
    const event = props.events[index]
    const viewRect = eventViewRect(event)
    if (
      point.x >= viewRect.left
      && point.x <= viewRect.left + viewRect.width
      && point.y >= viewRect.top
      && point.y <= viewRect.top + viewRect.height
    ) {
      return {
        point,
        event,
      }
    }
  }
  return {
    point,
    event: null,
  }
}

function conflictBadgeHit(pointX: number, pointY: number, event: ScheduleBoardEvent) {
  if (event.badgeVariant !== 'conflict')
    return false
  const canvas = canvasRef.value
  const ctx = canvas?.getContext('2d')
  if (!ctx)
    return false
  ctx.save()
  ctx.font = '700 10px sans-serif'
  const badgeWidth = eventBadgeWidth(ctx, event)
  ctx.restore()
  const rect = eventViewRect(event)
  const badgeX = rect.left + rect.width - badgeWidth
  return pointX >= badgeX && pointX <= badgeX + badgeWidth && pointY >= rect.top && pointY <= rect.top + 16
}

function handleCanvasMouseMove(event: MouseEvent) {
  const resolved = resolveEventFromPoint(event.clientX, event.clientY)
  if (!resolved?.event) {
    scheduleCloseHover(24)
    if (canvasRef.value)
      canvasRef.value.style.cursor = 'default'
    return
  }

  setHoveredEvent(resolved.event)
  if (canvasRef.value)
    canvasRef.value.style.cursor = 'pointer'
}

function handleCanvasMouseLeave() {
  if (canvasRef.value)
    canvasRef.value.style.cursor = 'default'
  scheduleCloseHover(32)
}

function handleCanvasClick(event: MouseEvent) {
  const resolved = resolveEventFromPoint(event.clientX, event.clientY)
  if (!resolved?.event || !resolved.point)
    return
  if (conflictBadgeHit(resolved.point.x, resolved.point.y, resolved.event)) {
    emit('conflict', resolved.event)
    return
  }
  emit('detail', resolved.event)
}

function handleViewportScroll(event: Event) {
  const currentTarget = event.currentTarget as HTMLElement | null
  if (!currentTarget)
    return
  scrollLeft.value = currentTarget.scrollLeft
  updateHoveredEventRect()
  scheduleRender()
  emit('scroll', scrollLeft.value)
}

function setScrollLeft(nextLeft: number) {
  const viewport = viewportRef.value
  if (!viewport)
    return
  if (Math.abs(viewport.scrollLeft - nextLeft) < 1)
    return
  viewport.scrollLeft = nextLeft
  scrollLeft.value = viewport.scrollLeft
  updateHoveredEventRect()
  scheduleRender()
}

function getScrollLeft() {
  return scrollLeft.value
}

function getViewportWidth() {
  return Number(viewportRef.value?.clientWidth || 0)
}

async function scrollToEvent(id: string) {
  await nextTick()
  const viewport = viewportRef.value
  const target = props.events.find(item => String(item.id || '').trim() === String(id || '').trim())
  if (!viewport || !target)
    return false

  const bodyVisibleWidth = Math.max(0, viewport.clientWidth - props.timeAxisWidth)
  const desiredLeft = clamp(
    target.left - Math.max(0, bodyVisibleWidth - target.width) / 2,
    0,
    Math.max(0, totalWidth.value - viewport.clientWidth),
  )
  viewport.scrollTo({
    left: desiredLeft,
    behavior: 'smooth',
  })
  scrollLeft.value = desiredLeft
  updateHoveredEventRect()
  scheduleRender()
  return true
}

watch(
  () => [props.columns, props.events, props.timelineHeight, props.currentLine],
  () => {
    updateHoveredEventRect()
    scheduleRender()
  },
)

watch(scrollLeft, () => {
  updateHoveredEventRect()
})

onMounted(() => {
  scheduleRender()
  if (viewportRef.value && typeof ResizeObserver !== 'undefined') {
    resizeObserver = new ResizeObserver(() => {
      updateHoveredEventRect()
      scheduleRender()
    })
    resizeObserver.observe(viewportRef.value)
  }
})

onUnmounted(() => {
  resizeObserver?.disconnect()
  resizeObserver = null
  clearHoverCloseTimer()
  textMeasureCache.clear()
  ellipsisCache.clear()
  if (pendingFrame)
    window.cancelAnimationFrame(pendingFrame)
})

defineExpose({
  getScrollLeft,
  getViewportWidth,
  scrollToEvent,
  setScrollLeft,
})
</script>

<template>
  <div
    ref="viewportRef"
    class="chsb-viewport"
    @scroll.passive="handleViewportScroll"
  >
    <div class="chsb-sticky-layer">
      <canvas
        ref="canvasRef"
        class="chsb-canvas"
        @mousemove="handleCanvasMouseMove"
        @mouseleave="handleCanvasMouseLeave"
        @click="handleCanvasClick"
      />

      <div class="chsb-overlay">
        <div class="chsb-overlay-slot">
          <slot name="overlay" :columns="overlayColumns" />
        </div>

        <TimetableScheduleHoverPopover
          v-if="hoveredEventKey && hoveredEventRect && eventMap.get(hoveredEventKey)"
          :open="openPopoverKey === hoveredEventKey"
          :floating-gap="10"
          :floating-overlap="0"
          :schedule-id="String(eventMap.get(hoveredEventKey)?.popover?.scheduleId || '')"
          :editable="Boolean(eventMap.get(hoveredEventKey)?.popover?.editable)"
          :batch-no="String(eventMap.get(hoveredEventKey)?.popover?.batchNo || '')"
          :batch-size="Number(eventMap.get(hoveredEventKey)?.popover?.batchSize || 0)"
          :lesson-date="String(eventMap.get(hoveredEventKey)?.popover?.lessonDate || '')"
          :call-status-key="String(eventMap.get(hoveredEventKey)?.popover?.callStatusKey || 'unsigned')"
          :mode-label="eventMap.get(hoveredEventKey)?.popover?.modeLabel || ''"
          :lesson-title="eventMap.get(hoveredEventKey)?.popover?.lessonTitle || ''"
          :teacher-name="eventMap.get(hoveredEventKey)?.popover?.teacherName || ''"
          :course-name="eventMap.get(hoveredEventKey)?.popover?.courseName || ''"
          :assistant-text="eventMap.get(hoveredEventKey)?.popover?.assistantText || ''"
          :student-text="eventMap.get(hoveredEventKey)?.popover?.studentText || ''"
          :classroom-name="eventMap.get(hoveredEventKey)?.popover?.classroomName || ''"
          :time-text="eventMap.get(hoveredEventKey)?.popover?.timeText || ''"
          :conflict-text="eventMap.get(hoveredEventKey)?.popover?.conflictText || ''"
          @open-change="handlePopoverOpenChange(eventMap.get(hoveredEventKey), $event)"
          @detail="emitDetailFromHover"
          @copy="emitActionFromHover('copy', $event)"
          @copy-current="emitActionFromHover('copy-current', $event)"
          @edit="emitActionFromHover('edit', $event)"
          @edit-current="emitActionFromHover('edit-current', $event)"
        >
          <div
            ref="hoverAnchorRef"
            class="chsb-hover-anchor"
            :style="{
              left: `${hoveredEventRect.left}px`,
              top: `${hoveredEventRect.top}px`,
              width: `${hoveredEventRect.width}px`,
              height: `${hoveredEventRect.height}px`,
            }"
            @mouseenter="setHoveredEvent(eventMap.get(hoveredEventKey))"
            @mouseleave="scheduleCloseHover(240)"
            @click="emitDetailFromHover"
          />
        </TimetableScheduleHoverPopover>
      </div>
    </div>

    <div
      class="chsb-spacer"
      :style="{
        width: `${totalWidth}px`,
        height: `${timelineHeight}px`,
      }"
    />
  </div>
</template>

<style scoped lang="less">
.chsb-viewport {
  position: relative;
  overflow-x: auto;
  overflow-y: visible;
  background: #fff;
}

.chsb-sticky-layer {
  position: sticky;
  top: 0;
  left: 0;
  z-index: 2;
  height: 0;
  overflow: visible;
}

.chsb-canvas {
  display: block;
}

.chsb-overlay {
  position: absolute;
  top: 0;
  left: 0;
  z-index: 3;
  width: 0;
  height: 0;
  pointer-events: none;
}

.chsb-overlay-slot {
  position: relative;
  pointer-events: none;
}

.chsb-hover-anchor {
  position: absolute;
  background: transparent;
  pointer-events: auto;
  cursor: pointer;
}

.chsb-spacer {
  position: relative;
  z-index: 1;
  pointer-events: none;
}
</style>
