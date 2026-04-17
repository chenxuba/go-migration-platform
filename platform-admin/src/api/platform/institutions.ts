import type { ResponseBody } from '@/utils/request'

export interface InstitutionItem {
  id: number
  organName: string
  organCode?: string
  loginName?: string
  mobile?: string
  principal?: string
  address?: string
  logo?: string
  enabled: boolean
  staffCount: number
  activeStaffCount: number
  adminCount: number
}

export interface InstitutionDetail {
  id: number
  organName: string
  organCode: string
  loginName: string
  mobile: string
  principal?: string
  provinceCode?: number
  province: string
  cityCode?: number
  city: string
  regionCode?: number
  region?: string
  address?: string
  concatPhone?: string
  fixedPhone?: string
  remark?: string
  logo?: string
  enabled: boolean
  status: number
  lng?: number
  lat?: number
  profile: InstitutionProfile
}

export interface InstitutionProfile {
  organLabel?: string
  description?: string
  businessTime?: string
  video?: string
  galleryImages?: string[]
}

export interface InstitutionSummary {
  totalCount: number
  enabledCount: number
  disabledCount: number
}

export interface InstitutionPagePayload {
  items: InstitutionItem[]
  total: number
  current: number
  size: number
  summary?: InstitutionSummary
}

export interface InstitutionPageParams {
  current?: number
  size?: number
  keyword?: string
  mobile?: string
  enabled?: boolean
}

export interface InstitutionMutationPayload {
  id?: number
  organName: string
  loginName: string
  mobile: string
  principal?: string
  provinceCode?: number
  province: string
  cityCode?: number
  city: string
  regionCode?: number
  region?: string
  address?: string
  concatPhone?: string
  fixedPhone?: string
  remark?: string
  logo?: string
  enabled?: boolean
  lng?: number
  lat?: number
  profile?: InstitutionProfile
}

export interface InstitutionGeocodePayload {
  province: string
  city: string
  region?: string
  address: string
}

export interface InstitutionGeocodeResult {
  lng: number
  lat: number
  source: string
  resolvedAddress?: string
}

export function pageInstitutionsApi(params: InstitutionPageParams) {
  return useGet<InstitutionItem[], InstitutionPageParams>('/api/v1/platform/institutions', params) as Promise<
    ResponseBody<InstitutionItem[]> & { data?: InstitutionPagePayload }
  >
}

export function getInstitutionDetailApi(params: { id: number }) {
  return useGet<InstitutionDetail, { id: number }>('/api/v1/platform/institutions/detail', params)
}

export function geocodeInstitutionApi(data: InstitutionGeocodePayload) {
  return usePost<InstitutionGeocodeResult, InstitutionGeocodePayload>('/api/v1/platform/institutions/geocode', data)
}

export function createInstitutionApi(data: InstitutionMutationPayload) {
  return usePost<{ id: number }, InstitutionMutationPayload>('/api/v1/platform/institutions/create', data)
}

export function updateInstitutionApi(data: InstitutionMutationPayload & { id: number }) {
  return usePost<boolean, InstitutionMutationPayload & { id: number }>('/api/v1/platform/institutions/update', data)
}

export function updateInstitutionStatusApi(data: { id: number, enabled: boolean }) {
  return usePost<boolean, { id: number, enabled: boolean }>('/api/v1/platform/institutions/status', data)
}
