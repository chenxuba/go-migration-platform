<script setup lang="ts">
import { CloseOutlined, ExclamationCircleFilled } from '@ant-design/icons-vue'
import { checkAssistantScheduleAvailabilityApi, checkOneToOneScheduleAvailabilityApi, type OneToOneScheduleAvailabilityItem, validateOneToOneSchedulesApi } from '@/api/edu-center/teaching-schedule'
import messageService from '@/utils/messageService'

interface ConflictWorkbenchPlan {
  date: string
  week: string
  rule: string
  time: string
  startTime: string
  endTime: string
  teacher: string
  assistant?: string
  classroom: string
  teacherId?: string
  assistantIds?: string[]
  classroomId?: string
  allowStudentConflict?: boolean
  allowClassroomConflict?: boolean
}

interface SimpleOption {
  value: string
  label: string
}

interface TimeOption {
  value: string
  label: string
  startTime: string
  endTime: string
}

interface PeriodGroupOption {
  key: string
  label: string
  teacherIds: string[]
  timeOptions: TimeOption[]
}

interface ConflictWorkbenchSubmitPayload {
  plans: ConflictWorkbenchPlan[]
  assistantIds?: string[]
}

interface ScheduleConflictItem {
  name?: string
  classTypeText?: string
  date?: string
  week?: string
  timeText?: string
  teacherId?: string
  teacherName?: string
  assistantNames?: string[]
  classroomName?: string
  studentNames?: string[]
  conflictTypes?: string[]
}

interface ConflictDetailView extends ScheduleConflictItem {
  key: string
  activeConflictTypes: string[]
  resolvedConflictTypes: string[]
  teacherTone: 'default' | 'danger' | 'success'
  assistantTone: 'default' | 'danger' | 'success'
  classroomTone: 'default' | 'danger' | 'success'
  studentTone: 'default' | 'danger' | 'success'
  isResolved: boolean
  isPartiallyResolved: boolean
}

const props = defineProps<{
  open: boolean
  oneToOneId: string
  assistantIds?: string[]
  validation?: {
    valid: boolean
    message?: string
    currentSchedules?: ScheduleConflictItem[]
    existingSchedules?: ScheduleConflictItem[]
    conflictTypes?: string[]
  } | null
  plans: ConflictWorkbenchPlan[]
  teacherOptions: SimpleOption[]
  assistantOptions?: SimpleOption[]
  classroomOptions: SimpleOption[]
  timeOptions: TimeOption[]
  periodGroups?: PeriodGroupOption[]
  defaultGroupKey?: string
  defaultTeacherId?: string
  defaultClassroomId?: string
  loading?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'submit', payload: ConflictWorkbenchSubmitPayload): void
}>()

interface WorkbenchRow {
  key: string
  index: number
  date: string
  week: string
  rule: string
  groupKey: string
  startTime: string
  endTime: string
  teacherId: string
  assistantIds: string[]
  classroomId: string
  allowStudentConflict: boolean
  allowClassroomConflict: boolean
}

interface SoftConflictSnapshotItem {
  allowStudentConflict: boolean
  allowClassroomConflict: boolean
}

interface TimeOptionAvailabilityView {
  status: 'free' | 'busy' | 'unknown'
  statusText: string
  reasonText: string
  conflictTypes: string[]
}

interface TimeOptionSelectView {
  value: string
  label: string
  baseLabel: string
  status: 'free' | 'busy' | 'unknown'
  statusText: string
  reasonText: string
}

interface AssistantOptionSelectView {
  value: string
  label: string
  baseLabel: string
  status: 'free' | 'busy' | 'unknown'
  statusText: string
}

const modalOpen = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const validating = ref(false)
const validationState = ref(props.validation || null)
const rowStates = ref<WorkbenchRow[]>([])
const initialConflictMap = ref<Record<string, string[]>>({})
const lastValidatedRowMap = ref<Record<string, WorkbenchRow>>({})
const bulkSoftConflictSnapshot = ref<Record<string, SoftConflictSnapshotItem> | null>(null)
let revalidateTimer: ReturnType<typeof setTimeout> | null = null
const timeOptionAvailabilityMap = ref<Record<string, TimeOptionAvailabilityView>>({})
const timeOptionAvailabilityLoading = ref(false)
let timeOptionAvailabilityTimer: ReturnType<typeof setTimeout> | null = null
let timeOptionAvailabilitySeq = 0
const assistantAvailabilityMap = ref<Record<string, TimeOptionAvailabilityView>>({})
const assistantAvailabilityLoading = ref(false)
let assistantAvailabilityTimer: ReturnType<typeof setTimeout> | null = null
let assistantAvailabilitySeq = 0

function parseTimeText(text?: string) {
  const m = String(text || '').match(/(\d{1,2}:\d{2})[~～](\d{1,2}:\d{2})/)
  if (!m)
    return null
  const toMinutes = (value: string) => {
    const [hour, minute] = value.split(':').map(Number)
    return hour * 60 + minute
  }
  return { start: toMinutes(m[1]), end: toMinutes(m[2]) }
}

function schedulesOverlap(
  current: { date?: string, timeText?: string },
  existing: { date?: string, timeText?: string },
) {
  if (current.date !== existing.date)
    return false
  const currentRange = parseTimeText(current.timeText)
  const existingRange = parseTimeText(existing.timeText)
  if (!currentRange || !existingRange)
    return false
  return currentRange.start < existingRange.end && currentRange.end > existingRange.start
}

function currentSchedulesByIndex() {
  return validationState.value?.currentSchedules || []
}

function uniqueConflictTypes(types: string[]) {
  return Array.from(new Set(types.filter(Boolean)))
}

function normalizeText(value?: string) {
  return String(value || '').trim()
}

function buildRowIdentity(date: string, startTime: string, endTime: string) {
  return `${date}|${startTime}|${endTime}`
}

function normalizedTeacherId(value: string) {
  return String(value || '').trim()
}

function normalizedAssistantIds(values?: Array<string | number>) {
  return Array.from(new Set((values || []).map(value => String(value || '').trim()).filter(Boolean)))
}

function allowedGroupOptionsByTeacher(teacherId: string) {
  const normalized = normalizedTeacherId(teacherId)
  return (props.periodGroups || []).filter(group =>
    !group.teacherIds.length || !normalized || group.teacherIds.includes(normalized),
  )
}

function timeValueForRow(row: Record<string, any>) {
  return `${row.groupKey}|${row.startTime}|${row.endTime}`
}

function buildTimeValue(groupKey: string, startTime: string, endTime: string) {
  return `${groupKey}|${startTime}|${endTime}`
}

function groupOptionsForRow(row: Record<string, any>) {
  const groups = allowedGroupOptionsByTeacher(row.teacherId)
  return groups.length ? groups : (props.periodGroups || [])
}

function timeOptionsForRow(row: Record<string, any>) {
  const groups = props.periodGroups || []
  const group = groups.find(item => item.key === row.groupKey) || groupOptionsForRow(row)[0]
  return group?.timeOptions || []
}

function buildTimeOptionAvailabilityKey(teacherId: string, lessonDate: string, startTime: string, endTime: string) {
  return `${normalizedTeacherId(teacherId)}|${lessonDate}|${startTime}|${endTime}`
}

function buildAssistantAvailabilityKey(rowKey: string, assistantId: string) {
  return `${rowKey}|${String(assistantId || '').trim()}`
}

