<script setup>
import { computed, onMounted, ref } from 'vue'
import { DownOutlined, InfoCircleOutlined } from '@ant-design/icons-vue'
import { useTableColumns } from '@/composables/useTableColumns'
import { getOrderDetailPagedApi } from '@/api/finance-center/order-manage'
import messageService from '@/utils/messageService'
import dayjs from 'dayjs'

const currentYearStart = dayjs().startOf('year').format('YYYY-MM-DD')
const today = dayjs().format('YYYY-MM-DD')
const defaultCreateTimeVals = ref([currentYearStart, today])
const defaultOrderStatusVals = ref([1, 3, 7, 8, 6, 2])
const dataSource = ref([])
const loading = ref(false)
const pagination = ref({
  current: 1,
  pageSize: 50,
  total: 0,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: total => `共 ${total} 条`,
})

const displayArray = ref([
  'orderNumber',
  'orderType',
  'orderTag',
  'orderSource',
  'orderStatus',
  'handleContent',
  'enrollType',
  'productType',
  'courseCategory',
  'salesPerson',
  'createUser',
  'dealDate',
  'createTime',
  'orderArrearStatus',
  'latestPaidTime',
])

const queryState = ref({
  orderNumber: undefined,
  orderTypeList: undefined,
  orderTagIds: undefined,
  orderSourceList: undefined,
  orderStatusList: [1, 3, 7, 8, 6, 2],
  courseIds: undefined,
  enrollTypes: undefined,
  productTypes: undefined,
  courseCategoryId: undefined,
  salePersonId: undefined,
  creatorId: undefined,
  studentId: undefined,
  dealDateBegin: undefined,
  dealDateEnd: undefined,
  createdTimeBegin: currentYearStart,
  createdTimeEnd: today,
  orderArrearStatus: undefined,
  latestPaidTimeBegin: undefined,
  latestPaidTimeEnd: undefined,
})

const orderTypeMap = {
  1: '报名续费',
  2: '储值账户充值',
  3: '退课',
  4: '储值账户退费',
  5: '转课',
  6: '退教材费',
  7: '退学杂费',
}

const orderSourceMap = {
  1: '线下办理',
  2: '微校报名',
  3: '线下导入',
  4: '续费订单',
}

const orderStatusMap = {
  1: '待付款',
  2: '审批中',
  3: '已完成',
  4: '已关闭',
  5: '已作废',
  6: '待处理',
  7: '退费中',
  8: '已退费',
}

const enrollTypeMap = {
  1: '新报',
  2: '续费',
  3: '扩科',
  0: '无',
  4: '无',
}

const productTypeMap = {
  1: '课程',
  2: '教学用品',
  3: '约课付费',
  4: '储值账户',
  5: '场地预约',
  6: '学杂费',
}

const unitMap = {
  1: '课时',
  2: '天',
  3: '月',
  4: '年',
  5: '元',
}

