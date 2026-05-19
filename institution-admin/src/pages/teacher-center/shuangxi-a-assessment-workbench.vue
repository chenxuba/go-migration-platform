<script setup lang="ts">
import {
  ArrowLeftOutlined,
  BulbOutlined,
  CheckCircleFilled,
  ClockCircleOutlined,
  EyeOutlined,
  FileDoneOutlined,
  HomeOutlined,
  LeftOutlined,
  MessageOutlined,
  RightOutlined,
  SaveOutlined,
  SlidersOutlined,
  SwapOutlined,
  TeamOutlined,
  ThunderboltOutlined,
  ToolOutlined,
} from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import { Modal } from 'ant-design-vue'
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  getShuangxiAAssessmentDraftDetailApi,
  getShuangxiAAssessmentFormTemplateItemApi,
  getShuangxiAAssessmentFormTemplateSummaryApi,
  getShuangxiAAssessmentRecordDetailApi,
  pageShuangxiAAssessmentDraftsApi,
  pageShuangxiAAssessmentRecordsApi,
  saveShuangxiAAssessmentDraftApi,
  saveShuangxiAAssessmentDraftItemApi,
  submitShuangxiAAssessmentDraftApi,
  updateShuangxiAAssessmentRecordApi,
  type ShuangxiAAssessmentItem,
  type ShuangxiADomainSummary,
  type ShuangxiADraftDetail,
  type ShuangxiADraftInput,
  type ShuangxiADraftSaveRequest,
  type ShuangxiADraftSummary,
  type ShuangxiAItemSummary,
  type ShuangxiAScoreOption,
  type ShuangxiASkillSummary,
  type ShuangxiATemplateSummary,
} from '@/api/edu-center/shuangxi-assessment'
import {
  getScaleAssessmentStudentCandidatesApi,
  updateScaleAssessmentStudentGenderApi,
} from '@/api/teacher-center/scale-library'
import messageService from '@/utils/messageService'

type DraftItemSaveStatus = 'saving' | 'saved' | 'error'
type NormalizedGender = 'male' | 'female' | ''
const AUTO_NEXT_DELAY_MS = 350

const route = useRoute()
const router = useRouter()

const templateLoading = ref(false)
const itemLoading = ref(false)
const saving = ref(false)
const submitting = ref(false)
const draftResumeModalOpen = ref(false)
const genderModalOpen = ref(false)
const genderSaving = ref(false)
const selectedGenderDraft = ref<NormalizedGender>('')
const selectedDomainCode = ref('')
const selectedItemNo = ref(numberFromQuery('itemNo') || 0)
const autoNext = ref(true)
const autoSaveLastSavedAt = ref('')
const currentProgress = ref<ShuangxiADraftDetail['progress']>()
const existingDraft = ref<ShuangxiADraftSummary>()
const template = ref<ShuangxiATemplateSummary>()
const itemCache = reactive<Record<number, ShuangxiAAssessmentItem>>({})
const itemScores = reactive<Record<number, number>>({})
const itemRemarks = reactive<Record<number, string>>({})
const previousItemScores = reactive<Record<number, number>>({})
const previousAssessmentDate = ref('')
const draftItemSaveStatus = ref<Record<number, DraftItemSaveStatus>>({})
const draftItemSaveErrors = ref<Record<number, string>>({})
const skillListRef = ref<HTMLElement | null>(null)
let draftSavePromise: Promise<ShuangxiADraftDetail | undefined> | undefined
let draftCreationPromise: Promise<ShuangxiADraftDetail | undefined> | undefined
let itemSaveChain: Promise<void> = Promise.resolve()
let autoNextTimer: ReturnType<typeof setTimeout> | undefined

const editingRecordId = ref(numberFromQuery('recordId') || 0)
const recordMode = computed(() => textFromQuery('recordMode').toLowerCase())
const isRecordReuseMode = computed(() => editingRecordId.value > 0 && recordMode.value === 'reuse')
const isSubmittedRecordMode = computed(() => editingRecordId.value > 0)
const isRecordEditMode = computed(() => editingRecordId.value > 0 && !isRecordReuseMode.value)

const editor = reactive({
  id: numberFromQuery('draftId') || undefined as number | undefined,
  studentId: numberFromQuery('childId') || undefined as number | undefined,
  studentName: textFromQuery('childName'),
  studentGender: textFromQuery('childGender'),
  examinerName: textFromQuery('examinerName'),
  birthDate: normalizeDateText(textFromQuery('childBirthDate') || textFromQuery('birthDate')),
  assessmentDate: normalizeDateText(textFromQuery('assessmentDate')) || dayjs().format('YYYY-MM-DD'),
  remark: '',
})

const scaleTitle = computed(() => textFromQuery('scaleName') || template.value?.title || '双溪课程评量表A')
const studentName = computed(() => editor.studentName || textFromQuery('childName') || '-')
const examinerName = computed(() => editor.examinerName || '当前老师')
const assessmentDateText = computed(() => formatDate(editor.assessmentDate))
const studentAge = computed(() => assessmentAgeText(editor.birthDate, editor.assessmentDate) || textFromQuery('childAge') || '-')
const allDomains = computed(() => template.value?.domains || [])
const allSkills = computed(() => allDomains.value.flatMap(domain => domain.skills || []))
const allItems = computed(() => allSkills.value.flatMap(skill => skill.items || []))
const selectedDomain = computed(() => allDomains.value.find(domain => domain.domainCode === selectedDomainCode.value) || allDomains.value[0])
const selectedDomainSkills = computed(() => selectedDomain.value?.skills || [])
const currentItemSummary = computed(() => summaryByNo(selectedItemNo.value))
const currentItem = computed(() => itemCache[selectedItemNo.value])
const currentScoreOptions = computed(() => scoreOptionsForItem(currentItem.value || currentItemSummary.value))
const currentScore = computed(() => effectiveItemScores.value[selectedItemNo.value])
const previousScore = computed(() => previousItemScores[selectedItemNo.value])
const currentRemark = computed({
  get: () => itemRemarks[selectedItemNo.value] || '',
  set: value => updateRemark(value),
})
const currentItemDisplayTitle = computed(() => displayNumberedItemTitle(currentItem.value || currentItemSummary.value))
const currentSkillName = computed(() => currentItemSummary.value?.skillName || currentItem.value?.skillName || '-')
const currentDomainName = computed(() => currentItemSummary.value?.domainName || currentItem.value?.domainName || '-')
const currentSkillDisplayTitle = computed(() => displaySkillTitle(currentItem.value || currentItemSummary.value))
const effectiveItemScores = computed(() => {
  const scores: Record<number, number> = { ...itemScores }
  const sanitaryPadScore = genderDefaultScoreForItem(82)
  if (sanitaryPadScore !== undefined)
    scores[82] = sanitaryPadScore
  const shaveScore = genderDefaultScoreForItem(83)
  if (shaveScore !== undefined)
    scores[83] = shaveScore
  return scores
})
const answeredItemCount = computed(() => allItems.value.filter(item => hasScore(item.itemNo)).length)
const totalItemCount = computed(() => template.value?.itemCount || allItems.value.length)
const missingItemCount = computed(() => Math.max(totalItemCount.value - answeredItemCount.value, 0))
const progressPercent = computed(() => totalItemCount.value ? Math.round((answeredItemCount.value / totalItemCount.value) * 100) : 0)
const currentIndex = computed(() => allItems.value.findIndex(item => item.itemNo === selectedItemNo.value))
const currentDisplayIndex = computed(() => currentIndex.value >= 0 ? currentIndex.value + 1 : 0)
const hasPreviousItem = computed(() => currentIndex.value > 0)
const hasNextItem = computed(() => currentIndex.value >= 0 && currentIndex.value < allItems.value.length - 1)
const firstMissingNo = computed(() => firstMissingItemNo())
const selectedSkillCode = computed(() => currentItemSummary.value?.skillCode || currentItem.value?.skillCode || '')
const previousScoreOption = computed(() => currentScoreOptions.value.find(option => option.value === previousScore.value))
const normalizedGender = computed(() => normalizeGender(editor.studentGender))
const genderText = computed(() => normalizedGender.value === 'male' ? '男' : normalizedGender.value === 'female' ? '女' : '未确认')
const autoSaveState = computed<'idle' | 'saving' | 'saved'>(() => {
  if (saving.value || Object.values(draftItemSaveStatus.value).some(status => status === 'saving'))
    return 'saving'
  return autoSaveLastSavedAt.value ? 'saved' : 'idle'
})
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
const donutStyle = computed(() => ({
  background: `radial-gradient(circle at center, #fff 54%, transparent 55%), conic-gradient(#2563eb 0 ${progressPercent.value}%, #e5e7eb ${progressPercent.value}% 100%)`,
}))
const domainCards = computed(() => allDomains.value.map((domain) => {
  const items = (domain.skills || []).flatMap(skill => skill.items || [])
  const answered = items.filter(item => hasScore(item.itemNo)).length
  const percent = items.length ? Math.round((answered / items.length) * 100) : 0
  return {
    code: domain.domainCode,
    name: domain.domainName,
    icon: domainIconForName(domain.domainName),
    answered,
    total: items.length,
    percent,
    active: domain.domainCode === selectedDomainCode.value,
  }
}))

