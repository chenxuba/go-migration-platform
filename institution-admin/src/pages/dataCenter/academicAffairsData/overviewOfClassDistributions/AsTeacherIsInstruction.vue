<script setup>
import { BarChartOutlined, LineChartOutlined, QuestionCircleOutlined } from '@ant-design/icons-vue'
import dateFiltering from '../../components/dateFiltering.vue'
import charColumn from '../../components/charColumn.vue'
import charLine from '../../components/charLine.vue'

// 数据简报
const reportList = [
  {
    title: '上课老师数量', // 标题
    value: 2, // 数值
    briefing: false, // 是否环比、同比
    popover_title: '上课老师数量', // 弹出框标题
    popover_content: '所选时间段内上课的老师总数', // 弹出框内容
    chain: '', // 环比
    onYear: '', // 同比
  },
  {
    title: '上课助教数量',
    value: 0,
    briefing: true,
    popover_title: '上课助教数量',
    popover_content: '所选时间段内上课的助教总数',
    chain: '',
    onYear: '',
  },
  {
    title: '课消课程数量',
    value: 2,
    briefing: true,
    popover_title: '课消课程数量',
    popover_content: '所选时间段内消耗课时的课程数量',
    chain: '',
    onYear: '',
  },
  {
    title: '消耗课时数量',
    value: 2,
    briefing: true,
    popover_title: '消耗课时数量',
    popover_content: '所选时间段内消耗的课时总数',
    chain: '',
    onYear: '',
  },
  {
    title: '消耗金额数量',
    value: 0,
    briefing: true,
    popover_title: '消耗金额数量',
    popover_content: '所选时间段内消耗的金额总数',
    chain: '',
    onYear: '',
  },
  {
    title: '拖欠课时数',
    value: 0,
    briefing: true,
    popover_title: '拖欠课时数',
    popover_content: '所选时间段内拖欠的课时总数',
    chain: '',
    onYear: '',
  },
  {
    title: '拖欠金额数',
    value: 0,
    briefing: true,
    popover_title: '拖欠金额数',
    popover_content: '所选时间段内拖欠的金额总数',
    chain: '',
    onYear: '',
  },
  {
    title: '消耗学费金额',
    value: 300,
    briefing: true,
    popover_title: '消耗学费金额',
    popover_content: '所选时间段内消耗的学费金额总数',
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
    title: '消耗课时数量',
    value: 0,
    popover_title: '消耗课时数量',
    popover_content: '所选时间段内消耗的课时总数',
  },
  {
    title: '到课消耗课时数',
    value: 1,
    popover_title: '到课消耗课时数',
    popover_content: '所选时间段内学员到课消耗的课时数量',
  },
  {
    title: '请假消耗课时数',
    value: 2,
    popover_title: '请假消耗课时数',
    popover_content: '所选时间段内学员请假消耗的课时数量',
  },
  {
    title: '旷课消耗课时数',
    value: 3,
    popover_title: '旷课消耗课时数',
    popover_content: '所选时间段内学员旷课消耗的课时数量',
  },
  {
    title: '消耗金额数量',
    value: 4,
    popover_title: '消耗金额数量',
    popover_content: '所选时间段内消耗的金额总数',
  },
  {
    title: '消耗学费金额',
    value: 5,
    popover_title: '消耗学费金额',
    popover_content: '所选时间段内消耗的学费金额总数',
  },
]

