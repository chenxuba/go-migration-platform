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

const INSTITUTION_GROUP_PREFIX = 'grp:'
const INSTITUTION_ROUTE_PREFIX = 'page:'
const INSTITUTION_AUTH_PREFIX = 'perm:'

const COMPACT_PHRASES = [
  { from: ['ONE', 'TO', 'ONE'], to: 'o2o' },
  { from: ['VALUE', 'ADDED'], to: 'va' },
  { from: ['THIRD', 'PARTY'], to: 'tp' },
  { from: ['DATA', 'CENTER'], to: 'dc' },
] as const

const COMPACT_TOKEN_MAP: Record<string, string> = {
  ADDED: 'add',
  AI: 'ai',
  ALBUM: 'alb',
  ALL: 'all',
  APPROVAL: 'apv',
  ARCHIVE: 'arc',
  ASSESSMENT: 'asm',
  ASSIGN: 'asn',
  ATTR: 'atr',
  AUTO: 'ato',
  BASIC: 'bsc',
  BILL: 'bil',
  BIZ: 'biz',
  BRAND: 'brd',
  BUSY: 'bsy',
  CALL: 'cal',
  CAMPAIGN: 'camp',
  CASHIER: 'csh',
  CATEGORY: 'ctg',
  CENTER: 'ctr',
  CHANGE: 'chg',
  CHANNEL: 'chn',
  CLAIM: 'clm',
  CLEAR: 'clr',
  CLASS: 'cls',
  CLOCK: 'clk',
  COLLECTION: 'col',
  CONVERT: 'cvt',
  COUNT: 'cnt',
  COURSE: 'crs',
  CREATE: 'crt',
  DATA: 'dat',
  DELETE: 'del',
  DEPT: 'dept',
  DETAIL: 'dtl',
  DISCOUNT: 'dct',
  EDU: 'edu',
  EDIT: 'edt',
  EFFECTIVE: 'eff',
  ENROLL: 'enr',
  EXAM: 'exm',
  EXPORT: 'exp',
  FACE: 'fac',
  FEEDBACK: 'fdb',
  FINANCE: 'fin',
  FOLLOW: 'flw',
  FORM: 'frm',
  GOODS: 'gds',
  GRADE: 'grd',
  GROWTH: 'grw',
  HOME: 'home',
  HOMEWORK: 'hwk',
  HOURS: 'hrs',
  IMPORT: 'imp',
  INCOME: 'inc',
  INFO: 'info',
  INTERNAL: 'intl',
  INTENTION: 'int',
  INTERACTIVE: 'iact',
  INVENTORY: 'inv',
  LEAVE: 'lev',
  LIST: 'lst',
  LOCKED: 'lck',
  MAKEUP: 'mkp',
  MANAGE: 'mng',
  MAX: 'max',
  MICRO: 'mic',
  MINIAPP: 'mini',
  MORE: 'mor',
  MY: 'my',
  NOTICE: 'ntc',
  OFFICIAL: 'off',
  OPERATE: 'opr',
  OPERATION: 'opn',
  ORDER: 'ord',
  ORG: 'org',
  OVERVIEW: 'ovw',
  PARTY: 'pty',
  PAYROLL: 'pay',
  PERFORMANCE: 'pfm',
  PERSONAL: 'psn',
  PHONE: 'phn',
  PLAN: 'pln',
  POINT: 'pnt',
  POOL: 'pol',
  PRINT: 'prt',
  PROMOTION: 'prm',
  PUBLIC: 'pub',
  RECHARGE: 'rch',
  RECORD: 'rec',
  RECOVERY: 'rcv',
  REFUND: 'rfd',
  RELATION: 'rel',
  REPLY: 'rpy',
  REPORT: 'rpt',
  REVIEW: 'rvw',
  ROLE: 'role',
  ROLL: 'rol',
  RULE: 'rul',
  SAFE: 'saf',
  SALES: 'sls',
  SCALE: 'scl',
  SCHOOL: 'sch',
  SCORE: 'scr',
  SCREEN: 'scr',
  SELF: 'self',
  SEND: 'snd',
  SENSITIVE: 'sns',
  SETTING: 'set',
  SIGN: 'sgn',
  SMS: 'sms',
  STAFF: 'stf',
  STATUS: 'sts',
  STUDENT: 'stu',
  STUDENTS: 'stus',
  STYLE: 'sty',
  SUMMARY: 'sum',
  SUPERVISE: 'sup',
  TARGET: 'tgt',
  TEACHER: 'tch',
  TEMPLATE: 'tpl',
  TEST: 'tst',
  THIRD: 'thd',
  TIMETABLE: 'tbl',
  TRIAL: 'trl',
  TUITION: 'tui',
  UPDATE: 'upd',
  VALUE: 'val',
  VIEW: 'view',
  WARNING: 'wrn',
  WITH: 'wth',
  WRITE: 'wrt',
}

