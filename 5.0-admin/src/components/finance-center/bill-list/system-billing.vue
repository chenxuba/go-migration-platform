<script setup>
import { computed, createVNode, onMounted, ref, watch } from 'vue'
import { CloseOutlined, DeleteOutlined, DownOutlined, ExclamationCircleFilled, ExclamationCircleOutlined, InfoCircleOutlined } from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'
import { useTableColumns } from '@/composables/useTableColumns'
import { cancelConfirmLedgerApi, confirmLedgerApi, getLedgerListApi, getLedgerStatisticsApi } from '@/api/finance-center/ledger'
import { getUserListApi } from '@/api/internal-manage/staff-manage'
import { getRecommenderPageApi } from '@/api/enroll-center/intention-student'
import messageService from '@/utils/messageService'

const sourceTypeOptions = [
  { id: 1, value: '系统同步' },
  { id: 2, value: '手动记账' },
]

const ledgerStatusOptions = [
  { id: 0, value: '待确认' },
  { id: 1, value: '已确认' },
  { id: 2, value: '退款中' },
  { id: 3, value: '退款失败' },
]

const payMethodMap = {
  1: '微信',
  2: '支付宝',
  3: '银行转账',
  4: 'POS机',
  5: '现金',
  6: '其他',
}

const ledgerTypeMap = {
  1: '收入',
  2: '支出',
}

const sourceTypeMap = {
  1: '系统同步',
  2: '手动记账',
}

const ledgerStatusMap = {
  0: '待确认',
  1: '已确认',
  2: '退款中',
  3: '退款失败',
}

/** 与后端 inst_ledger.ledger_sub_category_id 一致 */
const LEDGER_SUB_RECHARGE_ACCOUNT_REFUND = 'recharge-account-refund'

function isLedgerStoredValueRefund(record) {
  if (!record)
    return false
  if (String(record.ledgerSubCategoryId || '') === LEDGER_SUB_RECHARGE_ACCOUNT_REFUND)
    return true
  const name = String(record.ledgerSubCategoryName || '')
  return name.includes('储值账户退费') || name.includes('储值账户退款')
}

/** 待确认时的主操作文案：退费为「确认支出」，其余为「确认到账」 */
function getLedgerConfirmPrimaryActionLabel(record) {
  if (!record)
    return '确认到账'
  if (Number(record.ledgerConfirmStatus) === 1)
    return '取消确认'
  return isLedgerStoredValueRefund(record) ? '确认支出' : '确认到账'
}

const filterState = ref({
  orderNumber: '',
  payDate: [],
  ledgerNumber: '',
  bankSlipNo: '',
  orderId: '',
  studentId: '',
  sourceTypes: [],
  dealStaffId: '',
  ledgerConfirmStatuses: [],
  confirmStaffId: '',
  confirmTime: [],
})

const dataSource = ref([])
const loading = ref(false)
const statistics = ref({
  incomeAmount: 0,
  expenditureAmount: 0,
  balanceAmount: 0,
  totalConfirm: 0,
  totalUnConfirm: 0,
  totalRefunding: 0,
  totalRefundFailed: 0,
})
const pagination = ref({
  current: 1,
  pageSize: 20,
  total: 0,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: total => `共 ${total} 条`,
})

const staffOptions = ref([])
const studentOptions = ref([])
const allColumns = ref([
  { title: '账单编号', dataIndex: 'ledgerNumber', key: 'ledgerNumber', fixed: 'left', width: 200, required: true },
  { title: '记账类型', dataIndex: 'sourceType', key: 'sourceType', width: 120 },
  { title: '收款方式', dataIndex: 'payMethod', key: 'payMethod', width: 120 },
  { title: '收款账户', dataIndex: 'accountName', key: 'accountName', width: 130 },
  { title: '支付单号', dataIndex: 'bankSlipNo', key: 'bankSlipNo', width: 140 },
  { title: '关联订单', dataIndex: 'orderNumber', key: 'orderNumber', width: 210 },
  { title: '对方账户', dataIndex: 'reciprocalAccount', key: 'reciprocalAccount', width: 130 },
  { title: '经办人', dataIndex: 'dealStaffName', key: 'dealStaffName', width: 120 },
  { title: '收支类型', dataIndex: 'type', key: 'type', width: 100 },
  { title: '一级分类', dataIndex: 'ledgerCategoryName', key: 'ledgerCategoryName', width: 120 },
  { title: '二级分类', dataIndex: 'ledgerSubCategoryName', key: 'ledgerSubCategoryName', width: 120 },
  { title: '支付日期', dataIndex: 'payTime', key: 'payTime', width: 120 },
  { title: '操作时间', dataIndex: 'createdTime', key: 'createdTime', width: 160 },
  { title: '账单备注', dataIndex: 'paymentVoucher', key: 'paymentVoucher', width: 180 },
  { title: '学员/电话', dataIndex: 'studentName', key: 'studentName', width: 180 },
  { title: '办理内容', dataIndex: 'productItems', key: 'productItems', width: 160 },
  { title: '确认人员', dataIndex: 'confirmStaffName', key: 'confirmStaffName', width: 120 },
  { title: '确认时间', dataIndex: 'confirmTime', key: 'confirmTime', width: 160 },
  { title: '确认备注', dataIndex: 'confirmRemark', key: 'confirmRemark', width: 160 },
  { title: '金额', dataIndex: 'amount', key: 'amount', fixed: 'right', width: 100, required: true },
  { title: '账单状态', dataIndex: 'ledgerConfirmStatus', key: 'ledgerConfirmStatus', fixed: 'right', width: 100, required: true },
  { title: '操作', dataIndex: 'action', key: 'action', fixed: 'right', width: 180 },
])

const { selectedValues, columnOptions, filteredColumns, totalWidth }
  = useTableColumns({
    storageKey: 'system-billing',
    allColumns,
    excludeKeys: ['action'],
  })

