<script setup lang="ts">
import {
  ArrowLeftOutlined,
  CheckCircleFilled,
  FileDoneOutlined,
  FileTextOutlined,
  LeftOutlined,
  MessageOutlined,
  PlayCircleOutlined,
  RightOutlined,
  SaveOutlined,
  SlidersOutlined,
  SwapOutlined,
  WechatOutlined,
} from '@ant-design/icons-vue'
import QRCode from 'qrcode'
import dayjs from 'dayjs'
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { Modal } from 'ant-design-vue'
import { useRoute, useRouter } from 'vue-router'
import {
  getPEP3AssessmentDraftDetailApi,
  getPEP3AssessmentFormTemplateItemApi,
  getPEP3AssessmentFormTemplateSummaryApi,
  getPEP3AssessmentRecordDetailApi,
  invitePEP3CaregiverReportApi,
  pagePEP3AssessmentDraftsApi,
  pagePEP3AssessmentRecordsApi,
  savePEP3AssessmentDraftApi,
  savePEP3AssessmentDraftItemApi,
  submitPEP3AssessmentDraftApi,
  type PEP3AssessmentDraftDetail,
  type PEP3AssessmentDraftSummary,
  type PEP3AssessmentFormTemplateSummary,
  type PEP3AssessmentItem,
  type PEP3AssessmentItemGroupSummary,
  type PEP3AssessmentItemSummary,
  type PEP3AssessmentRecordDetail,
  type PEP3CaregiverReportInvite,
  type PEP3DraftSaveRequest,
  type PEP3ItemRecordField,
  type PEP3ScaleCode,
  type PEP3ScoreOption,
} from '@/api/edu-center/pep3-assessment'
import { getScaleAssessmentStudentCandidatesApi } from '@/api/teacher-center/scale-library'
import messageService from '@/utils/messageService'

type DraftItemSaveStatus = 'queued' | 'saving' | 'saved' | 'error'
type ScoreTone = 'green' | 'blue' | 'red'

interface WorkbenchScoreOption {
  value: number
  title: string
  desc: string
  standardText: string
  tone: ScoreTone
  checkColor: string
}

const route = useRoute()
const router = useRouter()

const templateLoading = ref(false)
const currentItemLoading = ref(false)
const saving = ref(false)
const submitting = ref(false)
const template = ref<PEP3AssessmentFormTemplateSummary>()
const itemCache = reactive<Record<number, PEP3AssessmentItem>>({})
const draft = ref<PEP3AssessmentDraftDetail>()
const currentProgress = ref<PEP3AssessmentDraftDetail['progress']>()
const currentItemNo = ref(numberFromQuery('itemNo') || 0)
const autoNext = ref(true)
const caregiverQRCodeDataUrl = ref('')
const caregiverInviteLoading = ref(false)
const caregiverInvite = ref<PEP3CaregiverReportInvite>()
const draftResumeModalOpen = ref(false)
const existingDraft = ref<PEP3AssessmentDraftSummary>()
const guidanceVideoOpen = ref(false)
const materialPreviewOpen = ref(false)
const previousScoreDate = ref('')
const previousItemScores = reactive<Record<number, number>>({})
const expandedGroupKeys = ref<string[]>([])
const draftItemSaveStatus = ref<Record<number, DraftItemSaveStatus>>({})
const draftItemSaveErrors = ref<Record<number, string>>({})
const draftItemSaveTimers = new Map<number, number>()
const draftItemSaveSeq = new Map<number, number>()
const draftItemSaveInFlight = new Set<number>()
const autoSaveLastSavedAt = ref('')
let draftCreationPromise: Promise<PEP3AssessmentDraftDetail | undefined> | undefined

const editor = reactive<{
  id?: number
  studentId?: number
  studentName: string
  examinerName: string
  birthDate?: string
  assessmentDate?: string
  remark: string
  allowMissingItems: boolean
  itemScores: Record<number, number | undefined>
  rawScores: Record<string, number | undefined>
  itemRecordValues: Record<number, Record<string, unknown>>
  caregiverReport?: PEP3DraftSaveRequest['caregiverReport']
}>({
  id: numberFromQuery('draftId') || undefined,
  studentId: numberFromQuery('childId') || undefined,
  studentName: textFromQuery('childName'),
  examinerName: textFromQuery('examinerName'),
  birthDate: normalizeDateText(textFromQuery('childBirthDate') || textFromQuery('birthDate')),
  assessmentDate: normalizeDateText(textFromQuery('assessmentDate')) || dayjs().format('YYYY-MM-DD'),
  remark: '',
  allowMissingItems: true,
  itemScores: {},
  rawScores: {},
  itemRecordValues: {},
})

const studentName = computed(() => editor.studentName || textFromQuery('childName') || '-')
const studentAge = computed(() => textFromQuery('childAge') || '-')
const scaleTitle = computed(() => {
  const name = textFromQuery('scaleName')
  if (name && /pep[-\s]?3/i.test(name))
    return 'PEP-3'
  if (name)
    return name
  const code = String(route.query.scaleCode || 'PEP3').trim()
  if (code.toUpperCase() === 'PEP3')
    return 'PEP-3'
  return code || 'PEP-3'
})
const assessmentDateText = computed(() => formatDate(editor.assessmentDate))
const examinerName = computed(() => draft.value?.examinerName || editor.examinerName || '当前老师')
const allItems = computed(() => template.value?.itemGroups?.flatMap(group => group.items || []) || [])
const totalItemCount = computed(() => template.value?.itemCount || allItems.value.length)
const currentIndex = computed(() => {
  const index = allItems.value.findIndex(item => item.itemNo === currentItemNo.value)
  return index >= 0 ? index : 0
})
const currentItemSummary = computed(() => allItems.value[currentIndex.value])
const currentItem = computed(() => itemCache[currentItemNo.value])
const currentItemTitle = computed(() => currentItem.value || currentItemSummary.value ? displayItemTitle(currentItem.value || currentItemSummary.value) : '-')
const currentDomainCode = computed(() => currentItem.value?.domainCode || currentItemSummary.value?.domainCode || '-')
const currentDomainName = computed(() => currentItem.value?.domainName || currentItemSummary.value?.domainName || '-')
const currentMaterialImageUrl = computed(() => firstNonEmpty(currentItem.value?.materialImages))
const guidanceVideoUrl = computed(() => normalizeText(currentItem.value?.guidanceVideo, ''))
const currentRecordFields = computed(() => currentItem.value?.recordFields || [])
const answeredItemCount = computed(() => Object.values(editor.itemScores).filter(isValidScore).length)
const missingItemCount = computed(() => Math.max(totalItemCount.value - answeredItemCount.value, 0))
const progressPercent = computed(() => totalItemCount.value ? Math.round((answeredItemCount.value / totalItemCount.value) * 100) : 0)
const donutStyle = computed(() => ({
  background: `radial-gradient(circle at center, #fff 54%, transparent 55%), conic-gradient(#2563eb 0 ${progressPercent.value}%, #e5e7eb ${progressPercent.value}% 100%)`,
}))
const selectedScore = computed(() => {
  const score = editor.itemScores[currentItemNo.value]
  return isValidScore(score) ? Number(score) : undefined
})
const previousScore = computed(() => {
  const score = previousItemScores[currentItemNo.value]
  return isValidScore(score) ? Number(score) : undefined
})
const currentScoreOptions = computed<WorkbenchScoreOption[]>(() => {
  const options = currentItem.value?.scoreOptions?.length
    ? currentItem.value.scoreOptions
    : template.value?.scoreOptions || []
  return options
    .map(toWorkbenchScoreOption)
    .sort((a, b) => b.value - a.value)
})
const previousScoreOption = computed(() => {
  if (previousScore.value === undefined)
    return undefined
  return currentScoreOptions.value.find(item => item.value === previousScore.value)
})
const autoSaveState = computed<'idle' | 'saving' | 'saved'>(() => {
  const statuses = Object.values(draftItemSaveStatus.value)
  if (statuses.some(status => status === 'queued' || status === 'saving'))
    return 'saving'
  return autoSaveLastSavedAt.value ? 'saved' : 'idle'
})
const autoSaveText = computed(() => {
  if (autoSaveState.value === 'saving')
    return '自动保存中...'
  if (autoSaveState.value === 'saved')
    return `已自动保存为草稿 ${autoSaveLastSavedAt.value}`
  return ''
})
const currentDisplayIndex = computed(() => totalItemCount.value ? currentIndex.value + 1 : 0)
const hasPreviousItem = computed(() => currentIndex.value > 0)
const hasNextItem = computed(() => currentIndex.value < allItems.value.length - 1)
const pageGroups = computed(() => {
  return (template.value?.itemGroups || []).map((group) => {
    const key = groupKey(group)
    const items = group.items || []
    const doneCount = items.filter(item => isValidScore(editor.itemScores[item.itemNo])).length
    const percent = items.length ? Math.round((doneCount / items.length) * 100) : 0
    return {
      key,
      title: groupTitle(group),
      count: `${doneCount}/${items.length} 题`,
      percent,
      expanded: expandedGroupKeys.value.includes(key),
      items: items.map(item => ({
        no: item.itemNo,
        name: displayItemTitle(item),
        status: item.itemNo === currentItemNo.value ? 'active' : isValidScore(editor.itemScores[item.itemNo]) ? 'done' : 'todo',
      })),
    }
  })
})

