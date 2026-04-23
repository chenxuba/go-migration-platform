<script setup lang="ts">
import { CloseOutlined } from '@ant-design/icons-vue'
import { computed, ref, watch } from 'vue'
import dayjs from 'dayjs'
import { getPendingRenewalReminderRecordDetailApi, type PendingRenewalReminderRecordDetailItem } from '@/api/edu-center/student-list'
import { useDrawer } from '@/composables/useDrawer'
import { Sex, SexLabel } from '@/enums'
import messageService from '@/utils/messageService'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  recordId: {
    type: String,
    default: '',
  },
})

const emit = defineEmits(['update:open'])

const { openDrawer } = useDrawer(props, emit)
const loading = ref(false)
const activeKey = ref('unsent')
const detail = ref<{
  recordId: string
  templateId: string
  templateName: string
  channelName: string
  readCount: number
  notifyCount: number
  successCount: number
  skippedCount: number
  failedCount: number
  unsentCount: number
  operatorName: string
  sendTime?: string | null
  sentList: PendingRenewalReminderRecordDetailItem[]
  unsentList: PendingRenewalReminderRecordDetailItem[]
}>({
  recordId: '',
  templateId: '',
  templateName: '',
  channelName: '',
  readCount: 0,
  notifyCount: 0,
  successCount: 0,
  skippedCount: 0,
  failedCount: 0,
  unsentCount: 0,
  operatorName: '',
  sendTime: '',
  sentList: [],
  unsentList: [],
})

const currentList = computed(() => (activeKey.value === 'sent' ? detail.value.sentList : detail.value.unsentList))
const currentListCount = computed(() => currentList.value.length)

const columns = [
  {
    title: '学员/性别',
    dataIndex: 'student',
    key: 'student',
    width: 220,
  },
  {
    title: '家校云',
    dataIndex: 'homeSchool',
    key: 'homeSchool',
    width: 160,
  },
]

function getGenderText(sex?: number) {
  const value = Number.isFinite(Number(sex)) ? Number(sex) : Sex.Unknown
  return SexLabel[value as Sex] || SexLabel[Sex.Unknown]
}

function formatSendTime(value?: string | null) {
  if (!value)
    return '-'
  const date = dayjs(value)
  return date.isValid() ? date.format('YYYY-MM-DD HH:mm') : value
}

function getHomeSchoolStatusClass(status?: number) {
  if (Number(status) === 1)
    return 'text-green-600'
  if (Number(status) === 2)
    return 'text-#ccc'
  return 'text-#ccc'
}

async function loadDetail() {
  if (!props.recordId || !props.open)
    return
  loading.value = true
  try {
    const res = await getPendingRenewalReminderRecordDetailApi({ recordId: props.recordId })
    if (res.code !== 200) {
      throw new Error(res.message || '获取消息详情失败')
    }
    const result = res.result
    detail.value = {
      recordId: result?.recordId || '',
      templateId: result?.templateId || '',
      templateName: result?.templateName || '续费提醒',
      channelName: result?.channelName || '微信提醒',
      readCount: Number(result?.readCount || 0),
      notifyCount: Number(result?.notifyCount || 0),
      successCount: Number(result?.successCount || 0),
      skippedCount: Number(result?.skippedCount || 0),
      failedCount: Number(result?.failedCount || 0),
      unsentCount: Number(result?.unsentCount || 0),
      operatorName: result?.operatorName || '-',
      sendTime: result?.sendTime || '',
      sentList: Array.isArray(result?.sentList) ? result.sentList : [],
      unsentList: Array.isArray(result?.unsentList) ? result.unsentList : [],
    }
    activeKey.value = detail.value.sentList.length > 0 ? 'sent' : 'unsent'
  }
  catch (error: any) {
    console.error('load pending renewal reminder detail failed', error)
    messageService.error(error?.message || '获取消息详情失败')
  }
  finally {
    loading.value = false
  }
}

watch(() => [props.open, props.recordId], () => {
  if (props.open)
    void loadDetail()
}, { immediate: true })
</script>

