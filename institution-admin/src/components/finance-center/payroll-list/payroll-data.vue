<script setup>
import { CloseOutlined } from '@ant-design/icons-vue'
import { useTableColumns } from '@/composables/useTableColumns'

const displayArray = ref([
  'intention',
  'intentionCourse',
  'reference',
  'studentStatus',
  'classEndingTime',
  'classStopTime',
])
const dataSource = ref([
  {
    key: '1',
    orderNum: '12345678901234567890',
    orderType: '收入',
    orderForm: '支付宝',
    orderTag: '默认账户',
    orderStatus: '-',
    handleContent: '2023-04-01',
    orderSalesperson: '张三',
    handledBy: '李四',
    handledDate: '2023-04-01',
    createTime: '1234567890',
    accountChanges: '2023-04-01',
    discount: '100',
    orderStatus: '已支付',
  },
])
const allColumns = ref([
  {
    title: '工资条名称/结算起止日期',
    dataIndex: 'titleOrDate',
    key: 'titleOrDate',
    width: 220,
  },

  {
    title: '结算员工',
    dataIndex: 'user',
    key: 'user',
    width: 140,
  },
  {
    title: '创建人',
    dataIndex: 'createUser',
    key: 'createUser',
    width: 120,
  },
  {
    title: '创建日期',
    dataIndex: 'createTime',
    key: 'createTime',
    width: 140,
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 120,
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    width: 120,
  },
])
const payrollSource = ref([{}, {}, {}, {}])
const columns = ref([
  {
    title: '工资条项目',
    dataIndex: 'project',
    key: 'project',
    width: 240,
  },
  {
    title: '类型',
    dataIndex: 'type',
    key: 'type',
    width: 120,
  },
  {
    title: '结算规则',
    dataIndex: 'rule',
    key: 'rule',
    width: 120,
  },
  {
    title: '规则模式',
    dataIndex: 'model',
    key: 'rule',
    width: 120,
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    width: 160,
  },

])
const { selectedValues, columnOptions, filteredColumns, totalWidth }
  = useTableColumns({
    storageKey: 'payroll-list', // 本地存储键名
    allColumns, // 原始列配置
    excludeKeys: ['action'], // 需要排除的列键
  })
</script>

<template>
  <div>
    <!-- 学员筛选条件 -->
    <div class="filter-wrap bg-white pl-3 pr-3 rounded-lb-4 rounded-rb-4">
      <all-filter
        :display-array="displayArray"
        :is-quick-show="false" :is-show-search-input="true" search-label="工资条名称"
      />
    </div>
    <div class="student-list mt-3 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            当前共计 {{ dataSource.length }} 个工资条
          </div>
          <div class="edit flex">
            <a-button type="primary" class="mr-2">
              创建工资条
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
              <template v-if="column.key === 'titleOrDate'">
                <div class="text-#222">
                  3月业绩提成工资条
                </div>
                <div class="text-3 text-#888">
                  结算起止：2025-03-01 ～ 2025-03-31
                </div>
              </template>
              <template v-if="column.key === 'user'">
                全部员工（12名）
              </template>
              <template v-if="column.key === 'createUser'">
                何洪武
              </template>
              <template v-if="column.key === 'createTime'">
                2025-04-02 15:18
              </template>
              <template v-if="column.key === 'status'">
                <div class="flex flex-items-center">
                  <span class="dot" />
                  <span>待确认</span>
                </div>
              </template>
              <template v-if="column.key === 'action'">
                <a-space :size="16">
                  <a class="font500">详情{{ record.a }}</a>
                  <a class="font500">导出{{ record.a }}</a>
                  <a class="font500">作废{{ record.a }}</a>
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
/* 添加旋转动画 */
@keyframes icon-rotate {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(180deg);
  }
}

.close-btn {
  &:hover {
    background: transparent;

    .close-icon {
      animation: icon-rotate 0.3s linear;
    }
  }
}
</style>
