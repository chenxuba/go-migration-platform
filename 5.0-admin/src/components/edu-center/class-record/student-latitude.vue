<script setup lang="ts">
import { DownOutlined, QuestionCircleOutlined } from '@ant-design/icons-vue'
import dayjs, { type Dayjs } from 'dayjs'
import { computed, onMounted, ref, watch } from 'vue'
import { getCourseIdAndNameApi } from '@/api/edu-center/registr-renewal'
import { pageGroupClassesApi } from '@/api/edu-center/group-class'
import { getOneToOneListApi } from '@/api/edu-center/one-to-one'
import { getUserListApi } from '@/api/internal-manage/staff-manage'
import { getStudentTeachingRecordPagedListApi, type StudentTeachingRecordItem } from '@/api/edu-center/class-record'
import StudentAvatar from '@/components/common/StudentAvatar.vue'
import { useTableColumns } from '@/composables/useTableColumns'

interface FilterOption {
  id: string
  value: string
}

const monthStart = dayjs().startOf('month')
const today = dayjs()
const defaultScheduleDateVals = [monthStart.format('YYYY-MM-DD'), today.format('YYYY-MM-DD')]
const displayArray = ref([
  'scheduleDate',
  'stuPhoneSearch',
  'scheduleCourse',
  'scheduleTeacher',
  'assistantTeacher',
  'scheduleClass',
  'scheduleOneToOne',
  'lastEditedTime',
  'scheduleType',
  'studentIdentity',
  'classStatus',
  'billingMode',
  'isArrears',
])
const scheduleTypeOptions = [
  { id: '1', value: '班级日程' },
  { id: '2', value: '1对1日程' },
  { id: '3', value: '试听日程' },
]
const studentIdentityOptions = [
  { id: 'class', value: '班级学员' },
  { id: 'one_to_one', value: '1对1学员' },
  { id: 'trial', value: '试听学员' },
  { id: 'temporary', value: '临时学员' },
  { id: 'makeup', value: '补课学员' },
]
const classStatusOptions = [
  { id: '1', value: '到课' },
  { id: '3', value: '请假' },
  { id: '2', value: '旷课' },
  { id: '0', value: '未记录' },
]
const defaultClassStatusVals = ['1', '3', '2']
const lessonChargingModeOptions = [
  { id: 1, value: '按课时' },
  { id: 2, value: '按时段' },
  { id: 3, value: '按金额' },
  { id: 4, value: '不记课时' },
]
const arrearOnlyOptions = [
  { id: 1, value: '仅显示有拖欠数据的记录' },
]

const loading = ref(false)
const openClassRecordDrawer = ref(false)
const currentTeachingRecordId = ref('')
const dataSource = ref<StudentTeachingRecordItem[]>([])
const filterDateRange = ref<[Dayjs, Dayjs]>([monthStart, today])
const filterUpdatedDateRange = ref<[Dayjs, Dayjs] | null>(null)
const filterStudentId = ref<string | undefined>(undefined)
const filterLessonId = ref<string | undefined>(undefined)
const filterTeacherIds = ref<string[]>([])
const filterAssistantTeacherIds = ref<string[]>([])
const filterClassId = ref<string | undefined>(undefined)
const filterOneToOneId = ref<string | undefined>(undefined)
const filterScheduleTypes = ref<string[]>([])
const filterStudentIdentityValues = ref<string[]>([])
const filterClassStatusValues = ref<string[]>([...defaultClassStatusVals])
const filterLessonChargingModeValues = ref<number[]>([])
const filterIsArrear = ref<boolean | null>(null)
const summary = ref({
  total: 0,
  totalClassTimes: 0,
  totalTuition: 0,
})
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

function handleSeeClassRecord(record?: Partial<StudentTeachingRecordItem>) {
  currentTeachingRecordId.value = String(record?.teachingRecordId || '').trim()
  openClassRecordDrawer.value = true
}

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

