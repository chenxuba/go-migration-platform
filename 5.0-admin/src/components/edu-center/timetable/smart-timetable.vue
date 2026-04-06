<script setup>
import { Modal } from 'ant-design-vue'
import dayjs from 'dayjs'
import { h } from 'vue'
import SmartTimetableConflictModal from './smart-timetable-conflict-modal.vue'
import SmartTimetableDragConfirmModal from './smart-timetable-drag-confirm-modal.vue'
import SmartTimetableDragConflictModal from './smart-timetable-drag-conflict-modal.vue'
import SmartTimetableGrid from './smart-timetable-grid.vue'
import SmartTimetableScheduledDetailModal from './smart-timetable-scheduled-detail-modal.vue'
import SmartTimetableToolbar from './smart-timetable-toolbar.vue'
import ScheduleConflictModal from './schedule-conflict-modal.vue'
import { listClassroomsApi } from '@/api/business-settings/classroom'
import { getOneToOneListApi } from '@/api/edu-center/one-to-one'
import { pageGroupClassesApi } from '@/api/edu-center/group-class'
import { getCourseIdAndNameApi } from '@/api/edu-center/registr-renewal'
import { batchUpdateTeachingSchedulesApi, cancelTeachingSchedulesApi, createOneToOneSchedulesApi, getTeachingScheduleConflictDetailApi, listTeachingSchedulesByTeacherMatrixApi, validateOneToOneSchedulesApi } from '@/api/edu-center/teaching-schedule'
import { getUserListApi } from '@/api/internal-manage/staff-manage'
import { useSmartTimetableAvailability } from '@/composables/useSmartTimetableAvailability'
import { useSmartTimetableClassMode } from '@/composables/useSmartTimetableClassMode'
import { useSmartTimetablePicker } from '@/composables/useSmartTimetablePicker'
import { useUserStore } from '@/stores/user'
import messageService from '@/utils/messageService'
import {
  DEFAULT_UNIFIED_TIME_PERIOD_CONFIG,
  buildQuickHourlySlots,
  configGroupsSorted,
  parseUnifiedTimePeriodConfig,
} from '@/utils/unified-time-period'
import emitter, { EVENTS } from '@/utils/eventBus'

const displayArray = ref([
  'scheduleTeacher',
  'scheduleClassroom',
  'scheduleClass',
  'scheduleOneToOne',
  'scheduleCourse',
  'scheduleType',
  'scheduleCallStatus',
])
const SMART_TIMETABLE_VIEW_MODE_KEY = 'smart-timetable-view-mode'
const DRAG_BATCH_VALIDATE_SINGLE_REQUEST_THRESHOLD = 500
const DRAG_BATCH_VALIDATE_CHUNK_SIZE = 300
const DRAG_BATCH_VALIDATE_CONCURRENCY = 2

function getSavedTimeView() {
  if (typeof window === 'undefined')
    return 'week'
  try {
    const saved = window.localStorage.getItem(SMART_TIMETABLE_VIEW_MODE_KEY)
    if (saved === 'day' || saved === 'week' || saved === 'swapWeek')
      return saved
  }
  catch {
  }
  return 'week'
}

// 当前选中的时间维度
const currentTime = ref(getSavedTimeView())
// 当前的日期区间 - 默认设置为本周
const currentWeek = ref(dayjs())
/** 课表时间视图：下拉与日期导航联动 */
const timeViewOptions = [
  { value: 'day', label: '日视图' },
  { value: 'week', label: '日期视图' },
  { value: 'swapWeek', label: '时间视图' },
]
/** 1=1v1，2=班课 */
const currentModel = ref('1')
const currentGroup = ref('A')
/** 与 matrixDays、表头节次列对齐；切换 A/B 时在新数据返回前不改，避免清空矩阵导致整页高度塌缩抖动 */
const displayedGroupKey = ref('A')
const timetableRootRef = ref(null)
const classId = ref(null)
const filterStudentId = ref(undefined)
const filterTeacherId = ref([])
const filterClassroomId = ref([])
const filterClassId = ref(undefined)
const filterOneToOneId = ref(undefined)
const filterCourseId = ref(undefined)
const filterScheduleType = ref([])
const filterCallStatus = ref(undefined)

const scheduleTypeOptions = [
  { id: 'group_class', value: '班级日程' },
  { id: 'one_to_one', value: '1对1日程' },
  { id: 'trial', value: '试听日程' },
]

const scheduleCallStatusOptions = [
  { id: 'unsigned', value: '未点名' },
  { id: 'signed', value: '已点名' },
]

function normalizeScheduleFilterValue(value) {
  if (Array.isArray(value))
    return value.length ? value[0] : undefined
  const text = String(value ?? '').trim()
  return text ? text : undefined
}

function normalizeScheduleFilterValues(value) {
  if (!Array.isArray(value))
    return []
  return value.map(item => String(item ?? '').trim()).filter(Boolean)
}

function handleScheduleTeacherFilter(value) {
  filterTeacherId.value = normalizeScheduleFilterValues(value)
}

function handleScheduleClassroomFilter(value) {
  filterClassroomId.value = normalizeScheduleFilterValues(value)
}

function handleScheduleClassFilter(value) {
  filterClassId.value = normalizeScheduleFilterValue(value)
}

function handleScheduleOneToOneFilter(value) {
  filterOneToOneId.value = normalizeScheduleFilterValue(value)
}

function handleScheduleCourseFilter(value) {
  filterCourseId.value = normalizeScheduleFilterValue(value)
}

function handleScheduleTypeFilter(value) {
  filterScheduleType.value = normalizeScheduleFilterValues(value)
}

function handleScheduleCallStatusFilter(value) {
  filterCallStatus.value = normalizeScheduleFilterValue(value)
}

function handleStuPhoneFilter(value) {
  filterStudentId.value = normalizeScheduleFilterValue(value)
}

function mergeFilterOptions(previous, incoming, selectedValues = []) {
  const selectedSet = new Set((Array.isArray(selectedValues) ? selectedValues : [selectedValues]).map(value => String(value || '')).filter(Boolean))
  const map = new Map()
  ;(Array.isArray(previous) ? previous : []).forEach((item) => {
    const key = String(item?.id || '').trim()
    if (key && selectedSet.has(key))
      map.set(key, item)
  })
  ;(Array.isArray(incoming) ? incoming : []).forEach((item) => {
    const key = String(item?.id || '').trim()
    if (key)
      map.set(key, item)
  })
  return [...map.values()]
}

async function loadScheduleTeacherOptions(searchKey = '', reset = true) {
  if (reset) {
    scheduleTeacherPagination.value.current = 1
    scheduleTeacherFinished.value = false
  }
  scheduleTeacherSearchKey.value = searchKey
  try {
    const res = await getUserListApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: scheduleTeacherPagination.value.pageSize,
        pageIndex: scheduleTeacherPagination.value.current,
        skipCount: 0,
      },
      queryModel: {
        searchKey,
        status: 0,
      },
      sortModel: {},
    })
    if (res.code !== 200)
      return
    const resultData = (Array.isArray(res.result) ? res.result : []).map(item => ({
      id: String(item.id ?? ''),
      value: String(item.nickName || item.name || item.value || item.id || '').trim(),
    })).filter(item => item.id && item.value)
    scheduleTeacherOptions.value = reset
      ? mergeFilterOptions(scheduleTeacherOptions.value, resultData, filterTeacherId.value)
      : mergeFilterOptions(scheduleTeacherOptions.value, [...scheduleTeacherOptions.value, ...resultData], filterTeacherId.value)
    const total = Number(res.total || resultData.length || 0)
    scheduleTeacherPagination.value.total = total
    scheduleTeacherFinished.value = scheduleTeacherOptions.value.length >= total
  }
  catch (error) {
    console.error('load schedule teacher options failed', error)
  }
}

async function loadScheduleClassroomOptions(searchKey = '') {
  scheduleClassroomSearchKey.value = searchKey
  try {
    const res = await listClassroomsApi({
      enabledOnly: true,
      searchKey,
    })
    if (res.code !== 200)
      return
    const resultData = (Array.isArray(res.result) ? res.result : []).map(item => ({
      id: String(item.id ?? ''),
      value: String(item.name || item.id || '').trim(),
    })).filter(item => item.id && item.value)
    scheduleClassroomOptions.value = mergeFilterOptions(scheduleClassroomOptions.value, resultData, filterClassroomId.value)
    scheduleClassroomFinished.value = true
  }
  catch (error) {
    console.error('load schedule classroom options failed', error)
  }
}

async function loadScheduleClassOptions(searchKey = '', reset = true) {
  if (reset) {
    scheduleClassPagination.value.current = 1
    scheduleClassFinished.value = false
  }
  scheduleClassSearchKey.value = searchKey
  try {
    const res = await pageGroupClassesApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: scheduleClassPagination.value.pageSize,
        pageIndex: scheduleClassPagination.value.current,
        skipCount: 0,
      },
      queryModel: {
        className: searchKey || undefined,
      },
    })
    if (res.code !== 200)
      return
    const list = Array.isArray(res.result?.list) ? res.result.list : []
    const resultData = list.map(item => ({
      id: String(item.id ?? ''),
      value: String(item.name || item.id || '').trim(),
    })).filter(item => item.id && item.value)
    scheduleClassOptions.value = reset
      ? mergeFilterOptions(scheduleClassOptions.value, resultData, filterClassId.value)
      : mergeFilterOptions(scheduleClassOptions.value, [...scheduleClassOptions.value, ...resultData], filterClassId.value)
    const total = Number(res.result?.total || resultData.length || 0)
    scheduleClassPagination.value.total = total
    scheduleClassFinished.value = scheduleClassOptions.value.length >= total
  }
  catch (error) {
    console.error('load schedule class options failed', error)
  }
}

async function loadScheduleOneToOneOptions(searchKey = '', reset = true) {
  if (reset) {
    scheduleOneToOnePagination.value.current = 1
    scheduleOneToOneFinished.value = false
  }
  scheduleOneToOneSearchKey.value = searchKey
  try {
    const res = await getOneToOneListApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: scheduleOneToOnePagination.value.pageSize,
        pageIndex: scheduleOneToOnePagination.value.current,
        skipCount: 0,
      },
      queryModel: {
        status: [1],
        searchKey,
      },
    })
    if (res.code !== 200)
      return
    const list = Array.isArray(res.result?.list) ? res.result.list : []
    const resultData = list.map(item => ({
      id: String(item.id ?? ''),
      value: `${String(item.studentName || item.name || item.id || '').trim()}～${String(item.lessonName || '').trim()}`.replace(/～$/, ''),
    })).filter(item => item.id && item.value)
    scheduleOneToOneOptions.value = reset
      ? mergeFilterOptions(scheduleOneToOneOptions.value, resultData, filterOneToOneId.value)
      : mergeFilterOptions(scheduleOneToOneOptions.value, [...scheduleOneToOneOptions.value, ...resultData], filterOneToOneId.value)
    const total = Number(res.result?.total || resultData.length || 0)
    scheduleOneToOnePagination.value.total = total
    scheduleOneToOneFinished.value = scheduleOneToOneOptions.value.length >= total
  }
  catch (error) {
    console.error('load schedule one to one options failed', error)
  }
}

async function loadScheduleCourseOptions(searchKey = '', reset = true) {
  if (reset) {
    scheduleCoursePagination.value.current = 1
    scheduleCourseFinished.value = false
  }
  scheduleCourseSearchKey.value = searchKey
  try {
    const res = await getCourseIdAndNameApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: scheduleCoursePagination.value.pageSize,
        pageIndex: scheduleCoursePagination.value.current,
        skipCount: 0,
      },
      queryModel: {
        searchKey,
      },
      sortModel: {},
    })
    if (res.code !== 200)
      return
    const resultData = (Array.isArray(res.result) ? res.result : []).map(item => ({
      id: String(item.id ?? ''),
      value: String(item.name || item.id || '').trim(),
    })).filter(item => item.id && item.value)
    scheduleCourseOptions.value = reset
      ? mergeFilterOptions(scheduleCourseOptions.value, resultData, filterCourseId.value)
      : mergeFilterOptions(scheduleCourseOptions.value, [...scheduleCourseOptions.value, ...resultData], filterCourseId.value)
    const total = Number(res.total || resultData.length || 0)
    scheduleCoursePagination.value.total = total
    scheduleCourseFinished.value = scheduleCourseOptions.value.length >= total
  }
  catch (error) {
    console.error('load schedule course options failed', error)
  }
}

async function onScheduleTeacherDropdownVisibleChange() {
  await loadScheduleTeacherOptions('', true)
}

async function onScheduleTeacherSearch(keyword) {
  await loadScheduleTeacherOptions(keyword || '', true)
}

async function loadMoreScheduleTeacher() {
  if (scheduleTeacherFinished.value)
    return
  scheduleTeacherPagination.value.current += 1
  await loadScheduleTeacherOptions(scheduleTeacherSearchKey.value, false)
}

async function onScheduleClassroomDropdownVisibleChange() {
  await loadScheduleClassroomOptions('')
}

async function onScheduleClassroomSearch(keyword) {
  await loadScheduleClassroomOptions(keyword || '')
}

async function onScheduleClassDropdownVisibleChange() {
  await loadScheduleClassOptions('', true)
}

async function onScheduleClassSearch(keyword) {
  await loadScheduleClassOptions(keyword || '', true)
}

async function loadMoreScheduleClass() {
  if (scheduleClassFinished.value)
    return
  scheduleClassPagination.value.current += 1
  await loadScheduleClassOptions(scheduleClassSearchKey.value, false)
}

async function onScheduleOneToOneDropdownVisibleChange() {
  await loadScheduleOneToOneOptions('', true)
}

async function onScheduleOneToOneSearch(keyword) {
  await loadScheduleOneToOneOptions(keyword || '', true)
}

async function loadMoreScheduleOneToOne() {
  if (scheduleOneToOneFinished.value)
    return
  scheduleOneToOnePagination.value.current += 1
  await loadScheduleOneToOneOptions(scheduleOneToOneSearchKey.value, false)
}

async function onScheduleCourseDropdownVisibleChange() {
  await loadScheduleCourseOptions('', true)
}

async function onScheduleCourseSearch(keyword) {
  await loadScheduleCourseOptions(keyword || '', true)
}

