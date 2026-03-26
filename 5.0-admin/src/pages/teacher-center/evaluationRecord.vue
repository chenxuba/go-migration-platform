<script setup>
import { DownOutlined } from '@ant-design/icons-vue'
import { useTableColumns } from '@/composables/useTableColumns'

const displayArray = ref(['intention', 'createTime', 'intentionCourse', 'reference', 'classEndingTime', 'classStopTime', 'currentStatus', 'orNotFenClass'])
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
    title: '测评类型',
    dataIndex: 'type',
    width: 180,
    fixed: 'left',
    key: 'type',
  },
  {
    title: '测评日期',
    key: 'date',
    dataIndex: 'date',
    width: 140,
  },
  {
    title: '展示对比',
    key: 'showCompare',
    dataIndex: 'showCompare',
    width: 100,

  },
  {
    title: '家长端是否展示',
    dataIndex: 'isShow',
    key: 'isShow',
    width: 110,
  },
  {
    title: '下次测评建议日期',
    dataIndex: 'nextEvaluationDate',
    key: 'nextEvaluationDate',
    width: 160,
  },
  // 操作
  {
    title: '操作',
    key: 'action',
    dataIndex: 'action',
    fixed: 'right',
    width: 120,
  },
])
const rowSelection = {
  onChange: (selectedRowKeys, selectedRows) => {
    console.log(`selectedRowKeys: ${selectedRowKeys}`, 'selectedRows: ', selectedRows)
  },
}
const defaultOrNotFenClass = ref([0])
const defaultCurrentStatus = ref([1, 3])
const { selectedValues, columnOptions, filteredColumns, totalWidth }
  = useTableColumns({
    storageKey: 'evaluationRecord', // 本地存储键名
    allColumns, // 原始列配置
    excludeKeys: ['action'], // 需要排除的列键
  })
</script>

<template>
  <div>
    <!-- 学员筛选条件 -->
    <div class="filter-wrap  bg-white  pl-3 pr-3 rounded-4">
      <all-filter
        :default-or-not-fen-class="defaultOrNotFenClass" :default-current-status="defaultCurrentStatus"
        :display-array="displayArray" :is-quick-show="false" :is-show-search-stu-phonefilter="true"
      />
    </div>
    <div class="student-list mt-2 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            总计 {{ dataSource.length }} 条
          </div>
          <div class="edit flex">
            <a-dropdown class="mr-2">
              <template #overlay>
                <a-menu>
                  <a-menu-item key="1">
                    批量停课
                  </a-menu-item>
                  <a-menu-item key="2">
                    批量复课
                  </a-menu-item>
                  <a-menu-item key="3">
                    批量结课
                  </a-menu-item>
                  <a-menu-item key="4">
                    批量转课
                  </a-menu-item>
                  <a-menu-item key="5">
                    批量修改有效期
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
              <template v-if="column.key === 'type'">
                语言沟通能力评估
              </template>
              <template v-if="column.key === 'date'">
                2024-07-10 10:06
              </template>
              <template v-if="column.key === 'showCompare'">
                <!-- 开关 -->
                <a-switch v-model:checked="record.a" />
              </template>
              <template v-if="column.key === 'isShow'">
                <a-switch v-model:checked="record.a" />
              </template>
              <template v-if="column.key === 'nextEvaluationDate'">
                <!-- 日期选择器 -->
                <a-date-picker v-model:value="record.a" />
              </template>
              <template v-if="column.key === 'action'">
                <a-space :size="15">
                  <a>查看报告</a>
                  <a>删除报告</a>
                </a-space>
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