function domainIconForName(name = '') {
  if (name.includes('感官') || name.includes('知觉'))
    return EyeOutlined
  if (name.includes('粗大'))
    return ThunderboltOutlined
  if (name.includes('精细'))
    return ToolOutlined
  if (name.includes('生活') || name.includes('自理'))
    return HomeOutlined
  if (name.includes('沟通'))
    return MessageOutlined
  if (name.includes('认知'))
    return BulbOutlined
  if (name.includes('社会') || name.includes('社交'))
    return TeamOutlined
  return SlidersOutlined
}

watch(selectedItemNo, (itemNo) => {
  if (itemNo > 0)
    void fetchItemDetail(itemNo)
  syncSelectedDomain()
  void revealSelectedItem()
})

watch(allItems, () => {
  ensureSelectedItem()
}, { immediate: true })

onMounted(() => {
  void initializeWorkbench()
})

onBeforeUnmount(() => {
  clearPendingAutoNext()
})

async function initializeWorkbench() {
  await fetchTemplate()
  if (isSubmittedRecordMode.value) {
    await fetchRecordForEdit(editingRecordId.value)
    return
  }
  if (editor.id) {
    await fetchDraftDetail(editor.id)
    await hydrateStudentInfo()
    await loadPreviousAssessment()
    selectInitialItem(currentProgress.value?.missingItemNos?.[0])
    ensureStudentGender()
    return
  }
  await hydrateStudentInfo()
  const draft = await findExistingDraft()
  if (draft) {
    existingDraft.value = draft
    draftResumeModalOpen.value = true
    selectInitialItem(draft.progress?.missingItemNos?.[0])
    return
  }
  await startNewAssessment()
}

async function fetchTemplate() {
  templateLoading.value = true
  try {
    const res = await getShuangxiAAssessmentFormTemplateSummaryApi()
    template.value = unwrap<ShuangxiATemplateSummary>(res)
    selectedDomainCode.value = template.value.domains?.[0]?.domainCode || ''
    selectInitialItem()
  }
  catch (error: any) {
    messageService.error(getErrorMessage(error, '获取双溪课程评量表A题目目录失败'))
  }
  finally {
    templateLoading.value = false
  }
}

async function fetchRecordForEdit(recordId: number) {
  if (!recordId)
    return
  try {
    const res = await getShuangxiAAssessmentRecordDetailApi(recordId)
    const detail = unwrap<any>(res)
    const input = normalizeDraftInputSnapshot(detail.input)
    const reuseAssessmentDate = dayjs().format('YYYY-MM-DD')
    editor.id = undefined
    editor.studentId = detail.studentId || input?.studentId || editor.studentId
    editor.studentName = detail.studentName || input?.studentName || editor.studentName
    editor.studentGender = detail.studentGender || input?.studentGender || editor.studentGender
    editor.examinerName = detail.examinerName || input?.examinerName || editor.examinerName
    editor.birthDate = normalizeDateText(detail.birthDate || input?.birthDate) || editor.birthDate
    editor.assessmentDate = isRecordReuseMode.value ? reuseAssessmentDate : normalizeDateText(detail.assessmentDate || input?.assessmentDate) || editor.assessmentDate
    editor.remark = detail.remark || input?.remark || ''
    applyDraftInput(input, { keepAssessmentDate: isRecordReuseMode.value })
    if (isRecordReuseMode.value)
      editor.assessmentDate = reuseAssessmentDate
    selectInitialItem()
    await loadPreviousAssessment()
    ensureStudentGender()
    messageService.info(isRecordReuseMode.value ? '当前正在复用已提交的双溪测评，提交后会生成新的正式记录' : '当前正在修改已提交的双溪测评记录，修改后请重新提交')
  }
  catch (error: any) {
    messageService.error(getErrorMessage(error, '获取双溪测评记录失败'))
    void router.push('/teacherCenter/evaluationRecord')
  }
}

async function fetchItemDetail(itemNo: number) {
  if (itemNo <= 0 || itemCache[itemNo])
    return
  itemLoading.value = true
  try {
    const res = await getShuangxiAAssessmentFormTemplateItemApi(itemNo)
    const detail = unwrap<ShuangxiAAssessmentItem>(res)
    if (detail?.itemNo === itemNo)
      itemCache[itemNo] = detail
  }
  catch (error: any) {
    messageService.error(getErrorMessage(error, `获取第 ${itemNo} 题失败`))
  }
  finally {
    itemLoading.value = false
  }
}

async function hydrateStudentInfo() {
  if (!editor.studentId)
    return
  if (editor.birthDate && normalizeGender(editor.studentGender))
    return
  try {
    const res = await getScaleAssessmentStudentCandidatesApi({
      scaleCode: 'SHUANGXI_A',
      keyword: editor.studentName || undefined,
      pageIndex: 1,
      pageSize: 100,
    })
    const data = unwrap<any>(res)
    const student = (data?.items || []).find((item: any) => Number(item.id) === Number(editor.studentId))
    if (student?.birthDate && !editor.birthDate)
      editor.birthDate = normalizeDateText(student.birthDate)
    if (student?.gender && !normalizeGender(editor.studentGender))
      editor.studentGender = student.gender
  }
  catch {
  }
}

async function findExistingDraft() {
  if (!editor.studentId)
    return undefined
  try {
    const res = await pageShuangxiAAssessmentDraftsApi({
      pageRequestModel: { pageIndex: 1, pageSize: 1 },
      queryModel: {
        assessmentCode: 'SHUANGXI_A',
        studentId: editor.studentId,
      },
      latestOnly: true,
    })
    const data = unwrap<any>(res)
    return (data?.items || [])[0] as ShuangxiADraftSummary | undefined
  }
  catch (error: any) {
    messageService.error(getErrorMessage(error, '查询双溪未完成草稿失败'))
    return undefined
  }
}

