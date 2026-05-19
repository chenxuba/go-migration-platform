<script setup>
import dayjs from 'dayjs'
import { Empty, Modal } from 'ant-design-vue'
import messageService from '@/utils/messageService'
import { useTableColumns } from '@/composables/useTableColumns'
import AssessmentRecordConfigModal from './components/assessment-record-config-modal.vue'
import GenerateIepModal from './components/generate-iep-modal.vue'
import {
  deletePEP3AssessmentRecordApi,
  downloadPEP3AssessmentBookletPdfApi,
  downloadPEP3AssessmentRecordReportInterpretationPdfApi,
  generatePEP3AssessmentRecordReportInterpretationStreamApi,
  getPEP3AssessmentRecordReportInterpretationApi,
  pagePEP3AssessmentRecordsApi,
} from '@/api/edu-center/pep3-assessment'
import {
  deleteERXinAssessmentRecordApi,
  downloadERXinAssessmentRecordReportCombinedPdfApi,
  downloadERXinAssessmentRecordReportInterpretationPdfApi,
  downloadERXinAssessmentRecordReportPdfApi,
  generateERXinAssessmentRecordReportInterpretationStreamApi,
  getERXinAssessmentRecordReportInterpretationApi,
  pageERXinAssessmentRecordsApi,
} from '@/api/edu-center/erxin-assessment'
import {
  deleteAutismDevAssessmentRecordApi,
  downloadAutismDevResultAnalysisWordApi,
  downloadAutismDevSelectedReportPdfApi,
  generateAutismDevAssessmentRecordReportInterpretationStreamApi,
  generateAutismDevResultAnalysisStreamApi,
  getAutismDevAssessmentRecordReportInterpretationApi,
  getAutismDevResultAnalysisApi,
  pageAutismDevAssessmentRecordsApi,
} from '@/api/edu-center/autismdev-assessment'
import {
  deleteShuangxiAAssessmentRecordApi,
  downloadShuangxiADevelopmentProfilePdfApi,
  pageShuangxiAAssessmentRecordsApi,
} from '@/api/edu-center/shuangxi-assessment'
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
const configModalOpen = ref(false)
const configTargetRecord = ref(null)
const reportPreviewUrl = ref('')
const reportPreviewRequestKey = ref(0)
const reportPdfReady = ref(false)
const reportTab = ref('result')
const interpretation = ref(null)
const interpretationLoading = ref(false)
const interpretationGenerating = ref(false)
const interpretationFetched = ref(false)
const interpretationError = ref('')
const interpretationProgress = ref('')
const interpretationStreamingText = ref('')
const interpretationStreamRef = ref(null)
const interpretationShellRef = ref(null)
const interpretationProgressRef = ref(null)
const streamingInterpretationPreview = computed(() => interpretationPreviewFromText(interpretationStreamingText.value))
const activeAutismDevReportSection = ref('assessmentInfo')
const autismDevAnalysis = ref(null)
const autismDevAnalysisLoading = ref(false)
const autismDevAnalysisGenerating = ref(false)
const autismDevAnalysisFetched = ref(false)
const autismDevAnalysisError = ref('')
const autismDevAnalysisProgress = ref('')
const autismDevAnalysisStreamText = ref('')
const autismDevAnalysisRecordKey = ref('')
const autismDevAnalysisWordExporting = ref(false)
const autismDevExportSections = ref([])
const autismDevExportAnalysisLoading = ref(false)
const simpleEmptyImage = Empty.PRESENTED_IMAGE_SIMPLE
const router = useRouter()
let reportPdfReadyTimer = 0
let interpretationAbortController = null
let autismDevAnalysisAbortController = null
let interpretationScrollFrame = 0
let interpretationProgressAnchored = false
const autismDevAnalysisStreamReadableText = computed(() => autismDevAnalysisReadableStreamText(autismDevAnalysisStreamText.value))
const autismDevAnalysisStreamRows = computed(() => autismDevAnalysisPreviewRows(autismDevAnalysisStreamReadableText.value))
const autismDevAnalysisStreamProgressPercent = computed(() => Math.round(autismDevResultAnalysisStreamProgress(autismDevAnalysisStreamText.value) * 100))

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
const erxinExportDimensionOptions = [
  {
    value: 'erxin_result',
    title: '评估记录',
    badge: '01',
    desc: '导出儿心量表评估结果记录PDF。',
    pages: '记录',
    recommended: true,
  },
  {
    value: 'erxin_interpretation',
    title: '报告解读',
    badge: '02',
    desc: '导出已生成的报告解读内容。',
    pages: '报告',
  },
  {
    value: 'erxin_combined',
    title: '记录+报告',
    badge: '03',
    desc: '合并导出评估记录和报告解读。',
    pages: '记录+报告',
  },
]
const autismDevExportDimensionOptions = [
  {
    value: 'assessmentInfo',
    title: '评估情况',
    badge: '01',
    desc: '查看发展能力计分汇总与评估基本情况。',
    pages: '评估情况',
    recommended: true,
  },
  {
    value: 'resultAnalysis',
    title: '评估结果分析',
    badge: '02',
    desc: 'AI生成八大领域的现况、优势、弱项与训练目标。',
    pages: '结果分析',
  },
  {
    value: 'interpretation',
    title: '报告解读',
    badge: '03',
    desc: 'AI生成综合报告解读，辅助IEP目标制定。',
    pages: '报告解读',
  },
  {
    value: 'training',
    title: '训练效果',
    badge: '04',
    desc: '查看与导出阶段训练效果分析。',
    pages: '训练效果',
  },
  {
    value: 'developmentProfile',
    title: '发展情况剖面图',
    badge: '05',
    desc: '查看发展能力剖面图。',
    pages: '发展剖面',
  },
  {
    value: 'behaviorProfile',
    title: '情绪行为表现图',
    badge: '06',
    desc: '查看情绪行为表现图。',
    pages: '情绪行为',
  },
]
const shuangxiExportDimensionOptions = [
  {
    value: 'shuangxi_development_profile',
    title: '综合发展侧面图',
    badge: '01',
    desc: '导出双溪课程评量表A综合发展侧面图。',
    pages: 'PDF',
    recommended: true,
  },
]
const defaultExportDimension = exportDimensionOptions.find(item => item.recommended)?.value || 'all'
const selectedExportDimension = ref(defaultExportDimension)
const reportModuleValues = ['test_score', 'development_profile', 'score_and_profile', 'scoring_tables']
const reportModuleOptions = exportDimensionOptions.filter(item => reportModuleValues.includes(item.value))
const defaultReportModule = reportModuleOptions.find(item => item.recommended)?.value || reportModuleOptions[0]?.value || 'test_score'
const activeReportModule = ref(defaultReportModule)
const activeExportDimensionOptions = computed(() => {
  if (isShuangxiARecord(exportTargetRecord.value))
    return shuangxiExportDimensionOptions
  if (isERXinRecord(exportTargetRecord.value))
    return erxinExportDimensionOptions
  if (isAutismDevRecord(exportTargetRecord.value))
    return autismDevExportDimensionOptions
  return exportDimensionOptions
})
const autismDevEnabledExportSections = computed(() => autismDevExportDimensionOptions
  .filter(option => isAutismDevExportSectionEnabled(option.value))
  .map(option => option.value))
const autismDevExportAllSelected = computed(() => {
  const enabled = autismDevEnabledExportSections.value
  return !!enabled.length && enabled.every(section => autismDevExportSections.value.includes(section))
})
const exportModalWidth = computed(() => {
  if (isAutismDevRecord(exportTargetRecord.value))
    return 700
  if (isERXinRecord(exportTargetRecord.value))
    return 760
  return 700
})

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
    width: 180,
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

