<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useTableColumns } from '@/composables/useTableColumns'

const router = useRouter()
const displayArray = ref([
  'createTime',
  'stuPhoneSearch',
  'performanceAllocationStatus',
  'orderType',
])

const dataSource = ref([
  {
    key: '1',
    orderNum: '20250413184836423728284',
    student: {
      name: '龙龙',
      phone: '176****1636',
    },
    allocationStatus: '待分配',
    performanceOwner: '-',
    allocationAmount: '0.00',
    handleContent: '初级感统课',
    orderType: '报名续费',
    totalAmount: '200.00',
    createTime: '2025-04-13 19:04',
  },
  {
    key: '2',
    orderNum: '20250413184836423728285',
    student: {
      name: '小明',
      phone: '138****2345',
    },
    allocationStatus: '待分配',
    performanceOwner: '-',
    allocationAmount: '0.00',
    handleContent: '高级感统课',
    orderType: '报名续费',
    totalAmount: '300.00',
    createTime: '2025-04-13 18:30',
  },
  {
    key: '3',
    orderNum: '20250413184836423728286',
    student: {
      name: '小红',
      phone: '159****7890',
    },
    allocationStatus: '待分配',
    performanceOwner: '-',
    allocationAmount: '0.00',
    handleContent: '语言训练课',
    orderType: '报名续费',
    totalAmount: '400.00',
    createTime: '2025-04-13 16:45',
  },
  {
    key: '4',
    orderNum: '20250413184836423728287',
    student: {
      name: '小华',
      phone: '186****5678',
    },
    allocationStatus: '待分配',
    performanceOwner: '-',
    allocationAmount: '0.00',
    handleContent: '专注力训练课',
    orderType: '报名续费',
    totalAmount: '350.00',
    createTime: '2025-04-13 15:20',
  },
])

// 添加表格多选相关变量
const state = reactive({
  selectedRowKeys: [],
  selectedRows: [],
})

// 多选变化处理函数
function onSelectChange(keys, rows) {
  state.selectedRowKeys = keys
  state.selectedRows = rows
  console.log('selectedRowKeys changed: ', keys)
  console.log('selectedRows changed: ', rows)
}

const allColumns = ref([
  {
    title: '订单编号',
    dataIndex: 'orderNum',
    key: 'orderNum',
    fixed: 'left',
    width: 250,
    required: true,
  },
  {
    title: '关联学员',
    dataIndex: 'student',
    key: 'student',
    width: 160,
  },
  {
    title: '分配状态',
    dataIndex: 'allocationStatus',
    key: 'allocationStatus',
    width: 120,
  },
  {
    title: '业绩归属人',
    dataIndex: 'performanceOwner',
    key: 'performanceOwner',
    width: 120,
  },
  {
    title: '业绩分配金额（元）',
    dataIndex: 'allocationAmount',
    key: 'allocationAmount',
    width: 150,
  },
  {
    title: '交易内容',
    dataIndex: 'handleContent',
    key: 'handleContent',
    width: 150,
  },
  {
    title: '订单类型',
    dataIndex: 'orderType',
    key: 'orderType',
    width: 120,
  },
  {
    title: '订单总金额（元）',
    dataIndex: 'totalAmount',
    key: 'totalAmount',
    width: 150,
  },
  {
    title: '订单创建时间',
    dataIndex: 'createTime',
    key: 'createTime',
    width: 160,
    sorter: true,
    // 默认倒叙
    defaultSortOrder: 'descend',
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    fixed: 'right',
    width: 160,
  },
])
const defaultCreateTimeVals = ref(['2025-04-01', '2025-04-13'])

const { selectedValues, columnOptions, filteredColumns, totalWidth }
  = useTableColumns({
    storageKey: 'performance-management', // 本地存储键名
    allColumns, // 原始列配置
    excludeKeys: ['action'], // 需要排除的列键
  })
function handleAllocation(record) {
  console.log(record)
  router.push({
    path: `/finance-center/performance-edit/${record.orderNum}`,
  })
}

function handleNoAllocation(record) {
  console.log(record)
}
</script>

<template>
  <div>
    <!-- 学员筛选条件 -->
    <div class="filter-wrap bg-white pl-3 pr-3 rounded-lb-4 rounded-rb-4">
      <all-filter
        type="noDelCreateTime" create-time-label="订单创建时间" :default-create-time-vals="defaultCreateTimeVals" :display-array="displayArray"
        :is-quick-show="false" :is-show-search-input="true" search-label="订单编号" search-placeholder="请输入订单编号"
      />
    </div>
    <div class="student-list mt-3 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total flex flex-col">
            <span>订单总金额：¥ 11500，已分配订单金额：¥ 0，无需分配订单金额：¥ 0，未分配订单金额：¥ 11500</span>
            <span class="text-#666">共8条可分配业绩订单，已分配0条，无需分配0条，待分配8条（不含待付款、已关闭、退费中、已作废、欠费的订单）</span>
          </div>
          <div class="operations">
            <a-space>
              <a-button type="primary">
                批量分配
              </a-button>
              <a-button>导出数据</a-button>
            </a-space>
          </div>
        </div>
        <div class="table-content mt-2">
          <a-table
            :data-source="dataSource" :pagination="dataSource.length > 10" :columns="filteredColumns"
            :scroll="{ x: totalWidth }" size="small" :row-selection="{
              selectedRowKeys: state.selectedRowKeys,
              onChange: onSelectChange,
            }"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'orderNum'">
                <a class="text-#000 cursor-pointer">{{ record.orderNum }}</a>
              </template>
              <template v-if="column.key === 'student'">
                <div class="flex flex-items-center">
                  <div class="name mt-0">
                    <div class="text-#222">
                      {{ record.student.name }}
                    </div>
                    <div class="text-14px text-#888">
                      {{ record.student.phone }}
                    </div>
                  </div>
                </div>
              </template>
              <template v-if="column.key === 'allocationStatus'">
                <a-badge status="processing" :text="record.allocationStatus" />
              </template>
              <template v-if="column.key === 'performanceOwner'">
                {{ record.performanceOwner }}
              </template>
              <template v-if="column.key === 'allocationAmount'">
                {{ record.allocationAmount }}
              </template>
              <template v-if="column.key === 'handleContent'">
                {{ record.handleContent }}
              </template>
              <template v-if="column.key === 'orderType'">
                {{ record.orderType }}
              </template>
              <template v-if="column.key === 'totalAmount'">
                {{ record.totalAmount }}
              </template>
              <template v-if="column.key === 'createTime'">
                {{ record.createTime }}
              </template>
              <template v-if="column.key === 'action'">
                <a-space :size="12">
                  <a @click="handleAllocation(record)">分配业绩</a>
                  <a @click="handleNoAllocation(record)">无需分配</a>
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
  justify-content: center;

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
