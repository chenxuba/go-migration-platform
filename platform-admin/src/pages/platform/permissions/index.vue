<template>
  <div class="permission-management">
    <a-card title="权限管理" :bordered="false">
      <div class="search-bar">
        <a-space>
          <a-segmented
            v-model:value="currentPortal"
            :options="portalOptions"
            @change="handlePortalChange"
          />
          <a-input
            v-model:value="keywordInput"
            placeholder="请输入菜单名称或标识"
            allow-clear
            style="width: 260px"
            @pressEnter="handleSearch"
          />
          <a-button type="primary" @click="handleSearch">
            搜索
          </a-button>
          <a-button @click="handleReset">
            重置
          </a-button>
          <a-button v-if="hasPermission(AccessEnum.menuPermissions_add)" type="primary" @click="handleAddRoot">
            <template #icon>
              <PlusOutlined />
            </template>
            新增
          </a-button>
          <a-button @click="toggleExpandAll">
            展开 / 折叠
          </a-button>
        </a-space>
      </div>

      <a-spin :spinning="loading">
        <div class="permission-tree-panel">
          <div
            v-if="tableData.length"
            ref="treeBodyRef"
            class="permission-tree-body"
          >
            <div class="permission-tree-header">
              <span>菜单名称</span>
              <span>权限标识</span>
              <span>层级</span>
              <span>排序</span>
              <span>权重</span>
              <span>描述</span>
              <span class="permission-tree-header__actions">操作</span>
            </div>

            <div v-for="row in flatRows" :key="row.record.id" class="permission-tree-node">
              <div class="permission-tree-row">
                <div class="permission-tree-row__name">
                  <div class="permission-tree-name-cell" :style="getIndentStyle(row.depth)">
                    <button
                      v-if="row.hasChildren"
                      type="button"
                      class="permission-tree-switcher"
                      :class="{ 'is-expanded': row.isExpanded }"
                      @click.stop="toggleRowExpand(row.record)"
                    >
                      <RightOutlined />
                    </button>
                    <span v-else class="permission-tree-switcher permission-tree-switcher--placeholder" />
                    <span class="menu-name-text" :title="row.record.menuName || '--'">
                      {{ row.record.menuName || '--' }}
                    </span>
                  </div>
                </div>

                <div class="permission-tree-row__code">
                  <span class="ellipsis-text" :title="row.record.displayMenuCode">
                    {{ row.record.displayMenuCode }}
                  </span>
                </div>

                <div class="permission-tree-row__level">
                  <a-tag :color="row.record.levelTagColor">
                    {{ row.record.levelLabel }}
                  </a-tag>
                </div>

                <div class="permission-tree-row__sort">
                  {{ row.record.sort ?? '--' }}
                </div>

                <div class="permission-tree-row__weight">
                  {{ row.record.weight ?? '--' }}
                </div>

                <div class="permission-tree-row__intro">
                  <span class="ellipsis-text" :title="row.record.introduceText">
                    {{ row.record.introduceText }}
                  </span>
                </div>

                <div class="permission-tree-row__actions" @click.stop>
                  <a-button
                    v-if="hasPermission(AccessEnum.menuPermissions_update)"
                    type="link"
                    size="small"
                    @click.stop="openDrawer('edit', row.record)"
                  >
                    修改
                  </a-button>
                  <a-button
                    v-if="hasPermission(AccessEnum.menuPermissions_add) && Number(row.record.depth || 1) < 3"
                    type="link"
                    size="small"
                    @click.stop="handleAddChild(row.record)"
                  >
                    新增
                  </a-button>
                  <a-popconfirm
                    v-if="hasPermission(AccessEnum.menuPermissions_delete)"
                    title="确定要删除该权限吗？"
                    @confirm="handleDeleteByRecord(row.record)"
                  >
                    <a-button type="link" size="small" danger @click.stop>
                      删除
                    </a-button>
                  </a-popconfirm>
                </div>
              </div>
            </div>
          </div>

          <a-empty v-else description="暂无权限数据" />
        </div>
      </a-spin>
    </a-card>

    <PlatformModalShell
      v-model:open="drawerVisible"
      :width="900"
      :title="formMode === 'edit' ? '编辑权限' : '新增权限'"
      modal-class="permission-form-modal"
      scrollable
      @close="closeFormModal"
    >
      <div class="permission-form-card">
        <a-form :model="formData" layout="vertical" class="permission-form">
          <div class="permission-form-grid">
            <div class="permission-form-grid__item">
              <a-form-item label="权限名称" required>
                <a-input v-model:value="formData.menuName" placeholder="请输入权限名称" />
              </a-form-item>
            </div>

            <div class="permission-form-grid__item">
              <a-form-item label="权限编码" required>
                <a-input
                  v-model:value="formData.menuCode"
                  :disabled="formMode === 'edit'"
                  placeholder="例如 systemModel:menuPermissions"
                />
              </a-form-item>
            </div>

            <div class="permission-form-grid__item">
              <a-form-item label="排序">
                <a-input-number v-model:value="formData.sort" :min="0" style="width: 100%" />
              </a-form-item>
            </div>

            <div class="permission-form-grid__item">
              <a-form-item label="权重">
                <a-input-number v-model:value="formData.weight" :min="0" style="width: 100%" />
              </a-form-item>
            </div>

            <div class="permission-form-grid__item">
              <a-form-item label="父级菜单">
                <a-input :value="parentName" disabled />
              </a-form-item>
            </div>

            <div class="permission-form-grid__item">
              <a-form-item label="所属端口">
                <a-input :value="currentPortal === PortalEnum.PLATFORM ? '平台端' : '机构端'" disabled />
              </a-form-item>
            </div>

            <div class="permission-form-grid__item permission-form-grid__item--full">
              <a-form-item label="备注">
                <a-input v-model:value="formData.remark" placeholder="请输入备注" />
              </a-form-item>
            </div>

            <div class="permission-form-grid__item permission-form-grid__item--full">
              <a-form-item label="权限描述">
                <a-textarea
                  v-model:value="formData.introduce"
                  :rows="3"
                  placeholder="请输入权限描述"
                />
              </a-form-item>
            </div>
          </div>
        </a-form>
      </div>

      <template #footer>
        <div class="permission-form-modal__footer">
          <a-popconfirm
            v-if="formMode === 'edit' && hasPermission(AccessEnum.menuPermissions_delete)"
            title="确定要删除该权限吗？"
            @confirm="handleDelete"
          >
            <a-button danger ghost>
              删除
            </a-button>
          </a-popconfirm>
          <a-button @click="closeFormModal">
            取消
          </a-button>
          <a-button
            v-if="hasPermission(formMode === 'edit' ? AccessEnum.menuPermissions_update : AccessEnum.menuPermissions_add)"
            type="primary"
            @click="handleSubmit"
          >
            保存
          </a-button>
        </div>
      </template>
    </PlatformModalShell>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, shallowRef, watch } from 'vue'
