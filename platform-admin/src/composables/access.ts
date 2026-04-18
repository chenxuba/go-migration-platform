import { toArray } from '@v-c/utils'
import { AccessEnum } from '~@/utils/constant'

export function useAccess() {
  const userStore = useUserStore()
  const roles = computed(() => userStore.roles)
  const hasAccess = (roles: (string | number)[] | string | number | AccessEnum) => {
    const accessRoles = Array.isArray(userStore.roles) ? userStore.roles : []
    if (accessRoles.includes(AccessEnum.superAdmin))
      return true

    const roleArr = toArray(roles).flat(1)
    return roleArr.some(role => accessRoles?.includes(role))
  }
  return {
    hasAccess,
    roles,
  }
}
