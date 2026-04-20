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
        </a-space>
      </div>

      <a-table
        v-model:expandedRowKeys="expandedRowKeys"
        class="permission-table"
        :columns="columns"
        :data-source="tableData"
        :loading="loading"
        :pagination="false"
        :scroll="{ x: 1280 }"
        table-layout="fixed"
        size="middle"
        row-key="id"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'menuName'">
            <span class="menu-name-text" :title="toPermissionRecord(record).menuName || '--'">
              {{ toPermissionRecord(record).menuName || '--' }}
            </span>
          </template>

          <template v-else-if="column.key === 'menuCode'">
            <span class="ellipsis-text" :title="toPermissionRecord(record).displayMenuCode">
              {{ toPermissionRecord(record).displayMenuCode }}
            </span>
          </template>

          <template v-else-if="column.key === 'levelLabel'">
            <a-tag :color="toPermissionRecord(record).levelTagColor">
              {{ toPermissionRecord(record).levelLabel }}
            </a-tag>
          </template>

          <template v-else-if="column.key === 'introduce'">
            <span class="ellipsis-text" :title="toPermissionRecord(record).introduceText">
              {{ toPermissionRecord(record).introduceText }}
            </span>
          </template>

          <template v-else-if="column.key === 'actions'">
            <div class="action-cell">
              <a-button
                v-if="hasPermission(AccessEnum.menuPermissions_update)"
                type="link"
                size="small"
                @click="openDrawer('edit', toPermissionRecord(record))"
              >
                修改
              </a-button>
              <a-button
                v-if="hasPermission(AccessEnum.menuPermissions_add) && Number(toPermissionRecord(record).depth || 1) < 3"
                type="link"
                size="small"
                @click="handleAddChild(toPermissionRecord(record))"
              >
                新增
              </a-button>
              <a-popconfirm
                v-if="hasPermission(AccessEnum.menuPermissions_delete)"
                title="确定要删除该权限吗？"
                @confirm="handleDeleteByRecord(toPermissionRecord(record))"
              >
                <a-button type="link" size="small" danger>
                  删除
                </a-button>
              </a-popconfirm>
            </div>
          </template>

          <template v-else>
            {{ (toPermissionRecord(record) as any)[column.dataIndex as string] ?? '--' }}
          </template>
        </template>
      </a-table>
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

            <div v-if="showAccessDeniedImageField" class="permission-form-grid__item permission-form-grid__item--full">
              <a-form-item>
                <div class="access-denied-image-field">
                  <div class="access-denied-image-field__header">
                    <div class="access-denied-image-field__title">
                      无权限展示图
                    </div>

                    <div class="access-denied-image-field__actions">
                      <a-upload
                        accept=".jpg,.jpeg,.png,.webp"
                        :show-upload-list="false"
                        :custom-request="handleAccessDeniedImageUpload"
                        :before-upload="beforeAccessDeniedImageUpload"
                      >
                        <a-button type="primary" :loading="accessDeniedImageUploading">
                          上传图片
                        </a-button>
                      </a-upload>
                      <a-button v-if="formData.accessDeniedImage" @click="clearAccessDeniedImage">
                        清空图片
                      </a-button>
                    </div>
                  </div>

                  <div class="access-denied-image-field__preview" :class="{ 'is-empty': !formData.accessDeniedImage }">
                    <img
                      v-if="formData.accessDeniedImage"
                      :src="formData.accessDeniedImage"
                      alt="无权限展示图"
                    >
                    <div v-else class="access-denied-image-field__placeholder">
                      机构端未开通该页面功能时，将优先展示这里上传的图片。
                    </div>
                  </div>
                </div>
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
import type { UploadRequestOption } from 'ant-design-vue/es/vc-upload/interface'
import { computed, onMounted, reactive, ref, shallowRef, watch } from 'vue'
import type { TableColumnsType } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import * as qiniu from 'qiniu-js'
import {
  createPermissionApi,
  deletePermissionApi,
  getPermissionTreeApi,
  updatePermissionApi,
  type PermissionMenuItem,
  type PermissionMutationPayload,
} from '@/api/platform/permissions'
import { getQiniuToken } from '@/api/qiniu'
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