watch(allItems, (items) => {
  if (!items.length)
    return
  if (!items.some(item => item.itemNo === currentItemNo.value))
    currentItemNo.value = items[0].itemNo
  expandCurrentGroup()
}, { immediate: true })

watch(currentItemNo, () => {
  guidanceVideoOpen.value = false
  materialPreviewOpen.value = false
  expandCurrentGroup()
  void fetchItemDetail(currentItemNo.value)
})

onMounted(() => {
  void initializeWorkbench()
})

onBeforeUnmount(() => {
  flushDraftItemSaves()
  clearDraftItemSaveTimers()
})

function unwrap<T>(res: any): T {
  return (res?.data ?? res?.result ?? res) as T
}

function getErrorMessage(error: any, fallback: string) {
  return error?.response?.data?.message || error?.message || fallback
}

function textFromQuery(key: string) {
  const raw = route.query[key]
  const value = Array.isArray(raw) ? raw[0] : raw
  return String(value || '').trim()
}

function numberFromQuery(key: string) {
  const value = Number(textFromQuery(key))
  return Number.isFinite(value) && value > 0 ? value : 0
}

function normalizeDateText(value?: string) {
  const text = String(value || '').trim()
  if (!text)
    return undefined
  const date = dayjs(text)
  return date.isValid() ? date.format('YYYY-MM-DD') : undefined
}

function formatDate(value?: string) {
  const text = String(value || '').trim()
  if (!text)
    return '-'
  const date = dayjs(text)
  return date.isValid() ? date.format('YYYY-MM-DD') : text
}

function formatDateTime(value?: string) {
  const text = String(value || '').trim()
  if (!text)
    return '-'
  const date = dayjs(text)
  return date.isValid() ? date.format('YYYY-MM-DD HH:mm') : text
}

function normalizeText(value?: string, fallback = '-') {
  const text = String(value || '').replace(/\s+/g, ' ').trim()
  return text || fallback
}

function firstNonEmpty(values?: string[]) {
  return (values || []).map(item => String(item || '').trim()).find(Boolean) || ''
}

function displayItemTitle(item: PEP3AssessmentItem | PEP3AssessmentItemSummary) {
  return normalizeText(item.itemTitle || item.testItem)
    .replace(/^[（(]?\d+[）)]?\s*[、.．-]?\s*/, '')
    .trim() || normalizeText(item.testItem)
}

function groupKey(group: PEP3AssessmentItemGroupSummary) {
  return group.groupCode || `booklet_page_${group.bookletPageNo}`
}

function groupTitle(group: PEP3AssessmentItemGroupSummary) {
  if (group.startItemNo && group.endItemNo)
    return `第 ${group.bookletPageNo} 页  ${group.startItemNo}-${group.endItemNo}题`
  return group.title || `第 ${group.bookletPageNo} 页`
}

function isValidScore(score: unknown) {
  return Number(score) === 0 || Number(score) === 1 || Number(score) === 2
}

function scoreTone(value: number): ScoreTone {
  if (value === 2)
    return 'green'
  if (value === 0)
    return 'red'
  return 'blue'
}

function scoreCheckColor(value: number) {
  if (value === 2)
    return '#0d9749'
  if (value === 0)
    return '#d41f1f'
  return '#0757e6'
}

function shortScoreLabel(value: number) {
  if (value === 2)
    return '通过'
  if (value === 1)
    return '部分通过'
  if (value === 0)
    return '未通过'
  return ''
}

function toWorkbenchScoreOption(option: PEP3ScoreOption): WorkbenchScoreOption {
  const value = Number(option.value)
  return {
    value,
    title: `${value} 分`,
    desc: shortScoreLabel(value) || normalizeText(option.label, ''),
    standardText: normalizeText(option.description || option.label),
    tone: scoreTone(value),
    checkColor: scoreCheckColor(value),
  }
}

function expandCurrentGroup() {
  const group = template.value?.itemGroups?.find(itemGroup => (itemGroup.items || []).some(item => item.itemNo === currentItemNo.value))
  if (!group)
    return
  const key = groupKey(group)
  if (!expandedGroupKeys.value.includes(key))
    expandedGroupKeys.value = [key]
}

function toggleGroup(key: string) {
  expandedGroupKeys.value = expandedGroupKeys.value.includes(key)
    ? expandedGroupKeys.value.filter(item => item !== key)
    : [...expandedGroupKeys.value, key]
}

async function initializeWorkbench() {
  await fetchTemplate()
  if (editor.id) {
    await fetchDraftDetail(editor.id)
    await hydrateStudentBirthDate()
    await fetchPreviousAssessment()
    if (validateDraftHeader(true))
      await refreshCaregiverInvite(true)
    return
  }

  await hydrateStudentBirthDate()
  const draftToResume = await findExistingDraft()
  if (draftToResume) {
    existingDraft.value = draftToResume
    draftResumeModalOpen.value = true
    return
  }

  await startNewAssessment()
}

async function startNewAssessment() {
  await fetchPreviousAssessment()
  if (validateDraftHeader(true)) {
    const detail = await saveDraft(true)
    if (detail?.id)
      await refreshCaregiverInvite(true)
  }
}

async function findExistingDraft() {
  if (!editor.studentId)
    return undefined
  try {
    const res = await pagePEP3AssessmentDraftsApi({
      pageRequestModel: { pageIndex: 1, pageSize: 1 },
      queryModel: {
        assessmentCode: 'PEP3',
        studentId: editor.studentId,
      },
    })
    const data = unwrap<any>(res)
    return (data?.items || [])[0] as PEP3AssessmentDraftSummary | undefined
  }
  catch (error: any) {
    messageService.error(getErrorMessage(error, '查询未完成草稿失败'))
    return undefined
  }
}

async function continueExistingDraft() {
  const target = existingDraft.value
  draftResumeModalOpen.value = false
  if (!target?.id)
    return
  await fetchDraftDetail(Number(target.id))
  const missingItemNo = target.progress?.missingItemNos?.[0]
  if (missingItemNo && allItems.value.some(item => item.itemNo === missingItemNo))
    currentItemNo.value = missingItemNo
  await fetchPreviousAssessment()
  await refreshCaregiverInvite(true)
}

async function restartAssessment() {
  draftResumeModalOpen.value = false
  resetAssessmentInputForRestart()
  await startNewAssessment()
}

function resetAssessmentInputForRestart() {
  clearDraftItemSaveTimers()
  draftItemSaveSeq.clear()
  draftItemSaveInFlight.clear()
  draftItemSaveStatus.value = {}
  draftItemSaveErrors.value = {}
  autoSaveLastSavedAt.value = ''
  draft.value = undefined
  currentProgress.value = undefined
  caregiverInvite.value = undefined
  caregiverQRCodeDataUrl.value = ''
  editor.id = undefined
  editor.remark = ''
  editor.allowMissingItems = true
  editor.itemScores = {}
  editor.rawScores = {}
  editor.itemRecordValues = {}
  editor.caregiverReport = undefined
}

