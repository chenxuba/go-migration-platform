import axios from 'axios'
import { STORAGE_AUTHORIZE_KEY, useAuthorization } from '~/composables/authorization'
import { useGet, usePost } from '~/utils/request'

/** 对标 CheckClassName：true = 名称已存在 */
export function checkGroupClassNameApi(data: {
  name: string
  isOne2One: boolean
  /** 编辑班级时传入当前班级 id，排除自身重名 */
  exceptId?: string
}) {
  return usePost<boolean>('/api/v1/group-classes/check-name', data)
}

/** 对标 Create 集体班 */
export function createGroupClassApi(data: {
  name: string
  lessonId: string
  classroomId?: string
  maxCount: number
  teacherIds: string[]
  defaultTeacherId: string
  defaultStudentClassTime: number
  defaultTeacherClassTime: number
  defaultClassTimeRecordMode: number
  isCopyStudent?: boolean
  copiedStudents?: unknown[]
  isCopyTimetable?: boolean
  classProperties?: unknown[]
  remark?: string
}) {
  // 业务错误用 HTTP 400，需当作「有 body 的成功响应」解析，否则会进 reject 且看不到 message
  return usePost<{ id: string, name: string }>('/api/v1/group-classes/create', data, {
    validateStatus: status => (status >= 200 && status < 300) || status === 400,
  })
}

/** 对标 Class/Update */
export function updateGroupClassApi(data: {
  id: string
  name: string
  lessonId: string
  classroomId?: string
  maxCount: number
  teacherIds: string[]
  defaultTeacherId: string
  defaultStudentClassTime: number
  defaultTeacherClassTime: number
  defaultClassTimeRecordMode: number
  copyFromClassId?: string
  isCopyStudent?: boolean
  copiedStudents?: unknown[]
  isCopyTimetable?: boolean
  classProperties?: unknown[]
  remark?: string
}) {
  return usePost<{ id: string, name: string }>('/api/v1/group-classes/update', data, {
    validateStatus: status => (status >= 200 && status < 300) || status === 400,
  })
}

export function closeGroupClassApi(data: { id: string }) {
  return usePost<boolean>('/api/v1/group-classes/close', data)
}

export interface GroupClassBatchIDsParams {
  ids: string[]
}

export interface GroupClassBatchTeacherParams {
  ids: string[]
  teacherIds: string[]
}

export interface GroupClassBatchClassTimeParams {
  ids: string[]
  defaultStudentClassTime: number
  defaultTeacherClassTime: number
  defaultClassTimeRecordMode: number
}

export interface GroupClassBatchMaxCountParams {
  ids: string[]
  maxCount: number
}

export function batchCloseGroupClassesApi(data: GroupClassBatchIDsParams) {
  return usePost('/api/v1/group-classes/batch-close', data)
}

export function reopenGroupClassApi(data: { id: string }) {
  return usePost<boolean>('/api/v1/group-classes/reopen', data)
}

export function batchAssignGroupClassTeacherApi(data: GroupClassBatchTeacherParams) {
  return usePost('/api/v1/group-classes/batch-assign-class-teacher', data)
}

export function batchReplaceGroupClassTeacherApi(data: GroupClassBatchTeacherParams) {
  return usePost('/api/v1/group-classes/batch-replace-class-teacher', data)
}

export function batchUpdateGroupClassClassTimeApi(data: GroupClassBatchClassTimeParams) {
  return usePost('/api/v1/group-classes/batch-update-class-time', data)
}

export function batchUpdateGroupClassMaxCountApi(data: GroupClassBatchMaxCountParams) {
  return usePost('/api/v1/group-classes/batch-update-max-count', data)
}

/** 对标 QueryClassList */
export function pageGroupClassesApi(data: {
  queryModel: Record<string, unknown>
  pageRequestModel: {
    needTotal?: boolean
    pageSize: number
    pageIndex: number
    skipCount?: number
  }
}) {
  return usePost<{ list: GroupClassRow[], total: number }>('/api/v1/group-classes/page', data)
}

export async function exportGroupClassesApi(data: {
  queryModel?: Record<string, unknown>
}) {
  const token = useAuthorization()
  return axios.post('/api/v1/group-classes/export', data, {
    responseType: 'blob',
    headers: {
      [STORAGE_AUTHORIZE_KEY]: token.value || '',
      Authorization: token.value ? `Bearer ${token.value}` : '',
      'Accept-Language': 'zh-CN',
    },
  })
}

export function pageMoveGroupClassCandidatesApi(data: {
  queryModel: {
    currentClassId: string
    studentId: string
    lessonId: string
    className?: string
    teacherId?: string
  }
  pageRequestModel: {
    needTotal?: boolean
    pageSize: number
    pageIndex: number
    skipCount?: number
  }
}) {
  return usePost<{ list: GroupClassRow[], total: number }>('/api/v1/group-classes/move-student-candidates', data)
}

