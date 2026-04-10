import dayjs from 'dayjs'
import type { TeachingScheduleBatchDetail, TeachingScheduleBatchMeta, TeachingScheduleItem } from '@/api/edu-center/teaching-schedule'

export type GroupClassBatchPlanScheduleType = 'groupClass'
export type GroupClassBatchPlanSchedulingMode = 'repeat' | 'free'
export type GroupClassBatchPlanRepeatRule = 'none' | 'weekly' | 'biweekly' | 'daily' | 'alternateDay'
export type GroupClassBatchPlanHolidayPolicy = 'include' | 'filter'

export interface GroupClassBatchPlanTimeBlockPreset {
  startTime: string
  endTime: string
}

export interface GroupClassBatchPlanModalPreset {
  batchNo?: string
  scheduleIds: string[]
  groupClassId: string
  scheduleType: GroupClassBatchPlanScheduleType
  schedulingMode: GroupClassBatchPlanSchedulingMode
  repeatRule: GroupClassBatchPlanRepeatRule
  holidayPolicy: GroupClassBatchPlanHolidayPolicy
  selectedWeekdays: string[]
  scheduleStartDate: string
  freeSelectedDates: string[]
  plannedClassCount: number
  teacherId?: string
  assistantIds: string[]
  classroomId?: string
  timeBlocks: GroupClassBatchPlanTimeBlockPreset[]
  warnings: string[]
  editable: boolean
  detail: TeachingScheduleBatchDetail
}

const weekdayLabels = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
const weekdayToNumber: Record<string, number> = {
  周日: 0,
  周一: 1,
  周二: 2,
  周三: 3,
  周四: 4,
  周五: 5,
  周六: 6,
}

function normalizeId(value: unknown) {
  const text = String(value ?? '').trim()
  if (!text || text === '0' || text === 'undefined' || text === 'null')
    return undefined
  return text
}

function normalizeTime(value: string) {
  return dayjs(value).format('HH:mm')
}

function sortSchedules(list: TeachingScheduleItem[]) {
  return [...list].sort((a, b) => {
    const left = `${a.lessonDate}|${normalizeTime(a.startAt)}|${normalizeTime(a.endAt)}|${a.id}`
    const right = `${b.lessonDate}|${normalizeTime(b.startAt)}|${normalizeTime(b.endAt)}|${b.id}`
    return left.localeCompare(right)
  })
}

function sliceSchedulesFromAnchor(list: TeachingScheduleItem[], anchorScheduleId?: string) {
  const sorted = sortSchedules(list)
  const anchorId = String(anchorScheduleId || '').trim()
  if (!anchorId)
    return sorted
  const index = sorted.findIndex(item => String(item.id || '').trim() === anchorId)
  if (index < 0)
    return sorted
  return sorted.slice(index)
}

function uniqueDates(list: TeachingScheduleItem[]) {
  return [...new Set(sortSchedules(list).map(item => item.lessonDate))]
}

function uniqueTimeBlocks(list: TeachingScheduleItem[]): GroupClassBatchPlanTimeBlockPreset[] {
  const seen = new Set<string>()
  const result: GroupClassBatchPlanTimeBlockPreset[] = []
  sortSchedules(list).forEach((item) => {
    const startTime = normalizeTime(item.startAt)
    const endTime = normalizeTime(item.endAt)
    const key = `${startTime}|${endTime}`
    if (seen.has(key))
      return
    seen.add(key)
    result.push({ startTime, endTime })
  })
  return result
}

function actualPairs(list: TeachingScheduleItem[]) {
  return sortSchedules(list).map(item => `${item.lessonDate}|${normalizeTime(item.startAt)}|${normalizeTime(item.endAt)}`)
}

function arraysEqual(left: string[], right: string[]) {
  if (left.length !== right.length)
    return false
  return left.every((item, index) => item === right[index])
}

function pickConsistentValue(values: Array<string | undefined>) {
  const normalized = values.map(item => normalizeId(item)).filter(Boolean) as string[]
  if (!normalized.length)
    return undefined
  const first = normalized[0]
  return normalized.every(item => item === first) ? first : undefined
}

function pickConsistentAssistantIds(list: TeachingScheduleItem[]) {
  const normalized = list.map(item =>
    (Array.isArray(item.assistantIds) ? item.assistantIds : [])
      .map(value => normalizeId(value))
      .filter(Boolean)
      .sort() as string[],
  )
  if (!normalized.length)
    return []
  const first = normalized[0]
  return normalized.every(item => arraysEqual(item, first)) ? [...first] : []
}

