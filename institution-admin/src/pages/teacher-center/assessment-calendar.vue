<script setup lang="ts">
import {
  BookOutlined,
  CheckCircleOutlined,
  FileTextOutlined,
  FormOutlined,
  ReloadOutlined,
} from '@ant-design/icons-vue'
import { computed, h, onMounted, reactive, ref } from 'vue'
import { Modal } from 'ant-design-vue'
import dayjs, { type Dayjs } from 'dayjs'
import {
  deletePEP3AssessmentDraftApi,
  downloadPEP3AssessmentBookletPdfApi,
  getPEP3AssessmentBookletApi,
  getPEP3AssessmentDraftDetailApi,
  getPEP3AssessmentFormTemplateApi,
  getPEP3AssessmentReportApi,
  pagePEP3AssessmentDraftsApi,
  pagePEP3AssessmentRecordsApi,
  savePEP3AssessmentDraftApi,
  submitPEP3AssessmentDraftApi,
  type PEP3AssessmentDraftDetail,
  type PEP3AssessmentDraftSummary,
  type PEP3AssessmentFormTemplate,
  type PEP3AssessmentRecordSummary,
  type PEP3Booklet,
  type PEP3DraftSaveRequest,
  type PEP3RawScoreField,
  type PEP3Report,
  type PEP3ScaleCode,
  type PEP3TemplateSection,
} from '@/api/edu-center/pep3-assessment'
import messageService from '@/utils/messageService'

type WorkbenchTab = 'drafts' | 'records'

interface DraftEditorState {
  id?: number
  studentId?: number
  studentName: string
  examinerName: string
  birthDate: Dayjs | null
  assessmentDate: Dayjs | null
  remark: string
  allowMissingItems: boolean
  itemScores: Record<number, number | undefined>
  rawScores: Record<string, number | undefined>
}

const activeTab = ref<WorkbenchTab>('drafts')
const templateLoading = ref(false)
const draftsLoading = ref(false)
const recordsLoading = ref(false)
const drawerOpen = ref(false)
const saving = ref(false)
const submitting = ref(false)
const previewLoading = ref(false)

const template = ref<PEP3AssessmentFormTemplate>()
const draftRows = ref<PEP3AssessmentDraftSummary[]>([])
const recordRows = ref<PEP3AssessmentRecordSummary[]>([])
const currentProgress = ref<PEP3AssessmentDraftDetail['progress']>()
const activeGroupKeys = ref<string[]>(['booklet_page_2'])
const reportModalOpen = ref(false)
const bookletModalOpen = ref(false)
const currentReport = ref<PEP3Report>()
const currentBooklet = ref<PEP3Booklet>()

const draftPagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showTotal: (total: number) => `共 ${total} 条`,
})

const recordPagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showTotal: (total: number) => `共 ${total} 条`,
})

const draftFilter = reactive({
  searchKey: '',
  status: undefined as string | undefined,
})

const recordFilter = reactive({
  searchKey: '',
})

const editor = reactive<DraftEditorState>({
  studentName: '',
  examinerName: '',
  birthDate: null,
  assessmentDate: dayjs(),
  remark: '',
  allowMissingItems: true,
  itemScores: {},
  rawScores: {},
})

const developmentRawScoreFields = computed(() => {
  return (template.value?.rawScoreFields || []).filter(item => item.category !== 'caregiver_report')
})

const caregiverRawScoreFields = computed(() => {
  return (template.value?.rawScoreFields || []).filter(item => item.category === 'caregiver_report')
})

const caregiverScaleCodes = computed(() => {
  const codes = caregiverRawScoreFields.value.map(item => String(item.scaleCode))
  return new Set(codes.length ? codes : ['PB', 'PSC', 'AB'])
})

const caregiverRawScoreTotal = computed(() => caregiverRawScoreFields.value.length || 3)

const itemGroups = computed(() => template.value?.itemGroups || [])

const answeredItemCount = computed(() => {
  return Object.values(editor.itemScores).filter(value => value === 0 || value === 1 || value === 2).length
})

const rawScoreCount = computed(() => {
  return Object.entries(editor.rawScores).filter(([scaleCode, value]) => {
    return caregiverScaleCodes.value.has(scaleCode) && typeof value === 'number' && Number.isFinite(value)
  }).length
})

const autoRawScoreRows = computed(() => {
  const rows = developmentRawScoreFields.value.map(field => ({
    scaleCode: String(field.scaleCode),
    scaleName: field.scaleName,
    itemCount: 0,
    answeredItemCount: 0,
    rawScore: 0,
    maxRawScore: field.maxScore,
  }))
  const rowMap = new Map(rows.map(row => [row.scaleCode, row]))
  itemGroups.value.forEach((group) => {
    group.items.forEach((item) => {
      const row = rowMap.get(String(item.domainCode))
      if (!row)
        return
      row.itemCount += 1
      const score = editor.itemScores[item.itemNo]
      if (score === 0 || score === 1 || score === 2) {
        row.answeredItemCount += 1
        row.rawScore += Number(score)
      }
    })
  })
  return rows
})

