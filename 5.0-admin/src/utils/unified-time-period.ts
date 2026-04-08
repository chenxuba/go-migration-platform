/** 机构统一上课时间段（与 inst-config 中 unifiedTimePeriodJson 字段结构一致） */

export interface UnifiedPeriodSlot {
  index: number
  start: string
  end: string
  enabled?: boolean
}

/** 时段组关联的老师（仅存机构用户 id + 展示名）；不同时段组可重复绑定同一老师 */
export interface UnifiedPeriodBoundTeacher {
  id: string
  name: string
}

export interface UnifiedPeriodGroup {
  id: string
  name: string
  sort: number
  slots: UnifiedPeriodSlot[]
  boundTeachers?: UnifiedPeriodBoundTeacher[]
}

export interface UnifiedTimePeriodConfig {
  version: number
  groups: UnifiedPeriodGroup[]
}

/** 午休从当天 12:00 开始，时长为 lunchBreakMinutes；为 0 则不插入午休空档 */
export interface SmartFillSlotParams {
  firstStart: string
  lessonMinutes: number
  breakBetweenMinutes: number
  lunchBreakMinutes: number
  maxSlots?: number
}

const LUNCH_START_MINUTES = 12 * 60

function hhmmToMinutes(hhmm: string): number | null {
  const t = String(hhmm || '').trim()
  const m = /^(\d{1,2}):(\d{2})$/.exec(t)
  if (!m)
    return null
  const h = Number(m[1])
  const mi = Number(m[2])
  if (!Number.isFinite(h) || !Number.isFinite(mi) || h < 0 || h > 23 || mi < 0 || mi > 59)
    return null
  return h * 60 + mi
}

function minutesToHHmm(total: number): string {
  const capped = Math.min(Math.max(0, total), 24 * 60 - 1)
  const h = Math.floor(capped / 60)
  const mi = capped % 60
  return `${String(h).padStart(2, '0')}:${String(mi).padStart(2, '0')}`
}

/**
 * 按「首节开始 + 课长 + 课间 +（12:00 起的）午休」自动生成节次列表。
 * 若某一节会跨过 12:00，则该节结束时间卡在 12:00，午休后再从午休结束起排课。
 */
export function generateSlotsSmartFill(p: SmartFillSlotParams): UnifiedPeriodSlot[] {
  const lesson = Math.max(5, Math.min(180, Math.round(Number(p.lessonMinutes) || 40)))
  const brk = Math.max(0, Math.min(120, Math.round(Number(p.breakBetweenMinutes) || 0)))
  const lunchLen = Math.max(0, Math.min(240, Math.round(Number(p.lunchBreakMinutes) || 0)))
  const maxSlots = Math.max(1, Math.min(32, p.maxSlots ?? 16))
  let cur = hhmmToMinutes(p.firstStart)
  if (cur == null)
    return []

  const lunchEnd = lunchLen > 0 ? LUNCH_START_MINUTES + lunchLen : -1
  const out: UnifiedPeriodSlot[] = []

  const skipIntoLunch = () => {
    if (lunchLen <= 0)
      return
    if (cur >= LUNCH_START_MINUTES && cur < lunchEnd)
      cur = lunchEnd
  }

  while (out.length < maxSlots && cur < 24 * 60) {
    skipIntoLunch()

    let periodEnd = cur + lesson
    if (periodEnd > 24 * 60)
      break

    if (lunchLen > 0 && cur < LUNCH_START_MINUTES && periodEnd > LUNCH_START_MINUTES) {
      periodEnd = LUNCH_START_MINUTES
      if (periodEnd <= cur) {
        cur = lunchEnd
        continue
      }
    }

    out.push({
      index: out.length + 1,
      start: minutesToHHmm(cur),
      end: minutesToHHmm(periodEnd),
      enabled: true,
    })

    cur = periodEnd + brk
    if (lunchLen > 0 && cur > LUNCH_START_MINUTES && cur < lunchEnd)
      cur = lunchEnd
  }

  return out.map((s, i) => ({ ...s, index: i + 1 }))
}

