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
