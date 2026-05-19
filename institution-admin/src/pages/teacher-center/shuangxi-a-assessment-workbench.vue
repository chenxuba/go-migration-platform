<script setup lang="ts">
import {
  ArrowLeftOutlined,
  CheckCircleFilled,
  FileDoneOutlined,
  FileTextOutlined,
  LeftOutlined,
  RightOutlined,
  SaveOutlined,
  SlidersOutlined,
  SwapOutlined,
} from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import { Modal } from 'ant-design-vue'
import { computed, nextTick, onMounted, reactive, ref, watch } from 'vue'
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
const currentItemTitle = computed(() => displayItemTitle(currentItem.value || currentItemSummary.value))
const currentSkillName = computed(() => currentItemSummary.value?.skillName || currentItem.value?.skillName || '-')
const currentDomainName = computed(() => currentItemSummary.value?.domainName || currentItem.value?.domainName || '-')
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
    answered,
    total: items.length,
    percent,
    active: domain.domainCode === selectedDomainCode.value,
  }
}))

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
      goNextItem()
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
      goNextItem()
  }
  catch (error: any) {
    const message = getErrorMessage(error, `第 ${itemNo} 题自动保存失败`)
    draftItemSaveStatus.value = { ...draftItemSaveStatus.value, [itemNo]: 'error' }
    draftItemSaveErrors.value = { ...draftItemSaveErrors.value, [itemNo]: message }
    messageService.error(message)
  }
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
  selectedItemNo.value = itemNo
}

function goPreviousItem() {
  if (!hasPreviousItem.value)
    return
  selectedItemNo.value = allItems.value[currentIndex.value - 1].itemNo
}

function goNextItem() {
  if (!hasNextItem.value)
    return
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
  return [...options].sort((a, b) => b.value - a.value)
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
        <span>{{ domain.name }}</span>
        <strong>{{ domain.answered }}/{{ domain.total }}</strong>
        <i><em :style="{ width: `${domain.percent}%` }"></em></i>
      </button>
    </section>

    <main class="workbench-main">
      <aside ref="skillListRef" class="skill-sidebar">
        <div class="sidebar-title">
          <span>技能项目</span>
          <SlidersOutlined />
        </div>
        <div v-for="skill in selectedDomainSkills" :key="skill.skillCode" class="skill-group">
          <div class="skill-group__head" :class="{ 'is-active': skill.skillCode === selectedSkillCode }">
            <span>{{ skill.skillName }}</span>
            <strong>{{ skillProgress(skill).answered }}/{{ skillProgress(skill).total }}</strong>
          </div>
          <div class="skill-group__progress">
            <i :style="{ width: `${skillProgress(skill).percent}%` }"></i>
          </div>
          <div class="question-list">
            <button
              v-for="item in skill.items"
              :key="item.itemNo"
              type="button"
              class="question-item"
              :class="`is-${itemStatus(item)}`"
              :data-item-no="item.itemNo"
              @click="selectItem(item.itemNo)"
            >
              <span>第 {{ item.itemNo }} 题</span>
              <strong>{{ displayItemTitle(item) }}</strong>
              <CheckCircleFilled v-if="itemStatus(item) === 'done'" />
              <i v-else-if="itemStatus(item) === 'active'"></i>
              <b v-else></b>
            </button>
          </div>
        </div>
      </aside>

      <section class="question-panel">
        <a-spin :spinning="templateLoading || itemLoading">
          <div class="question-title-row">
            <h1>第 {{ selectedItemNo || '-' }} 题&nbsp;&nbsp;{{ currentItemTitle }}</h1>
            <a-tag color="blue">{{ currentDomainName }} / {{ currentSkillName }}</a-tag>
          </div>

          <article class="instruction-card">
            <h2><FileTextOutlined />评量项目</h2>
            <p>{{ normalizeText(currentItem?.testItem || currentItemSummary?.testItem || currentItem?.itemTitle || currentItemSummary?.itemTitle) }}</p>
          </article>

          <article class="instruction-card">
            <h2><FileTextOutlined />评分说明</h2>
            <p v-for="option in currentScoreOptions" :key="option.value">
              <b>{{ scoreOptionLabel(option) }}：</b>{{ scoreOptionDescription(option) }}
            </p>
          </article>

          <div class="score-section">
            <div class="score-section__head">
              <h2>评分</h2>
              <div v-if="previousScore !== undefined && previousAssessmentDate" class="previous-score-summary">
                <span>上次评量 {{ previousAssessmentDate }}</span>
                <strong>{{ previousScore }}分</strong>
              </div>
              <span v-if="genderDefaultScoreForItem(selectedItemNo) !== undefined" class="fixed-score-tip">
                {{ genderDefaultReasonForItem(selectedItemNo) }}
              </span>
            </div>
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
                <strong>{{ scoreOptionLabel(option) }}</strong>
                <span>{{ scoreOptionDescription(option) }}</span>
                <em v-if="previousScore === option.value && previousAssessmentDate" class="score-option__previous-badge">
                  上次 {{ previousAssessmentDate.slice(5) }}
                </em>
                <CheckCircleFilled v-if="currentScore === option.value" class="score-option__check" />
              </button>
            </div>
          </div>
        </a-spin>
      </section>

      <aside class="right-rail">
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

        <section class="right-card remark-card">
          <h2>本题备注</h2>
          <a-textarea
            v-model:value="currentRemark"
            :auto-size="{ minRows: 6, maxRows: 8 }"
            placeholder="可记录现场观察、辅助方式或异常情况"
            @blur="persistRemark"
          />
        </section>

        <section class="right-card current-card">
          <h2>当前定位</h2>
          <div class="current-meta">
            <span>领域</span><strong>{{ currentDomainName }}</strong>
            <span>技能</span><strong>{{ currentSkillName }}</strong>
            <span>题号</span><strong>第 {{ selectedItemNo || '-' }} 题</strong>
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
  min-width: 1180px;
  overflow: hidden;
  color: #1f2937;
  background: #f4f7fb;
}