import { PlusOutlined, RightOutlined } from '@ant-design/icons-vue'
import {
  createPermissionApi,
  deletePermissionApi,
  getPermissionTreeApi,
  updatePermissionApi,
  type PermissionMenuItem,
  type PermissionMutationPayload,
} from '@/api/platform/permissions'
import messageService from '@/utils/messageService'
import { AccessEnum } from '~@/utils/constant'
import PlatformModalShell from '../shared/platform-modal-shell.vue'

enum PortalEnum {
  PLATFORM = 0,
  INSTITUTION = 2,
}

interface PermissionRecord extends PermissionMenuItem {
  children?: PermissionRecord[]
  key: number
  depth: number
  displayMenuCode: string
  levelLabel: string
  levelTagColor: string
  introduceText: string
  searchText: string
}

type FormMode = 'add' | 'edit' | null
interface FlatPermissionRow {
  record: PermissionRecord
  depth: number
  hasChildren: boolean
  isExpanded: boolean
}

const { hasAccess } = useAccess()

const hasPermission = (permissionCode: string) => {
  return hasAccess([permissionCode, AccessEnum.superAdmin])
}

const treeData = shallowRef<PermissionRecord[]>([])
const nodeMap = shallowRef<Map<number, PermissionRecord>>(new Map())
const keywordInput = ref('')
const searchKeyword = ref('')
const expandedRowKeys = ref<number[]>([])
const loading = ref(false)
const treeBodyRef = ref<HTMLDivElement | null>(null)

const currentPortal = ref<PortalEnum>(PortalEnum.INSTITUTION)
const portalOptions = [
  { label: '平台端', value: PortalEnum.PLATFORM, disabled: true },
  { label: '机构端', value: PortalEnum.INSTITUTION },
]

const formMode = ref<FormMode>(null)
const drawerVisible = ref(false)
const parentName = ref<string>('根菜单')

