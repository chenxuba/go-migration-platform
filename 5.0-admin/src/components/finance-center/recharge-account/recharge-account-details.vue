<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { CloseOutlined, DeleteOutlined, InfoCircleOutlined } from '@ant-design/icons-vue'
import { debounce } from 'lodash-es'
import dayjs from 'dayjs'
import { useTableColumns } from '@/composables/useTableColumns'
import { getRecommenderPageApi } from '@/api/enroll-center/intention-student'
import { getRechargeAccountDetailPageApi, getRechargeAccountExpendIncomeApi, getRechargeAccountItemPageApi } from '@/api/finance-center/recharge-account'
import { useStudentStore } from '@/stores/student'
import messageService from '@/utils/messageService'

const FLOW_TYPE_OPTIONS = [
  { id: 1, value: '储值账户充值' },
  { id: 2, value: '储值账户退费' },
  { id: 3, value: '报名订单支出' },
  { id: 4, value: '退费订单退回' },
  { id: 5, value: '作废储值充值' },
  { id: 6, value: '转课少补支出' },
  { id: 7, value: '转课退回' },
  { id: 8, value: '作废转课少补支出' },
  { id: 9, value: '作废转课退回' },
  { id: 10, value: '场地预约支出' },
  { id: 11, value: '场地预约退回' },
  { id: 12, value: '作废储值退费' },
]

const FLOW_TYPE_LABEL_MAP = FLOW_TYPE_OPTIONS.reduce((acc, item) => {
  acc[item.id] = item.value
  return acc
}, {})

const currentYearStart = dayjs().startOf('year').format('YYYY-MM-DD')
const today = dayjs().format('YYYY-MM-DD')

const filterState = ref({
  operationDate: [currentYearStart, today],
  flowTypes: [],
  studentId: '',
})

const searchKeyStuPhoneModel = ref(undefined)
const selectedStudentOption = ref(null)
const stuPhoneSearchOptions = ref([])
const stuPhoneSearchFinished = ref(false)
const stuPhoneSearchLoading = ref(false)
const stuPhoneSearchPagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
})

const dataSource = ref([])
const loading = ref(false)
const summary = ref({
  income: 0,
  expend: 0,
})
const pagination = ref({
  current: 1,
  pageSize: 20,
  total: 0,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: total => `共 ${total} 条`,
})

const studentStore = useStudentStore()
const openDrawer = ref(false)
const openOrderDetailDrawer = ref(false)
const openAccountDetailDrawer = ref(false)
const currentOrderId = ref('')
const currentAccountRecord = ref({})

const allColumns = ref([
  { title: '储值账户', dataIndex: 'rechargeAccountName', key: 'rechargeAccountName', width: 180 },
  { title: '明细关联学员', dataIndex: 'studentName', key: 'studentName', width: 220 },
  { title: '操作时间', dataIndex: 'createTime', key: 'createTime', width: 180 },
  { title: '明细类型', dataIndex: 'rechargeAccountFlowSourceType', key: 'rechargeAccountFlowSourceType', width: 160 },
  { title: '充值金额（元）', dataIndex: 'amount', key: 'amount', width: 140 },
  { title: '赠送金额（元）', dataIndex: 'givingAmount', key: 'givingAmount', width: 140 },
  { title: '残联金额（元）', dataIndex: 'residualAmount', key: 'residualAmount', width: 140 },
  { title: '订单编号', dataIndex: 'sourceOrderNumber', key: 'sourceOrderNumber', width: 220 },
  { title: '账单备注', dataIndex: 'remark', key: 'remark', width: 150 },
  { title: '总计（元）', dataIndex: 'totalAmount', key: 'totalAmount', width: 140 },
])

const { selectedValues, columnOptions, filteredColumns, totalWidth } = useTableColumns({
  storageKey: 'recharge-account-details',
  allColumns,
  excludeKeys: [],
})

const selectedFilterList = computed(() => {
  const list = [
    {
      key: 'operationDate',
      label: '操作日期',
      value: `${filterState.value.operationDate[0]}~${filterState.value.operationDate[1]}`,
      fixed: true,
    },
  ]
  if (filterState.value.flowTypes.length) {
    list.push({
      key: 'flowTypes',
      label: '明细类型',
      value: filterState.value.flowTypes.map(id => FLOW_TYPE_LABEL_MAP[id]).filter(Boolean).join('、'),
    })
  }
  if (filterState.value.studentId) {
    list.push({
      key: 'studentId',
      label: '学员/电话',
      value: selectedStudentOption.value?.stuName || filterState.value.studentId,
    })
  }
  return list
})

