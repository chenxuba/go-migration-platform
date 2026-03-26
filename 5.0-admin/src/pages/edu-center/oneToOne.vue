<script setup>
import { DownOutlined } from '@ant-design/icons-vue'
import { useTableColumns } from '@/composables/useTableColumns'

const displayArray = ref(['openClassStatus', 'doYouSchedule', 'billingMode', 'createUser', 'createTime', 'intentionCourse', 'reference', 'classEndingTime', 'classStopTime'])
const dataSource = ref([{}, {}])
const allColumns = ref([
  {
    title: '一对一',
    dataIndex: 'title',
    key: 'title',
    width: 200,
  },
  {
    title: '学员/性别',
    dataIndex: 'name',
    key: 'name',
    width: 140,
  },
  {
    title: '联系电话',
    key: 'phone',
    dataIndex: 'phone',
    width: 140,
  },
  {
    title: '班主任',
    key: 'headTeacher',
    dataIndex: 'headTeacher',
    width: 110,

  },
  {
    title: '上课时间',
    dataIndex: 'classTime',
    key: 'classTime',
    width: 300,
  },
  {
    title: '是否排课',
    dataIndex: 'doYouSchedule',
    key: 'doYouSchedule',
    width: 120,
  },
  {
    title: '已上/排课',
    dataIndex: 'alreadyOnOrtotal',
    key: 'alreadyOnOrtotal',
    width: 150,
  },
  {
    title: '创建时间',
    dataIndex: 'createTime',
    key: 'createTime',
    width: 180,

  },
  {
    title: '开课状态',
    dataIndex: 'openClassStatus',
    key: 'openClassStatus',
    width: 120,

  },

  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    fixed: 'right',
    width: 140,
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
    storageKey: 'oneToOne', // 本地存储键名
    allColumns, // 原始列配置
    excludeKeys: ['action'], // 需要排除的列键
  })
const openDrawer = ref(false)
function handleDetail(record) {
  openDrawer.value = true
}
</script>

<template>
  <div>
    <!-- 学员筛选条件 -->
    <div class="filter-wrap  bg-white  pl-3 pr-3 rounded-4">
      <all-filter
        :default-open-class-status="defaultOpenClassStatus" :display-array="displayArray"
        :is-quick-one-to-one-show="true" :is-show-one-to-one="true" search-label="班级名称"
      />
    </div>
    <div class="student-list mt-2 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            共{{ dataSource.length }}条，关联学员 1 人
          </div>
          <div class="edit flex">
            <a-dropdown class="mr-2">
              <template #overlay>
                <a-menu>
                  <a-menu-item key="1">
                    批量分配班主任
                  </a-menu-item>
                </a-menu>
              </template>
              <a-button>
                批量操作
                <DownOutlined :style="{ fontSize: '10px' }" />
              </a-button>
            </a-dropdown>
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
            <a-button type="primary" class="mr-2 w-25">
              报名
            </a-button>
            <!-- 自定义字段 -->
            <customize-code
              v-model:checked-values="selectedValues" :options="columnOptions" :total="allColumns.length - 1"
              :num="selectedValues.length - 1"
            />
          </div>
        </div>
        <div class="table-content mt-2">
          <a-table
            :data-source="dataSource" :pagination="dataSource.length > 10" :columns="filteredColumns"
            :row-selection="rowSelection" :scroll="{ x: totalWidth }" size="small"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'title'">
                妞妞-一对一认知课 {{ record.a }}
              </template>
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
                176****1613
              </template>
              <template v-if="column.key === 'studentNum'">
                0
              </template>
              <template v-if="column.key === 'headTeacher'">
                张晨
              </template>
              <template v-if="column.key === 'classRoom'">
                -
              </template>
              <template v-if="column.key === 'classTime'">
                <a-tooltip placement="right">
                  <template #title>
                    2025-04-10~2025-04-25
                  </template>
                  <div class="cursor-pointer hover-bg-#e6f4ff">
                    每周一，二，三，四，五 15:00 ~ 16:00
                  </div>
                </a-tooltip>
                <a-tooltip placement="right">
                  <template #title>
                    2025-04-10~2025-04-25
                  </template>
                  <div class="cursor-pointer hover-bg-#e6f4ff">
                    隔周一，三 16:10 ~ 17:00
                  </div>
                </a-tooltip>
              </template>
              <template v-if="column.key === 'doYouSchedule'">
                <div class="studentStatus">
                  <span class="dot" />
                  <span>已排课</span>
                </div>
              </template>
              <template v-if="column.key === 'alreadyOnOrtotal'">
                0/4节
              </template>
              <template v-if="column.key === 'openClassStatus'">
                <div class="text-#06f bg-#e6f0ff rounded-2.5 inline-block text-3 pt-0.5 pb-0.5 pl-2 pr-2">
                  开课中
                </div>
              </template>
              <template v-if="column.key === 'createTime'">
                2025-03-24 08:01
              </template>
              <template v-if="column.key === 'createUser'">
                龙钊
              </template>
              <template v-else-if="column.key === 'action'">
                <span class="flex action">
                  <a class="mr-3">排课</a>
                  <a class="mr-3">编辑</a>
                  <a @click="handleDetail(record)">详情</a>
                </span>
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>
    <one-to-one-drawer v-model:open="openDrawer" />
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
