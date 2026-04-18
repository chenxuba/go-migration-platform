export type AccessCode = string | number

export type AccessMetaType = 'system' | 'route' | 'action'

export class AccessItem {
  readonly code: string
  readonly title: string
  readonly type: AccessMetaType
  readonly menu: string
  readonly page: string
  readonly action?: string

  constructor(
    code: string,
    title: string,
    options: {
      type: AccessMetaType
      menu: string
      page: string
      action?: string
    },
  ) {
    this.code = code
    this.title = title
    this.type = options.type
    this.menu = options.menu
    this.page = options.page
    this.action = options.action
  }

  get label() {
    return [this.menu, this.page, this.action].filter(Boolean).join(' / ')
  }

  toString() {
    return this.code
  }
}

function system(code: string, title: string) {
  return new AccessItem(code, title, {
    type: 'system',
    menu: '系统内置',
    page: title,
  })
}

function route(code: string, menu: string, page: string) {
  return new AccessItem(code, page, {
    type: 'route',
    menu,
    page,
  })
}

function action(code: string, menu: string, page: string, actionName: string) {
  return new AccessItem(code, actionName, {
    type: 'action',
    menu,
    page,
    action: actionName,
  })
}

export const AccessEnum = {
  // 系统内置
  system_structure_model: system('structureModel', '结构模型'),
  system_user: system('USER', '用户'),

  // 品牌中心
  brand_official: route('INST_ROUTE_BRAND_OFFICIAL', '品牌中心', '专属公众号'),
  brand_miniapp: route('INST_ROUTE_BRAND_MINIAPP', '品牌中心', '专属小程序'),
  brand_micro: route('INST_ROUTE_BRAND_MICRO', '品牌中心', '微机构'),

  // 招生中心
  enroll_self_test: route('INST_ROUTE_ENROLL_SELF_TEST', '招生中心', '招生自测'),
  enroll_campaign: route('INST_ROUTE_ENROLL_CAMPAIGN', '招生中心', '超级裂变'),
  enroll_intention: route('INST_ROUTE_ENROLL_INTENTION', '招生中心', '意向学员'),
  enroll_follow: route('INST_ROUTE_ENROLL_FOLLOW', '招生中心', '跟进记录'),
  enroll_trial: route('INST_ROUTE_ENROLL_TRIAL', '招生中心', '试听管理'),

  // 教务中心
  edu_sign: route('INST_ROUTE_EDU_SIGN', '教务中心', '报名续费'),
  edu_student: route('INST_ROUTE_EDU_STUDENT', '教务中心', '学员管理'),
  edu_enroll_list: route('INST_ROUTE_EDU_ENROLL_LIST', '教务中心', '报读列表'),
  edu_class: route('INST_ROUTE_EDU_CLASS', '教务中心', '班级管理'),
  edu_one_to_one: route('INST_ROUTE_EDU_ONE_TO_ONE', '教务中心', '一对一'),
  edu_timetable: route('INST_ROUTE_EDU_TIMETABLE', '教务中心', '课表'),
  edu_roll_call: route('INST_ROUTE_EDU_ROLL_CALL', '教务中心', '上课点名'),
  edu_record: route('INST_ROUTE_EDU_RECORD', '教务中心', '上课记录'),
  edu_makeup: route('INST_ROUTE_EDU_MAKEUP', '教务中心', '补课'),
  edu_face: route('INST_ROUTE_EDU_FACE', '教务中心', '人脸考勤'),
  edu_course: route('INST_ROUTE_EDU_COURSE', '教务中心', '课程商品'),

  // 教研中心
  teacher_scale: route('INST_ROUTE_TEACHER_SCALE', '教研中心', '评估量表'),
  teacher_interactive: route('INST_ROUTE_TEACHER_INTERACTIVE', '教研中心', '交互训练'),
  teacher_plan: route('INST_ROUTE_TEACHER_PLAN', '教研中心', '教案中心'),
  teacher_record: route('INST_ROUTE_TEACHER_RECORD', '教研中心', '评估记录'),
  teacher_interactive_record: route('INST_ROUTE_TEACHER_INTERACTIVE_RECORD', '教研中心', '交互记录'),
  teacher_homework_record: route('INST_ROUTE_TEACHER_HOMEWORK_RECORD', '教研中心', '作业记录'),
  teacher_recovery_summary: route('INST_ROUTE_TEACHER_RECOVERY_SUMMARY', '教研中心', '康复小结'),
  teacher_recovery_archive: route('INST_ROUTE_TEACHER_RECOVERY_ARCHIVE', '教研中心', '康复档案'),

  // 家校服务
  home_recovery: route('INST_ROUTE_HOME_RECOVERY', '家校服务', '康复记录'),
  home_homework: route('INST_ROUTE_HOME_HOMEWORK', '家校服务', '课后任务'),
  home_notice: route('INST_ROUTE_HOME_NOTICE', '家校服务', '通知公告'),
  home_leave: route('INST_ROUTE_HOME_LEAVE', '家校服务', '请假管理'),

  // 财务中心
  finance_order: route('INST_ROUTE_FINANCE_ORDER', '财务中心', '订单管理'),
  finance_approval: route('INST_ROUTE_FINANCE_APPROVAL', '财务中心', '审批管理'),
  finance_discount: route('INST_ROUTE_FINANCE_DISCOUNT', '财务中心', '报名优惠'),
  finance_performance: route('INST_ROUTE_FINANCE_PERFORMANCE', '财务中心', '业绩管理'),
  finance_bill: route('INST_ROUTE_FINANCE_BILL', '财务中心', '账单管理'),
  finance_payroll: route('INST_ROUTE_FINANCE_PAYROLL', '财务中心', '工资管理'),
  finance_income_detail: route('INST_ROUTE_FINANCE_INCOME_DETAIL', '财务中心', '确认收入明细'),
  finance_tuition_change: route('INST_ROUTE_FINANCE_TUITION_CHANGE', '财务中心', '学费变动记录'),
  finance_recharge: route('INST_ROUTE_FINANCE_RECHARGE', '财务中心', '储值账户'),

  // 数据中心
  data_screen: route('INST_ROUTE_DATA_SCREEN', '数据中心', '数据大屏'),
  data_enroll: route('INST_ROUTE_DATA_ENROLL', '数据中心', '招生数据'),
  data_edu: route('INST_ROUTE_DATA_EDU', '数据中心', '教务数据'),
  data_hours: route('INST_ROUTE_DATA_HOURS', '数据中心', '课时统计'),
  data_home: route('INST_ROUTE_DATA_HOME', '数据中心', '家校数据'),
  data_finance: route('INST_ROUTE_DATA_FINANCE', '数据中心', '财务数据'),
  data_report: route('INST_ROUTE_DATA_REPORT', '数据中心', '报表管理'),

  // 内部管理
  internal_staff: route('INST_ROUTE_INTERNAL_STAFF', '内部管理', '员工管理'),
  internal_role: route('INST_ROUTE_INTERNAL_ROLE', '内部管理', '角色管理'),

  // 业务设置
  setting_enroll: route('INST_ROUTE_SETTING_ENROLL', '业务设置', '招生设置'),
  setting_edu: route('INST_ROUTE_SETTING_EDU', '业务设置', '教务设置'),
  setting_home: route('INST_ROUTE_SETTING_HOME', '业务设置', '家校设置'),
  setting_finance: route('INST_ROUTE_SETTING_FINANCE', '业务设置', '财务设置'),
  setting_more: route('INST_ROUTE_SETTING_MORE', '业务设置', '更多设置'),

  // 招生中心 / 意向学员
  enroll_intention_follow_status: action('INST_AUTH_ENROLL_INTENTION_FOLLOW_STATUS', '招生中心', '意向学员', '编辑学员跟进状态'),
  enroll_intention_view_all: action('INST_AUTH_ENROLL_INTENTION_ALL', '招生中心', '意向学员', '查看所有的意向学员'),
  enroll_intention_view_my: action('INST_AUTH_ENROLL_INTENTION_MY', '招生中心', '意向学员', '仅查看我的意向学员'),
  enroll_intention_view_dept: action('INST_AUTH_ENROLL_INTENTION_DEPT', '招生中心', '意向学员', '在PC端查看本部门及以下作为销售员的意向学员'),
  enroll_intention_manage: action('INST_AUTH_ENROLL_INTENTION_MANAGE', '招生中心', '意向学员', '管理意向学员'),
  enroll_intention_detail: action('INST_AUTH_ENROLL_INTENTION_DETAIL', '招生中心', '意向学员', '意向学员详情'),
  enroll_intention_channel_edit: action('INST_AUTH_ENROLL_INTENTION_CHANNEL_EDIT', '招生中心', '意向学员', '意向学员渠道编辑'),
  enroll_intention_import: action('INST_AUTH_ENROLL_INTENTION_IMPORT', '招生中心', '意向学员', '导入意向学员'),
  enroll_intention_export: action('INST_AUTH_ENROLL_INTENTION_EXPORT', '招生中心', '意向学员', '导出意向学员'),
  enroll_intention_assign_sales: action('INST_AUTH_ENROLL_INTENTION_ASSIGN_SALES', '招生中心', '意向学员', '分配销售员'),
  enroll_intention_transfer_public_pool: action('INST_AUTH_ENROLL_INTENTION_TRANSFER_PUBLIC_POOL', '招生中心', '意向学员', '批量转入公有池'),

  // 敏感数据 / 导入导出可见范围
  sensitive_export_all: action('INST_AUTH_SENSITIVE_EXPORT_ALL', '敏感数据', '导入导出可见范围', '可查看所有导出记录'),
  sensitive_export_my: action('INST_AUTH_SENSITIVE_EXPORT_MY', '敏感数据', '导入导出可见范围', '仅查看我的导出记录'),
  sensitive_import_all: action('INST_AUTH_SENSITIVE_IMPORT_ALL', '敏感数据', '导入导出可见范围', '可查看所有导入记录'),
  sensitive_import_my: action('INST_AUTH_SENSITIVE_IMPORT_MY', '敏感数据', '导入导出可见范围', '仅查看我的导入记录'),

  // 教务中心 / 班级管理
  edu_class_manage_with_students: action('INST_AUTH_EDU_CLASS_MANAGE_WITH_STUDENTS', '教务中心', '班级管理', '新建/编辑/结班/调整班级学员'),
  edu_class_manage: action('INST_AUTH_EDU_CLASS_MANAGE', '教务中心', '班级管理', '新建/编辑班级以及结班操作'),
  edu_class_adjust_students: action('INST_AUTH_EDU_CLASS_ADJUST_STUDENTS', '教务中心', '班级管理', '调整班级学员'),
  edu_class_edit_max_count: action('INST_AUTH_EDU_CLASS_MAX_COUNT', '教务中心', '班级管理', '编辑满班人数'),

  // 教务中心 / 课表
  edu_timetable_all_class_operation: action('INST_AUTH_EDU_TIMETABLE_ALL_CLASS_OPERATION', '教务中心', '课表', '新建/编辑/删除日程/添加补课学员时可选择所有班级/1v1'),
  edu_timetable_own_class_operation: action('INST_AUTH_EDU_TIMETABLE_OWN_CLASS_OPERATION', '教务中心', '课表', '新建/编辑/删除日程/添加补课学员时可选择自己的班级/1v1'),
  edu_timetable_conflict_list: action('INST_AUTH_EDU_TIMETABLE_CONFLICT_LIST', '教务中心', '课表', '冲突日程列表'),
} as const

