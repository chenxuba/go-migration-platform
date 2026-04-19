<script setup lang="ts">
import dayjs, { type Dayjs } from 'dayjs'
import { computed, ref, watch } from 'vue'
import scheduleClassImage from '@/assets/images/timetable/schedule-class.png'
import scheduleOneToOneImage from '@/assets/images/timetable/schedule-one2one.png'
import {
  exportClassCommentWordApi,
  getClassCommentStudentPagedListApi,
  type ClassCommentStudentItem,
} from '@/api/edu-center/class-record'
import { pageGroupClassesApi } from '@/api/edu-center/group-class'
import { getOneToOneListApi } from '@/api/edu-center/one-to-one'
import { getCourseIdAndNameApi } from '@/api/edu-center/registr-renewal'
import { getUserListApi } from '@/api/internal-manage/staff-manage'
import RehabRecordEditorDrawer from '@/components/home-center/class-comment/rehab-record-editor-drawer.vue'
import messageService from '@/utils/messageService'

interface FilterOption {
  id: string
  value: string
  mobile?: string
}

const props = withDefaults(defineProps<{
  studentId?: string | number
  open?: boolean
  active?: boolean
}>(), {
  studentId: '',
  open: false,
  active: false,
})

function createDefaultDateRange(): [Dayjs, Dayjs] {
  const start = dayjs().startOf('month')
  const end = dayjs()
  return [start, end]
}

function createDefaultScheduleDateVals() {
  const [start, end] = createDefaultDateRange()
  return [start.format('YYYY-MM-DD'), end.format('YYYY-MM-DD')]
}

const displayArray = ref([
  'scheduleCourse',
  'scheduleTeacher',
  'assistantTeacher',
  'scheduleClass',
  'scheduleOneToOne',
  'scheduleDate',
  'scheduleType',
  'commentStatus',
  'readStatus',
  'parentFeedbackStatus',
])

const scheduleTypeOptions = [
  { id: '1', value: '班级' },
  { id: '2', value: '1对1' },
]

const commentStatusOptions = [
  { id: '1', value: '已记录' },
  { id: '0', value: '未记录' },
]

const readStatusOptions = [
  { id: '1', value: '已读' },
  { id: '0', value: '未读' },
]

const parentFeedbackStatusOptions = [
  { id: '1', value: '已反馈' },
  { id: '0', value: '未反馈' },
]

const currentStudentId = computed(() => String(props.studentId || '').trim())
const defaultScheduleDateVals = ref<string[]>(createDefaultScheduleDateVals())
const filterRenderKey = ref(0)
const loading = ref(false)
const exportingWord = ref(false)
const dataSource = ref<ClassCommentStudentItem[]>([])
const drawerOpen = ref(false)
const currentActionRecord = ref<Partial<ClassCommentStudentItem> | null>(null)
const drawerMode = ref<'create' | 'view' | 'edit'>('view')
const sortStartTime = ref(2)

const filterDateRange = ref<[Dayjs, Dayjs]>(createDefaultDateRange())
const filterLessonId = ref<string | undefined>(undefined)
const filterTeacherIds = ref<string[]>([])
const filterAssistantTeacherIds = ref<string[]>([])
const filterClassId = ref<string | undefined>(undefined)
const filterOneToOneId = ref<string | undefined>(undefined)
const filterScheduleTypes = ref<string[]>([])
const filterCommentStatus = ref<boolean | undefined>(true)
const filterReadStatus = ref<boolean | undefined>(undefined)
const filterParentFeedbackStatus = ref<boolean | undefined>(undefined)

const pagination = ref({
  current: 1,
  pageSize: 50,
  total: 0,
})

const courseOptions = ref<FilterOption[]>([])
const courseFinished = ref(false)
const classOptions = ref<FilterOption[]>([])
const classFinished = ref(false)
const oneToOneOptions = ref<FilterOption[]>([])
const oneToOneFinished = ref(false)
const teacherOptions = ref<FilterOption[]>([])
const teacherFinished = ref(false)
const assistantTeacherOptions = ref<FilterOption[]>([])
const assistantTeacherFinished = ref(false)

