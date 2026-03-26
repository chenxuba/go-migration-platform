<script setup>
import { DownOutlined } from '@ant-design/icons-vue'

const displayArray = ref(['intention', 'followStatus', 'sex', 'createUser', 'createTime', 'intentionCourse', 'reference'])
const dataSource = ref([
  {
    key: '1',
    name: '胡彦斌',
    phone: 17601241636,
    intentionCourse: '初级言语课、高级感统课、中级认知课',
    channelType: '外部渠道',
    channel: '抖音',
    teacher: '张晨',
    status: '已邀约',
    followed: '2025-03-31 17:09',
    nextTime: '2025-03-31 17:09',
    createTime: '2025-03-31 17:09',
    createUser: '张晨',
    putType: '否',
    putPeo: '-',
    birthday: '2022-09-23',
    wxchat: '1115009958',
    grade: '一年级',
    school: '上海市第一人民小学',
    address: '上海市杨浦区纪念路8号财大科技园区5号楼102A',
    IDcard1: 'CL202209229932',
    IDcard2: '37292520220922883X',
  },
])
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
    dataIndex: 'phone',
    width: 120,
    key: 'phone',
  },
  {
    title: '欠费金额',
    dataIndex: 'owePrice',
    key: 'owePrice',
    width: 120,
  },
  {
    title: '订单总金额',
    dataIndex: 'orderTotalPrice',
    key: 'orderTotalPrice',
    width: 120,
  },
  {
    title: '已支付金额',
    dataIndex: 'payPrice',
    key: 'payPrice',
    width: 120,
  },
  {
    title: '原订单号',
    dataIndex: 'originalOrderNum',
    key: 'originalOrderNum',
    width: 240,

  },
  {
    title: '创建时间',
    dataIndex: 'createTime',
    key: 'createTime',
    width: 160,
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    fixed: 'right',
    width: 100,
    required: true,
  },
])
const rowSelection = {
  onChange: (selectedRowKeys, selectedRows) => {
    console.log(`selectedRowKeys: ${selectedRowKeys}`, 'selectedRows: ', selectedRows)
  },
}
// 从本地存储读取已保存的列配置
const savedSelected = localStorage.getItem('owePrice')
const keysArray = allColumns.value
  .map(column => column?.key) // 可选链操作符
  .filter(key => typeof key !== 'undefined') // 过滤未定义的值
const initialSelectedValues = savedSelected
  ? JSON.parse(savedSelected)
  : keysArray

// 选中的列（初始化包含重要字段）
const selectedValues = ref(initialSelectedValues)
// 生成字段选择选项（排除操作列）
const columnOptions = computed(() =>
  allColumns.value
    .filter(col => col.key !== 'action')
    .map(col => ({
      id: col.key,
      value: col.title,
      disabled: col.required, // 禁用必选字段
    })),
)
// 过滤后的列（自动包含必选列）
const filteredColumns = computed(() => {
  const requiredColumns = allColumns.value.filter(col => col.required)
  const optionalColumns = allColumns.value
    .filter(col =>
      selectedValues.value.includes(col.key)
      && !col.required,
    )

  // 保持固定列顺序：left -> normal -> right
  return [
    ...requiredColumns.filter(col => col.fixed === 'left'),
    ...optionalColumns,
    ...requiredColumns.filter(col => col.fixed === 'right'),
  ]
})
// 强制包含必选字段的监听
watch(selectedValues, (newVal) => {
  const requiredKeys = allColumns.value
    .filter(col => col.required)
    .map(col => col.key)

  // 自动补全必选字段
  if (!requiredKeys.every(k => newVal.includes(k))) {
    selectedValues.value = Array.from(new Set([
      ...newVal.filter(v => !requiredKeys.includes(v)),
      ...requiredKeys,
    ]))
  }
}, { deep: true })
// 自动保存列配置到本地存储
watch(selectedValues, (newVal) => {
  localStorage.setItem('owePrice', JSON.stringify(newVal))
}, { deep: true })
// 表格总宽度计算
const totalWidth = computed(() =>
  filteredColumns.value.reduce((acc, column) => acc + (column.width || 0), 0),
)
</script>

<template>
  <div>
    <!-- 学员筛选条件 -->
    <div class="filter-wrap mt-2 bg-white  pl-3 pr-3 rounded-4">
      <all-filter :display-array="displayArray" :is-quick-show="false" :is-show-search-stu-phonefilter="true" />
    </div>
    <div class="student-list mt-2 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            当前共{{ dataSource.length }}名学员
          </div>
          <div class="edit flex">
            <a-button class="mr-2">
              消息记录
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
            <a-dropdown class="mr-2">
              <template #overlay>
                <a-menu>
                  <a-menu-item key="1">
                    微信提醒
                  </a-menu-item>
                  <a-menu-item key="2">
                    短信提醒
                  </a-menu-item>
                </a-menu>
              </template>
              <a-button>
                批量发送欠费提醒
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
            :data-source="dataSource" :pagination="dataSource.length > 10" :columns="filteredColumns" :row-selection="rowSelection"
            :scroll="{ x: totalWidth }" size="small"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'name'">
                <student-avatar
                  name="龙龙"
                  gender="男"
                  :show-age="false"
                  default-active-key="0"
                />
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
              <template v-if="column.key === 'owePrice'">
                ¥ 100.00 {{ record.a }}
              </template>
              <template v-if="column.key === 'orderTotalPrice'">
                ¥ 3000.00
              </template>
              <template v-if="column.key === 'payPrice'">
                ¥ 200.00
              </template>
              <template v-if="column.key === 'originalOrderNum'">
                <clamped-text :lines="2" text="20250409001456920078965" />
              </template>
              <template v-if="column.key === 'action'">
                <div class="actionbtn">
                  补费
                </div>
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>
  </div>
</template>

<style lang="less" scoped>
  .actionbtn{
    color: var(--pro-ant-color-primary);
    cursor: pointer;
  }
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
</style>
