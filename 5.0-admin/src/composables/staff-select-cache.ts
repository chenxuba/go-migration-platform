export interface SharedStaffOption {
  id: number | string
  nickName: string
  mobile?: string
  status?: number
}

interface SharedStaffCacheEntry {
  items: SharedStaffOption[]
  total: number
  updatedAt: number
  promise?: Promise<{ items: SharedStaffOption[], total: number }>
}

const STAFF_CACHE_TTL = 5 * 60 * 1000
const sharedStaffCache = new Map<string, SharedStaffCacheEntry>()

function buildStaffCacheKey(fetchType: string, status?: number) {
  return `${fetchType}:${status ?? 'all'}`
}

export async function getCachedInitialStaffList(
  fetchType: string,
  status: number | undefined,
  loader: () => Promise<{ items: SharedStaffOption[], total: number }>,
) {
  const key = buildStaffCacheKey(fetchType, status)
  const current = sharedStaffCache.get(key)
  const now = Date.now()

  if (current?.promise) {
    return current.promise
  }

  if (current && current.items.length > 0 && now - current.updatedAt < STAFF_CACHE_TTL) {
    return {
      items: current.items,
      total: current.total,
    }
  }

  const promise = loader()
    .then((result) => {
      sharedStaffCache.set(key, {
        items: result.items,
        total: result.total,
        updatedAt: Date.now(),
      })
      return result
    })
    .catch((error) => {
      sharedStaffCache.delete(key)
      throw error
    })

  sharedStaffCache.set(key, {
    items: current?.items || [],
    total: current?.total || 0,
    updatedAt: current?.updatedAt || 0,
    promise,
  })

  return promise
}

export function findCachedStaff(fetchType: string, status: number | undefined, staffId: number | string) {
  const entry = sharedStaffCache.get(buildStaffCacheKey(fetchType, status))
  if (!entry)
    return null
  return entry.items.find(item => `${item.id}` === `${staffId}`) || null
}

export function mergeCachedStaff(fetchType: string, status: number | undefined, items: SharedStaffOption[], total?: number) {
  const key = buildStaffCacheKey(fetchType, status)
  const current = sharedStaffCache.get(key)
  const merged = [...(current?.items || [])]

  items.forEach((item) => {
    if (!merged.find(existing => `${existing.id}` === `${item.id}`)) {
      merged.push(item)
    }
  })

  sharedStaffCache.set(key, {
    items: merged,
    total: total ?? current?.total ?? merged.length,
    updatedAt: Date.now(),
  })
}
