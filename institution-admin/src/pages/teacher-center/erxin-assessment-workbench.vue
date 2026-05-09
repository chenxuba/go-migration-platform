<script setup lang="ts">
import {
  AimOutlined,
  AppstoreOutlined,
  ArrowLeftOutlined,
  ArrowRightOutlined,
  BulbOutlined,
  CheckCircleFilled,
  CheckCircleOutlined,
  CheckOutlined,
  CloseOutlined,
  DoubleLeftOutlined,
  DoubleRightOutlined,
  FileDoneOutlined,
  FileTextOutlined,
  HistoryOutlined,
  HighlightOutlined,
  InfoCircleOutlined,
  SaveOutlined,
  SoundOutlined,
  TeamOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import { Modal } from 'ant-design-vue'
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  getERXinAssessmentDraftDetailApi,
  getERXinAssessmentFormTemplateItemApi,
  getERXinAssessmentFormTemplateSummaryApi,
  pageERXinAssessmentDraftsApi,
  saveERXinAssessmentDraftApi,
  saveERXinAssessmentDraftItemApi,
  submitERXinAssessmentDraftApi,
  type ERXinAssessmentAgeGroupSummary,
  type ERXinAssessmentDomain,
  type ERXinAssessmentDraftDetail,
  type ERXinAssessmentDraftSummary,
  type ERXinAssessmentFormTemplateSummary,
  type ERXinAssessmentItem,
  type ERXinAssessmentItemSummary,
  type ERXinDraftInput,
  type ERXinDraftSaveRequest,
} from '@/api/edu-center/erxin-assessment'
import { getScaleAssessmentStudentCandidatesApi } from '@/api/teacher-center/scale-library'
import messageService from '@/utils/messageService'

type DraftItemSaveStatus = 'saving' | 'saved' | 'error'

interface RuleRow {
  label: string
  value: string
  done: boolean
  month?: number
  targetMonths?: number[]
  selected?: boolean
}

const route = useRoute()
const router = useRouter()

const templateLoading = ref(false)
const itemLoading = ref(false)
const saving = ref(false)
const submitting = ref(false)
const draftResumeModalOpen = ref(false)
const detailModalOpen = ref(false)
const allItemsModalOpen = ref(false)
const existingDraft = ref<ERXinAssessmentDraftSummary>()
const selectedDomainCode = ref('')
const selectedItemNo = ref(0)
const template = ref<ERXinAssessmentFormTemplateSummary>()
const currentProgress = ref<ERXinAssessmentDraftDetail['progress']>()
const itemCache = reactive<Record<number, ERXinAssessmentItem>>({})
const itemPasses = reactive<Record<number, boolean>>({})
const itemRemarks = reactive<Record<number, string>>({})
const previousStartIndexByDomain = reactive<Record<string, number>>({})
const futureEndIndexByDomain = reactive<Record<string, number>>({})
const futureVisibleDomains = ref<string[]>([])
const reviewMonthByDomain = reactive<Record<string, number | undefined>>({})
const draftItemSaveStatus = ref<Record<number, DraftItemSaveStatus>>({})
const draftItemSaveErrors = ref<Record<number, string>>({})
const autoSaveLastSavedAt = ref('')
const mainScrollRef = ref<HTMLElement | null>(null)
const ruleMonthListRef = ref<HTMLElement | null>(null)
const overviewScrollRef = ref<HTMLElement | null>(null)
const workspaceFlashMonths = ref<number[]>([])
const ruleFlashMonths = ref<number[]>([])
let draftCreationPromise: Promise<ERXinAssessmentDraftDetail | undefined> | undefined
let workspaceFlashTimer: number | undefined
let ruleFlashTimer: number | undefined
let workspaceFlashSerial = 0

const editor = reactive({
  id: numberFromQuery('draftId') || undefined as number | undefined,
  studentId: numberFromQuery('childId') || undefined as number | undefined,
  studentName: textFromQuery('childName'),
  examinerName: textFromQuery('examinerName'),
  birthDate: normalizeDateText(textFromQuery('childBirthDate') || textFromQuery('birthDate')),
  assessmentDate: normalizeDateText(textFromQuery('assessmentDate')) || dayjs().format('YYYY-MM-DD'),
  remark: '',
})

const domains = computed(() => template.value?.domains || [])
const ageGroups = computed(() => template.value?.ageGroups || [])
const standardAgeMonths = computed(() => [...new Set(ageGroups.value.map(group => Number(group.ageMonth)).filter(month => month > 0))].sort((a, b) => a - b))
const allItems = computed(() => ageGroups.value.flatMap(group => group.items || []))
const studentName = computed(() => editor.studentName || textFromQuery('childName') || '-')
const examinerName = computed(() => editor.examinerName || '当前老师')
const assessmentDateText = computed(() => formatDate(editor.assessmentDate))
const studentAge = computed(() => assessmentAgeText(editor.birthDate, editor.assessmentDate) || textFromQuery('childAge') || '-')
const mainAgeMonth = computed(() => {
  const months = actualAgeMonths(editor.birthDate, editor.assessmentDate)
  const ages = standardAgeMonths.value
  if (!ages.length || months <= 0)
    return 0
  return ages.reduce((best, ageMonth) => Math.abs(months - ageMonth) < Math.abs(months - best) ? ageMonth : best, ages[0])
})
const mainAgeIndex = computed(() => standardAgeMonths.value.indexOf(mainAgeMonth.value))
const defaultPreviousStartIndex = computed(() => mainAgeIndex.value < 0 ? 0 : Math.max(0, mainAgeIndex.value - 2))
const selectedDomainName = computed(() => domainName(selectedDomainCode.value))
const currentItemSummary = computed(() => summaryByNo(selectedItemNo.value))
const currentItem = computed(() => itemCache[selectedItemNo.value])
const visibleMonths = computed(() => visibleMonthsForDomain(selectedDomainCode.value))
const centerMonths = computed(() => centerMonthsForDomain(selectedDomainCode.value))
const workspacePreviousMonths = computed(() => previousMonthsForDomain(selectedDomainCode.value))
const workspaceFutureMonths = computed(() => futureMonthsForDomain(selectedDomainCode.value))
const reviewMonth = computed(() => reviewMonthByDomain[selectedDomainCode.value])
const isReviewingRecord = computed(() => reviewMonthByDomain[selectedDomainCode.value] !== undefined)
const completedDomainCount = computed(() => domains.value.filter(domain => domainStopRuleComplete(domain.domainCode)).length)
const savedItemCount = computed(() => Object.keys(itemPasses).length)
const selectedDomainProgress = computed(() => domainProgress(selectedDomainCode.value))
const selectedDomainCompletionText = computed(() => {
  if (domainStopRuleComplete(selectedDomainCode.value))
    return '已完成'
  const progress = selectedDomainProgress.value
  if (progress.total > 0 && progress.answered >= progress.total)
    return '当前可见完成（待推进）'
  if (progress.answered > 0)
    return '测查中'
  return '待测'
})
const autoSaveState = computed<'idle' | 'saving' | 'saved'>(() => {
  const statuses = Object.values(draftItemSaveStatus.value)
  if (statuses.some(status => status === 'saving'))
    return 'saving'
  return autoSaveLastSavedAt.value ? 'saved' : 'idle'
})
const autoSaveText = computed(() => {
  if (autoSaveState.value === 'saving')
    return '自动保存中...'
  if (autoSaveState.value === 'saved')
    return `已自动保存 ${autoSaveLastSavedAt.value}`
  return ''
})
const nextActionText = computed(() => buildNextActionText())
const nextActionIcon = computed(() => resolveNextActionIcon())
const ruleProgressHint = computed(() => buildRuleProgressHint())
const ruleRows = computed(() => recordRowsForDomain(selectedDomainCode.value))
const ruleMonthRows = computed(() => ruleRows.value.filter(row => row.month !== undefined))
const rulePinnedRows = computed(() => ruleRows.value.filter(row => row.month === undefined))
const overviewAgeGroups = computed(() => [...ageGroups.value].filter(group => (group.items || []).length).sort((left, right) => left.ageMonth - right.ageMonth))
const overviewSections = computed(() => chunkAgeGroups(overviewAgeGroups.value, 5))
const overviewSortedDomains = computed(() => [...domains.value].sort((left, right) => (left.sortNo || 0) - (right.sortNo || 0)))
const overviewTotalItemCount = computed(() => overviewAgeGroups.value.reduce((total, group) => total + (group.items || []).length, 0))
const overviewTargetAgeMonth = computed(() => {
  const selectedGroup = overviewAgeGroups.value.find(group => (group.items || []).some(item => item.itemNo === selectedItemNo.value))
  return selectedGroup?.ageMonth || mainAgeMonth.value
})
const overviewTargetSectionIndex = computed(() => {
  const index = overviewSections.value.findIndex(section => section.some(group => group.ageMonth === overviewTargetAgeMonth.value))
  return index < 0 ? 0 : index
})
const currentRecordRemark = computed(() => itemRemarks[selectedItemNo.value] || '')
const currentDetailTitle = computed(() => {
  const item = currentItem.value || currentItemSummary.value
  const itemNo = item?.itemNo || selectedItemNo.value
  const title = itemTitle(item)
  if (!itemNo)
    return '当前题目说明'
  return `${itemNo} ${title}${item?.parentReportAllowed ? '（R）' : ''}`
})

watch(selectedItemNo, (itemNo) => {
  if (itemNo > 0)
    void fetchItemDetail(itemNo)
})

watch(selectedDomainCode, () => {
  workspaceFlashMonths.value = []
  ruleFlashMonths.value = []
})

onMounted(() => {
  void initializeWorkbench()
})

