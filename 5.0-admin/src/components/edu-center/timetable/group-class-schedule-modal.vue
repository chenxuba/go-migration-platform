<script setup lang="ts">
import {
  CalendarOutlined,
  ClockCircleOutlined,
  CloseOutlined,
  EnvironmentOutlined,
  QuestionCircleOutlined,
  TeamOutlined,
  UserOutlined,
} from '@ant-design/icons-vue'
import dayjs, { type Dayjs } from 'dayjs'
import { computed, nextTick, ref, watch } from 'vue'
import type { GroupClassBatchPlanModalPreset } from './group-class-batch-plan-preset'
import GroupClassScheduleConflictWorkbenchModal from './group-class-schedule-conflict-workbench-modal.vue'
import { type ClassroomItem, listClassroomsApi } from '@/api/business-settings/classroom'
import { getInstPeriodConfigApi } from '@/api/common/config'
import { type GroupClassDetailVO, type GroupClassRow, listGroupClassStudentsByClassIdsApi, pageGroupClassesApi, getGroupClassDetailApi } from '@/api/edu-center/group-class'
import type { TeachingScheduleValidationResult } from '@/api/edu-center/teaching-schedule'
import { checkGroupClassAssistantScheduleAvailabilityApi, createGroupClassSchedulesApi, replaceTeachingScheduleBatchApi, validateGroupClassSchedulesApi } from '@/api/edu-center/teaching-schedule'
import { getUserListApi } from '@/api/internal-manage/staff-manage'
import StaffSelect from '@/components/common/staff-select.vue'
import { useUserStore } from '@/stores/user'
import emitter, { EVENTS } from '@/utils/eventBus'
import messageService from '@/utils/messageService'
import {
  DEFAULT_UNIFIED_TIME_PERIOD_CONFIG,
  buildQuickHourlySlots,
  configGroupsSorted,
  periodGroupIndexForKey as resolvePeriodGroupIndex,
  periodGroupKeyForIndex,
  parseUnifiedTimePeriodConfig,
} from '@/utils/unified-time-period'

type PreviewTone = 'pending' | 'blocked'
type SchedulingMode = 'repeat' | 'free'
type RepeatRule = 'none' | 'weekly' | 'biweekly' | 'daily' | 'alternateDay'
type HolidayPolicy = 'include' | 'filter'
type PeriodGroupKey = string

interface PreviewItem {
  date: string
  week: string
  rule: string
  time: string
  startTime: string
  endTime: string
  teacher: string
  assistant: string
  classroom: string
  teacherId?: string
  assistantIds?: string[]
  classroomId?: string
  allowStudentConflict?: boolean
  tone: PreviewTone
}

type BatchCreatePlan = Omit<PreviewItem, 'tone'> & {
  tone?: PreviewTone
}

interface SummaryItem {
  label: string
  value: string
}

interface OptionItem<T = string> {
  value: T
  label: string
  desc?: string
}

interface SchoolTimeSlot {
  value: string
  label: string
  desc: string
  start: string
  end: string
}

interface TimeBlock {
  key: string
  rangeText: string
  startTime: string
  endTime: string
  minutes: number
}

interface StaffOptionItem {
  id: string | number
  name?: string
  nickName?: string
  mobile?: string
  status?: number
  disabled?: boolean
}

interface AvailabilityBadgeView {
  status: 'free' | 'busy' | 'unknown'
  statusText: string
}

interface SlotSelectOptionView {
  value: string
  label: string
  baseLabel: string
  status: 'free' | 'busy' | 'unknown'
  statusText: string
}

interface AssistantSelectOptionView {
  value: string
  label: string
  baseLabel: string
  mobile?: string
  status: 'free' | 'busy' | 'unknown'
  statusText: string
}

interface GroupClassRecord {
  id: string
  name: string
  courseId: string
  courseName: string
  status: number
  defaultTeacherId: string
  defaultTeacherName: string
  teacherIds: string[]
  teacherNames: string[]
  classroomId: string
  classroomName: string
  studentCount: number
  studentNames: string[]
  remark: string
  detailLoaded?: boolean
}

const props = withDefaults(defineProps<{
  open: boolean
  mode?: 'create' | 'editBatch'
  batchPlanPreset?: GroupClassBatchPlanModalPreset | null
}>(), {
  mode: 'create',
  batchPlanPreset: null,
})

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'updated'): void
}>()

const modalOpen = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const isBatchPlanEditMode = computed(() => props.mode === 'editBatch')
const isSingleScheduleEditMode = computed(() => {
  if (!isBatchPlanEditMode.value || !props.batchPlanPreset)
    return false
  if (props.batchPlanPreset.editScope === 'current')
    return true
  const batchNo = String(props.batchPlanPreset.batchNo || '').trim()
  const scheduleIds = Array.isArray(props.batchPlanPreset.scheduleIds)
    ? props.batchPlanPreset.scheduleIds.map(id => String(id || '').trim()).filter(Boolean)
    : []
  return !batchNo && scheduleIds.length === 1
})
const showSchedulingModeSection = computed(() => !isBatchPlanEditMode.value)

const weekDayOptions = ['周一', '周二', '周三', '周四', '周五', '周六', '周日']
const weekdayToNumber: Record<string, number> = {
  周日: 0,
  周一: 1,
  周二: 2,
  周三: 3,
  周四: 4,
  周五: 5,
  周六: 6,
}
const weekDisplayMap: Record<number, string> = {
  0: '周日',
  1: '周一',
  2: '周二',
  3: '周三',
  4: '周四',
  5: '周五',
  6: '周六',
}

const schedulingModeOptions: OptionItem<SchedulingMode>[] = [
  { value: 'repeat', label: '重复排课', desc: '批量生成固定周期班课日程' },
  { value: 'free', label: '自由排课', desc: '先创建单次班课，后续再补排' },
]

const repeatRuleOptions: OptionItem<RepeatRule>[] = [
  { value: 'none', label: '不重复' },
  { value: 'weekly', label: '每周重复' },
  { value: 'biweekly', label: '隔周重复' },
  { value: 'daily', label: '每天重复' },
  { value: 'alternateDay', label: '隔天重复' },
]

const holidayPolicyOptions: OptionItem<HolidayPolicy>[] = [
  { value: 'include', label: '不过滤节假日' },
  { value: 'filter', label: '过滤节假日' },
]

const repeatRuleLabelMap: Record<RepeatRule, string> = {
  none: '不重复',
  weekly: '每周重复',
  biweekly: '隔周重复',
  daily: '每天重复',
  alternateDay: '隔天重复',
}

const schoolHolidaySet = new Set(['2026-05-01', '2026-05-02', '2026-05-03'])

const userStore = useUserStore()
const effectivePeriodConfigRaw = ref<unknown>(null)
const plannerShellRef = ref<HTMLElement | null>(null)

const periodConfig = computed(() => {
  const parsed = parseUnifiedTimePeriodConfig(effectivePeriodConfigRaw.value ?? userStore.instConfig?.unifiedTimePeriodJson)
  return parsed ?? DEFAULT_UNIFIED_TIME_PERIOD_CONFIG
})

const sortedPeriodGroups = computed(() => configGroupsSorted(periodConfig.value))

const groupOptions = computed(() => {
  const groups = sortedPeriodGroups.value
  if (!groups.length)
    return [{ key: 'A', label: '默认时段' }]
  return groups.map((group, index) => ({
    key: periodGroupKeyForIndex(index),
    label: group.name || `${periodGroupKeyForIndex(index)}时段`,
  }))
})

function slotsForGroupKey(key: PeriodGroupKey) {
  const groups = sortedPeriodGroups.value
  const fallback = buildQuickHourlySlots().filter(item => item.enabled !== false)
  if (!groups.length)
    return [...fallback].sort((a, b) => a.index - b.index)
  const idx = periodGroupIndexForKey(key)
  const current = groups[idx] || groups[0]
  return [...current.slots].filter(item => item.enabled !== false).sort((a, b) => a.index - b.index)
}

function periodGroupIndexForKey(key: PeriodGroupKey) {
  return resolvePeriodGroupIndex(key)
}

async function loadEffectivePeriodConfig(dateText?: string) {
  try {
    const res = await getInstPeriodConfigApi({
      effectiveDate: dateText || scheduleStartDate.value.format('YYYY-MM-DD'),
    })
    effectivePeriodConfigRaw.value = res.result?.unifiedTimePeriodJson ?? userStore.instConfig?.unifiedTimePeriodJson ?? null
  }
  catch (error) {
    console.warn('load effective period config for group class failed', error)
    if (!userStore.instConfig)
      await userStore.getInstConfig()
    effectivePeriodConfigRaw.value = userStore.instConfig?.unifiedTimePeriodJson ?? null
  }
}

function isTeacherAllowedInPeriodGroup(teacherIdStr: string, groupIndex: number): boolean {
  const groups = sortedPeriodGroups.value
  const current = groups[groupIndex]
  if (!current)
    return true
  const list = current.boundTeachers
  if (!Array.isArray(list) || list.length === 0)
    return true
  if (!teacherIdStr)
    return true
  return list.some(item => String(item.id ?? '').trim() === teacherIdStr)
}

function isValidStaffId(value: unknown) {
  const text = String(value ?? '').trim()
  return text !== '' && text !== '0' && text !== 'undefined' && text !== 'null'
}

function isValidClassroomId(value: unknown) {
  const text = String(value ?? '').trim()
  return text !== '' && text !== '0' && text !== 'undefined' && text !== 'null'
}

function sameStaffId(a: unknown, b: unknown) {
  return isValidStaffId(a) && isValidStaffId(b) && String(a) === String(b)
}

function displayMobileText(staff?: StaffOptionItem | null) {
  return String(staff?.mobile || '').trim()
}

function displayStaffName(staff?: StaffOptionItem | null) {
  const base = staff?.nickName || staff?.name || ''
  if (!base)
    return ''
  return staff?.disabled && !base.endsWith('（离职）') ? `${base}（离职）` : base
}

function scrollPlannerShellToTop() {
  const el = plannerShellRef.value
  if (el)
    el.scrollTop = 0
}

function scheduleSlotSelectGetPopupContainer() {
  return document.body
}

function scheduleSlotKeysEqual(a: string[], b: string[]) {
  if (a.length !== b.length)
    return false
  for (let i = 0; i < a.length; i++) {
    if (a[i] !== b[i])
      return false
  }
  return true
}

function findPresetGroupAndSlotKeys(timeBlocks: GroupClassBatchPlanModalPreset['timeBlocks']) {
  const normalized = Array.isArray(timeBlocks)
    ? timeBlocks.map(item => ({
        startTime: String(item?.startTime || '').trim().slice(0, 5),
        endTime: String(item?.endTime || '').trim().slice(0, 5),
      })).filter(item => item.startTime && item.endTime)
    : []
  if (!normalized.length)
    return { group: currentGroup.value, slotKeys: [] as string[] }

  for (const opt of groupOptions.value) {
    const key = opt.key as PeriodGroupKey
    const options = slotsForGroupKey(key)
    const slotKeys: string[] = []
    let matched = true
    for (const block of normalized) {
      const slot = options.find(item => String(item.start || '').slice(0, 5) === block.startTime && String(item.end || '').slice(0, 5) === block.endTime)
      if (!slot) {
        matched = false
        break
      }
      slotKeys.push(`period-${slot.index}`)
    }
    if (matched)
      return { group: key, slotKeys }
  }

  return { group: currentGroup.value, slotKeys: [] as string[] }
}

function scheduleAvailabilityBadge(status: 'free' | 'busy' | 'unknown', statusText: string): AvailabilityBadgeView {
  return { status, statusText }
}

function slotDurationMinutes(start: string, end: string) {
  const [sh, sm] = start.split(':').map(Number)
  const [eh, em] = end.split(':').map(Number)
  return Math.max(0, eh * 60 + em - (sh * 60 + sm))
}

function slotBaseLabel(slot: SchoolTimeSlot) {
  return `${slot.desc} · ${slot.start} - ${slot.end}`
}

function normalizeGroupClassRecord(value: Partial<GroupClassRecord> & { id: string, name: string }): GroupClassRecord {
  return {
    id: String(value.id || '').trim(),
    name: String(value.name || '').trim(),
    courseId: String(value.courseId || '').trim(),
    courseName: String(value.courseName || '').trim(),
    status: Number(value.status || 0),
    defaultTeacherId: String(value.defaultTeacherId || '').trim(),
    defaultTeacherName: String(value.defaultTeacherName || '').trim(),
    teacherIds: Array.isArray(value.teacherIds) ? value.teacherIds.map(item => String(item || '').trim()).filter(Boolean) : [],
    teacherNames: Array.isArray(value.teacherNames) ? value.teacherNames.map(item => String(item || '').trim()).filter(Boolean) : [],
    classroomId: String(value.classroomId || '').trim(),
    classroomName: String(value.classroomName || '').trim(),
    studentCount: Math.max(0, Number(value.studentCount || 0)),
    studentNames: Array.isArray(value.studentNames) ? value.studentNames.map(item => String(item || '').trim()).filter(Boolean) : [],
    remark: String(value.remark || '').trim(),
    detailLoaded: value.detailLoaded === true,
  }
}

function mapGroupClassRow(item: GroupClassRow): GroupClassRecord {
  const teachers = Array.isArray(item?.teachers) ? item.teachers : []
  return normalizeGroupClassRecord({
    id: String(item?.id ?? ''),
    name: String(item?.name || item?.id || '').trim(),
    courseId: String(item?.lessonId ?? ''),
    courseName: String(item?.lessonName || '').trim(),
    status: Number(item?.status || 0),
    defaultTeacherId: String(item?.defaultTeacherId ?? ''),
    defaultTeacherName: String(item?.defaultTeacherName || '').trim(),
    teacherIds: teachers.map(teacher => teacher?.id),
    teacherNames: teachers.map(teacher => teacher?.name),
    classroomName: String(item?.classRoomName || '').trim(),
    studentCount: Number(item?.studentCount || 0),
    remark: String(item?.remark || '').trim(),
  })
}

function mapGroupClassDetail(detail: GroupClassDetailVO): Partial<GroupClassRecord> {
  const teachers = Array.isArray(detail?.teachers) ? detail.teachers : []
  return {
    id: String(detail?.id ?? ''),
    name: String(detail?.name || detail?.id || '').trim(),
    courseId: String(detail?.lessonId ?? ''),
    courseName: String(detail?.lessonName || '').trim(),
    status: Number(detail?.status || 0),
    defaultTeacherId: String(detail?.defaultTeacherId ?? ''),
    defaultTeacherName: String(detail?.defaultTeacherName || '').trim(),
    teacherIds: teachers.map(teacher => teacher?.id),
    teacherNames: teachers.map(teacher => teacher?.name),
    classroomId: String(detail?.classroomId ?? ''),
    classroomName: String(detail?.classroomName || detail?.classRoomName || '').trim(),
    studentCount: Number(detail?.studentCount || 0),
    remark: String(detail?.remark || '').trim(),
    detailLoaded: true,
  }
}

function normalizeStudentBucketNames(list?: Array<{ name?: string }>) {
  return Array.isArray(list)
    ? list.map(item => String(item?.name || '').trim()).filter(Boolean)
    : []
}

const classroomList = ref<ClassroomItem[]>([])
const workbenchTeacherList = ref<StaffOptionItem[]>([])
const assistantWorkbenchStaffList = ref<StaffOptionItem[]>([])
const groupClassRecords = ref<GroupClassRecord[]>([])
const groupClassLoading = ref(false)
const groupClassDetailLoading = ref(false)
const selectedGroupClassId = ref<string | undefined>(undefined)
const schedulingMode = ref<SchedulingMode>('repeat')
const repeatRule = ref<RepeatRule>('weekly')
const holidayPolicy = ref<HolidayPolicy>('filter')
const currentGroup = ref<PeriodGroupKey>('A')
const selectedWeekdays = ref(['周一', '周三', '周五'])
const selectedSchoolTimeSlots = ref<string[]>([])
const selectedTeacher = ref<string | number | undefined>(undefined)
const selectedTeacherDisplay = ref<StaffOptionItem | null>(null)
const selectedAssistant = ref<Array<string | number> | undefined>(undefined)
const selectedAssistantDisplays = ref<StaffOptionItem[]>([])
const selectedClassroom = ref<string | undefined>(undefined)
const previewModalOpen = ref(false)
const creatingSchedules = ref(false)
const creatingWithSoftConflict = ref(false)
const previewValidating = ref(false)
const previewValidationMessage = ref('')
const previewHasConflict = ref(false)
const previewValidationResult = ref<TeachingScheduleValidationResult | null>(null)
const conflictModalOpen = ref(false)
const scheduleStartDate = ref(dayjs().startOf('day'))
const freeSelectedDates = ref<Dayjs[]>([dayjs().startOf('day')])
const freeCalendarPanelDate = ref(dayjs().startOf('month'))
const freeCalendarOpen = ref(false)
const plannedClassCount = ref(1)
const slotAvailabilityMap = ref<Record<string, AvailabilityBadgeView>>({})
const slotAvailabilityLoading = ref(false)
const assistantAvailabilityMap = ref<Record<string, AvailabilityBadgeView>>({})
const assistantAvailabilityLoading = ref(false)