const openBillDrawer = ref(false)
const currentLedger = ref(null)
const openOrderDetailDrawer = ref(false)
const currentOrderId = ref('')
const confirmModalOpen = ref(false)
const confirmSubmitting = ref(false)
const confirmTargetLedger = ref(null)

const confirmLedgerModalTitle = computed(() =>
  confirmTargetLedger.value && isLedgerStoredValueRefund(confirmTargetLedger.value)
    ? '确认支出'
    : '确认到账',
)
const confirmLedgerModalOkText = computed(() =>
  confirmTargetLedger.value && isLedgerStoredValueRefund(confirmTargetLedger.value)
    ? '确认支出'
    : '确认到账',
)
const confirmRemarkText = ref('')

const selectedFilterList = computed(() => {
  const list = []
  if (String(filterState.value.orderNumber || '').trim())
    list.push({ key: 'orderNumber', label: '订单编号', value: String(filterState.value.orderNumber).trim() })
  if (Array.isArray(filterState.value.payDate) && filterState.value.payDate.filter(Boolean).length === 2)
    list.push({ key: 'payDate', label: '支付日期', value: `${filterState.value.payDate[0]} ~ ${filterState.value.payDate[1]}` })
  if (String(filterState.value.ledgerNumber || '').trim())
    list.push({ key: 'ledgerNumber', label: '账单编号', value: String(filterState.value.ledgerNumber).trim() })
  if (String(filterState.value.bankSlipNo || '').trim())
    list.push({ key: 'bankSlipNo', label: '支付单号', value: String(filterState.value.bankSlipNo).trim() })
  if (String(filterState.value.orderId || '').trim())
    list.push({ key: 'orderId', label: '关联订单', value: String(filterState.value.orderId).trim() })
  if (String(filterState.value.studentId || '').trim())
    list.push({ key: 'studentId', label: '学员/电话', value: studentNameById(filterState.value.studentId) || String(filterState.value.studentId).trim() })
  if (Array.isArray(filterState.value.sourceTypes) && filterState.value.sourceTypes.length)
    list.push({ key: 'sourceTypes', label: '记账类型', value: formatOptionValues(filterState.value.sourceTypes, sourceTypeOptions) })
  if (String(filterState.value.dealStaffId || '').trim())
    list.push({ key: 'dealStaffId', label: '经办人', value: staffNameById(filterState.value.dealStaffId) || filterState.value.dealStaffId })
  if (Array.isArray(filterState.value.ledgerConfirmStatuses) && filterState.value.ledgerConfirmStatuses.length)
    list.push({ key: 'ledgerConfirmStatuses', label: '账单状态', value: formatOptionValues(filterState.value.ledgerConfirmStatuses, ledgerStatusOptions) })
  if (String(filterState.value.confirmStaffId || '').trim())
    list.push({ key: 'confirmStaffId', label: '确认人员', value: staffNameById(filterState.value.confirmStaffId) || filterState.value.confirmStaffId })
  if (Array.isArray(filterState.value.confirmTime) && filterState.value.confirmTime.filter(Boolean).length === 2)
    list.push({ key: 'confirmTime', label: '确认时间', value: `${filterState.value.confirmTime[0]} ~ ${filterState.value.confirmTime[1]}` })
  return list
})

const hasSelectedFilters = computed(() => selectedFilterList.value.length > 0)
const currentLedgerIndex = computed(() => {
  if (!currentLedger.value?.id)
    return -1
  return dataSource.value.findIndex(item => String(item.id) === String(currentLedger.value.id))
})
const hasPrevLedger = computed(() => currentLedgerIndex.value > 0)
const hasNextLedger = computed(() => currentLedgerIndex.value >= 0 && currentLedgerIndex.value < dataSource.value.length - 1)
const isCurrentLedgerConfirmed = computed(() => currentLedger.value?.ledgerConfirmStatus === 1)

function formatOptionValues(values, options) {
  return options
    .filter(item => values.includes(item.id))
    .map(item => item.value)
    .join('、')
}

function staffNameById(staffId) {
  const target = staffOptions.value.find(item => String(item.id) === String(staffId))
  return target?.name || target?.nickName || ''
}

function studentNameById(studentId) {
  const target = studentOptions.value.find(item => String(item.id) === String(studentId))
  return target?.stuName || target?.name || ''
}

function formatDate(dateStr, withTime = true) {
  if (!dateStr)
    return '-'
  const text = String(dateStr).replace('T', ' ')
  return withTime ? text.slice(0, 16) : text.slice(0, 10)
}

function formatMoney(amount = 0, type = 1) {
  const prefix = type === 2 ? '-' : '+'
  return `${prefix}${Number(amount || 0).toFixed(2)}`
}

function getSourceTypeText(sourceType) {
  return sourceTypeMap[sourceType] || '-'
}

function getLedgerStatusText(status) {
  return ledgerStatusMap[status] || '-'
}

function getPayMethodText(payMethod) {
  return payMethodMap[payMethod] || '-'
}

function getLedgerTypeText(type) {
  return ledgerTypeMap[type] || '-'
}

/** 系统账单：储值账户充值/退费，办理内容展示储值账户名 */
function isStoredValueRechargeLedger(record) {
  const t = Number(record?.orderType || 0)
  if (t === 2 || t === 4)
    return true
  const sub = String(record?.ledgerSubCategoryName || '')
  if (sub.includes('储值账户充值') || sub.includes('储值账户退费') || sub.includes('储值账户退款'))
    return true
  return false
}

function getStoredValueLedgerAccountDisplayName(record) {
  const keys = ['rechargeAccountName', 'storedValueAccountName']
  for (const k of keys) {
    const s = String(record?.[k] || '').trim()
    if (s)
      return s
  }
  // 接口未返账户名时不使用学员姓名兜底，避免与「账户名」语义混淆
  return '-'
}