function buildTimeOptionAvailabilityView(item: OneToOneScheduleAvailabilityItem): TimeOptionAvailabilityView {
  if (item.valid) {
    return {
      status: 'free',
      statusText: '空闲',
      reasonText: '当前老师该节次可排',
      conflictTypes: [],
    }
  }

  const conflictTypes = item.conflictTypes || []
  const reasonText = item.message || (conflictTypes.length ? `${conflictTypes.join('、')}冲突` : '该节次已有冲突')
  return {
    status: 'busy',
    statusText: '繁忙',
    reasonText,
    conflictTypes,
  }
}

function timeOptionAvailabilityFor(row: Record<string, any>, option: TimeOption): TimeOptionAvailabilityView {
  const teacherId = normalizedTeacherId(row.teacherId)
  if (!teacherId) {
    return {
      status: 'unknown',
      statusText: '先选老师',
      reasonText: '',
      conflictTypes: [],
    }
  }

  const key = buildTimeOptionAvailabilityKey(teacherId, row.date, option.startTime, option.endTime)
  return timeOptionAvailabilityMap.value[key] || {
    status: 'unknown',
    statusText: timeOptionAvailabilityLoading.value ? '检测中' : '待检测',
    reasonText: '',
    conflictTypes: [],
  }
}

function timeOptionViewsForRow(row: Record<string, any>): TimeOptionSelectView[] {
  return timeOptionsForRow(row).map((option) => {
    const availability = timeOptionAvailabilityFor(row, option)
    return {
      value: buildTimeValue(row.groupKey, option.startTime, option.endTime),
      label: `${option.label} · ${availability.statusText}`,
      baseLabel: option.label,
      status: availability.status,
      statusText: availability.statusText,
      reasonText: availability.reasonText,
    }
  })
}

function assistantAvailabilityForRow(row: Record<string, any>, assistantId: string): TimeOptionAvailabilityView {
  const normalized = String(assistantId || '').trim()
  if (!normalized) {
    return {
      status: 'unknown',
      statusText: '待定',
      reasonText: '',
      conflictTypes: [],
    }
  }
  return assistantAvailabilityMap.value[buildAssistantAvailabilityKey(row.key, normalized)] || {
    status: 'unknown',
    statusText: assistantAvailabilityLoading.value ? '检测中' : '待检测',
    reasonText: '',
    conflictTypes: [],
  }
}

function assistantOptionViewsForRow(row: Record<string, any>): AssistantOptionSelectView[] {
  return (props.assistantOptions || [])
    .filter(option => option.value !== row.teacherId)
    .map((option) => {
      const availability = assistantAvailabilityForRow(row, option.value)
      return {
        value: option.value,
        label: `${option.label} · ${availability.statusText}`,
        baseLabel: option.label,
        status: availability.status,
        statusText: availability.statusText,
      }
    })
}

function syncValidatedRowSnapshot(rows = rowStates.value) {
  lastValidatedRowMap.value = rows.reduce<Record<string, WorkbenchRow>>((map, row) => {
    map[row.key] = { ...row }
    return map
  }, {})
}

function teacherConflictResolvedLocally(row: WorkbenchRow, conflictTypes: string[]) {
  const validatedRow = lastValidatedRowMap.value[row.key]
  if (!validatedRow || !conflictTypes.includes('老师'))
    return false
  return normalizedTeacherId(row.teacherId) !== normalizedTeacherId(validatedRow.teacherId)
}

function classroomConflictResolvedLocally(row: WorkbenchRow, conflictTypes: string[]) {
  const validatedRow = lastValidatedRowMap.value[row.key]
  if (!validatedRow || !conflictTypes.includes('教室'))
    return false
  return normalizeText(row.classroomId) !== normalizeText(validatedRow.classroomId)
}

function matchesTeacherConflict(row: WorkbenchRow, item: ScheduleConflictItem) {
  const itemTeacherId = normalizedTeacherId(item.teacherId || '')
  const rowTeacherId = normalizedTeacherId(row.teacherId)
  if (itemTeacherId && rowTeacherId)
    return itemTeacherId === rowTeacherId
  return normalizeText(item.teacherName) === normalizeText(teacherNameById(row.teacherId))
}

function matchesClassroomConflict(row: WorkbenchRow, item: ScheduleConflictItem) {
  const itemClassroom = normalizeText(item.classroomName)
  const rowClassroom = normalizeText(classroomNameById(row.classroomId))
  if (!itemClassroom || itemClassroom === '-')
    return rowClassroom === '-' || !rowClassroom
  return itemClassroom === rowClassroom
}

function buildConflictDetailView(row: WorkbenchRow, item: ScheduleConflictItem, index: number): ConflictDetailView {
  const sourceConflictTypes = item.conflictTypes || []
  const resolvedConflictTypes = sourceConflictTypes.filter((type) => {
    if (type === '老师')
      return teacherConflictResolvedLocally(row, sourceConflictTypes) && !matchesTeacherConflict(row, item)
    if (type === '教室')
      return classroomConflictResolvedLocally(row, sourceConflictTypes) && !matchesClassroomConflict(row, item)
    return false
  })
  const activeConflictTypes = sourceConflictTypes.filter(type => !resolvedConflictTypes.includes(type))
  return {
    ...item,
    key: `${row.key}-${item.date || 'unknown'}-${item.timeText || 'unknown'}-${item.name || 'schedule'}-${index}`,
    activeConflictTypes,
    resolvedConflictTypes,
    teacherTone: activeConflictTypes.includes('老师') ? 'danger' : (resolvedConflictTypes.includes('老师') ? 'success' : 'default'),
    assistantTone: activeConflictTypes.includes('助教') ? 'danger' : (resolvedConflictTypes.includes('助教') ? 'success' : 'default'),
    classroomTone: activeConflictTypes.includes('教室') ? 'danger' : (resolvedConflictTypes.includes('教室') ? 'success' : 'default'),
    studentTone: activeConflictTypes.includes('学员') ? 'danger' : 'default',
    isResolved: !activeConflictTypes.length && resolvedConflictTypes.length > 0,
    isPartiallyResolved: activeConflictTypes.length > 0 && resolvedConflictTypes.length > 0,
  }
}

function resolvedConflictLabel(type: string) {
  if (type === '老师')
    return '原老师冲突已解除'
  if (type === '教室')
    return '原教室冲突已解除'
  return `已解除${type}冲突`
}

function toneClass(tone: 'default' | 'danger' | 'success') {
  return {
    'schedule-conflict__cell--danger': tone === 'danger',
    'schedule-conflict__cell--success': tone === 'success',
  }
}