watch(filterState, () => {
  pagination.value.current = 1
  fetchDetailData()
}, { deep: true })

onUnmounted(() => {
  debouncedSearchStuPhone.cancel()
})

function handleSeeStuData(studentId = '') {
  if (!studentId) {
    messageService.error('invalid studentId')
    return
  }
  studentStore.setStudentId(String(studentId))
  openDrawer.value = true
}

function handleOrderDetail(orderId = '', orderNumber = '') {
  if (!orderId) {
    messageService.error('invalid orderId')
    return
  }
  currentOrderId.value = String(orderId)
  openOrderDetailDrawer.value = true
}

function handleOpenLinkedOrderFromAccountDetailDrawer(orderId) {
  const id = String(orderId || '').trim()
  if (!id) {
    messageService.error('invalid orderId')
    return
  }
  openAccountDetailDrawer.value = false
  currentOrderId.value = id
  openOrderDetailDrawer.value = true
}

async function handleOpenRechargeAccountDetail(record) {
  const fallbackRecord = {
    rechargeAccountId: record?.rechargeAccountId,
    rechargeAccountName: record?.rechargeAccountName,
    phone: record?.phone || '',
    rechargeAccountStudents: Array.isArray(record?.rechargeAccountStudents) ? record.rechargeAccountStudents : [],
  }

  try {
    if (record?.studentId) {
      const { result } = await getRechargeAccountItemPageApi({
        queryModel: {
          studentId: String(record.studentId),
          showZeroBalanceAccount: true,
        },
        pageRequestModel: {
          needTotal: true,
          pageSize: 50,
          pageIndex: 1,
          skipCount: 0,
        },
        sortModel: {
          orderByUpdatedTime: 0,
        },
      })
      const accountList = Array.isArray(result?.list) ? result.list : []
      currentAccountRecord.value = accountList.find(item => String(item.rechargeAccountId) === String(record.rechargeAccountId)) || fallbackRecord
    }
    else {
      currentAccountRecord.value = fallbackRecord
    }
  }
  catch (error) {
    console.error('获取储值账户详情信息失败:', error)
    currentAccountRecord.value = fallbackRecord
  }

  openAccountDetailDrawer.value = true
}

function clearFilter(key) {
  if (key === 'flowTypes') {
    filterState.value.flowTypes = []
    return
  }
  if (key === 'studentId') {
    filterState.value.studentId = ''
    searchKeyStuPhoneModel.value = undefined
    selectedStudentOption.value = null
  }
}

function clearSelectableFilters() {
  filterState.value.flowTypes = []
  filterState.value.studentId = ''
  searchKeyStuPhoneModel.value = undefined
  selectedStudentOption.value = null
}

async function getStuPhoneSearchPage(params = { key: undefined, studentStatus: undefined }) {
  try {
    if (stuPhoneSearchFinished.value && stuPhoneSearchPagination.value.current !== 1) return
    stuPhoneSearchLoading.value = true
    const res = await getRecommenderPageApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: stuPhoneSearchPagination.value.pageSize,
        pageIndex: stuPhoneSearchPagination.value.current,
        skipCount: 0,
      },
      queryModel: {
        searchKey: params.key,
        studentStatus: params.studentStatus,
      },
      sortModel: {},
    })
    if (res.code === 200) {
      const resultData = Array.isArray(res.result) ? res.result : []
      if (stuPhoneSearchPagination.value.current === 1) stuPhoneSearchOptions.value = resultData
      else stuPhoneSearchOptions.value = [...stuPhoneSearchOptions.value, ...resultData]
      stuPhoneSearchPagination.value.total = res.total || 0
      stuPhoneSearchFinished.value = stuPhoneSearchOptions.value.length >= stuPhoneSearchPagination.value.total
    }
  }
  catch (error) {
    console.error('加载学员搜索失败:', error)
  }
  finally {
    stuPhoneSearchLoading.value = false
  }
}

function dropdownVisibleChangeFun(event) {
  if (!event) return
  stuPhoneSearchPagination.value.current = 1
  stuPhoneSearchOptions.value = []
  stuPhoneSearchFinished.value = false
  getStuPhoneSearchPage({})
}

const debouncedSearchStuPhone = debounce((value) => {
  stuPhoneSearchPagination.value.current = 1
  stuPhoneSearchFinished.value = false
  stuPhoneSearchOptions.value = []
  getStuPhoneSearchPage({ key: value })
}, 300)

function handleSearchStuPhone(value) {
  debouncedSearchStuPhone(value)
}

