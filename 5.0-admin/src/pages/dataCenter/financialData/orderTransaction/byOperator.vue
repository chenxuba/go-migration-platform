<script setup>
import { BarChartOutlined, LineChartOutlined, QuestionCircleOutlined } from '@ant-design/icons-vue'
import dateFiltering from '../../components/dateFiltering.vue'
import charColumn from '../../components/charColumn.vue'
import charLine from '../../components/charLine.vue'

// 数据简报
const reportList = [
  {
    title: '经办订单员工数',
    value: 2,
    briefing: true,
    popover_title: '经办订单员工数',
    popover_content: '经办订单的员工总数',
    chain: '',
    onYear: '',
  },
  {
    title: '订单应收',
    value: 6393.3,
    briefing: true,
    popover_title: '订单应收',
    popover_content: '订单的应收金额总数',
    chain: '',
    onYear: '',
  },
  {
    title: '订单实收',
    value: 6393.3,
    briefing: true,
    popover_title: '订单实收',
    popover_content: '订单的实收金额总数',
    chain: '',
    onYear: '',
  },
  {
    title: '订单欠费',
    value: 0,
    briefing: true,
    popover_title: '订单欠费',
    popover_content: '订单的欠费金额总数',
    chain: '',
    onYear: '',
  },
  {
    title: '订单支出',
    value: 0,
    briefing: true,
    popover_title: '订单支出',
    popover_content: '订单的支出金额总数',
    chain: '',
    onYear: '',
  },
  {
    title: '订单结余',
    value: 6393.3,
    briefing: true,
    popover_title: '订单结余',
    popover_content: '订单的结余金额总数',
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
const dataTypeList = [
  {
    title: '经办订单员工数',
    value: 0,
    popover_title: '经办订单员工数',
    popover_content: '经办订单的员工总数',
  },
  {
    title: '订单应收',
    value: 1,
    popover_title: '订单应收',
    popover_content: '订单的应收金额总数',
  },
  {
    title: '订单实收',
    value: 2,
    popover_title: '订单实收',
    popover_content: '订单的实收金额总数',
  },
  {
    title: '订单欠费',
    value: 3,
    popover_title: '订单欠费',
    popover_content: '订单的欠费金额总数',
  },
  {
    title: '订单支出',
    value: 4,
    popover_title: '订单支出',
    popover_content: '订单的支出金额总数',
  },
  {
    title: '订单结余',
    value: 5,
    popover_title: '订单结余',
    popover_content: '订单的结余金额总数',
  },
]
// 表格列
const allColumns = ref([
  {
    title: '经办人姓名',
    dataIndex: 'operatorName',
    key: 'operatorName',
    fixed: 'left',
    width: 160,
    sorter: (a, b) => a.operatorName.localeCompare(b.operatorName),
    required: true,
  },
  {
    title: '在职状态',
    dataIndex: 'employmentStatus',
    key: 'employmentStatus',
    width: 120,
    sorter: (a, b) => a.employmentStatus.localeCompare(b.employmentStatus),
  },
  {
    title: '入职创建时间',
    dataIndex: 'hireDate',
    key: 'hireDate',
    width: 160,
    sorter: (a, b) => a.hireDate.localeCompare(b.hireDate),
  },
  {
    title: '订单应收',
    dataIndex: 'orderReceivable',
    key: 'orderReceivable',
    width: 160,
    sorter: (a, b) => a.orderReceivable - b.orderReceivable,
  },
  {
    title: '订单实收',
    dataIndex: 'orderActualReceived',
    key: 'orderActualReceived',
    width: 160,
    sorter: (a, b) => a.orderActualReceived - b.orderActualReceived,
  },
  {
    title: '订单欠费',
    dataIndex: 'orderArrears',
    key: 'orderArrears',
    width: 160,
    sorter: (a, b) => a.orderArrears - b.orderArrears,
  },
  {
    title: '订单支出',
    dataIndex: 'orderExpenditure',
    key: 'orderExpenditure',
    width: 160,
    sorter: (a, b) => a.orderExpenditure - b.orderExpenditure,
  },
  {
    title: '订单结余',
    dataIndex: 'orderBalance',
    key: 'orderBalance',
    width: 160,
    sorter: (a, b) => a.orderBalance - b.orderBalance,
  },
  {
    title: '报课结余',
    dataIndex: 'courseBalance',
    key: 'courseBalance',
    width: 160,
    sorter: (a, b) => a.courseBalance - b.courseBalance,
  },
  {
    title: '学杂费结余',
    dataIndex: 'miscBalance',
    key: 'miscBalance',
    width: 160,
    sorter: (a, b) => a.miscBalance - b.miscBalance,
  },
  {
    title: '教学用品结余',
    dataIndex: 'teachingMaterialBalance',
    key: 'teachingMaterialBalance',
    width: 160,
    sorter: (a, b) => a.teachingMaterialBalance - b.teachingMaterialBalance,
  },
  {
    title: '储值账户结余',
    dataIndex: 'storedValueBalance',
    key: 'storedValueBalance',
    width: 160,
    sorter: (a, b) => a.storedValueBalance - b.storedValueBalance,
  },
  {
    title: '场地预约结余',
    dataIndex: 'venueBookingBalance',
    key: 'venueBookingBalance',
    width: 160,
    sorter: (a, b) => a.venueBookingBalance - b.venueBookingBalance,
  },
])

// 表格数据
const dataSource = ref([
  {
    operatorName: '张三',
    employmentStatus: '在职',
    hireDate: '2023-01-01',
    orderReceivable: 6000,
    orderActualReceived: 5800,
    orderArrears: 200,
    orderExpenditure: 1000,
    orderBalance: 4800,
    courseBalance: 3000,
    miscBalance: 500,
    teachingMaterialBalance: 300,
    storedValueBalance: 800,
    venueBookingBalance: 200,
  },
  {
    operatorName: '李四',
    employmentStatus: '在职',
    hireDate: '2023-02-01',
    orderReceivable: 7000,
    orderActualReceived: 7000,
    orderArrears: 0,
    orderExpenditure: 1200,
    orderBalance: 5800,
    courseBalance: 3500,
    miscBalance: 600,
    teachingMaterialBalance: 400,
    storedValueBalance: 1000,
    venueBookingBalance: 300,
  },
  {
    operatorName: '王五',
    employmentStatus: '离职',
    hireDate: '2023-03-01',
    orderReceivable: 5000,
    orderActualReceived: 4800,
    orderArrears: 200,
    orderExpenditure: 800,
    orderBalance: 4000,
    courseBalance: 2500,
    miscBalance: 400,
    teachingMaterialBalance: 300,
    storedValueBalance: 600,
    venueBookingBalance: 200,
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
