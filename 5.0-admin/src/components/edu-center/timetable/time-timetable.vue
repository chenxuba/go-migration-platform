<script setup>
import { LeftOutlined, RightOutlined } from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'
import dayjs from 'dayjs'
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import CreateSchedulePopover from './create-schedule-popover.vue'
import ScheduleBatchPlanEditModal from './schedule-batch-plan-edit-modal.vue'
import ScheduleConflictModal from './schedule-conflict-modal.vue'
import SmartTimetableScheduleDetailDrawer from './smart-timetable-schedule-detail-drawer.vue'
import TimetableScheduleHoverPopover from './timetable-schedule-hover-popover.vue'
import { listClassroomsApi } from '@/api/business-settings/classroom'
import { getOneToOneListApi } from '@/api/edu-center/one-to-one'
import { pageGroupClassesApi } from '@/api/edu-center/group-class'
import { getCourseIdAndNameApi } from '@/api/edu-center/registr-renewal'
import { cancelTeachingScheduleScopedApi, downloadTimeTimetableExcelApi, getTeachingScheduleConflictDetailApi, listTeachingSchedulesApi } from '@/api/edu-center/teaching-schedule'
import { getUserListApi } from '@/api/internal-manage/staff-manage'
import emitter, { EVENTS } from '@/utils/eventBus'
import messageService from '@/utils/messageService'
import { loadTeachingScheduleDeleteTargetCount } from './schedule-delete-scope'

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

const filterStudentId = ref(undefined)
const filterTeacherId = ref([])
const filterClassroomId = ref([])
const filterClassId = ref(undefined)
const filterOneToOneId = ref(undefined)
const filterCourseId = ref(undefined)
const filterScheduleType = ref([])
const filterCallStatus = ref(undefined)

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

const scheduleTypeOptions = [
  { id: 'group_class', value: '班级日程' },
  { id: 'one_to_one', value: '1对1日程' },
  { id: 'trial', value: '试听日程' },
]

const scheduleCallStatusOptions = [
  { id: 'unsigned', value: '未点名' },
  { id: 'signed', value: '已点名' },
]

function scheduleBadgeText(classType) {
  return Number(classType) === 1 ? '班课' : '1v1'
}

