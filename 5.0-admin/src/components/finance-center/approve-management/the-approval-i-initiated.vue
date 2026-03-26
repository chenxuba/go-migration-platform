<script setup>
import { QuestionCircleOutlined } from '@ant-design/icons-vue'
import { debounce } from 'lodash-es'
import { computed, onMounted, ref } from 'vue'
import dayjs from 'dayjs'
import ConfigApproveRuleDrawer from './configApproveRuleDrawer.vue'
import { useTableColumns } from '@/composables/useTableColumns'
import { getApprovalMyInitiatedPagedListApi } from '@/api/finance-center/approval-manage'
import messageService from '@/utils/messageService'

const displayArray = ref(['approveNumber', 'orderNumber', 'salesPerson', 'finishTime', 'createTime', 'approvalStatus'])
const currentMonthStart = dayjs().startOf('month').format('YYYY-MM-DD')
const today = dayjs().format('YYYY-MM-DD')
const defaultCreateTimeVals = ref([currentMonthStart, today])
const dataSource = ref([])
const loading = ref(false)
const allFilterRef = ref(null)
const pagination = ref({
  current: 1,
  pageSize: 50,
  total: 0,
  showSizeChanger: true,
  showTotal: total => `共 ${total} 条`,
  pageSizeOptions: ['10', '20', '50', '100'],
})

const queryState = ref({
  approveNumber: undefined,
  orderNumber: undefined,
  currentApproveStaffId: undefined,
  finishStartTime: undefined,
  finishEndTime: undefined,
  initiateStartTime: currentMonthStart,
  initiateEndTime: today,
  studentId: undefined,
  statuses: undefined,
})

const approvalStatusMap = {
  1: '审批中',
  2: '审批通过',
  3: '审批拒绝',
  4: '已作废',
}

const approvalTypeMap = {
  1: '报名订单',
  2: '转课订单',
  3: '退课订单',
  4: '储值充值',
  5: '储值退费',
  6: '退学杂教材费',
}

const allColumns = ref([
  {
    title: '审批编号',
    dataIndex: 'approveNumber',
    key: 'approveNumber',
    width: 210,
    fixed: 'left',
    required: true,
  },
  {
    title: '申请人',
    dataIndex: 'initiateStaffName',
    key: 'initiateStaffName',
    width: 100,
  },
  {
    title: '订单编号',
    dataIndex: 'orderNumber',
    key: 'orderNumber',
    width: 210,
  },
  {
    title: '学员/电话',
    dataIndex: 'studentName',
    key: 'studentName',
    width: 180,
  },
  {
    title: '当前审批人',
    dataIndex: 'approveFlows',
    key: 'approveFlows',
    width: 180,
  },
  {
    title: '审批状态',
    dataIndex: 'status',
    key: 'status',
    width: 120,
  },
  {
    title: '审批类型',
    dataIndex: 'type',
    key: 'type',
    width: 120,
  },
  {
    title: '审批完成时间',
    dataIndex: 'finishTime',
    key: 'finishTime',
    width: 150,
  },
  {
    title: '申请时间',
    dataIndex: 'initiateTime',
    key: 'initiateTime',
    width: 150,
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    fixed: 'right',
    width: 80,
  },
])

const { filteredColumns, totalWidth } = useTableColumns({
  storageKey: 'approval-initiated',
  allColumns,
  excludeKeys: ['action'],
})

const defaultStudentAvatar = 'https://cdn.schoolpal.cn/schoolpal/next-erp/avator_male.png?x-oss-process=image/resize,w_120'

const openOrderDetailDrawer = ref(false)
const currentOrderId = ref('')
const openApproveFlowsModal = ref(false)
const currentApproveFlows = ref([])
const openConfigApproveRuleDrawer = ref(false)

function handleOrderDetail(orderId) {
  currentOrderId.value = String(orderId || '')
  openOrderDetailDrawer.value = true
}

function handleOrderDetailDrawerClosed() {
  fetchApprovalList()
}

function handleViewApproveFlows(flows) {
  currentApproveFlows.value = Array.isArray(flows) ? flows : []
  openApproveFlowsModal.value = true
}

function handleConfigApproveRule() {
  openConfigApproveRuleDrawer.value = true
}

function normalizeApprovalItem(item) {
  return {
    ...item,
    approveFlows: Array.isArray(item.approveFlows) ? item.approveFlows : [],
  }
}

