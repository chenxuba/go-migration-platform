export interface TuitionAccountReadingItem {
  id?: string
  lessonId?: string
  lessonName?: string
  lessonType?: number
  totalQuantity?: number
  totalFreeQuantity?: number
  totalTuition?: number
  arrearTuition?: number
  isAdjustable?: boolean
  remainQuantity?: number
  tuition?: number
  remainFreeQuantity?: number
  enableExpireTime?: boolean
  expireTime?: string
  validDate?: string
  endDate?: string
  activedAt?: string
  assignedClass?: boolean
  status?: number
  changeStatusTime?: string
  lessonChargingMode?: number
  planSuspendTime?: string
  planResumeTime?: string
  hasGradeUpgrade?: boolean
  manualSort?: boolean
}

export interface TuitionAccountReadingListQueryParams {
  sortModel?: Record<string, any>
  queryModel: {
    studentId: string
  }
  pageRequestModel: {
    needTotal?: boolean
    pageSize: number
    pageIndex: number
    skipCount?: number
  }
}

export interface TuitionAccountReadingListResult {
  list?: TuitionAccountReadingItem[]
  total?: number
}

// 查询学生报读列表（学费账户在读列表）
export function getTuitionAccountReadingListApi(data: TuitionAccountReadingListQueryParams) {
  return usePost<TuitionAccountReadingListResult>('/api/v1/tuition-accounts/reading-list', data)
}
