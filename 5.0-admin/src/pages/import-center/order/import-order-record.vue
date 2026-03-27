<script setup>
import { ExclamationCircleOutlined, LeftOutlined } from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { createVNode } from 'vue'
import { useRouter } from 'vue-router'
import {
  clearOrderImportTaskListApi,
  deleteOrderImportTaskApi,
  getOrderImportTaskListApi,
} from '~@/api/finance-center/order-import'
import { useUserStore } from '~@/stores/user'
import messageService from '~@/utils/messageService'

const router = useRouter()
const userStore = useUserStore()
const schoolName = computed(() => userStore.userInfo?.orgName || '总机构')
const loading = ref(false)
const records = ref([])
let pollingTimer = null

function goBack() {
  router.replace('/import-center/import-order-starter')
}

function viewDetail(record) {
  router.push(`/import-center/import-order/record/${record.id}`)
}

function statusText(status) {
  if (status === 4)
    return '导入中'
  return status === 3 ? '待处理' : '已完成'
}

function statusClass(status) {
  if (status === 4)
    return 'status-chip status-chip--running'
  if (status === 3)
    return 'status-chip status-chip--pending'
  return 'status-chip status-chip--completed'
}

async function loadRecords() {
  loading.value = true
  try {
    const { result, data } = await getOrderImportTaskListApi()
    records.value = result?.list || data?.list || []
  }
  catch (error) {
    console.error(error)
    messageService.error('加载导入记录失败')
  }
  finally {
    loading.value = false
  }
}

function handleClearRecords() {
  Modal.confirm({
    title: '确认清空导入记录？',
    icon: createVNode(ExclamationCircleOutlined),
    centered: true,
    okText: '确认清空',
    okType: 'danger',
    cancelText: '取消',
    content: '将清空当前机构的全部学员订单导入记录，该操作不可恢复。',
    async onOk() {
      try {
        const res = await clearOrderImportTaskListApi()
        if (res.code === 200) {
          messageService.success('导入记录已清空')
          await loadRecords()
          return
        }
        return Promise.reject(new Error(res.message || '清空失败'))
      }
      catch (error) {
        console.error(error)
        messageService.error('清空失败，请稍后重试')
        return Promise.reject(error)
      }
    },
  })
}

function handleDeleteRecord(record) {
  Modal.confirm({
    title: '确认删除这条导入记录？',
    icon: createVNode(ExclamationCircleOutlined),
    centered: true,
    okText: '确认删除',
    okType: 'danger',
    cancelText: '取消',
    content: '删除后将无法恢复该待处理导入任务。',
    async onOk() {
      try {
        const res = await deleteOrderImportTaskApi({ taskId: record.id })
        if (res.code === 200) {
          messageService.success('导入记录已删除')
          await loadRecords()
          return
        }
        return Promise.reject(new Error(res.message || '删除失败'))
      }
      catch (error) {
        console.error(error)
        messageService.error('删除失败，请稍后重试')
        return Promise.reject(error)
      }
    },
  })
}

onMounted(() => {
  loadRecords()
  pollingTimer = window.setInterval(() => {
    if (records.value.some(item => item.status === 4))
      loadRecords()
  }, 2000)
})

onUnmounted(() => {
  if (pollingTimer) {
    clearInterval(pollingTimer)
    pollingTimer = null
  }
})
</script>

