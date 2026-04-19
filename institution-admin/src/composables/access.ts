import { toArray } from '@v-c/utils'
import { getAccessItem, normalizeAccessCode, type AccessCodeLike } from '~@/constants/access'

const READONLY_ALLOWED_ACTION_KEYWORDS = ['查看', '详情', '明细', '列表', '查询', '导出', '下载', '打印']
const READONLY_BLOCKED_ACTION_KEYWORDS = ['新增', '新建', '创建', '编辑', '修改', '删除', '移除', '管理', '导入', '分配', '转入', '转出', '认领', '设置', '保存', '审批', '审核', '提交', '发布', '上架', '下架', '启用', '禁用', '调整', '排课', '结班', '补课', '停课', '退费', '退款', '核销', '收款', '开单', '续费', '录入', '操作']

function isReadonlyBlockedAccess(code: string | number) {
  const normalizedCode = String(normalizeAccessCode(code) || '')
  if (!normalizedCode.startsWith('perm:'))
    return false
  if (normalizedCode.endsWith('Use'))
    return false

  const accessItem = getAccessItem(code)
  if (!accessItem || accessItem.type !== 'action')
    return false

  const actionText = String(accessItem.action || accessItem.title || '').trim()
  if (!actionText)
    return true
  if (READONLY_ALLOWED_ACTION_KEYWORDS.some(keyword => actionText.includes(keyword)))
    return false
  if (READONLY_BLOCKED_ACTION_KEYWORDS.some(keyword => actionText.includes(keyword)))
    return true

  return true
}

export function useAccess() {
  const userStore = useUserStore()
  const roles = computed(() => userStore.roles)
  const institutionReadonly = computed(() => userStore.institutionReadonly)
  const hasAccess = (roles: AccessCodeLike) => {
    const accessRoles = (userStore.roles ?? [])
      .map(item => normalizeAccessCode(item))
      .filter(Boolean)
    const roleArr = toArray(roles)
      .flat(1)
      .map(item => normalizeAccessCode(item))
      .filter(Boolean)
    return roleArr.some((role) => {
      if (institutionReadonly.value && isReadonlyBlockedAccess(role))
        return false
      return accessRoles?.includes(role)
    })
  }
  return {
    hasAccess,
    roles,
    institutionReadonly,
  }
}
