<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { DownOutlined } from '@ant-design/icons-vue'
import { useTableColumns } from '@/composables/useTableColumns'
import { getRechargeAccountItemPageApi, getRechargeAccountStatisticsApi } from '@/api/finance-center/recharge-account'
import { useStudentStore } from '@/stores/student'
import messageService from '@/utils/messageService'
import OrderDetailDrawer from '@/components/common/order-detail-drawer.vue'

const dataSource = ref([])
const openDrawer = ref(false)
const openAccountDrawer = ref(false)
const openRefundDrawer = ref(false)
const refundAccountId = ref(undefined)
const openAccountDetailDrawer = ref(false)
const openLinkedOrderDetailDrawer = ref(false)
const linkedOrderDetailId = ref('')
const checked = ref(true)
const stuId = ref(undefined)
const selectedStudentId = ref(undefined)
const currentAccountRecord = ref({})
const loading = ref(false)
const statistics = ref({
  rechargeAccountTotal: 0,
  residualAmountTotal: 0,
  amountTotal: 0,
  givingAmountTotal: 0,
})
const animatedSummary = ref({
  rechargeAccountTotal: 0,
  residualAmountTotal: 0,
  rechargeBalanceTotal: 0,
  givingAmountTotal: 0,
})
const summaryAnimationFrameId = ref(0)
const pagination = ref({
  current: 1,
  pageSize: 20,
  total: 0,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: total => `共 ${total} 条`,
})
const studentStore = useStudentStore()

function handleSeeStuData(studentId) {
  if (!studentId) {
    messageService.error('invalid studentId')
    return
  }
  studentStore.setStudentId(String(studentId))
  openDrawer.value = true
}
function handleAccount() {
  stuId.value = undefined
  openAccountDrawer.value = true
}

function handleOpenAccountDetail(record) {
  currentAccountRecord.value = { ...record }
  openAccountDetailDrawer.value = true
}

function handleOpenLinkedOrderFromAccountDetail(orderId) {
  const id = String(orderId || '').trim()
  if (!id)
    return
  openAccountDetailDrawer.value = false
  linkedOrderDetailId.value = id
  openLinkedOrderDetailDrawer.value = true
}

async function handleAccountDetailUpdated() {
  await Promise.all([fetchRechargeAccountList(), fetchRechargeAccountStatistics()])
  if (!openAccountDetailDrawer.value)
    return
  const accountId = currentAccountRecord.value?.rechargeAccountId
  if (!accountId)
    return
  currentAccountRecord.value = dataSource.value.find(item => String(item.rechargeAccountId) === String(accountId)) || currentAccountRecord.value
}

const mainStudentMap = computed(() => {
  const map = {}
  dataSource.value.forEach((item) => {
    const mainStudent = (item.rechargeAccountStudents || []).find(student => student.isMainStudent) || item.rechargeAccountStudents?.[0]
    if (mainStudent?.studentId) {
      map[item.rechargeAccountId] = mainStudent
    }
  })
  return map
})

const allColumns = ref([
  {
    title: '储值账户',
    dataIndex: 'rechargeAccount',
    key: 'rechargeAccount',
    width: 220,
  },
  {
    title: '关联学员',
    dataIndex: 'linkStudent',
    key: 'linkStudent',
    width: 250,
  },
  {
    title: '更新时间',
    dataIndex: 'updateTime',
    key: 'updateTime',
    width: 160,
    sorter: {
      compare: (a, b) => a.updateTime - b.updateTime,
    },
    defaultSortOrder: 'descend',
  },
  {
    title: '可用总余额（元）',
    dataIndex: 'canUseTotal',
    key: 'canUseTotal',
    width: 150,
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    fixed: 'right',
    width: 100,
  },

])
// 充值按钮触发
function rechargeFun(record) {
  const mainStudent = mainStudentMap.value?.[record?.rechargeAccountId]
  const mainStudentId = mainStudent?.studentId
  if (!mainStudentId) {
    messageService.error('该储值账户暂无关联学员，无法充值')
    return
  }
  openAccountDrawer.value = true
  stuId.value = mainStudentId
}

function refundFun(record) {
  if (!record?.rechargeAccountId) {
    messageService.error('无效的储值账户')
    return
  }
  refundAccountId.value = record.rechargeAccountId
  openRefundDrawer.value = true
}

