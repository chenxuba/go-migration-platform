<script setup>
import { useTableColumns } from '@/composables/useTableColumns'

const displayArray = ref([
  'intention',
  'intentionCourse',
  'reference',
  'studentStatus',
  'classEndingTime',
  'classStopTime',
])
const openPayrollDrawer = ref(false)
const dataSource = ref([{}, {}])
const allColumns = ref([
  {
    title: '确认收入创建时间',
    dataIndex: 'createTime',
    key: 'createTime',
    width: 200,
  },

  {
    title: '学员',
    dataIndex: 'student',
    key: 'student',
    width: 120,
  },
  {
    title: '课程名称',
    dataIndex: 'courseName',
    key: 'courseName',
    width: 140,
  },
  {
    title: '课程类别',
    dataIndex: 'courseType',
    key: 'courseType',
    width: 120,
  },
  {
    title: '授课方式',
    dataIndex: 'teachingMethod',
    key: 'teachingMethod',
    width: 120,
  },
  {
    title: '明细类型',
    dataIndex: 'detailsStatus',
    key: 'detailsStatus',
    width: 140,
  },
  {
    title: '上课教师',
    dataIndex: 'teachers',
    key: 'teachers',
    width: 120,
  },
  {
    title: '上课助教',
    dataIndex: 'subTeachers',
    key: 'subTeachers',
    width: 120,
  },
  {
    title: '课消所属班级',
    dataIndex: 'linkClass',
    key: 'linkClass',
    width: 140,
  },
  {
    title: '上课时间',
    dataIndex: 'courseTime',
    key: 'courseTime',
    width: 180,
  },
  {
    title: '点名时间',
    dataIndex: 'callNameTime',
    key: 'callNameTime',
    width: 180,
  },
  {
    title: '课程消耗',
    dataIndex: 'courseUse',
    key: 'courseUse',
    fixed: 'right',
    required: true, // 新增必选标识
    width: 120,
  },
  {
    title: '确认收入',
    dataIndex: 'confirmIncome',
    key: 'confirmIncome',
    fixed: 'right',
    required: true, // 新增必选标识
    width: 120,
  },
])
const { selectedValues, columnOptions, filteredColumns, totalWidth }
  = useTableColumns({
    storageKey: 'income-details', // 本地存储键名
    allColumns, // 原始列配置
    excludeKeys: ['action'], // 需要排除的列键
  })
</script>

<template>
  <div>
    <!-- 学员筛选条件 -->
    <div class="filter-wrap bg-white pl-3 pr-3 rounded-4">
      <all-filter
        :display-array="displayArray"
        :is-quick-show="false" :is-show-search-stu-phonefilter="true"
      />
    </div>
    <div class="student-list mt-3 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            共 {{ dataSource.length }} 条确认收入记录，确认收入金额总计：￥2288.00
          </div>
          <div class="edit flex">
            <a-button class="mr-2">
              导出数据
            </a-button>
            <!-- 自定义字段 -->
            <!-- <customize-code v-model:checkedValues="selectedValues" :options="columnOptions"
              :total="allColumns.length - 1" :num="selectedValues.length - 1" /> -->
          </div>
        </div>
        <div class="table-content mt-2">
          <a-table
            :data-source="dataSource" :pagination="dataSource.length > 10" :columns="filteredColumns"
            :scroll="{ x: totalWidth }" size="small"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'createTime'">
                2025-04-13 19:58{{ record.a }}
              </template>
              <template v-if="column.key === 'student'">
                妞妞
              </template>
              <template v-if="column.key === 'courseName'">
                奥夫音乐课
              </template>
              <template v-if="column.key === 'courseType'">
                -
              </template>
              <template v-if="column.key === 'teachingMethod'">
                班课
              </template>
              <template v-if="column.key === 'detailsStatus'">
                <div class="flex flex-items-center">
                  <span class="dot" />
                  <span>课时课消</span>
                </div>
              </template>
              <template v-if="column.key === 'teachers'">
                郭杨
              </template>
              <template v-if="column.key === 'subTeachers'">
                -
              </template>
              <template v-if="column.key === 'linkClass'">
                奥夫班
              </template>
              <template v-if="column.key === 'courseTime'">
                <div class="text-#222">
                  2025-04-14
                </div>
                <div class="text-3 text-#888">
                  时段：10:00~10:30
                </div>
              </template>
              <template v-if="column.key === 'callNameTime'">
                2025-04-14 08:19
              </template>
              <template v-if="column.key === 'courseUse'">
                1课时
              </template>
              <template v-if="column.key === 'confirmIncome'">
                + 300.00
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>
  </div>
</template>

<style lang="less" scoped>
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

span.dot {
  border-radius: 50%;
  display: inline-block;
  height: 6px;
  position: relative;
  vertical-align: middle;
  width: 6px;
  margin-right: 4px;
  background: #06f;
}
</style>
