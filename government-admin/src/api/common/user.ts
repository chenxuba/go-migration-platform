export interface UserInfo {
  id: number | string
  loginType?: string
  username: string
  mobile?: string
  nickName: string
  avatar: string
  logo?: string
  isAdmin: number
  roles?: (string | number)[]
  menuCodeList?: (string | number)[]
  deptId?: number | string
  deptName?: string
  roleId?: string
  roleName?: string
  orgName: string
  instId: number | string
  instUserId: number | string
  deptIds: number[]
}

export function getUserInfoApi() {
  return useGet<UserInfo>('/sso/info')
}
