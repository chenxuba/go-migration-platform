export interface UserInfo {
  id: number | string
  username: string
  nickName: string
  avatar: string
  logo?: string
  isAdmin: number
  roles?: (string | number)[]
  menuCodeList?: (string | number)[]
  orgName: string
  instId: number | string
  instUserId: number | string
  deptIds: number[]
  tenantId?: string
  tenantRole?: string
  tenantType?: string
}

export function getUserInfoApi() {
  return useGet<UserInfo>('/sso/info')
}