async function startNewAssessment() {
  resetAssessmentInput()
  selectInitialItem()
  ensureStudentGender()
  await loadPreviousAssessment()
  if (validateDraftHeader(true) && normalizeGender(editor.studentGender))
    await saveDraft(true)
}

async function continueExistingDraft() {
  const draft = existingDraft.value
  draftResumeModalOpen.value = false
  if (!draft?.id)
    return
  await fetchDraftDetail(Number(draft.id))
  await loadPreviousAssessment()
  selectInitialItem(draft.progress?.missingItemNos?.[0])
  ensureStudentGender()
}

async function restartAssessment() {
  draftResumeModalOpen.value = false
  existingDraft.value = undefined
  await startNewAssessment()
}

async function fetchDraftDetail(id: number) {
  try {
    const res = await getShuangxiAAssessmentDraftDetailApi(id)
    applyDraftDetail(unwrap<ShuangxiADraftDetail>(res))
  }
  catch (error: any) {
    messageService.error(getErrorMessage(error, '获取双溪测评草稿失败'))
  }
}

function applyDraftDetail(detail: ShuangxiADraftDetail) {
  editor.id = detail.id
  editor.studentId = detail.studentId || detail.input?.studentId || editor.studentId
  editor.studentName = detail.studentName || detail.input?.studentName || editor.studentName
  editor.studentGender = detail.input?.studentGender || editor.studentGender
  editor.examinerName = detail.examinerName || detail.input?.examinerName || editor.examinerName
  editor.birthDate = normalizeDateText(detail.birthDate || detail.input?.birthDate) || editor.birthDate
  editor.assessmentDate = normalizeDateText(detail.assessmentDate || detail.input?.assessmentDate) || editor.assessmentDate
  editor.remark = detail.input?.remark || detail.remark || ''
  currentProgress.value = detail.progress
  applyDraftInput(detail.input)
  autoSaveLastSavedAt.value = detail.updatedTime ? dayjs(detail.updatedTime).format('MM-DD HH:mm') : autoSaveLastSavedAt.value
}

function mergeDraftDetailInput(detail: ShuangxiADraftDetail) {
  const localScores = { ...itemScores }
  const localRemarks = { ...itemRemarks }
  applyDraftDetail(detail)
  Object.entries(localScores).forEach(([itemNo, score]) => {
    if (isValidScore(score))
      itemScores[Number(itemNo)] = Number(score)
  })
  Object.entries(localRemarks).forEach(([itemNo, remark]) => {
    if (remark?.trim())
      itemRemarks[Number(itemNo)] = remark
  })
}

function applyDraftInput(input?: ShuangxiADraftInput, options: { keepAssessmentDate?: boolean } = {}) {
  clearRecord(itemScores)
  clearRecord(itemRemarks)
  if (!input)
    return
  editor.studentId = input.studentId || editor.studentId
  editor.studentName = input.studentName || editor.studentName
  editor.studentGender = input.studentGender || editor.studentGender
  editor.examinerName = input.examinerName || editor.examinerName
  editor.birthDate = normalizeDateText(input.birthDate) || editor.birthDate
  if (!options.keepAssessmentDate)
    editor.assessmentDate = normalizeDateText(input.assessmentDate) || editor.assessmentDate
  editor.remark = input.remark || editor.remark
  Object.entries(input.itemScores || {}).forEach(([itemNo, score]) => {
    if (isValidScore(score))
      itemScores[Number(itemNo)] = Number(score)
  })
  ;(input.itemScoreList || []).forEach((item) => {
    if (item.itemNo > 0 && isValidScore(item.score))
      itemScores[item.itemNo] = Number(item.score)
    if (item.itemNo > 0 && item.remark?.trim())
      itemRemarks[item.itemNo] = item.remark
  })
  Object.entries(input.itemRemarks || {}).forEach(([itemNo, remark]) => {
    if (remark?.trim())
      itemRemarks[Number(itemNo)] = remark
  })
  ;(input.itemRemarkList || []).forEach((item) => {
    if (item.itemNo > 0 && item.remark?.trim())
      itemRemarks[item.itemNo] = item.remark
  })
}

function normalizeDraftInputSnapshot(input: unknown): ShuangxiADraftInput | undefined {
  if (!input)
    return undefined
  if (typeof input === 'string') {
    try {
      return JSON.parse(input) as ShuangxiADraftInput
    }
    catch {
      return undefined
    }
  }
  return input as ShuangxiADraftInput
}

function resetAssessmentInput() {
  editor.id = undefined
  editor.remark = ''
  clearRecord(itemScores)
  clearRecord(itemRemarks)
  currentProgress.value = undefined
  autoSaveLastSavedAt.value = ''
  draftItemSaveStatus.value = {}
  draftItemSaveErrors.value = {}
}

function selectInitialItem(preferredItemNo = 0) {
  if (preferredItemNo && allItems.value.some(item => item.itemNo === preferredItemNo)) {
    selectItem(preferredItemNo)
    return
  }
  ensureSelectedItem()
}

function ensureSelectedItem() {
  const items = allItems.value
  if (!items.length) {
    selectedItemNo.value = 0
    selectedDomainCode.value = ''
    return
  }
  if (items.some(item => item.itemNo === selectedItemNo.value)) {
    syncSelectedDomain()
    return
  }
  const target = items.find(item => !hasScore(item.itemNo)) || items[0]
  selectedItemNo.value = target.itemNo
  selectedDomainCode.value = target.domainCode
}

function validateDraftHeader(silent = false) {
  if (!editor.studentId || studentName.value === '-') {
    if (!silent)
      messageService.warning('缺少真实儿童，无法保存测评')
    return false
  }
  if (!normalizeGender(editor.studentGender)) {
    if (!silent) {
      messageService.warning('请先确认学生性别')
      ensureStudentGender()
    }
    return false
  }
  return true
}

