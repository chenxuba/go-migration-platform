<script setup lang="ts">
import type { TableColumnsType } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import { onMounted, reactive, ref } from 'vue'
import AllFilter from '@/components/common/all-filter.vue'
import {
  pageInstitutionsApi,
  updateInstitutionStatusApi,
  type InstitutionItem,
  type InstitutionSummary,
} from '@/api/platform/institutions'
import InstitutionFormDrawer from './components/institution-form-drawer.vue'
import messageService from '@/utils/messageService'

const displayArray = ['customSearch', 'enableStatus']
const directCountyCityLabels = new Set([
  '市辖区',
  '县',
  '省直辖县级行政区划',
  '自治区直辖县级行政区划',
])

const customSearchFilters = [
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
]

const listLoading = ref(false)
const statusSubmittingId = ref<number | null>(null)
const dataSource = ref<InstitutionItem[]>([])
const institutionDrawerOpen = ref(false)
const editingInstitutionId = ref<number | null>(null)
const summary = ref<InstitutionSummary>({
  totalCount: 0,
  enabledCount: 0,
  disabledCount: 0,
})

const filters = reactive<{
  keyword?: string
  mobile?: string
  enabled?: boolean
}>({
  keyword: undefined,
  mobile: undefined,
  enabled: undefined,
})

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
    title: '状态',
    dataIndex: 'enabled',
    key: 'enabled',
    width: 120,
    align: 'center' as const,
  },
  {
    title: '操作',
    key: 'action',
    width: 160,
    fixed: 'right' as const,
  },
]

let requestSerial = 0

function resetFilters() {
  filters.keyword = undefined
  filters.mobile = undefined
  filters.enabled = undefined
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

function getRowClassName(record: Partial<InstitutionItem>) {
  return record.enabled ? 'institution-row' : 'institution-row institution-row--disabled'
}

function openCreateDrawer() {
  editingInstitutionId.value = null
  institutionDrawerOpen.value = true
}

function openEditDrawer(record: Partial<InstitutionItem>) {
  editingInstitutionId.value = Number(record.id || 0) || null
  institutionDrawerOpen.value = true
}

function handleDrawerSaved() {
  institutionDrawerOpen.value = false
  editingInstitutionId.value = null
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
    messageService.success(enabled ? '机构已启用' : '机构已停用')
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
      enabled: filters.enabled,
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
      enabledCount: list.filter(item => item.enabled).length,
      disabledCount: list.filter(item => !item.enabled).length,
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
    }

    pagination.current = 1
    fetchInstitutions()
  },
  'update:enableStatusFilter': (value: number | undefined, isClearAll: boolean) => {
    if (isClearAll) {
      resetFilters()
    }
    else {
      filters.enabled = value === 1 ? true : value === 0 ? false : undefined
    }

    pagination.current = 1
    fetchInstitutions()
  },
}

onMounted(() => {
  fetchInstitutions()
})
</script>

<template>
  <div class="organization-page">
    <div class="filter-wrap">
      <AllFilter
        :display-array="displayArray"
        :is-quick-show="false"
        :custom-is-display-list="customSearchFilters"
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

        <a-button type="primary" @click="openCreateDrawer">
          <template #icon>
            <PlusOutlined />
          </template>
          新建机构
        </a-button>
      </div>

      <div class="table-content">
        <a-table
          class="organization-table"
          :columns="columns"
          :data-source="dataSource"
          :loading="listLoading"
          :pagination="pagination"
          :scroll="{ x: 920 }"
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

            <template v-else-if="column.key === 'account'">
              <div class="info-cell">
                <div class="cell-title cell-title--sm">
                  {{ record.mobile || '--' }}
                </div>
                <div class="cell-sub">
                  登录手机号
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

            <template v-else-if="column.key === 'enabled'">
              <span class="status-text" :class="record.enabled ? 'status-text--enabled' : 'status-text--disabled'">
                <span class="status-text__dot" />
                {{ record.enabled ? '启用中' : '已停用' }}
              </span>
            </template>

            <template v-else-if="column.key === 'action'">
              <div class="action-cell">
                <a-button type="link" class="action-cell__link" @click="openEditDrawer(record)">
                  编辑
                </a-button>
                <a-popconfirm
                  :title="record.enabled ? '确定停用该机构？' : '确定启用该机构？'"
                  ok-text="确定"
                  cancel-text="取消"
                  @confirm="toggleInstitutionStatus(record, !record.enabled)"
                >
                  <a-button
                    type="link"
                    class="action-cell__link"
                    :loading="statusSubmittingId === Number(record.id)"
                  >
                    {{ record.enabled ? '停用' : '启用' }}
                  </a-button>
                </a-popconfirm>
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

.action-cell {
  display: flex;
  align-items: center;
  gap: 18px;
}

.action-cell__link {
  padding-inline: 0;
  height: auto;
  color: #1677ff;
  font-size: 14px;
  font-weight: 500;
}

.action-cell__link:hover,
.action-cell__link:focus {
  color: #4096ff;
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

:deep(.organization-table .ant-pagination) {
  margin: 16px 8px 0;
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
  gap: 10px;
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
