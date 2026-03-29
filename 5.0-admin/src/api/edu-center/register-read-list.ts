export interface RegisterReadInfo {
  tuitionAccountId?: string
  studentId?: string
  studentName?: string
  avatar?: string
  sex?: number
  phone?: string
  lessonId?: string
  lessonName?: string
  lessonType?: number // 授课方式 1-班级授课 2-1v1授课
  lessonChargingMode?: number
  type?: number // 办理类型 0-试听 1-报读 2-续费 3-转课
  totalQuantity?: number
  totalFreeQuantity?: number
  totalTuition?: number
  quantity?: number
  freeQuantity?: number
  tuition?: number
  confirmedTuition?: number
  tuitionAccountStatus?: number
  assignedClass?: boolean
  enableExpireTime?: boolean
  expireTime?: string
  planSuspendTime?: string
  planResumeTime?: string
  changeStatusTime?: string
  canTransferTuitionAccount?: boolean
  advisorStaffId?: string
  advisorStaffName?: string
  studentManagerId?: string
  studentManagerName?: string
  classTeacherList?: Array<{
    id?: string
    name?: string
  }>
  suspendedTime?: string
  classEndingTime?: string
  paidTuition?: number
  shouldTuition?: number
  arrearTuition?: number
  chargeAgainstTuition?: number
  transferredTuition?: number
  paidRemaining?: number
  hasGradeUpgrade?: boolean
  lastestTeachingRecordTime?: string
}

export interface RegisterReadQueryParams {
  pageRequestModel: {
    needTotal?: boolean
    pageSize: number
    pageIndex: number
    skipCount?: number
  }
  queryModel?: {
    fromExpireTime?: string
    toExpireTime?: string
    fromSuspendedTime?: string
    toSuspendedTime?: string
    fromClosedTime?: string
    toClosedTime?: string
    isSetExpireTime?: boolean
    assignedClass?: boolean
    studentId?: string
    lessonType?: number
    remainLessonChargingMode?: number
    fromRemainQuantity?: number
    toRemainQuantity?: number
    lessonChargingList?: number[]
    statusList?: number[]
    classTeacherId?: string
    salespersonId?: string
    classIds?: string[]
    productIds?: string[]
    isArrears?: boolean
    lastestTeachingRecordStartTime?: string
    lastestTeachingRecordEndTime?: string
  }
}

export interface RegisterReadListResult {
  totalRemainedTuition?: number
  totalConfirmedTuition?: number
  totalPaidRemainedTuition?: number
  total?: number
  studentCount?: number
  studentTutionAccounts?: RegisterReadInfo[]
}

// 获取报读列表
export function getRegisterReadListApi(data: RegisterReadQueryParams) {
  return usePost<RegisterReadListResult>('/api/v1/orders/registration-list', data)
}