const allColumns = ref<any[]>([
  {
    title: '上课日期/时段',
    dataIndex: 'classDateTime',
    key: 'classDateTime',
    fixed: 'left',
    width: 160,
    required: true,
  },
  {
    title: '学员/电话',
    dataIndex: 'name',
    key: 'name',
    fixed: 'left',
    width: 160,
    required: true,
  },
  {
    title: '所属班级/1v1',
    key: 'linkClass1v1',
    dataIndex: 'cloud',
    width: 180,
  },
  {
    title: '所属课程',
    key: 'course',
    dataIndex: 'course',
    width: 160,
  },
  {
    title: '科目',
    dataIndex: 'subject',
    key: 'subject',
    width: 110,
  },
  {
    title: '日程类型',
    dataIndex: 'scheduleType',
    key: 'scheduleType',
    width: 140,
  },
  {
    title: '学员身份',
    dataIndex: 'studentIdentity',
    key: 'studentIdentity',
    width: 140,
  },
  {
    title: '上课状态',
    dataIndex: 'classStatus',
    key: 'classStatus',
    width: 120,
  },
  {
    title: '扣费课程账户',
    dataIndex: 'deductionAccount',
    key: 'deductionAccount',
    width: 160,
  },
  {
    title: '课消方式',
    key: 'courseNotMethod',
    dataIndex: 'courseNotMethod',
    width: 110,
  },
  {
    title: '上课点名数量',
    dataIndex: 'classCallNum',
    key: 'classCallNum',
    width: 160,
  },
  {
    title: '消耗数量',
    dataIndex: 'useNum',
    key: 'useNum',
    width: 140,
  },
  {
    title: '拖欠数量',
    dataIndex: 'oweNum',
    key: 'oweNum',
    width: 140,
  },
  {
    title: '消耗学费',
    dataIndex: 'usePrice',
    key: 'usePrice',
    width: 140,
  },
  {
    title: '上课老师',
    dataIndex: 'mainTeacher',
    key: 'mainTeacher',
    width: 140,
  },
  {
    title: '上课助教',
    dataIndex: 'subTeacher',
    key: 'subTeacher',
    width: 140,
  },
  {
    title: '点名更新时间',
    key: 'callupdateTime',
    dataIndex: 'callupdateTime',
    width: 200,
  },
  {
    title: '对内备注',
    dataIndex: 'externalRemarks',
    key: 'externalRemarks',
    width: 140,
  },
  {
    title: '对外备注',
    dataIndex: 'remarks',
    key: 'remarks',
    width: 140,
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    fixed: 'right',
    width: 140,
  },
])

const { selectedValues, columnOptions, filteredColumns, totalWidth } = useTableColumns({
  storageKey: 'student-latitude',
  allColumns,
  excludeKeys: ['action'],
})

const rowSelection = {
  onChange: (_selectedRowKeys: (string | number)[], _selectedRows: StudentTeachingRecordItem[]) => {},
}

const tablePagination = computed(() => ({
  current: pagination.value.current,
  pageSize: pagination.value.pageSize,
  total: pagination.value.total,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: (total: number) => `共 ${total} 条`,
}))

const headerTipMap: Record<string, string[]> = {
  studentIdentity: [
    '【学员身份】指学员当时是以什么身份来上课的',
    '【举例】A学员由于请假未上课，老师随后对A学员进行了“补课”操作，那么A学员的身份就是补课学员。',
  ],
  courseNotMethod: [
    '【课消方式】课消方式决定了点名时的记录内容。',
    '“按课时”：可以记录课时。',
    '“按金额”：可以记录课时和金额。',
    '“按时间”：可以记录课时。',
  ],
  useNum: [
    '【消耗数量】当次课程真实消耗了多少课时/金额。',
  ],
  oweNum: [
    '【拖欠数量】该学员“剩余数量 < 点名数量时”，会产生“拖欠数量”。',
  ],
  usePrice: [
    '【消耗学费】本次点名数量对应的学费（钱），即机构实际确认收入。',
  ],
  callupdateTime: [
    '【点名更新时间】最近一次编辑点名的时间和操作人',
  ],
}

