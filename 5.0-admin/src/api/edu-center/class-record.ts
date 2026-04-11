import { usePost } from '~/utils/request'

export interface ClassRecordQueryModel {
  beginStartTime?: string
  endStartTime?: string
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

export interface ScheduleTeachingRecordItem {
  teachingRecordId: string
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
