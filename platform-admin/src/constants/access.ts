export type AccessCode = string | number
export type AccessCodeLike = AccessItem | AccessCode | Array<AccessItem | AccessCode>
export type AccessMetaType = 'system' | 'group' | 'route' | 'action'

export class AccessItem {
  readonly code: string
  readonly title: string
  readonly type: AccessMetaType
  readonly menu: string
  readonly page: string
  readonly action?: string

  constructor(code: string, title: string, options: { type: AccessMetaType, menu: string, page: string, action?: string }) {
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
  return new AccessItem(code, title, { type: 'system', menu: '系统内置', page: title })
}

function group(code: string, title: string) {
  return new AccessItem(code, title, { type: 'group', menu: title, page: title })
}

function route(code: string, menu: string, page: string) {
  return new AccessItem(code, page, { type: 'route', menu, page })
}

function action(code: string, menu: string, page: string, actionName: string) {
  return new AccessItem(code, actionName, { type: 'action', menu, page, action: actionName })
}

export const PlatformAccessEnum = {
  superAdmin: system('super:admin', '超级管理员'),

  platformControl: group('grp:plt', '平台总控'),
  platformHome: route('page:pltHome', '平台总控', '总控首页'),
  platformHomeRefresh: action('perm:pltHomeRefresh', '平台总控', '总控首页', '刷新数据'),
  platformTenant: route('page:pltTenant', '平台总控', '租户管理'),
  platformTenantAdd: action('perm:pltTenantAdd', '平台总控', '租户管理', '新增租户'),
  platformTenantEdit: action('perm:pltTenantEdit', '平台总控', '租户管理', '编辑租户'),
  platformTenantStatus: action('perm:pltTenantStatus', '平台总控', '租户管理', '启停租户'),
  platformTenantLoginConfig: action('perm:pltTenantLoginCfg', '平台总控', '租户管理', '登录配置'),
  platformTenantLoginAddress: action('perm:pltTenantLoginAddr', '平台总控', '租户管理', '生成登录地址'),

  customerManage: group('grp:cust', '客户管理'),
  customerOrg: route('page:custOrg', '客户管理', '机构列表'),
  customerOrgAdd: action('perm:custOrgAdd', '客户管理', '机构列表', '新增机构'),
  customerOrgEdit: action('perm:custOrgEdit', '客户管理', '机构列表', '编辑机构'),
  customerOrgVersionAuth: action('perm:custOrgVersionAuth', '客户管理', '机构列表', '版本授权'),
  customerOrgRenew: action('perm:custOrgRenew', '客户管理', '机构列表', '机构续期'),
  customerOrgLoginConfig: action('perm:custOrgLoginCfg', '客户管理', '机构列表', '登录配置'),
  customerGov: route('page:custGov', '客户管理', '政府账户'),
  customerGovAdd: action('perm:custGovAdd', '客户管理', '政府账户', '新增账号'),
  customerGovEdit: action('perm:custGovEdit', '客户管理', '政府账户', '编辑账号'),
  customerGovStatus: action('perm:custGovStatus', '客户管理', '政府账户', '启停账号'),

  internalManage: group('grp:intl', '内部管理'),
  internalStaff: route('page:intlStf', '内部管理', '员工管理'),
  internalStaffAdd: action('perm:intlStfAdd', '内部管理', '员工管理', '新增员工'),
  internalStaffEdit: action('perm:intlStfEdit', '内部管理', '员工管理', '编辑员工'),
  internalStaffStatus: action('perm:intlStfStatus', '内部管理', '员工管理', '启停员工'),
  internalStaffAssignRole: action('perm:intlStfAssignRole', '内部管理', '员工管理', '分配角色'),
  internalRole: route('page:intlRole', '内部管理', '角色管理'),
  internalRoleAdd: action('perm:intlRoleAdd', '内部管理', '角色管理', '新增角色'),
  internalRoleEdit: action('perm:intlRoleEdit', '内部管理', '角色管理', '编辑角色'),
  internalRoleDelete: action('perm:intlRoleDel', '内部管理', '角色管理', '删除角色'),

  systemConfig: group('grp:sys', '系统配置'),
  defaultRole: route('page:sysDefRole', '系统配置', '默认角色'),
  defaultRoleEdit: action('perm:sysDefRoleEdit', '系统配置', '默认角色', '编辑默认角色'),
  defaultRolePreview: action('perm:sysDefRolePreview', '系统配置', '默认角色', '预览权限'),
  version: route('page:sysVer', '系统配置', '版本管理'),
  versionAdd: action('perm:sysVerAdd', '系统配置', '版本管理', '新增版本'),
  versionEdit: action('perm:sysVerEdit', '系统配置', '版本管理', '编辑版本'),
  versionDelete: action('perm:sysVerDel', '系统配置', '版本管理', '删除版本'),
  versionPermissionConfig: action('perm:sysVerPermCfg', '系统配置', '版本管理', '配置权限'),
  storage: route('page:sysOss', '系统配置', '云存储配置'),
  storageEdit: action('perm:sysOssEdit', '系统配置', '云存储配置', '编辑配置'),
  storageTest: action('perm:sysOssTest', '系统配置', '云存储配置', '测试配置'),
  loginTemplate: route('page:sysLoginTpl', '系统配置', '登录页模板'),
  loginTemplateAdd: action('perm:sysLoginTplAdd', '系统配置', '登录页模板', '新增模板'),
  loginTemplateEdit: action('perm:sysLoginTplEdit', '系统配置', '登录页模板', '编辑模板'),
  loginTemplateDelete: action('perm:sysLoginTplDel', '系统配置', '登录页模板', '删除模板'),
  loginTemplatePreview: action('perm:sysLoginTplPreview', '系统配置', '登录页模板', '真实预览'),
  permission: route('page:sysPerm', '系统配置', '权限管理'),
  permissionAdd: action('perm:sysPermAdd', '系统配置', '权限管理', '新增权限'),
  permissionEdit: action('perm:sysPermEdit', '系统配置', '权限管理', '修改权限'),
  permissionDelete: action('perm:sysPermDel', '系统配置', '权限管理', '删除权限'),

  scaleConfig: group('grp:scale', '量表配置'),
  scaleManage: route('page:sysScale', '量表配置', '量表管理'),
  scaleManageAdd: action('perm:sysScaleAdd', '量表配置', '量表管理', '新增量表'),
  scaleManageAuth: action('perm:sysScaleAuth', '量表配置', '量表管理', '机构授权'),
  scaleManageIepTarget: action('perm:sysScaleIepTarget', '量表配置', '量表管理', 'IEP目标库'),
  scaleManageReference: action('perm:sysScaleReference', '量表配置', '量表管理', '引用文献'),
  scaleManageThanks: action('perm:sysScaleThanks', '量表配置', '量表管理', '特别鸣谢'),
} as const

export function normalizePlatformAccessCode(value: AccessItem | AccessCode | null | undefined) {
  if (value == null)
    return ''
  if (value instanceof AccessItem)
    return value.code
  return String(value).trim()
}

export const PlatformAccessCodeMap = Object.values(PlatformAccessEnum).reduce<Record<string, AccessItem>>((map, item) => {
  map[item.code] = item
  return map
}, {})

export function getPlatformAccessItem(value: AccessItem | AccessCode | null | undefined) {
  return PlatformAccessCodeMap[normalizePlatformAccessCode(value)]
}