let selectedGroupClassSeq = 0
let slotAvailabilitySeq = 0
let slotAvailabilityTimer: ReturnType<typeof setTimeout> | null = null
let assistantAvailabilitySeq = 0
let assistantAvailabilityTimer: ReturnType<typeof setTimeout> | null = null
let applyingBatchPlanPreset = false

const selectedGroupClass = computed(() =>
  groupClassRecords.value.find(item => item.id === selectedGroupClassId.value),
)

const selectedAssistantValues = computed<Array<string | number>>(() =>
  Array.isArray(selectedAssistant.value) ? selectedAssistant.value : [],
)

const batchPresetSchedules = computed(() =>
  Array.isArray(props.batchPlanPreset?.detail?.schedules)
    ? props.batchPlanPreset.detail.schedules
    : [],
)

const schoolTimeSlotOptions = computed<SchoolTimeSlot[]>(() => {
  const slots = slotsForGroupKey(currentGroup.value)
  return slots.map(item => ({
    value: `period-${item.index}`,
    label: `第${item.index}节课`,
    desc: `第${item.index}节课`,
    start: String(item.start || '').slice(0, 5),
    end: String(item.end || '').slice(0, 5),
  }))
})

const selectedTeacherIdNormalized = computed(() =>
  isValidStaffId(selectedTeacher.value) ? String(selectedTeacher.value).trim() : '',
)

const eligiblePeriodGroupKeys = computed<PeriodGroupKey[]>(() => {
  const options = groupOptions.value
  const teacherId = selectedTeacherIdNormalized.value
  const keys: PeriodGroupKey[] = []
  for (const option of options) {
    const key = option.key as PeriodGroupKey
    if (isTeacherAllowedInPeriodGroup(teacherId, periodGroupIndexForKey(key)))
      keys.push(key)
  }
  return keys
})

function isPeriodGroupChoiceDisabled(key: PeriodGroupKey): boolean {
  const eligible = eligiblePeriodGroupKeys.value
  if (!eligible.length)
    return false
  return !eligible.includes(key)
}

function isAssistantAllowedInCurrentGroup(assistantIdStr: string) {
  return isTeacherAllowedInPeriodGroup(assistantIdStr, periodGroupIndexForKey(currentGroup.value))
}

function isPresetAssistantId(id: unknown) {
  if (!isBatchPlanEditMode.value)
    return false
  const target = String(id || '').trim()
  if (!target)
    return false
  return Array.isArray(props.batchPlanPreset?.assistantIds)
    && props.batchPlanPreset.assistantIds.some(item => String(item || '').trim() === target)
}

const activeTimeBlocks = computed<TimeBlock[]>(() => {
  const rows: (TimeBlock & { sortKey: string })[] = []
  for (const id of selectedSchoolTimeSlots.value) {
    const slot = schoolTimeSlotOptions.value.find(item => item.value === id)
    if (!slot)
      continue
    rows.push({
      key: slot.value,
      rangeText: `${slot.desc} · ${slot.start}-${slot.end}`,
      startTime: slot.start,
      endTime: slot.end,
      minutes: slotDurationMinutes(slot.start, slot.end),
      sortKey: slot.start,
    })
  }
  return rows.sort((a, b) => a.sortKey.localeCompare(b.sortKey)).map(({ sortKey: _sortKey, ...rest }) => rest)
})

const groupClassSelectOptions = computed(() =>
  groupClassRecords.value.map(item => ({
    value: item.id,
    label: `${item.name || '-'} · 课程：${item.courseName || '-'}`,
    statusText: Number(item.status || 0) === 2 ? '已结班' : '开班中',
    studentCount: item.studentNames.length || item.studentCount || 0,
  })),
)

const teacherPresetStaff = computed<StaffOptionItem[]>(() => {
  const records: StaffOptionItem[] = []
  const added = new Set<string>()
  const append = (id: unknown, name?: string) => {
    if (!isValidStaffId(id))
      return
    const label = String(name || '').trim()
    if (!label)
      return
    const key = String(id)
    if (added.has(key))
      return
    added.add(key)
    records.push({ id: key, name: label, nickName: label })
  }
  batchPresetSchedules.value.forEach(item => append(item.teacherId, item.teacherName))
  const current = selectedGroupClass.value
  current?.teacherIds.forEach((id, index) => append(id, current.teacherNames[index]))
  append(current?.defaultTeacherId, current?.defaultTeacherName)
  append(selectedTeacherDisplay.value?.id, displayStaffName(selectedTeacherDisplay.value))
  return records
})

const assistantPresetStaff = computed<StaffOptionItem[]>(() => {
  const records: StaffOptionItem[] = []
  const added = new Set<string>()
  const append = (id: unknown, name?: string) => {
    if (!isValidStaffId(id))
      return
    const label = String(name || '').trim()
    if (!label)
      return
    const key = String(id).trim()
    if (added.has(key))
      return
    added.add(key)
    records.push({ id: key, name: label, nickName: label })
  }
  batchPresetSchedules.value.forEach((item) => {
    const ids = Array.isArray(item.assistantIds) ? item.assistantIds : []
    const names = Array.isArray(item.assistantNames) ? item.assistantNames : []
    ids.forEach((id, index) => append(id, names[index]))
  })
  return records
})

function resolveStaffDisplayById(id: string | number | undefined) {
  if (!isValidStaffId(id))
    return null
  const target = String(id)
  return workbenchTeacherList.value.find(item => sameStaffId(item.id, target))
    || teacherPresetStaff.value.find(item => sameStaffId(item.id, target))
    || selectedAssistantDisplays.value.find(item => sameStaffId(item.id, target))
    || null
}

const scheduleStaffSelectKey = computed(() => selectedGroupClass.value?.id || 'empty')

const teacherSelectOptions = computed(() => {
  const merged = new Map<string, { value: string, label: string }>()
  const append = (item?: StaffOptionItem | null) => {
    if (!isValidStaffId(item?.id))
      return
    const value = String(item.id).trim()
    const label = displayStaffName(item) || value
    if (!merged.has(value))
      merged.set(value, { value, label })
  }
  teacherPresetStaff.value.forEach(item => append(item))
  workbenchTeacherList.value.forEach(item => append(item))
  return [...merged.values()]
})

const assistantSelectStaffs = computed<StaffOptionItem[]>(() => {
  const merged = new Map<string, StaffOptionItem>()
  const append = (item?: StaffOptionItem | null, force = false) => {
    if (!isValidStaffId(item?.id))
      return
    const id = String(item.id).trim()
    const existing = merged.get(id)
    if (!force && !existing && !isAssistantAllowedInCurrentGroup(id))
      return
    const displayName = displayStaffName(item) || id
    const mobile = displayMobileText(item)
    if (!existing) {
      merged.set(id, {
        id,
        name: displayName,
        nickName: displayName,
        mobile,
        status: item?.status,
        disabled: item?.disabled,
      })
      return
    }
    merged.set(id, {
      ...existing,
      name: existing.name || displayName,
      nickName: existing.nickName || displayName,
      mobile: existing.mobile || mobile,
      status: item?.status ?? existing.status,
      disabled: item?.disabled ?? existing.disabled,
    })
  }
  selectedAssistantDisplays.value.forEach(item => append(item, true))
  assistantPresetStaff.value.forEach(item => append(item, isPresetAssistantId(item.id)))
  assistantWorkbenchStaffList.value.forEach(item => append(item))
  return [...merged.values()]
})

const assistantSelectOptionViews = computed<AssistantSelectOptionView[]>(() =>
  assistantSelectStaffs.value
    .filter(staff => !sameStaffId(staff.id, selectedTeacher.value))
    .map((staff) => {
      const availability = assistantAvailabilityFor(staff)
      const baseLabel = displayStaffName(staff) || String(staff.id)
      return {
        value: String(staff.id),
        label: `${baseLabel} · ${availability.statusText}`,
        baseLabel,
        mobile: displayMobileText(staff),
        status: availability.status,
        statusText: availability.statusText,
      }
    }),
)

function assistantAvailabilityFor(staff: StaffOptionItem): AvailabilityBadgeView {
  if (!activeTimeBlocks.value.length)
    return scheduleAvailabilityBadge('unknown', '先选时段')
  return assistantAvailabilityMap.value[String(staff.id)]
    || scheduleAvailabilityBadge(assistantAvailabilityLoading.value ? 'unknown' : 'free', assistantAvailabilityLoading.value ? '检测中' : '空闲')
}

const assistantOptionList = computed(() =>
  assistantSelectStaffs.value
    .filter(item => !sameStaffId(item.id, selectedTeacher.value))
    .map(item => ({
      value: String(item.id),
      label: displayStaffName(item) || String(item.id),
      mobile: displayMobileText(item),
    })),
)

const classroomOptions = computed(() => {
  const classroomSet = new Map<string, { value: string, label: string }>()
  const append = (id?: string | number, name?: string) => {
    if (!isValidClassroomId(id))
      return
    const key = String(id).trim()
    const label = String(name || '').trim()
    if (!label || classroomSet.has(key))
      return
    classroomSet.set(key, { value: key, label })
  }
  classroomList.value.forEach(item => append(item.id, item.name))
  batchPresetSchedules.value.forEach(item => append(item.classroomId, item.classroomName))
  append(selectedGroupClass.value?.classroomId, selectedGroupClass.value?.classroomName)
  return [...classroomSet.values()]
})

const normalizedSelectedClassroomId = computed(() => {
  const current = String(selectedClassroom.value || '').trim()
  if (!current)
    return ''
  return classroomOptions.value.some(item => item.value === current) ? current : ''
})

const selectedTeacherText = computed(() => {
  if (!isValidStaffId(selectedTeacher.value))
    return '-'
  const current = selectedTeacherDisplay.value
  if (displayStaffName(current))
    return displayStaffName(current)
  const preset = teacherPresetStaff.value.find(item => sameStaffId(item.id, selectedTeacher.value))
  if (displayStaffName(preset))
    return displayStaffName(preset)
  return String(selectedTeacher.value)
})

const selectedAssistantText = computed(() =>
  selectedAssistantValues.value.length
    ? selectedAssistantValues.value.map((id) => {
        const current = selectedAssistantDisplays.value.find(item => sameStaffId(item.id, id))
        if (displayStaffName(current))
          return displayStaffName(current)
        return String(id)
      }).filter(Boolean).join('、')
    : '未安排',
)

const scheduledClassroomText = computed(() => {
  const current = classroomOptions.value.find(item => item.value === normalizedSelectedClassroomId.value)
  if (current?.label)
    return current.label
  return selectedGroupClass.value?.classroomName || '未设置教室'
})

const groupClassStatusText = computed(() =>
  Number(selectedGroupClass.value?.status) === 2 ? '已结班' : '开班中',
)

const selectedGroupClassTeacherText = computed(() => {
  const names = selectedGroupClass.value?.teacherNames || []
  return names.length ? names.join('、') : '-'
})

const selectedGroupClassStudentText = computed(() => {
  const names = selectedGroupClass.value?.studentNames || []
  return names.length ? names.join('、') : '暂无在班学员'
})

const selectedGroupClassStudentCountText = computed(() => {
  const current = selectedGroupClass.value
  const count = current?.studentNames.length || current?.studentCount || 0
  return `${count} 人`
})

const scheduleSessionMinutesText = computed(() => {
  const blocks = activeTimeBlocks.value
  if (!blocks.length)
    return '--'
  const uniq = [...new Set(blocks.map(item => `${item.minutes} 分钟`))]
  return uniq.join('、')
})

const selectedGroupClassSummary = computed<SummaryItem[]>(() => [
  { label: '课程', value: selectedGroupClass.value?.courseName || '-' },
  { label: '默认老师', value: selectedGroupClass.value?.defaultTeacherName || '-' },
  { label: '班级老师', value: selectedGroupClassTeacherText.value },
  { label: '默认教室', value: selectedGroupClass.value?.classroomName || '未设置教室' },
  { label: '在班学员', value: selectedGroupClassStudentCountText.value },
  { label: '本次时长', value: scheduleSessionMinutesText.value },
])

const selectedWeekdaysText = computed(() => selectedWeekdays.value.join(' / '))

const schoolSlotFieldLabelText = computed(() => isSingleScheduleEditMode.value ? '课表节次' : '课表节次（可多选）')
const schoolSlotTooltipText = computed(() =>
  isSingleScheduleEditMode.value
    ? '单个班课日程编辑时仅允许选择一个课表节次，保存后会更新当前这节班课。'
    : '同一上课日内按所选「第几节课」各生成一节班课日程；重复排课时每个上课日都会生成这些节。',
)
const schoolSlotPlaceholderText = computed(() =>
  isSingleScheduleEditMode.value ? '请选择一个课表节次' : '同一天内可勾选多节，例如上午 + 下午',
)
const schoolTimeSlotSelectValue = computed<string | string[] | undefined>({
  get() {
    if (isSingleScheduleEditMode.value)
      return selectedSchoolTimeSlots.value[0]
    return selectedSchoolTimeSlots.value
  },
  set(value) {
    if (isSingleScheduleEditMode.value) {
      const normalized = String(value || '').trim()
      selectedSchoolTimeSlots.value = normalized ? [normalized] : []
      return
    }
    selectedSchoolTimeSlots.value = Array.isArray(value)
      ? value.map(item => String(item || '').trim()).filter(Boolean)
      : []
  },
})

const freeSelectedDatesSorted = computed(() =>
  [...freeSelectedDates.value].sort((a, b) => a.valueOf() - b.valueOf()),
)

const freeSelectedDateKeys = computed(() =>
  new Set(freeSelectedDatesSorted.value.map(item => item.format('YYYY-MM-DD'))),
)

const freeSelectedDatesText = computed(() => {
  const dates = freeSelectedDatesSorted.value.map(item => item.format('YYYY-MM-DD'))
  if (!dates.length)
    return '请选择上课日期'
  return dates.join('、')
})

const scheduleTargetCount = computed(() => {
  const slots = Math.max(activeTimeBlocks.value.length, 1)
  if (schedulingMode.value === 'free')
    return freeSelectedDatesSorted.value.length * slots
  const planned = Math.max(0, Math.floor(Number(plannedClassCount.value) || 0))
  if (planned < 1)
    return 0
  if (repeatRule.value === 'none')
    return Math.min(planned, slots)
  return planned
})

const repeatRuleText = computed(() => {
  if (schedulingMode.value === 'free')
    return '自由排课 · 单次日程'
  const base = repeatRuleLabelMap[repeatRule.value]
  if (repeatRule.value === 'weekly' || repeatRule.value === 'biweekly')
    return `${base} · ${selectedWeekdaysText.value || '-'}`
  return base
})

function slotAvailabilityFor(slot: SchoolTimeSlot): AvailabilityBadgeView {
  if (!selectedGroupClass.value?.id)
    return scheduleAvailabilityBadge('unknown', '先选班级')
  if (!isValidStaffId(selectedTeacher.value))
    return scheduleAvailabilityBadge('unknown', '先选老师')
  if (!plannedDates.value.length)
    return scheduleAvailabilityBadge('unknown', '待定')
  return slotAvailabilityMap.value[slot.value]
    || scheduleAvailabilityBadge(slotAvailabilityLoading.value ? 'unknown' : 'free', slotAvailabilityLoading.value ? '检测中' : '可排')
}

const schoolTimeSlotOptionViews = computed<SlotSelectOptionView[]>(() =>
  schoolTimeSlotOptions.value.map((slot) => {
    const availability = slotAvailabilityFor(slot)
    return {
      value: slot.value,
      label: `${slotBaseLabel(slot)} · ${availability.statusText}`,
      baseLabel: slotBaseLabel(slot),
      status: availability.status,
      statusText: availability.statusText,
    }
  }),
)

function selectedTimeBlockStatusText(block: TimeBlock) {
  const matched = schoolTimeSlotOptionViews.value.find(item => item.value === block.key)
  return matched?.statusText || '待定'
}