const currentPortal = ref<PortalEnum>(PortalEnum.INSTITUTION)
const portalOptions = [
  { label: '平台端', value: PortalEnum.PLATFORM, disabled: true },
  { label: '机构端', value: PortalEnum.INSTITUTION },
]

const formMode = ref<FormMode>(null)
const drawerVisible = ref(false)
const parentName = ref<string>('根菜单')
const accessDeniedImageUploading = ref(false)

const formData = reactive<PermissionMutationPayload>({
  id: undefined,
  pid: 0,
  menuName: '',
  menuCode: '',
  sort: 0,
  weight: 0,
  remark: '',
  introduce: '',
  accessDeniedImage: '',
  ownType: PortalEnum.INSTITUTION,
})

const columns: TableColumnsType<PermissionRecord> = [
  {
    title: '菜单名称',
    dataIndex: 'menuName',
    key: 'menuName',
    width: 260,
  },
  {
    title: '权限标识',
    dataIndex: 'displayMenuCode',
    key: 'menuCode',
    width: 220,
  },
  {
    title: '层级',
    key: 'levelLabel',
    width: 110,
  },
  {
    title: '排序',
    dataIndex: 'sort',
    key: 'sort',
    width: 90,
  },
  {
    title: '权重',
    dataIndex: 'weight',
    key: 'weight',
    width: 90,
  },
  {
    title: '描述',
    dataIndex: 'introduce',
    key: 'introduce',
    width: 320,
  },
  {
    title: '操作',
    key: 'actions',
    width: 168,
    fixed: 'right',
  },
]

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
  if (raw.startsWith('grp:'))
    return `grp:${normalizeCamelSuffix(raw.slice('grp:'.length))}`
  if (raw.startsWith('page:'))
    return `page:${normalizeCamelSuffix(raw.slice('page:'.length))}`
  if (raw.startsWith('perm:'))
    return `perm:${normalizeCamelSuffix(raw.slice('perm:'.length))}`
  return raw
}

const isLegacyInstitutionPermissionCode = (code: string | number | null | undefined) => {
  const raw = String(code ?? '').trim()
  return /^INST_(GROUP|ROUTE|AUTH)_/.test(raw)
}

