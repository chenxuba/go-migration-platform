/** 机构统一上课时间段 JSON，存 inst_config.unified_time_period_json */

export interface UnifiedPeriodSlot {
  index: number
  start: string
  end: string
  enabled?: boolean
}

export interface UnifiedPeriodGroup {
  id: string
  name: string
  sort: number
  slots: UnifiedPeriodSlot[]
}

export interface UnifiedTimePeriodConfig {
  version: number
  groups: UnifiedPeriodGroup[]
}

/** 从 8:00 起连续 12 节整点（8–9 … 19–20），供默认配置与设置页「快捷生成」 */
export function buildQuickHourlySlots(): UnifiedPeriodSlot[] {
  return Array.from({ length: 12 }, (_, i) => {
    const h = 8 + i
    const sh = String(h).padStart(2, '0')
    const eh = String(h + 1).padStart(2, '0')
    return { index: i + 1, start: `${sh}:00`, end: `${eh}:00`, enabled: true }
  })
}

export const DEFAULT_UNIFIED_TIME_PERIOD_CONFIG: UnifiedTimePeriodConfig = {
  version: 1,
  groups: [
    { id: 'group-a', name: 'A时段', sort: 0, slots: buildQuickHourlySlots() },
    { id: 'group-b', name: 'B时段', sort: 1, slots: buildQuickHourlySlots() },
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

function normalizeGroup(g: Partial<UnifiedPeriodGroup>, fallbackIndex: number): UnifiedPeriodGroup {
  const id = String(g.id || `group-${fallbackIndex}`).trim() || `group-${fallbackIndex}`
  const name = String(g.name || `时段${fallbackIndex + 1}`).trim()
  const sort = typeof g.sort === 'number' ? g.sort : fallbackIndex
  const slots = Array.isArray(g.slots)
    ? g.slots.map((s, si) => normalizeSlot(s, si)).filter(s => s.start && s.end)
    : []
  return { id, name, sort, slots }
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

export function slotCountActive(g: UnifiedPeriodGroup): number {
  return g.slots.filter(s => s.enabled !== false).length
}
