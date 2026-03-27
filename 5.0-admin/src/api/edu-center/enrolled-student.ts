import axios from 'axios'
import { STORAGE_AUTHORIZE_KEY, useAuthorization } from '~/composables/authorization'

export interface EnrolledStudentInfo {
  id?: string
  stuName?: string
  avatarUrl?: string
  stuSex?: number
  mobile?: string
  phoneRelationship?: number
  isCollect?: boolean
  isBindChild?: boolean
  studentStatus?: number
  createTime?: string
  channelId?: string
  channelName?: string
  advisorId?: string
  advisorName?: string
  studentManagerId?: string
  studentManagerName?: string
  followUpTime?: string
  birthDay?: string
  weChatNumber?: string
  secondPhoneNumber?: string
  studySchool?: string
  grade?: string
  interest?: string
  address?: string
  recommendStudentId?: string
  recommendStudentName?: string
  remark?: string
  salesAssignedTime?: string
  salePerson?: string
  salePersonName?: string
  customInfo?: Array<{
    fieldId: number
    fieldName: string
    value: string
  }>
  isCrossSchoolStudent?: boolean
  createId?: string
  createName?: string
  followUpStatus?: number
  collectorStaffId?: string
  collectorStaffName?: string
  foregroundStaffId?: string
  foregroundStaffName?: string
  phoneSellStaffId?: string
  phoneSellStaffName?: string
  viceSellStaffStaffId?: string
  viceSellStaffStaffName?: string
  firstEnrolledTime?: string
}

export interface EnrolledStudentQueryParams {
  pageRequestModel: {
    pageSize: number
    pageIndex: number
  }
  queryModel?: {
    stuName?: string
    sexes?: number[]
    customFieldSearchList?: Array<{
      studentCustomFieldId: string
      type: number
      searchKey?: string
      searchOptions?: string[]
      searchTimeBegin?: string
      searchTimeEnd?: string
    }>
  }
}

// 获取在读学员列表
export function getEnrolledStudentListApi(data: EnrolledStudentQueryParams) {
  return usePost<EnrolledStudentInfo>('/api/v1/enrolled-students/page', data)
}

export interface ExportConditionItem {
  label: string
  value: string
}

export interface EnrolledStudentExportRecord {
  id: number
  fileName: string
  exporterName: string
  totalRows: number
  queryConditions: ExportConditionItem[]
  createdTime?: string
  expiresAt?: string
  downloadUrl?: string
}

export interface EnrolledStudentExportRequest {
  queryModel: Record<string, any>
  queryConditions: ExportConditionItem[]
}

export function exportEnrolledStudentsApi(data: EnrolledStudentExportRequest) {
  return usePost<EnrolledStudentExportRecord>('/api/v1/enrolled-students/export', data)
}

export function getEnrolledStudentExportRecordsApi() {
  return useGet<EnrolledStudentExportRecord[]>('/api/v1/enrolled-students/export-records')
}

export async function downloadEnrolledStudentExportRecordApi(recordId: number | string) {
  const token = useAuthorization()
  const response = await axios.get(`/api/v1/enrolled-students/export-records/download`, {
    params: { recordId },
    responseType: 'blob',
    headers: {
      [STORAGE_AUTHORIZE_KEY]: token.value || '',
      Authorization: token.value ? `Bearer ${token.value}` : '',
      'Accept-Language': 'zh-CN',
    },
  })
  return response
}