function formatDateTimeRange(record: Partial<StudentTeachingRecordItem> | Record<string, any>) {
  const start = dayjs(record.startTime)
  const end = dayjs(record.endTime)
  if (!start.isValid() || !end.isValid()) {
    return {
      dateText: '-',
      timeText: '--:-- ~ --:--',
    }
  }
  const weekday = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'][start.day()] || ''
  return {
    dateText: `${start.format('YYYY-MM-DD')} (${weekday})`,
    timeText: `${start.format('HH:mm')} ~ ${end.format('HH:mm')}`,
  }
}

function formatNumber(value?: number, suffix = '') {
  const num = Number(value || 0)
  if (!Number.isFinite(num))
    return suffix ? `0${suffix}` : '0'
  const text = Number.isInteger(num) ? String(num) : num.toFixed(2).replace(/\.?0+$/, '')
  return suffix ? `${text}${suffix}` : text
}

function formatCurrency(value?: number) {
  return `¥${Number(value || 0).toFixed(2)}`
}

function sourceTypeText(value?: number) {
  const type = Number(value || 0)
  if (type === 2)
    return '临时学员'
  if (type === 3 || type === 7)
    return '补课学员'
  if (type === 4)
    return '试听学员'
  if (type === 6)
    return '1对1学员'
  return '班级学员'
}

function buildStudentSourceTypes(values: string[]) {
  const result = new Set<number>()
  values.forEach((item) => {
    if (item === 'class')
      result.add(5)
    else if (item === 'one_to_one')
      result.add(6)
    else if (item === 'trial')
      result.add(4)
    else if (item === 'temporary')
      result.add(2)
    else if (item === 'makeup') {
      result.add(3)
      result.add(7)
    }
  })
  return [...result]
}

function buildClassStatusValues(values: string[]) {
  return values.map((item) => {
    if (item === '0')
      return 0
    return Number(item)
  }).filter(item => Number.isFinite(item))
}

function statusText(value?: number) {
  const status = Number(value || 0)
  if (status === 2)
    return '旷课'
  if (status === 3)
    return '请假'
  if (status === 4)
    return '未记录'
  return '到课'
}

function statusTagClass(value?: number) {
  const status = Number(value || 0)
  if (status === 2)
    return 'record-status-tag record-status-tag--absent'
  if (status === 3)
    return 'record-status-tag record-status-tag--leave'
  if (status === 4)
    return 'record-status-tag record-status-tag--pending'
  return 'record-status-tag record-status-tag--arrived'
}

function scheduleTypeText(value?: number) {
  const type = Number(value || 0)
  if (type === 2)
    return '1对1日程'
  if (type === 3)
    return '试听日程'
  return '班级日程'
}

function scheduleTypeTagClass(value?: number) {
  const type = Number(value || 0)
  if (type === 2)
    return 'record-meta-tag record-meta-tag--one-to-one'
  if (type === 3)
    return 'record-meta-tag record-meta-tag--trial'
  return 'record-meta-tag record-meta-tag--group'
}

function studentIdentityTagClass(value?: number) {
  const type = Number(value || 0)
  if (type === 2)
    return 'record-meta-tag record-meta-tag--temporary'
  if (type === 3 || type === 7)
    return 'record-meta-tag record-meta-tag--make-up'
  if (type === 4)
    return 'record-meta-tag record-meta-tag--trial-student'
  if (type === 6)
    return 'record-meta-tag record-meta-tag--one-to-one-student'
  return 'record-meta-tag record-meta-tag--class-student'
}

function chargingModeText(value?: number) {
  const mode = Number(value || 0)
  if (mode === 2)
    return '按时间'
  if (mode === 3)
    return '按金额'
  return '按课时'
}

function isTimeChargingMode(record: Partial<StudentTeachingRecordItem> | Record<string, any>) {
  return Number(record?.skuMode || 0) === 2
}

function isTrialStudent(record: Partial<StudentTeachingRecordItem> | Record<string, any>) {
  return Number(record?.sourceType || 0) === 4
}

function hasArrearQuantity(record: Partial<StudentTeachingRecordItem> | Record<string, any>) {
  return Number(record?.arrearQuantity || 0) > 0
}

function classDisplay(record: Partial<StudentTeachingRecordItem> | Record<string, any>) {
  return record.className || record.one2OneName || '-'
}

