<script setup>
import { BarChartOutlined, LineChartOutlined, QuestionCircleOutlined } from '@ant-design/icons-vue'
import dateFiltering from '../../components/dateFiltering.vue'
import charColumn from '../../components/charColumn.vue'
import charLine from '../../components/charLine.vue'

// 数据简报
// 数据简报
const reportList = [
  {
    title: '班主任数量', // 标题
    value: 3, // 数值
    briefing: false, // 是否环比、同比
    popover_title: '班主任数量', // 弹出框标题
    popover_content: '所选时间段内的班主任总数', // 弹出框内容
    chain: '', // 环比
    onYear: '', // 同比
  },
  {
    title: '开课次数',
    value: 3,
    briefing: false,
    popover_title: '开课次数',
    popover_content: '所选时间段内的开课总次数',
    chain: '',
    onYear: '',
  },
  {
    title: '应到人次',
    value: 4,
    briefing: true,
    popover_title: '应到人次',
    popover_content: '所选时间段内应该到课的学员人次总数',
    chain: '',
    onYear: '',
  },
  {
    title: '实到人次',
    value: 2,
    briefing: true,
    popover_title: '实到人次',
    popover_content: '所选时间段内实际到课的学员人次总数',
    chain: '',
    onYear: '',
  },
  {
    title: '学员出勤率',
    value: '50%',
    briefing: true,
    popover_title: '学员出勤率',
    popover_content: '实到人次/应到人次 *100%',
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
    title: '开课次数',
    value: 0,
    popover_title: '开课次数',
    popover_content: '所选时间段内的开课总次数',
  },
  {
    title: '应到人次',
    value: 1,
    popover_title: '应到人次',
    popover_content: '所选时间段内应该到课的学员人次总数',
  },
  {
    title: '实到人次',
    value: 2,
    popover_title: '实到人次',
    popover_content: '所选时间段内实际到课的学员人次总数',
  },
  {
    title: '学员出勤率',
    value: 3,
    popover_title: '学员出勤率',
    popover_content: '实到人次/应到人次 *100%',
  },
  {
    title: '消耗课时数量',
    value: 4,
    popover_title: '消耗课时数量',
    popover_content: '所选时间段内消耗的课时总数',
  },
  {
    title: '消耗金额数量',
    value: 5,
    popover_title: '消耗金额数量',
    popover_content: '所选时间段内消耗的金额总数',
  },
  {
    title: '消耗学费金额',
    value: 6,
    popover_title: '消耗学费金额',
    popover_content: '所选时间段内消耗的学费金额总数',
  },
]
// 表格列
// 表格列
const allColumns = ref([
  {
    title: '班主任姓名',
    dataIndex: 'headTeacherName',
    key: 'headTeacherName',
    fixed: 'left',
    width: 160,
    sorter: (a, b) => a.headTeacherName.localeCompare(b.headTeacherName),
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
    dataIndex: 'hireDate',
    key: 'hireDate',
    width: 160,
    sorter: (a, b) => a.hireDate.localeCompare(b.hireDate),
  },
  {
    title: '当前带班数量',
    dataIndex: 'currentClassCount',
    key: 'currentClassCount',
    width: 160,
    sorter: (a, b) => a.currentClassCount - b.currentClassCount,
  },
  {
    title: '当前带1对1数量',
    dataIndex: 'currentOneOnOneCount',
    key: 'currentOneOnOneCount',
    width: 160,
    sorter: (a, b) => a.currentOneOnOneCount - b.currentOneOnOneCount,
  },
  {
    title: '关联学员人数',
    dataIndex: 'relatedStudentCount',
    key: 'relatedStudentCount',
    width: 160,
    sorter: (a, b) => a.relatedStudentCount - b.relatedStudentCount,
  },
  {
    title: '开课次数',
    dataIndex: 'classCount',
    key: 'classCount',
    width: 160,
    sorter: (a, b) => a.classCount - b.classCount,
  },
  {
    title: '应到人次',
    dataIndex: 'expectedAttendance',
    key: 'expectedAttendance',
    width: 160,
    sorter: (a, b) => a.expectedAttendance - b.expectedAttendance,
  },
  {
    title: '实到人次',
    dataIndex: 'actualAttendance',
    key: 'actualAttendance',
    width: 160,
    sorter: (a, b) => a.actualAttendance - b.actualAttendance,
  },
  {
    title: '请假人次',
    dataIndex: 'leaveCount',
    key: 'leaveCount',
    width: 160,
    sorter: (a, b) => a.leaveCount - b.leaveCount,
  },
  {
    title: '旷课人次',
    dataIndex: 'absentCount',
    key: 'absentCount',
    width: 160,
    sorter: (a, b) => a.absentCount - b.absentCount,
  },
  {
    title: '未记录人次',
    dataIndex: 'unrecordedCount',
    key: 'unrecordedCount',
    width: 160,
    sorter: (a, b) => a.unrecordedCount - b.unrecordedCount,
  },
  {
    title: '学员出勤率',
    dataIndex: 'attendanceRate',
    key: 'attendanceRate',
    width: 160,
    sorter: (a, b) => a.attendanceRate - b.attendanceRate,
  },
  {
    title: '上课点名课时数',
    dataIndex: 'rollCallHours',
    key: 'rollCallHours',
    width: 160,
    sorter: (a, b) => a.rollCallHours - b.rollCallHours,
  },
  {
    title: '上课点名金额数',
    dataIndex: 'rollCallAmount',
    key: 'rollCallAmount',
    width: 160,
    sorter: (a, b) => a.rollCallAmount - b.rollCallAmount,
  },
  {
    title: '消耗课时数量',
    dataIndex: 'consumedHours',
    key: 'consumedHours',
    width: 160,
    sorter: (a, b) => a.consumedHours - b.consumedHours,
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
    headTeacherName: '张老师',
    employmentStatus: '在职',
    hireDate: '2022-01-01',
    currentClassCount: 3,
    currentOneOnOneCount: 5,
    relatedStudentCount: 25,
    classCount: 12,
    expectedAttendance: 45,
    actualAttendance: 35,
    leaveCount: 8,
    absentCount: 2,
    unrecordedCount: 0,
    attendanceRate: 77.8,
    rollCallHours: 35,
    rollCallAmount: 3500,
    consumedHours: 35,
    consumedAmount: 3500,
    arrearsHours: 2,
    arrearsAmount: 200,
    consumedTuitionAmount: 3300,
  },
  {
    headTeacherName: '李老师',
    employmentStatus: '在职',
    hireDate: '2022-03-15',
    currentClassCount: 2,
    currentOneOnOneCount: 8,
    relatedStudentCount: 30,
    classCount: 15,
    expectedAttendance: 50,
    actualAttendance: 42,
    leaveCount: 6,
    absentCount: 2,
    unrecordedCount: 0,
    attendanceRate: 84.0,
    rollCallHours: 42,
    rollCallAmount: 4200,
    consumedHours: 42,
    consumedAmount: 4200,
    arrearsHours: 1,
    arrearsAmount: 100,
    consumedTuitionAmount: 4100,
  },
  {
    headTeacherName: '王老师',
    employmentStatus: '离职',
    hireDate: '2021-09-01',
    currentClassCount: 0,
    currentOneOnOneCount: 0,
    relatedStudentCount: 15,
    classCount: 14,
    expectedAttendance: 48,
    actualAttendance: 40,
    leaveCount: 5,
    absentCount: 3,
    unrecordedCount: 0,
    attendanceRate: 83.3,
    rollCallHours: 40,
    rollCallAmount: 4000,
    consumedHours: 40,
    consumedAmount: 4000,
    arrearsHours: 2,
    arrearsAmount: 200,
    consumedTuitionAmount: 3800,
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
