<script setup lang="ts">
import { LeftOutlined } from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'
import { onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  clearPlatformPEP3IEPMaterialImportTaskListApi,
  deletePlatformPEP3IEPMaterialImportTaskApi,
  getPlatformPEP3IEPMaterialImportTaskListApi,
  type PEP3IEPMaterialImportTaskDetail,
} from '@/api/platform/scales'
import messageService from '@/utils/messageService'

const router = useRouter()
const loading = ref(false)
const records = ref<PEP3IEPMaterialImportTaskDetail[]>([])
let pollingTimer: number | null = null

function unwrap<T>(res: any): T {
  return (res?.result ?? res?.data ?? res) as T
}

function goBack() {
  router.replace('/platform/scales/pep3-iep-materials/import')
}

function viewDetail(record: PEP3IEPMaterialImportTaskDetail) {
  router.push(`/platform/scales/pep3-iep-materials/import/record/${record.id}`)
}

function statusText(status: number) {
  if (status === 4)
    return '导入中'
  return status === 3 ? '待处理' : '已完成'
}

function formatTime(value?: string) {
  return value ? value.replace('T', ' ').slice(0, 16) : '-'
}

async function loadRecords() {
  loading.value = true
  try {
    const data = unwrap<{ list: PEP3IEPMaterialImportTaskDetail[] }>(await getPlatformPEP3IEPMaterialImportTaskListApi())
    records.value = data?.list || []
  } catch (error: any) {
    messageService.error(error?.response?.data?.message || error?.message || '加载导入记录失败')
  } finally {
    loading.value = false
  }
}

function handleClearRecords() {
  Modal.confirm({
    title: '确认清空导入记录？',
    centered: true,
    okText: '确认清空',
    okType: 'danger',
    cancelText: '取消',
    content: '将清空平台端全部PEP3素材导入记录，该操作不可恢复。',
    async onOk() {
      try {
        await clearPlatformPEP3IEPMaterialImportTaskListApi()
        messageService.success('导入记录已清空')
        await loadRecords()
      } catch (error: any) {
        messageService.error(error?.response?.data?.message || error?.message || '清空失败，请稍后重试')
        return Promise.reject(error)
      }
    },
  })
}

function handleDeleteRecord(record: PEP3IEPMaterialImportTaskDetail) {
  Modal.confirm({
    title: '确认删除这条导入记录？',
    centered: true,
    okText: '确认删除',
    okType: 'danger',
    cancelText: '取消',
    content: '删除后将无法恢复该待处理导入任务。',
    async onOk() {
      try {
        await deletePlatformPEP3IEPMaterialImportTaskApi({ taskId: record.id })
        messageService.success('导入记录已删除')
        await loadRecords()
      } catch (error: any) {
        messageService.error(error?.response?.data?.message || error?.message || '删除失败，请稍后重试')
        return Promise.reject(error)
      }
    },
  })
}

