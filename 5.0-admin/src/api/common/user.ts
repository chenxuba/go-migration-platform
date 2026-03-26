export interface UserInfo {
  id: number | string
  username: string
  nickName: string
  avatar: string
  isAdmin: number
  roles?: (string | number)[]
  menuCodeList?: (string | number)[]
  orgName: string
  instId: number | string
  instUserId: number | string
  deptIds: number[]
}

export function getUserInfoApi() {
  return useGet<UserInfo>('/sso/sso/info')
}
