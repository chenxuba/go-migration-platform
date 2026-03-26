<script setup>
const displayArray = ref(['classEndingTime'])
const columns = [
  {
    title: '上课日期/时段',
    dataIndex: 'date',
    key: 'date',
    // 排序
    sorter: (a, b) => a.date.localeCompare(b.date),
    width: 120,
  },
  {
    title: '课程名称',
    dataIndex: 'courseName',
    key: 'courseName',
    width: 140,
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
    title: '上课教室',
    dataIndex: 'classroom',
    key: 'classroom',
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
    courseName: '课程名称',
    teacher: '张三',
    assistant: '李四',
    classroom: '101',
  },
  {
    date: '2025-05-14(周六)',
    courseName: '课程名称',
    teacher: '张三',
    assistant: '李四',
    classroom: '101',
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
        <custom-title title="当前共计 3 条待点名日程" font-size="14px" class="pb-12px" />
        <a-button type="primary" class="mb-12px">
          创建未排课点名
        </a-button>
      </div>
      <a-table :columns="columns" size="small" :data-source="data" :pagination="false">
        <template #bodyCell="{ column, record }">
          <!-- 重复规则 -->
          <template v-if="column.dataIndex === 'date'">
            <div>{{ record.date }}</div>
            <div>08:00 ～ 08:30</div>
          </template>
          <!-- 操作 -->
          <template v-if="column.dataIndex === 'action'">
            <a-space :size="12">
              <a>点名</a>
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