const timeModeText = computed(() => {
  const blocks = activeTimeBlocks.value
  const groupLabel = groupOptions.value.find(item => item.key === currentGroup.value)?.label || ''
  const prefix = groupLabel ? `${groupLabel} · ` : ''
  if (!blocks.length)
    return `${prefix}请选择课表节次`
  const blocksDesc = blocks.map((block) => {
    const statusText = selectedTimeBlockStatusText(block)
    return `${block.rangeText} · ${statusText}`
  }).join('；')
  const count = blocks.length
  const head = count > 1 ? `课表节次（共 ${count} 节）` : '课表节次'
  return `${prefix}${head} · ${blocksDesc}`
})

const overviewItems = computed<SummaryItem[]>(() => [
  { label: '排课方式', value: schedulingMode.value === 'repeat' ? '重复排课' : '自由排课' },
  { label: '日期设置', value: schedulingMode.value === 'free' ? freeSelectedDatesText.value : rangeText.value },
  { label: '重复规则', value: repeatRuleText.value },
  { label: '上课时间', value: timeModeText.value },
  { label: '上课老师', value: selectedTeacherText.value },
  { label: '上课助教', value: selectedAssistantText.value },
  { label: '上课教室', value: scheduledClassroomText.value },
  { label: '班级学员', value: selectedGroupClassStudentText.value },
])

const overviewTooltipLabels = new Set(['日期设置', '重复规则', '上课时间', '上课教室', '班级学员'])

function isHoliday(date: Dayjs) {
  return schoolHolidaySet.has(date.format('YYYY-MM-DD'))
}

const rawPlannedDates = computed(() => {
  if (schedulingMode.value === 'free')
    return freeSelectedDatesSorted.value.map(item => item.startOf('day'))

  const start = scheduleStartDate.value
  if (!start)
    return []

  const rangeStart = start.startOf('day')
  const sessionCap = scheduleTargetCount.value
  if (sessionCap <= 0)
    return []
  if (repeatRule.value === 'none')
    return [rangeStart]

  const slotCount = Math.max(activeTimeBlocks.value.length, 1)
  const dateTarget = Math.max(1, Math.ceil(sessionCap / slotCount))
  const result: Dayjs[] = []
  let cursor = rangeStart
  let guard = 0

  while (result.length < dateTarget && guard < 5000) {
    if (repeatRule.value === 'daily') {
      result.push(cursor)
    }
    else if (repeatRule.value === 'alternateDay') {
      const diff = cursor.diff(rangeStart, 'day')
      if (diff % 2 === 0)
        result.push(cursor)
    }
    else {
      const selectedNumbers = new Set(selectedWeekdays.value.map(day => weekdayToNumber[day]))
      const matchedWeekday = selectedNumbers.has(cursor.day())
      const weekDiff = Math.floor(cursor.startOf('day').diff(rangeStart, 'day') / 7)
      if (matchedWeekday && (repeatRule.value === 'weekly' || weekDiff % 2 === 0))
        result.push(cursor)
    }
    cursor = cursor.add(1, 'day')
    guard += 1
  }

  return result
})

const excludedHolidayCount = computed(() => {
  if (holidayPolicy.value !== 'filter')
    return 0
  return rawPlannedDates.value.filter(date => isHoliday(date)).length
})

const filteredHolidayDateLabels = computed(() =>
  rawPlannedDates.value
    .filter(date => holidayPolicy.value === 'filter' && isHoliday(date))
    .map(item => item.format('YYYY-MM-DD')),
)

const plannedDates = computed(() => {
  if (holidayPolicy.value !== 'filter')
    return rawPlannedDates.value
  return rawPlannedDates.value.filter(date => !isHoliday(date))
})

const autoScheduleEndDate = computed(() => {
  if (schedulingMode.value === 'free')
    return freeSelectedDatesSorted.value[freeSelectedDatesSorted.value.length - 1] || scheduleStartDate.value
  return plannedDates.value[plannedDates.value.length - 1] || scheduleStartDate.value
})

const rangeText = computed(() =>
  `${scheduleStartDate.value.format('YYYY-MM-DD')} 至 ${autoScheduleEndDate.value.format('YYYY-MM-DD')}`,
)

const blockedReason = computed(() => {
  if (groupClassLoading.value)
    return '班级数据加载中，请稍候。'
  if (!groupClassRecords.value.length)
    return '暂无可用班级，请先创建班级后再排课。'
  if (!selectedGroupClass.value?.id)
    return '请选择班级后继续排课。'
  if (Number(selectedGroupClass.value?.status) === 2)
    return '当前班级已结班，暂不可创建新日程。'
  if (!isValidStaffId(selectedTeacher.value))
    return '请先选择上课教师。'
  if (schedulingMode.value === 'free' && freeSelectedDatesSorted.value.length === 0)
    return '请至少选择一个上课日期。'
  const planned = Math.floor(Number(plannedClassCount.value) || 0)
  if (schedulingMode.value !== 'free' && planned < 1)
    return '请填写计划上课次数（至少为 1）。'
  if (selectedSchoolTimeSlots.value.length === 0)
    return '请至少选择一节课表节次。'
  if (!activeTimeBlocks.value.length)
    return '当前没有可用的上课时段，请检查时间配置。'
  if (schedulingMode.value === 'repeat' && (repeatRule.value === 'weekly' || repeatRule.value === 'biweekly') && selectedWeekdays.value.length === 0)
    return '请至少选择一个每周上课日。'
  return ''
})

const isSchedulable = computed(() => !blockedReason.value)

const emptyStudentHint = computed(() =>
  selectedGroupClass.value?.id && !(selectedGroupClass.value.studentNames.length || selectedGroupClass.value.studentCount)
    ? '当前班级暂无在班学员，继续创建会生成空日程。'
    : '',
)

const previewRuleText = computed(() => {
  if (schedulingMode.value === 'free')
    return '单次日程'
  if (repeatRule.value === 'weekly' || repeatRule.value === 'biweekly')
    return `${repeatRuleLabelMap[repeatRule.value]} · ${selectedWeekdaysText.value || '-'}`
  return repeatRuleLabelMap[repeatRule.value]
})

const previewPlans = computed<PreviewItem[]>(() => {
  const blocks = activeTimeBlocks.value
  if (!blocks.length)
    return []
  const tone: PreviewTone = isSchedulable.value ? 'pending' : 'blocked'
  const unclipped = plannedDates.value.flatMap(date =>
    blocks.map(block => ({
      date: date.format('YYYY-MM-DD'),
      week: weekDisplayMap[date.day()] || '-',
      rule: previewRuleText.value,
      time: block.rangeText,
      startTime: block.startTime,
      endTime: block.endTime,
      teacher: selectedTeacherText.value,
      assistant: selectedAssistantText.value,
      classroom: scheduledClassroomText.value,
      teacherId: selectedTeacherIdNormalized.value || undefined,
      assistantIds: selectedAssistantValues.value.map(id => String(id)),
      classroomId: normalizedSelectedClassroomId.value || undefined,
      allowStudentConflict: false,
      tone,
    })),
  )
  const cap = scheduleTargetCount.value
  if (cap <= 0)
    return []
  return unclipped.slice(0, cap)
})

const estimatedCount = computed(() => previewPlans.value.length)

function buildValidationItemKey(item: { lessonDate?: string, startTime?: string, endTime?: string }) {
  return `${String(item.lessonDate || '')}|${String(item.startTime || '')}|${String(item.endTime || '')}`
}

const previewPlanConflictMap = computed(() => {
  const map = new Map<string, string[]>()
  const items = previewValidationResult.value?.items || []
  items.forEach((item) => {
    if (item.valid === false)
      map.set(buildValidationItemKey(item), item.conflictTypes || [])
  })
  return map
})

function previewRowConflictTypes(plan: PreviewItem) {
  return previewPlanConflictMap.value.get(buildValidationItemKey({
    lessonDate: plan.date,
    startTime: plan.startTime,
    endTime: plan.endTime,
  })) || []
}

const previewHelperText = computed(() => {
  if (blockedReason.value)
    return blockedReason.value
  if (previewValidating.value)
    return '正在校验班级、老师、助教、学员与教室冲突，请稍候。'
  if (previewHasConflict.value)
    return previewValidationMessage.value || (isSingleScheduleEditMode.value ? '当前日程设置存在冲突，请返回修改后再尝试保存。' : (isBatchPlanEditMode.value ? '当前批次规则存在冲突，请返回修改后再尝试保存。' : '当前排课方案存在冲突，请返回修改后再尝试创建。'))
  if (!estimatedCount.value && excludedHolidayCount.value > 0)
    return '当前日期都命中节假日且已被过滤，请调整日期或关闭节假日过滤。'
  if (!estimatedCount.value)
    return '请先选择有效的排课日期。'
  if (excludedHolidayCount.value > 0)
    return `已根据节假日规则过滤 ${excludedHolidayCount.value} 节，剩余 ${estimatedCount.value} 节待${isBatchPlanEditMode.value ? '保存' : '创建'}。`
  return `已完成预检，可确认${isBatchPlanEditMode.value ? '保存' : '创建'}。正式提交时服务端仍会再校验一次。`
})

const modalTitleText = computed(() => isBatchPlanEditMode.value ? '编辑班级日程' : '创建班级日程')
const modalSubtitleText = computed(() =>
  isSingleScheduleEditMode.value
    ? '回显当前班课日程信息，可调整本节的日期与时间资源。'
    : (
  isBatchPlanEditMode.value
    ? '回显当前班课批次的生成条件，调整后会整体替换这批日程。'
    : '参考班级基础信息快速批量排课，先把规则配置清楚，再确认创建。'
      ),
)
const summaryCardTitleText = computed(() => isSingleScheduleEditMode.value ? '当前日程' : (isBatchPlanEditMode.value ? '当前批次' : '当前班级'))
const summaryCardDescText = computed(() =>
  isSingleScheduleEditMode.value
    ? '来自当前日程的基础信息与编辑摘要。'
    : (isBatchPlanEditMode.value ? '来自当前批次的基础信息与规则摘要。' : '来自当前班级档案的基础信息与创建摘要。'),
)
const overviewCardTitleText = computed(() => isSingleScheduleEditMode.value ? '编辑摘要' : (isBatchPlanEditMode.value ? '规则摘要' : '创建摘要'))
const formCardDescText = computed(() =>
  isSingleScheduleEditMode.value
    ? '当前为单节班课编辑，可调整开始日期与时间资源，其他日期规则已锁定。'
    : (isBatchPlanEditMode.value ? '回显当前批次的生成条件，调整后整体替换本批次。' : '按顺序完成排课方式、日期规则和时间资源。'),
)
const reviewTitleText = computed(() => isSingleScheduleEditMode.value ? '预计保存结果' : (isBatchPlanEditMode.value ? '预计替换清单' : '预计排课清单'))
const reviewSubtitleText = computed(() =>
  isSingleScheduleEditMode.value
    ? '先确认本次将保存的班课日程，再执行保存。'
    : (isBatchPlanEditMode.value ? '先确认本次将替换出的班课日程，再执行整体保存。' : '先确认本次将创建的班课日程，再执行批量创建。'),
)
const selectedRecordPlaceholderText = computed(() => {
  if (groupClassLoading.value)
    return '正在加载班级数据...'
  return isSingleScheduleEditMode.value ? '当前日程对应的班级' : (isBatchPlanEditMode.value ? '当前批次对应的班级' : '请选择班级')
})
const datePlanEndHintText = computed(() =>
  isSingleScheduleEditMode.value ? '单节日程的结束日期与开始日期保持一致' : '根据计划上课次数与重复规则自动推算',
)
const datePlanFooterHintText = computed(() =>
  isSingleScheduleEditMode.value ? '当前为单节班课编辑，开始日期可以调整，计划次数固定为 1。' : '可自由填写节数；结束日期由开始日期、重复规则与本次数推算。',
)

const footerTipText = computed(() => {
  if (selectedGroupClass.value?.remark)
    return `班级备注：${selectedGroupClass.value.remark}`
  if (isSingleScheduleEditMode.value)
    return '保存后会更新当前这节班课日程。'
  return isBatchPlanEditMode.value
    ? '保存后会整体替换当前批次；原批次日程会被撤销，新批次会按当前规则重建。'
    : '创建后仍可在课表中继续调整老师、教室和班级学员。'
})

const actionButtonText = computed(() => {
  if (isSingleScheduleEditMode.value)
    return '保存班课日程'
  if (isBatchPlanEditMode.value)
    return estimatedCount.value > 0 ? `保存并替换 ${estimatedCount.value} 节` : '保存班课规则'
  if (schedulingMode.value === 'free')
    return estimatedCount.value > 0 ? '创建单次日程' : '创建日程'
  return estimatedCount.value > 0 ? `批量创建 ${estimatedCount.value} 节` : '批量创建日程'
})

const conflictWorkbenchPeriodGroups = computed(() =>
  groupOptions.value.map((option) => {
    const group = sortedPeriodGroups.value[periodGroupIndexForKey(option.key)]
    const teacherIds = Array.isArray(group?.boundTeachers)
      ? group.boundTeachers.map(item => String(item.id ?? '').trim()).filter(Boolean)
      : []
    return {
      key: option.key,
      label: option.label,
      teacherIds,
      timeOptions: slotsForGroupKey(option.key).map(slot => ({
        value: `${option.key}|${slot.start}|${slot.end}`,
        label: `第${slot.index}节课 · ${slot.start}-${slot.end}`,
        startTime: slot.start,
        endTime: slot.end,
      })),
    }
  }),
)

async function fetchGroupClassRecords() {
  if (groupClassLoading.value)
    return
  groupClassLoading.value = true
  try {
    const pageSize = 200
    let pageIndex = 1
    let total = 0
    const allRows: GroupClassRecord[] = []

    do {
      const res = await pageGroupClassesApi({
        pageRequestModel: {
          needTotal: true,
          pageSize,
          pageIndex,
          skipCount: (pageIndex - 1) * pageSize,
        },
        queryModel: {
          statues: [1],
        },
      })
      if (res.code !== 200)
        throw new Error(res.message || '获取班级列表失败')
      const pageList = Array.isArray(res.result?.list) ? res.result.list : []
      total = Number(res.result?.total || 0)
      allRows.push(...pageList.map(mapGroupClassRow))
      if (!pageList.length)
        break
      pageIndex += 1
    } while (allRows.length < total)

    groupClassRecords.value = allRows
  }
  catch (error: any) {
    console.error('fetch group class records failed', error)
    groupClassRecords.value = []
    messageService.error(error?.message || '获取班级列表失败')
  }
  finally {
    groupClassLoading.value = false
  }
}

function upsertGroupClassRecord(next: GroupClassRecord) {
  const map = new Map(groupClassRecords.value.map(item => [item.id, item]))
  const existing = map.get(next.id)
  map.set(next.id, normalizeGroupClassRecord({
    ...existing,
    ...next,
    teacherIds: next.teacherIds.length ? next.teacherIds : existing?.teacherIds,
    teacherNames: next.teacherNames.length ? next.teacherNames : existing?.teacherNames,
    studentNames: next.studentNames.length ? next.studentNames : existing?.studentNames,
    studentCount: next.studentCount || existing?.studentCount,
    detailLoaded: next.detailLoaded || existing?.detailLoaded,
  }))
  groupClassRecords.value = [...map.values()]
  return map.get(next.id) || null
}

async function ensureSelectedGroupClassLoaded(id: string) {
  const normalizedId = String(id || '').trim()
  if (!normalizedId)
    return null
  const existing = groupClassRecords.value.find(item => item.id === normalizedId)
  if (existing?.detailLoaded && (existing.studentNames.length || existing.studentCount === 0))
    return existing

  groupClassDetailLoading.value = true
  try {
    const [detailRes, studentsRes] = await Promise.all([
      getGroupClassDetailApi({ id: normalizedId }),
      listGroupClassStudentsByClassIdsApi({ classIds: [normalizedId] }),
    ])
    if (detailRes.code !== 200 || !detailRes.result)
      throw new Error(detailRes.message || '获取班级详情失败')
    if (studentsRes.code !== 200)
      throw new Error(studentsRes.message || '获取班级学员失败')

    const bucket = Array.isArray(studentsRes.result)
      ? studentsRes.result.find(item => String(item?.classId || '').trim() === normalizedId)
      : null

    return upsertGroupClassRecord(normalizeGroupClassRecord({
      ...(existing || { id: normalizedId, name: detailRes.result.name || normalizedId }),
      ...mapGroupClassDetail(detailRes.result),
      studentNames: normalizeStudentBucketNames(bucket?.students),
      studentCount: Array.isArray(bucket?.students) ? bucket!.students.length : Number(detailRes.result.studentCount || 0),
      detailLoaded: true,
    }))
  }
  finally {
    groupClassDetailLoading.value = false
  }
}

