<script setup lang="ts">
import type { TableColumnsType } from 'ant-design-vue'
import { CloseOutlined, InfoCircleOutlined, PlusOutlined, ReloadOutlined, SearchOutlined, ThunderboltOutlined, UploadOutlined } from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'
import { computed, h, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  batchGeneratePlatformPEP3IEPMaterialAIApi,
  deletePlatformPEP3IEPMaterialGoalApi,
  deletePlatformPEP3IEPMaterialRuleApi,
  deletePlatformPEP3IEPMaterialTrainingApi,
  generatePlatformPEP3IEPMaterialAIApi,
  getScaleQuestionBankApi,
  pagePlatformPEP3IEPMaterialGoalsApi,
  pagePlatformPEP3IEPMaterialRulesApi,
  pagePlatformPEP3IEPMaterialTrainingApi,
  savePlatformPEP3IEPMaterialGoalApi,
  savePlatformPEP3IEPMaterialRuleApi,
  savePlatformPEP3IEPMaterialTrainingApi,
  type PEP3IEPMaterialAIGenerateResult,
  type PEP3IEPGoalMaterial,
  type PEP3IEPItemOptionRule,
  type PEP3IEPTrainingMaterial,
  type PlatformPageResult,
  type ScaleQuestionBank,
  type ScaleQuestionBankItem,
  type ScaleQuestionBankScoreOption,
} from '@/api/platform/scales'
import PlatformModalShell from '@/pages/platform/shared/platform-modal-shell.vue'
import messageService from '@/utils/messageService'

interface RuleForm extends Omit<PEP3IEPItemOptionRule, 'domainCode' | 'itemNo' | 'scoreValue'> {
  domainCode?: string
  itemNo?: number
  scoreValue?: number
  longGoalMaterialId?: number
  longGoal: string
}

type AIBatchMode = 'short_goal' | 'training'

interface AIBatchPreviewRow {
  id: string
  selected: boolean
  shortGoal?: string
  courseForm?: string
  trainingProject?: string
  trainingContent?: string
}

const router = useRouter()
const keyword = ref('')
const status = ref('')
const loading = ref(false)
const saving = ref(false)
const questionLoading = ref(false)
const materialDrawerOpen = ref(false)
const shortGoalDrawerOpen = ref(false)
const shortGoalModalOpen = ref(false)
const trainingModalOpen = ref(false)
const aiBatchModalOpen = ref(false)
const shortGoalLoading = ref(false)
const trainingLoading = ref(false)
const aiBatchLoading = ref(false)
const aiBatchSaving = ref(false)
const aiBatchMode = ref<AIBatchMode>('training')
const aiBatchCount = ref(5)
const aiBatchRows = ref<AIBatchPreviewRow[]>([])
const aiBatchLastError = ref('')
const aiGenerating = reactive({
  longGoal: false,
  shortGoal: false,
  training: false,
})

const questionBank = ref<ScaleQuestionBank | null>(null)
const totalRows = ref<PEP3IEPItemOptionRule[]>([])
const shortGoalRows = ref<PEP3IEPGoalMaterial[]>([])
const trainingRows = ref<PEP3IEPTrainingMaterial[]>([])
const activeRule = ref<PEP3IEPItemOptionRule | null>(null)
const activeShortGoal = ref<PEP3IEPGoalMaterial | null>(null)

const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
  showSizeChanger: true,
  showTotal: (total: number) => `共 ${total} 条`,
})

const ruleForm = reactive<RuleForm>({
  libraryScope: 'platform',
  instId: 0,
  itemNo: undefined,
  itemTitle: '',
  domainCode: undefined,
  domain: '',
  scoreValue: undefined,
  scoreLabel: '',
  scoreDescription: '',
  resultMeaning: '',
  generatePolicy: '',
  priority: 0,
  aiInstruction: '',
  status: 'active',
  goalMaterialIds: [],
  goalMaterials: [],
  longGoalMaterialId: undefined,
  longGoal: '',
})

const shortGoalForm = reactive<PEP3IEPGoalMaterial>({
  libraryScope: 'platform',
  instId: 0,
  materialType: 'short_term',
  parentGoalMaterialId: 0,
  domainCode: '',
  domain: '',
  longGoal: '',
  shortGoal: '',
  courseForm: undefined,
  priority: 100,
  status: 'active',
})

const trainingForm = reactive<PEP3IEPTrainingMaterial>({
  libraryScope: 'platform',
  instId: 0,
  goalMaterialId: undefined,
  trainingProject: '',
  trainingContent: '',
  priority: 100,
  status: 'active',
})

const statusOptions = [
  { label: '全部状态', value: '' },
  { label: '启用', value: 'active' },
  { label: '停用', value: 'inactive' },
]

const materialStatusOptions = statusOptions.filter(item => item.value)

const fallbackScoreOptions: ScaleQuestionBankScoreOption[] = [
  { label: '2分', value: 2, description: '通过 / 恰当' },
  { label: '1分', value: 1, description: '部分通过 / 轻微' },
  { label: '0分', value: 0, description: '未能通过 / 严重' },
]

const courseFormOptions = [
  { label: '个训', value: '个训' },
  { label: '集体课', value: '集体课' },
]

const questionItems = computed(() => questionBank.value?.items || [])

const domainOptions = computed(() => {
  const domains = questionBank.value?.domains || []
  return domains.map(item => ({
    label: item.scaleName || item.scaleCode,
    value: item.scaleCode,
  }))
})

const domainNameMap = computed(() => {
  const map = new Map<string, string>()
  for (const item of questionBank.value?.domains || [])
    map.set(item.scaleCode, item.scaleName)
  return map
})

const filteredQuestionOptions = computed(() => questionItems.value
  .filter(item => !ruleForm.domainCode || item.domainCode === ruleForm.domainCode)
  .map(item => ({
    label: questionTitle(item),
    value: Number(item.itemNo),
  })))

const selectedQuestion = computed(() => questionItems.value.find(item => Number(item.itemNo) === Number(ruleForm.itemNo)))

const scoreOptions = computed(() => {
  const options = selectedQuestion.value?.scoreOptions?.length ? selectedQuestion.value.scoreOptions : fallbackScoreOptions
  return [...options]
    .sort((a, b) => Number(b.value) - Number(a.value))
    .map(item => ({
      label: scoreOptionText(item),
      value: Number(item.value),
    }))
})

const activeLongGoal = computed(() => {
  const goals = activeRule.value?.goalMaterials || []
  return goals.find(item => item.materialType === 'long_term' || !item.parentGoalMaterialId) || goals[0]
})

const drawerTitle = computed(() => ruleForm.id ? '编辑题目选项长期目标' : '新增题目选项长期目标')
const shortGoalModalTitle = computed(() => shortGoalForm.id ? '编辑短期目标' : '新增短期目标')
const trainingModalTitle = computed(() => trainingForm.id ? '编辑训练内容' : '新增训练内容')

const columns = computed<TableColumnsType>(() => [
  { title: '题号', dataIndex: 'itemNo', key: 'itemNo', width: 76 },
  { title: '题目', key: 'question', width: 280, ellipsis: true },
  { title: '选项', dataIndex: 'scoreValue', key: 'scoreValue', width: 128 },
  { title: '领域', dataIndex: 'domain', key: 'domain', width: 140, ellipsis: true },
  { title: '关联长期目标', key: 'longGoal', ellipsis: true },
  { title: '状态', dataIndex: 'status', key: 'status', width: 88 },
  { title: '操作', key: 'action', width: 210, fixed: 'right' as const },
])

