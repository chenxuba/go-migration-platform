import axios from 'axios'
import { STORAGE_AUTHORIZE_KEY, useAuthorization } from '~/composables/authorization'
import { useGet, usePost } from '~/utils/request'

export interface ClassRecordQueryModel {
  beginStartTime?: string
  endStartTime?: string
  beginCreateTime?: string
  endCreateTime?: string
  beginUpdatedTime?: string
  endUpdatedTime?: string
  studentId?: string
  teacherIds?: string[]
  assistantTeacherIds?: string[]
  one2OneIds?: string[]
  timetableSourceTypes?: number[]
  studentSourceTypes?: number[]
  lessonChargingModeEnums?: number[]
  studentTeachingRecordStatuses?: number[]
  scheduleCallStatus?: number
  isArrear?: boolean | null
  lessonIds?: string[]
  classIds?: string[]
}

export interface StudentTeachingRecordItem {
  studentTeachingRecordId: string
  teachingRecordId: string
  studentId: string
  studentName: string
  studentPhone?: string
  avatar?: string
  teacherName?: string
  teacherEmployeeType?: number
  assistants?: string
  className?: string
  one2OneName?: string
  lessonName?: string
  status: number
  sourceType: number
  startTime?: string
  endTime?: string
  teachingRecordCreatedTime?: string
  timetableSourceType: number
  updatedTime?: string
  updatedStaffName?: string
  recordTime?: string
  quantity?: number
  actualQuantity?: number
  amount?: number
  skuMode?: number
  actualDeduct?: number
  actualTuition?: number
  arrearQuantity?: number
  remark?: string
  externalRemark?: string
  tuitionAccountId?: string
  tuitionAccountName?: string
  hasCompensated?: boolean
  subjectId?: string
  subjectName?: string
  advisorStaffId?: string
  advisorStaffName?: string
  studentManagerId?: string
  studentManagerName?: string
  teachingContent?: string
  teachingContentImages?: string[]
  classRoomName?: string
  one2OneTeachers?: string
  classTeachers?: string
  rollCallClassTeachers?: string
  currentClassTeachers?: string
}

export interface StudentTeachingRecordPagedResult {
  totalClassTimes?: number
  totalTuition?: number
  totalStudentCount?: number
  list?: StudentTeachingRecordItem[]
  total?: number
}

export interface TeachingRecordDetailTeacher {
  teacherId: string
  teacherName: string
  type: number
  status: number
  quantity: number
}

export interface TeachingRecordDetailStudent {
  studentTeachingRecordId: string
  studentId: string
  studentName: string
  studentPhone?: string
  avatar?: string
  sex?: number
  birthday?: string
  status: number
  sourceType: number
  quantity?: number
  actualQuantity?: number
  remark?: string
  externalRemark?: string
  isComment?: boolean
  isParentRead?: boolean | null
  tuitionAccountId?: string
  tuitionAccountName?: string
  isTuitionAccountActive?: boolean
  leftQuantity?: number
  skuMode?: number
  amount?: number
  actualDeduct?: number
  actualTuition?: number
  arrearQuantity?: number
  recordTime?: string
  updatedTime?: string
  updatedStaffName?: string
}

export interface TeachingRecordDetailResult {
  teachingRecordId: string
  sourceName?: string
  sourceType?: number
  sourceId?: string
  lessonId?: string
  lessonType?: number
  startTime?: string
  endTime?: string
  shouldAttendanceCount?: number
  actualAttendanceCount?: number
  leaveCount?: number
  truancyCount?: number
  teacherClassTime?: number
  studentTotalClassTime?: number
  studentActualTuition?: number
  teacherList?: TeachingRecordDetailTeacher[]
  studentList?: TeachingRecordDetailStudent[]
  createdTime?: string
  createdStaffName?: string
  timetableSourceType?: number
  classRoomName?: string
  classRoomId?: string
  defaultStudentClassTime?: number
  timetableSourceId?: string
  lessonName?: string
  teachingContent?: string
  subjectId?: string
  subjectName?: string
  teachingContentImages?: string[]
}

export interface UpdateStudentTeachingRecordParams {
  studentTeachingRecordId?: string
  teachingRecordId?: string
  studentId?: string
  sourceType?: number
  status: number
  quantity: number
  remark?: string
  externalRemark?: string
}

export interface UpdateTeachingRecordClassInfoParams {
  teachingRecordId: string
  teacherId: string
  assistantIds?: string[]
  classRoomId?: string
  teacherClassTime: number
}

export interface DeleteStudentTeachingRecordParams {
  studentTeachingRecordId: string
}

export interface ScheduleTeachingRecordItem {
  teachingRecordId: string
  timetableSourceId?: string
  startTime?: string
  endTime?: string
  timetableSourceType: number
  sourceName?: string
  sourceType?: number
  sourceId?: string
  lessonId?: string
  className?: string
  one2OneName?: string
  lessonName?: string
  subjectId?: string
  subjectName?: string
  rollCallStatus: number
  attendanceRate?: number
  attendCount?: number
  shouldAttendCount?: number
  leaveCount?: number
  absentCount?: number
  unrecordedCount?: number
  actualQuantity?: number
  actualTuition?: number
  teacherId?: string
  teacherName?: string
  assistants?: string
  classRoomName?: string
  teacherClassTime?: number
  commentCount?: number
  unCommentCount?: number
  readCount?: number
  unReadCount?: number
  createdTime?: string
  updatedTime?: string
}

export interface ScheduleTeachingRecordPagedResult {
  totalClassTimes?: number
  totalTeacherTimes?: number
  totalTuition?: number
  list?: ScheduleTeachingRecordItem[]
  total?: number
}