function formatLedgerProductItemsDisplay(record) {
  // 优先用列表接口的 productItems（含储值账户名）；接口未单独返 rechargeAccountName 时原先会误判为 '-'
  if (Array.isArray(record?.productItems) && record.productItems.length)
    return record.productItems.join('、')
  if (isStoredValueRechargeLedger(record))
    return getStoredValueLedgerAccountDisplayName(record)
  return '-'
}

function getLedgerStatusClass(status) {
  if (status === 1)
    return 'status-confirmed'
  if (status === 2)
    return 'status-refunding'
  if (status === 3)
    return 'status-failed'
  return 'status-pending'
}

function canEditLedger(record) {
  if (!record)
    return false
  if (record.sourceType === 1)
    return false
  if (record.ledgerConfirmStatus === 1)
    return false
  return true
}

function canDeleteLedger(record) {
  return canEditLedger(record)
}

function updateCurrentLedgerById(id) {
  if (!currentLedger.value || String(currentLedger.value.id) !== String(id))
    return
  const latest = dataSource.value.find(item => String(item.id) === String(id))
  if (latest)
    currentLedger.value = latest
}

function openLedgerByIndex(index) {
  const target = dataSource.value[index]
  if (!target)
    return
  currentLedger.value = target
}

function goPrevLedger() {
  if (!hasPrevLedger.value)
    return
  openLedgerByIndex(currentLedgerIndex.value - 1)
}

function goNextLedger() {
  if (!hasNextLedger.value)
    return
  openLedgerByIndex(currentLedgerIndex.value + 1)
}

async function handleConfirmLedger(record) {
  if (!record?.id)
    return
  try {
    if (record.ledgerConfirmStatus === 1) {
      Modal.confirm({
        title: '确定取消?',
        centered: true,
        icon: createVNode(ExclamationCircleFilled, { style: 'color:#fa8c16' }),
        content: createVNode('div', {}, [
          createVNode('div', {}, '取消后账单将被标记为待确认，确认人和确认时间将'),
          createVNode('div', {}, '被清空。'),
        ]),
        okText: '确定取消',
        cancelText: '再想想',
        onOk: async () => {
          await cancelConfirmLedgerApi({ id: record.id })
          messageService.success('取消确认成功')
          await fetchAll()
          updateCurrentLedgerById(record.id)
        },
      })
      return
    }
    confirmTargetLedger.value = record
    confirmRemarkText.value = record?.confirmRemark?.text || ''
    confirmModalOpen.value = true
  }
  catch (error) {
    console.error('确认到账失败:', error)
    messageService.error(error?.message || '操作失败')
  }
}

async function submitConfirmLedger() {
  if (!confirmTargetLedger.value?.id)
    return
  const outgoing = isLedgerStoredValueRefund(confirmTargetLedger.value)
  try {
    confirmSubmitting.value = true
    const ledgerId = confirmTargetLedger.value.id
    await confirmLedgerApi({
      id: ledgerId,
      confirmRemark: {
        text: confirmRemarkText.value || '',
        images: [],
      },
    })
    messageService.success(outgoing ? '确认支出成功' : '确认到账成功')
    closeConfirmModal()
    await fetchAll()
    updateCurrentLedgerById(ledgerId)
  }
  catch (error) {
    console.error(outgoing ? '确认支出失败:' : '确认到账失败:', error)
    messageService.error(error?.message || (outgoing ? '确认支出失败' : '确认到账失败'))
  }
  finally {
    confirmSubmitting.value = false
  }
}

function closeConfirmModal() {
  confirmModalOpen.value = false
  confirmSubmitting.value = false
  confirmTargetLedger.value = null
  confirmRemarkText.value = ''
}

function handleEditLedger(record) {
  if (!canEditLedger(record))
    return
  messageService.info(`编辑账单功能待实现：${record?.ledgerNumber || ''}`)
}

function handleDeleteLedger(record) {
  if (!canDeleteLedger(record))
    return
  messageService.info(`删除账单功能待实现：${record?.ledgerNumber || ''}`)
}

function getRequestPayload() {
  return {
    sortModel: {},
    queryModel: {
      orderNumber: filterState.value.orderNumber || undefined,
      bankSlipNo: filterState.value.bankSlipNo || undefined,
      ledgerNumber: filterState.value.ledgerNumber || undefined,
      orderId: filterState.value.orderId || undefined,
      studentId: filterState.value.studentId || undefined,
      sourceTypes: filterState.value.sourceTypes.length ? filterState.value.sourceTypes : undefined,
      dealStaffId: filterState.value.dealStaffId || undefined,
      ledgerConfirmStatuses: filterState.value.ledgerConfirmStatuses.length ? filterState.value.ledgerConfirmStatuses : undefined,
      confirmStaffId: filterState.value.confirmStaffId || undefined,
      payStartTime: filterState.value.payDate?.[0] || undefined,
      payEndTime: filterState.value.payDate?.[1] || undefined,
      confirmStartTime: filterState.value.confirmTime?.[0] || undefined,
      confirmEndTime: filterState.value.confirmTime?.[1] || undefined,
    },
    pageRequestModel: {
      needTotal: true,
      pageSize: pagination.value.pageSize,
      pageIndex: pagination.value.current,
      skipCount: (pagination.value.current - 1) * pagination.value.pageSize,
    },
  }
}

async function loadStaffOptions(searchKey = '') {
  try {
    const res = await getUserListApi({
      queryModel: {
        searchKey,
      },
      pageRequestModel: {
        needTotal: false,
        pageIndex: 1,
        pageSize: 50,
      },
    })
    if (res.code === 200) {
      const list = Array.isArray(res.result) ? res.result : []
      staffOptions.value = list.map(item => ({
        id: String(item.id),
        name: item.nickName,
        phone: item.mobile,
      }))
    }
  }
  catch (error) {
    console.error('加载员工列表失败:', error)
  }
}

