<script setup>
import { CloseOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import enrollmentDetail from './one-to-one-enrollment-detail.vue'
import schedule from './one-to-one-schedule.vue'
import waitingRollCallSchedule from './one-to-one-waiting-roll-call-schedule.vue'
import classRecord from './one-to-one-class-record.vue'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  record: {
    type: Object,
    default: () => ({}),
  },
  /** 当前 1 对 1 报读对应的学费账户（按 orderCourseDetailId 过滤后的列表） */
  tuitionAccounts: {
    type: Array,
    default: () => [],
  },
})
const emit = defineEmits(['update:open', 'edit'])
const activeKey = ref('0')

const openDrawer = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

watch(() => openDrawer.value, (newVal) => {
  if (newVal) {
    activeKey.value = '0'
  }
})

function formatCreatedTime(value) {
  if (!value || `${value}`.startsWith('0001-01-01')) {
    return '2026-03-28 20:51'
  }
  return dayjs(value).format('YYYY-MM-DD HH:mm')
}

function getStatusText(status) {
  return status === 2 ? '已结班' : '开班中'
}

const isClassClosed = computed(() => Number(props.record?.status) === 2)

function handleEdit() {
  emit('edit', props.record)
}
</script>

<template>
  <a-drawer
    v-model:open="openDrawer"
    :push="{ distance: 80 }"
    :body-style="{ padding: '0', background: '#f7f7fd' }"
    :closable="false"
    width="1165px"
    placement="right"
  >
    <template #title>
      <div class="custom-header flex justify-between h-4 flex-items-center">
        <div class="text-5">
          1对1详情
        </div>
        <a-button type="text" class="close-btn" @click="openDrawer = false">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>

    <div class="contenter flex justify-between bg-white px6 py4">
      <div class="flex flex-1">
        <div class="one-to-one-badge">
          <span>1v1</span>
        </div>
        <div class="info flex flex-1 ml-4 flex-col">
          <div class="top flex flex-items-center flex-wrap">
            <div
              class="name text-20px font-800 mr-3"
              :class="{ 'name--closed': isClassClosed }"
            >
              {{ record?.name || '柳一一一-感统课' }}
            </div>
            <span
              class="status-chip"
              :class="{ 'status-chip--closed': isClassClosed }"
            >
              {{ getStatusText(record?.status) }}
            </span>
          </div>
          <div class="bottom mt-2 text-#888 text-14px">
            龙钊 创建于 {{ formatCreatedTime(record?.createdTime) }}
          </div>
        </div>
      </div>
      <div v-if="!isClassClosed" class="ml-4">
        <a-button type="primary" @click="handleEdit">
          编辑
        </a-button>
      </div>
    </div>

    <div class="desc bg-white px6 pt2 pb4">
      <a-descriptions
        :column="3"
        size="small"
        :content-style="{ color: '#666' }"
      >
        <a-descriptions-item label="学生姓名">
          {{ record?.studentName || '柳一一' }}
        </a-descriptions-item>
        <a-descriptions-item label="课时记录方式">
          按固定课时记录
        </a-descriptions-item>
        <a-descriptions-item label="学员记录课时">
          1课时
        </a-descriptions-item>
        <a-descriptions-item label="老师授课课时">
          0课时
        </a-descriptions-item>
        <a-descriptions-item label="班主任">
          -
        </a-descriptions-item>
        <a-descriptions-item label="关联课程">
          {{ record?.lessonName || '感统课' }}
        </a-descriptions-item>
        <a-descriptions-item label="升期状态">
          未升期
        </a-descriptions-item>
        <a-descriptions-item label="默认上课教师">
          -
        </a-descriptions-item>
        <a-descriptions-item label="备注">
          -
        </a-descriptions-item>
      </a-descriptions>
    </div>

    <div class="tabs">
      <a-tabs
        v-model:active-key="activeKey"
        size="large"
        :tab-bar-style="{
          'border-radius': '0px',
          'padding-left': '24px',
        }"
      >
        <a-tab-pane key="0" tab="报读明细">
          <enrollment-detail :record="record" :accounts="tuitionAccounts" />
        </a-tab-pane>
        <a-tab-pane key="1" tab="日程">
          <schedule />
        </a-tab-pane>
        <a-tab-pane key="2" tab="待点名日程">
          <waiting-roll-call-schedule />
        </a-tab-pane>
        <a-tab-pane key="3" tab="上课记录">
          <class-record />
        </a-tab-pane>
      </a-tabs>
    </div>
  </a-drawer>
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

.one-to-one-badge {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  background: linear-gradient(180deg, #ff9696 0%, #ff6f6f 100%);
  box-shadow: 0 8px 20px rgba(255, 111, 111, 0.28);
  display: flex;
  align-items: center;
  justify-content: center;

  span {
    color: #fff;
    font-size: 20px;
    font-weight: bold;
    font-style: italic;
  }
}

.status-chip {
  display: inline-flex;
  align-items: center;
  height: 24px;
  padding: 0 16px;
  border-radius: 999px;
  background: #e9f2ff;
  color: #1f6fff;
  font-size: 12px;
}

.status-chip--closed {
  background: #fff1f0;
  color: #cf1322;
  border: 1px solid #ffccc7;
}

.name--closed {
  opacity: 0.55;
  color: #595959;
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
