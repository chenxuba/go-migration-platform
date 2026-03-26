<script setup>
import afterSchoolTasksModel from './components/afterSchoolTasksModel.vue'
import { useTableColumns } from '@/composables/useTableColumns'

const displayArray = ref([
  'intention',
  'followStatus',
  'sex',
  'createUser',
  'applyTime',
  'intentionCourse',
  'reference',
  'studentStatus',
  'classEndingTime',
  'classStopTime',
])
const dataSource = ref([{ key: 1 }, { key: 2 }])
const defaultCreateTimeVals = ref(['2025-04-01', '2025-04-13'])
const allColumns = ref([
  {
    title: '任务名称（班级/1v1）',
    dataIndex: 'homeworkName',
    key: 'homeworkName',
    width: 180,
  },
  {
    title: '发布内容',
    dataIndex: 'publishContent',
    key: 'publishContent',
    width: 180,
  },
  {
    title: '提交任务率',
    dataIndex: 'submitRate',
    key: 'submitRate',
    width: 150,
  },

  {
    title: '待批改数量',
    dataIndex: 'pendingCorrectionNum',
    key: 'pendingCorrectionNum',
    width: 150,
  },
  {
    title: '发布人',
    dataIndex: 'publishUser',
    key: 'publishUser',
    width: 150,
  },
  {
    title: '发布时间',
    dataIndex: 'publishTime',
    key: 'publishTime',
    width: 150,
    // 排序 ，默认倒序
    sorter: {
      compare: (a, b) => a.publishTime - b.publishTime,
    },
    defaultSortOrder: 'descend',
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    fixed: 'right',
    width: 130,
  },
])

const { selectedValues, columnOptions, filteredColumns, totalWidth }
  = useTableColumns({
    storageKey: 'homework', // 本地存储键名
    allColumns, // 原始列配置
    excludeKeys: ['action'], // 需要排除的列键
  })
const openDrawer = ref(false)

const schoolTasksModel = ref(false)

function handleCreateTask(record) {
  schoolTasksModel.value = true
  schoolTasksModelTitle.value = '新建课后任务'
}

function handleEdit(record) {
  schoolTasksModel.value = true
  schoolTasksModelTitle.value = '编辑课后任务'
}

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
        :default-create-time-vals="defaultCreateTimeVals"
      />
    </div>
    <div class="student-list mt-3 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            共 {{ dataSource.length }} 个课后任务
          </div>
          <div class="edit flex">
            <a-space>
              <a-button type="primary" @click="handleCreateTask">
                新建课后任务
              </a-button>
            </a-space>
            <!-- 自定义字段 -->
            <!-- <customize-code v-model:checkedValues="selectedValues" :options="columnOptions"
                :total="allColumns.length - 1" :num="selectedValues.length - 1" /> -->
          </div>
        </div>
        <div class="table-content mt-2">
          <a-alert
            class="mb2 text-#06f"
            message="家长分享课后任务，机构就可获得转介绍线索"
            type="info"
            show-icon
            closable
          />
          <a-table
            :data-source="dataSource"
            :pagination="dataSource.length > 10"
            :columns="filteredColumns"
            :scroll="{ x: totalWidth }"
            size="small"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'homeworkName'">
                <div class="text-#222">
                  4-16 语文作业
                </div>
                <div class="text-#888 text-3">
                  陈陈-一对一认知课
                </div>
              </template>
              <template v-if="column.key === 'publishContent'">
                <div class="text-#222 w-60%">
                  <clamped-text :lines="2" text="完成【静夜思】诗词抄写10遍，并完成背诵，家长录制背诵视频上传" />
                </div>
              </template>
              <template v-if="column.key === 'submitRate'">
                <div class="text-#222">
                  100%
                </div>
                <div class="text-#888 text-3">
                  已交1人 / 应交1人
                </div>
              </template>
              <template v-if="column.key === 'pendingCorrectionNum'">
                <div class="text-#222 text-#f90">
                  1 人待批改
                </div>
              </template>
              <template v-if="column.key === 'publishUser'">
                <div class="text-#222">
                  陈瑞
                </div>
              </template>
              <template v-if="column.key === 'publishTime'">
                <div class="text-#222">
                  2025-04-16 (周三)
                </div>
                <div class="text-#888 text-3">
                  18:12
                </div>
              </template>
              <template v-if="column.key === 'action'">
                <a-space :size="14">
                  <a class="font500" @click="handleEdit(record)">编辑{{ record.a }}</a>
                  <a class="font500" @click="handleCreateTask(record)">复制</a>
                  <a class="font500">删除</a>
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
    <afterSchoolTasksModel v-model="schoolTasksModel" :title="schoolTasksModelTitle" />
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
  background: #0c3;
}

.tip {
  padding: 10px 24px 10px 14px;
  background: #e6f0ff;
  color: #333;

  a {
    color: var(--pro-ant-color-primary);
  }
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
      color: var(--pro-ant-color-primary);
    }
  }
}
</style>
