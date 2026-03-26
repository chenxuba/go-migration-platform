<script setup>
import { CloseOutlined } from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'
import EditRowRollNameModal from './edit-row-roll-name-modal.vue'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  defaultTitle: {
    type: String,
    default: '',
  },
})
const emit = defineEmits(['update:open'])
// 处理双向绑定
const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

function closeFun() {
  openModal.value = false
}
const userName = ref('')
const columns = [
  {
    title: '学员姓名',
    dataIndex: 'name',
    key: 'name',
    width: 200,
  },
  {
    title: '联系电话',
    dataIndex: 'phone',
    key: 'phone',
    width: 100,
  },
]
const data = [
  {
    id: 1,
    name: '张三',
    phone: '17601241636',
  },
  {
    id: 2,
    name: '李四',
    phone: '17601241636',
    disabled: true,
  },
  {
    id: 3,
    name: '王五',
    phone: '17601241636',
  },
]
const editRowRollNameModals = ref(false)
// 点击行事件处理
function handleRowClick(record) {
  return {
    onClick: () => {
      // 如果行被禁用，直接返回
      if (record.disabled)
        return message.warning('该学生已在日程中')

      // 非禁用行：获取 id 并处理业务逻辑
      console.log('点击行ID:', record.id)
      // 这里可以调用其他方法，例如跳转页面、打开弹窗等
      // handleOpenDetail(record.id);
      editRowRollNameModals.value = true
    },
  }
}
</script>

<template>
  <a-modal
    v-model:open="openModal" centered class="modal-content-box" :footer="false" :keyboard="false"
    :closable="false" :mask-closable="false" :width="800"
  >
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>{{ defaultTitle }}</span>
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
      <!-- 学员姓名	联系电话 -->
      <a-table
        :columns="columns" :data-source="data" row-key="id" class="mt-12px" :pagination="false"
        :row-class-name="(record) => record.disabled ? 'disabled-row' : 'row-hover'" :custom-row="handleRowClick"
      >
        <template #bodyCell="{ column, record }">
          <div v-if="column.dataIndex === 'name'">
            <img
              src="https://cdn.schoolpal.cn/schoolpal/next-erp/avator_male.png" alt=""
              class="w-32px h-32px rounded-full mr-4px"
            >
            {{ record.name }}
          </div>
          <div v-if="column.dataIndex === 'phone'">
            <!-- 脱敏 -->
            {{ record.phone.replace(/(\d{3})(\d{4})(\d{4})/, '$1****$2') }}
          </div>
        </template>
      </a-table>
    </div>
    <!-- 点击行触发 -->
    <EditRowRollNameModal v-model:open="editRowRollNameModals" />
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
  padding: 12px 24px 24px;
}

/* 禁用行的样式 */
:deep(.disabled-row) {
  background-color: #eee;
  /* 背景颜色 */
  opacity: 0.8;
  /* 透明度 */
  cursor: not-allowed;
  /* 鼠标样式 */
}

:deep(.row-hover) {
  cursor: pointer;
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