const isCurrentInstitutionPermissionCode = (code: string | number | null | undefined) => {
  const raw = normalizePermissionCode(code)
  return /^(grp|page|perm):/.test(raw)
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

const toPermissionRecord = (record: Record<string, any>) => record as PermissionRecord

const currentParentNode = computed(() => {
  const pid = Number(formData.pid || 0)
  if (pid <= 0)
    return undefined
  return nodeMap.value.get(pid)
})

const showAccessDeniedImageField = computed(() => {
  const normalizedCode = normalizePermissionCode(formData.menuCode)
  const menuName = String(formData.menuName || '').trim()
  const parentCode = normalizePermissionCode(currentParentNode.value?.menuCode)

  if (menuName === '页面功能访问')
    return true
  if (/^perm:[a-z][a-zA-Z0-9]*Use$/.test(normalizedCode))
    return true
  return parentCode.startsWith('page:') && menuName.includes('页面功能')
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
  formData.accessDeniedImage = ''
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
    formData.accessDeniedImage = node.accessDeniedImage || ''
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
}

const handleReset = () => {
  keywordInput.value = ''
  searchKeyword.value = ''
  expandedRowKeys.value = []
}

function beforeAccessDeniedImageUpload(file: File) {
  if (!file.type.startsWith('image/')) {
    messageService.error('只能上传图片文件')
    return false
  }
  if (file.size / 1024 / 1024 >= 5) {
    messageService.error('图片大小不能超过 5MB')
    return false
  }
  return true
}

async function handleAccessDeniedImageUpload(options: UploadRequestOption) {
  const rawFile = options.file as File
  if (!rawFile || !beforeAccessDeniedImageUpload(rawFile)) {
    options.onError?.(new Error('invalid file'))
    return
  }

  accessDeniedImageUploading.value = true
  try {
    const tokenRes: any = await getQiniuToken()
    const { token, uuid, buckethostname } = tokenRes.result || {}
    if (!token || !uuid || !buckethostname)
      throw new Error('获取上传凭证失败')

    const ext = rawFile.name.includes('.') ? rawFile.name.slice(rawFile.name.lastIndexOf('.')) : '.png'
    const key = `permission/access-denied/${uuid}${ext}`
    const config = {
      useCdnDomain: true,
      region: qiniu.region.z0,
    }
    const putExtra = {
      fname: rawFile.name,
      mimeType: rawFile.type,
    }

    const observable = qiniu.upload(rawFile, key, token, putExtra, config)
    observable.subscribe({
      error(error) {
        messageService.error(error?.message || '上传图片失败')
        accessDeniedImageUploading.value = false
        options.onError?.(error)
      },
      complete(result) {
        formData.accessDeniedImage = `${buckethostname}${result.key}`
        accessDeniedImageUploading.value = false
        messageService.success('图片上传成功')
        options.onSuccess?.(result as any)
      },
    })
  }
  catch (error: any) {
    accessDeniedImageUploading.value = false
    messageService.error(error?.message || '上传图片失败')
    options.onError?.(error)
  }
}

const clearAccessDeniedImage = () => {
  formData.accessDeniedImage = ''
}

const handleSubmit = async () => {
  if (!String(formData.menuName || '').trim() || !String(formData.menuCode || '').trim()) {
    messageService.error('请填写权限名称和编码')
    return
  }

  if (currentPortal.value === PortalEnum.INSTITUTION && isLegacyInstitutionPermissionCode(formData.menuCode)) {
    messageService.error('机构端权限标识请使用 grp:/page:/perm: 新 code')
    return
  }

  if (currentPortal.value === PortalEnum.INSTITUTION && !isCurrentInstitutionPermissionCode(formData.menuCode)) {
    messageService.error('机构端权限标识必须使用 grp:/page:/perm: 前缀')
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
    accessDeniedImage: String(formData.accessDeniedImage || '').trim(),
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

  .ellipsis-text {
    display: inline-block;
    max-width: 100%;
    overflow: hidden;
    color: #000;
    text-overflow: ellipsis;
    white-space: nowrap;
    vertical-align: bottom;
  }

  .action-cell {
    display: flex;
    align-items: center;
    gap: 4px;
    white-space: nowrap;
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

  :deep(.permission-table .ant-table-container) {
    border: 1px solid #f0f0f0;
    border-radius: 8px;
    overflow: hidden;
  }

  :deep(.permission-table .ant-table-thead > tr > th) {
    padding: 12px 16px;
    background: #fafafa !important;
    color: #262626;
    font-size: 14px;
    font-weight: 500;
    border-bottom: 1px solid #f0f0f0;
  }

  :deep(.permission-table .ant-table-tbody > tr > td) {
    padding: 14px 16px;
    border-bottom: 1px solid #f5f5f5;
    vertical-align: middle;
    color: #000;
  }

  :deep(.permission-table .ant-table-tbody > tr:hover > td) {
    background: #fcfcfc;
  }

  :deep(.permission-table .ant-table-cell-fix-right) {
    background: #fff;
  }

  :deep(.permission-table .ant-table-thead > tr > th.ant-table-cell-fix-right) {
    background: #fafafa !important;
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

.access-denied-image-field {
  width: 100%;
}

.access-denied-image-field__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 12px;
}

.access-denied-image-field__title {
  color: #1f2329;
  font-size: 14px;
  font-weight: 500;
  line-height: 22px;
}

.access-denied-image-field__preview {
  width: 100%;
  min-height: 220px;
  overflow: hidden;
  border: 1px solid #dfe5f0;
  border-radius: 14px;
  background: #f7f9fc;
}

.access-denied-image-field__preview.is-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.access-denied-image-field__preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.access-denied-image-field__placeholder {
  color: #6b7280;
  font-size: 13px;
  line-height: 22px;
}

.access-denied-image-field__actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  flex-shrink: 0;
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

  .access-denied-image-field__header {
    align-items: flex-start;
    flex-direction: column;
  }

  .access-denied-image-field__preview {
    min-height: 180px;
  }

  .access-denied-image-field__actions {
    width: 100%;
    justify-content: flex-start;
  }
}
</style>