const defaultOpenClassStatus = ref(1)
const { selectedValues, columnOptions, filteredColumns, totalWidth }
  = useTableColumns({
    storageKey: 'recharge-account-list', // 本地存储键名
    allColumns, // 原始列配置
    excludeKeys: ['action'], // 需要排除的列键
  })

watch([selectedStudentId, checked], () => {
  pagination.value.current = 1
  fetchRechargeAccountList()
})

async function fetchRechargeAccountList() {
  try {
    loading.value = true
    const { result } = await getRechargeAccountItemPageApi({
      queryModel: {
        studentId: selectedStudentId.value || undefined,
        showZeroBalanceAccount: !checked.value,
      },
      pageRequestModel: {
        needTotal: true,
        pageSize: pagination.value.pageSize,
        pageIndex: pagination.value.current,
        skipCount: (pagination.value.current - 1) * pagination.value.pageSize,
      },
      sortModel: {
        orderByUpdatedTime: 0,
      },
    })
    dataSource.value = Array.isArray(result?.list) ? result.list : []
    pagination.value.total = result?.total || 0
  }
  catch (error) {
    console.error('获取储值账户列表失败:', error)
    messageService.error('获取储值账户列表失败')
  }
  finally {
    loading.value = false
  }
}

async function fetchRechargeAccountStatistics() {
  try {
    const { result } = await getRechargeAccountStatisticsApi()
    statistics.value = {
      rechargeAccountTotal: Number(result?.rechargeAccountTotal || 0),
      residualAmountTotal: Number(result?.residualAmountTotal || 0),
      amountTotal: Number(result?.amountTotal || 0),
      givingAmountTotal: Number(result?.givingAmountTotal || 0),
    }
    animateSummaryValues()
  }
  catch (error) {
    console.error('获取储值账户统计失败:', error)
  }
}

function formatDate(dateStr) {
  if (!dateStr)
    return '-'
  return String(dateStr).replace('T', ' ').slice(0, 16)
}

function formatMoney(value) {
  return Number(value || 0).toFixed(2)
}

function stopSummaryAnimation() {
  if (summaryAnimationFrameId.value) {
    cancelAnimationFrame(summaryAnimationFrameId.value)
    summaryAnimationFrameId.value = 0
  }
}

function animateSummaryValues() {
  stopSummaryAnimation()

  const targets = {
    rechargeAccountTotal: Number(statistics.value.rechargeAccountTotal || 0),
    residualAmountTotal: Number(statistics.value.residualAmountTotal || 0),
    rechargeBalanceTotal: Math.max(0, Number(statistics.value.amountTotal || 0) - Number(statistics.value.residualAmountTotal || 0)),
    givingAmountTotal: Number(statistics.value.givingAmountTotal || 0),
  }

  animatedSummary.value = {
    rechargeAccountTotal: 0,
    residualAmountTotal: 0,
    rechargeBalanceTotal: 0,
    givingAmountTotal: 0,
  }

  const duration = 1200
  const startTime = performance.now()
  const easeOutCubic = progress => 1 - (1 - progress) ** 3

  const step = (currentTime) => {
    const progress = Math.min((currentTime - startTime) / duration, 1)
    const eased = easeOutCubic(progress)

    animatedSummary.value = {
      rechargeAccountTotal: targets.rechargeAccountTotal * eased,
      residualAmountTotal: targets.residualAmountTotal * eased,
      rechargeBalanceTotal: targets.rechargeBalanceTotal * eased,
      givingAmountTotal: targets.givingAmountTotal * eased,
    }

    if (progress < 1) {
      summaryAnimationFrameId.value = requestAnimationFrame(step)
      return
    }

    animatedSummary.value = { ...targets }
    summaryAnimationFrameId.value = 0
  }

  summaryAnimationFrameId.value = requestAnimationFrame(step)
}

function handleTableChange(pag) {
  pagination.value.current = pag.current
  pagination.value.pageSize = pag.pageSize
  fetchRechargeAccountList()
}

onMounted(async () => {
  await Promise.all([fetchRechargeAccountList(), fetchRechargeAccountStatistics()])
})

onBeforeUnmount(() => {
  stopSummaryAnimation()
})
</script>

