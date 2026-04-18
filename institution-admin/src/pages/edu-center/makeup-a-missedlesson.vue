<script setup>
import { DownOutlined } from '@ant-design/icons-vue'
import { useTableColumns } from '@/composables/useTableColumns'

const displayArray = ref(['openClassStatus', 'doYouSchedule', 'billingMode', 'createUser', 'createTime', 'intentionCourse', 'reference', 'classEndingTime', 'classStopTime'])
const dataSource = ref([{}, {}])
const allColumns = ref([
  {
    title: '学员/性别',
    dataIndex: 'name',
    key: 'name',
    fixed: 'left',
    width: 120,
    required: true, // 新增必选标识
  },
  {
    title: '联系电话',
    key: 'phone',
    dataIndex: 'phone',
    width: 120,
  },
  {
    title: '补课状态',
    key: 'makeCourseStatus',
    dataIndex: 'makeCourseStatus',
    width: 120,
  },
  {
    title: '缺课课程',
    key: 'missSchoolCourse',
    dataIndex: 'missSchoolCourse',
    width: 120,

  },
  {
    title: '缺课时段',
    dataIndex: 'missSchoolTime',
    key: 'missSchoolTime',
    width: 160,
  },
  {
    title: '上课教室',
    dataIndex: 'classRoom',
    key: 'classRoom',
    width: 120,
  },
  {
    title: '到课状态',
    dataIndex: 'gotoCourseStatus',
    key: 'gotoCourseStatus',
    width: 120,
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    fixed: 'right',
    width: 220,
  },

])
const rowSelection = {
  onChange: (selectedRowKeys, selectedRows) => {
    console.log(`selectedRowKeys: ${selectedRowKeys}`, 'selectedRows: ', selectedRows)
  },
}
const defaultOpenClassStatus = ref(1)
const { selectedValues, columnOptions, filteredColumns, totalWidth }
  = useTableColumns({
    storageKey: 'makeup-a-missedlesson', // 本地存储键名
    allColumns, // 原始列配置
    excludeKeys: ['action'], // 需要排除的列键
  })
</script>

<template>
  <div>
    <!-- 学员筛选条件 -->
    <div class="filter-wrap  bg-white  pl-3 pr-3 rounded-4">
      <all-filter
        :default-open-class-status="defaultOpenClassStatus" :display-array="displayArray" :is-quick-show="false"
        :is-show-search-stu-phone="false" :is-show-search-stu-phonefilter="true" search-label="班级名称"
      />
    </div>
    <div class="student-list mt-2 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            当前共{{ dataSource.length }}条数据
          </div>
          <div class="edit flex">
            <a-button class="mr-2">
              无需补课
            </a-button>
            <a-dropdown class="mr-2">
              <template #overlay>
                <a-menu>
                  <a-menu-item key="1">
                    批量导出
                  </a-menu-item>
                  <a-menu-item key="2">
                    导出记录
                  </a-menu-item>
                </a-menu>
              </template>
              <a-button>
                导出数据
                <DownOutlined :style="{ fontSize: '10px' }" />
              </a-button>
            </a-dropdown>
            <!-- 自定义字段 -->
            <customize-code
              v-model:checked-values="selectedValues" :options="columnOptions" :total="allColumns.length"
              :num="selectedValues.length"
            />
          </div>
        </div>
        <div class="table-content mt-2">
          <a-table
            :data-source="dataSource" :pagination="dataSource.length > 10" :columns="filteredColumns"
            :row-selection="rowSelection" :scroll="{ x: totalWidth }" size="small"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'name'">
                <div class="flex">
                  <img
                    width="40" height="40" class="mr-2" style="border-radius: 100%;"
                    src="https://cdn.schoolpal.cn/schoolpal/next-erp/avator_male.png?x-oss-process=image/resize,w_120"
                    alt=""
                  >
                  <div class="name mt-1">
                    <div class="text-#222">
                      龙龙
                    </div>
                    <div class="text-3 text-#888 flex flex-items-center">
                      男
                    </div>
                  </div>
                </div>
              </template>
              <template v-if="column.key === 'phone'">
                <div class="name">
                  <div class="text-#222">
                    爸爸
                  </div>
                  <div class="text-3 text-#666">
                    176****1636
                  </div>
                </div>
              </template>
              <template v-if="column.key === 'makeCourseStatus'">
                未安排
              </template>
              <template v-if="column.key === 'missSchoolCourse'">
                初级认知课
              </template>
              <template v-if="column.key === 'missSchoolTime'">
                <div class="name">
                  <div class="text-#222">
                    2025-04-10 (周四)
                  </div>
                  <div class="text-3 text-#666">
                    09:00 ～ 10:00
                  </div>
                </div>
              </template>
              <template v-if="column.key === 'classRoom'">
                郭杨
              </template>
              <template v-if="column.key === 'gotoCourseStatus'">
                请假
              </template>
              <template v-if="column.key === 'action'">
                <span class="flex action">
                  <!-- 插班补课 、取消补课 -->
                  <a class="font500">一对一补课 {{ record.a }}</a>
                  <a class="mx-3 font500">添加补课记录</a>
                  <a class="font500">详情</a>
                </span>
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

.studentStatus {
  display: flex;
  align-items: center;

  span.dot {
    border-radius: 50%;
    display: inline-block;
    height: 6px;
    position: relative;
    vertical-align: middle;
    width: 6px;
    margin-right: 4px;
    background: var(--pro-ant-color-primary);
  }
}
</style>