async function fetchTemplate() {
  templateLoading.value = true
  try {
    const res = await getPEP3AssessmentFormTemplateSummaryApi()
    template.value = unwrap<PEP3AssessmentFormTemplateSummary>(res)
    if (currentItemNo.value > 0 && allItems.value.some(item => item.itemNo === currentItemNo.value))
      await fetchItemDetail(currentItemNo.value)
  }
  catch (error: any) {
    messageService.error(getErrorMessage(error, '获取PEP-3题目目录失败'))
  }
  finally {
    templateLoading.value = false
  }
}

async function fetchItemDetail(itemNo: number) {
  if (itemNo <= 0 || itemCache[itemNo])
    return
  currentItemLoading.value = true
  try {
    const res = await getPEP3AssessmentFormTemplateItemApi(itemNo)
    itemCache[itemNo] = unwrap<PEP3AssessmentItem>(res)
  }
  catch (error: any) {
    messageService.error(getErrorMessage(error, `获取第${itemNo}题失败`))
  }
  finally {
    currentItemLoading.value = false
  }
}

async function fetchDraftDetail(id: number) {
  try {
    const res = await getPEP3AssessmentDraftDetailApi(id)
    const detail = unwrap<PEP3AssessmentDraftDetail>(res)
    draft.value = detail
    currentProgress.value = detail.progress
    editor.id = detail.id
    applyDraftInput(normalizeInputSnapshot(detail.input))
  }
  catch (error: any) {
    messageService.error(getErrorMessage(error, '获取测评草稿失败'))
  }
}

async function hydrateStudentBirthDate() {
  if (editor.birthDate || !editor.studentId)
    return
  try {
    const res = await getScaleAssessmentStudentCandidatesApi({
      scaleCode: textFromQuery('scaleCode') || 'PEP3',
      keyword: editor.studentName || undefined,
      pageIndex: 1,
      pageSize: 100,
    })
    const data = unwrap<any>(res)
    const rows = data?.items || []
    const student = rows.find((item: any) => Number(item.id) === Number(editor.studentId))
    if (student?.birthDate)
      editor.birthDate = normalizeDateText(student.birthDate)
  }
  catch {
    // 出生日期缺失时仍允许保存草稿，正式提交时后端会校验。
  }
}

async function fetchPreviousAssessment() {
  if (!editor.studentId)
    return
  try {
    const res = await pagePEP3AssessmentRecordsApi({
      pageRequestModel: { pageIndex: 1, pageSize: 1 },
      queryModel: {
        assessmentCode: 'PEP3',
        studentId: editor.studentId,
      },
    })
    const data = unwrap<any>(res)
    const latest = (data?.items || [])[0]
    if (!latest?.id)
      return
    previousScoreDate.value = formatDate(latest.assessmentDate)
    const detailRes = await getPEP3AssessmentRecordDetailApi(Number(latest.id))
    const detail = unwrap<PEP3AssessmentRecordDetail>(detailRes)
    applyPreviousInput(normalizeInputSnapshot(detail.input))
  }
  catch {
    previousScoreDate.value = ''
  }
}

function normalizeInputSnapshot(input: any): PEP3DraftSaveRequest | undefined {
  if (!input)
    return undefined
  if (typeof input === 'string') {
    try {
      return JSON.parse(input) as PEP3DraftSaveRequest
    }
    catch {
      return undefined
    }
  }
  return input as PEP3DraftSaveRequest
}

function applyDraftInput(input?: PEP3DraftSaveRequest) {
  if (!input)
    return
  editor.studentId = input.studentId || editor.studentId
  editor.studentName = input.studentName || editor.studentName
  editor.examinerName = input.examinerName || editor.examinerName
  editor.birthDate = normalizeDateText(input.birthDate) || editor.birthDate
  editor.assessmentDate = normalizeDateText(input.assessmentDate) || editor.assessmentDate
  editor.remark = input.remark || ''
  editor.allowMissingItems = input.allowMissingItems ?? true
  editor.itemScores = {}
  editor.rawScores = {}
  editor.itemRecordValues = {}
  editor.caregiverReport = input.caregiverReport
  Object.entries(input.itemScores || {}).forEach(([itemNo, score]) => {
    if (isValidScore(score))
      editor.itemScores[Number(itemNo)] = Number(score)
  })
  ;(input.itemScoreList || []).forEach((item) => {
    if (isValidScore(item.score))
      editor.itemScores[item.itemNo] = Number(item.score)
  })
  Object.entries(input.rawScores || {}).forEach(([scaleCode, score]) => {
    if (typeof score === 'number' && Number.isFinite(score))
      editor.rawScores[scaleCode] = score
  })
  ;(input.rawScoreList || []).forEach((item) => {
    editor.rawScores[item.scaleCode] = item.rawScore
  })
  Object.entries(input.itemRecordValues || {}).forEach(([itemNo, values]) => {
    editor.itemRecordValues[Number(itemNo)] = { ...(values || {}) }
  })
  ;(input.itemRecordValueList || []).forEach((item) => {
    if (!editor.itemRecordValues[item.itemNo])
      editor.itemRecordValues[item.itemNo] = {}
    editor.itemRecordValues[item.itemNo][item.fieldKey] = item.value
  })
}

function applyPreviousInput(input?: PEP3DraftSaveRequest) {
  Object.keys(previousItemScores).forEach(key => delete previousItemScores[Number(key)])
  if (!input)
    return
  Object.entries(input.itemScores || {}).forEach(([itemNo, score]) => {
    if (isValidScore(score))
      previousItemScores[Number(itemNo)] = Number(score)
  })
  ;(input.itemScoreList || []).forEach((item) => {
    if (isValidScore(item.score))
      previousItemScores[item.itemNo] = Number(item.score)
  })
}

function buildPayload(): PEP3DraftSaveRequest {
  const itemScoreList = Object.entries(editor.itemScores)
    .filter(([, score]) => isValidScore(score))
    .map(([itemNo, score]) => ({ itemNo: Number(itemNo), score: Number(score) }))
    .sort((a, b) => a.itemNo - b.itemNo)
  const rawScoreList = Object.entries(editor.rawScores)
    .filter(([, score]) => typeof score === 'number' && Number.isFinite(score))
    .map(([scaleCode, rawScore]) => ({ scaleCode: scaleCode as PEP3ScaleCode, rawScore: Number(rawScore) }))
    .sort((a, b) => a.scaleCode.localeCompare(b.scaleCode))
  const itemRecordValueList = Object.entries(editor.itemRecordValues)
    .flatMap(([itemNo, values]) => Object.entries(values || {})
      .filter(([, value]) => !isEmptyRecordValue(value))
      .map(([fieldKey, value]) => ({ itemNo: Number(itemNo), fieldKey, value })))
    .sort((a, b) => a.itemNo - b.itemNo || a.fieldKey.localeCompare(b.fieldKey))

  return {
    id: editor.id,
    studentId: editor.studentId,
    studentName: studentName.value === '-' ? undefined : studentName.value,
    examinerName: editor.examinerName,
    birthDate: editor.birthDate,
    assessmentDate: editor.assessmentDate,
    remark: editor.remark,
    allowMissingItems: editor.allowMissingItems,
    itemScoreList,
    rawScoreList,
    itemRecordValueList,
    caregiverReport: editor.caregiverReport,
  }
}

function validateDraftHeader(silent = false) {
  if (!editor.studentId || studentName.value === '-') {
    if (!silent)
      messageService.warning('缺少真实儿童，无法保存测评')
    return false
  }
  return true
}

