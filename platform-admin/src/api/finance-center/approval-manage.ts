import { useGet, usePost } from '~/utils/request'

export interface ApprovalTemplateFlow {
  step: number
  staffIds: string[]
  staffNames?: string[]
}

export interface ApprovalTemplate {
  id: string
  enable: boolean
  type: number
  name?: string
  updatedStaffName?: string
  updatedTime?: string
  ruleJson?: string
  flowModels?: ApprovalTemplateFlow[]
}

export interface ApprovalTemplateSaveItem {
  id: number
  type: number
  enable: boolean
  ruleJson: string
  flowRequestModels: {
    step: number
    staffIds: number[]
  }[]
}

export interface ApprovalStaffSummary {
  id: string
  name: string
  phone: string
  superAdmin: boolean
  avatar: string
  status: number
  createdAt?: string
  employeeType: number
}

export interface ApprovalStaffSummaryResult {
  list: ApprovalStaffSummary[]
  total: number
}

export interface ApprovalFlowStaff {
  staffId: string
  staffName: string
  teacherStatus: number
  isApproveOperate: boolean
}

export interface ApprovalFlowStage {
  isCurrentStage: boolean
  status: number
  remark: string
  operateTime: string
  step: number
  flowStaffs: ApprovalFlowStaff[]
}

export interface ApprovalAllItem {
  id: string
  approveNumber: string
  type: number
  initiateStaffName: string
  studentName: string
  studentId: string
  studentAvatar: string
  studentPhone: string
  finishTime: string
  initiateTime: string
  status: number
  orderNumber: string
  orderId: string
  orderType: number
  approveFlows: ApprovalFlowStage[]
}

export interface ApprovalAllPagedResult {
  list: ApprovalAllItem[]
  total: number
}

export interface ApprovalDetailResult {
  approveNumber?: string
  status?: number
  initiateStaffName?: string
  finishTime?: string
  initiateTime?: string
  initiateReason?: string
  approveFlows?: ApprovalFlowStage[]
}

export interface ApprovalMyStatisticsResult {
  truntoMyApproveCount: number
  myHaveApprovedCount: number
  initiateApproveCount: number
}

export interface ApprovalOperateParams {
  id: string | number
  remark?: string
}

export function getApprovalTemplatesApi() {
  return useGet<ApprovalTemplate[]>('/api/v1/approval-templates/list')
}

export function saveApprovalTemplatesApi(data: { approveTemplateRequests: ApprovalTemplateSaveItem[] }) {
  return usePost<boolean>('/api/v1/approval-templates/save', data)
}

export function getStaffSummariesApi(data: {
  queryModel?: {
    schoolId?: string
    searchKey?: string
  }
  pageRequestModel: {
    needTotal?: boolean
    skipCount?: number
    pageSize: number
    pageIndex: number
  }
}) {
  return usePost<ApprovalStaffSummaryResult>('/api/v1/staffs/summaries', data)
}

export function getApprovalAllPagedListApi(data: {
  queryModel?: {
    approveNumber?: string
    initiateStaffId?: string | number
    orderNumber?: string
    currentApproveStaffId?: string | number
    finishStartTime?: string
    finishEndTime?: string
    initiateStartTime?: string
    initiateEndTime?: string
    studentId?: string | number
    statuses?: number[]
  }
  pageRequestModel: {
    needTotal?: boolean
    skipCount?: number
    pageSize: number
    pageIndex: number
  }
  sortModel?: {
    orderByInitiateTime?: number
    orderByFinishTime?: number
  }
}) {
  return usePost<ApprovalAllPagedResult>('/api/v1/approvals/all-paged-list', data)
}

export function getApprovalMyPagedListApi(data: {
  queryModel?: {
    approveNumber?: string
    orderNumber?: string
    currentApproveStaffId?: string | number
    finishStartTime?: string
    finishEndTime?: string
    initiateStartTime?: string
    initiateEndTime?: string
    studentId?: string | number
    statuses?: number[]
    truntoMyApprove?: boolean
    myHaveApproved?: boolean
  }
  pageRequestModel: {
    needTotal?: boolean
    skipCount?: number
    pageSize: number
    pageIndex: number
  }
  sortModel?: {
    orderByInitiateTime?: number
    orderByFinishTime?: number
  }
}) {
  return usePost<ApprovalAllPagedResult>('/api/v1/approvals/my-paged-list', data)
}

export function getApprovalMyStatisticsCountApi() {
  return useGet<ApprovalMyStatisticsResult>('/api/v1/approvals/my-approve-statistics-count')
}

export function getApprovalMyInitiatedPagedListApi(data: {
  queryModel?: {
    approveNumber?: string
    orderNumber?: string
    currentApproveStaffId?: string | number
    finishStartTime?: string
    finishEndTime?: string
    initiateStartTime?: string
    initiateEndTime?: string
    studentId?: string | number
    statuses?: number[]
  }
  pageRequestModel: {
    needTotal?: boolean
    skipCount?: number
    pageSize: number
    pageIndex: number
  }
  sortModel?: {
    orderByInitiateTime?: number
    orderByFinishTime?: number
  }
}) {
  return usePost<ApprovalAllPagedResult>('/api/v1/approvals/my-initiated-paged-list', data)
}

export function getApprovalDetailApi(params: { id: string }) {
  return useGet<ApprovalDetailResult>('/api/v1/approvals/detail', params)
}

export function approveApprovalApi(data: ApprovalOperateParams) {
  return usePost<boolean>('/api/v1/approvals/approve', data)
}

export function refuseApprovalApi(data: ApprovalOperateParams) {
  return usePost<boolean>('/api/v1/approvals/refuse', data)
}