const rowViews = computed(() =>
  rowStates.value.map((row, index) => {
    const current = currentSchedulesByIndex()[index]
    const conflictTypes = current?.conflictTypes || []
    const localResolvedConflictTypes = uniqueConflictTypes([
      teacherConflictResolvedLocally(row, conflictTypes) ? '老师' : '',
      classroomConflictResolvedLocally(row, conflictTypes) ? '教室' : '',
    ])
    const displayConflictTypes = conflictTypes.filter(type => !localResolvedConflictTypes.includes(type))
    const hasTeacherConflict = conflictTypes.includes('老师')
    const hasAssistantConflict = conflictTypes.includes('助教')
    const hasStudentConflict = conflictTypes.includes('学员')
    const hasClassroomConflict = conflictTypes.includes('教室')
    const displayHasTeacherConflict = displayConflictTypes.includes('老师')
    const displayHasAssistantConflict = displayConflictTypes.includes('助教')
    const displayHasStudentConflict = displayConflictTypes.includes('学员')
    const displayHasClassroomConflict = displayConflictTypes.includes('教室')
    const matches = (validationState.value?.existingSchedules || []).filter(item =>
      schedulesOverlap(
        current || { date: row.date, timeText: `${row.startTime}~${row.endTime}` },
        item,
      ),
    )
    const conflictDetails = matches.map((item, matchIndex) => buildConflictDetailView(row, item, matchIndex))
    const readyBySoftConflict = (!hasStudentConflict || row.allowStudentConflict)
      && (!hasClassroomConflict || row.allowClassroomConflict)
    const initialConflictTypes = initialConflictMap.value[row.key] || []
    const confirmedResolvedConflictTypes = initialConflictTypes.filter(type => !conflictTypes.includes(type))
    const resolvedConflictTypes = uniqueConflictTypes([
      ...confirmedResolvedConflictTypes,
      ...localResolvedConflictTypes,
    ])
    return {
      ...row,
      current,
      matches,
      conflictDetails,
      conflictTypes,
      displayConflictTypes,
      hasTeacherConflict,
      hasAssistantConflict,
      hasStudentConflict,
      hasClassroomConflict,
      displayHasTeacherConflict,
      displayHasAssistantConflict,
      displayHasStudentConflict,
      displayHasClassroomConflict,
      localResolvedConflictTypes,
      confirmedResolvedConflictTypes,
      resolvedConflictTypes,
      canCreate: !hasTeacherConflict && !hasAssistantConflict && readyBySoftConflict,
    }
  }),
)

function shouldRenderRowInWorkbench(row: { key: string }) {
  return (initialConflictMap.value[row.key] || []).length > 0
}

const workbenchRowViews = computed(() =>
  rowViews.value.filter(row => shouldRenderRowInWorkbench(row)),
)

const summary = computed(() => {
  const rows = workbenchRowViews.value
  const allRows = rowViews.value
  return {
    total: rows.length,
    teacher: rows.filter(row => row.hasTeacherConflict).length,
    assistant: rows.filter(row => row.hasAssistantConflict).length,
    student: rows.filter(row => row.hasStudentConflict).length,
    classroom: rows.filter(row => row.hasClassroomConflict).length,
    ready: allRows.filter(row => row.canCreate).length,
  }
})

const canEnableAllSoftConflicts = computed(() =>
  rowViews.value.some(row =>
    (row.displayHasStudentConflict && !row.allowStudentConflict)
    || (row.displayHasClassroomConflict && !row.allowClassroomConflict),
  ),
)

const canUndoEnableAllSoftConflicts = computed(() => {
  const snapshot = bulkSoftConflictSnapshot.value
  if (!snapshot)
    return false
  return rowStates.value.some((row) => {
    const previous = snapshot[row.key]
    if (!previous)
      return false
    return previous.allowStudentConflict !== row.allowStudentConflict
      || previous.allowClassroomConflict !== row.allowClassroomConflict
  })
})

function teacherNameById(id: string) {
  return props.teacherOptions.find(item => item.value === id)?.label || '-'
}

function classroomNameById(id: string) {
  if (!id)
    return '-'
  return props.classroomOptions.find(item => item.value === id)?.label || '-'
}

function classroomSelectValue(value?: string) {
  const normalized = String(value || '').trim()
  return normalized || undefined
}

function assistantNamesByIds(ids?: Array<string | number>) {
  const normalized = normalizedAssistantIds(ids)
  if (!normalized.length)
    return []
  const labels = normalized.map((id) => {
    return props.assistantOptions?.find(item => item.value === id)?.label || id
  }).filter(Boolean)
  return Array.from(new Set(labels))
}

function currentRowAssistantEntries(row: Record<string, any>) {
  return normalizedAssistantIds(row.assistantIds).map((id) => {
    const name = assistantNamesByIds([id])[0] || id
    const availability = assistantAvailabilityForRow(row, id)
    return {
      id,
      name,
      danger: availability.status === 'busy',
    }
  })
}

function selectedAssistantNameSetForRow(row: Record<string, any>) {
  return new Set(assistantNamesByIds(row.assistantIds).map(name => normalizeText(name)).filter(Boolean))
}

function selectedAssistantIdSetForRow(row: Record<string, any>) {
  return new Set(normalizedAssistantIds(row.assistantIds).map(id => normalizedTeacherId(id)).filter(Boolean))
}

function conflictItemAssistantEntries(row: Record<string, any>, item: ConflictDetailView) {
  return (item.assistantNames || []).filter(Boolean).map((name) => {
    return {
      name,
      danger: item.activeConflictTypes.includes('助教') && selectedAssistantNameSetForRow(row).has(normalizeText(name)),
    }
  })
}

function conflictItemTeacherTone(row: Record<string, any>, item: ConflictDetailView) {
  if (item.activeConflictTypes.includes('老师'))
    return 'danger'
  if (item.activeConflictTypes.includes('助教')) {
    const teacherId = normalizedTeacherId(item.teacherId || '')
    const teacherName = normalizeText(item.teacherName)
    if ((teacherId && selectedAssistantIdSetForRow(row).has(teacherId)) || (teacherName && selectedAssistantNameSetForRow(row).has(teacherName)))
      return 'danger'
  }
  if (item.resolvedConflictTypes.includes('老师'))
    return 'success'
  return 'default'
}

function buildValidationPayload() {
  const unionAssistantIds = Array.from(new Set(rowStates.value.flatMap(row => normalizedAssistantIds(row.assistantIds))))
  return {
    oneToOneId: props.oneToOneId,
    teacherId: props.defaultTeacherId || rowStates.value[0]?.teacherId || '',
    assistantIds: unionAssistantIds,
    classroomId: props.defaultClassroomId || '',
    schedules: rowStates.value.map(row => ({
      lessonDate: row.date,
      startTime: row.startTime,
      endTime: row.endTime,
      teacherId: row.teacherId,
      assistantIds: row.assistantIds,
      classroomId: row.classroomId,
    })),
  }
}

async function revalidateRows() {
  if (!props.oneToOneId || !rowStates.value.length)
    return
  validating.value = true
  try {
    const res = await validateOneToOneSchedulesApi(buildValidationPayload())
    if (res.code !== 200 || !res.result)
      throw new Error(res.message || '重新校验失败')
    validationState.value = res.result
    syncValidatedRowSnapshot()
  }
  catch (error: any) {
    console.error('revalidateRows failed', error)
    messageService.error(error?.response?.data?.message || error?.message || '重新校验失败')
  }
  finally {
    validating.value = false
  }
}

async function fetchTimeOptionAvailability() {
  const seq = ++timeOptionAvailabilitySeq
  const oneToOneId = String(props.oneToOneId || '').trim()
  const schedulesMap = new Map<string, {
    teacherId: string
    lessonDate: string
    startTime: string
    endTime: string
  }>()

  rowStates.value.forEach((row) => {
    const teacherId = normalizedTeacherId(row.teacherId)
    if (!teacherId)
      return
    timeOptionsForRow(row).forEach((option) => {
      const key = buildTimeOptionAvailabilityKey(teacherId, row.date, option.startTime, option.endTime)
      schedulesMap.set(key, {
        teacherId,
        lessonDate: row.date,
        startTime: option.startTime,
        endTime: option.endTime,
      })
    })
  })

  if (!props.open || !oneToOneId || !schedulesMap.size) {
    timeOptionAvailabilityMap.value = {}
    timeOptionAvailabilityLoading.value = false
    return
  }

  timeOptionAvailabilityLoading.value = true
  try {
    const res = await checkOneToOneScheduleAvailabilityApi({
      oneToOneId,
      schedules: Array.from(schedulesMap.values()),
    })
    if (seq !== timeOptionAvailabilitySeq)
      return
    if (res.code !== 200 || !res.result)
      throw new Error(res.message || '检测节次空闲状态失败')

    const nextMap: Record<string, TimeOptionAvailabilityView> = {}
    ;(res.result.items || []).forEach((item) => {
      nextMap[buildTimeOptionAvailabilityKey(item.teacherId, item.lessonDate, item.startTime, item.endTime)] = buildTimeOptionAvailabilityView(item)
    })
    timeOptionAvailabilityMap.value = nextMap
  }
  catch (error) {
    if (seq !== timeOptionAvailabilitySeq)
      return
    console.error('fetchTimeOptionAvailability failed', error)
    timeOptionAvailabilityMap.value = {}
  }
  finally {
    if (seq === timeOptionAvailabilitySeq)
      timeOptionAvailabilityLoading.value = false
  }
}

