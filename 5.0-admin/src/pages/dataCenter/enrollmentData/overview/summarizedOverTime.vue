<script setup>
import { BarChartOutlined, LineChartOutlined, QuestionCircleOutlined } from '@ant-design/icons-vue'
import dateFiltering from '../../components/dateFiltering.vue'
import charColumn from '../../components/charColumn.vue'
import charLine from '../../components/charLine.vue'

// 数据简报
const reportList = [
  {
    title: '现有意向学员数', // 标题
    value: 1, // 数值
    briefing: false, // 是否环比、同比
    popover_title: '现有意向学员数', // 弹出框标题
    popover_content: '截止当前时间，系统内的意向学员数', // 弹出框内容
    chain: '', // 环比
    onYear: '', // 同比
  },
  {
    title: '试听邀约数',
    value: 1,
    briefing: true,
    popover_title: '试听邀约数',
    popover_content: '试听记录数 <br /> 时间：试听记录邀约时间',
    chain: '',
    onYear: '',
  },
  {
    title: '试听完成数',
    value: 1,
    briefing: true,
    popover_title: '试听完成数',
    popover_content: '试听记录数 <br /> 时间：试听记录邀约时间 <br /> 试听记录状态：已完成+已转化',
    chain: '',
    onYear: '',
  },
  {
    title: '试听完成率',
    value: 1,
    briefing: true,
    popover_title: '试听完成率',
    popover_content: '试听完成数/试听邀约数 *100%',
    chain: '',
    onYear: '',
  },
  {
    title: '试听转化数',
    value: 1,
    briefing: true,
    popover_title: '试听转化数',
    popover_content: '试听记录数<br />时间：试听记录邀约时间 <br />试听记录状态：已转化',
    chain: '',
    onYear: '',
  },
  {
    title: '试听转化率',
    value: 1,
    briefing: true,
    popover_title: '试听转化率',
    popover_content: '试听转化数/试听完成数 *100%',
    chain: '',
    onYear: '',
  },
  {
    title: '报课人次',
    value: 1,
    briefing: true,
    popover_title: '试听邀约数',
    popover_content: '成单人次（学员购买商品人次数，即订单明细记录单数量）<br />时间：订单经办时间<br /> 条件：订单状态为待处理与已完成状态<br />维度：订单明细商品类型为课程商品',
    chain: '',
    onYear: '',
  },
  {
    title: '报课实收金额',
    value: 8960,
    briefing: true,
    popover_title: '报课实收金额',
    popover_content: '订单实收金额<br /> 时间：订单经办时间<br />条件：订单状态为待处理与已完成状态<br />维度：订单明细商品类型为课程商品',
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
    title: '新增意向学员数',
    value: 0,
    popover_title: '新增意向学员数', // 弹出框标题
    popover_content: '学员数 <br />时间：创建时间', // 弹出框内容
  },
  {
    title: '试听邀约数',
    value: 1,
    popover_title: '试听邀约数',
    popover_content: '试听记录数 <br /> 时间：试听记录邀约时间',
  },
  {
    title: '试听完成数',
    value: 2,
    popover_title: '试听完成数',
    popover_content: '试听记录数 <br /> 时间：试听记录邀约时间 <br /> 试听记录状态：已完成+已转化',
  },
  {
    title: '试听完成率',
    value: 3,
    popover_title: '试听完成率',
    popover_content: '试听完成数/试听邀约数 *100%',
  },
  {
    title: '试听转化数',
    value: 4,
    popover_title: '试听转化数',
    popover_content: '试听记录数<br />时间：试听记录邀约时间 <br />试听记录状态：已转化',
  },
  {
    title: '试听转化率',
    value: 5,
    popover_title: '试听转化率',
    popover_content: '试听转化数/试听完成数 *100%',
  },
  {
    title: '报课人次',
    value: 6,
    popover_title: '试听邀约数',
    popover_content: '成单人次（学员购买商品人次数，即订单明细记录单数量）<br />时间：订单经办时间<br /> 条件：订单状态为待处理与已完成状态<br />维度：订单明细商品类型为课程商品',
  },
  {
    title: '报课实收金额',
    value: 7,
    popover_title: '报课实收金额',
    popover_content: '订单实收金额<br /> 时间：订单经办时间<br />条件：订单状态为待处理与已完成状态<br />维度：订单明细商品类型为课程商品',
  },
]
// 表格列
const allColumns = ref([
  {
    title: '时间',
    dataIndex: 'date',
    key: 'date',
    fixed: 'left',
    width: 160,
    sorter: (a, b) => a.date - b.date,
    required: true, // 新增必选标识
  },
  {
    title: '新增意向学员数',
    dataIndex: 'newIntentStudents',
    key: 'newIntentStudents',
    width: 160,
    sorter: (a, b) => a.newIntentStudents - b.newIntentStudents,
  },
  {
    title: '跟进记录数',
    dataIndex: 'followUpRecords',
    key: 'followUpRecords',
    width: 160,
    sorter: (a, b) => a.followUpRecords - b.followUpRecords,
  },
  {
    title: '未回访跟进记录数',
    dataIndex: 'unvisitedFollowUpRecords',
    key: 'unvisitedFollowUpRecords',
    width: 160,
    sorter: (a, b) => a.unvisitedFollowUpRecords - b.unvisitedFollowUpRecords,
  },
  {
    title: '已回访跟进记录数',
    dataIndex: 'visitedFollowUpRecords',
    key: 'visitedFollowUpRecords',
    width: 160,
    sorter: (a, b) => a.visitedFollowUpRecords - b.visitedFollowUpRecords,
  },
  {
    title: '试听邀约数',
    dataIndex: 'trialInvitations',
    key: 'trialInvitations',
    width: 160,
    sorter: (a, b) => a.trialInvitations - b.trialInvitations,
  },
  {
    title: '试听完成数',
    dataIndex: 'completedTrials',
    key: 'completedTrials',
    width: 160,
    sorter: (a, b) => a.completedTrials - b.completedTrials,
  },
  {
    title: '试听完成率',
    dataIndex: 'trialCompletionRate',
    key: 'trialCompletionRate',
    width: 160,
    sorter: (a, b) => a.trialCompletionRate - b.trialCompletionRate,
  },
  {
    title: '试听转化数',
    dataIndex: 'trialConversions',
    key: 'trialConversions',
    width: 160,
    sorter: (a, b) => a.trialConversions - b.trialConversions,
  },
  {
    title: '试听转化率',
    dataIndex: 'trialConversionRate',
    key: 'trialConversionRate',
    width: 160,
    sorter: (a, b) => a.trialConversionRate - b.trialConversionRate,
  },
  {
    title: '报课人次',
    dataIndex: 'enrollmentCount',
    key: 'enrollmentCount',
    width: 160,
    sorter: (a, b) => a.enrollmentCount - b.enrollmentCount,
  },
  {
    title: '报课应收成交额',
    dataIndex: 'expectedEnrollmentAmount',
    key: 'expectedEnrollmentAmount',
    width: 160,
    sorter: (a, b) => a.expectedEnrollmentAmount - b.expectedEnrollmentAmount,
  },
  {
    title: '报课实收金额',
    dataIndex: 'actualEnrollmentAmount',
    key: 'actualEnrollmentAmount',
    width: 160,
    sorter: (a, b) => a.actualEnrollmentAmount - b.actualEnrollmentAmount,
  },
])
// 表格数据
const dataSource = ref([
  {
    date: '2023-01-01',
    newIntentStudents: 25,
    followUpRecords: 42,
    unvisitedFollowUpRecords: 15,
    visitedFollowUpRecords: 27,
    trialInvitations: 18,
    completedTrials: 15,
    trialCompletionRate: 83.3,
    trialConversions: 10,
    trialConversionRate: 66.7,
    enrollmentCount: 8,
    expectedEnrollmentAmount: 12000,
    actualEnrollmentAmount: 11500,
  },
  {
    date: '2023-01-02',
    newIntentStudents: 30,
    followUpRecords: 38,
    unvisitedFollowUpRecords: 12,
    visitedFollowUpRecords: 26,
    trialInvitations: 22,
    completedTrials: 18,
    trialCompletionRate: 81.8,
    trialConversions: 12,
    trialConversionRate: 66.7,
    enrollmentCount: 10,
    expectedEnrollmentAmount: 15000,
    actualEnrollmentAmount: 14200,
  },
  {
    date: '2023-01-03',
    newIntentStudents: 28,
    followUpRecords: 45,
    unvisitedFollowUpRecords: 18,
    visitedFollowUpRecords: 27,
    trialInvitations: 20,
    completedTrials: 16,
    trialCompletionRate: 80.0,
    trialConversions: 11,
    trialConversionRate: 68.8,
    enrollmentCount: 9,
    expectedEnrollmentAmount: 13500,
    actualEnrollmentAmount: 12800,
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
