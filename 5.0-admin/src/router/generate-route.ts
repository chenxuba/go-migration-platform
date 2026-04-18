import { isUrl, toArray } from '@v-c/utils'
import type { RouteRecordRaw } from 'vue-router'
import { omit } from 'lodash'
import { normalizeAccessCode } from '~@/constants/access'
import { basicRouteMap } from './router-modules'
import type { MenuData, MenuDataItem } from '~@/layouts/basic-layout/typing'
import dynamicRoutes from '~@/router/dynamic-routes'
import { ROOT_ROUTE_REDIRECT_PATH } from '~@/router/constant'
import { i18n } from '~@/locales'

let cache_key = 1

const getCacheKey = () => `Cache_Key_${cache_key++}`

function renderTitle(route: RouteRecordRaw) {
  const { title, locale } = route.meta || {}
  if (!title)
    return ''
  return locale ? (i18n.global as any).t(locale) : title
}

function formatMenu(route: RouteRecordRaw, path?: string) {
  return {
    id: route.meta?.id,
    parentId: route.meta?.parentId,
    title: () => renderTitle(route),
    icon: route.meta?.icon || '',
    path: path ?? route.path,
    hideInMenu: route.meta?.hideInMenu || false,
    parentKeys: route.meta?.parentKeys || [],
    hideInBreadcrumb: route.meta?.hideInBreadcrumb || false,
    hideChildrenInMenu: route.meta?.hideChildrenInMenu || false,
    locale: route.meta?.locale,
    keepAlive: route.meta?.keepAlive || false,
    name: route.name as string,
    new: route.meta?.new as boolean,
    url: route.meta?.url || '',
    target: route.meta?.target || '_blank',
  }
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

function normalizeRouteAccessPath(path?: string) {
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

function cloneRoute(route: RouteRecordRaw, children?: RouteRecordRaw[]) {
  const nextRoute = {
    ...route,
    meta: route.meta ? { ...route.meta } : undefined,
  } as RouteRecordRaw

  if (children)
    nextRoute.children = children
  else
    delete nextRoute.children

  return nextRoute
}

function normalizeAccessList(access?: RouteRecordRaw['meta'] extends infer T ? T extends { access?: infer U } ? U : never : never) {
  return toArray(access as any)
    .flat(1)
    .map(item => String(normalizeAccessCode(item) || '').trim())
    .filter(Boolean)
}

function buildRouteAccessMap(routes: RouteRecordRaw[]) {
  const accessMap = new Map<string, string[]>()

  function walk(items: RouteRecordRaw[]) {
    items.forEach((route) => {
      const routePath = normalizeRouteAccessPath(route.path)
      const routeAccess = normalizeAccessList(route.meta?.access as any)
      if (routePath && routeAccess.length > 0)
        accessMap.set(routePath, routeAccess)
      if (route.children?.length)
        walk(route.children)
    })
  }

  walk(routes)
  return accessMap
}

function resolveRouteAccess(route: RouteRecordRaw, accessMap: Map<string, string[]>, inheritedAccess: string[] = []) {
  const ownAccess = normalizeAccessList(route.meta?.access as any)
  if (ownAccess.length > 0)
    return ownAccess

  const parentKeys = Array.isArray(route.meta?.parentKeys)
    ? route.meta.parentKeys
    : route.meta?.parentKeys
      ? [route.meta.parentKeys]
      : []

  for (const key of parentKeys) {
    const access = accessMap.get(normalizeRouteAccessPath(String(key)))
    if (access?.length)
      return access
  }

  return inheritedAccess
}

function filterRoutesByAccess(
  routes: RouteRecordRaw[],
  hasAccess: (roles: (string | number)[] | string | number) => boolean,
  accessMap: Map<string, string[]>,
  inheritedAccess: string[] = [],
) {
  return routes.reduce<RouteRecordRaw[]>((items, route) => {
    const routeAccess = resolveRouteAccess(route, accessMap, inheritedAccess)
    const nextChildren = route.children?.length
      ? filterRoutesByAccess(route.children, hasAccess, accessMap, routeAccess)
      : undefined

    const hasChildren = Array.isArray(route.children) && route.children.length > 0
    if (hasChildren) {
      if (!nextChildren || nextChildren.length === 0) {
        const canVisitCurrentRoute = routeAccess.length === 0 || hasAccess(routeAccess)
        const hasRenderableComponent = !!checkComponent(route.component)
        if (!canVisitCurrentRoute || !hasRenderableComponent)
          return items
      }
    }
    else if (routeAccess.length > 0 && !hasAccess(routeAccess)) {
      return items
    }

    items.push(cloneRoute(route, nextChildren))
    return items
  }, [])
}

function findFirstAccessibleRoutePath(routes: RouteRecordRaw[]): string {
  for (const route of routes) {
    if (route.children?.length) {
      const childPath = findFirstAccessibleRoutePath(route.children)
      if (childPath)
        return childPath
    }

    const routePath = normalizeRouteAccessPath(route.path)
    if (routePath && routePath !== '/' && !route.meta?.hideInMenu)
      return routePath
  }

  return ROOT_ROUTE_REDIRECT_PATH
}

// 本地静态路由生成菜单的信息
export function genRoutes(routes: RouteRecordRaw[], parent?: MenuDataItem) {
  const menuData: MenuData = []
  routes.forEach((route) => {
    let path = route.path
    if (!path.startsWith('/') && !isUrl(path)) {
      // 判断当前是不是以 /开头，如果不是就表示当前的路由地址为不完全的地址
      if (parent)
        path = `${parent.path}/${path}`
      else
        path = `/${path}`
    }
    // 判断是不是存在name，如果不存在name的情况下，自动补充一个自定义的name，为了更容易的去实现保活的功能，name是必须的
    if (!route.name)
      route.name = getCacheKey()
    const item: MenuDataItem = formatMenu(route, path)
    item.children = []
    if (route.children && route.children.length)
      item.children = genRoutes(route.children, item)
    if (item.children?.length === 0)
      delete item.children
    menuData.push(item)
  })
  return menuData
}

export async function generateRoutes() {
  const { hasAccess } = useAccess()
  const accessMap = buildRouteAccessMap(dynamicRoutes)
  const accessRoutes = filterRoutesByAccess(dynamicRoutes, hasAccess, accessMap)
  const menuData = genRoutes(accessRoutes)

  return {
    menuData,
    routeData: accessRoutes,
    homePath: findFirstAccessibleRoutePath(accessRoutes),
  }
}

function checkComponent(component: RouteRecordRaw['component']) {
  for (const componentKey in basicRouteMap) {
    if (component === (basicRouteMap as any)[componentKey])
      return undefined
  }
  return component
}

// 路由拉平处理
function flatRoutes(routes: RouteRecordRaw[], parentName?: string, parentComps: RouteRecordRaw['component'][] = []) {
  const flatRouteData: RouteRecordRaw[] = []
  for (const route of routes) {
    const parentComponents = [...parentComps]
    const currentRoute = omit(route, ['children']) as RouteRecordRaw
    if (!currentRoute.meta)
      currentRoute.meta = {}
    if (parentName)
      currentRoute.meta.parentName = parentName
    if (parentComponents.length > 0)
      currentRoute.meta.parentComps = parentComponents
    currentRoute.meta.originPath = currentRoute.path
    flatRouteData.push(currentRoute)
    if (route.children && route.children.length) {
      const comp = checkComponent(route.component)
      if (comp)
        parentComponents.push(comp)
      flatRouteData.push(...flatRoutes(route.children, route.name as string, [...parentComponents]))
    }
  }
  return flatRouteData
}

export function generateFlatRoutes(routes: RouteRecordRaw[], redirectPath = ROOT_ROUTE_REDIRECT_PATH) {
  const flatRoutesList = flatRoutes(routes)
  // 拿到拉平后的路由，然后统一添加一个父级的路由,通过这层路由实现保活的功能
  const parentRoute: RouteRecordRaw = {
    path: '/',
    redirect: redirectPath,
    name: 'ROOT_EMPTY_PATH',
    // fix: https://github.com/antdv-pro/antdv-pro/issues/179
    // component: getRouterModule('RouteView'),
    children: flatRoutesList,
  }
  return [parentRoute]
}
