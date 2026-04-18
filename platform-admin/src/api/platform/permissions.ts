export interface PermissionMenuItem {
  id: number
  menuName: string
  menuCode: string
  pid: number
  sort?: number
  weight?: number
  remark?: string
  introduce?: string
  ownType?: number
  children?: PermissionMenuItem[]
}

export interface PermissionMutationPayload {
  id?: number
  menuName: string
  menuCode: string
  pid?: number
  sort?: number
  weight?: number
  remark?: string
  introduce?: string
  ownType?: number
}

export function getPermissionTreeApi(params: { ownType?: number, menuName?: string } = {}) {
  return useGet<PermissionMenuItem[], { ownType?: number, menuName?: string }>('/sso/menu/list', {
    ownType: Number(params.ownType ?? 0),
    menuName: params.menuName,
  })
}

export function createPermissionApi(data: PermissionMutationPayload) {
  return usePost<PermissionMenuItem, PermissionMutationPayload>('/sso/menu/create', data)
}

export function updatePermissionApi(data: PermissionMutationPayload & { id: number }) {
  return usePost<PermissionMenuItem, PermissionMutationPayload & { id: number }>('/sso/menu/update', data)
}

export function deletePermissionApi(data: { id: number }) {
  return usePost<boolean, { id: number }>('/sso/menu/delete', data)
}