async function saveDraft(silent = false) {
  if (!validateDraftHeader(silent))
    return undefined
  saving.value = true
  try {
    const res = await savePEP3AssessmentDraftApi(buildPayload())
    const detail = unwrap<PEP3AssessmentDraftDetail>(res)
    draft.value = detail
    currentProgress.value = detail.progress
    editor.id = detail.id
    editor.examinerName = detail.examinerName || editor.examinerName
    if (!silent)
      messageService.success('草稿已保存')
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
  flushDraftItemSaves()
  submitting.value = true
  try {
    const detail = await saveDraft(true)
    if (!detail?.id)
      return
    if (!canSubmitByRequiredFields(detail.progress)) {
      currentProgress.value = detail.progress
      messageService.warning(cannotSubmitMessage(detail.progress))
      return
    }
    if (!isAllQuestionsAnswered(detail.progress)) {
      currentProgress.value = detail.progress
      messageService.warning(incompleteItemsMessage(detail.progress))
      return
    }
    if (isCaregiverReportMissing(detail.progress)) {
      const confirmed = await confirmSubmitWithoutCaregiverReport()
      if (!confirmed)
        return
    }
    const res = await submitPEP3AssessmentDraftApi(detail.id)
    const result = unwrap<any>(res)
    messageService.success('已生成正式测评记录')
    if (result?.recordId)
      await router.push('/teacherCenter/evaluationRecord')
  }
  catch (error: any) {
    messageService.error(getErrorMessage(error, '提交正式记录失败'))
  }
  finally {
    submitting.value = false
  }
}

function canSubmitByRequiredFields(progress?: PEP3AssessmentDraftDetail['progress']) {
  const missing = progress?.missingRequiredFields || []
  return !missing.includes('birthDate') && !missing.includes('assessmentDate')
}

function isAllQuestionsAnswered(progress?: PEP3AssessmentDraftDetail['progress']) {
  const itemCount = progress?.itemCount || totalItemCount.value
  const answered = progress?.answeredItemCount ?? answeredItemCount.value
  return itemCount > 0 && answered >= itemCount && (progress?.missingItemCount ?? missingItemCount.value) <= 0
}

function isCaregiverReportMissing(progress?: PEP3AssessmentDraftDetail['progress']) {
  return (progress?.caregiverRawScoreCount || 0) < 3
}

function cannotSubmitMessage(progress?: PEP3AssessmentDraftDetail['progress']) {
  const missing = progress?.missingRequiredFields || []
  if (missing.includes('birthDate') || missing.includes('assessmentDate'))
    return '请补全出生日期和测评日期后再提交'
  if (!isAllQuestionsAnswered(progress))
    return incompleteItemsMessage(progress)
  return '当前草稿还不能提交，请补全必填内容'
}

function incompleteItemsMessage(progress?: PEP3AssessmentDraftDetail['progress']) {
  const itemCount = progress?.itemCount || totalItemCount.value
  const answered = progress?.answeredItemCount ?? answeredItemCount.value
  const missing = progress?.missingItemCount ?? Math.max(itemCount - answered, 0)
  return `还有 ${missing} 道题未评分，请完成全部 ${itemCount} 道题后再提交正式记录`
}

function confirmSubmitWithoutCaregiverReport() {
  return new Promise<boolean>((resolve) => {
    Modal.confirm({
      title: '照护者报告尚未提交',
      content: '当前测评题目已完成，但家长还没有提交照护者报告。可以先提交正式记录；家长后续提交后，系统会自动合并照护者报告数据并更新记录。是否继续提交？',
      okText: '先提交记录',
      cancelText: '暂不提交',
      centered: true,
      onOk: () => resolve(true),
      onCancel: () => resolve(false),
    })
  })
}

function setCurrentItemScore(value: number) {
  if (!currentItem.value)
    return
  setItemScoreValue(currentItem.value.itemNo, value)
  if (autoNext.value && hasNextItem.value) {
    window.setTimeout(() => {
      goNextItem()
    }, 180)
  }
}

function setItemScoreValue(itemNo: number, value: unknown) {
  const score = Number(value)
  if (isValidScore(score))
    editor.itemScores[itemNo] = score
  else
    delete editor.itemScores[itemNo]
  queueDraftItemSave(itemNo)
}

function draftItemRecordValues(itemNo: number) {
  return Object.entries(editor.itemRecordValues[itemNo] || {})
    .filter(([fieldKey, value]) => fieldKey.trim() && !isEmptyRecordValue(value))
    .reduce<Record<string, unknown>>((out, [fieldKey, value]) => {
      out[fieldKey.trim()] = value
      return out
    }, {})
}

function clearDraftItemSaveTimers() {
  draftItemSaveTimers.forEach(timer => window.clearTimeout(timer))
  draftItemSaveTimers.clear()
}

function queueDraftItemSave(itemNo: number) {
  if (itemNo <= 0)
    return
  draftItemSaveSeq.set(itemNo, (draftItemSaveSeq.get(itemNo) || 0) + 1)
  const existingTimer = draftItemSaveTimers.get(itemNo)
  if (existingTimer) {
    window.clearTimeout(existingTimer)
    draftItemSaveTimers.delete(itemNo)
  }
  setDraftItemSaveStatus(itemNo, 'saving')
  if (!draftItemSaveInFlight.has(itemNo))
    void saveDraftItem(itemNo)
}

async function ensureDraftForItemSave() {
  if (!validateDraftHeader(true))
    return false
  if (editor.id)
    return true
  if (!draftCreationPromise) {
    draftCreationPromise = saveDraft(true).finally(() => {
      draftCreationPromise = undefined
    })
  }
  const detail = await draftCreationPromise
  return !!detail?.id
}

async function saveDraftItem(itemNo: number) {
  if (itemNo <= 0)
    return
  if (draftItemSaveInFlight.has(itemNo))
    return

  draftItemSaveInFlight.add(itemNo)
  try {
    while (true) {
      const saveSeq = draftItemSaveSeq.get(itemNo) || 0
      setDraftItemSaveStatus(itemNo, 'saving')
      const canSave = await ensureDraftForItemSave()
      if (!canSave || !editor.id) {
        setDraftItemSaveStatus(itemNo, 'error', '草稿创建失败')
        return
      }
      const score = editor.itemScores[itemNo]
      const res = await savePEP3AssessmentDraftItemApi({
        draftId: editor.id,
        itemNo,
        score: isValidScore(score) ? Number(score) : undefined,
        recordValues: draftItemRecordValues(itemNo),
      })
      const detail = unwrap<PEP3AssessmentDraftDetail>(res)
      draft.value = detail
      currentProgress.value = detail.progress
      autoSaveLastSavedAt.value = dayjs().format('MM-DD HH:mm')
      if ((draftItemSaveSeq.get(itemNo) || 0) === saveSeq) {
        setDraftItemSaveStatus(itemNo, 'saved')
        return
      }
    }
  }
  catch (error: any) {
    const message = getErrorMessage(error, `第${itemNo}题自动保存失败`)
    setDraftItemSaveStatus(itemNo, 'error', message)
    messageService.error(message)
  }
  finally {
    draftItemSaveInFlight.delete(itemNo)
  }
}

function flushDraftItemSaves() {
  const pendingItemNos = Array.from(draftItemSaveTimers.keys())
  clearDraftItemSaveTimers()
  pendingItemNos.forEach(itemNo => void saveDraftItem(itemNo))
}

function setDraftItemSaveStatus(itemNo: number, status: DraftItemSaveStatus, error?: string) {
  draftItemSaveStatus.value = { ...draftItemSaveStatus.value, [itemNo]: status }
  const nextErrors = { ...draftItemSaveErrors.value }
  if (error)
    nextErrors[itemNo] = error
  else
    delete nextErrors[itemNo]
  draftItemSaveErrors.value = nextErrors
}

function getItemRecordValue(itemNo: number, fieldKey: string) {
  return editor.itemRecordValues[itemNo]?.[fieldKey]
}

function getCurrentRecordTextValue(fieldKey: string) {
  const value = getItemRecordValue(currentItemNo.value, fieldKey)
  return typeof value === 'string' || typeof value === 'number' ? String(value) : ''
}

function getCurrentRecordNumberValue(fieldKey: string) {
  const value = getItemRecordValue(currentItemNo.value, fieldKey)
  if (typeof value === 'number' && Number.isFinite(value))
    return value
  if (typeof value === 'string' && value.trim() !== '' && Number.isFinite(Number(value)))
    return Number(value)
  return undefined
}

function getCurrentRecordSelectValue(fieldKey: string) {
  const value = getItemRecordValue(currentItemNo.value, fieldKey)
  return typeof value === 'string' || typeof value === 'number' ? value : undefined
}

function getCurrentRecordArrayValue(fieldKey: string): string[] {
  const value = getItemRecordValue(currentItemNo.value, fieldKey)
  return Array.isArray(value) ? value.map(item => String(item)) : []
}

function recordFieldOptions(field: PEP3ItemRecordField) {
  return (field.options || []).map(option => ({
    label: option.label,
    value: option.value,
  }))
}

function setCurrentRecordValue(fieldKey: string, value: unknown) {
  if (!currentItem.value)
    return
  setItemRecordValue(currentItem.value.itemNo, fieldKey, value)
}

function setItemRecordValue(itemNo: number, fieldKey: string, value: unknown) {
  if (isEmptyRecordValue(value)) {
    clearItemRecordValue(itemNo, fieldKey)
    return
  }
  if (!editor.itemRecordValues[itemNo])
    editor.itemRecordValues[itemNo] = {}
  editor.itemRecordValues[itemNo][fieldKey] = value
  queueDraftItemSave(itemNo)
}

function clearItemRecordValue(itemNo: number, fieldKey: string) {
  if (!editor.itemRecordValues[itemNo]) {
    queueDraftItemSave(itemNo)
    return
  }
  delete editor.itemRecordValues[itemNo][fieldKey]
  if (!Object.keys(editor.itemRecordValues[itemNo]).length)
    delete editor.itemRecordValues[itemNo]
  queueDraftItemSave(itemNo)
}

function isEmptyRecordValue(value: unknown) {
  if (value === undefined || value === null)
    return true
  if (typeof value === 'string')
    return value.trim() === ''
  if (Array.isArray(value))
    return value.length === 0
  return false
}

async function refreshCaregiverInvite(silent = false) {
  caregiverInviteLoading.value = true
  try {
    const canSave = await ensureDraftForItemSave()
    if (!canSave || !editor.id)
      return
    const res = await invitePEP3CaregiverReportApi(editor.id)
    caregiverInvite.value = unwrap<PEP3CaregiverReportInvite>(res)
    await generateCaregiverQRCode(caregiverInvite.value)
    if (!silent)
      messageService.success('照护者报告入口已生成')
  }
  catch (error: any) {
    if (!silent)
      messageService.error(getErrorMessage(error, '生成照护者报告入口失败'))
  }
  finally {
    caregiverInviteLoading.value = false
  }
}

async function generateCaregiverQRCode(invite?: PEP3CaregiverReportInvite) {
  caregiverQRCodeDataUrl.value = ''
  if (invite?.miniProgramCodeDataUrl) {
    caregiverQRCodeDataUrl.value = invite.miniProgramCodeDataUrl
    return
  }
  const value = `${invite?.qrCodeValue || invite?.wechatUrlLink || invite?.miniProgramPath || invite?.url || ''}`.trim()
  if (!value)
    return
  caregiverQRCodeDataUrl.value = await QRCode.toDataURL(value, {
    width: 132,
    margin: 1,
    color: {
      dark: '#111827',
      light: '#FFFFFF',
    },
  })
}

async function handleCaregiverAction(type: 'sms' | 'wechat') {
  await refreshCaregiverInvite(true)
  if (!caregiverInvite.value) {
    messageService.warning('暂未生成照护者报告入口')
    return
  }
  messageService.success(type === 'sms' ? '已生成短信发送入口' : '已生成微信推送入口')
}

function goToItem(itemNo: number) {
  if (!allItems.value.some(item => item.itemNo === itemNo))
    return
  currentItemNo.value = itemNo
}

function goPreviousItem() {
  if (!hasPreviousItem.value)
    return
  currentItemNo.value = allItems.value[currentIndex.value - 1].itemNo
}

function goNextItem() {
  if (!hasNextItem.value)
    return
  currentItemNo.value = allItems.value[currentIndex.value + 1].itemNo
}

function jumpToMissingItem() {
  const missing = allItems.value.find(item => !isValidScore(editor.itemScores[item.itemNo]))
  if (!missing) {
    messageService.success('当前没有缺题')
    return
  }
  goToItem(missing.itemNo)
}

function goBack() {
  void router.push('/teacherCenter/scale-library')
}
</script>

<template>
  <div class="pep3-workbench-page">
    <header class="workbench-header">
      <button type="button" class="back-button" aria-label="返回量表库" @click="goBack">
        <ArrowLeftOutlined />
      </button>
      <strong class="workbench-title">{{ scaleTitle }} PC测评工作台</strong>
      <span class="header-divider"></span>
      <span class="header-meta">儿童：<b>{{ studentName }}</b></span>
      <span class="header-divider"></span>
      <span class="header-meta">年龄：<b>{{ studentAge }}</b></span>
      <span class="header-divider"></span>
      <span class="header-meta">测评日期：<b>{{ assessmentDateText }}</b></span>
      <span class="header-divider"></span>
      <span class="header-meta">施测者：<b>{{ examinerName }}</b></span>
      <div class="header-actions">
        <span
          v-if="autoSaveText"
          class="auto-save-status"
          :class="{ 'is-saving': autoSaveState === 'saving', 'is-saved': autoSaveState === 'saved' }"
        >
          {{ autoSaveText }}
        </span>
        <a-button size="large" class="outline-action" :loading="saving" @click="saveDraft(false)">
          <template #icon>
            <SaveOutlined />
          </template>
          保存草稿
        </a-button>
        <a-button size="large" type="primary" class="primary-action" :loading="submitting" @click="submitDraft">
          <template #icon>
            <FileDoneOutlined />
          </template>
          提交记录
        </a-button>
      </div>
    </header>

    <main class="workbench-main">
      <aside class="page-sidebar">
        <div class="sidebar-title">
          <span>记录册页面</span>
          <SlidersOutlined />
        </div>
        <div v-for="group in pageGroups" :key="group.key" class="page-group">
          <div class="page-group__head" @click="toggleGroup(group.key)">
            <span class="page-group__title">
              <RightOutlined v-if="!group.expanded" />
              <span v-else class="chevron-down">⌄</span>
              {{ group.title }}
            </span>
            <span>{{ group.count }}</span>
          </div>
          <div class="page-group__progress">
            <div class="progress-line">
              <i :style="{ width: `${group.percent}%` }"></i>
            </div>
            <span class="page-group__percent">{{ group.percent }}%</span>
          </div>

          <div v-if="group.expanded" class="question-list">
            <button
              v-for="item in group.items"
              :key="item.no"
              type="button"
              class="question-item"
              :class="`is-${item.status}`"
              @click="goToItem(item.no)"
            >
              <span>第 {{ item.no }} 题</span>
              <strong>{{ item.name }}</strong>
              <CheckCircleFilled v-if="item.status === 'done'" />
              <i v-else-if="item.status === 'active'"></i>
              <b v-else></b>
            </button>
          </div>
        </div>
      </aside>

      <section class="question-panel">
        <a-spin :spinning="templateLoading || currentItemLoading">
          <div class="question-title-row">
            <h1>第 {{ currentItemNo || '-' }} 题&nbsp;&nbsp;{{ currentItemTitle }}</h1>
            <a-tag color="blue">{{ currentDomainCode }} {{ currentDomainName }}</a-tag>
          </div>

          <article class="instruction-card material-card" :class="{ 'has-image': currentMaterialImageUrl }">
            <div class="material-card__text">
              <h2><FileTextOutlined />材料</h2>
              <p>{{ normalizeText(currentItem?.materials) }}</p>
            </div>
            <button v-if="currentMaterialImageUrl" type="button" class="material-card__image" @click="materialPreviewOpen = true">
              <img :src="currentMaterialImageUrl" alt="材料图片">
            </button>
          </article>

          <article class="instruction-card">
            <h2><FileTextOutlined />操作标准</h2>
            <p>{{ normalizeText(currentItem?.method) }}</p>
          </article>

          <article class="instruction-card guidance-card" :class="{ 'has-video': guidanceVideoUrl }">
            <div class="guidance-card__text">
              <h2><FileTextOutlined />指导语</h2>
              <p>{{ normalizeText(currentItem?.guidance) }}</p>
            </div>
            <button v-if="guidanceVideoUrl" type="button" class="guidance-video-entry" @click="guidanceVideoOpen = true">
              <span><PlayCircleOutlined /></span>
              <strong>指导视频</strong>
              <em>点击观看</em>
            </button>
          </article>

          <article class="instruction-card">
            <h2><FileTextOutlined />评分标准</h2>
            <p v-for="item in currentScoreOptions" :key="item.value">
              <b>{{ item.title }}（{{ item.desc }}）：</b> {{ item.standardText }}
            </p>
          </article>

          <div class="score-section">
            <div class="score-section__head">
              <h2>评分</h2>
              <div v-if="previousScoreOption" class="previous-score-summary" :class="`score-${previousScoreOption.tone}`">
                <span>上次测评 {{ previousScoreDate }}</span>
                <strong>{{ previousScoreOption.title }} · {{ previousScoreOption.desc }}</strong>
              </div>
            </div>
            <div class="score-options">
              <button
                v-for="item in currentScoreOptions"
                :key="item.value"
                type="button"
                class="score-option"
                :class="[`score-${item.tone}`, { 'is-selected': selectedScore === item.value, 'is-previous': previousScore === item.value }]"
                @click="setCurrentItemScore(item.value)"
              >
                <strong>{{ item.title }}</strong>
                <span>{{ item.desc }}</span>
                <em v-if="previousScore === item.value && previousScoreDate" class="score-option__previous-badge">
                  上次 {{ previousScoreDate.slice(5) }}
                </em>
                <CheckCircleFilled
                  v-if="selectedScore === item.value"
                  class="score-option__check"
                  :style="{ color: item.checkColor }"
                />
              </button>
            </div>
          </div>
        </a-spin>
      </section>

      <aside class="score-sidebar">
        <section class="right-card progress-card">
          <h2>当前进度</h2>
          <div class="progress-card__body">
            <div class="donut" :style="donutStyle">{{ progressPercent }}%</div>
            <div class="progress-stats">
              <span>已完成</span>
              <strong>{{ answeredItemCount }} <i>/ {{ totalItemCount }} 题</i></strong>
              <span>缺题</span>
              <strong class="danger">{{ missingItemCount }} <i>题</i></strong>
            </div>
          </div>
        </section>

        <section class="right-card training-record-card">
          <h2>儿童训练记录</h2>
          <div v-if="currentRecordFields.length" class="training-record-field">
            <div class="training-record-field__head">
              <span>{{ currentRecordFields[0].label }}</span>
              <a-tag class="training-record-field__type">{{ currentRecordFields[0].displayType || currentRecordFields[0].fieldType }}</a-tag>
            </div>
            <a-checkbox-group
              v-if="currentRecordFields[0].fieldType === 'checkbox_group'"
              :value="getCurrentRecordArrayValue(currentRecordFields[0].key)"
              class="training-record-checks"
              :options="recordFieldOptions(currentRecordFields[0])"
              @change="value => setCurrentRecordValue(currentRecordFields[0].key, value)"
            />
            <a-radio-group
              v-else-if="currentRecordFields[0].fieldType === 'radio'"
              :value="getCurrentRecordSelectValue(currentRecordFields[0].key)"
              :options="recordFieldOptions(currentRecordFields[0])"
              @change="event => setCurrentRecordValue(currentRecordFields[0].key, event.target.value)"
            />
            <a-input-number
              v-else-if="currentRecordFields[0].fieldType === 'number'"
              :value="getCurrentRecordNumberValue(currentRecordFields[0].key)"
              style="width: 100%"
              :placeholder="currentRecordFields[0].placeholder"
              @change="value => setCurrentRecordValue(currentRecordFields[0].key, value)"
            />
            <a-textarea
              v-else-if="currentRecordFields[0].fieldType === 'textarea'"
              :value="getCurrentRecordTextValue(currentRecordFields[0].key)"
              :placeholder="currentRecordFields[0].placeholder"
              :auto-size="{ minRows: 2, maxRows: 4 }"
              @change="event => setCurrentRecordValue(currentRecordFields[0].key, event.target.value)"
            />
            <a-input
              v-else
              :value="getCurrentRecordTextValue(currentRecordFields[0].key)"
              :placeholder="currentRecordFields[0].placeholder"
              @change="event => setCurrentRecordValue(currentRecordFields[0].key, event.target.value)"
            />
          </div>
          <div v-else class="training-record-empty">本题暂无训练记录项</div>
        </section>

        <section class="right-card caregiver-card">
          <h2>照护者报告</h2>
          <div class="caregiver-qrcode">
            <img v-if="caregiverQRCodeDataUrl" :src="caregiverQRCodeDataUrl" alt="照护者报告填写二维码">
            <a-spin v-else />
            <p>家长扫码填写照护者报告</p>
          </div>
          <div class="caregiver-actions">
            <a-button block class="caregiver-action" :loading="caregiverInviteLoading" @click="handleCaregiverAction('sms')">
              <template #icon>
                <MessageOutlined />
              </template>
              发送短信给家长
            </a-button>
            <a-button block type="primary" class="caregiver-action caregiver-action--wechat" :loading="caregiverInviteLoading" @click="handleCaregiverAction('wechat')">
              <template #icon>
                <WechatOutlined />
              </template>
              推送微信消息
            </a-button>
          </div>
        </section>
      </aside>
    </main>

    <footer class="workbench-footer">
      <a-button size="large" class="nav-button" :disabled="!hasPreviousItem" @click="goPreviousItem">
        <template #icon>
          <LeftOutlined />
        </template>
        上一题
      </a-button>
      <div class="question-counter">
        <strong>{{ currentDisplayIndex }}</strong>
        <span>/ {{ totalItemCount }}</span>
      </div>
      <a-button size="large" type="primary" class="next-button" :disabled="!hasNextItem" @click="goNextItem">
        下一题
        <RightOutlined />
      </a-button>
      <a-button size="large" class="nav-button" @click="jumpToMissingItem">
        <template #icon>
          <SwapOutlined />
        </template>
        跳到缺题
      </a-button>
      <div class="auto-next">
        <span>自动下一题</span>
        <a-switch v-model:checked="autoNext" />
      </div>
    </footer>

    <a-modal
      v-model:open="draftResumeModalOpen"
      title="发现未完成草稿"
      ok-text="继续测试"
      cancel-text="重新测试"
      :width="430"
      :closable="false"
      :mask-closable="false"
      centered
      @ok="continueExistingDraft"
      @cancel="restartAssessment"
    >
      <div class="draft-resume-tip">
        <p>当前儿童存在一份未提交的 PEP-3 测评草稿。</p>
        <div class="draft-resume-meta">
          <span>已完成：<b>{{ existingDraft?.answeredItemCount || 0 }}</b> / {{ existingDraft?.progress?.itemCount || totalItemCount }} 题</span>
          <span>更新时间：<b>{{ formatDateTime(existingDraft?.updatedTime) }}</b></span>
        </div>
      </div>
    </a-modal>

    <a-modal
      v-model:open="guidanceVideoOpen"
      title="指导视频"
      :footer="null"
      :width="720"
      centered
      destroy-on-close
    >
      <div class="guidance-video-player">
        <video v-if="guidanceVideoUrl" :src="guidanceVideoUrl" controls autoplay></video>
      </div>
    </a-modal>

    <a-modal
      v-model:open="materialPreviewOpen"
      title="材料图片"
      :footer="null"
      :width="520"
      centered
    >
      <div class="material-preview">
        <img v-if="currentMaterialImageUrl" :src="currentMaterialImageUrl" alt="材料图片">
      </div>
    </a-modal>
  </div>
</template>

<style scoped lang="less">
.pep3-workbench-page {
  display: flex;
  flex-direction: column;
  height: 100vh;
  min-height: 0;
  margin: 0;
  overflow: hidden;
  color: #1f2937;
  background: #f3f5f9;
}

.workbench-header {
  position: sticky;
  top: 0;
  z-index: 30;
  flex: 0 0 auto;
  display: flex;
  align-items: center;
  min-height: 52px;
  padding: 0 14px;
  background: rgba(255, 255, 255, 0.96);
  border: 1px solid #d8dfe8;
  border-radius: 0 0 10px 10px;
  box-shadow: 0 8px 22px rgba(15, 23, 42, 0.08);
  backdrop-filter: blur(8px);
}

.back-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  margin-right: 12px;
  color: #0f2a5f;
  background: transparent;
  border: 0;
  border-radius: 6px;
  cursor: pointer;
  font-size: 18px;

  &:hover {
    background: #eef4ff;
    color: #155bdc;
  }
}