async function fetchAssistantAvailability() {
  const seq = ++assistantAvailabilitySeq
  const oneToOneId = String(props.oneToOneId || '').trim()
  const assistantIds = normalizedAssistantIds((props.assistantOptions || []).map(item => item.value))
  if (!props.open || !oneToOneId || !assistantIds.length || !rowStates.value.length) {
    assistantAvailabilityMap.value = {}
    assistantAvailabilityLoading.value = false
    return
  }

  assistantAvailabilityLoading.value = true
  try {
    const results = await Promise.all(
      rowStates.value.map(async (row) => {
        const res = await checkAssistantScheduleAvailabilityApi({
          oneToOneId,
          assistantIds,
          schedules: [{
            lessonDate: row.date,
            startTime: row.startTime,
            endTime: row.endTime,
          }],
        })
        return { rowKey: row.key, res }
      }),
    )
    if (seq !== assistantAvailabilitySeq)
      return

    const nextMap: Record<string, TimeOptionAvailabilityView> = {}
    results.forEach(({ rowKey, res }) => {
      if (res.code !== 200 || !res.result)
        throw new Error(res.message || '检测助教空闲状态失败')
      assistantIds.forEach((assistantId) => {
        nextMap[buildAssistantAvailabilityKey(rowKey, assistantId)] = {
          status: 'free',
          statusText: '空闲',
          reasonText: '',
          conflictTypes: [],
        }
      })
      ;(res.result.items || []).forEach((item) => {
        nextMap[buildAssistantAvailabilityKey(rowKey, item.assistantId)] = {
          status: item.valid ? 'free' : 'busy',
          statusText: item.valid ? '空闲' : '繁忙',
          reasonText: item.message || '',
          conflictTypes: item.conflictTypes || [],
        }
      })
    })
    assistantAvailabilityMap.value = nextMap
  }
  catch (error) {
    if (seq !== assistantAvailabilitySeq)
      return
    console.error('fetchAssistantAvailability failed', error)
    assistantAvailabilityMap.value = {}
  }
  finally {
    if (seq === assistantAvailabilitySeq)
      assistantAvailabilityLoading.value = false
  }
}

function scheduleRevalidate() {
  if (revalidateTimer)
    clearTimeout(revalidateTimer)
  revalidateTimer = setTimeout(() => {
    void revalidateRows()
  }, 240)
}

function scheduleTimeOptionAvailabilityCheck() {
  if (timeOptionAvailabilityTimer)
    clearTimeout(timeOptionAvailabilityTimer)
  timeOptionAvailabilityTimer = setTimeout(() => {
    void fetchTimeOptionAvailability()
  }, 180)
}

function scheduleAssistantAvailabilityCheck() {
  if (assistantAvailabilityTimer)
    clearTimeout(assistantAvailabilityTimer)
  assistantAvailabilityTimer = setTimeout(() => {
    void fetchAssistantAvailability()
  }, 180)
}

function initializeRows() {
  rowStates.value = props.plans.map((plan, index) => {
    const defaultTeacherId = String(plan.teacherId || props.defaultTeacherId || '').trim()
    const defaultClassroomId = String(plan.classroomId || props.defaultClassroomId || '').trim()
    const defaultGroupKey = String(props.defaultGroupKey || '').trim() || props.periodGroups?.[0]?.key || ''
    return {
      key: `${plan.date}|${plan.startTime}|${plan.endTime}|${index}`,
      index: index + 1,
      date: plan.date,
      week: plan.week,
      rule: plan.rule,
      groupKey: defaultGroupKey,
      startTime: plan.startTime,
      endTime: plan.endTime,
      teacherId: defaultTeacherId,
      assistantIds: normalizedAssistantIds(plan.assistantIds || props.assistantIds || []),
      classroomId: defaultClassroomId,
      allowStudentConflict: Boolean(plan.allowStudentConflict),
      allowClassroomConflict: Boolean(plan.allowClassroomConflict),
    }
  })
  validationState.value = props.validation || null
  bulkSoftConflictSnapshot.value = null
  initialConflictMap.value = {}
  props.plans.forEach((plan, index) => {
    const current = props.validation?.currentSchedules?.[index]
    initialConflictMap.value[`${plan.date}|${plan.startTime}|${plan.endTime}|${index}`] = current?.conflictTypes || []
  })
  syncValidatedRowSnapshot()
  scheduleTimeOptionAvailabilityCheck()
  scheduleAssistantAvailabilityCheck()
}

watch(
  () => props.open,
  (open) => {
    if (open)
      initializeRows()
    else {
      timeOptionAvailabilityMap.value = {}
      assistantAvailabilityMap.value = {}
    }
  },
  { immediate: true },
)

watch(
  () => props.validation,
  (value) => {
    if (props.open) {
      validationState.value = value || null
      syncValidatedRowSnapshot()
    }
  },
)

watch(
  () => [
    props.open ? '1' : '0',
    props.oneToOneId,
    (props.assistantOptions || []).map(item => item.value).join(','),
    rowStates.value.map(row => `${row.key}|${row.date}|${row.startTime}|${row.endTime}`).join(','),
  ].join('|'),
  () => {
    if (props.open)
      scheduleAssistantAvailabilityCheck()
  },
)

function updateRow<K extends keyof WorkbenchRow>(key: string, field: K, value: WorkbenchRow[K]) {
  rowStates.value = rowStates.value.map((row) => {
    if (row.key !== key)
      return row
    const next = {
      ...row,
      [field]: value,
    }
    if (field === 'teacherId') {
      const groups = groupOptionsForRow(next)
      if (groups.length && !groups.some(group => group.key === next.groupKey))
        next.groupKey = groups[0].key
      const allowedTimeOptions = timeOptionsForRow(next)
      if (allowedTimeOptions.length && !allowedTimeOptions.some(option => option.startTime === next.startTime && option.endTime === next.endTime)) {
        next.startTime = allowedTimeOptions[0].startTime
        next.endTime = allowedTimeOptions[0].endTime
      }
    }
    if (field === 'groupKey') {
      const allowedTimeOptions = timeOptionsForRow(next)
      if (allowedTimeOptions.length) {
        next.startTime = allowedTimeOptions[0].startTime
        next.endTime = allowedTimeOptions[0].endTime
      }
    }
    return next
  })
  if (field === 'teacherId') {
    rowStates.value = rowStates.value.map((row) => {
      const teacherId = normalizedTeacherId(row.teacherId)
      if (!teacherId)
        return row
      const nextAssistantIds = normalizedAssistantIds(row.assistantIds).filter(id => id !== teacherId)
      if (nextAssistantIds.length === row.assistantIds.length)
        return row
      return {
        ...row,
        assistantIds: nextAssistantIds,
      }
    })
    scheduleAssistantAvailabilityCheck()
  }
  if (field === 'teacherId' || field === 'groupKey')
    scheduleTimeOptionAvailabilityCheck()
  if (field === 'groupKey' || field === 'startTime' || field === 'endTime')
    scheduleAssistantAvailabilityCheck()
  if (field === 'teacherId' || field === 'classroomId' || field === 'startTime' || field === 'endTime' || field === 'groupKey')
    scheduleRevalidate()
}

