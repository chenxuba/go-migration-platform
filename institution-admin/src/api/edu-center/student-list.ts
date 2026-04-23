import axios from 'axios'
import { STORAGE_AUTHORIZE_KEY, useAuthorization } from '~/composables/authorization'
import type { EnrolledStudentInfo } from './enrolled-student'

export interface FieldInfo {
  filter: any
  fieldKey: string
  fieldType: number
  required: boolean
  searched: boolean
  optionsJson: string
  remark: string
}

export interface StudentOverviewStatistics {
  totalStudents: number
  recentMonthNewStudents: number
  previousMonthNewStudents: number
  recentMonthGrowthRate: number
  readingStudents: number
  historyStudents: number
  intentStudents: number
  pendingRenewalStudents: number
  arrearStudents: number
  birthdayStudents: number
  pendingClassStudents: number
  pendingAttentionStudents: number
  absentStudents: number
}

// 学员属性 获取系统默认字段列表  /instStudentFieldKey/getDefaultField
export function getStuDefaultFieldApi(data: FieldInfo) {
  return useGet<FieldInfo>('/api/v1/student-field-keys/default', data)
}
// 学员属性 获取自定义字段列表 /instStudentFieldKey/getCustomField
export function getStuCustomFieldApi(data: FieldInfo) {
  return useGet<FieldInfo>('/api/v1/student-field-keys/custom', data)
}
// 更新字段展示状态 /instStudentFieldKey/updateDisplayStatus
export function updateStuDisplayStatusApi(data: FieldInfo) {
  return usePost<FieldInfo>('/api/v1/student-field-keys/display-status', data)
}
// 新增自定义学员属性 /instStudentFieldKey/addCustomField
export function addStuCustomFieldApi(data: FieldInfo) {
  return usePost<FieldInfo>('/api/v1/student-field-keys/create', data)
}
// 更新自定义学员属性 /instStudentFieldKey/updateCustomField
export function updateStuCustomFieldApi(data: FieldInfo) {
  return usePost<FieldInfo>('/api/v1/student-field-keys/update', data)
}
// 删除自定义学员属性 /instStudentFieldKey/deleteCustomField
export function deleteStuCustomFieldApi(data: FieldInfo) {
  return usePost<FieldInfo>('/api/v1/student-field-keys/delete', data)
}

// 学员管理顶部统计
export function getStudentOverviewStatisticsApi(data: Record<string, unknown> = {}) {
  return useGet<StudentOverviewStatistics>('/api/v1/students/overview-statistics', data)
}

export interface StudentRegistrationArrearItem {
  orderId: string
  orderNumber: string
  studentId: string
  studentName: string
  sex?: number
  avatar?: string
  phone?: string
  arrearAmount: number
  orderAmount: number
  paidAmount: number
  productName: string
  createdTime?: string
}

export interface StudentRegistrationArrearPagedResult {
  list?: StudentRegistrationArrearItem[]
  total?: number
}

export interface StudentRegistrationArrearStatistics {
  totalArrearAmount?: number
}

export interface StudentRegistrationArrearQueryParams {
  pageRequestModel: {
    needTotal?: boolean
    pageSize: number
    pageIndex: number
    skipCount?: number
  }
  queryModel?: {
    orderNumber?: string
    lessonId?: string
    studentId?: string
    keyword?: string
    keywordType?: string
    createdTimeBegin?: string
    createdTimeEnd?: string
  }
}

export function getStudentRegistrationArrearPagedListApi(data: StudentRegistrationArrearQueryParams) {
  return usePost<StudentRegistrationArrearPagedResult>('/api/v1/students/registration-arrears/page', data)
}

export function getStudentRegistrationArrearStatisticsApi(data: StudentRegistrationArrearQueryParams['queryModel'] = {}) {
  return usePost<StudentRegistrationArrearStatistics>('/api/v1/students/registration-arrears/statistics', { queryModel: data })
}

export async function exportStudentRegistrationArrearApi(data: {
  queryModel?: StudentRegistrationArrearQueryParams['queryModel']
}) {
  const token = useAuthorization()
  return axios.post('/api/v1/students/registration-arrears/export', data, {
    responseType: 'blob',
    headers: {
      [STORAGE_AUTHORIZE_KEY]: token.value || '',
      Authorization: token.value ? `Bearer ${token.value}` : '',
      'Accept-Language': 'zh-CN',
    },
  })
}

export interface StudentLessonArrearItem {
  studentId: string
  studentName: string
  sex?: number
  avatar?: string
  phone?: string
  lessonId: string
  lessonName: string
  tuitionAccountId: string
  lessonChargingMode: number
  beInArrearsTotal: number
  recordCount: number
  advisorStaffId?: string
  advisorStaffName?: string
  studentManagerId?: string
  studentManagerName?: string
}

export interface StudentLessonArrearPagedResult {
  list?: StudentLessonArrearItem[]
  total?: number
}

export interface StudentLessonArrearStatistics {
  totalArrearAmount?: number
  totalArrearTime?: number
}

export interface StudentLessonArrearQueryParams {
  pageRequestModel: {
    needTotal?: boolean
    pageSize: number
    pageIndex: number
    skipCount?: number
  }
  queryModel?: {
    lessonId?: string
    studentId?: string
    keyword?: string
    keywordType?: string
  }
}

