<script setup lang="ts">
import type { TableColumnsType } from 'ant-design-vue'
import { DownOutlined, PlusOutlined } from '@ant-design/icons-vue'
import { computed, onMounted, reactive, ref, watch } from 'vue'
import AllFilter from '@/components/common/all-filter.vue'
import {
  pageInstitutionsApi,
  updateInstitutionStatusApi,
  type InstitutionItem,
  type InstitutionSummary,
} from '@/api/platform/institutions'
import { regionData } from '@/constants/region-data'
import InstitutionFormDrawer from './components/institution-form-drawer.vue'
import InstitutionLoginBrandModal from './components/institution-login-brand-modal.vue'
import InstitutionPermissionBatchModal from './components/institution-permission-batch-modal.vue'
import InstitutionPermissionModal from './components/institution-permission-modal.vue'
import InstitutionRenewalModal from './components/institution-renewal-modal.vue'
import InstitutionVersionModal from './components/institution-version-modal.vue'
import messageService from '@/utils/messageService'

const displayArray = ['customSearch', 'createTime', 'enableStatus']
const institutionStatusOptions = [
  { id: 1, value: '启用' },
  { id: 3, value: '预警' },
  { id: 2, value: '停用' },
  { id: 4, value: '过期' },
]
const institutionOpenTypeOptions = [
  { id: 1, value: '体验版' },
  { id: 2, value: '基础版' },
  { id: 3, value: '高级版' },
  { id: 4, value: '旗舰版' },
]
const directCountyCityLabels = new Set([
  '市辖区',
  '县',
  '省直辖县级行政区划',
  '自治区直辖县级行政区划',
])

const listLoading = ref(false)
const statusSubmittingId = ref<number | null>(null)
const dataSource = ref<InstitutionItem[]>([])
const institutionDrawerOpen = ref(false)
const editingInstitutionId = ref<number | null>(null)
const institutionPermissionOpen = ref(false)
const permissionInstitutionId = ref<number | null>(null)
const institutionPermissionBatchOpen = ref(false)
const selectedInstitutionIds = ref<number[]>([])
const institutionVersionOpen = ref(false)
const versionInstitutionId = ref<number | null>(null)
const institutionRenewalOpen = ref(false)
const renewingInstitutionId = ref<number | null>(null)
const institutionLoginBrandOpen = ref(false)
const loginBrandInstitutionId = ref<number | null>(null)
const summary = ref<InstitutionSummary>({
  totalCount: 0,
  enabledCount: 0,
  disabledCount: 0,
})

const filters = reactive<{
  keyword?: string
  mobile?: string
  registerTime?: string[]
  status?: number
  openType?: number
  provinceCode?: string
  cityCode?: string
  regionCode?: string
}>({
  keyword: undefined,
  mobile: undefined,
  registerTime: undefined,
  status: undefined,
  openType: undefined,
  provinceCode: undefined,
  cityCode: undefined,
  regionCode: undefined,
})

const provinceFilterOptions = computed(() => regionData.map(item => ({
  id: item.value,
  value: item.label,
})))
const selectedProvinceFilter = computed(() => regionData.find(item => item.value === filters.provinceCode))
const cityFilterOptions = computed(() => (selectedProvinceFilter.value?.children || []).map(item => ({
  id: item.value,
  value: item.label,
})))
const selectedCityFilter = computed(() => selectedProvinceFilter.value?.children?.find(item => item.value === filters.cityCode))
const regionFilterOptions = computed(() => (selectedCityFilter.value?.children || []).map(item => ({
  id: item.value,
  value: item.label,
})))

const customSearchFilters = computed(() => [
  {
    id: 'keyword',
    fieldKey: '机构名称',
    fieldType: 1,
  },
  {
    id: 'mobile',
    fieldKey: '联系电话',
    fieldType: 1,
  },
  {
    id: 'openType',
    fieldKey: '开通版本',
    fieldType: 4,
    optionsList: institutionOpenTypeOptions,
  },
  {
    id: 'provinceCode',
    fieldKey: '省份',
    fieldType: 4,
    optionsList: provinceFilterOptions.value,
  },
  {
    id: 'cityCode',
    fieldKey: '城市',
    fieldType: 4,
    optionsList: cityFilterOptions.value,
  },
  {
    id: 'regionCode',
    fieldKey: '区县',
    fieldType: 4,
    optionsList: regionFilterOptions.value,
  },
])

