import { storeToRefs } from 'pinia'
import { computed } from 'vue'
import { useUserStore } from '~@/stores/user'

export enum ConsoleOwnType {
  PLATFORM = 0,
  TENANT = 1,
}

export function useConsoleOwnType() {
  const userStore = useUserStore()
  const { userInfo } = storeToRefs(userStore)
  return computed(() => userInfo.value?.tenantRole === 'tenant_admin' ? ConsoleOwnType.TENANT : ConsoleOwnType.PLATFORM)
}

export function resolveConsoleOwnType(userInfo?: { tenantRole?: string }) {
  return userInfo?.tenantRole === 'tenant_admin' ? ConsoleOwnType.TENANT : ConsoleOwnType.PLATFORM
}
