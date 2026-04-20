import type { ResponseBody } from '@/utils/request'

export type GovernmentLevel = 'super' | 'province' | 'city' | 'district'

export interface GovernmentAccountScopeItem {
  id?: number
  scopeLevel: GovernmentLevel | string
  provinceCode?: string
  provinceName?: string
  cityCode?: string
  cityName?: string
  districtCode?: string
  districtName?: string
  displayName?: string
}

export interface GovernmentAccountItem {
  id: number
  username: string
  mobile?: string
  nickName?: string
  roleId?: string
  roleName?: string
  isAdmin?: boolean
  disabled?: boolean
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

export interface GovernmentRoleOption {
  roleId: number
  roleName: string
  level: GovernmentLevel | string
  levelLabel?: string
  isAdmin?: boolean
}

export interface GovernmentAccountDetail {
  id: number
  username: string
  mobile: string
  nickName: string
  disabled: boolean
  level: GovernmentLevel | string
  levelLabel?: string
  roleId: number
  roleName?: string
  lastLoginTime?: string
  scopes: GovernmentAccountScopeItem[]
}

export interface GovernmentAccountMutationPayload {
  id?: number
  username: string
  password?: string
  mobile: string
  nickName: string
  disabled?: boolean
  level: GovernmentLevel | string
  roleId: number
  scopes: GovernmentAccountScopeItem[]
}

export function pageGovernmentAccountsApi(params: GovernmentAccountPageParams) {
  return useGet<GovernmentAccountItem[], GovernmentAccountPageParams>('/sso/governmentUsers', params) as Promise<
    ResponseBody<GovernmentAccountItem[]> & { data?: GovernmentAccountPagePayload }
  >
}

export function getGovernmentRoleOptionsApi() {
  return useGet<GovernmentRoleOption[]>('/sso/governmentRoles', undefined, { silentError: true })
}

export function getGovernmentAccountDetailApi(params: { id: number }) {
  return useGet<GovernmentAccountDetail, { id: number }>('/sso/governmentUserDetail', params, { silentError: true })
}

export function createGovernmentAccountApi(data: GovernmentAccountMutationPayload) {
  return usePost<{ id: number }, GovernmentAccountMutationPayload>('/sso/governmentUserCreate', data, { silentError: true })
}

export function updateGovernmentAccountApi(data: GovernmentAccountMutationPayload & { id: number }) {
  return usePost<boolean, GovernmentAccountMutationPayload & { id: number }>('/sso/governmentUserUpdate', data, { silentError: true })
}

export function updateGovernmentAccountStatusApi(data: { id: number, disabled: boolean }) {
  return usePost<boolean, { id: number, disabled: boolean }>('/sso/governmentUserStatus', data, { silentError: true })
}
