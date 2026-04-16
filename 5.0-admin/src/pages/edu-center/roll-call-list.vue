<script setup lang="ts">
import type { Dayjs } from 'dayjs'
import type {
  RollCallClassTimetableDetail,
  RollCallConfirmParams,
  RollCallTeachingRecordStudent,
  RollCallTeachingRecordStudentListResult,
} from '@/api/edu-center/roll-call'
import { ExclamationCircleFilled, ExclamationCircleOutlined, LoadingOutlined } from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'
import dayjs from 'dayjs'
import { computed, createVNode, nextTick, onMounted, ref, watch } from 'vue'
import { useTableColumns } from '@/composables/useTableColumns'
import { listClassroomsApi } from '@/api/business-settings/classroom'
import { pageGroupClassesApi } from '@/api/edu-center/group-class'
import { getOneToOneListApi } from '@/api/edu-center/one-to-one'
import { getCourseIdAndNameApi } from '@/api/edu-center/registr-renewal'
import {
  batchEstimateRollCallSufficientTuitionAccountApi,
  checkRollCallTeachingRecordByTeacherAndTimeApi,
  confirmRollCallApi,
  getRollCallClassTimetableApi,
  getRollCallPagedListApi,
  getRollCallStatisticsApi,
  getRollCallTeachingRecordStudentListApi,
} from '@/api/edu-center/roll-call'
import { cancelTeachingSchedulesApi, type TeachingScheduleItem } from '@/api/edu-center/teaching-schedule'
import messageService from '@/utils/messageService'

interface FilterOption {
  id: string
  value: string
}

interface AllFilterExpose {
  setScheduleDateFilter: (values?: string[], shouldEmit?: boolean) => void
}

interface TeacherRoleFilterPayload {
  teacherType?: number
  teacherId?: string
}

type DashboardFilter = 'today' | 'all' | 'partial' | 'custom'
type BatchRollCallStatusKey = 'attended' | 'absent' | 'leave' | 'unrecorded'

interface BatchRollCallStudentRecord {
  id: string
  studentAccount: string
  tuitionAccountId: string
  sourceType: number
  consumptionMethod: string
  rawChargingMode: number
  recordAttendance: boolean
  attendanceCount: number
  internalNote: string
  externalNote: string
  hasTeachingRecord: boolean
  locked: boolean
  autoRollCall: boolean
  attended: boolean
  absent: boolean
  leave: boolean
  unrecorded: boolean
}

interface BatchRollCallFailureItem {
  id: string
  scheduleName: string
  teacherName: string
  reason: string
}

const today = dayjs().format('YYYY-MM-DD')
const monthStart = dayjs().startOf('month')
const todayDayjs = dayjs()
const defaultScheduleDateVals = [monthStart.format('YYYY-MM-DD'), todayDayjs.format('YYYY-MM-DD')]
const batchRollCallLimit = 10

const displayArray = ref([
  'scheduleDate',
  'scheduleCourse',
  'scheduleClassroom',
  'scheduleClass',
  'scheduleOneToOne',
  'scheduleType',
])

const dataSource = ref<TeachingScheduleItem[]>([])
const tableLoading = ref(false)
const statisticsLoading = ref(false)
const batchDeleting = ref(false)
const batchRollCallSelecting = ref(false)
const batchRollCallSubmitting = ref(false)
const batchRollCallResultOpen = ref(false)
const openDrawer = ref(false)
const createUnscheduledRollCallOpen = ref(false)
const currentRollCallScheduleId = ref('')
const currentRollCallLessonDay = ref('')
const allFilterRef = ref<AllFilterExpose | null>(null)
const dashboardFilter = ref<DashboardFilter>('custom')
const dateRange = ref<[Dayjs, Dayjs]>([monthStart, todayDayjs])
const sortDirection = ref(2)

const pagination = ref({
  current: 1,
  pageSize: 50,
  total: 0,
})
const selectedRowKeys = ref<string[]>([])
const batchRollCallCheckedKeys = ref<string[]>([])

const dashboardStats = ref({
  todayCount: 0,
  allCount: 0,
  partialCount: 0,
})
const batchRollCallProgress = ref({
  total: 0,
  completed: 0,
  currentName: '',
})
const batchRollCallResult = ref({
  succeedCount: 0,
  failedCount: 0,
  failedItems: [] as BatchRollCallFailureItem[],
})

const filterTeacherRole = ref<TeacherRoleFilterPayload>({
  teacherType: 1,
  teacherId: undefined,
})
const filterClassroomId = ref<string[]>([])
const filterClassId = ref<string | undefined>(undefined)
const filterOneToOneId = ref<string | undefined>(undefined)
const filterCourseId = ref<string | undefined>(undefined)
const filterScheduleType = ref<string[]>([])

const scheduleClassroomOptions = ref<FilterOption[]>([])
const scheduleClassOptions = ref<FilterOption[]>([])
const scheduleOneToOneOptions = ref<FilterOption[]>([])
const scheduleCourseOptions = ref<FilterOption[]>([])

const scheduleClassroomFinished = ref(false)
const scheduleClassFinished = ref(false)
const scheduleOneToOneFinished = ref(false)
const scheduleCourseFinished = ref(false)

const scheduleClassPagination = ref({ current: 1, pageSize: 20, total: 0 })
const scheduleOneToOnePagination = ref({ current: 1, pageSize: 20, total: 0 })

const scheduleClassSearchKey = ref('')
const scheduleOneToOneSearchKey = ref('')

const scheduleTypeOptions = [
  { id: 'group_class', value: '班级日程' },
  { id: 'one_to_one', value: '1对1日程' },
  { id: 'trial', value: '试听日程' },
]

