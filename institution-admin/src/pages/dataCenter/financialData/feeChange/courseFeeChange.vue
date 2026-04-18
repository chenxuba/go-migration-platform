<script setup>
import { BarChartOutlined, LineChartOutlined, QuestionCircleOutlined } from '@ant-design/icons-vue'
import dateFiltering from '../../components/dateFiltering.vue'
import charColumn from '../../components/charColumn.vue'
import charLine from '../../components/charLine.vue'

// 数据简报
const reportList = [
  {
    title: '扣费账户课程数',
    value: 5,
    briefing: true,
    popover_title: '扣费账户课程数',
    popover_content: '当前扣费账户的课程总数',
    chain: '',
    onYear: '',
  },
  {
    title: '新增学费总计',
    value: 6683.3,
    briefing: true,
    popover_title: '新增学费总计',
    popover_content: '新增学费的总金额',
    chain: '',
    onYear: '',
  },
  {
    title: '减少学费总计',
    value: 4916.62,
    briefing: true,
    popover_title: '减少学费总计',
    popover_content: '减少学费的总金额',
    chain: '',
    onYear: '',
  },
  {
    title: '确认收入学费',
    value: 4626.62,
    briefing: true,
    popover_title: '确认收入学费',
    popover_content: '已确认收入的学费总金额',
    chain: '',
    onYear: '',
  },
  {
    title: '剩余学费',
    value: 4376.68,
    briefing: true,
    popover_title: '剩余学费',
    popover_content: '剩余未确认的学费总金额',
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
    title: '新增学费总计',
    value: 0,
    popover_title: '新增学费总计',
    popover_content: '新增学费的总金额',
  },
  {
    title: '减少学费总计',
    value: 1,
    popover_title: '减少学费总计',
    popover_content: '减少学费的总金额',
  },
  {
    title: '确认收入学费',
    value: 2,
    popover_title: '确认收入学费',
    popover_content: '已确认收入的学费总金额',
  },
]
// 表格列
const allColumns = ref([
  {
    title: '课程名称',
    dataIndex: 'courseName',
    key: 'courseName',
    fixed: 'left',
    width: 160,
    sorter: (a, b) => a.courseName.localeCompare(b.courseName),
    required: true,
  },
  {
    title: '课程类别',
    dataIndex: 'courseType',
    key: 'courseType',
    width: 160,
    sorter: (a, b) => a.courseType.localeCompare(b.courseType),
  },
  {
    title: '在读学员数',
    dataIndex: 'studentCount',
    key: 'studentCount',
    width: 160,
    sorter: (a, b) => a.studentCount - b.studentCount,
  },
  {
    title: '报名学费',
    dataIndex: 'enrollmentFee',
    key: 'enrollmentFee',
    width: 160,
    sorter: (a, b) => a.enrollmentFee - b.enrollmentFee,
  },
  {
    title: '转入学费',
    dataIndex: 'transferInFee',
    key: 'transferInFee',
    width: 160,
    sorter: (a, b) => a.transferInFee - b.transferInFee,
  },
  {
    title: '返还学费',
    dataIndex: 'refundFee',
    key: 'refundFee',
    width: 160,
    sorter: (a, b) => a.refundFee - b.refundFee,
  },
  {
    title: '新增学费总计',
    dataIndex: 'totalNewFee',
    key: 'totalNewFee',
    width: 160,
    sorter: (a, b) => a.totalNewFee - b.totalNewFee,
  },
  {
    title: '课消学费',
    dataIndex: 'consumedFee',
    key: 'consumedFee',
    width: 160,
    sorter: (a, b) => a.consumedFee - b.consumedFee,
  },
  {
    title: '结课学费',
    dataIndex: 'completedFee',
    key: 'completedFee',
    width: 160,
    sorter: (a, b) => a.completedFee - b.completedFee,
  },
  {
    title: '转出学费',
    dataIndex: 'transferOutFee',
    key: 'transferOutFee',
    width: 160,
    sorter: (a, b) => a.transferOutFee - b.transferOutFee,
  },
  {
    title: '退费学费',
    dataIndex: 'withdrawalFee',
    key: 'withdrawalFee',
    width: 160,
    sorter: (a, b) => a.withdrawalFee - b.withdrawalFee,
  },
  {
    title: '作废学费',
    dataIndex: 'canceledFee',
    key: 'canceledFee',
    width: 160,
    sorter: (a, b) => a.canceledFee - b.canceledFee,
  },
  {
    title: '减少学费总计',
    dataIndex: 'totalReducedFee',
    key: 'totalReducedFee',
    width: 160,
    sorter: (a, b) => a.totalReducedFee - b.totalReducedFee,
  },
  {
    title: '确认收入学费',
    dataIndex: 'confirmedIncomeFee',
    key: 'confirmedIncomeFee',
    width: 160,
    sorter: (a, b) => a.confirmedIncomeFee - b.confirmedIncomeFee,
  },
  {
    title: '剩余学费',
    dataIndex: 'remainingFee',
    key: 'remainingFee',
    width: 160,
    sorter: (a, b) => a.remainingFee - b.remainingFee,
  },
])

// 表格数据
const dataSource = ref([
  {
    courseName: '数学基础班',
    courseType: '常规课程',
    studentCount: 20,
    enrollmentFee: 5000,
    transferInFee: 1000,
    refundFee: 500,
    totalNewFee: 6500,
    consumedFee: 2000,
    completedFee: 1000,
    transferOutFee: 500,
    withdrawalFee: 300,
    canceledFee: 200,
    totalReducedFee: 4000,
    confirmedIncomeFee: 2500,
    remainingFee: 4000,
  },
  {
    courseName: '英语提高班',
    courseType: '特色课程',
    studentCount: 15,
    enrollmentFee: 6000,
    transferInFee: 800,
    refundFee: 400,
    totalNewFee: 6800,
    consumedFee: 2500,
    completedFee: 800,
    transferOutFee: 300,
    withdrawalFee: 200,
    canceledFee: 100,
    totalReducedFee: 3900,
    confirmedIncomeFee: 2900,
    remainingFee: 3900,
  },
  {
    courseName: '物理竞赛班',
    courseType: '竞赛课程',
    studentCount: 10,
    enrollmentFee: 7000,
    transferInFee: 500,
    refundFee: 300,
    totalNewFee: 7500,
    consumedFee: 3000,
    completedFee: 500,
    transferOutFee: 200,
    withdrawalFee: 100,
    canceledFee: 50,
    totalReducedFee: 3850,
    confirmedIncomeFee: 3650,
    remainingFee: 3850,
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
