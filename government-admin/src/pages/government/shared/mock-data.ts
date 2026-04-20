export interface GovernmentScope {
  level: '省级' | '市级' | '区级'
  regionName: string
  regionCode: string
  subordinateRegions: number
  institutionCount: number
  pendingAlerts: number
  pendingTasks: number
}

export const scopeMap: Record<GovernmentScope['level'], GovernmentScope> = {
  省级: {
    level: '省级',
    regionName: '浙江省',
    regionCode: '330000',
    subordinateRegions: 90,
    institutionCount: 218,
    pendingAlerts: 7,
    pendingTasks: 15,
  },
  市级: {
    level: '市级',
    regionName: '杭州市',
    regionCode: '330100',
    subordinateRegions: 13,
    institutionCount: 46,
    pendingAlerts: 3,
    pendingTasks: 6,
  },
  区级: {
    level: '区级',
    regionName: '西湖区',
    regionCode: '330106',
    subordinateRegions: 0,
    institutionCount: 12,
    pendingAlerts: 1,
    pendingTasks: 2,
  },
}

export const regionalSummary = [
  {
    key: '330100',
    regionName: '杭州市',
    level: '市级',
    institutionCount: 46,
    ongoingTasks: 8,
    alerts: 3,
    completionRate: '93%',
  },
  {
    key: '330200',
    regionName: '宁波市',
    level: '市级',
    institutionCount: 39,
    ongoingTasks: 5,
    alerts: 1,
    completionRate: '97%',
  },
  {
    key: '330300',
    regionName: '温州市',
    level: '市级',
    institutionCount: 31,
    ongoingTasks: 6,
    alerts: 2,
    completionRate: '91%',
  },
  {
    key: '330400',
    regionName: '嘉兴市',
    level: '市级',
    institutionCount: 24,
    ongoingTasks: 3,
    alerts: 1,
    completionRate: '96%',
  },
]

export const institutionList = [
  {
    key: '1',
    institutionName: '西湖童康康复中心',
    regionName: '杭州市 / 西湖区',
    level: '区级',
    institutionType: '民办康复机构',
    riskLevel: '低风险',
    status: '正常营业',
    latestInspectionAt: '2026-04-18',
    principal: '王敏',
    phone: '0571-88118811',
  },
  {
    key: '2',
    institutionName: '上城启航儿童康复站',
    regionName: '杭州市 / 上城区',
    level: '区级',
    institutionType: '社区康复点',
    riskLevel: '中风险',
    status: '限期整改',
    latestInspectionAt: '2026-04-15',
    principal: '赵宁',
    phone: '0571-88990012',
  },
  {
    key: '3',
    institutionName: '宁波甬爱康复门诊',
    regionName: '宁波市 / 鄞州区',
    level: '市级',
    institutionType: '医疗康复机构',
    riskLevel: '低风险',
    status: '正常营业',
    latestInspectionAt: '2026-04-11',
    principal: '林达',
    phone: '0574-66778899',
  },
  {
    key: '4',
    institutionName: '温州星语发育支持中心',
    regionName: '温州市 / 鹿城区',
    level: '市级',
    institutionType: '民办康复机构',
    riskLevel: '高风险',
    status: '暂停接诊',
    latestInspectionAt: '2026-04-09',
    principal: '陈悦',
    phone: '0577-66886688',
  },
]

export const supervisionTasks = [
  {
    key: 'task-1',
    taskName: '2026 年二季度机构安全巡查',
    taskType: '专项督导',
    regionName: '浙江省',
    status: '进行中',
    owner: '省级监管专班',
    dueDate: '2026-04-30',
    completionRate: '68%',
  },
  {
    key: 'task-2',
    taskName: '杭州市康复档案规范抽查',
    taskType: '日常抽查',
    regionName: '杭州市',
    status: '待下发',
    owner: '杭州市监管局',
    dueDate: '2026-04-25',
    completionRate: '0%',
  },
  {
    key: 'task-3',
    taskName: '西湖区消防整改复核',
    taskType: '整改复核',
    regionName: '西湖区',
    status: '待复核',
    owner: '西湖区监管员',
    dueDate: '2026-04-22',
    completionRate: '90%',
  },
]

export const accountRoles = [
  {
    key: 'role-1',
    roleName: '省级监管员',
    dataScope: '全省机构与下辖市区',
    typicalPermissions: '查看省级总览、下发督导任务、跨市统计分析',
  },
  {
    key: 'role-2',
    roleName: '市级监管员',
    dataScope: '本市及下辖区机构',
    typicalPermissions: '查看市级统计、督办区级整改、审核本市任务',
  },
  {
    key: 'role-3',
    roleName: '区级监管员',
    dataScope: '本区机构',
    typicalPermissions: '机构巡查、整改跟踪、日常预警处理',
  },
]