const allColumns = ref([
  { title: '订单编号', dataIndex: 'orderNumber', key: 'orderNumber', fixed: 'left', width: 210, required: true },
  { title: '报名学员', dataIndex: 'studentName', key: 'studentName', fixed: 'left', width: 180, required: true },
  { title: '订单类型', dataIndex: 'orderType', key: 'orderType', width: 120 },
  { title: '订单来源', dataIndex: 'orderSource', key: 'orderSource', width: 120 },
  { title: '订单标签', dataIndex: 'tagNames', key: 'tagNames', width: 160 },
  { title: '订单状态', dataIndex: 'orderStatus', key: 'orderStatus', width: 120 },
  { title: '办理内容', dataIndex: 'productName', key: 'productName', width: 160 },
  { title: '报读类型', dataIndex: 'enrollType', key: 'enrollType', width: 120 },
  { title: '商品类型', dataIndex: 'productType', key: 'productType', width: 120 },
  { title: '课程类别', dataIndex: 'productCategoryName', key: 'productCategoryName', width: 140 },
  { title: '报价单名称', dataIndex: 'quoteName', key: 'quoteName', width: 160 },
  { title: '报价单', dataIndex: 'skuName', key: 'skuName', width: 180 },
  { title: '购买份数', dataIndex: 'skuCount', key: 'skuCount', width: 100 },
  { title: '购买数量', dataIndex: 'quantity', key: 'quantity', width: 120 },
  { title: '赠送数量', dataIndex: 'freeQuantity', key: 'freeQuantity', width: 120 },
  { title: '单课优惠名称', dataIndex: 'discountName', key: 'discountName', width: 140 },
  { title: '单课优惠', dataIndex: 'discountNumber', key: 'discountNumber', width: 120 },
  { title: '分摊整单优惠', dataIndex: 'shareDiscount', key: 'shareDiscount', width: 140 },
  { title: '应收/应退', dataIndex: 'shouldAmount', key: 'shouldAmount', width: 120 },
  { title: '分摊优惠券', dataIndex: 'shareCouponAmount', key: 'shareCouponAmount', width: 120 },
  { title: '分摊储值账户充值余额', dataIndex: 'shareRechargeAccountAmount', key: 'shareRechargeAccountAmount', width: 170 },
  { title: '分摊储值账户赠送余额', dataIndex: 'shareRechargeAccountGivingAmount', key: 'shareRechargeAccountGivingAmount', width: 170 },
  { title: '实收/实退', dataIndex: 'actualPaidAmount', key: 'actualPaidAmount', width: 120 },
  { title: '欠费金额', dataIndex: 'arrearAmount', key: 'arrearAmount', width: 120 },
  { title: '坏账金额', dataIndex: 'badDebtAmount', key: 'badDebtAmount', width: 120 },
  { title: '平账抵扣', dataIndex: 'chargeAgainstAmount', key: 'chargeAgainstAmount', width: 120 },
  { title: '订单销售员', dataIndex: 'salePersonName', key: 'salePersonName', width: 120 },
  { title: '经办人', dataIndex: 'staffName', key: 'staffName', width: 120 },
  { title: '经办日期', dataIndex: 'dealDate', key: 'dealDate', width: 120, sorter: true },
  { title: '创建时间', dataIndex: 'createdTime', key: 'createdTime', width: 160, sorter: true },
])

const { selectedValues, columnOptions, filteredColumns, totalWidth } = useTableColumns({
  storageKey: 'order-detail-list',
  allColumns,
  excludeKeys: [],
})

const defaultStudentAvatar = 'https://cdn.schoolpal.cn/schoolpal/next-erp/avator_male.png?x-oss-process=image/resize,w_120'
const openOrderDetailDrawer = ref(false)
const currentOrderId = ref('')

function handleOrderDetail(orderId) {
  currentOrderId.value = String(orderId || '')
  openOrderDetailDrawer.value = true
}

function handleOrderDetailDrawerClosed() {
  fetchOrderDetailList()
}

function resetQueryState() {
  queryState.value = {
    orderNumber: undefined,
    orderTypeList: undefined,
    orderTagIds: undefined,
    orderSourceList: undefined,
    orderStatusList: [1, 3, 7, 8, 6, 2],
    courseIds: undefined,
    enrollTypes: undefined,
    productTypes: undefined,
    courseCategoryId: undefined,
    salePersonId: undefined,
    creatorId: undefined,
    studentId: undefined,
    dealDateBegin: undefined,
    dealDateEnd: undefined,
    createdTimeBegin: currentYearStart,
    createdTimeEnd: today,
    orderArrearStatus: undefined,
    latestPaidTimeBegin: undefined,
    latestPaidTimeEnd: undefined,
  }
}

async function fetchOrderDetailList(id, type) {
  try {
    loading.value = true
    const res = await getOrderDetailPagedApi({
      queryModel: queryState.value,
      sortModel: {},
      pageRequestModel: {
        needTotal: true,
        pageSize: pagination.value.pageSize,
        pageIndex: pagination.value.current,
        skipCount: (pagination.value.current - 1) * pagination.value.pageSize,
      },
    })
    if (res.code === 200) {
      dataSource.value = Array.isArray(res.result?.list) ? res.result.list : []
      pagination.value.total = res.result?.total || 0
      if (type) {
        allFilterRef.value?.clearQuickFilter(id, type)
      }
      return
    }
    messageService.error(res.message || '获取订单明细列表失败')
  }
  catch (error) {
    console.error('获取订单明细列表失败:', error)
    messageService.error('获取订单明细列表失败')
  }
  finally {
    loading.value = false
  }
}

