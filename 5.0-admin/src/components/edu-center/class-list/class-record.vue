<script setup>
import { ExclamationCircleOutlined } from '@ant-design/icons-vue'

const displayArray = ref(['classEndingTime', 'classStopTime'])
const columns = ref([
  {
    title: '上课日期/时段',
    dataIndex: 'date',
    key: 'date',
    fixed: 'left',
    // 排序
    sorter: (a, b) => a.date.localeCompare(b.date),
    width: 200,
  },
  // 授课课时
  {
    title: '授课课时',
    dataIndex: 'classTime',
    key: 'classTime',
    width: 120,
  },
  {
    title: '学员消耗课时',
    dataIndex: 'studentClassTime',
    key: 'studentClassTime',
    width: 140,
  },
  // 消耗学费
  {
    title: '消耗学费',
    dataIndex: 'consumeFee',
    key: 'consumeFee',
    width: 120,
  },
  {
    title: '上课教师',
    dataIndex: 'teacher',
    key: 'teacher',
    width: 120,
  },
  {
    title: '上课助教',
    dataIndex: 'assistant',
    key: 'assistant',
    width: 120,
  },
  // 出勤率
  {
    title: '出勤率',
    dataIndex: 'attendanceRate',
    key: 'attendanceRate',
    width: 120,
  },
  // 到课（人）
  {
    title: '到课（人）',
    dataIndex: 'attendance',
    key: 'attendance',
    width: 120,
  },
  // 请假（人）
  {
    title: '请假（人）',
    dataIndex: 'leave',
    key: 'leave',
    width: 120,
  },
  // 旷课（人）
  {
    title: '旷课（人）',
    dataIndex: 'absent',
    key: 'absent',
    width: 120,
  },
  // 未记录（人）
  {
    title: '未记录（人）',
    dataIndex: 'unrecorded',
    key: 'unrecorded',
    width: 120,
  },
  // 点名时间
  {
    title: '点名时间',
    dataIndex: 'rollCallTime',
    key: 'rollCallTime',
    width: 130,
  },
  // 操作
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    width: 170,
    fixed: 'right',
  },
])

const data = ref([
  {
    date: '2025-05-13(周五)',
    classTime: '1课时',
    studentClassTime: '1课时',
    consumeFee: '￥880',
    teacher: '张三',
    assistant: '-',
    attendanceRate: '100%',
    attendance: '10',
    leave: '1',
    absent: '1',
    unrecorded: '1',
    rollCallTime: '2025-05-13 10:00',
  },

])
// 计算表格总宽度
const totalWidth = computed(() =>
  columns.value.reduce((acc, col) => acc + (col.width || 0), 0),
)
</script>

<template>
  <div class="m-12px">
    <!-- 筛选条件 -->
    <all-filter :display-array="displayArray" />
  </div>
  <div class="m-12px">
    <div class="bg-#fff pt-18px px-20px rounded-10px">
      <div class="flex justify-between items-center">
        <custom-title title="共 5 条上课记录 学员总计 3 课时，上课教师总计 4 课时 ，共消耗学费 ￥ 880" font-size="14px" class="pb-12px" />
      </div>
      <a-table :columns="columns" size="small" :data-source="data" :pagination="false" :scroll="{ x: totalWidth }">
        <template #headerCell="{ column }">
          <template v-if="column.key === 'studentClassTime'">
            学员消耗课时
            <a-popover title="学员消耗课时">
              <template #content>
                <div>
                  本日程全部学员的消耗总课时
                </div>
              </template>
              <ExclamationCircleOutlined />
            </a-popover>
          </template>
          <!-- 出勤率 -->
          <template v-if="column.key === 'attendanceRate'">
            出勤率
            <a-popover title="出勤率">
              <template #content>
                <div>
                  【出勤率】根据当前设置的数
                  据范围来统计人数。
                  计算规则：
                  实到人数/应到人数=出勤率
                </div>
              </template>
              <ExclamationCircleOutlined />
            </a-popover>
          </template>
          <!-- 消耗学费 -->
          <template v-if="column.key === 'consumeFee'">
            消耗学费
            <a-popover title="消耗学费">
              <template #content>
                <div>
                  【消耗学费】本次点名数量对应的学费（钱），即机构实际确认收入。
                </div>
              </template>
              <ExclamationCircleOutlined />
            </a-popover>
          </template>
        </template>
        <template #bodyCell="{ column, record }">
          <!-- 重复规则 -->
          <template v-if="column.dataIndex === 'date'">
            <div>{{ record.date }}</div>
            <div>08:00 ～ 08:30</div>
          </template>
          <!-- 操作 -->
          <template v-if="column.dataIndex === 'action'">
            <a-space :size="12">
              <a>编辑点名</a>
              <a>详情</a>
              <a>删除</a>
            </a-space>
          </template>
        </template>
      </a-table>
    </div>
  </div>
</template>

<style lang="less" scoped></style>
