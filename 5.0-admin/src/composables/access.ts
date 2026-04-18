import { toArray } from '@v-c/utils'
import type { AccessCodeLike } from '~@/constants/access'

export function useAccess() {
  const userStore = useUserStore()
  const roles = computed(() => userStore.roles)
  const hasAccess = (roles: AccessCodeLike) => {
    const accessRoles = userStore.roles
    const roleArr = toArray(roles).flat(1)
    return roleArr.some(role => accessRoles?.includes(role))
  }
  return {
    hasAccess,
    roles,
  }
}