function weekdayLabelsFromDates(dates: string[]) {
  const seen = new Set<string>()
  const labels: string[] = []
  dates.forEach((item) => {
    const label = weekdayLabels[dayjs(item).day()]
    if (!label || seen.has(label))
      return
    seen.add(label)
    labels.push(label)
  })
  return labels
}

function buildRepeatPairs(options: {
  startDate: string
  plannedClassCount: number
  repeatRule: GroupClassBatchPlanRepeatRule
  selectedWeekdays: string[]
  timeBlocks: GroupClassBatchPlanTimeBlockPreset[]
}) {
  const start = dayjs(options.startDate).startOf('day')
  if (!start.isValid())
    return []
  const slotCount = Math.max(options.timeBlocks.length, 1)
  const planned = Math.max(0, Math.floor(Number(options.plannedClassCount) || 0))
  if (planned < 1)
    return []

  let dates = [start]
  if (options.repeatRule !== 'none') {
    dates = []
    const targetDates = Math.max(1, Math.ceil(planned / slotCount))
    let cursor = start
    let guard = 0
    const selectedNumbers = new Set(options.selectedWeekdays.map(item => weekdayToNumber[item]))
    while (dates.length < targetDates && guard < 5000) {
      if (options.repeatRule === 'daily') {
        dates.push(cursor)
      }
      else if (options.repeatRule === 'alternateDay') {
        if (cursor.diff(start, 'day') % 2 === 0)
          dates.push(cursor)
      }
      else {
        const matchedWeekday = selectedNumbers.has(cursor.day())
        const weekDiff = Math.floor(cursor.startOf('day').diff(start, 'day') / 7)
        if (matchedWeekday && (options.repeatRule === 'weekly' || weekDiff % 2 === 0))
          dates.push(cursor)
      }
      cursor = cursor.add(1, 'day')
      guard += 1
    }
  }

  const pairs = dates.flatMap(date =>
    options.timeBlocks.map(block => `${date.format('YYYY-MM-DD')}|${block.startTime}|${block.endTime}`),
  )
  const cap = options.repeatRule === 'none'
    ? Math.min(planned, Math.max(options.timeBlocks.length, 1))
    : planned
  return pairs.slice(0, cap)
}

function buildFreePairs(dates: string[], timeBlocks: GroupClassBatchPlanTimeBlockPreset[]) {
  return dates.flatMap(date =>
    timeBlocks.map(block => `${date}|${block.startTime}|${block.endTime}`),
  )
}

function normalizeStoredBatchMeta(meta?: TeachingScheduleBatchMeta) {
  if (!meta)
    return null
  const schedulingMode = meta.schedulingMode === 'free' ? 'free' : 'repeat'
  const repeatRule = (meta.repeatRule || 'none') as GroupClassBatchPlanRepeatRule
  return {
    schedulingMode,
    repeatRule,
    holidayPolicy: meta.holidayPolicy === 'filter' ? 'filter' as const : 'include' as const,
    selectedWeekdays: Array.isArray(meta.selectedWeekdays) ? meta.selectedWeekdays.filter(Boolean) : [],
    scheduleStartDate: String(meta.scheduleStartDate || '').trim(),
    freeSelectedDates: Array.isArray(meta.freeSelectedDates) ? meta.freeSelectedDates.filter(Boolean) : [],
    plannedClassCount: Math.max(0, Number(meta.plannedClassCount || 0)),
  }
}

function inferModeAndRule(list: TeachingScheduleItem[], timeBlocks: GroupClassBatchPlanTimeBlockPreset[]) {
  const dates = uniqueDates(list)
  const pairs = actualPairs(list)
  const startDate = dates[0] || dayjs().format('YYYY-MM-DD')
  const selectedWeekdays = weekdayLabelsFromDates(dates)
  const plannedClassCount = pairs.length

  if (dates.length <= 1) {
    return {
      editable: true,
      schedulingMode: 'free' as const,
      repeatRule: 'none' as const,
      selectedWeekdays,
      scheduleStartDate: startDate,
      freeSelectedDates: dates.length ? dates : [startDate],
      plannedClassCount,
    }
  }

  const repeatCandidates: GroupClassBatchPlanRepeatRule[] = ['daily', 'alternateDay', 'weekly', 'biweekly', 'none']
  for (const repeatRule of repeatCandidates) {
    if (repeatRule === 'none' && dates.length !== 1)
      continue
    if ((repeatRule === 'weekly' || repeatRule === 'biweekly') && selectedWeekdays.length === 0)
      continue
    const generated = buildRepeatPairs({
      startDate,
      plannedClassCount,
      repeatRule,
      selectedWeekdays,
      timeBlocks,
    })
    if (arraysEqual(generated, pairs)) {
      return {
        editable: true,
        schedulingMode: 'repeat' as const,
        repeatRule,
        selectedWeekdays,
        scheduleStartDate: startDate,
        freeSelectedDates: dates,
        plannedClassCount,
      }
    }
  }

  const freePairs = buildFreePairs(dates, timeBlocks)
  if (arraysEqual(freePairs, pairs)) {
    return {
      editable: true,
      schedulingMode: 'free' as const,
      repeatRule: 'none' as const,
      selectedWeekdays,
      scheduleStartDate: startDate,
      freeSelectedDates: dates,
      plannedClassCount,
    }
  }

  return {
    editable: false,
    schedulingMode: 'free' as const,
    repeatRule: 'none' as const,
    selectedWeekdays,
    scheduleStartDate: startDate,
    freeSelectedDates: dates,
    plannedClassCount,
  }
}

