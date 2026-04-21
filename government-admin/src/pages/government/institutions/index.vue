<script setup lang="ts">
import type { TableColumnsType } from 'ant-design-vue'
import { computed, onMounted, reactive, ref } from 'vue'
import AllFilter from '@/components/common/all-filter.vue'
import {
  pageGovernmentInstitutionsApi,
  type GovernmentInstitutionItem,
  type GovernmentInstitutionPagePayload,
  type GovernmentInstitutionSummary,
} from '@/api/government/institutions'
import messageService from '@/utils/messageService'

const displayArray = ['customSearch', 'enableStatus']
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

const loading = ref(false)
const dataSource = ref<GovernmentInstitutionItem[]>([])
const pageInfo = ref<GovernmentInstitutionPagePayload | null>(null)
const summary = ref<GovernmentInstitutionSummary>({
  totalCount: 0,
  enabledCount: 0,
  warningCount: 0,
  disabledCount: 0,
  expiredCount: 0,
  readingStudentCount: 0,
  intentStudentCount: 0,
  orderCount: 0,
})

const filters = reactive<{
  keyword?: string
  status?: number
  openType?: number
}>({
  keyword: undefined,
  status: undefined,
  openType: undefined,
})

const customSearchFilters = computed(() => [
  {
    id: 'keyword',
    fieldKey: '关键字',
    fieldType: 1,
  },
  {
    id: 'openType',
    fieldKey: '开通版本',
    fieldType: 4,
    optionsList: institutionOpenTypeOptions,
  },
])

const customSearchValues = computed(() => ({
  keyword: filters.keyword ?? '',
  openType: filters.openType ?? '',
}))

const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
  showSizeChanger: true,
  showTotal: (total: number) => `共 ${total} 家机构`,
})

const columns: TableColumnsType<GovernmentInstitutionItem> = [
  {
    title: '机构信息',
    key: 'organName',
    width: 260,
    fixed: 'left' as const,
  },
  {
    title: '机构地址',
    key: 'region',
    width: 300,
  },
  {
    title: '负责人 / 联系方式',
    key: 'contact',
    width: 180,
  },
  {
    title: '开通版本',
    key: 'openType',
    width: 110,
    align: 'center' as const,
  },
  {
    title: '机构状态',
    key: 'status',
    width: 110,
    align: 'center' as const,
  },
  {
    title: '注册时间',
    key: 'registerTime',
    width: 170,
  },
  {
    title: '到期时间',
    key: 'expireEndTime',
    width: 170,
  },
  {
    title: '意向学员',
    dataIndex: 'intentStudentCount',
    key: 'intentStudentCount',
    width: 100,
    align: 'right' as const,
  },
  {
    title: '订单数',
    dataIndex: 'orderCount',
    key: 'orderCount',
    width: 90,
    align: 'right' as const,
  },
  {
    title: '在职员工',
    key: 'staffCount',
    width: 100,
    align: 'right' as const,
    fixed: 'right' as const,
  },
  {
    title: '在读学员',
    dataIndex: 'readingStudentCount',
    key: 'readingStudentCount',
    width: 100,
    align: 'right' as const,
    fixed: 'right' as const,
  },
]

const statCards = computed(() => [
  {
    key: 'total',
    label: '纳管机构',
    value: `${summary.value.totalCount}`,
    unit: '家',
    tone: 'blue',
  },
  {
    key: 'enabled',
    label: '启用机构',
    value: `${summary.value.enabledCount}`,
    unit: '家',
    tone: 'green',
  },
  {
    key: 'warning',
    label: '预警机构',
    value: `${summary.value.warningCount}`,
    unit: '家',
    tone: 'orange',
  },
  {
    key: 'disabled',
    label: '停用 / 过期',
    value: `${summary.value.disabledCount + summary.value.expiredCount}`,
    unit: '家',
    tone: 'slate',
  },
  {
    key: 'reading',
    label: '在读学员',
    value: `${summary.value.readingStudentCount}`,
    unit: '人',
    tone: 'cyan',
  },
  {
    key: 'intent',
    label: '意向学员',
    value: `${summary.value.intentStudentCount}`,
    unit: '人',
    tone: 'gold',
  },
  {
    key: 'orders',
    label: '订单总量',
    value: `${summary.value.orderCount}`,
    unit: '笔',
    tone: 'red',
  },
])

let requestSerial = 0

function resolveRequestErrorMessage(error: any, fallback: string) {
  return String(error?.response?.data?.message || error?.message || fallback).trim() || fallback
}