const classPagination = ref({ current: 1, pageSize: 20, total: 0 })
const oneToOnePagination = ref({ current: 1, pageSize: 20, total: 0 })
const teacherPagination = ref({ current: 1, pageSize: 20, total: 0 })
const assistantTeacherPagination = ref({ current: 1, pageSize: 20, total: 0 })

const classSearchKey = ref('')
const oneToOneSearchKey = ref('')
const teacherSearchKey = ref('')
const assistantTeacherSearchKey = ref('')

const allColumns = ref<any[]>([
  {
    title: '上课日期/时段',
    dataIndex: 'classDateTime',
    key: 'classDateTime',
    fixed: 'left',
    width: 160,
    sorter: true,
    defaultSortOrder: 'descend',
  },
  {
    title: '类型',
    dataIndex: 'type',
    key: 'type',
    width: 120,
  },
  {
    title: '所属班级/1对1',
    key: 'linkClassOr1v1',
    dataIndex: 'linkClassOr1v1',
    width: 220,
  },
  {
    title: '所属课程',
    key: 'linkCourse',
    dataIndex: 'linkCourse',
    width: 160,
  },
  {
    title: '上课教师',
    key: 'teacher',
    dataIndex: 'teacher',
    width: 130,
  },
  {
    title: '是否记录',
    dataIndex: 'isComment',
    key: 'isComment',
    width: 100,
  },
  {
    title: '已读/未读',
    dataIndex: 'readStatus',
    key: 'readStatus',
    width: 120,
  },
  {
    title: '家长反馈',
    dataIndex: 'parentFeedback',
    key: 'parentFeedback',
    width: 120,
  },
  {
    title: '上课助教',
    dataIndex: 'subTeacher',
    key: 'subTeacher',
    width: 150,
  },
  {
    title: '上课教室',
    dataIndex: 'classRoom',
    key: 'classRoom',
    width: 130,
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    fixed: 'right',
    width: 140,
  },
])

const totalWidth = computed(() =>
  allColumns.value.reduce((acc, column) => acc + Number(column.width || 0), 0),
)

const tablePagination = computed(() => ({
  current: pagination.value.current,
  pageSize: pagination.value.pageSize,
  total: pagination.value.total,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: (total: number) => `共 ${total} 条`,
}))

const currentViewingStudent = computed(() => {
  if (!currentActionRecord.value)
    return null
  return {
    id: currentActionRecord.value.studentId,
    name: currentActionRecord.value.studentName,
    avatar: currentActionRecord.value.avatar,
  }
})

const currentViewingSession = computed(() => {
  if (!currentActionRecord.value)
    return null
  return {
    sourceName: currentActionRecord.value.sourceName,
    lessonName: currentActionRecord.value.lessonName,
    teacherName: currentActionRecord.value.teacherName,
    classRoomName: currentActionRecord.value.classRoomName,
    startTime: currentActionRecord.value.startTime,
    endTime: currentActionRecord.value.endTime,
  }
})

function normalizeFilterValue(value: unknown) {
  if (Array.isArray(value))
    return value.length ? String(value[0] ?? '').trim() || undefined : undefined
  const text = String(value ?? '').trim()
  return text || undefined
}

function normalizeFilterValues(value: unknown) {
  if (!Array.isArray(value))
    return []
  return value.map(item => String(item ?? '').trim()).filter(Boolean)
}

function normalizeBooleanFilter(value: unknown) {
  const normalized = normalizeFilterValue(value)
  if (normalized === '1')
    return true
  if (normalized === '0')
    return false
  return undefined
}

function mergeFilterOptions(previous: FilterOption[], incoming: FilterOption[], selectedValues: string | string[] | undefined = []) {
  const selectedSet = new Set((Array.isArray(selectedValues) ? selectedValues : [selectedValues]).map(value => String(value || '')).filter(Boolean))
  const map = new Map<string, FilterOption>()
  previous.forEach((item) => {
    if (selectedSet.has(item.id))
      map.set(item.id, item)
  })
  incoming.forEach((item) => {
    if (item.id)
      map.set(item.id, item)
  })
  return [...map.values()]
}

