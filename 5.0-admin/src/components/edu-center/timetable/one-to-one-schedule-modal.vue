<script setup lang="ts">
import {
  BookOutlined,
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
import ScheduleConflictWorkbenchModal from './schedule-conflict-workbench-modal.vue'
import { type ClassroomItem, listClassroomsApi } from '@/api/business-settings/classroom'
import { getUserListApi } from '@/api/internal-manage/staff-manage'
import StaffSelect from '@/components/common/staff-select.vue'
import { type OneToOneItem, getOneToOneListApi } from '@/api/edu-center/one-to-one'
import type { TeachingScheduleValidationResult } from '@/api/edu-center/teaching-schedule'
import { checkAssistantScheduleAvailabilityApi, checkOneToOneScheduleAvailabilityApi, createOneToOneSchedulesApi, validateOneToOneSchedulesApi } from '@/api/edu-center/teaching-schedule'
import messageService from '@/utils/messageService'
import emitter, { EVENTS } from '@/utils/eventBus'
import { useUserStore } from '@/stores/user'
import {
  DEFAULT_UNIFIED_TIME_PERIOD_CONFIG,
  buildQuickHourlySlots,
  configGroupsSorted,
  parseUnifiedTimePeriodConfig,
} from '@/utils/unified-time-period'

type PreviewTone = 'pending' | 'blocked'
type ScheduleType = 'oneToOne' | 'studentLesson'
type SchedulingMode = 'repeat' | 'free'
type RepeatRule = 'none' | 'weekly' | 'biweekly' | 'daily' | 'alternateDay'
type HolidayPolicy = 'include' | 'filter'
type PeriodGroupKey = 'A' | 'B'

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
  allowClassroomConflict?: boolean
  tone: PreviewTone
}

interface SummaryItem {
  label: string
  value: string
}

interface QuickTag {
  text: string
  tone: 'primary' | 'warning' | 'neutral'
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
  mobile?: string
  status: 'free' | 'busy' | 'unknown'
  statusText: string
}

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
}>()

const modalOpen = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

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

const scheduleTypeOptions: OptionItem<ScheduleType>[] = [
  { value: 'oneToOne', label: '按1对1', desc: '从现有1对1档案直接带出排课信息' },
  { value: 'studentLesson', label: '按学员和课程', desc: '按学员与课程组合查看待创建对象' },
]

const schedulingModeOptions: OptionItem<SchedulingMode>[] = [
  { value: 'repeat', label: '重复排课', desc: '批量生成固定周期日程' },
  { value: 'free', label: '自由排课', desc: '先创建单次日程，后续再补排' },
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

const periodConfig = computed(() => {
  const parsed = parseUnifiedTimePeriodConfig(userStore.instConfig?.unifiedTimePeriodJson)
  return parsed ?? DEFAULT_UNIFIED_TIME_PERIOD_CONFIG
})

const sortedPeriodGroups = computed(() => configGroupsSorted(periodConfig.value))

const groupOptions = computed(() => {
  const g = sortedPeriodGroups.value
  if (!g.length)
    return [{ key: 'A' as const, label: '默认时段' }]
  if (g.length === 1)
    return [{ key: 'A' as const, label: g[0].name || '时段' }]
  return [
    { key: 'A' as const, label: g[0].name || 'A时段' },
    { key: 'B' as const, label: g[1].name || 'B时段' },
  ]
})

function slotsForGroupKey(key: PeriodGroupKey) {
  const groups = sortedPeriodGroups.value
  const fallback = buildQuickHourlySlots().filter(s => s.enabled !== false)
  if (!groups.length)
    return [...fallback].sort((a, b) => a.index - b.index)
  const idx = key === 'B' ? 1 : 0
  const g = groups[idx] || groups[0]
  return [...g.slots].filter(s => s.enabled !== false).sort((a, b) => a.index - b.index)
}

function periodGroupIndexForKey(key: PeriodGroupKey): number {
  return key === 'B' ? 1 : 0
}

/** 该组未配置关联老师时，任意老师可选；配置后仅列表内老师可选 */
function isTeacherAllowedInPeriodGroup(teacherIdStr: string, groupIndex: number): boolean {
  const groups = sortedPeriodGroups.value
  const g = groups[groupIndex]
  if (!g)
    return true
  const list = g.boundTeachers
  if (!Array.isArray(list) || list.length === 0)
    return true
  if (!teacherIdStr)
    return true
  return list.some(t => String(t.id ?? '').trim() === teacherIdStr)
}

/** 助教跟随当前时段组筛选；当前组未配置关联老师时，仍允许全部老师可选 */
function isAssistantAllowedInCurrentGroup(assistantIdStr: string): boolean {
  return isTeacherAllowedInPeriodGroup(assistantIdStr, periodGroupIndexForKey(currentGroup.value))
}

const classroomList = ref<ClassroomItem[]>([])
const workbenchTeacherList = ref<StaffOptionItem[]>([])
const oneToOneRecords = ref<OneToOneItem[]>([])
const oneToOneLoading = ref(false)
const selectedOneToOneId = ref<string | undefined>(undefined)
const scheduleType = ref<ScheduleType>('oneToOne')
const schedulingMode = ref<SchedulingMode>('repeat')
const repeatRule = ref<RepeatRule>('weekly')
const holidayPolicy = ref<HolidayPolicy>('filter')
const currentGroup = ref<PeriodGroupKey>('A')
const selectedWeekdays = ref(['周一', '周三', '周五'])
/** 同一选课可多次勾选（例如上午一节 + 下午一节），每个上课日按所选时段各生成一节 */
const selectedSchoolTimeSlots = ref<string[]>([])

const schoolTimeSlotOptions = computed<SchoolTimeSlot[]>(() => {
  const slots = slotsForGroupKey(currentGroup.value)
  return slots.map(s => ({
    value: `period-${s.index}`,
    label: `第${s.index}节课`,
    desc: `第${s.index}节课`,
    start: String(s.start || '').slice(0, 5),
    end: String(s.end || '').slice(0, 5),
  }))
})

const selectedTeacher = ref<string | number | undefined>(undefined)
const selectedTeacherDisplay = ref<StaffOptionItem | null>(null)
const selectedAssistant = ref<Array<string | number> | undefined>(undefined)
const selectedAssistantDisplays = ref<StaffOptionItem[]>([])
const selectedClassroom = ref<string | undefined>(undefined)
const selectedStudentLessonPath = ref<string[] | undefined>(undefined)
const previewModalOpen = ref(false)
const creatingSchedules = ref(false)
const previewValidating = ref(false)
const previewValidationMessage = ref('')
const previewHasConflict = ref(false)
const conflictModalOpen = ref(false)
const previewValidationResult = ref<TeachingScheduleValidationResult | null>(null)
const creatingWithSoftConflict = ref(false)
const plannerShellRef = ref<HTMLElement | null>(null)
const teacherSlotAvailabilityMap = ref<Record<string, AvailabilityBadgeView>>({})
const teacherSlotAvailabilityLoading = ref(false)
const assistantAvailabilityMap = ref<Record<string, AvailabilityBadgeView>>({})
const assistantAvailabilityLoading = ref(false)
let teacherSlotAvailabilitySeq = 0
let assistantAvailabilitySeq = 0
let teacherSlotAvailabilityTimer: ReturnType<typeof setTimeout> | null = null
let assistantAvailabilityTimer: ReturnType<typeof setTimeout> | null = null

function scrollPlannerShellToTop() {
  const el = plannerShellRef.value
  if (el)
    el.scrollTop = 0
}

/** 课表节次下拉的容器：挂 body + 提高 z-index，避免滚动的 planner-shell / Modal 误判「点外部」而关窗 */
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

function scheduleAvailabilityBadge(status: 'free' | 'busy' | 'unknown', statusText: string): AvailabilityBadgeView {
  return { status, statusText }
}

function displayMobileText(staff?: StaffOptionItem | null) {
  return String(staff?.mobile || '').trim()
}

const scheduleStartDate = ref(dayjs().startOf('day'))
const freeSelectedDates = ref<Dayjs[]>([dayjs().startOf('day')])
const freeCalendarPanelDate = ref(dayjs().startOf('month'))
const freeCalendarOpen = ref(false)
const plannedClassCount = ref(1)

const selectedOneToOne = computed(() =>
  oneToOneRecords.value.find(item => item.id === selectedOneToOneId.value),
)

const selectedAssistantValues = computed<Array<string | number>>(() =>
  Array.isArray(selectedAssistant.value) ? selectedAssistant.value : [],
)

const oneToOneSelectOptions = computed(() =>
  oneToOneRecords.value.map(item => ({
    value: item.id || '',
    label: `${item.name || '-'} · 课程：${item.lessonName || '-'}`,
  })),
)

const studentLessonCascaderOptions = computed(() => {
  const grouped = new Map<string, { value: string, label: string, children: { value: string, label: string }[] }>()

  oneToOneRecords.value.forEach((item) => {
    const studentValue = item.studentId || `student-${item.id || item.name || Math.random()}`
    if (!grouped.has(studentValue)) {
      grouped.set(studentValue, {
        value: studentValue,
        label: item.studentName || '-',
        children: [],
      })
    }

    grouped.get(studentValue)?.children.push({
      value: item.id || '',
      label: `课程：${item.lessonName || '-'}`,
    })
  })

  return [...grouped.values()]
})

function isValidStaffId(value: unknown) {
  const text = String(value ?? '').trim()
  return text !== '' && text !== '0' && text !== 'undefined' && text !== 'null'
}

/** 与列表页一致：接口常用 0 / "0" 表示未设置教室，不应作为可选值展示 */
function isValidClassroomId(value: unknown) {
  const text = String(value ?? '').trim()
  return text !== '' && text !== '0' && text !== 'undefined' && text !== 'null'
}

function sameStaffId(a: unknown, b: unknown) {
  return isValidStaffId(a) && isValidStaffId(b) && String(a) === String(b)
}

const selectedTeacherIdNormalized = computed(() =>
  isValidStaffId(selectedTeacher.value) ? String(selectedTeacher.value).trim() : '',
)

const eligiblePeriodGroupKeys = computed<PeriodGroupKey[]>(() => {
  const opts = groupOptions.value
  const tid = selectedTeacherIdNormalized.value
  const keys: PeriodGroupKey[] = []
  for (const opt of opts) {
    const k = opt.key as PeriodGroupKey
    if (isTeacherAllowedInPeriodGroup(tid, periodGroupIndexForKey(k)))
      keys.push(k)
  }
  return keys
})

function isPeriodGroupChoiceDisabled(key: PeriodGroupKey): boolean {
  const elig = eligiblePeriodGroupKeys.value
  if (!elig.length)
    return false
  return !elig.includes(key)
}

function displayStaffName(staff?: StaffOptionItem | null) {
  return staff?.nickName || staff?.name || ''
}

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

  selectedOneToOne.value?.teacherList?.forEach(item => append(item.teacherId, item.name))
  append(selectedOneToOne.value?.defaultTeacherId, selectedOneToOne.value?.defaultTeacherName)
  return records
})

