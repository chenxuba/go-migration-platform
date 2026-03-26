<script setup>
import { CloseOutlined, FilterFilled } from '@ant-design/icons-vue'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  title: {
    type: String,
    default: '补课学员',
  },
})
const emit = defineEmits(['update:open'])
// 处理双向绑定
const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const userName = ref('')
// 学员姓名	课程账户
// 年龄
// 性别
// 剩余数量
// 分班状态

const columns = ref([
  {
    title: '学员姓名',
    dataIndex: 'name',
    key: 'name',
  },
  {
    title: '课程账户',
    dataIndex: 'courseAccount',
    key: 'courseAccount',
  },
  {
    title: '年龄',
    dataIndex: 'age',
    key: 'age',
  },
  {
    title: '性别',
    dataIndex: 'gender',
    key: 'gender',
    filters: [
      {
        text: '男',
        value: '1',
      },
      {
        text: '女',
        value: '0',
      },
      {
        text: '未知',
        value: '2',
      },
    ],
    onFilter: (value, record) => {
      return record.gender === value
    },
  },
  {
    title: '剩余数量',
    dataIndex: 'remainingQuantity',
    key: 'remainingQuantity',
  },
  {
    title: '分班状态',
    dataIndex: 'classStatus',
    key: 'classStatus',
  },
])
const data = ref([
  {
    id: 1,
    name: '张三',
    courseAccount: '视知觉训练',
    age: '9',
    gender: '1',
    remainingQuantity: '0',
    classStatus: '已分班',
    classStatusType: 1,
    disabled: true,
  },
  {
    id: 2,
    name: '李四',
    courseAccount: '视知觉训练',
    age: '24',
    gender: '2',
    remainingQuantity: '10',
    classStatus: '未分班',
    classStatusType: 2,
    disabled: false,
  },
])

function handleSubmit() {
  console.log(selectedRows.value)
}
function closeFun() {
  openModal.value = false
  selectedRowKeys.value = []
  selectedRows.value = []
}
// 多选相关状态
const selectedRowKeys = ref([])
const selectedRows = ref([])
// 多选配置
const rowSelection = computed(() => ({
  selectedRowKeys: selectedRowKeys.value,
  onChange: (keys, rows) => {
    selectedRowKeys.value = keys
    selectedRows.value = rows
    console.log('选中行数据:', toRaw(selectedRowKeys.value))
  },
  // 可选配置项
  getCheckboxProps: record => ({ disabled: record.disabled }),
  // type: 'checkbox' // 默认是多选，设为 radio 可切换单选模式
}))

const minAge = ref(null)
const maxAge = ref(null)
const originalData = ref([...data.value]) // 保存原始数据
const ageDropdownVisible = ref(false) // 控制年龄筛选下拉框的显示和隐藏
const isAgeFiltered = ref(false) // 控制图标颜色

function resetAge() {
  minAge.value = null
  maxAge.value = null
  // 移除立即重置数据的操作，仅在确认时才应用筛选
}

function handleAgeConfirm() {
  if (minAge.value == null && maxAge.value == null) {
    // 重置为初始数据
    data.value = [...originalData.value]
    ageDropdownVisible.value = false
    isAgeFiltered.value = false // 重置图标颜色
    return
  }
  let min = minAge.value
  let max = maxAge.value
  if (min != null && max != null && min > max) {
    // 交换
    [min, max] = [max, min]
    minAge.value = min
    maxAge.value = max
  }
  // 实现年龄筛选逻辑
  filterDataByAge()

  // 设置图标颜色
  isAgeFiltered.value = true

  // 隐藏下拉菜单
  ageDropdownVisible.value = false
}

// 添加年龄筛选函数
function filterDataByAge() {
  // 如果最小和最大年龄都为空，恢复原始数据
  if (minAge.value === null && maxAge.value === null) {
    data.value = [...originalData.value]
    return
  }

  // 筛选符合条件的数据
  data.value = originalData.value.filter((item) => {
    const ageInMonths = Number.parseInt(item.age)
    const ageInYears = ageInMonths / 12 // 假设数据是以月为单位

    // 根据输入的最小/最大年龄筛选
    if (minAge.value !== null && maxAge.value !== null) {
      return ageInYears >= minAge.value && ageInYears <= maxAge.value
    }
    else if (minAge.value !== null) {
      return ageInYears >= minAge.value
    }
    else if (maxAge.value !== null) {
      return ageInYears <= maxAge.value
    }

    return true
  })
}
</script>