function formatDateTime(record: Partial<ClassCommentStudentItem>) {
  const start = dayjs(record.startTime)
  const end = dayjs(record.endTime)
  if (!start.isValid() || !end.isValid()) {
    return {
      dateText: '-',
      timeText: '--:--～--:--',
    }
  }
  const weekday = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'][start.day()] || ''
  return {
    dateText: `${start.format('YYYY-MM-DD')}（${weekday}）`,
    timeText: `${start.format('HH:mm')}～${end.format('HH:mm')}`,
  }
}

function sourceTypeText(record: Partial<ClassCommentStudentItem>) {
  const sourceType = Number(record.sourceType || 0)
  if (sourceType === 2)
    return '1对1'
  if (sourceType === 3)
    return '试听'
  return '班级'
}

function sourceTypeImage(record: Partial<ClassCommentStudentItem>) {
  const sourceType = Number(record.sourceType || 0)
  if (sourceType === 2)
    return scheduleOneToOneImage
  if (sourceType === 1)
    return scheduleClassImage
  return ''
}

function displayText(value?: string | null) {
  const text = String(value ?? '').trim()
  return text || '-'
}

function readStatusText(record: Partial<ClassCommentStudentItem>) {
  if (record.isRead === true)
    return '已读'
  if (record.isRead === false)
    return '未读'
  return '-'
}

function parentFeedbackText(record: Partial<ClassCommentStudentItem>) {
  return record.isParentFeedback ? '已反馈' : '-'
}

function commentStatusText(record: Partial<ClassCommentStudentItem>) {
  return record.isComment ? '已记录' : '未记录'
}

function buildQueryModel() {
  return {
    isParentFeedback: filterParentFeedbackStatus.value,
    teachingStartTime: filterDateRange.value?.[0]?.format('YYYY-MM-DD'),
    teachingEndTime: filterDateRange.value?.[1]?.format('YYYY-MM-DD'),
    teachingRecordTypes: filterScheduleTypes.value.length
      ? filterScheduleTypes.value.map(item => Number(item)).filter(Boolean)
      : [1, 2],
    lessonId: filterLessonId.value,
    teacherIds: filterTeacherIds.value.length ? filterTeacherIds.value : undefined,
    assistantTeacherIds: filterAssistantTeacherIds.value.length ? filterAssistantTeacherIds.value : undefined,
    classId: filterClassId.value,
    one2OneId: filterOneToOneId.value,
    isComment: filterCommentStatus.value,
    isRead: filterReadStatus.value,
    classTeacherIds: [],
    one2OneTeacherIds: [],
    studentId: currentStudentId.value || undefined,
  }
}

function parseAttachmentFilenameFromHeader(headerValue?: string) {
  const header = String(headerValue || '')
  if (!header)
    return ''
  const utf8Match = header.match(/filename\*=UTF-8''([^;]+)/i)
  if (utf8Match?.[1]) {
    try {
      return decodeURIComponent(utf8Match[1])
    }
    catch {
      return utf8Match[1]
    }
  }
  const plainMatch = header.match(/filename="?([^";]+)"?/i)
  return plainMatch?.[1] || ''
}

function resetListData() {
  dataSource.value = []
  pagination.value.current = 1
  pagination.value.total = 0
}

function resetFilters() {
  defaultScheduleDateVals.value = createDefaultScheduleDateVals()
  filterDateRange.value = createDefaultDateRange()
  filterLessonId.value = undefined
  filterTeacherIds.value = []
  filterAssistantTeacherIds.value = []
  filterClassId.value = undefined
  filterOneToOneId.value = undefined
  filterScheduleTypes.value = []
  filterCommentStatus.value = true
  filterReadStatus.value = undefined
  filterParentFeedbackStatus.value = undefined
  sortStartTime.value = 2
  pagination.value.current = 1
  filterRenderKey.value += 1
}

