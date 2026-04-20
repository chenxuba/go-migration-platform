import type { UserInfo } from '~/api/common/user'

export function hasGovernmentPortalAccess(userInfo?: Partial<UserInfo> | null) {
  return userInfo?.loginType === 'government'
}
