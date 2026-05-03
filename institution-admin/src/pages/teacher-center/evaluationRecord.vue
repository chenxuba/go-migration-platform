<script setup>
import dayjs from 'dayjs'
import { Empty } from 'ant-design-vue'
import messageService from '@/utils/messageService'
import { useTableColumns } from '@/composables/useTableColumns'
import GenerateIepModal from './components/generate-iep-modal.vue'
import {
  deletePEP3AssessmentRecordApi,
  downloadPEP3AssessmentBookletPdfApi,
  pagePEP3AssessmentRecordsApi,
} from '@/api/edu-center/pep3-assessment'
import { getScaleCategoryOptionsApi } from '@/api/teacher-center/scale-library'

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
const iepModalOpen = ref(false)
const iepTargetRecord = ref(null)
const reportPreviewUrl = ref('')
const reportPreviewRequestKey = ref(0)
const reportPdfReady = ref(false)
const simpleEmptyImage = Empty.PRESENTED_IMAGE_SIMPLE
let reportPdfReadyTimer = 0

const exportDimensionOptions = [
  {
    value: 'test_score',
    title: '仅导出测验分数',
    badge: '01',
    desc: '导出首页测验分数汇总，适合快速归档总览。',
    pages: '第 1 页',
  },
  {
    value: 'development_profile',
    title: '仅导出发展表现图',
    badge: '02',
    desc: '只导出发展表现图，用于查看各领域发展曲线。',
    pages: '第 19 页',
  },
  {
    value: 'score_and_profile',
    title: '导出测验分数与发展表现图',
    badge: '03',
    desc: '包含测验分数汇总和发展表现图，适合简版报告。',
    pages: '第 1、19 页',
    recommended: true,
  },
  {
    value: 'scoring_tables',
    title: '仅导出测验评分表',
    badge: '04',
    desc: '导出儿童表现记录、评分统计和照顾者评分表。',
    pages: '第 2-18 页',
  },
  {
    value: 'education_plan',
    title: '仅导出教育计划分析用表',
    badge: '05',
    desc: '导出教育计划分析相关页，便于教学计划制定。',
    pages: '第 20-26 页',
  },
  {
    value: 'all',
    title: '全维度导出',
    badge: 'ALL',
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

const queryModel = reactive({
  scaleCategory: undefined,
  studentId: undefined,
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

const allColumns = ref([
  {
    title: '学员/性别',
    dataIndex: 'student',
    key: 'student',
    fixed: 'left',
    width: 210,
    required: true,
  },
  {
    title: '评估量表',
    dataIndex: 'assessmentName',
    key: 'assessmentName',
    width: 220,
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
    width: 160,
  },
  {
    title: '测评年龄',
    dataIndex: 'age',
    key: 'age',
    width: 120,
  },
  {
    title: '评估老师',
    dataIndex: 'examinerName',
    key: 'examinerName',
    width: 130,
  },
  {
    title: '创建时间',
    dataIndex: 'createdTime',
    key: 'createdTime',
    width: 160,
  },
  {
    title: '操作',
    key: 'action',
    dataIndex: 'action',
    fixed: 'right',
    width: 240,
  },
])

const { selectedValues, columnOptions, filteredColumns, totalWidth } = useTableColumns({
  storageKey: 'evaluationRecord',
  allColumns,
  excludeKeys: ['action'],
})

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
  'update:stuPhoneSearchFilter': (val, isClearAll) => {
    queryModel.studentId = isClearAll || Array.isArray(val) ? undefined : (val || undefined)
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

function formatCurrentAge(row) {
  const birth = dayjs(row?.birthDate).startOf('day')
  const today = dayjs().startOf('day')
  if (!birth.isValid() || birth.isAfter(today))
    return formatAge(row)

  const years = today.diff(birth, 'year')
  const afterYears = birth.add(years, 'year')
  const months = today.diff(afterYears, 'month')
  const afterMonths = afterYears.add(months, 'month')
  const days = today.diff(afterMonths, 'day')

  const parts = []
  if (years)
    parts.push(`${years}岁`)
  if (months)
    parts.push(`${months}月`)
  if (days || !parts.length)
    parts.push(`${days}天`)
  return parts.join('')
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

function iepActionText(record) {
  return record?.iepPlanStatus === 'confirmed' ? '查看IEP' : '生成IEP'
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

function reload() {
  pagination.current = 1
  fetchRecords()
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
  loading.value = true
  try {
    const res = await pagePEP3AssessmentRecordsApi({
      pageRequestModel: {
        pageIndex: pagination.current,
        pageSize: pagination.pageSize,
      },
      queryModel: {
        scaleCategory: queryModel.scaleCategory,
        studentId: queryModel.studentId,
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
  }
  finally {
    loading.value = false
  }
}

function handleTableChange(page) {
  pagination.current = page.current
  pagination.pageSize = page.pageSize
  fetchRecords()
}

async function viewReport(row) {
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

function openIepModal(row) {
  if (!row)
    return
  iepTargetRecord.value = row
  iepModalOpen.value = true
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
  deletingId.value = row.id
  try {
    await deletePEP3AssessmentRecordApi(row.id)
    messageService.success('评估记录已删除')
    if (dataSource.value.length === 1 && pagination.current > 1)
      pagination.current -= 1
    fetchRecords()
  }
  catch (error) {
    messageService.error(getErrorMessage(error, '删除评估记录失败'))
  }
  finally {
    deletingId.value = undefined
  }
}

onMounted(() => {
  fetchScaleCategories()
  fetchRecords()
})

onBeforeUnmount(() => {
  revokeReportPreviewUrl()
  resetReportPdfReady()
})
</script>

<template>
  <div>
    <div class="filter-wrap bg-white pl-3 pr-3 rounded-4">
      <all-filter
        :display-array="displayArray"
        :is-quick-show="false"
        :is-show-search-stu-phonefilter="true"
        :scale-category-options="scaleCategoryOptions"
        create-time-label="评估时间"
        v-on="filterUpdateHandlers"
      />
    </div>

    <div class="student-list mt-2 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            总计 {{ pagination.total }} 条
          </div>
          <div class="edit flex">
            <customize-code
              v-model:checked-values="selectedValues"
              :options="columnOptions"
              :total="allColumns.length"
              :num="selectedValues.length"
            />
          </div>
        </div>
        <div class="table-content mt-2">
          <a-table
            :data-source="dataSource"
            :pagination="pagination"
            :columns="filteredColumns"
            :loading="loading"
            :scroll="{ x: totalWidth }"
            row-key="id"
            size="small"
            @change="handleTableChange"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'student'">
                <student-avatar
                  :id="record.studentId"
                  :name="record.studentName || '-'"
                  :gender="record.studentGender || '-'"
                  :age="formatCurrentAge(record)"
                  :avatar-url="record.studentAvatar"
                  default-active-key="0"
                />
              </template>
              <template v-else-if="column.key === 'assessmentName'">
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
                  <a @click="openIepModal(record)">{{ iepActionText(record) }}</a>
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
      wrap-class-name="pep3-report-modal"
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
      wrap-class-name="pep3-export-dimension-modal"
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
        <div class="export-dimension__summary">
          <div class="export-dimension__file">
            PDF
          </div>
          <div class="export-dimension__record">
            <div class="export-dimension__name">
              {{ exportTargetRecord?.studentName || '学员' }}
            </div>
            <div class="export-dimension__meta">
              {{ exportTargetRecord?.assessmentName || '评估记录' }} · {{ formatDate(exportTargetRecord?.assessmentDate) }}
            </div>
          </div>
          <div class="export-dimension__current">
            <span>可选范围</span>
            <strong>{{ exportDimensionOptions.length }} 项</strong>
          </div>
        </div>
        <div class="export-dimension__chooser">
          <div class="export-dimension__matrix">
            <button
              v-for="option in exportDimensionOptions"
              :key="option.value"
              type="button"
              class="export-dimension-chip"
              :class="{ 'export-dimension-chip--active': selectedExportDimension === option.value }"
              :aria-pressed="selectedExportDimension === option.value"
              :title="option.title"
              :disabled="!!exportingId"
              @click="selectedExportDimension = option.value"
            >
              <span class="export-dimension-chip__dot" />
              <span class="export-dimension-chip__text">{{ option.title }}</span>
              <span v-if="option.recommended" class="export-dimension-chip__tag">推荐</span>
            </button>
          </div>
          <div class="export-dimension__detail">
            <span class="export-dimension__detail-label">导出内容</span>
            <strong>{{ exportDimensionTitle(selectedExportDimension) }}</strong>
            <p>{{ exportDimensionDesc(selectedExportDimension) }}</p>
            <div class="export-dimension__detail-pages">
              <span>包含页码</span>
              <em>{{ exportDimensionPages(selectedExportDimension) }}</em>
            </div>
          </div>
        </div>
        <div class="export-dimension__footer">
          <div class="export-dimension__selection">
            <span>将导出</span>
            <strong>{{ exportDimensionTitle(selectedExportDimension) }}</strong>
            <em>{{ exportDimensionPages(selectedExportDimension) }}</em>
          </div>
          <div class="export-dimension__actions">
            <a-button :disabled="!!exportingId" @click="closeExportModal">
              取消
            </a-button>
            <a-button type="primary" :loading="!!exportingId" @click="exportReport()">
              开始导出
            </a-button>
          </div>
        </div>
      </div>
    </a-modal>

    <generate-iep-modal
      v-model:open="iepModalOpen"
      :record="iepTargetRecord"
      @saved="fetchRecords"
      @confirmed="fetchRecords"
    />
  </div>
</template>

<style lang="less" scoped>
.total {
  position: relative;
  padding-left: 10px;
  color: #222;
  display: flex;
  align-items: center;

  &::before {
    display: inline-block;
    background: var(--pro-ant-color-primary);
    border-radius: 2px;
    content: "";
    height: 12px;
    left: 0;
    position: absolute;
    width: 4px;
  }
}

.single-line {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.action-links {
  white-space: nowrap;
}

.disabled {
  pointer-events: none;
  color: #999;
}

.report-modal-title {
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
  scrollbar-color: #c6d1df transparent;
  scrollbar-width: thin;
  transition: opacity 0.16s ease;

  &::-webkit-scrollbar {
    width: 10px;
    height: 10px;
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

  &::-webkit-scrollbar-thumb:hover {
    background: #aebccc;
    background-clip: padding-box;
    border: 2px solid transparent;
  }
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
  padding-top: 160px;
  box-sizing: border-box;
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

.export-modal-title {
  display: flex;
  flex-direction: column;
  gap: 2px;

  span {
    color: #1f2937;
    font-size: 16px;
    font-weight: 600;
    line-height: 24px;
  }

  small {
    color: #8a94a6;
    font-size: 12px;
    font-weight: 400;
    line-height: 18px;
  }
}

.export-dimension {
  padding: 16px 20px 0;
}

.export-dimension__summary {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  margin-bottom: 14px;
  background: #f8fafc;
  border: 1px solid #edf1f6;
  border-radius: 8px;
}

.export-dimension__file {
  flex: 0 0 auto;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 42px;
  height: 42px;
  color: var(--pro-ant-color-primary);
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0;
  background: #eef6ff;
  border: 1px solid #d9eaff;
  border-radius: 8px;
}

.export-dimension__record {
  flex: 1 1 auto;
  min-width: 0;
}

.export-dimension__name {
  overflow: hidden;
  color: #1f2937;
  font-size: 15px;
  font-weight: 600;
  line-height: 22px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.export-dimension__meta {
  margin-top: 2px;
  overflow: hidden;
  color: #7a8494;
  font-size: 12px;
  line-height: 18px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.export-dimension__current {
  flex: 0 0 auto;
  display: flex;
  flex-direction: column;
  min-width: 138px;
  padding-left: 14px;
  border-left: 1px solid #e5eaf2;

  span {
    color: #9aa4b2;
    font-size: 12px;
    line-height: 18px;
  }

  strong {
    overflow: hidden;
    color: #334155;
    font-size: 13px;
    font-weight: 600;
    line-height: 20px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.export-dimension__chooser {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 210px;
  gap: 12px;
  align-items: stretch;
}

.export-dimension__matrix {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
  max-height: 188px;
  overflow-y: auto;
  scrollbar-color: #cfd8e3 transparent;
  scrollbar-width: thin;

  &::-webkit-scrollbar {
    width: 6px;
  }

  &::-webkit-scrollbar-track {
    background: transparent;
  }

  &::-webkit-scrollbar-thumb {
    background: #cfd8e3;
    border-radius: 999px;
  }

  &::-webkit-scrollbar-thumb:hover {
    background: #aebacd;
  }
}

.export-dimension-chip {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
  height: 42px;
  padding: 0 10px;
  color: inherit;
  font: inherit;
  text-align: left;
  cursor: pointer;
  background: #fff;
  border: 1px solid #edf0f5;
  border-radius: 8px;
  transition: border-color 0.16s ease, box-shadow 0.16s ease, background 0.16s ease;

  &:hover {
    background: #fbfdff;
    border-color: #c9ddf7;
  }

  &:focus-visible {
    outline: 2px solid rgba(24, 144, 255, 0.22);
    outline-offset: 2px;
  }

  &:disabled {
    cursor: not-allowed;
    opacity: 0.72;
  }
}

.export-dimension-chip--active {
  background: #f7fbff;
  border-color: #8dc6ff;
  box-shadow: 0 3px 10px rgba(24, 144, 255, 0.08);
}

.export-dimension-chip__dot {
  flex: 0 0 auto;
  width: 8px;
  height: 8px;
  background: #cbd5e1;
  border-radius: 50%;
}

.export-dimension-chip--active .export-dimension-chip__dot {
  background: var(--pro-ant-color-primary);
  box-shadow: 0 0 0 3px rgba(24, 144, 255, 0.13);
}

.export-dimension-chip__text {
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

.export-dimension-chip--active .export-dimension-chip__text {
  color: #1f2937;
  font-weight: 600;
}

.export-dimension-chip__tag {
  flex: 0 0 auto;
  padding: 0 5px;
  color: var(--pro-ant-color-primary);
  font-size: 11px;
  font-weight: 500;
  line-height: 18px;
  background: #eef6ff;
  border-radius: 4px;
}

.export-dimension__detail {
  display: flex;
  flex-direction: column;
  min-width: 0;
  padding: 12px;
  background: #f8fafc;
  border: 1px solid #edf1f6;
  border-radius: 8px;
}

.export-dimension__detail-label {
  color: #98a2b3;
  font-size: 12px;
  line-height: 18px;
}

.export-dimension__detail strong {
  margin-top: 4px;
  overflow: hidden;
  color: #1f2937;
  font-size: 14px;
  font-weight: 600;
  line-height: 22px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.export-dimension__detail p {
  flex: 1 1 auto;
  display: -webkit-box;
  margin: 8px 0 12px;
  overflow: hidden;
  color: #687386;
  font-size: 12px;
  line-height: 19px;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 3;
}

.export-dimension__detail-pages {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding-top: 10px;
  color: #8a94a6;
  font-size: 12px;
  line-height: 18px;
  border-top: 1px solid #e8edf4;

  em {
    color: var(--pro-ant-color-primary);
    font-style: normal;
    font-weight: 600;
  }
}

.export-dimension__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 20px;
  margin: 16px -20px 0;
  background: #fbfcfe;
  border-top: 1px solid #eef1f5;
}

.export-dimension__selection {
  display: flex;
  align-items: center;
  min-width: 0;
  color: #7a8494;
  font-size: 13px;
  line-height: 22px;

  span {
    flex: 0 0 auto;
  }

  strong {
    min-width: 0;
    margin-left: 8px;
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

@media (max-width: 760px) {
  .report-head,
  .export-dimension__summary,
  .export-dimension__footer {
    align-items: flex-start;
  }

  .report-head {
    flex-direction: column;
  }

  .report-info {
    align-items: flex-start;
    flex-direction: column;
    gap: 2px;
    width: 100%;
  }

  .report-module-area {
    align-items: stretch;
    flex-direction: column;
  }

  .report-module-grid {
    width: 100%;
  }

  .report-module-summary {
    padding-top: 8px;
    padding-left: 0;
    border-top: 1px solid #e6edf6;
    border-left: 0;
  }

  .export-dimension__current {
    display: none;
  }

  .export-dimension__chooser {
    grid-template-columns: 1fr;
  }

  .export-dimension__matrix {
    max-height: 55vh;
  }

  .export-dimension__footer {
    flex-direction: column;
  }

  .export-dimension__selection,
  .export-dimension__actions {
    width: 100%;
  }

  .export-dimension__actions {
    justify-content: flex-end;
  }
}
</style>

<style lang="less">
.pep3-report-modal {
  .ant-modal {
    max-width: calc(100vw - 48px);
  }

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
    max-height: calc(100vh - 150px);
    padding: 0;
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

    &::-webkit-scrollbar-thumb:hover {
      background: #aebccc;
      background-clip: padding-box;
      border: 2px solid transparent;
    }
  }
}

.pep3-export-dimension-modal {
  .ant-modal-content {
    padding: 0;
    overflow: hidden;
    border-radius: 12px;
    box-shadow: 0 10px 32px rgba(15, 23, 42, 0.14);
  }

  .ant-modal-header {
    padding: 18px 24px 14px;
    margin: 0;
    border-bottom: 1px solid #eef1f5;
  }

  .ant-modal-title {
    margin: 0;
  }

  .ant-modal-close {
    top: 16px;
    inset-inline-end: 16px;
    color: #8a94a6;
  }

  .ant-modal-body {
    padding: 0;
  }
}

@media (max-width: 760px) {
  .pep3-report-modal,
  .pep3-export-dimension-modal {
    .ant-modal {
      width: calc(100vw - 32px) !important;
      max-width: calc(100vw - 32px);
    }
  }
}
</style>