.workbench-title {
  color: #0f2a5f;
  font-size: 18px;
  font-weight: 800;
  white-space: nowrap;
}

.header-divider {
  width: 1px;
  height: 18px;
  margin: 0 12px;
  background: #cbd5e1;
}

.header-meta {
  color: #111827;
  font-size: 13px;
  white-space: nowrap;

  b {
    font-weight: 700;
  }
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-left: auto;
}

.auto-save-status {
  min-width: 190px;
  color: #64748b;
  font-size: 13px;
  line-height: 20px;
  text-align: right;
  white-space: nowrap;

  &.is-saving {
    color: #2563eb;
  }

  &.is-saved {
    color: #475569;
  }
}

.outline-action,
.primary-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 7px;
  min-width: 104px;
  height: 32px;
  padding: 0 14px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 700;
  line-height: 1;

  :deep(.ant-btn-icon) {
    display: inline-flex;
    align-items: center;
    margin-inline-end: 0;
    line-height: 1;
  }

  :deep(.anticon) {
    display: inline-flex;
    align-items: center;
    line-height: 1;
  }

  :deep(.anticon + span) {
    margin-inline-start: 0;
  }
}

.outline-action {
  color: #155bdc;
  border-color: #2f6bff;
}

.primary-action {
  background: #0757e6;
  box-shadow: 0 10px 20px rgba(7, 87, 230, 0.24);
}

