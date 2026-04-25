<script setup lang="ts">
import type { TableColumnsType } from 'ant-design-vue'
import type { UploadRequestOption } from 'ant-design-vue/es/vc-upload/interface'
import { computed, onMounted, reactive, ref } from 'vue'
import QRCode from 'qrcode'
import * as qiniu from 'qiniu-js'
import {
  ApartmentOutlined,
  CheckCircleOutlined,
  DeleteOutlined,
  DownloadOutlined,
  GlobalOutlined,
  KeyOutlined,
  LinkOutlined,
  PlusOutlined,
  SafetyCertificateOutlined,
  SearchOutlined,
  ShopOutlined,
  TeamOutlined,
  UploadOutlined,
} from '@ant-design/icons-vue'
import messageService from '@/utils/messageService'
import { listTenantsApi, saveTenantApi } from '@/api/platform/tenants'
import { pageInstitutionsApi } from '@/api/platform/institutions'
import { pageVersionsApi } from '@/api/platform/versions'
import { getQiniuToken } from '@/api/qiniu'
import { resolveUploadErrorMessage, validateUploadFileByToken } from '@/utils/upload-limit'

interface TenantRecord {
  tenantId: string
  tenantName: string
  tenantType: string
  edition?: string
  status?: string
  isolationMode?: string
  institutionCount: number
  institutionIds?: number[]
  menuCount: number
  moduleCount?: number
  moduleIds?: number[]
  moduleNames?: string[]
  adminUsernames: string[]
  domains: string[]
  adminDomains?: string[]
  institutionDomains?: string[]
  loginBrand?: TenantLoginBrandConfig
  platformLoginBrand?: TenantLoginBrandConfig
  institutionLoginBrand?: TenantLoginBrandConfig
}

interface TenantLoginBrandConfig {
  brandName?: string
  template?: string
  logoUrl?: string
  loginTitle?: string
  loginSubtitle?: string
  backgroundUrl?: string
  primaryColor?: string
  copyright?: string
  heroBadge?: string
  heroTitle?: string
  heroDescription?: string
}

interface InstitutionOption {
  id: number
  organName: string
  tenantId?: string
  tenantName?: string
}

interface VersionOption {
  id: number
  name: string
  menuCount?: number
  price?: number
}

interface TenantFormState {
  tenantId: string
  tenantName: string
  edition: string
  status: string
  isolationMode: string
  adminDomain: string
  institutionDomain: string
  institutionIds: number[]
  moduleIds: number[]
  adminUsername: string
  adminPassword: string
  adminNickName: string
  adminMobile: string
  platformLoginBrand: Required<TenantLoginBrandConfig>
  institutionLoginBrand: Required<TenantLoginBrandConfig>
  remark: string
}

const loading = ref(false)
const saving = ref(false)
const modalOpen = ref(false)
const authorizationModalOpen = ref(false)
const authorizationSaving = ref(false)
const authorizationTenant = ref<TenantRecord | null>(null)
const authorizationModuleId = ref<number | undefined>(undefined)
const loginAddressModalOpen = ref(false)
const loginAddressTenant = ref<TenantRecord | null>(null)
const loginAddressQr = reactive<Record<'platform' | 'institution', string>>({
  platform: '',
  institution: '',
})
const editingTenantId = ref('')
const keyword = ref('')
const statusFilter = ref<'all' | 'active' | 'disabled' | 'incomplete'>('all')
const tenantRows = ref<TenantRecord[]>([])
const institutionOptions = ref<InstitutionOption[]>([])
const institutionLoading = ref(false)
const versionOptions = ref<VersionOption[]>([])
const versionLoading = ref(false)

const defaultLoginBrand: Required<TenantLoginBrandConfig> = {
  template: 'business-split',
  brandName: '',
  logoUrl: '',
  loginTitle: '',
  loginSubtitle: '',
  backgroundUrl: '',
  primaryColor: '#1677ff',
  copyright: '',
  heroBadge: '',
  heroTitle: '',
  heroDescription: '',
}

const platformTemplateOptions = [
  { label: '商务分屏登录', value: 'business-split' },
  { label: '居中品牌卡片', value: 'center-card' },
  { label: '极简企业门户', value: 'minimal-portal' },
]
const institutionTemplateOptions = [
  { label: '教务分屏登录', value: 'education-split' },
  { label: '校区品牌卡片', value: 'campus-card' },
  { label: '轻量门户登录', value: 'clean-portal' },
]
const brandColorOptions = [
  { label: '科技蓝', value: '#1677ff' },
  { label: '活力橙', value: '#fe8130' },
  { label: '教育绿', value: '#13ad74' },
  { label: '品牌紫', value: '#7c3aed' },
  { label: '商务青', value: '#08979c' },
  { label: '高级黑', value: '#1f2937' },
]
const brandUploading = reactive<Record<string, boolean>>({})
const brandUploadProgress = reactive<Record<string, number>>({})

type LoginBrandScope = 'platform' | 'institution'
type LoginBrandAssetField = 'logoUrl' | 'backgroundUrl'

function createDefaultLoginBrand(tenantName = '', template = 'business-split'): Required<TenantLoginBrandConfig> {
  return {
    ...defaultLoginBrand,
    template,
    brandName: tenantName,
    loginTitle: tenantName ? `${tenantName}管理后台` : '',
    loginSubtitle: '请输入账号密码登录',
    heroBadge: tenantName,
    heroTitle: tenantName ? `欢迎进入${tenantName}` : '',
    heroDescription: '独立租户后台，按客户域名、菜单权限和业务配置隔离运行。',
  }
}

function assignLoginBrand(target: Required<TenantLoginBrandConfig>, value?: TenantLoginBrandConfig, tenantName = '', template = 'business-split') {
  Object.assign(target, createDefaultLoginBrand(tenantName, template), value || {})
}

function getLoginBrand(scope: LoginBrandScope) {
  return scope === 'platform' ? formState.platformLoginBrand : formState.institutionLoginBrand
}

function getBrandUploadKey(scope: LoginBrandScope, field: LoginBrandAssetField) {
  return `${scope}-${field}`
}

function getBrandUploadLabel(field: LoginBrandAssetField) {
  return field === 'logoUrl' ? 'Logo' : '登录背景图'
}

function getBrandAssetFolder(scope: LoginBrandScope, field: LoginBrandAssetField) {
  const entryFolder = scope === 'platform' ? 'platform-admin' : 'institution-admin'
  const assetFolder = field === 'logoUrl' ? 'logo' : 'background'
  return `tenant-login/${entryFolder}/${assetFolder}`
}

function isBrandUploading(scope: LoginBrandScope, field: LoginBrandAssetField) {
  return !!brandUploading[getBrandUploadKey(scope, field)]
}

function getBrandUploadProgress(scope: LoginBrandScope, field: LoginBrandAssetField) {
  return brandUploadProgress[getBrandUploadKey(scope, field)] || 0
}

function selectBrandColor(scope: LoginBrandScope, color: string) {
  getLoginBrand(scope).primaryColor = color
}

function clearBrandAsset(scope: LoginBrandScope, field: LoginBrandAssetField) {
  getLoginBrand(scope)[field] = ''
}

function beforeBrandImageUpload(file: File, label: string) {
  const isImage = file.type.startsWith('image/')
  if (!isImage) {
    messageService.warning(`${label}只能上传图片文件`)
    return false
  }

  const isLt8M = file.size / 1024 / 1024 <= 8
  if (!isLt8M) {
    messageService.warning(`${label}大小不能超过 8MB`)
    return false
  }

  return true
}