async function loadCourseOptions(searchKey = '') {
  try {
    const res = await getCourseIdAndNameApi({
      searchKey: searchKey || '',
    })
    if (res.code !== 200)
      return
    const resultData = (Array.isArray(res.result) ? res.result : []).map(item => ({
      id: String(item.id ?? ''),
      value: String(item.name || item.value || item.id || '').trim(),
    })).filter(item => item.id && item.value)
    courseOptions.value = mergeFilterOptions(courseOptions.value, resultData, filterLessonId.value)
    courseFinished.value = true
  }
  catch (error) {
    console.error('load student rehab record courses failed', error)
  }
}

async function loadClassOptions(searchKey = '', reset = true) {
  if (reset) {
    classPagination.value.current = 1
    classFinished.value = false
  }
  classSearchKey.value = searchKey
  try {
    const res = await pageGroupClassesApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: classPagination.value.pageSize,
        pageIndex: classPagination.value.current,
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
    classOptions.value = reset
      ? mergeFilterOptions(classOptions.value, resultData, filterClassId.value)
      : mergeFilterOptions(classOptions.value, [...classOptions.value, ...resultData], filterClassId.value)
    classPagination.value.total = Number(res.result?.total || resultData.length || 0)
    classFinished.value = classOptions.value.length >= classPagination.value.total
  }
  catch (error) {
    console.error('load student rehab record classes failed', error)
  }
}

async function loadOneToOneOptions(searchKey = '', reset = true) {
  if (reset) {
    oneToOnePagination.value.current = 1
    oneToOneFinished.value = false
  }
  oneToOneSearchKey.value = searchKey
  try {
    const res = await getOneToOneListApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: oneToOnePagination.value.pageSize,
        pageIndex: oneToOnePagination.value.current,
        skipCount: 0,
      },
      queryModel: {
        searchKey: searchKey || undefined,
      },
    })
    if (res.code !== 200)
      return
    const list = Array.isArray(res.result?.list) ? res.result.list : []
    const resultData = list.map(item => ({
      id: String(item.id ?? ''),
      value: String(item.name || item.studentName || item.id || '').trim(),
    })).filter(item => item.id && item.value)
    oneToOneOptions.value = reset
      ? mergeFilterOptions(oneToOneOptions.value, resultData, filterOneToOneId.value)
      : mergeFilterOptions(oneToOneOptions.value, [...oneToOneOptions.value, ...resultData], filterOneToOneId.value)
    oneToOnePagination.value.total = Number(res.result?.total || resultData.length || 0)
    oneToOneFinished.value = oneToOneOptions.value.length >= oneToOnePagination.value.total
  }
  catch (error) {
    console.error('load student rehab record one to one failed', error)
  }
}

async function loadTeacherOptions(searchKey = '', reset = true) {
  if (reset) {
    teacherPagination.value.current = 1
    teacherFinished.value = false
  }
  teacherSearchKey.value = searchKey
  try {
    const res = await getUserListApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: teacherPagination.value.pageSize,
        pageIndex: teacherPagination.value.current,
        skipCount: 0,
      },
      queryModel: {
        searchKey: searchKey || undefined,
      },
      sortModel: {},
    })
    if (res.code !== 200)
      return
    const resultData = (Array.isArray(res.result) ? res.result : []).map(item => ({
      id: String(item.id ?? ''),
      value: String(item.nickName || item.name || item.value || item.id || '').trim(),
      mobile: String(item.mobile || item.phone || '').trim() || undefined,
    })).filter(item => item.id && item.value)
    teacherOptions.value = reset
      ? mergeFilterOptions(teacherOptions.value, resultData, filterTeacherIds.value)
      : mergeFilterOptions(teacherOptions.value, [...teacherOptions.value, ...resultData], filterTeacherIds.value)
    teacherPagination.value.total = Number(res.total || resultData.length || 0)
    teacherFinished.value = teacherOptions.value.length >= teacherPagination.value.total
  }
  catch (error) {
    console.error('load student rehab record teachers failed', error)
  }
}