function changeRowTime(key: string, value: string) {
  const [groupKey, startTime, endTime] = String(value).split('|')
  if (!groupKey || !startTime || !endTime)
    return
  rowStates.value = rowStates.value.map((row) => {
    if (row.key !== key)
      return row
    return {
      ...row,
      groupKey,
      startTime,
      endTime,
    }
  })
  scheduleAssistantAvailabilityCheck()
  scheduleRevalidate()
}

function enableAllSoftConflicts() {
  if (!canEnableAllSoftConflicts.value)
    return

  if (!bulkSoftConflictSnapshot.value) {
    bulkSoftConflictSnapshot.value = rowStates.value.reduce<Record<string, SoftConflictSnapshotItem>>((map, row) => {
      map[row.key] = {
        allowStudentConflict: row.allowStudentConflict,
        allowClassroomConflict: row.allowClassroomConflict,
      }
      return map
    }, {})
  }

  rowStates.value = rowStates.value.map((row) => {
    const currentView = rowViews.value.find(item => item.key === row.key)
    return {
      ...row,
      allowStudentConflict: currentView?.displayHasStudentConflict ? true : row.allowStudentConflict,
      allowClassroomConflict: currentView?.displayHasClassroomConflict ? true : row.allowClassroomConflict,
    }
  })
}

function restoreBulkSoftConflictSelection() {
  const snapshot = bulkSoftConflictSnapshot.value
  if (!snapshot)
    return

  rowStates.value = rowStates.value.map((row) => {
    const previous = snapshot[row.key]
    if (!previous)
      return row
    return {
      ...row,
      allowStudentConflict: previous.allowStudentConflict,
      allowClassroomConflict: previous.allowClassroomConflict,
    }
  })
  bulkSoftConflictSnapshot.value = null
}

function updateAssistantIds(key: string, value: Array<string | number>) {
  rowStates.value = rowStates.value.map((row) => {
    if (row.key !== key)
      return row
    const teacherId = normalizedTeacherId(row.teacherId)
    return {
      ...row,
      assistantIds: normalizedAssistantIds(value).filter(id => id !== teacherId),
    }
  })
  scheduleAssistantAvailabilityCheck()
  scheduleRevalidate()
}

function submitResolvedRows() {
  const selected = rowViews.value.filter(row => row.canCreate)
  if (!selected.length)
    return
  emit('submit', {
    plans: selected.map((row) => {
      const option = props.timeOptions.find(item => item.startTime === row.startTime && item.endTime === row.endTime)
      const group = (props.periodGroups || []).find(item => item.key === row.groupKey)
      return {
        date: row.date,
        week: row.week,
        rule: row.rule,
        time: option?.label || `${group?.label || ''} · ${row.startTime}-${row.endTime}`,
        startTime: row.startTime,
        endTime: row.endTime,
        teacher: teacherNameById(row.teacherId),
        assistant: assistantNamesByIds(row.assistantIds).join('、') || '未安排',
        classroom: classroomNameById(row.classroomId),
        teacherId: row.teacherId,
        assistantIds: [...row.assistantIds],
        classroomId: row.classroomId,
        allowStudentConflict: row.allowStudentConflict,
        allowClassroomConflict: row.allowClassroomConflict,
      }
    }),
    assistantIds: Array.from(new Set(selected.map(row => row.assistantIds).flat())),
  })
}

const columns = [
  { title: '待创建日程', key: 'current', dataIndex: 'current', width: '28%' },
  { title: '冲突详情', key: 'conflicts', dataIndex: 'conflicts', width: '36%' },
  { title: '处理方式', key: 'actions', dataIndex: 'actions', width: '36%' },
]
</script>

