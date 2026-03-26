<script setup>
import { CloseOutlined } from '@ant-design/icons-vue'

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
const columns = ref([
  {
    title: '学员姓名',
    dataIndex: 'name',
  },
  {
    title: '联系电话',
    dataIndex: 'phone',
  },
  {
    title: '缺课时间',
    dataIndex: 'absentTime',
  },
])
const data = ref([
  {
    id: 1,
    name: '张三',
    phone: '12345678901',
    absentTime: '2021-01-01 12:00-13:00',
    disabled: true,
  },
  {
    id: 2,
    name: '李四',
    phone: '12345678902',
    absentTime: '2021-01-02 12:00-13:00',
  },
  {
    id: 3,
    name: '王五',
    phone: '12345678903',
    absentTime: '2021-01-03 12:00-13:00',
  },
  {
    id: 4,
    name: '赵六',
    phone: '12345678904',
    absentTime: '2021-01-04 12:00-13:00',
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
</script>

<template>
  <a-modal
    v-model:open="openModal" centered class="modal-content-box" :keyboard="false" :closable="false"
    :mask-closable="false" :width="800" @cancel="closeFun"
  >
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>{{ title }}</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <div class="contenter scrollbar">
      <a-input v-model:value="userName" placeholder="搜索学员姓名" class="h-48px rounded-12px">
        <template #prefix>
          <img
            src="https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12181/static/magnifying.2bcc08ab.svg"
            alt="" class="pr-6px mt--2px"
          >
        </template>
      </a-input>
      <!-- 表格 -->
      <!-- 补课学员的字段 ： 学员姓名	联系电话	缺课时间 -->
      <!-- 临时学员的字段 ： 学员姓名	联系电话 -->
      <!-- 试听学员的字段 ： 学员姓名	联系电话 -->
      <!-- 多选 -->
      <a-table
        :columns="columns" :data-source="data" row-key="id" class="mt-12px" :pagination="false"
        :row-selection="rowSelection"
      >
        <template #bodyCell="{ column, record }">
          <div v-if="column.dataIndex === 'name'">
            {{ record.name }}
          </div>
          <div v-if="column.dataIndex === 'phone'">
            {{ record.phone }}
          </div>
          <!-- 缺课时段只有 补课学员才有 -->
          <div v-if="column.dataIndex === 'absentTime'">
            {{ record.absentTime }}
          </div>
        </template>
      </a-table>
    </div>
    <template #footer>
      <div class="flex justify-between">
        <span>已勾选{{ selectedRows.length }}人</span>
        <div>
          <a-button danger ghost @click="closeFun">
            关闭
          </a-button>
          <a-button type="primary" ghost :disabled="selectedRows.length === 0" @click="handleSubmit">
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