onMounted(() => {
  void loadRecords()
  pollingTimer = window.setInterval(() => {
    if (records.value.some(item => item.status === 4))
      void loadRecords()
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
    <div class="work-top">
      <div class="work-top-left">
        <div class="import-header-logo" title="导入中心" aria-hidden="true" />
        <span class="back-link" @click="goBack">
          <LeftOutlined /> 返回
        </span>
      </div>
      <div class="work-top-right">
        当前端：平台总控
      </div>
    </div>

    <div class="work-main">
      <div class="work-main-card">
        <div class="record-header">
          <div class="record-title">
            PEP3素材导入记录
          </div>
          <a-button danger ghost @click="handleClearRecords">
            清空记录
          </a-button>
        </div>

        <a-table
          :loading="loading"
          :data-source="records"
          :pagination="{ pageSize: 10, hideOnSinglePage: true }"
          :scroll="{ x: 1120, y: 'calc(100vh - 252px)' }"
          row-key="id"
          size="middle"
          class="mt-24px"
        >
          <a-table-column title="文件名称" data-index="fileName" key="fileName" width="360">
            <template #default="{ record }">
              <a-tooltip :title="record.fileName">
                <span class="file-name-cell">{{ record.fileName || '-' }}</span>
              </a-tooltip>
            </template>
          </a-table-column>
          <a-table-column title="状态" key="status" width="120">
            <template #default="{ record }">
              <span :class="['status-dot', { running: record.status === 4, pending: record.status === 3 }]" />
              {{ statusText(record.status) }}
            </template>
          </a-table-column>
          <a-table-column title="导入时间" key="createdTime" width="170">
            <template #default="{ record }">
              {{ formatTime(record.createdTime) }}
            </template>
          </a-table-column>
          <a-table-column title="导入人" data-index="uploadStaffName" key="uploadStaffName" width="150" />
          <a-table-column title="结果" key="result" width="260">
            <template #default="{ record }">
              <template v-if="record.status === 4">
                已导入{{ (record.executedRows || 0) + (record.errorRows || 0) }}/{{ record.totalRows || 0 }}
              </template>
              <template v-else>
                导入共计{{ record.totalRows || 0 }}（<span :class="record.executedRows > 0 ? 'success-count' : 'neutral-count'">成功{{ record.executedRows || 0 }}</span>/<span :class="record.errorRows > 0 ? 'fail-count' : 'neutral-count'">失败{{ record.errorRows || 0 }}</span>）
              </template>
            </template>
          </a-table-column>
          <a-table-column title="操作" key="action" width="120" fixed="right">
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
  height: 100vh;
  min-height: 100vh;
  overflow: hidden;
  background: #f7f7fd;
}

.work-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 52px;
  background: #fff;
}

.work-top-left {
  display: flex;
  align-items: center;
}

.work-top-right {
  padding-right: 24px;
  color: #000;
  font-size: 15px;
  font-weight: 500;
}

.import-header-logo {
  position: relative;
  width: 52px;
  height: 52px;
  flex-shrink: 0;
  overflow: hidden;
  background: linear-gradient(145deg, #2b8cff 0%, #0066ff 45%, #0050d8 100%);
}

.import-header-logo::before {
  position: absolute;
  top: 14px;
  left: 11px;
  width: 30px;
  height: 24px;
  background: #fff;
  box-shadow: inset 0 -8px 0 rgba(0, 102, 255, 0.12), inset 0 -16px 0 rgba(0, 102, 255, 0.08);
  content: '';
}

.back-link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  margin-left: 14px;
  color: #06f;
  font-size: 18px;
  font-weight: 500;
  cursor: pointer;
}

.work-main {
  box-sizing: border-box;
  display: flex;
  justify-content: center;
  height: calc(100vh - 52px);
  min-width: 1040px;
  padding: 24px 40px;
}

.work-main-card {
  box-sizing: border-box;
  width: min(1260px, 100%);
  height: 100%;
  min-height: 0;
  overflow: hidden;
  padding: 28px 48px 30px;
  border-radius: 16px;
  background: #fff;
  box-shadow: 0 12px 32px rgba(15, 35, 80, 0.08);
}

.record-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.record-title {
  color: #222;
  font-size: 22px;
  font-weight: 600;
  line-height: 1.3;
}

.work-main-card :deep(.ant-table-wrapper) {
  margin-top: 18px !important;
}

.work-main-card :deep(.ant-table) {
  table-layout: fixed;
}

.work-main-card :deep(.ant-table-thead > tr > th) {
  background: #f8fafc;
  color: #344054;
  font-size: 14px;
  font-weight: 600;
}

.work-main-card :deep(.ant-table-cell) {
  height: 50px;
  color: #1f2937;
  font-size: 14px;
  line-height: 1.5;
  white-space: nowrap;
}

.file-name-cell {
  display: block;
  overflow: hidden;
  max-width: 100%;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.status-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  margin-right: 8px;
  border-radius: 50%;
  background: #12b76a;
}

.status-dot.running {
  background: #1677ff;
}

.status-dot.pending {
  background: #faad14;
}

.success-count {
  color: #12b76a;
}

.fail-count {
  color: #f04438;
}

.neutral-count {
  color: #667085;
}

@media (max-height: 820px) {
  .work-main {
    padding: 18px 36px;
  }

  .work-main-card {
    padding: 24px 44px 26px;
  }
}
</style>