const handleFilterUpdate = (updates, isClearAll = false, id, type) => {
  if (isClearAll) {
    resetQueryState()
  }
  else {
    Object.entries(updates).forEach(([key, value]) => {
      queryState.value[key] = value
    })
  }
  pagination.value.current = 1
  fetchOrderDetailList(id, type)
}

function mapRange(fieldPrefix, value, isClearAll, id, type) {
  if (Array.isArray(value) && value.length === 2) {
    handleFilterUpdate({
      [`${fieldPrefix}Begin`]: value[0],
      [`${fieldPrefix}End`]: value[1],
    }, isClearAll, id, type)
    return
  }
  handleFilterUpdate({
    [`${fieldPrefix}Begin`]: undefined,
    [`${fieldPrefix}End`]: undefined,
  }, isClearAll, id, type)
}

const allFilterRef = ref(null)
const filterUpdateHandlers = computed(() => ({
  'update:orderNumberFilter': (val, isClearAll, id, type) => handleFilterUpdate({ orderNumber: val || undefined }, isClearAll, id, type),
  'update:orderTypeFilter': (val, isClearAll, id, type) => handleFilterUpdate({ orderTypeList: Array.isArray(val) && val.length ? val : undefined }, isClearAll, id, type),
  'update:orderTagFilter': (val, isClearAll, id, type) => handleFilterUpdate({ orderTagIds: Array.isArray(val) && val.length ? val.map(String) : undefined }, isClearAll, id, type),
  'update:orderSourceFilter': (val, isClearAll, id, type) => handleFilterUpdate({ orderSourceList: Array.isArray(val) && val.length ? val : undefined }, isClearAll, id, type),
  'update:orderStatusFilter': (val, isClearAll, id, type) => handleFilterUpdate({ orderStatusList: Array.isArray(val) && val.length ? val : undefined }, isClearAll, id, type),
  'update:handleContentFilter': (val, isClearAll, id, type) => handleFilterUpdate({ courseIds: Array.isArray(val) && val.length ? val.map(String) : undefined }, isClearAll, id, type),
  'update:enrollTypeFilter': (val, isClearAll, id, type) => handleFilterUpdate({ enrollTypes: Array.isArray(val) && val.length ? val : undefined }, isClearAll, id, type),
  'update:productTypeFilter': (val, isClearAll, id, type) => handleFilterUpdate({ productTypes: Array.isArray(val) && val.length ? val : undefined }, isClearAll, id, type),
  'update:courseCategoryFilter': (val, isClearAll, id, type) => handleFilterUpdate({ courseCategoryId: val || undefined }, isClearAll, id, type),
  'update:salesPersonFilter': (val, isClearAll, id, type) => handleFilterUpdate({ salePersonId: val || undefined }, isClearAll, id, type),
  'update:createUserFilter': (val, isClearAll, id, type) => handleFilterUpdate({ creatorId: val || undefined }, isClearAll, id, type),
  'update:dealDateFilter': (val, isClearAll, id, type) => mapRange('dealDate', val, isClearAll, id, type),
  'update:createTimeFilter': (val, isClearAll, id, type) => mapRange('createdTime', val, isClearAll, id, type),
  'update:orderArrearStatusFilter': (val, isClearAll, id, type) => handleFilterUpdate({ orderArrearStatus: Array.isArray(val) && val.length ? val : undefined }, isClearAll, id, type),
  'update:latestPaidTimeFilter': (val, isClearAll, id, type) => mapRange('latestPaidTime', val, isClearAll, id, type),
  'update:stuPhoneSearchFilter': (val, isClearAll, id, type) => handleFilterUpdate({ studentId: val || undefined }, isClearAll, id, type),
}))

function handleTableChange(pag) {
  pagination.value.current = pag.current
  pagination.value.pageSize = pag.pageSize
  fetchOrderDetailList()
}

function formatDate(value) {
  if (!value || String(value).startsWith('0001-01-01'))
    return '-'
  return String(value).replace('T', ' ').slice(0, 16)
}

function formatDateOnly(value) {
  if (!value || String(value).startsWith('0001-01-01'))
    return '-'
  return String(value).split('T')[0]
}