function buildQueryModel() {
  return {
    beginStartTime: filterDateRange.value[0]?.format('YYYY-MM-DD'),
    endStartTime: filterDateRange.value[1]?.format('YYYY-MM-DD'),
    beginUpdatedTime: filterUpdatedDateRange.value?.[0]?.format('YYYY-MM-DD'),
    endUpdatedTime: filterUpdatedDateRange.value?.[1]?.format('YYYY-MM-DD'),
    studentId: filterStudentId.value,
    lessonIds: filterLessonId.value ? [filterLessonId.value] : undefined,
    timetableSourceTypes: filterScheduleTypes.value.map(item => Number(item)).filter(Boolean),
    teacherIds: filterTeacherIds.value.length ? filterTeacherIds.value : undefined,
    assistantTeacherIds: filterAssistantTeacherIds.value.length ? filterAssistantTeacherIds.value : undefined,
    classIds: filterClassId.value ? [filterClassId.value] : undefined,
    one2OneIds: filterOneToOneId.value ? [filterOneToOneId.value] : undefined,
    studentSourceTypes: buildStudentSourceTypes(filterStudentIdentityValues.value),
    studentTeachingRecordStatuses: buildClassStatusValues(filterClassStatusValues.value),
    lessonChargingModeEnums: filterLessonChargingModeValues.value.length ? filterLessonChargingModeValues.value : undefined,
    isArrear: filterIsArrear.value,
  }
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
    console.error('load class record courses failed', error)
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
    console.error('load class record classes failed', error)
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
    console.error('load class record one to one failed', error)
  }
}

async function loadTeacherFilterOptions(
  targetOptions: typeof teacherOptions,
  targetFinished: typeof teacherFinished,
  targetPagination: typeof teacherPagination,
  selectedValues: string[] | string | undefined,
  searchKey = '',
  reset = true,
) {
  if (reset) {
    targetPagination.value.current = 1
    targetFinished.value = false
  }
  try {
    const res = await getUserListApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: targetPagination.value.pageSize,
        pageIndex: targetPagination.value.current,
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
    })).filter(item => item.id && item.value)
    targetOptions.value = reset
      ? mergeFilterOptions(targetOptions.value, resultData, selectedValues)
      : mergeFilterOptions(targetOptions.value, [...targetOptions.value, ...resultData], selectedValues)
    targetPagination.value.total = Number(res.total || resultData.length || 0)
    targetFinished.value = targetOptions.value.length >= targetPagination.value.total
  }
  catch (error) {
    console.error('load class record teachers failed', error)
  }
}

function loadTeacherOptions(searchKey = '', reset = true) {
  teacherSearchKey.value = searchKey
  return loadTeacherFilterOptions(teacherOptions, teacherFinished, teacherPagination, filterTeacherIds.value, searchKey, reset)
}

function loadAssistantTeacherOptions(searchKey = '', reset = true) {
  assistantTeacherSearchKey.value = searchKey
  return loadTeacherFilterOptions(assistantTeacherOptions, assistantTeacherFinished, assistantTeacherPagination, filterAssistantTeacherIds.value, searchKey, reset)
}

async function loadList() {
  loading.value = true
  try {
    const res = await getStudentTeachingRecordPagedListApi({
      queryModel: buildQueryModel(),
      pageRequestModel: {
        needTotal: true,
        pageIndex: pagination.value.current,
        pageSize: pagination.value.pageSize,
        skipCount: (pagination.value.current - 1) * pagination.value.pageSize,
      },
      sortModel: {
        startTime: 2,
        updatedTime: 0,
      },
    })
    if (res.code === 200) {
      dataSource.value = Array.isArray(res.result?.list) ? res.result.list : []
      summary.value = {
        total: Number(res.result?.total || 0),
        totalClassTimes: Number(res.result?.totalClassTimes || 0),
        totalTuition: Number(res.result?.totalTuition || 0),
      }
      pagination.value.total = Number(res.result?.total || 0)
      return
    }
    dataSource.value = []
    summary.value = { total: 0, totalClassTimes: 0, totalTuition: 0 }
    pagination.value.total = 0
  }
  catch (error) {
    console.error('load student teaching records failed', error)
    dataSource.value = []
    summary.value = { total: 0, totalClassTimes: 0, totalTuition: 0 }
    pagination.value.total = 0
  }
  finally {
    loading.value = false
  }
}