function handleStuPhonePopupScroll(event) {
  const { target } = event
  const { scrollTop, scrollHeight, clientHeight } = target
  if (scrollHeight - scrollTop - clientHeight < 1) {
    if (!stuPhoneSearchLoading.value && stuPhoneSearchPagination.value.current * stuPhoneSearchPagination.value.pageSize < stuPhoneSearchPagination.value.total) {
      stuPhoneSearchPagination.value.current += 1
      getStuPhoneSearchPage({})
    }
  }
}

function handleStuPhoneSearchChange(value) {
  const id = typeof value === 'object' && value !== null ? value.value : value
  if (id) selectedStudentOption.value = stuPhoneSearchOptions.value.find(user => String(user.id) === String(id)) || null
  else selectedStudentOption.value = null
  nextTick(() => {
    filterState.value.studentId = id || ''
  })
}

function formatDate(dateStr) {
  if (!dateStr) return '-'
  return String(dateStr).replace('T', ' ').slice(0, 16)
}

function formatMoney(value) {
  const num = Number(value || 0)
  return `${num >= 0 ? '+' : ''}${num.toFixed(2)}`
}

async function fetchDetailData() {
  try {
    loading.value = true
    const [listRes, summaryRes] = await Promise.all([
      getRechargeAccountDetailPageApi({
        queryModel: {
          studentId: filterState.value.studentId || undefined,
          startTime: filterState.value.operationDate?.[0],
          endTime: filterState.value.operationDate?.[1],
          flowTypes: filterState.value.flowTypes.length ? filterState.value.flowTypes : undefined,
        },
        pageRequestModel: {
          needTotal: true,
          pageSize: pagination.value.pageSize,
          pageIndex: pagination.value.current,
          skipCount: (pagination.value.current - 1) * pagination.value.pageSize,
        },
        sortModel: {
          orderByCreatedTime: 0,
        },
      }),
      getRechargeAccountExpendIncomeApi({
        studentId: filterState.value.studentId || undefined,
        startTime: filterState.value.operationDate?.[0],
        endTime: filterState.value.operationDate?.[1],
      }),
    ])

    dataSource.value = Array.isArray(listRes.result?.list) ? listRes.result.list : []
    pagination.value.total = listRes.result?.total || 0
    summary.value = {
      income: Number(summaryRes.result?.income || 0),
      expend: Number(summaryRes.result?.expend || 0),
    }
  }
  catch (error) {
    console.error('获取储值账户明细失败:', error)
    messageService.error('获取储值账户明细失败')
  }
  finally {
    loading.value = false
  }
}

function handleTableChange(pag) {
  pagination.value.current = pag.current
  pagination.value.pageSize = pag.pageSize
  fetchDetailData()
}

onMounted(() => {
  fetchDetailData()
})
</script>