async function handleBrandAssetUpload(options: UploadRequestOption, scope: LoginBrandScope, field: LoginBrandAssetField) {
  const rawFile = options.file as File
  const label = getBrandUploadLabel(field)
  if (!rawFile || !beforeBrandImageUpload(rawFile, label)) {
    options.onError?.(new Error('invalid file'))
    return
  }

  const uploadKey = getBrandUploadKey(scope, field)
  brandUploading[uploadKey] = true
  brandUploadProgress[uploadKey] = 0

  try {
    const tokenRes: any = await getQiniuToken()
    const { token, uuid, buckethostname } = tokenRes.result || {}
    if (!token || !uuid || !buckethostname)
      throw new Error(tokenRes?.message || '获取上传凭证失败')
    validateUploadFileByToken(rawFile, tokenRes.result, label)

    const ext = rawFile.name.includes('.') ? rawFile.name.slice(rawFile.name.lastIndexOf('.')) : '.png'
    const key = `${getBrandAssetFolder(scope, field)}/${uuid}${ext}`
    const observable = qiniu.upload(rawFile, key, token, {
      fname: rawFile.name,
      mimeType: rawFile.type,
    }, {
      useCdnDomain: true,
      region: qiniu.region.z0,
    })

    observable.subscribe({
      next(result) {
        brandUploadProgress[uploadKey] = Math.floor(result.total.percent)
      },
      error(error) {
        console.error('upload tenant login asset failed', error)
        messageService.error(resolveUploadErrorMessage(error, `${label}上传失败`))
        brandUploading[uploadKey] = false
        brandUploadProgress[uploadKey] = 0
        options.onError?.(error)
      },
      complete(result) {
        getLoginBrand(scope)[field] = `${buckethostname}${result.key}`
        brandUploading[uploadKey] = false
        brandUploadProgress[uploadKey] = 100
        messageService.success(`${label}上传成功`)
        options.onSuccess?.(result as any)
      },
    })
  }
  catch (error: any) {
    console.error('prepare tenant login asset upload failed', error)
    messageService.error(resolveUploadErrorMessage(error, `${label}上传失败`))
    brandUploading[uploadKey] = false
    brandUploadProgress[uploadKey] = 0
    options.onError?.(error)
  }
}

const formState = reactive<TenantFormState>({
  tenantId: '',
  tenantName: '',
  edition: 'enterprise',
  status: 'active',
  isolationMode: 'shared_db',
  adminDomain: '',
  institutionDomain: '',
  institutionIds: [],
  moduleIds: [],
  adminUsername: '',
  adminPassword: '',
  adminNickName: '',
  adminMobile: '',
  platformLoginBrand: createDefaultLoginBrand('', 'business-split'),
  institutionLoginBrand: createDefaultLoginBrand('', 'education-split'),
  remark: '',
})

const partnerRows = computed(() => tenantRows.value.filter(item => item.tenantType !== 'platform'))
const filteredRows = computed(() => {
  const searchValue = keyword.value.trim().toLowerCase()
  return partnerRows.value.filter((item) => {
    const matchesKeyword = !searchValue
      || item.tenantName.toLowerCase().includes(searchValue)
      || item.tenantId.toLowerCase().includes(searchValue)
      || item.domains?.some(domain => domain.toLowerCase().includes(searchValue))
      || item.adminUsernames?.some(username => username.toLowerCase().includes(searchValue))

    const incomplete = !item.domains?.length || !item.adminUsernames?.length || !item.institutionCount
    const matchesStatus = statusFilter.value === 'all'
      || item.status === statusFilter.value
      || (statusFilter.value === 'incomplete' && incomplete)

    return matchesKeyword && matchesStatus
  })
})

const activeTenantCount = computed(() => partnerRows.value.filter(item => item.status !== 'disabled').length)
const institutionTotal = computed(() => partnerRows.value.reduce((total, item) => total + Number(item.institutionCount || 0), 0))
const domainTotal = computed(() => partnerRows.value.reduce((total, item) => total + (item.domains?.length || 0), 0))
const incompleteCount = computed(() => partnerRows.value.filter(item => !item.domains?.length || !item.adminUsernames?.length || !item.institutionCount).length)
const institutionSelectOptions = computed(() => {
  const currentTenantId = formState.tenantId.trim()
  return institutionOptions.value.map((item) => {
    const boundTenantId = String(item.tenantId || '').trim()
    const isBoundToOtherTenant = !!boundTenantId && boundTenantId !== currentTenantId
    return {
      value: item.id,
      label: `${item.organName}（ID:${item.id}）`,
      organName: item.organName,
      institutionId: item.id,
      tenantId: boundTenantId,
      tenantName: item.tenantName || '',
      disabled: isBoundToOtherTenant,
    }
  })
})


const loginAddressItems = computed(() => {
  const tenant = loginAddressTenant.value
  if (!tenant)
    return []

  return [
    {
      key: 'platform' as const,
      title: '子总控后台',
      description: '客户管理员登录，管理租户内版本、机构和授权配置。',
      domain: getPrimaryAdminDomain(tenant),
      url: buildLoginUrl(getPrimaryAdminDomain(tenant), 'platform'),
      qr: loginAddressQr.platform,
    },
    {
      key: 'institution' as const,
      title: '机构端后台',
      description: '租户下属机构登录，按机构归属和权限范围进入业务后台。',
      domain: getPrimaryInstitutionDomain(tenant),
      url: buildLoginUrl(getPrimaryInstitutionDomain(tenant), 'institution'),
      qr: loginAddressQr.institution,
    },
  ]
})

const pagination = reactive({
  current: 1,
  pageSize: 20,
  showSizeChanger: true,
  showTotal: (total: number) => `共 ${total} 个合作客户`,
})

const columns: TableColumnsType<TenantRecord> = [
  { title: '合作客户', key: 'tenant', width: 270, fixed: 'left' as const },
  { title: '开通状态', key: 'status', width: 130 },
  { title: '机构规模', key: 'institutions', width: 150 },
  { title: '子总控账号', key: 'admins', width: 220 },
  { title: '独立域名', key: 'domains', width: 260 },
  { title: '授权版本', key: 'modules', width: 220 },
  { title: '授权资源', key: 'authorization', width: 150 },
  { title: '操作', key: 'action', width: 240, fixed: 'right' as const },
]

function statusClass(status?: string) {
  return status === 'disabled' ? 'status-pill--disabled' : 'status-pill--active'
}

function statusText(status?: string) {
  return status === 'disabled' ? '停用' : '正常运营'
}

function editionText(edition?: string) {
  const map: Record<string, string> = {
    enterprise: '企业版',
    professional: '专业版',
    platform: '平台版',
  }
  return map[edition || ''] || edition || '企业版'
}

function isolationText() {
  return '共享库'
}

function resetForm() {
  editingTenantId.value = ''
  formState.tenantId = ''
  formState.tenantName = ''
  formState.edition = 'enterprise'
  formState.status = 'active'
  formState.isolationMode = 'shared_db'
  formState.adminDomain = ''
  formState.institutionDomain = ''
  formState.institutionIds = []
  formState.moduleIds = []
  formState.adminUsername = ''
  formState.adminPassword = ''
  formState.adminNickName = ''
  formState.adminMobile = ''
  assignLoginBrand(formState.platformLoginBrand, undefined, '', 'business-split')
  assignLoginBrand(formState.institutionLoginBrand, undefined, '', 'education-split')
  formState.remark = ''
}

function normalizeTenantId() {
  formState.tenantId = formState.tenantId.trim().toLowerCase().replace(/\s+/g, '-').replace(/[^a-z0-9-_]/g, '')
}

function openCreateModal() {
  resetForm()
  modalOpen.value = true
}

function openEditModal(record: TenantRecord) {
  editingTenantId.value = record.tenantId
  formState.tenantId = record.tenantId
  formState.tenantName = record.tenantName
  formState.edition = record.edition || 'enterprise'
  formState.status = record.status || 'active'
  formState.isolationMode = record.isolationMode || 'shared_db'
  formState.adminDomain = (record.adminDomains?.[0] || record.domains?.[0] || '').trim()
  formState.institutionDomain = (record.institutionDomains?.[0] || '').trim()
  formState.institutionIds = [...(record.institutionIds || [])]
  formState.moduleIds = [...(record.moduleIds || [])]
  formState.adminUsername = record.adminUsernames?.[0] || ''
  formState.adminPassword = ''
  formState.adminNickName = ''
  formState.adminMobile = ''
  assignLoginBrand(formState.platformLoginBrand, record.platformLoginBrand || record.loginBrand, record.tenantName, 'business-split')
  assignLoginBrand(formState.institutionLoginBrand, record.institutionLoginBrand || record.loginBrand, record.tenantName, 'education-split')
  formState.remark = ''
  modalOpen.value = true
}