function handleScheduleDateFilter(value: unknown) {
  if (!Array.isArray(value) || value.length < 2) {
    filterDateRange.value = [monthStart, today]
    return
  }
  const start = dayjs(String(value[0] || ''))
  const end = dayjs(String(value[1] || ''))
  filterDateRange.value = [
    start.isValid() ? start : monthStart,
    end.isValid() ? end : today,
  ]
}

function handleUpdatedTimeFilter(value: unknown) {
  if (!Array.isArray(value) || value.length < 2) {
    filterUpdatedDateRange.value = null
    return
  }
  const start = dayjs(String(value[0] || ''))
  const end = dayjs(String(value[1] || ''))
  if (!start.isValid() || !end.isValid()) {
    filterUpdatedDateRange.value = null
    return
  }
  filterUpdatedDateRange.value = [start, end]
}

function handleStudentFilter(value: unknown) {
  filterStudentId.value = normalizeFilterValue(value)
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
  filterScheduleTypes.value = Array.isArray(value) ? value.map(item => String(item || '')).filter(Boolean) : []
}

function handleStudentIdentityFilter(value: unknown) {
  filterStudentIdentityValues.value = normalizeFilterValues(value)
}

function handleClassStatusFilter(value: unknown) {
  filterClassStatusValues.value = normalizeFilterValues(value)
}

function handleLessonChargingModeFilter(value: unknown) {
  filterLessonChargingModeValues.value = Array.isArray(value)
    ? value.map(item => Number(item)).filter(item => Number.isFinite(item))
    : []
}

function handleIsArrearsFilter(value: unknown) {
  if (value === null || value === undefined || value === '')
    filterIsArrear.value = null
  else
    filterIsArrear.value = Number(value) === 1
}

function handleTableChange(page: { current?: number, pageSize?: number }) {
  pagination.value.current = Number(page?.current || 1)
  pagination.value.pageSize = Number(page?.pageSize || pagination.value.pageSize)
  loadList()
}

watch(
  [
    filterDateRange,
    filterUpdatedDateRange,
    filterStudentId,
    filterLessonId,
    filterTeacherIds,
    filterAssistantTeacherIds,
    filterClassId,
    filterOneToOneId,
    filterScheduleTypes,
    filterStudentIdentityValues,
    filterClassStatusValues,
    filterLessonChargingModeValues,
    filterIsArrear,
  ],
  () => {
    pagination.value.current = 1
    loadList()
  },
  { deep: true },
)

onMounted(() => {
  loadList()
})
</script>

