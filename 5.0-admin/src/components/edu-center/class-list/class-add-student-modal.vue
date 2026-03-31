<script setup>
import { CloseOutlined, FilterFilled, SearchOutlined } from '@ant-design/icons-vue'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  /** 弹窗标题后缀：班级名称，如「感统集体课2班2」 */
  title: {
    type: String,
    default: '',
  },
  /** 顶部提示条中的课程/报读名称，默认同 title */
  lessonName: {
    type: String,
    default: '',
  },
})

const emit = defineEmits(['update:open'])

const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const bannerName = computed(() => {
  const n = (props.lessonName || props.title || '').trim()
  return n || '课程'
})

const userName = ref('')

/** 静态示例数据（后续接接口替换） */
function staticRows() {
  return [
    {
      id: 'static-1',
      displayName: '陈瑞',
      phone: '176****1111',
      courseAccount: '时段课程',
      age: '-',
      gender: '2',
      remainingQuantity: '8天',
      classStatus: '未分班',
      classStatusType: 2,
      disabled: false,
    },
  ]
}

const columns = [
  {
    title: '学员姓名',
    dataIndex: 'name',
    key: 'name',
    width: 200,
  },
  {
    title: '课程账户',
    dataIndex: 'courseAccount',
    key: 'courseAccount',
    width: 140,
  },
  {
    title: '年龄',
    dataIndex: 'age',
    key: 'age',
    width: 100,
  },
  {
    title: '性别',
    dataIndex: 'gender',
    key: 'gender',
    width: 100,
    filters: [
      { text: '男', value: '1' },
      { text: '女', value: '0' },
      { text: '未知', value: '2' },
    ],
    onFilter: (value, record) => record.gender === value,
  },
  {
    title: '剩余数量',
    dataIndex: 'remainingQuantity',
    key: 'remainingQuantity',
    width: 120,
  },
  {
    title: '分班状态',
    dataIndex: 'classStatus',
    key: 'classStatus',
    width: 140,
  },
]

const data = ref([])
const originalData = ref([])

const selectedRowKeys = ref([])
const selectedRows = ref([])

const rowSelection = computed(() => ({
  selectedRowKeys: selectedRowKeys.value,
  onChange: (keys, rows) => {
    selectedRowKeys.value = keys
    selectedRows.value = rows
  },
  getCheckboxProps: record => ({ disabled: record.disabled }),
}))

const minAge = ref(null)
const maxAge = ref(null)
const ageDropdownVisible = ref(false)
const isAgeFiltered = ref(false)

function resetAge() {
  minAge.value = null
  maxAge.value = null
}

function filterDataByAge() {
  if (minAge.value === null && maxAge.value === null) {
    data.value = [...originalData.value]
    return
  }
  data.value = originalData.value.filter((item) => {
    if (item.age === '-' || item.age == null || item.age === '')
      return false
    const ageNum = Number.parseInt(String(item.age), 10)
    if (Number.isNaN(ageNum))
      return false
    const ageInYears = ageNum / 12
    if (minAge.value != null && maxAge.value != null)
      return ageInYears >= minAge.value && ageInYears <= maxAge.value
    if (minAge.value != null)
      return ageInYears >= minAge.value
    if (maxAge.value != null)
      return ageInYears <= maxAge.value
    return true
  })
}

function handleAgeConfirm() {
  if (minAge.value == null && maxAge.value == null) {
    data.value = [...originalData.value]
    ageDropdownVisible.value = false
    isAgeFiltered.value = false
    return
  }
  let min = minAge.value
  let max = maxAge.value
  if (min != null && max != null && min > max) {
    [min, max] = [max, min]
    minAge.value = min
    maxAge.value = max
  }
  filterDataByAge()
  isAgeFiltered.value = true
  ageDropdownVisible.value = false
}

function handleSubmit() {
  // 静态阶段仅占位
  console.log('selected', toRaw(selectedRows.value))
}