async function loadMoreScheduleCourse() {
  if (scheduleCourseFinished.value)
    return
  scheduleCoursePagination.value.current += 1
  await loadScheduleCourseOptions(scheduleCourseSearchKey.value, false)
}

function getWeekStart(value = dayjs()) {
  const d = dayjs(value)
  const diff = d.day() === 0 ? -6 : 1 - d.day()
  return d.add(diff, 'day').startOf('day')
}

// 监听时间维度变化
watch(currentTime, () => {
  if (typeof window !== 'undefined') {
    try {
      window.localStorage.setItem(SMART_TIMETABLE_VIEW_MODE_KEY, currentTime.value)
    }
    catch {
    }
  }
  currentWeek.value = dayjs()
})

// 格式化日期显示
function formatDateRange(value) {
  if (!value)
    return ''

  switch (currentTime.value) {
    case 'day':
      return value.format('YYYY年MM月DD日')
    case 'week':
    case 'swapWeek': {
      const start = getWeekStart(value)
      const end = start.add(6, 'day')
      if (start.year() === end.year() && start.month() === end.month()) {
        return `${start.format('YYYY年MM月DD日')} ~ ${end.format('DD日')}`
      }
      else if (start.year() === end.year()) {
        return `${start.format('YYYY年MM月DD日')} ~ ${end.format('MM月DD日')}`
      }
      else {
        return `${start.format('YYYY年MM月DD日')} ~ ${end.format('YYYY年MM月DD日')}`
      }
    }
    case 'month':
      return value.format('YYYY年MM月')
    default:
      return ''
  }
}

// 处理前一个时间段
function handlePrev() {
  switch (currentTime.value) {
    case 'day':
      currentWeek.value = currentWeek.value.subtract(1, 'day')
      break
    case 'week':
    case 'swapWeek':
      currentWeek.value = currentWeek.value.subtract(1, 'week')
      break
    case 'month':
      currentWeek.value = currentWeek.value.subtract(1, 'month')
      break
  }
}

// 处理后一个时间段
function handleNext() {
  switch (currentTime.value) {
    case 'day':
      currentWeek.value = currentWeek.value.add(1, 'day')
      break
    case 'week':
    case 'swapWeek':
      currentWeek.value = currentWeek.value.add(1, 'week')
      break
    case 'month':
      currentWeek.value = currentWeek.value.add(1, 'month')
      break
  }
}

function handleThisWeek() {
  currentWeek.value = dayjs()
}

// 创建一个方法 用于格式化时间xx月-xx日
function formatDate(date) {
  return dayjs(date).format('MM-DD')
}
// 创建一个方法 用于格式化时间为周x，非星期x
function formatWeek(date) {
  const day = dayjs(date).day()
  const weekMap = {
    0: '日',
    1: '一',
    2: '二',
    3: '三',
    4: '四',
    5: '五',
    6: '六',
  }
  return `周${weekMap[day]}`
}
const userStore = useUserStore()
const matrixDays = ref([])
const timetableLoading = ref(false)
const creatingOneToOneSchedule = ref(false)
const deletingScheduledLesson = ref(false)
const forcingConflictSchedule = ref(false)
const conflictDetailModalOpen = ref(false)
const dragConflictDetailOpen = ref(false)
const scheduledConflictDetailOpen = ref(false)
const scheduledConflictDetailLoading = ref(false)
const scheduledConflictDetailValidation = ref(null)
const conflictDetailState = ref({
  summary: '',
  attempted: null,
  items: [],
})
const dragConflictDetailState = ref({
  summary: '',
  attempted: null,
  items: [],
})
const scheduledLessonDetailOpen = ref(false)
const scheduledLessonDetailState = ref({
  modeLabel: '',
  modeColor: '',
  lessonTitle: '',
  dateLabel: '',
  timeLabel: '',
  teacherName: '',
  mainTeacherId: '',
  assistantText: '',
  assistantIds: [],
  classroomId: '',
  groupLabel: '',
  studentText: '',
  courseName: '',
  scheduleId: '',
  courseType: null,
  isMain: true,
  text: null,
  column: null,
  record: null,
})
const draggingScheduleState = ref(null)
const draggingScheduleCellKey = ref('')
const dragPointerState = ref({
  x: 0,
  y: 0,
  visible: false,
})
const dragConfirmOpen = ref(false)
const dragConfirmSubmitting = ref(false)
const dragConfirmDetail = ref({
  source: null,
  target: null,
  payload: null,
})
const updatingDraggedSchedule = ref(false)
const dragHoverState = ref({
  key: '',
  teacherId: '',
  teacherName: '-',
  lessonDate: '',
  startTime: '',
  endTime: '',
  checking: false,
  valid: null,
  label: '',
  message: '',
  conflictTypes: [],
  existingSchedules: [],
})
const dragValidationStateMap = ref({})
const dragValidationCache = new Map()
const dragValidationPromises = new Map()
/** 防止快速切换周次/组别时旧请求晚到覆盖新矩阵 */
let matrixLoadSeq = 0
let pendingConflictJump = null
let focusedScheduleCellTimer = null
let pendingScheduleDragStart = null
let customScheduleDragMoveHandler = null
let customScheduleDragUpHandler = null
let blockedScheduleDragAttempt = null
let blockedScheduleDragMoveHandler = null
let blockedScheduleDragUpHandler = null
let lastBlockedScheduleDragHintAt = 0
let activeDragValidationSessionId = 0
const focusedScheduleCellKey = ref('')
const isSwapTimeGrid = computed(() => currentTime.value === 'swapWeek')
const isWeekLikeView = computed(() => currentTime.value === 'week' || currentTime.value === 'swapWeek')

const displayDates = computed(() => {
  if (currentTime.value === 'day')
    return [dayjs(currentWeek.value).startOf('day')]
  const start = getWeekStart(currentWeek.value)
  return Array.from({ length: 7 }, (_, i) => start.add(i, 'day'))
})

const queryDateRange = computed(() => {
  const dates = displayDates.value
  if (!dates.length) {
    const d = dayjs().format('YYYY-MM-DD')
    return { startDate: d, endDate: d }
  }
  return {
    startDate: dates[0].format('YYYY-MM-DD'),
    endDate: dates[dates.length - 1].format('YYYY-MM-DD'),
  }
})

const periodConfig = computed(() => {
  const parsed = parseUnifiedTimePeriodConfig(userStore.instConfig?.unifiedTimePeriodJson)
  return parsed ?? DEFAULT_UNIFIED_TIME_PERIOD_CONFIG
})

const sortedPeriodGroups = computed(() => configGroupsSorted(periodConfig.value))

const groupOptions = computed(() => {
  const g = sortedPeriodGroups.value
  if (!g.length)
    return [{ key: 'A', label: '默认时段' }]
  if (g.length === 1)
    return [{ key: 'A', label: g[0].name || '时段' }]
  return [
    { key: 'A', label: g[0].name || 'A时段' },
    { key: 'B', label: g[1].name || 'B时段' },
  ]
})

function slotsForGroupKey(key) {
  const groups = sortedPeriodGroups.value
  const fallback = buildQuickHourlySlots().filter(s => s.enabled !== false)
  if (!groups.length)
    return [...fallback].sort((a, b) => a.index - b.index)
  const idx = key === 'B' ? 1 : 0
  const g = groups[idx] || groups[0]
  return [...g.slots].filter(s => s.enabled !== false).sort((a, b) => a.index - b.index)
}

const activePeriodSlots = computed(() => slotsForGroupKey(displayedGroupKey.value))

function periodGroupForKey(key) {
  const groups = sortedPeriodGroups.value
  if (!groups.length)
    return null
  const idx = key === 'B' ? 1 : 0
  return groups[idx] || groups[0] || null
}

/** 矩阵接口：时段组 UUID + 回退 teacherIds（按请求时的组别快照，避免加载途中切换导致参数错位） */
function teacherMatrixGroupParamsForKey(key) {
  const g = periodGroupForKey(key)
  if (!g)
    return {}
  const periodGroupUuid = String(g.id || '').trim()
  const bound = g.boundTeachers
  const ids = Array.isArray(bound)
    ? bound.map(t => String(t.id ?? '').trim()).filter(Boolean)
    : []
  return {
    ...(periodGroupUuid ? { periodGroupUuid } : {}),
    ...(ids.length ? { matrixTeacherIds: ids.join(',') } : {}),
  }
}

function normalizeHHMM(t) {
  const s = String(t || '').trim()
  if (!s)
    return ''
  return s.length >= 5 ? s.slice(0, 5) : s
}

function minutesFromHHMM(t) {
  const n = normalizeHHMM(t)
  const m = /^(\d{1,2}):(\d{2})$/.exec(n)
  if (!m)
    return null
  const h = Number(m[1])
  const mi = Number(m[2])
  if (!Number.isFinite(h) || !Number.isFinite(mi))
    return null
  return h * 60 + mi
}

function emptyLessonCell(slot) {
  return {
    scheduleId: null,
    startTime: slot.start,
    endTime: slot.end,
    courseName: null,
    studentId: null,
    classId: null,
    className: null,
    studentNames: null,
    courseType: null,
    scheduledConflict: false,
    scheduledConflictTypes: [],
    conflict: false,
    conflictReason: null,
    serverConflict: false,
    serverConflictReason: null,
    assistantConflict: false,
    assistantConflictReason: null,
    isMain: undefined,
  }
}

function normalizeFilterOptionId(...values) {
  for (const value of values) {
    const text = String(value ?? '').trim()
    if (text)
      return text
  }
  return ''
}

function resolveLessonScheduleTypeKey(courseType, isMain) {
  if (courseType === 1)
    return 'one_to_one'
  return isMain === false ? 'group_assistant' : 'group_main'
}

function resolveLessonCallStatusKey(legacyItem) {
  const scheduleStatus = Number(legacyItem?.scheduleStatus ?? 0)
  const courseStatus = Number(legacyItem?.courseStatus ?? 0)
  const finishType = Number(legacyItem?.finishType ?? 0)
  if (finishType > 1 || courseStatus > 1 || scheduleStatus === 2)
    return 'signed'
  return 'unsigned'
}

function buildLessonsForRow(slots, legacyList, currentTeacherId) {
  const lessons = slots.map(s => emptyLessonCell(s))
  const list = Array.isArray(legacyList) ? legacyList : []
  for (const leg of list) {
    const st = normalizeHHMM(leg.scheduleStartTime)
    const et = normalizeHHMM(leg.scheduleEndTime)
    const sm = minutesFromHHMM(st)
    const em = minutesFromHHMM(et)
    let idx = lessons.findIndex(l => l.startTime === st && l.endTime === et)
    if (idx < 0 && sm != null && em != null) {
      let best = -1
      let bestOverlap = 0
      lessons.forEach((l, i) => {
        const ls = minutesFromHHMM(l.startTime)
        const le = minutesFromHHMM(l.endTime)
        if (ls == null || le == null)
          return
        const ov = Math.min(le, em) - Math.max(ls, sm)
        if (ov > bestOverlap) {
          bestOverlap = ov
          best = i
        }
      })
      if (best >= 0 && bestOverlap > 0)
        idx = best
    }
    if (idx < 0)
      continue

    const displayCourseType = leg.courseType === 2 ? 1 : 2
    const studentIds = (leg.studentList || []).map(s => String(s.id)).filter(Boolean)
    const names = (leg.studentList || []).map(s => ({ id: String(s.id), name: s.name }))
    const teacherPeople = Array.isArray(leg.teacherList) ? leg.teacherList : []
    const mainTeacher = teacherPeople[0]
    const assistants = teacherPeople.slice(1)
    const assistantNames = assistants.map(item => item?.name).filter(Boolean)
    const assistantIds = assistants.map(item => String(item?.id || '')).filter(Boolean)
    const mainTeacherId = String(mainTeacher?.id || '')
    const isMain = String(currentTeacherId || '') === mainTeacherId
    const scheduleTypeKey = resolveLessonScheduleTypeKey(displayCourseType, isMain)
    const callStatusKey = resolveLessonCallStatusKey(leg)

    lessons[idx] = {
      ...lessons[idx],
      scheduleId: leg.id != null ? String(leg.id) : null,
      courseName: leg.courseName || null,
      courseId: leg.courseId != null ? String(leg.courseId) : '',
      studentId: studentIds.length ? studentIds : null,
      classId: leg.classId != null ? String(leg.classId) : null,
      className: leg.className || null,
      studentNames: names.length ? names : null,
      courseType: displayCourseType,
      scheduledConflict: leg.conflict === true,
      scheduledConflictTypes: leg.conflictTypes || [],
      mainTeacherId,
      teacherName: mainTeacher?.name || null,
      assistantText: assistantNames.length ? assistantNames.join('、') : '未安排',
      assistantIds,
      classroomId: String(leg.classroomId || '').trim(),
      classroomName: leg.classroomName || '',
      isMain,
      scheduleTypeKey,
      callStatusKey,
      scheduleStatus: leg.scheduleStatus ?? null,
      courseStatus: leg.courseStatus ?? null,
      finishType: leg.finishType ?? null,
    }
  }
  return lessons
}

const rawGridRows = computed(() => {
  const slots = activePeriodSlots.value
  const dates = displayDates.value.map(d => d.format('YYYY-MM-DD'))
  if (!slots.length || !dates.length)
    return []

  let teacherCols = []
  for (const d of dates) {
    const day = matrixDays.value.find(x => x.scheduleDate === d)
    if (day?.scheduleListVoList?.length) {
      teacherCols = day.scheduleListVoList
      break
    }
  }
  if (!teacherCols.length && matrixDays.value[0]?.scheduleListVoList)
    teacherCols = matrixDays.value[0].scheduleListVoList || []

  if (!teacherCols.length)
    return []

  const dayMap = new Map()
  for (const day of matrixDays.value) {
    const m = new Map()
    for (const col of day.scheduleListVoList || []) {
      m.set(String(col.teacherId), col.scheduleInfoVoList || [])
    }
    dayMap.set(day.scheduleDate, m)
  }

  const rows = []
  for (const col of teacherCols) {
    const tid = String(col.teacherId)
    const tname = col.teacherName
    for (const dateStr of dates) {
      const schedules = dayMap.get(dateStr)?.get(tid) || []
      rows.push({
        key: `${tid}-${dateStr}`,
        name: tname,
        teacherId: tid,
        date: dateStr,
        lessons: buildLessonsForRow(slots, schedules, tid),
      })
    }
  }
  return rows
})

