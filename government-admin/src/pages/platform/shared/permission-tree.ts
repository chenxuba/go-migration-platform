export interface RawPermissionNode {
  id?: number | string
  menuId?: number | string
  menuName?: string
  title?: string
  checked?: boolean
  isSelect?: boolean
  children?: RawPermissionNode[]
}

export interface PermissionTreeNode {
  key: number
  title: string
  children?: PermissionTreeNode[]
  disableCheckbox?: boolean
  disabled?: boolean
}

function normalizeNodeKey(node: RawPermissionNode) {
  const rawValue = Number(node.id ?? node.menuId ?? 0)
  return Number.isFinite(rawValue) ? rawValue : 0
}

export function buildPermissionTreeData(
  nodes: RawPermissionNode[] = [],
  options?: {
    disabledKeySet?: Set<number>
    disableUnmatched?: boolean
  },
): PermissionTreeNode[] {
  return (nodes || [])
    .map((node) => {
      const key = normalizeNodeKey(node)
      if (!key)
        return null

      const children = buildPermissionTreeData(node.children || [], options)
      const shouldDisable = !!options?.disableUnmatched && !!options?.disabledKeySet && !options.disabledKeySet.has(key)

      return {
        key,
        title: String(node.menuName || node.title || ''),
        children: children.length ? children : undefined,
        disableCheckbox: shouldDisable,
        disabled: false,
      }
    })
    .filter(Boolean) as PermissionTreeNode[]
}

export function collectLeafCheckedKeys(nodes: RawPermissionNode[] = []) {
  const result: number[] = []

  const walk = (items: RawPermissionNode[]) => {
    items.forEach((item) => {
      const key = normalizeNodeKey(item)
      const children = Array.isArray(item.children) ? item.children : []
      if (children.length > 0) {
        walk(children)
        return
      }

      if ((item.checked || item.isSelect) && key > 0)
        result.push(key)
    })
  }

  walk(nodes)
  return result
}

export function collectLeafKeysBySelectedSet(nodes: PermissionTreeNode[] = [], selectedKeys: Array<number | string> = []) {
  const selectedKeySet = new Set(selectedKeys.map(item => Number(item)).filter(item => Number.isFinite(item) && item > 0))
  const result: number[] = []

  const walk = (items: PermissionTreeNode[]) => {
    items.forEach((item) => {
      const children = Array.isArray(item.children) ? item.children : []
      if (children.length > 0) {
        walk(children)
        return
      }

      if (selectedKeySet.has(Number(item.key)))
        result.push(Number(item.key))
    })
  }

  walk(nodes)
  return result
}

export function collectExpandKeysByKeyword(nodes: PermissionTreeNode[] = [], keyword = '') {
  const value = String(keyword || '').trim().toLowerCase()
  if (!value)
    return []

  const expandKeys = new Set<number>()

  const walk = (items: PermissionTreeNode[], parentKeys: number[] = []) => {
    items.forEach((item) => {
      const currentParents = [...parentKeys, item.key]
      if (String(item.title || '').toLowerCase().includes(value)) {
        parentKeys.forEach(key => expandKeys.add(key))
      }
      if (Array.isArray(item.children) && item.children.length > 0)
        walk(item.children, currentParents)
    })
  }

  walk(nodes)
  return Array.from(expandKeys)
}

export function collectAllKeys(nodes: PermissionTreeNode[] = []) {
  const result: number[] = []

  const walk = (items: PermissionTreeNode[]) => {
    items.forEach((item) => {
      result.push(Number(item.key))
      if (Array.isArray(item.children) && item.children.length > 0)
        walk(item.children)
    })
  }

  walk(nodes)
  return result
}

export function countLeafNodes(nodes: PermissionTreeNode[] = []) {
  let total = 0

  const walk = (items: PermissionTreeNode[]) => {
    items.forEach((item) => {
      const children = Array.isArray(item.children) ? item.children : []
      if (children.length > 0) {
        walk(children)
        return
      }
      total += 1
    })
  }

  walk(nodes)
  return total
}