async function fetchClassroomList() {
  try {
    const res = await listClassroomsApi({ enabledOnly: true })
    if (res.code !== 200) {
      messageService.error(res.message || '获取教室列表失败')
      return
    }
    classroomList.value = Array.isArray(res.result) ? res.result : []
  }
  catch (error: any) {
    console.error('fetch classroom list for group class failed', error)
    messageService.error(error?.message || '获取教室列表失败')
  }
}

async function fetchWorkbenchTeacherList() {
  try {
    const pageRequestModel = {
      needTotal: false,
      pageSize: 500,
      pageIndex: 1,
      skipCount: 0,
    }
    const [teacherRes, assistantRes] = await Promise.all([
      getUserListApi({
        pageRequestModel,
        queryModel: {
          status: 0,
        },
      }),
      getUserListApi({
        pageRequestModel,
        queryModel: {},
      }),
    ])
    if (teacherRes.code !== 200)
      throw new Error(teacherRes.message || '获取老师列表失败')
    if (assistantRes.code !== 200)
      throw new Error(assistantRes.message || '获取助教列表失败')

    const mapStaff = (item: any) => {
      const label = String(item.nickName || item.name || item.id || '').trim()
      return {
        id: String(item.id ?? '').trim(),
        name: label,
        nickName: label,
        mobile: String(item.mobile ?? '').trim(),
        disabled: item?.disabled === true,
      }
    }
    workbenchTeacherList.value = (Array.isArray(teacherRes.result) ? teacherRes.result : [])
      .map(mapStaff)
      .filter(item => item.id)
    assistantWorkbenchStaffList.value = (Array.isArray(assistantRes.result) ? assistantRes.result : [])
      .sort((left: any, right: any) => Number(Boolean(left?.disabled)) - Number(Boolean(right?.disabled)))
      .map(mapStaff)
      .filter(item => item.id)
  }
  catch (error: any) {
    console.error('fetch workbench teacher list for group class failed', error)
    messageService.error(error?.message || '获取老师/助教列表失败')
  }
}

function closeModal() {
  previewModalOpen.value = false
  conflictModalOpen.value = false
  modalOpen.value = false
}

function buildBatchMetaPayload() {
  return {
    schedulingMode: schedulingMode.value,
    repeatRule: repeatRule.value,
    holidayPolicy: holidayPolicy.value,
    selectedWeekdays: [...selectedWeekdays.value],
    scheduleStartDate: scheduleStartDate.value.format('YYYY-MM-DD'),
    freeSelectedDates: freeSelectedDatesSorted.value.map(item => item.format('YYYY-MM-DD')),
    plannedClassCount: Math.max(0, Math.floor(Number(plannedClassCount.value) || 0)),
  }
}

function buildScheduleCreatePayload(options: {
  assistantIds?: string[]
  excludeIds?: string[]
  plans?: BatchCreatePlan[]
} = {}) {
  const plans = Array.isArray(options.plans) && options.plans.length ? options.plans : previewPlans.value
  const normalizedPlans = plans.map((item) => ({
    lessonDate: item.date,
    startTime: item.startTime,
    endTime: item.endTime,
    teacherId: item.teacherId ? String(item.teacherId) : undefined,
    assistantIds: Array.isArray(item.assistantIds) ? item.assistantIds.map(id => String(id)) : undefined,
    classroomId: item.classroomId ? String(item.classroomId) : undefined,
    allowStudentConflict: item.allowStudentConflict === true,
  }))
  const unionAssistantIds = Array.from(new Set(normalizedPlans.flatMap(item => item.assistantIds || [])))
  const topTeacherId = normalizedPlans[0]?.teacherId || selectedTeacherIdNormalized.value || ''
  const topClassroomId = normalizedPlans[0]?.classroomId || normalizedSelectedClassroomId.value || ''
  return {
    groupClassId: String(selectedGroupClass.value?.id || ''),
    teacherId: topTeacherId,
    assistantIds: Array.isArray(options.assistantIds)
      ? options.assistantIds.map(id => String(id)).filter(Boolean)
      : unionAssistantIds,
    classroomId: topClassroomId,
    excludeIds: Array.isArray(options.excludeIds)
      ? options.excludeIds.map(id => String(id)).filter(Boolean)
      : undefined,
    batchMeta: buildBatchMetaPayload(),
    allowStudentConflict: normalizedPlans.some(item => item.allowStudentConflict === true),
    schedules: normalizedPlans,
  }
}

async function validatePreviewSchedules() {
  if (!selectedGroupClass.value?.id || previewPlans.value.length === 0)
    return
  previewValidating.value = true
  previewHasConflict.value = false
  previewValidationMessage.value = ''
  previewValidationResult.value = null
  try {
    const res = await validateGroupClassSchedulesApi(buildScheduleCreatePayload({
      excludeIds: isBatchPlanEditMode.value ? props.batchPlanPreset?.scheduleIds : undefined,
    }))
    if (res.code !== 200 || !res.result) {
      previewHasConflict.value = true
      previewValidationMessage.value = res.message || '预检失败，请稍后重试。'
      return
    }
    previewHasConflict.value = res.result.valid === false
    previewValidationMessage.value = res.result.message || ''
    previewValidationResult.value = res.result
    if (previewHasConflict.value)
      conflictModalOpen.value = true
  }
  catch (error: any) {
    console.error('validate group class preview schedules failed', error)
    previewHasConflict.value = true
    previewValidationMessage.value = error?.response?.data?.message || error?.message || '预检失败，请稍后重试。'
  }
  finally {
    previewValidating.value = false
  }
}

async function openPreviewModal() {
  if (!isSchedulable.value || estimatedCount.value === 0)
    return
  previewModalOpen.value = true
  await validatePreviewSchedules()
}

function closePreviewModal() {
  previewModalOpen.value = false
}

async function confirmBatchCreate(options: {
  assistantIds?: string[]
  plans?: BatchCreatePlan[]
} = {}) {
  if (!selectedGroupClass.value?.id)
    return
  const isSoftConflictCreate = Array.isArray(options.plans)
    ? options.plans.some(item => item.allowStudentConflict === true)
    : previewPlans.value.some(item => item.allowStudentConflict === true)
  if (isSoftConflictCreate)
    creatingWithSoftConflict.value = true
  else
    creatingSchedules.value = true
  try {
    const payload = buildScheduleCreatePayload(options)
    const res = isBatchPlanEditMode.value
      ? await replaceTeachingScheduleBatchApi({
        batchNo: props.batchPlanPreset?.batchNo,
        ids: props.batchPlanPreset?.scheduleIds,
        groupClassId: String(selectedGroupClass.value?.id || ''),
        ...payload,
      })
      : await createGroupClassSchedulesApi(payload)
    if (res.code !== 200)
      throw new Error(res.message || (isBatchPlanEditMode.value ? '保存班课规则失败' : '创建班课日程失败'))
    const count = res.result?.count || (options.plans?.length || previewPlans.value.length)
    if (isBatchPlanEditMode.value) {
      messageService.success(isSingleScheduleEditMode.value ? '已更新班课日程' : `已按新规则替换 ${count} 节班课日程`)
    }
    else {
      messageService.success(
        isSoftConflictCreate
          ? `已创建 ${count} 节班课日程，并保留学员冲突标记`
          : `已创建 ${count} 节班课日程`,
      )
    }
    emitter.emit(EVENTS.REFRESH_DATA)
    emit('updated')
    previewModalOpen.value = false
    conflictModalOpen.value = false
    modalOpen.value = false
  }
  catch (error: any) {
    console.error('create group class schedules failed', error)
    previewHasConflict.value = true
    previewValidationMessage.value = error?.response?.data?.message || error?.message || (isBatchPlanEditMode.value ? '保存班课规则失败' : '创建班课日程失败')
    if (previewValidationResult.value)
      conflictModalOpen.value = true
    messageService.error(previewValidationMessage.value)
  }
  finally {
    creatingSchedules.value = false
    creatingWithSoftConflict.value = false
  }
}

function handleConflictWorkbenchSubmit(payload: { plans: BatchCreatePlan[], assistantIds?: string[] }) {
  void confirmBatchCreate({
    assistantIds: payload.assistantIds,
    plans: payload.plans,
  })
}

function handleTeacherChange(_value: string | number | undefined, staff?: StaffOptionItem | null) {
  selectedTeacherDisplay.value = staff || null
}

function handleAssistantChange(values?: Array<string | number>) {
  const nextValues = Array.isArray(values) ? values : selectedAssistantValues.value
  selectedAssistant.value = nextValues.length ? nextValues : undefined
  selectedAssistantDisplays.value = nextValues.map((id) => {
    return assistantSelectStaffs.value.find(item => sameStaffId(item.id, id))
  }).filter(Boolean) as StaffOptionItem[]
}

async function applyBatchPlanPreset(preset?: GroupClassBatchPlanModalPreset | null) {
  if (!preset)
    return

  const groupClassId = String(preset.groupClassId || '').trim()
  if (!groupClassId)
    return

  applyingBatchPlanPreset = true
  try {
    selectedGroupClassId.value = groupClassId
    await ensureSelectedGroupClassLoaded(groupClassId)
    await nextTick()

    schedulingMode.value = preset.schedulingMode as SchedulingMode
    repeatRule.value = preset.repeatRule as RepeatRule
    holidayPolicy.value = preset.holidayPolicy
    selectedWeekdays.value = preset.selectedWeekdays.length ? [...preset.selectedWeekdays] : ['周一']
    scheduleStartDate.value = dayjs(preset.scheduleStartDate || dayjs()).startOf('day')
    freeSelectedDates.value = (preset.freeSelectedDates.length ? preset.freeSelectedDates : [preset.scheduleStartDate || dayjs().format('YYYY-MM-DD')])
      .map(item => dayjs(item).startOf('day'))
      .filter(item => item.isValid())
    if (!freeSelectedDates.value.length)
      freeSelectedDates.value = [dayjs().startOf('day')]
    freeCalendarPanelDate.value = freeSelectedDates.value[0].startOf('month')
    plannedClassCount.value = Math.max(1, Number(preset.plannedClassCount || 1))
    if (isSingleScheduleEditMode.value) {
      schedulingMode.value = 'repeat'
      repeatRule.value = 'none'
      holidayPolicy.value = 'include'
      selectedWeekdays.value = []
      plannedClassCount.value = 1
      freeSelectedDates.value = [scheduleStartDate.value.startOf('day')]
      freeCalendarPanelDate.value = scheduleStartDate.value.startOf('month')
    }

    const teacherId = String(preset.teacherId || '').trim()
    selectedTeacher.value = teacherId || undefined
    selectedClassroom.value = String(preset.classroomId || '').trim() || undefined

    const matchedSlots = findPresetGroupAndSlotKeys(preset.timeBlocks)
    currentGroup.value = matchedSlots.group
    await nextTick()
    selectedSchoolTimeSlots.value = isSingleScheduleEditMode.value ? matchedSlots.slotKeys.slice(0, 1) : matchedSlots.slotKeys
    selectedTeacherDisplay.value = resolveStaffDisplayById(teacherId)

    const assistantIds = Array.isArray(preset.assistantIds)
      ? preset.assistantIds.map(item => String(item).trim()).filter(Boolean)
      : []
    selectedAssistant.value = assistantIds.length ? assistantIds : undefined
    handleAssistantChange(assistantIds)
  }
  finally {
    applyingBatchPlanPreset = false
  }
}

function toggleFreeScheduleDate(date: Dayjs) {
  const key = date.startOf('day').format('YYYY-MM-DD')
  const exists = freeSelectedDates.value.some(item => item.startOf('day').format('YYYY-MM-DD') === key)
  freeSelectedDates.value = exists
    ? freeSelectedDates.value.filter(item => item.startOf('day').format('YYYY-MM-DD') !== key)
    : [...freeSelectedDates.value, date.startOf('day')]
}

function handleFreeCalendarPanelChange(date: Dayjs) {
  freeCalendarPanelDate.value = date.startOf('month')
}

function jumpFreeCalendarToday() {
  const today = dayjs().startOf('day')
  freeCalendarPanelDate.value = today.startOf('month')
  const todayKey = today.format('YYYY-MM-DD')
  const exists = freeSelectedDates.value.some(item => item.startOf('day').format('YYYY-MM-DD') === todayKey)
  if (!exists)
    freeSelectedDates.value = [...freeSelectedDates.value, today]
}

function clearFreeSelectedDates() {
  freeSelectedDates.value = []
}

function toggleWeekday(day: string) {
  const active = selectedWeekdays.value.includes(day)
  if (active && selectedWeekdays.value.length === 1)
    return
  selectedWeekdays.value = active
    ? selectedWeekdays.value.filter(item => item !== day)
    : [...selectedWeekdays.value, day]
}

const isAllWeekdaysSelected = computed(() => selectedWeekdays.value.length === weekDayOptions.length)

function toggleAllWeekdays() {
  selectedWeekdays.value = isAllWeekdaysSelected.value ? [] : [...weekDayOptions]
}

function invertWeekdays() {
  selectedWeekdays.value = weekDayOptions.filter(item => !selectedWeekdays.value.includes(item))
}

function filterAssistantOption(input: string, option?: { value?: string | number, label?: unknown }) {
  const keyword = String(input || '').trim().toLowerCase()
  if (!keyword)
    return true
  const matched = assistantSelectOptionViews.value.find(item => String(item.value) === String(option?.value ?? ''))
  const haystacks = [
    String(option?.label || ''),
    matched?.mobile || '',
  ]
  return haystacks.some(text => text.toLowerCase().includes(keyword))
}

async function fetchAssistantAvailability() {
  const seq = ++assistantAvailabilitySeq
  const groupClassId = String(selectedGroupClass.value?.id || '').trim()
  const assistantIds = assistantSelectOptionViews.value.map(item => item.value)
  if (!groupClassId || !assistantIds.length || !plannedDates.value.length || !activeTimeBlocks.value.length) {
    assistantAvailabilityMap.value = {}
    assistantAvailabilityLoading.value = false
    return
  }

  const schedules = plannedDates.value.flatMap(date =>
    activeTimeBlocks.value.map(block => ({
      lessonDate: date.format('YYYY-MM-DD'),
      startTime: block.startTime,
      endTime: block.endTime,
    })),
  )
  if (!schedules.length || schedules.length > 2000) {
    assistantAvailabilityMap.value = {}
    assistantAvailabilityLoading.value = false
    return
  }

  assistantAvailabilityLoading.value = true
  try {
    const res = await checkGroupClassAssistantScheduleAvailabilityApi({
      groupClassId,
      assistantIds,
      excludeIds: isBatchPlanEditMode.value ? props.batchPlanPreset?.scheduleIds : undefined,
      schedules,
    })
    if (seq !== assistantAvailabilitySeq)
      return
    if (res.code !== 200 || !res.result)
      throw new Error(res.message || '检测助教时段状态失败')

    const nextMap: Record<string, AvailabilityBadgeView> = {}
    assistantIds.forEach((id) => {
      nextMap[id] = scheduleAvailabilityBadge('free', '空闲')
    })
    ;(res.result.items || []).forEach((item) => {
      nextMap[item.assistantId] = scheduleAvailabilityBadge(item.valid ? 'free' : 'busy', item.valid ? '空闲' : '繁忙')
    })
    assistantAvailabilityMap.value = nextMap
  }
  catch (error) {
    if (seq !== assistantAvailabilitySeq)
      return
    console.error('fetch group class assistant availability failed', error)
    assistantAvailabilityMap.value = {}
  }
  finally {
    if (seq === assistantAvailabilitySeq)
      assistantAvailabilityLoading.value = false
  }
}

function scheduleAssistantAvailabilityCheck() {
  if (assistantAvailabilityTimer)
    clearTimeout(assistantAvailabilityTimer)
  assistantAvailabilityTimer = setTimeout(() => {
    void fetchAssistantAvailability()
  }, 180)
}

function schoolSlotMaxTagPlaceholder(omittedValues: { label?: unknown, value?: unknown }[]) {
  const count = omittedValues?.length ?? 0
  return count > 0 ? `+${count}` : ''
}

