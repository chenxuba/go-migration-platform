import { toArray } from '@v-c/utils'
import { PlatformAccessEnum, normalizePlatformAccessCode, type AccessCodeLike } from '~@/constants/access'

export function useAccess() {
  const userStore = useUserStore()
  const roles = computed(() => userStore.roles)
  const hasAccess = (roles: AccessCodeLike) => {
    const accessRoles = (Array.isArray(userStore.roles) ? userStore.roles : [])
      .map(item => normalizePlatformAccessCode(item))
      .filter(Boolean)
    if (accessRoles.includes(PlatformAccessEnum.superAdmin.code))
      return true

    const roleArr = toArray(roles as any)
      .flat(1)
      .map(item => normalizePlatformAccessCode(item))
      .filter(Boolean)
    return roleArr.some(role => accessRoles.includes(role))
  }
  return {
    hasAccess,
    roles,
  }
}
