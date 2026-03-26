export interface ChannelInfo {
  categoryId: number
  categoryName: string
  channelName: string
  createTime: string
  dealTransformCount: number
  dealTransformRate: number
  id: number
  invalidCount: number
  isDefault: boolean
  isDisabled: boolean
  remark: string
  uuid: string
  version: number
}

export interface StudentInfo {
  id?: number
  uuid?: string
  version?: number
  studentStatus?: number
  primaryCourseCount?: number
  stuName: string
  mobile: string
  avatar: string
  sex: number
  birthday: string
  grade: string
  studySchool: string
  interest: string
  phoneRelationship: number
  address: string
  channelId: number
  weChatNumber: string
  salespersonId: number
  remark: string
  followUpStatus: number
  customInfo: Array<{
    fieldId: number
    value: string
  }>
}

export interface FollowUpInfo {
  studentId: number
  followMethod: number
  intentLevel: number
  nextFollowUpTime: string
  followUpStatus: number
  content: string
  followImages: string
  intentCourseIds: Array<number>
}

export interface IntentionStudentImportColumn {
  key: string
  title: string
  required: boolean
  fieldType: number
  options?: string[]
}

export interface IntentionStudentImportCell {
  key: string
  title: string
  value: string
  error?: string
}

export interface IntentionStudentImportRow {
  rowNo: number
  hasError: boolean
  cells: IntentionStudentImportCell[]
}

export interface IntentionStudentImportParseResult {
  importId: string
  fileName: string
  instName: string
  columns: IntentionStudentImportColumn[]
  rows: IntentionStudentImportRow[]
  normalCount: number
  abnormalCount: number
}

export interface IntentionStudentImportTaskDetail {
  id: string
  fileName: string
  uploadStaffId: string
  uploadStaffName: string
  executeStaffId?: string
  executeStaffName?: string
  totalRows: number
  executedRows: number
  deletedRows: number
  errorRows: number
  createdTime?: string
  confirmTime?: string
  completeTime?: string
  status: number
  instName: string
}

export interface IntentionStudentImportTaskRecordListResult {
  list: IntentionStudentImportRow[]
  total: number
  columns: IntentionStudentImportColumn[]
}

export interface IntentionStudentImportTaskListResult {
  list: IntentionStudentImportTaskDetail[]
  total: number
}

export function buildIntentionStudentImportTemplateApi(data: Record<string, unknown> = {}) {
  return useGet<string>('/api/v1/intent-students/import-template', data)
}

export function parseIntentionStudentImportApi(data: FormData) {
  return usePost<IntentionStudentImportParseResult, FormData>('/api/v1/intent-students/import-parse', data, {
    headers: {
      'Content-Type': 'multipart/form-data;charset=UTF-8',
    },
  })
}

export function uploadIntentionStudentImportApi(data: FormData) {
  return usePost<{ fileUrl: string, fileName: string }, FormData>('/api/v1/intent-students/import-upload', data, {
    headers: {
      'Content-Type': 'multipart/form-data;charset=UTF-8',
    },
  })
}

export function submitIntentionStudentImportTaskApi(data: { fileUrl: string, fileName: string }) {
  return usePost<string>('/api/v1/intent-students/import-tasks/submit', data)
}

export function getIntentionStudentImportTaskDetailApi(params: { taskId: string }) {
  return useGet<IntentionStudentImportTaskDetail>('/api/v1/intent-students/import-tasks/detail', params)
}

export function getIntentionStudentImportTaskListApi() {
  return useGet<IntentionStudentImportTaskListResult>('/api/v1/intent-students/import-tasks/list')
}

export function clearIntentionStudentImportTaskListApi() {
  return usePost<boolean>('/api/v1/intent-students/import-tasks/clear')
}

export function deleteIntentionStudentImportTaskApi(data: { taskId: string }) {
  return usePost<boolean>('/api/v1/intent-students/import-tasks/delete', data)
}

export function getIntentionStudentImportTaskRecordListApi(data: {
  queryModel: { taskId: string, type: number }
  sortModel?: string
  pageRequestModel?: { needTotal?: boolean, pageSize?: number, pageIndex?: number, skipCount?: number }
}) {
  return usePost<IntentionStudentImportTaskRecordListResult>('/api/v1/intent-students/import-tasks/records', data)
}

export function batchSaveIntentionStudentImportTaskRecordsApi(data: { taskId: string, records: IntentionStudentImportRow[] }) {
  return usePost<IntentionStudentImportRow[]>('/api/v1/intent-students/import-tasks/batch-save-records', data)
}

export function startIntentionStudentImportTaskApi(data: { taskId: string }) {
  return usePost<boolean>('/api/v1/intent-students/import-tasks/start', data)
}

// 获取渠道列表
export function getChannelPCPageApi(data: ChannelInfo) {
  return usePost<ChannelInfo>('/api/v1/channels/pc-page', data)
}