.workbench-main {
  display: grid;
  grid-template-columns: 240px minmax(420px, 1fr) 256px;
  gap: 10px;
  flex: 1 1 auto;
  min-height: 0;
  overflow: hidden;
  padding: 10px 10px 0;
}

.page-sidebar,
.question-panel {
  min-height: 0;
  max-height: 100%;
  overflow-y: auto;
  overscroll-behavior: contain;
  scrollbar-color: #cbd5e1 transparent;
  scrollbar-width: thin;
}

.page-sidebar::-webkit-scrollbar,
.question-panel::-webkit-scrollbar,
.score-sidebar::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

.page-sidebar::-webkit-scrollbar-track,
.question-panel::-webkit-scrollbar-track,
.score-sidebar::-webkit-scrollbar-track {
  background: transparent;
}

.page-sidebar::-webkit-scrollbar-thumb,
.question-panel::-webkit-scrollbar-thumb,
.score-sidebar::-webkit-scrollbar-thumb {
  background: #cbd5e1;
  border-radius: 999px;
}

.page-sidebar::-webkit-scrollbar-thumb:hover,
.question-panel::-webkit-scrollbar-thumb:hover,
.score-sidebar::-webkit-scrollbar-thumb:hover {
  background: #94a3b8;
}

.page-sidebar,
.question-panel,
.right-card {
  background: rgba(255, 255, 255, 0.96);
  border: 1px solid #e1e7f0;
  border-radius: 8px;
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.06);
}

