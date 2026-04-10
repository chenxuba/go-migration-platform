<script setup lang="ts">
import type { Dayjs } from 'dayjs'
import { CopyOutlined, DownloadOutlined, LeftOutlined, RightOutlined, SettingOutlined } from '@ant-design/icons-vue'
import { Modal, message } from 'ant-design-vue'
import dayjs from 'dayjs'
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import ScheduleBatchPlanEditModal from './schedule-batch-plan-edit-modal.vue'
import ScheduleConflictModal from './schedule-conflict-modal.vue'
import SmartTimetableScheduleDetailDrawer from './smart-timetable-schedule-detail-drawer.vue'
import TimetableScheduleHoverPopover from './timetable-schedule-hover-popover.vue'
import TimetableScheduleSummary from './timetable-schedule-summary.vue'
import { listClassroomsApi } from '@/api/business-settings/classroom'
import { getOneToOneListApi } from '@/api/edu-center/one-to-one'
import { pageGroupClassesApi } from '@/api/edu-center/group-class'
import { getCourseIdAndNameApi } from '@/api/edu-center/registr-renewal'
import type {
  MatrixTeacherFilterParam,
  TeachingScheduleItem,
  TeachingScheduleMatrixDay,
  TeachingScheduleMatrixLegacyItem,
} from '@/api/edu-center/teaching-schedule'
import {
  cancelTeachingSchedulesApi,
  copyTeachingSchedulesWeekApi,
  downloadTeachingSchedulesTeacherMatrixExcelApi,
  getTeachingScheduleConflictDetailApi,
  listTeachingSchedulesByTeacherMatrixApi,
} from '@/api/edu-center/teaching-schedule'
import { getUserListApi } from '@/api/internal-manage/staff-manage'
import messageService from '@/utils/messageService'

const emit = defineEmits<{
  (e: 'week-range-change', value: { startDate: string, endDate: string }): void
}>()

interface FilterOption {
  id: string
  value: string
}

interface DrawerSummary {
  scheduleId?: string
  id?: string
  lessonTitle?: string
  courseName?: string
  teacherName?: string
  assistantText?: string
  classroomName?: string
  studentText?: string
  courseType?: number
}

const displayArray = ref([
  'scheduleTeacher',
  'scheduleClassroom',
  'scheduleClass',
  'scheduleOneToOne',
  'scheduleCourse',
  'scheduleType',
  'scheduleCallStatus',
])

const currentDate = ref(dayjs())
const now = ref(dayjs())
const loading = ref(false)
const matrixDays = ref<TeachingScheduleMatrixDay[]>([])
const scheduleDetailOpen = ref(false)
const currentDetailSchedule = ref<TeachingScheduleItem | null>(null)
const currentScheduleDetail = ref<DrawerSummary | null>(null)
const deletingScheduleDetail = ref(false)
const scheduleBatchPlanEditOpen = ref(false)
const currentBatchPlanSchedule = ref<TeachingScheduleItem | null>(null)
const scheduleConflictOpen = ref(false)
const scheduleConflictValidation = ref(null)
const scheduleConflictLoading = ref(false)

const headerDatesRef = ref<HTMLElement | null>(null)
const bodyScrollRef = ref<HTMLElement | null>(null)
let syncingScroll = false
let matrixLoadSeq = 0

const filterStudentId = ref<string | undefined>(undefined)
const filterTeacherId = ref<string[]>([])
const filterClassroomId = ref<string[]>([])
const filterClassId = ref<string | undefined>(undefined)
const filterOneToOneId = ref<string | undefined>(undefined)
const filterCourseId = ref<string | undefined>(undefined)
const filterScheduleType = ref<string[]>([])
const filterCallStatus = ref<string | undefined>(undefined)

const scheduleTeacherOptions = ref<FilterOption[]>([])
const scheduleClassroomOptions = ref<FilterOption[]>([])
const scheduleClassOptions = ref<FilterOption[]>([])
const scheduleOneToOneOptions = ref<FilterOption[]>([])
const scheduleCourseOptions = ref<FilterOption[]>([])

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

const scheduleTypeOptions: FilterOption[] = [
  { id: 'group_class', value: '班级日程' },
  { id: 'one_to_one', value: '1对1日程' },
  { id: 'trial', value: '试听日程' },
]

const scheduleCallStatusOptions: FilterOption[] = [
  { id: 'unsigned', value: '未点名' },
  { id: 'signed', value: '已点名' },
]

function scheduleBadgeText(classType: number) {
  return Number(classType) === 1 ? '班课' : '1v1'
}

function conflictBadgeTitle(event: CellSchedule) {
  const types = Array.isArray(event?.raw?.conflictTypes)
    ? event.raw.conflictTypes.map(item => String(item || '').trim()).filter(Boolean)
    : []
  return types.length ? `冲突原因：${types.join('、')}冲突，点击查看详情` : '当前课程存在冲突，点击查看详情'
}

function isOneToOneSchedule(schedule: TeachingScheduleItem | null | undefined) {
  return Number(schedule?.classType) === 2
}

function scheduleAssistantText(schedule: TeachingScheduleItem | null | undefined) {
  const list = Array.isArray(schedule?.assistantNames)
    ? schedule.assistantNames.map(item => String(item || '').trim()).filter(Boolean)
    : []
  return list.length ? list.join('、') : '未安排'
}

function scheduleStudentSummary(schedule: TeachingScheduleItem | null | undefined) {
  const text = String(schedule?.studentName || '').trim()
  return text || '-'
}

function scheduleConflictSummary(schedule: TeachingScheduleItem | null | undefined) {
  const types = Array.isArray(schedule?.conflictTypes)
    ? schedule.conflictTypes.map(item => String(item || '').trim()).filter(Boolean)
    : []
  if (types.length)
    return `${types.join('、')}冲突`
  return schedule?.conflict ? '当前课程存在冲突' : ''
}

function scheduleHoverTitle(schedule: TeachingScheduleItem | null | undefined) {
  if (isOneToOneSchedule(schedule))
    return String(schedule?.lessonName || '').trim() || '1对1日程'
  return String(schedule?.teachingClassName || '').trim()
    || String(schedule?.lessonName || '').trim()
    || '班课日程'
}

function scheduleTimeTextFromEvent(event: CellSchedule) {
  return `${event.startAt.format('YYYY-MM-DD')} ${event.startAt.format('HH:mm')} ~ ${event.endAt.format('HH:mm')}`
}

function buildScheduleDrawerDetail(schedule: TeachingScheduleItem | null | undefined): DrawerSummary | null {
  if (!schedule)
    return null
  return {
    scheduleId: String(schedule.id || '').trim(),
    id: String(schedule.id || '').trim(),
    lessonTitle: scheduleHoverTitle(schedule),
    courseName: String(schedule.lessonName || '').trim() || '-',
    teacherName: String(schedule.teacherName || '').trim() || '-',
    assistantText: scheduleAssistantText(schedule),
    classroomName: String(schedule.classroomName || '').trim() || '',
    studentText: scheduleStudentSummary(schedule),
    courseType: isOneToOneSchedule(schedule) ? 1 : 2,
  }
}

function normalizeScheduleFilterValue(value: unknown): string | undefined {
  if (Array.isArray(value))
    return value.length ? String(value[0] ?? '').trim() || undefined : undefined
  const text = String(value ?? '').trim()
  return text || undefined
}

function normalizeScheduleFilterValues(value: unknown): string[] {
  if (!Array.isArray(value))
    return []
  return value.map(item => String(item ?? '').trim()).filter(Boolean)
}

function handleScheduleTeacherFilter(value: unknown) {
  filterTeacherId.value = normalizeScheduleFilterValues(value)
}

function handleScheduleClassroomFilter(value: unknown) {
  filterClassroomId.value = normalizeScheduleFilterValues(value)
}

function handleScheduleClassFilter(value: unknown) {
  filterClassId.value = normalizeScheduleFilterValue(value)
}

function handleScheduleOneToOneFilter(value: unknown) {
  filterOneToOneId.value = normalizeScheduleFilterValue(value)
}

function handleScheduleCourseFilter(value: unknown) {
  filterCourseId.value = normalizeScheduleFilterValue(value)
}

function handleScheduleTypeFilter(value: unknown) {
  filterScheduleType.value = normalizeScheduleFilterValues(value)
}

function handleScheduleCallStatusFilter(value: unknown) {
  filterCallStatus.value = normalizeScheduleFilterValue(value)
}

function handleStuPhoneFilter(value: unknown) {
  filterStudentId.value = normalizeScheduleFilterValue(value)
}

