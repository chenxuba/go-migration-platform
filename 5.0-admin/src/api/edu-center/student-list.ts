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