<template>
  <div class="roll-call">
    <div class="databord bg-white  pt-0.1 pb-1 pl-5 pr-5 rounded-4   rounded-lt-none rounded-rt-none ">
        <div class="cards-scroll mt-3 mb-2">
          <div class="cards-grid gap-4">
          <div class="card-item bg-#fbfcff h-22.5 cursor-pointer rounded-5 hover-bg-#0066ff0d min-w-285px">
            <div class="contentMain">
              <div class="contentMainLeft">
                机构总账户余额（元）
              </div>
              <div class="contentMainRight font-900 mt1 mr-4 text-#06f">
                {{ formatMoney(animatedSummary.rechargeAccountTotal) }}
              </div>
            </div>
            <div class="contentSub">
              <div class="contentSubLeft">
                充值余额+残联余额+赠送余额
              </div>
            </div>
          </div>
          <div class="card-item bg-#fbfcff h-22.5 cursor-pointer rounded-5 hover-bg-#0066ff0d min-w-285px">
            <div class="contentMain">
              <div class="contentMainLeft">
                残联余额（元）
              </div>
              <div class="contentMainRight font-900 mt1 mr-4 text-#222">
                {{ formatMoney(animatedSummary.residualAmountTotal) }}
              </div>
            </div>
            <div class="contentSub">
              <div class="contentSubLeft">
                机构账户充值金额
              </div>
            </div>
          </div>
          <div class="card-item bg-#fbfcff h-22.5 cursor-pointer rounded-5 hover-bg-#0066ff0d min-w-285px">
            <div class="contentMain">
              <div class="contentMainLeft">
                充值余额（元）
              </div>
              <div class="contentMainRight font-900 mt1 mr-4 text-#222">
                {{ formatMoney(animatedSummary.rechargeBalanceTotal) }}
              </div>
            </div>
            <div class="contentSub">
              <div class="contentSubLeft">
                机构账户充值金额
              </div>
            </div>
          </div>
          <div class="card-item bg-#fbfcff h-22.5 cursor-pointer rounded-5 hover-bg-#0066ff0d min-w-285px">
            <div class="contentMain">
              <div class="contentMainLeft">
                赠送余额（元）
              </div>
              <div class="contentMainRight font-900 mt1 mr-4 text-#222">
                {{ formatMoney(animatedSummary.givingAmountTotal) }}
              </div>
            </div>
            <div class="contentSub">
              <div class="contentSubLeft">
                机构账户赠送金额
              </div>
            </div>
          </div>
        </div>
        </div>
    </div>
    <div class="bg-white rounded-4 mt-3 px-5 py-3">
      <div class="filter-right">
        <div class="label">
          学员/电话
        </div>
        <student-select
          v-model="selectedStudentId"
          width="240px"
          placeholder="搜索姓名/手机号"
          allow-clear
        />
      </div>
    </div>
    <div class="bg-white rounded-4  mt-3 py-3 px-5">
      <div class="table-title flex justify-between mb2">
        <div class="total">
          共 {{ pagination.total }} 个账户
        </div>
        <div class="edit flex">
          <div class="flex flex-items-center mr3">
            <a-switch v-model:checked="checked" size="small" /> <span class="ml1.2 text-#444 text-3.5">隐藏余额为0的账户</span>
          </div>
          <a-button class="mr-3" type="primary" @click="handleAccount()">
            账户充值
          </a-button>
          <a-dropdown>
            <template #overlay>
              <a-menu>
                <a-menu-item key="1">
                  导入储值账户
                </a-menu-item>
                <a-menu-item key="2">
                  导出储值账户
                </a-menu-item>
              </a-menu>
            </template>
            <a-button>
              导入/导出储值账户
              <DownOutlined :style="{ fontSize: '10px' }" />
            </a-button>
          </a-dropdown>
          <!-- 自定义字段 -->
          <!-- <customize-code v-model:checkedValues="selectedValues" :options="columnOptions" :total="allColumns.length-1"
            :num="selectedValues.length-1" /> -->
        </div>
      </div>
      <a-table
        :data-source="dataSource" :pagination="pagination" :columns="filteredColumns"
        :scroll="{ x: totalWidth }" :loading="loading" row-key="rechargeAccountId" size="small" @change="handleTableChange"
      >
        <template #headerCell="{ column }">
          <template v-if="column.key === 'canUseTotal'">
            <span class="flex-center">{{ column.title }}</span>
          </template>
        </template>
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'rechargeAccount'">
            <div class="flex justify-between flex-items-start gap-2 pr1">
              <a-tooltip class="min-w-0 flex-1">
                <template #title>
                  点击查看储值账户详情
                </template>
                <span class="cursor-pointer text-#1677ff hover-text-#0958d9" @click="handleOpenAccountDetail(record)">
                  {{ record.rechargeAccountName || '默认账户' }}
                </span>
              </a-tooltip>
              <span class="text-3 text-#888 whitespace-nowrap flex-shrink-0">含残联 ¥{{ formatMoney(record.residualBalance) }}</span>
            </div>
          </template>
          <template v-if="column.key === 'linkStudent'">
            <div class="text-#222 font-500">
              {{ (record.rechargeAccountStudents || []).length }}位
            </div>

            <div class="text-3 text-#222 font-500">
              <a-tooltip
                v-for="student in (record.rechargeAccountStudents || [])"
                :key="student.studentId"
              >
                <template #title>
                  查看学员详情
                </template>
                <span class="hover-text-#06f cursor-pointer mr-1" @click="handleSeeStuData(student.studentId)">
                  {{ student.studentName }}
                </span>
              </a-tooltip>
              <span v-if="!(record.rechargeAccountStudents || []).length">-</span>
            </div>
          </template>
          <template v-if="column.key === 'updateTime'">
            {{ formatDate(record.updateTime) }}
          </template>
          <template v-if="column.key === 'canUseTotal'">
            <span class="font-800 text-#222 flex-center">{{ formatMoney(record.balanceTotal) }}</span>
          </template>
          <template v-else-if="column.key === 'action'">
            <span class="flex action">
              <a-space :size="14">
                <a @click="rechargeFun(record)">充值</a>
                <a @click="refundFun(record)">退款</a>
              </a-space>
            </span>
          </template>
        </template>
      </a-table>
    </div>
    <student-info-drawer v-model:open="openDrawer" />
    <account-recharge-drawer
      v-model:open="openAccountDrawer"
      :stu-id="stuId"
      @submitted="() => { fetchRechargeAccountList(); fetchRechargeAccountStatistics() }"
    />
    <account-refund-drawer
      v-model:open="openRefundDrawer"
      :recharge-account-id="refundAccountId"
      @submitted="() => { fetchRechargeAccountList(); fetchRechargeAccountStatistics() }"
    />
    <recharge-account-detail-drawer
      v-model:open="openAccountDetailDrawer"
      :account="currentAccountRecord"
      @updated="handleAccountDetailUpdated"
      @open-linked-order-detail="handleOpenLinkedOrderFromAccountDetail"
    />
    <order-detail-drawer v-model:open="openLinkedOrderDetailDrawer" :order-id="linkedOrderDetailId" />
  </div>
