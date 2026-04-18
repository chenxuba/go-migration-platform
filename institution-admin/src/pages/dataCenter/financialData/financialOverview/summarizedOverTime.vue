<script setup>
import { BarChartOutlined, LineChartOutlined, QuestionCircleOutlined } from '@ant-design/icons-vue'
import dateFiltering from '../../components/dateFiltering.vue'
import charColumn from '../../components/charColumn.vue'
import charLine from '../../components/charLine.vue'

// 数据简报
const reportList = [
  {
    title: '总收入金额',
    value: 6393.3,
    briefing: false,
    popover_title: '总收入金额',
    popover_content: '所有渠道收入的总金额',
    chain: '',
    onYear: '',
  },
  {
    title: '总支出金额',
    value: 0,
    briefing: false,
    popover_title: '总支出金额',
    popover_content: '所有渠道支出的总金额',
    chain: '',
    onYear: '',
  },
  {
    title: '结余金额',
    value: 6393.3,
    briefing: false,
    popover_title: '结余金额',
    popover_content: '总收入金额减去总支出金额',
    chain: '',
    onYear: '',
  },
  {
    title: '新增学费',
    value: 6683.3,
    briefing: true,
    popover_title: '新增学费',
    popover_content: '本期新增的学费金额',
    chain: '',
    onYear: '',
  },
  {
    title: '减少学费',
    value: 4336.63,
    briefing: true,
    popover_title: '减少学费',
    popover_content: '本期减少的学费金额',
    chain: '',
    onYear: '',
  },
  {
    title: '确认收入学费',
    value: 4046.63,
    briefing: true,
    popover_title: '确认收入学费',
    popover_content: '已确认为收入的学费金额',
    chain: '',
    onYear: '',
  },
  {
    title: '剩余学费',
    value: 4956.67,
    briefing: true,
    popover_title: '剩余学费',
    popover_content: '尚未使用的学费金额',
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
    title: '总收入金额',
    value: 0,
    popover_title: '总收入金额',
    popover_content: '订单收入+记账收入',
  },
  {
    title: '总支出金额',
    value: 1,
    popover_title: '总支出金额',
    popover_content: '订单支出+记账支出',
  },
  {
    title: '结余金额',
    value: 2,
    popover_title: '结余金额',
    popover_content: '总收入-总支出',
  },
  {
    title: '新增学费',
    value: 3,
    popover_title: '新增学费',
    popover_content: '学费增加金额总计<br />时间：学费变动时间<br />维度：学费变动记录类型一级为报名学费、转入学费、返还学费',
  },
  {
    title: '减少学费',
    value: 4,
    popover_title: '减少学费',
    popover_content: '学费减少金额总计<br />时间：学费变动时间<br />维度：学费变动记录类型一级为课消学费、结课学费、转出学费、退费学费、作废学费',
  },
  {
    title: '确认收入学费',
    value: 5,
    popover_title: '确认收入学费',
    popover_content: '学费确认收入金额总计<br />时间：确认收入时间<br />维度：全部确认收入类型，包含转/退课手续费<br />注意，确认收入时间为：拥有上课记录的学费消耗统计其上课时间，无上课记录的学费消耗统计其学费变动时间',
  },
]

