<script setup>
import dayjs from 'dayjs'
import { Empty } from 'ant-design-vue'
import messageService from '@/utils/messageService'
import {
  deletePEP3AssessmentRecordApi,
  downloadPEP3AssessmentBookletPdfApi,
  pagePEP3AssessmentRecordsApi,
} from '@/api/edu-center/pep3-assessment'
import { getScaleCategoryOptionsApi } from '@/api/teacher-center/scale-library'
import { useStudentStore } from '@/stores/student'

const props = defineProps({
  active: {
    type: Boolean,
    default: true,
  },
})

const studentStore = useStudentStore()
const displayArray = ref(['scaleCategory', 'createTime'])
const loading = ref(false)
const previewLoading = ref(false)
const exportingId = ref()
const deletingId = ref()
const dataSource = ref([])
const scaleCategoryOptions = ref([])
const currentReport = ref(null)
const reportModalOpen = ref(false)
const exportModalOpen = ref(false)
const exportTargetRecord = ref(null)
const reportPreviewUrl = ref('')
const reportPreviewRequestKey = ref(0)
const reportPdfReady = ref(false)
const simpleEmptyImage = Empty.PRESENTED_IMAGE_SIMPLE
let reportPdfReadyTimer = 0

const queryModel = reactive({
  scaleCategory: undefined,
  assessmentDateBegin: undefined,
  assessmentDateEnd: undefined,
})

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showTotal: total => `共 ${total} 条`,
})

const exportDimensionOptions = [
  {
    value: 'test_score',
    title: '仅导出测验分数',
    desc: '导出首页测验分数汇总，适合快速归档总览。',
    pages: '第 1 页',
  },
  {
    value: 'development_profile',
    title: '仅导出发展表现图',
    desc: '只导出发展表现图，用于查看各领域发展曲线。',
    pages: '第 19 页',
  },
  {
    value: 'score_and_profile',
    title: '导出测验分数与发展表现图',
    desc: '包含测验分数汇总和发展表现图，适合简版报告。',
    pages: '第 1、19 页',
    recommended: true,
  },
  {
    value: 'scoring_tables',
    title: '仅导出测验评分表',
    desc: '导出儿童表现记录、评分统计和照顾者评分表。',
    pages: '第 2-18 页',
  },
  {
    value: 'education_plan',
    title: '仅导出教育计划分析用表',
    desc: '导出教育计划分析相关页，便于教学计划制定。',
    pages: '第 20-26 页',
  },
  {
    value: 'all',
    title: '全维度导出',
    desc: '导出完整测试员记录册，包含所有维度与分析表。',
    pages: '第 1-26 页',
  },
]
const defaultExportDimension = exportDimensionOptions.find(item => item.recommended)?.value || 'all'
const selectedExportDimension = ref(defaultExportDimension)
const reportModuleValues = ['test_score', 'development_profile', 'score_and_profile', 'scoring_tables']
const reportModuleOptions = exportDimensionOptions.filter(item => reportModuleValues.includes(item.value))
const defaultReportModule = reportModuleOptions.find(item => item.recommended)?.value || reportModuleOptions[0]?.value || 'test_score'
const activeReportModule = ref(defaultReportModule)

const columns = [
  {
    title: '评估量表',
    dataIndex: 'assessmentName',
    key: 'assessmentName',
    fixed: 'left',
    width: 240,
  },
  {
    title: '评估日期',
    dataIndex: 'assessmentDate',
    key: 'assessmentDate',
    width: 140,
  },
  {
    title: '量表分类',
    dataIndex: 'scaleCategory',
    key: 'scaleCategory',
    width: 180,
  },
  {
    title: '测评年龄',
    dataIndex: 'age',
    key: 'age',
    width: 130,
  },
  {
    title: '评估老师',
    dataIndex: 'examinerName',
    key: 'examinerName',
    width: 140,
  },
  {
    title: '创建时间',
    dataIndex: 'createdTime',
    key: 'createdTime',
    width: 170,
  },
  {
    title: '操作',
    key: 'action',
    dataIndex: 'action',
    fixed: 'right',
    width: 170,
  },
]