const scheduleTeacherOptions = ref([])
const scheduleClassroomOptions = ref([])
const scheduleClassOptions = ref([])
const scheduleOneToOneOptions = ref([])
const scheduleCourseOptions = ref([])

const scheduleTeacherFinished = ref(false)
const scheduleClassroomFinished = ref(false)
const scheduleClassFinished = ref(false)
const scheduleOneToOneFinished = ref(false)
const scheduleCourseFinished = ref(false)

const scheduleTeacherPagination = ref({ current: 1, pageSize: 20, total: 0 })
const scheduleClassPagination = ref({ current: 1, pageSize: 20, total: 0 })
const scheduleOneToOnePagination = ref({ current: 1, pageSize: 20, total: 0 })
const scheduleCoursePagination = ref({ current: 1, pageSize: 20, total: 0 })

const scheduleTeacherSearchKey = ref('')
const scheduleClassSearchKey = ref('')
const scheduleOneToOneSearchKey = ref('')
const scheduleCourseSearchKey = ref('')
const scheduleClassroomSearchKey = ref('')

const dataSource = computed(() => {
  return [...rawGridRows.value].sort((a, b) => {
    if (a.teacherId !== b.teacherId)
      return a.teacherId.localeCompare(b.teacherId)
    return a.date.localeCompare(b.date)
  })
})

const transposedDataSource = computed(() => {
  const slots = activePeriodSlots.value
  const dates = displayDates.value.map(d => d.format('YYYY-MM-DD'))
  const rows = []
  const teacherSeen = new Set()

  dataSource.value.forEach((row) => {
    if (teacherSeen.has(row.teacherId))
      return
    teacherSeen.add(row.teacherId)

    slots.forEach((slot, slotIndex) => {
      rows.push({
        key: `${row.teacherId}-slot-${slot.index}`,
        teacherId: row.teacherId,
        name: row.name,
        slotIndex,
        slotLabel: `第${slot.index}节课`,
        startTime: slot.start,
        endTime: slot.end,
        cells: dates.map((date) => {
          const matchedRow = dataSource.value.find(item => item.teacherId === row.teacherId && item.date === date)
          return matchedRow?.lessons?.[slotIndex] || emptyLessonCell(slot)
        }),
      })
    })
  })

  return rows
})

const tableDataSource = computed(() => isSwapTimeGrid.value ? transposedDataSource.value : dataSource.value)

function uniqueConflictTypes(list) {
  return Array.from(new Set((Array.isArray(list) ? list : []).map(item => String(item || '').trim()).filter(Boolean)))
}

function uniqueExistingSchedules(list) {
  const map = new Map()
  ;(Array.isArray(list) ? list : []).forEach((item) => {
    const key = [
      String(item?.teacherId || '').trim(),
      String(item?.date || '').trim(),
      String(item?.timeText || '').trim(),
      String(item?.name || '').trim(),
    ].join('|')
    if (!map.has(key))
      map.set(key, item)
  })
  return [...map.values()]
}

function mergeAvailabilityConflictReasons(reasons) {
  const normalized = (Array.isArray(reasons) ? reasons : []).filter(Boolean)
  const existingSchedules = uniqueExistingSchedules(normalized.flatMap(reason => reason?.existingSchedules || []))
  const conflictTypes = uniqueConflictTypes(normalized.flatMap(reason => reason?.conflictTypes || []))
  const messages = Array.from(new Set(normalized.map(reason => String(reason?.message || '').trim()).filter(Boolean)))
  return {
    type: existingSchedules.length ? '1v1-api' : (normalized[0]?.type || '1v1-api'),
    message: messages.join('；') || '该时间段不可排课',
    conflictTypes,
    existingSchedules,
  }
}

function syncLessonConflictState(lesson) {
  const reasons = []
  if (lesson.serverConflict && lesson.serverConflictReason)
    reasons.push(lesson.serverConflictReason)
  if (lesson.assistantConflict && lesson.assistantConflictReason)
    reasons.push(lesson.assistantConflictReason)

  if (!reasons.length) {
    lesson.conflict = false
    lesson.conflictReason = null
    return
  }

  lesson.conflict = true
  lesson.conflictReason = reasons.length === 1 ? reasons[0] : mergeAvailabilityConflictReasons(reasons)
}

function clearLessonConflictState(lesson, scope = 'all') {
  if (scope === 'all' || scope === 'server') {
    lesson.serverConflict = false
    lesson.serverConflictReason = null
  }
  if (scope === 'all' || scope === 'assistant') {
    lesson.assistantConflict = false
    lesson.assistantConflictReason = null
  }
  if (scope === 'all') {
    lesson.conflict = false
    lesson.conflictReason = null
  }
  else {
    syncLessonConflictState(lesson)
  }
}

function resetEmptyLessonConflicts(scope = 'all') {
  dataSource.value.forEach((teacher) => {
    teacher.lessons.forEach((lesson) => {
      if (!lesson.studentId)
        clearLessonConflictState(lesson, scope)
    })
  })
}

/** 当前展示范围内每位老师已占用的节次数（与格子里蓝色已排课一致：有 studentId 即计入） */
const scheduledLessonCountByTeacher = computed(() => {
  const map = new Map()
  for (const row of dataSource.value) {
    const tid = String(row.teacherId)
    let n = 0
    for (const lesson of row.lessons || []) {
      if (lesson.studentId)
        n++
    }
    map.set(tid, (map.get(tid) || 0) + n)
  }
  return map
})

function teacherScheduledLessonCount(teacherId) {
  if (teacherId == null)
    return 0
  return scheduledLessonCountByTeacher.value.get(String(teacherId)) ?? 0
}

function teacherLessonCountLabel(teacherId) {
  const n = teacherScheduledLessonCount(teacherId)
  const scope = currentTime.value === 'day' ? '当日' : '本周'
  return `${scope}共${n}节课`
}

const activeGroupLabel = computed(() => {
  return groupOptions.value.find(o => o.key === displayedGroupKey.value)?.label || ''
})

let detectOneToOneAvailabilityBridge = (_value) => {}
let handleClassBridge = (_value) => {}

const {
  assistantNameById,
  fetchAssistantOptions,
  fetchOneToOneOptionsForTimetable,
  filterOneToOneOption,
  handle1v1,
  handleOneToOneDropdownVisibleChange,
  normalizedSelectedAssistantIds,
  oneToOneData,
  oneToOneListLoading,
  oneToOnePickerOpen,
  oneToOneRecordId,
  renderOneToOneDropdown,
  resetOneToOnePickerState,
  selectedAssistantIds,
  selectedAssistantText,
} = useSmartTimetablePicker({
  activeGroupLabel,
  currentModel,
  displayedGroupKey,
  detectOneToOneAvailability: value => detectOneToOneAvailabilityBridge(value),
  periodGroupForKey,
  resetAssistantConflicts: () => resetEmptyLessonConflicts('assistant'),
})

const {
  cancelOneToOneAvailabilityCheck,
  detectOneToOneAvailability,
  oneToOneAvailabilityLoading,
} = useSmartTimetableAvailability({
  assistantNameById,
  buildAvailabilitySlotKey,
  currentModel,
  dataSource,
  normalizedSelectedAssistantIds,
  oneToOneData,
  parseConflictTimeRange,
  resetEmptyLessonConflicts,
  syncLessonConflictState,
  uniqueConflictTypes,
  uniqueExistingSchedules,
})

detectOneToOneAvailabilityBridge = value => detectOneToOneAvailability(value)

watch(
  groupOptions,
  (opts) => {
    if (!opts.some(o => o.key === currentGroup.value))
      currentGroup.value = 'A'
  },
  { immediate: true },
)

async function loadTimetableMatrix() {
  const seq = ++matrixLoadSeq
  const requestedGroup = currentGroup.value
  timetableLoading.value = true
  try {
    await userStore.getInstConfig()
    if (seq !== matrixLoadSeq)
      return
    const { startDate, endDate } = queryDateRange.value
    const scheduleTeacherIds = filterTeacherId.value.join(',')
    const classroomIds = filterClassroomId.value.join(',')
    const res = await listTeachingSchedulesByTeacherMatrixApi({
      startDate,
      endDate,
      studentId: filterStudentId.value,
      scheduleTeacherIds: scheduleTeacherIds || undefined,
      classroomIds: classroomIds || undefined,
      groupClassIds: filterClassId.value ? String(filterClassId.value) : undefined,
      oneToOneClassIds: filterOneToOneId.value ? String(filterOneToOneId.value) : undefined,
      lessonIds: filterCourseId.value ? String(filterCourseId.value) : undefined,
      scheduleTypes: filterScheduleType.value.length ? filterScheduleType.value.join(',') : undefined,
      callStatuses: filterCallStatus.value ? String(filterCallStatus.value) : undefined,
      ...teacherMatrixGroupParamsForKey(requestedGroup),
    })
    if (seq !== matrixLoadSeq)
      return
    if (res.code === 200 && Array.isArray(res.result))
      matrixDays.value = res.result
    else
      matrixDays.value = []
    displayedGroupKey.value = requestedGroup
    await nextTick()
    if (seq !== matrixLoadSeq)
      return
    if (currentModel.value === '1' && oneToOneRecordId.value)
      await detectOneToOneAvailability(oneToOneRecordId.value)
    else if (currentModel.value === '2' && classId.value)
      handleClassBridge(classId.value)
    else
      resetEmptyLessonConflicts()
    await flushPendingConflictJump()
  }
  catch (e) {
    console.error('loadTimetableMatrix failed', e)
  }
  finally {
    if (seq === matrixLoadSeq)
      timetableLoading.value = false
  }
}

watch(
  [currentWeek, () => (currentTime.value === 'swapWeek' ? 'week' : currentTime.value), currentModel, currentGroup],
  () => {
    void loadTimetableMatrix()
  },
)

watch(
  [filterStudentId, filterTeacherId, filterClassroomId, filterClassId, filterOneToOneId, filterCourseId, filterScheduleType, filterCallStatus],
  () => {
    void loadTimetableMatrix()
  },
  { deep: true },
)

function refreshTimetableRelatedData() {
  void loadTimetableMatrix()
  void fetchOneToOneOptionsForTimetable()
  void fetchAssistantOptions()
}

onMounted(() => {
  void loadTimetableMatrix()
  void fetchOneToOneOptionsForTimetable()
  void fetchAssistantOptions()
  emitter.on(EVENTS.REFRESH_DATA, refreshTimetableRelatedData)
})

onUnmounted(() => {
  emitter.off(EVENTS.REFRESH_DATA, refreshTimetableRelatedData)
  clearCustomScheduleDragListeners()
  clearBlockedScheduleDragAttempt()
  if (focusedScheduleCellTimer)
    clearTimeout(focusedScheduleCellTimer)
})
const columns = computed(() => {
  const slots = activePeriodSlots.value
  const teacherColumn = {
    title: '教师',
    dataIndex: 'name',
    key: 'name',
    width: 120,
    align: 'center',
    fixed: 'left',
    customCell: (_, index) => {
      if (!tableDataSource.value.length)
        return {}
      const currentTeacherId = tableDataSource.value[index].teacherId
      if (index === 0 || tableDataSource.value[index - 1].teacherId !== currentTeacherId) {
        let count = 1
        for (let i = index + 1; i < tableDataSource.value.length; i++) {
          if (tableDataSource.value[i].teacherId === currentTeacherId)
            count++
          else
            break
        }
        return { rowSpan: count }
      }
      return { rowSpan: 0 }
    },
  }

  if (isSwapTimeGrid.value) {
    const slotColumn = {
      title: '节次',
      dataIndex: 'slotLabel',
      key: 'slot',
      width: 120,
      fixed: 'left',
      align: 'center',
    }

    const dateColumns = displayDates.value.map((date, index) => ({
      title: formatWeek(date.format('YYYY-MM-DD')),
      date: date.format('YYYY-MM-DD'),
      dateText: formatDate(date),
      dataIndex: ['cells', index],
      key: `date-${date.format('YYYY-MM-DD')}`,
      width: 160,
      align: 'center',
    }))

    return [teacherColumn, slotColumn, ...dateColumns]
  }

  const baseColumns = [
    teacherColumn,
    {
      title: '日期',
      dataIndex: 'date',
      key: 'date',
      width: 80,
      fixed: 'left',
      align: 'center',
    },
  ]

  const lessonColumns = slots.map((slot, index) => ({
    title: `第${slot.index}节课`,
    startTime: slot.start,
    endTime: slot.end,
    dataIndex: ['lessons', index],
    key: `lesson-${slot.index}-${index}`,
    width: 160,
    align: 'center',
  }))

  return [...baseColumns, ...lessonColumns]
})

function isScheduleColumn(column) {
  const key = column?.dataIndex?.[0]
  return key === 'lessons' || key === 'cells'
}

function scheduleCellDate(column, record) {
  return isSwapTimeGrid.value ? column?.date : record?.date
}

function scheduleCellStartTime(column, record) {
  return isSwapTimeGrid.value ? record?.startTime : column?.startTime
}

function scheduleCellEndTime(column, record) {
  return isSwapTimeGrid.value ? record?.endTime : column?.endTime
}

function scheduleCellTeacherName(record) {
  return record?.name || '-'
}

function scheduleCellKey(column, record) {
  return buildAvailabilitySlotKey(
    record?.teacherId,
    scheduleCellDate(column, record),
    scheduleCellStartTime(column, record),
    scheduleCellEndTime(column, record),
  )
}

function scheduleCellContextRecord(column, record) {
  return {
    ...record,
    date: scheduleCellDate(column, record),
    name: scheduleCellTeacherName(record),
  }
}

function scheduleCellContextColumn(column, record) {
  return {
    ...column,
    startTime: scheduleCellStartTime(column, record),
    endTime: scheduleCellEndTime(column, record),
  }
}
// 课程列表数据
const courseList = ref([
  {
    id: '589251114063479808',
    name: '初级认知课',
    courseType: 1,
  },
  {
    id: '58925112157479108',
    name: '初级感统课',
    courseType: 1,
  },
  {
    id: '589251121574791081',
    name: 'PT治疗课',
    courseType: 1,
  },
  {
    id: '589251121574791082',
    name: 'OT精细课',
    courseType: 1,
  },
  {
    id: '589251121574791083',
    name: '口肌训练课',
    courseType: 1,
  },
  {
    id: '589251121574791084',
    name: '初级认知课',
    courseType: 2,
  },
])
// 当前视图下的全部行（时段 A/B 切换后数据源已重建；跨组检测以当前页为准）
const allDataSource = computed(() => dataSource.value)

