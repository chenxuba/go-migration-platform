import type { ResponseBody } from '@/utils/request'

export interface InstitutionItem {
  id: number
  organName: string
  organCode?: string
  loginName?: string
  mobile?: string
  principal?: string
  province?: string
  city?: string
  region?: string
  address?: string
  logo?: string
  enabled: boolean
  status?: number
  openType?: number
  openDuration?: string
  registerTime?: string
  expireEndTime?: string
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
  openType: number
  openDuration?: string
  expireStartTime?: string
  expireEndTime?: string
  lng?: number
  lat?: number
  profile: InstitutionProfile
}

export interface InstitutionProfile {
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
  status?: number
  openType?: number
  provinceCode?: number
  cityCode?: number
  regionCode?: number
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
  openType?: number
  openDuration?: string
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

export interface InstitutionRenewalRecord {
  id: number
  institutionId: number
  beforeOpenType: number
  beforeOpenDuration?: string
  beforeExpireEndTime?: string
  afterOpenType: number
  renewDuration?: string
  renewStartTime?: string
  afterExpireEndTime?: string
  operatorId?: number
  createTime?: string
}

export interface InstitutionRenewalMutationPayload {
  institutionId: number
  openType: number
  openDuration: string
}

export interface InstitutionRenewalResult {
  institutionId: number
  openType: number
  openDuration?: string
  expireStartTime?: string
  expireEndTime?: string
}

export interface InstitutionPermissionDetail {
  institutionId: number
  organName: string
  mobile?: string
  openType: number
  openDuration?: string
  status: number
  expireEndTime?: string
  currentModuleId?: number
  currentModuleName?: string
  adminRoleId?: number
  adminRoleName?: string
  templateMenuIds?: number[]
  effectiveMenuIds?: number[]
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
  return usePost<InstitutionGeocodeResult, InstitutionGeocodePayload>(
    '/api/v1/platform/institutions/geocode',
    data,
    { silentError: true },
  )
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

export function getInstitutionRenewalRecordsApi(params: { institutionId: number }) {
  return useGet<InstitutionRenewalRecord[], { institutionId: number }>(
    '/api/v1/platform/institutions/renewal-records',
    params,
  )
}

export function renewInstitutionApi(data: InstitutionRenewalMutationPayload) {
  return usePost<InstitutionRenewalResult, InstitutionRenewalMutationPayload>(
    '/api/v1/platform/institutions/renew',
    data,
  )
}

export function getInstitutionPermissionDetailApi(params: { institutionId: number }) {
  return useGet<InstitutionPermissionDetail, { institutionId: number }>(
    '/api/v1/platform/institutions/permission-detail',
    params,
  )
}

export function replaceInstitutionPermissionVersionApi(data: { institutionId: number, moduleId: number, menuIds?: number[] }) {
  return usePost<boolean, { institutionId: number, moduleId: number, menuIds?: number[] }>(
    '/api/v1/platform/institutions/permission-version',
    data,
  )
}
