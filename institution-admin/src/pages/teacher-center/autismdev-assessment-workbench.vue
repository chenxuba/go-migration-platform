<script setup lang="ts">
import {
  ArrowLeftOutlined,
  CheckCircleFilled,
  FileDoneOutlined,
  FileSearchOutlined,
  FileTextOutlined,
  FilterOutlined,
  LeftOutlined,
  RightOutlined,
  SaveOutlined,
  SlidersOutlined,
  SwapOutlined,
} from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import { computed, nextTick, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  getAutismDevAssessmentRecordDetailApi,
  getAutismDevAssessmentDraftDetailApi,
  getAutismDevAssessmentFormTemplateItemApi,
  getAutismDevAssessmentFormTemplateSummaryApi,
  pageAutismDevAssessmentDraftsApi,
  saveAutismDevAssessmentDraftApi,
  saveAutismDevAssessmentDraftItemApi,
  submitAutismDevAssessmentDraftApi,
  updateAutismDevAssessmentRecordApi,
  type AutismDevAssessmentRecordDetail,
  type AutismDevAssessmentDraftSummary,
  type AutismDevAssessmentItem,
  type AutismDevDomainGroup,
  type AutismDevDraftDetail,
  type AutismDevDraftInput,
  type AutismDevDraftSaveRequest,
  type AutismDevItemSummary,
  type AutismDevQuestionDisplayPreference,
  type AutismDevScopeMode,
  type AutismDevScoreOption,
  type AutismDevTemplateSummary,
} from '@/api/edu-center/autismdev-assessment'
import { getScaleAssessmentStudentCandidatesApi } from '@/api/teacher-center/scale-library'
import messageService from '@/utils/messageService'

type DraftItemSaveStatus = 'saving' | 'saved' | 'error'
type NormalizedQuestionPreference = 'all' | 'matchingAge' | 'ageAndBelow'
type NormalizedScopeMode = 'full' | 'custom'

interface RangeOption {
  label: string
  total: number
  done: number
}

const route = useRoute()
const router = useRouter()

const templateLoading = ref(false)
const itemLoading = ref(false)
const saving = ref(false)
const submitting = ref(false)
const editingRecordId = ref(numberFromQuery('recordId') || 0)
const recordMode = computed(() => {
  const raw = route.query.recordMode
  const value = Array.isArray(raw) ? raw[0] : raw
  return String(value || '').trim().toLowerCase()
})
const draftResumeModalOpen = ref(false)
const scopeEditorOpen = ref(false)
const preferenceModalOpen = ref(false)
const selectedDomainCode = ref('')
const selectedItemNo = ref(numberFromQuery('itemNo') || 0)
const selectedRangeFilter = ref('')
const autoNext = ref(true)
const expandedGroupKeys = ref<string[]>([])
const questionDisplayPreference = ref<NormalizedQuestionPreference>('ageAndBelow')
const scopeMode = ref<NormalizedScopeMode>('full')
const selectedScopeDomainCodes = ref<string[]>([])
const draftScopeMode = ref<NormalizedScopeMode>('full')
const draftScopeDomainCodes = ref<string[]>([])
const autoSaveLastSavedAt = ref('')
const currentProgress = ref<AutismDevDraftDetail['progress']>()
const existingDraft = ref<AutismDevAssessmentDraftSummary>()
const template = ref<AutismDevTemplateSummary>()
const itemCache = reactive<Record<number, AutismDevAssessmentItem>>({})
const itemScores = reactive<Record<number, string>>({})
const itemRemarks = reactive<Record<number, string>>({})
const draftItemSaveStatus = ref<Record<number, DraftItemSaveStatus>>({})
const draftItemSaveErrors = ref<Record<number, string>>({})
const itemListRef = ref<HTMLElement | null>(null)
let draftCreationPromise: Promise<AutismDevDraftDetail | undefined> | undefined
let draftSavePromise: Promise<AutismDevDraftDetail | undefined> | undefined
let itemSaveChain: Promise<void> = Promise.resolve()

const editor = reactive({
  id: numberFromQuery('draftId') || undefined as number | undefined,
  studentId: numberFromQuery('childId') || undefined as number | undefined,
  studentName: textFromQuery('childName'),
  examinerName: textFromQuery('examinerName'),
  birthDate: normalizeDateText(textFromQuery('childBirthDate') || textFromQuery('birthDate')),
  assessmentDate: normalizeDateText(textFromQuery('assessmentDate')) || dayjs().format('YYYY-MM-DD'),
  remark: '',
})

const preferenceOptions: Array<{ value: NormalizedQuestionPreference, label: string, desc: string }> = [
  { value: 'ageAndBelow', label: '月龄及以下', desc: '展示实足月龄及以下题目' },
  { value: 'matchingAge', label: '匹配月龄', desc: '只展示当前月龄段题目' },
  { value: 'all', label: '全部题目', desc: '不按月龄过滤' },
]

const scaleTitle = computed(() => autismDevScaleTitle(textFromQuery('scaleName') || template.value?.title || '孤独症儿童发展评估表'))
const studentName = computed(() => editor.studentName || textFromQuery('childName') || '-')
const examinerName = computed(() => editor.examinerName || '当前老师')
const assessmentDateText = computed(() => formatDate(editor.assessmentDate))
const studentAge = computed(() => assessmentAgeText(editor.birthDate, editor.assessmentDate) || textFromQuery('childAge') || '-')
const studentAgeMonths = computed(() => {
  const dateMonths = ageMonthsFromDates(editor.birthDate, editor.assessmentDate)
  return dateMonths ?? ageMonthsFromText(textFromQuery('childAge'))
})
const studentAgeMonthText = computed(() => studentAgeMonths.value === undefined ? '月龄未知' : `${studentAgeMonths.value}月`)
const allTemplateGroups = computed(() => template.value?.domainGroups || [])
const allItems = computed(() => allTemplateGroups.value.flatMap(group => group.items || []))
const allDisplayDomainGroups = computed(() => allTemplateGroups.value.map(displayGroupForPreference).filter(group => group.items.length > 0))
const isCustomScope = computed(() => scopeMode.value === 'custom' && selectedScopeDomainCodes.value.length > 0)
const scopeDomainCodesForPayload = computed(() => {
  if (!isCustomScope.value)
    return []
  const selected = new Set(selectedScopeDomainCodes.value)
  return allTemplateGroups.value.map(group => group.domainCode.trim()).filter(code => code && selected.has(code))
})
const displayDomainGroups = computed(() => {
  if (!isCustomScope.value)
    return allDisplayDomainGroups.value
  const selected = new Set(scopeDomainCodesForPayload.value)
  return allDisplayDomainGroups.value.filter(group => selected.has(group.domainCode.trim()))
})
const displayItems = computed(() => displayDomainGroups.value.flatMap(group => group.items || []))
const selectedGroup = computed(() => {
  const matched = displayDomainGroups.value.find(group => group.domainCode === selectedDomainCode.value)
  return matched || displayDomainGroups.value[0]
})
const selectedGroupItems = computed(() => selectedGroup.value?.items || [])
const filteredSelectedGroupItems = computed(() => {
  const filter = selectedRangeFilter.value.trim()
  if (!filter)
    return selectedGroupItems.value
  return selectedGroupItems.value.filter(item => assessmentRangeBucket(item) === filter)
})
const currentItemSummary = computed(() => summaryByNo(selectedItemNo.value))
const currentItem = computed(() => itemCache[selectedItemNo.value])
const currentScoreOptions = computed(() => scoreOptionsForItem(currentItem.value || currentItemSummary.value))
const selectedScore = computed(() => itemScores[selectedItemNo.value] || '')
const currentRemark = computed({
  get: () => itemRemarks[selectedItemNo.value] || '',
  set: (value: string) => updateItemRemark(value),
})
const answeredItemCount = computed(() => displayItems.value.filter(item => hasScore(item.itemNo)).length)
const totalItemCount = computed(() => displayItems.value.length)
const missingItemCount = computed(() => Math.max(totalItemCount.value - answeredItemCount.value, 0))
const progressPercent = computed(() => totalItemCount.value ? Math.round((answeredItemCount.value / totalItemCount.value) * 100) : 0)
const selectedDomainProgress = computed(() => domainProgress(selectedGroup.value))
const currentIndex = computed(() => {
  const index = displayItems.value.findIndex(item => item.itemNo === selectedItemNo.value)
  return index >= 0 ? index : 0
})
const currentDisplayIndex = computed(() => totalItemCount.value ? currentIndex.value + 1 : 0)
const hasPreviousItem = computed(() => currentIndex.value > 0)
const hasNextItem = computed(() => currentIndex.value < displayItems.value.length - 1)
const missingItems = computed(() => displayItems.value.filter(item => !hasScore(item.itemNo)).slice(0, 12))
const rangeOptions = computed(() => rangeOptionsForGroup(selectedGroup.value))
const scopeCheckboxOptions = computed(() => allDisplayDomainGroups.value.map(group => ({ label: group.domainName || group.title, value: group.domainCode })))
const scopeText = computed(() => isCustomScope.value ? `自定义 ${scopeDomainCodesForPayload.value.length}/${allDisplayDomainGroups.value.length}` : '全量')
const preferenceText = computed(() => questionPreferenceLabel(questionDisplayPreference.value))
const currentItemTitle = computed(() => currentItemSummary.value ? displayItemTitle(currentItemSummary.value) : '-')
const currentAgeRangeText = computed(() => ageRangeText(currentItem.value || currentItemSummary.value))
const currentRangeText = computed(() => assessmentRangeText(currentItem.value || currentItemSummary.value))
const donutStyle = computed(() => ({
  background: `radial-gradient(circle at center, #fff 54%, transparent 55%), conic-gradient(#2563eb 0 ${progressPercent.value}%, #e5e7eb ${progressPercent.value}% 100%)`,
}))
const pageGroups = computed(() => displayDomainGroups.value.map((group) => {
  const progress = domainProgress(group)
  const items = (group.domainCode === selectedDomainCode.value ? filteredSelectedGroupItems.value : group.items).map(item => ({
    no: item.itemNo,
    labelNo: item.domainItemNo || item.itemNo,
    name: displayItemTitle(item),
    status: navItemStatus(item),
  }))
  return {
    key: group.domainCode,
    title: group.domainName || group.title,
    count: `${progress.answered}/${progress.total}`,
    percent: progress.percent,
    expanded: expandedGroupKeys.value.includes(group.domainCode),
    items,
  }
}))
const autoSaveState = computed<'idle' | 'saving' | 'saved'>(() => {
  if (saving.value || Object.values(draftItemSaveStatus.value).some(status => status === 'saving'))
    return 'saving'
  return autoSaveLastSavedAt.value ? 'saved' : 'idle'
})
const isRecordReuseMode = computed(() => editingRecordId.value > 0 && recordMode.value === 'reuse')
const isSubmittedRecordMode = computed(() => editingRecordId.value > 0)
const isRecordEditMode = computed(() => editingRecordId.value > 0 && !isRecordReuseMode.value)
const autoSaveText = computed(() => {
  if (isRecordReuseMode.value)
    return '复用测评中，提交后生成新记录'
  if (isRecordEditMode.value)
    return '正式记录修改中，修改后请重新提交'
  if (autoSaveState.value === 'saving')
    return '保存中...'
  if (autoSaveState.value === 'saved')
    return `已自动保存 ${autoSaveLastSavedAt.value}`
  return '等待作答'
})
const submitActionText = computed(() => {
  if (isRecordReuseMode.value)
    return '提交新记录'
  if (isRecordEditMode.value)
    return '重新提交'
  return '提交记录'
})