const {
  applyClassSchedule,
  classData,
  findClassInfo,
  handleClass,
  resolveClassConflictMessage,
  resolveSelectedClassTarget,
} = useSmartTimetableClassMode({
  activeGroupLabel,
  allDataSource,
  dataSource,
  getLessonIndex,
  resetEmptyLessonConflicts,
})

handleClassBridge = value => handleClass(value)

function buildAvailabilitySlotKey(teacherId, lessonDate, startTime, endTime) {
  return `${String(teacherId)}|${lessonDate}|${startTime}|${endTime}`
}

function parseConflictTimeRange(timeText) {
  const m = String(timeText || '').trim().match(/(\d{2}:\d{2})\s*[~-]\s*(\d{2}:\d{2})/)
  if (!m)
    return null
  return { startTime: m[1], endTime: m[2] }
}

function resolveConflictScheduleGroupInfo(item) {
  const teacherId = String(item?.teacherId || '').trim()
  const timeRange = parseConflictTimeRange(item?.timeText)
  const matches = groupOptions.value
    .map((opt) => {
      const group = periodGroupForKey(opt.key)
      const teacherMatched = !teacherId || !(group?.boundTeachers?.length)
        || group.boundTeachers.some(t => String(t.id) === teacherId)
      const timeMatched = !timeRange || slotsForGroupKey(opt.key).some(slot =>
        slot.start === timeRange.startTime && slot.end === timeRange.endTime,
      )
      return teacherMatched && timeMatched
        ? { key: opt.key, label: opt.label }
        : null
    })
    .filter(Boolean)

  const unique = matches.filter((item, index, arr) =>
    arr.findIndex(x => x.key === item.key) === index,
  )
  return {
    keys: unique.map(item => item.key),
    labels: unique.map(item => item.label),
  }
}

function resolveConflictScheduleGroupLabel(item) {
  return resolveConflictScheduleGroupInfo(item).labels.join('/')
}

function setFocusedScheduleCell(key) {
  focusedScheduleCellKey.value = key || ''
  if (focusedScheduleCellTimer)
    clearTimeout(focusedScheduleCellTimer)
  if (key) {
    focusedScheduleCellTimer = window.setTimeout(() => {
      if (focusedScheduleCellKey.value === key)
        focusedScheduleCellKey.value = ''
    }, 3000)
  }
}

async function focusScheduleCell(key) {
  await nextTick()
  const root = timetableRootRef.value
  if (!root || !key)
    return false
  const cell = Array.from(root.querySelectorAll('[data-schedule-cell-key]')).find(
    el => el.getAttribute('data-schedule-cell-key') === key,
  )
  if (!cell)
    return false
  cell.scrollIntoView({
    behavior: 'smooth',
    block: 'center',
    inline: 'center',
  })
  setFocusedScheduleCell(key)
  return true
}

async function flushPendingConflictJump() {
  if (!pendingConflictJump?.cellKey)
    return
  const pending = pendingConflictJump
  const found = await focusScheduleCell(pending.cellKey)
  if (found) {
    pendingConflictJump = null
    messageService.success('已定位到冲突课程')
  }
}

function buildConflictDetailItems(reason, fallbackConflictTypes = ['时间']) {
  const existingSchedules = Array.isArray(reason?.existingSchedules) ? reason.existingSchedules : []
  return existingSchedules.map((item, index) => {
    const groupInfo = resolveConflictScheduleGroupInfo(item)
    const timeRange = parseConflictTimeRange(item.timeText)
    const conflictTypes = Array.isArray(item.conflictTypes) && item.conflictTypes.length ? item.conflictTypes : fallbackConflictTypes
    const jumpGroupKey = groupInfo.keys.includes(currentGroup.value)
      ? currentGroup.value
      : groupInfo.keys[0] || ''
    return {
      key: `${item.teacherId || item.teacherName || 'teacher'}-${item.date}-${item.timeText}-${index}`,
      name: item.name || '-',
      classTypeText: item.classTypeText || '日程',
      date: item.date,
      week: item.week || '',
      timeText: item.timeText,
      teacherId: item.teacherId || '',
      teacherName: item.teacherName || '-',
      groupLabel: groupInfo.labels.join('/') || '未知组别',
      classroomName: item.classroomName || '-',
      assistantText: (item.assistantNames || []).join('、') || '-',
      studentText: (item.studentNames || []).join('、') || '-',
      conflictTypes,
      hasTeacherConflict: conflictTypes.includes('老师'),
      hasAssistantConflict: conflictTypes.includes('助教'),
      hasStudentConflict: conflictTypes.includes('学员'),
      hasClassroomConflict: conflictTypes.includes('教室'),
      jumpCellKey: timeRange && item.teacherId
        ? buildAvailabilitySlotKey(item.teacherId, item.date, timeRange.startTime, timeRange.endTime)
        : '',
      jumpGroupKey,
    }
  })
}

function openConflictDetailModalWithAttempt(reason, attempted) {
  const attemptedConflictTypes = Array.isArray(reason?.conflictTypes) ? reason.conflictTypes : []
  const fallbackConflictTypes = attemptedConflictTypes.length ? attemptedConflictTypes : ['时间']
  const items = buildConflictDetailItems(reason, fallbackConflictTypes)
  conflictDetailState.value = {
    summary: `${reason.message || '当前空位存在时间冲突'}，共发现 ${items.length} 条冲突日程。`,
    attempted,
    items,
  }
  conflictDetailModalOpen.value = true
}

function openDragConflictDetailModal(validation, dragState, target) {
  const conflictTypes = Array.isArray(validation?.conflictTypes) ? validation.conflictTypes : []
  const fallbackConflictTypes = conflictTypes.length ? conflictTypes : ['时间']
  const items = buildConflictDetailItems({
    message: validation?.message,
    conflictTypes,
    existingSchedules: validation?.existingSchedules,
  }, fallbackConflictTypes)

  dragConflictDetailState.value = {
    summary: `${validation?.message || '当前调课存在时间冲突'}，共发现 ${items.length} 条冲突日程。`,
    attempted: {
      modeLabel: '1v1',
      studentText: dragState?.studentText || '未识别学员',
      courseName: dragState?.courseName || '未识别课程',
      date: target?.lessonDate || '',
      week: formatWeek(target?.lessonDate || ''),
      timeText: `${target?.startTime || ''}-${target?.endTime || ''}`,
      teacherName: target?.teacherName || '-',
      assistantText: dragState?.assistantText || '未安排',
      lessonIndex: getLessonIndex(target?.startTime || ''),
      groupLabel: activeGroupLabel.value || '当前组',
    },
    items,
  }
  dragConflictDetailOpen.value = true
}

function openApiConflictModal(reason, column, record) {
  const selectedTarget = resolveConflictAttemptTarget()
  const attemptedConflictTypes = Array.isArray(reason?.conflictTypes) ? reason.conflictTypes : []
  const forceAllowed = currentModel.value === '1'
    && Boolean(oneToOneRecordId.value)
    && attemptedConflictTypes.length > 0
    && attemptedConflictTypes.every(type => type === '学员')

  openConflictDetailModalWithAttempt(reason, {
    modeLabel: selectedTarget.modeLabel,
    targetLabel: selectedTarget.targetLabel,
    targetValue: selectedTarget.targetValue,
    courseName: selectedTarget.courseName,
    date: record.date,
    week: formatWeek(record.date),
    timeText: `${column.startTime}-${column.endTime}`,
    teacherName: record.name,
    assistantText: selectedAssistantText.value,
    lessonIndex: getLessonIndex(column.startTime),
    groupLabel: activeGroupLabel.value || '当前组',
    forceAllowed,
    forcePayload: forceAllowed
      ? {
          oneToOneId: String(oneToOneRecordId.value),
          teacherId: String(record.teacherId),
          assistantIds: normalizedSelectedAssistantIds.value,
          schedules: [{
            lessonDate: record.date,
            startTime: column.startTime,
            endTime: column.endTime,
            assistantIds: normalizedSelectedAssistantIds.value,
          }],
        }
      : null,
  })
}

async function forceScheduleDespiteStudentConflict() {
  const attempted = conflictDetailState.value.attempted
  if (!attempted?.forceAllowed || !attempted?.forcePayload) {
    messageService.warning('当前冲突类型暂不支持强制排课')
    return
  }

  forcingConflictSchedule.value = true
  try {
    const schedules = Array.isArray(attempted.forcePayload.schedules)
      ? attempted.forcePayload.schedules.map(item => ({
          ...item,
          allowStudentConflict: true,
        }))
      : []
    const res = await createOneToOneSchedulesApi({
      ...attempted.forcePayload,
      allowStudentConflict: true,
      schedules,
    })
    if (res.code !== 200)
      throw new Error(res.message || '强制排课失败')
    conflictDetailModalOpen.value = false
    messageService.success('已按学员冲突方式排课，课表将标记冲突')
    emitter.emit(EVENTS.REFRESH_DATA)
  }
  catch (error) {
    console.error('force schedule despite student conflict failed', error)
    messageService.error(error?.response?.data?.message || error?.message || '强制排课失败')
  }
  finally {
    forcingConflictSchedule.value = false
  }
}

async function openScheduledConflictDetail(text) {
  if (!text?.scheduledConflict || !text?.scheduleId) {
    messageService.warning('当前课程暂无可查看的冲突详情')
    return
  }

  scheduledConflictDetailLoading.value = true
  try {
    const res = await getTeachingScheduleConflictDetailApi({
      id: String(text.scheduleId),
    })
    if (res.code !== 200 || !res.result)
      throw new Error(res.message || '加载冲突详情失败')
    scheduledConflictDetailValidation.value = res.result
    scheduledConflictDetailOpen.value = true
  }
  catch (error) {
    console.error('openScheduledConflictDetail failed', error)
    messageService.error(error?.response?.data?.message || error?.message || '加载冲突详情失败')
  }
  finally {
    scheduledConflictDetailLoading.value = false
  }
}

async function jumpToConflictSchedule(item) {
  if (!item?.jumpCellKey) {
    messageService.warning('当前冲突课程暂不支持定位')
    return
  }

  conflictDetailModalOpen.value = false
  dragConflictDetailOpen.value = false
  pendingConflictJump = {
    cellKey: item.jumpCellKey,
  }

  let needReload = false
  if (item.jumpGroupKey && item.jumpGroupKey !== currentGroup.value) {
    currentGroup.value = item.jumpGroupKey
    needReload = true
  }

  const targetDate = String(item.date || '').trim()
  if (targetDate) {
    if (currentTime.value === 'day') {
      if (dayjs(currentWeek.value).format('YYYY-MM-DD') !== targetDate) {
        currentWeek.value = dayjs(targetDate)
        needReload = true
      }
    }
    else {
      const { startDate, endDate } = queryDateRange.value
      if (targetDate < startDate || targetDate > endDate) {
        currentWeek.value = dayjs(targetDate)
        needReload = true
      }
    }
  }

  if (!needReload) {
    const found = await focusScheduleCell(item.jumpCellKey)
    if (found) {
      pendingConflictJump = null
      messageService.success('已定位到冲突课程')
    }
    else {
      await loadTimetableMatrix()
    }
  }
}

function resolveConflictAttemptTarget() {
  if (currentModel.value === '1') {
    const selectedOneToOne = oneToOneData.value.find(item => item.id === String(oneToOneRecordId.value))
    return {
      modeLabel: '1v1',
      targetLabel: '排课学员',
      targetValue: selectedOneToOne?.studentName || '未选择学员',
      courseName: selectedOneToOne?.courseName || '未选择课程',
    }
  }

  return resolveSelectedClassTarget(classId.value)
}

// 处理冲突点击
function handleConflictClick(timeSlot, column, record) {
  let content = '该时间段已有课程安排，无法排课'

  // 根据冲突原因提供更详细的信息
  if (timeSlot.conflictReason) {
    const reason = timeSlot.conflictReason

    if (reason.type === '1v1-api') {
      openApiConflictModal(reason, column, record)
      return
    }
    else if (reason.type === '1v1-assistant-selection') {
      content = reason.message || '已选助教与当前上课老师重复，请调整助教后再排课'
    }
    else {
      content = resolveClassConflictMessage(reason) || content
    }
  }

  Modal.info({
    title: '时间冲突',
    content,
  })
}

function buildConfirmField(label, value, valueColor = '#1f2329') {
  return h('div', {
    style: {
      display: 'grid',
      gridTemplateColumns: '76px 1fr',
      gap: '10px',
      alignItems: 'start',
      fontSize: '14px',
      lineHeight: '22px',
    },
  }, [
    h('div', {
      style: {
        color: '#8c8c8c',
      },
    }, label),
    h('div', {
      style: {
        color: valueColor,
        fontWeight: 600,
        wordBreak: 'break-word',
      },
    }, value || '-'),
  ])
}

const SMART_TIMETABLE_SKIP_CONFIRM_KEY = 'smart-timetable-skip-confirm-date'

function todayScheduleConfirmKey() {
  return dayjs().format('YYYY-MM-DD')
}

function shouldSkipScheduleConfirmToday() {
  if (typeof window === 'undefined')
    return false
  try {
    return window.localStorage.getItem(SMART_TIMETABLE_SKIP_CONFIRM_KEY) === todayScheduleConfirmKey()
  }
  catch {
    return false
  }
}

function setSkipScheduleConfirmToday(enabled) {
  if (typeof window === 'undefined')
    return
  try {
    if (enabled)
      window.localStorage.setItem(SMART_TIMETABLE_SKIP_CONFIRM_KEY, todayScheduleConfirmKey())
    else
      window.localStorage.removeItem(SMART_TIMETABLE_SKIP_CONFIRM_KEY)
  }
  catch {
  }
}

