<script setup>
import { BarChartOutlined, LineChartOutlined, QuestionCircleOutlined } from '@ant-design/icons-vue'
import dateFiltering from '../../components/dateFiltering.vue'
import charColumn from '../../components/charColumn.vue'
import charLine from '../../components/charLine.vue'

// 数据简报
const reportList = [
  {
    title: '计划课时课消', // 标题
    value: 22, // 数值
    briefing: false, // 是否环比、同比
    popover_title: '计划课时课消', // 弹出框标题
    popover_content: '计划消耗的课时数量', // 弹出框内容
    chain: '', // 环比
    onYear: '', // 同比
  },
  {
    title: '消耗课时数量',
    value: 2,
    briefing: true,
    popover_title: '消耗课时数量',
    popover_content: '实际消耗的课时数量',
    chain: '',
    onYear: '',
  },
  {
    title: '计划达成率',
    value: '9.09%',
    briefing: true,
    popover_title: '计划达成率',
    popover_content: '消耗课时数量/计划课时课消 *100%',
    chain: '',
    onYear: '',
  },
  {
    title: '消耗金额数量',
    value: 0,
    briefing: true,
    popover_title: '消耗金额数量',
    popover_content: '实际消耗的金额数量',
    chain: '',
    onYear: '',
  },
  {
    title: '拖欠课时数',
    value: 0,
    briefing: true,
    popover_title: '拖欠课时数',
    popover_content: '拖欠的课时数量',
    chain: '',
    onYear: '',
  },
  {
    title: '拖欠金额数',
    value: 0,
    briefing: true,
    popover_title: '拖欠金额数',
    popover_content: '拖欠的金额数量',
    chain: '',
    onYear: '',
  },
  {
    title: '消耗学费金额',
    value: 300,
    briefing: true,
    popover_title: '消耗学费金额',
    popover_content: '消耗的学费金额',
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
    title: '消耗课时数量',
    value: 0,
    popover_title: '消耗课时数量', // 弹出框标题
    popover_content: '实际消耗的课时数量', // 弹出框内容
  },
  {
    title: '到课消耗课时数',
    value: 1,
    popover_title: '到课消耗课时数',
    popover_content: '学员到课消耗的课时数量',
  },
  {
    title: '请假消耗课时数',
    value: 2,
    popover_title: '请假消耗课时数',
    popover_content: '学员请假消耗的课时数量',
  },
  {
    title: '旷课消耗课时数',
    value: 3,
    popover_title: '旷课消耗课时数',
    popover_content: '学员旷课消耗的课时数量',
  },
  {
    title: '消耗金额数量',
    value: 4,
    popover_title: '消耗金额数量',
    popover_content: '实际消耗的金额数量',
  },
  {
    title: '消耗学费金额',
    value: 5,
    popover_title: '消耗学费金额',
    popover_content: '消耗的学费金额',
  },
  {
    title: '拖欠课时数',
    value: 6,
    popover_title: '拖欠课时数',
    popover_content: '拖欠的课时数量',
  },
  {
    title: '拖欠金额数',
    value: 7,
    popover_title: '拖欠金额数',
    popover_content: '拖欠的金额数量',
  },
]
// 表格列
// 表格列
const allColumns = ref([
  {
    title: '时间',
    dataIndex: 'date',
    key: 'date',
    fixed: 'left',
    width: 160,
    sorter: (a, b) => a.date.localeCompare(b.date),
    required: true, // 新增必选标识
  },
  {
    title: '开课次数',
    dataIndex: 'classCount',
    key: 'classCount',
    width: 160,
    sorter: (a, b) => a.classCount - b.classCount,
  },
  {
    title: '上课点名课时数',
    dataIndex: 'attendanceHours',
    key: 'attendanceHours',
    width: 160,
    sorter: (a, b) => a.attendanceHours - b.attendanceHours,
  },
  {
    title: '上课点名金额数',
    dataIndex: 'attendanceAmount',
    key: 'attendanceAmount',
    width: 160,
    sorter: (a, b) => a.attendanceAmount - b.attendanceAmount,
  },
  {
    title: '消耗课时数量',
    dataIndex: 'consumedHours',
    key: 'consumedHours',
    width: 160,
    sorter: (a, b) => a.consumedHours - b.consumedHours,
  },
  {
    title: '到课消耗课时数量',
    dataIndex: 'attendedConsumedHours',
    key: 'attendedConsumedHours',
    width: 160,
    sorter: (a, b) => a.attendedConsumedHours - b.attendedConsumedHours,
  },
  {
    title: '请假消耗课时数量',
    dataIndex: 'leaveConsumedHours',
    key: 'leaveConsumedHours',
    width: 160,
    sorter: (a, b) => a.leaveConsumedHours - b.leaveConsumedHours,
  },
  {
    title: '旷课消耗课时数量',
    dataIndex: 'absentConsumedHours',
    key: 'absentConsumedHours',
    width: 160,
    sorter: (a, b) => a.absentConsumedHours - b.absentConsumedHours,
  },
  {
    title: '消耗金额数量',
    dataIndex: 'consumedAmount',
    key: 'consumedAmount',
    width: 160,
    sorter: (a, b) => a.consumedAmount - b.consumedAmount,
  },
  {
    title: '拖欠课时数',
    dataIndex: 'arrearsHours',
    key: 'arrearsHours',
    width: 160,
    sorter: (a, b) => a.arrearsHours - b.arrearsHours,
  },
  {
    title: '拖欠金额数',
    dataIndex: 'arrearsAmount',
    key: 'arrearsAmount',
    width: 160,
    sorter: (a, b) => a.arrearsAmount - b.arrearsAmount,
  },
  {
    title: '消耗学费金额',
    dataIndex: 'consumedTuitionAmount',
    key: 'consumedTuitionAmount',
    width: 160,
    sorter: (a, b) => a.consumedTuitionAmount - b.consumedTuitionAmount,
  },
])
// 表格数据
const dataSource = ref([
  {
    date: '2023-01-01',
    classCount: 12,
    attendanceHours: 45,
    attendanceAmount: 4500,
    consumedHours: 40,
    attendedConsumedHours: 35,
    leaveConsumedHours: 3,
    absentConsumedHours: 2,
    consumedAmount: 4000,
    arrearsHours: 5,
    arrearsAmount: 500,
    consumedTuitionAmount: 3800,
  },
  {
    date: '2023-01-02',
    classCount: 15,
    attendanceHours: 50,
    attendanceAmount: 5000,
    consumedHours: 48,
    attendedConsumedHours: 42,
    leaveConsumedHours: 4,
    absentConsumedHours: 2,
    consumedAmount: 4800,
    arrearsHours: 2,
    arrearsAmount: 200,
    consumedTuitionAmount: 4600,
  },
  {
    date: '2023-01-03',
    classCount: 14,
    attendanceHours: 48,
    attendanceAmount: 4800,
    consumedHours: 45,
    attendedConsumedHours: 40,
    leaveConsumedHours: 3,
    absentConsumedHours: 2,
    consumedAmount: 4500,
    arrearsHours: 3,
    arrearsAmount: 300,
    consumedTuitionAmount: 4300,
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
