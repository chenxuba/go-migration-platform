<script setup>
import { CloseOutlined } from '@ant-design/icons-vue'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
})
const emit = defineEmits(['update:open'])
// 处理双向绑定
const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const list = ref([
  {
    id: 1,
    type: 'stop',
    name: '停课',
    date: '2025-05-15',
  },
  {
    id: 2,
    type: 'recovery',
    name: '复课',
    date: '2025-05-15',
  },
])

function closeFun() {
  openModal.value = false
}
</script>

<template>
  <a-modal
    v-model:open="openModal" centered class="modal-content-box" :keyboard="false" :closable="false"
    :mask-closable="false" :width="800" :footer="false"
  >
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>停/复课记录</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <div class="pb-2px">
      <div v-for="(item, index) in list" :key="index" class="contenter scrollbar">
        <a-descriptions class="descriptions" :column="3" size="small" :content-style="{ color: '#888' }">
          <template #title>
            <div class="flex flex-items-center mb-8px">
              <img
                v-if="item.type === 'recovery'"
                src="https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12185/static/recovery-course.29f340ff.svg"
                class="w-14px h-14px mr-4px mt-1px" alt=""
              >
              <img
                v-else
                src="https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12185/static/stop-course.a6619221.svg"
                class="w-14px h-14px mr-4px mt-1px" alt=""
              >
              <span class="text-14px">{{ item.name }}</span>
            </div>
          </template>
          <a-descriptions-item label="复课日期">
            2025-05-15
          </a-descriptions-item>
          <a-descriptions-item label="现有效期至">
            不限制
          </a-descriptions-item>
          <a-descriptions-item label="操作时间">
            2025-05-15
          </a-descriptions-item>
          <a-descriptions-item label="操作人">
            陈瑞
          </a-descriptions-item>
          <a-descriptions-item label="复课备注">
            -
          </a-descriptions-item>
        </a-descriptions>
      </div>
    </div>
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
  margin:18px 24px;
  border-radius: 14px;
  background: #f6f7f8;

  :deep(.descriptions .ant-descriptions-header) {
    margin-bottom: 0;
  }
}
</style>
