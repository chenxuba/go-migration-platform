<script setup>
import {
  ExclamationCircleOutlined,
} from '@ant-design/icons-vue'
import { useTableColumns } from '@/composables/useTableColumns'

const displayArray = ref([
  'intention',
  'followStatus',
  'sex',
  'createUser',
  'createTime',
])
const dataSource = ref([{ key: 1 }, { key: 2 }])
const allColumns = ref([
  {
    title: '学员/电话',
    dataIndex: 'name',
    key: 'name',
    width: 160,
  },
  {
    title: '请假类型',
    dataIndex: 'leaveType',
    key: 'leaveType',
    width: 120,
  },
  {
    title: '变更后该学员总的请假次数',
    key: 'totalleaveNum',
    dataIndex: 'totalleaveNum',
    width: 190,
  },
  {
    title: '课程名称',
    key: 'courseName',
    dataIndex: 'courseName',
    width: 160,
  },
  {
    title: '变更后该课程总的请假次数',
    key: 'totalCourseleaveNum',
    dataIndex: 'totalCourseleaveNum',
    width: 190,
  },
  {
    title: '操作人',
    key: 'createUser',
    dataIndex: 'createUser',
    width: 130,
  },
  {
    title: '操作时间',
    dataIndex: 'createTime',
    key: 'createTime',
    width: 150,
  },
])
const { selectedValues, columnOptions, filteredColumns, totalWidth }
  = useTableColumns({
    storageKey: 'leave-change-record', // 本地存储键名
    allColumns, // 原始列配置
    excludeKeys: ['action'], // 需要排除的列键
  })

const openDrawer = ref(false)
function handleSeeStuData() {
  openDrawer.value = true
}
const openOrderDetailDrawer = ref(false)
function handleOrderDetail() {
  openOrderDetailDrawer.value = true
}
</script>

<template>
  <div>
    <!-- 学员筛选条件 -->
    <div class="filter-wrap bg-white pl-3 pr-3 rounded-lb-4 rounded-rb-4">
      <all-filter
        :display-array="displayArray"
        :is-quick-show="false"
        :is-show-search-stu-phonefilter="true"
      />
    </div>
    <div class="student-list mt-3 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            共 {{ dataSource.length }} 条数据
          </div>
          <div class="edit flex">
            <!-- 自定义字段 -->
            <!-- <customize-code
                  v-model:checkedValues="selectedValues"
                  :options="columnOptions"
                  :total="allColumns.length - 1"
                  :num="selectedValues.length - 1"
                /> -->
          </div>
        </div>
        <div class="table-content mt-2">
          <!-- <div class="tip mb2 text-#06f"> <ExclamationCircleFilled />  </div> -->
          <a-table
            :data-source="dataSource"
            :pagination="dataSource.length > 10"
            :columns="filteredColumns"
            :scroll="{ x: totalWidth }"
            size="small"
          >
            <template #headerCell="{ column }">
              <template v-if="column.key === 'commentStatistics'">
                <span class="mr-1">{{ column.title }}</span>
                <a-tooltip color="#666">
                  <template #title>
                    已点评人数/应点评人数
                  </template>
                  <ExclamationCircleOutlined />
                </a-tooltip>
              </template>
            </template>
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'name'">
                <a-tooltip>
                  <template #title>
                    查看学员档案
                  </template>
                  <div
                    class="flex cursor-pointer hover"
                    @click="handleSeeStuData()"
                  >
                    <img
                      width="36"
                      height="36"
                      class="mr-0"
                      style="border-radius: 100%"
                      src="https://cdn.schoolpal.cn/schoolpal/next-erp/avator_male.png?x-oss-process=image/resize,w_120"
                      alt=""
                    >
                    <div class="name">
                      <div class="text-#222 name">
                        龙龙{{ record.a }}
                      </div>
                      <div class="text-3 text-#888 flex flex-items-center">
                        176****1636
                      </div>
                    </div>
                  </div>
                </a-tooltip>
              </template>
              <template v-if="column.key === 'leaveType'">
                <div class="text-#222">
                  按学员
                </div>
              </template>
              <template v-if="column.key === 'totalleaveNum'">
                <div class="text-#222">
                  1
                </div>
              </template>
              <template v-if="column.key === 'courseName'">
                <div class="text-#222">
                  语文
                </div>
              </template>
              <template v-if="column.key === 'totalCourseleaveNum'">
                <div class="text-#222">
                  1
                </div>
              </template>
              <template v-if="column.key === 'createUser'">
                <div class="text-#222">
                  龙龙
                </div>
              </template>
              <template v-if="column.key === 'createTime'">
                <div class="text-#222">
                  2024-07-10 20:32
                </div>
              </template>

              <template v-if="column.key === 'action'">
                <a-space :size="14">
                  <a class="font500">请假详情{{ record.a }}</a>
                </a-space>
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>
    <student-info-drawer v-model:open="openDrawer" />
    <order-detail-drawer
      v-model:open="openOrderDetailDrawer"
    />
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

.tip {
  padding: 10px 24px 10px 14px;
  background: #e6f0ff;
}

.upNew {
  position: relative;

  &::before {
    position: absolute;
    top: -12px;
    left: -22px;
    z-index: 999;
    width: 39px;
    height: 22px;
    background: url(data:image/png;base64,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);
    background-size: contain;
    content: "";
  }
}
.hover {
  &:hover {
    .name {
      color: #06f;
    }
  }
}
</style>
