import type { MenuTreeNode } from './versions'

export interface DefaultRoleTemplateItem {
  roleId: number
  roleName: string
  isDefault?: boolean
  roleIds?: number[]
  functionalAuthorityCount?: number
  dataAuthorityCount?: number
}

export interface DefaultRoleDetail {
  roleId: number
  roleName: string
  description?: string
  isDefault?: boolean
  menuIds?: MenuTreeNode[]
}

export interface SaveDefaultRolePayload {
  roleName: string
  description?: string
  menuIds?: number[]
  roleType?: number
}

export interface UpdateDefaultRolePayload {
  roleId: number
  roleName: string
  description?: string
  menuIds?: number[]
}

export interface DeleteDefaultRolePayload {
  roleId: number
}

export interface DeleteDefaultRoleResult {
  detachedUsers?: number
}

export function getDefaultRoleTemplatesApi(params: { roleType?: number } = {}) {
  return useGet<DefaultRoleTemplateItem[], { roleType?: number }>(
    '/sso/role/getRoleTemplate',
    params,
    { silentError: true },
  )
}

export function getDefaultRoleDetailApi(params: { roleId: number }) {
  return useGet<DefaultRoleDetail, { roleId: number }>(
    '/sso/role/getDefaultRoleDetail',
    params,
    { silentError: true },
  )
}

export function getRoleMenuIDsApi(params: { roleId: number, ownType?: number }) {
  return useGet<number[], { roleId: number, ownType?: number }>(
    '/sso/role/menuList',
    params,
    { silentError: true },
  )
}

export function createDefaultRoleApi(data: SaveDefaultRolePayload) {
  return usePost<boolean, SaveDefaultRolePayload>(
    '/sso/role/saveDefaultRole',
    data,
    { silentError: true },
  )
}

export function updateDefaultRoleApi(data: UpdateDefaultRolePayload) {
  return usePost<boolean, UpdateDefaultRolePayload>(
    '/sso/role/updateRole',
    data,
    { silentError: true },
  )
}

export function deleteDefaultRoleApi(data: DeleteDefaultRolePayload) {
  return usePost<DeleteDefaultRoleResult, DeleteDefaultRolePayload>(
    '/sso/role/deleteDefaultRole',
    data,
    { silentError: true },
  )
}
