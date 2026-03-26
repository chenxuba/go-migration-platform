<script setup>
import {
  DownOutlined,
} from '@ant-design/icons-vue'
import { useTableColumns } from '@/composables/useTableColumns'

const displayArray = ref(['doYouSchedule', 'openClassStatus'])

const dataSource = ref([{ key: 1 }, { key: 2 }])
const allColumns = ref([
  {
    title: '学杂费名称',
    dataIndex: 'name',
    key: 'name',
    fixed: 'left',
    width: 180,
    required: true, // 新增必选标识
  },
  {
    title: '售卖状态',
    dataIndex: 'sellStatus',
    key: 'sellStatus',
    width: 140,
  },
  {
    title: '是否开启微校售卖',
    dataIndex: 'isOpenSamllSchoolSell',
    key: 'isOpenSamllSchoolSell',
    width: 140,
  },
  {
    title: '售卖价格',
    key: 'price',
    dataIndex: 'price',
    width: 140,
  },
  {
    title: '总销量',
    dataIndex: 'totalSales',
    key: 'totalSales',
    width: 140,
  },
  {
    title: '更新时间',
    key: 'updateTime',
    dataIndex: 'updateTime',
    width: 140,
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
    console.log(
      `selectedRowKeys: ${selectedRowKeys}`,
      'selectedRows: ',
      selectedRows,
    )
  },
}
const { selectedValues, columnOptions, filteredColumns, totalWidth }
  = useTableColumns({
    storageKey: 'tuition-adn-fees', // 本地存储键名
    allColumns, // 原始列配置
    excludeKeys: ['action'], // 需要排除的列键
  })
</script>

<template>
  <div>
    <!-- 学员筛选条件 -->
    <div class="filter-wrap bg-white pl-3 pr-3 rounded-4 rounded-tl-0 rounded-tr-0">
      <all-filter
        :display-array="displayArray"
        :is-quick-show="false"
        :is-show-search-stu-phone="false"
        :is-show-search-input="true"
        search-label="学杂费名称"
      />
    </div>
    <div class="student-list mt-2 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            总计{{ dataSource.length }}条学杂费，1 条在售卖
          </div>
          <div class="edit flex">
            <a-button class="mr-2">
              导出数据
            </a-button>
            <a-dropdown class="mr-2">
              <template #overlay>
                <a-menu>
                  <a-menu-item key="1">
                    批量删除
                  </a-menu-item>
                  <a-menu-item key="2">
                    批量售卖
                  </a-menu-item>
                  <a-menu-item key="3">
                    批量停售
                  </a-menu-item>
                </a-menu>
              </template>
              <a-button>
                批量操作
                <DownOutlined :style="{ fontSize: '10px' }" />
              </a-button>
            </a-dropdown>
            <a-button type="primary" class="mr-2">
              创建学杂费
            </a-button>
            <!-- 自定义字段 -->
            <!-- <customize-code
              v-model:checkedValues="selectedValues"
              :options="columnOptions"
              :total="allColumns.length - 1"
              :num="selectedValues.length"
            /> -->
          </div>
        </div>
        <div class="table-content mt-2">
          <a-table
            :data-source="dataSource"
            :pagination="dataSource.length > 10"
            :columns="filteredColumns"
            :row-selection="rowSelection"
            :scroll="{ x: totalWidth }"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'name'">
                初级感统课 {{ record.a }}
              </template>
              <template v-if="column.key === 'price'">
                ¥500.00
              </template>

              <template v-if="column.key === 'sellStatus'">
                <span class="bg-#e6ffec text-#0c3 text-3 px3 py0.8 rounded-10">在售</span>
              </template>

              <template v-if="column.key === 'isOpenSamllSchoolSell'">
                <div class="studentStatus">
                  <span class="dot bg-#ccc" />
                  <span class="text-#999">否</span>
                </div>
              </template>
              <template v-if="column.key === 'totalSales'">
                14
              </template>
              <template v-if="column.key === 'updateTime'">
                2024-07-24 10:45
              </template>
              <template v-if="column.key === 'subject'">
                自费
              </template>
              <template v-else-if="column.key === 'action'">
                <span class="flex action">
                  <a class="mr-3">编辑{{ record.a }}</a>
                  <a class="mr-3">预览</a>
                  <a class="font500">停售</a>
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
  }
}
</style>
