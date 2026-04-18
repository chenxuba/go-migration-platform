import { logoutApi } from '~@/api/common/login'
import { getRouteMenusApi, type InstitutionMenuNode } from '~@/api/common/menu'
import type { UserInfo } from '~@/api/common/user'
import { getUserInfoApi } from '~@/api/common/user'
import { type InstConfig, getInstConfigApi } from '~@/api/common/config'
import type { MenuData } from '~@/layouts/basic-layout/typing'
import { rootRoute } from '~@/router/constant'
import { generateFlatRoutes, generateRoutes, generateTreeRoutes } from '~@/router/generate-route'
import { DYNAMIC_LOAD_WAY, DynamicLoadEnum } from '~@/utils/constant'
import type { RouteRecordRaw } from 'vue-router'

type DynamicRouteResult = {
  menuData: MenuData
  routeData: RouteRecordRaw[]
  permissionTree?: InstitutionMenuNode[]
  homePath?: string
}

export const useUserStore = defineStore('user', () => {
  const routerData = shallowRef()
  const menuData = shallowRef<MenuData>([])
  const permissionTree = shallowRef<InstitutionMenuNode[]>([])
  const userInfo = shallowRef<UserInfo>()
  const instConfig = shallowRef<InstConfig>()
  const token = useAuthorization()
  const isAdmin = computed(() => userInfo.value.isAdmin === 1)
  const avatar = computed(() => userInfo.value?.avatar)
  const nickname = computed(() => userInfo.value?.nickName ?? userInfo.value?.username)
  const roles = computed(() => userInfo.value?.menuCodeList)
  const instUserId = computed(() => userInfo.value?.instUserId)
  const deptIds = computed(() => userInfo.value?.deptIds)

  const getMenuRoutes = async (): Promise<DynamicRouteResult> => {
    const { result } = await getRouteMenusApi()
    return generateTreeRoutes((result ?? []) as unknown as MenuData) as DynamicRouteResult
  }

  function normalizeUserInfo(result: any): UserInfo {
    if (!result)
      return result

    return {
      ...result,
      avatar: result.avatar ?? result.logo ?? '',
      isAdmin: result.isAdmin ?? (result.admin ? 1 : 0),
      deptIds: result.deptIds ?? [],
      openType: result.openType ?? 0,
      versionName: result.versionName ?? '',
      orgName: result.orgName ?? '',
      instId: result.instId ?? '',
      instUserId: result.instUserId ?? '',
    }
  }

  const generateDynamicRoutes = async () => {
    // 这里判断是前端还是后端动态加载菜单 DYNAMIC_LOAD_WAY 是环境变量里面配置的
    const dynamicLoadWay: () => Promise<DynamicRouteResult> = DYNAMIC_LOAD_WAY === DynamicLoadEnum.BACKEND
      ? getMenuRoutes
      : generateRoutes
    const {
      menuData: treeMenuData,
      routeData,
      permissionTree: scopedPermissionTree = [],
      homePath,
    } = await dynamicLoadWay()
    // console.log(treeMenuData, routeData)

    menuData.value = treeMenuData
    permissionTree.value = scopedPermissionTree

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
      permissionTree.value = []
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
    permissionTree,
    generateDynamicRoutes,
    avatar,
    nickname,
    isAdmin,
    instUserId,
    deptIds
  }
})