async function loadAssistantTeacherOptions(searchKey = '', reset = true) {
  if (reset) {
    assistantTeacherPagination.value.current = 1
    assistantTeacherFinished.value = false
  }
  assistantTeacherSearchKey.value = searchKey
  try {
    const res = await getUserListApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: assistantTeacherPagination.value.pageSize,
        pageIndex: assistantTeacherPagination.value.current,
        skipCount: 0,
      },
      queryModel: {
        searchKey: searchKey || undefined,
      },
      sortModel: {},
    })
    if (res.code !== 200)
      return
    const resultData = (Array.isArray(res.result) ? res.result : []).map(item => ({
      id: String(item.id ?? ''),
      value: String(item.nickName || item.name || item.value || item.id || '').trim(),
      mobile: String(item.mobile || item.phone || '').trim() || undefined,
    })).filter(item => item.id && item.value)
    assistantTeacherOptions.value = reset
      ? mergeFilterOptions(assistantTeacherOptions.value, resultData, filterAssistantTeacherIds.value)
      : mergeFilterOptions(assistantTeacherOptions.value, [...assistantTeacherOptions.value, ...resultData], filterAssistantTeacherIds.value)
    assistantTeacherPagination.value.total = Number(res.total || resultData.length || 0)
    assistantTeacherFinished.value = assistantTeacherOptions.value.length >= assistantTeacherPagination.value.total
  }
  catch (error) {
    console.error('load student rehab record assistants failed', error)
  }
}

async function loadList() {
  if (!props.open || !props.active)
    return
  if (!currentStudentId.value) {
    resetListData()
    return
  }

  loading.value = true
  try {
    const res = await getClassCommentStudentPagedListApi({
      queryModel: buildQueryModel(),
      pageRequestModel: {
        needTotal: true,
        pageIndex: pagination.value.current,
        pageSize: pagination.value.pageSize,
        skipCount: (pagination.value.current - 1) * pagination.value.pageSize,
      },
      sortModel: {
        startTime: sortStartTime.value,
      },
    })
    if (res.code !== 200)
      throw new Error(res.message || '获取康复记录失败')
    dataSource.value = Array.isArray(res.result?.list) ? res.result.list : []
    pagination.value.total = Number(res.result?.total || 0)
  }
  catch (error: any) {
    console.error('load student rehab record list failed', error)
    resetListData()
    messageService.error(error?.response?.data?.message || error?.message || '获取康复记录失败')
  }
  finally {
    loading.value = false
  }
}

async function handleExportWord() {
  if (exportingWord.value)
    return
  if (!currentStudentId.value) {
    messageService.warning('当前学员信息缺失，暂不可导出')
    return
  }
  if (!filterDateRange.value || filterDateRange.value.length < 2) {
    messageService.warning('导出前请选择上课日期')
    return
  }
  const [start, end] = filterDateRange.value
  if (start.startOf('day').isAfter(end.startOf('day'))) {
    messageService.warning('上课日期筛选范围不正确')
    return
  }
  if (end.startOf('day').isAfter(start.startOf('day').add(1, 'month'))) {
    messageService.warning('导出时间范围最大支持一个月')
    return
  }
  if (pagination.value.total <= 0) {
    messageService.warning('暂无可导出的康复记录')
    return
  }

  exportingWord.value = true
  try {
    const res = await exportClassCommentWordApi({
      queryModel: buildQueryModel(),
      sortModel: {
        startTime: sortStartTime.value,
      },
    })
    const contentType = String(res.headers['content-type'] || '')
    if (contentType.includes('application/json')) {
      const text = await res.data.text()
      try {
        const payload = JSON.parse(text)
        messageService.error(payload?.message || '导出失败')
      }
      catch {
        messageService.error('导出失败')
      }
      return
    }
    const blob = new Blob([res.data], {
      type: contentType || 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
    })
    const filename = parseAttachmentFilenameFromHeader(res.headers['content-disposition'])
      || `康复记录-${dayjs().format('YYYYMMDDHHmmss')}.docx`
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = filename
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
    messageService.success('导出成功')
  }
  catch (error: any) {
    console.error('export student rehab records failed', error)
    const blobText = await error?.response?.data?.text?.()
    if (blobText) {
      try {
        const payload = JSON.parse(blobText)
        messageService.error(payload?.message || '导出失败')
        return
      }
      catch {
      }
    }
    messageService.error(error?.message || '导出失败')
  }
  finally {
    exportingWord.value = false
  }
}