/** 从 8:00 起连续 12 节整点（8–9 … 19–20），供默认配置与兼容旧「快捷生成」 */
export function buildQuickHourlySlots(): UnifiedPeriodSlot[] {
  return generateSlotsSmartFill({
    firstStart: '08:00',
    lessonMinutes: 60,
    breakBetweenMinutes: 0,
    lunchBreakMinutes: 0,
    maxSlots: 12,
  })
}

export const DEFAULT_UNIFIED_TIME_PERIOD_CONFIG: UnifiedTimePeriodConfig = {
  version: 1,
  groups: [
    { id: 'group-a', name: 'A时段', sort: 0, slots: buildQuickHourlySlots(), boundTeachers: [] },
    { id: 'group-b', name: 'B时段', sort: 1, slots: buildQuickHourlySlots(), boundTeachers: [] },
  ],
}

export function parseUnifiedTimePeriodConfig(raw: unknown): UnifiedTimePeriodConfig | null {
  if (raw == null || raw === '')
    return null
  try {
    let obj: unknown = raw
    if (typeof raw === 'string') {
      const t = raw.trim()
      if (!t)
        return null
      obj = JSON.parse(t) as unknown
    }
    if (typeof obj !== 'object' || obj === null || !Array.isArray((obj as UnifiedTimePeriodConfig).groups))
      return null
    const c = obj as UnifiedTimePeriodConfig
    return {
      version: Number(c.version) || 1,
      groups: c.groups.map((g, gi) => normalizeGroup(g, gi)),
    }
  }
  catch {
    return null
  }
}

function normalizeBoundTeachers(raw: unknown): UnifiedPeriodBoundTeacher[] {
  if (!Array.isArray(raw))
    return []
  const out: UnifiedPeriodBoundTeacher[] = []
  for (const item of raw) {
    if (typeof item !== 'object' || item === null)
      continue
    const o = item as Record<string, unknown>
    const id = String(o.id ?? o.teacherId ?? '').trim()
    if (!id)
      continue
    const name = String(o.name ?? o.teacherName ?? o.nickName ?? '').trim() || id
    out.push({ id, name })
  }
  return out
}

function normalizeGroup(g: Partial<UnifiedPeriodGroup>, fallbackIndex: number): UnifiedPeriodGroup {
  const id = String(g.id || `group-${fallbackIndex}`).trim() || `group-${fallbackIndex}`
  const name = String(g.name || `时段${fallbackIndex + 1}`).trim()
  const sort = typeof g.sort === 'number' ? g.sort : fallbackIndex
  const slots = Array.isArray(g.slots)
    ? g.slots.map((s, si) => normalizeSlot(s, si)).filter(s => s.start && s.end)
    : []
  const boundTeachers = normalizeBoundTeachers((g as UnifiedPeriodGroup).boundTeachers)
  return { id, name, sort, slots, boundTeachers }
}

function normalizeSlot(s: Partial<UnifiedPeriodSlot>, si: number): UnifiedPeriodSlot {
  const index = typeof s.index === 'number' && s.index > 0 ? s.index : si + 1
  return {
    index,
    start: String(s.start || '').slice(0, 5),
    end: String(s.end || '').slice(0, 5),
    enabled: s.enabled !== false,
  }
}

export function configGroupsSorted(config: UnifiedTimePeriodConfig | null): UnifiedPeriodGroup[] {
  if (!config?.groups?.length)
    return []
  return [...config.groups].sort((a, b) => a.sort - b.sort)
}

export function periodGroupKeyForIndex(index: number): string {
  const safeIndex = Number.isFinite(index) ? Math.max(0, Math.floor(index)) : 0
  if (safeIndex < 26)
    return String.fromCharCode(65 + safeIndex)
  return `G${safeIndex + 1}`
}

export function periodGroupIndexForKey(key: string): number {
  const normalized = String(key || '').trim().toUpperCase()
  if (/^[A-Z]$/.test(normalized))
    return normalized.charCodeAt(0) - 65
  const match = normalized.match(/^G(\d+)$/)
  if (match)
    return Math.max(0, Number(match[1]) - 1)
  return 0
}

export function slotCountActive(g: UnifiedPeriodGroup): number {
  return g.slots.filter(s => s.enabled !== false).length
}