onBeforeUnmount(() => {
  if (workspaceFlashTimer)
    window.clearInterval(workspaceFlashTimer)
  if (ruleFlashTimer)
    window.clearTimeout(ruleFlashTimer)
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

function actualAgeMonths(birthDate?: string, assessmentDate?: string) {
  const birth = dayjs(birthDate).startOf('day')
  const target = dayjs(assessmentDate).startOf('day')
  if (!birth.isValid() || !target.isValid() || birth.isAfter(target))
    return 0
  return target.diff(birth, 'day') / 30.4375
}

function normalizeText(value?: string, fallback = '-') {
  const text = String(value || '').replace(/\s+/g, ' ').trim()
  return text || fallback
}

async function initializeWorkbench() {
  await fetchTemplate()
  if (editor.id) {
    await fetchDraftDetail(editor.id)
    await hydrateStudentBirthDate()
    selectInitialItem()
    return
  }
  await hydrateStudentBirthDate()
  const draft = await findExistingDraft()
  if (draft) {
    existingDraft.value = draft
    draftResumeModalOpen.value = true
    selectInitialItem()
    return
  }
  await startNewAssessment()
}

async function fetchTemplate() {
  templateLoading.value = true
  try {
    const res = await getERXinAssessmentFormTemplateSummaryApi()
    template.value = unwrap<ERXinAssessmentFormTemplateSummary>(res)
    selectedDomainCode.value = template.value.domains?.[0]?.domainCode || ''
  }
  catch (error: any) {
    messageService.error(getErrorMessage(error, '获取儿心量表题目目录失败'))
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
    const res = await getERXinAssessmentFormTemplateItemApi(itemNo)
    itemCache[itemNo] = unwrap<ERXinAssessmentItem>(res)
  }
  catch (error: any) {
    messageService.error(getErrorMessage(error, `获取第${itemNo}题失败`))
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
      scaleCode: 'ERXIN2',
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
    const res = await pageERXinAssessmentDraftsApi({
      pageRequestModel: { pageIndex: 1, pageSize: 1 },
      queryModel: {
        assessmentCode: 'ERXIN2',
        studentId: editor.studentId,
        latestOnly: true,
      },
      latestOnly: true,
    })
    const data = unwrap<any>(res)
    return (data?.items || [])[0] as ERXinAssessmentDraftSummary | undefined
  }
  catch (error: any) {
    messageService.error(getErrorMessage(error, '查询儿心未完成草稿失败'))
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
    const res = await getERXinAssessmentDraftDetailApi(id)
    applyDraftDetail(unwrap<ERXinAssessmentDraftDetail>(res))
  }
  catch (error: any) {
    messageService.error(getErrorMessage(error, '获取儿心测评草稿失败'))
  }
}

function applyDraftDetail(detail: ERXinAssessmentDraftDetail) {
  editor.id = detail.id
  editor.studentId = detail.studentId || editor.studentId
  editor.studentName = detail.studentName || editor.studentName
  editor.examinerName = detail.examinerName || editor.examinerName
  editor.birthDate = normalizeDateText(detail.birthDate || detail.input?.birthDate) || editor.birthDate
  editor.assessmentDate = normalizeDateText(detail.assessmentDate || detail.input?.assessmentDate) || editor.assessmentDate
  editor.remark = detail.remark || detail.input?.remark || ''
  currentProgress.value = detail.progress
  applyDraftInput(detail.input)
  restoreAssessmentWindowsFromAnswers()
  autoSaveLastSavedAt.value = detail.updatedTime ? dayjs(detail.updatedTime).format('MM-DD HH:mm') : autoSaveLastSavedAt.value
}

function applyDraftInput(input?: ERXinDraftInput) {
  Object.keys(itemPasses).forEach(key => delete itemPasses[Number(key)])
  Object.keys(itemRemarks).forEach(key => delete itemRemarks[Number(key)])
  if (!input)
    return
  Object.entries(input.itemPasses || {}).forEach(([itemNo, passed]) => {
    if (typeof passed === 'boolean')
      itemPasses[Number(itemNo)] = passed
  })
  ;(input.itemPassList || []).forEach((item) => {
    if (item.itemNo > 0 && typeof item.passed === 'boolean')
      itemPasses[item.itemNo] = item.passed
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
}

function resetAssessmentInput() {
  editor.id = undefined
  editor.remark = ''
  Object.keys(itemPasses).forEach(key => delete itemPasses[Number(key)])
  Object.keys(itemRemarks).forEach(key => delete itemRemarks[Number(key)])
  Object.keys(previousStartIndexByDomain).forEach(key => delete previousStartIndexByDomain[key])
  Object.keys(futureEndIndexByDomain).forEach(key => delete futureEndIndexByDomain[key])
  Object.keys(reviewMonthByDomain).forEach(key => delete reviewMonthByDomain[key])
  futureVisibleDomains.value = []
  currentProgress.value = undefined
  autoSaveLastSavedAt.value = ''
  draftItemSaveStatus.value = {}
  draftItemSaveErrors.value = {}
}

function selectInitialItem(preferredItemNo = 0) {
  if (preferredItemNo && summaryByNo(preferredItemNo)) {
    const summary = summaryByNo(preferredItemNo)
    selectedDomainCode.value = summary?.domainCode || selectedDomainCode.value
    selectedItemNo.value = preferredItemNo
    return
  }
  if (!selectedDomainCode.value)
    selectedDomainCode.value = domains.value[0]?.domainCode || ''
  selectedItemNo.value = firstCurrentItemNo(selectedDomainCode.value) || firstVisibleItemNo(selectedDomainCode.value)
}

function validateDraftHeader(silent = false) {
  if (!editor.studentId || studentName.value === '-') {
    if (!silent)
      messageService.warning('缺少真实儿童，无法保存测评')
    return false
  }
  return true
}

function buildPayload(): ERXinDraftSaveRequest {
  const itemPassList = Object.entries(itemPasses)
    .map(([itemNo, passed]) => ({
      itemNo: Number(itemNo),
      passed: Boolean(passed),
      remark: itemRemarks[Number(itemNo)]?.trim() || '',
    }))
    .sort((a, b) => a.itemNo - b.itemNo)
  const itemRemarkList = Object.entries(itemRemarks)
    .filter(([, remark]) => remark.trim())
    .map(([itemNo, remark]) => ({ itemNo: Number(itemNo), remark: remark.trim() }))
    .sort((a, b) => a.itemNo - b.itemNo)
  return {
    id: editor.id,
    studentId: editor.studentId,
    studentName: studentName.value === '-' ? undefined : studentName.value,
    examinerName: editor.examinerName,
    birthDate: editor.birthDate,
    assessmentDate: editor.assessmentDate,
    remark: editor.remark,
    itemPassList,
    itemRemarkList,
  }
}

async function saveDraft(silent = false) {
  if (!validateDraftHeader(silent))
    return undefined
  saving.value = true
  try {
    const res = await saveERXinAssessmentDraftApi(buildPayload())
    const detail = unwrap<ERXinAssessmentDraftDetail>(res)
    applyDraftDetail(detail)
    if (!silent)
      messageService.success('儿心草稿已保存')
    return detail
  }
  catch (error: any) {
    messageService.error(getErrorMessage(error, '保存儿心草稿失败'))
    return undefined
  }
  finally {
    saving.value = false
  }
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

async function saveItem(itemNo: number) {
  if (itemNo <= 0 || !hasPass(itemNo))
    return
  draftItemSaveStatus.value = { ...draftItemSaveStatus.value, [itemNo]: 'saving' }
  try {
    const canSave = await ensureDraftForItemSave()
    if (!canSave || !editor.id)
      throw new Error('草稿创建失败')
    const res = await saveERXinAssessmentDraftItemApi({
      draftId: editor.id,
      itemNo,
      passed: itemPasses[itemNo],
      remark: itemRemarks[itemNo]?.trim() || '',
    })
    const detail = unwrap<ERXinAssessmentDraftDetail>(res)
    editor.id = detail.id
    currentProgress.value = detail.progress
    autoSaveLastSavedAt.value = dayjs().format('MM-DD HH:mm')
    draftItemSaveStatus.value = { ...draftItemSaveStatus.value, [itemNo]: 'saved' }
  }
  catch (error: any) {
    const message = getErrorMessage(error, `第${itemNo}题自动保存失败`)
    draftItemSaveStatus.value = { ...draftItemSaveStatus.value, [itemNo]: 'error' }
    draftItemSaveErrors.value = { ...draftItemSaveErrors.value, [itemNo]: message }
    messageService.error(message)
  }
}

async function submitDraft() {
  const blocker = localSubmitBlocker()
  if (blocker) {
    messageService.warning(blocker)
    return
  }
  submitting.value = true
  try {
    const detail = await saveDraft(true)
    const draftId = detail?.id || editor.id
    if (!draftId) {
      messageService.warning('请先保存草稿，再提交正式记录')
      return
    }
    await submitERXinAssessmentDraftApi(draftId)
    messageService.success('已提交正式测评记录')
    await router.push('/teacherCenter/evaluationRecord')
  }
  catch (error: any) {
    messageService.error(getErrorMessage(error, '提交儿心测评记录失败'))
  }
  finally {
    submitting.value = false
  }
}

function localSubmitBlocker() {
  if (!editor.birthDate || !editor.assessmentDate)
    return '缺少出生日期或测查日期，不能提交正式记录'
  for (const domain of domains.value) {
    const code = domain.domainCode
    for (const month of visibleMonthsForDomain(code)) {
      for (const item of itemsFor(code, month)) {
        if (!hasPass(item.itemNo)) {
          selectedDomainCode.value = code
          selectedItemNo.value = item.itemNo
          return `${domainName(code)}还有当前可见题目未记录，请补全后再提交`
        }
      }
    }
    if (canContinuePreviousMonthsForDomain(code)) {
      selectedDomainCode.value = code
      return `${domainName(code)}尚未形成连续两个往前月龄全通过，请继续往前测查`
    }
    if (canEnterFutureMonthsForDomain(code)) {
      selectedDomainCode.value = code
      return hasPreviousBaselineForDomain(code)
        ? `${domainName(code)}已建立前测基线，请先进入往后测查`
        : `${domainName(code)}已测至最低月龄，请先进入往后测查`
    }
    if (canContinueFutureMonthsForDomain(code)) {
      selectedDomainCode.value = code
      return `${domainName(code)}尚未形成连续两个往后月龄全不通过，请继续往后测查`
    }
    if (!domainStopRuleComplete(code)) {
      selectedDomainCode.value = code
      return `${domainName(code)}尚未满足儿心量表停止规则，请完成规则提示的测查`
    }
  }
  return ''
}

function selectDomain(code: string) {
  selectedDomainCode.value = code
  selectedItemNo.value = firstCurrentItemNo(code) || firstVisibleItemNo(code)
  void scrollMainToTop()
}

async function scrollMainToTop() {
  await nextTick()
  mainScrollRef.value?.scrollTo({ top: 0, behavior: 'smooth' })
}

function ruleRowHasUnmetResult(row: RuleRow) {
  if (row.done || row.month === undefined)
    return false
  return row.value.includes('未全') || row.value.includes('未通过')
}

function ruleRowClickable(row: RuleRow) {
  return row.month !== undefined || (row.done && Boolean(row.targetMonths?.length))
}

function revealRuleTargets(row: RuleRow) {
  if (row.month !== undefined) {
    openAssessmentRecord(row.month)
    return
  }
  if (!row.done || !row.targetMonths?.length)
    return
  revealRuleMonths(row.targetMonths)
  void revealWorkspaceMonths(row.targetMonths)
}

function revealRuleMonths(months: number[]) {
  const targetMonths = [...new Set(months)]
  if (!targetMonths.length)
    return
  const targetIndex = ruleMonthRows.value.findIndex(row => row.month !== undefined && targetMonths.includes(row.month))
  if (targetIndex < 0)
    return

  if (ruleFlashTimer)
    window.clearTimeout(ruleFlashTimer)
  ruleFlashMonths.value = targetMonths
  ruleFlashTimer = window.setTimeout(() => {
    ruleFlashMonths.value = []
    ruleFlashTimer = undefined
  }, 900)

  nextTick(() => {
    const container = ruleMonthListRef.value
    if (!container)
      return
    const maxOffset = Math.max(0, container.scrollHeight - container.clientHeight)
    const targetOffset = Math.min(maxOffset, Math.max(0, 6 + targetIndex * 40))
    container.scrollTo({ top: targetOffset, behavior: 'smooth' })
  })
}

async function revealWorkspaceMonths(months: number[]) {
  const rawTargets = [...new Set(months)]
  if (!rawTargets.length)
    return
  if (reviewMonthByDomain[selectedDomainCode.value] !== undefined) {
    delete reviewMonthByDomain[selectedDomainCode.value]
    await nextTick()
  }
  const targets = centerMonths.value.filter(month => rawTargets.includes(month))
  if (!targets.length)
    return
  flashWorkspaceMonths(targets)
  await nextTick()
  scrollWorkspaceMonthIntoView(targets[0])
}

function flashWorkspaceMonths(months: number[]) {
  workspaceFlashSerial += 1
  const serial = workspaceFlashSerial
  const targetMonths = [...months]
  if (workspaceFlashTimer)
    window.clearInterval(workspaceFlashTimer)
  workspaceFlashMonths.value = targetMonths
  let ticks = 0
  workspaceFlashTimer = window.setInterval(() => {
    if (serial !== workspaceFlashSerial) {
      window.clearInterval(workspaceFlashTimer)
      return
    }
    ticks += 1
    workspaceFlashMonths.value = ticks % 2 === 1 ? [] : targetMonths
    if (ticks >= 6) {
      window.clearInterval(workspaceFlashTimer)
      workspaceFlashTimer = undefined
      workspaceFlashMonths.value = []
    }
  }, 220)
}

function scrollWorkspaceMonthIntoView(month: number) {
  const container = mainScrollRef.value
  if (!container)
    return
  const key = `${selectedDomainCode.value}-${month}`
  const target = Array.from(container.querySelectorAll<HTMLElement>('[data-domain-month]'))
    .find(element => element.dataset.domainMonth === key)
  if (!target)
    return
  const containerRect = container.getBoundingClientRect()
  const targetRect = target.getBoundingClientRect()
  const targetTop = container.scrollTop + targetRect.top - containerRect.top
  container.scrollTo({ top: Math.max(0, targetTop), behavior: 'smooth' })
}

async function revealSelectedItem() {
  await nextTick()
  const container = mainScrollRef.value
  const itemNo = selectedItemNo.value
  if (!container || itemNo <= 0)
    return
  const activeItem = container.querySelector<HTMLElement>(`[data-item-no="${itemNo}"]`)
  if (!activeItem)
    return
  const activeCenter = activeItem.offsetTop + activeItem.offsetHeight / 2
  const targetTop = activeCenter - container.clientHeight / 2
  const maxTop = Math.max(0, container.scrollHeight - container.clientHeight)
  container.scrollTo({
    top: Math.max(0, Math.min(maxTop, targetTop)),
    behavior: 'smooth',
  })
}

function selectItem(itemNo: number) {
  if (!summaryByNo(itemNo))
    return
  selectedItemNo.value = itemNo
}

function scoreItem(itemNo: number, passed: boolean) {
  const summary = summaryByNo(itemNo)
  const domainCode = summary?.domainCode || selectedDomainCode.value
  const historyReview = reviewMonthByDomain[domainCode] !== undefined && summary?.ageMonth === reviewMonthByDomain[domainCode]
  const changedExisting = hasPass(itemNo) && itemPasses[itemNo] !== passed
  if (historyReview && changedExisting) {
    Modal.confirm({
      title: '确认修改历史题目？',
      content: `第 ${itemNo} 题已记录过结果，修改后会重新判断本能区测查规则。`,
      okText: '确认修改',
      cancelText: '取消',
      centered: true,
      onOk: () => applyScore(itemNo, passed, domainCode),
    })
    return
  }
  applyScore(itemNo, passed, domainCode)
}

function applyScore(itemNo: number, passed: boolean, domainCode: string) {
  itemPasses[itemNo] = passed
  reconcileAfterScore(domainCode)
  selectedDomainCode.value = domainCode
  const nextSelected = nextSelectedItemNoForDomain(domainCode, itemNo)
  selectedItemNo.value = nextSelected > 0 ? nextSelected : itemNo
  void revealSelectedItem()
  void saveItem(itemNo)
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
  if (itemNo > 0 && hasPass(itemNo))
    void saveItem(itemNo)
}

function openAssessmentRecord(month: number) {
  reviewMonthByDomain[selectedDomainCode.value] = month
  selectedItemNo.value = firstItemNoForMonth(selectedDomainCode.value, month)
  void revealSelectedItem()
}

function returnToCurrentAssessment() {
  delete reviewMonthByDomain[selectedDomainCode.value]
  selectedItemNo.value = firstCurrentItemNo(selectedDomainCode.value) || firstVisibleItemNo(selectedDomainCode.value)
  void revealSelectedItem()
}

function locateCurrentAssessmentItem() {
  delete reviewMonthByDomain[selectedDomainCode.value]
  selectedItemNo.value = firstPendingAssessmentItemNoForDomain(selectedDomainCode.value) || firstCurrentItemNo(selectedDomainCode.value) || firstVisibleItemNo(selectedDomainCode.value)
  void revealSelectedItem()
}

function continuePreviousMonths() {
  const domainCode = selectedDomainCode.value
  const currentStart = previousStartIndexForDomain(domainCode)
  if (currentStart <= 0)
    return
  let start = currentStart
  while (start > 0) {
    const lowestVisibleMonth = standardAgeMonths.value[start]
    const step = ageMonthAllPassed(domainCode, lowestVisibleMonth) ? 1 : 2
    const nextStart = Math.max(0, start - step)
    if (nextStart === start)
      break
    previousStartIndexByDomain[domainCode] = nextStart
    start = nextStart
    if (firstPendingCurrentItemNo(domainCode) > 0 || hasPreviousBaselineForDomain(domainCode) || !previousMonthsCompleteForDomain(domainCode))
      break
  }
  delete reviewMonthByDomain[domainCode]
  selectedItemNo.value = firstPendingCurrentItemNo(domainCode) || firstCurrentItemNo(domainCode) || firstVisibleItemNo(domainCode)
  void revealRuleMonths(standardAgeMonths.value.slice(start, currentStart).reverse())
  void revealSelectedItem()
}

function enterFutureMonths() {
  const index = mainAgeIndex.value
  const domainCode = selectedDomainCode.value
  if (index < 0 || index >= standardAgeMonths.value.length - 1)
    return
  const endIndex = Math.min(standardAgeMonths.value.length - 1, index + 2)
  delete reviewMonthByDomain[domainCode]
  if (!futureVisibleDomains.value.includes(domainCode))
    futureVisibleDomains.value = [...futureVisibleDomains.value, domainCode]
  futureEndIndexByDomain[domainCode] = endIndex
  selectedItemNo.value = firstItemNoForMonth(domainCode, standardAgeMonths.value[index + 1])
  void revealRuleMonths(standardAgeMonths.value.slice(index + 1, endIndex + 1))
  void revealSelectedItem()
}

function continueFutureMonths() {
  const index = mainAgeIndex.value
  const domainCode = selectedDomainCode.value
  if (index < 0)
    return
  const currentEnd = futureEndIndexByDomain[domainCode] ?? Math.min(standardAgeMonths.value.length - 1, index + 2)
  if (currentEnd >= standardAgeMonths.value.length - 1)
    return
  const highestVisibleMonth = standardAgeMonths.value[currentEnd]
  const step = ageMonthAllFailed(domainCode, highestVisibleMonth) ? 1 : 2
  const nextEnd = Math.min(standardAgeMonths.value.length - 1, currentEnd + step)
  delete reviewMonthByDomain[domainCode]
  futureEndIndexByDomain[domainCode] = nextEnd
  selectedItemNo.value = firstItemNoForMonth(domainCode, standardAgeMonths.value[currentEnd + 1])
  void revealRuleMonths(standardAgeMonths.value.slice(currentEnd + 1, nextEnd + 1))
  void revealSelectedItem()
}

function ruleAction() {
  if (canContinuePreviousMonthsForDomain(selectedDomainCode.value)) {
    continuePreviousMonths()
    return
  }
  if (canEnterFutureMonthsForDomain(selectedDomainCode.value)) {
    enterFutureMonths()
    return
  }
  if (canContinueFutureMonthsForDomain(selectedDomainCode.value)) {
    continueFutureMonths()
    return
  }
  if (firstPendingAssessmentItemNoForDomain(selectedDomainCode.value) > 0)
    locateCurrentAssessmentItem()
}

function ruleActionLabel() {
  if (canContinuePreviousMonthsForDomain(selectedDomainCode.value))
    return '继续往前测查'
  if (canContinueFutureMonthsForDomain(selectedDomainCode.value))
    return '继续往后测查'
  if (firstPendingAssessmentItemNoForDomain(selectedDomainCode.value) > 0)
    return futureVisibleDomains.value.includes(selectedDomainCode.value) ? '定位往后测查' : '定位当前题目'
  if (domainStopRuleComplete(selectedDomainCode.value))
    return '本能区测查完成'
  return '进入往后测查'
}

function resolveNextActionIcon() {
  if (canContinuePreviousMonthsForDomain(selectedDomainCode.value))
    return DoubleLeftOutlined
  if (canContinueFutureMonthsForDomain(selectedDomainCode.value))
    return DoubleRightOutlined
  if (firstPendingAssessmentItemNoForDomain(selectedDomainCode.value) > 0)
    return AimOutlined
  if (domainStopRuleComplete(selectedDomainCode.value))
    return CheckCircleOutlined
  return ArrowRightOutlined
}

function ruleActionDisabled() {
  return !canContinuePreviousMonthsForDomain(selectedDomainCode.value)
    && !canEnterFutureMonthsForDomain(selectedDomainCode.value)
    && !canContinueFutureMonthsForDomain(selectedDomainCode.value)
    && !firstPendingAssessmentItemNoForDomain(selectedDomainCode.value)
}

function previousStartIndexForDomain(domainCode: string) {
  const index = mainAgeIndex.value
  if (index < 0)
    return 0
  return previousStartIndexByDomain[domainCode] ?? defaultPreviousStartIndex.value
}

function previousMonthsForDomain(domainCode: string) {
  const index = mainAgeIndex.value
  if (index <= 0)
    return []
  const start = previousStartIndexForDomain(domainCode)
  return standardAgeMonths.value.slice(start, index)
}

function futureMonthsForDomain(domainCode: string) {
  if (!futureVisibleDomains.value.includes(domainCode))
    return []
  const index = mainAgeIndex.value
  if (index < 0)
    return []
  const end = Math.min(Math.max(futureEndIndexByDomain[domainCode] ?? Math.min(standardAgeMonths.value.length - 1, index + 2), index), standardAgeMonths.value.length - 1)
  if (index + 1 > end)
    return []
  return standardAgeMonths.value.slice(index + 1, end + 1)
}

function visibleMonthsBeforeFutureForDomain(domainCode: string) {
  return [mainAgeMonth.value, ...previousMonthsForDomain(domainCode).reverse()].filter(Boolean)
}

function visibleMonthsForDomain(domainCode: string) {
  return [...visibleMonthsBeforeFutureForDomain(domainCode), ...futureMonthsForDomain(domainCode)]
}

function centerMonthsForDomain(domainCode: string) {
  const reviewMonth = reviewMonthByDomain[domainCode]
  if (reviewMonth !== undefined)
    return [reviewMonth]
  return visibleMonthsForDomain(domainCode).filter(month => itemsFor(domainCode, month).length)
}

function recordMonthsForDomain(domainCode: string) {
  const months = new Set<number>(visibleMonthsForDomain(domainCode))
  for (const group of ageGroups.value) {
    const hasAnsweredItem = (group.items || []).some(item => item.domainCode === domainCode && hasPass(item.itemNo))
    if (hasAnsweredItem)
      months.add(group.ageMonth)
  }
  const mainAge = mainAgeMonth.value
  const previous = [...months].filter(month => month > 0 && month < mainAge).sort((a, b) => b - a)
  const future = [...months].filter(month => month > mainAge).sort((a, b) => a - b)
  return [months.has(mainAge) ? mainAge : 0, ...previous, ...future].filter(Boolean)
}

function previousMonthsCompleteForDomain(domainCode: string) {
  const previous = previousMonthsForDomain(domainCode)
  return previous.length > 0 && previous.every(month => ageMonthComplete(domainCode, month))
}

function previousBaselineMonthsForDomain(domainCode: string) {
  const mainIndex = mainAgeIndex.value
  if (mainIndex <= 1)
    return []
  const currentStart = previousStartIndexForDomain(domainCode)
  for (let index = mainIndex - 2; index >= currentStart; index--) {
    const lowerMonth = standardAgeMonths.value[index]
    const upperMonth = standardAgeMonths.value[index + 1]
    if (ageMonthAllPassed(domainCode, lowerMonth) && ageMonthAllPassed(domainCode, upperMonth))
      return [upperMonth, lowerMonth]
  }
  return []
}

function hasPreviousBaselineForDomain(domainCode: string) {
  return previousBaselineMonthsForDomain(domainCode).length > 0
}

function mainMonthCompleteForDomain(domainCode: string) {
  return mainAgeMonth.value > 0 && ageMonthComplete(domainCode, mainAgeMonth.value)
}

function canContinuePreviousMonthsForDomain(domainCode: string) {
  if (hasPreviousBaselineForDomain(domainCode) || !previousMonthsCompleteForDomain(domainCode) || !mainMonthCompleteForDomain(domainCode) || previousStartIndexForDomain(domainCode) <= 0)
    return false
  return true
}

function previousBoundaryStopForDomain(domainCode: string) {
  if (hasPreviousBaselineForDomain(domainCode) || previousStartIndexForDomain(domainCode) > 0)
    return false
  const previous = previousMonthsForDomain(domainCode)
  return previous.length === 0 || previousMonthsCompleteForDomain(domainCode)
}

function previousSearchResolvedForDomain(domainCode: string) {
  return hasPreviousBaselineForDomain(domainCode) || previousBoundaryStopForDomain(domainCode)
}

function canEnterFutureMonthsForDomain(domainCode: string) {
  return !futureVisibleDomains.value.includes(domainCode)
    && mainAgeIndex.value < standardAgeMonths.value.length - 1
    && mainMonthCompleteForDomain(domainCode)
    && previousSearchResolvedForDomain(domainCode)
}

function futureMonthsCompleteForDomain(domainCode: string) {
  const future = futureMonthsForDomain(domainCode)
  return future.length > 0 && future.every(month => ageMonthComplete(domainCode, month))
}

function futureCeilingMonthsForDomain(domainCode: string) {
  const future = futureMonthsForDomain(domainCode)
  for (let index = 0; index < future.length - 1; index++) {
    const current = future[index]
    const next = future[index + 1]
    const currentIndex = standardAgeMonths.value.indexOf(current)
    const nextIndex = standardAgeMonths.value.indexOf(next)
    if (nextIndex - currentIndex !== 1)
      continue
    if (ageMonthAllFailed(domainCode, current) && ageMonthAllFailed(domainCode, next))
      return [current, next]
  }
  return []
}

function hasFutureCeilingForDomain(domainCode: string) {
  return futureCeilingMonthsForDomain(domainCode).length > 0
}

function canContinueFutureMonthsForDomain(domainCode: string) {
  if (!futureVisibleDomains.value.includes(domainCode) || hasFutureCeilingForDomain(domainCode) || !futureMonthsCompleteForDomain(domainCode))
    return false
  const index = mainAgeIndex.value
  const end = futureEndIndexByDomain[domainCode] ?? Math.min(standardAgeMonths.value.length - 1, index + 2)
  return end < standardAgeMonths.value.length - 1
}

function futureBoundaryStopForDomain(domainCode: string) {
  const mainIndex = mainAgeIndex.value
  if (mainIndex < 0)
    return false
  if (!futureVisibleDomains.value.includes(domainCode))
    return mainIndex >= standardAgeMonths.value.length - 1
  if (!futureMonthsCompleteForDomain(domainCode))
    return false
  const end = futureEndIndexByDomain[domainCode] ?? Math.min(standardAgeMonths.value.length - 1, mainIndex + 2)
  return end >= standardAgeMonths.value.length - 1 && !hasFutureCeilingForDomain(domainCode)
}

function futureSearchResolvedForDomain(domainCode: string) {
  return hasFutureCeilingForDomain(domainCode) || futureBoundaryStopForDomain(domainCode)
}

function domainStopRuleComplete(domainCode: string) {
  if (!mainMonthCompleteForDomain(domainCode))
    return false
  if (!previousSearchResolvedForDomain(domainCode))
    return false
  return futureSearchResolvedForDomain(domainCode)
}

function recordRowsForDomain(domainCode: string): RuleRow[] {
  const mainAge = mainAgeMonth.value
  const reviewMonth = reviewMonthByDomain[domainCode]
  const baselineMonths = previousBaselineMonthsForDomain(domainCode)
  const ceilingMonths = futureCeilingMonthsForDomain(domainCode)
  const previousBoundaryStop = previousBoundaryStopForDomain(domainCode)
  const futureBoundaryStop = futureBoundaryStopForDomain(domainCode)
  const rows: RuleRow[] = recordMonthsForDomain(domainCode).map(month => ({
    label: month === mainAge ? `主测${month}月龄` : month < mainAge ? `往前${month}月龄` : `往后${month}月龄`,
    value: recordStatusText(domainCode, month),
    done: recordDone(domainCode, month),
    month,
    selected: reviewMonth === month,
  }))
  rows.push({
    label: '前测基线',
    value: hasPreviousBaselineForDomain(domainCode) ? '已建立' : previousBoundaryStop ? '已到最低月龄' : '未形成',
    done: hasPreviousBaselineForDomain(domainCode) || previousBoundaryStop,
    targetMonths: baselineMonths,
  })
  if (futureVisibleDomains.value.includes(domainCode) || hasAnsweredFutureMonth(domainCode) || futureBoundaryStop) {
    rows.push({
      label: '后测封顶',
      value: hasFutureCeilingForDomain(domainCode) ? '已建立' : futureBoundaryStop ? '已到最高月龄' : '未形成',
      done: hasFutureCeilingForDomain(domainCode) || futureBoundaryStop,
      targetMonths: ceilingMonths,
    })
  }
  return rows
}

function recordStatusText(domainCode: string, month: number) {
  if (month === mainAgeMonth.value)
    return ageMonthComplete(domainCode, month) ? '已完成' : '未完成'
  if (month < mainAgeMonth.value)
    return ageMonthAllPassed(domainCode, month) ? '全通过' : ageMonthComplete(domainCode, month) ? '未全通过' : '未完成'
  return ageMonthAllFailed(domainCode, month) ? '全不通过' : ageMonthComplete(domainCode, month) ? '未全不通过' : '未完成'
}

function recordDone(domainCode: string, month: number) {
  if (month === mainAgeMonth.value)
    return ageMonthComplete(domainCode, month)
  if (month < mainAgeMonth.value)
    return ageMonthAllPassed(domainCode, month)
  return ageMonthAllFailed(domainCode, month)
}

function buildNextActionText() {
  for (const month of visibleMonths.value) {
    for (const item of itemsFor(selectedDomainCode.value, month)) {
      if (!hasPass(item.itemNo))
        return `完成${month}月龄第${item.itemNo}题`
    }
  }
  if (canEnterFutureMonthsForDomain(selectedDomainCode.value))
    return hasPreviousBaselineForDomain(selectedDomainCode.value) ? '前测已达标，可以进入往后测查' : '已到最低月龄，前测强行停止，可进入往后测查'
  if (canContinuePreviousMonthsForDomain(selectedDomainCode.value)) {
    const currentStart = previousStartIndexForDomain(selectedDomainCode.value)
    const lowestVisibleMonth = standardAgeMonths.value[currentStart]
    const step = ageMonthAllPassed(selectedDomainCode.value, lowestVisibleMonth) ? 1 : 2
    const nextStart = Math.max(0, currentStart - step)
    const nextMonths = standardAgeMonths.value.slice(nextStart, currentStart)
    return `前测未形成连续全通过，继续追加${nextMonths.join('月、')}月`
  }
  if (canContinueFutureMonthsForDomain(selectedDomainCode.value)) {
    const index = mainAgeIndex.value
    const currentEnd = futureEndIndexByDomain[selectedDomainCode.value] ?? Math.min(standardAgeMonths.value.length - 1, index + 2)
    const highestVisibleMonth = standardAgeMonths.value[currentEnd]
    const step = ageMonthAllFailed(selectedDomainCode.value, highestVisibleMonth) ? 1 : 2
    const nextEnd = Math.min(standardAgeMonths.value.length - 1, currentEnd + step)
    const nextMonths = standardAgeMonths.value.slice(currentEnd + 1, nextEnd + 1)
    return `后测未形成连续全不通过，继续追加${nextMonths.join('月、')}月`
  }
  if (hasFutureCeilingForDomain(selectedDomainCode.value) || domainStopRuleComplete(selectedDomainCode.value))
    return '当前能区停止规则已满足，可切换下一个能区或提交'
  if (futureVisibleDomains.value.includes(selectedDomainCode.value) && !hasFutureCeilingForDomain(selectedDomainCode.value))
    return futureBoundaryStopForDomain(selectedDomainCode.value) ? '已到最高月龄，后测强行停止' : futureMonthsCompleteForDomain(selectedDomainCode.value) ? '已到最高可追测月龄，仍未形成连续全不通过' : '先完成当前可见的往后测查题目'
  if (!hasPreviousBaselineForDomain(selectedDomainCode.value))
    return previousBoundaryStopForDomain(selectedDomainCode.value) ? '已到最低可追测月龄，仍未形成连续全通过' : previousMonthsCompleteForDomain(selectedDomainCode.value) ? '已到最低可追测月龄，仍未形成连续全通过' : '先完成当前可见的往前测查题目'
  return '当前可见题目已完成'
}

function buildRuleProgressHint() {
  const domainCode = selectedDomainCode.value
  if (hasFutureCeilingForDomain(domainCode))
    return '往后测查已形成连续两个标准月龄全不通过，本能区达到停止规则。'
  if (futureBoundaryStopForDomain(domainCode))
    return '已测至最高标准月龄，仍未形成后测封顶，按边界规则强行停止。'
  if (hasPreviousBaselineForDomain(domainCode))
    return '前测已形成连续两个标准月龄全通过，继续往后寻找连续两个标准月龄全不通过。'
  if (previousBoundaryStopForDomain(domainCode))
    return '已测至最低标准月龄，仍未形成前测基线，按边界规则进入往后测查。'
  return '前测尚未形成连续两个标准月龄全通过，需继续向更低月龄追测。'
}

function hasAnsweredFutureMonth(domainCode: string) {
  return highestAnsweredFutureIndex(domainCode) >= 0
}

function highestAnsweredFutureIndex(domainCode: string) {
  let highest = -1
  for (const group of ageGroups.value) {
    const index = standardAgeMonths.value.indexOf(group.ageMonth)
    if (index <= mainAgeIndex.value)
      continue
    const hasAnswered = group.items.some(item => item.domainCode === domainCode && hasPass(item.itemNo))
    if (hasAnswered)
      highest = Math.max(highest, index)
  }
  return highest
}

function reconcileAfterScore(domainCode: string) {
  trimPreviousWindowToActiveBaseline(domainCode)
  const previousReady = mainMonthCompleteForDomain(domainCode) && previousSearchResolvedForDomain(domainCode)
  if (!previousReady) {
    futureVisibleDomains.value = futureVisibleDomains.value.filter(code => code !== domainCode)
    delete futureEndIndexByDomain[domainCode]
    return
  }
  const highestAnsweredFuture = highestAnsweredFutureIndex(domainCode)
  if (!futureVisibleDomains.value.includes(domainCode) && highestAnsweredFuture < 0) {
    delete futureEndIndexByDomain[domainCode]
    return
  }
  if (!futureVisibleDomains.value.includes(domainCode))
    futureVisibleDomains.value = [...futureVisibleDomains.value, domainCode]
  const defaultEnd = Math.min(standardAgeMonths.value.length - 1, mainAgeIndex.value + 2)
  const existingEnd = futureEndIndexByDomain[domainCode] ?? defaultEnd
  const end = Math.max(existingEnd, highestAnsweredFuture)
  const normalizedEnd = Math.max(defaultEnd, Math.min(end, standardAgeMonths.value.length - 1))
  futureEndIndexByDomain[domainCode] = trimFutureEndToActiveCeiling(domainCode, normalizedEnd)
}

function trimPreviousWindowToActiveBaseline(domainCode: string) {
  const mainIndex = mainAgeIndex.value
  if (mainIndex <= 0)
    return
  const currentStart = previousStartIndexForDomain(domainCode)
  let bestStart: number | undefined
  for (let index = mainIndex - 2; index >= currentStart; index--) {
    const current = standardAgeMonths.value[index]
    const next = standardAgeMonths.value[index + 1]
    if (ageMonthAllPassed(domainCode, current) && ageMonthAllPassed(domainCode, next)) {
      bestStart = index
      break
    }
  }
  if (bestStart !== undefined && bestStart !== currentStart)
    previousStartIndexByDomain[domainCode] = bestStart
}

function trimFutureEndToActiveCeiling(domainCode: string, endIndex: number) {
  const mainIndex = mainAgeIndex.value
  if (mainIndex < 0 || endIndex <= mainIndex + 1)
    return endIndex
  for (let index = mainIndex + 1; index < endIndex; index++) {
    const current = standardAgeMonths.value[index]
    const next = standardAgeMonths.value[index + 1]
    if (ageMonthAllFailed(domainCode, current) && ageMonthAllFailed(domainCode, next))
      return index + 1
  }
  return endIndex
}

function restoreAssessmentWindowsFromAnswers() {
  Object.keys(previousStartIndexByDomain).forEach(key => delete previousStartIndexByDomain[key])
  Object.keys(futureEndIndexByDomain).forEach(key => delete futureEndIndexByDomain[key])
  Object.keys(reviewMonthByDomain).forEach(key => delete reviewMonthByDomain[key])
  futureVisibleDomains.value = []
  const mainIndex = mainAgeIndex.value
  if (mainIndex < 0)
    return
  for (const domain of domains.value) {
    const domainCode = domain.domainCode
    let previousStart = defaultPreviousStartIndex.value
    let hasPreviousAnswer = false
    let futureEnd = Math.min(standardAgeMonths.value.length - 1, mainIndex + 2)
    let hasFutureAnswer = false
    for (const group of ageGroups.value) {
      const ageIndex = standardAgeMonths.value.indexOf(group.ageMonth)
      if (ageIndex < 0)
        continue
      const hasAnsweredItem = group.items.some(item => item.domainCode === domainCode && hasPass(item.itemNo))
      if (!hasAnsweredItem)
        continue
      if (ageIndex < mainIndex) {
        hasPreviousAnswer = true
        previousStart = Math.min(previousStart, ageIndex)
      }
      else if (ageIndex > mainIndex) {
        hasFutureAnswer = true
        futureEnd = Math.max(futureEnd, ageIndex)
      }
    }
    if (hasPreviousAnswer)
      previousStartIndexByDomain[domainCode] = previousStart
    if (hasFutureAnswer) {
      futureVisibleDomains.value = [...new Set([...futureVisibleDomains.value, domainCode])]
      futureEndIndexByDomain[domainCode] = Math.min(futureEnd, standardAgeMonths.value.length - 1)
    }
    trimPreviousWindowToActiveBaseline(domainCode)
    if (futureVisibleDomains.value.includes(domainCode))
      futureEndIndexByDomain[domainCode] = trimFutureEndToActiveCeiling(domainCode, futureEndIndexByDomain[domainCode] ?? futureEnd)
  }
}

function itemsFor(domainCode: string, ageMonth: number) {
  const group = ageGroups.value.find(item => item.ageMonth === ageMonth)
  return (group?.items || []).filter(item => item.domainCode === domainCode)
}

function ageMonthComplete(domainCode: string, ageMonth: number) {
  const items = itemsFor(domainCode, ageMonth)
  return items.length > 0 && items.every(item => hasPass(item.itemNo))
}

function ageMonthAllPassed(domainCode: string, ageMonth: number) {
  const items = itemsFor(domainCode, ageMonth)
  return items.length > 0 && items.every(item => itemPasses[item.itemNo] === true)
}

function ageMonthAllFailed(domainCode: string, ageMonth: number) {
  const items = itemsFor(domainCode, ageMonth)
  return items.length > 0 && items.every(item => itemPasses[item.itemNo] === false)
}

function hasPass(itemNo: number) {
  return typeof itemPasses[itemNo] === 'boolean'
}

function firstVisibleItemNo(domainCode: string) {
  for (const month of visibleMonthsForDomain(domainCode)) {
    const itemNo = firstItemNoForMonth(domainCode, month)
    if (itemNo > 0)
      return itemNo
  }
  return 0
}

function firstCurrentItemNo(domainCode: string) {
  return firstPendingCurrentItemNo(domainCode) || centerMonthsForDomain(domainCode).map(month => firstItemNoForMonth(domainCode, month)).find(Boolean) || 0
}

function firstPendingCurrentItemNo(domainCode: string) {
  for (const month of centerMonthsForDomain(domainCode)) {
    for (const item of itemsFor(domainCode, month)) {
      if (!hasPass(item.itemNo))
        return item.itemNo
    }
  }
  return 0
}

function firstPendingAssessmentItemNoForDomain(domainCode: string) {
  for (const month of visibleMonthsForDomain(domainCode)) {
    for (const item of itemsFor(domainCode, month)) {
      if (!hasPass(item.itemNo))
        return item.itemNo
    }
  }
  return 0
}

function nextSelectedItemNoForDomain(domainCode: string, fallbackItemNo: number) {
  if (reviewMonthByDomain[domainCode] !== undefined)
    return fallbackItemNo
  return firstPendingCurrentItemNo(domainCode) || fallbackItemNo
}

function firstItemNoForMonth(domainCode: string, ageMonth: number) {
  return itemsFor(domainCode, ageMonth)[0]?.itemNo || 0
}

function summaryByNo(itemNo: number): ERXinAssessmentItemSummary | undefined {
  return allItems.value.find(item => item.itemNo === itemNo)
}

function domainName(domainCode: string) {
  return domains.value.find(domain => domain.domainCode === domainCode)?.domainName || domainCode || '-'
}

function domainProgress(domainCode: string) {
  const items = visibleMonthsForDomain(domainCode).flatMap(month => itemsFor(domainCode, month))
  const answered = items.filter(item => hasPass(item.itemNo)).length
  return { answered, total: items.length, percent: items.length ? Math.round((answered / items.length) * 100) : 0 }
}

function domainVisibleComplete(domainCode: string) {
  const progress = domainProgress(domainCode)
  return progress.total > 0 && progress.answered >= progress.total
}

function domainStatusText(domainCode: string) {
  if (domainStopRuleComplete(domainCode))
    return '已完成'
  if (domainVisibleComplete(domainCode))
    return '待推进'
  const progress = domainProgress(domainCode)
  if (progress.answered > 0)
    return '测查中'
  return '待测'
}

function domainStatusClass(domainCode: string) {
  const completed = domainStopRuleComplete(domainCode)
  const progress = domainProgress(domainCode)
  const visibleComplete = progress.total > 0 && progress.answered >= progress.total
  return {
    'is-complete': completed,
    'is-active': selectedDomainCode.value === domainCode && !completed,
    'is-ready': !completed && visibleComplete,
  }
}

function domainIcon(domain: { domainCode: string, domainName: string }) {
  const code = domain.domainCode.toUpperCase()
  const name = domain.domainName
  if (code === 'GM' || name.includes('大运动'))
    return ThunderboltOutlined
  if (code === 'FM' || name.includes('精细'))
    return HighlightOutlined
  if (code === 'AD' || name.includes('适应'))
    return BulbOutlined
  if (code === 'LANG' || name.includes('语言'))
    return SoundOutlined
  if (code === 'SOC' || name.includes('社会') || name.includes('社交'))
    return TeamOutlined
  if (code.includes('AIM'))
    return AimOutlined
  return AppstoreOutlined
}

function itemTitle(item?: ERXinAssessmentItem | ERXinAssessmentItemSummary) {
  return normalizeText(item?.itemTitle || item?.testItem)
}

function itemStatusClass(item: ERXinAssessmentItemSummary) {
  if (item.itemNo === selectedItemNo.value)
    return 'is-active'
  if (!hasPass(item.itemNo))
    return 'is-todo'
  return itemPasses[item.itemNo] ? 'is-pass' : 'is-fail'
}

function overviewItemStatusClass(item: ERXinAssessmentItemSummary) {
  if (item.itemNo === selectedItemNo.value)
    return 'is-selected'
  if (!hasPass(item.itemNo))
    return 'is-empty'
  return itemPasses[item.itemNo] ? 'is-pass' : 'is-fail'
}

function overviewItemsForDomain(group: ERXinAssessmentAgeGroupSummary, domainCode: string) {
  return (group.items || [])
    .filter(item => item.domainCode === domainCode)
    .sort((left, right) => left.itemNo - right.itemNo)
}

function chunkAgeGroups(groups: ERXinAssessmentAgeGroupSummary[], size: number) {
  const sections: ERXinAssessmentAgeGroupSummary[][] = []
  for (let index = 0; index < groups.length; index += size)
    sections.push(groups.slice(index, index + size))
  return sections
}

function paperDomainName(value: string) {
  const text = value.trim()
  if (text.length <= 3)
    return text
  if (text.includes('大运动'))
    return '大运动'
  if (text.includes('精细'))
    return '精细动作'
  if (text.includes('适应'))
    return '适应能力'
  if (text.includes('语言'))
    return '语言'
  if (text.includes('社会'))
    return '社会行为'
  return text
}

function openAllItemsModal() {
  allItemsModalOpen.value = true
}

function monthAnsweredCount(domainCode: string, month: number) {
  return itemsFor(domainCode, month).filter(item => hasPass(item.itemNo)).length
}

function shouldShowPreviousDivider(index: number) {
  return !isReviewingRecord.value && index === 1 && workspacePreviousMonths.value.length > 0
}

function shouldShowFutureDivider(month: number) {
  return !isReviewingRecord.value && workspaceFutureMonths.value.length > 0 && month === workspaceFutureMonths.value[0]
}

function openDetailModal() {
  detailModalOpen.value = true
}

function goBack() {
  void router.push('/teacherCenter/scale-library')
}
</script>

<template>
  <div class="erxin-workbench-page">
    <header class="workbench-header">
      <button type="button" class="back-button" aria-label="返回量表库" @click="goBack">
        <ArrowLeftOutlined />
      </button>
      <strong class="workbench-title">儿心量表-II 测评工作台</strong>
      <span class="header-divider"></span>
      <span class="header-meta">儿童：<b>{{ studentName }}</b></span>
      <span class="header-divider"></span>
      <span class="header-meta">实足年龄：<b>{{ studentAge }}</b></span>
      <span class="header-divider"></span>
      <span class="header-meta">主测月龄：<b>{{ mainAgeMonth ? `${mainAgeMonth}月` : '-' }}</b></span>
      <span class="header-divider"></span>
      <span class="header-meta">测查日期：<b>{{ assessmentDateText }}</b></span>
      <div class="header-actions">
        <span v-if="autoSaveText" class="auto-save-status" :class="{ 'is-saving': autoSaveState === 'saving', 'is-saved': autoSaveState === 'saved' }">{{ autoSaveText }}</span>
        <a-button size="large" class="outline-action" :loading="saving" @click="saveDraft(false)">
          <template #icon><SaveOutlined /></template>
          保存草稿
        </a-button>
        <a-button size="large" type="primary" class="primary-action" :loading="submitting" @click="submitDraft">
          <template #icon><FileDoneOutlined /></template>
          提交记录
        </a-button>
      </div>
    </header>

    <main class="workbench-main">
      <aside class="domain-sidebar">
        <div class="sidebar-title">
          <span>能区进度</span>
          <b>{{ completedDomainCount }}/{{ domains.length }}</b>
        </div>
        <div class="domain-list">
          <button
            v-for="domain in domains"
            :key="domain.domainCode"
            type="button"
            class="domain-card"
            :class="{ 'is-active': domain.domainCode === selectedDomainCode, 'is-complete': domainStopRuleComplete(domain.domainCode) }"
            @click="selectDomain(domain.domainCode)"
          >
            <div class="domain-card__top">
              <span class="domain-card__icon">
                <component :is="domainIcon(domain)" />
              </span>
              <span class="domain-card__name">{{ domain.domainName }}</span>
              <span class="domain-card__count">
                {{ domainProgress(domain.domainCode).answered }}/{{ domainProgress(domain.domainCode).total }}
              </span>
            </div>
            <div class="domain-card__bottom">
              <div class="domain-card__progress">
                <i :style="{ width: `${domainProgress(domain.domainCode).percent}%` }"></i>
              </div>
              <span class="domain-card__status" :class="domainStatusClass(domain.domainCode)">
                {{ domainStatusText(domain.domainCode) }}
              </span>
            </div>
          </button>
        </div>
        <button type="button" class="all-items-button" @click="openAllItemsModal">
          <FileTextOutlined />
          <span>查看全部题目</span>
          <b>›</b>
        </button>
        <div class="progress-summary">
          <strong>完成情况</strong>
          <span>本能区：{{ selectedDomainCompletionText }}</span>
          <span>全量表：{{ completedDomainCount }}/{{ domains.length }} 能区完成</span>
          <span>已保存{{ savedItemCount }}题</span>
        </div>
      </aside>

      <section class="assessment-panel">
        <a-spin :spinning="templateLoading || itemLoading">
          <div class="workspace-area">
            <div class="center-head">
              <h1>
                {{ isReviewingRecord && reviewMonth !== undefined ? `${selectedDomainName} · ${reviewMonth}月龄记录` : `${selectedDomainName} · 当前测查` }}
              </h1>
              <div class="center-head__actions">
                <button v-if="isReviewingRecord" type="button" class="return-current-button" @click="returnToCurrentAssessment">
                  返回当前测查
                </button>
                <span class="small-badge">主测月龄 {{ mainAgeMonth || '-' }}月龄</span>
              </div>
            </div>

            <div ref="mainScrollRef" class="workspace-scroll">
              <div v-if="!centerMonths.length" class="current-empty-state">
                <FileDoneOutlined />
                <strong>当前题目已完成</strong>
                <span>请按右侧规则继续推进测查</span>
              </div>
              <div v-else class="month-stack">
                <template v-for="(month, index) in centerMonths" :key="`${selectedDomainCode}-${month}`">
                  <div v-if="shouldShowPreviousDivider(index)" class="workspace-divider">
                    <span></span>
                    <b>往前测查</b>
                    <span></span>
                  </div>
                  <div v-if="shouldShowFutureDivider(month)" class="workspace-divider">
                    <span></span>
                    <b>往后测查</b>
                    <span></span>
                  </div>

                  <section
                    class="month-section"
                    :class="{ 'is-flashing': workspaceFlashMonths.includes(month) }"
                    :data-domain-month="`${selectedDomainCode}-${month}`"
                  >
                    <div class="month-section__head" :class="{ 'is-main-age': month === mainAgeMonth }">
                      <strong>{{ month }}月龄</strong>
                      <span v-if="month === mainAgeMonth" class="month-main-tag">主测月龄</span>
                      <em>已测 {{ monthAnsweredCount(selectedDomainCode, month) }}/{{ itemsFor(selectedDomainCode, month).length }}</em>
                    </div>
                    <div class="item-list">
                      <div
                        v-for="(item, itemIndex) in itemsFor(selectedDomainCode, month)"
                        :key="item.itemNo"
                        class="item-row"
                        :class="[itemStatusClass(item), { 'has-divider': itemIndex < itemsFor(selectedDomainCode, month).length - 1 }]"
                        :data-item-no="item.itemNo"
                        @click="selectItem(item.itemNo)"
                      >
                        <span class="item-no">{{ item.itemNo }}</span>
                        <div class="item-title">
                          <strong>{{ itemTitle(item) }}</strong>
                          <i v-if="item.parentReportAllowed">R</i>
                          <i v-if="item.attentionIfFailed" class="is-warning">*</i>
                        </div>
                        <button
                          type="button"
                          class="inline-score is-pass"
                          :class="{ 'is-selected': itemPasses[item.itemNo] === true }"
                          @click.stop="scoreItem(item.itemNo, true)"
                        >
                          <CheckOutlined />
                          通过
                        </button>
                        <button
                          type="button"
                          class="inline-score is-fail"
                          :class="{ 'is-selected': itemPasses[item.itemNo] === false }"
                          @click.stop="scoreItem(item.itemNo, false)"
                        >
                          <CloseOutlined />
                          不通过
                        </button>
                      </div>
                    </div>
                  </section>
                </template>
              </div>
            </div>
          </div>

          <section class="detail-panel">
            <div class="detail-panel__head">
              <strong>当前题目说明：{{ currentDetailTitle }}</strong>
              <a-button size="small" class="full-detail-button" @click="openDetailModal">
                <template #icon><FileTextOutlined /></template>
                完整说明
              </a-button>
            </div>
            <div class="detail-boxes">
              <article class="detail-text-box">
                <h2>操作方法</h2>
                <p>{{ normalizeText(currentItem?.method, itemLoading ? '正在加载题目操作方法...' : '暂无内容') }}</p>
              </article>
              <article class="detail-text-box">
                <h2>通过标准</h2>
                <p>{{ normalizeText(currentItem?.passCriteria, itemLoading ? '正在加载通过标准...' : '暂无内容') }}</p>
              </article>
            </div>
          </section>
        </a-spin>
      </section>

      <aside class="rule-sidebar">
        <section class="right-card rule-card">
          <h3>规则判断</h3>
          <div class="rule-body">
            <div class="next-action">
              <component :is="nextActionIcon" />
              <div>
                <span>下一步</span>
                <strong>{{ nextActionText }}</strong>
              </div>
            </div>
            <h3>测评记录</h3>
            <div class="rule-list">
              <div ref="ruleMonthListRef" class="rule-month-list">
                <button
                  v-for="row in ruleMonthRows"
                  :key="`${row.label}-${row.month || row.targetMonths?.join('-') || ''}`"
                  type="button"
                  class="rule-row"
                  :class="{
                    'is-done': row.done,
                    'is-selected': row.selected,
                    'is-unmet': ruleRowHasUnmetResult(row),
                    'is-flashing': row.month !== undefined && ruleFlashMonths.includes(row.month),
                    'is-clickable': ruleRowClickable(row),
                  }"
                  :disabled="!ruleRowClickable(row)"
                  @click="revealRuleTargets(row)"
                >
                  <CheckCircleFilled v-if="row.done" />
                  <CloseOutlined v-else-if="ruleRowHasUnmetResult(row)" />
                  <InfoCircleOutlined v-else />
                  <span>{{ row.label }}</span>
                  <b>{{ row.value }}</b>
                  <HistoryOutlined />
                </button>
              </div>
              <div v-if="rulePinnedRows.length" class="rule-pinned-list">
                <button
                  v-for="row in rulePinnedRows"
                  :key="`${row.label}-${row.targetMonths?.join('-') || ''}`"
                  type="button"
                  class="rule-row is-pinned"
                  :class="{
                    'is-done': row.done,
                    'is-unmet': ruleRowHasUnmetResult(row),
                    'is-clickable': ruleRowClickable(row),
                  }"
                  :disabled="!ruleRowClickable(row)"
                  @click="revealRuleTargets(row)"
                >
                  <CheckCircleFilled v-if="row.done" />
                  <CloseOutlined v-else-if="ruleRowHasUnmetResult(row)" />
                  <InfoCircleOutlined v-else />
                  <span>{{ row.label }}</span>
                  <b>{{ row.value }}</b>
                  <AimOutlined v-if="ruleRowClickable(row)" />
                </button>
              </div>
            </div>
            <h3>测查推进</h3>
            <p class="rule-hint">{{ ruleProgressHint }}</p>
            <a-button block type="primary" class="rule-action" :disabled="ruleActionDisabled()" @click="ruleAction">
              {{ ruleActionLabel() }}
            </a-button>
          </div>
          <div class="right-remark-section">
            <h3>题目备注</h3>
            <a-textarea
              :value="currentRecordRemark"
              :auto-size="{ minRows: 4, maxRows: 4 }"
              placeholder="添加本题备注"
              @change="event => updateItemRemark(event.target.value)"
              @blur="finishItemRemarkEdit"
            />
          </div>
        </section>
      </aside>
    </main>

    <a-modal
      v-model:open="detailModalOpen"
      title="当前题目完整说明"
      :footer="null"
      :width="720"
      centered
    >
      <div class="detail-modal-body">
        <h3>{{ currentDetailTitle }}</h3>
        <section>
          <strong>操作方法</strong>
          <p>{{ normalizeText(currentItem?.method, '暂无内容') }}</p>
        </section>
        <section>
          <strong>通过标准</strong>
          <p>{{ normalizeText(currentItem?.passCriteria, '暂无内容') }}</p>
        </section>
      </div>
    </a-modal>

    <a-modal
      v-model:open="allItemsModalOpen"
      title="全部题目总览"
      :footer="null"
      :width="1280"
      centered
      wrap-class-name="erxin-all-items-modal"
      :body-style="{ padding: 0 }"
    >
      <template #title>
        <div class="overview-title">
          <strong>全部题目总览</strong>
          <span>已测 {{ savedItemCount }}/{{ overviewTotalItemCount }}</span>
          <span v-if="mainAgeMonth">主测月龄 {{ mainAgeMonth }}月龄</span>
        </div>
      </template>
      <div ref="overviewScrollRef" class="overview-scroll">
        <div v-if="!overviewSections.length" class="overview-empty">
          暂无题目
        </div>
        <template v-else>
          <section
            v-for="(section, sectionIndex) in overviewSections"
            :key="section.map(group => group.ageMonth).join('-')"
            class="overview-section"
            :class="{ 'is-target': sectionIndex === overviewTargetSectionIndex }"
            :data-overview-section="sectionIndex"
          >
            <table class="overview-table">
              <colgroup>
                <col class="overview-domain-col">
                <col v-for="group in section" :key="group.ageMonth">
              </colgroup>
              <thead>
                <tr>
                  <th>项目</th>
                  <th
                    v-for="group in section"
                    :key="group.ageMonth"
                    :class="{ 'is-main-age': group.ageMonth === mainAgeMonth }"
                  >
                    {{ group.ageMonth }} 月龄
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="domain in overviewSortedDomains" :key="`${sectionIndex}-${domain.domainCode}`">
                  <th>{{ paperDomainName(domain.domainName) }}</th>
                  <td v-for="group in section" :key="`${domain.domainCode}-${group.ageMonth}`">
                    <div
                      v-for="item in overviewItemsForDomain(group, domain.domainCode)"
                      :key="item.itemNo"
                      class="overview-item"
                      :class="overviewItemStatusClass(item)"
                    >
                      <span class="overview-check"></span>
                      <p>
                        <b>{{ item.itemNo }}</b>
                        {{ itemTitle(item) }}
                        <i v-if="item.parentReportAllowed">R</i>
                      </p>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </section>
        </template>
      </div>
    </a-modal>

    <a-modal
      v-model:open="draftResumeModalOpen"
      title="发现未完成草稿"
      ok-text="继续测评"
      cancel-text="重新测评"
      :width="430"
      :closable="false"
      :mask-closable="false"
      centered
      @ok="continueExistingDraft"
      @cancel="restartAssessment"
    >
      <div class="draft-resume-tip">
        <p>当前儿童存在一份未提交的儿心量表测评草稿。</p>
        <div class="draft-resume-meta">
          <span>已记录：<b>{{ existingDraft?.answeredItemCount || 0 }}</b> 题</span>
          <span>更新时间：<b>{{ formatDateTime(existingDraft?.updatedTime) }}</b></span>
        </div>
      </div>
    </a-modal>
  </div>
</template>

<style scoped lang="less">
.erxin-workbench-page {
  display: flex;
  flex-direction: column;
  height: 100vh;
  min-height: 0;
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
  font-size: 14px;
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
  grid-template-columns: 214px minmax(520px, 1fr) 296px;
  gap: 10px;
  flex: 1 1 auto;
  min-height: 0;
  overflow: hidden;
  padding: 10px 10px 0;
}

.domain-sidebar,
.assessment-panel,
.rule-sidebar {
  min-height: 0;
  max-height: 100%;
  overscroll-behavior: contain;
  scrollbar-color: #cbd5e1 transparent;
  scrollbar-width: thin;
}

.domain-sidebar,
.rule-sidebar {
  overflow-y: auto;
}

.domain-sidebar,
.assessment-panel,
.right-card {
  position: relative;
  background: rgba(255, 255, 255, 0.96);
  border: 1px solid #e1e7f0;
  border-radius: 8px;
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.06);
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
  color: #111827;
  font-size: 13px;
  font-weight: 700;
  backdrop-filter: blur(8px);

  b {
    font-size: 13px;
  }
}

.domain-list {
  padding: 10px 12px 0;
}

.domain-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
  height: 64px;
  margin: 0 0 8px;
  padding: 9px 10px;
  text-align: left;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  cursor: pointer;

  &.is-active {
    border-color: #2563eb;
    background: #f7faff;
    box-shadow: 0 8px 18px rgba(37, 99, 235, 0.12);
  }

  &.is-complete {
    border-color: #bde8cf;
  }
}

.domain-card__top {
  display: grid;
  grid-template-columns: 24px minmax(0, 1fr) auto;
  align-items: center;
  gap: 8px;
}

.domain-card__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  color: #667085;
  background: #eef4ff;
  border-radius: 7px;
  font-size: 15px;

  .is-active & {
    color: #fff;
    background: #2563eb;
  }
}