function buildScheduleConfirmContent({
  modeLabel,
  modeColor,
  targetLabel,
  targetValue,
  courseName,
  dateLabel,
  timeLabel,
  teacherName,
  assistantText,
  groupLabel,
  onSkipTodayChange,
}) {
  return h('div', {
    style: {
      display: 'flex',
      flexDirection: 'column',
      gap: '14px',
      marginTop: '4px',
    },
  }, [
    h('div', {
      style: {
        display: 'flex',
        alignItems: 'center',
        gap: '10px',
        flexWrap: 'wrap',
      },
    }, [
      h('span', {
        style: {
          display: 'inline-flex',
          alignItems: 'center',
          justifyContent: 'center',
          minWidth: '52px',
          height: '28px',
          padding: '0 10px',
          borderRadius: '999px',
          background: modeColor,
          color: '#fff',
          fontSize: '13px',
          fontWeight: 700,
        },
      }, modeLabel),
      h('span', {
        style: {
          color: '#262626',
          fontSize: '16px',
          fontWeight: 700,
        },
      }, `${targetValue}${courseName ? ` · ${courseName}` : ''}`),
    ]),
    h('div', {
      style: {
        padding: '14px 16px',
        borderRadius: '14px',
        background: '#f8fafc',
        border: '1px solid #edf2f7',
        display: 'flex',
        flexDirection: 'column',
        gap: '10px',
      },
    }, [
      buildConfirmField(targetLabel, targetValue),
      buildConfirmField('上课时间', `${dateLabel} · ${timeLabel}`, '#1677ff'),
      buildConfirmField('上课老师', teacherName),
      ...(assistantText != null ? [buildConfirmField('上课助教', assistantText)] : []),
      buildConfirmField('所在组别', groupLabel || '当前组'),
    ]),
    h('div', {
      style: {
        padding: '12px 14px',
        borderRadius: '12px',
        background: '#fff7e6',
        color: '#ad6800',
        fontSize: '13px',
        lineHeight: '22px',
      },
    }, '确认后将立即创建日程并占用该时段；如果此时课表已被别人占用，系统会再次拦截。'),
    h('label', {
      style: {
        display: 'inline-flex',
        alignItems: 'center',
        gap: '8px',
        cursor: 'pointer',
        color: '#595959',
        fontSize: '13px',
        userSelect: 'none',
      },
    }, [
      h('input', {
        type: 'checkbox',
        onChange: event => onSkipTodayChange?.(Boolean(event?.target?.checked)),
        style: {
          width: '14px',
          height: '14px',
          cursor: 'pointer',
        },
      }),
      h('span', null, '今日不再提示'),
    ]),
  ])
}

function confirmScheduleWithOptionalSkip({
  modeLabel,
  modeColor,
  targetLabel,
  targetValue,
  courseName,
  dateLabel,
  timeLabel,
  teacherName,
  assistantText,
  groupLabel,
  onConfirm,
}) {
  if (shouldSkipScheduleConfirmToday()) {
    return Promise.resolve(onConfirm())
  }

  let skipToday = false
  return new Promise((resolve, reject) => {
    Modal.confirm({
      title: '确认排课',
      width: 620,
      okText: '确认排课',
      cancelText: '再想想',
      content: buildScheduleConfirmContent({
        modeLabel,
        modeColor,
        targetLabel,
        targetValue,
        courseName,
        dateLabel,
        timeLabel,
        teacherName,
        assistantText,
        groupLabel,
        onSkipTodayChange: (checked) => { skipToday = checked },
      }),
      async onOk() {
        if (skipToday)
          setSkipScheduleConfirmToday(true)
        try {
          const result = await onConfirm()
          resolve(result)
          return result
        }
        catch (error) {
          reject(error)
          throw error
        }
      },
      onCancel() {
        resolve(false)
      },
    })
  })
}

function buildScheduledLessonCancelConfirmTitle(text, column, record) {
  const dateObj = dayjs(record.date)
  const month = dateObj.format('M')
  const day = dateObj.format('D')
  const lessonIndex = getLessonIndex(column.startTime)
  const studentText = Array.isArray(text.studentNames)
    ? text.studentNames.map(item => item.name).filter(Boolean).join('、')
    : '-'

  return h('div', {
    style: {
      display: 'flex',
      flexDirection: 'column',
      gap: '6px',
      maxWidth: '360px',
      lineHeight: '20px',
    },
  }, [
    h('div', {
      style: {
        color: '#262626',
        fontSize: '14px',
        fontWeight: 700,
      },
    }, '撤销这节 1v1 课程？'),
    h('div', {
      style: {
        color: '#595959',
        fontSize: '13px',
      },
    }, `${month}月${day}日 ${formatWeek(record.date)} 第${lessonIndex}节 · ${column.startTime}-${column.endTime}`),
    h('div', {
      style: {
        color: '#595959',
        fontSize: '13px',
      },
    }, `老师：${record.name} · ${activeGroupLabel.value || '当前组'} · 学员：${studentText}`),
    h('div', {
      style: {
        color: '#8c8c8c',
        fontSize: '12px',
      },
    }, text.courseName || '课程'),
  ])
}

function openScheduledLessonDetail(text, column, record) {
  const dateObj = dayjs(record.date)
  const month = dateObj.format('M')
  const day = dateObj.format('D')
  const lessonIndex = getLessonIndex(column.startTime)
  const studentText = Array.isArray(text.studentNames)
    ? text.studentNames.map(item => item.name).filter(Boolean).join('、')
    : '-'

  scheduledLessonDetailState.value = {
    modeLabel: text.courseType === 1 ? '1v1' : '班课',
    modeColor: text.courseType === 1 ? '#1677ff' : '#13c2c2',
    lessonTitle: `${studentText || '学员'} · ${text.courseName || '课程'}`,
    dateLabel: `${month}月${day}日 ${formatWeek(record.date)} 第${lessonIndex}节`,
    timeLabel: `${column.startTime}-${column.endTime}`,
    teacherName: text.teacherName || record.name,
    mainTeacherId: String(text.mainTeacherId || ''),
    assistantText: text.assistantText || '未安排',
    assistantIds: Array.isArray(text.assistantIds) ? text.assistantIds : [],
    classroomId: String(text.classroomId || ''),
    groupLabel: activeGroupLabel.value || '当前组',
    studentText: studentText || '-',
    courseName: text.courseName || '',
    scheduleId: String(text.scheduleId || ''),
    courseType: text.courseType,
    isMain: text.isMain !== false,
    text,
    column,
    record,
  }
  scheduledLessonDetailOpen.value = true
}

async function deleteScheduledLessonFromDetail() {
  const detail = scheduledLessonDetailState.value
  if (detail.courseType !== 1) {
    messageService.info('班课删除建议和主教/辅教联动一起设计，这里先不直接删除。')
    return
  }
  if (!detail.scheduleId) {
    messageService.warning('当前日程缺少可撤销标识，请刷新后重试')
    return
  }

  const dateObj = dayjs(detail.record?.date)
  const month = dateObj.format('M')
  const day = dateObj.format('D')
  const lessonIndex = getLessonIndex(detail.column?.startTime)

  deletingScheduledLesson.value = true
  try {
    if (detail.isMain === false) {
      const nextAssistantIds = (detail.assistantIds || []).filter(id => String(id) !== String(detail.record?.teacherId || ''))
      const res = await batchUpdateTeachingSchedulesApi({
        ids: [detail.scheduleId],
        teacherId: detail.mainTeacherId,
        assistantIds: nextAssistantIds,
        classroomId: detail.classroomId || '',
        startTime: detail.column?.startTime,
        endTime: detail.column?.endTime,
      })
      if (res.code !== 200)
        throw new Error(res.message || '移除助教失败')
      scheduledLessonDetailOpen.value = false
      messageService.success(`已移除 ${month}月${day}日 第${lessonIndex}节课的助教`)
      emitter.emit(EVENTS.REFRESH_DATA)
      return
    }

    const res = await cancelTeachingSchedulesApi({
      ids: [detail.scheduleId],
    })
    if (res.code !== 200)
      throw new Error(res.message || '删除日程失败')
    scheduledLessonDetailOpen.value = false
    messageService.success(`已删除 ${month}月${day}日 第${lessonIndex}节 1v1 日程，主教/助教课表已同步移除`)
    emitter.emit(EVENTS.REFRESH_DATA)
  }
  catch (error) {
    console.error('cancel teaching schedule failed', error)
    messageService.error(error?.response?.data?.message || error?.message || '删除日程失败')
    throw error
  }
  finally {
    deletingScheduledLesson.value = false
  }
}

// 排课
function handleScheduleClick(timeSlot, column, record) {
  if (currentModel.value === '1') {
    if (!oneToOneRecordId.value) {
      Modal.warning({
        title: '请先选择一对一',
        content: '请先在上方选择要排课的 1 对 1 记录',
      })
      return
    }

    const studentInfo = oneToOneData.value.find(
      item => item.id === String(oneToOneRecordId.value),
    )

    if (!studentInfo) {
      Modal.warning({
        title: '记录不存在',
        content: '所选 1 对 1 已不在列表中，请重新选择或刷新页面',
      })
      return
    }

    // 获取月份和日期信息
    const dateObj = dayjs(record.date)
    const month = dateObj.format('M')
    const day = dateObj.format('D')
    const lessonIndex = getLessonIndex(column.startTime)

    void confirmScheduleWithOptionalSkip({
      modeLabel: '1v1',
      modeColor: '#1677ff',
      targetLabel: '排课对象',
      targetValue: studentInfo.studentName,
      courseName: studentInfo.courseName,
      dateLabel: `${month}月${day}日 ${formatWeek(record.date)} 第${lessonIndex}节`,
      timeLabel: `${column.startTime}-${column.endTime}`,
      teacherName: record.name,
      assistantText: selectedAssistantText.value,
      groupLabel: activeGroupLabel.value || '当前组',
      async onConfirm() {
        creatingOneToOneSchedule.value = true
        try {
          const res = await createOneToOneSchedulesApi({
            oneToOneId: String(oneToOneRecordId.value),
            teacherId: String(record.teacherId),
            assistantIds: normalizedSelectedAssistantIds.value,
            schedules: [{
              lessonDate: record.date,
              startTime: column.startTime,
              endTime: column.endTime,
              assistantIds: normalizedSelectedAssistantIds.value,
            }],
          })
          if (res.code !== 200)
            throw new Error(res.message || '创建1对1日程失败')

          messageService.success(`已为 ${studentInfo.studentName} 创建 ${month}月${day}日 第${lessonIndex}节课`)
          emitter.emit(EVENTS.REFRESH_DATA)
        }
        catch (error) {
          console.error('create one-to-one schedule failed', error)
          messageService.error(error?.response?.data?.message || error?.message || '创建1对1日程失败')
          await loadTimetableMatrix()
        }
        finally {
          creatingOneToOneSchedule.value = false
        }
      },
    })
  }
  else {
    // 班课排课逻辑
    if (!classId.value) {
      Modal.warning({
        title: '请先选择班级',
        content: '请先在上方选择要排课的班级',
      })
      return
    }

    const classInfo = findClassInfo(classId.value)

    if (!classInfo) {
      Modal.warning({
        title: '班级信息不存在',
        content: '请选择有效的班级',
      })
      return
    }

    // 检查时间冲突
    if (timeSlot.conflict) {
      Modal.warning({
        title: '时间冲突',
        content: '该时间段已有冲突，不可排课',
      })
      return
    }

    // 获取月份和日期信息
    const dateObj = dayjs(record.date)
    const month = dateObj.format('M')
    const day = dateObj.format('D')
    const lessonIndex = getLessonIndex(column.startTime)

    void confirmScheduleWithOptionalSkip({
      modeLabel: '班课',
      modeColor: '#13c2c2',
      targetLabel: '排课班级',
      targetValue: classInfo.name,
      courseName: classInfo.courseName,
      dateLabel: `${month}月${day}日 ${formatWeek(record.date)} 第${lessonIndex}节`,
      timeLabel: `${column.startTime}-${column.endTime}`,
      teacherName: record.name,
      groupLabel: activeGroupLabel.value || '当前组',
      onConfirm() {
        console.log('确认排课', classInfo.name, column.startTime, column.endTime)
        applyClassSchedule({
          classInfo,
          column,
          record,
        })
      },
    })
  }
}

function getLessonIndex(startTime) {
  const slots = activePeriodSlots.value
  const i = slots.findIndex(s => s.start === startTime)
  return i >= 0 ? i + 1 : ''
}

function emptyLessonStatusText(lesson) {
  const conflictTypes = uniqueConflictTypes(lesson?.conflictReason?.conflictTypes || [])
  if (!lesson?.conflict)
    return '空闲时段(可排)'
  if (lesson?.conflictReason?.type === '1v1-assistant-selection')
    return '主助教同人(不可排)'
  if (conflictTypes.length === 1 && conflictTypes[0] === '助教')
    return '助教冲突(不可排)'
  return '时间冲突(不可排)'
}

function scheduleLessonTitle(text) {
  const className = String(text?.className || '').trim()
  const courseName = String(text?.courseName || '').trim()
  return className || courseName || '课程'
}

function scheduleLessonMeta(text) {
  const className = String(text?.className || '').trim()
  const courseName = String(text?.courseName || '').trim()
  if (className && courseName && className !== courseName)
    return courseName
  return courseName || className || '-'
}

function formatDragDateLabel(date) {
  if (!date)
    return '-'
  return `${dayjs(date).format('MM-DD')} (${formatWeek(date)})`
}

function formatDragTimeLabel(startTime, endTime, separator = '~') {
  if (!startTime || !endTime)
    return '-'
  return `${startTime}${separator}${endTime}`
}

function updateDragPointer(event) {
  dragPointerState.value = {
    x: Number(event?.clientX || 0),
    y: Number(event?.clientY || 0),
    visible: true,
  }
}

function clearCustomScheduleDragListeners() {
  if (typeof document === 'undefined')
    return
  pendingScheduleDragStart = null
  if (customScheduleDragMoveHandler)
    document.removeEventListener('mousemove', customScheduleDragMoveHandler)
  if (customScheduleDragUpHandler)
    document.removeEventListener('mouseup', customScheduleDragUpHandler)
  customScheduleDragMoveHandler = null
  customScheduleDragUpHandler = null
  document.body.style.userSelect = ''
}

