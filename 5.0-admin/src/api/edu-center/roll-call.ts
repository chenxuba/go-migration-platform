import { usePost } from '~/utils/request'
import type { TeachingScheduleItem } from './teaching-schedule'
import type { StudentLessonTuitionAccountsResult } from './one-to-one'

export interface RollCallQueryModel {
  startDate?: string
  endDate?: string
  lessonId?: string
  classroomId?: string
  classId?: string
  oneToOneId?: string
  teacherId?: string
  teacherTypes?: number[]
  scheduleTypes?: string[]
}

export interface RollCallStatisticsResult {
  todayCount: number
  allCount: number
  partialCount: number
}

export interface RollCallPagedListParams {
  queryModel: RollCallQueryModel
  pageRequestModel: {
    needTotal?: boolean
    pageSize: number
    pageIndex: number
    skipCount?: number
  }
  sortModel?: {
    byStartDate?: number
  }
}

export interface RollCallPagedListResult {
  list?: TeachingScheduleItem[]
  total?: number
}

export interface RollCallClassTimetableParams {
  id: string
  lessonDay?: string
}

export interface RollCallClassTimetableTeacher {
  teacherId: string
  teacherDuty: number
  teacherName: string
  teacherStatus: number
}

export interface RollCallClassTimetableStudent {
  sourceType: number
  sourceId: string
  studentId: string
  studentName: string
  studentAvatar?: string
  studentPhone?: string
  studentPhoneRelationshipType?: number
}

export interface RollCallClassTimetableLessonDay {
  lessonDay: string
  isFinished: boolean
  lessonDayIndex: number
  students?: RollCallClassTimetableStudent[]
  removeStudent?: RollCallClassTimetableStudent[]
  teachingRecordId?: string
}

export interface RollCallClassTimetableDetail {
  id: string
  classId: string
  className: string
  classTimes: number
  defaultStudentClassTime: number
  defaultTeacherClassTime: number
  defaultClassTimeRecordMode: number
  lessonPrice: number
  teachers?: RollCallClassTimetableTeacher[]
  addressType?: number
  addressId?: string
  addressName?: string
  lessonId: string
  lessonName: string
  lessonType?: number
  startMinutes: number
  endMinutes: number
  repeatSpan?: number
  weekDays?: number
  startDate?: string
  endDate?: string
  lessonCount?: number
  remark?: string
  externalRemark?: string
  lessonDays?: RollCallClassTimetableLessonDay[]
}

export interface RollCallClassTimetableResult {
  detail?: RollCallClassTimetableDetail
}

export interface RollCallTeachingRecordStudentListParams {
  timetableSourceId: string
  timetableSourceType: number
  classId: string
  lessonId: string
  one2OneId: string
  startDate: string
  endDate: string
  lessonDay: string
}

export interface RollCallTeachingRecordMeta {
  sourceName: string
  sourceType: number
  sourceId: string
  lessonId: string
  timetableSourceType: number
  tag: number
  timetableSourceId: string
  startTime: string
  endTime: string
  teacherClassTime: number
  classroomId: string
}

export interface RollCallTeachingRecordTeacher {
  teacherId: string
  type: number
}

export interface RollCallTeachingRecordStudent {
  studentId: string
  studentName: string
  avatar?: string
  isBindChild: boolean
  quantity: number
  paidRemaining: number
  chargingMode: number
  isTuitionAccountActive: boolean
  makeUpTeachingRecordId: string
  absentStudentType: number
  tuitionAccountId: string
  sourceType: number
  studentTeachingStatus: number
  defaultStudentTeachingStatus: number
  hasSignIn: boolean
  isCrossSchoolStudent: boolean
}

export interface RollCallTeachingRecordStudentListResult {
  data?: RollCallTeachingRecordMeta
  teachers?: RollCallTeachingRecordTeacher[]
  students?: RollCallTeachingRecordStudent[]
}

export interface RollCallStudentLeaveCountItem {
  studentId: string
  leaveCount: number
}

export interface RollCallStudentLeaveCountParams {
  studentIds: string[]
  lessonId: string
}

export interface RollCallStudentTuitionExtraInfoItem {
  studentId: string
  mutilTuition: boolean
  bestMatchProductName: string
}

export interface RollCallStudentTuitionExtraInfoParams {
  studentIds: string[]
  lessonId: string
}

export interface RollCallStudentTuitionAccountsParams {
  studentId: string
  lessonId: string
}

export function getRollCallClassTimetableApi(data: RollCallClassTimetableParams) {
  return usePost<RollCallClassTimetableResult>('/api/v1/roll-call/class-timetable', data)
}

export function getRollCallTeachingRecordStudentListApi(data: RollCallTeachingRecordStudentListParams) {
  return usePost<RollCallTeachingRecordStudentListResult>('/api/v1/roll-call/teaching-record/student-list', data)
}

export function getRollCallStudentLeaveCountApi(data: RollCallStudentLeaveCountParams) {
  return usePost<RollCallStudentLeaveCountItem[]>('/api/v1/roll-call/student-leave-count', data)
}

export function getRollCallStudentTuitionExtraInfoApi(data: RollCallStudentTuitionExtraInfoParams) {
  return usePost<RollCallStudentTuitionExtraInfoItem[]>('/api/v1/roll-call/student-tuition-extra-info', data)
}

export function getRollCallStudentTuitionAccountsApi(data: RollCallStudentTuitionAccountsParams) {
  return usePost<StudentLessonTuitionAccountsResult>('/api/v1/roll-call/student-tuition-accounts', data)
}

export function getRollCallStatisticsApi(data: { queryModel?: RollCallQueryModel }) {
  return usePost<RollCallStatisticsResult>('/api/v1/roll-call/statistics', data)
}

export function getRollCallPagedListApi(data: RollCallPagedListParams) {
  return usePost<RollCallPagedListResult>('/api/v1/roll-call/paged-list', data)
}