async function getDownloadErrorMessage(error, fallback) {
  const blobText = await error?.response?.data?.text?.()
  if (blobText) {
    try {
      const payload = JSON.parse(blobText)
      return payload?.message || fallback
    }
    catch {
      return blobText || fallback
    }
  }
  return getErrorMessage(error, fallback)
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

function isERXinRecord(record) {
  const source = String(record?._recordSource || record?.assessmentCode || '').trim().toUpperCase()
  return source === 'ERXIN' || source.startsWith('ERXIN')
}

function isAutismDevRecord(record) {
  const source = String(record?._recordSource || record?.assessmentCode || '').trim().toUpperCase()
  return source === 'AUTISMDEV'
}

function isShuangxiARecord(record) {
  const source = String(record?._recordSource || record?.assessmentCode || '').trim().toUpperCase()
  return source === 'SHUANGXI_A' || source === 'SHUANGXIA' || source.startsWith('SHUANGXI')
}

function recordSourceType(record) {
  if (isShuangxiARecord(record))
    return 'SHUANGXI_A'
  if (isAutismDevRecord(record))
    return 'AUTISMDEV'
  if (isERXinRecord(record))
    return 'ERXIN'
  return 'PEP3'
}

function recordActionKey(record) {
  if (!record?.id)
    return ''
  return `${recordSourceType(record)}-${record.id}`
}

function markRecordSource(record, source) {
  return {
    ...record,
    _recordSource: source,
  }
}

function recordSortValue(record) {
  const values = [record?.createdTime]
  for (const value of values) {
    const parsed = dayjs(value)
    if (parsed.isValid())
      return parsed.valueOf()
  }
  return 0
}

function compareRecordDesc(left, right) {
  const timeDiff = recordSortValue(right) - recordSortValue(left)
  if (timeDiff)
    return timeDiff
  return Number(right?.id || 0) - Number(left?.id || 0)
}

function currentReportIsERXin() {
  return isERXinRecord(currentReport.value?.record)
}

function currentReportIsAutismDev() {
  return isAutismDevRecord(currentReport.value?.record)
}

function currentReportIsShuangxiA() {
  return isShuangxiARecord(currentReport.value?.record)
}

function autismDevReportSectionOption(value) {
  return autismDevExportDimensionOptions.find(item => item.value === value) || autismDevExportDimensionOptions[0]
}

function autismDevReportSectionTitle(value = activeAutismDevReportSection.value) {
  return autismDevReportSectionOption(value)?.title || '评估情况'
}

function autismDevReportSectionDesc(value = activeAutismDevReportSection.value) {
  return autismDevReportSectionOption(value)?.desc || '查看孤独症儿童发展评估报告。'
}

function defaultAutismDevReportSection() {
  return autismDevExportDimensionOptions.find(item => item.recommended)?.value || 'assessmentInfo'
}

function normalizeAutismDevAnalysis(value) {
  const data = value || {}
  const rows = Array.isArray(data.rows) ? data.rows : []
  return {
    title: String(data.title || ''),
    model: String(data.model || ''),
    generatedBy: String(data.generatedBy || ''),
    generatedAt: String(data.generatedAt || ''),
    rows: rows.map(row => ({
      domain: String(row?.domain || ''),
      status: String(row?.status || ''),
      strengths: String(row?.strengths || ''),
      weaknesses: String(row?.weaknesses || ''),
      targets: String(row?.targets || ''),
    })),
  }
}

function emptyAutismDevAnalysis() {
  return normalizeAutismDevAnalysis({
    title: '孤独症儿童评估结果分析表',
    rows: ['感知觉', '粗大动作', '精细动作', '语言与沟通', '认知', '社会交往', '生活自理', '情绪与行为']
      .map(domain => ({ domain })),
  })
}

function canonicalAutismDevAnalysisDomain(value) {
  const normalized = String(value || '').replace(/\s+/g, '').replaceAll('和', '与').trim()
  return emptyAutismDevAnalysis().rows.find((row) => {
    const candidate = String(row.domain || '').replace(/\s+/g, '').replaceAll('和', '与').trim()
    return normalized === candidate
  })?.domain || ''
}

function autismDevStrengthWeaknessText(row) {
  const strengths = String(row?.strengths || '').trim()
  const weaknesses = String(row?.weaknesses || '').trim()
  const lines = []
  if (strengths)
    lines.push(`优势：${strengths}`)
  if (weaknesses)
    lines.push(`劣势：${weaknesses}`)
  return lines.join('\n')
}

function autismDevAnalysisPreviewRows(readableText) {
  const rows = emptyAutismDevAnalysis().rows.map(row => ({ ...row }))
  const byDomain = new Map(rows.map(row => [row.domain, row]))
  const text = String(readableText || '').trim()
  if (!text || text.startsWith('正在连接AI生成服务'))
    return rows

  let activeRow = null
  let activeField = ''
  const appendField = (value) => {
    if (!activeRow || !activeField)
      return
    const text = String(value || '').trim()
    if (!text)
      return
    activeRow[activeField] = [activeRow[activeField], text].filter(Boolean).join(activeRow[activeField] ? '\n' : '')
  }

  for (const rawLine of text.split(/\n+/)) {
    const line = rawLine.trim()
    if (!line)
      continue
    const domainMatch = line.match(/^【(.+?)】?$/)
    if (domainMatch) {
      const domain = canonicalAutismDevAnalysisDomain(domainMatch[1])
      activeRow = domain ? byDomain.get(domain) : null
      activeField = ''
      continue
    }
    const fieldMatch = line.match(/^(能力现状描述|优势|劣势|训练目标)：(.*)$/)
    if (fieldMatch) {
      const [, label, value] = fieldMatch
      activeField = {
        能力现状描述: 'status',
        优势: 'strengths',
        劣势: 'weaknesses',
        训练目标: 'targets',
      }[label] || ''
      appendField(value)
      continue
    }
    appendField(line)
  }
  return rows
}

function isAutismDevAnalysisReadableField(key) {
  return ['domain', 'status', 'strengths', 'weaknesses', 'targets'].includes(key)
}

function autismDevAnalysisStreamLabel(key) {
  return {
    status: '能力现状描述',
    strengths: '优势',
    weaknesses: '劣势',
    targets: '训练目标',
  }[key] || key
}

function autismDevAnalysisReadableStreamText(raw) {
  const text = String(raw || '').trim()
  if (!text)
    return '正在连接AI生成服务，准备读取评估记录...'

  let output = ''
  let token = ''
  let mode = 'outside'
  let currentKey = ''
  let visibleKey = ''
  let expectingValue = false
  let escaping = false
  let lastWasNewline = true

  const writeText = (value) => {
    if (!value)
      return
    output += value
    lastWasNewline = value.endsWith('\n')
  }

  const startVisibleField = (key) => {
    if (output && !lastWasNewline)
      writeText('\n')
    if (key === 'domain') {
      if (output)
        writeText('\n')
      writeText('【')
      return
    }
    writeText(`${autismDevAnalysisStreamLabel(key)}：`)
  }

  const endVisibleField = (key) => {
    if (key === 'domain') {
      writeText('】\n')
      return
    }
    if (output && !lastWasNewline)
      writeText('\n')
  }

  const writeEscaped = (char) => {
    if (['n', 'r', 't'].includes(char)) {
      writeText(' ')
      return
    }
    if (char === '"')
      writeText('"')
    else if (char === '/')
      writeText('/')
    else if (char === '\\')
      writeText('\\')
    else
      writeText(char)
  }

  for (const char of text) {
    if (escaping) {
      if (mode === 'visibleValue')
        writeEscaped(char)
      else if (mode === 'key')
        token += char
      escaping = false
      continue
    }
    if (char === '\\') {
      escaping = true
      continue
    }
    if (mode === 'outside') {
      if (char === '"') {
        token = ''
        if (expectingValue) {
          if (isAutismDevAnalysisReadableField(currentKey)) {
            visibleKey = currentKey
            startVisibleField(visibleKey)
            mode = 'visibleValue'
          }
          else {
            mode = 'hiddenValue'
          }
        }
        else {
          mode = 'key'
        }
      }
      else if (char === ':') {
        expectingValue = !!currentKey
      }
      else if (char === ',' || char === '}' || char === ']') {
        if (expectingValue) {
          expectingValue = false
          currentKey = ''
        }
      }
    }
    else if (mode === 'key') {
      if (char === '"') {
        currentKey = token
        token = ''
        mode = 'outside'
      }
      else {
        token += char
      }
    }
    else if (mode === 'visibleValue') {
      if (char === '"') {
        endVisibleField(visibleKey)
        mode = 'outside'
        expectingValue = false
        currentKey = ''
        visibleKey = ''
      }
      else {
        writeText(char)
      }
    }
    else if (mode === 'hiddenValue') {
      if (char === '"') {
        mode = 'outside'
        expectingValue = false
        currentKey = ''
      }
    }
  }

  const normalized = output.trimEnd()
  if (normalized)
    return normalized
  const fallback = text
    .replaceAll('{', '')
    .replaceAll('}', '')
    .replaceAll('[', '')
    .replaceAll(']', '')
    .replaceAll('"', '')
    .replaceAll(',', '')
    .trim()
  return fallback || '正在连接AI生成服务，准备读取评估记录...'
}

function autismDevResultAnalysisStreamProgress(raw) {
  const length = String(raw || '').trim().length
  const floor = 0.12
  const ceiling = 0.975
  if (length <= 0)
    return floor
  if (length <= 360)
    return floor + (0.32 - floor) * (length / 360)
  if (length <= 1600)
    return 0.32 + (0.88 - 0.32) * ((length - 360) / 1240)
  const tail = 1 - Math.exp(-(length - 1600) / 720)
  return 0.88 + (ceiling - 0.88) * Math.min(Math.max(tail, 0), 1)
}

function autismDevAnalysisIsEmpty(value = autismDevAnalysis.value) {
  const rows = Array.isArray(value?.rows) ? value.rows : []
  return !rows.some(row => String(row?.status || row?.strengths || row?.weaknesses || row?.targets || '').trim())
}

function autismDevAnalysisForExport() {
  return autismDevAnalysisIsEmpty() ? null : autismDevAnalysis.value
}

function isAutismDevExportSectionEnabled(section) {
  if (section === 'interpretation')
    return !(interpretationLoading.value || interpretationGenerating.value)
  if (section !== 'resultAnalysis')
    return true
  return !autismDevAnalysisGenerating.value && !autismDevAnalysisLoading.value && !autismDevAnalysisIsEmpty()
}

function autismDevExportSectionStatus(section) {
  if (section === 'interpretation') {
    if (interpretationGenerating.value)
      return '生成中'
    if (interpretationLoading.value)
      return '读取中'
    return ''
  }
  if (section !== 'resultAnalysis')
    return ''
  if (autismDevAnalysisGenerating.value)
    return '生成中'
  if (autismDevAnalysisLoading.value || autismDevExportAnalysisLoading.value)
    return '读取中'
  if (autismDevAnalysisIsEmpty())
    return '未生成'
  return ''
}

function normalizeAutismDevExportSections(sections = []) {
  const allowed = autismDevExportDimensionOptions.map(item => item.value)
  const enabled = autismDevEnabledExportSections.value
  return sections
    .filter(section => allowed.includes(section) && enabled.includes(section))
    .filter((section, index, array) => array.indexOf(section) === index)
}

function setAutismDevExportSections(sections = []) {
  const normalized = normalizeAutismDevExportSections(sections)
  autismDevExportSections.value = normalized.length ? normalized : normalizeAutismDevExportSections([defaultAutismDevReportSection()])
}

function toggleAutismDevExportSection(section) {
  if (!isAutismDevExportSectionEnabled(section))
    return
  const next = new Set(autismDevExportSections.value)
  if (next.has(section))
    next.delete(section)
  else
    next.add(section)
  setAutismDevExportSections(Array.from(next))
  if (section === 'interpretation' && next.has(section) && !interpretationFetched.value && !interpretationLoading.value)
    void loadSavedInterpretation(exportTargetRecord.value)
}

function toggleAllAutismDevExportSections() {
  if (autismDevExportAllSelected.value) {
    autismDevExportSections.value = []
    return
  }
  setAutismDevExportSections(autismDevEnabledExportSections.value)
}

function reportTitleForRecord(record) {
  if (isShuangxiARecord(record))
    return record?.assessmentName || '双溪课程评量表A'
  if (isAutismDevRecord(record))
    return record?.assessmentName || '孤独症儿童发展评估报告'
  return isERXinRecord(record)
    ? (record?.assessmentName || '儿心量表-II发育行为评估报告')
    : 'PEP-3测试员记录册'
}

function reportModalHint() {
  if (currentReportIsShuangxiA())
    return '查看双溪课程评量表A综合发展侧面图'
  if (currentReportIsAutismDev())
    return '按评估情况、结果分析、报告解读、训练效果和剖面图查看报告'
  return currentReportIsERXin() ? '查看儿心量表评估报告内容' : '按记录册导出维度查看报告内容'
}

function reportFrameTitle() {
  if (currentReportIsShuangxiA())
    return '双溪综合发展侧面图PDF预览'
  if (currentReportIsAutismDev())
    return `孤独症儿童发展评估报告-${autismDevReportSectionTitle()}PDF预览`
  return currentReportIsERXin() ? '儿心量表评估报告PDF预览' : 'PEP-3记录册PDF预览'
}

function reportInterpretationDefaultTitle(record = currentReport.value?.record) {
  if (isAutismDevRecord(record))
    return '孤独症儿童发展评估报告解读'
  return isERXinRecord(record) ? '儿心量表报告解读' : 'PEP-3报告解读'
}

function reportInterpretationGeneratingHint(record = currentReport.value?.record) {
  if (isAutismDevRecord(record))
    return '正在读取孤独症儿童发展评估结果...'
  return isERXinRecord(record) ? '正在分析全量表与五大能区结果...' : '正在分析PEP-3评估结果...'
}

function reportInterpretationDomainSectionTitle(record = currentReport.value?.record) {
  if (isAutismDevRecord(record))
    return '八大领域表现'
  return isERXinRecord(record) ? '能区表现' : '领域表现'
}

const interpretationFieldNameTextSet = new Set([
  'domain',
  'domainname',
  'name',
  'title',
  'analysis',
  'content',
  'text',
  'value',
  'summary',
  'description',
  'suggestion',
  'suggestions',
  'note',
  'notes',
  'domainanalysis',
])

function isInterpretationFieldNameText(value) {
  const text = String(value || '').trim().replace(/^[:：]+|[:：]+$/g, '').toLowerCase()
  return interpretationFieldNameTextSet.has(text)
}

function cleanInterpretationText(value, maxLength = 120) {
  let text = String(value || '').trim().replace(/^["'`]+|["'`]+$/g, '').trim()
  text = text.replace(/^\d+[.、)）]\s*/, '').trim()
  if (!text || isInterpretationFieldNameText(text))
    return ''
  const chars = Array.from(text)
  if (maxLength > 0 && chars.length > maxLength) {
    let cut = maxLength
    for (let index = maxLength; index > maxLength - 20 && index > 0; index -= 1) {
      if ('。；;，,、'.includes(chars[index - 1])) {
        cut = index
        break
      }
    }
    text = chars.slice(0, cut).join('').trim()
  }
  return text
}

function interpretationObjectText(source, keys) {
  if (!source || typeof source !== 'object')
    return ''
  for (const key of keys) {
    const text = cleanInterpretationText(source[key], 160)
    if (text)
      return text
  }
  return ''
}

function interpretationTextFromValue(item, key = '') {
  if (!item)
    return ''
  if (typeof item === 'object' && !Array.isArray(item)) {
    const domain = interpretationObjectText(item, ['domain', 'domainName', 'name', 'title'])
    let body = interpretationObjectText(item, ['analysis', 'text', 'content', 'summary', 'description', 'value', 'suggestion', 'note'])
    if (!body) {
      for (const [field, fieldValue] of Object.entries(item)) {
        if (isInterpretationFieldNameText(field))
          continue
        body = interpretationTextFromValue(fieldValue, key)
        if (body)
          break
      }
    }
    if (!body)
      return ''
    if (key === 'domainAnalysis' && domain && !body.startsWith(`${domain}：`) && !body.startsWith(`${domain}:`))
      return `${domain}：${body}`
    return body
  }
  const maxLength = key === 'suggestions' ? 80 : 120
  return cleanInterpretationText(item, maxLength)
}

function stringList(value, key = '') {
  if (!Array.isArray(value))
    return []
  const seen = new Set()
  const result = []
  value.forEach((item) => {
    const text = interpretationTextFromValue(item, key)
    if (!text || seen.has(text))
      return
    seen.add(text)
    result.push(text)
  })
  return result
}

function normalizeInterpretation(value) {
  const data = value || {}
  return {
    title: String(data.title || ''),
    model: String(data.model || ''),
    generatedBy: String(data.generatedBy || ''),
    generatedAt: String(data.generatedAt || ''),
    summary: cleanInterpretationText(data.summary, 140),
    domainAnalysis: stringList(data.domainAnalysis, 'domainAnalysis'),
    suggestions: stringList(data.suggestions, 'suggestions'),
    notes: stringList(data.notes, 'notes'),
  }
}

function previewInterpretationFromMap(value) {
  const data = value || {}
  return {
    summary: cleanInterpretationText(data.summary, 140),
    domainAnalysis: stringList(data.domainAnalysis, 'domainAnalysis'),
    suggestions: stringList(data.suggestions, 'suggestions'),
    notes: stringList(data.notes, 'notes'),
  }
}

function interpretationPreviewFromText(raw) {
  const text = String(raw || '').trim()
  if (!text)
    return previewInterpretationFromMap({})
  try {
    const decoded = JSON.parse(text)
    if (decoded && typeof decoded === 'object')
      return previewInterpretationFromMap(decoded)
  }
  catch {
  }
  return {
    summary: cleanInterpretationText(extractPartialJsonStringValue(text, 'summary'), 140),
    domainAnalysis: extractPartialJsonArrayValues(text, 'domainAnalysis'),
    suggestions: extractPartialJsonArrayValues(text, 'suggestions'),
    notes: extractPartialJsonArrayValues(text, 'notes'),
  }
}

function interpretationIsEmpty(value = interpretation.value) {
  if (!value)
    return true
  return !String(value.summary || '').trim()
    && !stringList(value.domainAnalysis, 'domainAnalysis').length
    && !stringList(value.suggestions, 'suggestions').length
    && !stringList(value.notes, 'notes').length
}

function streamingPreviewIsEmpty(value = streamingInterpretationPreview.value) {
  return !String(value?.summary || '').trim()
    && !stringList(value?.domainAnalysis, 'domainAnalysis').length
    && !stringList(value?.suggestions, 'suggestions').length
    && !stringList(value?.notes, 'notes').length
}

function extractPartialJsonStringValue(text, key) {
  const match = new RegExp(`"${key}"\\s*:\\s*"((?:\\\\.|[^"\\\\])*)`).exec(text)
  return match ? decodePartialJsonString(match[1] || '') : ''
}

function extractPartialJsonArrayValues(text, key) {
  const keyIndex = text.indexOf(`"${key}"`)
  if (keyIndex < 0)
    return []
  const arrayStart = text.indexOf('[', keyIndex)
  if (arrayStart < 0)
    return []
  const arrayEnd = text.indexOf(']', arrayStart)
  const body = text.slice(arrayStart + 1, arrayEnd >= 0 ? arrayEnd : text.length)
  const objectValues = extractPartialJsonObjectArrayValues(body, key)
  if (objectValues.length)
    return stringList(objectValues, key)
  const values = Array.from(body.matchAll(/"((?:\\.|[^"\\])*)"/g))
    .map(match => decodePartialJsonString(match[1] || '').trim())
    .filter(Boolean)
  const trailing = extractTrailingPartialJsonArrayString(body)
  if (trailing)
    values.push(trailing)
  return stringList(values, key)
}

function extractPartialJsonObjectArrayValues(body, key) {
  const objectFields = key === 'domainAnalysis'
    ? ['analysis', 'text', 'content', 'summary', 'description', 'value']
    : ['text', 'content', 'summary', 'description', 'value', 'suggestion', 'note']
  const fieldPattern = objectFields.join('|')
  if (!new RegExp(`"(${fieldPattern})"\\s*:`).test(body))
    return []
  const completeValues = []
  for (const objectMatch of body.matchAll(/\{([^{}]*)\}/g)) {
    const objectText = objectMatch[1] || ''
    const value = extractObjectFieldString(objectText, objectFields)
    if (!value)
      continue
    if (key !== 'domainAnalysis') {
      completeValues.push(value)
      continue
    }
    const domain = extractObjectFieldString(objectText, ['domain', 'domainName', 'name', 'title'])
    completeValues.push(domain ? `${domain}：${value}` : value)
  }
  if (completeValues.length)
    return completeValues
  const values = []
  const fieldRegex = new RegExp(`"(?:${fieldPattern})"\\s*:\\s*"((?:\\\\.|[^"\\\\])*)"`,'g')
  for (const match of body.matchAll(fieldRegex))
    values.push(decodePartialJsonString(match[1] || '').trim())
  return values
}

function extractObjectFieldString(body, keys) {
  const fieldPattern = keys.join('|')
  const match = new RegExp(`"(?:${fieldPattern})"\\s*:\\s*"((?:\\\\.|[^"\\\\])*)"`).exec(body)
  return match ? decodePartialJsonString(match[1] || '').trim() : ''
}

function extractTrailingPartialJsonArrayString(body) {
  let quoteIndex = -1
  let escaped = false
  let inString = false
  for (let index = 0; index < body.length; index += 1) {
    const char = body[index]
    if (escaped) {
      escaped = false
      continue
    }
    if (char === '\\') {
      escaped = true
      continue
    }
    if (char === '"') {
      inString = !inString
      quoteIndex = index
    }
  }
  if (!inString || quoteIndex < 0 || quoteIndex >= body.length - 1)
    return ''
  return decodePartialJsonString(body.slice(quoteIndex + 1)).trim()
}

function decodePartialJsonString(value) {
  try {
    return JSON.parse(`"${value}"`)
  }
  catch {
    return String(value || '')
      .replace(/\\"/g, '"')
      .replace(/\\n/g, '\n')
      .replace(/\\\\/g, '\\')
  }
}

function resetInterpretationState() {
  if (interpretationAbortController) {
    interpretationAbortController.abort()
    interpretationAbortController = null
  }
  if (interpretationScrollFrame) {
    window.cancelAnimationFrame(interpretationScrollFrame)
    interpretationScrollFrame = 0
  }
  interpretationProgressAnchored = false
  reportTab.value = 'result'
  interpretation.value = null
  interpretationLoading.value = false
  interpretationGenerating.value = false
  interpretationFetched.value = false
  interpretationError.value = ''
  interpretationProgress.value = ''
  interpretationStreamingText.value = ''
}

function resetAutismDevReportState() {
  if (autismDevAnalysisAbortController) {
    autismDevAnalysisAbortController.abort()
    autismDevAnalysisAbortController = null
  }
  activeAutismDevReportSection.value = defaultAutismDevReportSection()
  autismDevAnalysis.value = emptyAutismDevAnalysis()
  autismDevAnalysisLoading.value = false
  autismDevAnalysisGenerating.value = false
  autismDevAnalysisFetched.value = false
  autismDevAnalysisError.value = ''
  autismDevAnalysisProgress.value = ''
  autismDevAnalysisStreamText.value = ''
  autismDevAnalysisRecordKey.value = ''
  autismDevExportSections.value = []
  autismDevExportAnalysisLoading.value = false
}

function interpretationSectionItems(value, key) {
  return stringList(value?.[key], key)
}

function scrollInterpretationStreamToBottom() {
  if (!interpretationGenerating.value)
    return
  if (interpretationScrollFrame)
    window.cancelAnimationFrame(interpretationScrollFrame)
  interpretationScrollFrame = window.requestAnimationFrame(() => {
    interpretationScrollFrame = 0
    const streamElement = interpretationStreamRef.value
    if (streamElement)
      streamElement.scrollTo({ top: streamElement.scrollHeight, behavior: 'smooth' })
  })
}

watch([interpretationStreamingText, streamingInterpretationPreview], () => {
  nextTick(scrollInterpretationStreamToBottom)
})

function scrollInterpretationProgressIntoView() {
  if (interpretationProgressAnchored)
    return
  nextTick(() => {
    const progressElement = interpretationProgressRef.value
    const modalBodyElement = progressElement?.closest('.ant-modal-body')
    if (!progressElement || !modalBodyElement)
      return
    const progressRect = progressElement.getBoundingClientRect()
    const bodyRect = modalBodyElement.getBoundingClientRect()
    const targetTop = modalBodyElement.scrollTop + progressRect.top - bodyRect.top - 16
    modalBodyElement.scrollTo({
      top: Math.max(0, targetTop),
      behavior: 'smooth',
    })
    interpretationProgressAnchored = true
  })
}

function exportDimensionTitle(value) {
  if (value === 'shuangxi_development_profile')
    return '综合发展侧面图'
  if (value === 'pep3_interpretation')
    return '报告解读'
  if (isAutismDevRecord(exportTargetRecord.value))
    return autismDevExportDimensionOptions.find(item => item.value === value)?.title || '评估报告'
  return activeExportDimensionOptions.value.find(item => item.value === value)?.title || '全维度导出'
}

function exportDimensionPages(value) {
  if (value === 'shuangxi_development_profile')
    return 'PDF'
  if (value === 'pep3_interpretation')
    return '报告'
  if (isAutismDevRecord(exportTargetRecord.value))
    return autismDevExportDimensionOptions.find(item => item.value === value)?.pages || '报告'
  return activeExportDimensionOptions.value.find(item => item.value === value)?.pages || '第 1-26 页'
}

function exportDimensionDesc(value) {
  if (value === 'shuangxi_development_profile')
    return '导出双溪课程评量表A综合发展侧面图。'
  if (value === 'pep3_interpretation')
    return '导出已生成的PEP-3报告解读内容。'
  if (isAutismDevRecord(exportTargetRecord.value))
    return autismDevExportDimensionOptions.find(item => item.value === value)?.desc || '导出孤独症儿童发展评估报告。'
  return activeExportDimensionOptions.value.find(item => item.value === value)?.desc || '导出完整测试员记录册，包含所有维度与分析表。'
}

function exportModalTitle() {
  if (isShuangxiARecord(exportTargetRecord.value))
    return '导出双溪报告'
  if (isAutismDevRecord(exportTargetRecord.value))
    return '导出孤独症评估报告'
  return isERXinRecord(exportTargetRecord.value) ? '导出儿心报告' : '导出记录册'
}

function exportModalHint() {
  if (isShuangxiARecord(exportTargetRecord.value))
    return '导出双溪课程评量表A综合发展侧面图'
  if (isAutismDevRecord(exportTargetRecord.value))
    return '按评估情况、结果分析、训练效果和剖面图选择导出内容'
  return isERXinRecord(exportTargetRecord.value) ? '选择本次导出的报告内容' : '选择本次导出的内容范围'
}

function defaultExportDimensionForRecord(record) {
  if (isShuangxiARecord(record))
    return shuangxiExportDimensionOptions.find(item => item.recommended)?.value || shuangxiExportDimensionOptions[0]?.value || 'shuangxi_development_profile'
  if (isERXinRecord(record))
    return erxinExportDimensionOptions.find(item => item.recommended)?.value || erxinExportDimensionOptions[0]?.value || 'erxin_result'
  if (isAutismDevRecord(record))
    return defaultAutismDevReportSection()
  return defaultExportDimension
}

function normalizeSelectedExportDimension() {
  if (!activeExportDimensionOptions.value.some(item => item.value === selectedExportDimension.value))
    selectedExportDimension.value = defaultExportDimensionForRecord(exportTargetRecord.value)
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

function reportExportDimension(row, dimension = activeReportModule.value) {
  if (isAutismDevRecord(row)) {
    if (reportTab.value === 'interpretation')
      return 'interpretation'
    return activeAutismDevReportSection.value || defaultAutismDevReportSection()
  }
  if (!isERXinRecord(row) && reportTab.value === 'interpretation')
    return 'pep3_interpretation'
  return dimension
}

function reportExportTitle(row, dimension) {
  if (isShuangxiARecord(row))
    return exportDimensionTitle(dimension)
  if (isAutismDevRecord(row))
    return exportDimensionTitle(dimension)
  if (!isERXinRecord(row) && dimension === 'pep3_interpretation')
    return '报告解读'
  if (!isERXinRecord(row))
    return reportModuleTitle(dimension)
  return exportDimensionTitle(dimension)
}

function iepActionText(record) {
  return record?.iepPlanStatus === 'confirmed' ? '查看IEP' : '生成IEP'
}

function hasIepPlan(record) {
  return !!String(record?.iepPlanStatus || '').trim()
}

function canModifyAssessmentRecord(record) {
  return !!record?.id && !(hasIepPlan(record) && !isERXinRecord(record))
}

function assessmentRecordModifyTip(record) {
  if (!record?.id)
    return ''
  if (!canModifyAssessmentRecord(record))
    return '已生成IEP的评估记录，不支持直接修改。如需调整，请复用测评，提交一份新的测评记录。'
  return '修改当前测评记录，重新提交后会覆盖当前报告数据。'
}

function assessmentRecordReuseTip(record) {
  if (!record?.id)
    return ''
  return '复用当前测评的基本信息和作答结果，提交后生成新的正式记录，不覆盖原记录。'
}

function assessmentRecordConfirmTitle(record, mode = 'edit') {
  return mode === 'reuse' ? '确认复用测评？' : '确认修改测评？'
}

function assessmentRecordConfirmContent(record, mode = 'edit') {
  if (mode === 'reuse')
    return '将复制当前测评记录的基本信息和作答结果，打开测评页面后可继续调整；提交后会生成新的正式记录，不会覆盖原记录。'
  if (isERXinRecord(record))
    return '修改并重新提交后会覆盖当前儿心评估记录和报告数据，请确认后继续。'
  if (isAutismDevRecord(record))
    return '修改并重新提交后会覆盖当前孤独症儿童发展评估记录和报告数据，请确认后继续。'
  return '修改并重新提交后会覆盖当前评估记录和报告数据，请确认后继续。'
}

function confirmAssessmentRecordAction(record = currentReport.value?.record, mode = 'edit') {
  if (!record?.id)
    return
  if (isShuangxiARecord(record)) {
    messageService.warning(mode === 'reuse' ? '双溪量表记录暂不支持复用' : '双溪量表记录暂不支持修改')
    return
  }
  if (mode === 'edit' && !canModifyAssessmentRecord(record)) {
    messageService.warning('已生成IEP的评估记录不支持直接修改，请使用复用测评')
    return
  }
  Modal.confirm({
    title: assessmentRecordConfirmTitle(record, mode),
    content: assessmentRecordConfirmContent(record, mode),
    okText: mode === 'reuse' ? '确认复用' : '确认修改',
    cancelText: '取消',
    onOk: () => editAssessmentRecord(record, mode),
  })
}

function confirmReportExport(row = currentReport.value?.record, dimension = activeReportModule.value) {
  if (!row?.id || exportingId.value)
    return
  const exportDimension = reportExportDimension(row, dimension)
  if (isERXinRecord(row)) {
    openExportModal(row, reportTab.value === 'interpretation' ? 'erxin_interpretation' : 'erxin_result')
    return
  }
  if (isAutismDevRecord(row)) {
    openExportModal(row, reportTab.value === 'interpretation' ? 'interpretation' : activeAutismDevReportSection.value || defaultAutismDevReportSection())
    return
  }
  if (isShuangxiARecord(row)) {
    const dimension = 'shuangxi_development_profile'
    const content = `将导出「${row.studentName || '-'} / ${formatDate(row.assessmentDate)}」的${reportExportTitle(row, dimension)}PDF。`
    Modal.confirm({
      title: '确认导出评估报告？',
      content,
      okText: '确认导出',
      cancelText: '取消',
      onOk: () => exportReport(row, dimension),
    })
    return
  }
  const content = `将导出「${row.studentName || '-'} / ${formatDate(row.assessmentDate)}」的${reportExportTitle(row, exportDimension)}PDF。`
  Modal.confirm({
    title: '确认导出评估报告？',
    content,
    okText: '确认导出',
    cancelText: '取消',
    onOk: () => exportReport(row, exportDimension),
  })
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

function triggerBlobDownload(response, fallbackName, defaultContentType = 'application/octet-stream') {
  const contentType = String(response?.headers?.['content-type'] || response?.headers?.['Content-Type'] || defaultContentType)
  const blob = response?.data instanceof Blob ? response.data : new Blob([response?.data], { type: contentType })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = getDownloadFilename(response, fallbackName)
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  window.setTimeout(() => URL.revokeObjectURL(url), 60_000)
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
    const pageSize = Math.max(pagination.current * pagination.pageSize, pagination.pageSize)
    const request = {
      pageRequestModel: {
        pageIndex: 1,
        pageSize,
      },
      queryModel: {
        scaleCategory: queryModel.scaleCategory,
        studentId: queryModel.studentId,
        assessmentDateBegin: queryModel.assessmentDateBegin,
        assessmentDateEnd: queryModel.assessmentDateEnd,
      },
    }
    const recordRequests = [
      { label: 'PEP-3', source: 'PEP3', promise: pagePEP3AssessmentRecordsApi(request) },
      { label: '儿心量表', source: 'ERXIN', promise: pageERXinAssessmentRecordsApi(request) },
      { label: '孤独症儿童发展评估表', source: 'AUTISMDEV', promise: pageAutismDevAssessmentRecordsApi(request) },
      { label: '双溪量表A', source: 'SHUANGXI_A', promise: pageShuangxiAAssessmentRecordsApi(request) },
    ]
    const results = await Promise.allSettled(recordRequests.map(item => item.promise))
    const failedMessages = []
    const merged = []
    let total = 0
    results.forEach((result, index) => {
      const recordRequest = recordRequests[index]
      if (result.status === 'rejected') {
        failedMessages.push(`${recordRequest.label}加载失败`)
        return
      }
      const data = unwrap(result.value)
      total += Number(data?.total || 0)
      merged.push(...(data?.items || []).map(item => markRecordSource(item, recordRequest.source)))
    })
    if (failedMessages.length === recordRequests.length)
      throw new Error('获取评估记录失败')
    if (failedMessages.length)
      messageService.error(failedMessages.join('，'))
    merged.sort(compareRecordDesc)
    const start = (pagination.current - 1) * pagination.pageSize
    dataSource.value = merged.slice(start, start + pagination.pageSize)
    pagination.total = total
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
  resetInterpretationState()
  resetAutismDevReportState()
  reportTab.value = 'result'
  activeReportModule.value = defaultReportModule
  currentReport.value = {
    title: reportTitleForRecord(row),
    record: row,
  }
  reportModalOpen.value = true
  if (isAutismDevRecord(row)) {
    activeAutismDevReportSection.value = defaultAutismDevReportSection()
    void loadAutismDevResultAnalysis(row, { silent: true })
  }
  loadReportPdfPreview(row, defaultReportPreviewDimension(row))
}

function defaultReportPreviewDimension(row) {
  if (isShuangxiARecord(row))
    return 'shuangxi_development_profile'
  if (isAutismDevRecord(row))
    return defaultAutismDevReportSection()
  if (isERXinRecord(row))
    return 'erxin_result'
  return defaultReportModule
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
  if (isAutismDevRecord(row) && dimension === 'interpretation') {
    revokeReportPreviewUrl()
    if (!interpretationFetched.value && !interpretationLoading.value)
      loadSavedInterpretation()
    return
  }
  if (isAutismDevRecord(row) && dimension === 'resultAnalysis') {
    revokeReportPreviewUrl()
    void loadAutismDevResultAnalysis(row)
    return
  }
  const requestKey = reportPreviewRequestKey.value + 1
  reportPreviewRequestKey.value = requestKey
  resetReportPdfReady()
  previewLoading.value = true
  try {
    const autismDevSections = isAutismDevRecord(row) ? autismDevReportSectionsForDimension(dimension) : []
    const response = isShuangxiARecord(row)
      ? await downloadShuangxiADevelopmentProfilePdfApi(row.id)
      : isAutismDevRecord(row)
      ? await downloadAutismDevSelectedReportPdfApi(row.id, autismDevSections, autismDevSections.includes('resultAnalysis') ? autismDevAnalysisForExport() : null)
      : isERXinRecord(row)
      ? (dimension === 'erxin_interpretation'
          ? await downloadERXinAssessmentRecordReportInterpretationPdfApi(row.id)
          : await downloadERXinAssessmentRecordReportPdfApi(row.id))
      : await downloadPEP3AssessmentBookletPdfApi(row.id, dimension)
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
  if (currentReportIsERXin() || currentReportIsAutismDev() || currentReportIsShuangxiA())
    return
  reportTab.value = 'result'
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
  resetInterpretationState()
  resetAutismDevReportState()
}

function selectReportTab(tab) {
  reportTab.value = tab
  if (tab === 'interpretation' && !interpretationFetched.value && !interpretationLoading.value)
    loadSavedInterpretation()
}

async function loadSavedInterpretation(row = currentReport.value?.record) {
  if (!row?.id)
    return
  interpretationLoading.value = true
  interpretationGenerating.value = false
  interpretationFetched.value = true
  interpretationError.value = ''
  interpretationProgress.value = '正在读取已保存的报告解读...'
  interpretationStreamingText.value = ''
  try {
    const res = isERXinRecord(row)
      ? await getERXinAssessmentRecordReportInterpretationApi(row.id)
      : isAutismDevRecord(row)
      ? await getAutismDevAssessmentRecordReportInterpretationApi(row.id)
      : await getPEP3AssessmentRecordReportInterpretationApi(row.id)
    const data = normalizeInterpretation(unwrap(res))
    interpretation.value = data
    interpretationProgress.value = interpretationIsEmpty(data) ? '报告解读尚未生成' : '已读取保存的报告解读'
  }
  catch (error) {
    interpretationError.value = getErrorMessage(error, '报告解读读取失败')
  }
  finally {
    interpretationLoading.value = false
  }
}

function handleGenerateInterpretation() {
  if (!interpretationIsEmpty()) {
    Modal.confirm({
      title: '重新生成报告解读？',
      content: '重新生成会覆盖当前已保存的报告解读，确认继续吗？',
      okText: '重新生成',
      cancelText: '取消',
      onOk: () => {
        generateInterpretation(true)
      },
    })
    return
  }
  generateInterpretation(false)
}

async function generateInterpretation(regenerate = false) {
  const row = currentReport.value?.record
  if (!row?.id)
    return
  if (interpretationAbortController)
    interpretationAbortController.abort()
  const controller = new AbortController()
  interpretationAbortController = controller
  interpretationLoading.value = true
  interpretationGenerating.value = true
  interpretationFetched.value = true
  interpretationError.value = ''
  interpretationProgress.value = regenerate ? '正在重新生成报告解读...' : '正在生成报告解读...'
  interpretationStreamingText.value = ''
  interpretationProgressAnchored = false
  if (regenerate)
    interpretation.value = null
  scrollInterpretationProgressIntoView()
  try {
    const generateStreamApi = isERXinRecord(row)
      ? generateERXinAssessmentRecordReportInterpretationStreamApi
      : isAutismDevRecord(row)
      ? generateAutismDevAssessmentRecordReportInterpretationStreamApi
      : generatePEP3AssessmentRecordReportInterpretationStreamApi
    const data = await generateStreamApi(
      row.id,
      {
        onStatus: (message) => {
          interpretationProgress.value = message || 'AI 正在分析评估结果...'
          scrollInterpretationProgressIntoView()
        },
        onDelta: (text) => {
          if (text)
            interpretationStreamingText.value += text
          interpretationProgress.value = 'AI 正在生成报告解读...'
          nextTick(scrollInterpretationStreamToBottom)
        },
        onDone: (data) => {
          interpretation.value = normalizeInterpretation(data)
        },
      },
      { signal: controller.signal },
    )
    interpretation.value = normalizeInterpretation(data)
    interpretationProgress.value = '报告解读已生成'
    interpretationStreamingText.value = ''
  }
  catch (error) {
    if (error?.name !== 'AbortError')
      interpretationError.value = getErrorMessage(error, '报告解读生成失败')
  }
  finally {
    if (interpretationAbortController === controller) {
      interpretationLoading.value = false
      interpretationGenerating.value = false
      interpretationAbortController = null
    }
  }
}

async function loadAutismDevResultAnalysis(row = currentReport.value?.record, options = {}) {
  if (!row?.id || !isAutismDevRecord(row))
    return null
  const targetKey = recordActionKey(row)
  if (autismDevAnalysisFetched.value && autismDevAnalysisRecordKey.value === targetKey && !options.force)
    return autismDevAnalysis.value
  autismDevAnalysisRecordKey.value = targetKey
  autismDevAnalysisLoading.value = !options.silent
  autismDevAnalysisError.value = ''
  if (!options.keepProgress)
    autismDevAnalysisProgress.value = '正在读取评估结果分析...'
  try {
    const response = await getAutismDevResultAnalysisApi(row.id)
    if (autismDevAnalysisRecordKey.value !== targetKey)
      return null
    const data = normalizeAutismDevAnalysis(unwrap(response))
    autismDevAnalysis.value = data.rows.length ? data : emptyAutismDevAnalysis()
    autismDevAnalysisFetched.value = true
    autismDevAnalysisProgress.value = autismDevAnalysisIsEmpty(autismDevAnalysis.value) ? '评估结果分析尚未生成' : '已读取评估结果分析'
    return autismDevAnalysis.value
  }
  catch (error) {
    if (autismDevAnalysisRecordKey.value === targetKey) {
      autismDevAnalysisError.value = getErrorMessage(error, '评估结果分析读取失败')
      autismDevAnalysis.value = emptyAutismDevAnalysis()
      autismDevAnalysisFetched.value = true
    }
    return null
  }
  finally {
    if (autismDevAnalysisRecordKey.value === targetKey)
      autismDevAnalysisLoading.value = false
  }
}

async function loadAutismDevAnalysisForExport(row = exportTargetRecord.value) {
  if (!row?.id || !isAutismDevRecord(row))
    return
  autismDevExportAnalysisLoading.value = true
  try {
    await loadAutismDevResultAnalysis(row, { silent: true })
    setAutismDevExportSections(autismDevExportSections.value)
  }
  finally {
    autismDevExportAnalysisLoading.value = false
  }
}

async function generateAutismDevResultAnalysis() {
  const row = currentReport.value?.record
  if (!row?.id || !isAutismDevRecord(row) || autismDevAnalysisGenerating.value)
    return
  if (autismDevAnalysisAbortController)
    autismDevAnalysisAbortController.abort()
  const targetKey = recordActionKey(row)
  const controller = new AbortController()
  autismDevAnalysisAbortController = controller
  autismDevAnalysisRecordKey.value = targetKey
  autismDevAnalysisLoading.value = false
  autismDevAnalysisGenerating.value = true
  autismDevAnalysisFetched.value = true
  autismDevAnalysisError.value = ''
  autismDevAnalysisProgress.value = '正在生成评估结果分析...'
  autismDevAnalysisStreamText.value = ''
  autismDevAnalysis.value = emptyAutismDevAnalysis()
  try {
    const data = await generateAutismDevResultAnalysisStreamApi(
      row.id,
      {
        onStatus(message) {
          if (autismDevAnalysisRecordKey.value !== targetKey)
            return
          autismDevAnalysisProgress.value = message || '正在生成评估结果分析...'
        },
        onDelta(text) {
          if (autismDevAnalysisRecordKey.value !== targetKey)
            return
          if (text)
            autismDevAnalysisStreamText.value += text
          autismDevAnalysisProgress.value = 'AI正在生成评估结果分析...'
        },
        onDone(data) {
          if (autismDevAnalysisRecordKey.value !== targetKey)
            return
          autismDevAnalysis.value = normalizeAutismDevAnalysis(data)
        },
      },
      { signal: controller.signal },
    )
    if (autismDevAnalysisRecordKey.value !== targetKey)
      return
    autismDevAnalysis.value = normalizeAutismDevAnalysis(data)
    autismDevAnalysisProgress.value = '评估结果分析已生成'
    autismDevAnalysisStreamText.value = ''
    setAutismDevExportSections(autismDevExportSections.value)
  }
  catch (error) {
    if (error?.name !== 'AbortError' && autismDevAnalysisRecordKey.value === targetKey)
      autismDevAnalysisError.value = getErrorMessage(error, '评估结果分析生成失败')
  }
  finally {
    if (autismDevAnalysisAbortController === controller) {
      autismDevAnalysisGenerating.value = false
      autismDevAnalysisAbortController = null
    }
  }
}

async function exportAutismDevResultAnalysisWord() {
  const row = currentReport.value?.record
  if (!row?.id || !isAutismDevRecord(row) || autismDevAnalysisWordExporting.value)
    return
  if (autismDevAnalysisGenerating.value) {
    messageService.warning('评估结果分析生成中，请稍后导出')
    return
  }
  if (autismDevAnalysisLoading.value) {
    messageService.warning('评估结果分析读取中，请稍后')
    return
  }
  autismDevAnalysisWordExporting.value = true
  try {
    const targetKey = recordActionKey(row)
    if (!autismDevAnalysisFetched.value || autismDevAnalysisRecordKey.value !== targetKey)
      await loadAutismDevResultAnalysis(row, { silent: true })
    if (autismDevAnalysisIsEmpty()) {
      messageService.warning('请先生成评估结果分析后再导出')
      return
    }
    const response = await downloadAutismDevResultAnalysisWordApi(row.id, autismDevAnalysisForExport())
    triggerBlobDownload(
      response,
      `${row.studentName || '学员'}-孤独症儿童评估结果分析-${formatDate(row.assessmentDate)}.docx`,
      'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
    )
    messageService.success('导出成功')
  }
  catch (error) {
    messageService.error(await getDownloadErrorMessage(error, '导出Word失败'))
  }
  finally {
    autismDevAnalysisWordExporting.value = false
  }
}

function selectAutismDevReportSection(section) {
  if (!currentReportIsAutismDev())
    return
  if (activeAutismDevReportSection.value === section)
    return
  activeAutismDevReportSection.value = section
  if (section === 'interpretation') {
    reportTab.value = 'interpretation'
    revokeReportPreviewUrl()
    if (!interpretationFetched.value && !interpretationLoading.value)
      loadSavedInterpretation()
    return
  }
  reportTab.value = 'result'
  if (section === 'resultAnalysis') {
    revokeReportPreviewUrl()
    void loadAutismDevResultAnalysis()
    return
  }
  loadReportPdfPreview(currentReport.value?.record, section)
}

function openExportModal(row, dimension) {
  if (!row || exportingId.value)
    return
  exportTargetRecord.value = row
  if (isAutismDevRecord(row)) {
    setAutismDevExportSections(autismDevReportSectionsForDimension(dimension))
    selectedExportDimension.value = autismDevExportSections.value[0] || defaultAutismDevReportSection()
    exportModalOpen.value = true
    if (autismDevExportSections.value.includes('resultAnalysis'))
      void loadAutismDevAnalysisForExport(row)
    if (autismDevExportSections.value.includes('interpretation') && !interpretationFetched.value && !interpretationLoading.value)
      void loadSavedInterpretation(row)
    return
  }
  selectedExportDimension.value = dimension || defaultExportDimensionForRecord(row)
  normalizeSelectedExportDimension()
  exportModalOpen.value = true
}

function openIepModal(row) {
  if (!row)
    return
  iepTargetRecord.value = row
  iepModalOpen.value = true
}

function openConfigModal(row) {
  if (!row)
    return
  configTargetRecord.value = row
  configModalOpen.value = true
}

function editAssessmentRecord(row = currentReport.value?.record, mode = 'edit') {
  if (!row?.id)
    return
  if (isShuangxiARecord(row)) {
    messageService.warning(mode === 'reuse' ? '双溪量表记录暂不支持复用' : '双溪量表记录暂不支持修改')
    return
  }
  const recordMode = mode === 'reuse' ? 'reuse' : 'edit'
  const path = isAutismDevRecord(row)
    ? '/teacherCenter/autismdev-assessment-workbench'
    : isERXinRecord(row)
      ? '/teacherCenter/erxin-assessment-workbench'
      : '/teacherCenter/scale-assessment-workbench'
  closeReportModal()
  void router.push({
    path,
    query: {
      recordId: row.id,
      recordMode,
      scaleName: row.assessmentName || (isAutismDevRecord(row) ? '孤独症儿童发展评估表' : isERXinRecord(row) ? '儿心量表-II' : 'PEP-3'),
      scaleCode: row.assessmentCode || (isAutismDevRecord(row) ? 'AUTISMDEV' : isERXinRecord(row) ? 'ERXIN2' : 'PEP3'),
      childId: row.studentId,
      childName: row.studentName,
      childAge: formatAge(row),
      childBirthDate: formatDate(row.birthDate),
      assessmentDate: formatDate(row.assessmentDate),
      examinerName: row.examinerName,
    },
  })
}

function closeExportModal() {
  if (exportingId.value)
    return
  exportModalOpen.value = false
}

async function downloadERXinExportPdf(recordId, dimension) {
  if (dimension === 'erxin_interpretation')
    return downloadERXinAssessmentRecordReportInterpretationPdfApi(recordId)
  if (dimension === 'erxin_combined')
    return downloadERXinAssessmentRecordReportCombinedPdfApi(recordId)
  return downloadERXinAssessmentRecordReportPdfApi(recordId)
}

async function downloadPEP3ExportPdf(recordId, dimension) {
  if (dimension === 'pep3_interpretation')
    return downloadPEP3AssessmentRecordReportInterpretationPdfApi(recordId)
  return downloadPEP3AssessmentBookletPdfApi(recordId, dimension)
}

async function downloadShuangxiAExportPdf(recordId) {
  return downloadShuangxiADevelopmentProfilePdfApi(recordId)
}

function autismDevReportSectionsForDimension(dimension) {
  if (Array.isArray(dimension))
    return normalizeAutismDevExportSections(dimension)
  if (autismDevExportDimensionOptions.some(item => item.value === dimension))
    return [dimension]
  if (dimension === 'all' || dimension === 'autismdev_report')
    return normalizeAutismDevExportSections(autismDevExportDimensionOptions.map(item => item.value))
  return [defaultAutismDevReportSection()]
}

async function exportReport(row = exportTargetRecord.value, dimension = selectedExportDimension.value) {
  if (!row)
    return
  const autismDevSections = isAutismDevRecord(row)
    ? normalizeAutismDevExportSections(autismDevExportSections.value.length ? autismDevExportSections.value : autismDevReportSectionsForDimension(dimension))
    : []
  if (isAutismDevRecord(row)) {
    if (!autismDevSections.length) {
      messageService.warning('请选择导出内容')
      return
    }
    if (autismDevSections.includes('interpretation') && (interpretationLoading.value || interpretationGenerating.value)) {
      messageService.warning('报告解读生成或读取中，请稍后')
      return
    }
    if (autismDevSections.includes('resultAnalysis') && (autismDevExportAnalysisLoading.value || autismDevAnalysisLoading.value)) {
      messageService.warning('评估结果分析读取中，请稍后')
      return
    }
    if (autismDevSections.includes('resultAnalysis') && autismDevAnalysisIsEmpty()) {
      messageService.warning('请先生成评估结果分析后再导出')
      return
    }
  }
  exportingId.value = recordActionKey(row)
  try {
    const response = isShuangxiARecord(row)
      ? await downloadShuangxiAExportPdf(row.id)
      : isAutismDevRecord(row)
      ? await downloadAutismDevSelectedReportPdfApi(row.id, autismDevSections, autismDevSections.includes('resultAnalysis') ? autismDevAnalysisForExport() : null)
      : isERXinRecord(row)
      ? await downloadERXinExportPdf(row.id, dimension)
      : await downloadPEP3ExportPdf(row.id, dimension)
    const url = URL.createObjectURL(new Blob([response.data], { type: 'application/pdf' }))
    const link = document.createElement('a')
    link.href = url
    const fallbackName = isShuangxiARecord(row)
      ? `${row.studentName || '学员'}-双溪综合发展侧面图-${formatDate(row.assessmentDate)}.pdf`
      : isAutismDevRecord(row)
      ? `${row.studentName || '学员'}-孤独症儿童发展评估报告-${autismDevSections.map(autismDevReportSectionTitle).join('+')}-${formatDate(row.assessmentDate)}.pdf`
      : isERXinRecord(row)
      ? `${row.studentName || '学员'}-${exportDimensionTitle(dimension)}-${formatDate(row.assessmentDate)}.pdf`
      : `${row.studentName || '学员'}-${row.assessmentName || '评估记录'}-${exportDimensionTitle(dimension)}-${formatDate(row.assessmentDate)}.pdf`
    link.download = getDownloadFilename(response, fallbackName)
    link.click()
    window.setTimeout(() => URL.revokeObjectURL(url), 60_000)
    if (exportModalOpen.value)
      exportModalOpen.value = false
  }
  catch (error) {
    messageService.error(await getDownloadErrorMessage(error, '导出评估记录失败'))
  }
  finally {
    exportingId.value = undefined
  }
}

async function deleteRecord(row) {
  deletingId.value = recordActionKey(row)
  try {
    if (isShuangxiARecord(row))
      await deleteShuangxiAAssessmentRecordApi(row.id)
    else if (isAutismDevRecord(row))
      await deleteAutismDevAssessmentRecordApi(row.id)
    else if (isERXinRecord(row))
      await deleteERXinAssessmentRecordApi(row.id)
    else
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
  resetInterpretationState()
  resetAutismDevReportState()
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
            :row-key="recordActionKey"
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
                  default-active-key="6"
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
                <a-tooltip :title="record.examinerName || '-'">
                  <span class="single-line">{{ record.examinerName || '-' }}</span>
                </a-tooltip>
              </template>
              <template v-else-if="column.key === 'createdTime'">
                {{ formatDateTime(record.createdTime) }}
              </template>
              <template v-else-if="column.key === 'action'">
                <a-space :size="8" class="action-links">
                  <a :class="{ disabled: previewLoading }" @click="viewReport(record)">查看</a>
                  <a @click="openConfigModal(record)">配置</a>
                  <a-popconfirm title="确认删除这条评估记录？" ok-text="删除" cancel-text="取消" @confirm="deleteRecord(record)">
                    <a :class="{ disabled: deletingId === recordActionKey(record) }">删除</a>
                  </a-popconfirm>
                  <a :class="{ disabled: exportingId === recordActionKey(record) }" @click="openExportModal(record)">导出</a>
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
          <small>{{ reportModalHint() }}</small>
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
            <div v-if="!isERXinRecord(currentReport.record) && !isAutismDevRecord(currentReport.record) && !isShuangxiARecord(currentReport.record)" class="report-inline-summary">
              <strong>{{ reportTab === 'interpretation' ? '报告解读' : reportModulePages(activeReportModule) }}</strong>
              <span>{{ reportTab === 'interpretation' ? '查看或生成PEP-3报告解读。' : reportModuleDesc(activeReportModule) }}</span>
              <a-button
                v-if="reportTab === 'interpretation'"
                type="primary"
                size="small"
                :loading="interpretationGenerating"
                :disabled="interpretationLoading && !interpretationGenerating"
                @click="handleGenerateInterpretation"
              >
                {{ interpretationGenerating ? '生成中' : (interpretationIsEmpty() ? '生成解读' : '重新生成解读') }}
              </a-button>
            </div>
          </div>
          <div class="report-head__actions">
            <a-button
              type="primary"
              size="small"
              class="report-export-btn"
              :loading="exportingId === recordActionKey(currentReport.record)"
              @click="confirmReportExport(currentReport.record, activeReportModule)"
            >
              导出
            </a-button>
            <a-tooltip :title="assessmentRecordModifyTip(currentReport.record)" placement="top">
              <a-button
                size="small"
                class="report-edit-btn"
                :disabled="exportingId === recordActionKey(currentReport.record) || !canModifyAssessmentRecord(currentReport.record)"
                @click="confirmAssessmentRecordAction(currentReport.record, 'edit')"
              >
                修改
              </a-button>
            </a-tooltip>
            <a-tooltip :title="assessmentRecordReuseTip(currentReport.record)" placement="top">
              <a-button
                size="small"
                class="report-edit-btn"
                :disabled="exportingId === recordActionKey(currentReport.record)"
                @click="confirmAssessmentRecordAction(currentReport.record, 'reuse')"
              >
                复用
              </a-button>
            </a-tooltip>
          </div>
        </div>
        <div v-if="!isERXinRecord(currentReport.record) && !isAutismDevRecord(currentReport.record) && !isShuangxiARecord(currentReport.record)" class="report-module-area pep3-report-tabs">
          <div class="report-module-grid">
            <button
              v-for="option in reportModuleOptions"
              :key="option.value"
              type="button"
              class="report-module-chip"
              :class="{ 'report-module-chip--active': reportTab === 'result' && activeReportModule === option.value }"
              :title="option.title"
              @click="selectReportModule(option.value)"
            >
              <span class="report-module-chip__dot" />
              <span class="report-module-chip__text">{{ reportModuleShortTitle(option.value) }}</span>
              <span v-if="option.recommended" class="report-module-chip__tag">推荐</span>
            </button>
            <button
              type="button"
              class="report-module-chip"
              :class="{ 'report-module-chip--active': reportTab === 'interpretation' }"
              @click="selectReportTab('interpretation')"
            >
              <span class="report-module-chip__dot" />
              <span class="report-module-chip__text">报告解读</span>
            </button>
          </div>
        </div>
        <div v-else-if="isShuangxiARecord(currentReport.record)" class="report-module-area erxin-report-tabs">
          <div class="report-module-grid">
            <button
              type="button"
              class="report-module-chip report-module-chip--active"
            >
              <span class="report-module-chip__dot" />
              <span class="report-module-chip__text">综合发展侧面图</span>
            </button>
          </div>
          <div class="report-module-summary erxin-report-tabs__summary">
            <strong>PDF报告</strong>
            <span>查看双溪课程评量表A综合发展侧面图。</span>
          </div>
        </div>
        <div v-else-if="isAutismDevRecord(currentReport.record)" class="report-module-area erxin-report-tabs">
          <div class="report-module-grid autismdev-report-tabs">
            <button
              v-for="option in autismDevExportDimensionOptions"
              :key="option.value"
              type="button"
              class="report-module-chip"
              :class="{ 'report-module-chip--active': activeAutismDevReportSection === option.value }"
              :title="option.title"
              @click="selectAutismDevReportSection(option.value)"
            >
              <span class="report-module-chip__dot" />
              <span class="report-module-chip__text">{{ option.title }}</span>
            </button>
          </div>
          <div
            v-if="activeAutismDevReportSection === 'resultAnalysis'"
            class="report-module-summary report-module-summary--actions-only erxin-report-tabs__summary"
          >
            <div class="report-module-summary__actions">
              <a-button
                v-if="!autismDevAnalysisIsEmpty()"
                size="small"
                :loading="autismDevAnalysisWordExporting"
                :disabled="autismDevAnalysisLoading || autismDevAnalysisGenerating"
                @click="exportAutismDevResultAnalysisWord"
              >
                导出Word
              </a-button>
              <a-button
                type="primary"
                size="small"
                :loading="autismDevAnalysisGenerating"
                :disabled="autismDevAnalysisLoading || autismDevAnalysisWordExporting"
                @click="generateAutismDevResultAnalysis"
              >
                {{ autismDevAnalysisGenerating ? '生成中' : (autismDevAnalysisIsEmpty() ? 'AI生成' : '重新生成') }}
              </a-button>
            </div>
          </div>
          <div
            v-else-if="activeAutismDevReportSection === 'interpretation'"
            class="report-module-summary report-module-summary--actions-only erxin-report-tabs__summary"
          >
            <div class="report-module-summary__actions">
              <a-button
                type="primary"
                size="small"
                :loading="interpretationGenerating"
                :disabled="interpretationLoading && !interpretationGenerating"
                @click="handleGenerateInterpretation"
              >
                {{ interpretationGenerating ? '生成中' : (interpretationIsEmpty() ? '生成解读' : '重新生成解读') }}
              </a-button>
            </div>
          </div>
        </div>
        <div v-else class="report-module-area erxin-report-tabs">
          <div class="report-module-grid">
            <button
              type="button"
              class="report-module-chip"
              :class="{ 'report-module-chip--active': reportTab === 'result' }"
              @click="selectReportTab('result')"
            >
              <span class="report-module-chip__dot" />
              <span class="report-module-chip__text">评估结果记录</span>
            </button>
            <button
              type="button"
              class="report-module-chip"
              :class="{ 'report-module-chip--active': reportTab === 'interpretation' }"
              @click="selectReportTab('interpretation')"
            >
              <span class="report-module-chip__dot" />
              <span class="report-module-chip__text">报告解读</span>
            </button>
          </div>
          <div class="report-module-summary erxin-report-tabs__summary">
            <strong>{{ reportTab === 'interpretation' ? '报告解读' : 'PDF报告' }}</strong>
            <span>{{ reportTab === 'interpretation' ? '查看或生成儿心量表报告解读。' : '查看儿心量表评估结果记录。' }}</span>
            <a-button
              v-if="reportTab === 'interpretation'"
              type="primary"
              size="small"
              :loading="interpretationGenerating"
              :disabled="interpretationLoading && !interpretationGenerating"
              @click="handleGenerateInterpretation"
            >
              {{ interpretationGenerating ? '生成中' : (interpretationIsEmpty() ? '生成解读' : '重新生成解读') }}
            </a-button>
          </div>
        </div>

        <div class="report-module-content">
          <div v-if="currentReportIsAutismDev() && activeAutismDevReportSection === 'resultAnalysis'" class="autismdev-analysis-shell">
            <div v-if="autismDevAnalysisLoading" class="autismdev-analysis-state">
              <a-spin size="small" />
              <strong>评估结果分析读取中</strong>
              <span>{{ autismDevAnalysisProgress || '正在读取已保存的评估结果分析...' }}</span>
            </div>
            <div v-else-if="autismDevAnalysisGenerating" class="autismdev-analysis-stream">
              <div class="autismdev-analysis-stream__head">
                <span>AI</span>
                <div class="autismdev-analysis-stream__title">
                  <strong>正在生成 {{ currentReport?.record?.studentName || '学员' }} 的评估结果分析</strong>
                  <div class="autismdev-analysis-stream__progress-line">
                    <small>{{ autismDevAnalysisProgress || 'AI正在生成评估结果分析...' }}</small>
                    <div class="autismdev-analysis-progress">
                      <i :style="{ width: `${autismDevAnalysisStreamProgressPercent}%` }" />
                    </div>
                    <em>{{ autismDevAnalysisStreamProgressPercent }}%</em>
                  </div>
                </div>
              </div>
              <div class="autismdev-analysis-content autismdev-analysis-content--streaming">
                <div class="autismdev-analysis-doc-title">
                  孤独症儿童评估结果分析表
                </div>
                <div class="autismdev-analysis-meta">
                  <div>
                    <span>儿童姓名：</span>
                    <em>{{ currentReport?.record?.studentName || '-' }}</em>
                  </div>
                  <div>
                    <span>评估者：</span>
                    <em>{{ currentReport?.record?.examinerName || '-' }}</em>
                  </div>
                  <div>
                    <span>评估时间：</span>
                    <em>{{ formatDate(currentReport?.record?.assessmentDate) }}</em>
                  </div>
                </div>
                <div class="autismdev-analysis-table-wrap">
                  <table class="autismdev-analysis-table autismdev-analysis-table--streaming">
                    <colgroup>
                      <col class="autismdev-analysis-table__col-domain">
                      <col class="autismdev-analysis-table__col-status">
                      <col class="autismdev-analysis-table__col-strength">
                      <col class="autismdev-analysis-table__col-target">
                    </colgroup>
                    <thead>
                      <tr>
                        <th>领&nbsp;&nbsp;&nbsp;域</th>
                        <th>能力现状描述</th>
                        <th>优劣分析</th>
                        <th>训练目标</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr
                        v-for="(row, index) in autismDevAnalysisStreamRows"
                        :key="`stream-${row.domain}-${index}`"
                      >
                        <td>{{ row.domain || '' }}</td>
                        <td>{{ row.status || '' }}</td>
                        <td>{{ autismDevStrengthWeaknessText(row) || '' }}</td>
                        <td>{{ row.targets || '' }}</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
            <div v-else-if="autismDevAnalysisError" class="autismdev-analysis-state">
              <a-empty
                description="评估结果分析加载失败"
                :image="simpleEmptyImage"
                :image-style="{ height: '48px' }"
              />
              <p>{{ autismDevAnalysisError }}</p>
              <a-button size="small" type="primary" @click="generateAutismDevResultAnalysis">
                重新生成
              </a-button>
            </div>
            <div v-else-if="autismDevAnalysisIsEmpty()" class="autismdev-analysis-state">
              <a-empty
                description="评估结果分析尚未生成"
                :image="simpleEmptyImage"
                :image-style="{ height: '48px' }"
              />
              <p>点击“AI生成”后，系统会基于当前评估结果生成并保存分析表。</p>
              <a-button size="small" type="primary" @click="generateAutismDevResultAnalysis">
                AI生成
              </a-button>
            </div>
            <div v-else class="autismdev-analysis-content">
              <div class="autismdev-analysis-doc-title">
                {{ autismDevAnalysis?.title || '孤独症儿童评估结果分析表' }}
              </div>
              <div class="autismdev-analysis-meta">
                <div>
                  <span>儿童姓名：</span>
                  <em>{{ currentReport?.record?.studentName || '-' }}</em>
                </div>
                <div>
                  <span>评估者：</span>
                  <em>{{ currentReport?.record?.examinerName || '-' }}</em>
                </div>
                <div>
                  <span>评估时间：</span>
                  <em>{{ formatDate(currentReport?.record?.assessmentDate) }}</em>
                </div>
              </div>
              <div class="autismdev-analysis-table-wrap">
                <table class="autismdev-analysis-table">
                  <colgroup>
                    <col class="autismdev-analysis-table__col-domain">
                    <col class="autismdev-analysis-table__col-status">
                    <col class="autismdev-analysis-table__col-strength">
                    <col class="autismdev-analysis-table__col-target">
                  </colgroup>
                  <thead>
                    <tr>
                      <th>领&nbsp;&nbsp;&nbsp;域</th>
                      <th>能力现状描述</th>
                      <th>优劣分析</th>
                      <th>训练目标</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr
                      v-for="(row, index) in autismDevAnalysis.rows"
                      :key="`${row.domain}-${index}`"
                    >
                      <td>{{ row.domain || '' }}</td>
                      <td>{{ row.status || '' }}</td>
                      <td>{{ autismDevStrengthWeaknessText(row) || '' }}</td>
                      <td>{{ row.targets || '' }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
          <div v-else-if="reportTab === 'result'" class="report-pdf-shell">
            <iframe
              v-if="reportPreviewUrl"
              :key="reportPreviewUrl"
              class="report-pdf-frame"
              :class="{ 'report-pdf-frame--ready': reportPdfReady }"
              :src="`${reportPreviewUrl}#toolbar=0&navpanes=0`"
              :title="reportFrameTitle()"
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
          <div v-else ref="interpretationShellRef" class="erxin-interpretation-shell">
            <div v-if="interpretationLoading && interpretationGenerating" ref="interpretationProgressRef" class="erxin-interpretation-progress">
              <div class="erxin-interpretation-progress__header">
                <span class="erxin-interpretation-progress__icon">AI</span>
                <div class="erxin-interpretation-progress__title">
                  <strong>AI 正在生成报告解读</strong>
                  <span>{{ interpretationProgress || reportInterpretationGeneratingHint() }}</span>
                </div>
                <em class="erxin-streaming-indicator">流式生成</em>
              </div>
              <div ref="interpretationStreamRef" class="erxin-interpretation-stream">
                <div v-if="streamingPreviewIsEmpty()" class="erxin-interpretation-stream__empty">
                  正在建立报告结构，稍后开始输出解读内容...
                </div>
                <template v-else>
                  <section v-if="streamingInterpretationPreview.summary" class="erxin-interpretation-section erxin-interpretation-section--streaming">
                    <h4>综合解读</h4>
                    <p>{{ streamingInterpretationPreview.summary }}</p>
                  </section>
                  <section v-if="streamingInterpretationPreview.domainAnalysis.length" class="erxin-interpretation-section erxin-interpretation-section--streaming">
                    <h4>{{ reportInterpretationDomainSectionTitle() }}</h4>
                    <ol>
                      <li v-for="(item, index) in streamingInterpretationPreview.domainAnalysis" :key="`stream-domain-${index}`">
                        {{ item }}
                      </li>
                    </ol>
                  </section>
                  <section v-if="streamingInterpretationPreview.suggestions.length" class="erxin-interpretation-section erxin-interpretation-section--streaming">
                    <h4>发展建议</h4>
                    <ol>
                      <li v-for="(item, index) in streamingInterpretationPreview.suggestions" :key="`stream-suggestion-${index}`">
                        {{ item }}
                      </li>
                    </ol>
                  </section>
                  <section v-if="streamingInterpretationPreview.notes.length" class="erxin-interpretation-section erxin-interpretation-section--streaming">
                    <h4>注意事项</h4>
                    <ol>
                      <li v-for="(item, index) in streamingInterpretationPreview.notes" :key="`stream-note-${index}`">
                        {{ item }}
                      </li>
                    </ol>
                  </section>
                </template>
              </div>
            </div>
            <div v-else-if="interpretationLoading" class="erxin-interpretation-read-loading">
              <a-spin size="small" />
              <strong>报告解读读取中</strong>
              <span>{{ interpretationProgress || '正在读取已保存的报告解读...' }}</span>
            </div>
            <div v-else-if="!interpretationLoading && interpretationError" class="erxin-interpretation-state">
              <a-empty
                description="报告解读加载失败"
                :image="simpleEmptyImage"
                :image-style="{ height: '48px' }"
              />
              <p>{{ interpretationError }}</p>
              <a-button size="small" type="primary" @click="generateInterpretation(true)">
                重新生成
              </a-button>
            </div>
            <div v-else-if="!interpretationLoading && interpretationIsEmpty()" class="erxin-interpretation-state">
              <a-empty
                description="报告解读尚未生成"
                :image="simpleEmptyImage"
                :image-style="{ height: '48px' }"
              />
              <p>点击“生成解读”后，AI 会基于当前评估结果生成并保存。</p>
              <a-button size="small" type="primary" @click="generateInterpretation(false)">
                生成解读
              </a-button>
            </div>
            <div v-else-if="!interpretationLoading" class="erxin-interpretation-content">
              <div class="erxin-interpretation-title">
                <strong>{{ interpretation.title || reportInterpretationDefaultTitle() }}</strong>
                <span v-if="interpretation.generatedAt || interpretation.generatedBy">
                  {{ interpretation.generatedBy || 'AI' }} {{ interpretation.generatedAt ? `· ${formatDateTime(interpretation.generatedAt)}` : '' }}
                </span>
              </div>
              <section class="erxin-interpretation-section">
                <h4>综合解读</h4>
                <p>{{ interpretation.summary || '-' }}</p>
              </section>
              <section class="erxin-interpretation-section">
                <h4>{{ reportInterpretationDomainSectionTitle() }}</h4>
                <ol>
                  <li v-for="(item, index) in interpretationSectionItems(interpretation, 'domainAnalysis')" :key="`domain-${index}`">
                    {{ item }}
                  </li>
                </ol>
              </section>
              <section class="erxin-interpretation-section">
                <h4>发展建议</h4>
                <ol>
                  <li v-for="(item, index) in interpretationSectionItems(interpretation, 'suggestions')" :key="`suggestion-${index}`">
                    {{ item }}
                  </li>
                </ol>
              </section>
              <section v-if="interpretationSectionItems(interpretation, 'notes').length" class="erxin-interpretation-section">
                <h4>注意事项</h4>
                <ol>
                  <li v-for="(item, index) in interpretationSectionItems(interpretation, 'notes')" :key="`note-${index}`">
                    {{ item }}
                  </li>
                </ol>
              </section>
            </div>
          </div>
        </div>
      </div>
    </a-modal>

    <a-modal
      v-model:open="exportModalOpen"
      :width="exportModalWidth"
      :centered="true"
      :footer="null"
      wrap-class-name="pep3-export-dimension-modal"
      :mask-closable="!exportingId"
      @cancel="closeExportModal"
    >
      <template #title>
        <div class="export-modal-title">
          <span>{{ exportModalTitle() }}</span>
          <small>{{ exportModalHint() }}</small>
        </div>
      </template>
      <div
        class="export-dimension"
        :class="{
          'export-dimension--erxin': isERXinRecord(exportTargetRecord),
          'export-dimension--autismdev': isAutismDevRecord(exportTargetRecord),
        }"
      >
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
            <strong>{{ activeExportDimensionOptions.length }} 项</strong>
          </div>
        </div>
        <div v-if="isAutismDevRecord(exportTargetRecord)" class="autismdev-export-card-list">
          <div class="autismdev-export-toolbar">
            <span>按 pad 端报告分段选择导出内容</span>
            <button type="button" :disabled="!!exportingId" @click="toggleAllAutismDevExportSections">
              {{ autismDevExportAllSelected ? '清空' : '全选' }}
            </button>
          </div>
          <button
            v-for="option in autismDevExportDimensionOptions"
            :key="option.value"
            type="button"
            class="autismdev-export-row"
            :class="{
              'autismdev-export-row--active': autismDevExportSections.includes(option.value),
              'autismdev-export-row--disabled': !isAutismDevExportSectionEnabled(option.value),
            }"
            :disabled="!!exportingId || !isAutismDevExportSectionEnabled(option.value)"
            @click="toggleAutismDevExportSection(option.value)"
          >
            <span class="autismdev-export-row__check" />
            <span class="autismdev-export-row__body">
              <span class="autismdev-export-row__head">
                <strong>{{ option.title }}</strong>
                <em v-if="option.recommended">默认</em>
                <em v-if="autismDevExportSectionStatus(option.value)" class="is-muted">{{ autismDevExportSectionStatus(option.value) }}</em>
              </span>
              <span class="autismdev-export-row__desc">{{ option.desc }}</span>
            </span>
          </button>
        </div>
        <div v-else-if="isERXinRecord(exportTargetRecord)" class="erxin-export-card-list">
          <button
            v-for="option in activeExportDimensionOptions"
            :key="option.value"
            type="button"
            class="erxin-export-card"
            :class="{ 'erxin-export-card--active': selectedExportDimension === option.value }"
            :aria-pressed="selectedExportDimension === option.value"
            :disabled="!!exportingId"
            @click="selectedExportDimension = option.value"
          >
            <span class="erxin-export-card__check" />
            <span class="erxin-export-card__body">
              <span class="erxin-export-card__head">
                <strong>{{ option.title }}</strong>
                <em v-if="option.recommended">推荐</em>
              </span>
              <span class="erxin-export-card__desc">{{ option.desc }}</span>
              <span class="erxin-export-card__pages">{{ option.pages }}</span>
            </span>
          </button>
        </div>
        <div v-else class="export-dimension__chooser">
          <div class="export-dimension__matrix">
            <button
              v-for="option in activeExportDimensionOptions"
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
          <div v-if="isAutismDevRecord(exportTargetRecord)" class="export-dimension__selection">
            <span>将导出</span>
            <strong>{{ autismDevExportSections.length }} 项内容</strong>
            <em>{{ autismDevExportSections.map(autismDevReportSectionTitle).join('、') || '未选择' }}</em>
          </div>
          <div v-else class="export-dimension__selection">
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

    <assessment-record-config-modal
      v-model:open="configModalOpen"
      :record="configTargetRecord"
      @saved="fetchRecords"
    />

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
  display: inline-block;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  vertical-align: bottom;
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
  align-items: center;
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

.report-inline-summary {
  display: flex;
  align-items: center;
  flex: 1 1 auto;
  gap: 8px;
  min-width: 120px;
  padding-left: 12px;
  overflow: hidden;
  border-left: 1px solid #e6edf6;

  strong {
    flex: 0 0 auto;
    color: var(--pro-ant-color-primary);
    font-size: 13px;
    font-weight: 600;
    line-height: 20px;
    white-space: nowrap;
  }

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

  :deep(.ant-btn) {
    flex: 0 0 auto;
    height: 26px;
    padding: 0 10px;
    font-size: 12px;
  }
}

.report-head__actions {
  display: flex;
  flex: 0 0 auto;
  gap: 8px;
  align-items: center;
}

.report-export-btn {
  min-width: 56px;
  height: 28px;
  padding: 0 12px;
}

.report-edit-btn {
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

.report-module-summary__actions {
  flex: 0 0 auto;
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.report-module-summary--actions-only {
  flex: 0 0 auto;
  justify-content: flex-end;
  padding-left: 0;
  margin-left: auto;
  border-left: 0;
}

.pep3-report-tabs {
  align-items: center;
}

.pep3-report-tabs .report-module-grid {
  width: 100%;
}

.erxin-report-tabs__summary {
  :deep(.ant-btn) {
    flex: 0 0 auto;
    height: 28px;
    padding: 0 12px;
    font-size: 12px;
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

.erxin-interpretation-shell {
  min-height: 620px;
  padding: 18px;
  overflow-y: auto;
  background: #fbfdff;
  border: 1px solid #edf1f6;
  border-radius: 8px;
  scrollbar-color: #c6d1df transparent;
  scrollbar-width: thin;

  &::-webkit-scrollbar {
    width: 8px;
  }

  &::-webkit-scrollbar-track {
    background: transparent;
  }

  &::-webkit-scrollbar-thumb {
    background: #c6d1df;
    border-radius: 999px;
  }

  &::-webkit-scrollbar-thumb:hover {
    background: #aebccc;
  }
}

.erxin-interpretation-progress {
  display: flex;
  flex-direction: column;
  height: min(66vh, 620px);
  min-height: 540px;
  padding: 16px 18px 18px;
  background: #fff;
  border: 1px solid #e6edf6;
  border-radius: 10px;
  box-shadow: 0 8px 22px rgba(15, 23, 42, 0.04);
}

.erxin-interpretation-progress__header {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 0 0 auto;
  padding-bottom: 12px;
  margin-bottom: 12px;
  color: #64748b;
  border-bottom: 1px solid #edf1f6;
}

.erxin-interpretation-progress__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  width: 32px;
  height: 32px;
  color: var(--pro-ant-color-primary);
  font-size: 12px;
  font-weight: 700;
  line-height: 1;
  background: #eef6ff;
  border: 1px solid #d8eaff;
  border-radius: 10px;
}

.erxin-interpretation-progress__title {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  gap: 2px;
  min-width: 0;

  strong {
    overflow: hidden;
    color: #1f2937;
    font-size: 15px;
    font-weight: 600;
    line-height: 22px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  span {
    min-width: 0;
    overflow: hidden;
    color: #7a8494;
    font-size: 12px;
    font-weight: 500;
    line-height: 18px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.erxin-streaming-indicator {
  flex: 0 0 auto;
  padding: 0 9px;
  color: var(--pro-ant-color-primary);
  font-size: 11px;
  font-style: normal;
  font-weight: 600;
  line-height: 24px;
  background: #eef6ff;
  border: 1px solid #d8eaff;
  border-radius: 999px;
  animation: erxin-stream-pulse 0.9s ease-in-out infinite;
}

.erxin-interpretation-read-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  min-height: 540px;
  color: #7a8494;
  font-size: 13px;

  strong {
    color: #1f2937;
    font-size: 13px;
    font-weight: 600;
    line-height: 22px;
  }

  span {
    color: #7a8494;
    font-size: 12px;
    line-height: 18px;
  }
}

.erxin-interpretation-stream {
  display: flex;
  flex-direction: column;
  gap: 14px;
  flex: 1 1 auto;
  min-height: 0;
  padding: 2px 2px 4px;
  overflow-y: auto;
  background: transparent;
  scrollbar-color: #c6d1df transparent;
  scrollbar-width: thin;

  &::-webkit-scrollbar {
    width: 8px;
  }

  &::-webkit-scrollbar-track {
    background: transparent;
  }

  &::-webkit-scrollbar-thumb {
    background: #c6d1df;
    border-radius: 999px;
  }
}

.erxin-interpretation-stream__empty {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 430px;
  color: #8a94a6;
  font-size: 13px;
  font-weight: 600;
  line-height: 22px;
  background: #f8fafc;
  border: 1px dashed #dbe7f5;
  border-radius: 8px;
}

.erxin-interpretation-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-start;
  min-height: 360px;
  padding-top: 88px;
  color: #7a8494;

  :deep(.ant-empty) {
    margin-bottom: 8px;
  }

  :deep(.ant-empty-image) {
    height: 48px;
    margin-bottom: 8px;
  }

  :deep(.ant-empty-description) {
    color: #b7beca;
    font-size: 14px;
    line-height: 22px;
  }

  p {
    margin: 0 0 12px;
    color: #7a8494;
    font-size: 13px;
    line-height: 20px;
  }
}

.erxin-interpretation-content {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 0;
}

.erxin-interpretation-title {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 14px;
  background: #f8fafc;
  border: 1px solid #edf1f6;
  border-radius: 8px;

  strong {
    min-width: 0;
    overflow: hidden;
    color: #1f2937;
    font-size: 15px;
    font-weight: 600;
    line-height: 22px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  span {
    flex: 0 0 auto;
    color: #8a94a6;
    font-size: 12px;
    line-height: 18px;
  }
}

.erxin-interpretation-section {
  padding: 14px 16px;
  background: #fff;
  border: 1px solid #edf1f6;
  border-radius: 8px;

  h4 {
    position: relative;
    margin: 0 0 10px;
    padding-left: 10px;
    color: var(--pro-ant-color-primary);
    font-size: 14px;
    font-weight: 700;
    line-height: 22px;

    &::before {
      position: absolute;
      top: 5px;
      left: 0;
      width: 3px;
      height: 12px;
      content: "";
      background: var(--pro-ant-color-primary);
      border-radius: 999px;
    }
  }

  p {
    margin: 0;
    color: #334155;
    font-size: 13px;
    font-weight: 500;
    line-height: 24px;
  }

  ol {
    padding-left: 18px;
    margin: 0;
    color: #334155;
    font-size: 13px;
    font-weight: 500;
    line-height: 24px;
  }

  li + li {
    margin-top: 6px;
  }
}

.erxin-interpretation-section--streaming {
  position: relative;
  box-shadow: 0 4px 14px rgba(15, 23, 42, 0.03);
  animation: erxin-stream-card-in 0.18s ease-out;

  &::after {
    position: absolute;
    right: 12px;
    bottom: 10px;
    width: 6px;
    height: 6px;
    content: "";
    background: var(--pro-ant-color-primary);
    border-radius: 50%;
    opacity: 0.55;
    animation: erxin-stream-pulse 1s ease-in-out infinite;
  }
}

@keyframes erxin-stream-pulse {
  0%,
  100% {
    opacity: 0.45;
  }

  50% {
    opacity: 1;
  }
}

@keyframes erxin-stream-card-in {
  from {
    opacity: 0;
    transform: translateY(4px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.autismdev-report-tabs {
  flex: 1 1 auto;
  max-width: none;
  min-width: 0;
}

.autismdev-analysis-shell {
  min-height: 620px;
  padding: 18px;
  overflow-y: auto;
  background: #fbfdff;
  border: 1px solid #edf1f6;
  border-radius: 8px;
  scrollbar-color: #c6d1df transparent;
  scrollbar-width: thin;

  &::-webkit-scrollbar {
    width: 8px;
  }

  &::-webkit-scrollbar-track {
    background: transparent;
  }

  &::-webkit-scrollbar-thumb {
    background: #c6d1df;
    border-radius: 999px;
  }
}

.autismdev-analysis-state {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  gap: 8px;
  min-height: 430px;
  color: #7a8494;
  font-size: 13px;

  strong {
    color: #1f2937;
    font-size: 14px;
    font-weight: 600;
    line-height: 22px;
  }

  span,
  p {
    margin: 0;
    color: #7a8494;
    font-size: 13px;
    line-height: 20px;
  }
}

.autismdev-analysis-stream {
  display: flex;
  flex-direction: column;
  gap: 14px;
  min-height: 540px;
  padding: 16px 18px;
  background: #fff;
  border: 1px solid #e6edf6;
  border-radius: 10px;
  box-shadow: 0 8px 22px rgba(15, 23, 42, 0.04);
}

.autismdev-analysis-stream__head {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #edf1f6;

  > span {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    color: var(--pro-ant-color-primary);
    font-size: 12px;
    font-weight: 700;
    background: #eef6ff;
    border: 1px solid #d8eaff;
    border-radius: 10px;
  }
}

.autismdev-analysis-stream__title {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  gap: 7px;
  min-width: 0;

  strong {
    color: #1f2937;
    font-size: 15px;
    font-weight: 600;
    line-height: 22px;
  }

  small {
    color: #7a8494;
    font-size: 12px;
    line-height: 18px;
  }
}

.autismdev-analysis-stream__progress-line {
  display: grid;
  align-items: center;
  grid-template-columns: minmax(120px, auto) minmax(120px, 1fr) auto;
  gap: 10px;

  em {
    color: var(--pro-ant-color-primary);
    font-size: 12px;
    font-style: normal;
    font-weight: 700;
    line-height: 18px;
  }
}

.autismdev-analysis-progress {
  height: 8px;
  overflow: hidden;
  background: #eef4fb;
  border-radius: 999px;

  i {
    display: block;
    height: 100%;
    background: var(--pro-ant-color-primary);
    border-radius: inherit;
    transition: width 0.18s ease;
  }
}

.autismdev-analysis-stream__body {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  min-height: 410px;
  padding: 14px;
  overflow: hidden;
  background: #f8fafc;
  border: 1px solid #e6edf6;
  border-radius: 8px;
}

.autismdev-analysis-stream__body-title {
  flex: 0 0 auto;
  padding-bottom: 8px;
  margin-bottom: 10px;
  border-bottom: 1px solid #e6edf6;

  span {
    color: #1f2937;
    font-size: 12px;
    font-weight: 700;
    line-height: 18px;
  }
}

.autismdev-analysis-stream__text {
  flex: 1 1 auto;
  min-height: 0;
  overflow: auto;
  color: #334155;
  font-size: 13px;
  line-height: 24px;
  white-space: pre-wrap;
  scrollbar-color: #c6d1df transparent;
  scrollbar-width: thin;

  &::-webkit-scrollbar {
    width: 8px;
  }

  &::-webkit-scrollbar-track {
    background: transparent;
  }

  &::-webkit-scrollbar-thumb {
    background: #c6d1df;
    border-radius: 999px;
  }
}

.autismdev-analysis-stream__foot {
  color: #7a8494;
  font-size: 12px;
  line-height: 18px;
}

.autismdev-analysis-content {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.autismdev-analysis-content--streaming {
  min-height: 0;
}

.autismdev-analysis-doc-title {
  margin: 2px 0 12px;
  color: #000;
  font-size: 18px;
  font-weight: 500;
  line-height: 1.2;
  text-align: center;
}

.autismdev-analysis-meta {
  display: flex;
  align-items: center;
  gap: 20px;
  min-width: 0;
  margin-bottom: 8px;

  div {
    display: flex;
    align-items: center;
    min-width: 0;

    &:nth-child(1) {
      flex: 31 1 0;
    }

    &:nth-child(2) {
      flex: 25 1 0;
    }

    &:nth-child(3) {
      flex: 32 1 0;
    }
  }

  span {
    flex: 0 0 auto;
    color: #000;
    font-size: 13px;
    line-height: 1.2;
    white-space: nowrap;
  }

  em {
    flex: 1 1 auto;
    min-width: 0;
    height: 20px;
    overflow: hidden;
    color: #000;
    font-size: 13px;
    font-style: normal;
    line-height: 20px;
    text-align: center;
    text-overflow: ellipsis;
    white-space: nowrap;
    border-bottom: 1px solid #000;
  }
}

.autismdev-analysis-table-wrap {
  overflow-x: auto;
  background: #fff;
  scrollbar-color: #c6d1df transparent;
  scrollbar-width: thin;

  &::-webkit-scrollbar {
    height: 8px;
  }

  &::-webkit-scrollbar-track {
    background: transparent;
  }

  &::-webkit-scrollbar-thumb {
    background: #c6d1df;
    border-radius: 999px;
  }
}

.autismdev-analysis-table {
  width: 100%;
  min-width: 720px;
  color: #000;
  table-layout: fixed;
  background: #fff;
  border: 1px solid #000;
  border-collapse: collapse;
}

.autismdev-analysis-table__col-domain {
  width: 12%;
}

.autismdev-analysis-table__col-status {
  width: 25%;
}

.autismdev-analysis-table__col-strength {
  width: 34%;
}

.autismdev-analysis-table__col-target {
  width: 29%;
}

.autismdev-analysis-table th {
  height: 32px;
  padding: 0 6px;
  color: #000;
  font-size: 13px;
  font-weight: 400;
  line-height: 1.12;
  text-align: center;
  vertical-align: middle;
  background: #fff;
  border: 1px solid #000;
}

.autismdev-analysis-table td {
  height: 118px;
  padding: 8px 10px;
  color: #000;
  font-size: 12px;
  font-weight: 400;
  line-height: 1.55;
  white-space: pre-line;
  word-break: break-word;
  vertical-align: top;
  background: #fff;
  border: 1px solid #000;
}

.autismdev-analysis-table td:first-child {
  padding: 0 6px;
  font-size: 13px;
  line-height: 1.2;
  text-align: center;
  vertical-align: middle;
}

.autismdev-analysis-table--streaming {
  box-shadow: none;
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
    flex: 1 1 auto;
    min-width: 0;
    margin-left: 10px;
    padding-left: 10px;
    overflow: hidden;
    color: #9aa4b2;
    font-style: normal;
    text-overflow: ellipsis;
    white-space: nowrap;
    border-left: 1px solid #e2e8f0;
  }
}

.export-dimension__actions {
  flex: 0 0 auto;
  display: flex;
  align-items: center;
  gap: 8px;
}

.export-dimension--erxin {
  padding-top: 14px;

  .export-dimension__summary {
    margin-bottom: 12px;
  }

  .export-dimension__file {
    width: 38px;
    height: 38px;
    font-size: 12px;
  }

  .export-dimension__current {
    min-width: 82px;
  }

  .export-dimension__footer {
    margin-top: 14px;
  }
}

.export-dimension--autismdev {
  padding-top: 12px;

  .export-dimension__summary {
    gap: 10px;
    padding: 10px 12px;
    margin-bottom: 10px;
  }

  .export-dimension__file {
    width: 34px;
    height: 34px;
    font-size: 11px;
    border-radius: 7px;
  }

  .export-dimension__name {
    font-size: 14px;
    line-height: 20px;
  }

  .export-dimension__meta {
    margin-top: 1px;
  }

  .export-dimension__current {
    min-width: 76px;
    padding-left: 12px;
  }

  .export-dimension__footer {
    margin-top: 12px;
  }
}

.autismdev-export-card-list {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
}

.autismdev-export-toolbar {
  display: flex;
  grid-column: 1 / -1;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 34px;
  padding: 6px 10px;
  background: #f8fafc;
  border: 1px solid #e8edf4;
  border-radius: 8px;

  span {
    color: #64748b;
    font-size: 12px;
    line-height: 18px;
  }

  button {
    height: 24px;
    padding: 0 10px;
    color: var(--pro-ant-color-primary);
    font: inherit;
    font-size: 12px;
    font-weight: 600;
    cursor: pointer;
    background: #eef6ff;
    border: 1px solid #d8eaff;
    border-radius: 6px;

    &:disabled {
      cursor: not-allowed;
      opacity: 0.7;
    }
  }
}

.autismdev-export-row {
  display: flex;
  align-items: center;
  gap: 9px;
  width: 100%;
  min-height: 62px;
  padding: 8px 10px;
  color: inherit;
  font: inherit;
  text-align: left;
  cursor: pointer;
  background: #fff;
  border: 1px solid #e8edf4;
  border-radius: 8px;
  transition: border-color 0.16s ease, box-shadow 0.16s ease, background 0.16s ease;

  &:hover {
    background: #fbfdff;
    border-color: #c9ddf7;
  }

  &:disabled {
    cursor: not-allowed;
  }
}

.autismdev-export-row--active {
  background: #f7fbff;
  border-color: #8dc6ff;
  box-shadow: 0 3px 10px rgba(24, 144, 255, 0.08);
}

.autismdev-export-row--disabled {
  opacity: 0.66;
}

.autismdev-export-row__check {
  flex: 0 0 auto;
  position: relative;
  width: 15px;
  height: 15px;
  margin-top: 0;
  background: #fff;
  border: 1px solid #cbd5e1;
  border-radius: 4px;
}

.autismdev-export-row--active .autismdev-export-row__check {
  background: var(--pro-ant-color-primary);
  border-color: var(--pro-ant-color-primary);
}

.autismdev-export-row--active .autismdev-export-row__check::after {
  position: absolute;
  top: 2px;
  left: 4px;
  width: 4px;
  height: 8px;
  border: solid #fff;
  border-width: 0 2px 2px 0;
  content: "";
  transform: rotate(45deg);
}

.autismdev-export-row__body {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
  width: 100%;
}

.autismdev-export-row__head {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;

  strong {
    min-width: 0;
    color: #1f2937;
    font-size: 13px;
    font-weight: 600;
    line-height: 20px;
    white-space: normal;
  }

  em {
    flex: 0 0 auto;
    padding: 0 6px;
    color: var(--pro-ant-color-primary);
    font-size: 11px;
    font-style: normal;
    font-weight: 500;
    line-height: 18px;
    background: #eef6ff;
    border-radius: 999px;
  }

  .is-muted {
    color: #94a3b8;
    background: #f1f5f9;
  }
}

.autismdev-export-row__desc {
  display: block;
  overflow: hidden;
  color: #687386;
  font-size: 12px;
  line-height: 18px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.erxin-export-card-list {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.erxin-export-card {
  display: flex;
  align-items: flex-start;
  gap: 9px;
  min-width: 0;
  min-height: 116px;
  padding: 12px;
  color: inherit;
  font: inherit;
  text-align: left;
  cursor: pointer;
  background: #fff;
  border: 1px solid #e8edf4;
  border-radius: 10px;
  transition: border-color 0.16s ease, box-shadow 0.16s ease, background 0.16s ease;

  &:hover {
    background: #fbfdff;
    border-color: #c9ddf7;
  }

  &:focus-visible {
    outline: 2px solid rgba(24, 144, 255, 0.2);
    outline-offset: 2px;
  }

  &:disabled {
    cursor: not-allowed;
    opacity: 0.72;
  }
}

.erxin-export-card--active {
  background: #f7fbff;
  border-color: #8dc6ff;
  box-shadow: 0 6px 18px rgba(24, 144, 255, 0.08);
}

.erxin-export-card__check {
  flex: 0 0 auto;
  width: 16px;
  height: 16px;
  margin-top: 2px;
  background: #cbd5e1;
  border: 4px solid #f1f5f9;
  border-radius: 50%;
}

.erxin-export-card--active .erxin-export-card__check {
  background: var(--pro-ant-color-primary);
  border-color: #dcecff;
}

.erxin-export-card__body {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  min-width: 0;
}

.erxin-export-card__head {
  display: flex;
  align-items: center;
  gap: 8px;

  strong {
    min-width: 0;
    overflow: hidden;
    color: #1f2937;
    font-size: 13px;
    font-weight: 600;
    line-height: 22px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  em {
    flex: 0 0 auto;
    padding: 0 6px;
    color: var(--pro-ant-color-primary);
    font-size: 11px;
    font-style: normal;
    font-weight: 500;
    line-height: 18px;
    background: #eef6ff;
    border-radius: 999px;
  }
}

.erxin-export-card__desc {
  display: -webkit-box;
  min-height: 40px;
  margin-top: 6px;
  overflow: hidden;
  color: #687386;
  font-size: 12px;
  line-height: 20px;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.erxin-export-card__pages {
  align-self: flex-start;
  margin-top: auto;
  color: #1677ff;
  font-size: 12px;
  font-weight: 600;
  line-height: 18px;
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
    flex-wrap: wrap;
    padding-top: 8px;
    padding-left: 0;
    border-top: 1px solid #e6edf6;
    border-left: 0;
  }

  .report-module-summary__actions {
    width: 100%;
    justify-content: flex-end;
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

  .autismdev-export-card-list {
    grid-template-columns: repeat(2, minmax(0, 1fr));
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
