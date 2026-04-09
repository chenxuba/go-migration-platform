// GET /api/v1/inst-config
export interface InstConfig {
  actualStudentSourceTypeConfig: string
  addImportStudentRule: number
  addIntentionStudentRule: number
  allowOriginalRefund: string
  arriveClassDaysConfig: string
  arriveClassSwitch: string
  arriveClassTimesConfig: string
  auditionRecordAutomate: string
  autoAssignPerformance: string
  autoSendBirthdayMessage: string
  bookLessonLockedHours: string
  bookLessonOpeningDays: string
  bookLessonTime: string
  bookLessonTimes: string
  bookLessonTimesCycle: string
  chargeByPriceDefaultPrice: string
  classCommentParentFeedbackType: string
  createTime: string
  deductWhenLeave: string
  deductWhenTruancy: string
  defaultClassTime: string
  defaultStudentClassTime: string
  defaultTeacherClassTime: string
  discountsMode: string
  enableAdjustTuitionAccountOrder: string
  enableAdvisor: boolean
  enableArrearagedSendMessage: string
  enableAuditionSmsRemind: string
  enableAutoDeductStock: string
  enableBookLessonTimes: string
  enableByAutoTeaching: string
  enableByDateLesson: string
  enableByDateStudentAbsentRecord: string
  enableByDateStudentTeachingRecord: string
  enableByFaceAttendance: string
  enableByVoiceTips: string
  enableChargeByPrice: string
  enableChargeByPriceStudentAbsentRecord: string
  enableClassCommentParentFeedback: string
  enableCollectorStaff: boolean
  enableCompensationSendMessage: string
  enableComposeLesson: string
  enableCustomSku: string
  enableFaceAttendanceCheckInNotice: string
  enableFaceAttendanceCheckOutNotice: string
  enableFaceAttendanceRelateTeaching: string
  enableFilterHoliday: string
  enableForeground: boolean
  enableGoodsManagement: string
  enableGradeUpgrade: string
  enableLeaveApplyNumberLimit: string
  enableLeaveApplyTimeLimit: string
  enableLeaveDeductMoney: string
  enableLeftClassTimeRemind: string
  enableLiquidationRemindMessage: string
  enableOrderTagRequired: string
  enableOrgSendChildBindNoticeToAdmin: string
  enableOrgSendFaceAttendNoticeToAdmin: string
  enablePeerInfoAndServiceManagement: string
  enablePhoneSellStaff: boolean
  enablePointChangeRemindMessage: string
  enablePublicPool: boolean
  enableRechargeAccountChangeMessage: string
  enableRefundZero: string
  enableRenewClassNum: string
  enableRenewPrice: string
  enableRenewValidityDay: string
  enableSendChildBindNoticeToAdmin: string
  enableSendCouponRemindSms: string
  enableSendFaceAttendNoticeToAdmin: string
  enableSendFaceAttendNoticeToParent: string
  enableShowArrearsInformation: string
  enableShowLeftTuition: string
  enableShowRechargeAccountBalance: string
  enableShowSchoolOnOrderReceipt: string
  enableSpaceBookingNotice: string
  enableStudentManager: boolean
  enableStudentParentTranscriptChart: string
  enableSubject: string
  enableSubjectOnlineSaleFilter: string
  enableTeachingBillRemindSms: string
  enableQuickUnifiedPeriod?: boolean
  unifiedTimePeriodJson?: unknown
  enableTimetableTimeConfig: string
  enableTranOrderFinishedSendMessage: string
  enableTruantDeductMoney: string
  enableViceSellStaff: boolean
  enabledArrearsRollcall: string
  enabledBookLessonExcess: string
  enabledClassConsumptionReminder: string
  enabledClassReminder: string
  enabledOne2one: string
  enabledRenewReminder: string
  enabledShowBookLessonStudentCount: string
  faceAttendanceInterval: string
  faceAttendanceRelateRule: string
  faceAttendanceSplit: string
  id: number
  instId: number
  leaveApplyCycleLimit: string
  leaveApplyNumberLimit: string
  leaveApplyTimeLimit: string
  leaveApplyTypeLimit: string
  limitImportSameWeChat: boolean
  limitSameWeChat: boolean
  maxMicroSchoolOnlineUserCount: string
  maximumClassSizePolicy: string
  microSchoolOrderExpireMinutes: string
  renewClassNum: string
  renewPrice: string
  renewValidityDay: string
  schoolHomeBanner: string
  sendClassReminderMsgHour: string
  sendClassReminderSmsHour: string
  shouldStudentSourceTypeConfig: string
  studentAbsentClassSwitch: string
  studentAbsentClassValue: string
  timeTableChangedSendToCSwitch: string
  tuitionAccountPriority: string
  unfollowedTime: number
  updateTime: string
  uuid: string
  version: number
}
export interface GetInstConfigParams {
  effectiveDate?: string
}

export interface InstPeriodConfig {
  unifiedTimePeriodJson?: unknown
}

export interface SetInstConfigResult {
  success: boolean
  periodWeekStart?: string
  periodAppliedToday?: boolean
}

export interface InstPeriodRepairResult {
  success: boolean
  repairedVersions: number
}

export function getInstConfigApi(params?: GetInstConfigParams) {
  return useGet<InstConfig>('/api/v1/inst-config', params)
}

export function getInstPeriodConfigApi(params?: GetInstConfigParams) {
  return useGet<InstPeriodConfig>('/api/v1/inst-config/period', params)
}
// /instConfig/setInstConfig
export function setInstConfigApi(data: InstConfig) {
  return usePost<SetInstConfigResult>('/api/v1/inst-config/update', data)
}

export function previewInstPeriodEffectiveApi(data: { unifiedTimePeriodJson: unknown }) {
  return usePost<SetInstConfigResult>('/api/v1/inst-config/period-effective-preview', data)
}

export function repairInstPeriodVersionsApi() {
  return usePost<InstPeriodRepairResult>('/api/v1/inst-config/period-repair', {})
}
// 解密学员手机号
export function getStudentPhoneNumberApi(data: { studentId: number }) {
  return useGet<string>('/api/v1/intent-students/phone', data)
}