function openAuthorizationModal(record: TenantRecord) {
  authorizationTenant.value = record
  authorizationModuleId.value = record.moduleIds?.[0]
  authorizationModalOpen.value = true
}

async function handleSaveAuthorization() {
  const tenant = authorizationTenant.value
  if (!tenant)
    return

  authorizationSaving.value = true
  try {
    const adminDomains = [...(tenant.adminDomains || tenant.domains || [])]
    const institutionDomains = [...(tenant.institutionDomains || [])]
    await saveTenantApi({
      tenantId: tenant.tenantId,
      tenantName: tenant.tenantName,
      tenantType: 'partner',
      edition: tenant.edition || 'enterprise',
      status: tenant.status || 'active',
      isolationMode: tenant.isolationMode || 'shared_db',
      domains: adminDomains,
      adminDomains,
      institutionDomains,
      institutionIds: tenant.institutionIds || [],
      moduleIds: authorizationModuleId.value ? [authorizationModuleId.value] : [],
      adminUsername: tenant.adminUsernames?.[0] || '',
      adminPassword: '',
      adminNickName: '',
      adminMobile: '',
      loginBrand: tenant.platformLoginBrand || tenant.loginBrand,
      platformLoginBrand: tenant.platformLoginBrand || tenant.loginBrand,
      institutionLoginBrand: tenant.institutionLoginBrand || tenant.loginBrand,
      remark: '',
    })
    messageService.success('授权配置已保存')
    authorizationModalOpen.value = false
    await loadTenants()
  }
  catch (error) {
    console.error(error)
    messageService.error('授权保存失败，请检查版本授权范围')
  }
  finally {
    authorizationSaving.value = false
  }
}

async function loadTenants() {
  loading.value = true
  try {
    const res = await listTenantsApi({ keyword: keyword.value.trim() || undefined })
    tenantRows.value = (res.data || res.result || []) as TenantRecord[]
  }
  catch (error) {
    console.error(error)
    messageService.error('租户列表加载失败')
  }
  finally {
    loading.value = false
  }
}

async function loadInstitutionOptions() {
  institutionLoading.value = true
  try {
    const res = await pageInstitutionsApi({ current: 1, size: 300 })
    const payload = res.data as any
    institutionOptions.value = payload?.items || res.result || []
  }
  catch (error) {
    console.warn('load institution options failed', error)
    institutionOptions.value = []
  }
  finally {
    institutionLoading.value = false
  }
}

async function loadVersionOptions() {
  versionLoading.value = true
  try {
    const res = await pageVersionsApi({ current: 1, size: 300, type: 1 })
    const payload = res.data as any
    versionOptions.value = payload?.items || res.result || []
  }
  catch (error) {
    console.warn('load version options failed', error)
    versionOptions.value = []
  }
  finally {
    versionLoading.value = false
  }
}

async function handleSave() {
  normalizeTenantId()
  if (!formState.tenantId) {
    messageService.warning('请填写租户标识')
    return
  }
  if (!formState.tenantName.trim()) {
    messageService.warning('请填写客户名称')
    return
  }

  const adminDomains = formState.adminDomain.trim() ? [formState.adminDomain.trim().toLowerCase()] : []
  const institutionDomains = formState.institutionDomain.trim() ? [formState.institutionDomain.trim().toLowerCase()] : []

  saving.value = true
  try {
    await saveTenantApi({
      tenantId: formState.tenantId,
      tenantName: formState.tenantName.trim(),
      tenantType: 'partner',
      edition: formState.edition,
      status: formState.status,
      isolationMode: formState.isolationMode,
      domains: adminDomains,
      adminDomains,
      institutionDomains,
      institutionIds: formState.institutionIds,
      moduleIds: formState.moduleIds,
      adminUsername: formState.adminUsername.trim(),
      adminPassword: formState.adminPassword.trim(),
      adminNickName: formState.adminNickName.trim(),
      adminMobile: formState.adminMobile.trim(),
      loginBrand: {
        ...formState.platformLoginBrand,
        brandName: formState.platformLoginBrand.brandName.trim() || formState.tenantName.trim(),
        primaryColor: formState.platformLoginBrand.primaryColor.trim() || '#1677ff',
      },
      platformLoginBrand: {
        ...formState.platformLoginBrand,
        brandName: formState.platformLoginBrand.brandName.trim() || formState.tenantName.trim(),
        primaryColor: formState.platformLoginBrand.primaryColor.trim() || '#1677ff',
      },
      institutionLoginBrand: {
        ...formState.institutionLoginBrand,
        brandName: formState.institutionLoginBrand.brandName.trim() || formState.tenantName.trim(),
        primaryColor: formState.institutionLoginBrand.primaryColor.trim() || formState.platformLoginBrand.primaryColor.trim() || '#1677ff',
      },
      remark: formState.remark.trim(),
    })
    messageService.success('合作客户租户已保存')
    modalOpen.value = false
    await loadTenants()
  }
  catch (error) {
    console.error(error)
    messageService.error('保存失败，请检查客户标识、域名或账号是否重复')
  }
  finally {
    saving.value = false
  }
}


function getPrimaryAdminDomain(record: TenantRecord) {
  return String(record.adminDomains?.[0] || record.domains?.[0] || '').trim()
}

function getPrimaryInstitutionDomain(record: TenantRecord) {
  return String(record.institutionDomains?.[0] || '').trim()
}

