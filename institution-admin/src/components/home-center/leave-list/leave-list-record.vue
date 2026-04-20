<script setup>
import { computed, onMounted, ref } from 'vue'
import { debounce } from 'lodash-es'
import dayjs from 'dayjs'
import leaveDetailsDrawer from './components/leaveDetailsDrawer.vue'
import { useTableColumns } from '@/composables/useTableColumns'
import { getLeavePagedListApi } from '@/api/home-center/leave'
import messageService from '@/utils/messageService'

const displayArray = ref(['approvalStatus', 'leaveType', 'applyTime'])
const defaultApprovalStatusVals = ref([1])
const approvalStatusOptions = [
  { id: 1, value: '待处理' },
  { id: 2, value: '已通过' },
  { id: 3, value: '已拒绝' },
  { id: 4, value: '已撤销' },
]

const defaultStudentAvatar = 'https://cdn.schoolpal.cn/schoolpal/next-erp/avator_male.png?x-oss-process=image/resize,w_120'

const loading = ref(false)
const dataSource = ref([])
const allFilterRef = ref(null)
const openAddLeaveModal = ref(false)
const openLeaveDetailsDrawer = ref(false)
const currentLeaveId = ref('')

const pagination = ref({
  current: 1,
  pageSize: 20,
  total: 0,
  showSizeChanger: true,
  showTotal: total => `共 ${total} 条`,
  pageSizeOptions: ['10', '20', '50', '100'],
})

const queryState = ref({
  studentId: undefined,
  applyStartTime: undefined,
  applyEndTime: undefined,
  leaveTypes: undefined,
  statuses: [...defaultApprovalStatusVals.value],
})

const statusStyleMap = {
  1: {
    color: '#1677ff',
    background: '#e6f4ff',
  },
  2: {
    color: '#389e0d',
    background: '#f6ffed',
  },
  3: {
    color: '#cf1322',
    background: '#fff1f0',
  },
  4: {
    color: '#8c8c8c',
    background: '#f5f5f5',
  },
}

const allColumns = ref([
  {
    title: '学员/电话',
    dataIndex: 'studentName',
    key: 'studentName',
    width: 156,
    fixed: 'left',
  },
  {
    title: '开始时间',
    dataIndex: 'startTime',
    key: 'startTime',
    width: 170,
  },
  {
    title: '结束时间',
    dataIndex: 'endTime',
    key: 'endTime',
    width: 170,
  },
  {
    title: '请假类型',
    dataIndex: 'leaveTypeText',
    key: 'leaveTypeText',
    width: 120,
  },
  {
    title: '发起人',
    dataIndex: 'initiateStaffName',
    key: 'initiateStaffName',
    width: 140,
  },
  {
    title: '处理状态',
    dataIndex: 'status',
    key: 'status',
    width: 120,
  },
  {
    title: '审批人',
    dataIndex: 'currentApproverName',
    key: 'currentApproverName',
    width: 160,
  },
  {
    title: '申请时间',
    dataIndex: 'applyTime',
    key: 'applyTime',
    width: 170,
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    fixed: 'right',
    width: 110,
  },
])

const { filteredColumns, totalWidth } = useTableColumns({
  storageKey: 'leave-list-record',
  allColumns,
  excludeKeys: ['action'],
})

function formatDateTime(value) {
  if (!value || String(value).startsWith('0001-01-01'))
    return '-'
  const formatted = dayjs(value)
  return formatted.isValid() ? formatted.format('YYYY-MM-DD HH:mm') : String(value).replace('T', ' ').slice(0, 16)
}

function getStatusStyle(status) {
  return statusStyleMap[status] || statusStyleMap[4]
}

function formatInitiateStaffName(name, isAgent) {
  const normalized = String(name || '').trim()
  if (!normalized)
    return '-'
  if (!isAgent)
    return normalized
  if (normalized.includes('（代办）') || normalized.includes('(代办)'))
    return normalized
  return `${normalized}（代办）`
}

function formatApproverName(name, status) {
  const normalized = String(name || '').trim()
  if (normalized)
    return normalized
  if (Number(status) === 2)
    return '系统自动执行'
  return '-'
}

function getLeaveTypeClass(leaveType) {
  switch (Number(leaveType)) {
    case 1:
      return 'personal'
    case 2:
      return 'sick'
    case 3:
      return 'suspend'
    default:
      return 'default'
  }
}

async function fetchLeaveList(id, type) {
  try {
    loading.value = true
    const res = await getLeavePagedListApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: pagination.value.pageSize,
        pageIndex: pagination.value.current,
        skipCount: (pagination.value.current - 1) * pagination.value.pageSize,
      },
      queryModel: {
        ...queryState.value,
      },
      sortModel: {
        byApplyTime: -1,
      },
    })

    if (res.code === 200) {
      dataSource.value = (res.result?.list || []).map(item => ({
        ...item,
        key: item.id,
      }))
      pagination.value.total = res.result?.total || 0
      if (type) {
        allFilterRef.value?.clearQuickFilter(id, type)
      }
    }
  }
  catch (error) {
    console.error('获取请假列表失败:', error)
    messageService.error('获取请假列表失败')
  }
  finally {
    loading.value = false
  }
}

const handleFilterUpdate = debounce((updates, isClearAll = false, id, type) => {
  if (isClearAll) {
    queryState.value = {
      studentId: undefined,
      applyStartTime: undefined,
      applyEndTime: undefined,
      leaveTypes: undefined,
      statuses: undefined,
    }
  }
  else {
    Object.assign(queryState.value, updates)
  }

  pagination.value.current = 1
  fetchLeaveList(id, type)
}, 300, { leading: true, trailing: false })