</template>

<style lang="less" scoped>
.cards-scroll {
  overflow-x: auto;
  overflow-y: hidden;
  padding-bottom: 8px;

  &::-webkit-scrollbar {
    height: 6px;
  }

  &::-webkit-scrollbar-track {
    background: transparent;
  }

  &::-webkit-scrollbar-thumb {
    background: #d9e6ff;
    border-radius: 999px;
  }

  &::-webkit-scrollbar-thumb:hover {
    background: #b7d0ff;
  }

  scrollbar-width: thin;
  scrollbar-color: #d9e6ff transparent;
}

.cards-grid {
  display: grid;
  grid-auto-flow: column;
  grid-auto-columns: minmax(320px, 1fr);
  min-width: max-content;
}

.card-item {
  min-width: 320px;
}

.filter-right {
  display: flex;
  justify-content: flex-end;
  align-items: center;
}

.filter-right .label {
  border: 1px solid #f0f0f0;
  height: 32px;
  line-height: 32px;
  min-width: 104px;
  padding: 0 16px 0 8px;
  text-align: center;
  color: #222;
  font-size: 14px;
  border-radius: 8px 0 0 8px;
  border-right: 0;
  white-space: nowrap;
}

:deep(.filter-right .ant-select-selector) {
  border-radius: 0 6px 6px 0 !important;
}

.contentMain {
  box-sizing: content-box;
  padding: 16px 0px 6px 18px;
  height: 30px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: bold;

  .contentMainLeft {
    font-size: 14px;
    font-weight: 500;
    color: #222;
    flex-shrink: 0;
  }

  .contentMainRight {
    min-width: 72px;
    height: 30px;
    font-size: 30px;
    font-weight: 700;
    font-family: DINAlternate-Bold, DINAlternate;
    line-height: 30px;
    flex-shrink: 0;
    text-align: right;
  }
}

.contentSub {
  padding: 0 18px;
  height: 16px;
  line-height: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: bold;

  .contentSubLeft {
    font-size: 13px;
    color: #888;
  }

  .contentSubRight {
    font-size: 12px;
    color: #06f;
  }
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
