<script setup>
import { Modal } from 'ant-design-vue'
import dayjs from 'dayjs'
import { h } from 'vue'
import SmartTimetableConflictModal from './smart-timetable-conflict-modal.vue'
import SmartTimetableDragConfirmModal from './smart-timetable-drag-confirm-modal.vue'
import SmartTimetableDragConflictModal from './smart-timetable-drag-conflict-modal.vue'
import SmartTimetableGrid from './smart-timetable-grid.vue'
import SmartTimetableScheduleDetailDrawer from './smart-timetable-schedule-detail-drawer.vue'
import SmartTimetableToolbar from './smart-timetable-toolbar.vue'
import ScheduleBatchPlanEditModal from './schedule-batch-plan-edit-modal.vue'
import ScheduleConflictModal from './schedule-conflict-modal.vue'
import TimetableScheduleSummary from './timetable-schedule-summary.vue'
import { listClassroomsApi } from '@/api/business-settings/classroom'
import { getInstPeriodConfigApi } from '@/api/common/config'
import { getOneToOneListApi } from '@/api/edu-center/one-to-one'
import { pageGroupClassesApi } from '@/api/edu-center/group-class'
import { getCourseIdAndNameApi } from '@/api/edu-center/registr-renewal'
import { batchUpdateTeachingSchedulesApi, cancelTeachingSchedulesApi, checkAssistantScheduleAvailabilityApi, checkOneToOneScheduleAvailabilityApi, createGroupClassSchedulesApi, createOneToOneSchedulesApi, downloadSmartTimetableExcelApi, getTeachingScheduleConflictDetailApi, listTeachingSchedulesByTeacherMatrixApi, validateOneToOneSchedulesApi } from '@/api/edu-center/teaching-schedule'
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
  periodGroupIndexForKey,
  periodGroupKeyForIndex,
} from '@/utils/unified-time-period'
import emitter, { EVENTS } from '@/utils/eventBus'

const emit = defineEmits(['week-range-change'])

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
const allFilterRef = ref(null)
const classId = ref(null)
const exportLoading = ref(false)
const filterStudentId = ref(undefined)
const filterTeacherId = ref([])
const filterClassroomId = ref([])
const filterClassId = ref(undefined)
const filterOneToOneId = ref(undefined)
const filterCourseId = ref(undefined)
const filterScheduleType = ref([])
const filterCallStatus = ref(undefined)
const schedulingClassroomList = ref([])
const schedulingClassroomLoading = ref(false)
const selectedOneToOneClassroomId = ref(undefined)
const selectedClassClassroomId = ref(undefined)

const scheduleTypeOptions = [
  { id: 'group_class', value: '班级日程' },
  { id: 'one_to_one', value: '1对1日程' },
  { id: 'trial', value: '试听日程' },
]

const scheduleCallStatusOptions = [
  { id: 'unsigned', value: '未点名' },
  { id: 'signed', value: '已点名' },
]

function normalizeOptionalClassroomId(value) {
  const text = String(value ?? '').trim()
  if (!text || text === '0' || text.toLowerCase() === 'null' || text.toLowerCase() === 'undefined')
    return ''
  return text
}

function normalizeScheduleFilterValue(value) {
  if (Array.isArray(value))
    return value.length ? value[0] : undefined
  const text = String(value ?? '').trim()
  return text || undefined
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

function hasScheduledLesson(lesson) {
  if (!lesson || typeof lesson !== 'object')
    return false
  if (String(lesson.scheduleId || '').trim())
    return true
  if (String(lesson.classId || '').trim())
    return true
  if (Array.isArray(lesson.studentId) ? lesson.studentId.length > 0 : Boolean(lesson.studentId))
    return true
  return Array.isArray(lesson.studentNames) && lesson.studentNames.length > 0
}

function syncAllFilterScheduleTeacher(values) {
  allFilterRef.value?.setScheduleTeacherFilter?.(values, false)
  filterTeacherId.value = normalizeScheduleFilterValues(values)
}

function syncAllFilterScheduleClassroom(values) {
  allFilterRef.value?.setScheduleClassroomFilter?.(values, false)
  filterClassroomId.value = normalizeScheduleFilterValues(values)
}

function syncAllFilterScheduleClass(value) {
  allFilterRef.value?.setScheduleClassFilter?.(value, false)
  filterClassId.value = normalizeScheduleFilterValue(value)
}

function syncAllFilterScheduleOneToOne(value) {
  allFilterRef.value?.setScheduleOneToOneFilter?.(value, false)
  filterOneToOneId.value = normalizeScheduleFilterValue(value)
}

function syncAllFilterScheduleCourse(value) {
  allFilterRef.value?.setScheduleCourseFilter?.(value, false)
  filterCourseId.value = normalizeScheduleFilterValue(value)
}

function syncAllFilterScheduleType(values) {
  allFilterRef.value?.setScheduleTypeFilter?.(values, false)
  filterScheduleType.value = normalizeScheduleFilterValues(values)
}

function syncAllFilterScheduleCallStatus(value) {
  allFilterRef.value?.setScheduleCallStatusFilter?.(value, false)
  filterCallStatus.value = normalizeScheduleFilterValue(value)
}

function syncAllFilterStuPhone(value) {
  allFilterRef.value?.setStuPhoneSearchFilter?.(value, false)
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

async function loadSchedulingClassrooms() {
  schedulingClassroomLoading.value = true
  try {
    const res = await listClassroomsApi({
      enabledOnly: true,
    })
    if (res.code !== 200) {
      messageService.error(res.message || '获取教室列表失败')
      return
    }
    schedulingClassroomList.value = Array.isArray(res.result) ? res.result : []
  }
  catch (error) {
    console.error('load scheduling classrooms failed', error)
    messageService.error(error?.response?.data?.message || error?.message || '获取教室列表失败')
  }
  finally {
    schedulingClassroomLoading.value = false
  }
}

function classroomNameById(value, fallbackName = '') {
  const normalized = normalizeOptionalClassroomId(value)
  if (!normalized)
    return String(fallbackName || '').trim()
  const matched = schedulingClassroomList.value.find(item => String(item?.id ?? '').trim() === normalized)
  return String(matched?.name || fallbackName || '').trim()
}

const selectedOneToOneSchedulingClassroomName = computed(() =>
  classroomNameById(selectedOneToOneClassroomId.value),
)

const oneToOneSchedulingClassroomOptions = computed(() => {
  const optionMap = new Map()
  const append = (id, name) => {
    const normalizedId = normalizeOptionalClassroomId(id)
    const label = String(name || '').trim()
    if (!normalizedId || !label || optionMap.has(normalizedId))
      return
    optionMap.set(normalizedId, {
      value: normalizedId,
      label,
    })
  }

  schedulingClassroomList.value.forEach(item => append(item?.id, item?.name))
  append(selectedOneToOneClassroomId.value, selectedOneToOneSchedulingClassroomName.value)

  return [...optionMap.values()]
})

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

function emitCurrentWeekRange(value = currentWeek.value) {
  const start = getWeekStart(value)
  emit('week-range-change', {
    startDate: start.format('YYYY-MM-DD'),
    endDate: start.add(6, 'day').format('YYYY-MM-DD'),
  })
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

watch(currentWeek, value => emitCurrentWeekRange(value), { immediate: true })

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

function parseAttachmentFilenameFromHeader(cd) {
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
  return quoted?.[1] || undefined
}

async function exportSmartTimetable() {
  exportLoading.value = true
  try {
    const { startDate, endDate } = queryDateRange.value
    const scheduleTeacherIds = filterTeacherId.value.join(',')
    const classroomIds = filterClassroomId.value.join(',')
    const res = await downloadSmartTimetableExcelApi({
      startDate,
      endDate,
      viewMode: currentTime.value,
      studentId: filterStudentId.value,
      scheduleTeacherIds: scheduleTeacherIds || undefined,
      classroomIds: classroomIds || undefined,
      groupClassIds: filterClassId.value ? String(filterClassId.value) : undefined,
      oneToOneClassIds: filterOneToOneId.value ? String(filterOneToOneId.value) : undefined,
      lessonIds: filterCourseId.value ? String(filterCourseId.value) : undefined,
      scheduleTypes: filterScheduleType.value.length ? filterScheduleType.value.join(',') : undefined,
      callStatuses: filterCallStatus.value ? String(filterCallStatus.value) : undefined,
      ...teacherMatrixGroupParamsForKey(currentGroup.value),
    })
    const ct = String(res.headers['content-type'] || '')
    if (ct.includes('application/json')) {
      const text = await res.data.text()
      try {
        const j = JSON.parse(text)
        messageService.error(j.message || '导出失败')
      }
      catch {
        messageService.error('导出失败')
      }
      return
    }
    const blob = new Blob([res.data], {
      type: ct || 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    })
    const cd = res.headers['content-disposition']
    const filename = parseAttachmentFilenameFromHeader(cd)
      || `智慧课表_${dayjs().format('YYYYMMDDHHmmss')}.xlsx`
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    a.click()
    URL.revokeObjectURL(url)
    messageService.success('已导出课表')
  }
  catch (error) {
    console.error('export smart timetable failed', error)
    messageService.error('导出失败')
  }
  finally {
    exportLoading.value = false
  }
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
const scheduleBatchPlanEditOpen = ref(false)
const currentBatchPlanSchedule = ref(null)
const forcingConflictSchedule = ref(false)
const locatingConflictItemKey = ref('')
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
  classroomName: '',
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
const dragCopySubmitting = ref(false)
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
let suppressScheduledLessonClickUntil = 0
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

const effectivePeriodConfigRaw = ref(null)

const periodConfig = computed(() => {
  const parsed = parseUnifiedTimePeriodConfig(effectivePeriodConfigRaw.value ?? userStore.instConfig?.unifiedTimePeriodJson)
  return parsed ?? DEFAULT_UNIFIED_TIME_PERIOD_CONFIG
})

const sortedPeriodGroups = computed(() => configGroupsSorted(periodConfig.value))

const groupOptions = computed(() => {
  const g = sortedPeriodGroups.value
  if (!g.length)
    return [{ key: 'A', label: '默认时段' }]
  return g.map((group, index) => {
    const key = periodGroupKeyForIndex(index)
    return {
      key,
      label: group.name || `${key}时段`,
    }
  })
})

function slotsForGroupKey(key) {
  const groups = sortedPeriodGroups.value
  const fallback = buildQuickHourlySlots().filter(s => s.enabled !== false)
  if (!groups.length)
    return [...fallback].sort((a, b) => a.index - b.index)
  const idx = periodGroupIndexForKey(key)
  const g = groups[idx] || groups[0]
  return [...g.slots].filter(s => s.enabled !== false).sort((a, b) => a.index - b.index)
}

const activePeriodSlots = computed(() => slotsForGroupKey(displayedGroupKey.value))

function periodGroupForKey(key) {
  const groups = sortedPeriodGroups.value
  if (!groups.length)
    return null
  const idx = periodGroupIndexForKey(key)
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
    lessonDate: '',
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
  const explicitCallStatus = Number(legacyItem?.callStatus ?? 0)
  if (explicitCallStatus === 2)
    return 'signed'
  if (explicitCallStatus === 1)
    return 'unsigned'
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
      lessonDate: String(leg.scheduleDate || '').trim(),
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
const lessonConflictRenderTick = ref(0)

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

const tableDataSource = computed(() => {
  lessonConflictRenderTick.value
  const rows = isSwapTimeGrid.value ? transposedDataSource.value : dataSource.value
  return [...rows]
})

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
      if (!hasScheduledLesson(lesson))
        clearLessonConflictState(lesson, scope)
    })
  })
  lessonConflictRenderTick.value += 1
}