const allColumns = ref([
  {
    title: '上课日期/时段',
    dataIndex: 'classDateTime',
    key: 'classDateTime',
    width: 180,
    fixed: 'left',
    sorter: true,
    defaultSortOrder: 'descend',
  },
  {
    title: '日程类型',
    dataIndex: 'scheduleType',
    key: 'scheduleType',
    width: 120,
  },
  {
    title: '班级/1对1',
    key: 'classOr1v1',
    dataIndex: 'classOr1v1',
    width: 180,
  },
  {
    title: '课程名称',
    key: 'courseName',
    dataIndex: 'courseName',
    width: 140,
  },
  {
    title: '上课老师',
    key: 'mainTeacher',
    dataIndex: 'mainTeacher',
    width: 110,
  },
  {
    title: '上课助教',
    dataIndex: 'subTeacher',
    key: 'subTeacher',
    width: 140,
  },
  {
    title: '上课教室',
    dataIndex: 'classRoom',
    key: 'classRoom',
    width: 120,
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    fixed: 'right',
    width: 100,
  },
])

const { selectedValues, columnOptions, filteredColumns, totalWidth }
  = useTableColumns({
    storageKey: 'roll-call-list',
    allColumns,
    excludeKeys: ['action'],
  })

const tablePagination = computed(() => ({
  current: pagination.value.current,
  pageSize: pagination.value.pageSize,
  total: pagination.value.total,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: (total: number) => `共 ${total} 条`,
}))

const batchRollCallColumns = [
  {
    title: '上课日期/时段',
    dataIndex: 'classDateTime',
    key: 'classDateTime',
    width: 220,
  },
  {
    title: '班级/1对1',
    dataIndex: 'classOr1v1',
    key: 'classOr1v1',
    width: 180,
  },
  {
    title: '课程名称',
    dataIndex: 'courseName',
    key: 'courseName',
    width: 160,
  },
  {
    title: '上课教师',
    dataIndex: 'mainTeacher',
    key: 'mainTeacher',
    width: 140,
  },
]

const rowSelection = computed(() => ({
  selectedRowKeys: selectedRowKeys.value,
  onChange: (keys: (string | number)[]) => {
    selectedRowKeys.value = keys.map(key => String(key))
  },
}))

const selectedBatchRollCallRows = computed(() => {
  const selectedSet = new Set(selectedRowKeys.value.map(key => String(key)))
  return dataSource.value.filter(item => selectedSet.has(String(item.id)))
})

const batchRollCallCheckedRows = computed(() => {
  const selectedSet = new Set(batchRollCallCheckedKeys.value.map(key => String(key)))
  return selectedBatchRollCallRows.value.filter(item => selectedSet.has(String(item.id)))
})

const batchRollCallSelectRowSelection = computed(() => ({
  selectedRowKeys: batchRollCallCheckedKeys.value,
  onChange: (keys: (string | number)[]) => {
    const normalized = keys.map(key => String(key))
    if (normalized.length > batchRollCallLimit) {
      messageService.warning(`批量点名一次最多选择 ${batchRollCallLimit} 条日程`)
      return
    }
    batchRollCallCheckedKeys.value = normalized
  },
  getCheckboxProps: (record: TeachingScheduleItem) => {
    const recordId = String(record.id || '')
    const checked = batchRollCallCheckedKeys.value.includes(recordId)
    return {
      disabled: record.canRollCall === false || (!checked && batchRollCallCheckedKeys.value.length >= batchRollCallLimit),
    }
  },
}))

const batchRollCallProgressPercent = computed(() => {
  if (!batchRollCallProgress.value.total)
    return 0
  return Math.min(100, Math.round(batchRollCallProgress.value.completed / batchRollCallProgress.value.total * 100))
})

const batchRollCallFailureSummary = computed(() =>
  batchRollCallResult.value.failedItems
    .map(item => `${item.scheduleName}${item.reason ? `：${item.reason}` : ''}`)
    .join('；'),
)

function normalizeScheduleFilterValue(value: unknown) {
  if (Array.isArray(value))
    return value.length ? String(value[0] ?? '').trim() || undefined : undefined
  const text = String(value ?? '').trim()
  return text || undefined
}