.page-sidebar {
  scrollbar-gutter: stable;
}

.sidebar-title {
  position: sticky;
  top: 0;
  z-index: 2;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 34px;
  padding: 0 16px;
  background: rgba(255, 255, 255, 0.98);
  border-bottom: 1px solid #e5eaf1;
  border-radius: 8px 8px 0 0;
  font-size: 13px;
  font-weight: 700;
  backdrop-filter: blur(8px);
}

.page-group {
  padding: 10px 12px 8px;
  border-bottom: 1px solid #e5eaf1;
}

.page-group__head {
  display: flex;
  justify-content: space-between;
  color: #667085;
  cursor: pointer;
  font-size: 12px;
}

.page-group__title {
  display: inline-flex;
  gap: 8px;
  align-items: center;
  color: #111827;
  font-weight: 700;
}

.chevron-down {
  margin-top: -6px;
  font-size: 18px;
  line-height: 12px;
}

.page-group__progress {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 38px;
  align-items: center;
  gap: 10px;
  margin: 8px 0 2px 18px;
}

.progress-line {
  height: 5px;
  overflow: hidden;
  background: #e5e7eb;
  border-radius: 999px;

  i {
    display: block;
    height: 100%;
    background: #18a957;
    border-radius: inherit;
  }
}

.page-group__percent {
  color: #667085;
  text-align: right;
  font-size: 12px;
  white-space: nowrap;
}

.question-list {
  margin-top: 6px;
}

.question-item {
  display: grid;
  grid-template-columns: 50px minmax(0, 1fr) 16px;
  align-items: center;
  width: 100%;
  min-height: 30px;
  padding: 0 8px 0 26px;
  color: #4b5563;
  background: transparent;
  border: 0;
  border-radius: 0;
  cursor: pointer;
  font-size: 12px;
  text-align: left;

  strong {
    min-width: 0;
    overflow: hidden;
    color: inherit;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  span {
    white-space: nowrap;
  }

  .anticon {
    color: #18a957;
    font-size: 14px;
  }

  i,
  b {
    display: inline-block;
    width: 14px;
    height: 14px;
    border-radius: 50%;
  }

  i {
    background: radial-gradient(circle at center, #fff 19%, #1769e8 22% 100%);
  }

  b {
    border: 1px solid #b8c1cf;
  }

  &.is-active {
    position: relative;
    color: #0757e6;
    background: #eaf3ff;
    font-weight: 800;

    &::before {
      position: absolute;
      left: 0;
      width: 4px;
      height: 100%;
      background: #0757e6;
      border-radius: 0 4px 4px 0;
      content: "";
    }
  }
}

.question-panel {
  padding: 14px 16px 12px;
}

.question-title-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;

  h1 {
    margin: 0;
    color: #111827;
    font-size: 20px;
    font-weight: 900;
  }

  :deep(.ant-tag) {
    padding: 2px 8px;
    border-radius: 7px;
    font-size: 12px;
  }
}

.instruction-card {
  padding: 11px 13px;
  margin-bottom: 8px;
  background: #fff;
  border: 1px solid #d8e0eb;
  border-radius: 8px;
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.06);

  h2 {
    display: flex;
    align-items: center;
    gap: 6px;
    margin: 0 0 7px;
    color: #263247;
    font-size: 14px;
    font-weight: 700;

    .anticon {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      width: 14px;
      height: 14px;
      color: #155bdc;
      background: #edf4ff;
      border: 1px solid #c8dcff;
      border-radius: 4px;
      font-size: 9px;
    }
  }

  ul,
  ol {
    padding-left: 20px;
    margin: 0;
  }

  li,
  p {
    margin: 4px 0;
    color: #3f4856;
    font-size: 13px;
    line-height: 1.45;
  }
}

.material-card {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  align-items: center;
  gap: 16px;

  &.has-image {
    grid-template-columns: minmax(0, 1fr) 72px;
  }
}

.material-card__text {
  min-width: 0;
}

.material-card__image {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 72px;
  height: 52px;
  padding: 4px;
  background: #f8fafc;
  border: 1px solid #dbe5f0;
  border-radius: 6px;
  cursor: pointer;

  img {
    display: block;
    width: 100%;
    height: 100%;
    object-fit: contain;
  }

  &:hover {
    border-color: #75a7ff;
    background: #f3f8ff;
  }
}

.material-preview {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
  background: #f8fafc;
  border-radius: 8px;

  img {
    display: block;
    max-width: 100%;
    max-height: 420px;
    object-fit: contain;
  }
}

.guidance-card {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  align-items: center;
  gap: 16px;

  &.has-video {
    grid-template-columns: minmax(0, 1fr) 154px;
  }
}

.guidance-card__text {
  min-width: 0;
}

.guidance-video-entry {
  display: grid;
  grid-template-columns: 32px minmax(0, 1fr);
  grid-template-rows: auto auto;
  align-items: center;
  min-height: 58px;
  padding: 8px 10px;
  color: #155bdc;
  text-align: left;
  background: #f7faff;
  border: 1px solid #c8dcff;
  border-radius: 8px;
  cursor: pointer;

  span {
    grid-row: span 2;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    color: #fff;
    background: #0757e6;
    border-radius: 50%;
    font-size: 16px;
  }

  strong {
    overflow: hidden;
    color: #0f2a5f;
    font-size: 13px;
    font-weight: 800;
    line-height: 1.2;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  em {
    color: #667085;
    font-size: 12px;
    font-style: normal;
    line-height: 1.2;
  }

  &:hover {
    background: #eef5ff;
    border-color: #75a7ff;
  }
}

.guidance-video-player {
  overflow: hidden;
  background: #0f172a;
  border-radius: 8px;

  video {
    display: block;
    width: 100%;
    max-height: 420px;
    background: #0f172a;
  }
}

.draft-resume-tip {
  p {
    margin: 0 0 12px;
    color: #1f2937;
    font-size: 14px;
    line-height: 1.6;
  }
}

.draft-resume-meta {
  display: grid;
  gap: 8px;
  padding: 10px 12px;
  color: #667085;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 13px;

  b {
    color: #111827;
  }
}

.score-section {
  margin-top: 6px;

  h2 {
    margin: 0;
    font-size: 13px;
  }
}

.score-section__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 8px;
}