const editorPercent = computed(() => {
  if (currentProgress.value)
    return currentProgress.value.completionPercent || 0
  const total = (template.value?.itemCount || 172) + 3
  return Math.round((answeredItemCount.value + rawScoreCount.value) * 1000 / total) / 10
})

const summaryCards = computed(() => [
  {
    label: '草稿',
    value: draftPagination.total,
    desc: '待继续测评',
    icon: FormOutlined,
  },
  {
    label: '正式记录',
    value: recordPagination.total,
    desc: '已完成评分',
    icon: FileTextOutlined,
  },
  {
    label: '题目',
    value: template.value?.itemCount || 172,
    desc: '录入模板',
    icon: BookOutlined,
  },
  {
    label: '本次进度',
    value: `${editorPercent.value}%`,
    desc: '当前抽屉',
    icon: CheckCircleOutlined,
  },
])

const draftColumns: any[] = [
  { title: '儿童', dataIndex: 'studentName', key: 'studentName', width: 170, fixed: 'left' },
  { title: '评估日期', dataIndex: 'assessmentDate', key: 'assessmentDate', width: 130 },
  { title: '进度', dataIndex: 'progress', key: 'progress', width: 220 },
  { title: '状态', dataIndex: 'status', key: 'status', width: 130 },
  { title: '测试员', dataIndex: 'examinerName', key: 'examinerName', width: 140 },
  { title: '更新时间', dataIndex: 'updatedTime', key: 'updatedTime', width: 170 },
  { title: '备注', dataIndex: 'remark', key: 'remark', width: 200 },
  { title: '操作', dataIndex: 'action', key: 'action', width: 220, fixed: 'right' },
]

const recordColumns: any[] = [
  { title: '儿童', dataIndex: 'studentName', key: 'studentName', width: 170, fixed: 'left' },
  { title: '评估日期', dataIndex: 'assessmentDate', key: 'assessmentDate', width: 130 },
  { title: '实足年龄', dataIndex: 'age', key: 'age', width: 130 },
  { title: '常模月龄', dataIndex: 'normAgeMonths', key: 'normAgeMonths', width: 110 },
  { title: '测试员', dataIndex: 'examinerName', key: 'examinerName', width: 140 },
  { title: '数据状态', dataIndex: 'dataStatus', key: 'dataStatus', width: 260 },
  { title: '创建时间', dataIndex: 'createdTime', key: 'createdTime', width: 170 },
  { title: '操作', dataIndex: 'action', key: 'action', width: 250, fixed: 'right' },
]

function unwrap<T>(res: { data?: T, result?: T }): T {
  return (res.data ?? res.result) as T
}

function formatDate(value?: string) {
  if (!value)
    return '-'
  return dayjs(value).isValid() ? dayjs(value).format('YYYY-MM-DD') : value
}

function formatDateTime(value?: string) {
  if (!value)
    return '-'
  return dayjs(value).isValid() ? dayjs(value).format('YYYY-MM-DD HH:mm') : value
}

function statusMeta(status?: string) {
  const map: Record<string, { color: string, text: string }> = {
    draft: { color: 'default', text: '草稿' },
    ready_to_score: { color: 'processing', text: '可评分' },
    complete: { color: 'success', text: '已完整' },
    submitted: { color: 'purple', text: '已提交' },
  }
  return map[status || ''] || { color: 'default', text: status || '-' }
}

function getErrorMessage(error: any, fallback: string) {
  return error?.response?.data?.message || error?.message || fallback
}

function resetEditor() {
  editor.id = undefined
  editor.studentId = undefined
  editor.studentName = ''
  editor.examinerName = ''
  editor.birthDate = null
  editor.assessmentDate = dayjs()
  editor.remark = ''
  editor.allowMissingItems = true
  editor.itemScores = {}
  editor.rawScores = {}
  currentProgress.value = undefined
  activeGroupKeys.value = ['booklet_page_2']
}

function applyDraftInput(input?: PEP3DraftSaveRequest) {
  if (!input)
    return
  editor.studentId = input.studentId
  editor.studentName = input.studentName || ''
  editor.examinerName = input.examinerName || ''
  editor.birthDate = input.birthDate ? dayjs(input.birthDate) : null
  editor.assessmentDate = input.assessmentDate ? dayjs(input.assessmentDate) : null
  editor.remark = input.remark || ''
  editor.allowMissingItems = input.allowMissingItems ?? true
  editor.itemScores = {}
  editor.rawScores = {}
  Object.entries(input.itemScores || {}).forEach(([itemNo, score]) => {
    editor.itemScores[Number(itemNo)] = Number(score)
  })
  ;(input.itemScoreList || []).forEach((item) => {
    editor.itemScores[item.itemNo] = item.score
  })
  Object.entries(input.rawScores || {}).forEach(([scaleCode, score]) => {
    editor.rawScores[scaleCode] = Number(score)
  })
  ;(input.rawScoreList || []).forEach((item) => {
    editor.rawScores[item.scaleCode] = item.rawScore
  })
}