async function fetchSlotAvailability() {
  const seq = ++slotAvailabilitySeq
  const groupClassId = String(selectedGroupClass.value?.id || '').trim()
  const teacherId = selectedTeacherIdNormalized.value
  if (!groupClassId || !teacherId || !plannedDates.value.length || !schoolTimeSlotOptions.value.length) {
    slotAvailabilityMap.value = {}
    slotAvailabilityLoading.value = false
    return
  }

  const schedules = plannedDates.value.flatMap(date =>
    schoolTimeSlotOptions.value.map(slot => ({
      lessonDate: date.format('YYYY-MM-DD'),
      startTime: slot.start,
      endTime: slot.end,
      teacherId,
      assistantIds: selectedAssistantValues.value.map(id => String(id)),
      classroomId: normalizedSelectedClassroomId.value || undefined,
    })),
  )

  if (!schedules.length || schedules.length > 2000) {
    slotAvailabilityMap.value = {}
    slotAvailabilityLoading.value = false
    return
  }

  slotAvailabilityLoading.value = true
  try {
    const res = await validateGroupClassSchedulesApi({
      groupClassId,
      teacherId,
      assistantIds: selectedAssistantValues.value.map(id => String(id)),
      classroomId: normalizedSelectedClassroomId.value || undefined,
      excludeIds: isBatchPlanEditMode.value ? props.batchPlanPreset?.scheduleIds : undefined,
      schedules,
    })
    if (seq !== slotAvailabilitySeq)
      return
    if (res.code !== 200 || !res.result)
      throw new Error(res.message || '检测节次可排状态失败')

    const counter = new Map<string, { total: number, invalid: number }>()
    schoolTimeSlotOptions.value.forEach((slot) => {
      counter.set(slot.value, { total: 0, invalid: 0 })
    })
    ;(res.result.items || []).forEach((item) => {
      const slot = schoolTimeSlotOptions.value.find(candidate => candidate.start === item.startTime && candidate.end === item.endTime)
      if (!slot)
        return
      const current = counter.get(slot.value) || { total: 0, invalid: 0 }
      current.total += 1
      if (item.valid === false)
        current.invalid += 1
      counter.set(slot.value, current)
    })
    const nextMap: Record<string, AvailabilityBadgeView> = {}
    counter.forEach((stat, slotKey) => {
      if (stat.total === 0 || stat.invalid === 0) {
        nextMap[slotKey] = scheduleAvailabilityBadge('free', '可排')
        return
      }
      if (stat.invalid === stat.total) {
        nextMap[slotKey] = scheduleAvailabilityBadge('busy', '有冲突')
        return
      }
      nextMap[slotKey] = scheduleAvailabilityBadge('busy', '部分冲突')
    })
    slotAvailabilityMap.value = nextMap
  }
  catch (error) {
    if (seq !== slotAvailabilitySeq)
      return
    console.error('fetch group class slot availability failed', error)
    slotAvailabilityMap.value = {}
  }
  finally {
    if (seq === slotAvailabilitySeq)
      slotAvailabilityLoading.value = false
  }
}

function scheduleSlotAvailabilityCheck() {
  if (slotAvailabilityTimer)
    clearTimeout(slotAvailabilityTimer)
  slotAvailabilityTimer = setTimeout(() => {
    void fetchSlotAvailability()
  }, 180)
}

watch(classroomOptions, (options) => {
  const current = String(selectedClassroom.value || '').trim()
  if (!current)
    return
  if (!options.some(item => item.value === current))
    selectedClassroom.value = undefined
}, { immediate: true })

watch(
  groupOptions,
  (options) => {
    if (!options.some(item => item.key === currentGroup.value))
      currentGroup.value = options[0]?.key || 'A'
  },
  { immediate: true },
)

watch(
  eligiblePeriodGroupKeys,
  () => {
    if (applyingBatchPlanPreset)
      return
    const eligible = eligiblePeriodGroupKeys.value
    if (eligible.length === 1) {
      currentGroup.value = eligible[0]
      return
    }
    if (eligible.length > 1 && !eligible.includes(currentGroup.value))
      currentGroup.value = eligible[0]
  },
  { flush: 'post', immediate: true },
)

watch(
  schoolTimeSlotOptions,
  (options) => {
    if (!options.length) {
      if (selectedSchoolTimeSlots.value.length)
        selectedSchoolTimeSlots.value = []
      return
    }
    const valid = new Set(options.map(item => item.value))
    const next = selectedSchoolTimeSlots.value.filter(value => valid.has(value))
    if (!scheduleSlotKeysEqual(next, selectedSchoolTimeSlots.value))
      selectedSchoolTimeSlots.value = next
  },
  { immediate: true },
)

watch(
  () => selectedGroupClassId.value,
  async (value) => {
    const currentSeq = ++selectedGroupClassSeq
    const normalizedValue = String(value || '').trim()
    const preservePresetSelection = isBatchPlanEditMode.value
      && !!props.batchPlanPreset
      && normalizedValue !== ''
      && normalizedValue === String(props.batchPlanPreset.groupClassId || '').trim()
    if (!value) {
      selectedTeacher.value = undefined
      selectedTeacherDisplay.value = null
      selectedClassroom.value = undefined
      selectedAssistant.value = undefined
      selectedAssistantDisplays.value = []
      selectedSchoolTimeSlots.value = []
      previewValidationResult.value = null
      previewHasConflict.value = false
      previewValidationMessage.value = ''
      slotAvailabilityMap.value = {}
      assistantAvailabilityMap.value = {}
      if (preservePresetSelection)
        return
      return
    }
    try {
      const current = await ensureSelectedGroupClassLoaded(normalizedValue)
      if (currentSeq !== selectedGroupClassSeq || !current)
        return
      previewValidationResult.value = null
      previewHasConflict.value = false
      previewValidationMessage.value = ''
      slotAvailabilityMap.value = {}
      assistantAvailabilityMap.value = {}
      if (preservePresetSelection)
        return
      const defaultTeacherId = isValidStaffId(current.defaultTeacherId)
        ? current.defaultTeacherId
        : teacherPresetStaff.value[0]?.id
      selectedTeacher.value = defaultTeacherId || undefined
      selectedTeacherDisplay.value = resolveStaffDisplayById(defaultTeacherId || '')
      selectedClassroom.value = isValidClassroomId(current.classroomId)
        ? String(current.classroomId).trim()
        : undefined
      selectedAssistant.value = undefined
      selectedAssistantDisplays.value = []
      selectedSchoolTimeSlots.value = []
      scheduleStartDate.value = dayjs().startOf('day')
      freeSelectedDates.value = [dayjs().startOf('day')]
      freeCalendarPanelDate.value = freeSelectedDates.value[0].startOf('month')
      plannedClassCount.value = 1
    }
    catch (error: any) {
      if (currentSeq !== selectedGroupClassSeq)
        return
      console.error('load selected group class detail failed', error)
      messageService.error(error?.message || '加载班级详情失败')
    }
  },
  { immediate: true },
)

watch(selectedTeacher, (value) => {
  if (!isValidStaffId(value))
    return
  const currentAssistantValues = selectedAssistantValues.value
  const next = currentAssistantValues.filter(id => !sameStaffId(id, value))
  if (next.length !== currentAssistantValues.length) {
    selectedAssistant.value = next.length ? next : undefined
    handleAssistantChange(next)
  }
})

watch(
  () => [currentGroup.value, assistantSelectStaffs.value.map(item => String(item.id)).join(',')].join('|'),
  () => {
    if (applyingBatchPlanPreset)
      return
    const currentAssistantValues = selectedAssistantValues.value
    const next = currentAssistantValues.filter(id => isAssistantAllowedInCurrentGroup(String(id)) || isPresetAssistantId(id))
    if (next.length !== currentAssistantValues.length) {
      selectedAssistant.value = next.length ? next : undefined
      handleAssistantChange(next)
    }
  },
  { immediate: true },
)

watch(
  () => [
    String(selectedGroupClass.value?.id || ''),
    plannedDates.value.map(item => item.format('YYYY-MM-DD')).join(','),
    activeTimeBlocks.value.map(item => `${item.startTime}-${item.endTime}`).join(','),
    assistantSelectStaffs.value.map(item => String(item.id)).join(','),
  ].join('|'),
  () => {
    scheduleAssistantAvailabilityCheck()
  },
  { immediate: true },
)

watch(
  () => [
    String(selectedGroupClass.value?.id || ''),
    selectedTeacherIdNormalized.value,
    currentGroup.value,
    plannedDates.value.map(item => item.format('YYYY-MM-DD')).join(','),
    schoolTimeSlotOptions.value.map(item => `${item.start}-${item.end}`).join(','),
    selectedAssistantValues.value.map(item => String(item)).join(','),
    normalizedSelectedClassroomId.value,
  ].join('|'),
  () => {
    scheduleSlotAvailabilityCheck()
  },
  { immediate: true },
)

watch(scheduleStartDate, () => {
  if (modalOpen.value)
    void loadEffectivePeriodConfig()
})

watch(modalOpen, async (value) => {
  if (value) {
    selectedGroupClassId.value = undefined
    selectedTeacher.value = undefined
    selectedTeacherDisplay.value = null
    selectedAssistant.value = undefined
    selectedAssistantDisplays.value = []
    selectedClassroom.value = undefined
    previewValidationResult.value = null
    previewHasConflict.value = false
    previewValidationMessage.value = ''
    previewModalOpen.value = false
    conflictModalOpen.value = false
    await nextTick()
    scrollPlannerShellToTop()
    await Promise.all([
      fetchGroupClassRecords(),
      fetchClassroomList(),
      fetchWorkbenchTeacherList(),
    ])
    await loadEffectivePeriodConfig()
    if (isBatchPlanEditMode.value && props.batchPlanPreset)
      await applyBatchPlanPreset(props.batchPlanPreset)
    await nextTick()
    scrollPlannerShellToTop()
    requestAnimationFrame(() => scrollPlannerShellToTop())
  }
  else {
    previewModalOpen.value = false
    conflictModalOpen.value = false
    slotAvailabilityMap.value = {}
    assistantAvailabilityMap.value = {}
  }
}, { immediate: true })

watch(
  () => [modalOpen.value, isBatchPlanEditMode.value, props.batchPlanPreset] as const,
  async ([open, editMode, preset]) => {
    if (!open || !editMode || !preset || !groupClassRecords.value.length)
      return
    await applyBatchPlanPreset(preset)
  },
)
</script>