function handleScheduleDateFilter(value: unknown) {
  if (!Array.isArray(value) || value.length < 2)
    return
  const start = dayjs(String(value[0] || ''))
  const end = dayjs(String(value[1] || ''))
  if (!start.isValid() || !end.isValid())
    return
  if (
    filterDateRange.value?.[0]?.format('YYYY-MM-DD') === start.format('YYYY-MM-DD')
    && filterDateRange.value?.[1]?.format('YYYY-MM-DD') === end.format('YYYY-MM-DD')
  ) {
    return
  }
  filterDateRange.value = [start, end]
}

function handleLessonFilter(value: unknown) {
  filterLessonId.value = normalizeFilterValue(value)
}

function handleTeacherFilter(value: unknown) {
  filterTeacherIds.value = normalizeFilterValues(value)
}

function handleAssistantTeacherFilter(value: unknown) {
  filterAssistantTeacherIds.value = normalizeFilterValues(value)
}

function handleClassFilter(value: unknown) {
  filterClassId.value = normalizeFilterValue(value)
}

function handleOneToOneFilter(value: unknown) {
  filterOneToOneId.value = normalizeFilterValue(value)
}

function handleScheduleTypeFilter(value: unknown) {
  filterScheduleTypes.value = normalizeFilterValues(value)
}

function handleCommentStatusFilter(value: unknown) {
  filterCommentStatus.value = normalizeBooleanFilter(value)
}

function handleReadStatusFilter(value: unknown) {
  filterReadStatus.value = normalizeBooleanFilter(value)
}

function handleParentFeedbackStatusFilter(value: unknown) {
  filterParentFeedbackStatus.value = normalizeBooleanFilter(value)
}

function handleTableChange(page: { current?: number, pageSize?: number }, _filters: any, sorter: any) {
  pagination.value.current = Number(page?.current || 1)
  pagination.value.pageSize = Number(page?.pageSize || pagination.value.pageSize)
  if (sorter?.order === 'ascend')
    sortStartTime.value = 1
  else
    sortStartTime.value = 2
  loadList()
}

function handleViewRecord(record?: Partial<ClassCommentStudentItem>) {
  drawerMode.value = 'view'
  currentActionRecord.value = record ? { ...record } : null
  drawerOpen.value = true
}

function handleCommentRecord(record?: Partial<ClassCommentStudentItem>) {
  drawerMode.value = 'create'
  currentActionRecord.value = record ? { ...record } : null
  drawerOpen.value = true
}

watch(
  () => `${props.open}|${currentStudentId.value}`,
  (value, previousValue) => {
    if (!props.open) {
      drawerOpen.value = false
      currentActionRecord.value = null
      drawerMode.value = 'view'
      resetListData()
      resetFilters()
      return
    }
    if (value !== previousValue && currentStudentId.value) {
      resetFilters()
      resetListData()
    }
  },
  { immediate: true },
)

watch(
  [
    () => props.open,
    () => props.active,
    currentStudentId,
    filterDateRange,
    filterLessonId,
    filterTeacherIds,
    filterAssistantTeacherIds,
    filterClassId,
    filterOneToOneId,
    filterScheduleTypes,
    filterCommentStatus,
    filterReadStatus,
    filterParentFeedbackStatus,
  ],
  () => {
    if (!props.open || !props.active || !currentStudentId.value)
      return
    pagination.value.current = 1
    loadList()
  },
  { deep: true, immediate: true },
)
</script>

