<script setup>
import { CloseCircleOutlined, CloseOutlined, ExclamationCircleFilled } from '@ant-design/icons-vue'

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

// defineEmits(['update:open']);
const openModal = ref(false)
function handleDelete() {
  console.log('删除')
  openModal.value = true
}
const columns = ref([
  // 学员姓名	联系电话	状态	操作
  {
    title: '学员姓名',
    dataIndex: 'name',
    key: 'name',
  },
  {
    title: '联系电话',
    dataIndex: 'phone',
    key: 'phone',
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
  },
])
const data = ref([
  { name: '张三', phone: '13800138000', status: '已点名', action: '详情' },
])

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
            日程详情
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
            src="@/assets/images/timetable/schedule-one2one.png" alt=""
          >
        </div>
        <div class="info flex flex-1 ml-4 flex-col">
          <div class="top flex justify-between flex-center flex-1">
            <a-space>
              <div class="name text-5 font-800">
                半年-认知课
              </div>
            </a-space>
            <a-space>
              <a-button type='link' ghost>
                仅复制此日程
              </a-button>
              <a-button danger ghost @click="handleDelete">
                删除
              </a-button>
              <a-button type="primary" @click="handleEditRollName">
                编辑
              </a-button>
              <a-button type="primary">
                去点名
              </a-button>
            </a-space>
          </div>
          <div class="bottom flex-1 flex flex-items-center mt-2">
            <div class="birthday flex-center">
              <span class="text-4 text-#222">2025-04-14(周一)10:00 ~ 10:30</span>
              <span class="bg-#e6f0ff text-#06f text-3 px2 py1 rounded-10 ml2">30分钟</span>
            </div>
          </div>
        </div>
      </div>
      <div class="desc pt-4 bg-white px6 py3 pb0">
        <a-descriptions :column="4" size="small" :content-style="{ color: '#888' }">
          <a-descriptions-item label="上课教师">
            张晨
          </a-descriptions-item>
          <a-descriptions-item label="上课助教">
            陈瑞生
          </a-descriptions-item>
          <a-descriptions-item label="上课教室">
            -
          </a-descriptions-item>
          <a-descriptions-item label="重复规则">
            不重复
          </a-descriptions-item>
          <a-descriptions-item label="对内备注">
            -
          </a-descriptions-item>
          <a-descriptions-item label="对外备注">
            -
          </a-descriptions-item>
        </a-descriptions>
      </div>
      <div class="tabs">
        <a-tabs
          v-model:active-key="activeKey" size="large" :tab-bar-style="{
            'border-radius': '0px', 'padding-left': '24px',
          }"
        >
          <a-tab-pane key="0" tab="学员名单">
            <a-card title="上课学员">
              <a-table :columns="columns" :data-source="data" />
            </a-card>
          </a-tab-pane>
        </a-tabs>
      </div>
    </a-drawer>
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