<template>
  <div>
    <div class="filter-wrap bg-white  pl-3 pr-3 rounded-lb-4 rounded-rb-4">
      <all-filter
        :display-array="displayArray"
        :default-schedule-date-vals="defaultScheduleDateVals"
        :default-class-status-vals="defaultClassStatusVals"
        :schedule-date-disable-future="true"
        :schedule-type-options="scheduleTypeOptions"
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
        :student-identity-options="studentIdentityOptions"
        :class-status-options="classStatusOptions"
        :billing-mode-label="'课消方式'"
        :billing-mode-options-data="lessonChargingModeOptions"
        :is-arrears-label="'拖欠数量'"
        :is-arrears-options-data="arrearOnlyOptions"
        :schedule-course-label="'关联课程'"
        :last-edited-time-label="'点名更新时间'"
        :whole-condition-clear-types="['scheduleTeacher', 'assistantTeacher', 'scheduleType', 'studentIdentity', 'classStatus']"
        :is-quick-show="false"
        :student-status="1"
        @update:schedule-date-filter="handleScheduleDateFilter"
        @update:stu-phone-search-filter="handleStudentFilter"
        @update:schedule-course-filter="handleLessonFilter"
        @update:schedule-teacher-filter="handleTeacherFilter"
        @update:assistant-teacher-filter="handleAssistantTeacherFilter"
        @update:schedule-class-filter="handleClassFilter"
        @update:schedule-one-to-one-filter="handleOneToOneFilter"
        @update:last-edited-time-filter="handleUpdatedTimeFilter"
        @update:schedule-type-filter="handleScheduleTypeFilter"
        @update:student-identity-filter="handleStudentIdentityFilter"
        @update:class-status-filter="handleClassStatusFilter"
        @update:billing-mode-filter="handleLessonChargingModeFilter"
        @update:charging-method-filter="handleLessonChargingModeFilter"
        @update:is-arrears-filter="handleIsArrearsFilter"
      />
    </div>
    <div class="student-list mt-3 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            共 {{ summary.total }} 条记录 ，共记录 {{ formatNumber(summary.totalClassTimes, '课时') }}，共消耗学费 {{ formatCurrency(summary.totalTuition) }}
          </div>
          <div class="edit flex">
            <a-button class="mr-2">
              变更日志
            </a-button>
            <a-dropdown class="mr-2">
              <template #overlay>
                <a-menu>
                  <a-menu-item key="1">
                    批量导出
                  </a-menu-item>
                  <a-menu-item key="3">
                    导出记录
                  </a-menu-item>
                </a-menu>
              </template>
              <a-button>
                导出数据
                <DownOutlined :style="{ fontSize: '10px' }" />
              </a-button>
            </a-dropdown>
            <customize-code
              v-model:checked-values="selectedValues" :options="columnOptions" :total="allColumns.length - 1"
              :num="selectedValues.length - 1"
            />
          </div>
        </div>
        <div class="table-content mt-2">
          <a-table
            row-key="studentTeachingRecordId"
            :loading="loading"
            :data-source="dataSource"
            :pagination="tablePagination"
            :columns="filteredColumns"
            :row-selection="rowSelection"
            :scroll="{ x: totalWidth }"
            size="small"
            @change="handleTableChange"
          >
            <template #headerCell="{ column }">
              <template v-if="headerTipMap[String(column.key || '')]">
                <div class="table-header-with-tip">
                  <span>{{ column.title }}</span>
                  <a-tooltip placement="top">
                    <template #title>
                      <div class="identity-tip">
                        <div v-for="line in headerTipMap[String(column.key || '')]" :key="line">
                          {{ line }}
                        </div>
                      </div>
                    </template>
                    <QuestionCircleOutlined class="table-header-tip-icon" />
                  </a-tooltip>
                </div>
              </template>
            </template>
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'classDateTime'">
                <div class="name">
                  <div class="text-#000">
                    {{ formatDateTimeRange(record).dateText }}
                  </div>
                  <div class="text-3 text-#888 flex flex-items-center">
                    {{ formatDateTimeRange(record).timeText }}
                  </div>
                </div>
              </template>
              <template v-if="column.key === 'name'">
                <StudentAvatar
                  :id="record.studentId"
                  :name="record.studentName || '-'"
                  :avatar-url="record.avatar"
                  :phone="record.studentPhone"
                  :show-gender="false"
                  :show-age="false"
                  default-active-key="0"
                />
              </template>
              <template v-if="column.key === 'linkClass1v1'">
                {{ classDisplay(record) }}
              </template>
              <template v-if="column.key === 'course'">
                {{ record.lessonName || '-' }}
              </template>
              <template v-if="column.key === 'subject'">
                {{ record.subjectName || '-' }}
              </template>
              <template v-if="column.key === 'scheduleType'">
                <span :class="scheduleTypeTagClass(record.timetableSourceType)">
                  {{ scheduleTypeText(record.timetableSourceType) }}
                </span>
              </template>
              <template v-if="column.key === 'studentIdentity'">
                <span :class="studentIdentityTagClass(record.sourceType)">
                  {{ sourceTypeText(record.sourceType) }}
                </span>
              </template>
              <template v-if="column.key === 'classStatus'">
                <span :class="statusTagClass(record.status)">
                  {{ statusText(record.status) }}
                </span>
              </template>
              <template v-if="column.key === 'deductionAccount'">
                {{ isTrialStudent(record) ? '-' : (record.tuitionAccountName || '-') }}
              </template>
              <template v-if="column.key === 'courseNotMethod'">
                {{ isTrialStudent(record) ? '-' : chargingModeText(record.skuMode) }}
              </template>
              <template v-if="column.key === 'classCallNum'">
                {{ isTrialStudent(record) || isTimeChargingMode(record) ? '不记课时' : formatNumber(record.quantity, '课时') }}
              </template>
              <template v-if="column.key === 'useNum'">
                {{ isTrialStudent(record) || isTimeChargingMode(record) ? '-' : formatNumber(record.actualQuantity, '课时') }}
              </template>
              <template v-if="column.key === 'oweNum'">
                <span :class="{ 'owe-num-text': hasArrearQuantity(record) }">
                  {{ isTrialStudent(record) || isTimeChargingMode(record) ? '-' : formatNumber(record.arrearQuantity, '课时') }}
                </span>
              </template>
              <template v-if="column.key === 'usePrice'">
                <span class="use-price-text">
                  {{ isTrialStudent(record) || isTimeChargingMode(record) ? '-' : formatCurrency(record.actualTuition) }}
                </span>
              </template>
              <template v-if="column.key === 'mainTeacher'">
                {{ record.teacherName || '-' }}
              </template>
              <template v-if="column.key === 'subTeacher'">
                {{ record.assistants || '-' }}
              </template>
              <template v-if="column.key === 'callupdateTime'">
                <div class="update-time-cell">
                  <div>
                    {{ record.updatedTime ? dayjs(record.updatedTime).format('YYYY-MM-DD HH:mm') : '-' }}
                  </div>
                  <div class="update-time-operator">
                    {{ record.updatedStaffName || '-' }}
                  </div>
                </div>
              </template>
              <template v-if="column.key === 'externalRemarks'">
                {{ record.remark || '-' }}
              </template>
              <template v-if="column.key === 'remarks'">
                {{ record.externalRemark || '-' }}
              </template>
              <template v-if="column.key === 'action'">
                <a class="font500" @click="handleSeeClassRecord(record)">上课记录详情</a>
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>
    <class-record-details v-model:open="openClassRecordDrawer" :teaching-record-id="currentTeachingRecordId" />
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