function buildPayload(): ShuangxiADraftSaveRequest {
  const scores = effectiveItemScores.value
  const itemScoreList = Object.entries(scores)
    .filter(([, score]) => isValidScore(score))
    .map(([itemNo, score]) => ({
      itemNo: Number(itemNo),
      score: Number(score),
      remark: itemRemarks[Number(itemNo)]?.trim() || '',
    }))
    .sort((left, right) => left.itemNo - right.itemNo)
  const itemRemarkList = Object.entries(itemRemarks)
    .filter(([, remark]) => remark.trim())
    .map(([itemNo, remark]) => ({ itemNo: Number(itemNo), remark: remark.trim() }))
    .sort((left, right) => left.itemNo - right.itemNo)
  return {
    id: editor.id,
    studentId: editor.studentId,
    studentName: studentName.value === '-' ? undefined : studentName.value,
    studentGender: genderText.value,
    examinerName: editor.examinerName,
    birthDate: editor.birthDate,
    assessmentDate: editor.assessmentDate,
    remark: editor.remark,
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
  draftSavePromise = persistDraft().finally(() => {
    draftSavePromise = undefined
    saving.value = false
  })
  const detail = await draftSavePromise
  if (!silent && detail)
    messageService.success('双溪测评草稿已保存')
  return detail
}

async function persistDraft() {
  try {
    const res = await saveShuangxiAAssessmentDraftApi(buildPayload())
    const detail = unwrap<ShuangxiADraftDetail>(res)
    applyDraftDetail(detail)
    return detail
  }
  catch (error: any) {
    messageService.error(getErrorMessage(error, '保存双溪测评草稿失败'))
    return undefined
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

function selectScore(score: number) {
  if (selectedItemNo.value <= 0 || submitting.value)
    return
  const fixedScore = genderDefaultScoreForItem(selectedItemNo.value)
  if (fixedScore !== undefined) {
    itemScores[selectedItemNo.value] = fixedScore
    messageService.info(genderDefaultReasonForItem(selectedItemNo.value))
    queueSaveItem(selectedItemNo.value, false)
    return
  }
  const itemNo = selectedItemNo.value
  itemScores[itemNo] = score
  queueSaveItem(itemNo, autoNext.value)
}

function queueSaveItem(itemNo: number, moveNext = false) {
  if (isRecordEditMode.value) {
    if (moveNext)
      scheduleAutoNext(itemNo)
    return
  }
  itemSaveChain = itemSaveChain.then(() => persistItem(itemNo, moveNext)).catch(() => undefined)
}

async function persistItem(itemNo: number, moveNext = false) {
  if (itemNo <= 0 || !isValidScore(effectiveItemScores.value[itemNo]))
    return
  draftItemSaveStatus.value = { ...draftItemSaveStatus.value, [itemNo]: 'saving' }
  try {
    const canSave = await ensureDraftForItemSave()
    if (!canSave || !editor.id)
      throw new Error('草稿创建失败')
    const res = await saveShuangxiAAssessmentDraftItemApi({
      draftId: editor.id,
      itemNo,
      score: Number(effectiveItemScores.value[itemNo]),
      remark: itemRemarks[itemNo]?.trim() || '',
      studentGender: genderText.value,
    })
    mergeDraftDetailInput(unwrap<ShuangxiADraftDetail>(res))
    autoSaveLastSavedAt.value = dayjs().format('MM-DD HH:mm')
    draftItemSaveStatus.value = { ...draftItemSaveStatus.value, [itemNo]: 'saved' }
    if (moveNext && selectedItemNo.value === itemNo)
      scheduleAutoNext(itemNo)
  }
  catch (error: any) {
    const message = getErrorMessage(error, `第 ${itemNo} 题自动保存失败`)
    draftItemSaveStatus.value = { ...draftItemSaveStatus.value, [itemNo]: 'error' }
    draftItemSaveErrors.value = { ...draftItemSaveErrors.value, [itemNo]: message }
    messageService.error(message)
  }
}

function clearPendingAutoNext() {
  if (!autoNextTimer)
    return
  clearTimeout(autoNextTimer)
  autoNextTimer = undefined
}

function scheduleAutoNext(itemNo: number) {
  if (itemNo <= 0 || !hasNextItem.value)
    return
  clearPendingAutoNext()
  autoNextTimer = setTimeout(() => {
    autoNextTimer = undefined
    if (selectedItemNo.value === itemNo)
      goNextItem()
  }, AUTO_NEXT_DELAY_MS)
}

function updateRemark(value: string) {
  if (selectedItemNo.value <= 0)
    return
  if (value.trim())
    itemRemarks[selectedItemNo.value] = value
  else
    delete itemRemarks[selectedItemNo.value]
}

async function persistRemark() {
  if (selectedItemNo.value <= 0)
    return
  if (isRecordEditMode.value)
    return
  if (hasScore(selectedItemNo.value))
    queueSaveItem(selectedItemNo.value)
  else
    await saveDraft(true)
}

async function submitDraft() {
  if (!editor.birthDate || !editor.assessmentDate) {
    messageService.warning('缺少出生日期或评量日期，不能提交正式记录')
    return
  }
  if (!validateDraftHeader(false))
    return
  if (missingItemCount.value > 0) {
    const confirmed = await confirmFillMissingScores()
    if (!confirmed)
      return
    fillMissingScoresWithZero()
  }
  if (isRecordEditMode.value) {
    await submitRecordEdit()
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
    await submitShuangxiAAssessmentDraftApi(draftId)
    messageService.success(isRecordReuseMode.value ? '已基于复用测评生成新的正式记录' : '已提交正式测评记录')
    await router.push('/teacherCenter/evaluationRecord')
  }
  catch (error: any) {
    messageService.error(getErrorMessage(error, '提交双溪测评记录失败'))
  }
  finally {
    submitting.value = false
  }
}

async function submitRecordEdit() {
  submitting.value = true
  try {
    const res = await updateShuangxiAAssessmentRecordApi({
      ...buildPayload(),
      id: editingRecordId.value,
    })
    const result = unwrap<any>(res)
    messageService.success('已重新提交，并生成新的双溪评量报告')
    if (result?.id)
      await router.push('/teacherCenter/evaluationRecord')
  }
  catch (error: any) {
    messageService.error(getErrorMessage(error, '重新提交双溪测评记录失败'))
  }
  finally {
    submitting.value = false
  }
}

function confirmFillMissingScores() {
  return new Promise<boolean>((resolve) => {
    Modal.confirm({
      title: '仍有题目未评分',
      content: `当前还有 ${missingItemCount.value} 道题未评分。双溪正式提交会将未评分题目按 0 分处理，确认继续吗？`,
      okText: '按0分提交',
      cancelText: '返回补题',
      centered: true,
      onOk: () => resolve(true),
      onCancel: () => resolve(false),
    })
  })
}

function fillMissingScoresWithZero() {
  allItems.value.forEach((item) => {
    if (!hasScore(item.itemNo))
      itemScores[item.itemNo] = 0
  })
}

async function loadPreviousAssessment() {
  if (!editor.studentId)
    return
  clearRecord(previousItemScores)
  previousAssessmentDate.value = ''
  try {
    const res = await pageShuangxiAAssessmentRecordsApi({
      pageRequestModel: { pageIndex: 1, pageSize: isSubmittedRecordMode.value ? 20 : 5 },
      queryModel: {
        assessmentCode: 'SHUANGXI_A',
        studentId: editor.studentId,
        assessmentDateEnd: normalizeDateText(editor.assessmentDate),
      },
    })
    const data = unwrap<any>(res)
    const latest = (data?.items || []).find((item: any) => Number(item.id) !== editingRecordId.value)
    if (!latest?.id)
      return
    const detailRes = await getShuangxiAAssessmentRecordDetailApi(Number(latest.id))
    const detail = unwrap<any>(detailRes)
    previousAssessmentDate.value = formatDate(detail.assessmentDate || latest.assessmentDate)
    const input = normalizeDraftInputSnapshot(detail.input)
    Object.entries(input?.itemScores || {}).forEach(([itemNo, score]) => {
      if (isValidScore(score))
        previousItemScores[Number(itemNo)] = Number(score)
    })
    ;(input?.itemScoreList || []).forEach((item) => {
      if (item.itemNo > 0 && isValidScore(item.score))
        previousItemScores[item.itemNo] = Number(item.score)
    })
  }
  catch {
    clearRecord(previousItemScores)
    previousAssessmentDate.value = ''
  }
}

function selectDomain(code: string) {
  if (!code)
    return
  clearPendingAutoNext()
  selectedDomainCode.value = code
  const domain = allDomains.value.find(item => item.domainCode === code)
  const items = (domain?.skills || []).flatMap(skill => skill.items || [])
  const target = items.find(item => !hasScore(item.itemNo)) || items[0]
  if (target)
    selectedItemNo.value = target.itemNo
  void scrollSkillListToTop()
}

function selectItem(itemNo: number) {
  if (itemNo <= 0)
    return
  clearPendingAutoNext()
  selectedItemNo.value = itemNo
}

function goPreviousItem() {
  if (!hasPreviousItem.value)
    return
  clearPendingAutoNext()
  selectedItemNo.value = allItems.value[currentIndex.value - 1].itemNo
}

function goNextItem() {
  if (!hasNextItem.value)
    return
  clearPendingAutoNext()
  selectedItemNo.value = allItems.value[currentIndex.value + 1].itemNo
}

function jumpToMissingItem() {
  if (firstMissingNo.value > 0)
    selectItem(firstMissingNo.value)
  else
    messageService.success('当前没有缺题')
}

function firstMissingItemNo() {
  return allItems.value.find(item => !hasScore(item.itemNo))?.itemNo || 0
}

function hasScore(itemNo: number) {
  return isValidScore(effectiveItemScores.value[itemNo])
}

function skillProgress(skill: ShuangxiASkillSummary) {
  const items = skill.items || []
  const answered = items.filter(item => hasScore(item.itemNo)).length
  return {
    answered,
    total: items.length,
    percent: items.length ? Math.round((answered / items.length) * 100) : 0,
  }
}

function isSkillActive(skill: ShuangxiASkillSummary) {
  return skill.skillCode === selectedSkillCode.value
}

function selectSkill(skill: ShuangxiASkillSummary) {
  const items = skill.items || []
  const target = items.find(item => !hasScore(item.itemNo)) || items[0]
  if (target?.itemNo)
    selectItem(target.itemNo)
}

function itemStatus(item: ShuangxiAItemSummary) {
  if (item.itemNo === selectedItemNo.value)
    return 'active'
  return hasScore(item.itemNo) ? 'done' : 'todo'
}

function summaryByNo(itemNo: number) {
  return allItems.value.find(item => item.itemNo === itemNo)
}

function scoreOptionsForItem(item?: Partial<ShuangxiAAssessmentItem | ShuangxiAItemSummary>) {
  const options = 'scoreOptions' in (item || {}) && Array.isArray((item as ShuangxiAAssessmentItem)?.scoreOptions)
    ? (item as ShuangxiAAssessmentItem).scoreOptions
    : template.value?.scoreOptions || []
  return [...options].sort((a, b) => a.value - b.value)
}

function scoreOptionLabel(option: ShuangxiAScoreOption) {
  return option.label || `${option.value}分`
}

function scoreOptionDescription(option: ShuangxiAScoreOption) {
  return option.description || scoreOptionLabel(option)
}

function scoreTone(score: number) {
  if (score >= 3)
    return 'green'
  if (score >= 1)
    return 'blue'
  return 'red'
}

function displayItemTitle(item?: Partial<ShuangxiAItemSummary>) {
  return normalizeText(item?.testItem || item?.itemTitle || item?.itemCode, '-')
}

function stripLeadingCode(text: string, code: string) {
  if (!code)
    return text
  return text
    .replace(new RegExp(`^\\s*${code.replace(/\./g, '\\.')}\\s*[-—、.．]?\\s*`), '')
    .trim() || text
}

function numericCode(value?: string, maxParts = 0) {
  const code = String(value || '').trim().match(/\d+(?:\.\d+)*/)?.[0] || ''
  if (!code)
    return ''
  if (!maxParts)
    return code
  return code.split('.').slice(0, maxParts).join('.')
}

function displayNumberedItemTitle(item?: Partial<ShuangxiAItemSummary>) {
  const title = displayItemTitle(item)
  const code = numericCode(item?.itemCode)
  if (!code || title.startsWith(code))
    return title
  return `${code} ${title}`
}

function displaySkillTitle(item?: Partial<ShuangxiAItemSummary>) {
  const name = normalizeText(item?.skillName || currentSkillName.value, '-')
  const code = numericCode(item?.itemCode, 2) || numericCode(item?.skillCode)
  if (!code || name.startsWith(code))
    return name
  return `${code} ${name}`
}

function displaySidebarSkillTitle(skill: ShuangxiASkillSummary) {
  const name = normalizeText(skill.skillName, '-')
  const code = numericCode(skill.items?.[0]?.itemCode, 2) || numericCode(skill.skillCode)
  if (!code || name.startsWith(code))
    return name
  return `${code} ${name}`
}

function displaySidebarItemCode(item: ShuangxiAItemSummary) {
  return numericCode(item.itemCode) || `第 ${item.itemNo} 题`
}

function displaySidebarItemTitle(item: ShuangxiAItemSummary) {
  return stripLeadingCode(displayItemTitle(item), numericCode(item.itemCode))
}

function syncSelectedDomain() {
  const item = currentItemSummary.value || currentItem.value
  if (item?.domainCode)
    selectedDomainCode.value = item.domainCode
}

async function revealSelectedItem() {
  await nextTick()
  const el = skillListRef.value?.querySelector(`[data-item-no="${selectedItemNo.value}"]`) as HTMLElement | null
  el?.scrollIntoView({ block: 'nearest' })
}

async function scrollSkillListToTop() {
  await nextTick()
  if (skillListRef.value)
    skillListRef.value.scrollTop = 0
}

function ensureStudentGender() {
  if (normalizeGender(editor.studentGender) || !editor.studentId)
    return
  selectedGenderDraft.value = ''
  genderModalOpen.value = true
}

async function confirmStudentGender() {
  if (!selectedGenderDraft.value) {
    messageService.warning('请选择学生性别')
    return
  }
  if (!editor.studentId)
    return
  genderSaving.value = true
  try {
    const gender = selectedGenderDraft.value === 'male' ? '男' : '女'
    const res = await updateScaleAssessmentStudentGenderApi({
      studentId: editor.studentId,
      gender,
    })
    const data = unwrap<any>(res)
    editor.studentGender = data?.gender || gender
    genderModalOpen.value = false
    messageService.success('学生性别已更新')
    if (!isRecordEditMode.value)
      await saveDraft(true)
  }
  catch (error: any) {
    messageService.error(getErrorMessage(error, '学生性别更新失败'))
  }
  finally {
    genderSaving.value = false
  }
}

function genderDefaultScoreForItem(itemNo: number) {
  const gender = normalizeGender(editor.studentGender)
  if (gender === 'male' && itemNo === 82)
    return 0
  if (gender === 'female' && itemNo === 83)
    return 0
  return undefined
}

function genderDefaultReasonForItem(itemNo: number) {
  const gender = normalizeGender(editor.studentGender)
  if (gender === 'male' && itemNo === 82)
    return '该题为女性生理项目，男生默认计 0 分。'
  if (gender === 'female' && itemNo === 83)
    return '该题为男性生理项目，女生默认计 0 分。'
  return ''
}

function normalizeGender(value?: string): NormalizedGender {
  const text = String(value || '').trim().toLowerCase()
  if (!text)
    return ''
  if (['男', '男性', 'male', 'm', '1'].includes(text))
    return 'male'
  if (['女', '女性', 'female', 'f', '2'].includes(text))
    return 'female'
  return ''
}

function clearRecord<T>(record: Record<number, T>) {
  Object.keys(record).forEach(key => delete record[Number(key)])
}

function unwrap<T>(res: any): T {
  return (res?.data ?? res?.result ?? res) as T
}

function numberFromQuery(key: string) {
  const raw = route.query[key]
  const value = Array.isArray(raw) ? raw[0] : raw
  const number = Number(value)
  return Number.isFinite(number) && number > 0 ? number : 0
}

function textFromQuery(key: string) {
  const raw = route.query[key]
  const value = Array.isArray(raw) ? raw[0] : raw
  return String(value || '').trim()
}

function normalizeDateText(value?: string) {
  if (!value)
    return ''
  const parsed = dayjs(value)
  return parsed.isValid() ? parsed.format('YYYY-MM-DD') : ''
}

function formatDate(value?: string) {
  return normalizeDateText(value) || '-'
}

function assessmentAgeText(birthDate?: string, assessmentDate?: string) {
  const birth = dayjs(birthDate)
  const assess = dayjs(assessmentDate)
  if (!birth.isValid() || !assess.isValid() || assess.isBefore(birth))
    return ''
  let years = assess.year() - birth.year()
  let months = assess.month() - birth.month()
  let days = assess.date() - birth.date()
  if (days < 0) {
    months -= 1
    days += assess.subtract(1, 'month').daysInMonth()
  }
  if (months < 0) {
    years -= 1
    months += 12
  }
  return `${years}岁${months}月${days}天`
}

function normalizeText(value?: string, fallback = '暂无') {
  const text = String(value || '').trim()
  return text || fallback
}

function isValidScore(value: unknown) {
  const number = Number(value)
  return Number.isFinite(number) && number >= 0
}

function getErrorMessage(error: any, fallback: string) {
  return error?.response?.data?.message || error?.message || fallback
}

function goBack() {
  void router.push('/teacherCenter/scale-library')
}
</script>

<template>
  <div class="shuangxi-workbench-page">
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
      <span class="header-meta">性别：<b>{{ genderText }}</b></span>
      <span class="header-divider"></span>
      <span class="header-meta">评量日期：<b>{{ assessmentDateText }}</b></span>
      <span class="header-divider"></span>
      <span class="header-meta">评量者：<b>{{ examinerName }}</b></span>
      <div class="header-actions">
        <span
          class="auto-save-status"
          :class="{ 'is-saving': autoSaveState === 'saving', 'is-saved': autoSaveState === 'saved' }"
        >
          {{ autoSaveText }}
        </span>
        <a-button v-if="!isRecordEditMode" size="large" class="outline-action" :loading="saving" @click="saveDraft(false)">
          <template #icon>
            <SaveOutlined />
          </template>
          保存草稿
        </a-button>
        <a-button size="large" type="primary" class="primary-action" :loading="submitting" @click="submitDraft">
          <template #icon>
            <FileDoneOutlined />
          </template>
          {{ submitActionText }}
        </a-button>
      </div>
    </header>

    <section class="dimension-strip">
      <button
        v-for="domain in domainCards"
        :key="domain.code"
        type="button"
        class="dimension-card"
        :class="{ 'is-active': domain.active }"
        @click="selectDomain(domain.code)"
      >
        <span class="dimension-card__head">
          <b>
            <span class="dimension-card__icon">
              <component :is="domain.icon" />
            </span>
            <span class="dimension-card__name">{{ domain.name }}</span>
          </b>
          <strong>{{ domain.answered }}/{{ domain.total }}</strong>
        </span>
        <i><em :style="{ width: `${domain.percent}%` }"></em></i>
      </button>
    </section>

    <main class="workbench-main">
      <aside ref="skillListRef" class="page-sidebar">
        <div class="sidebar-title">
          <span>技能项目</span>
          <SlidersOutlined />
        </div>
        <div v-for="skill in selectedDomainSkills" :key="skill.skillCode" class="page-group">
          <div
            class="page-group__head"
            :class="{ 'is-active': isSkillActive(skill) }"
            @click="selectSkill(skill)"
          >
            <span class="page-group__title">
              <RightOutlined v-if="!isSkillActive(skill)" />
              <span v-else class="chevron-down">⌄</span>
              {{ displaySidebarSkillTitle(skill) }}
            </span>
            <span>{{ skillProgress(skill).answered }}/{{ skillProgress(skill).total }}</span>
          </div>
          <div class="page-group__progress">
            <div class="progress-line">
              <i :style="{ width: `${skillProgress(skill).percent}%` }"></i>
            </div>
            <span class="page-group__percent">{{ skillProgress(skill).percent }}%</span>
          </div>
          <div v-if="isSkillActive(skill)" class="question-list">
            <button
              v-for="item in skill.items"
              :key="item.itemNo"
              type="button"
              class="question-item"
              :class="`is-${itemStatus(item)}`"
              :data-item-no="item.itemNo"
              @click="selectItem(item.itemNo)"
            >
              <span>{{ displaySidebarItemCode(item) }}</span>
              <strong>{{ displaySidebarItemTitle(item) }}</strong>
              <CheckCircleFilled v-if="itemStatus(item) === 'done'" />
              <i v-else-if="itemStatus(item) === 'active'"></i>
              <b v-else></b>
            </button>
          </div>
        </div>
      </aside>

      <section class="question-panel">
        <a-spin :spinning="templateLoading || itemLoading">
          <div class="question-content">
            <div class="question-kicker-row">
              <div class="question-kicker">
                <i></i>
                <span>{{ currentSkillDisplayTitle }}</span>
              </div>
              <strong>第 {{ selectedItemNo || '-' }} 题</strong>
            </div>
            <h1 class="question-main-title">{{ currentItemDisplayTitle }}</h1>
            <div v-if="previousScore !== undefined && previousAssessmentDate" class="previous-score-banner">
              <ClockCircleOutlined />
              <span class="previous-score-banner__date">上次评量 {{ previousAssessmentDate }}</span>
              <strong>{{ previousScore }}分</strong>
              <em v-if="previousScoreOption">· {{ scoreOptionDescription(previousScoreOption) }}</em>
            </div>
            <div v-if="genderDefaultScoreForItem(selectedItemNo) !== undefined" class="fixed-score-tip">
              {{ genderDefaultReasonForItem(selectedItemNo) }}
            </div>

            <div class="score-section">
              <div class="score-options">
                <button
                  v-for="option in currentScoreOptions"
                  :key="option.value"
                  type="button"
                  class="score-option"
                  :class="[`score-${scoreTone(option.value)}`, {
                    'is-selected': currentScore === option.value,
                    'is-previous': previousScore === option.value,
                    'is-fixed': genderDefaultScoreForItem(selectedItemNo) === option.value,
                  }]"
                  @click="selectScore(option.value)"
                >
                  <i class="score-option__radio"></i>
                  <strong>{{ scoreOptionLabel(option) }}</strong>
                  <span>{{ scoreOptionDescription(option) }}</span>
                  <em v-if="previousScore === option.value && previousAssessmentDate" class="score-option__previous-badge">
                    上次 {{ previousAssessmentDate.slice(5) }}
                  </em>
                </button>
              </div>
            </div>
          </div>
        </a-spin>
      </section>

      <aside class="right-rail">
        <section class="right-panel">
          <div class="right-panel__section progress-card">
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
          </div>

          <div class="right-panel__section remark-card">
            <div class="remark-card__head">
              <h2>本题备注</h2>
              <span>{{ currentRemark.length }}/300</span>
            </div>
            <div class="remark-input-shell">
              <a-textarea
                v-model:value="currentRemark"
                class="remark-textarea"
                :auto-size="false"
                placeholder="可记录现场观察、辅助方式或异常情况"
                @blur="persistRemark"
              />
            </div>
          </div>

          <div class="right-panel__section missing-nav-card">
            <h2>缺题导航</h2>
            <div class="missing-nav-list">
              <button type="button" class="missing-nav-row" @click="autoNext = !autoNext">
                <span>自动下一题</span>
                <strong>{{ autoNext ? '已开启' : '已关闭' }}</strong>
              </button>
              <button type="button" class="missing-nav-row" @click="selectDomain(selectedDomain?.domainCode || '')">
                <span>当前维度</span>
                <strong>{{ currentDomainName }}</strong>
              </button>
              <button type="button" class="missing-nav-row">
                <span>当前技能</span>
                <strong>{{ currentSkillDisplayTitle }}</strong>
              </button>
              <button
                type="button"
                class="missing-nav-row"
                :disabled="!firstMissingNo"
                @click="jumpToMissingItem"
              >
                <span>第一缺题</span>
                <strong>{{ firstMissingNo ? displayNumberedItemTitle(summaryByNo(firstMissingNo)) : '无缺题' }}</strong>
              </button>
            </div>
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
      ok-text="继续测评"
      cancel-text="重新测评"
      :width="430"
      :closable="false"
      :mask-closable="false"
      @ok="continueExistingDraft"
      @cancel="restartAssessment"
    >
      <p class="modal-copy">
        当前儿童存在一份未提交的双溪课程评量表A草稿，可继续上次进度，也可以重新开始。
      </p>
    </a-modal>

    <a-modal
      v-model:open="genderModalOpen"
      title="确认学生性别"
      ok-text="确认"
      cancel-text="返回"
      :width="430"
      :confirm-loading="genderSaving"
      :closable="false"
      :mask-closable="false"
      @ok="confirmStudentGender"
      @cancel="goBack"
    >
      <p class="modal-copy">
        双溪课程评量表A包含按性别默认计分的题目，请先确认 {{ studentName }} 的性别。
      </p>
      <a-radio-group v-model:value="selectedGenderDraft" class="gender-options">
        <a-radio-button value="male">男</a-radio-button>
        <a-radio-button value="female">女</a-radio-button>
      </a-radio-group>
    </a-modal>
  </div>
</template>

<style scoped>
.shuangxi-workbench-page {
  display: flex;
  flex-direction: column;
  height: 100vh;
  min-height: 0;
  min-width: 1180px;
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
  cursor: pointer;
  background: transparent;
  border: 0;
  border-radius: 6px;
  font-size: 18px;
}

.back-button:hover {
  color: #155bdc;
  background: #eef4ff;
}

.workbench-title {
  flex: 0 0 auto;
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
  flex: 0 0 auto;
  color: #111827;
  font-size: 13px;
  white-space: nowrap;
}

.header-meta b {
  color: #1f2937;
  font-weight: 600;
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
}

.auto-save-status.is-saving {
  color: #2563eb;
}

.auto-save-status.is-saved {
  color: #475569;
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
}

.outline-action {
  color: #155bdc;
  border-color: #2f6bff;
}

.primary-action {
  background: #0757e6;
  box-shadow: 0 10px 20px rgba(7, 87, 230, 0.24);
}

.dimension-strip {
  display: grid;
  flex: 0 0 auto;
  grid-template-columns: repeat(auto-fit, minmax(128px, 1fr));
  gap: 10px;
  padding: 12px 18px 4px;
}

.dimension-card {
  display: grid;
  gap: 10px;
  min-width: 0;
  padding: 10px 12px;
  text-align: left;
  cursor: pointer;
  background: #fff;
  border: 1px solid #e5eaf2;
  border-radius: 8px;
  transition: border-color .16s ease, box-shadow .16s ease;
}

.dimension-card.is-active {
  border-color: #2563eb;
  box-shadow: 0 8px 18px rgba(37, 99, 235, 0.12);
}

.dimension-card__head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 10px;
  overflow: hidden;
}

