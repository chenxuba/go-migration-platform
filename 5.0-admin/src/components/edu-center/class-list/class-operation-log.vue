<script setup lang="ts">
import dayjs from 'dayjs'
import type { TableColumnsType } from 'ant-design-vue'
import { computed, reactive, ref, watch } from 'vue'
import { pageGroupClassOperationLogsApi, type GroupClassOperationLogItem } from '@/api/edu-center/group-class'
import messageService from '@/utils/messageService'

const props = withDefaults(defineProps<{
  open?: boolean
  classId?: string
  className?: string
}>(), {
  open: false,
  classId: '',
  className: '',
})

const displayArray = ['stuPhoneSearch', 'classStatus', 'createUser', 'classStopTime']
const operationTypeOptions = [
  { id: '1', value: '切换默认学费账户' },
  { id: '2', value: '在班状态变更' },
  { id: '3', value: '移出班级学员' },
  { id: '4', value: '添加班级学员' },
]

const columns: TableColumnsType<GroupClassOperationLogItem> = [
  {
    title: '操作时间',
    dataIndex: 'operateTime',
    key: 'operateTime',
    width: 180,
  },
  {
    title: '学员',
    dataIndex: 'studentName',
    key: 'studentName',
    width: 150,
  },
  {
    title: '操作类型',
    dataIndex: 'operationTypeText',
    key: 'operationTypeText',
    width: 150,
  },
  {
    title: '操作内容',
    dataIndex: 'operationContent',
    key: 'operationContent',
    width: 520,
  },
  {
    title: '操作人',
    dataIndex: 'operatorName',
    key: 'operatorName',
    width: 140,
  },
]

const loading = ref(false)
const allFilterKey = ref(0)
const dataSource = ref<GroupClassOperationLogItem[]>([])
const filterStudentId = ref<string>()
const filterOperatorId = ref<string>()
const filterOperationTypes = ref<number[]>([])
const filterOperateRange = ref<[string, string] | null>(null)
const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
})

let loadSeq = 0

const totalWidth = computed(() =>
  columns.reduce((sum, item) => sum + Number(item.width || 0), 0),
)

const tablePagination = computed(() => ({
  current: pagination.current,
  pageSize: pagination.pageSize,
  total: pagination.total,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: (total: number) => `共 ${total} 条`,
}))

async function loadList() {
  const classId = String(props.classId || '').trim()
  if (!props.open || !classId) {
    dataSource.value = []
    pagination.total = 0
    return
  }
  const currentSeq = ++loadSeq
  loading.value = true
  try {
    const res = await pageGroupClassOperationLogsApi({
      queryModel: {
        classId,
        studentId: filterStudentId.value,
        operatorId: filterOperatorId.value,
        operateStartAt: filterOperateRange.value?.[0],
        operateEndAt: filterOperateRange.value?.[1],
        operationTypes: filterOperationTypes.value,
      },
      pageRequestModel: {
        needTotal: true,
        pageIndex: pagination.current,
        pageSize: pagination.pageSize,
        skipCount: (pagination.current - 1) * pagination.pageSize,
      },
    })
    if (currentSeq !== loadSeq)
      return
    if (res.code !== 200)
      throw new Error(res.message || '加载操作日志失败')
    dataSource.value = Array.isArray(res.result?.list) ? res.result.list : []
    pagination.total = Number(res.result?.total || 0)
  }
  catch (error: any) {
    if (currentSeq !== loadSeq)
      return
    dataSource.value = []
    pagination.total = 0
    messageService.error(error?.response?.data?.message || error?.message || '加载操作日志失败')
  }
  finally {
    if (currentSeq === loadSeq)
      loading.value = false
  }
}

function resetFilters() {
  filterStudentId.value = undefined
  filterOperatorId.value = undefined
  filterOperationTypes.value = []
  filterOperateRange.value = null
  pagination.current = 1
}

function reloadFirstPage() {
  pagination.current = 1
  loadList()
}

function normalizeStringValue(value: unknown) {
  const normalized = String(value || '').trim()
  return normalized || undefined
}

function normalizeDateRange(value: unknown) {
  if (!Array.isArray(value) || value.length < 2)
    return null
  const start = String(value[0] || '').trim()
  const end = String(value[1] || '').trim()
  if (!start || !end)
    return null
  return [start, end] as [string, string]
}

function normalizeNumberArray(value: unknown) {
  if (!Array.isArray(value))
    return []
  return value
    .map(item => Number(item))
    .filter(item => Number.isFinite(item) && item > 0)
}

