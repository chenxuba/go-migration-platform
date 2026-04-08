<script setup>
import { CloseOutlined } from '@ant-design/icons-vue'
import { useDrawer } from '@/composables/useDrawer'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['update:open'])

const { openDrawer } = useDrawer(props, emit)
</script>

<template>
  <div>
    <a-drawer
      v-model:open="openDrawer"
      :push="{ distance: 80 }"
      :body-style="{ padding: '0', background: '#f7f7fd' }"
      :closable="false"
      width="1162px"
      placement="right"
    >
      <template #title>
        <div class="custom-header flex justify-between h-4 flex-items-center">
          <div class="text-5">
            日程详情
          </div>
          <a-button type="text" class="close-btn" @click="openDrawer = false">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>

      <div class="contenter">
        demo
      </div>
    </a-drawer>
  </div>
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
</style>