<template>
  <div class="roll-call">
    <div class="bg-white rounded-4 rounded-lt-none rounded-rt-none px-5 py-3">
      <div class="filter-section">
        <div class="standard-filters">
          <span class="section-title mt-0.5 text-#222">筛选条件：</span>
          <checkbox-filter
            v-model:checked-values="filterState.operationDate"
            label="操作日期"
            type="dateSelectType"
          />
          <checkbox-filter
            v-model:checked-values="filterState.flowTypes"
            :options="FLOW_TYPE_OPTIONS"
            label="明细类型"
            type="checkbox"
          />
        </div>
        <div class="student-filter-wrap">
          <div class="label">
            学员/电话
          </div>
          <a-select
            v-model:value="searchKeyStuPhoneModel"
            class="searchKeyStuPhone"
            allow-clear
            show-search
            placeholder="搜索学员姓名"
            :filter-option="false"
            style="width: 240px"
            option-label-prop="label"
            :label-in-value="true"
            @change="handleStuPhoneSearchChange"
            @dropdown-visible-change="dropdownVisibleChangeFun"
            @search="handleSearchStuPhone"
            @popup-scroll="handleStuPhonePopupScroll"
          >
            <a-select-option
              v-for="item in stuPhoneSearchOptions"
              :key="item.id"
              :value="item.id"
              :data="item"
              :label="item.stuName"
            >
              <div class="flex flex-center mb-1">
                <div>
                  <img class="w-10 rounded-10" :src="item.avatarUrl" alt="">
                </div>
                <div class="ml-2 mr-3">
                  <div class="text-sm text-#666 leading-7">
                    {{ item.stuName }}
                  </div>
                  <div class="text-xs text-#888">
                    {{ item.mobile }}
                  </div>
                </div>
                <div>
                  <a-tag v-if="item.studentStatus == 1" :bordered="false" color="processing">
                    在读学员
                  </a-tag>
                  <a-tag v-if="item.studentStatus == 0" :bordered="false" color="orange">
                    意向学员
                  </a-tag>
                </div>
              </div>
            </a-select-option>
            <a-select-option
              v-if="stuPhoneSearchFinished && stuPhoneSearchOptions.length > 0"
              key="no-more"
              :value="-1"
              :label="undefined"
            >
              <div class="text-center text-#999 text-3">
                ～没有更多了～
              </div>
            </a-select-option>
          </a-select>
        </div>
      </div>
      <div class="selected-conditions">
        <span class="section-title text-#222">已选条件：</span>
        <div class="condition-tags">
          <a-popconfirm title="确定要清空所有条件吗？" @confirm="clearSelectableFilters">
            <a-tag color="red" class="clear-all mb-2">
              清空所有
              <DeleteOutlined class="text-3 ml-4px mt-0.6px" />
            </a-tag>
          </a-popconfirm>
          <a-tag color="blue" class="condition-tag mb-2">
            <div class="tag-content">
              <span class="condition-label">操作日期：</span>
              <div class="condition-values">
                <span class="value-item">{{ filterState.operationDate[0] }}~{{ filterState.operationDate[1] }}</span>
              </div>
            </div>
          </a-tag>
          <a-tag v-for="item in selectedFilterList.filter(tag => !tag.fixed)" :key="item.key" color="blue" class="condition-tag mb-2">
            <div class="tag-content">
              <span class="condition-label">{{ item.label }}：</span>
              <div class="condition-values">
                <span class="value-item">
                  {{ item.value }}
                  <CloseOutlined class="close-icon" @click.stop="clearFilter(item.key)" />
                </span>
              </div>
            </div>
          </a-tag>
        </div>
      </div>
    </div>
    <div class="bg-white rounded-4 mt-3 py-3 px-5">
      <div class="table-title flex justify-between mb2">
        <div class="total">
          收入 ￥{{ summary.income.toFixed(2) }}，支出 ￥{{ summary.expend.toFixed(2) }}
        </div>
        <div class="edit flex">
          <a-button>
            导出数据
          </a-button>
        </div>
      </div>
      <a-table
        :data-source="dataSource"
        :pagination="pagination"
        :columns="filteredColumns"
        :scroll="{ x: totalWidth }"
        :loading="loading"
        row-key="rechargeAccountFlowId"
        size="small"
        @change="handleTableChange"
      >
        <template #headerCell="{ column }">
          <template v-if="column.key === 'studentName'">
            <span class="mr-1">明细关联学员</span>
            <a-popover placement="top" trigger="hover">
              <template #content>
                <div class="w-120 text-#666 leading-6">
                  <div style="color:#222;font-weight:500;padding-bottom:8px;margin-bottom:12px;border-bottom:1px solid #f0f0f0;">
                    明细关联学员
                  </div>
                  <div>创建储值账户明细流水时的关联学员。</div>
                  <div>如：储值充值/退费时，默认显示主要关联学员；发起报课消费、退课退回时的订单关联学员。</div>
                </div>
              </template>
              <InfoCircleOutlined class="cursor-pointer text-#999 hover:text-#1677ff" />
            </a-popover>
          </template>
          <template v-if="column.key === 'rechargeAccountFlowSourceType'">
            <span class="mr-1">明细类型</span>
            <a-popover placement="top" trigger="hover">
              <template #content>
                <div class="w-160 text-#666 leading-6">
                  <div style="color:#222;font-weight:500;padding-bottom:8px;margin-bottom:12px;border-bottom:1px solid #f0f0f0;">
                    明细类型
                  </div>
                  <div>储值账户充值：完成储值账户充值订单时</div>
                  <div>储值账户退费：完成储值账户退费订单时</div>
                  <div>报名订单支出：报名缴费订单使用储值支付时</div>
                  <div>退费订单退回：退课、退教材费、退学杂费的金额退回储值账户时</div>
                  <div>作废储值充值：作废储值账户充值订单时</div>
                  <div>转课少补支出：转课存在差价时，使用储值支付转课订单时</div>
                  <div>转课退回：转课存在差价时，多出金额退回储值账户时</div>
                  <div>作废转课少补支出：作废使用储值支付的转课订单时</div>
                  <div>作废转课退回：作废转课退回订单时</div>
                  <div>场地预约支出：场地预约订单使用储值账户时</div>
                  <div>场地预约退回：场地预约订单作废/订单关闭时</div>
                  <div>作废储值退费：作废储值退费订单时</div>
                </div>
              </template>
              <InfoCircleOutlined class="cursor-pointer text-#999 hover:text-#1677ff" />
            </a-popover>
          </template>
          <template v-if="column.key === 'totalAmount'">
            <span class="flex justify-end">总计（元）</span>
          </template>
        </template>
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'rechargeAccountName'">
            <a-tooltip>
              <template #title>
                点击查看储值账户详情
              </template>
              <span class="cursor-pointer text-#1677ff hover-text-#0958d9" @click="handleOpenRechargeAccountDetail(record)">
                {{ record.rechargeAccountName || '-' }}
              </span>
            </a-tooltip>
          </template>
          <template v-if="column.key === 'studentName'">
            <student-avatar
              :id="record.studentId"
              :name="record.studentName || '-'"
              :phone="record.studentPhone || ''"
              :avatar-url="record.studentAvatar"
              :show-gender="false"
              :show-age="false"
              default-active-key="0"
            />
          </template>
          <template v-if="column.key === 'createTime'">
            {{ formatDate(record.createTime) }}
          </template>
          <template v-if="column.key === 'rechargeAccountFlowSourceType'">
            <span class="bg-#e6f0ff text-#06f text-3 px2 py1 rounded-10">{{ FLOW_TYPE_LABEL_MAP[record.rechargeAccountFlowSourceType] || '-' }}</span>
          </template>
          <template v-if="column.key === 'amount'">
            {{ formatMoney(record.amount) }}
          </template>
          <template v-if="column.key === 'givingAmount'">
            {{ formatMoney(record.givingAmount) }}
          </template>
          <template v-if="column.key === 'residualAmount'">
            {{ formatMoney(record.residualAmount) }}
          </template>
          <template v-if="column.key === 'sourceOrderNumber'">
            <a
              v-if="record.sourceOrderNumber && Number(record.sourceId || 0) > 0"
              class="text-#06f cursor-pointer"
              @click="handleOrderDetail(record.sourceId, record.sourceOrderNumber)"
            >
              {{ record.sourceOrderNumber }}
            </a>
            <span v-else-if="record.sourceOrderNumber">{{ record.sourceOrderNumber }}</span>
            <span v-else>-</span>
          </template>
          <template v-if="column.key === 'remark'">
            {{ record.remark || '-' }}
          </template>
          <template v-if="column.key === 'totalAmount'">
            <span class="flex justify-end font-800">{{ formatMoney(record.totalAmount) }}</span>
          </template>
        </template>
      </a-table>
    </div>
    <student-info-drawer v-model:open="openDrawer" />
    <order-detail-drawer v-model:open="openOrderDetailDrawer" :order-id="currentOrderId" />
    <recharge-account-detail-drawer
      v-model:open="openAccountDetailDrawer"
      :account="currentAccountRecord"
      @open-linked-order-detail="handleOpenLinkedOrderFromAccountDetailDrawer"
    />
  </div>