.workbench-header {
  display: flex;
  align-items: center;
  flex: 0 0 72px;
  gap: 12px;
  padding: 0 22px;
  background: #fff;
  border-bottom: 1px solid #e5eaf2;
  box-shadow: 0 6px 18px rgba(15, 23, 42, 0.05);
}

.back-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  color: #334155;
  cursor: pointer;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
}

.workbench-title {
  flex: 0 0 auto;
  color: #111827;
  font-size: 18px;
  font-weight: 700;
}

.header-divider {
  width: 1px;
  height: 18px;
  background: #e5e7eb;
}

.header-meta {
  flex: 0 0 auto;
  color: #64748b;
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
  gap: 10px;
  margin-left: auto;
}

.auto-save-status {
  color: #64748b;
  font-size: 13px;
  white-space: nowrap;
}

.auto-save-status.is-saving {
  color: #2563eb;
}

.auto-save-status.is-saved {
  color: #16a34a;
}

.outline-action,
.primary-action {
  height: 38px;
  border-radius: 8px;
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
  gap: 6px;
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

.dimension-card span {
  overflow: hidden;
  color: #334155;
  font-size: 13px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dimension-card strong {
  color: #2563eb;
  font-size: 13px;
}

.dimension-card i,
.skill-group__progress {
  display: block;
  height: 4px;
  overflow: hidden;
  background: #e5e7eb;
  border-radius: 999px;
}

.dimension-card em,
.skill-group__progress i {
  display: block;
  height: 100%;
  background: #2563eb;
  border-radius: inherit;
}

.workbench-main {
  display: grid;
  grid-template-columns: 280px minmax(0, 1fr) 292px;
  flex: 1 1 auto;
  gap: 12px;
  min-height: 0;
  padding: 12px 18px 84px;
}

.skill-sidebar,
.question-panel,
.right-rail {
  min-height: 0;
}

.skill-sidebar {
  overflow-y: auto;
  padding: 14px;
  background: #fff;
  border: 1px solid #e5eaf2;
  border-radius: 10px;
}

.sidebar-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  color: #334155;
  font-size: 14px;
  font-weight: 700;
}

.skill-group {
  padding: 10px 0;
  border-top: 1px solid #eef2f7;
}

.skill-group:first-of-type {
  border-top: 0;
}

.skill-group__head {
  display: flex;
  justify-content: space-between;
  gap: 8px;
  color: #334155;
  font-size: 13px;
  font-weight: 600;
}

.skill-group__head.is-active {
  color: #2563eb;
}

.skill-group__progress {
  margin: 8px 0;
}

.question-list {
  display: grid;
  gap: 6px;
}

.question-item {
  position: relative;
  display: grid;
  grid-template-columns: 56px minmax(0, 1fr) 16px;
  align-items: center;
  gap: 8px;
  min-height: 38px;
  padding: 7px 8px;
  text-align: left;
  cursor: pointer;
  background: #f8fafc;
  border: 1px solid transparent;
  border-radius: 8px;
}

.question-item span {
  color: #64748b;
  font-size: 12px;
}

.question-item strong {
  overflow: hidden;
  color: #334155;
  font-size: 13px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.question-item.is-active {
  background: #eff6ff;
  border-color: #93c5fd;
}

.question-item.is-done :deep(.anticon) {
  color: #16a34a;
}

.question-item i,
.question-item b {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.question-item i {
  background: #2563eb;
}

.question-item b {
  border: 1px solid #cbd5e1;
}

.question-panel {
  overflow-y: auto;
  padding: 18px;
  background: #fff;
  border: 1px solid #e5eaf2;
  border-radius: 10px;
}

.question-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 14px;
}

.question-title-row h1 {
  min-width: 0;
  margin: 0;
  overflow: hidden;
  color: #111827;
  font-size: 22px;
  font-weight: 700;
  line-height: 32px;
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
.right-card h2 {
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
  padding: 16px;
  background: #fff;
  border: 1px solid #dbeafe;
  border-radius: 10px;
}

.score-section__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.fixed-score-tip {
  color: #64748b;
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
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 10px;
}

.score-option {
  position: relative;
  min-height: 94px;
  padding: 14px;
  text-align: left;
  cursor: pointer;
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 10px;
}

.score-option strong {
  display: block;
  margin-bottom: 8px;
  color: #111827;
  font-size: 20px;
  font-weight: 800;
}

.score-option span {
  color: #64748b;
  font-size: 13px;
  line-height: 1.5;
}

.score-option.is-selected {
  border-color: #2563eb;
  box-shadow: 0 10px 22px rgba(37, 99, 235, 0.14);
}

.score-option.is-previous:not(.is-selected) {
  background: #f8fafc;
  border-color: #cbd5e1;
}

.score-option.score-green.is-selected {
  border-color: #16a34a;
}

.score-option.score-red.is-selected {
  border-color: #ef4444;
}

.score-option__check {
  position: absolute;
  top: 10px;
  right: 10px;
  color: #2563eb;
  font-size: 18px;
}

.score-option__previous-badge {
  position: absolute;
  right: 10px;
  bottom: 8px;
  color: #64748b;
  font-size: 12px;
  font-style: normal;
}

.right-rail {
  display: grid;
  align-content: start;
  gap: 12px;
}

.right-card {
  padding: 16px;
  background: #fff;
  border: 1px solid #e5eaf2;
  border-radius: 10px;
}

.progress-card__body {
  display: flex;
  align-items: center;
  gap: 16px;
}

.donut {
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 88px;
  width: 88px;
  height: 88px;
  color: #2563eb;
  font-size: 18px;
  font-weight: 800;
  border-radius: 50%;
}

.progress-stats {
  display: grid;
  gap: 4px;
}

.progress-stats span {
  color: #64748b;
  font-size: 12px;
}

.progress-stats strong {
  color: #111827;
  font-size: 18px;
}

.progress-stats i {
  color: #94a3b8;
  font-size: 12px;
  font-style: normal;
}

.progress-stats .danger {
  color: #dc2626;
}

.current-meta {
  display: grid;
  grid-template-columns: 44px minmax(0, 1fr);
  gap: 8px 10px;
  color: #64748b;
  font-size: 13px;
}

.current-meta strong {
  min-width: 0;
  overflow: hidden;
  color: #1f2937;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.workbench-footer {
  position: fixed;
  right: 18px;
  bottom: 16px;
  left: 18px;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  height: 56px;
  background: rgba(255, 255, 255, 0.94);
  border: 1px solid #e5eaf2;
  border-radius: 12px;
  box-shadow: 0 10px 30px rgba(15, 23, 42, 0.12);
  backdrop-filter: blur(8px);
}

.nav-button,
.next-button {
  min-width: 108px;
  height: 38px;
  border-radius: 8px;
}

.question-counter {
  min-width: 74px;
  text-align: center;
}

.question-counter strong {
  color: #2563eb;
  font-size: 22px;
}

.question-counter span {
  color: #64748b;
  font-size: 13px;
}

.auto-next {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: #475569;
  font-size: 13px;
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
