<script setup>
import { BarChartOutlined, LineChartOutlined, QuestionCircleOutlined } from '@ant-design/icons-vue'
import dateFiltering from '../../components/dateFiltering.vue'
import charColumn from '../../components/charColumn.vue'
import charLine from '../../components/charLine.vue'

// 数据简报
const reportList = [
  {
    title: '当前收款账户数量',
    value: 8,
    briefing: true,
    popover_title: '当前收款账户数量',
    popover_content: '当前收款账户的总数量',
    chain: '',
    onYear: '',
  },
  {
    title: '订单收入',
    value: 6393.3,
    briefing: true,
    popover_title: '订单收入',
    popover_content: '订单收入的总金额',
    chain: '',
    onYear: '',
  },
  {
    title: '订单支出',
    value: 0,
    briefing: true,
    popover_title: '订单支出',
    popover_content: '订单支出的总金额',
    chain: '',
    onYear: '',
  },
  {
    title: '记账收入',
    value: 0,
    briefing: true,
    popover_title: '记账收入',
    popover_content: '记账收入的总金额',
    chain: '',
    onYear: '',
  },
  {
    title: '记账支出',
    value: 0,
    briefing: true,
    popover_title: '记账支出',
    popover_content: '记账支出的总金额',
    chain: '',
    onYear: '',
  },
  {
    title: '总收入',
    value: 6393.3,
    briefing: true,
    popover_title: '总收入',
    popover_content: '订单收入和记账收入的总和',
    chain: '',
    onYear: '',
  },
  {
    title: '总支出',
    value: 0,
    briefing: true,
    popover_title: '总支出',
    popover_content: '订单支出和记账支出的总和',
    chain: '',
    onYear: '',
  },
  {
    title: '总结余',
    value: 6393.3,
    briefing: true,
    popover_title: '总结余',
    popover_content: '总收入减去总支出的金额',
    chain: '',
    onYear: '',
  },
]

// 日期选择组件
const dateFilter = ref(null)
// 是否显示同比、环比
const showComparison = ref(false)
// 图表类型
const activeIcon = ref(0)
// 图表数据类型
const dataType = ref(0)
// 图表数据类型列表
const dataTypeList = [
  {
    title: '订单收入',
    value: 0,
    popover_title: '订单收入',
    popover_content: '订单收入的总金额',
  },
  {
    title: '订单支出',
    value: 1,
    popover_title: '订单支出',
    popover_content: '订单支出的总金额',
  },
  {
    title: '记账收入',
    value: 2,
    popover_title: '记账收入',
    popover_content: '记账收入的总金额',
  },
  {
    title: '记账支出',
    value: 3,
    popover_title: '记账支出',
    popover_content: '记账支出的总金额',
  },
  {
    title: '总收入',
    value: 4,
    popover_title: '总收入',
    popover_content: '订单收入和记账收入的总和',
  },
  {
    title: '总支出',
    value: 5,
    popover_title: '总支出',
    popover_content: '订单支出和记账支出的总和',
  },
]
// 表格列
const allColumns = ref([
  {
    title: '收款方式',
    dataIndex: 'paymentMethod',
    key: 'paymentMethod',
    fixed: 'left',
    width: 160,
    sorter: (a, b) => a.paymentMethod.localeCompare(b.paymentMethod),
    required: true, // 新增必选标识
  },
  {
    title: '收款账户名',
    dataIndex: 'accountName',
    key: 'accountName',
    width: 160,
    sorter: (a, b) => a.accountName.localeCompare(b.accountName),
  },
  {
    title: '订单收入',
    dataIndex: 'orderIncome',
    key: 'orderIncome',
    width: 160,
    sorter: (a, b) => a.orderIncome - b.orderIncome,
  },
  {
    title: '订单支出',
    dataIndex: 'orderExpense',
    key: 'orderExpense',
    width: 160,
    sorter: (a, b) => a.orderExpense - b.orderExpense,
  },
  {
    title: '记账收入',
    dataIndex: 'accountIncome',
    key: 'accountIncome',
    width: 160,
    sorter: (a, b) => a.accountIncome - b.accountIncome,
  },
  {
    title: '记账支出',
    dataIndex: 'accountExpense',
    key: 'accountExpense',
    width: 160,
    sorter: (a, b) => a.accountExpense - b.accountExpense,
  },
  {
    title: '总收入',
    dataIndex: 'totalIncome',
    key: 'totalIncome',
    width: 160,
    sorter: (a, b) => a.totalIncome - b.totalIncome,
  },
  {
    title: '总支出',
    dataIndex: 'totalExpense',
    key: 'totalExpense',
    width: 160,
    sorter: (a, b) => a.totalExpense - b.totalExpense,
  },
  {
    title: '总结余',
    dataIndex: 'totalBalance',
    key: 'totalBalance',
    width: 160,
    sorter: (a, b) => a.totalBalance - b.totalBalance,
  },
  {
    title: '已确认收入',
    dataIndex: 'confirmedIncome',
    key: 'confirmedIncome',
    width: 160,
    sorter: (a, b) => a.confirmedIncome - b.confirmedIncome,
  },
  {
    title: '未确认收入',
    dataIndex: 'unconfirmedIncome',
    key: 'unconfirmedIncome',
    width: 160,
    sorter: (a, b) => a.unconfirmedIncome - b.unconfirmedIncome,
  },
  {
    title: '已确认支出',
    dataIndex: 'confirmedExpense',
    key: 'confirmedExpense',
    width: 160,
    sorter: (a, b) => a.confirmedExpense - b.confirmedExpense,
  },
  {
    title: '未确认支出',
    dataIndex: 'unconfirmedExpense',
    key: 'unconfirmedExpense',
    width: 160,
    sorter: (a, b) => a.unconfirmedExpense - b.unconfirmedExpense,
  },
])

