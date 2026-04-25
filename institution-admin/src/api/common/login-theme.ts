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

export interface TenantPublicLoginTheme {
  tenantId: string
  tenantName: string
  entryType: string
  institutionId?: number
  institutionName?: string
  loginBrand: TenantLoginBrandConfig
  matchedBy?: string
}

export function getLoginThemeApi(entryType = 'platform-admin') {
  return useGet<TenantPublicLoginTheme, { entryType: string }>('/platform-api/api/v1/public/login-theme', { entryType }, {
    token: false,
    silentError: true,
  }) as Promise<ResponseBody<TenantPublicLoginTheme>>
}