const customSearchValues = computed(() => ({
  keyword: filters.keyword ?? '',
  mobile: filters.mobile ?? '',
  openType: filters.openType ?? '',
  provinceCode: filters.provinceCode ?? '',
  cityCode: filters.cityCode ?? '',
  regionCode: filters.regionCode ?? '',
}))

const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
  showSizeChanger: true,
  showTotal: (total: number) => `共 ${total} 家机构`,
})

const columns: TableColumnsType<InstitutionItem> = [
  {
    title: '机构信息',
    dataIndex: 'organName',
    key: 'organName',
    width: 220,
    fixed: 'left' as const,
  },
  {
    title: '所属租户',
    key: 'tenant',
    width: 160,
  },
  {
    title: '账号信息',
    key: 'account',
    width: 180,
  },
  {
    title: '联系方式',
    key: 'contact',
    width: 220,
  },
  {
    title: '开通版本',
    key: 'openType',
    width: 160,
  },
  {
    title: '注册时间',
    key: 'registerTime',
    width: 180,
  },
  {
    title: '过期时间',
    key: 'expireEndTime',
    width: 160,
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 140,
    align: 'center' as const,
  },
  {
    title: '操作',
    key: 'action',
    width: 220,
    fixed: 'right' as const,
  },
]

let requestSerial = 0

function resetFilters() {
  filters.keyword = undefined
  filters.mobile = undefined
  filters.registerTime = undefined
  filters.status = undefined
  filters.openType = undefined
  filters.provinceCode = undefined
  filters.cityCode = undefined
  filters.regionCode = undefined
}

