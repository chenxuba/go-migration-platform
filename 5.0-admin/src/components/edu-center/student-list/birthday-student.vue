<script setup>
import { DownOutlined } from '@ant-design/icons-vue'

const displayArray = ref(['intention', 'followStatus', 'sex', 'createUser', 'createTime', 'intentionCourse', 'reference'])
const allColumns = ref([
  {
    title: '学员/性别/年龄',
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
    title: '学员状态',
    dataIndex: 'studentStatus',
    key: 'studentStatus',
    width: 120,
  },
  {
    title: '生日日期',
    key: 'birthday',
    dataIndex: 'birthday',
    width: 120,
  },
])
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
// 从本地存储读取已保存的列配置
const savedSelected = localStorage.getItem('birthdayStudent')
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
  localStorage.setItem('birthdayStudent', JSON.stringify(newVal))
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
      <all-filter :display-array="displayArray" :is-quick-show="false" :is-show-search-stu-phone="false" />
    </div>
    <div class="student-list mt-2 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            当前共{{ dataSource.length }}名学员 <span class="text-#0066ff">（当前为未来2个月的生日学员）</span>
          </div>
          <div class="edit flex">
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
          </div>
        </div>
        <div class="table-content mt-2">
          <a-table :data-source="dataSource" :pagination="dataSource.length > 10" :columns="filteredColumns" :scroll="{ x: totalWidth }" size="small">
            <template #bodyCell="{ column, record }">
              <!-- 学员/性别列保持不变 -->
              <template v-if="column.key === 'name'">
                <student-avatar
                  name="龙龙"
                  gender="男"
                  :show-age="false"
                  default-active-key="0"
                />
              </template>

              <!-- 联系电话列保持不变 -->
              <template v-if="column.key === 'phone'">
                <div class="name">
                  <div class="text-#222">
                    爸爸
                  </div>
                  <div class="text-3 text-#666">
                    176​**​​**​1636
                  </div>
                </div>
              </template>

              <template v-if="column.key === 'studentStatus'">
                <div class="flex flex-items-center studentStatus">
                  <span class="dot" />
                  <span>在读学员</span>
                </div>
              </template>

              <!-- 新增创建时间列 -->
              <template v-if="column.key === 'createTime'">
                {{ record.createTime || '2024-03-23 12:32' }}
              </template>

              <!-- 新增操作列 -->
              <template v-if="column.key === 'action'">
                <a-space>
                  <a-button type="link" size="small">
                    去分班
                  </a-button>
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
    margin-right: 2px;
    background: var(--pro-ant-color-primary);
  }
}
</style>
