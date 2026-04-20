<script setup lang="ts">
import { CloseOutlined } from '@ant-design/icons-vue'

const props = withDefaults(defineProps<{
  open?: boolean
}>(), {
  open: false,
})

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'confirm'): void
}>()

const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

function closeModal() {
  openModal.value = false
}

function handleConfirm() {
  emit('confirm')
  closeModal()
}
</script>

<template>
  <a-modal
    v-model:open="openModal"
    class="modal-content-box"
    :keyboard="false"
    :closable="false"
    :mask-closable="false"
    :width="800"
    centered
  >
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>创建未排课点名</span>
        <a-button type="text" class="close-btn" @click="closeModal">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>

    <div class="create-roll-call-modal__body scrollbar">
      <div class="create-roll-call-modal__section">
        <div class="create-roll-call-modal__label">
          未排课点名弹窗
        </div>
        <div class="create-roll-call-modal__desc">
          已按 `modal-demo.vue` 的样式新建组件，并接入当前页面按钮点击弹出。
        </div>
      </div>
    </div>

    <template #footer>
      <a-button danger ghost @click="closeModal">
        关闭
      </a-button>
      <a-button type="primary" ghost @click="handleConfirm">
        确定
      </a-button>
    </template>
  </a-modal>
</template>

<style lang="less" scoped>
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

.create-roll-call-modal__body {
  padding: 20px 24px 24px;
  min-height: 180px;
}

.create-roll-call-modal__section {
  border: 1px solid #f0f0f0;
  border-radius: 12px;
  padding: 16px 18px;
  background: #fafcff;
}

.create-roll-call-modal__label {
  font-size: 16px;
  font-weight: 600;
  color: #1f1f1f;
  line-height: 24px;
}

.create-roll-call-modal__desc {
  margin-top: 8px;
  font-size: 13px;
  line-height: 22px;
  color: #666;
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
