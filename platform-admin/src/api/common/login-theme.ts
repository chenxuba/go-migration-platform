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
  loginBrand: TenantLoginBrandConfig
  matchedBy?: string
}

export function getLoginThemeApi(entryType = 'platform-admin') {
  return useGet<TenantPublicLoginTheme, { entryType: string, _t: string }>('/api/v1/public/login-theme', { entryType, _t: String(Date.now()) }, {
    headers: {
      'Cache-Control': 'no-cache',
      Pragma: 'no-cache',
    },
    token: false,
    silentError: true,
  }) as Promise<ResponseBody<TenantPublicLoginTheme>>
}