const totalWidth = columns.reduce((acc, column) => acc + (column.width || 0), 0)
const studentId = computed(() => String(studentStore.studentId || '').trim())
const tablePagination = computed(() => ({
  ...pagination,
  showTotal: total => `共 ${total} 条`,
}))

const filterUpdateHandlers = {
  'update:scaleCategoryFilter': (val, isClearAll) => {
    queryModel.scaleCategory = isClearAll ? undefined : (val || undefined)
    reload()
  },
  'update:createTimeFilter': (val, isClearAll) => {
    const range = Array.isArray(val) ? val : []
    queryModel.assessmentDateBegin = !isClearAll && range[0] ? dayjs(range[0]).format('YYYY-MM-DD') : undefined
    queryModel.assessmentDateEnd = !isClearAll && range[1] ? dayjs(range[1]).format('YYYY-MM-DD') : undefined
    reload()
  },
}

function unwrap(res) {
  return res?.data ?? res?.result ?? res
}

function getErrorMessage(error, fallback) {
  return error?.response?.data?.message || error?.message || fallback
}

function formatDate(value) {
  if (!value)
    return '-'
  return dayjs(value).isValid() ? dayjs(value).format('YYYY-MM-DD') : value
}

function formatDateTime(value) {
  if (!value)
    return '-'
  return dayjs(value).isValid() ? dayjs(value).format('YYYY-MM-DD HH:mm') : value
}

function formatAge(row) {
  const parts = []
  if (row.ageYears)
    parts.push(`${row.ageYears}岁`)
  if (row.ageMonths)
    parts.push(`${row.ageMonths}月`)
  if (row.ageDays)
    parts.push(`${row.ageDays}天`)
  return parts.join('') || '-'
}

function exportDimensionTitle(value) {
  return exportDimensionOptions.find(item => item.value === value)?.title || '全维度导出'
}

function exportDimensionPages(value) {
  return exportDimensionOptions.find(item => item.value === value)?.pages || '第 1-26 页'
}

function exportDimensionDesc(value) {
  return exportDimensionOptions.find(item => item.value === value)?.desc || '导出完整测试员记录册，包含所有维度与分析表。'
}

function reportModuleTitle(value) {
  return reportModuleOptions.find(item => item.value === value)?.title || reportModuleOptions[0]?.title || '测验分数'
}

function reportModuleShortTitle(value) {
  const titleMap = {
    test_score: '测验分数',
    development_profile: '发展表现图',
    score_and_profile: '分数+表现图',
    scoring_tables: '评分表',
  }
  return titleMap[value] || reportModuleTitle(value)
}

function reportModuleDesc(value) {
  return reportModuleOptions.find(item => item.value === value)?.desc || ''
}

function reportModulePages(value) {
  return reportModuleOptions.find(item => item.value === value)?.pages || ''
}

function getDownloadFilename(response, fallback) {
  const disposition = response?.headers?.['content-disposition'] || response?.headers?.['Content-Disposition'] || ''
  const matched = `${disposition}`.match(/filename\*=UTF-8''([^;]+)/i) || `${disposition}`.match(/filename="?([^";]+)"?/i)
  if (!matched?.[1])
    return fallback
  try {
    return decodeURIComponent(matched[1])
  }
  catch (error) {
    return matched[1] || fallback
  }
}

function currentStudentIdValue() {
  return studentId.value || undefined
}

async function fetchScaleCategories() {
  try {
    const res = await getScaleCategoryOptionsApi()
    const list = unwrap(res) || []
    scaleCategoryOptions.value = list.map(item => ({ id: item, value: item }))
  }
  catch (error) {
    messageService.error(getErrorMessage(error, '获取量表分类失败'))
  }
}