function buildInstitutionAddress(record: Partial<GovernmentInstitutionItem>) {
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

function getInstitutionOpenTypeLabel(record: Partial<GovernmentInstitutionItem>) {
  const currentValue = Number(record.openType || 0)
  return institutionOpenTypeOptions.find(item => item.id === currentValue)?.value || '基础版'
}

function getInstitutionStatusMeta(record: Partial<GovernmentInstitutionItem>) {
  const currentValue = Number(record.status || 0)
  if (currentValue === 1) {
    return {
      text: '启用',
      color: 'success',
      className: 'status-tag--enabled',
    }
  }
  if (currentValue === 3) {
    return {
      text: '预警',
      color: 'warning',
      className: 'status-tag--warning',
    }
  }
  if (currentValue === 4) {
    return {
      text: '过期',
      color: 'error',
      className: 'status-tag--expired',
    }
  }
  return {
    text: '停用',
    color: 'default',
    className: 'status-tag--disabled',
  }
}

function resetFilters() {
  filters.keyword = undefined
  filters.status = undefined
  filters.openType = undefined
  pagination.current = 1
  fetchInstitutions()
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
      return
    }

    const fieldId = id || payload?.item?.id
    const value = String(payload?.value ?? '').trim() || undefined

    if (fieldId === 'keyword')
      filters.keyword = value

    if (fieldId === 'openType')
      filters.openType = value ? Number(value) : undefined

    pagination.current = 1
    fetchInstitutions()
  },
  'update:enableStatusFilter': (value: number | undefined, isClearAll: boolean) => {
    if (isClearAll) {
      resetFilters()
      return
    }

    filters.status = value ? Number(value) : undefined
    pagination.current = 1
    fetchInstitutions()
  },
}

async function fetchInstitutions() {
  const currentRequest = ++requestSerial
  loading.value = true
  try {
    const response = await pageGovernmentInstitutionsApi({
      current: pagination.current,
      size: pagination.pageSize,
      keyword: String(filters.keyword || '').trim() || undefined,
      status: filters.status,
      openType: filters.openType,
    })

    if (currentRequest !== requestSerial)
      return

    if (response.code !== 200) {
      messageService.error(response.message || '获取监管机构列表失败')
      return
    }

    const payload = response.data || null
    const items = Array.isArray(response.result) ? response.result : []

    pageInfo.value = payload
    dataSource.value = items
    pagination.total = Number(response.total || payload?.total || 0)
    summary.value = payload?.summary || {
      totalCount: pagination.total,
      enabledCount: 0,
      warningCount: 0,
      disabledCount: 0,
      expiredCount: 0,
      readingStudentCount: 0,
      intentStudentCount: 0,
      orderCount: 0,
    }
  }
  catch (error: any) {
    if (currentRequest !== requestSerial)
      return
    console.error('fetch government institutions failed', error)
    messageService.error(resolveRequestErrorMessage(error, '获取监管机构列表失败'))
  }
  finally {
    if (currentRequest === requestSerial)
      loading.value = false
  }
}

onMounted(() => {
  fetchInstitutions()
})
</script>

<template>
  <div class="gov-page">
    <section class="page-header">
      <div class="page-header__main">
        <div class="page-header__eyebrow">
          G 端 / 机构监管
        </div>
        <h1 class="page-header__title">
          监管机构台账
        </h1>
      </div>

      <div class="page-header__badge">
        {{ pageInfo?.levelLabel || '监管视角' }}
      </div>
    </section>

    <section class="stat-grid">
      <article v-for="card in statCards" :key="card.key" class="stat-card" :class="`stat-card--${card.tone}`">
        <div class="stat-card__label">
          {{ card.label }}
        </div>
        <div class="stat-card__value">
          {{ card.value }}
          <span class="stat-card__unit">{{ card.unit }}</span>
        </div>
      </article>
    </section>

    <div class="filter-wrap">
      <AllFilter
        :display-array="displayArray"
        :is-quick-show="false"
        :custom-is-display-list="customSearchFilters"
        :custom-search-values="customSearchValues"
        enable-status-label="机构状态"
        :enable-status-options-override="institutionStatusOptions"
        v-on="filterUpdateHandlers"
      />
    </div>

    <a-card :bordered="false">
      <a-table
        :columns="columns"
        :data-source="dataSource"
        :loading="loading"
        :pagination="pagination"
        :scroll="{ x: 1600 }"
        row-key="id"
        size="small"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'organName'">
            <div class="cell-main">
              {{ record.organName || '--' }}
            </div>
            <div class="cell-sub">
              省市区：{{ [record.province, record.city, record.region].filter(Boolean).join(' / ') || '--' }}
            </div>
          </template>

          <template v-else-if="column.key === 'region'">
            <div class="address-cell">
              <a-tooltip v-if="buildInstitutionAddress(record)" :title="buildInstitutionAddress(record)" placement="topLeft">
                <div class="cell-main cell-main--clamp">
                  {{ buildInstitutionAddress(record) }}
                </div>
              </a-tooltip>
              <div v-else class="cell-main">
                --
              </div>
            </div>
          </template>

          <template v-else-if="column.key === 'contact'">
            <div class="cell-main">
              {{ record.principal || '--' }}
            </div>
            <div class="cell-sub">
              {{ record.mobile || '--' }}
            </div>
          </template>

          <template v-else-if="column.key === 'openType'">
            <span class="text-strong">{{ getInstitutionOpenTypeLabel(record) }}</span>
          </template>

          <template v-else-if="column.key === 'status'">
            <a-tag :color="getInstitutionStatusMeta(record).color" :class="getInstitutionStatusMeta(record).className">
              {{ getInstitutionStatusMeta(record).text }}
            </a-tag>
          </template>

          <template v-else-if="column.key === 'registerTime'">
            {{ formatDateMinute(record.registerTime) }}
          </template>

          <template v-else-if="column.key === 'expireEndTime'">
            {{ formatDateMinute(record.expireEndTime) }}
          </template>

          <template v-else-if="column.key === 'staffCount'">
            <div class="number-cell">
              {{ record.activeStaffCount || 0 }}
            </div>
          </template>
        </template>
      </a-table>
    </a-card>
  </div>