export interface ClassCommentQueryModel {
  teachingStartTime?: string
  teachingEndTime?: string
  teachingRecordTypes?: number[]
  lessonId?: string
  teacherIds?: string[]
  classId?: string
  one2OneId?: string
  classTeacherIds?: string[]
  one2OneTeacherIds?: string[]
}

export interface ClassCommentItem {
  teachingRecordId: string
  sourceName?: string
  sourceType?: number
  sourceId?: string
  lessonId?: string
  lessonName?: string
  createdTime?: string
  teacherId?: string
  teacherName?: string
  startTime?: string
  endTime?: string
  readCount?: number
  unReadCount?: number
  commentCount?: number
  unCommentCount?: number
  assistants?: string
  classRoomName?: string
}

export interface ClassCommentPagedParams {
  queryModel: ClassCommentQueryModel
  pageRequestModel: {
    needTotal?: boolean
    pageSize: number
    pageIndex: number
    skipCount?: number
  }
  sortModel?: {
    startTime?: number
  }
}

export interface ClassCommentPagedResult {
  list?: ClassCommentItem[]
  total?: number
}

export interface ClassCommentStudentQueryModel {
  isParentFeedback?: boolean
  teachingStartTime?: string
  teachingEndTime?: string
  teachingRecordTypes?: number[]
  lessonId?: string
  teacherIds?: string[]
  assistantTeacherIds?: string[]
  classId?: string
  one2OneId?: string
  isComment?: boolean
  isRead?: boolean
  classTeacherIds?: string[]
  one2OneTeacherIds?: string[]
  studentId?: string
}

export interface ClassCommentStudentItem {
  teachingRecordId: string
  studentTeachingRecordId: string
  sourceName?: string
  sourceType?: number
  sourceId?: string
  lessonId?: string
  lessonName?: string
  teacherId?: string
  teacherName?: string
  startTime?: string
  endTime?: string
  avatar?: string
  studentName?: string
  studentId?: string
  studentPhone?: string
  isComment?: boolean
  isRead?: boolean | null
  assistants?: string
  classRoomName?: string
  isParentFeedback?: boolean
  parentFeedbackType?: number
  parentFeedbackGrade?: number
  parentFeedbackContent?: string
}

export interface ClassCommentStudentPagedParams {
  queryModel: ClassCommentStudentQueryModel
  pageRequestModel: {
    needTotal?: boolean
    pageSize: number
    pageIndex: number
    skipCount?: number
  }
  sortModel?: {
    startTime?: number
  }
}

export interface ClassCommentStudentPagedResult {
  list?: ClassCommentStudentItem[]
  total?: number
}

export interface ClassRecordExportConditionItem {
  label: string
  value: string
}

export interface ClassRecordExportRecord {
  id: number
  exportType: string
  fileName: string
  exporterName: string
  totalRows: number
  queryConditions: ClassRecordExportConditionItem[]
  createdTime?: string
  expiresAt?: string
  downloadUrl?: string
}

export interface ClassRecordPagedParams {
  queryModel: ClassRecordQueryModel
  pageRequestModel: {
    needTotal?: boolean
    pageSize: number
    pageIndex: number
    skipCount?: number
  }
  sortModel?: {
    startTime?: number
    updatedTime?: number
  }
}

export function getStudentTeachingRecordPagedListApi(data: ClassRecordPagedParams) {
  return usePost<StudentTeachingRecordPagedResult>('/api/v1/class-records/student-paged-list', data)
}

export function getScheduleTeachingRecordPagedListApi(data: ClassRecordPagedParams) {
  return usePost<ScheduleTeachingRecordPagedResult>('/api/v1/class-records/schedule-paged-list', data)
}

export function getClassCommentPagedListApi(data: ClassCommentPagedParams) {
  return usePost<ClassCommentPagedResult>('/api/v1/class-comments/paged-list', data)
}

export function getClassCommentStudentPagedListApi(data: ClassCommentStudentPagedParams) {
  return usePost<ClassCommentStudentPagedResult>('/api/v1/class-comments/student-paged-list', data)
}

export function getTeachingRecordDetailApi(params: { teachingRecordId: string }) {
  return useGet<TeachingRecordDetailResult>('/api/v1/class-records/detail', params)
}

export function updateTeachingRecordClassInfoApi(data: UpdateTeachingRecordClassInfoParams) {
  return usePost<boolean>('/api/v1/class-records/update-class-info', data)
}

export function exportClassRecordsApi(data: {
  exportType: string
  recordIds: string[]
  queryConditions: ClassRecordExportConditionItem[]
}) {
  return usePost<ClassRecordExportRecord>('/api/v1/class-records/export', data)
}

export function getClassRecordExportRecordsApi(exportType: string) {
  return useGet<ClassRecordExportRecord[]>('/api/v1/class-records/export-records', { exportType })
}

export async function downloadClassRecordExportRecordApi(recordId: number | string, exportType: string) {
  const token = useAuthorization()
  const response = await axios.get('/api/v1/class-records/export-records/download', {
    params: { recordId, exportType },
    responseType: 'blob',
    headers: {
      [STORAGE_AUTHORIZE_KEY]: token.value || '',
      Authorization: token.value ? `Bearer ${token.value}` : '',
      'Accept-Language': 'zh-CN',
    },
  })
  return response
}

export function updateStudentTeachingRecordApi(data: UpdateStudentTeachingRecordParams) {
  return usePost<boolean>('/api/v1/class-records/student/update', data)
}

export function deleteStudentTeachingRecordApi(data: DeleteStudentTeachingRecordParams) {
  return usePost<boolean>('/api/v1/class-records/student/delete', data)
}

export function deleteTeachingRecordApi(data: { teachingRecordId: string }) {
  return usePost<boolean>('/api/v1/class-records/delete', data)
}
