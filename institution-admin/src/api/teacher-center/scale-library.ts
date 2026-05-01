import { useGet } from '~/utils/request'

export interface ScaleLibraryTextResource {
  id: number
  scaleId: number
  content: string
  sort: number
}

export type ScaleLibraryStatus = 'available' | 'unavailable' | string

export interface ScaleLibraryItem {
  id: number
  name: string
  code: string
  category: string
  scenario: string
  ageRange: string
  ageMinMonths: number
  ageMaxMonths: number
  duration: string
  durationMinMinutes: number
  durationMaxMinutes: number
  currentVersion: string
  itemCount: number
  domainCount: number
  institutionCount: number
  monthUsage: number
  usageCount: number
  latestUse: string
  dataStatus: string
  status: ScaleLibraryStatus
  statusText: string
  updatedAt: string
  summary: string
  posterUrl: string
  executionEntry: string
  apiPackage: string
  references: ScaleLibraryTextResource[]
  acknowledgements: ScaleLibraryTextResource[]
  authReserved: boolean
  authActionEnabled: boolean
}

export interface ScaleLibrarySummary {
  total: number
  available: number
  unavailable: number
  monthUsage: number
  usageCount: number
  reservedAuths: number
}

export interface ScaleLibraryFilterOptions {
  categories: string[]
  categoryCounts?: Record<string, number>
  scenarios: string[]
  statuses: string[]
  ageScopes: string[]
  durations: string[]
}

export interface ScaleLibraryResponse {
  items: ScaleLibraryItem[]
  summary: ScaleLibrarySummary
  filterOptions: ScaleLibraryFilterOptions
}

export interface ScaleLibraryQuery {
  keyword?: string
  category?: string
  scenario?: string
  status?: string
  ageScope?: string
  duration?: string
}

export interface ScaleAssessmentStudentCandidate {
  id: number
  shortName: string
  name: string
  avatarUrl: string
  gender: string
  age: string
  birthDate?: string
  contactPhone: string
  latestAssessment: string
}

export interface ScaleAssessmentStudentCandidateQuery {
  scaleCode?: string
  keyword?: string
  pageIndex?: number
  pageSize?: number
}

export interface ScaleAssessmentStudentCandidateResponse {
  items: ScaleAssessmentStudentCandidate[]
  total: number
  current: number
  size: number
}

export function getScaleLibraryApi(params?: ScaleLibraryQuery) {
  return useGet<ScaleLibraryResponse, ScaleLibraryQuery>('/api/v1/assessments/scales/library', params)
}

export function getScaleAssessmentStudentCandidatesApi(params?: ScaleAssessmentStudentCandidateQuery) {
  return useGet<ScaleAssessmentStudentCandidateResponse, ScaleAssessmentStudentCandidateQuery>('/api/v1/assessments/scales/student-candidates', params)
}