<template>
  <div class="import-record-layout">
    <div class="work-top flex justify-between items-center h-56px bg-#fff">
      <div class="work-top-left flex items-center">
        <div class="import-header-logo" title="导入中心" aria-hidden="true" />
        <span class="text-20px text-#06f font500 ml-16px flex items-center cursor-pointer" @click="goBack">
          <LeftOutlined class="mt--1px" /> 返回
        </span>
      </div>
      <div class="work-top-right pr-20px text-16px text-#000 font500">
        当前机构：{{ schoolName }}
      </div>
    </div>

    <div class="work-main">
      <div class="work-main-card">
        <div class="record-header">
          <div class="record-title">
            学员订单导入记录
          </div>
          <a-button danger ghost @click="handleClearRecords">
            清空记录
          </a-button>
        </div>

        <a-table
          :loading="loading"
          :data-source="records"
          :pagination="{ pageSize: 10, hideOnSinglePage: true }"
          row-key="id"
          class="mt-24px"
        >
          <a-table-column title="文件名称" data-index="fileName" key="fileName" />
          <a-table-column title="状态" key="status">
            <template #default="{ record }">
              <span :class="statusClass(record.status)">
                <span class="status-dot" />
                {{ statusText(record.status) }}
              </span>
            </template>
          </a-table-column>
          <a-table-column title="导入时间" key="createdTime">
            <template #default="{ record }">
              {{ record.createdTime ? record.createdTime.replace('T', ' ').slice(0, 16) : '-' }}
            </template>
          </a-table-column>
          <a-table-column title="导入人" data-index="uploadStaffName" key="uploadStaffName" />
          <a-table-column title="结果" key="result">
            <template #default="{ record }">
              <span>
                <template v-if="record.status === 4">
                  已导入{{ (record.executedRows || 0) + (record.errorRows || 0) }}/{{ record.totalRows || 0 }}
                </template>
                <template v-else>
                  导入共计{{ record.totalRows || 0 }}（<span :class="record.executedRows > 0 ? 'success-count' : 'neutral-count'">成功{{ record.executedRows || 0 }}</span>/<span :class="record.errorRows > 0 ? 'fail-count' : 'neutral-count'">失败{{ record.errorRows || 0 }}</span>）
                </template>
              </span>
            </template>
          </a-table-column>
          <a-table-column title="操作" key="action" width="100">
            <template #default="{ record }">
              <template v-if="record.status === 3">
                <a-button type="link" danger @click="handleDeleteRecord(record)">
                  删除
                </a-button>
              </template>
              <template v-else>
                <a-button type="link" @click="viewDetail(record)">
                  详情
                </a-button>
              </template>
            </template>
          </a-table-column>
        </a-table>
      </div>
    </div>
  </div>
</template>

<style lang="less" scoped>
.import-record-layout {
  min-height: 100vh;
  background: #f7f7fd;
}

.import-header-logo {
  flex-shrink: 0;
  width: 56px;
  height: 56px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.22) 0%, transparent 42%),
    linear-gradient(145deg, #2b8cff 0%, #0066ff 45%, #0050d8 100%);
  position: relative;
  overflow: hidden;
}

.import-header-logo::before {
  content: '';
  position: absolute;
  left: 12px;
  top: 15px;
  width: 32px;
  height: 26px;
  background-color: rgba(255, 255, 255, 0.94);
  background-image:
    linear-gradient(rgba(0, 102, 255, 0.22), rgba(0, 102, 255, 0.22)),
    linear-gradient(rgba(0, 102, 255, 0.18), rgba(0, 102, 255, 0.18)),
    linear-gradient(rgba(0, 102, 255, 0.14), rgba(0, 102, 255, 0.14));
  background-size: 24px 2px, 18px 2px, 22px 2px;
  background-position: 4px 8px, 4px 14px, 4px 20px;
  background-repeat: no-repeat;
}

.work-main {
  display: flex;
  justify-content: center;
  padding: 40px 0;
}

.work-main-card {
  width: 1300px;
  padding: 40px 48px 48px;
  border-radius: 24px;
  background: #fff;
  box-shadow: 0 0 32px rgba(0, 0, 0, 0.08);
}

.record-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.record-title {
  color: #222;
  font-size: 24px;
  font-weight: 600;
}

.status-chip {
  display: inline-flex;
  align-items: center;
  min-width: 72px;
  height: 28px;
  padding: 0 12px;
  border-radius: 999px;
  font-size: 14px;
  font-weight: 500;
  line-height: 28px;
}

.status-chip .status-dot {
  display: inline-block;
  width: 7px;
  height: 7px;
  margin-right: 8px;
  border-radius: 50%;
}

.status-chip--running {
  color: #d48806;
  background: #fff7e6;
}

.status-chip--running .status-dot {
  background: #faad14;
  box-shadow: 0 0 0 3px rgba(250, 173, 20, 0.16);
}

.status-chip--pending {
  color: #8c8c8c;
  background: #f5f5f5;
}

.status-chip--pending .status-dot {
  background: #bfbfbf;
}

.status-chip--completed {
  color: #1677ff;
  background: #edf5ff;
}

.status-chip--completed .status-dot {
  background: #1677ff;
}

.success-count {
  color: #1677ff;
}

.fail-count {
  color: #ff4d4f;
}

.neutral-count {
  color: #666;
}
</style>
