import { AxiosError } from 'axios'
import router from '~/router'
import { useMetaTitle } from '~/composables/meta-title'
import { setRouteEmitter } from '~@/utils/route-listener'
import { useAccess } from '@/composables/access'
import { resolveRouteMenuAccess } from './access-meta'

const allowList = ['/login', '/error', '/401', '/404', '/403','/502']
const loginPath = '/login'

router.beforeEach(async (to, from, next) => {
  // 正常路由处理流程
  setRouteEmitter(to)
  const userStore = useUserStore()
  const token = useAuthorization()
  const { hasAccess } = useAccess()

  const ensureRouteAccess = () => {
    const routeAccess = resolveRouteMenuAccess(to)
    if (!routeAccess || routeAccess.length === 0 || to.path === '/403')
      return true
    if (hasAccess(routeAccess))
      return true
    next({
      path: '/403',
      replace: true,
    })
    return false
  }

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
    if (!userStore.userInfo && !allowList.includes(to.path) && !to.path.startsWith('/redirect')) {
      try {
        // 获取用户信息
        await userStore.getUserInfo()
        // 获取路由菜单的信息
        const currentRoute = await userStore.generateDynamicRoutes()
        router.addRoute(currentRoute)
        next({
          ...to,
          replace: true,
        })
        return
      }
      catch (e) {
        if (e instanceof AxiosError && e?.response?.status === 401) {
          token.value = null
          next({
            path: loginPath,
            query: {
              redirect: encodeURIComponent(to.fullPath),
            },
          })
          return
        }
        token.value = null
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
      // 如果当前是登录页面就跳转到首页
      if (to.path === loginPath) {
        next({
          path: '/',
        })
        return
      }
    }
  }
  if (token.value && !allowList.includes(to.path) && !to.path.startsWith('/redirect')) {
    if (!ensureRouteAccess())
      return
  }
  next()
})

router.afterEach((to) => {
  useMetaTitle(to)
  useLoadingCheck()
  useScrollToTop()
})