<template>
  <a-modal
    v-model:open="modalOpen"
    centered
    class="schedule-conflict-workbench-modal"
    :footer="null"
    :width="1320"
    :body-style="{ paddingTop: '0px' }"
    :keyboard="false"
    :closable="false"
    :mask-closable="true"
  >
    <template #title>
      <div class="schedule-conflict__titlebar">
        <span>冲突处理</span>
        <a-button type="text" @click="modalOpen = false">
          <template #icon>
            <CloseOutlined />
          </template>
        </a-button>
      </div>
    </template>

    <div class="schedule-conflict">
      <div class="schedule-conflict__banner">
        <ExclamationCircleFilled />
        <span>{{ validationState?.message || '当前创建日程存在冲突' }}</span>
      </div>

      <div class="schedule-conflict__toolbar">
        <div class="schedule-conflict__toolbar-summary">
          共 {{ summary.total }} 节待处理日程，其中老师冲突 {{ summary.teacher }} 节，助教冲突 {{ summary.assistant }} 节，学员冲突 {{ summary.student }} 节，教室冲突 {{ summary.classroom }} 节，当前可直接创建 {{ summary.ready }} 节。
        </div>
        <div class="schedule-conflict__toolbar-actions">
          <a-button
            v-if="canUndoEnableAllSoftConflicts"
            type="link"
            class="schedule-conflict__toolbar-link schedule-conflict__toolbar-link--muted"
            @click="restoreBulkSoftConflictSelection"
          >
            撤销批量忽略
          </a-button>
          <a-button
            type="link"
            class="schedule-conflict__toolbar-link"
            :disabled="!canEnableAllSoftConflicts"
            @click="enableAllSoftConflicts"
          >
            忽略全部软冲突
          </a-button>
          <a-button size="small" :loading="validating" @click="revalidateRows">
            重新校验
          </a-button>
        </div>
      </div>

      <a-table
        class="schedule-conflict__workbench"
        :columns="columns"
        :data-source="workbenchRowViews"
        :pagination="false"
        row-key="key"
        :scroll="{ y: 560 }"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'current'">
            <div class="schedule-conflict__cell-card">
              <div class="schedule-conflict__cell-top">
                <span class="schedule-conflict__group-index">第 {{ record.index }} 节待创建</span>
                <span class="schedule-conflict__group-time">{{ record.date }} {{ record.startTime }}~{{ record.endTime }}</span>
              </div>
              <div class="schedule-conflict__cell-main">
                <strong>{{ record.current?.name || '-' }}</strong>
                <span>{{ record.current?.classTypeText || '1对1日程' }}</span>
                <span class="schedule-conflict__group-type">{{ (props.periodGroups || []).find(item => item.key === record.groupKey)?.label || '-' }}</span>
              </div>
              <div class="schedule-conflict__cell-meta">
                <span>老师：<strong :class="{ 'schedule-conflict__cell--danger': record.displayHasTeacherConflict, 'schedule-conflict__cell--success': !record.displayHasTeacherConflict && record.resolvedConflictTypes.includes('老师') }">{{ teacherNameById(record.teacherId) }}</strong></span>
                <span>
                  助教：
                  <span class="schedule-conflict__name-list">
                    <template v-if="currentRowAssistantEntries(record).length">
                      <span
                        v-for="(assistant, assistantIndex) in currentRowAssistantEntries(record)"
                        :key="`${record.key}-assistant-${assistant.id}`"
                        :class="{ 'schedule-conflict__cell--danger': assistant.danger }"
                      >
                        {{ assistant.name }}<template v-if="assistantIndex < currentRowAssistantEntries(record).length - 1">、</template>
                      </span>
                    </template>
                    <span v-else>未安排</span>
                  </span>
                </span>
                <span>教室：<strong :class="{ 'schedule-conflict__cell--danger': record.displayHasClassroomConflict, 'schedule-conflict__cell--success': !record.displayHasClassroomConflict && record.resolvedConflictTypes.includes('教室') }">{{ classroomNameById(record.classroomId) }}</strong></span>
                <span>学员：<strong :class="{ 'schedule-conflict__cell--danger': record.displayHasStudentConflict }">{{ (record.current?.studentNames || []).join('、') || '-' }}</strong></span>
              </div>
            </div>
          </template>

          <template v-else-if="column.key === 'conflicts'">
            <div class="schedule-conflict__cell-stack">
              <div class="schedule-conflict__panel-title">
                命中冲突
              </div>
              <div
                v-for="item in record.conflictDetails"
                :key="item.key"
                class="schedule-conflict__conflict-item"
                :class="{
                  'schedule-conflict__conflict-item--resolved': item.isResolved,
                  'schedule-conflict__conflict-item--mixed': item.isPartiallyResolved,
                }"
              >
                <div class="schedule-conflict__cell-main">
                  <strong>{{ item.name }}</strong>
                  <span>{{ item.classTypeText }}</span>
                  <span>{{ item.date }} {{ item.timeText }}</span>
                </div>
                <div class="schedule-conflict__cell-meta">
                  <span>老师：<strong :class="toneClass(conflictItemTeacherTone(record, item))">{{ item.teacherName || '-' }}</strong></span>
                  <span>
                    助教：
                    <span class="schedule-conflict__name-list">
                      <template v-if="conflictItemAssistantEntries(record, item).length">
                        <span
                          v-for="(assistant, assistantIndex) in conflictItemAssistantEntries(record, item)"
                          :key="`${item.key}-assistant-${assistant.name}-${assistantIndex}`"
                          :class="{ 'schedule-conflict__cell--danger': assistant.danger }"
                        >
                          {{ assistant.name }}<template v-if="assistantIndex < conflictItemAssistantEntries(record, item).length - 1">、</template>
                        </span>
                      </template>
                      <span v-else>未安排</span>
                    </span>
                  </span>
                  <span>教室：<strong :class="toneClass(item.classroomTone)">{{ item.classroomName || '-' }}</strong></span>
                  <span>冲突学员：<strong :class="toneClass(item.studentTone)">{{ (item.studentNames || []).join('、') || '-' }}</strong></span>
                </div>
                <div class="schedule-conflict__tags">
                  <span
                    v-for="tag in item.activeConflictTypes"
                    :key="`${item.key}-active-${tag}`"
                    class="schedule-conflict__status-tag schedule-conflict__status-tag--danger"
                  >
                    {{ tag }}冲突
                  </span>
                  <span
                    v-for="tag in item.resolvedConflictTypes"
                    :key="`${item.key}-resolved-${tag}`"
                    class="schedule-conflict__status-tag schedule-conflict__status-tag--success"
                  >
                    {{ resolvedConflictLabel(tag) }}
                  </span>
                </div>
              </div>
            </div>
          </template>

          <template v-else-if="column.key === 'actions'">
            <div class="schedule-conflict__action-panel">
              <div class="schedule-conflict__panel-title">
                本行处理
              </div>

              <div class="schedule-conflict__action-group">
                <span class="schedule-conflict__action-label">时段组</span>
                <a-select
                  :value="record.groupKey"
                  :options="groupOptionsForRow(record).map(item => ({ value: item.key, label: item.label }))"
                  class="schedule-conflict__control"
                  @change="value => updateRow(record.key, 'groupKey', String(value))"
                />
              </div>

              <div class="schedule-conflict__action-group">
                <span class="schedule-conflict__action-label">上课老师</span>
                <a-select
                  :value="record.teacherId"
                  :options="teacherOptions"
                  class="schedule-conflict__control"
                  @change="value => updateRow(record.key, 'teacherId', String(value))"
                />
              </div>

              <div class="schedule-conflict__action-group">
                <div class="schedule-conflict__action-label-row">
                  <span class="schedule-conflict__action-label">节次时间</span>
                  <small class="schedule-conflict__action-label-hint">
                    {{ timeOptionAvailabilityLoading ? '正在检测空闲状态' : '先选老师，再看节次空闲情况' }}
                  </small>
                </div>
                <a-select
                  :value="timeValueForRow(record)"
                  option-label-prop="label"
                  popup-class-name="schedule-conflict__time-dropdown"
                  class="schedule-conflict__control"
                  @change="value => changeRowTime(record.key, String(value))"
                >
                  <a-select-option
                    v-for="item in timeOptionViewsForRow(record)"
                    :key="item.value"
                    :value="item.value"
                    :label="item.label"
                  >
                    <div class="schedule-conflict__time-option">
                      <span class="schedule-conflict__time-option-label">{{ item.baseLabel }}</span>
                      <span
                        class="schedule-conflict__time-option-status"
                        :class="{
                          'schedule-conflict__time-option-status--free': item.status === 'free',
                          'schedule-conflict__time-option-status--busy': item.status === 'busy',
                          'schedule-conflict__time-option-status--unknown': item.status === 'unknown',
                        }"
                      >
                        {{ item.statusText }}
                      </span>
                    </div>
                  </a-select-option>
                </a-select>
              </div>

              <div class="schedule-conflict__action-group">
                <div class="schedule-conflict__action-label-row">
                  <span class="schedule-conflict__action-label">上课助教</span>
                  <small class="schedule-conflict__action-label-hint">本行独立助教</small>
                </div>
                <a-select
                  :value="record.assistantIds"
                  mode="multiple"
                  placeholder="请选择上课助教"
                  show-search
                  allow-clear
                  option-label-prop="label"
                  option-filter-prop="label"
                  popup-class-name="schedule-conflict__assistant-dropdown"
                  class="schedule-conflict__control schedule-conflict__control--multiple"
                  @change="value => updateAssistantIds(record.key, (value || []) as Array<string | number>)"
                >
                  <a-select-option
                    v-for="item in assistantOptionViewsForRow(record)"
                    :key="item.value"
                    :value="item.value"
                    :label="item.label"
                  >
                    <div class="schedule-conflict__staff-option">
                      <span class="schedule-conflict__staff-option-label">{{ item.baseLabel }}</span>
                      <span
                        class="schedule-conflict__time-option-status"
                        :class="{
                          'schedule-conflict__time-option-status--free': item.status === 'free',
                          'schedule-conflict__time-option-status--busy': item.status === 'busy',
                          'schedule-conflict__time-option-status--unknown': item.status === 'unknown',
                        }"
                      >
                        {{ item.statusText }}
                      </span>
                    </div>
                  </a-select-option>
                </a-select>
              </div>

              <div class="schedule-conflict__action-group">
                <span class="schedule-conflict__action-label">上课教室</span>
                <a-select
                  :value="classroomSelectValue(record.classroomId)"
                  allow-clear
                  placeholder="请选择上课教室"
                  :options="classroomOptions"
                  class="schedule-conflict__control"
                  @change="value => updateRow(record.key, 'classroomId', String(value || ''))"
                />
              </div>

              <label
                v-if="record.displayHasStudentConflict"
                class="schedule-conflict__action-option"
              >
                <a-checkbox
                  :checked="record.allowStudentConflict"
                  @change="event => updateRow(record.key, 'allowStudentConflict', Boolean(event.target.checked))"
                />
                <div class="schedule-conflict__action-option-main">
                  <span>忽略学员冲突</span>
                  <small>允许学员并行上课，创建后标记冲突</small>
                </div>
              </label>

              <label
                v-if="record.displayHasClassroomConflict"
                class="schedule-conflict__action-option"
              >
                <a-checkbox
                  :checked="record.allowClassroomConflict"
                  @change="event => updateRow(record.key, 'allowClassroomConflict', Boolean(event.target.checked))"
                />
                <div class="schedule-conflict__action-option-main">
                  <span>忽略教室冲突</span>
                  <small>允许共享教室资源，创建后标记冲突</small>
                </div>
              </label>

              <div v-if="record.displayHasTeacherConflict" class="schedule-conflict__action-tip schedule-conflict__action-tip--danger">
                <span class="schedule-conflict__action-badge schedule-conflict__action-badge--danger">老师冲突</span>
                <span>请调整老师或节次后再创建</span>
              </div>
              <div v-if="record.displayHasAssistantConflict" class="schedule-conflict__action-tip schedule-conflict__action-tip--danger">
                <span class="schedule-conflict__action-badge schedule-conflict__action-badge--danger">助教冲突</span>
                <span>可直接调整上课助教后再创建</span>
              </div>
              <div
                v-else-if="!record.displayHasAssistantConflict && record.resolvedConflictTypes.includes('老师')"
                class="schedule-conflict__action-tip schedule-conflict__action-tip--success"
              >
                <span class="schedule-conflict__action-badge schedule-conflict__action-badge--success">已解决</span>
                <span>{{ record.localResolvedConflictTypes.includes('老师') ? '原老师冲突已解除，正在同步复校结果' : '原老师冲突已解除' }}</span>
              </div>

              <div
                class="schedule-conflict__action-result"
                :class="{
                  'schedule-conflict__action-result--ready': record.canCreate,
                  'schedule-conflict__action-result--resolved': !record.canCreate && record.resolvedConflictTypes.length,
                }"
              >
                <span class="schedule-conflict__action-result-label">处理结果</span>
                <strong>{{ record.canCreate ? '本节可创建' : '本节暂不可创建' }}</strong>
                <div v-if="record.resolvedConflictTypes.length" class="schedule-conflict__resolved-list">
                  <span
                    v-for="tag in record.resolvedConflictTypes"
                    :key="`${record.key}-resolved-${tag}`"
                    class="schedule-conflict__resolved-tag"
                  >
                    {{ resolvedConflictLabel(tag) }}
                  </span>
                </div>
                <span
                  v-if="record.localResolvedConflictTypes.length"
                  class="schedule-conflict__action-result-hint"
                >
                  当前已按行内调整先弱化原冲突，最终结果以重新校验为准。
                </span>
              </div>
            </div>
          </template>
        </template>
      </a-table>

      <div class="schedule-conflict__footer">
        <div class="schedule-conflict__footer-hint">
          学员冲突、教室冲突可在工作台内忽略；老师冲突请先改老师或改节次；助教冲突请返回创建弹窗修改助教。
        </div>
        <div class="schedule-conflict__footer-actions">
          <a-button @click="modalOpen = false">
            返回修改
          </a-button>
          <a-button
            type="primary"
            :loading="props.loading"
            :disabled="summary.ready === 0"
            @click="submitResolvedRows"
          >
            创建已处理项（{{ summary.ready }} 节）
          </a-button>
        </div>
      </div>
    </div>
  </a-modal>