async function fetchTemplate() {
  templateLoading.value = true
  try {
    const res = await getPEP3AssessmentFormTemplateApi()
    template.value = unwrap<PEP3AssessmentFormTemplate>(res)
  }
  catch (error: any) {
    messageService.error(getErrorMessage(error, '获取PEP-3题库失败'))
  }
  finally {
    templateLoading.value = false
  }
}

async function fetchDrafts() {
  draftsLoading.value = true
  try {
    const res = await pagePEP3AssessmentDraftsApi({
      pageRequestModel: {
        pageIndex: draftPagination.current,
        pageSize: draftPagination.pageSize,
      },
      queryModel: {
        searchKey: draftFilter.searchKey || undefined,
        status: draftFilter.status,
      },
    })
    const data = unwrap<any>(res)
    draftRows.value = data?.items || res.result || []
    draftPagination.total = data?.total ?? res.total ?? 0
  }
  catch (error: any) {
    messageService.error(getErrorMessage(error, '获取草稿列表失败'))
  }
  finally {
    draftsLoading.value = false
  }
}

async function fetchRecords() {
  recordsLoading.value = true
  try {
    const res = await pagePEP3AssessmentRecordsApi({
      pageRequestModel: {
        pageIndex: recordPagination.current,
        pageSize: recordPagination.pageSize,
      },
      queryModel: {
        searchKey: recordFilter.searchKey || undefined,
      },
    })
    const data = unwrap<any>(res)
    recordRows.value = data?.items || res.result || []
    recordPagination.total = data?.total ?? res.total ?? 0
  }
  catch (error: any) {
    messageService.error(getErrorMessage(error, '获取正式记录失败'))
  }
  finally {
    recordsLoading.value = false
  }
}

function handleDraftTableChange(pagination: any) {
  draftPagination.current = pagination.current
  draftPagination.pageSize = pagination.pageSize
  fetchDrafts()
}

function handleRecordTableChange(pagination: any) {
  recordPagination.current = pagination.current
  recordPagination.pageSize = pagination.pageSize
  fetchRecords()
}

function buildPayload(): PEP3DraftSaveRequest {
  const itemScoreList = Object.entries(editor.itemScores)
    .filter(([, score]) => score === 0 || score === 1 || score === 2)
    .map(([itemNo, score]) => ({ itemNo: Number(itemNo), score: Number(score) }))
    .sort((a, b) => a.itemNo - b.itemNo)
  const rawScoreList = Object.entries(editor.rawScores)
    .filter(([scaleCode, score]) => caregiverScaleCodes.value.has(scaleCode) && typeof score === 'number' && Number.isFinite(score))
    .map(([scaleCode, rawScore]) => ({ scaleCode: scaleCode as PEP3ScaleCode, rawScore: Number(rawScore) }))
    .sort((a, b) => a.scaleCode.localeCompare(b.scaleCode))

  return {
    id: editor.id,
    studentId: editor.studentId,
    studentName: editor.studentName.trim(),
    examinerName: editor.examinerName.trim(),
    birthDate: editor.birthDate ? editor.birthDate.format('YYYY-MM-DD') : undefined,
    assessmentDate: editor.assessmentDate ? editor.assessmentDate.format('YYYY-MM-DD') : undefined,
    remark: editor.remark.trim(),
    allowMissingItems: editor.allowMissingItems,
    itemScoreList,
    rawScoreList,
  }
}

async function saveDraft(silent = false) {
  saving.value = true
  try {
    const res = await savePEP3AssessmentDraftApi(buildPayload())
    const detail = unwrap<PEP3AssessmentDraftDetail>(res)
    editor.id = detail.id
    currentProgress.value = detail.progress
    if (!silent)
      messageService.success('草稿已保存')
    await fetchDrafts()
    return detail
  }
  catch (error: any) {
    messageService.error(getErrorMessage(error, '保存草稿失败'))
    return undefined
  }
  finally {
    saving.value = false
  }
}

async function submitDraft() {
  submitting.value = true
  try {
    const detail = await saveDraft(true)
    if (!detail?.id)
      return
    if (!detail.progress?.canScore) {
      currentProgress.value = detail.progress
      messageService.warning(cannotSubmitMessage(detail.progress))
      return
    }
    const res = await submitPEP3AssessmentDraftApi(detail.id)
    const submitResult = unwrap<any>(res)
    messageService.success('已生成正式测评记录')
    drawerOpen.value = false
    activeTab.value = 'records'
    await Promise.all([fetchDrafts(), fetchRecords()])
    if (submitResult?.recordId)
      await openReport({ id: submitResult.recordId } as PEP3AssessmentRecordSummary)
  }
  catch (error: any) {
    messageService.error(getErrorMessage(error, '提交正式记录失败'))
  }
  finally {
    submitting.value = false
  }
}

