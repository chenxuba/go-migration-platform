<script setup>
import { CloseOutlined } from '@ant-design/icons-vue'
import { computed, onMounted, onUnmounted, ref } from 'vue'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
})
const emit = defineEmits(['update:open'])
const formRef = ref()
// 处理双向绑定
const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

function closeFun() {
  openModal.value = false
}

// 设置表格最大高度，窗口变化时动态调整
const totalHeight = ref(window.innerHeight - 200)

// 监听窗口大小变化
function updateTableHeight() {
  totalHeight.value = window.innerHeight - 200
}

onMounted(() => {
  window.addEventListener('resize', updateTableHeight)
})

onUnmounted(() => {
  window.removeEventListener('resize', updateTableHeight)
})

const columns = [
  {
    title: '变动类型',
    dataIndex: 'type',
  },
  {
    title: '变动时间',
    dataIndex: 'time',
  },
  {
    title: '学费变动',
    dataIndex: 'fee',
  },
  {
    title: '数量变动',
    dataIndex: 'num',
  },
]
const data = [
  {
    type: '学费变动',
    time: '2025-04-14 10:00',
    fee: '+ ¥ 2300.00',
    num: '+ 10 课时',
  },
  {
    type: '课消退还',
    time: '2025-04-14 10:00',
    fee: '- ¥ 2300.00',
    num: '- 10 课时',
  },
  {
    type: '课消补扣',
    time: '2025-04-14 10:00',
    fee: '- ¥ 2300.00',
    num: '- 10 课时',
  },
  {
    type: '课消',
    time: '2025-04-14 10:00',
    fee: '- ¥ 2300.00',
    num: '- 10 课时',
  },
  {
    type: '报名',
    time: '2025-04-14 10:00',
    fee: '+ ¥ 2300.00',
    num: '+ 10 课时',
  },
  {
    type: '报名',
    time: '2025-04-14 10:00',
    fee: '+ ¥ 2300.00',
    num: '+ 10 课时',
  },
  {
    type: '报名',
    time: '2025-04-14 10:00',
    fee: '+ ¥ 2300.00',
    num: '+ 10 课时',
  },
  {
    type: '报名',
    time: '2025-04-14 10:00',
    fee: '+ ¥ 2300.00',
    num: '+ 10 课时',
  },

]
</script>

<template>
  <a-modal
    v-model:open="openModal" centered class="modal-content-box" :keyboard="false" :closable="false"
    :mask-closable="false" :width="800" :footer="false"
  >
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>学费变动记录</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <div class="contenter scrollbar">
      <div class="bg-#fff py-16px px-24px">
        <div class="top mb-16px flex justify-between flex-center">
          <div class="top-left flex flex-col">
            <span class="text-20px text-#222 font-500">视知觉训练</span>
            <a-space>
              <span class="bg-#e6f0ff text-#06f text-3 px3 py2px rounded-10">班级授课</span>
              <span class="bg-#e6f0ff text-#06f text-3 px3 py2px rounded-10">课时</span>
            </a-space>
          </div>
          <div class="top-right">
            <div class="custom-num-font-family text-20px text-#06f font-500">
              ¥ 2300.00
            </div>
            <span class="text-12px text-#999 flex justify-end">剩余学费</span>
          </div>
        </div>
        <!-- 描述列表 -->
        <a-descriptions :column="3" :content-style="{ color: '#888' }">
          <a-descriptions-item label="总学费">
            ¥ 2300.00
          </a-descriptions-item>
          <a-descriptions-item label="总课时">
            21课时
          </a-descriptions-item>
          <a-descriptions-item label="剩余课时">
            10课时
          </a-descriptions-item>
        </a-descriptions>
      </div>
      <div class="tables p-16px rounded-10">
        <div class="bg-#fff rounded-12px p-16px h-600px">
          <custom-title title="共 8 条记录" font-size="14px" class="mb-12px" />
          <!-- 变动类型	变动时间	学费变动	数量变动 -->
          <a-table :columns="columns" :data-source="data" :scroll="{ y: totalHeight }" />
        </div>
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
  background: #f7f6fd;
  border-radius: 14px;
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