async function fetchApprovalList(id, type) {
  try {
    loading.value = true
    const res = await getApprovalMyInitiatedPagedListApi({
      queryModel: queryState.value,
      pageRequestModel: {
        needTotal: true,
        pageSize: pagination.value.pageSize,
        pageIndex: pagination.value.current,
        skipCount: (pagination.value.current - 1) * pagination.value.pageSize,
      },
      sortModel: {
        orderByInitiateTime: 0,
        orderByFinishTime: 0,
      },
    })
    if (res.code === 200) {
      dataSource.value = (res.result?.list || []).map(normalizeApprovalItem)
      pagination.value.total = res.result?.total || 0
      if (type) {
        allFilterRef.value?.clearQuickFilter(id, type)
      }
      return
    }
    messageService.error(res.message || '获取我发起的审批失败')
  }
  catch (error) {
    console.error('获取我发起的审批失败:', error)
    messageService.error('获取我发起的审批失败')
  }
  finally {
    loading.value = false
  }
}

const handleFilterUpdate = debounce((updates, isClearAll = false, id, type) => {
  if (isClearAll) {
    queryState.value = {
      approveNumber: undefined,
      orderNumber: undefined,
      currentApproveStaffId: undefined,
      finishStartTime: undefined,
      finishEndTime: undefined,
      initiateStartTime: currentMonthStart,
      initiateEndTime: today,
      studentId: undefined,
      statuses: undefined,
    }
  }
  else {
    Object.assign(queryState.value, updates)
  }
  pagination.value.current = 1
  fetchApprovalList(id, type)
}, 300, { leading: true, trailing: false })

const filterUpdateHandlers = computed(() => ({
  'update:approveNumberFilter': (val, isClearAll, id, type) => handleFilterUpdate({ approveNumber: val || undefined }, isClearAll, id, type),
  'update:orderNumberFilter': (val, isClearAll, id, type) => handleFilterUpdate({ orderNumber: val || undefined }, isClearAll, id, type),
  'update:salesPersonFilter': (val, isClearAll, id, type) => handleFilterUpdate({ currentApproveStaffId: val || undefined }, isClearAll, id, type),
  'update:finishTimeFilter': (val, isClearAll, id, type) => {
    if (Array.isArray(val) && val.length === 2) {
      handleFilterUpdate({ finishStartTime: val[0], finishEndTime: val[1] }, isClearAll, id, type)
      return
    }
    handleFilterUpdate({ finishStartTime: undefined, finishEndTime: undefined }, isClearAll, id, type)
  },
  'update:createTimeFilter': (val, isClearAll, id, type) => {
    if (Array.isArray(val) && val.length === 2) {
      handleFilterUpdate({ initiateStartTime: val[0], initiateEndTime: val[1] }, isClearAll, id, type)
      return
    }
    handleFilterUpdate({ initiateStartTime: currentMonthStart, initiateEndTime: today }, isClearAll, id, type)
  },
  'update:approvalStatusFilter': (val, isClearAll, id, type) => handleFilterUpdate({ statuses: Array.isArray(val) && val.length ? val : undefined }, isClearAll, id, type),
  'update:stuPhoneSearchFilter': (val, isClearAll, id, type) => handleFilterUpdate({ studentId: val || undefined }, isClearAll, id, type),
}))

function handleTableChange(pag) {
  pagination.value.current = pag.current
  pagination.value.pageSize = pag.pageSize
  fetchApprovalList()
}

function formatDateTime(value) {
  if (!value || String(value).startsWith('0001-01-01'))
    return '-'
  return String(value).replace('T', ' ').slice(0, 16)
}

function getCurrentApproverText(flows) {
  const currentFlow = flows.find(item => item.isCurrentStage)
  if (!currentFlow)
    return '-'
  return (currentFlow.flowStaffs || []).map(item => item.staffName).filter(Boolean).join('、') || '-'
}

function getFlowStaffNames(flow) {
  const names = Array.isArray(flow?.flowStaffs)
    ? flow.flowStaffs.map(item => item?.staffName).filter(Boolean)
    : []
  return names.length ? names.join('、') : '-'
}

onMounted(() => {
  fetchApprovalList()
})
</script>