async function fetchRecords() {
  const currentStudentId = currentStudentIdValue()
  if (!currentStudentId) {
    dataSource.value = []
    pagination.total = 0
    return
  }
  loading.value = true
  try {
    const res = await pagePEP3AssessmentRecordsApi({
      pageRequestModel: {
        pageIndex: pagination.current,
        pageSize: pagination.pageSize,
      },
      queryModel: {
        studentId: currentStudentId,
        scaleCategory: queryModel.scaleCategory,
        assessmentDateBegin: queryModel.assessmentDateBegin,
        assessmentDateEnd: queryModel.assessmentDateEnd,
      },
    })
    const data = unwrap(res)
    dataSource.value = data?.items || []
    pagination.total = data?.total || 0
  }
  catch (error) {
    messageService.error(getErrorMessage(error, '获取评估记录失败'))
    dataSource.value = []
    pagination.total = 0
  }
  finally {
    loading.value = false
  }
}

function reload() {
  pagination.current = 1
  if (props.active)
    fetchRecords()
}

function handleTableChange(page) {
  pagination.current = page.current
  pagination.pageSize = page.pageSize
  fetchRecords()
}

function resetReportPdfReady() {
  if (reportPdfReadyTimer) {
    window.clearTimeout(reportPdfReadyTimer)
    reportPdfReadyTimer = 0
  }
  reportPdfReady.value = false
}

function revokeReportPreviewUrl() {
  if (!reportPreviewUrl.value)
    return
  URL.revokeObjectURL(reportPreviewUrl.value)
  reportPreviewUrl.value = ''
  resetReportPdfReady()
}

async function loadReportPdfPreview(row = currentReport.value?.record, dimension = activeReportModule.value) {
  if (!row?.id)
    return
  const requestKey = reportPreviewRequestKey.value + 1
  reportPreviewRequestKey.value = requestKey
  resetReportPdfReady()
  previewLoading.value = true
  try {
    const response = await downloadPEP3AssessmentBookletPdfApi(row.id, dimension)
    const nextUrl = URL.createObjectURL(new Blob([response.data], { type: 'application/pdf' }))
    if (requestKey !== reportPreviewRequestKey.value) {
      URL.revokeObjectURL(nextUrl)
      return
    }
    revokeReportPreviewUrl()
    reportPreviewUrl.value = nextUrl
  }
  catch (error) {
    if (requestKey === reportPreviewRequestKey.value)
      messageService.error(getErrorMessage(error, '加载PDF预览失败'))
  }
  finally {
    if (requestKey === reportPreviewRequestKey.value)
      previewLoading.value = false
  }
}

function viewReport(row) {
  if (!row)
    return
  activeReportModule.value = defaultReportModule
  currentReport.value = {
    title: 'PEP-3测试员记录册',
    record: row,
  }
  reportModalOpen.value = true
  loadReportPdfPreview(row, defaultReportModule)
}

function selectReportModule(value) {
  if (activeReportModule.value === value)
    return
  activeReportModule.value = value
  loadReportPdfPreview()
}

function handleReportPdfFrameLoad() {
  resetReportPdfReady()
  reportPdfReadyTimer = window.setTimeout(() => {
    reportPdfReady.value = true
    reportPdfReadyTimer = 0
  }, 180)
}

function closeReportModal() {
  reportModalOpen.value = false
  reportPreviewRequestKey.value += 1
  previewLoading.value = false
  revokeReportPreviewUrl()
}

function openExportModal(row) {
  if (!row || exportingId.value)
    return
  exportTargetRecord.value = row
  selectedExportDimension.value = defaultExportDimension
  exportModalOpen.value = true
}

function closeExportModal() {
  if (exportingId.value)
    return
  exportModalOpen.value = false
}

async function exportReport(row = exportTargetRecord.value, dimension = selectedExportDimension.value) {
  if (!row)
    return
  exportingId.value = row.id
  try {
    const response = await downloadPEP3AssessmentBookletPdfApi(row.id, dimension)
    const url = URL.createObjectURL(new Blob([response.data], { type: 'application/pdf' }))
    const link = document.createElement('a')
    link.href = url
    const fallbackName = `${row.studentName || '学员'}-${row.assessmentName || '评估记录'}-${exportDimensionTitle(dimension)}-${formatDate(row.assessmentDate)}.pdf`
    link.download = getDownloadFilename(response, fallbackName)
    link.click()
    window.setTimeout(() => URL.revokeObjectURL(url), 60_000)
    if (exportModalOpen.value)
      exportModalOpen.value = false
  }
  catch (error) {
    messageService.error(getErrorMessage(error, '导出评估记录失败'))
  }
  finally {
    exportingId.value = undefined
  }
}