.domain-card__name {
  overflow: hidden;
  color: #111827;
  font-size: 13px;
  font-weight: 800;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.domain-card__count {
  color: #667085;
  font-size: 12px;
  font-weight: 700;
  white-space: nowrap;

  &::after {
    content: ' 题';
  }
}

.domain-card__bottom {
  display: flex;
  align-items: center;
  gap: 10px;
}

.domain-card__progress {
  flex: 1 1 auto;
  height: 4px;
  overflow: hidden;
  background: #e5e7eb;
  border-radius: 999px;

  i {
    display: block;
    height: 100%;
    background: #2563eb;
    border-radius: inherit;
  }
}

.domain-card.is-complete .domain-card__progress i {
  background: #18a957;
}

.domain-card__status {
  color: #98a2b3;
  font-size: 11px;
  line-height: 1;
  font-weight: 800;
  white-space: nowrap;

  &.is-active {
    color: #2563eb;
  }

  &.is-ready {
    color: #155bdc;
  }

  &.is-complete {
    color: #18a957;
  }
}

.all-items-button {
  display: flex;
  align-items: center;
  gap: 8px;
  width: calc(100% - 26px);
  min-height: 38px;
  margin: 2px 12px 0 14px;
  padding: 0 12px;
  color: #155bdc;
  background: #f7faff;
  border: 1px solid #c8dcff;
  border-radius: 8px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 900;

  span {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-align: left;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  b {
    flex: 0 0 auto;
    font-size: 18px;
    line-height: 1;
  }

  .anticon {
    flex: 0 0 auto;
    font-size: 17px;
  }
}

.progress-summary {
  position: absolute;
  right: 12px;
  bottom: 14px;
  left: 14px;
  display: grid;
  gap: 6px;
  padding: 12px;
  color: #667085;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 12px;
  font-weight: 700;

  strong {
    color: #111827;
    font-size: 14px;
    font-weight: 800;
  }
}

.assessment-panel {
  display: flex;
  flex-direction: column;
  overflow: hidden;

  :deep(.ant-spin-nested-loading),
  :deep(.ant-spin-container) {
    height: 100%;
  }

  :deep(.ant-spin-container) {
    display: flex;
    flex-direction: column;
    min-height: 0;
  }
}

.workspace-area {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  min-height: 0;
  padding: 10px 12px 6px 16px;
  background: #fff;
}

.center-head {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 12px;
  justify-content: space-between;
  margin-bottom: 8px;

  h1 {
    margin: 0;
    color: #111827;
    font-size: 18px;
    font-weight: 800;
  }
}

.center-head__actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.return-current-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 28px;
  padding: 0 11px;
  color: #155bdc;
  background: #f7faff;
  border: 1px solid #c8dcff;
  border-radius: 999px;
  cursor: pointer;
  font-size: 12px;
  font-weight: 800;
  line-height: 1;
  white-space: nowrap;

  &:hover {
    color: #0757e6;
    background: #eaf3ff;
    border-color: #9fc2ff;
  }
}

.small-badge {
  display: inline-flex;
  align-items: center;
  height: 28px;
  padding: 0 10px;
  color: #0f2a5f;
  background: #f7faff;
  border: 1px solid #c8dcff;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 900;
  white-space: nowrap;
}

.workspace-scroll {
  flex: 1 1 auto;
  min-height: 0;
  overflow-y: auto;
  scrollbar-color: #cbd5e1 transparent;
  scrollbar-width: thin;
}

.month-stack {
  display: grid;
  gap: 6px;
}

.current-empty-state {
  display: grid;
  place-items: center;
  gap: 8px;
  width: 420px;
  max-width: 100%;
  margin: 80px auto;
  padding: 20px 22px 18px;
  color: #667085;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;

  .anticon {
    color: #0757e6;
    font-size: 34px;
  }

  strong {
    color: #111827;
    font-size: 16px;
    font-weight: 900;
  }
}

.workspace-divider {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  align-items: center;
  gap: 10px;
  padding: 4px 0;

  span {
    height: 1px;
    background: #e2e8f0;
  }

  b {
    padding: 2px 10px;
    color: #475467;
    background: #f8fafc;
    border: 1px solid #e2e8f0;
    border-radius: 999px;
    font-size: 12px;
    font-weight: 700;
  }
}

.month-section {
  overflow: hidden;
  background: #fff;
  border: 1px solid #e5eaf1;
  border-radius: 8px;

  &.is-flashing {
    background: #eaf3ff;
    border-color: #0757e6;
    box-shadow: 0 0 0 2px rgba(7, 87, 230, 0.10);
  }
}

.month-section__head {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 34px;
  padding: 0 12px;
  background: #f8fafc;
  border-bottom: 1px solid #e5eaf1;

  strong {
    color: #111827;
    font-size: 13px;
    font-weight: 800;
  }

  em {
    margin-left: auto;
    color: #475467;
    font-size: 12px;
    font-style: normal;
    font-weight: 700;
  }

  &.is-main-age {
    background: #eaf3ff;

    strong {
      color: #0757e6;
    }
  }
}

.month-main-tag {
  display: inline-flex;
  align-items: center;
  height: 22px;
  padding: 0 8px;
  color: #0757e6;
  background: #fff;
  border: 1px solid #c8dcff;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 700;
}

.item-list {
  display: grid;
}

.item-row {
  display: grid;
  grid-template-columns: 48px minmax(0, 1fr) 82px 90px;
  align-items: center;
  gap: 8px;
  height: 48px;
  padding: 0 12px;
  background: #fff;
  cursor: pointer;

  &.has-divider {
    border-bottom: 1px solid #e5eaf1;
  }

  &.is-active {
    background: #eaf3ff;
  }
}

.item-no {
  color: #64748b;
  font-size: 13px;
  font-weight: 700;
}

.item-title {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;

  strong {
    overflow: hidden;
    color: #111827;
    font-size: 13px;
    font-weight: 700;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  i {
    display: inline-flex;
    flex: 0 0 auto;
    align-items: center;
    justify-content: center;
    min-width: 18px;
    height: 18px;
    padding: 0 5px;
    color: #0757e6;
    background: #eef4ff;
    border: 1px solid #c8dcff;
    border-radius: 999px;
    font-size: 11px;
    font-style: normal;
    font-weight: 900;

    &.is-warning {
      color: #155bdc;
      background: #f7faff;
      border-color: #c8dcff;
    }
  }
}

.inline-score {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  height: 32px;
  color: #475467;
  background: #fff;
  border: 1px solid #e5eaf1;
  border-radius: 8px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 700;

  &.is-pass.is-selected {
    color: #18a957;
    background: #ecfdf3;
    border-color: #a6e4be;
  }

  &.is-fail.is-selected {
    color: #d92d20;
    background: #fff8f7;
    border-color: #f5c5c1;
  }
}

.detail-panel {
  flex: 0 0 150px;
  padding: 10px 12px 8px 16px;
  background: #fff;
  border-top: 1px solid #e5eaf1;
}

.detail-panel__head {
  display: flex;
  align-items: center;
  gap: 12px;
  justify-content: space-between;
  margin-bottom: 8px;

  strong {
    overflow: hidden;
    color: #111827;
    font-size: 14px;
    font-weight: 800;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.full-detail-button {
  flex: 0 0 auto;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 700;
}

.detail-boxes {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
  height: 96px;
}

.detail-text-box {
  min-width: 0;
  padding: 8px 10px;
  overflow: hidden;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;

  h2 {
    margin: 0 0 4px;
    color: #111827;
    font-size: 13px;
    font-weight: 800;
  }

  p {
    display: -webkit-box;
    margin: 0;
    overflow: hidden;
    color: #475467;
    font-size: 13px;
    font-weight: 500;
    line-height: 1.28;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 4;
  }
}

.rule-sidebar {
  display: flex;
  flex-direction: column;
  padding-right: 2px;
}

.right-card {
  padding: 11px 12px 10px;

  h2 {
    margin: 0 0 8px;
    color: #111827;
    font-size: 13px;
    font-weight: 800;
  }
}

.rule-card {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  min-height: 0;
}

.rule-body {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  min-height: 0;

  h3 {
    margin: 10px 0 6px;
    color: #111827;
    font-size: 13px;
    font-weight: 800;
  }
}

.next-action {
  display: grid;
  grid-template-columns: 24px minmax(0, 1fr);
  gap: 8px;
  padding: 9px 10px;
  color: #155bdc;
  background: #f7faff;
  border: 1px solid #c8dcff;
  border-radius: 8px;

  .anticon {
    margin-top: 2px;
    font-size: 16px;
  }

  span {
    display: block;
    color: #667085;
    font-size: 12px;
  }

  strong {
    display: block;
    margin-top: 2px;
    color: #0f2a5f;
    font-size: 13px;
    font-weight: 800;
    line-height: 1.35;
  }
}

.rule-list {
  display: flex;
  flex-direction: column;
  flex: 1 1 auto;
  min-height: 0;
  overflow: hidden;
  background: #fff;
  border: 1px solid #e5eaf1;
  border-radius: 8px;
}

.rule-month-list {
  flex: 1 1 auto;
  min-height: 0;
  overflow-y: auto;
  scrollbar-color: #cbd5e1 transparent;
  scrollbar-width: thin;
}

.rule-pinned-list {
  flex: 0 0 auto;
  background: #f8fafc;
  border-top: 1px solid #e5eaf1;
}

.rule-row {
  display: grid;
  grid-template-columns: 16px minmax(0, 1fr) auto 14px;
  align-items: center;
  gap: 8px;
  width: 100%;
  height: 34px;
  padding: 0 10px;
  color: #475467;
  text-align: left;
  background: #fff;
  border: 0;
  border-bottom: 1px solid #e5eaf1;
  cursor: default;

  span {
    overflow: hidden;
    color: #111827;
    font-size: 12px;
    font-weight: 700;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  b {
    color: #475467;
    font-size: 12px;
    font-weight: 700;
    white-space: nowrap;
  }

  .anticon {
    color: #98a2b3;
    font-size: 14px;
  }

  &.is-done {
    .anticon:first-child,
    b {
      color: #18a957;
    }
  }

  &.is-unmet {
    .anticon:first-child,
    b {
      color: #d92d20;
    }
  }

  &.is-selected {
    background: #eaf3ff;

    span,
    .anticon:last-child {
      color: #0757e6;
    }
  }

  &.is-flashing {
    background: #eaf3ff;
    box-shadow: inset 2px 0 0 #0757e6;

    span,
    .anticon:last-child {
      color: #0757e6;
    }
  }

  &.is-pinned {
    background: #f8fafc;

    &:last-child {
      border-bottom: 0;
    }
  }

  &.is-clickable {
    cursor: pointer;

    &:hover {
      background: #f7faff;
    }
  }

  &:disabled {
    cursor: default;
  }
}

.rule-hint {
  min-height: 34px;
  margin: 0 0 6px;
  overflow: hidden;
  color: #667085;
  font-size: 12px;
  line-height: 1.28;
  text-overflow: ellipsis;
}

.rule-action {
  height: 34px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 700;
}

.right-remark-section {
  flex: 0 0 auto;
  padding-top: 10px;

  h3 {
    margin: 0 0 6px;
    color: #111827;
    font-size: 13px;
    font-weight: 800;
  }

  :deep(.ant-input) {
    background: #f8fafc;
    border-color: #e2e8f0;
    border-radius: 8px;
    font-size: 13px;
  }
}

.detail-modal-body {
  display: grid;
  gap: 14px;

  h3 {
    margin: 0;
    color: #111827;
    font-size: 16px;
    font-weight: 900;
  }

  section {
    padding: 12px;
    background: #f8fafc;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
  }

  strong {
    display: block;
    margin-bottom: 6px;
    color: #111827;
    font-size: 14px;
    font-weight: 900;
  }

  p {
    margin: 0;
    color: #475467;
    font-size: 13px;
    line-height: 1.6;
  }
}

.overview-title {
  display: flex;
  align-items: center;
  gap: 10px;

  strong {
    color: #111827;
    font-size: 18px;
    font-weight: 900;
  }

  span {
    display: inline-flex;
    align-items: center;
    height: 26px;
    padding: 0 10px;
    color: #475467;
    background: #f7faff;
    border: 1px solid #d8e6ff;
    border-radius: 999px;
    font-size: 12px;
    font-weight: 800;
  }
}

.overview-scroll {
  display: grid;
  gap: 14px;
  max-height: calc(100vh - 150px);
  padding: 22px 18px 16px;
  overflow-y: auto;
  background: #f8fafc;
  scrollbar-color: #cbd5e1 transparent;
  scrollbar-width: thin;
}

.overview-empty {
  display: grid;
  place-items: center;
  min-height: 320px;
  color: #667085;
  font-size: 14px;
  font-weight: 800;
}

.overview-section {
  overflow: hidden;
  background: #fff;
  border: 1px solid #d8dfe8;
  border-radius: 8px;

  &.is-target {
    border-color: #0757e6;
    box-shadow: 0 0 0 2px rgba(7, 87, 230, 0.08);
  }
}

.overview-table {
  width: 100%;
  table-layout: fixed;
  border-collapse: collapse;
  color: #111827;

  th,
  td {
    border: 1px solid #d8dfe8;
  }

  thead th {
    height: 40px;
    color: #0f172a;
    background: #f8fafc;
    font-size: 14px;
    font-weight: 900;

    &.is-main-age {
      color: #0757e6;
      background: #eaf3ff;
    }
  }

  tbody th {
    width: 82px;
    padding: 8px 6px;
    color: #111827;
    background: #fbfdff;
    font-size: 14px;
    font-weight: 900;
    text-align: center;
  }

  td {
    min-height: 48px;
    padding: 7px 8px;
    vertical-align: top;
    background: #fff;
  }
}

.overview-domain-col {
  width: 82px;
}

.overview-item {
  display: grid;
  grid-template-columns: 14px minmax(0, 1fr);
  gap: 4px;
  align-items: start;
  padding: 2px 3px;
  border-radius: 4px;

  + .overview-item {
    margin-top: 2px;
  }

  &.is-selected {
    background: #eaf3ff;
  }

  p {
    min-width: 0;
    margin: 0;
    overflow: hidden;
    color: #1f2937;
    font-size: 13px;
    font-weight: 600;
    line-height: 1.25;
    text-overflow: ellipsis;
  }

  b {
    font-weight: 900;
  }

  i {
    margin-left: 3px;
    color: #0757e6;
    font-size: 11px;
    font-style: normal;
    font-weight: 900;
  }
}

.overview-check {
  width: 13px;
  height: 13px;
  margin-top: 2px;
  border: 1px solid #b8c0cc;
  border-radius: 3px;
}

.overview-item.is-pass .overview-check {
  position: relative;
  background: #18a957;
  border-color: #18a957;

  &::after {
    position: absolute;
    top: 1px;
    left: 4px;
    width: 4px;
    height: 8px;
    border: solid #fff;
    border-width: 0 2px 2px 0;
    content: '';
    transform: rotate(45deg);
  }
}

.overview-item.is-fail .overview-check {
  position: relative;
  background: #d92d20;
  border-color: #d92d20;

  &::after {
    position: absolute;
    top: 5px;
    left: 3px;
    width: 7px;
    height: 2px;
    background: #fff;
    border-radius: 999px;
    content: '';
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

@media (max-width: 1440px) {
  .workbench-main {
    grid-template-columns: 208px minmax(480px, 1fr) 276px;
  }

  .header-meta {
    font-size: 14px;
  }
}
</style>
