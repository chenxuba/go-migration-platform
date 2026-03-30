import { usePost } from '~/utils/request'

export interface LessonIncomeTeacher {
  id?: string
  name?: string
}

export interface LessonIncomeItem {
  id: string
  studentId: string
  studentName: string
  studentPhone?: string
  studentAvatar?: string
  teachingCourseId?: string
  teachingCourseName?: string
  lessonId: string
  lessonName: string
  lessonType?: number
  teachingMethod?: number
  sourceType: number
  lessonDay?: string
  startMinutes?: number
  endMinutes?: number
  teachingTime?: string
  rollCallTime?: string
  quantity: number
  lessonChargingMode?: number
  tuition: number
  createdTime?: string
  teachers?: LessonIncomeTeacher[]
  teacherName?: string
  assistantTeachers?: LessonIncomeTeacher[]
  assistantName?: string
  productCategoryId?: string
  productCategoryName?: string
  classId?: string
  className?: string
  conformIncomeTime?: string
  teachingRecordId?: string
}

export interface LessonIncomePagedResult {
  list?: LessonIncomeItem[]
  total?: number
}

export interface LessonIncomeStatistics {
  totalCount?: number
  totalTuition?: number
}

export interface LessonIncomeQueryModel {
  startDate?: string
  endDate?: string
  sourceTypes?: number[]
  studentId?: string
  staffId?: string
  lessonId?: string
  lessonDayStartDate?: string
  lessonDayEndDate?: string
  classId?: string
  productCategoryId?: string
  conformIncomeTimeStartDate?: string
  conformIncomeTimeEndDate?: string
}

export interface LessonIncomeQueryParams {
  queryModel: LessonIncomeQueryModel
  sortModel?: {
    orderByCreatedTime?: number
  }
  pageRequestModel: {
    needTotal?: boolean
    pageSize: number
    pageIndex: number
    skipCount?: number
  }
}

export function getLessonIncomePagedListApi(data: LessonIncomeQueryParams) {
  return usePost<LessonIncomePagedResult>('/api/v1/lesson-incomes/query-paged-list', data)
}

export function getLessonIncomeStatisticsApi(data: LessonIncomeQueryModel) {
  return usePost<LessonIncomeStatistics>('/api/v1/lesson-incomes/statistics', data)
}
