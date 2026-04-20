<script setup lang="ts">
import type { TableColumnsType } from 'ant-design-vue'
import { computed, onMounted, reactive, ref } from 'vue'
import AllFilter from '@/components/common/all-filter.vue'
import {
  pageGovernmentAccountsApi,
  type GovernmentAccountItem,
} from '@/api/platform/government-accounts'
import messageService from '@/utils/messageService'

const displayArray = ['customSearch']

const listLoading = ref(false)
const dataSource = ref<GovernmentAccountItem[]>([])

const filters = reactive<{
  username?: string
  mobile?: string
}>({
  username: undefined,
  mobile: undefined,
})

const customSearchFilters = computed(() => [
  {
    id: 'username',
    fieldKey: '姓名/账号',
    fieldType: 1,
  },
  {
    id: 'mobile',
    fieldKey: '手机号',
    fieldType: 1,
  },
])

const customSearchValues = computed(() => ({
  username: filters.username ?? '',
  mobile: filters.mobile ?? '',
}))

const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
  showSizeChanger: true,
  showTotal: (total: number) => `共 ${total} 个政府账户`,
})

const columns: TableColumnsType<GovernmentAccountItem> = [
  {
    title: '姓名/账号',
    key: 'name',
    width: 220,
    fixed: 'left' as const,
  },
  {
    title: '手机号',
    dataIndex: 'mobile',
    key: 'mobile',
    width: 160,
  },
  {
    title: '账号状态',
    dataIndex: 'status',
    key: 'status',
    width: 120,
    align: 'center' as const,
  },
  {
    title: '监管层级',
    dataIndex: 'level',
    key: 'level',
    width: 140,
    align: 'center' as const,
  },
  {
    title: '管辖范围',
    dataIndex: 'scope',
    key: 'scope',
    width: 180,
  },
  {
    title: '已分配角色',
    dataIndex: 'roleName',
    key: 'roleName',
    width: 260,
  },
  {
    title: '最近登录时间',
    dataIndex: 'lastLoginTime',
    key: 'lastLoginTime',
    width: 180,
  },
  {
    title: '操作',
    key: 'action',
    width: 100,
    fixed: 'right' as const,
    align: 'center' as const,
  },
]

let requestSerial = 0

function resetFilters() {
  filters.username = undefined
  filters.mobile = undefined
}

function getDisplayName(record: GovernmentAccountItem) {
  return String(record.nickName || record.username || '').trim() || '--'
}

function getAccountName(record: GovernmentAccountItem) {
  return String(record.username || '').trim() || '--'
}

function normalizeText(value?: string) {
  const text = String(value || '').trim()
  return text || '--'
}

function toGovernmentAccount(record: unknown) {
  return record as GovernmentAccountItem
}

function getStatusColor(status?: string) {
  const value = String(status || '').trim()
  if (value === '正常')
    return 'success'
  if (value === '未登录')
    return 'warning'
  return 'default'
}

function getLevelColor(level?: string) {
  const value = String(level || '').trim()
  if (value === '超级监管')
    return 'gold'
  if (value === '省级')
    return 'blue'
  if (value === '市级')
    return 'cyan'
  if (value === '区县级')
    return 'geekblue'
  if (value === '多层级')
    return 'purple'
  return 'default'
}

async function fetchGovernmentAccounts() {
  const currentRequest = ++requestSerial
  listLoading.value = true
  try {
    const res = await pageGovernmentAccountsApi({
      current: pagination.current,
      size: pagination.pageSize,
      username: filters.username,
      mobile: filters.mobile,
    })

    if (currentRequest !== requestSerial)
      return

    const payload = res.data
    dataSource.value = payload?.items || []
    pagination.total = payload?.total || 0
    pagination.current = payload?.current || pagination.current
    pagination.pageSize = payload?.size || pagination.pageSize
  }
  catch (error: any) {
    if (currentRequest !== requestSerial)
      return
    console.error('fetch government accounts failed', error)
    messageService.error(error?.message || '获取政府账户失败')
  }
  finally {
    if (currentRequest === requestSerial)
      listLoading.value = false
  }
}