<template>
  <div>
    <div class="filter-wrap bg-white pl-3 pr-3 rounded-lb-4 rounded-rb-4">
      <all-filter
        :key="filterRenderKey"
        :display-array="displayArray"
        :default-schedule-date-vals="defaultScheduleDateVals"
        :schedule-date-disable-future="true"
        :schedule-course-options="courseOptions"
        :schedule-course-finished="courseFinished"
        :on-schedule-course-dropdown-visible-change="() => loadCourseOptions()"
        :on-schedule-course-search="loadCourseOptions"
        :schedule-teacher-options="teacherOptions"
        :schedule-teacher-finished="teacherFinished"
        :on-schedule-teacher-dropdown-visible-change="() => loadTeacherOptions('', true)"
        :on-schedule-teacher-search="keyword => loadTeacherOptions(keyword, true)"
        :on-schedule-teacher-load-more="() => { teacherPagination.current += 1; return loadTeacherOptions(teacherSearchKey, false) }"
        :assistant-teacher-options="assistantTeacherOptions"
        :assistant-teacher-finished="assistantTeacherFinished"
        :on-assistant-teacher-dropdown-visible-change="() => loadAssistantTeacherOptions('', true)"
        :on-assistant-teacher-search="keyword => loadAssistantTeacherOptions(keyword, true)"
        :on-assistant-teacher-load-more="() => { assistantTeacherPagination.current += 1; return loadAssistantTeacherOptions(assistantTeacherSearchKey, false) }"
        :schedule-class-options="classOptions"
        :schedule-class-finished="classFinished"
        :on-schedule-class-dropdown-visible-change="() => loadClassOptions('', true)"
        :on-schedule-class-search="keyword => loadClassOptions(keyword, true)"
        :on-schedule-class-load-more="() => { classPagination.current += 1; return loadClassOptions(classSearchKey, false) }"
        :schedule-one-to-one-options="oneToOneOptions"
        :schedule-one-to-one-finished="oneToOneFinished"
        :on-schedule-one-to-one-dropdown-visible-change="() => loadOneToOneOptions('', true)"
        :on-schedule-one-to-one-search="keyword => loadOneToOneOptions(keyword, true)"
        :on-schedule-one-to-one-load-more="() => { oneToOnePagination.current += 1; return loadOneToOneOptions(oneToOneSearchKey, false) }"
        :schedule-course-label="'课程'"
        :schedule-class-label="'班级'"
        :schedule-one-to-one-label="'1对1'"
        :schedule-type-label="'类型'"
        :schedule-type-options="scheduleTypeOptions"
        :comment-status-label="'是否记录'"
        :comment-status-options="commentStatusOptions"
        :default-comment-status-val="'1'"
        :read-status-label="'已读/未读'"
        :read-status-options="readStatusOptions"
        :parent-feedback-status-label="'家长反馈'"
        :parent-feedback-status-options="parentFeedbackStatusOptions"
        :is-quick-show="false"
        @update:schedule-date-filter="handleScheduleDateFilter"
        @update:schedule-course-filter="handleLessonFilter"
        @update:schedule-teacher-filter="handleTeacherFilter"
        @update:assistant-teacher-filter="handleAssistantTeacherFilter"
        @update:schedule-class-filter="handleClassFilter"
        @update:schedule-one-to-one-filter="handleOneToOneFilter"
        @update:schedule-type-filter="handleScheduleTypeFilter"
        @update:comment-status-filter="handleCommentStatusFilter"
        @update:read-status-filter="handleReadStatusFilter"
        @update:parent-feedback-status-filter="handleParentFeedbackStatusFilter"
      />
    </div>
    <div class="student-list mt-3 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            共 {{ pagination.total }} 条数据
          </div>
          <div class="edit flex">
            <a-tooltip>
              <template #title>
                请先筛选上课日期后再导出
              </template>
              <a-button class="export-word-btn" :loading="exportingWord" @click="handleExportWord">
                导出 Word
              </a-button>
            </a-tooltip>
          </div>
        </div>
        <div class="table-content mt-2">
          <a-table
            row-key="studentTeachingRecordId"
            :loading="loading"
            :data-source="dataSource"
            :pagination="tablePagination"
            :columns="allColumns"
            :scroll="{ x: totalWidth }"
            size="small"
            @change="handleTableChange"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'classDateTime'">
                <div class="text-#222">
                  {{ formatDateTime(record).dateText }}
                </div>
                <div class="text-#222">
                  {{ formatDateTime(record).timeText }}
                </div>
              </template>
              <template v-else-if="column.key === 'type'">
                <div class="justify-between flex-center">
                  <span>{{ sourceTypeText(record) }}</span>
                  <img
                    v-if="sourceTypeImage(record)"
                    class="type-tag-image"
                    height="45"
                    :src="sourceTypeImage(record)"
                    alt=""
                  >
                </div>
              </template>
              <template v-else-if="column.key === 'linkClassOr1v1'">
                <div class="text-#222">
                  {{ record.sourceName || '-' }}
                </div>
              </template>
              <template v-else-if="column.key === 'linkCourse'">
                <div class="text-#222">
                  {{ record.lessonName || '-' }}
                </div>
              </template>
              <template v-else-if="column.key === 'teacher'">
                <div class="text-#222">
                  {{ record.teacherName || '-' }}
                </div>
              </template>
              <template v-else-if="column.key === 'isComment'">
                <span
                  :class="record.isComment ? 'bg-#e6f4ff text-#1677ff' : 'bg-#fff7e6 text-#fa8c16'"
                  class="text-3 px2 py1 rounded-10"
                >
                  {{ commentStatusText(record) }}
                </span>
              </template>
              <template v-else-if="column.key === 'readStatus'">
                <div class="text-#222">
                  {{ readStatusText(record) }}
                </div>
              </template>
              <template v-else-if="column.key === 'parentFeedback'">
                <div class="text-#222">
                  {{ parentFeedbackText(record) }}
                </div>
              </template>
              <template v-else-if="column.key === 'subTeacher'">
                <div class="text-#222">
                  {{ displayText(record.assistants) }}
                </div>
              </template>
              <template v-else-if="column.key === 'classRoom'">
                <div class="text-#222">
                  {{ displayText(record.classRoomName) }}
                </div>
              </template>
              <template v-else-if="column.key === 'action'">
                <a-space :size="14">
                  <a v-if="!record.isComment" class="font500" @click="handleCommentRecord(record)">
                    去记录
                  </a>
                  <a class="font500" @click="handleViewRecord(record)">
                    查看
                  </a>
                </a-space>
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>

    <RehabRecordEditorDrawer
      v-model="drawerOpen"
      :mode="drawerMode"
      :student-teaching-record-id="currentActionRecord?.studentTeachingRecordId"
      :student="currentViewingStudent"
      :session="currentViewingSession"
      @published="loadList"
    />
  </div>
</template>

<style lang="less" scoped>
.total {
  position: relative;
  padding-left: 10px;
  color: #222;
  display: flex;
  align-items: center;

  &::before {
    display: inline-block;
    background: var(--pro-ant-color-primary);
    border-radius: 2px;
    content: "";
    height: 12px;
    left: 0;
    position: absolute;
    width: 4px;
  }
}

.export-word-btn {
  color: var(--pro-ant-color-primary);
  background: #fff;
  border-color: var(--pro-ant-color-primary);
  box-shadow: none;

  &:hover,
  &:focus {
    color: var(--pro-ant-color-primary);
    background: #fff;
    border-color: var(--pro-ant-color-primary);
    box-shadow: none;
  }

  &:active {
    color: var(--pro-ant-color-primary);
    background: #f7fbff;
    border-color: var(--pro-ant-color-primary);
    box-shadow: none;
  }
}

.type-tag-image {
  opacity: 0.4;
}
</style>