function abbreviateToken(token: string) {
  const upper = String(token || '').trim().toUpperCase()
  if (!upper)
    return ''
  if (COMPACT_TOKEN_MAP[upper])
    return COMPACT_TOKEN_MAP[upper]
  return upper.toLowerCase().slice(0, 3)
}

function compactParts(parts: string[]) {
  const normalizedParts = parts
    .map(part => String(part || '').trim().toUpperCase())
    .filter(Boolean)

  const result: string[] = []
  for (let index = 0; index < normalizedParts.length; index += 1) {
    let matched = false
    for (const phrase of COMPACT_PHRASES) {
      const current = normalizedParts.slice(index, index + phrase.from.length)
      if (current.length === phrase.from.length && current.every((item, i) => item === phrase.from[i])) {
        result.push(phrase.to)
        index += phrase.from.length - 1
        matched = true
        break
      }
    }
    if (!matched)
      result.push(abbreviateToken(normalizedParts[index]))
  }
  return result
}

function upperSnakeToCompactCamel(value: string) {
  const parts = String(value || '')
    .trim()
    .split(/[_:\-\s]+/)
    .filter(Boolean)

  return compactParts(parts).reduce((result, part, index) => {
    if (!part)
      return result
    if (index === 0)
      return `${result}${part}`
    return `${result}${part.charAt(0).toUpperCase()}${part.slice(1)}`
  }, '')
}

function normalizeCamelSuffix(value: string) {
  const raw = String(value || '').trim()
  if (!raw)
    return ''
  if (/[_:\-\s]/.test(raw))
    return upperSnakeToCompactCamel(raw)
  return raw.charAt(0).toLowerCase() + raw.slice(1)
}

export function normalizeInstitutionAccessCode(code: string | number | null | undefined) {
  if (code == null)
    return ''

  const raw = String(code).trim()
  if (!raw)
    return ''

  if (raw.startsWith(INSTITUTION_GROUP_PREFIX))
    return `${INSTITUTION_GROUP_PREFIX}${normalizeCamelSuffix(raw.slice(INSTITUTION_GROUP_PREFIX.length))}`
  if (raw.startsWith(INSTITUTION_ROUTE_PREFIX))
    return `${INSTITUTION_ROUTE_PREFIX}${normalizeCamelSuffix(raw.slice(INSTITUTION_ROUTE_PREFIX.length))}`
  if (raw.startsWith(INSTITUTION_AUTH_PREFIX))
    return `${INSTITUTION_AUTH_PREFIX}${normalizeCamelSuffix(raw.slice(INSTITUTION_AUTH_PREFIX.length))}`
  return raw
}

export function buildPageUsePermissionCode(routeCode: string | number | null | undefined) {
  const normalized = normalizeInstitutionAccessCode(routeCode)
  if (!normalized || !normalized.startsWith(INSTITUTION_ROUTE_PREFIX))
    return ''
  return `${INSTITUTION_AUTH_PREFIX}${normalized.slice(INSTITUTION_ROUTE_PREFIX.length)}Use`
}

function system(code: string, title: string) {
  return new AccessItem(code, title, {
    type: 'system',
    menu: '系统内置',
    page: title,
  })
}

function route(code: string, menu: string, page: string) {
  return new AccessItem(normalizeInstitutionAccessCode(code), page, {
    type: 'route',
    menu,
    page,
  })
}