function scheduleTitle(item) {
  return item.teachingClassName
    || item.studentName
    || item.lessonName
    || (Number(item.classType) === 1 ? '班课日程' : '1对1日程')
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
  scheduleCourseSearchKey.value = searchKey
  try {
    const res = await getCourseIdAndNameApi({ searchKey })
    if (res.code !== 200)
      return
    const resultData = (Array.isArray(res.result) ? res.result : []).map(item => ({
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
  scheduleCourseFinished.value = true
}

const timeOptions = [
  { key: 'day', label: '日' },
  { key: 'week', label: '周' },
]

const currentTime = ref('week')
const currentDate = ref(dayjs())
const now = ref(dayjs())
const exportLoading = ref(false)
const scheduleLoading = ref(false)
const scheduleRows = ref([])
const scheduleDetailOpen = ref(false)
const currentDetailSchedule = ref(null)
const currentScheduleDetail = ref(null)
const deletingScheduleDetail = ref(false)
const scheduleBatchPlanEditOpen = ref(false)
const currentBatchPlanSchedule = ref(null)
const scheduleBatchPlanEditScope = ref('batch')
const scheduleBatchPlanAction = ref('edit')
const scheduleConflictOpen = ref(false)
const scheduleConflictValidation = ref(null)
const scheduleConflictLoading = ref(false)
const locatingConflictItemKey = ref('')
const focusedScheduleId = ref('')
const headerScrollRef = ref(null)
const boardScrollRef = ref(null)
const scheduleViewportWidth = ref(0)
let syncingScroll = false
let scheduleLoadSeq = 0
let focusedScheduleTimer = null
let pendingConflictJump = null

/** 与教师矩阵课表一致：日期条横向滚动时「钉」在可视区中心的浮动芯片 */
const timeHeaderTimeColWidth = 84
const floatingDatePillWidth = 156
const floatingDateStyles = ref({})
let layoutResizeObserver = null

const weekdayLabels = ['周一', '周二', '周三', '周四', '周五', '周六', '周日']
const timelineStart = 8 * 60
const timelineEnd = 22 * 60
const timelineTopPadding = 18
const hourRowHeight = 96
const timelineBottomPadding = 28
/** 避免过短日程信息完全挤爆，但不能大到吃掉真实课间距 */
const scheduleCardMinHeight = 60
const scheduleCardMinWidth = 152
const scheduleCardGap = 5
/** 列宽 = 左 6px + 卡片 + 右 6px；与 eventStyle 里 +6 对齐，避免只左边留白、右边贴线 */
const scheduleColumnHorizontalInset = 6
const baseDateColumnWidth
  = scheduleCardMinWidth + scheduleColumnHorizontalInset * 2
const overlapExtraWidth = scheduleCardMinWidth + scheduleCardGap

const scheduleLegend = [
  {
    key: 'unsigned',
    label: '未点名（教师/课程）',
    type: 'bar',
    color: 'linear-gradient(90deg, #39b8ff 0%, #6c5cff 50%, #74d87f 100%)',
  },
  {
    key: 'signed',
    label: '已点名',
    type: 'bar',
    color: '#b7bec8',
  },
  {
    key: 'trial',
    label: '含试听学员',
    type: 'icon',
  },
  {
    key: 'conflict',
    label: '日程冲突',
    type: 'icon-danger',
  },
]

function getWeekStart(value = dayjs()) {
  const current = dayjs(value)
  const diff = current.day() === 0 ? -6 : 1 - current.day()
  return current.add(diff, 'day').startOf('day')
}

function emitCurrentWeekRange(value = currentDate.value) {
  const start = getWeekStart(value)
  emit('week-range-change', {
    startDate: start.format('YYYY-MM-DD'),
    endDate: start.add(6, 'day').format('YYYY-MM-DD'),
  })
}

let nowTimer = null

onMounted(() => {
  nowTimer = setInterval(() => {
    now.value = dayjs()
  }, 30 * 1000)
  loadSchedules()
  emitter.on(EVENTS.REFRESH_DATA, loadSchedules)
})

onUnmounted(() => {
  if (nowTimer) {
    clearInterval(nowTimer)
    nowTimer = null
  }
  if (focusedScheduleTimer) {
    clearTimeout(focusedScheduleTimer)
    focusedScheduleTimer = null
  }
  emitter.off(EVENTS.REFRESH_DATA, loadSchedules)
  if (layoutResizeObserver) {
    layoutResizeObserver.disconnect()
    layoutResizeObserver = null
  }
})

watch(currentTime, () => {
  currentDate.value = dayjs()
})

watch(currentDate, value => emitCurrentWeekRange(value), { immediate: true })

function formatDateRange(value) {
  if (!value)
    return ''

  if (currentTime.value === 'day')
    return value.format('YYYY年MM月DD日')

  const start = getWeekStart(value)
  const end = start.add(6, 'day')

  if (start.year() === end.year() && start.month() === end.month())
    return `${start.format('YYYY年MM月DD日')} ~ ${end.format('DD日')}`
  if (start.year() === end.year())
    return `${start.format('YYYY年MM月DD日')} ~ ${end.format('MM月DD日')}`
  return `${start.format('YYYY年MM月DD日')} ~ ${end.format('YYYY年MM月DD日')}`
}

function handlePrev() {
  currentDate.value
    = currentTime.value === 'day'
      ? currentDate.value.subtract(1, 'day')
      : currentDate.value.subtract(1, 'week')
}

function handleNext() {
  currentDate.value
    = currentTime.value === 'day'
      ? currentDate.value.add(1, 'day')
      : currentDate.value.add(1, 'week')
}

function handleGoThisWeek() {
  currentDate.value = dayjs()
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

function resolveExportDateRange() {
  if (currentTime.value === 'day') {
    const date = dayjs(currentDate.value).startOf('day')
    const dateText = date.format('YYYY-MM-DD')
    return {
      startDate: dateText,
      endDate: dateText,
    }
  }
  const start = getWeekStart(currentDate.value)
  return {
    startDate: start.format('YYYY-MM-DD'),
    endDate: start.add(6, 'day').format('YYYY-MM-DD'),
  }
}

async function exportTimeTimetable() {
  exportLoading.value = true
  try {
    const scheduleTeacherIds = filterTeacherId.value.join(',')
    const classroomIds = filterClassroomId.value.join(',')
    const { startDate, endDate } = resolveExportDateRange()
    const res = await downloadTimeTimetableExcelApi({
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
      || `时间课表明细_${dayjs().format('YYYYMMDDHHmmss')}.xlsx`
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    a.click()
    URL.revokeObjectURL(url)
    messageService.success('已导出课表')
  }
  catch (error) {
    console.error('export time timetable failed', error)
    messageService.error('导出失败')
  }
  finally {
    exportLoading.value = false
  }
}

const isViewingTodayOrThisWeek = computed(() => {
  if (currentTime.value === 'day')
    return currentDate.value.isSame(now.value, 'day')
  return getWeekStart(currentDate.value).isSame(getWeekStart(now.value), 'day')
})

function formatClock(minutes) {
  const hour = String(Math.floor(minutes / 60)).padStart(2, '0')
  const minute = String(minutes % 60).padStart(2, '0')
  return `${hour}:${minute}`
}

const displayDates = computed(() => {
  if (currentTime.value === 'day')
    return [dayjs(currentDate.value).startOf('day')]
  const start = getWeekStart(currentDate.value)
  return Array.from({ length: 7 }, (_, index) => start.add(index, 'day'))
})

const todayKey = computed(() => now.value.format('YYYY-MM-DD'))
const currentTimeMinutes = computed(
  () => now.value.hour() * 60 + now.value.minute(),
)
const currentTimeLabel = computed(() => now.value.format('HH:mm'))
const showCurrentTimeLine = computed(() => {
  if (
    currentTimeMinutes.value < timelineStart
    || currentTimeMinutes.value > timelineEnd
  ) {
    return false
  }
  return displayDates.value.some(
    date => date.format('YYYY-MM-DD') === todayKey.value,
  )
})

const queryDateRange = computed(() => {
  const dates = displayDates.value
  return {
    startDate: dates[0]?.format('YYYY-MM-DD') || dayjs().format('YYYY-MM-DD'),
    endDate:
      dates[dates.length - 1]?.format('YYYY-MM-DD')
      || dayjs().format('YYYY-MM-DD'),
  }
})

async function loadSchedules() {
  const seq = ++scheduleLoadSeq
  scheduleLoading.value = true
  try {
    const scheduleTeacherIds = filterTeacherId.value.join(',')
    const classroomIds = filterClassroomId.value.join(',')
    const res = await listTeachingSchedulesApi({
      startDate: queryDateRange.value.startDate,
      endDate: queryDateRange.value.endDate,
      studentId: filterStudentId.value,
      scheduleTeacherIds: scheduleTeacherIds || undefined,
      classroomIds: classroomIds || undefined,
      groupClassIds: filterClassId.value ? String(filterClassId.value) : undefined,
      oneToOneClassIds: filterOneToOneId.value ? String(filterOneToOneId.value) : undefined,
      lessonIds: filterCourseId.value ? String(filterCourseId.value) : undefined,
      scheduleTypes: filterScheduleType.value.length ? filterScheduleType.value.join(',') : undefined,
      callStatuses: filterCallStatus.value ? String(filterCallStatus.value) : undefined,
    })
    if (seq !== scheduleLoadSeq)
      return
    if (res.code === 200) {
      scheduleRows.value = Array.isArray(res.result) ? res.result : []
      return
    }
    scheduleRows.value = []
  }
  catch (error) {
    console.error('load schedules failed', error)
    if (seq !== scheduleLoadSeq)
      return
    scheduleRows.value = []
  }
  finally {
    if (seq === scheduleLoadSeq) {
      scheduleLoading.value = false
      await nextTick()
      updateScheduleViewportWidth()
      updateFloatingDatePositions(boardScrollRef.value?.scrollLeft ?? headerScrollRef.value?.scrollLeft ?? 0)
      await flushPendingConflictJump()
    }
  }
}

watch(currentDate, () => {
  loadSchedules()
})

watch(
  [filterStudentId, filterTeacherId, filterClassroomId, filterClassId, filterOneToOneId, filterCourseId, filterScheduleType, filterCallStatus],
  () => {
    loadSchedules()
  },
  { deep: true },
)

const mockSchedules = computed(() =>
  scheduleRows.value.map(item => ({
    id: item.id,
    batchNo: item.batchNo,
    batchSize: item.batchSize || 1,
    classType: item.classType,
    dateKey: dayjs(item.startAt).format('YYYY-MM-DD'),
    startAt: dayjs(item.startAt),
    endAt: dayjs(item.endAt),
    title: scheduleTitle(item),
    course: item.lessonName || '-',
    teacher: item.teacherName || '-',
    classroom: item.classroomName || '-',
    studentText: item.studentName ? `学员：${item.studentName}` : '-',
    status: Number(item.callStatus) === 2 ? 'signed' : 'unsigned',
    conflict: item.conflict === true,
    conflictTypes: item.conflictTypes || [],
    hasTrial: false,
    raw: item,
  })),
)

const headerSummaries = computed(() =>
  displayDates.value.map((date) => {
    const key = date.format('YYYY-MM-DD')
    const count = mockSchedules.value.filter(
      item => item.dateKey === key,
    ).length
    return {
      key,
      date,
      count,
    }
  }),
)

const unsignedCount = computed(
  () => mockSchedules.value.filter(item => item.status === 'unsigned').length,
)

const hourMarks = computed(() =>
  Array.from(
    { length: timelineEnd / 60 - timelineStart / 60 + 1 },
    (_, index) => timelineStart + index * 60,
  ),
)

const hoverSlots = computed(() =>
  hourMarks.value.slice(0, -1).map((startMinutes, index) => ({
    key: `slot-${startMinutes}`,
    startMinutes,
    endMinutes: hourMarks.value[index + 1],
  })),
)

const timelineHeight = computed(
  () =>
    timelineTopPadding
    + (hourMarks.value.length - 1) * hourRowHeight
    + timelineBottomPadding,
)

function minuteOffset(minutes) {
  return timelineTopPadding + ((minutes - timelineStart) / 60) * hourRowHeight
}

function normalizeScheduleItem(item) {
  return {
    ...item,
    startMinutes: item.startAt.hour() * 60 + item.startAt.minute(),
    endMinutes: item.endAt.hour() * 60 + item.endAt.minute(),
  }
}

function isOneToOneSchedule(schedule) {
  return Number(schedule?.classType) === 2
}

function scheduleAssistantText(schedule) {
  const list = Array.isArray(schedule?.assistantNames)
    ? schedule.assistantNames.map(item => String(item || '').trim()).filter(Boolean)
    : []
  return list.length ? list.join('、') : '未安排'
}

function scheduleStudentSummary(schedule) {
  const text = String(schedule?.studentName || '').trim()
  return text || '-'
}

function scheduleConflictSummary(schedule) {
  const types = Array.isArray(schedule?.conflictTypes)
    ? schedule.conflictTypes.map(item => String(item || '').trim()).filter(Boolean)
    : []
  if (types.length)
    return `${types.join('、')}冲突`
  return schedule?.conflict ? '当前课程存在冲突' : ''
}

function conflictBadgeTooltip(event) {
  const types = Array.isArray(event?.raw?.conflictTypes)
    ? event.raw.conflictTypes.map(item => String(item || '').trim()).filter(Boolean)
    : []
  if (types.length)
    return `冲突原因：${types.join('、')}冲突，点击查看详情`
  return '当前课程存在冲突，点击查看详情'
}

function scheduleHoverTitle(schedule) {
  if (isOneToOneSchedule(schedule))
    return String(schedule?.lessonName || '').trim() || '1对1日程'
  return String(schedule?.teachingClassName || '').trim()
    || String(schedule?.lessonName || '').trim()
    || '班课日程'
}

function scheduleTimeTextFromEvent(event) {
  if (!event?.startAt || !event?.endAt)
    return '-'
  return `${event.startAt.format('YYYY-MM-DD')} ${event.startAt.format('HH:mm')} ~ ${event.endAt.format('HH:mm')}`
}

function buildScheduleDrawerDetail(schedule) {
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

function openScheduleDetail(item) {
  const schedule = item?.raw || item || null
  if (!schedule)
    return
  currentDetailSchedule.value = schedule
  currentScheduleDetail.value = buildScheduleDrawerDetail(schedule)
  scheduleDetailOpen.value = true
}

function computeOverlapPeak(items = []) {
  const columns = []
  let peak = 1
  items.forEach((item) => {
    let columnIndex = columns.findIndex(
      endValue => endValue <= item.startMinutes,
    )
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

function updateScheduleViewportWidth() {
  const headerWidth = Number(headerScrollRef.value?.clientWidth || 0)
  const boardWidth = Number(boardScrollRef.value?.clientWidth || 0)
  scheduleViewportWidth.value = Math.max(headerWidth, boardWidth, 0)
}

const intrinsicDateColumnWidths = computed(() => {
  const map = new Map()
  headerSummaries.value.forEach((item) => {
    const list = mockSchedules.value
      .filter(schedule => schedule.dateKey === item.key)
      .map(normalizeScheduleItem)
      .sort((a, b) => a.startAt.valueOf() - b.startAt.valueOf())
    const peak = computeOverlapPeak(list)
    map.set(
      item.key,
      baseDateColumnWidth + Math.max(0, peak - 1) * overlapExtraWidth,
    )
  })
  return map
})

const dateColumnWidths = computed(() => {
  const map = new Map()
  const items = headerSummaries.value
  if (!items.length)
    return map

  const entries = items.map(item => ({
    key: item.key,
    width: intrinsicDateColumnWidths.value.get(item.key) || baseDateColumnWidth,
  }))
  const minGridWidth = timeHeaderTimeColWidth + entries.reduce((sum, item) => sum + item.width, 0)
  const extraWidthPerColumn = scheduleViewportWidth.value > minGridWidth
    ? (scheduleViewportWidth.value - minGridWidth) / entries.length
    : 0

  entries.forEach((item) => {
    map.set(item.key, item.width + extraWidthPerColumn)
  })

  return map
})

const totalGridWidth = computed(() =>
  timeHeaderTimeColWidth
  + headerSummaries.value.reduce(
    (sum, item) => sum + (dateColumnWidths.value.get(item.key) || baseDateColumnWidth),
    0,
  ),
)

const gridTrackStyle = computed(() => ({
  width: '100%',
  minWidth: `${totalGridWidth.value}px`,
}))

function getDayColumnLeftAndWidth(dayIndex) {
  const items = headerSummaries.value
  if (dayIndex < 0 || dayIndex >= items.length)
    return { left: 0, width: 0 }
  let left = timeHeaderTimeColWidth
  for (let i = 0; i < dayIndex; i++) {
    const w = dateColumnWidths.value.get(items[i].key) || baseDateColumnWidth
    left += w
  }
  const key = items[dayIndex].key
  const width = dateColumnWidths.value.get(key) || baseDateColumnWidth
  return { left, width }
}

/** 以表体 scrollLeft 参与计算；可视宽度取表头滚动容器（与教师矩阵一致）。 */
function updateFloatingDatePositions(scrollLeftOverride) {
  const headerEl = headerScrollRef.value
  const boardEl = boardScrollRef.value
  const items = headerSummaries.value
  if (!headerEl || !items.length) {
    floatingDateStyles.value = {}
    return
  }
  const scrollLeft = scrollLeftOverride ?? boardEl?.scrollLeft ?? headerEl.scrollLeft
  const containerWidth
    = headerEl.clientWidth > 0
      ? headerEl.clientWidth
      : Math.max(300, (boardEl?.clientWidth ?? 800) - 100)
  const pillW = floatingDatePillWidth
  const next = {}

  items.forEach((item, index) => {
    const { left: leftOffset, width: dayColumnWidth } = getDayColumnLeftAndWidth(index)
    const rightOffset = leftOffset + dayColumnWidth
    const expandedScrollLeft = Math.max(0, scrollLeft - 200)
    const expandedScrollRight = scrollLeft + containerWidth + 200
    const inPlay = rightOffset > expandedScrollLeft && leftOffset < expandedScrollRight

    if (!inPlay || dayColumnWidth <= 0) {
      next[item.key] = {
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

    next[item.key] = {
      left: `${finalPosition}px`,
      opacity: '1',
      visibility: 'visible',
    }
  })
  floatingDateStyles.value = next
}

function getFloatingStyle(dateKey) {
  return floatingDateStyles.value[dateKey] ?? {
    left: '0',
    opacity: '0',
    visibility: 'hidden',
  }
}

function floatingPillStyle(dateKey) {
  const s = getFloatingStyle(dateKey)
  return {
    width: `${floatingDatePillWidth}px`,
    left: s.left,
    opacity: s.opacity,
    visibility: s.visibility,
  }
}

const gridTemplateStyle = computed(() => {
  const dateCols = headerSummaries.value.map((item) => {
    const w = dateColumnWidths.value.get(item.key) || baseDateColumnWidth
    return `${w}px`
  })
  return {
    gridTemplateColumns: `84px ${dateCols.join(' ')}`,
    width: '100%',
    minWidth: `${totalGridWidth.value}px`,
  }
})

function buildClusterLayouts(clusterItems = []) {
  const columns = []
  let peakColumns = 0

  const assigned = clusterItems.map((item) => {
    let columnIndex = columns.findIndex(
      endValue => endValue <= item.startMinutes,
    )
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
      timeText: `${item.startAt.format('HH:mm')} - ${item.endAt.format(
        'HH:mm',
      )}`,
    }
  })

  return assigned.map(item => ({
    ...item,
    displayColumnIndex: item.columnIndex,
    displayColumnCount: peakColumns,
  }))
}

function buildDayLayouts(items = []) {
  const sorted = [...items]
    .map(normalizeScheduleItem)
    .sort((a, b) => a.startAt.valueOf() - b.startAt.valueOf())

  const clusters = []
  let currentCluster = []
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

const layoutsByDate = computed(() => {
  const map = new Map()
  headerSummaries.value.forEach((item) => {
    const list = mockSchedules.value.filter(
      schedule => schedule.dateKey === item.key,
    )
    map.set(item.key, buildDayLayouts(list))
  })
  return map
})

function eventStyle(item) {
  const leftOffset
    = (item.displayColumnIndex || 0) * (scheduleCardMinWidth + scheduleCardGap)
      + scheduleColumnHorizontalInset
  return {
    top: `${minuteOffset(item.startMinutes)}px`,
    height: `${Math.max(
      scheduleCardMinHeight,
      ((item.endMinutes - item.startMinutes) / 60) * hourRowHeight,
    )}px`,
    left: `${leftOffset}px`,
    width: `${scheduleCardMinWidth}px`,
  }
}

function eventClass(item) {
  return {
    'schedule-event': true,
    'schedule-event--unsigned': item.status === 'unsigned',
    'schedule-event--signed': item.status === 'signed',
    'schedule-event--conflict': item.conflict,
    'schedule-event--focused': focusedScheduleId.value === String(item.id),
  }
}

function parseConflictTimeRange(timeText) {
  const matched = String(timeText || '').trim().match(/(\d{2}:\d{2})\s*[~-]\s*(\d{2}:\d{2})/)
  if (!matched)
    return null
  return {
    startTime: matched[1],
    endTime: matched[2],
  }
}

function setFocusedSchedule(id) {
  focusedScheduleId.value = id ? String(id) : ''
  if (focusedScheduleTimer)
    clearTimeout(focusedScheduleTimer)
  if (focusedScheduleId.value) {
    focusedScheduleTimer = window.setTimeout(() => {
      if (focusedScheduleId.value === String(id))
        focusedScheduleId.value = ''
    }, 3000)
  }
}

function closeScheduleConflictModalIfOpen(shouldClose) {
  if (shouldClose)
    scheduleConflictOpen.value = false
}

async function focusScheduleEvent(scheduleId) {
  await nextTick()
  const root = boardScrollRef.value
  const targetId = String(scheduleId || '').trim()
  if (!root || !targetId)
    return false
  const eventNode = Array.from(root.querySelectorAll('[data-schedule-event-id]')).find(
    el => el.getAttribute('data-schedule-event-id') === targetId,
  )
  if (!eventNode)
    return false
  eventNode.scrollIntoView({
    behavior: 'smooth',
    block: 'center',
    inline: 'center',
  })
  setFocusedSchedule(targetId)
  return true
}

function buildConflictJumpLocator(item) {
  const timeRange = parseConflictTimeRange(item?.timeText)
  return {
    date: String(item?.date || '').trim(),
    teacherId: String(item?.teacherId || '').trim(),
    teacherName: String(item?.teacherName || '').trim(),
    startTime: timeRange?.startTime || '',
    endTime: timeRange?.endTime || '',
  }
}

function findScheduleForLocator(locator) {
  const targetDate = String(locator?.date || '').trim()
  const targetTeacherId = String(locator?.teacherId || '').trim()
  const targetTeacherName = String(locator?.teacherName || '').trim()
  const targetStartTime = String(locator?.startTime || '').trim()
  const targetEndTime = String(locator?.endTime || '').trim()
  return mockSchedules.value.find((item) => {
    const raw = item?.raw || {}
    const matchesDate = item.dateKey === targetDate
    const matchesTime = item.startAt?.format('HH:mm') === targetStartTime
      && item.endAt?.format('HH:mm') === targetEndTime
    const rawTeacherId = String(raw.teacherId || '').trim()
    const rawTeacherName = String(raw.teacherName || '').trim()
    const matchesTeacher = (targetTeacherId && rawTeacherId === targetTeacherId)
      || (!targetTeacherId && targetTeacherName && rawTeacherName === targetTeacherName)
      || (!targetTeacherId && !targetTeacherName)
    return matchesDate && matchesTime && matchesTeacher
  }) || null
}

async function flushPendingConflictJump() {
  if (!pendingConflictJump?.locator)
    return

  const pending = pendingConflictJump
  const matched = findScheduleForLocator(pending.locator)
  if (matched?.id) {
    const found = await focusScheduleEvent(matched.id)
    if (found) {
      pendingConflictJump = null
      locatingConflictItemKey.value = ''
      closeScheduleConflictModalIfOpen(pending.closeScheduleConflictModal)
      messageService.success('已定位到冲突课程')
      return
    }
  }

  if (pending.allowAppendTeacherFilter && !pending.teacherFilterExpanded) {
    const teacherId = String(pending.locator?.teacherId || '').trim()
    const teacherName = String(pending.locator?.teacherName || '').trim()
    if (teacherId && !filterTeacherId.value.includes(teacherId)) {
      const nextTeacherIds = [...filterTeacherId.value, teacherId]
      scheduleTeacherOptions.value = mergeFilterOptions(scheduleTeacherOptions.value, [{
        id: teacherId,
        value: teacherName || teacherId,
      }], nextTeacherIds)
      filterTeacherId.value = nextTeacherIds
      pendingConflictJump = {
        ...pending,
        teacherFilterExpanded: true,
      }
      return
    }
  }

  pendingConflictJump = null
  locatingConflictItemKey.value = ''
  messageService.warning('未定位到课程，请检查筛选条件或日期范围')
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

async function openEventConflictDetail(event) {
  if (!event?.conflict)
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
  catch (error) {
    console.error('openEventConflictDetail failed', error)
    messageService.error(error?.response?.data?.message || error?.message || '加载冲突详情失败')
  }
  finally {
    scheduleConflictLoading.value = false
  }
}

async function jumpToConflictSchedule(item) {
  await openConflictLocatingState([
    String(item?.teacherId || '').trim(),
    String(item?.teacherName || '').trim(),
    String(item?.date || '').trim(),
    String(item?.timeText || '').trim(),
  ].join('|'))
  const locator = buildConflictJumpLocator(item)
  if (!locator.date || !locator.startTime || !locator.endTime) {
    locatingConflictItemKey.value = ''
    messageService.warning('当前冲突课程暂不支持定位')
    return
  }

  const teacherFilterMissing = locator.teacherId && !filterTeacherId.value.includes(locator.teacherId)
  const closeScheduleConflictModal = scheduleConflictOpen.value
  let needReload = false
  if (currentTime.value === 'day') {
    if (dayjs(currentDate.value).format('YYYY-MM-DD') !== locator.date) {
      currentDate.value = dayjs(locator.date)
      needReload = true
    }
  }
  else {
    const { startDate, endDate } = queryDateRange.value
    if (locator.date < startDate || locator.date > endDate) {
      currentDate.value = dayjs(locator.date)
      needReload = true
    }
  }

  if (!needReload) {
    const matched = findScheduleForLocator(locator)
    if (matched?.id) {
      const found = await focusScheduleEvent(matched.id)
      if (found) {
        locatingConflictItemKey.value = ''
        closeScheduleConflictModalIfOpen(closeScheduleConflictModal)
        messageService.success('已定位到冲突课程')
        return
      }
    }
  }

  pendingConflictJump = {
    closeScheduleConflictModal,
    locator,
    allowAppendTeacherFilter: Boolean(teacherFilterMissing),
    teacherFilterExpanded: false,
  }
  await loadSchedules()
}

function openScheduleEdit(item) {
  openScheduleDetail(item)
}

function openBatchPlanEdit(schedule, scope = 'batch', payload, action = 'edit') {
  const nextSchedule = payload
    ? {
        ...schedule,
        batchMeta: payload.batchMeta,
        batchNo: payload.batchNo || schedule?.batchNo,
        batchSize: Number(payload.batchSize || schedule?.batchSize || 0) || schedule?.batchSize,
      }
    : schedule
  scheduleBatchPlanAction.value = action
  scheduleBatchPlanEditScope.value = scope
  currentBatchPlanSchedule.value = nextSchedule || null
  scheduleBatchPlanEditOpen.value = true
}

const isCurrentDetailEditable = computed(() => Boolean(currentDetailSchedule.value?.id))

async function handleScheduleDetailDelete(scope = 'current') {
  const schedule = currentDetailSchedule.value
  const scheduleId = String(schedule?.id || '').trim()
  if (!scheduleId) {
    messageService.warning('当前日程缺少删除标识，请刷新后重试')
    return
  }

  let deleteCount = 1
  if (scope === 'future') {
    try {
      deleteCount = await loadTeachingScheduleDeleteTargetCount(schedule, 'future')
    }
    catch (error) {
      console.error('load batch delete count failed', error)
      messageService.error(error?.response?.data?.message || error?.message || '加载待删除日程失败')
      return
    }
  }

  Modal.confirm({
    title: scope === 'future' ? '删除后续全部日程?' : '删除日程?',
    content: scope === 'future'
      ? `后续 ${deleteCount} 个日程将被全部删除，删除后不可恢复，请谨慎操作`
      : '删除后将不可恢复，请谨慎操作',
    okText: '删除',
    cancelText: '取消',
    async onOk() {
      deletingScheduleDetail.value = true
      try {
        const res = await cancelTeachingScheduleScopedApi({
          id: scheduleId,
          scope,
        })
        if (res.code !== 200)
          throw new Error(res.message || '删除日程失败')
        scheduleDetailOpen.value = false
        currentDetailSchedule.value = null
        currentScheduleDetail.value = null
        messageService.success(
          scope === 'future'
            ? `已删除后续 ${deleteCount} 节${isOneToOneSchedule(schedule) ? '1对1' : '班课'}日程`
            : `已删除${isOneToOneSchedule(schedule) ? '1对1' : '班课'}日程`,
        )
        await loadSchedules()
      }
      catch (error) {
        console.error('delete schedule detail failed', error)
        messageService.error(error?.response?.data?.message || error?.message || '删除日程失败')
        throw error
      }
      finally {
        deletingScheduleDetail.value = false
      }
    },
  })
}

function handleScheduleDetailEdit(payload) {
  const schedule = currentDetailSchedule.value
  if (!schedule?.id)
    return
  openBatchPlanEdit(schedule, 'batch', payload)
}

function handleScheduleDetailEditCurrent(payload) {
  const schedule = currentDetailSchedule.value
  if (!schedule?.id)
    return
  openBatchPlanEdit(schedule, 'current', payload)
}

function handleBatchPlanUpdated() {
  scheduleBatchPlanEditOpen.value = false
  scheduleDetailOpen.value = false
  scheduleBatchPlanEditScope.value = 'batch'
  scheduleBatchPlanAction.value = 'edit'
  loadSchedules()
}

function isActiveColumn(dateKey) {
  return dateKey === todayKey.value
}

function isMutedTimeLabel(mark) {
  if (!showCurrentTimeLine.value)
    return false
  return Math.abs(mark - currentTimeMinutes.value) <= 20
}

function isFutureDateKey(dateKey) {
  const current = dayjs(dateKey)
  const today = now.value.startOf('day')
  if (current.isAfter(today, 'day'))
    return true
  if (current.isBefore(today, 'day'))
    return false
  return true
}

function isFutureSlot(dateKey, startMinutes, endMinutes) {
  if (!isFutureDateKey(dateKey))
    return false
  const current = dayjs(dateKey)
  const today = now.value.startOf('day')
  if (current.isAfter(today, 'day'))
    return true
  return endMinutes > currentTimeMinutes.value
}

function hasEventInSlot(dateKey, startMinutes, endMinutes) {
  const list = layoutsByDate.value.get(dateKey) || []
  return list.some(
    event =>
      !(event.endMinutes <= startMinutes || event.startMinutes >= endMinutes),
  )
}

function createSlotStyle(startMinutes, endMinutes) {
  return {
    top: `${minuteOffset(startMinutes)}px`,
    height: `${((endMinutes - startMinutes) / 60) * hourRowHeight}px`,
  }
}

function syncScroll(source, target) {
  if (!source || !target || syncingScroll)
    return
  syncingScroll = true
  target.scrollLeft = source.scrollLeft
  requestAnimationFrame(() => {
    syncingScroll = false
  })
}

function handleHeaderScroll(event) {
  syncScroll(event.target, boardScrollRef.value)
  updateFloatingDatePositions(
    boardScrollRef.value?.scrollLeft ?? event.target.scrollLeft,
  )
}

function handleBoardScroll(event) {
  syncScroll(event.target, headerScrollRef.value)
  updateFloatingDatePositions(event.target.scrollLeft)
}

watch(
  [() => headerScrollRef.value, () => boardScrollRef.value],
  ([headerEl, boardEl]) => {
    if (layoutResizeObserver) {
      layoutResizeObserver.disconnect()
      layoutResizeObserver = null
    }

    updateScheduleViewportWidth()
    if (typeof ResizeObserver === 'undefined') {
      nextTick(() => updateFloatingDatePositions())
      return
    }

    layoutResizeObserver = new ResizeObserver(() => {
      updateScheduleViewportWidth()
      updateFloatingDatePositions()
    })
    if (headerEl)
      layoutResizeObserver.observe(headerEl)
    if (boardEl && boardEl !== headerEl)
      layoutResizeObserver.observe(boardEl)

    nextTick(() => {
      updateScheduleViewportWidth()
      updateFloatingDatePositions()
    })
  },
  { immediate: true },
)

watch(gridTemplateStyle, () => nextTick(() => updateFloatingDatePositions()))
</script>

<template>
  <div>
    <div
      class="filter-wrap bg-white pl-3 pr-3 rounded-4 rounded-lt-0 rounded-rt-0"
    >
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

    <div class="time-page mt2">
      <div class="toolbar-card">
        <div class="toolbar-main">
          <div class="toolbar-group">
            <a-radio-group
              v-model:value="currentTime"
              button-style="solid"
              size="small"
            >
              <a-radio-button
                v-for="opt in timeOptions"
                :key="opt.key"
                :value="opt.key"
              >
                {{ opt.label }}
              </a-radio-button>
            </a-radio-group>
          </div>

          <div class="toolbar-date time-selector ml3 font-800 text-5 flex-center">
            <a-popover trigger="hover">
              <template #content>
                {{ currentTime === 'day' ? '前一天' : '上一周' }}
              </template>
              <span
                class="cursor-pointer text-3 text-#888 flex w6 h6 bg-#eee rounded-10 flex-center font-500 hover-text-#06f hover-bg-#e6f0ff"
                @click="handlePrev"
              >
                <LeftOutlined />
              </span>
            </a-popover>
            <span class="mx-2">
              <div
                class="relative cursor-pointer toolbar-date-range"
                :class="isViewingTodayOrThisWeek ? 'text-#0061ff' : 'text-#222'"
              >
                {{ formatDateRange(currentDate) }}
                <a-date-picker
                  v-if="currentTime === 'day'"
                  v-model:value="currentDate"
                  class="absolute left-0 top-0 right-0 bottom-0 z-10 opacity-0"
                  :allow-clear="false"
                  :bordered="false"
                  :format="formatDateRange"
                  style="cursor: pointer"
                />
                <a-date-picker
                  v-else
                  v-model:value="currentDate"
                  class="absolute left-0 top-0 right-0 bottom-0 z-10 opacity-0"
                  picker="week"
                  :allow-clear="false"
                  :bordered="false"
                  :format="formatDateRange"
                  style="cursor: pointer"
                />
              </div>
            </span>
            <a-popover trigger="hover">
              <template #content>
                {{ currentTime === 'day' ? '后一天' : '下一周' }}
              </template>
              <span
                class="cursor-pointer text-3 text-#888 flex w6 h6 bg-#eee rounded-10 flex-center font-500 hover-text-#06f hover-bg-#e6f0ff"
                @click="handleNext"
              >
                <RightOutlined />
              </span>
            </a-popover>
            <a-popover trigger="hover">
              <template #content>
                {{ currentTime === 'day' ? '回到今天' : '回到本周' }}
              </template>
              <a-button
                type="default"
                size="small"
                class="toolbar-today-week-btn ml2"
                :class="{
                  'toolbar-today-week-btn--active': isViewingTodayOrThisWeek,
                  'toolbar-today-week-btn--inactive': !isViewingTodayOrThisWeek,
                }"
                @click="handleGoThisWeek"
              >
                {{ currentTime === 'day' ? '今天' : '本周' }}
              </a-button>
            </a-popover>
          </div>

          <a-space>
            <CreateSchedulePopover />
            <a-button :loading="exportLoading" @click="exportTimeTimetable">
              导出课表
            </a-button>
          </a-space>
        </div>
      </div>

      <div class="schedule-card" :class="{ 'schedule-card--loading': scheduleLoading }">
        <a-spin
          :spinning="scheduleLoading"
          class="schedule-area-spin"
        >
          <div class="schedule-sticky-shell">
            <div class="schedule-summary">
              <div class="schedule-summary__left">
                <span class="summary-accent" />
                <span>共 {{ mockSchedules.length }} 个日程（未点名
                  {{ unsignedCount }} 个日程）</span>
              </div>
              <div class="schedule-summary__right">
                <span
                  v-for="item in scheduleLegend"
                  :key="item.key"
                  class="legend-item"
                >
                  <span
                    v-if="item.type === 'bar'"
                    class="legend-item__bar"
                    :style="{ background: item.color }"
                  />
                  <span
                    v-else-if="item.type === 'icon'"
                    class="legend-item__icon legend-item__icon--trial"
                  />
                  <span
                    v-else
                    class="legend-item__icon legend-item__icon--danger"
                  />
                  {{ item.label }}
                </span>
              </div>
            </div>

            <div
              ref="headerScrollRef"
              class="schedule-header-scroll"
              @scroll="handleHeaderScroll"
            >
              <div class="schedule-header-track" :style="gridTrackStyle">
                <div class="schedule-floating-date-layer">
                  <div
                    v-for="item in headerSummaries"
                    :key="`float-${item.key}`"
                    class="schedule-floating-chip"
                    :class="{ 'schedule-floating-chip--today': isActiveColumn(item.key) }"
                    :style="floatingPillStyle(item.key)"
                  >
                    <div class="schedule-floating-chip__line">
                      <span class="schedule-floating-chip__date">{{ item.date.format("M/D") }}</span>
                      <span class="schedule-floating-chip__week">{{
                        ["周日", "周一", "周二", "周三", "周四", "周五", "周六"][item.date.day()]
                      }}</span>
                    </div>
                    <div class="schedule-floating-chip__meta">
                      共 <strong>{{ item.count }}</strong> 个
                    </div>
                  </div>
                </div>
                <div class="schedule-header-grid" :style="gridTemplateStyle">
                  <div class="schedule-time-header" />

                  <div
                    v-for="(item, di) in headerSummaries"
                    :key="item.key"
                    class="schedule-column-header"
                    :class="{
                      'schedule-column-header--active': isActiveColumn(item.key),
                      'schedule-column-header--day-divider':
                        di < headerSummaries.length - 1,
                    }"
                  >
                    <div class="schedule-column-header__title schedule-column-header__title--ghost" aria-hidden="true">
                      {{
                        currentTime === "day"
                          ? "当日"
                          : weekdayLabels[
                            item.date.day() === 0 ? 6 : item.date.day() - 1
                          ]
                      }}
                      <span class="schedule-column-header__date">（{{ item.date.format("M-D") }}）</span>
                    </div>
                    <div class="schedule-column-header__count schedule-column-header__count--ghost" aria-hidden="true">
                      {{ item.count }}个
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div
            ref="boardScrollRef"
            class="schedule-board"
            @scroll="handleBoardScroll"
          >
            <div class="schedule-grid" :style="gridTemplateStyle">
              <div class="schedule-time-axis">
                <div
                  v-for="(mark, index) in hourMarks"
                  :key="mark"
                  class="schedule-time-axis__label"
                  :class="{
                    'schedule-time-axis__label--first': index === 0,
                    'schedule-time-axis__label--muted': isMutedTimeLabel(mark),
                  }"
                  :style="{ top: `${minuteOffset(mark)}px` }"
                >
                  <span class="schedule-time-axis__text">{{
                    formatClock(mark)
                  }}</span>
                </div>
                <div
                  v-if="showCurrentTimeLine"
                  class="schedule-now-axis"
                  :style="{ top: `${minuteOffset(currentTimeMinutes)}px` }"
                >
                  <span class="schedule-now-axis__text">{{
                    currentTimeLabel
                  }}</span>
                  <span class="schedule-now-axis__dot" />
                </div>
              </div>

              <div
                v-for="(item, di) in headerSummaries"
                :key="`${item.key}-body`"
                class="schedule-column"
                :class="{
                  'schedule-column--active': isActiveColumn(item.key),
                  'schedule-column--day-divider':
                    di < headerSummaries.length - 1,
                }"
              >
                <div
                  class="schedule-column__body"
                  :style="{ height: `${timelineHeight}px` }"
                >
                  <div
                    v-for="mark in hourMarks"
                    :key="`${item.key}-${mark}`"
                    class="schedule-column__line"
                    :style="{ top: `${minuteOffset(mark)}px` }"
                  />
                  <template
                    v-for="slot in hoverSlots"
                    :key="`${item.key}-${slot.key}`"
                  >
                    <div
                      v-if="
                        isFutureSlot(
                          item.key,
                          slot.startMinutes,
                          slot.endMinutes,
                        )
                          && !hasEventInSlot(
                            item.key,
                            slot.startMinutes,
                            slot.endMinutes,
                          )
                      "
                      class="schedule-create-slot"
                      :style="createSlotStyle(slot.startMinutes, slot.endMinutes)"
                    >
                      <CreateSchedulePopover trigger="click">
                        <button
                          type="button"
                          class="schedule-create-slot__trigger"
                        >
                          点击创建排课日程
                        </button>
                      </CreateSchedulePopover>
                    </div>
                  </template>
                  <div
                    v-if="showCurrentTimeLine"
                    class="schedule-now-line"
                    :style="{ top: `${minuteOffset(currentTimeMinutes)}px` }"
                  />

                  <TimetableScheduleHoverPopover
                    v-for="event in layoutsByDate.get(item.key) || []"
                    :key="event.id"
                    :schedule-id="String(event.id || '')"
                    :batch-no="String(event.raw?.batchNo || '')"
                    :batch-size="Number(event.raw?.batchSize || 0)"
                    :mode-label="scheduleBadgeText(event.classType)"
                    :lesson-title="scheduleHoverTitle(event.raw)"
                    :teacher-name="event.teacher"
                    :course-name="event.course"
                    :assistant-text="scheduleAssistantText(event.raw)"
                    :student-text="scheduleStudentSummary(event.raw)"
                    :classroom-name="event.classroom"
                    :time-text="scheduleTimeTextFromEvent(event)"
                    :conflict-text="scheduleConflictSummary(event.raw)"
                    @detail="openScheduleDetail(event)"
                    @copy="payload => openBatchPlanEdit(event.raw, 'batch', payload, 'copy')"
                    @copy-current="payload => openBatchPlanEdit(event.raw, 'current', payload, 'copy')"
                    @edit="payload => openBatchPlanEdit(event.raw, 'batch', payload)"
                    @edit-current="payload => openBatchPlanEdit(event.raw, 'current', payload)"
                  >
                    <div
                      :class="eventClass(event)"
                      :style="eventStyle(event)"
                      :data-schedule-event-id="String(event.id)"
                      @click="openScheduleEdit(event)"
                    >
                      <div class="schedule-event__top">
                        <div class="schedule-event__time">
                          {{ event.timeText }}
                        </div>
                        <div class="schedule-event__badges">
                          <a-tooltip v-if="event.conflict" :title="conflictBadgeTooltip(event)" placement="top" @click.stop>
                            <span
                              class="schedule-event__badge schedule-event__badge--conflict"
                              @click.stop="openEventConflictDetail(event)"
                            >
                              冲突
                            </span>
                          </a-tooltip>
                          <span
                            v-else
                            class="schedule-event__badge"
                            :class="event.classType === 1 ? 'schedule-event__badge--group-class' : 'schedule-event__badge--one-to-one'"
                          >
                            {{ scheduleBadgeText(event.classType) }}
                          </span>
                          <span v-if="event.hasTrial" class="schedule-event__badge">
                            试听
                          </span>
                        </div>
                      </div>
                      <div class="schedule-event__body">
                        <div class="schedule-event__title">
                          {{ event.title }}
                        </div>
                        <div class="schedule-event__meta schedule-event__meta__course">
                          {{ event.course }}
                        </div>
                        <div
                          class="schedule-event__meta schedule-event__meta--muted"
                        >
                          {{ event.teacher }}
                          <template
                            v-if="event.classroom && event.classroom !== '-'"
                          >
                            · {{ event.classroom }}
                          </template>
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
    </div>
    <SmartTimetableScheduleDetailDrawer
      v-model:open="scheduleDetailOpen"
      :detail="currentScheduleDetail"
      :deleting="deletingScheduleDetail"
      :editable="isCurrentDetailEditable"
      @delete="handleScheduleDetailDelete"
      @delete-current="handleScheduleDetailDelete('current')"
      @delete-future="handleScheduleDetailDelete('future')"
      @copy="payload => openBatchPlanEdit(currentDetailSchedule, 'batch', payload, 'copy')"
      @copy-current="payload => openBatchPlanEdit(currentDetailSchedule, 'current', payload, 'copy')"
      @edit="handleScheduleDetailEdit"
      @edit-current="handleScheduleDetailEditCurrent"
    />
    <ScheduleBatchPlanEditModal
      v-model:open="scheduleBatchPlanEditOpen"
      :schedule="currentBatchPlanSchedule"
      :scope="scheduleBatchPlanEditScope"
      :action="scheduleBatchPlanAction"
      @updated="handleBatchPlanUpdated"
    />
    <ScheduleConflictModal
      v-model:open="scheduleConflictOpen"
      :validation="scheduleConflictValidation"
      :locating="Boolean(locatingConflictItemKey)"
      title="冲突详情"
      current-title="当前冲突日程"
      existing-title="与其冲突的日程"
      fallback-message="当前日程与已有日程存在冲突"
      @jump="jumpToConflictSchedule"
    />
  </div>
</template>

<style scoped lang="less">
.time-page {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.toolbar-card,
.schedule-card {
  border: 1px solid #e5ebf3;
  border-radius: 16px;
  background: #fff;
  box-shadow: 0 10px 24px rgb(15 23 42 / 4%);
}

.toolbar-card {
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
  min-width: 240px;
  text-align: center;
}

.schedule-card {
  overflow: visible;
  border-top: none;
  border-top-left-radius: 0;
  border-top-right-radius: 0;
}

.schedule-card--loading {
  .schedule-sticky-shell,
  .schedule-board {
    pointer-events: none;
  }
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

/* 课表区域轻量 loading（替代 v-loading 黑色半透明蒙层） */
.schedule-area-spin {
  display: block;
  width: 100%;

  :deep(.ant-spin-nested-loading) {
    width: 100%;
  }

  :deep(.ant-spin-container) {
    overflow: visible;
  }

  :deep(.ant-spin-container.ant-spin-blur) {
    pointer-events: none;
  }

  :deep(.ant-spin-blur) {
    opacity: 0.72;
    filter: none;
    -webkit-filter: none;
  }

  :deep(.ant-spin) {
    max-height: none;
    z-index: 60;
  }

  :deep(.ant-spin-dot) {
    font-size: 14px;
  }
}

.schedule-sticky-shell {
  position: sticky;
  top: 8px;
  z-index: 40;
  background: #fff;
  box-shadow: 0 10px 22px rgb(15 23 42 / 6%);
}

.schedule-summary {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 16px 10px;
  border-bottom: 1px solid #edf2f7;
  background: rgb(255 255 255 / 98%);
  backdrop-filter: blur(12px);
}

.schedule-summary__left,
.schedule-summary__right {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
}

.schedule-summary__left {
  color: #1f2937;
  font-size: 13px;
  font-weight: 600;
}

.summary-accent {
  width: 4px;
  height: 16px;
  border-radius: 999px;
  background: #1677ff;
}

.legend-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #4b5563;
  font-size: 12px;
}

.legend-item__bar {
  display: inline-block;
  width: 18px;
  height: 4px;
  border-radius: 999px;
}

.legend-item__icon {
  position: relative;
  display: inline-block;
  width: 12px;
  height: 12px;
  border-radius: 3px;
  background: #fff;
  border: 1px solid #cbd5e1;
}

.legend-item__icon--trial::after {
  position: absolute;
  left: 2px;
  top: 2px;
  width: 6px;
  height: 6px;
  background: #b5bfcf;
  border-radius: 1px;
  content: "";
}

.legend-item__icon--danger {
  border-color: #ff7875;
}

.legend-item__icon--danger::after {
  position: absolute;
  left: 1px;
  right: 1px;
  top: 50%;
  height: 2px;
  background: #ff4d4f;
  transform: translateY(-50%);
  content: "";
}

.schedule-header-scroll {
  overflow-x: auto;
  overflow-y: hidden;
  background: #fff;
}

.schedule-header-track {
  position: relative;
  width: max-content;
  min-width: 100%;
}

.schedule-floating-date-layer {
  position: absolute;
  top: 0;
  left: 0;
  z-index: 4;
  width: 100%;
  height: 48px;
  min-height: 48px;
  pointer-events: none;
}

.schedule-floating-chip {
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

.schedule-floating-chip__line {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  justify-content: center;
  gap: 6px;
  line-height: 1.15;
}

.schedule-floating-chip__date {
  color: #0f172a;
  font-size: 13px;
  font-weight: 700;
  letter-spacing: -0.02em;
}

.schedule-floating-chip__week {
  color: #64748b;
  font-size: 12px;
  font-weight: 600;
}

.schedule-floating-chip__meta {
  margin: 0;
  color: #94a3b8;
  font-size: 11px;
  font-weight: 500;
  line-height: 1.2;
  letter-spacing: 0.02em;
}

.schedule-floating-chip__meta strong {
  color: #64748b;
  font-weight: 700;
}

.schedule-floating-chip--today {
  border-color: #91caff;
  background: #f5f9ff;
  box-shadow:
    0 1px 3px rgb(22 119 255 / 12%),
    inset 0 0 0 1px rgb(255 255 255 / 90%);
}

.schedule-floating-chip--today::before {
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

.schedule-floating-chip--today .schedule-floating-chip__date {
  color: #0958d9;
}

.schedule-floating-chip--today .schedule-floating-chip__week {
  color: #1677ff;
}

.schedule-floating-chip--today .schedule-floating-chip__meta,
.schedule-floating-chip--today .schedule-floating-chip__meta strong {
  color: #1677ff;
}

.schedule-floating-chip--today .schedule-floating-chip__meta {
  opacity: 0.9;
}

.schedule-column-header__title--ghost,
.schedule-column-header__count--ghost {
  visibility: hidden;
}

.schedule-header-grid {
  display: grid;
  width: max-content;
  min-width: 100%;
}

.schedule-board {
  overflow-x: auto;
  overflow-y: visible;
  background: #fff;
}

/* 表头与正文横向滚动条样式一致（两处 scrollLeft 会同步） */
.schedule-header-scroll,
.schedule-board {
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

.schedule-grid {
  display: grid;
  width: max-content;
  min-width: 100%;
  position: relative;
}

.schedule-time-header,
.schedule-column-header {
  height: 48px;
  border-right: 1px solid #dde5f0;
  border-bottom: 1px solid #dde5f0;
  background: #eff4fb;
}

.schedule-time-header {
  position: sticky;
  left: 0;
  z-index: 20;
  background: #fff;
  box-shadow: 4px 0 14px -6px rgb(15 23 42 / 12%);
}

.schedule-column-header {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: #374151;
  font-size: 14px;
  font-weight: 700;
}

.schedule-column-header--active {
  color: #1677ff;
  box-shadow: inset 0 3px 0 #1677ff;
}

.schedule-column-header--day-divider {
  border-right-width: 2px;
  border-right-color: #a8b8cc;
}

.schedule-column-header__date {
  color: #6b7280;
  font-weight: 600;
}

.schedule-column-header__count {
  color: #6b7280;
  font-size: 13px;
  font-weight: 600;
}

.schedule-time-axis {
  position: sticky;
  left: 0;
  z-index: 20;
  border-right: 1px solid #dde5f0;
  background: #fff;
  box-shadow: 4px 0 14px -6px rgb(15 23 42 / 12%);
}

.schedule-time-axis__label {
  position: absolute;
  left: 0;
  right: 0;
  transform: translateY(-50%);
  color: #1f2937;
  font-size: 14px;
  text-align: center;
  z-index: 6;
}

.schedule-time-axis__label::before {
  position: absolute;
  left: 0;
  right: 0;
  top: 50%;
  border-top: 1px solid #dde5f0;
  content: "";
  z-index: 0;
}

.schedule-time-axis__label--first {
  transform: translateY(-50%);
}

.schedule-time-axis__label--muted .schedule-time-axis__text {
  color: rgb(31 41 55 / 22%);
}

.schedule-time-axis__text {
  position: relative;
  z-index: 1;
  display: inline-block;
  padding: 0 10px;
  background: #fff;
}

.schedule-column {
  position: relative;
  border-right: 1px solid #dde5f0;
  background: #fff;
}

.schedule-column--active {
  background: #f3f9ff;
}

.schedule-column--day-divider {
  border-right-width: 2px;
  border-right-color: #a8b8cc;
}

.schedule-column__body {
  position: relative;
}

.schedule-column__line {
  position: absolute;
  left: 0;
  right: 0;
  border-top: 1px solid #dde5f0;
}

.schedule-now-axis {
  position: absolute;
  z-index: 8;
  pointer-events: none;
  left: 0;
  right: 0;
}

.schedule-now-axis__text {
  position: absolute;
  top: -7px;
  left: -3px;
  right: -4px;
  color: #ff4d4f;
  font-size: 12px;
  font-weight: 600;
  line-height: 1.2;
  text-align: center;
}

.schedule-now-axis__dot {
  position: absolute;
  top: -3px;
  left: 81px;
  width: 6px;
  height: 6px;
  border-radius: 999px;
  background: #ff4d4f;
}

.schedule-now-line {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  border-top: 1px solid #ffb3b3;
  z-index: 1;
  pointer-events: none;
}

.schedule-create-slot {
  position: absolute;
  left: 0;
  right: 0;
  z-index: 1;
}

.schedule-create-slot__trigger {
  width: calc(100% - 12px);
  height: calc(100% - 12px);
  margin: 6px;
  border: 1px dashed transparent;
  border-radius: 8px;
  background: transparent;
  color: transparent;
  font: inherit;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: color 0.18s ease, border-color 0.18s ease,
    background-color 0.18s ease;
}

.schedule-create-slot__trigger:hover {
  border-color: #cfe0ff;
  background: rgb(24 119 255 / 4%);
  color: #1677ff;
}

.schedule-event {
  position: absolute;
  z-index: 2;
  display: flex;
  flex-direction: column;
  border: none;
  border-radius: 4px;
  background: #ffffff;
  overflow: hidden;
  box-shadow: 0 6px 16px rgb(22 119 255 / 10%);
}

.schedule-event--unsigned {
  background: #ffffff;
}

.schedule-event--signed {
  background: #f5f7fa;
}

.schedule-event--conflict {
  box-shadow: 0 0 0 2px rgba(255, 77, 79, 0.4);
}

.schedule-event--focused {
  box-shadow: 0 0 0 3px rgba(22, 119, 255, 0.45), 0 14px 28px rgba(22, 119, 255, 0.2);
}

.schedule-event__top {
  position: relative;
  display: flex;
  align-items: center;
  min-height: 24px;
  padding: 3px 56px 3px 10px;
  background: #1677ff;
}

.schedule-event--signed .schedule-event__top {
  background: #98a2b3;
}

.schedule-event__time {
  flex: 1;
  min-width: 0;
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  line-height: 1.2;
  white-space: nowrap;
  letter-spacing: 0.01em;
}

.schedule-event__badges {
  position: absolute;
  top: 0;
  right: 0;
  display: flex;
  flex-direction: row-reverse;
  align-items: flex-start;
  gap: 4px;
  flex-shrink: 0;
}

.schedule-event__badge {
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
  border: none;
}

.schedule-event__badge--one-to-one {
  padding: 0 8px 0 9px;
  border-radius: 0 4px 0 8px;
  background: rgb(0 0 0 / 50%);
  color: #fff;
}

.schedule-event__badge--group-class {
  padding: 0 8px 0 9px;
  border-radius: 0 4px 0 8px;
  background: #d46b08;
  color: #fff;
}

.schedule-event__badge--conflict {
  padding: 0 8px 0 9px;
  border-radius: 0 4px 0 8px;
  background: #ff4d4f;
  color: #fff;
  cursor: pointer;
}

.schedule-event__body {
  display: flex;
  flex-direction: column;
  padding: 4px 0 0 10px;
}

.schedule-event__title {
  color: #0f172a;
  font-size: 13px;
  font-weight: 700;
  line-height: 1;
  letter-spacing: 0.01em;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  overflow: hidden;
  -webkit-line-clamp: 2;
}

.schedule-event__meta {
  color: #64748b;
  font-size: 12px;
  line-height: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.schedule-event__meta__course{
  margin: 4px 0;
}

.schedule-event__meta--muted {
  color: #334155;
  font-weight: 600;
}

.time-selector {
  font-family: DINAlternate-Bold, DINAlternate;
}

.toolbar-card :deep(.ant-radio-button-wrapper) {
  padding: 0 16px;
}

@media (max-width: 1200px) {
  .toolbar-main {
    flex-wrap: wrap;
    justify-content: flex-start;
  }
}

@media (max-width: 768px) {
  .schedule-summary {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