function clearBlockedScheduleDragAttempt() {
  blockedScheduleDragAttempt = null
  if (typeof document === 'undefined')
    return
  if (blockedScheduleDragMoveHandler)
    document.removeEventListener('mousemove', blockedScheduleDragMoveHandler)
  if (blockedScheduleDragUpHandler)
    document.removeEventListener('mouseup', blockedScheduleDragUpHandler)
  blockedScheduleDragMoveHandler = null
  blockedScheduleDragUpHandler = null
}

function showBlockedScheduleDragHint(message) {
  const now = Date.now()
  if (now - lastBlockedScheduleDragHintAt < 1200)
    return
  lastBlockedScheduleDragHintAt = now
  messageService.info(message)
}

function createEmptyDragHoverState() {
  return {
    key: '',
    teacherId: '',
    teacherName: '-',
    lessonDate: '',
    startTime: '',
    endTime: '',
    checking: false,
    valid: null,
    label: '',
    message: '',
    conflictTypes: [],
    existingSchedules: [],
  }
}

function setDragValidationState(target, payload = {}, options = {}) {
  const key = String(target?.key || payload?.key || '').trim()
  if (!key)
    return null

  if (options.sessionId != null && options.sessionId !== activeDragValidationSessionId)
    return null

  if (options.dragState && draggingScheduleState.value?.scheduleId !== options.dragState.scheduleId)
    return null

  const previous = dragValidationStateMap.value[key] || createEmptyDragHoverState()
  const nextConflictTypes = payload?.conflictTypes ?? previous.conflictTypes ?? []
  const nextExistingSchedules = Array.isArray(payload?.existingSchedules)
    ? payload.existingSchedules
    : (Array.isArray(previous.existingSchedules) ? previous.existingSchedules : [])

  const next = {
    ...previous,
    ...(target || {}),
    ...(payload || {}),
    key,
    conflictTypes: uniqueConflictTypes(nextConflictTypes),
    existingSchedules: nextExistingSchedules,
  }
  if (payload?.checking == null && (payload?.valid === true || payload?.valid === false))
    next.checking = false
  dragValidationStateMap.value[key] = next
  if (dragHoverState.value.key === key)
    dragHoverState.value = { ...next }
  return next
}

function resetDragScheduleState() {
  clearCustomScheduleDragListeners()
  clearBlockedScheduleDragAttempt()
  activeDragValidationSessionId += 1
  draggingScheduleState.value = null
  draggingScheduleCellKey.value = ''
  dragPointerState.value = {
    x: 0,
    y: 0,
    visible: false,
  }
  dragHoverState.value = createEmptyDragHoverState()
  dragValidationStateMap.value = {}
  dragValidationCache.clear()
  dragValidationPromises.clear()
}

function scheduleStudentText(text) {
  return Array.isArray(text?.studentNames)
    ? text.studentNames.map(item => item.name).filter(Boolean).join('、')
    : ''
}

function isScheduleDraggable(text) {
  return currentModel.value === '1'
    && text?.courseType === 1
    && text?.isMain !== false
    && Boolean(text?.scheduleId)
    && Boolean(text?.classId)
}

function resolveScheduleDragBlockedMessage(text) {
  if (text?.courseType === 1 && text?.isMain === false)
    return '当前是助教课表，暂不支持拖拽调课，请在主教老师所在行操作'
  if (text?.courseType === 2)
    return '当前仅支持 1v1 主教课程拖拽调课'
  if (text?.courseType === 1)
    return '当前课程暂不支持拖拽调课'
  return ''
}

function buildDraggingScheduleState(text, column, record) {
  return {
    scheduleId: String(text?.scheduleId || '').trim(),
    oneToOneId: String(text?.classId || '').trim(),
    assistantIds: Array.isArray(text?.assistantIds)
      ? text.assistantIds.map(id => String(id || '').trim()).filter(Boolean)
      : [],
    assistantText: text?.assistantText || '未安排',
    classroomId: String(text?.classroomId || '').trim(),
    courseName: text?.courseName || '',
    lessonTitle: scheduleLessonTitle(text),
    lessonMeta: scheduleLessonMeta(text),
    studentText: scheduleStudentText(text),
    sourceDate: record?.date || '',
    sourceStartTime: column?.startTime || '',
    sourceEndTime: column?.endTime || '',
    sourceCellKey: buildAvailabilitySlotKey(record?.teacherId, record?.date, column?.startTime, column?.endTime),
  }
}

function buildDragTarget(column, record) {
  return {
    key: buildAvailabilitySlotKey(record?.teacherId, record?.date, column?.startTime, column?.endTime),
    teacherId: String(record?.teacherId || '').trim(),
    teacherName: record?.name || '-',
    lessonDate: record?.date || '',
    startTime: column?.startTime || '',
    endTime: column?.endTime || '',
  }
}

function dragConflictStateFromTypes(conflictTypes, message) {
  const types = uniqueConflictTypes(conflictTypes)
  if (types.length === 1) {
    if (types[0] === '助教') {
      return {
        valid: false,
        label: '助教冲突(不可调)',
        message: message || '所选助教该时间段已有安排',
        conflictTypes: types,
      }
    }
    if (types[0] === '学员') {
      return {
        valid: false,
        label: '学员冲突(不可调)',
        message: message || '当前学员该时间段已有安排',
        conflictTypes: types,
      }
    }
    if (types[0] === '老师') {
      return {
        valid: false,
        label: '老师冲突(不可调)',
        message: message || '当前老师该时间段已有安排',
        conflictTypes: types,
      }
    }
    if (types[0] === '教室') {
      return {
        valid: false,
        label: '教室冲突(不可调)',
        message: message || '当前教室该时间段已有安排',
        conflictTypes: types,
      }
    }
  }
  return {
    valid: false,
    label: '冲突时段(不可调)',
    message: message || '当前空点不可调课',
    conflictTypes: types,
  }
}

function buildDragValidationResultFromValidationItem(target, item) {
  if (!item) {
    return {
      valid: false,
      label: '检测失败',
      message: '未返回当前空点的检测结果',
      conflictTypes: [],
      existingSchedules: [],
    }
  }

  if (item.valid) {
    return {
      valid: true,
      label: '可调课',
      message: item.message || `${target.teacherName} ${target.lessonDate} ${target.startTime}-${target.endTime} 可调课`,
      conflictTypes: [],
      existingSchedules: Array.isArray(item.existingSchedules) ? item.existingSchedules : [],
    }
  }

  return {
    ...dragConflictStateFromTypes(item.conflictTypes || [], item.message || '当前空点不可调课'),
    existingSchedules: Array.isArray(item.existingSchedules) ? item.existingSchedules : [],
  }
}

function validateDragTargetLocally(dragState, target) {
  if (!dragState?.scheduleId) {
    return {
      valid: false,
      label: '拖拽失效',
      message: '当前拖拽状态已失效，请重新拖拽',
      conflictTypes: [],
    }
  }
  if (!dragState.oneToOneId) {
    return {
      valid: false,
      label: '信息不完整',
      message: '当前课程缺少1对1标识，暂不支持拖拽调课',
      conflictTypes: [],
    }
  }
  if (dragState.assistantIds.includes(target.teacherId)) {
    return {
      valid: false,
      label: '主助教同人(不可调)',
      message: '目标老师已在本节课助教列表中，主教与助教不能为同一人',
      conflictTypes: ['助教'],
    }
  }
  return null
}

function buildDragValidationCacheKey(dragState, target) {
  return [
    dragState.scheduleId,
    dragState.oneToOneId,
    target.teacherId,
    target.lessonDate,
    target.startTime,
    target.endTime,
    dragState.assistantIds.join(','),
    dragState.classroomId,
  ].join('|')
}

function scheduleCellValue(column, record) {
  const root = column?.dataIndex?.[0]
  const index = Number(column?.dataIndex?.[1])
  if (!Number.isInteger(index))
    return null
  if (root === 'lessons')
    return record?.lessons?.[index] || null
  if (root === 'cells')
    return record?.cells?.[index] || null
  return null
}

function collectVisibleEmptyDragTargets() {
  const targets = []
  const seen = new Set()
  const scheduleColumns = columns.value.filter(isScheduleColumn)

  tableDataSource.value.forEach((row) => {
    scheduleColumns.forEach((column) => {
      const lesson = scheduleCellValue(column, row)
      if (!lesson || lesson.studentId)
        return

      const target = buildDragTarget(scheduleCellContextColumn(column, row), scheduleCellContextRecord(column, row))
      if (!target.key || seen.has(target.key))
        return

      seen.add(target.key)
      targets.push(target)
    })
  })

  return targets
}

function applyDragValidationResult(target, result, options = {}) {
  const dragState = options.dragState || draggingScheduleState.value
  if (dragState)
    dragValidationCache.set(buildDragValidationCacheKey(dragState, target), result)
  if (options.updateCellState !== false)
    setDragValidationState(target, result, { sessionId: options.sessionId, dragState })
  if (options.apply !== false && dragHoverState.value.key === target.key)
    applyDragHoverState(target.key, { ...target, ...result })
}

async function validateDragTargetsInBatch(targets, options = {}) {
  const dragState = options.dragState || draggingScheduleState.value
  if (!dragState || !targets.length)
    return

  try {
    const res = await validateOneToOneSchedulesApi({
      oneToOneId: dragState.oneToOneId,
      teacherId: targets[0]?.teacherId || '',
      assistantIds: dragState.assistantIds,
      classroomId: dragState.classroomId,
      excludeIds: [dragState.scheduleId],
      schedules: targets.map(target => ({
        lessonDate: target.lessonDate,
        startTime: target.startTime,
        endTime: target.endTime,
        teacherId: target.teacherId,
        assistantIds: dragState.assistantIds,
        classroomId: dragState.classroomId,
      })),
    })
    if (res.code !== 200 || !res.result)
      throw new Error(res.message || '批量检测调课空点失败')

    const itemMap = new Map(
      (Array.isArray(res.result.items) ? res.result.items : [])
        .map(item => [buildAvailabilitySlotKey(item.teacherId, item.lessonDate, item.startTime, item.endTime), item]),
    )

    targets.forEach((target) => {
      const item = itemMap.get(target.key)
      const result = buildDragValidationResultFromValidationItem(target, item)
      applyDragValidationResult(target, result, {
        dragState,
        sessionId: options.sessionId,
        apply: options.apply === true && dragHoverState.value.key === target.key,
      })
    })
  }
  catch (error) {
    console.error('validate drag targets in batch failed', error)
    const result = {
      valid: false,
      label: '检测失败',
      message: error?.response?.data?.message || error?.message || '批量检测调课空点失败',
      conflictTypes: [],
      existingSchedules: [],
    }
    targets.forEach(target => applyDragValidationResult(target, result, {
      dragState,
      sessionId: options.sessionId,
      apply: options.apply === true && dragHoverState.value.key === target.key,
    }))
  }
}

const dragPreviewStyle = computed(() => {
  const dragState = draggingScheduleState.value
  if (!dragState?.previewWidth || !dragState?.previewHeight) {
    const x = Math.max(12, Math.round(dragPointerState.value.x + 18))
    const y = Math.max(12, Math.round(dragPointerState.value.y - 22))
    return {
      transform: `translate3d(${x}px, ${y}px, 0)`,
    }
  }

  const left = Math.round(dragPointerState.value.x - dragState.offsetX)
  const top = Math.max(8, Math.round(dragPointerState.value.y - dragState.offsetY - 22))
  return {
    transform: `translate3d(${left}px, ${top}px, 0)`,
  }
})

const draggingScheduleStyle = computed(() => {
  const dragState = draggingScheduleState.value
  if (!dragState?.previewWidth || !dragState?.previewHeight)
    return {}
  const left = Math.round(dragPointerState.value.x - dragState.offsetX)
  const top = Math.round(dragPointerState.value.y - dragState.offsetY)
  return {
    position: 'fixed',
    left: `${left}px`,
    top: `${top}px`,
    width: `${Math.round(dragState.previewWidth)}px`,
    height: `${Math.round(dragState.previewHeight)}px`,
    zIndex: 1200,
  }
})

const dragPreviewTargetText = computed(() => {
  if (!draggingScheduleState.value)
    return ''
  if (!dragHoverState.value.key)
    return '拖动到空闲时段调课'
  const parts = []
  if (dragHoverState.value.valid === false)
    parts.push('不可调')
  parts.push('调整到：')
  parts.push(formatDragDateLabel(dragHoverState.value.lessonDate))
  parts.push(dragHoverState.value.startTime || '')
  return parts.filter(Boolean).join(' ')
})

function buildDragConfirmState(dragState, target) {
  return {
    source: {
      dateLabel: formatDragDateLabel(dragState.sourceDate),
      timeLabel: formatDragTimeLabel(dragState.sourceStartTime, dragState.sourceEndTime),
      lessonTitle: dragState.lessonTitle || '-',
      courseName: dragState.lessonMeta || '-',
      studentText: dragState.studentText || '-',
    },
    target: {
      dateLabel: formatDragDateLabel(target.lessonDate),
      timeLabel: formatDragTimeLabel(target.startTime, target.endTime),
      lessonTitle: dragState.lessonTitle || '-',
      courseName: dragState.lessonMeta || '-',
      studentText: dragState.studentText || '-',
    },
    payload: {
      dragState,
      target,
    },
  }
}

function applyDragHoverState(targetKey, payload) {
  if (!targetKey)
    return
  dragHoverState.value = {
    key: targetKey,
    checking: false,
    valid: null,
    label: '',
    message: '',
    conflictTypes: [],
    ...payload,
  }
}