<template>
  <div>
    <div class="filter-wrap bg-white pl-3 pr-3 rounded-lb-4 rounded-rb-4">
      <all-filter
        ref="allFilterRef"
        type="noDelCreateTime"
        :display-array="displayArray"
        :default-create-time-vals="defaultCreateTimeVals"
        :is-show-search-stu-phonefilter="true"
        create-time-label="申请时间"
        sales-person-label="当前审批人"
        sales-person-placeholder="搜索审批人"
        v-on="filterUpdateHandlers"
      />
    </div>
    <div class="student-list mt-3 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            当前共计 {{ pagination.total }} 条审批
          </div>
          <div class="edit flex">
            <a-space>
              <a-button @click="handleConfigApproveRule()">
                配置审批规则
              </a-button>
              <a-button>
                导出数据
              </a-button>
            </a-space>
          </div>
        </div>
        <div class="table-content mt-2">
          <a-table
            :data-source="dataSource"
            :pagination="pagination"
            :columns="filteredColumns"
            :scroll="{ x: totalWidth }"
            :loading="loading"
            size="small"
            @change="handleTableChange"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'approveNumber'">
                <span class="text-#06f flex-center justify-start cursor-pointer" @click="handleOrderDetail(record.orderId)">
                  {{ record.approveNumber }}
                </span>
              </template>
              <template v-else-if="column.key === 'studentName'">
                <student-avatar
                  :id="record.studentId"
                  :name="record.studentName || '-'"
                  :avatar-url="record.studentAvatar || defaultStudentAvatar"
                  :phone="record.studentPhone"
                  :show-gender="false"
                  :show-age="false"
                  default-active-key="0"
                />
              </template>
              <template v-else-if="column.key === 'orderNumber'">
                <clamped-text class="text-#06f cursor-pointer" :lines="1" :text="record.orderNumber || '-'" @click="handleOrderDetail(record.orderId)" />
              </template>
              <template v-else-if="column.key === 'initiateStaffName'">
                {{ record.initiateStaffName || '-' }}
              </template>
              <template v-else-if="column.key === 'approveFlows'">
                <div class="flex flex-items-center text-#222">
                  <span class="mr-1">{{ getCurrentApproverText(record.approveFlows) }}</span>
                  <a-tooltip placement="top">
                    <template #title>查看各级审批人</template>
                    <span class="inline-flex cursor-pointer text-#06f hover:opacity-80" @click="handleViewApproveFlows(record.approveFlows)">
                      <QuestionCircleOutlined />
                    </span>
                  </a-tooltip>
                </div>
              </template>
              <template v-else-if="column.key === 'status'">
                <div class="flex flex-items-center">
                  <span class="dot" :class="{ rejected: record.status === 3, pending: record.status === 1 }" />
                  <span>{{ approvalStatusMap[record.status] || '-' }}</span>
                </div>
              </template>
              <template v-else-if="column.key === 'type'">
                {{ approvalTypeMap[record.type] || '-' }}
              </template>
              <template v-else-if="column.key === 'finishTime'">
                {{ formatDateTime(record.finishTime) }}
              </template>
              <template v-else-if="column.key === 'initiateTime'">
                {{ formatDateTime(record.initiateTime) }}
              </template>
              <template v-else-if="column.key === 'action'">
                <a class="font500" @click="handleOrderDetail(record.orderId)">查看/处理</a>
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>
    <order-detail-drawer
      v-model:open="openOrderDetailDrawer"
      :order-id="currentOrderId"
      @closed="handleOrderDetailDrawerClosed"
    />
    <ConfigApproveRuleDrawer v-model:open="openConfigApproveRuleDrawer" />
    <a-modal
      v-model:open="openApproveFlowsModal"
      title="各级审批人"
      :footer="null"
      width="460px"
      centered
      :body-style="{ padding: '0 24px 24px' }"
    >
      <div class="approve-flow-list pt-3">
        <div v-for="flow in currentApproveFlows" :key="flow.step" class="approve-flow-item">
          <span class="text-#666">{{ getFlowStaffNames(flow) }}</span>
          <span v-if="flow.isCurrentStage" class="current-stage-tag">当前阶段</span>
        </div>
      </div>
    </a-modal>
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

span.dot {
  border-radius: 50%;
  display: inline-block;
  height: 6px;
  position: relative;
  vertical-align: middle;
  width: 6px;
  margin-right: 4px;
  background: #0c3;

  &.rejected {
    background: #f33;
  }

  &.pending {
    background: #f90;
  }
}

.current-stage-tag {
  margin-left: 10px;
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 12px;
  line-height: 18px;
  color: #1677ff;
  background: #eef4ff;
}

.approve-flow-list {
  padding-left: 2px;
}

.approve-flow-item {
  position: relative;
  min-height: 28px;
  padding-left: 22px;
  padding-bottom: 14px;
  display: flex;
  align-items: center;
  line-height: 24px;
}

.approve-flow-item:last-child {
  padding-bottom: 0;
}

.approve-flow-item::before {
  content: '';
  position: absolute;
  left: 3px;
  top: 7px;
  width: 9px;
  height: 9px;
  border-radius: 50%;
  background: #1677ff;
}

.approve-flow-item::after {
  content: '';
  position: absolute;
  left: 7px;
  top: 16px;
  bottom: -8px;
  width: 1px;
  background: #d0e4ff;
}

.approve-flow-item:last-child::after {
  display: none;
}
</style>