async function loadStudentOptions(searchKey = '') {
  try {
    const res = await getRecommenderPageApi({
      pageRequestModel: {
        needTotal: true,
        pageIndex: 1,
        pageSize: 20,
        skipCount: 0,
      },
      queryModel: {
        searchKey,
      },
      sortModel: {},
    })
    if (res.code === 200) {
      studentOptions.value = Array.isArray(res.result) ? res.result : []
    }
  }
  catch (error) {
    console.error('加载学员列表失败:', error)
  }
}

async function fetchLedgerList() {
  try {
    loading.value = true
    const { result } = await getLedgerListApi(getRequestPayload())
    dataSource.value = Array.isArray(result?.list) ? result.list : []
    pagination.value.total = result?.total || 0
  }
  catch (error) {
    console.error('获取账单列表失败:', error)
    messageService.error('获取账单列表失败')
  }
  finally {
    loading.value = false
  }
}

async function fetchLedgerStatistics() {
  try {
    const { result } = await getLedgerStatisticsApi(getRequestPayload())
    statistics.value = {
      incomeAmount: Number(result?.incomeAmount || 0),
      expenditureAmount: Number(result?.expenditureAmount || 0),
      balanceAmount: Number(result?.balanceAmount || 0),
      totalConfirm: Number(result?.totalConfirm || 0),
      totalUnConfirm: Number(result?.totalUnConfirm || 0),
      totalRefunding: Number(result?.totalRefunding || 0),
      totalRefundFailed: Number(result?.totalRefundFailed || 0),
    }
  }
  catch (error) {
    console.error('获取账单统计失败:', error)
  }
}

async function fetchAll() {
  await Promise.all([fetchLedgerList(), fetchLedgerStatistics()])
}

function openLedgerDetail(record) {
  currentLedger.value = record
  openBillDrawer.value = true
}

function handleOrderDetail(orderId) {
  currentOrderId.value = String(orderId || '')
  openOrderDetailDrawer.value = true
}

function handleTableChange(pag) {
  pagination.value.current = pag.current
  pagination.value.pageSize = pag.pageSize
  fetchAll()
}

function resetAllFilters() {
  filterState.value = {
    orderNumber: '',
    payDate: [],
    ledgerNumber: '',
    bankSlipNo: '',
    orderId: '',
    studentId: '',
    sourceTypes: [],
    dealStaffId: '',
    ledgerConfirmStatuses: [],
    confirmStaffId: '',
    confirmTime: [],
  }
}

function clearFilter(key) {
  switch (key) {
    case 'payDate':
    case 'confirmTime':
      filterState.value[key] = []
      break
    case 'sourceTypes':
    case 'ledgerConfirmStatuses':
      filterState.value[key] = []
      break
    default:
      filterState.value[key] = ''
  }
}

watch(
  filterState,
  () => {
    pagination.value.current = 1
    fetchAll()
  },
  { deep: true },
)

onMounted(async () => {
  await loadStaffOptions()
  await loadStudentOptions()
  fetchAll()
})
</script>