</template>

<style scoped lang="less">
.schedule-conflict__titlebar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.schedule-conflict {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.schedule-conflict__banner {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  border-radius: 12px;
  background: #fff7f7;
  color: #ff7875;
  font-size: 13px;
  font-weight: 600;
  border: 1px solid #ffe1e0;
}

.schedule-conflict__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}

.schedule-conflict__toolbar-summary {
  color: #4b5563;
  font-size: 13px;
  line-height: 1.6;
}

.schedule-conflict__toolbar-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.schedule-conflict__toolbar-link {
  padding: 0;
  height: auto;
  font-weight: 600;
}

.schedule-conflict__toolbar-link--muted {
  color: #667085;
}

.schedule-conflict__toolbar-link--muted:hover {
  color: #475467;
}

.schedule-conflict__workbench {
  :deep(.ant-table-thead > tr > th) {
    padding: 12px 14px;
    color: #4b5563;
    font-size: 13px;
    font-weight: 700;
    background: #f8fafc;
  }

  :deep(.ant-table-thead > tr > th::before) {
    display: none;
  }

  :deep(.ant-table-tbody > tr > td) {
    padding: 12px 14px;
    vertical-align: top;
    background: #fff;
  }
}

.schedule-conflict__cell-card,
.schedule-conflict__conflict-item,
.schedule-conflict__action-panel {
  padding: 12px 14px;
  border: 1px solid #edf2f7;
  border-radius: 12px;
  background: #f8fafc;
}

.schedule-conflict__conflict-item + .schedule-conflict__conflict-item {
  margin-top: 8px;
}

.schedule-conflict__conflict-item--mixed {
  border-color: #d6e4ff;
  background: #fcfdff;
}

.schedule-conflict__conflict-item--resolved {
  border-color: #c6f6d5;
  background: #f6ffed;
}

.schedule-conflict__cell-stack {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.schedule-conflict__panel-title {
  color: #8c8c8c;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.02em;
}

.schedule-conflict__cell-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.schedule-conflict__group-index {
  color: #1f2329;
  font-size: 14px;
  font-weight: 700;
}

.schedule-conflict__group-time {
  color: #1677ff;
  font-size: 13px;
  font-weight: 600;
}

.schedule-conflict__cell-main {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-top: 8px;
  color: #8c8c8c;
  font-size: 12px;
}

.schedule-conflict__cell-main strong {
  color: #1f2329;
  font-size: 15px;
  font-weight: 700;
}

.schedule-conflict__cell-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  margin-top: 8px;
  color: #4b5563;
  font-size: 13px;
  line-height: 1.7;
}

.schedule-conflict__name-list {
  color: inherit;
}

.schedule-conflict__cell--danger {
  color: #ff4d4f;
  font-weight: 700;
}

.schedule-conflict__cell--success {
  color: #389e0d;
  font-weight: 700;
}

.schedule-conflict__tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 8px;
}

.schedule-conflict__status-tag {
  display: inline-flex;
  align-items: center;
  min-height: 24px;
  padding: 0 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
  line-height: 24px;
}

.schedule-conflict__status-tag--danger {
  background: #fff1f0;
  color: #ff4d4f;
}

.schedule-conflict__status-tag--success {
  background: #f6ffed;
  color: #389e0d;
}

.schedule-conflict__action-panel {
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-height: 100%;
}

.schedule-conflict__action-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.schedule-conflict__action-label {
  color: #8c8c8c;
  font-size: 12px;
  font-weight: 600;
}

.schedule-conflict__action-label-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.schedule-conflict__action-label-hint {
  color: #98a2b3;
  font-size: 11px;
  font-weight: 500;
  line-height: 1.4;
}

.schedule-conflict__control {
  width: 100%;
}