</template>

<style scoped lang="less">
.gov-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
  width: 100%;
  min-width: 0;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 18px 22px;
  border-radius: 14px;
  background: linear-gradient(135deg, #f3fbf9 0%, #ffffff 100%);
}

.page-header__main {
  min-width: 0;
}

.page-header__eyebrow {
  margin-bottom: 10px;
  color: #0f766e;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.08em;
}

.page-header__title {
  margin: 0;
  color: #102a43;
  font-size: 20px;
  font-weight: 700;
}

.page-header__badge {
  flex-shrink: 0;
  padding: 8px 14px;
  border-radius: 999px;
  background: rgba(15, 118, 110, 0.08);
  color: #0f766e;
  font-size: 14px;
  font-weight: 700;
  line-height: 20px;
  white-space: nowrap;
}

.stat-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 12px;
}

.stat-card {
  padding: 18px;
  border-radius: 14px;
  background: #ffffff;
}

.stat-card__label {
  margin-bottom: 10px;
  color: #64748b;
  font-size: 13px;
}

.stat-card__value {
  color: #0f172a;
  font-size: 28px;
  font-weight: 700;
  line-height: 1;
}

.stat-card__unit {
  margin-left: 6px;
  color: #64748b;
  font-size: 13px;
  font-weight: 500;
}

.stat-card--blue {
  background: linear-gradient(135deg, #eff6ff 0%, #ffffff 100%);
}

.stat-card--green {
  background: linear-gradient(135deg, #ecfdf5 0%, #ffffff 100%);
}

.stat-card--orange {
  background: linear-gradient(135deg, #fff7ed 0%, #ffffff 100%);
}

.stat-card--slate {
  background: linear-gradient(135deg, #f8fafc 0%, #ffffff 100%);
}

.stat-card--cyan {
  background: linear-gradient(135deg, #ecfeff 0%, #ffffff 100%);
}

.stat-card--gold {
  background: linear-gradient(135deg, #fffbeb 0%, #ffffff 100%);
}

.stat-card--red {
  background: linear-gradient(135deg, #fef2f2 0%, #ffffff 100%);
}

.filter-wrap {
  padding: 0 12px;
  overflow: hidden;
  border: 1px solid #e9edf3;
  border-radius: 16px;
  background: #fff;
}

.cell-main {
  color: #0f172a;
  font-weight: 600;
  line-height: 22px;
}

.cell-main--clamp {
  display: -webkit-box;
  overflow: hidden;
  line-height: 22px;
  white-space: normal;
  word-break: break-all;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.cell-sub {
  margin-top: 4px;
  color: #64748b;
  font-size: 12px;
  line-height: 18px;
}

.address-cell {
  width: 100%;
  box-sizing: border-box;
  padding-right: 72px;
}

.cell-sub--right {
  text-align: right;
}

.number-cell {
  color: #0f172a;
  font-weight: 700;
  text-align: right;
}

.text-strong {
  color: #0f172a;
  font-weight: 600;
}

.status-tag--enabled {
  border-color: #86efac;
}

.status-tag--warning {
  border-color: #fdba74;
}

.status-tag--expired {
  border-color: #fca5a5;
}

.status-tag--disabled {
  border-color: #cbd5e1;
}

@media (max-width: 960px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