function closeFun() {
  openModal.value = false
  selectedRowKeys.value = []
  selectedRows.value = []
  resetAge()
  isAgeFiltered.value = false
}

watch(
  () => props.open,
  (v) => {
    if (!v)
      return
    const rows = staticRows()
    data.value = [...rows]
    originalData.value = [...rows]
    userName.value = ''
    selectedRowKeys.value = []
    selectedRows.value = []
    resetAge()
    isAgeFiltered.value = false
  },
)

const filteredTableData = computed(() => {
  const q = userName.value.trim()
  if (!q)
    return data.value
  return data.value.filter(
    row => String(row.displayName || '').includes(q),
  )
})

function isRemainingZero(v) {
  if (v === 0 || v === '0')
    return true
  const s = String(v)
  return s === '0课时' || s === '0天'
}
</script>

<template>
  <a-modal
    v-model:open="openModal"
    centered
    class="add-class-student-modal"
    :keyboard="false"
    :closable="false"
    :mask-closable="false"
    :width="920"
    :footer="null"
    @cancel="closeFun"
  >
    <template #title>
      <div class="text-5 flex justify-between flex-center pr-1">
        <span class="font-500 text-#222">添加学员-{{ title || '班级' }}</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-4.5 text-#8c8c8c" />
          </template>
        </a-button>
      </div>
    </template>

    <a-alert
      class="add-student-banner border-0 rounded-0 !mb-0 pl-30px"
      type="info"
      show-icon
    >
      <template #message>
        <span class="text-#1677ff">以下为「{{ bannerName }}」报读下的学员</span>
      </template>
    </a-alert>

    <div class="modal-body-inner">
      <a-input
        v-model:value="userName"
        allow-clear
        placeholder="请输入学员姓名"
        class="student-search-input"
      >
        <template #prefix>
          <SearchOutlined class="text-#bfbfbf" />
        </template>
      </a-input>

      <div class="table-wrap mt-3">
        <a-table
          :columns="columns"
          :data-source="filteredTableData"
          row-key="id"
          :pagination="false"
          :row-selection="rowSelection"
          :scroll="{ x: 840 }"
        >
          <template #headerCell="{ column }">
            <template v-if="column.dataIndex === 'age'">
              <div class="flex items-center justify-between gap-1">
                <span>{{ column.title }}</span>
                <a-dropdown v-model:open="ageDropdownVisible" placement="bottomRight" :trigger="['click']">
                  <FilterFilled
                    class="text-12px cursor-pointer shrink-0"
                    :class="[
                      isAgeFiltered ? 'text-#1677ff' : 'text-#00000040 hover:text-#00000073',
                    ]"
                  />
                  <template #overlay>
                    <a-menu class="!p-0">
                      <div class="age-filter">
                        <div class="flex justify-between mb-2 text-3 text-#666">
                          <span>最小年龄（岁）</span>
                          <span>最大年龄（岁）</span>
                        </div>
                        <div class="flex items-center mb-3 gap-2">
                          <a-input-number
                            v-model:value="minAge"
                            :controls="false"
                            class="flex-1"
                            :min="0"
                            :max="100"
                            :precision="0"
                            placeholder="最小"
                          />
                          <span class="text-#d9d9d9">-</span>
                          <a-input-number
                            v-model:value="maxAge"
                            :controls="false"
                            class="flex-1"
                            :min="0"
                            :max="100"
                            :precision="0"
                            placeholder="最大"
                          />
                        </div>
                        <div class="flex justify-between gap-2">
                          <a-button
                            size="small"
                            type="link"
                            class="!px-1"
                            :disabled="minAge == null && maxAge == null"
                            @click="resetAge"
                          >
                            重置
                          </a-button>
                          <a-button size="small" type="primary" @click="handleAgeConfirm">
                            确定
                          </a-button>
                        </div>
                      </div>
                    </a-menu>
                  </template>
                </a-dropdown>
              </div>
            </template>
          </template>

          <template #bodyCell="{ column, record }">
            <template v-if="column.dataIndex === 'name'">
              <div class="flex items-center gap-3">
                <a-avatar
                  class="shrink-0 !bg-#1677ff"
                  :size="36"
                >
                  {{ (record.displayName || '?').slice(0, 1) }}
                </a-avatar>
                <div class="min-w-0">
                  <div class="text-#222 truncate">
                    {{ record.displayName }}
                  </div>
                  <div class="text-3 text-#8c8c8c truncate">
                    {{ record.phone || '-' }}
                  </div>
                </div>
              </div>
            </template>
            <template v-else-if="column.dataIndex === 'age'">
              {{ record.age === '-' ? '-' : `${record.age} 个月` }}
            </template>
            <template v-else-if="column.dataIndex === 'gender'">
              {{ record.gender === '1' ? '男' : record.gender === '0' ? '女' : '未知' }}
            </template>
            <template v-else-if="column.dataIndex === 'remainingQuantity'">
              <span
                :class="isRemainingZero(record.remainingQuantity) ? 'text-#ff4d4f' : 'text-#222'"
              >
                {{ record.remainingQuantity }}
              </span>
            </template>
            <template v-else-if="column.dataIndex === 'classStatus'">
              <div class="flex items-center">
                <span
                  class="inline-block h-6px w-6px rounded-full shrink-0"
                  :class="record.classStatusType === 1 ? 'bg-#1677ff' : 'bg-#fa8c16'"
                />
                <span
                  class="text-3.5 px-1"
                  :class="record.classStatusType === 1 ? 'text-#1677ff' : ' text-#fa8c16'"
                >
                  {{ record.classStatus }}
                </span>
              </div>
            </template>
          </template>
        </a-table>
      </div>
    </div>

    <div class="modal-footer-bar flex items-center justify-between border-t border-#f0f0f0 bg-white px-6 py-3">
      <span class="text-#222">已选择: {{ selectedRowKeys.length }} 人</span>
      <a-space :size="12">
        <a-button @click="closeFun">
          取消
        </a-button>
        <a-button type="primary" :disabled="selectedRowKeys.length === 0" @click="handleSubmit">
          确定
        </a-button>
      </a-space>
    </div>
  </a-modal>
