<script setup>
import { CloseOutlined, InfoCircleOutlined } from '@ant-design/icons-vue'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  // 渠道分类列表
  channelCategoryList: {
    type: Array,
    default: () => [],
  },
  // 已选渠道数据
  selectedRowsData: {
    type: Array,
    default: () => [],
  },
  // 渠道分类id
  propCategoryId: {
    type: Number,
    default: 0,
  },
  // 弹窗类型
  modalType: {
    type: String,
    default: '',
  },
})

const emit = defineEmits(['update:open', 'update:adjustChannel'])

const categoryId = ref(0)

// 处理双向绑定
const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})
// 监听openModal，如果打开则重置categoryId
watch(openModal, (newVal) => {
  if (newVal) {
    categoryId.value = props.propCategoryId
  }
})

function closeFun() {
  openModal.value = false
}
function adjustChannel() {
  console.log('categoryId: ', categoryId.value)
  // emit
  emit('update:adjustChannel', categoryId.value)
}
</script>

<template>
  <a-modal
    v-model:open="openModal" centered class="modal-content-box" :keyboard="false" :closable="false"
    :mask-closable="false" :width="700"
  >
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>{{ modalType === 'adjust' ? '批量调整渠道' : '请选择' }}</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <a-alert v-if="modalType === 'adjust'" type="info" :message="`已选 ${selectedRowsData.length} 个渠道，请选择`" show-icon class="rounded-0 border-none text-#06f" />
    <div class="contenter scrollbar">
      <!-- 单选组 -->
      <div class="single-select-group">
        <a-radio-group v-model:value="categoryId" class="flex flex-col custom-radio">
          <a-radio :value="0" class="mb-14px pb-14px border-b border-#eee border-solid border-t-0 border-x-0">
            将已选渠道移出对应分类
            <a-tooltip>
              <template #title>
                将已选渠道移出对应分类
              </template>
              <InfoCircleOutlined class="text-#06f" />
            </a-tooltip>
          </a-radio>
          <a-radio
            v-for="item in channelCategoryList" :key="item.id" :value="item.id"
            class="mb-14px pb-14px border-b border-#eee border-solid border-t-0 border-x-0"
          >
            分类：{{ item.categoryName }}
          </a-radio>
        </a-radio-group>
      </div>
    </div>
    <template #footer>
      <div class="footer">
        <a-button @click="closeFun">
          取消
        </a-button>
        <a-button type="primary" @click="adjustChannel">
          确认调整
        </a-button>
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
  height: calc(100vh - 300px);
  overflow-y: auto;
  padding: 12px 24px;

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
