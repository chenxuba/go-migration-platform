import dayjs from 'dayjs'
import { getTeachingScheduleBatchDetailApi, type TeachingScheduleItem } from '@/api/edu-center/teaching-schedule'

export type TeachingScheduleDeleteScope = 'current' | 'future'

function normalizeTime(value?: string) {
  return dayjs(String(value || '')).format('HH:mm')
}

export function sortTeachingScheduleItemsByTimeline(list: TeachingScheduleItem[] = []) {
  return [...list].sort((left, right) => {
    const leftKey = `${String(left.lessonDate || '').trim()}|${normalizeTime(left.startAt)}|${normalizeTime(left.endAt)}|${String(left.id || '').trim()}`
    const rightKey = `${String(right.lessonDate || '').trim()}|${normalizeTime(right.startAt)}|${normalizeTime(right.endAt)}|${String(right.id || '').trim()}`
    return leftKey.localeCompare(rightKey)
  })
}

export function resolveTeachingScheduleDeleteTargets(list: TeachingScheduleItem[] = [], anchorId?: string, scope: TeachingScheduleDeleteScope = 'current') {
  const sorted = sortTeachingScheduleItemsByTimeline(list)
  const currentId = String(anchorId || '').trim()
  if (!currentId)
    return []
  const index = sorted.findIndex(item => String(item.id || '').trim() === currentId)
  if (index < 0)
    return []
  if (scope === 'future')
    return sorted.slice(index)
  return sorted.slice(index, index + 1)
}

export async function loadTeachingScheduleDeleteTargetCount(schedule?: Pick<TeachingScheduleItem, 'id' | 'batchNo'> | null, scope: TeachingScheduleDeleteScope = 'current') {
  const scheduleId = String(schedule?.id || '').trim()
  const batchNo = String(schedule?.batchNo || '').trim()
  if (!scheduleId)
    return 0
  if (scope === 'current' || !batchNo)
    return 1
  const res = await getTeachingScheduleBatchDetailApi({ batchNo })
  if (res.code !== 200 || !res.result)
    throw new Error(res.message || '加载批次规则失败')
  const targets = resolveTeachingScheduleDeleteTargets(res.result.schedules || [], scheduleId, scope)
  return Math.max(targets.length, 1)
}