.dimension-card b {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
  overflow: hidden;
  color: #334155;
  font-size: 13px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dimension-card__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  width: 18px;
  height: 18px;
  color: #2563eb;
  background: #eff6ff;
  border-radius: 5px;
  font-size: 12px;
}

.dimension-card__name {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dimension-card strong {
  flex: 0 0 auto;
  color: #2563eb;
  font-size: 13px;
}

.dimension-card i {
  display: block;
  height: 4px;
  overflow: hidden;
  background: #e5e7eb;
  border-radius: 999px;
}

.dimension-card em {
  display: block;
  height: 100%;
  background: #2563eb;
  border-radius: inherit;
}

.workbench-main {
  display: grid;
  grid-template-columns: 300px minmax(420px, 1fr) 292px;
  flex: 1 1 auto;
  gap: 10px;
  min-height: 0;
  overflow: hidden;
  padding: 10px 10px 0;
}

.page-sidebar,
.question-panel,
.right-rail {
  min-height: 0;
  max-height: 100%;
}

.page-sidebar {
  overflow-y: auto;
  padding: 0;
  overscroll-behavior: contain;
  scrollbar-color: #cbd5e1 transparent;
  scrollbar-gutter: stable;
  scrollbar-width: thin;
  background: rgba(255, 255, 255, 0.96);
  border: 1px solid #e1e7f0;
  border-radius: 8px;
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.06);
}

