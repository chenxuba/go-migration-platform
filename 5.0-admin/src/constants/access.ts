export type AccessCode = string | number
export type AccessCodeLike = AccessCode | readonly AccessCode[]
export type AccessMeta = {
  type: 'system' | 'route' | 'action'
  menu: string
  page: string
  action?: string
}

export enum AccessEnum {
  // 系统内置
  structureModel = 'structureModel',
  USER = 'USER',

  // 品牌中心
  INST_ROUTE_BRAND_OFFICIAL = 'INST_ROUTE_BRAND_OFFICIAL',
  INST_ROUTE_BRAND_MINIAPP = 'INST_ROUTE_BRAND_MINIAPP',
  INST_ROUTE_BRAND_MICRO = 'INST_ROUTE_BRAND_MICRO',

  // 招生中心
  INST_ROUTE_ENROLL_SELF_TEST = 'INST_ROUTE_ENROLL_SELF_TEST',
  INST_ROUTE_ENROLL_CAMPAIGN = 'INST_ROUTE_ENROLL_CAMPAIGN',
  INST_ROUTE_ENROLL_INTENTION = 'INST_ROUTE_ENROLL_INTENTION',
  INST_ROUTE_ENROLL_FOLLOW = 'INST_ROUTE_ENROLL_FOLLOW',
  INST_ROUTE_ENROLL_TRIAL = 'INST_ROUTE_ENROLL_TRIAL',

  // 教务中心
  INST_ROUTE_EDU_SIGN = 'INST_ROUTE_EDU_SIGN',
  INST_ROUTE_EDU_STUDENT = 'INST_ROUTE_EDU_STUDENT',
  INST_ROUTE_EDU_ENROLL_LIST = 'INST_ROUTE_EDU_ENROLL_LIST',
  INST_ROUTE_EDU_CLASS = 'INST_ROUTE_EDU_CLASS',
  INST_ROUTE_EDU_ONE_TO_ONE = 'INST_ROUTE_EDU_ONE_TO_ONE',
  INST_ROUTE_EDU_TIMETABLE = 'INST_ROUTE_EDU_TIMETABLE',
  INST_ROUTE_EDU_ROLL_CALL = 'INST_ROUTE_EDU_ROLL_CALL',
  INST_ROUTE_EDU_RECORD = 'INST_ROUTE_EDU_RECORD',
  INST_ROUTE_EDU_MAKEUP = 'INST_ROUTE_EDU_MAKEUP',
  INST_ROUTE_EDU_FACE = 'INST_ROUTE_EDU_FACE',
  INST_ROUTE_EDU_COURSE = 'INST_ROUTE_EDU_COURSE',

  // 教研中心
  INST_ROUTE_TEACHER_SCALE = 'INST_ROUTE_TEACHER_SCALE',
  INST_ROUTE_TEACHER_INTERACTIVE = 'INST_ROUTE_TEACHER_INTERACTIVE',
  INST_ROUTE_TEACHER_PLAN = 'INST_ROUTE_TEACHER_PLAN',
  INST_ROUTE_TEACHER_RECORD = 'INST_ROUTE_TEACHER_RECORD',
  INST_ROUTE_TEACHER_INTERACTIVE_RECORD = 'INST_ROUTE_TEACHER_INTERACTIVE_RECORD',
  INST_ROUTE_TEACHER_HOMEWORK_RECORD = 'INST_ROUTE_TEACHER_HOMEWORK_RECORD',
  INST_ROUTE_TEACHER_RECOVERY_SUMMARY = 'INST_ROUTE_TEACHER_RECOVERY_SUMMARY',
  INST_ROUTE_TEACHER_RECOVERY_ARCHIVE = 'INST_ROUTE_TEACHER_RECOVERY_ARCHIVE',

  // 家校服务
  INST_ROUTE_HOME_RECOVERY = 'INST_ROUTE_HOME_RECOVERY',
  INST_ROUTE_HOME_HOMEWORK = 'INST_ROUTE_HOME_HOMEWORK',
  INST_ROUTE_HOME_NOTICE = 'INST_ROUTE_HOME_NOTICE',
  INST_ROUTE_HOME_LEAVE = 'INST_ROUTE_HOME_LEAVE',

