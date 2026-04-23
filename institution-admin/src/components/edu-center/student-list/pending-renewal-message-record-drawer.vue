<script setup lang="ts">
import { CloseOutlined } from '@ant-design/icons-vue'
import { ref, watch } from 'vue'
import dayjs from 'dayjs'
import {
  pagePendingRenewalReminderRecordsApi,
  type PendingRenewalReminderRecordPageItem,
} from '@/api/edu-center/student-list'
import { useDrawer } from '@/composables/useDrawer'
import messageService from '@/utils/messageService'
import PendingRenewalMessageDetailDrawer from './pending-renewal-message-detail-drawer.vue'
import PendingRenewalMessageTemplateExampleModal from './pending-renewal-message-template-example-modal.vue'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['update:open'])

const { openDrawer } = useDrawer(props, emit)
const loading = ref(false)
const templateExampleOpen = ref(false)
const detailOpen = ref(false)
const currentRecordId = ref('')
const dataSource = ref<PendingRenewalReminderRecordPageItem[]>([])
const pagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  pageSizeOptions: ['10', '20', '50'],
  showTotal: (total: number) => `共 ${total} 条消息`,
})

const columns = [
  {
    title: '消息模板名称',
    dataIndex: 'templateName',
    key: 'templateName',
    width: 130,
  },
  {
    title: '发送方式',
    dataIndex: 'channelName',
    key: 'channelName',
    width: 120,
  },
  {
    title: '通知人数',
    dataIndex: 'notifyCount',
    key: 'notifyCount',
    width: 100,
  },
  {
    title: '发送人',
    dataIndex: 'operatorName',
    key: 'operatorName',
    width: 120,
  },
  {
    title: '发送时间',
    dataIndex: 'sendTime',
    key: 'sendTime',
    width: 180,
  },
]

function formatDateTime(value?: string | null) {
  if (!value)
    return '-'
  const date = dayjs(value)
  return date.isValid() ? date.format('YYYY-MM-DD HH:mm') : value
}

async function getList() {
  if (!props.open)
    return
  loading.value = true
  try {
    const res = await pagePendingRenewalReminderRecordsApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: pagination.value.pageSize,
        pageIndex: pagination.value.current,
        skipCount: (pagination.value.current - 1) * pagination.value.pageSize,
      },
    })
    if (res.code !== 200) {
      throw new Error(res.message || '获取消息记录失败')
    }
    const result = res.result
    dataSource.value = Array.isArray(result?.list) ? result.list : []
    pagination.value.total = Number(result?.total || 0)
  }
  catch (error: any) {
    console.error('load pending renewal reminder records failed', error)
    messageService.error(error?.message || '获取消息记录失败')
  }
  finally {
    loading.value = false
  }
}

function handleTableChange(pageInfo: { current?: number, pageSize?: number }) {
  pagination.value.current = Number(pageInfo.current || 1)
  pagination.value.pageSize = Number(pageInfo.pageSize || pagination.value.pageSize)
  void getList()
}

function handleOpenDetail(recordId: string) {
  currentRecordId.value = recordId
  detailOpen.value = true
}

watch(() => props.open, (val) => {
  if (!val)
    return
  pagination.value.current = 1
  void getList()
}, { immediate: true })

defineExpose({
  getList,
})
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
            消息记录
          </div>
          <a-button type="text" class="close-btn" @click="openDrawer = false">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>

      <div class="drawer-content">
        <div class="table-card">
          <custom-title :title="`共 ${pagination.total} 条消息`" font-size="14px" font-weight="500" class="table-card__title">
            <template #right>
              <a-button @click="templateExampleOpen = true">
                模板示例
              </a-button>
            </template>
          </custom-title>

          <a-table
            :loading="loading"
            :columns="columns"
            :data-source="dataSource"
            :pagination="pagination"
            row-key="recordId"
            :scroll="{ x: 730 }"
            size="small"
            @change="handleTableChange"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'templateName'">
                <a-tooltip title="点击查看详情">
                  <a-button type="link" class="record-link" @click="handleOpenDetail(record.recordId)">
                    {{ record.templateName || '续费提醒' }}
                  </a-button>
                </a-tooltip>
              </template>

              <template v-if="column.key === 'channelName'">
                <span class="channel-badge">
                  {{ record.channelName || '微信提醒' }}
                </span>
              </template>

              <template v-if="column.key === 'notifyCount'">
                {{ record.readCount || 0 }}/{{ record.notifyCount || 0 }}
              </template>

              <template v-if="column.key === 'sendTime'">
                {{ formatDateTime(record.sendTime) }}
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </a-drawer>

    <pending-renewal-message-template-example-modal
      v-model:open="templateExampleOpen"
    />

    <pending-renewal-message-detail-drawer
      v-model:open="detailOpen"
      :record-id="currentRecordId"
    />
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

.drawer-content {
  min-height: 100%;
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

.record-link {
  padding-left: 0;
  color: #1f2329;
  font-weight: 400;

  &:hover,
  &:focus {
    color: #1677ff;
  }
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
</style>
