import type { RouteRecordRaw } from 'vue-router'

type ComponentLoader = () => Promise<unknown>

const loadedComponents = new Set<string>()
let preloadingAllRoutes = false

function canPreload() {
  if (typeof window === 'undefined')
    return false

  const connection = (navigator as any).connection
  if (connection?.saveData)
    return false

  return !/2g/i.test(connection?.effectiveType || '')
}

function wait(ms: number) {
  return new Promise(resolve => window.setTimeout(resolve, ms))
}

function runWhenIdle(task: () => void) {
  const requestIdleCallback = (window as any).requestIdleCallback
  if (requestIdleCallback) {
    requestIdleCallback(task, { timeout: 2500 })
    return
  }

  window.setTimeout(task, 1200)
}

function isComponentLoader(component: RouteRecordRaw['component']): component is ComponentLoader {
  return typeof component === 'function'
}

function flattenRoutes(routes: RouteRecordRaw[] = [], output: RouteRecordRaw[] = []) {
  for (const route of routes) {
    output.push(route)
    if (route.children?.length)
      flattenRoutes(route.children, output)
  }
  return output
}

async function preloadComponent(component: RouteRecordRaw['component'], key: string) {
  if (!isComponentLoader(component) || loadedComponents.has(key))
    return

  loadedComponents.add(key)
  try {
    await component()
  }
  catch (error) {
    loadedComponents.delete(key)
    console.warn('preload route component failed', error)
  }
}

function getRouteComponents(routerData: RouteRecordRaw, route: RouteRecordRaw) {
  const components: Array<{ key: string, component: RouteRecordRaw['component'] }> = []

  if (routerData.component) {
    components.push({
      key: 'root-layout',
      component: routerData.component,
    })
  }

  const parentComps = route.meta?.parentComps
  if (Array.isArray(parentComps)) {
    parentComps.forEach((component, index) => {
      components.push({
        key: `${route.path}:parent:${index}`,
        component,
      })
    })
  }

  if (route.component) {
    components.push({
      key: `${route.path}:self`,
      component: route.component,
    })
  }

  return components
}

function findRouteByPath(routerData: RouteRecordRaw, path: string) {
  const routes = flattenRoutes([routerData])
  return routes.find(route => route.path === path || route.meta?.originPath === path)
}

export function preloadRouteByPath(routerData: RouteRecordRaw | undefined, path: string) {
  if (!routerData || !path || !canPreload())
    return

  const route = findRouteByPath(routerData, path)
  if (!route)
    return

  for (const item of getRouteComponents(routerData, route))
    void preloadComponent(item.component, item.key)
}

export function scheduleAccessibleRoutePreload(routerData: RouteRecordRaw | undefined, currentPath: string) {
  if (!routerData || preloadingAllRoutes || !canPreload())
    return

  preloadingAllRoutes = true
  runWhenIdle(() => {
    void (async () => {
      const routes = flattenRoutes([routerData])
        .filter(route => route.path && route.path !== currentPath && !route.redirect && route.component)

      for (const route of routes) {
        for (const item of getRouteComponents(routerData, route))
          await preloadComponent(item.component, item.key)

        await wait(80)
      }
    })()
  })
}
