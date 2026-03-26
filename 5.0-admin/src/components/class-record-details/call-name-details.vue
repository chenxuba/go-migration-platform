<script setup>
import { ExclamationCircleOutlined } from '@ant-design/icons-vue'

const displayArray = ref(['intention', 'followStatus', 'sex'])
const dataSource = ref([{ key: 1 }, { key: 2 }])
const openDrawer = ref(false)
function handleSeeStuData() {
  openDrawer.value = true
}
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
    title: '详情',
    key: 'detail',
    dataIndex: 'detail',
    fixed: 'left',
    width: 80,
  },
  {
    title: '学员身份',
    dataIndex: 'studentId',
    key: 'studentId',
    width: 120,
  },
  {
    title: '上课状态',
    dataIndex: 'classStatus',
    key: 'classStatus',
    width: 120,

  },
  {
    title: '扣费课程账户',
    dataIndex: 'deductionAccount',
    key: 'deductionAccount',
    width: 160,

  },
  {
    title: '课消方式',
    key: 'courseNotMethod',
    dataIndex: 'courseNotMethod',
    width: 110,
  },
  {
    title: '上课点名数量',
    dataIndex: 'classCallNum',
    key: 'classCallNum',
    width: 160,
  },
  {
    title: '消耗数量',
    dataIndex: 'useNum',
    key: 'useNum',
    width: 140,
  },
  {
    title: '拖欠数量',
    dataIndex: 'oweNum',
    key: 'oweNum',
    width: 140,
  },
  {
    title: '消耗学费',
    dataIndex: 'usePrice',
    key: 'usePrice',
    width: 140,
  },

  {
    title: '对内备注',
    dataIndex: 'externalRemarks',
    key: 'externalRemarks',
    width: 140,
  },
  {
    title: '对外备注',
    dataIndex: 'remarks',
    key: 'remarks',
    width: 140,
  },
  {
    title: '点名更新时间',
    key: 'callupdateTime',
    dataIndex: 'callupdateTime',
    width: 150,
    fixed: 'right',
  },

])
const rowSelection = {
  onChange: (selectedRowKeys, selectedRows) => {
    console.log(`selectedRowKeys: ${selectedRowKeys}`, 'selectedRows: ', selectedRows)
  },
}
const defaultStudentStatus = ref(1)
// 从本地存储读取已保存的列配置
const savedSelected = localStorage.getItem('call-name-details')
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
  localStorage.setItem('call-name-details', JSON.stringify(newVal))
}, { deep: true })
// 表格总宽度计算
const totalWidth = computed(() =>
  filteredColumns.value.reduce((acc, column) => acc + (column.width || 0), 0),
)
</script>

<template>
  <div>
    <!-- 学员筛选条件 -->
    <div class="filter-wrap bg-white  pl-3 pr-3 rounded-4">
      <all-filter
        :default-student-status="defaultStudentStatus" :display-array="displayArray" :is-quick-show="false"
        :is-show-search-input="true" search-label="学员姓名"
      />
    </div>
    <div class="student-list mt-3 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            共 <span class="text-4 mx-2 text-#06f">{{ dataSource.length }}</span> 条记录
          </div>
        </div>
        <div class="table-content mt-2">
          <a-table
            :data-source="dataSource" :pagination="dataSource.length > 10" :columns="filteredColumns"
            :scroll="{ x: totalWidth }" size="small"
          >
            <template #headerCell="{ column }">
              <template v-if="column.key === 'courseNotMethod'">
                <span class="mr-1">{{ column.title }}</span>
                <a-popover placement="top">
                  <template #title>
                    <div style="border-bottom:1px solid #eee;padding-bottom: 6px;">
                      课消方式
                    </div>
                  </template>
                  <template #content>
                    <div>【课消方式】课消方式决定了点名时的记录内容。</div>
                    <div>“按课时”：可以记录课时。</div>
                    <div>“按金额”：可以记录课时和金额。</div>
                    <div>“按时间”：可以记录课时。</div>
                  </template>
                  <ExclamationCircleOutlined />
                </a-popover>
              </template>
              <template v-if="column.key === 'useNum'">
                <span class="mr-1">{{ column.title }}</span>
                <a-popover placement="top">
                  <template #title>
                    <div style="border-bottom:1px solid #eee;padding-bottom: 6px;">
                      消耗数量
                    </div>
                  </template>
                  <template #content>
                    <div>【消耗数量】当次课程真实消耗了多少课时/金额。</div>
                  </template>
                  <ExclamationCircleOutlined />
                </a-popover>
              </template>
              <template v-if="column.key === 'oweNum'">
                <span class="mr-1">{{ column.title }}</span>
                <a-popover placement="top">
                  <template #title>
                    <div style="border-bottom:1px solid #eee;padding-bottom: 6px;">
                      拖欠数量
                    </div>
                  </template>
                  <template #content>
                    <div>【拖欠数量】该学员“剩余数量＜点名数量时”，会产生“拖欠数量”。</div>
                  </template>
                  <ExclamationCircleOutlined />
                </a-popover>
              </template>
              <template v-if="column.key === 'usePrice'">
                <span class="mr-1">{{ column.title }}</span>
                <a-popover placement="top">
                  <template #title>
                    <div style="border-bottom:1px solid #eee;padding-bottom: 6px;">
                      消耗学费
                    </div>
                  </template>
                  <template #content>
                    <div>【消耗学费】本次点名数量对应的学费（钱），即机构实际确认收入。</div>
                  </template>
                  <ExclamationCircleOutlined />
                </a-popover>
              </template>
            </template>
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'name'">
                <a-tooltip>
                  <template #title>
                    查看学员档案{{ record.a }}
                  </template>
                  <div class="flex cursor-pointer  hover" @click="handleSeeStuData()">
                    <img
                      width="36" height="36" class="mr-1" style="border-radius: 100%;"
                      src="https://pcsys.admin.ybc365.com/c04d0ea2-a8b0-4001-b19b-946a980cb726.png"
                      alt=""
                    >
                    <div class="name mt-0">
                      <div class="text-#222 name">
                        龙龙
                      </div>
                      <div class="text-3 text-#888 flex flex-items-center">
                        176****1636
                      </div>
                    </div>
                  </div>
                </a-tooltip>
              </template>
              <template v-if="column.key === 'detail'">
                <a-tooltip>
                  <template #title>
                    学员点名详情{{ record.a }}
                  </template>
                  <span class="cursor-pointer hover-text-#06f">详情</span>
                </a-tooltip>
              </template>
              <template v-if="column.key === 'studentId'">
                班级学员
              </template>
              <template v-if="column.key === 'classStatus'">
                <span class="bg-#e6f0ff text-#06f text-3 px2 py1 rounded-10 ml2">到课</span>
              </template>
              <template v-if="column.key === 'deductionAccount'">
                初级认知课
              </template>
              <template v-if="column.key === 'courseNotMethod'">
                按课时
              </template>
              <template v-if="column.key === 'classCallNum'">
                1课时
              </template>
              <template v-if="column.key === 'useNum'">
                1课时
              </template>
              <template v-if="column.key === 'oweNum'">
                -
              </template>
              <template v-if="column.key === 'usePrice'">
                ¥200.00
              </template>
              <template v-if="column.key === 'callupdateTime'">
                2024-12-23 13:22
              </template>
              <template v-if="column.key === 'externalRemarks'">
                -
              </template>
              <template v-if="column.key === 'remarks'">
                -
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>
    <student-info-drawer v-model:open="openDrawer" />
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