async function openNewAssessment() {
  if (!template.value)
    await fetchTemplate()
  resetEditor()
  drawerOpen.value = true
}

async function continueDraft(row: PEP3AssessmentDraftSummary) {
  if (!template.value)
    await fetchTemplate()
  try {
    const res = await getPEP3AssessmentDraftDetailApi(row.id)
    const detail = unwrap<PEP3AssessmentDraftDetail>(res)
    resetEditor()
    editor.id = detail.id
    currentProgress.value = detail.progress
    applyDraftInput(detail.input)
    drawerOpen.value = true
  }
  catch (error: any) {
    messageService.error(getErrorMessage(error, '获取草稿详情失败'))
  }
}

function confirmDeleteDraft(row: PEP3AssessmentDraftSummary) {
  Modal.confirm({
    title: '删除草稿',
    content: `确认删除 ${row.studentName || '未命名儿童'} 的PEP-3草稿？`,
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      try {
        await deletePEP3AssessmentDraftApi(row.id)
        messageService.success('草稿已删除')
        await fetchDrafts()
      }
      catch (error: any) {
        messageService.error(getErrorMessage(error, '删除草稿失败'))
      }
    },
  })
}

async function openReport(row: PEP3AssessmentRecordSummary) {
  previewLoading.value = true
  try {
    const res = await getPEP3AssessmentReportApi(row.id)
    currentReport.value = unwrap<PEP3Report>(res)
    reportModalOpen.value = true
  }
  catch (error: any) {
    messageService.error(getErrorMessage(error, '获取报告失败'))
  }
  finally {
    previewLoading.value = false
  }
}

async function openBooklet(row: PEP3AssessmentRecordSummary) {
  previewLoading.value = true
  try {
    const res = await getPEP3AssessmentBookletApi(row.id)
    currentBooklet.value = unwrap<PEP3Booklet>(res)
    bookletModalOpen.value = true
  }
  catch (error: any) {
    messageService.error(getErrorMessage(error, '获取记录册失败'))
  }
  finally {
    previewLoading.value = false
  }
}

async function openBookletPdf(row: PEP3AssessmentRecordSummary) {
  previewLoading.value = true
  const previewWindow = window.open('', '_blank')
  try {
    const response = await downloadPEP3AssessmentBookletPdfApi(row.id)
    const url = URL.createObjectURL(new Blob([response.data], { type: 'application/pdf' }))
    if (previewWindow)
      previewWindow.location.href = url
    else
      window.open(url, '_blank', 'noopener,noreferrer')
    window.setTimeout(() => URL.revokeObjectURL(url), 60_000)
  }
  catch (error: any) {
    previewWindow?.close()
    messageService.error(getErrorMessage(error, '生成记录册PDF失败'))
  }
  finally {
    previewLoading.value = false
  }
}

function continueDraftRow(row: Record<string, any>) {
  return continueDraft(row as PEP3AssessmentDraftSummary)
}

function submitDraftRow(row: Record<string, any>) {
  return continueDraft(row as PEP3AssessmentDraftSummary).then(() => submitDraft())
}

function confirmDeleteDraftRow(row: Record<string, any>) {
  confirmDeleteDraft(row as PEP3AssessmentDraftSummary)
}

function openReportRow(row: Record<string, any>) {
  return openReport(row as PEP3AssessmentRecordSummary)
}

function openBookletRow(row: Record<string, any>) {
  return openBooklet(row as PEP3AssessmentRecordSummary)
}

function openBookletPdfRow(row: Record<string, any>) {
  return openBookletPdf(row as PEP3AssessmentRecordSummary)
}

function antTableColumns(section: PEP3TemplateSection) {
  const columns: any[] = []
  for (const column of section.table?.columns || []) {
    const antColumn = {
      title: column.label,
      dataIndex: column.key,
      key: column.key,
      width: column.width,
      align: column.align as any,
      customRender: ({ text }: { text: unknown }) => formatPreviewCell(text, column.key),
    }
    if (!column.group) {
      columns.push(antColumn)
      continue
    }
    const last = columns[columns.length - 1]
    if (last?.key === `group:${column.group}` && Array.isArray(last.children)) {
      last.children.push(antColumn)
    }
    else {
      columns.push({
        title: column.group,
        key: `group:${column.group}`,
        children: [antColumn],
      })
    }
  }
  return columns
}

