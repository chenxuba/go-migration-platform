<script setup>
import leaveDetailsDrawer from './components/leaveDetailsDrawer.vue'
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
    width: 140,
  },
  {
    title: '开始时间',
    dataIndex: 'startTime',
    key: 'startTime',
    width: 160,
  },
  {
    title: '结束时间',
    key: 'endTime',
    dataIndex: 'endTime',
    width: 160,
  },
  {
    title: '发起人',
    key: 'createUser',
    dataIndex: 'createUser',
    width: 160,
  },
  {
    title: '处理状态',
    key: 'followStatus',
    dataIndex: 'followStatus',
    width: 130,
  },
  {
    title: '审批人',
    key: 'approvePeo',
    dataIndex: 'approvePeo',
    width: 130,
  },
  {
    title: '申请时间',
    dataIndex: 'createTime',
    key: 'createTime',
    width: 150,
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    fixed: 'right',
    width: 120,
  },
])
const { selectedValues, columnOptions, filteredColumns, totalWidth }
  = useTableColumns({
    storageKey: 'leave-list-record', // 本地存储键名
    allColumns, // 原始列配置
    excludeKeys: ['action'], // 需要排除的列键
  })

const openDrawer = ref(false)
const openLeaveDetailsDrawer = ref(false)

function handleLeaveDetails(item) {
  openLeaveDetailsDrawer.value = true
}

function handleSeeStuData() {
  openDrawer.value = true
}
const openOrderDetailDrawer = ref(false)
function handleOrderDetail() {
  openOrderDetailDrawer.value = true
}
// 请假代办
const openAddLeaveModal = ref(false)
function handleAddLeave() {
  openAddLeaveModal.value = true
}
</script>

<template>
  <div>
    <!-- 学员筛选条件 -->
    <div class="filter-wrap bg-white pl-3 pr-3 rounded-lb-4 rounded-rb-4">
      <all-filter
        :display-array="displayArray" :is-quick-show="false"
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
            <!-- 请假代办 按钮 -->
            <a-button type="primary" @click="handleAddLeave">
              请假代办
            </a-button>
          </div>
        </div>
        <div class="table-content mt-2">
          <!-- <div class="tip mb2 text-#06f"> <ExclamationCircleFilled />  </div> -->
          <a-table
            :data-source="dataSource" :pagination="dataSource.length > 10" :columns="filteredColumns"
            :scroll="{ x: totalWidth }" size="small"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'name'">
                <a-tooltip>
                  <template #title>
                    查看学员档案
                  </template>
                  <div class="flex cursor-pointer hover flex-items-center h-4 w-30 my-3" @click="handleSeeStuData()">
                    <img
                      width="36" height="36" class="mr-2" style="border-radius: 100%;"
                      src="https://cdn.schoolpal.cn/schoolpal/next-erp/avator_male.png?x-oss-process=image/resize,w_120"
                      alt=""
                    >
                    <div class="name mt-1">
                      <div class="text-#222 name">
                        龙龙
                      </div>
                      <div class="text-3 text-#888 flex flex-items-center name">
                        男 <span
                          class="inline-block w-0.2 h-2.5 bg-#ccc ml-1.5 mr-1.5 name"
                        /> 1个月
                      </div>
                    </div>
                  </div>
                </a-tooltip>
              </template>
              <template v-if="column.key === 'startTime'">
                <div class="text-#222">
                  2024-07-10 20:32
                </div>
              </template>
              <template v-if="column.key === 'endTime'">
                <div class="text-#222">
                  2024-07-10 20:32
                </div>
              </template>
              <template v-if="column.key === 'createUser'">
                <div class="text-#222">
                  龙龙（代办）
                </div>
              </template>
              <template v-if="column.key === 'followStatus'">
                <div class="text-#222">
                  <span class="bg-#e6ffec text-#0c3 text-3 px2 py1 rounded-10 font500">已通过</span>
                </div>
              </template>
              <template v-if="column.key === 'approvePeo'">
                <div class="text-#222">
                  龙龙{{ record.a }}
                </div>
              </template>
              <template v-if="column.key === 'createTime'">
                <div class="text-#222">
                  2024-07-10 20:32
                </div>
              </template>
              <template v-if="column.key === 'classRoom'">
                <div class="text-#222">
                  -
                </div>
              </template>
              <template v-if="column.key === 'action'">
                <a-space :size="14">
                  <a class="font500" @click="handleLeaveDetails(record)">请假详情{{ record.a }}</a>
                </a-space>
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>
    <leaveDetailsDrawer v-model="openLeaveDetailsDrawer" />
    <student-info-drawer v-model:open="openDrawer" />
    <order-detail-drawer v-model:open="openOrderDetailDrawer" />
    <add-leave-modal v-model:open="openAddLeaveModal" />
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
