<script setup lang="ts">
import { Empty } from 'ant-design-vue'
import dayjs from 'dayjs'
import { computed, ref, watch } from 'vue'
import { downloadClassRecordExportRecordApi, getClassRecordExportRecordsApi, type ClassRecordExportRecord } from '@/api/edu-center/class-record'
import messageService from '@/utils/messageService'

const props = withDefaults(defineProps<{
  open?: boolean
  exportType?: string
}>(), {
  open: false,
  exportType: 'student',
})

const emit = defineEmits(['update:open'])
const simpleImage = Empty.PRESENTED_IMAGE_SIMPLE

const modalOpen = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const loading = ref(false)
const records = ref<ClassRecordExportRecord[]>([])

function formatTime(value?: string) {
  if (!value)
    return '-'
  const date = dayjs(value)
  return date.isValid() ? date.format('YYYY-MM-DD HH:mm') : value
}

function resolveTitle() {
  return props.exportType === 'schedule' ? '按日程导出记录' : '按学员导出记录'
}

function triggerBlobDownload(response: any, fallbackName: string) {
  const blob = new Blob([response.data], { type: response.headers['content-type'] || 'application/octet-stream' })
  const disposition = response.headers['content-disposition'] || ''
  const matched = disposition.match(/filename\*=UTF-8''([^;]+)/i)
  const fileName = matched ? decodeURIComponent(matched[1]) : fallbackName
  const url = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = fileName
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  window.URL.revokeObjectURL(url)
}

async function loadRecords() {
  if (!modalOpen.value)
    return
  loading.value = true
  try {
    const res = await getClassRecordExportRecordsApi(String(props.exportType || 'student'))
    if (res.code !== 200) {
      throw new Error(res.message || '获取导出记录失败')
    }
    records.value = Array.isArray(res.result) ? res.result : []
  }
  catch (error: any) {
    records.value = []
    messageService.error(error?.response?.data?.message || error?.message || '获取导出记录失败')
  }
  finally {
    loading.value = false
  }
}

async function handleDownload(record: ClassRecordExportRecord) {
  try {
    const response = await downloadClassRecordExportRecordApi(record.id, String(props.exportType || 'student'))
    triggerBlobDownload(response, record.fileName || `上课记录导出-${dayjs().format('YYYYMMDDHHmmss')}.xlsx`)
  }
  catch (error: any) {
    messageService.error(error?.response?.data?.message || error?.message || '下载失败，请稍后重试')
  }
}

watch(
  () => `${modalOpen.value}|${props.exportType}`,
  () => {
    if (!modalOpen.value) {
      records.value = []
      loading.value = false
      return
    }
    loadRecords()
  },
  { immediate: true },
)
</script>

<template>
  <a-modal
    v-model:open="modalOpen"
    :title="resolveTitle()"
    :footer="null"
    :width="820"
    destroy-on-close
  >
    <a-spin :spinning="loading">
      <div v-if="records.length" class="export-record-list">
        <div v-for="record in records" :key="record.id" class="export-record-card">
          <div class="export-record-card__header">
            <div class="export-record-card__title">
              {{ record.fileName || '-' }}
            </div>
            <a-button type="link" class="p-0" @click="handleDownload(record)">
              下载
            </a-button>
          </div>
          <div class="export-record-card__meta">
            <span>导出人：{{ record.exporterName || '-' }}</span>
            <span>导出条数：{{ record.totalRows || 0 }}</span>
            <span>导出时间：{{ formatTime(record.createdTime) }}</span>
          </div>
          <div v-if="record.queryConditions?.length" class="export-record-card__conditions">
            <span
              v-for="item in record.queryConditions"
              :key="`${record.id}-${item.label}-${item.value}`"
              class="export-record-tag"
            >
              {{ item.label }}：{{ item.value }}
            </span>
          </div>
        </div>
      </div>
      <a-empty v-else :image="simpleImage" description="暂无导出记录" />
    </a-spin>
  </a-modal>
</template>

<style lang="less" scoped>
.export-record-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.export-record-card {
  border: 1px solid #eef1f6;
  border-radius: 12px;
  padding: 14px 16px;
  background: #fff;
}

.export-record-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.export-record-card__title {
  color: #1f2329;
  font-weight: 600;
  line-height: 22px;
  word-break: break-all;
}

.export-record-card__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 20px;
  margin-top: 8px;
  color: #4e5969;
  font-size: 13px;
}

.export-record-card__conditions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 10px;
}

.export-record-tag {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  border-radius: 999px;
  background: #f5f7fb;
  color: #4e5969;
  font-size: 12px;
  line-height: 18px;
}
</style>