const shortGoalColumns = computed<TableColumnsType>(() => [
  { title: '短期目标', dataIndex: 'shortGoal', key: 'shortGoal', ellipsis: true, customCell: record => shortGoalCellProps(record as PEP3IEPGoalMaterial, 'first') },
  { title: '课程形式', dataIndex: 'courseForm', key: 'courseForm', width: 100, customCell: record => shortGoalCellProps(record as PEP3IEPGoalMaterial) },
  { title: '状态', dataIndex: 'status', key: 'status', width: 88, customCell: record => shortGoalCellProps(record as PEP3IEPGoalMaterial) },
  { title: '操作', key: 'action', width: 120, fixed: 'right', customCell: record => shortGoalCellProps(record as PEP3IEPGoalMaterial, 'last') },
])

const trainingColumns: TableColumnsType = [
  { title: '训练项目', dataIndex: 'trainingProject', key: 'trainingProject', width: 220, ellipsis: true },
  { title: '训练内容', dataIndex: 'trainingContent', key: 'trainingContent', ellipsis: true },
  { title: '状态', dataIndex: 'status', key: 'status', width: 88 },
  { title: '操作', key: 'action', width: 120, fixed: 'right' },
]

const aiBatchModalTitle = computed(() => aiBatchMode.value === 'short_goal' ? 'AI批量生成短期目标' : 'AI批量生成训练内容')

const aiBatchHelpText = computed(() => {
  if (aiBatchMode.value === 'short_goal')
    return '根据当前题目选项和长期目标，一次生成多条短期目标和课程形式。生成结果先预览，勾选后再保存。'
  return '根据当前短期目标，一次生成多条训练项目和训练内容。生成结果先预览，勾选后再保存。'
})

const aiBatchColumns = computed<TableColumnsType>(() => {
  if (aiBatchMode.value === 'short_goal') {
    return [
      { title: '短期目标', dataIndex: 'shortGoal', key: 'shortGoal' },
      { title: '课程形式', dataIndex: 'courseForm', key: 'courseForm', width: 150 },
    ]
  }
  return [
    { title: '训练项目', dataIndex: 'trainingProject', key: 'trainingProject', width: 220 },
    { title: '训练内容', dataIndex: 'trainingContent', key: 'trainingContent' },
  ]
})

const aiBatchSelectedCount = computed(() => aiBatchRows.value.filter(item => item.selected).length)

const aiBatchRowSelection = computed(() => ({
  selectedRowKeys: aiBatchRows.value.filter(item => item.selected).map(item => item.id),
  onChange: (keys: Array<string | number>) => {
    const selected = new Set(keys.map(key => String(key)))
    aiBatchRows.value.forEach((item) => {
      item.selected = selected.has(item.id)
    })
  },
}))

function unwrap<T>(res: any): T {
  return (res?.data ?? res?.result ?? res) as T
}

function questionTitle(item: ScaleQuestionBankItem) {
  return item.itemTitle || item.testItem || '未命名题目'
}

function pagePayload(extra: Record<string, any> = {}) {
  return {
    pageRequestModel: {
      pageIndex: pagination.current,
      pageSize: pagination.pageSize,
    },
    queryModel: {
      keyword: keyword.value.trim() || undefined,
      status: status.value || undefined,
      ...extra,
    },
  }
}

function resetRuleForm() {
  Object.assign(ruleForm, {
    id: undefined,
    libraryScope: 'platform',
    instId: 0,
    itemNo: undefined,
    itemTitle: '',
    domainCode: undefined,
    domain: '',
    scoreValue: undefined,
    scoreLabel: '',
    scoreDescription: '',
    resultMeaning: '',
    generatePolicy: '',
    priority: 0,
    aiInstruction: '',
    status: 'active',
    goalMaterialIds: [],
    goalMaterials: [],
    longGoalMaterialId: undefined,
    longGoal: '',
  })
}

function resetShortGoalForm(parent?: PEP3IEPGoalMaterial) {
  Object.assign(shortGoalForm, {
    id: undefined,
    libraryScope: 'platform',
    instId: 0,
    materialType: 'short_term',
    parentGoalMaterialId: parent?.id || 0,
    domainCode: parent?.domainCode || '',
    domain: parent?.domain || '',
    longGoal: parent?.longGoal || '',
    shortGoal: '',
    courseForm: undefined,
    priority: 100,
    status: 'active',
  })
}

function resetTrainingForm(goal?: PEP3IEPGoalMaterial) {
  Object.assign(trainingForm, {
    id: undefined,
    libraryScope: 'platform',
    instId: 0,
    goalMaterialId: goal?.id,
    trainingProject: '',
    trainingContent: '',
    priority: 100,
    status: 'active',
  })
}

async function loadQuestionBank() {
  questionLoading.value = true
  try {
    questionBank.value = unwrap<ScaleQuestionBank>(await getScaleQuestionBankApi({ scaleCode: 'PEP3' }))
  } catch (error: any) {
    questionBank.value = null
    messageService.error(error?.response?.data?.message || error?.message || '加载PEP3题库失败')
  } finally {
    questionLoading.value = false
  }
}

async function fetchCurrent() {
  loading.value = true
  try {
    const data = unwrap<PlatformPageResult<PEP3IEPItemOptionRule>>(await pagePlatformPEP3IEPMaterialRulesApi(pagePayload()))
    totalRows.value = data.items || []
    pagination.total = Number(data.total || 0)
  } catch (error: any) {
    messageService.error(error?.response?.data?.message || error?.message || '获取素材库失败')
  } finally {
    loading.value = false
  }
}

async function loadShortGoals(parentID: number) {
  if (!parentID) {
    shortGoalRows.value = []
    selectShortGoalForTraining(null)
    return
  }
  shortGoalLoading.value = true
  try {
    const data = unwrap<PlatformPageResult<PEP3IEPGoalMaterial>>(await pagePlatformPEP3IEPMaterialGoalsApi({
      pageRequestModel: { pageIndex: 1, pageSize: 200 },
      queryModel: { materialType: 'short_term', parentGoalMaterialId: parentID },
    }))
    shortGoalRows.value = data.items || []
    const current = shortGoalRows.value.find(item => Number(item.id) === Number(activeShortGoal.value?.id))
    selectShortGoalForTraining(current || shortGoalRows.value[0] || null)
  } catch (error: any) {
    shortGoalRows.value = []
    selectShortGoalForTraining(null)
    messageService.error(error?.response?.data?.message || error?.message || '获取短期目标失败')
  } finally {
    shortGoalLoading.value = false
  }
}

async function loadTrainingRows(goalID: number) {
  if (!goalID) {
    trainingRows.value = []
    return
  }
  trainingLoading.value = true
  try {
    const data = unwrap<PlatformPageResult<PEP3IEPTrainingMaterial>>(await pagePlatformPEP3IEPMaterialTrainingApi({
      pageRequestModel: { pageIndex: 1, pageSize: 200 },
      queryModel: { goalMaterialId: goalID },
    }))
    trainingRows.value = data.items || []
  } catch (error: any) {
    trainingRows.value = []
    messageService.error(error?.response?.data?.message || error?.message || '获取训练内容失败')
  } finally {
    trainingLoading.value = false
  }
}

function handleTableChange(page: any) {
  pagination.current = page.current || 1
  pagination.pageSize = page.pageSize || 20
  void fetchCurrent()
}

function handleSearch() {
  pagination.current = 1
  void fetchCurrent()
}

