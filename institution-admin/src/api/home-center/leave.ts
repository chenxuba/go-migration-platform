import { useGet, usePost } from '~/utils/request'

export interface LeaveScheduleItem {
  scheduleId: string
  classType: number
  teachingClassId: string
  teachingClassName: string
  lessonId: string
  lessonName: string
  teacherId: string
  teacherName: string
  startTime?: string
  endTime?: string
  rosterStatusBefore: number
}

export interface LeavePagedItem {
  id: string
  studentId: string
  studentName: string
  studentAvatarUrl?: string
  studentPhone: string
  startTime?: string
  endTime?: string
  isAgent: boolean
  leaveType: number
  leaveTypeText: string
  initiateStaffName: string
  operatorName?: string
  status: number
  statusText: string
  currentApproverName: string
  approverName?: string
  applyTime?: string
}

export interface LeavePagedResult {
  list: LeavePagedItem[]
  total: number
}

export interface LeaveProcessItem {
  actionType: number
  name: string
  status: string
  actionTime?: string
  pending: boolean
  remark?: string
}

export interface LeaveDetailApproveItem {
  operatorId: string
  operatorName: string
  operatorAvatar: string
  operationDate?: string
  remark: string
  approveStatus: number
  approveStatusText: string
  actionType: number
}

export interface LeaveDetail {
  id: string
  studentId: string
  studentName: string
  studentAvatarUrl?: string
  studentPhone: string
  studentSex: number
  startTime?: string
  endTime?: string
  startDate?: string
  endDate?: string
  isAgent: boolean
  leaveType: number
  leaveTypeText: string
  reason: string
  proofMaterials: string[]
  remark: string
  status: number
  statusText: string
  initiateStaffName: string
  operatorId: string
  operatorName?: string
  operatorAvatar: string
  operationDate?: string
  currentApproverName: string
  approverName?: string
  approve?: LeaveDetailApproveItem
  approves: LeaveDetailApproveItem[]
  applyTime?: string
  schedules: LeaveScheduleItem[]
  processes: LeaveProcessItem[]
}

export interface LeaveCreateResult {
  id: string
  status: number
}

export interface LeaveDetailScheduleTeacherItem {
  teacherColor: string
  teacherId: string
  teacherDuty: number
  teacherName: string
  teacherStatus: number
}

export interface LeaveDetailScheduleMemberItem {
  memberType: number
  memberId: string
  memberName: string
  timetableId: string
}

export interface LeaveDetailScheduleItem {
  orgId: string
  schoolId: string
  schoolName: string
  id: string
  title: string
  lessonDay?: string
  lessonType: number
  isFinished: boolean
  startMinutes: number
  endMinutes: number
  remark: string
  externalRemark: string
  lessonId: string
  lessonName: string
  lessonColor: string
  mainTeacherId: string
  mainTeacherName: string
  mainTeacherColor: string
  mainTeacherStatus: number
  mainTeacherAvatar: string
  tags: string[]
  tagsString: string
  sourceType: number
  sourceId: string
  address: string | null
  addressType: number
  addressId: string
  addressName: string
  members: LeaveDetailScheduleMemberItem[]
  teachers: LeaveDetailScheduleTeacherItem[]
  repeatSpan: number
  weekDays: number
  scheduleSourceType: number
  scheduleSourceId: string
  maxStudentCount: number
  bookedStudentCount: number
  subjectId: string
  subjectName: string
  isOrgCreated: boolean
  isOpenLiveRecord: boolean
  isOpenLive: boolean
  startTime?: string
  endTime?: string
}

export interface LeaveDetailScheduleResult {
  list: LeaveDetailScheduleItem[]
  total: number
}

export function previewLeaveSchedulesApi(data: {
  studentId: string
  startTime: string
  endTime: string
}) {
  return usePost<{ list: LeaveScheduleItem[], total: number }>('/api/v1/leaves/preview-schedules', data)
}

export function createLeaveAgentApi(data: {
  studentId: string
  startTime: string
  endTime: string
  leaveType: number
  reason?: string
  proofMaterials?: string[]
  remark?: string
}) {
  return usePost<LeaveCreateResult>('/api/v1/leaves/agent', data)
}

export function getLeavePagedListApi(data: {
  pageRequestModel: {
    needTotal?: boolean
    pageSize: number
    pageIndex: number
    skipCount?: number
  }
  queryModel?: {
    studentId?: string
    applyStartTime?: string
    applyEndTime?: string
    leaveTypes?: number[]
    statuses?: number[]
  }
  sortModel?: {
    byApplyTime?: number
  }
}) {
  return usePost<LeavePagedResult>('/api/v1/leaves/paged-list', data)
}

export function getLeaveDetailApi(params: { id: string }) {
  return useGet<LeaveDetail>('/api/v1/leaves/detail', params)
}

export function getLeaveDetailSchedulesApi(data: {
  pageRequestModel: {
    needTotal?: boolean
    pageSize: number
    pageIndex: number
    skipCount?: number
  }
  queryModel: {
    studentId: string | number
    startDateTime: string
    endDateTime: string
    timeRangeSearchType?: number
  }
  sortModel?: {
    byStartDate?: number
  }
}) {
  return usePost<LeaveDetailScheduleResult>('/api/v1/leaves/detail-schedules', data)
}