function isRechargeAccountOrder(record) {
  return [2, 4].includes(Number(record?.orderType || 0)) || Number(record?.productType || 0) === 4
}

function isRefundDisplayOrder(record) {
  return [3, 4, 6, 7].includes(Number(record?.orderType || 0))
}

function getTimeSlotTotalDays(record) {
  if (Number(record?.chargingMode || 0) !== 2 || !record?.validDate || !record?.endDate) {
    return null
  }
  const startDate = dayjs(record.validDate)
  const endDate = dayjs(record.endDate)
  if (!startDate.isValid() || !endDate.isValid()) {
    return null
  }
  return Math.max(endDate.diff(startDate, 'day') + 1, 0)
}

function getQuantityText(value, unit, record, type = 'purchase') {
  if (isRechargeAccountOrder(record))
    return '-'
  if (value === undefined || value === null)
    return '-'
  if (Number(record?.chargingMode || 0) === 2) {
    const totalDays = getTimeSlotTotalDays(record)
    const freeDays = Number(record?.freeQuantity || 0)
    if (totalDays !== null) {
      if (type === 'gift') {
        return `${freeDays}天`
      }
      return `${Math.max(totalDays - freeDays, 0)}天`
    }
  }
  return `${value}${unitMap[unit] || ''}`
}

function getQuoteDisplay(record) {
  if (isRechargeAccountOrder(record))
    return '-'
  if (Number(record?.chargingMode || 0) === 3 && record?.quoteName === '自定义') {
    return `充值金额${Number(record.tuition || 0).toFixed(2)}元`
  }
  const quantity = Number(record.quantity || 0)
  const tuition = Number(record.tuition || 0)
  const unitText = unitMap[record.skuUnit] || ''
  if (quantity > 0 && unitText) {
    return `${quantity}${unitText}/${tuition.toFixed(2)}元`
  }
  return tuition ? `${tuition.toFixed(2)}元` : '-'
}

function formatMoney(value) {
  return `${Number(value || 0).toFixed(2)}`
}

function getDiscountText(record) {
  if (isRechargeAccountOrder(record))
    return '-'
  if (!record.discountType || !record.discountNumber) {
    return '-'
  }
  if (record.discountType === 2) {
    return `${record.discountNumber}折`
  }
  return `-¥${Number(record.discountNumber || 0).toFixed(2)}`
}

function getEnrollTypeText(record) {
  if (isRechargeAccountOrder(record))
    return '-'
  return enrollTypeMap[record.enrollType] || '-'
}

function getSkuCountText(record) {
  if (isRechargeAccountOrder(record))
    return '1份'
  return record.skuCount ? `${record.skuCount}份` : '-'
}

function getRechargePlaceholderText(record, value) {
  if (isRechargeAccountOrder(record))
    return '-'
  return value || '-'
}

function getRechargeMoneyText(record, value, prefix = '¥ ') {
  if (isRechargeAccountOrder(record) && Number(value || 0) === 0)
    return '-'
  return `${prefix}${formatMoney(value)}`
}

onMounted(() => {
  resetQueryState()
  fetchOrderDetailList()
})

defineExpose({
  refreshData: () => fetchOrderDetailList(),
})
</script>