function resetFilters() {
  keyword.value = ''
  status.value = ''
  pagination.current = 1
  void fetchCurrent()
}

function openCreate() {
  resetRuleForm()
  if (!questionBank.value)
    void loadQuestionBank()
  materialDrawerOpen.value = true
}

function openEdit(record: PEP3IEPItemOptionRule) {
  resetRuleForm()
  Object.assign(ruleForm, JSON.parse(JSON.stringify(record || {})), { libraryScope: 'platform', instId: 0 })
  const longGoal = firstLongGoal(record)
  if (longGoal) {
    ruleForm.longGoalMaterialId = Number(longGoal.id)
    ruleForm.longGoal = longGoal.longGoal || ''
  }
  if (!questionBank.value)
    void loadQuestionBank()
  materialDrawerOpen.value = true
}

async function openShortGoalDrawer(record: PEP3IEPItemOptionRule) {
  const longGoal = firstLongGoal(record)
  if (!longGoal?.id) {
    messageService.warning('请先保存长期目标')
    return
  }
  activeRule.value = record
  activeShortGoal.value = null
  trainingRows.value = []
  resetShortGoalForm(longGoal)
  resetTrainingForm()
  shortGoalDrawerOpen.value = true
  await loadShortGoals(Number(longGoal.id))
}

function firstLongGoal(record?: PEP3IEPItemOptionRule | null) {
  const goals = record?.goalMaterials || []
  return goals.find(item => item.materialType === 'long_term' || !item.parentGoalMaterialId) || goals[0]
}

function asRuleRecord(record: unknown): PEP3IEPItemOptionRule {
  return record as PEP3IEPItemOptionRule
}

function asGoalRecord(record: unknown): PEP3IEPGoalMaterial {
  return record as PEP3IEPGoalMaterial
}

function asTrainingRecord(record: unknown): PEP3IEPTrainingMaterial {
  return record as PEP3IEPTrainingMaterial
}

function applyDomainToRule(domainCode?: string) {
  ruleForm.domainCode = domainCode || undefined
  ruleForm.domain = domainNameMap.value.get(ruleForm.domainCode || '') || ''
  ruleForm.itemNo = undefined
  ruleForm.itemTitle = ''
  ruleForm.scoreValue = undefined
  ruleForm.scoreLabel = ''
  ruleForm.scoreDescription = ''
}

function applyQuestionToRule(itemNo?: number) {
  const question = questionItems.value.find(item => Number(item.itemNo) === Number(itemNo))
  if (!question)
    return
  ruleForm.itemNo = Number(question.itemNo)
  ruleForm.itemTitle = questionTitle(question)
  ruleForm.domainCode = question.domainCode || ruleForm.domainCode
  ruleForm.domain = question.domainName || domainNameMap.value.get(ruleForm.domainCode || '') || ruleForm.domain
  const values = scoreOptions.value.map(item => Number(item.value))
  ruleForm.scoreValue = values.includes(0) ? 0 : undefined
  applyScoreToRule()
}

function applyScoreToRule() {
  const option = selectedScoreOption()
  if (ruleForm.scoreValue === undefined || ruleForm.scoreValue === null) {
    ruleForm.scoreLabel = ''
    ruleForm.scoreDescription = ''
    return
  }
  ruleForm.scoreLabel = option?.label || `${ruleForm.scoreValue}分`
  ruleForm.scoreDescription = option?.description || ''
}

function selectedScoreOption() {
  const options = selectedQuestion.value?.scoreOptions?.length ? selectedQuestion.value.scoreOptions : fallbackScoreOptions
  return options.find(item => Number(item.value) === Number(ruleForm.scoreValue))
}

function normalizedScoreValue(value: unknown) {
  if (value === undefined || value === null || value === '')
    return undefined
  const numberValue = Number(value)
  return Number.isNaN(numberValue) ? undefined : numberValue
}

function currentRuleAIContext() {
  const record = activeRule.value
  return {
    domain: record?.domain || shortGoalForm.domain || activeLongGoal.value?.domain || ruleForm.domain || '',
    domainCode: record?.domainCode || shortGoalForm.domainCode || activeLongGoal.value?.domainCode || ruleForm.domainCode || '',
    itemNo: normalizedScoreValue(record?.itemNo ?? ruleForm.itemNo),
    itemTitle: record?.itemTitle || ruleForm.itemTitle || '',
    scoreValue: normalizedScoreValue(record?.scoreValue ?? ruleForm.scoreValue),
    scoreLabel: record?.scoreLabel || ruleForm.scoreLabel || '',
    scoreDescription: record?.scoreDescription || ruleForm.scoreDescription || '',
  }
}

function compactExistingTexts(values: string[]) {
  const seen = new Set<string>()
  const result: string[] = []
  for (const value of values) {
    const text = String(value || '').trim()
    if (!text || seen.has(text))
      continue
    seen.add(text)
    result.push(text)
    if (result.length >= 20)
      break
  }
  return result
}

function existingShortGoalsForAI() {
  const currentID = Number(shortGoalForm.id || 0)
  return compactExistingTexts(shortGoalRows.value
    .filter(item => !currentID || Number(item.id) !== currentID)
    .map(item => item.shortGoal))
}

function existingTrainingProjectsForAI() {
  const currentID = Number(trainingForm.id || 0)
  return compactExistingTexts(trainingRows.value
    .filter(item => !currentID || Number(item.id) !== currentID)
    .map(item => item.trainingProject))
}

function existingTrainingContentsForAI() {
  const currentID = Number(trainingForm.id || 0)
  return compactExistingTexts(trainingRows.value
    .filter(item => !currentID || Number(item.id) !== currentID)
    .map(item => item.trainingContent))
}

async function generateLongGoal() {
  if (!Number(ruleForm.itemNo)) {
    messageService.warning('请先选择题目')
    return
  }
  if (ruleForm.scoreValue === undefined || ruleForm.scoreValue === null) {
    messageService.warning('请先选择选项')
    return
  }
  aiGenerating.longGoal = true
  try {
    const result = unwrap<PEP3IEPMaterialAIGenerateResult>(await generatePlatformPEP3IEPMaterialAIApi({
      target: 'long_goal',
      domain: ruleForm.domain,
      domainCode: ruleForm.domainCode,
      itemNo: Number(ruleForm.itemNo),
      itemTitle: ruleForm.itemTitle,
      scoreValue: Number(ruleForm.scoreValue),
      scoreLabel: ruleForm.scoreLabel,
      scoreDescription: ruleForm.scoreDescription,
    }))
    if (result.longGoal)
      ruleForm.longGoal = result.longGoal
  } catch (error: any) {
    messageService.error(error?.response?.data?.message || error?.message || 'AI生成失败')
  } finally {
    aiGenerating.longGoal = false
  }
}

async function generateShortGoal() {
  const context = currentRuleAIContext()
  const longGoal = activeLongGoal.value?.longGoal || shortGoalForm.longGoal
  if (!Number(context.itemNo)) {
    messageService.warning('缺少题目信息')
    return
  }
  if (context.scoreValue === undefined || context.scoreValue === null) {
    messageService.warning('缺少选项信息')
    return
  }
  if (!String(longGoal || '').trim()) {
    messageService.warning('请先保存或填写长期目标')
    return
  }
  aiGenerating.shortGoal = true
  try {
    const result = unwrap<PEP3IEPMaterialAIGenerateResult>(await generatePlatformPEP3IEPMaterialAIApi({
      target: 'short_goal',
      ...context,
      longGoal,
      existingShortGoals: existingShortGoalsForAI(),
    }))
    if (result.shortGoal)
      shortGoalForm.shortGoal = result.shortGoal
    if (result.courseForm)
      shortGoalForm.courseForm = result.courseForm
  } catch (error: any) {
    messageService.error(error?.response?.data?.message || error?.message || 'AI生成失败')
  } finally {
    aiGenerating.shortGoal = false
  }
}