function mergeFilterOptions(previous: FilterOption[], incoming: FilterOption[], selectedValues: string[] | string | undefined = []) {
  const selectedSet = new Set((Array.isArray(selectedValues) ? selectedValues : [selectedValues]).map(value => String(value || '')).filter(Boolean))
  const map = new Map<string, FilterOption>()
  previous.forEach((item) => {
    const key = String(item?.id || '').trim()
    if (key && selectedSet.has(key))
      map.set(key, item)
  })
  incoming.forEach((item) => {
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
      },
      sortModel: {},
    })
    if (res.code !== 200)
      return
    const resultData: FilterOption[] = (Array.isArray(res.result) ? res.result : []).map(item => ({
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
  try {
    const res = await listClassroomsApi({
      enabledOnly: true,
      searchKey,
    })
    if (res.code !== 200)
      return
    const resultData: FilterOption[] = (Array.isArray(res.result) ? res.result : []).map(item => ({
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
    const resultData: FilterOption[] = list.map(item => ({
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
    const resultData: FilterOption[] = list.map(item => ({
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
  scheduleCourseSearchKey.value = searchKey
  try {
    const res = await getCourseIdAndNameApi({ searchKey })
    if (res.code !== 200)
      return
    const resultData: FilterOption[] = (Array.isArray(res.result) ? res.result : []).map(item => ({
      id: String(item.id ?? ''),
      value: String(item.name || item.id || '').trim(),
    })).filter(item => item.id && item.value)
    scheduleCourseOptions.value = mergeFilterOptions(
      reset ? [] : scheduleCourseOptions.value,
      resultData,
      filterCourseId.value,
    )
    scheduleCoursePagination.value.current = 1
    scheduleCoursePagination.value.total = resultData.length
    scheduleCourseFinished.value = true
  }
  catch (error) {
    console.error('load schedule course options failed', error)
  }
}

async function onScheduleTeacherDropdownVisibleChange() {
  await loadScheduleTeacherOptions('', true)
}

async function onScheduleTeacherSearch(keyword: string) {
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

async function onScheduleClassroomSearch(keyword: string) {
  await loadScheduleClassroomOptions(keyword || '')
}

async function onScheduleClassDropdownVisibleChange() {
  await loadScheduleClassOptions('', true)
}

async function onScheduleClassSearch(keyword: string) {
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

async function onScheduleOneToOneSearch(keyword: string) {
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

async function onScheduleCourseSearch(keyword: string) {
  await loadScheduleCourseOptions(keyword || '', true)
}

async function loadMoreScheduleCourse() {
  scheduleCourseFinished.value = true
}

/** 日期条横向「钉」在可视区内（对齐旧版 orgAllCourseData.vue updateFloatingDatePositions） */
/** 浮动日期芯片宽度（与样式 padding、边框配合） */
const floatingDatePillWidth = 156

interface FloatingDateStyle {
  left: string
  opacity: string
  visibility: string
}
const floatingDateStyles = ref<Record<string, FloatingDateStyle>>({})

let headerDatesResizeObserver: ResizeObserver | null = null

/* eslint-disable ts/no-use-before-define */
function getDayColumnRange(dayIndex: number) {
  const groups = dateTeacherGroups.value
  if (dayIndex < 0 || dayIndex >= groups.length)
    return { left: 0, width: 0 }
  let left = 0
  for (let i = 0; i < dayIndex; i++) {
    const g = groups[i]
    for (const t of g.teachers) {
      const entry = layoutsByCell.value.get(`${g.dateKey}|${t.key}`)
      const w = Math.max(teacherColWidth, entry?.colWidth ?? teacherColWidth)
      left += w
    }
  }
  const g = groups[dayIndex]
  let width = 0
  for (const t of g.teachers) {
    const entry = layoutsByCell.value.get(`${g.dateKey}|${t.key}`)
    const w = Math.max(teacherColWidth, entry?.colWidth ?? teacherColWidth)
    width += w
  }
  return { left, width }
}

/**
 * 对齐旧版 orgAllCourseData.vue：以课表主体 scrollLeft 参与计算；可视宽度与旧版 scheduleBody.clientWidth-100
 * 一致时可用 header 日期区宽度；left 为内容坐标 finalPosition（旧版 floatingElement.style.left = finalPosition + 'px'）。
 */
function updateFloatingDatePositions(scrollLeftOverride?: number) {
  const headerEl = headerDatesRef.value
  const boardEl = bodyScrollRef.value
  const groups = dateTeacherGroups.value
  if (!headerEl || !groups.length) {
    floatingDateStyles.value = {}
    return
  }
  const scrollLeft
    = scrollLeftOverride ?? boardEl?.scrollLeft ?? headerEl.scrollLeft
  const containerWidth
    = headerEl.clientWidth > 0
      ? headerEl.clientWidth
      : Math.max(300, (boardEl?.clientWidth ?? 800) - 100)
  const pillW = floatingDatePillWidth
  const next: Record<string, FloatingDateStyle> = {}

  groups.forEach((g, index) => {
    const { left: leftOffset, width: dayColumnWidth } = getDayColumnRange(index)
    const rightOffset = leftOffset + dayColumnWidth
    const expandedScrollLeft = Math.max(0, scrollLeft - 200)
    const expandedScrollRight = scrollLeft + containerWidth + 200
    const inPlay = rightOffset > expandedScrollLeft && leftOffset < expandedScrollRight

    if (!inPlay || dayColumnWidth <= 0) {
      next[g.dateKey] = {
        left: `${leftOffset}px`,
        opacity: '0',
        visibility: 'hidden',
      }
      return
    }

    let finalPosition = leftOffset + (dayColumnWidth - pillW) / 2
    if (leftOffset >= scrollLeft && rightOffset <= scrollLeft + containerWidth) {
      finalPosition = leftOffset + (dayColumnWidth - pillW) / 2
    }
    else {
      const visibleLeft = Math.max(leftOffset, scrollLeft)
      const visibleRight = Math.min(rightOffset, scrollLeft + containerWidth)
      const visibleWidth = visibleRight - visibleLeft
      if (visibleWidth >= pillW)
        finalPosition = visibleLeft + (visibleWidth - pillW) / 2
      else
        finalPosition = visibleLeft
    }
    if (finalPosition < leftOffset)
      finalPosition = leftOffset
    if (finalPosition + pillW > rightOffset)
      finalPosition = rightOffset - pillW

    next[g.dateKey] = {
      left: `${finalPosition}px`,
      opacity: '1',
      visibility: 'visible',
    }
  })
  floatingDateStyles.value = next
}
/* eslint-enable ts/no-use-before-define */

function getFloatingStyle(dateKey: string): FloatingDateStyle {
  return floatingDateStyles.value[dateKey] ?? {
    left: '0',
    opacity: '0',
    visibility: 'hidden',
  }
}

function floatingPillStyle(dateKey: string): Record<string, string> {
  const s = getFloatingStyle(dateKey)
  return {
    width: `${floatingDatePillWidth}px`,
    left: s.left,
    opacity: s.opacity,
    visibility: s.visibility,
  }
}

const timelineStart = 8 * 60
const timelineEnd = 22 * 60
const timelineTopPadding = 18
const timelineBottomPadding = 28

const TM_DISPLAY_STORAGE_KEY = 'teacher-matrix-display-v1'

interface MatrixDisplayPreferences {
  weekdays: number[]
  teacherFilter: MatrixTeacherFilterParam
}

/** 每小时行高（px），固定为原「宽松 50」档，不再提供配置项 */
const HOUR_ROW_HEIGHT_PX = 120

function defaultMatrixDisplay(): MatrixDisplayPreferences {
  return {
    weekdays: [1, 2, 3, 4, 5, 6, 7],
    teacherFilter: 'all',
  }
}

function loadMatrixDisplayFromStorage(): MatrixDisplayPreferences {
  try {
    const raw = localStorage.getItem(TM_DISPLAY_STORAGE_KEY)
    if (!raw)
      return defaultMatrixDisplay()
    const o = JSON.parse(raw) as Partial<MatrixDisplayPreferences>
    const wd = Array.isArray(o.weekdays) && o.weekdays.length
      ? o.weekdays.filter(n => n >= 1 && n <= 7)
      : defaultMatrixDisplay().weekdays
    const tf = o.teacherFilter === 'has_class' || o.teacherFilter === 'no_class'
      ? o.teacherFilter
      : 'all'
    return { weekdays: wd, teacherFilter: tf }
  }
  catch {
    return defaultMatrixDisplay()
  }
}

const matrixDisplay = ref<MatrixDisplayPreferences>(loadMatrixDisplayFromStorage())
const displayConfigOpen = ref(false)
const tempMatrixDisplay = ref<MatrixDisplayPreferences>({ ...loadMatrixDisplayFromStorage() })

/** 复制周课表：目标周（周选择器值） */
const copyWeekModalOpen = ref(false)
const copyTargetWeek = ref(dayjs())
const copyWeekSubmitting = ref(false)

const weekdayOptions = [
  { value: 1, label: '周一' },
  { value: 2, label: '周二' },
  { value: 3, label: '周三' },
  { value: 4, label: '周四' },
  { value: 5, label: '周五' },
  { value: 6, label: '周六' },
  { value: 7, label: '周日' },
]

function openMatrixDisplayConfig() {
  tempMatrixDisplay.value = {
    weekdays: [...matrixDisplay.value.weekdays],
    teacherFilter: matrixDisplay.value.teacherFilter,
  }
  displayConfigOpen.value = true
}

function applyMatrixDisplayConfig() {
  let wd = tempMatrixDisplay.value.weekdays.filter(n => n >= 1 && n <= 7)
  if (!wd.length)
    wd = [...defaultMatrixDisplay().weekdays]
  wd = [...new Set(wd)].sort((a, b) => a - b)
  matrixDisplay.value = {
    weekdays: wd,
    teacherFilter: tempMatrixDisplay.value.teacherFilter,
  }
  try {
    localStorage.setItem(TM_DISPLAY_STORAGE_KEY, JSON.stringify(matrixDisplay.value))
  }
  catch {}
  displayConfigOpen.value = false
  loadMatrix()
}
const timeColWidth = 84
const teacherColWidth = 168
const scheduleCardMinWidth = 152
const scheduleCardGap = 5
const scheduleColumnHorizontalInset = 6
const baseDateColumnWidth = scheduleCardMinWidth + scheduleColumnHorizontalInset * 2
const overlapExtraWidth = scheduleCardMinWidth + scheduleCardGap

function getWeekStart(value: Dayjs = dayjs()) {
  const current = dayjs(value)
  const diff = current.day() === 0 ? -6 : 1 - current.day()
  return current.add(diff, 'day').startOf('day')
}

function emitCurrentWeekRange(value: Dayjs = currentDate.value) {
  const start = getWeekStart(value)
  emit('week-range-change', {
    startDate: start.format('YYYY-MM-DD'),
    endDate: start.add(6, 'day').format('YYYY-MM-DD'),
  })
}

const displayDates = computed(() => {
  const start = getWeekStart(currentDate.value)
  return Array.from({ length: 7 }, (_, i) => start.add(i, 'day'))
})

const todayKey = computed(() => now.value.format('YYYY-MM-DD'))

const queryRange = computed(() => ({
  startDate: displayDates.value[0]?.format('YYYY-MM-DD') ?? dayjs().format('YYYY-MM-DD'),
  endDate: displayDates.value[6]?.format('YYYY-MM-DD') ?? dayjs().format('YYYY-MM-DD'),
}))

function formatWeekRange(value: Dayjs) {
  const start = getWeekStart(value)
  const end = start.add(6, 'day')
  if (start.year() === end.year() && start.month() === end.month())
    return `${start.format('YYYY年MM月DD日')} ~ ${end.format('DD日')}`
  if (start.year() === end.year())
    return `${start.format('YYYY年MM月DD日')} ~ ${end.format('MM月DD日')}`
  return `${start.format('YYYY年MM月DD日')} ~ ${end.format('YYYY年MM月DD日')}`
}

const isThisWeek = computed(() =>
  getWeekStart(currentDate.value).isSame(getWeekStart(now.value), 'day'),
)

function handlePrevWeek() {
  currentDate.value = currentDate.value.subtract(1, 'week')
}

function handleNextWeek() {
  currentDate.value = currentDate.value.add(1, 'week')
}

function handleThisWeek() {
  currentDate.value = dayjs()
}

function buildMatrixQueryParams() {
  const prefs = matrixDisplay.value
  const fullWeek = [1, 2, 3, 4, 5, 6, 7]
  const sorted = [...prefs.weekdays].sort((a, b) => a - b)
  const weekdaysStr
    = sorted.length === fullWeek.length && sorted.every((v, i) => v === fullWeek[i])
      ? undefined
      : sorted.join(',')
  const teacherFilter
    = prefs.teacherFilter === 'all' ? undefined : prefs.teacherFilter
  return { weekdaysStr, teacherFilter }
}

/** 解析下载响应里的文件名：优先 RFC5987 的 filename*=UTF-8''（与后端 url.QueryEscape 一致） */
function parseAttachmentFilenameFromHeader(cd: string | undefined): string | undefined {
  if (!cd)
    return undefined
  const star = /filename\*=(?:UTF-8'')?([^;\n]+)/i.exec(cd)
  if (star?.[1]) {
    const raw = star[1].trim().replace(/^["']|["']$/g, '')
    try {
      return decodeURIComponent(raw.replace(/\+/g, ' '))
    }
    catch {
      return raw
    }
  }
  const quoted = /filename="([^"]*)"/i.exec(cd)
  if (quoted?.[1])
    return quoted[1]
  return undefined
}

function openCopyWeekModal() {
  copyTargetWeek.value = getWeekStart(currentDate.value).add(1, 'week')
  copyWeekModalOpen.value = true
}

/**
 * 将当前 queryRange 对应周的 1 对 1 课表复制到目标周（与后端 copy-week 一致）。
 * 返回 rejected Promise 时可阻止弹窗关闭（Ant Design Vue Modal async ok）。
 */
async function handleCopyWeekConfirm() {
  const srcStart = queryRange.value.startDate
  const srcEnd = queryRange.value.endDate
  const tStart = getWeekStart(copyTargetWeek.value).format('YYYY-MM-DD')
  const tEnd = getWeekStart(copyTargetWeek.value).add(6, 'day').format('YYYY-MM-DD')
  if (srcStart === tStart && srcEnd === tEnd) {
    message.warning('目标周不能与当前周相同')
    return Promise.reject(new Error('same week'))
  }
  copyWeekSubmitting.value = true
  try {
    const res = await copyTeachingSchedulesWeekApi({
      sourceStartDate: srcStart,
      sourceEndDate: srcEnd,
      targetStartDate: tStart,
      targetEndDate: tEnd,
      classType: 2,
    })
    const raw = res.data ?? res.result
    const created = typeof raw === 'object' && raw && 'created' in raw
      ? Number((raw as { created: number }).created)
      : 0
    message.success(
      created > 0
        ? `已复制 ${created} 节日程到 ${formatWeekRange(copyTargetWeek.value)}`
        : '未新增日程（当周可能没有可复制的课，或目标周存在冲突）',
    )
    copyWeekModalOpen.value = false
    currentDate.value = getWeekStart(copyTargetWeek.value)
  }
  catch {
    return Promise.reject(new Error('copy week failed'))
  }
  finally {
    copyWeekSubmitting.value = false
  }
}

async function exportTeacherMatrixExcel() {
  try {
    const { weekdaysStr, teacherFilter } = buildMatrixQueryParams()
    const scheduleTeacherIds = filterTeacherId.value.join(',')
    const classroomIds = filterClassroomId.value.join(',')
    const res = await downloadTeachingSchedulesTeacherMatrixExcelApi({
      startDate: queryRange.value.startDate,
      endDate: queryRange.value.endDate,
      studentId: filterStudentId.value,
      scheduleTeacherIds: scheduleTeacherIds || undefined,
      classroomIds: classroomIds || undefined,
      groupClassIds: filterClassId.value ? String(filterClassId.value) : undefined,
      oneToOneClassIds: filterOneToOneId.value ? String(filterOneToOneId.value) : undefined,
      lessonIds: filterCourseId.value ? String(filterCourseId.value) : undefined,
      callStatuses: filterCallStatus.value ? String(filterCallStatus.value) : undefined,
      scheduleTypes: filterScheduleType.value.length ? filterScheduleType.value.join(',') : undefined,
      ...(weekdaysStr ? { weekdays: weekdaysStr } : {}),
      ...(teacherFilter ? { teacherFilter } : {}),
    })
    const ct = String(res.headers['content-type'] || '')
    if (ct.includes('application/json')) {
      const text = await (res.data as Blob).text()
      try {
        const j = JSON.parse(text) as { message?: string }
        message.error(j.message || '导出失败')
      }
      catch {
        message.error('导出失败')
      }
      return
    }
    const blob = new Blob([res.data as BlobPart], {
      type: ct || 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    })
    const cd = res.headers['content-disposition']
    const filename = parseAttachmentFilenameFromHeader(cd)
      ?? `课表导出_${queryRange.value.startDate}_${queryRange.value.endDate}.xlsx`
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    a.click()
    URL.revokeObjectURL(url)
    message.success('已导出课表')
  }
  catch (e) {
    console.error('export teacher matrix failed', e)
    message.error('导出失败')
  }
}

async function loadMatrix() {
  const seq = ++matrixLoadSeq
  loading.value = true
  try {
    const { weekdaysStr, teacherFilter } = buildMatrixQueryParams()
    const scheduleTeacherIds = filterTeacherId.value.join(',')
    const classroomIds = filterClassroomId.value.join(',')
    const res = await listTeachingSchedulesByTeacherMatrixApi({
      startDate: queryRange.value.startDate,
      endDate: queryRange.value.endDate,
      studentId: filterStudentId.value,
      scheduleTeacherIds: scheduleTeacherIds || undefined,
      classroomIds: classroomIds || undefined,
      groupClassIds: filterClassId.value ? String(filterClassId.value) : undefined,
      oneToOneClassIds: filterOneToOneId.value ? String(filterOneToOneId.value) : undefined,
      lessonIds: filterCourseId.value ? String(filterCourseId.value) : undefined,
      scheduleTypes: filterScheduleType.value.length ? filterScheduleType.value.join(',') : undefined,
      callStatuses: filterCallStatus.value ? String(filterCallStatus.value) : undefined,
      ...(weekdaysStr ? { weekdays: weekdaysStr } : {}),
      ...(teacherFilter ? { teacherFilter } : {}),
    })
    if (seq !== matrixLoadSeq)
      return
    if (res.code === 200 && Array.isArray(res.result))
      matrixDays.value = res.result
    else
      matrixDays.value = []
  }
  catch (e) {
    console.error('load teacher matrix failed', e)
    if (seq !== matrixLoadSeq)
      return
    matrixDays.value = []
  }
  finally {
    if (seq === matrixLoadSeq) {
      loading.value = false
      await nextTick()
      syncScroll(bodyScrollRef.value, headerDatesRef.value)
      updateFloatingDatePositions(bodyScrollRef.value?.scrollLeft ?? 0)
    }
  }
}

function syncScroll(source: HTMLElement | null, target: HTMLElement | null) {
  if (!source || !target || syncingScroll)
    return
  syncingScroll = true
  target.scrollLeft = source.scrollLeft
  requestAnimationFrame(() => {
    syncingScroll = false
  })
}

function handleHeaderScroll(event: Event) {
  const header = event.target as HTMLElement
  syncScroll(header, bodyScrollRef.value)
  updateFloatingDatePositions(
    bodyScrollRef.value?.scrollLeft ?? header.scrollLeft,
  )
}

function handleBoardScroll(event: Event) {
  const board = event.target as HTMLElement
  syncScroll(board, headerDatesRef.value)
  updateFloatingDatePositions(board.scrollLeft)
}

onMounted(() => {
  loadMatrix()
})

onUnmounted(() => {
  headerDatesResizeObserver?.disconnect()
  headerDatesResizeObserver = null
})

watch(currentDate, () => loadMatrix())

watch(currentDate, value => emitCurrentWeekRange(value), { immediate: true })

watch(
  [filterStudentId, filterTeacherId, filterClassroomId, filterClassId, filterOneToOneId, filterCourseId, filterScheduleType, filterCallStatus],
  () => loadMatrix(),
  { deep: true },
)

watch(
  () => headerDatesRef.value,
  (el) => {
    headerDatesResizeObserver?.disconnect()
    headerDatesResizeObserver = null
    if (!el || typeof ResizeObserver === 'undefined')
      return
    headerDatesResizeObserver = new ResizeObserver(() => updateFloatingDatePositions())
    headerDatesResizeObserver.observe(el)
    nextTick(() => updateFloatingDatePositions())
  },
)

const dateTeacherGroups = computed(() => {
  const days = matrixDays.value
  if (!days.length)
    return []
  return days.map((day) => {
    const cols = day.scheduleListVoList ?? []
    const dayCount = cols.reduce(
      (n, c) => n + (c.scheduleInfoVoList?.length ?? 0),
      0,
    )
    return {
      dateKey: day.scheduleDate,
      date: dayjs(day.scheduleDate),
      dayCount,
      teachers: cols.map(t => ({
        key: `id:${t.teacherId}`,
        name: t.teacherName,
        empty: false,
        count: t.scheduleInfoVoList?.length ?? 0,
      })),
    }
  })
})

function legacyToTeachingScheduleItem(
  info: TeachingScheduleMatrixLegacyItem,
  col: { teacherId: number, teacherName: string },
): TeachingScheduleItem {
  const teacherList = Array.isArray(info.teacherList) ? info.teacherList : []
  const studentList = Array.isArray(info.studentList) ? info.studentList : []
  const mainTeacher = teacherList[0]
  const assistantList = teacherList.slice(1)
  const studentText = studentList
    .map(item => String(item?.name || '').trim())
    .filter(Boolean)
    .join('、')
  const tid = mainTeacher?.id ?? col.teacherId
  const tname = mainTeacher?.name ?? col.teacherName
  const start = dayjs(`${info.scheduleDate} ${info.scheduleStartTime}`, 'YYYY-MM-DD HH:mm')
  const end = dayjs(`${info.scheduleDate} ${info.scheduleEndTime}`, 'YYYY-MM-DD HH:mm')
  return {
    id: String(info.id),
    batchNo: String(info.batchNo || '').trim() || (info.batchId != null ? String(info.batchId) : undefined),
    batchSize: 1,
    classType: info.courseType ?? 0,
    teachingClassId: info.classId != null ? String(info.classId) : '',
    teachingClassName: info.className ?? '',
    studentId: studentList[0] != null ? String(studentList[0].id) : '',
    studentName: studentText,
    lessonId: info.courseId != null ? String(info.courseId) : '',
    lessonName: info.courseName ?? '',
    teacherId: String(tid),
    teacherName: tname,
    assistantIds: assistantList.map(item => String(item.id)),
    assistantNames: assistantList.map(item => String(item.name || '').trim()).filter(Boolean),
    classroomId: '',
    classroomName: '',
    lessonDate: info.scheduleDate,
    startAt: start.format('YYYY-MM-DD HH:mm:ss'),
    endAt: end.format('YYYY-MM-DD HH:mm:ss'),
    status: info.scheduleStatus ?? 1,
    callStatus: info.callStatus ?? 1,
    callStatusText: info.callStatusText ?? (Number(info.callStatus) === 2 ? '已点名' : '未点名'),
    conflict: info.conflict === true,
    conflictTypes: info.conflictTypes ?? [],
  }
}

function resolveLessonCallStatusKey(info: TeachingScheduleMatrixLegacyItem) {
  const explicitCallStatus = Number(info?.callStatus ?? 0)
  if (explicitCallStatus === 2)
    return 'signed'
  if (explicitCallStatus === 1)
    return 'unsigned'
  const scheduleStatus = Number(info?.scheduleStatus ?? 0)
  const courseStatus = Number(info?.courseStatus ?? 0)
  const finishType = Number(info?.finishType ?? 0)
  if (finishType > 1 || courseStatus > 1 || scheduleStatus === 2)
    return 'signed'
  return 'unsigned'
}

interface CellSchedule {
  id: string
  dateKey: string
  teacherKey: string
  startAt: Dayjs
  endAt: Dayjs
  startMinutes: number
  endMinutes: number
  title: string
  course: string
  teacher: string
  classroom: string
  classType: number
  status: string
  raw: TeachingScheduleItem
  displayColumnIndex?: number
  displayColumnCount?: number
  timeText?: string
}

const internalSchedules = computed((): CellSchedule[] => {
  const out: CellSchedule[] = []
  for (const day of matrixDays.value ?? []) {
    for (const col of day.scheduleListVoList ?? []) {
      for (const info of col.scheduleInfoVoList ?? []) {
        const start = dayjs(
          `${info.scheduleDate} ${info.scheduleStartTime}`,
          'YYYY-MM-DD HH:mm',
        )
        const end = dayjs(
          `${info.scheduleDate} ${info.scheduleEndTime}`,
          'YYYY-MM-DD HH:mm',
        )
        const raw = legacyToTeachingScheduleItem(info, col)
        out.push({
          id: String(info.id),
          dateKey: day.scheduleDate,
          teacherKey: `id:${col.teacherId}`,
          startAt: start,
          endAt: end,
          startMinutes: start.hour() * 60 + start.minute(),
          endMinutes: end.hour() * 60 + end.minute(),
          title: (info.className || info.courseName || '日程') as string,
          course: info.courseName || '-',
          teacher: col.teacherName,
          classroom: '-',
          classType: info.courseType ?? 0,
          status: resolveLessonCallStatusKey(info),
          raw,
        })
      }
    }
  }
  return out
})

const flatColumns = computed(() => {
  const cols: Array<{
    dateKey: string
    teacherKey: string
    teacherName: string
    empty: boolean
    count: number
  }> = []
  for (const g of dateTeacherGroups.value) {
    for (const t of g.teachers) {
      cols.push({
        dateKey: g.dateKey,
        teacherKey: t.key,
        teacherName: t.name,
        empty: false,
        count: t.count ?? 0,
      })
    }
  }
  return cols
})

const hourMarks = computed(() =>
  Array.from(
    { length: timelineEnd / 60 - timelineStart / 60 + 1 },
    (_, i) => timelineStart + i * 60,
  ),
)

const timelineHeight = computed(
  () =>
    timelineTopPadding
    + (hourMarks.value.length - 1) * HOUR_ROW_HEIGHT_PX
    + timelineBottomPadding,
)

function minuteOffset(minutes: number) {
  return timelineTopPadding + ((minutes - timelineStart) / 60) * HOUR_ROW_HEIGHT_PX
}

function computeOverlapPeak(items: { startMinutes: number, endMinutes: number }[]) {
  const columns: number[] = []
  let peak = 1
  items.forEach((item) => {
    let columnIndex = columns.findIndex(endValue => endValue <= item.startMinutes)
    if (columnIndex === -1) {
      columnIndex = columns.length
      columns.push(item.endMinutes)
    }
    else {
      columns[columnIndex] = item.endMinutes
    }
    peak = Math.max(peak, columns.length)
  })
  return peak
}

function buildClusterLayouts(clusterItems: CellSchedule[]) {
  const columns: number[] = []
  let peakColumns = 0
  const assigned = clusterItems.map((item) => {
    let columnIndex = columns.findIndex(endValue => endValue <= item.startMinutes)
    if (columnIndex === -1) {
      columnIndex = columns.length
      columns.push(item.endMinutes)
    }
    else {
      columns[columnIndex] = item.endMinutes
    }
    peakColumns = Math.max(peakColumns, columns.length)
    return {
      ...item,
      columnIndex,
      timeText: `${item.startAt.format('HH:mm')} - ${item.endAt.format('HH:mm')}`,
    }
  })
  return assigned.map(item => ({
    ...item,
    displayColumnIndex: item.columnIndex,
    displayColumnCount: peakColumns,
  }))
}

function buildDayLayouts(items: CellSchedule[]) {
  const sorted = [...items].sort((a, b) => a.startAt.valueOf() - b.startAt.valueOf())
  const clusters: CellSchedule[][] = []
  let currentCluster: CellSchedule[] = []
  let currentEnd = -1
  sorted.forEach((item) => {
    if (currentCluster.length === 0) {
      currentCluster = [item]
      currentEnd = item.endMinutes
      return
    }
    if (item.startMinutes < currentEnd) {
      currentCluster.push(item)
      currentEnd = Math.max(currentEnd, item.endMinutes)
      return
    }
    clusters.push(currentCluster)
    currentCluster = [item]
    currentEnd = item.endMinutes
  })
  if (currentCluster.length)
    clusters.push(currentCluster)
  return clusters.flatMap(cluster => buildClusterLayouts(cluster))
}

const layoutsByCell = computed(() => {
  const map = new Map<string, { layouts: CellSchedule[], colWidth: number }>()
  for (const col of flatColumns.value) {
    const list = internalSchedules.value.filter(
      s => s.dateKey === col.dateKey && s.teacherKey === col.teacherKey,
    )
    const peak = computeOverlapPeak(list)
    const colWidth
      = baseDateColumnWidth + Math.max(0, peak - 1) * overlapExtraWidth
    map.set(`${col.dateKey}|${col.teacherKey}`, {
      layouts: buildDayLayouts(list),
      colWidth: Math.max(teacherColWidth, colWidth),
    })
  }
  return map
})

const gridTemplateStyleHeader = computed(() => {
  const n = flatColumns.value.length
  if (!n) {
    return {
      gridTemplateColumns: '1fr',
      width: '100%' as const,
      minWidth: '100%' as const,
    }
  }
  const tracks = flatColumns.value.map((col) => {
    const entry = layoutsByCell.value.get(`${col.dateKey}|${col.teacherKey}`)
    const w = Math.max(teacherColWidth, entry?.colWidth ?? teacherColWidth)
    return `${w}px`
  })
  return {
    gridTemplateColumns: tracks.join(' '),
    width: 'max-content' as const,
    minWidth: '100%' as const,
  }
})

watch(gridTemplateStyleHeader, () => nextTick(() => updateFloatingDatePositions()))

const gridTemplateStyle = computed(() => {
  const n = flatColumns.value.length
  if (!n) {
    return {
      gridTemplateColumns: `${timeColWidth}px`,
      width: '100%' as const,
      minWidth: '100%' as const,
    }
  }
  const tracks = flatColumns.value.map((col) => {
    const entry = layoutsByCell.value.get(`${col.dateKey}|${col.teacherKey}`)
    const w = Math.max(teacherColWidth, entry?.colWidth ?? teacherColWidth)
    return `${w}px`
  })
  return {
    gridTemplateColumns: `${timeColWidth}px ${tracks.join(' ')}`,
    width: 'max-content' as const,
    minWidth: '100%' as const,
  }
})

function eventStyle(event: CellSchedule, col: { dateKey: string, teacherKey: string }) {
  const entry = layoutsByCell.value.get(`${col.dateKey}|${col.teacherKey}`)
  const colWidth = entry?.colWidth ?? teacherColWidth
  const leftOffset
    = (event.displayColumnIndex || 0) * (scheduleCardMinWidth + scheduleCardGap)
      + scheduleColumnHorizontalInset
  return {
    top: `${minuteOffset(event.startMinutes)}px`,
    height: `${Math.max(
      82,
      ((event.endMinutes - event.startMinutes) / 60) * HOUR_ROW_HEIGHT_PX,
    )}px`,
    left: `${leftOffset}px`,
    width: `${Math.min(scheduleCardMinWidth, colWidth - scheduleColumnHorizontalInset * 2)}px`,
  }
}

function isActiveDate(dateKey: string) {
  return dateKey === todayKey.value
}

function formatClock(minutes: number) {
  const hour = String(Math.floor(minutes / 60)).padStart(2, '0')
  const minute = String(minutes % 60).padStart(2, '0')
  return `${hour}:${minute}`
}

const currentTimeMinutes = computed(() => now.value.hour() * 60 + now.value.minute())
const currentTimeLabel = computed(() => now.value.format('HH:mm'))
const showCurrentTimeLine = computed(() => {
  if (
    currentTimeMinutes.value < timelineStart
    || currentTimeMinutes.value > timelineEnd
  ) {
    return false
  }
  return displayDates.value.some(d => d.format('YYYY-MM-DD') === todayKey.value)
})

function openScheduleEdit(event: CellSchedule) {
  const schedule = event.raw
  currentDetailSchedule.value = schedule
  currentScheduleDetail.value = buildScheduleDrawerDetail(schedule)
  scheduleDetailOpen.value = true
}

function onBatchPlanUpdated() {
  scheduleBatchPlanEditOpen.value = false
  scheduleDetailOpen.value = false
  loadMatrix()
}

const isCurrentDetailEditable = computed(() => Boolean(currentDetailSchedule.value?.id))

async function openEventConflictDetail(event: CellSchedule) {
  if (!event?.raw?.conflict)
    return

  scheduleConflictLoading.value = true
  try {
    const res = await getTeachingScheduleConflictDetailApi({
      id: String(event.id),
    })
    if (res.code !== 200 || !res.result)
      throw new Error(res.message || '加载冲突详情失败')
    scheduleConflictValidation.value = res.result
    scheduleConflictOpen.value = true
  }
  catch (error: any) {
    console.error('openEventConflictDetail failed', error)
    messageService.error(error?.response?.data?.message || error?.message || '加载冲突详情失败')
  }
  finally {
    scheduleConflictLoading.value = false
  }
}

function handleScheduleDetailDelete() {
  const schedule = currentDetailSchedule.value
  const scheduleId = String(schedule?.id || '').trim()
  if (!scheduleId) {
    message.warning('当前日程缺少删除标识，请刷新后重试')
    return
  }

  Modal.confirm({
    title: '删除日程?',
    content: '删除后将不可恢复，请谨慎操作',
    okText: '删除',
    cancelText: '取消',
    async onOk() {
      deletingScheduleDetail.value = true
      try {
        const res = await cancelTeachingSchedulesApi({
          ids: [scheduleId],
        })
        if (res.code !== 200)
          throw new Error(res.message || '删除日程失败')
        scheduleDetailOpen.value = false
        currentDetailSchedule.value = null
        currentScheduleDetail.value = null
        message.success(`已删除${isOneToOneSchedule(schedule) ? '1对1' : '班课'}日程`)
        await loadMatrix()
      }
      catch (error: any) {
        console.error('delete schedule detail failed', error)
        message.error(error?.response?.data?.message || error?.message || '删除日程失败')
        throw error
      }
      finally {
        deletingScheduleDetail.value = false
      }
    },
  })
}

function handleScheduleDetailEdit() {
  const schedule = currentDetailSchedule.value
  if (!schedule?.id)
    return
  currentBatchPlanSchedule.value = schedule
  scheduleBatchPlanEditOpen.value = true
}

const totalLessons = computed(() => internalSchedules.value.length)
const unsignedLessons = computed(() =>
  internalSchedules.value.filter(item => item.status === 'unsigned').length,
)
</script>

<template>
  <div class="tm-api-root">
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
        :whole-condition-clear-types="['scheduleTeacher', 'scheduleClassroom', 'scheduleType']"
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
    <div class="tm-api-toolbar-card mt2">
      <div class="toolbar-main">
        <div class="toolbar-group tm-toolbar-ghost" aria-hidden="true">
          <a-radio-group button-style="solid" size="small">
            <a-radio-button value="day">
              日
            </a-radio-button>
            <a-radio-button value="week">
              周
            </a-radio-button>
          </a-radio-group>
        </div>

        <div class="toolbar-date time-selector ml3 font-800 text-5 flex-center">
          <a-popover trigger="hover">
            <template #content>
              上一周
            </template>
            <span
              class="cursor-pointer text-3 text-#888 flex w6 h6 bg-#eee rounded-10 flex-center font-500 hover-text-#06f hover-bg-#e6f0ff"
              @click="handlePrevWeek"
            >
              <LeftOutlined />
            </span>
          </a-popover>
          <span class="mx-2">
            <div
              class="relative cursor-pointer toolbar-date-range"
              :class="isThisWeek ? 'text-#0061ff' : 'text-#222'"
            >
              {{ formatWeekRange(currentDate) }}
              <a-date-picker
                v-model:value="currentDate"
                class="absolute left-0 top-0 right-0 bottom-0 z-10 opacity-0"
                picker="week"
                :allow-clear="false"
                :bordered="false"
                :format="formatWeekRange"
                style="cursor: pointer"
              />
            </div>
          </span>
          <a-popover trigger="hover">
            <template #content>
              下一周
            </template>
            <span
              class="cursor-pointer text-3 text-#888 flex w6 h6 bg-#eee rounded-10 flex-center font-500 hover-text-#06f hover-bg-#e6f0ff"
              @click="handleNextWeek"
            >
              <RightOutlined />
            </span>
          </a-popover>
          <a-popover trigger="hover">
            <template #content>
              回到本周
            </template>
            <a-button
              type="default"
              size="small"
              class="toolbar-today-week-btn ml2"
              :class="{
                'toolbar-today-week-btn--active': isThisWeek,
                'toolbar-today-week-btn--inactive': !isThisWeek,
              }"
              @click="handleThisWeek"
            >
              本周
            </a-button>
          </a-popover>
        </div>

        <a-space>
          <a-button
            type="default"
            size="small"
            @click="openCopyWeekModal"
          >
            <template #icon>
              <CopyOutlined />
            </template>
            复制周课表
          </a-button>
          <a-button
            type="default"
            size="small"
            @click="exportTeacherMatrixExcel"
          >
            <template #icon>
              <DownloadOutlined />
            </template>
            导出课表
          </a-button>
          <a-button
            type="default"
            size="small"
            @click="openMatrixDisplayConfig"
          >
            <template #icon>
              <SettingOutlined />
            </template>
            展示配置
          </a-button>
        </a-space>
      </div>
    </div>

    <a-modal
      v-model:open="displayConfigOpen"
      title="课表展示配置"
      width="620px"
      ok-text="确定"
      cancel-text="取消"
      @ok="applyMatrixDisplayConfig"
    >
      <div class="tm-display-config">
        <div class="tm-dc-row">
          <div class="tm-dc-head">
            <span class="tm-dc-title">日期展示维度</span>
          </div>
          <a-select
            v-model:value="tempMatrixDisplay.weekdays"
            mode="multiple"
            placeholder="选择星期"
            style="width: 100%"
            :options="weekdayOptions"
          />
        </div>
        <div class="tm-dc-row">
          <div class="tm-dc-head">
            <span class="tm-dc-title">筛选老师</span>
            <span class="tm-dc-hint">按日当天日程筛列</span>
          </div>
          <a-radio-group v-model:value="tempMatrixDisplay.teacherFilter" class="custom-radio">
            <a-radio value="all">
              全部老师
            </a-radio>
            <a-radio value="has_class">
              仅有课老师
            </a-radio>
            <a-radio value="no_class">
              仅无课老师
            </a-radio>
          </a-radio-group>
        </div>
      </div>
    </a-modal>

    <a-modal
      v-model:open="copyWeekModalOpen"
      title="复制周课表"
      ok-text="开始复制"
      cancel-text="取消"
      :confirm-loading="copyWeekSubmitting"
      @ok="handleCopyWeekConfirm"
    >
      <p class="tm-copy-week-hint">
        将当前周 <strong>{{ formatWeekRange(currentDate) }}</strong> 的 <strong>1 对 1</strong> 课表按星期对齐复制到目标周（批量排课会保持新的批量关系）。
      </p>
      <div class="tm-copy-week-picker">
        <span class="tm-copy-week-label">复制到</span>
        <a-date-picker
          v-model:value="copyTargetWeek"
          picker="week"
          :allow-clear="false"
          style="width: 100%"
          :format="formatWeekRange"
        />
      </div>
    </a-modal>

    <div class="tm-api-card">
      <a-spin :spinning="loading" :delay="120" size="small" class="tm-api-spin">
        <div class="tm-sticky-shell">
          <div class="tm-api-summary">
            <TimetableScheduleSummary
              :total="totalLessons"
              :unsigned-count="unsignedLessons"
            />
          </div>

          <div v-if="matrixDays.length" class="tm-schedule-header">
            <div class="tm-header-time-corner">
              <div class="tm-header-time-corner__label">
                时间
              </div>
            </div>
            <div
              ref="headerDatesRef"
              class="tm-header-dates"
              @scroll="handleHeaderScroll"
            >
              <!-- 与旧版一致：浮动层放在「与列同宽的滚动内容」里，left 才用内容坐标 finalPosition -->
              <div class="tm-header-dates-track">
                <div class="tm-floating-date-layer">
                  <div
                    v-for="g in dateTeacherGroups"
                    :key="`float-${g.dateKey}`"
                    class="tm-floating-chip"
                    :class="{ 'tm-floating-chip--today': isActiveDate(g.dateKey) }"
                    :style="floatingPillStyle(g.dateKey)"
                  >
                    <div class="tm-floating-chip__line">
                      <span class="tm-floating-chip__date">{{ g.date.format('M/D') }}</span>
                      <span class="tm-floating-chip__week">{{
                        ['周日', '周一', '周二', '周三', '周四', '周五', '周六'][g.date.day()]
                      }}</span>
                    </div>
                    <div class="tm-floating-chip__meta">
                      共 <strong>{{ g.dayCount }}</strong> 节
                    </div>
                  </div>
                </div>
                <div class="tm-header-grid" :style="gridTemplateStyleHeader">
                  <div
                    v-for="(g, gi) in dateTeacherGroups"
                    :key="g.dateKey"
                    class="tm-date-banner"
                    :class="{
                      'tm-date-banner--active': isActiveDate(g.dateKey),
                      'tm-date-banner--day-divider': gi < dateTeacherGroups.length - 1,
                    }"
                    :style="{ gridColumn: `span ${g.teachers.length}` }"
                  >
                    <div class="tm-date-banner__inner tm-date-banner__inner--ghost" aria-hidden="true">
                      <span class="tm-date-banner__date">{{ g.date.format('M/D') }}</span>
                      <span class="tm-date-banner__week">{{ ['周日', '周一', '周二', '周三', '周四', '周五', '周六'][g.date.day()] }}</span>
                      <span class="tm-date-banner__count">共 {{ g.dayCount }} 节</span>
                    </div>
                  </div>
                </div>
                <div class="tm-subheader-grid" :style="gridTemplateStyleHeader">
                  <div
                    v-for="(g, gi) in dateTeacherGroups"
                    :key="`w-${g.dateKey}`"
                    class="tm-teacher-head-group"
                  >
                    <div
                      v-for="(t, ti) in g.teachers"
                      :key="`${g.dateKey}-${t.key}`"
                      class="tm-teacher-head"
                      :class="{
                        'tm-teacher-head--active': isActiveDate(g.dateKey),
                        'tm-teacher-head--has-class': (t.count ?? 0) > 0,
                        'tm-teacher-head--no-class': (t.count ?? 0) === 0,
                        'tm-teacher-head--day-divider':
                          ti === g.teachers.length - 1 && gi < dateTeacherGroups.length - 1,
                      }"
                    >
                      <div class="tm-teacher-head__avatar">
                        {{ (t.name || '?').slice(0, 1) }}
                      </div>
                      <div class="tm-teacher-head__meta">
                        <div class="tm-teacher-head__name">
                          {{ t.name }}
                        </div>
                        <div
                          class="tm-teacher-head__count"
                          :class="{
                            'tm-teacher-head__count--zero': (t.count ?? 0) === 0,
                            'tm-teacher-head__count--has': (t.count ?? 0) > 0,
                          }"
                        >
                          共 {{ t.count ?? 0 }} 节
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div v-if="!matrixDays.length && !loading" class="tm-api-empty">
          本周暂无数据
        </div>

        <div
          v-if="matrixDays.length"
          ref="bodyScrollRef"
          class="tm-schedule-board"
          @scroll="handleBoardScroll"
        >
          <div class="tm-board-grid" :style="gridTemplateStyle">
            <div class="tm-time-axis">
              <div
                v-for="mark in hourMarks"
                :key="mark"
                class="tm-time-label"
                :style="{ top: `${minuteOffset(mark)}px` }"
              >
                <span>{{ formatClock(mark) }}</span>
              </div>
              <div
                v-if="showCurrentTimeLine"
                class="tm-now-line"
                :style="{ top: `${minuteOffset(currentTimeMinutes)}px` }"
              >
                <span class="tm-now-line__text">{{ currentTimeLabel }}</span>
              </div>
            </div>

            <div
              v-for="(col, ci) in flatColumns"
              :key="`${col.dateKey}-${col.teacherKey}`"
              class="tm-column"
              :class="{
                'tm-column--active': isActiveDate(col.dateKey),
                'tm-column--no-class': col.count === 0,
                'tm-column--has-class': col.count > 0,
                'tm-column--day-divider':
                  ci < flatColumns.length - 1 && flatColumns[ci + 1].dateKey !== col.dateKey,
              }"
            >
              <div
                class="tm-column__body"
                :style="{ height: `${timelineHeight}px` }"
              >
                <div
                  v-for="mark in hourMarks"
                  :key="`${col.dateKey}-${col.teacherKey}-${mark}`"
                  class="tm-column__line"
                  :style="{ top: `${minuteOffset(mark)}px` }"
                />
                <div
                  v-if="showCurrentTimeLine && col.dateKey === todayKey"
                  class="tm-now-marker"
                  :style="{ top: `${minuteOffset(currentTimeMinutes)}px` }"
                />

                <TimetableScheduleHoverPopover
                  v-for="event in (layoutsByCell.get(`${col.dateKey}|${col.teacherKey}`)?.layouts ?? [])"
                  :key="event.id"
                  :schedule-id="String(event.id || '')"
                  :mode-label="scheduleBadgeText(event.classType)"
                  :lesson-title="scheduleHoverTitle(event.raw)"
                  :teacher-name="event.teacher"
                  :course-name="event.course"
                  :assistant-text="scheduleAssistantText(event.raw)"
                  :student-text="scheduleStudentSummary(event.raw)"
                  :classroom-name="event.classroom"
                  :time-text="scheduleTimeTextFromEvent(event)"
                  :conflict-text="scheduleConflictSummary(event.raw)"
                  @detail="openScheduleEdit(event)"
                >
                  <div
                    class="tm-event"
                    :class="{ 'tm-event--conflict': event.raw?.conflict }"
                    :style="eventStyle(event, col)"
                    @click="openScheduleEdit(event)"
                  >
                    <div class="tm-event__top">
                      <div class="tm-event__time">
                        {{ event.timeText }}
                      </div>
                      <div class="tm-event__badges">
                        <a-tooltip v-if="event.raw?.conflict" :title="conflictBadgeTitle(event)" placement="top" @click.stop>
                          <span
                            class="tm-event__badge tm-event__badge--conflict"
                            @click.stop="openEventConflictDetail(event)"
                          >
                            冲突
                          </span>
                        </a-tooltip>
                        <span
                          v-else-if="event.classType === 1 || event.classType === 2"
                          class="tm-event__badge"
                          :class="event.classType === 1 ? 'tm-event__badge--group-class' : 'tm-event__badge--one-to-one'"
                        >
                          {{ scheduleBadgeText(event.classType) }}
                        </span>
                      </div>
                    </div>
                    <div class="tm-event__body">
                      <div class="tm-event__title">
                        {{ event.title }}
                      </div>
                      <div class="tm-event__meta">
                        {{ event.course }}
                      </div>
                      <div class="tm-event__meta tm-event__meta--muted">
                        {{ event.teacher }}
                      </div>
                    </div>
                  </div>
                </TimetableScheduleHoverPopover>
              </div>
            </div>
          </div>
        </div>
      </a-spin>
    </div>

    <SmartTimetableScheduleDetailDrawer
      v-model:open="scheduleDetailOpen"
      :detail="currentScheduleDetail"
      :deleting="deletingScheduleDetail"
      :editable="isCurrentDetailEditable"
      @delete="handleScheduleDetailDelete"
      @edit="handleScheduleDetailEdit"
      @updated="onBatchPlanUpdated"
    />
    <ScheduleBatchPlanEditModal
      v-model:open="scheduleBatchPlanEditOpen"
      :schedule="currentBatchPlanSchedule"
      @updated="onBatchPlanUpdated"
    />
    <ScheduleConflictModal
      v-model:open="scheduleConflictOpen"
      :validation="scheduleConflictValidation"
      :locating="scheduleConflictLoading"
      title="冲突详情"
      current-title="当前冲突日程"
      existing-title="与其冲突的日程"
      fallback-message="当前日程与已有日程存在冲突"
    />
  </div>
</template>

<style scoped lang="less">
.tm-api-root {
  min-width: 0;
}

.tm-api-toolbar-card,
.tm-api-card {
  border: 1px solid #e5ebf3;
  border-radius: 16px;
  background: #fff;
  box-shadow: 0 10px 24px rgb(15 23 42 / 4%);
  overflow: visible;
}

.tm-api-toolbar-card {
  padding: 14px 18px;
  border-bottom: none;
  border-bottom-left-radius: 0;
  border-bottom-right-radius: 0;
}

.toolbar-main {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.toolbar-group {
  display: flex;
  align-items: center;
  gap: 12px;
}

.toolbar-date {
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 1;
  min-width: 0;
}

.toolbar-date-range {
  display: inline-block;
  width: 240px;
  min-width: 240px;
  max-width: 240px;
  text-align: center;
}

.tm-toolbar-ghost {
  visibility: hidden;
  pointer-events: none;
}

.time-selector {
  font-family: DINAlternate-Bold, DINAlternate;
  gap: 6px;
}

.toolbar-today-week-btn {
  padding: 0 10px;
  border-radius: 8px;
  transition:
    color 0.18s ease,
    border-color 0.18s ease,
    background-color 0.18s ease,
    box-shadow 0.18s ease;
}

.toolbar-today-week-btn--active {
  color: #1677ff;
  border-color: #91caff;
  background: #f5f9ff;
  box-shadow: inset 0 0 0 1px rgb(255 255 255 / 72%);
}

.toolbar-today-week-btn--inactive {
  color: #222;
}

.tm-api-toolbar-card :deep(.ant-radio-button-wrapper) {
  padding: 0 16px;
}

.tm-display-config {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding: 4px 0 8px;
}

.tm-copy-week-hint {
  margin: 0 0 16px;
  color: #475569;
  font-size: 14px;
  line-height: 1.55;
}

.tm-copy-week-picker {
  display: flex;
  align-items: center;
  gap: 12px;
}

.tm-copy-week-label {
  flex-shrink: 0;
  color: #334155;
  font-size: 14px;
  font-weight: 600;
}

.tm-dc-row {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.tm-dc-head {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 8px;
}

.tm-dc-title {
  color: #1f2937;
  font-size: 14px;
  font-weight: 700;
}

.tm-dc-hint {
  color: #94a3b8;
  font-size: 12px;
  font-weight: 500;
}

.tm-api-card {
  border-top: none;
  border-top-left-radius: 0;
  border-top-right-radius: 0;
}

.tm-api-spin {
  display: block;
  width: 100%;

  :deep(.ant-spin-nested-loading) {
    width: 100%;
  }

  :deep(.ant-spin-container) {
    overflow: visible;
  }
}

/* 与时间课表 .schedule-sticky-shell 一致：统计 + 表头随页面滚动吸顶 */
.tm-sticky-shell {
  position: sticky;
  top: 8px;
  z-index: 40;
  background: #fff;
  box-shadow: 0 10px 22px rgb(15 23 42 / 6%);
}

.tm-api-summary {
  background: rgb(255 255 255 / 98%);
}

.tm-api-empty {
  padding: 48px;
  text-align: center;
  color: #94a3b8;
  font-size: 14px;
}

.tm-schedule-header {
  display: flex;
  min-height: 104px;
  border-bottom: 1px solid #dde5f0;
  background: #fff;
}

.tm-header-time-corner {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  width: 84px;
  min-width: 84px;
  border-right: 1px solid #dde5f0;
  background: #eff4fb;
  color: #1f2937;
  font-size: 13px;
  font-weight: 700;
}

.tm-header-dates {
  flex: 1;
  min-width: 0;
  overflow-x: auto;
  overflow-y: hidden;
  scrollbar-width: none;
  -ms-overflow-style: none;

  &::-webkit-scrollbar {
    display: none;
  }
}

.tm-header-dates-track {
  position: relative;
  width: max-content;
  min-width: 100%;
}

.tm-floating-date-layer {
  position: absolute;
  top: 0;
  left: 0;
  z-index: 4;
  width: 100%;
  height: 48px;
  min-height: 48px;
  pointer-events: none;
}

/**
 * 浮动日期：独立「芯片」样式，与表头 #eff4fb 协调，盖住格线又不像弹层卡片
 */
.tm-floating-chip {
  position: absolute;
  top: 50%;
  left: 0;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 3px;
  min-height: 44px;
  padding: 6px 12px 7px;
  border: 1px solid #cdd9ea;
  border-radius: 8px;
  background: #fff;
  box-shadow:
    0 1px 2px rgb(15 23 42 / 6%),
    0 0 0 1px rgb(255 255 255 / 80%) inset;
  text-align: center;
  isolation: isolate;
  transform: translateY(-50%);
}

.tm-floating-chip__line {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  justify-content: center;
  gap: 6px;
  line-height: 1.15;
}

.tm-floating-chip__date {
  color: #0f172a;
  font-size: 13px;
  font-weight: 700;
  letter-spacing: -0.02em;
}

.tm-floating-chip__week {
  color: #64748b;
  font-size: 12px;
  font-weight: 600;
}

.tm-floating-chip__meta {
  margin: 0;
  color: #94a3b8;
  font-size: 11px;
  font-weight: 500;
  line-height: 1.2;
  letter-spacing: 0.02em;
}

.tm-floating-chip__meta strong {
  color: #64748b;
  font-weight: 700;
}

.tm-floating-chip--today {
  border-color: #91caff;
  background: #f5f9ff;
  box-shadow:
    0 1px 3px rgb(22 119 255 / 12%),
    inset 0 0 0 1px rgb(255 255 255 / 90%);
}

.tm-floating-chip--today::before {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  width: 34px;
  height: 3px;
  border-radius: 0 0 4px 4px;
  background: #1677ff;
  transform: translateX(-50%);
}

.tm-floating-chip--today .tm-floating-chip__date {
  color: #0958d9;
}

.tm-floating-chip--today .tm-floating-chip__week {
  color: #1677ff;
}

.tm-floating-chip--today .tm-floating-chip__meta,
.tm-floating-chip--today .tm-floating-chip__meta strong {
  color: #1677ff;
}

.tm-floating-chip--today .tm-floating-chip__meta {
  opacity: 0.9;
}

/* 与时间课表 .schedule-board：横向滚动 + 纵向由页面承担 */
.tm-schedule-board {
  position: relative;
  overflow-x: auto;
  overflow-y: visible;
  background: #fff;
  scrollbar-width: thin;
  scrollbar-color: #b8c9de #eef2f8;

  &::-webkit-scrollbar {
    height: 8px;
  }

  &::-webkit-scrollbar-track {
    margin: 0 4px;
    border-radius: 999px;
    background: #eef2f8;
  }

  &::-webkit-scrollbar-thumb {
    border-radius: 999px;
    background: linear-gradient(180deg, #c5d4ea 0%, #a8bad4 100%);
    box-shadow: inset 0 1px 0 rgb(255 255 255 / 45%);

    &:hover {
      background: linear-gradient(180deg, #a8bad4 0%, #8fa3bd 100%);
    }
  }
}

.tm-header-grid,
.tm-subheader-grid,
.tm-board-grid {
  display: grid;
  width: max-content;
  min-width: 100%;
}

.tm-date-banner {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 48px;
  padding: 8px;
  border-right: 1px solid #dde5f0;
  border-bottom: 1px solid #dde5f0;
  background: #eff4fb;
  text-align: center;
}

.tm-date-banner--active {
  color: #1677ff;
  box-shadow: inset 0 3px 0 #1677ff;
}

/* 相邻两天的分界：略深、略粗，区别于同一天内教师列 */
.tm-date-banner--day-divider {
  border-right-width: 2px;
  border-right-color: #a8b8cc;
}

.tm-date-banner__inner {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  gap: 6px 10px;
}

.tm-date-banner__inner--ghost {
  visibility: hidden;
}

.tm-date-banner__date {
  color: #374151;
  font-size: 14px;
  font-weight: 700;
}

.tm-date-banner__week {
  color: #6b7280;
  font-size: 13px;
  font-weight: 600;
}

.tm-date-banner__count {
  color: #6b7280;
  font-size: 13px;
  font-weight: 600;
}

.tm-subheader-grid {
  background: #fff;
  border-bottom: 1px solid #dde5f0;
}

.tm-teacher-head-group {
  display: contents;
}

.tm-teacher-head {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 10px 12px;
  border-right: 1px solid #dde5f0;
  background: #eff4fb;
  min-width: 0;
}

.tm-teacher-head--active {
  background: #f3f9ff;
}

.tm-teacher-head--has-class {
  box-shadow: inset 0 0 0 1px rgb(22 119 255 / 12%);
}

.tm-teacher-head--day-divider {
  border-right-width: 2px;
  border-right-color: #a8b8cc;
}

.tm-teacher-head__count--zero {
  color: #9ca3af;
  font-weight: 600;
}

.tm-teacher-head__count--has {
  color: #1677ff;
  font-weight: 700;
}

.tm-teacher-head__meta {
  min-width: 0;
  text-align: center;
}

.tm-teacher-head__avatar {
  flex-shrink: 0;
  width: 32px;
  height: 32px;
  border-radius: 999px;
  background: #e5e7eb;
  color: #4b5563;
  font-size: 13px;
  font-weight: 700;
  line-height: 32px;
  text-align: center;
}

.tm-teacher-head__name {
  overflow: hidden;
  color: #374151;
  font-size: 13px;
  font-weight: 700;
  line-height: 1.25;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tm-teacher-head__count {
  margin-top: 2px;
  font-size: 12px;
  font-weight: 600;
}

.tm-time-axis {
  position: relative;
  border-right: 1px solid #dde5f0;
  background: #fff;
  box-shadow: 4px 0 14px -6px rgb(15 23 42 / 12%);
}

.tm-board-grid .tm-time-axis {
  position: sticky;
  left: 0;
  z-index: 3;
  min-width: 84px;
  max-width: 84px;
}

.tm-time-label {
  position: absolute;
  left: 0;
  right: 0;
  transform: translateY(-50%);
  text-align: center;
  font-size: 14px;
  pointer-events: none;
}

.tm-time-label span {
  position: relative;
  z-index: 1;
  display: inline-block;
  padding: 0 10px;
  background: #fff;
  color: #1f2937;
  font-weight: 600;
}

.tm-time-label::before {
  position: absolute;
  left: 0;
  right: 0;
  top: 50%;
  border-top: 1px solid #dde5f0;
  content: "";
  z-index: 0;
}

.tm-now-line {
  position: absolute;
  left: 0;
  right: 0;
  z-index: 3;
  transform: translateY(-50%);
  pointer-events: none;
}

.tm-now-line__text {
  position: absolute;
  top: -7px;
  left: -3px;
  right: -4px;
  color: #ff4d4f;
  font-size: 12px;
  font-weight: 600;
  text-align: center;
}

.tm-column {
  position: relative;
  border-right: 1px solid #dde5f0;
  background: #fff;
}

.tm-column--day-divider {
  border-right-width: 2px;
  border-right-color: #a8b8cc;
}

.tm-column--active {
  background: #f3f9ff;
}

.tm-column--no-class {
  background: #f9fafb;
}

.tm-column--has-class {
  background: #fff;
}

.tm-column__body {
  position: relative;
}

.tm-column__line {
  position: absolute;
  left: 0;
  right: 0;
  border-top: 1px solid #dde5f0;
  pointer-events: none;
}

.tm-now-marker {
  position: absolute;
  left: 0;
  right: 0;
  z-index: 2;
  border-top: 1px solid #ffb3b3;
  pointer-events: none;
}

.tm-event {
  position: absolute;
  z-index: 2;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border-radius: 4px;
  background: #fff;
  box-shadow: 0 6px 16px rgb(22 119 255 / 10%);
  cursor: pointer;
}

.tm-event--conflict {
  box-shadow: 0 0 0 2px rgba(255, 77, 79, 0.4);
}

.tm-event__top {
  position: relative;
  display: flex;
  align-items: center;
  min-height: 24px;
  padding: 3px 56px 3px 10px;
  background: #1677ff;
}

.tm-event__time {
  flex: 1;
  min-width: 0;
  color: #fff;
  font-size: 12px;
  font-weight: 600;
  white-space: nowrap;
}

.tm-event__badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 16px;
  padding: 0 7px;
  border-radius: 10px;
  background: rgb(9 61 149 / 24%);
  color: #fff;
  font-size: 10px;
  font-weight: 700;
  line-height: 1;
}

.tm-event__badges {
  position: absolute;
  top: 0;
  right: 0;
  display: flex;
  flex-direction: row-reverse;
  align-items: flex-start;
  gap: 4px;
  flex-shrink: 0;
}

.tm-event__badge--one-to-one {
  padding: 0 8px 0 9px;
  border-radius: 0 4px 0 8px;
  background: rgb(0 0 0 / 50%);
}

.tm-event__badge--group-class {
  padding: 0 8px 0 9px;
  border-radius: 0 4px 0 8px;
  background: #d46b08;
}

.tm-event__badge--conflict {
  padding: 0 8px 0 9px;
  border-radius: 0 4px 0 8px;
  background: #ff4d4f;
  cursor: pointer;
}

.tm-event__body {
  display: flex;
  flex-direction: column;
  padding: 4px 0 0 10px;
}

.tm-event__title {
  display: -webkit-box;
  overflow: hidden;
  color: #0f172a;
  font-size: 13px;
  font-weight: 700;
  line-height: 1.25;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.tm-event__meta {
  overflow: hidden;
  color: #64748b;
  font-size: 12px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tm-event__meta--muted {
  color: #334155;
  font-weight: 600;
}
</style>