// 编辑渠道
export function updateChannelApi(data: ChannelInfo) {
  return usePost<ChannelInfo>('/api/v1/channels/update', data)
}

// 创建渠道
export function createChannelApi(data: ChannelInfo) {
  return usePost<ChannelInfo>('/api/v1/channels/create', data)
}
// 启用停用
export function updateChannelStatusApi(data: ChannelInfo) {
  return usePost<ChannelInfo>('/api/v1/channels/status', data)
}
// 获取渠道分类
export function getChannelCategoryListApi(data: ChannelInfo) {
  return useGet<ChannelInfo>('/api/v1/channel-categories', data)
}
// 创建渠道分类
export function saveCategoryApi(data: ChannelInfo) {
  return usePost<ChannelInfo>('/api/v1/channel-categories/create', data)
}
// 编辑渠道分类
export function updateCategoryApi(data: ChannelInfo) {
  return usePost<ChannelInfo>('/api/v1/channel-categories/update', data)
}
// 删除分类
export function deleteCategoryApi(data: ChannelInfo) {
  return usePost<ChannelInfo>('/api/v1/channel-categories/delete', data)
}
// 批量调整渠道
export function adjustChannelApi(data: ChannelInfo) {
  return usePost<ChannelInfo>('/api/v1/channels/adjust', data)
}

export function getDefaultChannelList() {
  return useGet('/api/v1/channels/default')
}

// 获取自定义渠道列表
export function getChannelListWithChannelsApi(data: ChannelInfo) {
  return useGet<ChannelInfo>('/api/v1/channels/grouped', data)
}
// 获取意向学生列表
export function getIntentStudentListApi(data: ChannelInfo) {
  return usePost<ChannelInfo>('/api/v1/intent-students/page', data)
}
// 添加意向学员
export function addIntendedStudentApi(data: StudentInfo) {
  return usePost<StudentInfo>('/api/v1/intent-students/create', data)
}
// 校验学员重复
export function checkStudentRepeatApi(data: StudentInfo) {
  return usePost<StudentInfo>('/api/v1/intent-students/check-repeat', data)
}
// 设置跟进状态
export function updateStatusApi(data: StudentInfo) {
  return usePost<StudentInfo>('/api/v1/intent-students/status', data)
}
// 批量删除意向学员
export function batchDeleteIntendedStudentApi(data: StudentInfo) {
  return usePost<StudentInfo>('/api/v1/intent-students/delete', data)
}
// 获取意向学生详情
export function getIntentStudentDetailApi(data: StudentInfo) {
  return useGet<StudentInfo>('/api/v1/intent-students/detail', data)
}
// 获取统计跟进数量
export function getFollowUpCountApi(data: StudentInfo) {
  return useGet<StudentInfo>('/api/v1/follow-records/count', data)
}
// 获取渠道树形结构
export function getChannelTreeApi(data: StudentInfo) {
  return useGet<StudentInfo>('/api/v1/channel-tree', data)
}
// 修改意向学员
export function updateIntendedStudentApi(data: StudentInfo) {
  return usePost<StudentInfo>('/api/v1/intent-students/update', data)
}
// 获取推荐人分页
export function getRecommenderPageApi(data: StudentInfo) {
  return usePost<StudentInfo>('/api/v1/recommenders/page', data)
}
// 添加跟进记录
export function createStudentFollowUpApi(data: FollowUpInfo) {
  return usePost<FollowUpInfo>('/api/v1/follow-records/create', data)
}
// 获取跟进记录
export function getFollowUpRecordPagedApi(data: FollowUpInfo) {
  return usePost<FollowUpInfo>('/api/v1/follow-records/page', data)
}
// 修改回访状态
export function updateVisitStatusApi(data: FollowUpInfo) {
  return usePost<FollowUpInfo>('/api/v1/follow-records/visit-status', data)
}
// 统计跟进记录（与 getFollowUpCount 同后端）
export function getFollowRecordCountApi(data: FollowUpInfo) {
  return useGet<FollowUpInfo>('/api/v1/follow-records/count', data)
}
// 修改跟进记录
export function updateFollowRecordApi(data: FollowUpInfo) {
  return usePost<FollowUpInfo>('/api/v1/follow-records/update', data)
}
// 批量分配销售
export function batchAssignSalespersonApi(data: StudentInfo) {
  return usePost<StudentInfo>('/api/v1/intent-students/assign-sales', data)
}
// 批量转入公有池
export function batchTransferToPublicPoolApi(data: StudentInfo) {
  return usePost<StudentInfo>('/api/v1/intent-students/public-pool', data)
}
// 获取学生变更信息
export function listStudentChangeInfoApi(data: StudentInfo) {
  return useGet<StudentInfo>('/api/v1/students/change-records', data)
}