// 表格数据
const dataSource = ref([
  {
    paymentMethod: '微信支付',
    accountName: '微信商户号',
    orderIncome: 11500,
    orderExpense: 0,
    accountIncome: 0,
    accountExpense: 0,
    totalIncome: 11500,
    totalExpense: 0,
    totalBalance: 11500,
    confirmedIncome: 11500,
    unconfirmedIncome: 0,
    confirmedExpense: 0,
    unconfirmedExpense: 0,
  },
  {
    paymentMethod: '支付宝',
    accountName: '支付宝商户号',
    orderIncome: 14200,
    orderExpense: 0,
    accountIncome: 0,
    accountExpense: 0,
    totalIncome: 14200,
    totalExpense: 0,
    totalBalance: 14200,
    confirmedIncome: 14200,
    unconfirmedIncome: 0,
    confirmedExpense: 0,
    unconfirmedExpense: 0,
  },
  {
    paymentMethod: '现金',
    accountName: '现金账户',
    orderIncome: 12800,
    orderExpense: 0,
    accountIncome: 0,
    accountExpense: 0,
    totalIncome: 12800,
    totalExpense: 0,
    totalBalance: 12800,
    confirmedIncome: 12800,
    unconfirmedIncome: 0,
    confirmedExpense: 0,
    unconfirmedExpense: 0,
  },
])
// 图表数据源
const { selectedValues, columnOptions, filteredColumns, totalWidth }
  = useTableColumns({
    storageKey: 'enrollmentDataOverviewSummarizedOverTime', // 本地存储键名
    allColumns, // 原始列配置
    excludeKeys: [], // 需要排除的列键
  })
// 日期选择
function handleDateChange({ date }) {
  dateFilter.value = date
  console.log(dateFilter.value)
}
</script>