</template>

<style lang="less" scoped>
.filter-section {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.standard-filters {
  display: flex;
  align-items: flex-start;
  flex-wrap: wrap;
}

.section-title {
  white-space: nowrap;
}

.student-filter-wrap {
  align-items: center;
  display: flex;
  margin-left: auto;
}

.student-filter-wrap .label {
  border: 1px solid #f0f0f0;
  height: 32px;
  padding: 0 10px;
  line-height: 32px;
  text-align: center;
  min-width: 104px;
  border-radius: 8px 0 0 8px !important;
  color: #222;
  font-size: 14px;
  border-right: 0;
}

:deep(.student-filter-wrap .ant-select-selector) {
  border-radius: 0 6px 6px 0 !important;
}

.selected-conditions {
  display: flex;
  align-items: flex-start;
}

.condition-tags {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
}

.condition-tag {
  display: flex;
  align-items: center;
  border-radius: 4px;
}

.clear-all {
  cursor: pointer;
  display: flex;
  align-items: center;
}

.tag-content {
  display: flex;
  align-items: center;
}

.condition-values {
  display: flex;
  align-items: center;
}

.value-item {
  display: inline-flex;
  align-items: center;
}

.close-icon {
  margin-left: 6px;
  font-size: 12px;
  cursor: pointer;
  color: rgba(92, 92, 92, 0.45);
  transition: color 0.3s;
}

.close-icon:hover {
  color: rgba(0, 0, 0, 0.75);
}

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
</style>