watch(selectedItemNo, (itemNo) => {
  if (itemNo > 0)
    void fetchItemDetail(itemNo)
  void revealSelectedItem()
})

watch(displayItems, () => {
  ensureSelectedDisplayItem()
})

onMounted(() => {
  void initializeWorkbench()
})

function unwrap<T>(res: any): T {
  return (res?.data ?? res?.result ?? res) as T
}

function getErrorMessage(error: any, fallback: string) {
  return error?.response?.data?.message || error?.message || fallback
}

function isDraftNotFoundError(error: any) {
  return getErrorMessage(error, '').toLowerCase().includes('assessment draft not found')
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

function assessmentAgeText(birthDate?: string, assessmentDate?: string) {
  const birth = dayjs(birthDate).startOf('day')
  const target = dayjs(assessmentDate).startOf('day')
  if (!birth.isValid() || !target.isValid() || birth.isAfter(target))
    return ''
  const years = target.diff(birth, 'year')
  const afterYears = birth.add(years, 'year')
  const months = target.diff(afterYears, 'month')
  const afterMonths = afterYears.add(months, 'month')
  const days = target.diff(afterMonths, 'day')
  const parts: string[] = []
  if (years)
    parts.push(`${years}岁`)
  if (months)
    parts.push(`${months}月`)
  if (days || !parts.length)
    parts.push(`${days}天`)
  return parts.join('')
}

function ageMonthsFromDates(birthDate?: string, assessmentDate?: string) {
  const birth = dayjs(birthDate).startOf('day')
  const target = dayjs(assessmentDate).startOf('day')
  if (!birth.isValid() || !target.isValid() || birth.isAfter(target))
    return undefined
  let months = (target.year() - birth.year()) * 12 + target.month() - birth.month()
  if (target.date() < birth.date())
    months -= 1
  return Math.max(months, 0)
}

function ageMonthsFromText(ageText?: string) {
  const text = String(ageText || '').trim()
  if (!text || text === '未知')
    return undefined
  const monthMatch = text.match(/(\d+)\s*(?:个)?月/)
  if (monthMatch)
    return Number(monthMatch[1])
  const yearMonthMatch = text.match(/(\d+)\s*岁\s*(\d+)?/)
  if (yearMonthMatch)
    return Number(yearMonthMatch[1]) * 12 + Number(yearMonthMatch[2] || 0)
  const raw = Number(text)
  return Number.isFinite(raw) ? raw : undefined
}

function normalizeText(value?: string, fallback = '-') {
  const text = String(value || '').trim()
  return text || fallback
}

function compactText(value?: string, fallback = '-') {
  const text = String(value || '').replace(/\s+/g, ' ').trim()
  return text || fallback
}

function autismDevScaleTitle(raw: string) {
  const title = raw.trim() || '孤独症儿童发展评估表'
  const cleaned = title.replace(/\s*[（(]?\s*试行\s*[）)]?\s*$/u, '').trim()
  return cleaned || title
}

async function initializeWorkbench() {
  await fetchTemplate()
  if (isSubmittedRecordMode.value) {
    await fetchRecordForEdit(editingRecordId.value)
    return
  }
  if (editor.id) {
    await fetchDraftDetail(editor.id)
    await hydrateStudentBirthDate()
    selectInitialItem(currentProgress.value?.missingItemNos?.[0])
    return
  }
  await hydrateStudentBirthDate()
  const draft = await findExistingDraft()
  if (draft) {
    existingDraft.value = draft
    applyDraftQuestionPreference(draft.progress?.questionDisplayPreference)
    draftResumeModalOpen.value = true
    selectInitialItem(draft.progress?.missingItemNos?.[0])
    return
  }
  await startNewAssessment()
}

async function fetchRecordForEdit(recordId: number) {
  if (!recordId)
    return
  try {
    const res = await getAutismDevAssessmentRecordDetailApi(recordId)
    const detail = unwrap<AutismDevAssessmentRecordDetail>(res)
    const input = normalizeDraftInputSnapshot(detail.input)
    const reuseAssessmentDate = dayjs().format('YYYY-MM-DD')
    editor.id = undefined
    editor.studentId = detail.studentId || editor.studentId
    editor.studentName = detail.studentName || editor.studentName
    editor.examinerName = detail.examinerName || editor.examinerName
    editor.birthDate = normalizeDateText(detail.birthDate || input?.birthDate) || editor.birthDate
    editor.assessmentDate = isRecordReuseMode.value ? reuseAssessmentDate : normalizeDateText(detail.assessmentDate || input?.assessmentDate) || editor.assessmentDate
    editor.remark = detail.remark || input?.remark || ''
    applyDraftInput(input)
    applyDraftQuestionPreference(input?.questionDisplayPreference)
    if (isRecordReuseMode.value)
      editor.assessmentDate = reuseAssessmentDate
    selectInitialItem()
    messageService.info(isRecordReuseMode.value ? '当前正在复用已提交的孤独症儿童发展评估，提交后会生成新的正式记录' : '当前正在修改已提交的孤独症儿童发展评估记录，修改后请重新提交')
  }
  catch (error: any) {
    messageService.error(getErrorMessage(error, '获取孤独症儿童发展评估记录失败'))
    void router.push('/teacherCenter/evaluationRecord')
  }
}

async function fetchTemplate() {
  templateLoading.value = true
  try {
    const res = await getAutismDevAssessmentFormTemplateSummaryApi()
    template.value = unwrap<AutismDevTemplateSummary>(res)
    selectedDomainCode.value = displayDomainGroups.value[0]?.domainCode || template.value.domainGroups?.[0]?.domainCode || ''
    expandedGroupKeys.value = selectedDomainCode.value ? [selectedDomainCode.value] : []
    selectInitialItem()
  }
  catch (error: any) {
    messageService.error(getErrorMessage(error, '获取孤独症儿童发展评估表题目目录失败'))
  }
  finally {
    templateLoading.value = false
  }
}

async function fetchItemDetail(itemNo: number) {
  if (itemNo <= 0 || itemCache[itemNo])
    return
  itemLoading.value = true
  try {
    const res = await getAutismDevAssessmentFormTemplateItemApi(itemNo)
    const detail = unwrap<AutismDevAssessmentItem>(res)
    if (detail?.itemNo === itemNo)
      itemCache[itemNo] = detail
  }
  catch (error: any) {
    messageService.error(getErrorMessage(error, `获取第 ${itemNo} 项题目说明失败`))
  }
  finally {
    itemLoading.value = false
  }
}

async function hydrateStudentBirthDate() {
  if (editor.birthDate || !editor.studentId)
    return
  try {
    const res = await getScaleAssessmentStudentCandidatesApi({
      scaleCode: 'AUTISMDEV',
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
  }
}

async function findExistingDraft() {
  if (!editor.studentId)
    return undefined
  try {
    const res = await pageAutismDevAssessmentDraftsApi({
      pageRequestModel: { pageIndex: 1, pageSize: 1 },
      queryModel: {
        assessmentCode: 'AUTISMDEV',
        studentId: editor.studentId,
        latestOnly: true,
      },
      latestOnly: true,
    })
    const data = unwrap<any>(res)
    return (data?.items || [])[0] as AutismDevAssessmentDraftSummary | undefined
  }
  catch (error: any) {
    messageService.error(getErrorMessage(error, '查询孤独症儿童发展评估草稿失败'))
    return undefined
  }
}

async function startNewAssessment() {
  resetAssessmentInput()
  selectInitialItem()
  if (validateDraftHeader(true))
    await saveDraft(true)
}

async function continueExistingDraft() {
  const draft = existingDraft.value
  draftResumeModalOpen.value = false
  if (!draft?.id)
    return
  await fetchDraftDetail(Number(draft.id))
  selectInitialItem(draft.progress?.missingItemNos?.[0])
}

async function restartAssessment() {
  draftResumeModalOpen.value = false
  existingDraft.value = undefined
  await startNewAssessment()
}

async function fetchDraftDetail(id: number) {
  try {
    const res = await getAutismDevAssessmentDraftDetailApi(id)
    applyDraftDetail(unwrap<AutismDevDraftDetail>(res))
  }
  catch (error: any) {
    messageService.error(getErrorMessage(error, '获取孤独症儿童发展评估草稿失败'))
  }
}

function applyDraftDetail(detail: AutismDevDraftDetail) {
  editor.id = detail.id
  editor.studentId = detail.studentId || editor.studentId
  editor.studentName = detail.studentName || editor.studentName
  editor.examinerName = detail.examinerName || editor.examinerName
  editor.birthDate = normalizeDateText(detail.birthDate || detail.input?.birthDate) || editor.birthDate
  editor.assessmentDate = normalizeDateText(detail.assessmentDate || detail.input?.assessmentDate) || editor.assessmentDate
  editor.remark = detail.remark || detail.input?.remark || ''
  currentProgress.value = detail.progress
  applyDraftInput(detail.input)
  applyDraftQuestionPreference(detail.input?.questionDisplayPreference || detail.progress?.questionDisplayPreference)
  autoSaveLastSavedAt.value = detail.updatedTime ? dayjs(detail.updatedTime).format('MM-DD HH:mm') : autoSaveLastSavedAt.value
}

function mergeDraftDetailInput(detail: AutismDevDraftDetail) {
  const localScores = { ...itemScores }
  const localRemarks = { ...itemRemarks }
  applyDraftDetail(detail)
  Object.entries(localScores).forEach(([itemNo, score]) => {
    if (score)
      itemScores[Number(itemNo)] = score
  })
  Object.entries(localRemarks).forEach(([itemNo, remark]) => {
    if (remark?.trim())
      itemRemarks[Number(itemNo)] = remark.trim()
  })
}

function applyDraftInput(input?: AutismDevDraftInput) {
  clearRecord(itemScores)
  clearRecord(itemRemarks)
  if (!input)
    return
  Object.entries(input.itemScores || {}).forEach(([itemNo, score]) => {
    const normalized = normalizeScoreValue(score)
    if (normalized)
      itemScores[Number(itemNo)] = normalized
  })
  ;(input.itemScoreList || []).forEach((item) => {
    const normalized = normalizeScoreValue(item.score)
    if (item.itemNo > 0 && normalized)
      itemScores[item.itemNo] = normalized
    if (item.itemNo > 0 && item.remark?.trim())
      itemRemarks[item.itemNo] = item.remark.trim()
  })
  Object.entries(input.itemRemarks || {}).forEach(([itemNo, remark]) => {
    if (remark?.trim())
      itemRemarks[Number(itemNo)] = remark.trim()
  })
  ;(input.itemRemarkList || []).forEach((item) => {
    if (item.itemNo > 0 && item.remark?.trim())
      itemRemarks[item.itemNo] = item.remark.trim()
  })
  applyDraftScope(input)
}

function normalizeDraftInputSnapshot(input: unknown): AutismDevDraftInput | undefined {
  if (!input)
    return undefined
  if (typeof input === 'string') {
    try {
      return JSON.parse(input) as AutismDevDraftInput
    }
    catch {
      return undefined
    }
  }
  return input as AutismDevDraftInput
}

function clearRecord<T>(record: Record<number, T>) {
  Object.keys(record).forEach(key => delete record[Number(key)])
}

function resetAssessmentInput() {
  editor.id = undefined
  editor.remark = ''
  clearRecord(itemScores)
  clearRecord(itemRemarks)
  currentProgress.value = undefined
  autoSaveLastSavedAt.value = ''
  selectedRangeFilter.value = ''
  scopeMode.value = 'full'
  selectedScopeDomainCodes.value = []
  draftItemSaveStatus.value = {}
  draftItemSaveErrors.value = {}
}

function applyDraftScope(input?: AutismDevDraftInput) {
  const validCodes = new Set(allTemplateGroups.value.map(group => group.domainCode.trim()).filter(Boolean))
  const codes = (input?.scopeDomainCodes || []).map(code => code.trim()).filter(code => validCodes.has(code))
  const mode = String(input?.scopeMode || '').trim().toLowerCase()
  if (mode === 'custom' && codes.length) {
    scopeMode.value = 'custom'
    selectedScopeDomainCodes.value = [...new Set(codes)]
    return
  }
  scopeMode.value = 'full'
  selectedScopeDomainCodes.value = []
}

function applyDraftQuestionPreference(value?: AutismDevQuestionDisplayPreference) {
  const normalized = normalizeQuestionPreference(value)
  if (!normalized)
    return
  questionDisplayPreference.value = normalized
  selectedRangeFilter.value = ''
}

function normalizeQuestionPreference(value?: AutismDevQuestionDisplayPreference) {
  const text = String(value || '').trim()
  if (text === 'all' || text === 'matchingAge' || text === 'ageAndBelow')
    return text as NormalizedQuestionPreference
  return undefined
}

function normalizeScopeMode(value?: AutismDevScopeMode) {
  return String(value || '').trim().toLowerCase() === 'custom' ? 'custom' : 'full'
}

function normalizeScoreValue(value?: string) {
  return String(value || '').trim().toUpperCase()
}

function selectInitialItem(preferredItemNo = 0) {
  if (preferredItemNo && displayItems.value.some(item => item.itemNo === preferredItemNo)) {
    selectItem(preferredItemNo)
    return
  }
  ensureSelectedDisplayItem()
}

function ensureSelectedDisplayItem() {
  const items = displayItems.value
  if (!items.length) {
    selectedDomainCode.value = ''
    selectedItemNo.value = 0
    return
  }
  const current = items.find(item => item.itemNo === selectedItemNo.value)
  if (current) {
    selectedDomainCode.value = current.domainCode
    if (selectedRangeFilter.value && assessmentRangeBucket(current) !== selectedRangeFilter.value)
      selectedRangeFilter.value = ''
    return
  }
  const target = items.find(item => !hasScore(item.itemNo)) || items[0]
  selectedDomainCode.value = target.domainCode
  selectedItemNo.value = target.itemNo
}

function validateDraftHeader(silent = false) {
  if (!editor.studentId || studentName.value === '-') {
    if (!silent)
      messageService.warning('缺少真实儿童，无法保存测评')
    return false
  }
  return true
}

function buildPayload(): AutismDevDraftSaveRequest {
  const itemScoreList = Object.entries(itemScores)
    .map(([itemNo, score]) => ({
      itemNo: Number(itemNo),
      score: normalizeScoreValue(score),
      remark: itemRemarks[Number(itemNo)]?.trim() || '',
    }))
    .filter(item => item.itemNo > 0 && item.score)
    .sort((left, right) => left.itemNo - right.itemNo)
  const itemRemarkList = Object.entries(itemRemarks)
    .filter(([, remark]) => remark.trim())
    .map(([itemNo, remark]) => ({ itemNo: Number(itemNo), remark: remark.trim() }))
    .sort((left, right) => left.itemNo - right.itemNo)
  return {
    id: editor.id,
    studentId: editor.studentId,
    studentName: studentName.value === '-' ? undefined : studentName.value,
    examinerName: editor.examinerName,
    birthDate: editor.birthDate,
    assessmentDate: editor.assessmentDate,
    remark: editor.remark,
    scopeMode: scopeMode.value,
    scopeDomainCodes: scopeDomainCodesForPayload.value,
    questionDisplayPreference: questionDisplayPreference.value,
    itemScoreList,
    itemRemarkList,
  }
}

async function saveDraft(silent = false) {
  if (!validateDraftHeader(silent))
    return undefined
  if (draftSavePromise)
    return await draftSavePromise
  saving.value = true
  draftSavePromise = persistDraftWithRecovery().finally(() => {
    draftSavePromise = undefined
    saving.value = false
  })
  const detail = await draftSavePromise
  if (!silent && detail)
    messageService.success('孤独症儿童发展评估草稿已保存')
  return detail
}

async function persistDraftWithRecovery() {
  try {
    return await persistDraftPayload(buildPayload())
  }
  catch (error: any) {
    if (editor.id && isDraftNotFoundError(error)) {
      editor.id = undefined
      draftCreationPromise = undefined
      try {
        return await persistDraftPayload({ ...buildPayload(), id: undefined })
      }
      catch (retryError: any) {
        messageService.error(getErrorMessage(retryError, '保存孤独症儿童发展评估草稿失败'))
        return undefined
      }
    }
    messageService.error(getErrorMessage(error, '保存孤独症儿童发展评估草稿失败'))
    return undefined
  }
}

async function persistDraftPayload(payload: AutismDevDraftSaveRequest) {
  const res = await saveAutismDevAssessmentDraftApi(payload)
  const detail = unwrap<AutismDevDraftDetail>(res)
  applyDraftDetail(detail)
  return detail
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

function queueSaveItem(itemNo: number, moveNext = false) {
  itemSaveChain = itemSaveChain.then(() => persistItem(itemNo, moveNext)).catch(() => undefined)
}

async function persistItem(itemNo: number, moveNext = false) {
  if (itemNo <= 0 || !hasScore(itemNo))
    return
  draftItemSaveStatus.value = { ...draftItemSaveStatus.value, [itemNo]: 'saving' }
  try {
    const canSave = await ensureDraftForItemSave()
    if (!canSave || !editor.id)
      throw new Error('草稿创建失败')
    const detail = await persistDraftItemWithRecovery(itemNo)
    if (!detail)
      throw new Error('草稿保存失败')
    mergeDraftDetailInput(detail)
    autoSaveLastSavedAt.value = dayjs().format('MM-DD HH:mm')
    draftItemSaveStatus.value = { ...draftItemSaveStatus.value, [itemNo]: 'saved' }
    if (moveNext && selectedItemNo.value === itemNo)
      goNextItem()
  }
  catch (error: any) {
    const message = getErrorMessage(error, `第 ${itemNo} 项自动保存失败`)
    draftItemSaveStatus.value = { ...draftItemSaveStatus.value, [itemNo]: 'error' }
    draftItemSaveErrors.value = { ...draftItemSaveErrors.value, [itemNo]: message }
    messageService.error(message)
  }
}

async function persistDraftItemWithRecovery(itemNo: number) {
  try {
    return await persistDraftItem(itemNo, editor.id)
  }
  catch (error: any) {
    if (!isDraftNotFoundError(error))
      throw error
    editor.id = undefined
    draftCreationPromise = undefined
    const detail = await saveDraft(true)
    const draftId = detail?.id || editor.id
    if (!draftId)
      throw error
    return await persistDraftItem(itemNo, draftId)
  }
}

async function persistDraftItem(itemNo: number, draftId?: number) {
  if (!draftId)
    throw new Error('草稿创建失败')
  const res = await saveAutismDevAssessmentDraftItemApi({
    draftId,
    itemNo,
    score: itemScores[itemNo],
    remark: itemRemarks[itemNo]?.trim() || '',
  })
  return unwrap<AutismDevDraftDetail>(res)
}

async function submitDraft() {
  if (isRecordEditMode.value) {
    await submitRecordEdit()
    return
  }
  if (!editor.birthDate || !editor.assessmentDate) {
    messageService.warning('缺少出生日期或测查日期，不能提交正式记录')
    return
  }
  const missingNo = firstUnansweredItemNo()
  if (missingNo > 0) {
    selectItem(missingNo)
    messageService.warning(`本次范围还有 ${missingItemCount.value} 道题未评分，完成后再提交`)
    return
  }
  if (isCustomScope.value) {
    await saveDraft(true)
    messageService.warning('自定义范围已保存为草稿；部分领域正式记录待后端支持')
    return
  }
  submitting.value = true
  try {
    await itemSaveChain
    const detail = await saveDraft(true)
    const draftId = detail?.id || editor.id
    if (!draftId) {
      messageService.warning('请先保存草稿，再提交正式记录')
      return
    }
    await submitAutismDevAssessmentDraftApi(draftId)
    messageService.success('已提交正式测评记录')
    await router.push('/teacherCenter/evaluationRecord')
  }
  catch (error: any) {
    messageService.error(getErrorMessage(error, '提交孤独症儿童发展评估记录失败'))
  }
  finally {
    submitting.value = false
  }
}

async function submitRecordEdit() {
  if (!editor.birthDate || !editor.assessmentDate) {
    messageService.warning('缺少出生日期或测查日期，不能提交正式记录')
    return
  }
  const missingNo = firstUnansweredItemNo()
  if (missingNo > 0) {
    selectItem(missingNo)
    messageService.warning(`本次范围还有 ${missingItemCount.value} 道题未评分，完成后再重新提交`)
    return
  }
  if (isCustomScope.value) {
    messageService.warning('自定义范围暂不支持直接修改正式记录')
    return
  }
  submitting.value = true
  try {
    const res = await updateAutismDevAssessmentRecordApi({
      ...buildPayload(),
      id: editingRecordId.value,
    })
    const result = unwrap<AutismDevAssessmentRecordDetail>(res)
    messageService.success('已重新提交，并生成新的孤独症儿童发展评估报告')
    if (result?.id)
      await router.push('/teacherCenter/evaluationRecord')
  }
  catch (error: any) {
    messageService.error(getErrorMessage(error, '重新提交孤独症儿童发展评估记录失败'))
  }
  finally {
    submitting.value = false
  }
}

function selectDomain(code: string) {
  if (!code)
    return
  selectedDomainCode.value = code
  if (!expandedGroupKeys.value.includes(code))
    expandedGroupKeys.value = [code]
  selectedRangeFilter.value = ''
  const group = displayDomainGroups.value.find(item => item.domainCode === code)
  const target = group?.items.find(item => !hasScore(item.itemNo)) || group?.items[0]
  if (target)
    selectedItemNo.value = target.itemNo
  void scrollItemListToTop()
}

function selectItem(itemNo: number) {
  const item = displayItems.value.find(row => row.itemNo === itemNo)
  if (!item)
    return
  selectedItemNo.value = item.itemNo
  selectedDomainCode.value = item.domainCode
  if (!expandedGroupKeys.value.includes(item.domainCode))
    expandedGroupKeys.value = [item.domainCode]
  if (selectedRangeFilter.value && assessmentRangeBucket(item) !== selectedRangeFilter.value)
    selectedRangeFilter.value = ''
}

function scoreCurrent(score: string) {
  scoreItem(selectedItemNo.value, score, autoNext.value)
}

function scoreItem(itemNo: number, score: string, moveNext = false) {
  if (itemNo <= 0 || submitting.value)
    return
  const item = displayItems.value.find(row => row.itemNo === itemNo)
  if (!item)
    return
  selectedItemNo.value = item.itemNo
  selectedDomainCode.value = item.domainCode
  itemScores[itemNo] = normalizeScoreValue(score)
  if (isRecordEditMode.value) {
    if (moveNext)
      goNextItem()
    return
  }
  queueSaveItem(itemNo, moveNext)
}

function updateItemRemark(value: string) {
  const itemNo = selectedItemNo.value
  if (itemNo <= 0)
    return
  const normalized = value.trim()
  if (normalized)
    itemRemarks[itemNo] = normalized
  else
    delete itemRemarks[itemNo]
}

function finishItemRemarkEdit() {
  const itemNo = selectedItemNo.value
  if (!isRecordEditMode.value && itemNo > 0 && hasScore(itemNo))
    queueSaveItem(itemNo)
}

function setQuestionDisplayPreference(value: NormalizedQuestionPreference) {
  if (questionDisplayPreference.value === value)
    return
  questionDisplayPreference.value = value
  selectedRangeFilter.value = ''
  ensureSelectedDisplayItem()
  if (!isRecordEditMode.value)
    void saveDraft(true)
}

function openQuestionPreferenceEditor() {
  preferenceModalOpen.value = true
}

function chooseQuestionDisplayPreference(value: NormalizedQuestionPreference) {
  setQuestionDisplayPreference(value)
  preferenceModalOpen.value = false
}

function openScopeEditor() {
  draftScopeMode.value = scopeMode.value
  draftScopeDomainCodes.value = [...selectedScopeDomainCodes.value]
  scopeEditorOpen.value = true
}

function applyScopeEdit() {
  const mode = normalizeScopeMode(draftScopeMode.value)
  if (mode === 'custom' && !draftScopeDomainCodes.value.length) {
    messageService.warning('自定义范围至少选择 1 个领域')
    return
  }
  scopeMode.value = mode
  selectedScopeDomainCodes.value = mode === 'custom' ? orderedValidScopeCodes(draftScopeDomainCodes.value) : []
  scopeEditorOpen.value = false
  selectedRangeFilter.value = ''
  ensureSelectedDisplayItem()
  if (!isRecordEditMode.value)
    void saveDraft(true)
}

function orderedValidScopeCodes(codes: string[]) {
  const selected = new Set(codes.map(code => code.trim()).filter(Boolean))
  return allTemplateGroups.value.map(group => group.domainCode.trim()).filter(code => code && selected.has(code))
}

function selectRangeFilter(label: string) {
  const next = selectedRangeFilter.value === label ? '' : label
  selectedRangeFilter.value = next
  const items = filteredSelectedGroupItems.value
  const target = items.find(item => !hasScore(item.itemNo)) || items[0]
  if (target)
    selectItem(target.itemNo)
}

function goPreviousItem() {
  if (!hasPreviousItem.value)
    return
  selectItem(displayItems.value[currentIndex.value - 1].itemNo)
}

function goNextItem() {
  if (!hasNextItem.value)
    return
  selectItem(displayItems.value[currentIndex.value + 1].itemNo)
}

function jumpToMissing() {
  const itemNo = firstUnansweredItemNo()
  if (!itemNo) {
    messageService.success('当前没有缺题')
    return
  }
  selectItem(itemNo)
}

function firstUnansweredItemNo() {
  return displayItems.value.find(item => !hasScore(item.itemNo))?.itemNo || 0
}

function hasScore(itemNo: number) {
  return Boolean(itemScores[itemNo]?.trim())
}

function itemStatusClass(item: AutismDevItemSummary) {
  if (item.itemNo === selectedItemNo.value)
    return 'is-active'
  if (!hasScore(item.itemNo))
    return 'is-empty'
  const score = itemScores[item.itemNo]
  if (score === 'P' || score === 'A')
    return 'is-strong'
  if (score === 'E' || score === 'M')
    return 'is-middle'
  if (score === 'F' || score === 'S')
    return 'is-weak'
  return 'is-exempt'
}

function navItemStatus(item: AutismDevItemSummary) {
  if (item.itemNo === selectedItemNo.value)
    return 'active'
  return hasScore(item.itemNo) ? 'done' : 'todo'
}

function toggleGroup(key: string) {
  if (expandedGroupKeys.value.includes(key)) {
    expandedGroupKeys.value = expandedGroupKeys.value.filter(item => item !== key)
    return
  }
  expandedGroupKeys.value = [key]
  if (selectedDomainCode.value !== key)
    selectDomain(key)
}

function goToItem(itemNo: number) {
  selectItem(itemNo)
}

function itemSaveStatusText(itemNo: number) {
  const status = draftItemSaveStatus.value[itemNo]
  if (status === 'saving')
    return '保存中'
  if (status === 'saved')
    return '已保存'
  if (status === 'error')
    return '保存失败'
  return hasScore(itemNo) ? '已记录' : '未记录'
}

function scoreOptionsForItem(item?: AutismDevAssessmentItem | AutismDevItemSummary) {
  if (!item)
    return []
  if ('scoreOptions' in item && item.scoreOptions?.length)
    return item.scoreOptions
  const scoreType = item.scoreType || selectedGroup.value?.scoreType || ''
  return (template.value?.scoreOptions || []).filter(option => option.scoreType.toUpperCase() === scoreType.toUpperCase())
}

function displayGroupForPreference(group: AutismDevDomainGroup): AutismDevDomainGroup {
  const items = (group.items || []).filter(item => shouldDisplayItemForPreference(item, questionDisplayPreference.value))
  return {
    ...group,
    itemCount: items.length,
    items,
  }
}

function shouldDisplayItemForPreference(item: AutismDevItemSummary, preference: NormalizedQuestionPreference) {
  if (isEmotionBehaviorItem(item) || preference === 'all')
    return true
  const ageMonths = studentAgeMonths.value
  if (ageMonths === undefined)
    return true
  const minMonth = Number(item.ageMinMonth || 0)
  const maxMonth = Number(item.ageMaxMonth || 0)
  if (minMonth <= 0 && maxMonth <= 0)
    return true
  if (preference === 'matchingAge')
    return minMonth <= ageMonths && (maxMonth <= 0 || ageMonths <= maxMonth)
  return minMonth <= ageMonths
}

function isEmotionBehaviorItem(item?: AutismDevItemSummary) {
  return String(item?.domainCode || '').trim().toUpperCase() === 'EB' || String(item?.scoreType || '').trim().toUpperCase() === 'AMS'
}

function summaryByNo(itemNo: number) {
  return allItems.value.find(item => item.itemNo === itemNo)
}

function displayItemTitle(item?: AutismDevItemSummary) {
  return compactText(item?.itemTitle || item?.testItem)
}

function assessmentRangeBucket(item?: AutismDevItemSummary) {
  const parts = String(item?.assessmentRange || '')
    .split('/')
    .map(part => part.trim())
    .filter(Boolean)
  if (parts.length >= 2)
    return parts[1]
  return parts[0] || '未分类'
}

function assessmentRangeText(item?: AutismDevItemSummary | AutismDevAssessmentItem) {
  return normalizeText(item?.assessmentRange)
}

function ageRangeText(item?: AutismDevItemSummary | AutismDevAssessmentItem) {
  if (!item)
    return '-'
  const min = Number(item.ageMinMonth || 0)
  const max = Number(item.ageMaxMonth || 0)
  if (min <= 0 && max <= 0)
    return '-'
  if (max <= 0)
    return `${min}月以上`
  return `${Math.max(min, 0)}-${max}月`
}

function rangeOptionsForGroup(group?: AutismDevDomainGroup) {
  const optionByLabel = new Map<string, RangeOption>()
  for (const item of group?.items || []) {
    const label = assessmentRangeBucket(item)
    const option = optionByLabel.get(label) || { label, total: 0, done: 0 }
    option.total += 1
    if (hasScore(item.itemNo))
      option.done += 1
    optionByLabel.set(label, option)
  }
  return [...optionByLabel.values()]
}

function domainProgress(group?: AutismDevDomainGroup) {
  const items = group?.items || []
  const answered = items.filter(item => hasScore(item.itemNo)).length
  const total = items.length
  return {
    answered,
    total,
    percent: total ? Math.round((answered / total) * 100) : 0,
    complete: total > 0 && answered >= total,
  }
}

function questionPreferenceLabel(value: NormalizedQuestionPreference) {
  if (value === 'all')
    return '全部题目'
  if (value === 'matchingAge')
    return '匹配月龄'
  return '月龄及以下'
}

function scoreOptionTitle(option: AutismDevScoreOption) {
  const label = option.label.trim()
  const prefix = option.value.trim()
  if (label && prefix && label.startsWith(prefix))
    return label.slice(prefix.length).trim() || label
  return label || prefix
}

function scoreColor(score?: string) {
  switch (String(score || '').trim().toUpperCase()) {
    case 'P':
    case 'A':
      return '#0f8a5f'
    case 'E':
    case 'M':
      return '#c26816'
    case 'F':
    case 'S':
      return '#c43d3d'
    case 'X':
    default:
      return '#64748b'
  }
}

function scoreTone(score?: string) {
  switch (String(score || '').trim().toUpperCase()) {
    case 'P':
    case 'A':
      return 'green'
    case 'E':
    case 'M':
      return 'blue'
    case 'F':
    case 'S':
      return 'red'
    case 'X':
    default:
      return 'gray'
  }
}

async function revealSelectedItem() {
  await nextTick()
  const container = itemListRef.value
  if (!container || selectedItemNo.value <= 0)
    return
  const target = container.querySelector<HTMLElement>(`[data-item-no="${selectedItemNo.value}"]`)
  if (!target)
    return
  const targetTop = target.offsetTop - container.clientHeight / 2 + target.offsetHeight / 2
  container.scrollTo({ top: Math.max(0, targetTop), behavior: 'smooth' })
}

async function scrollItemListToTop() {
  await nextTick()
  itemListRef.value?.scrollTo({ top: 0, behavior: 'smooth' })
}

function goBack() {
  void router.push('/teacherCenter/scale-library')
}
</script>

<template>
  <div class="autismdev-workbench-page">
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
        <span class="auto-save-status" :class="{ 'is-saving': autoSaveState === 'saving', 'is-saved': autoSaveState === 'saved' }">{{ autoSaveText }}</span>
        <a-button v-if="!isRecordEditMode" size="large" class="outline-action" :loading="saving" :disabled="saving" @click="saveDraft(false)">
          <template #icon><SaveOutlined /></template>
          保存草稿
        </a-button>
        <a-button size="large" type="primary" class="primary-action" :loading="submitting" @click="submitDraft">
          <template #icon><FileDoneOutlined /></template>
          {{ submitActionText }}
        </a-button>
      </div>
    </header>

    <main class="workbench-main">
      <aside ref="itemListRef" class="page-sidebar">
        <div class="sidebar-title">
          <span>领域任务</span>
          <button type="button" class="scope-link" @click="openScopeEditor">
            {{ scopeText }}
            <SlidersOutlined />
          </button>
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
              :data-item-no="item.no"
              @click="goToItem(item.no)"
            >
              <span>第 {{ item.labelNo }} 项</span>
              <strong>{{ item.name }}</strong>
              <CheckCircleFilled v-if="item.status === 'done'" />
              <i v-else-if="item.status === 'active'"></i>
              <b v-else></b>
            </button>
          </div>
        </div>
      </aside>

      <section class="question-panel">
        <a-spin :spinning="templateLoading || itemLoading">
          <template v-if="currentItemSummary">
            <div class="question-content">
              <div class="question-title-row">
                <h1>第 {{ currentItemSummary.domainItemNo || currentItemSummary.itemNo }} 项&nbsp;&nbsp;{{ currentItemTitle }}</h1>
                <a-tag color="blue">{{ currentItemSummary.domainName }}</a-tag>
                <a-button size="small" class="preference-button" @click="openQuestionPreferenceEditor">
                  <template #icon><FilterOutlined /></template>
                  {{ preferenceText }}
                </a-button>
              </div>

              <div class="range-age-row">
                <article class="instruction-card range-age-card">
                  <h2><FileTextOutlined />评估范围</h2>
                  <p>{{ currentRangeText }}</p>
                </article>
                <article class="instruction-card range-age-card">
                  <h2><FileTextOutlined />参考年龄</h2>
                  <p>{{ currentAgeRangeText }}</p>
                </article>
              </div>

              <article class="instruction-card">
                <h2><FileTextOutlined />评估材料</h2>
                <p>{{ normalizeText(currentItem?.materials || currentItemSummary.materials, '暂无材料说明') }}</p>
              </article>

              <article class="instruction-card">
                <h2><FileTextOutlined />评估方法</h2>
                <p>{{ normalizeText(currentItem?.method || currentItemSummary.method, '暂无操作方法') }}</p>
              </article>

              <article class="instruction-card">
                <h2><FileTextOutlined />评分标准</h2>
                <p>{{ normalizeText(currentItem?.passCriteria || currentItemSummary.passCriteria, '暂无评分标准') }}</p>
              </article>
            </div>

            <div class="score-section">
              <div class="score-section__head">
                <h2>评分</h2>
                <span>{{ itemSaveStatusText(selectedItemNo) }}</span>
              </div>
              <div class="score-options">
                <button
                  v-for="option in currentScoreOptions"
                  :key="`${option.scoreType}-${option.value}`"
                  type="button"
                  class="score-option"
                  :class="[`score-${scoreTone(option.value)}`, { 'is-selected': selectedScore === option.value }]"
                  @click="scoreCurrent(option.value)"
                >
                  <span class="score-option__title">
                    <strong>{{ option.value }}</strong>
                    <b>{{ scoreOptionTitle(option) }}</b>
                  </span>
                  <em>{{ option.description || '点击记录本题评分' }}</em>
                  <CheckCircleFilled v-if="selectedScore === option.value" class="score-option__check" />
                </button>
              </div>
            </div>
          </template>
          <div v-else class="empty-state">
            <FileSearchOutlined />
            <strong>暂无可评估题目</strong>
            <span>请检查题目偏好或测评范围设置</span>
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
              <strong>{{ answeredItemCount }} <i>/ {{ totalItemCount }} 项</i></strong>
              <span>缺题</span>
              <strong class="danger">{{ missingItemCount }} <i>项</i></strong>
            </div>
          </div>
        </section>

        <section class="right-card range-card">
          <div class="range-card__head">
            <h3>{{ selectedGroup?.domainName || '当前领域' }}</h3>
            <span>{{ rangeOptions.length }}类</span>
          </div>
          <div class="range-list">
            <button
              type="button"
              class="range-row"
              :class="{ 'is-active': !selectedRangeFilter }"
              @click="selectRangeFilter('')"
            >
              <span>全部</span>
              <b>{{ selectedDomainProgress.answered }}/{{ selectedDomainProgress.total }}</b>
            </button>
            <button
              v-for="option in rangeOptions"
              :key="option.label"
              type="button"
              class="range-row"
              :class="{ 'is-active': selectedRangeFilter === option.label }"
              @click="selectRangeFilter(option.label)"
            >
              <span>{{ option.label }}</span>
              <b>{{ option.done }}/{{ option.total }}</b>
            </button>
          </div>
        </section>

        <section class="right-card remark-card">
          <h3>备注</h3>
          <a-textarea
            v-model:value="currentRemark"
            :auto-size="{ minRows: 3, maxRows: 3 }"
            placeholder="请输入备注"
            @blur="finishItemRemarkEdit"
          />
        </section>
      </aside>
    </main>

    <footer class="workbench-footer">
      <a-button size="large" class="nav-button" :disabled="!hasPreviousItem" @click="goPreviousItem">
        <template #icon><LeftOutlined /></template>
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
      <a-button size="large" class="nav-button" @click="jumpToMissing">
        <template #icon><SwapOutlined /></template>
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
      ok-text="继续测评"
      cancel-text="重新测评"
      :width="430"
      :closable="false"
      :mask-closable="false"
      wrap-class-name="autismdev-draft-modal"
      centered
      @ok="continueExistingDraft"
      @cancel="restartAssessment"
    >
      <div class="draft-resume-tip">
        <p>当前儿童存在一份未提交的孤独症儿童发展评估表草稿。</p>
        <div class="draft-resume-meta">
          <span>已完成：<b>{{ existingDraft?.answeredItemCount || 0 }}</b> / {{ existingDraft?.progress?.itemCount || totalItemCount }} 项</span>
          <span>更新时间：<b>{{ formatDateTime(existingDraft?.updatedTime) }}</b></span>
        </div>
      </div>
    </a-modal>

    <a-modal
      v-model:open="scopeEditorOpen"
      title="测评范围"
      ok-text="应用范围"
      cancel-text="取消"
      :width="560"
      centered
      @ok="applyScopeEdit"
    >
      <div class="scope-editor">
        <a-radio-group v-model:value="draftScopeMode" button-style="solid">
          <a-radio-button value="full">全量测评</a-radio-button>
          <a-radio-button value="custom">自定义领域</a-radio-button>
        </a-radio-group>
        <a-checkbox-group
          v-if="draftScopeMode === 'custom'"
          v-model:value="draftScopeDomainCodes"
          class="scope-checkboxes"
          :options="scopeCheckboxOptions"
        />
        <p class="scope-editor-tip">自定义范围可保存草稿；正式提交仍按后端支持范围校验。</p>
      </div>
    </a-modal>

    <a-modal
      v-model:open="preferenceModalOpen"
      title="题目偏好配置"
      :footer="null"
      :width="520"
      centered
    >
      <div class="preference-dialog">
        <button
          v-for="option in preferenceOptions"
          :key="option.value"
          type="button"
          class="preference-dialog__row"
          :class="{ 'is-active': questionDisplayPreference === option.value }"
          @click="chooseQuestionDisplayPreference(option.value)"
        >
          <b>{{ option.label }}</b>
          <span>{{ option.desc }}</span>
        </button>
        <p>情绪与行为领域始终展示。</p>
      </div>
    </a-modal>
  </div>
</template>

<style scoped lang="less">
.autismdev-workbench-page {
  display: flex;
  flex-direction: column;
  height: 100vh;
  min-height: 0;
  overflow: hidden;
  color: #172033;
  background: #f3f6fb;
}

.workbench-header {
  position: sticky;
  top: 0;
  z-index: 30;
  flex: 0 0 auto;
  display: flex;
  align-items: center;
  min-height: 64px;
  padding: 0 18px;
  background: rgba(255, 255, 255, 0.97);
  border: 1px solid #d8e1ee;
  border-radius: 0 0 12px 12px;
  box-shadow: 0 8px 22px rgba(15, 23, 42, 0.07);
}

.back-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  margin-right: 12px;
  color: #14305f;
  background: #fff;
  border: 1px solid #dbe4f0;
  border-radius: 10px;
  cursor: pointer;
  font-size: 22px;

  &:hover {
    background: #eef4ff;
    color: #155bdc;
  }
}

.workbench-title {
  max-width: 430px;
  overflow: hidden;
  color: #0d2759;
  font-size: 26px;
  font-weight: 900;
  line-height: 1;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.header-divider {
  width: 1px;
  height: 18px;
  margin: 0 12px;
  background: #cbd5e1;
}

.header-meta {
  color: #64748b;
  font-size: 13px;
  font-weight: 700;
  white-space: nowrap;

  b {
    color: #334155;
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
  min-width: 164px;
  color: #94a3b8;
  font-size: 13px;
  font-weight: 700;
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
  min-width: 116px;
  height: 40px;
  padding: 0 14px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 800;
  line-height: 1;
}

.outline-action {
  color: #155bdc;
  border-color: #2f6bff;
}

.primary-action {
  background: #0757e6;
  box-shadow: 0 10px 20px rgba(7, 87, 230, 0.22);
}

.workbench-main {
  display: grid;
  grid-template-columns: 280px minmax(560px, 1fr) 296px;
  gap: 12px;
  flex: 1 1 auto;
  min-height: 0;
  overflow: hidden;
  padding: 12px 12px 0;
}

.domain-sidebar,
.assessment-panel,
.right-sidebar {
  min-height: 0;
  max-height: 100%;
}

.domain-sidebar,
.right-sidebar {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.sidebar-card,
.assessment-panel,
.right-card {
  background: rgba(255, 255, 255, 0.96);
  border: 1px solid #dbe4f0;
  border-radius: 8px;
  box-shadow: 0 10px 22px rgba(15, 23, 42, 0.06);
}

.sidebar-card {
  flex: 1 1 auto;
  min-height: 0;
  overflow: hidden;
}

.sidebar-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 58px;
  padding: 0 12px 0 14px;
  border-bottom: 1px solid #e5ebf3;

  span {
    color: #0f2548;
    font-size: 18px;
    font-weight: 900;
  }

  b {
    color: #475569;
    font-size: 13px;
  }
}

.sidebar-title__actions {
  display: flex;
  gap: 8px;
}

.scope-chip {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  height: 30px;
  padding: 0 10px;
  color: #2563eb;
  background: #fff;
  border: 1px solid #bfdbfe;
  border-radius: 8px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 800;

  &:hover {
    background: #eff6ff;
  }

  &.is-outline {
    color: #475569;
    border-color: #dbe4f0;
  }
}

.domain-scroll {
  height: calc(100% - 58px);
  overflow-y: auto;
  padding: 10px 10px 14px;
  scrollbar-color: #cbd5e1 transparent;
  scrollbar-width: thin;
}

.domain-collapse {
  background: transparent;

  :deep(.ant-collapse-item) {
    --domain-color: #2563eb;
    margin-bottom: 8px;
    overflow: hidden;
    background: #fff;
    border: 1px solid #e1e8f2;
    border-radius: 8px;
  }

  :deep(.ant-collapse-item-active) {
    background: #f7fbff;
    border-color: #93c5fd;
  }

  :deep(.ant-collapse-header) {
    align-items: flex-start;
    padding: 10px 10px 9px 12px !important;
    color: #0f2548;
  }

  :deep(.ant-collapse-expand-icon) {
    height: 28px;
    padding-inline-end: 7px !important;
    color: #475569;
    font-size: 12px;
  }

  :deep(.ant-collapse-header-text) {
    min-width: 0;
    flex: 1 1 auto;
  }

  :deep(.ant-collapse-content-box) {
    padding: 0 8px 8px !important;
  }
}

.domain-collapse-header {
  display: grid;
  grid-template-columns: 10px minmax(0, 1fr) auto;
  gap: 8px;
  align-items: center;
  min-width: 0;

  b {
    color: #64748b;
    font-size: 13px;
    font-weight: 900;
  }
}

.domain-block__dot {
  width: 10px;
  height: 10px;
  background: var(--domain-color);
  border-radius: 999px;
}

.domain-block__name {
  min-width: 0;
  overflow: hidden;
  color: #0f2548;
  font-size: 15px;
  font-weight: 900;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.domain-block__progress {
  grid-column: 2 / 4;
  height: 5px;
  margin-top: 6px;
  overflow: hidden;
  background: #e8edf5;
  border-radius: 999px;

  i {
    display: block;
    height: 100%;
    background: var(--domain-color);
    border-radius: inherit;
  }
}

.domain-item-row {
  display: grid;
  grid-template-columns: 52px minmax(0, 1fr) 22px;
  gap: 10px;
  align-items: center;
  width: 100%;
  min-height: 38px;
  padding: 6px 8px 6px 34px;
  color: #475569;
  text-align: left;
  background: transparent;
  border: 1px solid transparent;
  border-radius: 7px;
  cursor: pointer;

  &:hover {
    background: #f5f8fc;
  }

  &.is-active {
    color: #1d4ed8;
    background: #eff6ff;
  }

  &.is-strong .domain-item-row__no {
    color: #0f8a5f;
  }

  &.is-middle .domain-item-row__no {
    color: #c26816;
  }

  &.is-weak .domain-item-row__no {
    color: #c43d3d;
  }
}

.domain-item-row__no {
  color: #64748b;
  font-size: 13px;
  font-weight: 900;
}

.domain-item-row__title {
  min-width: 0;
  overflow: hidden;
  font-size: 13px;
  font-weight: 800;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.domain-item-row__score,
.domain-item-row__empty {
  justify-self: end;
}

.domain-item-row__score {
  font-style: normal;
  font-weight: 900;
}

.domain-item-row__empty {
  width: 18px;
  height: 18px;
  border: 2px solid #cbd5e1;
  border-radius: 50%;
}

.assessment-panel {
  min-height: 0;
  overflow: hidden;

  :deep(.ant-spin-nested-loading),
  :deep(.ant-spin-container) {
    height: 100%;
  }
}

.workspace-scroll {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
  overflow-y: auto;
  padding: 22px 22px 18px;
  scrollbar-color: #cbd5e1 transparent;
  scrollbar-width: thin;
}

.question-head {
  display: flex;
  gap: 12px;
  align-items: flex-start;
  justify-content: space-between;
  padding-bottom: 18px;

  h1 {
    margin: 8px 0 0;
    color: #0f172a;
    font-size: 28px;
    font-weight: 900;
    line-height: 1.2;
  }
}

.question-kicker {
  color: #64748b;
  font-size: 13px;
  font-weight: 800;
}

.preference-button {
  flex: 0 0 auto;
  height: 38px;
  color: #1d4ed8;
  border-color: #bfdbfe;
  border-radius: 8px;
  font-weight: 800;
}

.question-meta-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 220px;
  gap: 12px;
  margin-bottom: 12px;
}

.question-meta-card {
  min-width: 0;
  padding: 16px 18px;
  background: #fff;
  border: 1px solid #e1e8f2;
  border-radius: 8px;

  span {
    display: block;
    color: #64748b;
    font-size: 14px;
    font-weight: 800;
  }

  b {
    display: block;
    min-width: 0;
    margin-top: 8px;
    overflow: hidden;
    color: #172033;
    font-size: 17px;
    font-weight: 900;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.detail-section {
  flex: 1 1 auto;
  min-height: 0;
  overflow-y: auto;
  padding-right: 2px;
  scrollbar-color: #cbd5e1 transparent;
  scrollbar-width: thin;

  article {
    margin-bottom: 12px;
    padding: 16px 18px;
    background: #fff;
    border: 1px solid #e1e8f2;
    border-radius: 8px;
  }

  h2 {
    margin: 0 0 10px;
    color: #0f2548;
    font-size: 18px;
    font-weight: 900;
  }

  p {
    margin: 0;
    color: #374151;
    font-size: 15px;
    font-weight: 700;
    line-height: 1.65;
    white-space: pre-line;
  }
}

.score-section {
  flex: 0 0 auto;
  padding-top: 4px;
}

.score-section__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 0 0 10px;

  strong {
    color: #0f2548;
    font-size: 18px;
    font-weight: 900;
  }

  span {
    color: #64748b;
    font-size: 13px;
  }
}

.score-option-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.score-option {
  --score-color: #64748b;
  display: flex;
  align-items: center;
  min-width: 0;
  height: 90px;
  padding: 12px 14px;
  text-align: left;
  background: #fff;
  border: 1px solid #dfe7f2;
  border-radius: 8px;
  cursor: pointer;

  &:hover,
  &.is-selected {
    background: #f8fbff;
    border-color: var(--score-color);
  }

  .anticon {
    margin-left: auto;
    color: #cbd5e1;
    font-size: 20px;
  }

  &.is-selected .anticon {
    color: var(--score-color);
  }
}

.score-option__badge {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  margin-right: 12px;
  color: var(--score-color);
  background: #f8fafc;
  border: 1px solid #dbe4f0;
  border-radius: 20px;
  font-size: 21px;
  font-weight: 900;
}

.score-option__text {
  min-width: 0;

  b,
  i {
    display: block;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  b {
    color: #111827;
    font-size: 14px;
    font-weight: 900;
    white-space: nowrap;
  }

  i {
    margin-top: 4px;
    color: #64748b;
    font-size: 12px;
    font-style: normal;
    line-height: 1.3;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
  }
}

.right-sidebar {
  gap: 12px;
  overflow: hidden;
  padding-bottom: 8px;
}

.right-card {
  padding: 14px;

  h3 {
    margin: 0 0 10px;
    color: #0f2548;
    font-size: 18px;
    font-weight: 900;
  }
}

.progress-card,
.remark-card {
  flex: 0 0 auto;
}

.range-card {
  flex: 1 1 auto;
  min-height: 0;
  overflow: hidden;
}

.progress-top {
  display: flex;
  align-items: center;
  gap: 14px;

  strong {
    display: block;
    color: #1d4ed8;
    font-size: 22px;
    font-weight: 900;
  }

  span,
  em {
    display: block;
    color: #64748b;
    font-size: 13px;
    font-style: normal;
    font-weight: 700;
  }
}

.progress-meta {
  display: grid;
  gap: 8px;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #e5ebf3;

  span {
    display: flex;
    justify-content: space-between;
    color: #64748b;
    font-size: 13px;
  }

  b {
    color: #1f2937;
  }
}

.range-card__head {
  display: flex;
  align-items: center;
  justify-content: space-between;

  h3 {
    margin-bottom: 10px;
  }

  span {
    color: #334155;
    font-size: 13px;
    font-weight: 900;
  }
}

.range-list {
  display: grid;
  gap: 8px;
  max-height: calc(100% - 34px);
  overflow-y: auto;
  padding-right: 2px;
  scrollbar-color: #cbd5e1 transparent;
  scrollbar-width: thin;
}

.range-row {
  display: flex;
  align-items: center;
  width: 100%;
  padding: 9px 10px;
  text-align: left;
  background: #fff;
  border: 1px solid #e1e8f2;
  border-radius: 7px;
  cursor: pointer;

  &:hover,
  &.is-active {
    color: #fff;
    background: #2563eb;
    border-color: #2563eb;

    span,
    b {
      color: #fff;
    }
  }
}

.range-row {
  justify-content: space-between;

  span {
    min-width: 0;
    overflow: hidden;
    color: #334155;
    font-size: 13px;
    font-weight: 800;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  b {
    color: #64748b;
    font-size: 12px;
  }
}

.remark-card {
  :deep(.ant-input) {
    border-radius: 7px;
    font-size: 13px;
    background: #fbfdff;
  }
}

.empty-small,
.empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  color: #64748b;
}

.empty-small {
  height: 64px;
  font-size: 13px;
}

.empty-state {
  flex-direction: column;
  height: 100%;

  .anticon {
    color: #94a3b8;
    font-size: 42px;
  }

  strong {
    margin-top: 12px;
    color: #334155;
    font-size: 18px;
    font-weight: 900;
  }

  span {
    margin-top: 6px;
    font-size: 13px;
  }
}

.workbench-footer {
  flex: 0 0 62px;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 0 18px;
  background: rgba(255, 255, 255, 0.97);
  border: 1px solid #d8e1ee;
  border-radius: 10px 10px 0 0;
  box-shadow: 0 -8px 20px rgba(15, 23, 42, 0.06);
}

.footer-button {
  min-width: 112px;
  height: 36px;
  border-radius: 7px;
  font-weight: 800;

  &.is-primary {
    background: #0757e6;
  }
}

.footer-center {
  display: flex;
  align-items: baseline;
  justify-content: center;
  min-width: 96px;
  margin-left: auto;

  b {
    color: #0f172a;
    font-size: 26px;
    font-weight: 900;
  }

  span {
    margin-left: 4px;
    color: #64748b;
    font-size: 15px;
    font-weight: 800;
  }
}

.draft-resume-tip {
  p {
    margin: 0 0 12px;
    color: #334155;
  }
}

.draft-resume-meta {
  display: grid;
  gap: 8px;
  padding: 12px;
  background: #f8fafc;
  border: 1px solid #e5ebf3;
  border-radius: 8px;

  span {
    display: flex;
    justify-content: space-between;
    color: #64748b;
    font-size: 13px;
  }

  b {
    color: #1f2937;
  }
}

.scope-editor {
  display: grid;
  gap: 16px;
}

.scope-checkboxes {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px 12px;
  padding: 14px;
  background: #f8fafc;
  border: 1px solid #e5ebf3;
  border-radius: 8px;
}

.scope-editor-tip {
  margin: 0;
  color: #64748b;
  font-size: 13px;
}

.preference-dialog {
  display: grid;
  gap: 10px;

  p {
    margin: 0;
    color: #64748b;
    font-size: 13px;
  }
}

.preference-dialog__row {
  display: block;
  width: 100%;
  padding: 14px 16px;
  text-align: left;
  background: #fff;
  border: 1px solid #dbe4f0;
  border-radius: 8px;
  cursor: pointer;

  &:hover,
  &.is-active {
    background: #eff6ff;
    border-color: #93c5fd;
  }

  b,
  span {
    display: block;
  }

  b {
    color: #0f2548;
    font-size: 16px;
    font-weight: 900;
  }

  span {
    margin-top: 6px;
    color: #64748b;
    font-size: 13px;
    font-weight: 700;
  }
}

@media (max-width: 1280px) {
  .header-meta:nth-of-type(n + 3),
  .header-divider:nth-of-type(n + 3) {
    display: none;
  }

  .workbench-main {
    grid-template-columns: 248px minmax(460px, 1fr) 268px;
  }

  .score-option-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

/* PEP-3 PC workbench alignment overrides. */
.autismdev-workbench-page {
  margin: 0;
  color: #1f2937;
  background: #f3f5f9;
}

.workbench-header {
  min-height: 52px;
  padding: 0 14px;
  border-color: #d8dfe8;
  border-radius: 0 0 10px 10px;
  box-shadow: 0 8px 22px rgba(15, 23, 42, 0.08);
  backdrop-filter: blur(8px);
}

.back-button {
  width: 28px;
  height: 28px;
  color: #0f2a5f;
  background: transparent;
  border: 0;
  border-radius: 6px;
  font-size: 18px;
}

.workbench-title {
  max-width: none;
  color: #0f2a5f;
  font-size: 18px;
  font-weight: 800;
  line-height: normal;
}

.header-meta {
  color: #111827;
  font-size: 13px;
  font-weight: 400;

  b {
    color: inherit;
    font-weight: 700;
  }
}

.auto-save-status {
  min-width: 190px;
  color: #64748b;
  font-size: 13px;
  font-weight: 400;
  line-height: 20px;
}

.outline-action,
.primary-action {
  min-width: 104px;
  height: 32px;
  padding: 0 14px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 700;

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

.primary-action {
  box-shadow: 0 10px 20px rgba(7, 87, 230, 0.24);
}

.workbench-main {
  grid-template-columns: 240px minmax(420px, 1fr) 256px;
  gap: 10px;
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

.page-sidebar::-webkit-scrollbar,
.question-panel::-webkit-scrollbar,
.score-sidebar::-webkit-scrollbar,
.range-list::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

.page-sidebar::-webkit-scrollbar-track,
.question-panel::-webkit-scrollbar-track,
.score-sidebar::-webkit-scrollbar-track,
.range-list::-webkit-scrollbar-track {
  background: transparent;
}

.page-sidebar::-webkit-scrollbar-thumb,
.question-panel::-webkit-scrollbar-thumb,
.score-sidebar::-webkit-scrollbar-thumb,
.range-list::-webkit-scrollbar-thumb {
  background: #cbd5e1;
  border-radius: 999px;
}

.sidebar-title {
  position: sticky;
  top: 0;
  z-index: 2;
  height: 34px;
  padding: 0 12px 0 16px;
  background: rgba(255, 255, 255, 0.98);
  border-bottom: 1px solid #e5eaf1;
  border-radius: 8px 8px 0 0;
  font-size: 13px;
  font-weight: 700;
  backdrop-filter: blur(8px);

  span {
    color: #111827;
    font-size: 13px;
    font-weight: 700;
  }
}

.scope-link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  height: 24px;
  padding: 0 8px;
  color: #155bdc;
  background: #f7faff;
  border: 1px solid #c8dcff;
  border-radius: 6px;
  cursor: pointer;
  font-size: 12px;
  font-weight: 700;

  &:hover {
    background: #eef4ff;
    border-color: #75a7ff;
  }
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
  min-width: 0;
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
  grid-template-columns: 58px minmax(0, 1fr) 16px;
  align-items: center;
  width: 100%;
  min-height: 30px;
  padding: 0 8px 0 20px;
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
  display: flex;
  flex-direction: column;
  min-height: 0;
  padding: 14px 16px 12px;
  overflow: hidden;

  :deep(.ant-spin-nested-loading),
  :deep(.ant-spin-container) {
    display: flex;
    flex: 1 1 auto;
    flex-direction: column;
    height: 100%;
    min-height: 0;
  }
}

.question-content {
  flex: 1 1 auto;
  min-height: 0;
  overflow-y: auto;
  padding-right: 2px;
  scrollbar-color: #cbd5e1 transparent;
  scrollbar-width: thin;
}

.question-content::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

.question-content::-webkit-scrollbar-track {
  background: transparent;
}

.question-content::-webkit-scrollbar-thumb {
  background: #cbd5e1;
  border-radius: 999px;
}

.question-title-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;

  h1 {
    min-width: 0;
    margin: 0;
    overflow: hidden;
    color: #111827;
    font-size: 20px;
    font-weight: 900;
    line-height: 1.25;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  :deep(.ant-tag) {
    flex: 0 0 auto;
    padding: 2px 8px;
    border-radius: 7px;
    font-size: 12px;
  }
}

.preference-button {
  flex: 0 0 auto;
  height: 28px;
  margin-left: auto;
  color: #155bdc;
  border-color: #c8dcff;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 700;
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

  p {
    margin: 0;
    color: #3f4856;
    font-size: 13px;
    line-height: 1.45;
    white-space: pre-line;
  }
}

.range-age-row {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
  margin-bottom: 8px;

  .instruction-card {
    margin-bottom: 0;
  }
}





.score-section {
  flex: 0 0 auto;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #e5ebf3;

  h2 {
    margin: 0;
    color: #263247;
    font-size: 13px;
    font-weight: 800;
  }
}

.score-section__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 8px;

  span {
    color: #64748b;
    font-size: 12px;
  }
}

.score-options {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 8px;
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
  min-width: 0;
  height: 66px;
  min-height: 66px;
  padding: 8px 34px 8px 12px;
  color: #111827;
  text-align: left;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  cursor: pointer;

  &::after {
    position: absolute;
    top: 12px;
    right: 10px;
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

  em {
    display: block;
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .score-option__title {
    display: flex;
    align-items: baseline;
    gap: 8px;
    max-width: 100%;
    min-width: 0;
    overflow: hidden;
  }

  strong {
    flex: 0 0 auto;
    color: var(--score-color);
    font-size: 18px;
    font-weight: 700;
    line-height: 1.2;
  }

  b {
    min-width: 0;
    overflow: hidden;
    color: #334155;
    font-size: 13px;
    font-weight: 700;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  em {
    margin-top: 3px;
    color: #667085;
    font-size: 12px;
    font-style: normal;
    line-height: 1.3;
    display: -webkit-box;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 1;
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

    &::after {
      opacity: 0;
    }
  }
}

.score-option__check {
  position: absolute;
  top: 10px;
  right: 9px;
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

  h2,
  h3 {
    margin: 0 0 8px;
    color: #0f2548;
    font-size: 13px;
    font-weight: 800;
  }
}

.progress-card,
.remark-card {
  flex: 0 0 auto;
}

.progress-card__body {
  display: grid;
  grid-template-columns: 88px minmax(0, 1fr);
  align-items: center;
  gap: 10px;
  min-height: 88px;
}

.donut {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 82px;
  height: 82px;
  color: #1f2a44;
  border-radius: 50%;
  font-size: 20px;
  font-weight: 700;
  line-height: 1;
  text-align: center;
  white-space: nowrap;
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

.range-card {
  flex: 1 1 auto;
  min-height: 0;
  overflow: hidden;
}

.range-card__head {
  display: flex;
  align-items: center;
  justify-content: space-between;

  span {
    color: #334155;
    font-size: 12px;
    font-weight: 800;
  }
}

.range-list {
  display: grid;
  gap: 8px;
  max-height: calc(100% - 30px);
  overflow-y: auto;
  padding-right: 2px;
  scrollbar-color: #cbd5e1 transparent;
  scrollbar-width: thin;
}

.range-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 8px 10px;
  text-align: left;
  background: #fff;
  border: 1px solid #e1e8f2;
  border-radius: 7px;
  cursor: pointer;

  &:hover,
  &.is-active {
    color: #fff;
    background: #2563eb;
    border-color: #2563eb;

    span,
    b {
      color: #fff;
    }
  }

  span {
    min-width: 0;
    overflow: hidden;
    color: #334155;
    font-size: 13px;
    font-weight: 700;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  b {
    color: #64748b;
    font-size: 12px;
  }
}

.remark-card {
  :deep(.ant-input) {
    background: #fbfdff;
    border-radius: 7px;
    font-size: 13px;
    line-height: 1.45;
    min-height: 72px !important;
    max-height: 72px !important;
    resize: none;
  }
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  flex: 1 1 auto;
  min-height: 0;
  min-height: 420px;
  color: #64748b;

  .anticon {
    color: #94a3b8;
    font-size: 42px;
  }

  strong {
    margin-top: 12px;
    color: #334155;
    font-size: 18px;
    font-weight: 900;
  }

  span {
    margin-top: 6px;
    font-size: 13px;
  }
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

  span {
    display: block;
  }

  b {
    color: #111827;
  }
}

.scope-editor {
  display: grid;
  gap: 16px;
}

.scope-checkboxes {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px 12px;
  padding: 14px;
  background: #f8fafc;
  border: 1px solid #e5ebf3;
  border-radius: 8px;
}

.scope-editor-tip {
  margin: 0;
  color: #64748b;
  font-size: 13px;
}

.preference-dialog {
  display: grid;
  gap: 10px;

  p {
    margin: 0;
    color: #64748b;
    font-size: 13px;
  }
}

.preference-dialog__row {
  display: block;
  width: 100%;
  padding: 14px 16px;
  text-align: left;
  background: #fff;
  border: 1px solid #dbe4f0;
  border-radius: 8px;
  cursor: pointer;

  &:hover,
  &.is-active {
    background: #eff6ff;
    border-color: #93c5fd;
  }

  b,
  span {
    display: block;
  }

  b {
    color: #0f2548;
    font-size: 16px;
    font-weight: 900;
  }

  span {
    margin-top: 6px;
    color: #64748b;
    font-size: 13px;
    font-weight: 700;
  }
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

<style lang="less">
.autismdev-draft-modal {
  .ant-modal-content {
    border-radius: 8px;
  }

  .ant-modal-title {
    color: #111827;
    font-size: 16px;
    font-weight: 700;
  }

  .ant-modal-body {
    padding-top: 14px;
  }

  .ant-modal-footer {
    margin-top: 16px;

    .ant-btn {
      min-width: 86px;
      height: 32px;
      padding: 0 16px;
      border-radius: 6px;
      font-size: 14px;
    }
  }
}
</style>
