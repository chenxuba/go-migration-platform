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
  status: number
  sourceType: number
  quantity?: number
  actualQuantity?: number
  remark?: string
  externalRemark?: string
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
  timetableSourceId?: string
  lessonName?: string
  teachingContent?: string
  subjectId?: string
  subjectName?: string
  teachingContentImages?: string[]
}

export interface ScheduleTeachingRecordItem {
  teachingRecordId: string
  timetableSourceId?: string
  startTime?: string
  endTime?: string
  timetableSourceType: number
  className?: string
  one2OneName?: string
  lessonName?: string
  subjectId?: string
  subjectName?: string
  rollCallStatus: number
  attendanceRate?: number
  attendCount?: number
  shouldAttendCount?: number
  actualQuantity?: number
  actualTuition?: number
  teacherName?: string
  assistants?: string
  teacherClassTime?: number
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

export function getTeachingRecordDetailApi(params: { teachingRecordId: string }) {
  return useGet<TeachingRecordDetailResult>('/api/v1/class-records/detail', params)
}

export function deleteTeachingRecordApi(data: { teachingRecordId: string }) {
  return usePost<boolean>('/api/v1/class-records/delete', data)
}