// 表格数据 - 直接定义一个完整的表格
const combinedColumns = [
  { title: '时间', dataIndex: '时间', width: 230, fixed: 'left', sorter: (a, b) => a.时间.localeCompare(b.时间) },
  { title: '订单数量', dataIndex: '订单数量', width: 150, sorter: (a, b) => a.订单数量 - b.订单数量 },
  { title: '订单应收', dataIndex: '订单应收', width: 150, sorter: (a, b) => a.订单应收 - b.订单应收 },
  { title: '订单实收', dataIndex: '订单实收', width: 150, sorter: (a, b) => a.订单实收 - b.订单实收 },
  { title: '订单失效', dataIndex: '订单失效', width: 150, sorter: (a, b) => a.订单失效 - b.订单失效 },
  { title: '订单未收', dataIndex: '订单未收', width: 150, sorter: (a, b) => a.订单未收 - b.订单未收 },
  { title: '订单支出', dataIndex: '订单支出', width: 150, sorter: (a, b) => a.订单支出 - b.订单支出 },
  { title: '订单结余', dataIndex: '订单结余', width: 150, sorter: (a, b) => a.订单结余 - b.订单结余 },
  { title: '报课结余', dataIndex: '报课结余', width: 150, sorter: (a, b) => a.报课结余 - b.报课结余 },
  { title: '学杂费结余', dataIndex: '学杂费结余', width: 150, sorter: (a, b) => a.学杂费结余 - b.学杂费结余 },
  { title: '教学用品结余', dataIndex: '教学用品结余', width: 150, sorter: (a, b) => a.教学用品结余 - b.教学用品结余 },
  { title: '约课付费结余', dataIndex: '约课付费结余', width: 150, sorter: (a, b) => a.约课付费结余 - b.约课付费结余 },
  { title: '健值服产结余', dataIndex: '健值服产结余', width: 150, sorter: (a, b) => a.健值服产结余 - b.健值服产结余 },
  { title: '校校杯课程结余', dataIndex: '校校杯课程结余', width: 150, sorter: (a, b) => a.校校杯课程结余 - b.校校杯课程结余 },
  { title: '捐赠预付结余', dataIndex: '捐赠预付结余', width: 150, sorter: (a, b) => a.捐赠预付结余 - b.捐赠预付结余 },
  { title: '订单验单收入', dataIndex: '订单验单收入', width: 150, sorter: (a, b) => a.订单验单收入 - b.订单验单收入 },
  { title: '订单验单支出', dataIndex: '订单验单支出', width: 150, sorter: (a, b) => a.订单验单支出 - b.订单验单支出 },
  { title: '记账收入', dataIndex: '记账收入', width: 150, sorter: (a, b) => a.记账收入 - b.记账收入 },
  { title: '记账支出', dataIndex: '记账支出', width: 150, sorter: (a, b) => a.记账支出 - b.记账支出 },
  { title: '总收入', dataIndex: '总收入', width: 150, sorter: (a, b) => a.总收入 - b.总收入 },
  { title: '总支出', dataIndex: '总支出', width: 150, sorter: (a, b) => a.总支出 - b.总支出 },
  { title: '总结余', dataIndex: '总结余', width: 150, sorter: (a, b) => a.总结余 - b.总结余 },
  { title: '新增学费总计', dataIndex: '新增学费总计', width: 150, sorter: (a, b) => a.新增学费总计 - b.新增学费总计 },
  { title: '减少学费总计', dataIndex: '减少学费总计', width: 150, sorter: (a, b) => a.减少学费总计 - b.减少学费总计 },
  { title: '确认收入学费', dataIndex: '确认收入学费', width: 150, sorter: (a, b) => a.确认收入学费 - b.确认收入学费 },
]

const combinedTableData = [
  {
    '时间': '2025-04-28-2025-05-04',
    '订单数量': 1,
    '订单应收': 2800,
    '订单实收': 2800,
    '订单失效': 0,
    '订单未收': 0,
    '订单支出': 2800,
    '订单结余': 2800,
    '报课结余': 0,
    '学杂费结余': 0,
    '教学用品结余': 0,
    '约课付费结余': 0,
    '健值服产结余': 0,
    '校校杯课程结余': 0,
    '捐赠预付结余': 0,
    '订单验单收入': 2800,
    '订单验单支出': 0,
    '记账收入': 0,
    '记账支出': 0,
    '总收入': 2800,
    '总支出': 0,
    '总结余': 2800,
    '新增学费总计': 2800,
    '减少学费总计': 559.98,
    '确认收入学费': 849.98,
  },
  {
    '时间': '2025-05-05-2025-05-11',
    '订单数量': 5,
    '订单应收': 3593.3,
    '订单实收': 3593.3,
    '订单失效': 0,
    '订单未收': 0,
    '订单支出': 3593.3,
    '订单结余': 3593.3,
    '报课结余': 0,
    '学杂费结余': 0,
    '教学用品结余': 0,
    '约课付费结余': 0,
    '健值服产结余': 0,
    '校校杯课程结余': 0,
    '捐赠预付结余': 0,
    '订单验单收入': 3593.3,
    '订单验单支出': 0,
    '记账收入': 0,
    '记账支出': 0,
    '总收入': 3593.3,
    '总支出': 0,
    '总结余': 3593.3,
    '新增学费总计': 3883.3,
    '减少学费总计': 3776.65,
    '确认收入学费': 3196.65,
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
        <a-table :data-source="combinedTableData" :columns="combinedColumns" :pagination="false" :scroll="{ x: 2000 }">
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
                <a-table-summary-cell v-for="(col, index) in combinedColumns.slice(1)" :key="index" :index="index + 1">
                  {{ combinedTableData.reduce((sum, record) => sum + (Number(record[col.dataIndex]) || 0), 0) }}
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