async function deleteRecord(row) {
  if (!row?.id)
    return
  deletingId.value = row.id
  try {
    await deletePEP3AssessmentRecordApi(row.id)
    messageService.success('评估记录已删除')
    fetchRecords()
  }
  catch (error) {
    messageService.error(getErrorMessage(error, '删除评估记录失败'))
  }
  finally {
    deletingId.value = undefined
  }
}

watch(
  [studentId, () => props.active],
  () => {
    if (props.active)
      reload()
  },
  { immediate: true },
)

onMounted(() => {
  fetchScaleCategories()
})

onBeforeUnmount(() => {
  revokeReportPreviewUrl()
  resetReportPdfReady()
})
</script>

<template>
  <div class="student-assessment-record">
    <div class="filter-wrap bg-white pl-3 pr-3 rounded-4">
      <all-filter
        :display-array="displayArray"
        :is-quick-show="false"
        :is-show-search-stu-phonefilter="false"
        :scale-category-options="scaleCategoryOptions"
        create-time-label="评估时间"
        v-on="filterUpdateHandlers"
      />
    </div>

    <div class="student-list mt-3 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            总计 {{ pagination.total }} 条
          </div>
        </div>
        <div class="table-content mt-2">
          <a-table
            :data-source="dataSource"
            :pagination="pagination.total > pagination.pageSize ? tablePagination : false"
            :columns="columns"
            :loading="loading"
            :scroll="{ x: totalWidth }"
            row-key="id"
            size="small"
            @change="handleTableChange"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'assessmentName'">
                <span class="single-line">{{ record.assessmentName || '-' }}</span>
              </template>
              <template v-else-if="column.key === 'assessmentDate'">
                {{ formatDate(record.assessmentDate) }}
              </template>
              <template v-else-if="column.key === 'scaleCategory'">
                {{ record.scaleCategory || '-' }}
              </template>
              <template v-else-if="column.key === 'age'">
                {{ formatAge(record) }}
              </template>
              <template v-else-if="column.key === 'examinerName'">
                {{ record.examinerName || '-' }}
              </template>
              <template v-else-if="column.key === 'createdTime'">
                {{ formatDateTime(record.createdTime) }}
              </template>
              <template v-else-if="column.key === 'action'">
                <a-space :size="12" class="action-links">
                  <a :class="{ disabled: previewLoading }" @click="viewReport(record)">查看</a>
                  <a-popconfirm title="确认删除这条评估记录？" ok-text="删除" cancel-text="取消" @confirm="deleteRecord(record)">
                    <a :class="{ disabled: deletingId === record.id }">删除</a>
                  </a-popconfirm>
                  <a :class="{ disabled: exportingId === record.id }" @click="openExportModal(record)">导出</a>
                </a-space>
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>

    <a-modal
      v-model:open="reportModalOpen"
      width="842px"
      :centered="true"
      :footer="null"
      wrap-class-name="student-assessment-report-modal"
      @cancel="closeReportModal"
    >
      <template #title>
        <div class="report-modal-title">
          <span>评估报告</span>
          <small>按记录册导出维度查看报告内容</small>
        </div>
      </template>
      <div v-if="currentReport" class="report-preview">
        <div class="report-head">
          <div class="report-info">
            <div class="report-title">
              {{ currentReport.title || currentReport.record?.assessmentName || '评估报告' }}
            </div>
            <div class="report-subtitle">
              {{ currentReport.record?.studentName || '-' }} / {{ formatDate(currentReport.record?.assessmentDate) }}
            </div>
          </div>
          <a-button
            type="primary"
            size="small"
            class="report-export-btn"
            :loading="exportingId === currentReport.record?.id"
            @click="exportReport(currentReport.record, activeReportModule)"
          >
            导出
          </a-button>
        </div>
        <div class="report-module-area">
          <div class="report-module-grid">
            <button
              v-for="option in reportModuleOptions"
              :key="option.value"
              type="button"
              class="report-module-chip"
              :class="{ 'report-module-chip--active': activeReportModule === option.value }"
              :title="option.title"
              @click="selectReportModule(option.value)"
            >
              <span class="report-module-chip__dot" />
              <span class="report-module-chip__text">{{ reportModuleShortTitle(option.value) }}</span>
              <span v-if="option.recommended" class="report-module-chip__tag">推荐</span>
            </button>
          </div>
          <div class="report-module-summary">
            <strong>{{ reportModulePages(activeReportModule) }}</strong>
            <span>{{ reportModuleDesc(activeReportModule) }}</span>
          </div>
        </div>

        <div class="report-module-content">
          <div class="report-pdf-shell">
            <iframe
              v-if="reportPreviewUrl"
              :key="reportPreviewUrl"
              class="report-pdf-frame"
              :class="{ 'report-pdf-frame--ready': reportPdfReady }"
              :src="`${reportPreviewUrl}#toolbar=0&navpanes=0`"
              title="PEP-3记录册PDF预览"
              @load="handleReportPdfFrameLoad"
            />
            <div v-if="previewLoading || (reportPreviewUrl && !reportPdfReady)" class="report-pdf-loading">
              <a-spin size="small" />
              <span>PDF加载中...</span>
            </div>
            <a-empty
              v-if="!reportPreviewUrl && !previewLoading"
              class="report-pdf-empty"
              description="暂无PDF预览"
              :image="simpleEmptyImage"
              :image-style="{ height: '48px' }"
            />
          </div>
        </div>
      </div>
    </a-modal>

    <a-modal
      v-model:open="exportModalOpen"
      width="700px"
      :centered="true"
      :footer="null"
      wrap-class-name="student-assessment-export-modal"
      :mask-closable="!exportingId"
      @cancel="closeExportModal"
    >
      <template #title>
        <div class="export-modal-title">
          <span>导出记录册</span>
          <small>选择本次导出的内容范围</small>
        </div>
      </template>
      <div class="export-dimension">
        <div class="export-dimension__list">
          <button
            v-for="option in exportDimensionOptions"
            :key="option.value"
            type="button"
            class="export-dimension-chip"
            :class="{ 'export-dimension-chip--active': selectedExportDimension === option.value }"
            :disabled="!!exportingId"
            @click="selectedExportDimension = option.value"
          >
            <span class="export-dimension-chip__title">{{ option.title }}</span>
            <span v-if="option.recommended" class="export-dimension-chip__tag">推荐</span>
            <span class="export-dimension-chip__pages">{{ option.pages }}</span>
          </button>
        </div>
        <div class="export-dimension__footer">
          <div class="export-dimension__current">
            <strong>{{ exportDimensionTitle(selectedExportDimension) }}</strong>
            <em>{{ exportDimensionPages(selectedExportDimension) }}</em>
          </div>
          <div class="export-dimension__actions">
            <a-button :disabled="!!exportingId" @click="closeExportModal">
              取消
            </a-button>
            <a-button type="primary" :loading="!!exportingId" @click="exportReport()">
              导出
            </a-button>
          </div>
        </div>
      </div>
    </a-modal>
  </div>
