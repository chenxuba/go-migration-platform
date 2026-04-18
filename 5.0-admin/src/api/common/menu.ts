export interface MenuMetaInfo {
  id: number
  menuName: string
  menuCode: string
  introduce?: string
  accessDeniedImage?: string
}

export function getMenuByCodeApi(params: { menuCode: string, ownType?: number }, options?: { silentError?: boolean }) {
  return useGet<MenuMetaInfo, { menuCode: string, ownType: number }>('/sso/menu/by-code', {
    menuCode: params.menuCode,
    ownType: Number(params.ownType ?? 2) || 2,
  }, {
    silentError: options?.silentError,
  })
}