<template>
  <div class="group-class-schedule-modal-root">
    <a-modal
      v-model:open="modalOpen"
      centered
      class="group-class-schedule-modal"
      :width="1140"
      :body-style="{ padding: '0' }"
      :keyboard="false"
      :closable="false"
      :mask-closable="true"
      @cancel="closeModal"
    >
      <template #title>
        <div class="planner-head">
          <div class="planner-head__main">
            <div class="planner-head__title">
              {{ modalTitleText }}
            </div>
            <div class="planner-head__subtitle">
              {{ modalSubtitleText }}
            </div>
          </div>

          <div class="planner-head__stats">
            <div class="planner-stat">
              <strong>{{ estimatedCount }}</strong>
              <span>节预计生成</span>
            </div>
            <div v-if="excludedHolidayCount" class="planner-stat planner-stat--subtle">
              <strong>{{ excludedHolidayCount }}</strong>
              <span>节已过滤</span>
            </div>
          </div>

          <a-button type="text" class="planner-head__close" @click="closeModal">
            <template #icon>
              <CloseOutlined />
            </template>
          </a-button>
        </div>
      </template>

      <template #footer>
        <div class="planner-footer">
          <div class="planner-footer__tip">
            {{ footerTipText }}
          </div>

          <div class="planner-footer__actions">
            <a-button @click="closeModal">
              取消
            </a-button>
            <a-button type="primary" :disabled="!isSchedulable || estimatedCount === 0" :loading="creatingSchedules" @click="openPreviewModal">
              {{ actionButtonText }}
            </a-button>
          </div>
        </div>
      </template>

      <div ref="plannerShellRef" class="planner-shell">
        <section class="planner-strip planner-strip--top">
          <div class="planner-inline planner-inline--top">
            <div class="planner-inline__group planner-inline__group--record">
              <span class="planner-label planner-label--inline">
                <TeamOutlined />
                选择班级
              </span>
              <a-select
                v-model:value="selectedGroupClassId"
                size="large"
                show-search
                option-filter-prop="label"
                option-label-prop="label"
                popup-class-name="planner-record-select-dropdown"
                allow-clear
                :loading="groupClassLoading || groupClassDetailLoading"
                :disabled="isBatchPlanEditMode || groupClassLoading || !groupClassSelectOptions.length"
                :not-found-content="groupClassLoading ? '正在加载班级数据...' : '暂无班级数据'"
                :placeholder="selectedRecordPlaceholderText"
                class="planner-control planner-control--record"
              >
                <a-select-option
                  v-for="item in groupClassSelectOptions"
                  :key="item.value"
                  :value="item.value"
                  :label="item.label"
                >
                  <div class="planner-option">
                    <div class="planner-option__title">
                      {{ item.label }}
                    </div>
                    <div class="planner-option__meta">
                      <span>{{ item.statusText }}</span>
                      <span>{{ item.studentCount }} 名学员</span>
                    </div>
                  </div>
                </a-select-option>
              </a-select>
            </div>

            <div class="planner-inline__group planner-inline__group--status">
              <div class="planner-balance">
                <div class="planner-balance__item">
                  <span>状态</span>
                  <strong>{{ groupClassStatusText }}</strong>
                </div>
                <div class="planner-balance__item">
                  <span>在班学员</span>
                  <strong>{{ selectedGroupClassStudentCountText }}</strong>
                </div>
              </div>
            </div>
          </div>
        </section>

        <div v-if="selectedGroupClass?.id" class="planner-layout">
          <aside class="planner-aside">
            <section class="planner-card planner-card--summary">
              <div class="planner-card__head">
                <div class="planner-card__title">
                  {{ summaryCardTitleText }}
                </div>
                <div class="planner-card__desc">
                  {{ summaryCardDescText }}
                </div>
              </div>

              <div class="planner-profile">
                <div class="planner-profile__name">
                  {{ selectedGroupClass?.name || '-' }}
                </div>
                <div class="planner-profile__meta">
                  {{ selectedGroupClass?.courseName || '-' }} · {{ selectedGroupClassStudentCountText }}
                </div>
              </div>

              <div class="planner-summary-list">
                <div
                  v-for="item in selectedGroupClassSummary"
                  :key="item.label"
                  class="planner-summary-list__row"
                >
                  <span>{{ item.label }}</span>
                  <strong class="planner-summary-list__value">{{ item.value }}</strong>
                </div>
              </div>

              <div class="planner-card__subhead">
                {{ overviewCardTitleText }}
              </div>

              <div class="planner-summary-list planner-summary-list--secondary">
                <div
                  v-for="item in overviewItems"
                  :key="item.label"
                  class="planner-summary-list__row"
                >
                  <span>{{ item.label }}</span>
                  <a-tooltip v-if="overviewTooltipLabels.has(item.label)" :title="item.value">
                    <span class="planner-summary-list__value-wrap">
                      <strong class="planner-summary-list__value">{{ item.value }}</strong>
                    </span>
                  </a-tooltip>
                  <strong v-else class="planner-summary-list__value">{{ item.value }}</strong>
                </div>
              </div>

              <div v-if="selectedGroupClass?.remark" class="planner-note">
                {{ selectedGroupClass.remark }}
              </div>

              <div
                v-for="(warning, index) in (isBatchPlanEditMode ? props.batchPlanPreset?.warnings || [] : [])"
                :key="`group-preset-warning-${index}`"
                class="planner-alert planner-alert--soft"
              >
                {{ warning }}
              </div>

              <div v-if="emptyStudentHint" class="planner-alert planner-alert--soft">
                {{ emptyStudentHint }}
              </div>

              <div v-if="blockedReason" class="planner-alert">
                {{ blockedReason }}
              </div>
            </section>
          </aside>

          <main class="planner-main">
            <section class="planner-card planner-card--form">
              <div class="planner-card__head">
                <div class="planner-card__title">
                  排课设置
                </div>
                <div class="planner-card__desc">
                  {{ formCardDescText }}
                </div>
              </div>

              <div v-if="showSchedulingModeSection" class="planner-section">
                <div class="planner-section__title">
                  排课方式
                </div>

                <div class="planner-choice-row">
                  <button
                    v-for="item in schedulingModeOptions"
                    :key="item.value"
                    type="button"
                    class="planner-choice"
                    :class="{ 'planner-choice--active': schedulingMode === item.value }"
                    @click="schedulingMode = item.value"
                  >
                    <span class="planner-choice__title">{{ item.label }}</span>
                    <span class="planner-choice__desc">{{ item.desc }}</span>
                  </button>
                </div>
              </div>

              <div class="planner-section">
                <div class="planner-section__title">
                  日期规则
                </div>

                <div class="planner-form-grid">
                  <div v-if="schedulingMode === 'repeat'" class="planner-date-plan-row">
                    <label class="planner-field planner-date-plan-row__cell planner-date-plan-row__cell--start">
                      <span class="planner-label planner-label--required">
                        <CalendarOutlined />
                        开始日期
                      </span>
                      <a-date-picker
                        v-model:value="scheduleStartDate"
                        size="large"
                        class="planner-control"
                        :allow-clear="false"
                      />
                    </label>

                    <label class="planner-field planner-date-plan-row__cell planner-date-plan-row__cell--count">
                      <span class="planner-label planner-label--required">
                        计划上课次数
                      </span>
                      <a-input-number
                        v-model:value="plannedClassCount"
                        :min="1"
                        :precision="0"
                        :disabled="isSingleScheduleEditMode"
                        size="large"
                        class="planner-control"
                      />
                    </label>

                    <div class="planner-field planner-date-plan-row__cell">
                      <span class="planner-label">
                        <CalendarOutlined />
                        结束日期
                      </span>
                      <div class="planner-static-field planner-static-field--compact planner-static-field--inline-hint">
                        <strong>{{ autoScheduleEndDate.format('YYYY-MM-DD') }}</strong>
                        <span>{{ datePlanEndHintText }}</span>
                      </div>
                    </div>

                    <span class="planner-field__hint planner-date-plan-row__hint">{{ datePlanFooterHintText }}</span>
                  </div>

                  <div
                    v-else
                    class="planner-date-plan-row planner-date-plan-row--free"
                  >
                    <div class="planner-field planner-date-plan-row__cell planner-date-plan-row__cell--start">
                      <div class="planner-field__label-row">
                        <span class="planner-label planner-label--required">
                          <CalendarOutlined />
                          上课日期
                        </span>
                        <a-button
                          type="link"
                          size="small"
                          :disabled="freeSelectedDatesSorted.length === 0"
                          @click.stop.prevent="clearFreeSelectedDates"
                        >
                          一键清空
                        </a-button>
                      </div>
                      <a-popover
                        v-model:open="freeCalendarOpen"
                        trigger="click"
                        placement="bottomLeft"
                        overlay-class-name="planner-free-calendar-popover"
                      >
                        <template #content>
                          <div class="planner-free-calendar">
                            <a-calendar
                              :fullscreen="false"
                              :value="freeCalendarPanelDate"
                              @panel-change="handleFreeCalendarPanelChange"
                              @select="toggleFreeScheduleDate"
                            >
                              <template #dateFullCellRender="{ current }">
                                <div
                                  class="planner-free-calendar__cell"
                                  :class="{ 'planner-free-calendar__cell--selected': freeSelectedDateKeys.has(current.format('YYYY-MM-DD')) }"
                                >
                                  {{ current.date() }}
                                </div>
                              </template>
                            </a-calendar>
                            <div class="planner-free-calendar__footer">
                              <a-button type="link" size="small" @click="jumpFreeCalendarToday">
                                今天
                              </a-button>
                            </div>
                          </div>
                        </template>

                        <div class="planner-static-field planner-static-field--compact planner-static-field--free-trigger">
                          <strong :title="freeSelectedDatesText">{{ freeSelectedDatesText }}</strong>
                        </div>
                      </a-popover>
                    </div>
                  </div>

                  <div v-if="schedulingMode === 'repeat' && !isSingleScheduleEditMode" class="planner-field planner-field--full">
                    <span class="planner-label planner-label--required">
                      重复规则
                    </span>
                    <div class="planner-chip-row">
                      <button
                        v-for="item in repeatRuleOptions"
                        :key="item.value"
                        type="button"
                        class="planner-chip"
                        :class="{ 'planner-chip--active': repeatRule === item.value }"
                        @click="repeatRule = item.value"
                      >
                        {{ item.label }}
                      </button>
                    </div>
                  </div>

                  <div
                    v-if="schedulingMode === 'repeat' && (repeatRule === 'weekly' || repeatRule === 'biweekly')"
                    class="planner-field planner-field--full"
                  >
                    <span class="planner-label planner-label--required">
                      每周上课日
                    </span>
                    <div class="planner-chip-row planner-chip-row--weekday">
                      <button
                        v-for="item in weekDayOptions"
                        :key="item"
                        type="button"
                        class="planner-chip"
                        :class="{ 'planner-chip--active': selectedWeekdays.includes(item) }"
                        @click="toggleWeekday(item)"
                      >
                        {{ item }}
                      </button>
                      <div class="planner-chip-actions">
                        <a-button type="link" size="small" @click="toggleAllWeekdays">
                          {{ isAllWeekdaysSelected ? '取消全选' : '全选' }}
                        </a-button>
                        <a-button type="link" size="small" @click="invertWeekdays">
                          反选
                        </a-button>
                      </div>
                    </div>
                  </div>

                  <div v-if="!isSingleScheduleEditMode" class="planner-field planner-field--full">
                    <span class="planner-label planner-label--required">
                      节假日
                      <a-tooltip title="当前示例会按学校节假日配置过滤 2026-05-01 至 2026-05-03。">
                        <QuestionCircleOutlined class="planner-label__tip" />
                      </a-tooltip>
                    </span>
                    <div class="planner-chip-row">
                      <button
                        v-for="item in holidayPolicyOptions"
                        :key="item.value"
                        type="button"
                        class="planner-chip"
                        :class="{ 'planner-chip--active': holidayPolicy === item.value }"
                        @click="holidayPolicy = item.value"
                      >
                        {{ item.label }}
                      </button>
                      <div
                        v-if="holidayPolicy === 'filter' && filteredHolidayDateLabels.length > 0"
                        class="planner-holiday-inline"
                      >
                        <span class="planner-holiday-inline__label">已过滤日期</span>
                        <div class="planner-holiday-inline__list">
                          <span
                            v-for="item in filteredHolidayDateLabels"
                            :key="item"
                            class="planner-holiday-inline__item"
                          >
                            {{ item }}
                          </span>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <div class="planner-section">
                <div class="planner-section__title">
                  时间与资源
                </div>

                <div class="planner-form-grid">
                  <label class="planner-field">
                    <span class="planner-label planner-label--required">
                      <UserOutlined />
                      上课教师
                    </span>
                    <StaffSelect
                      :key="`${scheduleStaffSelectKey}-teacher`"
                      v-model="selectedTeacher"
                      size="large"
                      placeholder="请选择上课教师"
                      width="100%"
                      :multiple="false"
                      :status="0"
                      :allow-clear="false"
                      :preset-staff="teacherPresetStaff"
                      class="planner-control"
                      @change="handleTeacherChange"
                    />
                  </label>

                  <label class="planner-field">
                    <span class="planner-label">
                      <EnvironmentOutlined />
                      上课教室
                    </span>
                    <a-select
                      v-model:value="selectedClassroom"
                      size="large"
                      allow-clear
                      placeholder="不选则默认班级教室"
                      :options="classroomOptions"
                      class="planner-control"
                    />
                  </label>

                  <div class="planner-field planner-field--major planner-field--full">
                    <div class="planner-field__label-row planner-field__label-row--with-period-group">
                      <span class="planner-label planner-label--required planner-label--with-inline-tip">
                        <ClockCircleOutlined />
                        {{ schoolSlotFieldLabelText }}
                        <a-tooltip :title="schoolSlotTooltipText">
                          <QuestionCircleOutlined class="planner-label__tip" />
                        </a-tooltip>
                      </span>
                      <span class="planner-period-group-wrap">
                        <a-radio-group
                          v-model:value="currentGroup"
                          button-style="solid"
                          size="small"
                          class="planner-period-group-radio planner-period-group-radio--inline"
                        >
                          <a-radio-button
                            v-for="opt in groupOptions"
                            :key="opt.key"
                            :value="opt.key"
                            :disabled="isPeriodGroupChoiceDisabled(opt.key)"
                          >
                            {{ opt.label }}
                          </a-radio-button>
                        </a-radio-group>
                        <a-tooltip title="节次来自机构「时段设置」当前组。选择上课教师后，若该老师仅绑定一个时段组，将自动切换并锁定组别；绑定多个组时可自由切换。">
                          <QuestionCircleOutlined class="planner-label__tip planner-period-group-wrap__tip" />
                        </a-tooltip>
                      </span>
                    </div>
                    <span class="planner-control-tooltip-wrap">
                      <a-select
                        v-model:value="schoolTimeSlotSelectValue"
                        :mode="isSingleScheduleEditMode ? undefined : 'multiple'"
                        size="large"
                        option-label-prop="label"
                        allow-clear
                        :placeholder="schoolSlotPlaceholderText"
                        max-tag-count="responsive"
                        :max-tag-placeholder="schoolSlotMaxTagPlaceholder"
                        :get-popup-container="scheduleSlotSelectGetPopupContainer"
                        popup-class-name="planner-schedule-slot-select-dropdown"
                        class="planner-control planner-control--major planner-multi-slot-select"
                      >
                        <a-select-option
                          v-for="item in schoolTimeSlotOptionViews"
                          :key="item.value"
                          :value="item.value"
                          :label="item.label"
                        >
                          <div class="planner-slot-option">
                            <span class="planner-slot-option__label">{{ item.baseLabel }}</span>
                            <span
                              class="planner-slot-option__status"
                              :class="{
                                'planner-slot-option__status--free': item.status === 'free',
                                'planner-slot-option__status--busy': item.status === 'busy',
                                'planner-slot-option__status--unknown': item.status === 'unknown',
                              }"
                            >
                              {{ item.statusText }}
                            </span>
                          </div>
                        </a-select-option>
                      </a-select>
                    </span>
                  </div>

                  <label class="planner-field planner-field--full">
                    <span class="planner-label">
                      <TeamOutlined />
                      上课助教
                    </span>
                    <a-select
                      v-model:value="selectedAssistant"
                      mode="multiple"
                      show-search
                      :placeholder="`可不选，仅${groupOptions.find(item => item.key === currentGroup)?.label || '当前组'}可选`"
                      size="large"
                      allow-clear
                      option-label-prop="label"
                      :filter-option="filterAssistantOption"
                      popup-class-name="planner-assistant-select-dropdown"
                      class="planner-control planner-multi-slot-select planner-assistant-select"
                      @change="handleAssistantChange"
                    >
                      <a-select-option
                        v-for="item in assistantSelectOptionViews"
                        :key="item.value"
                        :value="item.value"
                        :label="item.label"
                      >
                        <div class="planner-staff-option">
                          <div class="planner-staff-option__main">
                            <span class="planner-staff-option__label">{{ item.baseLabel }}</span>
                            <span v-if="item.mobile" class="planner-staff-option__mobile">{{ item.mobile }}</span>
                          </div>
                          <span
                            class="planner-slot-option__status"
                            :class="{
                              'planner-slot-option__status--free': item.status === 'free',
                              'planner-slot-option__status--busy': item.status === 'busy',
                              'planner-slot-option__status--unknown': item.status === 'unknown',
                            }"
                          >
                            {{ item.statusText }}
                          </span>
                        </div>
                      </a-select-option>
                    </a-select>
                  </label>
                </div>
              </div>
            </section>
          </main>
        </div>

        <section v-else class="planner-card planner-card--empty-state">
          <a-empty
            :description="groupClassLoading ? '正在加载班级数据...' : '请选择班级后查看档案并继续排课'"
          />
        </section>
      </div>
    </a-modal>

    <a-modal
      v-model:open="previewModalOpen"
      centered
      class="planner-review-modal"
      :footer="null"
      :width="980"
      :body-style="{ padding: '0 24px 16px' }"
      :keyboard="false"
      :mask-closable="false"
      @cancel="closePreviewModal"
    >
      <template #title>
        <div class="planner-review__head">
          <div>
            <div class="planner-review__title">
              {{ reviewTitleText }}
            </div>
            <div class="planner-review__subtitle">
              {{ reviewSubtitleText }}
            </div>
          </div>
          <div class="planner-review__count">
            共 {{ estimatedCount }} 节
          </div>
        </div>
      </template>

      <div class="planner-review">
        <div
          class="planner-review__tip"
          :class="{ 'planner-review__tip--warning': !isSchedulable || previewHasConflict }"
        >
          <span class="planner-review__tip-text">{{ previewHelperText }}</span>
          <a-button
            v-if="previewHasConflict && previewValidationResult"
            type="link"
            size="small"
            class="planner-review__conflict-link"
            @click="conflictModalOpen = true"
          >
            查看冲突详情
          </a-button>
        </div>

        <div class="planner-table-wrap planner-table-wrap--modal">
          <table class="planner-table">
            <thead>
              <tr>
                <th>日期</th>
                <th>星期</th>
                <th>规则</th>
                <th>时间</th>
                <th>老师</th>
                <th>助教</th>
                <th>教室</th>
                <th class="planner-table__status-cell">
                  状态
                </th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="!previewPlans.length">
                <td colspan="8" class="planner-table__empty">
                  暂无预计日程
                </td>
              </tr>
              <template v-else>
                <tr
                  v-for="(item, planIdx) in previewPlans"
                  :key="`${item.date}-${item.time}-${planIdx}`"
                  :class="{ 'planner-table__row--blocked': item.tone === 'blocked' }"
                >
                  <td>{{ item.date }}</td>
                  <td>{{ item.week }}</td>
                  <td>{{ item.rule }}</td>
                  <td>{{ item.time }}</td>
                  <td :class="{ 'planner-table__cell--danger': previewRowConflictTypes(item).includes('老师') }">
                    {{ item.teacher }}
                  </td>
                  <td :class="{ 'planner-table__cell--danger': previewRowConflictTypes(item).includes('助教') }">
                    {{ item.assistant }}
                  </td>
                  <td :class="{ 'planner-table__cell--danger': previewRowConflictTypes(item).includes('教室') }">
                    {{ item.classroom }}
                  </td>
                  <td class="planner-table__status-cell">
                    <template v-if="item.tone === 'blocked'">
                      <span class="planner-tag planner-tag--table planner-tag--warning">
                        不可创建
                      </span>
                    </template>
                    <template v-else-if="previewRowConflictTypes(item).length">
                      <a-tag
                        v-for="tag in previewRowConflictTypes(item)"
                        :key="`${planIdx}-${tag}`"
                        color="error"
                        class="planner-review__conflict-tag"
                        :bordered="false"
                      >
                        {{ tag }}冲突
                      </a-tag>
                    </template>
                    <span v-else class="planner-tag planner-tag--table">
                      正常
                    </span>
                  </td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>

        <div class="planner-review__footer">
          <div class="planner-footer__tip">
            {{ footerTipText }}
          </div>

          <div class="planner-footer__actions">
            <a-button @click="closePreviewModal">
              返回修改
            </a-button>
            <a-button type="primary" :disabled="!isSchedulable || estimatedCount === 0 || previewValidating || previewHasConflict" :loading="creatingSchedules" @click="() => confirmBatchCreate()">
              {{ actionButtonText }}
            </a-button>
          </div>
        </div>
      </div>
    </a-modal>
  </div>

  <GroupClassScheduleConflictWorkbenchModal
    v-model:open="conflictModalOpen"
    :group-class-id="String(selectedGroupClass?.id || '')"
    :plans="previewPlans"
    :validation="previewValidationResult"
    :teacher-options="teacherSelectOptions"
    :assistant-options="assistantOptionList"
    :classroom-options="classroomOptions"
    :period-groups="conflictWorkbenchPeriodGroups"
    :time-options="schoolTimeSlotOptions.map(item => ({
      value: item.value,
      label: `${item.desc} · ${item.start}-${item.end}`,
      startTime: item.start,
      endTime: item.end,
    }))"
    :default-group-key="currentGroup"
    :default-teacher-id="String(selectedTeacher || '')"
    :default-classroom-id="normalizedSelectedClassroomId"
    :loading="creatingSchedules || creatingWithSoftConflict"
    @submit="handleConflictWorkbenchSubmit"
  />
</template>

<style scoped lang="less">
.group-class-schedule-modal-root {
  display: contents;
}