.page-sidebar::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

.page-sidebar::-webkit-scrollbar-track {
  background: transparent;
}

.page-sidebar::-webkit-scrollbar-thumb {
  background: #cbd5e1;
  border-radius: 999px;
}

.page-sidebar::-webkit-scrollbar-thumb:hover {
  background: #94a3b8;
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
  color: #111827;
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
  align-items: center;
  gap: 8px;
  min-width: 0;
  color: #111827;
  font-weight: 700;
}

.page-group__title :deep(.anticon) {
  flex: 0 0 auto;
  font-size: 12px;
}

.page-group__title {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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
}

.progress-line i {
  display: block;
  height: 100%;
  background: #18a957;
  border-radius: inherit;
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
  position: relative;
  display: grid;
  grid-template-columns: 50px minmax(0, 1fr) 16px;
  align-items: center;
  width: 100%;
  min-height: 30px;
  padding: 0 8px 0 26px;
  color: #4b5563;
  cursor: pointer;
  background: transparent;
  border: 0;
  border-radius: 0;
  font-size: 12px;
  text-align: left;
}

.question-item span {
  white-space: nowrap;
}

.question-item strong {
  min-width: 0;
  overflow: hidden;
  color: inherit;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.question-item :deep(.anticon) {
  color: #18a957;
  font-size: 14px;
}

.question-item i,
.question-item b {
  display: inline-block;
  width: 14px;
  height: 14px;
  border-radius: 50%;
}

.question-item i {
  background: radial-gradient(circle at center, #fff 19%, #1769e8 22% 100%);
}

.question-item b {
  border: 1px solid #b8c1cf;
}

.question-item.is-active {
  position: relative;
  color: #0757e6;
  background: #eaf3ff;
  font-weight: 800;
}

.question-item.is-active::before {
  position: absolute;
  left: 0;
  width: 4px;
  height: 100%;
  background: #0757e6;
  border-radius: 0 4px 4px 0;
  content: "";
}

.question-panel {
  overflow-y: auto;
  padding: 20px 24px 0;
  background: #fff;
  border: 1px solid #e5eaf2;
  border-radius: 10px;
}

.question-panel :deep(.ant-spin-nested-loading),
.question-panel :deep(.ant-spin-container) {
  min-height: 100%;
}

.question-panel :deep(.ant-spin-container) {
  display: flex;
  flex-direction: column;
}

.question-content {
  display: flex;
  flex-direction: column;
  min-height: 100%;
  padding-bottom: 22px;
}

.question-kicker-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.question-kicker {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
  color: #2563eb;
  font-size: 13px;
  font-weight: 800;
}

.question-kicker i {
  width: 4px;
  height: 18px;
  background: #2563eb;
  border-radius: 999px;
}

.question-kicker span {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.question-kicker-row > strong {
  color: #64748b;
  font-size: 13px;
  font-weight: 700;
}

.question-main-title {
  min-width: 0;
  margin: 0 0 14px;
  color: #111827;
  font-size: 24px;
  font-weight: 800;
  line-height: 32px;
  text-overflow: ellipsis;
}

.previous-score-banner {
  display: flex;
  align-items: center;
  gap: 8px;
  min-height: 38px;
  padding: 6px 12px;
  margin-bottom: 14px;
  color: #166534;
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 700;
}

.previous-score-banner strong {
  color: #16a34a;
}

.previous-score-banner__date {
  color: #111827;
}

.previous-score-banner em {
  min-width: 0;
  overflow: hidden;
  color: #15803d;
  font-style: normal;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.instruction-card {
  padding: 16px;
  margin-bottom: 14px;
  background: #f8fafc;
  border: 1px solid #e5eaf2;
  border-radius: 10px;
}

.instruction-card h2,
.score-section h2,
.right-panel h2 {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0 0 10px;
  color: #111827;
  font-size: 15px;
  font-weight: 700;
}

.instruction-card p {
  margin: 6px 0 0;
  color: #475569;
  font-size: 14px;
  line-height: 1.7;
}

.score-section {
  padding: 0;
  margin-top: 0;
  background: transparent;
  border: 0;
  border-radius: 0;
  box-shadow: none;
}

.score-section__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.fixed-score-tip {
  padding: 8px 12px;
  margin: -8px 0 16px;
  color: #64748b;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 13px;
}

.previous-score-summary {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 4px 8px;
  margin-left: auto;
  color: #475569;
  font-size: 12px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
}

.previous-score-summary strong {
  color: #2563eb;
}

.score-options {
  display: grid;
  grid-template-columns: 1fr;
  gap: 10px;
}

.score-option {
  position: relative;
  display: grid;
  grid-template-columns: 24px 48px minmax(0, 1fr) auto;
  align-items: center;
  gap: 10px;
  min-height: 56px;
  padding: 10px 16px;
  text-align: left;
  cursor: pointer;
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 10px;
}

.score-option strong {
  margin: 0;
  color: #111827;
  font-size: 17px;
  font-weight: 700;
  white-space: nowrap;
}

.score-option span {
  min-width: 0;
  overflow: hidden;
  color: #111827;
  font-size: 13px;
  font-weight: 700;
  line-height: 1.5;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.score-option.is-selected {
  border-color: #2563eb;
  box-shadow: 0 10px 22px rgba(37, 99, 235, 0.14);
}

.score-option.is-previous:not(.is-selected) {
  background: #fbfdff;
  border-color: #cbd5e1;
}

.score-option.score-green.is-selected,
.score-option.score-red.is-selected {
  border-color: #2563eb;
}

.score-option__radio {
  position: relative;
  width: 16px;
  height: 16px;
  background-color: transparent;
  border: 1px solid #d9d9d9;
  border-radius: 50%;
  transition: border-color 0.2s ease;
}

.score-option__radio::after {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 8px;
  height: 8px;
  content: '';
  background-color: #2563eb;
  border-radius: 50%;
  opacity: 0;
  transform: translate(-50%, -50%) scale(0);
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.score-option.is-selected .score-option__radio {
  border-color: #2563eb;
}

.score-option.is-selected .score-option__radio::after {
  opacity: 1;
  transform: translate(-50%, -50%) scale(1);
}

.score-option__check {
  position: absolute;
  top: 10px;
  right: 10px;
  color: #2563eb;
  font-size: 18px;
}

.score-option__previous-badge {
  justify-self: end;
  padding: 4px 9px;
  color: #15803d;
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
  border-radius: 7px;
  font-size: 12px;
  font-style: normal;
  font-weight: 700;
  white-space: nowrap;
}

.right-rail {
  min-height: 0;
  overflow: hidden;
}

.right-panel {
  display: grid;
  grid-template-rows: auto minmax(88px, 1fr) auto;
  gap: 10px;
  height: 100%;
  min-height: 0;
  padding: 12px;
  background: #fff;
  border: 1px solid #e5eaf2;
  border-radius: 10px;
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.06);
  overflow: hidden;
}

.right-panel__section {
  min-width: 0;
}

.progress-card,
.remark-card {
  padding-bottom: 10px;
  border-bottom: 1px solid #eef2f7;
}

.remark-card {
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-height: 0;
}

.remark-card__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.remark-card__head h2 {
  margin-bottom: 0;
}

.remark-card__head span {
  color: #94a3b8;
  font-size: 12px;
  white-space: nowrap;
}

.remark-input-shell {
  display: flex;
  flex: 1 1 auto;
  min-height: 0;
}

.remark-input-shell :deep(.ant-input),
.remark-input-shell :deep(textarea) {
  height: 100% !important;
  min-height: 0 !important;
  resize: none;
}

.progress-card__body {
  display: flex;
  align-items: center;
  gap: 12px;
}

.donut {
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 72px;
  width: 72px;
  height: 72px;
  color: #2563eb;
  font-size: 16px;
  font-weight: 800;
  border-radius: 50%;
}

.progress-stats {
  display: grid;
  gap: 2px;
}

.progress-stats span {
  color: #64748b;
  font-size: 12px;
}

.progress-stats strong {
  color: #111827;
  font-size: 16px;
}

.progress-stats i {
  color: #94a3b8;
  font-size: 12px;
  font-style: normal;
}

.progress-stats .danger {
  color: #dc2626;
}

.missing-nav-list {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 6px;
}

.missing-nav-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  align-content: center;
  gap: 2px;
  min-height: 42px;
  padding: 5px 8px;
  color: #64748b;
  text-align: left;
  cursor: pointer;
  background: #f8fafc;
  border: 1px solid #e5eaf2;
  border-radius: 7px;
  font-size: 13px;
}

.missing-nav-row:disabled {
  cursor: not-allowed;
  opacity: 0.68;
}

.missing-nav-row strong {
  min-width: 0;
  overflow: hidden;
  color: #1f2937;
  font-weight: 700;
  font-size: 12px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 1400px) {
  .workbench-main {
    grid-template-columns: 260px minmax(400px, 1fr) 240px;
    gap: 8px;
    padding-inline: 8px;
  }

  .right-rail {
    min-height: 0;
  }

  .right-panel {
    grid-template-rows: auto minmax(72px, 1fr) auto;
    gap: 8px;
    padding: 10px;
    border-radius: 8px;
  }

  .progress-card__body {
    gap: 12px;
  }

  .donut {
    flex-basis: 62px;
    width: 62px;
    height: 62px;
    font-size: 14px;
  }

  .progress-stats strong {
    font-size: 15px;
  }

  .missing-nav-row {
    min-height: 40px;
    padding-inline: 7px;
    font-size: 12px;
  }

  .remark-card {
    min-height: 0;
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
}

.question-counter strong {
  color: #1f2937;
  font-size: 24px;
  letter-spacing: 0;
}

.question-counter span {
  margin-left: 6px;
  color: #1f2937;
  font-size: 14px;
}

.auto-next {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  color: #374151;
  font-size: 12px;
}

.modal-copy {
  margin: 0 0 16px;
  color: #475569;
  line-height: 1.7;
}

.gender-options {
  width: 100%;
}

.gender-options :deep(.ant-radio-button-wrapper) {
  width: 50%;
  text-align: center;
}
</style>
