/**
 * 家长关系枚举
 */
export enum ParentRelationship {
  Father = 1,
  Mother = 2,
  GrandFather = 3,
  GrandMother = 4,
  MaternalGrandfather = 5,
  MaternalGrandmother = 6,
  Other = 7,
}

export const ParentRelationshipLabel: Record<ParentRelationship, string> = {
  [ParentRelationship.Father]: '爸爸',
  [ParentRelationship.Mother]: '妈妈',
  [ParentRelationship.GrandFather]: '爷爷',
  [ParentRelationship.GrandMother]: '奶奶',
  [ParentRelationship.MaternalGrandfather]: '外公',
  [ParentRelationship.MaternalGrandmother]: '外婆',
  [ParentRelationship.Other]: '其他',
}

/**
 * 跟进状态枚举
 */
export enum FollowUpStatus {
  Pending = 0,
  InProgress = 1,
  NoAnswer = 2,
  Invited = 3,
  Audited = 4,
  Visited = 5,
  Invalid = 6,
}

export const FollowUpStatusLabel: Record<FollowUpStatus, string> = {
  [FollowUpStatus.Pending]: '待跟进',
  [FollowUpStatus.InProgress]: '跟进中',
  [FollowUpStatus.NoAnswer]: '未接听',
  [FollowUpStatus.Invited]: '已邀约',
  [FollowUpStatus.Audited]: '已试听',
  [FollowUpStatus.Visited]: '已到访',
  [FollowUpStatus.Invalid]: '已失效',
}

export const FollowUpStatusStyle: Record<FollowUpStatus, { className: string }> = {
  [FollowUpStatus.Pending]: {
    className: 'bg-#ffe6e6 text-#f33',
  },
  [FollowUpStatus.InProgress]: {
    className: 'bg-#e6f0ff text-#06f',
  },
  [FollowUpStatus.NoAnswer]: {
    className: 'bg-#ffe6e6 text-#f33',
  },
  [FollowUpStatus.Invited]: {
    className: 'bg-#fff5e6 text-#f90',
  },
  [FollowUpStatus.Audited]: {
    className: 'bg-#e6ffec text-#0c3',
  },
  [FollowUpStatus.Visited]: {
    className: 'bg-#f3e6ff text-#581893',
  },
  [FollowUpStatus.Invalid]: {
    className: 'bg-#eee text-#888',
  },
}

/**
 * 意向度枚举
 */
export enum IntentionLevel {
  High = 4,
  Medium = 3,
  Low = 2,
  Unknown = 1,
}

export const IntentionLevelLabel: Record<IntentionLevel, string> = {
  [IntentionLevel.High]: '高',
  [IntentionLevel.Medium]: '中',
  [IntentionLevel.Low]: '低',
  [IntentionLevel.Unknown]: '未知',
}

export const IntentionLevelStyle: Record<IntentionLevel, { color: string }> = {
  [IntentionLevel.High]: { color: '#f33' },
  [IntentionLevel.Medium]: { color: '#f90' },
  [IntentionLevel.Low]: { color: '#0c3' },
  [IntentionLevel.Unknown]: { color: '#d9d9d9' },
}

/**
 * 性别枚举
 */
export enum Sex {
  Male = 1,
  Female = 0,
  Unknown = 2,
}

export const SexLabel: Record<Sex, string> = {
  [Sex.Male]: '男',
  [Sex.Female]: '女',
  [Sex.Unknown]: '未知',
}

/**
 * 跟进方式枚举
 */
export enum FollowMethod {
  None = 0,
  Phone = 1,
  WeChat = 2,
  Interview = 3,
  Other = 4,
}

export const FollowMethodLabel: Record<FollowMethod, string> = {
  [FollowMethod.None]: '无',
  [FollowMethod.Phone]: '电话',
  [FollowMethod.WeChat]: '微信',
  [FollowMethod.Interview]: '面谈',
  [FollowMethod.Other]: '其他',
}

/**
 * 回访状态枚举
 */
export enum VisitStatus {
  Visited = 1,
  NotVisited = 0,
}
export const VisitStatusLabel: Record<VisitStatus, string> = {
  [VisitStatus.Visited]: '已回访',
  [VisitStatus.NotVisited]: '未回访',
}

/**
 * 学员状态枚举 意向学员 在读学员  历史学员
 */
export enum StudentStatus {
  Intention = 0,
  Reading = 1,
  History = 2,
}

export const StudentStatusLabel: Record<StudentStatus, string> = {
  [StudentStatus.Intention]: '意向学员',
  [StudentStatus.Reading]: '在读学员',
  [StudentStatus.History]: '历史学员',
}

/**
 * 授课方式枚举 班级授课 1v1授课
 */
export enum TeachingMethod {
  Class = 1,
  OneToOne = 2,
}
export const TeachingMethodLabel: Record<TeachingMethod, string> = {
  [TeachingMethod.Class]: '班级授课',
  [TeachingMethod.OneToOne]: '1v1授课',
}

/**
 * 公共是否枚举 是 否
 */
export enum IsCommonYesNo {
  Yes = 1,
  No = 0,
}
export const IsCommonYesNoLabel: Record<IsCommonYesNo, string> = {
  [IsCommonYesNo.Yes]: '是',
  [IsCommonYesNo.No]: '否',
}

/**
 * 售卖状态枚举 在售 停售
 */
export enum SellStatus {
  OnSale = 1,
  StopSale = 0,
}
export const SellStatusLabel: Record<SellStatus, string> = {
  [SellStatus.OnSale]: '在售',
  [SellStatus.StopSale]: '停售',
}
