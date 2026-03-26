<script setup>
import { computed, ref } from 'vue'
import { LeftOutlined } from '@ant-design/icons-vue'
import { useModalStore } from '~/stores/modal'
import BusinessSettings from '~/components/business-settings/index.vue'

const modalStore = useModalStore()
const isConfirmVisible = ref(false)
const modalTitle = computed(() => modalStore.modalTitle)
const showBackButton = computed(() => modalStore.modalBackButton)
const businessSettingsRef = ref(null)

// 判断各类模态框是否显示
const isBusinessSettingsVisible = computed(() =>
  modalStore.isModalActive('BusinessSettings'),
)

// 处理返回按钮点击
function handleBackClick() {
  if (businessSettingsRef.value) {
    businessSettingsRef.value.goBack()
  }
}

// 关闭模态框的处理方法
function handleCloseModal() {
  // 先显示确认对话框
  isConfirmVisible.value = true
}

// 确认关闭
function confirmClose() {
  // 先关闭确认对话框
  isConfirmVisible.value = false
  // 然后关闭主模态框
  modalStore.closeModal()
}
</script>

<template>
  <div>
    <!-- 业务设置模态框 -->
    <a-modal
      v-model:open="isBusinessSettingsVisible"
      style="top:40px;overflow: hidden;"
      wrap-class-name="business-settings-modal"
      :width="375"
      :footer="null"
      :closable="!showBackButton"
      :mask-closable="false"
      @cancel="handleCloseModal"
    >
      <template #title>
        <div class="modal-custom-title">
          <span v-if="showBackButton" class="modal-back-button" @click="handleBackClick">
            <LeftOutlined />
          </span>
          <span>{{ modalTitle }}</span>
        </div>
      </template>
      <BusinessSettings ref="businessSettingsRef" />
    </a-modal>

    <!-- 确认关闭提示框 -->
    <a-modal
      v-model:open="isConfirmVisible" :centered="true" :closable="false" title="手动刷新生效" :footer="null"
      :mask-closable="false" :keyboard="false" :width="500"
    >
      <p>关闭当前弹窗后，需手动刷新页面使设置生效</p>
      <div class="text-right mt-4">
        <a-button type="primary" @click="confirmClose">
          知道了
        </a-button>
      </div>
    </a-modal>

    <!-- 可以添加更多模态框类型 -->
  </div>
</template>

<style lang="less" scoped>
.modal-custom-title {
  display: flex;
  align-items: center;
}

.modal-back-button {
  margin-right: 12px;
  cursor: pointer;
  font-size: 16px;
}
</style>