<template>
  <div>
    <a-drawer
      v-model:open="openDrawer"
      :push="{ distance: 80 }"
      :body-style="{ padding: '0', background: '#f7f7fd' }"
      :closable="false"
      width="800px"
      placement="right"
    >
      <template #title>
        <div class="custom-header flex justify-between h-4 flex-items-center">
          <div class="text-5">
            消息详情
          </div>
          <a-button type="text" class="close-btn" @click="openDrawer = false">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>

      <div class="detail-content">
        <div class="detail-header">
          <div class="detail-title-wrap">
            <div class="detail-title">
              {{ detail.templateName || '续费提醒' }}
            </div>
            <span class="channel-badge">
              {{ detail.channelName || '微信提醒' }}
            </span>
          </div>
          <div class="detail-meta">
            {{ detail.operatorName || '-' }} 发布于 {{ formatSendTime(detail.sendTime) }}，通知人数：{{ detail.notifyCount || 0 }}
          </div>

          <div class="tabs">
            <a-tabs
              v-model:active-key="activeKey"
              size="large"
              :tab-bar-style="{ 'border-radius': '0px', 'padding-left': '24px' }"
            >
              <a-tab-pane key="sent" :tab="`已读学员（${detail.sentList.length}）`" />
              <a-tab-pane key="unsent" :tab="`未读学员（${detail.unsentList.length}）`" />
            </a-tabs>
          </div>
        </div>

        <div class="detail-body">
          <div class="table-card">
            <custom-title :title="`共 ${currentListCount} 个学员`" font-size="14px" font-weight="500" class="table-card__title" />

            <a-table
              :loading="loading"
              :columns="columns"
              :data-source="currentList"
              row-key="itemId"
              :pagination="false"
              :scroll="{ x: 420 }"
              size="small"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'student'">
                  <student-avatar
                    :id="record.studentId"
                    :name="record.studentName || '-'"
                    :gender="getGenderText(record.sex)"
                    :avatar-url="record.avatar || ''"
                    :show-age="false"
                    default-active-key="0"
                  />
                </template>

                <template v-if="column.key === 'homeSchool'">
                  <a-tooltip placement="right">
                    <template #title>
                      <span>{{ record.homeSchoolStatusText || '-' }}</span>
                    </template>
                    <div class="home-school-cell flex flex-items-center">
                      <span class="whitespace-nowrap" :class="getHomeSchoolStatusClass(record.homeSchoolStatus)">
                        {{ record.homeSchoolStatusText || '-' }}
                      </span>
                      <img
                        class="ml-2 follow-bind-icon"
                        :class="{ 'follow-bind-icon--inactive': record.homeSchoolStatus !== 1 }"
                        src="~@/assets/images/follow.svg"
                        alt=""
                      >
                    </div>
                  </a-tooltip>
                </template>
              </template>
            </a-table>
          </div>
        </div>
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

.detail-content {
  min-height: 100%;
}

.detail-header {
  background: #fff;
  border-bottom: 1px solid #f0f0f0;
  padding: 20px 24px 0;
}

.detail-title-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
}

.detail-title {
  color: #1f2329;
  font-size: 18px;
  font-weight: 600;
}

.channel-badge {
  color: #0c3;
  border: none;
  border-radius: 10px;
  background-color: #e6ffec;
  white-space: nowrap;
  flex-shrink: 0;
  padding: 0 10px;
  font-size: 12px;
  font-weight: 400;
  line-height: 20px;
  display: inline-block;
}

.detail-meta {
  color: #8a94a6;
  font-size: 14px;
  line-height: 22px;
  margin-top: 10px;
}

.tabs {
  width: 100%;
  line-height: 40px;
  margin-top: 12px;

  :deep(.ant-tabs-nav) {
    border-radius: 16px;
    padding: 0 24px;
    margin-bottom: 0;
  }

  :deep(.ant-tabs-tab) {
    font-size: 16px;
  }

  :deep(.ant-tabs-ink-bar) {
    text-align: center;
    height: 9px !important;
    background: transparent;
    bottom: 1px !important;

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

.detail-body {
  min-height: calc(100% - 132px);
  padding: 16px;
}

.table-card {
  border: 1px solid #edf1f7;
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 8px 24px rgb(15 23 42 / 4%);
  overflow: hidden;
  padding: 12px 16px 16px;
}

.table-card__title {
  margin-bottom: 16px;
}

.cell-primary {
  color: #1f2329;
  line-height: 22px;
}

.home-school-cell {
  line-height: 22px;
}

.follow-bind-icon {
  width: 16px;
  height: 16px;
}

.follow-bind-icon--inactive {
  filter: grayscale(1) opacity(0.45);
}
</style>