export type AccessItemValue = (typeof AccessEnum)[keyof typeof AccessEnum]
export type AccessValueLike = AccessCode | AccessItemValue
export type AccessCodeLike = AccessValueLike | readonly AccessValueLike[]

export function isAccessItem(value: unknown): value is AccessItem {
  return value instanceof AccessItem
}

export function normalizeAccessCode(value: AccessValueLike | null | undefined): AccessCode | '' {
  if (value == null)
    return ''
  if (isAccessItem(value))
    return value.code
  return value
}

export const AccessCodeMap = Object.values(AccessEnum).reduce<Record<string, AccessItem>>((items, item) => {
  items[item.code] = item
  return items
}, {})

export function getAccessItem(value: AccessValueLike | null | undefined) {
  if (value == null)
    return undefined
  if (isAccessItem(value))
    return value
  return typeof value === 'string' ? AccessCodeMap[value] : undefined
}

export function getAccessLabel(value: AccessValueLike | null | undefined) {
  const item = getAccessItem(value)
  if (!item)
    return String(normalizeAccessCode(value) || '')
  return item.label || item.title
}

export const AccessGroup = {
  edu_class_manage: [
    AccessEnum.edu_class_manage_with_students,
    AccessEnum.edu_class_manage,
  ],
  edu_class_adjust: [
    AccessEnum.edu_class_manage_with_students,
    AccessEnum.edu_class_adjust_students,
  ],
  edu_timetable_manage: [
    AccessEnum.edu_timetable_all_class_operation,
    AccessEnum.edu_timetable_own_class_operation,
  ],
} as const satisfies Record<string, readonly AccessItemValue[]>