function handleTableChange(page: { current?: number, pageSize?: number }) {
  pagination.current = page.current || 1
  pagination.pageSize = page.pageSize || 20
  fetchGovernmentAccounts()
}

const filterUpdateHandlers = {
  'update:customSearchInputFilter': (payload: any, isClearAll: boolean, id?: string) => {
    if (isClearAll) {
      resetFilters()
    }
    else {
      const fieldId = id || payload?.item?.id
      const value = String(payload?.value ?? '').trim() || undefined

      if (fieldId === 'username')
        filters.username = value

      if (fieldId === 'mobile')
        filters.mobile = value
    }

    pagination.current = 1
    fetchGovernmentAccounts()
  },
}

onMounted(() => {
  fetchGovernmentAccounts()
})
</script>

<template>
  <div class="government-account-page">
    <div class="filter-wrap">
      <AllFilter
        :display-array="displayArray"
        :is-quick-show="false"
        :custom-is-display-list="customSearchFilters"
        :custom-search-values="customSearchValues"
        v-on="filterUpdateHandlers"
      />
    </div>

    <div class="account-list">
      <div class="table-title">
        <div class="table-title__left">
          <div class="total">
            共 {{ pagination.total }} 个政府账户
          </div>
          <div class="hint">
            管辖范围后续会跟随区域权限配置同步展示
          </div>
        </div>
      </div>

      <div class="table-content">
        <a-table
          class="account-table"
          :columns="columns"
          :data-source="dataSource"
          :loading="listLoading"
          :pagination="pagination"
          :scroll="{ x: 1360 }"
          row-key="id"
          size="small"
          @change="handleTableChange"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'name'">
              <div class="name-cell">
                <div class="cell-title">
                  {{ getDisplayName(toGovernmentAccount(record)) }}
                </div>
                <div class="cell-sub">
                  账号：{{ getAccountName(toGovernmentAccount(record)) }}
                </div>
              </div>
            </template>

            <template v-else-if="column.key === 'status'">
              <a-tag :color="getStatusColor(toGovernmentAccount(record).status)" class="status-tag">
                {{ normalizeText(toGovernmentAccount(record).status) }}
              </a-tag>
            </template>

            <template v-else-if="column.key === 'level'">
              <a-tag :color="getLevelColor(toGovernmentAccount(record).level)" class="level-tag">
                {{ normalizeText(toGovernmentAccount(record).level) }}
              </a-tag>
            </template>

            <template v-else-if="column.key === 'roleName'">
              <a-tooltip :title="normalizeText(toGovernmentAccount(record).roleName)" placement="topLeft">
                <div class="ellipsis-text">
                  {{ normalizeText(toGovernmentAccount(record).roleName) }}
                </div>
              </a-tooltip>
            </template>

            <template v-else-if="column.key === 'action'">
              <span class="action-placeholder">--</span>
            </template>

            <template v-else>
              {{ normalizeText((record as any)[column.dataIndex as string]) }}
            </template>
          </template>
        </a-table>
      </div>
    </div>
  </div>
</template>

<style scoped lang="less">
.government-account-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.account-list {
  padding: 16px 24px 8px;
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 8px 24px rgba(15, 23, 42, 0.04);
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
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.total {
  position: relative;
  padding-left: 10px;
  display: flex;
  align-items: center;
  color: #262626;
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

.hint {
  color: #8c8c8c;
  font-size: 12px;
  line-height: 20px;
}

.name-cell,
.info-cell {
  min-width: 0;
}

.cell-title {
  color: #262626;
  font-size: 14px;
  font-weight: 600;
  line-height: 22px;
}

.cell-sub {
  color: #8c8c8c;
  font-size: 12px;
  line-height: 20px;
}

.ellipsis-text {
  overflow: hidden;
  color: #262626;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.status-tag,
.level-tag {
  min-width: 72px;
  text-align: center;
}

.action-placeholder {
  color: #b0b7c3;
}

:deep(.account-table .ant-table-thead > tr > th) {
  color: #5b6472;
  font-weight: 600;
}

:deep(.account-table .ant-table-tbody > tr > td) {
  padding-top: 12px;
  padding-bottom: 12px;
}
</style>