// 表格列
const allColumns = ref([
  {
    title: '老师姓名',
    dataIndex: 'teacherName',
    key: 'teacherName',
    fixed: 'left',
    width: 160,
    sorter: (a, b) => a.teacherName.localeCompare(b.teacherName),
    required: true, // 新增必选标识
  },
  {
    title: '课程名称',
    dataIndex: 'courseName',
    key: 'courseName',
    width: 160,
    sorter: (a, b) => a.courseName.localeCompare(b.courseName),
  },
  {
    title: '开课次数',
    dataIndex: 'classCount',
    key: 'classCount',
    width: 160,
    sorter: (a, b) => a.classCount - b.classCount,
  },
  {
    title: '老师授课课时',
    dataIndex: 'teachingHours',
    key: 'teachingHours',
    width: 160,
    sorter: (a, b) => a.teachingHours - b.teachingHours,
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
    title: '计划课时课消',
    dataIndex: 'plannedHours',
    key: 'plannedHours',
    width: 160,
    sorter: (a, b) => a.plannedHours - b.plannedHours,
  },
  {
    title: '消耗课时数量',
    dataIndex: 'consumedHours',
    key: 'consumedHours',
    width: 160,
    sorter: (a, b) => a.consumedHours - b.consumedHours,
  },
  {
    title: '到课消耗课时数',
    dataIndex: 'attendedConsumedHours',
    key: 'attendedConsumedHours',
    width: 160,
    sorter: (a, b) => a.attendedConsumedHours - b.attendedConsumedHours,
  },
  {
    title: '请假消耗课时数',
    dataIndex: 'leaveConsumedHours',
    key: 'leaveConsumedHours',
    width: 160,
    sorter: (a, b) => a.leaveConsumedHours - b.leaveConsumedHours,
  },
  {
    title: '旷课消耗课时数',
    dataIndex: 'absentConsumedHours',
    key: 'absentConsumedHours',
    width: 160,
    sorter: (a, b) => a.absentConsumedHours - b.absentConsumedHours,
  },
  {
    title: '计划达成率',
    dataIndex: 'planCompletionRate',
    key: 'planCompletionRate',
    width: 160,
    sorter: (a, b) => a.planCompletionRate - b.planCompletionRate,
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
    title: '消耗课时单价',
    dataIndex: 'hourPrice',
    key: 'hourPrice',
    width: 160,
    sorter: (a, b) => a.hourPrice - b.hourPrice,
  },
  {
    title: '消耗课时学费',
    dataIndex: 'hourTuition',
    key: 'hourTuition',
    width: 160,
    sorter: (a, b) => a.hourTuition - b.hourTuition,
  },
  {
    title: '消耗学费金额',
    dataIndex: 'consumedTuition',
    key: 'consumedTuition',
    width: 160,
    sorter: (a, b) => a.consumedTuition - b.consumedTuition,
  },
  {
    title: '补课课时数',
    dataIndex: 'makeupHours',
    key: 'makeupHours',
    width: 160,
    sorter: (a, b) => a.makeupHours - b.makeupHours,
  },
  {
    title: '补课金额数',
    dataIndex: 'makeupAmount',
    key: 'makeupAmount',
    width: 160,
    sorter: (a, b) => a.makeupAmount - b.makeupAmount,
  },
  {
    title: '补课消耗学费金额',
    dataIndex: 'makeupTuition',
    key: 'makeupTuition',
    width: 160,
    sorter: (a, b) => a.makeupTuition - b.makeupTuition,
  },
])
// 表格数据
const dataSource = ref([
  {
    teacherName: '张老师',
    courseName: '初级英语',
    classCount: 12,
    teachingHours: 24,
    attendanceHours: 36,
    attendanceAmount: 3600,
    plannedHours: 40,
    consumedHours: 36,
    attendedConsumedHours: 30,
    leaveConsumedHours: 4,
    absentConsumedHours: 2,
    planCompletionRate: 90,
    consumedAmount: 3600,
    arrearsHours: 0,
    arrearsAmount: 0,
    hourPrice: 100,
    hourTuition: 100,
    consumedTuition: 3600,
    makeupHours: 5,
    makeupAmount: 500,
    makeupTuition: 500,
  },
  {
    teacherName: '李老师',
    courseName: '中级英语',
    classCount: 15,
    teachingHours: 30,
    attendanceHours: 45,
    attendanceAmount: 4500,
    plannedHours: 50,
    consumedHours: 45,
    attendedConsumedHours: 40,
    leaveConsumedHours: 3,
    absentConsumedHours: 2,
    planCompletionRate: 90,
    consumedAmount: 4500,
    arrearsHours: 0,
    arrearsAmount: 0,
    hourPrice: 100,
    hourTuition: 100,
    consumedTuition: 4500,
    makeupHours: 4,
    makeupAmount: 400,
    makeupTuition: 400,
  },
  {
    teacherName: '王老师',
    courseName: '高级英语',
    classCount: 14,
    teachingHours: 28,
    attendanceHours: 42,
    attendanceAmount: 4200,
    plannedHours: 45,
    consumedHours: 42,
    attendedConsumedHours: 38,
    leaveConsumedHours: 2,
    absentConsumedHours: 2,
    planCompletionRate: 93.3,
    consumedAmount: 4200,
    arrearsHours: 0,
    arrearsAmount: 0,
    hourPrice: 100,
    hourTuition: 100,
    consumedTuition: 4200,
    makeupHours: 3,
    makeupAmount: 300,
    makeupTuition: 300,
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
