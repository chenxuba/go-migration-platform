import { AxiosError } from 'axios'
import router from '~/router'
import { useMetaTitle } from '~/composables/meta-title'
import { hasGovernmentPortalAccess } from '~/utils/government-auth'
import { setRouteEmitter } from '~@/utils/route-listener'

const allowList = ['/government/login', '/government/error', '/government/401', '/government/404', '/government/403', '/government/502']
const loginPath = '/government/login'

router.beforeEach(async (to, from, next) => {
  // 正常路由处理流程
  setRouteEmitter(to)
  const userStore = useUserStore()
  const token = useAuthorization()
  const { hasAccess } = useAccess()

  if (!token.value) {
    //  如果token不存在就跳转到登录页面
    if (!allowList.includes(to.path) && !to.path.startsWith('/government/redirect')) {
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
    if ((!userStore.userInfo || !userStore.routerData) && !allowList.includes(to.path) && !to.path.startsWith('/government/redirect')) {
      try {
        // 获取用户信息
        const userInfo = userStore.userInfo || await userStore.getUserInfo()
        if (!hasGovernmentPortalAccess(userInfo)) {
          await userStore.logout()
          next({
            path: '/government/403',
            replace: true,
          })
          return
        }
        // 获取机构配置
        await userStore.getInstConfig()
        // 获取路由菜单的信息
        const currentRoute = await userStore.generateDynamicRoutes()
        router.addRoute(currentRoute)
        if (to.meta?.access && !hasAccess(to.meta.access)) {
          next({
            path: '/government/403',
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
            path: '/government/401',
          })
        }
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

  if (token.value && to.meta?.access && !hasAccess(to.meta.access)) {
    next({
      path: '/government/403',
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
})
