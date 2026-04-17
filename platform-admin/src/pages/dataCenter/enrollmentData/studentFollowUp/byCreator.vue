<script setup>
import { BarChartOutlined, LineChartOutlined, QuestionCircleOutlined } from '@ant-design/icons-vue'
import dateFiltering from '../../components/dateFiltering.vue'
import charColumn from '../../components/charColumn.vue'
import charLine from '../../components/charLine.vue'
// 数据简报
const reportList = [
  {
    title: '创建跟进记录员工数', // 标题
    value: 1, // 数值
    briefing: false, // 是否环比、同比
    popover_title: '创建跟进记录员工数', // 弹出框标题
    popover_content: '查询时间内，有创建过跟进记录的员工数', // 弹出框内容
    chain: '', // 环比
    onYear: '', // 同比
  },
  {
    title: '已跟进人数',
    value: 1,
    briefing: true,
    popover_title: '已跟进人数',
    popover_content: '跟进记录数（按学员去重）<br />时间：跟进记录创建时间',
    chain: '',
    onYear: '',
  },
  {
    title: '跟进记录数',
    value: 1,
    briefing: true,
    popover_title: '跟进记录数',
    popover_content: '跟进记录数 <br />时间：跟进记录创建时间',
    chain: '',
    onYear: '',
  },
  {
    title: '未回访跟进记录数',
    value: 1,
    briefing: true,
    popover_title: '未回访跟进记录数',
    popover_content: '跟进记录数 <br />时间：跟进记录创建时间 <br />跟进记录状态：未回访',
    chain: '',
    onYear: '',
  },
  {
    title: '已回访跟进记录数',
    value: 0,
    briefing: true,
    popover_title: '已回访跟进记录数',
    popover_content: '跟进记录数 <br />时间：跟进记录创建时间 <br />跟进记录状态：已回访',
    chain: '',
    onYear: '',
  },
  {
    title: '逾期未回访跟进记录数',
    value: 1,
    briefing: false,
    popover_title: '逾期未回访跟进记录数',
    popover_content: '跟进记录数 <br />时间：截止当前 <br />条件：跟进记录下次跟进时间已超过当前时间',
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
    title: '已跟进人数',
    value: 0,
    popover_title: '已跟进人数',
    popover_content: '已经进行跟进的学员人数',
  },
  {
    title: '跟进记录数',
    value: 1,
    popover_title: '跟进记录数',
    popover_content: '跟进记录数 <br />时间：跟进记录创建时间',
  },
  {
    title: '未回访跟进记录数',
    value: 2,
    popover_title: '未回访跟进记录数',
    popover_content: '跟进记录数 <br />时间：跟进记录创建时间 <br />跟进记录状态：未回访',
  },
  {
    title: '已回访跟进记录数',
    value: 3,
    popover_title: '已回访跟进记录数',
    popover_content: '跟进记录数 <br />时间：跟进记录创建时间 <br />跟进记录状态：已回访',
  },
]
// 表格列
const allColumns = ref([
  {
    title: '创建人姓名',
    dataIndex: 'salesName',
    key: 'salesName',
    fixed: 'left',
    width: 160,
    sorter: (a, b) => a.salesName.localeCompare(b.salesName),
    required: true, // 新增必选标识
  },
  {
    title: '在职状态',
    dataIndex: 'employmentStatus',
    key: 'employmentStatus',
    width: 160,
    sorter: (a, b) => a.employmentStatus.localeCompare(b.employmentStatus),
  },
  {
    title: '入职创建时间',
    dataIndex: 'creationTime',
    key: 'creationTime',
    width: 160,
    sorter: (a, b) => a.creationTime.localeCompare(b.creationTime),
  },
  {
    title: '已跟进人数',
    dataIndex: 'followedPersons',
    key: 'followedPersons',
    width: 160,
    sorter: (a, b) => a.followedPersons - b.followedPersons,
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
    title: '逾期未回访跟进记录数',
    dataIndex: 'overdueUnvisitedRecords',
    key: 'overdueUnvisitedRecords',
    width: 160,
    sorter: (a, b) => a.overdueUnvisitedRecords - b.overdueUnvisitedRecords,
  },
])
// 表格数据
const dataSource = ref([
  {
    salesName: '张三',
    employmentStatus: '在职',
    creationTime: '2023-01-01',
    assignedIntentStudents: 25,
    newIntentStudents: 15,
    assignedEnrolledStudents: 18,
    followedPersons: 22,
    followUpRecords: 42,
    unvisitedFollowUpRecords: 15,
    visitedFollowUpRecords: 27,
    overdueUnvisitedRecords: 5,
  },
  {
    salesName: '李四',
    employmentStatus: '在职',
    creationTime: '2023-01-15',
    assignedIntentStudents: 30,
    newIntentStudents: 20,
    assignedEnrolledStudents: 22,
    followedPersons: 25,
    followUpRecords: 38,
    unvisitedFollowUpRecords: 12,
    visitedFollowUpRecords: 26,
    overdueUnvisitedRecords: 3,
  },
  {
    salesName: '王五',
    employmentStatus: '离职',
    creationTime: '2022-10-01',
    assignedIntentStudents: 28,
    newIntentStudents: 18,
    assignedEnrolledStudents: 20,
    followedPersons: 24,
    followUpRecords: 45,
    unvisitedFollowUpRecords: 18,
    visitedFollowUpRecords: 27,
    overdueUnvisitedRecords: 4,
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