<template>
  <div>
    <div class="filter-wrap bg-white pl-3 pr-3 rounded-lb-4 rounded-rb-4">
      <div class="filter-section pt-3">
        <span class="section-title mt-0.5 text-#222">筛选条件：</span>
        <div class="standard-filters">
          <checkbox-filter v-model:checked-values="filterState.orderNumber" label="订单编号" type="inputType" placeholder="请输入订单编号" />
          <checkbox-filter v-model:checked-values="filterState.payDate" label="支付日期" type="dateSelectType" />
          <checkbox-filter v-model:checked-values="filterState.ledgerNumber" label="账单编号" type="inputType" placeholder="请输入账单编号" />
          <checkbox-filter v-model:checked-values="filterState.bankSlipNo" label="支付单号" type="inputType" placeholder="请输入支付单号" />
          <checkbox-filter v-model:checked-values="filterState.orderId" label="关联订单" type="inputType" placeholder="请输入关联订单" />
          <checkbox-filter
            v-model:checked-values="filterState.studentId"
            :options="studentOptions"
            label="学员/电话"
            type="radio"
            category="stu"
            placeholder="请输入学员/电话"
            @on-dropdown-visible-change="loadStudentOptions"
            @on-search="loadStudentOptions"
          />
          <checkbox-filter v-model:checked-values="filterState.sourceTypes" :options="sourceTypeOptions" label="记账类型" type="checkbox" />
          <checkbox-filter
            v-model:checked-values="filterState.dealStaffId"
            :options="staffOptions"
            label="经办人"
            type="radio"
            category="teacher"
            placeholder="请输入经办人"
            @on-dropdown-visible-change="loadStaffOptions"
            @on-search="loadStaffOptions"
          />
          <checkbox-filter v-model:checked-values="filterState.ledgerConfirmStatuses" :options="ledgerStatusOptions" label="账单状态" type="checkbox" />
          <checkbox-filter
            v-model:checked-values="filterState.confirmStaffId"
            :options="staffOptions"
            label="确认人员"
            type="radio"
            category="teacher"
            placeholder="请输入确认人员"
            @on-dropdown-visible-change="loadStaffOptions"
            @on-search="loadStaffOptions"
          />
          <checkbox-filter v-model:checked-values="filterState.confirmTime" label="确认时间" type="dateSelectType" />
        </div>
      </div>
      <div v-if="hasSelectedFilters" class="selected-conditions mt-2">
        <span class="section-title text-#222">已选条件：</span>
        <div class="condition-tags">
          <a-popconfirm title="确定要清空所有条件吗？" @confirm="resetAllFilters">
            <a-tag color="red" class="clear-all mb-2">
              清空已选
              <DeleteOutlined class="text-3 ml-4px mt-0.6px" />
            </a-tag>
          </a-popconfirm>
          <a-tag v-for="item in selectedFilterList" :key="item.key" color="blue" class="condition-tag mb-2">
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
    <div class="student-list mt-3 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            共 {{ pagination.total }} 条信息
          </div>
          <div class="edit flex">
            <a-dropdown class="mr-2">
              <template #overlay>
                <a-menu>
                  <a-menu-item key="1">
                    批量导出
                  </a-menu-item>
                  <a-menu-item key="3">
                    导出记录
                  </a-menu-item>
                </a-menu>
              </template>
              <a-button>
                导出数据
                <DownOutlined :style="{ fontSize: '10px' }" />
              </a-button>
            </a-dropdown>
            <customize-code
              v-model:checked-values="selectedValues" :options="columnOptions"
              :total="allColumns.length - 1" :num="selectedValues.length"
            />
          </div>
        </div>
        <div class="table-content mt-2">
          <div class="tip">
            收入：¥ {{ statistics.incomeAmount.toFixed(2) }}， 支出：¥ {{ statistics.expenditureAmount.toFixed(2) }}，
            结算：¥ {{ statistics.balanceAmount.toFixed(2) }}， 已确认 {{ statistics.totalConfirm }} 条，
            待确认 {{ statistics.totalUnConfirm }} 条， 退款中 {{ statistics.totalRefunding }} 条，
            退款失败 {{ statistics.totalRefundFailed }} 条
          </div>
          <a-table
            :data-source="dataSource" :pagination="pagination" :columns="filteredColumns"
            :scroll="{ x: totalWidth }" :loading="loading" row-key="id" size="small" @change="handleTableChange"
          >
            <template #headerCell="{ column }">
              <template v-if="column.key === 'action'">
                <span class="mr-1">操作</span>
                <a-popover placement="top" trigger="hover">
                  <template #content>
                    <div class="w-50 text-#666 leading-6">
                      <div style="color:#222;font-weight:500;padding-bottom:8px;margin-bottom:12px;border-bottom:1px solid #f0f0f0;">
                        操作说明
                      </div>
                      <div>已确认账单不支持编辑或删除</div>
                      <div>系统同步账单不支持编辑或删除</div>
                      <div>收银宝账单不支持编辑</div>
                    </div>
                  </template>
                  <InfoCircleOutlined class="cursor-pointer text-#999 hover:text-#1677ff" />
                </a-popover>
              </template>
            </template>
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'ledgerNumber'">
                <a-tooltip>
                  <template #title>
                    查看账单详情
                  </template>
                  <a class="text-#06f cursor-pointer" @click="openLedgerDetail(record)">{{ record.ledgerNumber }}</a>
                </a-tooltip>
              </template>
              <template v-if="column.key === 'sourceType'">
                {{ getSourceTypeText(record.sourceType) }}
              </template>
              <template v-if="column.key === 'payMethod'">
                {{ getPayMethodText(record.payMethod) }}
              </template>
              <template v-if="column.key === 'accountName'">
                {{ record.accountName || '-' }}
              </template>
              <template v-if="column.key === 'bankSlipNo'">
                {{ record.bankSlipNo || '-' }}
              </template>
              <template v-if="column.key === 'orderNumber'">
                <a-tooltip>
                  <template #title>
                    查看订单详情
                  </template>
                  <a v-if="record.orderId" class="text-#06f cursor-pointer" @click="handleOrderDetail(record.orderId)">{{ record.orderNumber || '-' }}</a>
                  <span v-else>{{ record.orderNumber || '-' }}</span>
                </a-tooltip>
              </template>
              <template v-if="column.key === 'reciprocalAccount'">
                {{ record.reciprocalAccount || '-' }}
              </template>
              <template v-if="column.key === 'dealStaffName'">
                {{ record.dealStaffName || '-' }}
              </template>
              <template v-if="column.key === 'type'">
                {{ getLedgerTypeText(record.type) }}
              </template>
              <template v-if="column.key === 'createdTime'">
                {{ formatDate(record.createdTime) }}
              </template>
              <template v-if="column.key === 'paymentVoucher'">
                {{ record.paymentVoucher?.text || '-' }}
              </template>
              <template v-if="column.key === 'productItems'">
                {{ formatLedgerProductItemsDisplay(record) }}
              </template>
              <template v-if="column.key === 'confirmStaffName'">
                {{ record.confirmStaffName || '-' }}
              </template>
              <template v-if="column.key === 'ledgerCategoryName'">
                {{ record.ledgerCategoryName || '-' }}
              </template>
              <template v-if="column.key === 'ledgerSubCategoryName'">
                {{ record.ledgerSubCategoryName || '-' }}
              </template>
              <template v-if="column.key === 'payTime'">
                {{ formatDate(record.payTime, false) }}
              </template>
              <template v-if="column.key === 'confirmTime'">
                {{ formatDate(record.confirmTime) }}
              </template>
              <template v-if="column.key === 'confirmRemark'">
                {{ record.confirmRemark?.text || '-' }}
              </template>
              <template v-if="column.key === 'studentName'">
                <student-avatar
                  :id="record.studentId"
                  :name="record.studentName || '-'"
                  :phone="record.studentPhone || ''"
                  :show-gender="false"
                  :show-age="false"
                  default-active-key="0"
                />
              </template>
              <template v-if="column.key === 'amount'">
                <span :class="record.type === 2 ? 'amount-expense' : 'amount-income'">
                  {{ formatMoney(record.amount, record.type) }}
                </span>
              </template>
              <template v-if="column.key === 'ledgerConfirmStatus'">
                <span class="status-pill" :class="getLedgerStatusClass(record.ledgerConfirmStatus)">
                  {{ getLedgerStatusText(record.ledgerConfirmStatus) }}
                </span>
              </template>
              <template v-if="column.key === 'action'">
                <a-space :size="12">
                  <a class="font500" @click="handleConfirmLedger(record)">{{ getLedgerConfirmPrimaryActionLabel(record) }}</a>
                  <a
                    :class="canEditLedger(record) ? 'font500' : 'action-disabled'"
                    @click="handleEditLedger(record)"
                  >编辑</a>
                  <a
                    :class="canDeleteLedger(record) ? 'font500' : 'action-disabled'"
                    @click="handleDeleteLedger(record)"
                  >删除</a>
                </a-space>
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>

    <!-- 账单详情 -->
    <a-drawer
      v-model:open="openBillDrawer" :body-style="{ padding: '0', background: '#fff' }" :closable="false"
      width="800px" placement="right"
    >
      <!-- 自定义头部 -->
      <template #title>
        <div class="custom-header flex justify-between h-4 flex-items-center">
          <div class="text-5">
            账单详情
          </div>
          <a-button type="text" class="close-btn" @click="openBillDrawer = false">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>
      <div class="contenter pb20">
        <div class="h-11 bg-#e6f0ff text-#06f flex flex-items-center font-400 pl3.5">
          <ExclamationCircleOutlined class="font-800 mr1" />
          {{ currentLedger?.sourceType === 1 ? '自动同步的订单账单暂不支持编辑或删除' : '手动记账能力后续补齐' }}
        </div>
        <div class="ledger-amount-panel h-36 pt-8 pb-6 flex-center flex-col border border-b-#eee border-solid border-x-none border-t-none">
          <img
            v-if="isCurrentLedgerConfirmed"
            class="confirmed-stamp"
            src="https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/14953/static/png/confirmed-C02Wt4HX.png"
            alt="已确认"
          >
          <div class="mb1">
            <svg width="40px" height="40px" viewBox="0 0 40 40" class="icon income">
              <title>切片</title>
              <g id="\u9875\u9762-1" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd">
                <g id="\u8BBE\u7F6E-\u5207\u6362" transform="translate(-576.000000, -160.000000)">
                  <g id="\u7F16\u7EC4-22\u5907\u4EFD-4" transform="translate(539.000000, 152.000000)">
                    <g id="\u7F16\u7EC4-21" transform="translate(28.000000, 8.000000)">
                      <g id="\u7F16\u7EC4-34\u5907\u4EFD-50" transform="translate(9.000000, 0.000000)">
                        <circle id="\u692D\u5706\u5F62" fill="#e6f0ff" cx="20" cy="20" r="20" />
                        <path
                          id="\u5F62\u72B6\u7ED3\u5408"
                          d="M28,9 C29.6568542,9 31,10.3431458 31,12 L31,12 L31,26 C31,28.7614237 28.7614237,31 26,31 L26,31 L11.5,31 C9.01471863,31 7,28.9852814 7,26.5 L7,26.5 L7,25 C7,23.3431458 8.34314575,22 10,22 L10,22 L14,22 L14,12 C14,10.4023191 15.24892,9.09633912 16.8237272,9.00509269 L17,9 Z M14,24 L10,24 C9.44771525,24 9,24.4477153 9,25 L9,25 L9,26.5 C9,27.8807119 10.1192881,29 11.5,29 L11.5,29 L14,29 L14,24 Z M28,11 L17,11 C16.4477153,11 16,11.4477153 16,12 L16,12 L16,22 L20,22 C21.5976809,22 22.9036609,23.24892 22.9949073,24.8237272 L23,25 L23,26.5 C23,27.8807119 24.1192881,29 25.5,29 L25.5,29 L26,29 C27.5976809,29 28.9036609,27.75108 28.9949073,26.1762728 L28.9949073,26.1762728 L29,26 L29,12 C29,11.4477153 28.5522847,11 28,11 L28,11 Z M20,24 L16,24 L16,29 L21.758,29 C21.3146237,28.3362246 21.0439378,27.5542808 21.004898,26.7118357 L21.004898,26.7118357 L21,26.5 L21,25 C21,24.4477153 20.5522847,24 20,24 L20,24 Z M22,18 C22.5522847,18 23,18.4477153 23,19 C23,19.5128358 22.6139598,19.9355072 22.1166211,19.9932723 L22,20 L19,20 C18.4477153,20 18,19.5522847 18,19 C18,18.4871642 18.3860402,18.0644928 18.8833789,18.0067277 L19,18 L22,18 Z M25.6498204,14 C26.2021052,14 26.6498204,14.4477153 26.6498204,15 C26.6498204,15.5128358 26.2637802,15.9355072 25.7664415,15.9932723 L25.6498204,16 L19,16 C18.4477153,16 18,15.5522847 18,15 C18,14.4871642 18.3860402,14.0644928 18.8833789,14.0067277 L19,14 L25.6498204,14 Z" fill="#06f"
                        />
                      </g>
                    </g>
                  </g>
                </g>
              </g>
            </svg>
          </div>
          <div class="text-8 font-400 text-#222">
            {{ formatMoney(currentLedger?.amount, currentLedger?.type) }}
          </div>
        </div>
        <div class="mt-4 px-6 border border-b-#eee border-solid border-x-none border-t-none pb2">
          <a-descriptions :column="2" :content-style="{ color: '#888' }">
            <a-descriptions-item label="账单编号">
              {{ currentLedger?.ledgerNumber || '-' }}
            </a-descriptions-item>
            <a-descriptions-item label="来源">
              {{ getSourceTypeText(currentLedger?.sourceType) }}
            </a-descriptions-item>
            <a-descriptions-item label="收款方式">
              {{ getPayMethodText(currentLedger?.payMethod) }}
            </a-descriptions-item>
            <a-descriptions-item label="收款账户">
              {{ currentLedger?.accountName || '-' }}
            </a-descriptions-item>
            <a-descriptions-item label="支付单号">
              {{ currentLedger?.bankSlipNo || '-' }}
            </a-descriptions-item>
            <a-descriptions-item label="关联订单">
              {{ currentLedger?.orderNumber || '-' }}
              <span v-if="currentLedger?.orderId" class="text-#06f ml2 cursor-pointer" @click="handleOrderDetail(currentLedger?.orderId)">查看</span>
            </a-descriptions-item>
            <a-descriptions-item label="对方账户">
              {{ currentLedger?.reciprocalAccount || '-' }}
            </a-descriptions-item>
            <a-descriptions-item label="经办人">
              {{ currentLedger?.dealStaffName || '-' }}
            </a-descriptions-item>
            <a-descriptions-item label="一级分类">
              {{ currentLedger?.ledgerCategoryName || '-' }}
            </a-descriptions-item>
            <a-descriptions-item label="二级分类">
              {{ currentLedger?.ledgerSubCategoryName || '-' }}
            </a-descriptions-item>
            <a-descriptions-item label="支付日期">
              {{ formatDate(currentLedger?.payTime, false) }}
            </a-descriptions-item>
            <a-descriptions-item label="操作时间">
              {{ formatDate(currentLedger?.createdTime) }}
            </a-descriptions-item>
            <a-descriptions-item label="学员信息">
              {{ currentLedger?.studentName || '-' }} {{ currentLedger?.studentPhone || '' }}
            </a-descriptions-item>
            <a-descriptions-item label="办理内容">
              {{ formatLedgerProductItemsDisplay(currentLedger) }}
            </a-descriptions-item>
            <a-descriptions-item label="账单备注">
              {{ currentLedger?.paymentVoucher?.text || '-' }}
            </a-descriptions-item>
          </a-descriptions>
        </div>
        <div class="mt-4 px-6 border border-b-#eee border-solid border-x-none border-t-none pb2">
          <a-descriptions :column="2" :content-style="{ color: '#888' }">
            <a-descriptions-item label="确认时间">
              {{ formatDate(currentLedger?.confirmTime) }}
            </a-descriptions-item>
            <a-descriptions-item label="账单状态">
              {{ getLedgerStatusText(currentLedger?.ledgerConfirmStatus) }}
            </a-descriptions-item>
            <a-descriptions-item label="确认人员">
              {{ currentLedger?.confirmStaffName || '-' }}
            </a-descriptions-item>
          </a-descriptions>
        </div>
        <div class="mt-4 px-6 border border-b-#eee border-solid border-x-none border-t-none pb2">
          <a-descriptions :column="1" :content-style="{ color: '#888' }">
            <a-descriptions-item label="确认备注">
              {{ currentLedger?.confirmRemark?.text || '-' }}
            </a-descriptions-item>
          </a-descriptions>
        </div>
      </div>
      <template #footer>
        <div class="h-15 flex flex-center justify-between">
          <div>
            <a-button
              :type="isCurrentLedgerConfirmed ? 'default' : 'primary'"
              :danger="isCurrentLedgerConfirmed"
              class="h-12 w-35 text-5"
              @click="handleConfirmLedger(currentLedger)"
            >
              {{ getLedgerConfirmPrimaryActionLabel(currentLedger) }}
            </a-button>
          </div>
          <a-space :size="14">
            <a-tooltip :title="hasPrevLedger ? '' : '当前为列表页第一个'">
              <span>
                <a-button type="primary" ghost class="h-12 w-35 text-5" :disabled="!hasPrevLedger" @click="goPrevLedger">
                  上一个
                </a-button>
              </span>
            </a-tooltip>
            <a-tooltip :title="hasNextLedger ? '' : '当前为列表页最后一个'">
              <span>
                <a-button type="primary" ghost class="h-12 w-35 text-5" :disabled="!hasNextLedger" @click="goNextLedger">
                  下一个
                </a-button>
              </span>
            </a-tooltip>
          </a-space>
        </div>
      </template>
    </a-drawer>
    <a-modal
      v-model:open="confirmModalOpen"
      :title="confirmLedgerModalTitle"
      :confirm-loading="confirmSubmitting"
      :ok-text="confirmLedgerModalOkText"
      cancel-text="取消"
      @ok="submitConfirmLedger"
      @cancel="closeConfirmModal"
    >
      <a-form layout="vertical">
        <a-form-item label="确认备注">
          <a-textarea
            v-model:value="confirmRemarkText"
            :rows="4"
            :maxlength="100"
            placeholder="请输入确认备注"
            show-count
          />
        </a-form-item>
      </a-form>
    </a-modal>
    <order-detail-drawer v-model:open="openOrderDetailDrawer" :order-id="currentOrderId" />
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

