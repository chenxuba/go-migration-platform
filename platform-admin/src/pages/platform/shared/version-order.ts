export const systemDefaultVersionNames = ['体验版', '基础版', '高级版', '旗舰版'] as const

const versionDisplayOrder = [...systemDefaultVersionNames]

const versionDisplayOrderMap = new Map<string, number>(
  versionDisplayOrder.map((name, index) => [name, index] as const),
)

export function getVersionDisplayRank(name?: string) {
  const normalizedName = String(name || '').trim()
  return versionDisplayOrderMap.get(normalizedName) ?? Number.MAX_SAFE_INTEGER
}

export function sortVersionsByDisplayOrder<T extends { name?: string }>(items: T[]) {
  return items
    .map((item, index) => ({ item, index }))
    .sort((left, right) => {
      const rankDiff = getVersionDisplayRank(left.item.name) - getVersionDisplayRank(right.item.name)
      if (rankDiff !== 0)
        return rankDiff
      return left.index - right.index
    })
    .map(({ item }) => item)
}

export function sortVersionsByDisplayOrderDesc<T extends { name?: string }>(items: T[]) {
  return items
    .map((item, index) => ({ item, index }))
    .sort((left, right) => {
      const rankDiff = getVersionDisplayRank(right.item.name) - getVersionDisplayRank(left.item.name)
      if (rankDiff !== 0)
        return rankDiff
      return left.index - right.index
    })
    .map(({ item }) => item)
}

export function filterSystemDefaultVersions<T extends { name?: string }>(items: T[]) {
  const defaultNameSet = new Set<string>(systemDefaultVersionNames)
  return items.filter(item => defaultNameSet.has(String(item.name || '').trim()))
}
