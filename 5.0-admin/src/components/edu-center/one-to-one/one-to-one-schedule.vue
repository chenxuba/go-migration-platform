<script setup>
import { ExclamationCircleOutlined } from '@ant-design/icons-vue'

const columns = [
  {
    title: '重复规则',
    dataIndex: 'repeatRule',
    key: 'repeatRule',
    width: 180,
  },
  {
    title: '上课时间',
    dataIndex: 'time',
    key: 'time',
    width: 100,
  },
  {
    title: '已上/排课',
    dataIndex: 'status',
    key: 'status',
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
    repeatRule: '每天重复',
    type: 1,
    time: '10:00~12:00',
    status: '1/2节',
    teacher: '张三',
    assistant: '李四',
    classroom: '101',
  },
  {
    repeatRule: '单次',
    type: 2,
    time: '10:00~12:00',
    status: '1/9节',
    teacher: '张三',
    assistant: '-',
    classroom: '-',
  },
]
</script>

<template>
  <div class="m-12px">
    <div class="bg-#fff pt-18px px-20px rounded-10px">
      <div class="flex justify-between items-center">
        <custom-title title="共 2 个日程" font-size="14px" class="pb-12px" />
        <a-button type="primary" class="mb-12px">
          一键排课
        </a-button>
      </div>
      <a-table :columns="columns" size="small" :data-source="data" :pagination="false">
        <template #headerCell="{ column }">
          <template v-if="column.key === 'status'">
            已上/排课
            <a-popover title="已上/排课">
              <template #content>
                <div>
                  已完成日程数/排课日程总数
                </div>
              </template>
              <ExclamationCircleOutlined />
            </a-popover>
          </template>
        </template>
        <template #bodyCell="{ column, record }">
          <!-- 重复规则 -->
          <template v-if="column.dataIndex === 'repeatRule'">
            <div class="flex flex-items-center">
              <img
                v-if="record.type === 1" class="w-34px h-34px"
                src="https://pcsys.admin.ybc365.com/56fe4dce-75ac-4965-bc5f-825a5e273f0a.png" alt=""
              >
              <img
                v-if="record.type === 2" class="w-34px h-34px"
                src="https://pcsys.admin.ybc365.com/0fab2373-0087-4745-bc70-a717e65172b4.png" alt=""
              >
              <div class="ml-12px text-#666">
                <div class="text-14px ">
                  {{ record.repeatRule }}
                </div>
                <div class="text-14px ">
                  2025-05-13~2025-05-18
                </div>
              </div>
            </div>
          </template>
          <!-- 上课时间 -->
          <template v-if="column.dataIndex === 'time'">
            <div>{{ record.time }}</div>
            <div class="text-#888">
              每天
            </div>
          </template>
          <!-- 操作 -->
          <template v-if="column.dataIndex === 'action'">
            <a-space :size="12">
              <a type="primary">详情</a>
              <a type="primary">编辑</a>
              <a type="primary">删除</a>
            </a-space>
          </template>
        </template>
      </a-table>
    </div>
  </div>
</template>

<style lang="less" scoped></style>
