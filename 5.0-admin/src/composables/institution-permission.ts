import type { MaybeRefOrGetter } from 'vue'
import { toValue } from 'vue'
import type { InstitutionMenuNode } from '~@/api/common/menu'

type PermissionMatcher =
  | string
  | string[]
  | {
    code?: string | string[]
    groupCode?: string | string[]
  }

function getSingleAccessCode(value?: unknown) {
  if (Array.isArray(value)) {
    for (const item of value) {
      const code = normalizeText(String(item))
      if (code)
        return code
    }
    return ''
  }

  return normalizeText(String(value || ''))
}

function normalizeText(value?: string) {
  return String(value || '').trim()
}

function normalizePath(path?: string) {
  const raw = String(path || '').trim()
  if (!raw)
    return ''

  const withoutQuery = raw.split('?')[0]?.split('#')[0] || ''
  if (!withoutQuery)
    return ''

  if (withoutQuery === '/')
    return '/'

  return withoutQuery.replace(/\/+$/, '')
}

function ensureArray<T>(value?: T | T[]) {
  if (Array.isArray(value))
    return value
  if (value == null)
    return []
  return [value]
}

function matchesText(target: string, candidates: string[]) {
  if (!target || candidates.length === 0)
    return false

  return candidates.some(item => normalizeText(item) === target)
}

function matchMenuNode(node: InstitutionMenuNode, matcher: PermissionMatcher) {
  if (!matcher)
    return false

  if (typeof matcher === 'string' || Array.isArray(matcher)) {
    const codes = ensureArray(matcher).map(item => normalizeText(String(item))).filter(Boolean)
    return matchesText(normalizeText(node.menuCode), codes)
  }

  const codes = ensureArray(matcher.code).map(item => normalizeText(String(item))).filter(Boolean)
  if (matchesText(normalizeText(node.menuCode), codes))
    return true

  const groupCodes = ensureArray(matcher.groupCode).map(item => normalizeText(String(item))).filter(Boolean)
  return matchesText(normalizeText(node.groupCode), groupCodes)
}

function findRouteNode(nodes: InstitutionMenuNode[], matcher: string): InstitutionMenuNode | null {
  const normalizedCode = normalizeText(matcher)
  const normalizedPath = normalizePath(matcher)
  if (!normalizedCode && !normalizedPath)
    return null

  for (const node of nodes) {
    if (normalizedCode && normalizeText(node.menuCode) === normalizedCode)
      return node

    if (normalizedPath && normalizePath(node.urlPath) === normalizedPath)
      return node

    if (node.children?.length) {
      const matched = findRouteNode(node.children, matcher)
      if (matched)
        return matched
    }
  }

  return null
}

export function useInstitutionPermission(routeCode?: MaybeRefOrGetter<string | undefined>) {
  const route = useRoute()
  const userStore = useUserStore()

  const currentRouteCode = computed(() =>
    normalizeText(toValue(routeCode) || getSingleAccessCode(route.meta?.access) || ''),
  )
  const currentPath = computed(() => normalizePath(route.path))
  const permissionTree = computed(() => userStore.permissionTree || [])
  const routeNode = computed(() => findRouteNode(permissionTree.value, currentRouteCode.value || currentPath.value))
  const actionNodes = computed(() => routeNode.value?.children || [])

  const hasRouteAccess = (matcher?: string) =>
    !!findRouteNode(permissionTree.value, normalizeText(matcher || currentRouteCode.value) || normalizePath(matcher || currentPath.value))

  const hasActionAccess = (matcher: PermissionMatcher) => {
    return actionNodes.value.some(node => matchMenuNode(node, matcher))
  }

  const actionNames = computed(() =>
    actionNodes.value
      .map(item => normalizeText(item.menuName))
      .filter(Boolean),
  )

  return {
    routeNode,
    actionNodes,
    actionNames,
    hasRouteAccess,
    hasActionAccess,
  }
}
