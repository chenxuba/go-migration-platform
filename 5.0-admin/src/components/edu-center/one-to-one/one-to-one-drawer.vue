<script setup>
import { CloseCircleOutlined, CloseOutlined, ExclamationCircleFilled } from '@ant-design/icons-vue'
import schedule from './one-to-one-schedule.vue'
import waitingRollCallSchedule from './one-to-one-waiting-roll-call-schedule.vue'
import classRecord from './one-to-one-class-record.vue'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
})
const emit = defineEmits(['update:open'])
const activeKey = ref('0')
// 处理双向绑定
const openDrawer = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

// 监听抽屉打开，重置 activeKey 为 '0'
watch(() => openDrawer.value, (newVal) => {
  if (newVal) {
    activeKey.value = '0'
  }
})

// defineEmits(['update:open']);
const openModal = ref(false)
function handleDelete() {
  console.log('删除')
  openModal.value = true
}
// 编辑上课信息
const editClassInfoModal = ref(false)
function handleEditClassInfo() {
  editClassInfoModal.value = true
}
// 编辑点名
const editRollNameModal = ref(false)
function handleEditRollName() {
  editRollNameModal.value = true
}
</script>

<template>
  <div>
    <a-drawer
      v-model:open="openDrawer" :push="{ distance: 80 }" :body-style="{ padding: '0', background: '#f7f7fd' }"
      :closable="false" width="1165px" placement="right"
    >
      <!-- 自定义头部 -->
      <template #title>
        <div class="custom-header flex justify-between h-4 flex-items-center">
          <div class="text-5">
            一对一详情
          </div>
          <a-button type="text" class="close-btn" @click="openDrawer = false">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>
      <div class="contenter flex flex-center bg-white px6 py3">
        <div class="avatarBox w-16 h-16 relative">
          <img
            width="64" height="64" class=" rounded-100"
            src="https://pcsys.admin.ybc365.com/83b8fd68-2f9b-4a35-979f-1fd0ea349889.png" alt=""
          >
        </div>
        <div class="info flex flex-1 ml-4 flex-col">
          <div class="top flex justify-between flex-center flex-1">
            <a-space>
              <div class="name text-5 font-800">
                李宇豪-听觉训练课
              </div>
              <span class="bg-#e6f0ff text-#06f text-3 px3 py2px rounded-10 ml2">开课中</span>
            </a-space>
            <a-space>
              <a-button type="primary">
                编辑
              </a-button>
            </a-space>
          </div>
          <div class="bottom flex-1 flex flex-items-center mt-2">
            <div class="birthday flex-center">
              <span class="text-14px text-#888">创建于 2025-05-15 17:36</span>
            </div>
          </div>
        </div>
      </div>
      <div class="desc pt-4 bg-white px6 py3 pb0">
        <a-descriptions :column="4" size="small" :content-style="{ color: '#888' }">
          <a-descriptions-item label="学生姓名">
            李宇豪
          </a-descriptions-item>
          <a-descriptions-item label="学员记录课时">
            1
          </a-descriptions-item>
          <a-descriptions-item label="老师授课课时">
            1
          </a-descriptions-item>
          <a-descriptions-item label="剩余课时">
            0 课时
          </a-descriptions-item>
          <a-descriptions-item label="班主任">
            -
          </a-descriptions-item>
          <a-descriptions-item label="关联课程">
            听觉训练课
          </a-descriptions-item>
          <a-descriptions-item label="升期状态">
            未升期
          </a-descriptions-item>
        </a-descriptions>
      </div>
      <div class="tabs">
        <a-tabs
          v-model:active-key="activeKey" size="large" :tab-bar-style="{
            'border-radius': '0px', 'padding-left': '24px',
          }"
        >
          <a-tab-pane key="0" tab="日程">
            <schedule />
          </a-tab-pane>
          <a-tab-pane key="1" tab="待点名日程">
            <waiting-roll-call-schedule />
          </a-tab-pane>
          <a-tab-pane key="2" tab="上课记录">
            <class-record />
          </a-tab-pane>
        </a-tabs>
      </div>
    </a-drawer>
    <a-modal
      v-model:open="openModal" centered :footer="false" :closable="false" :mask-closable="false"
      :keyboard="false" width="420px" @ok="handleDelete"
    >
      <div class="text-18px mb-12px font500">
        <CloseCircleOutlined class="text-#f00 mr2 text-5" /> 删除上课点名记录？
      </div>
      <div class="pl-30px text-#666">
        <div>1.删除后已点名扣费学员将会反还学费，并减少对应的已确认收入;</div>
        <div>2.若包含试听、补课学员状态不会修改，已试听状态变成已取消状态，已补课学员将退回至未安排状态，并删除上课记录;</div>
        <div>3.删除上课点名记录后，（除了试听日程）所对应的日程中的学员点名状态变成未点名;</div>
        <div>4.删除上课点名记录后，日程状态从已点名变成未点名。</div>
        <div class="text-#f00 mt-12px">
          <ExclamationCircleFilled /> 此操作不可撤销，请谨慎操作
        </div>
      </div>
      <a-space class="mt-24px flex justify-end">
        <a-button danger ghost>
          删除
        </a-button>
        <a-button class="text-#666" @click="openModal = false">
          再想想
        </a-button>
      </a-space>
    </a-modal>
  </div>
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

.tabs {
  width: 100%;
  border-radius: 10px;

  :deep(.ant-tabs-nav) {
    background: #fff;
    margin: 0;
  }

  :deep(.ant-tabs-ink-bar) {
    text-align: center;
    height: 12px !important;
    background: transparent;
    bottom: 0px !important;

    &::after {
      position: absolute;
      top: 0;
      left: calc(50% - 12px);
      width: 24px !important;
      height: 4px !important;
      border-radius: 2px;
      background-color: var(--pro-ant-color-primary);
      content: "";
    }
  }
}
</style>