const filterUpdateHandlers = computed(() => ({
  'update:approvalStatusFilter': (val, isClearAll, id, type) => handleFilterUpdate({
    statuses: Array.isArray(val) && val.length ? val : undefined,
  }, isClearAll, id, type),
  'update:leaveTypeFilter': (val, isClearAll, id, type) => handleFilterUpdate({
    leaveTypes: Array.isArray(val) && val.length ? val : undefined,
  }, isClearAll, id, type),
  'update:applyTimeFilter': (val, isClearAll, id, type) => {
    if (Array.isArray(val) && val.length === 2) {
      handleFilterUpdate({
        applyStartTime: val[0],
        applyEndTime: val[1],
      }, isClearAll, id, type)
      return
    }
    handleFilterUpdate({
      applyStartTime: undefined,
      applyEndTime: undefined,
    }, isClearAll, id, type)
  },
  'update:stuPhoneSearchFilter': (val, isClearAll, id, type) => handleFilterUpdate({
    studentId: val || undefined,
  }, isClearAll, id, type),
}))

function handleTableChange(pag) {
  pagination.value.current = pag.current
  pagination.value.pageSize = pag.pageSize
  fetchLeaveList()
}

function handleAddLeave() {
  openAddLeaveModal.value = true
}

function handleLeaveCreated() {
  openAddLeaveModal.value = false
  pagination.value.current = 1
  fetchLeaveList()
}

function handleLeaveDetails(record) {
  currentLeaveId.value = String(record.id || '')
  openLeaveDetailsDrawer.value = true
}

function handleLeaveDetailsClosed() {
  fetchLeaveList()
}

function handleLeaveChanged() {
  fetchLeaveList()
}

onMounted(() => {
  fetchLeaveList()
})
</script>

<template>
  <div>
    <div class="filter-wrap bg-white pl-3 pr-3 rounded-lb-4 rounded-rb-4">
      <all-filter
        ref="allFilterRef"
        :display-array="displayArray"
        :is-quick-show="false"
        :is-show-search-stu-phonefilter="true"
        :default-approval-status-vals="defaultApprovalStatusVals"
        :approval-status-options-override="approvalStatusOptions"
        apply-time-label="申请时间"
        v-on="filterUpdateHandlers"
      />
    </div>

    <div class="student-list mt-3 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            当前共计 {{ pagination.total }} 条请假申请
          </div>
          <div class="edit flex">
            <a-button type="primary" @click="handleAddLeave">
              请假代办
            </a-button>
          </div>
        </div>

        <div class="table-content mt-2">
          <a-table
            row-key="id"
            :data-source="dataSource"
            :pagination="pagination"
            :columns="filteredColumns"
            :loading="loading"
            :scroll="{ x: totalWidth }"
            size="small"
            @change="handleTableChange"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'studentName'">
                <student-avatar
                  :id="record.studentId"
                  :name="record.studentName || '-'"
                  :avatar-url="record.studentAvatarUrl || defaultStudentAvatar"
                  :phone="record.studentPhone"
                  :auto-width="false"
                  :show-gender="false"
                  :show-age="false"
                  default-active-key="0"
                />
              </template>

              <template v-else-if="column.key === 'startTime'">
                {{ formatDateTime(record.startTime) }}
              </template>

              <template v-else-if="column.key === 'endTime'">
                {{ formatDateTime(record.endTime) }}
              </template>

              <template v-else-if="column.key === 'leaveTypeText'">
                <span class="leave-type-pill" :class="getLeaveTypeClass(record.leaveType)">
                  {{ record.leaveTypeText || '-' }}
                </span>
              </template>

              <template v-else-if="column.key === 'initiateStaffName'">
                {{ formatInitiateStaffName(record.operatorName || record.initiateStaffName, record.isAgent) }}
              </template>

              <template v-else-if="column.key === 'status'">
                <span class="status-chip" :style="getStatusStyle(record.status)">
                  {{ record.statusText || '-' }}
                </span>
              </template>

              <template v-else-if="column.key === 'currentApproverName'">
                {{ formatApproverName(record.approverName || record.currentApproverName, record.status) }}
              </template>

              <template v-else-if="column.key === 'applyTime'">
                {{ formatDateTime(record.applyTime) }}
              </template>

              <template v-else-if="column.key === 'action'">
                <a class="font500" @click="handleLeaveDetails(record)">
                  请假详情
                </a>
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>

    <leaveDetailsDrawer
      v-model="openLeaveDetailsDrawer"
      :leave-id="currentLeaveId"
      @changed="handleLeaveChanged"
      @closed="handleLeaveDetailsClosed"
    />
    <add-leave-modal v-model:open="openAddLeaveModal" @success="handleLeaveCreated" />
  </div>
</template>

<style lang="less" scoped>
.total {
  position: relative;
  padding-left: 10px;
  color: #222;
  display: flex;
  align-items: center;

  &::before {
    display: inline-block;
    background: var(--pro-ant-color-primary);
    border-radius: 2px;
    content: "";
    height: 12px;
    left: 0;
    position: absolute;
    width: 4px;
  }
}

.leave-type-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 54px;
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
  line-height: 20px;
}

.leave-type-pill.personal {
  color: #1677ff;
  background: #e6f4ff;
}

.leave-type-pill.sick {
  color: #d46b08;
  background: #fff7e6;
}

.leave-type-pill.suspend {
  color: #531dab;
  background: #f9f0ff;
}

.leave-type-pill.default {
  color: #595959;
  background: #f5f5f5;
}

.status-chip {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 72px;
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 500;
  line-height: 20px;
}
</style>