function action(code: string, menu: string, page: string, actionName: string) {
  return new AccessItem(normalizeInstitutionAccessCode(code), actionName, {
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
  brand_official: route('page:brdOff', '品牌中心', '专属公众号'),
  brand_miniapp: route('page:brdMini', '品牌中心', '专属小程序'),
  brand_micro: route('page:brdMic', '品牌中心', '微机构'),

  // 招生中心
  enroll_self_test: route('page:enrSelfTst', '招生中心', '招生自测'),
  enroll_campaign: route('page:enrCamp', '招生中心', '超级裂变'),
  enroll_intention: route('page:enrInt', '招生中心', '意向学员'),
  enroll_follow: route('page:enrFlw', '招生中心', '跟进记录'),
  enroll_trial: route('page:enrTrl', '招生中心', '试听管理'),

  // 教务中心
  edu_sign: route('page:eduSgn', '教务中心', '报名续费'),
  edu_student: route('page:eduStu', '教务中心', '学员管理'),
  edu_enroll_list: route('page:eduEnrLst', '教务中心', '报读列表'),
  edu_class: route('page:eduCls', '教务中心', '班级管理'),
  edu_one_to_one: route('page:eduO2o', '教务中心', '一对一'),
  edu_timetable: route('page:eduTbl', '教务中心', '课表'),
  edu_roll_call: route('page:eduRolCal', '教务中心', '上课点名'),
  edu_record: route('page:eduRec', '教务中心', '上课记录'),
  edu_makeup: route('page:eduMkp', '教务中心', '补课'),
  edu_face: route('page:eduFac', '教务中心', '人脸考勤'),
  edu_course: route('page:eduCrs', '教务中心', '课程商品'),

  // 教研中心
  teacher_scale: route('page:tchScl', '教研中心', '量表库'),
  teacher_record: route('page:tchRec', '教研中心', '评估记录'),
  teacher_plan: route('page:tchPln', '教研中心', '教案中心'),
  teacher_interactive: route('page:tchIact', '教研中心', '交互训练'),
  teacher_interactive_record: route('page:tchIactRec', '教研中心', '交互记录'),
  teacher_homework_record: route('page:tchHwkRec', '教研中心', '作业记录'),
  teacher_recovery_summary: route('page:tchRcvSum', '教研中心', '康复小结'),
  teacher_recovery_archive: route('page:tchRcvArc', '教研中心', '康复档案'),

  // 家校服务
  home_recovery: route('page:homeRcv', '家校服务', '康复记录'),
  home_homework: route('page:homeHwk', '家校服务', '课后任务'),
  home_notice: route('page:homeNtc', '家校服务', '通知公告'),
  home_leave: route('page:homeLev', '家校服务', '请假管理'),

  // 财务中心
  finance_order: route('page:finOrd', '财务中心', '订单管理'),
  finance_approval: route('page:finApv', '财务中心', '审批管理'),
  finance_discount: route('page:finDct', '财务中心', '报名优惠'),
  finance_performance: route('page:finPfm', '财务中心', '业绩管理'),
  finance_bill: route('page:finBil', '财务中心', '账单管理'),
  finance_payroll: route('page:finPay', '财务中心', '工资管理'),
  finance_income_detail: route('page:finIncDtl', '财务中心', '确认收入明细'),
  finance_tuition_change: route('page:finTuiChg', '财务中心', '学费变动记录'),
  finance_recharge: route('page:finRch', '财务中心', '储值账户'),

  // 数据中心
  data_screen: route('page:datScr', '数据中心', '数据大屏'),
  data_enroll: route('page:datEnr', '数据中心', '招生数据'),
  data_edu: route('page:datEdu', '数据中心', '教务数据'),
  data_hours: route('page:datHrs', '数据中心', '课时统计'),
  data_home: route('page:datHome', '数据中心', '家校数据'),
  data_finance: route('page:datFin', '数据中心', '财务数据'),
  data_report: route('page:datRpt', '数据中心', '报表管理'),

  // 内部管理
  internal_staff: route('page:intlStf', '内部管理', '员工管理'),
  internal_role: route('page:intlRole', '内部管理', '角色管理'),

  // 业务设置
  setting_enroll: route('page:setEnr', '业务设置', '招生设置'),
  setting_edu: route('page:setEdu', '业务设置', '教务设置'),
  setting_home: route('page:setHome', '业务设置', '家校设置'),
  setting_finance: route('page:setFin', '业务设置', '财务设置'),
  setting_more: route('page:setMor', '业务设置', '更多设置'),

  // 招生中心 / 意向学员
  enroll_intention_follow_status: action('perm:enrIntFlwSts', '招生中心', '意向学员', '编辑学员跟进状态'),
  enroll_intention_view_all: action('perm:enrIntAll', '招生中心', '意向学员', '查看所有的意向学员'),
  enroll_intention_view_my: action('perm:enrIntMy', '招生中心', '意向学员', '仅查看我的意向学员'),
  enroll_intention_view_dept: action('perm:enrIntDept', '招生中心', '意向学员', '在PC端查看本部门及以下作为销售员的意向学员'),
  enroll_intention_manage: action('perm:enrIntMng', '招生中心', '意向学员', '管理意向学员'),
  enroll_intention_detail: action('perm:enrIntDtl', '招生中心', '意向学员', '意向学员详情'),
  enroll_intention_channel_edit: action('perm:enrIntChnEdt', '招生中心', '意向学员', '意向学员渠道编辑'),
  enroll_intention_import: action('perm:enrIntImp', '招生中心', '意向学员', '导入意向学员'),
  enroll_intention_export: action('perm:enrIntExp', '招生中心', '意向学员', '导出意向学员'),
  enroll_intention_assign_sales: action('perm:enrIntAsnSls', '招生中心', '意向学员', '分配销售员'),
  enroll_intention_transfer_public_pool: action('perm:enrIntTranPubPol', '招生中心', '意向学员', '批量转入公有池'),
  enroll_public_pool_setting: action('perm:enrPubPolSet', '招生中心', '公有池', '设置公有池'),
  enroll_public_pool_claim: action('perm:enrPubPolClm', '招生中心', '公有池', '批量认领'),
  enroll_public_pool_assign: action('perm:enrPubPolAsn', '招生中心', '公有池', '批量分配'),

  // 招生中心 / 跟进记录
  enroll_follow_view_all: action('perm:enrFlwAll', '招生中心', '跟进记录', '查看所有跟进记录'),
  enroll_follow_view_my: action('perm:enrFlwMy', '招生中心', '跟进记录', '仅查看我的跟进记录'),
  enroll_follow_view_dept: action('perm:enrFlwDept', '招生中心', '跟进记录', '在PC端查看本部门及以下作为销售员的跟进记录'),
  enroll_follow_edit: action('perm:enrFlwEdt', '招生中心', '跟进记录', '编辑跟进记录'),
  enroll_follow_export: action('perm:enrFlwExp', '招生中心', '跟进记录', '导出跟进记录'),

  // 业务设置 / 招生设置
  setting_enroll_intention_input: action('perm:bizSetEnrIntInput', '业务设置', '招生设置', '意向学员录入设置'),

  // 敏感数据 / 导入导出可见范围
  sensitive_export_all: action('perm:snsExpAll', '敏感数据', '导入导出可见范围', '可查看所有导出记录'),
  sensitive_export_my: action('perm:snsExpMy', '敏感数据', '导入导出可见范围', '仅查看我的导出记录'),
  sensitive_import_all: action('perm:snsImpAll', '敏感数据', '导入导出可见范围', '可查看所有导入记录'),
  sensitive_import_my: action('perm:snsImpMy', '敏感数据', '导入导出可见范围', '仅查看我的导入记录'),

  // 教务中心 / 班级管理
  edu_class_manage_with_students: action('perm:eduClsMngWthStus', '教务中心', '班级管理', '新建/编辑/结班/调整班级学员'),
  edu_class_manage: action('perm:eduClsMng', '教务中心', '班级管理', '新建/编辑班级以及结班操作'),
  edu_class_adjust_students: action('perm:eduClsAdjustStus', '教务中心', '班级管理', '调整班级学员'),
  edu_class_edit_max_count: action('perm:eduClsMaxCnt', '教务中心', '班级管理', '编辑满班人数'),

  // 教务中心 / 课表
  edu_timetable_all_class_operation: action('perm:eduTblAllClsOpn', '教务中心', '课表', '新建/编辑/删除日程/添加补课学员时可选择所有班级/1v1'),
  edu_timetable_own_class_operation: action('perm:eduTblOwnClsOpn', '教务中心', '课表', '新建/编辑/删除日程/添加补课学员时可选择自己的班级/1v1'),
  edu_timetable_conflict_list: action('perm:eduTblConfLst', '教务中心', '课表', '冲突日程列表'),
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
  return normalizeInstitutionAccessCode(value)
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
  return typeof value === 'string' ? AccessCodeMap[normalizeInstitutionAccessCode(value)] : undefined
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