.previous-score-summary {
  --previous-score-color: #64748b;

  display: inline-flex;
  align-items: center;
  gap: 10px;
  height: 30px;
  padding: 0 12px;
  color: #64748b;
  background: #f8fafc;
  border: 1px solid #dbe3ed;
  border-radius: 6px;
  font-size: 12px;
  box-shadow: 0 4px 12px rgba(15, 23, 42, 0.05);
  white-space: nowrap;

  &::before {
    width: 6px;
    height: 6px;
    background: var(--previous-score-color);
    border-radius: 50%;
    content: "";
  }

  strong {
    color: #1f2937;
    font-size: 12px;
    font-weight: 800;
  }

  &.score-green {
    --previous-score-color: #0d9749;
    background: #f6fff9;
    border-color: #bde8cf;
  }

  &.score-blue {
    --previous-score-color: #0757e6;
    background: #f7faff;
    border-color: #c4d7ff;
  }

  &.score-red {
    --previous-score-color: #d41f1f;
    background: #fff8f7;
    border-color: #f5c5c1;
  }
}

.score-options {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
}

.score-option {
  --score-color: #0757e6;
  --score-hover-bg: #f8fbff;
  --score-selected-bg: #f4f8ff;
  --score-check-bg: #0757e6;

  position: relative;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  justify-content: center;
  min-height: 64px;
  padding: 10px 42px 10px 16px;
  color: #111827;
  text-align: left;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  cursor: pointer;
  transition:
    background 0.18s ease,
    border-color 0.18s ease,
    box-shadow 0.18s ease;

  &::after {
    position: absolute;
    top: 13px;
    right: 14px;
    width: 15px;
    height: 15px;
    background: #fff;
    border: 1px solid #cbd5e1;
    border-radius: 50%;
    content: "";
  }

  &:hover {
    background: var(--score-hover-bg);
    border-color: var(--score-color);
  }

  strong,
  span {
    display: block;
  }

  strong {
    color: var(--score-color);
    font-size: 19px;
    font-weight: 700;
    line-height: 1.2;
  }

  span {
    margin-top: 4px;
    color: #334155;
    font-size: 13px;
    font-weight: 500;
  }

  &.score-green {
    --score-color: #0d9749;
    --score-hover-bg: #f7fff9;
    --score-selected-bg: #f6fff9;
    --score-check-bg: #0d9749;
  }

  &.score-blue {
    --score-color: #0757e6;
    --score-hover-bg: #f7faff;
    --score-selected-bg: #f7faff;
    --score-check-bg: #0757e6;
  }

  &.score-red {
    --score-color: #d41f1f;
    --score-hover-bg: #fff8f7;
    --score-selected-bg: #fff8f7;
    --score-check-bg: #d41f1f;
  }

  &.is-selected {
    background: var(--score-selected-bg);
    border-color: var(--score-color);
    box-shadow: none;

    &::after {
      opacity: 0;
    }
  }

  &.is-previous:not(.is-selected) {
    border-color: #cad5e2;
    background: #fbfdff;
  }
}

.score-option__previous-badge {
  position: absolute;
  top: 11px;
  right: 42px;
  display: inline-flex;
  align-items: center;
  height: 22px;
  padding: 0 8px;
  color: var(--score-color);
  background: #fff;
  border: 1px solid var(--score-color);
  border-radius: 6px;
  font-size: 11px;
  font-style: normal;
  font-weight: 700;
  line-height: 20px;
}

.score-option__check {
  position: absolute;
  top: 11px;
  right: 13px;
  color: var(--score-check-bg);
  font-size: 17px;

  :deep(svg) {
    color: inherit;
    fill: currentcolor;
  }
}

.score-sidebar {
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-height: 0;
  max-height: 100%;
  overflow-y: auto;
  overscroll-behavior: contain;
  padding-right: 2px;
  scrollbar-color: #cbd5e1 transparent;
  scrollbar-gutter: stable;
  scrollbar-width: thin;
}

.right-card {
  padding: 11px 12px 10px;

  h2 {
    margin: 0 0 8px;
    font-size: 13px;
    font-weight: 800;
  }
}

.progress-card__body {
  display: grid;
  grid-template-columns: 88px 1fr;
  align-items: center;
  gap: 10px;
}

.donut {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 82px;
  height: 82px;
  color: #1f2a44;
  background:
    radial-gradient(circle at center, #fff 54%, transparent 55%),
    conic-gradient(#2563eb 0 68%, #e5e7eb 68% 100%);
  border-radius: 50%;
  font-size: 20px;
  font-weight: 700;
}

.progress-stats {
  display: grid;
  gap: 4px;
  color: #667085;
  font-size: 12px;

  strong {
    color: #0757e6;
    font-size: 15px;

    i {
      color: #4b5563;
      font-size: 12px;
      font-style: normal;
      font-weight: 500;
    }
  }

  .danger {
    color: #f04438;
  }
}

.caregiver-card {
  padding: 14px 16px;
}

.caregiver-qrcode {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 10px 0 12px;
  text-align: center;

  img {
    width: 132px;
    height: 132px;
    padding: 6px;
    background: #fff;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
  }

  p {
    margin: 8px 0 0;
    color: #667085;
    font-size: 12px;
  }
}

.caregiver-actions {
  display: grid;
  gap: 8px;
}

.caregiver-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 30px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 700;

  :deep(.ant-btn-icon) {
    display: inline-flex;
    align-items: center;
    margin-inline-end: 0;
  }
}

.caregiver-action--wechat {
  background: #0757e6;
}

.training-record-card {
  padding: 14px 16px;

  h2 {
    margin-bottom: 10px;
  }
}

.training-record-field {
  padding: 10px;
  background: #f8fafc;
  border: 1px solid #e7edf3;
  border-radius: 6px;
}

.training-record-field__head {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 8px;
  color: #475467;
  font-size: 12px;
  line-height: 22px;
}

.training-record-field__type {
  margin-inline-end: 0;
  font-size: 11px;
  line-height: 18px;
}

.training-record-checks {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 14px;
}

.training-record-checks :deep(.ant-checkbox-wrapper) {
  margin-inline-start: 0;
  color: #344054;
  font-size: 12px;
}

.training-record-empty {
  padding: 12px 10px;
  color: #98a2b3;
  background: #f8fafc;
  border: 1px solid #e7edf3;
  border-radius: 6px;
  font-size: 12px;
}

.workbench-footer {
  position: sticky;
  bottom: 0;
  z-index: 11;
  flex: 0 0 auto;
  display: grid;
  grid-template-columns: 136px 1fr 146px 158px 142px;
  align-items: center;
  gap: 16px;
  min-height: 58px;
  padding: 0 20px;
  background: rgba(255, 255, 255, 0.96);
  border: 1px solid #d8dfe8;
  border-radius: 10px 10px 0 0;
  box-shadow: 0 -8px 24px rgba(15, 23, 42, 0.08);
}

.nav-button,
.next-button {
  height: 34px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 800;
}

.nav-button {
  color: #0757e6;
  border-color: #9bbcff;
}

.next-button {
  background: #0757e6;
  box-shadow: 0 10px 20px rgba(7, 87, 230, 0.22);
}

.question-counter {
  text-align: center;

  strong {
    color: #1f2937;
    font-size: 24px;
    letter-spacing: 0;
  }

  span {
    margin-left: 6px;
    color: #1f2937;
    font-size: 14px;
  }
}

.auto-next {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  color: #374151;
  font-size: 12px;
}

@media (max-width: 1400px) {
  .workbench-main {
    grid-template-columns: 220px minmax(400px, 1fr) 240px;
  }

  .header-divider {
    margin: 0 16px;
  }

  .header-meta {
    font-size: 14px;
  }

  .workbench-footer {
    gap: 24px;
  }
}
</style>
