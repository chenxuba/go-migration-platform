import { usePost } from '~/utils/request'

export const CAMPUS_DATA_CLEAR_SCOPE_BUSINESS_ONLY = 'business_only'

export interface CampusDataClearResult {
  scope: string
  scopeName: string
  cleared: {
    students: number
    studentFieldValues: number
    studentChangeRecords: number
    studentTeachingRecords: number
    studentTeachingChangeLogs: number
    studentRehabRecords: number
    followRecords: number
    homeworkTasks: number
    noticeRecords: number
    orders: number
    orderCourseDetails: number
    orderPaymentDetails: number
    ledgers: number
    approvalRecords: number
    approvalHistories: number
    tuitionAccounts: number
    tuitionAccountFlows: number
    closeTuitionAccountOrders: number
    refundTuitionOrders: number
    refundTuitionOrderItems: number
    suspendResumeTuitionOrders: number
    rechargeAccounts: number
    rechargeAccountStudents: number
    rechargeAccountFlows: number
    rechargeAccountOrders: number
    rechargeAccountOrderTags: number
    rechargeAccountBills: number
    rechargeAccountBillFlows: number
    courses: number
    courseDetails: number
    courseQuotations: number
    coursePropertyResults: number
    productPackages: number
    productPackageItems: number
    productPackageProperties: number
    importTasks: number
    importTaskRecords: number
    orderImportTasks: number
    orderImportTaskRecords: number
    rechargeImportTasks: number
    rechargeImportTaskRecords: number
    exportRecords: number
    templateMessageRecords: number
    templateMessageRecordItems: number
    weChatBindTickets: number
    weChatStudentBindings: number
    courseSaleVolumesReset: number
    teachingClasses: number
    teachingClassStudents: number
    teachingClassTeachers: number
    teachingSchedules: number
    teachingScheduleStudents: number
    teachingScheduleBatchMetas: number
    teachingRecords: number
    teachingClassOperationLogs: number
    teachingClassEntryExits: number
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