function formatPreviewCell(value: unknown, columnKey: string) {
  if (Array.isArray(value))
    return value.join('、')
  if (value === undefined || value === null || value === '') {
    return ['developmentAge', 'percentileRank', 'scaledScore', 'standardScoreSum', 'level'].includes(columnKey)
      ? '待校对'
      : '-'
  }
  return value
}

function caregiverRawScoreDisplay(row: { progress?: { caregiverRawScoreCount?: number }, rawScoreCount?: number }) {
  const count = row.progress?.caregiverRawScoreCount ?? Math.min(row.rawScoreCount || 0, caregiverRawScoreTotal.value)
  return `${count}/${caregiverRawScoreTotal.value}`
}

function cannotSubmitMessage(progress?: PEP3AssessmentDraftDetail['progress']) {
  const missing = progress?.missingRequiredFields || []
  if (missing.includes('birthDate') || missing.includes('assessmentDate'))
    return '请先填写出生日期和评估日期'
  if (!progress?.answeredItemCount)
    return '请先录入逐题得分，照顾者报告原始分不能单独生成正式记录'
  return '当前草稿还不能生成正式记录，请继续补充分数'
}

function rawScoreLabel(field: PEP3RawScoreField) {
  return field.maxScore ? `${field.scaleName}（${field.scaleCode}，0-${field.maxScore}）` : `${field.scaleName}（${field.scaleCode}）`
}

onMounted(async () => {
  await fetchTemplate()
  await Promise.all([fetchDrafts(), fetchRecords()])
})
</script>