<template>
  <div>
    <div class="filter-wrap bg-white pl-3 pr-3 rounded-lb-4 rounded-rb-4">
      <all-filter
        ref="allFilterRef"
        :display-array="displayArray"
        :default-create-time-vals="defaultCreateTimeVals"
        :default-order-status-vals="defaultOrderStatusVals"
        :is-quick-show="false"
        :is-show-search-stu-phonefilter="true"
        create-user-label="经办人"
        create-user-placeholder="请输入经办人"
        sales-person-label="订单销售员"
        sales-person-placeholder="请输入订单销售员"
        create-time-label="创建时间"
        v-on="filterUpdateHandlers"
      />
    </div>
    <div class="student-list mt-3 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            当前共 {{ pagination.total }} 笔订单明细
          </div>
          <div class="edit flex">
            <a-dropdown class="mr-2">
              <template #overlay>
                <a-menu>
                  <a-menu-item key="1">导出数据</a-menu-item>
                </a-menu>
              </template>
              <a-button>
                导出数据
                <DownOutlined :style="{ fontSize: '10px' }" />
              </a-button>
            </a-dropdown>
            <customize-code
              v-model:checked-values="selectedValues"
              :options="columnOptions"
              :total="allColumns.length"
              :num="selectedValues.length"
            />
          </div>
        </div>
        <div class="table-content mt-2">
          <a-table
            :data-source="dataSource"
            :pagination="pagination"
            :columns="filteredColumns"
            :scroll="{ x: totalWidth }"
            :sticky="{ offsetHeader: 100 }"
            :loading="loading"
            size="small"
            @change="handleTableChange"
          >
            <template #headerCell="{ column }">
              <template v-if="column.key === 'orderStatus'">
                <span class="mr-1">{{ column.title }}</span>
                <a-popover placement="top" trigger="hover">
                  <template #content>
                    <div class="w-92 text-#666 leading-6">
                      <div style="color:#222;font-weight:500;padding-bottom:8px;margin-bottom:12px;border-bottom:1px solid #f0f0f0;">
                        字段说明
                      </div>
                      <div>待付款：尚未完成付款的订单。</div>
                      <div>待处理：尚未设置授课课程开始时间或教务用品待发货的订单。</div>
                      <div>已完成：已完成一次支付的订单。</div>
                      <div>退费中：尚未完成退费的订单。</div>
                      <div>已退费：已完成退费的订单。</div>
                      <div>已关闭：未完成支付已关闭的订单。</div>
                      <div>已作废：已完成支付已作废的订单。</div>
                      <div>审批中：已触发审批在审核状态中的订单。</div>
                    </div>
                  </template>
                  <InfoCircleOutlined class="cursor-pointer text-#999 hover:text-#1677ff" />
                </a-popover>
              </template>
            </template>
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'orderNumber'">
                <span class="text-#06f inline-flex flex-items-center whitespace-nowrap cursor-pointer" @click="handleOrderDetail(record.orderId)">
                  {{ record.orderNumber }}
                  <a-tooltip v-if="record.isBadDebt">
                    <template #title>订单已设为坏账</template>
                    <span
                      class="w-5 h-5 block text-#333 bg-#DDD font-600 text-3 ml-1 text-center line-height-5 rounded-1"
                    >坏</span>
                  </a-tooltip>
                  <a-tooltip v-else-if="record.arrearAmount > 0 && record.orderStatus !== 4">
                    <template #title>订单欠费未缴清</template>
                    <span
                      class="w-5 h-5 block text-red bg-#FBE7E6 font-600 text-3 ml-1 text-center line-height-5 rounded-1"
                    >欠</span>
                  </a-tooltip>
                </span>
              </template>
              <template v-else-if="column.key === 'studentName'">
                <student-avatar
                  :id="record.studentId"
                  :name="record.studentName || '-'"
                  :avatar-url="record.studentAvatar || defaultStudentAvatar"
                  :phone="record.studentPhone"
                  :show-gender="true"
                  :gender="record.sex === 0 ? '女' : record.sex === 1 ? '男' : '未知'"
                  :show-age="false"
                  default-active-key="0"
                />
              </template>
              <template v-else-if="column.key === 'orderType'">
                <span
                  :class="
                    Number(record.orderType) === 4
                      ? 'text-#ff3333 font-normal'
                      : ''
                  "
                >
                  {{ orderTypeMap[record.orderType] || '-' }}
                </span>
              </template>
              <template v-else-if="column.key === 'orderSource'">
                {{ orderSourceMap[record.orderSource] || '-' }}
              </template>
              <template v-else-if="column.key === 'tagNames'">
                <span v-if="record.tagNames?.length">{{ record.tagNames.map(tag => `【${tag}】`).join('、') }}</span>
                <span v-else>-</span>
              </template>
              <template v-else-if="column.key === 'orderStatus'">
                <div class="flex flex-items-center">
                  <span class="dot" :style="{
                    background:
                      record.orderStatus === 3 ? '#1890ff' :
                        record.orderStatus === 6 ? '#1677ff' :
                          record.orderStatus === 8 ? '#52c41a' :
                            record.orderStatus === 7 ? '#fa8c16' :
                              record.orderStatus === 1 ? '#faad14' :
                                record.orderStatus === 2 ? '#faad14' :
                                  record.orderStatus === 4 ? '#d9d9d9' :
                                    record.orderStatus === 5 ? '#ff4d4f' :
                                      '#d9d9d9'
                  }" />
                  <span>{{ orderStatusMap[record.orderStatus] || '-' }}</span>
                </div>
              </template>
              <template v-else-if="column.key === 'productName'">
                {{ record.productName || '-' }}
              </template>
              <template v-else-if="column.key === 'enrollType'">
                {{ getEnrollTypeText(record) }}
              </template>
              <template v-else-if="column.key === 'productType'">
                {{ productTypeMap[record.productType] || '-' }}
              </template>
              <template v-else-if="column.key === 'productCategoryName'">
                {{ getRechargePlaceholderText(record, record.productCategoryName) }}
              </template>
              <template v-else-if="column.key === 'quoteName'">
                {{ getRechargePlaceholderText(record, record.quoteName) }}
              </template>
              <template v-else-if="column.key === 'skuName'">
                {{ getQuoteDisplay(record) }}
              </template>
              <template v-else-if="column.key === 'skuCount'">
                {{ getSkuCountText(record) }}
              </template>
              <template v-else-if="column.key === 'quantity'">
                {{ getQuantityText(record.quantity, record.skuUnit, record, 'purchase') }}
              </template>
              <template v-else-if="column.key === 'freeQuantity'">
                {{ getQuantityText(record.freeQuantity, record.skuUnit, record, 'gift') }}
              </template>
              <template v-else-if="column.key === 'discountName'">
                -
              </template>
              <template v-else-if="column.key === 'discountNumber'">
                {{ getDiscountText(record) }}
              </template>
              <template v-else-if="column.key === 'shareDiscount'">
                {{ isRechargeAccountOrder(record) && Number(record.shareDiscount || 0) === 0 ? '-' : `- ¥${Number(record.shareDiscount || 0).toFixed(2)}` }}
              </template>
              <template v-else-if="column.key === 'shouldAmount'">
                <span
                  class="font-800"
                  :class="isRefundDisplayOrder(record) ? 'text-#ff3333' : 'text-#222'"
                >
                  {{ isRefundDisplayOrder(record) ? '-' : '+' }}¥ {{ formatMoney(Math.abs(Number(record.shouldAmount) || 0)) }}
                </span>
              </template>
              <template v-else-if="column.key === 'shareCouponAmount'">
                {{ getRechargeMoneyText(record, record.shareCouponAmount, '-¥ ') }}
              </template>
              <template v-else-if="column.key === 'shareRechargeAccountAmount'">
                {{ getRechargeMoneyText(record, record.shareRechargeAccountAmount, '-¥ ') }}
              </template>
              <template v-else-if="column.key === 'shareRechargeAccountGivingAmount'">
                {{ getRechargeMoneyText(record, record.shareRechargeAccountGivingAmount, '-¥ ') }}
              </template>
              <template v-else-if="column.key === 'actualPaidAmount'">
                <span
                  class="font-800"
                  :class="isRefundDisplayOrder(record) ? 'text-#ff3333' : 'text-#222'"
                >
                  {{ isRefundDisplayOrder(record) ? '-' : '+' }}¥ {{ formatMoney(Math.abs(Number(record.actualPaidAmount) || 0)) }}
                </span>
              </template>
              <template v-else-if="column.key === 'arrearAmount'">
                {{ getRechargeMoneyText(record, record.arrearAmount) }}
              </template>
              <template v-else-if="column.key === 'badDebtAmount'">
                {{ getRechargeMoneyText(record, record.badDebtAmount) }}
              </template>
              <template v-else-if="column.key === 'chargeAgainstAmount'">
                {{ getRechargeMoneyText(record, record.chargeAgainstAmount) }}
              </template>
              <template v-else-if="column.key === 'salePersonName'">
                {{ record.salePersonName || '-' }}
              </template>
              <template v-else-if="column.key === 'staffName'">
                {{ record.staffName || '-' }}
              </template>
              <template v-else-if="column.key === 'dealDate'">
                {{ formatDateOnly(record.dealDate) }}
              </template>
              <template v-else-if="column.key === 'createdTime'">
                {{ formatDate(record.createdTime) }}
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
}
</style>