/** 对标 QueryClassStatisticsInfo（请求体与 queryModel 字段一致） */
export function groupClassStatisticsApi(queryModel: Record<string, unknown>) {
  return usePost<{
    classCount: number
    openClassCount: number
    studentCount: number
    studentPersonTime: number
  }>('/api/v1/group-classes/statistics', queryModel)
}

export interface GroupClassTeacher {
  id: string
  name: string
  /** 机构员工手机号，与 StaffSelect 右侧展示一致 */
  mobile?: string
  status: number
  avatar?: string
}

export interface GroupClassRow {
  id: string
  name: string
  classTime: number
  lessonId: string
  lessonName: string
  isMultiProduct: boolean
  studentCount: number
  lockStudentCount: number
  maxCount: number
  teachers: GroupClassTeacher[]
  defaultTeacherId: string
  defaultTeacherName: string
  classRoomName: string
  classLessonTimes: unknown[]
  isScheduled: boolean
  classLessonDayInfos: {
    lessonDayCount: number
    completeLessonDayCount: number
  }
  status: number
  closedTime: string
  createdTime: string
  createdStaffName: string
  remark: string
  classProperties: unknown[]
  defaultStudentClassTime: number
  defaultTeacherClassTime: number
  defaultClassTimeRecordMode: number
}

/** 对标 Class/Get，编辑弹窗拉详情 */
export interface GroupClassDetailVO extends GroupClassRow {
  classroomId: string
  classroomName: string
  classroomEnabled: boolean
  classroomAddressCharge: number
  teacherCount: number
  lessonType: number
  lessonScope: number
  lessonPrice: number
  defaultTeacherStatus: number
}

export function getGroupClassDetailApi(params: { id: string }) {
  return useGet<GroupClassDetailVO>('/api/v1/group-classes/detail', params)
}

export async function downloadGroupClassRollCallSheetApi(params: { classId: string }) {
  const token = useAuthorization()
  return axios.get('/api/v1/group-classes/export-roll-call-sheet', {
    params,
    responseType: 'blob',
    headers: {
      [STORAGE_AUTHORIZE_KEY]: token.value || '',
      Authorization: token.value ? `Bearer ${token.value}` : '',
      'Accept-Language': 'zh-CN',
    },
  })
}

export interface GroupClassDrawerScheduleItem {
  key: string
  classId: string
  detailScheduleId: string
  batchNo?: string
  scheduleCount: number
  completedCount: number
  type: number
  repeatRule: string
  dateRangeText: string
  timeText: string
  weekdayText: string
  teacherName: string
  assistantText: string
  classroomName: string
  lessonName: string
  batchMeta?: {
    schedulingMode?: string
    repeatRule?: string
    holidayPolicy?: string
    selectedWeekdays?: string[]
    scheduleStartDate?: string
    freeSelectedDates?: string[]
    plannedClassCount?: number
  }
}

export interface GroupClassDrawerWaitingRollCallScheduleItem {
  id: string
  batchNo?: string
  batchSize: number
  classId: string
  lessonName: string
  lessonDate: string
  startAt: string
  endAt: string
  teacherName: string
  assistantText: string
  classroomName: string
  callStatus: number
  callStatusText?: string
  canRollCall?: boolean
  rollCallDisabledReason?: string
}

export function getGroupClassDrawerSchedulesApi(data: { classId: string }) {
  return usePost<{ list: GroupClassDrawerScheduleItem[], total: number }>('/api/v1/group-classes/schedules', data)
}

export function getGroupClassDrawerWaitingRollCallSchedulesApi(data: {
  classId: string
  startDate?: string
  endDate?: string
}) {
  return usePost<{ list: GroupClassDrawerWaitingRollCallScheduleItem[], total: number }>('/api/v1/group-classes/waiting-roll-call-schedules', data)
}

export interface GroupClassOperationLogItem {
  id: string
  operateTime: string
  studentId: string
  studentName: string
  operationType: number
  operationTypeText: string
  operationContent: string
  operatorId: string
  operatorName: string
}

export interface GroupClassEntryExitRecordItem {
  id: string
  studentId: string
  studentName: string
  avatar?: string
  phone?: string
  phoneRelationship?: number
  entryExitStatus: number
  entryExitStatusText: string
  entryExitTime: string
  operatorId: string
  operatorName: string
  operateTime: string
}

export function pageGroupClassOperationLogsApi(data: {
  queryModel: {
    classId: string
    studentId?: string
    operatorId?: string
    operateStartAt?: string
    operateEndAt?: string
    operationTypes?: number[]
  }
  pageRequestModel: {
    needTotal?: boolean
    pageSize: number
    pageIndex: number
    skipCount?: number
  }
}) {
  return usePost<{ list: GroupClassOperationLogItem[], total: number }>('/api/v1/group-classes/operation-log-paged-list', data)
}

export function pageGroupClassEntryExitRecordsApi(data: {
  queryModel: {
    classId: string
    studentId?: string
    recordStartDate?: string
    recordEndDate?: string
    entryExitStatuses?: number[]
  }
  pageRequestModel: {
    needTotal?: boolean
    pageSize: number
    pageIndex: number
    skipCount?: number
  }
}) {
  return usePost<{ list: GroupClassEntryExitRecordItem[], total: number, studentCount: number }>('/api/v1/group-classes/entry-exit-record-paged-list', data)
}

