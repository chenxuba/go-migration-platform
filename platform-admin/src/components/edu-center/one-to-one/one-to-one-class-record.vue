<script setup>
import { ExclamationCircleOutlined } from '@ant-design/icons-vue'

const displayArray = ref(['classEndingTime', 'classStopTime'])
const columns = [
  {
    title: '上课日期/时段',
    dataIndex: 'date',
    key: 'date',
    // 排序
    sorter: (a, b) => a.date.localeCompare(b.date),
    width: 120,
  },
  // 授课课时
  {
    title: '授课课时',
    dataIndex: 'classTime',
    key: 'classTime',
    width: 80,
  },
  {
    title: '学员消耗课时',
    dataIndex: 'studentClassTime',
    key: 'studentClassTime',
    width: 80,
  },
  // 消耗学费
  {
    title: '消耗学费',
    dataIndex: 'consumeFee',
    key: 'consumeFee',
    width: 80,
  },
  {
    title: '上课教师',
    dataIndex: 'teacher',
    key: 'teacher',
    width: 80,
  },
  {
    title: '上课助教',
    dataIndex: 'assistant',
    key: 'assistant',
    width: 80,
  },
  {
    title: '点名时间',
    dataIndex: 'rollCallTime',
    key: 'rollCallTime',
    width: 80,
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    width: 100,
  },
]

const data = [
  {
    date: '2025-05-13(周五)',
    classTime: '1课时',
    studentClassTime: '1课时',
    consumeFee: '￥880',
    teacher: '张三',
    assistant: '-',
    rollCallTime: '2025-05-13 10:00',
  },
  {
    date: '2025-05-14(周六)',
    classTime: '1课时',
    studentClassTime: '1课时',
    consumeFee: '￥880',
    teacher: '张三',
    assistant: '-',
    rollCallTime: '2025-05-14 10:00',
  },
]
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
      <a-table :columns="columns" size="small" :data-source="data" :pagination="false">
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