async function generateTraining() {
  const context = currentRuleAIContext()
  const shortGoal = activeShortGoal.value?.shortGoal || shortGoalForm.shortGoal
  if (!String(shortGoal || '').trim()) {
    messageService.warning('请先选择或填写短期目标')
    return
  }
  aiGenerating.training = true
  try {
    const result = unwrap<PEP3IEPMaterialAIGenerateResult>(await generatePlatformPEP3IEPMaterialAIApi({
      target: 'training',
      ...context,
      longGoal: activeLongGoal.value?.longGoal || shortGoalForm.longGoal || '',
      shortGoal,
      courseForm: activeShortGoal.value?.courseForm || shortGoalForm.courseForm,
      existingTrainingProjects: existingTrainingProjectsForAI(),
      existingTrainingContents: existingTrainingContentsForAI(),
    }))
    if (result.trainingProject)
      trainingForm.trainingProject = result.trainingProject
    if (result.trainingContent)
      trainingForm.trainingContent = result.trainingContent
  } catch (error: any) {
    messageService.error(error?.response?.data?.message || error?.message || 'AI生成失败')
  } finally {
    aiGenerating.training = false
  }
}

function resetAIBatch(mode: AIBatchMode) {
  aiBatchMode.value = mode
  aiBatchCount.value = 5
  aiBatchRows.value = []
  aiBatchLastError.value = ''
}

function openBatchShortGoalModal() {
  if (!activeLongGoal.value?.id) {
    messageService.warning('请先保存长期目标')
    return
  }
  const context = currentRuleAIContext()
  if (!Number(context.itemNo)) {
    messageService.warning('缺少题目信息')
    return
  }
  if (context.scoreValue === undefined || context.scoreValue === null) {
    messageService.warning('缺少选项信息')
    return
  }
  resetAIBatch('short_goal')
  aiBatchModalOpen.value = true
}

function openBatchTrainingModal() {
  if (!activeShortGoal.value?.id) {
    messageService.warning('请先选择短期目标')
    return
  }
  resetAIBatch('training')
  aiBatchModalOpen.value = true
}

function buildAIBatchRequest() {
  const context = currentRuleAIContext()
  if (aiBatchMode.value === 'short_goal') {
    const longGoal = activeLongGoal.value?.longGoal || shortGoalForm.longGoal
    return {
      target: 'short_goal' as const,
      count: Number(aiBatchCount.value || 5),
      ...context,
      longGoal,
      existingShortGoals: existingShortGoalsForAI(),
    }
  }
  const shortGoal = activeShortGoal.value?.shortGoal || shortGoalForm.shortGoal
  return {
    target: 'training' as const,
    count: Number(aiBatchCount.value || 5),
    ...context,
    longGoal: activeLongGoal.value?.longGoal || shortGoalForm.longGoal || '',
    shortGoal,
    courseForm: activeShortGoal.value?.courseForm || shortGoalForm.courseForm,
    existingTrainingProjects: existingTrainingProjectsForAI(),
    existingTrainingContents: existingTrainingContentsForAI(),
  }
}

function validateAIBatchRequest() {
  if (aiBatchMode.value === 'short_goal') {
    if (!activeLongGoal.value?.id)
      return '请先保存长期目标'
    const context = currentRuleAIContext()
    if (!Number(context.itemNo))
      return '缺少题目信息'
    if (context.scoreValue === undefined || context.scoreValue === null)
      return '缺少选项信息'
  } else if (!activeShortGoal.value?.id) {
    return '请先选择短期目标'
  }
  if (Number(aiBatchCount.value) < 1)
    return '生成数量不能小于1'
  return ''
}

async function runAIBatchGenerate() {
  const warning = validateAIBatchRequest()
  if (warning) {
    messageService.warning(warning)
    return
  }
  aiBatchLoading.value = true
  aiBatchLastError.value = ''
  try {
    const result = unwrap<{
      items: PEP3IEPMaterialAIGenerateResult[]
      failed: number
      lastError?: string
    }>(await batchGeneratePlatformPEP3IEPMaterialAIApi(buildAIBatchRequest()))
    aiBatchRows.value = (result.items || []).map((item: PEP3IEPMaterialAIGenerateResult, index: number) => ({
      id: `${Date.now()}-${index}`,
      selected: true,
      shortGoal: item.shortGoal || '',
      courseForm: item.courseForm || undefined,
      trainingProject: item.trainingProject || '',
      trainingContent: item.trainingContent || '',
    }))
    aiBatchLastError.value = result.lastError || ''
    if (result.failed > 0 && result.items?.length)
      messageService.warning(`已生成 ${result.items.length} 条，另有 ${result.failed} 次生成未通过校验`)
    else
      messageService.success(`已生成 ${result.items?.length || 0} 条`)
  } catch (error: any) {
    aiBatchRows.value = []
    messageService.error(error?.response?.data?.message || error?.message || 'AI批量生成失败')
  } finally {
    aiBatchLoading.value = false
  }
}

function selectedAIBatchRows() {
  return aiBatchRows.value.filter(item => item.selected)
}

function validateAIBatchPreviewRows(rows: AIBatchPreviewRow[]) {
  if (!rows.length)
    return '请先选择要保存的数据'
  if (aiBatchMode.value === 'short_goal') {
    if (rows.some(item => !String(item.shortGoal || '').trim()))
      return '短期目标不能为空'
    if (rows.some(item => !String(item.courseForm || '').trim()))
      return '课程形式不能为空'
  } else {
    if (rows.some(item => !String(item.trainingProject || '').trim()))
      return '训练项目不能为空'
    if (rows.some(item => !String(item.trainingContent || '').trim()))
      return '训练内容不能为空'
  }
  return ''
}

async function saveAIBatchRows() {
  const rows = selectedAIBatchRows()
  const warning = validateAIBatchPreviewRows(rows)
  if (warning) {
    messageService.warning(warning)
    return
  }
  aiBatchSaving.value = true
  let saved = 0
  let failed = 0
  try {
    if (aiBatchMode.value === 'short_goal') {
      for (const row of rows) {
        try {
          await savePlatformPEP3IEPMaterialGoalApi({
            libraryScope: 'platform',
            instId: 0,
            materialType: 'short_term',
            parentGoalMaterialId: Number(activeLongGoal.value?.id || 0),
            domainCode: activeLongGoal.value?.domainCode || '',
            domain: activeLongGoal.value?.domain || '',
            longGoal: activeLongGoal.value?.longGoal || '',
            shortGoal: String(row.shortGoal || '').trim(),
            courseForm: String(row.courseForm || '').trim(),
            priority: 100,
            status: 'active',
          })
          saved++
        } catch {
          failed++
        }
      }
      await loadShortGoals(Number(activeLongGoal.value?.id || 0))
      await fetchCurrent()
    } else {
      for (const row of rows) {
        try {
          await savePlatformPEP3IEPMaterialTrainingApi({
            libraryScope: 'platform',
            instId: 0,
            goalMaterialId: Number(activeShortGoal.value?.id || 0),
            trainingProject: String(row.trainingProject || '').trim(),
            trainingContent: String(row.trainingContent || '').trim(),
            priority: 100,
            status: 'active',
          })
          saved++
        } catch {
          failed++
        }
      }
      await loadTrainingRows(Number(activeShortGoal.value?.id || 0))
    }
    if (saved > 0)
      messageService.success(`已保存 ${saved} 条`)
    if (failed > 0)
      messageService.warning(`${failed} 条保存失败，请检查后重试`)
    if (saved > 0 && failed === 0)
      aiBatchModalOpen.value = false
  } finally {
    aiBatchSaving.value = false
  }
}

