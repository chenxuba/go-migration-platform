export type LoginTemplateScope = 'platform' | 'institution'

export interface LoginTemplateMeta {
  label: string
  value: string
  scope: LoginTemplateScope[]
  description: string
  layout: 'split' | 'card' | 'portal'
}

export const loginTemplateRegistry: LoginTemplateMeta[] = [
  {
    label: '商务分屏登录',
    value: 'business-split',
    scope: ['platform'],
    description: '左侧品牌宣传，右侧账号登录，适合客户子总控后台。',
    layout: 'split',
  },
  {
    label: '居中品牌卡片',
    value: 'center-card',
    scope: ['platform'],
    description: '居中卡片式登录，品牌露出集中，适合轻量管理后台。',
    layout: 'card',
  },
  {
    label: '极简企业门户',
    value: 'minimal-portal',
    scope: ['platform'],
    description: '大标题门户风格，强调企业形象和入口识别。',
    layout: 'portal',
  },
  {
    label: '教务分屏登录',
    value: 'education-split',
    scope: ['institution'],
    description: '教务业务分屏布局，适合机构端日常运营入口。',
    layout: 'split',
  },
  {
    label: '校区品牌卡片',
    value: 'campus-card',
    scope: ['institution'],
    description: '突出校区 Logo 与登录卡片，适合机构独立品牌页。',
    layout: 'card',
  },
  {
    label: '轻量门户登录',
    value: 'clean-portal',
    scope: ['institution'],
    description: '更轻的门户风格，适合多机构统一但保持品牌差异。',
    layout: 'portal',
  },
]

export function getLoginTemplates(scope: LoginTemplateScope) {
  return loginTemplateRegistry.filter(item => item.scope.includes(scope))
}

export function getLoginTemplateOptions(scope: LoginTemplateScope, includeDefault = false) {
  const options = getLoginTemplates(scope).map(item => ({
    label: item.label,
    value: item.value,
    description: item.description,
  }))
  return includeDefault ? [{ label: '跟随租户默认', value: '', description: '使用租户机构端默认登录页模板' }, ...options] : options
}

export function getLoginTemplateMeta(value?: string) {
  return loginTemplateRegistry.find(item => item.value === value)
}
