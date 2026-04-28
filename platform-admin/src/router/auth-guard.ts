import type { NavigationGuardNext, RouteLocationNormalized } from 'vue-router'
import { AxiosError } from 'axios'
import router from '~/router'
import { useAccess } from '~/composables/access'
import { useAuthorization } from '~/composables/authorization'
import { useUserStore } from '~/stores/user'
import { preloadRouteByPath, scheduleAccessibleRoutePreload } from './route-preload'

const errorPageList = ['/platform/401', '/platform/404', '/platform/403', '/platform/502']
const loginPath = '/platform/login'

function findFirstMenuPath(items: any[] = []): string {
  for (const item of items) {
    const childPath = findFirstMenuPath(item?.children || [])
    if (childPath)
      return childPath
    if (item?.path && !item.hideInMenu)
      return item.path
  }
  return ''
}

function resolveAccessibleHome(userStore: ReturnType<typeof useUserStore>) {
  return findFirstMenuPath(userStore.menuData as any[]) || '/platform/control-overview'
}

function shouldRedirectToHome(path: string) {
  return path === loginPath || path === '/' || errorPageList.includes(path)
}

export async function handleAuthGuard(
  to: RouteLocationNormalized,
  _from: RouteLocationNormalized,
  next: NavigationGuardNext,
) {
  const userStore = useUserStore()
  const { hasAccess } = useAccess()

  if ((!userStore.userInfo || !userStore.routerData) && !to.path.startsWith('/platform/redirect')) {
    try {
      // 登录接口已经返回 user 时直接复用，刷新页面才补请求 /sso/info。
      if (!userStore.userInfo)
        await userStore.getUserInfo()
      // 获取机构配置
      if (!userStore.instConfig)
        await userStore.getInstConfig()
      // 获取路由菜单的信息
      const currentRoute = await userStore.generateDynamicRoutes()
      router.addRoute(currentRoute)
      const accessibleHome = resolveAccessibleHome(userStore)
      const nextPath = shouldRedirectToHome(to.path) ? accessibleHome : to.path
      preloadRouteByPath(currentRoute, nextPath)
      if (shouldRedirectToHome(to.path)) {
        next({ path: accessibleHome, replace: true })
        return
      }
      if (to.meta?.access && !hasAccess(to.meta.access)) {
        next({
          path: to.path === '/platform/control-overview' ? accessibleHome : '/platform/403',
          replace: true,
        })
        return
      }
      next({
        ...to,
        replace: true,
      })
      return
    }
    catch (e) {
      if (e instanceof AxiosError && e?.response?.status === 401) {
        next({
          path: '/platform/401',
        })
        return
      }
      throw e
    }
  }

  const accessibleHome = resolveAccessibleHome(userStore)
  // 如果当前是登录页面或错误页就跳转到有权限首页
  if (shouldRedirectToHome(to.path)) {
    next({
      path: accessibleHome,
      replace: true,
    })
    return
  }

  if (to.meta?.access && !hasAccess(to.meta.access)) {
    next({
      path: to.path === '/platform/control-overview' ? accessibleHome : '/platform/403',
      replace: true,
    })
    return
  }

  next()
}

export function handleAuthAfterEach(to: RouteLocationNormalized) {
  const token = useAuthorization()
  const userStore = useUserStore()
  if (token.value && userStore.routerData)
    scheduleAccessibleRoutePreload(userStore.routerData, to.path)
}
