export interface InstitutionMenuNode {
  id?: number | string
  menuName: string
  menuCode?: string
  urlPath?: string
  remark?: string
  introduce?: string
  groupCode?: string
  children?: InstitutionMenuNode[]
}

export function getRouteMenusApi() {
  return useGet<InstitutionMenuNode[]>('/sso/menu')
}
