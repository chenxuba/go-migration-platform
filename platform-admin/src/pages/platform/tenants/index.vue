<script setup lang="ts">
import type { TableColumnsType } from 'ant-design-vue'
import { computed, onMounted, reactive, ref } from 'vue'
import {
  ApartmentOutlined,
  CheckCircleOutlined,
  GlobalOutlined,
  KeyOutlined,
  PlusOutlined,
  SafetyCertificateOutlined,
  SearchOutlined,
  ShopOutlined,
  TeamOutlined,
} from '@ant-design/icons-vue'
import messageService from '@/utils/messageService'
import { listTenantsApi, saveTenantApi } from '@/api/platform/tenants'
import { pageInstitutionsApi } from '@/api/platform/institutions'
import { pageVersionsApi } from '@/api/platform/versions'

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
}

interface InstitutionOption {
  id: number
  organName: string
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
  remark: string
}

const loading = ref(false)
const saving = ref(false)
const modalOpen = ref(false)
const authorizationModalOpen = ref(false)
const authorizationSaving = ref(false)
const authorizationTenant = ref<TenantRecord | null>(null)
const authorizationModuleId = ref<number | undefined>(undefined)
const editingTenantId = ref('')
const keyword = ref('')
const statusFilter = ref<'all' | 'active' | 'disabled' | 'incomplete'>('all')
const tenantRows = ref<TenantRecord[]>([])
const institutionOptions = ref<InstitutionOption[]>([])
const institutionLoading = ref(false)
const versionOptions = ref<VersionOption[]>([])
const versionLoading = ref(false)

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
  { title: '操作', key: 'action', width: 180, fixed: 'right' as const },
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
      :scroll="{ x: 1450 }"
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
          </a-space>
        </template>
      </template>
    </a-table>

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
              placeholder="选择机构"
              :options="institutionOptions.map(item => ({ value: item.id, label: `${item.organName}（ID:${item.id}）` }))"
            />
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
          <span>只配置该客户可售卖、可分配给机构的唯一版本包。</span>
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
          保存后，该客户只能基于这个授权版本给下属机构开通或续费。
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

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 16px;

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