.schedule-conflict__control--multiple {
  :deep(.ant-select-selector) {
    align-items: center !important;
    min-height: 42px !important;
    row-gap: 4px;
  }

  :deep(.ant-select-selection-overflow) {
    flex-wrap: wrap !important;
    align-items: center !important;
  }

  :deep(.ant-select-selection-overflow-item) {
    max-width: none !important;
  }

  :deep(.ant-select-selection-item) {
    display: inline-flex;
    align-items: center;
    min-height: 26px;
    margin-top: 4px;
    margin-bottom: 4px;
    padding: 0 10px;
    border: 1px solid #d6e4ff;
    border-radius: 8px;
    background: #eef6ff;
    color: #1677ff;
    font-size: 12px;
    font-weight: 600;
  }

  :deep(.ant-select-selection-item-remove) {
    color: #6da7ff;
  }

  :deep(.ant-select-selection-item-remove:hover) {
    color: #1677ff;
  }
}

.schedule-conflict__action-group :deep(.ant-select-selector) {
  min-height: 42px !important;
  padding: 4px 12px !important;
  border: 1px solid #dfe7f1 !important;
  border-radius: 12px !important;
  background: #fff !important;
  box-shadow: none !important;
}

.schedule-conflict__action-group :deep(.ant-select-selection-item),
.schedule-conflict__action-group :deep(.ant-select-selection-placeholder) {
  line-height: 32px !important;
}

.schedule-conflict__action-group :deep(.ant-select-arrow) {
  color: #b3b8c2;
}

.schedule-conflict__time-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  width: 100%;
}

.schedule-conflict__staff-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  width: 100%;
}

.schedule-conflict__staff-option-label {
  min-width: 0;
  color: #1f2329;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.45;
}

.schedule-conflict__time-option-label {
  min-width: 0;
  color: #1f2329;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.4;
}

.schedule-conflict__time-option-status {
  display: inline-flex;
  flex-shrink: 0;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 600;
  line-height: 1;
}

.schedule-conflict__time-option-status::before {
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: currentColor;
  content: '';
}

.schedule-conflict__time-option-status--free {
  color: #52c41a;
}

.schedule-conflict__time-option-status--busy {
  color: #ff7a45;
}

.schedule-conflict__time-option-status--unknown {
  color: #98a2b3;
}

:deep(.schedule-conflict__time-dropdown) {
  .ant-select-item {
    padding: 0;
  }

  .ant-select-item-option {
    padding: 0 !important;
  }

  .ant-select-item-option-content {
    padding: 12px 16px;
  }

  .ant-select-item-option-active:not(.ant-select-item-option-disabled) .ant-select-item-option-content {
    background: #f8fbff;
  }

  .ant-select-item-option-selected:not(.ant-select-item-option-disabled) .ant-select-item-option-content {
    background: #edf5ff;
  }

  .ant-select-item-option-state {
    display: none;
  }

  .ant-select-item-option-selected .schedule-conflict__time-option-label {
    color: #1677ff;
  }

  .ant-select-item-option + .ant-select-item-option {
    border-top: 1px solid #f2f4f7;
  }
}

:deep(.schedule-conflict__assistant-dropdown) {
  .ant-select-item {
    padding: 0;
  }

  .ant-select-item-option {
    padding: 0 !important;
  }

  .ant-select-item-option-content {
    padding: 12px 16px;
  }

  .ant-select-item-option-active:not(.ant-select-item-option-disabled) .ant-select-item-option-content {
    background: #f8fbff;
  }

  .ant-select-item-option-selected:not(.ant-select-item-option-disabled) .ant-select-item-option-content {
    background: #edf5ff;
  }

  .ant-select-item-option-state {
    display: none;
  }

  .ant-select-item-option + .ant-select-item-option {
    border-top: 1px solid #f2f4f7;
  }
}

.schedule-conflict__action-option {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 10px;
  background: #fff;
  border: 1px solid #e9eef5;
  color: #4b5563;
  font-size: 13px;
}

.schedule-conflict__action-option-main {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.schedule-conflict__action-option-main span {
  color: #1f2329;
  font-weight: 600;
}

.schedule-conflict__action-option-main small {
  color: #8c8c8c;
  font-size: 12px;
  line-height: 1.5;
}

.schedule-conflict__action-tip {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border-radius: 10px;
  background: #fff;
  border: 1px solid #e9eef5;
  color: #8c8c8c;
  font-size: 13px;
  line-height: 1.6;
}

.schedule-conflict__action-tip--danger {
  color: #ff7875;
  font-weight: 600;
}

.schedule-conflict__action-tip--success {
  border-color: #d9f7be;
  background: #f6ffed;
  color: #389e0d;
  font-weight: 600;
}

.schedule-conflict__action-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 56px;
  height: 24px;
  padding: 0 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 700;
}

.schedule-conflict__action-badge--danger {
  background: #fff1f0;
  color: #ff4d4f;
}

.schedule-conflict__action-badge--success {
  background: #d9f7be;
  color: #237804;
}

.schedule-conflict__action-result {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-top: auto;
  padding: 10px 12px;
  border-radius: 10px;
  background: #fff;
  border: 1px dashed #d9e1ea;
  color: #8c8c8c;
  font-size: 13px;
}

.schedule-conflict__action-result-label {
  color: #8c8c8c;
  font-size: 12px;
  font-weight: 600;
}

.schedule-conflict__action-result strong {
  font-size: 14px;
  font-weight: 700;
}

.schedule-conflict__action-result--ready {
  border-color: #b7d9ff;
  background: #f3f9ff;
  color: #1677ff;
}

.schedule-conflict__action-result--resolved {
  border-color: #d9f7be;
  background: #fcfff8;
}

.schedule-conflict__resolved-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 4px;
}

.schedule-conflict__resolved-tag {
  display: inline-flex;
  align-items: center;
  min-height: 26px;
  padding: 0 10px;
  border-radius: 999px;
  background: #f6ffed;
  border: 1px solid #b7eb8f;
  color: #237804;
  font-size: 12px;
  font-weight: 700;
  line-height: 24px;
}

.schedule-conflict__action-result-hint {
  color: #8c8c8c;
  font-size: 12px;
  line-height: 1.5;
}

.schedule-conflict__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding-top: 8px;
}

.schedule-conflict__footer-hint {
  color: #8c8c8c;
  font-size: 13px;
  line-height: 1.6;
}

.schedule-conflict__footer-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-left: auto;
}

@media (max-width: 1200px) {
  .schedule-conflict__cell-top,
  .schedule-conflict__cell-meta {
    align-items: flex-start;
    flex-direction: column;
    gap: 8px;
  }

  .schedule-conflict__footer {
    align-items: stretch;
    flex-direction: column;
  }

  .schedule-conflict__footer-actions {
    width: 100%;
    justify-content: flex-end;
  }
}
</style>

<style lang="less">
.schedule-conflict__assistant-dropdown.ant-select-dropdown,
.schedule-conflict__time-dropdown.ant-select-dropdown {
  .ant-select-item-option-state {
    display: none !important;
  }
}

.schedule-conflict__assistant-dropdown.ant-select-dropdown {
  .ant-select-item {
    padding: 0;
  }

  .ant-select-item-option {
    min-height: auto;
    padding: 0 !important;
  }

  .ant-select-item-option-content {
    padding: 12px 16px;
  }

  .ant-select-item-option-active:not(.ant-select-item-option-disabled) .ant-select-item-option-content {
    background: #f8fbff;
  }

  .ant-select-item-option-selected:not(.ant-select-item-option-disabled) .ant-select-item-option-content {
    background: #edf5ff;
  }

  .ant-select-item-option + .ant-select-item-option {
    border-top: 1px solid #f2f4f7;
  }
}
</style>