<template>
  <div class="pep3-page">
    <div class="pep3-header">
      <div>
        <div class="pep3-eyebrow">PEP-3儿童心理教育评核</div>
        <h1>评估量表</h1>
      </div>
      <a-space>
        <a-button :icon="h(ReloadOutlined)" @click="() => { fetchDrafts(); fetchRecords() }">
          刷新
        </a-button>
        <a-button type="primary" :icon="h(FormOutlined)" @click="openNewAssessment">
          新建测评
        </a-button>
      </a-space>
    </div>

    <div class="summary-grid">
      <div v-for="item in summaryCards" :key="item.label" class="summary-item">
        <component :is="item.icon" class="summary-item__icon" />
        <div>
          <div class="summary-item__value">{{ item.value }}</div>
          <div class="summary-item__label">{{ item.label }} · {{ item.desc }}</div>
        </div>
      </div>
    </div>

    <div class="content-panel">
      <a-tabs v-model:active-key="activeTab">
        <a-tab-pane key="drafts" tab="测评草稿">
          <div class="table-toolbar">
            <a-space>
              <a-input-search
                v-model:value="draftFilter.searchKey"
                allow-clear
                placeholder="搜索儿童/测试员"
                style="width: 240px"
                @search="() => { draftPagination.current = 1; fetchDrafts() }"
              />
              <a-select
                v-model:value="draftFilter.status"
                allow-clear
                placeholder="草稿状态"
                style="width: 150px"
                @change="() => { draftPagination.current = 1; fetchDrafts() }"
              >
                <a-select-option value="draft">草稿</a-select-option>
                <a-select-option value="ready_to_score">可评分</a-select-option>
                <a-select-option value="complete">已完整</a-select-option>
                <a-select-option value="submitted">已提交</a-select-option>
              </a-select>
            </a-space>
            <a-button type="primary" @click="openNewAssessment">新建测评</a-button>
          </div>

          <a-table
            row-key="id"
            size="small"
            :loading="draftsLoading"
            :columns="draftColumns"
            :data-source="draftRows"
            :pagination="draftPagination"
            :scroll="{ x: 1220 }"
            @change="handleDraftTableChange"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'studentName'">
                <div class="primary-cell">
                  <span>{{ record.studentName || '未填写姓名' }}</span>
                  <small>ID {{ record.studentId || '-' }}</small>
                </div>
              </template>
              <template v-else-if="column.key === 'assessmentDate'">
                {{ formatDate(record.assessmentDate) }}
              </template>
              <template v-else-if="column.key === 'progress'">
                <div class="progress-cell">
                  <a-progress :percent="record.progress?.completionPercent || 0" size="small" />
                  <span>{{ record.answeredItemCount }}/{{ record.progress?.itemCount || 172 }}题，照顾者{{ caregiverRawScoreDisplay(record) }}</span>
                </div>
              </template>
              <template v-else-if="column.key === 'status'">
                <a-tag :color="statusMeta(record.status).color">{{ statusMeta(record.status).text }}</a-tag>
              </template>
              <template v-else-if="column.key === 'updatedTime'">
                {{ formatDateTime(record.updatedTime) }}
              </template>
              <template v-else-if="column.key === 'remark'">
                <a-tooltip :title="record.remark">
                  <span class="ellipsis">{{ record.remark || '-' }}</span>
                </a-tooltip>
              </template>
              <template v-else-if="column.key === 'action'">
                <a-space>
                  <a @click="continueDraftRow(record)">继续</a>
                  <a @click="submitDraftRow(record)">提交</a>
                  <a class="danger-link" @click="confirmDeleteDraftRow(record)">删除</a>
                </a-space>
              </template>
            </template>
          </a-table>
        </a-tab-pane>

        <a-tab-pane key="records" tab="正式记录">
          <div class="table-toolbar">
            <a-input-search
              v-model:value="recordFilter.searchKey"
              allow-clear
              placeholder="搜索儿童/测试员"
              style="width: 240px"
              @search="() => { recordPagination.current = 1; fetchRecords() }"
            />
          </div>
          <a-table
            row-key="id"
            size="small"
            :loading="recordsLoading"
            :columns="recordColumns"
            :data-source="recordRows"
            :pagination="recordPagination"
            :scroll="{ x: 1300 }"
            @change="handleRecordTableChange"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'studentName'">
                <div class="primary-cell">
                  <span>{{ record.studentName || '未填写姓名' }}</span>
                  <small>ID {{ record.studentId || '-' }}</small>
                </div>
              </template>
              <template v-else-if="column.key === 'assessmentDate'">
                {{ formatDate(record.assessmentDate) }}
              </template>
              <template v-else-if="column.key === 'age'">
                {{ record.ageYears }}岁{{ record.ageMonths }}个月{{ record.ageDays }}天
              </template>
              <template v-else-if="column.key === 'dataStatus'">
                <a-tooltip :title="record.dataStatus">
                  <span class="ellipsis">{{ record.dataStatus || '-' }}</span>
                </a-tooltip>
              </template>
              <template v-else-if="column.key === 'createdTime'">
                {{ formatDateTime(record.createdTime) }}
              </template>
              <template v-else-if="column.key === 'action'">
                <a-space>
                  <a @click="openReportRow(record)">解释报告</a>
                  <a @click="openBookletRow(record)">记录册</a>
                  <a @click="openBookletPdfRow(record)">记录册PDF</a>
                </a-space>
              </template>
            </template>
          </a-table>
        </a-tab-pane>
      </a-tabs>
    </div>

    <a-drawer
      v-model:open="drawerOpen"
      width="960"
      class="pep3-drawer"
      title="PEP-3测评录入"
      :destroy-on-close="false"
    >
      <a-spin :spinning="templateLoading">
        <div class="editor-layout">
          <div class="editor-main">
            <a-form layout="vertical">
              <div class="form-grid">
                <a-form-item label="儿童姓名">
                  <a-input v-model:value="editor.studentName" placeholder="儿童姓名" />
                </a-form-item>
                <a-form-item label="学员ID">
                  <a-input-number v-model:value="editor.studentId" :min="1" style="width: 100%" placeholder="系统学员ID" />
                </a-form-item>
                <a-form-item label="出生日期">
                  <a-date-picker v-model:value="editor.birthDate" style="width: 100%" />
                </a-form-item>
                <a-form-item label="评估日期">
                  <a-date-picker v-model:value="editor.assessmentDate" style="width: 100%" />
                </a-form-item>
                <a-form-item label="测试员">
                  <a-input v-model:value="editor.examinerName" placeholder="默认取当前登录员工" />
                </a-form-item>
                <a-form-item label="缺题评分">
                  <a-switch v-model:checked="editor.allowMissingItems" checked-children="允许" un-checked-children="不允许" />
                </a-form-item>
              </div>
              <a-form-item label="备注">
                <a-textarea v-model:value="editor.remark" :rows="2" placeholder="测评过程备注" />
              </a-form-item>
            </a-form>

            <a-divider orientation="left">逐题得分</a-divider>
            <a-collapse v-model:active-key="activeGroupKeys" class="item-collapse">
              <a-collapse-panel
                v-for="group in itemGroups"
                :key="group.groupCode"
                :header="group.title"
              >
                <div class="item-list">
                  <div v-for="item in group.items" :key="item.itemNo" class="score-row">
                    <div class="score-row__main">
                      <div class="score-row__title">
                        <a-tag>{{ item.domainCode }}</a-tag>
                        <span>{{ item.itemTitle }}</span>
                      </div>
                      <div class="score-row__meta">
                        <span v-if="item.materials">材料：{{ item.materials }}</span>
                        <span v-if="item.sourcePages?.length">页码：{{ item.sourcePages.join(', ') }}</span>
                      </div>
                      <a-popover title="评分标准" placement="rightTop">
                        <template #content>
                          <div class="standard-popover">
                            <p v-for="option in item.scoreOptions" :key="option.value">
                              <b>{{ option.label }}</b>{{ option.description }}
                            </p>
                          </div>
                        </template>
                        <a class="standard-link">评分标准</a>
                      </a-popover>
                    </div>
                    <a-radio-group v-model:value="editor.itemScores[item.itemNo]" button-style="solid" size="small">
                      <a-radio-button v-for="option in item.scoreOptions" :key="option.value" :value="option.value">
                        {{ option.value }}
                      </a-radio-button>
                    </a-radio-group>
                  </div>
                </div>
              </a-collapse-panel>
            </a-collapse>

            <a-divider orientation="left">照顾者报告原始分</a-divider>
            <div class="raw-score-grid">
              <a-form-item v-for="field in caregiverRawScoreFields" :key="field.scaleCode" :label="rawScoreLabel(field)">
                <a-input-number
                  v-model:value="editor.rawScores[field.scaleCode]"
                  :min="field.minScore"
                  :max="field.maxScore"
                  style="width: 100%"
                />
              </a-form-item>
            </div>
          </div>

          <aside class="editor-side">
            <div class="side-block">
              <div class="side-block__title">当前进度</div>
              <a-progress type="circle" :percent="editorPercent" :width="92" />
              <div class="side-metrics">
                <div><span>已评分题目</span><b>{{ answeredItemCount }}</b></div>
                <div><span>照顾者原始分</span><b>{{ rawScoreCount }}/{{ caregiverRawScoreTotal }}</b></div>
                <div><span>缺题</span><b>{{ currentProgress?.missingItemCount ?? '-' }}</b></div>
              </div>
            </div>

            <div class="side-block">
              <div class="side-block__title">发展/行为原始分</div>
              <div class="side-block__hint">由逐题 0/1/2 自动汇总，不需要手填。</div>
              <div class="auto-score-list">
                <div v-for="row in autoRawScoreRows" :key="row.scaleCode" class="auto-score-item">
                  <div class="auto-score-item__meta">
                    <strong>{{ row.scaleCode }}</strong>
                    <span>{{ row.scaleName }}</span>
                  </div>
                  <div class="auto-score-item__value">
                    <b>{{ row.rawScore }}</b>
                    <span>{{ row.answeredItemCount }}/{{ row.itemCount }}</span>
                  </div>
                </div>
              </div>
            </div>
          </aside>
        </div>
      </a-spin>

      <template #footer>
        <div class="drawer-footer">
          <a-space>
            <a-button @click="drawerOpen = false">关闭</a-button>
            <a-button :loading="saving" @click="saveDraft(false)">保存草稿</a-button>
            <a-button type="primary" :loading="submitting" @click="submitDraft">提交正式记录</a-button>
          </a-space>
        </div>
      </template>
    </a-drawer>

    <a-modal
      v-model:open="reportModalOpen"
      title="PEP-3解释性报告"
      width="980px"
      :footer="null"
      :confirm-loading="previewLoading"
    >
      <div class="preview-content">
        <section v-for="section in currentReport?.sections || []" :key="section.sectionCode" class="preview-section">
          <h3>{{ section.title }}</h3>
          <div v-if="section.fields?.length" class="section-fields">
            <div v-for="field in section.fields" :key="field.key" class="section-field">
              <span class="section-field__label">{{ field.label }}</span>
              <strong class="section-field__value">{{ field.value || '-' }}</strong>
            </div>
          </div>
          <div v-if="section.textItems?.length" class="section-text-list">
            <p v-for="text in section.textItems" :key="text">{{ text }}</p>
          </div>
          <a-table
            v-if="section.table"
            size="small"
            :pagination="false"
            :columns="antTableColumns(section)"
            :data-source="section.table.rows"
            :scroll="{ x: 760 }"
          />
        </section>
      </div>
    </a-modal>

    <a-modal
      v-model:open="bookletModalOpen"
      title="PEP-3测试员记录册"
      width="1080px"
      :footer="null"
      :confirm-loading="previewLoading"
    >
      <div class="preview-content">
        <a-alert
          v-if="currentBooklet?.warnings?.length"
          type="warning"
          show-icon
          class="mb-3"
          :message="currentBooklet.warnings.join('；')"
        />
        <a-tabs>
          <a-tab-pane v-for="page in currentBooklet?.pages || []" :key="page.pageNo" :tab="`第${page.pageNo}页`">
            <section v-for="section in page.sections" :key="section.sectionCode" class="preview-section">
              <h3>{{ section.title }}</h3>
              <div v-if="section.fields?.length" class="section-fields">
                <div v-for="field in section.fields" :key="field.key" class="section-field">
                  <span class="section-field__label">{{ field.label }}</span>
                  <strong class="section-field__value">{{ field.value || '-' }}</strong>
                </div>
              </div>
              <div v-if="section.textItems?.length" class="section-text-list">
                <p v-for="text in section.textItems" :key="text">{{ text }}</p>
              </div>
              <a-table
                v-if="section.table"
                size="small"
                :pagination="false"
                :columns="antTableColumns(section)"
                :data-source="section.table.rows"
                :scroll="{ x: 900 }"
              />
            </section>
          </a-tab-pane>
        </a-tabs>
      </div>
    </a-modal>
  </div>