function scoreOptionText(option?: ScaleQuestionBankScoreOption) {
  if (!option)
    return ''
  const valueText = `${option.value}分`
  const label = String(option.label || '').trim()
  const description = String(option.description || '').trim()
  if (description)
    return `${valueText}：${description}`
  if (label && label !== valueText)
    return `${valueText}：${label}`
  return valueText
}

function validateRuleForm() {
  if (!ruleForm.domainCode)
    return '请选择领域'
  if (!Number(ruleForm.itemNo))
    return '请选择题目'
  if (![0, 1, 2].includes(Number(ruleForm.scoreValue)))
    return '请选择选项'
  if (!String(ruleForm.longGoal || '').trim())
    return '请填写长期目标'
  return ''
}

function validateShortGoalForm() {
  if (!Number(shortGoalForm.parentGoalMaterialId))
    return '缺少所属长期目标'
  if (!String(shortGoalForm.shortGoal || '').trim())
    return '请填写短期目标'
  if (!String(shortGoalForm.courseForm || '').trim())
    return '请选择课程形式'
  return ''
}

function validateTrainingForm() {
  if (!Number(trainingForm.goalMaterialId))
    return '请选择短期目标'
  if (!String(trainingForm.trainingProject || '').trim())
    return '请填写训练项目'
  if (!String(trainingForm.trainingContent || '').trim())
    return '请填写训练内容'
  return ''
}

