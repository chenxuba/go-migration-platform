<script setup>
import dayjs from 'dayjs'
import messageService from '@/utils/messageService'
import { useTableColumns } from '@/composables/useTableColumns'
import {
  deletePEP3AssessmentRecordApi,
  downloadPEP3AssessmentBookletPdfApi,
  getPEP3AssessmentReportApi,
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
const selectedExportDimension = ref('all')

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
    width: 170,
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

function exportDimensionTitle(value) {
  return exportDimensionOptions.find(item => item.value === value)?.title || '全维度导出'
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
  previewLoading.value = true
  try {
    const res = await getPEP3AssessmentReportApi(row.id)
    currentReport.value = unwrap(res)
    reportModalOpen.value = true
  }
  catch (error) {
    messageService.error(getErrorMessage(error, '获取评估报告失败'))
  }
  finally {
    previewLoading.value = false
  }
}

function openExportModal(row) {
  if (!row || exportingId.value)
    return
  exportTargetRecord.value = row
  selectedExportDimension.value = 'all'
  exportModalOpen.value = true
}

function closeExportModal() {
  if (exportingId.value)
    return
  exportModalOpen.value = false
}

async function exportReport(row = exportTargetRecord.value) {
  if (!row)
    return
  exportingId.value = row.id
  try {
    const response = await downloadPEP3AssessmentBookletPdfApi(row.id, selectedExportDimension.value)
    const url = URL.createObjectURL(new Blob([response.data], { type: 'application/pdf' }))
    const link = document.createElement('a')
    link.href = url
    const fallbackName = `${row.studentName || '学员'}-${row.assessmentName || '评估记录'}-${exportDimensionTitle(selectedExportDimension.value)}-${formatDate(row.assessmentDate)}.pdf`
    link.download = getDownloadFilename(response, fallbackName)
    link.click()
    window.setTimeout(() => URL.revokeObjectURL(url), 60_000)
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
                  :age="formatAge(record)"
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
                </a-space>
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>

    <a-modal v-model:open="reportModalOpen" width="840px" title="评估报告" :footer="null">
      <a-spin :spinning="previewLoading">
        <div v-if="currentReport" class="report-preview">
          <div class="report-head">
            <div>
              <div class="report-title">
                {{ currentReport.title || currentReport.record?.assessmentName || '评估报告' }}
              </div>
              <div class="report-subtitle">
                {{ currentReport.record?.studentName || '-' }} / {{ formatDate(currentReport.record?.assessmentDate) }}
              </div>
            </div>
            <a-button type="primary" @click="openExportModal(currentReport.record)">
              导出
            </a-button>
          </div>
          <div class="report-section-list">
            <div v-for="section in currentReport.sections || []" :key="section.sectionCode" class="report-section">
              <div class="report-section-title">
                {{ section.title }}
              </div>
              <div v-if="section.fields?.length" class="report-fields">
                <div v-for="field in section.fields" :key="field.key" class="report-field">
                  <span>{{ field.label }}</span>
                  <strong>{{ field.value || '-' }}</strong>
                </div>
              </div>
              <div v-if="section.textItems?.length" class="report-text">
                <p v-for="(text, index) in section.textItems" :key="index">
                  {{ text }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </a-spin>
    </a-modal>

    <a-modal
      v-model:open="exportModalOpen"
      width="680px"
      :centered="true"
      title="选择导出维度"
      wrap-class-name="pep3-export-dimension-modal"
      :mask-closable="!exportingId"
      :confirm-loading="!!exportingId"
      ok-text="开始导出"
      cancel-text="取消"
      @ok="exportReport()"
      @cancel="closeExportModal"
    >
      <div class="export-dimension">
        <div class="export-dimension__summary">
          <div>
            <div class="export-dimension__name">
              {{ exportTargetRecord?.studentName || '学员' }}
            </div>
            <div class="export-dimension__meta">
              {{ exportTargetRecord?.assessmentName || '评估记录' }} · {{ formatDate(exportTargetRecord?.assessmentDate) }}
            </div>
          </div>
          <div class="export-dimension__type">
            PDF
          </div>
        </div>
        <div class="export-dimension__grid">
          <button
            v-for="option in exportDimensionOptions"
            :key="option.value"
            type="button"
            class="export-dimension-card"
            :class="{ 'export-dimension-card--active': selectedExportDimension === option.value }"
            :aria-pressed="selectedExportDimension === option.value"
            :disabled="!!exportingId"
            @click="selectedExportDimension = option.value"
          >
            <span class="export-dimension-card__badge">{{ option.badge }}</span>
            <span class="export-dimension-card__content">
              <span class="export-dimension-card__title">
                {{ option.title }}
                <span v-if="option.recommended" class="export-dimension-card__tag">推荐</span>
              </span>
              <span class="export-dimension-card__desc">{{ option.desc }}</span>
              <span class="export-dimension-card__pages">{{ option.pages }}</span>
            </span>
          </button>
        </div>
      </div>
    </a-modal>
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

.report-preview {
  max-height: 68vh;
  overflow-y: auto;
  padding-right: 4px;
}

.report-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 16px;
  border-bottom: 1px solid #f0f0f0;
}

.report-title {
  color: #222;
  font-size: 16px;
  font-weight: 600;
  line-height: 24px;
}

.report-subtitle {
  margin-top: 4px;
  color: #888;
  font-size: 13px;
}

.report-section {
  padding: 14px 0;
  border-bottom: 1px solid #f5f5f5;
}

.report-section-title {
  color: #222;
  font-weight: 600;
  line-height: 22px;
}

.report-fields {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px 18px;
  margin-top: 10px;
}

.report-field {
  display: flex;
  justify-content: space-between;
  min-width: 0;
  color: #666;

  span,
  strong {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  strong {
    max-width: 58%;
    color: #222;
    font-weight: 500;
  }
}

.report-text {
  margin-top: 10px;
  color: #555;
  line-height: 22px;

  p {
    margin-bottom: 6px;
  }
}

.export-dimension {
  padding-top: 2px;
}

.export-dimension__summary {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
  margin-bottom: 16px;
  background: #f7f9fc;
  border: 1px solid #eef1f6;
  border-radius: 8px;
}

.export-dimension__name {
  color: #1f2937;
  font-size: 15px;
  font-weight: 600;
  line-height: 22px;
}

.export-dimension__meta {
  margin-top: 3px;
  color: #7a8494;
  font-size: 12px;
  line-height: 18px;
}

.export-dimension__type {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 46px;
  height: 28px;
  color: var(--pro-ant-color-primary);
  font-size: 13px;
  font-weight: 700;
  background: #eef5ff;
  border: 1px solid #d7e8ff;
  border-radius: 6px;
}

.export-dimension__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.export-dimension-card {
  display: flex;
  gap: 12px;
  width: 100%;
  min-height: 104px;
  padding: 14px;
  color: inherit;
  font: inherit;
  text-align: left;
  cursor: pointer;
  background: #fff;
  border: 1px solid #edf0f5;
  border-radius: 8px;
  transition: border-color 0.16s ease, box-shadow 0.16s ease, background 0.16s ease;

  &:hover {
    border-color: #bfd9ff;
    box-shadow: 0 6px 18px rgba(24, 144, 255, 0.08);
  }

  &:focus-visible {
    outline: 2px solid rgba(24, 144, 255, 0.28);
    outline-offset: 2px;
  }

  &:disabled {
    cursor: not-allowed;
    opacity: 0.72;
  }
}

.export-dimension-card--active {
  background: #f7fbff;
  border-color: var(--pro-ant-color-primary);
  box-shadow: 0 8px 22px rgba(24, 144, 255, 0.12);
}

.export-dimension-card__badge {
  flex: 0 0 auto;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 38px;
  height: 38px;
  color: #53708f;
  font-size: 12px;
  font-weight: 700;
  background: #f2f5f9;
  border-radius: 8px;
}

.export-dimension-card--active .export-dimension-card__badge {
  color: #fff;
  background: var(--pro-ant-color-primary);
}

.export-dimension-card__content {
  min-width: 0;
}

.export-dimension-card__title {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
  color: #202733;
  font-size: 14px;
  font-weight: 600;
  line-height: 20px;
}

.export-dimension-card__tag {
  flex: 0 0 auto;
  padding: 0 5px;
  color: var(--pro-ant-color-primary);
  font-size: 11px;
  font-weight: 500;
  line-height: 18px;
  background: #eef5ff;
  border-radius: 4px;
}

.export-dimension-card__desc {
  display: block;
  margin-top: 6px;
  color: #687386;
  font-size: 12px;
  line-height: 18px;
}

.export-dimension-card__pages {
  display: block;
  margin-top: 8px;
  color: #98a2b3;
  font-size: 12px;
  line-height: 18px;
}
</style>