</template>

<style scoped lang="less">
.pep3-page {
  color: #1f2937;
}

.pep3-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 18px 22px;
  background: #fff;
  border: 1px solid #eef1f5;
  border-radius: 8px;

  h1 {
    margin: 4px 0 0;
    font-size: 22px;
    font-weight: 650;
    line-height: 30px;
  }
}

.pep3-eyebrow {
  font-size: 13px;
  color: #667085;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
  margin-top: 12px;
}

.summary-item {
  display: flex;
  align-items: center;
  gap: 12px;
  min-height: 86px;
  padding: 16px;
  background: #fff;
  border: 1px solid #eef1f5;
  border-radius: 8px;
}

.summary-item__icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 38px;
  height: 38px;
  color: var(--pro-ant-color-primary);
  background: #eef4ff;
  border-radius: 8px;
  font-size: 18px;
}

.summary-item__value {
  font-size: 24px;
  font-weight: 700;
  line-height: 28px;
}

.summary-item__label {
  margin-top: 4px;
  font-size: 13px;
  color: #667085;
}

.content-panel {
  margin-top: 12px;
  padding: 0 18px 18px;
  background: #fff;
  border: 1px solid #eef1f5;
  border-radius: 8px;
}

.table-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 8px 0 14px;
}

.primary-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;

  span {
    color: #111827;
    font-weight: 600;
  }

  small {
    color: #98a2b3;
  }
}

