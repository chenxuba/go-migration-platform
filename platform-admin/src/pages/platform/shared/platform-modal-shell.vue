<script setup lang="ts">
import { CloseOutlined } from '@ant-design/icons-vue'
import { computed, useSlots } from 'vue'

const props = withDefaults(defineProps<{
  open: boolean
  title: string
  width?: number | string
  scrollable?: boolean
  destroyOnClose?: boolean
  modalClass?: string
}>(), {
  width: 960,
  scrollable: false,
  destroyOnClose: true,
  modalClass: '',
})

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'close'): void
}>()

const slots = useSlots()

const openModel = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const modalClassName = computed(() => {
  return ['createStu-modal-content-box', 'platform-modal-shell', props.modalClass]
    .filter(Boolean)
    .join(' ')
})

const hasFooter = computed(() => Boolean(slots.footer))

function handleClose() {
  emit('update:open', false)
  emit('close')
}
</script>

<template>
  <a-modal
    v-model:open="openModel"
    centered
    :destroy-on-close="destroyOnClose"
    :keyboard="false"
    :closable="false"
    :mask-closable="false"
    :width="width"
    :class="modalClassName"
    :footer="hasFooter ? undefined : null"
    @cancel="handleClose"
  >
    <template #title>
      <div class="platform-modal-shell__titlebar">
        <span>{{ title }}</span>
        <a-button type="text" class="close-btn" @click="handleClose">
          <template #icon>
            <CloseOutlined class="close-icon" />
          </template>
        </a-button>
      </div>
    </template>

    <div class="platform-modal-shell__body" :class="{ 'platform-modal-shell__body--scrollable': scrollable }">
      <slot />
    </div>

    <template v-if="hasFooter" #footer>
      <slot name="footer" />
    </template>
  </a-modal>
</template>

<style scoped lang="less">
:deep(.createStu-modal-content-box.platform-modal-shell .ant-modal-content) {
  border-radius: 22px;
  overflow: hidden;
  box-shadow: 0 18px 46px rgba(15, 23, 42, 0.14);
}

:deep(.createStu-modal-content-box.platform-modal-shell .ant-modal-header) {
  padding: 24px 28px 14px;
  margin-bottom: 0;
  border-bottom: none;
}

:deep(.createStu-modal-content-box.platform-modal-shell .ant-modal-body) {
  padding: 0 28px 0;
}

:deep(.createStu-modal-content-box.platform-modal-shell .ant-modal-footer) {
  padding: 18px 28px 24px;
  border-top: none;
}

.platform-modal-shell__titlebar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  color: #1f2329;
  font-size: 20px;
  font-weight: 700;
  line-height: 32px;
}

.platform-modal-shell__body--scrollable {
  max-height: calc(100vh - 220px);
  overflow-y: auto;
  overflow-x: hidden;
}

.close-btn {
  width: 40px;
  height: 40px;
  color: #1f2329;
  font-size: 22px;
}

.close-btn:hover {
  background: transparent;
}

.close-btn:hover .close-icon {
  animation: icon-rotate 0.3s linear;
}

@keyframes icon-rotate {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(180deg);
  }
}
</style>
