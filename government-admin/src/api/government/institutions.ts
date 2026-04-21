import type { ResponseBody } from '@/utils/request'

export interface GovernmentInstitutionItem {
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
  enabled: boolean
  status?: number
  openType?: number
  openDuration?: string
  registerTime?: string
  expireEndTime?: string
  staffCount: number
  activeStaffCount: number
  adminCount: number
  readingStudentCount: number
  intentStudentCount: number
  orderCount: number
}

export interface GovernmentInstitutionSummary {
  totalCount: number
  enabledCount: number
  warningCount: number
  disabledCount: number
  expiredCount: number
  readingStudentCount: number
  intentStudentCount: number
  orderCount: number
}

export interface GovernmentInstitutionPagePayload {
  items: GovernmentInstitutionItem[]
  total: number
  current: number
  size: number
  level: string
  levelLabel: string
  scopeText: string
  scopeCodeText: string
  scopeCount: number
  summary?: GovernmentInstitutionSummary
}

export interface GovernmentInstitutionPageParams {
  current?: number
  size?: number
  keyword?: string
  status?: number
  openType?: number
}

export function pageGovernmentInstitutionsApi(params: GovernmentInstitutionPageParams) {
  return useGet<GovernmentInstitutionItem[], GovernmentInstitutionPageParams>(
    '/api/v1/platform/government/institutions',
    params,
    { silentError: true },
  ) as Promise<ResponseBody<GovernmentInstitutionItem[]> & { data?: GovernmentInstitutionPagePayload }>
}
