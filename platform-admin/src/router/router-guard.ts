import { AxiosError } from 'axios'
import router from '~/router'
import { useMetaTitle } from '~/composables/meta-title'
import { setRouteEmitter } from '~@/utils/route-listener'
import { useModalStore } from '~/stores/modal'
import { useLayoutMenu } from '~/stores/layout-menu'

const allowList = ['/login', '/error', '/401', '/404', '/403','/502']
const loginPath = '/login'

router.beforeEach(async (to, from, next) => {
  // 处理模态框路由
  const modalStore = useModalStore()
  if (to.meta.isModal) {
    // 打开对应的模态框
    modalStore.openModal(to.name as string, to.params)

    // 获取 layout-menu store，防止菜单选中状态更新
    const layoutMenu = useLayoutMenu()
    const currentSelectedKeys = [...layoutMenu.selectedKeys]
    const currentOpenKeys = [...layoutMenu.openKeys]

    // 返回到来源页面，保持当前路由状态不变
    next({
      path: from.path,
      replace: true,
    })

    // 确保菜单选中状态不变
    setTimeout(() => {
      layoutMenu.selectedKeys = currentSelectedKeys
      layoutMenu.openKeys = currentOpenKeys
    }, 10)

    return
  }

  // 正常路由处理流程
  setRouteEmitter(to)
  const userStore = useUserStore()
  const token = useAuthorization()

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
        // 获取机构配置
        await userStore.getInstConfig()
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
          // 跳转到error页面
          next({
            path: '/401',
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
  next()
})

router.afterEach((to) => {
  useMetaTitle(to)
  useLoadingCheck()
  useScrollToTop()
})
