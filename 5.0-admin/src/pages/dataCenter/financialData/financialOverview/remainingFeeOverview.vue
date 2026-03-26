<script setup>
import { BarChartOutlined, LineChartOutlined, QuestionCircleOutlined } from '@ant-design/icons-vue'
import dateFiltering from '../../components/dateFiltering.vue'
import charColumn from '../../components/charColumn.vue'
import charLine from '../../components/charLine.vue'

// 数据简报
const reportList = [
  {
    title: '总剩余学费',
    value: 4376.68,
    briefing: true,
    popover_title: '总剩余学费',
    popover_content: '所有课程剩余的学费总额',
    chain: '',
    onYear: '',
  },
  {
    title: '总剩余课时',
    value: 40,
    briefing: true,
    popover_title: '总剩余课时',
    popover_content: '所有课程剩余的课时总数',
    chain: '',
    onYear: '',
  },
  {
    title: '总剩余天数',
    value: 39,
    briefing: true,
    popover_title: '总剩余天数',
    popover_content: '所有课程剩余的有效天数',
    chain: '',
    onYear: '',
  },
  {
    title: '总剩余金额',
    value: 0,
    briefing: true,
    popover_title: '总剩余金额',
    popover_content: '所有课程剩余的金额总数',
    chain: '',
    onYear: '',
  },
]

// 日期选择组件
const dateFilter = ref(null)
// 是否显示同比、环比
const showComparison = ref(false)
const activeIcon = ref(0)
const dataType = ref(0)
const dataTypeList = [
  {
    title: '总剩余学费',
    value: 0,
    popover_title: '总剩余学费',
    popover_content: '所有课程剩余的学费总额',
  },
  {
    title: '总剩余课时',
    value: 1,
    popover_title: '总剩余课时',
    popover_content: '所有课程剩余的课时总数',
  },
  {
    title: '总剩余天数',
    value: 2,
    popover_title: '总剩余天数',
    popover_content: '所有课程剩余的有效天数',
  },
  {
    title: '总剩余金额',
    value: 3,
    popover_title: '总剩余金额',
    popover_content: '所有课程剩余的金额总数',
  },
]

// 表格数据 - 直接定义一个完整的表格
const allColumns = [
  { title: '日期', dataIndex: '日期', width: 230, fixed: 'left', sorter: (a, b) => a.日期.localeCompare(b.日期) },
  { title: '总剩余学费', dataIndex: '总剩余学费', width: 150, sorter: (a, b) => a.总剩余学费 - b.总剩余学费 },
  { title: '总剩余课时', dataIndex: '总剩余课时', width: 150, sorter: (a, b) => a.总剩余课时 - b.总剩余课时 },
  { title: '剩余天数', dataIndex: '剩余天数', width: 150, sorter: (a, b) => a.剩余天数 - b.剩余天数 },
  { title: '剩余金额', dataIndex: '剩余金额', width: 150, sorter: (a, b) => a.剩余金额 - b.剩余金额 },
]

const dataSource = [
  {
    '日期': '2025-04-28-2025-05-04',
    '总剩余学费': 4376.68,
    '总剩余课时': 40,
    '剩余天数': 39,
    '剩余金额': 0,
  },
  {
    '日期': '2025-05-05-2025-05-11',
    '总剩余学费': 4376.68,
    '总剩余课时': 40,
    '剩余天数': 39,
    '剩余金额': 0,
  },
]

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
      <div class="flex justify-between items-center pt-8px">
        <span>共计1条数据</span>
        <a-button ghost type="primary">
          下载报表
        </a-button>
      </div>
      <div class="table-content mt-2">
        <a-table :data-source="dataSource" :columns="allColumns" :pagination="false" :scroll="{ x: 2000 }">
          <template #bodyCell="{ column, record }">
            <template v-if="column.dataIndex === '时间'">
              <span>{{ record.时间 }}</span>
            </template>
          </template>
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

  td, th {
    background-color: #f0f7ff !important;
  }
}

:deep(.ant-table-summary) {
  .ant-table-cell {
    &.ant-table-cell-fix-left {
      background-color: #f0f7ff !important;
    }
  }
}
</style>