</template>

<style lang="less" scoped>
.close-btn {
  &:hover {
    background: transparent;
  }
}

.modal-body-inner {
  padding: 16px 24px 8px;
}

.student-search-input {
  height: 44px;
  border-radius: 10px;

 
}

.table-wrap {
  :deep(.ant-table-thead > tr > th) {
    background: #fafafa;
    color: #595959;
    font-weight: 500;
  }
}

.age-filter {
  padding: 12px;
  min-width: 260px;
}
</style>

<style lang="less">
.add-class-student-modal .ant-modal-content {
  padding: 0;
  overflow: hidden;
}

.add-class-student-modal .ant-modal-header {
  padding: 16px 24px;
  margin-bottom: 0;
  border-bottom: 1px solid #f0f0f0;
}

.add-class-student-modal .ant-modal-body {
  padding: 0;
}

.add-student-banner.ant-alert {
  margin: 0;
  border-radius: 0;
  border: none;
  background: #e6f4ff;
}

.add-student-banner .ant-alert-icon {
  color: #1677ff;
}

/* 选中行无高亮底；hover 仅极浅灰（覆盖 Ant Design 默认选中蓝、hover 蓝） */
.add-class-student-modal .ant-table-tbody > tr > td {
  background: #fff !important;
}

.add-class-student-modal .ant-table-tbody > tr:hover > td {
  background: #f7f7f7 !important;
}

.add-class-student-modal .ant-table-tbody > tr.ant-table-row-selected > td {
  background: #fff !important;
}

.add-class-student-modal .ant-table-tbody > tr.ant-table-row-selected:hover > td {
  background: #f7f7f7 !important;
}
</style>
