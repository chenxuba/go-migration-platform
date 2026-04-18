import { toArray } from '@v-c/utils'
import type { RouteMeta, RouteRecordRaw } from 'vue-router'
import {
  buildPageUsePermissionCode,
  normalizeAccessCode,
  type AccessCodeLike,
} from '~@/constants/access'
import dynamicRoutes from '~@/router/dynamic-routes'

type RouteLike = {
  path?: string
  meta?: Partial<RouteMeta>
}

export function normalizePath(path?: string) {
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

export function normalizeRouteAccessPath(path?: string) {
  const normalized = normalizePath(path)
  if (!normalized || normalized === '/')
    return normalized

  const segments = normalized.split('/').filter(Boolean)
  const staticSegments = []

  for (const segment of segments) {
    if (segment.startsWith(':'))
      break
    staticSegments.push(segment)
  }

  return `/${staticSegments.join('/')}`
}

export function normalizeAccessList(access?: AccessCodeLike) {
  return toArray(access as any)
    .flat(1)
    .map(item => String(normalizeAccessCode(item) || '').trim())
    .filter(Boolean)
}

export function buildRouteAccessMap(routes: RouteRecordRaw[]) {
  const accessMap = new Map<string, string[]>()

  function walk(items: RouteRecordRaw[]) {
    items.forEach((route) => {
      const routePath = normalizeRouteAccessPath(route.path)
      const routeAccess = normalizeAccessList((route.meta?.menuAccess ?? route.meta?.access) as any)
      if (routePath && routeAccess.length > 0)
        accessMap.set(routePath, routeAccess)
      if (route.children?.length)
        walk(route.children)
    })
  }

  walk(routes)
  return accessMap
}

export function resolveMenuAccessMeta(
  meta?: Partial<RouteMeta>,
  accessMap?: Map<string, string[]>,
  inheritedAccess: string[] = [],
) {
  const ownAccess = normalizeAccessList((meta?.menuAccess ?? meta?.access) as any)
  if (ownAccess.length > 0)
    return ownAccess

  const parentKeys = Array.isArray(meta?.parentKeys)
    ? meta?.parentKeys
    : meta?.parentKeys
      ? [meta.parentKeys]
      : []

  for (const key of parentKeys) {
    const access = accessMap?.get(normalizeRouteAccessPath(String(key)))
    if (access?.length)
      return access
  }

  return inheritedAccess
}

export function buildPageAccessFromMenuAccess(accessList: string[]) {
  const routeAccess = accessList.find(code => String(code || '').startsWith('page:'))
  if (!routeAccess)
    return []

  const pageUseCode = buildPageUsePermissionCode(routeAccess)
  return pageUseCode ? [pageUseCode] : []
}

const dynamicRouteAccessMap = buildRouteAccessMap(dynamicRoutes)

export function resolveRouteMenuAccess(route: RouteLike) {
  return resolveMenuAccessMeta(route.meta, dynamicRouteAccessMap)
}

export function resolveRoutePageAccess(route: RouteLike) {
  const explicitPageAccess = normalizeAccessList(route.meta?.pageAccess as any)
  if (explicitPageAccess.length > 0)
    return explicitPageAccess

  return buildPageAccessFromMenuAccess(resolveRouteMenuAccess(route))
}