function normalizeLogo(logo?: string) {
  const value = String(logo || '').trim()
  if (!value)
    return undefined

  if (/^(https?:)?\/\//.test(value) || value.startsWith('/') || value.startsWith('data:'))
    return value

  return undefined
}

function getInitial(name?: string) {
  return String(name || '').trim().slice(0, 1) || '机'
}

function buildInstitutionAddress(record: Partial<InstitutionItem>) {
  const province = String(record.province || '').trim()
  const rawCity = String(record.city || '').trim()
  const city = !rawCity || rawCity === province || directCountyCityLabels.has(rawCity) ? '' : rawCity
  const segments = [
    province,
    city,
    String(record.region || '').trim(),
    String(record.address || '').trim(),
  ].filter(Boolean)

  const deduped: string[] = []
  for (const segment of segments) {
    if (deduped[deduped.length - 1] === segment)
      continue
    deduped.push(segment)
  }

  return deduped.join('')
}

function formatDateMinute(value?: string) {
  const raw = String(value || '').trim()
  if (!raw)
    return '--'

  if (raw.length >= 16)
    return raw.slice(0, 16)

  return raw
}

function getInstitutionOpenTypeLabel(record: Partial<InstitutionItem>) {
  const currentModuleName = String(record.currentModuleName || '').trim()
  if (currentModuleName)
    return currentModuleName

  const currentValue = Number(record.openType || 0)
  return institutionOpenTypeOptions.find(item => item.id === currentValue)?.value || '基础版'
}

function getInstitutionExpireAt(record: Partial<InstitutionItem>) {
  const raw = String(record.expireEndTime || '').trim()
  if (!raw)
    return null
  const expireAt = new Date(raw.replace(/-/g, '/'))
  if (Number.isNaN(expireAt.getTime()))
    return null
  return expireAt
}

function getInstitutionStatusValue(record: Partial<InstitutionItem>) {
  if (record.enabled === false)
    return 2

  if (isInstitutionExpiredByTime(record))
    return 4

  if (isInstitutionWarningByTime(record))
    return 3

  if (record.enabled === true)
    return 1

  const rawStatus = Number(record.status || 0)
  return [1, 2, 3, 4].includes(rawStatus) ? rawStatus : 1
}

function getInstitutionStatusMeta(record: Partial<InstitutionItem>) {
  const status = getInstitutionStatusValue(record)
  if (status === 1) {
    return {
      value: 1,
      text: '启用',
      className: 'status-text--enabled',
    }
  }
  if (status === 4) {
    return {
      value: 4,
      text: '过期',
      className: 'status-text--expired',
    }
  }
  if (status === 3) {
    return {
      value: 3,
      text: '预警',
      className: 'status-text--warning',
    }
  }
  return {
    value: 2,
    text: '停用',
    className: 'status-text--disabled',
  }
}

function isInstitutionExpired(record: Partial<InstitutionItem>) {
  return getInstitutionStatusValue(record) === 4
}

function isInstitutionExpiredByTime(record: Partial<InstitutionItem>) {
  const expireAt = getInstitutionExpireAt(record)
  if (!expireAt)
    return false
  return expireAt.getTime() < Date.now()
}

function isInstitutionWarningByTime(record: Partial<InstitutionItem>) {
  const expireAt = getInstitutionExpireAt(record)
  if (!expireAt)
    return false
  const now = new Date()
  if (expireAt.getTime() < now.getTime())
    return false
  const warningDeadline = new Date(now)
  warningDeadline.setMonth(warningDeadline.getMonth() + 1)
  return expireAt.getTime() <= warningDeadline.getTime()
}

function canToggleInstitutionStatus(record: Partial<InstitutionItem>) {
  const status = getInstitutionStatusValue(record)
  return status === 1 || status === 2 || status === 3 || status === 4
}

function getToggleTargetEnabled(record: Partial<InstitutionItem>) {
  return getInstitutionStatusValue(record) === 2
}

function getRowClassName(record: Partial<InstitutionItem>) {
  const status = getInstitutionStatusValue(record)
  return status === 2 || status === 4 ? 'institution-row institution-row--disabled' : 'institution-row'
}

function openCreateDrawer() {
  editingInstitutionId.value = null
  institutionDrawerOpen.value = true
}

function openEditDrawer(record: Partial<InstitutionItem>) {
  editingInstitutionId.value = Number(record.id || 0) || null
  institutionDrawerOpen.value = true
}

function openRenewalModal(record: Partial<InstitutionItem>) {
  renewingInstitutionId.value = Number(record.id || 0) || null
  institutionRenewalOpen.value = true
}

function openPermissionModal(record: Partial<InstitutionItem>) {
  permissionInstitutionId.value = Number(record.id || 0) || null
  institutionPermissionOpen.value = true
}

function openBatchPermissionModal() {
  if (!selectedInstitutionIds.value.length) {
    messageService.warning('请先勾选机构后再批量配置权限')
    return
  }
  institutionPermissionBatchOpen.value = true
}

function openVersionModal(record: Partial<InstitutionItem>) {
  versionInstitutionId.value = Number(record.id || 0) || null
  institutionVersionOpen.value = true
}

function openLoginBrandModal(record: Partial<InstitutionItem>) {
  loginBrandInstitutionId.value = Number(record.id || 0) || null
  institutionLoginBrandOpen.value = true
}

function clearSelectedInstitutions() {
  selectedInstitutionIds.value = []
}

function handleDrawerSaved() {
  institutionDrawerOpen.value = false
  editingInstitutionId.value = null
  fetchInstitutions()
}

function handleRenewalSaved() {
  fetchInstitutions()
}

function handlePermissionSaved() {
  fetchInstitutions()
}

function handleBatchPermissionSaved() {
  institutionPermissionBatchOpen.value = false
  fetchInstitutions()
}

function handleVersionSaved() {
  fetchInstitutions()
}

function handleLoginBrandSaved() {
  fetchInstitutions()
}

async function toggleInstitutionStatus(record: Partial<InstitutionItem>, enabled: boolean) {
  const id = Number(record.id || 0)
  if (!id)
    return

  statusSubmittingId.value = id
  try {
    const res = await updateInstitutionStatusApi({ id, enabled })
    if (res.code !== 200) {
      messageService.error(res.message || '更新机构状态失败')
      return
    }
    const nextMessage = enabled
      ? (
          isInstitutionExpiredByTime(record)
            ? '机构已启用，当前状态为过期'
            : (isInstitutionWarningByTime(record) ? '机构已启用，当前状态为预警' : '机构已启用')
        )
      : '机构已停用'
    messageService.success(nextMessage)
    await fetchInstitutions()
  }
  catch (error: any) {
    console.error('toggle institution status failed', error)
    messageService.error(error?.message || '更新机构状态失败')
  }
  finally {
    statusSubmittingId.value = null
  }
}

async function fetchInstitutions() {
  const currentRequest = ++requestSerial
  listLoading.value = true
  try {
    const response = await pageInstitutionsApi({
      current: pagination.current,
      size: pagination.pageSize,
      keyword: filters.keyword,
      mobile: filters.mobile,
      registerTimeBegin: filters.registerTime?.[0],
      registerTimeEnd: filters.registerTime?.[1],
      status: filters.status,
      openType: filters.openType,
      provinceCode: filters.provinceCode ? Number(filters.provinceCode) : undefined,
      cityCode: filters.cityCode ? Number(filters.cityCode) : undefined,
      regionCode: filters.regionCode ? Number(filters.regionCode) : undefined,
    })

    if (currentRequest !== requestSerial)
      return

    if (response.code !== 200) {
      messageService.error(response.message || '获取机构列表失败')
      return
    }

    const list = Array.isArray(response.result) ? response.result : []
    dataSource.value = list
    pagination.total = Number(response.total || 0)
    summary.value = response.data?.summary || {
      totalCount: pagination.total,
      enabledCount: list.filter(item => [1, 3].includes(getInstitutionStatusValue(item))).length,
      disabledCount: list.filter(item => [2, 4].includes(getInstitutionStatusValue(item))).length,
    }
  }
  catch (error: any) {
    if (currentRequest !== requestSerial)
      return
    console.error('fetch institutions failed', error)
    messageService.error(error?.message || '获取机构列表失败')
  }
  finally {
    if (currentRequest === requestSerial)
      listLoading.value = false
  }
}

function handleTableChange(page: { current?: number, pageSize?: number }) {
  pagination.current = page.current || 1
  pagination.pageSize = page.pageSize || 20
  fetchInstitutions()
}

const rowSelection = computed(() => {
  return {
    selectedRowKeys: selectedInstitutionIds.value,
    preserveSelectedRowKeys: true,
    onChange: (keys: Array<string | number>) => {
      selectedInstitutionIds.value = keys
        .map(key => Number(key))
        .filter(key => Number.isFinite(key) && key > 0)
    },
  }
})

const filterUpdateHandlers = {
  'update:customSearchInputFilter': (payload: any, isClearAll: boolean, id?: string) => {
    if (isClearAll) {
      resetFilters()
    }
    else {
      const fieldId = id || payload?.item?.id
      const value = String(payload?.value ?? '').trim() || undefined

      if (fieldId === 'keyword')
        filters.keyword = value

      if (fieldId === 'mobile')
        filters.mobile = value

      if (fieldId === 'openType')
        filters.openType = value ? Number(value) : undefined

      if (fieldId === 'provinceCode') {
        filters.provinceCode = value
        filters.cityCode = undefined
        filters.regionCode = undefined
      }

      if (fieldId === 'cityCode') {
        filters.cityCode = value
        filters.regionCode = undefined
      }

      if (fieldId === 'regionCode')
        filters.regionCode = value
    }

    pagination.current = 1
    fetchInstitutions()
  },
  'update:createTimeFilter': (value: string[] | undefined, isClearAll: boolean) => {
    if (isClearAll) {
      resetFilters()
    }
    else {
      const normalized = Array.isArray(value)
        ? value.map(item => String(item || '').trim()).filter(Boolean)
        : []
      filters.registerTime = normalized.length === 2 ? normalized : undefined
    }

    pagination.current = 1
    fetchInstitutions()
  },
  'update:enableStatusFilter': (value: number | undefined, isClearAll: boolean) => {
    if (isClearAll) {
      resetFilters()
    }
    else {
      filters.status = value ? Number(value) : undefined
    }

    pagination.current = 1
    fetchInstitutions()
  },
}

onMounted(() => {
  fetchInstitutions()
})

watch(institutionDrawerOpen, (open) => {
  if (!open)
    editingInstitutionId.value = null
})

watch(institutionPermissionOpen, (open) => {
  if (!open)
    permissionInstitutionId.value = null
})

watch(institutionPermissionBatchOpen, (open) => {
  if (!open)
    selectedInstitutionIds.value = []
})

watch(institutionVersionOpen, (open) => {
  if (!open)
    versionInstitutionId.value = null
})

watch(institutionRenewalOpen, (open) => {
  if (!open)
    renewingInstitutionId.value = null
})
</script>

<template>
  <div class="organization-page">
    <div class="filter-wrap">
      <AllFilter
        :display-array="displayArray"
        :is-quick-show="false"
        create-time-label="注册时间"
        :custom-is-display-list="customSearchFilters"
        :custom-search-values="customSearchValues"
        enable-status-label="机构状态"
        :enable-status-options-override="institutionStatusOptions"
        v-on="filterUpdateHandlers"
      />
    </div>

    <div class="organization-list">
      <div class="table-title">
        <div class="table-title__left">
          <div class="total">
            共 {{ summary.totalCount || pagination.total }} 家机构
          </div>
        </div>

        <div class="table-title__actions">
          <a-button :disabled="!selectedInstitutionIds.length" @click="openBatchPermissionModal">
            批量权限配置
          </a-button>
          <a-button type="link" :disabled="!selectedInstitutionIds.length" @click="clearSelectedInstitutions">
            清空已选
          </a-button>
          <a-button type="primary" @click="openCreateDrawer">
            <template #icon>
              <PlusOutlined />
            </template>
            新建机构
          </a-button>
        </div>
      </div>

      <div class="table-content">
        <a-table
          class="organization-table"
          :columns="columns"
          :data-source="dataSource"
          :loading="listLoading"
          :pagination="pagination"
          :row-selection="rowSelection"
          :scroll="{ x: 1520 }"
          :row-class-name="getRowClassName"
          row-key="id"
          size="small"
          @change="handleTableChange"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'organName'">
              <div class="org-cell">
                <a-avatar :src="normalizeLogo(record.logo)" class="org-cell__avatar">
                  {{ getInitial(record.organName) }}
                </a-avatar>
                <div class="org-cell__meta">
                  <div class="cell-title">
                    {{ record.organName || '--' }}
                  </div>
                  <div class="cell-sub cell-sub--light">
                    负责人：{{ record.principal || '--' }}
                  </div>
                </div>
              </div>
            </template>

            <template v-else-if="column.key === 'tenant'">
              <div class="info-cell">
                <div class="cell-title cell-title--sm">
                  {{ record.tenantName || '--' }}
                </div>
                <div class="cell-sub">
                  {{ record.tenantId || '未绑定租户' }}
                </div>
              </div>
            </template>

            <template v-else-if="column.key === 'account'">
              <div class="info-cell">
                <div class="cell-title cell-title--sm">
                  {{ record.loginName || '--' }}
                </div>
                <div class="cell-sub">
                  登录账号
                </div>
              </div>
            </template>

            <template v-else-if="column.key === 'contact'">
              <div class="info-cell">
                <div class="cell-title cell-title--sm">
                  {{ record.mobile || '--' }}
                </div>
                <a-tooltip v-if="buildInstitutionAddress(record)" :title="buildInstitutionAddress(record)" placement="topLeft">
                  <div class="cell-sub contact-address">
                    {{ buildInstitutionAddress(record) }}
                  </div>
                </a-tooltip>
                <div v-else class="cell-sub contact-address">
                  未填写机构地址
                </div>
              </div>
            </template>

            <template v-else-if="column.key === 'openType'">
              <div class="info-cell">
                <div class="cell-title cell-title--sm">
                  {{ getInstitutionOpenTypeLabel(record) }}
                </div>
                <div class="cell-sub">
                  开通版本
                </div>
              </div>
            </template>

            <template v-else-if="column.key === 'registerTime'">
              <div class="info-cell">
                <div class="cell-title cell-title--sm">
                  {{ formatDateMinute(record.registerTime) }}
                </div>
                <div class="cell-sub">
                  注册时间
                </div>
              </div>
            </template>

            <template v-else-if="column.key === 'expireEndTime'">
              <div class="info-cell">
                <div class="cell-title cell-title--sm" :class="{ 'cell-title--expired': isInstitutionExpiredByTime(record) }">
                  {{ formatDateMinute(record.expireEndTime) }}
                </div>
                <div class="cell-sub">
                  账号到期时间
                </div>
              </div>
            </template>

            <template v-else-if="column.key === 'status'">
              <span class="status-text" :class="getInstitutionStatusMeta(record).className">
                <span class="status-text__dot" />
                {{ getInstitutionStatusMeta(record).text }}
              </span>
            </template>

            <template v-else-if="column.key === 'action'">
              <div class="action-cell action-cell--text">
                <a class="action-link" @click="openEditDrawer(record)">
                  编辑
                </a>
                <a class="action-link" @click="openRenewalModal(record)">
                  续期
                </a>
                <a-popconfirm
                  v-if="canToggleInstitutionStatus(record)"
                  :title="getToggleTargetEnabled(record) ? '确定启用该机构？' : '确定停用该机构？'"
                  ok-text="确定"
                  cancel-text="取消"
                  @confirm="toggleInstitutionStatus(record, getToggleTargetEnabled(record))"
                >
                  <a class="action-link" :class="getToggleTargetEnabled(record) ? 'action-link--success' : 'action-link--danger'">
                    {{ getToggleTargetEnabled(record) ? '启用' : '停用' }}
                  </a>
                </a-popconfirm>
                <a-dropdown placement="bottomRight" :trigger="['click']">
                  <a class="action-link action-more-link">
                    更多
                    <DownOutlined class="action-more-link__arrow" />
                  </a>
                  <template #overlay>
                    <a-menu class="action-more-menu">
                      <a-menu-item key="permission" @click="openPermissionModal(record)">
                        机构权限
                      </a-menu-item>
                      <a-menu-item key="version" @click="openVersionModal(record)">
                        切换版本
                      </a-menu-item>
                      <a-menu-item key="login-brand" @click="openLoginBrandModal(record)">
                        独立登录页
                      </a-menu-item>
                    </a-menu>
                  </template>
                </a-dropdown>
              </div>
            </template>
          </template>
        </a-table>
      </div>
    </div>

    <InstitutionFormDrawer
      v-model:open="institutionDrawerOpen"
      :institution-id="editingInstitutionId"
      @saved="handleDrawerSaved"
    />
    <InstitutionRenewalModal
      v-model:open="institutionRenewalOpen"
      :institution-id="renewingInstitutionId"
      @renewed="handleRenewalSaved"
    />
    <InstitutionPermissionModal
      v-model:open="institutionPermissionOpen"
      :institution-id="permissionInstitutionId"
      @saved="handlePermissionSaved"
    />
    <InstitutionPermissionBatchModal
      v-model:open="institutionPermissionBatchOpen"
      :institution-ids="selectedInstitutionIds"
      @saved="handleBatchPermissionSaved"
    />
    <InstitutionVersionModal
      v-model:open="institutionVersionOpen"
      :institution-id="versionInstitutionId"
      @saved="handleVersionSaved"
    />
    <InstitutionLoginBrandModal
      v-model:open="institutionLoginBrandOpen"
      :institution-id="loginBrandInstitutionId"
      @saved="handleLoginBrandSaved"
    />
  </div>
</template>

<style scoped>
.organization-page {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.filter-wrap,
.organization-list {
  background: #fff;
  border: 1px solid #e9edf3;
  border-radius: 16px;
}

.filter-wrap {
  padding: 0 12px;
}

.organization-list {
  padding: 16px 24px 8px;
  overflow: hidden;
}

.table-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 12px;
}

.table-title__left {
  min-width: 0;
}

.table-title__actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.total {
  position: relative;
  padding-left: 10px;
  color: #262626;
  display: flex;
  align-items: center;
  font-size: 14px;
  font-weight: 600;
  line-height: 24px;
}

.total::before {
  position: absolute;
  left: 0;
  top: 6px;
  width: 4px;
  height: 12px;
  border-radius: 2px;
  background: var(--pro-ant-color-primary);
  content: "";
}

.table-content {
  min-width: 0;
}

.org-cell {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.org-cell__avatar {
  flex-shrink: 0;
  background: linear-gradient(135deg, #1677ff 0%, #69b1ff 100%);
  font-weight: 700;
}

.org-cell__meta,
.info-cell {
  min-width: 0;
}

.cell-title {
  color: #262626;
  font-size: 14px;
  line-height: 22px;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.cell-title--sm {
  font-size: 13px;
  font-weight: 500;
}

.cell-title--expired {
  color: #ff4d4f;
}

.cell-sub {
  color: #8c8c8c;
  font-size: 12px;
  line-height: 20px;
  word-break: break-all;
}

.cell-sub--light {
  color: #b0b0b0;
}

.contact-address {
  display: block;
  width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  word-break: normal;
}

.status-text {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 500;
}

.status-text__dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.status-text--enabled {
  color: #1677ff;
}

.status-text--enabled .status-text__dot {
  background: #1677ff;
}

.status-text--disabled {
  color: #8c8c8c;
}

.status-text--disabled .status-text__dot {
  background: #bfbfbf;
}

.status-text--warning {
  color: #d97706;
}

.status-text--warning .status-text__dot {
  background: #f59e0b;
}

.status-text--expired {
  color: #ff4d4f;
}

.status-text--expired .status-text__dot {
  background: #ff4d4f;
}

.action-cell {
  display: flex;
  align-items: center;
  gap: 8px;
  padding-right: 4px;
  white-space: nowrap;
}

.action-cell--text {
  gap: 12px;
}

.action-link {
  color: #1677ff;
  font-size: 14px;
  line-height: 22px;
  cursor: pointer;
  white-space: nowrap;
  transition: color 0.2s ease;
}

.action-link:hover {
  color: #4096ff;
}

.action-link--success {
  color: #15803d;
}

.action-link--success:hover {
  color: #16a34a;
}

.action-link--danger {
  color: #d46b08;
}

.action-link--danger:hover {
  color: #fa8c16;
}

.action-more-link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.action-more-link__arrow {
  font-size: 10px;
}

:deep(.organization-table .ant-table) {
  background: transparent;
}

:deep(.organization-table .ant-table-thead > tr > th) {
  padding: 12px 16px;
  background: #fafafa !important;
  color: #262626;
  font-size: 14px;
  font-weight: 500;
  border-bottom: 1px solid #f0f0f0;
}

:deep(.organization-table .ant-table-thead > tr > th .ant-table-column-title) {
  color: #262626;
  font-weight: 500;
}

:deep(.organization-table .ant-table-thead > tr > th.ant-table-cell-fix-right) {
  text-align: left;
}

:deep(.organization-table .ant-table-cell-fix-right-last) {
  background: #fff;
}

:deep(.organization-table .ant-table-tbody > tr > td) {
  padding: 16px;
  border-bottom: 1px solid #f5f5f5;
  vertical-align: middle;
}

:deep(.organization-table .ant-table-tbody > tr:hover > td) {
  background: #fcfcfc;
}

:deep(.organization-table .institution-row--disabled > td) {
  background: #fcfcfc;
}

:deep(.organization-table .institution-row--disabled > td.ant-table-cell-fix-right-last) {
  background: #fcfcfc;
}

:deep(.organization-table .ant-pagination) {
  margin: 16px 8px 0;
}

:deep(.action-more-menu.ant-dropdown-menu) {
  min-width: 124px;
  padding: 8px 0;
  border-radius: 12px;
  box-shadow: 0 10px 28px rgba(15, 35, 95, 0.12);
}

:deep(.action-more-menu .ant-dropdown-menu-item) {
  min-height: 40px;
  padding: 8px 16px;
  border-radius: 0;
  font-size: 14px;
  color: #262626;
}

:deep(.filter-wrap .filter-section) {
  padding: 0 4px;
}

:deep(.filter-wrap .section-title) {
  color: #434343;
  font-weight: 600;
}

:deep(.filter-wrap .standard-filters) {
  display: flex;
  flex-wrap: wrap;
  /* gap: 10px;  */
}

:deep(.filter-wrap .selectBox .label) {
  color: #434343;
  font-weight: 600;
}

@media (max-width: 960px) {
  .organization-list {
    padding: 16px 16px 8px;
  }

  .table-title {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
