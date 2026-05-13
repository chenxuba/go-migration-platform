<script setup lang="ts">
import { LeftOutlined } from '@ant-design/icons-vue'
import { Empty } from 'ant-design-vue'
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  getPlatformPEP3IEPMaterialImportTaskDetailApi,
  getPlatformPEP3IEPMaterialImportTaskRecordListApi,
  type PEP3IEPMaterialImportColumn,
  type PEP3IEPMaterialImportRow,
} from '@/api/platform/scales'
import messageService from '@/utils/messageService'

const router = useRouter()
const route = useRoute()
const simpleImage = Empty.PRESENTED_IMAGE_SIMPLE
const taskId = computed(() => String(route.params.id || ''))
const loading = ref(false)
const activeTab = ref<'success' | 'fail'>('success')
const detail = reactive({
  fileName: '',
  uploadStaffName: '',
  executeStaffName: '',
  totalRows: 0,
  executedRows: 0,
  errorRows: 0,
  createdTime: '',
  completeTime: '',
})
const columns = ref<PEP3IEPMaterialImportColumn[]>([])
const successRows = ref<PEP3IEPMaterialImportRow[]>([])
const failRows = ref<PEP3IEPMaterialImportRow[]>([])
const displayedRows = computed(() => activeTab.value === 'success' ? successRows.value : failRows.value)
const tableMinWidth = computed(() => columns.value.reduce((total, column) => total + getColumnWidth(column.title), 0) + 70 + 140)

function unwrap<T>(res: any): T {
  return (res?.result ?? res?.data ?? res) as T
}

function goBack() {
  router.replace('/platform/scales/pep3-iep-materials/import/record')
}

function formatTime(value?: string) {
  return value ? value.replace('T', ' ').slice(0, 16) : '-'
}

function getColumnWidth(title: string) {
  switch (`${title || ''}`.trim()) {
    case '领域':
      return 150
    case '题目':
      return 260
    case '选项':
    case '状态':
      return 110
    case '课程形式':
      return 120
    case '训练项目':
      return 180
    case '长期目标':
    case '短期目标':
    case '训练内容':
      return 260
    default:
      return 160
  }
}

async function loadDetail() {
  loading.value = true
  try {
    const [detailRes, failRes, successRes] = await Promise.all([
      getPlatformPEP3IEPMaterialImportTaskDetailApi({ taskId: taskId.value }),
      getPlatformPEP3IEPMaterialImportTaskRecordListApi({
        queryModel: { taskId: taskId.value, type: 0 },
        pageRequestModel: { needTotal: true, pageSize: 1000, pageIndex: 1, skipCount: 0 },
      }),
      getPlatformPEP3IEPMaterialImportTaskRecordListApi({
        queryModel: { taskId: taskId.value, type: 1 },
        pageRequestModel: { needTotal: true, pageSize: 1000, pageIndex: 1, skipCount: 0 },
      }),
    ])
    const task = unwrap<any>(detailRes)
    const fail = unwrap<any>(failRes)
    const success = unwrap<any>(successRes)
    Object.assign(detail, {
      fileName: task?.fileName || '',
      uploadStaffName: task?.uploadStaffName || '',
      executeStaffName: task?.executeStaffName || '',
      totalRows: task?.totalRows || 0,
      executedRows: task?.executedRows || 0,
      errorRows: task?.errorRows || 0,
      createdTime: task?.createdTime || '',
      completeTime: task?.completeTime || '',
    })
    columns.value = success?.columns?.length ? success.columns : (fail?.columns || [])
    successRows.value = success?.list || []
    failRows.value = fail?.list || []
    if (successRows.value.length === 0 && failRows.value.length > 0)
      activeTab.value = 'fail'
  } catch (error: any) {
    messageService.error(error?.response?.data?.message || error?.message || '加载导入详情失败')
    router.replace('/platform/scales/pep3-iep-materials/import/record')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void loadDetail()
})
</script>