async function ensureDragTargetValidation(target, options = {}) {
  const dragState = options.dragState || draggingScheduleState.value
  if (!dragState)
    return null

  const localResult = validateDragTargetLocally(dragState, target)
  if (localResult) {
    if (options.updateCellState !== false)
      setDragValidationState(target, localResult, { sessionId: options.sessionId, dragState })
    if (options.apply !== false && dragHoverState.value.key === target.key)
      applyDragHoverState(target.key, { ...target, ...localResult })
    return localResult
  }

  const cacheKey = buildDragValidationCacheKey(dragState, target)
  if (dragValidationCache.has(cacheKey)) {
    const cached = dragValidationCache.get(cacheKey)
    if (options.updateCellState !== false)
      setDragValidationState(target, cached, { sessionId: options.sessionId, dragState })
    if (options.apply !== false && dragHoverState.value.key === target.key)
      applyDragHoverState(target.key, { ...target, ...cached })
    return cached
  }

  if (dragValidationPromises.has(cacheKey)) {
    const pending = dragValidationPromises.get(cacheKey)
    const result = await pending
    if (options.updateCellState !== false)
      setDragValidationState(target, result, { sessionId: options.sessionId, dragState })
    if (options.apply !== false && dragHoverState.value.key === target.key)
      applyDragHoverState(target.key, { ...target, ...result })
    return result
  }

  const promise = (async () => {
    try {
      const res = await validateOneToOneSchedulesApi({
        oneToOneId: dragState.oneToOneId,
        teacherId: target.teacherId,
        assistantIds: dragState.assistantIds,
        classroomId: dragState.classroomId,
        excludeIds: [dragState.scheduleId],
        schedules: [{
          lessonDate: target.lessonDate,
          startTime: target.startTime,
          endTime: target.endTime,
          teacherId: target.teacherId,
          assistantIds: dragState.assistantIds,
          classroomId: dragState.classroomId,
        }],
      })
      if (res.code !== 200 || !res.result)
        throw new Error(res.message || '检测调课空点失败')

      const item = Array.isArray(res.result.items) && res.result.items.length === 1
        ? res.result.items[0]
        : null
      const result = item
        ? buildDragValidationResultFromValidationItem(target, item)
        : res.result.valid
          ? {
              valid: true,
              label: '可调课',
              message: `${target.teacherName} ${target.lessonDate} ${target.startTime}-${target.endTime} 可调课`,
              conflictTypes: [],
              existingSchedules: [],
            }
          : {
              ...dragConflictStateFromTypes(res.result.conflictTypes || [], res.result.message || '当前空点不可调课'),
              existingSchedules: Array.isArray(res.result.existingSchedules) ? res.result.existingSchedules : [],
            }
      dragValidationCache.set(cacheKey, result)
      return result
    }
    catch (error) {
      console.error('validate drag target failed', error)
      const result = {
        valid: false,
        label: '检测失败',
        message: error?.response?.data?.message || error?.message || '检测调课空点失败',
        conflictTypes: [],
      }
      dragValidationCache.set(cacheKey, result)
      return result
    }
    finally {
      if (dragValidationPromises.get(cacheKey) === promise)
        dragValidationPromises.delete(cacheKey)
    }
  })()

  dragValidationPromises.set(cacheKey, promise)
  const result = await promise
  if (options.updateCellState !== false)
    setDragValidationState(target, result, { sessionId: options.sessionId, dragState })
  if (options.apply !== false && dragHoverState.value.key === target.key && draggingScheduleState.value?.scheduleId === dragState.scheduleId) {
    applyDragHoverState(target.key, {
      ...target,
      ...result,
    })
  }
  return result
}

async function primeDragValidationForVisibleTargets(dragState, sessionId) {
  const remoteTargets = []

  collectVisibleEmptyDragTargets().forEach((target) => {
    const localResult = validateDragTargetLocally(dragState, target)
    if (localResult) {
      setDragValidationState(target, localResult, { sessionId, dragState })
      return
    }

    setDragValidationState(target, {
      checking: true,
      valid: null,
      label: '检测中...',
      message: '正在检测当前空点是否可调',
      conflictTypes: [],
      existingSchedules: [],
    }, { sessionId, dragState })
    remoteTargets.push(target)
  })

  if (!remoteTargets.length)
    return

  if (remoteTargets.length <= DRAG_BATCH_VALIDATE_SINGLE_REQUEST_THRESHOLD) {
    await validateDragTargetsInBatch(remoteTargets, {
      dragState,
      sessionId,
    })
    return
  }

  const chunks = []
  for (let i = 0; i < remoteTargets.length; i += DRAG_BATCH_VALIDATE_CHUNK_SIZE)
    chunks.push(remoteTargets.slice(i, i + DRAG_BATCH_VALIDATE_CHUNK_SIZE))

  let index = 0
  const workerCount = Math.min(DRAG_BATCH_VALIDATE_CONCURRENCY, chunks.length)
  const workers = Array.from({ length: workerCount }, async () => {
    while (index < chunks.length) {
      if (sessionId !== activeDragValidationSessionId || draggingScheduleState.value?.scheduleId !== dragState.scheduleId)
        return
      const chunk = chunks[index++]
      await validateDragTargetsInBatch(chunk, {
        dragState,
        sessionId,
      })
    }
  })

  await Promise.all(workers)
}

function emptyLessonDragState(column, record) {
  const target = buildDragTarget(column, record)
  return dragValidationStateMap.value[target.key]
    || (dragHoverState.value.key === target.key ? dragHoverState.value : null)
}

function handleSchedulePointerDown(event, text, column, record) {
  clearCustomScheduleDragListeners()
  clearBlockedScheduleDragAttempt()
  if (isScheduleDraggable(text)) {
    if (typeof document === 'undefined')
      return
    const dragElement = event.currentTarget instanceof HTMLElement ? event.currentTarget : null
    if (!dragElement)
      return

    const rect = dragElement.getBoundingClientRect()
    pendingScheduleDragStart = {
      startX: Number(event?.clientX || 0),
      startY: Number(event?.clientY || 0),
      dragState: buildDraggingScheduleState(text, column, record),
      rect,
      offsetX: Math.max(8, Math.min(rect.width - 8, Number(event?.clientX || 0) - rect.left)),
      offsetY: Math.max(8, Math.min(rect.height - 8, Number(event?.clientY || 0) - rect.top)),
    }

    customScheduleDragMoveHandler = (moveEvent) => {
      const moveX = Number(moveEvent?.clientX || 0)
      const moveY = Number(moveEvent?.clientY || 0)
      if (!draggingScheduleState.value) {
        if (!pendingScheduleDragStart)
          return
        const deltaX = Math.abs(moveX - pendingScheduleDragStart.startX)
        const deltaY = Math.abs(moveY - pendingScheduleDragStart.startY)
        if (deltaX < 4 && deltaY < 4)
          return

        draggingScheduleState.value = {
          ...pendingScheduleDragStart.dragState,
          previewWidth: pendingScheduleDragStart.rect.width,
          previewHeight: pendingScheduleDragStart.rect.height,
          offsetX: pendingScheduleDragStart.offsetX,
          offsetY: pendingScheduleDragStart.offsetY,
        }
        activeDragValidationSessionId += 1
        draggingScheduleCellKey.value = pendingScheduleDragStart.dragState.sourceCellKey
        dragHoverState.value = createEmptyDragHoverState()
        dragValidationStateMap.value = {}
        dragValidationCache.clear()
        dragValidationPromises.clear()
        const sessionId = activeDragValidationSessionId
        pendingScheduleDragStart = null
        document.body.style.userSelect = 'none'
        void primeDragValidationForVisibleTargets(draggingScheduleState.value, sessionId)
      }

      updateDragPointer(moveEvent)
      const target = resolvePointerDragTarget(moveX, moveY)
      if (!target) {
        dragHoverState.value = createEmptyDragHoverState()
        return
      }
      const existingState = dragValidationStateMap.value[target.key]
      if (existingState) {
        applyDragHoverState(target.key, {
          ...existingState,
          ...target,
        })
        return
      }
      setDragValidationState(target, {
        checking: true,
        valid: null,
        label: '检测中...',
        message: '正在检测当前空点是否可调',
        conflictTypes: [],
        existingSchedules: [],
      }, { sessionId: activeDragValidationSessionId, dragState: draggingScheduleState.value })
      applyDragHoverState(target.key, dragValidationStateMap.value[target.key] || {
        ...target,
        checking: true,
        valid: null,
        label: '检测中...',
        message: '正在检测当前空点是否可调',
        conflictTypes: [],
      })
      void ensureDragTargetValidation(target, {
        dragState: draggingScheduleState.value,
        sessionId: activeDragValidationSessionId,
      })
    }

    customScheduleDragUpHandler = async (upEvent) => {
      if (!draggingScheduleState.value) {
        clearCustomScheduleDragListeners()
        return
      }

      const dragState = draggingScheduleState.value
      const target = resolvePointerDragTarget(Number(upEvent?.clientX || 0), Number(upEvent?.clientY || 0))
      if (!target) {
        resetDragScheduleState()
        return
      }

      const validation = await ensureDragTargetValidation(target, { dragState })
      if (!validation?.valid) {
        const warningMessage = validation?.message || '当前空点不可调课'
        messageService.warning(warningMessage)
        if (Array.isArray(validation?.existingSchedules) && validation.existingSchedules.length) {
          openDragConflictDetailModal(validation, dragState, target)
        }
        resetDragScheduleState()
        return
      }

      openDragScheduleConfirm(dragState, target)
      resetDragScheduleState()
    }

    document.addEventListener('mousemove', customScheduleDragMoveHandler)
    document.addEventListener('mouseup', customScheduleDragUpHandler)
    return
  }

  const message = resolveScheduleDragBlockedMessage(text)
  if (!message || typeof document === 'undefined')
    return

  blockedScheduleDragAttempt = {
    startX: Number(event?.clientX || 0),
    startY: Number(event?.clientY || 0),
    message,
  }

  blockedScheduleDragMoveHandler = (moveEvent) => {
    if (!blockedScheduleDragAttempt)
      return
    const deltaX = Math.abs(Number(moveEvent?.clientX || 0) - blockedScheduleDragAttempt.startX)
    const deltaY = Math.abs(Number(moveEvent?.clientY || 0) - blockedScheduleDragAttempt.startY)
    if (deltaX < 6 && deltaY < 6)
      return
    showBlockedScheduleDragHint(blockedScheduleDragAttempt.message)
    clearBlockedScheduleDragAttempt()
  }

  blockedScheduleDragUpHandler = () => {
    clearBlockedScheduleDragAttempt()
  }

  document.addEventListener('mousemove', blockedScheduleDragMoveHandler)
  document.addEventListener('mouseup', blockedScheduleDragUpHandler)
}

function openDragScheduleConfirm(dragState, target) {
  dragConfirmDetail.value = buildDragConfirmState(dragState, target)
  dragConfirmOpen.value = true
}

async function submitDragScheduleAdjustment() {
  const payload = dragConfirmDetail.value?.payload
  if (!payload?.dragState || !payload?.target) {
    dragConfirmOpen.value = false
    return
  }

  const { dragState, target } = payload
  const dateLabel = dayjs(target.lessonDate).format('M月D日')
  const lessonIndex = getLessonIndex(target.startTime)
  dragConfirmSubmitting.value = true
  updatingDraggedSchedule.value = true
  try {
    const res = await batchUpdateTeachingSchedulesApi({
      ids: [dragState.scheduleId],
      teacherId: target.teacherId,
      assistantIds: dragState.assistantIds,
      classroomId: dragState.classroomId,
      lessonDate: target.lessonDate,
      startTime: target.startTime,
      endTime: target.endTime,
    })
    if (res.code !== 200)
      throw new Error(res.message || '拖拽调课失败')

    dragConfirmOpen.value = false
    dragConfirmDetail.value = {
      source: null,
      target: null,
      payload: null,
    }
    const lessonText = dragState.studentText || dragState.courseName || '课程'
    messageService.success(`已将 ${lessonText} 调到 ${dateLabel} ${formatWeek(target.lessonDate)} 第${lessonIndex}节`)
    emitter.emit(EVENTS.REFRESH_DATA)
  }
  catch (error) {
    console.error('drag update teaching schedule failed', error)
    messageService.error(error?.response?.data?.message || error?.message || '拖拽调课失败')
    await loadTimetableMatrix()
  }
  finally {
    dragConfirmSubmitting.value = false
    updatingDraggedSchedule.value = false
  }
}

function resolvePointerDragTarget(clientX, clientY) {
  if (typeof document === 'undefined')
    return null
  const el = document.elementFromPoint(clientX, clientY)
  const targetEl = el instanceof HTMLElement
    ? el.closest('[data-empty-schedule-cell-key]')
    : null
  if (!(targetEl instanceof HTMLElement))
    return null

  const key = String(targetEl.dataset.emptyScheduleCellKey || '').trim()
  if (!key)
    return null
  return {
    key,
    teacherId: String(targetEl.dataset.dragTargetTeacherId || '').trim(),
    teacherName: String(targetEl.dataset.dragTargetTeacherName || '').trim() || '-',
    lessonDate: String(targetEl.dataset.dragTargetLessonDate || '').trim(),
    startTime: String(targetEl.dataset.dragTargetStartTime || '').trim(),
    endTime: String(targetEl.dataset.dragTargetEndTime || '').trim(),
  }
}

// 添加监听，当模式切换时清空之前的选择
watch(currentModel, (newValue) => {
  console.log('切换模式', newValue)

  resetDragScheduleState()
  dragConfirmOpen.value = false
  dragConflictDetailOpen.value = false
  dragConfirmDetail.value = {
    source: null,
    target: null,
    payload: null,
  }
  dragConflictDetailState.value = {
    summary: '',
    attempted: null,
    items: [],
  }
  cancelOneToOneAvailabilityCheck()
  resetEmptyLessonConflicts()

  if (newValue === '1') {
    // 切换到1v1模式，清空班级选择
    classId.value = null
    className.value = null
  }
  else {
    resetOneToOnePickerState()
    courseId.value = null
    courseName.value = null
  }
})

watch(dragConfirmOpen, (open) => {
  if (!open && !dragConfirmSubmitting.value) {
    dragConfirmDetail.value = {
      source: null,
      target: null,
      payload: null,
    }
  }
})

watch(dragConflictDetailOpen, (open) => {
  if (!open) {
    dragConflictDetailState.value = {
      summary: '',
      attempted: null,
      items: [],
    }
  }
})
</script>