  // 财务中心
  INST_ROUTE_FINANCE_ORDER = 'INST_ROUTE_FINANCE_ORDER',
  INST_ROUTE_FINANCE_APPROVAL = 'INST_ROUTE_FINANCE_APPROVAL',
  INST_ROUTE_FINANCE_DISCOUNT = 'INST_ROUTE_FINANCE_DISCOUNT',
  INST_ROUTE_FINANCE_PERFORMANCE = 'INST_ROUTE_FINANCE_PERFORMANCE',
  INST_ROUTE_FINANCE_BILL = 'INST_ROUTE_FINANCE_BILL',
  INST_ROUTE_FINANCE_PAYROLL = 'INST_ROUTE_FINANCE_PAYROLL',
  INST_ROUTE_FINANCE_INCOME_DETAIL = 'INST_ROUTE_FINANCE_INCOME_DETAIL',
  INST_ROUTE_FINANCE_TUITION_CHANGE = 'INST_ROUTE_FINANCE_TUITION_CHANGE',
  INST_ROUTE_FINANCE_RECHARGE = 'INST_ROUTE_FINANCE_RECHARGE',

  // 数据中心
  INST_ROUTE_DATA_SCREEN = 'INST_ROUTE_DATA_SCREEN',
  INST_ROUTE_DATA_ENROLL = 'INST_ROUTE_DATA_ENROLL',
  INST_ROUTE_DATA_EDU = 'INST_ROUTE_DATA_EDU',
  INST_ROUTE_DATA_HOURS = 'INST_ROUTE_DATA_HOURS',
  INST_ROUTE_DATA_HOME = 'INST_ROUTE_DATA_HOME',
  INST_ROUTE_DATA_FINANCE = 'INST_ROUTE_DATA_FINANCE',
  INST_ROUTE_DATA_REPORT = 'INST_ROUTE_DATA_REPORT',

  // 内部管理
  INST_ROUTE_INTERNAL_STAFF = 'INST_ROUTE_INTERNAL_STAFF',
  INST_ROUTE_INTERNAL_ROLE = 'INST_ROUTE_INTERNAL_ROLE',

  // 业务设置
  INST_ROUTE_SETTING_ENROLL = 'INST_ROUTE_SETTING_ENROLL',
  INST_ROUTE_SETTING_EDU = 'INST_ROUTE_SETTING_EDU',
  INST_ROUTE_SETTING_HOME = 'INST_ROUTE_SETTING_HOME',
  INST_ROUTE_SETTING_FINANCE = 'INST_ROUTE_SETTING_FINANCE',
  INST_ROUTE_SETTING_MORE = 'INST_ROUTE_SETTING_MORE',

  // 招生中心 / 意向学员
  INST_AUTH_ENROLL_INTENTION_ALL = 'INST_AUTH_ENROLL_INTENTION_ALL',
  INST_AUTH_ENROLL_INTENTION_MY = 'INST_AUTH_ENROLL_INTENTION_MY',
  INST_AUTH_ENROLL_INTENTION_DEPT = 'INST_AUTH_ENROLL_INTENTION_DEPT',
  INST_AUTH_ENROLL_INTENTION_TRANSFER_PUBLIC_POOL = 'INST_AUTH_ENROLL_INTENTION_TRANSFER_PUBLIC_POOL',

  // 教务中心 / 班级管理
  INST_AUTH_EDU_CLASS_MANAGE_WITH_STUDENTS = 'INST_AUTH_EDU_CLASS_MANAGE_WITH_STUDENTS',
  INST_AUTH_EDU_CLASS_MANAGE = 'INST_AUTH_EDU_CLASS_MANAGE',
  INST_AUTH_EDU_CLASS_ADJUST_STUDENTS = 'INST_AUTH_EDU_CLASS_ADJUST_STUDENTS',
  INST_AUTH_EDU_CLASS_MAX_COUNT = 'INST_AUTH_EDU_CLASS_MAX_COUNT',

