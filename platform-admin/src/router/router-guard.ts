import { AxiosError } from 'axios'
import router from '~/router'
import { useMetaTitle } from '~/composables/meta-title'
import { setRouteEmitter } from '~@/utils/route-listener'
import { preloadRouteByPath, scheduleAccessibleRoutePreload } from './route-preload'

const allowList = ['/login', '/login-template-preview', '/error', '/401', '/404', '/403', '/502']
const errorPageList = ['/401', '/404', '/403', '/502']
const loginPath = '/login'

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
  return findFirstMenuPath(userStore.menuData as any[]) || '/'
}

function shouldRedirectToHome(path: string) {
  return path === loginPath || path === '/' || errorPageList.includes(path)
}

router.beforeEach(async (to, from, next) => {
  // 正常路由处理流程
  setRouteEmitter(to)
  const userStore = useUserStore()
  const token = useAuthorization()
  const { hasAccess } = useAccess()

  if (!token.value) {
    //  如果token不存在就跳转到登录页面
    if (!allowList.includes(to.path) && !to.path.startsWith('/redirect')) {
      next({
        path: loginPath,
        query: {
          redirect: encodeURIComponent(to.fullPath),
        },
      })
      return
    }
  }
  else {
    if ((!userStore.userInfo || !userStore.routerData) && !to.path.startsWith('/redirect')) {
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
            path: to.path === '/platform/control-overview' ? accessibleHome : '/403',
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
          // 跳转到error页面
          next({
            path: '/401',
          })
        }
      }
    }
    else {
      const accessibleHome = resolveAccessibleHome(userStore)
      // 如果当前是登录页面或错误页就跳转到有权限首页
      if (shouldRedirectToHome(to.path)) {
        next({
          path: accessibleHome,
          replace: true,
        })
        return
      }
    }
  }

  if (token.value && to.meta?.access && !hasAccess(to.meta.access)) {
    const accessibleHome = resolveAccessibleHome(userStore)
    next({
      path: to.path === '/platform/control-overview' ? accessibleHome : '/403',
      replace: true,
    })
    return
  }
  next()
})

router.afterEach((to) => {
  useMetaTitle(to)
  useLoadingCheck()
  useScrollToTop()

  const token = useAuthorization()
  const userStore = useUserStore()
  if (token.value && userStore.routerData)
    scheduleAccessibleRoutePreload(userStore.routerData, to.path)
})