function clearOneToOneAvailabilityHighlights() {
  cancelOneToOneAvailabilityCheck()
  resetEmptyLessonConflicts()
}

/** 当前展示范围内每位老师已占用的节次数（与格子里蓝色已排课一致：有已排日程即计入） */
const scheduledLessonCountByTeacher = computed(() => {
  const map = new Map()
  for (const row of dataSource.value) {
    const tid = String(row.teacherId)
    let n = 0
    for (const lesson of row.lessons || []) {
      if (hasScheduledLesson(lesson))
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

const visibleScheduledLessons = computed(() => {
  const lessons = []
  dataSource.value.forEach((row) => {
    ;(row.lessons || []).forEach((lesson) => {
      if (hasScheduledLesson(lesson))
        lessons.push(lesson)
    })
  })
  return lessons
})

const smartTimetableTotalSchedules = computed(() => visibleScheduledLessons.value.length)
const smartTimetableUnsignedSchedules = computed(() =>
  visibleScheduledLessons.value.filter(lesson => lesson.callStatusKey === 'unsigned').length,
)

const activeGroupLabel = computed(() => {
  return groupOptions.value.find(o => o.key === displayedGroupKey.value)?.label || ''
})

let detectOneToOneAvailabilityBridge = (_value) => {}
let handleClassBridge = (_value) => {}

const {
  assistantNameById,
  assistantOptions,
  assistantOptionsLoading,
  assistantTextForIds,
  buildOneToOneScheduleAssignment,
  fetchAssistantOptions,
  fetchOneToOneOptionsForTimetable,
  filterOneToOneOption,
  handle1v1,
  handleOneToOneDropdownVisibleChange,
  isAssistantAllowedInDisplayedGroup,
  normalizedSelectedAssistantIds,
  oneToOneData,
  oneToOneListLoading,
  oneToOnePickerOpen,
  oneToOneRecordId,
  renderOneToOneDropdown,
  resetOneToOnePickerState,
  selectedAssistantIds,
} = useSmartTimetablePicker({
  activeGroupLabel,
  classroomId: selectedOneToOneClassroomId,
  classroomLoading: schedulingClassroomLoading,
  classroomNameById: value => classroomNameById(value),
  classroomOptions: oneToOneSchedulingClassroomOptions,
  classroomPlaceholder: computed(() => '不选则不占用教室'),
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
  hasScheduledLesson,
  normalizedSelectedAssistantIds,
  normalizedSelectedClassroomId: computed(() => normalizeOptionalClassroomId(selectedOneToOneClassroomId.value)),
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
      currentGroup.value = opts[0]?.key || 'A'
  },
  { immediate: true },
)

async function loadTimetableMatrix() {
  const seq = ++matrixLoadSeq
  const requestedGroup = currentGroup.value
  clearClassConflictCache()
  timetableLoading.value = true
  try {
    const { startDate, endDate } = queryDateRange.value
    try {
      const cfgRes = await getInstPeriodConfigApi({ effectiveDate: startDate })
      if (seq !== matrixLoadSeq)
        return
      effectivePeriodConfigRaw.value = cfgRes.result?.unifiedTimePeriodJson ?? userStore.instConfig?.unifiedTimePeriodJson ?? null
    }
    catch (error) {
      console.warn('load effective inst period config failed, fallback to latest', error)
      if (!userStore.instConfig)
        await userStore.getInstConfig()
      if (seq !== matrixLoadSeq)
        return
      effectivePeriodConfigRaw.value = userStore.instConfig?.unifiedTimePeriodJson ?? null
    }
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
  [currentWeek, () => (currentTime.value === 'swapWeek' ? 'week' : currentTime.value), currentGroup],
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
  void loadClassOptions()
  void loadSchedulingClassrooms()
}

onMounted(() => {
  void loadTimetableMatrix()
  void fetchOneToOneOptionsForTimetable()
  void fetchAssistantOptions()
  void loadClassOptions()
  void loadSchedulingClassrooms()
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
const classPickerOpen = ref(false)
const selectedClassAssistantIds = ref([])
const classAssistantKeyword = ref('')
const classClassroomKeyword = ref('')
const autoClassTeacherFilterId = ref('')
let classSelectionSyncing = false
let preserveClassPickerOpen = false
let lastHandledClassId = ''

function normalizeClassAssistantIds(values) {
  return Array.from(
    new Set(
      (Array.isArray(values) ? values : [])
        .map(value => String(value || '').trim())
        .filter(Boolean),
    ),
  )
}

const normalizedSelectedClassAssistantIds = computed(() =>
  normalizeClassAssistantIds(selectedClassAssistantIds.value),
)

const {
  buildClassScheduleAssignment,
  classData,
  classConflictLoading,
  classListLoading,
  clearClassConflictCache,
  ensureClassLoaded,
  findClassInfo,
  handleClass,
  loadClassOptions,
  resolveClassConflictMessage,
  resolveSelectedClassTarget,
} = useSmartTimetableClassMode({
  activeGroupLabel,
  dataSource,
  getLessonIndex,
  hasScheduledLesson,
  normalizedSelectedAssistantIds: normalizedSelectedClassAssistantIds,
  queryDateRange,
  resetEmptyLessonConflicts,
  selectedClassroomId: computed(() => normalizeOptionalClassroomId(selectedClassClassroomId.value)),
  resolveClassroomName: value => classroomNameById(value),
})

const selectedClassSchedulingClassroomName = computed(() => {
  const selectedClass = findClassInfo(classId.value)
  const defaultClassroomId = normalizeOptionalClassroomId(selectedClass?.classroomId)
  const defaultClassroomName = String(selectedClass?.classroomName || '').trim()
  const explicitClassroomId = normalizeOptionalClassroomId(selectedClassClassroomId.value)
  if (!explicitClassroomId)
    return defaultClassroomName
  return classroomNameById(
    explicitClassroomId,
    explicitClassroomId === defaultClassroomId ? defaultClassroomName : '',
  )
})

const schedulingClassroomOptions = computed(() => {
  const optionMap = new Map()
  const append = (id, name) => {
    const normalizedId = normalizeOptionalClassroomId(id)
    const label = String(name || '').trim()
    if (!normalizedId || !label || optionMap.has(normalizedId))
      return
    optionMap.set(normalizedId, {
      value: normalizedId,
      label,
    })
  }

  schedulingClassroomList.value.forEach(item => append(item?.id, item?.name))
  const selectedClass = findClassInfo(classId.value)
  append(selectedClass?.classroomId, selectedClass?.classroomName)
  append(selectedOneToOneClassroomId.value, selectedOneToOneSchedulingClassroomName.value)
  append(selectedClassClassroomId.value, selectedClassSchedulingClassroomName.value)

  return [...optionMap.values()]
})

const schedulingClassroomPlaceholder = computed(() =>
  currentModel.value === '1'
    ? '不选则不占用教室'
    : (normalizeOptionalClassroomId(findClassInfo(classId.value)?.classroomId) ? '不选则默认班级教室' : '请选择教室'),
)

const classSchedulingClassroomSummary = computed(() => {
  if (selectedClassSchedulingClassroomName.value)
    return `已选教室：${selectedClassSchedulingClassroomName.value}`
  if (normalizeOptionalClassroomId(findClassInfo(classId.value)?.classroomId))
    return '未选择教室时，默认使用班级教室。'
  return '当前班级未配置默认教室，可在此选择。'
})

function resolveSelectedOneToOneClassroom() {
  const classroomId = normalizeOptionalClassroomId(selectedOneToOneClassroomId.value)
  return {
    classroomId,
    classroomName: classroomNameById(classroomId),
  }
}

function resolveSelectedClassroomForClass(classInfo) {
  const explicitClassroomId = normalizeOptionalClassroomId(selectedClassClassroomId.value)
  const fallbackClassroomId = normalizeOptionalClassroomId(classInfo?.classroomId)
  const classroomId = explicitClassroomId || fallbackClassroomId
  return {
    classroomId,
    classroomName: classroomNameById(
      classroomId,
      classroomId === fallbackClassroomId ? String(classInfo?.classroomName || '').trim() : '',
    ),
  }
}

handleClassBridge = (value) => {
  void handleClass(value)
}

const classAssistantOptionsInPicker = computed(() => {
  const keyword = String(classAssistantKeyword.value || '').trim().toLowerCase()
  return assistantOptions.value.filter((item) => {
    if (!isAssistantAllowedInDisplayedGroup(item.value))
      return false
    if (!keyword)
      return true
    const blob = `${item.label || ''} ${item.mobile || ''} ${item.value || ''}`.toLowerCase()
    return blob.includes(keyword)
  })
})

const classClassroomOptionsInPicker = computed(() => {
  const keyword = String(classClassroomKeyword.value || '').trim().toLowerCase()
  return schedulingClassroomOptions.value.filter((item) => {
    if (!keyword)
      return true
    return String(item?.label || '').trim().toLowerCase().includes(keyword)
  })
})

watch(
  [displayedGroupKey, assistantOptions],
  () => {
    if (!normalizedSelectedClassAssistantIds.value.length)
      return
    const next = normalizedSelectedClassAssistantIds.value.filter(id => isAssistantAllowedInDisplayedGroup(id))
    if (next.length === normalizedSelectedClassAssistantIds.value.length)
      return
    const removedCount = normalizedSelectedClassAssistantIds.value.length - next.length
    handleClassAssistantSelectChange(next)
    if (removedCount > 0) {
      messageService.warning(`已切换到${activeGroupLabel.value || '当前组'}，自动移除 ${removedCount} 位非本组助教`, { duration: 4500 })
    }
  },
  { immediate: true },
)

function classAssistantTextForTeacher(teacherId) {
  const normalizedTeacherId = String(teacherId || '').trim()
  return classAssistantTextForIds(
    normalizedSelectedClassAssistantIds.value.filter(id => id !== normalizedTeacherId),
  )
}

function classAssistantTextForIds(assistantIds) {
  const names = (Array.isArray(assistantIds) ? assistantIds : [])
    .map(id => assistantNameById(id))
    .filter(Boolean)
  return names.length ? names.join('、') : '未安排'
}

const selectedClassAssistantText = computed(() => classAssistantTextForTeacher(''))

function requestKeepClassPickerOpen() {
  preserveClassPickerOpen = true
  classPickerOpen.value = true
  requestAnimationFrame(() => {
    preserveClassPickerOpen = false
  })
}

function handleClassDropdownVisibleChange(open) {
  if (!open && preserveClassPickerOpen) {
    classPickerOpen.value = true
    return
  }
  classPickerOpen.value = open
}

function handleClassAssistantSelectChange(value) {
  selectedClassAssistantIds.value = normalizeClassAssistantIds(value)
  if (currentModel.value === '2' && classId.value)
    handleClassBridge(classId.value)
  else
    resetEmptyLessonConflicts('assistant')
}

function toggleClassAssistantOption(value, checked) {
  const normalized = String(value || '').trim()
  if (!normalized)
    return
  const next = new Set(normalizedSelectedClassAssistantIds.value)
  if (checked)
    next.add(normalized)
  else
    next.delete(normalized)
  handleClassAssistantSelectChange([...next])
  requestKeepClassPickerOpen()
}

function renderClassDropdown() {
  const selectedClass = findClassInfo(classId.value)
  const classOptionNodes = classData.value.length
    ? classData.value.map((item) => {
        const selected = String(classId.value || '').trim() === String(item.id || '').trim()
        return h('div', {
          class: 'st-top-1v1-dropdown__class-option',
          key: item.id,
          style: {
            padding: '12px 16px',
            borderBottom: '1px solid #f5f5f5',
            background: selected ? '#e6f4ff' : '#fff',
            cursor: 'pointer',
            transition: 'background 0.2s ease',
          },
          onMousedown: (event) => {
            event.preventDefault()
            event.stopPropagation()
          },
          onClick: async () => {
            const nextClassId = String(item.id || '').trim()
            classId.value = nextClassId || null
            await handleClassSelectionChange(nextClassId)
            requestKeepClassPickerOpen()
          },
        }, [
          h('div', {
            style: {
              color: '#262626',
              fontSize: '14px',
              fontWeight: 700,
              lineHeight: '22px',
              marginBottom: '4px',
            },
          }, item.name || '-'),
          h('div', {
            style: {
              color: '#666',
              fontSize: '12px',
              lineHeight: '18px',
            },
          }, `主教：${item.mainTeacherName || '-'}`),
        ])
      })
    : [
        h('div', {
          class: 'st-top-1v1-dropdown__empty',
          style: {
            padding: '16px',
            color: '#8c8c8c',
            fontSize: '12px',
            lineHeight: '18px',
          },
        }, classListLoading.value ? '班级加载中...' : '暂无匹配班级'),
      ]
  const classroomOptionNodes = classClassroomOptionsInPicker.value.length
    ? classClassroomOptionsInPicker.value.map((item) => {
        const checked = normalizeOptionalClassroomId(selectedClassClassroomId.value) === String(item.value || '').trim()
        return h('div', {
          class: 'st-top-1v1-dropdown__assistant-item',
          key: item.value,
          style: {
            display: 'flex',
            alignItems: 'center',
            gap: '6px',
            minHeight: '30px',
            padding: '2px 0px',
            borderRadius: '10px',
            cursor: 'pointer',
            boxSizing: 'border-box',
            userSelect: 'none',
          },
          onMousedown: (event) => {
            event.preventDefault()
            event.stopPropagation()
          },
          onClick: () => {
            selectedClassClassroomId.value = String(item.value || '').trim() || undefined
            requestKeepClassPickerOpen()
          },
        }, [
          h('span', {
            class: 'st-top-1v1-dropdown__assistant-checkbox',
            style: {
              display: 'inline-flex',
              alignItems: 'center',
              justifyContent: 'center',
              width: '16px',
              height: '16px',
              borderRadius: '4px',
              border: checked ? '1px solid #1677ff' : '1px solid #8c8c8c',
              background: checked ? '#1677ff' : '#fff',
              color: '#fff',
              flex: '0 0 auto',
              fontSize: '11px',
              fontWeight: 700,
              lineHeight: 1,
            },
          }, checked ? '✓' : ''),
          h('span', {
            class: 'st-top-1v1-dropdown__assistant-name',
            style: {
              flex: 1,
              minWidth: 0,
              color: '#262626',
              fontSize: '12px',
              fontWeight: 600,
              lineHeight: '20px',
            },
          }, item.label),
        ])
      })
    : [
        h('div', {
          class: 'st-top-1v1-dropdown__empty',
          style: {
            padding: '8px 0 12px',
            color: '#8c8c8c',
            fontSize: '12px',
            lineHeight: '18px',
          },
        }, '暂无匹配教室'),
      ]
  const assistantChildren = [
    h('div', {
      class: 'st-top-1v1-dropdown__section-head',
      style: {
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        gap: '12px',
        marginBottom: '12px',
      },
    }, [
      h('span', {
        class: 'st-top-1v1-dropdown__section-title',
        style: {
          color: '#262626',
          fontSize: '14px',
          fontWeight: 700,
          lineHeight: 1,
        },
      }, '选择助教'),
      h('span', {
        class: 'st-top-1v1-dropdown__section-hint',
        style: {
          color: '#8c8c8c',
          fontSize: '12px',
          lineHeight: 1,
        },
      }, classId.value ? '多选，可不选' : '先选班级后配置'),
    ]),
  ]
  const classroomChildren = [
    h('div', {
      class: 'st-top-1v1-dropdown__section-head',
      style: {
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        gap: '12px',
        marginBottom: '10px',
      },
    }, [
      h('span', {
        class: 'st-top-1v1-dropdown__section-title',
        style: {
          color: '#262626',
          fontSize: '14px',
          fontWeight: 700,
          lineHeight: 1,
        },
      }, '选择教室'),
      h('span', {
        class: 'st-top-1v1-dropdown__section-hint',
        style: {
          color: '#8c8c8c',
          fontSize: '12px',
          lineHeight: 1,
        },
      }, '单选，可不选'),
    ]),
    h('input', {
      class: 'st-top-1v1-dropdown__search-input',
      value: classClassroomKeyword.value,
      placeholder: '搜索教室',
      style: {
        width: '100%',
        height: '30px',
        padding: '0 10px',
        color: '#262626',
        fontSize: '12px',
        background: '#fff',
        border: '1px solid #d9d9d9',
        borderRadius: '8px',
        outline: 'none',
        boxSizing: 'border-box',
        marginBottom: '4px',
      },
      onInput: (event) => {
        classClassroomKeyword.value = event?.target?.value || ''
      },
      onFocus: () => {
        requestKeepClassPickerOpen()
      },
      onClick: () => {
        requestKeepClassPickerOpen()
      },
    }),
    h('div', {
      class: 'st-top-1v1-dropdown__summary',
      style: {
        marginBottom: '14px',
        color: '#5b6475',
        fontSize: '12px',
        lineHeight: '1.5',
      },
    }, classSchedulingClassroomSummary.value),
    schedulingClassroomLoading.value
      ? h('div', {
          class: 'st-top-1v1-dropdown__empty',
          style: {
            padding: '8px 0 12px',
            color: '#8c8c8c',
            fontSize: '12px',
            lineHeight: '18px',
            marginBottom: '12px',
          },
        }, '教室加载中...')
      : h('div', {
          class: 'st-top-1v1-dropdown__assistant-list',
          style: {
            display: 'flex',
            flexDirection: 'column',
            gap: '0px',
            maxHeight: '200px',
            overflowY: 'auto',
            paddingRight: '4px',
            marginBottom: '14px',
          },
        }, [
          h('div', {
            class: 'st-top-1v1-dropdown__assistant-item',
            style: {
              display: 'flex',
              alignItems: 'center',
              gap: '6px',
              minHeight: '30px',
              padding: '2px 0px',
              borderRadius: '10px',
              cursor: 'pointer',
              boxSizing: 'border-box',
              userSelect: 'none',
            },
            onMousedown: (event) => {
              event.preventDefault()
              event.stopPropagation()
            },
            onClick: () => {
              selectedClassClassroomId.value = undefined
              requestKeepClassPickerOpen()
            },
          }, [
            h('span', {
              class: 'st-top-1v1-dropdown__assistant-checkbox',
              style: {
                display: 'inline-flex',
                alignItems: 'center',
                justifyContent: 'center',
                width: '16px',
                height: '16px',
                borderRadius: '4px',
                border: !normalizeOptionalClassroomId(selectedClassClassroomId.value) ? '1px solid #1677ff' : '1px solid #8c8c8c',
                background: !normalizeOptionalClassroomId(selectedClassClassroomId.value) ? '#1677ff' : '#fff',
                color: '#fff',
                flex: '0 0 auto',
                fontSize: '11px',
                fontWeight: 700,
                lineHeight: 1,
              },
            }, !normalizeOptionalClassroomId(selectedClassClassroomId.value) ? '✓' : ''),
            h('span', {
              class: 'st-top-1v1-dropdown__assistant-name',
              style: {
                flex: 1,
                minWidth: 0,
                color: '#262626',
                fontSize: '12px',
                fontWeight: 600,
                lineHeight: '20px',
              },
            }, schedulingClassroomPlaceholder.value),
          ]),
          ...classroomOptionNodes,
        ]),
  ]

  if (classId.value) {
    if (selectedClass?.mainTeacherName) {
      assistantChildren.push(
        h('div', {
          class: 'st-top-1v1-dropdown__summary',
          style: {
            marginBottom: '8px',
            color: '#5b6475',
            fontSize: '12px',
            lineHeight: '1.5',
          },
        }, `默认主教筛选：${selectedClass.mainTeacherName}`),
      )
    }

    assistantChildren.push(
      h('input', {
        class: 'st-top-1v1-dropdown__search-input',
        value: classAssistantKeyword.value,
        placeholder: '搜索助教',
        style: {
          width: '100%',
          height: '30px',
          padding: '0 10px',
          color: '#262626',
          fontSize: '12px',
          background: '#fff',
          border: '1px solid #d9d9d9',
          borderRadius: '8px',
          outline: 'none',
          boxSizing: 'border-box',
          marginBottom: '4px',
        },
        onInput: (event) => {
          classAssistantKeyword.value = event?.target?.value || ''
        },
        onFocus: () => {
          requestKeepClassPickerOpen()
        },
        onClick: () => {
          requestKeepClassPickerOpen()
        },
      }),
    )

    if (normalizedSelectedClassAssistantIds.value.length) {
      assistantChildren.push(
        h('div', {
          class: 'st-top-1v1-dropdown__summary',
          style: {
            marginBottom: '2px',
            color: '#5b6475',
            fontSize: '12px',
            lineHeight: '1.5',
          },
        }, `已选助教：${selectedClassAssistantText.value}`),
      )
    }

    if (assistantOptionsLoading.value) {
      assistantChildren.push(h('div', {
        class: 'st-top-1v1-dropdown__empty',
        style: {
          padding: '14px 0 4px',
          color: '#8c8c8c',
          fontSize: '12px',
          lineHeight: '18px',
        },
      }, '助教加载中...'))
    }
    else if (classAssistantOptionsInPicker.value.length) {
      assistantChildren.push(
        h('div', {
          class: 'st-top-1v1-dropdown__assistant-list',
          style: {
            display: 'flex',
            flexDirection: 'column',
            gap: '0px',
            flex: 1,
            minHeight: 0,
            overflowY: 'auto',
            paddingRight: '4px',
          },
        }, classAssistantOptionsInPicker.value.map((item) => {
          const checked = normalizedSelectedClassAssistantIds.value.includes(String(item.value))
          return h('div', {
            class: 'st-top-1v1-dropdown__assistant-item',
            key: item.value,
            style: {
              display: 'flex',
              alignItems: 'center',
              gap: '6px',
              minHeight: '30px',
              padding: '2px 0px',
              borderRadius: '10px',
              cursor: 'pointer',
              boxSizing: 'border-box',
              userSelect: 'none',
            },
            onMousedown: (event) => {
              event.preventDefault()
              event.stopPropagation()
            },
            onClick: () => {
              toggleClassAssistantOption(item.value, !checked)
            },
          }, [
            h('span', {
              class: 'st-top-1v1-dropdown__assistant-checkbox',
              style: {
                display: 'inline-flex',
                alignItems: 'center',
                justifyContent: 'center',
                width: '16px',
                height: '16px',
                borderRadius: '4px',
                border: checked ? '1px solid #1677ff' : '1px solid #8c8c8c',
                background: checked ? '#1677ff' : '#fff',
                color: '#fff',
                flex: '0 0 auto',
                fontSize: '11px',
                fontWeight: 700,
                lineHeight: 1,
              },
            }, checked ? '✓' : ''),
            h('span', {
              class: 'st-top-1v1-dropdown__assistant-name',
              style: {
                flex: 1,
                minWidth: 0,
                color: '#262626',
                fontSize: '12px',
                fontWeight: 600,
                lineHeight: '20px',
              },
            }, item.label),
            item.mobile
              ? h('span', {
                  class: 'st-top-1v1-dropdown__assistant-mobile',
                  style: {
                    color: '#8c8c8c',
                    fontSize: '11px',
                    lineHeight: '20px',
                    flex: '0 0 auto',
                  },
                }, item.mobile)
              : null,
          ])
        })),
      )
    }
    else {
      assistantChildren.push(h('div', {
        class: 'st-top-1v1-dropdown__empty',
        style: {
          padding: '14px 0 4px',
          color: '#8c8c8c',
          fontSize: '12px',
          lineHeight: '18px',
        },
      }, '暂无匹配助教'))
    }
  }
  else {
    assistantChildren.push(h('div', {
      class: 'st-top-1v1-dropdown__empty',
      style: {
        padding: '14px 0 18px',
        color: '#8c8c8c',
        fontSize: '12px',
        lineHeight: '18px',
      },
    }, '先选班级，再在右侧勾选助教。'))
  }

  return h('div', {
    class: 'st-top-1v1-dropdown',
    style: {
      display: 'flex',
      width: '820px',
      minWidth: '820px',
      maxWidth: '820px',
      minHeight: '300px',
      maxHeight: '300px',
      background: '#fff',
      borderRadius: '12px',
      overflow: 'hidden',
    },
  }, [
    h('div', {
      class: 'st-top-1v1-dropdown__list',
      style: {
        flex: '0 0 278px',
        minWidth: '278px',
        maxWidth: '278px',
        overflowY: 'auto',
        borderRight: '1px solid #f0f0f0',
      },
    }, classOptionNodes),
    h('div', {
      class: 'st-top-1v1-dropdown__side',
      style: {
        display: 'flex',
        flex: '0 0 271px',
        flexDirection: 'column',
        minHeight: 0,
        minWidth: 0,
        padding: '14px 16px 16px',
        background: 'linear-gradient(180deg, #fcfdff 0%, #fff 100%)',
        overflow: 'hidden',
        borderRight: '1px solid #f0f0f0',
      },
      onMousedown: event => event.stopPropagation(),
    }, assistantChildren),
    h('div', {
      class: 'st-top-1v1-dropdown__side',
      style: {
        display: 'flex',
        flex: 1,
        flexDirection: 'column',
        minWidth: 0,
        padding: '14px 16px 16px',
        background: 'linear-gradient(180deg, #fcfdff 0%, #fff 100%)',
        overflowY: 'auto',
      },
      onMousedown: event => event.stopPropagation(),
    }, classroomChildren),
  ])
}

function applyClassDefaultTeacherQuickFilter(classInfo) {
  const defaultTeacherId = String(classInfo?.mainTeacherId || '').trim()
  const defaultTeacherName = String(classInfo?.mainTeacherName || defaultTeacherId).trim()
  const previousAutoTeacherId = String(autoClassTeacherFilterId.value || '').trim()
  autoClassTeacherFilterId.value = defaultTeacherId
  if (!defaultTeacherId) {
    const shouldClear = Boolean(previousAutoTeacherId && filterTeacherId.value.length === 1 && filterTeacherId.value[0] === previousAutoTeacherId)
    if (previousAutoTeacherId && filterTeacherId.value.length === 1 && filterTeacherId.value[0] === previousAutoTeacherId) {
      nextTick(() => {
        if (allFilterRef.value?.setScheduleTeacherFilter) {
          allFilterRef.value.setScheduleTeacherFilter([], true)
        }
        else {
          handleScheduleTeacherFilter([])
        }
      })
    }
    return shouldClear
  }
  const shouldUpdate = filterTeacherId.value.length !== 1 || filterTeacherId.value[0] !== defaultTeacherId
  scheduleTeacherOptions.value = mergeFilterOptions(scheduleTeacherOptions.value, [{
    id: defaultTeacherId,
    value: defaultTeacherName || defaultTeacherId,
  }], [defaultTeacherId])
  nextTick(() => {
    if (allFilterRef.value?.setScheduleTeacherFilter) {
      allFilterRef.value.setScheduleTeacherFilter([defaultTeacherId], true)
    }
    else {
      handleScheduleTeacherFilter([defaultTeacherId])
    }
  })
  return shouldUpdate
}

function clearClassAutoTeacherFilter() {
  const autoTeacherId = String(autoClassTeacherFilterId.value || '').trim()
  autoClassTeacherFilterId.value = ''
  if (!autoTeacherId)
    return
  if (filterTeacherId.value.length !== 1 || filterTeacherId.value[0] !== autoTeacherId)
    return
  nextTick(() => {
    if (allFilterRef.value?.setScheduleTeacherFilter) {
      allFilterRef.value.setScheduleTeacherFilter([], true)
    }
    else {
      handleScheduleTeacherFilter([])
    }
  })
}

async function handleClassSelectionChange(value) {
  const nextClassId = String(value || '').trim()
  if (nextClassId !== lastHandledClassId) {
    selectedClassAssistantIds.value = []
    classAssistantKeyword.value = ''
    classClassroomKeyword.value = ''
    selectedClassClassroomId.value = undefined
  }
  lastHandledClassId = nextClassId
  if (!nextClassId) {
    preserveClassPickerOpen = false
    classPickerOpen.value = false
    clearClassAutoTeacherFilter()
    selectedClassClassroomId.value = undefined
    await handleClass(value)
    return
  }
  requestKeepClassPickerOpen()
  classSelectionSyncing = true
  try {
    const classInfo = await ensureClassLoaded(value)
    if (classInfo) {
      selectedClassClassroomId.value = normalizeOptionalClassroomId(classInfo.classroomId) || undefined
      const teacherFilterChanged = applyClassDefaultTeacherQuickFilter(classInfo)
      if (!teacherFilterChanged)
        await handleClass(value)
    }
    else {
      await handleClass(value)
    }
  }
  finally {
    nextTick(() => {
      classSelectionSyncing = false
    })
  }
}

function buildAvailabilitySlotKey(teacherId, lessonDate, startTime, endTime) {
  return `${String(teacherId)}|${lessonDate}|${startTime}|${endTime}`
}

function parseConflictTimeRange(timeText) {
  const m = String(timeText || '').trim().match(/(\d{2}:\d{2})\s*[~-]\s*(\d{2}:\d{2})/)
  if (!m)
    return null
  return { startTime: m[1], endTime: m[2] }
}

function findBestSlotForTimeRange(slots, timeRange) {
  if (!timeRange)
    return null
  const exact = (Array.isArray(slots) ? slots : []).find(slot =>
    slot.start === timeRange.startTime && slot.end === timeRange.endTime,
  )
  if (exact)
    return exact

  const targetStart = minutesFromHHMM(timeRange.startTime)
  const targetEnd = minutesFromHHMM(timeRange.endTime)
  if (targetStart == null || targetEnd == null)
    return null

  let bestSlot = null
  let bestOverlap = 0
  ;(Array.isArray(slots) ? slots : []).forEach((slot) => {
    const slotStart = minutesFromHHMM(slot?.start)
    const slotEnd = minutesFromHHMM(slot?.end)
    if (slotStart == null || slotEnd == null)
      return
    const overlap = Math.min(slotEnd, targetEnd) - Math.max(slotStart, targetStart)
    if (overlap > bestOverlap) {
      bestOverlap = overlap
      bestSlot = slot
    }
  })
  return bestOverlap > 0 ? bestSlot : null
}

function resolveConflictScheduleGroupInfo(item) {
  const teacherId = String(item?.teacherId || '').trim()
  const timeRange = parseConflictTimeRange(item?.timeText)
  const matches = groupOptions.value
    .map((opt) => {
      const group = periodGroupForKey(opt.key)
      const teacherMatched = !teacherId || !(group?.boundTeachers?.length)
        || group.boundTeachers.some(t => String(t.id) === teacherId)
      const timeMatched = !timeRange || Boolean(findBestSlotForTimeRange(slotsForGroupKey(opt.key), timeRange))
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

function conflictJumpActionKey(item) {
  return String(item?.key || [
    String(item?.teacherId || '').trim(),
    String(item?.teacherName || '').trim(),
    String(item?.date || '').trim(),
    String(item?.timeText || '').trim(),
  ].join('|'))
}

function resolveConflictScheduleTimeLabel(item) {
  const timeRange = parseConflictTimeRange(item?.timeText)
  const groupInfo = resolveConflictScheduleGroupInfo(item)
  const groupKey = groupInfo.keys.includes(currentGroup.value)
    ? currentGroup.value
    : (groupInfo.keys[0] || '')
  const matchedSlot = findBestSlotForTimeRange(
    slotsForGroupKey(groupKey || currentGroup.value),
    timeRange,
  )
  if (!matchedSlot)
    return ''
  const groupLabel = groupOptions.value.find(opt => opt.key === (groupKey || currentGroup.value))?.label
    || groupInfo.labels[0]
    || activeGroupLabel.value
    || '当前组'
  return `${groupLabel} 第${matchedSlot.index}节`
}

async function openConflictLocatingState(key) {
  locatingConflictItemKey.value = String(key || '').trim()
  await nextTick()
  await new Promise((resolve) => {
    if (typeof window !== 'undefined' && typeof window.requestAnimationFrame === 'function')
      window.requestAnimationFrame(() => resolve(true))
    else
      resolve(true)
  })
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

function closeConflictModalsByFlags(flags = {}) {
  if (flags.closeScheduledConflictModal)
    scheduledConflictDetailOpen.value = false
  if (flags.closeConflictDetailModal)
    conflictDetailModalOpen.value = false
  if (flags.closeDragConflictModal)
    dragConflictDetailOpen.value = false
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
    locatingConflictItemKey.value = ''
    closeConflictModalsByFlags(pending)
    messageService.success('已定位到冲突课程')
    return
  }

  if (pending.allowAppendTeacherFilter && !pending.teacherFilterExpanded) {
    const teacherId = String(pending.teacherId || '').trim()
    const teacherName = String(pending.teacherName || '').trim()
    if (teacherId && !filterTeacherId.value.includes(teacherId)) {
      const nextTeacherIds = [...filterTeacherId.value, teacherId]
      scheduleTeacherOptions.value = mergeFilterOptions(scheduleTeacherOptions.value, [{
        id: teacherId,
        value: teacherName || teacherId,
      }], nextTeacherIds)
      syncAllFilterScheduleTeacher(nextTeacherIds)
      pendingConflictJump = {
        ...pending,
        teacherFilterExpanded: true,
      }
      return
    }
  }

  if (!pending.filtersRelaxed) {
    const clearedFilterLabels = []
    if (filterStudentId.value) {
      syncAllFilterStuPhone(undefined)
      clearedFilterLabels.push('学员/电话')
    }
    if (filterClassroomId.value.length) {
      syncAllFilterScheduleClassroom([])
      clearedFilterLabels.push('上课教室')
    }
    if (filterClassId.value) {
      syncAllFilterScheduleClass(undefined)
      clearedFilterLabels.push('班级')
    }
    if (filterOneToOneId.value) {
      syncAllFilterScheduleOneToOne(undefined)
      clearedFilterLabels.push('1对1')
    }
    if (filterCourseId.value) {
      syncAllFilterScheduleCourse(undefined)
      clearedFilterLabels.push('课程')
    }
    if (filterScheduleType.value.length) {
      syncAllFilterScheduleType([])
      clearedFilterLabels.push('日程类型')
    }
    if (filterCallStatus.value) {
      syncAllFilterScheduleCallStatus(undefined)
      clearedFilterLabels.push('点名状态')
    }

    if (clearedFilterLabels.length) {
      pendingConflictJump = {
        ...pending,
        filtersRelaxed: true,
        clearedFilterLabels,
      }
      return
    }
  }

  pendingConflictJump = null
  locatingConflictItemKey.value = ''
  const adjustedLabels = []
  if (pending.teacherFilterExpanded)
    adjustedLabels.push('目标老师')
  if (Array.isArray(pending.clearedFilterLabels) && pending.clearedFilterLabels.length)
    adjustedLabels.push(...pending.clearedFilterLabels)
  const adjustedText = adjustedLabels.length
    ? `已自动调整${adjustedLabels.join('、')}筛选，`
    : ''
  messageService.warning(`${adjustedText}仍未定位到课程，请检查日期或组别`)
}

function buildConflictJumpItem(item, index = 0, fallbackConflictTypes = ['时间']) {
  const groupInfo = resolveConflictScheduleGroupInfo(item)
  const timeRange = parseConflictTimeRange(item?.timeText)
  const conflictTypes = uniqueConflictTypes([
    ...(Array.isArray(item?.conflictTypes) ? item.conflictTypes : []),
    ...(Array.isArray(fallbackConflictTypes) ? fallbackConflictTypes : []),
  ])
  const jumpGroupKey = groupInfo.keys.includes(currentGroup.value)
    ? currentGroup.value
    : groupInfo.keys[0] || ''
  const matchedSlot = findBestSlotForTimeRange(
    slotsForGroupKey(jumpGroupKey || currentGroup.value),
    timeRange,
  )
  return {
    key: `${item?.teacherId || item?.teacherName || 'teacher'}-${item?.date}-${item?.timeText}-${index}`,
    name: item?.name || '-',
    classTypeText: item?.classTypeText || '日程',
    date: item?.date,
    week: item?.week || '',
    timeText: item?.timeText,
    teacherId: item?.teacherId || '',
    teacherName: item?.teacherName || '-',
    groupLabel: groupInfo.labels.join('/') || '未知组别',
    classroomName: item?.classroomName || '-',
    assistantText: (item?.assistantNames || []).join('、') || '-',
    studentText: (item?.studentNames || []).join('、') || '-',
    conflictTypes,
    hasTeacherConflict: conflictTypes.includes('老师'),
    hasAssistantConflict: conflictTypes.includes('助教'),
    hasStudentConflict: conflictTypes.includes('学员'),
    hasClassroomConflict: conflictTypes.includes('教室'),
    jumpCellKey: matchedSlot && item?.teacherId
      ? buildAvailabilitySlotKey(item.teacherId, item.date, matchedSlot.start, matchedSlot.end)
      : '',
    jumpGroupKey,
  }
}

function buildConflictDetailItems(reason, fallbackConflictTypes = ['时间']) {
  const existingSchedules = Array.isArray(reason?.existingSchedules) ? reason.existingSchedules : []
  return existingSchedules.map((item, index) => buildConflictJumpItem(item, index, fallbackConflictTypes))
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
      conflictTypes,
      forceAllowed: Boolean(dragState?.scheduleId)
        && Boolean(dragState?.oneToOneId)
        && conflictTypes.length > 0
        && conflictTypes.every(type => type === '学员'),
      forceDisabledReason: buildForceScheduleDisabledReason(conflictTypes, '调课'),
      forcePayload: Boolean(dragState?.scheduleId)
        && Boolean(dragState?.oneToOneId)
        && conflictTypes.length > 0
        && conflictTypes.every(type => type === '学员')
        ? {
            ids: [String(dragState.scheduleId)],
            teacherId: String(target?.teacherId || '').trim(),
            assistantIds: Array.isArray(dragState?.assistantIds) ? dragState.assistantIds : [],
            classroomId: String(dragState?.classroomId || '').trim() || undefined,
            lessonDate: String(target?.lessonDate || '').trim(),
            startTime: String(target?.startTime || '').trim(),
            endTime: String(target?.endTime || '').trim(),
          }
        : null,
    },
    items,
  }
  dragConflictDetailOpen.value = true
}

function buildForceScheduleDisabledReason(conflictTypes = [], actionText = '排课') {
  const normalized = uniqueConflictTypes(conflictTypes)
  if (!normalized.length)
    return `当前冲突类型暂不支持仍要${actionText}。`
  if (normalized.every(type => type === '学员'))
    return ''

  const blockingTypes = normalized.filter(type => type !== '学员')
  if (!blockingTypes.length)
    return `当前冲突类型暂不支持仍要${actionText}。`

  return `当前还存在${blockingTypes.join('、')}冲突，仅纯学员冲突支持仍要${actionText}。`
}

function openApiConflictModal(reason, column, record) {
  const selectedTarget = resolveConflictAttemptTarget()
  const attemptedConflictTypes = Array.isArray(reason?.conflictTypes) ? reason.conflictTypes : []
  const selectedClassroom = resolveSelectedOneToOneClassroom()
  const assignment = buildOneToOneScheduleAssignment(
    record.teacherId,
    normalizedSelectedAssistantIds.value,
  )
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
    assistantText: assistantTextForIds(assignment.assistantIds),
    classroomId: selectedClassroom.classroomId,
    classroomName: selectedClassroom.classroomName,
    warningText: assignment.removedAssistantIds.length > 0 ? '主教与助教不能为同一人，系统已自动忽略重复助教。' : '',
    lessonIndex: getLessonIndex(column.startTime),
    groupLabel: activeGroupLabel.value || '当前组',
    conflictTypes: attemptedConflictTypes,
    removedAssistantIds: assignment.removedAssistantIds,
    forceAllowed,
    forceDisabledReason: forceAllowed ? '' : buildForceScheduleDisabledReason(attemptedConflictTypes),
    forcePayload: forceAllowed
      ? {
          oneToOneId: String(oneToOneRecordId.value),
          teacherId: String(record.teacherId),
          assistantIds: assignment.assistantIds,
          classroomId: selectedClassroom.classroomId || undefined,
          schedules: [{
            lessonDate: record.date,
            startTime: column.startTime,
            endTime: column.endTime,
            assistantIds: assignment.assistantIds,
            classroomId: selectedClassroom.classroomId || undefined,
          }],
        }
      : null,
  })
}

function openClassConflictModal(reason, column, record) {
  const selectedTarget = resolveConflictAttemptTarget()
  const attemptedConflictTypes = Array.isArray(reason?.conflictTypes) ? reason.conflictTypes : []
  const selectedClass = findClassInfo(classId.value)
  const hasStudentConflict = attemptedConflictTypes.includes('学员')
  const classStudentText = Array.isArray(selectedClass?.studentNames) && selectedClass.studentNames.length
    ? selectedClass.studentNames.join('、')
    : '暂无班级学员信息'
  const conflictingStudentText = Array.isArray(reason?.conflictingStudentNames) && reason.conflictingStudentNames.length
    ? reason.conflictingStudentNames.join('、')
    : (hasStudentConflict ? '未识别到具体冲突学员' : '')
  const selectedClassroom = resolveSelectedClassroomForClass(selectedClass)
  const assignment = selectedClass
    ? buildClassScheduleAssignment(
        selectedClass,
        record.teacherId,
        normalizedSelectedClassAssistantIds.value,
      )
    : {
        teacherId: String(record.teacherId || '').trim(),
        assistantIds: [],
        removedAssistantIds: [],
      }
  const forceAllowed = currentModel.value === '2'
    && Boolean(classId.value)
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
    assistantText: classAssistantTextForIds(assignment.assistantIds),
    studentLabel: '班级学员',
    studentText: classStudentText,
    conflictStudentLabel: hasStudentConflict ? '冲突学员' : '',
    conflictStudentText: conflictingStudentText,
    classroomId: selectedClassroom.classroomId,
    classroomName: selectedClassroom.classroomName,
    warningText: assignment.removedAssistantIds.length > 0 ? '主教与助教不能为同一人，系统已自动忽略重复助教。' : '',
    lessonIndex: getLessonIndex(column.startTime),
    groupLabel: activeGroupLabel.value || '当前组',
    conflictTypes: attemptedConflictTypes,
    removedAssistantIds: assignment.removedAssistantIds,
    forceAllowed,
    forceDisabledReason: forceAllowed ? '' : buildForceScheduleDisabledReason(attemptedConflictTypes),
    forcePayload: forceAllowed
      ? {
          groupClassId: String(classId.value),
          teacherId: assignment.teacherId,
          assistantIds: assignment.assistantIds,
          classroomId: selectedClassroom.classroomId || undefined,
          schedules: [{
            lessonDate: record.date,
            startTime: column.startTime,
            endTime: column.endTime,
            teacherId: assignment.teacherId,
            assistantIds: assignment.assistantIds,
            classroomId: selectedClassroom.classroomId || undefined,
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
    const payload = {
      ...attempted.forcePayload,
      allowStudentConflict: true,
      schedules,
    }
    const res = attempted.forcePayload.oneToOneId
      ? await createOneToOneSchedulesApi(payload)
      : await createGroupClassSchedulesApi(payload)
    if (res.code !== 200)
      throw new Error(res.message || '强制排课失败')
    conflictDetailModalOpen.value = false
    messageService.success(
      Array.isArray(attempted?.removedAssistantIds) && attempted.removedAssistantIds.length > 0
        ? '已自动忽略与主教重复的助教，并按学员冲突方式排课'
        : '已按学员冲突方式排课，课表将标记冲突',
    )
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

async function forceDragScheduleDespiteStudentConflict() {
  const attempted = dragConflictDetailState.value.attempted
  if (!attempted?.forceAllowed || !attempted?.forcePayload) {
    messageService.warning('当前冲突类型暂不支持强制调课')
    return
  }

  forcingConflictSchedule.value = true
  updatingDraggedSchedule.value = true
  try {
    const res = await batchUpdateTeachingSchedulesApi({
      ...attempted.forcePayload,
      allowStudentConflict: true,
    })
    if (res.code !== 200)
      throw new Error(res.message || '强制调课失败')
    dragConflictDetailOpen.value = false
    const lessonText = attempted.studentText || attempted.courseName || '课程'
    messageService.success(`已按学员冲突方式调课：${lessonText}`)
    emitter.emit(EVENTS.REFRESH_DATA)
  }
  catch (error) {
    console.error('force drag schedule despite student conflict failed', error)
    messageService.error(error?.response?.data?.message || error?.message || '强制调课失败')
    await loadTimetableMatrix()
  }
  finally {
    forcingConflictSchedule.value = false
    updatingDraggedSchedule.value = false
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
  await openConflictLocatingState(conflictJumpActionKey(item))
  if (!item?.jumpCellKey) {
    locatingConflictItemKey.value = ''
    messageService.warning('当前冲突课程暂不支持定位')
    return
  }

  const jumpTeacherId = String(item.teacherId || '').trim()
  const jumpTeacherName = String(item.teacherName || '').trim()
  const teacherFilterMissing = jumpTeacherId && !filterTeacherId.value.includes(jumpTeacherId)
  const modalFlags = {
    closeScheduledConflictModal: scheduledConflictDetailOpen.value,
    closeConflictDetailModal: conflictDetailModalOpen.value,
    closeDragConflictModal: dragConflictDetailOpen.value,
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
      locatingConflictItemKey.value = ''
      closeConflictModalsByFlags(modalFlags)
      messageService.success('已定位到冲突课程')
      return
    }
  }

  pendingConflictJump = {
    ...modalFlags,
    cellKey: item.jumpCellKey,
    teacherId: jumpTeacherId,
    teacherName: jumpTeacherName,
    allowAppendTeacherFilter: Boolean(teacherFilterMissing),
    teacherFilterExpanded: false,
  }
  await loadTimetableMatrix()
}

async function jumpToScheduledConflictSchedule(item) {
  const jumpItem = buildConflictJumpItem(item, 0)
  await openConflictLocatingState([
    String(item?.teacherId || '').trim(),
    String(item?.teacherName || '').trim(),
    String(item?.date || '').trim(),
    String(item?.timeText || '').trim(),
  ].join('|'))
  await jumpToConflictSchedule(jumpItem)
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
    else if (currentModel.value === '2' && Array.isArray(reason?.existingSchedules) && reason.existingSchedules.length) {
      openClassConflictModal(reason, column, record)
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
  classroomText,
  warningText,
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
      ...(classroomText != null ? [buildConfirmField('上课教室', classroomText)] : []),
      buildConfirmField('所在组别', groupLabel || '当前组'),
    ]),
    ...(warningText
      ? [
          h('div', {
            style: {
              padding: '12px 14px',
              borderRadius: '12px',
              background: '#fff7e6',
              border: '1px solid #ffd591',
              color: '#ad6800',
              fontSize: '13px',
              lineHeight: '22px',
            },
          }, warningText),
        ]
      : []),
    h('div', {
      style: {
        padding: '12px 14px',
        borderRadius: '12px',
        background: '#f5f5f5',
        color: '#595959',
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
  classroomText,
  warningText,
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
        classroomText,
        warningText,
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
    lessonTitle: text.courseType === 1
      ? `${studentText || '学员'} · ${text.courseName || '课程'}`
      : scheduleLessonTitle(text),
    dateLabel: `${month}月${day}日 ${formatWeek(record.date)} 第${lessonIndex}节`,
    timeLabel: `${column.startTime}-${column.endTime}`,
    teacherName: text.teacherName || record.name,
    mainTeacherId: String(text.mainTeacherId || ''),
    assistantText: text.assistantText || '未安排',
    assistantIds: Array.isArray(text.assistantIds) ? text.assistantIds : [],
    classroomId: String(text.classroomId || ''),
    classroomName: text.classroomName || '',
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
    const scheduleLabel = detail.courseType === 1 ? '1v1' : '班课'
    messageService.success(`已删除 ${month}月${day}日 第${lessonIndex}节 ${scheduleLabel}日程，主教/助教课表已同步移除`)
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

function buildBatchPlanScheduleFromDetail(detail) {
  if (!detail?.scheduleId)
    return null
  const studentNames = Array.isArray(detail?.text?.studentNames)
    ? detail.text.studentNames.map(item => String(item?.name || '').trim()).filter(Boolean)
    : []
  const firstStudentId = Array.isArray(detail?.text?.studentNames)
    ? String(detail.text.studentNames[0]?.id || '').trim()
    : ''
  const assistantNames = detail.assistantText && detail.assistantText !== '未安排'
    ? String(detail.assistantText).split('、').map(item => item.trim()).filter(Boolean)
    : []
  const lessonDate = String(detail.record?.date || '').trim()
  const startTime = String(detail.column?.startTime || '').trim()
  const endTime = String(detail.column?.endTime || '').trim()

  return {
    id: String(detail.scheduleId || '').trim(),
    batchNo: String(detail.text?.batchNo || '').trim() || undefined,
    batchSize: 1,
    classType: detail.courseType === 1 ? 2 : 1,
    teachingClassId: String(detail.text?.classId || '').trim(),
    teachingClassName: String(detail.text?.className || '').trim(),
    studentId: firstStudentId,
    studentName: studentNames.join('、'),
    lessonId: String(detail.text?.courseId || '').trim(),
    lessonName: String(detail.courseName || '').trim(),
    teacherId: String(detail.text?.teacherId || detail.record?.teacherId || '').trim(),
    teacherName: String(detail.teacherName || '').trim(),
    assistantIds: Array.isArray(detail.assistantIds) ? detail.assistantIds : [],
    assistantNames,
    classroomId: String(detail.classroomId || '').trim(),
    classroomName: String(detail.classroomName || '').trim(),
    lessonDate,
    startAt: lessonDate && startTime ? `${lessonDate} ${startTime}:00` : '',
    endAt: lessonDate && endTime ? `${lessonDate} ${endTime}:00` : '',
    status: 1,
    callStatus: detail.text?.callStatusKey === 'signed' ? 2 : 1,
    callStatusText: detail.text?.callStatusKey === 'signed' ? '已点名' : '未点名',
    conflict: Boolean(detail.text?.scheduledConflict),
    conflictTypes: Array.isArray(detail.text?.scheduledConflictTypes) ? detail.text.scheduledConflictTypes : [],
  }
}

function openScheduledLessonBatchPlanEdit() {
  const detail = scheduledLessonDetailState.value
  if (!(detail.courseType === 1 && detail.isMain !== false))
    return
  const schedule = buildBatchPlanScheduleFromDetail(detail)
  if (!schedule) {
    messageService.warning('当前日程缺少编辑标识，请刷新后重试')
    return
  }
  scheduledLessonDetailOpen.value = false
  currentBatchPlanSchedule.value = schedule
  scheduleBatchPlanEditOpen.value = true
}

function handleBatchPlanUpdated() {
  scheduleBatchPlanEditOpen.value = false
  currentBatchPlanSchedule.value = null
  emitter.emit(EVENTS.REFRESH_DATA)
}

// 排课
async function handleScheduleClick(timeSlot, column, record) {
  if (currentModel.value === '1') {
    if (!oneToOneRecordId.value) {
      messageService.warning('请先在上方选择要排课的 1 对 1 记录')
      return
    }

    const studentInfo = oneToOneData.value.find(
      item => item.id === String(oneToOneRecordId.value),
    )

    if (!studentInfo) {
      messageService.warning('所选 1 对 1 已不在列表中，请重新选择或刷新页面')
      return
    }

    // 获取月份和日期信息
    const dateObj = dayjs(record.date)
    const month = dateObj.format('M')
    const day = dateObj.format('D')
    const lessonIndex = getLessonIndex(column.startTime)
    const assignment = buildOneToOneScheduleAssignment(
      record.teacherId,
      normalizedSelectedAssistantIds.value,
    )
    const selectedClassroom = resolveSelectedOneToOneClassroom()

    void confirmScheduleWithOptionalSkip({
      modeLabel: '1v1',
      modeColor: '#1677ff',
      targetLabel: '排课对象',
      targetValue: studentInfo.studentName,
      courseName: studentInfo.courseName,
      dateLabel: `${month}月${day}日 ${formatWeek(record.date)} 第${lessonIndex}节`,
      timeLabel: `${column.startTime}-${column.endTime}`,
      teacherName: record.name,
      assistantText: assistantTextForIds(assignment.assistantIds),
      classroomText: selectedClassroom.classroomName || '未设置教室',
      warningText: assignment.removedAssistantIds.length > 0 ? '主教与助教不能为同一人，系统已自动忽略重复助教。' : '',
      groupLabel: activeGroupLabel.value || '当前组',
      async onConfirm() {
        creatingOneToOneSchedule.value = true
        try {
          const res = await createOneToOneSchedulesApi({
            oneToOneId: String(oneToOneRecordId.value),
            teacherId: String(record.teacherId),
            assistantIds: assignment.assistantIds,
            classroomId: selectedClassroom.classroomId || undefined,
            schedules: [{
              lessonDate: record.date,
              startTime: column.startTime,
              endTime: column.endTime,
              assistantIds: assignment.assistantIds,
              classroomId: selectedClassroom.classroomId || undefined,
            }],
          })
          if (res.code !== 200)
            throw new Error(res.message || '创建1对1日程失败')

          messageService.success(
            assignment.removedAssistantIds.length > 0
              ? `已自动忽略与主教重复的助教，并为 ${studentInfo.studentName} 创建 ${month}月${day}日 第${lessonIndex}节课`
              : `已为 ${studentInfo.studentName} 创建 ${month}月${day}日 第${lessonIndex}节课`,
          )
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
    if (!classId.value) {
      messageService.warning('请先在上方选择要排课的班级')
      return
    }

    const classInfo = await handleClass(classId.value)

    if (!classInfo) {
      messageService.warning('请选择有效的班级')
      return
    }

    // 检查时间冲突
    if (timeSlot.conflict) {
      messageService.warning('该时间段已有冲突，不可排课')
      return
    }

    // 获取月份和日期信息
    const dateObj = dayjs(record.date)
    const month = dateObj.format('M')
    const day = dateObj.format('D')
    const lessonIndex = getLessonIndex(column.startTime)
    const classAssignment = buildClassScheduleAssignment(
      classInfo,
      record.teacherId,
      normalizedSelectedClassAssistantIds.value,
    )
    const hasDuplicateClassAssistant = classAssignment.removedAssistantIds.length > 0
    const previewAssistantText = classAssistantTextForIds(classAssignment.assistantIds)
    const previewClassroom = resolveSelectedClassroomForClass(classInfo)

    void confirmScheduleWithOptionalSkip({
      modeLabel: '班课',
      modeColor: '#13c2c2',
      targetLabel: '排课班级',
      targetValue: classInfo.name,
      courseName: classInfo.courseName,
      dateLabel: `${month}月${day}日 ${formatWeek(record.date)} 第${lessonIndex}节`,
      timeLabel: `${column.startTime}-${column.endTime}`,
      teacherName: record.name,
      assistantText: previewAssistantText,
      classroomText: previewClassroom.classroomName || '未设置教室',
      warningText: hasDuplicateClassAssistant ? '主教与助教不能为同一人，系统已自动忽略重复助教。' : '',
      groupLabel: activeGroupLabel.value || '当前组',
      async onConfirm() {
        creatingOneToOneSchedule.value = true
        try {
          const ensuredClassInfo = await ensureClassLoaded(classInfo.id) || classInfo
          const selectedClassroom = resolveSelectedClassroomForClass(ensuredClassInfo)
          const assignment = buildClassScheduleAssignment(
            ensuredClassInfo,
            record.teacherId,
            normalizedSelectedClassAssistantIds.value,
          )
          const res = await createGroupClassSchedulesApi({
            groupClassId: ensuredClassInfo.id,
            teacherId: assignment.teacherId,
            assistantIds: assignment.assistantIds,
            classroomId: selectedClassroom.classroomId || undefined,
            schedules: [{
              lessonDate: record.date,
              startTime: column.startTime,
              endTime: column.endTime,
              teacherId: assignment.teacherId,
              assistantIds: assignment.assistantIds,
              classroomId: selectedClassroom.classroomId || undefined,
            }],
          })
          if (res.code !== 200)
            throw new Error(res.message || '创建班课日程失败')

          messageService.success(
            assignment.removedAssistantIds.length > 0
              ? `已自动忽略与主教重复的助教，并为 ${ensuredClassInfo.name} 创建 ${month}月${day}日 第${lessonIndex}节课`
              : `已为 ${ensuredClassInfo.name} 创建 ${month}月${day}日 第${lessonIndex}节课`,
          )
          emitter.emit(EVENTS.REFRESH_DATA)
        }
        catch (error) {
          console.error('create group class schedule failed', error)
          messageService.error(error?.response?.data?.message || error?.message || '创建班课日程失败')
          await loadTimetableMatrix()
        }
        finally {
          creatingOneToOneSchedule.value = false
        }
      },
    })
  }
}

function getLessonIndex(startTime) {
  const slots = activePeriodSlots.value
  const i = slots.findIndex(s => s.start === startTime)
  return i >= 0 ? i + 1 : ''
}

function hasActiveScheduleTarget() {
  if (currentModel.value === '1')
    return Boolean(String(oneToOneRecordId.value || '').trim())
  return Boolean(String(classId.value || '').trim())
}

function emptyLessonStatusText(lesson) {
  if (!hasActiveScheduleTarget())
    return ''
  if (currentModel.value === '2' && classId.value && classConflictLoading.value)
    return '空闲时段(检测中...)'
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
  if (!draggingScheduleState.value)
    document.body.style.userSelect = ''
}

function showBlockedScheduleDragHint(message) {
  const now = Date.now()
  if (now - lastBlockedScheduleDragHintAt < 1200)
    return
  lastBlockedScheduleDragHintAt = now
  messageService.info(message)
}

function suppressScheduledLessonClick(duration = 220) {
  suppressScheduledLessonClickUntil = Date.now() + duration
}

function consumeScheduledLessonClickSuppressed() {
  if (Date.now() <= suppressScheduledLessonClickUntil) {
    suppressScheduledLessonClickUntil = 0
    return true
  }
  return false
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

function isScheduleBeforeToday(text) {
  const lessonDate = String(text?.lessonDate || '').trim()
  if (!lessonDate)
    return false
  return dayjs(lessonDate).isBefore(dayjs().startOf('day'), 'day')
}

function isScheduleDraggable(text) {
  return currentModel.value === '1'
    && text?.courseType === 1
    && text?.isMain !== false
    && !isScheduleBeforeToday(text)
    && text?.callStatusKey !== 'signed'
    && Boolean(text?.scheduleId)
    && Boolean(text?.classId)
}

function resolveScheduleDragBlockedMessage(text) {
  if (text?.callStatusKey === 'signed')
    return '当前课程已点名，暂不支持拖拽调课'
  if (isScheduleBeforeToday(text))
    return '过去的日程，不允许拖拽调课'
  if (text?.courseType === 1 && text?.isMain === false)
    return '当前是助教课表，暂不支持拖拽调课，请在主教老师所在行操作'
  if (text?.courseType === 2)
    return '当前仅支持 1v1 主教课程拖拽调课'
  if (text?.courseType === 1)
    return '当前课程暂不支持拖拽调课'
  return ''
}

function buildDraggingScheduleState(text, column, record) {
  const lessonDate = scheduleCellDate(column, record)
  const startTime = scheduleCellStartTime(column, record)
  const endTime = scheduleCellEndTime(column, record)
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
    sourceDate: lessonDate || '',
    sourceStartTime: startTime || '',
    sourceEndTime: endTime || '',
    sourceCellKey: buildAvailabilitySlotKey(record?.teacherId, lessonDate, startTime, endTime),
  }
}

function buildDragTarget(column, record) {
  const lessonDate = scheduleCellDate(column, record)
  const startTime = scheduleCellStartTime(column, record)
  const endTime = scheduleCellEndTime(column, record)
  return {
    key: buildAvailabilitySlotKey(record?.teacherId, lessonDate, startTime, endTime),
    teacherId: String(record?.teacherId || '').trim(),
    teacherName: record?.name || '-',
    lessonDate: lessonDate || '',
    startTime: startTime || '',
    endTime: endTime || '',
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

function mergeDragExistingSchedules(...groups) {
  const map = new Map()
  groups.flat().forEach((item) => {
    if (!item)
      return
    const key = [
      item.date,
      item.timeText,
      item.teacherId,
      item.teacherName,
      item.name,
      item.classroomName,
      (item.conflictTypes || []).join(','),
    ].join('|')
    if (!map.has(key))
      map.set(key, item)
  })
  return [...map.values()]
}

function mergeDragValidationResults(primary, secondary) {
  if (!secondary)
    return primary
  if (primary?.valid === false && secondary?.valid !== false)
    return primary
  if (primary?.valid !== false && secondary?.valid === false)
    return secondary
  if (primary?.valid !== false && secondary?.valid !== false)
    return primary

  const conflictTypes = uniqueConflictTypes([
    ...(primary?.conflictTypes || []),
    ...(secondary?.conflictTypes || []),
  ])
  return {
    ...dragConflictStateFromTypes(conflictTypes, secondary?.message || primary?.message || '当前空点不可调课'),
    existingSchedules: mergeDragExistingSchedules(primary?.existingSchedules || [], secondary?.existingSchedules || []),
  }
}

async function checkDragTargetsTeacherAvailability(targets, dragState) {
  const res = await checkOneToOneScheduleAvailabilityApi({
    oneToOneId: dragState.oneToOneId,
    excludeIds: [dragState.scheduleId],
    schedules: targets.map(target => ({
      teacherId: target.teacherId,
      lessonDate: target.lessonDate,
      startTime: target.startTime,
      endTime: target.endTime,
    })),
  })
  if (res.code !== 200 || !res.result)
    throw new Error(res.message || '检测调课空点失败')

  return new Map(
    (Array.isArray(res.result.items) ? res.result.items : [])
      .map(item => [buildAvailabilitySlotKey(item.teacherId, item.lessonDate, item.startTime, item.endTime), item]),
  )
}

async function checkDragTargetAssistantAvailability(target, dragState) {
  if (!Array.isArray(dragState?.assistantIds) || !dragState.assistantIds.length)
    return null

  const res = await checkAssistantScheduleAvailabilityApi({
    oneToOneId: dragState.oneToOneId,
    assistantIds: dragState.assistantIds,
    excludeIds: [dragState.scheduleId],
    schedules: [{
      lessonDate: target.lessonDate,
      startTime: target.startTime,
      endTime: target.endTime,
    }],
  })
  if (res.code !== 200 || !res.result)
    throw new Error(res.message || '检测助教空闲状态失败')

  const invalidItems = (Array.isArray(res.result.items) ? res.result.items : []).filter(item => item.valid === false)
  if (!invalidItems.length) {
    return {
      valid: true,
      label: '可调课',
      message: `${target.teacherName} ${target.lessonDate} ${target.startTime}-${target.endTime} 可调课`,
      conflictTypes: [],
      existingSchedules: [],
    }
  }

  return {
    ...dragConflictStateFromTypes(
      uniqueConflictTypes(invalidItems.flatMap(item => item.conflictTypes || [])),
      invalidItems[0]?.message || '所选助教该时间段已有安排',
    ),
    existingSchedules: mergeDragExistingSchedules(invalidItems.flatMap(item => item.existingSchedules || [])),
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
  if (dayjs(String(target?.lessonDate || '').trim()).isBefore(dayjs().startOf('day'), 'day')) {
    return {
      valid: false,
      label: '过去日期(不可调)',
      message: '不允许拖拽调课到过去的日期',
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
      if (!lesson || hasScheduledLesson(lesson))
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
    try {
      const itemMap = await checkDragTargetsTeacherAvailability(targets, dragState)
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
    catch (fallbackError) {
      console.error('fallback drag availability check failed', fallbackError)
      const result = {
        valid: false,
        label: '检测失败',
        message: fallbackError?.response?.data?.message || fallbackError?.message || error?.response?.data?.message || error?.message || '批量检测调课空点失败',
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
      try {
        const itemMap = await checkDragTargetsTeacherAvailability([target], dragState)
        const teacherResult = buildDragValidationResultFromValidationItem(target, itemMap.get(target.key))
        const assistantResult = await checkDragTargetAssistantAvailability(target, dragState)
        const result = mergeDragValidationResults(teacherResult, assistantResult)
        dragValidationCache.set(cacheKey, result)
        return result
      }
      catch (fallbackError) {
        console.error('fallback validate drag target failed', fallbackError)
        const result = {
          valid: false,
          label: '检测失败',
          message: fallbackError?.response?.data?.message || fallbackError?.message || error?.response?.data?.message || error?.message || '检测调课空点失败',
          conflictTypes: [],
        }
        dragValidationCache.set(cacheKey, result)
        return result
      }
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
        suppressScheduledLessonClick()
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
  document.body.style.userSelect = 'none'

  blockedScheduleDragMoveHandler = (moveEvent) => {
    if (!blockedScheduleDragAttempt)
      return
    const deltaX = Math.abs(Number(moveEvent?.clientX || 0) - blockedScheduleDragAttempt.startX)
    const deltaY = Math.abs(Number(moveEvent?.clientY || 0) - blockedScheduleDragAttempt.startY)
    if (deltaX < 6 && deltaY < 6)
      return
    suppressScheduledLessonClick()
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

async function copyDraggedScheduleToTarget() {
  const payload = dragConfirmDetail.value?.payload
  if (!payload?.dragState || !payload?.target) {
    dragConfirmOpen.value = false
    return
  }

  const { dragState, target } = payload
  if (!dragState.oneToOneId) {
    messageService.warning('当前课程缺少1对1标识，暂不支持复制')
    return
  }

  const dateLabel = dayjs(target.lessonDate).format('M月D日')
  const lessonIndex = getLessonIndex(target.startTime)
  dragCopySubmitting.value = true
  creatingOneToOneSchedule.value = true
  try {
    const res = await createOneToOneSchedulesApi({
      oneToOneId: dragState.oneToOneId,
      teacherId: target.teacherId,
      assistantIds: dragState.assistantIds,
      classroomId: dragState.classroomId || undefined,
      schedules: [{
        lessonDate: target.lessonDate,
        startTime: target.startTime,
        endTime: target.endTime,
        teacherId: target.teacherId,
        assistantIds: dragState.assistantIds,
        classroomId: dragState.classroomId || undefined,
      }],
    })
    if (res.code !== 200)
      throw new Error(res.message || '复制课程失败')

    dragConfirmOpen.value = false
    dragConfirmDetail.value = {
      source: null,
      target: null,
      payload: null,
    }
    const lessonText = dragState.studentText || dragState.courseName || '课程'
    messageService.success(`已复制 ${lessonText} 到 ${dateLabel} ${formatWeek(target.lessonDate)} 第${lessonIndex}节`)
    emitter.emit(EVENTS.REFRESH_DATA)
  }
  catch (error) {
    console.error('copy teaching schedule failed', error)
    messageService.error(error?.response?.data?.message || error?.message || '复制课程失败')
    await loadTimetableMatrix()
  }
  finally {
    dragCopySubmitting.value = false
    creatingOneToOneSchedule.value = false
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
    classPickerOpen.value = false
    selectedClassAssistantIds.value = []
    classAssistantKeyword.value = ''
    classClassroomKeyword.value = ''
    selectedClassClassroomId.value = undefined
    preserveClassPickerOpen = false
    lastHandledClassId = ''
    clearClassAutoTeacherFilter()
  }
  else {
    resetOneToOnePickerState()
    selectedOneToOneClassroomId.value = undefined
  }
})

watch(
  () => String(oneToOneRecordId.value || '').trim(),
  (value, previousValue) => {
    if (value || !previousValue)
      return
    clearOneToOneAvailabilityHighlights()
  },
)

watch(
  () => String(filterOneToOneId.value || '').trim(),
  (value, previousValue) => {
    if (value || !previousValue || String(oneToOneRecordId.value || '').trim())
      return
    clearOneToOneAvailabilityHighlights()
  },
)

watch(
  () => normalizeOptionalClassroomId(selectedOneToOneClassroomId.value),
  (value, previousValue) => {
    if (value === previousValue)
      return
    if (currentModel.value !== '1' || !String(oneToOneRecordId.value || '').trim())
      return
    void detectOneToOneAvailability(oneToOneRecordId.value)
  },
)

watch(
  () => normalizeOptionalClassroomId(selectedClassClassroomId.value),
  (value, previousValue) => {
    if (value === previousValue)
      return
    if (classSelectionSyncing)
      return
    if (currentModel.value !== '2' || !String(classId.value || '').trim())
      return
    handleClassBridge(classId.value)
  },
)

watch(dragConfirmOpen, (open) => {
  if (!open && !dragConfirmSubmitting.value && !dragCopySubmitting.value) {
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
        ref="allFilterRef"
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
    <SmartTimetableToolbar
      v-model:current-model="currentModel"
      v-model:one-to-one-record-id="oneToOneRecordId"
      v-model:one-to-one-picker-open="oneToOnePickerOpen"
      v-model:class-id="classId"
      v-model:class-picker-open="classPickerOpen"
      v-model:current-time="currentTime"
      v-model:current-week="currentWeek"
      v-model:current-group="currentGroup"
      :one-to-one-list-loading="oneToOneListLoading"
      :one-to-one-data="oneToOneData"
      :render-one-to-one-dropdown="renderOneToOneDropdown"
      :filter-one-to-one-option="filterOneToOneOption"
      :render-class-dropdown="renderClassDropdown"
      :class-data="classData"
      :class-list-loading="classListLoading"
      :time-view-options="timeViewOptions"
      :format-date-range="formatDateRange"
      :is-week-like-view="isWeekLikeView"
      :group-options="groupOptions"
      :on-one-to-one-change="handle1v1"
      :on-one-to-one-dropdown-visible-change="handleOneToOneDropdownVisibleChange"
      :on-class-change="handleClassSelectionChange"
      :on-class-dropdown-visible-change="handleClassDropdownVisibleChange"
      :on-class-search="loadClassOptions"
      :on-prev="handlePrev"
      :on-next="handleNext"
      :on-this-week="handleThisWeek"
      :on-export="exportSmartTimetable"
      :export-loading="exportLoading"
    />
    <div class="st-summary-shell">
      <TimetableScheduleSummary
        :total="smartTimetableTotalSchedules"
        :unsigned-count="smartTimetableUnsignedSchedules"
      />
    </div>
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
      :has-scheduled-lesson="hasScheduledLesson"
      :open-scheduled-lesson-detail="openScheduledLessonDetail"
      :open-scheduled-conflict-detail="openScheduledConflictDetail"
      :handle-conflict-click="handleConflictClick"
      :handle-schedule-click="handleScheduleClick"
      :consume-scheduled-lesson-click-suppressed="consumeScheduledLessonClickSuppressed"
      :handle-schedule-pointer-down="handleSchedulePointerDown"
      :is-schedule-draggable="isScheduleDraggable"
      :resolve-schedule-drag-blocked-message="resolveScheduleDragBlockedMessage"
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

    <SmartTimetableScheduleDetailDrawer
      v-model:open="scheduledLessonDetailOpen"
      :detail="scheduledLessonDetailState"
      :deleting="deletingScheduledLesson"
      :editable="scheduledLessonDetailState.courseType === 1 && scheduledLessonDetailState.isMain !== false"
      @delete="deleteScheduledLessonFromDetail"
      @edit="openScheduledLessonBatchPlanEdit"
    />

    <ScheduleBatchPlanEditModal
      v-model:open="scheduleBatchPlanEditOpen"
      :schedule="currentBatchPlanSchedule"
      @updated="handleBatchPlanUpdated"
    />

    <ScheduleConflictModal
      v-model:open="scheduledConflictDetailOpen"
      :validation="scheduledConflictDetailValidation"
      :locating="Boolean(locatingConflictItemKey)"
      :get-time-extra-label="resolveConflictScheduleTimeLabel"
      title="冲突详情"
      current-title="当前冲突日程"
      existing-title="与其冲突的日程"
      fallback-message="当前日程与已有日程存在冲突"
      @jump="jumpToScheduledConflictSchedule"
    />

    <SmartTimetableConflictModal
      v-model:open="conflictDetailModalOpen"
      :forcing="forcingConflictSchedule"
      :locating="Boolean(locatingConflictItemKey)"
      :conflict-detail-state="conflictDetailState"
      @force="forceScheduleDespiteStudentConflict"
      @jump="jumpToConflictSchedule"
    />

    <SmartTimetableDragConflictModal
      v-model:open="dragConflictDetailOpen"
      :forcing="forcingConflictSchedule"
      :locating="Boolean(locatingConflictItemKey)"
      :detail="dragConflictDetailState"
      @force="forceDragScheduleDespiteStudentConflict"
      @jump="jumpToConflictSchedule"
    />

    <SmartTimetableDragConfirmModal
      v-model:open="dragConfirmOpen"
      :detail="dragConfirmDetail"
      :submitting="dragConfirmSubmitting"
      :copying="dragCopySubmitting"
      @confirm="submitDragScheduleAdjustment"
      @copy="copyDraggedScheduleToTarget"
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

.st-summary-shell {
  overflow: hidden;
  background: #fff;
}

.st-summary-shell :deep(.timetable-summary) {
  padding-top: 0;
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
