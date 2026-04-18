<script setup>
import { DownOutlined, ExclamationCircleFilled, ExclamationCircleOutlined } from '@ant-design/icons-vue'

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
    title: '学员/电话',
    dataIndex: 'name',
    key: 'name',
    fixed: 'left',
    width: 160,
    required: true, // 新增必选标识
  },
  {
    title: '未记录次数',
    dataIndex: 'notRecordedNumber',
    width: 120,
    key: 'notRecordedNumber',
  },
  {
    title: '旷课次数',
    dataIndex: 'truancyNumber',
    key: 'truancyNumber',
    width: 120,
  },
  {
    title: '请假次数',
    dataIndex: 'leaveNumber',
    key: 'leaveNumber',
    width: 120,
  },
  {
    title: '距离上次上课天数',
    dataIndex: 'daysSinceLastClass',
    key: 'daysSinceLastClass',
    width: 150,
  },
  {
    title: '提醒标记',
    dataIndex: 'reminderMark',
    key: 'reminderMark',
    width: 120,
  },
  {
    title: '班主任',
    dataIndex: 'headTeacher',
    key: 'headTeacher',
    width: 120,
  },
  {
    title: '销售',
    key: 'sale',
    dataIndex: 'sale',
    width: 120,
  },
  {
    title: '上次提醒时间',
    dataIndex: 'lastReminderTime',
    key: 'lastReminderTime',
    width: 160,
  },
  {
    title: '上次提醒操作人',
    dataIndex: 'lastReminderUser',
    key: 'lastReminderUser',
    width: 130,
  },
])

const rowSelection = {
  onChange: (selectedRowKeys, selectedRows) => {
    console.log(`selectedRowKeys: ${selectedRowKeys}`, 'selectedRows: ', selectedRows)
  },
}
// 从本地存储读取已保存的列配置
const savedSelected = localStorage.getItem('missSchool')
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
  localStorage.setItem('missSchool', JSON.stringify(newVal))
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
            <a-dropdown class="mr-2">
              <template #overlay>
                <a-menu>
                  <a-menu-item key="1">
                    短信提醒
                  </a-menu-item>
                  <a-menu-item key="2">
                    微信提醒
                  </a-menu-item>
                </a-menu>
              </template>
              <a-button>
                批量提醒
                <DownOutlined :style="{ fontSize: '10px' }" />
              </a-button>
            </a-dropdown>
            <a-dropdown class="mr-2">
              <template #overlay>
                <a-menu>
                  <a-menu-item key="1">
                    无需提醒
                  </a-menu-item>
                  <a-menu-item key="2">
                    需提醒
                  </a-menu-item>
                </a-menu>
              </template>
              <a-button>
                批量编辑提醒标记
                <DownOutlined :style="{ fontSize: '10px' }" />
              </a-button>
            </a-dropdown>
            <a-button class="mr-2">
              导出
            </a-button>
            <!-- 自定义字段 -->
            <customize-code
              v-model:checked-values="selectedValues" :options="columnOptions" :total="allColumns.length"
              :num="selectedValues.length"
            />
          </div>
        </div>
        <div class="table-content mt-2">
          <div class="tip flex justify-between">
            <div>
              <ExclamationCircleFilled :style="{ fontSize: '14px', color: '#0066ff' }" />
              家缺课学员条件设置：距离最近一次上课为到课状态的上课日期大于 30 天
            </div>
            <div class="cursor-pointer">
              缺课条件配置
            </div>
          </div>
          <a-table
            :data-source="dataSource" :pagination="dataSource.length > 10" :columns="filteredColumns"
            :row-selection="rowSelection" :scroll="{ x: totalWidth }" size="small"
          >
            <template #headerCell="{ column }">
              <template v-if="column.key === 'notRecordedNumber'">
                <span class="mr-1">{{ column.title }}</span>
                <a-tooltip color="#666">
                  <template #title>
                    仅展示30天内的未记录次数
                  </template>
                  <ExclamationCircleOutlined />
                </a-tooltip>
              </template>
              <template v-if="column.key === 'truancyNumber'">
                <span class="mr-1">{{ column.title }}</span>
                <a-tooltip color="#666">
                  <template #title>
                    仅展示30天内的旷课次数
                  </template>
                  <ExclamationCircleOutlined />
                </a-tooltip>
              </template>
              <template v-if="column.key === 'leaveNumber'">
                <span class="mr-1">{{ column.title }}</span>
                <a-tooltip color="#666">
                  <template #title>
                    仅展示30天内的请假次数
                  </template>
                  <ExclamationCircleOutlined />
                </a-tooltip>
              </template>
            </template>
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'name'">
                <student-avatar name="龙龙" gender="男" phone="17601241636" default-active-key="0" />
              </template>
              <!-- 未记录次数 -->
              <template v-if="column.key === 'notRecordedNumber'">
                1{{ record.a }}
              </template>

              <!-- 旷课次数 -->
              <template v-if="column.key === 'truancyNumber'">
                2
              </template>

              <!-- 请假次数 -->
              <template v-if="column.key === 'leaveNumber'">
                0
              </template>

              <!-- 距离上次上课天数 -->
              <template v-if="column.key === 'daysSinceLastClass'">
                2天
              </template>

              <!-- 提醒标记 -->
              <template v-if="column.key === 'reminderMark'">
                无
              </template>

              <!-- 班主任 -->
              <template v-if="column.key === 'headTeacher'">
                张晨
              </template>

              <!-- 销售 -->
              <template v-if="column.key === 'sale'">
                陈瑞生
              </template>

              <!-- 上次提醒时间 -->
              <template v-if="column.key === 'lastReminderTime'">
                2025-02-12 12:22
              </template>

              <!-- 上次提醒操作人 -->
              <template v-if="column.key === 'lastReminderUser'">
                陈瑞生
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

.tip {
  padding: 10px 24px 10px 14px;
  background: #e6f0ff;
  color: #333;
  color: var(--pro-ant-color-primary);
}
</style>
