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

const LOGIN_THEME_CACHE_PREFIX = 'PLATFORM_ADMIN_LOGIN_THEME'

export function getLoginThemeCacheKey(entryType = 'platform-admin') {
  if (typeof window === 'undefined')
    return `${LOGIN_THEME_CACHE_PREFIX}:${entryType}`

  const hostname = window.location.hostname.toLowerCase() || 'unknown'
  return `${LOGIN_THEME_CACHE_PREFIX}:${entryType}:${hostname}`
}

export function readCachedLoginBrand(entryType = 'platform-admin') {
  if (typeof window === 'undefined')
    return undefined

  const cacheKey = getLoginThemeCacheKey(entryType)
  const legacyCacheKey = `${LOGIN_THEME_CACHE_PREFIX}:${entryType}`
  try {
    const raw = window.localStorage.getItem(cacheKey) || window.localStorage.getItem(legacyCacheKey)
    if (!raw)
      return undefined
    return JSON.parse(raw) as TenantLoginBrandConfig
  }
  catch {
    return undefined
  }
}

export function writeCachedLoginBrand(next?: TenantLoginBrandConfig, entryType = 'platform-admin') {
  if (typeof window === 'undefined' || !next)
    return

  const cacheKey = getLoginThemeCacheKey(entryType)
  const legacyCacheKey = `${LOGIN_THEME_CACHE_PREFIX}:${entryType}`
  try {
    window.localStorage.setItem(cacheKey, JSON.stringify(next))
    window.localStorage.removeItem(legacyCacheKey)
  }
  catch {
    // Ignore storage quota and privacy-mode errors.
  }
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
