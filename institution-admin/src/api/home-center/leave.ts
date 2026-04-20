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

export interface LeaveDetail {
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
  reason: string
  proofMaterials: string[]
  remark: string
  status: number
  statusText: string
  initiateStaffName: string
  operatorName?: string
  currentApproverName: string
  approverName?: string
  applyTime?: string
  schedules: LeaveScheduleItem[]
  processes: LeaveProcessItem[]
}

export interface LeaveCreateResult {
  id: string
  status: number
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