export interface GroupClassStudentQueryModel {
  id?: string
  classId?: string
  status?: number[]
  ignoreSuspendedTuitionAccount?: boolean
}

export interface GroupClassStudentStatisticsVO {
  studentCount: number
  noneBindCount: number
  noneFaceCount: number
}

export interface GroupClassStudentTuitionAccountInfo {
  tuitionAccountId: string
  productName: string
  productId: string
  remainQuantity: number
  remainFreeQuantity: number
  remainTuition: number
  arrearTuition: number
  lessonChargingMode: number
  enableExpireTime: boolean
  startTime?: string
  expireTime?: string
  isTuitionAccountActive: boolean
  totalQuantity: number
  totalFreeQuantity: number
  totalTuition: number
}

export interface GroupClassStudentPagedItem {
  id: string
  name: string
  sex: number
  avatar?: string
  phone?: string
  isBind: boolean
  studentFaceInfoId?: string
  isStudentFace: boolean
  phoneRelationship?: number
  tuitionAccountId?: string
  classStudentTuitionAccountInfo?: GroupClassStudentTuitionAccountInfo
  status: number
  totalQuantity: number
  totalFreeQuantity: number
  totalTuition: number
  quantity: number
  freeQuantity: number
  tuition: number
  confirmedTuition: number
  tuitionAccountStatus: number
  enableExpireTime: boolean
  expireTime?: string
  suspendedTime?: string
  classEndingTime?: string
  advisorId?: string
  advisorName?: string
  studentManagerId?: string
  studentManagerName?: string
  customInfo?: unknown[]
  balance?: number
  point?: string
  usedClassTime?: number
  birthday?: string
  weChatNumber?: string
  grade?: string
  studySchool?: string
  address?: string
  interest?: string
  channelName?: string
  joinTime?: string
}

export interface GroupClassStudentTeachingRecordCountItem {
  studentId: string
  studentAttendCount: number
  studentLeaveCount: number
  studentTruancyCount: number
}

export function getGroupClassStudentStatisticsApi(queryModel: GroupClassStudentQueryModel) {
  return usePost<GroupClassStudentStatisticsVO>('/api/v1/group-classes/student-statistics', queryModel)
}

export function pageGroupClassStudentsApi(data: {
  queryModel: GroupClassStudentQueryModel
  pageRequestModel: {
    needTotal?: boolean
    pageSize: number
    pageIndex: number
    skipCount?: number
  }
}) {
  return usePost<{ list: GroupClassStudentPagedItem[], total: number }>('/api/v1/group-classes/student-paged-list', data)
}

export function getGroupClassFinishCoursePreviewApi(data: {
  queryModel: {
    id: string
    classStudentStatus: number[]
  }
  sortModel?: {
    orderByJoinTime?: number
    totalTuition?: number
    tuition?: number
    confirmedTuition?: number
    expireTime?: number
  }
  pageRequestModel: {
    needTotal?: boolean
    pageSize: number
    pageIndex: number
    skipCount?: number
  }
}) {
  return usePost<{ list: GroupClassStudentPagedItem[], total: number }>('/api/v1/group-classes/finish-course-preview', data)
}

export function getGroupClassStudentTeachingRecordCountApi(data: {
  studentIds: string[]
  classId: string
  studentTeachingRecordStatuses?: number[]
}) {
  return usePost<GroupClassStudentTeachingRecordCountItem[]>('/api/v1/group-classes/student-teaching-record-count', data)
}

/** 对标 Class/GetStudentListByClassIds：各班已在班学员 */
export interface GroupClassStudentInClassBucket {
  classId: string
  students: GroupClassStudentInClassItem[]
}

export interface GroupClassStudentInClassItem {
  id: string
  name: string
  avatar?: string
  phone?: string
  sex?: number
  tuitionAccountId?: string
  classId?: string
}

export function listGroupClassStudentsByClassIdsApi(data: { classIds: string[] }) {
  return usePost<GroupClassStudentInClassBucket[]>('/api/v1/group-classes/students-by-class-ids', data)
}

/** 对标 Class/BatchAssignStudents：批量将学员编入集体班 */
export function batchAssignGroupClassStudentsApi(data: {
  classIds: string[]
  students: { studentId: string, tuitionAccountId: string }[]
  enforceClassAssign?: boolean
}) {
  return usePost<{ success: boolean }>('/api/v1/group-classes/batch-assign-students', data, {
    validateStatus: status => (status >= 200 && status < 300) || status === 400,
  })
}

export function removeGroupClassStudentApi(data: {
  classId: string
  studentId: string
}) {
  return usePost<boolean>('/api/v1/group-classes/remove-student', data)
}

export function moveGroupClassStudentApi(data: {
  fromClassId: string
  toClassId: string
  studentId: string
}) {
  return usePost<boolean>('/api/v1/group-classes/move-student', data)
}