  // 教务中心 / 课表
  INST_AUTH_EDU_TIMETABLE_ALL_CLASS_OPERATION = 'INST_AUTH_EDU_TIMETABLE_ALL_CLASS_OPERATION',
  INST_AUTH_EDU_TIMETABLE_OWN_CLASS_OPERATION = 'INST_AUTH_EDU_TIMETABLE_OWN_CLASS_OPERATION',
  INST_AUTH_EDU_TIMETABLE_CONFLICT_LIST = 'INST_AUTH_EDU_TIMETABLE_CONFLICT_LIST',
}

export const AccessMetaMap: Record<AccessEnum, AccessMeta> = {
  [AccessEnum.structureModel]: { type: 'system', menu: '系统内置', page: '结构模型' },
  [AccessEnum.USER]: { type: 'system', menu: '系统内置', page: '用户' },

  [AccessEnum.INST_ROUTE_BRAND_OFFICIAL]: { type: 'route', menu: '品牌中心', page: '专属公众号' },
  [AccessEnum.INST_ROUTE_BRAND_MINIAPP]: { type: 'route', menu: '品牌中心', page: '专属小程序' },
  [AccessEnum.INST_ROUTE_BRAND_MICRO]: { type: 'route', menu: '品牌中心', page: '微机构' },

  [AccessEnum.INST_ROUTE_ENROLL_SELF_TEST]: { type: 'route', menu: '招生中心', page: '招生自测' },
  [AccessEnum.INST_ROUTE_ENROLL_CAMPAIGN]: { type: 'route', menu: '招生中心', page: '超级裂变' },
  [AccessEnum.INST_ROUTE_ENROLL_INTENTION]: { type: 'route', menu: '招生中心', page: '意向学员' },
  [AccessEnum.INST_ROUTE_ENROLL_FOLLOW]: { type: 'route', menu: '招生中心', page: '跟进记录' },
  [AccessEnum.INST_ROUTE_ENROLL_TRIAL]: { type: 'route', menu: '招生中心', page: '试听管理' },

  [AccessEnum.INST_ROUTE_EDU_SIGN]: { type: 'route', menu: '教务中心', page: '报名续费' },
  [AccessEnum.INST_ROUTE_EDU_STUDENT]: { type: 'route', menu: '教务中心', page: '学员管理' },
  [AccessEnum.INST_ROUTE_EDU_ENROLL_LIST]: { type: 'route', menu: '教务中心', page: '报读列表' },
  [AccessEnum.INST_ROUTE_EDU_CLASS]: { type: 'route', menu: '教务中心', page: '班级管理' },
  [AccessEnum.INST_ROUTE_EDU_ONE_TO_ONE]: { type: 'route', menu: '教务中心', page: '一对一' },
  [AccessEnum.INST_ROUTE_EDU_TIMETABLE]: { type: 'route', menu: '教务中心', page: '课表' },
  [AccessEnum.INST_ROUTE_EDU_ROLL_CALL]: { type: 'route', menu: '教务中心', page: '上课点名' },
  [AccessEnum.INST_ROUTE_EDU_RECORD]: { type: 'route', menu: '教务中心', page: '上课记录' },
  [AccessEnum.INST_ROUTE_EDU_MAKEUP]: { type: 'route', menu: '教务中心', page: '补课' },
  [AccessEnum.INST_ROUTE_EDU_FACE]: { type: 'route', menu: '教务中心', page: '人脸考勤' },
  [AccessEnum.INST_ROUTE_EDU_COURSE]: { type: 'route', menu: '教务中心', page: '课程商品' },

  [AccessEnum.INST_ROUTE_TEACHER_SCALE]: { type: 'route', menu: '教研中心', page: '评估量表' },
  [AccessEnum.INST_ROUTE_TEACHER_INTERACTIVE]: { type: 'route', menu: '教研中心', page: '交互训练' },
  [AccessEnum.INST_ROUTE_TEACHER_PLAN]: { type: 'route', menu: '教研中心', page: '教案中心' },
  [AccessEnum.INST_ROUTE_TEACHER_RECORD]: { type: 'route', menu: '教研中心', page: '评估记录' },
  [AccessEnum.INST_ROUTE_TEACHER_INTERACTIVE_RECORD]: { type: 'route', menu: '教研中心', page: '交互记录' },
  [AccessEnum.INST_ROUTE_TEACHER_HOMEWORK_RECORD]: { type: 'route', menu: '教研中心', page: '作业记录' },
  [AccessEnum.INST_ROUTE_TEACHER_RECOVERY_SUMMARY]: { type: 'route', menu: '教研中心', page: '康复小结' },
  [AccessEnum.INST_ROUTE_TEACHER_RECOVERY_ARCHIVE]: { type: 'route', menu: '教研中心', page: '康复档案' },

  [AccessEnum.INST_ROUTE_HOME_RECOVERY]: { type: 'route', menu: '家校服务', page: '康复记录' },
  [AccessEnum.INST_ROUTE_HOME_HOMEWORK]: { type: 'route', menu: '家校服务', page: '课后任务' },
  [AccessEnum.INST_ROUTE_HOME_NOTICE]: { type: 'route', menu: '家校服务', page: '通知公告' },
  [AccessEnum.INST_ROUTE_HOME_LEAVE]: { type: 'route', menu: '家校服务', page: '请假管理' },

  [AccessEnum.INST_ROUTE_FINANCE_ORDER]: { type: 'route', menu: '财务中心', page: '订单管理' },
  [AccessEnum.INST_ROUTE_FINANCE_APPROVAL]: { type: 'route', menu: '财务中心', page: '审批管理' },
  [AccessEnum.INST_ROUTE_FINANCE_DISCOUNT]: { type: 'route', menu: '财务中心', page: '报名优惠' },
  [AccessEnum.INST_ROUTE_FINANCE_PERFORMANCE]: { type: 'route', menu: '财务中心', page: '业绩管理' },
  [AccessEnum.INST_ROUTE_FINANCE_BILL]: { type: 'route', menu: '财务中心', page: '账单管理' },
  [AccessEnum.INST_ROUTE_FINANCE_PAYROLL]: { type: 'route', menu: '财务中心', page: '工资管理' },
  [AccessEnum.INST_ROUTE_FINANCE_INCOME_DETAIL]: { type: 'route', menu: '财务中心', page: '确认收入明细' },
  [AccessEnum.INST_ROUTE_FINANCE_TUITION_CHANGE]: { type: 'route', menu: '财务中心', page: '学费变动记录' },
  [AccessEnum.INST_ROUTE_FINANCE_RECHARGE]: { type: 'route', menu: '财务中心', page: '储值账户' },

  [AccessEnum.INST_ROUTE_DATA_SCREEN]: { type: 'route', menu: '数据中心', page: '数据大屏' },
  [AccessEnum.INST_ROUTE_DATA_ENROLL]: { type: 'route', menu: '数据中心', page: '招生数据' },
  [AccessEnum.INST_ROUTE_DATA_EDU]: { type: 'route', menu: '数据中心', page: '教务数据' },
  [AccessEnum.INST_ROUTE_DATA_HOURS]: { type: 'route', menu: '数据中心', page: '课时统计' },
  [AccessEnum.INST_ROUTE_DATA_HOME]: { type: 'route', menu: '数据中心', page: '家校数据' },
  [AccessEnum.INST_ROUTE_DATA_FINANCE]: { type: 'route', menu: '数据中心', page: '财务数据' },
  [AccessEnum.INST_ROUTE_DATA_REPORT]: { type: 'route', menu: '数据中心', page: '报表管理' },

  [AccessEnum.INST_ROUTE_INTERNAL_STAFF]: { type: 'route', menu: '内部管理', page: '员工管理' },
  [AccessEnum.INST_ROUTE_INTERNAL_ROLE]: { type: 'route', menu: '内部管理', page: '角色管理' },

  [AccessEnum.INST_ROUTE_SETTING_ENROLL]: { type: 'route', menu: '业务设置', page: '招生设置' },
  [AccessEnum.INST_ROUTE_SETTING_EDU]: { type: 'route', menu: '业务设置', page: '教务设置' },
  [AccessEnum.INST_ROUTE_SETTING_HOME]: { type: 'route', menu: '业务设置', page: '家校设置' },
  [AccessEnum.INST_ROUTE_SETTING_FINANCE]: { type: 'route', menu: '业务设置', page: '财务设置' },
  [AccessEnum.INST_ROUTE_SETTING_MORE]: { type: 'route', menu: '业务设置', page: '更多设置' },

  [AccessEnum.INST_AUTH_ENROLL_INTENTION_ALL]: { type: 'action', menu: '招生中心', page: '意向学员', action: '查看所有的意向学员' },
  [AccessEnum.INST_AUTH_ENROLL_INTENTION_MY]: { type: 'action', menu: '招生中心', page: '意向学员', action: '仅查看我的意向学员' },
  [AccessEnum.INST_AUTH_ENROLL_INTENTION_DEPT]: { type: 'action', menu: '招生中心', page: '意向学员', action: '在PC端查看本部门及以下作为销售员的意向学员' },
  [AccessEnum.INST_AUTH_ENROLL_INTENTION_TRANSFER_PUBLIC_POOL]: { type: 'action', menu: '招生中心', page: '意向学员', action: '批量转入公有池' },

  [AccessEnum.INST_AUTH_EDU_CLASS_MANAGE_WITH_STUDENTS]: { type: 'action', menu: '教务中心', page: '班级管理', action: '新建/编辑/结班/调整班级学员' },
  [AccessEnum.INST_AUTH_EDU_CLASS_MANAGE]: { type: 'action', menu: '教务中心', page: '班级管理', action: '新建/编辑班级以及结班操作' },
  [AccessEnum.INST_AUTH_EDU_CLASS_ADJUST_STUDENTS]: { type: 'action', menu: '教务中心', page: '班级管理', action: '调整班级学员' },
  [AccessEnum.INST_AUTH_EDU_CLASS_MAX_COUNT]: { type: 'action', menu: '教务中心', page: '班级管理', action: '编辑满班人数' },

  [AccessEnum.INST_AUTH_EDU_TIMETABLE_ALL_CLASS_OPERATION]: { type: 'action', menu: '教务中心', page: '课表', action: '新建/编辑/删除日程/添加补课学员时可选择所有班级/1v1' },
  [AccessEnum.INST_AUTH_EDU_TIMETABLE_OWN_CLASS_OPERATION]: { type: 'action', menu: '教务中心', page: '课表', action: '新建/编辑/删除日程/添加补课学员时可选择自己的班级/1v1' },
  [AccessEnum.INST_AUTH_EDU_TIMETABLE_CONFLICT_LIST]: { type: 'action', menu: '教务中心', page: '课表', action: '冲突日程列表' },
}

export function getAccessLabel(code: AccessCode) {
  if (typeof code !== 'string')
    return String(code)

  const meta = AccessMetaMap[code as AccessEnum]
  if (!meta)
    return code

  return [meta.menu, meta.page, meta.action].filter(Boolean).join(' / ')
}

export const AccessGroup = {
  INST_AUTH_EDU_CLASS_MANAGE: [
    AccessEnum.INST_AUTH_EDU_CLASS_MANAGE_WITH_STUDENTS,
    AccessEnum.INST_AUTH_EDU_CLASS_MANAGE,
  ],
  INST_AUTH_EDU_CLASS_ADJUST: [
    AccessEnum.INST_AUTH_EDU_CLASS_MANAGE_WITH_STUDENTS,
    AccessEnum.INST_AUTH_EDU_CLASS_ADJUST_STUDENTS,
  ],
  INST_AUTH_EDU_TIMETABLE_MANAGE: [
    AccessEnum.INST_AUTH_EDU_TIMETABLE_ALL_CLASS_OPERATION,
    AccessEnum.INST_AUTH_EDU_TIMETABLE_OWN_CLASS_OPERATION,
  ],
} as const satisfies Record<string, readonly AccessEnum[]>