<template>
  <a-modal
    v-model:open="openModal" centered class="modal-content-box" :keyboard="false" :closable="false"
    :mask-closable="false" :width="800" @cancel="closeFun"
  >
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>添加学员-{{ title }}</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <a-alert message="以下为&quot;视知觉训练&quot;报读下的学员" type="info" class="border-0 rounded-0 text-#06f" show-icon />
    <div class="contenter scrollbar">
      <a-input v-model:value="userName" placeholder="搜索学员姓名" class="h-48px rounded-12px">
        <template #prefix>
          <img
            src="https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12181/static/magnifying.2bcc08ab.svg"
            alt="" class="pr-6px mt--2px"
          >
        </template>
      </a-input>
      <a-table
        :columns="columns" :data-source="data" row-key="id" class="mt-12px" :pagination="false"
        :row-selection="rowSelection"
      >
        <template #headerCell="{ column }">
          <div v-if="column.dataIndex === 'age'" class="flex justify-between">
            <span>{{ column.title }}</span>
            <span>

              <a-dropdown v-model:open="ageDropdownVisible" placement="bottomRight" :trigger="['click']">
                <FilterFilled
                  class="text-12px cursor-pointer" :class="[
                    isAgeFiltered ? 'text-#06f' : 'text-#0000004a hover:text-#0000006a',
                  ]"
                />
                <template #overlay>
                  <a-menu>
                    <div class="age-filter">
                      <div class="flex justify-between mb-2">
                        <span>最小年龄（岁）</span>
                        <span>最大年龄（岁）</span>
                      </div>
                      <div class="flex items-center mb-2">
                        <a-input-number
                          v-model:value="minAge" :controls="false" class="w-25 mr-2" :min="0" :max="100"
                          :precision="0" placeholder="最小"
                        />
                        <span>-</span>
                        <a-input-number
                          v-model:value="maxAge" :controls="false" class="w-25 ml-2" :min="0" :max="100"
                          :precision="0" placeholder="最大"
                        />
                      </div>
                      <div class="flex justify-between px4 b-btn">
                        <a-button
                          size="small" type="text" class="mr-2 w-18 text-#06f reset-btn"
                          :disabled="minAge == null && maxAge == null" @click="resetAge"
                        >重置</a-button>
                        <a-button size="small" class="w-18" type="primary" @click="handleAgeConfirm">确定</a-button>
                      </div>
                    </div>
                  </a-menu>
                </template>
              </a-dropdown>
            </span>
          </div>
        </template>
        <template #bodyCell="{ column, record }">
          <div v-if="column.dataIndex === 'name'">
            {{ record.name }}
          </div>
          <!-- 年龄 -->
          <div v-if="column.dataIndex === 'age'">
            {{ record.age }} 个月
          </div>
          <!-- 性别 -->
          <div v-if="column.dataIndex === 'gender'">
            {{ record.gender === '1' ? '男' : record.gender === '2' ? '女' : '未知' }}
          </div>
          <!-- 剩余课时 -->
          <div v-if="column.dataIndex === 'remainingQuantity'">
            <span :class="record.remainingQuantity > 0 ? 'text-#666' : 'text-#ff3000'">
              {{ record.remainingQuantity }}课时
            </span>
          </div>
          <!-- 分班状态 -->
          <div v-if="column.dataIndex === 'classStatus'">
            <span v-if="record.classStatusType === 1" class="bg-#e6f0ff text-#06f text-3 px3 py2px rounded-10">{{
              record.classStatus }}</span>
            <span v-if="record.classStatusType === 2" class="bg-#fff5e6 text-#f90 text-3 px3 py2px rounded-10">{{
              record.classStatus }}</span>
          </div>
        </template>
      </a-table>
    </div>
    <template #footer>
      <div class="flex justify-between">
        <span>已勾选{{ selectedRows.length }}人</span>
        <div>
          <a-button danger ghost @click="closeFun">
            取消
          </a-button>
          <a-button type="primary" :disabled="selectedRows.length === 0" @click="handleSubmit">
            确定
          </a-button>
        </div>
      </div>
    </template>
  </a-modal>
</template>

<style lang="less" scoped>
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

.contenter {
  padding: 24px;
  padding-top: 12px;
}

.age-filter {
  display: inline-block;
  padding: 8px;
  border-radius: 6px;
  min-width: 100px;
  color: #666;
}

/* 自定义镂空样式 */
.custom-radio ::v-deep(.ant-radio-wrapper:hover .ant-radio),
.custom-radio ::v-deep(.ant-radio:hover .ant-radio-inner),
.custom-radio ::v-deep(.ant-radio-input:focus + .ant-radio-inner) {
  border-color: var(--pro-ant-color-primary);
}

.custom-radio ::v-deep(.ant-radio-inner) {
  background-color: transparent;
  border-color: #d9d9d9;
}

.custom-radio ::v-deep(.ant-radio-checked .ant-radio-inner) {
  background-color: transparent;
  border-color: var(--pro-ant-color-primary);
}

.custom-radio ::v-deep(.ant-radio-inner::after) {
  background-color: var(--pro-ant-color-primary);
  transform: scale(0.5);
}
</style>

<style>
.modal-content-box .ant-modal-header {
  padding: 10px 16px !important;
  margin-bottom: 0;
}

.modal-content-box .ant-modal-body {
  padding: 0 !important;
}
</style>