function normalizeScheduleFilterValues(value: unknown) {
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

function isSameDateRange(nextStart: Dayjs, nextEnd: Dayjs) {
  return dateRange.value[0]?.format('YYYY-MM-DD') === nextStart.format('YYYY-MM-DD')
    && dateRange.value[1]?.format('YYYY-MM-DD') === nextEnd.format('YYYY-MM-DD')
}

function buildStatisticsQueryModel() {
  return {
    lessonId: filterCourseId.value,
    classroomId: filterClassroomId.value[0],
    classId: filterClassId.value,
    oneToOneId: filterOneToOneId.value,
    teacherId: filterTeacherRole.value.teacherId,
    teacherTypes: filterTeacherRole.value.teacherId ? [Number(filterTeacherRole.value.teacherType || 1)] : undefined,
    scheduleTypes: filterScheduleType.value.length ? filterScheduleType.value : undefined,
  }
}

function buildDateRangeForList() {
  if (dashboardFilter.value === 'today') {
    return {
      startDate: today,
      endDate: today,
    }
  }
  if (dashboardFilter.value === 'all' || dashboardFilter.value === 'partial') {
    return {
      startDate: undefined,
      endDate: today,
    }
  }
  return {
    startDate: dateRange.value[0]?.format('YYYY-MM-DD'),
    endDate: dateRange.value[1]?.format('YYYY-MM-DD'),
  }
}

function buildListQueryModel() {
  const range = buildDateRangeForList()
  return {
    ...buildStatisticsQueryModel(),
    startDate: range.startDate,
    endDate: range.endDate,
    callStatusMode: dashboardFilter.value === 'partial' ? 'partial' : 'incomplete',
  }
}

async function loadStatistics() {
  statisticsLoading.value = true
  try {
    const res = await getRollCallStatisticsApi({
      queryModel: buildStatisticsQueryModel(),
    })
    if (res.code === 200) {
      dashboardStats.value = {
        todayCount: Number(res.result?.todayCount || 0),
        allCount: Number(res.result?.allCount || 0),
        partialCount: Number(res.result?.partialCount || 0),
      }
      return
    }
    dashboardStats.value = { todayCount: 0, allCount: 0, partialCount: 0 }
  }
  catch (error) {
    console.error('load roll call statistics failed', error)
    dashboardStats.value = { todayCount: 0, allCount: 0, partialCount: 0 }
  }
  finally {
    statisticsLoading.value = false
  }
}

async function loadList() {
  tableLoading.value = true
  try {
    const res = await getRollCallPagedListApi({
      queryModel: buildListQueryModel(),
      pageRequestModel: {
        needTotal: true,
        pageIndex: pagination.value.current,
        pageSize: pagination.value.pageSize,
        skipCount: (pagination.value.current - 1) * pagination.value.pageSize,
      },
      sortModel: {
        byStartDate: sortDirection.value,
      },
    })
    if (res.code === 200) {
      dataSource.value = Array.isArray(res.result?.list) ? res.result.list : []
      pagination.value.total = Number(res.result?.total || 0)
      selectedRowKeys.value = selectedRowKeys.value.filter(key => dataSource.value.some(item => String(item.id) === key))
      batchRollCallCheckedKeys.value = batchRollCallCheckedKeys.value.filter(key => dataSource.value.some(item => String(item.id) === key))
      return
    }
    dataSource.value = []
    pagination.value.total = 0
    selectedRowKeys.value = []
    batchRollCallCheckedKeys.value = []
  }
  catch (error) {
    console.error('load roll call list failed', error)
    dataSource.value = []
    pagination.value.total = 0
    selectedRowKeys.value = []
    batchRollCallCheckedKeys.value = []
  }
  finally {
    tableLoading.value = false
  }
}

function resetToFirstPage() {
  pagination.value.current = 1
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

function handleScheduleDateFilter(value: unknown) {
  const normalized = normalizeScheduleFilterValues(value)
  if (normalized.length >= 2) {
    const start = dayjs(normalized[0])
    const end = dayjs(normalized[1])
    if (start.isValid() && end.isValid()) {
      if (dashboardFilter.value === 'custom' && isSameDateRange(start, end))
        return
      dateRange.value = [start, end]
      dashboardFilter.value = 'custom'
      return
    }
  }
  if (dashboardFilter.value === 'custom' && isSameDateRange(monthStart, todayDayjs))
    return
  dateRange.value = [monthStart, todayDayjs]
  dashboardFilter.value = 'custom'
}

function handleScheduleTypeFilter(value: unknown) {
  filterScheduleType.value = normalizeScheduleFilterValues(value)
}

function handleTeacherRoleFilter(payload: TeacherRoleFilterPayload) {
  filterTeacherRole.value = {
    teacherType: Number(payload?.teacherType || 1),
    teacherId: payload?.teacherId ? String(payload.teacherId) : undefined,
  }
}

function handleQuickFilter(type: DashboardFilter) {
  dashboardFilter.value = type
  if (type === 'today') {
    dateRange.value = [todayDayjs, todayDayjs]
    allFilterRef.value?.setScheduleDateFilter([today, today])
  }
  else if (type === 'custom') {
    allFilterRef.value?.setScheduleDateFilter([
      dateRange.value[0].format('YYYY-MM-DD'),
      dateRange.value[1].format('YYYY-MM-DD'),
    ])
  }
}

function handleRollCall(record?: Partial<TeachingScheduleItem> | Record<string, any>) {
  currentRollCallScheduleId.value = String(record?.id || '').trim()
  currentRollCallLessonDay.value = String(record?.lessonDate || '').trim()
  openDrawer.value = true
}

function handleOpenCreateUnscheduledRollCall() {
  createUnscheduledRollCallOpen.value = true
}

async function refreshAfterRollCall() {
  resetToFirstPage()
  await loadStatistics()
  await loadList()
}

function handleBatchDelete() {
  if (!selectedRowKeys.value.length) {
    messageService.warning('请先选择要删除的日程')
    return
  }

  Modal.confirm({
    title: '确定删除？',
    icon: createVNode(ExclamationCircleOutlined, { style: 'color: #ff4d4f' }),
    content: `删除的日程无法恢复，并不会再显示在课表中。已选择 ${selectedRowKeys.value.length} 条，确认删除？`,
    okText: '确定删除',
    cancelText: '再想想',
    async onOk() {
      batchDeleting.value = true
      try {
        const res = await cancelTeachingSchedulesApi({
          ids: selectedRowKeys.value,
        })
        if (res.code !== 200)
          throw new Error(res.message || '批量删除失败')
        messageService.success(`已删除 ${Number(res.result?.canceled || selectedRowKeys.value.length)} 条日程`)
        selectedRowKeys.value = []
        await loadStatistics()
        await loadList()
      }
      catch (error: any) {
        console.error('batch delete roll call schedules failed', error)
        messageService.error(error?.response?.data?.message || error?.message || '批量删除失败')
        throw error
      }
      finally {
        batchDeleting.value = false
      }
    },
  })
}

function handleOpenBatchRollCall() {
  if (!selectedRowKeys.value.length) {
    messageService.warning('请先选择要批量点名的日程')
    return
  }

  const selectableRows = selectedBatchRollCallRows.value.filter(item => item.canRollCall !== false)
  if (!selectableRows.length) {
    messageService.warning('已选日程暂不可批量点名')
    return
  }

  batchRollCallCheckedKeys.value = selectableRows
    .slice(0, batchRollCallLimit)
    .map(item => String(item.id || ''))

  if (selectableRows.length > batchRollCallLimit)
    messageService.warning(`批量点名一次最多选择 ${batchRollCallLimit} 条日程`)

  batchRollCallSelecting.value = true
}

async function handleBatchRollCallConfirm() {
  if (batchRollCallSubmitting.value)
    return

  const rows = [...batchRollCallCheckedRows.value]
  if (!rows.length) {
    messageService.warning('请先勾选要批量点名的日程')
    return
  }

  batchRollCallSelecting.value = false
  batchRollCallSubmitting.value = true
  batchRollCallProgress.value = {
    total: rows.length,
    completed: 0,
    currentName: '',
  }
  batchRollCallResult.value = {
    succeedCount: 0,
    failedCount: 0,
    failedItems: [],
  }
  await nextTick()

  const nextResult = {
    succeedCount: 0,
    failedCount: 0,
    failedItems: [] as BatchRollCallFailureItem[],
  }

  try {
    for (let index = 0; index < rows.length; index += 1) {
      const item = rows[index]
      batchRollCallProgress.value.currentName = getBatchRollCallScheduleName(item)
      await nextTick()
      try {
        await executeBatchRollCall(item)
        nextResult.succeedCount += 1
      }
      catch (error) {
        console.error('batch roll call failed', error)
        nextResult.failedCount += 1
        nextResult.failedItems.push({
          id: String(item.id || ''),
          scheduleName: getBatchRollCallScheduleName(item),
          teacherName: String(item.teacherName || '-') || '-',
          reason: getBatchRollCallErrorMessage(error, '点名提交失败'),
        })
      }
      batchRollCallProgress.value.completed = index + 1
    }

    batchRollCallResult.value = nextResult
    selectedRowKeys.value = []
    await refreshAfterRollCall()
    batchRollCallResultOpen.value = true
  }
  finally {
    batchRollCallSubmitting.value = false
  }
}

function handleTableChange(page: { current?: number, pageSize?: number }, _filters: unknown, sorter: { order?: string } | Array<{ order?: string }>) {
  pagination.value.current = Number(page?.current || 1)
  pagination.value.pageSize = Number(page?.pageSize || pagination.value.pageSize)
  const currentSorter = Array.isArray(sorter) ? sorter[0] : sorter
  if (currentSorter?.order === 'ascend')
    sortDirection.value = 1
  else if (currentSorter?.order === 'descend')
    sortDirection.value = 2
  else
    sortDirection.value = 2
  loadList()
}

function scheduleTypeLabel(record: Record<string, any>) {
  return Number(record.classType) === 2 ? '1对1日程' : '班级日程'
}

function classDisplayName(record: Record<string, any>) {
  return record.teachingClassName || record.studentName || '-'
}

function assistantTeacherText(record: Record<string, any>) {
  if (Array.isArray(record.assistantNames) && record.assistantNames.length)
    return record.assistantNames.join('、')
  return '-'
}

function formatLessonWeekday(lessonDate?: string) {
  if (!lessonDate)
    return ''
  const weekdays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
  return weekdays[dayjs(lessonDate).day()] || ''
}

function formatLessonTime(record: Record<string, any>) {
  const dateText = record.lessonDate ? `${record.lessonDate} (${formatLessonWeekday(record.lessonDate)})` : '-'
  const startText = record.startAt ? dayjs(record.startAt).format('HH:mm') : '--:--'
  const endText = record.endAt ? dayjs(record.endAt).format('HH:mm') : '--:--'
  return {
    dateText,
    timeText: `${startText} ~ ${endText}`,
  }
}

function normalizeLessonDayValue(value?: string) {
  const text = String(value || '').trim()
  if (!text)
    return ''
  return text.includes('T') ? text : `${text}T00:00:00`
}

function parseBatchRollCallNumber(value: unknown) {
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : 0
}

function defaultTeachingStatusToKey(value: unknown): BatchRollCallStatusKey {
  if (Number(value) === 2)
    return 'absent'
  if (Number(value) === 3)
    return 'leave'
  if (Number(value) === 4)
    return 'unrecorded'
  return 'attended'
}

function buildBatchRollCallStudentRecord(item: RollCallTeachingRecordStudent, detail: RollCallClassTimetableDetail): BatchRollCallStudentRecord {
  const hasTeachingRecord = Boolean(item?.hasTeachingRecord)
  const statusKey = hasTeachingRecord
    ? defaultTeachingStatusToKey(item?.studentTeachingStatus)
    : defaultTeachingStatusToKey(item?.defaultStudentTeachingStatus)
  const effectiveChargingMode = hasTeachingRecord && Number(item?.recordedSkuMode || 0) > 0
    ? Number(item?.recordedSkuMode || 0)
    : Number(item?.chargingMode || 0)
  const attendanceCount = hasTeachingRecord
    ? Number(item?.recordedQuantity || 0)
    : Number(detail?.defaultStudentClassTime || 1)

  return {
    id: String(item?.studentId || ''),
    studentAccount: String(item?.studentName || ''),
    tuitionAccountId: String(item?.recordedTuitionAccountId || item?.tuitionAccountId || ''),
    sourceType: Number(item?.sourceType || 0),
    consumptionMethod: String(effectiveChargingMode || ''),
    rawChargingMode: effectiveChargingMode,
    recordAttendance: effectiveChargingMode === 1 && statusKey !== 'unrecorded',
    attendanceCount,
    internalNote: hasTeachingRecord ? String(item?.recordedRemark || '') : '',
    externalNote: hasTeachingRecord ? String(item?.recordedExternalRemark || '') : '',
    hasTeachingRecord,
    locked: Boolean(item?.locked),
    autoRollCall: Boolean(item?.autoRollCall),
    attended: statusKey === 'attended',
    absent: statusKey === 'absent',
    leave: statusKey === 'leave',
    unrecorded: statusKey === 'unrecorded',
  }
}

function shouldSkipBatchRollCallStudent(record: BatchRollCallStudentRecord) {
  return Boolean(record?.locked) || Boolean(record?.autoRollCall) || Boolean(record?.hasTeachingRecord)
}

function getBatchRollCallRecordStatus(record: BatchRollCallStudentRecord) {
  if (record?.absent)
    return 2
  if (record?.leave)
    return 3
  if (record?.unrecorded)
    return 4
  return 1
}

function buildBatchRollCallStudentQuantity(record: BatchRollCallStudentRecord) {
  if (!record || record.unrecorded || !record.recordAttendance)
    return 0
  return Math.max(parseBatchRollCallNumber(record.attendanceCount), 0)
}

function buildBatchRollCallEstimatePayload(records: BatchRollCallStudentRecord[]) {
  return records
    .filter(record =>
      !shouldSkipBatchRollCallStudent(record)
      && Number(record?.consumptionMethod || 0) === 1
      && buildBatchRollCallStudentQuantity(record) > 0,
    )
    .map(record => ({
      quantity: buildBatchRollCallStudentQuantity(record),
      tuitionAccountId: String(record?.tuitionAccountId || ''),
      studentName: String(record?.studentAccount || ''),
    }))
}

function buildBatchRollCallConfirmPayload(
  detail: RollCallClassTimetableDetail,
  recordResult: RollCallTeachingRecordStudentListResult,
  students: BatchRollCallStudentRecord[],
  scheduleId: string,
): RollCallConfirmParams {
  const meta = (recordResult?.data || {}) as Record<string, any>
  const teacherList = Array.isArray(recordResult?.teachers) ? recordResult.teachers : []
  return {
    sourceName: String(meta?.sourceName || detail?.className || ''),
    teachingContent: '',
    teachingContentImages: [],
    timetableSourceType: Number(meta?.timetableSourceType || 0),
    timetableSourceId: String(meta?.timetableSourceId || scheduleId || ''),
    teachingRecordId: String(meta?.teachingRecordId || ''),
    sourceId: String(meta?.sourceId || detail?.classId || ''),
    sourceType: Number(meta?.sourceType || 0),
    lessonId: String(meta?.lessonId || detail?.lessonId || ''),
    startTime: String(meta?.startTime || ''),
    endTime: String(meta?.endTime || ''),
    teacherClassTime: Number(meta?.teacherClassTime || detail?.defaultTeacherClassTime || 0),
    studentShouldDeduct: Number(detail?.defaultStudentClassTime || 1),
    teacherList: teacherList.map(item => ({
      teacherId: String(item?.teacherId || ''),
      type: Number(item?.type || 0),
    })),
    studentList: students
      .filter(record => !shouldSkipBatchRollCallStudent(record))
      .map((record) => {
        const quantity = buildBatchRollCallStudentQuantity(record)
        return {
          studentShouldDeduct: quantity,
          studentName: String(record?.studentAccount || ''),
          studentId: String(record?.id || ''),
          tuitionAccountId: String(record?.tuitionAccountId || '0'),
          absentTeachingRecordId: '0',
          status: getBatchRollCallRecordStatus(record),
          sourceType: Number(record?.sourceType || 0),
          remark: String(record?.internalNote || ''),
          externalRemark: String(record?.externalNote || ''),
          skuMode: Number(record?.consumptionMethod || record?.rawChargingMode || 0),
          amount: 0,
          quantity,
        }
      }),
    subjectId: '0',
    classRoomId: String(meta?.classroomId || detail?.addressId || '0'),
  }
}

function getBatchRollCallScheduleName(record: Partial<TeachingScheduleItem>) {
  const { dateText, timeText } = formatLessonTime(record as Record<string, any>)
  return [dateText, timeText, String(record?.lessonName || '').trim()].filter(Boolean).join(' ')
}

function getBatchRollCallErrorMessage(error: any, fallback: string) {
  return String(error?.response?.data?.message || error?.message || fallback).trim() || fallback
}

async function executeBatchRollCall(record: TeachingScheduleItem) {
  const scheduleId = String(record?.id || '').trim()
  if (!scheduleId)
    throw new Error('缺少日程标识')

  const detailRes = await getRollCallClassTimetableApi({
    id: scheduleId,
    lessonDay: normalizeLessonDayValue(record?.lessonDate) || undefined,
  })
  if (detailRes.code !== 200 || !detailRes.result?.detail)
    throw new Error(detailRes.message || '加载点名课表失败')

  const detail = detailRes.result.detail
  const lessonDay = String(detail.lessonDays?.[0]?.lessonDay || normalizeLessonDayValue(record?.lessonDate) || '')
  const recordRes = await getRollCallTeachingRecordStudentListApi({
    timetableSourceId: scheduleId,
    timetableSourceType: 1,
    classId: String(detail.classId || ''),
    lessonId: String(detail.lessonId || ''),
    one2OneId: '0',
    startDate: String(detail.startDate || '0001-01-01T00:00:00'),
    endDate: String(detail.endDate || '0001-01-01T00:00:00'),
    lessonDay,
  })
  if (recordRes.code !== 200 || !recordRes.result)
    throw new Error(recordRes.message || '加载点名学员失败')

  const students = (Array.isArray(recordRes.result.students) ? recordRes.result.students : [])
    .map(item => buildBatchRollCallStudentRecord(item, detail))
  const payload = buildBatchRollCallConfirmPayload(detail, recordRes.result, students, scheduleId)
  if (!payload.startTime || !payload.endTime)
    throw new Error('缺少上课时间，请刷新后重试')
  if (!Array.isArray(payload.studentList) || payload.studentList.length === 0)
    throw new Error('当前没有可提交的点名学员')

  const mainTeacher = payload.teacherList.find(item => Number(item?.type) === 1)
  if (mainTeacher?.teacherId) {
    const checkRes = await checkRollCallTeachingRecordByTeacherAndTimeApi({
      startTime: payload.startTime,
      endTime: payload.endTime,
      teacherId: String(mainTeacher.teacherId || ''),
      timetableSourceId: String(payload.timetableSourceId || ''),
      excludeTeachingRecordId: String(payload.teachingRecordId || ''),
    })
    if (checkRes.code !== 200)
      throw new Error(checkRes.message || '上课教师时间冲突校验失败')
  }

  const estimatePayload = buildBatchRollCallEstimatePayload(students)
  if (estimatePayload.length > 0) {
    const estimateRes = await batchEstimateRollCallSufficientTuitionAccountApi({
      tuitionInfoList: estimatePayload,
    })
    if (estimateRes.code !== 200)
      throw new Error(estimateRes.message || '扣费账户剩余校验失败')

    const insufficientIdSet = new Set(
      (Array.isArray(estimateRes.result?.tuitionInfoList) ? estimateRes.result.tuitionInfoList : [])
        .filter(item => item?.isSufficient === false)
        .map(item => String(item?.tuitionAccountId || '')),
    )
    const insufficientNames = Array.from(new Set(
      estimatePayload
        .filter(item => insufficientIdSet.has(String(item?.tuitionAccountId || '')))
        .map(item => String(item?.studentName || '').trim())
        .filter(Boolean),
    ))
    if (insufficientNames.length > 0)
      throw new Error(`学费不足：${insufficientNames.join('、')}`)
  }

  const confirmRes = await confirmRollCallApi(payload)
  if (confirmRes.code !== 200)
    throw new Error(confirmRes.message || '点名提交失败')
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
    console.error('load roll call classrooms failed', error)
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
    scheduleClassPagination.value.total = Number(res.result?.total || resultData.length || 0)
    scheduleClassFinished.value = scheduleClassOptions.value.length >= scheduleClassPagination.value.total
  }
  catch (error) {
    console.error('load roll call classes failed', error)
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
    scheduleOneToOnePagination.value.total = Number(res.result?.total || resultData.length || 0)
    scheduleOneToOneFinished.value = scheduleOneToOneOptions.value.length >= scheduleOneToOnePagination.value.total
  }
  catch (error) {
    console.error('load roll call one-to-one options failed', error)
  }
}

async function loadScheduleCourseOptions(searchKey = '', reset = true) {
  try {
    const res = await getCourseIdAndNameApi({ searchKey })
    if (res.code !== 200)
      return
    const resultData = (Array.isArray(res.result) ? res.result : []).map(item => ({
      id: String(item.id ?? ''),
      value: String(item.name || item.id || '').trim(),
    })).filter(item => item.id && item.value)
    scheduleCourseOptions.value = mergeFilterOptions(reset ? [] : scheduleCourseOptions.value, resultData, filterCourseId.value)
    scheduleCourseFinished.value = true
  }
  catch (error) {
    console.error('load roll call courses failed', error)
  }
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

watch(
  [filterTeacherRole, filterClassroomId, filterClassId, filterOneToOneId, filterCourseId, filterScheduleType],
  () => {
    resetToFirstPage()
    loadStatistics()
    loadList()
  },
  { deep: true },
)

watch(
  [dashboardFilter, dateRange],
  () => {
    resetToFirstPage()
    loadList()
  },
  { deep: true },
)

onMounted(async () => {
  await loadStatistics()
  await loadList()
})
</script>

<template>
  <div class="roll-call">
    <div class="databord bg-white pt-3 pb-3 pl-5 pr-5 rounded-4">
      <custom-title title="关键数据看板" font-size="14px" font-weight="500" />
      <div class="flex justify-between mt-3 mb-2">
        <div class="flex-1 bg-#fbfcff h-22.5 cursor-pointer rounded-5 hover-bg-#0066ff0d" @click="handleQuickFilter('today')">
          <div class="contentMain">
            <div class="contentMainLeft">
              今日待点名
            </div>
            <div class="contentMainRight">
              {{ statisticsLoading ? '-' : dashboardStats.todayCount }}
            </div>
          </div>
          <div class="contentSub">
            <div class="contentSubLeft">
              今日待点名的日程
            </div>
            <div class="contentSubRight">
              快捷筛选
            </div>
          </div>
        </div>
        <div class="flex-1 bg-#fbfcff h-22.5 cursor-pointer rounded-5 hover-bg-#0066ff0d ml-4 mr-4" @click="handleQuickFilter('all')">
          <div class="contentMain">
            <div class="contentMainLeft">
              全部待点名
            </div>
            <div class="contentMainRight">
              {{ statisticsLoading ? '-' : dashboardStats.allCount }}
            </div>
          </div>
          <div class="contentSub">
            <div class="contentSubLeft">
              过去至今未完成全部点名的日程
            </div>
            <div class="contentSubRight">
              快捷筛选
            </div>
          </div>
        </div>
        <div class="flex-1 bg-#fbfcff h-22.5 cursor-pointer rounded-5 hover-bg-#0066ff0d" @click="handleQuickFilter('partial')">
          <div class="contentMain">
            <div class="contentMainLeft">
              部分点名
            </div>
            <div class="contentMainRight">
              {{ statisticsLoading ? '-' : dashboardStats.partialCount }}
            </div>
          </div>
          <div class="contentSub">
            <div class="contentSubLeft">
              已点名但未完成全部点名的日程
            </div>
            <div class="contentSubRight">
              前往处理
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="bg-white rounded-4 mt-3 pl-2 pr-2 py-3">
      <all-filter
        ref="allFilterRef"
        :display-array="displayArray"
        :default-schedule-date-vals="defaultScheduleDateVals"
        :schedule-date-disable-future="true"
        is-show-chang-teacher-search
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
        :schedule-type-options="scheduleTypeOptions"
        @update:schedule-classroom-filter="handleScheduleClassroomFilter"
        @update:schedule-class-filter="handleScheduleClassFilter"
        @update:schedule-one-to-one-filter="handleScheduleOneToOneFilter"
        @update:schedule-course-filter="handleScheduleCourseFilter"
        @update:schedule-date-filter="handleScheduleDateFilter"
        @update:schedule-type-filter="handleScheduleTypeFilter"
        @update:teacher-role-filter="handleTeacherRoleFilter"
      />
    </div>

    <div class="bg-white rounded-4 mt-3 py-3 px-5">
      <div class="table-title flex justify-between">
        <div class="total">
          当前共计 {{ pagination.total }} 条待点名日程
        </div>
        <div class="edit flex">
          <a-button class="mr-3" :loading="batchDeleting" @click="handleBatchDelete">
            批量删除
          </a-button>
          <a-button class="mr-3" @click="handleOpenBatchRollCall">
            批量点名
          </a-button>
          <a-button class="mr-3" type="primary" @click="handleOpenCreateUnscheduledRollCall">
            创建未排课点名
          </a-button>
          <customize-code
            v-model:checked-values="selectedValues"
            :options="columnOptions"
            :total="allColumns.length - 1"
            :num="selectedValues.length - 1"
          />
        </div>
      </div>
      <a-table
        row-key="id"
        :loading="tableLoading"
        :data-source="dataSource"
        :pagination="tablePagination"
        :row-selection="rowSelection"
        :columns="filteredColumns"
        :scroll="{ x: totalWidth }"
        size="small"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'classDateTime'">
            <div>{{ formatLessonTime(record).dateText }}</div>
            <div>{{ formatLessonTime(record).timeText }}</div>
          </template>
          <template v-else-if="column.key === 'scheduleType'">
            <span>{{ scheduleTypeLabel(record) }}</span>
          </template>
          <template v-else-if="column.key === 'classOr1v1'">
            {{ classDisplayName(record) }}
          </template>
          <template v-else-if="column.key === 'courseName'">
            {{ record.lessonName || '-' }}
          </template>
          <template v-else-if="column.key === 'mainTeacher'">
            {{ record.teacherName || '-' }}
          </template>
          <template v-else-if="column.key === 'subTeacher'">
            {{ assistantTeacherText(record) }}
          </template>
          <template v-else-if="column.key === 'classRoom'">
            {{ record.classroomName || '-' }}
          </template>
          <template v-else-if="column.key === 'action'">
            <span class="flex action">
              <a class="font500" @click="handleRollCall(record)">点名</a>
            </span>
          </template>
        </template>
      </a-table>
    </div>

    <roll-call-drawer
      v-model:open="openDrawer"
      :schedule-id="currentRollCallScheduleId"
      :lesson-day="currentRollCallLessonDay"
      @updated="refreshAfterRollCall"
      @confirmed="refreshAfterRollCall"
    />

    <create-unscheduled-roll-call-modal v-model:open="createUnscheduledRollCallOpen" />

    <a-modal
      v-model:open="batchRollCallSelecting"
      title="批量点名"
      :width="760"
      centered
      :mask-closable="false"
      ok-text="确认点名"
      cancel-text="取消"
      @ok="handleBatchRollCallConfirm"
    >
      <div class="batch-roll-call-modal">
        <div class="batch-roll-call-hint">
          请确认需要批量点名的日程，系统会按当前默认点名规则逐条提交。
        </div>
        <a-table
          row-key="id"
          size="small"
          :data-source="selectedBatchRollCallRows"
          :columns="batchRollCallColumns"
          :pagination="false"
          :row-selection="batchRollCallSelectRowSelection"
          :scroll="{ y: 360 }"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'classDateTime'">
              <div>{{ formatLessonTime(record).dateText }}</div>
              <div class="batch-roll-call-sub">
                {{ formatLessonTime(record).timeText }}
              </div>
            </template>
            <template v-else-if="column.key === 'classOr1v1'">
              <div>{{ classDisplayName(record) }}</div>
              <div v-if="record.canRollCall === false && record.rollCallDisabledReason" class="batch-roll-call-disabled">
                {{ record.rollCallDisabledReason }}
              </div>
            </template>
            <template v-else-if="column.key === 'courseName'">
              {{ record.lessonName || '-' }}
            </template>
            <template v-else-if="column.key === 'mainTeacher'">
              {{ record.teacherName || '-' }}
            </template>
          </template>
        </a-table>
        <div class="batch-roll-call-footer">
          已勾选：{{ batchRollCallCheckedKeys.length }} / {{ batchRollCallLimit }}
        </div>
      </div>
    </a-modal>

    <a-modal
      :open="batchRollCallSubmitting"
      title="批量点名中"
      :width="420"
      centered
      :footer="null"
      :closable="false"
      :mask-closable="false"
      :keyboard="false"
    >
      <div class="batch-roll-call-progress">
        <div class="batch-roll-call-progress-icon">
          <loading-outlined />
        </div>
        <div class="batch-roll-call-progress-title">
          正在批量点名
        </div>
        <div class="batch-roll-call-progress-desc">
          已处理 {{ batchRollCallProgress.completed }} / {{ batchRollCallProgress.total }} 条日程
        </div>
        <div v-if="batchRollCallProgress.currentName" class="batch-roll-call-progress-current">
          当前日程：{{ batchRollCallProgress.currentName }}
        </div>
        <a-progress :percent="batchRollCallProgressPercent" :show-info="false" status="active" />
      </div>
    </a-modal>

    <a-modal
      v-model:open="batchRollCallResultOpen"
      :width="456"
      centered
      :mask-closable="false"
      :footer="null"
    >
      <template #title>
        <div class="batch-roll-call-result-title">
          <span class="batch-roll-call-result-title-icon">
            <exclamation-circle-filled />
          </span>
          <span>操作结果</span>
        </div>
      </template>
      <div class="batch-roll-call-result">
        <div class="batch-roll-call-result-block is-success">
          <div class="batch-roll-call-result-count">
            成功：{{ batchRollCallResult.succeedCount }}条
          </div>
          <div class="batch-roll-call-result-text">
            说明：批量点名已将学员标记为“到课”，点名后，如需编辑请前往“上课记录”中查看
          </div>
        </div>

        <div class="batch-roll-call-result-block is-fail">
          <div class="batch-roll-call-result-count">
            失败：{{ batchRollCallResult.failedCount }}条
          </div>
          <div class="batch-roll-call-result-text">
            说明：没有学员的日程，或未开启消超且学费不足的学员所在的日程，无法批量点名
          </div>
        </div>

        <div v-if="batchRollCallFailureSummary" class="batch-roll-call-result-detail">
          失败明细：{{ batchRollCallFailureSummary }}
        </div>

        <div class="batch-roll-call-result-actions">
          <a-button type="primary" @click="batchRollCallResultOpen = false">
            好的，知道了
          </a-button>
        </div>
      </div>
    </a-modal>
  </div>
</template>

<style lang="less" scoped>
.contentMain {
  box-sizing: content-box;
  padding: 16px 12px 6px 24px;
  height: 30px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: bold;

  .contentMainLeft {
    font-size: 14px;
    font-weight: 500;
    color: #222;
    flex-shrink: 0;
  }

  .contentMainRight {
    min-width: 72px;
    height: 30px;
    font-size: 30px;
    font-weight: 700;
    font-family: DINAlternate-Bold, DINAlternate;
    color: #06f;
    line-height: 30px;
    flex-shrink: 0;
    text-align: center;
  }
}

.contentSub {
  padding: 0 24px;
  height: 16px;
  line-height: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: bold;

  .contentSubLeft {
    font-size: 13px;
    color: #888;
  }

  .contentSubRight {
    font-size: 12px;
    color: #06f;
  }
}

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

.batch-roll-call-modal {
  .batch-roll-call-hint {
    margin-bottom: 12px;
    font-size: 13px;
    color: #666;
  }

  .batch-roll-call-sub {
    color: #888;
  }

  .batch-roll-call-disabled {
    margin-top: 2px;
    color: #ff7a45;
    font-size: 12px;
    line-height: 18px;
  }

  .batch-roll-call-footer {
    display: flex;
    justify-content: flex-end;
    margin-top: 12px;
    color: #666;
    font-size: 13px;
  }
}

.batch-roll-call-progress {
  padding: 8px 4px 4px;
  text-align: center;

  .batch-roll-call-progress-icon {
    margin-bottom: 12px;
    font-size: 30px;
    color: #06f;
  }

  .batch-roll-call-progress-title {
    font-size: 16px;
    font-weight: 600;
    color: #222;
  }

  .batch-roll-call-progress-desc {
    margin-top: 8px;
    color: #666;
  }

  .batch-roll-call-progress-current {
    margin: 12px 0 16px;
    color: #444;
    word-break: break-all;
  }
}

.batch-roll-call-result {
  padding: 4px 8px 0 16px;

  .batch-roll-call-result-block + .batch-roll-call-result-block {
    margin-top: 18px;
  }

  .batch-roll-call-result-count {
    font-size: 15px;
    font-weight: 600;
    line-height: 24px;
  }

  .batch-roll-call-result-text {
    margin-top: 4px;
    max-width: 340px;
    color: #666;
    font-size: 13px;
    line-height: 22px;
  }

  .is-success .batch-roll-call-result-count {
    color: #1677ff;
  }

  .is-fail .batch-roll-call-result-count {
    color: #ff4d4f;
  }

  .batch-roll-call-result-detail {
    margin-top: 18px;
    max-height: 108px;
    overflow-y: auto;
    color: #666;
    font-size: 12px;
    line-height: 20px;
  }

  .batch-roll-call-result-actions {
    display: flex;
    justify-content: flex-end;
    margin-top: 24px;
  }
}

.batch-roll-call-result-title {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 15px;
  font-weight: 600;
  color: #222;
}

.batch-roll-call-result-title-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  color: #1677ff;
  font-size: 22px;
}
</style>
