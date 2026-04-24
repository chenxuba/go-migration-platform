import type { ResponseBody } from '@/utils/request'

export interface VersionItem {
  id: number
  tenantId?: string
  ownerType?: string
  sourceModuleId?: number
  name: string
  type: number
  price: number
  remark?: string
  menuCount: number
  orgCount: number
  createTime?: string
  updateTime?: string
}

export interface VersionDetail {
  moduleId: number
  tenantId?: string
  ownerType?: string
  sourceModuleId?: number
  moduleName: string
  moduleType: number
  price: number
  remark?: string
  menuCount: number
  orgCount: number
  createTime?: string
  updateTime?: string
  selectedMenuIds?: number[]
  menuIds?: MenuTreeNode[]
}

export interface MenuTreeNode {
  id?: number
  menuId?: string | number
  menuName: string
  introduce?: string
  menuType?: number
  groupCode?: string
  weight?: number
  isSelect?: boolean
  checked?: boolean
  children?: MenuTreeNode[]
}

export interface VersionPagePayload {
  items: VersionItem[]
  total: number
  current: number
  size: number
}

export interface VersionPageParams {
  current?: number
  size?: number
  name?: string
  type?: number
}

export interface VersionMutationPayload {
  id?: number
  name: string
  type: number
  price?: number
  remark?: string
  menuIds?: number[]
}

export function pageVersionsApi(params: VersionPageParams) {
  return useGet<VersionItem[], VersionPageParams>('/api/v1/platform/modules', params) as Promise<
    ResponseBody<VersionItem[]> & { data?: VersionPagePayload }
  >
}

export function getVersionDetailApi(params: { moduleId: number }) {
  return useGet<VersionDetail, { moduleId: number }>('/api/v1/platform/modules/detail', params)
}

export function createVersionApi(data: VersionMutationPayload) {
  return usePost<{ id: number }, VersionMutationPayload>('/api/v1/platform/modules/create', data)
}

export function updateVersionApi(data: VersionMutationPayload & { id: number }) {
  return usePost<boolean, VersionMutationPayload & { id: number }>('/api/v1/platform/modules/update', data)
}

export function saveVersionMenusApi(data: { id: number, menuIds: number[] }) {
  return usePost<boolean, { id: number, menuIds: number[] }>('/api/v1/platform/modules/permissions', data)
}

export function getVersionMenuTreeApi(params: { type?: number } = {}) {
  return useGet<MenuTreeNode[], { type?: number }>('/api/v1/platform/modules/menu-tree', {
    type: Number(params.type || 1) || 1,
  })
}

export function getInstitutionMenuTreeApi(params: { ownType: number | string }) {
  const normalizedOwnType = String(params.ownType || '').toUpperCase() === 'INSTITUTION'
    ? 2
    : params.ownType

  return useGet<MenuTreeNode[], { ownType: number | string }>('/sso/menu/list', {
    ...params,
    ownType: normalizedOwnType,
  })
}