const scheduleStaffSelectKey = computed(() => selectedOneToOne.value?.id || 'empty')

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
  return [...classroomSet.values()]
})

const normalizedSelectedClassroomId = computed(() => {
  const current = String(selectedClassroom.value || '').trim()
  if (!current)
    return ''
  return classroomOptions.value.some(item => item.value === current) ? current : ''
})

watch(classroomOptions, (options) => {
  const current = String(selectedClassroom.value || '').trim()
  if (!current)
    return
  if (!options.some(item => item.value === current))
    selectedClassroom.value = undefined
}, { immediate: true })

watch(
  groupOptions,
  (opts) => {
    if (!opts.some(o => o.key === currentGroup.value))
      currentGroup.value = 'A'
  },
  { immediate: true },
)

watch(
  eligiblePeriodGroupKeys,
  () => {
    const elig = eligiblePeriodGroupKeys.value
    if (elig.length === 1) {
      currentGroup.value = elig[0]
      return
    }
    if (elig.length > 1 && !elig.includes(currentGroup.value))
      currentGroup.value = elig[0]
  },
  { flush: 'post', immediate: true },
)

watch(
  schoolTimeSlotOptions,
  (opts) => {
    if (!opts.length) {
      if (selectedSchoolTimeSlots.value.length)
        selectedSchoolTimeSlots.value = []
      return
    }
    const valid = new Set(opts.map(s => s.value))
    const next = selectedSchoolTimeSlots.value.filter(v => valid.has(v))
    if (!scheduleSlotKeysEqual(next, selectedSchoolTimeSlots.value))
      selectedSchoolTimeSlots.value = next
  },
  { immediate: true },
)

watch(
  selectedOneToOne,
  (value) => {
    const defaultTeacherId = isValidStaffId(value?.defaultTeacherId)
      ? value?.defaultTeacherId
      : teacherPresetStaff.value[0]?.id
    selectedTeacher.value = defaultTeacherId
    selectedTeacherDisplay.value = teacherPresetStaff.value.find(item => sameStaffId(item.id, defaultTeacherId)) || null
    selectedClassroom.value = value && isValidClassroomId(value.classRoomId)
      ? String(value.classRoomId).trim()
      : undefined
    selectedAssistant.value = undefined
    selectedAssistantDisplays.value = []
    selectedSchoolTimeSlots.value = []
    teacherSlotAvailabilityMap.value = {}
    assistantAvailabilityMap.value = {}
    scheduleStartDate.value = dayjs().startOf('day')
    freeSelectedDates.value = [dayjs().startOf('day')]
    freeCalendarPanelDate.value = freeSelectedDates.value[0].startOf('month')
  },
  { immediate: true },
)

watch(selectedOneToOneId, (value) => {
  const current = oneToOneRecords.value.find(item => item.id === value)
  if (!current?.studentId || !current?.id) {
    selectedStudentLessonPath.value = undefined
    plannedClassCount.value = 1
    return
  }
  selectedStudentLessonPath.value = [current.studentId, current.id]
  const ta = current?.tuitionAccount
  const remain = Number(ta?.remainQuantity || 0) + Number(ta?.remainFreeQuantity || 0)
  plannedClassCount.value = remain > 0 ? Math.max(1, Math.ceil(remain)) : 1
}, { immediate: true })

watch(modalOpen, async (value) => {
  if (value) {
    selectedOneToOneId.value = undefined
    selectedStudentLessonPath.value = undefined
    await nextTick()
    scrollPlannerShellToTop()
    await Promise.all([
      userStore.getInstConfig(),
      fetchOneToOneRecords(),
      fetchClassroomList(),
      fetchWorkbenchTeacherList(),
    ])
    await nextTick()
    scrollPlannerShellToTop()
    requestAnimationFrame(() => scrollPlannerShellToTop())
  }
  if (!value) {
    previewModalOpen.value = false
    teacherSlotAvailabilityMap.value = {}
    assistantAvailabilityMap.value = {}
  }
})

const oneToOneStatusText = computed(() =>
  Number(selectedOneToOne.value?.status) === 2 ? '已结班' : '开班中',
)

const oneToOneStudentStatusText = computed(() => {
  const status = Number(selectedOneToOne.value?.classStudentStatus)
  if (status === 2)
    return '停课中'
  if (status === 3)
    return '已结课'
  return '开课中'
})

const recordModeText = computed(() =>
  Number(selectedOneToOne.value?.defaultClassTimeRecordMode) === 2 ? '按上课时长记录' : '按固定课时记录',
)

const effectiveLessonChargingMode = computed(() => {
  const ta = selectedOneToOne.value?.tuitionAccount
  const mode = Number(ta?.lessonChargingMode || 0)
  if (mode > 0)
    return mode
  if (ta?.enableExpireTime && Number(ta?.totalQuantity || 0) > 0)
    return 2
  return mode || 0
})

const quantityUnit = computed(() => {
  if (effectiveLessonChargingMode.value === 1)
    return '课时'
  if (effectiveLessonChargingMode.value === 2)
    return '天'
  if (effectiveLessonChargingMode.value === 3)
    return '元'
  return ''
})

function formatBalanceValue(value: number) {
  if (!Number.isFinite(value))
    return '--'
  return `${value.toFixed(2)}${quantityUnit.value ? ` ${quantityUnit.value}` : ''}`
}

const recordClassroomText = computed(() => selectedOneToOne.value?.classRoomName || '-')
const scheduledClassroomText = computed(() => {
  const current = classroomOptions.value.find(item => item.value === selectedClassroom.value)
  return current?.label || '-'
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
  const append = (item?: StaffOptionItem | null) => {
    if (!isValidStaffId(item?.id))
      return
    const id = String(item.id).trim()
    if (!isAssistantAllowedInCurrentGroup(id))
      return
    if (!merged.has(id)) {
      merged.set(id, {
        id,
        name: displayStaffName(item),
        nickName: displayStaffName(item) || id,
        mobile: displayMobileText(item),
        status: item?.status,
      })
    }
  }
  selectedAssistantDisplays.value.forEach(item => append(item))
  workbenchTeacherList.value.forEach(item => append(item))
  return [...merged.values()]
})

const conflictWorkbenchPeriodGroups = computed(() =>
  groupOptions.value.map((opt) => {
    const group = sortedPeriodGroups.value[periodGroupIndexForKey(opt.key)]
    const teacherIds = Array.isArray(group?.boundTeachers)
      ? group.boundTeachers.map(item => String(item.id ?? '').trim()).filter(Boolean)
      : []
    return {
      key: opt.key,
      label: opt.label,
      teacherIds,
      timeOptions: slotsForGroupKey(opt.key).map(slot => ({
        value: `${opt.key}|${slot.start}|${slot.end}`,
        label: `第${slot.index}节课 · ${slot.start}-${slot.end}`,
        startTime: slot.start,
        endTime: slot.end,
      })),
    }
  }),
)
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