const formData = reactive<PermissionMutationPayload>({
  id: undefined,
  pid: 0,
  menuName: '',
  menuCode: '',
  sort: 0,
  weight: 0,
  remark: '',
  introduce: '',
  ownType: PortalEnum.INSTITUTION,
})

const compactPhrases = [
  { from: ['ONE', 'TO', 'ONE'], to: 'o2o' },
  { from: ['VALUE', 'ADDED'], to: 'va' },
  { from: ['THIRD', 'PARTY'], to: 'tp' },
  { from: ['DATA', 'CENTER'], to: 'dc' },
]

const compactTokenMap: Record<string, string> = {
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

const abbreviateToken = (token: string) => {
  const upper = String(token || '').trim().toUpperCase()
  if (!upper)
    return ''
  if (compactTokenMap[upper])
    return compactTokenMap[upper]
  return upper.toLowerCase().slice(0, 3)
}

const compactParts = (parts: string[]) => {
  const normalizedParts = parts
    .map(part => String(part || '').trim().toUpperCase())
    .filter(Boolean)

  const result: string[] = []
  for (let index = 0; index < normalizedParts.length; index += 1) {
    let matched = false
    for (const phrase of compactPhrases) {
      const current = normalizedParts.slice(index, index + phrase.from.length)
      if (current.length === phrase.from.length && current.every((item, currentIndex) => item === phrase.from[currentIndex])) {
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

const toCompactCamel = (value: string) => {
  return compactParts(
    String(value || '')
      .trim()
      .split(/[_:\-\s]+/)
      .filter(Boolean),
  ).reduce((result, part, index) => {
    if (!part)
      return result
    if (index === 0)
      return `${result}${part}`
    return `${result}${part.charAt(0).toUpperCase()}${part.slice(1)}`
  }, '')
}

const normalizeCamelSuffix = (value: string) => {
  const raw = String(value || '').trim()
  if (!raw)
    return ''
  if (/[_:\-\s]/.test(raw))
    return toCompactCamel(raw)
  return raw.charAt(0).toLowerCase() + raw.slice(1)
}

const normalizePermissionCode = (code: string | number | null | undefined) => {
  const raw = String(code ?? '').trim()
  if (!raw)
    return ''
  if (raw.startsWith('g:'))
    return `g:${normalizeCamelSuffix(raw.slice('g:'.length))}`
  if (raw.startsWith('r:'))
    return `r:${normalizeCamelSuffix(raw.slice('r:'.length))}`
  if (raw.startsWith('a:'))
    return `a:${normalizeCamelSuffix(raw.slice('a:'.length))}`
  if (raw.startsWith('INST_GROUP_'))
    return `g:${toCompactCamel(raw.slice('INST_GROUP_'.length))}`
  if (raw.startsWith('INST_ROUTE_'))
    return `r:${toCompactCamel(raw.slice('INST_ROUTE_'.length))}`
  if (raw.startsWith('INST_AUTH_'))
    return `a:${toCompactCamel(raw.slice('INST_AUTH_'.length))}`
  return raw
}

const normalizeTree = (list: PermissionMenuItem[] = [], depth = 1): PermissionRecord[] => {
  return list.map((item) => {
    const children = Array.isArray(item.children) ? normalizeTree(item.children, depth + 1) : []
    const displayMenuCode = normalizePermissionCode(item.menuCode) || '--'
    const introduceText = String(item.introduce || item.remark || '').trim() || '--'
    const levelLabel = depth <= 1 ? '1级菜单' : depth === 2 ? '2级界面' : '3级权限'
    const levelTagColor = depth <= 1 ? 'blue' : depth === 2 ? 'green' : 'orange'
    const menuName = String(item.menuName || '')
    return {
      ...item,
      key: Number(item.id),
      menuCode: displayMenuCode,
      sort: Number(item.sort || 0),
      weight: Number(item.weight || 0),
      pid: Number(item.pid || 0),
      ownType: Number(item.ownType || currentPortal.value),
      depth,
      displayMenuCode,
      levelLabel,
      levelTagColor,
      introduceText,
      searchText: `${menuName.toLowerCase()} ${displayMenuCode.toLowerCase()}`,
      children,
    }
  })
}

const tableData = computed<PermissionRecord[]>(() => {
  const rawKeyword = searchKeyword.value.trim()
  if (!rawKeyword)
    return treeData.value

  const keyword = rawKeyword.toLowerCase()
  const normalizedKeyword = normalizePermissionCode(rawKeyword).toLowerCase()
  const keywordSet = new Set([keyword, normalizedKeyword].filter(Boolean))

  const filterTree = (list: PermissionRecord[]): PermissionRecord[] => {
    const result: PermissionRecord[] = []
    list.forEach((item) => {
      const children = item.children ? filterTree(item.children) : []
      const hit
        = Array.from(keywordSet).some(currentKeyword =>
          item.searchText.includes(currentKeyword))

      if (hit || children.length) {
        result.push({
          ...item,
          children,
        })
      }
    })
    return result
  }

  return filterTree(treeData.value)
})

const collectExpandableIds = (list: PermissionRecord[]): number[] => {
  const ids: number[] = []
  list.forEach((item) => {
    if (item.children?.length) {
      ids.push(Number(item.id))
      ids.push(...collectExpandableIds(item.children))
    }
  })
  return ids
}

const expandableRowKeys = computed<number[]>(() => collectExpandableIds(tableData.value))
const expandedKeySet = computed(() => new Set(expandedRowKeys.value.map(key => Number(key))))
const flatRows = computed<FlatPermissionRow[]>(() => {
  const rows: FlatPermissionRow[] = []

  const walk = (list: PermissionRecord[]) => {
    list.forEach((item) => {
      const hasChildren = Boolean(item.children?.length)
      const isExpanded = expandedKeySet.value.has(Number(item.id))

      rows.push({
        record: item,
        depth: item.depth,
        hasChildren,
        isExpanded,
      })

      if (hasChildren && isExpanded)
        walk(item.children!)
    })
  }

  walk(tableData.value)
  return rows
})

const buildNodeMap = (list: PermissionRecord[]) => {
  const map = new Map<number, PermissionRecord>()
  const travel = (nodes: PermissionRecord[]) => {
    nodes.forEach((item) => {
      map.set(Number(item.id), item)
      if (item.children?.length)
        travel(item.children)
    })
  }
  travel(list)
  return map
}

const resetTreeScroll = () => {
  if (treeBodyRef.value)
    treeBodyRef.value.scrollTop = 0
}

const loadTree = async () => {
  loading.value = true
  try {
    const res = await getPermissionTreeApi({ ownType: currentPortal.value })
    if (res.code !== 200) {
      messageService.error(res.message || '加载权限树失败')
      return
    }
    const normalizedTree = normalizeTree(Array.isArray(res.result) ? res.result : [])
    treeData.value = normalizedTree
    nodeMap.value = buildNodeMap(normalizedTree)
    resetTreeScroll()
  }
  catch (error: any) {
    messageService.error(error?.message || '加载权限树失败')
  }
  finally {
    loading.value = false
  }
}

const handlePortalChange = async (value: number) => {
  currentPortal.value = value as PortalEnum
  keywordInput.value = ''
  searchKeyword.value = ''
  expandedRowKeys.value = []
  await loadTree()
}

const resetForm = () => {
  formData.id = undefined
  formData.pid = 0
  formData.menuName = ''
  formData.menuCode = ''
  formData.sort = 0
  formData.weight = 0
  formData.remark = ''
  formData.introduce = ''
  formData.ownType = currentPortal.value
  parentName.value = '根菜单'
}

const handleAddRoot = () => {
  resetForm()
  formMode.value = 'add'
  formData.pid = 0
  parentName.value = '根菜单'
  drawerVisible.value = true
}

const handleAddChild = (parent?: PermissionRecord) => {
  const target = parent
  if (!target) {
    messageService.warning('请先选择父级节点')
    return
  }

  resetForm()
  formMode.value = 'add'
  formData.pid = Number(target.id)
  parentName.value = target.menuName
  drawerVisible.value = true
}

const openDrawer = (mode: FormMode, node?: PermissionRecord) => {
  if (mode === 'edit') {
    if (!node) {
      messageService.warning('请先选择要编辑的权限')
      return
    }

    formData.id = Number(node.id)
    formData.pid = Number(node.pid || 0)
    if (Number(node.pid || 0) === 0) {
      parentName.value = '根菜单'
    }
    else {
      const parentNode = nodeMap.value.get(Number(node.pid))
      parentName.value = parentNode ? parentNode.menuName : ''
    }
    formData.menuName = node.menuName
    formData.menuCode = node.displayMenuCode
    formData.sort = Number(node.sort || 0)
    formData.weight = Number(node.weight || 0)
    formData.remark = node.remark || ''
    formData.introduce = node.introduce || ''
    formData.ownType = Number(node.ownType || currentPortal.value)
  }
  formMode.value = mode
  drawerVisible.value = true
}

const closeFormModal = () => {
  drawerVisible.value = false
}

const handleSearch = () => {
  searchKeyword.value = keywordInput.value.trim()
  if (!searchKeyword.value)
    expandedRowKeys.value = []
  resetTreeScroll()
}

const handleReset = () => {
  keywordInput.value = ''
  searchKeyword.value = ''
  expandedRowKeys.value = []
  resetTreeScroll()
}

const toggleExpandAll = () => {
  if (expandedRowKeys.value.length) {
    expandedRowKeys.value = []
    return
  }

  expandedRowKeys.value = [...expandableRowKeys.value]
}

const toggleRowExpand = (record: PermissionRecord) => {
  if (!record.children?.length)
    return

  const targetId = Number(record.id)
  const nextKeys = new Set(expandedRowKeys.value.map(key => Number(key)))
  if (nextKeys.has(targetId))
    nextKeys.delete(targetId)
  else
    nextKeys.add(targetId)
  expandedRowKeys.value = Array.from(nextKeys)
}

const getIndentStyle = (depth: number) => {
  return {
    paddingLeft: `${Math.max(0, depth - 1) * 24}px`,
  }
}

const handleSubmit = async () => {
  if (!String(formData.menuName || '').trim() || !String(formData.menuCode || '').trim()) {
    messageService.error('请填写权限名称和编码')
    return
  }

  const payload: PermissionMutationPayload = {
    id: formData.id || undefined,
    pid: Number(formData.pid || 0),
    menuName: String(formData.menuName || '').trim(),
    menuCode: normalizePermissionCode(formData.menuCode),
    sort: Number(formData.sort || 0),
    weight: Number(formData.weight || 0),
    remark: String(formData.remark || '').trim(),
    introduce: String(formData.introduce || '').trim(),
    ownType: currentPortal.value,
  }

  try {
    if (formMode.value === 'edit' && payload.id) {
      const res = await updatePermissionApi(payload as PermissionMutationPayload & { id: number })
      if (res.code !== 200) {
        messageService.error(res.message || '更新失败')
        return
      }
      messageService.success('更新成功')
    }
    else {
      const res = await createPermissionApi(payload)
      if (res.code !== 200) {
        messageService.error(res.message || '新增失败')
        return
      }
      messageService.success('新增成功')
    }
    drawerVisible.value = false
    await loadTree()
  }
  catch (error: any) {
    messageService.error(error?.message || '保存失败')
  }
}

const handleDelete = async () => {
  if (!formData.id)
    return
  try {
    const res = await deletePermissionApi({ id: Number(formData.id) })
    if (res.code !== 200) {
      messageService.error(res.message || '删除失败')
      return
    }
    messageService.success('删除成功')
    formMode.value = null
    drawerVisible.value = false
    await loadTree()
  }
  catch (error: any) {
    messageService.error(error?.message || '删除失败')
  }
}

const handleDeleteByRecord = async (record: PermissionRecord) => {
  try {
    const res = await deletePermissionApi({ id: Number(record.id) })
    if (res.code !== 200) {
      messageService.error(res.message || '删除失败')
      return
    }
    messageService.success('删除成功')
    await loadTree()
  }
  catch (error: any) {
    messageService.error(error?.message || '删除失败')
  }
}

watch(
  [tableData, searchKeyword],
  () => {
    if (searchKeyword.value) {
      expandedRowKeys.value = [...expandableRowKeys.value]
      return
    }

    const availableKeys = new Set(expandableRowKeys.value.map(key => Number(key)))
    expandedRowKeys.value = expandedRowKeys.value.filter(key => availableKeys.has(Number(key)))
  },
  { flush: 'post' },
)

onMounted(() => {
  loadTree()
})
</script>

<style scoped lang="less">
.permission-management {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial,
    'Noto Sans', sans-serif, 'Apple Color Emoji', 'Segoe UI Emoji', 'Segoe UI Symbol',
    'Noto Color Emoji';
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;

  .search-bar {
    margin-bottom: 16px;
  }

  .permission-tree-panel {
    border: 1px solid #f0f0f0;
    border-radius: 8px;
    background: #fff;
    overflow: hidden;
  }

  .permission-tree-header,
  .permission-tree-row {
    display: grid;
    grid-template-columns: minmax(300px, 2.1fr) 200px 108px 88px 88px minmax(320px, 2.3fr) 220px;
    column-gap: 16px;
    align-items: center;
    min-width: 1388px;
  }

  .permission-tree-header {
    min-height: 44px;
    padding: 0 16px;
    border-bottom: 1px solid #f0f0f0;
    background: #fafafa;
    color: rgba(0, 0, 0, 0.85);
    font-size: 13px;
    font-weight: 500;
    line-height: 20px;
  }

  .permission-tree-header > span {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .permission-tree-body {
    position: relative;
    z-index: 0;
    overflow: auto;
    background: #fff;
    isolation: isolate;
    scrollbar-width: thin;
    scrollbar-color: #d9d9d9 transparent;
  }

  .permission-tree-node {
    min-height: 56px;
    min-width: 1388px;
    border-bottom: 1px solid #f0f0f0;
  }

  .permission-tree-row {
    min-height: 56px;
    padding: 10px 16px;
    color: #000;
    font-size: 13px;
    line-height: 20px;
  }

  .permission-tree-name-cell {
    display: flex;
    align-items: flex-start;
    min-width: 0;
  }

  .permission-tree-switcher {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex: 0 0 20px;
    width: 20px;
    height: 20px;
    margin-right: 6px;
    border: none;
    background: transparent;
    color: rgba(0, 0, 0, 0.45);
    cursor: pointer;
    padding: 0;
    transition: transform 0.16s ease, color 0.16s ease;
  }

  .permission-tree-switcher.is-expanded {
    color: rgba(0, 0, 0, 0.72);
    transform: rotate(90deg);
  }

  .permission-tree-switcher--placeholder {
    visibility: hidden;
    cursor: default;
  }

  .permission-tree-row__name,
  .permission-tree-row__code,
  .permission-tree-row__level,
  .permission-tree-row__sort,
  .permission-tree-row__weight,
  .permission-tree-row__intro,
  .permission-tree-row__actions {
    min-width: 0;
  }

  .permission-tree-header__actions,
  .permission-tree-row__level,
  .permission-tree-row__sort,
  .permission-tree-row__weight {
    display: flex;
    align-items: center;
  }

  .permission-tree-header__actions,
  .permission-tree-row__actions {
    position: sticky;
    right: 0;
    z-index: 1;
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 4px;
    white-space: nowrap;
    box-shadow: -10px 0 12px -12px rgba(15, 23, 42, 0.18);
  }

  .permission-tree-header__actions {
    background: #fafafa;
    z-index: 2;
  }

  .permission-tree-row__actions {
    background: #fff;
  }

  .ellipsis-text {
    display: inline-block;
    max-width: 100%;
    overflow: hidden;
    color: #000;
    text-overflow: ellipsis;
    white-space: nowrap;
    vertical-align: bottom;
  }

  :deep(.ant-card),
  :deep(.ant-card-head),
  :deep(.ant-card-body),
  :deep(.ant-space),
  :deep(.ant-segmented),
  :deep(.ant-input),
  :deep(.ant-btn),
  :deep(.ant-tag),
  :deep(.ant-drawer),
  :deep(.ant-form),
  :deep(.ant-form-item),
  :deep(.ant-form-item-label > label) {
    font-family: inherit;
  }

  .permission-tree-body::-webkit-scrollbar {
    width: 8px;
    height: 8px;
  }

  .permission-tree-body::-webkit-scrollbar-thumb {
    border-radius: 999px;
    background: rgba(0, 0, 0, 0.18);
  }

  .permission-tree-body::-webkit-scrollbar-track {
    background: transparent;
  }
}
</style>

<style scoped lang="less">
.menu-name-text {
  display: block;
  color: #000;
  line-height: 22px;
  vertical-align: middle;
  word-break: break-word;
  white-space: normal;
}

.permission-form-card {
  margin-top: 8px;
  border: 1px solid #e8edf5;
  border-radius: 18px;
  background:
    linear-gradient(180deg, rgba(22, 119, 255, 0.06) 0%, rgba(22, 119, 255, 0) 132px),
    #fff;
  box-shadow: 0 14px 32px rgba(15, 23, 42, 0.06);
  padding: 20px 22px 6px;
}

.permission-form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 20px;
}

.permission-form-grid__item {
  min-width: 0;
}

.permission-form-grid__item--full {
  grid-column: 1 / -1;
}

.permission-form :deep(.ant-form-item) {
  margin-bottom: 16px;
}

.permission-form-modal__footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 12px;
}

@media (max-width: 768px) {
  .permission-form-grid {
    grid-template-columns: 1fr;
    gap: 0;
  }

  .permission-form-grid__item--full {
    grid-column: auto;
  }

  .permission-form-modal__footer {
    flex-wrap: wrap;
  }
}
</style>