<template>
  <div ref="timetableRootRef">
    <div class="filter-wrap bg-white pl-3 pr-3 rounded-4 rounded-lt-0 rounded-rt-0">
      <all-filter
        :display-array="displayArray"
        :is-show-search-stu-phonefilter="true"
        :schedule-teacher-options="scheduleTeacherOptions"
        :schedule-teacher-finished="scheduleTeacherFinished"
        :on-schedule-teacher-dropdown-visible-change="onScheduleTeacherDropdownVisibleChange"
        :on-schedule-teacher-search="onScheduleTeacherSearch"
        :on-schedule-teacher-load-more="loadMoreScheduleTeacher"
        :schedule-classroom-options="scheduleClassroomOptions"
        :schedule-classroom-finished="scheduleClassroomFinished"
        :on-schedule-classroom-dropdown-visible-change="onScheduleClassroomDropdownVisibleChange"
        :on-schedule-classroom-search="onScheduleClassroomSearch"
        :schedule-class-options="scheduleClassOptions"
        :schedule-class-finished="scheduleClassFinished"
        :on-schedule-class-dropdown-visible-change="onScheduleClassDropdownVisibleChange"
        :on-schedule-class-search="onScheduleClassSearch"
        :on-schedule-class-load-more="loadMoreScheduleClass"
        :schedule-one-to-one-options="scheduleOneToOneOptions"
        :schedule-one-to-one-finished="scheduleOneToOneFinished"
        :on-schedule-one-to-one-dropdown-visible-change="onScheduleOneToOneDropdownVisibleChange"
        :on-schedule-one-to-one-search="onScheduleOneToOneSearch"
        :on-schedule-one-to-one-load-more="loadMoreScheduleOneToOne"
        :schedule-course-options="scheduleCourseOptions"
        :schedule-course-finished="scheduleCourseFinished"
        :on-schedule-course-dropdown-visible-change="onScheduleCourseDropdownVisibleChange"
        :on-schedule-course-search="onScheduleCourseSearch"
        :on-schedule-course-load-more="loadMoreScheduleCourse"
        :schedule-type-options="scheduleTypeOptions"
        :schedule-call-status-options="scheduleCallStatusOptions"
        @update:schedule-teacher-filter="handleScheduleTeacherFilter"
        @update:schedule-classroom-filter="handleScheduleClassroomFilter"
        @update:schedule-class-filter="handleScheduleClassFilter"
        @update:schedule-one-to-one-filter="handleScheduleOneToOneFilter"
        @update:schedule-course-filter="handleScheduleCourseFilter"
        @update:schedule-type-filter="handleScheduleTypeFilter"
        @update:schedule-call-status-filter="handleScheduleCallStatusFilter"
        @update:stu-phone-search-filter="handleStuPhoneFilter"
      />
    </div>
    <SmartTimetableToolbar
      v-model:current-model="currentModel"
      v-model:one-to-one-record-id="oneToOneRecordId"
      v-model:one-to-one-picker-open="oneToOnePickerOpen"
      v-model:class-id="classId"
      v-model:current-time="currentTime"
      v-model:current-week="currentWeek"
      v-model:current-group="currentGroup"
      :one-to-one-list-loading="oneToOneListLoading"
      :one-to-one-data="oneToOneData"
      :render-one-to-one-dropdown="renderOneToOneDropdown"
      :filter-one-to-one-option="filterOneToOneOption"
      :class-data="classData"
      :time-view-options="timeViewOptions"
      :format-date-range="formatDateRange"
      :is-week-like-view="isWeekLikeView"
      :group-options="groupOptions"
      :on-one-to-one-change="handle1v1"
      :on-one-to-one-dropdown-visible-change="handleOneToOneDropdownVisibleChange"
      :on-class-change="handleClass"
      :on-prev="handlePrev"
      :on-next="handleNext"
      :on-this-week="handleThisWeek"
    />
    <SmartTimetableGrid
      :spinning="timetableLoading || oneToOneAvailabilityLoading || creatingOneToOneSchedule || updatingDraggedSchedule"
      :table-data-source="tableDataSource"
      :columns="columns"
      :is-swap-time-grid="isSwapTimeGrid"
      :focused-schedule-cell-key="focusedScheduleCellKey"
      :dragging-schedule-cell-key="draggingScheduleCellKey"
      :is-schedule-column="isScheduleColumn"
      :schedule-cell-key="scheduleCellKey"
      :schedule-cell-start-time="scheduleCellStartTime"
      :schedule-cell-end-time="scheduleCellEndTime"
      :schedule-cell-context-column="scheduleCellContextColumn"
      :schedule-cell-context-record="scheduleCellContextRecord"
      :open-scheduled-lesson-detail="openScheduledLessonDetail"
      :open-scheduled-conflict-detail="openScheduledConflictDetail"
      :handle-conflict-click="handleConflictClick"
      :handle-schedule-click="handleScheduleClick"
      :handle-schedule-pointer-down="handleSchedulePointerDown"
      :is-schedule-draggable="isScheduleDraggable"
      :dragging-schedule-style="draggingScheduleStyle"
      :empty-lesson-drag-state="emptyLessonDragState"
      :empty-lesson-status-text="emptyLessonStatusText"
      :teacher-lesson-count-label="teacherLessonCountLabel"
      :format-week="formatWeek"
      :format-date="formatDate"
    />

    <div
      v-if="draggingScheduleState && dragPointerState.visible"
      class="st-drag-preview"
      :style="dragPreviewStyle"
    >
      <div
        class="st-drag-preview__target"
        :class="{
          'st-drag-preview__target--invalid': dragHoverState.valid === false,
          'st-drag-preview__target--active': dragHoverState.valid === true,
        }"
      >
        {{ dragPreviewTargetText }}
      </div>
    </div>

    <SmartTimetableScheduledDetailModal
      v-model:open="scheduledLessonDetailOpen"
      :detail="scheduledLessonDetailState"
      :deleting="deletingScheduledLesson"
      @confirm="deleteScheduledLessonFromDetail"
    />

    <ScheduleConflictModal
      v-model:open="scheduledConflictDetailOpen"
      :validation="scheduledConflictDetailValidation"
      title="冲突详情"
      current-title="当前冲突日程"
      existing-title="与其冲突的日程"
      fallback-message="当前日程与已有日程存在冲突"
    />

    <SmartTimetableConflictModal
      v-model:open="conflictDetailModalOpen"
      :forcing="forcingConflictSchedule"
      :conflict-detail-state="conflictDetailState"
      @force="forceScheduleDespiteStudentConflict"
      @jump="jumpToConflictSchedule"
    />

    <SmartTimetableDragConflictModal
      v-model:open="dragConflictDetailOpen"
      :detail="dragConflictDetailState"
      @jump="jumpToConflictSchedule"
    />

    <SmartTimetableDragConfirmModal
      v-model:open="dragConfirmOpen"
      :detail="dragConfirmDetail"
      :submitting="dragConfirmSubmitting"
      @confirm="submitDragScheduleAdjustment"
    />
  </div>
</template>

<style lang="less" scoped>
.st-drag-preview {
  position: fixed;
  top: 0;
  left: 0;
  z-index: 1100;
  pointer-events: none;
}

.st-drag-preview__target {
  display: inline-block;
  align-items: center;
  max-width: 320px;
  color: #1668ff;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.2;
  white-space: normal;
  text-shadow:
    0 1px 2px rgba(255, 255, 255, 0.96),
    0 0 10px rgba(255, 255, 255, 0.72);
}

.st-drag-preview__target--active {
  color: #1677ff;
}

.st-drag-preview__target--invalid {
  color: #cf1322;
}

/* 与班课下拉同量级宽度，避免顶栏把日期区挤换行 */
.st-top-1v1-select {
  width: 180px;
  max-width: 180px;
}

.st-top-class-select {
  width: 180px;
  max-width: 180px;
}

.st-top-1v1-dropdown {
  display: flex;
  width: 520px;
  min-width: 520px;
  max-width: 520px;
  min-height: 280px;
  max-height: 280px;
  background: #fff;
}

.st-top-1v1-dropdown__list {
  flex: 0 0 278px;
  min-width: 278px;
  max-width: 278px;
  overflow-y: auto;
  border-right: 1px solid #f0f0f0;
}

.st-top-1v1-dropdown__side {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-width: 0;
  padding: 10px 12px 12px;
  background: #fff;
}

.st-top-1v1-dropdown__section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 10px;
}

.st-top-1v1-dropdown__section-title {
  color: #262626;
  font-size: 12px;
  font-weight: 700;
}

.st-top-1v1-dropdown__section-hint {
  color: #8c8c8c;
  font-size: 12px;
}

.st-top-1v1-dropdown__search {
  margin-bottom: 10px;
}

.st-top-1v1-dropdown__summary {
  margin-bottom: 8px;
  color: #5b6475;
  font-size: 12px;
  line-height: 18px;
}

.st-top-1v1-dropdown__assistant-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1;
  overflow-y: auto;
  padding-right: 2px;
}

.st-top-1v1-dropdown__assistant-list {
  scrollbar-width: thin;
  scrollbar-color: #cfd6e4 transparent;
}

.st-top-1v1-dropdown__assistant-list::-webkit-scrollbar {
  width: 8px;
}

.st-top-1v1-dropdown__assistant-list::-webkit-scrollbar-track {
  background: transparent;
}

.st-top-1v1-dropdown__assistant-list::-webkit-scrollbar-thumb {
  background: #cfd6e4;
  border-radius: 999px;
  border: 2px solid transparent;
  background-clip: padding-box;
}

.st-top-1v1-dropdown__assistant-list::-webkit-scrollbar-thumb:hover {
  background: #b8c2d6;
  border-radius: 999px;
  border: 2px solid transparent;
  background-clip: padding-box;
}

.st-top-1v1-dropdown__assistant-item {
  display: flex;
  align-items: center;
  gap: 8px;
  min-height: 32px;
  padding: 6px 8px;
  border-radius: 8px;
  cursor: pointer;
  transition: background-color 0.18s ease;
}

.st-top-1v1-dropdown__assistant-item:hover {
  background: #f7faff;
}

.st-top-1v1-dropdown__assistant-name {
  flex: 1;
  min-width: 0;
  color: #262626;
  font-size: 13px;
  font-weight: 500;
}

.st-top-1v1-dropdown__assistant-mobile {
  color: #8c8c8c;
  font-size: 12px;
}

.st-top-1v1-dropdown__empty {
  padding: 12px 0 4px;
  color: #8c8c8c;
  font-size: 12px;
  line-height: 18px;
}

:deep(.st-top-1v1-select-dropdown .st-top-1v1-dropdown) {
  display: flex;
  width: 520px;
  min-width: 520px;
  max-width: 520px;
  min-height: 280px;
  max-height: 280px;
  background: #fff;
  border-radius: 12px;
  overflow: hidden;
}

:deep(.st-top-1v1-select-dropdown .st-top-1v1-dropdown__list) {
  flex: 0 0 278px;
  min-width: 278px;
  max-width: 278px;
  overflow-y: auto;
  border-right: 1px solid #f0f0f0;
}

:deep(.st-top-1v1-select-dropdown .st-top-1v1-dropdown__side) {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-width: 0;
  padding: 14px 16px 16px;
  background: linear-gradient(180deg, #fcfdff 0%, #fff 100%);
}

:deep(.st-top-1v1-select-dropdown .st-top-1v1-dropdown__section-head) {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

:deep(.st-top-1v1-select-dropdown .st-top-1v1-dropdown__section-title) {
  color: #262626;
  font-size: 14px;
  font-weight: 700;
  line-height: 1;
}

:deep(.st-top-1v1-select-dropdown .st-top-1v1-dropdown__section-hint) {
  color: #8c8c8c;
  font-size: 12px;
  line-height: 1;
}

:deep(.st-top-1v1-select-dropdown .st-top-1v1-dropdown__search-input) {
  width: 100%;
  height: 38px;
  padding: 0 12px;
  color: #262626;
  font-size: 14px;
  background: #fff;
  border: 1px solid #d9d9d9;
  border-radius: 10px;
  outline: none;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
  box-sizing: border-box;
}

:deep(.st-top-1v1-select-dropdown .st-top-1v1-dropdown__search-input:focus) {
  border-color: #1677ff;
  box-shadow: 0 0 0 2px rgba(22, 119, 255, 0.12);
}

:deep(.st-top-1v1-select-dropdown .st-top-1v1-dropdown__summary) {
  margin: 10px 0 8px;
  color: #5b6475;
  font-size: 12px;
  line-height: 18px;
}

:deep(.st-top-1v1-select-dropdown .st-top-1v1-dropdown__assistant-list) {
  display: flex;
  flex-direction: column;
  gap: 6px;
  flex: 1;
  overflow-y: auto;
  padding-right: 2px;
}

:deep(.st-top-1v1-select-dropdown .st-top-1v1-dropdown__assistant-item) {
  display: flex;
  align-items: center;
  gap: 10px;
  min-height: 42px;
  padding: 8px 10px;
  border-radius: 10px;
  cursor: pointer;
  transition: background-color 0.18s ease;
  box-sizing: border-box;
}

:deep(.st-top-1v1-select-dropdown .st-top-1v1-dropdown__assistant-item:hover) {
  background: #f5f9ff;
}

:deep(.st-top-1v1-select-dropdown .st-top-1v1-dropdown__assistant-checkbox) {
  width: 18px;
  height: 18px;
  margin: 0;
  accent-color: #1677ff;
  flex: 0 0 auto;
}

:deep(.st-top-1v1-select-dropdown .st-top-1v1-dropdown__assistant-name) {
  flex: 1;
  min-width: 0;
  color: #262626;
  font-size: 14px;
  font-weight: 600;
  line-height: 20px;
}

:deep(.st-top-1v1-select-dropdown .st-top-1v1-dropdown__assistant-mobile) {
  color: #8c8c8c;
  font-size: 13px;
  line-height: 20px;
  flex: 0 0 auto;
}

:deep(.st-top-1v1-select-dropdown .st-top-1v1-dropdown__empty) {
  padding: 14px 0 4px;
  color: #8c8c8c;
  font-size: 12px;
  line-height: 18px;
}
</style>

<style lang="less">
.st-top-1v1-dropdown__assistant-list {
  scrollbar-width: thin;
  scrollbar-color: #c9d3e6 transparent;
}

.st-top-1v1-dropdown__assistant-list::-webkit-scrollbar {
  width: 10px;
}

.st-top-1v1-dropdown__assistant-list::-webkit-scrollbar-track {
  background: transparent;
}

.st-top-1v1-dropdown__assistant-list::-webkit-scrollbar-thumb {
  background: #c9d3e6;
  border-radius: 999px;
  border: 2px solid #fff;
}

.st-top-1v1-dropdown__assistant-list::-webkit-scrollbar-thumb:hover {
  background: #aebbd4;
}
</style>