</template>

<style lang="less" scoped>
.student-assessment-record {
  padding: 12px 0 0;
}

.total {
  position: relative;
  display: flex;
  align-items: center;
  padding-left: 10px;
  color: #222;

  &::before {
    position: absolute;
    left: 0;
    display: inline-block;
    width: 4px;
    height: 12px;
    content: "";
    background: var(--pro-ant-color-primary);
    border-radius: 2px;
  }
}

.single-line {
  display: inline-block;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  vertical-align: bottom;
  white-space: nowrap;
}

.action-links {
  white-space: nowrap;

  a {
    color: var(--pro-ant-color-primary);
    cursor: pointer;

    &.disabled {
      color: #b8c2d0;
      pointer-events: none;
    }
  }
}

.report-modal-title,
.export-modal-title {
  display: flex;
  flex-direction: column;
  gap: 2px;

  span {
    color: #1f2937;
    font-size: 18px;
    font-weight: 600;
    line-height: 26px;
  }

  small {
    color: #8a94a6;
    font-size: 12px;
    font-weight: 400;
    line-height: 18px;
  }
}

.report-preview {
  padding: 14px 24px 0;
}

.report-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 8px 0;
  border-bottom: 1px solid #edf1f6;
}

.report-info {
  display: flex;
  align-items: baseline;
  flex: 1 1 auto;
  gap: 10px;
  min-width: 0;
}

