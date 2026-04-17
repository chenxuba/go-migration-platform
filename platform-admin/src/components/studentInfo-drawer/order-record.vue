<script setup>
import { computed, ref, watch } from 'vue'
import dayjs from 'dayjs'
import OrderDetailDrawer from '@/components/common/order-detail-drawer.vue'
import { getOrderListApi } from '@/api/finance-center/order-manage'
import { useStudentStore } from '@/stores/student'
import messageService from '@/utils/messageService'

const studentStore = useStudentStore()

const openOrderDetailDrawer = ref(false)
const currentOrderId = ref('')
const loading = ref(false)
const dataSource = ref([])
const pagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
})

const allColumns = ref([
  {
    title: '订单编号',
    dataIndex: 'orderNumber',
    key: 'orderNumber',
    fixed: 'left',
    width: 210,
  },
  {
    title: '订单类型',
    key: 'orderType',
    dataIndex: 'orderType',
    width: 120,
  },
  {
    title: '订单来源',
    key: 'orderSource',
    dataIndex: 'orderSource',
    width: 120,
  },
  {
    title: '订单状态',
    key: 'orderStatus',
    dataIndex: 'orderStatus',
    width: 120,
  },
  {
    title: '办理内容',
    dataIndex: 'handleContent',
    key: 'handleContent',
    width: 220,
  },
  {
    title: '订单创建时间',
    dataIndex: 'createdTime',
    key: 'createdTime',
    width: 180,
  },
  {
    title: '订单总金额(元)',
    dataIndex: 'orderTotalPrice',
    key: 'orderTotalPrice',
    width: 160,
  },
])

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

const refundOrderTypes = new Set([3, 4, 6, 7])

const totalText = computed(() => `共${pagination.value.total}条订单`)

function formatDate(dateStr) {
  if (!dateStr)
    return '-'
  return dayjs(dateStr).format('YYYY-MM-DD HH:mm')
}

function formatMoney(amount) {
  const value = Number(amount || 0)
  if (!Number.isFinite(value))
    return '0.00'
  return value.toFixed(2)
}

function orderProductCount(record) {
  return Array.isArray(record?.productItems) ? record.productItems.length : 0
}

function orderHandleContent(record) {
  const list = Array.isArray(record?.productItems) ? record.productItems.filter(Boolean) : []
  return list.length ? list.join('、') : '-'
}

function orderAmountText(record) {
  const amount = Math.abs(Number(record?.totalAmount ?? record?.amount ?? 0))
  const sign = refundOrderTypes.has(Number(record?.orderType || 0)) ? '-' : '+'
  return `${sign}${formatMoney(amount)}`
}

function shouldShowArrearBadge(record) {
  return !record?.isBadDebt && Number(record?.orderStatus || 0) !== 4 && Number(record?.arrearAmount || 0) > 0
}

function handleOrderDetail(orderId) {
  const id = String(orderId || '').trim()
  if (!id)
    return
  currentOrderId.value = id
  openOrderDetailDrawer.value = true
}

async function fetchOrderList() {
  const studentId = String(studentStore.studentId || '').trim()
  if (!studentId) {
    dataSource.value = []
    pagination.value.total = 0
    return
  }
  loading.value = true
  try {
    const { result } = await getOrderListApi({
      sortModel: {},
      queryModel: {
        studentId,
      },
      pageRequestModel: {
        needTotal: true,
        pageSize: pagination.value.pageSize,
        pageIndex: pagination.value.current,
      },
    })
    const list = Array.isArray(result?.list) ? result.list : []
    dataSource.value = list
    pagination.value.total = Number(result?.total || 0)
  }
  catch (error) {
    console.error('获取订单记录失败:', error)
    messageService.error('获取订单记录失败')
    dataSource.value = []
    pagination.value.total = 0
  }
  finally {
    loading.value = false
  }
}

function handleTableChange(page) {
  pagination.value.current = Number(page?.current || 1)
  pagination.value.pageSize = Number(page?.pageSize || 10)
  fetchOrderList()
}

watch(
  () => String(studentStore.studentId || '').trim(),
  () => {
    pagination.value.current = 1
    fetchOrderList()
  },
  { immediate: true },
)
</script>

<template>
  <div class="order-record p3 pr0">
    <div class="record bg-white rounded-3 p3">
      <div class="total mb-1.5">
        {{ totalText }}
      </div>
      <a-table
        :sticky="true"
        :loading="loading"
        :data-source="dataSource"
        :columns="allColumns"
        :pagination="pagination.total > pagination.pageSize ? pagination : false"
        row-key="orderId"
        style="color: #666;"
        size="small"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'orderNumber'">
            <span class="text-#06f flex-center justify-start cursor-pointer" @click="handleOrderDetail(record.orderId)">
              {{ record.orderNumber || '-' }}
              <a-tooltip v-if="shouldShowArrearBadge(record)">
                <template #title>订单欠费未缴清</template>
                <span class="w-5 h-5 block text-red bg-#FBE7E6 text-3 ml-1 text-center line-height-5 rounded-1">欠</span>
              </a-tooltip>
            </span>
          </template>
          <template v-else-if="column.key === 'orderType'">
            {{ orderTypeMap[record.orderType] || '-' }}
          </template>
          <template v-else-if="column.key === 'orderSource'">
            {{ orderSourceMap[record.orderSource] || '-' }}
          </template>
          <template v-else-if="column.key === 'orderStatus'">
            <div class="flex-center justify-start">
              <span class="dot" /><span>{{ orderStatusMap[record.orderStatus] || '-' }}</span>
            </div>
          </template>
          <template v-else-if="column.key === 'handleContent'">
            {{ orderHandleContent(record) }}
          </template>
          <template v-else-if="column.key === 'createdTime'">
            {{ formatDate(record.createdTime) }}
          </template>
          <template v-else-if="column.key === 'orderTotalPrice'">
            <div class="text-center">
              <div class="text-#222 font-500">
                {{ orderAmountText(record) }}
              </div>
              <div class="text-#888">
                共{{ orderProductCount(record) }}件商品
              </div>
            </div>
          </template>
        </template>
      </a-table>
    </div>
    <OrderDetailDrawer v-model:open="openOrderDetailDrawer" :order-id="currentOrderId" />
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
  background: #06f;
}
</style>