.studentStatus {
  span.dot {
    border-radius: 50%;
    display: inline-block;
    height: 6px;
    position: relative;
    vertical-align: middle;
    width: 6px;
    margin-right: 4px;
    background: var(--pro-ant-color-primary);
  }
}

.tip {
  padding: 10px 24px 10px 14px;
  background: #e6f0ff;
  color: #333;

  a {
    color: var(--pro-ant-color-primary);
  }
}

.filter-section {
  display: flex;
  align-items: flex-start;
}

.section-title {
  white-space: nowrap;
}

.standard-filters {
  display: flex;
  flex-wrap: wrap;
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

.clear-all {
  cursor: pointer;
  display: flex;
  align-items: center;
}

.student-cell {
  line-height: 1.5;
}

.amount-income {
  color: #222;
  font-weight: 600;
}

.amount-expense {
  color: #222;
  font-weight: 600;
}

.status-pill {
  border-radius: 12px;
  display: inline-flex;
  font-size: 12px;
  font-weight: 500;
  line-height: 20px;
  padding: 0 10px;
}

.status-pending {
  background: #fff7e6;
  color: #fa8c16;
}

.status-confirmed {
  background: #f6ffed;
  color: #52c41a;
}

.status-refunding {
  background: #e6f4ff;
  color: #1677ff;
}

.status-failed {
  background: #fff1f0;
  color: #f5222d;
}

.action-disabled {
  color: #d9d9d9 !important;
  cursor: not-allowed;
  pointer-events: auto;
}

.ledger-amount-panel {
  position: relative;
}

.confirmed-stamp {
  position: absolute;
  right: 28px;
  top: 10px;
  width: 122px;
  pointer-events: none;
}

.upNew {
  position: relative;

  &::before {
    position: absolute;
    top: -12px;
    left: -22px;
    z-index: 999;
    width: 39px;
    height: 22px;
    background: url(data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAE4AAAAsCAYAAADLlo5MAAAAAXNSR0IArs4c6QAABjtJREFUaEPtm3lo1EcUxz+zRrwtgmiNf4hBvEFkd0m8Fa1XbdGWBlERFVsFj1ovPEGsfxk86omK4IEiFg/EQkHFekATknjfSETQKKKoVfFKdsrbybq7yR6//e3+4prkwWJI3nsz8913z6hIgrTWipycbHy+b/H5slAqE8hEa/m3aRKqUyeq1CvgEVCK1qW4XCW4XH+Rn1+glNJ2F1J2BLXXOwStfwK+R+uv7ej47DJKPQaOodSfqrDwZKL7SQg4nZ2dQ1nZaqBfogulOf85MjIWqoKCfKv7tASc9nqz0DoPrX+wqviL5FPqMEotUIWFJfH2Hxc4v1v6fAeBFvGU1ZC/P8flyo3nvjGB0273LJRah9b1aggo1o6hVDla/6aKizdGE4gKnHa71wO/WlupxnL9oYqL50Q6XUTg/JYGG2osHIkdbHYky6sCXEWp8Xetc8+oPqnKUWp45ZgXBpw/e/p8RbUoEVi1PUkYntBsGw6cx3OoxpccVqGqzKfUYVVU9GPg15+Aqyhu/7Wrt1bIZWT0ChTJQeDc7nNA35QC0KULTJliVC5dCh8+2FffsiUsXgxZWbBsGVy/bl2XywXdukH9+nDhgnW5qpznVXGxv2vyA1dR5J5IRmNE2X79YN068yf5+e3b5JbYvBmys+H4cVixoqqujAwQgAOfVq2gZ08j07w5PH8Oo0fDmzf29+FyfSOJwgDndm8HfravLYpkssBNngwDBgSVt2gBbdvCx49w+3b4otu2QY8eMHVq5M1obWTWrIGLF+0fVantqqhomvKPhrxeGbmkfsqRLHDikmIhVmj5cmjXzgAnFnXzJpSWms+9e1BUBC9fWtEUm0emKoWFmcrRpJAscJ07Q2YmNG1qYtuVK8FDNWgAbjcUFEB5Ody4YUAW4M6ehblzkwcpmgZJEtrr/R2fb5kjqyQLnGyqQwfYtQvevYPhw6GszGxVXFjc7u5dGDvW/G769OoBzuVapbTbvQ8Yl7bAycYOHjQWN2cOnD9vtirJYdQoA+qmTdULHOxX2uM5jdYDHQduy5bY5YiUKgJQKPXqBU2aQP/+MHIk5OfD0aOGQ8qbZs1gwwYTx0pKYOhQY3Hi0lu3Rj/SpUsmwdglpf4R4G6jdUe7OmLKhbpqvAUkcA8eHM516JAJ+FZoxw5QKnpWDdUhX8KTJ1a0RuZR6o64qlxmOHOxEgqcfMsSxKORZMLKAX3lSmjdOijRuDFIUS1UWZ/UdlKqiMWJNQVqNUkijRqZtV/JUTEx8elT+8DBa7G4/9C6WTJaosqmIjmEKu/UCfZJSAYGDoTXr8OXjpQccnNh4UK4dQsmTEjZMavPVe10Dg0bGmsJkGTYQOwaMyYcuBcvYNq0qlnVQeCqJznYAW7iRJg925qVDBsG48eDyJw8CYsWGTnHgEvnckRca8aMIHAS/KUfFZJ6TtqoAElpsmABDBkCu3fDxorrAseAS/cCOF6Mk+D//r3h2rMHunaFVauCZYtjwJlLZmfmcKlIDu3bw9q1JoseOBBMDpIIpD+9fz/ozqdOwVdfmQ5CelNHXTWdm3w5+KRJMHOmKX7F/QJZVWqxI0egXj0YMcIU12fOGLDEbR/LCwcHY5zo1h7PNrT+xVoUToArFRYnLVX37rB6NVy+HF6OSNslZUlengFKelcBsE+fYPxzylX9wJnb+vQbZEqxu3dv0IrEDUPruL59TTy7ds0MATweY3Xz5gW/XSeB84Pndp9N+DGNVODSfEejNm1A+k2hY8eCk41YRvvwocmKQuvXg4Ajjb00+JULYMmqs2bBnTuwZImRkc5B4mGAHAfOTpKQqUROTgK+a4FVGnS5p5Bpr4AtBbCAIe4qHyk3JIsOGhQcGsyfb9qoq1dBpsah5DRwFbEusevBceNiW5wFnKqwPHhgRkVCYrHSIchkZf9+6FgxizhxwlzcBEj62Z07TYw7ffozAJfOF9IyxJSJsCQIybCVL35kUvzoUXhRLBBKXde7Nzx7ZrJwiqjuCYRNIOse3aQSOH+8q3vmFRPSuoeFqba4gL5a+JTVEpRx3wD73ba2PJ62BJlhsgTcJ+szRXJeyh/nJLDhdGFNCLhK7puLUt858nQiXdCJsQ9bwH0C8Ev4L0kOfQn/A6jssToWH7guAAAAAElFTkSuQmCC);
    background-size: contain;
    content: "";
  }
}

/* 添加旋转动画 */
@keyframes icon-rotate {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(180deg);
  }
}

.close-btn {
  &:hover {
    background: transparent;

    .close-icon {
      animation: icon-rotate 0.3s linear;
    }
  }
}
</style>