export function getStudentLessonArrearPagedListApi(data: StudentLessonArrearQueryParams) {
  return usePost<StudentLessonArrearPagedResult>('/api/v1/students/lesson-arrears/page', data)
}

export function getStudentLessonArrearStatisticsApi(data: StudentLessonArrearQueryParams['queryModel'] = {}) {
  return usePost<StudentLessonArrearStatistics>('/api/v1/students/lesson-arrears/statistics', { queryModel: data })
}

export async function exportStudentLessonArrearApi(data: {
  queryModel?: StudentLessonArrearQueryParams['queryModel']
}) {
  const token = useAuthorization()
  return axios.post('/api/v1/students/lesson-arrears/export', data, {
    responseType: 'blob',
    headers: {
      [STORAGE_AUTHORIZE_KEY]: token.value || '',
      Authorization: token.value ? `Bearer ${token.value}` : '',
      'Accept-Language': 'zh-CN',
    },
  })
}

export interface PendingAttentionStudentQueryParams {
  pageRequestModel: {
    pageSize: number
    pageIndex: number
    needTotal?: boolean
    skipCount?: number
  }
  queryModel?: {
    studentId?: string
    sexes?: number[]
    studentStatuses?: number[]
    classIds?: string[]
    ageMin?: number
    ageMax?: number
  }
}

export function getPendingAttentionStudentPagedListApi(data: PendingAttentionStudentQueryParams) {
  return usePost<EnrolledStudentInfo>('/api/v1/students/pending-attention/page', data)
}

export interface PendingRenewalStudentItem {
  tuitionAccountId?: string
  studentId?: string
  sex?: number
  avatar?: string
  lessonId?: string
  lessonName?: string
  studentName?: string
  leftQuantity?: number
  leftFreeQuantity?: number
  enableExpireTime?: boolean
  expireTime?: string
  lessonChargingMode?: number
  totalQuantity?: number
  latestStartTime?: string
  phone?: string
  tuition?: number
  classTeacherList?: Array<{
    id?: string
    name?: string
  }>
  status?: number
  advisorStaffId?: string
  advisorStaffName?: string
  studentManagerId?: string
  studentManagerName?: string
}

export interface PendingRenewalStudentQueryParams {
  pageRequestModel: {
    needTotal?: boolean
    pageSize: number
    pageIndex: number
    skipCount?: number
  }
  queryModel?: {
    studentId?: string
    productId?: string
    productIds?: string[]
    classTeacherId?: string
    classIds?: string[]
    statusList?: number[]
  }
  sortModel?: {
    expriedTime?: number
  }
}

export interface PendingRenewalStudentPagedResult {
  list?: PendingRenewalStudentItem[]
  total?: number
  studentCount?: number
}

export function getPendingRenewalStudentsPagedListApi(data: PendingRenewalStudentQueryParams) {
  return usePost<PendingRenewalStudentPagedResult>('/api/v1/students/pending-renewals/page', data)
}

export interface PendingRenewalReminderSendParams {
  tuitionAccountIds: string[]
}

export interface PendingRenewalReminderSendResult {
  recordId: string
  notifyCount: number
  successCount: number
  skippedCount: number
  failedCount: number
}

export interface PendingRenewalReminderRecordPageParams {
  pageRequestModel: {
    pageSize: number
    pageIndex: number
    needTotal?: boolean
    skipCount?: number
  }
}

export interface PendingRenewalReminderRecordPageItem {
  recordId: string
  templateId: string
  templateName: string
  channel: number
  channelName: string
  readCount: number
  notifyCount: number
  successCount: number
  skippedCount: number
  failedCount: number
  operatorId: string
  operatorName: string
  sendTime?: string | null
  unsentCount: number
}

export interface PendingRenewalReminderRecordPageResult {
  list: PendingRenewalReminderRecordPageItem[]
  total: number
}

export interface PendingRenewalReminderRecordDetailItem {
  itemId: string
  tuitionAccountId: string
  studentId: string
  studentName: string
  sex?: number
  avatar?: string
  phone?: string
  lessonName: string
  remainingText: string
  homeSchoolStatus: number
  homeSchoolStatusText: string
}

export interface PendingRenewalReminderRecordDetailResult {
  recordId: string
  templateId: string
  templateName: string
  channel: number
  channelName: string
  readCount: number
  notifyCount: number
  successCount: number
  skippedCount: number
  failedCount: number
  unsentCount: number
  operatorId: string
  operatorName: string
  sendTime?: string | null
  sentList: PendingRenewalReminderRecordDetailItem[]
  unsentList: PendingRenewalReminderRecordDetailItem[]
}

export function sendPendingRenewalWechatReminderApi(data: PendingRenewalReminderSendParams) {
  return usePost<PendingRenewalReminderSendResult>('/api/v1/students/pending-renewals/reminders/wechat/send', data)
}

export function pagePendingRenewalReminderRecordsApi(data: PendingRenewalReminderRecordPageParams) {
  return usePost<PendingRenewalReminderRecordPageResult>('/api/v1/students/pending-renewals/reminder-records/page', data)
}

export function getPendingRenewalReminderRecordDetailApi(data: { recordId: string }) {
  return useGet<PendingRenewalReminderRecordDetailResult>('/api/v1/students/pending-renewals/reminder-records/detail', data)
}