const usedQuantityValue = computed(() => {
  if (!selectedOneToOne.value?.id)
    return Number.NaN
  const ta = selectedOneToOne.value?.tuitionAccount
  const total = Number(ta?.totalQuantity || 0) + Number(ta?.totalFreeQuantity || 0)
  const remain = Number(ta?.remainQuantity || 0) + Number(ta?.remainFreeQuantity || 0)
  return Math.max(total - remain, 0)
})

const remainQuantityValue = computed(() => {
  if (!selectedOneToOne.value?.id)
    return Number.NaN
  const ta = selectedOneToOne.value?.tuitionAccount
  return Number(ta?.remainQuantity || 0) + Number(ta?.remainFreeQuantity || 0)
})

/** 剩余可排课节数上限（与学费账户粒度一致，向上取整） */
const remainSessionCap = computed(() => {
  const r = remainQuantityValue.value
  return r > 0 ? Math.ceil(r) : 0
})

function slotDurationMinutes(start: string, end: string) {
  const [sh, sm] = start.split(':').map(Number)
  const [eh, em] = end.split(':').map(Number)
  return Math.max(0, eh * 60 + em - (sh * 60 + sm))
}

const activeTimeBlocks = computed<TimeBlock[]>(() => {
  const rows: (TimeBlock & { sortKey: string })[] = []
  for (const id of selectedSchoolTimeSlots.value) {
    const slot = schoolTimeSlotOptions.value.find(s => s.value === id)
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
  return rows
    .sort((a, b) => a.sortKey.localeCompare(b.sortKey))
    .map(({ sortKey: _s, ...r }) => r)
})

function slotBaseLabel(slot: SchoolTimeSlot) {
  return `${slot.desc} · ${slot.start} - ${slot.end}`
}

function teacherSlotAvailabilityFor(slot: SchoolTimeSlot): AvailabilityBadgeView {
  if (!isValidStaffId(selectedTeacher.value))
    return scheduleAvailabilityBadge('unknown', '先选老师')
  if (!plannedDates.value.length)
    return scheduleAvailabilityBadge('unknown', '待定')
  return teacherSlotAvailabilityMap.value[slot.value]
    || scheduleAvailabilityBadge(teacherSlotAvailabilityLoading.value ? 'unknown' : 'free', teacherSlotAvailabilityLoading.value ? '检测中' : '空闲')
}

const schoolTimeSlotOptionViews = computed<SlotSelectOptionView[]>(() =>
  schoolTimeSlotOptions.value.map((slot) => {
    const availability = teacherSlotAvailabilityFor(slot)
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

async function fetchTeacherSlotAvailability() {
  const seq = ++teacherSlotAvailabilitySeq
  const oneToOneId = String(selectedOneToOne.value?.id || '').trim()
  const teacherId = selectedTeacherIdNormalized.value
  if (!oneToOneId || !teacherId || !plannedDates.value.length || !schoolTimeSlotOptions.value.length) {
    teacherSlotAvailabilityMap.value = {}
    teacherSlotAvailabilityLoading.value = false
    return
  }

  const schedules = plannedDates.value.flatMap(date =>
    schoolTimeSlotOptions.value.map(slot => ({
      teacherId,
      lessonDate: date.format('YYYY-MM-DD'),
      startTime: slot.start,
      endTime: slot.end,
    })),
  )
  if (!schedules.length || schedules.length > 2000) {
    teacherSlotAvailabilityMap.value = {}
    teacherSlotAvailabilityLoading.value = false
    return
  }

  teacherSlotAvailabilityLoading.value = true
  try {
    const res = await checkOneToOneScheduleAvailabilityApi({
      oneToOneId,
      schedules,
    })
    if (seq !== teacherSlotAvailabilitySeq)
      return
    if (res.code !== 200 || !res.result)
      throw new Error(res.message || '检测老师节次空闲状态失败')

    const nextMap: Record<string, AvailabilityBadgeView> = {}
    schoolTimeSlotOptions.value.forEach((slot) => {
      nextMap[slot.value] = scheduleAvailabilityBadge('free', '空闲')
    })
    ;(res.result.items || []).forEach((item) => {
      if (item.valid !== false)
        return
      const matched = schoolTimeSlotOptions.value.find(slot => slot.start === item.startTime && slot.end === item.endTime)
      if (!matched)
        return
      nextMap[matched.value] = scheduleAvailabilityBadge('busy', '繁忙')
    })
    teacherSlotAvailabilityMap.value = nextMap
  }
  catch (error: any) {
    if (seq !== teacherSlotAvailabilitySeq)
      return
    console.error('fetchTeacherSlotAvailability failed', error)
    teacherSlotAvailabilityMap.value = {}
  }
  finally {
    if (seq === teacherSlotAvailabilitySeq)
      teacherSlotAvailabilityLoading.value = false
  }
}

function scheduleTeacherSlotAvailabilityCheck() {
  if (teacherSlotAvailabilityTimer)
    clearTimeout(teacherSlotAvailabilityTimer)
  teacherSlotAvailabilityTimer = setTimeout(() => {
    void fetchTeacherSlotAvailability()
  }, 180)
}

function assistantAvailabilityFor(staff: StaffOptionItem): AvailabilityBadgeView {
  if (!activeTimeBlocks.value.length)
    return scheduleAvailabilityBadge('unknown', '先选时段')
  return assistantAvailabilityMap.value[String(staff.id)]
    || scheduleAvailabilityBadge(assistantAvailabilityLoading.value ? 'unknown' : 'free', assistantAvailabilityLoading.value ? '检测中' : '空闲')
}

const assistantSelectOptionViews = computed<AssistantSelectOptionView[]>(() =>
  assistantSelectStaffs.value
    .filter(staff => !sameStaffId(staff.id, selectedTeacher.value))
    .map((staff) => {
      const availability = assistantAvailabilityFor(staff)
      return {
        value: String(staff.id),
        label: displayStaffName(staff) || String(staff.id),
        mobile: displayMobileText(staff),
        status: availability.status,
        statusText: availability.statusText,
      }
    }),
)

const assistantOptionList = computed(() =>
  assistantSelectStaffs.value
    .filter(item => !sameStaffId(item.id, selectedTeacher.value))
    .map(item => ({
      value: String(item.id),
      label: displayStaffName(item) || String(item.id),
    })),
)

async function fetchAssistantAvailability() {
  const seq = ++assistantAvailabilitySeq
  const oneToOneId = String(selectedOneToOne.value?.id || '').trim()
  const assistantIds = assistantSelectOptionViews.value.map(item => item.value)
  if (!oneToOneId || !assistantIds.length || !plannedDates.value.length || !activeTimeBlocks.value.length) {
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
    const res = await checkAssistantScheduleAvailabilityApi({
      oneToOneId,
      assistantIds,
      schedules,
    })
    if (seq !== assistantAvailabilitySeq)
      return
    if (res.code !== 200 || !res.result)
      throw new Error(res.message || '检测助教空闲状态失败')

    const nextMap: Record<string, AvailabilityBadgeView> = {}
    assistantIds.forEach((id) => {
      nextMap[id] = scheduleAvailabilityBadge('free', '空闲')
    })
    ;(res.result.items || []).forEach((item) => {
      nextMap[item.assistantId] = scheduleAvailabilityBadge(item.valid ? 'free' : 'busy', item.valid ? '空闲' : '繁忙')
    })
    assistantAvailabilityMap.value = nextMap
  }
  catch (error: any) {
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

function scheduleAssistantAvailabilityCheck() {
  if (assistantAvailabilityTimer)
    clearTimeout(assistantAvailabilityTimer)
  assistantAvailabilityTimer = setTimeout(() => {
    void fetchAssistantAvailability()
  }, 180)
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

const scheduleSessionMinutesText = computed(() => {
  const blocks = activeTimeBlocks.value
  if (!blocks.length)
    return '--'
  const uniq = [...new Set(blocks.map(b => `${b.minutes} 分钟`))]
  return uniq.join('、')
})

const selectedWeekdaysText = computed(() => selectedWeekdays.value.join(' / '))
const freeSelectedDatesSorted = computed(() =>
  [...freeSelectedDates.value].sort((a, b) => a.valueOf() - b.valueOf()),
)

const freeSelectedDateKeys = computed(() =>
  new Set(freeSelectedDatesSorted.value.map(item => item.format('YYYY-MM-DD'))),
)

const freeSelectedDatesText = computed(() => {
  const dates = freeSelectedDatesSorted.value.map(item => item.format('YYYY-MM-DD'))
  if (dates.length === 0)
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
  if (schedulingMode.value === 'repeat')
    return planned
  return 0
})

const repeatRuleText = computed(() => {
  if (schedulingMode.value === 'free')
    return '自由排课 · 单次日程'
  const base = repeatRuleLabelMap[repeatRule.value]
  if (repeatRule.value === 'weekly' || repeatRule.value === 'biweekly')
    return `${base} · ${selectedWeekdaysText.value || '-'}`
  return base
})

const timeModeText = computed(() => {
  const blocks = activeTimeBlocks.value
  const groupLabel = groupOptions.value.find(o => o.key === currentGroup.value)?.label || ''
  const prefix = groupLabel ? `${groupLabel} · ` : ''
  if (!blocks.length)
    return `${prefix}请选择课表节次`
  const blocksDesc = blocks.map((b) => {
    const statusText = selectedTimeBlockStatusText(b)
    return `${b.rangeText} · ${statusText}`
  }).join('；')
  const n = blocks.length
  const head = n > 1 ? `课表节次（共 ${n} 节）` : '课表节次'
  return `${prefix}${head} · ${blocksDesc}`
})

const selectedOneToOneSummary = computed<SummaryItem[]>(() => [
  { label: '学员', value: selectedOneToOne.value?.studentName || '-' },
  { label: '课程', value: selectedOneToOne.value?.lessonName || '-' },
  { label: '默认老师', value: selectedOneToOne.value?.defaultTeacherName || '-' },
  { label: '班主任', value: selectedOneToOne.value?.classTeacherName || '-' },
  { label: '默认教室', value: recordClassroomText.value },
  { label: '本次时段', value: scheduleSessionMinutesText.value },
  { label: '记录方式', value: recordModeText.value },
  { label: '学费账户', value: `${selectedOneToOne.value?.tuitionAccountCount ?? 0} 个` },
])

/** 侧边「创建摘要」里 value 会被单行省略，这些行悬停展示全文 */
const overviewTooltipLabels = new Set(['重复规则', '计划上课次数', '已选日期', '上课时间', '上课教室'])

const planExceedsRemainHint = computed(() => {
  if (schedulingMode.value === 'free') {
    const cap = remainSessionCap.value
    const planned = scheduleTargetCount.value
    if (planned < 1)
      return ''
    if (cap <= 0)
      return '当前剩余可排为 0，已选日期仅用于预估；正式创建前请处理账户。'
    if (planned > cap)
      return `当前已选 ${planned} 节，多于剩余可排 ${cap} 节；正式创建时请核对账户或减少日期。`
    return ''
  }
  const cap = remainSessionCap.value
  const planned = Math.floor(Number(plannedClassCount.value) || 0)
  if (planned < 1)
    return ''
  if (cap <= 0)
    return '当前剩余可排为 0，仍可按手填节数推算结束日期；正式创建前请处理账户。'
  if (planned > cap)
    return `当前手填 ${planned} 节，多于剩余可排 ${cap} 节；结束日期仍按手填次数推算，正式创建时请核对账户/是否截断。`
  return ''
})

const recordQuickTags = computed<QuickTag[]>(() => [
  { text: oneToOneStatusText.value, tone: 'primary' },
  {
    text: oneToOneStudentStatusText.value,
    tone: oneToOneStudentStatusText.value === '开课中' ? 'neutral' : 'warning',
  },
])

const blockedReason = computed(() => {
  if (oneToOneLoading.value)
    return '1对1数据加载中，请稍候。'
  if (!oneToOneRecords.value.length)
    return '暂无可用的1对1档案，请先创建1对1后再排课。'
  if (!selectedOneToOne.value?.id)
    return '请选择1对1后继续排课。'
  if (Number(selectedOneToOne.value?.status) === 2)
    return '当前 1 对 1 已结班，暂不可创建新日程。'
  if (Number(selectedOneToOne.value?.classStudentStatus) === 2)
    return '当前学员处于停课中，恢复后再创建日程。'
  if (Number(selectedOneToOne.value?.classStudentStatus) === 3)
    return '当前学员已结课，暂不可创建新日程。'
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

  const slotsPerDay = Math.max(activeTimeBlocks.value.length, 1)
  const dateTarget = Math.max(1, Math.ceil(sessionCap / slotsPerDay))

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

const filteredHolidayDates = computed(() => {
  if (holidayPolicy.value !== 'filter')
    return []
  return rawPlannedDates.value.filter(date => isHoliday(date))
})

const filteredHolidayDateLabels = computed(() =>
  filteredHolidayDates.value.map(item => item.format('YYYY-MM-DD')),
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

const dateSettingText = computed(() => {
  if (schedulingMode.value === 'free')
    return freeSelectedDatesText.value
  return rangeText.value
})

const overviewItems = computed<SummaryItem[]>(() => [
  { label: '排课类型', value: scheduleType.value === 'oneToOne' ? '按1对1' : '按学员和课程' },
  { label: '排课方式', value: schedulingMode.value === 'repeat' ? '重复排课' : '自由排课' },
  { label: '日期设置', value: dateSettingText.value },
  { label: '重复规则', value: repeatRuleText.value },
  { label: '上课时间', value: timeModeText.value },
  { label: '上课老师', value: selectedTeacherText.value },
  { label: '上课助教', value: selectedAssistantText.value },
  { label: '上课教室', value: scheduledClassroomText.value },
])

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
      allowClassroomConflict: false,
      tone,
    })),
  )
  const cap = scheduleTargetCount.value
  if (cap <= 0)
    return []
  return unclipped.slice(0, cap)
})

watch(
  () => [
    String(selectedOneToOne.value?.id || ''),
    selectedTeacherIdNormalized.value,
    currentGroup.value,
    plannedDates.value.map(item => item.format('YYYY-MM-DD')).join(','),
    schoolTimeSlotOptions.value.map(item => `${item.start}-${item.end}`).join(','),
  ].join('|'),
  () => {
    scheduleTeacherSlotAvailabilityCheck()
  },
  { immediate: true },
)

watch(
  selectedTeacher,
  (value) => {
    if (!isValidStaffId(value))
      return
    const currentAssistantValues = selectedAssistantValues.value
    const next = currentAssistantValues.filter(id => !sameStaffId(id, value))
    if (next.length !== currentAssistantValues.length) {
      selectedAssistant.value = next.length ? next : undefined
      handleAssistantChange(next)
    }
  },
)

watch(
  () => [currentGroup.value, assistantSelectStaffs.value.map(item => String(item.id)).join(',')].join('|'),
  () => {
    const currentAssistantValues = selectedAssistantValues.value
    const next = currentAssistantValues.filter(id => isAssistantAllowedInCurrentGroup(String(id)))
    if (next.length !== currentAssistantValues.length) {
      selectedAssistant.value = next.length ? next : undefined
      handleAssistantChange(next)
    }
  },
  { immediate: true },
)

watch(
  () => [
    String(selectedOneToOne.value?.id || ''),
    plannedDates.value.map(item => item.format('YYYY-MM-DD')).join(','),
    activeTimeBlocks.value.map(item => `${item.startTime}-${item.endTime}`).join(','),
    assistantSelectStaffs.value.map(item => String(item.id)).join(','),
  ].join('|'),
  () => {
    scheduleAssistantAvailabilityCheck()
  },
  { immediate: true },
)

const estimatedCount = computed(() => previewPlans.value.length)

type ScheduleConflictRow = NonNullable<
  TeachingScheduleValidationResult['currentSchedules']
>[number]

function normalizeHm(t: string) {
  const m = String(t || '').trim().match(/^(\d{1,2}):(\d{2})/)
  if (!m)
    return String(t || '').trim().slice(0, 5)
  const h = Number.parseInt(m[1], 10)
  return `${String(h).padStart(2, '0')}:${m[2]}`
}

function parseScheduleTimeText(timeText: string) {
  const m = String(timeText).match(/^(\d{1,2}:\d{2})[~～](\d{1,2}:\d{2})$/)
  if (!m)
    return null
  return { start: normalizeHm(m[1]), end: normalizeHm(m[2]) }
}

function previewPlanMatchesSchedule(plan: PreviewItem, schedule: ScheduleConflictRow) {
  if (plan.date !== schedule.date)
    return false
  const parsed = parseScheduleTimeText(schedule.timeText)
  if (!parsed)
    return false
  return (
    normalizeHm(plan.startTime) === parsed.start
    && normalizeHm(plan.endTime) === parsed.end
  )
}

/** 预检结果与清单行对齐，用于标红老师/教室列、状态标签 */
const previewPlanConflictMap = computed(() => {
  const val = previewValidationResult.value
  const plans = previewPlans.value
  const map = new Map<string, string[]>()
  if (!val?.currentSchedules?.length || !plans.length)
    return map
  for (const plan of plans) {
    const found = val.currentSchedules.find(s => previewPlanMatchesSchedule(plan, s))
    if (found?.conflictTypes?.length)
      map.set(`${plan.date}|${plan.startTime}|${plan.endTime}`, [...found.conflictTypes])
  }
  return map
})

function previewRowConflictTypes(plan: PreviewItem) {
  return previewPlanConflictMap.value.get(
    `${plan.date}|${plan.startTime}|${plan.endTime}`,
  ) || []
}

function openConflictDetailModal() {
  if (previewValidationResult.value)
    conflictModalOpen.value = true
}

const previewHelperText = computed(() => {
  if (blockedReason.value)
    return blockedReason.value
  if (previewValidating.value)
    return '正在校验老师、教室与时间冲突，请稍候。'
  if (previewHasConflict.value)
    return previewValidationMessage.value || '当前排课方案存在冲突，请返回修改后再尝试创建。'
  if (!estimatedCount.value && excludedHolidayCount.value > 0)
    return '当前日期都命中节假日且已被过滤，请调整日期或关闭节假日过滤。'
  if (!estimatedCount.value)
    return '请先选择有效的排课日期。'
  if (excludedHolidayCount.value > 0)
    return `已根据节假日规则过滤 ${excludedHolidayCount.value} 节，剩余 ${estimatedCount.value} 节待创建。`
  return '已完成预检，可确认创建。正式创建时服务端仍会再校验一次。'
})

const footerTipText = computed(() => {
  if (selectedOneToOne.value?.remark)
    return `档案备注：${selectedOneToOne.value.remark}`
  return '创建后仍可在日程列表中继续调整老师、教室和具体时间。'
})

const actionButtonText = computed(() => {
  if (schedulingMode.value === 'free')
    return estimatedCount.value > 0 ? '创建单次日程' : '创建日程'
  return estimatedCount.value > 0 ? `批量创建 ${estimatedCount.value} 节` : '批量创建日程'
})

/** 多选时段超出可视区域时展示 +N（配合 maxTagCount=responsive） */
function schoolSlotMaxTagPlaceholder(omittedValues: { label?: unknown, value?: unknown }[]) {
  const n = omittedValues?.length ?? 0
  return n > 0 ? `+${n}` : ''
}

async function fetchOneToOneRecords() {
  if (oneToOneLoading.value)
    return

  oneToOneLoading.value = true
  try {
    const pageSize = 200
    let pageIndex = 1
    let total = 0
    const allRecords: OneToOneItem[] = []

    do {
      const res = await getOneToOneListApi({
        pageRequestModel: {
          needTotal: true,
          pageSize,
          pageIndex,
          skipCount: (pageIndex - 1) * pageSize,
        },
        queryModel: {},
      })

      if (res.code !== 200)
        throw new Error(res.message || '获取1对1列表失败')

      const pageList = Array.isArray(res.result?.list) ? res.result.list : []
      total = Number(res.result?.total || 0)
      allRecords.push(...pageList)

      if (!pageList.length)
        break
      pageIndex += 1
    } while (allRecords.length < total)

    oneToOneRecords.value = allRecords

    if (!allRecords.length) {
      selectedOneToOneId.value = undefined
      selectedStudentLessonPath.value = undefined
      plannedClassCount.value = 1
      return
    }

    const preserved = allRecords.find(item => item.id === selectedOneToOneId.value)
    selectedOneToOneId.value = preserved?.id
  }
  catch (error: any) {
    console.error('fetch one to one records failed', error)
    oneToOneRecords.value = []
    selectedOneToOneId.value = undefined
    selectedStudentLessonPath.value = undefined
    messageService.error(error?.message || '获取1对1列表失败')
  }
  finally {
    oneToOneLoading.value = false
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
    console.error('fetch classroom list failed', error)
    messageService.error(error?.message || '获取教室列表失败')
  }
}

async function fetchWorkbenchTeacherList() {
  try {
    const res = await getUserListApi({
      pageRequestModel: {
        needTotal: false,
        pageSize: 500,
        pageIndex: 1,
        skipCount: 0,
      },
      queryModel: {
        status: 0,
      },
    })
    if (res.code !== 200) {
      messageService.error(res.message || '获取老师列表失败')
      return
    }
    const rows = Array.isArray(res.result) ? res.result : []
    workbenchTeacherList.value = rows.map((item: any) => {
      const label = String(item.nickName || item.name || item.id || '').trim()
      return {
        id: String(item.id ?? '').trim(),
        name: label,
        nickName: label,
        mobile: String(item.mobile ?? '').trim(),
      }
    }).filter(item => item.id)
  }
  catch (error: any) {
    console.error('fetch workbench teacher list failed', error)
    messageService.error(error?.message || '获取老师列表失败')
  }
}

function closeModal() {
  previewModalOpen.value = false
  modalOpen.value = false
}

function buildScheduleCreatePayload(options: {
  allowStudentConflict?: boolean
  allowClassroomConflict?: boolean
  assistantIds?: string[]
  plans?: PreviewItem[]
} = {}) {
  const plans = Array.isArray(options.plans) && options.plans.length
    ? options.plans
    : previewPlans.value
  return {
    oneToOneId: String(selectedOneToOne.value?.id || ''),
    teacherId: String(selectedTeacher.value || ''),
    assistantIds: Array.isArray(options.assistantIds)
      ? options.assistantIds.map(id => String(id))
      : selectedAssistantValues.value.map(id => String(id)),
    classroomId: normalizedSelectedClassroomId.value || '',
    allowStudentConflict: options.allowStudentConflict === true,
    allowClassroomConflict: options.allowClassroomConflict === true,
    schedules: plans.map(item => ({
      lessonDate: item.date,
      startTime: item.startTime,
      endTime: item.endTime,
      teacherId: item.teacherId ? String(item.teacherId) : undefined,
      assistantIds: Array.isArray(item.assistantIds) ? item.assistantIds.map(id => String(id)) : undefined,
      classroomId: item.classroomId ? String(item.classroomId) : undefined,
      allowStudentConflict: item.allowStudentConflict === true || options.allowStudentConflict === true,
      allowClassroomConflict: item.allowClassroomConflict === true || options.allowClassroomConflict === true,
    })),
  }
}

async function validatePreviewSchedules() {
  if (!selectedOneToOne.value?.id || previewPlans.value.length === 0)
    return
  previewValidating.value = true
  previewHasConflict.value = false
  previewValidationMessage.value = ''
  previewValidationResult.value = null
  try {
    const res = await validateOneToOneSchedulesApi(buildScheduleCreatePayload())
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
    console.error('validate preview schedules failed', error)
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
  allowStudentConflict?: boolean
  allowClassroomConflict?: boolean
  assistantIds?: string[]
  plans?: PreviewItem[]
} = {}) {
  if (!selectedOneToOne.value?.id)
    return
  const isSoftConflictCreate = options.allowStudentConflict === true || options.allowClassroomConflict === true
  if (isSoftConflictCreate)
    creatingWithSoftConflict.value = true
  else
    creatingSchedules.value = true
  try {
    const res = await createOneToOneSchedulesApi(buildScheduleCreatePayload(options))
    if (res.code !== 200)
      throw new Error(res.message || '创建1对1日程失败')
    const count = res.result?.count || (options.plans?.length || previewPlans.value.length)
    if (isSoftConflictCreate) {
      messageService.success(`已创建 ${count} 节1对1日程，并标记冲突`)
    }
    else {
      messageService.success(`已创建 ${count} 节1对1日程`)
    }
    emitter.emit(EVENTS.REFRESH_DATA)
    conflictModalOpen.value = false
    previewModalOpen.value = false
    modalOpen.value = false
  }
  catch (error: any) {
    console.error('create one-to-one schedules failed', error)
    previewHasConflict.value = true
    previewValidationMessage.value = error?.response?.data?.message || error?.message || '创建1对1日程失败'
    if (previewValidationResult.value)
      conflictModalOpen.value = true
    messageService.error(previewValidationMessage.value)
  }
  finally {
    creatingSchedules.value = false
    creatingWithSoftConflict.value = false
  }
}

function handleConflictWorkbenchSubmit(payload: {
  plans: PreviewItem[]
  assistantIds?: string[]
}) {
  void confirmBatchCreate({
    assistantIds: payload.assistantIds,
    plans: payload.plans,
  })
}

function handleTeacherChange(_value: string | number | undefined, staff?: StaffOptionItem | null) {
  selectedTeacherDisplay.value = staff || null
}

function handleAssistantChange(values?: Array<string | number>) {
  const nextValues = Array.isArray(values)
    ? values
    : selectedAssistantValues.value
  selectedAssistant.value = nextValues.length ? nextValues : undefined
  selectedAssistantDisplays.value = nextValues.map((id) => {
    return assistantSelectStaffs.value.find(item => sameStaffId(item.id, id))
  }).filter(Boolean) as StaffOptionItem[]
}

function handleStudentLessonChange(value?: string[]) {
  selectedStudentLessonPath.value = value?.length ? value : undefined
  if (value?.length >= 2)
    selectedOneToOneId.value = value[1]
  else
    selectedOneToOneId.value = undefined
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
  const inverted = weekDayOptions.filter(item => !selectedWeekdays.value.includes(item))
  selectedWeekdays.value = inverted
}
</script>

<template>
  <div class="one-to-one-schedule-modal-root">
    <a-modal
      v-model:open="modalOpen"
      centered
      class="one-to-one-schedule-modal"
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
              创建1对1日程
            </div>
            <div class="planner-head__subtitle">
              参考当前1对1档案快速生成日程，先把规则配置清楚，再确认创建。
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
            <div class="planner-inline__group planner-inline__group--type">
              <span class="planner-label planner-label--inline">排课类型</span>
              <div class="planner-choice-row planner-choice-row--type-inline">
                <button
                  v-for="item in scheduleTypeOptions"
                  :key="item.value"
                  type="button"
                  class="planner-choice planner-choice--type"
                  :class="{ 'planner-choice--active': scheduleType === item.value }"
                  @click="scheduleType = item.value"
                >
                  <span class="planner-choice__title">{{ item.label }}</span>
                </button>
              </div>
            </div>

            <div class="planner-inline__group planner-inline__group--record">
              <span class="planner-label planner-label--inline">
                <BookOutlined />
                {{ scheduleType === 'oneToOne' ? '选择1对1' : '选择学员和课程' }}
              </span>
              <a-select
                v-if="scheduleType === 'oneToOne'"
                v-model:value="selectedOneToOneId"
                size="large"
                show-search
                option-filter-prop="label"
                option-label-prop="label"
                popup-class-name="planner-record-select-dropdown"
                allow-clear
                :loading="oneToOneLoading"
                :disabled="oneToOneLoading || !oneToOneSelectOptions.length"
                :not-found-content="oneToOneLoading ? '正在加载1对1数据...' : '暂无1对1数据'"
                :placeholder="oneToOneLoading ? '正在加载1对1数据...' : '请选择1对1'"
                class="planner-control planner-control--record"
              >
                <a-select-option
                  v-for="item in oneToOneSelectOptions"
                  :key="item.value"
                  :value="item.value"
                  :label="item.label"
                >
                  <div class="planner-option">
                    <div class="planner-option__title">
                      {{ item.label }}
                    </div>
                  </div>
                </a-select-option>
              </a-select>
              <a-cascader
                v-else
                v-model:value="selectedStudentLessonPath"
                :options="studentLessonCascaderOptions"
                allow-clear
                :disabled="oneToOneLoading || !studentLessonCascaderOptions.length"
                :placeholder="oneToOneLoading ? '正在加载1对1数据...' : '请选择学员和课程'"
                class="planner-control planner-control--record"
                @change="handleStudentLessonChange"
              />
            </div>

            <div class="planner-inline__group planner-inline__group--status">
              <span class="planner-label planner-label--inline">账户余额</span>
              <div class="planner-balance">
                <div class="planner-balance__item">
                  <span>已用数量</span>
                  <strong>{{ formatBalanceValue(usedQuantityValue) }}</strong>
                </div>
                <div class="planner-balance__item">
                  <span>剩余数量</span>
                  <strong>{{ formatBalanceValue(remainQuantityValue) }}</strong>
                </div>
              </div>
            </div>
          </div>
        </section>

        <div v-if="selectedOneToOne?.id" class="planner-layout">
          <aside class="planner-aside">
            <section class="planner-card planner-card--summary">
              <div class="planner-card__head">
                <div class="planner-card__title">
                  当前档案
                </div>
                <div class="planner-card__desc">
                  来自当前1对1记录的基础信息与创建摘要。
                </div>
              </div>

              <div class="planner-profile">
                <div class="planner-profile__name">
                  {{ selectedOneToOne?.name || '-' }}
                </div>
                <div class="planner-profile__meta">
                  {{ selectedOneToOne?.studentName || '-' }} · {{ selectedOneToOne?.lessonName || '-' }}
                </div>
              </div>

              <div class="planner-summary-list">
                <div
                  v-for="item in selectedOneToOneSummary"
                  :key="item.label"
                  class="planner-summary-list__row"
                >
                  <span>{{ item.label }}</span>
                  <strong class="planner-summary-list__value">{{ item.value }}</strong>
                </div>
              </div>

              <div class="planner-card__subhead">
                创建摘要
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

              <div v-if="selectedOneToOne?.remark" class="planner-note">
                {{ selectedOneToOne.remark }}
              </div>

              <div v-if="planExceedsRemainHint" class="planner-alert planner-alert--soft">
                {{ planExceedsRemainHint }}
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
                  按顺序完成排课方式、日期规则和时间资源。
                </div>
              </div>

              <div class="planner-section">
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
                  <div
                    v-if="schedulingMode === 'repeat'"
                    class="planner-date-plan-row"
                  >
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
                        <span>根据计划上课次数与重复规则自动推算</span>
                      </div>
                    </div>

                    <span class="planner-field__hint planner-date-plan-row__hint">可自由填写节数；结束日期由开始日期、重复规则与本次数推算。参考剩余可排 {{ remainSessionCap }} 节。</span>
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

                  <div v-if="schedulingMode === 'repeat'" class="planner-field planner-field--full">
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

                  <div class="planner-field planner-field--full">
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
                      placeholder="可不选"
                      :options="classroomOptions"
                      class="planner-control"
                    />
                  </label>

                  <div class="planner-field planner-field--major planner-field--full">
                    <div class="planner-field__label-row planner-field__label-row--with-period-group">
                      <span class="planner-label planner-label--required planner-label--with-inline-tip">
                        <ClockCircleOutlined />
                        课表节次（可多选）
                        <a-tooltip title="同一上课日内按所选「第几节课」各生成一节日程；重复排课时每个上课日都会生成这些节，总节数仍受剩余课时/额度约束。">
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
                        v-model:value="selectedSchoolTimeSlots"
                        mode="multiple"
                        size="large"
                        option-label-prop="label"
                        allow-clear
                        placeholder="同一天内可勾选多节，例如上午 + 下午"
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
                      :placeholder="`可不选，仅${groupOptions.find(o => o.key === currentGroup)?.label || '当前组'}可选`"
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
                            <span class="planner-staff-option__label">{{ item.label }}</span>
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
            :description="oneToOneLoading ? '正在加载1对1数据...' : '请选择1对1后查看档案并继续排课'"
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
              预计排课清单
            </div>
            <div class="planner-review__subtitle">
              先确认本次将创建的日程，再执行批量创建。
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
            @click="openConflictDetailModal"
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
                  <td
                    :class="{
                      'planner-table__cell--danger': previewRowConflictTypes(item).includes('老师'),
                    }"
                  >
                    {{ item.teacher }}
                  </td>
                  <td
                    :class="{
                      'planner-table__cell--danger': previewRowConflictTypes(item).includes('助教'),
                    }"
                  >
                    {{ item.assistant }}
                  </td>
                  <td
                    :class="{
                      'planner-table__cell--danger': previewRowConflictTypes(item).includes('教室'),
                    }"
                  >
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
                    <span
                      v-else
                      class="planner-tag planner-tag--table"
                    >
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
            <a-button type="primary" :disabled="!isSchedulable || estimatedCount === 0 || previewValidating || previewHasConflict" :loading="creatingSchedules" @click="confirmBatchCreate">
              {{ actionButtonText }}
            </a-button>
          </div>
        </div>
      </div>
    </a-modal>
  </div>

  <ScheduleConflictWorkbenchModal
    v-model:open="conflictModalOpen"
    :one-to-one-id="String(selectedOneToOne?.id || '')"
    :assistant-ids="selectedAssistantValues.map(id => String(id))"
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
.one-to-one-schedule-modal-root {
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

.planner-head__close {
  color: #94a3b8;
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

.planner-shell {
  --planner-major-height: 42px;
  display: flex;
  max-height: calc(100vh - 170px);
  flex-direction: column;
  gap: 10px;
  overflow: auto;
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
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.planner-inline--top {
  align-items: center;
  gap: 18px;
}

.planner-inline__group {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 12px;
}

.planner-inline__group--record {
  flex: 1;
}

.planner-inline__group--status {
  flex-shrink: 0;
  justify-content: flex-end;
}

.planner-inline__group--type {
  flex: 0 0 auto;
}

.planner-control--record {
  flex: 1;
  min-width: 0;
}

.planner-label {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #4e5969;
  font-size: 13px;
  font-weight: 600;
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

.planner-label__tip {
  color: #9aa4b2;
  font-size: 14px;
}

.planner-tag-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: flex-start;
}

.planner-balance {
  display: flex;
  gap: 6px;
}

.planner-balance__item {
  min-width: 100px;
  padding: 6px 10px;
  border: 1px solid #f90;
  border-radius: 10px;
  background: #fff5e6;
}

.planner-balance__item span {
  display: block;
  color: #86909c;
  font-size: 11px;
  line-height: 1.4;
}

.planner-balance__item strong {
  display: block;
  margin-top: 2px;
  color: #1f2329;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.45;
}

.planner-tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 30px;
  padding: 0 12px;
  border-radius: 999px;
  background: #f2f4f7;
  color: #4e5969;
  font-size: 12px;
  font-weight: 500;
}

.planner-tag--primary {
  background: #e6f4ff;
  color: #1677ff;
}

.planner-tag--warning {
  background: #fff7e6;
  color: #d46b08;
}

.planner-tag--table {
  min-height: 26px;
  padding: 0 10px;
}

.planner-choice-row {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.planner-choice-row--type-inline {
  display: flex;
  flex-wrap: nowrap;
  gap: 8px;
}

button.planner-choice {
  display: flex;
  width: 100%;
  min-height: 58px;
  flex-direction: column;
  justify-content: center;
  gap: 4px;
  padding: 10px 12px;
  border: 1px solid #d8e1eb;
  border-radius: 14px;
  background: #fff;
  cursor: pointer;
  appearance: none;
  font: inherit;
  outline: none;
  text-align: left;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    box-shadow 0.2s ease;
}

button.planner-choice:hover {
  border-color: #91caff;
}

button.planner-choice.planner-choice--active {
  border-color: #1677ff;
  background: #f7fbff;
  box-shadow: 0 0 0 3px rgb(22 119 255 / 8%);
}

.planner-choice__title {
  color: #1f2329;
  font-size: 14px;
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

button.planner-choice.planner-choice--type {
  width: auto;
  min-height: 40px;
  flex-direction: row;
  align-items: center;
  gap: 0;
  padding: 0 14px;
  border-radius: 999px;
}

button.planner-choice.planner-choice--type .planner-choice__title {
  font-size: 14px;
}

.planner-option__title {
  color: #1f2329;
  font-size: 13px;
  line-height: 1.5;
}

.planner-option__desc {
  color: #86909c;
  font-size: 12px;
  line-height: 1.5;
}

.planner-option--inline {
  display: flex;
  align-items: center;
}

.planner-layout {
  display: grid;
  grid-template-columns: 300px minmax(0, 1fr);
  gap: 14px;
  align-items: stretch;
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
}

.planner-card--empty-state {
  display: flex;
  min-height: 460px;
  align-items: center;
  justify-content: center;
  padding: 24px;
}

.planner-card__head {
  padding: 16px 18px 0;
}

.planner-card__title {
  color: #1f2329;
  font-size: 15px;
  font-weight: 600;
  line-height: 1.5;
}

.planner-card__desc {
  margin-top: 4px;
  color: #86909c;
  font-size: 12px;
  line-height: 1.6;
}

.planner-card__subhead {
  margin: 0 18px 10px;
  padding-top: 14px;
  border-top: 1px solid #f1f4f8;
  color: #1f2329;
  font-size: 13px;
  font-weight: 600;
  line-height: 1.5;
}

.planner-profile {
  padding: 14px 18px 0;
}

.planner-profile__name {
  color: #1f2329;
  font-size: 18px;
  font-weight: 600;
  line-height: 1.4;
}

.planner-profile__meta {
  margin-top: 4px;
  color: #86909c;
  font-size: 13px;
  line-height: 1.6;
}

.planner-summary-list {
  display: flex;
  flex-direction: column;
  gap: 0;
  padding: 14px 18px 16px;
}

.planner-summary-list--secondary {
  padding-top: 0;
}

.planner-summary-list__row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 0;
  border-bottom: 1px dashed #edf0f3;
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
  min-width: 0;
  flex: 1;
  text-align: right;
  cursor: default;
}

.planner-summary-list__value {
  display: block;
  min-width: 0;
  flex: 1;
  overflow: hidden;
  color: #1f2329;
  font-size: 12px;
  font-weight: 600;
  line-height: 1.6;
  text-align: right;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.planner-summary-list__value-wrap .planner-summary-list__value {
  flex: none;
  width: 100%;
}

.planner-note,
.planner-alert {
  margin: 0 18px 18px;
  padding: 12px 14px;
  border-radius: 12px;
  font-size: 13px;
  line-height: 1.7;
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

.planner-field__hint {
  display: block;
  margin-top: 6px;
  color: #86909c;
  font-size: 12px;
  line-height: 1.55;
}

.planner-section {
  padding: 16px 18px 18px;
}

.planner-section + .planner-section {
  border-top: 1px solid #f1f4f8;
}

.planner-section__title {
  margin-bottom: 10px;
  color: #1f2329;
  font-size: 15px;
  font-weight: 600;
  line-height: 1.5;
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
  display: grid;
  /* 开始日期固定舒适宽度，结束日期列仍用 1fr 吃剩余空间 */
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
}

:deep(.planner-date-plan-row__cell--start .ant-picker) {
  width: 100% !important;
  max-width: 200px !important;
}

/* 计划次数输入撑满中间列，避免单元格内右侧留白导致与结束日期间距「看起来」不对称 */
.planner-date-plan-row__cell--count {
  width: 100%;
  min-width: 0;
}

.planner-date-plan-row__cell--count .planner-control {
  width: 100%;
}

:deep(.planner-date-plan-row__cell--count .ant-input-number) {
  width: 100% !important;
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

.planner-static-field--free-trigger {
  cursor: pointer;
}

.planner-static-field--free-trigger strong {
  display: block;
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.planner-static-field--free-trigger span {
  font-size: 11px;
}

.planner-free-calendar {
  width: 300px;
}

.planner-free-calendar__footer {
  display: flex;
  justify-content: center;
  padding-top: 8px;
  border-top: 1px solid #f0f0f0;
}

.planner-free-calendar__cell {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
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

/* 与左侧列「图标 + gap」同宽，使「各时段课长」与「上课助教」文案纵列对齐 */
.planner-label__lead-spacer {
  display: inline-block;
  width: calc(14px + 6px);
  flex-shrink: 0;
}

.planner-date-plan-row__cell .planner-static-field--inline-hint span {
  white-space: normal;
}

.planner-date-plan-row__cell .planner-static-field--inline-hint strong {
  white-space: nowrap;
  flex-shrink: 0;
}

.planner-field {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 8px;
}

.planner-field--full {
  grid-column: 1 / -1;
}

.planner-field--major {
  align-self: start;
}

.planner-field__label-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
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
  display: inline-flex;
  align-items: center;
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
  /* 避免 grid/flex 子项 min-width:0 把多选压成一条竖线 */
  min-width: 320px;
}

.planner-chip-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}

.planner-chip-row--weekday {
  align-items: center;
  flex-wrap: nowrap;
  overflow-x: auto;
}

.planner-chip-actions {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  margin-left: 6px;
  flex-shrink: 0;
  white-space: nowrap;
}

.planner-holiday-inline {
  display: inline-flex;
  align-items: center;
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
  font-size: 12px;
  font-weight: 600;
  line-height: 1.5;
}

.planner-holiday-inline__list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.planner-holiday-inline__item {
  display: inline-flex;
  align-items: center;
  height: 26px;
  padding: 0 10px;
  border-radius: 999px;
  background: #fff5e6;
  color: #d46b08;
  font-size: 12px;
  line-height: 1;
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
  cursor: pointer;
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

.planner-static-field {
  display: flex;
  min-height: 42px;
  flex-direction: column;
  justify-content: center;
  gap: 4px;
  padding: 10px 12px;
  border: 1px solid #e5ebf2;
  border-radius: 12px;
  background: #fafbfd;
}

.planner-static-field strong {
  color: #1f2329;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.5;
}

.planner-static-field span {
  color: #86909c;
  font-size: 12px;
  line-height: 1.5;
}

.planner-static-field--compact {
  min-height: 42px;
  justify-content: center;
  padding: 9px 12px;
}

.planner-static-field--compact strong {
  font-size: 14px;
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
  margin-top: 0;
  text-align: right;
  white-space: nowrap;
}

.planner-static-field--major {
  height: var(--planner-major-height);
  padding: 10px 12px;
  border-radius: 12px;
}

.planner-static-field--inline {
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.planner-static-field--inline strong {
  font-size: 14px;
  flex-shrink: 0;
}

.planner-static-field--inline span {
  font-size: 12px;
  text-align: right;
}

.planner-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 14px 6px 0;
}

.planner-footer__tip {
  flex: 1;
  min-width: 0;
  color: #86909c;
  font-size: 12px;
  line-height: 1.6;
}

.planner-footer__actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.planner-review {
  display: flex;
  max-height: calc(100vh - 220px);
  flex-direction: column;
}

.planner-review__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.planner-review__title {
  color: #1f2329;
  font-size: 20px;
  font-weight: 600;
  line-height: 1.4;
}

.planner-review__subtitle {
  margin-top: 4px;
  color: #86909c;
  font-size: 13px;
  line-height: 1.6;
}

.planner-review__count {
  flex-shrink: 0;
  padding-right: 36px;
  color: #1677ff;
  font-size: 16px;
  font-weight: 600;
  line-height: 1.6;
}

.planner-review__tip {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px 12px;
  padding: 12px 14px;
  border-radius: 12px;
  background: #f6f8fa;
  color: #4e5969;
  font-size: 13px;
  line-height: 1.6;
}

.planner-review__tip-text {
  flex: 1;
  min-width: 200px;
}

.planner-review__conflict-link {
  flex-shrink: 0;
  padding: 0;
  height: auto;
  font-weight: 600;
}

.planner-review__tip--warning {
  background: #fff7e6;
  color: #d46b08;
}

.planner-review__conflict-tag {
  margin-inline-end: 0;
}

.planner-table__cell--danger {
  color: #ff4d4f;
  font-weight: 700;
}

.planner-review__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding-top: 16px;
  border-top: 1px solid #f0f0f0;
}

.planner-table-wrap {
  overflow: auto;
}

.planner-table-wrap--modal {
  max-height: min(56vh, 540px);
  overflow-x: auto;
  overflow-y: auto;
}

.planner-table {
  width: max-content;
  min-width: 100%;
  border-collapse: separate;
  border-spacing: 0;
}

.planner-table thead th {
  position: sticky;
  top: 0;
  z-index: 2;
  padding: 12px 14px;
  border-bottom: 1px solid #f0f0f0;
  background: #fafafa;
  color: #4e5969;
  font-size: 13px;
  font-weight: 600;
  text-align: left;
  white-space: nowrap;
}

.planner-table tbody td {
  padding: 14px;
  border-bottom: 1px solid #f5f5f5;
  color: #1f2329;
  font-size: 13px;
  line-height: 1.6;
  vertical-align: middle;
  white-space: nowrap;
}

.planner-table tbody tr:last-child td {
  border-bottom: none;
}

.planner-table__row--blocked td {
  background: #fffaf0;
}

.planner-table__empty {
  padding: 36px 14px !important;
  color: #86909c !important;
  text-align: center;
}

.planner-table th:nth-child(1),
.planner-table td:nth-child(1) {
  min-width: 120px;
}

.planner-table th:nth-child(2),
.planner-table td:nth-child(2) {
  min-width: 92px;
}

.planner-table th:nth-child(3),
.planner-table td:nth-child(3) {
  min-width: 420px;
}

.planner-table th:nth-child(4),
.planner-table td:nth-child(4) {
  min-width: 220px;
}

.planner-table th:nth-child(5),
.planner-table td:nth-child(5) {
  min-width: 110px;
}

.planner-table th:nth-child(6),
.planner-table td:nth-child(6) {
  min-width: 110px;
}

.planner-table th:nth-child(7),
.planner-table td:nth-child(7) {
  min-width: 110px;
}

.planner-table__status-cell {
  position: sticky;
  right: 0;
  z-index: 3;
  min-width: 160px;
  background: #fff;
  box-shadow: -8px 0 12px -10px rgb(15 23 42 / 18%);
}

.planner-table thead .planner-table__status-cell {
  z-index: 4;
  background: #fafafa;
}

.planner-table__row--blocked .planner-table__status-cell {
  background: #fffaf0;
}

:deep(.one-to-one-schedule-modal .ant-modal-content),
:deep(.planner-review-modal .ant-modal-content) {
  overflow: hidden;
  padding: 0;
  border-radius: 20px;
}

:deep(.one-to-one-schedule-modal .ant-modal-header) {
  margin-bottom: 0;
  padding: 14px 20px 10px;
  border-bottom: none;
}

:deep(.one-to-one-schedule-modal .ant-modal-body) {
  padding: 0;
}

:deep(.one-to-one-schedule-modal .ant-modal-footer) {
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

:deep(.planner-control.ant-picker),
:deep(.planner-control .ant-select-selector) {
  border-color: #d9e1ea !important;
  border-radius: 12px !important;
  box-shadow: none !important;
  min-height: 42px;
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
}

:deep(.planner-control.ant-select-single .ant-select-selection-search-input) {
  height: 40px !important;
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

/* 与胶囊标签的 margin + min-height 对齐，避免 caret 相对标签行偏上 */
:deep(.planner-multi-slot-select.ant-select-multiple .ant-select-selection-search) {
  display: inline-flex !important;
  align-items: center !important;
  margin-top: 8px;
  margin-bottom: 6px;
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
  flex-shrink: 0;
  white-space: nowrap !important;
  display: inline-flex;
  align-items: center;
  min-height: 28px;
  height: auto;
  margin-top: 6px;
  margin-bottom: 6px;
  padding: 0 10px 0 12px;
  border: 1px solid #bfd8ff;
  border-radius: 8px;
  background: #eef6ff;
  color: #1677ff;
  font-size: 13px;
  font-weight: 400;
  max-width: none !important;
}

:deep(.planner-multi-slot-select.ant-select-multiple .ant-select-selection-item-content) {
  color: inherit;
  max-width: none !important;
  overflow: visible !important;
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

:deep(.planner-control--record .ant-select-selector),
:deep(.planner-control--record .ant-select-selection-item),
:deep(.planner-control--record .ant-select-selection-search),
:deep(.planner-control--record .ant-select-selection-search-input),
:deep(.planner-control--record .ant-select-selection-placeholder) {
  font-size: 14px !important;
  font-weight: 400 !important;
}

:deep(.planner-control--record .ant-select-selector),
:deep(.planner-control--record .ant-cascader-picker),
:deep(.planner-control--record .ant-cascader-picker-label),
:deep(.planner-control--record .ant-cascader-input) {
  display: flex;
  align-items: center;
  min-height: 42px;
}

:deep(.planner-control--record .ant-cascader-picker-label) {
  inset: 0 40px 0 12px;
  line-height: 42px;
}

:deep(.ant-select:not(.planner-multi-slot-select) .ant-select-selection-item) {
  font-size: 14px !important;
}

:deep(.ant-select-selection-search-input),
:deep(.ant-select-selection-placeholder) {
  font-size: 14px !important;
}

:deep(.planner-control:not(.planner-multi-slot-select) .ant-select-selection-item),
:deep(.planner-control.ant-picker .ant-picker-input > input) {
  font-size: 14px !important;
}

:deep(.planner-record-select-dropdown) {
  padding: 8px;
}

:deep(.planner-record-select-dropdown .ant-select-item) {
  padding: 0;
}

:deep(.planner-record-select-dropdown .ant-select-item-option) {
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

:deep(.planner-record-select-dropdown .ant-select-item-option:last-child) {
  margin-bottom: 0;
}

:deep(.planner-record-select-dropdown .ant-select-item-option-content) {
  padding: 10px 12px;
}

:deep(.planner-record-select-dropdown .ant-select-item-option-active:not(.ant-select-item-option-disabled)) {
  border-color: #dbeafe;
  background: #f8fbff;
}

:deep(.planner-record-select-dropdown .ant-select-item-option-selected:not(.ant-select-item-option-disabled)) {
  border-color: #bfd8ff;
  background: linear-gradient(180deg, #f7fbff 0%, #eef6ff 100%);
  box-shadow: inset 0 0 0 1px rgb(22 119 255 / 10%);
}

:deep(.planner-record-select-dropdown .planner-option) {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

:deep(.planner-record-select-dropdown .planner-option__title) {
  overflow: hidden;
  color: #1f2329;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.45;
  text-overflow: ellipsis;
  white-space: nowrap;
}

:deep(.planner-record-select-dropdown .planner-option__desc) {
  overflow: hidden;
  color: #7d8898;
  font-size: 12px;
  line-height: 1.5;
  text-overflow: ellipsis;
  white-space: nowrap;
}

:deep(.one-to-one-schedule-modal .ant-select-focused .ant-select-selector),
:deep(.one-to-one-schedule-modal .ant-picker-focused) {
  border-color: #1677ff !important;
}

@media (max-width: 1200px) {
  .planner-layout {
    grid-template-columns: minmax(0, 1fr);
  }

  .planner-inline,
  .planner-inline__group {
    align-items: stretch;
  }

  .planner-inline {
    flex-direction: column;
  }

  .planner-inline__group--record,
  .planner-inline__group--status,
  .planner-inline__group--type {
    width: 100%;
  }

  .planner-inline__group--status {
    justify-content: flex-start;
  }

  .planner-balance {
    flex-wrap: wrap;
  }

  .planner-choice-row--type-inline {
    flex-wrap: wrap;
  }

  .planner-control--record {
    width: 100%;
  }
}

@media (max-width: 768px) {
  .planner-head,
  .planner-footer,
  .planner-review__head,
  .planner-review__footer {
    flex-direction: column;
    align-items: stretch;
  }

  .planner-head__stats {
    justify-content: space-between;
  }

  .planner-choice-row,
  .planner-form-grid {
    grid-template-columns: minmax(0, 1fr);
  }

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
    align-items: flex-start;
  }

  .planner-choice-row--type-inline {
    width: 100%;
    flex-direction: column;
  }
}
</style>

<style lang="less">
/* 渲染在 body 上，需高于 Modal(1000) / Mask，避免被挡或焦点异常 */
.planner-schedule-slot-select-dropdown.ant-select-dropdown {
  z-index: 3000 !important;
}

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

.planner-record-select-dropdown .planner-option__desc {
  overflow: hidden;
  color: #7d8898;
  font-size: 12px;
  line-height: 1.5;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