<template>
  <div>
    <!-- 日期选择组件 -->
    <date-filtering @change="handleDateChange" />
    <!-- 数据简报 -->
    <div class="card-white">
      <div class="flex justify-between ml-2">
        <div class="total font-bold">
          数据简报
        </div>
        <div class="flex items-center gap-2">
          <span class="font-size-14px">计算同比、环比：</span>
          <a-switch v-model:checked="showComparison" />
        </div>
      </div>
      <div class="flex align-center p-12px gap-24px">
        <div v-for="(item, index) in reportList" :key="index" class="flex-1 p-12px bg-#fbfcff border-radius-12px">
          <div class="block_top flex align-center gap-1">
            <span class="text-#888 font-size-12px whitespace-nowrap">{{ item.title }}</span>
            <a-popover color="#fff" :title="item.popover_title">
              <template #content>
                <div v-html="item.popover_content" />
              </template>
              <QuestionCircleOutlined class="text-#888 font-size-12px" />
            </a-popover>
          </div>
          <div class="block_bottom">
            <div class="font-size-24px font-bold" style="font-family: 'DIN Alternate'">
              {{ item.value }}
            </div>
          </div>
          <!-- 分割线 -->
          <template v-if="item.briefing && showComparison">
            <div class="line" />
            <div class="font-size-12px text-#111">
              <div class="flex align-center gap-2">
                <span>环比</span>
                <span>{{ item.chain || '-' }}</span>
              </div>
              <div class="flex align-center gap-2">
                <span>同比</span>
                <span>{{ item.onYear || '-' }}</span>
              </div>
            </div>
          </template>
        </div>
      </div>
    </div>
    <!-- 趋势分析 -->
    <div class="card-white">
      <div class="flex justify-between ml-2">
        <div class="total font-bold">
          趋势分析
        </div>
        <div class="flex items-center gap-4">
          <a-tooltip>
            <template #title>
              柱状图
            </template>
            <BarChartOutlined
              class="font-size-16px" :class="activeIcon === 0 ? 'text-#1890ff' : 'text-#888'"
              @click="activeIcon = 0"
            />
          </a-tooltip>
          <a-tooltip>
            <template #title>
              趋势图
            </template>
            <LineChartOutlined
              class="font-size-16px" :class="activeIcon === 1 ? 'text-#1890ff' : 'text-#888'"
              @click="activeIcon = 1"
            />
          </a-tooltip>
        </div>
      </div>
      <div class="py-12px ">
        <div class="flex items-center ml-6">
          <div>数据类型：</div>
          <a-radio-group v-model:value="dataType" class="custom-radio">
            <a-radio v-for="(item, index) in dataTypeList" :key="index" :value="index">
              {{ item.title }}
              <a-popover color="#fff" :title="item.popover_title">
                <template #content>
                  <div v-html="item.popover_content" />
                </template>
                <QuestionCircleOutlined class="text-#888 font-size-12px" />
              </a-popover>
            </a-radio>
          </a-radio-group>
        </div>
        <div class="px-44px py-24px">
          <charColumn v-if="activeIcon === 0" />
          <charLine v-if="activeIcon === 1" />
        </div>
      </div>
    </div>
    <!-- 数据说明 -->
    <div class="card-white" style="padding-left: 24px;">
      <custom-title title="数据明细" font-size="15px" font-weight="500" />
      <div class="flex justify-between align-center py-8px">
        <span>共计1条数据</span>
        <a-button ghost type="primary">
          下载报表
        </a-button>
      </div>
      <div class="table-content mt-2">
        <a-table
          :data-source="dataSource" :pagination="false" :columns="filteredColumns"
          :scroll="{ x: totalWidth }" size="small"
        >
          <template #summary>
            <a-table-summary fixed>
              <a-table-summary-row class="summary-row">
                <a-table-summary-cell index="0">
                  总计
                </a-table-summary-cell>
                <a-table-summary-cell v-for="(col, index) in allColumns.slice(1)" :key="index" :index="index + 1">
                  {{ dataSource.reduce((sum, record) => sum + (Number(record[col.dataIndex]) || 0), 0) }}
                </a-table-summary-cell>
              </a-table-summary-row>
            </a-table-summary>
          </template>
        </a-table>
      </div>
    </div>
  </div>
</template>

<style lang="less" scoped>
.card-white {
  background: #fff;
  margin-top: 8px;
  padding: 12px;
  border-radius: 12px;

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
}

.line {
  background-color: #eee;
  height: 1px;
  margin: 8px 0;
}

:deep(.summary-row) {
  background-color: #f0f7ff;
  font-weight: bold;

  td,
  th {
    background-color: #f0f7ff !important;
  }
}

/* 自定义镂空样式 */
.custom-radio ::v-deep(.ant-radio-wrapper:hover .ant-radio),
.custom-radio ::v-deep(.ant-radio:hover .ant-radio-inner),
.custom-radio ::v-deep(.ant-radio-input:focus + .ant-radio-inner) {
  border-color: var(--pro-ant-color-primary);
}

.custom-radio ::v-deep(.ant-radio-inner) {
  background-color: transparent;
  border-color: #d9d9d9;
}

.custom-radio ::v-deep(.ant-radio-checked .ant-radio-inner) {
  background-color: transparent;
  border-color: var(--pro-ant-color-primary);
}

.custom-radio ::v-deep(.ant-radio-inner::after) {
  background-color: var(--pro-ant-color-primary);
  transform: scale(0.5);
}
</style>
