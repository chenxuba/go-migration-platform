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
    width: 120,
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

async function exportReport(row) {
  exportingId.value = row.id
  try {
    const response = await downloadPEP3AssessmentBookletPdfApi(row.id)
    const url = URL.createObjectURL(new Blob([response.data], { type: 'application/pdf' }))
    const link = document.createElement('a')
    link.href = url
    link.download = `${row.studentName || '学员'}-${row.assessmentName || '评估记录'}-${formatDate(row.assessmentDate)}.pdf`
    link.click()
    window.setTimeout(() => URL.revokeObjectURL(url), 60_000)
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
                <div class="student-cell">
                  <img class="student-avatar" :src="record.studentAvatar" alt="">
                  <div class="student-info">
                    <div class="student-name">
                      {{ record.studentName || '-' }}
                    </div>
                    <div class="student-meta">
                      {{ record.studentGender || '-' }}
                    </div>
                  </div>
                </div>
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
                  <a :class="{ disabled: exportingId === record.id }" @click="exportReport(record)">导出</a>
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
            <a-button type="primary" @click="exportReport(currentReport.record)">
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

.student-cell {
  display: flex;
  align-items: center;
  min-width: 0;
}

.student-avatar {
  width: 40px;
  height: 40px;
  margin-right: 8px;
  border-radius: 50%;
  object-fit: cover;
  flex: 0 0 40px;
}

.student-info {
  min-width: 0;
}

.student-name,
.single-line {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.student-name {
  color: #222;
}

.student-meta {
  margin-top: 2px;
  color: #888;
  font-size: 12px;
  line-height: 18px;
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
</style>
