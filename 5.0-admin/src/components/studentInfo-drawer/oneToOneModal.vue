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
function closeFun() {
  openModal.value = false
}
const list = [
  {
    type: 1,
    name: '周 8:00-8:30',
    time: '重复，2025-05-13 ~ 2025-05-18',
  },
  {
    type: 1,
    name: '周 8:00-8:30',
    time: '重复，2025-05-13 ~ 2025-05-18',
  },
  {
    type: 2,
    name: '周五 13:00-14:06',
    time: '自由，2025-05-13',
  },
  {
    type: 2,
    name: '周五 13:00-14:06',
    time: '自由，2025-05-13',
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
        <span>查看一对一</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <div class="contenter scrollbar">
      <div class="flex flex-center bg-white  py3">
        <div class="avatarBox w-16 h-16 ">
          <img
            width="64" height="64" src="https://pcsys.admin.ybc365.com/83b8fd68-2f9b-4a35-979f-1fd0ea349889.png"
            alt=""
          >
        </div>
        <div class="info flex flex-1 ml-4 flex-col ">
          <div class="top flex justify-between flex-center flex-1">
            <a-space>
              <div class="name text-16px font-800">
                陈陈-初级感统课
              </div>
            </a-space>
          </div>
          <div class="bottom flex-1 flex flex-items-center mt-2">
            <div class="birthday flex-center">
              <span class="text-14px text-#888">听觉训练课</span>
            </div>
          </div>
        </div>
      </div>
      <div class="text-14px text-#888">
        班主任：未分配
      </div>
      <div class="bg-#f6f7f8 rounded-12px p-16px mt-12px">
        <custom-title title="排课日程共 6 个" font-size="14px" class="mb-12px" />
        <a-descriptions :column="2" :content-style="{ color: '#888' }">
          <a-descriptions-item v-for="(item, index) in list" :key="index" style="padding: 15px 0;">
            <div class="flex flex-center">
              <img
                v-if="item.type === 1" class="w-34px h-34px" src="https://pcsys.admin.ybc365.com/56fe4dce-75ac-4965-bc5f-825a5e273f0a.png"
                alt=""
              >
              <img
                v-if="item.type === 2" class="w-34px h-34px" src="https://pcsys.admin.ybc365.com/0fab2373-0087-4745-bc70-a717e65172b4.png"
                alt=""
              >
              <div class="ml-12px">
                <div class="text-14px text-#666">
                  {{ item.name }}
                </div>
                <div class="text-12px">
                  {{ item.time }}
                </div>
              </div>
            </div>
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
