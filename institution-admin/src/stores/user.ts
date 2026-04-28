import { logoutApi } from '~@/api/common/login'
import type { UserInfo } from '~@/api/common/user'
import { getUserInfoApi } from '~@/api/common/user'
import { type InstConfig, getInstConfigApi } from '~@/api/common/config'
import type { MenuData } from '~@/layouts/basic-layout/typing'
import { rootRoute } from '~@/router/constant'
import { generateFlatRoutes, generateRoutes } from '~@/router/generate-route'
import type { RouteRecordRaw } from 'vue-router'

type DynamicRouteResult = {
  menuData: MenuData
  routeData: RouteRecordRaw[]
  homePath?: string
}

export const useUserStore = defineStore('user', () => {
  const routerData = shallowRef()
  const menuData = shallowRef<MenuData>([])
  const userInfo = shallowRef<UserInfo>()
  const instConfig = shallowRef<InstConfig>()
  const token = useAuthorization()
  const isAdmin = computed(() => userInfo.value.isAdmin === 1)
  const avatar = computed(() => userInfo.value?.avatar)
  const nickname = computed(() => userInfo.value?.nickName ?? userInfo.value?.username)
  const roles = computed(() => userInfo.value?.menuCodeList)
  const instUserId = computed(() => userInfo.value?.instUserId)
  const deptIds = computed(() => userInfo.value?.deptIds)
  const institutionStatus = computed(() => userInfo.value?.institutionStatus ?? 'normal')
  const institutionReadonly = computed(() => Boolean(userInfo.value?.institutionReadonly))

  function normalizeUserInfo(result: any): UserInfo {
    if (!result)
      return result

    return {
      ...result,
      avatar: result.avatar ?? '',
      logo: result.logo ?? '',
      isAdmin: result.isAdmin ?? (result.admin ? 1 : 0),
      deptIds: result.deptIds ?? [],
      openType: result.openType ?? 0,
      versionName: result.versionName ?? '',
      orgName: result.orgName ?? '',
      instId: result.instId ?? '',
      instUserId: result.instUserId ?? '',
      institutionStatus: result.institutionStatus ?? 'normal',
      institutionReadonly: Boolean(result.institutionReadonly),
    }
  }

  const generateDynamicRoutes = async () => {
    const {
      menuData: treeMenuData,
      routeData,
      homePath,
    } = await generateRoutes()

    menuData.value = treeMenuData

    const redirectPath = homePath || (typeof rootRoute.redirect === 'string' ? rootRoute.redirect : '/') || '/'
    routerData.value = {
      ...rootRoute,
      redirect: redirectPath,
      children: generateFlatRoutes(routeData, redirectPath),
    }
    return routerData.value
  }

  // 获取用户信息
  const getUserInfo = async () => {
    // 获取用户信息
    const { result } = await getUserInfoApi()
    const normalized = normalizeUserInfo(result)
    userInfo.value = normalized
    return normalized
  }
  // 获取机构配置
  const getInstConfig = async () => {
    try {
      const { result } = await getInstConfigApi()
      instConfig.value = result
    }
    catch (error) {
      console.warn('getInstConfig failed, continue with empty config', error)
      instConfig.value = {} as InstConfig
    }
  }

  const logout = async () => {
    // 退出登录
    // 1. 清空用户信息
    try {
      await logoutApi()
    }
    finally {
      token.value = null
      userInfo.value = undefined
      routerData.value = undefined
      menuData.value = []
    }
  }

  return {
    userInfo,
    instConfig,
    roles,
    getUserInfo,
    getInstConfig,
    logout,
    routerData,
    menuData,
    generateDynamicRoutes,
    avatar,
    nickname,
    isAdmin,
    instUserId,
    deptIds,
    institutionStatus,
    institutionReadonly,
  }
})
