import type { ResponseBody } from '@/utils/request'

export interface GovernmentAccountItem {
  id: number
  username: string
  mobile?: string
  nickName?: string
  roleId?: string
  roleName?: string
  isAdmin?: boolean
  status?: string
  level?: string
  scope?: string
  lastLoginTime?: string
}

export interface GovernmentAccountPagePayload {
  items: GovernmentAccountItem[]
  total: number
  current: number
  size: number
}

export interface GovernmentAccountPageParams {
  current?: number
  size?: number
  username?: string
  mobile?: string
}

export function pageGovernmentAccountsApi(params: GovernmentAccountPageParams) {
  return useGet<GovernmentAccountItem[], GovernmentAccountPageParams>('/sso/governmentUsers', params) as Promise<
    ResponseBody<GovernmentAccountItem[]> & { data?: GovernmentAccountPagePayload }
  >
}
