import type { ResponseBody } from '@/utils/request'

export interface TenantLoginBrandConfig {
  template?: string
  brandName?: string
  logoUrl?: string
  loginTitle?: string
  loginSubtitle?: string
  backgroundUrl?: string
  primaryColor?: string
  copyright?: string
  heroBadge?: string
  heroTitle?: string
  heroDescription?: string
}

export interface TenantBootstrapSummary {
  tenantId: string
  tenantName: string
  tenantType: string
  edition?: string
  status?: string
  isolationMode?: string
  institutionCount: number
  menuCount: number
  moduleCount?: number
  moduleIds?: number[]
  moduleNames?: string[]
  adminUsernames: string[]
  domains: string[]
  adminDomains?: string[]
  institutionDomains?: string[]
  loginBrand?: TenantLoginBrandConfig
  platformLoginBrand?: TenantLoginBrandConfig
  institutionLoginBrand?: TenantLoginBrandConfig
}

export type TenantListItem = TenantBootstrapSummary

export interface TenantMutationPayload {
  tenantId: string
  tenantName: string
  tenantType?: string
  edition?: string
  status?: string
  isolationMode?: string
  domains?: string[]
  adminDomains?: string[]
  institutionDomains?: string[]
  institutionIds?: number[]
  menuIds?: number[]
  moduleIds?: number[]
  adminUsername?: string
  adminPassword?: string
  adminNickName?: string
  adminMobile?: string
  loginBrand?: TenantLoginBrandConfig
  platformLoginBrand?: TenantLoginBrandConfig
  institutionLoginBrand?: TenantLoginBrandConfig
  remark?: string
}

export function listTenantsApi(params: { keyword?: string } = {}) {
  return useGet<TenantListItem[], { keyword?: string }>('/api/v1/platform/tenants', params) as Promise<
    ResponseBody<TenantListItem[]>
  >
}

export function getTenantBootstrapSummaryApi() {
  return useGet<TenantBootstrapSummary>('/api/v1/platform/tenants/bootstrap-summary') as Promise<
    ResponseBody<TenantBootstrapSummary>
  >
}

export function saveTenantApi(data: TenantMutationPayload) {
  return usePost<boolean, TenantMutationPayload>('/api/v1/platform/tenants/save', data)
}