function handleStudentFilter(value: unknown) {
  const next = normalizeStringValue(value)
  if (filterStudentId.value === next)
    return
  filterStudentId.value = next
  reloadFirstPage()
}

function handleOperatorFilter(value: unknown) {
  const next = normalizeStringValue(value)
  if (filterOperatorId.value === next)
    return
  filterOperatorId.value = next
  reloadFirstPage()
}

function handleOperationTypeFilter(value: unknown) {
  const next = normalizeNumberArray(value)
  if (JSON.stringify(filterOperationTypes.value) === JSON.stringify(next))
    return
  filterOperationTypes.value = next
  reloadFirstPage()
}

function handleOperateTimeFilter(value: unknown) {
  const next = normalizeDateRange(value)
  if (JSON.stringify(filterOperateRange.value) === JSON.stringify(next))
    return
  filterOperateRange.value = next
  reloadFirstPage()
}

function handleTableChange(page: { current?: number, pageSize?: number }) {
  const nextCurrent = Number(page.current || 1)
  const nextSize = Number(page.pageSize || pagination.pageSize)
  const changed = nextCurrent !== pagination.current || nextSize !== pagination.pageSize
  pagination.current = nextCurrent
  pagination.pageSize = nextSize
  if (changed)
    loadList()
}

function formatDateTime(value?: string) {
  if (!value)
    return '-'
  const date = dayjs(value)
  return date.isValid() ? date.format('YYYY-MM-DD HH:mm') : '-'
}

watch(
  () => `${props.open}|${String(props.classId || '').trim()}`,
  () => {
    resetFilters()
    allFilterKey.value += 1
    loadList()
  },
  { immediate: true },
)
</script>

<template>
  <div class="group-class-history-tab">
    <div class="filter-wrap">
      <all-filter
        :key="allFilterKey"
        :display-array="displayArray"
        :class-status-options="operationTypeOptions"
        class-status-label="操作类型"
        create-user-label="操作人"
        class-stop-time-label="操作时间"
        :is-quick-show="false"
        @update:stu-phone-search-filter="handleStudentFilter"
        @update:class-status-filter="handleOperationTypeFilter"
        @update:create-user-filter="handleOperatorFilter"
        @update:class-stop-time-filter="handleOperateTimeFilter"
      />
    </div>

    <div class="history-card">
      <div class="history-card__summary">
        <custom-title :title="`共 ${pagination.total} 条操作记录`" font-size="14px" class="pb-12px" />
      </div>

      <a-table
        row-key="id"
        size="middle"
        :loading="loading"
        :columns="columns"
        :data-source="dataSource"
        :pagination="tablePagination"
        :scroll="{ x: totalWidth }"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.dataIndex === 'operateTime'">
            {{ formatDateTime(record.operateTime) }}
          </template>
          <template v-else-if="column.dataIndex === 'studentName'">
            <div class="student-name">
              {{ record.studentName || '-' }}
            </div>
          </template>
          <template v-else-if="column.dataIndex === 'operationTypeText'">
            <a-tag class="type-tag" color="blue">
              {{ record.operationTypeText || '-' }}
            </a-tag>
          </template>
          <template v-else-if="column.dataIndex === 'operationContent'">
            <div class="content-cell" :title="record.operationContent || '-'">
              {{ record.operationContent || '-' }}
            </div>
          </template>
          <template v-else-if="column.dataIndex === 'operatorName'">
            {{ record.operatorName || '-' }}
          </template>
        </template>
      </a-table>
    </div>
  </div>
</template>

<style lang="less" scoped>
.group-class-history-tab {
  padding: 12px;
}

.filter-wrap {
  margin-bottom: 12px;
}

.history-card {
  border: 1px solid #eef1f6;
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 8px 24px rgba(31, 35, 41, 0.06);
  overflow: hidden;
}

.history-card__summary {
  padding: 18px 20px 0;
}

.student-name {
  color: #1f2329;
  font-weight: 500;
}

.type-tag {
  margin-inline-end: 0;
  border-radius: 999px;
}

.content-cell {
  color: #4e5969;
  line-height: 22px;
  word-break: break-all;
}

:deep(.ant-table-wrapper) {
  padding: 0 20px 20px;
}

:deep(.ant-table-thead > tr > th) {
  background: #f8fafc;
  color: #4e5969;
  font-weight: 600;
}
</style>
