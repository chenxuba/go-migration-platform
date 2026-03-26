import { usePost } from '~/utils/request'

export const CAMPUS_DATA_CLEAR_SCOPE_BUSINESS_ONLY = 'business_only'

export interface CampusDataClearResult {
  scope: string
  scopeName: string
  cleared: {
    students: number
    studentFieldValues: number
    studentChangeRecords: number
    followRecords: number
    orders: number
    orderCourseDetails: number
    orderPaymentDetails: number
    approvalRecords: number
    approvalHistories: number
    tuitionAccounts: number
    importTasks: number
    importTaskRecords: number
    courseSaleVolumesReset: number
  }
  preserved: string[]
  intentStudentIndexCleared: boolean
  intentStudentIndexMessage?: string
}

export interface CampusDataClearPayload {
  scope?: string
}

export function clearCampusDataApi(data: CampusDataClearPayload = { scope: CAMPUS_DATA_CLEAR_SCOPE_BUSINESS_ONLY }) {
  return usePost<CampusDataClearResult>('/api/v1/campus-data/clear', data)
}