.report-title {
  min-width: 0;
  overflow: hidden;
  color: #1f2937;
  font-size: 15px;
  font-weight: 600;
  line-height: 22px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.report-subtitle {
  flex: 0 0 auto;
  color: #7a8494;
  font-size: 13px;
  line-height: 20px;
  white-space: nowrap;
}

.report-export-btn {
  flex: 0 0 auto;
  min-width: 56px;
  height: 28px;
  padding: 0 12px;
}

.report-module-area {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 10px;
  padding: 8px 10px;
  background: #f8fafc;
  border: 1px solid #edf1f6;
  border-radius: 8px;
}

.report-module-grid {
  display: flex;
  flex: 0 0 auto;
  gap: 6px;
  max-width: 100%;
  overflow-x: auto;
  scrollbar-width: none;

  &::-webkit-scrollbar {
    display: none;
  }
}

.report-module-chip {
  display: flex;
  align-items: center;
  flex: 0 0 auto;
  gap: 7px;
  min-width: 0;
  height: 32px;
  padding: 0 10px;
  color: inherit;
  font: inherit;
  text-align: left;
  cursor: pointer;
  background: #fff;
  border: 1px solid #e7edf5;
  border-radius: 6px;
  transition: background 0.16s ease, border-color 0.16s ease, box-shadow 0.16s ease;

  &:hover {
    background: #fbfdff;
    border-color: #bfd9ff;
  }
}

.report-module-chip--active {
  background: #f7fbff;
  border-color: #7dbbff;
  box-shadow: 0 2px 8px rgba(24, 144, 255, 0.08);
}

.report-module-chip__dot {
  flex: 0 0 auto;
  width: 6px;
  height: 6px;
  background: #cbd5e1;
  border-radius: 50%;
}

.report-module-chip--active .report-module-chip__dot {
  background: var(--pro-ant-color-primary);
  box-shadow: 0 0 0 3px rgba(24, 144, 255, 0.12);
}

.report-module-chip__text {
  flex: 1 1 auto;
  min-width: 0;
  overflow: hidden;
  color: #334155;
  font-size: 13px;
  font-weight: 500;
  line-height: 20px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.report-module-chip--active .report-module-chip__text {
  color: var(--pro-ant-color-primary);
  font-weight: 600;
}

.report-module-chip__tag {
  flex: 0 0 auto;
  padding: 0 5px;
  color: var(--pro-ant-color-primary);
  font-size: 12px;
  line-height: 18px;
  background: #eef6ff;
  border-radius: 4px;
}

.report-module-summary {
  display: flex;
  align-items: center;
  flex: 1 1 auto;
  gap: 8px;
  min-width: 0;
  padding-left: 12px;
  border-left: 1px solid #e6edf6;

  span {
    flex: 1 1 auto;
    min-width: 0;
    overflow: hidden;
    color: #687386;
    font-size: 12px;
    line-height: 18px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  strong {
    flex: 0 0 auto;
    overflow: hidden;
    color: var(--pro-ant-color-primary);
    font-size: 13px;
    font-weight: 600;
    line-height: 20px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.report-module-content {
  padding: 16px 0 22px;
}

.report-pdf-shell {
  position: relative;
  min-height: 620px;
  overflow: hidden;
  background: #fff;
  border: 1px solid #edf1f6;
  border-radius: 8px;
}

.report-pdf-frame {
  display: block;
  width: 100%;
  height: min(72vh, 760px);
  min-height: 620px;
  opacity: 0;
  background: #fff;
  border: 0;
  transition: opacity 0.16s ease;
}

.report-pdf-frame--ready {
  opacity: 1;
}

.report-pdf-loading {
  position: absolute;
  inset: 0;
  z-index: 2;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  gap: 10px;
  box-sizing: border-box;
  padding-top: 160px;
  color: #7a8494;
  font-size: 13px;
  background: #fff;
}

.report-pdf-empty {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: flex-start;
  flex-direction: column;
  margin: 0;
  padding-top: 96px;
  box-sizing: border-box;

  :deep(.ant-empty-image) {
    margin-bottom: 10px;
  }

  :deep(.ant-empty-description) {
    color: #6b7280;
    font-size: 14px;
    line-height: 22px;
  }
}

.export-dimension {
  padding: 18px 24px 20px;
}

.export-dimension__list {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
}

.export-dimension-chip {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
  height: 42px;
  padding: 0 12px;
  color: inherit;
  font: inherit;
  text-align: left;
  cursor: pointer;
  background: #fff;
  border: 1px solid #e7edf5;
  border-radius: 8px;

  &:hover {
    border-color: #bfd9ff;
  }
}

.export-dimension-chip--active {
  background: #f7fbff;
  border-color: #7dbbff;
}

.export-dimension-chip__title {
  flex: 1 1 auto;
  min-width: 0;
  overflow: hidden;
  color: #334155;
  font-size: 13px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.export-dimension-chip__tag {
  flex: 0 0 auto;
  padding: 0 5px;
  color: var(--pro-ant-color-primary);
  font-size: 12px;
  line-height: 18px;
  background: #eef6ff;
  border-radius: 4px;
}

.export-dimension-chip__pages {
  flex: 0 0 auto;
  color: #8a94a6;
  font-size: 12px;
}

.export-dimension__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-top: 16px;
  padding-top: 14px;
  border-top: 1px solid #eef1f5;
}

.export-dimension__current {
  display: flex;
  align-items: center;
  min-width: 0;

  strong {
    overflow: hidden;
    color: #1f2937;
    font-weight: 600;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  em {
    flex: 0 0 auto;
    margin-left: 10px;
    padding-left: 10px;
    color: #9aa4b2;
    font-style: normal;
    border-left: 1px solid #e2e8f0;
  }
}

.export-dimension__actions {
  flex: 0 0 auto;
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>

<style lang="less">
.student-assessment-report-modal,
.student-assessment-export-modal {
  .ant-modal-content {
    padding: 0;
    overflow: hidden;
    border-radius: 12px;
    box-shadow: 0 10px 32px rgba(15, 23, 42, 0.14);
  }

  .ant-modal-header {
    padding: 20px 24px 14px;
    margin: 0;
    border-bottom: 1px solid #eef1f5;
  }

  .ant-modal-title {
    margin: 0;
  }

  .ant-modal-close {
    top: 18px;
    inset-inline-end: 18px;
    color: #8a94a6;
  }

  .ant-modal-body {
    padding: 0;
  }
}

.student-assessment-report-modal {
  .ant-modal {
    max-width: calc(100vw - 48px);
  }

  .ant-modal-body {
    max-height: calc(100vh - 150px);
    overflow-y: auto;
    scrollbar-color: #c6d1df transparent;
    scrollbar-width: thin;

    &::-webkit-scrollbar {
      width: 10px;
    }

    &::-webkit-scrollbar-track {
      background: transparent;
    }

    &::-webkit-scrollbar-thumb {
      background: #c6d1df;
      background-clip: padding-box;
      border: 2px solid transparent;
      border-radius: 999px;
    }
  }
}
</style>