.studentStatus {
  span.dot {
    border-radius: 50%;
    display: inline-block;
    height: 6px;
    margin-right: 6px;
    width: 6px;
  }
}

.table-header-with-tip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.table-header-tip-icon {
  color: #999;
  cursor: pointer;
  font-size: 14px;
}

.identity-tip {
  max-width: 240px;
  line-height: 22px;
  white-space: normal;
}

.update-time-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.update-time-operator {
  color: #8c8c8c;
  font-size: 12px;
  line-height: 18px;
}

.record-status-tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 48px;
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 12px;
  line-height: 20px;
  white-space: nowrap;
}

.record-meta-tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 72px;
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 12px;
  line-height: 20px;
  white-space: nowrap;
  border: 1px solid transparent;
}

.record-meta-tag--group,
.record-meta-tag--class-student {
  color: #166534;
  background: #f0fdf4;
  border-color: #bbf7d0;
}

.record-meta-tag--one-to-one,
.record-meta-tag--one-to-one-student {
  color: #1d4ed8;
  background: #eff6ff;
  border-color: #bfdbfe;
}

.record-meta-tag--trial,
.record-meta-tag--trial-student {
  color: #c2410c;
  background: #fff7ed;
  border-color: #fed7aa;
}

.record-meta-tag--temporary {
  color: #7c3aed;
  background: #f5f3ff;
  border-color: #ddd6fe;
}

.record-meta-tag--make-up {
  color: #0f766e;
  background: #f0fdfa;
  border-color: #99f6e4;
}

.record-status-tag--arrived {
  color: #2f6bff;
  background: #eef4ff;
}

.record-status-tag--leave {
  color: #fa8c16;
  background: #fff4e8;
}

.record-status-tag--absent {
  color: #f5222d;
  background: #fff1f0;
}

.record-status-tag--pending {
  color: #8c8c8c;
  background: #f5f5f5;
}

.use-price-text {
  font-weight: 600;
}

.owe-num-text {
  color: #f5222d;
}
</style>
