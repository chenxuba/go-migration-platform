export interface GovernmentOverviewEntry {
  regionCode: string
  regionName: string
  levelLabel: string
  institutionCount: number
  readingStudentCount: number
  intentStudentCount: number
  orderCount: number
}

export interface GovernmentOverviewPayload {
  level: string
  levelLabel: string
  scopeText: string
  scopeCodeText: string
  scopeCount: number
  institutionCount: number
  subordinateRegionCount: number
  readingStudentCount: number
  orderCount: number
  regionalSummary: GovernmentOverviewEntry[]
}

export function getGovernmentOverviewApi() {
  return useGet<GovernmentOverviewPayload>('/api/v1/platform/government/overview', undefined, { silentError: true })
}