.planner-head {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.planner-head__main {
  flex: 1;
  min-width: 0;
}

.planner-head__title {
  color: #1f2329;
  font-size: 20px;
  font-weight: 600;
  line-height: 1.25;
}

.planner-head__subtitle {
  margin-top: 2px;
  color: #86909c;
  font-size: 12px;
  line-height: 1.5;
}

.planner-head__stats {
  display: flex;
  align-items: center;
  gap: 8px;
}

.planner-stat {
  min-width: 108px;
  padding: 8px 10px;
  border-radius: 12px;
  background: linear-gradient(180deg, #f6fbff 0%, #edf6ff 100%);
  text-align: right;
}

.planner-stat strong {
  display: block;
  color: #1677ff;
  font-size: 20px;
  font-weight: 700;
  line-height: 1;
}

.planner-stat span {
  display: block;
  margin-top: 4px;
  color: #6b7785;
  font-size: 11px;
  line-height: 1.4;
}

.planner-stat--subtle {
  background: #f7f8fa;
}

.planner-stat--subtle strong {
  color: #1f2329;
}

.planner-head__close {
  color: #94a3b8;
}

.planner-shell {
  max-height: 78vh;
  overflow: auto;
  padding: 0 24px 24px;
  background: linear-gradient(180deg, #f7faff 0%, #fff 180px);
}

.planner-strip {
  padding-top: 20px;
}

.planner-inline {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

.planner-inline__group {
  display: flex;
  align-items: center;
  gap: 14px;
  min-width: 0;
}

.planner-inline__group--record {
  flex: 1;
  min-width: 360px;
}

.planner-inline__group--status {
  min-width: 260px;
  justify-content: flex-end;
}

.planner-label {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #4b5563;
  font-size: 13px;
  font-weight: 600;
}

.planner-label--inline {
  flex: 0 0 auto;
  margin-bottom: 0;
  white-space: nowrap;
}

.planner-label--required::after {
  content: '*';
  color: #ff4d4f;
}

.planner-label__tip {
  color: #98a2b3;
}

.planner-control {
  width: 100%;
}

.planner-control--record {
  flex: 1 1 auto;
  min-width: 0;
}

:deep(.planner-control .ant-select-selector),
:deep(.planner-control.ant-picker),
:deep(.planner-control.ant-input-number) {
  min-height: 44px !important;
  border-radius: 12px !important;
  border-color: #dfe7f1 !important;
  box-shadow: none !important;
}



.planner-balance {
  display: flex;
  flex: 1 1 auto;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 14px;
  background: #f7f9fc;
}

.planner-balance__item {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 88px;
}

.planner-balance__item span {
  color: #86909c;
  font-size: 12px;
}

.planner-balance__item strong {
  color: #1f2329;
  font-size: 16px;
  font-weight: 700;
}

.planner-layout {
  display: grid;
  grid-template-columns: 320px minmax(0, 1fr);
  gap: 20px;
  margin-top: 20px;
}

.planner-card {
  border: 1px solid #edf2f7;
  border-radius: 20px;
  background: #fff;
  box-shadow: 0 10px 28px rgb(15 23 42 / 6%);
}

.planner-card--summary,
.planner-card--form {
  padding: 20px;
}

.planner-card--empty-state {
  margin-top: 20px;
  padding: 40px 0;
}

.planner-card__head {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.planner-card__title {
  color: #1f2329;
  font-size: 16px;
  font-weight: 700;
}

.planner-card__desc {
  color: #86909c;
  font-size: 12px;
  line-height: 1.6;
}

.planner-card__subhead {
  margin-top: 18px;
  color: #1f2329;
  font-size: 14px;
  font-weight: 700;
}

.planner-profile {
  margin-top: 18px;
  padding: 14px 16px;
  border-radius: 16px;
  background: #f7faff;
}

.planner-profile__name {
  color: #1f2329;
  font-size: 18px;
  font-weight: 700;
}

.planner-profile__meta {
  margin-top: 4px;
  color: #667085;
  font-size: 13px;
}

.planner-summary-list {
  margin-top: 14px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.planner-summary-list__row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  color: #667085;
  font-size: 13px;
}

.planner-summary-list__value-wrap {
  display: inline-flex;
  min-width: 0;
}

.planner-summary-list__value {
  max-width: 180px;
  color: #1f2329;
  font-weight: 700;
  text-align: right;
  line-height: 1.5;
  word-break: break-word;
}

.planner-note,
.planner-alert {
  margin-top: 16px;
  padding: 12px 14px;
  border-radius: 14px;
  background: #fff7e6;
  color: #ad6800;
  font-size: 13px;
  line-height: 1.7;
  border: 1px solid #ffe7ba;
}

.planner-alert {
  background: #fff1f0;
  color: #cf1322;
  border-color: #ffccc7;
}

.planner-alert--soft {
  background: #f6ffed;
  color: #389e0d;
  border-color: #d9f7be;
}

.planner-section + .planner-section {
  margin-top: 22px;
}

.planner-section__title {
  color: #1f2329;
  font-size: 14px;
  font-weight: 700;
  margin-bottom: 14px;
}

.planner-choice-row {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.planner-choice {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-height: 86px;
  padding: 14px 16px;
  border: 1px solid #dfe7f1;
  border-radius: 16px;
  background: #fff;
  text-align: left;
  cursor: pointer;
  transition: all 0.18s ease;
}

.planner-choice:hover {
  border-color: #91caff;
  background: #f7fbff;
}

.planner-choice--active {
  border-color: #1677ff;
  background: #f2f8ff;
  box-shadow: 0 8px 16px rgb(22 119 255 / 10%);
}

.planner-choice__title {
  color: #1f2329;
  font-size: 14px;
  font-weight: 700;
}

.planner-choice__desc {
  color: #667085;
  font-size: 12px;
  line-height: 1.6;
}

.planner-form-grid {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.planner-date-plan-row {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
}

.planner-date-plan-row--free {
  grid-template-columns: minmax(0, 1fr);
}

.planner-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.planner-field--full {
  width: 100%;
}

.planner-field__label-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.planner-field__label-row--with-period-group {
  align-items: flex-start;
}

.planner-static-field {
  display: flex;
  flex-direction: column;
  justify-content: center;
  min-height: 44px;
  padding: 10px 14px;
  border-radius: 12px;
  background: #f8fafc;
  border: 1px solid #e5edf5;
}

.planner-static-field strong {
  color: #1f2329;
  font-size: 14px;
}

.planner-static-field span {
  margin-top: 2px;
  color: #86909c;
  font-size: 12px;
  line-height: 1.5;
}

.planner-static-field--free-trigger {
  cursor: pointer;
}

.planner-chip-row {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.planner-chip {
  min-height: 36px;
  padding: 0 14px;
  border: 1px solid #dfe7f1;
  border-radius: 999px;
  background: #fff;
  color: #1f2329;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
}

.planner-chip--active {
  border-color: #1677ff;
  background: #edf5ff;
  color: #1677ff;
}

.planner-chip-actions {
  display: inline-flex;
  align-items: center;
  gap: 2px;
}

.planner-holiday-inline {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.planner-holiday-inline__label {
  color: #86909c;
  font-size: 12px;
}

.planner-holiday-inline__list {
  display: inline-flex;
  flex-wrap: wrap;
  gap: 6px;
}

.planner-holiday-inline__item {
  display: inline-flex;
  min-height: 24px;
  padding: 0 10px;
  border-radius: 999px;
  background: #fff7e6;
  color: #ad6800;
  font-size: 12px;
  align-items: center;
}

.planner-period-group-wrap {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.planner-review {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.planner-review__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.planner-review__title {
  color: #1f2329;
  font-size: 18px;
  font-weight: 700;
}

.planner-review__subtitle {
  margin-top: 4px;
  color: #86909c;
  font-size: 12px;
}

.planner-review__count {
  color: #1677ff;
  font-size: 14px;
  font-weight: 700;
}

.planner-review__tip {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 14px;
  border-radius: 14px;
  background: #f6ffed;
  color: #389e0d;
  border: 1px solid #d9f7be;
  font-size: 13px;
}

.planner-review__tip--warning {
  background: #fff7e6;
  color: #ad6800;
  border-color: #ffe7ba;
}

.planner-table-wrap {
  border: 1px solid #edf2f7;
  border-radius: 16px;
  overflow: auto;
  background: #fff;
}

.planner-table {
  width: 100%;
  border-collapse: collapse;
  min-width: 760px;
}

.planner-table th,
.planner-table td {
  padding: 12px 14px;
  border-bottom: 1px solid #f2f4f7;
  color: #4b5563;
  font-size: 13px;
  text-align: left;
}

.planner-table thead th {
  background: #f8fafc;
  color: #475467;
  font-weight: 700;
}

.planner-table__empty {
  text-align: center !important;
  color: #98a2b3 !important;
  padding: 32px 0 !important;
}

.planner-table__cell--danger {
  color: #ff4d4f !important;
  font-weight: 700;
}

.planner-table__status-cell {
  min-width: 130px;
}

.planner-review__conflict-tag {
  margin-inline-end: 6px;
  margin-bottom: 4px;
}

.planner-tag {
  display: inline-flex;
  align-items: center;
  min-height: 24px;
  padding: 0 10px;
  border-radius: 999px;
  background: #f0fdf4;
  color: #389e0d;
  font-size: 12px;
  font-weight: 700;
}

.planner-tag--warning {
  background: #fff1f0;
  color: #cf1322;
}

.planner-review__footer,
.planner-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.planner-footer__tip {
  color: #86909c;
  font-size: 12px;
  line-height: 1.7;
}

.planner-footer__actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.planner-option {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.planner-option__title {
  color: #1f2329;
  font-size: 14px;
  font-weight: 600;
}

.planner-option__meta {
  display: flex;
  gap: 12px;
  color: #98a2b3;
  font-size: 12px;
}

.planner-free-calendar__cell {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  border-radius: 8px;
}

.planner-free-calendar__cell--selected {
  background: #e6f4ff;
  color: #1677ff;
  font-weight: 700;
}

.planner-free-calendar__footer {
  display: flex;
  justify-content: flex-end;
}

.planner-shell {
  --planner-major-height: 42px;
  display: flex;
  max-height: calc(100vh - 170px);
  flex-direction: column;
  gap: 10px;
  padding: 10px 14px 10px;
  background: #f6f8fb;
}

.planner-strip,
.planner-card {
  border: 1px solid #e8edf3;
  border-radius: 16px;
  background: #fff;
  box-shadow: 0 8px 24px rgb(15 23 42 / 4%);
}

.planner-strip {
  padding: 14px 18px;
}

.planner-inline {
  justify-content: space-between;
}

.planner-inline--top {
  align-items: center;
  gap: 18px;
}

.planner-inline__group {
  gap: 12px;
}

.planner-inline__group--record {
  min-width: 0;
}

.planner-inline__group--status {
  min-width: 0;
  flex-shrink: 0;
}

.planner-label {
  color: #4e5969;
  white-space: nowrap;
}

.planner-label--inline {
  flex-shrink: 0;
  margin-right: 2px;
}

.planner-label--required::before {
  content: '*';
  color: #ff4d4f;
}

.planner-label--required::after {
  content: none;
}

.planner-label__tip {
  color: #9aa4b2;
  font-size: 14px;
}

:deep(.planner-control.ant-picker),
:deep(.planner-control .ant-select-selector),
:deep(.planner-control.ant-input-number) {
  min-height: 42px !important;
  border-color: #d9e1ea !important;
  border-radius: 12px !important;
  box-shadow: none !important;
}

:deep(.planner-control.ant-select-single .ant-select-selector) {
  display: flex;
  align-items: center;
}

:deep(.planner-control.ant-select-single .ant-select-selection-item),
:deep(.planner-control.ant-select-single .ant-select-selection-placeholder),
:deep(.planner-control.ant-select-single .ant-select-selection-search) {
  display: flex;
  align-items: center;
  min-height: 40px;
  font-size: 14px !important;
  font-weight: 400 !important;
}

:deep(.planner-control.ant-select-single .ant-select-selection-search-input) {
  height: 40px !important;
}

:deep(.planner-control.ant-picker .ant-picker-input > input),
:deep(.planner-control.ant-input-number .ant-input-number-input) {
  font-size: 14px !important;
  font-weight: 400 !important;
}

:deep(.planner-control .ant-select-selection-placeholder),
:deep(.planner-control .ant-select-selection-search-input) {
  font-size: 14px !important;
  font-weight: 400 !important;
}

.planner-balance {
  flex: 0 0 auto;
  gap: 6px;
  padding: 0;
  background: transparent;
}

.planner-balance__item {
  min-width: 100px;
  padding: 6px 10px;
  border: 1px solid #f90;
  border-radius: 10px;
  background: #fff5e6;
}

.planner-balance__item span,
.planner-balance__item strong {
  display: block;
}

.planner-balance__item span {
  font-size: 11px;
  line-height: 1.4;
}

.planner-balance__item strong {
  margin-top: 2px;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.45;
}

.planner-layout {
  grid-template-columns: 300px minmax(0, 1fr);
  gap: 14px;
  align-items: stretch;
  margin-top: 0;
}

.planner-aside,
.planner-main {
  display: flex;
  min-width: 0;
}

.planner-card {
  overflow: hidden;
}

.planner-card--summary,
.planner-card--form {
  display: flex;
  flex: 1;
  flex-direction: column;
  padding: 0;
}

.planner-card--empty-state {
  display: flex;
  min-height: 460px;
  align-items: center;
  justify-content: center;
  margin-top: 0;
  padding: 24px;
}

.planner-card__head {
  padding: 16px 18px 0;
}

.planner-card__title {
  font-size: 15px;
  font-weight: 600;
  line-height: 1.5;
}

.planner-card__subhead {
  margin: 0 18px 10px;
  padding-top: 14px;
  border-top: 1px solid #f1f4f8;
  font-size: 13px;
  font-weight: 600;
  line-height: 1.5;
}

.planner-profile {
  margin-top: 0;
  padding: 14px 18px 0;
  border-radius: 0;
  background: transparent;
}

.planner-profile__name {
  font-size: 18px;
  font-weight: 600;
  line-height: 1.4;
}

.planner-profile__meta {
  color: #86909c;
  line-height: 1.6;
}

.planner-summary-list {
  margin-top: 0;
  gap: 0;
  padding: 14px 18px 16px;
}

.planner-summary-list--secondary {
  padding-top: 0;
}

.planner-summary-list__row {
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px dashed #edf0f3;
  color: #86909c;
}

.planner-summary-list__row:last-child {
  border-bottom: none;
}

.planner-summary-list__row span {
  flex-shrink: 0;
  width: 72px;
  color: #86909c;
  font-size: 13px;
}

.planner-summary-list__value-wrap {
  display: block;
  flex: 1;
  text-align: right;
  cursor: default;
}

.planner-summary-list__value {
  display: block;
  min-width: 0;
  max-width: none;
  overflow: hidden;
  font-size: 12px;
  font-weight: 600;
  line-height: 1.6;
  text-align: right;
  text-overflow: ellipsis;
  white-space: nowrap;
  word-break: normal;
}

.planner-summary-list__value-wrap .planner-summary-list__value {
  width: 100%;
}

.planner-note,
.planner-alert {
  margin: 0 18px 18px;
  border-radius: 12px;
  border: none;
}

.planner-note {
  background: #fafbfc;
  color: #4e5969;
}

.planner-alert {
  background: #fff7e6;
  color: #d46b08;
}

.planner-alert--soft {
  background: #e8f4ff;
  color: #1664c0;
}

.planner-section {
  padding: 16px 18px 18px;
}

.planner-section + .planner-section {
  margin-top: 0;
  border-top: 1px solid #f1f4f8;
}

.planner-section__title {
  margin-bottom: 10px;
  font-size: 15px;
  font-weight: 600;
  line-height: 1.5;
}

.planner-choice-row {
  gap: 12px;
}

button.planner-choice {
  width: 100%;
  min-height: 58px;
  justify-content: center;
  gap: 4px;
  padding: 10px 12px;
  border: 1px solid #d8e1eb;
  border-radius: 14px;
  appearance: none;
  font: inherit;
  outline: none;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    box-shadow 0.2s ease;
}

button.planner-choice:hover {
  border-color: #91caff;
  background: #fff;
}

button.planner-choice.planner-choice--active {
  border-color: #1677ff;
  background: #f7fbff;
  box-shadow: 0 0 0 3px rgb(22 119 255 / 8%);
}

.planner-choice__title {
  font-weight: 600;
  line-height: 1.5;
}

.planner-choice__desc {
  color: #7d8898;
  font-size: 11px;
  line-height: 1.45;
}

button.planner-choice.planner-choice--active .planner-choice__title {
  color: #1677ff;
}

.planner-form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px 16px;
}

.planner-form-grid > .planner-field .planner-control {
  width: 100%;
}

:deep(.planner-form-grid > .planner-field .ant-select) {
  width: 100% !important;
}

.planner-date-plan-row {
  grid-column: 1 / -1;
  grid-template-columns: minmax(148px, 200px) minmax(100px, 132px) minmax(0, 1fr);
  column-gap: 14px;
  row-gap: 10px;
  align-items: start;
}

.planner-date-plan-row__cell {
  min-width: 0;
}

.planner-date-plan-row__cell--start {
  width: 100%;
  max-width: 200px;
}

.planner-date-plan-row__cell--start .planner-control {
  width: 100%;
  max-width: 200px;
  height: 42px;
}

:deep(.planner-date-plan-row__cell--start .ant-picker) {
  width: 100% !important;
  height: 42px !important;
  max-width: 200px !important;
}

.planner-date-plan-row__cell--count {
  width: 100%;
  min-width: 0;
}

.planner-date-plan-row__cell--count .planner-control {
  width: 100%;
  height: 42px;
}

:deep(.planner-date-plan-row__cell--count .ant-input-number) {
  width: 100% !important;
  height: 42px !important;
}

:deep(.planner-date-plan-row .ant-picker-input),
:deep(.planner-date-plan-row .ant-input-number-input-wrap),
:deep(.planner-date-plan-row .ant-input-number-handler-wrap) {
  height: 100%;
}

:deep(.planner-date-plan-row .ant-picker-input > input),
:deep(.planner-date-plan-row .ant-input-number-input) {
  height: 40px !important;
  line-height: 40px !important;
  padding-block: 0 !important;
}

.planner-date-plan-row__hint {
  grid-column: 1 / -1;
  margin-top: 0;
}

.planner-date-plan-row--free {
  grid-template-columns: minmax(0, 1fr);
  column-gap: 14px;
  row-gap: 10px;
}

.planner-date-plan-row--free .planner-date-plan-row__cell--start {
  max-width: none;
}

.planner-date-plan-row--free .planner-date-plan-row__cell--start .planner-control {
  max-width: none;
}

:deep(.planner-date-plan-row--free .planner-date-plan-row__cell--start .ant-picker) {
  max-width: none !important;
}

.planner-field {
  min-width: 0;
}

.planner-field--full {
  grid-column: 1 / -1;
  width: auto;
}

.planner-field--major {
  align-self: start;
}

.planner-field__hint {
  display: block;
  margin-top: 6px;
  color: #86909c;
  font-size: 12px;
  line-height: 1.55;
}

.planner-field__label-row--with-period-group {
  align-items: center;
  min-height: 32px;
  margin-bottom: 2px;
}

.planner-label--with-inline-tip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.planner-period-group-wrap {
  gap: 6px;
  flex-shrink: 0;
}

.planner-period-group-wrap__tip {
  cursor: help;
  color: #98a2b3;
}

.planner-period-group-radio--inline {
  flex-shrink: 0;
}

.planner-control-tooltip-wrap {
  display: block;
  width: 100%;
  min-width: 320px;
}

.planner-chip-row {
  gap: 8px;
  align-items: center;
}

.planner-chip-row--weekday {
  flex-wrap: nowrap;
  overflow-x: auto;
}

.planner-chip-actions {
  margin-left: 6px;
  flex-shrink: 0;
  white-space: nowrap;
}

button.planner-chip {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 68px;
  height: 34px;
  padding: 0 14px;
  border: 1px solid #d7dfe9;
  border-radius: 999px;
  background: #fff;
  color: #4e5969;
  font-size: 13px;
  font-weight: 500;
  appearance: none;
  font: inherit;
  outline: none;
  transition:
    color 0.2s ease,
    border-color 0.2s ease,
    background-color 0.2s ease,
    box-shadow 0.2s ease;
}

button.planner-chip:hover {
  border-color: #91caff;
  color: #1677ff;
}

button.planner-chip.planner-chip--active {
  border-color: #1677ff;
  background: #e6f4ff;
  color: #1677ff;
  box-shadow: 0 0 0 3px rgb(22 119 255 / 8%);
}

.planner-holiday-inline {
  gap: 8px;
  margin-left: 8px;
  padding: 3px 10px;
  border: 1px dashed #ffd591;
  border-radius: 12px;
  background: #fffaf0;
  flex-wrap: nowrap;
}

.planner-holiday-inline__label {
  flex-shrink: 0;
  color: #d46b08;
  font-weight: 600;
  line-height: 1.5;
}

.planner-holiday-inline__list {
  display: flex;
  gap: 8px;
}

.planner-holiday-inline__item {
  height: 26px;
  align-items: center;
  line-height: 1;
}

.planner-static-field {
  min-height: 42px;
  gap: 4px;
  padding: 10px 12px;
  border: 1px solid #e5ebf2;
  border-radius: 12px;
  background: #fafbfd;
}

.planner-static-field strong {
  font-weight: 600;
  line-height: 1.5;
}

.planner-static-field span {
  margin-top: 0;
  line-height: 1.5;
}

.planner-static-field--compact {
  min-height: 42px;
  justify-content: center;
  padding: 9px 12px;
}

.planner-static-field--compact strong {
  line-height: 1.35;
}

.planner-static-field--compact span {
  margin-top: 2px;
  font-size: 11px;
  line-height: 1.35;
}

.planner-static-field--inline-hint {
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.planner-static-field--inline-hint span {
  text-align: right;
  white-space: nowrap;
}

.planner-static-field--free-trigger strong {
  display: block;
  overflow: hidden;
  font-size: 13px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.planner-static-field--free-trigger span {
  font-size: 11px;
}

.planner-date-plan-row__cell .planner-static-field--inline-hint span {
  white-space: normal;
}

.planner-date-plan-row__cell .planner-static-field--inline-hint strong {
  white-space: nowrap;
  flex-shrink: 0;
}

.planner-free-calendar {
  width: 300px;
}

.planner-free-calendar__footer {
  justify-content: center;
  padding-top: 8px;
  border-top: 1px solid #f0f0f0;
}

.planner-free-calendar__cell {
  height: 34px;
  border-radius: 10px;
  color: #1f2329;
  transition: all 0.18s ease;
}

.planner-free-calendar__cell--selected {
  background: #1677ff;
  color: #fff;
  font-weight: 600;
}

:deep(.group-class-schedule-modal .ant-modal-content),
:deep(.planner-review-modal .ant-modal-content) {
  overflow: hidden;
  padding: 0;
  border-radius: 20px;
}

:deep(.group-class-schedule-modal .ant-modal-header) {
  margin-bottom: 0;
  padding: 14px 20px 10px;
  border-bottom: none;
}

:deep(.group-class-schedule-modal .ant-modal-body) {
  padding: 0;
}

:deep(.group-class-schedule-modal .ant-modal-footer) {
  padding: 14px 20px 16px;
  border-top: 1px solid #eef2f6;
  background: #fff;
}

:deep(.planner-review-modal .ant-modal-header) {
  margin-bottom: 0;
  padding: 18px 24px 14px;
  border-bottom: none;
}

:deep(.planner-review-modal .ant-modal-body) {
  padding: 0 24px 16px;
}

:deep(.planner-multi-slot-select.ant-select) {
  width: 100% !important;
  min-width: 320px !important;
}

:deep(.planner-multi-slot-select .ant-select-selector) {
  width: 100% !important;
  min-width: 0;
  flex-wrap: wrap !important;
  align-items: center !important;
  row-gap: 4px;
}

:deep(.planner-multi-slot-select .ant-select-selection-overflow) {
  flex-wrap: wrap !important;
  align-items: center !important;
}

:deep(.planner-multi-slot-select .ant-select-selection-overflow-item) {
  max-width: none !important;
}

:deep(.planner-multi-slot-select.ant-select-multiple .ant-select-selection-search) {
  display: inline-flex !important;
  align-items: center !important;
  margin-top: 8px;
  margin-bottom: 6px;
}

:deep(.planner-control--major.planner-multi-slot-select.ant-select-multiple .ant-select-selection-search) {
  margin-top: 4px;
  margin-bottom: 4px;
}

:deep(.planner-multi-slot-select.ant-select-multiple .ant-select-selection-search-input),
:deep(.planner-multi-slot-select.ant-select-multiple .ant-select-selection-search-mirror) {
  height: 28px !important;
  min-height: 28px !important;
  line-height: 28px !important;
}

:deep(.planner-control--major.ant-picker),
:deep(.planner-control--major .ant-select-selector) {
  height: var(--planner-major-height) !important;
  min-height: var(--planner-major-height) !important;
  padding-inline: 12px !important;
  border-radius: 12px !important;
}

:deep(.planner-multi-slot-select.ant-select-multiple .ant-select-selector) {
  height: auto !important;
  min-height: var(--planner-major-height) !important;
}

:deep(.planner-control--major.planner-multi-slot-select.ant-select-multiple .ant-select-selector) {
  padding-top: 2px !important;
  padding-bottom: 2px !important;
  row-gap: 2px;
}

:deep(.planner-control--major.ant-picker .ant-picker-input) {
  height: 100%;
}

:deep(.planner-control--major.ant-picker .ant-picker-input > input) {
  height: 100%;
  font-size: 14px;
  font-weight: 400;
}

:deep(.planner-control--major:not(.planner-multi-slot-select) .ant-select-selection-item) {
  display: flex;
  align-items: center;
  font-size: 14px;
  font-weight: 400;
  line-height: 1.45;
  white-space: normal;
}

:deep(.planner-multi-slot-select.ant-select-multiple .ant-select-selection-item) {
  display: inline-flex;
  align-items: center;
  min-height: 28px;
  height: auto;
  max-width: none !important;
  margin-top: 6px;
  margin-bottom: 6px;
  padding: 0 10px 0 12px;
  border: 1px solid #bfd8ff;
  border-radius: 8px;
  background: #eef6ff;
  color: #1677ff;
  font-size: 13px;
  font-weight: 400;
  white-space: nowrap !important;
}

:deep(.planner-control--major.planner-multi-slot-select.ant-select-multiple .ant-select-selection-item) {
  min-height: 26px;
  margin-top: 4px;
  margin-bottom: 4px;
}

:deep(.planner-multi-slot-select.ant-select-multiple .ant-select-selection-item-content) {
  max-width: none !important;
  overflow: visible !important;
  color: inherit;
  text-overflow: clip !important;
  white-space: nowrap !important;
}

:deep(.planner-multi-slot-select.ant-select-multiple .ant-select-selection-item-remove) {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  border-radius: 6px;
  color: #6da7ff;
  transition: all 0.2s ease;
}

:deep(.planner-multi-slot-select.ant-select-multiple .ant-select-selection-item-remove:hover) {
  background: rgb(22 119 255 / 10%);
  color: #1677ff;
}

:deep(.planner-multi-slot-select.ant-select-multiple .ant-select-selection-overflow-item-rest .ant-select-selection-item) {
  border-color: #d7dfe9;
  border-radius: 8px;
  background: #f7f8fa;
  color: #5f6b7c;
}

:deep(.planner-control--major .ant-select-arrow),
:deep(.planner-control--major.ant-picker .ant-picker-suffix) {
  color: #b0b8c5;
  font-size: 16px;
}

:deep(.group-class-schedule-modal .ant-select-focused .ant-select-selector),
:deep(.group-class-schedule-modal .ant-picker-focused) {
  border-color: #1677ff !important;
}

@media (max-width: 1080px) {
  .planner-layout {
    grid-template-columns: minmax(0, 1fr);
  }

  .planner-balance {
    flex-wrap: wrap;
  }
}

@media (max-width: 860px) {
  .planner-head,
  .planner-review__head,
  .planner-review__footer,
  .planner-footer,
  .planner-inline {
    flex-direction: column;
    align-items: stretch;
  }

  .planner-choice-row,
  .planner-form-grid,
  .planner-date-plan-row {
    grid-template-columns: minmax(0, 1fr);
  }

  .planner-date-plan-row__cell--start,
  .planner-date-plan-row__cell--start .planner-control {
    max-width: none;
  }

  :deep(.planner-date-plan-row__cell--start .ant-picker) {
    max-width: none !important;
  }

  .planner-inline__group {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
  }

  .planner-inline__group--record,
  .planner-inline__group--status {
    min-width: 0;
  }

  .planner-label--inline {
    white-space: normal;
  }

  .planner-footer__actions {
    justify-content: flex-end;
  }

  .planner-control-tooltip-wrap {
    min-width: 0;
  }
}
</style>

<style lang="less">
.planner-record-select-dropdown.ant-select-dropdown,
.planner-schedule-slot-select-dropdown.ant-select-dropdown,
.planner-assistant-select-dropdown.ant-select-dropdown {
  z-index: 3000 !important;
}

.planner-schedule-slot-select-dropdown .ant-select-item,
.planner-assistant-select-dropdown .ant-select-item {
  padding: 0;
}

.planner-schedule-slot-select-dropdown .ant-select-item-option,
.planner-assistant-select-dropdown .ant-select-item-option {
  min-height: auto;
  padding: 0;
}

.planner-schedule-slot-select-dropdown .ant-select-item-option-content,
.planner-assistant-select-dropdown .ant-select-item-option-content {
  padding: 12px 14px;
}

.planner-schedule-slot-select-dropdown .ant-select-item-option + .ant-select-item-option,
.planner-assistant-select-dropdown .ant-select-item-option + .ant-select-item-option {
  border-top: 1px solid #f2f4f7;
}

.planner-schedule-slot-select-dropdown .ant-select-item-option-active:not(.ant-select-item-option-disabled) .ant-select-item-option-content,
.planner-assistant-select-dropdown .ant-select-item-option-active:not(.ant-select-item-option-disabled) .ant-select-item-option-content {
  background: #f8fbff;
}

.planner-schedule-slot-select-dropdown .ant-select-item-option-selected:not(.ant-select-item-option-disabled) .ant-select-item-option-content,
.planner-assistant-select-dropdown .ant-select-item-option-selected:not(.ant-select-item-option-disabled) .ant-select-item-option-content {
  background: #edf5ff;
}

.planner-schedule-slot-select-dropdown .ant-select-item-option-state,
.planner-assistant-select-dropdown .ant-select-item-option-state {
  display: none;
}

.planner-slot-option,
.planner-staff-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.planner-slot-option__label,
.planner-staff-option__label {
  min-width: 0;
  color: #1f2329;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.45;
}

.planner-staff-option__main {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 2px;
}

.planner-staff-option__mobile {
  color: #98a2b3;
  font-size: 12px;
  line-height: 1.4;
}

.planner-slot-option__status {
  display: inline-flex;
  flex-shrink: 0;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 600;
  line-height: 1;
}

.planner-slot-option__status::before {
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: currentColor;
  content: '';
}

.planner-slot-option__status--free {
  color: #52c41a;
}

.planner-slot-option__status--busy {
  color: #ff7a45;
}

.planner-slot-option__status--unknown {
  color: #98a2b3;
}

.planner-record-select-dropdown {
  padding: 8px;
}

.planner-record-select-dropdown .ant-select-item {
  padding: 0;
}

.planner-record-select-dropdown .ant-select-item-option {
  min-height: auto;
  margin-bottom: 6px;
  padding: 0;
  border: 1px solid transparent;
  border-radius: 12px;
  background: #fff;
  transition:
    border-color 0.18s ease,
    background-color 0.18s ease,
    box-shadow 0.18s ease;
}

.planner-record-select-dropdown .ant-select-item-option:last-child {
  margin-bottom: 0;
}

.planner-record-select-dropdown .ant-select-item-option-content {
  padding: 10px 12px;
}

.planner-record-select-dropdown .ant-select-item-option-active:not(.ant-select-item-option-disabled) {
  border-color: #dbeafe;
  background: #f8fbff;
}

.planner-record-select-dropdown .ant-select-item-option-selected:not(.ant-select-item-option-disabled) {
  border-color: #bfd8ff;
  background: linear-gradient(180deg, #f7fbff 0%, #eef6ff 100%);
  box-shadow: inset 0 0 0 1px rgb(22 119 255 / 10%);
}

.planner-record-select-dropdown .planner-option {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.planner-record-select-dropdown .planner-option__title {
  overflow: hidden;
  color: #1f2329;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.45;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.planner-record-select-dropdown .planner-option__meta {
  display: flex;
  gap: 10px;
  overflow: hidden;
  color: #7d8898;
  font-size: 12px;
  line-height: 1.5;
}
</style>