export function inferGroupClassBatchPlanPreset(detail: TeachingScheduleBatchDetail, anchorScheduleId?: string): GroupClassBatchPlanModalPreset {
  const fullList = Array.isArray(detail.schedules) ? sortSchedules(detail.schedules) : []
  const list = sliceSchedulesFromAnchor(fullList, anchorScheduleId)
  const warnings: string[] = []
  const timeBlocks = uniqueTimeBlocks(list)
  const fullTimeBlocks = uniqueTimeBlocks(fullList)
  const teacherId = pickConsistentValue(list.map(item => item.teacherId))
  const classroomId = pickConsistentValue(list.map(item => item.classroomId))
  const assistantIds = pickConsistentAssistantIds(list)

  if (!teacherId)
    warnings.push('当前批次的上课教师已被逐节调整，打开后请重新确认主教。')
  if (!classroomId && list.some(item => normalizeId(item.classroomId)))
    warnings.push('当前批次的上课教室不完全一致，打开后请重新确认教室。')
  if (!assistantIds.length && list.some(item => Array.isArray(item.assistantIds) && item.assistantIds.length > 0))
    warnings.push('当前批次的上课助教不完全一致，打开后请重新确认助教。')

  const storedMeta = normalizeStoredBatchMeta(detail.batchMeta)
  const remainingDates = uniqueDates(list)
  const inferredBase = inferModeAndRule(fullList, fullTimeBlocks)

  if (storedMeta) {
    return {
      batchNo: detail.batchNo,
      scheduleIds: list.map(item => item.id),
      groupClassId: detail.teachingClassId,
      scheduleType: 'groupClass',
      schedulingMode: storedMeta.schedulingMode === 'free' ? 'free' : 'repeat',
      repeatRule: (storedMeta.repeatRule as GroupClassBatchPlanRepeatRule) || 'none',
      holidayPolicy: storedMeta.holidayPolicy === 'filter' ? 'filter' : 'include',
      selectedWeekdays: Array.isArray(storedMeta.selectedWeekdays) && storedMeta.selectedWeekdays.length
        ? [...storedMeta.selectedWeekdays]
        : weekdayLabelsFromDates(remainingDates),
      scheduleStartDate: remainingDates[0] || storedMeta.scheduleStartDate || dayjs().format('YYYY-MM-DD'),
      freeSelectedDates: remainingDates.length
        ? remainingDates
        : (Array.isArray(storedMeta.freeSelectedDates) ? storedMeta.freeSelectedDates : []),
      plannedClassCount: Math.max(1, list.length || Number(storedMeta.plannedClassCount || 1)),
      teacherId,
      assistantIds,
      classroomId,
      timeBlocks,
      warnings,
      editable: true,
      detail,
    }
  }

  if (!inferredBase.editable)
    warnings.push('当前批次的日期与时段组合已被逐节调整，暂不能完整按规则回显；如需整体重排，请先统一这批日程的时间结构。')

  return {
    batchNo: detail.batchNo,
    scheduleIds: list.map(item => item.id),
    groupClassId: detail.teachingClassId,
    scheduleType: 'groupClass',
    schedulingMode: inferredBase.schedulingMode,
    repeatRule: inferredBase.repeatRule,
    holidayPolicy: 'include',
    selectedWeekdays: inferredBase.selectedWeekdays,
    scheduleStartDate: remainingDates[0] || inferredBase.scheduleStartDate,
    freeSelectedDates: remainingDates.length ? remainingDates : inferredBase.freeSelectedDates,
    plannedClassCount: Math.max(1, list.length || inferredBase.plannedClassCount),
    teacherId,
    assistantIds,
    classroomId,
    timeBlocks,
    warnings,
    editable: inferredBase.editable,
    detail,
  }
}