async function saveRule() {
  const warning = validateRuleForm()
  if (warning) {
    messageService.warning(warning)
    return
  }
  saving.value = true
  try {
    const savedGoal = unwrap<PEP3IEPGoalMaterial>(await savePlatformPEP3IEPMaterialGoalApi({
      id: ruleForm.longGoalMaterialId,
      libraryScope: 'platform',
      instId: 0,
      materialType: 'long_term',
      parentGoalMaterialId: 0,
      domainCode: ruleForm.domainCode,
      domain: ruleForm.domain,
      longGoal: String(ruleForm.longGoal || '').trim(),
      shortGoal: '',
      courseForm: '',
      priority: 100,
      status: ruleForm.status || 'active',
    }))
    await savePlatformPEP3IEPMaterialRuleApi({
      id: ruleForm.id,
      libraryScope: 'platform',
      instId: 0,
      itemNo: Number(ruleForm.itemNo),
      itemTitle: ruleForm.itemTitle || '',
      domainCode: ruleForm.domainCode || '',
      domain: ruleForm.domain || '',
      scoreValue: Number(ruleForm.scoreValue),
      scoreLabel: ruleForm.scoreLabel || '',
      scoreDescription: ruleForm.scoreDescription || '',
      resultMeaning: defaultResultMeaning(Number(ruleForm.scoreValue)),
      generatePolicy: defaultGeneratePolicy(Number(ruleForm.scoreValue)),
      priority: 0,
      aiInstruction: '',
      status: ruleForm.status || 'active',
      goalMaterialIds: savedGoal.id ? [Number(savedGoal.id)] : [],
    })
    messageService.success('保存成功')
    materialDrawerOpen.value = false
    await fetchCurrent()
  } catch (error: any) {
    messageService.error(error?.response?.data?.message || error?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function saveShortGoal() {
  const warning = validateShortGoalForm()
  if (warning) {
    messageService.warning(warning)
    return
  }
  saving.value = true
  try {
    await savePlatformPEP3IEPMaterialGoalApi({
      ...shortGoalForm,
      libraryScope: 'platform',
      instId: 0,
      materialType: 'short_term',
      parentGoalMaterialId: Number(shortGoalForm.parentGoalMaterialId),
      longGoal: activeLongGoal.value?.longGoal || shortGoalForm.longGoal || '',
      priority: 100,
    })
    messageService.success('短期目标已保存')
    shortGoalModalOpen.value = false
    resetShortGoalForm(activeLongGoal.value)
    await loadShortGoals(Number(activeLongGoal.value?.id || 0))
    await fetchCurrent()
  } catch (error: any) {
    messageService.error(error?.response?.data?.message || error?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function saveTraining() {
  if (activeShortGoal.value?.id)
    trainingForm.goalMaterialId = Number(activeShortGoal.value.id)
  const warning = validateTrainingForm()
  if (warning) {
    messageService.warning(warning)
    return
  }
  saving.value = true
  try {
    await savePlatformPEP3IEPMaterialTrainingApi({
      ...trainingForm,
      libraryScope: 'platform',
      instId: 0,
      goalMaterialId: Number(trainingForm.goalMaterialId),
      priority: 100,
    })
    messageService.success('训练内容已保存')
    trainingModalOpen.value = false
    resetTrainingForm(activeShortGoal.value || undefined)
    await loadTrainingRows(Number(activeShortGoal.value?.id || 0))
  } catch (error: any) {
    messageService.error(error?.response?.data?.message || error?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

function editShortGoal(record: PEP3IEPGoalMaterial) {
  Object.assign(shortGoalForm, JSON.parse(JSON.stringify(record || {})), {
    libraryScope: 'platform',
    instId: 0,
    materialType: 'short_term',
    parentGoalMaterialId: Number(record.parentGoalMaterialId || activeLongGoal.value?.id || 0),
  })
}

function openCreateShortGoalModal() {
  if (!activeLongGoal.value?.id) {
    messageService.warning('请先保存长期目标')
    return
  }
  resetShortGoalForm(activeLongGoal.value)
  shortGoalModalOpen.value = true
}

function openEditShortGoalModal(record: PEP3IEPGoalMaterial) {
  editShortGoal(record)
  shortGoalModalOpen.value = true
}

function selectShortGoalForTraining(record: PEP3IEPGoalMaterial | null) {
  activeShortGoal.value = record
  resetTrainingForm(record || undefined)
  if (record?.id)
    void loadTrainingRows(Number(record.id))
  else
    trainingRows.value = []
}

function isActiveShortGoal(record?: PEP3IEPGoalMaterial | null) {
  return Number(record?.id) === Number(activeShortGoal.value?.id)
}

function shortGoalCellProps(record: PEP3IEPGoalMaterial, position: 'first' | 'middle' | 'last' = 'middle') {
  const active = isActiveShortGoal(record)
  const style: Record<string, string> = {
    cursor: 'pointer',
  }
  if (active) {
    Object.assign(style, {
      backgroundColor: '#eef6ff',
    })
    if (position === 'first') {
      Object.assign(style, {
        color: '#1f2937',
        fontWeight: '650',
        boxShadow: 'inset 3px 0 0 #4096ff',
      })
    }
  }
  return { style }
}

function shortGoalRowProps(record: PEP3IEPGoalMaterial) {
  return {
    onClick: (event: MouseEvent) => {
      const target = event.target as HTMLElement | null
      if (target?.closest('a,button,.ant-btn'))
        return
      selectShortGoalForTraining(record)
    },
  }
}

function editTraining(record: PEP3IEPTrainingMaterial) {
  Object.assign(trainingForm, JSON.parse(JSON.stringify(record || {})), {
    libraryScope: 'platform',
    instId: 0,
    goalMaterialId: Number(record.goalMaterialId || activeShortGoal.value?.id || 0),
  })
}

function openCreateTrainingModal(goal?: PEP3IEPGoalMaterial | null) {
  const targetGoal = goal || activeShortGoal.value
  if (!targetGoal?.id) {
    messageService.warning('请先选择短期目标')
    return
  }
  selectShortGoalForTraining(targetGoal)
  resetTrainingForm(targetGoal)
  trainingModalOpen.value = true
}

function openEditTrainingModal(record: PEP3IEPTrainingMaterial) {
  editTraining(record)
  trainingModalOpen.value = true
}

function confirmDeleteRule(record: PEP3IEPItemOptionRule) {
  Modal.confirm({
    title: '确认删除这条题目选项长期目标？',
    content: '删除后不会再参与后续IEP匹配。',
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      try {
        await deletePlatformPEP3IEPMaterialRuleApi(Number(record.id))
        messageService.success('删除成功')
        await fetchCurrent()
      } catch (error: any) {
        messageService.error(error?.response?.data?.message || error?.message || '删除失败')
      }
    },
  })
}

function confirmDeleteShortGoal(record: PEP3IEPGoalMaterial) {
  Modal.confirm({
    title: '确认删除这条短期目标？',
    content: '删除后它下面的训练内容不会再参与后续IEP匹配。',
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      try {
        await deletePlatformPEP3IEPMaterialGoalApi(Number(record.id))
        messageService.success('删除成功')
        await loadShortGoals(Number(activeLongGoal.value?.id || 0))
      } catch (error: any) {
        messageService.error(error?.response?.data?.message || error?.message || '删除失败')
      }
    },
  })
}

function confirmDeleteTraining(record: PEP3IEPTrainingMaterial) {
  Modal.confirm({
    title: '确认删除这条训练内容？',
    content: '删除后不会再作为AI拆解月计划或周计划的素材。',
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      try {
        await deletePlatformPEP3IEPMaterialTrainingApi(Number(record.id))
        messageService.success('删除成功')
        await loadTrainingRows(Number(activeShortGoal.value?.id || 0))
      } catch (error: any) {
        messageService.error(error?.response?.data?.message || error?.message || '删除失败')
      }
    },
  })
}

function scoreBriefText(value?: number) {
  return `${value ?? '-'}分`
}

function scoreFullText(value?: number, record?: PEP3IEPItemOptionRule) {
  if (record?.scoreDescription)
    return `${value ?? '-'}分：${record.scoreDescription}`
  if (record?.scoreLabel && record.scoreLabel !== `${value ?? '-'}分`)
    return `${value ?? '-'}分：${record.scoreLabel}`
  const option = fallbackScoreOptions.find(item => Number(item.value) === Number(value))
  return option ? scoreOptionText(option) : `${value ?? '-'}分`
}

function questionText(record: PEP3IEPItemOptionRule) {
  return record.itemTitle || '未命名题目'
}

function longGoalText(record: PEP3IEPItemOptionRule) {
  return firstLongGoal(record)?.longGoal || '未关联长期目标'
}

function statusText(value?: string) {
  return value === 'inactive' ? '停用' : '启用'
}

function statusColor(value?: string) {
  return value === 'inactive' ? 'default' : 'green'
}

function defaultResultMeaning(score: number) {
  if (score === 2)
    return '已通过，默认用于维持、泛化或提高独立性。'
  if (score === 1)
    return '部分通过，优先转化为季度或半年度IEP目标。'
  if (score === 0)
    return '未通过，生成前备或基础目标。'
  return ''
}

function defaultGeneratePolicy(score: number) {
  if (score === 2)
    return 'skip_or_generalize'
  if (score === 0)
    return 'prerequisite_goal'
  return 'primary_goal'
}

onMounted(() => {
  void loadQuestionBank()
  void fetchCurrent()
})
</script>

<template>
  <div class="pep3-material-page">
    <div class="page-head">
      <div>
        <h2>PEP3 IEP素材库</h2>
        <p>题目选项关联长期目标；长期目标下维护短期目标；短期目标下维护训练项目和训练内容。</p>
      </div>
      <a-space>
        <a-button :icon="h(ReloadOutlined)" @click="fetchCurrent">
          刷新
        </a-button>
        <a-button :icon="h(UploadOutlined)" @click="router.push('/platform/scales/pep3-iep-materials/import')">
          批量导入
        </a-button>
        <a-button type="primary" :icon="h(PlusOutlined)" @click="openCreate">
          新增题目选项长期目标
        </a-button>
      </a-space>
    </div>

    <div class="toolbar">
      <div class="toolbar-title">
        题目选项关联长期目标
      </div>
      <div class="filters">
        <a-input
          v-model:value="keyword"
          allow-clear
          class="keyword"
          placeholder="搜索题目或目标"
          @press-enter="handleSearch"
        >
          <template #prefix>
            <SearchOutlined />
          </template>
        </a-input>
        <a-select v-model:value="status" class="status-select" :options="statusOptions" @change="handleSearch" />
        <a-button @click="resetFilters">
          清空
        </a-button>
      </div>
    </div>

    <a-table
      row-key="id"
      size="middle"
      :columns="columns"
      :data-source="totalRows"
      :loading="loading"
      :pagination="pagination"
      :scroll="{ x: 1080 }"
      @change="handleTableChange"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'question'">
          <a-tooltip :title="questionText(asRuleRecord(record))">
            <span>{{ questionText(asRuleRecord(record)) }}</span>
          </a-tooltip>
        </template>
        <template v-else-if="column.key === 'scoreValue'">
          <div class="score-option-cell">
            <a-tag color="blue" class="score-tag">{{ scoreBriefText(asRuleRecord(record).scoreValue) }}</a-tag>
            <a-tooltip
              :title="scoreFullText(asRuleRecord(record).scoreValue, asRuleRecord(record))"
              :overlay-style="{ maxWidth: '460px', whiteSpace: 'normal' }"
            >
              <InfoCircleOutlined class="score-info-icon" />
            </a-tooltip>
          </div>
        </template>
        <template v-else-if="column.key === 'longGoal'">
          <a-tooltip :title="longGoalText(asRuleRecord(record))" :overlay-style="{ maxWidth: '520px', whiteSpace: 'normal' }">
            <span :class="{ muted: !firstLongGoal(asRuleRecord(record)) }">{{ longGoalText(asRuleRecord(record)) }}</span>
          </a-tooltip>
        </template>
        <template v-else-if="column.key === 'status'">
          <a-tag :color="statusColor(record.status)">{{ statusText(record.status) }}</a-tag>
        </template>
        <template v-else-if="column.key === 'action'">
          <a-space>
            <a @click="openEdit(asRuleRecord(record))">编辑</a>
            <a @click="openShortGoalDrawer(asRuleRecord(record))">关联短期目标</a>
            <a class="danger-link" @click="confirmDeleteRule(asRuleRecord(record))">删除</a>
          </a-space>
        </template>
      </template>
    </a-table>

    <a-drawer
      v-model:open="materialDrawerOpen"
      width="820"
      :body-style="{ padding: '0', background: '#f7f7fd' }"
      :closable="false"
      :push="{ distance: 80 }"
      placement="right"
      destroy-on-close
    >
      <template #title>
        <div class="custom-header flex justify-between h-4 flex-items-center">
          <div class="text-5">
            {{ drawerTitle }}
          </div>
          <a-button type="text" class="close-btn" @click="materialDrawerOpen = false">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>
      <div class="contenter pep3-drawer-content">
        <a-form layout="vertical">
          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item label="领域" required>
                <a-select
                  v-model:value="ruleForm.domainCode"
                  allow-clear
                  show-search
                  option-filter-prop="label"
                  :loading="questionLoading"
                  :options="domainOptions"
                  placeholder="请选择领域"
                  @change="applyDomainToRule"
                />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="状态">
                <a-select v-model:value="ruleForm.status" :options="materialStatusOptions" placeholder="请选择状态" />
              </a-form-item>
            </a-col>
          </a-row>
          <a-form-item label="题目" required>
            <a-select
              v-model:value="ruleForm.itemNo"
              allow-clear
              show-search
              option-filter-prop="label"
              :disabled="!ruleForm.domainCode"
              :loading="questionLoading"
              :options="filteredQuestionOptions"
              placeholder="选择领域后再选择题目"
              @change="applyQuestionToRule"
            />
          </a-form-item>
          <a-form-item label="选项" required>
            <a-select v-model:value="ruleForm.scoreValue" :options="scoreOptions" placeholder="请选择选项" @change="applyScoreToRule" />
          </a-form-item>
          <a-form-item required>
            <template #label>
              <div class="form-label-action">
                <span>长期目标</span>
                <a-button
                  type="link"
                  size="small"
                  class="ai-generate-btn"
                  :loading="aiGenerating.longGoal"
                  @click="generateLongGoal"
                >
                  <template #icon>
                    <ThunderboltOutlined />
                  </template>
                  AI生成
                </a-button>
              </div>
            </template>
            <a-textarea v-model:value="ruleForm.longGoal" :rows="5" placeholder="填写该题目选项对应的长期目标" />
          </a-form-item>
        </a-form>
      </div>

      <template #footer>
        <div class="drawer-footer">
          <span class="scope-hint">平台标准素材</span>
          <a-space>
            <a-button @click="materialDrawerOpen = false">
              取消
            </a-button>
            <a-button type="primary" :loading="saving" @click="saveRule">
              保存
            </a-button>
          </a-space>
        </div>
      </template>
    </a-drawer>

    <a-drawer
      v-model:open="shortGoalDrawerOpen"
      width="980"
      :body-style="{ padding: '0', background: '#f7f7fd' }"
      :closable="false"
      :push="{ distance: 80 }"
      placement="right"
      destroy-on-close
    >
      <template #title>
        <div class="custom-header flex justify-between h-4 flex-items-center">
          <div class="text-5">
            关联短期目标和训练内容
          </div>
          <a-button type="text" class="close-btn" @click="shortGoalDrawerOpen = false">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>
      <div class="contenter pep3-drawer-content">
        <div class="long-goal-panel">
          <span>{{ activeRule?.domain || '未分领域' }}</span>
          <strong>{{ activeLongGoal?.longGoal || '未关联长期目标' }}</strong>
        </div>

        <div class="section-head">
          <div class="section-title">
            短期目标
          </div>
          <a-space>
            <a-button size="small" :icon="h(ThunderboltOutlined)" @click="openBatchShortGoalModal">
              AI批量生成
            </a-button>
            <a-button type="primary" size="small" :icon="h(PlusOutlined)" @click="openCreateShortGoalModal">
              新增短期目标
            </a-button>
          </a-space>
        </div>
        <a-table
          row-key="id"
          size="small"
          :columns="shortGoalColumns"
          :data-source="shortGoalRows"
          :loading="shortGoalLoading"
          :pagination="false"
          :scroll="{ x: 760 }"
          :custom-row="shortGoalRowProps"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'shortGoal'">
              <span>{{ record.shortGoal }}</span>
            </template>
            <template v-else-if="column.key === 'status'">
              <a-tag :color="statusColor(record.status)">{{ statusText(record.status) }}</a-tag>
            </template>
            <template v-else-if="column.key === 'action'">
              <a-space>
                <a @click.stop="openEditShortGoalModal(asGoalRecord(record))">编辑</a>
                <a class="danger-link" @click.stop="confirmDeleteShortGoal(asGoalRecord(record))">删除</a>
              </a-space>
            </template>
          </template>
        </a-table>

        <div class="training-section">
          <div class="section-head">
            <div class="section-title">
              训练内容
            </div>
            <a-space v-if="activeShortGoal">
              <a-button size="small" :icon="h(ThunderboltOutlined)" @click="openBatchTrainingModal">
                AI批量生成
              </a-button>
              <a-button
                type="primary"
                size="small"
                :icon="h(PlusOutlined)"
                @click="openCreateTrainingModal()"
              >
                新增训练内容
              </a-button>
            </a-space>
          </div>
          <a-alert
            v-if="!activeShortGoal"
            type="info"
            show-icon
            message="请先选择或保存一个短期目标，再维护训练内容。"
          />
          <template v-else>
            <div class="short-goal-panel">
              <span>当前短期目标</span>
              <strong>{{ activeShortGoal.shortGoal }}</strong>
            </div>
            <a-table
              row-key="id"
              size="small"
              :columns="trainingColumns"
              :data-source="trainingRows"
              :loading="trainingLoading"
              :pagination="false"
              :scroll="{ x: 760 }"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'status'">
                  <a-tag :color="statusColor(record.status)">{{ statusText(record.status) }}</a-tag>
                </template>
                <template v-else-if="column.key === 'action'">
                  <a-space>
                    <a @click="openEditTrainingModal(asTrainingRecord(record))">编辑</a>
                    <a class="danger-link" @click="confirmDeleteTraining(asTrainingRecord(record))">删除</a>
                  </a-space>
                </template>
              </template>
            </a-table>
          </template>
        </div>
      </div>

      <template #footer>
        <div class="drawer-footer">
          <span class="scope-hint">短期目标和训练内容都归属当前长期目标</span>
          <a-button @click="shortGoalDrawerOpen = false">
            关闭
          </a-button>
        </div>
      </template>
    </a-drawer>

    <PlatformModalShell
      v-model:open="shortGoalModalOpen"
      :title="shortGoalModalTitle"
      :width="720"
      scrollable
    >
      <a-form layout="vertical" class="modal-form">
        <a-form-item required>
          <template #label>
            <div class="form-label-action">
              <span>短期目标</span>
              <a-button
                type="link"
                size="small"
                class="ai-generate-btn"
                :loading="aiGenerating.shortGoal"
                @click="generateShortGoal"
              >
                <template #icon>
                  <ThunderboltOutlined />
                </template>
                AI生成
              </a-button>
            </div>
          </template>
          <a-textarea v-model:value="shortGoalForm.shortGoal" :rows="4" placeholder="请输入该长期目标下的短期目标" />
        </a-form-item>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="课程形式" required>
              <a-select v-model:value="shortGoalForm.courseForm" allow-clear :options="courseFormOptions" placeholder="请选择课程形式" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="状态">
              <a-select v-model:value="shortGoalForm.status" :options="materialStatusOptions" placeholder="请选择状态" />
            </a-form-item>
          </a-col>
        </a-row>
      </a-form>
      <template #footer>
        <div class="modal-footer">
          <a-button @click="shortGoalModalOpen = false">
            取消
          </a-button>
          <a-button type="primary" :loading="saving" @click="saveShortGoal">
            保存
          </a-button>
        </div>
      </template>
    </PlatformModalShell>

    <PlatformModalShell
      v-model:open="trainingModalOpen"
      :title="trainingModalTitle"
      :width="720"
      scrollable
    >
      <a-form layout="vertical" class="modal-form">
        <a-form-item label="所属短期目标">
          <div class="readonly-goal">
            {{ activeShortGoal?.shortGoal || '-' }}
          </div>
        </a-form-item>
        <a-form-item label="训练项目" required>
          <a-input v-model:value="trainingForm.trainingProject" placeholder="请输入训练项目，例如：平衡木行走" />
        </a-form-item>
        <a-form-item required>
          <template #label>
            <div class="form-label-action">
              <span>训练内容</span>
              <a-button
                type="link"
                size="small"
                class="ai-generate-btn"
                :loading="aiGenerating.training"
                @click="generateTraining"
              >
                <template #icon>
                  <ThunderboltOutlined />
                </template>
                AI生成
              </a-button>
            </div>
          </template>
          <a-textarea v-model:value="trainingForm.trainingContent" :rows="4" placeholder="请输入简明训练活动，例如：准备材料、教师提示、儿童操作和泛化方式" />
        </a-form-item>
        <a-form-item label="状态">
          <a-select v-model:value="trainingForm.status" :options="materialStatusOptions" placeholder="请选择状态" />
        </a-form-item>
      </a-form>
      <template #footer>
        <div class="modal-footer">
          <a-button @click="trainingModalOpen = false">
            取消
          </a-button>
          <a-button type="primary" :loading="saving" @click="saveTraining">
            保存
          </a-button>
        </div>
      </template>
    </PlatformModalShell>

    <PlatformModalShell
      v-model:open="aiBatchModalOpen"
      :title="aiBatchModalTitle"
      :width="960"
      scrollable
    >
      <div class="batch-generate-box">
        <a-alert type="info" show-icon :message="aiBatchHelpText" />
        <div class="batch-toolbar">
          <div class="batch-count">
            <span>生成数量</span>
            <a-input-number v-model:value="aiBatchCount" :min="1" :max="10" />
          </div>
          <a-button type="primary" :loading="aiBatchLoading" :icon="h(ThunderboltOutlined)" @click="runAIBatchGenerate">
            开始生成
          </a-button>
        </div>
        <a-alert
          v-if="aiBatchLastError"
          type="warning"
          show-icon
          :message="`部分生成未通过校验：${aiBatchLastError}`"
        />
        <a-table
          row-key="id"
          size="small"
          class="batch-preview-table"
          :columns="aiBatchColumns"
          :data-source="aiBatchRows"
          :loading="aiBatchLoading"
          :pagination="false"
          :row-selection="aiBatchRowSelection"
          :scroll="{ x: 820 }"
        >
          <template #emptyText>
            设置数量后点击开始生成，结果会显示在这里。
          </template>
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'shortGoal'">
              <a-textarea v-model:value="record.shortGoal" :rows="3" placeholder="短期目标" />
            </template>
            <template v-else-if="column.key === 'courseForm'">
              <a-select v-model:value="record.courseForm" class="full" :options="courseFormOptions" placeholder="课程形式" />
            </template>
            <template v-else-if="column.key === 'trainingProject'">
              <a-input v-model:value="record.trainingProject" placeholder="训练项目" />
            </template>
            <template v-else-if="column.key === 'trainingContent'">
              <a-textarea v-model:value="record.trainingContent" :rows="3" placeholder="训练内容" />
            </template>
          </template>
        </a-table>
      </div>
      <template #footer>
        <div class="modal-footer batch-footer">
          <span>已选择 {{ aiBatchSelectedCount }} 条</span>
          <a-space>
            <a-button @click="aiBatchModalOpen = false">
              取消
            </a-button>
            <a-button type="primary" :disabled="!aiBatchSelectedCount" :loading="aiBatchSaving" @click="saveAIBatchRows">
              保存选中
            </a-button>
          </a-space>
        </div>
      </template>
    </PlatformModalShell>
  </div>
</template>

<style scoped>

.page-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 8px;
}

.page-head h2 {
  margin: 0;
  color: #1f2937;
  font-size: 22px;
  font-weight: 650;
}

.page-head p {
  margin: 0;
  color: #6b7280;
  font-size: 14px;
}

.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 14px 16px;
  margin-bottom: 14px;
  background: #fff;
  border: 1px solid #eef0f4;
  border-radius: 8px;
}

.toolbar-title {
  color: #1f2937;
  font-size: 15px;
  font-weight: 650;
}

.filters {
  display: flex;
  align-items: center;
  gap: 10px;
}

.keyword {
  width: 360px;
}

.status-select {
  width: 150px;
}

.score-tag {
  margin: 0;
}

.score-option-cell {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  max-width: 100%;
}

.score-info-icon {
  flex: 0 0 auto;
  color: #8c8c8c;
  font-size: 15px;
  cursor: help;
}

.score-info-icon:hover {
  color: #1677ff;
}

.muted {
  color: #9ca3af;
}

.danger-link {
  color: #ef4444;
}

.pep3-drawer-content {
  padding: 24px;
}

.close-btn:hover {
  background: transparent;
}

.close-btn:hover .close-icon {
  animation: icon-rotate 0.3s linear;
}

.long-goal-panel,
.short-goal-panel {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px 14px;
  margin-bottom: 16px;
  background: #f8fafc;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
}

.long-goal-panel span,
.short-goal-panel span {
  color: #6b7280;
  font-size: 13px;
}

.long-goal-panel strong,
.short-goal-panel strong {
  color: #1f2937;
  font-size: 14px;
  line-height: 1.6;
}

.section-title {
  margin: 0 0 12px;
  color: #111827;
  font-size: 15px;
  font-weight: 650;
}

.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.section-head .section-title {
  margin-bottom: 0;
}

.training-section {
  padding-top: 20px;
  margin-top: 20px;
  border-top: 1px solid #eef0f4;
}

.drawer-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.modal-form {
  padding-top: 4px;
}

.form-label-action {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  gap: 12px;
}

.ai-generate-btn {
  height: 22px;
  padding: 0;
  font-size: 13px;
  line-height: 22px;
}

.ai-generate-btn :deep(.anticon) {
  font-size: 13px;
}

.readonly-goal {
  min-height: 38px;
  padding: 8px 11px;
  color: #374151;
  line-height: 1.6;
  background: #f9fafb;
  border: 1px solid #d9d9d9;
  border-radius: 6px;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  width: 100%;
}

.batch-generate-box {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding-top: 4px;
}

.batch-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 14px;
  background: #f9fafb;
  border: 1px solid #eef0f4;
  border-radius: 8px;
}

.batch-count {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  color: #374151;
  font-size: 14px;
}

.batch-preview-table :deep(.ant-table-cell) {
  vertical-align: top;
}

.batch-footer {
  align-items: center;
  justify-content: space-between;
  color: #6b7280;
}

.scope-hint {
  color: #8c8c8c;
  font-size: 13px;
}

.full {
  width: 100%;
}

@keyframes icon-rotate {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(180deg);
  }
}

@media (max-width: 900px) {
  .page-head,
  .toolbar {
    align-items: stretch;
    flex-direction: column;
  }

  .filters {
    align-items: stretch;
    flex-direction: column;
  }

  .keyword,
  .status-select {
    width: 100%;
  }
}
</style>