<template>
  <div class="record-detail-layout">
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
        <div class="record-title">
          PEP3素材导入详情
        </div>
        <div class="summary-grid">
          <div>
            <span>文件名称</span>
            <strong :title="detail.fileName">{{ detail.fileName || '-' }}</strong>
          </div>
          <div>
            <span>导入人</span>
            <strong>{{ detail.executeStaffName || detail.uploadStaffName || '-' }}</strong>
          </div>
          <div>
            <span>导入时间</span>
            <strong>{{ formatTime(detail.completeTime || detail.createdTime) }}</strong>
          </div>
          <div>
            <span>导入结果</span>
            <strong>共{{ detail.totalRows }}条，成功{{ detail.executedRows }}条，失败{{ detail.errorRows }}条</strong>
          </div>
        </div>

        <div class="tab-row">
          <div class="tabs">
            <span :class="['tab', { active: activeTab === 'success' }]" @click="activeTab = 'success'">成功({{ successRows.length }})</span>
            <span :class="['tab', { active: activeTab === 'fail' }]" @click="activeTab = 'fail'">失败({{ failRows.length }})</span>
          </div>
        </div>

        <div class="table-wrap">
          <a-spin :spinning="loading">
            <table class="detail-table" :style="{ minWidth: `${tableMinWidth}px` }">
              <colgroup>
                <col style="width: 70px">
                <col v-for="column in columns" :key="column.key" :style="{ width: `${getColumnWidth(column.title)}px` }">
                <col style="width: 140px">
              </colgroup>
              <thead>
                <tr>
                  <th class="index-column">序号</th>
                  <th v-for="column in columns" :key="column.key">{{ column.title }}</th>
                  <th>导入结果</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="row in displayedRows" :key="row.id">
                  <td class="index-column">{{ row.rowNo }}</td>
                  <td v-for="cell in row.cells" :key="cell.key">
                    <span class="cell-text">{{ cell.value || '-' }}</span>
                  </td>
                  <td>
                    <span :class="row.status === 1 ? 'success-count' : 'fail-count'">{{ row.result || '-' }}</span>
                  </td>
                </tr>
                <tr v-if="displayedRows.length === 0">
                  <td :colspan="columns.length + 2" class="empty-table-cell">
                    <a-empty :image="simpleImage" />
                  </td>
                </tr>
              </tbody>
            </table>
          </a-spin>
        </div>
      </div>
    </div>
  </div>
</template>

<style lang="less" scoped>
.record-detail-layout {
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

.record-title {
  color: #222;
  font-size: 22px;
  font-weight: 600;
  line-height: 1.3;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
  margin-top: 18px;
}

.summary-grid > div {
  min-height: 72px;
  padding: 12px 16px;
  border: 1px solid #eef0f4;
  border-radius: 8px;
  background: #fbfcff;
}

.summary-grid span {
  display: block;
  color: #667085;
  font-size: 13px;
}

.summary-grid strong {
  display: block;
  margin-top: 7px;
  overflow: hidden;
  color: #1f2937;
  font-size: 15px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tab-row {
  margin-top: 20px;
}

.tabs {
  display: flex;
  gap: 28px;
}

.tab {
  position: relative;
  padding: 8px 0;
  color: #667085;
  font-size: 16px;
  cursor: pointer;
}

.tab.active {
  color: #1677ff;
  font-weight: 600;
}

.tab.active::after {
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  height: 2px;
  border-radius: 2px;
  background: #1677ff;
  content: '';
}

.table-wrap {
  margin-top: 16px;
  max-height: calc(100vh - 330px);
  overflow: auto;
  border: 1px solid #eef0f4;
  border-radius: 8px;
}

.detail-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  table-layout: fixed;
}

.detail-table th {
  height: 44px;
  padding: 0 12px;
  border-bottom: 1px solid #e5e7eb;
  background: #f8fafc;
  color: #344054;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.5;
  text-align: left;
}

.detail-table td {
  min-height: 52px;
  padding: 10px 12px;
  border-bottom: 1px solid #f0f2f5;
  color: #1f2937;
  font-size: 14px;
  line-height: 1.5;
  vertical-align: top;
}

.cell-text {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.index-column {
  text-align: center !important;
}

.success-count {
  color: #12b76a;
}

.fail-count {
  color: #f04438;
}

.empty-table-cell {
  height: 240px;
}
</style>
