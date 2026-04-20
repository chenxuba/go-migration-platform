<script setup>
import { ExclamationCircleOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
// 时间范围
const dateRange = ref([])
// 快捷日期范围选项
const dateRangeOptions = [
  { label: '本周', value: 'week' },
  { label: '上周', value: 'lastWeek' },
  { label: '本月', value: 'month' },
  { label: '上月', value: 'lastMonth' },
]
// 快捷日期选项
const activeDateType = ref('')

// 监听日期选项变化
watch(activeDateType, (newVal, oldVal) => {
  switch (newVal) {
    case 'week':
      dateRange.value = [dayjs().startOf('week'), dayjs().endOf('week')]
      break
    case 'lastWeek':
      dateRange.value = [dayjs().subtract(1, 'week').startOf('week'), dayjs().subtract(1, 'week').endOf('week')]
      break
    case 'month':
      dateRange.value = [dayjs().startOf('month'), dayjs().endOf('month')]
      break
    case 'lastMonth':
      dateRange.value = [dayjs().subtract(1, 'month').startOf('month'), dayjs().subtract(1, 'month').endOf('month')]
      break
    default:
      break
  }
})

// 维度选项
const dimensionOptions = [
  { label: '销售员', value: 'salesman' },
  { label: '上课教师', value: 'teacher' },
  { label: '渠道', value: 'channel' },
]

// 维度
const dimension = ref('salesman')

// 销售员维度指标
const salesmanDimension = [
  { label: '销售员姓名', value: 0, disabled: true },
  { label: '现有意向学员数', value: 1, tooltip: '查询结果不根据日期范围筛选变化' },
  { label: '现有在读学员数', value: 2, tooltip: '查询结果不根据日期范围筛选变化' },
  { label: '现有历史学员数', value: 3, tooltip: '查询结果不根据日期范围筛选变化' },
  { label: '跟进记录数', value: 4 },
  { label: '试听邀约数', value: 5 },
  { label: '试听完成数', value: 6 },
  { label: '试听转化数', value: 7, tooltip: '试听记录手动或自动标记为已转化的数量' },
  { label: '试听完成率', value: 8, tooltip: '试听完成数/试听邀约数' },
  { label: '试听转化率', value: 9, tooltip: '试听转化数/试听完成数' },
  { label: '成交人数', value: 10 },
  { label: '成交订单数', value: 11 },
  { label: '应收成交额', value: 12 },
  { label: '实收成交额', value: 13 },
  { label: '业绩金额', value: 14, tooltip: '试听转化数/试听完成数' },
]

// 教师维度指标
const teacherDimension = [
  { label: '上课教师', value: 0, disabled: true },
  { label: '试听邀约数', value: 1 },
  { label: '试听完成数', value: 2 },
  { label: '试听转化数', value: 3, tooltip: '试听记录手动或自动标记为已转化的数量' },
  { label: '试听完成率', value: 4, tooltip: '试听完成数/试听邀约数' },
  { label: '试听转化率', value: 5, tooltip: '试听转化数/试听完成数' },
]

// 渠道维度指标
const channelDimension = [
  { label: '渠道名称', value: 0, disabled: true },
  { label: '现有意向学员数', value: 1, tooltip: '查询结果不根据日期范围筛选变化' },
  { label: '现有在读学员数', value: 2, tooltip: '查询结果不根据日期范围筛选变化' },
  { label: '现有历史学员数', value: 3, tooltip: '查询结果不根据日期范围筛选变化' },
  { label: '试听邀约数', value: 4 },
  { label: '试听完成数', value: 5 },
  { label: '试听转化数', value: 6, tooltip: '试听记录手动或自动标记为已转化的数量' },
  { label: '试听完成率', value: 7, tooltip: '试听完成数/试听邀约数' },
  { label: '试听转化率', value: 8, tooltip: '试听转化数/试听完成数' },
  { label: '成交人数', value: 9 },
  { label: '成交订单数', value: 10 },
  { label: '应收成交额', value: 11 },
  { label: '实收成交额', value: 12 },
]

const dimensionMap = {
  salesman: salesmanDimension,
  teacher: teacherDimension,
  channel: channelDimension,
}

// 维度映射
const plainOptions = computed(() => {
  return dimensionMap[dimension.value]
})
// 多选状态
const state = reactive({
  indeterminate: true,
  checkAll: false,
  checkedList: [0, 1, 2], // 修改为对应的value值
})

// 全选
function onCheckAllChange(e) {
  Object.assign(state, {
    checkedList: e.target.checked ? plainOptions.value.map(option => option.value) : [],
    indeterminate: false,
  })
};

// 表格列
const allColumns = ref([
  {
    title: '销售员姓名',
    dataIndex: 'salesmanName',
    key: 'salesmanName',
    fixed: 'left',
    width: 160,
    sorter: (a, b) => a.salesmanName.localeCompare(b.salesmanName),
    required: true,
  },
  {
    title: '现有意向学员数',
    dataIndex: 'potentialStudents',
    key: 'potentialStudents',
    width: 160,
    sorter: (a, b) => a.potentialStudents - b.potentialStudents,
  },
  {
    title: '现有在读学员数',
    dataIndex: 'currentStudents',
    key: 'currentStudents',
    width: 160,
    sorter: (a, b) => a.currentStudents - b.currentStudents,
  },
  {
    title: '现有历史学员数',
    dataIndex: 'historicalStudents',
    key: 'historicalStudents',
    width: 160,
    sorter: (a, b) => a.historicalStudents - b.historicalStudents,
  },
  {
    title: '跟进记录数',
    dataIndex: 'followUpRecords',
    key: 'followUpRecords',
    width: 160,
    sorter: (a, b) => a.followUpRecords - b.followUpRecords,
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
    dataIndex: 'trialCompletions',
    key: 'trialCompletions',
    width: 160,
    sorter: (a, b) => a.trialCompletions - b.trialCompletions,
  },
  {
    title: '试听转化数',
    dataIndex: 'trialConversions',
    key: 'trialConversions',
    width: 160,
    sorter: (a, b) => a.trialConversions - b.trialConversions,
  },
  {
    title: '试听完成率',
    dataIndex: 'trialCompletionRate',
    key: 'trialCompletionRate',
    width: 160,
    sorter: (a, b) => a.trialCompletionRate - b.trialCompletionRate,
  },
  {
    title: '试听转化率',
    dataIndex: 'trialConversionRate',
    key: 'trialConversionRate',
    width: 160,
    sorter: (a, b) => a.trialConversionRate - b.trialConversionRate,
  },
  {
    title: '成交人数',
    dataIndex: 'dealCount',
    key: 'dealCount',
    width: 160,
    sorter: (a, b) => a.dealCount - b.dealCount,
  },
  {
    title: '成交订单数',
    dataIndex: 'dealOrders',
    key: 'dealOrders',
    width: 160,
    sorter: (a, b) => a.dealOrders - b.dealOrders,
  },
  {
    title: '应收成交额',
    dataIndex: 'receivableAmount',
    key: 'receivableAmount',
    width: 160,
    sorter: (a, b) => a.receivableAmount - b.receivableAmount,
  },
  {
    title: '实收成交额',
    dataIndex: 'receivedAmount',
    key: 'receivedAmount',
    width: 160,
    sorter: (a, b) => a.receivedAmount - b.receivedAmount,
  },
  {
    title: '业绩金额',
    dataIndex: 'performanceAmount',
    key: 'performanceAmount',
    width: 160,
    sorter: (a, b) => a.performanceAmount - b.performanceAmount,
  },
])

// 表格数据
const dataSource = ref([
  {
    salesmanName: '张三',
    potentialStudents: 15,
    currentStudents: 30,
    historicalStudents: 50,
    followUpRecords: 45,
    trialInvitations: 20,
    trialCompletions: 15,
    trialConversions: 8,
    trialCompletionRate: 75,
    trialConversionRate: 53.3,
    dealCount: 5,
    dealOrders: 5,
    receivableAmount: 25000,
    receivedAmount: 20000,
    performanceAmount: 18000,
  },
  {
    salesmanName: '李四',
    potentialStudents: 20,
    currentStudents: 25,
    historicalStudents: 40,
    followUpRecords: 35,
    trialInvitations: 15,
    trialCompletions: 12,
    trialConversions: 6,
    trialCompletionRate: 80,
    trialConversionRate: 50,
    dealCount: 4,
    dealOrders: 4,
    receivableAmount: 20000,
    receivedAmount: 18000,
    performanceAmount: 16000,
  },
])
// 图表数据源
const { selectedValues, columnOptions, filteredColumns, totalWidth }
  = useTableColumns({
    storageKey: 'enrollmentDataReport', // 本地存储键名
    allColumns, // 原始列配置
    excludeKeys: [], // 需要排除的列键
  })
</script>

<template>
  <div>
    <div class="card-white flex flex-col gap-20px">
      <!-- 时间 -->
      <div class="flex items-center gap-5px">
        <div class="label text-gray">
          时间：
        </div>
        <div class="value">
          <a-range-picker v-model:value="dateRange" class="mr-10px" value-format="YYYY-MM-DD" />
          <a-radio-group v-model:value="activeDateType" button-style="solid">
            <a-radio-button v-for="(item, index) in dateRangeOptions" :key="index" :value="item.value">
              {{ item.label }}
            </a-radio-button>
          </a-radio-group>
        </div>
      </div>
      <!-- 维度 -->
      <div class="flex items-center gap-5px">
        <div class="label text-gray">
          维度：
        </div>
        <div class="value">
          <a-radio-group v-model:value="dimension">
            <a-radio v-for="(item, index) in dimensionOptions" :key="index" class="custom-radio" :value="item.value">
              {{ item.label }}
            </a-radio>
          </a-radio-group>
        </div>
      </div>
      <!-- 指标 -->
      <div class="flex items-start gap-5px">
        <div class="label text-gray" style="white-space: nowrap;">
          指标：
        </div>
        <div class="value flex items-start">
          <div class="min-w-65px">
            <a-checkbox
              v-model:checked="state.checkAll" :indeterminate="state.indeterminate"
              @change="onCheckAllChange"
            >
              全选
            </a-checkbox>
          </div>

          <a-checkbox-group v-model:value="state.checkedList" class="gap-10px" :options="plainOptions">
            <template #label="{ label, tooltip }">
              <div class="flex items-center">
                <span class="mr-5px">{{ label }}</span>
                <a-popover v-if="tooltip" color="#fff" title="字段说明">
                  <template #content>
                    <div v-html="tooltip" />
                  </template>
                  <ExclamationCircleOutlined class="text-#888 font-size-14px" />
                </a-popover>
              </div>
            </template>
          </a-checkbox-group>
        </div>
      </div>
      <!-- 按钮 -->
      <div class="flex justify-end gap-20px">
        <a-button type="default">
          历史报表
        </a-button>
        <a-button type="primary">
          生成报表
        </a-button>
      </div>
      <!-- 报表配置项 -->
      <div class="flex flex-col gap-10px p-x-50px">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-15px">
            <img
              class="w-25px h-25px" src="https://pcsys.admin.ybc365.com/600beacc-92a4-49d0-8769-43751a503456.png"
              alt=""
            >
            <span class="font-bold font-size-20px">2025-05-12 ~ 2025-05-18</span>
          </div>
          <a-button type="primary">
            下载
          </a-button>
        </div>
        <div>
          <span class="text-#888">生成类型：</span>
          <span>招生数据</span>
        </div>
        <div>
          <span class="text-#888">创建日期：</span>
          <span>2025-05-12 17:16 | 何红武</span>
        </div>
        <div>
          <span class="text-#888">维度：</span>
          <span>销售员</span>
        </div>
        <div>
          <span class="text-#888">指标：</span>
          <span v-for="(item, index) in state.checkedList" :key="index">
            {{ plainOptions.find(option => option.value === item)?.label }}
            {{ index < state.checkedList.length - 1 ? '、' : '' }} </span>
        </div>
      </div>
      <!-- 报表明细 -->
      <div class="p-x-30px">
        <custom-title title="报表明细" font-size="15px" font-weight="500" />
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
  // margin-top: 8px;
  padding: 12px;
  border-bottom-left-radius: 12px;
  border-bottom-right-radius: 12px;
}
</style>