function buildLoginUrl(domain: string, entryType: 'platform' | 'institution') {
  const normalized = String(domain || '').trim()
  if (!normalized)
    return ''
  if (/^https?:\/\//i.test(normalized))
    return normalized

  const protocol = typeof window !== 'undefined' ? window.location.protocol : 'https:'
  const hasPort = /:\d+$/.test(normalized)
  const devPort = import.meta.env.DEV && !hasPort
    ? (entryType === 'institution' ? ':6678' : (window.location.port ? `:${window.location.port}` : ''))
    : ''
  return `${protocol}//${normalized}${devPort}/`
}

async function generateLoginAddressQrs() {
  loginAddressQr.platform = ''
  loginAddressQr.institution = ''
  const items = loginAddressItems.value
  await Promise.all(items.map(async (item) => {
    if (!item.url)
      return
    loginAddressQr[item.key] = await QRCode.toDataURL(item.url, {
      width: 180,
      margin: 1,
      color: {
        dark: '#111827',
        light: '#ffffff',
      },
    })
  }))
}

async function openLoginAddressModal(record: TenantRecord) {
  loginAddressTenant.value = record
  loginAddressModalOpen.value = true
  await generateLoginAddressQrs()
}

async function copyLoginAddress(url: string) {
  if (!url)
    return
  try {
    await navigator.clipboard.writeText(url)
    messageService.success('登录地址已复制')
  }
  catch (error) {
    console.warn('copy login address failed', error)
    messageService.warning('复制失败，请手动复制')
  }
}

function drawRoundRect(ctx: CanvasRenderingContext2D, x: number, y: number, width: number, height: number, radius: number) {
  ctx.beginPath()
  ctx.moveTo(x + radius, y)
  ctx.arcTo(x + width, y, x + width, y + height, radius)
  ctx.arcTo(x + width, y + height, x, y + height, radius)
  ctx.arcTo(x, y + height, x, y, radius)
  ctx.arcTo(x, y, x + width, y, radius)
  ctx.closePath()
}

function loadCanvasImage(src: string) {
  return new Promise<HTMLImageElement>((resolve, reject) => {
    const image = new Image()
    image.onload = () => resolve(image)
    image.onerror = reject
    image.src = src
  })
}

function drawCanvasText(ctx: CanvasRenderingContext2D, text: string, x: number, y: number, maxWidth: number) {
  const normalized = String(text || '')
  if (ctx.measureText(normalized).width <= maxWidth) {
    ctx.fillText(normalized, x, y)
    return
  }
  let value = normalized
  while (value.length > 0 && ctx.measureText(`${value}...`).width > maxWidth)
    value = value.slice(0, -1)
  ctx.fillText(`${value}...`, x, y)
}

async function downloadLoginAddressImage() {
  const tenant = loginAddressTenant.value
  if (!tenant)
    return

  const canvas = document.createElement('canvas')
  canvas.width = 1120
  canvas.height = 720
  const ctx = canvas.getContext('2d')
  if (!ctx)
    return

  ctx.fillStyle = '#f3f6fb'
  ctx.fillRect(0, 0, canvas.width, canvas.height)
  drawRoundRect(ctx, 48, 42, 1024, 636, 28)
  ctx.fillStyle = '#ffffff'
  ctx.fill()

  ctx.fillStyle = '#111827'
  ctx.font = '700 34px Arial, sans-serif'
  drawCanvasText(ctx, tenant.tenantName, 88, 104, 720)
  ctx.fillStyle = '#6b7280'
  ctx.font = '400 18px Arial, sans-serif'
  ctx.fillText(`租户标识：${tenant.tenantId}`, 88, 140)

  const items = loginAddressItems.value
  for (let index = 0; index < items.length; index += 1) {
    const item = items[index]
    const y = 190 + index * 230
    drawRoundRect(ctx, 88, y, 944, 190, 20)
    ctx.fillStyle = '#f9fafb'
    ctx.fill()
    ctx.strokeStyle = '#e5e7eb'
    ctx.lineWidth = 1
    ctx.stroke()

    const tagX = 118
    const tagY = y + 32
    const tagWidth = 106
    const tagHeight = 34
    ctx.fillStyle = item.key === 'platform' ? '#1677ff' : '#13ad74'
    drawRoundRect(ctx, tagX, tagY, tagWidth, tagHeight, tagHeight / 2)
    ctx.fill()
    ctx.fillStyle = '#ffffff'
    ctx.font = '700 17px Arial, sans-serif'
    ctx.textAlign = 'center'
    ctx.textBaseline = 'middle'
    ctx.fillText(item.title, tagX + tagWidth / 2, tagY + tagHeight / 2)
    ctx.textAlign = 'start'
    ctx.textBaseline = 'alphabetic'

    ctx.fillStyle = '#111827'
    ctx.font = '700 22px Arial, sans-serif'
    ctx.fillText(item.url ? item.url : '暂未配置登录域名', 118, y + 102)
    ctx.fillStyle = '#6b7280'
    ctx.font = '400 16px Arial, sans-serif'
    drawCanvasText(ctx, item.description, 118, y + 135, 620)

    if (item.qr) {
      const image = await loadCanvasImage(item.qr)
      ctx.drawImage(image, 848, y + 30, 128, 128)
    }
  }

  ctx.fillStyle = '#9ca3af'
  ctx.font = '400 15px Arial, sans-serif'
  ctx.fillText('请使用对应账号从对应入口登录，跨域名禁止登录', 88, 636)

  const link = document.createElement('a')
  const safeName = tenant.tenantName.replace(/[\\/:*?"<>|\s]+/g, '-')
  link.download = `${safeName || tenant.tenantId}-登录地址.png`
  link.href = canvas.toDataURL('image/png')
  link.click()
}

function handleSearch() {
  pagination.current = 1
  loadTenants()
}

onMounted(() => {
  loadTenants()
  loadInstitutionOptions()
  loadVersionOptions()
})
</script>

<template>
  <div class="tenant-page">
    <div class="tenant-toolbar">
      <div>
        <h1>租户管理</h1>
        <p>管理合作客户的独立售卖环境、子总控账号、机构归属、域名和版本授权。</p>
      </div>
      <div class="tenant-toolbar__actions">
        <a-input
          v-model:value="keyword"
          allow-clear
          placeholder="搜索客户名称、租户标识、域名、管理员"
          class="tenant-toolbar__search"
          @press-enter="handleSearch"
        >
          <template #prefix><SearchOutlined /></template>
        </a-input>
        <a-select v-model:value="statusFilter" class="tenant-toolbar__status" @change="pagination.current = 1">
          <a-select-option value="all">全部客户</a-select-option>
          <a-select-option value="active">正常运营</a-select-option>
          <a-select-option value="disabled">已停用</a-select-option>
          <a-select-option value="incomplete">待完善</a-select-option>
        </a-select>
        <a-button @click="loadTenants">刷新</a-button>
        <a-button type="primary" @click="openCreateModal">
          <template #icon><PlusOutlined /></template>
          开通客户
        </a-button>
      </div>
    </div>

    <div class="tenant-metrics tenant-metrics--compact">
      <div class="metric-item">
        <ShopOutlined class="metric-item__icon metric-item__icon--blue" />
        <span>合作客户</span>
        <strong>{{ partnerRows.length }}</strong>
        <small>{{ activeTenantCount }} 正常</small>
      </div>
      <div class="metric-item">
        <ApartmentOutlined class="metric-item__icon metric-item__icon--green" />
        <span>下游机构</span>
        <strong>{{ institutionTotal }}</strong>
        <small>已归属</small>
      </div>
      <div class="metric-item">
        <GlobalOutlined class="metric-item__icon metric-item__icon--purple" />
        <span>独立域名</span>
        <strong>{{ domainTotal }}</strong>
        <small>访问入口</small>
      </div>
      <div class="metric-item">
        <SafetyCertificateOutlined class="metric-item__icon metric-item__icon--orange" />
        <span>待完善</span>
        <strong>{{ incompleteCount }}</strong>
        <small>缺配置</small>
      </div>
    </div>

    <a-alert
      v-if="incompleteCount"
      show-icon
      type="warning"
      class="tenant-alert"
      :message="`有 ${incompleteCount} 个客户待完善：请补齐子总控账号、独立域名或机构绑定。`"
    />

    <a-table
      row-key="tenantId"
      :columns="columns"
      :data-source="filteredRows"
      :loading="loading"
      :pagination="pagination"
      :scroll="{ x: 1510 }"
      class="tenant-table"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'tenant'">
          <div class="tenant-cell">
            <div class="tenant-cell__name">{{ record.tenantName }}</div>
            <div class="tenant-cell__meta">
              {{ editionText(record.edition) }} · 共享库
            </div>
          </div>
        </template>

        <template v-if="column.key === 'status'">
          <div class="status-stack">
            <span class="status-pill" :class="statusClass(record.status)">
              <span class="status-pill__dot" />
              {{ statusText(record.status) }}
            </span>
            <span v-if="!record.domains?.length || !record.adminUsernames?.length || !record.institutionCount" class="status-pill status-pill--pending">
              <span class="status-pill__dot" />
              待完善
            </span>
          </div>
        </template>

        <template v-if="column.key === 'institutions'">
          <div class="count-cell">
            <strong>{{ record.institutionCount || 0 }}</strong>
            <span>个机构</span>
          </div>
        </template>

        <template v-if="column.key === 'admins'">
          <div class="chip-list">
            <span v-for="username in record.adminUsernames" :key="username" class="text-chip text-chip--account">
              {{ username }}
            </span>
            <span v-if="!record.adminUsernames?.length" class="muted">未开通</span>
          </div>
        </template>

        <template v-if="column.key === 'domains'">
          <div class="domain-stack">
            <div class="domain-line">
              <span class="domain-label">总控</span>
              <span v-for="domain in record.adminDomains" :key="`admin-${domain}`" class="domain-value">
                {{ domain }}
              </span>
              <span v-if="!record.adminDomains?.length" class="muted">未配置</span>
            </div>
            <div class="domain-line">
              <span class="domain-label">机构</span>
              <span v-for="domain in record.institutionDomains" :key="`inst-${domain}`" class="domain-value domain-value--institution">
                {{ domain }}
              </span>
              <span v-if="!record.institutionDomains?.length" class="muted">未配置</span>
            </div>
          </div>
        </template>

        <template v-if="column.key === 'modules'">
          <div class="chip-list">
            <span v-for="moduleName in record.moduleNames" :key="moduleName" class="text-chip text-chip--version">
              {{ moduleName }}
            </span>
            <span v-if="!record.moduleNames?.length" class="muted">未授权版本</span>
          </div>
        </template>

        <template v-if="column.key === 'authorization'">
          <div class="count-cell">
            <strong>{{ record.menuCount || 0 }}</strong>
            <span>个菜单权限</span>
          </div>
        </template>

        <template v-if="column.key === 'action'">
          <a-space>
            <a-button type="link" size="small" @click="openEditModal(record as TenantRecord)">编辑</a-button>
            <a-button type="link" size="small" @click="openAuthorizationModal(record as TenantRecord)">授权配置</a-button>
            <a-button type="link" size="small" @click="openLoginAddressModal(record as TenantRecord)">登录地址</a-button>
          </a-space>
        </template>
      </template>
    </a-table>


    <a-modal
      v-model:open="loginAddressModalOpen"
      :width="760"
      centered
      title="登录地址"
      ok-text="下载图片"
      cancel-text="关闭"
      wrap-class-name="tenant-login-address-modal"
      @ok="downloadLoginAddressImage"
    >
      <div class="login-address-panel">
        <div class="login-address-header">
          <div>
            <strong>{{ loginAddressTenant?.tenantName || '--' }}</strong>
            <span>租户标识：{{ loginAddressTenant?.tenantId || '--' }}</span>
          </div>
          <LinkOutlined />
        </div>

        <div class="login-address-grid">
          <div v-for="item in loginAddressItems" :key="item.key" class="login-address-card">
            <div class="login-address-card__main">
              <div class="login-address-card__title">
                {{ item.title }}
              </div>
              <div class="login-address-card__desc">
                {{ item.description }}
              </div>
              <div class="login-address-card__url" :class="{ 'login-address-card__url--empty': !item.url }">
                {{ item.url || '暂未配置登录域名' }}
              </div>
              <a-space>
                <a-button size="small" :disabled="!item.url" @click="copyLoginAddress(item.url)">
                  复制地址
                </a-button>
              </a-space>
            </div>
            <div class="login-address-card__qr">
              <img v-if="item.qr" :src="item.qr" :alt="item.title">
              <span v-else>未配置</span>
            </div>
          </div>
        </div>

        <div class="login-address-tip">
          <DownloadOutlined />
          下载图片会生成包含两个入口和二维码的交付图，可直接发给客户。
        </div>
      </div>
    </a-modal>

    <a-modal
      v-model:open="modalOpen"
      :width="980"
      centered
      :confirm-loading="saving"
      ok-text="保存租户"
      cancel-text="取消"
      :body-style="{ maxHeight: '72vh', overflowY: 'auto', padding: 0 }"
      wrap-class-name="tenant-business-modal"
      @ok="handleSave"
    >
      <template #title>
        <div class="modal-title">
          <strong>{{ editingTenantId ? '编辑合作客户' : '开通合作客户' }}</strong>
          <span>配置客户租户、访问域名、子总控账号和机构归属。</span>
        </div>
      </template>

      <div class="tenant-form-shell">
        <section class="form-section">
          <div class="form-section__head">
            <div>
              <h3>客户基础信息</h3>
              <p>租户标识用于系统隔离和域名识别，创建后不建议修改。</p>
            </div>
          </div>
          <div class="form-grid">
            <a-form-item label="客户名称" required>
              <a-input v-model:value="formState.tenantName" placeholder="例如：A客户集团" />
            </a-form-item>
            <a-form-item label="租户标识" required>
              <a-input
                v-model:value="formState.tenantId"
                :disabled="Boolean(editingTenantId)"
                placeholder="例如：tenant-a"
                @blur="normalizeTenantId"
              />
            </a-form-item>
            <a-form-item label="客户版本">
              <a-select v-model:value="formState.edition">
                <a-select-option value="enterprise">企业版</a-select-option>
                <a-select-option value="professional">专业版</a-select-option>
              </a-select>
            </a-form-item>
            <a-form-item label="运营状态">
              <a-select v-model:value="formState.status">
                <a-select-option value="active">正常运营</a-select-option>
                <a-select-option value="disabled">停用</a-select-option>
              </a-select>
            </a-form-item>
            <a-form-item label="数据隔离方式">
              <a-select v-model:value="formState.isolationMode" disabled>
                <a-select-option value="shared_db">共享库</a-select-option>
              </a-select>
            </a-form-item>
            <a-form-item label="备注">
              <a-input v-model:value="formState.remark" placeholder="交付说明、商务备注等" />
            </a-form-item>
          </div>
        </section>

        <section class="form-section">
          <div class="form-section__head">
            <div>
              <h3>子总控账号</h3>
              <p>客户使用该账号进入自己的子总控后台，只能管理授权范围内的资源。</p>
            </div>
            <KeyOutlined />
          </div>
          <div class="form-grid">
            <a-form-item label="登录账号">
              <a-input v-model:value="formState.adminUsername" placeholder="例如：tenant_a_admin" />
            </a-form-item>
            <a-form-item label="初始密码">
              <a-input-password v-model:value="formState.adminPassword" placeholder="新账号留空默认 123456，老账号留空不改密码" />
            </a-form-item>
            <a-form-item label="管理员名称">
              <a-input v-model:value="formState.adminNickName" placeholder="例如：A客户管理员" />
            </a-form-item>
            <a-form-item label="手机号">
              <a-input v-model:value="formState.adminMobile" placeholder="可选" />
            </a-form-item>
          </div>
        </section>


        <section class="form-section">
          <div class="form-section__head">
            <div>
              <h3>访问域名</h3>
              <p>分别配置客户子总控后台和机构端登录入口，各填写一个域名。</p>
            </div>
            <GlobalOutlined />
          </div>
          <div class="form-grid">
            <a-form-item label="子总控后台域名">
              <a-input v-model:value="formState.adminDomain" placeholder="例如：admin.tenant-a.example.com" />
            </a-form-item>
            <a-form-item label="机构端登录域名">
              <a-input v-model:value="formState.institutionDomain" placeholder="例如：school.tenant-a.example.com" />
            </a-form-item>
          </div>
        </section>

        <section class="form-section">
          <div class="form-section__head">
            <div>
              <h3>登录页模板</h3>
              <p>子总控后台和机构端是两套独立登录页，可以选择完全不同的页面模板和品牌内容。</p>
            </div>
            <GlobalOutlined />
          </div>
          <a-tabs class="login-template-tabs">
            <a-tab-pane key="platform" tab="子总控登录页">
              <div class="login-brand-editor">
                <div class="form-grid">
                  <a-form-item label="页面模板">
                    <a-select v-model:value="formState.platformLoginBrand.template" :options="platformTemplateOptions" />
                  </a-form-item>
                  <a-form-item label="品牌名称">
                    <a-input v-model:value="formState.platformLoginBrand.brandName" placeholder="默认使用客户名称" />
                  </a-form-item>
                  <a-form-item label="登录标题">
                    <a-input v-model:value="formState.platformLoginBrand.loginTitle" placeholder="例如：肯纳集团管理后台" />
                  </a-form-item>
                  <a-form-item label="宣传标题">
                    <a-input v-model:value="formState.platformLoginBrand.heroTitle" placeholder="例如：欢迎进入肯纳集团" />
                  </a-form-item>
                </div>

                <a-form-item label="主色调" class="brand-color-form-item">
                  <div class="brand-color-picker">
                    <button
                      v-for="color in brandColorOptions"
                      :key="`platform-${color.value}`"
                      type="button"
                      class="brand-color-swatch"
                      :class="{ 'brand-color-swatch--active': formState.platformLoginBrand.primaryColor === color.value }"
                      :style="{ '--brand-color': color.value }"
                      @click="selectBrandColor('platform', color.value)"
                    >
                      <span class="brand-color-swatch__dot" />
                      <span>{{ color.label }}</span>
                    </button>
                  </div>
                </a-form-item>

                <div class="brand-asset-grid">
                  <a-form-item label="Logo">
                    <div class="brand-upload-field">
                      <a-upload
                        :custom-request="options => handleBrandAssetUpload(options, 'platform', 'logoUrl')"
                        :show-upload-list="false"
                        accept="image/*"
                        :disabled="isBrandUploading('platform', 'logoUrl')"
                      >
                        <div class="brand-upload-card brand-upload-card--logo">
                          <img v-if="formState.platformLoginBrand.logoUrl" :src="formState.platformLoginBrand.logoUrl" alt="Logo">
                          <div v-else class="brand-upload-empty">
                            <UploadOutlined />
                            <span>上传 Logo</span>
                          </div>
                          <div v-if="isBrandUploading('platform', 'logoUrl')" class="brand-upload-mask">
                            上传中 {{ getBrandUploadProgress('platform', 'logoUrl') }}%
                          </div>
                        </div>
                      </a-upload>
                      <a-button v-if="formState.platformLoginBrand.logoUrl" size="small" @click="clearBrandAsset('platform', 'logoUrl')">
                        <template #icon><DeleteOutlined /></template>
                        清除
                      </a-button>
                    </div>
                  </a-form-item>
                  <a-form-item label="登录背景">
                    <div class="brand-upload-field">
                      <a-upload
                        :custom-request="options => handleBrandAssetUpload(options, 'platform', 'backgroundUrl')"
                        :show-upload-list="false"
                        accept="image/*"
                        :disabled="isBrandUploading('platform', 'backgroundUrl')"
                      >
                        <div class="brand-upload-card brand-upload-card--banner">
                          <img v-if="formState.platformLoginBrand.backgroundUrl" :src="formState.platformLoginBrand.backgroundUrl" alt="登录背景">
                          <div v-else class="brand-upload-empty">
                            <UploadOutlined />
                            <span>上传背景图</span>
                          </div>
                          <div v-if="isBrandUploading('platform', 'backgroundUrl')" class="brand-upload-mask">
                            上传中 {{ getBrandUploadProgress('platform', 'backgroundUrl') }}%
                          </div>
                        </div>
                      </a-upload>
                      <a-button v-if="formState.platformLoginBrand.backgroundUrl" size="small" @click="clearBrandAsset('platform', 'backgroundUrl')">
                        <template #icon><DeleteOutlined /></template>
                        清除
                      </a-button>
                    </div>
                  </a-form-item>
                </div>

                <a-form-item label="宣传文案" class="brand-wide-form-item">
                  <a-input v-model:value="formState.platformLoginBrand.heroDescription" placeholder="子总控登录页展示文案" />
                </a-form-item>
              </div>
            </a-tab-pane>
            <a-tab-pane key="institution" tab="机构端登录页">
              <div class="login-brand-editor">
                <div class="form-grid">
                  <a-form-item label="页面模板">
                    <a-select v-model:value="formState.institutionLoginBrand.template" :options="institutionTemplateOptions" />
                  </a-form-item>
                  <a-form-item label="品牌名称">
                    <a-input v-model:value="formState.institutionLoginBrand.brandName" placeholder="默认使用客户名称" />
                  </a-form-item>
                  <a-form-item label="登录标题">
                    <a-input v-model:value="formState.institutionLoginBrand.loginTitle" placeholder="例如：肯纳集团机构端" />
                  </a-form-item>
                  <a-form-item label="宣传标题">
                    <a-input v-model:value="formState.institutionLoginBrand.heroTitle" placeholder="例如：校区业务管理入口" />
                  </a-form-item>
                </div>

                <a-form-item label="主色调" class="brand-color-form-item">
                  <div class="brand-color-picker">
                    <button
                      v-for="color in brandColorOptions"
                      :key="`institution-${color.value}`"
                      type="button"
                      class="brand-color-swatch"
                      :class="{ 'brand-color-swatch--active': formState.institutionLoginBrand.primaryColor === color.value }"
                      :style="{ '--brand-color': color.value }"
                      @click="selectBrandColor('institution', color.value)"
                    >
                      <span class="brand-color-swatch__dot" />
                      <span>{{ color.label }}</span>
                    </button>
                  </div>
                </a-form-item>

                <div class="brand-asset-grid">
                  <a-form-item label="Logo">
                    <div class="brand-upload-field">
                      <a-upload
                        :custom-request="options => handleBrandAssetUpload(options, 'institution', 'logoUrl')"
                        :show-upload-list="false"
                        accept="image/*"
                        :disabled="isBrandUploading('institution', 'logoUrl')"
                      >
                        <div class="brand-upload-card brand-upload-card--logo">
                          <img v-if="formState.institutionLoginBrand.logoUrl" :src="formState.institutionLoginBrand.logoUrl" alt="Logo">
                          <div v-else class="brand-upload-empty">
                            <UploadOutlined />
                            <span>上传 Logo</span>
                          </div>
                          <div v-if="isBrandUploading('institution', 'logoUrl')" class="brand-upload-mask">
                            上传中 {{ getBrandUploadProgress('institution', 'logoUrl') }}%
                          </div>
                        </div>
                      </a-upload>
                      <a-button v-if="formState.institutionLoginBrand.logoUrl" size="small" @click="clearBrandAsset('institution', 'logoUrl')">
                        <template #icon><DeleteOutlined /></template>
                        清除
                      </a-button>
                    </div>
                  </a-form-item>
                  <a-form-item label="登录背景">
                    <div class="brand-upload-field">
                      <a-upload
                        :custom-request="options => handleBrandAssetUpload(options, 'institution', 'backgroundUrl')"
                        :show-upload-list="false"
                        accept="image/*"
                        :disabled="isBrandUploading('institution', 'backgroundUrl')"
                      >
                        <div class="brand-upload-card brand-upload-card--banner">
                          <img v-if="formState.institutionLoginBrand.backgroundUrl" :src="formState.institutionLoginBrand.backgroundUrl" alt="登录背景">
                          <div v-else class="brand-upload-empty">
                            <UploadOutlined />
                            <span>上传背景图</span>
                          </div>
                          <div v-if="isBrandUploading('institution', 'backgroundUrl')" class="brand-upload-mask">
                            上传中 {{ getBrandUploadProgress('institution', 'backgroundUrl') }}%
                          </div>
                        </div>
                      </a-upload>
                      <a-button v-if="formState.institutionLoginBrand.backgroundUrl" size="small" @click="clearBrandAsset('institution', 'backgroundUrl')">
                        <template #icon><DeleteOutlined /></template>
                        清除
                      </a-button>
                    </div>
                  </a-form-item>
                </div>

                <a-form-item label="宣传文案" class="brand-wide-form-item">
                  <a-input v-model:value="formState.institutionLoginBrand.heroDescription" placeholder="机构端登录页展示文案" />
                </a-form-item>
              </div>
            </a-tab-pane>
          </a-tabs>
        </section>

        <section class="form-section">
          <div class="form-section__head">
            <div>
              <h3>机构归属</h3>
              <p>选择哪些机构归属该客户租户，机构端登录会按租户做归属校验。</p>
            </div>
            <TeamOutlined />
          </div>
          <a-form-item>
            <a-select
              v-model:value="formState.institutionIds"
              mode="multiple"
              show-search
              option-filter-prop="label"
              :loading="institutionLoading"
              :max-tag-count="3"
              :max-tag-placeholder="omittedValues => `+${omittedValues.length}`"
              placeholder="选择机构"
              :options="institutionSelectOptions"
              class="institution-owner-select"
            >
              <template #option="option">
                <div class="institution-option">
                  <span class="institution-option__name">{{ option.label }}</span>
                  <span v-if="option.tenantName" class="institution-option__tenant">
                    已归属：{{ option.tenantName }}
                  </span>
                </div>
              </template>
            </a-select>
          </a-form-item>
        </section>
      </div>
    </a-modal>

    <a-modal
      v-model:open="authorizationModalOpen"
      :width="620"
      centered
      :confirm-loading="authorizationSaving"
      ok-text="保存授权"
      cancel-text="取消"
      wrap-class-name="tenant-auth-modal"
      @ok="handleSaveAuthorization"
    >
      <template #title>
        <div class="modal-title">
          <strong>授权配置</strong>
          <span>给客户配置基础版本，授权范围会下发到该租户下属机构。</span>
        </div>
      </template>

      <div class="authorization-panel">
        <div class="authorization-tenant">
          <div>
            <span>合作客户</span>
            <strong>{{ authorizationTenant?.tenantName || '-' }}</strong>
          </div>
          <SafetyCertificateOutlined />
        </div>

        <a-form layout="vertical">
          <a-form-item label="授权版本">
            <a-select
              v-model:value="authorizationModuleId"
              allow-clear
              show-search
              option-filter-prop="label"
              :loading="versionLoading"
              placeholder="选择一个授权版本"
              :options="versionOptions.map(item => ({ value: item.id, label: `${item.name}（${item.menuCount || 0} 个菜单）` }))"
            />
          </a-form-item>
        </a-form>

        <div class="authorization-note">
          保存后，版本权限只对该租户生效；租户下属机构会按授权版本展示可用菜单。
        </div>
      </div>
    </a-modal>
  </div>
</template>

<style scoped lang="less">
.tenant-page {
  padding: 14px 16px;
}

.tenant-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  min-height: 48px;
  margin-bottom: 10px;

  h1 {
    margin: 0;
    color: rgba(0, 0, 0, 0.88);
    font-size: 20px;
    font-weight: 650;
    line-height: 28px;
  }

  p {
    margin: 2px 0 0;
    color: rgba(0, 0, 0, 0.45);
    font-size: 13px;
    line-height: 20px;
  }

  &__actions {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  &__search {
    width: 360px;
  }

  &__status {
    width: 130px;
  }
}

.tenant-metrics {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0;
  margin-bottom: 10px;
}

.tenant-metrics--compact {
  padding: 8px 12px;
  border: 1px solid rgba(5, 5, 5, 0.06);
  border-radius: 12px;
  background: #fff;
}

.metric-item {
  display: grid;
  grid-template-columns: 28px auto minmax(44px, max-content);
  grid-template-rows: 16px 18px;
  align-items: center;
  column-gap: 8px;
  min-height: 38px;
  padding: 0 12px;
  border-right: 1px solid rgba(0, 0, 0, 0.06);

  &:last-child {
    border-right: 0;
  }

  &__icon {
    grid-row: 1 / span 2;
    width: 28px;
    height: 28px;
    display: grid;
    place-items: center;
    border-radius: 8px;
    font-size: 15px;
  }

  span {
    color: rgba(0, 0, 0, 0.8);
    font-size: 12px;
    line-height: 16px;
  }

  strong {
    grid-column: 3;
    grid-row: 1 / span 2;
    color: rgba(0, 0, 0, 0.88);
    font-size: 20px;
    line-height: 1;
    text-align: right;
  }

  small {
    color: rgba(0, 0, 0, 0.35);
    font-size: 12px;
    line-height: 16px;
  }

  &__icon--blue { color: #1677ff; background: #eaf3ff; }
  &__icon--green { color: #13a86b; background: #eafaf3; }
  &__icon--purple { color: #722ed1; background: #f5edff; }
  &__icon--orange { color: #d46b08; background: #fff3e6; }
}

.tenant-alert {
  margin-bottom: 10px;
  padding: 7px 12px;
}

.tenant-table {
  border-radius: 16px;
  overflow: hidden;
  background: #fff;
}

.tenant-cell {
  &__name {
    color: rgba(0, 0, 0, 0.88);
    font-weight: 600;
  }

  &__code,
  &__meta {
    margin-top: 4px;
    color: rgba(0, 0, 0, 0.45);
    font-size: 12px;
  }
}

.count-cell {
  strong,
  span {
    display: block;
  }

  strong {
    color: rgba(0, 0, 0, 0.88);
    font-size: 18px;
    line-height: 22px;
  }

  span {
    color: rgba(0, 0, 0, 0.45);
  }
}

.status-stack {
  display: inline-grid;
  gap: 6px;
}

.status-pill {
  display: inline-flex;
  align-items: center;
  width: max-content;
  height: 24px;
  padding: 0 9px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 500;
  line-height: 22px;
  white-space: nowrap;

  &__dot {
    width: 6px;
    height: 6px;
    margin-right: 6px;
    border-radius: 50%;
    background: currentColor;
  }

  &--active {
    color: #15945b;
    background: rgba(19, 168, 107, 0.1);
  }

  &--disabled {
    color: rgba(0, 0, 0, 0.45);
    background: rgba(0, 0, 0, 0.05);
  }

  &--pending {
    color: #b26a00;
    background: rgba(250, 173, 20, 0.13);
  }
}

.chip-list {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
}

.text-chip {
  display: inline-flex;
  align-items: center;
  max-width: 190px;
  height: 24px;
  padding: 0 8px;
  border-radius: 6px;
  overflow: hidden;
  font-size: 12px;
  line-height: 24px;
  text-overflow: ellipsis;
  white-space: nowrap;

  &--account {
    color: #155fbd;
    background: #f3f8ff;
  }

  &--version {
    color: #5b2aa0;
    background: #f7f2ff;
  }
}

.domain-stack {
  display: grid;
  gap: 7px;
}

.domain-line {
  display: flex;
  align-items: center;
  gap: 7px;
  min-width: 0;
}

.domain-label {
  flex: none;
  min-width: 32px;
  color: rgba(0, 0, 0, 0.42);
  font-size: 12px;
}

.domain-value {
  min-width: 0;
  max-width: 170px;
  padding: 0 2px;
  overflow: hidden;
  color: rgba(0, 0, 0, 0.72);
  font-size: 12px;
  line-height: 20px;
  text-overflow: ellipsis;
  white-space: nowrap;

  &--institution {
    color: #17804e;
  }
}

.muted {
  color: rgba(0, 0, 0, 0.35);
}

.modal-title {
  display: grid;
  gap: 4px;

  strong {
    color: rgba(0, 0, 0, 0.88);
    font-size: 17px;
  }

  span {
    color: rgba(0, 0, 0, 0.45);
    font-size: 13px;
  }
}

.login-address-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.login-address-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 18px;
  border: 1px solid rgba(22, 119, 255, 0.12);
  border-radius: 14px;
  background: #f7fbff;

  strong {
    display: block;
    color: rgba(0, 0, 0, 0.88);
    font-size: 17px;
    font-weight: 700;
  }

  span {
    display: block;
    margin-top: 4px;
    color: rgba(0, 0, 0, 0.45);
    font-size: 13px;
  }

  > .anticon {
    color: #1677ff;
    font-size: 24px;
  }
}

.login-address-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 12px;
}

.login-address-card {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 116px;
  gap: 16px;
  align-items: center;
  padding: 16px;
  border: 1px solid #e5e7eb;
  border-radius: 14px;
  background: #fff;
}

.login-address-card__title {
  color: rgba(0, 0, 0, 0.88);
  font-size: 16px;
  font-weight: 700;
}

.login-address-card__desc {
  margin-top: 4px;
  color: rgba(0, 0, 0, 0.45);
  font-size: 13px;
  line-height: 20px;
}

.login-address-card__url {
  margin: 12px 0;
  padding: 9px 10px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  background: #f9fafb;
  color: #1677ff;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.login-address-card__url--empty {
  color: rgba(0, 0, 0, 0.35);
}

.login-address-card__qr {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 116px;
  height: 116px;
  border: 1px solid #eef0f4;
  border-radius: 12px;
  background: #fafafa;
  color: rgba(0, 0, 0, 0.35);
  font-size: 13px;

  img {
    width: 98px;
    height: 98px;
  }
}

.login-address-tip {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border-radius: 10px;
  background: rgba(22, 119, 255, 0.06);
  color: rgba(0, 0, 0, 0.56);
  font-size: 13px;

  .anticon {
    color: #1677ff;
  }
}

.tenant-form-shell {
  padding: 20px 22px 4px;
}

.form-section {
  margin-bottom: 18px;
  padding: 18px;
  border: 1px solid rgba(0, 0, 0, 0.06);
  border-radius: 14px;
  background: #fff;

  &__head {
    display: flex;
    justify-content: space-between;
    gap: 16px;
    margin-bottom: 14px;

    h3 {
      margin: 0;
      color: rgba(0, 0, 0, 0.88);
      font-size: 16px;
      font-weight: 600;
    }

    p {
      margin: 4px 0 0;
      color: rgba(0, 0, 0, 0.45);
    }

    > .anticon {
      color: #1677ff;
      font-size: 20px;
    }
  }
}


.login-template-tabs {
  :deep(.ant-tabs-nav) {
    margin: 0 0 16px;
    border-bottom: 1px solid rgba(5, 5, 5, 0.08);
  }

  :deep(.ant-tabs-nav::before) {
    border-bottom: 0;
  }

  :deep(.ant-tabs-nav-list) {
    gap: 34px;
  }

  :deep(.ant-tabs-tab) {
    margin: 0;
    padding: 12px 0;
    color: rgba(0, 0, 0, 0.88);
    font-size: 15px;
    font-weight: 400;
  }

  :deep(.ant-tabs-tab + .ant-tabs-tab) {
    margin: 0;
  }

  :deep(.ant-tabs-tab-active .ant-tabs-tab-btn) {
    color: var(--pro-ant-color-primary, #1677ff);
    font-weight: 500;
  }

  :deep(.ant-tabs-ink-bar) {
    bottom: 1px !important;
    height: 9px !important;
    background: transparent !important;

    &::after {
      position: absolute;
      top: 0;
      left: calc(50% - 12px);
      width: 24px !important;
      height: 4px !important;
      border-radius: 2px;
      background-color: var(--pro-ant-color-primary, #1677ff);
      content: '';
    }
  }

  :deep(.ant-tabs-content-holder) {
    padding-top: 0;
  }
}

.login-brand-editor {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.brand-color-form-item,
.brand-wide-form-item {
  :deep(.ant-form-item-row) {
    display: grid;
    grid-template-columns: 112px minmax(0, 1fr);
    align-items: center;
    flex-wrap: nowrap;
  }

  :deep(.ant-form-item-label) {
    max-width: 112px;
    text-align: right;

    > label {
      color: rgba(0, 0, 0, 0.65);
      font-size: 14px;
      font-weight: 500;
    }
  }
}

.brand-color-picker {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.brand-color-swatch {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 32px;
  padding: 0 12px;
  color: rgba(0, 0, 0, 0.65);
  font-size: 13px;
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;

  &:hover {
    color: var(--brand-color);
    border-color: var(--brand-color);
  }
}

.brand-color-swatch--active {
  color: var(--brand-color);
  background: color-mix(in srgb, var(--brand-color) 8%, #fff);
  border-color: var(--brand-color);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--brand-color) 12%, transparent);
}

.brand-color-swatch__dot {
  width: 14px;
  height: 14px;
  background: var(--brand-color);
  border-radius: 50%;
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.55);
}

.brand-asset-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 16px;

  :deep(.ant-form-item) {
    margin-bottom: 14px;
  }

  :deep(.ant-form-item-row) {
    display: grid;
    grid-template-columns: 112px minmax(0, 1fr);
    align-items: flex-start;
    flex-wrap: nowrap;
  }

  :deep(.ant-form-item-label) {
    max-width: 112px;
    padding-top: 10px;
    text-align: right;

    > label {
      color: rgba(0, 0, 0, 0.65);
      font-size: 14px;
      font-weight: 500;
    }
  }
}

.brand-upload-field {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.brand-upload-card {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fafafa;
  border: 1px dashed #d9d9d9;
  border-radius: 10px;
  cursor: pointer;
  overflow: hidden;
  transition: all 0.2s ease;

  &:hover {
    border-color: var(--pro-ant-color-primary, #1677ff);
    background: #f7fbff;
  }

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
}

.brand-upload-card--logo {
  width: 96px;
  height: 64px;

  img {
    object-fit: contain;
    padding: 8px;
    background: #fff;
  }
}

.brand-upload-card--banner {
  width: 168px;
  height: 76px;
}

.brand-upload-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  color: rgba(0, 0, 0, 0.45);
  font-size: 13px;

  .anticon {
    color: var(--pro-ant-color-primary, #1677ff);
    font-size: 18px;
  }
}

.brand-upload-mask {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 13px;
  background: rgba(0, 0, 0, 0.45);
}

.institution-owner-select {
  width: 100%;

  :deep(.ant-select-selector) {
    flex-wrap: nowrap;
    min-height: 34px;
    overflow: hidden;
  }

  :deep(.ant-select-selection-overflow) {
    flex-wrap: nowrap;
    overflow: hidden;
  }

  :deep(.ant-select-selection-overflow-item) {
    flex: none;
  }

  :deep(.ant-select-selection-item) {
    max-width: 230px;
  }
}

.institution-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  width: 100%;
  min-width: 0;
}

.institution-option__name {
  min-width: 0;
  overflow: hidden;
  color: rgba(0, 0, 0, 0.88);
  font-weight: 500;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.institution-option__tenant {
  flex: none;
  color: rgba(0, 0, 0, 0.45);
  font-size: 12px;
}

:deep(.ant-select-item-option-disabled) {
  .institution-option__name,
  .institution-option__tenant {
    color: rgba(0, 0, 0, 0.25);
  }
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 16px;
  align-items: center;

  :deep(.ant-form-item) {
    margin-bottom: 14px;
  }

  :deep(.ant-form-item-row) {
    display: grid;
    grid-template-columns: 112px minmax(0, 1fr);
    align-items: center;
    flex-wrap: nowrap;
  }

  :deep(.ant-form-item-label) {
    max-width: 112px;
    padding: 0 10px 0 0;
    text-align: right;
    white-space: nowrap;
  }

  :deep(.ant-form-item-label > label) {
    white-space: nowrap;
  }

  :deep(.ant-form-item-control) {
    min-width: 0;
  }

  :deep(.ant-input),
  :deep(.ant-input-affix-wrapper),
  :deep(.ant-select) {
    width: 100%;
  }
}

.form-grid__full-row {
  grid-column: 1 / -1;
}


.authorization-tenant {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 18px;
  padding: 14px 16px;
  border: 1px solid rgba(22, 119, 255, 0.12);
  border-radius: 12px;
  background: #f7fbff;

  span {
    display: block;
    margin-bottom: 4px;
    color: rgba(0, 0, 0, 0.45);
    font-size: 12px;
  }

  strong {
    color: rgba(0, 0, 0, 0.88);
    font-size: 16px;
    font-weight: 600;
  }

  > .anticon {
    color: #1677ff;
    font-size: 22px;
  }
}

.authorization-note {
  margin-top: 10px;
  padding: 10px 12px;
  border-radius: 8px;
  color: rgba(0, 0, 0, 0.56);
  font-size: 13px;
  line-height: 20px;
  background: rgba(0, 0, 0, 0.025);
}

@media (max-width: 1280px) {
  .tenant-toolbar {
    flex-direction: column;
  }

  .tenant-metrics {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