.progress-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;

  span {
    font-size: 12px;
    color: #667085;
  }
}

.ellipsis {
  display: inline-block;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  vertical-align: bottom;
}

.danger-link {
  color: #d92d20;
}

.editor-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 250px;
  gap: 18px;
}

.editor-main {
  min-width: 0;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0 12px;
}

.item-collapse {
  background: transparent;
}

.item-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.score-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border: 1px solid #eef1f5;
  border-radius: 8px;
  background: #fff;
}

.score-row__title {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #111827;
  font-weight: 600;
}

.score-row__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 4px;
  color: #667085;
  font-size: 12px;
}

.standard-link {
  display: inline-block;
  margin-top: 4px;
  font-size: 12px;
}

.standard-popover {
  max-width: 380px;

  p {
    margin: 0 0 8px;
    color: #475467;
    line-height: 1.6;
  }

  b {
    margin-right: 8px;
    color: #111827;
  }
}

.raw-score-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0 12px;
}

.editor-side {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.side-block {
  padding: 14px;
  background: #f8fafc;
  border: 1px solid #eef1f5;
  border-radius: 8px;
}

.side-block__title {
  margin-bottom: 12px;
  color: #111827;
  font-weight: 650;
}

.side-metrics {
  display: grid;
  gap: 8px;
  margin-top: 14px;

  div {
    display: flex;
    align-items: center;
    justify-content: space-between;
    color: #667085;
  }

  b {
    color: #111827;
  }
}

.side-block__hint {
  margin: -6px 0 10px;
  color: #667085;
  font-size: 12px;
  line-height: 18px;
}

.auto-score-list {
  display: grid;
  gap: 8px;
}

.auto-score-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding: 8px 10px;
  background: #fff;
  border: 1px solid #eef1f5;
  border-radius: 8px;
}

.auto-score-item__meta {
  display: flex;
  flex-direction: column;
  min-width: 0;

  strong {
    color: #111827;
    font-size: 13px;
    line-height: 18px;
  }

  span {
    overflow: hidden;
    color: #667085;
    font-size: 12px;
    line-height: 18px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.auto-score-item__value {
  display: flex;
  flex-direction: column;
  align-items: flex-end;

  b {
    color: #111827;
    font-size: 16px;
    line-height: 20px;
  }

  span {
    color: #98a2b3;
    font-size: 12px;
    line-height: 18px;
  }
}

.drawer-footer {
  display: flex;
  justify-content: flex-end;
}

.preview-content {
  max-height: 68vh;
  overflow-y: auto;
  padding-right: 4px;
}

.preview-section {
  margin-bottom: 18px;

  h3 {
    margin: 0 0 10px;
    font-size: 15px;
    font-weight: 650;
  }
}

.section-fields {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.section-field {
  padding: 10px 12px;
  background: #f8fafc;
  border-radius: 8px;
}

.section-field__label {
  display: block;
  color: #667085;
  font-size: 12px;
}

.section-field__value {
  display: block;
  margin-top: 4px;
  color: #111827;
}

.section-text-list {
  padding: 10px 12px;
  background: #f8fafc;
  border-radius: 8px;

  p {
    margin: 0 0 6px;
  }
}

@media (max-width: 1280px) {
  .summary-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .editor-layout {
    grid-template-columns: 1fr;
  }

  .editor-side {
    order: -1;
  }
}
</style>
